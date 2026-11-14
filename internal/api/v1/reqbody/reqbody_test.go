package reqbody

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// trackingBody records whether the original body was fully consumed and closed.
type trackingBody struct {
	io.Reader
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

func TestBuffer_ReadsBodyUpFront(t *testing.T) {
	original := &trackingBody{Reader: strings.NewReader(`{"entries":{"a":1}}`)}
	req := httptest.NewRequest(http.MethodPost, "/", original)
	c := echo.New().NewContext(req, httptest.NewRecorder())

	require.NoError(t, Buffer(c))
	assert.True(t, original.closed, "the network body should be drained and closed before the lock is taken")

	// Handlers still see the full body.
	data, err := io.ReadAll(c.Request().Body)
	require.NoError(t, err)
	assert.Equal(t, `{"entries":{"a":1}}`, string(data))
}

func TestBuffer_NoBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := echo.New().NewContext(req, httptest.NewRecorder())
	assert.NoError(t, Buffer(c))
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestBuffer_ReadError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", failingReader{})
	c := echo.New().NewContext(req, httptest.NewRecorder())
	assert.ErrorIs(t, Buffer(c), io.ErrUnexpectedEOF)
}
