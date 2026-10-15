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

func TestCache_Delete_NumericKeyInObject(t *testing.T) {
	cache := New()
	defer cache.Close()
	ctx := context.Background()

	err := cache.Create(ctx, map[string]any{
		"users": map[string]any{
			"123": map[string]any{"name": "alice"},
			"456": map[string]any{"name": "bob"},
		},
		"list": []any{"a", "b", "c"},
	})
	assert.NoError(t, err)

	// Numeric segment under an object is a map key, not an array index.
	assert.NoError(t, cache.Delete(ctx, "users/123"))
	_, err = cache.Get(ctx, "users/123")
	assert.Error(t, err)
	_, err = cache.Get(ctx, "users/456")
	assert.NoError(t, err)

	// Numeric segment under an array still removes the element.
	assert.NoError(t, cache.Delete(ctx, "list/1"))
	val, err := cache.Get(ctx, "list")
	assert.NoError(t, err)
	assert.Equal(t, []any{"a", "c"}, val)
}
