package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeSystemMaintenanceService struct {
	status service.SystemMaintenanceStatus
	health service.MaintenanceHealthSummary
	items  []service.MaintenanceWindowView
	audits []service.MaintenanceAuditLogView
	item   service.MaintenanceWindowView
	err    error

	listLimit        int32
	listOffset       int32
	createIn         service.MaintenanceWindowInput
	updateID         pgtype.UUID
	updateIn         service.MaintenanceWindowInput
	activateID       pgtype.UUID
	activateBy       pgtype.UUID
	activateReason   string
	deactivateID     pgtype.UUID
	deactivateBy     pgtype.UUID
	deactivateReason string
	auditFilter      service.MaintenanceAuditFilter
}

func (f *fakeSystemMaintenanceService) Status(context.Context) (service.SystemMaintenanceStatus, error) {
	return f.status, f.err
}
func (f *fakeSystemMaintenanceService) HealthSummary(context.Context) (service.MaintenanceHealthSummary, error) {
	return f.health, f.err
}
func (f *fakeSystemMaintenanceService) ListWindows(_ context.Context, limit, offset int32) ([]service.MaintenanceWindowView, error) {
	f.listLimit, f.listOffset = limit, offset
	return f.items, f.err
}
func (f *fakeSystemMaintenanceService) CreateWindow(_ context.Context, in service.MaintenanceWindowInput) (service.MaintenanceWindowView, error) {
	f.createIn = in
	return f.item, f.err
}
func (f *fakeSystemMaintenanceService) UpdateWindow(_ context.Context, id pgtype.UUID, in service.MaintenanceWindowInput) (service.MaintenanceWindowView, error) {
	f.updateID, f.updateIn = id, in
	return f.item, f.err
}
func (f *fakeSystemMaintenanceService) ActivateWindow(_ context.Context, id, actor pgtype.UUID, reason string) (service.MaintenanceWindowView, error) {
	f.activateID, f.activateBy, f.activateReason = id, actor, reason
	return f.item, f.err
}
func (f *fakeSystemMaintenanceService) DeactivateWindow(_ context.Context, id, actor pgtype.UUID, reason string) (service.MaintenanceWindowView, error) {
	f.deactivateID, f.deactivateBy, f.deactivateReason = id, actor, reason
	return f.item, f.err
}
func (f *fakeSystemMaintenanceService) ListAuditLogs(_ context.Context, filter service.MaintenanceAuditFilter) ([]service.MaintenanceAuditLogView, error) {
	f.auditFilter = filter
	return f.audits, f.err
}

func maintenanceAdminRequest(method, target, body string, userID pgtype.UUID) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "uid": userID.String(), "sub": userID.String()})
}

