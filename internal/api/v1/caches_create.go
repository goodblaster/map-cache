package v1

import (
	"encoding/json"
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type CreateCacheBody struct {
	Name       string `json:"name"`
	Expiration string `json:"expiration"`
}

// HandleCreateCache - Handler for creating new cache.
func HandleCreateCache() echo.HandlerFunc {
	return func(c echo.Context) error {
		var body CreateCacheBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return ApiError(c, http.StatusBadRequest, "invalid request body")
		}

		err := caches.AddCache(body.Name)
		if err != nil {
			return ApiError(c, http.StatusInternalServerError, "failed to create cache") // or StatusBadRequest? depends on error
		}

		// Expiration?
		//

		return c.NoContent(http.StatusCreated)
	}
}
