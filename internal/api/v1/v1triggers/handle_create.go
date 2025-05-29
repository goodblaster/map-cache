package v1triggers

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// CreateTriggerRequest is for adding a single trigger.
type CreateTriggerRequest struct {
	Key string            `json:"key,required"`
	Raw caches.RawCommand `json:"command,required"`
}

// handleCreateTrigger creates a new trigger based on key and command.
func handleCreateTrigger() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var input CreateTriggerRequest
		if err := c.Bind(&input); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid JSON payload"))
		}

		cache := Cache(c)
		id, err := cache.CreateTrigger(ctx, input.Key, input.Raw.Command)
		if err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "failed to add trigger"))
		}

		return c.JSON(http.StatusOK, id)
	}
}
