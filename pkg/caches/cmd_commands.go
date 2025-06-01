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
	Type() CommandType
	MarshalJSON() ([]byte, error)
}

type CommandGroup struct {
	actions []Command
}

type CommandType string

const (
	CommandTypeIf      CommandType = "IF"
	CommandTypeFor     CommandType = "FOR"
	CommandTypeReplace CommandType = "REPLACE"
	CommandTypeReturn  CommandType = "RETURN"
	CommandTypePrint   CommandType = "PRINT"
	CommandTypeGet     CommandType = "GET"
	CommandTypeInc     CommandType = "INC"
	CommandTypeNoop    CommandType = "NOOP"
	CommandTypeGroup   CommandType = "COMMANDS"
)

func (CommandGroup) Type() CommandType {
	return CommandTypeGroup
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
