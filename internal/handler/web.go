package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dukerupert/faa-aircraft-search/internal/db"
	"github.com/dukerupert/faa-aircraft-search/internal/middleware"
	"github.com/dukerupert/faa-aircraft-search/web/templates/components"
	"github.com/dukerupert/faa-aircraft-search/web/templates/pages"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// Home renders the main search page with aircraft list
func (h *Handlers) Home(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// Get page parameter, default to 1
	page := 1
	if pageParam := c.QueryParam("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	// Set limit to 10 per page
	limit := 10
	offset := int32((page - 1) * limit)

	// Get aircraft with pagination
	aircraft, err := h.db.Queries.GetAllAircraft(ctx, db.GetAllAircraftParams{
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	// Get total count
	total, err := h.db.Queries.CountAircraft(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	return pages.Home(aircraft, total, page, limit).Render(ctx, c.Response().Writer)
}

// Search handles HTMX search requests
func (h *Handlers) Search(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	query := strings.TrimSpace(c.QueryParam("q"))
	
	// Get page parameter, default to 1
	page := 1
	if pageParam := c.QueryParam("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	// Set limit to 10 per page
	limit := 10
	offset := int32((page - 1) * limit)
	
	// If query is empty, return to all aircraft view
	if query == "" {
		// Get aircraft with pagination
		aircraft, err := h.db.Queries.GetAllAircraft(ctx, db.GetAllAircraftParams{
			Limit:  int32(limit),
			Offset: offset,
		})
		if err != nil {
			return c.String(http.StatusInternalServerError, "Database error")
		}

		// Get total count
		total, err := h.db.Queries.CountAircraft(ctx)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Database error")
		}

		return components.AircraftContainer(aircraft, total, page, limit).Render(ctx, c.Response().Writer)
	}

	// Search aircraft with pagination
	searchTerm := "%" + strings.ToUpper(query) + "%"
	
	aircraft, err := h.db.Queries.SearchAircraft(ctx, db.SearchAircraftParams{
		SearchTerm: searchTerm,
		Limit:      int32(limit),
		Offset:     offset,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	// Get total count for the search
	total, err := h.db.Queries.CountSearchAircraft(ctx, searchTerm)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	return components.SearchResults(aircraft, total, query, page, limit).Render(ctx, c.Response().Writer)
}

// AircraftList handles HTMX requests for paginated aircraft list
func (h *Handlers) AircraftList(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// Get page parameter, default to 1
	page := 1
	if pageParam := c.QueryParam("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	// Set limit to 10 per page
	limit := 10
	offset := int32((page - 1) * limit)

	// Get aircraft with pagination
	aircraft, err := h.db.Queries.GetAllAircraft(ctx, db.GetAllAircraftParams{
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	// Get total count
	total, err := h.db.Queries.CountAircraft(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}

	// Check if this is an HTMX request
	if c.Request().Header.Get("HX-Request") == "true" {
		// Return the aircraft container for HTMX requests
		return components.AircraftContainer(aircraft, total, page, limit).Render(ctx, c.Response().Writer)
	}

	// For non-HTMX requests, redirect to home with page parameter
	return c.Redirect(http.StatusSeeOther, "/?page="+strconv.Itoa(page))
}

// AircraftDetails handles GET /aircraft-details/:id
func (h *Handlers) AircraftDetails(c echo.Context) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	// Parse ID parameter
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 32)
	if err != nil {
		middleware.RecordDatabaseQuery("get_details", time.Since(start), false)
		return c.String(http.StatusBadRequest, "Invalid aircraft ID")
	}

	aircraft, err := h.db.Queries.GetAircraft(ctx, int32(id))
	middleware.RecordDatabaseQuery("get_details", time.Since(start), err == nil)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.String(http.StatusNotFound, "Aircraft not found")
		}
		return c.String(http.StatusInternalServerError, "Database error")
	}

	// Record detail view metric
	middleware.RecordAircraftDetailView()

	return components.AircraftDetails(aircraft).Render(ctx, c.Response().Writer)
}