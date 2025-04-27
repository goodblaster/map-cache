package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CreateKeysBody map[string]any

// HandleCreateKeys - Handler for creating new keys in the cache.
func HandleCreateKeys() echo.HandlerFunc {
	return func(c echo.Context) error {
		var body CreateKeysBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		cache := Cache(c)
		if err := cache.Create(c.Request().Context(), body); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, WebError(err))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusCreated)
	}
}
