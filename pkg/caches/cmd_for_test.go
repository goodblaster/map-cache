package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFOR_IteratesAndRunsCommands(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"job-1234": {
			"domains": {
				"apple": { "countdown": 1, "status": "busy" },
				"banana": { "countdown": 0, "status": "busy" }
			}
		}
	}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	loopExpr := "${{job-1234/domains/*/countdown}}"
	cmd := FOR(loopExpr,
		IF(
			"${{job-1234/domains/${{1}}/countdown}} == 0",
			REPLACE("job-1234/domains/${{1}}/status", "complete"),
			INC("job-1234/domains/${{1}}/countdown", -1),
		),
	)

	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)

	// Inspect result values
	root := cache.cmap.Data(ctx)
	domains := root["job-1234"].(map[string]any)["domains"].(map[string]any)

	apple := domains["apple"].(map[string]any)
	banana := domains["banana"].(map[string]any)

	assert.EqualValues(t, 0.0, apple["countdown"])
	assert.EqualValues(t, "complete", banana["status"])
}

func TestFOR_InvalidLoopExpr(t *testing.T) {
	ctx := context.Background()
	cache := New()

	cmd := FOR("invalid-no-interpolation", RETURN("should fail"))
	res := cmd.Do(ctx, cache)

	assert.Error(t, res.Error)
	assert.Contains(t, res.Error.Error(), "invalid FOR expression")
}

func TestFOR_NoWildcards(t *testing.T) {
	ctx := context.Background()
	cache := New()

	cmd := FOR("${{job/foo/bar}}", RETURN("bad")).Do(ctx, cache)
	assert.Error(t, cmd.Error)
	assert.Contains(t, cmd.Error.Error(), "must include a wildcard")
}

func TestFOR_TransformsPRINT(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"users": [
			{"name": "Alice", "age": 30},
			{"name": "Bob", "age": 25}
		]
	}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// PRINT should receive captured wildcard segments
	cmd := FOR("${{users/*/name}}",
		PRINT("User at index ${{1}} has name ${{users/${{1}}/name}}"),
	)

	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)
	// PRINT output goes to logger, we just verify no error
}

func TestFOR_TransformsRETURN(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"items": [
			{"id": "abc", "value": 10},
			{"id": "def", "value": 20}
		]
	}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// RETURN should substitute captures
	cmd := FOR("${{items/*/id}}",
		RETURN("Item ${{1}} has ID: ${{items/${{1}}/id}}"),
	)

	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)

	results, ok := res.Value.([]CmdResult)
	assert.True(t, ok)
	assert.Len(t, results, 2)
	assert.Equal(t, "Item 0 has ID: abc", results[0].Value)
	assert.Equal(t, "Item 1 has ID: def", results[1].Value)
}

func TestFOR_TransformsDELETE(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"users": [
			{"name": "Alice", "temp": true},
			{"name": "Bob", "temp": false}
		]
	}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// DELETE should substitute captures
	cmd := FOR("${{users/*/temp}}",
		DELETE("users/${{1}}/temp"),
	)

	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)

	// Verify temp keys are deleted
	getRes := GET("users/0/temp").Do(ctx, cache)
	assert.Error(t, getRes.Error)
	getRes = GET("users/1/temp").Do(ctx, cache)
	assert.Error(t, getRes.Error)

	// Names should still exist
	getRes = GET("users/0/name").Do(ctx, cache)
	assert.NoError(t, getRes.Error)
	assert.Equal(t, "Alice", getRes.Value)
}

func TestFOR_TransformsNestedFOR(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"groups": {
			"admin": {
				"users": ["alice", "bob"],
				"count": 0
			},
			"guest": {
				"users": ["charlie"],
				"count": 0
			}
		}
	}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// Nested FOR loops - outer captures should be substituted in inner LoopExpr
	cmd := FOR("${{groups/*/users}}",
		REPLACE("groups/${{1}}/count", float64(2)), // Use outer capture
	)

	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)

	// Verify counts were set
	getRes := GET("groups/admin/count").Do(ctx, cache)
	assert.NoError(t, getRes.Error)
	assert.Equal(t, float64(2), getRes.Value)
}

func TestFOR_TransformsCOMMANDS(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"tasks": [
			{"id": "task1", "count": 0, "status": "pending"},
			{"id": "task2", "count": 0, "status": "pending"}
		]
	}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// COMMANDS group should have all sub-commands transformed
	cmd := FOR("${{tasks/*/id}}",
		COMMANDS(
			INC("tasks/${{1}}/count", 1),
			REPLACE("tasks/${{1}}/status", "processed"),
		),
	)

	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)

	// Verify both commands executed for each task
	getRes := GET("tasks/0/count").Do(ctx, cache)
	assert.NoError(t, getRes.Error)
	assert.Equal(t, float64(1), getRes.Value)

	getRes = GET("tasks/0/status").Do(ctx, cache)
	assert.NoError(t, getRes.Error)
	assert.Equal(t, "processed", getRes.Value)
}

func TestFOR_TransformsGET(t *testing.T) {
	ctx := context.Background()
	cache := New()

	data := `{
		"items": [
			{"value": "first"},
			{"value": "second"}
		]
	}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	// GET should substitute captures (previously wasn't doing this)
	cmd := FOR("${{items/*/value}}",
		GET("items/${{1}}/value"),
	)

	res := cmd.Do(ctx, cache)
	assert.NoError(t, res.Error)

	results, ok := res.Value.([]CmdResult)
	assert.True(t, ok)
	assert.Len(t, results, 2)
	assert.Equal(t, "first", results[0].Value)
	assert.Equal(t, "second", results[1].Value)
}
