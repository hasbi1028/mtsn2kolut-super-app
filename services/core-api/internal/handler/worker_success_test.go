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

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeWorkerJobService struct {
	claimWorkerID string
	claimRow      db.ClaimJobRow
	claimErr      error

	getID  pgtype.UUID
	getRow db.GetJobRow
	getErr error

	completeID                 pgtype.UUID
	completeWorkerID           string
	completeErr                error
	completeAttendanceID       pgtype.UUID
	completeAttendanceWorkerID string
	completeAttendanceInput    *service.PusakaJobAttendanceInput
	completeAttendanceErr      error

	failID       pgtype.UUID
	failWorkerID string
	failError    string
	failRetry    pgtype.Text
	failErr      error
	statsRow     db.GetJobStatsRow
	statsErr     error
	statsCalls   int
}

func (f *fakeWorkerJobService) Claim(ctx context.Context, workerID string) (db.ClaimJobRow, error) {
	f.claimWorkerID = workerID
	return f.claimRow, f.claimErr
}

func (f *fakeWorkerJobService) Get(ctx context.Context, id pgtype.UUID) (db.GetJobRow, error) {
	f.getID = id
	return f.getRow, f.getErr
}

func (f *fakeWorkerJobService) Complete(ctx context.Context, id pgtype.UUID, workerID string) error {
	f.completeID = id
	f.completeWorkerID = workerID
	return f.completeErr
}

func (f *fakeWorkerJobService) CompleteWithAttendance(ctx context.Context, id pgtype.UUID, workerID string, attendance *service.PusakaJobAttendanceInput) error {
	f.completeAttendanceID = id
	f.completeAttendanceWorkerID = workerID
	f.completeAttendanceInput = attendance
	return f.completeAttendanceErr
}

func (f *fakeWorkerJobService) Fail(ctx context.Context, id pgtype.UUID, workerID string, errorMessage string, retryAfter pgtype.Text) error {
	f.failID = id
	f.failWorkerID = workerID
	f.failError = errorMessage
	f.failRetry = retryAfter
	return f.failErr
}

func (f *fakeWorkerJobService) Stats(context.Context) (db.GetJobStatsRow, error) {
	f.statsCalls++
	return f.statsRow, f.statsErr
}

type fakeWorkerAttendanceService struct {
	upsertArg db.UpsertAttendanceParams
	upsertErr error
}

func (f *fakeWorkerAttendanceService) Upsert(ctx context.Context, p db.UpsertAttendanceParams) (db.AttendanceRecord, error) {
	f.upsertArg = p
	return db.AttendanceRecord{EmployeeID: p.EmployeeID, Tanggal: p.Tanggal, JamMasuk: p.JamMasuk, JamPulang: p.JamPulang}, f.upsertErr
}

