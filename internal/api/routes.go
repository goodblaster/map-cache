package api

import (
	"net/http"

	"github.com/goodblaster/map-cache/internal/api/v1"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	groupV1 := e.Group("/api/v1")
	groupV1.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	V1 := v1.V1{}

	// V1 cache handlers
	cacheV1 := groupV1.Group("/cache", cacheMW)

	// --- Create keys ---
	cacheV1.POST("", V1.HandleCreateKeys())

	// --- Read keys ---
	cacheV1.GET("/:key", V1.HandleGetValue()) // Get single key
	cacheV1.POST("/get", V1.HandleGetBatch()) // Get multiple keys (batch)

	// --- Update keys ---
	cacheV1.PUT("/:key", V1.HandlePut()) // Full replace single
	cacheV1.PUT("", V1.HandlePutBatch()) // Full replace batch

	// ---
	//cacheV1.PATCH("/:key", V1.HandlePatch()) // Partial update single
	//cacheV1.PATCH("", V1.HandlePatchBatch()) // Partial update batch

	// --- Delete keys ---
	cacheV1.DELETE("/:key", V1.HandleDelete())      // Delete single key
	cacheV1.POST("/delete", V1.HandleDeleteBatch()) // Delete batch (POST because DELETE doesn't accept bodies cleanly)

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
			return echo.NewHTTPError(http.StatusFailedDependency, "cache not found") // todo better understanding of this error
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
