package v1caches

import (
	"net/http"

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
	Expiration string `json:"expiration"`
} // @name CreateCacheRequest

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
		var body createCacheRequest
		if err := c.Bind(&body); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		err := caches.AddCache(body.Name)
		if err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, "failed to create cache")
		}

		// TODO: Handle expiration if implemented later

		return c.NoContent(http.StatusCreated)
	}
}
