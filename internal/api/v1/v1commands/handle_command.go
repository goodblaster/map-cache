package v1commands

import (
	"net/http"

	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type commandRequest struct {
	Commands []caches.RawCommand `json:"commands"`
}

func (req commandRequest) Validate() error {
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

		cache := Cache(c)
		result := cache.Execute(c.Request().Context(), cmds...)

		if result.Error != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, result.Error)
		}

		return c.JSON(http.StatusOK, result.Value)
	}
}
