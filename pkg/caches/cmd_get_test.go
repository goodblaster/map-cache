package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGET(t *testing.T) {
	ctx := context.Background()
	cache := New()

	j := `{"num":1}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(j), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	res := GET("num").Do(ctx, cache)
	assert.NoError(t, res.Error)

	if assert.Len(t, res.Values, 1) {
		assert.EqualValues(t, 1, res.Values[0].(map[string]any)["num"])
		assert.EqualValues(t, cache.cmap.Data(ctx)["num"], res.Values[0].(map[string]any)["num"])
	}
}
