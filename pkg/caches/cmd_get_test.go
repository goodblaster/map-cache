package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGET_SingleKey(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{"foo": "bar"}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	res := GET("foo").Do(ctx, cache)
	assert.NoError(t, res.Error)

	assert.Len(t, res.Values, 1)

	valueMap, ok := res.Values[0].(map[string]any)
	assert.True(t, ok, "expected map[string]any in result")

	assert.Equal(t, "bar", valueMap["foo"])
}

func TestGET_WildcardKey(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"user": [
			{ "name": "Alice" },
			{ "name": "Bob" }
		]
	}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	res := GET("user/*/name").Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Len(t, res.Values, 1)

	valueMap, ok := res.Values[0].(map[string]any)
	assert.True(t, ok, "expected map[string]any in result")

	assert.Equal(t, "Alice", valueMap["user/0/name"])
	assert.Equal(t, "Bob", valueMap["user/1/name"])
}

func TestGET_MissingKey(t *testing.T) {
	ctx := context.Background()
	cache := New()

	res := GET("missing").Do(ctx, cache)
	assert.Error(t, res.Error)
	assert.Contains(t, res.Error.Error(), "not found")
}

func TestGET_MultipleKeys(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{"a": 1, "b": 2}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	res := GET("a", "b").Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Len(t, res.Values, 2)

	valueMap0, ok := res.Values[0].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, 1.0, valueMap0["a"])

	valueMap1, ok := res.Values[1].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, 2.0, valueMap1["b"])
}
