package v1keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/logos"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/labstack/echo/v4"
)

// handlePatchRequest represents the request body for patching a single cache value.
type handlePatchRequest struct {
	Operations []patchOperation `json:"operations,required"`
	Flags      map[string]any   `json:"flags,omitempty"` // Optional flags for the patch operation
}

type patchOperation struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value any    `json:"value,omitempty"`
}

func (req handlePatchRequest) Validate() error {
	return nil
}

// handlePatch applies a series of patch operations to the cache.
func handlePatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)

		var input handlePatchRequest
		if err := c.Bind(&input); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := input.Validate(); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request body"))
		}

		if err := cache.Patch(c.Request().Context()); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not apply entire patch"))
		}

		return c.NoContent(http.StatusOK)
	}
}
