package caches

import (
	"sync"

	"github.com/goodblaster/map-cache/pkg/containers"
)

var caches = sync.Map{}

const DefaultName = "default"

func List() []string {
	var cacheNames []string
	caches.Range(func(key, value interface{}) bool {
		cacheNames = append(cacheNames, key.(string))
		return true
	})
	return cacheNames
}

func AddCache(name string) error {
	cache := New()
	if _, loaded := caches.LoadOrStore(name, cache); loaded {
		cache.Close()
		return ErrCacheAlreadyExists
	}
	return nil
}

// FetchCache - do not automatically acquire the cache lock.
func FetchCache(name string) (*Cache, error) {
	val, exists := caches.Load(name)
	if !exists {
		return nil, ErrCacheNotFound
	}

	cache := val.(*Cache)
	return cache, nil
}

// DeleteCache - delete the cache and stop its timers and background worker.
// The caller must not hold the cache lock.
func DeleteCache(name string) error {
	val, exists := caches.Load(name)
	if !exists {
		return ErrCacheNotFound
	}

	removeCache(name, val.(*Cache))
	return nil
}

// removeCache unregisters cache and shuts it down, but only if it is still the
// cache registered under name. This keeps a stale expiration timer from
// removing a newer cache that was created (or restored) with the same name.
func removeCache(name string, cache *Cache) {
	if caches.CompareAndDelete(name, cache) {
		cache.Close()
	}
}

// Acquire - Acquire the cache if you already have a reference to it.
func (cache *Cache) Acquire(tag string) *Cache {
	cache.mutex.Lock()
	if cache.tag != nil {
		panic("cache was improperly released")
	}
	cache.tag = &tag
	return cache
}

// Release - Release the cache if you acquired it.
func (cache *Cache) Release(tag string) {
	if cache.tag == nil {
		panic("releasing untagged cache")
	}
	if tag != *cache.tag {
		panic("releasing unowned cache")
	}
	cache.tag = nil
	cache.mutex.Unlock()
}

// Clear - Clear all data and key TTLs from the cache. Triggers are kept.
// Must already be acquired.
func (cache *Cache) Clear() {
	cache.recordActivity()
	cache.cmap = containers.NewGabsMap()

	// Stop key timers so they can't delete keys created after the clear.
	for _, timer := range cache.keyExps {
		timer.Stop()
	}
	cache.keyExps = map[string]*Timer{}
}
