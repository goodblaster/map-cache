package caches

import (
	"context"
	"strconv"
	"strings"

	"github.com/goodblaster/map-cache/internal/config"
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

		// Clear any TTLs for this key or keys nested beneath it.
		value, err := cache.cmap.Get(ctx, path...)
		cache.clearKeyTTLs(key, err == nil && isContainer(value))

		// Delete the key - log errors but don't fail (deletion is best-effort)
		if err := cache.cmap.Delete(ctx, path...); err != nil {
			log.WithError(err).With("key", key).Warn("failed to delete key")
		}
	}
	return nil
}

// isContainer reports whether keys can be nested beneath value.
func isContainer(value any) bool {
	switch value.(type) {
	case map[string]any, []any:
		return true
	}
	return false
}

// clearKeyTTLs stops and forgets the TTL timer for key. If nested is true it
// also clears timers for keys beneath it, which requires scanning every TTL -
// callers should only ask for that when key holds an object or array.
func (cache *Cache) clearKeyTTLs(key string, nested bool) {
	if len(cache.keyExps) == 0 {
		return
	}

	if timer, ok := cache.keyExps[key]; ok {
		timer.Stop()
		delete(cache.keyExps, key)
	}

	if !nested {
		return
	}

	prefix := key + config.KeyDelimiter
	for k, timer := range cache.keyExps {
		if strings.HasPrefix(k, prefix) {
			timer.Stop()
			delete(cache.keyExps, k)
		}
	}
}
