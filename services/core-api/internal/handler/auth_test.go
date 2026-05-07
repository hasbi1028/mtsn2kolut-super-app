package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeAuthService struct {
	loginResult        domain.TokenPair
	loginErr           error
	lastLoginUsername  string
	lastLoginPassword  string
	lastLoginMeta      service.SessionMeta
	refreshResult      domain.TokenPair
	refreshErr         error
	lastRefreshToken   string
	lastRefreshMeta    service.SessionMeta
	logoutErr          error
	lastLogoutToken    string
	logoutAllErr       error
	lastLogoutAllUser  pgtype.UUID
	accountResult      db.GetUserAccountSummaryRow
	accountErr         error
	lastAccountUserID  pgtype.UUID
	historyResult      []service.AccountChangeHistoryItem
	historyErr         error
	lastHistoryUserID  pgtype.UUID
	lastHistoryLimit   int32
	updateContactErr   error
	lastContactUserID  pgtype.UUID
	lastContactPatch   service.AccountContactPatch
	saveAvatarInput    service.UploadAccountAvatarInput
	saveAvatarErr      error
	deleteAvatarErr    error
	lastAvatarDeleteID pgtype.UUID
	avatarPath         string
	avatarPathFound    bool
	avatarPathErr      error
	lastAvatarPathID   pgtype.UUID
	lastAvatarFilename string
	revokeErr          error
	lastRevokeUserID   pgtype.UUID
	lastRevokeSessID   pgtype.UUID
	sessionListResult  []db.AuthSession
	sessionListErr     error
	updateLabelErr     error
	lastLabelUserID    pgtype.UUID
	lastLabelSessionID pgtype.UUID
	lastDeviceLabel    string
	getPrefsResult     service.SidebarPreferences
	getPrefsErr        error
	updatePrefsResult  service.SidebarPreferences
	updatePrefsErr     error
	lastUserID         pgtype.UUID
	lastUpdatedPrefs   service.SidebarPreferences
	changePasswordErr  error
	lastChangeUser     string
	lastOldPassword    string
	lastNewPassword    string
}

func (f *fakeAuthService) Login(ctx context.Context, username, password string, meta service.SessionMeta) (domain.TokenPair, error) {
	f.lastLoginUsername = username
	f.lastLoginPassword = password
	f.lastLoginMeta = meta
	return f.loginResult, f.loginErr
}

func (f *fakeAuthService) Refresh(ctx context.Context, refreshToken string, meta service.SessionMeta) (domain.TokenPair, error) {
	f.lastRefreshToken = refreshToken
	f.lastRefreshMeta = meta
	return f.refreshResult, f.refreshErr
}

func (f *fakeAuthService) Logout(ctx context.Context, refreshToken string) error {
	f.lastLogoutToken = refreshToken
	return f.logoutErr
}

func (f *fakeAuthService) LogoutAll(ctx context.Context, userID pgtype.UUID) error {
	f.lastLogoutAllUser = userID
	return f.logoutAllErr
}

func (f *fakeAuthService) GetAccount(ctx context.Context, userID pgtype.UUID) (db.GetUserAccountSummaryRow, error) {
	f.lastAccountUserID = userID
	return f.accountResult, f.accountErr
}

func (f *fakeAuthService) ListAccountChangeHistory(ctx context.Context, userID pgtype.UUID, limit int32) ([]service.AccountChangeHistoryItem, error) {
	f.lastHistoryUserID = userID
	f.lastHistoryLimit = limit
	return f.historyResult, f.historyErr
}

func (f *fakeAuthService) UpdateAccountContact(ctx context.Context, userID pgtype.UUID, patch service.AccountContactPatch) (db.GetUserAccountSummaryRow, error) {
	f.lastContactUserID = userID
	f.lastContactPatch = patch
	if f.updateContactErr != nil {
		return db.GetUserAccountSummaryRow{}, f.updateContactErr
	}
	return f.accountResult, nil
}

func (f *fakeAuthService) SaveAccountAvatar(ctx context.Context, input service.UploadAccountAvatarInput) (db.GetUserAccountSummaryRow, error) {
	f.saveAvatarInput = input
	if f.saveAvatarErr != nil {
		return db.GetUserAccountSummaryRow{}, f.saveAvatarErr
	}
	return f.accountResult, nil
}

func (f *fakeAuthService) DeleteAccountAvatar(ctx context.Context, userID pgtype.UUID) (db.GetUserAccountSummaryRow, error) {
	f.lastAvatarDeleteID = userID
	if f.deleteAvatarErr != nil {
		return db.GetUserAccountSummaryRow{}, f.deleteAvatarErr
	}
	return f.accountResult, nil
}

func (f *fakeAuthService) AccountAvatarPath(ctx context.Context, userID pgtype.UUID, filename string) (string, bool, error) {
	f.lastAvatarPathID = userID
	f.lastAvatarFilename = filename
	return f.avatarPath, f.avatarPathFound, f.avatarPathErr
}

func (f *fakeAuthService) ListActiveSessions(ctx context.Context, userID pgtype.UUID) ([]db.AuthSession, error) {
	f.lastUserID = userID
	return f.sessionListResult, f.sessionListErr
}

