package caches

import (
	"context"
	"strings"
)

type CommandGet struct {
	keys []string
}

func GET(keys ...string) Command {
	return CommandGet{keys: keys}
}

func (p CommandGet) Do(ctx context.Context, cache *Cache) CmdResult {
	var res CmdResult
	for _, key := range p.keys {
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
