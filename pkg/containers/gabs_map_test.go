package containers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGabsMap_WildKeys_SingleWildcard(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create data with multiple users
	err := gMap.Set(ctx, map[string]any{
		"users": map[string]any{
			"alice": map[string]any{"name": "Alice", "age": 30},
			"bob":   map[string]any{"name": "Bob", "age": 25},
			"carol": map[string]any{"name": "Carol", "age": 35},
		},
	})
	require.NoError(t, err)

	// Test: Match users/*/name
	keys := gMap.WildKeys(ctx, "users/*/name")

	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "users/alice/name")
	assert.Contains(t, keys, "users/bob/name")
	assert.Contains(t, keys, "users/carol/name")
}

func TestGabsMap_WildKeys_MultipleWildcards(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Nested structure with multiple levels
	err := gMap.Set(ctx, map[string]any{
		"data": map[string]any{
			"group1": map[string]any{
				"item1": map[string]any{"value": 10},
				"item2": map[string]any{"value": 20},
			},
			"group2": map[string]any{
				"item3": map[string]any{"value": 30},
			},
		},
	})
	require.NoError(t, err)

	// Test: Match data/*/*/value
	keys := gMap.WildKeys(ctx, "data/*/*/value")

	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "data/group1/item1/value")
	assert.Contains(t, keys, "data/group1/item2/value")
	assert.Contains(t, keys, "data/group2/item3/value")
}

func TestGabsMap_WildKeys_WithArrays(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Structure with arrays
	err := gMap.Set(ctx, map[string]any{
		"items": []any{
			map[string]any{"name": "item1", "status": "active"},
			map[string]any{"name": "item2", "status": "inactive"},
			map[string]any{"name": "item3", "status": "active"},
		},
	})
	require.NoError(t, err)

	// Test: Match items/*/status
	keys := gMap.WildKeys(ctx, "items/*/status")

	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "items/0/status")
	assert.Contains(t, keys, "items/1/status")
	assert.Contains(t, keys, "items/2/status")
}

func TestGabsMap_WildKeys_NoMatches(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Simple data
	err := gMap.Set(ctx, map[string]any{
		"users": map[string]any{
			"alice": map[string]any{"name": "Alice"},
		},
	})
	require.NoError(t, err)

	// Test: Pattern that doesn't match anything
	keys := gMap.WildKeys(ctx, "nonexistent/*/value")

	assert.Empty(t, keys)
}

func TestGabsMap_WildKeys_NoWildcards(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	err := gMap.Set(ctx, map[string]any{
		"user": map[string]any{"name": "Alice"},
	})
	require.NoError(t, err)

	// Test: Exact path (no wildcards)
	keys := gMap.WildKeys(ctx, "user/name")

	assert.Len(t, keys, 1)
	assert.Equal(t, "user/name", keys[0])
}

func TestGabsMap_WildKeys_EmptyPath(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	err := gMap.Set(ctx, map[string]any{
		"key": "value",
	})
	require.NoError(t, err)

	// Test: Empty string pattern
	keys := gMap.WildKeys(ctx, "")

	// Empty path returns empty slice (no tokens to match)
	assert.Empty(t, keys)
}

func TestGabsMap_WildKeys_DeeplyNested(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: 5 levels deep
	err := gMap.Set(ctx, map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": "value",
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Test: Match deeply nested path
	keys := gMap.WildKeys(ctx, "a/b/c/d/e")

	assert.Len(t, keys, 1)
	assert.Equal(t, "a/b/c/d/e", keys[0])
}

func TestGabsMap_ArrayResize_Grow(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create array with 3 elements
	err := gMap.Set(ctx, []any{"a", "b", "c"}, "items")
	require.NoError(t, err)

	// Test: Grow to 5 elements
	err = gMap.ArrayResize(ctx, 5, "items")
	assert.NoError(t, err)

	// Verify: Should have 5 elements with last 2 being nil
	data, err := gMap.Get(ctx, "items")
	require.NoError(t, err)

	arr, ok := data.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 5)
	assert.Equal(t, "a", arr[0])
	assert.Equal(t, "b", arr[1])
	assert.Equal(t, "c", arr[2])
	assert.Nil(t, arr[3])
	assert.Nil(t, arr[4])
}

