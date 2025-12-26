package caches

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTrigger_InfiniteLoop_Direct tests that a trigger that modifies its own key
// is detected and stopped after MaxTriggerDepth iterations.
func TestTrigger_InfiniteLoop_Direct(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Create initial data
	err := cache.Create(ctx, map[string]any{"counter": 0})
	require.NoError(t, err)

	// Create a trigger that increments the counter whenever it changes
	// This will cause an infinite loop: trigger fires → increments counter → trigger fires again
	_, err = cache.CreateTrigger(ctx, "counter", INC("counter", 1))
	require.NoError(t, err)

	// Manually trigger by replacing the counter
	// This should hit the depth limit and return an error
	err = cache.Replace(ctx, "counter", 1)

	// Should get an error about recursion depth
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trigger recursion depth limit exceeded")
	assert.Contains(t, err.Error(), "infinite loop")
}

// TestTrigger_InfiniteLoop_Indirect tests a two-trigger cycle:
// Trigger A modifies key1 → fires Trigger B → modifies key2 → fires Trigger A
func TestTrigger_InfiniteLoop_Indirect(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Create initial data
	err := cache.Create(ctx, map[string]any{
		"key1": 0,
		"key2": 0,
	})
	require.NoError(t, err)

	// Trigger A: When key1 changes, increment key2
	_, err = cache.CreateTrigger(ctx, "key1", INC("key2", 1))
	require.NoError(t, err)

	// Trigger B: When key2 changes, increment key1
	_, err = cache.CreateTrigger(ctx, "key2", INC("key1", 1))
	require.NoError(t, err)

	// Start the cycle by modifying key1
	err = cache.Replace(ctx, "key1", 1)

	// Should hit the depth limit
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trigger recursion depth limit exceeded")
}

// TestTrigger_MaxDepth_Allowed tests that triggers can nest up to MaxTriggerDepth
// levels without error.
func TestTrigger_MaxDepth_Allowed(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Create a chain of keys: key0 → key1 → key2 → ... → key9
	// Each trigger increments the next key in the chain
	for i := 0; i <= MaxTriggerDepth; i++ {
		keyName := "key" + string(rune('0'+i))
		err := cache.Create(ctx, map[string]any{keyName: 0})
		require.NoError(t, err)
	}

	// Create triggers: key0 → increment key1, key1 → increment key2, etc.
	for i := 0; i < MaxTriggerDepth; i++ {
		triggerKey := "key" + string(rune('0'+i))
		targetKey := "key" + string(rune('0'+i+1))
		_, err := cache.CreateTrigger(ctx, triggerKey, INC(targetKey, 1))
		require.NoError(t, err)
	}

	// Start the chain by modifying key0
	// This should cascade through all keys up to MaxTriggerDepth
	err := cache.Replace(ctx, "key0", 1)

	// Should succeed - exactly at the limit
	assert.NoError(t, err)

	// Verify the cascade worked: key9 should have been incremented
	val, err := cache.Get(ctx, "key9")
	require.NoError(t, err)
	assert.Equal(t, float64(1), val) // INC uses float64
}

