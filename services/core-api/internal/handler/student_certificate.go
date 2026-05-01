package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type StudentCertificate struct {
	svc *service.StudentCertificate
}

func NewStudentCertificate(svc *service.StudentCertificate) *StudentCertificate {
	return &StudentCertificate{svc: svc}
}

// GET /api/tu/surat-keterangan/templates
func (h *StudentCertificate) ListTemplates(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.ListTemplates(r.Context(), true)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

// GET /api/tu/surat-keterangan/students
func (h *StudentCertificate) ListStudents(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.ListStudents(r.Context(), r.URL.Query().Get("search"), r.URL.Query().Get("status"))
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, data)
}

// GET /api/tu/surat-keterangan
func (h *StudentCertificate) List(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.List(
		r.Context(),
		r.URL.Query().Get("search"),
		r.URL.Query().Get("status"),
		r.URL.Query().Get("template_code"),
	)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, data)
}

// POST /api/tu/surat-keterangan
func (h *StudentCertificate) Create(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		TemplateID   string `json:"template_id"`
		StudentID    string `json:"student_id"`
		TanggalSurat string `json:"tanggal_surat"`
		Purpose      string `json:"purpose"`
		Recipient    string `json:"recipient"`
		Remarks      string `json:"remarks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	data, err := h.svc.Create(r.Context(), service.CreateStudentCertificateInput{
		TemplateID:         body.TemplateID,
		StudentID:          body.StudentID,
		TanggalSurat:       body.TanggalSurat,
		Purpose:            body.Purpose,
		Recipient:          body.Recipient,
		Remarks:            body.Remarks,
		CreatedByUserID:    inventoryActorUserID(r),
		IssuedByEmployeeID: tuEmployeeID(r),
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "uq_outgoing") {
			api.Conflict(w, "nomor surat sudah digunakan")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			api.NotFound(w)
			return
		}
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, data)
}

// GET /api/tu/surat-keterangan/{id}
func (h *StudentCertificate) Get(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	data, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.NotFound(w)
			return
		}
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

// POST /api/tu/surat-keterangan/{id}/cancel
func (h *StudentCertificate) Cancel(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body struct {
		Remarks string `json:"remarks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	data, err := h.svc.Cancel(r.Context(), id, body.Remarks)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.NotFound(w)
			return
		}
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, data)
}
