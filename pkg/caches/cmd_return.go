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

func (CommandReturn) Type() CommandType {
	return CommandTypeReturn
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
	matches := re.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return s, nil
	}

	var result strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start, end := match[0], match[1]
		keyStart, keyEnd := match[2], match[3]

		result.WriteString(s[lastIndex:start])

		key := strings.TrimSpace(s[keyStart:keyEnd])
		val, err := cache.Get(ctx, key)
		if err != nil {
			return "", fmt.Errorf("interpolation error for key %q: %w", key, err)
		}

		result.WriteString(fmt.Sprintf("%v", val))
		lastIndex = end
	}

	result.WriteString(s[lastIndex:])
	return result.String(), nil
}
