package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_Create(t *testing.T) {
	cache := New()
	ctx := context.Background()

	// Test creating a new key
	err := cache.Create(ctx, map[string]any{"key1": "value1"})
	assert.NoError(t, err)

	// Test creating a key that already exists
	err = cache.Create(ctx, map[string]any{"key1": "value2"})
	assert.Error(t, err)

	// Test creating a key with an invalid path
	err = cache.Create(ctx, map[string]any{"key1/key2": "value3"})
	assert.Error(t, err)
}
