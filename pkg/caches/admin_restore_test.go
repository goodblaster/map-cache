package caches

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestore(t *testing.T) {
	ctx := context.Background()

	// Create a temporary backup file
	tmpfile, err := os.CreateTemp("", "restore_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Create backup data
	backup := RestoreContainer{
		Data: map[string]any{
			"key1": "value1",
			"key2": 42,
			"nested": map[string]any{
				"inner": "value",
			},
		},
		KeyExpirations: map[string]int64{},
		Triggers:       map[string][]RawTrigger{},
	}

	// Write backup
	err = json.NewEncoder(tmpfile).Encode(backup)
	require.NoError(t, err)
	tmpfile.Close()

	// Restore
	err = Restore(ctx, "test-restore", tmpfile.Name())
	assert.NoError(t, err)

	// Verify cache was created
	cache, err := FetchCache("test-restore")
	assert.NoError(t, err)
	assert.NotNil(t, cache)

	// Verify data
	val, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	val, err = cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, float64(42), val) // JSON unmarshals numbers as float64

	// Cleanup
	DeleteCache("test-restore")
}

func TestRestoreWithExpiration(t *testing.T) {
	ctx := context.Background()

	tmpfile, err := os.CreateTemp("", "restore_exp_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Create backup with future expiration
	futureTime := time.Now().Add(1 * time.Hour).Unix()
	backup := RestoreContainer{
		Data: map[string]any{
			"key1": "value1",
		},
		KeyExpirations: map[string]int64{
			"key1": futureTime,
		},
		Triggers: map[string][]RawTrigger{},
	}

	err = json.NewEncoder(tmpfile).Encode(backup)
	require.NoError(t, err)
	tmpfile.Close()

	err = Restore(ctx, "test-restore-exp", tmpfile.Name())
	assert.NoError(t, err)

	cache, err := FetchCache("test-restore-exp")
	assert.NoError(t, err)
	assert.NotNil(t, cache)

	// Verify key still exists (not expired yet since futureTime is 1 hour from now)
	val, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Cleanup
	DeleteCache("test-restore-exp")
}

func TestRestoreWithExpiredKey(t *testing.T) {
	ctx := context.Background()

	tmpfile, err := os.CreateTemp("", "restore_expired_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Create backup with past expiration (already expired)
	pastTime := time.Now().Add(-1 * time.Hour).Unix()
	backup := RestoreContainer{
		Data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		KeyExpirations: map[string]int64{
			"key1": pastTime, // This key is already expired
		},
		Triggers: map[string][]RawTrigger{},
	}

	err = json.NewEncoder(tmpfile).Encode(backup)
	require.NoError(t, err)
	tmpfile.Close()

	err = Restore(ctx, "test-restore-expired", tmpfile.Name())
	assert.NoError(t, err)

	cache, err := FetchCache("test-restore-expired")
	assert.NoError(t, err)
	assert.NotNil(t, cache)

	// Verify both keys exist in the restored cache
	// (expired keys are still restored, expiration just isn't set for past times)
	val, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	val, err = cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	// Cleanup
	DeleteCache("test-restore-expired")
}

func TestRestoreWithTriggers(t *testing.T) {
	ctx := context.Background()

	tmpfile, err := os.CreateTemp("", "restore_triggers_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Create backup with triggers
	backup := RestoreContainer{
		Data: map[string]any{
			"counter": 0,
		},
		KeyExpirations: map[string]int64{},
		Triggers: map[string][]RawTrigger{
			"counter": {
				{
					Id:  "trigger1",
					Key: "counter",
					Command: RawCommand{
						Command: NOOP(),
					},
				},
			},
		},
	}

	err = json.NewEncoder(tmpfile).Encode(backup)
	require.NoError(t, err)
	tmpfile.Close()

	err = Restore(ctx, "test-restore-triggers", tmpfile.Name())
	assert.NoError(t, err)

	cache, err := FetchCache("test-restore-triggers")
	assert.NoError(t, err)
	assert.NotNil(t, cache)

	// Verify data was restored
	val, err := cache.Get(ctx, "counter")
	assert.NoError(t, err)
	assert.Equal(t, float64(0), val) // JSON unmarshals numbers as float64

	// Cleanup
	DeleteCache("test-restore-triggers")
}

func TestRestoreDefaultCache(t *testing.T) {
	ctx := context.Background()

	tmpfile, err := os.CreateTemp("", "restore_default_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	backup := RestoreContainer{
		Data: map[string]any{
			"key1": "value1",
		},
		KeyExpirations: map[string]int64{},
		Triggers:       map[string][]RawTrigger{},
	}

	err = json.NewEncoder(tmpfile).Encode(backup)
	require.NoError(t, err)
	tmpfile.Close()

	// Restore with empty name should use DefaultName
	err = Restore(ctx, "", tmpfile.Name())
	assert.NoError(t, err)

	cache, err := FetchCache(DefaultName)
	assert.NoError(t, err)

	val, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestRestoreInvalidFile(t *testing.T) {
	ctx := context.Background()

	err := Restore(ctx, "test", "/nonexistent/file.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening backup file")
}

func TestRestoreInvalidJSON(t *testing.T) {
	ctx := context.Background()

	tmpfile, err := os.CreateTemp("", "restore_invalid_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write invalid JSON
	_, err = tmpfile.WriteString("{invalid json")
	require.NoError(t, err)
	tmpfile.Close()

	err = Restore(ctx, "test", tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding backup file")
}