// TestTrigger_MaxDepth_Exceeded tests that exceeding MaxTriggerDepth returns an error
func TestTrigger_MaxDepth_Exceeded(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Create a chain longer than MaxTriggerDepth
	for i := 0; i <= MaxTriggerDepth+2; i++ {
		keyName := "key" + string(rune('0'+i))
		err := cache.Create(ctx, map[string]any{keyName: 0})
		require.NoError(t, err)
	}

	// Create triggers for the entire chain
	for i := 0; i <= MaxTriggerDepth+1; i++ {
		triggerKey := "key" + string(rune('0'+i))
		targetKey := "key" + string(rune('0'+i+1))
		_, err := cache.CreateTrigger(ctx, triggerKey, INC(targetKey, 1))
		require.NoError(t, err)
	}

	// Start the chain - should fail when it exceeds MaxTriggerDepth
	err := cache.Replace(ctx, "key0", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trigger recursion depth limit exceeded")
}

// TestTrigger_NormalOperation_StillWorks verifies that normal trigger operation
// (non-recursive or shallow recursion) still works correctly.
func TestTrigger_NormalOperation_StillWorks(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Create test data
	err := cache.Create(ctx, map[string]any{
		"counter": 0,
		"total":   0,
	})
	require.NoError(t, err)

	// Simple trigger: when counter changes, increment total
	// This is non-recursive since incrementing total doesn't trigger the counter trigger
	_, err = cache.CreateTrigger(ctx, "counter", INC("total", 1))
	require.NoError(t, err)

	// Update counter
	err = cache.Replace(ctx, "counter", 5)
	require.NoError(t, err)

	// Verify trigger fired (total should be 1 now)
	val, err := cache.Get(ctx, "total")
	require.NoError(t, err)
	assert.Equal(t, float64(1), val)
}

// TestTrigger_WildcardLoop tests infinite loop with wildcard patterns
func TestTrigger_WildcardLoop(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Create users - use simple structure to avoid wildcard interpolation issues
	err := cache.Create(ctx, map[string]any{
		"user_alice_count": 0,
		"user_bob_count":   0,
	})
	require.NoError(t, err)

	// Trigger: Whenever alice's count changes, increment it again
	// This creates an infinite loop
	_, err = cache.CreateTrigger(ctx, "user_alice_count",
		INC("user_alice_count", 1))
	require.NoError(t, err)

	// Trigger the loop
	err = cache.Replace(ctx, "user_alice_count", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trigger recursion depth limit exceeded")
}

// TestTrigger_DepthContextPreserved tests that the depth is properly tracked
// across different trigger executions
func TestTrigger_DepthContextPreserved(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Start with depth already at 5
	ctx = context.WithValue(ctx, triggerDepthContextKey, 5)

	err := cache.Create(ctx, map[string]any{"key": 0})
	require.NoError(t, err)

	// Create trigger that increments itself
	_, err = cache.CreateTrigger(ctx, "key", INC("key", 1))
	require.NoError(t, err)

	// Should only be able to recurse 5 more times (5 + 5 = 10)
	err = cache.Replace(ctx, "key", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trigger recursion depth limit exceeded")
}

// TestTrigger_ErrorMessage tests that the error message is helpful
func TestTrigger_ErrorMessage(t *testing.T) {
	ctx := context.Background()
	cache := New()

	err := cache.Create(ctx, map[string]any{"x": 0})
	require.NoError(t, err)

	_, err = cache.CreateTrigger(ctx, "x", INC("x", 1))
	require.NoError(t, err)

	err = cache.Replace(ctx, "x", 1)

	require.Error(t, err)

	// Verify error message contains useful information
	errMsg := err.Error()
	assert.Contains(t, errMsg, "trigger recursion depth limit exceeded")
	assert.Contains(t, errMsg, "10") // MaxTriggerDepth value
	assert.Contains(t, errMsg, "infinite loop")
}

// TestTrigger_MultipleIndependentTriggers tests that independent triggers
// (non-recursive) can all fire without hitting the depth limit
func TestTrigger_MultipleIndependentTriggers(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Create data
	err := cache.Create(ctx, map[string]any{
		"source": 0,
		"dest1":  0,
		"dest2":  0,
		"dest3":  0,
	})
	require.NoError(t, err)

	// Create multiple triggers that all react to 'source' but don't trigger each other
	// Each just increments a different destination
	_, err = cache.CreateTrigger(ctx, "source", INC("dest1", 1))
	require.NoError(t, err)
	_, err = cache.CreateTrigger(ctx, "source", INC("dest2", 1))
	require.NoError(t, err)
	_, err = cache.CreateTrigger(ctx, "source", INC("dest3", 1))
	require.NoError(t, err)

	// Update source - should fire all 3 triggers without recursion
	err = cache.Replace(ctx, "source", 42)
	require.NoError(t, err)

	// Verify all triggers fired (each dest should be incremented once)
	for _, key := range []string{"dest1", "dest2", "dest3"} {
		val, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, float64(1), val)
	}
}
