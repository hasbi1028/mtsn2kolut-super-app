package handler

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeBrandingService struct {
	settings  service.BrandingSettings
	err       error
	updateErr error
	assetErr  error
	resetErr  error

	updateArg    service.BrandingSettings
	assetPurpose string
	assetURL     string
	assetHash    string
	resetPurpose string
}

func (f *fakeBrandingService) Branding(context.Context) (service.BrandingSettings, error) {
	if f.err != nil {
		return service.BrandingSettings{}, f.err
	}
	if f.settings.AppName == "" {
		f.settings.AppName = "MTsN 2 Kolut"
	}
	return f.settings, nil
}

func (f *fakeBrandingService) UpdateBranding(_ context.Context, settings service.BrandingSettings) (service.BrandingSettings, error) {
	f.updateArg = settings
	if f.updateErr != nil {
		return service.BrandingSettings{}, f.updateErr
	}
	settings.Version = "updated"
	return settings, nil
}

func (f *fakeBrandingService) UpdateBrandingAsset(_ context.Context, purpose, url, hash string) (service.BrandingSettings, error) {
	f.assetPurpose = purpose
	f.assetURL = url
	f.assetHash = hash
	if f.assetErr != nil {
		return service.BrandingSettings{}, f.assetErr
	}
	return service.BrandingSettings{AppName: "MTsN 2 Kolut", LogoURL: url, Version: hash[:8]}, nil
}

func (f *fakeBrandingService) ResetBrandingAsset(_ context.Context, purpose string) (service.BrandingSettings, error) {
	f.resetPurpose = purpose
	if f.resetErr != nil {
		return service.BrandingSettings{}, f.resetErr
	}
	return service.BrandingSettings{AppName: "MTsN 2 Kolut", Version: "reset"}, nil
}

