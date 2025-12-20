package caches

import (
	"context"
	"strings"
)

type CommandDelete struct {
	Key string `json:"key,required"`
}

func (CommandDelete) Type() CommandType {
	return CommandTypeDelete
}

func DELETE(key string) Command {
	return CommandDelete{Key: key}
}

func (p CommandDelete) Do(ctx context.Context, cache *Cache) CmdResult {
	// Check if pattern contains wildcards
	if strings.Contains(p.Key, "*") {
		// Get all matching keys first (to return their values)
		keys := cache.cmap.WildKeys(ctx, p.Key)
		values := make([]any, 0, len(keys))

		for _, key := range keys {
			// Get value before deleting
			val, err := cache.Get(ctx, key)
			if err == nil {
				values = append(values, val)
			}
		}

		// Delete all matching keys
		if err := cache.Delete(ctx, keys...); err != nil {
			return CmdResult{Error: err}
		}

		return CmdResult{Value: values}
	}

	// Single key deletion
	val, err := cache.Get(ctx, p.Key)
	if err != nil {
		// Key doesn't exist - that's okay for delete
		// Return nil to indicate nothing was deleted
		return CmdResult{Value: nil}
	}

	if err := cache.Delete(ctx, p.Key); err != nil {
		return CmdResult{Error: err}
	}

	return CmdResult{Value: val}
}
