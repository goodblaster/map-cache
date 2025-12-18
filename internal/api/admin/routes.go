package admin

import (
	_ "github.com/goodblaster/map-cache/internal/api/v1/docs"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupRoutes(e *echo.Echo) {
	e.Pre(middleware.RemoveTrailingSlash())
	admin := e.Group("/admin", adminMW)
	admin.POST("/backup", handleBackup)
	admin.POST("/restore", handleRestore)
}

// TODO: Implement admin authentication middleware
//
// This middleware currently allows unrestricted access to admin endpoints.
// Consider implementing one of these authentication methods:
//
// Option 1: API Key in header
//   - Check for "X-Admin-Key" header
//   - Compare against environment variable ADMIN_API_KEY
//
// Option 2: Basic Auth
//   - Use middleware.BasicAuth() with credentials from env vars
//
// Option 3: JWT tokens
//   - Validate JWT token from Authorization header
//
// For now, this is a no-op and allows all requests through.
// SECURITY RISK: Anyone can backup/restore caches without authentication.
func adminMW(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: Add authentication check here
		// Example:
		// apiKey := c.Request().Header.Get("X-Admin-Key")
		// if apiKey != os.Getenv("ADMIN_API_KEY") {
		//     return echo.NewHTTPError(http.StatusUnauthorized, "invalid admin key")
		// }
		return next(c)
	}
}
