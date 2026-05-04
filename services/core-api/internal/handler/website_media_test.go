package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func websiteMediaAdminMultipartRequest(t *testing.T, target string, body *bytes.Buffer, contentType string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, target, body)
	req.Header.Set("Content-Type", contentType)
	return withClaims(req, jwt.MapClaims{
		"roles": []any{"admin"},
		"uid":   handlerTestUUID(1).String(),
	})
}

func websiteMediaBody(t *testing.T, filename, contentType string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("CreatePart() error = %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("multipart write error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart close error = %v", err)
	}
	return &body, writer.FormDataContentType()
}

func TestWebsiteMediaUploadValidationAndSuccess(t *testing.T) {
	h := NewWebsiteMedia(t.TempDir())
	if h == nil {
		t.Fatal("NewWebsiteMedia() = nil")
	}

	rec := httptest.NewRecorder()
	h.Upload(rec, adminRequest(http.MethodPost, "/api/website/media", "not multipart"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Upload(invalid multipart) status = %d, want 400", rec.Code)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart close error = %v", err)
	}
	rec = httptest.NewRecorder()
	h.Upload(rec, websiteMediaAdminMultipartRequest(t, "/api/website/media", &body, writer.FormDataContentType()))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Upload(missing file) status = %d, want 400", rec.Code)
	}

	textBody, textContentType := websiteMediaBody(t, "catatan.txt", "text/plain", []byte("bukan gambar"))
	rec = httptest.NewRecorder()
	h.Upload(rec, websiteMediaAdminMultipartRequest(t, "/api/website/media", textBody, textContentType))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Upload(non image) status = %d, want 400", rec.Code)
	}

	imageBody, imageContentType := websiteMediaBody(t, "Foto Rapat Komite Dengan Nama Sangat Panjang Sekali!!.PNG", "image/png", []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\b\x02\x00\x00\x00"))
	rec = httptest.NewRecorder()
	h.Upload(rec, websiteMediaAdminMultipartRequest(t, "/api/website/media", imageBody, imageContentType))
	if rec.Code != http.StatusCreated {
		t.Fatalf("Upload(image) status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data struct {
			URL  string `json:"url"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Upload(image) json error = %v", err)
	}
	if !strings.HasPrefix(payload.Data.URL, "/api/website/media/") || !strings.HasSuffix(payload.Data.Name, ".png") {
		t.Fatalf("Upload(image) payload = %+v, want media URL and png name", payload.Data)
	}
	if strings.ContainsAny(payload.Data.Name, " !") {
		t.Fatalf("Upload(image) stored name = %q, want sanitized name", payload.Data.Name)
	}
	if _, err := os.Stat(filepath.Join(h.storageDir, payload.Data.Name)); err != nil {
		t.Fatalf("uploaded file stat error = %v", err)
	}
}

func TestWebsiteMediaUploadRejectsOversizedMultipart(t *testing.T) {
	h := NewWebsiteMedia(t.TempDir())
	body, contentType := websiteMediaBody(t, "foto.png", "image/png", bytes.Repeat([]byte("x"), (8<<20)+1))
	rec := httptest.NewRecorder()

	h.Upload(rec, websiteMediaAdminMultipartRequest(t, "/api/website/media", body, contentType))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Upload(oversized) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestWebsiteMediaUploadStorageError(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "storage-file")
	if err := os.WriteFile(blocker, []byte("not a dir"), 0o600); err != nil {
		t.Fatalf("setup storage blocker error = %v", err)
	}
	h := &WebsiteMedia{storageDir: filepath.Join(blocker, "child")}
	body, contentType := websiteMediaBody(t, "foto.png", "image/png", []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\b\x02\x00\x00\x00"))

	rec := httptest.NewRecorder()
	h.Upload(rec, websiteMediaAdminMultipartRequest(t, "/api/website/media", body, contentType))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Upload(storage error) status = %d, want 500", rec.Code)
	}
}

func TestWebsiteMediaFileServesSafeFiles(t *testing.T) {
	dir := t.TempDir()
	filename := "gambar.png"
	if err := os.WriteFile(filepath.Join(dir, filename), []byte("png-data"), 0o600); err != nil {
		t.Fatalf("setup media file error = %v", err)
	}
	h := &WebsiteMedia{storageDir: dir}

	rec := httptest.NewRecorder()
	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/website/media/"+filename, nil), "filename", filename)
	h.File(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "png-data" {
		t.Fatalf("File(success) status/body = %d/%q, want 200/png-data", rec.Code, rec.Body.String())
	}

	for _, bad := range []string{"../secret.png", "folder/file.png"} {
		rec = httptest.NewRecorder()
		req = withRouteParam(httptest.NewRequest(http.MethodGet, "/api/website/media/"+bad, nil), "filename", bad)
		h.File(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("File(%q) status = %d, want 404", bad, rec.Code)
		}
	}
}
