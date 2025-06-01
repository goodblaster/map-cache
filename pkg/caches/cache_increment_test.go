package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_Increment(t *testing.T) {
	ctx := context.Background()
	cache := New()

	err := cache.Create(ctx, map[string]any{
		"key1": 10.0,
	})
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Increment a value in the cache
	newValue, err := cache.Increment(ctx, "key1", 5.0)
	if assert.NoError(t, err) {
		assert.Equal(t, 15.0, newValue)

		// Verify the incremented value
		value, err := cache.Get(ctx, "key1")
		if assert.NoError(t, err) {
			assert.Equal(t, 15.0, value)
		}
	}
}
