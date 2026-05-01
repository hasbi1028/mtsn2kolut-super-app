package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type notificationStore interface {
	ListAppNotifications(ctx context.Context, arg db.ListAppNotificationsParams) ([]db.AppNotification, error)
	CountUnreadAppNotifications(ctx context.Context, userID pgtype.UUID) (int64, error)
	MarkAppNotificationRead(ctx context.Context, arg db.MarkAppNotificationReadParams) (db.AppNotification, error)
	MarkAllAppNotificationsRead(ctx context.Context, userID pgtype.UUID) (int64, error)
}

type Notification struct {
	q notificationStore
}

func NewNotification(q *db.Queries) *Notification {
	return &Notification{q: q}
}

func (s *Notification) List(ctx context.Context, userID pgtype.UUID, unreadOnly bool, limit, offset int32) ([]db.AppNotification, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.q.ListAppNotifications(ctx, db.ListAppNotificationsParams{
		UserID:      userID,
		UnreadOnly:  unreadOnly,
		LimitCount:  limit,
		OffsetCount: offset,
	})
}

func (s *Notification) CountUnread(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return s.q.CountUnreadAppNotifications(ctx, userID)
}

func (s *Notification) MarkRead(ctx context.Context, userID, notificationID pgtype.UUID) (db.AppNotification, error) {
	row, err := s.q.MarkAppNotificationRead(ctx, db.MarkAppNotificationReadParams{
		ID:     notificationID,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.AppNotification{}, domain.ErrNotFound
	}
	return row, err
}

func (s *Notification) MarkAllRead(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return s.q.MarkAllAppNotificationsRead(ctx, userID)
}
