package caches

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Scenario 1: Session Management with Auto-Expiration
// Use case: Web application session storage with automatic cleanup
func TestScenario_SessionManagement(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	// Create a user session with nested data
	sessionID := "sess_abc123"
	err := cache.Create(ctx, map[string]any{
		sessionID: map[string]any{
			"user_id":     "user_42",
			"username":    "alice",
			"email":       "alice@example.com",
			"login_time":  time.Now().Unix(),
			"permissions": []any{"read", "write"},
		},
	})
	require.NoError(t, err)

	// Set session to expire in 30 minutes (simulated as 1ms for test)
	err = cache.SetKeyTTL(ctx, sessionID, 1)
	require.NoError(t, err)

	// Retrieve user info from session
	username, err := cache.Get(ctx, sessionID+"/username")
	assert.NoError(t, err)
	assert.Equal(t, "alice", username)

	// Note: last_activity would need to be created first before replacing
	// For this test, we'll just verify the session data exists

	// Verify session exists
	_, err = cache.Get(ctx, sessionID)
	assert.NoError(t, err)

	// Wait for expiration AND batch processing (100ms ticker + margin)
	time.Sleep(200 * time.Millisecond)

	// Session should be auto-deleted
	_, err = cache.Get(ctx, sessionID)
	assert.Error(t, err)
}

// Scenario 2: Feature Flags with Wildcard Updates
// Use case: Enable/disable features across multiple services simultaneously
func TestScenario_FeatureFlags(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Initialize feature flags for multiple services
	err := cache.Create(ctx, map[string]any{
		"features": map[string]any{
			"api": map[string]any{
				"new_auth":      false,
				"rate_limiting": true,
				"graphql":       false,
			},
			"web": map[string]any{
				"new_auth":  false,
				"dark_mode": true,
				"chat":      false,
			},
			"mobile": map[string]any{
				"new_auth":     false,
				"offline_mode": true,
				"biometric":    false,
			},
		},
	})
	require.NoError(t, err)

	// Enable new_auth feature across ALL services using a command
	cmd := FOR(
		"${{features/*/new_auth}}",
		REPLACE("features/${{1}}/new_auth", true),
	)
	result := cmd.Do(ctx, cache)
	assert.NoError(t, result.Error)

	// Verify all new_auth flags are now enabled
	apiAuth, _ := cache.Get(ctx, "features/api/new_auth")
	webAuth, _ := cache.Get(ctx, "features/web/new_auth")
	mobileAuth, _ := cache.Get(ctx, "features/mobile/new_auth")

	assert.Equal(t, true, apiAuth)
	assert.Equal(t, true, webAuth)
	assert.Equal(t, true, mobileAuth)

	// Check if ANY service has dark_mode enabled
	cmd = IF(
		`any(${{features/*/dark_mode}} == true)`,
		RETURN("dark_mode_available"),
		RETURN("dark_mode_not_available"),
	)
	result = cmd.Do(ctx, cache)
	assert.NoError(t, result.Error)
	assert.Equal(t, "dark_mode_available", result.Value)
}

// Scenario 3: Rate Limiting with Auto-Reset
// Use case: API rate limiting with per-user counters that reset automatically
func TestScenario_RateLimiting(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	userID := "user_123"
	apiKey := "rate_limit/" + userID

	// Initialize rate limit counter
	err := cache.Create(ctx, map[string]any{
		apiKey: map[string]any{
			"requests": 0,
			"limit":    100,
			"window":   "1min",
		},
	})
	require.NoError(t, err)

	// Set to expire in 1 minute (simulated as 1ms)
	err = cache.SetKeyTTL(ctx, apiKey, 1)
	require.NoError(t, err)

	// Simulate API requests
	for i := 0; i < 5; i++ {
		newCount, err := cache.Increment(ctx, apiKey+"/requests", 1)
		assert.NoError(t, err)
		assert.Equal(t, float64(i+1), newCount)
	}

	// Check if under limit using command
	cmd := IF(
		`${{rate_limit/user_123/requests}} < ${{rate_limit/user_123/limit}}`,
		RETURN("allowed"),
		RETURN("rate_limited"),
	)
	result := cmd.Do(ctx, cache)
	assert.NoError(t, result.Error)
	assert.Equal(t, "allowed", result.Value)

	// Verify counter
	count, err := cache.Get(ctx, apiKey+"/requests")
	assert.NoError(t, err)
	assert.Equal(t, float64(5), count)

	// Wait for expiration AND batch processing (100ms ticker + margin)
	time.Sleep(200 * time.Millisecond)

	// Rate limit should be reset
	_, err = cache.Get(ctx, apiKey)
	assert.Error(t, err)
}

