package v1

import (
	"encoding/json"
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type DeleteCacheBody struct {
	Name string `json:"name"`
}

// HandleDeleteCache - Handler for deleting a cache.
func HandleDeleteCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		var body DeleteCacheBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return ApiError(c, http.StatusBadRequest, "invalid request body")
		}

		err := caches.DeleteCache(body.Name)
		if err != nil {
			return ApiError(c, http.StatusNotFound, "could not find cache")
		}

		return c.NoContent(http.StatusOK)
	}
}
