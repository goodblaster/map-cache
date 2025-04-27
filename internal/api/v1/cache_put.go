package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandlePutBody - Body for the HandlePut function.
type HandlePutBody any

func HandlePut() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		var body HandlePutBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		if err := cache.Replace(c.Request().Context(), key, body); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, WebError(err))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}

// ReplaceBatchBody - Body for the HandleReplaceBatch function.
type ReplaceBatchBody map[string]any

// HandleReplaceBatch - Handler for batch modifying values in the cache.
func HandleReplaceBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var body ReplaceBatchBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		if err := cache.ReplaceBatch(c.Request().Context(), body); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, WebError(err))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}
