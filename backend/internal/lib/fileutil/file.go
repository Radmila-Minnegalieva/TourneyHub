// Package fileutil проверяет содержимое и имена загружаемых файлов.
package fileutil

import (
	"bytes"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func Validate(data []byte, allowPDF bool, maxBytes int) (string, error) {
	invalid := func(message string) (string, error) { return "", apperror.New(apperror.Validation, message) }
	if len(data) == 0 || len(data) > maxBytes {
		return invalid("file must contain 1..5242880 bytes")
	}
	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/png", "image/jpeg":
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return invalid("invalid image")
		}
		if cfg.Width > 10000 || cfg.Height > 10000 {
			return invalid("image dimensions exceed 10000 pixels")
		}
	case "application/pdf":
		if !allowPDF {
			return invalid("PDF is only supported for match protocols")
		}
	default:
		return invalid("only PDF, PNG and JPEG files are supported")
	}
	return contentType, nil
}
func CleanName(filename string) (string, error) {
	name := filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	if name == "." || name == "" || utf8.RuneCountInString(name) > 255 || strings.ContainsAny(name, "\r\n\x00") {
		return "", apperror.New(apperror.Validation, "invalid file name")
	}
	return name, nil
}
