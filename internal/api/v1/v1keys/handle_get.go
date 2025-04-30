package v1keys

import (
	"net/http"

	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/labstack/echo/v4"
)

// handleGetValue - Handler for getting a value from the cache.
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

type getBatchRequest struct {
	Keys []string `json:"keys"`
}

// handleGetBatch - Handler for batch getting values from the cache.
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
