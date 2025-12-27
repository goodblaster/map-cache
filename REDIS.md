# Redis Protocol Support

Map-cache implements the Redis Serialization Protocol (RESP2), allowing you to use standard Redis clients to interact with the cache. This provides a familiar interface for developers already using Redis while leveraging map-cache's unique features like nested paths, triggers, and advanced command execution.

## Quick Start

### Starting the Server

Enable RESP support via environment variable:

```bash
RESP_ENABLED=true ./map-cache
```

The RESP server listens on port `6379` by default (configurable via `RESP_ADDRESS`).

### Using redis-cli

```bash
$ redis-cli
127.0.0.1:6379> SET mykey "Hello World"
OK
127.0.0.1:6379> GET mykey
"Hello World"
127.0.0.1:6379> HSET user:1000 name "Alice" email "alice@example.com"
(integer) 2
127.0.0.1:6379> HGETALL user:1000
1) "name"
2) "Alice"
3) "email"
4) "alice@example.com"
```

### Using Redis Client Libraries

**Go (go-redis):**
```go
import "github.com/redis/go-redis/v9"

client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

ctx := context.Background()
client.Set(ctx, "key", "value", 0)
val, _ := client.Get(ctx, "key").Result()
```

**Python (redis-py):**
```python
import redis

r = redis.Redis(host='localhost', port=6379, decode_responses=True)
r.set('key', 'value')
print(r.get('key'))
```

**Node.js (ioredis):**
```javascript
const Redis = require('ioredis');
const redis = new Redis();

await redis.set('key', 'value');
const value = await redis.get('key');
```

## Supported Commands

Map-cache implements **71 core Redis commands** organized into five categories:

### String Commands (21)

| Command | Description | Example |
|---------|-------------|---------|
| **GET** | Get value of key | `GET mykey` |
| **SET** | Set key to value | `SET mykey "value"` |
| **DEL** | Delete one or more keys | `DEL key1 key2` |
| **EXISTS** | Check if keys exist | `EXISTS key1 key2 key3` |
| **INCR** | Increment integer value by 1 | `INCR counter` |
| **DECR** | Decrement integer value by 1 | `DECR counter` |
| **INCRBY** | Increment by amount | `INCRBY counter 5` |
| **DECRBY** | Decrement by amount | `DECRBY counter 5` |
| **MGET** | Get multiple values | `MGET key1 key2 key3` |
| **MSET** | Set multiple key-value pairs | `MSET k1 v1 k2 v2` |
| **GETSET** | Set new value, return old | `GETSET mykey "new"` |
| **SETNX** | Set if not exists | `SETNX lock "acquired"` |
| **SETEX** | Set with expiration (seconds) | `SETEX key 60 "value"` |
| **PSETEX** | Set with expiration (milliseconds) | `PSETEX key 60000 "value"` |
| **STRLEN** | Get string length | `STRLEN mykey` |
| **APPEND** | Append to string | `APPEND mykey " more"` |
| **GETRANGE** | Get substring by range | `GETRANGE mykey 0 4` |
| **SETRANGE** | Overwrite part of string | `SETRANGE mykey 7 "Redis"` |
| **GETEX** | Get value and set expiration | `GETEX mykey EX 60` |
| **GETDEL** | Get value and delete key | `GETDEL mykey` |
| **INCRBYFLOAT** | Increment by float amount | `INCRBYFLOAT price 2.5` |

### Key Management Commands (14)

| Command | Description | Example |
|---------|-------------|---------|
| **EXPIRE** | Set TTL in seconds | `EXPIRE key 60` |
| **PEXPIRE** | Set TTL in milliseconds | `PEXPIRE key 60000` |
| **PERSIST** | Remove TTL | `PERSIST key` |
| **TTL** | Get remaining TTL (seconds) | `TTL key` |
| **PTTL** | Get remaining TTL (milliseconds) | `PTTL key` |
| **EXPIRETIME** | Get absolute expiration Unix timestamp (seconds) | `EXPIRETIME key` |
| **PEXPIRETIME** | Get absolute expiration Unix timestamp (ms) | `PEXPIRETIME key` |
| **KEYS** | Find keys matching pattern | `KEYS user:*` |
| **EXPIREAT** | Set expiration at Unix timestamp | `EXPIREAT key 1735689600` |
| **PEXPIREAT** | Set expiration at Unix timestamp (ms) | `PEXPIREAT key 1735689600000` |
| **RENAME** | Rename key (preserves TTL) | `RENAME oldkey newkey` |
| **RENAMENX** | Rename if new key doesn't exist | `RENAMENX oldkey newkey` |
| **COPY** | Copy key to new key (preserves TTL) | `COPY source dest REPLACE` |
| **TYPE** | Get key type | `TYPE mykey` |

