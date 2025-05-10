package v1caches

import (
	"net/http"
	"time"

	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// updateCacheRequest represents the payload for updating a cache.
//
// swagger:model updateCacheRequest
type updateCacheRequest struct {
	// TTL for the cache in seconds
	// required: true
	TTL *int64 `json:"ttl"`
} // @name UpdateCacheRequest

// handleUpdateCache updates the expiration time of a cache.
//
// @Summary Update cache expiration
// @Description Updates the expiration time of a named cache. If no TTL is provided, it removes the expiration.
// @Tags caches
// @Param body body updateCacheRequest true "Cache update payload"
// @Produce json
// @Success 204
// @Failure 400 {object} v1errors.ErrorResponse "Invalid request or bad payload"
// @Failure 500 {object} v1errors.ErrorResponse "Internal server error"
// @Router /caches/{name} [put]
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
			duration := time.Duration(*input.TTL) * time.Second
			err := caches.SetCacheTTL(cacheName, duration)
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
