package caches

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/logos"
	v1errors "github.com/goodblaster/map-cache/internal/api/v1/errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// createCacheRequest represents the payload for creating a cache.
type createCacheRequest struct {
	Name string `json:"name,required"`
	TTL  *int64 `json:"ttl,omitempty"` // millisecond
}

func (req createCacheRequest) Validate() error {
	if req.Name == "" {
		return errors.New("cache name is required")
	}

	return nil
}

// handleCreateCache creates a new cache.
func handleCreateCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req createCacheRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := req.Validate(); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, err)
		}

		err := caches.AddCache(req.Name)
		if err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, "failed to create cache")
		}

		// Expiration
		if req.TTL != nil {
			if err := caches.SetCacheTTL(req.Name, *req.TTL); err != nil {
				logos.WithError(err).Error("could not set cache expiration")
			}
		}

		return c.NoContent(http.StatusCreated)
	}
}
