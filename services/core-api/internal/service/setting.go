package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type settingStore interface {
	ListSettings(ctx context.Context) ([]db.AppSetting, error)
	GetSetting(ctx context.Context, key string) (db.AppSetting, error)
	UpsertSetting(ctx context.Context, arg db.UpsertSettingParams) error
}

type Setting struct {
	q settingStore
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

func (s *Setting) SeedDefaults(ctx context.Context) error {
	defaults := map[string]string{
		"default_max_attempts": "3",
		"max_concurrent":       "5",
		"headless":             "true",
		"scheduler_last_error": "",
	}

	for key, value := range defaults {
		_, err := s.q.GetSetting(ctx, key)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: key, Value: value}); err != nil {
			return err
		}
	}

	return nil
}
