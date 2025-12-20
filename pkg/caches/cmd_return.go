package caches

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
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
		keyExpr := strings.TrimSpace(s[matches[0][2]:matches[0][3]])

		// Check for fallback syntax: key || default
		if strings.Contains(keyExpr, "||") {
			return evaluateWithFallback(ctx, cache, keyExpr)
		}

		if strings.Contains(keyExpr, "*") {
			// Wildcard expression
			keys := cache.cmap.WildKeys(ctx, keyExpr)
			var results []any
			for _, k := range keys {
				val, err := cache.Get(ctx, k)
				if err != nil {
					return nil, ErrWildcardInterpolation.Format(k, err)
				}
				results = append(results, val)
			}
			return results, nil
		}

		// Non-wildcard direct fetch
		val, err := cache.Get(ctx, keyExpr)
		if err != nil {
			return nil, ErrInterpolation.Format(keyExpr, err)
		}
		return val, nil
	}

	// Partial interpolations (template-style string)
	var builder strings.Builder
	lastIndex := 0

	for _, match := range matches {
		start, end := match[0], match[1]
		keyStart, keyEnd := match[2], match[3]
		keyExpr := strings.TrimSpace(s[keyStart:keyEnd])

		if strings.Contains(keyExpr, "*") {
			return nil, ErrWildcardInTemplate.Format(keyExpr)
		}

		// Append literal before interpolation
		builder.WriteString(s[lastIndex:start])

		var val any
		var err error

		// Check for fallback syntax in template
		if strings.Contains(keyExpr, "||") {
			val, err = evaluateWithFallback(ctx, cache, keyExpr)
		} else {
			val, err = cache.Get(ctx, keyExpr)
		}

		if err != nil {
			return nil, ErrInterpolation.Format(keyExpr, err)
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

// evaluateWithFallback handles "key || default" syntax
// The key is tried first, if it doesn't exist, the default literal value is returned
func evaluateWithFallback(ctx context.Context, cache *Cache, expr string) (any, error) {
	parts := strings.Split(expr, "||")

	// Must have exactly 2 parts: key || default
	if len(parts) != 2 {
		return nil, ErrInvalidFallbackExpression.Format(len(parts), expr)
	}

	keyPart := strings.TrimSpace(parts[0])
	defaultPart := strings.TrimSpace(parts[1])

	// Disallow wildcards with fallback
	if strings.Contains(keyPart, "*") {
		return nil, ErrWildcardWithFallback.Format(keyPart)
	}

	// Try to get the key from cache
	val, err := cache.Get(ctx, keyPart)
	if err == nil {
		return val, nil
	}

	// Key doesn't exist, return the default literal value
	return parseLiteral(defaultPart), nil
}

// parseLiteral converts a string to its appropriate type
func parseLiteral(s string) any {
	s = strings.TrimSpace(s)

	// Try boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Try null
	if s == "null" || s == "nil" {
		return nil
	}

	// Try integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Try quoted string (remove quotes)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}

	// Return as-is (unquoted string literal)
	return s
}
