package caches

import "time"

// SetCacheExpiration - set the expiration timer for the cache.
func SetCacheExpiration(name string, after time.Duration) error {
	cache, err := FetchCache(name)
	if err != nil {
		return err
	}

	if cache.exp != nil {
		cache.exp.Stop()
	}

	cache.exp = time.AfterFunc(after, func() {
		caches.Delete(name)
	})

	return nil
}

// CancelCacheExpiration - cancel the expiration timer.
func CancelCacheExpiration(name string) error {
	cache, err := FetchCache(name)
	if err != nil {
		return err
	}

	if cache.exp != nil {
		cache.exp.Stop()
		cache.exp = nil
	}

	return nil
}
