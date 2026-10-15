package resp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPTTL_MillisecondPrecision(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "k", "v", 0).Err())
	require.NoError(t, client.PExpire(ctx, "k", 1700*time.Millisecond).Err())

	pttl, err := client.PTTL(ctx, "k").Result()
	require.NoError(t, err)
	assert.Greater(t, pttl, 1500*time.Millisecond)
	assert.LessOrEqual(t, pttl, 1700*time.Millisecond)

	ttl, err := client.TTL(ctx, "k").Result()
	require.NoError(t, err)
	assert.Equal(t, 2*time.Second, ttl, "TTL rounds like Redis")

	expMs, err := client.Do(ctx, "PEXPIRETIME", "k").Int64()
	require.NoError(t, err)
	assert.InDelta(t, time.Now().Add(1700*time.Millisecond).UnixMilli(), expMs, 100)
}

func TestRename_KeepsSubSecondTTL(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "a", "v", 800*time.Millisecond).Err())
	require.NoError(t, client.Rename(ctx, "a", "b").Err())

	pttl, err := client.PTTL(ctx, "b").Result()
	require.NoError(t, err)
	assert.Greater(t, pttl, time.Duration(0), "renamed key lost its TTL")
}
