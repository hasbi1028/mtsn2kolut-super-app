package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestAuthContactHelperExtraBranches(t *testing.T) {
	for _, tt := range []struct {
		profile string
		want    []string
	}{
		{profile: "employee", want: []string{"phone", "email", "address"}},
		{profile: "student", want: []string{"phone", "address"}},
		{profile: "parent", want: []string{"phone", "address"}},
		{profile: "other", want: []string{}},
	} {
		got := contactEditableFields(tt.profile)
		if len(got) != len(tt.want) {
			t.Fatalf("contactEditableFields(%q) len = %d, want %d (%#v)", tt.profile, len(got), len(tt.want), got)
		}
		for i := range tt.want {
			if got[i] != tt.want[i] {
				t.Fatalf("contactEditableFields(%q)[%d] = %q, want %q", tt.profile, i, got[i], tt.want[i])
			}
		}
	}

	if got := stringSliceFromJSONValue(nil); len(got) != 0 {
		t.Fatalf("stringSliceFromJSONValue(nil) = %#v, want empty", got)
	}
	if got := stringSliceFromJSONValue([]byte(`["admin","guru"]`)); len(got) != 2 || got[0] != "admin" || got[1] != "guru" {
		t.Fatalf("stringSliceFromJSONValue([]byte) = %#v, want admin/guru", got)
	}
	if got := stringSliceFromJSONValue(`["siswa"]`); len(got) != 1 || got[0] != "siswa" {
		t.Fatalf("stringSliceFromJSONValue(string) = %#v, want siswa", got)
	}
	if got := stringSliceFromJSONValue([]string{"wali", "staf"}); len(got) != 2 || got[0] != "wali" || got[1] != "staf" {
		t.Fatalf("stringSliceFromJSONValue([]string) = %#v, want wali/staf", got)
	}
	if got := stringSliceFromJSONValue(func() {}); len(got) != 0 {
		t.Fatalf("stringSliceFromJSONValue(unmarshalable) = %#v, want empty", got)
	}
	if got := stringSliceFromJSONValue(`not-json`); len(got) != 0 {
		t.Fatalf("stringSliceFromJSONValue(invalid json) = %#v, want empty", got)
	}

	empty, err := decodeContactString(json.RawMessage(` null `))
	if err != nil || empty == nil || *empty != "" {
		t.Fatalf("decodeContactString(null) = %v/%v, want pointer to empty string", empty, err)
	}
	value, err := decodeContactString(json.RawMessage(`" 0812 "`))
	if err != nil || value == nil || *value != " 0812 " {
		t.Fatalf("decodeContactString(string) = %v/%v, want original string", value, err)
	}
	if _, err := decodeContactString(json.RawMessage(`123`)); err == nil {
		t.Fatal("decodeContactString(number) error = nil, want json type error")
	}
}

func TestDecodeAccountContactPatchRejectsUnknownAndInvalidFields(t *testing.T) {
	for _, tt := range []struct {
		name string
		body string
	}{
		{name: "malformed", body: `{`},
		{name: "unknown field", body: `{"nickname":"x"}`},
		{name: "non-string", body: `{"phone":123}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/account/contact", bytes.NewBufferString(tt.body))
			if _, ok := decodeAccountContactPatch(rec, req); ok {
				t.Fatal("decodeAccountContactPatch() ok = true, want false")
			}
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
		})
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/account/contact", bytes.NewBufferString(`{"phone":null,"email":"a@example.test","address":"Jl"}`))
	patch, ok := decodeAccountContactPatch(rec, req)
	if !ok || patch.Phone == nil || *patch.Phone != "" || patch.Email == nil || *patch.Email != "a@example.test" || patch.Address == nil || *patch.Address != "Jl" {
		t.Fatalf("decodeAccountContactPatch(valid) = %+v/%v, want parsed nil-as-empty fields", patch, ok)
	}
}

func TestAuthDeleteAccountAvatarExtraBranches(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"

	rec := httptest.NewRecorder()
	NewAuth(&fakeAuthService{}, nil).DeleteAccountAvatar(rec, httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/account/avatar", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("DeleteAccountAvatar(no claims) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/account/avatar", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "not-a-uuid"))
	NewAuth(&fakeAuthService{}, nil).DeleteAccountAvatar(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("DeleteAccountAvatar(bad claims) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/account/avatar", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	NewAuth(&fakeAuthService{deleteAvatarErr: domain.ErrForbidden}, nil).DeleteAccountAvatar(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("DeleteAccountAvatar(forbidden) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/account/avatar", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	NewAuth(&fakeAuthService{deleteAvatarErr: context.Canceled}, nil).DeleteAccountAvatar(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("DeleteAccountAvatar(internal) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}

	audit := &fakeAuthAuditWriter{}
	svc := &fakeAuthService{accountResult: db.GetUserAccountSummaryRow{
		ID:          mustUUID(t, userID),
		Username:    "admin",
		DisplayName: "Admin",
		ProfileType: "employee",
		PhotoUrl:    "",
		IsActive:    true,
	}}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/account/avatar", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	NewAuth(svc, audit).DeleteAccountAvatar(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("DeleteAccountAvatar(success) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.lastAvatarDeleteID != mustUUID(t, userID) {
		t.Fatalf("DeleteAccountAvatar userID = %v, want %s", svc.lastAvatarDeleteID, userID)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "AUTH_ACCOUNT_AVATAR_DELETE" {
		t.Fatalf("DeleteAccountAvatar audit = %#v, want avatar delete audit", audit.entries)
	}
}
