package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestAuthRolesBytesHandlesSupportedAndUnsupportedValues(t *testing.T) {
	if got := authRolesBytes(nil); got != nil {
		t.Fatalf("authRolesBytes(nil) = %q, want nil", got)
	}

	raw := []byte(`["admin"]`)
	if got := authRolesBytes(raw); string(got) != string(raw) {
		t.Fatalf("authRolesBytes([]byte) = %q, want %q", got, raw)
	}
	if got := authRolesBytes(`["guru"]`); string(got) != `["guru"]` {
		t.Fatalf("authRolesBytes(string) = %q, want legacy json string", got)
	}

	got := authRolesBytes([]string{"admin", "guru"})
	var decoded []string
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("authRolesBytes([]string) returned invalid json %q: %v", got, err)
	}
	if len(decoded) != 2 || decoded[0] != "admin" || decoded[1] != "guru" {
		t.Fatalf("authRolesBytes([]string) decoded = %#v, want admin/guru", decoded)
	}

	if got := authRolesBytes(func() {}); got != nil {
		t.Fatalf("authRolesBytes(unmarshalable) = %q, want nil", got)
	}
}

func TestNormalizeAccountChangeHistoryHelpersAndFiltering(t *testing.T) {
	if got := normalizeAccountChangeHistoryLimit(0); got != 30 {
		t.Fatalf("normalizeAccountChangeHistoryLimit(0) = %d, want 30", got)
	}
	if got := normalizeAccountChangeHistoryLimit(99); got != 50 {
		t.Fatalf("normalizeAccountChangeHistoryLimit(99) = %d, want 50", got)
	}
	if got := normalizeAccountChangeHistoryLimit(12); got != 12 {
		t.Fatalf("normalizeAccountChangeHistoryLimit(12) = %d, want 12", got)
	}

	for _, tt := range []struct {
		action string
		want   string
	}{
		{action: " AUTH_ACCOUNT_CONTACT_UPDATE ", want: "contact_update"},
		{action: "AUTH_ACCOUNT_AVATAR_UPDATE", want: "avatar_update"},
		{action: "AUTH_ACCOUNT_AVATAR_DELETE", want: "avatar_delete"},
		{action: "ACCOUNT_CHANGE_REQUEST_CREATED", want: "change_request_created"},
		{action: "ACCOUNT_CHANGE_REQUEST_CANCELLED", want: "change_request_cancelled"},
		{action: "ACCOUNT_CHANGE_REQUEST_APPROVED", want: "change_request_approved"},
		{action: "ACCOUNT_CHANGE_REQUEST_REJECTED", want: "change_request_rejected"},
		{action: "UNKNOWN", want: ""},
	} {
		if got := normalizeAccountChangeHistoryAction(tt.action); got != tt.want {
			t.Fatalf("normalizeAccountChangeHistoryAction(%q) = %q, want %q", tt.action, got, tt.want)
		}
	}

	created := pgtype.Timestamptz{Time: time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC), Valid: true}
	item, ok := accountChangeHistoryItemFromRow(db.ListOwnAccountChangeHistoryRow{
		Action:              "AUTH_ACCOUNT_AVATAR_DELETE",
		FieldKey:            " Nama ",
		Status:              " odd-status ",
		CreatedAt:           created,
		ReviewerUsername:    " reviewer ",
		ReviewerDisplayName: " Reviewer Name ",
		ReviewNote:          " done ",
	})
	if !ok {
		t.Fatal("accountChangeHistoryItemFromRow(valid) ok = false, want true")
	}
	if item.Action != "avatar_delete" || item.FieldKey != "nama" || item.Status != "completed" || item.ReviewerUsername != "reviewer" || item.ReviewerDisplayName != "Reviewer Name" || item.ReviewNote != "done" {
		t.Fatalf("accountChangeHistoryItemFromRow(valid) = %+v, want normalized/truncated fields", item)
	}
	if _, ok := accountChangeHistoryItemFromRow(db.ListOwnAccountChangeHistoryRow{Action: "unknown"}); ok {
		t.Fatal("accountChangeHistoryItemFromRow(unknown action) ok = true, want false")
	}
}

