package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandlePutBody - Body for the HandlePut function.
type HandlePutBody any

func (v V1) HandlePut() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := v.Cache(c)
		key := c.Param("key")

		// todo - remove this or add it everywhere?
		if key == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "key is required")
		}

		var body HandlePutBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		// TODO: System to filter error to a WebError?
		if err := cache.Replace(c.Request().Context(), key, body); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, v.WebError(err))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}

// HandlePutBatchBody - Body for the HandlePutBatch function.
type HandlePutBatchBody struct {
	Values   map[string]any `json:"values"`   // key/value pairs - Must be new
	Triggers []string       `json:"triggers"` // for now ...
}

// HandlePutBatch - Handler for batch modifying values in the cache.
func (v V1) HandlePutBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := v.Cache(c)
		var body HandlePutBatchBody
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		// TODO: System to filter error to a WebError?
		if err := cache.ReplaceBatch(c.Request().Context(), body.Values); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, v.WebError(err))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}
