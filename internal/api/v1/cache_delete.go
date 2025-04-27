package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandleDelete - Handler for deleting a single key from the cache.
func (v V1) HandleDelete() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := v.Cache(c)
		key := c.Param("key")

		if err := cache.Delete(c.Request().Context(), key); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, v.WebError(err))
		}

		return c.NoContent(http.StatusOK)
	}
}

// HandleDeleteBatch - Handler for deleting multiple keys from the cache.
func (v V1) HandleDeleteBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := v.Cache(c)
		var keys []string
		if err := json.NewDecoder(c.Request().Body).Decode(&keys); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
		}

		if err := cache.Delete(c.Request().Context(), keys...); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, v.WebError(err))
		}

		return c.NoContent(http.StatusOK)
	}
}
