package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type rombelStore interface {
	ListRombels(ctx context.Context) ([]db.ListRombelsRow, error)
	GetRombelDetail(ctx context.Context, id pgtype.UUID) (db.GetRombelDetailRow, error)
	ListStudentsByClassWithParents(ctx context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error)
	ListRombelSubjectAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error)
	GetRombelSubjectAssignment(ctx context.Context, arg db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error)
	CreateRombelSubjectAssignment(ctx context.Context, arg db.CreateRombelSubjectAssignmentParams) (db.CreateRombelSubjectAssignmentRow, error)
	UpdateRombelSubjectAssignment(ctx context.Context, arg db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error)
	CountRombelSubjectAssignmentDependents(ctx context.Context, arg db.CountRombelSubjectAssignmentDependentsParams) (db.CountRombelSubjectAssignmentDependentsRow, error)
	DeleteRombelSubjectAssignment(ctx context.Context, arg db.DeleteRombelSubjectAssignmentParams) (int64, error)
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

func (s *Rombel) GetSubjectAssignment(ctx context.Context, arg db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error) {
	return s.q.GetRombelSubjectAssignment(ctx, arg)
}

func (s *Rombel) CreateSubjectAssignment(ctx context.Context, arg db.CreateRombelSubjectAssignmentParams) (db.CreateRombelSubjectAssignmentRow, error) {
	return s.q.CreateRombelSubjectAssignment(ctx, arg)
}

func (s *Rombel) UpdateSubjectAssignment(ctx context.Context, arg db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error) {
	current, err := s.q.GetRombelSubjectAssignment(ctx, db.GetRombelSubjectAssignmentParams{
		ClassID: arg.ClassID,
		ID:      arg.ID,
	})
	if err != nil {
		return db.UpdateRombelSubjectAssignmentRow{}, err
	}
	if !sameRombelUUID(current.SubjectID, arg.SubjectID) {
		counts, err := s.q.CountRombelSubjectAssignmentDependents(ctx, db.CountRombelSubjectAssignmentDependentsParams{
			ClassID: arg.ClassID,
			ID:      arg.ID,
		})
		if err != nil {
			return db.UpdateRombelSubjectAssignmentRow{}, err
		}
		if counts.TotalTimetableSlots > 0 || counts.TotalJournalSessions > 0 || counts.TotalGradeComponents > 0 || counts.TotalGradeFinalizations > 0 {
			return db.UpdateRombelSubjectAssignmentRow{}, fmt.Errorf("%w: mata pelajaran tidak dapat diganti karena penugasan sudah dipakai oleh jadwal, jurnal, atau nilai", domain.ErrConflict)
		}
	}
	return s.q.UpdateRombelSubjectAssignment(ctx, arg)
}

func (s *Rombel) DeleteSubjectAssignment(ctx context.Context, arg db.DeleteRombelSubjectAssignmentParams) error {
	counts, err := s.q.CountRombelSubjectAssignmentDependents(ctx, db.CountRombelSubjectAssignmentDependentsParams{
		ClassID: arg.ClassID,
		ID:      arg.ID,
	})
	if err != nil {
		return err
	}
	if counts.TotalTimetableSlots > 0 || counts.TotalJournalSessions > 0 || counts.TotalGradeComponents > 0 || counts.TotalGradeFinalizations > 0 {
		return fmt.Errorf("%w: penugasan guru mapel masih dipakai oleh jadwal, jurnal, atau nilai", domain.ErrConflict)
	}
	rows, err := s.q.DeleteRombelSubjectAssignment(ctx, arg)
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func sameRombelUUID(a, b pgtype.UUID) bool {
	if a.Valid != b.Valid {
		return false
	}
	return a.Bytes == b.Bytes
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
