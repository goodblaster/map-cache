package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleRestore_Success(t *testing.T) {
	// Setup: Create a backup file
	tmpfile := "test-restore-success.json"
	defer os.Remove(tmpfile)

	backup := map[string]any{
		"data": map[string]any{
			"key1": "value1",
			"key2": float64(42),
		},
		"key_expirations": map[string]any{},
		"triggers":        map[string]any{},
		"expiration":      nil,
	}

	data, err := json.Marshal(backup)
	require.NoError(t, err)
	err = os.WriteFile(tmpfile, data, 0644)
	require.NoError(t, err)

	// Create Echo instance
	e := echo.New()

	// Create request
	reqBody := `{"cache":"restored-cache","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = handleRestore(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify cache was restored
	cache, err := caches.FetchCache("default") // Restore uses default cache name
	require.NoError(t, err)

	ctx := context.Background()
	cache.Acquire("test")
	val, err := cache.Get(ctx, "key1")
	cache.Release("test")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Cleanup
	caches.DeleteCache("default")
}

func TestHandleRestore_EmptyFilename(t *testing.T) {
	// Setup
	e := echo.New()

	// Create request with empty filename
	reqBody := `{"cache":"test","filename":""}`
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handleRestore(c)

	// Assert - should return 400 for empty filename
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestHandleRestore_NonExistentFile(t *testing.T) {
	// Setup
	e := echo.New()

	// Create request for non-existent file
	reqBody := `{"cache":"test","filename":"nonexistent-file-12345.json"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handleRestore(c)

	// Assert - should return 400 for non-existent file
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestHandleRestore_EmptyFile(t *testing.T) {
	// Setup: Create empty file
	tmpfile := "test-restore-empty.json"
	defer os.Remove(tmpfile)

	err := os.WriteFile(tmpfile, []byte(""), 0644)
	require.NoError(t, err)

	e := echo.New()

	// Create request
	reqBody := `{"cache":"test","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = handleRestore(c)

	// Assert - should return 400 for empty file
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestHandleRestore_MalformedJSON(t *testing.T) {
	// Setup: Create file with invalid JSON
	tmpfile := "test-restore-malformed.json"
	defer os.Remove(tmpfile)

	err := os.WriteFile(tmpfile, []byte("invalid json {"), 0644)
	require.NoError(t, err)

	e := echo.New()

	// Create request
	reqBody := `{"cache":"test","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = handleRestore(c)

	// Assert - should return 500 for malformed JSON (restore failed)
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

// CRITICAL BUG TEST: Test restoring a backup with expired TTL
func TestHandleRestore_ExpiredTTL_Bug(t *testing.T) {
	// Setup: Create a backup file with TTL in the past
	tmpfile := "test-restore-expired-ttl.json"
	defer os.Remove(tmpfile)

	// TTL timestamp is 1 hour in the past
	pastTimestamp := time.Now().Add(-1 * time.Hour).Unix()

	backup := map[string]any{
		"data": map[string]any{
			"key1": "value1",
		},
		"key_expirations": map[string]int64{
			"key1": pastTimestamp,
		},
		"triggers":   map[string]any{},
		"expiration": nil,
	}

	data, err := json.Marshal(backup)
	require.NoError(t, err)
	err = os.WriteFile(tmpfile, data, 0644)
	require.NoError(t, err)

	e := echo.New()

	// Create request
	reqBody := `{"cache":"test","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute - This should not panic or crash
	// The bug is that negative duration calculation doesn't handle expired TTLs properly
	err = handleRestore(c)

	// Assert - Restore should succeed even with expired TTL
	// The implementation may either:
	// 1. Skip expired keys (ideal)
	// 2. Accept negative duration (bug but shouldn't crash)
	// We're testing that it doesn't crash
	assert.NoError(t, err)

	// Cleanup
	caches.DeleteCache("default")
}

// CRITICAL BUG TEST: Test restore with multiple key expirations (closure bug)
func TestHandleRestore_MultipleKeyExpirations_ClosureBug(t *testing.T) {
	// Setup: Create a backup file with multiple keys with different TTLs
	tmpfile := "test-restore-multiple-ttl.json"
	defer os.Remove(tmpfile)

	// TTLs 5 seconds in the future
	futureTimestamp1 := time.Now().Add(5 * time.Second).Unix()
	futureTimestamp2 := time.Now().Add(10 * time.Second).Unix()
	futureTimestamp3 := time.Now().Add(15 * time.Second).Unix()

	backup := map[string]any{
		"data": map[string]any{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
		"key_expirations": map[string]int64{
			"key1": futureTimestamp1,
			"key2": futureTimestamp2,
			"key3": futureTimestamp3,
		},
		"triggers":   map[string]any{},
		"expiration": nil,
	}

	data, err := json.Marshal(backup)
	require.NoError(t, err)
	err = os.WriteFile(tmpfile, data, 0644)
	require.NoError(t, err)

	e := echo.New()

	// Create request
	reqBody := `{"cache":"test","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = handleRestore(c)

	// Assert
	assert.NoError(t, err)

	// Verify all keys were restored
	cache, err := caches.FetchCache("default")
	require.NoError(t, err)

	ctx := context.Background()
	cache.Acquire("test")

	val1, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val1)

	val2, err := cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val2)

	val3, err := cache.Get(ctx, "key3")
	assert.NoError(t, err)
	assert.Equal(t, "value3", val3)

	cache.Release("test")

	// Cleanup
	caches.DeleteCache("default")
}

func TestHandleRestore_WithTriggers(t *testing.T) {
	// Setup: Create a backup file with triggers
	tmpfile := "test-restore-triggers.json"
	defer os.Remove(tmpfile)

	backup := map[string]any{
		"data": map[string]any{
			"counter": float64(0),
		},
		"key_expirations": map[string]any{},
		"triggers": map[string]any{
			"counter": []any{
				map[string]any{
					"id": "trigger-1",
					"command": map[string]any{
						"type": "NOOP",
					},
				},
			},
		},
		"expiration": nil,
	}

	data, err := json.Marshal(backup)
	require.NoError(t, err)
	err = os.WriteFile(tmpfile, data, 0644)
	require.NoError(t, err)

	e := echo.New()

	// Create request
	reqBody := `{"cache":"test","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = handleRestore(c)

	// Assert
	assert.NoError(t, err)

	// Verify data was restored
	cache, err := caches.FetchCache("default")
	require.NoError(t, err)

	ctx := context.Background()
	cache.Acquire("test")
	val, err := cache.Get(ctx, "counter")
	cache.Release("test")
	assert.NoError(t, err)
	assert.Equal(t, float64(0), val)

	// Cleanup
	caches.DeleteCache("default")
}

func TestHandleRestore_InvalidJSON(t *testing.T) {
	// Setup
	e := echo.New()

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/admin/restore", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handleRestore(c)

	// Assert - should return 400 for invalid JSON
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestAdminRestoreRequest_Validate(t *testing.T) {
	// Create a temp file for testing
	tmpfile := "test-validate.json"
	defer os.Remove(tmpfile)
	err := os.WriteFile(tmpfile, []byte("data"), 0644)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		req       adminRestoreRequest
		expectErr bool
	}{
		{
			name: "valid request",
			req: adminRestoreRequest{
				CacheName: "test",
				Filename:  tmpfile,
			},
			expectErr: false,
		},
		{
			name: "empty filename",
			req: adminRestoreRequest{
				CacheName: "test",
				Filename:  "",
			},
			expectErr: true,
		},
		{
			name: "non-existent file",
			req: adminRestoreRequest{
				CacheName: "test",
				Filename:  "nonexistent-12345.json",
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