func TestSystemMaintenanceEndpointSuccesses(t *testing.T) {
	windowID := handlerTestUUID(201)
	actorID := handlerTestUUID(202)
	starts := time.Date(2026, 5, 17, 20, 0, 0, 0, time.UTC)
	ends := starts.Add(time.Hour)
	view := service.MaintenanceWindowView{ID: windowID.String(), Title: "Upgrade", Message: "Maintenance", Mode: "readonly", AffectedModules: []string{"cbt"}, StartsAt: &starts, EndsAt: &ends, IsActive: true, AllowAdminBypass: true, Severity: service.MaintenanceSeverityWarning, Status: "active", CreatedAt: starts, UpdatedAt: starts}
	fake := &fakeSystemMaintenanceService{
		status: service.SystemMaintenanceStatus{Active: true, Mode: "readonly", ServerTime: starts, Window: &view},
		health: service.MaintenanceHealthSummary{CoreAPI: service.MaintenanceHealthComponent{Status: "ok", ServerTime: starts}},
		items:  []service.MaintenanceWindowView{view},
		item:   view,
		audits: []service.MaintenanceAuditLogView{{ID: handlerTestUUID(203).String(), MaintenanceID: windowID.String(), Action: "activate", CreatedAt: starts}},
	}
	h := NewSystemMaintenance(fake)

	for _, tc := range []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		req  *http.Request
		want int
	}{
		{"status", h.Status, httptest.NewRequest(http.MethodGet, "/api/system/maintenance/status", nil), http.StatusOK},
		{"health", h.HealthSummary, httptest.NewRequest(http.MethodGet, "/api/system/maintenance/health", nil), http.StatusOK},
		{"list", h.ListWindows, httptest.NewRequest(http.MethodGet, "/api/system/maintenance/windows?limit=7&offset=3", nil), http.StatusOK},
		{"create", h.CreateWindow, maintenanceAdminRequest(http.MethodPost, "/api/system/maintenance/windows", `{"title":"Upgrade","message":"Maintenance","mode":"readonly","affected_modules":["cbt"],"starts_at":"2026-05-17T20:00:00Z","ends_at":"2026-05-17T21:00:00Z","is_active":true,"allow_admin_bypass":false,"bypass_roles":["admin"],"severity":"warning","reason":"release"}`, actorID), http.StatusCreated},
		{"update", h.UpdateWindow, withRouteParam(maintenanceAdminRequest(http.MethodPatch, "/api/system/maintenance/windows/"+windowID.String(), `{"title":"Upgrade 2","message":"Maintenance","mode":"readonly","severity":"info","reason":"edit"}`, actorID), "id", windowID.String()), http.StatusOK},
		{"activate", h.ActivateWindow, withRouteParam(maintenanceAdminRequest(http.MethodPost, "/api/system/maintenance/windows/"+windowID.String()+"/activate", `{"reason":"go"}`, actorID), "id", windowID.String()), http.StatusOK},
		{"deactivate", h.DeactivateWindow, withRouteParam(maintenanceAdminRequest(http.MethodPost, "/api/system/maintenance/windows/"+windowID.String()+"/deactivate", `{"reason":"done"}`, actorID), "id", windowID.String()), http.StatusOK},
		{"audits", h.ListAuditLogs, httptest.NewRequest(http.MethodGet, "/api/system/maintenance/audit-logs?action=activate&limit=9&offset=2&from=2026-05-17T00:00:00Z&to=2026-05-18T00:00:00Z", nil), http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.fn(rec, tc.req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d want %d body=%s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
	if fake.listLimit != 7 || fake.listOffset != 3 {
		t.Fatalf("ListWindows args = %d/%d", fake.listLimit, fake.listOffset)
	}
	if fake.createIn.Title != "Upgrade" || fake.createIn.ActorUserID != actorID || fake.createIn.AllowAdminBypass == nil || *fake.createIn.AllowAdminBypass {
		t.Fatalf("CreateWindow input = %+v", fake.createIn)
	}
	if fake.updateID != windowID || fake.updateIn.Title != "Upgrade 2" || fake.updateIn.ActorUserID != actorID {
		t.Fatalf("UpdateWindow input id=%v in=%+v", fake.updateID, fake.updateIn)
	}
	if fake.activateID != windowID || fake.activateBy != actorID || fake.activateReason != "go" {
		t.Fatalf("ActivateWindow args = %v %v %q", fake.activateID, fake.activateBy, fake.activateReason)
	}
	if fake.deactivateID != windowID || fake.deactivateBy != actorID || fake.deactivateReason != "done" {
		t.Fatalf("DeactivateWindow args = %v %v %q", fake.deactivateID, fake.deactivateBy, fake.deactivateReason)
	}
	if fake.auditFilter.Action != "activate" || fake.auditFilter.Limit != 9 || fake.auditFilter.Offset != 2 || fake.auditFilter.FromAt == nil || fake.auditFilter.ToAt == nil {
		t.Fatalf("audit filter = %+v", fake.auditFilter)
	}
}

func TestSystemMaintenanceValidationAndErrors(t *testing.T) {
	windowID := handlerTestUUID(204)
	h := NewSystemMaintenance(&fakeSystemMaintenanceService{err: domain.ErrBadRequest})
	for _, tc := range []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		req  *http.Request
		want int
	}{
		{"create invalid json", h.CreateWindow, httptest.NewRequest(http.MethodPost, "/api/system/maintenance/windows", strings.NewReader(`{`)), http.StatusBadRequest},
		{"update invalid id", h.UpdateWindow, withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/system/maintenance/windows/bad", strings.NewReader(`{}`)), "id", "bad"), http.StatusBadRequest},
		{"activate invalid id", h.ActivateWindow, withRouteParam(httptest.NewRequest(http.MethodPost, "/api/system/maintenance/windows/bad/activate", strings.NewReader(`{}`)), "id", "bad"), http.StatusBadRequest},
		{"deactivate invalid id", h.DeactivateWindow, withRouteParam(httptest.NewRequest(http.MethodPost, "/api/system/maintenance/windows/bad/deactivate", strings.NewReader(`{}`)), "id", "bad"), http.StatusBadRequest},
		{"create service bad request", h.CreateWindow, httptest.NewRequest(http.MethodPost, "/api/system/maintenance/windows", strings.NewReader(`{"title":"x"}`)), http.StatusBadRequest},
		{"update service bad request", h.UpdateWindow, withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/system/maintenance/windows/"+windowID.String(), strings.NewReader(`{"title":"x"}`)), "id", windowID.String()), http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.fn(rec, tc.req)
			if rec.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}

	errHandler := NewSystemMaintenance(&fakeSystemMaintenanceService{err: errors.New("db down")})
	for _, tc := range []struct {
		name string
		fn   func(*SystemMaintenance, http.ResponseWriter, *http.Request)
		req  *http.Request
	}{
		{"status", (*SystemMaintenance).Status, httptest.NewRequest(http.MethodGet, "/", nil)},
		{"health", (*SystemMaintenance).HealthSummary, httptest.NewRequest(http.MethodGet, "/", nil)},
		{"list", (*SystemMaintenance).ListWindows, httptest.NewRequest(http.MethodGet, "/", nil)},
		{"audits", (*SystemMaintenance).ListAuditLogs, httptest.NewRequest(http.MethodGet, "/", nil)},
	} {
		t.Run(tc.name+" internal", func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.fn(errHandler, rec, tc.req)
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
			}
		})
	}
}
