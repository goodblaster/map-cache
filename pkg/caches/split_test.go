package caches

import (
	"testing"

	"github.com/goodblaster/map-cache/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSplitKey(t *testing.T) {
	// Save original delimiter
	originalDelimiter := config.KeyDelimiter
	defer func() { config.KeyDelimiter = originalDelimiter }()

	tests := []struct {
		name      string
		delimiter string
		key       string
		expected  []string
	}{
		{
			name:      "simple path",
			delimiter: "/",
			key:       "users/123/name",
			expected:  []string{"users", "123", "name"},
		},
		{
			name:      "single segment",
			delimiter: "/",
			key:       "users",
			expected:  []string{"users"},
		},
		{
			name:      "empty string",
			delimiter: "/",
			key:       "",
			expected:  []string{""},
		},
		{
			name:      "custom delimiter",
			delimiter: ".",
			key:       "users.123.name",
			expected:  []string{"users", "123", "name"},
		},
		{
			name:      "leading delimiter",
			delimiter: "/",
			key:       "/users/123",
			expected:  []string{"", "users", "123"},
		},
		{
			name:      "trailing delimiter",
			delimiter: "/",
			key:       "users/123/",
			expected:  []string{"users", "123", ""},
		},
		{
			name:      "multiple consecutive delimiters",
			delimiter: "/",
			key:       "users//123",
			expected:  []string{"users", "", "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.KeyDelimiter = tt.delimiter
			result := SplitKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkSplitKey(b *testing.B) {
	config.KeyDelimiter = "/"

	benchmarks := []struct {
		name string
		key  string
	}{
		{"short", "user"},
		{"medium", "users/123/profile"},
		{"long", "users/123/profile/settings/preferences/theme"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				SplitKey(bm.key)
			}
		})
	}
}
