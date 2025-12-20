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
	// Extract pattern like ${{job-1234/domains/*/countdown}} using shared regex
	match := InterpolationPattern.FindStringSubmatch(f.LoopExpr)
	if len(match) < 2 {
		return CmdResult{Error: ErrInvalidForExpression.Format(f.LoopExpr)}
	}

	keyPattern := match[1]
	// TrimSpace only if needed
	if len(keyPattern) > 0 && (keyPattern[0] == ' ' || keyPattern[len(keyPattern)-1] == ' ' || keyPattern[0] == '\t') {
		keyPattern = strings.TrimSpace(keyPattern)
	}

	// Check for wildcard (avoid strings.Contains)
	hasWildcard := false
	for i := 0; i < len(keyPattern); i++ {
		if keyPattern[i] == '*' {
			hasWildcard = true
			break
		}
	}
	if !hasWildcard {
		return CmdResult{Error: ErrForExpressionNeedsWildcard.Format(keyPattern)}
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
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return CmdResult{Error: err}
		}

		submatches := keyRegex.FindStringSubmatch(key)
		if len(submatches) != starCount+1 {
			// No match or incorrect group count
			continue
		}

		for _, cmd := range f.Commands {
			// Check for context cancellation
			if err := ctx.Err(); err != nil {
				return CmdResult{Error: err}
			}

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
		return &CommandGet{
			Key: substituteCaptures(c.Key, captures),
		}
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
	case CommandDelete:
		return &CommandDelete{
			Key: substituteCaptures(c.Key, captures),
		}
	case CommandPrint:
		transformed := CommandPrint{Messages: make([]string, len(c.Messages))}
		for i, msg := range c.Messages {
			transformed.Messages[i] = substituteCaptures(msg, captures)
		}
		return &transformed
	case CommandReturn:
		if str, ok := c.Key.(string); ok {
			return &CommandReturn{Key: substituteCaptures(str, captures)}
		}
		return &c
	case CommandFor:
		transformed := CommandFor{
			LoopExpr: substituteCaptures(c.LoopExpr, captures),
			Commands: make([]Command, len(c.Commands)),
		}
		for i, cmd := range c.Commands {
			transformed.Commands[i] = transformCommand(cmd, captures)
		}
		return &transformed
	case CommandGroup:
		transformed := CommandGroup{actions: make([]Command, len(c.actions))}
		for i, action := range c.actions {
			transformed.actions[i] = transformCommand(action, captures)
		}
		return &transformed
	case CommandNoop:
		return &c
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
