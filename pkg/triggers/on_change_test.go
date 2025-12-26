package triggers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSame_Primitives tests primitive type comparisons
func TestSame_Primitives(t *testing.T) {
	testCases := []struct {
		name     string
		a        any
		b        any
		expected bool
	}{
		// Strings
		{"identical strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},
		{"empty strings", "", "", true},

		// Numbers
		{"identical int", 42, 42, true},
		{"different int", 42, 43, false},
		{"identical float64", 3.14, 3.14, true},
		{"different float64", 3.14, 3.15, false},
		{"zero int", 0, 0, true},

		// Booleans
		{"both true", true, true, true},
		{"both false", false, false, true},
		{"different bool", true, false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Same(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSame_TypeMismatches tests comparisons between different types
func TestSame_TypeMismatches(t *testing.T) {
	testCases := []struct {
		name string
		a    any
		b    any
	}{
		{"int vs string", 42, "42"},
		{"int vs float64", 5, 5.0},
		{"string vs bool", "true", true},
		{"int vs bool", 1, true},
		{"string vs map", "hello", map[string]any{"key": "value"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Same(tc.a, tc.b)
			assert.False(t, result, "Different types should not be same")
		})
	}
}

// TestSame_NilHandling tests nil value comparisons
func TestSame_NilHandling(t *testing.T) {
	testCases := []struct {
		name     string
		a        any
		b        any
		expected bool
	}{
		{"both nil", nil, nil, true},
		{"nil vs string", nil, "value", false},
		{"string vs nil", "value", nil, false},
		{"nil vs number", nil, 42, false},
		{"nil vs bool", nil, false, false},
		{"nil vs map", nil, map[string]any{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Same(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSame_MapComparison tests map[string]any comparisons
func TestSame_MapComparison(t *testing.T) {
	testCases := []struct {
		name     string
		a        any
		b        any
		expected bool
	}{
		{
			name:     "identical simple maps",
			a:        map[string]any{"a": 1, "b": 2},
			b:        map[string]any{"a": 1, "b": 2},
			expected: true,
		},
		{
			name:     "identical single key",
			a:        map[string]any{"key": "value"},
			b:        map[string]any{"key": "value"},
			expected: true,
		},
		{
			name:     "different keys",
			a:        map[string]any{"a": 1},
			b:        map[string]any{"b": 1},
			expected: false,
		},
		{
			name:     "different values",
			a:        map[string]any{"a": 1},
			b:        map[string]any{"a": 2},
			expected: false,
		},
		{
			name:     "different lengths",
			a:        map[string]any{"a": 1},
			b:        map[string]any{"a": 1, "b": 2},
			expected: false,
		},
		{
			name:     "empty maps",
			a:        map[string]any{},
			b:        map[string]any{},
			expected: true,
		},
		{
			name:     "nested identical maps",
			a:        map[string]any{"outer": map[string]any{"inner": "value"}},
			b:        map[string]any{"outer": map[string]any{"inner": "value"}},
			expected: true,
		},
		{
			name:     "nested different maps",
			a:        map[string]any{"outer": map[string]any{"inner": "value1"}},
			b:        map[string]any{"outer": map[string]any{"inner": "value2"}},
			expected: false,
		},
		{
			name:     "complex nested identical",
			a:        map[string]any{"a": 1, "b": map[string]any{"c": 2, "d": map[string]any{"e": 3}}},
			b:        map[string]any{"a": 1, "b": map[string]any{"c": 2, "d": map[string]any{"e": 3}}},
			expected: true,
		},
		{
			name:     "mixed value types",
			a:        map[string]any{"str": "hello", "num": 42, "bool": true},
			b:        map[string]any{"str": "hello", "num": 42, "bool": true},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Same(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSame_NonComparableTypes tests non-comparable types
func TestSame_NonComparableTypes(t *testing.T) {
	testCases := []struct {
		name string
		a    any
		b    any
	}{
		{
			name: "slices",
			a:    []string{"a", "b"},
			b:    []string{"a", "b"},
		},
		{
			name: "empty slices",
			a:    []int{},
			b:    []int{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Same(tc.a, tc.b)
			// Slices are not comparable, should return false
			assert.False(t, result)
		})
	}
}

// TestSame_EdgeCases tests edge cases and boundary conditions
func TestSame_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		a        any
		b        any
		expected bool
	}{
		{
			name:     "zero vs nil",
			a:        0,
			b:        nil,
			expected: false,
		},
		{
			name:     "empty string vs nil",
			a:        "",
			b:        nil,
			expected: false,
		},
		{
			name:     "false vs nil",
			a:        false,
			b:        nil,
			expected: false,
		},
		{
			name:     "negative numbers",
			a:        -42,
			b:        -42,
			expected: true,
		},
		{
			name:     "large numbers",
			a:        9999999999,
			b:        9999999999,
			expected: true,
		},
		{
			name:     "unicode strings",
			a:        "hello 世界 🌍",
			b:        "hello 世界 🌍",
			expected: true,
		},
		{
			name:     "map with nil value",
			a:        map[string]any{"key": nil},
			b:        map[string]any{"key": nil},
			expected: true,
		},
		{
			name:     "map with nil vs non-nil value",
			a:        map[string]any{"key": nil},
			b:        map[string]any{"key": "value"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Same(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSame_Symmetry verifies that Same is symmetric (a==b implies b==a)
func TestSame_Symmetry(t *testing.T) {
	testCases := []struct {
		name string
		a    any
		b    any
	}{
		{"strings", "hello", "world"},
		{"numbers", 1, 2},
		{"maps", map[string]any{"a": 1}, map[string]any{"b": 2}},
		{"nil vs value", nil, "value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result1 := Same(tc.a, tc.b)
			result2 := Same(tc.b, tc.a)
			assert.Equal(t, result1, result2, "Same should be symmetric")
		})
	}
}

// TestSame_Reflexivity verifies that Same is reflexive (a==a)
func TestSame_Reflexivity(t *testing.T) {
	testCases := []struct {
		name  string
		value any
	}{
		{"string", "hello"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"map", map[string]any{"a": 1, "b": 2}},
		{"empty map", map[string]any{}},
		{"nil", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Same(tc.value, tc.value)
			assert.True(t, result, "Same should be reflexive")
		})
	}
}
