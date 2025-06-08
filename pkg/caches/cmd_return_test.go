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
