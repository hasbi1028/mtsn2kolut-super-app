package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type PusakaSchedule struct {
	q *db.Queries
}

func NewPusakaSchedule(q *db.Queries) *PusakaSchedule { return &PusakaSchedule{q: q} }

func (s *PusakaSchedule) List(ctx context.Context) ([]db.Schedule, error) {
	return s.q.ListSchedules(ctx)
}

func (s *PusakaSchedule) Get(ctx context.Context, id pgtype.UUID) (db.Schedule, error) {
	return s.q.GetSchedule(ctx, id)
}

func (s *PusakaSchedule) Create(ctx context.Context, p db.CreateScheduleParams) (db.Schedule, error) {
	return s.q.CreateSchedule(ctx, p)
}

func (s *PusakaSchedule) UpdateByID(ctx context.Context, p db.UpdateScheduleByIDParams) (db.Schedule, error) {
	return s.q.UpdateScheduleByID(ctx, p)
}

func (s *PusakaSchedule) DeleteByID(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteScheduleByID(ctx, id)
}
