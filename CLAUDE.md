# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Build the binary
go build -o map-cache ./cmd/cache/main.go

# Run the server
./map-cache

# Stop the server gracefully (send SIGTERM or SIGINT)
# The server will wait up to 10 seconds for existing requests to complete
pkill -TERM map-cache
# Or press Ctrl+C if running in foreground

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -v -run TestName ./pkg/caches

# Run stress tests (tests concurrent access)
go test -v -run TestStress ./pkg/caches

# Run the big countdown scenario test
go test -v -run Test_Big ./pkg/caches

# Run integration tests
go test -v ./tests/...

# Run all benchmarks (skip regular tests with -run=^$)
go test -bench=. -benchmem -run=^$ ./...

# Run benchmarks for a specific package
go test -bench=. -benchmem -run=^$ ./pkg/caches
go test -bench=. -benchmem -run=^$ ./pkg/containers

# Run specific benchmark
go test -bench=BenchmarkCache_Get -benchmem -run=^$ ./pkg/caches

# Run benchmarks with custom duration (faster)
go test -bench=. -benchmem -run=^$ -benchtime=500ms ./pkg/caches

# Run benchmarks and save results for comparison
go test -bench=. -benchmem -run=^$ ./pkg/caches > bench-old.txt
# After making changes:
go test -bench=. -benchmem -run=^$ ./pkg/caches > bench-new.txt
# Compare with benchstat (install: go install golang.org/x/perf/cmd/benchstat@latest)
benchstat bench-old.txt bench-new.txt

