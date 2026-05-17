package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/domain"
)

func TestCbtPackageExtraAccessAndOptionalEventBranches(t *testing.T) {
	fake := &fakeCbtPackageService{}
	h := &CbtPackage{svc: fake}

	rec := httptest.NewRecorder()
	h.Readiness(rec, httptest.NewRequest(http.MethodGet, "/api/cbt/packages/readiness", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Readiness(no claims) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	if fake.readinessEventID.Valid {
		t.Fatalf("Readiness(no claims) reached service with event id %v", fake.readinessEventID)
	}

	rec = httptest.NewRecorder()
	h.List(rec, cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages?event_id=", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List(empty event_id) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listEventID.Valid {
		t.Fatalf("List(empty event_id) event id = %v, want zero optional uuid", fake.listEventID)
	}

	rec = httptest.NewRecorder()
	h.List(rec, cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages?event_id=bad", ""))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("List(bad event_id) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtPackageExtraLowBranchErrorMapping(t *testing.T) {
	packageID := handlerTestUUID(170)
	tests := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		req  *http.Request
		want int
	}{
		{
			name: "readiness service error",
			fn:   (&CbtPackage{svc: &fakeCbtPackageService{readinessErr: errors.New("db down")}}).Readiness,
			req:  cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages/readiness", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "get not found",
			fn:   (&CbtPackage{svc: &fakeCbtPackageService{detailErr: domain.ErrNotFound}}).Get,
			req:  cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodGet, "/api/cbt/packages/"+packageID.String(), ""), "id", packageID.String()),
			want: http.StatusNotFound,
		},
		{
			name: "update forbidden service error",
			fn:   (&CbtPackage{svc: &fakeCbtPackageService{updateErr: domain.ErrForbidden}}).Update,
			req:  cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPatch, "/api/cbt/packages/"+packageID.String(), `{"title":"P"}`), "id", packageID.String()),
			want: http.StatusForbidden,
		},
		{
			name: "replace conflict service error",
			fn:   (&CbtPackage{svc: &fakeCbtPackageService{replaceErr: domain.ErrConflict}}).ReplaceQuestions,
			req:  cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPut, "/api/cbt/packages/"+packageID.String()+"/questions", `{"question_ids":[]}`), "id", packageID.String()),
			want: http.StatusConflict,
		},
		{
			name: "clone not found service error",
			fn:   (&CbtPackage{svc: &fakeCbtPackageService{cloneErr: domain.ErrNotFound}}).Clone,
			req:  cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages/"+packageID.String()+"/clone", `{"title":"Clone"}`), "id", packageID.String()),
			want: http.StatusNotFound,
		},
		{
			name: "lock internal service error",
			fn:   (&CbtPackage{svc: &fakeCbtPackageService{lockErr: errors.New("db down")}}).Lock,
			req:  cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages/"+packageID.String()+"/lock", ``), "id", packageID.String()),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestCbtPackageExtraLockAuthBranches(t *testing.T) {
	packageID := handlerTestUUID(171)
	h := &CbtPackage{svc: &fakeCbtPackageService{}}

	rec := httptest.NewRecorder()
	req := cbtPackageRouteParam(cbtPackageAuthedRequest(http.MethodPost, "/api/cbt/packages/"+packageID.String()+"/lock", `{"reason":"x"}`), "id", "bad")
	h.Lock(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Lock(bad id) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = cbtPackageRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/packages/"+packageID.String()+"/lock", strings.NewReader(`{"reason":"x"}`)), jwt.MapClaims{"roles": []any{"admin"}}), "id", packageID.String())
	h.Lock(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Lock(missing uid) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
}
