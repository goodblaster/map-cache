package admin

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type adminRestoreRequest struct {
	CacheName string `json:"cache,required"`
	Filename  string `json:"filename,required"`
}

func (req adminRestoreRequest) Validate() error {
	if req.Filename == "" {
		return errors.New("filename is required")
	}

	// Security: Prevent path traversal attacks
	// Only allow simple filenames without directory traversal
	cleanPath := filepath.Clean(req.Filename)
	if strings.Contains(cleanPath, "..") {
		return errors.New("filename cannot contain '..'")
	}

	// Prevent absolute paths that could read from system directories
	if filepath.IsAbs(cleanPath) {
		return errors.New("filename must be a relative path")
	}

	// Make sure the file exists.
	f, err := os.Stat(req.Filename)
	if err != nil {
		return errors.New("could not stat file")
	}

	if f.Size() == 0 {
		return errors.New("file is empty")
	}

	return nil
}

func handleRestore(c echo.Context) error {
	ctx := c.Request().Context()
	var input adminRestoreRequest
	if err := c.Bind(&input); err != nil {
		return v1errors.ApiError(c, http.StatusBadRequest, err)
	}

	if err := input.Validate(); err != nil {
		return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "input failed validation"))
	}

	err := caches.Restore(ctx, "", input.Filename)
	if err != nil {
		return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "restore failed"))
	}

	return c.NoContent(http.StatusOK)
}
