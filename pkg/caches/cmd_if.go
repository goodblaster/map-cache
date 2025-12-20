package caches

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Knetic/govaluate"
)

// keyIdentifierReplacer is reused for converting keys to identifiers
var keyIdentifierReplacer = strings.NewReplacer(".", "_", "/", "_", "-", "_")

// exprCache caches compiled expressions to avoid recompiling the same expression
var exprCache sync.Map // map[string]*govaluate.EvaluableExpression

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

	// Sub in contextual variables.
	conditionExpr = substituteContextVars(ctx, conditionExpr)

	// Handle any(...) or all(...) first
	conditionExpr, err := expandAnyAll(conditionExpr, cache, parameters, ctx)
	if err != nil {
		return CmdResult{Error: err}
	}

	// Now handle remaining simple ${{...}} references using shared regex
	matches := InterpolationPattern.FindAllStringSubmatch(conditionExpr, -1)
	for _, match := range matches {
		fullMatch := match[0]
		key := match[1]
		// TrimSpace only if needed
		if len(key) > 0 && (key[0] == ' ' || key[len(key)-1] == ' ' || key[0] == '\t') {
			key = strings.TrimSpace(key)
		}

		val, err := cache.Get(ctx, key)
		if err != nil {
			val = nil
		}
		varName := keyToIdentifier(key)
		parameters[varName] = val
		conditionExpr = strings.ReplaceAll(conditionExpr, fullMatch, varName)
	}

	// Check cache first
	var expr *govaluate.EvaluableExpression
	if cached, ok := exprCache.Load(conditionExpr); ok {
		expr = cached.(*govaluate.EvaluableExpression)
	} else {
		// Compile and cache
		var err error
		expr, err = govaluate.NewEvaluableExpression(conditionExpr)
		if err != nil {
			return CmdResult{Error: ErrInvalidExpression.Format(err)}
		}
		exprCache.Store(conditionExpr, expr)
	}

	result, err := expr.Evaluate(parameters)
	if err != nil {
		return CmdResult{Error: ErrEvaluationError.Format(err)}
	}

	isTrue, ok := result.(bool)
	if !ok {
		return CmdResult{Error: ErrExpressionNotBoolean}
	}

	if isTrue {
		return p.IfTrue.Do(ctx, cache)
	}
	return p.IfFalse.Do(ctx, cache)
}

func expandAnyAll(expr string, cache *Cache, parameters map[string]any, ctx context.Context) (string, error) {
	// Use shared pre-compiled regex for aggregation functions
	return AggregationPattern.ReplaceAllStringFunc(expr, func(m string) string {
		matches := AggregationPattern.FindStringSubmatch(m)
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

		// Use strings.Builder for efficient string concatenation
		var builder strings.Builder
		join := " || "
		if mode == "all" {
			join = " && "
		}

		builder.WriteByte('(')
		for i, key := range keys {
			varName := keyToIdentifier(key)
			val, err := cache.Get(ctx, key)
			if err != nil {
				val = nil
			}
			parameters[varName] = val

			if i > 0 {
				builder.WriteString(join)
			}
			builder.WriteString(varName)
			builder.WriteByte(' ')
			builder.WriteString(op)
			builder.WriteByte(' ')
			builder.WriteString(right)
		}
		builder.WriteByte(')')

		return builder.String()
	}), nil
}

func keyToIdentifier(key string) string {
	return keyIdentifierReplacer.Replace(key)
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

func substituteContextVars(ctx context.Context, expr string) string {
	val := ctx.Value(triggerVarsContextKey)
	vars, ok := val.([]string)
	if !ok {
		return expr
	}

	for i := range vars {
		v := fmt.Sprintf("${{%d}}", i+1)
		expr = strings.ReplaceAll(expr, v, vars[i])
	}

	return expr
}
