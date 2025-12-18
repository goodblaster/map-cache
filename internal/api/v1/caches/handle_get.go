package caches

import (
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// handleGetCacheList returns a list of all active caches.
func handleGetCacheList() echo.HandlerFunc {
	return func(c echo.Context) error {
		list := caches.List()
		return c.JSON(http.StatusOK, list)
	}
}
