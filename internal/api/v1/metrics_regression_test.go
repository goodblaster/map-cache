package v1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsMiddleware_RecordsErrorStatus(t *testing.T) {
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Total number of HTTP requests"},
		[]string{"method", "path", "status"},
	)

	e := echo.New()
	e.Use(MetricsMiddleware)
	e.GET("/missing", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "key not found")
	})
	e.GET("/broken", func(c echo.Context) error {
		return errors.New("unexpected")
	})

	for _, path := range []string{"/missing", "/broken"} {
		e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	assert.Equal(t, float64(1), testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/missing", "404")))
	assert.Equal(t, float64(1), testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/broken", "500")))
	assert.Equal(t, float64(0), testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/missing", "200")),
		"failed requests must not be counted as 200s")
}

func TestUpdateCacheMetrics_DropsSeriesForDeletedCaches(t *testing.T) {
	const name = "metrics-regression-temp"
	_ = caches.DeleteCache(name)
	require.NoError(t, caches.AddCache(name))

	UpdateCacheMetrics()
	assert.True(t, hasCacheSeries(t, name), "series should exist while the cache does")

	require.NoError(t, caches.DeleteCache(name))
	UpdateCacheMetrics()
	assert.False(t, hasCacheSeries(t, name), "series should be removed with the cache")
}

// hasCacheSeries reports whether cache_size_bytes currently has a series for
// the named cache. (WithLabelValues can't be used to check: it creates one.)
func hasCacheSeries(t *testing.T, name string) bool {
	t.Helper()
	ch := make(chan prometheus.Metric, 1024)
	cacheSizeBytes.Collect(ch)
	close(ch)

	for m := range ch {
		var out dto.Metric
		require.NoError(t, m.Write(&out))
		for _, label := range out.GetLabel() {
			if label.GetName() == "cache" && label.GetValue() == name {
				return true
			}
		}
	}
	return false
}
