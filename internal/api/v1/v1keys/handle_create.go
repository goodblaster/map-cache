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
	TTL     map[string]int64 `json:"ttl"`
} // @name CreateKeysRequest

// Validate - Validates the createKeysRequest.
func (req createKeysRequest) Validate() error { return nil }

// handleCreate creates new entries in a cache.
//
// @Summary Create cache entries
// @Description Creates one or more keys in the cache with values
// @Tags keys
// @Accept json
// @Produce json
// @Param  body  body  createKeysRequest  true  "Request body"
// @Success 201 {string} string "Created"
// @Failure 400 {object} v1errors.ErrorResponse "Bad request – invalid JSON or failed validation"
// @Failure 409 {object} v1errors.ErrorResponse "Conflict – cache key already exists"
// @Failure 500 {object} v1errors.ErrorResponse "Internal server error"
// @Router /keys [post]
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

		// Triggers?
		//

		return c.NoContent(http.StatusCreated)
	}
}
