package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/dukerupert/faa-aircraft-search/internal/db"
	"github.com/dukerupert/faa-aircraft-search/web/templates/components"
	"github.com/dukerupert/faa-aircraft-search/web/templates/pages"
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
		// Return just the aircraft container for HTMX requests
		return components.AircraftContainer(aircraft, total, page, limit).Render(ctx, c.Response().Writer)
	}

	// For non-HTMX requests, redirect to home with page parameter
	return c.Redirect(http.StatusSeeOther, "/?page="+strconv.Itoa(page))
}

// SimpleTest handles a simple test endpoint
func (h *Handlers) SimpleTest(c echo.Context) error {
	return components.SimpleSearchResults("This is a simple test message!").Render(c.Request().Context(), c.Response().Writer)
}