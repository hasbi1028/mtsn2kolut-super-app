package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type RBACService interface {
	ListMatrix(ctx context.Context) (service.RBACMatrix, error)
	GetUserPermissions(ctx context.Context, userID pgtype.UUID) ([]string, error)
	GetUserRoles(ctx context.Context, userID pgtype.UUID) ([]string, error)
	ReplaceRolePermissions(ctx context.Context, roleCode string, permissionCodes []string, actorID pgtype.UUID) error
	ReplaceUserRoles(ctx context.Context, userID pgtype.UUID, roleCodes []string, actorID pgtype.UUID) error
}

type RBAC struct {
	svc RBACService
}

func NewRBAC(svc RBACService) *RBAC { return &RBAC{svc: svc} }

func (h *RBAC) ListMatrix(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
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
	if !adminAccessAllowed(r) {
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
	if !adminAccessAllowed(r) {
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

type updateRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func (h *RBAC) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	roleCode := strings.TrimSpace(chi.URLParam(r, "code"))
	if roleCode == "" {
		api.BadRequest(w, "role wajib diisi")
		return
	}
	var req updateRolePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "payload tidak valid")
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	userID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id user tidak valid")
		return
	}
	var req updateUserRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "payload tidak valid")
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
