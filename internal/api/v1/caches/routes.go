package caches

import "github.com/labstack/echo/v4"

func SetupRoutes(group *echo.Group) {
	gCaches := group.Group("/caches")

	// Get cache name list
	gCaches.GET("", handleGetCacheList())

	// Get cache statistics
	gCaches.GET("/stats", handleGetStats())

	// Create a cache
	gCaches.POST("", handleCreateCache())

	// Update cache expiration
	gCaches.PUT("/:name", handleUpdateCache())

	// Delete a cache
	gCaches.DELETE("/:name", handleDeleteCache())
}
