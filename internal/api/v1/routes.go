package v1

import (
	"net/http"
	"path"

	_ "github.com/goodblaster/map-cache/internal/api/v1/docs"
	"github.com/goodblaster/map-cache/internal/api/v1/v1caches"
	"github.com/goodblaster/map-cache/internal/api/v1/v1commands"
	"github.com/goodblaster/map-cache/internal/api/v1/v1keys"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo) {
	e.Pre(middleware.RemoveTrailingSlash())
	v1 := e.Group("/api/v1")
	v1.GET("", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, path.Join(c.Request().URL.Path, "/index.html"))
	})
	v1.GET("/*", echoSwagger.WrapHandler)

	//v1.GET("", func(c echo.Context) error {
	//	return c.String(http.StatusOK, "OK")
	//})

	v1keys.SetupRoutes(v1)
	v1caches.SetupRoutes(v1)
	v1commands.SetupRoutes(v1)
}
