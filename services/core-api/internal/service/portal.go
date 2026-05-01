package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type Portal struct {
	q *db.Queries
}

func NewPortal(q *db.Queries) *Portal { return &Portal{q: q} }

func (s *Portal) StudentOverview(ctx context.Context, studentID pgtype.UUID) (db.GetStudentByIDRow, []db.ListStudentParentsRow, []db.ListStudentExamSessionsRow, error) {
	student, err := s.q.GetStudentByID(ctx, studentID)
	if err != nil {
		return db.GetStudentByIDRow{}, nil, nil, err
	}
	parents, err := s.q.ListStudentParents(ctx, studentID)
	if err != nil {
		return db.GetStudentByIDRow{}, nil, nil, err
	}
	sessions, err := s.q.ListStudentExamSessions(ctx, studentID)
	if err != nil {
		return db.GetStudentByIDRow{}, nil, nil, err
	}
	return student, parents, sessions, nil
}

func (s *Portal) StudentTimetable(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	return s.q.ListStudentTimetable(ctx, studentID)
}

func (s *Portal) TeacherTimetable(ctx context.Context, employeeID pgtype.UUID) ([]db.ListTeacherTimetableRow, error) {
	return s.q.ListTeacherTimetable(ctx, employeeID)
}

func (s *Portal) ParentOverview(ctx context.Context, parentID pgtype.UUID) (db.Parent, []db.ListParentChildrenRow, error) {
	parent, err := s.q.GetParent(ctx, parentID)
	if err != nil {
		return db.Parent{}, nil, err
	}
	children, err := s.q.ListParentChildren(ctx, parentID)
	if err != nil {
		return db.Parent{}, nil, err
	}
	return parent, children, nil
}