**TTL Return Values:**
- `-2`: Key doesn't exist
- `-1`: Key exists but has no TTL
- Positive number: Remaining TTL

**EXPIRETIME Return Values:**
- `-2`: Key doesn't exist
- `-1`: Key exists but has no TTL
- Positive number: Unix timestamp when key expires

### Hash Commands (14)

| Command | Description | Example |
|---------|-------------|---------|
| **HGET** | Get hash field value | `HGET user:1 name` |
| **HSET** | Set hash field(s) | `HSET user:1 name "Alice"` |
| **HGETALL** | Get all fields and values | `HGETALL user:1` |
| **HDEL** | Delete hash field(s) | `HDEL user:1 email` |
| **HEXISTS** | Check if field exists | `HEXISTS user:1 name` |
| **HLEN** | Count fields in hash | `HLEN user:1` |
| **HKEYS** | Get all field names | `HKEYS user:1` |
| **HVALS** | Get all field values | `HVALS user:1` |
| **HMGET** | Get multiple field values | `HMGET user:1 name email` |
| **HMSET** | Set multiple fields | `HMSET user:1 name "Alice" age 30` |
| **HINCRBY** | Increment field by integer | `HINCRBY user:1 visits 1` |
| **HINCRBYFLOAT** | Increment field by float | `HINCRBYFLOAT user:1 balance 10.5` |
| **HSETNX** | Set field if not exists | `HSETNX user:1 created_at 1234567890` |
| **HRANDFIELD** | Get random field(s) from hash | `HRANDFIELD user:1 2 WITHVALUES` |

### List Commands (14)

| Command | Description | Example |
|---------|-------------|---------|
| **LPUSH** | Prepend values to list | `LPUSH mylist "value"` |
| **LPUSHX** | Prepend values only if list exists | `LPUSHX mylist "value"` |
| **RPUSH** | Append values to list | `RPUSH mylist "value"` |
| **RPUSHX** | Append values only if list exists | `RPUSHX mylist "value"` |
| **LPOP** | Remove and return first element | `LPOP mylist` |
| **RPOP** | Remove and return last element | `RPOP mylist` |
| **LLEN** | Get list length | `LLEN mylist` |
| **LRANGE** | Get range of elements | `LRANGE mylist 0 -1` |
| **LINDEX** | Get element at index | `LINDEX mylist 0` |
| **LSET** | Set element at index | `LSET mylist 0 "new"` |
| **LTRIM** | Trim list to range | `LTRIM mylist 0 99` |
| **LINSERT** | Insert before/after element | `LINSERT mylist BEFORE "pivot" "value"` |
| **LREM** | Remove elements by value | `LREM mylist 2 "value"` |
| **LPOS** | Find position of element | `LPOS mylist "value"` |

### Generic Commands (8)

| Command | Description | Example |
|---------|-------------|---------|
| **PING** | Test connection | `PING` or `PING "hello"` |
| **ECHO** | Echo message | `ECHO "Hello"` |
| **SELECT** | Switch cache/database | `SELECT 0` |
| **COMMAND** | Get command info | `COMMAND COUNT` |
| **HELLO** | Protocol negotiation | `HELLO` |
| **CLIENT** | Client management | `CLIENT SETINFO` |
| **FLUSHDB** | Delete all keys in current cache | `FLUSHDB` |
| **FLUSHALL** | Delete all keys (same as FLUSHDB) | `FLUSHALL` |

## Key Translation

Map-cache uses `/` as the path delimiter for nested data, while Redis conventionally uses `:`. The RESP server automatically translates between these formats:

**Redis key** → **Map-cache path**
- `user:123:name` → `user/123/name`
- `session:abc:data` → `session/abc/data`
- `config:app:version` → `config/app/version`

This translation is **bidirectional** and **transparent**:

```bash
# Via Redis protocol
redis-cli> SET user:100:email "alice@example.com"
OK

# Via HTTP API (same data)
curl http://localhost:8080/api/v1/keys/user/100/email \
  -H "X-Cache-Name: default"
# Returns: "alice@example.com"
```

### Translation Configuration

Control translation behavior via `RESP_KEY_MODE`:

```bash
# Translate : to / (default, recommended)
RESP_KEY_MODE=translate

# Preserve keys as-is (not recommended)
RESP_KEY_MODE=preserve
```

