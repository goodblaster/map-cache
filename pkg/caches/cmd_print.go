package caches

import (
	"context"
	"strings"

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
		// First, interpolate wildcard variables (e.g., ${{1}} → actual wildcard match)
		msg = substituteContextVars(ctx, msg)

		// Then replace key references with their values. The message is not
		// treated as a format string, so a literal % prints as-is.
		var missing error
		formatted := InterpolationPattern.ReplaceAllStringFunc(msg, func(m string) string {
			key := strings.TrimSpace(InterpolationPattern.FindStringSubmatch(m)[1])
			v, err := cache.Get(ctx, key)
			if err != nil {
				if missing == nil {
					missing = ErrKeyNotFound.Format(key)
				}
				return m
			}
			return FormatValue(v)
		})
		if missing != nil {
			return CmdResult{Error: missing}
		}

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
