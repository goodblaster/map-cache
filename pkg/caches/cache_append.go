package caches

import (
	"context"
	"reflect"
)

// ArrayAppend - Append entry to existing array.
func (cache *Cache) ArrayAppend(ctx context.Context, key string, value any) error {
	// Make sure the array exists.
	path := SplitKey(key)
	val, err := cache.cmap.Get(ctx, path...)
	if err != nil {
		return ErrKeyNotFound.Format(path)
	}

	if reflect.TypeOf(val).Kind() != reflect.Slice {
		return ErrNotAnArray.Format(path)
	}

	if err := cache.cmap.ArrayAppend(ctx, value, path...); err != nil {
		return err
	}

	return nil
}
