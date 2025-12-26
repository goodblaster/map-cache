package caches

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/log"
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
			return echo.NewHTTPError(http.StatusBadRequest, "invalid json payload").SetInternal(err)
		}

		if err := req.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "validation failed").SetInternal(err)
		}

		err := caches.AddCache(req.Name)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create cache").SetInternal(err)
		}

		// Expiration
		if req.TTL != nil {
			if err := caches.SetCacheTTL(req.Name, *req.TTL); err != nil {
				log.FromContext(c.Request().Context()).WithError(err).With("cache", req.Name).With("ttl_ms", *req.TTL).Error("could not set cache expiration")
			}
		}

		return c.NoContent(http.StatusCreated)
	}
}
