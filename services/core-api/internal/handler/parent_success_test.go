package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeParentService struct {
	getID pgtype.UUID

	createNama    string
	createPhone   string
	createAddress string

	updateID      pgtype.UUID
	updateNama    string
	updatePhone   string
	updateAddress string

	deleteID pgtype.UUID

	linkParentID  pgtype.UUID
	linkStudentID pgtype.UUID

	unlinkParentID  pgtype.UUID
	unlinkStudentID pgtype.UUID

	childrenParentID pgtype.UUID
}

func (f *fakeParentService) List(ctx context.Context) ([]db.Parent, error) {
	return []db.Parent{{ID: handlerTestUUID(210), Nama: "Orang Tua"}}, nil
}

func (f *fakeParentService) Get(ctx context.Context, id pgtype.UUID) (db.Parent, error) {
	f.getID = id
	return db.Parent{ID: id, Nama: "Orang Tua"}, nil
}

func (f *fakeParentService) Create(ctx context.Context, nama, phone, address string) (db.Parent, error) {
	f.createNama = nama
	f.createPhone = phone
	f.createAddress = address
	return db.Parent{ID: handlerTestUUID(211), Nama: nama, Phone: phone, Address: address}, nil
}

func (f *fakeParentService) Update(ctx context.Context, id pgtype.UUID, nama, phone, address string) (db.Parent, error) {
	f.updateID = id
	f.updateNama = nama
	f.updatePhone = phone
	f.updateAddress = address
	return db.Parent{ID: id, Nama: nama, Phone: phone, Address: address}, nil
}

func (f *fakeParentService) Delete(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return nil
}

func (f *fakeParentService) LinkStudent(ctx context.Context, parentID, studentID pgtype.UUID) error {
	f.linkParentID = parentID
	f.linkStudentID = studentID
	return nil
}

func (f *fakeParentService) UnlinkStudent(ctx context.Context, parentID, studentID pgtype.UUID) error {
	f.unlinkParentID = parentID
	f.unlinkStudentID = studentID
	return nil
}

func (f *fakeParentService) ListChildren(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	f.childrenParentID = parentID
	return []db.ListParentChildrenRow{}, nil
}

func TestParentSuccessHandlersForwardPayloads(t *testing.T) {
	parentID := handlerTestUUID(212)
	studentID := handlerTestUUID(213)
	svc := &fakeParentService{}
	h := &Parent{svc: svc}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/parents", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodGet, "/api/parents/"+parentID.String(), ""), "id", parentID.String())
	h.Get(rec, req)
	if rec.Code != http.StatusOK || svc.getID != parentID {
		t.Fatalf("Get() status/id = %d/%v", rec.Code, svc.getID)
	}

	body := `{"nama":"Orang Tua","phone":"0812","address":"Lasusua"}`
	rec = httptest.NewRecorder()
	h.Create(rec, adminRequest(http.MethodPost, "/api/parents", body))
	if rec.Code != http.StatusCreated || svc.createNama != "Orang Tua" || svc.createPhone != "0812" || svc.createAddress != "Lasusua" {
		t.Fatalf("Create() status/args = %d/%q/%q/%q", rec.Code, svc.createNama, svc.createPhone, svc.createAddress)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/parents/"+parentID.String(), body), "id", parentID.String())
	h.Update(rec, req)
	if rec.Code != http.StatusOK || svc.updateID != parentID || svc.updateNama != "Orang Tua" {
		t.Fatalf("Update() status/args = %d/%v/%q", rec.Code, svc.updateID, svc.updateNama)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/parents/"+parentID.String(), ""), "id", parentID.String())
	h.Delete(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteID != parentID {
		t.Fatalf("Delete() status/id = %d/%v", rec.Code, svc.deleteID)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPost, "/api/parents/"+parentID.String()+"/link", `{"student_id":"`+studentID.String()+`"}`), "id", parentID.String())
	h.LinkStudent(rec, req)
	if rec.Code != http.StatusOK || svc.linkParentID != parentID || svc.linkStudentID != studentID {
		t.Fatalf("LinkStudent() status/args = %d/%v/%v", rec.Code, svc.linkParentID, svc.linkStudentID)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPost, "/api/parents/"+parentID.String()+"/unlink", `{"student_id":"`+studentID.String()+`"}`), "id", parentID.String())
	h.UnlinkStudent(rec, req)
	if rec.Code != http.StatusOK || svc.unlinkParentID != parentID || svc.unlinkStudentID != studentID {
		t.Fatalf("UnlinkStudent() status/args = %d/%v/%v", rec.Code, svc.unlinkParentID, svc.unlinkStudentID)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodGet, "/api/parents/"+parentID.String()+"/children", ""), "id", parentID.String())
	h.ListChildren(rec, req)
	if rec.Code != http.StatusOK || svc.childrenParentID != parentID {
		t.Fatalf("ListChildren() status/id = %d/%v", rec.Code, svc.childrenParentID)
	}
}
