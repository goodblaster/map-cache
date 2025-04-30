package v1caches

import (
	"net/http"

	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type CreateCacheRequest struct {
	Name       string `json:"name"`
	Expiration string `json:"expiration"`
}

// handleCreateCache - Handler for creating new cache.
func handleCreateCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		var body CreateCacheRequest
		if err := c.Bind(&body); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		err := caches.AddCache(body.Name)
		if err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, "failed to create cache") // or StatusBadRequest? depends on error
		}

		// Expiration?
		//

		return c.NoContent(http.StatusCreated)
	}
}
