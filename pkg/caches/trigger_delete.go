package caches

import (
	"context"
	"slices"
)

func (cache *Cache) DeleteTrigger(ctx context.Context, id string) error {
	for k, v := range cache.triggers {
		remaining := slices.DeleteFunc(v, func(t Trigger) bool {
			return t.Id == id
		})
		// Drop patterns with no triggers left so OnChange doesn't keep checking them.
		if len(remaining) == 0 {
			delete(cache.triggers, k)
		} else {
			cache.triggers[k] = remaining
		}
	}
	return nil
}
