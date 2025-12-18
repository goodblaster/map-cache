package v1keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/logos"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
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
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid json payload"))
		}

		if err := req.Validate(); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request body"))
		}

		cache := Cache(c)
		if err := cache.Create(ctx, req.Entries); err != nil {
			if errors.Is(err, caches.ErrKeyAlreadyExists) {
				return v1errors.ApiError(c, http.StatusConflict, errors.Wrap(err, "keys already exist"))
			}
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "failed to create keys"))
		}

		// TTLs
		for key, ttl := range req.TTL {
			if err := cache.SetKeyTTL(ctx, key, ttl); err != nil {
				logos.WithError(err).Errorf("could not set cache expiration for key %q", key)
			}
		}

		return c.NoContent(http.StatusCreated)
	}
}
