package v1keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/labstack/echo/v4"
)

// handlePutRequest represents the request body for replacing a single cache value.
//
// swagger:model handlePutRequest
type handlePutRequest struct {
	// New value to store for the key
	// required: true
	Value any `json:"value"`
}

// handlePut replaces the value of a single key in the cache.
//
// @Summary Replace a single value
// @Description Replaces the value of a key in the cache
// @Tags keys
// @Accept json
// @Produce json
// @Param key path string true "Key to update"
// @Param body body handlePutRequest true "New value for the key"
// @Success 200 {string} string "Value replaced successfully"
// @Failure 400 {object} v1errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} v1errors.ErrorResponse "Internal server error"
// @Router /api/v1/keys/{key} [put]
func handlePut() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		var req handlePutRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := cache.Replace(c.Request().Context(), key, req.Value); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not replace contents"))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}

// replaceBatchRequest represents the request body for batch replacing values.
//
// swagger:model replaceBatchRequest
type replaceBatchRequest struct {
	// Map of keys to their new values
	// required: true
	Entries map[string]any `json:"entries"`
}

// handleReplaceBatch replaces multiple key-value pairs in the cache.
//
// @Summary Replace multiple values
// @Description Replaces multiple entries in the cache
// @Tags keys
// @Accept json
// @Produce json
// @Param body body replaceBatchRequest true "Map of key-value pairs to replace"
// @Success 200 {string} string "Values replaced successfully"
// @Failure 400 {object} v1errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} v1errors.ErrorResponse "Internal server error"
// @Router /api/v1/keys [put]
func handleReplaceBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req replaceBatchRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := cache.ReplaceBatch(c.Request().Context(), req.Entries); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not replace contents"))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}
