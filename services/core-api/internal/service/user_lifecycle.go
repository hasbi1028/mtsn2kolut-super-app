package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type userLifecycleStore interface {
	UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) error
	RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error)
}

type userLifecycleTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type UserLifecycle struct {
	q  userLifecycleStore
	tx userLifecycleTxStarter
}

func NewUserLifecycle(q *db.Queries) *UserLifecycle {
	return &UserLifecycle{q: q}
}

func NewUserLifecycleWithPool(pool *pgxpool.Pool) *UserLifecycle {
	return &UserLifecycle{q: db.New(pool), tx: pool}
}

func (s *UserLifecycle) DeleteAsDeactivate(ctx context.Context, id pgtype.UUID) error {
	return s.setActive(ctx, id, false)
}

func (s *UserLifecycle) UpdateStatus(ctx context.Context, id pgtype.UUID, isActive bool) error {
	return s.setActive(ctx, id, isActive)
}

func (s *UserLifecycle) setActive(ctx context.Context, id pgtype.UUID, isActive bool) error {
	apply := func(store userLifecycleStore) error {
		if err := store.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
			ID:       id,
			IsActive: isActive,
		}); err != nil {
			return err
		}
		if !isActive {
			_, err := store.RevokeAllAuthSessionsForUser(ctx, id)
			return err
		}
		return nil
	}

	if s.tx == nil {
		return apply(s.q)
	}

	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := apply(db.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
