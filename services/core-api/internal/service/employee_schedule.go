package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type EmployeeSchedule struct {
	q *db.Queries
}

func NewEmployeeSchedule(q *db.Queries) *EmployeeSchedule { return &EmployeeSchedule{q: q} }

func (s *EmployeeSchedule) List(ctx context.Context, employeeID pgtype.UUID) ([]db.ListEmployeeSchedulesRow, error) {
	return s.q.ListEmployeeSchedules(ctx, employeeID)
}

func (s *EmployeeSchedule) Upsert(ctx context.Context, p db.UpsertEmployeeScheduleParams) (db.EmployeeSchedule, error) {
	return s.q.UpsertEmployeeSchedule(ctx, p)
}

func (s *EmployeeSchedule) Delete(ctx context.Context, id, employeeID pgtype.UUID) error {
	return s.q.DeleteEmployeeSchedule(ctx, db.DeleteEmployeeScheduleParams{
		ID:         id,
		EmployeeID: employeeID,
	})
}
