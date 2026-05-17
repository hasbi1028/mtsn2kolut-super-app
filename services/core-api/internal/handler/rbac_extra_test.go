package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestRBACExtraRoleAndPermissionMutationErrors(t *testing.T) {
	adminID := handlerTestUUID(201)
	tests := []struct {
		name       string
		handler    func(*RBAC, http.ResponseWriter, *http.Request)
		req        *http.Request
		wantStatus int
	}{
		{
			name: "list roles service error",
			handler: func(h *RBAC, w http.ResponseWriter, r *http.Request) {
				h.ListRoles(w, r)
			},
			req:        adminRBACRequest(http.MethodGet, "/api/rbac/roles", "", adminID),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "create role invalid json",
			handler: func(h *RBAC, w http.ResponseWriter, r *http.Request) {
				h.CreateRole(w, r)
			},
			req:        adminRBACRequest(http.MethodPost, "/api/rbac/roles", `{`, adminID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "update role missing code",
			handler: func(h *RBAC, w http.ResponseWriter, r *http.Request) {
				h.UpdateRole(w, r)
			},
			req:        adminRBACRequest(http.MethodPut, "/api/rbac/roles/", `{"name":"Operator"}`, adminID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "role status invalid json",
			handler: func(h *RBAC, w http.ResponseWriter, r *http.Request) {
				h.SetRoleStatus(w, r)
			},
			req:        withRouteParam(adminRBACRequest(http.MethodPatch, "/api/rbac/roles/operator/status", `{`, adminID), "code", "operator"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "create permission invalid json",
			handler: func(h *RBAC, w http.ResponseWriter, r *http.Request) {
				h.CreatePermission(w, r)
			},
			req:        adminRBACRequest(http.MethodPost, "/api/rbac/permissions", `{`, adminID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "update permission missing code",
			handler: func(h *RBAC, w http.ResponseWriter, r *http.Request) {
				h.UpdatePermission(w, r)
			},
			req:        adminRBACRequest(http.MethodPut, "/api/rbac/permissions/", `{"module":"reports","action":"read"}`, adminID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "permission status invalid json",
			handler: func(h *RBAC, w http.ResponseWriter, r *http.Request) {
				h.SetPermissionStatus(w, r)
			},
			req:        withRouteParam(adminRBACRequest(http.MethodPatch, "/api/rbac/permissions/reports.read/status", `{`, adminID), "code", "reports.read"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "permission create forbidden without claims",
			handler: func(h *RBAC, w http.ResponseWriter, r *http.Request) {
				h.CreatePermission(w, r)
			},
			req:        httptest.NewRequest(http.MethodPost, "/api/rbac/permissions", nil),
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &RBAC{svc: &fakeRBACService{err: errors.New("service down")}}
			rec := httptest.NewRecorder()
			tt.handler(h, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestRBACExtraRoleAndPermissionRequireActorClaims(t *testing.T) {
	adminID := handlerTestUUID(202)
	svc := &fakeRBACService{}
	h := &RBAC{svc: svc}

	req := withClaims(httptest.NewRequest(http.MethodPost, "/api/rbac/roles", strings.NewReader(`{"code":"operator","name":"Operator"}`)), jwt.MapClaims{"roles": []any{"admin"}})
	rec := httptest.NewRecorder()
	h.CreateRole(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("CreateRole without uid status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	req = withRouteParam(adminRBACRequest(http.MethodPatch, "/api/rbac/permissions/reports.read/status", `{"is_active":true}`, adminID), "code", " reports.read ")
	rec = httptest.NewRecorder()
	h.SetPermissionStatus(rec, req)
	if rec.Code != http.StatusOK || svc.permissionStatusCode != "reports.read" || !svc.permissionStatusActive {
		t.Fatalf("SetPermissionStatus status=%d code=%q active=%v body=%s", rec.Code, svc.permissionStatusCode, svc.permissionStatusActive, rec.Body.String())
	}
}
