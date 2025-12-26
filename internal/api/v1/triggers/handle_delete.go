package triggers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// handleDeleteTrigger deletes a trigger by id.
func handleDeleteTrigger() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		id := c.Param("id")
		if id == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing trigger id")
		}

		cache := Cache(c)
		if err := cache.DeleteTrigger(ctx, id); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "could not delete trigger").SetInternal(err)
		}

		return c.NoContent(http.StatusOK)
	}
}
