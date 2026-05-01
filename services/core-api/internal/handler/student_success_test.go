package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeStudentService struct {
	listCalled bool

	listTeacherID pgtype.UUID

	createArg db.CreateStudentParams
	updateArg db.UpdateStudentParams
	deleteID  pgtype.UUID

	lifecycleID     pgtype.UUID
	lifecycleStatus db.StudentStatusEnum
}

func (f *fakeStudentService) List(ctx context.Context) ([]db.ListStudentsRow, error) {
	f.listCalled = true
	return []db.ListStudentsRow{}, nil
}

func (f *fakeStudentService) ListByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListStudentsByTeacherRow, error) {
	f.listTeacherID = teacherEmployeeID
	return []db.ListStudentsByTeacherRow{}, nil
}

func (f *fakeStudentService) Create(ctx context.Context, p db.CreateStudentParams) (db.Student, error) {
	f.createArg = p
	return db.Student{ID: handlerTestUUID(220), Nis: p.Nis, Nama: p.Nama, Status: p.Status}, nil
}

func (f *fakeStudentService) Update(ctx context.Context, p db.UpdateStudentParams) (db.Student, error) {
	f.updateArg = p
	return db.Student{ID: p.ID, Nis: p.Nis, Nama: p.Nama, Status: p.Status}, nil
}

func (f *fakeStudentService) Delete(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return nil
}

func (f *fakeStudentService) UpdateLifecycle(ctx context.Context, id pgtype.UUID, status db.StudentStatusEnum) error {
	f.lifecycleID = id
	f.lifecycleStatus = status
	return nil
}

func guruStudentRequest(method, target, body string, employeeID pgtype.UUID) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
		"roles": []any{"guru"},
		"eid":   employeeID.String(),
	}))
}

func TestStudentSuccessHandlersForwardPayloads(t *testing.T) {
	studentID := handlerTestUUID(221)
	classID := handlerTestUUID(222)
	teacherID := handlerTestUUID(223)
	svc := &fakeStudentService{}
	h := &Student{svc: svc}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/students", ""))
	if rec.Code != http.StatusOK || !svc.listCalled {
		t.Fatalf("List() status/called = %d/%v, want 200/true", rec.Code, svc.listCalled)
	}

	rec = httptest.NewRecorder()
	h.GuruAwareList(rec, guruStudentRequest(http.MethodGet, "/api/students", "", teacherID))
	if rec.Code != http.StatusOK || svc.listTeacherID != teacherID {
		t.Fatalf("GuruAwareList() status/teacher = %d/%v, want 200/%v", rec.Code, svc.listTeacherID, teacherID)
	}

	body := `{"nis":"123","nisn":"456","nama":"Siswa A","gender":"L","parent_name":"Ortu","parent_phone":"0812","class_id":"` + classID.String() + `","is_active":true,"status":"active"}`
	rec = httptest.NewRecorder()
	h.Create(rec, adminRequest(http.MethodPost, "/api/students", body))
	if rec.Code != http.StatusCreated || svc.createArg.Nis != "123" || svc.createArg.ClassID != classID || svc.createArg.Status != db.StudentStatusEnumActive {
		t.Fatalf("Create() status/arg = %d/%+v", rec.Code, svc.createArg)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodPut, "/api/students/"+studentID.String(), body), "id", studentID.String())
	h.Update(rec, req)
	if rec.Code != http.StatusOK || svc.updateArg.ID != studentID || svc.updateArg.ClassID != classID || svc.updateArg.Nama != "Siswa A" {
		t.Fatalf("Update() status/arg = %d/%+v", rec.Code, svc.updateArg)
	}

	rec = httptest.NewRecorder()
	h.PublicRegister(rec, httptest.NewRequest(http.MethodPost, "/api/public/register-student", strings.NewReader(`{"nis":"789","nama":"Calon Siswa","gender":"P","parent_name":"Ortu","parent_phone":"0813"}`)))
	if rec.Code != http.StatusCreated || svc.createArg.Nis != "789" || svc.createArg.Status != db.StudentStatusEnumProspective || !svc.createArg.IsActive {
		t.Fatalf("PublicRegister() status/arg = %d/%+v", rec.Code, svc.createArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/students/"+studentID.String(), ""), "id", studentID.String())
	h.Delete(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteID != studentID {
		t.Fatalf("Delete() status/id = %d/%v", rec.Code, svc.deleteID)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPatch, "/api/students/"+studentID.String()+"/lifecycle", `{"status":"alumni"}`), "id", studentID.String())
	h.UpdateLifecycle(rec, req)
	if rec.Code != http.StatusOK || svc.lifecycleID != studentID || svc.lifecycleStatus != db.StudentStatusEnumAlumni {
		t.Fatalf("UpdateLifecycle() status/args = %d/%v/%s", rec.Code, svc.lifecycleID, svc.lifecycleStatus)
	}
}