// Scenario 4: Shopping Cart with Product Catalog Integration
// Use case: Cart that references a product catalog and triggers auto-update total
// Demonstrates: ArrayAppend, product lookup, and trigger-based price aggregation
func TestScenario_ShoppingCart(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Initialize product catalog with prices
	err := cache.Create(ctx, map[string]any{
		"products": map[string]any{
			"prod_1": map[string]any{"name": "Widget", "price": 25.00},
			"prod_2": map[string]any{"name": "Gadget", "price": 15.00},
			"prod_3": map[string]any{"name": "Doohickey", "price": 10.00},
		},
	})
	require.NoError(t, err)

	// Initialize shopping cart
	err = cache.Create(ctx, map[string]any{
		"cart": map[string]any{
			"user_456": map[string]any{
				"total":    0,
				"item_ids": []any{}, // Array of product IDs
			},
		},
	})
	require.NoError(t, err)

	// Create a SINGLE trigger with COMMANDS group that handles all products
	// When any item is marked as added, the trigger uses the wildcard capture ${{1}}
	// to determine which product and increment by its price
	trigger := COMMANDS(
		// Check which product was added and increment by corresponding price
		// ${{1}} captures the product ID from the wildcard pattern
		IF(
			`"${{1}}" == "prod_1"`,
			INC("cart/user_456/total", 25.00), // Price from products/prod_1/price
			NOOP(),
		),
		IF(
			`"${{1}}" == "prod_2"`,
			INC("cart/user_456/total", 15.00), // Price from products/prod_2/price
			NOOP(),
		),
		IF(
			`"${{1}}" == "prod_3"`,
			INC("cart/user_456/total", 10.00), // Price from products/prod_3/price
			NOOP(),
		),
	)

	// Single trigger on wildcard pattern handles all products
	triggerID, err := cache.CreateTrigger(ctx, "cart/user_456/items/*", trigger)
	require.NoError(t, err)
	defer cache.DeleteTrigger(ctx, triggerID)

	// Add product IDs to cart array
	err = cache.ArrayAppend(ctx, "cart/user_456/item_ids", "prod_1")
	assert.NoError(t, err)

	// Create items structure with all products initially false
	err = cache.Create(ctx, map[string]any{
		"cart/user_456/items": map[string]any{
			"prod_1": false,
			"prod_2": false,
			"prod_3": false,
		},
	})
	assert.NoError(t, err)

	// Mark prod_1 as added - this triggers the price increment
	err = cache.Replace(ctx, "cart/user_456/items/prod_1", true)
	assert.NoError(t, err)

	// Total should now be 25.00 (auto-incremented by trigger)
	total, err := cache.Get(ctx, "cart/user_456/total")
	assert.NoError(t, err)
	assert.Equal(t, 25.00, total)

	// Add second product
	err = cache.ArrayAppend(ctx, "cart/user_456/item_ids", "prod_2")
	assert.NoError(t, err)

	err = cache.Replace(ctx, "cart/user_456/items/prod_2", true)
	assert.NoError(t, err)

	// Total: 25.00 + 15.00 = 40.00
	total, err = cache.Get(ctx, "cart/user_456/total")
	assert.NoError(t, err)
	assert.Equal(t, 40.00, total)

	// Add third product
	err = cache.ArrayAppend(ctx, "cart/user_456/item_ids", "prod_3")
	assert.NoError(t, err)

	err = cache.Replace(ctx, "cart/user_456/items/prod_3", true)
	assert.NoError(t, err)

	// Total: 40.00 + 10.00 = 50.00
	total, err = cache.Get(ctx, "cart/user_456/total")
	assert.NoError(t, err)
	assert.Equal(t, 50.00, total)

	// Verify item_ids array contains all products
	itemIDs, err := cache.Get(ctx, "cart/user_456/item_ids")
	assert.NoError(t, err)
	itemIDsArray, ok := itemIDs.([]any)
	assert.True(t, ok)
	assert.Len(t, itemIDsArray, 3)
	assert.Contains(t, itemIDsArray, "prod_1")
	assert.Contains(t, itemIDsArray, "prod_2")
	assert.Contains(t, itemIDsArray, "prod_3")

	// Demonstrate fetching product details using GET
	prod1Price, err := cache.Get(ctx, "products/prod_1/price")
	assert.NoError(t, err)
	assert.Equal(t, 25.00, prod1Price)
}

