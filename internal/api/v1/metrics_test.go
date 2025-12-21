package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	// Reset Prometheus registry for clean test
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	prometheus.DefaultGatherer = prometheus.DefaultRegisterer.(prometheus.Gatherer)

	// Re-register metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	prometheus.MustRegister(httpRequestsTotal)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	prometheus.MustRegister(httpRequestDuration)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)
	prometheus.MustRegister(httpRequestsInFlight)

	// Create test handler
	e := echo.New()
	e.Use(MetricsMiddleware)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	// Make a request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify metrics were recorded
	count := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/test", "200"))
	assert.Equal(t, float64(1), count, "Request counter should be incremented")
}

func TestMetricsMiddleware_InFlight(t *testing.T) {
	// Reset registry
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	prometheus.DefaultGatherer = prometheus.DefaultRegisterer.(prometheus.Gatherer)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)
	prometheus.MustRegister(httpRequestsInFlight)

	e := echo.New()
	e.Use(MetricsMiddleware)

	// Initially should be 0
	inFlight := testutil.ToFloat64(httpRequestsInFlight)
	assert.Equal(t, float64(0), inFlight)

	// Handler that checks in-flight count
	e.GET("/test", func(c echo.Context) error {
		// During request, in-flight should be 1
		inFlight := testutil.ToFloat64(httpRequestsInFlight)
		assert.Equal(t, float64(1), inFlight)
		return c.String(http.StatusOK, "test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// After request, should be back to 0
	inFlight = testutil.ToFloat64(httpRequestsInFlight)
	assert.Equal(t, float64(0), inFlight)
}
