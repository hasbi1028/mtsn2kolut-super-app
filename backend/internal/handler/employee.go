package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pusaka/backend/internal/api"
	db "github.com/pusaka/backend/internal/repository/postgres"
	"github.com/pusaka/backend/internal/service"
)

type Employee struct {
	svc *service.Employee
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
	api.OK(w, employees)
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
	api.OK(w, emp)
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
	api.Created(w, emp)
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
	api.OK(w, emp)
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

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	return u, u.Scan(s)
}
