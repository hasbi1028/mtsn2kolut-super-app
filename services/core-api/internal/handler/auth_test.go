package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeAuthService struct {
	getPrefsResult    service.SidebarPreferences
	getPrefsErr       error
	updatePrefsResult service.SidebarPreferences
	updatePrefsErr    error
	lastUserID        pgtype.UUID
	lastUpdatedPrefs  service.SidebarPreferences
}

func (f *fakeAuthService) Login(ctx context.Context, username, password string, meta service.SessionMeta) (domain.TokenPair, error) {
	return domain.TokenPair{}, nil
}

func (f *fakeAuthService) Refresh(ctx context.Context, refreshToken string, meta service.SessionMeta) (domain.TokenPair, error) {
	return domain.TokenPair{}, nil
}

func (f *fakeAuthService) Logout(ctx context.Context, refreshToken string) error { return nil }

func (f *fakeAuthService) LogoutAll(ctx context.Context, userID pgtype.UUID) error { return nil }

func (f *fakeAuthService) ListActiveSessions(ctx context.Context, userID pgtype.UUID) ([]db.AuthSession, error) {
	return nil, nil
}

func (f *fakeAuthService) RevokeSession(ctx context.Context, userID, sessionID pgtype.UUID) error {
	return nil
}

func (f *fakeAuthService) UpdateSessionLabel(ctx context.Context, userID, sessionID pgtype.UUID, deviceLabel string) error {
	return nil
}

func (f *fakeAuthService) GetSidebarPreferences(ctx context.Context, userID pgtype.UUID) (service.SidebarPreferences, error) {
	f.lastUserID = userID
	return f.getPrefsResult, f.getPrefsErr
}

func (f *fakeAuthService) UpdateSidebarPreferences(ctx context.Context, userID pgtype.UUID, prefs service.SidebarPreferences) (service.SidebarPreferences, error) {
	f.lastUserID = userID
	f.lastUpdatedPrefs = prefs
	return f.updatePrefsResult, f.updatePrefsErr
}

func (f *fakeAuthService) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	return nil
}

type fakeAuthAuditWriter struct {
	entries []db.CreateAuditLogParams
}

func (f *fakeAuthAuditWriter) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.entries = append(f.entries, arg)
	return db.AuditLog{}, nil
}

func withAuthClaims(reqCtx context.Context, userID string) context.Context {
	return context.WithValue(reqCtx, api.ClaimsKey, jwt.MapClaims{"sub": userID})
}

func mustUUID(t *testing.T, raw string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		t.Fatalf("invalid uuid %q: %v", raw, err)
	}
	return id
}

func TestAuthGetSidebarPreferencesWritesWrappedJSON(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	svc := &fakeAuthService{
		getPrefsResult: service.SidebarPreferences{
			PinnedItems: []string{"/grades", "/inventory"},
			RecentItems: []string{"/jadwal", "/website/posts"},
		},
	}
	h := NewAuth(svc, nil)

	req := httptest.NewRequest("GET", "http://internal/api/auth/preferences/sidebar", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.GetSidebarPreferences(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	wantUserID := mustUUID(t, userID)
	if svc.lastUserID != wantUserID {
		t.Fatalf("service user id = %#v, want %#v", svc.lastUserID, wantUserID)
	}

	var payload struct {
		Data service.SidebarPreferences `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if len(payload.Data.PinnedItems) != 2 || payload.Data.PinnedItems[0] != "/grades" {
		t.Fatalf("PinnedItems = %#v", payload.Data.PinnedItems)
	}
	if len(payload.Data.RecentItems) != 2 || payload.Data.RecentItems[1] != "/website/posts" {
		t.Fatalf("RecentItems = %#v", payload.Data.RecentItems)
	}
}

func TestAuthGetSidebarPreferencesUnauthorizedWithoutClaims(t *testing.T) {
	h := NewAuth(&fakeAuthService{}, nil)

	req := httptest.NewRequest("GET", "http://internal/api/auth/preferences/sidebar", nil)
	rec := httptest.NewRecorder()

	h.GetSidebarPreferences(rec, req)

	if rec.Code != 401 {
		t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "unauthorized" {
		t.Fatalf("error = %q, want %q", payload.Error, "unauthorized")
	}
}

func TestAuthUpdateSidebarPreferencesWritesWrappedJSONAndAudit(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	svc := &fakeAuthService{
		updatePrefsResult: service.SidebarPreferences{
			PinnedItems: []string{"/grades", "/inventory"},
			RecentItems: []string{"/jadwal"},
		},
	}
	audit := &fakeAuthAuditWriter{}
	h := NewAuth(svc, audit)

	body := bytes.NewBufferString(`{"pinned_items":["/grades","/inventory"],"recent_items":["/jadwal"]}`)
	req := httptest.NewRequest("PATCH", "http://internal/api/auth/preferences/sidebar", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.UpdateSidebarPreferences(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	wantUserID := mustUUID(t, userID)
	if svc.lastUserID != wantUserID {
		t.Fatalf("service user id = %#v, want %#v", svc.lastUserID, wantUserID)
	}
	if len(svc.lastUpdatedPrefs.PinnedItems) != 2 || svc.lastUpdatedPrefs.PinnedItems[1] != "/inventory" {
		t.Fatalf("last updated pinned = %#v", svc.lastUpdatedPrefs.PinnedItems)
	}

	var payload struct {
		Data service.SidebarPreferences `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if len(payload.Data.RecentItems) != 1 || payload.Data.RecentItems[0] != "/jadwal" {
		t.Fatalf("RecentItems = %#v", payload.Data.RecentItems)
	}

	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	if audit.entries[0].Action != "AUTH_SIDEBAR_PREFS_UPDATE" {
		t.Fatalf("audit action = %q", audit.entries[0].Action)
	}
}

func TestAuthUpdateSidebarPreferencesRejectsInvalidJSON(t *testing.T) {
	h := NewAuth(&fakeAuthService{}, nil)

	req := httptest.NewRequest("PATCH", "http://internal/api/auth/preferences/sidebar", bytes.NewBufferString(`{"pinned_items":`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withAuthClaims(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec := httptest.NewRecorder()

	h.UpdateSidebarPreferences(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "invalid json" {
		t.Fatalf("error = %q, want %q", payload.Error, "invalid json")
	}
}

func TestAuthUpdateSidebarPreferencesMapsBadRequest(t *testing.T) {
	svc := &fakeAuthService{
		updatePrefsErr: domain.ErrBadRequest,
	}
	h := NewAuth(svc, nil)

	req := httptest.NewRequest("PATCH", "http://internal/api/auth/preferences/sidebar", bytes.NewBufferString(`{"pinned_items":["grades"]}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withAuthClaims(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec := httptest.NewRecorder()

	h.UpdateSidebarPreferences(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "invalid sidebar preferences" {
		t.Fatalf("error = %q, want %q", payload.Error, "invalid sidebar preferences")
	}
}

func TestAuthUpdateSidebarPreferencesUnauthorizedWithoutClaims(t *testing.T) {
	h := NewAuth(&fakeAuthService{}, nil)

	req := httptest.NewRequest("PATCH", "http://internal/api/auth/preferences/sidebar", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateSidebarPreferences(rec, req)

	if rec.Code != 401 {
		t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "unauthorized" {
		t.Fatalf("error = %q, want %q", payload.Error, "unauthorized")
	}
}
