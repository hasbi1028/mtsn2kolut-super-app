package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeRombelService struct {
	listCalled           bool
	getClassID           pgtype.UUID
	listStudentsClassID  pgtype.UUID
	listHomeroomClassID  pgtype.UUID
	createHomeroomArg    db.CreateHomeroomAssignmentParams
	updateHomeroomArg    db.UpdateHomeroomAssignmentParams
	deleteHomeroomID     pgtype.UUID
	updateIdentityArg    db.UpdateRombelIdentityParams
	matrixUpdateArg      service.SubjectAssignmentMatrixCellInput
	matrixCalled         bool
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
	f.listCalled = true
	return []db.ListRombelsRow{{ID: handlerTestUUID(199), Code: "VII.A", Name: "VII A", Level: "VII", IsActive: true}}, f.err
}

func (f *fakeRombelService) Get(_ context.Context, id pgtype.UUID) (db.GetRombelDetailRow, error) {
	f.getClassID = id
	return db.GetRombelDetailRow{ID: id, Code: "VII.A", Name: "VII A", Level: "VII", IsActive: true}, f.err
}

func (f *fakeRombelService) UpdateIdentity(_ context.Context, arg db.UpdateRombelIdentityParams) (db.UpdateRombelIdentityRow, error) {
	f.updateIdentityArg = arg
	if f.err != nil {
		return db.UpdateRombelIdentityRow{}, f.err
	}
	return db.UpdateRombelIdentityRow{
		ID:       arg.ID,
		Code:     arg.Code,
		Name:     arg.Name,
		Level:    arg.Level,
		IsActive: arg.IsActive,
	}, nil
}

