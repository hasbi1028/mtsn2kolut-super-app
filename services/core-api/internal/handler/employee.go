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
	ID             pgtype.UUID        `json:"id"`
	Nip            string             `json:"nip"`
	Nama           string             `json:"nama"`
	UnitKerja      string             `json:"unit_kerja"`
	PusakaUsername string             `json:"pusaka_username"`
	IsActive       bool               `json:"is_active"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
}

func NewEmployee(svc *service.Employee) *Employee { return &Employee{svc: svc} }

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
	rows, err := h.svc.ListWithStatus(r.Context())
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
	var p db.CreateEmployeeParams
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	emp, err := h.svc.Create(r.Context(), p)
	if err != nil {
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
		"employee_id":    id,
		"pusaka_username": emp.PusakaUsername,
		"configured":     configured,
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
		ID:            emp.ID,
		Nip:           emp.Nip,
		Nama:          emp.Nama,
		UnitKerja:     emp.UnitKerja,
		PusakaUsername: body.PusakaUsername,
		PusakaPassword: body.PusakaPassword,
		IsActive:      emp.IsActive,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{
		"status":    "success",
		"configured": updated.PusakaUsername != "",
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
		"employee_id":       id,
		"pusaka_username":    emp.PusakaUsername,
		"configured":         configured,
		"last_tested_at":    nil,
		"last_test_status":   nil,
	})
}

func (h *Employee) TestPusakaCredentials(w http.ResponseWriter, r *http.Request) {
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
	if emp.PusakaUsername == "" || emp.PusakaPassword == "" {
		api.BadRequest(w, "Pusaka credentials not configured")
		return
	}
	// Create a test job for the worker to verify credentials
	// For now, just return success - actual test would be done by worker
	api.OK(w, map[string]string{
		"status":  "success",
		"message": "Credentials configured",
	})
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	return u, u.Scan(s)
}

func sanitizeEmployee(emp db.Employee) employeeResponse {
	return employeeResponse{
		ID:             emp.ID,
		Nip:            emp.Nip,
		Nama:           emp.Nama,
		UnitKerja:      emp.UnitKerja,
		PusakaUsername: emp.PusakaUsername,
		IsActive:       emp.IsActive,
		CreatedAt:      emp.CreatedAt,
		UpdatedAt:      emp.UpdatedAt,
	}
}

func (h *Employee) TestPusakaCredentials(w http.ResponseWriter, r *http.Request) {
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
	if emp.PusakaUsername == "" || emp.PusakaPassword == "" {
		api.BadRequest(w, "Pusaka credentials not configured")
		return
	}
	// Create a test job for the worker to verify credentials
	// For now, just return success - actual test would be done by worker
	api.OK(w, map[string]string{
		"status":  "success",
		"message": "Credentials configured",
	})
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	return u, u.Scan(s)
}

func sanitizeEmployee(emp db.Employee) employeeResponse {
	return employeeResponse{
		ID:             emp.ID,
		Nip:            emp.Nip,
		Nama:           emp.Nama,
		UnitKerja:      emp.UnitKerja,
		PusakaUsername: emp.PusakaUsername,
		IsActive:       emp.IsActive,
		CreatedAt:      emp.CreatedAt,
		UpdatedAt:      emp.UpdatedAt,
	}
}

func sanitizeEmployees(employees []db.Employee) []employeeResponse {
	items := make([]employeeResponse, 0, len(employees))
	for _, emp := range employees {
		items = append(items, sanitizeEmployee(emp))
	}
	return items
}
	emp, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	configured := emp.PusakaUsername != "" && emp.PusakaPassword != ""
	api.OK(w, map[string]any{
		"employee_id":    id,
		"pusaka_username": emp.PusakaUsername,
		"configured":     configured,
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
		ID:            emp.ID,
		Nip:           emp.Nip,
		Nama:          emp.Nama,
		UnitKerja:     emp.UnitKerja,
		PusakaUsername: body.PusakaUsername,
		PusakaPassword: body.PusakaPassword,
		IsActive:      emp.IsActive,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{
		"status":    "success",
		"configured": updated.PusakaUsername != "",
	})
}

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	return u, u.Scan(s)
}

func sanitizeEmployee(emp db.Employee) employeeResponse {
	return employeeResponse{
		ID:             emp.ID,
		Nip:            emp.Nip,
		Nama:           emp.Nama,
		UnitKerja:      emp.UnitKerja,
		PusakaUsername: emp.PusakaUsername,
		IsActive:       emp.IsActive,
		CreatedAt:      emp.CreatedAt,
		UpdatedAt:      emp.UpdatedAt,
	}
}

func sanitizeEmployees(employees []db.Employee) []employeeResponse {
	items := make([]employeeResponse, 0, len(employees))
	for _, emp := range employees {
		items = append(items, sanitizeEmployee(emp))
	}
	return items
}
	return items
}
