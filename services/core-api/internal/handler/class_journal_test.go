package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func TestNewClassJournalWiresOptionalAuditWriter(t *testing.T) {
	svc := &service.ClassJournal{}
	plain := NewClassJournal(svc)
	if plain == nil || plain.svc == nil || plain.audit != nil {
		t.Fatalf("NewClassJournal without audit = %#v, want service and nil audit", plain)
	}

	audit := &fakeCbtSessionAuditWriter{}
	withAudit := NewClassJournal(svc, audit)
	if withAudit == nil || withAudit.svc == nil || withAudit.audit != audit {
		t.Fatalf("NewClassJournal with audit = %#v, want provided audit", withAudit)
	}
}

func TestClassJournalOpenSessionFromTimetableSlotAcceptsLegacyTanggalAndDefaults(t *testing.T) {
	classID := handlerTestUUID(40)
	slotID := handlerTestUUID(41)
	sessionID := handlerTestUUID(42)
	assignmentID := handlerTestUUID(43)
	employeeID := handlerTestUUID(44)
	fake := &fakeClassJournalService{
		ClassJournal: &service.ClassJournal{},
		openResult: service.JournalSessionOpenResult{
			Session:       db.GetJournalSessionRow{ID: sessionID, AssignmentID: assignmentID},
			TimetableSlot: db.GetRombelTimetableSlotRow{ID: slotID, AssignmentID: assignmentID},
			Created:       false,
		},
	}
	h := &ClassJournal{svc: fake}
	req := guruJournalRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/timetable-slots/"+slotID.String()+"/journal-session", `{"tanggal":"2026-05-13","materi":"Materi langsung","catatan":"Catatan langsung"}`, employeeID)
	req = withRouteParams(req, "id", classID.String(), "slotID", slotID.String())
	rec := httptest.NewRecorder()

	h.OpenSessionFromTimetableSlot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("OpenSessionFromTimetableSlot status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.openClassID != classID || fake.openSlotID != slotID || fake.openEmployeeID != employeeID {
		t.Fatalf("forwarded ids class=%v slot=%v employee=%v, want %v/%v/%v", fake.openClassID, fake.openSlotID, fake.openEmployeeID, classID, slotID, employeeID)
	}
	if !fake.openTanggal.Valid || fake.openMateri != "Materi langsung" || fake.openCatatan != "Catatan langsung" || !fake.openGuruHadir {
		t.Fatalf("forwarded body date=%v materi=%q catatan=%q guru_hadir=%v, want legacy tanggal/direct fields/default true", fake.openTanggal, fake.openMateri, fake.openCatatan, fake.openGuruHadir)
	}
}

func TestClassJournalScopePermissionBranches(t *testing.T) {
	employeeID := handlerTestUUID(45)
	readAllReq := httptest.NewRequest(http.MethodGet, "/api/class-journal", nil)
	readAllReq = withClaims(readAllReq, jwt.MapClaims{"permissions": []any{"journal.read_all"}})
	if employee, ok := journalScope(readAllReq, false); !ok || employee.Valid {
		t.Fatalf("journalScope(read_all) = (%v, %v), want allowed all-scope", employee, ok)
	}

	manageReq := httptest.NewRequest(http.MethodPost, "/api/class-journal/sessions", nil)
	manageReq = withClaims(manageReq, jwt.MapClaims{"permissions": []any{"journal.manage"}, "eid": employeeID.String()})
	if employee, ok := journalScope(manageReq, true); !ok || employee != employeeID {
		t.Fatalf("journalScope(manage scoped) = (%v, %v), want employee scope", employee, ok)
	}

	deniedReq := httptest.NewRequest(http.MethodGet, "/api/class-journal", nil)
	deniedReq = withClaims(deniedReq, jwt.MapClaims{"roles": []any{"siswa"}, "eid": employeeID.String()})
	if employee, ok := journalScope(deniedReq, false); ok || employee.Valid {
		t.Fatalf("journalScope(disallowed role) = (%v, %v), want denied", employee, ok)
	}
}

func TestClassJournalHandlersRejectMalformedRequests(t *testing.T) {
	employeeID := handlerTestUUID(46)
	h := &ClassJournal{svc: &fakeClassJournalService{ClassJournal: &service.ClassJournal{}}}
	tests := []struct {
		name    string
		handle  func(http.ResponseWriter, *http.Request)
		req     *http.Request
		want    int
		message string
	}{
		{
			name:    "overview invalid assignment id",
			handle:  h.Overview,
			req:     guruJournalRequest(http.MethodGet, "/api/class-journal?assignment_id=bad", "", employeeID),
			want:    http.StatusBadRequest,
			message: "assignment_id tidak valid",
		},
		{
			name:    "create malformed json",
			handle:  h.CreateSession,
			req:     guruJournalRequest(http.MethodPost, "/api/class-journal/sessions", `{`, employeeID),
			want:    http.StatusBadRequest,
			message: "Data yang dikirim tidak valid",
		},
		{
			name:    "open missing date",
			handle:  h.OpenSessionFromTimetableSlot,
			req:     withRouteParams(guruJournalRequest(http.MethodPost, "/api/academic/rombel/"+handlerTestUUID(47).String()+"/timetable-slots/"+handlerTestUUID(48).String()+"/journal-session", `{}`, employeeID), "id", handlerTestUUID(47).String(), "slotID", handlerTestUUID(48).String()),
			want:    http.StatusBadRequest,
			message: "date wajib diisi",
		},
		{
			name:    "bulk malformed json",
			handle:  h.BulkUpsertAttendances,
			req:     withRouteParam(guruJournalRequest(http.MethodPut, "/api/class-journal/sessions/"+handlerTestUUID(49).String()+"/attendances", `{`, employeeID), "id", handlerTestUUID(49).String()),
			want:    http.StatusBadRequest,
			message: "Data yang dikirim tidak valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handle(rec, tt.req)
			if rec.Code != tt.want || !strings.Contains(rec.Body.String(), tt.message) {
				t.Fatalf("status/body = %d/%s, want %d containing %q", rec.Code, rec.Body.String(), tt.want, tt.message)
			}
		})
	}
}
