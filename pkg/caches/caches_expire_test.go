package caches

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetCacheExpiration(t *testing.T) {
	err := AddCache("test")
	assert.NoError(t, err)

	// Expire almost immediately
	err = SetCacheTTL("test", time.Millisecond)
	assert.NoError(t, err)

	// Sleep for a bit to let the expiration happen
	time.Sleep(time.Millisecond * 5)

	// Check if the cache is expired
	_, err = FetchCache("test")
	assert.Error(t, err)
}

func TestCacheExpirationChange(t *testing.T) {
	err := AddCache("test")
	assert.NoError(t, err)

	// Set expiration to 10 seconds
	err = SetCacheTTL("test", time.Second)
	assert.NoError(t, err)

	// Change expiration to 1 millisecond
	err = SetCacheTTL("test", time.Millisecond)
	assert.NoError(t, err)

	// Sleep for a bit to let the expiration happen
	time.Sleep(time.Millisecond * 5)

	// Check if the cache is expired
	_, err = FetchCache("test")
	assert.Error(t, err)
}

func TestCancelCacheExpiration(t *testing.T) {
	err := AddCache("test")
	assert.NoError(t, err)

	// Set expiration to 10 seconds
	err = SetCacheTTL("test", time.Millisecond*50)
	assert.NoError(t, err)

	// Cancel the expiration
	err = CancelCacheExpiration("test")
	assert.NoError(t, err)

	// Sleep for a bit to let the expiration happen (it shouldn't)
	time.Sleep(time.Millisecond * 55)

	// Check if the cache is still there
	cache, err := FetchCache("test")
	assert.NoError(t, err)
	assert.NotNil(t, cache)
}
