package v1caches

import (
	"net/http"

	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// handleDeleteCache deletes a named cache.
//
// @Summary Delete a cache
// @Description Deletes a cache with the specified name.
// @Tags caches
// @Produce json
// @Param name path string true "Name of the cache to delete"
// @Success 204
// @Failure 400 {object} v1errors.ErrorResponse "Invalid cache name"
// @Failure 404 {object} v1errors.ErrorResponse "Cache not found"
// @Router /caches/{name} [delete]
func handleDeleteCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("name")
		if name == "" {
			return v1errors.ApiError(c, http.StatusBadRequest, "missing cache name")
		}

		err := caches.DeleteCache(name)
		if err != nil {
			return v1errors.ApiError(c, http.StatusNotFound, "could not find cache")
		}

		return c.NoContent(http.StatusOK)
	}
}
