package handler

import (
	"fmt"
	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
	"net/http"
)

type KesiswaanHandler struct {
	svc *service.KesiswaanService
}

func NewKesiswaanHandler(svc *service.KesiswaanService) *KesiswaanHandler {
	return &KesiswaanHandler{svc: svc}
}

// GET /api/kesiswaan/murid?search=&status=&class_id=&page=1&per_page=25
func (h *KesiswaanHandler) ListMurid(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	classID := r.URL.Query().Get("class_id")
	page := parseIntParam(r.URL.Query().Get("page"), 1)
	perPage := parseIntParam(r.URL.Query().Get("per_page"), 25)

	if perPage > 100 {
		perPage = 100
	}

	items, total, err := h.svc.ListStudents(r.Context(), search, status, classID, page, perPage)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.JSON(w, http.StatusOK, map[string]any{
		"data":     items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"pages":    (total + int64(perPage) - 1) / int64(perPage),
	})
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var v int
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil || v < 1 {
		return defaultVal
	}
	return v
}

// PUT /api/kesiswaan/murid/{id}/profile
func (h *KesiswaanHandler) UpdateMuridProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id murid wajib diisi")
		return
	}

	var body struct {
		NIK          string `json:"nik"`
		TempatLahir  string `json:"tempat_lahir"`
		TanggalLahir string `json:"tanggal_lahir"`
		Alamat       string `json:"alamat"`
		Agama        string `json:"agama"`
		AnakKe       int    `json:"anak_ke"`
		Phone        string `json:"phone"`
		ParentName   string `json:"parent_name"`
		ParentPhone  string `json:"parent_phone"`
	}
	if !decodeJSON(w, r, &body, 0) {
		return
	}

	err := h.svc.UpdateProfile(r.Context(), id, body.NIK, body.TempatLahir, body.Alamat, body.Agama, body.Phone, body.ParentName, body.ParentPhone, body.TanggalLahir, body.AnakKe)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

// POST /api/kesiswaan/murid
func (h *KesiswaanHandler) CreateMurid(w http.ResponseWriter, r *http.Request) {
	var body service.CreateStudentRequest
	if !decodeJSON(w, r, &body, 0) {
		return
	}
	if body.NIS == "" || body.Nama == "" || body.Gender == "" {
		api.BadRequest(w, "NIS, nama, dan gender wajib diisi")
		return
	}
	if body.Gender != "L" && body.Gender != "P" {
		api.BadRequest(w, "Gender harus L atau P")
		return
	}
	id, err := h.svc.CreateStudent(r.Context(), body)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal membuat murid")
		return
	}
	api.OK(w, map[string]string{"id": id})
}

// DELETE /api/kesiswaan/murid/{id}
func (h *KesiswaanHandler) DeleteMurid(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id murid wajib diisi")
		return
	}
	err := h.svc.DeleteStudent(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus murid")
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}
