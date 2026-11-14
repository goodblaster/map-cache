package resp

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	respProto "github.com/tidwall/resp"
)

func TestIncr_OnValueWrittenBySet(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "n", "10", 0).Err())

	n, err := client.Incr(ctx, "n").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(11), n)

	n, err = client.DecrBy(ctx, "n", 5).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(6), n)

	got, err := client.Get(ctx, "n").Result()
	require.NoError(t, err)
	assert.Equal(t, "6", got)

	require.NoError(t, client.Set(ctx, "word", "abc", 0).Err())
	assert.ErrorContains(t, client.Incr(ctx, "word").Err(), "not an integer")
}

func TestIncr_LargeValuesStayExactAndOverflowIsRejected(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	// 2^53 + 1 is not representable as a float64.
	require.NoError(t, client.Set(ctx, "big", "9007199254740992", 0).Err())
	n, err := client.Incr(ctx, "big").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(9007199254740993), n)
	n, err = client.Incr(ctx, "big").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(9007199254740994), n)

	require.NoError(t, client.Set(ctx, "max", "9223372036854775807", 0).Err())
	assert.ErrorContains(t, client.Incr(ctx, "max").Err(), "overflow")
}

func TestDel_CountsOnlyExistingKeys(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "a", "1", 0).Err())

	n, err := client.Del(ctx, "a", "missing", "also-missing").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestGetEx_Options(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "k", "v", time.Minute).Err())
	val, err := client.Do(ctx, "GETEX", "k", "PERSIST").Text()
	require.NoError(t, err)
	assert.Equal(t, "v", val)

	ttl, err := client.TTL(ctx, "k").Result()
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl, "PERSIST must remove the TTL")

	assert.ErrorContains(t, client.Do(ctx, "GETEX", "k", "EX", "0").Err(), "invalid expire time")

	// An absolute expiry in the past removes the key.
	val, err = client.Do(ctx, "GETEX", "k", "EXAT", "1").Text()
	require.NoError(t, err)
	assert.Equal(t, "v", val)
	exists, err := client.Exists(ctx, "k").Result()
	require.NoError(t, err)
	assert.Zero(t, exists)
}

func TestSet_RejectsNonPositiveExpiry(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	assert.ErrorContains(t, client.Do(ctx, "SET", "k", "v", "EX", "0").Err(), "invalid expire time")
	assert.ErrorContains(t, client.Do(ctx, "SETEX", "k", "-1", "v").Err(), "invalid expire time")
	assert.ErrorContains(t, client.Do(ctx, "PSETEX", "k", "0", "v").Err(), "invalid expire time")

	exists, err := client.Exists(ctx, "k").Result()
	require.NoError(t, err)
	assert.Zero(t, exists)
}

func TestCopy_IsIndependentOfSource(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.HSet(ctx, "h1", "f", "v").Err())
	require.NoError(t, client.Copy(ctx, "h1", "h2", 0, false).Err())
	require.NoError(t, client.HSet(ctx, "h2", "g", "w").Err())

	h1, err := client.HGetAll(ctx, "h1").Result()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"f": "v"}, h1)

	require.NoError(t, client.RPush(ctx, "l1", "a", "b").Err())
	require.NoError(t, client.Copy(ctx, "l1", "l2", 0, false).Err())
	require.NoError(t, client.LSet(ctx, "l2", 0, "changed").Err())

	l1, err := client.LRange(ctx, "l1", 0, -1).Result()
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, l1)
}

func TestLPush_MultipleValuesOrder(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	// Redis pushes one at a time, so the last argument ends up at the head.
	n, err := client.LPush(ctx, "l", "a", "b", "c").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)

	n, err = client.LPushX(ctx, "l", "d", "e").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)

	list, err := client.LRange(ctx, "l", 0, -1).Result()
	require.NoError(t, err)
	assert.Equal(t, []string{"e", "d", "c", "b", "a"}, list)
}

func TestLRem_NegativeCountRemovesFromTail(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.RPush(ctx, "l", "x", "a", "x", "b", "x").Err())
	n, err := client.LRem(ctx, "l", -2, "x").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	list, err := client.LRange(ctx, "l", 0, -1).Result()
	require.NoError(t, err)
	assert.Equal(t, []string{"x", "a", "b"}, list)
}

func TestLPos_CountBeforeOtherOptions(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.RPush(ctx, "l", "a", "b", "a").Err())
	positions, err := client.Do(ctx, "LPOS", "l", "a", "COUNT", "0", "RANK", "1").Int64Slice()
	require.NoError(t, err)
	assert.Equal(t, []int64{0, 2}, positions)
}

func TestStringCommands_FormatLargeNumbersConsistently(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.IncrBy(ctx, "n", 1000000).Err())

	length, err := client.StrLen(ctx, "n").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(7), length)

	require.NoError(t, client.Append(ctx, "n", "x").Err())
	got, err := client.Get(ctx, "n").Result()
	require.NoError(t, err)
	assert.Equal(t, "1000000x", got)
}

func TestSetRange_RejectsHugeOffset(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	err := client.SetRange(ctx, "k", 9223372036854775807, "x").Err()
	assert.ErrorContains(t, err, "maximum allowed size")
	err = client.SetRange(ctx, "k", 1<<40, "x").Err()
	assert.ErrorContains(t, err, "maximum allowed size")

	// Server is still healthy.
	require.NoError(t, client.Ping(ctx).Err())
}

func TestKeys_RoundTripsKeyTranslation(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "user:1:name", "a", 0).Err())
	require.NoError(t, client.Set(ctx, "user:2:name", "b", 0).Err())

	keys, err := client.Keys(ctx, "user:*:name").Result()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"user:1:name", "user:2:name"}, keys)
}

func TestGetSet_KeyHoldingNull(t *testing.T) {
	_, client := startTestServer(t)
	ctx := context.Background()

	require.NoError(t, client.HSet(ctx, "h", "f", "v").Err())
	// GETSET on a missing key returns nil and creates it.
	err := client.GetSet(ctx, "fresh", "v1").Err()
	assert.ErrorIs(t, err, redis.Nil)
	old, err := client.GetSet(ctx, "fresh", "v2").Result()
	require.NoError(t, err)
	assert.Equal(t, "v1", old)
}

func TestPanicInHandler_DoesNotKillServer(t *testing.T) {
	RegisterCommand("TESTPANIC", func(s *Session, args []respProto.Value) error {
		var m map[string]int
		m["boom"] = 1 // nil map write
		return nil
	})
	t.Cleanup(func() { delete(commandRegistry, "TESTPANIC") })

	_, client := startTestServer(t)
	ctx := context.Background()

	assert.ErrorContains(t, client.Do(ctx, "TESTPANIC").Err(), "internal error")

	// Same connection and the server as a whole keep working.
	require.NoError(t, client.Set(ctx, "k", "v", 0).Err())
	got, err := client.Get(ctx, "k").Result()
	require.NoError(t, err)
	assert.Equal(t, "v", got)
}
