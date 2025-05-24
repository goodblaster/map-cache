package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrigger_Simple(t *testing.T) {
	ctx := context.Background()

	j := `
		{
			"a": {
				"b": {"c": 2},
				"z": "busy"
			}
		}`

	m := map[string]any{}
	err := json.Unmarshal([]byte(j), &m)
	assert.Nil(t, err)

	cache := New()
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	id, err := cache.CreateTrigger(ctx, "a/b/c", IF(
		"${{a/b/c}} == 0",
		REPLACE("a/z", "complete"),
		NOOP(),
	))
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	val, err := cache.Get(ctx, "a/b/c")
	assert.NoError(t, err)
	assert.EqualValues(t, 2, val)

	res := INC("a/b/c", -1).Do(ctx, cache)
	assert.NoError(t, res.Error)

	val, err = cache.Get(ctx, "a/b/c")
	assert.NoError(t, err)
	assert.EqualValues(t, 1, val)

	val, err = cache.Get(ctx, "a/z")
	assert.Nil(t, err)
	assert.EqualValues(t, "busy", val)

	res = INC("a/b/c", -1).Do(ctx, cache)
	assert.NoError(t, res.Error)

	val, err = cache.Get(ctx, "a/b/c")
	assert.NoError(t, err)
	assert.EqualValues(t, 0, val)

	val, err = cache.Get(ctx, "a/z")
	assert.NoError(t, err)
	assert.EqualValues(t, "complete", val)
}

func TestTrigger_Wildcard(t *testing.T) {
	ctx := context.Background()

	j := `
		{
			"a": {
				"b": {"c": 2},
				"d": {"c": 1},
				"z": "busy"
			}
		}`

	m := map[string]any{}
	err := json.Unmarshal([]byte(j), &m)
	assert.Nil(t, err)

	cache := New()
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	id, err := cache.CreateTrigger(ctx, "a/*/c", IF(
		"all(${{a/*/c}} == 0)",
		REPLACE("a/z", "complete"),
		NOOP(),
	))
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	val, err := cache.Get(ctx, "a/b/c")
	assert.NoError(t, err)
	assert.EqualValues(t, 2, val)

	res := INC("a/b/c", -1).Do(ctx, cache)
	assert.NoError(t, res.Error)

	val, err = cache.Get(ctx, "a/b/c")
	assert.NoError(t, err)
	assert.EqualValues(t, 1, val)

	val, err = cache.Get(ctx, "a/z")
	assert.Nil(t, err)
	assert.EqualValues(t, "busy", val)

	res = INC("a/b/c", -1).Do(ctx, cache)
	assert.NoError(t, res.Error)

	val, err = cache.Get(ctx, "a/b/c")
	assert.NoError(t, err)
	assert.EqualValues(t, 0, val)

	val, err = cache.Get(ctx, "a/z")
	assert.NoError(t, err)
	assert.EqualValues(t, "busy", val)

	res = INC("a/d/c", -1).Do(ctx, cache)
	assert.NoError(t, res.Error)

	val, err = cache.Get(ctx, "a/d/c")
	assert.NoError(t, err)
	assert.EqualValues(t, 0, val)

	val, err = cache.Get(ctx, "a/z")
	assert.NoError(t, err)
	assert.EqualValues(t, "complete", val)
}

func TestTrigger_Nested(t *testing.T) {
	ctx := context.Background()
	j := `
		{
			"a": {
				"countdown": 2,
				"state": "busy"
			},
			"b": {
				"countdown": 1,
				"state": "busy"
			},
			"state": "busy"
		}`

	m := map[string]any{}
	err := json.Unmarshal([]byte(j), &m)
	assert.Nil(t, err)

	cache := New()
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	id, err := cache.CreateTrigger(ctx, "*/countdown",
		FOR("${{*/countdown}}",
			IF("(${{${{1}}/countdown}} == 0)",
				REPLACE("${{1}}/state", "complete"),
				NOOP(),
			),
		))
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	id, err = cache.CreateTrigger(ctx, "*/state", IF(
		`all(${{*/state}} == "complete")`,
		REPLACE("state", "complete"),
		NOOP(),
	))
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	val, err := cache.Get(ctx, "a/countdown")
	assert.NoError(t, err)
	assert.EqualValues(t, 2, val)

	res := INC("a/countdown", -1).Do(ctx, cache)
	assert.NoError(t, res.Error)

	val, err = cache.Get(ctx, "a/countdown")
	assert.NoError(t, err)
	assert.EqualValues(t, 1, val)

	val, err = cache.Get(ctx, "a/state")
	assert.Nil(t, err)
	assert.EqualValues(t, "busy", val)

	res = INC("a/countdown", -1).Do(ctx, cache)
	assert.NoError(t, res.Error)

	val, err = cache.Get(ctx, "a/countdown")
	assert.NoError(t, err)
	assert.EqualValues(t, 0, val)

	val, err = cache.Get(ctx, "a/state")
	assert.NoError(t, err)
	assert.EqualValues(t, "complete", val)

	val, err = cache.Get(ctx, "state")
	assert.NoError(t, err)
	assert.EqualValues(t, "busy", val)

	res = INC("b/countdown", -1).Do(ctx, cache)
	assert.NoError(t, res.Error)

	val, err = cache.Get(ctx, "b/countdown")
	assert.NoError(t, err)
	assert.EqualValues(t, 0, val)

	val, err = cache.Get(ctx, "b/state")
	assert.NoError(t, err)
	assert.EqualValues(t, "complete", val)

	val, err = cache.Get(ctx, "state")
	assert.NoError(t, err)
	assert.EqualValues(t, "complete", val)
}
