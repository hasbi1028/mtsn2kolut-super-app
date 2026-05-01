package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeClassJournalService struct {
	*service.ClassJournal

	overviewAssignmentID pgtype.UUID
	overviewEmployeeID   pgtype.UUID
	overviewResult       service.JournalOverview
	overviewErr          error
	createAssignmentID   pgtype.UUID
	createTanggal        pgtype.Date
	createMateri         string
	createKegiatan       string
	createCatatan        string
	createGuruHadir      bool
	createEmployeeID     pgtype.UUID
	createResult         service.JournalSessionDetail
	createErr            error
	getID                pgtype.UUID
	getEmployeeID        pgtype.UUID
	getResult            service.JournalSessionDetail
	getErr               error
	updateID             pgtype.UUID
	updateMateri         string
	updateKegiatan       string
	updateCatatan        string
	updateGuruHadir      bool
	updateEmployeeID     pgtype.UUID
	updateResult         db.ClassJournalSession
	updateErr            error
	deleteID             pgtype.UUID
	deleteEmployeeID     pgtype.UUID
	deleteErr            error
	bulkID               pgtype.UUID
	bulkEntries          []service.JournalAttendanceEntry
	bulkEmployeeID       pgtype.UUID
	bulkErr              error
}

func (f *fakeClassJournalService) Overview(_ context.Context, assignmentID, employeeID pgtype.UUID) (service.JournalOverview, error) {
	f.overviewAssignmentID = assignmentID
	f.overviewEmployeeID = employeeID
	return f.overviewResult, f.overviewErr
}

func (f *fakeClassJournalService) CreateSession(_ context.Context, assignmentID pgtype.UUID, tanggal pgtype.Date, materi, kegiatan, catatan string, guruHadir bool, employeeID pgtype.UUID) (service.JournalSessionDetail, error) {
	f.createAssignmentID = assignmentID
	f.createTanggal = tanggal
	f.createMateri = materi
	f.createKegiatan = kegiatan
	f.createCatatan = catatan
	f.createGuruHadir = guruHadir
	f.createEmployeeID = employeeID
	return f.createResult, f.createErr
}

func (f *fakeClassJournalService) GetSession(_ context.Context, id, employeeID pgtype.UUID) (service.JournalSessionDetail, error) {
	f.getID = id
	f.getEmployeeID = employeeID
	return f.getResult, f.getErr
}

func (f *fakeClassJournalService) UpdateSession(_ context.Context, id pgtype.UUID, materi, kegiatan, catatan string, guruHadir bool, employeeID pgtype.UUID) (db.ClassJournalSession, error) {
	f.updateID = id
	f.updateMateri = materi
	f.updateKegiatan = kegiatan
	f.updateCatatan = catatan
	f.updateGuruHadir = guruHadir
	f.updateEmployeeID = employeeID
	return f.updateResult, f.updateErr
}

func (f *fakeClassJournalService) DeleteSession(_ context.Context, id, employeeID pgtype.UUID) error {
	f.deleteID = id
	f.deleteEmployeeID = employeeID
	return f.deleteErr
}

func (f *fakeClassJournalService) BulkUpsertAttendances(_ context.Context, sessionID pgtype.UUID, entries []service.JournalAttendanceEntry, employeeID pgtype.UUID) error {
	f.bulkID = sessionID
	f.bulkEntries = entries
	f.bulkEmployeeID = employeeID
	return f.bulkErr
}

func guruJournalRequest(method, target, body string, employeeID pgtype.UUID) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{
		"role":  "guru",
		"roles": []any{"guru"},
		"eid":   employeeID.String(),
		"usr":   "guru.ipa",
		"uid":   "01000000-0000-0000-0000-000000000000",
		"sub":   "01000000-0000-0000-0000-000000000000",
		"ssid":  "sess-journal-1",
	})
}

