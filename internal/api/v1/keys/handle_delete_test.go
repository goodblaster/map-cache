package keys

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDelete(t *testing.T) {
	e := echo.New()

	t.Run("delete existing key", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create key
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{"key1": "value1"})
		cache.Release("test")
		require.NoError(t, err)

		// Delete the key
		req := httptest.NewRequest(http.MethodDelete, "/keys/key1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("key1")
		c.Set("cache", cache)

		h := handleDelete()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify key was deleted
		cache.Acquire("test")
		_, err = cache.Get(ctx, "key1")
		cache.Release("test")
		assert.Error(t, err)
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		cache := caches.New()

		req := httptest.NewRequest(http.MethodDelete, "/keys/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("nonexistent")
		c.Set("cache", cache)

		h := handleDelete()
		// Should succeed even if key doesn't exist (idempotent)
		assert.NoError(t, h(c))
	})
}

func TestHandleDeleteBatch(t *testing.T) {
	e := echo.New()

	t.Run("delete multiple keys", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create keys
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		})
		cache.Release("test")
		require.NoError(t, err)

		// Delete keys
		reqBody := map[string]any{
			"keys": []string{"key1", "key2"},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/keys/delete", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleDeleteBatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify keys were deleted
		cache.Acquire("test")
		_, err1 := cache.Get(ctx, "key1")
		_, err2 := cache.Get(ctx, "key2")
		val3, err3 := cache.Get(ctx, "key3")
		cache.Release("test")
		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.NoError(t, err3) // key3 should still exist
		assert.Equal(t, "value3", val3)
	})

	t.Run("empty keys array", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"keys": []string{},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/keys/delete", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleDeleteBatch()
		err := h(c)
		assert.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
	})

	t.Run("empty key in array", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"keys": []string{"key1", ""},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/keys/delete", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleDeleteBatch()
		err := h(c)
		assert.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
	})
}
