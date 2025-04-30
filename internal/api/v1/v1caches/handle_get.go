package v1caches

import (
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// handleGetCacheList - Handler for getting list of all caches.
func handleGetCacheList() echo.HandlerFunc {
	return func(c echo.Context) error {
		list := caches.List()
		return c.JSON(http.StatusOK, list)
	}
}
