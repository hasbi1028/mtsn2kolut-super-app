package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func TestInventoryExtraSchoolRoomDefaultsAndErrorMapping(t *testing.T) {
	roomID := handlerTestUUID(41)
	fake := &fakeInventoryService{Inventory: &service.Inventory{}, roomRow: db.SchoolRoom{ID: roomID, Code: "R-1", Name: "Ruang 1"}}
	h := &Inventory{svc: fake}

	rec := httptest.NewRecorder()
	h.CreateSchoolRoom(rec, adminRequest(http.MethodPost, "/api/inventory/school-rooms", `{"code":" R-1 ","name":" Ruang 1 ","room_type":" kelas ","condition":" baik "}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateSchoolRoom status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.roomCreate.Code != "R-1" || fake.roomCreate.Name != "Ruang 1" || fake.roomCreate.RoomType != "kelas" || fake.roomCreate.Condition != "baik" || !fake.roomCreate.IsExamEligible {
		t.Fatalf("CreateSchoolRoom arg = %+v, want trimmed fields and default exam eligible true", fake.roomCreate)
	}

	fake.roomErr = domain.ErrConflict
	rec = httptest.NewRecorder()
	h.UpdateSchoolRoom(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/inventory/school-rooms/"+roomID.String(), `{"code":"R-1","name":"Ruang 1","is_exam_eligible":false}`), "id", roomID.String()))
	if rec.Code != http.StatusConflict || fake.roomUpdate.ID != roomID || fake.roomUpdate.IsExamEligible {
		t.Fatalf("UpdateSchoolRoom status/arg = %d/%+v, want 409, route id, explicit exam eligible false; body=%s", rec.Code, fake.roomUpdate, rec.Body.String())
	}

	fake.roomErr = nil
	fake.roomDeleteErr = domain.ErrConflict
	rec = httptest.NewRecorder()
	h.DeleteSchoolRoom(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/inventory/school-rooms/"+roomID.String(), ""), "id", roomID.String()))
	if rec.Code != http.StatusConflict || fake.roomID != roomID {
		t.Fatalf("DeleteSchoolRoom status/id = %d/%s, want 409/%s; body=%s", rec.Code, fake.roomID.String(), roomID.String(), rec.Body.String())
	}
}

func TestInventoryExtraSchoolRoomInvalidRequestsAndInternalErrors(t *testing.T) {
	roomID := handlerTestUUID(42)
	boom := errors.New("room backend failed")
	fake := &fakeInventoryService{Inventory: &service.Inventory{}, roomErr: boom}
	h := &Inventory{svc: fake}

	tests := []struct {
		name       string
		fn         http.HandlerFunc
		req        *http.Request
		wantStatus int
	}{
		{name: "list service error", fn: h.ListSchoolRooms, req: adminRequest(http.MethodGet, "/api/inventory/school-rooms?search=x", ""), wantStatus: http.StatusInternalServerError},
		{name: "get invalid id", fn: h.GetSchoolRoom, req: withRouteParam(adminRequest(http.MethodGet, "/api/inventory/school-rooms/bad", ""), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "get service error", fn: h.GetSchoolRoom, req: withRouteParam(adminRequest(http.MethodGet, "/api/inventory/school-rooms/"+roomID.String(), ""), "id", roomID.String()), wantStatus: http.StatusInternalServerError},
		{name: "create invalid json", fn: h.CreateSchoolRoom, req: adminRequest(http.MethodPost, "/api/inventory/school-rooms", `{`), wantStatus: http.StatusBadRequest},
		{name: "update invalid id", fn: h.UpdateSchoolRoom, req: withRouteParam(adminRequest(http.MethodPatch, "/api/inventory/school-rooms/bad", `{}`), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "update invalid json", fn: h.UpdateSchoolRoom, req: withRouteParam(adminRequest(http.MethodPatch, "/api/inventory/school-rooms/"+roomID.String(), `{`), "id", roomID.String()), wantStatus: http.StatusBadRequest},
		{name: "delete invalid id", fn: h.DeleteSchoolRoom, req: withRouteParam(adminRequest(http.MethodDelete, "/api/inventory/school-rooms/bad", ""), "id", "bad"), wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestInventoryExtraSchoolRoomForbiddenDoesNotCallService(t *testing.T) {
	fake := &fakeInventoryService{Inventory: &service.Inventory{}, roomRows: []db.SchoolRoom{{ID: pgtype.UUID{Valid: true}}}}
	h := &Inventory{svc: fake}
	req := httptest.NewRequest(http.MethodGet, "/api/inventory/school-rooms", nil)
	rec := httptest.NewRecorder()
	h.ListSchoolRooms(rec, req)
	if rec.Code != http.StatusForbidden || fake.roomSearch != "" || len(fake.roomRows) != 1 {
		t.Fatalf("ListSchoolRooms forbidden status/search = %d/%q, want 403/no service call", rec.Code, fake.roomSearch)
	}
}
