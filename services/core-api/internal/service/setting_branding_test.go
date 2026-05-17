package service

import (
	"context"
	"strings"
	"testing"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestBrandingPureHelpersNormalizeAndRoundTrip(t *testing.T) {
	fallback := BrandingSettings{
		AppName: "Fallback App", ShortName: "FB", Tagline: "Tag",
		PrimaryColor: "#111111", ThemeColor: "#222222",
		LogoURL: "/logo.svg", MarkURL: "/mark.svg", FaviconURL: "/favicon.ico",
		AppleTouchIconURL: "/apple.png", PWAIcon192URL: "/192.png", PWAIcon512URL: "/512.png",
		FormalLogoURL: "/formal.png", Version: "v1",
	}
	settings := normalizeBrandingSettings(BrandingSettings{
		AppName: "  Custom App  ", ShortName: " ", Tagline: "  Welcome  ",
		PrimaryColor: " ", ThemeColor: "#ABCDEF",
		LogoURL: " //evil.example/logo.png ", MarkURL: "/brand/mark.svg",
		FaviconURL: "/safe/../bad.ico", Version: " ",
	}, fallback)

	if settings.AppName != "Custom App" || settings.ShortName != fallback.ShortName || settings.Tagline != "Welcome" {
		t.Fatalf("text fields not normalized/fallbacked: %+v", settings)
	}
	if settings.PrimaryColor != fallback.PrimaryColor || settings.ThemeColor != "#ABCDEF" {
		t.Fatalf("colors not normalized/fallbacked: %+v", settings)
	}
	if settings.LogoURL != fallback.LogoURL || settings.MarkURL != "/brand/mark.svg" || settings.FaviconURL != fallback.FaviconURL {
		t.Fatalf("asset URLs not sanitized/fallbacked: %+v", settings)
	}
	if settings.Version != fallback.Version {
		t.Fatalf("Version = %q, want fallback", settings.Version)
	}

	mapped := brandingToMap(settings)
	roundTrip := brandingFromMap(mapped)
	if roundTrip.AppName != settings.AppName || roundTrip.MarkURL != settings.MarkURL || roundTrip.ThemeColor != settings.ThemeColor {
		t.Fatalf("branding round trip = %+v, want %+v", roundTrip, settings)
	}
}

func TestBrandingURLColorVersionAndAssetHelpers(t *testing.T) {
	for _, tc := range []struct{ in, fallback, want string }{
		{" /brand/logo.svg ", "/fallback.svg", "/brand/logo.svg"},
		{"//cdn.example/logo.svg", "/fallback.svg", "/fallback.svg"},
		{"https://example.test/logo.svg", "/fallback.svg", "/fallback.svg"},
		{"/brand/../secret.svg", "/fallback.svg", "/fallback.svg"},
		{"", "/fallback.svg", "/fallback.svg"},
	} {
		if got := safeBrandURL(tc.in, tc.fallback); got != tc.want {
			t.Fatalf("safeBrandURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	for _, color := range []string{"#166534", "#ABCDEF", "#123abc"} {
		if !validHexColor(color) {
			t.Fatalf("validHexColor(%q) = false", color)
		}
	}
	for _, color := range []string{"166534", "#123abz", "#123", " #166534 "} {
		if validHexColor(color) {
			t.Fatalf("validHexColor(%q) = true", color)
		}
	}
	if v := newBrandingVersion(); len(v) != 14 || strings.Trim(v, "0123456789") != "" {
		t.Fatalf("newBrandingVersion() = %q, want 14 digits", v)
	}

	var settings BrandingSettings
	for purpose, want := range map[string]string{
		"logo": "/u/logo.png", "mark": "/u/mark.png", "favicon": "/u/favicon.ico",
		"apple_touch_icon": "/u/apple.png", "pwa_icon_192": "/u/192.png", "pwa_icon_512": "/u/512.png",
		"formal_logo": "/u/formal.png",
	} {
		if !setBrandingAssetURL(&settings, purpose, want) {
			t.Fatalf("setBrandingAssetURL(%q) = false", purpose)
		}
	}
	if settings.LogoURL != "/u/logo.png" || settings.PWAIcon512URL != "/u/512.png" || settings.FormalLogoURL != "/u/formal.png" {
		t.Fatalf("asset fields not set: %+v", settings)
	}
	if setBrandingAssetURL(&settings, "logo", "https://evil.test/logo.png") {
		t.Fatal("setBrandingAssetURL accepted unsafe URL")
	}
	if setBrandingAssetURL(&settings, "unknown", "/ok.png") {
		t.Fatal("setBrandingAssetURL accepted unknown purpose")
	}
}

func TestSettingBrandingWrappersUseStoreAndValidate(t *testing.T) {
	store := newFakeStore()
	store.settings[brandingPrefix+"app_name"] = "Custom"
	store.settings[brandingPrefix+"logo_url"] = "https://evil.test/logo.png"
	svc := &Setting{q: store}

	loaded, err := svc.Branding(context.Background())
	if err != nil {
		t.Fatalf("Branding() error = %v", err)
	}
	if loaded.AppName != "Custom" || loaded.LogoURL != "/brand/madrasah-mark.svg" || loaded.PrimaryColor != "#166534" {
		t.Fatalf("Branding() = %+v", loaded)
	}

	updated, err := svc.UpdateBranding(context.Background(), BrandingSettings{
		AppName: " Updated ", ShortName: "Short", PrimaryColor: "#000000", ThemeColor: "#ffffff",
		LogoURL: "/updated.svg", MarkURL: "/updated-mark.svg",
	})
	if err != nil {
		t.Fatalf("UpdateBranding() error = %v", err)
	}
	if updated.AppName != "Updated" || updated.LogoURL != "/updated.svg" || updated.Version == "" || updated.Version == "default" {
		t.Fatalf("UpdateBranding() = %+v", updated)
	}
	if store.settings[brandingPrefix+"app_name"] != "Updated" || store.settings[brandingPrefix+"version"] != updated.Version {
		t.Fatalf("stored branding not updated: %+v", store.settings)
	}
	if _, err := svc.UpdateBranding(context.Background(), BrandingSettings{AppName: "Bad", PrimaryColor: "green", ThemeColor: "#ffffff"}); err == nil {
		t.Fatal("UpdateBranding() accepted invalid color")
	}

	asset, err := svc.UpdateBrandingAsset(context.Background(), "logo", "/uploads/logo.png", "abc123")
	if err != nil {
		t.Fatalf("UpdateBrandingAsset() error = %v", err)
	}
	if asset.LogoURL != "/uploads/logo.png" || store.settings[brandingPrefix+"logo_sha256"] != "abc123" {
		t.Fatalf("asset update not stored: asset=%+v settings=%+v", asset, store.settings)
	}
	if _, err := svc.UpdateBrandingAsset(context.Background(), "unknown", "/x.png", ""); err == nil {
		t.Fatal("UpdateBrandingAsset() accepted unknown purpose")
	}

	reset, err := svc.ResetBrandingAsset(context.Background(), "logo")
	if err != nil {
		t.Fatalf("ResetBrandingAsset() error = %v", err)
	}
	if reset.LogoURL != "/brand/madrasah-mark.svg" || store.settings[brandingPrefix+"logo_url"] != "/brand/madrasah-mark.svg" {
		t.Fatalf("ResetBrandingAsset() = %+v settings=%+v", reset, store.settings)
	}
}

func TestSettingBrandingReturnsStoreErrors(t *testing.T) {
	listErr := assertErr("list failed")
	upsertErr := assertErr("upsert failed")
	if _, err := (&Setting{q: failingSettingStore{listErr: listErr}}).Branding(context.Background()); err != listErr {
		t.Fatalf("Branding() error = %v, want %v", err, listErr)
	}
	_, err := (&Setting{q: failingSettingStore{upsertErr: upsertErr}}).UpdateBranding(context.Background(), BrandingSettings{AppName: "App", PrimaryColor: "#000000", ThemeColor: "#111111"})
	if err != upsertErr {
		t.Fatalf("UpdateBranding() error = %v, want %v", err, upsertErr)
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

var _ settingStore = (*fakeBrandingOnlyStore)(nil)

type fakeBrandingOnlyStore struct{ settings []db.AppSetting }

func (f fakeBrandingOnlyStore) ListSettings(context.Context) ([]db.AppSetting, error) {
	return f.settings, nil
}
func (f fakeBrandingOnlyStore) GetSetting(context.Context, string) (db.AppSetting, error) {
	return db.AppSetting{}, nil
}
func (f fakeBrandingOnlyStore) UpsertSetting(context.Context, db.UpsertSettingParams) error {
	return nil
}
