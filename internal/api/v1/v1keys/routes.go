package v1keys

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes(group *echo.Group) {
	// cache handlers
	group = group.Group("/keys", cacheMW)

	// --- Create keys ---
	group.POST("", handleCreate())

	// --- Read keys ---
	group.GET("/:key", handleGetValue()) // Get single key
	group.POST("/get", handleGetBatch()) // Get multiple keys (batch)

	// --- Update keys ---
	group.PUT("/:key", handlePut())     // Full replace single
	group.PUT("", handleReplaceBatch()) // Full replace batch

	// ---
	group.PATCH("/:key", handlePatch()) // Partial update single

	// --- Delete keys ---
	group.DELETE("/:key", handleDelete())      // Delete single key
	group.POST("/delete", handleDeleteBatch()) // Delete batch (POST because DELETE doesn't accept bodies cleanly)
}
