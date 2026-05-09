package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeAuditWriter struct {
	called bool
	arg    db.CreateAuditLogParams
}

func (f *fakeAuditWriter) CreateAuditLog(_ context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.called = true
	f.arg = arg
	return db.AuditLog{ID: pgtype.UUID{Valid: true}}, nil
}

func signedMiddlewareToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return token
}

func TestJWTValidatesVersionSessionAndInjectsClaims(t *testing.T) {
	versionCalled := false
	sessionCalled := false
	middleware := JWT(
		"secret",
		func(_ context.Context, sub string) (int64, error) {
			versionCalled = true
			if sub != "user-1" {
				t.Fatalf("currentVersion sub = %q, want user-1", sub)
			}
			return 3, nil
		},
		func(_ context.Context, sub, sessionID string) (bool, error) {
			sessionCalled = true
			if sub != "user-1" || sessionID != "session-1" {
				t.Fatalf("validateSession args = %q/%q, want user/session", sub, sessionID)
			}
			return true, nil
		},
	)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := api.ClaimsFromContext(r.Context())
		if !ok || claims["sub"] != "user-1" {
			t.Fatalf("claims = %+v/%v, want injected claims", claims, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+signedMiddlewareToken(t, "secret", jwt.MapClaims{
		"type": "access",
		"sub":  "user-1",
		"ver":  float64(3),
		"ssid": "session-1",
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent || !versionCalled || !sessionCalled {
		t.Fatalf("JWT status/version/session = %d/%v/%v, want 204/true/true", rec.Code, versionCalled, sessionCalled)
	}
}

func TestJWTRejectsInvalidVersionAndSessionClaims(t *testing.T) {
	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{name: "wrong type", claims: jwt.MapClaims{"type": "refresh", "sub": "user-1", "ver": float64(3), "ssid": "s"}},
		{name: "missing version", claims: jwt.MapClaims{"type": "access", "sub": "user-1", "ssid": "s"}},
		{name: "missing subject", claims: jwt.MapClaims{"type": "access", "ver": float64(3), "ssid": "s"}},
		{name: "missing session", claims: jwt.MapClaims{"type": "access", "sub": "user-1", "ver": float64(3)}},
		{name: "stale version", claims: jwt.MapClaims{"type": "access", "sub": "user-1", "ver": float64(2), "ssid": "s"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := JWT("secret", func(context.Context, string) (int64, error) {
				return 3, nil
			}, func(context.Context, string, string) (bool, error) {
				return true, nil
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("next handler should not be called")
			}))
			req := httptest.NewRequest(http.MethodGet, "/private", nil)
			req.Header.Set("Authorization", "Bearer "+signedMiddlewareToken(t, "secret", tt.claims))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("JWT(%s) status = %d, want 401; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}

	handler := JWT("secret", func(context.Context, string) (int64, error) {
		return 3, nil
	}, func(context.Context, string, string) (bool, error) {
		return false, nil
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+signedMiddlewareToken(t, "secret", jwt.MapClaims{"type": "access", "sub": "user-1", "ver": float64(3), "ssid": "s"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("JWT(invalid session) status = %d, want 401", rec.Code)
	}
}

func TestInternalKeyOrJWTAndWorkerKey(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	internal := InternalKeyOrJWT("internal", "secret", nil, nil)(next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal", nil)
	req.Header.Set("X-Internal-Key", "internal")
	internal.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("InternalKeyOrJWT(internal) status = %d, want 204", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/internal", nil)
	req.Header.Set("Authorization", "Bearer "+signedMiddlewareToken(t, "secret", jwt.MapClaims{"type": "access"}))
	internal.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("InternalKeyOrJWT(jwt) status = %d, want 204", rec.Code)
	}

	internalOnly := InternalKey("public-analytics-key")(next)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal-analytics/public-events", nil)
	req.Header.Set("X-Internal-Key", "public-analytics-key")
	internalOnly.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("InternalKey(valid) status = %d, want 204", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal-analytics/public-events", nil)
	req.Header.Set("X-Internal-Key", "wrong")
	internalOnly.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("InternalKey(invalid) status = %d, want 401", rec.Code)
	}

	worker := WorkerKey("worker-secret")(next)
	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/worker", nil),
		httptest.NewRequest(http.MethodPost, "/worker", nil),
	} {
		if req.Header.Get("X-Worker-Key") == "" {
			req.Header.Set("X-Worker-Key", "worker-secret")
		}
		rec = httptest.NewRecorder()
		worker.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("WorkerKey(valid) status = %d, want 204", rec.Code)
		}
		break
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/worker", nil)
	req.Header.Set("Authorization", "Bearer worker-secret")
	worker.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("WorkerKey(bearer) status = %d, want 204", rec.Code)
	}

	rec = httptest.NewRecorder()
	worker.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/worker", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("WorkerKey(missing) status = %d, want 401", rec.Code)
	}
}

func TestAuditLogsSuccessfulMutationsOnly(t *testing.T) {
	writer := &fakeAuditWriter{}
	handler := Audit(writer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/students/1", nil)
	req.Header.Set("X-Request-ID", "req-1")
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"uid": "01000000-0000-0000-0000-000000000000",
		"usr": "admin",
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated || !writer.called {
		t.Fatalf("Audit status/called = %d/%v, want 201/true", rec.Code, writer.called)
	}
	if writer.arg.Action != http.MethodPost || writer.arg.EntityType != "students" || writer.arg.EntityID != "/api/students/1" || !writer.arg.UserID.Valid {
		t.Fatalf("audit arg = %+v, want mutation metadata", writer.arg)
	}
	var metadata map[string]any
	if err := json.Unmarshal(writer.arg.Metadata, &metadata); err != nil {
		t.Fatalf("metadata unmarshal error = %v", err)
	}
	if metadata["actor"] != "admin" || metadata["request_id"] != "req-1" {
		t.Fatalf("audit metadata = %+v, want actor/request id", metadata)
	}

	writer = &fakeAuditWriter{}
	Audit(writer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/students", nil))
	if writer.called {
		t.Fatal("Audit(500) wrote audit log, want skipped")
	}

	writer = &fakeAuditWriter{}
	Audit(writer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/students", nil))
	if writer.called {
		t.Fatal("Audit(GET) wrote audit log, want skipped")
	}

	if got := entityFromPath("plain/path"); got != "plain" {
		t.Fatalf("entityFromPath() = %q, want plain", got)
	}
}

func TestAuditUsesChiRouteParamAsEntityID(t *testing.T) {
	writer := &fakeAuditWriter{}
	r := chi.NewRouter()
	r.With(Audit(writer)).Put("/api/students/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/students/01000000-0000-0000-0000-000000000001", nil))

	if rec.Code != http.StatusNoContent || !writer.called {
		t.Fatalf("Audit status/called = %d/%v, want 204/true", rec.Code, writer.called)
	}
	if writer.arg.EntityID != "01000000-0000-0000-0000-000000000001" {
		t.Fatalf("EntityID = %q, want route id param", writer.arg.EntityID)
	}

	var metadata map[string]any
	if err := json.Unmarshal(writer.arg.Metadata, &metadata); err != nil {
		t.Fatalf("metadata unmarshal error = %v", err)
	}
	if metadata["route"] != "/api/students/{id}" {
		t.Fatalf("route metadata = %v, want route pattern", metadata["route"])
	}
	params, ok := metadata["route_params"].(map[string]any)
	if !ok || params["id"] != "01000000-0000-0000-0000-000000000001" {
		t.Fatalf("route_params = %+v, want id param", metadata["route_params"])
	}
}

func TestAuditFallbackEntityIDUsesSanitizedPath(t *testing.T) {
	writer := &fakeAuditWriter{}
	handler := Audit(writer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/students/1?secret=token", nil))

	if rec.Code != http.StatusOK || !writer.called {
		t.Fatalf("Audit status/called = %d/%v, want 200/true", rec.Code, writer.called)
	}
	if writer.arg.EntityID != "/api/students/1" {
		t.Fatalf("EntityID = %q, want path without query", writer.arg.EntityID)
	}
}

func TestAuditDoesNotPanicOutsideChiContext(t *testing.T) {
	writer := &fakeAuditWriter{}
	handler := Audit(writer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/plain", nil))

	if rec.Code != http.StatusCreated || !writer.called {
		t.Fatalf("Audit status/called = %d/%v, want 201/true", rec.Code, writer.called)
	}
	if writer.arg.EntityID != "/api/plain" {
		t.Fatalf("EntityID = %q, want fallback path", writer.arg.EntityID)
	}
}

func TestExamTokenMiddlewareBranches(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := ParticipantFromContext(r.Context()); !ok {
			t.Fatal("participant missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	active := ExamToken(func(_ context.Context, token string) (db.GetParticipantByTokenRow, error) {
		return db.GetParticipantByTokenRow{
			Token:             token,
			SessionStatus:     db.CbtSessionStatusEnumActive,
			DeviceFingerprint: pgtype.Text{String: "device-1", Valid: true},
		}, nil
	})(next)
	req := httptest.NewRequest(http.MethodPost, "/exam", nil)
	req.Header.Set("X-Exam-Token", "token-1")
	req.Header.Set(DeviceFingerprintHdr, "device-1")
	rec := httptest.NewRecorder()
	active.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("ExamToken(active) status = %d, want 204", rec.Code)
	}

	missing := ExamToken(func(context.Context, string) (db.GetParticipantByTokenRow, error) {
		t.Fatal("lookup should not be called for missing token")
		return db.GetParticipantByTokenRow{}, nil
	})(next)
	rec = httptest.NewRecorder()
	missing.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/exam", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ExamToken(missing) status = %d, want 401", rec.Code)
	}

	failing := ExamToken(func(context.Context, string) (db.GetParticipantByTokenRow, error) {
		return db.GetParticipantByTokenRow{}, errors.New("not found")
	})(next)
	req = httptest.NewRequest(http.MethodPost, "/exam", nil)
	req.Header.Set("X-Exam-Token", "bad")
	req.Header.Set(DeviceFingerprintHdr, "device-1")
	rec = httptest.NewRecorder()
	failing.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ExamToken(lookup error) status = %d, want 401", rec.Code)
	}

	inactive := ExamToken(func(context.Context, string) (db.GetParticipantByTokenRow, error) {
		return db.GetParticipantByTokenRow{SessionStatus: db.CbtSessionStatusEnumDraft}, nil
	})(next)
	req = httptest.NewRequest(http.MethodPost, "/exam", nil)
	req.Header.Set("X-Exam-Token", "token-2")
	req.Header.Set(DeviceFingerprintHdr, "device-1")
	rec = httptest.NewRecorder()
	inactive.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("ExamToken(inactive) status = %d, want 403", rec.Code)
	}

	mismatch := ExamToken(func(context.Context, string) (db.GetParticipantByTokenRow, error) {
		return db.GetParticipantByTokenRow{
			SessionStatus:     db.CbtSessionStatusEnumActive,
			DeviceFingerprint: pgtype.Text{String: "device-1", Valid: true},
		}, nil
	})(next)
	req = httptest.NewRequest(http.MethodPost, "/exam", nil)
	req.Header.Set("X-Exam-Token", "token-3")
	req.Header.Set(DeviceFingerprintHdr, "device-2")
	rec = httptest.NewRecorder()
	mismatch.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("ExamToken(device mismatch) status = %d, want 409", rec.Code)
	}

	missingFingerprint := ExamToken(func(context.Context, string) (db.GetParticipantByTokenRow, error) {
		return db.GetParticipantByTokenRow{
			SessionStatus:     db.CbtSessionStatusEnumActive,
			DeviceFingerprint: pgtype.Text{String: "device-1", Valid: true},
		}, nil
	})(next)
	req = httptest.NewRequest(http.MethodPost, "/exam", nil)
	req.Header.Set("X-Exam-Token", "token-4")
	rec = httptest.NewRecorder()
	missingFingerprint.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ExamToken(missing fingerprint) status = %d, want 401", rec.Code)
	}

	unbound := ExamToken(func(context.Context, string) (db.GetParticipantByTokenRow, error) {
		return db.GetParticipantByTokenRow{
			SessionStatus: db.CbtSessionStatusEnumActive,
		}, nil
	})(next)
	req = httptest.NewRequest(http.MethodPost, "/exam", nil)
	req.Header.Set("X-Exam-Token", "token-5")
	req.Header.Set(DeviceFingerprintHdr, "device-1")
	rec = httptest.NewRecorder()
	unbound.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ExamToken(unbound participant) status = %d, want 401", rec.Code)
	}
}

func TestRequestLogPassesResponseThrough(t *testing.T) {
	handler := RequestLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ok"))
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted || rec.Body.String() != "ok" {
		t.Fatalf("RequestLog response = %d/%q, want 202/ok", rec.Code, rec.Body.String())
	}
	if got := bearerToken(httptest.NewRequest(http.MethodGet, "/", nil)); got != "" {
		t.Fatalf("bearerToken(no header) = %q, want empty", got)
	}
	if !HasAnyRole(jwt.MapClaims{"roles": []string{"staf"}}, "staf") {
		t.Fatal("HasAnyRole([]string) = false, want true")
	}
}
