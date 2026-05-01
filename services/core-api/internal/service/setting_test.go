package service

import (
	"context"
	"testing"
)

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
