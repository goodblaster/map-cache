package caches

import "context"

func (cache *Cache) Execute(ctx context.Context, commands ...Command) CmdResult {
	return COMMANDS(commands...).Do(ctx, cache)
}