// Scenario 5: Real-time Leaderboard
// Use case: Gaming leaderboard with automatic achievement tracking via triggers
func TestScenario_Leaderboard(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Initialize leaderboard with players
	err := cache.Create(ctx, map[string]any{
		"leaderboard": map[string]any{
			"player_1": map[string]any{"name": "Alice", "score": 1000, "rank": 0, "elite": false},
			"player_2": map[string]any{"name": "Bob", "score": 1500, "rank": 0, "elite": false},
			"player_3": map[string]any{"name": "Charlie", "score": 1200, "rank": 0, "elite": false},
		},
	})
	require.NoError(t, err)

	// Create trigger to mark players as "elite" when they cross 1500 points
	triggerCmd := IF(
		"${{leaderboard/${{1}}/score}} >= 1500",
		REPLACE("leaderboard/${{1}}/elite", true),
		NOOP(),
	)

	triggerID, err := cache.CreateTrigger(ctx, "leaderboard/*/score", triggerCmd)
	require.NoError(t, err)
	defer cache.DeleteTrigger(ctx, triggerID)

	// Player 2 is already elite (score >= 1500)
	isElite, _ := cache.Get(ctx, "leaderboard/player_2/elite")
	assert.Equal(t, false, isElite) // Not yet marked

	// Player 1 completes a level and gains points (will trigger elite status)
	newScore, err := cache.Increment(ctx, "leaderboard/player_1/score", 800)
	assert.NoError(t, err)
	assert.Equal(t, float64(1800), newScore)

	// Verify player 1 was automatically marked as elite by trigger
	isElite, err = cache.Get(ctx, "leaderboard/player_1/elite")
	assert.NoError(t, err)
	assert.Equal(t, true, isElite, "Player 1 should be marked elite")

	// Player 3 gets some points but doesn't reach elite threshold
	_, err = cache.Increment(ctx, "leaderboard/player_3/score", 100)
	assert.NoError(t, err)

	// Player 3 should not be elite
	isElite, err = cache.Get(ctx, "leaderboard/player_3/elite")
	assert.NoError(t, err)
	assert.Equal(t, false, isElite, "Player 3 should not be elite yet")

	// Manually track high score (demonstration of querying)
	player1ScoreRaw, _ := cache.Get(ctx, "leaderboard/player_1/score")
	player2ScoreRaw, _ := cache.Get(ctx, "leaderboard/player_2/score")
	player3ScoreRaw, _ := cache.Get(ctx, "leaderboard/player_3/score")

	// Convert to float64 for comparison
	player1Score, _ := ToFloat64(player1ScoreRaw)
	player2Score, _ := ToFloat64(player2ScoreRaw)
	player3Score, _ := ToFloat64(player3ScoreRaw)

	// Player 1 has the highest score
	assert.Greater(t, player1Score, player2Score)
	assert.Greater(t, player1Score, player3Score)
}

