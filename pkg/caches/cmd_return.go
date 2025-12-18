package caches

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type CommandReturn struct {
	Key any `json:"key,required"`
}

func (CommandReturn) Type() CommandType {
	return CommandTypeReturn
}

func RETURN(key any) Command {
	return CommandReturn{Key: key}
}

func (p CommandReturn) Do(ctx context.Context, cache *Cache) CmdResult {
	switch str := p.Key.(type) {
	case string:
		resolvedVal, err := evaluateInterpolations(ctx, cache, str)
		if err != nil {
			return CmdResult{Error: err}
		}
		return CmdResult{Value: resolvedVal}
	default:
		return CmdResult{Value: p.Key}
	}
}

func evaluateInterpolations(ctx context.Context, cache *Cache, s string) (any, error) {
	re := regexp.MustCompile(`\${{\s*([^}]+?)\s*}}`)
	matches := re.FindAllStringSubmatchIndex(s, -1)

	if len(matches) == 0 {
		// No interpolations
		return s, nil
	}

	if len(matches) == 1 && matches[0][0] == 0 && matches[0][1] == len(s) {
		// Entire string is a single interpolation
		key := strings.TrimSpace(s[matches[0][2]:matches[0][3]])
		if strings.Contains(key, "*") {
			// Wildcard expression
			keys := cache.cmap.WildKeys(ctx, key)
			var results []any
			for _, k := range keys {
				val, err := cache.Get(ctx, k)
				if err != nil {
					return nil, fmt.Errorf("wildcard interpolation error for key %q: %w", k, err)
				}
				results = append(results, val)
			}
			return results, nil
		}

		// Non-wildcard direct fetch
		val, err := cache.Get(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("interpolation error for key %q: %w", key, err)
		}
		return val, nil
	}

	// Partial interpolations (template-style string)
	var builder strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start, end := match[0], match[1]
		keyStart, keyEnd := match[2], match[3]
		key := strings.TrimSpace(s[keyStart:keyEnd])

		if strings.Contains(key, "*") {
			return nil, fmt.Errorf("wildcards not allowed in templated string: %q", key)
		}

		// Append literal before interpolation
		builder.WriteString(s[lastIndex:start])

		val, err := cache.Get(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("interpolation error for key %q: %w", key, err)
		}
		// Optimize: avoid fmt.Sprintf if value is already a string
		if str, ok := val.(string); ok {
			builder.WriteString(str)
		} else {
			builder.WriteString(fmt.Sprintf("%v", val))
		}

		lastIndex = end
	}

	// Append the rest of the string
	builder.WriteString(s[lastIndex:])

	return builder.String(), nil
}