func (f *fakeRombelService) ListStudentsWithParents(_ context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error) {
	f.listStudentsClassID = classID
	return []db.ListStudentsByClassWithParentsRow{{StudentID: handlerTestUUID(198), Nis: "123", StudentName: "Siswa Satu", Gender: db.GenderEnumL, IsActive: true, Status: db.StudentStatusEnumActive}}, f.err
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

func (f *fakeRombelService) GetSubjectAssignmentMatrix(context.Context) (service.SubjectAssignmentMatrix, error) {
	f.matrixCalled = true
	if f.err != nil {
		return service.SubjectAssignmentMatrix{}, f.err
	}
	return service.SubjectAssignmentMatrix{
		AcademicYearID:   handlerTestUUID(208),
		AcademicYearName: "2025/2026",
		Classes:          []db.ListSubjectAssignmentMatrixClassesRow{{ID: handlerTestUUID(209), Code: "VII.A", Name: "VII A", Level: "VII"}},
		Subjects:         []db.ListSubjectAssignmentMatrixSubjectsRow{{ID: handlerTestUUID(210), Code: "MTK", Name: "Matematika"}},
		Teachers:         []db.ListSubjectAssignmentMatrixTeachersRow{{ID: handlerTestUUID(211), Nama: "Guru Mapel"}},
		Cells:            []db.ListSubjectAssignmentMatrixCellsRow{},
	}, nil
}

func (f *fakeRombelService) UpdateSubjectAssignmentMatrixCell(_ context.Context, in service.SubjectAssignmentMatrixCellInput) (db.GetSubjectAssignmentMatrixCellRow, error) {
	f.matrixUpdateArg = in
	if f.err != nil {
		return db.GetSubjectAssignmentMatrixCellRow{}, f.err
	}
	return db.GetSubjectAssignmentMatrixCellRow{
		ClassID:           in.ClassID,
		SubjectID:         in.SubjectID,
		TeacherEmployeeID: in.TeacherEmployeeID,
		Status:            "complete",
	}, nil
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

func (f *fakeRombelService) ListHomeroomAssignments(_ context.Context, classID pgtype.UUID) ([]db.ListHomeroomAssignmentsByClassRow, error) {
	f.listHomeroomClassID = classID
	return []db.ListHomeroomAssignmentsByClassRow{{ID: handlerTestUUID(207), ClassID: classID, EmployeeID: handlerTestUUID(208), EmployeeName: "Wali Kelas", IsActive: true}}, f.err
}

func (f *fakeRombelService) CreateHomeroomAssignment(_ context.Context, arg db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error) {
	f.createHomeroomArg = arg
	return db.CreateHomeroomAssignmentRow{ID: handlerTestUUID(209), ClassID: arg.ClassID, EmployeeID: arg.EmployeeID, IsActive: arg.HomeroomIsActive, Notes: arg.Notes}, f.err
}

func (f *fakeRombelService) UpdateHomeroomAssignment(_ context.Context, arg db.UpdateHomeroomAssignmentParams) (db.UpdateHomeroomAssignmentRow, error) {
	f.updateHomeroomArg = arg
	return db.UpdateHomeroomAssignmentRow{ID: arg.ID, EmployeeID: arg.EmployeeID, IsActive: arg.HomeroomIsActive, Notes: arg.Notes}, f.err
}

func (f *fakeRombelService) DeleteHomeroomAssignment(_ context.Context, id pgtype.UUID) error {
	f.deleteHomeroomID = id
	return f.err
}

func TestRombelListGetAndStudentsSuccess(t *testing.T) {
	classID := handlerTestUUID(250)
	fake := &fakeRombelService{}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/academic/rombel", ""))
	if rec.Code != http.StatusOK || !fake.listCalled {
		t.Fatalf("List() status/called = %d/%v, want 200/true; body=%s", rec.Code, fake.listCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Get(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String(), ""), "id", classID.String()))
	if rec.Code != http.StatusOK || fake.getClassID != classID || fake.listStudentsClassID != classID || fake.listSubjectClassID != classID || fake.listTimetableClassID != classID || fake.listHomeroomClassID != classID {
		t.Fatalf("Get() status/ids = %d get=%s students=%s subjects=%s timetable=%s homeroom=%s, want all %s; body=%s", rec.Code, fake.getClassID.String(), fake.listStudentsClassID.String(), fake.listSubjectClassID.String(), fake.listTimetableClassID.String(), fake.listHomeroomClassID.String(), classID.String(), rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListStudents(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/students", ""), "id", classID.String()))
	if rec.Code != http.StatusOK || fake.listStudentsClassID != classID {
		t.Fatalf("ListStudents() status/class = %d/%s, want 200/%s; body=%s", rec.Code, fake.listStudentsClassID.String(), classID.String(), rec.Body.String())
	}
}

func TestRombelHomeroomAssignmentsSuccess(t *testing.T) {
	classID := handlerTestUUID(252)
	assignmentID := handlerTestUUID(253)
	employeeID := handlerTestUUID(254)
	academicYearID := handlerTestUUID(255)
	fake := &fakeRombelService{}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.ListHomeroomAssignments(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", ""), "id", classID.String()))
	if rec.Code != http.StatusOK || fake.listHomeroomClassID != classID {
		t.Fatalf("ListHomeroomAssignments() status/class = %d/%s, want 200/%s; body=%s", rec.Code, fake.listHomeroomClassID.String(), classID.String(), rec.Body.String())
	}

	body := `{"employee_id":"` + employeeID.String() + `","academic_year_id":"` + academicYearID.String() + `","start_date":"2025-07-01","end_date":"2026-06-30","is_active":false,"notes":" Wali utama "}`
	rec = httptest.NewRecorder()
	h.CreateHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homeroom-assignments", body), "id", classID.String()))
	if rec.Code != http.StatusCreated || fake.createHomeroomArg.ClassID != classID || fake.createHomeroomArg.EmployeeID != employeeID || fake.createHomeroomArg.HomeroomIsActive || fake.createHomeroomArg.Notes != "Wali utama" {
		t.Fatalf("CreateHomeroomAssignment() status/arg = %d/%+v, want 201 mapped payload", rec.Code, fake.createHomeroomArg)
	}

	body = `{"employee_id":"` + employeeID.String() + `","is_active":true,"notes":" Wali revisi "}`
	rec = httptest.NewRecorder()
	h.UpdateHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodPut, "/api/academic/homeroom-assignments/"+assignmentID.String(), body), "assignmentID", assignmentID.String()))
	if rec.Code != http.StatusOK || fake.updateHomeroomArg.ID != assignmentID || fake.updateHomeroomArg.EmployeeID != employeeID || !fake.updateHomeroomArg.HomeroomIsActive || fake.updateHomeroomArg.Notes != "Wali revisi" {
		t.Fatalf("UpdateHomeroomAssignment() status/arg = %d/%+v, want 200 mapped payload", rec.Code, fake.updateHomeroomArg)
	}

	rec = httptest.NewRecorder()
	h.DeleteHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/academic/homeroom-assignments/"+assignmentID.String(), ""), "assignmentID", assignmentID.String()))
	if rec.Code != http.StatusNoContent || fake.deleteHomeroomID != assignmentID {
		t.Fatalf("DeleteHomeroomAssignment() status/id = %d/%s, want 204/%s", rec.Code, fake.deleteHomeroomID.String(), assignmentID.String())
	}
}

func TestRombelUpdateIdentitySuccess(t *testing.T) {
	classID := handlerTestUUID(200)
	fake := &fakeRombelService{}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	body := `{"code":"VII.A","name":"VII A","level":"VII","is_active":true}`
	h.UpdateIdentity(rec, withRouteParam(adminRequest(http.MethodPut, "/api/academic/rombel/"+classID.String(), body), "id", classID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateIdentity() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateIdentityArg.ID != classID || fake.updateIdentityArg.Code != "VII.A" || fake.updateIdentityArg.Name != "VII A" || fake.updateIdentityArg.Level != "VII" || !fake.updateIdentityArg.IsActive {
		t.Fatalf("UpdateIdentity() arg = %+v, want forwarded rombel identity", fake.updateIdentityArg)
	}
}

func TestRombelUpdateIdentityValidationAndConflict(t *testing.T) {
	classID := handlerTestUUID(201)
	h := &Rombel{svc: &fakeRombelService{}}

	tests := []struct {
		name       string
		req        *http.Request
		svc        rombelService
		wantStatus int
	}{
		{
			name:       "unauthenticated",
			req:        withRouteParam(httptest.NewRequest(http.MethodPut, "/api/academic/rombel/"+classID.String(), nil), "id", classID.String()),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "ID data tidak valid",
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/academic/rombel/bad", `{}`), "id", "bad"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing is active",
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/academic/rombel/"+classID.String(), `{"code":"VII.A","name":"VII A","level":"VII"}`), "id", classID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service conflict",
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/academic/rombel/"+classID.String(), `{"code":"VII.A","name":"VII A","level":"VII","is_active":true}`), "id", classID.String()),
			svc:        &fakeRombelService{err: errors.Join(domain.ErrConflict, errors.New("kode rombel sudah dipakai"))},
			wantStatus: http.StatusConflict,
		},
	}
	for _, tt := range tests {
		if tt.svc != nil {
			h = &Rombel{svc: tt.svc}
		} else {
			h = &Rombel{svc: &fakeRombelService{}}
		}
		rec := httptest.NewRecorder()
		h.UpdateIdentity(rec, tt.req)
		if rec.Code != tt.wantStatus {
			t.Fatalf("%s: status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
		}
	}
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

	rec = httptest.NewRecorder()
	req := withRouteParams(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/subject-assignments/"+assignmentID.String(), ""), "id", classID.String(), "assignmentID", assignmentID.String())
	h.GetSubjectAssignment(rec, req)
	if rec.Code != http.StatusOK || fake.getSubjectArg.ClassID != classID || fake.getSubjectArg.ID != assignmentID {
		t.Fatalf("GetSubjectAssignment() status/arg = %d/%+v, want 200 scoped get arg", rec.Code, fake.getSubjectArg)
	}

	body := `{"subject_id":"` + subjectID.String() + `","teacher_employee_id":"` + teacherID.String() + `"}`
	rec = httptest.NewRecorder()
	h.CreateSubjectAssignment(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/subject-assignments", body), "id", classID.String()))
	if rec.Code != http.StatusCreated || fake.createSubjectArg.ClassID != classID || fake.createSubjectArg.SubjectID != subjectID || fake.createSubjectArg.TeacherEmployeeID != teacherID {
		t.Fatalf("CreateSubjectAssignment() status/arg = %d/%+v, want 201 class/subject/teacher", rec.Code, fake.createSubjectArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParams(adminRequest(http.MethodPut, "/api/academic/rombel/"+classID.String()+"/subject-assignments/"+assignmentID.String(), body), "id", classID.String(), "assignmentID", assignmentID.String())
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

func TestRombelSubjectAssignmentMatrixSuccess(t *testing.T) {
	classID := handlerTestUUID(214)
	subjectID := handlerTestUUID(215)
	teacherID := handlerTestUUID(216)
	fake := &fakeRombelService{}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.GetSubjectAssignmentMatrix(rec, adminRequest(http.MethodGet, "/api/academic/subject-assignment-matrix", ""))
	if rec.Code != http.StatusOK || !fake.matrixCalled {
		t.Fatalf("GetSubjectAssignmentMatrix() status/called = %d/%v, want 200/true; body=%s", rec.Code, fake.matrixCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	body := `{"class_id":"` + classID.String() + `","subject_id":"` + subjectID.String() + `","teacher_employee_id":"` + teacherID.String() + `"}`
	h.UpdateSubjectAssignmentMatrixCell(rec, adminRequest(http.MethodPut, "/api/academic/subject-assignment-matrix", body))
	if rec.Code != http.StatusOK || fake.matrixUpdateArg.ClassID != classID || fake.matrixUpdateArg.SubjectID != subjectID || fake.matrixUpdateArg.TeacherEmployeeID != teacherID {
		t.Fatalf("UpdateSubjectAssignmentMatrixCell() status/arg = %d/%+v, want 200 mapped ids", rec.Code, fake.matrixUpdateArg)
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
