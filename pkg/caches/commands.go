package caches

import (
	"context"
)

type CmdResult struct {
	Error  error
	Values []any
}

type Command interface {
	Do(ctx context.Context, cache *Cache) CmdResult
	Type() string
}

type CommandGroup struct {
	actions []Command
}

func (CommandGroup) Type() string {
	return "COMMANDS"
}

func COMMANDS(actions ...Command) Command {
	return CommandGroup{actions: actions}
}

func (p CommandGroup) Do(ctx context.Context, cache *Cache) CmdResult {
	var res CmdResult
	for _, action := range p.actions {
		actionRes := action.Do(ctx, cache)
		if actionRes.Error != nil {
			return actionRes
		}
		if actionRes.Values != nil {
			res.Values = append(res.Values, actionRes.Values)
		}
	}
	return res
}