func TestPusakaWorkerSuccessHandlersForwardPayloads(t *testing.T) {
	jobID := handlerTestUUID(180)
	employeeID := handlerTestUUID(181)
	sourceJobID := handlerTestUUID(182)
	jobs := &fakeWorkerJobService{
		claimRow: db.ClaimJobRow{ID: jobID, EmployeeID: employeeID},
		getRow:   db.GetJobRow{ID: jobID, EmployeeID: employeeID},
		statsRow: db.GetJobStatsRow{Queued: 1, Running: 2, Success: 3, Failed: 4},
	}
	attendance := &fakeWorkerAttendanceService{}
	settings := &fakeSettingService{
		listRows: []db.AppSetting{
			{Key: "max_concurrent", Value: "5"},
			{Key: "headless", Value: "true"},
			{Key: "pusaka_geo_base_lat", Value: "-3.2163111"},
			{Key: "pusaka_geo_base_lng", Value: "121.0428659"},
			{Key: "pusaka_geo_default_radius_m", Value: "50"},
			{Key: "pusaka_geo_checkin_radius_m", Value: "55"},
			{Key: "pusaka_geo_checkout_radius_m", Value: "28"},
			{Key: "ignored", Value: "value"},
			{Key: "worker_status:active", Value: `{"worker_id":"active","reported_at":"` + time.Now().UTC().Format(time.RFC3339) + `","active_consumers":2}`},
			{Key: "worker_status:stale", Value: `{"worker_id":"stale","reported_at":"` + time.Now().UTC().Add(-10*time.Minute).Format(time.RFC3339) + `"}`},
			{Key: "worker_status:bad", Value: `{bad json`},
		},
	}
	h := &PusakaWorker{jobs: jobs, att: attendance, sett: settings}

	rec := httptest.NewRecorder()
	h.Claim(rec, httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/claim", strings.NewReader(`{"worker_id":"worker-1"}`)))
	if rec.Code != http.StatusOK || jobs.claimWorkerID != "worker-1" {
		t.Fatalf("Claim() status/worker = %d/%q, want 200/worker-1", rec.Code, jobs.claimWorkerID)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/jobs/"+jobID.String()+"/complete", strings.NewReader(`{"worker_id":"worker-1","tanggal":"2026-05-01","jam_masuk":"07:10","jam_pulang":"15:00"}`)), "id", jobID.String())
	h.Complete(rec, req)
	if rec.Code != http.StatusNoContent || jobs.completeAttendanceID != jobID || jobs.completeAttendanceWorkerID != "worker-1" {
		t.Fatalf("Complete() status/complete/worker = %d/%v/%q", rec.Code, jobs.completeAttendanceID, jobs.completeAttendanceWorkerID)
	}
	if jobs.completeAttendanceInput == nil || !jobs.completeAttendanceInput.Tanggal.Valid || jobs.completeAttendanceInput.JamMasuk != "07:10" || jobs.completeAttendanceInput.JamPulang != "15:00" {
		t.Fatalf("Complete() attendance input = %+v, want parsed attendance", jobs.completeAttendanceInput)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/jobs/"+jobID.String()+"/fail", strings.NewReader(`{"worker_id":"worker-1","error":"timeout","retry_after_secs":"120"}`)), "id", jobID.String())
	h.Fail(rec, req)
	if rec.Code != http.StatusNoContent || jobs.failID != jobID || jobs.failWorkerID != "worker-1" || jobs.failError != "timeout" || jobs.failRetry.String != "120" {
		t.Fatalf("Fail() status/args = %d/%v/%q/%q/%+v", rec.Code, jobs.failID, jobs.failWorkerID, jobs.failError, jobs.failRetry)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/jobs/"+jobID.String()+"/fail", strings.NewReader(`{"worker_id":"worker-1","error":"timeout","retry_after_secs":180}`)), "id", jobID.String())
	h.Fail(rec, req)
	if rec.Code != http.StatusNoContent || jobs.failRetry.String != "180" {
		t.Fatalf("Fail(numeric retry) status/retry = %d/%q, want 204/180", rec.Code, jobs.failRetry.String)
	}

	rec = httptest.NewRecorder()
	h.UpsertAttendance(rec, httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/attendance", strings.NewReader(`{"employee_id":"`+employeeID.String()+`","tanggal":"2026-05-02","jam_masuk":"07:20","jam_pulang":"14:45","source_job_id":"`+sourceJobID.String()+`"}`)))
	if rec.Code != http.StatusOK || attendance.upsertArg.EmployeeID != employeeID || attendance.upsertArg.SourceJobID != sourceJobID {
		t.Fatalf("UpsertAttendance() status/arg = %d/%+v", rec.Code, attendance.upsertArg)
	}

	rec = httptest.NewRecorder()
	h.Config(rec, httptest.NewRequest(http.MethodGet, "/api/pusaka/worker/config", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Config() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var cfg struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("Config() json error = %v", err)
	}
	if cfg.Data["max_concurrent"] != "5" ||
		cfg.Data["headless"] != "true" ||
		cfg.Data["pusaka_geo_base_lat"] != "-3.2163111" ||
		cfg.Data["pusaka_geo_checkin_radius_m"] != "55" ||
		cfg.Data["ignored"] != "" {
		t.Fatalf("Config() data = %+v, want only worker config keys", cfg.Data)
	}

	rec = httptest.NewRecorder()
	h.Heartbeat(rec, httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/heartbeat", strings.NewReader(`{"worker_id":"worker-1","active_consumers":2,"target_concurrency":5,"headless":true,"last_sync_at":"2026-05-01T00:00:00Z"}`)))
	if rec.Code != http.StatusOK || settings.upsertKey != "worker_status:worker-1" {
		t.Fatalf("Heartbeat() status/upsert key = %d/%q", rec.Code, settings.upsertKey)
	}
	var heartbeat map[string]any
	if err := json.Unmarshal([]byte(settings.upsertValue), &heartbeat); err != nil {
		t.Fatalf("Heartbeat() stored json error = %v; value=%s", err, settings.upsertValue)
	}
	if heartbeat["worker_id"] != "worker-1" || heartbeat["active_consumers"].(float64) != 2 {
		t.Fatalf("Heartbeat() payload = %+v, want worker status", heartbeat)
	}

	rec = httptest.NewRecorder()
	h.GetStatus(rec, adminRequest(http.MethodGet, "/api/pusaka/worker/status", ""))
	if rec.Code != http.StatusOK || jobs.statsCalls != 1 {
		t.Fatalf("GetStatus() status/stats = %d/%d, want 200/1", rec.Code, jobs.statsCalls)
	}
	var status struct {
		Data struct {
			Total int `json:"total"`
			Queue struct {
				Queued  int64 `json:"queued"`
				Running int64 `json:"running"`
				Success int64 `json:"success"`
				Failed  int64 `json:"failed"`
			} `json:"queue"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatalf("GetStatus() json error = %v", err)
	}
	if status.Data.Total != 1 || status.Data.Queue.Queued != 1 || status.Data.Queue.Failed != 4 {
		t.Fatalf("GetStatus() data = %+v, want active worker and queue stats", status.Data)
	}
}

func TestPusakaWorkerClaimNoJobReturnsNoContent(t *testing.T) {
	h := &PusakaWorker{jobs: &fakeWorkerJobService{claimErr: domain.ErrNoJob}}
	rec := httptest.NewRecorder()
	h.Claim(rec, httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/claim", strings.NewReader(`{"worker_id":"worker-1"}`)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Claim(no job) status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
}

func TestNewPusakaWorker(t *testing.T) {
	if NewPusakaWorker(nil, nil, nil) == nil {
		t.Fatal("NewPusakaWorker(nil) = nil")
	}
}

func TestPusakaWorkerValidationAndServiceErrors(t *testing.T) {
	jobID := handlerTestUUID(183)
	employeeID := handlerTestUUID(184)

	t.Run("claim rejects invalid body and maps service error", func(t *testing.T) {
		h := &PusakaWorker{jobs: &fakeWorkerJobService{}}
		rec := httptest.NewRecorder()
		h.Claim(rec, httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/claim", strings.NewReader(`{bad`)))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Claim(invalid json) status = %d, want 400", rec.Code)
		}

		h.jobs = &fakeWorkerJobService{claimErr: errors.New("claim failed")}
		rec = httptest.NewRecorder()
		h.Claim(rec, httptest.NewRequest(http.MethodPost, "/api/pusaka/worker/claim", strings.NewReader(`{"worker_id":"worker-1"}`)))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Claim(service error) status = %d, want 500", rec.Code)
		}
	})

	t.Run("complete validates id body and maps service errors", func(t *testing.T) {
		jobs := &fakeWorkerJobService{getRow: db.GetJobRow{ID: jobID, EmployeeID: employeeID}}
		attendance := &fakeWorkerAttendanceService{}
		h := &PusakaWorker{jobs: jobs, att: attendance}

		rec := httptest.NewRecorder()
		h.Complete(rec, withRouteParam(httptest.NewRequest(http.MethodPost, "/job/bad/complete", strings.NewReader(`{}`)), "id", "bad"))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Complete(invalid id) status = %d, want 400", rec.Code)
		}

		jobs.completeAttendanceErr = errors.New("complete failed")
		rec = httptest.NewRecorder()
		req := withRouteParam(httptest.NewRequest(http.MethodPost, "/job/"+jobID.String()+"/complete", strings.NewReader(`{"worker_id":"worker-1","tanggal":"2026-05-01"}`)), "id", jobID.String())
		h.Complete(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Complete(service error) status = %d, want 500", rec.Code)
		}

		jobs.completeAttendanceErr = nil
		rec = httptest.NewRecorder()
		req = withRouteParam(httptest.NewRequest(http.MethodPost, "/job/"+jobID.String()+"/complete", strings.NewReader(`{"worker_id":"worker-1","tanggal":"bad"}`)), "id", jobID.String())
		h.Complete(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Complete(invalid tanggal) status = %d, want 400", rec.Code)
		}

		jobs.completeAttendanceErr = errors.New("complete failed")
		jobs.getID = pgtype.UUID{}
		rec = httptest.NewRecorder()
		req = withRouteParam(httptest.NewRequest(http.MethodPost, "/job/"+jobID.String()+"/complete", strings.NewReader(`{"worker_id":"worker-1"}`)), "id", jobID.String())
		h.Complete(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Complete(complete error) status = %d, want 500", rec.Code)
		}
		if jobs.getID.Valid {
			t.Fatalf("Complete(no tanggal) unexpectedly loaded job id %v", jobs.getID)
		}
		if jobs.completeAttendanceInput != nil {
			t.Fatalf("Complete(no tanggal) attendance input = %+v, want nil", jobs.completeAttendanceInput)
		}
	})

	t.Run("fail validates request and defaults retry", func(t *testing.T) {
		jobs := &fakeWorkerJobService{}
		h := &PusakaWorker{jobs: jobs}

		rec := httptest.NewRecorder()
		h.Fail(rec, withRouteParam(httptest.NewRequest(http.MethodPost, "/job/bad/fail", strings.NewReader(`{}`)), "id", "bad"))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Fail(invalid id) status = %d, want 400", rec.Code)
		}

		rec = httptest.NewRecorder()
		req := withRouteParam(httptest.NewRequest(http.MethodPost, "/job/"+jobID.String()+"/fail", strings.NewReader(`{bad`)), "id", jobID.String())
		h.Fail(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Fail(invalid json) status = %d, want 400", rec.Code)
		}

		rec = httptest.NewRecorder()
		req = withRouteParam(httptest.NewRequest(http.MethodPost, "/job/"+jobID.String()+"/fail", strings.NewReader(`{"worker_id":"worker-1","error":"timeout"}`)), "id", jobID.String())
		h.Fail(rec, req)
		if rec.Code != http.StatusNoContent || jobs.failRetry.String != "60" {
			t.Fatalf("Fail(default retry) status/retry = %d/%q, want 204/60", rec.Code, jobs.failRetry.String)
		}

		rec = httptest.NewRecorder()
		req = withRouteParam(httptest.NewRequest(http.MethodPost, "/job/"+jobID.String()+"/fail", strings.NewReader(`{"worker_id":"worker-1","error":"timeout","retry_after_secs":120}`)), "id", jobID.String())
		h.Fail(rec, req)
		if rec.Code != http.StatusNoContent || jobs.failRetry.String != "120" {
			t.Fatalf("Fail(numeric retry) status/retry = %d/%q, want 204/120", rec.Code, jobs.failRetry.String)
		}

		for name, retryBody := range map[string]string{
			"negative":  `-1`,
			"decimal":   `1.5`,
			"text":      `"soon"`,
			"too large": `86401`,
		} {
			rec = httptest.NewRecorder()
			req = withRouteParam(httptest.NewRequest(http.MethodPost, "/job/"+jobID.String()+"/fail", strings.NewReader(`{"worker_id":"worker-1","error":"timeout","retry_after_secs":`+retryBody+`}`)), "id", jobID.String())
			h.Fail(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("Fail(invalid retry %s) status = %d, want 400", name, rec.Code)
			}
		}

		jobs.failErr = errors.New("fail failed")
		rec = httptest.NewRecorder()
		req = withRouteParam(httptest.NewRequest(http.MethodPost, "/job/"+jobID.String()+"/fail", strings.NewReader(`{"worker_id":"worker-1","error":"timeout"}`)), "id", jobID.String())
		h.Fail(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Fail(service error) status = %d, want 500", rec.Code)
		}
	})

	t.Run("upsert attendance validates fields and maps service error", func(t *testing.T) {
		attendance := &fakeWorkerAttendanceService{}
		h := &PusakaWorker{att: attendance}

		for name, body := range map[string]string{
			"Data yang dikirim tidak valid": `{bad`,
			"invalid employee":              `{"employee_id":"bad","tanggal":"2026-05-01"}`,
			"invalid date":                  `{"employee_id":"` + employeeID.String() + `","tanggal":"bad"}`,
			"invalid source id":             `{"employee_id":"` + employeeID.String() + `","tanggal":"2026-05-01","source_job_id":"bad"}`,
		} {
			rec := httptest.NewRecorder()
			h.UpsertAttendance(rec, httptest.NewRequest(http.MethodPost, "/attendance", strings.NewReader(body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("UpsertAttendance(%s) status = %d, want 400", name, rec.Code)
			}
		}

		attendance.upsertErr = errors.New("upsert failed")
		rec := httptest.NewRecorder()
		h.UpsertAttendance(rec, httptest.NewRequest(http.MethodPost, "/attendance", strings.NewReader(`{"employee_id":"`+employeeID.String()+`","tanggal":"2026-05-01"}`)))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("UpsertAttendance(service error) status = %d, want 500", rec.Code)
		}
	})

	t.Run("config heartbeat and status map settings errors", func(t *testing.T) {
		settings := &fakeSettingService{listErr: errors.New("settings failed")}
		h := &PusakaWorker{sett: settings, jobs: &fakeWorkerJobService{}}

		rec := httptest.NewRecorder()
		h.Config(rec, httptest.NewRequest(http.MethodGet, "/config", nil))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Config(service error) status = %d, want 500", rec.Code)
		}

		rec = httptest.NewRecorder()
		h.Heartbeat(rec, httptest.NewRequest(http.MethodPost, "/heartbeat", strings.NewReader(`{bad`)))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Heartbeat(invalid json) status = %d, want 400", rec.Code)
		}

		settings.listErr = nil
		settings.upsertErr = errors.New("upsert failed")
		rec = httptest.NewRecorder()
		h.Heartbeat(rec, httptest.NewRequest(http.MethodPost, "/heartbeat", strings.NewReader(`{"worker_id":"worker-1"}`)))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Heartbeat(upsert error) status = %d, want 500", rec.Code)
		}

		rec = httptest.NewRecorder()
		h.GetStatus(rec, httptest.NewRequest(http.MethodGet, "/status", nil))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("GetStatus(forbidden) status = %d, want 403", rec.Code)
		}

		settings.listErr = errors.New("list failed")
		rec = httptest.NewRecorder()
		h.GetStatus(rec, adminRequest(http.MethodGet, "/status", ""))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("GetStatus(list error) status = %d, want 500", rec.Code)
		}

		settings.listErr = nil
		h.jobs = &fakeWorkerJobService{statsErr: errors.New("stats failed")}
		rec = httptest.NewRecorder()
		h.GetStatus(rec, adminRequest(http.MethodGet, "/status", ""))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("GetStatus(stats error) status = %d, want 500", rec.Code)
		}
	})
}
