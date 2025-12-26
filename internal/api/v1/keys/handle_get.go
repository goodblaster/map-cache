package keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/labstack/echo/v4"
)

// getBatchRequest represents the request body for retrieving multiple keys.
type getBatchRequest struct {
	Keys []string `json:"keys,required"`
}

func (req getBatchRequest) Validate() error {
	if len(req.Keys) == 0 {
		return errors.New("at least one key is required")
	}
	for _, key := range req.Keys {
		if key == "" {
			return errors.New("key cannot be empty")
		}
	}
	return nil
}

// handleGetValue retrieves a single value from the cache.
func handleGetValue() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		value, err := cache.Get(c.Request().Context(), key)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "key not found").SetInternal(err)
		}

		return c.JSON(http.StatusOK, value)
	}
}

// handleGetBatch retrieves multiple values from the cache.
func handleGetBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req getBatchRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid json payload").SetInternal(err)
		}

		if err := req.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "validation failed").SetInternal(err)
		}

		value, err := cache.BatchGet(c.Request().Context(), req.Keys...)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "key not found").SetInternal(err)
		}

		return c.JSON(http.StatusOK, value)
	}
}
