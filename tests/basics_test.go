package tests

import (
	"context"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/stretchr/testify/assert"
)

func Test_Basic(t *testing.T) {
	name := "test"
	if err := caches.AddCache(name); err != nil {
		t.Fatalf("Failed to add cache: %v", err)
	}

	cache, err := caches.FetchCache(name)
	if err != nil {
		t.Fatalf("Failed to fetch cache: %v", err)
	}

	cache.Acquire("test_basic")
	defer cache.Release("test_basic")

	// Initialize the cache with some data
	m := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": map[string]any{
			"innerKey1": "innerValue1",
			"innerKey2": "innerValue2",
			"outerKey1": "outerValue1",
		},
		"key4": []any{"item1", "item2", "item3"},
		"key5": 12345,
	}

	ctx := context.Background()
	err = cache.Create(ctx, m)
	assert.NoError(t, err, "Failed to replace cache")

	// Get a value from the cache
	value, err := cache.Get(ctx, "key5")
	if assert.NoError(t, err, "Failed to fetch value from cache") {
		assert.EqualValues(t, 12345, value)
	}

	// Get a value from a slice in the cache
	value, err = cache.Get(ctx, "key4/1")
	if assert.NoError(t, err, "Failed to fetch slice value from cache") {
		assert.EqualValues(t, "item2", value)
	}

	// Get a value from a nested map in the cache
	value, err = cache.Get(ctx, "key3/innerKey1")
	if assert.NoError(t, err, "Failed to fetch nested value from cache") {
		assert.EqualValues(t, "innerValue1", value)
	}

	// BatchGet a value from the cache
	batch, err := cache.BatchGet(ctx, "key1")
	if assert.NoError(t, err, "Failed to fetch batch from cache") {
		assert.EqualValues(t, "value1", batch["key1"], "Expected value1, got %v", batch)
	}

	// BatchGet a nested batch from the cache
	batch, err = cache.BatchGet(ctx, "key3/innerKey1")
	if assert.NoError(t, err, "Failed to fetch nested value from cache") {
		assert.EqualValues(t, "innerValue1", batch["key3/innerKey1"])
	}

	// BatchGet a value from a slice in the cache
	batch, err = cache.BatchGet(ctx, "key4/1")
	if assert.NoError(t, err, "Failed to fetch slice value from cache") {
		assert.EqualValues(t, "item2", batch["key4/1"])
	}

	// Change one value
	err = cache.Replace(ctx, "key3/outerKey1", "newOuterValue1")
	if assert.NoError(t, err, "Failed to replace cache") {
		// Verify the change
		value, err = cache.Get(ctx, "key3/outerKey1")
		if assert.NoError(t, err, "Failed to fetch value from cache") {
			assert.EqualValues(t, "newOuterValue1", value)
		}
	}

	// Change some values
	err = cache.ReplaceBatch(ctx, map[string]any{
		"key1":           "newValue1",
		"key3/innerKey1": "newInnerValue1",
		"key4/1":         "newItem2",
	})

	if assert.NoError(t, err, "Failed to replace cache") {
		// Verify the changes
		value, err = cache.Get(ctx, "key1")
		if assert.NoError(t, err, "Failed to fetch value from cache") {
			assert.EqualValues(t, "newValue1", value)
		}

		value, err = cache.Get(ctx, "key3/innerKey1")
		if assert.NoError(t, err, "Failed to fetch nested value from cache") {
			assert.EqualValues(t, "newInnerValue1", value)
		}

		value, err = cache.Get(ctx, "key4/1")
		if assert.NoError(t, err, "Failed to fetch slice value from cache") {
			assert.EqualValues(t, "newItem2", value)
		}
	}

	// Delete a value from the cache
	err = cache.Delete(ctx, "key1")
	if assert.NoError(t, err, "Failed to delete value from cache") {
		// Verify the deletion
		_, err = cache.Get(ctx, "key1")
		assert.Error(t, err, "Expected error when fetching deleted key")
	}
}
