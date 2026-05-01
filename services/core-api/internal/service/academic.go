package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type academicStore interface {
	ListAcademicYears(ctx context.Context) ([]db.AcademicYear, error)
	ListSchoolClasses(ctx context.Context) ([]db.ListSchoolClassesRow, error)
	ListSubjects(ctx context.Context) ([]db.Subject, error)
	ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error)
	ListTimetableSlots(ctx context.Context) ([]db.ListTimetableSlotsRow, error)
	GetAcademicStats(ctx context.Context) (db.GetAcademicStatsRow, error)
	CreateAcademicYear(ctx context.Context, arg db.CreateAcademicYearParams) (db.AcademicYear, error)
	CreateSchoolClass(ctx context.Context, arg db.CreateSchoolClassParams) (db.SchoolClass, error)
	CreateSubject(ctx context.Context, arg db.CreateSubjectParams) (db.Subject, error)
	CreateClassSubjectAssignment(ctx context.Context, arg db.CreateClassSubjectAssignmentParams) (db.ClassSubjectAssignment, error)
	CreateTimetableSlot(ctx context.Context, arg db.CreateTimetableSlotParams) (db.TimetableSlot, error)
	GetTimetableSlot(ctx context.Context, id pgtype.UUID) (db.TimetableSlot, error)
	UpdateTimetableSlot(ctx context.Context, arg db.UpdateTimetableSlotParams) (db.TimetableSlot, error)
	DeleteAcademicYear(ctx context.Context, id pgtype.UUID) error
	DeleteSchoolClass(ctx context.Context, id pgtype.UUID) error
	DeleteSubject(ctx context.Context, id pgtype.UUID) error
	DeleteClassSubjectAssignment(ctx context.Context, id pgtype.UUID) error
	DeleteTimetableSlot(ctx context.Context, id pgtype.UUID) error
	GetClassSubjectAssignment(ctx context.Context, id pgtype.UUID) (db.GetClassSubjectAssignmentRow, error)
	CountTimetableConflicts(ctx context.Context, arg db.CountTimetableConflictsParams) (int32, error)
	CountTimetableRoomConflicts(ctx context.Context, arg db.CountTimetableRoomConflictsParams) (int32, error)
}

type Academic struct {
	q academicStore
}

func NewAcademic(q *db.Queries) *Academic { return &Academic{q: q} }

func (s *Academic) ListYears(ctx context.Context) ([]db.AcademicYear, error) {
	return s.q.ListAcademicYears(ctx)
}

func (s *Academic) ListClasses(ctx context.Context) ([]db.ListSchoolClassesRow, error) {
	return s.q.ListSchoolClasses(ctx)
}

func (s *Academic) ListSubjects(ctx context.Context) ([]db.Subject, error) {
	return s.q.ListSubjects(ctx)
}

func (s *Academic) ListAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error) {
	return s.q.ListClassSubjectAssignments(ctx)
}

func (s *Academic) ListTimetableSlots(ctx context.Context) ([]db.ListTimetableSlotsRow, error) {
	return s.q.ListTimetableSlots(ctx)
}

func (s *Academic) GetStats(ctx context.Context) (db.GetAcademicStatsRow, error) {
	return s.q.GetAcademicStats(ctx)
}

func (s *Academic) CreateYear(ctx context.Context, p db.CreateAcademicYearParams) (db.AcademicYear, error) {
	return s.q.CreateAcademicYear(ctx, p)
}

func (s *Academic) CreateClass(ctx context.Context, p db.CreateSchoolClassParams) (db.SchoolClass, error) {
	return s.q.CreateSchoolClass(ctx, p)
}

func (s *Academic) CreateSubject(ctx context.Context, p db.CreateSubjectParams) (db.Subject, error) {
	return s.q.CreateSubject(ctx, p)
}

func (s *Academic) CreateAssignment(ctx context.Context, p db.CreateClassSubjectAssignmentParams) (db.ClassSubjectAssignment, error) {
	return s.q.CreateClassSubjectAssignment(ctx, p)
}

func (s *Academic) CreateTimetableSlot(ctx context.Context, p db.CreateTimetableSlotParams) (db.TimetableSlot, error) {
	if err := s.ensureTimetableSlotAvailable(ctx, p.AssignmentID, p.DayOfWeek, p.StartTime, p.EndTime, p.RoomLabel, pgtype.UUID{}); err != nil {
		return db.TimetableSlot{}, err
	}
	return s.q.CreateTimetableSlot(ctx, p)
}

func (s *Academic) UpdateTimetableSlot(ctx context.Context, p db.UpdateTimetableSlotParams) (db.TimetableSlot, error) {
	if _, err := s.q.GetTimetableSlot(ctx, p.ID); err != nil {
		return db.TimetableSlot{}, err
	}
	if err := s.ensureTimetableSlotAvailable(ctx, p.AssignmentID, p.DayOfWeek, p.StartTime, p.EndTime, p.RoomLabel, p.ID); err != nil {
		return db.TimetableSlot{}, err
	}
	return s.q.UpdateTimetableSlot(ctx, p)
}

func (s *Academic) DeleteYear(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteAcademicYear(ctx, id)
}

func (s *Academic) DeleteClass(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteSchoolClass(ctx, id)
}

func (s *Academic) DeleteSubject(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteSubject(ctx, id)
}

func (s *Academic) DeleteAssignment(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteClassSubjectAssignment(ctx, id)
}

func (s *Academic) DeleteTimetableSlot(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteTimetableSlot(ctx, id)
}

func (s *Academic) ensureTimetableSlotAvailable(ctx context.Context, assignmentID pgtype.UUID, dayOfWeek int16, startTime, endTime pgtype.Time, roomLabel string, excludeSlotID pgtype.UUID) error {
	assignment, err := s.q.GetClassSubjectAssignment(ctx, assignmentID)
	if err != nil {
		return fmt.Errorf("assignment tidak ditemukan")
	}
	conflicts, err := s.q.CountTimetableConflicts(ctx, db.CountTimetableConflictsParams{
		DayOfWeek:         dayOfWeek,
		StartTime:         startTime,
		EndTime:           endTime,
		ClassID:           assignment.ClassID,
		TeacherEmployeeID: assignment.TeacherEmployeeID,
		ExcludeSlotID:     excludeSlotID,
	})
	if err != nil {
		return err
	}
	if conflicts > 0 {
		return fmt.Errorf("slot bentrok dengan jadwal kelas atau guru pada waktu yang sama")
	}
	roomConflicts, err := s.q.CountTimetableRoomConflicts(ctx, db.CountTimetableRoomConflictsParams{
		DayOfWeek:     dayOfWeek,
		StartTime:     startTime,
		EndTime:       endTime,
		RoomLabel:     strings.TrimSpace(roomLabel),
		ExcludeSlotID: excludeSlotID,
	})
	if err != nil {
		return err
	}
	if roomConflicts > 0 {
		return fmt.Errorf("slot bentrok dengan penggunaan ruang pada waktu yang sama")
	}
	return nil
}

func ParseAcademicTimeInput(value string) (pgtype.Time, error) {
	var out pgtype.Time
	raw := strings.TrimSpace(value)
	if raw == "" {
		return out, fmt.Errorf("waktu wajib diisi")
	}
	if len(raw) == 5 {
		raw += ":00"
	}
	if err := out.Scan(raw); err != nil {
		return out, fmt.Errorf("format waktu tidak valid")
	}
	return out, nil
}
