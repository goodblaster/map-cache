package v1keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/labstack/echo/v4"
)

// handleDelete - Handler for deleting a single key from the cache.
func handleDelete() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		if err := cache.Delete(c.Request().Context(), key); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not delete key"))
		}

		return c.NoContent(http.StatusOK)
	}
}

type deleteBatchRequest struct {
	Keys []string `json:"keys"`
}

// handleDeleteBatch - Handler for deleting multiple V1Keys from the cache.
func handleDeleteBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req deleteBatchRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := cache.Delete(c.Request().Context(), req.Keys...); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not delete V1Keys"))
		}

		return c.NoContent(http.StatusOK)
	}
}
