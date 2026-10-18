package caches

import (
	"context"
	"slices"
)

func (cache *Cache) ReplaceTrigger(ctx context.Context, id string, newTrigger Trigger) error {
	for k, v := range cache.triggers {
		for i, t := range v {
			if t.Id != id {
				continue
			}

			if newTrigger.Key == k {
				// Replace the trigger at the same index
				cache.triggers[k][i] = newTrigger
				return nil
			}

			// The pattern changed: move the trigger so it fires on the new key.
			remaining := slices.Delete(v, i, i+1)
			if len(remaining) == 0 {
				delete(cache.triggers, k)
			} else {
				cache.triggers[k] = remaining
			}
			cache.triggers[newTrigger.Key] = append(cache.triggers[newTrigger.Key], newTrigger)
			return nil
		}
	}
	return ErrTriggerNotFound
}
