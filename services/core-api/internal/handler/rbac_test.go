package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeRBACService struct {
	matrix service.RBACMatrix
	err    error

	replaceRoleCode        string
	replacePermissionCodes []string
	replaceRoleActor       pgtype.UUID
	replaceRoleErr         error

	replaceUserID    pgtype.UUID
	replaceRoleCodes []string
	replaceUserActor pgtype.UUID
	replaceUserErr   error
}

func (f *fakeRBACService) ListMatrix(ctx context.Context) (service.RBACMatrix, error) {
	return f.matrix, f.err
}
func (f *fakeRBACService) GetUserPermissions(ctx context.Context, userID pgtype.UUID) ([]string, error) {
	return nil, nil
}
func (f *fakeRBACService) GetUserRoles(ctx context.Context, userID pgtype.UUID) ([]string, error) {
	return nil, nil
}
func (f *fakeRBACService) ReplaceRolePermissions(ctx context.Context, roleCode string, permissionCodes []string, actorID pgtype.UUID) error {
	f.replaceRoleCode = roleCode
	f.replacePermissionCodes = append([]string(nil), permissionCodes...)
	f.replaceRoleActor = actorID
	return f.replaceRoleErr
}
func (f *fakeRBACService) ReplaceUserRoles(ctx context.Context, userID pgtype.UUID, roleCodes []string, actorID pgtype.UUID) error {
	f.replaceUserID = userID
	f.replaceRoleCodes = append([]string(nil), roleCodes...)
	f.replaceUserActor = actorID
	return f.replaceUserErr
}

func TestRBACHandlerListsMatrixRolesAndPermissions(t *testing.T) {
	adminID := handlerTestUUID(90)
	svc := &fakeRBACService{matrix: service.RBACMatrix{
		Roles:           []db.RbacRole{{Code: "admin", Name: "Administrator"}},
		Permissions:     []db.RbacPermission{{Code: "users.read", Module: "users"}},
		RolePermissions: []db.ListRbacRolePermissionsRow{{RoleCode: "admin", PermissionCode: "users.read"}},
	}}
	h := &RBAC{svc: svc}

	rec := httptest.NewRecorder()
	h.ListMatrix(rec, adminRBACRequest(http.MethodGet, "/api/rbac/matrix", "", adminID))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListMatrix() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var res struct {
		Data service.RBACMatrix `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("ListMatrix() json error = %v", err)
	}
	if len(res.Data.Roles) != 1 || len(res.Data.Permissions) != 1 || len(res.Data.RolePermissions) != 1 {
		t.Fatalf("ListMatrix() data = %+v, want matrix slices", res.Data)
	}

	rec = httptest.NewRecorder()
	h.ListRoles(rec, adminRBACRequest(http.MethodGet, "/api/rbac/roles", "", adminID))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListRoles() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListPermissions(rec, adminRBACRequest(http.MethodGet, "/api/rbac/permissions", "", adminID))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListPermissions() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func TestRBACHandlerUpdatesRolePermissionsAndUserRoles(t *testing.T) {
	adminID := handlerTestUUID(91)
	userID := handlerTestUUID(92)
	svc := &fakeRBACService{}
	h := &RBAC{svc: svc}

	rec := httptest.NewRecorder()
	req := withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/roles/guru/permissions", `{"permissions":["bank_soal.read","bank_soal.create"]}`, adminID), "code", "guru")
	h.UpdateRolePermissions(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateRolePermissions() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.replaceRoleCode != "guru" || !reflect.DeepEqual(svc.replacePermissionCodes, []string{"bank_soal.read", "bank_soal.create"}) || svc.replaceRoleActor != adminID {
		t.Fatalf("role update args = %q/%v/%v", svc.replaceRoleCode, svc.replacePermissionCodes, svc.replaceRoleActor)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRBACRequest(http.MethodPatch, "/api/users/"+userID.String()+"/roles", `{"roles":["guru","staf"]}`, adminID), "id", userID.String())
	h.UpdateUserRoles(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateUserRoles() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.replaceUserID != userID || !reflect.DeepEqual(svc.replaceRoleCodes, []string{"guru", "staf"}) || svc.replaceUserActor != adminID {
		t.Fatalf("user role args = %v/%v/%v", svc.replaceUserID, svc.replaceRoleCodes, svc.replaceUserActor)
	}
}

func TestRBACHandlerValidationAndAccessErrors(t *testing.T) {
	adminID := handlerTestUUID(93)
	userID := handlerTestUUID(94)
	svc := &fakeRBACService{err: errors.New("db down")}
	h := &RBAC{svc: svc}

	tests := []struct {
		name       string
		handler    func(http.ResponseWriter, *http.Request)
		req        *http.Request
		wantStatus int
	}{
		{name: "matrix forbidden", handler: h.ListMatrix, req: httptest.NewRequest(http.MethodGet, "/api/rbac/matrix", nil), wantStatus: http.StatusForbidden},
		{name: "matrix store error", handler: h.ListMatrix, req: adminRBACRequest(http.MethodGet, "/api/rbac/matrix", "", adminID), wantStatus: http.StatusInternalServerError},
		{name: "role permissions invalid json", handler: h.UpdateRolePermissions, req: withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/roles/guru/permissions", `{`, adminID), "code", "guru"), wantStatus: http.StatusBadRequest},
		{name: "role permissions missing body", handler: h.UpdateRolePermissions, req: withRouteParam(adminRBACRequest(http.MethodPut, "/api/rbac/roles/guru/permissions", `{}`, adminID), "code", "guru"), wantStatus: http.StatusBadRequest},
		{name: "user roles invalid id", handler: h.UpdateUserRoles, req: withRouteParam(adminRBACRequest(http.MethodPatch, "/api/users/bad/roles", `{"roles":["guru"]}`, adminID), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "user roles missing roles", handler: h.UpdateUserRoles, req: withRouteParam(adminRBACRequest(http.MethodPatch, "/api/users/"+userID.String()+"/roles", `{}`, adminID), "id", userID.String()), wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func adminRBACRequest(method, target, body string, userID pgtype.UUID) *http.Request {
	req := adminRequest(method, target, body)
	if body != "" && req.Body == nil {
		req.Body = httptest.NewRequest(method, target, strings.NewReader(body)).Body
	}
	return withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "uid": userID.String(), "sub": userID.String()})
}
