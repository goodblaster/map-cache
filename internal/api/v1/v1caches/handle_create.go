package v1caches

import (
	"net/http"
	"time"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// createCacheRequest represents the payload for creating a cache.
//
// swagger:model createCacheRequest
type createCacheRequest struct {
	// Name of the cache to create
	// required: true
	Name string `json:"name"`

	// Expiration duration for the cache in Go duration format (e.g., "5m", "1h").
	// Currently not implemented.
	Expiration *time.Time `json:"expiration,omitempty"`
} // @name CreateCacheRequest

func (req createCacheRequest) Validate() error {
	if req.Name == "" {
		return errors.New("cache name is required")
	}

	return nil
}

// handleCreateCache creates a new cache.
//
// @Summary Create a new cache
// @Description Creates a new named cache. Optionally accepts expiration (not yet implemented).
// @Tags caches
// @Accept json
// @Produce json
// @Param body body createCacheRequest true "Cache creation payload"
// @Success 201 {string} string "Created"
// @Failure 400 {object} v1errors.ErrorResponse "Invalid request or bad payload"
// @Failure 500 {object} v1errors.ErrorResponse "Internal server error"
// @Router /caches [post]
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

		// TODO: Handle expiration if implemented later

		return c.NoContent(http.StatusCreated)
	}
}
