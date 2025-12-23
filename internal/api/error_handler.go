package api

import (
	"net/http"
	"strings"

	v1errors "github.com/goodblaster/map-cache/internal/api/v1/errors"
	"github.com/goodblaster/map-cache/internal/log"
	"github.com/labstack/echo/v4"
)

// CustomErrorHandler is a centralized error handler for all Echo errors.
// It handles logging and response formatting in one place, following the principle:
// "Handlers return errors, middleware logs them"
//
// Logging strategy:
//   - All errors (4xx and 5xx): ERROR level for comprehensive error tracking
//   - Full error details logged when SetInternal() is used
//   - Includes request_id from context for correlation
func CustomErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// Get request-scoped logger for correlation
	logger := log.FromContext(c.Request().Context())

	// Default to 500 Internal Server Error
	code := http.StatusInternalServerError
	message := http.StatusText(code)
	var internal error

	// Extract HTTP error details
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if he.Internal != nil {
			internal = he.Internal
		}
		// Use custom message if provided
		if he.Message != nil {
			switch v := he.Message.(type) {
			case string:
				message = v
			case error:
				// For error messages, extract user-friendly message
				// This prevents exposing internal error chains to users
				message = extractUserFriendlyMessage(v)
			default:
				message = http.StatusText(code)
			}
		} else {
			message = http.StatusText(code)
		}
	} else {
		// Plain error - treat as internal server error
		internal = err
		// Don't expose internal error details to user
		message = "Internal server error"
	}

	// Log based on error severity
	logFields := logger.
		With("status", code).
		With("path", c.Request().URL.Path).
		With("method", c.Request().Method)

	if internal != nil {
		logFields = logFields.WithError(internal)
	}

	// Log all errors at ERROR level for comprehensive error tracking
	logFields.Error(message)

	// Send JSON response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, map[string]interface{}{
				"message": message,
			})
		}
		if err != nil {
			logger.WithError(err).Error("failed to send error response")
		}
	}
}

// extractUserFriendlyMessage extracts a user-friendly error message from potentially
// nested/wrapped errors. This prevents exposing internal implementation details.
//
// Strategy:
//  1. Look for v1errors.Error type in the error chain (user-friendly errors)
//  2. If not found, use only the first line of the error message
//  3. This ensures wrapped error chains don't leak internal details
func extractUserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

	// Use WebError to find user-friendly error in chain
	webErr := v1errors.WebError(err)
	if webErr != nil {
		msg := webErr.Error()
		// Only use first line to avoid exposing stack traces
		if lines := strings.Split(msg, "\n"); len(lines) > 0 {
			return strings.TrimSpace(lines[0])
		}
		return msg
	}

	return "An error occurred"
}
