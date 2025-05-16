package admin

import (
	"net/http"

	"github.com/goodblaster/errors"
	"github.com/goodblaster/map-cache/internal/api/v1/v1errors"
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
	return nil
}

func handleBackup(c echo.Context) error {
	ctx := c.Request().Context()
	var input adminBackupRequest
	if err := c.Bind(&input); err != nil {
		return v1errors.ApiError(c, http.StatusBadRequest, err)
	}

	if err := input.Validate(); err != nil {
		return v1errors.ApiError(c, http.StatusBadRequest, errors.Wrap(err, "input failed validation"))
	}

	err := caches.Backup(ctx, input.CacheName, input.Filename)
	if err != nil {
		return v1errors.ApiError(c, http.StatusInternalServerError, errors.Wrap(err, "backup failed"))
	}

	return c.NoContent(http.StatusOK)
}
