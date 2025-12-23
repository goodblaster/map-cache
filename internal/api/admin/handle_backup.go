package admin

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

type adminBackupRequest struct {
	CacheName string `json:"cache,required"`
	Filename  string `json:"filename,required"`
}

func (req adminBackupRequest) Validate() error {
	if req.Filename == "" {
		return errors.New("filename is required")
	}

	// Security: Prevent path traversal attacks
	// Only allow simple filenames without directory traversal
	cleanPath := filepath.Clean(req.Filename)
	if strings.Contains(cleanPath, "..") {
		return errors.New("filename cannot contain '..'")
	}

	// Prevent absolute paths that could write to system directories
	if filepath.IsAbs(cleanPath) {
		return errors.New("filename must be a relative path")
	}

	return nil
}

func handleBackup(c echo.Context) error {
	ctx := c.Request().Context()
	var input adminBackupRequest
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json payload").SetInternal(err)
	}

	if err := input.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "validation failed").SetInternal(err)
	}

	err := caches.Backup(ctx, input.CacheName, input.Filename)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "backup failed").SetInternal(err)
	}

	return c.NoContent(http.StatusOK)
}
