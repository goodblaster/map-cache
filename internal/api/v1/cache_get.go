package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandleGetValue - Handler for getting a value from the cache.
func (v V1) HandleGetValue() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := v.Cache(c)
		key := c.Param("key")

		value, err := cache.Get(c.Request().Context(), key)
		if err != nil {
			return c.String(http.StatusNotFound, "Not Found")
		}

		return c.JSON(http.StatusOK, value)
	}
}

// HandleGetBatch - Handler for batch getting values from the cache.
func (v V1) HandleGetBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := v.Cache(c)
		var keys []string
		if err := json.NewDecoder(c.Request().Body).Decode(&keys); err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid request body")
		}

		value, err := cache.BatchGet(c.Request().Context(), keys...)
		if err != nil {
			return c.String(http.StatusNotFound, "Not Found")
		}

		return c.JSON(http.StatusOK, value)
	}
}
