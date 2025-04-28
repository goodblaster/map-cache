package v1

import (
	"encoding/json"
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type CreateKeysBody map[string]any

// HandleCreateKeys - Handler for creating new keys in the cache.
func HandleCreateKeys() echo.HandlerFunc {
	return func(c echo.Context) error {
		var body CreateKeysBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request body"))
		}

		cache := Cache(c)
		if err := cache.Create(c.Request().Context(), body); err != nil {
			if errors.Is(err, caches.ErrKeyAlreadyExists) {
				return ApiError(c, http.StatusConflict, errors.Wrap(err, "cache already exists"))
			}
			return ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "failed to create cache"))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusCreated)
	}
}
