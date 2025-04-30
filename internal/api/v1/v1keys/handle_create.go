package v1keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// createKeysRequest is the request body for creating new cache entries.
type createKeysRequest struct {
	Entries map[string]any `json:"entries"`
}

// Validate - Validates the createKeysRequest.
func (req createKeysRequest) Validate() error { return nil }

// handleCreate creates new entries in a cache.
//
// @Summary Create cache entries
// @Description Creates one or more keys in the cache with values
// @Tags cache
// @Accept json
// @Produce json
// @Param  body  body  createKeysRequest  true  "Request body"
// @Success 201 {string} string "Created"
// @Failure 400 {object} echo.HTTPError "Invalid request body"
// @Failure 409 {object} echo.HTTPError "Cache key already exists"
// @Failure 500 {object} echo.HTTPError "Internal server error"
// @Router /api/v1/keys [post]
func handleCreate() echo.HandlerFunc {
	return func(c echo.Context) error {
		var body createKeysRequest
		if err := c.Bind(&body); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid json payload"))
		}

		if err := c.Validate(&body); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request body"))
		}

		cache := Cache(c)
		if err := cache.Create(c.Request().Context(), body.Entries); err != nil {
			if errors.Is(err, caches.ErrKeyAlreadyExists) {
				return v1errors.ApiError(c, http.StatusConflict, errors.Wrap(err, "V1Keys(s) already exists"))
			}
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "failed to create V1Keys"))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusCreated)
	}
}
