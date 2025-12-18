package caches

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleCreateCache(t *testing.T) {
	e := echo.New()

	t.Run("valid request", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "testcache",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/caches", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handleCreateCache()

		// Clear caches before test
		_ = caches.DeleteCache("testcache")

		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		reqBody := map[string]interface{}{}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/caches", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handleCreateCache()

		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), "cache name is required")
		}
	})

	t.Run("duplicate cache name", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "dupCache",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/caches", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		h := handleCreateCache()

		// Create first time (should succeed)
		_ = caches.DeleteCache("dupCache")
		assert.NoError(t, h(c))
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Create second time (should fail)
		req = httptest.NewRequest(http.MethodPost, "/caches", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		h = handleCreateCache()
		assert.NoError(t, h(c))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "failed to create cache")
	})
}
