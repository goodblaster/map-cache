package caches

import (
	"context"
	"strings"
)

type CommandGet struct {
	Key string `json:"key,required"`
}

func (CommandGet) Type() CommandType {
	return CommandTypeGet
}

func GET(key string) Command {
	return CommandGet{Key: key}
}

func (p CommandGet) Do(ctx context.Context, cache *Cache) CmdResult {
	key := p.Key

	if !strings.Contains(key, "*") {
		val, err := cache.Get(ctx, key)
		if err != nil {
			return CmdResult{Error: err}
		}
		return CmdResult{Value: val}
	}

	// Wildcard path
	matchingKeys := cache.cmap.WildKeys(ctx, key)
	values := make(map[string]any, len(matchingKeys))

	for _, k := range matchingKeys {
		val, err := cache.Get(ctx, k)
		if err != nil {
			return CmdResult{Error: err}
		}
		values[k] = val
	}

	return CmdResult{Value: values}
}
