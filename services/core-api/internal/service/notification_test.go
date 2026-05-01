package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeNotificationStore struct {
	listArg      db.ListAppNotificationsParams
	listRows     []db.AppNotification
	listErr      error
	countUserID  pgtype.UUID
	unreadCount  int64
	markArg      db.MarkAppNotificationReadParams
	markRow      db.AppNotification
	markErr      error
	markAllUser  pgtype.UUID
	markAllCount int64
}

func (f *fakeNotificationStore) ListAppNotifications(ctx context.Context, arg db.ListAppNotificationsParams) ([]db.AppNotification, error) {
	f.listArg = arg
	return f.listRows, f.listErr
}

func (f *fakeNotificationStore) CountUnreadAppNotifications(ctx context.Context, userID pgtype.UUID) (int64, error) {
	f.countUserID = userID
	return f.unreadCount, nil
}

func (f *fakeNotificationStore) MarkAppNotificationRead(ctx context.Context, arg db.MarkAppNotificationReadParams) (db.AppNotification, error) {
	f.markArg = arg
	if f.markErr != nil {
		return db.AppNotification{}, f.markErr
	}
	return f.markRow, nil
}

func (f *fakeNotificationStore) MarkAllAppNotificationsRead(ctx context.Context, userID pgtype.UUID) (int64, error) {
	f.markAllUser = userID
	return f.markAllCount, nil
}

func TestNotificationListClampsPaging(t *testing.T) {
	userID := documentCycleTestUUID(31)
	store := &fakeNotificationStore{}
	svc := &Notification{q: store}

	if _, err := svc.List(context.Background(), userID, true, 500, -5); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.listArg.UserID != userID || !store.listArg.UnreadOnly {
		t.Fatalf("ListAppNotifications() arg = %+v, want user/unread", store.listArg)
	}
	if store.listArg.LimitCount != 100 || store.listArg.OffsetCount != 0 {
		t.Fatalf("ListAppNotifications() paging = limit %d offset %d, want 100/0", store.listArg.LimitCount, store.listArg.OffsetCount)
	}

	if _, err := svc.List(context.Background(), userID, false, 0, 5); err != nil {
		t.Fatalf("List(default limit) error = %v", err)
	}
	if store.listArg.LimitCount != 30 || store.listArg.OffsetCount != 5 || store.listArg.UnreadOnly {
		t.Fatalf("List(default limit) arg = %+v, want limit 30 offset 5 all notifications", store.listArg)
	}
}

func TestNotificationListPropagatesStoreError(t *testing.T) {
	expected := errors.New("list failed")
	store := &fakeNotificationStore{listErr: expected}
	svc := &Notification{q: store}

	_, err := svc.List(context.Background(), documentCycleTestUUID(36), false, 10, 0)
	if !errors.Is(err, expected) {
		t.Fatalf("List(store error) = %v, want %v", err, expected)
	}
}

func TestNotificationMarkReadMapsMissingRow(t *testing.T) {
	store := &fakeNotificationStore{markErr: pgx.ErrNoRows}
	svc := &Notification{q: store}

	_, err := svc.MarkRead(context.Background(), documentCycleTestUUID(32), documentCycleTestUUID(33))
	if err != domain.ErrNotFound {
		t.Fatalf("MarkRead() error = %v, want ErrNotFound", err)
	}
}

func TestNotificationCountAndMarkAllForwardUserID(t *testing.T) {
	userID := documentCycleTestUUID(34)
	store := &fakeNotificationStore{
		unreadCount:  7,
		markAllCount: 5,
	}
	svc := &Notification{q: store}

	count, err := svc.CountUnread(context.Background(), userID)
	if err != nil {
		t.Fatalf("CountUnread() error = %v", err)
	}
	if count != 7 || store.countUserID != userID {
		t.Fatalf("CountUnread() = %d, userID=%v; want 7/%v", count, store.countUserID, userID)
	}
	updated, err := svc.MarkAllRead(context.Background(), userID)
	if err != nil {
		t.Fatalf("MarkAllRead() error = %v", err)
	}
	if updated != 5 || store.markAllUser != userID {
		t.Fatalf("MarkAllRead() = %d, userID=%v; want 5/%v", updated, store.markAllUser, userID)
	}

	notificationID := documentCycleTestUUID(35)
	store.markRow = db.AppNotification{ID: notificationID, UserID: userID}
	row, err := svc.MarkRead(context.Background(), userID, notificationID)
	if err != nil {
		t.Fatalf("MarkRead() error = %v", err)
	}
	if row.ID != notificationID || store.markArg.ID != notificationID || store.markArg.UserID != userID {
		t.Fatalf("MarkRead() row/arg = %+v/%+v, want forwarded ids", row, store.markArg)
	}
}
