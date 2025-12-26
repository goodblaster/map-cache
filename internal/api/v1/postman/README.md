# map-cache Postman Collections

This directory contains Postman collections for testing and exploring the map-cache API.

## Collections

### 1. map-cache - Comprehensive Examples
**File:** `map-cache-examples.postman_collection.json`

**Purpose:** Learn map-cache capabilities through practical examples

**Contents:**
- **Getting Started** - Basic operations for newcomers
- **Command System** - All command types (INC, REPLACE, DELETE, IF, FOR, RETURN, etc.)
- **New Features** - Optional values, expression caching, DELETE with wildcards
- **Redis Replacements** - Common Redis patterns:
  - Rate limiting
  - Session management
  - Shopping carts
  - Leaderboards
  - Counter patterns
- **Advanced Patterns** - Triggers, workflows, cascading operations
- **Multi-tenant Scenarios** - Managing data for multiple tenants
- **Wildcard Operations** - Pattern matching and bulk operations
- **Performance Examples** - Batch operations, atomic transactions

**Best for:**
- Learning map-cache features
- Understanding Redis workflow replacements
- Exploring advanced patterns
- Performance optimization examples

### 2. map-cache API - Full Reference
**File:** `map-cache-api-full.postman_collection.json`

**Purpose:** Complete API reference documentation

**Contents:**
- Keys - CRUD operations on cache keys
- Caches - Cache lifecycle management
- Triggers - Reactive programming with triggers
- Commands - Command execution endpoint
- Admin - Backup and restore operations

**Best for:**
- API reference
- Testing all endpoints
- Integration testing
- Quick API exploration

### 3. Scenario - Countdown with Trigger to Complete
**File:** `map-cache-scenario-countdown.postman_collection.json`

**Purpose:** Step-by-step walkthrough of a reactive countdown scenario

**Demonstrates:**
- Creating a cache
- Setting up initial state
- Creating cascading triggers
- Decrementing counters
- Auto-completion via triggers

**Best for:**
- Understanding trigger mechanics
- Learning reactive patterns
- Step-by-step tutorials

## Environment

**File:** `map-cache-api-environment.postman_environment.json`

Contains environment variables:
- `host`: Server address (default: `localhost:8080`)

## Getting Started

### 1. Import Collections

1. Open Postman
2. Click "Import"
3. Select all JSON files from this directory
4. Collections will appear in your workspace

### 2. Set Environment

1. Select "Map Cache API Environment" from the environment dropdown
2. Verify `host` is set to your server address (default: `localhost:8080`)

### 3. Start the Server

```bash
# Build and run map-cache server
go build -o map-cache ./cmd/cache/main.go
./map-cache
```

Server will start on `http://localhost:8080` by default.

### 4. Run Examples

#### For Learning:
Start with **"map-cache - Comprehensive Examples"**
1. Open "1. Getting Started"
2. Run "Create a Cache"
3. Follow the examples in order

#### For API Testing:
Use **"map-cache API - Full Reference"**
1. Create a cache first
2. Test individual endpoints
3. Use `{{name}}` and `{{key}}` variables for dynamic values

#### For Tutorials:
Try **"Scenario - Countdown with Trigger to Complete"**
1. Run steps in order (1-7)
2. Observe how triggers cascade
3. See reactive programming in action

## Common Patterns

### Creating a Cache
All operations require a cache. Always start with:
```
POST /api/v1/caches
{
  "name": "my-cache"
}
```

### Using Cache Headers
Most endpoints require the cache name in headers:
```
X-Cache-Name: my-cache
```

### Value Interpolation
Use `${{key/path}}` to reference values:
```json
{
  "type": "RETURN",
  "key": "User: ${{username}}, Score: ${{score}}"
}
```

### Optional Values with Fallbacks
Chain fallbacks with `||`:
```json
{
  "type": "RETURN",
  "key": "${{user/premium || user/trial || false}}"
}
```

