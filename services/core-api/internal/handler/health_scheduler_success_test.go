package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestHealthGetReportsDatabaseState(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Health{pool: fakeHealthDB{}}).Get(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Health.Get(ok) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Health{pool: fakeHealthDB{err: errors.New("down")}}).Get(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("Health.Get(error) status = %d, want 503; body=%s", rec.Code, rec.Body.String())
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
