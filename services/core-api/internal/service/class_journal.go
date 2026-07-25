package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type ClassJournal struct {
	q *db.Queries
}

func NewClassJournalService(q *db.Queries) *ClassJournal {
	return &ClassJournal{q: q}
}

// ─── Types ───

type JournalSessionOpenResult struct {
	Session       db.GetJournalSessionRow
	TimetableSlot db.GetRombelTimetableSlotRow
	Created       bool
}

// ─── Helpers ───

type CreateSessionResult struct {
	ID          string `json:"id"`
	PertemuanKe int32  `json:"pertemuan_ke"`
}

// ─── Overview ───

type JournalOverviewItem struct {
	ID              string `json:"id"`
	AssignmentID    string `json:"assignment_id"`
	Tanggal         string `json:"tanggal"`
	PertemuanKe     int32  `json:"pertemuan_ke"`
	Materi          string `json:"materi"`
	Kegiatan        string `json:"kegiatan"`
	Catatan         string `json:"catatan"`
	GuruHadir       bool   `json:"guru_hadir"`
	ClassName       string `json:"class_name"`
	ClassCode       string `json:"class_code"`
	SubjectName     string `json:"subject_name"`
	TeacherName     string `json:"teacher_name"`
}

func (s *ClassJournal) Overview(ctx context.Context, assignmentID string) ([]JournalOverviewItem, error) {
	rows, err := s.q.ListJournalSessions(ctx, pgUUID(assignmentID))
	if err != nil {
		return nil, fmt.Errorf("list journal sessions: %w", err)
	}
	items := make([]JournalOverviewItem, len(rows))
	for i, r := range rows {
		items[i] = JournalOverviewItem{
			ID:           pgUUIDString(r.ID),
			AssignmentID: pgUUIDString(r.AssignmentID),
			Tanggal:      r.Tanggal.Time.Format("2006-01-02"),
			PertemuanKe:  r.PertemuanKe,
			Materi:       r.Materi, Kegiatan: r.Kegiatan, Catatan: r.Catatan,
			GuruHadir: r.GuruHadir, ClassName: r.ClassName,
			ClassCode: r.ClassCode, SubjectName: r.SubjectName,
			TeacherName: r.TeacherName,
		}
	}
	return items, nil
}

// ─── Create Session ───

type CreateSessionRequest struct {
	AssignmentID    string `json:"assignment_id"`
	TimetableSlotID string `json:"timetable_slot_id,omitempty"`
	Tanggal         string `json:"tanggal"`
	Materi          string `json:"materi"`
	Kegiatan        string `json:"kegiatan"`
	Catatan         string `json:"catatan"`
	GuruHadir       bool   `json:"guru_hadir"`
}

