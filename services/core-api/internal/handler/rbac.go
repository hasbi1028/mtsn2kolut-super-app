package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type RBACService interface {
	ListMatrix(ctx context.Context) (service.RBACMatrix, error)
	GetUserPermissions(ctx context.Context, userID pgtype.UUID) ([]string, error)
	GetUserRoles(ctx context.Context, userID pgtype.UUID) ([]string, error)
	CreateRole(ctx context.Context, input service.RBACRoleInput, actorID pgtype.UUID) (db.RbacRole, error)
	UpdateRole(ctx context.Context, code string, input service.RBACRoleInput, actorID pgtype.UUID) (db.RbacRole, error)
	SetRoleActive(ctx context.Context, code string, active bool, actorID pgtype.UUID) error
	CreatePermission(ctx context.Context, input service.RBACPermissionInput, actorID pgtype.UUID) (db.RbacPermission, error)
	UpdatePermission(ctx context.Context, code string, input service.RBACPermissionInput, actorID pgtype.UUID) (db.RbacPermission, error)
	SetPermissionActive(ctx context.Context, code string, active bool, actorID pgtype.UUID) error
	ReplaceRolePermissions(ctx context.Context, roleCode string, permissionCodes []string, actorID pgtype.UUID) error
	ReplaceUserRoles(ctx context.Context, userID pgtype.UUID, roleCodes []string, actorID pgtype.UUID) error
}

type RBAC struct {
	svc RBACService
}

func NewRBAC(svc RBACService) *RBAC { return &RBAC{svc: svc} }

func (h *RBAC) ListMatrix(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.read", "roles.manage") {
		api.Forbidden(w)
		return
	}
	matrix, err := h.svc.ListMatrix(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, matrix)
}

func (h *RBAC) ListRoles(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.read", "roles.manage") {
		api.Forbidden(w)
		return
	}
	matrix, err := h.svc.ListMatrix(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, matrix.Roles)
}

func (h *RBAC) ListPermissions(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.read", "roles.manage") {
		api.Forbidden(w)
		return
	}
	matrix, err := h.svc.ListMatrix(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, matrix.Permissions)
}

func (h *RBAC) CreateRole(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.manage") {
		api.Forbidden(w)
		return
	}
	var req service.RBACRoleInput
	if !decodeJSON(w, r, &req, 16<<10, disallowUnknownJSONFields) {
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	role, err := h.svc.CreateRole(r.Context(), req, actorID)
	if err != nil {
		writeClientError(w, err, "role tidak dapat dibuat")
		return
	}
	api.Created(w, role)
}

func (h *RBAC) UpdateRole(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.manage") {
		api.Forbidden(w)
		return
	}
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if code == "" {
		api.BadRequest(w, "role wajib diisi")
		return
	}
	var req service.RBACRoleInput
	if !decodeJSON(w, r, &req, 16<<10, disallowUnknownJSONFields) {
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	role, err := h.svc.UpdateRole(r.Context(), code, req, actorID)
	if err != nil {
		writeClientError(w, err, "role tidak dapat diperbarui")
		return
	}
	api.OK(w, role)
}

type updateRBACStatusRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *RBAC) SetRoleStatus(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.manage") {
		api.Forbidden(w)
		return
	}
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if code == "" {
		api.BadRequest(w, "role wajib diisi")
		return
	}
	var req updateRBACStatusRequest
	if !decodeJSON(w, r, &req, 8<<10, disallowUnknownJSONFields) {
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	if err := h.svc.SetRoleActive(r.Context(), code, req.IsActive, actorID); err != nil {
		writeClientError(w, err, "status role tidak dapat diperbarui")
		return
	}
	api.OK(w, map[string]any{"ok": true})
}

func (h *RBAC) CreatePermission(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.manage") {
		api.Forbidden(w)
		return
	}
	var req service.RBACPermissionInput
	if !decodeJSON(w, r, &req, 16<<10, disallowUnknownJSONFields) {
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	permission, err := h.svc.CreatePermission(r.Context(), req, actorID)
	if err != nil {
		writeClientError(w, err, "permission tidak dapat dibuat")
		return
	}
	api.Created(w, permission)
}

func (h *RBAC) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.manage") {
		api.Forbidden(w)
		return
	}
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if code == "" {
		api.BadRequest(w, "permission wajib diisi")
		return
	}
	var req service.RBACPermissionInput
	if !decodeJSON(w, r, &req, 16<<10, disallowUnknownJSONFields) {
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	permission, err := h.svc.UpdatePermission(r.Context(), code, req, actorID)
	if err != nil {
		writeClientError(w, err, "permission tidak dapat diperbarui")
		return
	}
	api.OK(w, permission)
}

func (h *RBAC) SetPermissionStatus(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.manage") {
		api.Forbidden(w)
		return
	}
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if code == "" {
		api.BadRequest(w, "permission wajib diisi")
		return
	}
	var req updateRBACStatusRequest
	if !decodeJSON(w, r, &req, 8<<10, disallowUnknownJSONFields) {
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	if err := h.svc.SetPermissionActive(r.Context(), code, req.IsActive, actorID); err != nil {
		writeClientError(w, err, "status permission tidak dapat diperbarui")
		return
	}
	api.OK(w, map[string]any{"ok": true})
}

type updateRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func (h *RBAC) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "roles.manage") {
		api.Forbidden(w)
		return
	}
	roleCode := strings.TrimSpace(chi.URLParam(r, "code"))
	if roleCode == "" {
		api.BadRequest(w, "role wajib diisi")
		return
	}
	var req updateRolePermissionsRequest
	if !decodeJSON(w, r, &req, 16<<10, disallowUnknownJSONFields) {
		return
	}
	if len(req.Permissions) == 0 {
		api.BadRequest(w, "minimal satu permission wajib dipilih")
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	if err := h.svc.ReplaceRolePermissions(r.Context(), roleCode, req.Permissions, actorID); err != nil {
		writeClientError(w, err, "permission role tidak dapat diperbarui")
		return
	}
	api.OK(w, map[string]any{"ok": true})
}

type updateUserRolesRequest struct {
	Roles []string `json:"roles"`
}

func (h *RBAC) UpdateUserRoles(w http.ResponseWriter, r *http.Request) {
	if !rbacAccessAllowed(r, "users.manage_roles") {
		api.Forbidden(w)
		return
	}
	userID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id user tidak valid")
		return
	}
	var req updateUserRolesRequest
	if !decodeJSON(w, r, &req, 16<<10, disallowUnknownJSONFields) {
		return
	}
	if len(req.Roles) == 0 {
		api.BadRequest(w, "minimal satu role wajib dipilih")
		return
	}
	actorID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	if err := h.svc.ReplaceUserRoles(r.Context(), userID, req.Roles, actorID); err != nil {
		writeClientError(w, err, "role user tidak dapat diperbarui")
		return
	}
	api.OK(w, map[string]any{"ok": true})
}

func rbacAccessAllowed(r *http.Request, permissions ...string) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin") || mw.HasAnyPermission(claims, permissions...)
}

func currentActorUUID(r *http.Request) (pgtype.UUID, error) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return pgtype.UUID{}, http.ErrNoCookie
	}
	for _, key := range []string{"uid", "sub"} {
		if raw, ok := claims[key].(string); ok && strings.TrimSpace(raw) != "" {
			return parseUUID(raw)
		}
	}
	return pgtype.UUID{}, http.ErrNoCookie
}
