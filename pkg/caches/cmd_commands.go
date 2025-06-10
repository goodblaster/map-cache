package caches

import (
	"context"
)

type CmdResult struct {
	Error error
	Value any
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
	var resValues []any

	for _, action := range p.actions {
		actionRes := action.Do(ctx, cache)
		if actionRes.Error != nil {
			return actionRes
		}
		resValues = append(resValues, actionRes.Value)
	}
	res.Value = resValues
	return res
}