func TestBrandingGetAliasesPublic(t *testing.T) {
	fake := &fakeBrandingService{settings: service.BrandingSettings{AppName: "Alias Portal"}}
	h := NewBranding(fake, "")
	if h == nil || h.storageDir != "" {
		t.Fatalf("NewBranding() = %+v, want handler preserving storage dir", h)
	}

	rec := httptest.NewRecorder()
	h.Get(rec, httptest.NewRequest(http.MethodGet, "/api/branding", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Alias Portal") {
		t.Fatalf("Get() status/body = %d/%s, want 200 with settings", rec.Code, rec.Body.String())
	}
}

func TestBrandingPublicUpdateResetAndAsset(t *testing.T) {
	fake := &fakeBrandingService{settings: service.BrandingSettings{AppName: "Madrasah", PrimaryColor: "#155e75"}}
	h := NewBranding(fake, t.TempDir())

	rec := httptest.NewRecorder()
	h.Public(rec, httptest.NewRequest(http.MethodGet, "/api/branding/public", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Madrasah") {
		t.Fatalf("Public() status/body = %d/%s, want 200 with app name", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Cache-Control"); !strings.Contains(got, "max-age=60") {
		t.Fatalf("Public() Cache-Control = %q, want public cache header", got)
	}

	rec = httptest.NewRecorder()
	h.Update(rec, adminRequest(http.MethodPut, "/api/branding", `{"app_name":"Portal","primary_color":"#0f766e"}`))
	if rec.Code != http.StatusOK || fake.updateArg.AppName != "Portal" || fake.updateArg.PrimaryColor != "#0f766e" {
		t.Fatalf("Update() status/arg/body = %d/%+v/%s, want mapped update", rec.Code, fake.updateArg, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ResetAsset(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/branding/assets/logo", ""), "purpose", "logo"))
	if rec.Code != http.StatusOK || fake.resetPurpose != "logo" || !strings.Contains(rec.Body.String(), "reset") {
		t.Fatalf("ResetAsset() status/purpose/body = %d/%q/%s, want 200 logo reset", rec.Code, fake.resetPurpose, rec.Body.String())
	}

	assetDir := t.TempDir()
	assetName := "logo-test.png"
	if err := os.WriteFile(filepath.Join(assetDir, assetName), validPNG(t, 128, 48), 0644); err != nil {
		t.Fatalf("write asset fixture: %v", err)
	}
	h.storageDir = assetDir
	rec = httptest.NewRecorder()
	h.Asset(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/api/branding/file/"+assetName, nil), "filename", assetName))
	if rec.Code != http.StatusOK || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("Asset() status/nosniff = %d/%q, want 200/nosniff", rec.Code, rec.Header().Get("X-Content-Type-Options"))
	}
}

func TestBrandingUploadAssetValidationAndSuccess(t *testing.T) {
	fake := &fakeBrandingService{}
	dir := t.TempDir()
	h := NewBranding(fake, dir)

	rec := httptest.NewRecorder()
	h.UploadAsset(rec, withRouteParam(adminRequest(http.MethodPost, "/api/branding/assets/unknown", ""), "purpose", "unknown"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadAsset(unknown purpose) status = %d, want 400", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.UploadAsset(rec, withRouteParam(multipartBrandingRequest(t, "mark", "mark.png", validPNG(t, 128, 64)), "purpose", "mark"))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "persegi") {
		t.Fatalf("UploadAsset(non-square mark) status/body = %d/%s, want square validation", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UploadAsset(rec, withRouteParam(multipartBrandingRequest(t, "logo", "logo.png", validPNG(t, 128, 48)), "purpose", "logo"))
	if rec.Code != http.StatusCreated || fake.assetPurpose != "logo" || fake.assetURL == "" || fake.assetHash == "" {
		t.Fatalf("UploadAsset(valid) status/asset = %d/%q/%q/%q; body=%s", rec.Code, fake.assetPurpose, fake.assetURL, fake.assetHash, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, strings.TrimPrefix(fake.assetURL, "/api/branding/file/"))); err != nil {
		t.Fatalf("uploaded asset was not stored: %v", err)
	}
}

func TestBrandingErrorPaths(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*Branding, http.ResponseWriter, *http.Request)
		svc  *fakeBrandingService
		req  *http.Request
		want int
	}{
		{name: "public internal", fn: (*Branding).Public, svc: &fakeBrandingService{err: errors.New("db down")}, req: httptest.NewRequest(http.MethodGet, "/api/branding/public", nil), want: http.StatusInternalServerError},
		{name: "update invalid json", fn: (*Branding).Update, svc: &fakeBrandingService{}, req: adminRequest(http.MethodPut, "/api/branding", `{`), want: http.StatusBadRequest},
		{name: "update service validation", fn: (*Branding).Update, svc: &fakeBrandingService{updateErr: errors.New("warna tidak valid")}, req: adminRequest(http.MethodPut, "/api/branding", `{"app_name":"Portal"}`), want: http.StatusBadRequest},
		{name: "reset service validation", fn: (*Branding).ResetAsset, svc: &fakeBrandingService{resetErr: errors.New("purpose tidak valid")}, req: withRouteParam(adminRequest(http.MethodDelete, "/api/branding/assets/logo", ""), "purpose", "logo"), want: http.StatusBadRequest},
		{name: "asset traversal", fn: (*Branding).Asset, svc: &fakeBrandingService{}, req: withRouteParam(httptest.NewRequest(http.MethodGet, "/api/branding/file/../secret", nil), "filename", "../secret"), want: http.StatusNotFound},
		{name: "asset bad prefix", fn: (*Branding).Asset, svc: &fakeBrandingService{}, req: withRouteParam(httptest.NewRequest(http.MethodGet, "/api/branding/file/other.png", nil), "filename", "other.png"), want: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(NewBranding(tt.svc, t.TempDir()), rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func validPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: 120, B: 180, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func multipartBrandingRequest(t *testing.T, purpose, filename string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := adminRequest(http.MethodPost, "/api/branding/assets/"+purpose, "")
	req.Body = io.NopCloser(bytes.NewReader(body.Bytes()))
	req.ContentLength = int64(body.Len())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
