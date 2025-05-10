package caches

import (
	"context"
	"strconv"
	"strings"
)

func (cache *Cache) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		path := SplitKey(key)

		// If the last value of the key is a number, assume it is an array element.
		// We need to handle those differently.
		if len(path) > 1 {
			lastStr := path[len(path)-1]
			if i, err := strconv.Atoi(lastStr); err == nil {
				_ = cache.Map.ArrayRemove(i, path[:len(path)-1]...)
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

		_ = cache.Map.Delete(path...)
	}
	return nil
}
