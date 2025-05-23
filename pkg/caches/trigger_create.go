package caches

import (
	"context"

	"github.com/google/uuid"
)

func (cache *Cache) CreateTrigger(ctx context.Context, key string, command Command) (string, error) {
	trigger := Trigger{
		Id:      uuid.New().String(),
		Key:     key,
		Command: command,
	}

	cache.triggers[key] = append(cache.triggers[key], trigger)
	return trigger.Id, nil
}
