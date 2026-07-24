package handler

import (
	"net/http"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type TimetableHandler struct {
	svc *service.TimetableService
}

func NewTimetableHandler(svc *service.TimetableService) *TimetableHandler {
	return &TimetableHandler{svc: svc}
}

// GET /api/academic/timetable/weekly?academic_year_id=
func (h *TimetableHandler) GetWeeklyData(w http.ResponseWriter, r *http.Request) {
	ayID := r.URL.Query().Get("academic_year_id")
	if ayID == "" {
		activeAY, err := h.svc.GetActiveAcademicYear(r.Context())
		if err != nil {
			api.BadRequest(w, "tidak ada tahun akademik aktif")
			return
		}
		ayID = activeAY
	}
	data, err := h.svc.GetWeeklyData(r.Context(), ayID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

// POST /api/academic/timetable/slots
func (h *TimetableHandler) CreateSlot(w http.ResponseWriter, r *http.Request) {
	var body service.CreateSlotRequest
	if !decodeJSON(w, r, &body, 0) {
		return
	}
	if err := h.svc.CreateSlot(r.Context(), body); err != nil {
		writeDomainOrInternal(w, err, "gagal membuat slot jadwal")
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

// PUT /api/academic/timetable/slots/{id}
func (h *TimetableHandler) UpdateSlot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id slot wajib diisi")
		return
	}
	var body service.UpdateSlotRequest
	if !decodeJSON(w, r, &body, 0) {
		return
	}
	body.ID = id
	if err := h.svc.UpdateSlot(r.Context(), body); err != nil {
		writeDomainOrInternal(w, err, "gagal mengupdate slot jadwal")
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

// DELETE /api/academic/timetable/slots/{id}
func (h *TimetableHandler) DeleteSlot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id slot wajib diisi")
		return
	}
	if err := h.svc.DeleteSlot(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus slot jadwal")
		return
	}
	api.NoContent(w)
}
