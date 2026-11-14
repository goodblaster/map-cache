// Package reqbody provides request body helpers shared by the v1 API packages.
package reqbody

import (
	"bytes"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Buffer reads the whole request body into memory and replaces it with an
// in-memory reader.
//
// The cache middleware calls this before acquiring the cache lock. Handlers
// decode the body while holding that lock, so without buffering a client that
// uploads slowly (or stalls mid-body) would block every other request to the
// same cache for as long as its upload takes.
func Buffer(c echo.Context) error {
	req := c.Request()
	if req.Body == nil || req.Body == http.NoBody {
		return nil
	}

	data, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return err
	}

	req.Body = io.NopCloser(bytes.NewReader(data))
	return nil
}
