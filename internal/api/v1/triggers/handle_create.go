package triggers

import (
	"net/http"

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
			return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON payload").SetInternal(err)
		}

		cache := Cache(c)
		id, err := cache.CreateTrigger(ctx, input.Key, input.Raw.Command)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to add trigger").SetInternal(err)
		}

		return c.JSON(http.StatusOK, id)
	}
}
