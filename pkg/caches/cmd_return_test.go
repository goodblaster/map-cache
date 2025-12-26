package caches

import (
	"context"
	"encoding/json"
	//"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRETURN_LiteralsOnly(t *testing.T) {
	ctx := context.Background()
	cache := New()

	cmd := RETURN("hello")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "hello", res.Value)

	cmd = RETURN(123)
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, 123, res.Value)

	cmd = RETURN(true)
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, true, res.Value)
}

func TestRETURN_WithInterpolation(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{"foo": "bar", "num": 42}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	cmd := RETURN("${{foo}}")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "bar", res.Value)

	cmd = RETURN("${{num}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.EqualValues(t, 42, res.Value)
}

func TestRETURN_WithBadInterpolation(t *testing.T) {
	ctx := context.Background()
	cache := New()

	cmd := RETURN("${{missing_key}}")
	res := cmd.Do(ctx, cache)

	assert.Error(t, res.Error)
	assert.Nil(t, res.Value)
}

func TestRETURN_FallbackToLiteral(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Missing key falls back to string literal
	cmd := RETURN("${{missing || default}}")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "default", res.Value)

	// Fallback to integer
	cmd = RETURN("${{missing || 42}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, int64(42), res.Value)

	// Fallback to float
	cmd = RETURN("${{missing || 3.14}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, 3.14, res.Value)

	// Fallback to boolean
	cmd = RETURN("${{missing || true}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, true, res.Value)

	// Fallback to quoted string
	cmd = RETURN("${{missing || \"quoted string\"}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "quoted string", res.Value)
}

func TestRETURN_FallbackWithExistingKey(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{"username": "Alice", "age": 25}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// Key exists - use it (ignore default)
	cmd := RETURN("${{username || Guest}}")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "Alice", res.Value)

	// Key missing - use default literal
	cmd = RETURN("${{status || active}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "active", res.Value)
}

func TestRETURN_FallbackOnlyTwoParts(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// More than 2 parts should error
	cmd := RETURN("${{primary || secondary || tertiary}}")
	res := cmd.Do(ctx, cache)
	assert.Error(t, res.Error)
	assert.Contains(t, res.Error.Error(), "must have exactly 2 parts")

	// Single part (no fallback) is valid but will error if key missing
	cmd = RETURN("${{nonexistent}}")
	res = cmd.Do(ctx, cache)
	assert.Error(t, res.Error)
}

func TestRETURN_FallbackInTemplate(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{"name": "Alice"}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// Use fallback in templated string
	cmd := RETURN("Hello, ${{name || Guest}}!")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "Hello, Alice!", res.Value)

	// Missing key in template
	cmd = RETURN("Status: ${{status || unknown}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "Status: unknown", res.Value)
}

func TestRETURN_FallbackWithWildcard_ShouldError(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Wildcards not allowed with fallback
	cmd := RETURN("${{users/*/name || unknown}}")
	res := cmd.Do(ctx, cache)
	assert.Error(t, res.Error)
	assert.Contains(t, res.Error.Error(), "wildcards not allowed with fallback")
}

func TestRETURN_FallbackPreservesTypes(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{"count": 5, "rate": 2.5, "enabled": false}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// Existing keys preserve their types
	cmd := RETURN("${{count || 0}}")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, float64(5), res.Value) // JSON unmarshal makes it float64

	cmd = RETURN("${{enabled || true}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, false, res.Value)

	// Missing keys use default types
	cmd = RETURN("${{missing_int || 10}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, int64(10), res.Value)
}
