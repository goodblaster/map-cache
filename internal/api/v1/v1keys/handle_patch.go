package v1keys

import (
	"github.com/labstack/echo/v4"
)

// handlePatch - Handler for patching the cache.
func handlePatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		//cache := Cache(c)
		return nil
		//var patches []caches.Patch
		//if err := c.Bind(&patches); err != nil {
		//	return ApiError(c, http.StatusBadRequest, "invalid json payload")
		//}
		//
		//if err := cache.Patch(c.Request().Context(), patches...); err != nil {
		//	return ApiError(c, http.StatusInternalServerError, v.Error(err))
		//}
		//
		//// Triggers?
		////
		//
		//return c.NoContent(http.StatusOK)
	}
}
