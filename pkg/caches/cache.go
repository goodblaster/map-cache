package caches

import (
	"sync"

	"github.com/Jeffail/gabs/v2"
)

type Cache struct {
	Map     *gabs.Container
	mutex   *sync.Mutex
	tag     *string           // who owns this
	exp     *Timer            // expiration timer
	keyExps map[string]*Timer // key-based expiration timers
}

func New() *Cache {
	return &Cache{
		Map:     gabs.New(),
		mutex:   &sync.Mutex{},
		keyExps: map[string]*Timer{},
	}
}
