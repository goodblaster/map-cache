package keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/log"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// createKeysRequest is the request body for creating new cache entries.
type createKeysRequest struct {
	Entries map[string]any   `json:"entries"`
	TTL     map[string]int64 `json:"ttl"` // milliseconds
}

// Validate - Validates the createKeysRequest.
func (req createKeysRequest) Validate() error {
	if len(req.Entries) == 0 {
		return errors.New("at least one entry is required")
	}
	for key := range req.Entries {
		if key == "" {
			return errors.New("key cannot be empty")
		}
	}
	// Validate TTL keys exist in Entries
	for key := range req.TTL {
		if _, exists := req.Entries[key]; !exists {
			return errors.Newf("TTL specified for non-existent key: %s", key)
		}
	}
	return nil
}

// handleCreate creates new entries in a cache.
func handleCreate() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var req createKeysRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid json payload").SetInternal(err)
		}

		if err := req.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "validation failed").SetInternal(err)
		}

		cache := Cache(c)
		if err := cache.Create(ctx, req.Entries); err != nil {
			if errors.Is(err, caches.ErrKeyAlreadyExists) {
				return echo.NewHTTPError(http.StatusConflict, "keys already exist").SetInternal(err)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create keys").SetInternal(err)
		}

		// TTLs
		for key, ttl := range req.TTL {
			if err := cache.SetKeyTTL(ctx, key, ttl); err != nil {
				log.FromContext(ctx).WithError(err).With("key", key).With("ttl_ms", ttl).Error("could not set cache expiration")
			}
		}

		return c.NoContent(http.StatusCreated)
	}
}
