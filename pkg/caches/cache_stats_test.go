package caches

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOperationStats_RecordLongOperation(t *testing.T) {
	os := NewOperationStats(5)

	// Record a long operation
	os.RecordLongOperation(100*time.Millisecond, "POST /keys", true, false)

	stats := os.GetStats()
	assert.Len(t, stats.RecentHistory, 1)

	// Verify history record
	exec := stats.RecentHistory[0]
	assert.Equal(t, 100*time.Millisecond, exec.Duration)
	assert.Equal(t, "POST /keys", exec.Operation)
	assert.True(t, exec.Success)
	assert.False(t, exec.TimedOut)
}

func TestOperationStats_RecordTimeout(t *testing.T) {
	os := NewOperationStats(5)

	// Record a timed-out operation
	os.RecordLongOperation(5*time.Second, "POST /commands", false, true)

	stats := os.GetStats()
	assert.Len(t, stats.RecentHistory, 1)

	// Verify history record
	exec := stats.RecentHistory[0]
	assert.Equal(t, 5*time.Second, exec.Duration)
	assert.Equal(t, "POST /commands", exec.Operation)
	assert.False(t, exec.Success)
	assert.True(t, exec.TimedOut)
}

func TestOperationStats_RingBufferBehavior(t *testing.T) {
	os := NewOperationStats(3) // Small buffer for testing

	// Fill the buffer
	for i := 0; i < 3; i++ {
		os.RecordLongOperation(time.Duration(i)*time.Millisecond, "GET /keys", true, false)
	}

	stats := os.GetStats()
	assert.Len(t, stats.RecentHistory, 3)

	// Add one more - should drop the oldest
	os.RecordLongOperation(99*time.Millisecond, "POST /keys", true, false)

	stats = os.GetStats()
	assert.Len(t, stats.RecentHistory, 3, "Buffer should stay at max size")

	// First record should be the second one we added (1ms)
	assert.Equal(t, 1*time.Millisecond, stats.RecentHistory[0].Duration)

	// Last record should be the newest one
	assert.Equal(t, 99*time.Millisecond, stats.RecentHistory[2].Duration)
}

func TestOperationStats_MultipleOperationTypes(t *testing.T) {
	os := NewOperationStats(10)

	// Record different operation types
	os.RecordLongOperation(100*time.Millisecond, "POST /keys", true, false)
	os.RecordLongOperation(200*time.Millisecond, "POST /commands", true, false)
	os.RecordLongOperation(150*time.Millisecond, "DELETE /keys", true, false)
	os.RecordLongOperation(300*time.Millisecond, "POST /triggers", false, true)

	stats := os.GetStats()
	assert.Len(t, stats.RecentHistory, 4)

	// Verify operation types are preserved
	assert.Equal(t, "POST /keys", stats.RecentHistory[0].Operation)
	assert.Equal(t, "POST /commands", stats.RecentHistory[1].Operation)
	assert.Equal(t, "DELETE /keys", stats.RecentHistory[2].Operation)
	assert.Equal(t, "POST /triggers", stats.RecentHistory[3].Operation)
}

func TestOperationStats_ThreadSafety(t *testing.T) {
	os := NewOperationStats(100)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				op := "POST /keys"
				if idx%2 == 0 {
					op = "POST /commands"
				}
				os.RecordLongOperation(time.Millisecond, op, true, false)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = os.GetStats()
			}
		}()
	}

	wg.Wait()

	stats := os.GetStats()
	assert.Equal(t, 100, len(stats.RecentHistory), "Should have 100 operations (ring buffer max)")
}

func TestOperationStats_SnapshotIsolation(t *testing.T) {
	os := NewOperationStats(5)

	// Record initial operation
	os.RecordLongOperation(100*time.Millisecond, "POST /keys", true, false)

	// Get snapshot
	snapshot1 := os.GetStats()

	// Record more operations
	os.RecordLongOperation(200*time.Millisecond, "POST /commands", true, false)

	// Get another snapshot
	snapshot2 := os.GetStats()

	// First snapshot should be unchanged
	assert.Len(t, snapshot1.RecentHistory, 1)

	// Second snapshot should have new data
	assert.Len(t, snapshot2.RecentHistory, 2)

	// Modifying snapshot1's history should not affect os
	if len(snapshot1.RecentHistory) > 0 {
		snapshot1.RecentHistory[0].Operation = "MODIFIED"
	}

	// Get fresh snapshot and verify it wasn't affected
	snapshot3 := os.GetStats()
	assert.Equal(t, "POST /keys", snapshot3.RecentHistory[0].Operation, "Original data should be unchanged")
}

func TestOperationStats_EmptyStats(t *testing.T) {
	os := NewOperationStats(10)

	stats := os.GetStats()
	assert.Empty(t, stats.RecentHistory)
}

func TestOperationStats_MultipleTimeouts(t *testing.T) {
	os := NewOperationStats(10)

	// Record mixed operations
	os.RecordLongOperation(100*time.Millisecond, "POST /keys", true, false)     // Success
	os.RecordLongOperation(5*time.Second, "POST /commands", false, true)        // Timeout
	os.RecordLongOperation(200*time.Millisecond, "DELETE /keys", true, false)   // Success
	os.RecordLongOperation(10*time.Second, "POST /triggers", false, true)       // Timeout

	stats := os.GetStats()
	assert.Len(t, stats.RecentHistory, 4)

	// Count timeouts
	timeoutCount := 0
	for _, exec := range stats.RecentHistory {
		if exec.TimedOut {
			timeoutCount++
		}
	}
	assert.Equal(t, 2, timeoutCount)
}
