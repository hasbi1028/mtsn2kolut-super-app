package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type Schedule struct {
	q *db.Queries
}

func NewSchedule(q *db.Queries) *Schedule { return &Schedule{q: q} }

func (s *Schedule) List(ctx context.Context) ([]db.Schedule, error) {
	return s.q.ListSchedules(ctx)
}

func (s *Schedule) Get(ctx context.Context, id pgtype.UUID) (db.Schedule, error) {
	return s.q.GetSchedule(ctx, id)
}

func (s *Schedule) Create(ctx context.Context, p db.CreateScheduleParams) (db.Schedule, error) {
	return s.q.CreateSchedule(ctx, p)
}

func (s *Schedule) UpdateByID(ctx context.Context, p db.UpdateScheduleByIDParams) (db.Schedule, error) {
	return s.q.UpdateScheduleByID(ctx, p)
}

func (s *Schedule) DeleteByID(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteScheduleByID(ctx, id)
}
