package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeSettingService struct {
	*service.Setting

	listRows         []db.AppSetting
	listErr          error
	upsertKey        string
	upsertValue      string
	upsertErr        error
	profile          service.SchoolProfile
	profileErr       error
	updateProfileArg service.SchoolProfile
	updateProfile    service.SchoolProfile
	updateProfileErr error
}

func (f *fakeSettingService) List(context.Context) ([]db.AppSetting, error) {
	return f.listRows, f.listErr
}

func (f *fakeSettingService) Upsert(_ context.Context, key, value string) error {
	f.upsertKey = key
	f.upsertValue = value
	return f.upsertErr
}

func (f *fakeSettingService) SchoolProfile(context.Context) (service.SchoolProfile, error) {
	return f.profile, f.profileErr
}

func (f *fakeSettingService) UpdateSchoolProfile(_ context.Context, profile service.SchoolProfile) (service.SchoolProfile, error) {
	f.updateProfileArg = profile
	return f.updateProfile, f.updateProfileErr
}

type fakePusakaJobService struct {
	*service.PusakaJob

	listStatus        string
	listLimit         int32
	listOffset        int32
	listRows          []db.ListJobsRow
	listTotal         int64
	listErr           error
	createEmployeeID  pgtype.UUID
	createRunType     string
	createAttempts    int32
	createRow         db.Job
	createErr         error
	statsRow          db.GetJobStatsRow
	statsErr          error
	runAllType        string
	runAllAttempts    int32
	runAllInserted    int
	runAllSkipped     int
	runAllErr         error
	cancelEmployeeID  pgtype.UUID
	cancelEmployeeN   int64
	cancelEmployeeErr error
	cancelAllN        int64
	cancelAllErr      error
}

func (f *fakePusakaJobService) List(_ context.Context, status string, limit, offset int32) ([]db.ListJobsRow, int64, error) {
	f.listStatus = status
	f.listLimit = limit
	f.listOffset = offset
	return f.listRows, f.listTotal, f.listErr
}

func (f *fakePusakaJobService) Create(_ context.Context, employeeID pgtype.UUID, runType string, maxAttempts int32) (db.Job, error) {
	f.createEmployeeID = employeeID
	f.createRunType = runType
	f.createAttempts = maxAttempts
	return f.createRow, f.createErr
}

func (f *fakePusakaJobService) Stats(context.Context) (db.GetJobStatsRow, error) {
	return f.statsRow, f.statsErr
}

func (f *fakePusakaJobService) RunAll(_ context.Context, runType string, maxAttempts int32) (int, int, error) {
	f.runAllType = runType
	f.runAllAttempts = maxAttempts
	return f.runAllInserted, f.runAllSkipped, f.runAllErr
}

func (f *fakePusakaJobService) CancelEmployee(_ context.Context, employeeID pgtype.UUID) (int64, error) {
	f.cancelEmployeeID = employeeID
	return f.cancelEmployeeN, f.cancelEmployeeErr
}

func (f *fakePusakaJobService) CancelAll(context.Context) (int64, error) {
	return f.cancelAllN, f.cancelAllErr
}

