package caches

import (
	"context"

	"github.com/goodblaster/errors"
)

// Increment - Increment single value in the cache.
func (cache *Cache) Increment(ctx context.Context, key string, value any) error {
	// Check key first. Error if does not exist.
	oldValue, err := cache.cmap.Get(ctx, SplitKey(key)...)

	if err != nil {
		return ErrKeyNotFound.Format(key)
	}

	f64, ok := ToFloat64(oldValue)
	if !ok {
		return errors.New("not a number")
	}

	inc, ok := ToFloat64(value)
	if !ok {
		return errors.New("increment value must be a number")
	}

	f64 += inc

	// Now set the value.
	if err := cache.cmap.Set(ctx, f64, SplitKey(key)...); err != nil {
		return errors.Wrap(err, "could not set value")
	}

	return nil
}
