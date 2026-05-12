package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Employee struct {
	svc employeeService
}

type employeeService interface {
	List(ctx context.Context) ([]db.ListEmployeesRow, error)
	ListWithStatus(ctx context.Context) ([]db.ListEmployeesWithStatusRow, error)
	ListPusakaEligibleWithStatus(ctx context.Context) ([]db.ListPusakaEligibleEmployeesWithStatusRow, error)
	Get(ctx context.Context, id pgtype.UUID) (db.GetEmployeeRow, error)
	Create(ctx context.Context, nip, nama, unitKerja, employmentType string, tanggalLahir pgtype.Date, jenisKelamin, tempatLahir, pusakaUsername, pusakaPassword string, isActive bool) (db.GetEmployeeRow, error)
	Update(ctx context.Context, p db.UpdateEmployeeParams) (db.GetEmployeeRow, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	SetActive(ctx context.Context, id pgtype.UUID, isActive bool) error
	UpsertPusakaAccount(ctx context.Context, employeeID pgtype.UUID, username, password string, isEnabled bool) error
	SetPusakaAccountEnabled(ctx context.Context, employeeID pgtype.UUID, isEnabled bool) error
	DeletePusakaAccount(ctx context.Context, employeeID pgtype.UUID) error
	CreateAuditLog(ctx context.Context, userID pgtype.UUID, action, entityType, entityID string, metadata []byte) error
	ListPusakaAuditLogs(ctx context.Context, employeeID string, limit, offset int32) ([]db.ListEntityAuditLogsRow, error)
}

type employeeResponse struct {
	ID               pgtype.UUID        `json:"id"`
	PegawaiUID       string             `json:"pegawai_uid"`
	Nip              string             `json:"nip"`
	Nama             string             `json:"nama"`
	UnitKerja        string             `json:"unit_kerja"`
	EmploymentType   string             `json:"employment_type"`
	JenisKelamin     string             `json:"jenis_kelamin"`
	TempatLahir      string             `json:"tempat_lahir"`
	PusakaUsername   string             `json:"pusaka_username"`
	PusakaIsEnabled  bool               `json:"pusaka_is_enabled"`
	PusakaEligible   bool               `json:"pusaka_eligible"`
	HasPusakaAccount bool               `json:"has_pusaka_account"`
	IsActive         bool               `json:"is_active"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
	TanggalLahir     string             `json:"tanggal_lahir"`
}

func NewEmployee(svc *service.Employee) *Employee { return &Employee{svc: svc} }

func employeeClientMessage(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	switch err.Error() {
	case "only pns or pppk employees can have pusaka accounts":
		return "akun PUSAKA hanya untuk pegawai PNS atau PPPK"
	case "Jenis kepegawaian tidak valid":
		return "jenis kepegawaian tidak valid"
	case "Jenis kelamin tidak valid":
		return "jenis kelamin tidak valid"
	case "disable or remove the pusaka account before changing employee type":
		return "nonaktifkan atau hapus akun PUSAKA sebelum mengubah jenis kepegawaian"
	case "pusaka account is not configured":
		return "akun PUSAKA belum dikonfigurasi"
	default:
		return safeClientMessage(err, fallback)
	}
}

func writeEmployeeClientError(w http.ResponseWriter, err error, fallback string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			switch pgErr.ConstraintName {
			case "uq_employees_nip_not_empty", "employees_nip_key":
				api.Conflict(w, "NIP sudah terdaftar")
			case "uq_employees_pegawai_uid":
				api.Conflict(w, "ID pegawai bertabrakan, coba ulangi")
			default:
				api.Conflict(w, "data pegawai sudah terdaftar")
			}
			return
		case "23514":
			api.BadRequest(w, employeeClientMessage(err, fallback))
			return
		}
	}
	api.BadRequest(w, employeeClientMessage(err, fallback))
}

func auditUserID(r *http.Request) pgtype.UUID {
	var uid pgtype.UUID
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if raw, ok := claims["uid"].(string); ok {
			_ = uid.Scan(raw)
		}
	}
	return uid
}

func (h *Employee) List(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	if r.URL.Query().Get("with_status") == "1" {
		h.ListWithStatus(w, r)
		return
	}
	employees, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, sanitizeEmployees(employees))
}

func (h *Employee) ListWithStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.ListWithStatus(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Employee) ListPusakaEligibleWithStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.ListPusakaEligibleWithStatus(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Employee) Get(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	emp, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, sanitizeEmployee(emp))
}

func (h *Employee) Create(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		Nip            string `json:"nip"`
		Nama           string `json:"nama"`
		UnitKerja      string `json:"unit_kerja"`
		EmploymentType string `json:"employment_type"`
		TanggalLahir   string `json:"tanggal_lahir"`
		JenisKelamin   string `json:"jenis_kelamin"`
		TempatLahir    string `json:"tempat_lahir"`
		PusakaUsername string `json:"pusaka_username"`
		PusakaPassword string `json:"pusaka_password"`
		IsActive       bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	tanggalLahir, err := parseEmployeeDate(body.TanggalLahir)
	if err != nil {
		api.BadRequest(w, "tanggal_lahir tidak valid")
		return
	}
	emp, err := h.svc.Create(r.Context(), body.Nip, body.Nama, body.UnitKerja, body.EmploymentType, tanggalLahir, body.JenisKelamin, body.TempatLahir, body.PusakaUsername, body.PusakaPassword, body.IsActive)
	if err != nil {
		writeEmployeeClientError(w, err, "Data pegawai tidak valid")
		return
	}
	api.Created(w, sanitizeEmployee(emp))
}

func (h *Employee) Update(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		Nip            string `json:"nip"`
		Nama           string `json:"nama"`
		UnitKerja      string `json:"unit_kerja"`
		EmploymentType string `json:"employment_type"`
		TanggalLahir   string `json:"tanggal_lahir"`
		JenisKelamin   string `json:"jenis_kelamin"`
		TempatLahir    string `json:"tempat_lahir"`
		IsActive       bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	tanggalLahir, err := parseEmployeeDate(body.TanggalLahir)
	if err != nil {
		api.BadRequest(w, "tanggal_lahir tidak valid")
		return
	}
	p := db.UpdateEmployeeParams{
		ID:             id,
		Nip:            body.Nip,
		Nama:           body.Nama,
		UnitKerja:      body.UnitKerja,
		EmploymentType: body.EmploymentType,
		TanggalLahir:   tanggalLahir,
		JenisKelamin:   body.JenisKelamin,
		TempatLahir:    body.TempatLahir,
		IsActive:       body.IsActive,
	}
	emp, err := h.svc.Update(r.Context(), p)
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		writeEmployeeClientError(w, err, "Perubahan data pegawai tidak valid")
		return
	}
	api.OK(w, sanitizeEmployee(emp))
}

func (h *Employee) Delete(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *Employee) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	if err := h.svc.SetActive(r.Context(), id, body.IsActive); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"id":        id,
		"is_active": body.IsActive,
	})
}

func (h *Employee) GetPusakaStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	emp, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	configured := emp.PusakaUsername != "" && emp.PusakaPassword != ""
	api.OK(w, map[string]any{
		"employee_id":     id,
		"pusaka_username": emp.PusakaUsername,
		"configured":      configured,
	})
}

func (h *Employee) UpdatePusakaCredentials(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		PusakaUsername string `json:"pusaka_username"`
		PusakaPassword string `json:"pusaka_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	emp, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	updated, err := h.svc.Update(r.Context(), db.UpdateEmployeeParams{
		ID:             emp.ID,
		Nip:            emp.Nip,
		Nama:           emp.Nama,
		UnitKerja:      emp.UnitKerja,
		EmploymentType: emp.EmploymentType,
		TanggalLahir:   emp.TanggalLahir,
		JenisKelamin:   emp.JenisKelamin,
		TempatLahir:    emp.TempatLahir,
		IsActive:       emp.IsActive,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	isEnabled := emp.IsActive
	if emp.PusakaUsername != "" {
		isEnabled = emp.PusakaIsEnabled
	}
	if err := h.svc.UpsertPusakaAccount(r.Context(), id, body.PusakaUsername, body.PusakaPassword, isEnabled); err != nil {
		api.Internal(w, err)
		return
	}
	meta, _ := json.Marshal(map[string]any{
		"employee_id":     pgUUIDString(id),
		"employment_type": emp.EmploymentType,
		"pusaka_username": body.PusakaUsername,
		"configured":      body.PusakaUsername != "",
	})
	_ = h.svc.CreateAuditLog(r.Context(), auditUserID(r), "PUSAKA_ACCOUNT_UPDATE", "pusaka_account", pgUUIDString(id), meta)
	api.OK(w, map[string]any{
		"status":     "success",
		"configured": body.PusakaUsername != "" || updated.PusakaUsername != "",
	})
}

func (h *Employee) UpdatePusakaAccountStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		IsEnabled bool `json:"is_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	if err := h.svc.SetPusakaAccountEnabled(r.Context(), id, body.IsEnabled); err != nil {
		writeEmployeeClientError(w, err, "Perubahan status akun PUSAKA tidak valid")
		return
	}
	meta, _ := json.Marshal(map[string]any{
		"employee_id": pgUUIDString(id),
		"is_enabled":  body.IsEnabled,
	})
	_ = h.svc.CreateAuditLog(r.Context(), auditUserID(r), "PUSAKA_ACCOUNT_TOGGLE", "pusaka_account", pgUUIDString(id), meta)
	api.OK(w, map[string]any{
		"employee_id": id,
		"is_enabled":  body.IsEnabled,
	})
}

func (h *Employee) DeletePusakaAccount(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if err := h.svc.DeletePusakaAccount(r.Context(), id); err != nil {
		writeEmployeeClientError(w, err, "Penghapusan akun PUSAKA tidak valid")
		return
	}
	meta, _ := json.Marshal(map[string]any{
		"employee_id": pgUUIDString(id),
		"deleted":     true,
	})
	_ = h.svc.CreateAuditLog(r.Context(), auditUserID(r), "PUSAKA_ACCOUNT_DELETE", "pusaka_account", pgUUIDString(id), meta)
	api.OK(w, map[string]any{
		"employee_id": id,
		"deleted":     true,
	})
}

func (h *Employee) ListPusakaAuditLogs(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	q := r.URL.Query()
	limit := int32(pageSize(q.Get("per_page"), 20))
	page := pageNum(q.Get("page"), 1)
	offset := int32((page - 1) * int(limit))
	rows, err := h.svc.ListPusakaAuditLogs(r.Context(), pgUUIDString(id), limit, offset)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func employeeDateString(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func parseEmployeeDate(value string) (pgtype.Date, error) {
	if value == "" {
		return pgtype.Date{}, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: parsed, Valid: true}, nil
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	return u, u.Scan(s)
}

func pgUUIDString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return u.String()
}

func sanitizeEmployee(emp db.GetEmployeeRow) employeeResponse {
	return employeeResponse{
		ID:               emp.ID,
		PegawaiUID:       emp.PegawaiUid,
		Nip:              emp.Nip,
		Nama:             emp.Nama,
		UnitKerja:        emp.UnitKerja,
		EmploymentType:   emp.EmploymentType,
		JenisKelamin:     emp.JenisKelamin,
		TempatLahir:      emp.TempatLahir,
		PusakaUsername:   emp.PusakaUsername,
		PusakaIsEnabled:  emp.PusakaIsEnabled,
		PusakaEligible:   emp.EmploymentType == "pns" || emp.EmploymentType == "pppk",
		HasPusakaAccount: emp.PusakaUsername != "",
		IsActive:         emp.IsActive,
		CreatedAt:        emp.CreatedAt,
		UpdatedAt:        emp.UpdatedAt,
		TanggalLahir:     employeeDateString(emp.TanggalLahir),
	}
}

func sanitizeEmployees(employees []db.ListEmployeesRow) []employeeResponse {
	items := make([]employeeResponse, 0, len(employees))
	for _, emp := range employees {
		items = append(items, employeeResponse{
			ID:               emp.ID,
			PegawaiUID:       emp.PegawaiUid,
			Nip:              emp.Nip,
			Nama:             emp.Nama,
			UnitKerja:        emp.UnitKerja,
			EmploymentType:   emp.EmploymentType,
			JenisKelamin:     emp.JenisKelamin,
			TempatLahir:      emp.TempatLahir,
			PusakaUsername:   emp.PusakaUsername,
			PusakaIsEnabled:  emp.PusakaIsEnabled,
			PusakaEligible:   emp.EmploymentType == "pns" || emp.EmploymentType == "pppk",
			HasPusakaAccount: emp.PusakaUsername != "",
			IsActive:         emp.IsActive,
			CreatedAt:        emp.CreatedAt,
			UpdatedAt:        emp.UpdatedAt,
			TanggalLahir:     employeeDateString(emp.TanggalLahir),
		})
	}
	return items
}
