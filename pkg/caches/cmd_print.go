package caches

import (
	"context"
	"fmt"

	"github.com/goodblaster/map-cache/internal/log"
)

type CommandPrint struct {
	Messages []string `json:"messages,required"`
}

func (CommandPrint) Type() CommandType {
	return CommandTypePrint
}

func PRINT(msgs ...string) Command {
	return CommandPrint{Messages: msgs}
}

func (p CommandPrint) Do(ctx context.Context, cache *Cache) CmdResult {
	var res CmdResult
	var resValues []any
	for _, msg := range p.Messages {
		msg, keys := ExtractAndReplaceParams(msg)
		var params []any

		for _, key := range keys {
			v, err := cache.Get(ctx, key)
			if err != nil {
				return CmdResult{Error: ErrKeyNotFound.Format(key)}
			}
			params = append(params, v)
		}

		formatted := fmt.Sprintf(msg, params...)
		log.Print(formatted)
		resValues = append(resValues, formatted)
	}
	res.Value = resValues
	return res
}

// ExtractAndReplaceParams - Handle ${{var}} using shared regex
func ExtractAndReplaceParams(input string) (string, []string) {
	var params []string

	// Use shared pre-compiled regex
	result := InterpolationPattern.ReplaceAllStringFunc(input, func(m string) string {
		submatch := InterpolationPattern.FindStringSubmatch(m)
		if len(submatch) > 1 {
			params = append(params, submatch[1])
		}
		return "%v"
	})

	return result, params
}
