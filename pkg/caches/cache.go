package caches

import (
	"sync"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type Cache struct {
	Map     *gabs.Container
	mutex   *sync.Mutex
	tag     *string                // who owns this
	exp     *time.Timer            // expiration timer
	keyExps map[string]*time.Timer // key-based expiration timers
}

func New() *Cache {
	return &Cache{
		Map:     gabs.New(),
		mutex:   &sync.Mutex{},
		keyExps: map[string]*time.Timer{},
	}
}
