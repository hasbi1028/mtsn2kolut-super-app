package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type ClassJournal struct {
	svc *service.ClassJournal
}

func NewClassJournal(svc *service.ClassJournal) *ClassJournal { return &ClassJournal{svc: svc} }

func (h *ClassJournal) Overview(w http.ResponseWriter, r *http.Request) {
	if !journalAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	assignmentID, err := optionalUUID(r.URL.Query().Get("assignment_id"))
	if err != nil {
		api.BadRequest(w, "assignment_id tidak valid")
		return
	}
	employeeID := journalEmployeeID(r)
	data, err := h.svc.Overview(r.Context(), assignmentID, employeeID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *ClassJournal) CreateSession(w http.ResponseWriter, r *http.Request) {
	if !journalAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		AssignmentID string `json:"assignment_id"`
		Tanggal      string `json:"tanggal"`
		Materi       string `json:"materi"`
		Kegiatan     string `json:"kegiatan"`
		Catatan      string `json:"catatan"`
		GuruHadir    bool   `json:"guru_hadir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	assignmentID, err := parseUUID(body.AssignmentID)
	if err != nil {
		api.BadRequest(w, "assignment_id tidak valid")
		return
	}
	if strings.TrimSpace(body.Tanggal) == "" {
		api.BadRequest(w, "tanggal wajib diisi")
		return
	}
	var tanggal pgtype.Date
	if err := tanggal.Scan(body.Tanggal); err != nil {
		api.BadRequest(w, "format tanggal tidak valid (gunakan YYYY-MM-DD)")
		return
	}
	employeeID := journalEmployeeID(r)
	detail, err := h.svc.CreateSession(r.Context(), assignmentID, tanggal, body.Materi, body.Kegiatan, body.Catatan, body.GuruHadir, employeeID)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "akses ditolak") {
			api.Forbidden(w)
			return
		}
		if strings.Contains(msg, "sudah ada") {
			api.Conflict(w, msg)
			return
		}
		api.BadRequest(w, msg)
		return
	}
	api.Created(w, detail)
}

func (h *ClassJournal) GetSession(w http.ResponseWriter, r *http.Request) {
	if !journalAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	detail, err := h.svc.GetSession(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, detail)
}

func (h *ClassJournal) UpdateSession(w http.ResponseWriter, r *http.Request) {
	if !journalAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body struct {
		Materi    string `json:"materi"`
		Kegiatan  string `json:"kegiatan"`
		Catatan   string `json:"catatan"`
		GuruHadir bool   `json:"guru_hadir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	employeeID := journalEmployeeID(r)
	row, err := h.svc.UpdateSession(r.Context(), id, body.Materi, body.Kegiatan, body.Catatan, body.GuruHadir, employeeID)
	if err != nil {
		if strings.Contains(err.Error(), "akses ditolak") {
			api.Forbidden(w)
			return
		}
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *ClassJournal) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteSession(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *ClassJournal) BulkUpsertAttendances(w http.ResponseWriter, r *http.Request) {
	if !journalAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body struct {
		Entries []service.JournalAttendanceEntry `json:"entries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	employeeID := journalEmployeeID(r)
	if err := h.svc.BulkUpsertAttendances(r.Context(), id, body.Entries, employeeID); err != nil {
		if strings.Contains(err.Error(), "akses ditolak") {
			api.Forbidden(w)
			return
		}
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

func journalAccessAllowed(r *http.Request) bool {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if rawRoles, ok := claims["roles"].([]any); ok {
			for _, role := range rawRoles {
				if role == "admin" || role == "guru" {
					return true
				}
			}
		}
		if role, _ := claims["role"].(string); role == "admin" || role == "guru" {
			return true
		}
		return false
	}
	return true
}

func journalEmployeeID(r *http.Request) pgtype.UUID {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if rawRoles, ok := claims["roles"].([]any); ok {
			isGuru := false
			for _, role := range rawRoles {
				if role == "guru" {
					isGuru = true
					break
				}
			}
			if role, _ := claims["role"].(string); role == "guru" {
				isGuru = true
			}
			if isGuru {
				if eid, _ := claims["eid"].(string); eid != "" {
					var id pgtype.UUID
					if err := id.Scan(strings.TrimSpace(eid)); err == nil {
						return id
					}
				}
			}
		}
	}
	return pgtype.UUID{}
}