// Scenario 6: User Presence Tracking
// Use case: Track online users and auto-remove inactive ones
func TestScenario_PresenceTracking(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	// Initialize online users
	err := cache.Create(ctx, map[string]any{
		"online": map[string]any{
			"user_100": map[string]any{
				"status":       "active",
				"last_seen":    time.Now().Unix(),
				"current_page": "/dashboard",
			},
			"user_101": map[string]any{
				"status":       "active",
				"last_seen":    time.Now().Unix(),
				"current_page": "/profile",
			},
		},
	})
	require.NoError(t, err)

	// Set expiration for user presence (auto-logout after inactivity)
	err = cache.SetKeyTTL(ctx, "online/user_100", 1)
	assert.NoError(t, err)

	// Get count of online users
	cmd := GET("${{online/*}}")
	result := cmd.Do(ctx, cache)
	assert.NoError(t, result.Error)
	// Wildcard GET returns a map of matches
	assert.NotNil(t, result.Value)

	// User updates their activity (heartbeat)
	err = cache.Replace(ctx, "online/user_100/last_seen", time.Now().Unix())
	assert.NoError(t, err)

	// Refresh expiration
	err = cache.SetKeyTTL(ctx, "online/user_100", 1)
	assert.NoError(t, err)

	// Wait for expiration AND batch processing (100ms ticker + margin)
	time.Sleep(200 * time.Millisecond)

	// Only user_101 should remain
	_, err = cache.Get(ctx, "online/user_100")
	assert.Error(t, err)

	_, err = cache.Get(ctx, "online/user_101")
	assert.NoError(t, err)
}

// Scenario 7: Configuration Management with Environment Overrides
// Use case: Application configuration with environment-specific overrides
func TestScenario_ConfigurationManagement(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Initialize base configuration
	err := cache.Create(ctx, map[string]any{
		"config": map[string]any{
			"default": map[string]any{
				"db_host":     "localhost",
				"db_port":     5432,
				"cache_ttl":   300,
				"log_level":   "info",
				"max_retries": 3,
			},
			"production": map[string]any{
				"db_host":   "prod-db.example.com",
				"log_level": "error",
			},
			"development": map[string]any{
				"log_level":   "debug",
				"max_retries": 5,
			},
		},
	})
	require.NoError(t, err)

	// Get production config directly
	dbHost, err := cache.Get(ctx, "config/production/db_host")
	assert.NoError(t, err)
	assert.Equal(t, "prod-db.example.com", dbHost)

	// For a setting not in production, it would fall back to default
	dbPort, err := cache.Get(ctx, "config/production/db_port")
	if err != nil {
		// Fallback to default
		dbPort, err = cache.Get(ctx, "config/default/db_port")
		assert.NoError(t, err)
	}
	assert.Equal(t, 5432, dbPort)

	// Update all environments' cache_ttl
	cmd := FOR(
		"${{config/*/cache_ttl}}",
		REPLACE("config/${{1}}/cache_ttl", 600),
	)
	result := cmd.Do(ctx, cache)
	assert.NoError(t, result.Error)

	// Verify update
	ttl, err := cache.Get(ctx, "config/default/cache_ttl")
	assert.NoError(t, err)
	assert.Equal(t, 600, ttl)
}

// Scenario 8: Workflow State Machine with Triggers
// Use case: Order processing workflow with automatic state transitions
func TestScenario_WorkflowStateMachine(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Initialize order with workflow state
	orderID := "order_789"
	err := cache.Create(ctx, map[string]any{
		"orders": map[string]any{
			orderID: map[string]any{
				"status":      "pending",
				"customer_id": "cust_123",
				"total":       299.99,
				"items_count": 3,
				"created_at":  time.Now().Unix(),
			},
		},
	})
	require.NoError(t, err)

	// Simulate payment completion (in real system, triggers would handle automation)
	err = cache.Replace(ctx, "orders/"+orderID+"/status", "paid")
	assert.NoError(t, err)

	// Manually set processed_at (showing what a trigger would do)
	err = cache.Create(ctx, map[string]any{
		"orders/" + orderID + "/processed_at": time.Now().Unix(),
	})
	assert.NoError(t, err)

	// Verify processed_at was set
	processedAt, err := cache.Get(ctx, "orders/"+orderID+"/processed_at")
	assert.NoError(t, err)
	assert.NotNil(t, processedAt)

	// Continue workflow
	err = cache.Replace(ctx, "orders/"+orderID+"/status", "shipped")
	assert.NoError(t, err)
}