func TestGabsMap_ArrayResize_Shrink(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create array with 5 elements
	err := gMap.Set(ctx, []any{"a", "b", "c", "d", "e"}, "items")
	require.NoError(t, err)

	// Test: Shrink to 2 elements
	err = gMap.ArrayResize(ctx, 2, "items")
	assert.NoError(t, err)

	// Verify: Should have 2 elements
	data, err := gMap.Get(ctx, "items")
	require.NoError(t, err)

	arr, ok := data.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 2)
	assert.Equal(t, "a", arr[0])
	assert.Equal(t, "b", arr[1])
}

func TestGabsMap_ArrayResize_SameSize(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create array with 3 elements
	err := gMap.Set(ctx, []any{"a", "b", "c"}, "items")
	require.NoError(t, err)

	// Test: Resize to same size (no-op)
	err = gMap.ArrayResize(ctx, 3, "items")
	assert.NoError(t, err)

	// Verify: Should still have 3 elements unchanged
	data, err := gMap.Get(ctx, "items")
	require.NoError(t, err)

	arr, ok := data.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 3)
	assert.Equal(t, "a", arr[0])
}

func TestGabsMap_ArrayResize_NegativeSize(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	err := gMap.Set(ctx, []any{"a", "b"}, "items")
	require.NoError(t, err)

	// Test: Negative size should error
	err = gMap.ArrayResize(ctx, -1, "items")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
}

