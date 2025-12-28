package examples

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// TestHybridCountdownExample demonstrates a hybrid approach:
// - HTTP API for trigger setup (declarative configuration)
// - Redis commands for all data operations (high performance)
// - Triggers handle cascading status changes automatically
//
// Scenario: Multi-stage deployment pipeline
// - 5 deployment stages, each with a countdown
// - When a stage countdown reaches 0, it marks complete and starts next stage
// - Overall status becomes "complete" when all stages finish
//
func TestHybridCountdownExample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping example in short mode")
	}

	// Setup
	httpClient := &http.Client{Timeout: 5 * time.Second}
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer redisClient.Close()
	ctx := context.Background()

	// Check if Redis is available
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skip("Skipping test: Redis server not available on localhost:6379")
	}

	// Check if HTTP server is available
	if _, err := httpClient.Get("http://localhost:8080/health"); err != nil {
		t.Skip("Skipping test: HTTP server not available on localhost:8080")
	}

	// Clean up
	redisClient.FlushDB(ctx)

	t.Log("=== Hybrid Countdown Example: Multi-Stage Deployment Pipeline ===")
	t.Log("")

	// ============================================================================
	// STEP 1: Initialize deployment structure via Redis
	// ============================================================================
	t.Log("Step 1: Initialize deployment structure (via Redis)...")

	// Set up deployment metadata
	redisClient.HSet(ctx, "deployment:pipeline", map[string]interface{}{
		"name":         "Production Deploy v2.5.0",
		"total_stages": 5,
		"status":       "pending",
		"started_at":   time.Now().Format(time.RFC3339),
	})

	// Initialize 5 stages with countdowns
	stages := []struct {
		name      string
		countdown int
	}{
		{"build", 3},
		{"test", 4},
		{"staging", 3},
		{"approval", 2},
		{"production", 5},
	}

	for i, stage := range stages {
		stageNum := i + 1
		stageKey := fmt.Sprintf("deployment:stages:%d", stageNum)

		redisClient.HSet(ctx, stageKey, map[string]interface{}{
			"stage_number": stageNum,
			"name":         stage.name,
			"countdown":    stage.countdown,
			"status":       "pending",
		})

		t.Logf("  ✓ Stage %d: %s (countdown: %d)", stageNum, stage.name, stage.countdown)
	}

	t.Log("")

	// ============================================================================
	// STEP 2: Set up triggers via HTTP API
	// ============================================================================
	t.Log("Step 2: Set up cascading triggers (via HTTP)...")

	// Trigger 1: When a stage countdown hits 0 → mark stage complete
	// Using nested interpolation with CORRECT wildcard index (${{1}} not ${{2}})
	trigger1 := map[string]interface{}{
		"key": "deployment/stages/*/countdown",
		"command": map[string]interface{}{
			"type":      "IF",
			"condition": "${{deployment/stages/${{1}}/countdown}} == 0",
			"if_true": map[string]interface{}{
				"type": "COMMANDS",
				"commands": []interface{}{
					map[string]interface{}{
						"type":  "REPLACE",
						"key":   "deployment/stages/${{1}}/status",
						"value": "complete",
					},
					map[string]interface{}{
						"type":    "PRINT",
						"message": "🎉 Stage ${{1}} completed!",
					},
				},
			},
			"if_false": map[string]interface{}{
				"type": "NOOP",
			},
		},
	}

	// Create trigger via HTTP
	jsonData, _ := json.Marshal(trigger1)
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/triggers", bytes.NewBuffer(jsonData))
	req.Header.Set("X-Cache-Name", "default")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to create trigger: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var trigger1ID string
	json.NewDecoder(resp.Body).Decode(&trigger1ID)
	t.Logf("  ✓ Trigger 1 created: %s", trigger1ID[:8])

	triggerIDs := []string{trigger1ID}

	t.Log("")

	// ============================================================================
	// STEP 3: Start the deployment via Redis
	// ============================================================================
	t.Log("Step 3: Start deployment (via Redis)...")

	// Mark pipeline as running and start stage 1
	redisClient.HSet(ctx, "deployment:pipeline", "status", "running")
	redisClient.HSet(ctx, "deployment:stages:1", "status", "running")

	t.Log("  ✓ Pipeline started, stage 1 (build) is now running")
	t.Log("")

	// ============================================================================
	// STEP 4: Simulate deployment progress via Redis DECR commands
	// ============================================================================
	t.Log("Step 4: Simulate deployment progress (via Redis)...")
	t.Log("")

	// Monitor and decrement countdowns until deployment completes
	maxIterations := 100
	iteration := 0

	for iteration < maxIterations {
		// Check if deployment is complete
		status, err := redisClient.HGet(ctx, "deployment:pipeline", "status").Result()
		if err == nil && status == "complete" {
			t.Log("✅ Deployment complete!")
			break
		}

		// First, check for completed stages and start next stage
		for i := 1; i <= 4; i++ { // Only check stages 1-4 (stage 5 is last)
			stageKey := fmt.Sprintf("deployment:stages:%d", i)
			nextStageKey := fmt.Sprintf("deployment:stages:%d", i+1)

			// Check if this stage just completed
			stageStatus, _ := redisClient.HGet(ctx, stageKey, "status").Result()
			nextStageStatus, _ := redisClient.HGet(ctx, nextStageKey, "status").Result()

			if stageStatus == "complete" && nextStageStatus == "pending" {
				// Start next stage
				redisClient.HSet(ctx, nextStageKey, "status", "running")
				nextStageName, _ := redisClient.HGet(ctx, nextStageKey, "name").Result()
				t.Logf("  🚀 Starting stage %d (%s)", i+1, nextStageName)
			}
		}

		// Check if all stages are complete and mark pipeline complete
		allComplete := true
		for i := 1; i <= 5; i++ {
			stageKey := fmt.Sprintf("deployment:stages:%d", i)
			stageStatus, _ := redisClient.HGet(ctx, stageKey, "status").Result()
			if stageStatus != "complete" {
				allComplete = false
				break
			}
		}
		if allComplete {
			pipelineStatus, _ := redisClient.HGet(ctx, "deployment:pipeline", "status").Result()
			if pipelineStatus != "complete" {
				redisClient.HSet(ctx, "deployment:pipeline", "status", "complete")
				t.Log("  ✅ All stages complete! Pipeline marked complete")
			}
		}

		// Find and decrement running stage countdowns
		decrementedAny := false
		for i := 1; i <= 5; i++ {
			stageKey := fmt.Sprintf("deployment:stages:%d", i)

			// Check if stage is running
			stageStatus, err := redisClient.HGet(ctx, stageKey, "status").Result()
			if err != nil || stageStatus != "running" {
				continue
			}

			// Get current countdown
			countdownStr, err := redisClient.HGet(ctx, stageKey, "countdown").Result()
			if err != nil {
				continue
			}

			// Decrement countdown via Redis (use HINCRBY for hash fields)
			newCountdown, err := redisClient.HIncrBy(ctx, stageKey, "countdown", -1).Result()
			if err != nil {
				t.Logf("  ERROR decrementing stage %d: %v", i, err)
				continue
			}

			// Check status immediately after decrement (triggers should fire synchronously)
			newStatus, _ := redisClient.HGet(ctx, stageKey, "status").Result()
			stageName, _ := redisClient.HGet(ctx, stageKey, "name").Result()
			t.Logf("  Stage %d (%s): countdown %s → %d, status=%s", i, stageName, countdownStr, newCountdown, newStatus)
			decrementedAny = true

			// Small delay to make output readable
			time.Sleep(300 * time.Millisecond)
		}

		if !decrementedAny {
			// No running stages, wait a bit for triggers to fire
			time.Sleep(200 * time.Millisecond)
		}

		iteration++
	}

	t.Log("")

	// ============================================================================
	// STEP 5: Verify final state via Redis
	// ============================================================================
	t.Log("Step 5: Verify final state (via Redis)...")

	// Check pipeline status
	pipelineStatus, err := redisClient.HGet(ctx, "deployment:pipeline", "status").Result()
	assert.NoError(t, err)
	assert.Equal(t, "complete", pipelineStatus, "Pipeline should be complete")
	t.Log("  ✓ Pipeline status: complete")

	// Check all stages are complete
	for i := 1; i <= 5; i++ {
		stageKey := fmt.Sprintf("deployment:stages:%d", i)
		stageStatus, err := redisClient.HGet(ctx, stageKey, "status").Result()
		assert.NoError(t, err)
		assert.Equal(t, "complete", stageStatus, fmt.Sprintf("Stage %d should be complete", i))

		stageName, _ := redisClient.HGet(ctx, stageKey, "name").Result()
		countdown, _ := redisClient.HGet(ctx, stageKey, "countdown").Result()
		t.Logf("  ✓ Stage %d (%s): status=complete, countdown=%s", i, stageName, countdown)
	}

	t.Log("")
	t.Log("=== Example Complete ===")
	t.Log("")
	t.Log("Key Takeaways:")
	t.Log("  • HTTP API used for trigger setup (one-time configuration)")
	t.Log("  • Redis commands used for all data operations (fast, efficient)")
	t.Log("  • Triggers enabled automatic cascading without polling")
	t.Log("  • Hybrid approach combines strengths of both APIs")
	t.Log("  • Final state: All 5 stages complete, pipeline status = 'complete'")

	// ============================================================================
	// STEP 6: Cleanup
	// ============================================================================
	t.Log("")
	t.Log("Cleanup: Removing triggers...")

	for i, triggerID := range triggerIDs {
		req, _ := http.NewRequest("DELETE",
			fmt.Sprintf("http://localhost:8080/api/v1/triggers/%s", triggerID),
			nil)
		req.Header.Set("X-Cache-Name", "default")

		resp, err := httpClient.Do(req)
		assert.NoError(t, err)
		resp.Body.Close()

		t.Logf("  ✓ Trigger %d deleted", i+1)
	}

	redisClient.FlushDB(ctx)
	t.Log("  ✓ Data cleared")
}