## Multi-Cache Support (SELECT)

Map-cache supports multiple named caches. The `SELECT` command maps Redis database numbers to cache names:

- **SELECT 0** → `default` cache
- **SELECT 1** → cache named `1`
- **SELECT 2** → cache named `2`
- **SELECT N** → cache named `N` (where N is a number)

### Creating Numbered Caches

Create caches via HTTP API with numeric names:

```bash
# Create cache "1"
curl -X POST http://localhost:8080/api/v1/caches \
  -H "Content-Type: application/json" \
  -d '{"name":"1"}'

# Create cache "2"
curl -X POST http://localhost:8080/api/v1/caches \
  -H "Content-Type: application/json" \
  -d '{"name":"2"}'
```

### Using SELECT

```bash
redis-cli> SELECT 0
OK
redis-cli> SET mykey "in default cache"
OK

redis-cli> SELECT 1
OK
redis-cli> SET mykey "in cache 1"
OK
redis-cli> GET mykey
"in cache 1"

redis-cli> SELECT 0
OK
redis-cli> GET mykey
"in default cache"
```

**Important:** Caches must be named with numbers to be accessible via SELECT. Named caches like "sessions" or "users" can only be accessed via the HTTP API.

## Data Model Mapping

Map-cache stores data as nested JSON structures. Redis data types map naturally:

### Strings

Redis strings map to JSON strings or numbers:

```bash
SET name "Alice"           # → {"name": "Alice"}
SET count 42               # → {"count": 42}
INCR counter               # → {"counter": 1}
```

### Hashes

Redis hashes map to nested JSON objects:

```bash
HSET user:100 name "Alice" email "alice@example.com"
# → {"user": {"100": {"name": "Alice", "email": "alice@example.com"}}}
```

The hash key becomes the parent path, and fields become child keys:
- Hash: `user:100`
- Fields: `name`, `email`
- Stored as: `user/100/name` and `user/100/email`

### Lists

Redis lists map to JSON arrays:

```bash
RPUSH mylist "value1" "value2" "value3"
# → {"mylist": ["value1", "value2", "value3"]}
```

List operations (LPUSH, RPUSH, LPOP, RPOP, etc.) manipulate the underlying JSON array. This provides familiar queue and stack patterns:

```bash
# Queue pattern (FIFO)
RPUSH queue "job1" "job2"
LPOP queue  # Returns "job1"

# Stack pattern (LIFO)
RPUSH stack "item1" "item2"
RPOP stack  # Returns "item2"
```

### Sets and Sorted Sets

**Not implemented.** These Redis data structures don't have natural JSON equivalents in map-cache's architecture.

## Persistence

Map-cache provides persistence through backup and restore operations, similar to Redis `SAVE`/`BGSAVE`:

### Via HTTP API

```bash
# Create backup
curl -X POST "http://localhost:8080/api/v1/admin/backup?cache=default&filename=backup.json"

# Restore from backup
curl -X POST "http://localhost:8080/api/v1/admin/restore?cache=default&filename=backup.json"
```

### Configuration

Set backup directory:

```bash
RESP_BACKUP_DIR=./backups
```

**Note:** Unlike Redis, map-cache does not have automatic persistence (AOF/RDB). Backups must be triggered manually or via scheduled jobs.

## Performance

Map-cache prioritizes **correctness** and **feature richness** over raw speed. Expected performance:

| Operation | Latency | Notes |
|-----------|---------|-------|
| Simple GET | 0.1-1ms | RESP protocol overhead + in-memory lookup |
| Simple SET | 0.1-1ms | Create or replace operation |
| INCR | 0.1-1ms | Atomic increment with auto-initialization |
| HGETALL | 0.5-2ms | Depends on hash size |
| Pattern match (KEYS) | 1-10ms | Iterates all keys, use sparingly |

**Comparison to Redis:**
- Redis: ~0.1-0.2ms for simple operations
- Map-cache: ~0.1-1ms (3-5x slower due to additional abstraction layers)

**Optimization Tips:**
1. Use specific paths instead of wildcards when possible
2. Keep hash sizes reasonable (< 100 fields for best performance)
3. Use batch operations (MGET/MSET) to reduce round trips
4. Avoid `KEYS *` in production (use `KEYS` with specific patterns)

## Advanced Features

Map-cache offers features beyond standard Redis:

### 1. Nested Path Access

Access deeply nested data naturally:

