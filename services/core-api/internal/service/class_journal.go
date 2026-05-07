package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type classJournalStore interface {
	CountJournalSessionsForAssignment(ctx context.Context, assignmentID pgtype.UUID) (int32, error)
	CreateJournalSession(ctx context.Context, arg db.CreateJournalSessionParams) (db.ClassJournalSession, error)
	GetJournalSession(ctx context.Context, id pgtype.UUID) (db.GetJournalSessionRow, error)
	GetJournalSessionIDByAssignmentDate(ctx context.Context, arg db.GetJournalSessionIDByAssignmentDateParams) (pgtype.UUID, error)
	ListJournalSessions(ctx context.Context, assignmentID pgtype.UUID) ([]db.ListJournalSessionsRow, error)
	UpdateJournalSession(ctx context.Context, arg db.UpdateJournalSessionParams) (db.ClassJournalSession, error)
	DeleteJournalSession(ctx context.Context, id pgtype.UUID) error
	UpsertJournalAttendance(ctx context.Context, arg db.UpsertJournalAttendanceParams) (db.ClassJournalAttendance, error)
	ListJournalAttendances(ctx context.Context, sessionID pgtype.UUID) ([]db.ListJournalAttendancesRow, error)
	ListJournalAttendanceSummary(ctx context.Context, assignmentID pgtype.UUID) ([]db.ListJournalAttendanceSummaryRow, error)
	ListActiveStudentsByClassID(ctx context.Context, classID pgtype.UUID) ([]db.ListActiveStudentsByClassIDRow, error)
	ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error)
	GetClassSubjectAssignment(ctx context.Context, id pgtype.UUID) (db.GetClassSubjectAssignmentRow, error)
	GetRombelTimetableSlot(ctx context.Context, arg db.GetRombelTimetableSlotParams) (db.GetRombelTimetableSlotRow, error)
}

type classJournalTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type ClassJournal struct {
	q  classJournalStore
	tx classJournalTxStarter
}

func NewClassJournal(q *db.Queries) *ClassJournal { return &ClassJournal{q: q} }

func NewClassJournalWithPool(pool *pgxpool.Pool) *ClassJournal {
	if pool == nil {
		return &ClassJournal{q: db.New(nil)}
	}
	return &ClassJournal{q: db.New(pool), tx: pool}
}

type JournalSessionDetail struct {
	Session     db.GetJournalSessionRow        `json:"session"`
	Attendances []db.ListJournalAttendancesRow `json:"attendances"`
}

type JournalSessionOpenResult struct {
	Session       db.GetJournalSessionRow        `json:"session"`
	Attendances   []db.ListJournalAttendancesRow `json:"attendances"`
	TimetableSlot db.GetRombelTimetableSlotRow   `json:"timetable_slot"`
	Created       bool                           `json:"created"`
}

type JournalOverview struct {
	Assignments []db.ListClassSubjectAssignmentsRow  `json:"assignments"`
	Sessions    []db.ListJournalSessionsRow          `json:"sessions"`
	Summary     []db.ListJournalAttendanceSummaryRow `json:"summary"`
}

type JournalAttendanceEntry struct {
	StudentID string `json:"student_id"`
	Status    string `json:"status"`
	Catatan   string `json:"catatan"`
}

func (s *ClassJournal) Overview(ctx context.Context, assignmentID, employeeID pgtype.UUID) (JournalOverview, error) {
	allAssignments, err := s.q.ListClassSubjectAssignments(ctx)
	if err != nil {
		return JournalOverview{}, err
	}
	assignments := allAssignments
	if employeeID.Valid {
		assignments = filterJournalAssignmentsByTeacher(allAssignments, employeeID)
	}
	out := JournalOverview{
		Assignments: assignments,
		Sessions:    []db.ListJournalSessionsRow{},
		Summary:     []db.ListJournalAttendanceSummaryRow{},
	}
	if assignmentID.Valid {
		if employeeID.Valid {
			assignment, err := s.q.GetClassSubjectAssignment(ctx, assignmentID)
			if err != nil {
				return JournalOverview{}, err
			}
			if assignment.TeacherEmployeeID != employeeID {
				return JournalOverview{}, fmt.Errorf("akses ditolak")
			}
		}
		out.Sessions, err = s.q.ListJournalSessions(ctx, assignmentID)
		if err != nil {
			return JournalOverview{}, err
		}
		out.Summary, err = s.q.ListJournalAttendanceSummary(ctx, assignmentID)
		if err != nil {
			return JournalOverview{}, err
		}
	}
	return out, nil
}

func (s *ClassJournal) GetSession(ctx context.Context, id, employeeID pgtype.UUID) (JournalSessionDetail, error) {
	return getJournalSessionWithStore(ctx, s.q, id, employeeID)
}

