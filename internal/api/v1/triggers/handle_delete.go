package triggers

import (
	"net/http"

	"github.com/goodblaster/errors"
	v1errors "github.com/goodblaster/map-cache/internal/api/v1/errors"
	"github.com/labstack/echo/v4"
)

// handleDeleteTrigger deletes a trigger by id.
func handleDeleteTrigger() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		id := c.Param("id")
		if id == "" {
			return v1errors.ApiError(c, http.StatusBadRequest, "missing trigger id")
		}

		cache := Cache(c)
		if err := cache.DeleteTrigger(ctx, id); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not delete trigger"))
		}

		return c.NoContent(http.StatusOK)
	}
}
