package caches

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBackupAndRestore(t *testing.T) {
	ctx := context.Background()
	cacheName := uuid.NewString()
	err := AddCache(cacheName)
	assert.NoError(t, err)

	cache, err := FetchCache(cacheName)
	assert.NoError(t, err, "Failed to fetch cache")

	// Create some initial data
	err = cache.Create(ctx, map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
	assert.NoError(t, err)

	// Create a timeout for the key1
	err = cache.SetKeyTTL(ctx, "key1", 3600*1000) // 1 hour
	assert.NoError(t, err, "Failed to set key1 TTL")

	// Create a timeout for the entire cache
	err = SetCacheTTL(cacheName, 3600*1000) // 1 hour
	assert.NoError(t, err, "Failed to set cache TTL")

	// Create a trigger for key1
	triggerId, err := cache.CreateTrigger(ctx, "key1", NOOP())
	assert.NoError(t, err, "Failed to create trigger for key1")

	// Backup the cache
	err = Backup(ctx, cacheName, cacheName)
	assert.NoError(t, err, "Failed to backup cache")

	// Verify backup file exists
	backupFile := cacheName
	_, err = os.Stat(backupFile)
	if os.IsNotExist(err) {
		t.Fatalf("Backup file %q does not exist", backupFile)
	}

	// Ensure the backup file is cleaned up after the test
	defer os.Remove(backupFile)

	/////////////////// RESTORE TEST ///////////////////

	// Delete the cache so we can more easily test restore
	err = DeleteCache(cacheName)
	assert.NoError(t, err, "Failed to delete cache")

	// Verify cache is deleted
	_, err = FetchCache(cacheName)
	assert.Error(t, err)

	// Restore the cache from backup
	err = Restore(ctx, cacheName, backupFile)
	assert.NoError(t, err, "Failed to restore cache from backup")

	// Fetch the restored cache
	restoredCache, err := FetchCache(cacheName)
	assert.NoError(t, err, "Failed to fetch restored cache")

	// Verify the data in the restored cache
	value1, err := restoredCache.Get(ctx, "key1")
	if assert.NoError(t, err, "Failed to get key1 from restored cache") {
		assert.Equal(t, "value1", value1, "Restored value for key1 does not match")
	}

	value2, err := restoredCache.Get(ctx, "key2")
	if assert.NoError(t, err, "Failed to get key2 from restored cache") {
		assert.Equal(t, "value2", value2, "Restored value for key2 does not match")
	}

	// Verify there is a TTL for key1
	_, ok := restoredCache.keyExps["key1"]
	assert.True(t, ok, "Key1 TTL should exist in restored cache")

	// Verify there is a cache expiration
	assert.NotNil(t, restoredCache.exp)

	// Verify the trigger for key1 exists
	triggers, ok := restoredCache.triggers["key1"]
	assert.True(t, ok, "Trigger for key1 should exist in restored cache")
	assert.Len(t, triggers, 1, "There should be one trigger for key1")

	// Verify the trigger ID matches
	for _, trigger := range triggers {
		assert.Equal(t, triggerId, trigger.Id, "Trigger ID for key1 does not match")
	}
}
