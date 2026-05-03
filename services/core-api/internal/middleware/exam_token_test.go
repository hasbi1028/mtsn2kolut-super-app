package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestExamTokenOrJWTRejectsInternalKeyHeaderOnly(t *testing.T) {
	middleware := ExamTokenOrJWT("secret", nil, nil, func(ctx context.Context, token string) (db.GetParticipantByTokenRow, error) {
		return db.GetParticipantByTokenRow{}, errors.New("unused")
	})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/cbt/assets/abc/file", nil)
	req.Header.Set("X-Internal-Key", "shared-secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestExamTokenOrJWTAllowsActiveExamToken(t *testing.T) {
	packageID := pgtype.UUID{Valid: true}
	middleware := ExamTokenOrJWT("secret", nil, nil, func(ctx context.Context, token string) (db.GetParticipantByTokenRow, error) {
		return db.GetParticipantByTokenRow{
			Token:             token,
			PackageID:         packageID,
			SessionStatus:     db.CbtSessionStatusEnumActive,
			DeviceFingerprint: pgtype.Text{String: "device-1", Valid: true},
		}, nil
	})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := ParticipantFromContext(r.Context()); !ok {
			t.Fatal("participant missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/cbt/assets/abc/file", nil)
	req.Header.Set("X-Exam-Token", "token-1")
	req.Header.Set(DeviceFingerprintHdr, "device-1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestExamTokenOrJWTRejectsLeakedQueryExamToken(t *testing.T) {
	middleware := ExamTokenOrJWT("secret", nil, nil, func(ctx context.Context, token string) (db.GetParticipantByTokenRow, error) {
		t.Fatal("lookup should not be called for query token")
		return db.GetParticipantByTokenRow{}, nil
	})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/cbt/assets/abc/file?exam_token=token-1", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
