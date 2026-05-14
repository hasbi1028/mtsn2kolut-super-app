package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type settingStore interface {
	ListSettings(ctx context.Context) ([]db.AppSetting, error)
	GetSetting(ctx context.Context, key string) (db.AppSetting, error)
	UpsertSetting(ctx context.Context, arg db.UpsertSettingParams) error
}

type Setting struct {
	q settingStore
}

func NewSetting(q *db.Queries) *Setting { return &Setting{q: q} }

type SchoolProfile struct {
	Name         string `json:"name"`
	NSM          string `json:"nsm"`
	NPSN         string `json:"npsn"`
	MinistryLine string `json:"ministry_line"`
	OfficeLine   string `json:"office_line"`
	Address      string `json:"address"`
	Village      string `json:"village"`
	District     string `json:"district"`
	Regency      string `json:"regency"`
	Province     string `json:"province"`
	PostalCode   string `json:"postal_code"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Website      string `json:"website"`
	HeadName     string `json:"head_name"`
	HeadNIP      string `json:"head_nip"`
}

type BrandingSettings struct {
	AppName           string `json:"app_name"`
	ShortName         string `json:"short_name"`
	Tagline           string `json:"tagline"`
	PrimaryColor      string `json:"primary_color"`
	ThemeColor        string `json:"theme_color"`
	LogoURL           string `json:"logo_url"`
	MarkURL           string `json:"mark_url"`
	FaviconURL        string `json:"favicon_url"`
	AppleTouchIconURL string `json:"apple_touch_icon_url"`
	PWAIcon192URL     string `json:"pwa_icon_192_url"`
	PWAIcon512URL     string `json:"pwa_icon_512_url"`
	FormalLogoURL     string `json:"formal_logo_url"`
	Version           string `json:"version"`
}

func (s *Setting) List(ctx context.Context) ([]db.AppSetting, error) {
	return s.q.ListSettings(ctx)
}

func (s *Setting) Get(ctx context.Context, key string) (db.AppSetting, error) {
	return s.q.GetSetting(ctx, key)
}

func (s *Setting) Upsert(ctx context.Context, key, value string) error {
	return s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: key, Value: value})
}

func (s *Setting) SchoolProfile(ctx context.Context) (SchoolProfile, error) {
	rows, err := s.q.ListSettings(ctx)
	if err != nil {
		return SchoolProfile{}, err
	}
	values := schoolProfileDefaultMap()
	for _, row := range rows {
		if _, ok := values[row.Key]; ok {
			values[row.Key] = row.Value
		}
	}
	return schoolProfileFromMap(values), nil
}

func (s *Setting) Branding(ctx context.Context) (BrandingSettings, error) {
	rows, err := s.q.ListSettings(ctx)
	if err != nil {
		return BrandingSettings{}, err
	}
	values := brandingDefaultMap()
	for _, row := range rows {
		if _, ok := values[row.Key]; ok {
			values[row.Key] = row.Value
		}
	}
	return brandingFromMap(values), nil
}

func (s *Setting) UpdateBranding(ctx context.Context, settings BrandingSettings) (BrandingSettings, error) {
	current, err := s.Branding(ctx)
	if err != nil {
		return BrandingSettings{}, err
	}
	settings = normalizeBrandingSettings(settings, current)
	if settings.AppName == "" {
		return BrandingSettings{}, errors.New("nama aplikasi wajib diisi")
	}
	if !validHexColor(settings.PrimaryColor) || !validHexColor(settings.ThemeColor) {
		return BrandingSettings{}, errors.New("warna branding wajib format hex, contoh #166534")
	}
	settings.Version = newBrandingVersion()
	for key, value := range brandingToMap(settings) {
		if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: key, Value: value}); err != nil {
			return BrandingSettings{}, err
		}
	}
	return settings, nil
}

func (s *Setting) UpdateBrandingAsset(ctx context.Context, purpose, url, hash string) (BrandingSettings, error) {
	settings, err := s.Branding(ctx)
	if err != nil {
		return BrandingSettings{}, err
	}
	if !setBrandingAssetURL(&settings, purpose, url) {
		return BrandingSettings{}, errors.New("jenis aset branding tidak didukung")
	}
	settings.Version = newBrandingVersion()
	values := brandingToMap(settings)
	values[brandingPrefix+purpose+"_sha256"] = strings.TrimSpace(hash)
	for key, value := range values {
		if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: key, Value: value}); err != nil {
			return BrandingSettings{}, err
		}
	}
	return settings, nil
}

func (s *Setting) ResetBrandingAsset(ctx context.Context, purpose string) (BrandingSettings, error) {
	settings, err := s.Branding(ctx)
	if err != nil {
		return BrandingSettings{}, err
	}
	defaults := brandingFromMap(brandingDefaultMap())
	var ok bool
	switch purpose {
	case "logo":
		settings.LogoURL, ok = defaults.LogoURL, true
	case "mark":
		settings.MarkURL, ok = defaults.MarkURL, true
	case "favicon":
		settings.FaviconURL, ok = defaults.FaviconURL, true
	case "apple_touch_icon":
		settings.AppleTouchIconURL, ok = defaults.AppleTouchIconURL, true
	case "pwa_icon_192":
		settings.PWAIcon192URL, ok = defaults.PWAIcon192URL, true
	case "pwa_icon_512":
		settings.PWAIcon512URL, ok = defaults.PWAIcon512URL, true
	case "formal_logo":
		settings.FormalLogoURL, ok = defaults.FormalLogoURL, true
	}
	if !ok {
		return BrandingSettings{}, errors.New("jenis aset branding tidak didukung")
	}
	settings.Version = newBrandingVersion()
	for key, value := range brandingToMap(settings) {
		if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: key, Value: value}); err != nil {
			return BrandingSettings{}, err
		}
	}
	return settings, nil
}

func (s *Setting) UpdateSchoolProfile(ctx context.Context, profile SchoolProfile) (SchoolProfile, error) {
	profile = normalizeSchoolProfile(profile)
	if profile.Name == "" {
		return SchoolProfile{}, errors.New("nama madrasah wajib diisi")
	}
	values := schoolProfileToMap(profile)
	for key, value := range values {
		if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: key, Value: value}); err != nil {
			return SchoolProfile{}, err
		}
	}
	return profile, nil
}

func (s *Setting) SeedDefaults(ctx context.Context) error {
	defaults := map[string]string{
		"default_max_attempts":         "3",
		"max_concurrent":               "5",
		"headless":                     "true",
		"pusaka_geo_base_lat":          "-3.2163111",
		"pusaka_geo_base_lng":          "121.0428659",
		"pusaka_geo_default_radius_m":  "50",
		"pusaka_geo_checkin_radius_m":  "55",
		"pusaka_geo_checkout_radius_m": "28",
		"scheduler_last_error":         "",
	}
	for key, value := range schoolProfileDefaultMap() {
		defaults[key] = value
	}
	for key, value := range brandingDefaultMap() {
		defaults[key] = value
	}

	for key, value := range defaults {
		_, err := s.q.GetSetting(ctx, key)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: key, Value: value}); err != nil {
			return err
		}
	}

	return nil
}

const schoolProfilePrefix = "school_profile."

func schoolProfileDefaultMap() map[string]string {
	return map[string]string{
		schoolProfilePrefix + "name":          "MTs Negeri 2 Kolaka Utara",
		schoolProfilePrefix + "nsm":           "",
		schoolProfilePrefix + "npsn":          DefaultEmployeeAccountNPSN,
		schoolProfilePrefix + "ministry_line": "Kementerian Agama Republik Indonesia",
		schoolProfilePrefix + "office_line":   "Kantor Kementerian Agama Kabupaten Kolaka Utara",
		schoolProfilePrefix + "address":       "",
		schoolProfilePrefix + "village":       "",
		schoolProfilePrefix + "district":      "",
		schoolProfilePrefix + "regency":       "Kolaka Utara",
		schoolProfilePrefix + "province":      "Sulawesi Tenggara",
		schoolProfilePrefix + "postal_code":   "",
		schoolProfilePrefix + "phone":         "",
		schoolProfilePrefix + "email":         "",
		schoolProfilePrefix + "website":       "",
		schoolProfilePrefix + "head_name":     "",
		schoolProfilePrefix + "head_nip":      "",
	}
}

func schoolProfileFromMap(values map[string]string) SchoolProfile {
	return SchoolProfile{
		Name:         values[schoolProfilePrefix+"name"],
		NSM:          values[schoolProfilePrefix+"nsm"],
		NPSN:         values[schoolProfilePrefix+"npsn"],
		MinistryLine: values[schoolProfilePrefix+"ministry_line"],
		OfficeLine:   values[schoolProfilePrefix+"office_line"],
		Address:      values[schoolProfilePrefix+"address"],
		Village:      values[schoolProfilePrefix+"village"],
		District:     values[schoolProfilePrefix+"district"],
		Regency:      values[schoolProfilePrefix+"regency"],
		Province:     values[schoolProfilePrefix+"province"],
		PostalCode:   values[schoolProfilePrefix+"postal_code"],
		Phone:        values[schoolProfilePrefix+"phone"],
		Email:        values[schoolProfilePrefix+"email"],
		Website:      values[schoolProfilePrefix+"website"],
		HeadName:     values[schoolProfilePrefix+"head_name"],
		HeadNIP:      values[schoolProfilePrefix+"head_nip"],
	}
}

func schoolProfileToMap(profile SchoolProfile) map[string]string {
	return map[string]string{
		schoolProfilePrefix + "name":          profile.Name,
		schoolProfilePrefix + "nsm":           profile.NSM,
		schoolProfilePrefix + "npsn":          profile.NPSN,
		schoolProfilePrefix + "ministry_line": profile.MinistryLine,
		schoolProfilePrefix + "office_line":   profile.OfficeLine,
		schoolProfilePrefix + "address":       profile.Address,
		schoolProfilePrefix + "village":       profile.Village,
		schoolProfilePrefix + "district":      profile.District,
		schoolProfilePrefix + "regency":       profile.Regency,
		schoolProfilePrefix + "province":      profile.Province,
		schoolProfilePrefix + "postal_code":   profile.PostalCode,
		schoolProfilePrefix + "phone":         profile.Phone,
		schoolProfilePrefix + "email":         profile.Email,
		schoolProfilePrefix + "website":       profile.Website,
		schoolProfilePrefix + "head_name":     profile.HeadName,
		schoolProfilePrefix + "head_nip":      profile.HeadNIP,
	}
}

func normalizeSchoolProfile(profile SchoolProfile) SchoolProfile {
	profile.Name = strings.TrimSpace(profile.Name)
	profile.NSM = strings.TrimSpace(profile.NSM)
	profile.NPSN = strings.TrimSpace(profile.NPSN)
	profile.MinistryLine = strings.TrimSpace(profile.MinistryLine)
	profile.OfficeLine = strings.TrimSpace(profile.OfficeLine)
	profile.Address = strings.TrimSpace(profile.Address)
	profile.Village = strings.TrimSpace(profile.Village)
	profile.District = strings.TrimSpace(profile.District)
	profile.Regency = strings.TrimSpace(profile.Regency)
	profile.Province = strings.TrimSpace(profile.Province)
	profile.PostalCode = strings.TrimSpace(profile.PostalCode)
	profile.Phone = strings.TrimSpace(profile.Phone)
	profile.Email = strings.TrimSpace(profile.Email)
	profile.Website = strings.TrimSpace(profile.Website)
	profile.HeadName = strings.TrimSpace(profile.HeadName)
	profile.HeadNIP = strings.TrimSpace(profile.HeadNIP)
	if profile.MinistryLine == "" {
		profile.MinistryLine = "Kementerian Agama Republik Indonesia"
	}
	if profile.OfficeLine == "" {
		profile.OfficeLine = "Kantor Kementerian Agama Kabupaten Kolaka Utara"
	}
	return profile
}

const brandingPrefix = "branding."

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func brandingDefaultMap() map[string]string {
	return map[string]string{
		brandingPrefix + "app_name":             "MTs Negeri 2 Kolaka Utara",
		brandingPrefix + "short_name":           "MTsN 2 Kolut",
		brandingPrefix + "tagline":              "Super App Madrasah",
		brandingPrefix + "primary_color":        "#166534",
		brandingPrefix + "theme_color":          "#166534",
		brandingPrefix + "logo_url":             "/brand/madrasah-mark.svg",
		brandingPrefix + "mark_url":             "/brand/madrasah-mark.svg",
		brandingPrefix + "favicon_url":          "/favicon.ico",
		brandingPrefix + "apple_touch_icon_url": "/apple-touch-icon.png",
		brandingPrefix + "pwa_icon_192_url":     "/pwa-icon-192.png",
		brandingPrefix + "pwa_icon_512_url":     "/pwa-icon-512.png",
		brandingPrefix + "formal_logo_url":      "/brand/logo-kemenag.png",
		brandingPrefix + "version":              "default",
	}
}

func brandingFromMap(values map[string]string) BrandingSettings {
	settings := BrandingSettings{
		AppName:           values[brandingPrefix+"app_name"],
		ShortName:         values[brandingPrefix+"short_name"],
		Tagline:           values[brandingPrefix+"tagline"],
		PrimaryColor:      values[brandingPrefix+"primary_color"],
		ThemeColor:        values[brandingPrefix+"theme_color"],
		LogoURL:           values[brandingPrefix+"logo_url"],
		MarkURL:           values[brandingPrefix+"mark_url"],
		FaviconURL:        values[brandingPrefix+"favicon_url"],
		AppleTouchIconURL: values[brandingPrefix+"apple_touch_icon_url"],
		PWAIcon192URL:     values[brandingPrefix+"pwa_icon_192_url"],
		PWAIcon512URL:     values[brandingPrefix+"pwa_icon_512_url"],
		FormalLogoURL:     values[brandingPrefix+"formal_logo_url"],
		Version:           values[brandingPrefix+"version"],
	}
	return normalizeBrandingSettings(settings, brandingFromMapWithoutNormalize(brandingDefaultMap()))
}

func brandingFromMapWithoutNormalize(values map[string]string) BrandingSettings {
	return BrandingSettings{
		AppName:           values[brandingPrefix+"app_name"],
		ShortName:         values[brandingPrefix+"short_name"],
		Tagline:           values[brandingPrefix+"tagline"],
		PrimaryColor:      values[brandingPrefix+"primary_color"],
		ThemeColor:        values[brandingPrefix+"theme_color"],
		LogoURL:           values[brandingPrefix+"logo_url"],
		MarkURL:           values[brandingPrefix+"mark_url"],
		FaviconURL:        values[brandingPrefix+"favicon_url"],
		AppleTouchIconURL: values[brandingPrefix+"apple_touch_icon_url"],
		PWAIcon192URL:     values[brandingPrefix+"pwa_icon_192_url"],
		PWAIcon512URL:     values[brandingPrefix+"pwa_icon_512_url"],
		FormalLogoURL:     values[brandingPrefix+"formal_logo_url"],
		Version:           values[brandingPrefix+"version"],
	}
}

func brandingToMap(settings BrandingSettings) map[string]string {
	return map[string]string{
		brandingPrefix + "app_name":             settings.AppName,
		brandingPrefix + "short_name":           settings.ShortName,
		brandingPrefix + "tagline":              settings.Tagline,
		brandingPrefix + "primary_color":        settings.PrimaryColor,
		brandingPrefix + "theme_color":          settings.ThemeColor,
		brandingPrefix + "logo_url":             settings.LogoURL,
		brandingPrefix + "mark_url":             settings.MarkURL,
		brandingPrefix + "favicon_url":          settings.FaviconURL,
		brandingPrefix + "apple_touch_icon_url": settings.AppleTouchIconURL,
		brandingPrefix + "pwa_icon_192_url":     settings.PWAIcon192URL,
		brandingPrefix + "pwa_icon_512_url":     settings.PWAIcon512URL,
		brandingPrefix + "formal_logo_url":      settings.FormalLogoURL,
		brandingPrefix + "version":              settings.Version,
	}
}

func normalizeBrandingSettings(settings BrandingSettings, fallback BrandingSettings) BrandingSettings {
	settings.AppName = strings.TrimSpace(settings.AppName)
	settings.ShortName = strings.TrimSpace(settings.ShortName)
	settings.Tagline = strings.TrimSpace(settings.Tagline)
	settings.PrimaryColor = strings.TrimSpace(settings.PrimaryColor)
	settings.ThemeColor = strings.TrimSpace(settings.ThemeColor)
	settings.LogoURL = safeBrandURL(settings.LogoURL, fallback.LogoURL)
	settings.MarkURL = safeBrandURL(settings.MarkURL, fallback.MarkURL)
	settings.FaviconURL = safeBrandURL(settings.FaviconURL, fallback.FaviconURL)
	settings.AppleTouchIconURL = safeBrandURL(settings.AppleTouchIconURL, fallback.AppleTouchIconURL)
	settings.PWAIcon192URL = safeBrandURL(settings.PWAIcon192URL, fallback.PWAIcon192URL)
	settings.PWAIcon512URL = safeBrandURL(settings.PWAIcon512URL, fallback.PWAIcon512URL)
	settings.FormalLogoURL = safeBrandURL(settings.FormalLogoURL, fallback.FormalLogoURL)
	settings.Version = strings.TrimSpace(settings.Version)
	if settings.AppName == "" {
		settings.AppName = fallback.AppName
	}
	if settings.ShortName == "" {
		settings.ShortName = fallback.ShortName
	}
	if settings.PrimaryColor == "" {
		settings.PrimaryColor = fallback.PrimaryColor
	}
	if settings.ThemeColor == "" {
		settings.ThemeColor = fallback.ThemeColor
	}
	if settings.Version == "" {
		settings.Version = fallback.Version
	}
	return settings
}

func setBrandingAssetURL(settings *BrandingSettings, purpose, url string) bool {
	url = safeBrandURL(url, "")
	if url == "" {
		return false
	}
	switch purpose {
	case "logo":
		settings.LogoURL = url
	case "mark":
		settings.MarkURL = url
	case "favicon":
		settings.FaviconURL = url
	case "apple_touch_icon":
		settings.AppleTouchIconURL = url
	case "pwa_icon_192":
		settings.PWAIcon192URL = url
	case "pwa_icon_512":
		settings.PWAIcon512URL = url
	case "formal_logo":
		settings.FormalLogoURL = url
	default:
		return false
	}
	return true
}

func safeBrandURL(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && !strings.Contains(value, "..") {
		return value
	}
	return fallback
}

func validHexColor(value string) bool { return hexColorPattern.MatchString(value) }

func newBrandingVersion() string { return time.Now().UTC().Format("20060102150405") }
