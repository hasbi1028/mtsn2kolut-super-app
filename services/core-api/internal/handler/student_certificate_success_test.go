package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeStudentCertificateHandlerService struct {
	activeOnly  bool
	templateErr error

	listStudentsSearch string
	listStudentsStatus string
	listStudentsErr    error

	listSearch       string
	listStatus       string
	listTemplateCode string
	listErr          error

	createInput service.CreateStudentCertificateInput
	createErr   error
	getID       pgtype.UUID
	getErr      error
	cancelID    pgtype.UUID
	cancelNote  string
	cancelErr   error
}

func (f *fakeStudentCertificateHandlerService) ListTemplates(ctx context.Context, activeOnly bool) ([]db.CertificateTemplate, error) {
	f.activeOnly = activeOnly
	if f.templateErr != nil {
		return nil, f.templateErr
	}
	return []db.CertificateTemplate{{ID: handlerTestUUID(160), Code: "SKA", IsActive: true}}, nil
}

func (f *fakeStudentCertificateHandlerService) ListStudents(ctx context.Context, search, status string) ([]db.ListStudentCertificateOptionsRow, error) {
	f.listStudentsSearch = search
	f.listStudentsStatus = status
	if f.listStudentsErr != nil {
		return nil, f.listStudentsErr
	}
	return []db.ListStudentCertificateOptionsRow{{ID: handlerTestUUID(161), Nama: "Siswa A"}}, nil
}

func (f *fakeStudentCertificateHandlerService) List(ctx context.Context, search, status, templateCode string) ([]db.ListStudentCertificatesRow, error) {
	f.listSearch = search
	f.listStatus = status
	f.listTemplateCode = templateCode
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListStudentCertificatesRow{{ID: handlerTestUUID(162), Status: "issued"}}, nil
}

func (f *fakeStudentCertificateHandlerService) Create(ctx context.Context, input service.CreateStudentCertificateInput) (db.GetStudentCertificateRow, error) {
	f.createInput = input
	if f.createErr != nil {
		return db.GetStudentCertificateRow{}, f.createErr
	}
	return db.GetStudentCertificateRow{ID: handlerTestUUID(163), Status: "issued"}, nil
}

func (f *fakeStudentCertificateHandlerService) Get(ctx context.Context, id pgtype.UUID) (db.GetStudentCertificateRow, error) {
	f.getID = id
	if f.getErr != nil {
		return db.GetStudentCertificateRow{}, f.getErr
	}
	return db.GetStudentCertificateRow{ID: id, Status: "issued"}, nil
}

func (f *fakeStudentCertificateHandlerService) Cancel(ctx context.Context, id pgtype.UUID, remarks string) (db.StudentCertificate, error) {
	f.cancelID = id
	f.cancelNote = remarks
	if f.cancelErr != nil {
		return db.StudentCertificate{}, f.cancelErr
	}
	return db.StudentCertificate{ID: id, Remarks: remarks, Status: "canceled"}, nil
}

func studentCertificateStaffRequest(method, target, body string, employeeID pgtype.UUID) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{
		"roles": []any{"staf"},
		"uid":   handlerTestUUID(1).String(),
		"eid":   employeeID.String(),
	})
}

