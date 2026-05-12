package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeEmployeeService struct {
	*service.Employee

	listRows              []db.ListEmployeesRow
	listErr               error
	listCalled            bool
	listWithStatusRows    []db.ListEmployeesWithStatusRow
	listWithStatusErr     error
	listWithStatusCalled  bool
	listPusakaRows        []db.ListPusakaEligibleEmployeesWithStatusRow
	listPusakaErr         error
	listPusakaCalled      bool
	getRow                db.GetEmployeeRow
	getErr                error
	getID                 pgtype.UUID
	createRow             db.GetEmployeeRow
	createErr             error
	createNip             string
	createNama            string
	createUnitKerja       string
	createEmploymentType  string
	createTanggalLahir    pgtype.Date
	createJenisKelamin    string
	createTempatLahir     string
	createPusakaUsername  string
	createPusakaPassword  string
	createIsActive        bool
	updateRow             db.GetEmployeeRow
	updateErr             error
	updateArg             db.UpdateEmployeeParams
	deleteID              pgtype.UUID
	deleteErr             error
	setActiveID           pgtype.UUID
	setActiveValue        bool
	setActiveErr          error
	upsertID              pgtype.UUID
	upsertUsername        string
	upsertPassword        string
	upsertEnabled         bool
	upsertErr             error
	setPusakaEnabledID    pgtype.UUID
	setPusakaEnabledValue bool
	setPusakaEnabledErr   error
	deletePusakaID        pgtype.UUID
	deletePusakaErr       error
	auditEntries          []employeeAuditEntry
	auditRows             []db.ListEntityAuditLogsRow
	auditRowsErr          error
	auditEmployeeID       string
	auditLimit            int32
	auditOffset           int32
}

type employeeAuditEntry struct {
	userID     pgtype.UUID
	action     string
	entityType string
	entityID   string
	metadata   []byte
}

func (f *fakeEmployeeService) List(context.Context) ([]db.ListEmployeesRow, error) {
	f.listCalled = true
	return f.listRows, f.listErr
}

func (f *fakeEmployeeService) ListWithStatus(context.Context) ([]db.ListEmployeesWithStatusRow, error) {
	f.listWithStatusCalled = true
	return f.listWithStatusRows, f.listWithStatusErr
}

func (f *fakeEmployeeService) ListPusakaEligibleWithStatus(context.Context) ([]db.ListPusakaEligibleEmployeesWithStatusRow, error) {
	f.listPusakaCalled = true
	return f.listPusakaRows, f.listPusakaErr
}

func (f *fakeEmployeeService) Get(_ context.Context, id pgtype.UUID) (db.GetEmployeeRow, error) {
	f.getID = id
	return f.getRow, f.getErr
}

func (f *fakeEmployeeService) Create(_ context.Context, nip, nama, unitKerja, employmentType string, tanggalLahir pgtype.Date, jenisKelamin, tempatLahir, pusakaUsername, pusakaPassword string, isActive bool) (db.GetEmployeeRow, error) {
	f.createNip = nip
	f.createNama = nama
	f.createUnitKerja = unitKerja
	f.createEmploymentType = employmentType
	f.createTanggalLahir = tanggalLahir
	f.createJenisKelamin = jenisKelamin
	f.createTempatLahir = tempatLahir
	f.createPusakaUsername = pusakaUsername
	f.createPusakaPassword = pusakaPassword
	f.createIsActive = isActive
	return f.createRow, f.createErr
}

func (f *fakeEmployeeService) Update(_ context.Context, p db.UpdateEmployeeParams) (db.GetEmployeeRow, error) {
	f.updateArg = p
	return f.updateRow, f.updateErr
}

func (f *fakeEmployeeService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeEmployeeService) SetActive(_ context.Context, id pgtype.UUID, isActive bool) error {
	f.setActiveID = id
	f.setActiveValue = isActive
	return f.setActiveErr
}

func (f *fakeEmployeeService) UpsertPusakaAccount(_ context.Context, employeeID pgtype.UUID, username, password string, isEnabled bool) error {
	f.upsertID = employeeID
	f.upsertUsername = username
	f.upsertPassword = password
	f.upsertEnabled = isEnabled
	return f.upsertErr
}

func (f *fakeEmployeeService) SetPusakaAccountEnabled(_ context.Context, employeeID pgtype.UUID, isEnabled bool) error {
	f.setPusakaEnabledID = employeeID
	f.setPusakaEnabledValue = isEnabled
	return f.setPusakaEnabledErr
}

