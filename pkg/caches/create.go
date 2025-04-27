package caches

import (
	"context"

	"github.com/goodblaster/map-cache/internal/mapkeys"
)

// Create - Create root-level keys/value pairs.
func (cache *Cache) Create(ctx context.Context, values map[string]any) error {
	// Check all keys first.
	// Keys must be a single path segment.
	// And the keys must not already exist.
	for key := range values {
		path := mapkeys.Split(key)
		if len(path) != 1 {
			return ErrSinglePathKeyRequired
		}

		if cache.Map.Exists(path[0]) {
			return ErrKeyAlreadyExists
		}
	}

	// Now set the values.
	for key, value := range values {
		_, err := cache.Map.Set(value, mapkeys.Split(key)...)
		if err != nil {
			return err // todo wrap
		}
	}

	return nil
}
