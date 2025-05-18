package caches

import "context"

type CommandReplace struct {
	Key   string `json:"key,required"`
	Value any    `json:"value,required"`
}

func (CommandReplace) Type() string {
	return "REPLACE"
}

func REPLACE(key string, value any) Command {
	return CommandReplace{Key: key, Value: value}
}

func (p CommandReplace) Do(ctx context.Context, cache *Cache) CmdResult {
	return CmdResult{Error: cache.Replace(ctx, p.Key, p.Value)}
}
