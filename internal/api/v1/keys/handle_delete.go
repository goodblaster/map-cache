package keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	v1errors "github.com/goodblaster/map-cache/internal/api/v1/errors"
	"github.com/labstack/echo/v4"
)

// deleteBatchRequest represents the request body for batch key deletion.
type deleteBatchRequest struct {
	// List of keys to delete
	// required: true
	Keys []string `json:"keys"`
}

// Validate - Validates the deleteBatchRequest.
func (req deleteBatchRequest) Validate() error {
	if len(req.Keys) == 0 {
		return errors.New("at least one key is required")
	}
	for _, key := range req.Keys {
		if key == "" {
			return errors.New("key cannot be empty")
		}
	}
	return nil
}

// handleDelete handles deletion of a single cache key.
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
func handleDeleteBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req deleteBatchRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := req.Validate(); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request body"))
		}

		if err := cache.Delete(c.Request().Context(), req.Keys...); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not delete keys"))
		}

		return c.NoContent(http.StatusOK)
	}
}
