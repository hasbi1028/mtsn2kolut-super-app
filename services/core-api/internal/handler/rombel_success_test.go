package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeRombelService struct {
	listSubjectClassID   pgtype.UUID
	getSubjectArg        db.GetRombelSubjectAssignmentParams
	createSubjectArg     db.CreateRombelSubjectAssignmentParams
	updateSubjectArg     db.UpdateRombelSubjectAssignmentParams
	deleteSubjectArg     db.DeleteRombelSubjectAssignmentParams
	listTimetableClassID pgtype.UUID
	getTimetableArg      db.GetRombelTimetableSlotParams
	createTimetableArg   db.CreateRombelTimetableSlotParams
	updateTimetableArg   db.UpdateRombelTimetableSlotParams
	deleteTimetableArg   db.DeleteRombelTimetableSlotParams
	err                  error
}

func (f *fakeRombelService) List(context.Context) ([]db.ListRombelsRow, error) {
	return []db.ListRombelsRow{}, f.err
}

func (f *fakeRombelService) Get(context.Context, pgtype.UUID) (db.GetRombelDetailRow, error) {
	return db.GetRombelDetailRow{}, f.err
}

func (f *fakeRombelService) ListStudentsWithParents(context.Context, pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error) {
	return []db.ListStudentsByClassWithParentsRow{}, f.err
}

func (f *fakeRombelService) ListSubjectAssignments(_ context.Context, classID pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error) {
	f.listSubjectClassID = classID
	if f.err != nil {
		return nil, f.err
	}
	return []db.ListRombelSubjectAssignmentsRow{{
		ID:                handlerTestUUID(201),
		ClassID:           classID,
		SubjectID:         handlerTestUUID(202),
		SubjectName:       "Matematika",
		SubjectCode:       "MTK",
		TeacherEmployeeID: handlerTestUUID(203),
		TeacherName:       "Guru Mapel",
	}}, nil
}

func (f *fakeRombelService) GetSubjectAssignment(_ context.Context, arg db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error) {
	f.getSubjectArg = arg
	if f.err != nil {
		return db.GetRombelSubjectAssignmentRow{}, f.err
	}
	return db.GetRombelSubjectAssignmentRow{ID: arg.ID, ClassID: arg.ClassID, SubjectName: "Matematika"}, nil
}

func (f *fakeRombelService) CreateSubjectAssignment(_ context.Context, arg db.CreateRombelSubjectAssignmentParams) (db.CreateRombelSubjectAssignmentRow, error) {
	f.createSubjectArg = arg
	if f.err != nil {
		return db.CreateRombelSubjectAssignmentRow{}, f.err
	}
	return db.CreateRombelSubjectAssignmentRow{
		ID:                handlerTestUUID(204),
		ClassID:           arg.ClassID,
		SubjectID:         arg.SubjectID,
		TeacherEmployeeID: arg.TeacherEmployeeID,
	}, nil
}

func (f *fakeRombelService) UpdateSubjectAssignment(_ context.Context, arg db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error) {
	f.updateSubjectArg = arg
	if f.err != nil {
		return db.UpdateRombelSubjectAssignmentRow{}, f.err
	}
	return db.UpdateRombelSubjectAssignmentRow{
		ID:                arg.ID,
		ClassID:           arg.ClassID,
		SubjectID:         arg.SubjectID,
		TeacherEmployeeID: arg.TeacherEmployeeID,
	}, nil
}

func (f *fakeRombelService) DeleteSubjectAssignment(_ context.Context, arg db.DeleteRombelSubjectAssignmentParams) error {
	f.deleteSubjectArg = arg
	return f.err
}

func (f *fakeRombelService) ListTimetableSlots(_ context.Context, classID pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error) {
	f.listTimetableClassID = classID
	if f.err != nil {
		return nil, f.err
	}
	return []db.ListRombelTimetableSlotsRow{{ID: handlerTestUUID(205), ClassID: classID, SubjectName: "Matematika"}}, nil
}

