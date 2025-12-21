package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/internal/log"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_StoresLoggerInContext(t *testing.T) {
	// Set up a mock logger
	mockLogger := log.NewMockLogger()
	originalLogger := log.Default()
	log.SetDefault(mockLogger)
	defer log.SetDefault(originalLogger) // Reset after test

	e := echo.New()
	e.Use(RequestIDMiddleware)
	e.Use(LoggingMiddleware)

	var loggerFromContext log.Logger
	e.GET("/test", func(c echo.Context) error {
		loggerFromContext = log.FromContext(c.Request().Context())
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify logger was stored in context
	assert.NotNil(t, loggerFromContext, "Logger should be stored in context")
}

func TestLoggingMiddleware_LogsRequestCompletion(t *testing.T) {
	// Set up a mock logger
	mockLogger := log.NewMockLogger()
	originalLogger := log.Default()
	log.SetDefault(mockLogger)
	defer log.SetDefault(originalLogger) // Reset after test

	e := echo.New()
	e.Use(RequestIDMiddleware)
	e.Use(LoggingMiddleware)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify request completion was logged at Info level
	assert.Contains(t, mockLogger.Messages, "INFO", "Info should be called to log request completion")
}

func TestLoggingMiddleware_IncludesRequestID(t *testing.T) {
	// Set up a mock logger
	mockLogger := log.NewMockLogger()
	originalLogger := log.Default()
	log.SetDefault(mockLogger)
	defer log.SetDefault(originalLogger) // Reset after test

	e := echo.New()
	e.Use(RequestIDMiddleware)
	e.Use(LoggingMiddleware)

	expectedRequestID := "test-request-id-123"
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, expectedRequestID)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify request ID was included in logged fields
	assert.Contains(t, mockLogger.Fields, "request_id", "Should log request_id field")
	if requestID, ok := mockLogger.Fields["request_id"]; ok {
		assert.Equal(t, expectedRequestID, requestID, "Request ID should match")
	}
}

func TestLoggingMiddleware_LogsStructuredFields(t *testing.T) {
	// Set up a mock logger
	mockLogger := log.NewMockLogger()
	originalLogger := log.Default()
	log.SetDefault(mockLogger)
	defer log.SetDefault(originalLogger) // Reset after test

	e := echo.New()
	e.Use(RequestIDMiddleware)
	e.Use(LoggingMiddleware)

	e.POST("/test/path", func(c echo.Context) error {
		return c.NoContent(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/test/path", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify structured fields were logged
	fields := mockLogger.Fields
	assert.Equal(t, "POST", fields["method"], "Should log HTTP method")
	assert.Equal(t, "/test/path", fields["path"], "Should log request path")
	assert.Equal(t, 201, fields["status"], "Should log response status")
	assert.NotNil(t, fields["duration_ms"], "Should log request duration")
	assert.NotNil(t, fields["remote_ip"], "Should log remote IP")
}

func TestLoggingMiddleware_LogsErrorStatus(t *testing.T) {
	// Set up a mock logger
	mockLogger := log.NewMockLogger()
	originalLogger := log.Default()
	log.SetDefault(mockLogger)
	defer log.SetDefault(originalLogger) // Reset after test

	e := echo.New()
	e.Use(RequestIDMiddleware)
	e.Use(LoggingMiddleware)

	e.GET("/test", func(c echo.Context) error {
		// Explicitly set status before returning error
		// This is how handlers typically report errors
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify error status was logged
	fields := mockLogger.Fields
	assert.Equal(t, 400, fields["status"], "Should log error status code")
}

func TestLoggingMiddleware_OrderMatters(t *testing.T) {
	// Test that LoggingMiddleware must come after RequestIDMiddleware
	mockLogger := log.NewMockLogger()
	originalLogger := log.Default()
	log.SetDefault(mockLogger)
	defer log.SetDefault(originalLogger) // Reset after test

	e := echo.New()
	// Apply in correct order: RequestID first, then Logging
	e.Use(RequestIDMiddleware)
	e.Use(LoggingMiddleware)

	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should have request_id in logged fields
	assert.Contains(t, mockLogger.Fields, "request_id", "Should have request_id when middleware order is correct")
	assert.NotEmpty(t, mockLogger.Fields["request_id"], "Request ID should not be empty")
}
