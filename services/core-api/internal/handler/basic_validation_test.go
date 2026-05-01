package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func adminRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "uid": "01000000-0000-0000-0000-000000000000"})
}

func withRouteParams(req *http.Request, pairs ...string) *http.Request {
	chiCtx := chi.NewRouteContext()
	for i := 0; i+1 < len(pairs); i += 2 {
		chiCtx.URLParams.Add(pairs[i], pairs[i+1])
	}
	return req.WithContext(contextWithRoute(req.Context(), chiCtx))
}

func contextWithRoute(ctx context.Context, chiCtx *chi.Context) context.Context {
	return context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
}

func TestAcademicRejectsInvalidAdminRequests(t *testing.T) {
	h := &Academic{svc: &service.Academic{}}
	tests := []struct {
		name       string
		fn         http.HandlerFunc
		entity     string
		id         string
		body       string
		wantStatus int
	}{
		{name: "create unknown entity", fn: h.Create, entity: "unknown", body: `{}`, wantStatus: http.StatusNotFound},
		{name: "create years invalid json", fn: h.Create, entity: "years", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create years bad start", fn: h.Create, entity: "years", body: `{"start_date":"bad","end_date":"2026-12-31"}`, wantStatus: http.StatusBadRequest},
		{name: "create classes bad year", fn: h.Create, entity: "classes", body: `{"academic_year_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "create assignments bad class", fn: h.Create, entity: "assignments", body: `{"class_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "create timetables bad day", fn: h.Create, entity: "timetables", body: `{"assignment_id":"01000000-0000-0000-0000-000000000000","day_of_week":7}`, wantStatus: http.StatusBadRequest},
		{name: "update bad id", fn: h.Update, entity: "timetables", id: "bad", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "update unknown entity", fn: h.Update, entity: "unknown", id: "01000000-0000-0000-0000-000000000000", body: `{}`, wantStatus: http.StatusNotFound},
		{name: "delete bad id", fn: h.Delete, entity: "years", id: "bad", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "delete unknown entity", fn: h.Delete, entity: "unknown", id: "01000000-0000-0000-0000-000000000000", body: `{}`, wantStatus: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(http.MethodPost, "/academic/"+tt.entity, tt.body)
			if tt.id != "" {
				req = withRouteParams(req, "entity", tt.entity, "id", tt.id)
			} else {
				req = withRouteParams(req, "entity", tt.entity)
			}
			tt.fn(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestStudentRejectsInvalidAdminAndPublicRequests(t *testing.T) {
	h := &Student{svc: &service.Student{}}
	tests := []struct {
		name       string
		fn         http.HandlerFunc
		body       string
		id         string
		wantStatus int
	}{
		{name: "create invalid json", fn: h.Create, body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create invalid class", fn: h.Create, body: `{"class_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update invalid id", fn: h.Update, body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update invalid json", fn: h.Update, body: `{`, id: "01000000-0000-0000-0000-000000000000", wantStatus: http.StatusBadRequest},
		{name: "delete invalid id", fn: h.Delete, body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "lifecycle invalid id", fn: h.UpdateLifecycle, body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "lifecycle invalid json", fn: h.UpdateLifecycle, body: `{`, id: "01000000-0000-0000-0000-000000000000", wantStatus: http.StatusBadRequest},
		{name: "lifecycle invalid status", fn: h.UpdateLifecycle, body: `{"status":"blocked"}`, id: "01000000-0000-0000-0000-000000000000", wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(http.MethodPost, "/students", tt.body)
			if tt.id != "" {
				req = withRouteParam(req, "id", tt.id)
			}
			tt.fn(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}

	rec := httptest.NewRecorder()
	h.PublicRegister(rec, httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PublicRegister(invalid json) status = %d, want 400", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.PublicRegister(rec, httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"nis":"123"}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PublicRegister(missing nama) status = %d, want 400", rec.Code)
	}
}

func TestEmployeeRejectsInvalidAdminRequestsAndSanitizesResponses(t *testing.T) {
	h := &Employee{svc: &service.Employee{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
		body string
		id   string
	}{
		{name: "get invalid id", fn: h.Get, id: "bad"},
		{name: "create invalid json", fn: h.Create, body: `{`},
		{name: "update invalid id", fn: h.Update, id: "bad", body: `{}`},
		{name: "update invalid json", fn: h.Update, id: "01000000-0000-0000-0000-000000000000", body: `{`},
		{name: "delete invalid id", fn: h.Delete, id: "bad"},
		{name: "status invalid id", fn: h.UpdateStatus, id: "bad", body: `{}`},
		{name: "status invalid json", fn: h.UpdateStatus, id: "01000000-0000-0000-0000-000000000000", body: `{`},
		{name: "pusaka status invalid id", fn: h.GetPusakaStatus, id: "bad"},
		{name: "pusaka credentials invalid id", fn: h.UpdatePusakaCredentials, id: "bad", body: `{}`},
		{name: "pusaka credentials invalid json", fn: h.UpdatePusakaCredentials, id: "01000000-0000-0000-0000-000000000000", body: `{`},
		{name: "pusaka account status invalid id", fn: h.UpdatePusakaAccountStatus, id: "bad", body: `{}`},
		{name: "pusaka account status invalid json", fn: h.UpdatePusakaAccountStatus, id: "01000000-0000-0000-0000-000000000000", body: `{`},
		{name: "delete pusaka invalid id", fn: h.DeletePusakaAccount, id: "bad"},
		{name: "audit logs invalid id", fn: h.ListPusakaAuditLogs, id: "bad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(http.MethodPost, "/employees", tt.body)
			if tt.id != "" {
				req = withRouteParam(req, "id", tt.id)
			}
			tt.fn(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s status = %d, want 400; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}

	emp := sanitizeEmployee(db.GetEmployeeRow{
		ID:              pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Nip:             "1980",
		Nama:            "Guru PNS",
		UnitKerja:       "Madrasah",
		EmploymentType:  "pns",
		PusakaUsername:  "guru.pns",
		PusakaPassword:  "secret",
		PusakaIsEnabled: true,
		IsActive:        true,
	})
	if emp.PusakaEligible != true || emp.HasPusakaAccount != true || emp.PusakaUsername != "guru.pns" {
		t.Fatalf("sanitizeEmployee() = %+v, want eligible configured PUSAKA account without password", emp)
	}
	items := sanitizeEmployees([]db.ListEmployeesRow{
		{ID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true}, Nama: "Honorer", EmploymentType: "honorer"},
		{ID: pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, Nama: "PPPK", EmploymentType: "pppk", PusakaUsername: "pppk"},
	})
	if len(items) != 2 || items[0].PusakaEligible || !items[1].PusakaEligible || !items[1].HasPusakaAccount {
		t.Fatalf("sanitizeEmployees() = %+v, want eligibility by employment type", items)
	}
}

func TestParentRejectsInvalidAdminRequests(t *testing.T) {
	h := &Parent{svc: &service.Parent{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
		body string
		id   string
	}{
		{name: "get invalid id", fn: h.Get, id: "bad"},
		{name: "create invalid json", fn: h.Create, body: `{`},
		{name: "update invalid id", fn: h.Update, id: "bad", body: `{}`},
		{name: "update invalid json", fn: h.Update, id: "01000000-0000-0000-0000-000000000000", body: `{`},
		{name: "delete invalid id", fn: h.Delete, id: "bad"},
		{name: "link invalid json", fn: h.LinkStudent, id: "01000000-0000-0000-0000-000000000000", body: `{`},
		{name: "link invalid student", fn: h.LinkStudent, id: "01000000-0000-0000-0000-000000000000", body: `{"student_id":"bad"}`},
		{name: "unlink invalid id", fn: h.UnlinkStudent, id: "bad", body: `{}`},
		{name: "unlink invalid json", fn: h.UnlinkStudent, id: "01000000-0000-0000-0000-000000000000", body: `{`},
		{name: "children invalid id", fn: h.ListChildren, id: "bad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(http.MethodPost, "/parents", tt.body)
			if tt.id != "" {
				req = withRouteParam(req, "id", tt.id)
			}
			tt.fn(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s status = %d, want 400; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestWebsiteAndUserRejectInvalidRequests(t *testing.T) {
	website := &Website{svc: &service.Website{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
		body string
		id   string
	}{
		{name: "website create invalid json", fn: website.Create, body: `{`},
		{name: "website update invalid id", fn: website.Update, body: `{}`, id: "bad"},
		{name: "website update invalid json", fn: website.Update, body: `{`, id: "01000000-0000-0000-0000-000000000000"},
		{name: "website delete invalid id", fn: website.Delete, body: `{}`, id: "bad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(http.MethodPost, "/website", tt.body)
			if tt.id != "" {
				req = withRouteParam(req, "id", tt.id)
			}
			tt.fn(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s status = %d, want 400; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}

	user := NewUser(nil)
	rec := httptest.NewRecorder()
	user.Create(rec, adminRequest(http.MethodPost, "/users", `{`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("User.Create(invalid json) status = %d, want 400", rec.Code)
	}
	rec = httptest.NewRecorder()
	user.Create(rec, adminRequest(http.MethodPost, "/users", `{"username":"admin"}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("User.Create(missing password) status = %d, want 400", rec.Code)
	}
}

func TestValidateUserCreateRoleProfileMatrix(t *testing.T) {
	validID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	tests := []struct {
		name    string
		roles   []string
		empID   pgtype.UUID
		stuID   pgtype.UUID
		parID   pgtype.UUID
		wantErr string
	}{
		{name: "admin no profile", roles: []string{"admin"}},
		{name: "guru with employee", roles: []string{"guru"}, empID: validID},
		{name: "invalid role", roles: []string{"operator"}, wantErr: "role tidak valid"},
		{name: "guru missing employee", roles: []string{"guru"}, wantErr: "role guru/staf/kesiswaan wajib ditautkan ke pegawai"},
		{name: "siswa missing student", roles: []string{"siswa"}, wantErr: "role siswa wajib ditautkan ke siswa"},
		{name: "ortu missing parent", roles: []string{"ortu"}, wantErr: "role ortu wajib ditautkan ke orang tua"},
		{name: "employee without employee role", roles: []string{"admin"}, empID: validID, wantErr: "tautan pegawai hanya boleh untuk role guru/staf/kesiswaan"},
		{name: "student without student role", roles: []string{"admin"}, stuID: validID, wantErr: "tautan siswa hanya boleh untuk role siswa"},
		{name: "parent without parent role", roles: []string{"admin"}, parID: validID, wantErr: "tautan orang tua hanya boleh untuk role ortu"},
		{name: "multiple profiles", roles: []string{"guru", "siswa"}, empID: validID, stuID: validID, wantErr: "satu akun hanya boleh ditautkan ke satu jenis profil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserCreate(tt.roles, tt.empID, tt.stuID, tt.parID)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateUserCreate() error = %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("validateUserCreate() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
