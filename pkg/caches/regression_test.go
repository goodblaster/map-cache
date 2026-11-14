package caches

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/goodblaster/map-cache/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func countCachedExpressions() int {
	n := 0
	exprCache.Range(func(_, _ any) bool { n++; return true })
	return n
}

func TestPrint_LiteralPercentAndLargeNumbers(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()
	require.NoError(t, cache.Create(ctx, map[string]any{"pct": 50.0, "total": 1000000.0}))

	res := cache.Execute(ctx, PRINT("done ${{pct}}% of ${{total}}"))
	require.NoError(t, res.Error)
	assert.Equal(t, []any{[]any{"done 50% of 1000000"}}, res.Value)

	res = cache.Execute(ctx, PRINT("${{missing}}"))
	assert.ErrorIs(t, res.Error, ErrKeyNotFound)
}

func TestReturn_TemplateDoesNotUseExponentNotation(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()
	require.NoError(t, cache.Create(ctx, map[string]any{"count": 1000000.0, "ratio": 0.5}))

	res := RETURN("count=${{count}} ratio=${{ratio}}").Do(ctx, cache)
	require.NoError(t, res.Error)
	assert.Equal(t, "count=1000000 ratio=0.5", res.Value)
}

func TestTrigger_WildcardVarsInAllCommands(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()
	require.NoError(t, cache.Create(ctx, map[string]any{
		"d": map[string]any{
			"a": map[string]any{"n": 1.0, "hits": 0.0, "tmp": "x", "tasks": map[string]any{"t1": 0.0, "t2": 0.0}},
		},
		"seen": "",
	}))

	_, err := cache.CreateTrigger(ctx, "d/*/n", COMMANDS(
		INC("d/${{1}}/hits", 1),
		DELETE("d/${{1}}/tmp"),
		IF("${{d/${{1}}/hits}} == 1", REPLACE("seen", "first"), NOOP()),
		FOR("${{d/${{1}}/tasks/*}}", INC("d/a/tasks/${{1}}", 5)),
	))
	require.NoError(t, err)

	require.NoError(t, cache.Replace(ctx, "d/a/n", 2.0))

	hits, err := cache.Get(ctx, "d/a/hits")
	require.NoError(t, err)
	assert.Equal(t, 1.0, hits)

	_, err = cache.Get(ctx, "d/a/tmp")
	assert.Error(t, err, "DELETE with ${{1}} should have removed the key")

	seen, _ := cache.Get(ctx, "seen")
	assert.Equal(t, "first", seen)

	t1, _ := cache.Get(ctx, "d/a/tasks/t1")
	t2, _ := cache.Get(ctx, "d/a/tasks/t2")
	assert.Equal(t, 5.0, t1)
	assert.Equal(t, 5.0, t2)
}

func TestIf_KeysThatAreNotIdentifiers(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()
	require.NoError(t, cache.Create(ctx, map[string]any{
		"9lives": 9.0,
		"a/b":    1.0,
		"a_b":    2.0,
	}))

	res := IF("${{9lives}} == 9", RETURN("yes"), RETURN("no")).Do(ctx, cache)
	require.NoError(t, res.Error)
	assert.Equal(t, "yes", res.Value)

	// a/b and a_b used to map to the same parameter name.
	res = IF("${{a/b}} == 1 && ${{a_b}} == 2", RETURN("yes"), RETURN("no")).Do(ctx, cache)
	require.NoError(t, res.Error)
	assert.Equal(t, "yes", res.Value)
}

func TestIf_NilBranchIsNoop(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	res := IF("1 == 2", RETURN("yes"), nil).Do(ctx, cache)
	require.NoError(t, res.Error)
	assert.Nil(t, res.Value)
}

func TestIf_InvalidExpressionMessage(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	res := IF("1 ==", NOOP(), NOOP()).Do(ctx, cache)
	require.Error(t, res.Error)
	assert.NotContains(t, res.Error.Error(), "%!", "error message should be formatted cleanly")
}

func TestIf_ExpressionCacheDoesNotGrowPerKey(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	_, err := cache.CreateTrigger(ctx, "items/*/n", IF("${{items/${{1}}/n}} > 100", NOOP(), NOOP()))
	require.NoError(t, err)

	require.NoError(t, cache.Create(ctx, map[string]any{"items/first/n": 0.0}))
	require.NoError(t, cache.Replace(ctx, "items/first/n", 1.0))
	before := countCachedExpressions()

	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("items/k%d/n", i)
		require.NoError(t, cache.Create(ctx, map[string]any{key: 0.0}))
		require.NoError(t, cache.Replace(ctx, key, 1.0))
	}

	assert.Equal(t, before, countCachedExpressions(),
		"the same condition evaluated for different keys should compile once")
}

