package containers

import (
	"context"
	"strconv"
	"testing"
)

// BenchmarkGabsMap_Get benchmarks retrieving values
func BenchmarkGabsMap_Get(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	_ = gMap.Set(ctx, map[string]any{
		"user": map[string]any{
			"name":  "Alice",
			"email": "alice@example.com",
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gMap.Get(ctx, "user", "name")
	}
}

// BenchmarkGabsMap_Set benchmarks setting values
func BenchmarkGabsMap_Set(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gMap.Set(ctx, i, "key")
	}
}

// BenchmarkGabsMap_SetNested benchmarks setting deeply nested values
func BenchmarkGabsMap_SetNested(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gMap.Set(ctx, i, "level1", "level2", "level3", "value")
	}
}

// BenchmarkGabsMap_Exists benchmarks checking if keys exist
func BenchmarkGabsMap_Exists(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	_ = gMap.Set(ctx, map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gMap.Exists(ctx, "a")
	}
}

// BenchmarkGabsMap_Delete benchmarks deleting keys
func BenchmarkGabsMap_Delete(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		key := "key-" + strconv.Itoa(i)
		_ = gMap.Set(ctx, i, key)
		b.StartTimer()

		_ = gMap.Delete(ctx, key)
	}
}

// BenchmarkGabsMap_WildKeys_SingleWildcard benchmarks wildcard matching with one wildcard
func BenchmarkGabsMap_WildKeys_SingleWildcard(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create 100 users
	users := make(map[string]any)
	for i := 0; i < 100; i++ {
		users["user-"+strconv.Itoa(i)] = map[string]any{"name": "User" + strconv.Itoa(i)}
	}
	_ = gMap.Set(ctx, users, "users")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gMap.WildKeys(ctx, "users/*/name")
	}
}

// BenchmarkGabsMap_WildKeys_MultipleWildcards benchmarks wildcard matching with multiple wildcards
func BenchmarkGabsMap_WildKeys_MultipleWildcards(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create nested structure
	data := make(map[string]any)
	for i := 0; i < 10; i++ {
		group := make(map[string]any)
		for j := 0; j < 10; j++ {
			group["item-"+strconv.Itoa(j)] = map[string]any{"value": j}
		}
		data["group-"+strconv.Itoa(i)] = group
	}
	_ = gMap.Set(ctx, data, "data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gMap.WildKeys(ctx, "data/*/*/value")
	}
}

// BenchmarkGabsMap_WildKeys_NoMatch benchmarks wildcard matching with no results
func BenchmarkGabsMap_WildKeys_NoMatch(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	_ = gMap.Set(ctx, map[string]any{
		"users": map[string]any{
			"alice": map[string]any{"name": "Alice"},
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gMap.WildKeys(ctx, "nonexistent/*/value")
	}
}

// BenchmarkGabsMap_ArrayAppend benchmarks appending to arrays
func BenchmarkGabsMap_ArrayAppend(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	_ = gMap.Set(ctx, []any{}, "items")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gMap.ArrayAppend(ctx, i, "items")
	}
}

// BenchmarkGabsMap_ArrayRemove benchmarks removing from arrays
func BenchmarkGabsMap_ArrayRemove(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup: Create array with 100 items
		items := make([]any, 100)
		for j := 0; j < 100; j++ {
			items[j] = j
		}
		_ = gMap.Set(ctx, items, "items")
		b.StartTimer()

		_ = gMap.ArrayRemove(ctx, 50, "items")
	}
}

// BenchmarkGabsMap_ArrayResize_Grow benchmarks growing arrays
func BenchmarkGabsMap_ArrayResize_Grow(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		_ = gMap.Set(ctx, []any{1, 2, 3}, "items")
		b.StartTimer()

		_ = gMap.ArrayResize(ctx, 100, "items")
	}
}

// BenchmarkGabsMap_ArrayResize_Shrink benchmarks shrinking arrays
func BenchmarkGabsMap_ArrayResize_Shrink(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		items := make([]any, 100)
		for j := 0; j < 100; j++ {
			items[j] = j
		}
		_ = gMap.Set(ctx, items, "items")
		b.StartTimer()

		_ = gMap.ArrayResize(ctx, 10, "items")
	}
}

// BenchmarkGabsMap_Data benchmarks getting the full data map
func BenchmarkGabsMap_Data(b *testing.B) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create large dataset
	data := make(map[string]any)
	for i := 0; i < 1000; i++ {
		data["key-"+strconv.Itoa(i)] = i
	}
	_ = gMap.Set(ctx, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gMap.Data(ctx)
	}
}
