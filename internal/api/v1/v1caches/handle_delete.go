package v1caches

import (
	"net/http"

	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// handleDeleteCache deletes a named cache.
func handleDeleteCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("name")
		if name == "" {
			return v1errors.ApiError(c, http.StatusBadRequest, "missing cache name")
		}

		// Deleting the cache, will also clear the TTL timers.
		err := caches.DeleteCache(name)
		if err != nil {
			return v1errors.ApiError(c, http.StatusNotFound, "could not find cache")
		}

		return c.NoContent(http.StatusOK)
	}
}
