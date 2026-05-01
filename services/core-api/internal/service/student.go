package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type studentStore interface {
	ListStudents(ctx context.Context) ([]db.ListStudentsRow, error)
	ListStudentsByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListStudentsByTeacherRow, error)
	CreateStudent(ctx context.Context, arg db.CreateStudentParams) (db.Student, error)
	UpdateStudent(ctx context.Context, arg db.UpdateStudentParams) (db.Student, error)
	DeleteStudent(ctx context.Context, id pgtype.UUID) error
	UpdateStudentStatus(ctx context.Context, arg db.UpdateStudentStatusParams) error
	GetStudentByID(ctx context.Context, id pgtype.UUID) (db.GetStudentByIDRow, error)
	UpdateStudentLifecycle(ctx context.Context, arg db.UpdateStudentLifecycleParams) error
	ListUsersByStudentID(ctx context.Context, studentID pgtype.UUID) ([]db.ListUsersByStudentIDRow, error)
	UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) error
}

type Student struct {
	q studentStore
}

func NewStudent(q *db.Queries) *Student { return &Student{q: q} }

func (s *Student) List(ctx context.Context) ([]db.ListStudentsRow, error) {
	return s.q.ListStudents(ctx)
}

func (s *Student) ListByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListStudentsByTeacherRow, error) {
	return s.q.ListStudentsByTeacher(ctx, teacherEmployeeID)
}

func (s *Student) Create(ctx context.Context, p db.CreateStudentParams) (db.Student, error) {
	return s.q.CreateStudent(ctx, p)
}

func (s *Student) Update(ctx context.Context, p db.UpdateStudentParams) (db.Student, error) {
	return s.q.UpdateStudent(ctx, p)
}

func (s *Student) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteStudent(ctx, id)
}

func (s *Student) UpdateStatus(ctx context.Context, id pgtype.UUID, status db.StudentStatusEnum) error {
	return s.q.UpdateStudentStatus(ctx, db.UpdateStudentStatusParams{
		ID:     id,
		Status: status,
	})
}

func (s *Student) Get(ctx context.Context, id pgtype.UUID) (db.GetStudentByIDRow, error) {
	return s.q.GetStudentByID(ctx, id)
}

func (s *Student) UpdateLifecycle(ctx context.Context, id pgtype.UUID, status db.StudentStatusEnum) error {
	isActive := status == db.StudentStatusEnumActive || status == db.StudentStatusEnumProspective
	if err := s.q.UpdateStudentLifecycle(ctx, db.UpdateStudentLifecycleParams{
		ID:       id,
		Status:   status,
		IsActive: isActive,
	}); err != nil {
		return err
	}
	users, err := s.q.ListUsersByStudentID(ctx, id)
	if err != nil {
		return err
	}
	for _, user := range users {
		if err := s.q.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
			ID:       user.ID,
			IsActive: isActive,
		}); err != nil {
			return err
		}
	}
	return nil
}
