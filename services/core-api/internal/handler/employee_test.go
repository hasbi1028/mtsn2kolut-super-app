package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func employeeErrorBody(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v; body=%q", err, rr.Body.String())
	}
	return body.Error
}

func TestWriteEmployeeClientErrorMapsPostgresConstraints(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{name: "nip unique", err: &pgconn.PgError{Code: "23505", ConstraintName: "uq_employees_nip_not_empty"}, wantStatus: http.StatusConflict, wantError: "NIP sudah terdaftar"},
		{name: "legacy nip unique", err: &pgconn.PgError{Code: "23505", ConstraintName: "employees_nip_key"}, wantStatus: http.StatusConflict, wantError: "NIP sudah terdaftar"},
		{name: "pegawai uid unique", err: &pgconn.PgError{Code: "23505", ConstraintName: "uq_employees_pegawai_uid"}, wantStatus: http.StatusConflict, wantError: "ID pegawai bertabrakan, coba ulangi"},
		{name: "other unique", err: &pgconn.PgError{Code: "23505", ConstraintName: "other"}, wantStatus: http.StatusConflict, wantError: "data pegawai sudah terdaftar"},
		{name: "check constraint falls back safely", err: &pgconn.PgError{Code: "23514", Message: "Jenis kelamin tidak valid"}, wantStatus: http.StatusBadRequest, wantError: "fallback"},
		{name: "service message", err: errors.New("pusaka account is not configured"), wantStatus: http.StatusBadRequest, wantError: "akun PUSAKA belum dikonfigurasi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			writeEmployeeClientError(rr, tt.err, "fallback")
			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if got := employeeErrorBody(t, rr); got != tt.wantError {
				t.Fatalf("error = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestEmployeeCreateMapsValidationAndServiceErrors(t *testing.T) {
	h := &Employee{svc: &fakeEmployeeService{createErr: errors.New("only pns or pppk employees can have pusaka accounts")}}
	rr := httptest.NewRecorder()
	h.Create(rr, adminRequest(http.MethodPost, "/employees", `{"nip":"123","nama":"Ali","employment_type":"honorer","tanggal_lahir":"2026-99-01"}`))
	if rr.Code != http.StatusBadRequest || employeeErrorBody(t, rr) != "tanggal_lahir tidak valid" {
		t.Fatalf("invalid date response = %d %q, want tanggal validation", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.Create(rr, adminRequest(http.MethodPost, "/employees", `{"nip":"123","nama":"Ali","employment_type":"honorer","tanggal_lahir":"2026-05-01","pusaka_username":"ali"}`))
	if rr.Code != http.StatusBadRequest || employeeErrorBody(t, rr) != "akun PUSAKA hanya untuk pegawai PNS atau PPPK" {
		t.Fatalf("service error response = %d %q, want mapped pusaka eligibility error", rr.Code, rr.Body.String())
	}
}

func TestEmployeeUpdateAndGetSimpleErrorMapping(t *testing.T) {
	validID := "00000000-0000-0000-0000-000000000009"

	h := &Employee{svc: &fakeEmployeeService{getErr: pgx.ErrNoRows}}
	rr := httptest.NewRecorder()
	h.Get(rr, withRouteParam(adminRequest(http.MethodGet, "/employees/"+validID, ""), "id", validID))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Get(no rows) status = %d, want 404", rr.Code)
	}

	h = &Employee{svc: &fakeEmployeeService{updateErr: pgx.ErrNoRows}}
	rr = httptest.NewRecorder()
	h.Update(rr, withRouteParam(adminRequest(http.MethodPut, "/employees/"+validID, `{"nip":"123","nama":"Ali","employment_type":"pns","tanggal_lahir":"2026-05-01","is_active":true}`), "id", validID))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Update(no rows) status = %d, want 404", rr.Code)
	}

	h = &Employee{svc: &fakeEmployeeService{setPusakaEnabledErr: errors.New("pusaka account is not configured")}}
	rr = httptest.NewRecorder()
	h.UpdatePusakaAccountStatus(rr, withRouteParam(adminRequest(http.MethodPatch, "/employees/"+validID+"/pusaka/status", `{"is_enabled":true}`), "id", validID))
	if rr.Code != http.StatusBadRequest || employeeErrorBody(t, rr) != "akun PUSAKA belum dikonfigurasi" {
		t.Fatalf("UpdatePusakaAccountStatus error = %d %q, want mapped bad request", rr.Code, rr.Body.String())
	}
}
