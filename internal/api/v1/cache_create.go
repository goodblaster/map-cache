package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HandlePostBody map[string]any

// HandleCreateKeys - Handler for creating new keys in the cache.
func (v V1) HandleCreateKeys() echo.HandlerFunc {
	return func(c echo.Context) error {
		var body HandlePostBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		cache := v.Cache(c)

		// TODO: System to filter error to a WebError?
		if err := cache.Create(c.Request().Context(), body); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, v.WebError(err))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusCreated)
	}
}