func (s *ClassJournal) CreateSession(ctx context.Context, req CreateSessionRequest) (*CreateSessionResult, error) {
	tanggal, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		return nil, fmt.Errorf("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}

	// Check if session already exists for this assignment+date
	if req.TimetableSlotID == "" {
		existingID, err := s.q.GetJournalSessionIDByAssignmentDate(ctx, db.GetJournalSessionIDByAssignmentDateParams{
			AssignmentID: pgUUID(req.AssignmentID),
			Tanggal:      pgtype.Date{Time: tanggal, Valid: true},
		})
		if err == nil {
			// Session already exists, return existing ID
			return &CreateSessionResult{ID: pgUUIDString(existingID)}, nil
		}
	} else {
		existingID, err := s.q.GetJournalSessionIDByTimetableSlotDate(ctx, db.GetJournalSessionIDByTimetableSlotDateParams{
			TimetableSlotID: pgUUID(req.TimetableSlotID),
			Tanggal:         pgtype.Date{Time: tanggal, Valid: true},
		})
		if err == nil {
			return &CreateSessionResult{ID: pgUUIDString(existingID)}, nil
		}
	}

	// Get next pertemuan_ke
	next, err := s.q.NextJournalMeetingNumber(ctx, pgUUID(req.AssignmentID))
	if err != nil {
		return nil, fmt.Errorf("next meeting: %w", err)
	}

	var slotID pgtype.UUID
	if req.TimetableSlotID != "" {
		slotID = pgUUID(req.TimetableSlotID)
	}

	session, err := s.q.CreateJournalSession(ctx, db.CreateJournalSessionParams{
		AssignmentID:    pgUUID(req.AssignmentID),
		TimetableSlotID: slotID,
		Tanggal:         pgtype.Date{Time: tanggal, Valid: true},
		PertemuanKe:     next,
		Materi:          req.Materi,
		Kegiatan:        req.Kegiatan,
		Catatan:         req.Catatan,
		GuruHadir:       req.GuruHadir,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &CreateSessionResult{ID: pgUUIDString(session.ID), PertemuanKe: next}, nil
}

// ─── Open From Timetable Slot ───

func (s *ClassJournal) OpenFromTimetableSlot(ctx context.Context, classID, slotID, tanggal, materi, catatan string, guruHadir bool) (*JournalSessionOpenResult, error) {
	// Parse tanggal
	tgl, err := time.Parse("2006-01-02", tanggal)
	if err != nil {
		return nil, fmt.Errorf("format tanggal tidak valid: %w", err)
	}

	// Get or create the session
	// First try to find existing session for this slot+date
	existingID, err := s.q.GetJournalSessionIDByTimetableSlotDate(ctx, db.GetJournalSessionIDByTimetableSlotDateParams{
		TimetableSlotID: pgUUID(slotID),
		Tanggal:         pgtype.Date{Time: tgl, Valid: true},
	})
	created := false
	var session db.GetJournalSessionRow

	if err != nil {
		// Not found — create new
		// Get the timetable slot to find assignment
		slotRow, err := s.q.GetRombelTimetableSlot(ctx, db.GetRombelTimetableSlotParams{
			ClassID: pgUUID(classID),
			ID:      pgUUID(slotID),
		})
		if err != nil {
			return nil, fmt.Errorf("get timetable slot: %w", err)
		}

		next, err := s.q.NextJournalMeetingNumber(ctx, slotRow.AssignmentID)
		if err != nil {
			return nil, fmt.Errorf("next meeting: %w", err)
		}

		newSession, err := s.q.CreateJournalSession(ctx, db.CreateJournalSessionParams{
			AssignmentID:    slotRow.AssignmentID,
			TimetableSlotID: pgUUID(slotID),
			Tanggal:         pgtype.Date{Time: tgl, Valid: true},
			PertemuanKe:     next,
			Materi:          materi,
			Catatan:         catatan,
			GuruHadir:       guruHadir,
		})
		if err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}

		// Fetch full session data
		session, err = s.q.GetJournalSession(ctx, newSession.ID)
		if err != nil {
			return nil, fmt.Errorf("get session: %w", err)
		}
		created = true
	} else {
		session, err = s.q.GetJournalSession(ctx, existingID)
		if err != nil {
			return nil, fmt.Errorf("get session: %w", err)
		}
	}

	return &JournalSessionOpenResult{
		Session: session,
		Created: created,
	}, nil
}

// ─── Bulk Upsert Attendances ───

type AttendanceRecord struct {
	StudentID string `json:"student_id"`
	Status    string `json:"status"` // hadir, sakit, izin, alpha
	Catatan   string `json:"catatan"`
}

type BulkAttendancesRequest struct {
	SessionID string             `json:"session_id"`
	Records   []AttendanceRecord `json:"records"`
}

func (s *ClassJournal) BulkUpsertAttendances(ctx context.Context, sessionID string, records []AttendanceRecord) error {
	for _, rec := range records {
		_, err := s.q.UpsertJournalAttendance(ctx, db.UpsertJournalAttendanceParams{
			SessionID: pgUUID(sessionID),
			StudentID: pgUUID(rec.StudentID),
			Status:    db.JournalAttendanceStatus(rec.Status),
			Catatan:   rec.Catatan,
		})
		if err != nil {
			return fmt.Errorf("upsert attendance for student %s: %w", rec.StudentID, err)
		}
	}
	return nil
}

// ─── Attendance List ───

type AttendanceItem struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	StudentID string `json:"student_id"`
	Status    string `json:"status"`
	Catatan   string `json:"catatan"`
	NIS       string `json:"nis"`
	NISN      string `json:"nisn"`
	Nama      string `json:"nama"`
	Gender    string `json:"gender"`
}

func (s *ClassJournal) ListAttendances(ctx context.Context, sessionID string) ([]AttendanceItem, error) {
	rows, err := s.q.ListJournalAttendances(ctx, pgUUID(sessionID))
	if err != nil {
		return nil, fmt.Errorf("list attendances: %w", err)
	}
	items := make([]AttendanceItem, len(rows))
	for i, r := range rows {
		items[i] = AttendanceItem{
			ID: pgUUIDString(r.ID), SessionID: pgUUIDString(r.SessionID),
			StudentID: pgUUIDString(r.StudentID), Status: string(r.Status),
			Catatan: r.Catatan, NIS: r.Nis, NISN: r.Nisn,
			Nama: r.Nama, Gender: string(r.Gender),
		}
	}
	return items, nil
}

// ─── Attendance Summary ───

type AttendanceSummaryItem struct {
	StudentID     string `json:"student_id"`
	NIS           string `json:"nis"`
	NISN          string `json:"nisn"`
	Nama          string `json:"nama"`
	TotalPertemuan int32 `json:"total_pertemuan"`
	Hadir         int32  `json:"hadir"`
	Sakit         int32  `json:"sakit"`
	Izin          int32  `json:"izin"`
	Alpha         int32  `json:"alpha"`
}

func (s *ClassJournal) AttendanceSummary(ctx context.Context, assignmentID string) ([]AttendanceSummaryItem, error) {
	rows, err := s.q.ListJournalAttendanceSummary(ctx, pgUUID(assignmentID))
	if err != nil {
		return nil, fmt.Errorf("attendance summary: %w", err)
	}
	items := make([]AttendanceSummaryItem, len(rows))
	for i, r := range rows {
		items[i] = AttendanceSummaryItem{
			StudentID: pgUUIDString(r.StudentID), NIS: r.Nis, NISN: r.Nisn,
			Nama: r.Nama, TotalPertemuan: r.TotalPertemuan,
			Hadir: r.Hadir, Sakit: r.Sakit, Izin: r.Izin, Alpha: r.Alpha,
		}
	}
	return items, nil
}

// ─── Delete Session ───

func (s *ClassJournal) DeleteSession(ctx context.Context, id string) error {
	err := s.q.DeleteJournalSession(ctx, pgUUID(id))
	if err != nil {
		return fmt.Errorf("delete journal session: %w", err)
	}
	return nil
}