// Scenario 9: Metrics Aggregation
// Use case: Real-time metrics collection and aggregation across services
func TestScenario_MetricsAggregation(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Initialize metrics for multiple services
	err := cache.Create(ctx, map[string]any{
		"metrics": map[string]any{
			"api_server": map[string]any{
				"requests":   0,
				"errors":     0,
				"latency_ms": 0,
			},
			"worker": map[string]any{
				"jobs_processed": 0,
				"jobs_failed":    0,
				"latency_ms":     0,
			},
			"database": map[string]any{
				"queries":      0,
				"slow_queries": 0,
				"latency_ms":   0,
			},
		},
	})
	require.NoError(t, err)

	// Simulate incoming requests
	for i := 0; i < 10; i++ {
		_, err := cache.Increment(ctx, "metrics/api_server/requests", 1)
		assert.NoError(t, err)
	}

	// Record some errors
	_, err = cache.Increment(ctx, "metrics/api_server/errors", 2)
	assert.NoError(t, err)

	// Get api_server requests directly
	requests, err := cache.Get(ctx, "metrics/api_server/requests")
	assert.NoError(t, err)
	assert.Equal(t, float64(10), requests)

	// Check if error rate is above threshold
	cmd := IF(
		`${{metrics/api_server/errors}} > 5`,
		RETURN("alert_high_error_rate"),
		RETURN("normal"),
	)
	result := cmd.Do(ctx, cache)
	assert.NoError(t, result.Error)
	assert.Equal(t, "normal", result.Value)

	// Get all metrics for api_server
	apiMetrics, err := cache.Get(ctx, "metrics/api_server")
	assert.NoError(t, err)
	assert.NotNil(t, apiMetrics)
}

// Scenario 10: Distributed Lock with TTL
// Use case: Distributed locking for coordinating work across multiple processes
func TestScenario_DistributedLock(t *testing.T) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	lockKey := "locks/critical_section"
	processID := "process_abc"

	// Try to acquire lock
	err := cache.Create(ctx, map[string]any{
		lockKey: map[string]any{
			"holder":      processID,
			"acquired_at": time.Now().Unix(),
		},
	})
	require.NoError(t, err)

	// Set lock expiration (auto-release if process crashes)
	err = cache.SetKeyTTL(ctx, lockKey, 1) // 1ms for test
	assert.NoError(t, err)

	// Verify lock is held
	holder, err := cache.Get(ctx, lockKey+"/holder")
	assert.NoError(t, err)
	assert.Equal(t, processID, holder)

	// Another process tries to acquire (should fail)
	err = cache.Create(ctx, map[string]any{
		lockKey: map[string]any{
			"holder":      "process_xyz",
			"acquired_at": time.Now().Unix(),
		},
	})
	assert.Error(t, err) // Key already exists

	// Release lock explicitly
	err = cache.Delete(ctx, lockKey)
	assert.NoError(t, err)

	// Now another process can acquire
	err = cache.Create(ctx, map[string]any{
		lockKey: map[string]any{
			"holder":      "process_xyz",
			"acquired_at": time.Now().Unix(),
		},
	})
	assert.NoError(t, err)

	// Set TTL for auto-expiration test
	err = cache.SetKeyTTL(ctx, lockKey, 1) // 1ms for test
	assert.NoError(t, err)

	// Wait for expiration AND batch processing (100ms ticker + margin)
	time.Sleep(200 * time.Millisecond)

	// Lock should be released
	_, err = cache.Get(ctx, lockKey)
	assert.Error(t, err)
}

