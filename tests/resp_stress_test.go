package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// checkRedisAvailable skips the test if Redis is not available
func checkRedisAvailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testClient := redis.NewClient(&redis.Options{
		Addr:            "localhost:6379",
		MaxRetries:      1,
		ConnMaxLifetime: time.Second,
	})
	defer testClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := testClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping test: Redis server not available")
	}
}

// TestRESP_ConcurrentConnections tests 100+ simultaneous client connections
func TestRESP_ConcurrentConnections(t *testing.T) {
	checkRedisAvailable(t)

	const numClients = 150
	const opsPerClient = 100

	ctx := context.Background()
	var wg sync.WaitGroup
	var successCount atomic.Int64
	var errorCount atomic.Int64

	// Launch concurrent clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			// Each goroutine gets its own client connection
			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			defer client.Close()

			// Verify connection
			if err := client.Ping(ctx).Err(); err != nil {
				errorCount.Add(1)
				t.Logf("Client %d failed to connect: %v", clientID, err)
				return
			}

			// Perform operations
			for j := 0; j < opsPerClient; j++ {
				key := fmt.Sprintf("concurrent:client%d:op%d", clientID, j)
				value := fmt.Sprintf("value-%d-%d", clientID, j)

				// SET
				if err := client.Set(ctx, key, value, 0).Err(); err != nil {
					errorCount.Add(1)
					continue
				}

				// GET
				result, err := client.Get(ctx, key).Result()
				if err != nil {
					errorCount.Add(1)
					continue
				}

				if result == value {
					successCount.Add(1)
				} else {
					errorCount.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	// Report results
	totalOps := int64(numClients * opsPerClient)
	successRate := float64(successCount.Load()) / float64(totalOps) * 100

	t.Logf("Concurrent test results:")
	t.Logf("  Clients: %d", numClients)
	t.Logf("  Ops per client: %d", opsPerClient)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Successful: %d", successCount.Load())
	t.Logf("  Errors: %d", errorCount.Load())
	t.Logf("  Success rate: %.2f%%", successRate)

	// Assert at least 99% success rate
	assert.GreaterOrEqual(t, successRate, 99.0, "Success rate should be at least 99%")
}

// TestRESP_SustainedLoad tests sustained 10k+ ops/sec
func TestRESP_SustainedLoad(t *testing.T) {
	checkRedisAvailable(t)

	const duration = 10 * time.Second
	const targetOpsPerSec = 10000
	const numWorkers = 50

	ctx := context.Background()
	var wg sync.WaitGroup
	var totalOps atomic.Int64
	var errors atomic.Int64

	stopChan := make(chan struct{})

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			defer client.Close()

			opCount := 0
			for {
				select {
				case <-stopChan:
					return
				default:
					key := fmt.Sprintf("load:worker%d:op%d", workerID, opCount)
					value := fmt.Sprintf("v%d", opCount)

					// Mix of operations
					switch opCount % 4 {
					case 0:
						err := client.Set(ctx, key, value, 0).Err()
						if err != nil {
							errors.Add(1)
						}
					case 1:
						_, err := client.Get(ctx, key).Result()
						if err != nil && err != redis.Nil {
							errors.Add(1)
						}
					case 2:
						err := client.Incr(ctx, fmt.Sprintf("load:counter:%d", workerID)).Err()
						if err != nil {
							errors.Add(1)
						}
					case 3:
						err := client.HSet(ctx, fmt.Sprintf("load:hash:%d", workerID), "field", value).Err()
						if err != nil {
							errors.Add(1)
						}
					}

					totalOps.Add(1)
					opCount++
				}
			}
		}(i)
	}

	// Run for specified duration
	time.Sleep(duration)
	close(stopChan)
	wg.Wait()

	// Calculate results
	total := totalOps.Load()
	errorCount := errors.Load()
	opsPerSec := float64(total) / duration.Seconds()
	errorRate := float64(errorCount) / float64(total) * 100

	t.Logf("Sustained load test results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Workers: %d", numWorkers)
	t.Logf("  Total operations: %d", total)
	t.Logf("  Operations/sec: %.0f", opsPerSec)
	t.Logf("  Errors: %d (%.2f%%)", errorCount, errorRate)

	// Assert we hit target throughput
	assert.GreaterOrEqual(t, opsPerSec, float64(targetOpsPerSec),
		"Should sustain at least %d ops/sec", targetOpsPerSec)
	assert.Less(t, errorRate, 1.0, "Error rate should be less than 1%")
}

