package caches

import "context"

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
	if err := cache.Replace(ctx, p.Key, p.Value); err != nil {
		return CmdResult{Error: err}
	}
	return CmdResult{Value: p.Value}
}
