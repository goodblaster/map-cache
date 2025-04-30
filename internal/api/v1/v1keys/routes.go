package v1keys

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes(group *echo.Group) {
	// cache handlers
	group = group.Group("/keys", cacheMW)

	// --- Create V1Keys ---
	group.POST("", handleCreate())

	// --- Read V1Keys ---
	group.GET("/:key", handleGetValue()) // Get single key
	group.POST("/get", handleGetBatch()) // Get multiple v1Keys (batch)

	// --- Update V1Keys ---
	group.PUT("/:key", handlePut())     // Full replace single
	group.PUT("", handleReplaceBatch()) // Full replace batch

	// ---
	//group.PATCH("/:key", handlePatch()) // Partial update single
	//group.PATCH("", handlePatchBatch()) // Partial update batch

	// --- Delete V1Keys ---
	group.DELETE("/:key", handleDelete())      // Delete single key
	group.POST("/delete", handleDeleteBatch()) // Delete batch (POST because DELETE doesn't accept bodies cleanly)
}
