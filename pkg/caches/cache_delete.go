package caches

import (
	"context"
	"strconv"
	"strings"

	"github.com/goodblaster/map-cache/internal/log"
)

func (cache *Cache) Delete(ctx context.Context, keys ...string) error {
	cache.recordActivity()

	for _, key := range keys {
		path := SplitKey(key)

		// If the last segment is a number and the parent is an array, it is an
		// array element and must be handled differently. Numeric keys inside
		// objects (e.g. users/123) fall through to a normal delete.
		if len(path) > 1 {
			lastStr := path[len(path)-1]
			if i, err := strconv.Atoi(lastStr); err == nil {
				parentPath := path[:len(path)-1]
				if parent, err := cache.cmap.Get(ctx, parentPath...); err == nil {
					if _, isArray := parent.([]any); isArray {
						if err := cache.cmap.ArrayRemove(ctx, i, parentPath...); err != nil {
							// Log but don't fail - deletion is best-effort for array elements
							log.WithError(err).With("key", key).With("index", i).Warn("failed to remove array element")
						}
						continue
					}
				}
			}
		}

		// Clear any TTLs that start with this key.
		for k, timer := range cache.keyExps {
			if strings.HasPrefix(k, key) {
				timer.Stop()
				delete(cache.keyExps, k)
			}
		}

		// Delete the key - log errors but don't fail (deletion is best-effort)
		if err := cache.cmap.Delete(ctx, path...); err != nil {
			log.WithError(err).With("key", key).Warn("failed to delete key")
		}
	}
	return nil
}
