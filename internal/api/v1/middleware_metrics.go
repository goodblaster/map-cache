package v1

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
		},
		[]string{"method", "path"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100, 1000, 10000, ...
		},
		[]string{"method", "path"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)
)

// MetricsMiddleware tracks HTTP request metrics for Prometheus
func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Increment in-flight requests
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// Track request size
		if c.Request().ContentLength > 0 {
			httpRequestSize.WithLabelValues(
				c.Request().Method,
				c.Path(),
			).Observe(float64(c.Request().ContentLength))
		}

		// Call next handler
		err := next(c)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		status := c.Response().Status
		if status == 0 {
			status = 200 // Default if not set
		}

		// Record metrics
		httpRequestsTotal.WithLabelValues(
			c.Request().Method,
			c.Path(),
			strconv.Itoa(status),
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request().Method,
			c.Path(),
		).Observe(duration)

		// Track response size
		if c.Response().Size > 0 {
			httpResponseSize.WithLabelValues(
				c.Request().Method,
				c.Path(),
			).Observe(float64(c.Response().Size))
		}

		return err
	}
}