func (f *fakeRombelService) GetTimetableSlot(_ context.Context, arg db.GetRombelTimetableSlotParams) (db.GetRombelTimetableSlotRow, error) {
	f.getTimetableArg = arg
	if f.err != nil {
		return db.GetRombelTimetableSlotRow{}, f.err
	}
	return db.GetRombelTimetableSlotRow{ID: arg.ID, ClassID: arg.ClassID, SubjectName: "Matematika"}, nil
}

func (f *fakeRombelService) CreateTimetableSlot(_ context.Context, arg db.CreateRombelTimetableSlotParams) (db.CreateRombelTimetableSlotRow, error) {
	f.createTimetableArg = arg
	if f.err != nil {
		return db.CreateRombelTimetableSlotRow{}, f.err
	}
	return db.CreateRombelTimetableSlotRow{
		ID:           handlerTestUUID(206),
		ClassID:      arg.ClassID,
		AssignmentID: arg.AssignmentID,
		DayOfWeek:    arg.DayOfWeek,
		RoomLabel:    arg.RoomLabel,
		Notes:        arg.Notes,
	}, nil
}

func (f *fakeRombelService) UpdateTimetableSlot(_ context.Context, arg db.UpdateRombelTimetableSlotParams) (db.UpdateRombelTimetableSlotRow, error) {
	f.updateTimetableArg = arg
	if f.err != nil {
		return db.UpdateRombelTimetableSlotRow{}, f.err
	}
	return db.UpdateRombelTimetableSlotRow{
		ID:           arg.ID,
		ClassID:      arg.ClassID,
		AssignmentID: arg.AssignmentID,
		DayOfWeek:    arg.DayOfWeek,
		RoomLabel:    arg.RoomLabel,
		Notes:        arg.Notes,
	}, nil
}

func (f *fakeRombelService) DeleteTimetableSlot(_ context.Context, arg db.DeleteRombelTimetableSlotParams) error {
	f.deleteTimetableArg = arg
	return f.err
}

func (f *fakeRombelService) ListHomeroomAssignments(context.Context, pgtype.UUID) ([]db.ListHomeroomAssignmentsByClassRow, error) {
	return []db.ListHomeroomAssignmentsByClassRow{}, f.err
}

func (f *fakeRombelService) CreateHomeroomAssignment(context.Context, db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error) {
	return db.CreateHomeroomAssignmentRow{}, f.err
}

func (f *fakeRombelService) UpdateHomeroomAssignment(context.Context, db.UpdateHomeroomAssignmentParams) (db.UpdateHomeroomAssignmentRow, error) {
	return db.UpdateHomeroomAssignmentRow{}, f.err
}

func (f *fakeRombelService) DeleteHomeroomAssignment(context.Context, pgtype.UUID) error {
	return f.err
}

