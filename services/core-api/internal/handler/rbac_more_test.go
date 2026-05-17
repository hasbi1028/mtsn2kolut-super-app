package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/service"
)

func TestRBACMoreListPermissionsEdges(t *testing.T) {
	adminID := handlerTestUUID(190)
	tests := []struct {
		name       string
		req        *http.Request
		svc        *fakeRBACService
		wantStatus int
	}{
		{
			name:       "forbidden without claims",
			req:        httptest.NewRequest(http.MethodGet, "/api/rbac/permissions", nil),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "service error",
			req:        adminRBACRequest(http.MethodGet, "/api/rbac/permissions", "", adminID),
			svc:        &fakeRBACService{err: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&RBAC{svc: tt.svc}).ListPermissions(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestRBACMoreRoleMutationEdges(t *testing.T) {
	adminID := handlerTestUUID(191)
	tests := []struct {
		name       string
		handler    func(*RBAC, http.ResponseWriter, *http.Request)
		req        *http.Request
		svc        *fakeRBACService
		wantStatus int
	}{
		{
			name:       "update role forbidden",
			handler:    (*RBAC).UpdateRole,
			req:        httptest.NewRequest(http.MethodPut, "/api/rbac/roles/operator", nil),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "update role invalid json",
			handler:    (*RBAC).UpdateRole,
			req:        withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/roles/operator", `{`, adminID), "code", "operator"),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update role missing actor",
			handler:    (*RBAC).UpdateRole,
			req:        withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/api/rbac/roles/operator", strings.NewReader(`{"name":"Operator"}`)), jwt.MapClaims{"roles": []any{"admin"}}), "code", "operator"),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "update role service error",
			handler:    (*RBAC).UpdateRole,
			req:        withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/roles/operator", `{"name":"Operator"}`, adminID), "code", "operator"),
			svc:        &fakeRBACService{err: errors.New("role tidak ditemukan")},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "set role status missing code",
			handler:    (*RBAC).SetRoleStatus,
			req:        withRouteParam(adminRBACRequest(http.MethodPatch, "/api/rbac/roles//status", `{"is_active":true}`, adminID), "code", " "),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "set role status missing actor",
			handler:    (*RBAC).SetRoleStatus,
			req:        withRouteParam(withClaims(httptest.NewRequest(http.MethodPatch, "/api/rbac/roles/operator/status", strings.NewReader(`{"is_active":true}`)), jwt.MapClaims{"roles": []any{"admin"}}), "code", "operator"),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "set role status service error",
			handler:    (*RBAC).SetRoleStatus,
			req:        withRouteParam(adminRBACRequest(http.MethodPatch, "/api/rbac/roles/operator/status", `{"is_active":false}`, adminID), "code", "operator"),
			svc:        &fakeRBACService{err: errors.New("duplicate role")},
			wantStatus: http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&RBAC{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestRBACMorePermissionMutationEdges(t *testing.T) {
	adminID := handlerTestUUID(192)
	tests := []struct {
		name       string
		handler    func(*RBAC, http.ResponseWriter, *http.Request)
		req        *http.Request
		svc        *fakeRBACService
		wantStatus int
	}{
		{
			name:       "update permission forbidden",
			handler:    (*RBAC).UpdatePermission,
			req:        httptest.NewRequest(http.MethodPut, "/api/rbac/permissions/reports.read", nil),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "update permission invalid json",
			handler:    (*RBAC).UpdatePermission,
			req:        withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/permissions/reports.read", `{`, adminID), "code", "reports.read"),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update permission missing actor",
			handler:    (*RBAC).UpdatePermission,
			req:        withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/api/rbac/permissions/reports.read", strings.NewReader(`{"module":"reports","action":"read"}`)), jwt.MapClaims{"roles": []any{"admin"}}), "code", "reports.read"),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "update permission service error",
			handler:    (*RBAC).UpdatePermission,
			req:        withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/permissions/reports.read", `{"module":"reports","action":"read"}`, adminID), "code", "reports.read"),
			svc:        &fakeRBACService{err: errors.New("permission tidak ditemukan")},
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&RBAC{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestRBACMoreAssignmentMutationEdges(t *testing.T) {
	adminID := handlerTestUUID(193)
	userID := handlerTestUUID(194)
	tests := []struct {
		name       string
		handler    func(*RBAC, http.ResponseWriter, *http.Request)
		req        *http.Request
		svc        *fakeRBACService
		wantStatus int
	}{
		{
			name:       "role permissions forbidden",
			handler:    (*RBAC).UpdateRolePermissions,
			req:        httptest.NewRequest(http.MethodPut, "/api/rbac/roles/guru/permissions", nil),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "role permissions missing code",
			handler:    (*RBAC).UpdateRolePermissions,
			req:        withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/roles//permissions", `{"permissions":["a.b"]}`, adminID), "code", " "),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "role permissions missing actor",
			handler:    (*RBAC).UpdateRolePermissions,
			req:        withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/api/rbac/roles/guru/permissions", strings.NewReader(`{"permissions":["a.b"]}`)), jwt.MapClaims{"roles": []any{"admin"}}), "code", "guru"),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "role permissions service error",
			handler:    (*RBAC).UpdateRolePermissions,
			req:        withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/roles/guru/permissions", `{"permissions":["a.b"]}`, adminID), "code", "guru"),
			svc:        &fakeRBACService{replaceRoleErr: errors.New("permission not found")},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "user roles forbidden",
			handler:    (*RBAC).UpdateUserRoles,
			req:        httptest.NewRequest(http.MethodPatch, "/api/users/"+userID.String()+"/roles", nil),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "user roles invalid json",
			handler:    (*RBAC).UpdateUserRoles,
			req:        withRouteParam(adminRBACRequest(http.MethodPatch, "/api/users/"+userID.String()+"/roles", `{`, adminID), "id", userID.String()),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "user roles missing actor",
			handler:    (*RBAC).UpdateUserRoles,
			req:        withRouteParam(withClaims(httptest.NewRequest(http.MethodPatch, "/api/users/"+userID.String()+"/roles", strings.NewReader(`{"roles":["guru"]}`)), jwt.MapClaims{"roles": []any{"admin"}}), "id", userID.String()),
			svc:        &fakeRBACService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "user roles service error",
			handler:    (*RBAC).UpdateUserRoles,
			req:        withRouteParam(adminRBACRequest(http.MethodPatch, "/api/users/"+userID.String()+"/roles", `{"roles":["guru"]}`, adminID), "id", userID.String()),
			svc:        &fakeRBACService{replaceUserErr: errors.New("role duplicate")},
			wantStatus: http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&RBAC{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestRBACMorePayloadForwardingTrimsCodes(t *testing.T) {
	adminID := handlerTestUUID(195)
	svc := &fakeRBACService{}
	h := &RBAC{svc: svc}

	rec := httptest.NewRecorder()
	h.UpdateRole(rec, withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/roles/operator", `{"name":" Operator "}`, adminID), "code", " operator "))
	if rec.Code != http.StatusOK || svc.updatedRoleCode != "operator" || svc.updatedRoleInput.Name != " Operator " {
		t.Fatalf("UpdateRole status=%d code=%q input=%+v body=%s", rec.Code, svc.updatedRoleCode, svc.updatedRoleInput, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdatePermission(rec, withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/permissions/reports.read", `{"module":"reports","action":"read","description":"Read reports"}`, adminID), "code", " reports.read "))
	if rec.Code != http.StatusOK || svc.updatedPermissionCode != "reports.read" || svc.updatedPermissionInput.Action != "read" {
		t.Fatalf("UpdatePermission status=%d code=%q input=%+v body=%s", rec.Code, svc.updatedPermissionCode, svc.updatedPermissionInput, rec.Body.String())
	}

	var _ service.RBACRoleInput
}
