package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateUploadedFileSecurityRejectsUnsafeInputs(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		body      string
		imageOnly bool
		wantErr   string
	}{
		{name: "extension mime mismatch", filename: "foto.png", body: "%PDF-1.4\nbody", wantErr: "jenis file tidak didukung"},
		{name: "empty file", filename: "foto.png", body: "", wantErr: "file kosong"},
		{name: "svg rejected", filename: "icon.svg", body: `<svg xmlns="http://www.w3.org/2000/svg"></svg>`, wantErr: "jenis file tidak didukung"},
		{name: "polyglot script rejected", filename: "foto.jpg", body: "\xff\xd8\xff\xe0" + strings.Repeat("A", 32) + "<script>alert(1)</script>", wantErr: "file scriptable tidak diperbolehkan"},
		{name: "image only rejects pdf", filename: "dokumen.pdf", body: "%PDF-1.4\nbody", imageOnly: true, wantErr: "hanya file gambar yang diperbolehkan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateUploadedFile(tt.filename, strings.NewReader(tt.body), 1024, tt.imageOnly)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateUploadedFile() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestSecureFileResponseHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	secureFileResponseHeaders(rec, "application/pdf", `../rapor "kelas"\\akhir.pdf`)

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("Content-Type = %q, want application/pdf", got)
	}
	got := rec.Header().Get("Content-Disposition")
	if !strings.HasPrefix(got, "inline; ") {
		t.Fatalf("Content-Disposition = %q, want inline", got)
	}
	if strings.Contains(got, "../") || strings.Contains(got, "\r") || strings.Contains(got, "\n") {
		t.Fatalf("Content-Disposition = %q, contains unsafe filename content", got)
	}
	if !strings.Contains(got, `filename="rapor \"kelas\"\\\\akhir.pdf"`) {
		t.Fatalf("Content-Disposition = %q, want safely quoted basename", got)
	}

	archive := httptest.NewRecorder()
	secureFileResponseHeaders(archive, "application/zip", "bundle.zip")
	if got := archive.Header().Get("Content-Disposition"); !strings.HasPrefix(got, "attachment; ") {
		t.Fatalf("archive Content-Disposition = %q, want attachment", got)
	}
	if got := archive.Result().Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("archive X-Content-Type-Options = %q, want nosniff", got)
	}
}

func TestValidateUploadedFileAcceptsSafePNG(t *testing.T) {
	png := "\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 64)
	got, err := validateUploadedFile("foto.png", strings.NewReader(png), 1024, true)
	if err != nil {
		t.Fatalf("validateUploadedFile(safe png) error = %v", err)
	}
	if got.MimeType != "image/png" || got.Ext != ".png" {
		t.Fatalf("validateUploadedFile(safe png) = mime %q ext %q, want image/png .png", got.MimeType, got.Ext)
	}
}

func TestSecureFileResponseHeadersDoesNotWriteStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	secureFileResponseHeaders(rec, "text/plain", "plain.txt")
	if rec.Code != http.StatusOK {
		t.Fatalf("recorder status = %d, want implicit 200", rec.Code)
	}
}
