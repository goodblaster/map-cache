package caches

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

type CommandIf struct {
	Condition string  `json:"condition,required"`
	IfTrue    Command `json:"if_true,required"`
	IfFalse   Command `json:"if_false,required"`
}

func (CommandIf) Type() CommandType {
	return CommandTypeIf
}

func IF(condition string, ifTrue, ifFalse Command) Command {
	return CommandIf{Condition: condition, IfTrue: ifTrue, IfFalse: ifFalse}
}

func (p CommandIf) Do(ctx context.Context, cache *Cache) CmdResult {
	parameters := map[string]any{}
	conditionExpr := p.Condition

	// Handle any(...) or all(...) first
	conditionExpr, err := expandAnyAll(conditionExpr, cache, parameters, ctx)
	if err != nil {
		return CmdResult{Error: err}
	}

	// Now handle remaining simple ${{...}} references
	re := regexp.MustCompile(`\${{\s*([^}]+?)\s*}}`)
	matches := re.FindAllStringSubmatch(conditionExpr, -1)
	for _, match := range matches {
		fullMatch := match[0]
		key := strings.TrimSpace(match[1])

		val, err := cache.Get(ctx, key)
		if err != nil {
			val = nil
		}
		varName := keyToIdentifier(key)
		parameters[varName] = val
		conditionExpr = strings.ReplaceAll(conditionExpr, fullMatch, varName)
	}

	expr, err := govaluate.NewEvaluableExpression(conditionExpr)
	if err != nil {
		return CmdResult{Error: fmt.Errorf("invalid expression: %w", err)}
	}

	result, err := expr.Evaluate(parameters)
	if err != nil {
		return CmdResult{Error: fmt.Errorf("evaluation error: %w", err)}
	}

	isTrue, ok := result.(bool)
	if !ok {
		return CmdResult{Error: fmt.Errorf("expression did not return a boolean")}
	}

	if isTrue {
		return p.IfTrue.Do(ctx, cache)
	}
	return p.IfFalse.Do(ctx, cache)
}

func expandAnyAll(expr string, cache *Cache, parameters map[string]any, ctx context.Context) (string, error) {
	re := regexp.MustCompile(`\b(any|all)\(\s*\${{\s*([^}]+?)\s*}}\s*([!<>=]=?|==)\s*([^\)]+?)\s*\)`)

	return re.ReplaceAllStringFunc(expr, func(m string) string {
		matches := re.FindStringSubmatch(m)
		if len(matches) != 5 {
			return m
		}
		mode := matches[1] // "any" or "all"
		keyPattern := matches[2]
		op := matches[3]
		right := matches[4]

		keys := cache.cmap.WildKeys(ctx, keyPattern)
		if len(keys) == 0 {
			return "false" // or error
		}

		var parts []string
		for _, key := range keys {
			varName := keyToIdentifier(key)
			val, err := cache.Get(ctx, key)
			if err != nil {
				val = nil
			}
			parameters[varName] = val
			parts = append(parts, fmt.Sprintf("%s %s %s", varName, op, right))
		}

		join := " || "
		if mode == "all" {
			join = " && "
		}
		return "(" + strings.Join(parts, join) + ")"
	}), nil
}

func keyToIdentifier(key string) string {
	replacer := strings.NewReplacer(".", "_", "/", "_", "-", "_")
	return replacer.Replace(key)
}

func (c *CommandIf) UnmarshalJSON(data []byte) error {
	// Define an alias to avoid infinite recursion
	type Alias CommandIf
	aux := struct {
		IfTrue  json.RawMessage `json:"if_true"`
		IfFalse json.RawMessage `json:"if_false"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Use your existing RawCommand logic
	var ifTrue RawCommand
	if err := json.Unmarshal(aux.IfTrue, &ifTrue); err != nil {
		return err
	}
	c.IfTrue = ifTrue.Command

	var ifFalse RawCommand
	if err := json.Unmarshal(aux.IfFalse, &ifFalse); err != nil {
		return err
	}
	c.IfFalse = ifFalse.Command

	return nil
}
