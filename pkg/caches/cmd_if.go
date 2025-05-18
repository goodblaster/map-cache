package caches

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

type CommandIf struct {
	condition string
	ifTrue    Command
	ifFalse   Command
}

func IF(condition string, ifTrue, ifFalse Command) Command {
	return CommandIf{condition: condition, ifTrue: ifTrue, ifFalse: ifFalse}
}

func (p CommandIf) Do(ctx context.Context, cache *Cache) CmdResult {
	parameters := map[string]any{}
	conditionExpr := p.condition

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
		return p.ifTrue.Do(ctx, cache)
	}
	return p.ifFalse.Do(ctx, cache)
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
