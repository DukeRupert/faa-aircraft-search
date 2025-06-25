package handler

import (
	"github.com/dukerupert/faa-aircraft-search/web/templates/components"
	"github.com/dukerupert/faa-aircraft-search/web/templates/pages"
	"github.com/labstack/echo/v4"
)

// Home renders the main search page
func (h *Handlers) Home(c echo.Context) error {
	return pages.Home().Render(c.Request().Context(), c.Response().Writer)
}

// SimpleTest handles a simple test endpoint
func (h *Handlers) SimpleTest(c echo.Context) error {
	return components.SimpleSearchResults("This is a simple test message!").Render(c.Request().Context(), c.Response().Writer)
}