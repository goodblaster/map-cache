package caches

import (
	"context"
)

// Get - Get one specific value from the cache.
func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	cache.recordActivity()
	return cache.cmap.Get(ctx, SplitKey(key)...)
}

// BatchGet - BatchGet values from the cache.
func (cache *Cache) BatchGet(ctx context.Context, keys ...string) ([]any, error) {
	cache.recordActivity()
	var vals []any

	for _, key := range keys {
		// Use cmap.Get directly to avoid double-counting activity
		val, err := cache.cmap.Get(ctx, SplitKey(key)...)
		if err != nil {
			return vals, err
		}
		vals = append(vals, val)
	}

	return vals, nil
}
