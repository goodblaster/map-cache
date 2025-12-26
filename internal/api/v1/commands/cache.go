package commands

import (
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

func Cache(c echo.Context) *caches.Cache {
	value := c.Get("cache")
	if value == nil {
		panic("cache value is not set")
	}

	cache, ok := value.(*caches.Cache)
	if !ok {
		panic("cache value is not of type *caches.Cache")
	}

	return cache
}
