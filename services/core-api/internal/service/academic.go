package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type Academic struct {
	q *db.Queries
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
	return s.q.CreateTimetableSlot(ctx, p)
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
