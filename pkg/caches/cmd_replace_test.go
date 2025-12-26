package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestREPLACE_NewKey(t *testing.T) {
	ctx := context.Background()
	cache := New()

	res := REPLACE("foo", "bar").Do(ctx, cache)
	assert.Error(t, res.Error)
	assert.Nil(t, res.Value)

	assert.Equal(t, "key not found: foo", res.Error.Error())
}

func TestREPLACE_OverwriteExistingKey(t *testing.T) {
	ctx := context.Background()
	cache := New()

	err := cache.Create(ctx, map[string]any{"num": 123})
	assert.NoError(t, err)

	res := REPLACE("num", 456).Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, 456, res.Value)

	assert.EqualValues(t, 456.0, cache.cmap.Data(ctx)["num"])
}

func TestREPLACE_NestedPath(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"user": [
			{ "name": "Alice" },
			{ "name": "Bob" }
		]
	}`
	var m map[string]any
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// Replace "Bob" with "Robert"
	res := REPLACE("user/1/name", "Robert").Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, "Robert", res.Value)

	// Navigate the actual nested structure
	root := cache.cmap.Data(ctx)

	user, ok := root["user"].([]any)
	assert.True(t, ok)
	assert.Len(t, user, 2)

	obj, ok := user[1].(map[string]any)
	assert.True(t, ok)

	assert.EqualValues(t, "Robert", obj["name"])
}