// TestRESP_ConnectionPoolExhaustion tests behavior under connection limits
func TestRESP_ConnectionPoolExhaustion(t *testing.T) {
	checkRedisAvailable(t)

	const numConnections = 500
	ctx := context.Background()

	clients := make([]*redis.Client, numConnections)
	var wg sync.WaitGroup
	var successfulConnections atomic.Int64

	// Try to open many connections simultaneously
	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			clients[idx] = client

			// Try to execute a command
			if err := client.Ping(ctx).Err(); err == nil {
				successfulConnections.Add(1)
			}
		}(i)
	}

	wg.Wait()

	// Clean up
	for _, client := range clients {
		if client != nil {
			client.Close()
		}
	}

	successCount := successfulConnections.Load()
	successRate := float64(successCount) / float64(numConnections) * 100

	t.Logf("Connection pool test results:")
	t.Logf("  Attempted connections: %d", numConnections)
	t.Logf("  Successful connections: %d", successCount)
	t.Logf("  Success rate: %.2f%%", successRate)

	// Server should handle at least 90% of connections
	assert.GreaterOrEqual(t, successRate, 90.0,
		"Server should handle at least 90%% of connection attempts")
}

// TestRESP_TTLExpirationAtScale tests TTL expiration with 10k+ keys
func TestRESP_TTLExpirationAtScale(t *testing.T) {
	checkRedisAvailable(t)

	const numKeys = 10000
	const ttlSeconds = 5

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx := context.Background()

	// Clean up first
	client.FlushDB(ctx)

	start := time.Now()

	// Create many keys with TTL
	t.Logf("Creating %d keys with %ds TTL...", numKeys, ttlSeconds)
	pipeline := client.Pipeline()
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("ttl:key:%d", i)
		pipeline.Set(ctx, key, fmt.Sprintf("value%d", i), time.Duration(ttlSeconds)*time.Second)
	}
	_, err := pipeline.Exec(ctx)
	assert.NoError(t, err)

	creationTime := time.Since(start)
	t.Logf("Created %d keys in %v (%.0f keys/sec)", numKeys, creationTime,
		float64(numKeys)/creationTime.Seconds())

	// Verify all keys exist
	existingKeys := 0
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("ttl:key:%d", i)
		exists, _ := client.Exists(ctx, key).Result()
		existingKeys += int(exists)
	}
	t.Logf("Verified %d/%d keys exist", existingKeys, numKeys)
	assert.Equal(t, numKeys, existingKeys, "All keys should exist initially")

	// Wait for TTL to expire (plus buffer)
	sleepDuration := time.Duration(ttlSeconds+2) * time.Second
	t.Logf("Waiting %v for keys to expire...", sleepDuration)
	time.Sleep(sleepDuration)

	// Verify keys have expired
	remainingKeys := 0
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("ttl:key:%d", i)
		exists, _ := client.Exists(ctx, key).Result()
		remainingKeys += int(exists)
	}

	expirationRate := float64(numKeys-remainingKeys) / float64(numKeys) * 100
	t.Logf("Expiration results:")
	t.Logf("  Expired: %d/%d keys", numKeys-remainingKeys, numKeys)
	t.Logf("  Remaining: %d keys", remainingKeys)
	t.Logf("  Expiration rate: %.2f%%", expirationRate)

	// At least 99% should have expired
	assert.GreaterOrEqual(t, expirationRate, 99.0,
		"At least 99%% of keys should expire")
}

