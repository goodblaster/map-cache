package docs

import (
	"mime"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(v1 *echo.Group) {
	v1.GET("/openapi.yaml", func(c echo.Context) error {
		data, err := ApiFiles.ReadFile("openapi.yaml")
		if err != nil {
			return c.String(http.StatusInternalServerError, "spec not found")
		}
		return c.Blob(http.StatusOK, "application/yaml", data)
	})

	v1.GET("/*", func(c echo.Context) error {
		file := c.Param("*")
		if file == "" || file == "/" {
			file = "swagger-ui/index.html"
		} else {
			file = path.Join("swagger-ui", file)
		}

		data, err := ApiFiles.ReadFile(file)
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}

		ctype := mime.TypeByExtension(path.Ext(file))
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		return c.Blob(http.StatusOK, ctype, data)
	})

	v1.GET("", func(c echo.Context) error {
		data, err := ApiFiles.ReadFile("swagger-ui/index.html")
		if err != nil {
			return c.String(http.StatusInternalServerError, "index.html not found")
		}
		return c.Blob(http.StatusOK, "text/html", data)
	})

}
