// @title Web Cache API
// @version 1.0
// @description API for managing web cache keys
// @BasePath /api/v1
package main

import (
	"github.com/goodblaster/logos"
	"github.com/goodblaster/map-cache/internal/api/admin"
	v1 "github.com/goodblaster/map-cache/internal/api/v1"
	"github.com/goodblaster/map-cache/internal/build"
	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.Init()

	err := caches.AddCache(caches.DefaultName)
	if err != nil {
		logos.WithError(err).Fatal("failed to add default cache")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	v1.SetupRoutes(e)
	admin.SetupRoutes(e)

	// Health check route
	e.GET("/status", func(c echo.Context) error {
		return c.JSON(200, map[string]any{
			"status": "ok",
			"build":  build.Info(),
		})
	})

	if err := e.Start(config.WebAddress); err != nil {
		logos.WithError(err).Fatal("failed to start web server")
	}
}
