package triggers

import (
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func cacheMW(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check headers for cache name
		cacheName := c.Request().Header.Get("X-Cache-Name")
		if cacheName == "" {
			cacheName = caches.DefaultName
		}

		// Make sure it exists
		cache, err := caches.FetchCache(cacheName)
		if err != nil {
			return echo.NewHTTPError(http.StatusFailedDependency, "cache not found").SetInternal(err)
		}

		// Generate a request ID and set it in the context
		requestId := uuid.New().String()
		c.Set("request_id", requestId)

		// Acquire the cache for this request
		cache.Acquire(requestId)
		defer cache.Release(requestId)

		// Set the cache in the context
		c.Set("cache", cache)
		return next(c)
	}
}
