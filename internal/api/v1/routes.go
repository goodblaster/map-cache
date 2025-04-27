package v1

import (
	"net/http"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupRoutes(e *echo.Echo) {

	e.Pre(middleware.RemoveTrailingSlash())
	v1 := e.Group("/api/v1")
	v1.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	SetupCacheRoutes(v1)
}

func SetupCacheRoutes(v1 *echo.Group) {
	// cache handlers
	gCache := v1.Group("/cache", cacheMW)

	// --- Create keys ---
	gCache.POST("", HandleCreateKeys())

	// --- Read keys ---
	gCache.GET("/:key", HandleGetValue()) // Get single key
	gCache.POST("/get", HandleGetBatch()) // Get multiple keys (batch)

	// --- Update keys ---
	gCache.PUT("/:key", HandlePut())     // Full replace single
	gCache.PUT("", HandleReplaceBatch()) // Full replace batch

	// ---
	//gCache.PATCH("/:key", HandlePatch()) // Partial update single
	//gCache.PATCH("", HandlePatchBatch()) // Partial update batch

	// --- Delete keys ---
	gCache.DELETE("/:key", HandleDelete())      // Delete single key
	gCache.POST("/delete", HandleDeleteBatch()) // Delete batch (POST because DELETE doesn't accept bodies cleanly)
}

func SetupCachesRoutes() {

}

func SetupTriggerRoutes() {

}

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
			return echo.NewHTTPError(http.StatusFailedDependency, "cache not found")
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
