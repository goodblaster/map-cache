package caches

import (
	"sync"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type Cache struct {
	Map   *gabs.Container
	mutex *sync.Mutex
	tag   *string     // who owns this
	exp   *time.Timer // expiration timer
}

func New() *Cache {
	return &Cache{
		Map:   gabs.New(),
		mutex: &sync.Mutex{},
	}
}