```bash
# Redis protocol
HSET user:100 profile:address:city "San Francisco"

# Internally stored as:
# user/100/profile/address/city = "San Francisco"
```

### 2. Triggers (Not accessible via RESP)

Triggers are configured via HTTP API and fire on key updates:

```bash
curl -X POST http://localhost:8080/api/v1/triggers \
  -H "X-Cache-Name: default" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "users/*/last_login",
    "command": {
      "type": "REPLACE",
      "key": "users/${{2}}/status",
      "value": "active"
    }
  }'
```

Now when `SET user:123:last_login "2023-10-01"` is called via Redis, the trigger fires and sets `user/123/status` to "active".

### 3. Command Execution (Not accessible via RESP)

Execute complex command sequences via HTTP API:

```json
POST /api/v1/commands
{
  "type": "COMMANDS",
  "commands": [
    {"type": "INC", "key": "stats/visits", "increment": 1},
    {"type": "IF", "condition": "${{stats/visits}} > 100", "then": [
      {"type": "REPLACE", "key": "status", "value": "popular"}
    ]}
  ]
}
```

### 4. Wildcard Operations

Map-cache supports wildcard patterns more broadly:

```bash
# KEYS command
KEYS users:*:email

# Returns all email keys like:
# users:100:email
# users:200:email
```

## Limitations and Differences from Redis

### Not Implemented

The following Redis features are **not supported**:

1. **Sets** (SADD, SMEMBERS, SINTER, etc.)
   - Redis uses hash tables for O(1) membership testing
   - JSON arrays require O(n) linear search
   - Workaround: Use objects like `{"member1": true, "member2": true}` for O(1) lookup
   - Limitation: Set operations (SINTER, SUNION) still require iteration and are slower than Redis

2. **Sorted Sets** (ZADD, ZRANGE, ZRANK, etc.)
   - Redis uses skip lists for O(log n) sorted operations
   - JSON has no native sorted structure
   - Arrays require re-sorting on every insert (O(n log n))
   - Objects can't maintain sorted order efficiently
   - Implementing tree structures on top of JSON defeats the purpose of a simple cache

3. **Pub/Sub** (PUBLISH, SUBSCRIBE, PSUBSCRIBE)
   - Different model from map-cache's trigger system
   - Map-cache uses pattern-based triggers (push) vs pub/sub channels (pull)

4. **Transactions** (MULTI/EXEC/WATCH - only basic queuing)
   - Redis provides optimistic locking with WATCH
   - Map-cache has cache-level locking, not key-level

5. **Blocking Operations** (BLPOP, BRPOP)
   - No blocking primitives in map-cache
   - Would require condition variables and significant complexity

6. **Lua Scripting** (EVAL, EVALSHA)
   - Map-cache has its own command DSL (IF, FOR, COMMANDS)
   - Lua → map-cache translation is complex

7. **Clustering/Replication**
   - Single-node architecture

8. **Automatic Persistence** (AOF/RDB)
   - Manual backup/restore only via HTTP API or SAVE/BGSAVE commands

9. **Streams** (XADD, XREAD, etc.)
   - Complex data structure not compatible with JSON storage model

### Behavioral Differences

1. **INCR auto-initialization**: In Redis, INCR on a non-existent key creates it with value 1. Map-cache initializes to 0, then increments to 1 (same result, different implementation).

2. **KEYS performance**: Map-cache iterates all keys in memory. For large datasets, use specific patterns.

3. **Hash storage**: Redis stores hashes as flat key-value pairs. Map-cache stores them as nested JSON objects, which may affect memory usage for deeply nested structures.

4. **TTL precision**: Both use milliseconds internally, but Redis may have tighter precision guarantees.

5. **Error messages**: Error messages may differ slightly from Redis.

## Configuration

All RESP-related settings are controlled via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `RESP_ENABLED` | `false` | Enable RESP server |
| `RESP_ADDRESS` | `:6379` | Listen address |
| `RESP_KEY_MODE` | `translate` | Key translation: `translate` or `preserve` |
| `RESP_DEFAULT_CACHE` | `default` | Default cache name |
| `RESP_MAX_CONNECTIONS` | `1000` | Max concurrent connections |
| `RESP_BACKUP_DIR` | `./backups` | Backup directory |

### Example Configuration

```bash
export RESP_ENABLED=true
export RESP_ADDRESS=:6379
export RESP_KEY_MODE=translate
export RESP_DEFAULT_CACHE=default
export RESP_MAX_CONNECTIONS=1000

./map-cache
```