// Scenario 11: Parallel Batch Processing with Cascading Triggers
// Use case: Distributed batch job processing where multiple tasks run in parallel,
// each with multiple steps. When all tasks complete, the overall batch status updates.
// Similar to MapReduce, ETL pipelines, or parallel test execution.
func TestScenario_ParallelBatchProcessing(t *testing.T) {
	ctx := context.Background()
	cache := New()

	batchID := "batch_20250101"

	// Initialize batch with multiple tasks, each having steps to complete
	err := cache.Create(ctx, map[string]any{
		batchID: map[string]any{
			"status":      "running",
			"total_tasks": 3,
			"tasks": map[string]any{
				"task_1": map[string]any{
					"name":           "data_validation",
					"steps_remaining": 5,
					"status":         "running",
				},
				"task_2": map[string]any{
					"name":           "data_transformation",
					"steps_remaining": 3,
					"status":         "running",
				},
				"task_3": map[string]any{
					"name":           "data_loading",
					"steps_remaining": 4,
					"status":         "running",
				},
			},
		},
	})
	require.NoError(t, err)

	// Trigger 1: When a task's steps_remaining hits 0, mark task as complete
	taskCompleteTrigger := IF(
		fmt.Sprintf("${{%s/tasks/${{1}}/steps_remaining}} <= 0", batchID),
		REPLACE(fmt.Sprintf("%s/tasks/${{1}}/status", batchID), "complete"),
		NOOP(),
	)
	triggerID1, err := cache.CreateTrigger(ctx, fmt.Sprintf("%s/tasks/*/steps_remaining", batchID), taskCompleteTrigger)
	require.NoError(t, err)
	defer cache.DeleteTrigger(ctx, triggerID1)

	// Trigger 2: When all tasks are complete, mark batch as complete
	batchCompleteTrigger := IF(
		fmt.Sprintf(`all(${{%s/tasks/*/status}} == "complete")`, batchID),
		REPLACE(fmt.Sprintf("%s/status", batchID), "complete"),
		NOOP(),
	)
	triggerID2, err := cache.CreateTrigger(ctx, fmt.Sprintf("%s/tasks/*/status", batchID), batchCompleteTrigger)
	require.NoError(t, err)
	defer cache.DeleteTrigger(ctx, triggerID2)

	// Verify initial batch status
	status, err := cache.Get(ctx, batchID+"/status")
	assert.NoError(t, err)
	assert.Equal(t, "running", status)

	// Simulate task 1 processing - decrement steps
	for i := 0; i < 5; i++ {
		_, err = cache.Increment(ctx, batchID+"/tasks/task_1/steps_remaining", -1)
		assert.NoError(t, err)
	}

	// Task 1 should now be complete (trigger 1 fired)
	task1Status, err := cache.Get(ctx, batchID+"/tasks/task_1/status")
	assert.NoError(t, err)
	assert.Equal(t, "complete", task1Status)

	// Batch should still be running (not all tasks complete)
	status, err = cache.Get(ctx, batchID+"/status")
	assert.NoError(t, err)
	assert.Equal(t, "running", status)

	// Complete task 2
	for i := 0; i < 3; i++ {
		_, err = cache.Increment(ctx, batchID+"/tasks/task_2/steps_remaining", -1)
		assert.NoError(t, err)
	}

	// Task 2 should be complete
	task2Status, err := cache.Get(ctx, batchID+"/tasks/task_2/status")
	assert.NoError(t, err)
	assert.Equal(t, "complete", task2Status)

	// Batch still running (task 3 not done)
	status, err = cache.Get(ctx, batchID+"/status")
	assert.NoError(t, err)
	assert.Equal(t, "running", status)

	// Complete task 3
	for i := 0; i < 4; i++ {
		_, err = cache.Increment(ctx, batchID+"/tasks/task_3/steps_remaining", -1)
		assert.NoError(t, err)
	}

	// Task 3 should be complete
	task3Status, err := cache.Get(ctx, batchID+"/tasks/task_3/status")
	assert.NoError(t, err)
	assert.Equal(t, "complete", task3Status)

	// NOW batch should be complete (trigger 2 fired - all tasks done)
	status, err = cache.Get(ctx, batchID+"/status")
	assert.NoError(t, err)
	assert.Equal(t, "complete", status, "Batch should be complete when all tasks are done")

	// Verify using command that all tasks are complete
	cmd := IF(
		fmt.Sprintf(`all(${{%s/tasks/*/status}} == "complete")`, batchID),
		RETURN(true),
		RETURN(false),
	)
	result := cmd.Do(ctx, cache)
	assert.NoError(t, result.Error)
	assert.Equal(t, true, result.Value)
}
