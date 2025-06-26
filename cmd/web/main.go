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
	"github.com/dukerupert/faa-aircraft-search/internal/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx := context.Background()

	// Initialize database connection
	db, err := database.InitDatabase(ctx)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Update total aircraft count metric on startup
	if count, err := db.Queries.CountAircraft(ctx); err == nil {
		middleware.UpdateTotalAircraftCount(float64(count))
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Prometheus metrics middleware
	e.Use(middleware.PrometheusMiddleware())

	// Request timeout middleware
	e.Use(echomiddleware.TimeoutWithConfig(echomiddleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// Static file serving (for any additional static assets)
	e.Static("/static", "web/static")

	// Initialize handler with database
	h := handler.New(db)

	// Metrics endpoint (exclude from metrics middleware to avoid recursion)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Web routes (HTML pages)
	e.GET("/", h.Home)
	e.GET("/search", h.Search)
	e.GET("/aircraft-list", h.AircraftList)
	e.GET("/aircraft-details/:id", h.AircraftDetails)

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

	// Static file serving (for any additional static assets)
	e.Static("/static", "web/static")

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
