package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dukerupert/faa-aircraft-search/internal/database"
	"github.com/dukerupert/faa-aircraft-search/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	db *database.Database
}

// SearchRequest represents the search query parameters
type SearchRequest struct {
	Query string `query:"q"`
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
}

// SearchResponse represents the search API response
type SearchResponse struct {
	Aircraft []db.AircraftDatum `json:"aircraft"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status        string    `json:"status"`
	Database      string    `json:"database"`
	AircraftCount int64     `json:"aircraft_count"`
	Timestamp     time.Time `json:"timestamp"`
}

func New(db *database.Database) *Handlers {
	return &Handlers{db: db}
}

// SearchAircraft handles GET /api/aircraft/search
func (h *Handlers) SearchAircraft(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// Bind query parameters
	req := &SearchRequest{
		Page:  1,
		Limit: 50,
	}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid query parameters",
		})
	}

	// Validate pagination parameters
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 50
	}

	offset := int32((req.Page - 1) * req.Limit)
	limit := int32(req.Limit)

	var aircraft []db.AircraftDatum
	var total int64
	var err error

	if req.Query == "" {
		// Get all aircraft with pagination
		aircraft, err = h.db.Queries.GetAllAircraft(ctx, db.GetAllAircraftParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to retrieve aircraft data",
			})
		}

		// Get total count
		total, err = h.db.Queries.CountAircraft(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to count aircraft records",
			})
		}
	} else {
		// Search aircraft with the query
		searchTerm := "%" + strings.ToUpper(req.Query) + "%"
		
		aircraft, err = h.db.Queries.SearchAircraft(ctx, db.SearchAircraftParams{
			SearchTerm: searchTerm,
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to search aircraft data",
			})
		}

		// Get search result count
		total, err = h.db.Queries.CountSearchAircraft(ctx, searchTerm)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to count search results",
			})
		}
	}

	response := SearchResponse{
		Aircraft: aircraft,
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
	}

	return c.JSON(http.StatusOK, response)
}

// GetAircraft handles GET /api/aircraft/:id
func (h *Handlers) GetAircraft(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	// Parse ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid aircraft ID",
		})
	}

	aircraft, err := h.db.Queries.GetAircraft(ctx, int32(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Aircraft not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve aircraft data",
		})
	}

	return c.JSON(http.StatusOK, aircraft)
}

// HealthCheck handles GET /api/health
func (h *Handlers) HealthCheck(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	// Test database connection
	err := h.db.Ping(ctx)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "database_unavailable",
			Message: "Database connection failed",
		})
	}

	// Get record count
	count, err := h.db.Queries.CountAircraft(ctx)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "database_query_failed",
			Message: "Failed to query database",
		})
	}

	response := HealthResponse{
		Status:        "healthy",
		Database:      "connected",
		AircraftCount: count,
		Timestamp:     time.Now().UTC(),
	}

	return c.JSON(http.StatusOK, response)
}