func TestSettingSuccessHandlersForwardPayloads(t *testing.T) {
	fake := &fakeSettingService{
		Setting:       &service.Setting{},
		listRows:      []db.AppSetting{{Key: "school_name", Value: "MTsN 2"}},
		profile:       service.SchoolProfile{Name: "MTsN 2"},
		updateProfile: service.SchoolProfile{Name: "MTsN 2 Kolaka Utara", Email: "info@example.sch.id"},
	}
	h := &Setting{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/settings", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "school_name") {
		t.Fatalf("List status/body = %d/%s, want settings", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Upsert(rec, withRouteParam(adminRequest(http.MethodPut, "/api/settings/theme", `{"value":"green"}`), "key", "theme"))
	if rec.Code != http.StatusOK {
		t.Fatalf("Upsert status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.upsertKey != "theme" || fake.upsertValue != "green" {
		t.Fatalf("Upsert args = (%q, %q), want theme/green", fake.upsertKey, fake.upsertValue)
	}

	rec = httptest.NewRecorder()
	h.SchoolProfile(rec, httptest.NewRequest(http.MethodGet, "/api/school-profile", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "MTsN 2") {
		t.Fatalf("SchoolProfile status/body = %d/%s, want profile", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdateSchoolProfile(rec, adminRequest(http.MethodPut, "/api/school-profile", `{"name":"MTsN 2 Kolaka Utara","email":"info@example.sch.id"}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateSchoolProfile status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateProfileArg.Name != "MTsN 2 Kolaka Utara" || fake.updateProfileArg.Email != "info@example.sch.id" {
		t.Fatalf("UpdateSchoolProfile arg = %+v, want decoded profile", fake.updateProfileArg)
	}
}

func TestPusakaJobSuccessHandlersForwardPayloads(t *testing.T) {
	employeeID := handlerTestUUID(148)
	jobID := handlerTestUUID(149)
	fake := &fakePusakaJobService{
		PusakaJob:       &service.PusakaJob{},
		listRows:        []db.ListJobsRow{{ID: jobID, EmployeeID: employeeID, Status: db.JobStatusEnumQueued}},
		listTotal:       42,
		createRow:       db.Job{ID: jobID, EmployeeID: employeeID, RunType: db.RunTypeEnumMorning, MaxAttempts: 3},
		statsRow:        db.GetJobStatsRow{Queued: 2, Running: 1},
		runAllInserted:  3,
		runAllSkipped:   1,
		cancelEmployeeN: 2,
		cancelAllN:      5,
	}
	h := &PusakaJob{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/pusaka/jobs?status=queued&page=3&per_page=15", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listStatus != "queued" || fake.listLimit != 15 || fake.listOffset != 30 || !strings.Contains(rec.Body.String(), `"total":42`) {
		t.Fatalf("List args/body = (%q, %d, %d)/%s, want paged result", fake.listStatus, fake.listLimit, fake.listOffset, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Create(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs", `{"employee_id":"`+employeeID.String()+`","run_type":"morning"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createEmployeeID != employeeID || fake.createRunType != "morning" || fake.createAttempts != 3 {
		t.Fatalf("Create args = (%v, %q, %d), want employee/morning/default attempts", fake.createEmployeeID, fake.createRunType, fake.createAttempts)
	}

	rec = httptest.NewRecorder()
	h.Stats(rec, adminRequest(http.MethodGet, "/api/pusaka/jobs/stats", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"queued":2`) {
		t.Fatalf("Stats status/body = %d/%s, want stats", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.RunAll(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs/run-all", `{}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("RunAll status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.runAllType != "morning" || fake.runAllAttempts != 3 {
		t.Fatalf("RunAll args = (%q, %d), want defaults", fake.runAllType, fake.runAllAttempts)
	}

	rec = httptest.NewRecorder()
	h.CancelEmployee(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs/cancel", `{"employee_id":"`+employeeID.String()+`"}`))
	if rec.Code != http.StatusOK || fake.cancelEmployeeID != employeeID {
		t.Fatalf("CancelEmployee status/id = %d/%v, want employee id", rec.Code, fake.cancelEmployeeID)
	}

	rec = httptest.NewRecorder()
	h.CancelAll(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs/cancel-all", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"cancelled":5`) {
		t.Fatalf("CancelAll status/body = %d/%s, want cancelled count", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.SyncAttendance(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs/sync-attendance", ""))
	if rec.Code != http.StatusOK || fake.runAllType != "scrape" || fake.runAllAttempts != 3 {
		t.Fatalf("SyncAttendance status/runAll = %d/%q/%d, want scrape/3", rec.Code, fake.runAllType, fake.runAllAttempts)
	}
}

func TestSettingAndPusakaJobHandlersMapServiceErrors(t *testing.T) {
	errDB := errors.New("db down")
	employeeID := handlerTestUUID(150)
	tests := []struct {
		name       string
		run        func(*httptest.ResponseRecorder)
		wantStatus int
	}{
		{name: "setting list", run: func(rec *httptest.ResponseRecorder) {
			(&Setting{svc: &fakeSettingService{Setting: &service.Setting{}, listErr: errDB}}).List(rec, adminRequest(http.MethodGet, "/api/settings", ""))
		}, wantStatus: http.StatusInternalServerError},
		{name: "setting upsert", run: func(rec *httptest.ResponseRecorder) {
			(&Setting{svc: &fakeSettingService{Setting: &service.Setting{}, upsertErr: errDB}}).Upsert(rec, withRouteParam(adminRequest(http.MethodPut, "/api/settings/theme", `{"value":"green"}`), "key", "theme"))
		}, wantStatus: http.StatusInternalServerError},
		{name: "profile public", run: func(rec *httptest.ResponseRecorder) {
			(&Setting{svc: &fakeSettingService{Setting: &service.Setting{}, profileErr: errDB}}).SchoolProfile(rec, httptest.NewRequest(http.MethodGet, "/api/school-profile", nil))
		}, wantStatus: http.StatusInternalServerError},
		{name: "profile update", run: func(rec *httptest.ResponseRecorder) {
			(&Setting{svc: &fakeSettingService{Setting: &service.Setting{}, updateProfileErr: errors.New("nama sekolah wajib diisi")}}).UpdateSchoolProfile(rec, adminRequest(http.MethodPut, "/api/school-profile", `{}`))
		}, wantStatus: http.StatusBadRequest},
		{name: "job list", run: func(rec *httptest.ResponseRecorder) {
			(&PusakaJob{svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, listErr: errDB}}).List(rec, adminRequest(http.MethodGet, "/api/pusaka/jobs", ""))
		}, wantStatus: http.StatusInternalServerError},
		{name: "job create conflict", run: func(rec *httptest.ResponseRecorder) {
			(&PusakaJob{svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, createErr: domain.ErrConflict}}).Create(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs", `{"employee_id":"`+employeeID.String()+`","run_type":"morning"}`))
		}, wantStatus: http.StatusConflict},
		{name: "job stats", run: func(rec *httptest.ResponseRecorder) {
			(&PusakaJob{svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, statsErr: errDB}}).Stats(rec, adminRequest(http.MethodGet, "/api/pusaka/jobs/stats", ""))
		}, wantStatus: http.StatusInternalServerError},
		{name: "job run all", run: func(rec *httptest.ResponseRecorder) {
			(&PusakaJob{svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, runAllErr: errDB}}).RunAll(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs/run-all", `{}`))
		}, wantStatus: http.StatusInternalServerError},
		{name: "job cancel employee", run: func(rec *httptest.ResponseRecorder) {
			(&PusakaJob{svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, cancelEmployeeErr: errDB}}).CancelEmployee(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs/cancel", `{"employee_id":"`+employeeID.String()+`"}`))
		}, wantStatus: http.StatusInternalServerError},
		{name: "job cancel all", run: func(rec *httptest.ResponseRecorder) {
			(&PusakaJob{svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, cancelAllErr: errDB}}).CancelAll(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs/cancel-all", ""))
		}, wantStatus: http.StatusInternalServerError},
		{name: "job sync attendance", run: func(rec *httptest.ResponseRecorder) {
			(&PusakaJob{svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, runAllErr: errDB}}).SyncAttendance(rec, adminRequest(http.MethodPost, "/api/pusaka/jobs/sync-attendance", ""))
		}, wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.run(rec)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestPusakaJobValidationBranches(t *testing.T) {
	employeeID := handlerTestUUID(151)
	errDB := errors.New("db down")
	plainRequest := func(method, target, body string) *http.Request {
		return httptest.NewRequest(method, target, strings.NewReader(body))
	}

	tests := []struct {
		name       string
		handler    func(*PusakaJob, http.ResponseWriter, *http.Request)
		svc        *fakePusakaJobService
		req        *http.Request
		wantStatus int
	}{
		{name: "create forbidden", handler: (*PusakaJob).Create, req: plainRequest(http.MethodPost, "/api/pusaka/jobs", `{}`), wantStatus: http.StatusForbidden},
		{name: "create invalid json", handler: (*PusakaJob).Create, req: adminRequest(http.MethodPost, "/api/pusaka/jobs", `{`), wantStatus: http.StatusBadRequest},
		{name: "create invalid employee", handler: (*PusakaJob).Create, req: adminRequest(http.MethodPost, "/api/pusaka/jobs", `{"employee_id":"bad"}`), wantStatus: http.StatusBadRequest},
		{name: "create internal error", handler: (*PusakaJob).Create, svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, createErr: errDB}, req: adminRequest(http.MethodPost, "/api/pusaka/jobs", `{"employee_id":"`+employeeID.String()+`","run_type":"morning","max_attempts":2}`), wantStatus: http.StatusInternalServerError},
		{name: "stats forbidden", handler: (*PusakaJob).Stats, req: plainRequest(http.MethodGet, "/api/pusaka/jobs/stats", ""), wantStatus: http.StatusForbidden},
		{name: "run all forbidden", handler: (*PusakaJob).RunAll, req: plainRequest(http.MethodPost, "/api/pusaka/jobs/run-all", `{}`), wantStatus: http.StatusForbidden},
		{name: "run all invalid json", handler: (*PusakaJob).RunAll, req: adminRequest(http.MethodPost, "/api/pusaka/jobs/run-all", `{`), wantStatus: http.StatusBadRequest},
		{name: "run all custom payload", handler: (*PusakaJob).RunAll, svc: &fakePusakaJobService{PusakaJob: &service.PusakaJob{}, runAllInserted: 1}, req: adminRequest(http.MethodPost, "/api/pusaka/jobs/run-all", `{"run_type":"afternoon","max_attempts":5}`), wantStatus: http.StatusCreated},
		{name: "cancel employee forbidden", handler: (*PusakaJob).CancelEmployee, req: plainRequest(http.MethodPost, "/api/pusaka/jobs/cancel", `{}`), wantStatus: http.StatusForbidden},
		{name: "cancel employee invalid json", handler: (*PusakaJob).CancelEmployee, req: adminRequest(http.MethodPost, "/api/pusaka/jobs/cancel", `{`), wantStatus: http.StatusBadRequest},
		{name: "cancel employee invalid id", handler: (*PusakaJob).CancelEmployee, req: adminRequest(http.MethodPost, "/api/pusaka/jobs/cancel", `{"employee_id":"bad"}`), wantStatus: http.StatusBadRequest},
		{name: "cancel all forbidden", handler: (*PusakaJob).CancelAll, req: plainRequest(http.MethodPost, "/api/pusaka/jobs/cancel-all", ""), wantStatus: http.StatusForbidden},
		{name: "sync attendance forbidden", handler: (*PusakaJob).SyncAttendance, req: plainRequest(http.MethodPost, "/api/pusaka/jobs/sync-attendance", ""), wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakePusakaJobService{PusakaJob: &service.PusakaJob{}}
			}
			rec := httptest.NewRecorder()
			tt.handler(&PusakaJob{svc: svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.name == "run all custom payload" && (svc.runAllType != "afternoon" || svc.runAllAttempts != 5) {
				t.Fatalf("RunAll custom args = %q/%d, want afternoon/5", svc.runAllType, svc.runAllAttempts)
			}
		})
	}
}
