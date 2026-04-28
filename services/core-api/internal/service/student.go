package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type Student struct {
	q *db.Queries
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
