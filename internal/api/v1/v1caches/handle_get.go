package v1caches

import (
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// handleGetCacheList returns a list of all active caches.
//
// @Summary List all caches
// @Description Returns a list of currently registered cache names.
// @Tags caches
// @Produce json
// @Success 200 {array} string "List of cache names"
// @Router /caches [get]
func handleGetCacheList() echo.HandlerFunc {
	return func(c echo.Context) error {
		list := caches.List()
		return c.JSON(http.StatusOK, list)
	}
}
