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
	input         service.CreateInternalAnalyticsEventInput
	dailyInput    service.InternalAnalyticsDailyQuery
	summaryInput  service.InternalAnalyticsSummaryQuery
	exportInput   service.InternalAnalyticsExportQuery
	row           service.InternalAnalyticsEventReceipt
	dailyResult   service.InternalAnalyticsDailyResult
	summary       service.InternalAnalyticsSummary
	exportResult  service.InternalAnalyticsExport
	err           error
	publicErr     error
	dailyErr      error
	summaryErr    error
	exportErr     error
	called        bool
	publicCalled  bool
	dailyCalled   bool
	summaryCalled bool
	exportCalled  bool
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

func (f *fakeInternalAnalyticsService) CreatePublicEvent(_ context.Context, in service.CreatePublicAnalyticsEventInput) (service.InternalAnalyticsEventReceipt, error) {
	f.publicCalled = true
	f.input = service.CreateInternalAnalyticsEventInput{
		EventName:     in.EventName,
		EventGroup:    "public",
		SourceSurface: "public_website",
		RouteGroup:    in.RouteGroup,
		Metadata:      in.Metadata,
		Result:        in.Result,
	}
	if f.publicErr != nil {
		return service.InternalAnalyticsEventReceipt{}, f.publicErr
	}
	return service.InternalAnalyticsEventReceipt{
		ID:                 "02000000-0000-0000-0000-000000000000",
		EventName:          in.EventName,
		EventGroup:         "public",
		OccurredAt:         time.Date(2026, 5, 9, 10, 31, 0, 0, time.UTC),
		RetentionExpiresAt: time.Date(2026, 8, 7, 10, 31, 0, 0, time.UTC),
	}, nil
}

func (f *fakeInternalAnalyticsService) ListDailyAggregates(_ context.Context, in service.InternalAnalyticsDailyQuery) (service.InternalAnalyticsDailyResult, error) {
	f.dailyCalled = true
	f.dailyInput = in
	if f.dailyErr != nil {
		return service.InternalAnalyticsDailyResult{}, f.dailyErr
	}
	if f.dailyResult.Items == nil {
		f.dailyResult = service.InternalAnalyticsDailyResult{
			Days:       in.Days,
			EventGroup: in.EventGroup,
			Items: []service.InternalAnalyticsDailyItem{{
				AggregateDate: "2026-05-09",
				EventGroup:    "dashboard",
				EventName:     "dashboard.view",
				SourceSurface: "web_admin",
				Count:         3,
			}},
		}
	}
	return f.dailyResult, nil
}

func (f *fakeInternalAnalyticsService) Summary(_ context.Context, in service.InternalAnalyticsSummaryQuery) (service.InternalAnalyticsSummary, error) {
	f.summaryCalled = true
	f.summaryInput = in
	if f.summaryErr != nil {
		return service.InternalAnalyticsSummary{}, f.summaryErr
	}
	if f.summary.TotalCount == 0 {
		f.summary = service.InternalAnalyticsSummary{
			Days:       in.Days,
			TotalCount: 5,
			Groups:     []service.InternalAnalyticsGroupSummary{{EventGroup: "dashboard", Count: 5}},
			TopEvents:  []service.InternalAnalyticsEventSummary{{EventName: "dashboard.view", EventGroup: "dashboard", Count: 5}},
		}
	}
	return f.summary, nil
}

