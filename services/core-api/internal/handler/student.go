package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Student struct {
	svc *service.Student
}

func NewStudent(svc *service.Student) *Student { return &Student{svc: svc} }

func (h *Student) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Student) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nis         string `json:"nis"`
		Nisn        string `json:"nisn"`
		Nama        string `json:"nama"`
		Gender      string `json:"gender"`
		ParentName  string `json:"parent_name"`
		ParentPhone string `json:"parent_phone"`
		ClassID     string `json:"class_id"`
		IsActive    bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	var classID pgtype.UUID
	if body.ClassID != "" {
		id, err := parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "class_id invalid")
			return
		}
		classID = id
	}
	row, err := h.svc.Create(r.Context(), db.CreateStudentParams{
		Nis:         body.Nis,
		Nisn:        body.Nisn,
		Nama:        body.Nama,
		Gender:      db.GenderEnum(body.Gender),
		ParentName:  body.ParentName,
		ParentPhone: body.ParentPhone,
		ClassID:     classID,
		IsActive:    body.IsActive,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *Student) Delete(w http.ResponseWriter, r *http.Request) {
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
