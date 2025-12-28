package caches

import (
	"context"
	"fmt"
	"strings"
)

type CommandReplace struct {
	Key   string `json:"key,required"`
	Value any    `json:"value,required"`
}

func (CommandReplace) Type() CommandType {
	return CommandTypeReplace
}

func REPLACE(key string, value any) Command {
	return CommandReplace{Key: key, Value: value}
}

func (p CommandReplace) Do(ctx context.Context, cache *Cache) CmdResult {
	// Interpolate wildcard variables in key (e.g., ${{1}} → actual wildcard match)
	key := p.Key
	if val := ctx.Value(triggerVarsContextKey); val != nil {
		if vars, ok := val.([]string); ok {
			for i, v := range vars {
				placeholder := fmt.Sprintf("${{%d}}", i+1)
				key = strings.ReplaceAll(key, placeholder, v)
			}
		}
	}

	if err := cache.Replace(ctx, key, p.Value); err != nil {
		return CmdResult{Error: err}
	}
	return CmdResult{Value: p.Value}
}
