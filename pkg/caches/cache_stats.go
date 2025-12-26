package caches

import (
	"sync"
	"time"
)

// OperationExecution represents a single long-running operation record
type OperationExecution struct {
	Timestamp time.Time     // When the operation started
	Duration  time.Duration // How long it took
	Operation string        // Type of operation (e.g., "POST /keys", "POST /commands")
	Success   bool          // Whether it completed successfully
	TimedOut  bool          // Whether it exceeded timeout
}

// OperationStats tracks long-running operation metrics for a cache
type OperationStats struct {
	mu             sync.RWMutex
	recentHistory  []OperationExecution
	maxHistorySize int
}

// NewOperationStats creates a new operation stats tracker
func NewOperationStats(maxHistorySize int) *OperationStats {
	return &OperationStats{
		recentHistory:  make([]OperationExecution, 0, maxHistorySize),
		maxHistorySize: maxHistorySize,
	}
}

// RecordLongOperation adds a long-running operation to the history
// This should only be called for operations exceeding the threshold
func (os *OperationStats) RecordLongOperation(duration time.Duration, operation string, success bool, timedOut bool) {
	os.mu.Lock()
	defer os.mu.Unlock()

	// Add to recent history
	exec := OperationExecution{
		Timestamp: time.Now(),
		Duration:  duration,
		Operation: operation,
		Success:   success,
		TimedOut:  timedOut,
	}

	// Ring buffer behavior: remove oldest if at capacity
	if len(os.recentHistory) >= os.maxHistorySize {
		// Shift left, drop first element
		copy(os.recentHistory, os.recentHistory[1:])
		os.recentHistory[len(os.recentHistory)-1] = exec
	} else {
		os.recentHistory = append(os.recentHistory, exec)
	}
}

// GetStats returns a snapshot of current statistics
func (os *OperationStats) GetStats() OperationStatsSnapshot {
	os.mu.RLock()
	defer os.mu.RUnlock()

	// Create a copy of history to avoid race conditions
	historyCopy := make([]OperationExecution, len(os.recentHistory))
	copy(historyCopy, os.recentHistory)

	return OperationStatsSnapshot{
		RecentHistory: historyCopy,
	}
}

// OperationStatsSnapshot is a thread-safe snapshot of operation statistics
type OperationStatsSnapshot struct {
	RecentHistory []OperationExecution
}