func TestIssueTokenPairUsesLegacyRolesFallbackAndStoresSession(t *testing.T) {
	store := newFakeStore()
	userID := documentCycleTestUUID(31)
	studentID := documentCycleTestUUID(32)
	store.userPermissions[userID] = []string{" read:data ", "read:data", "write:data", ""}
	svc := &Auth{q: store, jwtSecret: []byte("secret")}

	pair, err := svc.issueTokenPair(context.Background(), authUserRecord{
		ID:                 userID,
		Username:           "student1",
		StudentID:          studentID,
		IsActive:           true,
		AuthVersion:        7,
		MustChangePassword: true,
		Roles:              []byte(`[" siswa ","siswa","guru",""]`),
	}, SessionMeta{IPAddress: "10.0.0.1", UserAgent: "UA", DeviceLabel: "Phone"})
	if err != nil {
		t.Fatalf("issueTokenPair() error = %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" || !pair.MustChangePassword {
		t.Fatalf("issueTokenPair() = %+v, want populated tokens and must-change flag", pair)
	}
	if len(store.authSessions) != 1 {
		t.Fatalf("created sessions = %d, want 1", len(store.authSessions))
	}
	var session db.AuthSession
	for _, s := range store.authSessions {
		session = s
	}
	if session.UserID != userID || session.IpAddress != "10.0.0.1" || session.UserAgent != "UA" || session.DeviceLabel != "Phone" || session.RefreshTokenHash != hashToken(pair.RefreshToken) {
		t.Fatalf("stored session = %+v, want user/meta/hash", session)
	}

	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(pair.AccessToken, claims, func(token *jwt.Token) (any, error) { return []byte("secret"), nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("ParseWithClaims(access) = %v/%v, want valid", parsed, err)
	}
	roles, ok := claims["roles"].([]any)
	if !ok || len(roles) != 2 || roles[0] != "siswa" || roles[1] != "guru" || claims["role"] != "siswa" {
		t.Fatalf("access roles claims = roles:%#v role:%#v, want normalized legacy fallback", claims["roles"], claims["role"])
	}
	permissions, ok := claims["permissions"].([]any)
	if !ok || len(permissions) != 2 || permissions[0] != "read:data" || permissions[1] != "write:data" {
		t.Fatalf("access permissions claim = %#v, want normalized permissions", claims["permissions"])
	}
	if claims["sid"] != studentID.String() || claims["ver"] != float64(7) || claims["must_change_password"] != true {
		t.Fatalf("access identity claims = %#v, want student/version/must-change", claims)
	}
}

func TestListAccountChangeHistoryNormalizesLimitAndRejectsInvalidUser(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret")}
	if _, err := svc.ListAccountChangeHistory(context.Background(), pgtype.UUID{}, 10); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("ListAccountChangeHistory(invalid user) = %v, want unauthorized", err)
	}
	userID := documentCycleTestUUID(33)
	store.accountHistory = []db.ListOwnAccountChangeHistoryRow{{Action: "AUTH_ACCOUNT_CONTACT_UPDATE"}, {Action: "ignored"}}
	items, err := svc.ListAccountChangeHistory(context.Background(), userID, 100)
	if err != nil {
		t.Fatalf("ListAccountChangeHistory() error = %v", err)
	}
	if store.lastHistoryArgs.UserID != userID || store.lastHistoryArgs.LimitCount != 50 {
		t.Fatalf("history args = %+v, want user and clamped limit 50", store.lastHistoryArgs)
	}
	if len(items) != 1 || items[0].Action != "contact_update" {
		t.Fatalf("history items = %+v, want filtered normalized contact update", items)
	}
}
