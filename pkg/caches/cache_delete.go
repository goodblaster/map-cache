package caches

import (
	"context"
	"strconv"
	"strings"

	"github.com/goodblaster/logos"
)

func (cache *Cache) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		path := SplitKey(key)

		// If the last value of the key is a number, assume it is an array element.
		// We need to handle those differently.
		if len(path) > 1 {
			lastStr := path[len(path)-1]
			if i, err := strconv.Atoi(lastStr); err == nil {
				if err := cache.cmap.ArrayRemove(ctx, i, path[:len(path)-1]...); err != nil {
					// Log but don't fail - deletion is best-effort for array elements
					logos.WithError(err).Warnf("failed to remove array element at index %d for key %s", i, key)
				}
				continue
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
			logos.WithError(err).Warnf("failed to delete key %s", key)
		}
	}
	return nil
}