func getJournalSessionWithStore(ctx context.Context, store classJournalStore, id, employeeID pgtype.UUID) (JournalSessionDetail, error) {
	session, err := store.GetJournalSession(ctx, id)
	if err != nil {
		return JournalSessionDetail{}, err
	}
	if employeeID.Valid && session.TeacherEmployeeID != employeeID {
		return JournalSessionDetail{}, fmt.Errorf("akses ditolak")
	}
	attendances, err := store.ListJournalAttendances(ctx, id)
	if err != nil {
		return JournalSessionDetail{}, err
	}
	return JournalSessionDetail{Session: session, Attendances: attendances}, nil
}

func (s *ClassJournal) CreateSession(ctx context.Context, assignmentID pgtype.UUID, tanggal pgtype.Date, materi, kegiatan, catatan string, guruHadir bool, employeeID pgtype.UUID) (JournalSessionDetail, error) {
	var detail JournalSessionDetail
	err := s.withClassJournalStore(ctx, func(store classJournalStore) error {
		assignment, err := store.GetClassSubjectAssignment(ctx, assignmentID)
		if err != nil {
			return fmt.Errorf("assignment tidak ditemukan")
		}
		if employeeID.Valid && assignment.TeacherEmployeeID != employeeID {
			return fmt.Errorf("akses ditolak")
		}

		count, err := store.CountJournalSessionsForAssignment(ctx, assignmentID)
		if err != nil {
			return err
		}

		session, err := store.CreateJournalSession(ctx, db.CreateJournalSessionParams{
			AssignmentID: assignmentID,
			Tanggal:      tanggal,
			PertemuanKe:  count + 1,
			Materi:       strings.TrimSpace(materi),
			Kegiatan:     strings.TrimSpace(kegiatan),
			Catatan:      strings.TrimSpace(catatan),
			GuruHadir:    guruHadir,
		})
		if err != nil {
			if strings.Contains(err.Error(), "uq_journal_session_date") || strings.Contains(err.Error(), "unique") {
				return fmt.Errorf("pertemuan pada tanggal ini sudah ada untuk mata pelajaran ini")
			}
			return err
		}

		if err := seedJournalAttendances(ctx, store, session.ID, assignment.ClassID); err != nil {
			return err
		}

		detail, err = getJournalSessionWithStore(ctx, store, session.ID, employeeID)
		return err
	})
	if err != nil {
		return JournalSessionDetail{}, err
	}
	return detail, nil
}

