package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_Replace(t *testing.T) {
	ctx := context.Background()
	cache := NewTestCache(ctx, t)

	// Replace a value in the cache
	err := cache.Replace(ctx, "key5", 67890)
	if assert.NoError(t, err) {
		value, err := cache.Get(ctx, "key5")
		if assert.NoError(t, err) {
			assert.Equal(t, 67890, value)
		}
	}

	// Replace a value in a slice in the cache
	err = cache.Replace(ctx, "key4/1", "newItem2")
	if assert.NoError(t, err) {
		value, err := cache.Get(ctx, "key4/1")
		if assert.NoError(t, err) {
			assert.Equal(t, "newItem2", value)
		}
	}

	// Replace a value in a nested map in the cache
	err = cache.Replace(ctx, "key3/innerKey1", "newInnerValue1")
	if assert.NoError(t, err) {
		value, err := cache.Get(ctx, "key3/innerKey1")
		if assert.NoError(t, err) {
			assert.Equal(t, "newInnerValue1", value)
		}
	}
}

func TestCache_ReplaceBatch(t *testing.T) {
	ctx := context.Background()
	cache := NewTestCache(ctx, t)

	// Replace multiple values in the cache
	err := cache.ReplaceBatch(ctx, map[string]any{
		"key5":           67890,
		"key4/1":         "newItem2",
		"key3/innerKey1": "newInnerValue1",
	})

	if assert.NoError(t, err) {
		value, err := cache.Get(ctx, "key5")
		if assert.NoError(t, err) {
			assert.Equal(t, 67890, value)
		}

		value, err = cache.Get(ctx, "key4/1")
		if assert.NoError(t, err) {
			assert.Equal(t, "newItem2", value)
		}

		value, err = cache.Get(ctx, "key3/innerKey1")
		if assert.NoError(t, err) {
			assert.Equal(t, "newInnerValue1", value)
		}
	}
}
