package fileutil

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func testPNG(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestUploadedFileContentValidation(t *testing.T) {
	for _, tc := range []struct {
		name        string
		purpose     string
		contentType string
		data        []byte
		wantError   bool
	}{
		{name: "PNG avatar", purpose: "avatar", contentType: "image/png", data: testPNG(t)},
		{name: "PDF protocol", purpose: "match_protocol", contentType: "application/pdf", data: []byte("%PDF-1.4\n%%EOF\n")},
		{name: "PDF avatar", purpose: "avatar", data: []byte("%PDF-1.4\n%%EOF\n"), wantError: true},
		{name: "HTML disguised as image", purpose: "avatar", data: []byte("<html>active content</html>"), wantError: true},
		{name: "truncated PNG", purpose: "avatar", data: []byte("\x89PNG\r\n\x1a\ninvalid"), wantError: true},
		{name: "empty file", purpose: "match_protocol", wantError: true},
		{name: "oversized file", purpose: "match_protocol", data: make([]byte, 5*1024*1024+1), wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			contentType, err := Validate(tc.data, tc.purpose == "match_protocol", 5*1024*1024)
			if (err != nil) != tc.wantError || (!tc.wantError && contentType != tc.contentType) {
				t.Fatalf("content type=%q error=%v", contentType, err)
			}
		})
	}
}
