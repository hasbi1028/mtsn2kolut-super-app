package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Employee struct {
	svc *service.Employee
}

type employeeResponse struct {
	ID              pgtype.UUID        `json:"id"`
	Nip             string             `json:"nip"`
	Nama            string             `json:"nama"`
	UnitKerja       string             `json:"unit_kerja"`
	EmploymentType  string             `json:"employment_type"`
	PusakaUsername  string             `json:"pusaka_username"`
	PusakaIsEnabled bool               `json:"pusaka_is_enabled"`
	PusakaEligible  bool               `json:"pusaka_eligible"`
	HasPusakaAccount bool              `json:"has_pusaka_account"`
	IsActive        bool               `json:"is_active"`
	CreatedAt       pgtype.Timestamptz `json:"created_at"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
}

func NewEmployee(svc *service.Employee) *Employee { return &Employee{svc: svc} }

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
	if r.URL.Query().Get("with_status") == "1" {
		h.listWithStatus(w, r)
		return
	}
	employees, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, sanitizeEmployees(employees))
}

func (h *Employee) listWithStatus(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("scope") == "pusaka" {
		h.listPusakaEligibleWithStatus(w, r)
		return
	}
	rows, err := h.svc.ListWithStatus(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Employee) listPusakaEligibleWithStatus(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.ListPusakaEligibleWithStatus(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Employee) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
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
	var body struct {
		Nip            string `json:"nip"`
		Nama           string `json:"nama"`
		UnitKerja      string `json:"unit_kerja"`
		EmploymentType string `json:"employment_type"`
		PusakaUsername string `json:"pusaka_username"`
		PusakaPassword string `json:"pusaka_password"`
		IsActive       bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	emp, err := h.svc.Create(r.Context(), body.Nip, body.Nama, body.UnitKerja, body.EmploymentType, body.PusakaUsername, body.PusakaPassword, body.IsActive)
	if err != nil {
		if err.Error() == "only pns or pppk employees can have pusaka accounts" || err.Error() == "invalid employment type" {
			api.BadRequest(w, err.Error())
			return
		}
		api.Internal(w, err)
		return
	}
	api.Created(w, sanitizeEmployee(emp))
}

func (h *Employee) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var p db.UpdateEmployeeParams
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	p.ID = id
	emp, err := h.svc.Update(r.Context(), p)
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		if err.Error() == "invalid employment type" || err.Error() == "disable or remove the pusaka account before changing employee type" {
			api.BadRequest(w, err.Error())
			return
		}
		api.Internal(w, err)
		return
	}
	api.OK(w, sanitizeEmployee(emp))
}

func (h *Employee) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *Employee) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
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
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
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
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		PusakaUsername string `json:"pusaka_username"`
		PusakaPassword string `json:"pusaka_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
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
		IsActive:       emp.IsActive,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	if err := h.svc.UpsertPusakaAccount(r.Context(), id, body.PusakaUsername, body.PusakaPassword, emp.IsActive); err != nil {
		api.Internal(w, err)
		return
	}
	meta, _ := json.Marshal(map[string]any{
		"employee_id":      pgUUIDString(id),
		"employment_type":  emp.EmploymentType,
		"pusaka_username":  body.PusakaUsername,
		"configured":       body.PusakaUsername != "",
	})
	_ = h.svc.CreateAuditLog(r.Context(), auditUserID(r), "PUSAKA_ACCOUNT_UPDATE", "pusaka_account", pgUUIDString(id), meta)
	api.OK(w, map[string]any{
		"status":     "success",
		"configured": body.PusakaUsername != "" || updated.PusakaUsername != "",
	})
}

func (h *Employee) UpdatePusakaAccountStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		IsEnabled bool `json:"is_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if err := h.svc.SetPusakaAccountEnabled(r.Context(), id, body.IsEnabled); err != nil {
		if err.Error() == "pusaka account is not configured" {
			api.BadRequest(w, err.Error())
			return
		}
		api.Internal(w, err)
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
		Nip:              emp.Nip,
		Nama:             emp.Nama,
		UnitKerja:        emp.UnitKerja,
		EmploymentType:   emp.EmploymentType,
		PusakaUsername:   emp.PusakaUsername,
		PusakaIsEnabled: emp.PusakaIsEnabled,
		PusakaEligible:   emp.EmploymentType == "pns" || emp.EmploymentType == "pppk",
		HasPusakaAccount: emp.PusakaUsername != "",
		IsActive:         emp.IsActive,
		CreatedAt:        emp.CreatedAt,
		UpdatedAt:        emp.UpdatedAt,
	}
}

func sanitizeEmployees(employees []db.ListEmployeesRow) []employeeResponse {
	items := make([]employeeResponse, 0, len(employees))
	for _, emp := range employees {
		items = append(items, employeeResponse{
			ID:               emp.ID,
			Nip:              emp.Nip,
			Nama:             emp.Nama,
			UnitKerja:        emp.UnitKerja,
			EmploymentType:   emp.EmploymentType,
			PusakaUsername:   emp.PusakaUsername,
			PusakaIsEnabled: emp.PusakaIsEnabled,
			PusakaEligible:   emp.EmploymentType == "pns" || emp.EmploymentType == "pppk",
			HasPusakaAccount: emp.PusakaUsername != "",
			IsActive:         emp.IsActive,
			CreatedAt:        emp.CreatedAt,
			UpdatedAt:        emp.UpdatedAt,
		})
	}
	return items
}
