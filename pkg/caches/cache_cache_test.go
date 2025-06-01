package caches

import (
	"testing"

	"github.com/stretchr/testify/assert"
)
import "context"

func TestingMap() map[string]any {
	return map[string]any{
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
}

func NewTestCache(ctx context.Context, t *testing.T) *Cache {
	cache := New()
	err := cache.Create(ctx, TestingMap())
	if !assert.NoError(t, err) {
		t.Fatalf("Failed to create cache: %v", err)
	}
	return cache
}
