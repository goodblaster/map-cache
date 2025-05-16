package caches

import (
	"sync"

	"github.com/goodblaster/errors"
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
	_, exists := caches.Load(name)
	if exists {
		return errors.New("cache already exists")
	}

	cache := New()
	caches.Store(name, cache)
	return nil
}

// FetchCache - do not automatically acquire the cache lock.
func FetchCache(name string) (*Cache, error) {
	val, exists := caches.Load(name)
	if !exists {
		return nil, errors.New("cache not found")
	}

	cache := val.(*Cache)
	return cache, nil
}

// DeleteCache - delete the cache.
func DeleteCache(name string) error {
	val, exists := caches.Load(name)
	if !exists {
		return ErrKeyNotFound.Format(name)
	}

	cache := val.(*Cache)

	// Clear TTL timers for all keys.
	for _, timer := range cache.keyExps {
		timer.Stop()
	}

	// And the cache ttl
	if cache.exp != nil {
		cache.exp.Stop()
	}

	caches.Delete(name)
	return nil
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

// Clear - Clear the cache. Must already be acquired.
func (cache *Cache) Clear() {
	cache.cmap = containers.NewGabsMap()
}
