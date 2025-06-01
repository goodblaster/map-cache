package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_ArrayAppend(t *testing.T) {
	ctx := context.Background()
	cache := New()

	err := cache.Create(ctx, map[string]any{
		"key1": []any{"item1", "item2", "item3"},
	})

	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	err = cache.ArrayAppend(ctx, "key1", "item4")
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	value, err := cache.Get(ctx, "key1")
	if assert.NoError(t, err) {
		assert.Equal(t, []any{"item1", "item2", "item3", "item4"}, value)
	}
}
