package caches

import (
	"context"
	"slices"
)

func (cache *Cache) DeleteTrigger(ctx context.Context, id string) error {
	for k, v := range cache.triggers {
		cache.triggers[k] = slices.DeleteFunc(v, func(t Trigger) bool {
			return t.Id == id
		})
	}
	return nil
}
