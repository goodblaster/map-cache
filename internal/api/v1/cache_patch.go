package v1

import (
	"github.com/labstack/echo/v4"
)

// HandlePatch - Handler for patching the cache.
func (v V1) HandlePatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		//cache := v.Cache(c)
		return nil
		//var patches []caches.Patch
		//if err := json.NewDecoder(c.Request().Body).Decode(&patches); err != nil {
		//	return c.JSON(http.StatusBadRequest, "Invalid request body")
		//}
		//
		//// TODO: System to filter error to a WebError?
		//if err := cache.Patch(c.Request().Context(), patches...); err != nil {
		//	return c.JSON(http.StatusInternalServerError, v.WebError(err))
		//}
		//
		//// Triggers?
		////
		//
		//return c.NoContent(http.StatusOK)
	}
}
