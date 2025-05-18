package caches

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type CommandFor struct {
	loopExpr string
	cmds     []Command
}

// todo: commandS
func FOR(loopExpr string, cmds ...Command) Command {
	return CommandFor{loopExpr: loopExpr, cmds: cmds}
}

func (f CommandFor) Do(ctx context.Context, cache *Cache) CmdResult {
	// Extract pattern like ${{job-1234/domains/*/countdown}}
	re := regexp.MustCompile(`\${{\s*([^}]+?)\s*}}`)
	match := re.FindStringSubmatch(f.loopExpr)
	if len(match) < 2 {
		return CmdResult{Error: fmt.Errorf("invalid FOR expression: %s", f.loopExpr)}
	}

	keyPattern := strings.TrimSpace(match[1])
	if !strings.Contains(keyPattern, "*") {
		return CmdResult{Error: fmt.Errorf("FOR expression must include a wildcard: %s", keyPattern)}
	}

	// Build a regex from the wildcard pattern
	regexPattern := regexp.QuoteMeta(keyPattern)
	starCount := strings.Count(keyPattern, "*")
	for i := 0; i < starCount; i++ {
		regexPattern = strings.Replace(regexPattern, "\\*", "([^/]+)", 1)
	}
	keyRegex := regexp.MustCompile("^" + regexPattern + "$")

	// Resolve keys
	keys := cache.cmap.WildKeys(ctx, keyPattern)
	var allResults []CmdResult

	for _, key := range keys {
		submatches := keyRegex.FindStringSubmatch(key)
		if len(submatches) != starCount+1 {
			// No match or incorrect group count
			continue
		}

		for _, cmd := range f.cmds {
			// Replace ${{1}}, ${{2}}, ... with the captured fragments
			transformed := transformCommand(cmd, submatches[1:])

			result := transformed.Do(ctx, cache)
			allResults = append(allResults, result)

			if result.Error != nil {
				return result // stop on first error
			}
		}
	}

	var cmdResult CmdResult
	for _, res := range allResults {
		if res.Values != nil {
			cmdResult.Values = append(cmdResult.Values, res.Values)
		}
	}
	return cmdResult
}

func transformCommand(cmd Command, captures []string) Command {
	switch c := cmd.(type) {
	case CommandIf:
		return CommandIf{
			condition: substituteCaptures(c.condition, captures),
			ifTrue:    transformCommand(c.ifTrue, captures),
			ifFalse:   transformCommand(c.ifFalse, captures),
		}
	case CommandGet:
		keys := make([]string, len(c.keys))
		for i, k := range c.keys {
			keys[i] = substituteCaptures(k, captures)
		}
		return CommandGet{keys: keys}
	case CommandReplace:
		return CommandReplace{
			key: substituteCaptures(c.key, captures),
			val: c.val,
		}
	case CommandInc:
		return CommandInc{
			key: substituteCaptures(c.key, captures),
			val: c.val,
		}
	default:
		return cmd // unknown command type
	}
}

func substituteCaptures(s string, captures []string) string {
	for i, val := range captures {
		placeholder := fmt.Sprintf("${{%d}}", i+1)
		s = strings.ReplaceAll(s, placeholder, val)
	}
	return s
}
