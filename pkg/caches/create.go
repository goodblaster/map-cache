package caches

import (
	"context"

	"github.com/goodblaster/errors"
)

// Create - Create root-level keys/value pairs.
func (cache *Cache) Create(ctx context.Context, entries map[string]any) error {
	// Check all keys first.
	// Keys must be a single path segment.
	// And the keys must not already exist.
	for key := range entries {
		path := SplitKey(key)
		if len(path) != 1 {
			return ErrSinglePathKeyRequired.Format(key)
		}

		if cache.Map.Exists(path[0]) {
			return ErrKeyAlreadyExists.Format(path)
		}
	}

	// Now set the entries.
	for key, value := range entries {
		_, err := cache.Map.Set(value, SplitKey(key)...)
		if err != nil {
			return errors.Wrap(err, "could not set value")
		}
	}

	return nil
}
