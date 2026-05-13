package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeHealthDB struct {
	err error
}

func (f fakeHealthDB) Ping(context.Context) error {
	return f.err
}

type fakeSchedulerTickService struct {
	called bool
	err    error
}

func (f *fakeSchedulerTickService) Tick(ctx context.Context, now time.Time) (service.PusakaSchedulerResult, error) {
	f.called = true
	return service.PusakaSchedulerResult{Processed: 2, Enqueued: 3, Skipped: 1}, f.err
}

type fakeHealthAnalytics struct {
	backlog service.InternalAnalyticsExpiredBacklog
	err     error
	called  bool
}

func (f *fakeHealthAnalytics) ExpiredBacklog(context.Context) (service.InternalAnalyticsExpiredBacklog, error) {
	f.called = true
	return f.backlog, f.err
}

func TestHealthGetReportsDatabaseState(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Health{pool: fakeHealthDB{}}).Get(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Health.Get(ok) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"clock"`) || !strings.Contains(rec.Body.String(), `"server_utc"`) {
		t.Fatalf("Health.Get(ok) missing clock diagnostics: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Health{pool: fakeHealthDB{err: errors.New("down")}}).Get(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("Health.Get(error) status = %d, want 503; body=%s", rec.Code, rec.Body.String())
	}
}

func TestHealthGetReportsInternalAnalyticsExpiredBacklog(t *testing.T) {
	oldest := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	analytics := &fakeHealthAnalytics{backlog: service.InternalAnalyticsExpiredBacklog{
		ExpiredEventBacklogCount: 6,
		OldestExpiredEventAt:     &oldest,
	}}
	rec := httptest.NewRecorder()

	(&Health{pool: fakeHealthDB{}, analytics: analytics}).Get(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("Health.Get() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !analytics.called {
		t.Fatalf("analytics backlog service was not called")
	}
	body := rec.Body.String()
	for _, want := range []string{`"analytics"`, `"expired_event_backlog_count":6`, `"oldest_expired_event_at":"2026-05-01T00:00:00Z"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("health body missing %q: %s", want, body)
		}
	}
	for _, forbidden := range []string{"metadata", "actor_user_id", "raw_user_agent", "session_id"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("health analytics summary leaked %q: %s", forbidden, body)
		}
	}
}

func TestPusakaSchedulerTickHandlerForwardsToService(t *testing.T) {
	fake := &fakeSchedulerTickService{}
	h := &PusakaScheduler{svc: fake}
	rec := httptest.NewRecorder()
	h.Tick(rec, adminRequest(http.MethodPost, "/api/pusaka/scheduler/tick", ""))
	if rec.Code != http.StatusOK || !fake.called {
		t.Fatalf("Tick() status/called = %d/%v, want 200/true; body=%s", rec.Code, fake.called, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h = &PusakaScheduler{svc: &fakeSchedulerTickService{err: errors.New("tick failed")}}
	h.Tick(rec, adminRequest(http.MethodPost, "/api/pusaka/scheduler/tick", ""))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Tick(error) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
}
