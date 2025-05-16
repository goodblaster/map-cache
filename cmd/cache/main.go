// @title Web Cache API
// @version 1.0
// @description API for managing web cache keys
// @BasePath /api/v1
package main

import (
	"github.com/goodblaster/map-cache/internal/api/admin"
	v1 "github.com/goodblaster/map-cache/internal/api/v1"
	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

func main() {
	config.Init()
	err := caches.AddCache(caches.DefaultName)
	if err != nil {
		panic(err)
	}

	e := echo.New()
	v1.SetupRoutes(e)
	admin.SetupRoutes(e)
	_ = e.Start(":8080")
}
