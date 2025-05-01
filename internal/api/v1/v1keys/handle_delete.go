package v1keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/labstack/echo/v4"
)

// deleteBatchRequest represents the request body for batch key deletion.
//
// swagger:model deleteBatchRequest
type deleteBatchRequest struct {
	// List of keys to delete
	// required: true
	Keys []string `json:"keys"`
} // @name DeleteBatchRequest

// handleDelete handles deletion of a single cache key.
//
// @Summary Delete a single key
// @Description Deletes a single key from the specified cache
// @Tags keys
// @Produce json
// @Param key path string true "Key to delete"
// @Success 200 {string} string "Key deleted successfully"
// @Failure 500 {object} v1errors.ErrorResponse "Server error"
// @Router /api/v1/keys/{key} [delete]
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

// handleDeleteBatch handles deletion of multiple keys.
//
// @Summary Delete multiple keys
// @Description Deletes multiple keys from the specified cache
// @Tags keys
// @Accept json
// @Produce json
// @Param body body deleteBatchRequest true "List of keys to delete"
// @Success 200 {string} string "Keys deleted successfully"
// @Failure 400 {object} v1errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} v1errors.ErrorResponse "Server error"
// @Router /api/v1/keys/delete [post]
func handleDeleteBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req deleteBatchRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := cache.Delete(c.Request().Context(), req.Keys...); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not delete keys"))
		}

		return c.NoContent(http.StatusOK)
	}
}
