package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNOOP(t *testing.T) {
	ctx := context.Background()
	cache := New()
	res := NOOP().Do(ctx, cache)
	assert.NoError(t, res.Error)
	assert.Nil(t, res.Value)
}
