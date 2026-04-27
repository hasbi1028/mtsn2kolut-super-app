package service

import (
	"context"

	db "github.com/pusaka/backend/internal/repository/postgres"
)

type Setting struct {
	q *db.Queries
}

func NewSetting(q *db.Queries) *Setting { return &Setting{q: q} }

func (s *Setting) List(ctx context.Context) ([]db.AppSetting, error) {
	return s.q.ListSettings(ctx)
}

func (s *Setting) Get(ctx context.Context, key string) (db.AppSetting, error) {
	return s.q.GetSetting(ctx, key)
}

func (s *Setting) Upsert(ctx context.Context, key, value string) error {
	return s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: key, Value: value})
}
