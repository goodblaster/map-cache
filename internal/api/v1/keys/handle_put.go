package keys

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/log"
	"github.com/labstack/echo/v4"
)

// handlePutRequest represents the request body for replacing a single cache value.
type handlePutRequest struct {
	// New value to store for the key
	Value any `json:"value"`
}

func (req handlePutRequest) Validate() error {
	return nil
}

// handlePut replaces the value of a single key in the cache.
func handlePut() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		var req handlePutRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid json payload").SetInternal(err)
		}

		if err := req.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "validation failed").SetInternal(err)
		}

		if err := cache.Replace(c.Request().Context(), key, req.Value); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "could not replace contents").SetInternal(err)
		}

		return c.NoContent(http.StatusOK)
	}
}

// replaceBatchRequest represents the request body for batch replacing values.
type replaceBatchRequest struct {
	// Map of keys to their new values
	Entries map[string]any    `json:"entries,required"`
	TTL     map[string]*int64 `json:"ttl"` // milliseconds
}

func (req replaceBatchRequest) Validate() error {
	for key := range req.Entries {
		if key == "" {
			return errors.New("key cannot be empty")
		}
	}
	return nil
}

// handleReplaceBatch replaces multiple key-value pairs in the cache.
func handleReplaceBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req replaceBatchRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid json payload").SetInternal(err)
		}

		if err := req.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "validation failed").SetInternal(err)
		}

		if err := cache.ReplaceBatch(c.Request().Context(), req.Entries); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "could not replace contents").SetInternal(err)
		}

		// TTLs
		for key, ttl := range req.TTL {
			if ttl == nil {
				if err := cache.CancelKeyTTL(c.Request().Context(), key); err != nil {
					log.FromContext(c.Request().Context()).WithError(err).With("key", key).Warn("could not cancel cache expiration")
				}
				continue
			}
			if err := cache.SetKeyTTL(c.Request().Context(), key, *ttl); err != nil {
				log.FromContext(c.Request().Context()).WithError(err).With("key", key).With("ttl_ms", *ttl).Warn("could not set cache expiration")
			}
		}

		return c.NoContent(http.StatusOK)
	}
}
