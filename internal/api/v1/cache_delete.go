package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandleDelete - Handler for deleting a single key from the cache.
func HandleDelete() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		if err := cache.Delete(c.Request().Context(), key); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, WebError(err))
		}

		return c.NoContent(http.StatusOK)
	}
}

type DeleteBatchBody []string

// HandleDeleteBatch - Handler for deleting multiple keys from the cache.
func HandleDeleteBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var keys DeleteBatchBody
		if err := json.NewDecoder(c.Request().Body).Decode(&keys); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
		}

		if err := cache.Delete(c.Request().Context(), keys...); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, WebError(err))
		}

		return c.NoContent(http.StatusOK)
	}
}
