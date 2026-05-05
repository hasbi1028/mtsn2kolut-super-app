package service

import (
	"context"
	"errors"
	"strings"

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
		schoolProfilePrefix + "npsn":          "",
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
