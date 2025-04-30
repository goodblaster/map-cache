package v1caches

import (
	"net/http"

	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type deleteCacheRequest struct {
	Name string `json:"name"`
}

// handleDeleteCache - Handler for deleting a cache.
func handleDeleteCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req deleteCacheRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		err := caches.DeleteCache(req.Name)
		if err != nil {
			return v1errors.ApiError(c, http.StatusNotFound, "could not find cache")
		}

		return c.NoContent(http.StatusOK)
	}
}
