package triggers

import "github.com/labstack/echo/v4"

func SetupRoutes(group *echo.Group) {
	triggers := group.Group("/triggers", cacheMW)

	// Create trigger(s)
	triggers.POST("", handleCreateTrigger())

	// Delete trigger
	triggers.DELETE(":id", handleDeleteTrigger())

	// Replace trigger
	triggers.PUT(":id", handleReplaceTrigger())
}
