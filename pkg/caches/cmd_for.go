package caches

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type CommandFor struct {
	LoopExpr string    `json:"loop_expr,required"`
	Commands []Command `json:"commands,required"`
}

func (CommandFor) Type() CommandType {
	return CommandTypeFor
}

func FOR(loopExpr string, cmds ...Command) Command {
	return CommandFor{LoopExpr: loopExpr, Commands: cmds}
}

func (f CommandFor) Do(ctx context.Context, cache *Cache) CmdResult {
	// Extract pattern like ${{job-1234/domains/*/countdown}}
	re := regexp.MustCompile(`\${{\s*([^}]+?)\s*}}`)
	match := re.FindStringSubmatch(f.LoopExpr)
	if len(match) < 2 {
		return CmdResult{Error: fmt.Errorf("invalid FOR expression: %s", f.LoopExpr)}
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

		for _, cmd := range f.Commands {
			// Replace ${{1}}, ${{2}}, ... with the captured fragments
			transformed := transformCommand(cmd, submatches[1:])

			result := transformed.Do(ctx, cache)
			allResults = append(allResults, result)

			if result.Error != nil {
				return result // stop on first error
			}
		}
	}

	return CmdResult{
		Value: allResults,
	}
}

func transformCommand(cmd Command, captures []string) Command {
	if cmd == nil {
		return nil
	}

	// Convert pointer to value if needed
	val := reflect.ValueOf(cmd)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		cmd = val.Interface().(Command)
	}

	switch c := cmd.(type) {
	case CommandIf:
		return &CommandIf{
			Condition: substituteCaptures(c.Condition, captures),
			IfTrue:    transformCommand(c.IfTrue, captures),
			IfFalse:   transformCommand(c.IfFalse, captures),
		}
	case CommandGet:
		return &CommandGet{Key: c.Key}
	case CommandReplace:
		return &CommandReplace{
			Key:   substituteCaptures(c.Key, captures),
			Value: c.Value,
		}
	case CommandInc:
		return &CommandInc{
			Key:   substituteCaptures(c.Key, captures),
			Value: c.Value,
		}
	default:
		return cmd
	}
}

func substituteCaptures(s string, captures []string) string {
	for i, val := range captures {
		placeholder := fmt.Sprintf("${{%d}}", i+1)
		s = strings.ReplaceAll(s, placeholder, val)
	}
	return s
}

func (c *CommandFor) UnmarshalJSON(data []byte) error {
	type Alias CommandFor
	aux := struct {
		Commands []json.RawMessage `json:"commands"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, cmdData := range aux.Commands {
		var rc RawCommand
		if err := json.Unmarshal(cmdData, &rc); err != nil {
			return err
		}
		c.Commands = append(c.Commands, rc.Command)
	}

	return nil
}