func (f *fakeAuthService) RevokeSession(ctx context.Context, userID, sessionID pgtype.UUID) error {
	f.lastRevokeUserID = userID
	f.lastRevokeSessID = sessionID
	return f.revokeErr
}

func (f *fakeAuthService) UpdateSessionLabel(ctx context.Context, userID, sessionID pgtype.UUID, deviceLabel string) error {
	f.lastLabelUserID = userID
	f.lastLabelSessionID = sessionID
	f.lastDeviceLabel = deviceLabel
	return f.updateLabelErr
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
	f.lastChangeUser = username
	f.lastOldPassword = oldPassword
	f.lastNewPassword = newPassword
	return f.changePasswordErr
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

func signedToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}

func mustAuditMeta(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var meta map[string]any
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	return meta
}

func multipartAvatarRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/account/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestAuthUserIDBranches(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	got, err := authUserID(jwt.MapClaims{"uid": userID})
	if err != nil {
		t.Fatalf("authUserID(uid fallback) error = %v", err)
	}
	if got != mustUUID(t, userID) {
		t.Fatalf("authUserID(uid fallback) = %v, want %s", got, userID)
	}
	if _, err := authUserID(jwt.MapClaims{}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("authUserID(empty claims) error = %v, want unauthorized", err)
	}
}

func TestAuthAuditHelpersGuardAndWriteEvents(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	sessionID := "22222222-2222-2222-2222-222222222222"
	ctx := context.Background()

	(&Auth{}).auditTokenPair(ctx, "AUTH_LOGIN", domain.TokenPair{AccessToken: "not-a-token"}, nil)
	(&Auth{}).auditClaimsEvent(ctx, "AUTH_LOGOUT_ALL", nil)
	(&Auth{}).auditRefreshTokenEvent(ctx, "AUTH_LOGOUT", " ", nil)
	(&Auth{}).auditWithClaims(ctx, "AUTH_IGNORED", jwt.MapClaims{"uid": userID}, nil)

	audit := &fakeAuthAuditWriter{}
	h := &Auth{audit: audit}
	h.auditTokenPair(ctx, "AUTH_LOGIN", domain.TokenPair{AccessToken: "not-a-token"}, map[string]any{"ignored": true})
	h.auditClaimsEvent(ctx, "AUTH_LOGOUT_ALL", map[string]any{"scope": "none"})
	h.auditRefreshTokenEvent(ctx, "AUTH_LOGOUT", " ", nil)
	h.auditRefreshTokenEvent(ctx, "AUTH_LOGOUT", "not-a-token", nil)
	h.auditRefreshTokenEvent(ctx, "AUTH_LOGOUT", signedToken(t, jwt.MapClaims{"uid": userID, "type": "access"}), nil)
	if len(audit.entries) != 0 {
		t.Fatalf("guarded audit entries = %d, want 0", len(audit.entries))
	}

	h.auditTokenPair(ctx, "AUTH_LOGIN", domain.TokenPair{AccessToken: signedToken(t, jwt.MapClaims{
		"uid":  userID,
		"usr":  "admin",
		"ssid": sessionID,
	})}, map[string]any{"source": "login"})
	if len(audit.entries) != 1 {
		t.Fatalf("auditTokenPair entries = %d, want 1", len(audit.entries))
	}
	if audit.entries[0].Action != "AUTH_LOGIN" || audit.entries[0].EntityID != sessionID || audit.entries[0].UserID != mustUUID(t, userID) {
		t.Fatalf("auditTokenPair entry = %+v, want login session audit", audit.entries[0])
	}
	meta := mustAuditMeta(t, audit.entries[0].Metadata)
	if meta["username"] != "admin" || meta["source"] != "login" {
		t.Fatalf("auditTokenPair metadata = %+v, want username/source", meta)
	}

	claimsCtx := context.WithValue(ctx, api.ClaimsKey, jwt.MapClaims{
		"uid": userID,
		"usr": "admin",
	})
	h.auditClaimsEvent(claimsCtx, "AUTH_LOGOUT_ALL", map[string]any{"scope": "all_sessions"})
	if len(audit.entries) != 2 {
		t.Fatalf("auditClaimsEvent entries = %d, want 2", len(audit.entries))
	}
	if audit.entries[1].EntityID != userID {
		t.Fatalf("auditClaimsEvent entity id = %q, want uid fallback", audit.entries[1].EntityID)
	}
	meta = mustAuditMeta(t, audit.entries[1].Metadata)
	if meta["scope"] != "all_sessions" {
		t.Fatalf("auditClaimsEvent metadata = %+v, want scope", meta)
	}

	h.auditRefreshTokenEvent(ctx, "AUTH_LOGOUT", signedToken(t, jwt.MapClaims{
		"uid":  userID,
		"usr":  "admin",
		"ssid": sessionID,
		"type": "refresh",
	}), map[string]any{"scope": "single_session"})
	if len(audit.entries) != 3 {
		t.Fatalf("auditRefreshTokenEvent entries = %d, want 3", len(audit.entries))
	}
	if audit.entries[2].Action != "AUTH_LOGOUT" || audit.entries[2].EntityID != sessionID {
		t.Fatalf("auditRefreshTokenEvent entry = %+v, want logout session audit", audit.entries[2])
	}

	h.auditWithClaims(ctx, "AUTH_BAD_META", jwt.MapClaims{"uid": userID}, map[string]any{"bad": func() {}})
	if len(audit.entries) != 3 {
		t.Fatalf("auditWithClaims(bad metadata) entries = %d, want unchanged 3", len(audit.entries))
	}
}