func TestStudentCertificateSuccessHandlersForwardPayloads(t *testing.T) {
	templateID := handlerTestUUID(164)
	studentID := handlerTestUUID(165)
	certificateID := handlerTestUUID(166)
	employeeID := handlerTestUUID(167)
	svc := &fakeStudentCertificateHandlerService{}
	h := &StudentCertificate{svc: svc}

	rec := httptest.NewRecorder()
	h.ListTemplates(rec, studentCertificateStaffRequest(http.MethodGet, "/api/tu/surat-keterangan/templates", "", employeeID))
	if rec.Code != http.StatusOK || !svc.activeOnly {
		t.Fatalf("ListTemplates() status/activeOnly = %d/%v, want 200/true", rec.Code, svc.activeOnly)
	}

	rec = httptest.NewRecorder()
	h.ListStudents(rec, studentCertificateStaffRequest(http.MethodGet, "/api/tu/surat-keterangan/students?search=andi&status=active", "", employeeID))
	if rec.Code != http.StatusOK || svc.listStudentsSearch != "andi" || svc.listStudentsStatus != "active" {
		t.Fatalf("ListStudents() status/args = %d/%q/%q", rec.Code, svc.listStudentsSearch, svc.listStudentsStatus)
	}

	rec = httptest.NewRecorder()
	h.List(rec, studentCertificateStaffRequest(http.MethodGet, "/api/tu/surat-keterangan?search=nomor&status=issued&template_code=SKA", "", employeeID))
	if rec.Code != http.StatusOK || svc.listSearch != "nomor" || svc.listStatus != "issued" || svc.listTemplateCode != "SKA" {
		t.Fatalf("List() status/args = %d/%q/%q/%q", rec.Code, svc.listSearch, svc.listStatus, svc.listTemplateCode)
	}

	body := `{"template_id":"` + templateID.String() + `","student_id":"` + studentID.String() + `","tanggal_surat":"2026-05-01","purpose":"Beasiswa","recipient":"Komite","remarks":"Prioritas"}`
	rec = httptest.NewRecorder()
	h.Create(rec, studentCertificateStaffRequest(http.MethodPost, "/api/tu/surat-keterangan", body, employeeID))
	if rec.Code != http.StatusCreated || svc.createInput.TemplateID != templateID.String() || svc.createInput.StudentID != studentID.String() || svc.createInput.CreatedByUserID != handlerTestUUID(1) || svc.createInput.IssuedByEmployeeID != employeeID {
		t.Fatalf("Create() status/input = %d/%+v", rec.Code, svc.createInput)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(studentCertificateStaffRequest(http.MethodGet, "/api/tu/surat-keterangan/"+certificateID.String(), "", employeeID), "id", certificateID.String())
	h.Get(rec, req)
	if rec.Code != http.StatusOK || svc.getID != certificateID {
		t.Fatalf("Get() status/id = %d/%v", rec.Code, svc.getID)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(studentCertificateStaffRequest(http.MethodPost, "/api/tu/surat-keterangan/"+certificateID.String()+"/cancel", `{"remarks":"salah data"}`, employeeID), "id", certificateID.String())
	h.Cancel(rec, req)
	if rec.Code != http.StatusOK || svc.cancelID != certificateID || svc.cancelNote != "salah data" {
		t.Fatalf("Cancel() status/args = %d/%v/%q", rec.Code, svc.cancelID, svc.cancelNote)
	}
}

func TestStudentCertificateValidationAndServiceErrors(t *testing.T) {
	certificateID := handlerTestUUID(168)
	employeeID := handlerTestUUID(169)
	if NewStudentCertificate(nil) == nil {
		t.Fatal("NewStudentCertificate(nil) = nil")
	}

	t.Run("forbidden requests", func(t *testing.T) {
		h := &StudentCertificate{svc: &fakeStudentCertificateHandlerService{}}
		check := func(name string, call func(*httptest.ResponseRecorder)) {
			t.Helper()
			rec := httptest.NewRecorder()
			call(rec)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("%s status = %d, want 403", name, rec.Code)
			}
		}
		check("ListTemplates", func(rec *httptest.ResponseRecorder) {
			h.ListTemplates(rec, httptest.NewRequest(http.MethodGet, "/templates", nil))
		})
		check("ListStudents", func(rec *httptest.ResponseRecorder) {
			h.ListStudents(rec, httptest.NewRequest(http.MethodGet, "/students", nil))
		})
		check("List", func(rec *httptest.ResponseRecorder) {
			h.List(rec, httptest.NewRequest(http.MethodGet, "/certificates", nil))
		})
		check("Create", func(rec *httptest.ResponseRecorder) {
			h.Create(rec, httptest.NewRequest(http.MethodPost, "/certificates", strings.NewReader(`{}`)))
		})
		check("Get", func(rec *httptest.ResponseRecorder) {
			req := withRouteParam(httptest.NewRequest(http.MethodGet, "/certificates/"+certificateID.String(), nil), "id", certificateID.String())
			h.Get(rec, req)
		})
		check("Cancel", func(rec *httptest.ResponseRecorder) {
			req := withRouteParam(httptest.NewRequest(http.MethodPost, "/certificates/"+certificateID.String()+"/cancel", strings.NewReader(`{}`)), "id", certificateID.String())
			h.Cancel(rec, req)
		})
	})

	t.Run("list handlers map service errors", func(t *testing.T) {
		h := &StudentCertificate{svc: &fakeStudentCertificateHandlerService{templateErr: errors.New("templates failed")}}
		rec := httptest.NewRecorder()
		h.ListTemplates(rec, studentCertificateStaffRequest(http.MethodGet, "/templates", "", employeeID))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("ListTemplates(error) status = %d, want 500", rec.Code)
		}

		h.svc = &fakeStudentCertificateHandlerService{listStudentsErr: errors.New("status tidak valid")}
		rec = httptest.NewRecorder()
		h.ListStudents(rec, studentCertificateStaffRequest(http.MethodGet, "/students", "", employeeID))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("ListStudents(error) status = %d, want 400", rec.Code)
		}

		h.svc = &fakeStudentCertificateHandlerService{listErr: errors.New("filter tidak valid")}
		rec = httptest.NewRecorder()
		h.List(rec, studentCertificateStaffRequest(http.MethodGet, "/certificates", "", employeeID))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("List(error) status = %d, want 400", rec.Code)
		}
	})

	t.Run("create validates body and maps domain errors", func(t *testing.T) {
		h := &StudentCertificate{svc: &fakeStudentCertificateHandlerService{}}
		rec := httptest.NewRecorder()
		h.Create(rec, studentCertificateStaffRequest(http.MethodPost, "/certificates", `{bad`, employeeID))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Create(invalid json) status = %d, want 400", rec.Code)
		}

		for name, tc := range map[string]struct {
			err  error
			want int
		}{
			"duplicate":  {err: errors.New("duplicate key value violates uq_outgoing"), want: http.StatusConflict},
			"not found":  {err: pgx.ErrNoRows, want: http.StatusNotFound},
			"validation": {err: errors.New("tanggal surat tidak valid"), want: http.StatusBadRequest},
		} {
			h.svc = &fakeStudentCertificateHandlerService{createErr: tc.err}
			rec = httptest.NewRecorder()
			h.Create(rec, studentCertificateStaffRequest(http.MethodPost, "/certificates", `{}`, employeeID))
			if rec.Code != tc.want {
				t.Fatalf("Create(%s) status = %d, want %d", name, rec.Code, tc.want)
			}
		}
	})

	t.Run("get validates id and maps errors", func(t *testing.T) {
		h := &StudentCertificate{svc: &fakeStudentCertificateHandlerService{}}
		rec := httptest.NewRecorder()
		req := withRouteParam(studentCertificateStaffRequest(http.MethodGet, "/certificates/bad", "", employeeID), "id", "bad")
		h.Get(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Get(invalid id) status = %d, want 400", rec.Code)
		}

		h.svc = &fakeStudentCertificateHandlerService{getErr: pgx.ErrNoRows}
		rec = httptest.NewRecorder()
		req = withRouteParam(studentCertificateStaffRequest(http.MethodGet, "/certificates/"+certificateID.String(), "", employeeID), "id", certificateID.String())
		h.Get(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("Get(not found) status = %d, want 404", rec.Code)
		}

		h.svc = &fakeStudentCertificateHandlerService{getErr: errors.New("db failed")}
		rec = httptest.NewRecorder()
		req = withRouteParam(studentCertificateStaffRequest(http.MethodGet, "/certificates/"+certificateID.String(), "", employeeID), "id", certificateID.String())
		h.Get(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Get(internal) status = %d, want 500", rec.Code)
		}
	})

	t.Run("cancel validates body and maps errors", func(t *testing.T) {
		h := &StudentCertificate{svc: &fakeStudentCertificateHandlerService{}}
		rec := httptest.NewRecorder()
		req := withRouteParam(studentCertificateStaffRequest(http.MethodPost, "/certificates/bad/cancel", `{}`, employeeID), "id", "bad")
		h.Cancel(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Cancel(invalid id) status = %d, want 400", rec.Code)
		}

		rec = httptest.NewRecorder()
		req = withRouteParam(studentCertificateStaffRequest(http.MethodPost, "/certificates/"+certificateID.String()+"/cancel", `{bad`, employeeID), "id", certificateID.String())
		h.Cancel(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Cancel(invalid json) status = %d, want 400", rec.Code)
		}

		h.svc = &fakeStudentCertificateHandlerService{cancelErr: pgx.ErrNoRows}
		rec = httptest.NewRecorder()
		req = withRouteParam(studentCertificateStaffRequest(http.MethodPost, "/certificates/"+certificateID.String()+"/cancel", `{}`, employeeID), "id", certificateID.String())
		h.Cancel(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("Cancel(not found) status = %d, want 404", rec.Code)
		}

		h.svc = &fakeStudentCertificateHandlerService{cancelErr: errors.New("alasan wajib diisi")}
		rec = httptest.NewRecorder()
		req = withRouteParam(studentCertificateStaffRequest(http.MethodPost, "/certificates/"+certificateID.String()+"/cancel", `{}`, employeeID), "id", certificateID.String())
		h.Cancel(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Cancel(validation) status = %d, want 400", rec.Code)
		}
	})
}
