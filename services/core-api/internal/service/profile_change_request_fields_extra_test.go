package service

import (
	"errors"
	"strings"
	"testing"

	"mtsn2kolut-super-app/backend/internal/domain"
)

func TestProfileChangeFieldLabelsAndOfficialNormalization(t *testing.T) {
	field, err := normalizeOfficialChangeField(" student ", "birth-date")
	if err != nil || field != "tanggal_lahir" {
		t.Fatalf("normalizeOfficialChangeField birth-date = %q, %v", field, err)
	}
	field, err = normalizeOfficialChangeField("parent", "WHATSAPP")
	if err != nil || field != "phone" {
		t.Fatalf("normalizeOfficialChangeField whatsapp = %q, %v", field, err)
	}
	if _, err := normalizeOfficialChangeField("employee", "parent_name"); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("normalizeOfficialChangeField wrong profile error = %v, want bad request", err)
	}
	if got := profileChangeFieldLabel("student", "nama-orang-tua"); got != "Nama orang tua/wali" {
		t.Fatalf("profileChangeFieldLabel alias = %q", got)
	}
	if got := profileChangeFieldLabel("", "unknown-field"); got != "unknown_field" {
		t.Fatalf("profileChangeFieldLabel fallback = %q", got)
	}
}

func TestProfileChangeFieldFilterAndValuesRejectInvalidInput(t *testing.T) {
	filter, err := normalizeProfileChangeFieldFilter(" all ")
	if err != nil || filter != "" {
		t.Fatalf("normalizeProfileChangeFieldFilter(all) = %q, %v", filter, err)
	}
	filter, err = normalizeProfileChangeFieldFilter("nama-orang-tua")
	if err != nil || filter != "parent_name" {
		t.Fatalf("normalizeProfileChangeFieldFilter(alias) = %q, %v", filter, err)
	}
	if _, err := normalizeProfileChangeFieldFilter("not-a-field"); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("normalizeProfileChangeFieldFilter invalid error = %v", err)
	}

	dateConfig, err := findProfileChangeFieldConfig("student", "tanggal_lahir", true)
	if err != nil {
		t.Fatalf("find date config: %v", err)
	}
	value, err := normalizeOfficialChangeValue(dateConfig, " 2026-05-17 ")
	if err != nil || value != "2026-05-17" {
		t.Fatalf("normalizeOfficialChangeValue(date) = %q, %v", value, err)
	}
	if _, err := normalizeOfficialChangeValue(dateConfig, "17/05/2026"); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("normalizeOfficialChangeValue bad date error = %v", err)
	}

	textConfig, err := findProfileChangeFieldConfig("student", "alamat", true)
	if err != nil {
		t.Fatalf("find text config: %v", err)
	}
	value, err = normalizeOfficialChangeValue(textConfig, "  Jalan Merdeka  ")
	if err != nil || value != "Jalan Merdeka" {
		t.Fatalf("normalizeOfficialChangeValue(text) = %q, %v", value, err)
	}
	if _, err := normalizeOfficialChangeValue(textConfig, strings.Repeat("x", 201)); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("normalizeOfficialChangeValue long text error = %v", err)
	}
}