func (f *fakeEmployeeService) DeletePusakaAccount(_ context.Context, employeeID pgtype.UUID) error {
	f.deletePusakaID = employeeID
	return f.deletePusakaErr
}

func (f *fakeEmployeeService) CreateAuditLog(_ context.Context, userID pgtype.UUID, action, entityType, entityID string, metadata []byte) error {
	f.auditEntries = append(f.auditEntries, employeeAuditEntry{
		userID:     userID,
		action:     action,
		entityType: entityType,
		entityID:   entityID,
		metadata:   metadata,
	})
	return nil
}

func (f *fakeEmployeeService) ListPusakaAuditLogs(_ context.Context, employeeID string, limit, offset int32) ([]db.ListEntityAuditLogsRow, error) {
	f.auditEmployeeID = employeeID
	f.auditLimit = limit
	f.auditOffset = offset
	return f.auditRows, f.auditRowsErr
}

func employeeTestRow(id pgtype.UUID, nama, employmentType string) db.GetEmployeeRow {
	return db.GetEmployeeRow{
		ID:              id,
		PegawaiUid:      "4040603180001",
		Nip:             "198001012006041001",
		Nama:            nama,
		UnitKerja:       "MTsN 2 Kolaka Utara",
		EmploymentType:  employmentType,
		JenisKelamin:    "L",
		TempatLahir:     "Kolaka",
		PusakaUsername:  "guru.pns",
		PusakaPassword:  "secret-password",
		PusakaIsEnabled: true,
		IsActive:        true,
	}
}

func decodeEmployeeData(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("response JSON decode failed: %v; body=%s", err, rec.Body.String())
	}
	if err := json.Unmarshal(envelope.Data, out); err != nil {
		t.Fatalf("data JSON decode failed: %v; data=%s", err, string(envelope.Data))
	}
}

