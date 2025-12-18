package caches

import (
	"context"

	"github.com/goodblaster/errors"
)

// Replace - Replace single value in the cache.
func (cache *Cache) Replace(ctx context.Context, key string, value any) error {
	key = substituteContextVars(ctx, key)

	// Check key first. Error if does not exist.
	oldValue, err := cache.cmap.Get(ctx, SplitKey(key)...)
	_ = oldValue
	if err != nil {
		return ErrKeyNotFound.Format(key)
	}

	// Now set the value.
	if err := cache.cmap.Set(ctx, value, SplitKey(key)...); err != nil {
		return errors.Wrap(err, "could not set value")
	}

	// Fire triggers - return error if trigger execution fails (including infinite loops)
	if err := cache.OnChange(ctx, key, oldValue, value); err != nil {
		return errors.Wrap(err, "trigger execution failed")
	}

	return nil
}

// ReplaceBatch - Replace multiple, existing values in the cache.
// Each key in the values map is a path to a value in the cache (/a/b/c).
//
// TODO: This function does NOT fire triggers (OnChange) for performance reasons.
// This creates an inconsistency with Replace() which does fire triggers.
// Consider: Should batch operations fire triggers? If yes, implement it.
// If no, document this behavior clearly in API documentation.
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
