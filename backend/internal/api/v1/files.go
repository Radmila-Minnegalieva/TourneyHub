package api

import (
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"mime"
)

func (a *API) UploadFile(c echo.Context) error {
	if err := c.Request().ParseMultipartForm(6 * 1024 * 1024); err != nil {
		return echo.NewHTTPError(400, "invalid multipart body").SetInternal(err)
	}
	defer c.Request().MultipartForm.RemoveAll()
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(400, "file is required")
	}
	stream, err := file.Open()
	if err != nil {
		return err
	}
	defer stream.Close()
	data, err := io.ReadAll(io.LimitReader(stream, service.MaxFileBytes+1))
	if err != nil {
		return err
	}
	result, err := a.resources.UploadFile(c.Request().Context(), model.FileUpload{Filename: file.Filename, Purpose: model.FilePurpose(c.FormValue("purpose")), Data: data})
	if err != nil {
		return err
	}
	return c.JSON(201, fromModelFile(result))
}
func (a *API) DownloadFile(c echo.Context, id uuid.UUID) error {
	result, err := a.resources.DownloadFile(c.Request().Context(), id)
	if err != nil {
		return err
	}
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	c.Response().Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": result.File.Filename}))
	c.Response().Header().Set("Cache-Control", "private, max-age=0")
	return c.Blob(200, string(result.File.ContentType), result.Data)
}
