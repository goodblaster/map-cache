package resp

import (
	"context"
	"testing"
	"time"

	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// startTestServer starts a RESP server on a random local port backed by a
// fresh default cache, and returns a connected client.
func startTestServer(t *testing.T) (*Server, *redis.Client) {
	t.Helper()

	_ = caches.DeleteCache(caches.DefaultName)
	require.NoError(t, caches.AddCache(caches.DefaultName))

	prevAddr := config.RESPAddress
	config.RESPAddress = "127.0.0.1:0"
	srv := NewServer()
	require.NoError(t, srv.Start())
	config.RESPAddress = prevAddr

	client := redis.NewClient(&redis.Options{
		Addr:       srv.listener.Addr().String(),
		MaxRetries: 0,
	})

	t.Cleanup(func() {
		client.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		_ = caches.DeleteCache(caches.DefaultName)
	})

	return srv, client
}

func TestShutdown_WithIdleClient(t *testing.T) {
	srv, client := startTestServer(t)
	ctx := context.Background()

	// Establish an idle connection.
	require.NoError(t, client.Ping(ctx).Err())

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	start := time.Now()
	require.NoError(t, srv.Shutdown(shutdownCtx))
	require.Less(t, time.Since(start), time.Second)
	require.Equal(t, 0, srv.ActiveConnections())
}

func TestFlushDB_ClearsKeysAndTTLs(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "a", "1", time.Minute).Err())
	require.NoError(t, client.HSet(ctx, "h", "f", "v").Err())
	require.NoError(t, client.FlushDB(ctx).Err())

	n, err := client.Exists(ctx, "a", "h").Result()
	require.NoError(t, err)
	require.Zero(t, n)

	require.NoError(t, client.Set(ctx, "a", "2", 0).Err())
	ttl, err := client.TTL(ctx, "a").Result()
	require.NoError(t, err)
	require.Equal(t, time.Duration(-1), ttl)
}
