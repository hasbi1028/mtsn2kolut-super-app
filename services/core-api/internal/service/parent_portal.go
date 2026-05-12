package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type parentPortalStore interface {
	GetPortalParentIDByUserID(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error)
	ListParentChildren(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error)
	GetParentPortalChildAccess(ctx context.Context, arg db.GetParentPortalChildAccessParams) (pgtype.UUID, error)
	GetParentPortalChildProfile(ctx context.Context, arg db.GetParentPortalChildProfileParams) (db.GetParentPortalChildProfileRow, error)
	ListParentPortalChildTimetable(ctx context.Context, arg db.ListParentPortalChildTimetableParams) ([]db.ListParentPortalChildTimetableRow, error)
	ListParentPortalChildExamSessions(ctx context.Context, arg db.ListParentPortalChildExamSessionsParams) ([]db.ListParentPortalChildExamSessionsRow, error)
}

type ParentPortal struct {
	q parentPortalStore
}

func NewParentPortal(q *db.Queries) *ParentPortal {
	return &ParentPortal{q: q}
}

func (s *ParentPortal) Children(ctx context.Context, userID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	parentID, err := s.parentIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.ChildrenByParentID(ctx, parentID)
}

func (s *ParentPortal) ChildrenByParentID(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	if !parentID.Valid {
		return nil, domain.ErrBadRequest
	}
	return s.q.ListParentChildren(ctx, parentID)
}

func (s *ParentPortal) ChildProfile(ctx context.Context, userID, studentID pgtype.UUID) (db.GetParentPortalChildProfileRow, error) {
	parentID, err := s.parentIDForUser(ctx, userID)
	if err != nil {
		return db.GetParentPortalChildProfileRow{}, err
	}
	return s.ChildProfileByParentID(ctx, parentID, studentID)
}

func (s *ParentPortal) ChildProfileByParentID(ctx context.Context, parentID, studentID pgtype.UUID) (db.GetParentPortalChildProfileRow, error) {
	row, err := s.q.GetParentPortalChildProfile(ctx, db.GetParentPortalChildProfileParams{
		ParentID:  parentID,
		StudentID: studentID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetParentPortalChildProfileRow{}, domain.ErrForbidden
	}
	return row, err
}

func (s *ParentPortal) ChildSchedule(ctx context.Context, userID, studentID pgtype.UUID) ([]db.ListParentPortalChildTimetableRow, error) {
	parentID, err := s.parentIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.ChildScheduleByParentID(ctx, parentID, studentID)
}

func (s *ParentPortal) ChildScheduleByParentID(ctx context.Context, parentID, studentID pgtype.UUID) ([]db.ListParentPortalChildTimetableRow, error) {
	if err := s.ensureChildAccess(ctx, parentID, studentID); err != nil {
		return nil, err
	}
	return s.q.ListParentPortalChildTimetable(ctx, db.ListParentPortalChildTimetableParams{
		ParentID:  parentID,
		StudentID: studentID,
	})
}

func (s *ParentPortal) ChildResults(ctx context.Context, userID, studentID pgtype.UUID) ([]db.ListParentPortalChildExamSessionsRow, error) {
	parentID, err := s.parentIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.ChildResultsByParentID(ctx, parentID, studentID)
}

func (s *ParentPortal) ChildResultsByParentID(ctx context.Context, parentID, studentID pgtype.UUID) ([]db.ListParentPortalChildExamSessionsRow, error) {
	if err := s.ensureChildAccess(ctx, parentID, studentID); err != nil {
		return nil, err
	}
	return s.q.ListParentPortalChildExamSessions(ctx, db.ListParentPortalChildExamSessionsParams{
		ParentID:  parentID,
		StudentID: studentID,
	})
}

func (s *ParentPortal) parentIDForUser(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error) {
	if !userID.Valid {
		return pgtype.UUID{}, domain.ErrUnauthorized
	}
	parentID, err := s.q.GetPortalParentIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgtype.UUID{}, domain.ErrForbidden
		}
		return pgtype.UUID{}, err
	}
	if !parentID.Valid {
		return pgtype.UUID{}, domain.ErrForbidden
	}
	return parentID, nil
}

func (s *ParentPortal) ensureChildAccess(ctx context.Context, parentID, studentID pgtype.UUID) error {
	if !parentID.Valid || !studentID.Valid {
		return domain.ErrForbidden
	}
	_, err := s.q.GetParentPortalChildAccess(ctx, db.GetParentPortalChildAccessParams{
		ParentID:  parentID,
		StudentID: studentID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrForbidden
	}
	return err
}

var _ parentPortalStore = (*db.Queries)(nil)
