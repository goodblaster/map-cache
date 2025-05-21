package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPRINT_LiteralOnly(t *testing.T) {
	ctx := context.Background()
	cache := New()

	res := PRINT("hello world", "static message").Do(ctx, cache)
	assert.NoError(t, res.Error)

	expected := []any{"hello world", "static message"}
	assert.Equal(t, expected, res.Values)
}

func TestPRINT_WithInterpolation(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"foo": "bar",
		"num": 42
	}`
	var m map[string]any
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	res := PRINT("value is ${{foo}}", "number: ${{num}}").Do(ctx, cache)
	assert.NoError(t, res.Error)

	expected := []any{
		"value is bar",
		"number: 42",
	}
	assert.Equal(t, expected, res.Values)
}

func TestPRINT_MissingKey(t *testing.T) {
	ctx := context.Background()
	cache := New()

	res := PRINT("this will fail: ${{nope}}").Do(ctx, cache)
	assert.Error(t, res.Error)
	assert.Contains(t, res.Error.Error(), "nope")
}
