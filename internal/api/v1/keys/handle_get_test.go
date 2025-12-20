package keys

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetValue_Success(t *testing.T) {
	// Setup
	e := echo.New()
	cache := caches.New()
	cache.Acquire("test")
	defer cache.Release("test")

	// Create a test key
	err := cache.Create(context.Background(), map[string]any{
		"user/name": "John Doe",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/keys/user/name", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("key")
	c.SetParamValues("user/name")
	c.Set("cache", cache)

	// Execute
	handler := handleGetValue()
	err = handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "John Doe")
}

func TestHandleGetValue_NotFound(t *testing.T) {
	// Setup
	e := echo.New()
	cache := caches.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/keys/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("key")
	c.SetParamValues("nonexistent")
	c.Set("cache", cache)

	// Execute
	handler := handleGetValue()
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "key not found")
}

func TestHandleGetBatch_Success(t *testing.T) {
	// Setup
	e := echo.New()
	cache := caches.New()
	cache.Acquire("test")
	defer cache.Release("test")

	// Create test keys
	err := cache.Create(context.Background(), map[string]any{
		"user/name": "John Doe",
		"user/age":  30,
	})
	assert.NoError(t, err)

	reqBody := `{"keys": ["user/name", "user/age"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys/get", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("cache", cache)

	// Execute
	handler := handleGetBatch()
	err = handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result []any
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestHandleGetBatch_InvalidJSON(t *testing.T) {
	// Setup
	e := echo.New()
	cache := caches.New()

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys/get", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("cache", cache)

	// Execute
	handler := handleGetBatch()
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid json payload")
}

func TestHandleGetBatch_EmptyKeys(t *testing.T) {
	// Setup
	e := echo.New()
	cache := caches.New()

	reqBody := `{"keys": []}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys/get", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("cache", cache)

	// Execute
	handler := handleGetBatch()
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request body")
}

func TestHandleGetBatch_EmptyKeyInArray(t *testing.T) {
	// Setup
	e := echo.New()
	cache := caches.New()

	reqBody := `{"keys": ["valid", ""]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys/get", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("cache", cache)

	// Execute
	handler := handleGetBatch()
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request body")
}

func TestHandleGetBatch_KeyNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	cache := caches.New()

	reqBody := `{"keys": ["nonexistent"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys/get", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("cache", cache)

	// Execute
	handler := handleGetBatch()
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "key not found")
}

func TestGetBatchRequest_Validate_Valid(t *testing.T) {
	req := getBatchRequest{
		Keys: []string{"key1", "key2"},
	}
	err := req.Validate()
	assert.NoError(t, err)
}

func TestGetBatchRequest_Validate_EmptyKeys(t *testing.T) {
	req := getBatchRequest{
		Keys: []string{},
	}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one key is required")
}

func TestGetBatchRequest_Validate_EmptyKeyInArray(t *testing.T) {
	req := getBatchRequest{
		Keys: []string{"valid", ""},
	}
	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key cannot be empty")
}
