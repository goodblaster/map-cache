package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_Get(t *testing.T) {
	ctx := context.Background()
	cache := NewTestCache(ctx, t)

	value, err := cache.Get(ctx, "key1")
	if assert.NoError(t, err, "Failed to fetch value from cache") {
		assert.EqualValues(t, "value1", value)
	}

	value, err = cache.Get(ctx, "key4/1")
	if assert.NoError(t, err, "Failed to fetch slice value from cache") {
		assert.EqualValues(t, "item2", value)
	}

	value, err = cache.Get(ctx, "key3/innerKey1")
	if assert.NoError(t, err, "Failed to fetch nested value from cache") {
		assert.EqualValues(t, "innerValue1", value)
	}
}

func TestCache_BatchGet(t *testing.T) {
	ctx := context.Background()
	cache := NewTestCache(ctx, t)

	batch, err := cache.BatchGet(ctx, "key1")
	if assert.NoError(t, err, "Failed to fetch batch from cache") {
		assert.EqualValues(t, "value1", batch["key1"], "Expected value1, got %v", batch)
	}

	batch, err = cache.BatchGet(ctx, "key3/innerKey1")
	if assert.NoError(t, err, "Failed to fetch nested value from cache") {
		assert.EqualValues(t, "innerValue1", batch["key3/innerKey1"])
	}

	batch, err = cache.BatchGet(ctx, "key4/1")
	if assert.NoError(t, err, "Failed to fetch slice value from cache") {
		assert.EqualValues(t, "item2", batch["key4/1"])
	}

	// Test for non-existing key
	batch, err = cache.BatchGet(ctx, "nonExistingKey")
	if assert.NoError(t, err, "Failed to fetch non-existing key from cache") {
		assert.Empty(t, batch, "Expected empty batch for non-existing key")
	}

	// Test for multiple keys
	batch, err = cache.BatchGet(ctx, "key1", "key2")
	if assert.NoError(t, err, "Failed to fetch multiple keys from cache") {
		assert.EqualValues(t, "value1", batch["key1"], "Expected value1, got %v", batch)
		assert.EqualValues(t, "value2", batch["key2"], "Expected value2, got %v", batch)
	}

	// Test for multiple keys with one non-existing key
	batch, err = cache.BatchGet(ctx, "key1", "nonExistingKey")
	if assert.NoError(t, err, "Failed to fetch multiple keys from cache") {
		assert.EqualValues(t, "value1", batch["key1"], "Expected value1, got %v", batch)
		assert.Empty(t, batch["nonExistingKey"], "Expected empty value for non-existing key")
	}
}
