package caches

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type CommandReturn struct {
	Values []any `json:"values,required"`
}

func (CommandReturn) Type() string {
	return "RETURN"
}

func RETURN(values ...any) Command {
	return CommandReturn{Values: values}
}

func (p CommandReturn) Do(ctx context.Context, cache *Cache) CmdResult {
	resolved := make([]any, len(p.Values))

	for i, val := range p.Values {
		switch str := val.(type) {
		case string:
			resolvedStr, err := evaluateInterpolations(str, cache, ctx)
			if err != nil {
				return CmdResult{Error: err}
			}
			resolved[i] = resolvedStr
		default:
			resolved[i] = val
		}
	}

	return CmdResult{Values: resolved}
}

func evaluateInterpolations(s string, cache *Cache, ctx context.Context) (string, error) {
	re := regexp.MustCompile(`\${{\s*([^}]+?)\s*}}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		key := strings.TrimSpace(re.FindStringSubmatch(match)[1])
		val, err := cache.Get(ctx, key)
		if err != nil {
			return match // leave as-is or optionally return error
		}
		return fmt.Sprintf("%v", val)
	}), nil
}
