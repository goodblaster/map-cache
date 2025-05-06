package caches

import (
	"sync"

	"github.com/Jeffail/gabs/v2"
)

type Cache struct {
	Map   *gabs.Container
	mutex *sync.Mutex
	tag   *string // who owns this
}

func New() *Cache {
	return &Cache{
		Map:   gabs.New(),
		mutex: &sync.Mutex{},
	}
}
