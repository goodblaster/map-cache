package v1

import (
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

func ApiError(c echo.Context, code int, errmsg any) *echo.HTTPError {
	err, _ := errmsg.(error)

	log := logos.With("request", c.Request().RequestURI).With("status", code)
	if err != nil {
		log.WithError(err).Error(errmsg)
		return echo.NewHTTPError(code, WebError(err))
	}

	log.Error(errmsg)
	return echo.NewHTTPError(code, errmsg)
}
