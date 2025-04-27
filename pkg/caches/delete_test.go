package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_Delete(t *testing.T) {
	cache := New()
	ctx := context.Background()

	// Populate the cache with some data
	err := cache.Create(ctx, map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
	assert.NoError(t, err)

	// Make sure the data is there
	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Delete a key
	err = cache.Delete(ctx, "key1")
	assert.NoError(t, err)

	// Check that the key is gone
	value, err = cache.Get(ctx, "key1")
	assert.Error(t, err)

	// Deleting a non-existent key should not return an error
	err = cache.Delete(ctx, "nonExistentKey")
	assert.NoError(t, err)
}
