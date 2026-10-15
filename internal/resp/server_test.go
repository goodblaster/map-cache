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
