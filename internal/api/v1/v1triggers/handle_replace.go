package v1triggers

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// replaceTriggerRequest is for replacing a single trigger.
type replaceTriggerRequest struct {
	Id  string            `json:"id,required"`
	Key string            `json:"key,required"`
	Raw caches.RawCommand `json:"command,required"`
}

// handleDeleteCache deletes a trigger by id.
func handleReplaceTrigger() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		id := c.Param("id")
		if id == "" {
			return v1errors.ApiError(c, http.StatusBadRequest, "missing trigger id")
		}

		var input replaceTriggerRequest
		if err := c.Bind(&input); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid JSON payload"))
		}

		if input.Id != id {
			return v1errors.ApiError(c, http.StatusBadRequest, "payload id must match request id")
		}

		newTrigger := caches.Trigger{
			Id:      id,
			Key:     input.Key,
			Command: input.Raw.Command,
		}

		cache := Cache(c)
		if err := cache.ReplaceTrigger(ctx, id, newTrigger); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "failed to replace trigger"))
		}

		return c.NoContent(http.StatusNoContent)
	}
}