func TestEmployeeListHandlersForwardAndSanitize(t *testing.T) {
	employeeID := handlerTestUUID(240)
	fake := &fakeEmployeeService{
		Employee: &service.Employee{},
		listRows: []db.ListEmployeesRow{
			{
				ID:              employeeID,
				Nip:             "198001012006041001",
				Nama:            "Guru PNS",
				UnitKerja:       "Madrasah",
				EmploymentType:  "pns",
				PusakaUsername:  "guru.pns",
				PusakaPassword:  "secret-password",
				PusakaIsEnabled: true,
				IsActive:        true,
			},
			{ID: handlerTestUUID(241), Nama: "Honorer", EmploymentType: "honorer"},
		},
		listWithStatusRows: []db.ListEmployeesWithStatusRow{
			{ID: employeeID, Nama: "Guru PNS", ActiveStatus: "running", PusakaEligible: true},
		},
		listPusakaRows: []db.ListPusakaEligibleEmployeesWithStatusRow{
			{ID: employeeID, Nama: "Guru PNS", EmploymentType: "pns", PusakaEligible: true},
		},
	}
	h := &Employee{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/employees", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("List status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.listCalled {
		t.Fatal("List did not call service List")
	}
	if strings.Contains(rec.Body.String(), "secret-password") || strings.Contains(rec.Body.String(), "pusaka_password") {
		t.Fatalf("List response leaked PUSAKA password: %s", rec.Body.String())
	}
	var rows []map[string]any
	decodeEmployeeData(t, rec, &rows)
	if len(rows) != 2 || rows[0]["pusaka_eligible"] != true || rows[1]["pusaka_eligible"] != false {
		t.Fatalf("List data = %#v, want sanitized eligibility flags", rows)
	}

	rec = httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/employees?with_status=1", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List with status status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.listWithStatusCalled {
		t.Fatal("List with_status did not call service ListWithStatus")
	}

	rec = httptest.NewRecorder()
	h.ListPusakaEligibleWithStatus(rec, adminRequest(http.MethodGet, "/api/employees/pusaka-eligible", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListPusakaEligibleWithStatus status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.listPusakaCalled {
		t.Fatal("ListPusakaEligibleWithStatus did not call service")
	}
}

func TestEmployeeCreateGetUpdateDeleteAndStatus(t *testing.T) {
	id := handlerTestUUID(242)
	fake := &fakeEmployeeService{
		Employee:  &service.Employee{},
		getRow:    employeeTestRow(id, "Guru Lama", "pns"),
		createRow: employeeTestRow(id, "Guru Baru", "pppk"),
		updateRow: employeeTestRow(id, "Guru Revisi", "pns"),
	}
	h := &Employee{svc: fake}

	rec := httptest.NewRecorder()
	h.Create(rec, adminRequest(http.MethodPost, "/api/employees", `{"nip":"1980","nama":"Guru Baru","unit_kerja":"TU","employment_type":"pppk","tanggal_lahir":"1980-01-02","jenis_kelamin":"P","tempat_lahir":"Kolaka","pusaka_username":"baru","pusaka_password":"rahasia","is_active":true}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createNip != "1980" || fake.createNama != "Guru Baru" || fake.createUnitKerja != "TU" || fake.createEmploymentType != "pppk" || fake.createJenisKelamin != "P" || fake.createTempatLahir != "Kolaka" || fake.createPusakaUsername != "baru" || fake.createPusakaPassword != "rahasia" || !fake.createIsActive {
		t.Fatalf("Create forwarded = (%q, %q, %q, %q, %q, %q, %q, %q, %v)", fake.createNip, fake.createNama, fake.createUnitKerja, fake.createEmploymentType, fake.createJenisKelamin, fake.createTempatLahir, fake.createPusakaUsername, fake.createPusakaPassword, fake.createIsActive)
	}

	rec = httptest.NewRecorder()
	h.Get(rec, withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+id.String(), ""), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Get status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.getID != id || strings.Contains(rec.Body.String(), "secret-password") || strings.Contains(rec.Body.String(), "pusaka_password") {
		t.Fatalf("Get forwarded id=%v body=%s, want sanitized response", fake.getID, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Update(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String(), `{"nip":"1990","nama":"Guru Revisi","unit_kerja":"Kurikulum","employment_type":"pns","tanggal_lahir":"1980-01-02","jenis_kelamin":"L","tempat_lahir":"Kolaka","is_active":false}`), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Update status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateArg.ID != id || fake.updateArg.Nip != "1990" || fake.updateArg.Nama != "Guru Revisi" || fake.updateArg.UnitKerja != "Kurikulum" || fake.updateArg.EmploymentType != "pns" || fake.updateArg.IsActive {
		t.Fatalf("Update params = %+v, want route id and decoded payload", fake.updateArg)
	}
	if fake.updateArg.JenisKelamin != "L" || fake.updateArg.TempatLahir != "Kolaka" {
		t.Fatalf("Update identity params = %+v, want gender/place", fake.updateArg)
	}

	rec = httptest.NewRecorder()
	h.Delete(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/employees/"+id.String(), ""), "id", id.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Delete status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID != id {
		t.Fatalf("Delete id = %v, want %v", fake.deleteID, id)
	}

	rec = httptest.NewRecorder()
	h.UpdateStatus(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String()+"/status", `{"is_active":false}`), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateStatus status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.setActiveID != id || fake.setActiveValue {
		t.Fatalf("SetActive = (%v, %v), want (%v, false)", fake.setActiveID, fake.setActiveValue, id)
	}
}

func TestEmployeePusakaCredentialsStatusDeleteAndAudit(t *testing.T) {
	id := handlerTestUUID(243)
	getRow := employeeTestRow(id, "Guru PNS", "pns")
	getRow.PusakaUsername = "old.user"
	getRow.PusakaIsEnabled = false
	fake := &fakeEmployeeService{
		Employee:  &service.Employee{},
		getRow:    getRow,
		updateRow: employeeTestRow(id, "Guru PNS", "pns"),
		auditRows: []db.ListEntityAuditLogsRow{
			{ID: handlerTestUUID(244), Action: "PUSAKA_ACCOUNT_UPDATE", EntityType: "pusaka_account", EntityID: id.String()},
		},
	}
	h := &Employee{svc: fake}

	rec := httptest.NewRecorder()
	h.GetPusakaStatus(rec, withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+id.String()+"/pusaka-status", ""), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetPusakaStatus status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"configured":true`) {
		t.Fatalf("GetPusakaStatus body = %s, want configured true", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdatePusakaCredentials(rec, withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+id.String()+"/pusaka-credentials", `{"pusaka_username":"new.user","pusaka_password":"new-secret"}`), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdatePusakaCredentials status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateArg.ID != id || fake.updateArg.Nip != getRow.Nip || fake.updateArg.Nama != getRow.Nama || fake.updateArg.EmploymentType != getRow.EmploymentType {
		t.Fatalf("UpdatePusakaCredentials update params = %+v, want current employee identity", fake.updateArg)
	}
	if fake.upsertID != id || fake.upsertUsername != "new.user" || fake.upsertPassword != "new-secret" || fake.upsertEnabled {
		t.Fatalf("UpsertPusakaAccount = (%v, %q, %q, %v), want existing disabled account preserved", fake.upsertID, fake.upsertUsername, fake.upsertPassword, fake.upsertEnabled)
	}
	if len(fake.auditEntries) != 1 || fake.auditEntries[0].action != "PUSAKA_ACCOUNT_UPDATE" || fake.auditEntries[0].entityID != id.String() {
		t.Fatalf("audit entries after credentials = %+v, want one update audit", fake.auditEntries)
	}
	meta := mustAuditMeta(t, fake.auditEntries[0].metadata)
	if meta["pusaka_username"] != "new.user" || meta["configured"] != true {
		t.Fatalf("credential audit metadata = %#v, want username and configured", meta)
	}

	rec = httptest.NewRecorder()
	h.UpdatePusakaAccountStatus(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String()+"/pusaka-account/status", `{"is_enabled":false}`), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdatePusakaAccountStatus status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.setPusakaEnabledID != id || fake.setPusakaEnabledValue {
		t.Fatalf("SetPusakaAccountEnabled = (%v, %v), want (%v, false)", fake.setPusakaEnabledID, fake.setPusakaEnabledValue, id)
	}

	rec = httptest.NewRecorder()
	h.DeletePusakaAccount(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/employees/"+id.String()+"/pusaka-account", ""), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("DeletePusakaAccount status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deletePusakaID != id {
		t.Fatalf("DeletePusakaAccount id = %v, want %v", fake.deletePusakaID, id)
	}
	if len(fake.auditEntries) != 3 {
		t.Fatalf("audit entries len = %d, want 3", len(fake.auditEntries))
	}

	rec = httptest.NewRecorder()
	h.ListPusakaAuditLogs(rec, withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+id.String()+"/pusaka-audit-logs?page=3&per_page=15", ""), "id", id.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListPusakaAuditLogs status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.auditEmployeeID != id.String() || fake.auditLimit != 15 || fake.auditOffset != 30 {
		t.Fatalf("ListPusakaAuditLogs params = (%q, %d, %d), want (%q, 15, 30)", fake.auditEmployeeID, fake.auditLimit, fake.auditOffset, id.String())
	}
}

func TestEmployeeHandlersMapServiceErrors(t *testing.T) {
	id := handlerTestUUID(245)
	tests := []struct {
		name       string
		handler    func(*Employee, http.ResponseWriter, *http.Request)
		svc        *fakeEmployeeService
		req        *http.Request
		wantStatus int
		wantBody   string
	}{
		{
			name:       "list internal",
			handler:    (*Employee).List,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, listErr: errors.New("db down")},
			req:        adminRequest(http.MethodGet, "/api/employees", ""),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "list with status internal",
			handler:    (*Employee).List,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, listWithStatusErr: errors.New("db down")},
			req:        adminRequest(http.MethodGet, "/api/employees?with_status=1", ""),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "list pusaka eligible internal",
			handler:    (*Employee).ListPusakaEligibleWithStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, listPusakaErr: errors.New("db down")},
			req:        adminRequest(http.MethodGet, "/api/employees/pusaka-eligible", ""),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "get invalid id",
			handler:    (*Employee).Get,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/employees/bad", ""), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "get not found",
			handler:    (*Employee).Get,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, getErr: pgx.ErrNoRows},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+id.String(), ""), "id", id.String()),
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "get internal",
			handler:    (*Employee).Get,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, getErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+id.String(), ""), "id", id.String()),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "create invalid json",
			handler:    (*Employee).Create,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        adminRequest(http.MethodPost, "/api/employees", `{`),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Data yang dikirim tidak valid",
		},
		{
			name:       "create validation message",
			handler:    (*Employee).Create,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, createErr: errors.New("jenis kepegawaian tidak valid")},
			req:        adminRequest(http.MethodPost, "/api/employees", `{"nip":"1","nama":"Guru","employment_type":"bad"}`),
			wantStatus: http.StatusBadRequest,
			wantBody:   "jenis kepegawaian tidak valid",
		},
		{
			name:       "update invalid id",
			handler:    (*Employee).Update,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/bad", `{}`), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "update invalid json",
			handler:    (*Employee).Update,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String(), `{`), "id", id.String()),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Data yang dikirim tidak valid",
		},
		{
			name:       "update not found",
			handler:    (*Employee).Update,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, updateErr: pgx.ErrNoRows},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String(), `{}`), "id", id.String()),
			wantStatus: http.StatusNotFound,
			wantBody:   "not found",
		},
		{
			name:       "update validation message",
			handler:    (*Employee).Update,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, updateErr: errors.New("disable or remove the pusaka account before changing employee type")},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String(), `{}`), "id", id.String()),
			wantStatus: http.StatusBadRequest,
			wantBody:   "nonaktifkan atau hapus akun PUSAKA sebelum mengubah jenis kepegawaian",
		},
		{
			name:       "delete invalid id",
			handler:    (*Employee).Delete,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/api/employees/bad", ""), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "delete internal",
			handler:    (*Employee).Delete,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, deleteErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/api/employees/"+id.String(), ""), "id", id.String()),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "status invalid id",
			handler:    (*Employee).UpdateStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/bad/status", `{}`), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "status invalid json",
			handler:    (*Employee).UpdateStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String()+"/status", `{`), "id", id.String()),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Data yang dikirim tidak valid",
		},
		{
			name:       "status internal",
			handler:    (*Employee).UpdateStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, setActiveErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String()+"/status", `{"is_active":true}`), "id", id.String()),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "pusaka status invalid id",
			handler:    (*Employee).GetPusakaStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/employees/bad/pusaka-status", ""), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "pusaka status internal",
			handler:    (*Employee).GetPusakaStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, getErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+id.String()+"/pusaka-status", ""), "id", id.String()),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "pusaka credentials invalid id",
			handler:    (*Employee).UpdatePusakaCredentials,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/bad/pusaka-credentials", `{}`), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "pusaka credentials invalid json",
			handler:    (*Employee).UpdatePusakaCredentials,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+id.String()+"/pusaka-credentials", `{`), "id", id.String()),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Data yang dikirim tidak valid",
		},
		{
			name:       "pusaka credentials get internal",
			handler:    (*Employee).UpdatePusakaCredentials,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, getErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+id.String()+"/pusaka-credentials", `{}`), "id", id.String()),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "pusaka credentials update internal",
			handler:    (*Employee).UpdatePusakaCredentials,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, getRow: employeeTestRow(id, "Guru PNS", "pns"), updateErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+id.String()+"/pusaka-credentials", `{}`), "id", id.String()),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "pusaka credentials upsert internal",
			handler:    (*Employee).UpdatePusakaCredentials,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, getRow: employeeTestRow(id, "Guru PNS", "pns"), updateRow: employeeTestRow(id, "Guru PNS", "pns"), upsertErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+id.String()+"/pusaka-credentials", `{"pusaka_username":"guru"}`), "id", id.String()),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
		{
			name:       "pusaka account status invalid id",
			handler:    (*Employee).UpdatePusakaAccountStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/bad/pusaka-account/status", `{}`), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "pusaka account status invalid json",
			handler:    (*Employee).UpdatePusakaAccountStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String()+"/pusaka-account/status", `{`), "id", id.String()),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Data yang dikirim tidak valid",
		},
		{
			name:       "pusaka account missing",
			handler:    (*Employee).UpdatePusakaAccountStatus,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, setPusakaEnabledErr: errors.New("pusaka account is not configured")},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/employees/"+id.String()+"/pusaka-account/status", `{"is_enabled":true}`), "id", id.String()),
			wantStatus: http.StatusBadRequest,
			wantBody:   "akun PUSAKA belum dikonfigurasi",
		},
		{
			name:       "delete pusaka account invalid id",
			handler:    (*Employee).DeletePusakaAccount,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/api/employees/bad/pusaka-account", ""), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "delete pusaka account service error",
			handler:    (*Employee).DeletePusakaAccount,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, deletePusakaErr: errors.New("pusaka account is not configured")},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/api/employees/"+id.String()+"/pusaka-account", ""), "id", id.String()),
			wantStatus: http.StatusBadRequest,
			wantBody:   "akun PUSAKA belum dikonfigurasi",
		},
		{
			name:       "pusaka audit invalid id",
			handler:    (*Employee).ListPusakaAuditLogs,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/employees/bad/pusaka-audit-logs", ""), "id", "bad"),
			wantStatus: http.StatusBadRequest,
			wantBody:   "ID data tidak valid",
		},
		{
			name:       "pusaka audit internal",
			handler:    (*Employee).ListPusakaAuditLogs,
			svc:        &fakeEmployeeService{Employee: &service.Employee{}, auditRowsErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+id.String()+"/pusaka-audit-logs", ""), "id", id.String()),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Employee{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want substring %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}
