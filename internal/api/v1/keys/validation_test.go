package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateKeysRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		req       createKeysRequest
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid request",
			req: createKeysRequest{
				Entries: map[string]any{"key1": "value1"},
			},
			expectErr: false,
		},
		{
			name: "valid request with TTL",
			req: createKeysRequest{
				Entries: map[string]any{"key1": "value1"},
				TTL:     map[string]int64{"key1": 1000},
			},
			expectErr: false,
		},
		{
			name:      "empty entries",
			req:       createKeysRequest{Entries: map[string]any{}},
			expectErr: true,
			errMsg:    "at least one entry is required",
		},
		{
			name: "empty key",
			req: createKeysRequest{
				Entries: map[string]any{"": "value"},
			},
			expectErr: true,
			errMsg:    "key cannot be empty",
		},
		{
			name: "TTL for non-existent key",
			req: createKeysRequest{
				Entries: map[string]any{"key1": "value1"},
				TTL:     map[string]int64{"key2": 1000},
			},
			expectErr: true,
			errMsg:    "TTL specified for non-existent key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandlePutRequest_Validate(t *testing.T) {
	// handlePutRequest.Validate() allows any value including nil (represents JSON null)
	tests := []struct {
		name  string
		req   handlePutRequest
		valid bool
	}{
		{
			name:  "string value",
			req:   handlePutRequest{Value: "test"},
			valid: true,
		},
		{
			name:  "number value",
			req:   handlePutRequest{Value: 42},
			valid: true,
		},
		{
			name:  "nil value (JSON null)",
			req:   handlePutRequest{Value: nil},
			valid: true,
		},
		{
			name:  "object value",
			req:   handlePutRequest{Value: map[string]any{"nested": "value"}},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			assert.NoError(t, err) // All values are valid
		})
	}
}
