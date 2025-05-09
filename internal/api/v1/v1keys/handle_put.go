package v1keys

import (
	"net/http"
	"time"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/logos"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/labstack/echo/v4"
)

// handlePutRequest represents the request body for replacing a single cache value.
//
// swagger:model handlePutRequest
type handlePutRequest struct {
	// New value to store for the key
	// required: true
	Value any `json:"value"`
} // @name HandlePutRequest

func (req handlePutRequest) Validate() error {
	return nil
}

// handlePut replaces the value of a single key in the cache.
//
// @Summary Replace a single value
// @Description Replaces the value of a key in the cache
// @Tags keys
// @Accept json
// @Produce json
// @Param key path string true "Key to update"
// @Param body body handlePutRequest true "New value for the key"
// @Success 200 {string} string "Value replaced successfully"
// @Failure 400 {object} v1errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} v1errors.ErrorResponse "Internal server error"
// @Router /keys/{key} [put]
func handlePut() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		key := c.Param("key")

		var req handlePutRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := req.Validate(); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request body"))
		}

		if err := cache.Replace(c.Request().Context(), key, req.Value); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not replace contents"))
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}

// replaceBatchRequest represents the request body for batch replacing values.
//
// swagger:model replaceBatchRequest
type replaceBatchRequest struct {
	// Map of keys to their new values
	// required: true
	Entries map[string]any `json:"entries"`

	// Map of keys to their new TTLs (in seconds)
	// required: false
	TTL map[string]*int64 `json:"ttl"`
} // @name ReplaceBatchRequest

func (req replaceBatchRequest) Validate() error {
	for key := range req.Entries {
		if key == "" {
			return errors.New("key cannot be empty")
		}
	}
	return nil
}

// handleReplaceBatch replaces multiple key-value pairs in the cache.
//
// @Summary Replace multiple values
// @Description Replaces multiple entries in the cache
// @Tags keys
// @Accept json
// @Produce json
// @Param body body replaceBatchRequest true "Map of key-value pairs to replace"
// @Success 200 {string} string "Values replaced successfully"
// @Failure 400 {object} v1errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} v1errors.ErrorResponse "Internal server error"
// @Router /keys [put]
func handleReplaceBatch() echo.HandlerFunc {
	return func(c echo.Context) error {
		cache := Cache(c)
		var req replaceBatchRequest
		if err := c.Bind(&req); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, "invalid json payload")
		}

		if err := req.Validate(); err != nil {
			return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request body"))
		}

		if err := cache.ReplaceBatch(c.Request().Context(), req.Entries); err != nil {
			return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "could not replace contents"))
		}

		// TTLs
		for key, ttl := range req.TTL {
			if ttl == nil {
				if err := cache.CancelKeyTTL(c.Request().Context(), key); err != nil {
					logos.WithError(err).Warnf("could not cancel cache expiration for key %q", key)
				}
				continue
			}
			if err := cache.SetKeyTTL(c.Request().Context(), key, time.Second*(time.Duration(*ttl))); err != nil {
				logos.WithError(err).Warnf("could not set cache expiration for key %q", key)
			}
		}

		// Triggers?
		//

		return c.NoContent(http.StatusOK)
	}
}
