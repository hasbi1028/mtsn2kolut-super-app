package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type studentPortalStore interface {
	GetPortalStudentIDByUserID(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error)
	GetStudentByID(ctx context.Context, id pgtype.UUID) (db.GetStudentByIDRow, error)
	ListStudentTimetable(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error)
	ListStudentExamSessions(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error)
}

type StudentPortal struct {
	q studentPortalStore
}

func NewStudentPortal(q *db.Queries) *StudentPortal {
	return &StudentPortal{q: q}
}

func (s *StudentPortal) Profile(ctx context.Context, userID pgtype.UUID) (db.GetStudentByIDRow, error) {
	studentID, err := s.studentIDForUser(ctx, userID)
	if err != nil {
		return db.GetStudentByIDRow{}, err
	}
	return s.q.GetStudentByID(ctx, studentID)
}

func (s *StudentPortal) Schedule(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	studentID, err := s.studentIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.q.ListStudentTimetable(ctx, studentID)
}

func (s *StudentPortal) Results(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error) {
	studentID, err := s.studentIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.q.ListStudentExamSessions(ctx, studentID)
}

func (s *StudentPortal) studentIDForUser(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error) {
	if !userID.Valid {
		return pgtype.UUID{}, domain.ErrUnauthorized
	}
	studentID, err := s.q.GetPortalStudentIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgtype.UUID{}, domain.ErrForbidden
		}
		return pgtype.UUID{}, err
	}
	if !studentID.Valid {
		return pgtype.UUID{}, domain.ErrForbidden
	}
	return studentID, nil
}

var _ studentPortalStore = (*db.Queries)(nil)