func (s *ClassJournal) OpenSessionFromTimetableSlot(ctx context.Context, classID, slotID pgtype.UUID, tanggal pgtype.Date, materi, kegiatan, catatan string, guruHadir bool, employeeID pgtype.UUID) (JournalSessionOpenResult, error) {
	slot, err := s.q.GetRombelTimetableSlot(ctx, db.GetRombelTimetableSlotParams{
		ClassID: classID,
		ID:      slotID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return JournalSessionOpenResult{}, fmt.Errorf("%w: slot jadwal rombel tidak ditemukan", domain.ErrNotFound)
		}
		return JournalSessionOpenResult{}, err
	}
	if employeeID.Valid && slot.TeacherEmployeeID != employeeID {
		return JournalSessionOpenResult{}, fmt.Errorf("%w: akses ditolak", domain.ErrForbidden)
	}

	existing, err := s.getSessionByAssignmentDate(ctx, slot.AssignmentID, tanggal, employeeID)
	if err == nil {
		return JournalSessionOpenResult{
			Session:       existing.Session,
			Attendances:   existing.Attendances,
			TimetableSlot: slot,
			Created:       false,
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return JournalSessionOpenResult{}, err
	}

	var result JournalSessionOpenResult
	err = s.withClassJournalStore(ctx, func(store classJournalStore) error {
		count, err := store.CountJournalSessionsForAssignment(ctx, slot.AssignmentID)
		if err != nil {
			return err
		}
		session, err := store.CreateJournalSession(ctx, db.CreateJournalSessionParams{
			AssignmentID: slot.AssignmentID,
			Tanggal:      tanggal,
			PertemuanKe:  count + 1,
			Materi:       strings.TrimSpace(materi),
			Kegiatan:     strings.TrimSpace(kegiatan),
			Catatan:      strings.TrimSpace(catatan),
			GuruHadir:    guruHadir,
		})
		if err != nil {
			return err
		}

		if err := seedJournalAttendances(ctx, store, session.ID, slot.ClassID); err != nil {
			return err
		}

		detail, err := getJournalSessionWithStore(ctx, store, session.ID, employeeID)
		if err != nil {
			return err
		}
		result = JournalSessionOpenResult{
			Session:       detail.Session,
			Attendances:   detail.Attendances,
			TimetableSlot: slot,
			Created:       true,
		}
		return nil
	})
	if err != nil {
		if isJournalDuplicateDateError(err) {
			existing, getErr := s.getSessionByAssignmentDate(ctx, slot.AssignmentID, tanggal, employeeID)
			if getErr == nil {
				return JournalSessionOpenResult{
					Session:       existing.Session,
					Attendances:   existing.Attendances,
					TimetableSlot: slot,
					Created:       false,
				}, nil
			}
		}
		return JournalSessionOpenResult{}, err
	}
	return result, nil
}

func (s *ClassJournal) getSessionByAssignmentDate(ctx context.Context, assignmentID pgtype.UUID, tanggal pgtype.Date, employeeID pgtype.UUID) (JournalSessionDetail, error) {
	return getSessionByAssignmentDateWithStore(ctx, s.q, assignmentID, tanggal, employeeID)
}

func getSessionByAssignmentDateWithStore(ctx context.Context, store classJournalStore, assignmentID pgtype.UUID, tanggal pgtype.Date, employeeID pgtype.UUID) (JournalSessionDetail, error) {
	sessionID, err := store.GetJournalSessionIDByAssignmentDate(ctx, db.GetJournalSessionIDByAssignmentDateParams{
		AssignmentID: assignmentID,
		Tanggal:      tanggal,
	})
	if err != nil {
		return JournalSessionDetail{}, err
	}
	return getJournalSessionWithStore(ctx, store, sessionID, employeeID)
}

func (s *ClassJournal) UpdateSession(ctx context.Context, id pgtype.UUID, materi, kegiatan, catatan string, guruHadir bool, employeeID pgtype.UUID) (db.ClassJournalSession, error) {
	session, err := s.q.GetJournalSession(ctx, id)
	if err != nil {
		return db.ClassJournalSession{}, fmt.Errorf("sesi tidak ditemukan")
	}
	if employeeID.Valid && session.TeacherEmployeeID != employeeID {
		return db.ClassJournalSession{}, fmt.Errorf("akses ditolak")
	}
	return s.q.UpdateJournalSession(ctx, db.UpdateJournalSessionParams{
		ID:        id,
		Materi:    strings.TrimSpace(materi),
		Kegiatan:  strings.TrimSpace(kegiatan),
		Catatan:   strings.TrimSpace(catatan),
		GuruHadir: guruHadir,
	})
}

func (s *ClassJournal) DeleteSession(ctx context.Context, id, employeeID pgtype.UUID) error {
	if employeeID.Valid {
		session, err := s.q.GetJournalSession(ctx, id)
		if err != nil {
			return fmt.Errorf("sesi tidak ditemukan")
		}
		if session.TeacherEmployeeID != employeeID {
			return fmt.Errorf("akses ditolak")
		}
	}
	return s.q.DeleteJournalSession(ctx, id)
}

func (s *ClassJournal) BulkUpsertAttendances(ctx context.Context, sessionID pgtype.UUID, entries []JournalAttendanceEntry, employeeID pgtype.UUID) error {
	session, err := s.q.GetJournalSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("sesi tidak ditemukan")
	}
	if employeeID.Valid && session.TeacherEmployeeID != employeeID {
		return fmt.Errorf("akses ditolak")
	}
	for _, e := range entries {
		var studentID pgtype.UUID
		if err := studentID.Scan(strings.TrimSpace(e.StudentID)); err != nil {
			return fmt.Errorf("student_id tidak valid: %s", e.StudentID)
		}
		status, err := normalizeJournalAttendanceStatus(e.Status)
		if err != nil {
			return err
		}
		if _, err := s.q.UpsertJournalAttendance(ctx, db.UpsertJournalAttendanceParams{
			SessionID: sessionID,
			StudentID: studentID,
			Status:    status,
			Catatan:   strings.TrimSpace(e.Catatan),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *ClassJournal) withClassJournalStore(ctx context.Context, fn func(classJournalStore) error) error {
	if s.tx == nil {
		return fn(s.q)
	}
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(db.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func seedJournalAttendances(ctx context.Context, store classJournalStore, sessionID, classID pgtype.UUID) error {
	students, err := store.ListActiveStudentsByClassID(ctx, classID)
	if err != nil {
		return err
	}
	for _, st := range students {
		if _, err := store.UpsertJournalAttendance(ctx, db.UpsertJournalAttendanceParams{
			SessionID: sessionID,
			StudentID: st.ID,
			Status:    db.JournalAttendanceStatusHadir,
			Catatan:   "",
		}); err != nil {
			return err
		}
	}
	return nil
}

func normalizeJournalAttendanceStatus(raw string) (db.JournalAttendanceStatus, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "hadir":
		return db.JournalAttendanceStatusHadir, nil
	case "sakit":
		return db.JournalAttendanceStatusSakit, nil
	case "izin":
		return db.JournalAttendanceStatusIzin, nil
	case "alpha":
		return db.JournalAttendanceStatusAlpha, nil
	default:
		return "", fmt.Errorf("status kehadiran tidak valid: %s", raw)
	}
}

func isJournalDuplicateDateError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "uq_journal_session_date") ||
		strings.Contains(message, "duplicate") ||
		strings.Contains(message, "unique")
}

func filterJournalAssignmentsByTeacher(all []db.ListClassSubjectAssignmentsRow, employeeID pgtype.UUID) []db.ListClassSubjectAssignmentsRow {
	out := make([]db.ListClassSubjectAssignmentsRow, 0)
	for _, a := range all {
		if a.TeacherEmployeeID == employeeID {
			out = append(out, a)
		}
	}
	return out
}
