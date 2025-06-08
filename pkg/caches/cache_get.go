package caches

import (
	"context"
)

// Get - Get one specific value from the cache.
func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	return cache.cmap.Get(ctx, SplitKey(key)...)
}

// BatchGet - BatchGet values from the cache.
func (cache *Cache) BatchGet(ctx context.Context, keys ...string) ([]any, error) {
	var vals []any

	for _, key := range keys {
		val, err := cache.Get(ctx, key)
		if err != nil {
			return vals, err
		}
		vals = append(vals, val)
	}

	return vals, nil
}
