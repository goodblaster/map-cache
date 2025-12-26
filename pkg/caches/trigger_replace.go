package caches

import (
	"context"
)

func (cache *Cache) ReplaceTrigger(ctx context.Context, id string, newTrigger Trigger) error {
	for k, v := range cache.triggers {
		for i, t := range v {
			if t.Id == id {
				// Replace the trigger at the same index
				cache.triggers[k][i] = newTrigger
				return nil
			}
		}
	}
	return ErrTriggerNotFound
}
