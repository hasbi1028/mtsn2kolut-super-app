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

type studentService interface {
	List(ctx context.Context) ([]db.ListStudentsRow, error)
	ListByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListStudentsByTeacherRow, error)
	Create(ctx context.Context, p db.CreateStudentParams) (db.Student, error)
	Update(ctx context.Context, p db.UpdateStudentParams) (db.Student, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	UpdateLifecycle(ctx context.Context, id pgtype.UUID, status db.StudentStatusEnum) error
}

type Student struct {
	svc studentService
}

func NewStudent(svc *service.Student) *Student { return &Student{svc: svc} }

func (h *Student) List(w http.ResponseWriter, r *http.Request) {
	if !studentAdminAccessAllowed(r) {
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

func (h *Student) GuruAwareList(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if ok && mw.HasAnyRole(claims, "admin") {
		h.List(w, r)
		return
	}
	if ok && mw.HasAnyRole(claims, "guru") {
		eidRaw, _ := claims["eid"].(string)
		if eidRaw == "" {
			api.Forbidden(w)
			return
		}
		var eid pgtype.UUID
		if err := eid.Scan(eidRaw); err != nil {
			api.Forbidden(w)
			return
		}
		rows, err := h.svc.ListByTeacher(r.Context(), eid)
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.OK(w, rows)
		return
	}
	api.Forbidden(w)
}

func (h *Student) Create(w http.ResponseWriter, r *http.Request) {
	if !studentAdminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		Nis         string `json:"nis"`
		Nisn        string `json:"nisn"`
		Nama        string `json:"nama"`
		Gender      string `json:"gender"`
		ParentName  string `json:"parent_name"`
		ParentPhone string `json:"parent_phone"`
		ClassID     string `json:"class_id"`
		IsActive    bool   `json:"is_active"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	var classID pgtype.UUID
	if body.ClassID != "" {
		id, err := parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "Rombel tidak valid")
			return
		}
		classID = id
	}

	status := db.StudentStatusEnumActive
	if body.Status != "" {
		status = db.StudentStatusEnum(body.Status)
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
		Status:      status,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *Student) Update(w http.ResponseWriter, r *http.Request) {
	if !studentAdminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		Nis         string `json:"nis"`
		Nisn        string `json:"nisn"`
		Nama        string `json:"nama"`
		Gender      string `json:"gender"`
		ParentName  string `json:"parent_name"`
		ParentPhone string `json:"parent_phone"`
		ClassID     string `json:"class_id"`
		IsActive    bool   `json:"is_active"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	var classID pgtype.UUID
	if body.ClassID != "" {
		cid, err := parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "Rombel tidak valid")
			return
		}
		classID = cid
	}

	status := db.StudentStatusEnumActive
	if body.Status != "" {
		status = db.StudentStatusEnum(body.Status)
	}

	row, err := h.svc.Update(r.Context(), db.UpdateStudentParams{
		ID:          id,
		Nis:         body.Nis,
		Nisn:        body.Nisn,
		Nama:        body.Nama,
		Gender:      db.GenderEnum(body.Gender),
		ParentName:  body.ParentName,
		ParentPhone: body.ParentPhone,
		ClassID:     classID,
		IsActive:    body.IsActive,
		Status:      status,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Student) PublicRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nis         string `json:"nis"`
		Nama        string `json:"nama"`
		Gender      string `json:"gender"`
		ParentName  string `json:"parent_name"`
		ParentPhone string `json:"parent_phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	if body.Nama == "" || body.Nis == "" {
		api.BadRequest(w, "nama and nis required")
		return
	}

	row, err := h.svc.Create(r.Context(), db.CreateStudentParams{
		Nis:         body.Nis,
		Nama:        body.Nama,
		Gender:      db.GenderEnum(body.Gender),
		ParentName:  body.ParentName,
		ParentPhone: body.ParentPhone,
		IsActive:    true,
		Status:      db.StudentStatusEnumProspective,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *Student) Delete(w http.ResponseWriter, r *http.Request) {
	if !studentAdminAccessAllowed(r) {
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

func (h *Student) UpdateLifecycle(w http.ResponseWriter, r *http.Request) {
	if !studentAdminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	status := db.StudentStatusEnum(body.Status)
	switch status {
	case db.StudentStatusEnumProspective, db.StudentStatusEnumActive, db.StudentStatusEnumAlumni, db.StudentStatusEnumMutated:
	default:
		api.BadRequest(w, "status tidak didukung")
		return
	}
	if err := h.svc.UpdateLifecycle(r.Context(), id, status); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"id":     id,
		"status": status,
	})
}

func studentAdminAccessAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin")
}
