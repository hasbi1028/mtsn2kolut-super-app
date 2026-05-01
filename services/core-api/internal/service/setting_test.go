package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestSettingListGetAndUpsertForwardStoreCalls(t *testing.T) {
	store := newFakeStore()
	store.settings["site_title"] = "Madrasah"
	svc := &Setting{q: store}

	rows, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Key != "site_title" {
		t.Fatalf("List() = %+v, want site_title row", rows)
	}
	got, err := svc.Get(context.Background(), "site_title")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Key != "site_title" || got.Value != "Madrasah" {
		t.Fatalf("Get() = %+v, want site title value", got)
	}
	if err := svc.Upsert(context.Background(), "site_title", "MTsN 2"); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if store.settings["site_title"] != "MTsN 2" {
		t.Fatalf("stored site_title = %q, want MTsN 2", store.settings["site_title"])
	}
}

func TestSettingSchoolProfileUsesDefaultsForUnknownAndMissingRows(t *testing.T) {
	store := newFakeStore()
	store.settings["school_profile.name"] = "Nama Khusus"
	store.settings["unknown"] = "ignored"
	svc := &Setting{q: store}

	profile, err := svc.SchoolProfile(context.Background())
	if err != nil {
		t.Fatalf("SchoolProfile() error = %v", err)
	}
	if profile.Name != "Nama Khusus" {
		t.Fatalf("Name = %q, want custom value", profile.Name)
	}
	if profile.MinistryLine != "Kementerian Agama Republik Indonesia" || profile.Regency != "Kolaka Utara" {
		t.Fatalf("defaults were not preserved: %+v", profile)
	}
}

func TestSettingSeedDefaultsSkipsExistingRows(t *testing.T) {
	store := newFakeStore()
	store.settings["default_max_attempts"] = "9"
	svc := &Setting{q: store}

	if err := svc.SeedDefaults(context.Background()); err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}
	if store.settings["default_max_attempts"] != "9" {
		t.Fatalf("default_max_attempts = %q, want existing value 9", store.settings["default_max_attempts"])
	}
	if store.settings["max_concurrent"] != "5" {
		t.Fatalf("max_concurrent = %q, want seeded 5", store.settings["max_concurrent"])
	}
}

func TestSettingSeedDefaultsReturnsStoreErrors(t *testing.T) {
	lookupErr := errors.New("lookup failed")
	svc := &Setting{q: failingSettingStore{getErr: lookupErr}}
	if err := svc.SeedDefaults(context.Background()); !errors.Is(err, lookupErr) {
		t.Fatalf("SeedDefaults() lookup error = %v, want %v", err, lookupErr)
	}

	upsertErr := errors.New("upsert failed")
	svc = &Setting{q: failingSettingStore{getErr: pgx.ErrNoRows, upsertErr: upsertErr}}
	if err := svc.SeedDefaults(context.Background()); !errors.Is(err, upsertErr) {
		t.Fatalf("SeedDefaults() upsert error = %v, want %v", err, upsertErr)
	}
}

func TestSettingReturnsStoreErrors(t *testing.T) {
	listErr := errors.New("list failed")
	upsertErr := errors.New("upsert failed")
	svc := &Setting{q: failingSettingStore{listErr: listErr, upsertErr: upsertErr}}

	if _, err := svc.SchoolProfile(context.Background()); !errors.Is(err, listErr) {
		t.Fatalf("SchoolProfile() error = %v, want list failed", err)
	}
	if _, err := svc.UpdateSchoolProfile(context.Background(), SchoolProfile{Name: "MTs"}); !errors.Is(err, upsertErr) {
		t.Fatalf("UpdateSchoolProfile() error = %v, want upsert failed", err)
	}
}

type failingSettingStore struct {
	listErr   error
	getErr    error
	upsertErr error
}

func (f failingSettingStore) ListSettings(ctx context.Context) ([]db.AppSetting, error) {
	return nil, f.listErr
}

func (f failingSettingStore) GetSetting(ctx context.Context, key string) (db.AppSetting, error) {
	return db.AppSetting{}, f.getErr
}

func (f failingSettingStore) UpsertSetting(ctx context.Context, arg db.UpsertSettingParams) error {
	return f.upsertErr
}

func TestSettingSeedDefaults(t *testing.T) {
	store := newFakeStore()
	svc := &Setting{q: store}
	if err := svc.SeedDefaults(context.Background()); err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	for key, want := range map[string]string{
		"default_max_attempts": "3",
		"max_concurrent":       "5",
		"headless":             "true",
		"scheduler_last_error": "",
		"school_profile.name":  "MTs Negeri 2 Kolaka Utara",
	} {
		if got := store.settings[key]; got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestSettingSchoolProfile(t *testing.T) {
	store := newFakeStore()
	svc := &Setting{q: store}

	got, err := svc.UpdateSchoolProfile(context.Background(), SchoolProfile{
		Name:         "  MTs Negeri 2 Kolaka Utara  ",
		MinistryLine: " ",
		OfficeLine:   "",
		Address:      " Jl. Pendidikan ",
		HeadName:     " Kepala Madrasah ",
	})
	if err != nil {
		t.Fatalf("UpdateSchoolProfile() error = %v", err)
	}
	if got.Name != "MTs Negeri 2 Kolaka Utara" {
		t.Fatalf("Name = %q", got.Name)
	}
	if got.MinistryLine != "Kementerian Agama Republik Indonesia" {
		t.Fatalf("MinistryLine = %q", got.MinistryLine)
	}
	if got.Address != "Jl. Pendidikan" || got.HeadName != "Kepala Madrasah" {
		t.Fatalf("profile was not normalized: %+v", got)
	}

	loaded, err := svc.SchoolProfile(context.Background())
	if err != nil {
		t.Fatalf("SchoolProfile() error = %v", err)
	}
	if loaded.Name != got.Name || loaded.HeadName != got.HeadName {
		t.Fatalf("loaded profile = %+v, want %+v", loaded, got)
	}
}

func TestSettingSchoolProfileRequiresName(t *testing.T) {
	store := newFakeStore()
	svc := &Setting{q: store}

	if _, err := svc.UpdateSchoolProfile(context.Background(), SchoolProfile{}); err == nil {
		t.Fatal("UpdateSchoolProfile() error = nil, want error")
	}
}
