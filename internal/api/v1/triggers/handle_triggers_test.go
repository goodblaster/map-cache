package triggers

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

func TestHandleCreateTrigger(t *testing.T) {
	e := echo.New()

	t.Run("create trigger", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial data
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{
			"counter": float64(0),
			"total":   float64(0),
		})
		cache.Release("test")
		require.NoError(t, err)

		// Create trigger: when counter changes, increment total
		reqBody := map[string]any{
			"key": "counter",
			"command": map[string]any{
				"type":  "INC",
				"key":   "total",
				"value": 1,
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/triggers", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCreateTrigger()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			// Response should contain trigger ID
			var result string
			err := json.Unmarshal(rec.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		}

		// Verify trigger fires when counter changes
		cache.Acquire("test")
		err = cache.Replace(ctx, "counter", float64(5))
		cache.Release("test")
		require.NoError(t, err)

		cache.Acquire("test")
		total, err := cache.Get(ctx, "total")
		cache.Release("test")
		assert.NoError(t, err)
		assert.Equal(t, float64(1), total) // Should have been incremented
	})

	t.Run("invalid JSON", func(t *testing.T) {
		cache := caches.New()
		req := httptest.NewRequest(http.MethodPost, "/triggers", bytes.NewReader([]byte("invalid")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCreateTrigger()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}

func TestHandleDeleteTrigger(t *testing.T) {
	e := echo.New()

	t.Run("delete existing trigger", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create trigger
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{"key": "value"})
		triggerID, err := cache.CreateTrigger(ctx, "key", caches.NOOP())
		cache.Release("test")
		require.NoError(t, err)

		// Delete trigger
		req := httptest.NewRequest(http.MethodDelete, "/triggers/"+triggerID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(triggerID)
		c.Set("cache", cache)

		h := handleDeleteTrigger()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("delete non-existent trigger", func(t *testing.T) {
		cache := caches.New()
		req := httptest.NewRequest(http.MethodDelete, "/triggers/nonexistent", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("nonexistent-id")
		c.Set("cache", cache)

		h := handleDeleteTrigger()
		if assert.NoError(t, h(c)) {
			// Should handle gracefully
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
}

func TestHandleReplaceTrigger(t *testing.T) {
	e := echo.New()

	t.Run("replace existing trigger", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial data and trigger
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{
			"key":   "value",
			"count": float64(0),
		})
		triggerID, err := cache.CreateTrigger(ctx, "key", caches.INC("count", 1))
		cache.Release("test")
		require.NoError(t, err)

		// Replace trigger with new command (must include id and key in payload)
		reqBody := map[string]any{
			"id":  triggerID,
			"key": "key",
			"command": map[string]any{
				"type":  "INC",
				"key":   "count",
				"value": 10, // Different increment value
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/triggers/"+triggerID, bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(triggerID)
		c.Set("cache", cache)

		h := handleReplaceTrigger()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code) // Returns 204 not 200
		}

		// Verify new trigger fires with new value
		cache.Acquire("test")
		err = cache.Replace(ctx, "key", "newvalue")
		count, _ := cache.Get(ctx, "count")
		cache.Release("test")
		assert.Equal(t, float64(10), count) // Should be incremented by 10
	})

	t.Run("invalid JSON", func(t *testing.T) {
		cache := caches.New()
		req := httptest.NewRequest(http.MethodPut, "/triggers/some-id", bytes.NewReader([]byte("invalid")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("some-id")
		c.Set("cache", cache)

		h := handleReplaceTrigger()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}
