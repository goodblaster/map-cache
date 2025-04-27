package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandleGetValue - Handler for getting a value from the cache.
func HandleGetValue() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		value, err := cache.Get(c.Request().Context(), key)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "key not found")
		}

		return c.JSON(http.StatusOK, value)
	}
}

type GetBatchBody []string

// HandleGetBatch - Handler for batch getting values from the cache.
func HandleGetBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var keys GetBatchBody
		if err := json.NewDecoder(c.Request().Body).Decode(&keys); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		value, err := cache.BatchGet(c.Request().Context(), keys...)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "key not found")
		}

		return c.JSON(http.StatusOK, value)
	}
}
