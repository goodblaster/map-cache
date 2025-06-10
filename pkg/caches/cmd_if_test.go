package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIF_TrueCondition(t *testing.T) {
	ctx := context.Background()
	cache := New()

	err := cache.Create(ctx, map[string]any{"flag": true})
	assert.NoError(t, err)

	trueCmd := RETURN("yes")
	falseCmd := RETURN("no")

	cmd := IF("${{flag}}", trueCmd, falseCmd)
	res := cmd.Do(ctx, cache)

	assert.NoError(t, res.Error)
	assert.Equal(t, "yes", res.Value)
}

func TestIF_FalseCondition(t *testing.T) {
	ctx := context.Background()
	cache := New()

	err := cache.Create(ctx, map[string]any{"flag": false})
	assert.NoError(t, err)

	trueCmd := RETURN("yes")
	falseCmd := RETURN("no")

	cmd := IF("${{flag}}", trueCmd, falseCmd)
	res := cmd.Do(ctx, cache)

	assert.NoError(t, res.Error)
	assert.Equal(t, "no", res.Value)
}

func TestIF_ComparisonCondition(t *testing.T) {
	ctx := context.Background()
	cache := New()

	err := cache.Create(ctx, map[string]any{"x": 5})
	assert.NoError(t, err)

	trueCmd := RETURN("gt")
	falseCmd := RETURN("le")

	cmd := IF("${{x}} > 3", trueCmd, falseCmd)
	res := cmd.Do(ctx, cache)

	assert.NoError(t, res.Error)
	assert.Equal(t, "gt", res.Value)
}

func TestIF_AllCondition(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `
	{"a": {
		"one": "done",
		"two": "done"
	}}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	trueCmd := RETURN("all done")
	falseCmd := RETURN("not all done")

	cmd := IF(`all(${{a/*}} == "done")`, trueCmd, falseCmd)
	res := cmd.Do(ctx, cache)

	assert.NoError(t, res.Error)
	assert.Equal(t, "all done", res.Value)
}

func TestIF_AnyCondition(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `
	{"a": {
		"one": "pending",
		"two": "done"
	}}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	trueCmd := RETURN("some done")
	falseCmd := RETURN("none done")

	cmd := IF(`any(${{a/*}} == "done")`, trueCmd, falseCmd)
	res := cmd.Do(ctx, cache)

	assert.NoError(t, res.Error)
	assert.Equal(t, "some done", res.Value)
}

func TestIF_InvalidExpression(t *testing.T) {
	ctx := context.Background()
	cache := New()

	trueCmd := RETURN("yes")
	falseCmd := RETURN("no")

	cmd := IF("this is not valid", trueCmd, falseCmd)
	res := cmd.Do(ctx, cache)

	assert.Error(t, res.Error)
	assert.Contains(t, res.Error.Error(), "invalid expression")
}