func TestGabsMap_ArrayResize_PathNotFound(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Test: Resize non-existent path
	err := gMap.ArrayResize(ctx, 5, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestGabsMap_ArrayResize_NotAnArray(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create a non-array value
	err := gMap.Set(ctx, "string value", "notAnArray")
	require.NoError(t, err)

	// Test: Try to resize non-array
	err = gMap.ArrayResize(ctx, 5, "notAnArray")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a slice or array")
}

func TestGabsMap_ArrayResize_EmptyArray(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create empty array
	err := gMap.Set(ctx, []any{}, "items")
	require.NoError(t, err)

	// Test: Resize to 0 (should be no-op)
	err = gMap.ArrayResize(ctx, 0, "items")
	assert.NoError(t, err)

	// Test: Resize to 3
	err = gMap.ArrayResize(ctx, 3, "items")
	assert.NoError(t, err)

	data, err := gMap.Get(ctx, "items")
	require.NoError(t, err)

	arr, ok := data.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 3)
}

func TestGabsMap_Get_ExistingKey(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	err := gMap.Set(ctx, "test value", "key")
	require.NoError(t, err)

	// Test
	data, err := gMap.Get(ctx, "key")
	assert.NoError(t, err)
	assert.Equal(t, "test value", data)
}

func TestGabsMap_Get_NonExistentKey(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Test
	_, err := gMap.Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestGabsMap_Get_DeeplyNested(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: 5 levels deep
	err := gMap.Set(ctx, map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": "deeply nested value",
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Test
	data, err := gMap.Get(ctx, "a", "b", "c", "d", "e")
	assert.NoError(t, err)
	assert.Equal(t, "deeply nested value", data)
}

func TestGabsMap_Set_CreateNested(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Test: Set creates nested structure automatically
	err := gMap.Set(ctx, "value", "path", "to", "nested", "key")
	assert.NoError(t, err)

	// Verify
	data, err := gMap.Get(ctx, "path", "to", "nested", "key")
	assert.NoError(t, err)
	assert.Equal(t, "value", data)
}

func TestGabsMap_Set_OverwriteDifferentType(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create string value
	err := gMap.Set(ctx, "string", "key")
	require.NoError(t, err)

	// Test: Overwrite with number
	err = gMap.Set(ctx, 42, "key")
	assert.NoError(t, err)

	// Verify - Gabs preserves Go types when set directly
	data, err := gMap.Get(ctx, "key")
	assert.NoError(t, err)
	assert.Equal(t, 42, data)
}

func TestGabsMap_Exists_True(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	err := gMap.Set(ctx, "value", "key")
	require.NoError(t, err)

	// Test
	exists := gMap.Exists(ctx, "key")
	assert.True(t, exists)
}

func TestGabsMap_Exists_False(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Test
	exists := gMap.Exists(ctx, "nonexistent")
	assert.False(t, exists)
}

func TestGabsMap_Exists_PartialPath(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	err := gMap.Set(ctx, "value", "path", "to", "key")
	require.NoError(t, err)

	// Test: Check intermediate paths
	assert.True(t, gMap.Exists(ctx, "path"))
	assert.True(t, gMap.Exists(ctx, "path", "to"))
	assert.True(t, gMap.Exists(ctx, "path", "to", "key"))
	assert.False(t, gMap.Exists(ctx, "path", "to", "key", "deeper"))
}

func TestGabsMap_Delete_ExistingKey(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	err := gMap.Set(ctx, "value", "key")
	require.NoError(t, err)

	// Test
	err = gMap.Delete(ctx, "key")
	assert.NoError(t, err)

	// Verify
	exists := gMap.Exists(ctx, "key")
	assert.False(t, exists)
}

func TestGabsMap_Delete_NestedKey(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	err := gMap.Set(ctx, map[string]any{
		"parent": map[string]any{
			"child1": "value1",
			"child2": "value2",
		},
	})
	require.NoError(t, err)

	// Test: Delete one child
	err = gMap.Delete(ctx, "parent", "child1")
	assert.NoError(t, err)

	// Verify: child1 gone, child2 still exists
	assert.False(t, gMap.Exists(ctx, "parent", "child1"))
	assert.True(t, gMap.Exists(ctx, "parent", "child2"))
}

func TestGabsMap_ArrayAppend_Success(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create array
	err := gMap.Set(ctx, []any{"a", "b"}, "items")
	require.NoError(t, err)

	// Test
	err = gMap.ArrayAppend(ctx, "c", "items")
	assert.NoError(t, err)

	// Verify
	data, err := gMap.Get(ctx, "items")
	require.NoError(t, err)

	arr, ok := data.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 3)
	assert.Equal(t, "c", arr[2])
}

func TestGabsMap_ArrayRemove_Success(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup: Create array
	err := gMap.Set(ctx, []any{"a", "b", "c"}, "items")
	require.NoError(t, err)

	// Test: Remove index 1
	err = gMap.ArrayRemove(ctx, 1, "items")
	assert.NoError(t, err)

	// Verify
	data, err := gMap.Get(ctx, "items")
	require.NoError(t, err)

	arr, ok := data.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 2)
	assert.Equal(t, "a", arr[0])
	assert.Equal(t, "c", arr[1]) // "b" removed, "c" shifted
}

func TestGabsMap_Data_ReturnsMap(t *testing.T) {
	ctx := context.Background()
	gMap := NewGabsMap()

	// Setup
	testData := map[string]any{
		"key1": "value1",
		"key2": 42,
	}
	err := gMap.Set(ctx, testData)
	require.NoError(t, err)

	// Test
	data := gMap.Data(ctx)

	assert.NotNil(t, data)
	assert.Equal(t, "value1", data["key1"])
	assert.Equal(t, 42, data["key2"]) // Preserves Go int type
}

func TestGabsMap_NewGabsMap(t *testing.T) {
	gMap := NewGabsMap()

	assert.NotNil(t, gMap)
	assert.NotNil(t, gMap.container)
}