func (f *fakeInternalAnalyticsService) ExportAggregatesCSV(_ context.Context, in service.InternalAnalyticsExportQuery) (service.InternalAnalyticsExport, error) {
	f.exportCalled = true
	f.exportInput = in
	if f.exportErr != nil {
		return service.InternalAnalyticsExport{}, f.exportErr
	}
	if len(f.exportResult.Content) == 0 {
		f.exportResult = service.InternalAnalyticsExport{
			Filename: "internal-analytics-aggregate-2026-05-09.csv",
			Content:  []byte("aggregate_date,event_group,event_name,source_surface,role,result,count\n2026-05-09,dashboard,dashboard.view,web_admin,admin,success,5\n"),
		}
	}
	return f.exportResult, nil
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

func TestInternalAnalyticsHandlerCreatesPublicEventWithoutClaims(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	req := httptest.NewRequest(http.MethodPost, "/api/internal-analytics/public-events", strings.NewReader(`{"event_name":"public.page_view","route_group":"profil","result":"success","metadata":{"page_key":"profil","raw_user_agent":"must-not-forward"}}`))
	rec := httptest.NewRecorder()

	h.CreatePublicEvent(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.publicCalled {
		t.Fatalf("public service was not called")
	}
	if fake.input.EventName != "public.page_view" || fake.input.RouteGroup != "profil" || fake.input.Result != "success" {
		t.Fatalf("public input = %+v, want decoded public event", fake.input)
	}
	if fake.input.Metadata["page_key"] != "profil" || fake.input.Metadata["raw_user_agent"] != "must-not-forward" {
		t.Fatalf("public metadata decode = %+v, want service-layer sanitizer to own filtering", fake.input.Metadata)
	}
	if strings.Contains(rec.Body.String(), "raw_user_agent") || strings.Contains(rec.Body.String(), "must-not-forward") {
		t.Fatalf("public response leaked metadata: %s", rec.Body.String())
	}
}

func TestInternalAnalyticsHandlerRejectsStrictPublicJSON(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	req := httptest.NewRequest(http.MethodPost, "/api/internal-analytics/public-events", strings.NewReader(`{"event_name":"public.page_view","source_surface":"public_website"}`))
	rec := httptest.NewRecorder()

	h.CreatePublicEvent(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if fake.publicCalled {
		t.Fatalf("service should not be called for public payload with unknown fields")
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

func TestInternalAnalyticsHandlerListsDailyAggregatesFromQuery(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	req := analyticsRequest(``)
	req = httptest.NewRequest(http.MethodGet, "/api/internal-analytics/daily?event_group=dashboard&days=14&limit=25", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "permissions": []any{"analytics.read"}})
	rec := httptest.NewRecorder()

	h.ListDailyAggregates(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.dailyCalled {
		t.Fatalf("daily service was not called")
	}
	if fake.dailyInput.EventGroup != "dashboard" || fake.dailyInput.Days != 14 || fake.dailyInput.Limit != 25 {
		t.Fatalf("daily input = %+v, want query filters", fake.dailyInput)
	}
	if strings.Contains(rec.Body.String(), "metadata") || strings.Contains(rec.Body.String(), "raw_user_agent") {
		t.Fatalf("daily response leaked raw metadata: %s", rec.Body.String())
	}
}

func TestInternalAnalyticsHandlerRejectsInvalidDailyQuery(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	req := httptest.NewRequest(http.MethodGet, "/api/internal-analytics/daily?days=0", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "permissions": []any{"analytics.read"}})
	rec := httptest.NewRecorder()

	h.ListDailyAggregates(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if fake.dailyCalled {
		t.Fatalf("daily service should not be called for invalid query")
	}
}

func TestInternalAnalyticsHandlerReturnsSummary(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	req := httptest.NewRequest(http.MethodGet, "/api/internal-analytics/summary?days=30", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "permissions": []any{"analytics.read"}})
	rec := httptest.NewRecorder()

	h.Summary(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.summaryCalled || fake.summaryInput.Days != 30 {
		t.Fatalf("summary input = %+v, called=%v; want days query", fake.summaryInput, fake.summaryCalled)
	}
	if strings.Contains(rec.Body.String(), "metadata") || strings.Contains(rec.Body.String(), "actor_user_id") {
		t.Fatalf("summary response leaked raw event fields: %s", rec.Body.String())
	}
}

func TestInternalAnalyticsHandlerExportsAggregateCSV(t *testing.T) {
	fake := &fakeInternalAnalyticsService{}
	h := &InternalAnalytics{svc: fake}
	req := httptest.NewRequest(http.MethodGet, "/api/internal-analytics/export?event_group=dashboard&days=7&limit=100", nil)
	req = withClaims(req, jwt.MapClaims{
		"sub":   "01000000-0000-0000-0000-000000000001",
		"roles": []any{"admin"},
	})
	rec := httptest.NewRecorder()

	h.ExportAggregates(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.exportCalled {
		t.Fatalf("export service was not called")
	}
	if fake.exportInput.EventGroup != "dashboard" || fake.exportInput.Days != 7 || fake.exportInput.Limit != 100 {
		t.Fatalf("export input = %+v, want query filters", fake.exportInput)
	}
	if !fake.exportInput.ActorUserID.Valid || fake.exportInput.ActorRole != "admin" {
		t.Fatalf("export actor = %v/%q, want JWT actor context", fake.exportInput.ActorUserID, fake.exportInput.ActorRole)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("content-type = %q, want CSV", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("nosniff header = %q, want nosniff", got)
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("content-disposition = %q, want attachment", rec.Header().Get("Content-Disposition"))
	}
	for _, forbidden := range []string{"metadata", "actor_user_id", "session_id", "raw_user_agent", "ip_address", "user_agent"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Fatalf("export response leaked %q: %s", forbidden, rec.Body.String())
		}
	}
}