func TestAuthLoginWritesAuditAndForwardsSessionMeta(t *testing.T) {
	svc := &fakeAuthService{
		loginResult: domain.TokenPair{
			AccessToken:  signedToken(t, jwt.MapClaims{"uid": "11111111-1111-1111-1111-111111111111", "usr": "admin", "ssid": "22222222-2222-2222-2222-222222222222"}),
			RefreshToken: "refresh-token",
		},
	}
	audit := &fakeAuthAuditWriter{}
	h := NewAuth(svc, audit)

	req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/login", bytes.NewBufferString(`{"username":"admin","password":"secretpass"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-IP", "203.0.113.10")
	req.Header.Set("X-Client-User-Agent", "MTsN2 Test Client/1.0")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastLoginUsername != "admin" || svc.lastLoginPassword != "secretpass" {
		t.Fatalf("login credentials forwarded = (%q, %q)", svc.lastLoginUsername, svc.lastLoginPassword)
	}
	if svc.lastLoginMeta.IPAddress != "192.0.2.1" {
		t.Fatalf("login ip = %q, want remote address for untrusted client", svc.lastLoginMeta.IPAddress)
	}
	if svc.lastLoginMeta.UserAgent != "MTsN2 Test Client/1.0" {
		t.Fatalf("login user agent = %q", svc.lastLoginMeta.UserAgent)
	}
	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	if audit.entries[0].Action != "AUTH_LOGIN" {
		t.Fatalf("audit action = %q", audit.entries[0].Action)
	}
	meta := mustAuditMeta(t, audit.entries[0].Metadata)
	if meta["username"] != "admin" {
		t.Fatalf("audit username = %#v, want %q", meta["username"], "admin")
	}
	if meta["session_id"] != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("audit session_id = %#v", meta["session_id"])
	}
}