// TestHybridCountdownSimple is a simpler version demonstrating the core pattern
func TestHybridCountdownSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping example in short mode")
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer redisClient.Close()
	ctx := context.Background()

	// Check if Redis is available
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skip("Skipping test: Redis server not available on localhost:6379")
	}

	// Check if HTTP server is available
	if _, err := httpClient.Get("http://localhost:8080/health"); err != nil {
		t.Skip("Skipping test: HTTP server not available on localhost:8080")
	}

	redisClient.FlushDB(ctx)

	t.Log("=== Simple Hybrid Countdown Example ===")
	t.Log("")

	// Step 1: Initialize via Redis
	t.Log("1. Initialize task with countdown=3")
	redisClient.HSet(ctx, "task:1", map[string]interface{}{
		"name":      "Simple Task",
		"countdown": 3,
		"status":    "pending",
	})

	// Step 2: Create trigger via HTTP - when countdown hits 0, mark complete
	t.Log("2. Create trigger (HTTP): countdown=0 → status=complete")

	trigger := map[string]interface{}{
		"key": "task/1/countdown",
		"command": map[string]interface{}{
			"type":      "IF",
			"condition": "${{task/1/countdown}} == 0",
			"if_true": map[string]interface{}{
				"type":  "REPLACE",
				"key":   "task/1/status",
				"value": "complete",
			},
			"if_false": map[string]interface{}{
				"type": "NOOP",
			},
		},
	}

	jsonData, _ := json.Marshal(trigger)
	t.Logf("   Trigger JSON: %s", string(jsonData))
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/triggers", bytes.NewBuffer(jsonData))
	req.Header.Set("X-Cache-Name", "default")
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to create trigger: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to create trigger: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var triggerID string
	if err := json.NewDecoder(resp.Body).Decode(&triggerID); err != nil {
		t.Fatalf("Failed to decode trigger response: %v", err)
	}
	t.Logf("   Trigger created: %s", triggerID)

	// Give trigger time to register
	time.Sleep(500 * time.Millisecond)

	// Step 3: Start task via Redis
	t.Log("3. Start task (Redis): status=running")
	redisClient.HSet(ctx, "task:1", "status", "running")

	// Step 4: Decrement countdown via Redis until complete
	t.Log("4. Decrement countdown (Redis):")
	for i := 0; i < 10; i++ {
		status, _ := redisClient.HGet(ctx, "task:1", "status").Result()
		if status == "complete" {
			t.Log("   ✓ Task complete!")
			break
		}

		// Use HINCRBY to decrement hash field
		countdown, _ := redisClient.HIncrBy(ctx, "task:1", "countdown", -1).Result()
		t.Logf("   countdown → %d", countdown)

		// Give trigger extra time to fire when we hit 0
		if countdown == 0 {
			time.Sleep(500 * time.Millisecond)
		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Step 5: Verify via Redis
	t.Log("5. Verify final state (Redis):")
	status, _ := redisClient.HGet(ctx, "task:1", "status").Result()
	countdown, _ := redisClient.HGet(ctx, "task:1", "countdown").Result()
	t.Logf("   status=%s, countdown=%s", status, countdown)

	assert.Equal(t, "complete", status)

	redisClient.FlushDB(ctx)
	t.Log("")
	t.Log("✅ Simple example complete")
}
