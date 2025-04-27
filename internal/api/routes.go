package api

import (
	"github.com/goodblaster/map-cache/internal/api/v1"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	v1.SetupRoutes(e)
}
