package caches

import "context"

type CommandReplace struct {
	key string
	val any
}

func REPLACE(key string, value any) Command {
	return CommandReplace{key: key, val: value}
}

func (p CommandReplace) Do(ctx context.Context, cache *Cache) CmdResult {
	return CmdResult{Error: cache.Replace(ctx, p.key, p.val)}
}
