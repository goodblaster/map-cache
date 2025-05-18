package caches

import (
	"context"
)

type CommandNoop struct{}

func (CommandNoop) Type() string {
	return "NOOP"
}

func NOOP() Command {
	return CommandNoop{}
}

func (p CommandNoop) Do(ctx context.Context, cache *Cache) CmdResult {
	return CmdResult{}
}
