package v1

import (
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// HandleGetCacheList - Handler for getting list of all caches.
func HandleGetCacheList() echo.HandlerFunc {
	return func(c echo.Context) error {
		list := caches.List()
		return c.JSON(http.StatusOK, list)
	}
}
