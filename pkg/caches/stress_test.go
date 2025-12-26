package caches

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStress(t *testing.T) {
	const (
		numCaches  = 10
		numKeys    = 10
		numThreads = 1000
		numActions = 1000
	)

	ctx := context.Background()
	start := time.Now()

	// Prepare caches
	for i := 0; i < numCaches; i++ {
		name := "cache-" + strconv.Itoa(i)
		if err := AddCache(name); err != nil {
			t.Fatal(err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(numThreads)
	for i := 0; i < numThreads; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < numActions; j++ {
				// Pick a random cache
				cacheID := "cache-" + strconv.Itoa(int(randInt(0, numCaches)))
				cache, err := FetchCache(cacheID)
				if err != nil {
					t.Errorf("fetch cache %s error: %v", cacheID, err)
					continue
				}

				tag := uuid.New().String()
				cache.Acquire(tag)

				// Random action
				key := "key-" + strconv.Itoa(int(randInt(0, numKeys)))
				value := randInt(0, 1000000)

				// Randomly choose an operation. Ignore errors.
				switch randInt(0, 5) {
				case 0:
					// Create
					_ = cache.Create(ctx, map[string]any{key: value})
				case 1:
					// Replace
					_ = cache.Replace(ctx, key, value)
				case 2:
					// Batch Replace
					batch := map[string]any{
						key: value,
					}
					_ = cache.ReplaceBatch(ctx, batch)
				case 3:
					// Get
					_, _ = cache.Get(ctx, key)
				case 4:
					// Delete
					_ = cache.Delete(ctx, key)
				}
				
				cache.Release(tag)
			}
		}(i)
	}
	wg.Wait()
	t.Logf("done - %v", time.Since(start))
}

func randInt(min, max int) int64 {
	return int64(min) + int64(rand.Intn(max-min))
}
