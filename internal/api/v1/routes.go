package v1

import (
	"github.com/goodblaster/map-cache/internal/api/v1/docs"
	"github.com/goodblaster/map-cache/internal/api/v1/caches"
	"github.com/goodblaster/map-cache/internal/api/v1/commands"
	"github.com/goodblaster/map-cache/internal/api/v1/keys"
	"github.com/goodblaster/map-cache/internal/api/v1/triggers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupRoutes(e *echo.Echo) {
	e.Pre(middleware.RemoveTrailingSlash())
	v1 := e.Group("/api/v1")

	docs.SetupRoutes(v1)
	keys.SetupRoutes(v1)
	caches.SetupRoutes(v1)
	commands.SetupRoutes(v1)
	triggers.SetupRoutes(v1)
}
