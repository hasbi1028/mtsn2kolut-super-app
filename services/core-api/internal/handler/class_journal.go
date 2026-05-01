package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type ClassJournal struct {
	svc   classJournalService
	audit cbtAuthoringAuditWriter
}

type classJournalService interface {
	Overview(ctx context.Context, assignmentID, employeeID pgtype.UUID) (service.JournalOverview, error)
	CreateSession(ctx context.Context, assignmentID pgtype.UUID, tanggal pgtype.Date, materi, kegiatan, catatan string, guruHadir bool, employeeID pgtype.UUID) (service.JournalSessionDetail, error)
	GetSession(ctx context.Context, id, employeeID pgtype.UUID) (service.JournalSessionDetail, error)
	UpdateSession(ctx context.Context, id pgtype.UUID, materi, kegiatan, catatan string, guruHadir bool, employeeID pgtype.UUID) (db.ClassJournalSession, error)
	DeleteSession(ctx context.Context, id, employeeID pgtype.UUID) error
	BulkUpsertAttendances(ctx context.Context, sessionID pgtype.UUID, entries []service.JournalAttendanceEntry, employeeID pgtype.UUID) error
}

func NewClassJournal(svc *service.ClassJournal, audit ...cbtAuthoringAuditWriter) *ClassJournal {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &ClassJournal{svc: svc, audit: writer}
}

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
		writeClientError(w, err, "Data sesi jurnal tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CLASS_JOURNAL_SESSION_CREATE", "class_journal_session", pgUUIDString(detail.Session.ID), map[string]any{
		"assignment_id": pgUUIDString(detail.Session.AssignmentID),
		"tanggal":       body.Tanggal,
		"materi":        body.Materi,
	})
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
	detail, err := h.svc.GetSession(r.Context(), id, journalEmployeeID(r))
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
		writeClientError(w, err, "Perubahan sesi jurnal tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CLASS_JOURNAL_SESSION_UPDATE", "class_journal_session", pgUUIDString(row.ID), map[string]any{
		"assignment_id": pgUUIDString(row.AssignmentID),
		"materi":        row.Materi,
	})
	api.OK(w, row)
}

func (h *ClassJournal) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteSession(r.Context(), id, journalEmployeeID(r)); err != nil {
		api.Internal(w, err)
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CLASS_JOURNAL_SESSION_DELETE", "class_journal_session", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
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
		writeClientError(w, err, "Data kehadiran jurnal tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CLASS_JOURNAL_ATTENDANCE_BULK_UPSERT", "class_journal_session", pgUUIDString(id), map[string]any{
		"entry_count": len(body.Entries),
		"updated_by":  currentUsername(r),
	})
	api.OK(w, map[string]string{"status": "ok"})
}

func journalAccessAllowed(r *http.Request) bool {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		return mw.HasAnyRole(claims, "admin", "guru")
	}
	return false
}

func journalEmployeeID(r *http.Request) pgtype.UUID {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if mw.HasAnyRole(claims, "admin") {
			return pgtype.UUID{}
		}
		isGuru := false
		if rawRoles, ok := claims["roles"].([]any); ok {
			for _, role := range rawRoles {
				if role == "guru" {
					isGuru = true
					break
				}
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
	return pgtype.UUID{}
}
