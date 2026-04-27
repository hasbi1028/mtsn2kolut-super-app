package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type Employee struct {
	q *db.Queries
}

func NewEmployee(q *db.Queries) *Employee { return &Employee{q: q} }

func (s *Employee) List(ctx context.Context) ([]db.Employee, error) {
	return s.q.ListEmployees(ctx)
}

func (s *Employee) ListActive(ctx context.Context) ([]db.Employee, error) {
	return s.q.ListActiveEmployees(ctx)
}

func (s *Employee) Get(ctx context.Context, id pgtype.UUID) (db.Employee, error) {
	return s.q.GetEmployee(ctx, id)
}

func (s *Employee) Create(ctx context.Context, p db.CreateEmployeeParams) (db.Employee, error) {
	return s.q.CreateEmployee(ctx, p)
}

func (s *Employee) Update(ctx context.Context, p db.UpdateEmployeeParams) (db.Employee, error) {
	return s.q.UpdateEmployee(ctx, p)
}

func (s *Employee) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteEmployee(ctx, id)
}

func (s *Employee) ListWithStatus(ctx context.Context) ([]db.ListEmployeesWithStatusRow, error) {
	return s.q.ListEmployeesWithStatus(ctx)
}
