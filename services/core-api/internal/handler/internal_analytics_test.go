package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeInternalAnalyticsService struct {
	input  service.CreateInternalAnalyticsEventInput
	row    service.InternalAnalyticsEventReceipt
	err    error
	called bool
}

func (f *fakeInternalAnalyticsService) CreateEvent(_ context.Context, in service.CreateInternalAnalyticsEventInput) (service.InternalAnalyticsEventReceipt, error) {
	f.called = true
	f.input = in
	if f.err != nil {
		return service.InternalAnalyticsEventReceipt{}, f.err
	}
	if f.row.ID == "" {
		f.row = service.InternalAnalyticsEventReceipt{
			ID:                 "01000000-0000-0000-0000-000000000000",
			EventName:          in.EventName,
			EventGroup:         in.EventGroup,
			OccurredAt:         time.Date(2026, 5, 9, 10, 31, 0, 0, time.UTC),
			RetentionExpiresAt: time.Date(2026, 8, 7, 10, 31, 0, 0, time.UTC),
		}
	}
	return f.row, nil
}

func analyticsRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/internal-analytics/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return withClaims(req, jwt.MapClaims{
		"sub":         "01000000-0000-0000-0000-000000000001",
		"roles":       []any{"guru"},
		"permissions": []any{"dashboard.read"},
	})
}

func TestInternalAnalyticsHandlerCreatesSafeEventFromClaims(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	body := `{"event_name":"dashboard.view","event_group":"dashboard","source_surface":"web_admin","metadata":{"route_group":"dashboard","result":"success"}}`
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, analyticsRequest(body))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.called {
		t.Fatalf("service was not called")
	}
	if fake.input.EventName != "dashboard.view" || fake.input.EventGroup != "dashboard" || fake.input.SourceSurface != "web_admin" {
		t.Fatalf("input identity = %+v, want decoded payload", fake.input)
	}
	if !fake.input.ActorUserID.Valid || fake.input.ActorUserID.String() != "01000000-0000-0000-0000-000000000001" {
		t.Fatalf("actor_user_id = %v, want sub claim", fake.input.ActorUserID)
	}
	if fake.input.ActorRole != "guru" {
		t.Fatalf("actor role = %q, want first role claim", fake.input.ActorRole)
	}
	if fake.input.Metadata["route_group"] != "dashboard" {
		t.Fatalf("metadata = %+v, want object decoded", fake.input.Metadata)
	}
	if strings.Contains(rec.Body.String(), "metadata") || strings.Contains(rec.Body.String(), "route_group") {
		t.Fatalf("response leaked metadata: %s", rec.Body.String())
	}
	var response api.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("response json: %v", err)
	}
	if response.Data == nil {
		t.Fatalf("response data is nil")
	}
}

func TestInternalAnalyticsHandlerRejectsUnauthenticatedRequest(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	req := httptest.NewRequest(http.MethodPost, "/api/internal-analytics/events", strings.NewReader(`{"event_name":"dashboard.view"}`))
	rec := httptest.NewRecorder()

	h.CreateEvent(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
	if fake.called {
		t.Fatalf("service should not be called for unauthenticated request")
	}
}

func TestInternalAnalyticsHandlerRejectsStrictJSONFailures(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "trailing json", body: `{"event_name":"dashboard.view","event_group":"dashboard","source_surface":"web_admin"} {}`, wantStatus: http.StatusBadRequest},
		{name: "null payload", body: `null`, wantStatus: http.StatusBadRequest},
		{name: "array payload", body: `[]`, wantStatus: http.StatusBadRequest},
		{name: "metadata null", body: `{"event_name":"dashboard.view","event_group":"dashboard","source_surface":"web_admin","metadata":null}`, wantStatus: http.StatusBadRequest},
		{name: "metadata array", body: `{"event_name":"dashboard.view","event_group":"dashboard","source_surface":"web_admin","metadata":[]}`, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeInternalAnalyticsService{}
			h := &InternalAnalytics{svc: fake}
			rec := httptest.NewRecorder()

			h.CreateEvent(rec, analyticsRequest(tt.body))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if fake.called {
				t.Fatalf("service should not be called for invalid json")
			}
		})
	}
}

func TestInternalAnalyticsHandlerRejectsOversizedPayload(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	rec := httptest.NewRecorder()
	body := `{"event_name":"dashboard.view","event_group":"dashboard","source_surface":"web_admin","metadata":{"widget_key":"` + strings.Repeat("x", internalAnalyticsBodyLimit) + `"}}`

	h.CreateEvent(rec, analyticsRequest(body))

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413; body=%s", rec.Code, rec.Body.String())
	}
	if fake.called {
		t.Fatalf("service should not be called for oversized request")
	}
}

func TestInternalAnalyticsHandlerMapsValidationErrorsSafely(t *testing.T) {
	fake := &fakeInternalAnalyticsService{
		err: &service.InternalAnalyticsValidationError{
			Code:  service.InternalAnalyticsErrForbiddenMetadataKey,
			Field: "access_token",
		},
	}
	h := &InternalAnalytics{svc: fake}
	rec := httptest.NewRecorder()
	body := `{"event_name":"dashboard.view","event_group":"dashboard","source_surface":"web_admin","metadata":{"widget_key":"safe"}}`

	h.CreateEvent(rec, analyticsRequest(body))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "access-token-value") {
		t.Fatalf("response leaked raw sensitive payload: %s", rec.Body.String())
	}
	var validationErr *service.InternalAnalyticsValidationError
	if !errors.As(fake.err, &validationErr) {
		t.Fatalf("test setup error is not validation error")
	}
}
