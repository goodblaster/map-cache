package caches

import (
	"net/http"

	v1errors "github.com/goodblaster/map-cache/internal/api/v1/errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// updateCacheRequest represents the payload for updating a cache.
type updateCacheRequest struct {
	TTL *int64 `json:"ttl"` // milliseconds
}

// handleUpdateCache updates the expiration time of a cache.
func handleUpdateCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input updateCacheRequest
		if err := c.Bind(&input); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		cacheName := c.Param("name")
		if cacheName == "" {
			return v1errors.ApiError(c, http.StatusBadRequest, "cache name is required")
		}

		// Cannot modify the default cache
		if cacheName == caches.DefaultName {
			return v1errors.ApiError(c, http.StatusBadRequest, "cannot modify the default cache")
		}

		if input.TTL != nil {
			// Update the cache expiration
			err := caches.SetCacheTTL(cacheName, *input.TTL)
			if err != nil {
				return v1errors.ApiError(c, http.StatusInternalServerError, "failed to update cache")
			}

			return c.NoContent(http.StatusNoContent)
		}

		// If the TTL is nil, we interpret it as a request to remove the TTL
		err := caches.CancelCacheExpiration(cacheName)
		if err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, "could not remove cache expiration")
		}

		return c.NoContent(http.StatusNoContent)
	}
}
