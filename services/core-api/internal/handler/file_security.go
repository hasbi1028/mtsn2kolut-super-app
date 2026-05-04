package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

type validatedUpload struct {
	Data     []byte
	MimeType string
	Ext      string
}

var allowedUploadTypes = map[string]map[string]bool{
	".jpg":  {"image/jpeg": true},
	".jpeg": {"image/jpeg": true},
	".png":  {"image/png": true},
	".gif":  {"image/gif": true},
	".webp": {"image/webp": true},
	".pdf":  {"application/pdf": true},
	".mp3":  {"audio/mpeg": true},
	".wav":  {"audio/wave": true, "audio/wav": true, "audio/x-wav": true},
	".m4a":  {"audio/mp4": true, "video/mp4": true},
}

func validateUploadedFile(filename string, r io.Reader, maxBytes int64, imageOnly bool) (validatedUpload, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return validatedUpload{}, err
	}
	if int64(len(data)) > maxBytes {
		return validatedUpload{}, fmt.Errorf("ukuran file melebihi batas")
	}
	if len(data) == 0 {
		return validatedUpload{}, fmt.Errorf("file kosong")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := http.DetectContentType(data)
	allowedForExt := allowedUploadTypes[ext]
	if len(allowedForExt) == 0 || !allowedForExt[mimeType] {
		return validatedUpload{}, fmt.Errorf("jenis file tidak didukung")
	}
	if imageOnly && !strings.HasPrefix(mimeType, "image/") {
		return validatedUpload{}, fmt.Errorf("hanya file gambar yang diperbolehkan")
	}
	if ext == ".svg" || bytes.Contains(bytes.ToLower(data[:min(len(data), 512)]), []byte("<script")) {
		return validatedUpload{}, fmt.Errorf("file scriptable tidak diperbolehkan")
	}
	return validatedUpload{Data: data, MimeType: mimeType, Ext: ext}, nil
}

func secureFileResponseHeaders(w http.ResponseWriter, mimeType, filename string) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}
	disposition := "inline"
	if archiveLike(filename, mimeType) {
		disposition = "attachment"
	}
	w.Header().Set("Content-Disposition", disposition+`; filename="`+contentDispositionFilename(filename)+`"`)
}

func contentDispositionFilename(filename string) string {
	name := filepath.Base(filename)
	name = strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == 0 {
			return -1
		}
		return r
	}, name)
	name = strings.ReplaceAll(name, `\`, `\\`)
	name = strings.ReplaceAll(name, `"`, `\"`)
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "download"
	}
	return name
}

func archiveLike(filename, mimeType string) bool {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".apk", ".exe", ".bin":
		return true
	}
	return strings.HasPrefix(mimeType, "application/zip") || strings.HasPrefix(mimeType, "application/x-")
}
