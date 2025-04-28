package v1

import (
	"github.com/labstack/echo/v4"
)

// HandlePatch - Handler for patching the cache.
func HandlePatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		//cache := Cache(c)
		return nil
		//var patches []caches.Patch
		//if err := json.NewDecoder(c.Request().Body).Decode(&patches); err != nil {
		//	return ApiError(c, http.StatusBadRequest, "invalid request body")
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
