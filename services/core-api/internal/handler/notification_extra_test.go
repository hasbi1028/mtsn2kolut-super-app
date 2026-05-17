package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func TestNotificationExtraUserIDFromClaims(t *testing.T) {
	userID := handlerTestUUID(230)

	t.Run("uses sub claim", func(t *testing.T) {
		req := withClaims(httptest.NewRequest(http.MethodGet, "/api/notifications", nil), jwt.MapClaims{"sub": userID.String(), "uid": handlerTestUUID(231).String()})
		rec := httptest.NewRecorder()
		got, ok := notificationUserID(rec, req)
		if !ok || got != userID || rec.Code != http.StatusOK {
			t.Fatalf("notificationUserID(sub) = %v/%v code=%d, want sub user", got, ok, rec.Code)
		}
	})

	t.Run("falls back to uid claim", func(t *testing.T) {
		req := withClaims(httptest.NewRequest(http.MethodGet, "/api/notifications", nil), jwt.MapClaims{"uid": userID.String()})
		rec := httptest.NewRecorder()
		got, ok := notificationUserID(rec, req)
		if !ok || got != userID || rec.Code != http.StatusOK {
			t.Fatalf("notificationUserID(uid) = %v/%v code=%d, want uid user", got, ok, rec.Code)
		}
	})

	t.Run("rejects missing claims", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
		rec := httptest.NewRecorder()
		got, ok := notificationUserID(rec, req)
		if ok || got.Valid || rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "unauthorized") {
			t.Fatalf("notificationUserID(no claims) = %v/%v code=%d body=%s, want unauthorized", got, ok, rec.Code, rec.Body.String())
		}
	})

	t.Run("rejects invalid claim uuid", func(t *testing.T) {
		req := withClaims(httptest.NewRequest(http.MethodGet, "/api/notifications", nil), jwt.MapClaims{"sub": "not-a-uuid"})
		rec := httptest.NewRecorder()
		got, ok := notificationUserID(rec, req)
		if ok || got.Valid || rec.Code != http.StatusUnauthorized {
			t.Fatalf("notificationUserID(invalid) = %v/%v code=%d, want unauthorized", got, ok, rec.Code)
		}
	})
}

func TestNotificationExtraHandlersStopBeforeServiceWhenUnauthorized(t *testing.T) {
	fake := &fakeNotificationService{Notification: &service.Notification{}}
	h := &Notification{svc: fake}

	cases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		req     *http.Request
	}{
		{name: "list", handler: h.List, req: httptest.NewRequest(http.MethodGet, "/api/notifications", nil)},
		{name: "count", handler: h.CountUnread, req: httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)},
		{name: "mark", handler: h.MarkRead, req: withRouteParam(httptest.NewRequest(http.MethodPatch, "/api/notifications/11111111-1111-1111-1111-111111111111/read", nil), "id", "11111111-1111-1111-1111-111111111111")},
		{name: "mark all", handler: h.MarkAllRead, req: httptest.NewRequest(http.MethodPatch, "/api/notifications/read-all", nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.handler(rec, tc.req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
			}
		})
	}

	if fake.listUserID.Valid || fake.countUserID.Valid || fake.markUserID.Valid || fake.markAllUserID.Valid {
		t.Fatalf("service was called despite unauthorized request: list=%v count=%v mark=%v markAll=%v", fake.listUserID, fake.countUserID, fake.markUserID, fake.markAllUserID)
	}
}

func TestNotificationExtraMarkReadRejectsInvalidNotificationID(t *testing.T) {
	userID := handlerTestUUID(232)
	fake := &fakeNotificationService{Notification: &service.Notification{}}
	h := &Notification{svc: fake}
	req := withClaims(httptest.NewRequest(http.MethodPatch, "/api/notifications/not-a-uuid/read", nil), jwt.MapClaims{"sub": userID.String()})
	req = withRouteParam(req, "id", "not-a-uuid")
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if fake.markUserID.Valid || fake.markID.Valid {
		t.Fatalf("MarkRead service called with user/id %v/%v, want blocked", fake.markUserID, fake.markID)
	}
}

func TestNotificationExtraResponseCopiesAllFields(t *testing.T) {
	id := handlerTestUUID(233)
	userID := handlerTestUUID(234)
	readAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC), Valid: true}
	createdAt := pgtype.Timestamptz{Time: readAt.Time.Add(-1), Valid: true}
	row := notificationTestRow(id, userID)
	row.ReadAt = readAt
	row.CreatedAt = createdAt

	resp := notificationResponseFromRow(row)
	if resp.ID != id || resp.UserID != userID || resp.Category != row.Category || resp.Title != row.Title || resp.Body != row.Body || resp.EntityType != row.EntityType || resp.EntityID != row.EntityID || resp.LinkPath != row.LinkPath || resp.ReadAt != readAt || resp.CreatedAt != createdAt {
		t.Fatalf("notificationResponseFromRow() = %+v, want copy of %+v", resp, row)
	}

	list := notificationResponses(nil)
	if len(list) != 0 || list == nil {
		t.Fatalf("notificationResponses(nil) = %#v, want non-nil empty slice", list)
	}
	list = notificationResponses([]db.AppNotification{row})
	if len(list) != 1 || list[0].ID != id {
		t.Fatalf("notificationResponses(row) = %+v, want one copied row", list)
	}
}

func TestNotificationExtraClaimsContextRejectsUnexpectedClaimType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, map[string]any{"sub": handlerTestUUID(235).String()}))
	rec := httptest.NewRecorder()
	got, ok := notificationUserID(rec, req)
	if ok || got.Valid || rec.Code != http.StatusUnauthorized {
		t.Fatalf("notificationUserID(plain map) = %v/%v code=%d, want unauthorized", got, ok, rec.Code)
	}
}