func TestClassJournalSuccessHandlersForwardPayloads(t *testing.T) {
	assignmentID := handlerTestUUID(255)
	sessionID := handlerTestUUID(1)
	employeeID := handlerTestUUID(2)
	studentID := handlerTestUUID(3)
	fake := &fakeClassJournalService{
		ClassJournal: &service.ClassJournal{},
		overviewResult: service.JournalOverview{
			Assignments: []db.ListClassSubjectAssignmentsRow{{ID: assignmentID, SubjectName: "IPA"}},
			Sessions:    []db.ListJournalSessionsRow{{ID: sessionID, AssignmentID: assignmentID, Materi: "Ekosistem"}},
			Summary:     []db.ListJournalAttendanceSummaryRow{{StudentID: studentID, Nama: "Siswa", Hadir: 1}},
		},
		createResult: service.JournalSessionDetail{
			Session:     db.GetJournalSessionRow{ID: sessionID, AssignmentID: assignmentID, Materi: "Ekosistem"},
			Attendances: []db.ListJournalAttendancesRow{{StudentID: studentID, Nama: "Siswa", Status: db.JournalAttendanceStatusHadir}},
		},
		getResult: service.JournalSessionDetail{
			Session:     db.GetJournalSessionRow{ID: sessionID, AssignmentID: assignmentID, Materi: "Ekosistem"},
			Attendances: []db.ListJournalAttendancesRow{{StudentID: studentID, Nama: "Siswa", Status: db.JournalAttendanceStatusHadir}},
		},
		updateResult: db.ClassJournalSession{ID: sessionID, AssignmentID: assignmentID, Materi: "Ekosistem revisi"},
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &ClassJournal{svc: fake, audit: audit}

	rec := httptest.NewRecorder()
	h.Overview(rec, guruJournalRequest(http.MethodGet, "/api/class-journal?assignment_id="+assignmentID.String(), "", employeeID))
	if rec.Code != http.StatusOK {
		t.Fatalf("Overview status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.overviewAssignmentID != assignmentID || fake.overviewEmployeeID != employeeID {
		t.Fatalf("Overview params = (%v, %v), want (%v, %v)", fake.overviewAssignmentID, fake.overviewEmployeeID, assignmentID, employeeID)
	}

	rec = httptest.NewRecorder()
	h.CreateSession(rec, guruJournalRequest(http.MethodPost, "/api/class-journal/sessions", `{"assignment_id":"`+assignmentID.String()+`","tanggal":"2026-05-01","materi":"Ekosistem","kegiatan":"Diskusi","catatan":"Aman","guru_hadir":true}`, employeeID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateSession status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createAssignmentID != assignmentID || fake.createEmployeeID != employeeID || fake.createMateri != "Ekosistem" || fake.createKegiatan != "Diskusi" || fake.createCatatan != "Aman" || !fake.createGuruHadir || !fake.createTanggal.Valid {
		t.Fatalf("CreateSession forwarded assignment=%v employee=%v date=%v materi=%q kegiatan=%q catatan=%q hadir=%v", fake.createAssignmentID, fake.createEmployeeID, fake.createTanggal, fake.createMateri, fake.createKegiatan, fake.createCatatan, fake.createGuruHadir)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "CLASS_JOURNAL_SESSION_CREATE" {
		t.Fatalf("CreateSession audit = %#v, want create audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.GetSession(rec, withRouteParam(guruJournalRequest(http.MethodGet, "/api/class-journal/sessions/"+sessionID.String(), "", employeeID), "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetSession status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.getID != sessionID || fake.getEmployeeID != employeeID {
		t.Fatalf("GetSession params = (%v, %v), want (%v, %v)", fake.getID, fake.getEmployeeID, sessionID, employeeID)
	}

	rec = httptest.NewRecorder()
	h.UpdateSession(rec, withRouteParam(guruJournalRequest(http.MethodPatch, "/api/class-journal/sessions/"+sessionID.String(), `{"materi":"Ekosistem revisi","kegiatan":"Praktik","catatan":"Lengkap","guru_hadir":true}`, employeeID), "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateSession status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateID != sessionID || fake.updateEmployeeID != employeeID || fake.updateMateri != "Ekosistem revisi" || fake.updateKegiatan != "Praktik" || fake.updateCatatan != "Lengkap" || !fake.updateGuruHadir {
		t.Fatalf("UpdateSession forwarded id=%v employee=%v materi=%q kegiatan=%q catatan=%q hadir=%v", fake.updateID, fake.updateEmployeeID, fake.updateMateri, fake.updateKegiatan, fake.updateCatatan, fake.updateGuruHadir)
	}
	if len(audit.entries) != 2 || audit.entries[1].Action != "CLASS_JOURNAL_SESSION_UPDATE" {
		t.Fatalf("UpdateSession audit = %#v, want update audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.BulkUpsertAttendances(rec, withRouteParam(guruJournalRequest(http.MethodPut, "/api/class-journal/sessions/"+sessionID.String()+"/attendances", `{"entries":[{"student_id":"`+studentID.String()+`","status":"hadir","catatan":"ok"}]}`, employeeID), "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("BulkUpsertAttendances status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.bulkID != sessionID || fake.bulkEmployeeID != employeeID || len(fake.bulkEntries) != 1 || fake.bulkEntries[0].StudentID != studentID.String() {
		t.Fatalf("BulkUpsertAttendances params = (%v, %v, %+v), want one forwarded entry", fake.bulkID, fake.bulkEmployeeID, fake.bulkEntries)
	}
	if len(audit.entries) != 3 || audit.entries[2].Action != "CLASS_JOURNAL_ATTENDANCE_BULK_UPSERT" {
		t.Fatalf("BulkUpsertAttendances audit = %#v, want bulk audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.DeleteSession(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/class-journal/sessions/"+sessionID.String(), ""), "id", sessionID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteSession status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID != sessionID || fake.deleteEmployeeID.Valid {
		t.Fatalf("DeleteSession params = (%v, %v), want admin delete with empty employee scope", fake.deleteID, fake.deleteEmployeeID)
	}
	if len(audit.entries) != 4 || audit.entries[3].Action != "CLASS_JOURNAL_SESSION_DELETE" {
		t.Fatalf("DeleteSession audit = %#v, want delete audit", audit.entries)
	}
}

func TestClassJournalHandlersMapServiceErrors(t *testing.T) {
	assignmentID := handlerTestUUID(4)
	sessionID := handlerTestUUID(5)
	employeeID := handlerTestUUID(6)
	tests := []struct {
		name       string
		handler    func(*ClassJournal, http.ResponseWriter, *http.Request)
		svc        *fakeClassJournalService
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "overview internal",
			handler:    (*ClassJournal).Overview,
			svc:        &fakeClassJournalService{ClassJournal: &service.ClassJournal{}, overviewErr: errors.New("db down")},
			req:        guruJournalRequest(http.MethodGet, "/api/class-journal?assignment_id="+assignmentID.String(), "", employeeID),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "create forbidden",
			handler:    (*ClassJournal).CreateSession,
			svc:        &fakeClassJournalService{ClassJournal: &service.ClassJournal{}, createErr: errors.New("akses ditolak")},
			req:        guruJournalRequest(http.MethodPost, "/api/class-journal/sessions", `{"assignment_id":"`+assignmentID.String()+`","tanggal":"2026-05-01"}`, employeeID),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "get internal",
			handler:    (*ClassJournal).GetSession,
			svc:        &fakeClassJournalService{ClassJournal: &service.ClassJournal{}, getErr: errors.New("db down")},
			req:        withRouteParam(guruJournalRequest(http.MethodGet, "/api/class-journal/sessions/"+sessionID.String(), "", employeeID), "id", sessionID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "update not found",
			handler:    (*ClassJournal).UpdateSession,
			svc:        &fakeClassJournalService{ClassJournal: &service.ClassJournal{}, updateErr: errors.New("sesi tidak ditemukan")},
			req:        withRouteParam(guruJournalRequest(http.MethodPatch, "/api/class-journal/sessions/"+sessionID.String(), `{}`, employeeID), "id", sessionID.String()),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "delete internal",
			handler:    (*ClassJournal).DeleteSession,
			svc:        &fakeClassJournalService{ClassJournal: &service.ClassJournal{}, deleteErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/api/class-journal/sessions/"+sessionID.String(), ""), "id", sessionID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "bulk bad request",
			handler:    (*ClassJournal).BulkUpsertAttendances,
			svc:        &fakeClassJournalService{ClassJournal: &service.ClassJournal{}, bulkErr: errors.New("student_id tidak valid")},
			req:        withRouteParam(guruJournalRequest(http.MethodPut, "/api/class-journal/sessions/"+sessionID.String()+"/attendances", `{"entries":[{"student_id":"bad","status":"hadir"}]}`, employeeID), "id", sessionID.String()),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&ClassJournal{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
