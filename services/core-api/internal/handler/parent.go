package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Parent struct {
	svc *service.Parent
}

func NewParent(svc *service.Parent) *Parent { return &Parent{svc: svc} }

func (h *Parent) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Parent) Get(w http.ResponseWriter, r *http.Request) {
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
	parentID, _ := parseUUID(chi.URLParam(r, "id"))
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

func (h *Parent) ListChildren(w http.ResponseWriter, r *http.Request) {
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