// TestRESP_MixedWorkloadStress tests realistic mixed operations under load
func TestRESP_MixedWorkloadStress(t *testing.T) {
	checkRedisAvailable(t)

	const duration = 10 * time.Second
	const numWorkers = 20

	ctx := context.Background()
	var wg sync.WaitGroup
	var stats struct {
		sets     atomic.Int64
		gets     atomic.Int64
		incrs    atomic.Int64
		hsets    atomic.Int64
		lpushes  atomic.Int64
		deletes  atomic.Int64
		errors   atomic.Int64
	}

	stopChan := make(chan struct{})

	// Start workers with mixed operations
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			defer client.Close()

			opCount := 0
			for {
				select {
				case <-stopChan:
					return
				default:
					// Realistic workload: 40% reads, 30% writes, 30% complex ops
					op := opCount % 10
					key := fmt.Sprintf("mixed:w%d:k%d", workerID, opCount/10)

					var err error
					switch op {
					case 0, 1, 2, 3: // 40% - GETs
						_, err = client.Get(ctx, key).Result()
						if err != nil && err != redis.Nil {
							stats.errors.Add(1)
						} else {
							stats.gets.Add(1)
						}
					case 4, 5, 6: // 30% - SETs
						err = client.Set(ctx, key, fmt.Sprintf("v%d", opCount), 0).Err()
						if err != nil {
							stats.errors.Add(1)
						} else {
							stats.sets.Add(1)
						}
					case 7: // 10% - INCRs
						err = client.Incr(ctx, fmt.Sprintf("mixed:counter:%d", workerID)).Err()
						if err != nil {
							stats.errors.Add(1)
						} else {
							stats.incrs.Add(1)
						}
					case 8: // 10% - HSETs
						err = client.HSet(ctx, fmt.Sprintf("mixed:hash:%d", workerID),
							fmt.Sprintf("field%d", opCount), "value").Err()
						if err != nil {
							stats.errors.Add(1)
						} else {
							stats.hsets.Add(1)
						}
					case 9: // 10% - LPUSHes
						err = client.LPush(ctx, fmt.Sprintf("mixed:list:%d", workerID),
							fmt.Sprintf("item%d", opCount)).Err()
						if err != nil {
							stats.errors.Add(1)
						} else {
							stats.lpushes.Add(1)
						}
					}

					opCount++
				}
			}
		}(i)
	}

	// Run for specified duration
	time.Sleep(duration)
	close(stopChan)
	wg.Wait()

	// Calculate totals
	totalOps := stats.sets.Load() + stats.gets.Load() + stats.incrs.Load() +
		stats.hsets.Load() + stats.lpushes.Load()
	opsPerSec := float64(totalOps) / duration.Seconds()
	errorRate := float64(stats.errors.Load()) / float64(totalOps+stats.errors.Load()) * 100

	t.Logf("Mixed workload stress test results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Workers: %d", numWorkers)
	t.Logf("  Total operations: %d (%.0f ops/sec)", totalOps, opsPerSec)
	t.Logf("  Breakdown:")
	t.Logf("    SETs:    %d", stats.sets.Load())
	t.Logf("    GETs:    %d", stats.gets.Load())
	t.Logf("    INCRs:   %d", stats.incrs.Load())
	t.Logf("    HSETs:   %d", stats.hsets.Load())
	t.Logf("    LPUSHes: %d", stats.lpushes.Load())
	t.Logf("  Errors: %d (%.2f%%)", stats.errors.Load(), errorRate)

	assert.Less(t, errorRate, 1.0, "Error rate should be less than 1%")
}

// TestRESP_ConcurrentHashOperations tests concurrent access to same hash
func TestRESP_ConcurrentHashOperations(t *testing.T) {
	checkRedisAvailable(t)

	const numWorkers = 50
	const opsPerWorker = 100
	const hashKey = "concurrent:shared:hash"

	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	// Clean up
	client.Del(ctx, hashKey)

	var wg sync.WaitGroup
	var successCount atomic.Int64

	// Multiple workers hammering the same hash
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			defer client.Close()

			for j := 0; j < opsPerWorker; j++ {
				field := fmt.Sprintf("worker%d:field%d", workerID, j)
				value := fmt.Sprintf("value%d", j)

				// HSET
				if err := client.HSet(ctx, hashKey, field, value).Err(); err != nil {
					continue
				}

				// HGET
				result, err := client.HGet(ctx, hashKey, field).Result()
				if err == nil && result == value {
					successCount.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify hash integrity
	hashLen, err := client.HLen(ctx, hashKey).Result()
	assert.NoError(t, err)

	totalOps := int64(numWorkers * opsPerWorker)
	successRate := float64(successCount.Load()) / float64(totalOps) * 100

	t.Logf("Concurrent hash operations results:")
	t.Logf("  Workers: %d", numWorkers)
	t.Logf("  Operations per worker: %d", opsPerWorker)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Successful: %d (%.2f%%)", successCount.Load(), successRate)
	t.Logf("  Final hash size: %d fields", hashLen)

	assert.GreaterOrEqual(t, successRate, 99.0, "Success rate should be at least 99%")
	assert.Equal(t, totalOps, hashLen, "Hash should contain all fields")
}
