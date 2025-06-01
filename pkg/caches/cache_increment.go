package caches

import (
	"context"

	"github.com/goodblaster/errors"
)

// Increment - Increment single value in the cache.
func (cache *Cache) Increment(ctx context.Context, key string, value any) (float64, error) {
	// Check key first. Error if does not exist.
	oldValue, err := cache.cmap.Get(ctx, SplitKey(key)...)

	if err != nil {
		return 0, ErrKeyNotFound.Format(key)
	}

	f64, ok := ToFloat64(oldValue)
	if !ok {
		return 0, errors.New("not a number")
	}

	inc, ok := ToFloat64(value)
	if !ok {
		return 0, errors.New("increment value must be a number")
	}

	f64 += inc

	// Now set the value.
	err = cache.Replace(ctx, key, f64)
	return f64, err
}
