package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type SystemMaintenance struct {
	svc systemMaintenanceService
}

type systemMaintenanceService interface {
	Status(ctx context.Context) (service.SystemMaintenanceStatus, error)
	ListWindows(ctx context.Context, limit, offset int32) ([]service.MaintenanceWindowView, error)
	CreateWindow(ctx context.Context, in service.MaintenanceWindowInput) (service.MaintenanceWindowView, error)
	UpdateWindow(ctx context.Context, id pgtype.UUID, in service.MaintenanceWindowInput) (service.MaintenanceWindowView, error)
	ActivateWindow(ctx context.Context, id, actor pgtype.UUID, reason string) (service.MaintenanceWindowView, error)
	DeactivateWindow(ctx context.Context, id, actor pgtype.UUID, reason string) (service.MaintenanceWindowView, error)
	ListAuditLogs(ctx context.Context, filter service.MaintenanceAuditFilter) ([]service.MaintenanceAuditLogView, error)
	HealthSummary(ctx context.Context) (service.MaintenanceHealthSummary, error)
}

type maintenanceWindowRequest struct {
	Title            string     `json:"title"`
	Message          string     `json:"message"`
	Mode             string     `json:"mode"`
	AffectedModules  []string   `json:"affected_modules"`
	StartsAt         *time.Time `json:"starts_at"`
	EndsAt           *time.Time `json:"ends_at"`
	IsActive         bool       `json:"is_active"`
	AllowAdminBypass *bool      `json:"allow_admin_bypass"`
	BypassRoles      []string   `json:"bypass_roles"`
	Severity         string     `json:"severity"`
	Reason           string     `json:"reason"`
}

type maintenanceReasonRequest struct {
	Reason string `json:"reason"`
}

func NewSystemMaintenance(svc systemMaintenanceService) *SystemMaintenance {
	return &SystemMaintenance{svc: svc}
}

func (h *SystemMaintenance) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.svc.Status(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, status)
}

func (h *SystemMaintenance) HealthSummary(w http.ResponseWriter, r *http.Request) {
	status, err := h.svc.HealthSummary(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, status)
}

func (h *SystemMaintenance) ListWindows(w http.ResponseWriter, r *http.Request) {
	limit, offset := listLimitOffset(r, 50)
	items, err := h.svc.ListWindows(r.Context(), limit, offset)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"items": items})
}

func (h *SystemMaintenance) CreateWindow(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeMaintenanceWindowRequest(w, r)
	if !ok {
		return
	}
	item, err := h.svc.CreateWindow(r.Context(), req.toServiceInput(actorUserID(r)))
	if err != nil {
		writeDomainOrInternal(w, err, "maintenance tidak valid")
		return
	}
	api.Created(w, item)
}

func (h *SystemMaintenance) UpdateWindow(w http.ResponseWriter, r *http.Request) {
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	req, ok := decodeMaintenanceWindowRequest(w, r)
	if !ok {
		return
	}
	item, err := h.svc.UpdateWindow(r.Context(), id, req.toServiceInput(actorUserID(r)))
	if err != nil {
		writeDomainOrInternal(w, err, "maintenance tidak valid")
		return
	}
	api.OK(w, item)
}

func (h *SystemMaintenance) ActivateWindow(w http.ResponseWriter, r *http.Request) {
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	reason := decodeMaintenanceReason(r)
	item, err := h.svc.ActivateWindow(r.Context(), id, actorUserID(r), reason)
	if err != nil {
		writeDomainOrInternal(w, err, "maintenance tidak dapat diaktifkan")
		return
	}
	api.OK(w, item)
}

func (h *SystemMaintenance) DeactivateWindow(w http.ResponseWriter, r *http.Request) {
	id, ok := urlUUID(w, r, "id")
	if !ok {
		return
	}
	reason := decodeMaintenanceReason(r)
	item, err := h.svc.DeactivateWindow(r.Context(), id, actorUserID(r), reason)
	if err != nil {
		writeDomainOrInternal(w, err, "maintenance tidak dapat dinonaktifkan")
		return
	}
	api.OK(w, item)
}

func (h *SystemMaintenance) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := service.MaintenanceAuditFilter{
		Action: query.Get("action"),
		Limit:  int32Param(query.Get("limit"), 50),
		Offset: int32Param(query.Get("offset"), 0),
	}
	filter.FromAt = optionalTimeQuery(query.Get("from"))
	filter.ToAt = optionalTimeQuery(query.Get("to"))

	items, err := h.svc.ListAuditLogs(r.Context(), filter)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"items": items})
}

func decodeMaintenanceWindowRequest(w http.ResponseWriter, r *http.Request) (maintenanceWindowRequest, bool) {
	var req maintenanceWindowRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "payload maintenance tidak valid")
		return req, false
	}
	return req, true
}

func decodeMaintenanceReason(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	defer r.Body.Close()
	var req maintenanceReasonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		return ""
	}
	return strings.TrimSpace(req.Reason)
}

func (r maintenanceWindowRequest) toServiceInput(actor pgtype.UUID) service.MaintenanceWindowInput {
	return service.MaintenanceWindowInput{
		Title:            r.Title,
		Message:          r.Message,
		Mode:             r.Mode,
		AffectedModules:  r.AffectedModules,
		StartsAt:         r.StartsAt,
		EndsAt:           r.EndsAt,
		IsActive:         r.IsActive,
		AllowAdminBypass: r.AllowAdminBypass,
		BypassRoles:      r.BypassRoles,
		Severity:         r.Severity,
		ActorUserID:      actor,
		Reason:           r.Reason,
	}
}

func actorUserID(r *http.Request) pgtype.UUID {
	var id pgtype.UUID
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return id
	}
	raw, _ := claims["uid"].(string)
	if strings.TrimSpace(raw) == "" {
		raw, _ = claims["sub"].(string)
	}
	_ = id.Scan(raw)
	return id
}

func urlUUID(w http.ResponseWriter, r *http.Request, name string) (pgtype.UUID, bool) {
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, name)); err != nil || !id.Valid {
		api.BadRequest(w, "id maintenance tidak valid")
		return id, false
	}
	return id, true
}

func listLimitOffset(r *http.Request, defaultLimit int32) (int32, int32) {
	query := r.URL.Query()
	return int32Param(query.Get("limit"), defaultLimit), int32Param(query.Get("offset"), 0)
}

func int32Param(raw string, fallback int32) int32 {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return fallback
	}
	return int32(value)
}

func optionalTimeQuery(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &value
}
