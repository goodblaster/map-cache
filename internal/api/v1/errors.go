package v1

import (
	"strings"

	"github.com/goodblaster/errors"
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