func TestAuthRefreshMapsSuspendedAndWritesAudit(t *testing.T) {
	suspended := &fakeAuthService{refreshErr: domain.ErrSuspended}
	h := NewAuth(suspended, nil)
	req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/refresh", bytes.NewBufferString(`{"refresh_token":"r1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Refresh(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("suspended status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	svc := &fakeAuthService{
		refreshResult: domain.TokenPair{
			AccessToken:  signedToken(t, jwt.MapClaims{"uid": "11111111-1111-1111-1111-111111111111", "usr": "admin", "ssid": "33333333-3333-3333-3333-333333333333"}),
			RefreshToken: "next-refresh",
		},
	}
	audit := &fakeAuthAuditWriter{}
	h = NewAuth(svc, audit)
	req = httptest.NewRequest(http.MethodPost, "http://internal/api/auth/refresh", bytes.NewBufferString(`{"refresh_token":"r1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.11, 10.0.0.5")
	req.Header.Set("User-Agent", "browser/2.0")
	rec = httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastRefreshToken != "r1" {
		t.Fatalf("refresh token forwarded = %q", svc.lastRefreshToken)
	}
	if svc.lastRefreshMeta.IPAddress != "192.0.2.1" {
		t.Fatalf("refresh ip = %q, want remote address for untrusted client", svc.lastRefreshMeta.IPAddress)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "AUTH_REFRESH" {
		t.Fatalf("audit entries = %#v", audit.entries)
	}
}

func TestAuthLogoutWritesAuditFromRefreshTokenClaims(t *testing.T) {
	svc := &fakeAuthService{}
	audit := &fakeAuthAuditWriter{}
	h := NewAuth(svc, audit)
	refreshToken := signedToken(t, jwt.MapClaims{
		"uid":  "11111111-1111-1111-1111-111111111111",
		"usr":  "admin",
		"ssid": "44444444-4444-4444-4444-444444444444",
		"type": "refresh",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/logout", bytes.NewBufferString(`{"refresh_token":"`+refreshToken+`"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastLogoutToken != refreshToken {
		t.Fatalf("logout token forwarded = %q", svc.lastLogoutToken)
	}
	if len(audit.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(audit.entries))
	}
	if audit.entries[0].Action != "AUTH_LOGOUT" {
		t.Fatalf("audit action = %q", audit.entries[0].Action)
	}
	meta := mustAuditMeta(t, audit.entries[0].Metadata)
	if meta["scope"] != "single_session" {
		t.Fatalf("audit scope = %#v", meta["scope"])
	}
	if meta["session_id"] != "44444444-4444-4444-4444-444444444444" {
		t.Fatalf("audit session_id = %#v", meta["session_id"])
	}
}

func TestAuthLogoutAllUnauthorizedWithoutClaimsAndAuditsWithClaims(t *testing.T) {
	h := NewAuth(&fakeAuthService{}, nil)
	req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/logout-all", nil)
	rec := httptest.NewRecorder()
	h.LogoutAll(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}

	userID := "11111111-1111-1111-1111-111111111111"
	svc := &fakeAuthService{}
	audit := &fakeAuthAuditWriter{}
	h = NewAuth(svc, audit)
	req = httptest.NewRequest(http.MethodPost, "http://internal/api/auth/logout-all", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec = httptest.NewRecorder()
	h.LogoutAll(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastLogoutAllUser != mustUUID(t, userID) {
		t.Fatalf("logout-all user id = %#v", svc.lastLogoutAllUser)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "AUTH_LOGOUT_ALL" {
		t.Fatalf("audit entries = %#v", audit.entries)
	}
}

func TestAuthGetAccountForwardsCurrentUserAndSanitizesResponse(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	employeeID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	lastLogin := pgtype.Timestamptz{}
	_ = lastLogin.Scan(time.Date(2026, time.May, 6, 7, 30, 0, 0, time.UTC))
	svc := &fakeAuthService{
		accountResult: db.GetUserAccountSummaryRow{
			ID:             mustUUID(t, userID),
			Username:       "guru.ipa",
			DisplayName:    "Guru IPA",
			Roles:          []byte(`["guru","staf"]`),
			EmployeeID:     employeeID,
			ProfileType:    "employee",
			ProfileNama:    "Nama Pegawai",
			PhotoUrl:       "/api/auth/account/avatar/avatar_guru.png",
			ContactPhone:   "081234",
			ContactEmail:   "guru@example.id",
			ContactAddress: "Kolaka Utara",
			IsActive:       true,
			LastLoginAt:    lastLogin,
		},
	}
	h := NewAuth(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "http://internal/api/auth/account", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.GetAccount(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastAccountUserID != mustUUID(t, userID) {
		t.Fatalf("account user id = %v, want %s", svc.lastAccountUserID, userID)
	}
	body := rec.Body.String()
	for _, expected := range []string{"guru.ipa", "Guru IPA", "Nama Pegawai", "employee", "/api/auth/account/avatar/avatar_guru.png", `"avatar_url":"/api/auth/account/avatar/avatar_guru.png"`, "081234", "guru@example.id", "Kolaka Utara", "2026-05-06T07:30:00Z"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("GetAccount body = %s, want %q", body, expected)
		}
	}
	secretFreeBody := strings.ReplaceAll(body, "must_change_password", "")
	secretFreeBody = strings.ReplaceAll(secretFreeBody, "password_changed_at", "")
	if strings.Contains(secretFreeBody, "password") || strings.Contains(body, "auth_version") {
		t.Fatalf("GetAccount body = %s, must not expose secret or auth internals", body)
	}
}

func TestAuthGetAccountMapsUnauthorizedAndInternalError(t *testing.T) {
	h := NewAuth(&fakeAuthService{}, nil)
	rec := httptest.NewRecorder()
	h.GetAccount(rec, httptest.NewRequest(http.MethodGet, "http://internal/api/auth/account", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GetAccount(no claims) status = %d, want 401", rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "http://internal/api/auth/account", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "not-a-uuid"))
	rec = httptest.NewRecorder()
	h.GetAccount(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GetAccount(bad claims) status = %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "http://internal/api/auth/account", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{accountErr: domain.ErrUnauthorized}, nil).GetAccount(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GetAccount(unauthorized service) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{accountErr: context.Canceled}, nil).GetAccount(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GetAccount(internal) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthGetAccountChangeHistoryForwardsCurrentUserAndSanitizesResponse(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	createdAt := pgtype.Timestamptz{}
	_ = createdAt.Scan(time.Date(2026, time.May, 7, 8, 0, 0, 0, time.UTC))
	svc := &fakeAuthService{
		historyResult: []service.AccountChangeHistoryItem{
			{
				Action:           "change_request_approved",
				FieldKey:         "nama",
				Status:           "approved",
				CreatedAt:        createdAt,
				ReviewerUsername: "admin",
				ReviewNote:       "Sesuai dokumen",
			},
			{
				Action:    "avatar_delete",
				FieldKey:  "avatar",
				Status:    "completed",
				CreatedAt: createdAt,
			},
		},
	}
	h := NewAuth(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "http://internal/api/auth/account/change-history?per_page=2&user_id=22222222-2222-2222-2222-222222222222", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.GetAccountChangeHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastHistoryUserID != mustUUID(t, userID) || svc.lastHistoryLimit != 2 {
		t.Fatalf("history args = %v/%d, want current user and per_page", svc.lastHistoryUserID, svc.lastHistoryLimit)
	}
	body := rec.Body.String()
	for _, expected := range []string{"change_request_approved", "nama", "approved", "admin", "Sesuai dokumen", "avatar_delete", "2026-05-07T08:00:00Z"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("GetAccountChangeHistory body = %s, want %q", body, expected)
		}
	}
	for _, forbidden := range []string{"token", "password", "metadata", "22222222-2222-2222-2222-222222222222"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("GetAccountChangeHistory body = %s, must not expose %q", body, forbidden)
		}
	}
}

func TestAuthGetAccountChangeHistoryMapsUnauthorizedAndInternalError(t *testing.T) {
	h := NewAuth(&fakeAuthService{}, nil)
	rec := httptest.NewRecorder()
	h.GetAccountChangeHistory(rec, httptest.NewRequest(http.MethodGet, "http://internal/api/auth/account/change-history", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GetAccountChangeHistory(no claims) status = %d, want 401", rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "http://internal/api/auth/account/change-history", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{historyErr: domain.ErrUnauthorized}, nil).GetAccountChangeHistory(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GetAccountChangeHistory(unauthorized service) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{historyErr: context.Canceled}, nil).GetAccountChangeHistory(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GetAccountChangeHistory(internal) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthUpdateAccountContactForwardsCurrentUserAndAudits(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	svc := &fakeAuthService{
		accountResult: db.GetUserAccountSummaryRow{
			ID:             mustUUID(t, userID),
			Username:       "guru.ipa",
			DisplayName:    "Guru IPA",
			ProfileType:    "employee",
			ProfileNama:    "Nama Pegawai",
			ContactPhone:   "081234567890",
			ContactEmail:   "guru@example.id",
			ContactAddress: "Kolaka Utara",
			Roles:          []byte(`["guru"]`),
		},
	}
	audit := &fakeAuthAuditWriter{}
	h := NewAuth(svc, audit)
	req := httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/account/contact", bytes.NewBufferString(`{"phone":" 081234567890 ","email":"guru@example.id","address":"Kolaka Utara"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.UpdateAccountContact(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastContactUserID != mustUUID(t, userID) {
		t.Fatalf("contact user id = %v, want %s", svc.lastContactUserID, userID)
	}
	if svc.lastContactPatch.Phone == nil || *svc.lastContactPatch.Phone != " 081234567890 " {
		t.Fatalf("phone patch = %#v, want raw phone", svc.lastContactPatch.Phone)
	}
	if !strings.Contains(rec.Body.String(), `"editable_fields":["phone","email","address"]`) {
		t.Fatalf("body = %s, want employee editable fields", rec.Body.String())
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "AUTH_ACCOUNT_CONTACT_UPDATE" {
		t.Fatalf("audit entries = %#v", audit.entries)
	}
	meta := mustAuditMeta(t, audit.entries[0].Metadata)
	if meta["profile_type"] != "employee" {
		t.Fatalf("audit metadata = %+v, want profile type", meta)
	}
}

func TestAuthUpdateAccountContactRejectsOfficialFieldsAndMapsErrors(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	req := httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/account/contact", bytes.NewBufferString(`{"username":"baru"}`))
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	NewAuth(&fakeAuthService{}, nil).UpdateAccountContact(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UpdateAccountContact(official field) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/account/contact", bytes.NewBufferString(`{"phone":"0812"}`))
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{updateContactErr: domain.ErrForbidden}, nil).UpdateAccountContact(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("UpdateAccountContact(forbidden) status = %d, want 403", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/account/contact", bytes.NewBufferString(`{"email":"bad"}`))
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{updateContactErr: domain.ErrBadRequest}, nil).UpdateAccountContact(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UpdateAccountContact(bad request) status = %d, want 400", rec.Code)
	}
}

func TestAuthAccountAvatarUploadDeleteAndFile(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	sessionID := "22222222-2222-2222-2222-222222222222"
	svc := &fakeAuthService{
		accountResult: db.GetUserAccountSummaryRow{
			ID:          mustUUID(t, userID),
			Username:    "guru.ipa",
			DisplayName: "Guru IPA",
			ProfileType: "employee",
			PhotoUrl:    "/api/auth/account/avatar/avatar_baru.png",
			Roles:       []byte(`["guru"]`),
		},
	}
	audit := &fakeAuthAuditWriter{}
	h := NewAuth(svc, audit)
	claimsCtx := context.WithValue(context.Background(), api.ClaimsKey, jwt.MapClaims{
		"sub":  userID,
		"uid":  userID,
		"usr":  "guru.ipa",
		"ssid": sessionID,
	})
	req := multipartAvatarRequest(t, "foto.png", []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"))
	req = req.WithContext(claimsCtx)
	rec := httptest.NewRecorder()

	h.UploadAccountAvatar(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("UploadAccountAvatar() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.saveAvatarInput.UserID != mustUUID(t, userID) || svc.saveAvatarInput.MimeType != "image/png" || svc.saveAvatarInput.Ext != ".png" {
		t.Fatalf("avatar input = %+v, want owned png upload", svc.saveAvatarInput)
	}
	if !strings.Contains(rec.Body.String(), `"avatar_url":"/api/auth/account/avatar/avatar_baru.png"`) {
		t.Fatalf("body = %s, want avatar_url", rec.Body.String())
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "AUTH_ACCOUNT_AVATAR_UPDATE" {
		t.Fatalf("upload audit entries = %#v", audit.entries)
	}

	req = httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/account/avatar", nil).WithContext(claimsCtx)
	rec = httptest.NewRecorder()
	h.DeleteAccountAvatar(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("DeleteAccountAvatar() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.lastAvatarDeleteID != mustUUID(t, userID) {
		t.Fatalf("delete user id = %v, want %s", svc.lastAvatarDeleteID, userID)
	}
	if len(audit.entries) != 2 || audit.entries[1].Action != "AUTH_ACCOUNT_AVATAR_DELETE" {
		t.Fatalf("delete audit entries = %#v", audit.entries)
	}

	file, err := os.CreateTemp(t.TempDir(), "avatar-*.png")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	if _, err := file.WriteString("avatar-bytes"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	svc.avatarPath = file.Name()
	svc.avatarPathFound = true
	req = httptest.NewRequest(http.MethodGet, "http://internal/api/auth/account/avatar/avatar_baru.png", nil).WithContext(claimsCtx)
	req = withRouteParam(req, "filename", "avatar_baru.png")
	rec = httptest.NewRecorder()
	h.AccountAvatarFile(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "avatar-bytes" {
		t.Fatalf("AccountAvatarFile() status/body = %d/%q, want file bytes", rec.Code, rec.Body.String())
	}
	if svc.lastAvatarFilename != "avatar_baru.png" || rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("file filename/headers = %q/%q, want scoped nosniff file", svc.lastAvatarFilename, rec.Header().Get("X-Content-Type-Options"))
	}
}

func TestAuthAccountAvatarRejectsUnsupportedUpload(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	req := multipartAvatarRequest(t, "foto.gif", []byte("GIF89a\x01\x00\x01\x00"))
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	NewAuth(&fakeAuthService{}, nil).UploadAccountAvatar(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadAccountAvatar(gif) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthRevokeSessionMapsNotFound(t *testing.T) {
	h := NewAuth(&fakeAuthService{revokeErr: domain.ErrNotFound}, nil)
	req := httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/sessions/33333333-3333-3333-3333-333333333333", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec := httptest.NewRecorder()
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "33333333-3333-3333-3333-333333333333")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	h.RevokeSession(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestAuthListSessionsForwardsUserID(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	sessionID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	svc := &fakeAuthService{
		sessionListResult: []db.AuthSession{{ID: sessionID, DeviceLabel: "Laptop TU", RefreshTokenHash: "secret-hash"}},
	}
	h := NewAuth(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "http://internal/api/auth/sessions", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.ListSessions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastUserID != mustUUID(t, userID) {
		t.Fatalf("ListActiveSessions user id = %v, want %s", svc.lastUserID, userID)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("Laptop TU")) {
		t.Fatalf("ListSessions body = %s, want device label", rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("refresh_token_hash")) || bytes.Contains(rec.Body.Bytes(), []byte("secret-hash")) {
		t.Fatalf("ListSessions body = %s, must not expose refresh token hash", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "http://internal/api/auth/sessions", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{"uid": userID}))
	rec = httptest.NewRecorder()
	h.ListSessions(rec, req)
	if rec.Code != http.StatusOK || svc.lastUserID != mustUUID(t, userID) {
		t.Fatalf("ListSessions(uid fallback) status/user = %d/%v, want 200/%s", rec.Code, svc.lastUserID, userID)
	}
}

func TestAuthListSessionsMapsUnauthorizedAndInternalError(t *testing.T) {
	h := NewAuth(&fakeAuthService{}, nil)
	rec := httptest.NewRecorder()
	h.ListSessions(rec, httptest.NewRequest(http.MethodGet, "http://internal/api/auth/sessions", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ListSessions(no claims) status = %d, want 401", rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "http://internal/api/auth/sessions", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{"sub": "not-a-uuid"}))
	rec = httptest.NewRecorder()
	h.ListSessions(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ListSessions(invalid claims) status = %d, want 401", rec.Code)
	}

	h = NewAuth(&fakeAuthService{sessionListErr: context.Canceled}, nil)
	req = httptest.NewRequest(http.MethodGet, "http://internal/api/auth/sessions", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec = httptest.NewRecorder()
	h.ListSessions(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("ListSessions(error) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthUpdateSessionLabelForwardsAndAudits(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	sessionID := "33333333-3333-3333-3333-333333333333"
	svc := &fakeAuthService{}
	audit := &fakeAuthAuditWriter{}
	h := NewAuth(svc, audit)
	req := httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/sessions/"+sessionID, bytes.NewBufferString(`{"device_label":" Laptop TU "}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", sessionID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	rec := httptest.NewRecorder()

	h.UpdateSessionLabel(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastLabelUserID != mustUUID(t, userID) || svc.lastLabelSessionID != mustUUID(t, sessionID) || svc.lastDeviceLabel != " Laptop TU " {
		t.Fatalf("UpdateSessionLabel args = %v/%v/%q, want user/session/raw label", svc.lastLabelUserID, svc.lastLabelSessionID, svc.lastDeviceLabel)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "AUTH_SESSION_RENAME" {
		t.Fatalf("audit entries = %#v", audit.entries)
	}
	meta := mustAuditMeta(t, audit.entries[0].Metadata)
	if meta["device_label"] != "Laptop TU" || meta["renamed_session_id"] != sessionID {
		t.Fatalf("audit metadata = %+v, want trimmed label/session id", meta)
	}
}

func TestAuthUpdateSessionLabelMapsValidationAndServiceErrors(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	sessionID := "33333333-3333-3333-3333-333333333333"
	withRoute := func(req *http.Request, id string) *http.Request {
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", id)
		return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	}
	tests := []struct {
		name string
		svc  *fakeAuthService
		req  *http.Request
		want int
	}{
		{
			name: "unauthorized",
			svc:  &fakeAuthService{},
			req:  withRoute(httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/sessions/"+sessionID, bytes.NewBufferString(`{}`)), sessionID),
			want: http.StatusUnauthorized,
		},
		{
			name: "invalid claims",
			svc:  &fakeAuthService{},
			req:  withRoute(httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/sessions/"+sessionID, bytes.NewBufferString(`{"device_label":"HP"}`)).WithContext(withAuthClaims(context.Background(), "not-a-uuid")), sessionID),
			want: http.StatusUnauthorized,
		},
		{
			name: "invalid id",
			svc:  &fakeAuthService{},
			req:  withRoute(httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/sessions/bad", bytes.NewBufferString(`{}`)).WithContext(withAuthClaims(context.Background(), userID)), "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "invalid json",
			svc:  &fakeAuthService{},
			req:  withRoute(httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/sessions/"+sessionID, bytes.NewBufferString(`{`)).WithContext(withAuthClaims(context.Background(), userID)), sessionID),
			want: http.StatusBadRequest,
		},
		{
			name: "bad request",
			svc:  &fakeAuthService{updateLabelErr: domain.ErrBadRequest},
			req:  withRoute(httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/sessions/"+sessionID, bytes.NewBufferString(`{"device_label":""}`)).WithContext(withAuthClaims(context.Background(), userID)), sessionID),
			want: http.StatusBadRequest,
		},
		{
			name: "not found",
			svc:  &fakeAuthService{updateLabelErr: domain.ErrNotFound},
			req:  withRoute(httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/sessions/"+sessionID, bytes.NewBufferString(`{"device_label":"HP"}`)).WithContext(withAuthClaims(context.Background(), userID)), sessionID),
			want: http.StatusNotFound,
		},
		{
			name: "internal",
			svc:  &fakeAuthService{updateLabelErr: context.Canceled},
			req:  withRoute(httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/sessions/"+sessionID, bytes.NewBufferString(`{"device_label":"HP"}`)).WithContext(withAuthClaims(context.Background(), userID)), sessionID),
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewAuth(tt.svc, nil).UpdateSessionLabel(rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestAuthChangePasswordWritesAuditAndUsesJWTUsername(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	svc := &fakeAuthService{}
	audit := &fakeAuthAuditWriter{}
	h := NewAuth(svc, audit)

	req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/change-password", bytes.NewBufferString(`{"username":"ignored","old_password":"oldpass","new_password":"newpass123"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"sub": userID,
		"uid": userID,
		"usr": "admin",
	}))
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if svc.lastChangeUser != "admin" {
		t.Fatalf("change password username = %q, want %q", svc.lastChangeUser, "admin")
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "AUTH_PASSWORD_CHANGE" {
		t.Fatalf("audit entries = %#v", audit.entries)
	}
}

func TestAuthChangePasswordMapsSuspendedAndWeakPassword(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		code int
		msg  string
	}{
		{name: "suspended", err: domain.ErrSuspended, code: http.StatusForbidden, msg: "account is suspended"},
		{name: "weak_password", err: domain.ErrWeakPassword, code: http.StatusBadRequest, msg: "password baru minimal 8 karakter dan maksimal 72 karakter, tidak boleh sama dengan username, dan tidak boleh hanya angka"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAuth(&fakeAuthService{changePasswordErr: tt.err}, nil)
			req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/change-password", bytes.NewBufferString(`{"username":"admin","old_password":"old","new_password":"new"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.ChangePassword(rec, req)

			if rec.Code != tt.code {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.code, rec.Body.String())
			}
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json unmarshal failed: %v", err)
			}
			if payload.Error != tt.msg {
				t.Fatalf("error = %q, want %q", payload.Error, tt.msg)
			}
		})
	}
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

func TestAuthLoginMapsValidationAndServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		svc  *fakeAuthService
		body string
		want int
	}{
		{name: "invalid json", svc: &fakeAuthService{}, body: `{`, want: http.StatusBadRequest},
		{name: "unauthorized", svc: &fakeAuthService{loginErr: domain.ErrUnauthorized}, body: `{"username":"admin","password":"bad"}`, want: http.StatusUnauthorized},
		{name: "suspended", svc: &fakeAuthService{loginErr: domain.ErrSuspended}, body: `{"username":"admin","password":"secret"}`, want: http.StatusForbidden},
		{name: "internal", svc: &fakeAuthService{loginErr: context.Canceled}, body: `{"username":"admin","password":"secret"}`, want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewAuth(tt.svc, nil).Login(rec, httptest.NewRequest(http.MethodPost, "http://internal/api/auth/login", bytes.NewBufferString(tt.body)))
			if rec.Code != tt.want {
				t.Fatalf("Login() status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestAuthRefreshMapsValidationUnauthorizedAndInternal(t *testing.T) {
	tests := []struct {
		name string
		svc  *fakeAuthService
		body string
		want int
	}{
		{name: "invalid json", svc: &fakeAuthService{}, body: `{`, want: http.StatusBadRequest},
		{name: "missing token", svc: &fakeAuthService{}, body: `{}`, want: http.StatusBadRequest},
		{name: "unauthorized", svc: &fakeAuthService{refreshErr: domain.ErrUnauthorized}, body: `{"refresh_token":"r1"}`, want: http.StatusUnauthorized},
		{name: "internal", svc: &fakeAuthService{refreshErr: context.Canceled}, body: `{"refresh_token":"r1"}`, want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewAuth(tt.svc, nil).Refresh(rec, httptest.NewRequest(http.MethodPost, "http://internal/api/auth/refresh", bytes.NewBufferString(tt.body)))
			if rec.Code != tt.want {
				t.Fatalf("Refresh() status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestAuthLogoutAndLogoutAllMapServiceErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	NewAuth(&fakeAuthService{logoutErr: context.Canceled}, nil).Logout(rec, httptest.NewRequest(http.MethodPost, "http://internal/api/auth/logout", bytes.NewBufferString(`{"refresh_token":"r1"}`)))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Logout(error) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/logout-all", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "not-a-uuid"))
	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{}, nil).LogoutAll(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("LogoutAll(invalid claims) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "http://internal/api/auth/logout-all", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{logoutAllErr: domain.ErrUnauthorized}, nil).LogoutAll(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("LogoutAll(unauthorized service) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "http://internal/api/auth/logout-all", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "11111111-1111-1111-1111-111111111111"))
	rec = httptest.NewRecorder()
	NewAuth(&fakeAuthService{logoutAllErr: context.Canceled}, nil).LogoutAll(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("LogoutAll(internal) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthRevokeSessionValidationSuccessAndInternal(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	sessionID := "33333333-3333-3333-3333-333333333333"
	withRoute := func(req *http.Request, id string) *http.Request {
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", id)
		return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	}

	rec := httptest.NewRecorder()
	NewAuth(&fakeAuthService{}, nil).RevokeSession(rec, withRoute(httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/sessions/"+sessionID, nil), sessionID))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("RevokeSession(no claims) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req := withRoute(httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/sessions/"+sessionID, nil).WithContext(withAuthClaims(context.Background(), "not-a-uuid")), sessionID)
	NewAuth(&fakeAuthService{}, nil).RevokeSession(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("RevokeSession(invalid claims) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRoute(httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/sessions/bad", nil).WithContext(withAuthClaims(context.Background(), userID)), "bad")
	NewAuth(&fakeAuthService{}, nil).RevokeSession(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("RevokeSession(bad id) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	audit := &fakeAuthAuditWriter{}
	svc := &fakeAuthService{}
	rec = httptest.NewRecorder()
	req = withRoute(httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/sessions/"+sessionID, nil).WithContext(withAuthClaims(context.Background(), userID)), sessionID)
	NewAuth(svc, audit).RevokeSession(rec, req)
	if rec.Code != http.StatusOK || svc.lastRevokeUserID != mustUUID(t, userID) || svc.lastRevokeSessID != mustUUID(t, sessionID) {
		t.Fatalf("RevokeSession(success) status/args = %d/%v/%v", rec.Code, svc.lastRevokeUserID, svc.lastRevokeSessID)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "AUTH_SESSION_REVOKE" {
		t.Fatalf("RevokeSession audit = %#v, want revoke audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	req = withRoute(httptest.NewRequest(http.MethodDelete, "http://internal/api/auth/sessions/"+sessionID, nil).WithContext(withAuthClaims(context.Background(), userID)), sessionID)
	NewAuth(&fakeAuthService{revokeErr: context.Canceled}, nil).RevokeSession(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("RevokeSession(internal) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthPreferencesAndChangePasswordMapRemainingErrors(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://internal/api/auth/preferences/sidebar", nil)
	req = req.WithContext(withAuthClaims(req.Context(), "bad"))
	NewAuth(&fakeAuthService{}, nil).GetSidebarPreferences(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GetSidebarPreferences(bad claims) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "http://internal/api/auth/preferences/sidebar", nil)
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	NewAuth(&fakeAuthService{getPrefsErr: context.Canceled}, nil).GetSidebarPreferences(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GetSidebarPreferences(error) status = %d, want 500", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/preferences/sidebar", bytes.NewBufferString(`{}`))
	req = req.WithContext(withAuthClaims(req.Context(), "bad"))
	NewAuth(&fakeAuthService{}, nil).UpdateSidebarPreferences(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("UpdateSidebarPreferences(bad claims) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "http://internal/api/auth/preferences/sidebar", bytes.NewBufferString(`{}`))
	req = req.WithContext(withAuthClaims(req.Context(), userID))
	NewAuth(&fakeAuthService{updatePrefsErr: context.Canceled}, nil).UpdateSidebarPreferences(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("UpdateSidebarPreferences(error) status = %d, want 500", rec.Code)
	}

	for _, tt := range []struct {
		name string
		body string
		ctx  context.Context
		err  error
		want int
	}{
		{name: "invalid json", body: `{`, want: http.StatusBadRequest},
		{name: "missing username", body: `{"old_password":"old","new_password":"newpass123"}`, want: http.StatusBadRequest},
		{name: "missing new password", body: `{"username":"admin","old_password":"old"}`, want: http.StatusBadRequest},
		{name: "unauthorized", body: `{"username":"admin","old_password":"bad","new_password":"newpass123"}`, err: domain.ErrUnauthorized, want: http.StatusUnauthorized},
		{name: "internal", body: `{"username":"admin","old_password":"old","new_password":"newpass123"}`, err: context.Canceled, want: http.StatusInternalServerError},
	} {
		t.Run("ChangePassword "+tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "http://internal/api/auth/change-password", bytes.NewBufferString(tt.body))
			if tt.ctx != nil {
				req = req.WithContext(tt.ctx)
			}
			NewAuth(&fakeAuthService{changePasswordErr: tt.err}, nil).ChangePassword(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("ChangePassword() status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
