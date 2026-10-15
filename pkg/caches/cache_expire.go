package caches

import (
	"context"
	"time"

	"github.com/goodblaster/map-cache/internal/log"
)

// expiredKey identifies a key whose timer fired. The timer is carried along so
// the worker can tell whether the key's TTL was replaced or cancelled after the
// timer fired but before the batch was processed.
type expiredKey struct {
	key   string
	timer *Timer
}

// expirationWorker processes expired keys in batches to prevent goroutine storms.
// This runs in a dedicated goroutine for the lifetime of the cache.
func (cache *Cache) expirationWorker() {
	defer cache.expirationWg.Done()

	// Batch collection parameters
	const (
		maxBatchSize = 100                    // Maximum keys to delete at once
		batchTimeout = 100 * time.Millisecond // Maximum time to wait for batch
	)

	batch := make([]expiredKey, 0, maxBatchSize)
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	processBatch := func() {
		if len(batch) == 0 {
			return
		}

		// Create a background context for deletion
		ctx := context.Background()

		// Acquire lock once for the entire batch
		tag := "ttl-expiration-worker"
		cache.Acquire(tag)

		// Delete all keys in batch
		for _, expired := range batch {
			key := expired.key

			// Skip keys whose TTL was reset or cancelled after this timer fired.
			if cache.keyExps[key] != expired.timer {
				continue
			}

			// Record activity for the batch
			cache.recordActivity()

			// Split the key into path components
			path := SplitKey(key)

			// Delete the key from the underlying map
			if err := cache.cmap.Delete(ctx, path...); err != nil {
				log.WithError(err).With("key", key).Warn("failed to delete expired key")
			}

			// Clean up timer reference (already stopped by FutureFunc)
			delete(cache.keyExps, key)
		}

		cache.Release(tag)

		// Clear batch for reuse
		batch = batch[:0]
	}

	for {
		select {
		case expired := <-cache.expirationChan:
			// Add key to batch
			batch = append(batch, expired)

			// Process if batch is full
			if len(batch) >= maxBatchSize {
				processBatch()
			}

		case <-ticker.C:
			// Process batch on timeout (even if not full)
			processBatch()

		case <-cache.expirationStop:
			// Final batch processing before shutdown
			processBatch()
			return
		}
	}
}

// SetKeyTTL - set the expiration timer for a key.
func (cache *Cache) SetKeyTTL(ctx context.Context, key string, milliseconds int64) error {
	// Cancel existing timer if it exists.
	timer, ok := cache.keyExps[key]
	if ok {
		timer.Stop()
		delete(cache.keyExps, key)
	}

	// Create a new timer that sends the key to the expiration channel.
	// The expiration worker will handle the actual deletion in batches.
	cache.keyExps[key] = futureTimer(milliseconds, func(t *Timer) {
		// Send key to expiration channel (non-blocking)
		select {
		case cache.expirationChan <- expiredKey{key: key, timer: t}:
			// Key sent successfully
		default:
			// Channel full - delete directly as fallback
			// This should be rare with a 1000-key buffer
			if err := cache.Delete(ctx, key); err != nil {
				log.WithError(err).With("key", key).Warn("failed to delete expired key (channel full)")
			}
			delete(cache.keyExps, key)
		}
	})

	return nil
}

// CancelKeyTTL - cancel the expiration timer.
func (cache *Cache) CancelKeyTTL(ctx context.Context, key string) error {
	if timer, ok := cache.keyExps[key]; ok {
		timer.Stop()
		delete(cache.keyExps, key)
	}
	return nil
}
