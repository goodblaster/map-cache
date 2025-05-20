package v1commands

import "github.com/labstack/echo/v4"

func SetupRoutes(group *echo.Group) {
	gCaches := group.Group("/commands", cacheMW)

	// Execute command(s).
	gCaches.POST("/execute", handleCommand())
}
