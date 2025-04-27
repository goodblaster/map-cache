package v1

import (
	"strings"

	"github.com/goodblaster/errors"
)

// WebError - in an error stack, this is the error that will be returned to the user.
type WebError struct {
	Msg string
}

// Error - implement the error interface.
func (e WebError) Error() string {
	return e.Msg
}

// WebError - try to find a WebError in the error stack.
func (V1) WebError(err error) error {
	if err == nil {
		return nil
	}

	var webErr WebError
	if errors.As(err, &webErr) {
		return &webErr
	}

	// If we can't find a WebError, return the top-most error string.
	msg := strings.Split(err.Error(), "\n")[0]

	return &WebError{Msg: msg}
}
