package resp

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet_ClearsTTL(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "k", "a", time.Minute).Err())
	require.NoError(t, client.Set(ctx, "k", "b", 0).Err())

	ttl, err := client.TTL(ctx, "k").Result()
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl, "plain SET should discard the TTL")
}

func TestSet_KeepTTL(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "k", "a", time.Minute).Err())
	require.NoError(t, client.Set(ctx, "k", "b", redis.KeepTTL).Err())

	ttl, err := client.TTL(ctx, "k").Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0))

	val, err := client.Get(ctx, "k").Result()
	require.NoError(t, err)
	assert.Equal(t, "b", val)
}

func TestSet_SyntaxErrors(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	assert.ErrorContains(t, client.Do(ctx, "SET", "k", "v", "NX", "XX").Err(), "syntax error")
	assert.ErrorContains(t, client.Do(ctx, "SET", "k", "v", "EX", "10", "KEEPTTL").Err(), "syntax error")
}
