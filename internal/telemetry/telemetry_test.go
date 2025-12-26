package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestInit_Disabled(t *testing.T) {
	shutdown, err := Init(Config{
		Enabled:  false,
		Exporter: "none",
	})

	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Should not panic
	err = shutdown(context.Background())
	assert.NoError(t, err)
}

func TestInit_NoneExporter(t *testing.T) {
	shutdown, err := Init(Config{
		Enabled:  true,
		Exporter: "none",
	})

	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	err = shutdown(context.Background())
	assert.NoError(t, err)
}

func TestInit_Stdout(t *testing.T) {
	shutdown, err := Init(Config{
		Enabled:        true,
		Exporter:       "stdout",
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	})

	require.NoError(t, err)
	require.NotNil(t, shutdown)

	defer shutdown(context.Background())

	// Verify tracer is not nil
	tracer := Tracer()
	assert.NotNil(t, tracer)
}

func TestInit_InvalidExporter(t *testing.T) {
	shutdown, err := Init(Config{
		Enabled:  true,
		Exporter: "invalid",
	})

	assert.Error(t, err)
	assert.Nil(t, shutdown)
	assert.Contains(t, err.Error(), "unknown exporter type")
}

func TestTracer(t *testing.T) {
	// Tracer should work even without initialization
	tracer := Tracer()
	assert.NotNil(t, tracer)
}

func TestMiddleware_CreatesSpan(t *testing.T) {
	// Initialize tracing with stdout exporter for testing
	shutdown, err := Init(Config{
		Enabled:     true,
		Exporter:    "stdout",
		ServiceName: "test",
	})
	require.NoError(t, err)
	defer shutdown(context.Background())

	e := echo.New()
	e.Use(Middleware())

	e.GET("/test", func(c echo.Context) error {
		// Verify context has a valid span
		span := trace.SpanFromContext(c.Request().Context())
		assert.True(t, span.SpanContext().IsValid())
		return c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
}

func TestMiddleware_ExtractsTraceContext(t *testing.T) {
	// Initialize tracing
	shutdown, err := Init(Config{
		Enabled:     true,
		Exporter:    "stdout",
		ServiceName: "test",
	})
	require.NoError(t, err)
	defer shutdown(context.Background())

	e := echo.New()
	e.Use(Middleware())

	var extractedTraceID string

	e.GET("/test", func(c echo.Context) error {
		span := trace.SpanFromContext(c.Request().Context())
		extractedTraceID = span.SpanContext().TraceID().String()
		return c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Add W3C Trace Context header
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", extractedTraceID)
}

func TestMiddleware_WithCacheName(t *testing.T) {
	// Initialize tracing
	shutdown, err := Init(Config{
		Enabled:     true,
		Exporter:    "stdout",
		ServiceName: "test",
	})
	require.NoError(t, err)
	defer shutdown(context.Background())

	e := echo.New()
	e.Use(Middleware())

	e.GET("/test", func(c echo.Context) error {
		span := trace.SpanFromContext(c.Request().Context())
		assert.True(t, span.SpanContext().IsValid())
		return c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Cache-Name", "my-cache")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, 200, rec.Code)
}

func TestMiddleware_RecordsError(t *testing.T) {
	// Initialize tracing
	shutdown, err := Init(Config{
		Enabled:     true,
		Exporter:    "stdout",
		ServiceName: "test",
	})
	require.NoError(t, err)
	defer shutdown(context.Background())

	e := echo.New()
	e.Use(Middleware())

	e.GET("/test", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	// Middleware still returns the error, status code will be set by Echo's error handler
	assert.Equal(t, 500, rec.Code)
}
