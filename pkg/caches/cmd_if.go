package caches

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Knetic/govaluate"
)

// exprCache caches compiled expressions to avoid recompiling the same expression
var exprCache sync.Map // map[string]*govaluate.EvaluableExpression

// maxCachedExpressions bounds exprCache; when exceeded the cache is reset.
const maxCachedExpressions = 4096

var exprCacheSize atomic.Int64

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
	// Sub in contextual variables.
	conditionExpr := substituteContextVars(ctx, p.Condition)

	binder := paramBinder{params: map[string]any{}}

	// Handle any(...) or all(...) first
	if strings.Contains(conditionExpr, "any(") || strings.Contains(conditionExpr, "all(") {
		conditionExpr = expandAnyAll(ctx, conditionExpr, cache, &binder)
	}

	// Now handle remaining simple ${{...}} references using shared regex
	if matches := InterpolationPattern.FindAllStringSubmatchIndex(conditionExpr, -1); len(matches) > 0 {
		var builder strings.Builder
		builder.Grow(len(conditionExpr))
		last := 0
		for _, match := range matches {
			key := strings.TrimSpace(conditionExpr[match[2]:match[3]])
			val, err := cache.Get(ctx, key)
			if err != nil {
				val = nil
			}
			builder.WriteString(conditionExpr[last:match[0]])
			builder.WriteString(binder.bind(key, val))
			last = match[1]
		}
		builder.WriteString(conditionExpr[last:])
		conditionExpr = builder.String()
	}

	expr, err := compileExpression(conditionExpr)
	if err != nil {
		return CmdResult{Error: ErrInvalidExpression.Format(err)}
	}

	result, err := expr.Evaluate(binder.params)
	if err != nil {
		return CmdResult{Error: ErrEvaluationError.Format(err)}
	}

	isTrue, ok := result.(bool)
	if !ok {
		return CmdResult{Error: ErrExpressionNotBoolean}
	}

	if isTrue {
		if p.IfTrue == nil {
			return CmdResult{}
		}
		return p.IfTrue.Do(ctx, cache)
	}
	if p.IfFalse == nil {
		return CmdResult{}
	}
	return p.IfFalse.Do(ctx, cache)
}

// paramBinder assigns expression parameter names to cache keys. Names are
// positional (k0, k1, ...) rather than derived from the key, so that:
//   - the expression text - and therefore its compiled, cached form - is the
//     same no matter which keys a trigger or loop substituted in;
//   - keys that are not valid identifiers (e.g. starting with a digit) work;
//   - distinct keys can never collide (a/b vs a_b).
type paramBinder struct {
	params map[string]any
	keys   []string // keys[i] is bound to parameter name k<i>
}

func (b *paramBinder) bind(key string, val any) string {
	// Conditions reference a handful of keys, so a linear scan beats a map.
	for i, bound := range b.keys {
		if bound == key {
			return paramName(i)
		}
	}
	name := paramName(len(b.keys))
	b.keys = append(b.keys, key)
	b.params[name] = val
	return name
}

var paramNames = [...]string{"k0", "k1", "k2", "k3", "k4", "k5", "k6", "k7"}

func paramName(i int) string {
	if i < len(paramNames) {
		return paramNames[i]
	}
	return "k" + strconv.Itoa(i)
}

// compileExpression returns the compiled form of expr, caching the result.
// The cache is bounded: expressions can embed values substituted from keys, so
// the set of distinct expressions is not necessarily small.
func compileExpression(expr string) (*govaluate.EvaluableExpression, error) {
	if cached, ok := exprCache.Load(expr); ok {
		return cached.(*govaluate.EvaluableExpression), nil
	}

	compiled, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		return nil, err
	}

	if exprCacheSize.Add(1) > maxCachedExpressions {
		exprCache.Clear()
		exprCacheSize.Store(1)
	}
	exprCache.Store(expr, compiled)
	return compiled, nil
}

func expandAnyAll(ctx context.Context, expr string, cache *Cache, binder *paramBinder) string {
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
			val, err := cache.Get(ctx, key)
			if err != nil {
				val = nil
			}
			varName := binder.bind(key, val)

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
	})
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

// substituteContextVars replaces ${{1}}, ${{2}}, ... with the wildcard segments
// captured by the trigger that is currently executing, if any.
func substituteContextVars(ctx context.Context, expr string) string {
	if !strings.Contains(expr, "${{") {
		return expr
	}

	vars, ok := ctx.Value(triggerVarsContextKey).([]string)
	if !ok {
		return expr
	}

	return substituteCaptures(expr, vars)
}
