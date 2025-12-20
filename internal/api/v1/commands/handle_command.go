package commands

import (
	"context"
	"net/http"
	"time"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/config"
	v1errors "github.com/goodblaster/map-cache/internal/api/v1/errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type commandRequest struct {
	Commands []caches.RawCommand `json:"commands"`
}

func (req commandRequest) Validate() error {
	if len(req.Commands) == 0 {
		return errors.New("at least one command is required")
	}
	return nil
}

func handleCommand() echo.HandlerFunc {
	return func(c echo.Context) error {
		var input commandRequest
		if err := c.Bind(&input); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := input.Validate(); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, err)
		}

		var cmds []caches.Command
		for _, rawCommand := range input.Commands {
			cmds = append(cmds, rawCommand.Command)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request().Context(), time.Duration(config.CommandTimeoutMs)*time.Millisecond)
		defer cancel()

		cache := Cache(c)
		result := cache.Execute(ctx, cmds...)

		// Check if execution timed out
		if ctx.Err() == context.DeadlineExceeded {
			return v1errors.ApiError(c, http.StatusRequestTimeout, "command execution timed out")
		}

		if result.Error != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, result.Error)
		}

		return c.JSON(http.StatusOK, result.Value)
	}
}
