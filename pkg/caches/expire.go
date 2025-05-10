package caches

import (
	"context"
	"time"
)

// SetKeyTTL - set the expiration timer for a key.
func (cache *Cache) SetKeyTTL(ctx context.Context, key string, after time.Duration) error {
	// Cancel existing timer if it exists.
	timer, ok := cache.keyExps[key]
	if ok {
		timer.Stop()
		delete(cache.keyExps, key)
	}

	// Create a new timer.
	cache.keyExps[key] = time.AfterFunc(after, func() {
		_ = cache.Delete(ctx, key)
		delete(cache.keyExps, key)
	})

	return nil
}

// CancelKeyTTL - cancel the expiration timer.
func (cache *Cache) CancelKeyTTL(ctx context.Context, key string) error {
	if timer, ok := cache.keyExps[key]; ok {
		timer.Stop()
		delete(cache.keyExps, key)
	}
	return nil
}
