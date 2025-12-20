// @title Web Cache API
// @version 1.0
// @description API for managing web cache keys
// @BasePath /api/v1
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goodblaster/logos"
	"github.com/goodblaster/map-cache/internal/api/admin"
	v1 "github.com/goodblaster/map-cache/internal/api/v1"
	"github.com/goodblaster/map-cache/internal/build"
	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/internal/log"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
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

	err := caches.AddCache(caches.DefaultName)
	if err != nil {
		log.WithError(err).With("cache", caches.DefaultName).Fatal("failed to add default cache")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	v1.SetupRoutes(e)
	admin.SetupRoutes(e)

	// Health check route
	e.GET("/status", func(c echo.Context) error {
		return c.JSON(200, map[string]any{
			"status": "ok",
			"build":  build.Info(),
		})
	})

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
