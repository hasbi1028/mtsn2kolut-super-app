package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeInternalAnalyticsStore struct {
	calls []db.CreateInternalAnalyticsEventParams
	row   db.InternalAnalyticsEvent
	err   error
}

func (f *fakeInternalAnalyticsStore) CreateInternalAnalyticsEvent(_ context.Context, arg db.CreateInternalAnalyticsEventParams) (db.InternalAnalyticsEvent, error) {
	f.calls = append(f.calls, arg)
	if f.err != nil {
		return db.InternalAnalyticsEvent{}, f.err
	}
	row := f.row
	if !row.ID.Valid {
		row.ID = serviceTestUUID(11)
	}
	if row.EventName == "" {
		row.EventName = arg.EventName
	}
	if row.EventGroup == "" {
		row.EventGroup = arg.EventGroup
	}
	if !row.OccurredAt.Valid {
		row.OccurredAt = pgtype.Timestamptz{Time: fixedAnalyticsNow(), Valid: true}
	}
	if !row.RetentionExpiresAt.Valid {
		row.RetentionExpiresAt = arg.RetentionExpiresAt
	}
	return row, nil
}

func fixedAnalyticsNow() time.Time {
	return time.Date(2026, 5, 9, 10, 30, 0, 0, time.UTC)
}

func serviceTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func newTestInternalAnalytics(store *fakeInternalAnalyticsStore) *InternalAnalytics {
	return &InternalAnalytics{
		q:   store,
		now: fixedAnalyticsNow,
	}
}

func expectInternalAnalyticsValidationCode(t *testing.T, err error, code string) {
	t.Helper()
	var validationErr *InternalAnalyticsValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want InternalAnalyticsValidationError", err)
	}
	if validationErr.Code != code {
		t.Fatalf("validation code = %q, want %q", validationErr.Code, code)
	}
}

func TestInternalAnalyticsAcceptsAllowlistedSafeEvent(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	receipt, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "public.page_view",
		EventGroup:    "public",
		SourceSurface: "public_website",
		Metadata: map[string]any{
			"route_group":  "home",
			"device_class": "desktop",
		},
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if receipt.EventName != "public.page_view" || receipt.EventGroup != "public" {
		t.Fatalf("receipt = %+v, want safe event receipt", receipt)
	}
	if len(store.calls) != 1 {
		t.Fatalf("CreateInternalAnalyticsEvent calls = %d, want 1", len(store.calls))
	}
	call := store.calls[0]
	if call.EventName != "public.page_view" || call.EventGroup != "public" || call.SourceSurface != "public_website" {
		t.Fatalf("stored event params = %+v, want event identity and source", call)
	}
	wantRetention := fixedAnalyticsNow().Add(90 * 24 * time.Hour)
	if !call.RetentionExpiresAt.Valid || !call.RetentionExpiresAt.Time.Equal(wantRetention) {
		t.Fatalf("retention = %v, want public default %v", call.RetentionExpiresAt, wantRetention)
	}
	var metadata map[string]any
	if err := json.Unmarshal(call.Metadata, &metadata); err != nil {
		t.Fatalf("stored metadata is not json object: %v", err)
	}
	if metadata["route_group"] != "home" || metadata["device_class"] != "desktop" {
		t.Fatalf("metadata = %+v, want sanitized safe keys", metadata)
	}
}

func TestInternalAnalyticsDefaultRetentionByGroup(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		group     string
		wantDays  int
	}{
		{name: "public", eventName: "public.page_view", group: "public", wantDays: 90},
		{name: "internal", eventName: "auth.login_success", group: "auth", wantDays: 180},
		{name: "security", eventName: "security.forbidden", group: "security", wantDays: 180},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeInternalAnalyticsStore{}
			svc := newTestInternalAnalytics(store)
			_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
				EventName:     tt.eventName,
				EventGroup:    tt.group,
				SourceSurface: "core_api",
				Metadata:      map[string]any{"result": "success"},
			})
			if err != nil {
				t.Fatalf("CreateEvent() error = %v", err)
			}
			got := store.calls[0].RetentionExpiresAt.Time
			want := fixedAnalyticsNow().Add(time.Duration(tt.wantDays) * 24 * time.Hour)
			if !got.Equal(want) {
				t.Fatalf("retention = %v, want %v", got, want)
			}
		})
	}
}

func TestInternalAnalyticsRejectsUnknownEvent(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "unknown.event",
		EventGroup:    "unknown",
		SourceSurface: "web_admin",
		Metadata:      map[string]any{},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrEventNotAllowlisted)
	if len(store.calls) != 0 {
		t.Fatalf("store called for rejected event")
	}
}

func TestInternalAnalyticsRejectsGroupMismatch(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "public.page_view",
		EventGroup:    "auth",
		SourceSurface: "public_website",
		Metadata:      map[string]any{},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrGroupMismatch)
	if len(store.calls) != 0 {
		t.Fatalf("store called for rejected event")
	}
}

func TestInternalAnalyticsRejectsForbiddenMetadataKeysCaseInsensitiveAndNested(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "dashboard.view",
		EventGroup:    "dashboard",
		SourceSurface: "web_admin",
		Metadata: map[string]any{
			"safe": map[string]any{
				"Access_Token": "must-not-be-returned",
			},
		},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrForbiddenMetadataKey)
	if strings.Contains(err.Error(), "must-not-be-returned") {
		t.Fatalf("validation error leaked raw sensitive value: %q", err.Error())
	}
	if len(store.calls) != 0 {
		t.Fatalf("store called for rejected metadata")
	}
}

func TestInternalAnalyticsRejectsMetadataTooLarge(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "dashboard.view",
		EventGroup:    "dashboard",
		SourceSurface: "web_admin",
		Metadata: map[string]any{
			"widget_key": strings.Repeat("x", internalAnalyticsMetadataMaxBytes+1),
		},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrMetadataTooLarge)
	if len(store.calls) != 0 {
		t.Fatalf("store called for oversized metadata")
	}
}

func TestInternalAnalyticsRejectsInvalidSourceSurface(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "dashboard.view",
		EventGroup:    "dashboard",
		SourceSurface: "browser",
		Metadata:      map[string]any{},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrInvalidSourceSurface)
	if len(store.calls) != 0 {
		t.Fatalf("store called for invalid source surface")
	}
}
