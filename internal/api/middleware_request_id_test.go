package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware_GeneratesUUID(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware)

	var capturedRequestID string
	e.GET("/test", func(c echo.Context) error {
		capturedRequestID = GetRequestID(c)
		return c.String(http.StatusOK, capturedRequestID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify request ID was generated
	assert.NotEmpty(t, capturedRequestID, "Request ID should be generated")

	// Verify it's 12 hex characters (shortened UUID)
	assert.Equal(t, 12, len(capturedRequestID), "Request ID should be 12 characters")
	assert.Regexp(t, "^[0-9a-f]{12}$", capturedRequestID, "Request ID should be 12 hex characters")

	// Verify response header contains the request ID
	responseHeader := rec.Header().Get(RequestIDHeader)
	assert.Equal(t, capturedRequestID, responseHeader, "Response header should contain request ID")
}

func TestRequestIDMiddleware_UsesExistingHeader(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware)

	expectedRequestID := "custom-request-id-12345"
	var capturedRequestID string

	e.GET("/test", func(c echo.Context) error {
		capturedRequestID = GetRequestID(c)
		return c.String(http.StatusOK, capturedRequestID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, expectedRequestID)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify the existing request ID was used
	assert.Equal(t, expectedRequestID, capturedRequestID, "Should use existing request ID from header")

	// Verify response header contains the same request ID
	responseHeader := rec.Header().Get(RequestIDHeader)
	assert.Equal(t, expectedRequestID, responseHeader, "Response header should contain the same request ID")
}

func TestRequestIDMiddleware_StoresInContext(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware)

	var requestIDFromContext string
	var requestIDFromGetter string

	e.GET("/test", func(c echo.Context) error {
		// Test direct context access
		if id := c.Get(RequestIDContextKey); id != nil {
			if idStr, ok := id.(string); ok {
				requestIDFromContext = idStr
			}
		}

		// Test helper function
		requestIDFromGetter = GetRequestID(c)

		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Both methods should return the same ID
	assert.NotEmpty(t, requestIDFromContext)
	assert.NotEmpty(t, requestIDFromGetter)
	assert.Equal(t, requestIDFromContext, requestIDFromGetter, "Context access and getter should return same ID")
}

func TestGetRequestID_ReturnsEmptyWhenNotSet(t *testing.T) {
	e := echo.New()

	// No RequestIDMiddleware applied
	var capturedRequestID string
	e.GET("/test", func(c echo.Context) error {
		capturedRequestID = GetRequestID(c)
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return empty string when request ID not set
	assert.Empty(t, capturedRequestID, "Should return empty string when request ID not set")
}

func TestGetRequestID_ReturnsEmptyWhenWrongType(t *testing.T) {
	e := echo.New()

	// Set wrong type in context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(RequestIDContextKey, 12345) // Set an integer instead of string
			return next(c)
		}
	})

	var capturedRequestID string
	e.GET("/test", func(c echo.Context) error {
		capturedRequestID = GetRequestID(c)
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should return empty string when type is wrong
	assert.Empty(t, capturedRequestID, "Should return empty string when value is not a string")
}

func TestRequestIDMiddleware_MultipleRequests(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware)

	requestIDs := make([]string, 0, 3)
	e.GET("/test", func(c echo.Context) error {
		requestIDs = append(requestIDs, GetRequestID(c))
		return c.NoContent(http.StatusOK)
	})

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Verify all request IDs are unique
	assert.Len(t, requestIDs, 3)
	assert.NotEqual(t, requestIDs[0], requestIDs[1], "Request IDs should be unique")
	assert.NotEqual(t, requestIDs[1], requestIDs[2], "Request IDs should be unique")
	assert.NotEqual(t, requestIDs[0], requestIDs[2], "Request IDs should be unique")
}

func TestRequestIDMiddleware_PropagatesErrors(t *testing.T) {
	e := echo.New()
	e.Use(RequestIDMiddleware)

	expectedError := echo.NewHTTPError(http.StatusBadRequest, "test error")
	e.GET("/test", func(c echo.Context) error {
		return expectedError
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Error should propagate, but request ID should still be in response
	assert.NotEmpty(t, rec.Header().Get(RequestIDHeader), "Request ID should be in response even on error")
}