func TestRombelSubjectAssignmentsSuccess(t *testing.T) {
	classID := handlerTestUUID(210)
	assignmentID := handlerTestUUID(211)
	subjectID := handlerTestUUID(212)
	teacherID := handlerTestUUID(213)
	fake := &fakeRombelService{}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.ListSubjectAssignments(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/subject-assignments", ""), "id", classID.String()))
	if rec.Code != http.StatusOK || fake.listSubjectClassID != classID {
		t.Fatalf("ListSubjectAssignments() status/class = %d/%s, want 200/%s; body=%s", rec.Code, fake.listSubjectClassID.String(), classID.String(), rec.Body.String())
	}

	body := `{"subject_id":"` + subjectID.String() + `","teacher_employee_id":"` + teacherID.String() + `"}`
	rec = httptest.NewRecorder()
	h.CreateSubjectAssignment(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/subject-assignments", body), "id", classID.String()))
	if rec.Code != http.StatusCreated || fake.createSubjectArg.ClassID != classID || fake.createSubjectArg.SubjectID != subjectID || fake.createSubjectArg.TeacherEmployeeID != teacherID {
		t.Fatalf("CreateSubjectAssignment() status/arg = %d/%+v, want 201 class/subject/teacher", rec.Code, fake.createSubjectArg)
	}

	rec = httptest.NewRecorder()
	req := withRouteParams(adminRequest(http.MethodPut, "/api/academic/rombel/"+classID.String()+"/subject-assignments/"+assignmentID.String(), body), "id", classID.String(), "assignmentID", assignmentID.String())
	h.UpdateSubjectAssignment(rec, req)
	if rec.Code != http.StatusOK || fake.updateSubjectArg.ClassID != classID || fake.updateSubjectArg.ID != assignmentID || fake.updateSubjectArg.SubjectID != subjectID || fake.updateSubjectArg.TeacherEmployeeID != teacherID {
		t.Fatalf("UpdateSubjectAssignment() status/arg = %d/%+v, want 200 scoped update arg", rec.Code, fake.updateSubjectArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParams(adminRequest(http.MethodDelete, "/api/academic/rombel/"+classID.String()+"/subject-assignments/"+assignmentID.String(), ""), "id", classID.String(), "assignmentID", assignmentID.String())
	h.DeleteSubjectAssignment(rec, req)
	if rec.Code != http.StatusNoContent || fake.deleteSubjectArg.ClassID != classID || fake.deleteSubjectArg.ID != assignmentID {
		t.Fatalf("DeleteSubjectAssignment() status/arg = %d/%+v, want 204 scoped delete arg", rec.Code, fake.deleteSubjectArg)
	}
}

func TestRombelSubjectAssignmentsValidationAndAuth(t *testing.T) {
	classID := handlerTestUUID(220)
	assignmentID := handlerTestUUID(221)
	teacherID := handlerTestUUID(222)
	h := &Rombel{svc: &fakeRombelService{}}

	tests := []struct {
		name       string
		fn         func(http.ResponseWriter, *http.Request)
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "create unauthenticated returns 401",
			fn:         h.CreateSubjectAssignment,
			req:        withRouteParam(httptest.NewRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/subject-assignments", nil), "id", classID.String()),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "create invalid subject",
			fn:         h.CreateSubjectAssignment,
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/subject-assignments", `{"subject_id":"bad","teacher_employee_id":"`+teacherID.String()+`"}`), "id", classID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update invalid assignment id",
			fn:         h.UpdateSubjectAssignment,
			req:        withRouteParams(adminRequest(http.MethodPut, "/api/academic/rombel/"+classID.String()+"/subject-assignments/bad", `{}`), "id", classID.String(), "assignmentID", "bad"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete invalid class id",
			fn:         h.DeleteSubjectAssignment,
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/academic/rombel/bad/subject-assignments/"+assignmentID.String(), ""), "id", "bad", "assignmentID", assignmentID.String()),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestRombelTimetableSlotsSuccess(t *testing.T) {
	classID := handlerTestUUID(230)
	assignmentID := handlerTestUUID(231)
	slotID := handlerTestUUID(232)
	fake := &fakeRombelService{}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.ListTimetableSlots(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/timetable-slots", ""), "id", classID.String()))
	if rec.Code != http.StatusOK || fake.listTimetableClassID != classID {
		t.Fatalf("ListTimetableSlots() status/class = %d/%s, want 200/%s; body=%s", rec.Code, fake.listTimetableClassID.String(), classID.String(), rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req := withRouteParams(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/timetable-slots/"+slotID.String(), ""), "id", classID.String(), "slotID", slotID.String())
	h.GetTimetableSlot(rec, req)
	if rec.Code != http.StatusOK || fake.getTimetableArg.ClassID != classID || fake.getTimetableArg.ID != slotID {
		t.Fatalf("GetTimetableSlot() status/arg = %d/%+v, want 200 scoped get arg", rec.Code, fake.getTimetableArg)
	}

	body := `{"assignment_id":"` + assignmentID.String() + `","day_of_week":2,"start_time":"07:30","end_time":"08:50","room":"Lab IPA","notes":"Praktikum"}`
	rec = httptest.NewRecorder()
	h.CreateTimetableSlot(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/timetable-slots", body), "id", classID.String()))
	if rec.Code != http.StatusCreated || fake.createTimetableArg.ClassID != classID || fake.createTimetableArg.AssignmentID != assignmentID || fake.createTimetableArg.DayOfWeek != 2 || fake.createTimetableArg.RoomLabel != "Lab IPA" || fake.createTimetableArg.Notes != "Praktikum" {
		t.Fatalf("CreateTimetableSlot() status/arg = %d/%+v, want 201 scoped create arg", rec.Code, fake.createTimetableArg)
	}

	body = `{"assignment_id":"` + assignmentID.String() + `","day_of_week":3,"start_time":"09:00","end_time":"10:20","room_label":"Ruang 2","notes":"Ulangan"}`
	rec = httptest.NewRecorder()
	req = withRouteParams(adminRequest(http.MethodPut, "/api/academic/rombel/"+classID.String()+"/timetable-slots/"+slotID.String(), body), "id", classID.String(), "slotID", slotID.String())
	h.UpdateTimetableSlot(rec, req)
	if rec.Code != http.StatusOK || fake.updateTimetableArg.ClassID != classID || fake.updateTimetableArg.ID != slotID || fake.updateTimetableArg.AssignmentID != assignmentID || fake.updateTimetableArg.DayOfWeek != 3 || fake.updateTimetableArg.RoomLabel != "Ruang 2" {
		t.Fatalf("UpdateTimetableSlot() status/arg = %d/%+v, want 200 scoped update arg", rec.Code, fake.updateTimetableArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParams(adminRequest(http.MethodDelete, "/api/academic/rombel/"+classID.String()+"/timetable-slots/"+slotID.String(), ""), "id", classID.String(), "slotID", slotID.String())
	h.DeleteTimetableSlot(rec, req)
	if rec.Code != http.StatusNoContent || fake.deleteTimetableArg.ClassID != classID || fake.deleteTimetableArg.ID != slotID {
		t.Fatalf("DeleteTimetableSlot() status/arg = %d/%+v, want 204 scoped delete arg", rec.Code, fake.deleteTimetableArg)
	}
}

func TestRombelTimetableSlotsValidationAndAuth(t *testing.T) {
	classID := handlerTestUUID(240)
	assignmentID := handlerTestUUID(241)
	slotID := handlerTestUUID(242)
	h := &Rombel{svc: &fakeRombelService{}}

	tests := []struct {
		name       string
		fn         func(http.ResponseWriter, *http.Request)
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "create unauthenticated returns 401",
			fn:         h.CreateTimetableSlot,
			req:        withRouteParam(httptest.NewRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/timetable-slots", nil), "id", classID.String()),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "create invalid assignment",
			fn:         h.CreateTimetableSlot,
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/timetable-slots", `{"assignment_id":"bad","day_of_week":2,"start_time":"07:30","end_time":"08:50"}`), "id", classID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "create invalid day",
			fn:         h.CreateTimetableSlot,
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/timetable-slots", `{"assignment_id":"`+assignmentID.String()+`","day_of_week":7,"start_time":"07:30","end_time":"08:50"}`), "id", classID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "create invalid range",
			fn:         h.CreateTimetableSlot,
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/timetable-slots", `{"assignment_id":"`+assignmentID.String()+`","day_of_week":2,"start_time":"08:50","end_time":"07:30"}`), "id", classID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update invalid slot id",
			fn:         h.UpdateTimetableSlot,
			req:        withRouteParams(adminRequest(http.MethodPut, "/api/academic/rombel/"+classID.String()+"/timetable-slots/bad", `{}`), "id", classID.String(), "slotID", "bad"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete invalid class id",
			fn:         h.DeleteTimetableSlot,
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/academic/rombel/bad/timetable-slots/"+slotID.String(), ""), "id", "bad", "slotID", slotID.String()),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
