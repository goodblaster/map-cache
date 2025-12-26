package caches

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goodblaster/map-cache/pkg/containers"
)

type Cache struct {
	cmap          containers.Map
	mutex         *sync.Mutex
	tag           *string              // who owns this
	exp           *Timer               // expiration timer
	expMillis     *int64               // TTL in milliseconds (for stats)
	keyExps       map[string]*Timer    // key-based expiration timers
	triggers      map[string][]Trigger // key-based triggers
	lastAccessed  *time.Time           // last access timestamp
	activityCount atomic.Int64         // count of operations (thread-safe)
	opStats       *OperationStats      // long-running operation tracking
}

func New() *Cache {
	return &Cache{
		cmap:     containers.NewGabsMap(),
		mutex:    &sync.Mutex{},
		keyExps:  map[string]*Timer{},
		triggers: map[string][]Trigger{},
		opStats:  NewOperationStats(100), // Keep last 100 long operations
	}
}

// SizeBytes returns the approximate size of the cache data in bytes.
// This method is NOT thread-safe - caller must acquire the cache lock first.
func (cache *Cache) SizeBytes(ctx context.Context) int {
	data := cache.cmap.Data(ctx)
	if data == nil {
		return 0
	}

	// Marshal to JSON to get byte size
	bytes, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	return len(bytes)
}

// TTLMillis returns the cache-level TTL in milliseconds, or nil if not set.
// This method is NOT thread-safe - caller must acquire the cache lock first.
func (cache *Cache) TTLMillis() *int64 {
	return cache.expMillis
}

// LastAccessed returns the last access timestamp, or nil if never accessed.
// This method is NOT thread-safe - caller must acquire the cache lock first.
func (cache *Cache) LastAccessed() *time.Time {
	return cache.lastAccessed
}

// ActivityCount returns the total number of operations performed on this cache.
// This method IS thread-safe (uses atomic operations).
func (cache *Cache) ActivityCount() int64 {
	return cache.activityCount.Load()
}

// recordActivity increments the activity counter and updates last accessed time.
// This should be called for any cache operation.
// This method is NOT thread-safe for lastAccessed - caller must hold the lock.
func (cache *Cache) recordActivity() {
	cache.activityCount.Add(1)
	now := time.Now()
	cache.lastAccessed = &now
}

// WildKeys returns all keys matching the wildcard pattern.
// This method is NOT thread-safe - caller must acquire the cache lock first.
func (cache *Cache) WildKeys(ctx context.Context, pattern string) []string {
	return cache.cmap.WildKeys(ctx, pattern)
}

// KeyExpirations returns the map of keys with expiration timers.
// This method is NOT thread-safe - caller must acquire the cache lock first.
func (cache *Cache) KeyExpirations() map[string]*Timer {
	return cache.keyExps
}

// Triggers returns the map of trigger patterns to their triggers.
// This method is NOT thread-safe - caller must acquire the cache lock first.
func (cache *Cache) Triggers() map[string][]Trigger {
	return cache.triggers
}

// Expiration returns the cache-level expiration timer, or nil if not set.
// This method is NOT thread-safe - caller must acquire the cache lock first.
func (cache *Cache) Expiration() *Timer {
	return cache.exp
}

// RecordLongOperation records a long-running operation.
// This method IS thread-safe (delegates to OperationStats which handles locking).
func (cache *Cache) RecordLongOperation(duration time.Duration, operation string, success bool, timedOut bool) {
	cache.opStats.RecordLongOperation(duration, operation, success, timedOut)
}

// OperationStatsSnapshot returns the operation statistics for this cache.
// This method IS thread-safe (OperationStats uses internal locking).
func (cache *Cache) OperationStatsSnapshot() OperationStatsSnapshot {
	return cache.opStats.GetStats()
}
