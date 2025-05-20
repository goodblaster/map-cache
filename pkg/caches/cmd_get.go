package caches

import (
	"context"
	"strings"
)

type CommandGet struct {
	Keys []string `json:"keys,required"`
}

func (CommandGet) Type() CommandType {
	return CommandTypeGet
}

func GET(keys ...string) Command {
	return CommandGet{Keys: keys}
}

func (p CommandGet) Do(ctx context.Context, cache *Cache) CmdResult {
	var res CmdResult
	for _, key := range p.Keys {
		values := map[string]any{}
		if !strings.Contains(key, "*") {
			v, err := cache.Get(ctx, key)
			if err != nil {
				return CmdResult{Error: err}
			}
			values[key] = v
			res.Values = append(res.Values, values)
			continue
		}

		wildkeys := cache.cmap.WildKeys(ctx, key)
		for _, key := range wildkeys {
			v, err := cache.Get(ctx, key)
			if err != nil {
				return CmdResult{Error: err}
			}
			values[key] = v
		}
		res.Values = append(res.Values, values)
	}
	return res
}
