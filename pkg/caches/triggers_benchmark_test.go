package caches

import (
	"context"
	"strconv"
	"testing"
)

// BenchmarkReplace_WithWildcardTrigger measures Replace on a large cache that
// has a wildcard trigger, which runs pattern matching on every write.
func BenchmarkReplace_WithWildcardTrigger(b *testing.B) {
	ctx := context.Background()
	cache := New()
	defer cache.Close()

	domains := map[string]any{}
	for i := 0; i < 10_000; i++ {
		domains[strconv.Itoa(i)] = map[string]any{"countdown": 10, "status": "pending"}
	}
	cache.Acquire("bench")
	defer cache.Release("bench")
	if err := cache.Create(ctx, map[string]any{"domains": domains}); err != nil {
		b.Fatal(err)
	}
	if _, err := cache.CreateTrigger(ctx, "domains/*/countdown", NOOP()); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cache.Replace(ctx, "domains/42/countdown", i); err != nil {
			b.Fatal(err)
		}
	}
}
