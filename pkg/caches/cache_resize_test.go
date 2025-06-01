package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_ArrayResize(t *testing.T) {
	cache := New()
	ctx := context.Background()

	// Create a cache with initial values
	initialData := map[string]any{
		"key1": []any{"item1", "item2", "item3"},
	}

	err := cache.Create(ctx, initialData)
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Check the size of the string array.
	value, err := cache.Get(ctx, "key1")
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to get key1: %v", err)
	}

	if !assert.IsType(t, []any{}, value) {
		t.Fatalf("Expected key1 to be a slice, got %T", value)
	}

	if assert.Len(t, value, 3) {
		assert.Equal(t, []any{"item1", "item2", "item3"}, value)
	}

	// Resize the array to a smaller size
	err = cache.ArrayResize(ctx, "key1", 2)
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to resize key1: %v", err)
	}

	// Check the resized array
	value, err = cache.Get(ctx, "key1")
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to get resized key1: %v", err)
	}

	if !assert.IsType(t, []any{}, value) {
		t.Fatalf("Expected key1 to be a slice, got %T", value)
	}

	if assert.Len(t, value, 2) {
		assert.Equal(t, []any{"item1", "item2"}, value)
	}

	// Resize the array to a larger size
	err = cache.ArrayResize(ctx, "key1", 5)
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to resize key1: %v", err)
	}

	// Check the resized array
	value, err = cache.Get(ctx, "key1")
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to get resized key1: %v", err)
	}

	if !assert.IsType(t, []any{}, value) {
		t.Fatalf("Expected key1 to be a slice, got %T", value)
	}

	if assert.Len(t, value, 5) {
		assert.Equal(t, []any{"item1", "item2", nil, nil, nil}, value)
	}

	// Resize the array to the same size
	err = cache.ArrayResize(ctx, "key1", 5)
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to resize key1: %v", err)
	}

	// Check the resized array
	value, err = cache.Get(ctx, "key1")
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to get resized key1: %v", err)
	}

	if !assert.IsType(t, []any{}, value) {
		t.Fatalf("Expected key1 to be a slice, got %T", value)
	}

	if assert.Len(t, value, 5) {
		assert.Equal(t, []any{"item1", "item2", nil, nil, nil}, value)
	}

	// Resize the array to a negative size
	err = cache.ArrayResize(ctx, "key1", -1)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "newSize cannot be negative")
	}

	// Check that the array is unchanged
	value, err = cache.Get(ctx, "key1")
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to get key1 after negative resize: %v", err)
	}

	if !assert.IsType(t, []any{}, value) {
		t.Fatalf("Expected key1 to be a slice, got %T", value)
	}

	if assert.Len(t, value, 5) {
		assert.Equal(t, []any{"item1", "item2", nil, nil, nil}, value)
	}
}
