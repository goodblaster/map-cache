package v1keys

import (
	"net/http"

	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/labstack/echo/v4"
)

// getBatchRequest represents the request body for retrieving multiple keys.
//
// swagger:model getBatchRequest
type getBatchRequest struct {
	// List of keys to retrieve
	// required: true
	Keys []string `json:"keys"`
} // @name GetBatchRequest

// handleGetValue retrieves a single value from the cache.
//
// @Summary Get a single value
// @Description Retrieves the value associated with a single cache key
// @Tags keys
// @Produce json
// @Param key path string true "Key to retrieve"
// @Success 200 {object} interface{} "Value for the given key"
// @Failure 404 {object} v1errors.ErrorResponse "Key not found"
// @Router /api/v1/keys/{key} [get]
func handleGetValue() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		value, err := cache.Get(c.Request().Context(), key)
		if err != nil {
			return v1errors.ApiError(c, http.StatusNotFound, "key not found")
		}

		return c.JSON(http.StatusOK, value)
	}
}

// handleGetBatch retrieves multiple values from the cache.
//
// @Summary Get multiple values
// @Description Retrieves values for a list of cache keys
// @Tags keys
// @Accept json
// @Produce json
// @Param body body getBatchRequest true "List of keys to retrieve"
// @Success 200 {object} map[string]interface{} "Map of keys to values"
// @Failure 400 {object} v1errors.ErrorResponse "Invalid request body"
// @Failure 404 {object} v1errors.ErrorResponse "One or more keys not found"
// @Router /api/v1/keys/get [post]
func handleGetBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req getBatchRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		value, err := cache.BatchGet(c.Request().Context(), req.Keys...)
		if err != nil {
			return v1errors.ApiError(c, http.StatusNotFound, "key not found")
		}

		return c.JSON(http.StatusOK, value)
	}
}