# Note: The -run=^$ flag skips regular tests and only runs benchmarks.
# Without it, you'll see test log output (which includes expected error messages from error-condition tests).
```

## Configuration

The service is configured via environment variables:
- `LISTEN_ADDRESS` - Server address (default: `:8080`)
- `KEY_DELIMITER` - Path delimiter for nested keys (default: `/`)
- `LOG_FORMAT` - Log format: `json` or `text` (default: `json`)

## Architecture Overview

### Core Components

**1. Cache Layer (`pkg/caches/`)**
- `caches_caches.go`: Global cache registry using `sync.Map` for thread-safe cache management
- `cache_cache.go`: Individual `Cache` struct with nested data storage, mutex-based locking, TTL timers, and triggers
- Cache acquisition pattern: All cache operations require acquiring a lock via `Acquire(tag)` and releasing via `Release(tag)` with matching tags to prevent improper concurrent access

**2. Storage Model**
- Each cache uses `containers.Map` (Gabs wrapper) for nested JSON-like data structures
- Keys are path-based with `/` delimiter (e.g., `users/123/profile/name`)
- Values can be any JSON type: objects, arrays, strings, numbers, booleans, null

**3. Command System**
- Commands provide atomic batch operations with conditional logic and loops
- All commands implement the `Command` interface with a `Do(ctx, cache)` method
- **All commands now return values** - INC/REPLACE/DELETE return their new/deleted values
- Command types in `cmd_*.go` files:
  - `INC`: Increment/decrement numeric values (returns new value)
  - `REPLACE`: Overwrite key values (returns new value)
  - `DELETE`: Remove keys with wildcard support (returns deleted value(s))
  - `GET`: Retrieve values with wildcard support
  - `RETURN`: Return computed values (typically final command)
  - `IF`: Conditional execution based on expressions (uses expression caching)
  - `FOR`: Loop over wildcard patterns (e.g., `users/*/name`)
  - `COMMANDS`: Group multiple commands (returns array of results)
  - `PRINT`: Log formatted messages
  - `NOOP`: No-operation placeholder
- Commands execute in transactions via `cache.Execute()` for consistency

**4. Value Interpolation**
- Syntax: `${{key/path}}` for dynamic value references
- Wildcard patterns: `${{users/*/name}}` returns array of matches
- **Optional values** (NEW): `${{key || default}}` provides graceful fallbacks
  - Fallback chain: `${{primary || secondary || default}}`
  - Type-preserving: `${{count || 0}}` returns integer 0
  - Works in templates: `"Hello, ${{name || Guest}}!"`
  - Not allowed with wildcards: `${{users/*/name || unknown}}` errors
- Loop captures: In `FOR` loops, `${{1}}`, `${{2}}` access wildcard segments
- String embedding: `"Status: ${{status}}, Count: ${{count}}"`
- Expression evaluation: Uses `github.com/Knetic/govaluate` for conditionals
  - **Expression caching** (NEW): Compiled expressions cached globally via `sync.Map`
  - Provides 19-76% performance improvement for repeated IF conditions
- Aggregation functions: `all()` and `any()` for pattern-based conditions

**5. Trigger System**
- Triggers in `triggers.go` and `trigger_*.go` execute commands when keys matching patterns are updated
- Pattern-based: Triggers match keys using wildcards (e.g., `domains/*/countdown`)
- Execution: Fires after key updates complete, can cascade to other triggers
- Storage: Per-cache map of pattern -> trigger list, ordered by creation time
- **Infinite Loop Protection**: Trigger recursion is limited to `MaxTriggerDepth` (10 levels) to prevent infinite loops
  - If trigger A fires trigger B which fires trigger A again, the cycle is detected and an error is returned
  - The depth limit ensures the server doesn't crash from runaway trigger chains
  - Example: A trigger that modifies its own watched key will be stopped after 10 iterations

**6. TTL/Expiration**
- Two levels: per-key TTL and cache-level TTL
- `Timer` type in `timer.go` wraps `time.AfterFunc` for cleanup
- Key expirations stored in `cache.keyExps` map
- Automatic cleanup when timers fire or cache/keys are deleted

### API Structure (`internal/api/`)

**v1 API** (`internal/api/v1/`)
- `v1caches/`: Cache CRUD operations
- `v1keys/`: Key CRUD with path-based access
- `v1commands/`: Command execution endpoint
- `v1triggers/`: Trigger management
- `docs/`: Embedded Swagger UI and OpenAPI spec

**Admin API** (`internal/api/admin/`)
- Backup: Serialize entire cache to JSON
- Restore: Load cache from JSON backup

**Middleware Pattern**
- Each API package has middleware to extract cache from `X-Cache-Name` header
- Cache is acquired, stored in context, then released after handler completes
- Prevents deadlocks and ensures proper cleanup

### Testing Strategy

- Unit tests alongside implementation files (`*_test.go`)
- `big_test.go`: Complex scenario test with 100+ domains, cascading triggers, and countdown logic
- `stress_test.go`: Concurrent access tests
- Integration tests in `tests/` directory
- Benchmark tests (`*_benchmark_test.go`):
  - `cache_benchmark_test.go`: Benchmarks for core cache operations (Get, Create, Replace, Delete, Increment, etc.)
  - `cmd_benchmark_test.go`: Benchmarks for command execution (INC, REPLACE, IF, FOR, interpolation, etc.)
  - `gabs_map_benchmark_test.go`: Benchmarks for container operations and wildcard pattern matching
- Use `testify/assert` for assertions

## Key Patterns

### Cache Access Pattern
```go
cache, err := caches.FetchCache(name)
if err != nil {
    return err
}
cache.Acquire("operation-name")
defer cache.Release("operation-name")
// ... perform operations
```

### Command Implementation
- Each command type has its own file: `cmd_<type>.go`
- Commands must be JSON-marshalable (see `cmd_marshaling.go`)
- Use `CmdResult` to return values from command execution
- **All commands now return meaningful values**:
  - INC: returns new value after increment (not nil)
  - REPLACE: returns new value (not nil)
  - DELETE: returns deleted value(s) - array for wildcards
  - GET: returns fetched value(s) - map for wildcards
  - RETURN: returns computed value
  - IF/FOR/COMMANDS: return results from nested commands

### Trigger Pattern Matching
- Wildcards (`*`) match path segments
- Triggers store pattern and compiled command
- On key update, all matching triggers fire in order

### Path-Based Key Access
- Keys use `/` delimiter for nested access
- `user/profile/name` navigates: `cache["user"]["profile"]["name"]`
- Supports creation, retrieval, update, deletion at any depth

## Important Implementation Notes

- **Thread Safety**: All cache operations are mutex-protected; never hold multiple cache locks simultaneously
  - For HTTP API: Middleware automatically handles `Acquire(tag)`/`Release(tag)`
  - For direct library use: **MUST** manually call `Acquire(tag)` and `defer cache.Release(tag)`
- **Tag Enforcement**: Cache acquire/release tags must match; panics on mismatch to catch locking bugs
- **Wildcard Performance**: Pattern matching iterates all keys; use specific paths when possible
- **Trigger Cascading**: Triggers can modify keys that fire other triggers
  - **Infinite Loop Protection**: Recursion limited to 10 levels (`MaxTriggerDepth`)
  - Error returned if limit exceeded: `"trigger recursion depth limit exceeded (max: 10) - possible infinite loop detected"`
  - Safe trigger patterns: Ensure trigger chains don't create cycles (A→B→A)
- **TTL Precision**: TTL values are in milliseconds throughout the codebase
- **Context Propagation**: Commands receive `context.Context` for cancellation support

## External Dependencies

- **Echo**: Web framework (`github.com/labstack/echo/v4`)
- **Gabs**: JSON manipulation (`github.com/Jeffail/gabs/v2`)
- **govaluate**: Expression evaluation (`github.com/Knetic/govaluate`)
- **UUID**: Trigger IDs (`github.com/google/uuid`)
- **logos**: Structured logging (`github.com/goodblaster/logos`)
