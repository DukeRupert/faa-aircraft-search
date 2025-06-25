package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dukerupert/faa-aircraft-search/internal/database"
	"github.com/dukerupert/faa-aircraft-search/internal/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	ctx := context.Background()

	// Initialize database connection
	db, err := database.InitDatabase(ctx)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Request timeout middleware
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// Initialize handler with database
	h := handler.New(db)

	// Base health check route
	e.GET("/health", h.HealthCheck)

	// API v1 routes
	v1 := e.Group("/api/v1")
	{
		aircraft := v1.Group("/aircraft")
		{
			aircraft.GET("/search", h.SearchAircraft)
			aircraft.GET("/:id", h.GetAircraft)
		}
	}

	// Static file serving (for frontend)
	e.Static("/", "web/static")

	// Start server in a goroutine
	go func() {
		log.Println("Starting server on :8080")
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}