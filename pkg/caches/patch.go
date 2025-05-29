package caches

import (
	"context"

	"github.com/goodblaster/errors"
)

type PatchCommandType string

const (
	PatchCreate      PatchCommandType = "CREATE"  // Like a POS
	PatchReplace     PatchCommandType = "REPLACE" // Like a PUT
	PatchDelete      PatchCommandType = "DELETE"  // Like a DELETE
	PatchIncrement   PatchCommandType = "INC"     // Increment a number by a value
	PatchAppend      PatchCommandType = "APPEND"  // Append a value to an array
	PatchResizeArray PatchCommandType = "RESIZE"  // Resize an array to a specific length
)

var PatchCommandTypes = []PatchCommandType{
	PatchCreate,
	PatchReplace,
	PatchDelete,
	PatchIncrement,
	PatchAppend,
	PatchResizeArray,
}

type PatchOperation struct {
	Type  PatchCommandType `json:"type"`
	Key   string           `json:"key"`             // Slash-separated path, e.g., "user/0/name"
	Value any              `json:"value,omitempty"` // Used by SET, INC, APPEND, etc.
}

func (op PatchOperation) Validate() error {
	return nil
}

type PatchOperationResponse struct {
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"` // Error message if the operation failed
}

type PatchResponse struct {
	Results []PatchOperationResponse `json:"results"` // Results of each operation
}

// PatchRequest represents a request to apply a series of patch operations.
type PatchRequest struct {
	Operations []PatchOperation `json:"operations,required"` // List of operations to apply
	Flags      map[string]any   `json:"flags,omitempty"`     // Optional flags for the patch operation
}

func (req PatchRequest) Validate() error {
	if len(req.Operations) == 0 {
		return errors.New("at least one operation is required")
	}
	for _, op := range req.Operations {
		if op.Key == "" {
			return errors.New("operation key cannot be empty")
		}
		if op.Type == "" {
			return errors.New("operation type cannot be empty")
		}

		switch op.Type {
		case PatchCreate, PatchReplace, PatchDelete, PatchIncrement, PatchAppend, PatchResizeArray:
			// Valid operation types
		default:
			return errors.New("invalid operation type: " + string(op.Type))
		}
	}
	return nil
}

func (cache *Cache) Patch(ctx context.Context, patch PatchRequest) (*PatchResponse, error) {
	if err := patch.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid patch request")
	}

	var results PatchResponse

	// Process each operation
	for _, op := range patch.Operations {
		var result PatchOperationResponse
		switch op.Type {
		case PatchCreate:
			if err := cache.Create(ctx, map[string]any{op.Key: op.Value}); err != nil {
				result.Error = errors.Wrap(err, "failed to create key").Error()
				break
			}
			result.Result = op.Value

		case PatchReplace:
			if err := cache.Replace(ctx, op.Key, op.Value); err != nil {
				result.Error = errors.Wrap(err, "failed to replace key").Error()
				break
			}
			result.Result = op.Value

		case PatchDelete:
			if err := cache.Delete(ctx, op.Key); err != nil {
				result.Error = errors.Wrap(err, "failed to delete key").Error()
				break
			}
			result.Result = nil // Deletion doesn't return a value

		case PatchIncrement:
			if op.Value == nil {
				result.Error = "increment value cannot be nil"
				break
			}

			inc, ok := ToFloat64(op.Value)
			if !ok {
				result.Error = "increment value must be an integer"
				break
			}

			v, err := cache.Get(ctx, op.Key)
			if err != nil {
				result.Error = errors.Wrap(err, "failed to increment key").Error()
				break
			}

			f64, ok := ToFloat64(v)
			if !ok {
				result.Error = "value is not a number"
				break
			}

			f64 += inc
			err = cache.Replace(ctx, op.Key, f64)
			if err != nil {
				result.Error = errors.Wrap(err, "failed to increment key").Error()
				break
			}

			result.Result = f64

		case PatchAppend:
			if op.Value == nil {
				result.Error = "append value cannot be nil"
				break
			}

			newValue, err := cache.Append(ctx, op.Key, op.Value)
			if err != nil {
				result.Error = errors.Wrap(err, "failed to append to key").Error()
			} else {
				result.Result = newValue
			}

		case PatchResizeArray:
			if op.Value == nil {
				result.Error = "resize value cannot be nil"
			} else {
				newSize, ok := op.Value.(int)
				if !ok {
					result.Error = "resize value must be an integer"
				} else {
					newValue, err := cache.ResizeArray(ctx, op.Key, newSize)
					if err != nil {
						result.Error = errors.Wrap(err, "failed to resize array").Error()
					} else {
						result.Result = newValue
					}
				}
			}
		default:
			result.Error = "unknown operation type: " + string(op.Type)
		}
		results.Results = append(results.Results, result)
	}
	return &results, nil
}
