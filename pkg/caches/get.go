package caches

import (
	"context"
)

// Get - Get one specific value from the cache.
func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	c := cache.Map.Search(SplitKey(key)...)
	if c == nil {
		return nil, ErrKeyNotFound
	}

	return c.Data(), nil
}

// BatchGet - BatchGet values from the cache.
func (cache *Cache) BatchGet(ctx context.Context, keys ...string) (map[string]any, error) {
	vals := map[string]any{}

	for _, key := range keys {
		c := cache.Map.Search(SplitKey(key)...)
		if c == nil {
			continue // todo error on an failure? return what we CAN find?
		}
		vals[key] = c.Data()
	}

	return vals, nil
}
