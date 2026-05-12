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
	OpenSessionFromTimetableSlot(ctx context.Context, classID, slotID pgtype.UUID, tanggal pgtype.Date, materi, kegiatan, catatan string, guruHadir bool, employeeID pgtype.UUID) (service.JournalSessionOpenResult, error)
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
	employeeID, ok := journalReadScope(w, r)
	if !ok {
		return
	}
	assignmentID, err := optionalUUID(r.URL.Query().Get("assignment_id"))
	if err != nil {
		api.BadRequest(w, "assignment_id tidak valid")
		return
	}
	data, err := h.svc.Overview(r.Context(), assignmentID, employeeID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *ClassJournal) CreateSession(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := journalManageScope(w, r)
	if !ok {
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
		api.BadRequest(w, "Data yang dikirim tidak valid")
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

func (h *ClassJournal) OpenSessionFromTimetableSlot(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := journalManageScopeOrUnauthorized(w, r)
	if !ok {
		return
	}
	classID, slotID, ok := parseRombelTimetableSlotRoute(w, r)
	if !ok {
		return
	}
	var body struct {
		Date      string `json:"date"`
		Tanggal   string `json:"tanggal"`
		Topic     string `json:"topic"`
		Materi    string `json:"materi"`
		Kegiatan  string `json:"kegiatan"`
		Notes     string `json:"notes"`
		Catatan   string `json:"catatan"`
		GuruHadir *bool  `json:"guru_hadir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	dateValue := strings.TrimSpace(body.Date)
	if dateValue == "" {
		dateValue = strings.TrimSpace(body.Tanggal)
	}
	if dateValue == "" {
		api.BadRequest(w, "date wajib diisi")
		return
	}
	var tanggal pgtype.Date
	if err := tanggal.Scan(dateValue); err != nil {
		api.BadRequest(w, "format date tidak valid (gunakan YYYY-MM-DD)")
		return
	}
	materi := body.Materi
	if strings.TrimSpace(materi) == "" {
		materi = body.Topic
	}
	catatan := body.Catatan
	if strings.TrimSpace(catatan) == "" {
		catatan = body.Notes
	}
	guruHadir := true
	if body.GuruHadir != nil {
		guruHadir = *body.GuruHadir
	}

	result, err := h.svc.OpenSessionFromTimetableSlot(r.Context(), classID, slotID, tanggal, materi, body.Kegiatan, catatan, guruHadir, employeeID)
	if err != nil {
		writeClientError(w, err, "Sesi jurnal dari jadwal rombel tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CLASS_JOURNAL_SESSION_OPEN_FROM_TIMETABLE", "class_journal_session", pgUUIDString(result.Session.ID), map[string]any{
		"assignment_id":     pgUUIDString(result.Session.AssignmentID),
		"timetable_slot_id": pgUUIDString(slotID),
		"class_id":          pgUUIDString(classID),
		"tanggal":           dateValue,
		"created":           result.Created,
		"opened_by":         currentUsername(r),
	})
	if result.Created {
		api.Created(w, result)
		return
	}
	api.OK(w, result)
}

func (h *ClassJournal) GetSession(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := journalReadScope(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	detail, err := h.svc.GetSession(r.Context(), id, employeeID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, detail)
}

func (h *ClassJournal) UpdateSession(w http.ResponseWriter, r *http.Request) {
	employeeID, ok := journalManageScope(w, r)
	if !ok {
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
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
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
	employeeID, ok := journalManageScope(w, r)
	if !ok {
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
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
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
	_, ok := journalScope(r, false)
	return ok
}

func journalReadScope(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	return writeJournalScope(w, r, false)
}

func journalManageScope(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	return writeJournalScope(w, r, true)
}

func journalManageScopeOrUnauthorized(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	return writeJournalScopeWithMissingClaimsStatus(w, r, true, http.StatusUnauthorized)
}

func writeJournalScope(w http.ResponseWriter, r *http.Request, manage bool) (pgtype.UUID, bool) {
	return writeJournalScopeWithMissingClaimsStatus(w, r, manage, http.StatusForbidden)
}

func writeJournalScopeWithMissingClaimsStatus(w http.ResponseWriter, r *http.Request, manage bool, missingClaimsStatus int) (pgtype.UUID, bool) {
	if _, ok := api.ClaimsFromContext(r.Context()); !ok {
		if missingClaimsStatus == http.StatusUnauthorized {
			api.Unauthorized(w)
		} else {
			api.Forbidden(w)
		}
		return pgtype.UUID{}, false
	}
	employeeID, ok := journalScope(r, manage)
	if !ok {
		api.Forbidden(w)
		return pgtype.UUID{}, false
	}
	return employeeID, true
}

func journalScope(r *http.Request, manage bool) (pgtype.UUID, bool) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return pgtype.UUID{}, false
	}
	if mw.HasAnyRole(claims, "admin") {
		return pgtype.UUID{}, true
	}
	if manage {
		if mw.HasAnyPermission(claims, "journal.manage_all") {
			return pgtype.UUID{}, true
		}
		if !mw.HasAnyRole(claims, "guru") && !mw.HasAnyPermission(claims, "journal.manage") {
			return pgtype.UUID{}, false
		}
	} else {
		if mw.HasAnyPermission(claims, "journal.read_all", "journal.manage_all") {
			return pgtype.UUID{}, true
		}
		if !mw.HasAnyRole(claims, "guru") && !mw.HasAnyPermission(claims, "journal.read", "journal.manage") {
			return pgtype.UUID{}, false
		}
	}
	employeeID := journalEmployeeID(r)
	return employeeID, employeeID.Valid
}

func journalEmployeeID(r *http.Request) pgtype.UUID {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if mw.HasAnyRole(claims, "admin") || mw.HasAnyPermission(claims, "journal.read_all", "journal.manage_all") {
			return pgtype.UUID{}
		}
		if eid, _ := claims["eid"].(string); eid != "" {
			var id pgtype.UUID
			if err := id.Scan(strings.TrimSpace(eid)); err == nil {
				return id
			}
		}
	}
	return pgtype.UUID{}
}
