package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

var assertAnError = errors.New("assertAnError")

// ─── Fake service ───

type fakeSubjectAssignmentService struct {
	matrix    *service.AssignmentMatrixData
	matrixErr error

	upsertErr error
	deleteErr error

	activeAY   string
	activeAYErr error
}

func (f *fakeSubjectAssignmentService) GetMatrixData(ctx context.Context, ayID string) (*service.AssignmentMatrixData, error) {
	return f.matrix, f.matrixErr
}

func (f *fakeSubjectAssignmentService) UpsertCell(ctx context.Context, req service.UpsertCellRequest) error {
	return f.upsertErr
}

func (f *fakeSubjectAssignmentService) DeleteCell(ctx context.Context, id string) error {
	return f.deleteErr
}

func (f *fakeSubjectAssignmentService) GetActiveAcademicYear(ctx context.Context) (string, error) {
	return f.activeAY, f.activeAYErr
}

func TestSubjectAssignmentHandler_GetMatrix(t *testing.T) {
	fake := &fakeSubjectAssignmentService{
		matrix: &service.AssignmentMatrixData{
			Classes:  []service.AssignmentMatrixClass{{ID: "c1", Code: "VII.A", Name: "VII.A", Level: "VII"}},
			Subjects: []service.AssignmentMatrixSubject{{ID: "s1", Code: "MTK", Name: "Matematika", Category: "M"}},
			Teachers: []service.AssignmentMatrixTeacher{{ID: "t1", NIP: "123", Nama: "Guru A"}},
			Cells:    []service.AssignmentMatrixCell{{ClassID: "c1", SubjectID: "s1", Status: "complete", TeacherName: "Guru A"}},
		},
	}
	h := &SubjectAssignmentHandler{svc: fake}

	req := httptest.NewRequest(http.MethodGet, "/api/academic/subject-assignments", nil)
	rec := httptest.NewRecorder()
	h.GetMatrix(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var apiResp struct {
		Data *service.AssignmentMatrixData `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&apiResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if apiResp.Data == nil || len(apiResp.Data.Classes) != 1 {
		t.Fatalf("expected 1 class, got %d classes", len(apiResp.Data.Classes))
	}
}

func TestSubjectAssignmentHandler_GetMatrix_Error(t *testing.T) {
	fake := &fakeSubjectAssignmentService{matrixErr: assertAnError}
	h := &SubjectAssignmentHandler{svc: fake}
	req := httptest.NewRequest(http.MethodGet, "/api/academic/subject-assignments", nil)
	rec := httptest.NewRecorder()
	h.GetMatrix(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestSubjectAssignmentHandler_UpsertCell(t *testing.T) {
	fake := &fakeSubjectAssignmentService{}
	h := &SubjectAssignmentHandler{svc: fake}
	body := `{"class_id":"c1","subject_id":"s1","teacher_employee_id":"t1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/academic/subject-assignments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.UpsertCell(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
}

func TestSubjectAssignmentHandler_UpsertCell_MissingFields(t *testing.T) {
	fake := &fakeSubjectAssignmentService{}
	h := &SubjectAssignmentHandler{svc: fake}
	tests := []struct{ name, body string }{
		{"empty class", `{"class_id":"","subject_id":"s1","teacher_employee_id":"t1"}`},
		{"empty subject", `{"class_id":"c1","subject_id":"","teacher_employee_id":"t1"}`},
		{"empty teacher", `{"class_id":"c1","subject_id":"s1","teacher_employee_id":""}`},
		{"invalid json", `invalid`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/academic/subject-assignments", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.UpsertCell(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestSubjectAssignmentHandler_DeleteCell(t *testing.T) {
	fake := &fakeSubjectAssignmentService{}
	h := &SubjectAssignmentHandler{svc: fake}
	req := httptest.NewRequest(http.MethodDelete, "/api/academic/subject-assignments/c1", nil)
	req.SetPathValue("id", "c1")
	rec := httptest.NewRecorder()
	h.DeleteCell(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

func TestSubjectAssignmentHandler_DeleteCell_MissingID(t *testing.T) {
	fake := &fakeSubjectAssignmentService{}
	h := &SubjectAssignmentHandler{svc: fake}
	req := httptest.NewRequest(http.MethodDelete, "/api/academic/subject-assignments/", nil)
	rec := httptest.NewRecorder()
	h.DeleteCell(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
