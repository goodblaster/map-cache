package caches

import (
	"sync"

	"github.com/goodblaster/map-cache/pkg/containers"
)

type Cache struct {
	cmap    containers.Map
	mutex   *sync.Mutex
	tag     *string           // who owns this
	exp     *Timer            // expiration timer
	keyExps map[string]*Timer // key-based expiration timers
}

func New() *Cache {
	return &Cache{
		cmap:    containers.NewGabsMap(),
		mutex:   &sync.Mutex{},
		keyExps: map[string]*Timer{},
	}
}
