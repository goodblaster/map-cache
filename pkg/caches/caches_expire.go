package caches

import (
	"github.com/goodblaster/errors"
)

var ErrCannotExpireDefaultCache = errors.New("cannot set expiration for default cache")

// SetCacheTTL - set the expiration timer for the cache.
func SetCacheTTL(name string, milliseconds int64) error {
	if name == DefaultName {
		return ErrCannotExpireDefaultCache
	}

	cache, err := FetchCache(name)
	if err != nil {
		return err
	}

	if cache.exp != nil {
		cache.exp.Stop()
	}

	cache.exp = FutureFunc(milliseconds, func() {
		caches.Delete(name)
	})
	cache.expMillis = &milliseconds

	return nil
}

// CancelCacheExpiration - cancel the expiration timer.
func CancelCacheExpiration(name string) error {
	if name == DefaultName {
		return ErrCannotExpireDefaultCache
	}

	cache, err := FetchCache(name)
	if err != nil {
		return err
	}

	if cache.exp != nil {
		cache.exp.Stop()
		cache.exp = nil
		cache.expMillis = nil
	}

	return nil
}
