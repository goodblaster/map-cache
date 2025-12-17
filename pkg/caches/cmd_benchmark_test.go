package caches

import (
	"context"
	"strconv"
	"testing"
)

// BenchmarkCmd_INC benchmarks the INC command
func BenchmarkCmd_INC(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup
	_ = cache.Create(ctx, map[string]any{"counter": 0.0})

	cmd := INC("counter", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Do(ctx, cache)
	}
}

// BenchmarkCmd_REPLACE benchmarks the REPLACE command
func BenchmarkCmd_REPLACE(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup
	_ = cache.Create(ctx, map[string]any{"key": "initial"})

	cmd := REPLACE("key", "updated")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Do(ctx, cache)
	}
}

// BenchmarkCmd_IF benchmarks the IF command with simple condition
func BenchmarkCmd_IF(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup
	_ = cache.Create(ctx, map[string]any{"value": 50})

	cmd := IF(
		"${{value}} > 10",
		REPLACE("status", "high"),
		REPLACE("status", "low"),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Do(ctx, cache)
	}
}

// BenchmarkCmd_FOR benchmarks the FOR loop command
func BenchmarkCmd_FOR(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create 10 items
	items := make(map[string]any)
	for i := 0; i < 10; i++ {
		items["item-"+strconv.Itoa(i)] = map[string]any{"count": 0.0}
	}
	_ = cache.Create(ctx, map[string]any{"items": items})

	cmd := FOR(
		"${{items/*/count}}",
		INC("items/${{1}}/count", 1),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Do(ctx, cache)
	}
}

// BenchmarkCmd_ExecuteSequence benchmarks executing a sequence of commands
func BenchmarkCmd_ExecuteSequence(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup
	_ = cache.Create(ctx, map[string]any{
		"counter": 0.0,
		"status":  "idle",
	})

	commands := []Command{
		INC("counter", 1),
		IF(
			"${{counter}} > 5",
			REPLACE("status", "active"),
			NOOP(),
		),
		RETURN("${{counter}}"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Execute(ctx, commands...)
	}
}

// BenchmarkCmd_InterpolateSimple benchmarks simple value interpolation
func BenchmarkCmd_InterpolateSimple(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup
	_ = cache.Create(ctx, map[string]any{
		"name":  "Alice",
		"count": 42,
	})

	cmd := RETURN("User ${{name}} has ${{count}} items")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Do(ctx, cache)
	}
}

// BenchmarkCmd_InterpolateWildcard benchmarks wildcard interpolation
func BenchmarkCmd_InterpolateWildcard(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create 10 users
	users := make(map[string]any)
	for i := 0; i < 10; i++ {
		users["user-"+strconv.Itoa(i)] = map[string]any{"name": "User" + strconv.Itoa(i)}
	}
	_ = cache.Create(ctx, map[string]any{"users": users})

	cmd := RETURN("${{users/*/name}}")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Do(ctx, cache)
	}
}

// BenchmarkCmd_ComplexScenario benchmarks a complex real-world scenario
func BenchmarkCmd_ComplexScenario(b *testing.B) {
	ctx := context.Background()
	cache := New()

	// Setup: Create domain structure
	domains := make(map[string]any)
	for i := 0; i < 5; i++ {
		domains["domain-"+strconv.Itoa(i)] = map[string]any{
			"countdown": 10.0,
			"status":    "busy",
		}
	}
	_ = cache.Create(ctx, map[string]any{
		"domains": domains,
		"overall": "running",
	})

	commands := []Command{
		FOR(
			"${{domains/*/countdown}}",
			INC("domains/${{1}}/countdown", -1),
			IF(
				"${{domains/${{1}}/countdown}} <= 0",
				REPLACE("domains/${{1}}/status", "complete"),
				NOOP(),
			),
		),
		IF(
			"all(${{domains/*/status}} == \"complete\")",
			REPLACE("overall", "complete"),
			NOOP(),
		),
		RETURN("${{overall}}"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Execute(ctx, commands...)
	}
}
