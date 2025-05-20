package v1

import (
	"github.com/goodblaster/map-cache/internal/api/v1/docs"
	_ "github.com/goodblaster/map-cache/internal/api/v1/docs"
	"github.com/goodblaster/map-cache/internal/api/v1/v1caches"
	"github.com/goodblaster/map-cache/internal/api/v1/v1commands"
	"github.com/goodblaster/map-cache/internal/api/v1/v1keys"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupRoutes(e *echo.Echo) {
	e.Pre(middleware.RemoveTrailingSlash())
	v1 := e.Group("/api/v1")

	docs.SetupRoutes(v1)
	v1keys.SetupRoutes(v1)
	v1caches.SetupRoutes(v1)
	v1commands.SetupRoutes(v1)
}
