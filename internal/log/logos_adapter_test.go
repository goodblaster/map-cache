package log

import (
	"os"
	"testing"

	"github.com/goodblaster/logos"
	"github.com/stretchr/testify/assert"
)

func TestLogosAdapter(t *testing.T) {
	// Create a logos logger
	logosLogger := logos.NewLogger(logos.LevelInfo, logos.TextFormatter(), os.Stdout)

	// Wrap it
	adapter := LogosAdapter(logosLogger)

	assert.NotNil(t, adapter)

	// Test that it implements Logger interface
	var _ Logger = adapter

	// These should not panic
	adapter.Info("test")
	adapter.Infof("test %s", "value")
	adapter.Warn("test")
	adapter.Warnf("test %s", "value")
	adapter.Error("test")
	adapter.Errorf("test %s", "value")
	adapter.Print("test")
}

func TestLogosAdapterWithError(t *testing.T) {
	logosLogger := logos.NewLogger(logos.LevelInfo, logos.TextFormatter(), os.Stdout)
	adapter := LogosAdapter(logosLogger)

	logger := adapter.WithError(assert.AnError)
	assert.NotNil(t, logger)

	// Should still be a Logger
	var _ Logger = logger
}

func TestLogosAdapterWith(t *testing.T) {
	logosLogger := logos.NewLogger(logos.LevelInfo, logos.TextFormatter(), os.Stdout)
	adapter := LogosAdapter(logosLogger)

	logger := adapter.With("key", "value")
	assert.NotNil(t, logger)

	// Should still be a Logger
	var _ Logger = logger
}

func BenchmarkLogosAdapter(b *testing.B) {
	logosLogger := logos.NewLogger(logos.LevelInfo, logos.TextFormatter(), os.Stdout)
	adapter := LogosAdapter(logosLogger)

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			adapter.Info("benchmark")
		}
	})

	b.Run("WithError", func(b *testing.B) {
		err := assert.AnError
		for i := 0; i < b.N; i++ {
			adapter.WithError(err)
		}
	})

	b.Run("With", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			adapter.With("key", "value")
		}
	})
}
