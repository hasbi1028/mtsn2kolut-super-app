package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type rombelStore interface {
	ListRombels(ctx context.Context) ([]db.ListRombelsRow, error)
	GetRombelDetail(ctx context.Context, id pgtype.UUID) (db.GetRombelDetailRow, error)
	ListStudentsByClassWithParents(ctx context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error)
	ListRombelSubjectAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error)
	ListRombelTimetableSlots(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error)
	ListHomeroomAssignmentsByClass(ctx context.Context, classID pgtype.UUID) ([]db.ListHomeroomAssignmentsByClassRow, error)
	CreateHomeroomAssignment(ctx context.Context, arg db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error)
	UpdateHomeroomAssignment(ctx context.Context, arg db.UpdateHomeroomAssignmentParams) (db.UpdateHomeroomAssignmentRow, error)
	DeleteHomeroomAssignment(ctx context.Context, id pgtype.UUID) error
}

type Rombel struct {
	q rombelStore
}

func NewRombel(q *db.Queries) *Rombel { return &Rombel{q: q} }

func (s *Rombel) List(ctx context.Context) ([]db.ListRombelsRow, error) {
	return s.q.ListRombels(ctx)
}

func (s *Rombel) Get(ctx context.Context, id pgtype.UUID) (db.GetRombelDetailRow, error) {
	return s.q.GetRombelDetail(ctx, id)
}

func (s *Rombel) ListStudentsWithParents(ctx context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error) {
	return s.q.ListStudentsByClassWithParents(ctx, classID)
}

func (s *Rombel) ListSubjectAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error) {
	return s.q.ListRombelSubjectAssignments(ctx, classID)
}

func (s *Rombel) ListTimetableSlots(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error) {
	return s.q.ListRombelTimetableSlots(ctx, classID)
}

func (s *Rombel) ListHomeroomAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListHomeroomAssignmentsByClassRow, error) {
	return s.q.ListHomeroomAssignmentsByClass(ctx, classID)
}

func (s *Rombel) CreateHomeroomAssignment(ctx context.Context, arg db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error) {
	return s.q.CreateHomeroomAssignment(ctx, arg)
}

func (s *Rombel) UpdateHomeroomAssignment(ctx context.Context, arg db.UpdateHomeroomAssignmentParams) (db.UpdateHomeroomAssignmentRow, error) {
	return s.q.UpdateHomeroomAssignment(ctx, arg)
}

func (s *Rombel) DeleteHomeroomAssignment(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteHomeroomAssignment(ctx, id)
}
