package v1

import (
	"github.com/goodblaster/map-cache/internal/api/v1/caches"
	"github.com/goodblaster/map-cache/internal/api/v1/commands"
	"github.com/goodblaster/map-cache/internal/api/v1/docs"
	"github.com/goodblaster/map-cache/internal/api/v1/keys"
	"github.com/goodblaster/map-cache/internal/api/v1/triggers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(e *echo.Echo) {
	e.Pre(middleware.RemoveTrailingSlash())

	// Prometheus metrics endpoint (global, not under /api/v1)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	v1 := e.Group("/api/v1")

	// Apply metrics middleware to track all v1 API requests
	v1.Use(MetricsMiddleware)

	// Apply timing middleware to all v1 API operations
	v1.Use(TimingMiddleware)

	docs.SetupRoutes(v1)
	keys.SetupRoutes(v1)
	caches.SetupRoutes(v1)
	commands.SetupRoutes(v1)
	triggers.SetupRoutes(v1)
}
