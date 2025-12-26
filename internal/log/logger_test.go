package log

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDefaultAndDefault(t *testing.T) {
	// Save original logger
	originalLogger := Default()
	defer SetDefault(originalLogger)

	// Create a new mock logger
	mockLog := NewMockLogger()
	SetDefault(mockLog)

	// Verify it's set
	assert.Equal(t, mockLog, Default())
}

func TestPackageLevelFunctions(t *testing.T) {
	mockLog := NewMockLogger()
	SetDefault(mockLog)
	defer SetDefault(&noopLogger{})

	tests := []struct {
		name     string
		fn       func()
		expected string
	}{
		{
			name:     "Info",
			fn:       func() { Info("test") },
			expected: "INFO",
		},
		{
			name:     "Infof",
			fn:       func() { Infof("test %s", "value") },
			expected: "INFO: test %s",
		},
		{
			name:     "Warn",
			fn:       func() { Warn("test") },
			expected: "WARN",
		},
		{
			name:     "Warnf",
			fn:       func() { Warnf("test %s", "value") },
			expected: "WARN: test %s",
		},
		{
			name:     "Error",
			fn:       func() { Error("test") },
			expected: "ERROR",
		},
		{
			name:     "Errorf",
			fn:       func() { Errorf("test %s", "value") },
			expected: "ERROR: test %s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog.Messages = []string{} // Reset
			tt.fn()
			assert.Contains(t, mockLog.Messages, tt.expected)
		})
	}
}

func TestWithError(t *testing.T) {
	mockLog := NewMockLogger()
	SetDefault(mockLog)
	defer SetDefault(&noopLogger{})

	err := assert.AnError
	logger := WithError(err)

	assert.NotNil(t, logger)
	assert.Contains(t, mockLog.Errors, err)
}

func TestWith(t *testing.T) {
	mockLog := NewMockLogger()
	SetDefault(mockLog)
	defer SetDefault(&noopLogger{})

	logger := With("key", "value")

	assert.NotNil(t, logger)
	assert.Equal(t, "value", mockLog.Fields["key"])
}

func TestFromContext(t *testing.T) {
	mockLog := NewMockLogger()

	tests := []struct {
		name     string
		ctx      context.Context
		expected Logger
	}{
		{
			name:     "with logger in context",
			ctx:      WithLogger(context.Background(), mockLog),
			expected: mockLog,
		},
		{
			name:     "without logger in context",
			ctx:      context.Background(),
			expected: Default(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := FromContext(tt.ctx)
			assert.Equal(t, tt.expected, logger)
		})
	}
}

func TestWithLogger(t *testing.T) {
	mockLog := NewMockLogger()
	ctx := WithLogger(context.Background(), mockLog)

	retrieved := FromContext(ctx)
	assert.Equal(t, mockLog, retrieved)
}

func TestNoopLogger(t *testing.T) {
	noop := &noopLogger{}

	// These should not panic
	noop.Info("test")
	noop.Infof("test %s", "value")
	noop.Warn("test")
	noop.Warnf("test %s", "value")
	noop.Error("test")
	noop.Errorf("test %s", "value")
	noop.Fatal("test")
	noop.Print("test")

	// These should return the noop logger
	assert.Equal(t, noop, noop.WithError(assert.AnError))
	assert.Equal(t, noop, noop.With("key", "value"))
}

func TestConcurrentAccess(t *testing.T) {
	// Test thread-safety of Default/SetDefault
	mockLog := NewMockLogger()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			SetDefault(mockLog)
		}()

		go func() {
			defer wg.Done()
			_ = Default()
		}()
	}

	wg.Wait()
	// Should not panic
}

func BenchmarkPackageFunctions(b *testing.B) {
	mockLog := NewMockLogger()
	SetDefault(mockLog)
	defer SetDefault(&noopLogger{})

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Info("benchmark")
		}
	})

	b.Run("Infof", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Infof("benchmark %d", i)
		}
	})

	b.Run("WithError", func(b *testing.B) {
		err := assert.AnError
		for i := 0; i < b.N; i++ {
			WithError(err)
		}
	})

	b.Run("With", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			With("key", "value")
		}
	})
}

func BenchmarkFromContext(b *testing.B) {
	mockLog := NewMockLogger()
	ctx := WithLogger(context.Background(), mockLog)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromContext(ctx)
	}
}
