# map-cache

A powerful, in-memory caching service with HTTP and **Redis-compatible** APIs and a sophisticated query language, built in Go. Think **Redis meets a workflow engine** - combining high-performance caching with reactive programming, conditional logic, and pattern-based automation.

---

## 🎯 Why map-cache?

### The Problem It Solves

Traditional caching solutions like Redis and Memcached excel at simple key-value operations but fall short when you need:

- **Complex conditional logic** on cached data
- **Reactive workflows** that respond to data changes
- **Pattern-based batch operations** across multiple keys
- **Workflow orchestration** without external tools
- **Type-safe nested structures** with path-based access

map-cache fills this void by combining the speed of in-memory caching with a **declarative query language** that enables sophisticated data transformations, conditional workflows, and event-driven automation - all in a single, dependency-free binary.

### What Makes It Unique

| Feature | Redis | Memcached | map-cache |
|---------|-------|-----------|-----------|
| **In-memory performance** | ✅ | ✅ | ✅ |
| **Redis protocol (RESP)** | ✅ | ❌ | ✅ 71 commands |
| **Nested JSON structures** | ⚠️ Limited | ❌ | ✅ Native |
| **Conditional logic** | ⚠️ Lua scripts | ❌ | ✅ Built-in |
| **Pattern matching** | ⚠️ SCAN | ❌ | ✅ Wildcards |
| **Reactive triggers** | ❌ | ❌ | ✅ Event-driven |
| **Workflow orchestration** | ❌ | ❌ | ✅ Commands + Triggers |
| **Type preservation** | ⚠️ Strings only | ⚠️ Strings only | ✅ Full JSON types |
| **Zero dependencies** | ❌ | ❌ | ✅ Single binary |
| **RESTful API** | ❌ | ❌ | ✅ OpenAPI/Swagger |

### Real-World Use Cases

map-cache excels at scenarios requiring **stateful coordination** and **reactive behavior**:

1. **Job Orchestration** - Track distributed job progress with automatic completion detection
2. **Workflow Engines** - State machines with trigger-based transitions
3. **Real-time Dashboards** - Aggregate metrics with automatic threshold alerts
4. **Rate Limiting** - Complex rate limiting with conditional logic and automatic resets
5. **Session Management** - Rich session data with automatic cleanup and activity tracking
6. **Feature Flags** - Centralized flags with pattern-based bulk updates
7. **Game State** - Player stats with automatic achievement unlocking via triggers
8. **Configuration Management** - Hierarchical configs with environment-specific overrides
9. **Distributed Coordination** - Locking, leader election, task distribution
10. **Event-Driven Systems** - Triggers enable reactive programming without external message queues

### When to Choose map-cache Over Redis

**Choose map-cache** when you need:
- Complex conditional workflows on cached data
- Reactive behavior (triggers fire on changes)
- Deep nested structures with path-based access
- Pattern-based bulk operations
- Embedded workflow logic without external orchestrators

**Choose Redis** when you need:
- Persistence to disk
- Pub/sub messaging
- Distributed deployment (clustering)
- Mature ecosystem with extensive tooling
- Battle-tested production reliability at massive scale

### Performance Highlights

Recent optimizations deliver exceptional performance:

- **76% faster** conditional expressions (IF commands) with expression caching
- **78% less memory** for complex pattern matching operations
- **Sub-microsecond** simple key lookups (33 ns/op)
- **Zero allocation** cache fetches from global registry
- **21% faster** complex workflow scenarios
- Commands return values with **zero performance cost**

**RESP vs HTTP Protocol Performance:**
- **2-10x faster** operations via Redis protocol vs HTTP/JSON
- **18x faster** batch operations (MGET)
- **5-25x less memory** per operation

