package commands

import (
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/stretchr/testify/assert"
)

func TestCommandRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		req       commandRequest
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid request with commands",
			req: commandRequest{
				Commands: []caches.RawCommand{
					{Command: caches.NOOP()},
				},
			},
			expectErr: false,
		},
		{
			name: "valid request with multiple commands",
			req: commandRequest{
				Commands: []caches.RawCommand{
					{Command: caches.NOOP()},
					{Command: caches.NOOP()},
				},
			},
			expectErr: false,
		},
		{
			name:      "empty commands array",
			req:       commandRequest{Commands: []caches.RawCommand{}},
			expectErr: true,
			errMsg:    "at least one command is required",
		},
		{
			name:      "nil commands array",
			req:       commandRequest{Commands: nil},
			expectErr: true,
			errMsg:    "at least one command is required",
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
