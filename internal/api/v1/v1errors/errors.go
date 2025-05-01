package v1errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/logos"
	"github.com/labstack/echo/v4"
)

// Error - in an error stack, this is the error that will be returned to the user.
type Error struct {
	Msg string
}

// Error - implement the error interface.
func (e Error) Error() string {
	return e.Msg
}

// WebError - try to find Error in the error stack.
func WebError(err error) error {
	if err == nil {
		return nil
	}

	var webErr *Error
	if errors.As(err, &webErr) {
		return webErr
	}

	// If we can't find a Error, return the top-most error string.
	msg := strings.Split(err.Error(), "\n")[0]

	return &Error{Msg: msg}
}

func ApiError(c echo.Context, code int, errmsg any) *ErrorResponse {
	err, _ := errmsg.(error)

	log := logos.With("request", c.Request().RequestURI).With("status", code)
	if err != nil {
		log.WithError(err).Error(errmsg)
		return NewErrorResponse(code, WebError(err))
	}

	log.Error(errmsg)
	return NewErrorResponse(code, errmsg)
}

// ErrorResponse represents a standard error response.
//
// swagger:model ErrorResponse
type ErrorResponse struct {
	// A human-readable message or structured error detail
	Message any `json:"message"`

	// Internal is not exposed in Swagger
	Internal error `json:"-"`

	// Code is not exposed in Swagger
	Code int `json:"-"`
} // @name ErrorResponse

func NewErrorResponse(code int, message ...interface{}) *ErrorResponse {
	he := &ErrorResponse{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

// Error makes it compatible with `error` interface.
func (he *ErrorResponse) Error() string {
	if he.Internal == nil {
		return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
	}
	return fmt.Sprintf("code=%d, message=%v, internal=%v", he.Code, he.Message, he.Internal)
}

// SetInternal sets error to HTTPError.Internal
func (he *ErrorResponse) SetInternal(err error) *ErrorResponse {
	he.Internal = err
	return he
}

// WithInternal returns clone of HTTPError with err set to HTTPError.Internal field
func (he *ErrorResponse) WithInternal(err error) *ErrorResponse {
	return &ErrorResponse{
		Code:     he.Code,
		Message:  he.Message,
		Internal: err,
	}
}

// Unwrap satisfies the Go 1.13 error wrapper interface.
func (he *ErrorResponse) Unwrap() error {
	return he.Internal
}
