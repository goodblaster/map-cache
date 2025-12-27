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

### RESP Protocol Implementation (`internal/resp/`)

**Architecture Overview**
- RESP server runs in parallel with HTTP server, sharing the same cache layer
- Implements RESP2 protocol using `github.com/tidwall/resp` library
- Supports **62 core Redis commands** across strings, hashes, lists, keys, and generic operations
- Enabled via `RESP_ENABLED=true` environment variable (default port: 6379)

**Package Structure**
- `server.go`: TCP listener and connection management
- `protocol.go`: RESP2 encoding/decoding utilities
- `session.go`: Per-connection state (selected cache, connection ID)
- `mapper.go`: Key translation utilities (Redis `:` ↔ map-cache `/`)
- `commands_*.go`: Command handlers organized by category:
  - `commands_string.go`: 20 string commands (GET, SET, INCR, GETRANGE, etc.)
  - `commands_hash.go`: 13 hash commands (HGET, HSET, HINCRBY, etc.)
  - `commands_list.go`: 12 list commands (LPUSH, RPUSH, LTRIM, LPOS, etc.)
  - `commands_key.go`: 11 key commands (EXPIRE, TTL, RENAME, TYPE, etc.)
  - `commands_generic.go`: 6 generic commands (PING, ECHO, SELECT, etc.)

**Command Handler Pattern**
```go
func HandleCommandName(s *Session, args []respProto.Value) error {
    // 1. Validate arguments
    if len(args) != expectedCount {
        return s.WriteError("ERR wrong number of arguments")
    }

    // 2. Translate Redis key to map-cache path
    key := TranslateKey(args[0].String())

    // 3. Fetch cache from session
    cache, err := caches.FetchCache(s.SelectedCache())
    if err != nil {
        return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
    }

    // 4. Acquire cache lock with unique tag
    tag := s.Tag("COMMANDNAME")
    cache.Acquire(tag)
    defer cache.Release(tag)

    // 5. Create context with timeout
    ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
    defer cancel()

    // 6. Perform cache operation
    result, err := cache.Operation(ctx, key, value)

    // 7. Return RESP-encoded response
    return s.WriteValue(BulkString(result))
}
```

**Key Translation**
- Automatic bidirectional translation between Redis and map-cache key formats
- Redis convention: `user:123:name` (colon-separated)
- Map-cache convention: `user/123/name` (slash-separated nested paths)
- `TranslateKey()` function handles conversion transparently
- Configurable via `RESP_KEY_MODE` (default: `translate`)

**Session Management**
- Each TCP connection has a `Session` struct tracking:
  - Connection ID (for logging and tagging)
  - Selected cache name (default: "default", changeable via SELECT command)
  - RESP reader/writer for protocol encoding
- Sessions automatically clean up on disconnect

**Command Registration**
- Commands registered in `init()` functions via `RegisterCommand(name, handler)`
- Dispatcher routes incoming RESP commands to registered handlers
- Unrecognized commands return proper RESP error

**Data Type Mapping**
- Redis strings → JSON strings/numbers
- Redis hashes → Nested JSON objects (e.g., `user:1` hash → `user/1/*` paths)
- Redis lists → JSON arrays
- Redis sets → Not supported (hash tables incompatible with JSON)
- Redis sorted sets → Not supported (skip lists incompatible with JSON)

**Testing Approach**
- Integration tests in `tests/resp_new_commands_test.go`
- Uses official `github.com/redis/go-redis/v9` client library
- Each test function:
  1. Creates fresh Redis client connection
  2. Cleans up test keys via `cleanupKeys()` helper
  3. Executes command operations
  4. Validates responses using `testify/assert`
- All 62 commands have test coverage

**Performance Characteristics**
- Binary RESP protocol ~3-10x faster than HTTP/JSON
- Expected latency: 0.1-1ms per operation (vs 1-5ms for HTTP)
- Still slower than native Redis (0.05-0.2ms) due to additional abstraction layers
- Shared cache layer means HTTP and RESP clients see same data instantly

**Key Implementation Details**
- **LPOS command**: Supports RANK, COUNT, MAXLEN options; returns array when COUNT specified
- **RENAME/RENAMENX**: Preserves TTL when renaming keys
- **TYPE command**: Detects JSON type and returns appropriate Redis type (string/hash/list/none)
- **GETEX**: Supports EX, PX, EXAT, PXAT, PERSIST options for flexible expiration control
- **INCRBYFLOAT/HINCRBYFLOAT**: Handles type coercion from int/float/string

For complete Redis protocol documentation, see [REDIS.md](REDIS.md).

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

## RESP Protocol Support

Map-cache implements the Redis Serialization Protocol (RESP2) to support standard Redis clients. The RESP server runs alongside the HTTP server and shares the same cache registry.

