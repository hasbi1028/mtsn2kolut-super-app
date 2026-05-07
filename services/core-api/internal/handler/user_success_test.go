package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeUserStore struct {
	listRows []db.ListUsersRow
	listErr  error

	createArg db.CreateUserParams
	createErr error
	roles     []db.AddUserRoleParams
	rbacRoles []db.AddUserRbacRoleByCodeParams
	roleErr   error

	auditArg  db.ListAuditLogsParams
	auditRows []db.ListAuditLogsRow
	auditErr  error
}

type fakeUserLifecycle struct {
	deleteID             pgtype.UUID
	deleteActorID        pgtype.UUID
	deleteErr            error
	statusID             pgtype.UUID
	statusActorID        pgtype.UUID
	statusIsActive       bool
	statusErr            error
	resetID              pgtype.UUID
	resetActorID         pgtype.UUID
	resetPassword        string
	resetErr             error
	forcePasswordID      pgtype.UUID
	forcePasswordActorID pgtype.UUID
	forcePasswordErr     error
	profileID            pgtype.UUID
	profileActorID       pgtype.UUID
	profileLink          service.ProfileLink
	profileErr           error
}

func (f *fakeUserStore) ListUsers(ctx context.Context) ([]db.ListUsersRow, error) {
	return f.listRows, f.listErr
}

func (f *fakeUserStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error) {
	f.createArg = arg
	id := handlerTestUUID(230)
	return db.CreateUserRow{ID: id, Username: arg.Username, IsActive: arg.IsActive}, f.createErr
}

func (f *fakeUserStore) AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error {
	f.roles = append(f.roles, arg)
	return f.roleErr
}

func (f *fakeUserStore) AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error {
	f.rbacRoles = append(f.rbacRoles, arg)
	return f.roleErr
}

func (f *fakeUserLifecycle) DeleteAsDeactivate(ctx context.Context, id pgtype.UUID, actorID pgtype.UUID) error {
	f.deleteID = id
	f.deleteActorID = actorID
	return f.deleteErr
}

func (f *fakeUserLifecycle) UpdateStatus(ctx context.Context, id pgtype.UUID, isActive bool, actorID pgtype.UUID) error {
	f.statusID = id
	f.statusActorID = actorID
	f.statusIsActive = isActive
	return f.statusErr
}

func (f *fakeUserLifecycle) ResetPassword(ctx context.Context, id pgtype.UUID, newPassword string, actorID pgtype.UUID) error {
	f.resetID = id
	f.resetPassword = newPassword
	f.resetActorID = actorID
	return f.resetErr
}

func (f *fakeUserLifecycle) ForcePasswordChange(ctx context.Context, id pgtype.UUID, actorID pgtype.UUID) error {
	f.forcePasswordID = id
	f.forcePasswordActorID = actorID
	return f.forcePasswordErr
}

func (f *fakeUserLifecycle) UpdateProfileLink(ctx context.Context, id pgtype.UUID, link service.ProfileLink, actorID pgtype.UUID) error {
	f.profileID = id
	f.profileLink = link
	f.profileActorID = actorID
	return f.profileErr
}

func (f *fakeUserStore) ListAuditLogs(ctx context.Context, arg db.ListAuditLogsParams) ([]db.ListAuditLogsRow, error) {
	f.auditArg = arg
	return f.auditRows, f.auditErr
}

