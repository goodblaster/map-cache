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

func TestRETURN_FallbackToExistingKey(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{"primary": "value1", "secondary": "value2"}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// Primary key exists - use it
	cmd := RETURN("${{primary || secondary}}")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "value1", res.Value)

	// Primary missing - fall back to secondary
	cmd = RETURN("${{missing || secondary}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "value2", res.Value)
}

func TestRETURN_FallbackChain(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{"tertiary": "value3"}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// Try primary, then secondary, then tertiary
	cmd := RETURN("${{primary || secondary || tertiary}}")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "value3", res.Value)

	// All keys missing - use literal default
	cmd = RETURN("${{primary || secondary || fallback_value}}")
	res = cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "fallback_value", res.Value)
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