### Wildcard Patterns
Use `*` for pattern matching:
```json
{
  "type": "FOR",
  "loop_expr": "${{users/*/score}}",
  "command": {
    "type": "INC",
    "key": "users/${{1}}/score",
    "value": 10
  }
}
```

## Tips

### Performance
- Use batch operations with `COMMANDS` for multiple updates
- Expression caching provides 19-76% performance gain for repeated conditionals
- Atomic operations prevent race conditions

### Debugging
- Use `PRINT` commands to log values during execution
- Check trigger recursion depth (max: 10 levels)
- Use `RETURN` to inspect state after operations

### Best Practices
- Always clean up test caches
- Use descriptive cache names for multi-tenant scenarios
- Leverage triggers for reactive patterns instead of polling
- Use wildcards for bulk operations

## Redis Comparison

map-cache provides similar functionality to Redis but with different approaches:

| Redis Command | map-cache Equivalent |
|--------------|---------------------|
| `SET key value` | `POST /api/v1/keys` with entries |
| `GET key` | `GET /api/v1/keys/{key}` |
| `INCR key` | `INC` command |
| `DEL key` | `DELETE` command or `DELETE /api/v1/keys/{key}` |
| `HSET hash field value` | Create nested structure |
| `HGETALL hash` | GET with path to nested object |
| `ZADD` (sorted sets) | Store with score, query with patterns |
| `EXPIRE key seconds` | Cache TTL or key-level TTL |
| `WATCH/MULTI/EXEC` | `COMMANDS` for atomic batch operations |

## Additional Resources

- **API Documentation:** http://localhost:8080/api/v1/docs (when server is running)
- **Repository:** Check CLAUDE.md and README.md for architecture details
- **Benchmarks:** See `/tmp/bench-comparison.md` for performance data

## Examples Highlight

### Rate Limiting (Redis Replacement)
```json
{
  "commands": [
    {
      "type": "IF",
      "condition": "${{ratelimit/user123/count}} < 100",
      "if_true": {
        "type": "COMMANDS",
        "commands": [
          {"type": "INC", "key": "ratelimit/user123/count", "value": 1},
          {"type": "RETURN", "key": {"allowed": true}}
        ]
      },
      "if_false": {
        "type": "RETURN",
        "key": {"allowed": false, "retry_after": 60}
      }
    }
  ]
}
```

### Reactive Workflow with Triggers
```json
// Setup auto-complete trigger
{
  "key": "jobs/*/progress",
  "command": {
    "type": "IF",
    "condition": "${{jobs/${{1}}/progress}} >= ${{jobs/${{1}}/total}}",
    "if_true": {
      "type": "REPLACE",
      "key": "jobs/${{1}}/status",
      "value": "complete"
    }
  }
}
```

### Multi-tenant Bulk Operations
```json
{
  "commands": [
    {
      "type": "FOR",
      "loop_expr": "${{tenants/*/users/*/quota}}",
      "command": {
        "type": "INC",
        "key": "tenants/${{1}}/users/${{2}}/quota",
        "value": 100
      }
    }
  ]
}
```

## Troubleshooting

### Server Not Responding
- Verify server is running: `curl http://localhost:8080/api/v1/caches`
- Check port configuration in environment variables
- Review server logs for errors

### Cache Not Found Errors
- Ensure cache exists: Run "Get Cache List" to verify
- Check `X-Cache-Name` header is set correctly
- Cache names are case-sensitive

### Expression Errors
- Verify expression syntax with govaluate rules
- Check that all referenced keys exist
- Use `PRINT` to debug intermediate values

### Trigger Not Firing
- Verify pattern matches the key being updated
- Check trigger ID was returned on creation
- Review trigger recursion depth (max: 10)
- Ensure trigger command is valid

## Contributing

Found issues or have suggestions for new examples? Please open an issue or PR in the repository.