See [BENCHMARKS.md](BENCHMARKS.md) for detailed RESP vs HTTP comparison and the [benchmarks](#-testing) section for core operation metrics.

---

## 🚀 Features

- **Redis Protocol Support**: Drop-in replacement for Redis clients - 71 commands including GET, SET, HSET, LPUSH, EXPIRE, and more
- **Multiple Named Caches**: Create and manage multiple independent cache instances
- **Nested Key-Value Storage**: Store complex nested data structures with path-based access
- **Atomic Commands**: Execute batch operations with conditional logic, loops, and value interpolation
- **Event-Driven Triggers**: Automatically react to data changes with pattern-based triggers
- **Key Expiration (TTL)**: Set time-to-live for individual keys or entire caches
- **Dual APIs**: Full REST API with OpenAPI/Swagger + Redis-compatible RESP protocol on port 6379
- **Wildcard Patterns**: Use wildcards in keys for pattern matching and bulk operations
- **Value Interpolation**: Reference and compute values dynamically using `${{...}}` syntax
- **Optional Values**: Graceful fallbacks with `${{key || default}}` syntax
- **Expression Caching**: Automatic caching of compiled expressions for 76% faster conditionals
- **Backup & Restore**: Admin endpoints for cache backup and restoration
- **Zero Dependencies**: Single binary with no external requirements

---

## 📦 Installation

### Using Docker (Recommended)

```bash
# Pull and run the latest image
docker run -d -p 8080:8080 goodblaster/map-cache:latest

# Or use docker-compose
docker-compose up -d
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/goodblaster/map-cache.git
cd map-cache

# Build the binary
go build -o map-cache ./cmd/cache/main.go

# Run the server
./map-cache
```

The server will start on `http://localhost:8080` by default. You can configure the port using the `LISTEN_ADDRESS` environment variable:

```bash
LISTEN_ADDRESS=":3000" ./map-cache
```

### View API Documentation

Once the server is running, visit:
- **Swagger UI**: `http://localhost:8080/api/v1/docs`
- **OpenAPI Spec**: `http://localhost:8080/api/v1/docs/openapi.yaml`

---

## 🎯 Quick Start

### 1. Check Server Health

```bash
curl http://localhost:8080/healthz
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-10T12:34:56Z",
  "uptime_seconds": 3600,
  "build": {
    "version": "1.0.0",
    "commit": "abc123",
    "date": "2024-01-01T00:00:00Z"
  },
  "system": {
    "goroutines": 42,
    "memory_alloc_mb": 12,
    "memory_sys_mb": 24,
    "gc_count": 5
  },
  "caches": {
    "count": 1,
    "names": ["default"]
  }
}
```

### 2. Create a Named Cache

```bash
curl -X POST http://localhost:8080/api/v1/caches \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-cache"
  }'
```

### 3. Add Data to the Cache

```bash
curl -X POST http://localhost:8080/api/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: my-cache" \
  -d '{
    "entries": {
      "user": {
        "id": 12345,
        "name": "Alice",
        "email": "alice@example.com",
        "preferences": {
          "theme": "dark",
          "notifications": true
        }
      },
      "counter": 0,
      "tags": ["go", "cache", "api"]
    }
  }'
```

### 4. Retrieve Data

```bash
# Get a single value
curl http://localhost:8080/api/v1/keys/user/name \
  -H "X-Cache-Name: my-cache"

# Response: "Alice"

# Get nested value
curl http://localhost:8080/api/v1/keys/user/preferences/theme \
  -H "X-Cache-Name: my-cache"

# Response: "dark"
```

### 5. Update Data

```bash
# Full replace (PUT)
curl -X PUT http://localhost:8080/api/v1/keys/counter \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: my-cache" \
  -d '42'

# Partial update (PATCH)
curl -X PATCH http://localhost:8080/api/v1/keys/user \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: my-cache" \
  -d '{
    "commands": [
      {"type": "REPLACE", "path": "email", "value": "alice.new@example.com"}
    ]
  }'
```

---

## 🔌 Redis Protocol Support

Map-cache implements the Redis Serialization Protocol (RESP2), allowing you to use standard Redis clients alongside the HTTP API. Both protocols share the same underlying cache storage.

### Quick Start with Redis

```bash
# Enable RESP server (default port 6379)
RESP_ENABLED=true ./map-cache

# Use redis-cli
redis-cli

# Basic operations
127.0.0.1:6379> SET mykey "Hello World"
OK
127.0.0.1:6379> GET mykey
"Hello World"
127.0.0.1:6379> INCR counter
(integer) 1
127.0.0.1:6379> EXPIRE mykey 60
(integer) 1
127.0.0.1:6379> TTL mykey
(integer) 60
```

### Using Redis Client Libraries

**Go (go-redis):**
```go
import "github.com/redis/go-redis/v9"

client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
client.Set(ctx, "user:100:name", "Alice", 0)
val, _ := client.Get(ctx, "user:100:name").Result()
```

**Python (redis-py):**
```python
import redis

r = redis.Redis(host='localhost', port=6379, decode_responses=True)
r.hset('user:100', mapping={'name': 'Alice', 'email': 'alice@example.com'})
print(r.hgetall('user:100'))
```

### Supported Commands (45 total)

**String (15)**: GET, SET, DEL, EXISTS, INCR, DECR, INCRBY, DECRBY, MGET, MSET, GETSET, SETNX, SETEX, STRLEN, APPEND

**Hash (10)**: HGET, HSET, HGETALL, HDEL, HEXISTS, HLEN, HKEYS, HVALS, HMGET, HMSET

**List (8)**: LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE, LINDEX, LSET

**Key (6)**: EXPIRE, PEXPIRE, PERSIST, TTL, PTTL, KEYS

**Generic (6)**: PING, ECHO, SELECT, COMMAND, HELLO, CLIENT

### Key Features

- **Automatic key translation**: Redis keys (`user:123:name`) automatically map to paths (`user/123/name`)
- **Multi-cache support**: Use `SELECT` command to switch between numbered caches
- **Hash commands**: Map naturally to nested JSON structures
- **Both APIs access same data**: Set via Redis, retrieve via HTTP (and vice versa)

For complete Redis protocol documentation, limitations, and migration guide, see **[REDIS.md](./REDIS.md)**.

---

## 📚 API Reference

### Cache Management

#### List All Caches
```http
GET /api/v1/caches
```

#### Create a Cache
```http
POST /api/v1/caches
Content-Type: application/json

{
  "name": "cache-name",
  "ttl": 3600000  // Optional: TTL in milliseconds
}
```

#### Delete a Cache
```http
DELETE /api/v1/caches/:name
```

---

### Key Operations

All key operations require the `X-Cache-Name` header to specify which cache to use. If omitted, the default cache is used.

#### Create Keys (POST)
Creates new keys. Returns 409 Conflict if keys already exist.

```http
POST /api/v1/keys
X-Cache-Name: my-cache
Content-Type: application/json

{
  "entries": {
    "key1": "value1",
    "key2": 42,
    "nested": {
      "deep": {
        "value": "hello"
      }
    }
  },
  "ttl": {  // Optional: per-key TTL in milliseconds
    "key1": 5000,
    "key2": 10000
  }
}
```

#### Get Single Key (GET)
```http
GET /api/v1/keys/:key
X-Cache-Name: my-cache
```

**Path-based access**: Use `/` to access nested values:
- `GET /api/v1/keys/user` → entire user object
- `GET /api/v1/keys/user/name` → just the name
- `GET /api/v1/keys/user/preferences/theme` → nested value

#### Get Multiple Keys (POST)
```http
POST /api/v1/keys/get
X-Cache-Name: my-cache
Content-Type: application/json

{
  "keys": ["key1", "key2", "nested/deep/value"]
}
```

Response:
```json
["value1", 42, "hello"]
```

#### Replace Key (PUT)
Full replacement of a key's value. **Now returns the new value.**

```http
PUT /api/v1/keys/:key
X-Cache-Name: my-cache
Content-Type: application/json

"new value"
```

#### Replace Multiple Keys (PUT)
```http
PUT /api/v1/keys
X-Cache-Name: my-cache
Content-Type: application/json

{
  "entries": {
    "key1": "updated1",
    "key2": "updated2"
  }
}
```

#### Partial Update (PATCH)
Update specific paths within a key.

```http
PATCH /api/v1/keys/:key
X-Cache-Name: my-cache
Content-Type: application/json

{
  "commands": [
    {"type": "REPLACE", "path": "email", "value": "new@example.com"},
    {"type": "DELETE", "path": "oldField"}
  ]
}
```

#### Delete Key (DELETE)
```http
DELETE /api/v1/keys/:key
X-Cache-Name: my-cache
```

#### Delete Multiple Keys (POST)
```http
POST /api/v1/keys/delete
X-Cache-Name: my-cache
Content-Type: application/json

{
  "keys": ["key1", "key2", "key3"]
}
```

---

## ⚡ Commands

Commands enable atomic batch operations with conditional logic, loops, and value interpolation. They execute in a single transaction, ensuring consistency.

### Execute Commands

```http
POST /api/v1/commands/execute
X-Cache-Name: my-cache
Content-Type: application/json

{
  "commands": [
    {
      "type": "INC",
      "key": "counter",
      "value": 1
    },
    {
      "type": "RETURN",
      "key": "${{counter}}"
    }
  ]
}
```

Response (all commands now return values):
```json
[5, 5]
```

### Command Types

#### INC - Increment/Decrement
Increment or decrement a numeric value. **Now returns the new value.**

```json
{
  "type": "INC",
  "key": "domains/domain-1/countdown",
  "value": -1
}
```

**Returns**: The new value after increment (e.g., `4`)

#### REPLACE - Overwrite Value
Replace a key's value completely. **Now returns the new value.**

```json
{
  "type": "REPLACE",
  "key": "status",
  "value": "complete"
}
```

**Returns**: The new value (e.g., `"complete"`)

#### DELETE - Remove Key 🆕
Delete a key from the cache. Supports wildcards for bulk deletion. **Returns the deleted value(s).**

```json
{
  "type": "DELETE",
  "key": "users/123/temp"
}
```

**Wildcard deletion**:
```json
{
  "type": "DELETE",
  "key": "sessions/*/expired"
}
```

**Returns**:
- Single key: the deleted value (or `null` if not found)
- Wildcard: array of deleted values

#### GET - Retrieve Value
Fetch a value from the cache. Supports wildcards.

```json
{
  "type": "GET",
  "key": "users/*/name"
}
```

**Returns**:
- Single key: the value
- Wildcard: map of `key → value` pairs

#### RETURN - Return Value
Return a value or computed expression. This is typically the last command in a sequence.

```json
{
  "type": "RETURN",
  "key": "${{status}}"
}
```

**String interpolation**:
```json
{
  "type": "RETURN",
  "key": "Status is ${{status}}, count is ${{counter}}"
}
```

**Optional values with fallback** 🆕:
```json
{
  "type": "RETURN",
  "key": "${{user/name || Guest}}"
}
```

Syntax: `${{primary || fallback || default}}`

**Returns**: The computed value with proper type preservation

#### IF - Conditional Execution
Execute one of two commands based on a condition.

```json
{
  "type": "IF",
  "condition": "${{countdown}} <= 0",
  "if_true": {
    "type": "REPLACE",
    "key": "status",
    "value": "complete"
  },
  "if_false": {
    "type": "NOOP"
  }
}
```

**Supported operators**: `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`

**Aggregation functions**:
- `all(${{pattern}} == value)` - Returns true if all matching values satisfy the condition
- `any(${{pattern}} == value)` - Returns true if any matching value satisfies the condition

**Performance**: Expressions are automatically cached, making repeated IF conditions **76% faster**.

#### FOR - Loop Over Pattern
Iterate over keys matching a wildcard pattern.

```json
{
  "type": "FOR",
  "loop_expr": "${{domains/*/countdown}}",
  "commands": [
    {
      "type": "INC",
      "key": "domains/${{1}}/countdown",
      "value": -1
    }
  ]
}
```

The `loop_expr` uses wildcards (`*`) to match multiple keys. Captured values are available as `${{1}}`, `${{2}}`, etc.

**All command types now work in FOR loops**, including DELETE, PRINT, RETURN, and nested FOR/COMMANDS.

#### COMMANDS - Group Commands
Execute multiple commands sequentially. Returns an array of all results.

```json
{
  "type": "COMMANDS",
  "commands": [
    {"type": "INC", "key": "counter", "value": 1},
    {"type": "REPLACE", "key": "status", "value": "updated"},
    {"type": "RETURN", "key": "${{counter}}"}
  ]
}
```

**Returns**: Array of results from each command: `[5, "updated", 5]`

#### NOOP - No Operation
A no-op command that does nothing. Useful in conditional branches.

```json
{
  "type": "NOOP"
}
```

**Returns**: `null`

---

## 🔗 Value Interpolation

Use `${{...}}` syntax to reference values dynamically within commands and triggers.

### Basic Interpolation

- `${{key}}` → Gets the value at `key`
- `${{parent/child}}` → Gets nested values
- `${{some/*/value}}` → Wildcard pattern (returns array of matching values)

### Optional Values with Fallback 🆕

Gracefully handle missing keys with fallback syntax:

```json
{
  "type": "RETURN",
  "key": "${{config/timeout || 30}}"
}
```

**Fallback chain**:
```json
{
  "type": "RETURN",
  "key": "${{primary || secondary || default}}"
}
```

**Features**:
- Tries each key in order until one exists
- Last value is treated as literal default
- Supports numbers, booleans, strings, null
- Works in string templates: `"Hello, ${{name || Guest}}!"`
- Type-preserving: `${{count || 0}}` returns integer 0, not string "0"

**Not allowed**: `${{users/*/name || unknown}}` (wildcards with fallback)

### Captured Values in FOR Loops

When using `FOR` with wildcards, captured segments are available:

```json
{
  "type": "FOR",
  "loop_expr": "${{users/*/profile/name}}",
  "commands": [
    {
      "type": "DELETE",
      "key": "users/${{1}}/cache"
    },
    {
      "type": "PRINT",
      "messages": ["Cleared cache for user ${{1}}"]
    }
  ]
}
```

In this example:
- `${{1}}` = the value captured by the first `*` (e.g., "user-123")
- The loop iterates over all matching `users/*/profile/name` paths
- All command types (DELETE, PRINT, RETURN, etc.) now receive captures

### String Interpolation

You can embed interpolated values in strings:

```json
{
  "type": "RETURN",
  "key": "Status is ${{status}} and count is ${{count}}"
}
```

### Expression Evaluation

Conditions support expressions with automatic caching for performance:

```json
{
  "type": "IF",
  "condition": "${{count}} > 10 && ${{status}} == \"active\"",
  "if_true": { "type": "REPLACE", "key": "ready", "value": true },
  "if_false": { "type": "NOOP" }
}
```

---

## 🎯 Triggers

Triggers automatically execute commands when keys matching a pattern are updated. They enable event-driven workflows and reactive programming.

### Create a Trigger

```http
POST /api/v1/triggers
X-Cache-Name: my-cache
Content-Type: application/json

{
  "key": "domains/*/countdown",
  "command": {
    "type": "IF",
    "condition": "${{domains/${{1}}/countdown}} <= 0",
    "if_true": {
      "type": "REPLACE",
      "key": "domains/${{1}}/status",
      "value": "complete"
    },
    "if_false": {
      "type": "NOOP"
    }
  }
}
```

**Response**: Returns a trigger ID (UUID)

### Delete a Trigger

```http
DELETE /api/v1/triggers/:id
X-Cache-Name: my-cache
```

### Replace a Trigger

```http
PUT /api/v1/triggers/:id
X-Cache-Name: my-cache
Content-Type: application/json

{
  "key": "new-pattern/*/key",
  "command": { ... }
}
```

### Trigger Behavior

- Triggers fire **after** the key update completes
- Multiple triggers can match the same key pattern
- Triggers execute in the order they were created
- Trigger commands can modify other keys, which may fire additional triggers (cascading)

**⚠️ Infinite Loop Protection:**
- Trigger recursion is automatically limited to 10 levels deep
- If a trigger chain exceeds this depth (e.g., trigger A fires trigger B which fires A again), an error is returned
- This prevents server crashes from runaway trigger loops
- Design triggers carefully to avoid circular dependencies

---

## ⏰ Expiration (TTL)

Set time-to-live for keys or entire caches. TTL values are specified in **milliseconds**.

### Per-Key TTL

Set TTL when creating keys:

```json
{
  "entries": {
    "session": "abc123",
    "token": "xyz789"
  },
  "ttl": {
    "session": 3600000,  // 1 hour
    "token": 1800000     // 30 minutes
  }
}
```

### Cache-Level TTL

Set TTL when creating a cache:

```json
{
  "name": "temp-cache",
  "ttl": 7200000  // 2 hours - entire cache expires
}
```

---

## 💡 Use Cases & Examples

### Example 1: Job Progress Tracking with Cascading Triggers

Track progress of a distributed job across multiple domains with automatic completion detection.

**Step 1: Create cache and initialize domains**

```bash
curl -X POST http://localhost:8080/api/v1/caches \
  -H "Content-Type: application/json" \
  -d '{"name": "job-1234"}'

curl -X POST http://localhost:8080/api/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: job-1234" \
  -d '{
    "entries": {
      "domains": {
        "domain-1": {"status": "busy", "countdown": 5},
        "domain-2": {"status": "busy", "countdown": 3},
        "domain-3": {"status": "busy", "countdown": 7}
      },
      "status": "running"
    }
  }'
```

**Step 2: Create trigger to mark domain complete when countdown reaches zero**

```bash
curl -X POST http://localhost:8080/api/v1/triggers \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: job-1234" \
  -d '{
    "key": "domains/*/countdown",
    "command": {
      "type": "IF",
      "condition": "${{domains/${{1}}/countdown}} <= 0",
      "if_true": {
        "type": "REPLACE",
        "key": "domains/${{1}}/status",
        "value": "complete"
      },
      "if_false": {"type": "NOOP"}
    }
  }'
```

**Step 3: Create trigger to mark job complete when all domains are done**

```bash
curl -X POST http://localhost:8080/api/v1/triggers \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: job-1234" \
  -d '{
    "key": "domains/*/status",
    "command": {
      "type": "IF",
      "condition": "all(${{domains/*/status}} == \"complete\")",
      "if_true": {
        "type": "REPLACE",
        "key": "status",
        "value": "complete"
      },
      "if_false": {"type": "NOOP"}
    }
  }'
```

**Step 4: Decrement countdowns and see automatic completion**

```bash
curl -X POST http://localhost:8080/api/v1/commands/execute \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: job-1234" \
  -d '{
    "commands": [
      {
        "type": "FOR",
        "loop_expr": "${{domains/*/countdown}}",
        "commands": [
          {"type": "INC", "key": "domains/${{1}}/countdown", "value": -1}
        ]
      },
      {
        "type": "RETURN",
        "key": "${{status}}"
      }
    ]
  }'
```

The triggers will automatically fire as countdowns reach zero, cascading to mark the entire job complete.

### Example 2: Optional Values for Configuration

```bash
curl -X POST http://localhost:8080/api/v1/commands/execute \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: config" \
  -d '{
    "commands": [
      {
        "type": "RETURN",
        "key": "Timeout: ${{config/timeout || 30}}s, Retries: ${{config/retries || 3}}"
      }
    ]
  }'
```

Returns graceful defaults even if config keys don't exist: `"Timeout: 30s, Retries: 3"`

### Example 3: Bulk Cleanup with DELETE

```bash
curl -X POST http://localhost:8080/api/v1/commands/execute \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: my-cache" \
  -d '{
    "commands": [
      {
        "type": "DELETE",
        "key": "sessions/*/expired"
      },
      {
        "type": "RETURN",
        "key": "Cleaned up ${{deleted_count}} expired sessions"
      }
    ]
  }'
```

### Example 4: Real-time Counter with Improved Return Values

```bash
curl -X POST http://localhost:8080/api/v1/commands/execute \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: my-cache" \
  -d '{
    "commands": [
      {
        "type": "INC",
        "key": "visitorCount",
        "value": 1
      },
      {
        "type": "IF",
        "condition": "${{visitorCount}} >= 1000",
        "if_true": {
          "type": "REPLACE",
          "key": "milestone",
          "value": "reached"
        },
        "if_false": {"type": "NOOP"}
      },
      {
        "type": "RETURN",
        "key": "Current visitors: ${{visitorCount}}"
      }
    ]
  }'
```

Now returns: `[1001, "reached", "Current visitors: 1001"]` - all commands return their values!

---

## 🎓 Real-World Use Case Scenarios

The `pkg/caches/scenarios_test.go` file contains comprehensive, production-ready examples demonstrating what map-cache excels at. These scenarios serve as both documentation and executable tests.

### View the Scenarios

```bash
# Run all scenario tests
go test -v -run TestScenario ./pkg/caches

# Run a specific scenario
go test -v -run TestScenario_SessionManagement ./pkg/caches
```

### Featured Scenarios

#### 1. **Session Management** (`TestScenario_SessionManagement`)
- Store user sessions with nested data (user info, permissions, timestamps)
- Automatic session cleanup with TTL-based expiration
- Perfect for web applications requiring session storage

#### 2. **Feature Flags** (`TestScenario_FeatureFlags`)
- Centralized feature flag management across multiple services
- Bulk enable/disable features using wildcard patterns
- Check feature availability with `any()` and `all()` functions
- Ideal for gradual feature rollouts and A/B testing

#### 3. **Rate Limiting** (`TestScenario_RateLimiting`)
- Per-user API rate limiting with automatic window resets
- Track request counts and enforce limits
- Auto-reset counters using TTL expiration
- Production-ready for API throttling

#### 4. **Shopping Cart** (`TestScenario_ShoppingCart`)
- Product catalog integration with ArrayAppend for adding items
- Triggers auto-increment cart total when products are marked as added
- Demonstrates product lookup and trigger-based price aggregation
- Shows pattern for maintaining cart state with product references

#### 5. **Leaderboards** (`TestScenario_Leaderboard`)
- Real-time gaming leaderboards with score tracking
- Trigger-based automatic "elite" status when players cross 1500 points
- Demonstrates conditional logic with triggers
- Shows how triggers can update related fields when scores change

#### 6. **User Presence Tracking** (`TestScenario_PresenceTracking`)
- Track online/offline user status
- Auto-remove inactive users with TTL
- Heartbeat-based activity updates
- Great for chat applications and collaborative tools

#### 7. **Configuration Management** (`TestScenario_ConfigurationManagement`)
- Environment-specific configuration with fallbacks
- Bulk configuration updates across environments
- Hierarchical config with default values
- Enterprise configuration management

#### 8. **Workflow State Machine** (`TestScenario_WorkflowStateMachine`)
- Order processing workflows with state transitions
- Trigger-based automation for state changes
- Automatic timestamp tracking
- Business process automation

#### 9. **Metrics Aggregation** (`TestScenario_MetricsAggregation`)
- Real-time metrics collection across services
- Counter increments for requests, errors, jobs
- Threshold-based alerting with conditional logic
- Observability and monitoring

#### 10. **Distributed Locking** (`TestScenario_DistributedLock`)
- Distributed lock implementation with TTL
- Automatic lock release on expiration
- Process coordination across instances
- Prevent race conditions in distributed systems

#### 11. **Parallel Batch Processing** (`TestScenario_ParallelBatchProcessing`)
- Distributed batch job processing with cascading triggers
- Multiple tasks with countdown-based completion tracking
- Two-level trigger cascade: task completion → batch completion
- Perfect for ETL pipelines, MapReduce, parallel test execution, or render farms
- Demonstrates `all()` aggregation function for conditional logic

### Why These Scenarios Matter

Each scenario demonstrates:
- **Real production patterns** - Not toy examples, but actual use cases
- **Best practices** - Proper error handling, TTL usage, and atomic operations
- **Feature showcase** - Highlights specific map-cache capabilities
- **Copy-paste ready** - Use as templates for your own implementation

These tests run in CI/CD, ensuring the examples stay accurate and functional as the project evolves.

---

## 🛠️ Admin Endpoints

### Backup Cache

```http
POST /admin/backup
Content-Type: application/json

{
  "cache_name": "my-cache"
}
```

Returns a JSON representation of the entire cache.

### Restore Cache

```http
POST /admin/restore
Content-Type: application/json

{
  "cache_name": "my-cache",
  "data": { ... }
}
```

---

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run Test_Big ./pkg/caches

# Run stress tests
go test -v -run TestStress ./pkg/caches
```

The `Test_Big` test simulates a countdown scenario with 100+ domains, demonstrating cascading completion logic.

### Benchmarks

Performance benchmarks are available to measure the speed of cache operations:

```bash
# Run all benchmarks (skip regular tests)
go test -bench=. -benchmem -run=^$ ./...

# Run cache operation benchmarks
go test -bench=. -benchmem -run=^$ ./pkg/caches

# Run specific benchmark
go test -bench=BenchmarkCache_Get -benchmem -run=^$ ./pkg/caches

# Run benchmarks with shorter duration (faster)
go test -bench=. -benchmem -run=^$ -benchtime=500ms ./pkg/caches
```

**Note**: The `-run=^$` flag skips regular tests and only runs benchmarks, avoiding test log output.

### Performance Highlights

Recent optimizations deliver exceptional results:

| Operation | Performance | Improvement |
|-----------|-------------|-------------|
| **Simple Get** | 33 ns/op | Baseline |
| **Nested Get** | 83 ns/op | Baseline |
| **Replace** | 112 ns/op | Zero cost for return value |
| **Increment** | 142 ns/op | Slightly faster despite return value |
| **IF (simple)** | 5,882 ns/op | **19% faster** with caching |
| **IF (any/all)** | 58,152 ns/op | **76% faster** with caching |
| **Complex Workflow** | 58,523 ns/op | **22% faster overall** |

**Memory improvements**:
- IF expressions: **6-78% less memory** with caching
- Complex scenarios: **14% less memory**, **26% fewer allocations**

Available benchmarks include:
- **Cache Operations**: Create, Get, Replace, Delete, Increment, nested operations, batch operations
- **Command Execution**: INC, REPLACE, IF, FOR, value interpolation, complex scenarios
- **Wildcard Patterns**: Pattern matching with single/multiple wildcards
- **Container Operations**: Array operations, data retrieval, wildcard key matching
- **Concurrent Access**: Multi-threaded Get, Replace, and mixed operations

---

## 📦 Postman Collections

The repository includes comprehensive Postman collections:

- **Comprehensive Examples** (`map-cache-examples.postman_collection.json`) - 50+ examples including:
  - All command types (INC, REPLACE, DELETE, IF, FOR, RETURN)
  - Redis workflow replacements (rate limiting, sessions, shopping carts, leaderboards)
  - Advanced patterns (triggers, cascading workflows, multi-tenant scenarios)
  - New features (optional values, expression caching, wildcard operations)
- **API Reference** (`map-cache-api-full.postman_collection.json`) - Complete endpoint documentation
- **Countdown Tutorial** (`map-cache-scenario-countdown.postman_collection.json`) - Step-by-step reactive pattern example

See `internal/api/v1/postman/README.md` for detailed usage instructions and pattern examples.

---

## 🔧 Configuration

Configure the service using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_ADDRESS` | `:8080` | Address and port to listen on |
| `KEY_DELIMITER` | `/` | Delimiter for nested key paths |
| `LOG_FORMAT` | `json` | Log format (json/text) |

---

## 🏗️ Architecture

- **In-Memory Storage**: Fast, map-based storage with thread-safe operations
- **Atomic Operations**: Commands execute in transactions for consistency
- **Pattern Matching**: Wildcard support for flexible key matching
- **Event System**: Triggers enable reactive programming patterns
- **Expression Caching**: Compiled expressions cached for 76% faster conditionals
- **Type Preservation**: JSON values maintain their native types through all operations
- **Zero Allocations**: Optimized hot paths with zero-allocation cache lookups

---

## 📄 License

MIT

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

For questions or support, contact [dave@goodblaster.com](mailto:dave@goodblaster.com)

---

## 📊 Observability & Monitoring

### Health Check

The `/healthz` endpoint provides detailed server health information:

```bash
curl http://localhost:8080/healthz
```

Returns runtime metrics including uptime, memory usage, goroutine count, and cache statistics.

### Prometheus Metrics

Map-cache exposes Prometheus-compatible metrics at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

**HTTP Metrics:**
- `http_requests_total{method, path, status}` - Total HTTP requests by method, path, and status code
- `http_request_duration_seconds{method, path}` - HTTP request latency histogram
- `http_request_size_bytes{method, path}` - HTTP request size histogram
- `http_response_size_bytes{method, path}` - HTTP response size histogram
- `http_requests_in_flight` - Current number of HTTP requests being processed

**Cache Metrics** (updated every 10 seconds):
- `cache_size_bytes{cache}` - Current size of cache in bytes
- `cache_keys_total{cache}` - Total number of keys in cache
- `cache_activity_total{cache}` - Total operations performed on cache
- `cache_long_operations_total{cache}` - Total long-running operations
- `cache_timeout_operations_total{cache}` - Total timed-out operations
- `caches_total` - Total number of active caches

**Example Prometheus Queries:**
```promql
# P95 latency for API endpoints
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Request rate by endpoint
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m])

# Cache memory usage
sum(cache_size_bytes) by (cache)
```

**Grafana Dashboard:**
Import metrics into Grafana to visualize:
- Request rates and latencies
- Error rates and status code distribution
- Cache sizes and key counts
- Long-running operation trends

### Performance Profiling (pprof)

Map-cache exposes Go's standard pprof endpoints for runtime profiling and debugging:

**Available Endpoints:**
- `/debug/pprof/heap` - Heap memory allocation profile
- `/debug/pprof/goroutine` - Stack traces of all current goroutines
- `/debug/pprof/threadcreate` - Stack traces that led to thread creation
- `/debug/pprof/block` - Stack traces that led to blocking on synchronization primitives
- `/debug/pprof/mutex` - Stack traces of mutex contention
- `/debug/pprof/allocs` - All past memory allocations (sampling)
- `/debug/pprof/profile` - CPU profile (30-second duration by default)
- `/debug/pprof/trace` - Execution trace (duration specified via `?seconds=N`)
- `/debug/pprof/cmdline` - Command line arguments
- `/debug/pprof/symbol` - Symbol lookup

**Common Usage:**

```bash
# CPU profiling (30 seconds)
curl http://localhost:8080/debug/pprof/profile > cpu.prof
go tool pprof cpu.prof

# Heap memory snapshot
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Interactive CPU profiling with web visualization
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile

# Goroutine analysis (text format)
curl "http://localhost:8080/debug/pprof/goroutine?debug=1"

# Trace analysis
curl "http://localhost:8080/debug/pprof/trace?seconds=5" > trace.out
go tool trace trace.out

# Block profiling (enable first)
curl "http://localhost:8080/debug/pprof/block?debug=1"

# Mutex contention profiling
curl "http://localhost:8080/debug/pprof/mutex?debug=1"
```

**Production Profiling Tips:**
- CPU profiling has minimal overhead (~5%)
- Use `?seconds=N` to control trace/profile duration
- Heap profiles are snapshots and safe to collect anytime
- Add `?debug=1` for human-readable text output
- Use `go tool pprof -http` for interactive web-based analysis
- Profiles can be compared to detect regressions: `go tool pprof -base old.prof new.prof`

---

## 🔗 Related Resources

- **API Documentation**: Available at `/api/v1/docs` when the server is running
- **OpenAPI Spec**: Available at `/api/v1/docs/openapi.yaml`
- **Health Check**: `GET /healthz`
- **Prometheus Metrics**: `GET /metrics`
- **Performance Profiling**: `GET /debug/pprof/*` (heap, goroutine, cpu, etc.)
- **GitHub**: https://github.com/goodblaster/map-cache