## Migration from Redis

### Use Case: Drop-in Replacement

Map-cache can replace Redis for applications that use:
- Basic key-value operations (strings)
- Hash operations for structured data
- Simple TTL-based expiration
- Pattern matching via KEYS

### Not Suitable For

Map-cache is **not** suitable if you rely on:
- Lists, sets, or sorted sets
- Pub/Sub messaging
- Lua scripting
- Clustering or replication
- Blocking operations
- Sub-millisecond latency requirements

### Migration Steps

1. **Assess Command Usage**: Check which Redis commands your application uses (see supported list above)

2. **Create Caches**: If using SELECT, create numbered caches:
   ```bash
   curl -X POST http://localhost:8080/api/v1/caches -d '{"name":"1"}'
   ```

3. **Update Connection**: Change connection string to map-cache:
   ```python
   # Before
   r = redis.Redis(host='redis.example.com', port=6379)

   # After
   r = redis.Redis(host='map-cache.example.com', port=6379)
   ```

4. **Test Thoroughly**: Map-cache has different performance characteristics and limitations

5. **Monitor Performance**: Expect 3-5x higher latency than Redis

## Troubleshooting

### Connection Refused

```bash
# Check if RESP is enabled
echo $RESP_ENABLED  # should be "true"

# Check server logs
tail -f /tmp/map-cache-test.log
```

### Unknown Command Error

```bash
ERR unknown command 'LPUSH'
```

Command not implemented. See supported commands list above.

### Cache Not Found

```bash
redis-cli> SELECT 5
OK
redis-cli> SET key value
ERR cache "5" not found
```

Create the cache first via HTTP API:
```bash
curl -X POST http://localhost:8080/api/v1/caches -d '{"name":"5"}'
```

### Key Not Found After SET

Check cache selection:
```bash
redis-cli> SELECT 0
OK
redis-cli> SET mykey "value"
OK
redis-cli> SELECT 1
OK
redis-cli> GET mykey
(nil)  # Wrong cache!
```

### Performance Issues

1. Avoid `KEYS *` on large datasets
2. Use specific patterns: `KEYS user:*` instead of `KEYS *`
3. Batch operations with MGET/MSET
4. Consider using HTTP API for complex operations

## Examples

### Session Storage

```bash
# Create session
SETEX session:abc123 3600 "user_data"

# Get session
GET session:abc123

# Check TTL
TTL session:abc123

# Delete session
DEL session:abc123
```

### User Profiles

```bash
# Create user profile
HSET user:1000 name "Alice" email "alice@example.com" age 30

# Get specific field
HGET user:1000 email

# Get all fields
HGETALL user:1000

# Update field
HSET user:1000 last_login "2023-10-01"

# Delete field
HDEL user:1000 age
```

### Counters and Statistics

```bash
# Increment page views
INCR stats:pageviews

# Increment by amount
INCRBY stats:downloads 5

# Get current value
GET stats:pageviews

# Set expiration
EXPIRE stats:pageviews 86400  # 24 hours
```

### Pattern-Based Queries

```bash
# Store multiple user emails
SET user:100:email "alice@example.com"
SET user:200:email "bob@example.com"
SET user:300:email "charlie@example.com"

# Find all user emails
KEYS user:*:email

# Returns:
# 1) "user:100:email"
# 2) "user:200:email"
# 3) "user:300:email"
```

## Further Reading

- [Map-cache HTTP API Documentation](./README.md)
- [Architecture Overview](./CLAUDE.md)
- [Redis Protocol Specification (RESP2)](https://redis.io/docs/reference/protocol-spec/)

## Summary

Map-cache's Redis protocol support provides:

**Strengths:**
- ✅ Familiar Redis interface for 62 core commands
- ✅ Drop-in replacement for key-value, hash, and list use cases
- ✅ Automatic key translation (`:` ↔ `/`)
- ✅ Multi-cache support via SELECT
- ✅ Hash commands map naturally to nested JSON
- ✅ List commands work with JSON arrays
- ✅ Standard client library compatibility

**Limitations:**
- ❌ No sets or sorted sets
- ❌ No pub/sub or streams
- ❌ No Lua scripting
- ❌ No clustering or replication
- ⚠️ 3-5x slower than Redis (still fast at 0.1-1ms)

Map-cache excels at providing Redis-like operations on nested JSON data with unique features like triggers and complex command execution. It's ideal for applications that need structured data storage with a familiar Redis interface, but not for use cases requiring Redis's advanced data structures or sub-millisecond performance.
