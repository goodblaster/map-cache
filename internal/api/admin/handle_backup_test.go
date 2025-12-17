package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleBackup_Success(t *testing.T) {
	// Setup
	e := echo.New()
	cacheName := "test-backup-cache"
	tmpfile := "/tmp/test-backup.json"
	defer os.Remove(tmpfile)

	// Create cache with data
	err := caches.AddCache(cacheName)
	require.NoError(t, err)
	defer caches.DeleteCache(cacheName)

	cache, err := caches.FetchCache(cacheName)
	require.NoError(t, err)

	ctx := context.Background()
	cache.Acquire("test")
	err = cache.Create(ctx, map[string]any{
		"key1": "value1",
		"key2": 42,
	})
	cache.Release("test")
	require.NoError(t, err)

	// Create request
	reqBody := `{"cache":"` + cacheName + `","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/backup", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = handleBackup(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify file was created
	_, err = os.Stat(tmpfile)
	assert.NoError(t, err)
}

func TestHandleBackup_WithTTL(t *testing.T) {
	// Setup
	e := echo.New()
	cacheName := "test-ttl-cache"
	tmpfile := "/tmp/test-ttl-backup.json"
	defer os.Remove(tmpfile)

	// Create cache with TTL
	err := caches.AddCache(cacheName)
	require.NoError(t, err)
	defer caches.DeleteCache(cacheName)

	cache, err := caches.FetchCache(cacheName)
	require.NoError(t, err)

	ctx := context.Background()
	cache.Acquire("test")
	err = cache.Create(ctx, map[string]any{"key": "value"})
	cache.Release("test")
	require.NoError(t, err)

	// Set cache TTL
	err = caches.SetCacheTTL(cacheName, 3600000) // 1 hour
	require.NoError(t, err)

	// Create request
	reqBody := `{"cache":"` + cacheName + `","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/backup", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = handleBackup(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify file contains expiration
	data, err := os.ReadFile(tmpfile)
	require.NoError(t, err)

	var backup map[string]any
	err = json.Unmarshal(data, &backup)
	require.NoError(t, err)

	// Should have expiration field
	assert.Contains(t, backup, "expiration")
}

func TestHandleBackup_NonExistentCache(t *testing.T) {
	// Setup
	e := echo.New()
	tmpfile := "/tmp/nonexistent-cache.json"
	defer os.Remove(tmpfile)

	// Create request for non-existent cache
	reqBody := `{"cache":"nonexistent","filename":"` + tmpfile + `"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/backup", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handleBackup(c)

	// Assert - Echo handles error internally, returns nil to avoid panic
	// Check that it doesn't panic and completes
	assert.NoError(t, err)
}

func TestHandleBackup_EmptyFilename(t *testing.T) {
	// Setup
	e := echo.New()

	// Create request with empty filename
	reqBody := `{"cache":"test","filename":""}`
	req := httptest.NewRequest(http.MethodPost, "/admin/backup", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handleBackup(c)

	// Assert - Echo handles validation errors, doesn't panic
	assert.NoError(t, err)
}

func TestHandleBackup_InvalidJSON(t *testing.T) {
	// Setup
	e := echo.New()

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/admin/backup", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handleBackup(c)

	// Assert - Echo handles JSON parsing errors, doesn't panic
	assert.NoError(t, err)
}

func TestAdminBackupRequest_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		req       adminBackupRequest
		expectErr bool
	}{
		{
			name: "valid request",
			req: adminBackupRequest{
				CacheName: "test",
				Filename:  "/tmp/test.json",
			},
			expectErr: false,
		},
		{
			name: "empty filename",
			req: adminBackupRequest{
				CacheName: "test",
				Filename:  "",
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
