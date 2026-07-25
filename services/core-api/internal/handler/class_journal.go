package handler

import (
	"net/http"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type ClassJournal struct {
	svc *service.ClassJournal
}

func NewClassJournal(svc *service.ClassJournal) *ClassJournal {
	return &ClassJournal{svc: svc}
}

// GET /api/class-journal?assignment_id=
func (h *ClassJournal) Overview(w http.ResponseWriter, r *http.Request) {
	assignmentID := r.URL.Query().Get("assignment_id")
	if assignmentID == "" {
		api.BadRequest(w, "assignment_id wajib diisi")
		return
	}
	items, err := h.svc.Overview(r.Context(), assignmentID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, items)
}

// POST /api/class-journal/sessions
func (h *ClassJournal) CreateSession(w http.ResponseWriter, r *http.Request) {
	var body service.CreateSessionRequest
	if !decodeJSON(w, r, &body, 0) {
		return
	}
	if body.AssignmentID == "" {
		api.BadRequest(w, "assignment_id wajib diisi")
		return
	}
	if body.Tanggal == "" {
		api.BadRequest(w, "tanggal wajib diisi")
		return
	}
	result, err := h.svc.CreateSession(r.Context(), body)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal membuat jurnal")
		return
	}
	api.OK(w, result)
}

// POST /api/academic/rombel/{id}/timetable-slots/{slotID}/journal-session
func (h *ClassJournal) OpenSessionFromTimetableSlot(w http.ResponseWriter, r *http.Request) {
	classID := r.PathValue("id")
	slotID := r.PathValue("slotID")
	if classID == "" || slotID == "" {
		api.BadRequest(w, "rombel_id dan slot_id wajib diisi")
		return
	}

	var body struct {
		Tanggal   string `json:"tanggal"`
		Materi    string `json:"materi"`
		Kegiatan  string `json:"kegiatan"`
		Catatan   string `json:"catatan"`
		GuruHadir *bool  `json:"guru_hadir"`
	}
	if !decodeJSON(w, r, &body, 0) {
		return
	}
	if body.Tanggal == "" {
		api.BadRequest(w, "date wajib diisi")
		return
	}
	guruHadir := true
	if body.GuruHadir != nil {
		guruHadir = *body.GuruHadir
	}

	result, err := h.svc.OpenFromTimetableSlot(r.Context(), classID, slotID, body.Tanggal, body.Materi, body.Catatan, guruHadir)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal membuka jurnal")
		return
	}
	api.OK(w, result)
}

// PUT /api/class-journal/sessions/{id}/attendances
func (h *ClassJournal) BulkUpsertAttendances(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		api.BadRequest(w, "session_id wajib diisi")
		return
	}

	var body struct {
		Records []service.AttendanceRecord `json:"records"`
	}
	if !decodeJSON(w, r, &body, 0) {
		return
	}

	if err := h.svc.BulkUpsertAttendances(r.Context(), sessionID, body.Records); err != nil {
		writeDomainOrInternal(w, err, "gagal menyimpan kehadiran")
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

// GET /api/class-journal/sessions/{id}/attendances
func (h *ClassJournal) ListAttendances(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		api.BadRequest(w, "session_id wajib diisi")
		return
	}
	items, err := h.svc.ListAttendances(r.Context(), sessionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, items)
}

// DELETE /api/class-journal/sessions/{id}
func (h *ClassJournal) DeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		api.BadRequest(w, "session_id wajib diisi")
		return
	}
	if err := h.svc.DeleteSession(r.Context(), sessionID); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus jurnal")
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

// GET /api/class-journal/summary?assignment_id=
func (h *ClassJournal) AttendanceSummary(w http.ResponseWriter, r *http.Request) {
	assignmentID := r.URL.Query().Get("assignment_id")
	if assignmentID == "" {
		api.BadRequest(w, "assignment_id wajib diisi")
		return
	}
	items, err := h.svc.AttendanceSummary(r.Context(), assignmentID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, items)
}
