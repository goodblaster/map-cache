package v1keys

import (
	"context"
	"net/http"
	"reflect"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// handlePatchRequest represents the request body for patching a single cache value.
type handlePatchRequest struct {
	Operations []patchOperation `json:"operations,required"`
	Flags      map[string]any   `json:"flags,omitempty"` // Optional flags for the patch operation
}

type patchOperation struct {
	Type  PatchCommandType `json:"type"`
	Key   string           `json:"key"`
	Value any              `json:"value,omitempty"`
}

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

type PatchOperationResponse struct {
	Result any   `json:"result"`
	Error  error `json:"error,omitempty"` // Error message if the operation failed
}

type PatchResponse struct {
	Results []PatchOperationResponse `json:"results"` // Results of each operation
}

func (req handlePatchRequest) Validate() error {
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

func (req handlePatchRequest) ValidateOperations(ctx context.Context, cache *caches.Cache) error {
	// Validate the keys exist in the cache (or don't)
	for _, op := range req.Operations {
		v, err := cache.Get(ctx, op.Key)

		switch op.Type {
		case PatchDelete:
			// Ok if there's nothing to delete
			break

		case PatchCreate:
			if err == nil {
				return errors.New("key already exists for create operation: " + op.Key)
			}

		case PatchReplace, PatchIncrement, PatchAppend, PatchResizeArray:
			// Path must exist for these operations
			if err != nil {
				return err
			}

			// PatchIncrement must contain a numeric value
			if op.Type == PatchIncrement {
				if v == nil {
					return errors.New("no value to increment for key: " + op.Key)
				}
				if _, ok := caches.ToFloat64(op.Value); !ok {
					return errors.New("non-numeric value for increment operation on key: " + op.Key)
				}
			}

			// PatchAppend and PatchResizeArray must contain an array value
			if (op.Type == PatchAppend || op.Type == PatchResizeArray) && v != nil {
				if reflect.TypeOf(v).Kind() != reflect.Slice {
					return errors.New("can only append or resize arrays, but found non-array value for key: " + op.Key)
				}
			}

			// PatchResizeArray must contain a valid integer value
			if op.Type == PatchResizeArray {
				if op.Value == nil {
					return errors.New("resize operation requires a numeric value for key: " + op.Key)
				}
				if newSize, ok := caches.ToInt64(op.Value); !ok || newSize < 0 {
					return errors.New("resize operation requires a valid numeric value for key: " + op.Key)
				}
			}
		}
	}
	return nil
}

// handlePatch applies a series of patch operations to the cache.
func handlePatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		cache := Cache(c)

		var input handlePatchRequest
		if err := c.Bind(&input); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := input.Validate(); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request body"))
		}

		if err := input.ValidateOperations(ctx, cache); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid operation(s)"))
		}

		var results PatchResponse

		// Process each operation
		for _, op := range input.Operations {
			var result PatchOperationResponse
			switch op.Type {
			case PatchCreate:
				if err := cache.Create(ctx, map[string]any{op.Key: op.Value}); err != nil {
					result.Error = errors.Wrap(err, "failed to create key")
					break
				}

			case PatchReplace:
				if err := cache.Replace(ctx, op.Key, op.Value); err != nil {
					result.Error = errors.Wrap(err, "failed to replace key")
					break
				}

			case PatchDelete:
				if err := cache.Delete(ctx, op.Key); err != nil {
					result.Error = errors.Wrap(err, "failed to delete key")
					break
				}

			case PatchIncrement:
				f64, err := cache.Increment(ctx, op.Key, op.Value)
				if err != nil {
					result.Error = errors.Wrap(err, "failed to increment key")
					break
				}
				result.Result = f64

			case PatchAppend:
				if op.Value == nil {
					result.Error = errors.New("append value cannot be nil")
					break
				}

				if err := cache.ArrayAppend(ctx, op.Key, op.Value); err != nil {
					result.Error = errors.Wrap(err, "append failed")
					break
				}

			case PatchResizeArray:
				if op.Value == nil {
					result.Error = errors.New("resize value cannot be nil")
					break
				}

				newSize, ok := caches.ToInt64(op.Value)
				if !ok {
					result.Error = errors.New("resize value must be an integer")
					break
				}

				if err := cache.ArrayResize(ctx, op.Key, int(newSize)); err != nil {
					result.Error = errors.Wrap(err, "failed to resize array")
					break
				}

			default:
				result.Error = errors.New("invalid operation type: " + string(op.Type))
			}

			results.Results = append(results.Results, result)
			if result.Error != nil {
				// If any operation fails, we return the error immediately
				return v1errors.ApiError(c, http.StatusInternalServerError, results)
			}
		}

		return c.JSON(http.StatusOK, results)
	}
}
