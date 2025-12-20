package errors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	err := &Error{Msg: "test error message"}
	assert.Equal(t, "test error message", err.Error())
}

func TestWebError_WithErrorType(t *testing.T) {
	var err error = &Error{Msg: "web error"}
	err = errors.Wrap(err, "wrapped error")
	err = errors.Wrap(err, "another wrapped error")

	assert.EqualValues(t, "web error", WebError(err).Error())
}

func TestWebError_Nil(t *testing.T) {
	result := WebError(nil)
	assert.Nil(t, result)
}

func TestWebError_NonErrorType(t *testing.T) {
	err := fmt.Errorf("regular error")
	result := WebError(err)
	assert.NotNil(t, result)
	assert.Equal(t, "regular error", result.Error())
}

func TestWebError_MultilineError(t *testing.T) {
	err := fmt.Errorf("first line\nsecond line\nthird line")
	result := WebError(err)
	// Should only return first line
	assert.Equal(t, "first line", result.Error())
}

func TestApiError_WithErrorType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testErr := fmt.Errorf("database connection failed")
	err := ApiError(c, http.StatusInternalServerError, testErr)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "database connection failed")
}

func TestApiError_WithStringMessage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ApiError(c, http.StatusBadRequest, "invalid input")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid input")
}

func TestNewErrorResponse_WithMessage(t *testing.T) {
	resp := NewErrorResponse(http.StatusNotFound, "resource not found")
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, "resource not found", resp.Message)
	assert.Nil(t, resp.Internal)
}

func TestNewErrorResponse_NoMessage(t *testing.T) {
	resp := NewErrorResponse(http.StatusOK)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, http.StatusText(http.StatusOK), resp.Message)
}

func TestErrorResponse_Error_NoInternal(t *testing.T) {
	resp := &ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "bad request",
	}
	errStr := resp.Error()
	assert.Contains(t, errStr, "code=400")
	assert.Contains(t, errStr, "message=bad request")
	assert.NotContains(t, errStr, "internal=")
}

func TestErrorResponse_Error_WithInternal(t *testing.T) {
	internalErr := fmt.Errorf("internal error details")
	resp := &ErrorResponse{
		Code:     http.StatusInternalServerError,
		Message:  "server error",
		Internal: internalErr,
	}
	errStr := resp.Error()
	assert.Contains(t, errStr, "code=500")
	assert.Contains(t, errStr, "message=server error")
	assert.Contains(t, errStr, "internal=internal error details")
}

func TestErrorResponse_SetInternal(t *testing.T) {
	resp := &ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "bad request",
	}

	internalErr := fmt.Errorf("validation failed")
	result := resp.SetInternal(internalErr)

	assert.Equal(t, resp, result) // Should return same instance
	assert.Equal(t, internalErr, resp.Internal)
}

func TestErrorResponse_WithInternal(t *testing.T) {
	original := &ErrorResponse{
		Code:    http.StatusBadRequest,
		Message: "bad request",
	}

	internalErr := fmt.Errorf("validation failed")
	clone := original.WithInternal(internalErr)

	// Should be different instances
	assert.NotEqual(t, original, clone)
	assert.Nil(t, original.Internal)
	assert.Equal(t, internalErr, clone.Internal)
	assert.Equal(t, original.Code, clone.Code)
	assert.Equal(t, original.Message, clone.Message)
}

func TestErrorResponse_Unwrap(t *testing.T) {
	internalErr := fmt.Errorf("wrapped error")
	resp := &ErrorResponse{
		Code:     http.StatusInternalServerError,
		Message:  "server error",
		Internal: internalErr,
	}

	unwrapped := resp.Unwrap()
	assert.Equal(t, internalErr, unwrapped)
}

func TestErrorResponse_Unwrap_Nil(t *testing.T) {
	resp := &ErrorResponse{
		Code:    http.StatusOK,
		Message: "ok",
	}

	unwrapped := resp.Unwrap()
	assert.Nil(t, unwrapped)
}
