package caches

import (
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// handleDeleteCache deletes a named cache.
func handleDeleteCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("name")
		if name == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing cache name")
		}

		// Deleting the cache, will also clear the TTL timers.
		err := caches.DeleteCache(name)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "could not find cache").SetInternal(err)
		}

		return c.NoContent(http.StatusOK)
	}
}