func TestUserSuccessHandlersForwardPayloads(t *testing.T) {
	userID := handlerTestUUID(231)
	employeeID := handlerTestUUID(237)
	roleJSON, _ := json.Marshal([]string{"admin", "staf"})
	store := &fakeUserStore{
		listRows:  []db.ListUsersRow{{ID: userID, Username: "admin", Roles: roleJSON}},
		auditRows: []db.ListAuditLogsRow{{ID: handlerTestUUID(232), Action: "AUTH_LOGIN"}},
	}
	lifecycle := &fakeUserLifecycle{}
	h := &User{q: store, lifecycle: lifecycle}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/users", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var listed struct {
		Data []struct {
			Username string   `json:"username"`
			Roles    []string `json:"roles"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("List() json error = %v", err)
	}
	if len(listed.Data) != 1 || listed.Data[0].Username != "admin" || len(listed.Data[0].Roles) != 2 {
		t.Fatalf("List() data = %+v, want decoded roles", listed.Data)
	}

	rec = httptest.NewRecorder()
	h.Create(rec, adminRequest(http.MethodPost, "/api/users", `{"username":"operator","password":"secret123","roles":["admin"]}`))
	if rec.Code != http.StatusCreated || store.createArg.Username != "operator" || !store.createArg.IsActive || store.createArg.PasswordHash == "" || len(store.roles) != 1 || store.roles[0].Role != db.UserRoleAdmin || len(store.rbacRoles) != 1 || store.rbacRoles[0].Code != "admin" {
		t.Fatalf("Create() status/arg/roles/rbac = %d/%+v/%+v/%+v", rec.Code, store.createArg, store.roles, store.rbacRoles)
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("Create() json error = %v", err)
	}
	data, ok := created["data"].(map[string]any)
	if !ok || data["password_hash"] != nil {
		t.Fatalf("Create() data exposes password_hash: %#v", created["data"])
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodDelete, "/api/users/"+userID.String(), ""), "id", userID.String())
	h.Delete(rec, req)
	if rec.Code != http.StatusNoContent || lifecycle.deleteID != userID {
		t.Fatalf("Delete() status/id = %d/%v", rec.Code, lifecycle.deleteID)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/status", `{"is_active":false}`), "id", userID.String())
	h.UpdateStatus(rec, req)
	if rec.Code != http.StatusOK || lifecycle.statusID != userID || lifecycle.statusIsActive {
		t.Fatalf("UpdateStatus() status/arg = %d/%v/%v", rec.Code, lifecycle.statusID, lifecycle.statusIsActive)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPost, "/api/users/"+userID.String()+"/reset-password", `{"password":"newSecret123"}`), "id", userID.String())
	h.ResetPassword(rec, req)
	if rec.Code != http.StatusOK || lifecycle.resetID != userID || lifecycle.resetPassword != "newSecret123" {
		t.Fatalf("ResetPassword() status/arg = %d/%v/%q", rec.Code, lifecycle.resetID, lifecycle.resetPassword)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/profile-link", `{"employee_id":"`+employeeID.String()+`"}`), "id", userID.String())
	h.UpdateProfileLink(rec, req)
	if rec.Code != http.StatusOK || lifecycle.profileID != userID || lifecycle.profileLink.EmployeeID != employeeID {
		t.Fatalf("UpdateProfileLink() status/arg = %d/%v/%+v", rec.Code, lifecycle.profileID, lifecycle.profileLink)
	}

	rec = httptest.NewRecorder()
	h.ListAuditLogs(rec, adminRequest(http.MethodGet, "/api/settings/audit-logs?page=2&per_page=25", ""))
	if rec.Code != http.StatusOK || store.auditArg.Limit != 25 || store.auditArg.Offset != 25 {
		t.Fatalf("ListAuditLogs() status/arg = %d/%+v", rec.Code, store.auditArg)
	}
}

func TestUserValidationAndStoreErrors(t *testing.T) {
	userID := handlerTestUUID(233)
	employeeID := handlerTestUUID(234)
	studentID := handlerTestUUID(235)
	parentID := handlerTestUUID(236)
	errDB := errors.New("db down")

	plainRequest := func(method, target, body string) *http.Request {
		return httptest.NewRequest(method, target, strings.NewReader(body))
	}
	tests := []struct {
		name       string
		handler    func(*User, http.ResponseWriter, *http.Request)
		store      *fakeUserStore
		lifecycle  *fakeUserLifecycle
		req        *http.Request
		wantStatus int
	}{
		{name: "list forbidden", handler: (*User).List, req: plainRequest(http.MethodGet, "/api/users", ""), wantStatus: http.StatusForbidden},
		{name: "list store error", handler: (*User).List, store: &fakeUserStore{listErr: errDB}, req: adminRequest(http.MethodGet, "/api/users", ""), wantStatus: http.StatusInternalServerError},
		{name: "create forbidden", handler: (*User).Create, req: plainRequest(http.MethodPost, "/api/users", `{}`), wantStatus: http.StatusForbidden},
		{name: "create invalid json", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{`), wantStatus: http.StatusBadRequest},
		{name: "create missing credentials", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{"username":"operator"}`), wantStatus: http.StatusBadRequest},
		{name: "create missing role", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{"username":"operator","password":"secret123"}`), wantStatus: http.StatusBadRequest},
		{name: "create invalid role", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{"username":"operator","password":"secret123","roles":["super"]}`), wantStatus: http.StatusBadRequest},
		{name: "create employee role without employee", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{"username":"guru","password":"secret123","roles":["guru"]}`), wantStatus: http.StatusBadRequest},
		{name: "create student role without student", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{"username":"siswa","password":"secret123","roles":["siswa"]}`), wantStatus: http.StatusBadRequest},
		{name: "create parent role without parent", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{"username":"ortu","password":"secret123","roles":["ortu"]}`), wantStatus: http.StatusBadRequest},
		{name: "create employee link on admin", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{"username":"admin2","password":"secret123","roles":["admin"],"employee_id":"`+employeeID.String()+`"}`), wantStatus: http.StatusBadRequest},
		{name: "create multiple links", handler: (*User).Create, req: adminRequest(http.MethodPost, "/api/users", `{"username":"campur","password":"secret123","roles":["guru","siswa"],"employee_id":"`+employeeID.String()+`","student_id":"`+studentID.String()+`"}`), wantStatus: http.StatusBadRequest},
		{name: "create store error", handler: (*User).Create, store: &fakeUserStore{createErr: errDB}, req: adminRequest(http.MethodPost, "/api/users", `{"username":"operator","password":"secret123","roles":["admin"]}`), wantStatus: http.StatusInternalServerError},
		{name: "delete forbidden", handler: (*User).Delete, req: withRouteParam(plainRequest(http.MethodDelete, "/api/users/"+userID.String(), ""), "id", userID.String()), wantStatus: http.StatusForbidden},
		{name: "delete invalid id", handler: (*User).Delete, req: withRouteParam(adminRequest(http.MethodDelete, "/api/users/bad", ""), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "delete service error", handler: (*User).Delete, lifecycle: &fakeUserLifecycle{deleteErr: errDB}, req: withRouteParam(adminRequest(http.MethodDelete, "/api/users/"+userID.String(), ""), "id", userID.String()), wantStatus: http.StatusInternalServerError},
		{name: "status forbidden", handler: (*User).UpdateStatus, req: withRouteParam(plainRequest(http.MethodPatch, "/api/users/"+userID.String()+"/status", `{}`), "id", userID.String()), wantStatus: http.StatusForbidden},
		{name: "status invalid id", handler: (*User).UpdateStatus, req: withRouteParam(adminRequest(http.MethodPatch, "/api/users/bad/status", `{}`), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "status invalid json", handler: (*User).UpdateStatus, req: withRouteParam(adminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/status", `{`), "id", userID.String()), wantStatus: http.StatusBadRequest},
		{name: "status service error", handler: (*User).UpdateStatus, lifecycle: &fakeUserLifecycle{statusErr: errDB}, req: withRouteParam(adminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/status", `{"is_active":true}`), "id", userID.String()), wantStatus: http.StatusInternalServerError},
		{name: "reset password forbidden", handler: (*User).ResetPassword, req: withRouteParam(plainRequest(http.MethodPost, "/api/users/"+userID.String()+"/reset-password", `{}`), "id", userID.String()), wantStatus: http.StatusForbidden},
		{name: "reset password invalid id", handler: (*User).ResetPassword, req: withRouteParam(adminRequest(http.MethodPost, "/api/users/bad/reset-password", `{}`), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "reset password invalid json", handler: (*User).ResetPassword, req: withRouteParam(adminRequest(http.MethodPost, "/api/users/"+userID.String()+"/reset-password", `{`), "id", userID.String()), wantStatus: http.StatusBadRequest},
		{name: "reset password service error", handler: (*User).ResetPassword, lifecycle: &fakeUserLifecycle{resetErr: errDB}, req: withRouteParam(adminRequest(http.MethodPost, "/api/users/"+userID.String()+"/reset-password", `{"password":"newSecret123"}`), "id", userID.String()), wantStatus: http.StatusInternalServerError},
		{name: "profile link forbidden", handler: (*User).UpdateProfileLink, req: withRouteParam(plainRequest(http.MethodPatch, "/api/users/"+userID.String()+"/profile-link", `{}`), "id", userID.String()), wantStatus: http.StatusForbidden},
		{name: "profile link invalid id", handler: (*User).UpdateProfileLink, req: withRouteParam(adminRequest(http.MethodPatch, "/api/users/bad/profile-link", `{}`), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "profile link invalid json", handler: (*User).UpdateProfileLink, req: withRouteParam(adminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/profile-link", `{`), "id", userID.String()), wantStatus: http.StatusBadRequest},
		{name: "profile link invalid uuid", handler: (*User).UpdateProfileLink, req: withRouteParam(adminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/profile-link", `{"employee_id":"bad"}`), "id", userID.String()), wantStatus: http.StatusBadRequest},
		{name: "profile link service error", handler: (*User).UpdateProfileLink, lifecycle: &fakeUserLifecycle{profileErr: errDB}, req: withRouteParam(adminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/profile-link", `{"employee_id":"`+employeeID.String()+`"}`), "id", userID.String()), wantStatus: http.StatusInternalServerError},
		{name: "audit forbidden", handler: (*User).ListAuditLogs, req: plainRequest(http.MethodGet, "/api/settings/audit-logs", ""), wantStatus: http.StatusForbidden},
		{name: "audit store error", handler: (*User).ListAuditLogs, store: &fakeUserStore{auditErr: errDB}, req: adminRequest(http.MethodGet, "/api/settings/audit-logs", ""), wantStatus: http.StatusInternalServerError},
		{name: "create valid parent link surfaces role add error", handler: (*User).Create, store: &fakeUserStore{roleErr: errDB}, req: adminRequest(http.MethodPost, "/api/users", `{"username":"ortu","password":"secret123","roles":["ortu"],"parent_id":"`+parentID.String()+`"}`), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.store
			if store == nil {
				store = &fakeUserStore{}
			}
			lifecycle := tt.lifecycle
			if lifecycle == nil {
				lifecycle = &fakeUserLifecycle{}
			}
			rec := httptest.NewRecorder()
			tt.handler(&User{q: store, lifecycle: lifecycle}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
