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

// TODO - admin middleware
func adminMW(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
