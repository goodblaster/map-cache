package caches

import (
	"context"

	"github.com/goodblaster/errors"
)

// Replace - Replace single value in the cache.
func (cache *Cache) Replace(ctx context.Context, key string, value any) error {
	// Check key first. Error if does not exist.
	if !cache.cmap.Exists(ctx, SplitKey(key)...) {
		return ErrKeyNotFound.Format(key)
	}

	// Now set the value.
	if err := cache.cmap.Set(ctx, value, SplitKey(key)...); err != nil {
		return errors.Wrap(err, "could not set value")
	}

	return nil
}

// ReplaceBatch - Replace multiple, existing values in the cache.
// Each key in the values map is a path to a value in the cache (/a/b/c).
func (cache *Cache) ReplaceBatch(ctx context.Context, values map[string]any) error {
	// Check all keys first. Error if any do not exist.
	for key := range values {
		if !cache.cmap.Exists(ctx, SplitKey(key)...) {
			return ErrKeyNotFound.Format(key)
		}
	}

	// Now set the values.
	for key, value := range values {
		if err := cache.cmap.Set(ctx, value, SplitKey(key)...); err != nil {
			return errors.Wrap(err, "could not set value")
		}
	}

	return nil
}