func TestIf_ExpressionCacheIsBounded(t *testing.T) {
	for i := 0; i < maxCachedExpressions+10; i++ {
		_, err := compileExpression(fmt.Sprintf("%d == %d", i, i))
		require.NoError(t, err)
	}
	assert.LessOrEqual(t, countCachedExpressions(), maxCachedExpressions)
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		key, pattern string
		captures     []string
		ok           bool
	}{
		{"a/b/c", "a/b/c", nil, true},
		{"a/b/c", "a/*/c", []string{"b"}, true},
		{"a/b/c", "*/*/*", []string{"a", "b", "c"}, true},
		{"/a/b/", "a/*", []string{"b"}, true},
		{"a/b", "a/*/c", nil, false},
		{"a/b/c", "a/*", nil, false},
		{"a/b/c", "a/x/c", nil, false},
		{"a//c", "a/*/c", nil, false},
		{"abc", "a*", nil, false},
	}
	for _, tt := range tests {
		captures, ok := matchPattern(tt.key, tt.pattern)
		assert.Equal(t, tt.ok, ok, "%s vs %s", tt.key, tt.pattern)
		assert.Equal(t, tt.captures, captures, "%s vs %s", tt.key, tt.pattern)

		// Must agree with the exported matcher.
		_, err := ExtractWildcardMatches(tt.key, tt.pattern)
		assert.Equal(t, tt.ok, err == nil, "%s vs %s", tt.key, tt.pattern)
	}
}

func TestDelete_ClearsOnlyRelatedTTLs(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()
	require.NoError(t, cache.Create(ctx, map[string]any{
		"obj":    map[string]any{"child": 1.0},
		"object": 2.0,
		"plain":  3.0,
	}))
	require.NoError(t, cache.SetKeyTTL(ctx, "obj/child", 60000))
	require.NoError(t, cache.SetKeyTTL(ctx, "object", 60000))
	require.NoError(t, cache.SetKeyTTL(ctx, "plain", 60000))

	require.NoError(t, cache.Delete(ctx, "obj", "plain"))

	assert.NotContains(t, cache.keyExps, "obj/child")
	assert.NotContains(t, cache.keyExps, "plain")
	assert.Contains(t, cache.keyExps, "object", "a key that merely shares a prefix must keep its TTL")
}

func TestExpire_ParentExpiryClearsNestedTTLs(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	tag := "test"
	cache.Acquire(tag)
	require.NoError(t, cache.Create(ctx, map[string]any{"obj": map[string]any{"child": 1.0}}))
	require.NoError(t, cache.SetKeyTTL(ctx, "obj/child", 60000))
	require.NoError(t, cache.SetKeyTTL(ctx, "obj", 1))
	cache.Release(tag)

	require.Eventually(t, func() bool {
		cache.Acquire(tag)
		defer cache.Release(tag)
		_, err := cache.Get(ctx, "obj")
		return err != nil
	}, 2*time.Second, 20*time.Millisecond)

	cache.Acquire(tag)
	defer cache.Release(tag)
	assert.Empty(t, cache.keyExps, "the nested key's timer must not outlive its parent")
}

func TestSizeBytes_TracksJSONSize(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()
	assert.Equal(t, 2, cache.SizeBytes(ctx)) // {}

	require.NoError(t, cache.Create(ctx, map[string]any{
		"s": "hello",
		"n": 12.5,
		"b": true,
		"z": nil,
		"a": []any{1.0, "x"},
		"o": map[string]any{"k": false},
	}))
	want := len(`{"s":"hello","n":12.5,"b":true,"z":null,"a":[1,"x"],"o":{"k":false}}`)
	assert.Equal(t, want, cache.SizeBytes(ctx))
}

func TestLastAccessed(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()
	assert.Nil(t, cache.LastAccessed())

	_, _ = cache.Get(ctx, "anything")
	require.NotNil(t, cache.LastAccessed())
	assert.WithinDuration(t, time.Now(), *cache.LastAccessed(), time.Second)
}

func TestFormatValue(t *testing.T) {
	assert.Equal(t, "1000000", FormatValue(1000000.0))
	assert.Equal(t, "0.25", FormatValue(0.25))
	assert.Equal(t, "-3", FormatValue(-3.0))
	assert.Equal(t, "text", FormatValue("text"))
	assert.Equal(t, "true", FormatValue(true))
	assert.Equal(t, "42", FormatValue(int64(42)))
	assert.Equal(t, "1e+30", FormatValue(1e30))
}

func TestCustomKeyDelimiter_WildcardsAndTriggers(t *testing.T) {
	prev := config.KeyDelimiter
	config.KeyDelimiter = "."
	defer func() { config.KeyDelimiter = prev }()

	ctx := context.Background()
	cache := New()
	defer cache.Close()
	require.NoError(t, cache.Create(ctx, map[string]any{
		"users": map[string]any{
			"u1": map[string]any{"n": 1.0, "seen": 0.0},
			"u2": map[string]any{"n": 2.0, "seen": 0.0},
		},
	}))

	assert.ElementsMatch(t, []string{"users.u1.n", "users.u2.n"}, cache.WildKeys(ctx, "users.*.n"))

	_, err := cache.CreateTrigger(ctx, "users.*.n", INC("users.${{1}}.seen", 1))
	require.NoError(t, err)
	require.NoError(t, cache.Replace(ctx, "users.u2.n", 5.0))

	seen, err := cache.Get(ctx, "users.u2.seen")
	require.NoError(t, err)
	assert.Equal(t, 1.0, seen)

	res := FOR("${{users.*.n}}", INC("users.${{1}}.seen", 10)).Do(ctx, cache)
	require.NoError(t, res.Error)
	seen, _ = cache.Get(ctx, "users.u1.seen")
	assert.Equal(t, 10.0, seen)
}
