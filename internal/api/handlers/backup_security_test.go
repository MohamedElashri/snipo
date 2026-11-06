package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type spacesReader struct{}

func (spacesReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	return len(p), nil
}

func TestImportUploadLimit(t *testing.T) {
	for _, size := range []int64{(50 << 20) + 1, 52 << 20} {
		// A streaming body without Content-Length proves this is an actual read bound.
		body := io.MultiReader(strings.NewReader("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"backup.json\"\r\n\r\n"), io.LimitReader(spacesReader{}, size), strings.NewReader("\r\n--boundary--\r\n"))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/backup/import", body)
		req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
		w := httptest.NewRecorder()
		NewBackupHandler(nil, nil).Import(w, req)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("size %d: expected 413, got %d: %s", size, w.Code, w.Body.String())
		}
	}
}
