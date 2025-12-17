package caches

import (
	"context"
	"strconv"
	"testing"
)

// BenchmarkCache_Create benchmarks creating new keys in the cache
func BenchmarkCache_Create(b *testing.B) {
	ctx := context.Background()
	cache := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key-" + strconv.Itoa(i)
		_ = cache.Create(ctx, map[string]any{key: i})
	}
}

// BenchmarkCache_Get benchmarks retrieving values from the cache
func BenchmarkCache_Get(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create 1000 keys
	for i := 0; i < 1000; i++ {
		key := "key-" + strconv.Itoa(i)
		_ = cache.Create(ctx, map[string]any{key: i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key-" + strconv.Itoa(i%1000)
		_, _ = cache.Get(ctx, key)
	}
}

// BenchmarkCache_Replace benchmarks replacing existing values
func BenchmarkCache_Replace(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create 1000 keys
	for i := 0; i < 1000; i++ {
		key := "key-" + strconv.Itoa(i)
		_ = cache.Create(ctx, map[string]any{key: i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key-" + strconv.Itoa(i%1000)
		_ = cache.Replace(ctx, key, i*2)
	}
}

// BenchmarkCache_Delete benchmarks deleting keys
func BenchmarkCache_Delete(b *testing.B) {
	ctx := context.Background()
	cache := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		key := "key-" + strconv.Itoa(i)
		_ = cache.Create(ctx, map[string]any{key: i})
		b.StartTimer()

		_ = cache.Delete(ctx, key)
	}
}

// BenchmarkCache_Increment benchmarks incrementing numeric values
func BenchmarkCache_Increment(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create counter
	_ = cache.Create(ctx, map[string]any{"counter": 0})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Increment(ctx, "counter", 1)
	}
}

// BenchmarkCache_NestedGet benchmarks retrieving deeply nested values
func BenchmarkCache_NestedGet(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create deeply nested structure
	_ = cache.Create(ctx, map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"value": "deep",
					},
				},
			},
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(ctx, "level1/level2/level3/level4/value")
	}
}

// BenchmarkCache_NestedReplace benchmarks replacing deeply nested values
func BenchmarkCache_NestedReplace(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create deeply nested structure
	_ = cache.Create(ctx, map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": map[string]any{
						"value": "deep",
					},
				},
			},
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Replace(ctx, "level1/level2/level3/level4/value", i)
	}
}

// BenchmarkCache_ReplaceBatch benchmarks batch replacement of multiple keys
func BenchmarkCache_ReplaceBatch(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create 100 keys
	for i := 0; i < 100; i++ {
		key := "key-" + strconv.Itoa(i)
		_ = cache.Create(ctx, map[string]any{key: i})
	}

	// Prepare batch update
	batch := make(map[string]any)
	for i := 0; i < 10; i++ {
		key := "key-" + strconv.Itoa(i)
		batch[key] = i * 2
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.ReplaceBatch(ctx, batch)
	}
}

// BenchmarkCache_ArrayAppend benchmarks appending to arrays
func BenchmarkCache_ArrayAppend(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create empty array
	_ = cache.Create(ctx, map[string]any{"items": []any{}})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.ArrayAppend(ctx, "items", i)
	}
}
