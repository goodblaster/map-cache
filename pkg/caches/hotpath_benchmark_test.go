package caches

import (
	"context"
	"fmt"
	"strconv"
	"testing"
)

// Benchmark hot paths - the most frequently used operations

func BenchmarkCacheHotPaths(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Pre-populate with data for read benchmarks
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		_ = cache.Create(ctx, map[string]any{key: i})
	}

	b.Run("Get/simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(ctx, "key500")
		}
	})

	b.Run("Get/nested", func(b *testing.B) {
		_ = cache.Create(ctx, map[string]any{
			"users": map[string]any{
				"123": map[string]any{
					"profile": map[string]any{
						"name": "John",
					},
				},
			},
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(ctx, "users/123/profile/name")
		}
	})

	b.Run("Replace/simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cache.Replace(ctx, "key500", i)
		}
	})

	b.Run("Replace/nested", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cache.Replace(ctx, "users/123/profile/name", "Jane")
		}
	})

	b.Run("Create/simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("newkey%d", i)
			_ = cache.Create(ctx, map[string]any{key: i})
		}
	})

	b.Run("Delete/simple", func(b *testing.B) {
		// Pre-create keys
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delkey%d", i)
			_ = cache.Create(ctx, map[string]any{key: i})
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delkey%d", i)
			_ = cache.Delete(ctx, key)
		}
	})

	b.Run("Increment/number", func(b *testing.B) {
		_ = cache.Create(ctx, map[string]any{"counter": 0})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cache.Increment(ctx, "counter", 1)
		}
	})
}

func BenchmarkCommandExecution(b *testing.B) {
	ctx := context.Background()
	cache := New()

	_ = cache.Create(ctx, map[string]any{
		"value":   10,
		"status":  "pending",
		"counter": 0,
	})

	b.Run("REPLACE", func(b *testing.B) {
		cmd := REPLACE("value", 20)
		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("GET", func(b *testing.B) {
		cmd := GET("value")
		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("INC", func(b *testing.B) {
		cmd := INC("counter", 1)
		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("IF/simple", func(b *testing.B) {
		cmd := IF(
			`${{value}} > 5`,
			REPLACE("status", "high"),
			REPLACE("status", "low"),
		)
		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("RETURN/simple", func(b *testing.B) {
		cmd := RETURN("value")
		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("RETURN/interpolation", func(b *testing.B) {
		cmd := RETURN("${{value}}")
		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("COMMANDS/chain", func(b *testing.B) {
		cmd := COMMANDS(
			REPLACE("value", 100),
			INC("counter", 1),
			RETURN("${{counter}}"),
		)
		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})
}

func BenchmarkWildcardOperations(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Create data with wildcard patterns
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("users/%d/status", i)
		_ = cache.Create(ctx, map[string]any{key: "active"})
	}

	b.Run("FOR/wildcard", func(b *testing.B) {
		cmd := FOR(
			"${{users/*/status}}",
			REPLACE("users/${{1}}/status", "inactive"),
		)

		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("RETURN/wildcard", func(b *testing.B) {
		cmd := RETURN("${{users/*/status}}")

		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("IF/any", func(b *testing.B) {
		cmd := IF(
			`any(${{users/*/status}} == "active")`,
			NOOP(),
			NOOP(),
		)

		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})

	b.Run("IF/all", func(b *testing.B) {
		cmd := IF(
			`all(${{users/*/status}} == "inactive")`,
			NOOP(),
			NOOP(),
		)

		for i := 0; i < b.N; i++ {
			cmd.Do(ctx, cache)
		}
	})
}

func BenchmarkConcurrentAccess(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Pre-populate
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		_ = cache.Create(ctx, map[string]any{key: i})
	}

	b.Run("Concurrent/Get", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key%d", i%100)
				cache.Acquire("bench")
				_, _ = cache.Get(ctx, key)
				cache.Release("bench")
				i++
			}
		})
	})

	b.Run("Concurrent/Replace", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key%d", i%100)
				cache.Acquire("bench")
				_ = cache.Replace(ctx, key, i)
				cache.Release("bench")
				i++
			}
		})
	})

	b.Run("Concurrent/Mixed", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key%d", i%100)
				cache.Acquire("bench")
				if i%2 == 0 {
					_, _ = cache.Get(ctx, key)
				} else {
					_ = cache.Replace(ctx, key, i)
				}
				cache.Release("bench")
				i++
			}
		})
	})
}

func BenchmarkGlobalCacheOperations(b *testing.B) {
	b.Run("AddCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			name := "cache" + strconv.Itoa(i)
			_ = AddCache(name)
		}
	})

	b.Run("FetchCache", func(b *testing.B) {
		_ = AddCache("benchmark")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = FetchCache("benchmark")
		}
	})

	b.Run("DeleteCache", func(b *testing.B) {
		// Pre-create caches
		for i := 0; i < b.N; i++ {
			name := "delcache" + strconv.Itoa(i)
			_ = AddCache(name)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			name := "delcache" + strconv.Itoa(i)
			_ = DeleteCache(name)
		}
	})
}

func BenchmarkArrayOperations(b *testing.B) {
	ctx := context.Background()
	cache := New()

	_ = cache.Create(ctx, map[string]any{
		"items": []any{1, 2, 3, 4, 5},
	})

	b.Run("ArrayAppend", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cache.ArrayAppend(ctx, "items", i)
		}
	})

	b.Run("ArrayResize/grow", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cache.ArrayResize(ctx, "items", 100)
		}
	})

	b.Run("ArrayResize/shrink", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cache.ArrayResize(ctx, "items", 5)
		}
	})
}
