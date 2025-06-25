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
)

func main() {
	ctx := context.Background()

	// Initialize database connection
	pool, err := database.InitDatabase(ctx)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close(pool)

	// Initialize handler with database pool
	h := handler.New(pool)

	// Setup routes
	mux := http.NewServeMux()
	
	// API routes
	mux.HandleFunc("/api/aircraft/search", h.SearchAircraft)
	mux.HandleFunc("/api/aircraft/{id}", h.GetAircraft)
	mux.HandleFunc("/api/health", h.HealthCheck)
	
	// Static file serving (for frontend)
	mux.Handle("/", http.FileServer(http.Dir("./web/static/")))

	// Server configuration
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}