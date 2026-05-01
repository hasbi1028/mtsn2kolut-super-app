package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeNotificationService struct {
	*service.Notification

	listUserID     pgtype.UUID
	listUnreadOnly bool
	listLimit      int32
	listOffset     int32
	listRows       []db.AppNotification
	listErr        error
	countUserID    pgtype.UUID
	count          int64
	countErr       error
	markUserID     pgtype.UUID
	markID         pgtype.UUID
	markRow        db.AppNotification
	markErr        error
	markAllUserID  pgtype.UUID
	markAllCount   int64
	markAllErr     error
}

func (f *fakeNotificationService) List(_ context.Context, userID pgtype.UUID, unreadOnly bool, limit, offset int32) ([]db.AppNotification, error) {
	f.listUserID = userID
	f.listUnreadOnly = unreadOnly
	f.listLimit = limit
	f.listOffset = offset
	return f.listRows, f.listErr
}

func (f *fakeNotificationService) CountUnread(_ context.Context, userID pgtype.UUID) (int64, error) {
	f.countUserID = userID
	return f.count, f.countErr
}

func (f *fakeNotificationService) MarkRead(_ context.Context, userID, notificationID pgtype.UUID) (db.AppNotification, error) {
	f.markUserID = userID
	f.markID = notificationID
	return f.markRow, f.markErr
}

func (f *fakeNotificationService) MarkAllRead(_ context.Context, userID pgtype.UUID) (int64, error) {
	f.markAllUserID = userID
	return f.markAllCount, f.markAllErr
}

func notificationTestRow(id, userID pgtype.UUID) db.AppNotification {
	return db.AppNotification{
		ID:         id,
		UserID:     userID,
		Category:   "document_cycle",
		Title:      "Deadline dokumen",
		Body:       "Dokumen mendekati deadline",
		EntityType: "document_cycle_obligation",
		EntityID:   "obligation-1",
		LinkPath:   "/document-cycle",
	}
}

func TestNotificationSuccessHandlersForwardUserScope(t *testing.T) {
	userID := handlerTestUUID(122)
	notificationID := handlerTestUUID(123)
	fake := &fakeNotificationService{
		Notification: &service.Notification{},
		listRows:     []db.AppNotification{notificationTestRow(notificationID, userID)},
		count:        3,
		markRow:      notificationTestRow(notificationID, userID),
		markAllCount: 2,
	}
	h := &Notification{svc: fake}
	reqForUser := func(method, target, body string) *http.Request {
		return withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{"sub": userID.String()})
	}

	rec := httptest.NewRecorder()
	h.List(rec, reqForUser(http.MethodGet, "/api/notifications?unread_only=true&limit=50&offset=5", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listUserID != userID || !fake.listUnreadOnly || fake.listLimit != 50 || fake.listOffset != 5 {
		t.Fatalf("List args = (%v, %v, %d, %d), want user/query scope", fake.listUserID, fake.listUnreadOnly, fake.listLimit, fake.listOffset)
	}
	if strings.Contains(rec.Body.String(), "dedupe") || !strings.Contains(rec.Body.String(), "Deadline dokumen") {
		t.Fatalf("List body = %s, want response DTO without dedupe key", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.CountUnread(rec, reqForUser(http.MethodGet, "/api/notifications/unread-count", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"unread":3`) {
		t.Fatalf("CountUnread status/body = %d/%s, want unread count", rec.Code, rec.Body.String())
	}
	if fake.countUserID != userID {
		t.Fatalf("CountUnread user = %v, want %v", fake.countUserID, userID)
	}

	rec = httptest.NewRecorder()
	h.MarkRead(rec, withRouteParam(reqForUser(http.MethodPatch, "/api/notifications/"+notificationID.String()+"/read", ""), "id", notificationID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("MarkRead status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.markUserID != userID || fake.markID != notificationID {
		t.Fatalf("MarkRead args = (%v, %v), want (%v, %v)", fake.markUserID, fake.markID, userID, notificationID)
	}

	rec = httptest.NewRecorder()
	h.MarkAllRead(rec, reqForUser(http.MethodPatch, "/api/notifications/read-all", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"marked":2`) {
		t.Fatalf("MarkAllRead status/body = %d/%s, want marked count", rec.Code, rec.Body.String())
	}
	if fake.markAllUserID != userID {
		t.Fatalf("MarkAllRead user = %v, want %v", fake.markAllUserID, userID)
	}
}

func TestNotificationHandlersMapServiceErrors(t *testing.T) {
	userID := handlerTestUUID(124)
	notificationID := handlerTestUUID(125)
	reqForUser := func(method, target string) *http.Request {
		return withClaims(httptest.NewRequest(method, target, nil), jwt.MapClaims{"uid": userID.String()})
	}
	errDB := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*Notification, http.ResponseWriter, *http.Request)
		svc        *fakeNotificationService
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "list",
			handler:    (*Notification).List,
			svc:        &fakeNotificationService{Notification: &service.Notification{}, listErr: errDB},
			req:        reqForUser(http.MethodGet, "/api/notifications"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "count",
			handler:    (*Notification).CountUnread,
			svc:        &fakeNotificationService{Notification: &service.Notification{}, countErr: errDB},
			req:        reqForUser(http.MethodGet, "/api/notifications/unread-count"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "mark not found",
			handler:    (*Notification).MarkRead,
			svc:        &fakeNotificationService{Notification: &service.Notification{}, markErr: domain.ErrNotFound},
			req:        withRouteParam(reqForUser(http.MethodPatch, "/api/notifications/"+notificationID.String()+"/read"), "id", notificationID.String()),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "mark internal",
			handler:    (*Notification).MarkRead,
			svc:        &fakeNotificationService{Notification: &service.Notification{}, markErr: errDB},
			req:        withRouteParam(reqForUser(http.MethodPatch, "/api/notifications/"+notificationID.String()+"/read"), "id", notificationID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "mark all",
			handler:    (*Notification).MarkAllRead,
			svc:        &fakeNotificationService{Notification: &service.Notification{}, markAllErr: errDB},
			req:        reqForUser(http.MethodPatch, "/api/notifications/read-all"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Notification{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
