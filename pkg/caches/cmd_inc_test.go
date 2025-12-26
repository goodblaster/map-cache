package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestINC(t *testing.T) {
	ctx := context.Background()
	cache := New()

	j := `{"num":1}`
	m := map[string]any{}
	err := json.Unmarshal([]byte(j), &m)
	assert.NoError(t, err)

	err = cache.Create(ctx, m)
	assert.NoError(t, err)

	res := INC("num", 1).Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Equal(t, float64(2), res.Value)

	assert.EqualValues(t, 2, cache.cmap.Data(ctx)["num"])
}
