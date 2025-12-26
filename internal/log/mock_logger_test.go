package log

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockLogger(t *testing.T) {
	mock := NewMockLogger()

	assert.NotNil(t, mock)
	assert.Empty(t, mock.Messages)
	assert.Empty(t, mock.Errors)
	assert.NotNil(t, mock.Fields)
}

func TestMockLoggerInfo(t *testing.T) {
	mock := NewMockLogger()

	mock.Info("test")
	assert.Contains(t, mock.Messages, "INFO")

	mock.Infof("test %s", "value")
	assert.Contains(t, mock.Messages, "INFO: test %s")
}

func TestMockLoggerWarn(t *testing.T) {
	mock := NewMockLogger()

	mock.Warn("test")
	assert.Contains(t, mock.Messages, "WARN")

	mock.Warnf("test %s", "value")
	assert.Contains(t, mock.Messages, "WARN: test %s")
}

func TestMockLoggerError(t *testing.T) {
	mock := NewMockLogger()

	mock.Error("test")
	assert.Contains(t, mock.Messages, "ERROR")

	mock.Errorf("test %s", "value")
	assert.Contains(t, mock.Messages, "ERROR: test %s")
}

func TestMockLoggerFatal(t *testing.T) {
	mock := NewMockLogger()

	mock.Fatal("test")
	assert.Contains(t, mock.Messages, "FATAL")
	// Should not actually exit
}

func TestMockLoggerPrint(t *testing.T) {
	mock := NewMockLogger()

	mock.Print("test")
	assert.Contains(t, mock.Messages, "PRINT")
}

func TestMockLoggerWithError(t *testing.T) {
	mock := NewMockLogger()

	err := assert.AnError
	logger := mock.WithError(err)

	assert.Equal(t, mock, logger)
	assert.Contains(t, mock.Errors, err)
}

func TestMockLoggerWith(t *testing.T) {
	mock := NewMockLogger()

	logger := mock.With("key", "value")

	assert.Equal(t, mock, logger)
	assert.Equal(t, "value", mock.Fields["key"])
}

func TestMockLoggerHasMessage(t *testing.T) {
	mock := NewMockLogger()

	mock.Info("important message")

	assert.True(t, mock.HasMessage("INFO"))
	assert.False(t, mock.HasMessage("NOTFOUND"))
}

func TestMockLoggerConcurrency(t *testing.T) {
	mock := NewMockLogger()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			mock.Info("concurrent")
		}()

		go func() {
			defer wg.Done()
			mock.WithError(assert.AnError)
		}()

		go func() {
			defer wg.Done()
			mock.With("key", "value")
		}()
	}

	wg.Wait()
	// Should not panic, messages should be captured
	assert.NotEmpty(t, mock.Messages)
}

func BenchmarkMockLogger(b *testing.B) {
	mock := NewMockLogger()

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mock.Info("benchmark")
		}
	})

	b.Run("WithError", func(b *testing.B) {
		err := assert.AnError
		for i := 0; i < b.N; i++ {
			mock.WithError(err)
		}
	})

	b.Run("HasMessage", func(b *testing.B) {
		mock.Info("test message")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.HasMessage("test")
		}
	})
}
