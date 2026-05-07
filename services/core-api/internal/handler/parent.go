package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type parentService interface {
	List(ctx context.Context) ([]db.Parent, error)
	Get(ctx context.Context, id pgtype.UUID) (db.Parent, error)
	Create(ctx context.Context, nama, phone, address string) (db.Parent, error)
	Update(ctx context.Context, id pgtype.UUID, nama, phone, address string) (db.Parent, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	LinkStudent(ctx context.Context, parentID, studentID pgtype.UUID) error
	UnlinkStudent(ctx context.Context, parentID, studentID pgtype.UUID) error
	ListChildren(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error)
}

type Parent struct {
	svc parentService
}

func NewParent(svc *service.Parent) *Parent { return &Parent{svc: svc} }

func (h *Parent) List(w http.ResponseWriter, r *http.Request) {
	if !parentAccessAllowed(r, "parents.read", "parents.manage") {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Parent) Get(w http.ResponseWriter, r *http.Request) {
	if !parentAccessAllowed(r, "parents.read", "parents.manage") {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Parent) Create(w http.ResponseWriter, r *http.Request) {
	if !parentAccessAllowed(r, "parents.manage") {
		api.Forbidden(w)
		return
	}
	var body struct {
		Nama    string `json:"nama"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.Create(r.Context(), body.Nama, body.Phone, body.Address)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *Parent) Update(w http.ResponseWriter, r *http.Request) {
	if !parentAccessAllowed(r, "parents.manage") {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Nama    string `json:"nama"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.Update(r.Context(), id, body.Nama, body.Phone, body.Address)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Parent) Delete(w http.ResponseWriter, r *http.Request) {
	if !parentAccessAllowed(r, "parents.manage") {
		api.Forbidden(w)
		return
	}
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

func (h *Parent) LinkStudent(w http.ResponseWriter, r *http.Request) {
	if !parentAccessAllowed(r, "parents.manage") {
		api.Forbidden(w)
		return
	}
	parentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		StudentID string `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	studentID, err := parseUUID(body.StudentID)
	if err != nil {
		api.BadRequest(w, "invalid student_id")
		return
	}
	if err := h.svc.LinkStudent(r.Context(), parentID, studentID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "linked"})
}

func (h *Parent) UnlinkStudent(w http.ResponseWriter, r *http.Request) {
	if !parentAccessAllowed(r, "parents.manage") {
		api.Forbidden(w)
		return
	}
	parentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		StudentID string `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	studentID, err := parseUUID(body.StudentID)
	if err != nil {
		api.BadRequest(w, "invalid student_id")
		return
	}
	if err := h.svc.UnlinkStudent(r.Context(), parentID, studentID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "unlinked"})
}

func parentAccessAllowed(r *http.Request, permissions ...string) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin") || mw.HasAnyPermission(claims, permissions...)
}

func (h *Parent) ListChildren(w http.ResponseWriter, r *http.Request) {
	if !parentAccessAllowed(r, "parents.read", "parents.manage") {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.ListChildren(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}
