package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRETURN_LiteralsOnly(t *testing.T) {
	ctx := context.Background()
	cache := New()

	cmd := RETURN("hello", 123, true)
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, []any{"hello", 123, true}, res.Values)
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

	cmd := RETURN("${{foo}}", "${{num}}", "raw string")
	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)

	assert.Equal(t, []any{"bar", "42", "raw string"}, res.Values)
}

func TestRETURN_WithBadInterpolation(t *testing.T) {
	ctx := context.Background()
	cache := New()

	cmd := RETURN("${{missing_key}}")
	res := cmd.Do(ctx, cache)

	assert.Error(t, res.Error)
	assert.Nil(t, res.Values)
}
