package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dukerupert/faa-aircraft-search/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Database wraps the connection pool and SQLC queries
type Database struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

// GetConfigFromEnv reads database configuration from environment variables
func GetConfigFromEnv() *Config {
	config := &Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "5432"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	// Validate required environment variables
	if config.User == "" {
		log.Fatal("POSTGRES_USER environment variable is required")
	}
	if config.Password == "" {
		log.Fatal("POSTGRES_PASSWORD environment variable is required")
	}
	if config.DBName == "" {
		log.Fatal("POSTGRES_DB environment variable is required")
	}

	return config
}

// InitDatabase initializes the database connection and creates the database if it doesn't exist
func InitDatabase(ctx context.Context) (*Database, error) {
	config := GetConfigFromEnv()

	// First, connect to the default 'postgres' database to check if our target database exists
	adminConnString := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.SSLMode)

	log.Printf("Connecting to PostgreSQL server at %s:%s as user %s", config.Host, config.Port, config.User)

	// Connect to the admin database with retry logic
	var adminConn *pgx.Conn
	var err error
	
	for attempts := 0; attempts < 30; attempts++ {
		adminConn, err = pgx.Connect(ctx, adminConnString)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to PostgreSQL (attempt %d/30): %v", attempts+1, err)
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL after 30 attempts: %w", err)
	}
	defer adminConn.Close(ctx)

	// Check if the target database exists
	var exists bool
	err = adminConn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", config.DBName).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create the database if it doesn't exist
	if !exists {
		log.Printf("Database '%s' does not exist. Creating it...", config.DBName)
		
		// Note: Database names cannot be parameterized in PostgreSQL, but we validate the name
		if !isValidDatabaseName(config.DBName) {
			return nil, fmt.Errorf("invalid database name: %s", config.DBName)
		}
		
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{config.DBName}.Sanitize())
		_, err = adminConn.Exec(ctx, createDBQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to create database '%s': %w", config.DBName, err)
		}
		log.Printf("Database '%s' created successfully", config.DBName)
	} else {
		log.Printf("Database '%s' already exists", config.DBName)
	}

	// Now connect to the target database
	targetConnString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.DBName, config.SSLMode)

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(targetConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set connection pool settings
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create SQLC queries instance
	queries := db.New(pool)

	log.Printf("Successfully connected to database '%s'", config.DBName)
	
	return &Database{
		Pool:    pool,
		Queries: queries,
	}, nil
}

// BeginTx starts a new transaction and returns a Queries instance that uses the transaction
func (d *Database) BeginTx(ctx context.Context) (pgx.Tx, *db.Queries, error) {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	queries := d.Queries.WithTx(tx)
	return tx, queries, nil
}

// Ping tests the database connection
func (d *Database) Ping(ctx context.Context) error {
	return d.Pool.Ping(ctx)
}

// Close closes the database connection pool
func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
		log.Println("Database connection pool closed")
	}
}

// Legacy function for backward compatibility
func InitDatabaseLegacy(ctx context.Context) (*pgxpool.Pool, error) {
	db, err := InitDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return db.Pool, nil
}

// Legacy function for backward compatibility
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
		log.Println("Database connection pool closed")
	}
}

// getEnvOrDefault returns the environment variable value or a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// isValidDatabaseName validates that the database name contains only safe characters
func isValidDatabaseName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	
	// Database names should start with a letter or underscore
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	
	// Check remaining characters (letters, digits, underscores, and hyphens are allowed)
	for _, char := range name[1:] {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '_' || char == '-') {
			return false
		}
	}
	
	return true
}