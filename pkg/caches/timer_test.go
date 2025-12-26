package caches

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFutureFunc(t *testing.T) {
	var called atomic.Bool

	timer := FutureFunc(100, func() {
		called.Store(true)
	})

	assert.NotNil(t, timer)

	// Wait for timer to fire
	time.Sleep(150 * time.Millisecond)

	assert.True(t, called.Load(), "function should have been called")
}

func TestTimerStop(t *testing.T) {
	var called atomic.Bool

	timer := FutureFunc(100, func() {
		called.Store(true)
	})

	// Stop the timer before it fires
	timer.Stop()

	// Wait to ensure it doesn't fire
	time.Sleep(150 * time.Millisecond)

	assert.False(t, called.Load(), "function should not have been called")
}

func TestTimerStopMultipleTimes(t *testing.T) {
	var called atomic.Bool

	timer := FutureFunc(100, func() {
		called.Store(true)
	})

	// Stopping multiple times should not panic
	timer.Stop()
	timer.Stop()
	timer.Stop()

	assert.False(t, called.Load())
}

func TestFutureFuncZeroDelay(t *testing.T) {
	var called atomic.Bool

	timer := FutureFunc(0, func() {
		called.Store(true)
	})

	assert.NotNil(t, timer)

	// Should fire almost immediately
	time.Sleep(50 * time.Millisecond)

	assert.True(t, called.Load())
}

func TestFutureFuncNegativeDelay(t *testing.T) {
	var called atomic.Bool

	timer := FutureFunc(-100, func() {
		called.Store(true)
	})

	assert.NotNil(t, timer)

	// Should fire immediately
	time.Sleep(50 * time.Millisecond)

	assert.True(t, called.Load())
}

func TestMultipleTimers(t *testing.T) {
	var count atomic.Int32

	timer1 := FutureFunc(50, func() {
		count.Add(1)
	})
	timer2 := FutureFunc(100, func() {
		count.Add(1)
	})
	timer3 := FutureFunc(150, func() {
		count.Add(1)
	})

	assert.NotNil(t, timer1)
	assert.NotNil(t, timer2)
	assert.NotNil(t, timer3)

	// Wait for all to fire
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, int32(3), count.Load())
}

func TestTimerStopRace(t *testing.T) {
	// Test that Stop() is safe to call concurrently
	timer := FutureFunc(100, func() {})

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			timer.Stop()
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkFutureFunc(b *testing.B) {
	b.Run("create", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			timer := FutureFunc(1000, func() {})
			timer.Stop()
		}
	})

	b.Run("create_and_fire", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			done := make(chan struct{})
			FutureFunc(1, func() {
				close(done)
			})
			<-done
		}
	})

	b.Run("stop", func(b *testing.B) {
		timers := make([]*Timer, b.N)
		for i := 0; i < b.N; i++ {
			timers[i] = FutureFunc(10000, func() {})
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			timers[i].Stop()
		}
	})
}
