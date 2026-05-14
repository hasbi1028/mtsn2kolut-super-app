package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

const brandingUploadMaxBytes int64 = 2 << 20

var brandingPurposeRules = map[string]struct {
	MinWidth  int
	MinHeight int
	MaxWidth  int
	MaxHeight int
	Square    bool
}{
	"logo":             {MinWidth: 128, MinHeight: 48, MaxWidth: 2400, MaxHeight: 1200},
	"mark":             {MinWidth: 64, MinHeight: 64, MaxWidth: 1024, MaxHeight: 1024, Square: true},
	"favicon":          {MinWidth: 32, MinHeight: 32, MaxWidth: 512, MaxHeight: 512, Square: true},
	"apple_touch_icon": {MinWidth: 180, MinHeight: 180, MaxWidth: 1024, MaxHeight: 1024, Square: true},
	"pwa_icon_192":     {MinWidth: 192, MinHeight: 192, MaxWidth: 1024, MaxHeight: 1024, Square: true},
	"pwa_icon_512":     {MinWidth: 512, MinHeight: 512, MaxWidth: 2048, MaxHeight: 2048, Square: true},
	"formal_logo":      {MinWidth: 128, MinHeight: 128, MaxWidth: 2400, MaxHeight: 2400},
}

type Branding struct {
	svc        brandingService
	storageDir string
}

type brandingService interface {
	Branding(ctx context.Context) (service.BrandingSettings, error)
	UpdateBranding(ctx context.Context, settings service.BrandingSettings) (service.BrandingSettings, error)
	UpdateBrandingAsset(ctx context.Context, purpose, url, hash string) (service.BrandingSettings, error)
	ResetBrandingAsset(ctx context.Context, purpose string) (service.BrandingSettings, error)
}

func NewBranding(svc brandingService, storageDir string) *Branding {
	return &Branding{svc: svc, storageDir: storageDir}
}

func (h *Branding) Public(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.Branding(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	api.OK(w, settings)
}

func (h *Branding) Get(w http.ResponseWriter, r *http.Request) { h.Public(w, r) }

func (h *Branding) Update(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body service.BrandingSettings
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data branding tidak valid")
		return
	}
	settings, err := h.svc.UpdateBranding(r.Context(), body)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, settings)
}

func (h *Branding) UploadAsset(w http.ResponseWriter, r *http.Request) {
	purpose := chi.URLParam(r, "purpose")
	rule, ok := brandingPurposeRules[purpose]
	if !ok {
		api.BadRequest(w, "Jenis aset branding tidak didukung")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, brandingUploadMaxBytes+1024)
	if err := r.ParseMultipartForm(brandingUploadMaxBytes); err != nil {
		api.BadRequest(w, "multipart form tidak valid (maks 2 MB)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "file wajib diisi")
		return
	}
	defer file.Close()
	validated, err := validateUploadedFile(header.Filename, file, brandingUploadMaxBytes, true)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	if validated.Ext != ".png" && validated.Ext != ".jpg" && validated.Ext != ".jpeg" {
		api.BadRequest(w, "branding hanya menerima PNG atau JPEG agar dimensi dapat divalidasi")
		return
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(validated.Data))
	if err != nil {
		api.BadRequest(w, "file gambar tidak valid")
		return
	}
	if cfg.Width < rule.MinWidth || cfg.Height < rule.MinHeight || cfg.Width > rule.MaxWidth || cfg.Height > rule.MaxHeight {
		api.BadRequest(w, fmt.Sprintf("dimensi %s harus antara %dx%d dan %dx%d piksel", purpose, rule.MinWidth, rule.MinHeight, rule.MaxWidth, rule.MaxHeight))
		return
	}
	if rule.Square && cfg.Width != cfg.Height {
		api.BadRequest(w, "ikon wajib berbentuk persegi")
		return
	}
	if err := os.MkdirAll(h.storageDir, 0755); err != nil {
		api.Internal(w, err)
		return
	}
	sum := sha256.Sum256(validated.Data)
	hash := hex.EncodeToString(sum[:])
	storedName := fmt.Sprintf("%s-%s%s", purpose, hash[:16], normalizedBrandingExt(validated.Ext))
	if err := os.WriteFile(filepath.Join(h.storageDir, storedName), validated.Data, 0644); err != nil {
		api.Internal(w, err)
		return
	}
	settings, err := h.svc.UpdateBrandingAsset(r.Context(), purpose, "/api/branding/file/"+storedName, hash)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, map[string]any{"url": "/api/branding/file/" + storedName, "name": storedName, "width": cfg.Width, "height": cfg.Height, "branding": settings})
}

func (h *Branding) ResetAsset(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.ResetBrandingAsset(r.Context(), chi.URLParam(r, "purpose"))
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, settings)
}

func (h *Branding) Asset(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		api.NotFound(w)
		return
	}
	if !strings.HasPrefix(filename, "logo-") && !strings.HasPrefix(filename, "mark-") && !strings.HasPrefix(filename, "favicon-") && !strings.HasPrefix(filename, "apple_touch_icon-") && !strings.HasPrefix(filename, "pwa_icon_192-") && !strings.HasPrefix(filename, "pwa_icon_512-") && !strings.HasPrefix(filename, "formal_logo-") {
		api.NotFound(w)
		return
	}
	path := filepath.Join(h.storageDir, filename)
	secureFileResponseHeaders(w, mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))), filename)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, path)
}

func normalizedBrandingExt(ext string) string {
	ext = strings.ToLower(ext)
	if ext == ".jpeg" {
		return ".jpg"
	}
	return ext
}
