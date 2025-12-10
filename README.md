# map-cache

A powerful, in-memory caching service with an HTTP API built in Go. Designed for scenarios requiring simple map-based storage, atomic operations, conditional logic, and event-driven automation through triggers.

## 🚀 Features

- **Multiple Named Caches**: Create and manage multiple independent cache instances
- **Nested Key-Value Storage**: Store complex nested data structures with path-based access
- **Atomic Commands**: Execute batch operations with conditional logic and loops
- **Event-Driven Triggers**: Automatically react to data changes with pattern-based triggers
- **Key Expiration (TTL)**: Set time-to-live for individual keys or entire caches
- **RESTful API**: Full REST API with OpenAPI/Swagger documentation
- **Wildcard Patterns**: Use wildcards in keys for pattern matching and bulk operations
- **Value Interpolation**: Reference and compute values dynamically using `${{...}}` syntax
- **Backup & Restore**: Admin endpoints for cache backup and restoration

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

### 1. Check Server Status

```bash
curl http://localhost:8080/status
```

Response:
```json
{
  "status": "ok",
  "build": {
    "version": "1.0.0",
    "commit": "abc123",
    "date": "2024-01-01T00:00:00Z"
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
Full replacement of a key's value.

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

### Command Types

#### INC - Increment/Decrement
Increment or decrement a numeric value.

```json
{
  "type": "INC",
  "key": "domains/domain-1/countdown",
  "value": -1
}
```

#### REPLACE - Overwrite Value
Replace a key's value completely.

```json
{
  "type": "REPLACE",
  "key": "status",
  "value": "complete"
}
```

#### RETURN - Return Value
Return a value or computed expression. This is typically the last command in a sequence.

```json
{
  "type": "RETURN",
  "key": "${{status}}"
}
```

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

#### NOOP - No Operation
A no-op command that does nothing. Useful in conditional branches.

```json
{
  "type": "NOOP"
}
```

---

## 🔗 Value Interpolation

Use `${{...}}` syntax to reference values dynamically within commands and triggers.

### Basic Interpolation

- `${{key}}` → Gets the value at `key`
- `${{parent/child}}` → Gets nested values
- `${{some/*/value}}` → Wildcard pattern (returns array of matching values)

### Captured Values in FOR Loops

When using `FOR` with wildcards, captured segments are available:

```json
{
  "type": "FOR",
  "loop_expr": "${{users/*/profile/name}}",
  "commands": [
    {
      "type": "REPLACE",
      "key": "users/${{1}}/lastSeen",
      "value": "${{timestamp}}"
    }
  ]
}
```

In this example:
- `${{1}}` = the value captured by the first `*` (e.g., "user-123")
- The loop iterates over all matching `users/*/profile/name` paths

### String Interpolation

You can embed interpolated values in strings:

```json
{
  "type": "RETURN",
  "key": "Status is ${{status}} and count is ${{count}}"
}
```

### Expression Evaluation

Conditions support expressions:

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

### Example 1: Job Progress Tracking

Track progress of a distributed job across multiple domains.

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

**Step 4: Decrement countdowns**

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
          {
            "type": "INC",
            "key": "domains/${{1}}/countdown",
            "value": -1
          }
        ]
      },
      {
        "type": "RETURN",
        "key": "${{status}}"
      }
    ]
  }'
```

The triggers will automatically fire as countdowns reach zero, updating statuses accordingly.

### Example 2: User Session Management

Manage user sessions with automatic expiration.

```bash
# Create session cache
curl -X POST http://localhost:8080/api/v1/caches \
  -H "Content-Type: application/json" \
  -d '{"name": "sessions"}'

# Create session with 30-minute TTL
curl -X POST http://localhost:8080/api/v1/keys \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: sessions" \
  -d '{
    "entries": {
      "user-123": {
        "userId": 123,
        "email": "user@example.com",
        "lastActivity": "2024-01-01T12:00:00Z"
      }
    },
    "ttl": {
      "user-123": 1800000
    }
  }'

# Update last activity
curl -X PATCH http://localhost:8080/api/v1/keys/user-123 \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: sessions" \
  -d '{
    "commands": [
      {
        "type": "REPLACE",
        "path": "lastActivity",
        "value": "2024-01-01T12:05:00Z"
      }
    ]
  }'
```

### Example 3: Real-time Counter with Conditions

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

### Example 4: Monitoring with any() Function

Check if any service is down:

```bash
curl -X POST http://localhost:8080/api/v1/commands/execute \
  -H "Content-Type: application/json" \
  -H "X-Cache-Name: monitoring" \
  -d '{
    "commands": [
      {
        "type": "IF",
        "condition": "any(${{services/*/status}} == \"down\")",
        "if_true": {
          "type": "REPLACE",
          "key": "alert",
          "value": "Service outage detected"
        },
        "if_false": {
          "type": "REPLACE",
          "key": "alert",
          "value": "All services operational"
        }
      },
      {
        "type": "RETURN",
        "key": "${{alert}}"
      }
    ]
  }'
```

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

---

## 📦 Postman Collections

The repository includes Postman collections for testing:

- `internal/api/v1/postman/map-cache-api-full.postman_collection.json` - Complete API collection
- `internal/api/v1/postman/map-cache-scenario-countdown.postman_collection.json` - Countdown scenario walkthrough

Import these into Postman to explore the API interactively.

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

---

## 📄 License

MIT

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

For questions or support, contact [dave@goodblaster.com](mailto:dave@goodblaster.com)

---

## 🔗 Related Resources

- **API Documentation**: Available at `/api/v1/docs` when the server is running
- **OpenAPI Spec**: Available at `/api/v1/docs/openapi.yaml`
- **Health Check**: `GET /status`

