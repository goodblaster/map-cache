package v1keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/labstack/echo/v4"
)

// handlePutRequest - Body for the HandlePut function.
type handlePutRequest struct {
	Value any `json:"value"`
}

func handlePut() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		var req handlePutRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := cache.Replace(c.Request().Context(), key, req.Value); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not replace contents"))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}

// replaceBatchRequest - Body for the HandleReplaceBatch function.
type replaceBatchRequest struct {
	Entries map[string]any `json:"entries"`
}

// handleReplaceBatch - Handler for batch modifying values in the cache.
func handleReplaceBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req replaceBatchRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := cache.ReplaceBatch(c.Request().Context(), req.Entries); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not replace contents"))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}
