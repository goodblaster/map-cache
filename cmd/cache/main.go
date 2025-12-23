// @title Web Cache API
// @version 1.0
// @description API for managing web cache keys
// @BasePath /api/v1
package main

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/goodblaster/logos"
	"github.com/goodblaster/map-cache/internal/api"
	"github.com/goodblaster/map-cache/internal/api/admin"
	v1 "github.com/goodblaster/map-cache/internal/api/v1"
	"github.com/goodblaster/map-cache/internal/build"
	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/internal/log"
	"github.com/goodblaster/map-cache/internal/telemetry"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Track server start time for uptime calculation
	startTime := time.Now()

	// Configure and create logger (this is the ONLY place logos is used directly)
	var formatter logos.Formatter
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat == "json" {
		formatter = logos.JSONFormatter()
	} else {
		formatter = logos.TextFormatter()
	}

	// Create the logos logger and wrap in adapter
	// Adapter needed because logos methods return logos.Logger (struct) not log.Logger (interface)
	logosLogger := logos.NewLogger(logos.LevelInfo, formatter, os.Stdout)
	log.SetDefault(log.LogosAdapter(logosLogger))

	// Initialize configuration
	config.Init(log.Default())

	// Initialize telemetry
	shutdown, err := telemetry.Init(telemetry.Config{
		Enabled:        config.TelemetryEnabled,
		Exporter:       config.TelemetryExporter,
		OTLPEndpoint:   config.OTLPEndpoint,
		ServiceName:    config.ServiceName,
		ServiceVersion: build.Version,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to initialize telemetry")
	}
	defer shutdown(context.Background())

	err = caches.AddCache(caches.DefaultName)
	if err != nil {
		log.WithError(err).With("cache", caches.DefaultName).Fatal("failed to add default cache")
	}

	e := echo.New()

	// Custom error handler - centralized error logging and response formatting
	e.HTTPErrorHandler = api.CustomErrorHandler

	e.Use(middleware.Recover())

	// Request ID middleware - MUST be first to ensure all logs/traces have correlation IDs
	e.Use(api.RequestIDMiddleware)

	// Logging middleware - placed after RequestIDMiddleware to include request IDs in logs
	e.Use(api.LoggingMiddleware)

	// Only add telemetry middleware if enabled
	if config.TelemetryEnabled && config.TelemetryExporter != "none" {
		e.Use(telemetry.Middleware())
	}

	v1.SetupRoutes(e)
	admin.SetupRoutes(e)

	// Start periodic cache metrics update (every 10 seconds)
	metricsStopChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		// Update immediately on start
		v1.UpdateCacheMetrics()

		for {
			select {
			case <-ticker.C:
				v1.UpdateCacheMetrics()
			case <-metricsStopChan:
				return
			}
		}
	}()
	defer close(metricsStopChan)

	// Health check endpoint (Kubernetes-friendly)
	e.GET("/healthz", func(c echo.Context) error {
		// Collect runtime statistics
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Calculate uptime
		uptime := time.Since(startTime)

		// Get cache information
		cacheList := caches.List()

		return c.JSON(http.StatusOK, map[string]any{
			"status":         "healthy",
			"timestamp":      time.Now().UTC(),
			"uptime_seconds": int64(uptime.Seconds()),
			"build":          build.Info(),
			"system": map[string]any{
				"goroutines":      runtime.NumGoroutine(),
				"memory_alloc_mb": memStats.Alloc / 1024 / 1024,
				"memory_sys_mb":   memStats.Sys / 1024 / 1024,
				"gc_count":        memStats.NumGC,
			},
			"caches": map[string]any{
				"count": len(cacheList),
				"names": cacheList,
			},
		})
	})

	// pprof profiling endpoints
	// These are standard Go profiling endpoints for production debugging
	e.GET("/debug/pprof", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	e.GET("/debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	e.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	e.GET("/debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	e.GET("/debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	e.GET("/debug/pprof/allocs", echo.WrapHandler(pprof.Handler("allocs")))
	e.GET("/debug/pprof/block", echo.WrapHandler(pprof.Handler("block")))
	e.GET("/debug/pprof/goroutine", echo.WrapHandler(pprof.Handler("goroutine")))
	e.GET("/debug/pprof/heap", echo.WrapHandler(pprof.Handler("heap")))
	e.GET("/debug/pprof/mutex", echo.WrapHandler(pprof.Handler("mutex")))
	e.GET("/debug/pprof/threadcreate", echo.WrapHandler(pprof.Handler("threadcreate")))

	// Start server in a goroutine so we can handle shutdown signals
	go func() {
		log.With("address", config.WebAddress).Info("starting server")
		if err := e.Start(config.WebAddress); err != nil && err != http.ErrServerClosed {
			log.WithError(err).With("address", config.WebAddress).Fatal("failed to start web server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Give the server 10 seconds to finish handling existing requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("server forced to shutdown")
	}

	log.Info("server exited gracefully")
}