### Architecture

**RESP Server (`internal/resp/`)**:
- `server.go`: TCP listener on port 6379 with connection management
- `session.go`: Per-connection state (selected cache, connection ID, context)
- `handler.go`: Command dispatcher with registry pattern
- `protocol.go`: RESP encoding/decoding wrappers (uses `tidwall/resp` library)
- `mapper.go`: Key translation utilities (`:` ↔ `/` conversion)
- `commands_string.go`: String command handlers (GET, SET, INCR, etc.)
- `commands_key.go`: Key management handlers (EXPIRE, TTL, KEYS, etc.)
- `commands_hash.go`: Hash command handlers (HGET, HSET, HGETALL, etc.)
- `commands_list.go`: List command handlers (LPUSH, RPUSH, LPOP, LRANGE, etc.)

### Key Translation

Redis keys use `:` as separator (e.g., `user:123:name`), while map-cache uses `/` (e.g., `user/123/name`). The RESP server automatically translates between these formats in both directions.

**Translation mode** controlled by `RESP_KEY_MODE`:
- `translate` (default): Convert `:` to `/` and vice versa
- `preserve`: Use keys as-is without translation

### Command Implementation Pattern

All RESP commands follow this pattern:

```go
func HandleCommand(s *Session, args []respProto.Value) error {
    // 1. Validate arguments
    if len(args) != expectedCount {
        return s.WriteError("ERR wrong number of arguments")
    }

    // 2. Translate key from Redis format
    key := TranslateKey(args[0].String())

    // 3. Fetch cache
    cache, err := caches.FetchCache(s.SelectedCache())
    if err != nil {
        return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
    }

    // 4. Acquire lock with tag
    tag := s.Tag("COMMAND_NAME")
    cache.Acquire(tag)
    defer cache.Release(tag)

    // 5. Create context with timeout
    ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
    defer cancel()

    // 6. Execute cache operation
    value, err := cache.Get(ctx, key)
    if err != nil {
        return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
    }

    // 7. Return RESP-encoded response
    return s.WriteValue(ConvertToRESP(value))
}
```

### Hash Command Mapping

Hash commands map naturally to map-cache's nested paths:
- `HSET user:100 name "Alice"` → Create/replace at path `user/100/name`
- `HGET user:100 name` → Get value at path `user/100/name`
- `HGETALL user:100` → Get entire object at path `user/100`

This leverages map-cache's native support for nested JSON structures.

### Multi-Cache Support (SELECT)

The `SELECT` command maps Redis database numbers to cache names:
- `SELECT 0` → cache "default"
- `SELECT N` → cache "N" (where N is any number)

Caches must be created with numeric names to be accessible via SELECT.

### Session Management

Each TCP connection has a `Session` struct that tracks:
- Connection ID (for lock tagging)
- Selected cache name (default: "default")
- RESP reader/writer (tidwall/resp connection wrapper)
- Context for cancellation
- MULTI mode state (for transaction queuing)

### Testing

RESP commands are tested with the official `go-redis` client library:
- `test_redis_client.go`: Basic connectivity (PING, ECHO)
- `test_string_commands.go`: String operations
- `test_key_commands.go`: TTL and pattern matching
- `test_hash_commands.go`: Hash operations
- `test_list_commands.go`: List operations (queue/stack patterns)
- `test_select_caches.go`: Multi-cache isolation
- `test_integration.go`: Comprehensive end-to-end test

### Implemented Commands (45 total)

**String (15)**: GET, SET, DEL, EXISTS, INCR, DECR, INCRBY, DECRBY, MGET, MSET, GETSET, SETNX, SETEX, STRLEN, APPEND

**Key (6)**: EXPIRE, PEXPIRE, PERSIST, TTL, PTTL, KEYS

**Hash (10)**: HGET, HSET, HGETALL, HDEL, HEXISTS, HLEN, HKEYS, HVALS, HMGET, HMSET

**List (8)**: LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE, LINDEX, LSET

**Generic (6)**: PING, ECHO, SELECT, COMMAND, HELLO, CLIENT

See [REDIS.md](./REDIS.md) for complete documentation on Redis protocol support, including usage examples, limitations, and migration guide.

## External Dependencies

- **Echo**: Web framework (`github.com/labstack/echo/v4`)
- **Gabs**: JSON manipulation (`github.com/Jeffail/gabs/v2`)
- **govaluate**: Expression evaluation (`github.com/Knetic/govaluate`)
- **UUID**: Trigger IDs (`github.com/google/uuid`)
- **logos**: Structured logging (`github.com/goodblaster/logos`)
- **tidwall/resp**: RESP protocol parser/encoder (`github.com/tidwall/resp`)
