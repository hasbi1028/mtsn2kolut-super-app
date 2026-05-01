package handler

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"mtsn2kolut-super-app/backend/internal/api"
)

var websiteMediaSafeFilename = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

type WebsiteMedia struct {
	storageDir string
}

func NewWebsiteMedia(storageDir string) *WebsiteMedia {
	return &WebsiteMedia{storageDir: storageDir}
}

func (h *WebsiteMedia) Upload(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		api.BadRequest(w, "multipart form tidak valid (maks 8 MB)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "file wajib diisi")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		api.BadRequest(w, "hanya file gambar yang diperbolehkan")
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, 8<<20))
	if err != nil {
		api.Internal(w, err)
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum(data))
	ext := strings.ToLower(filepath.Ext(header.Filename))
	baseName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	baseName = websiteMediaSafeFilename.ReplaceAllString(baseName, "_")
	if len(baseName) > 40 {
		baseName = baseName[:40]
	}
	storedName := fmt.Sprintf("%s_%s%s", hash[:12], baseName, ext)

	if err := os.MkdirAll(h.storageDir, 0755); err != nil {
		api.Internal(w, err)
		return
	}
	destPath := filepath.Join(h.storageDir, storedName)
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		api.Internal(w, err)
		return
	}

	api.Created(w, map[string]string{
		"url":  "/api/website/media/" + storedName,
		"name": storedName,
	})
}

func (h *WebsiteMedia) File(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		api.NotFound(w)
		return
	}
	path := filepath.Join(h.storageDir, filename)
	w.Header().Del("Content-Type")
	http.ServeFile(w, r, path)
}
