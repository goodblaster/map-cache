package caches

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCache_StopsWorker(t *testing.T) {
	before := runtime.NumGoroutine()

	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("lifecycle-%d", i)
		require.NoError(t, AddCache(name))
		require.NoError(t, DeleteCache(name))
	}

	assert.Eventually(t, func() bool {
		return runtime.NumGoroutine() <= before+2
	}, 2*time.Second, 20*time.Millisecond, "expiration workers leaked")
}

func TestDeleteCache_NotFound(t *testing.T) {
	assert.ErrorIs(t, DeleteCache("lifecycle-missing"), ErrCacheNotFound)
}

func TestAddCache_ConcurrentSameName(t *testing.T) {
	const name = "lifecycle-concurrent"
	defer DeleteCache(name)

	var wg sync.WaitGroup
	var succeeded atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if AddCache(name) == nil {
				succeeded.Add(1)
			}
		}()
	}
	wg.Wait()

	assert.EqualValues(t, 1, succeeded.Load())
}

func TestCacheTTL_StaleTimerDoesNotRemoveNewCache(t *testing.T) {
	const name = "lifecycle-stale-ttl"
	defer DeleteCache(name)

	require.NoError(t, AddCache(name))
	old, err := FetchCache(name)
	require.NoError(t, err)
	require.NoError(t, SetCacheTTL(name, 50))

	// Replace the cache under the same name before the old TTL fires. Bypass
	// DeleteCache so the old timer is still armed.
	caches.Delete(name)
	require.NoError(t, AddCache(name))

	time.Sleep(150 * time.Millisecond)

	current, err := FetchCache(name)
	require.NoError(t, err, "old cache's TTL removed the new cache")
	assert.NotSame(t, old, current)
	old.Close()
}

func TestCacheTTL_ExpiryClosesCache(t *testing.T) {
	const name = "lifecycle-ttl-close"
	require.NoError(t, AddCache(name))
	cache, err := FetchCache(name)
	require.NoError(t, err)
	require.NoError(t, SetCacheTTL(name, 10))

	assert.Eventually(t, func() bool {
		select {
		case <-cache.expirationStop:
			_, err := FetchCache(name)
			return err != nil
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}
