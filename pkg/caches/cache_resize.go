package caches

import (
	"context"
)

// ArrayResize - Resize an existing array in the cache.
func (cache *Cache) ArrayResize(ctx context.Context, key string, newSize int) error {
	return cache.cmap.ArrayResize(ctx, newSize, SplitKey(key)...)
}
