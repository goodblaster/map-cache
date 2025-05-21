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
