package caches

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetKeyExpiration(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	err := cache.Create(ctx, map[string]any{
		"test": "value",
	})
	assert.NoError(t, err)

	// Make sure we can read it.
	val, err := cache.Get(ctx, "test")
	assert.NoError(t, err)
	assert.EqualValues(t, "value", val)

	// Set brief expiration
	err = cache.SetKeyTTL(ctx, "test", 1)

	// Sleep long enough for expiration AND batch processing (100ms ticker + margin)
	time.Sleep(time.Millisecond * 150)

	// Check if the cache is expired
	_, err = cache.Get(ctx, "test")
	assert.Error(t, err)
}

func TestSetKeyExpirationNested(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	err := cache.Create(ctx, map[string]any{
		"test": map[string]any{
			"nested": "value",
		},
	})
	assert.NoError(t, err)

	// Make sure we can read it.
	val, err := cache.Get(ctx, "test/nested")
	assert.NoError(t, err)
	assert.EqualValues(t, "value", val)

	// Set brief expiration
	err = cache.SetKeyTTL(ctx, "test/nested", 1)

	// Sleep long enough for expiration AND batch processing (100ms ticker + margin)
	time.Sleep(time.Millisecond * 150)

	// test should still be there
	_, err = cache.Get(ctx, "test")
	assert.NoError(t, err)

	// nested should not
	_, err = cache.Get(ctx, "test/nested")
	assert.Error(t, err)
}

func TestKeyExpirationChange(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	err := cache.Create(ctx, map[string]any{
		"test": "value",
	})
	assert.NoError(t, err)

	// Set expiration to 10 seconds
	err = cache.SetKeyTTL(ctx, "test", 10_000)
	assert.NoError(t, err)

	// Change expiration to millisecond
	err = cache.SetKeyTTL(ctx, "test", 1)
	assert.NoError(t, err)

	// Sleep long enough for expiration AND batch processing (100ms ticker + margin)
	time.Sleep(time.Millisecond * 150)

	// Check if the cache is expired
	_, err = cache.Get(ctx, "test")
	assert.Error(t, err)
}

func TestKeyExpirationCancel(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	err := cache.Create(ctx, map[string]any{
		"test": "value",
	})
	assert.NoError(t, err)

	// Set short expiration
	err = cache.SetKeyTTL(ctx, "test", 50)
	assert.NoError(t, err)

	// Cancel expiration
	err = cache.CancelKeyTTL(ctx, "test")
	assert.NoError(t, err)

	// Sleep for a bit to let the expiration happen
	time.Sleep(time.Millisecond * 55)

	// Check if the cache is still there
	val, err := cache.Get(ctx, "test")
	assert.NoError(t, err)
	assert.EqualValues(t, "value", val)
}

// A timer that has already fired must not delete the key if its TTL was reset
// or cancelled before the expiration worker processed it.
func TestKeyTTL_ResetAfterTimerFired(t *testing.T) {
	for _, tc := range []struct {
		name   string
		update func(ctx context.Context, cache *Cache) error
	}{
		{"reset", func(ctx context.Context, cache *Cache) error { return cache.SetKeyTTL(ctx, "k", 60_000) }},
		{"cancel", func(ctx context.Context, cache *Cache) error { return cache.CancelKeyTTL(ctx, "k") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cache := New()
			defer cache.Close()

			// Hold the lock so the worker can't process the fired timer until
			// after the TTL has been updated.
			cache.Acquire("test")
			assert.NoError(t, cache.Create(ctx, map[string]any{"k": "v"}))
			assert.NoError(t, cache.SetKeyTTL(ctx, "k", 1))
			time.Sleep(20 * time.Millisecond)
			assert.NoError(t, tc.update(ctx, cache))
			cache.Release("test")

			time.Sleep(250 * time.Millisecond)

			cache.Acquire("test")
			defer cache.Release("test")
			_, err := cache.Get(ctx, "k")
			assert.NoError(t, err, "key should survive")
		})
	}
}

// More keys expiring at once than the expiration channel can buffer must all be
// deleted by the worker (run with -race to catch unlocked fallback deletes).
func TestKeyTTL_ExpirationOverflow(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	const n = 3000 // > expirationChan buffer (1000)
	cache.Acquire("test")
	for i := 0; i < n; i++ {
		key := "k" + strconv.Itoa(i)
		assert.NoError(t, cache.Create(ctx, map[string]any{key: i}))
		assert.NoError(t, cache.SetKeyTTL(ctx, key, 1))
	}
	cache.Release("test")

	assert.Eventually(t, func() bool {
		cache.Acquire("test")
		defer cache.Release("test")
		return len(cache.cmap.Data(ctx)) == 0 && len(cache.keyExps) == 0
	}, 5*time.Second, 50*time.Millisecond)
}

// Clearing the cache must stop key timers so they can't delete a key with the
// same name created afterwards.
func TestClear_StopsKeyTimers(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	cache.Acquire("test")
	assert.NoError(t, cache.Create(ctx, map[string]any{"k": "old"}))
	assert.NoError(t, cache.SetKeyTTL(ctx, "k", 50))
	cache.Clear()
	assert.NoError(t, cache.Create(ctx, map[string]any{"k": "new"}))
	cache.Release("test")

	time.Sleep(250 * time.Millisecond)

	cache.Acquire("test")
	defer cache.Release("test")
	val, err := cache.Get(ctx, "k")
	assert.NoError(t, err)
	assert.Equal(t, "new", val)
	assert.Empty(t, cache.keyExps)
}
