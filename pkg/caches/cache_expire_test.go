package caches

import (
	"context"
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
