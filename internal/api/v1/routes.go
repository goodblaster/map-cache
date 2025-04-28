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

func SetupCacheRoutes(group *echo.Group) {
	// cache handlers
	group = group.Group("/cache", cacheMW)

	// --- Create keys ---
	group.POST("", HandleCreateKeys())

	// --- Read keys ---
	group.GET("/:key", HandleGetValue()) // Get single key
	group.POST("/get", HandleGetBatch()) // Get multiple keys (batch)

	// --- Update keys ---
	group.PUT("/:key", HandlePut())     // Full replace single
	group.PUT("", HandleReplaceBatch()) // Full replace batch

	// ---
	//group.PATCH("/:key", HandlePatch()) // Partial update single
	//group.PATCH("", HandlePatchBatch()) // Partial update batch

	// --- Delete keys ---
	group.DELETE("/:key", HandleDelete())      // Delete single key
	group.POST("/delete", HandleDeleteBatch()) // Delete batch (POST because DELETE doesn't accept bodies cleanly)
}

func SetupCachesRoutes(group *echo.Group) {
	gCaches := group.Group("/caches")

	// Get cache name list
	gCaches.GET("", HandleGetCacheList())

	// Create a cache
	gCaches.POST("", HandleCreateCache())

	// Delete a cache
	gCaches.DELETE("/:name", HandleDeleteCache())
}

func SetupTriggerRoutes(v1 *echo.Group) {

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
			return ApiError(c, http.StatusFailedDependency, "cache not found")
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
