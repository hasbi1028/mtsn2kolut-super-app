package handler

import (
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

// GET /api/kesiswaan/murid?search=&status=&class_id=
func (h *KesiswaanHandler) ListMurid(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	classID := r.URL.Query().Get("class_id")

	items, err := h.svc.ListStudents(r.Context(), search, status, classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, items)
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
