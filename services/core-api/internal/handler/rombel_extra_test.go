package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type rombelStudentsFake struct {
	*fakeRombelService
	rows []db.ListStudentsByClassWithParentsRow
	err  error
}

func (f *rombelStudentsFake) ListStudentsWithParents(_ context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error) {
	f.listStudentsClassID = classID
	return f.rows, f.err
}

func TestRombelExtraListStudentsGroupsParentsAndUsesRouteID(t *testing.T) {
	classID := handlerTestUUID(31)
	studentID := handlerTestUUID(32)
	fake := &rombelStudentsFake{
		fakeRombelService: &fakeRombelService{},
		rows: []db.ListStudentsByClassWithParentsRow{
			{StudentID: studentID, Nis: "001", StudentName: "Siswa", Gender: db.GenderEnumL, IsActive: true, Status: db.StudentStatusEnumActive, ParentID: handlerTestUUID(33), ParentNama: "Ayah", Relationship: "father", IsPrimaryContact: true},
			{StudentID: studentID, Nis: "001", StudentName: "Siswa", Gender: db.GenderEnumL, IsActive: true, Status: db.StudentStatusEnumActive, ParentID: handlerTestUUID(34), ParentNama: "Ibu", Relationship: "mother"},
		},
	}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.ListStudents(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String()+"/students", ""), "id", classID.String()))
	if rec.Code != http.StatusOK || fake.listStudentsClassID != classID {
		t.Fatalf("ListStudents status/id = %d/%s, want 200/%s; body=%s", rec.Code, fake.listStudentsClassID.String(), classID.String(), rec.Body.String())
	}
	var got struct {
		Data []rombelStudent `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("ListStudents response JSON error = %v; body=%s", err, rec.Body.String())
	}
	if len(got.Data) != 1 || got.Data[0].ID != studentID || len(got.Data[0].Parents) != 2 || got.Data[0].Parents[0].Nama != "Ayah" || got.Data[0].Parents[1].Relationship != "mother" {
		t.Fatalf("ListStudents grouped response = %+v, want one student with two parents", got)
	}
}

func TestRombelExtraGetStopsOnChildLoadError(t *testing.T) {
	classID := handlerTestUUID(35)
	fake := &rombelStudentsFake{fakeRombelService: &fakeRombelService{}, err: errors.New("students failed")}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.Get(rec, withRouteParam(adminRequest(http.MethodGet, "/api/academic/rombel/"+classID.String(), ""), "id", classID.String()))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Get child-load error status/body = %d/%s, want 500 with error", rec.Code, rec.Body.String())
	}
	if fake.listSubjectClassID.Valid || fake.listHomeroomClassID.Valid {
		t.Fatalf("Get continued after student error: subjectID=%v homeroomID=%v", fake.listSubjectClassID, fake.listHomeroomClassID)
	}
}

func TestRombelExtraSubjectMatrixValidationAndServiceErrors(t *testing.T) {
	fake := &fakeRombelService{}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.GetSubjectAssignmentMatrix(rec, adminRequest(http.MethodGet, "/api/academic/rombel/subject-assignment-matrix", ""))
	if rec.Code != http.StatusOK || !fake.matrixCalled {
		t.Fatalf("GetSubjectAssignmentMatrix status/called = %d/%v, want 200/true; body=%s", rec.Code, fake.matrixCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdateSubjectAssignmentMatrixCell(rec, adminRequest(http.MethodPatch, "/api/academic/rombel/subject-assignment-matrix/cell", `{"class_id":"bad","subject_id":"bad"}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UpdateSubjectAssignmentMatrixCell invalid class status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	classID := handlerTestUUID(36)
	subjectID := handlerTestUUID(37)
	fake.err = domain.ErrConflict
	rec = httptest.NewRecorder()
	h.UpdateSubjectAssignmentMatrixCell(rec, adminRequest(http.MethodPatch, "/api/academic/rombel/subject-assignment-matrix/cell", `{"class_id":"`+classID.String()+`","subject_id":"`+subjectID.String()+`","teacher_employee_id":null,"additional_weekly_hours":1.5,"customization_notes":" butuh tambahan "}`))
	if rec.Code != http.StatusConflict {
		t.Fatalf("UpdateSubjectAssignmentMatrixCell conflict status = %d, want 409; body=%s", rec.Code, rec.Body.String())
	}
	if fake.matrixUpdateArg.ClassID != classID || fake.matrixUpdateArg.SubjectID != subjectID || fake.matrixUpdateArg.TeacherEmployeeID.Valid || fake.matrixUpdateArg.AdditionalWeeklyHours == nil || *fake.matrixUpdateArg.AdditionalWeeklyHours != 1.5 {
		t.Fatalf("UpdateSubjectAssignmentMatrixCell arg = %+v, want parsed ids, nil teacher, additional hours", fake.matrixUpdateArg)
	}
}

func TestRombelExtraHomeroomValidationAndDeleteErrorMapping(t *testing.T) {
	classID := handlerTestUUID(38)
	assignmentID := handlerTestUUID(39)
	teacherID := handlerTestUUID(40)
	active := true
	fake := &fakeRombelService{}
	h := &Rombel{svc: fake}

	rec := httptest.NewRecorder()
	h.CreateHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/rombel/"+classID.String()+"/homerooms", `{"employee_id":"`+teacherID.String()+`","academic_year_id":"`+classID.String()+`","start_date":"2026-07-01","end_date":"2027-06-30","is_active":true,"notes":" wali "}`), "id", classID.String()))
	if rec.Code != http.StatusCreated || fake.createHomeroomArg.ClassID != classID || fake.createHomeroomArg.EmployeeID != teacherID || fake.createHomeroomArg.HomeroomIsActive != active || fake.createHomeroomArg.Notes != "wali" {
		t.Fatalf("CreateHomeroomAssignment status/arg = %d/%+v, want created forwarded payload; body=%s", rec.Code, fake.createHomeroomArg, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdateHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/academic/rombel/homerooms/"+assignmentID.String(), `{"employee_id":"bad","is_active":true}`), "assignmentID", assignmentID.String()))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UpdateHomeroomAssignment invalid employee status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	fake.err = domain.ErrConflict
	rec = httptest.NewRecorder()
	h.DeleteHomeroomAssignment(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/academic/rombel/homerooms/"+assignmentID.String(), ""), "assignmentID", assignmentID.String()))
	if rec.Code != http.StatusConflict || fake.deleteHomeroomID != assignmentID {
		t.Fatalf("DeleteHomeroomAssignment status/id = %d/%s, want 409/%s; body=%s", rec.Code, fake.deleteHomeroomID.String(), assignmentID.String(), rec.Body.String())
	}
}

var _ rombelService = (*rombelStudentsFake)(nil)
