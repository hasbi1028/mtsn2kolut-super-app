package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mtsn2kolut-super-app/backend/internal/domain"
)

func TestAcademicMoreReadHandlersLowBranches(t *testing.T) {
	dbErr := errors.New("db down")
	tests := []struct {
		name       string
		fn         func(*Academic, http.ResponseWriter, *http.Request)
		svc        *fakeAcademicService
		req        *http.Request
		wantStatus int
		wantCalled func(*fakeAcademicService) bool
	}{
		{
			name:       "dashboard forbidden skips service",
			fn:         (*Academic).GetDashboard,
			svc:        &fakeAcademicService{},
			req:        httptest.NewRequest(http.MethodGet, "/api/academic/dashboard", nil),
			wantStatus: http.StatusForbidden,
			wantCalled: func(f *fakeAcademicService) bool { return !f.dashboardCalled },
		},
		{
			name:       "dashboard internal",
			fn:         (*Academic).GetDashboard,
			svc:        &fakeAcademicService{dashboardErr: dbErr},
			req:        adminRequest(http.MethodGet, "/api/academic/dashboard", ""),
			wantStatus: http.StatusInternalServerError,
			wantCalled: func(f *fakeAcademicService) bool { return f.dashboardCalled },
		},
		{
			name:       "readiness forbidden skips service",
			fn:         (*Academic).GetReadiness,
			svc:        &fakeAcademicService{},
			req:        httptest.NewRequest(http.MethodGet, "/api/academic/readiness", nil),
			wantStatus: http.StatusForbidden,
			wantCalled: func(f *fakeAcademicService) bool { return !f.readinessCalled },
		},
		{
			name:       "readiness internal",
			fn:         (*Academic).GetReadiness,
			svc:        &fakeAcademicService{readinessErr: dbErr},
			req:        adminRequest(http.MethodGet, "/api/academic/readiness", ""),
			wantStatus: http.StatusInternalServerError,
			wantCalled: func(f *fakeAcademicService) bool { return f.readinessCalled },
		},
		{
			name:       "curriculum summary forbidden skips service",
			fn:         (*Academic).GetCurriculumSummary,
			svc:        &fakeAcademicService{},
			req:        httptest.NewRequest(http.MethodGet, "/api/academic/curriculum/summary?profile_id="+handlerTestUUID(51).String(), nil),
			wantStatus: http.StatusForbidden,
			wantCalled: func(f *fakeAcademicService) bool { return !f.curriculumSummaryProfileID.Valid },
		},
		{
			name:       "curriculum summary invalid profile",
			fn:         (*Academic).GetCurriculumSummary,
			svc:        &fakeAcademicService{},
			req:        adminRequest(http.MethodGet, "/api/academic/curriculum/summary?profile_id=bad", ""),
			wantStatus: http.StatusBadRequest,
			wantCalled: func(f *fakeAcademicService) bool { return !f.curriculumSummaryProfileID.Valid },
		},
		{
			name:       "curriculum summary internal",
			fn:         (*Academic).GetCurriculumSummary,
			svc:        &fakeAcademicService{curriculumErr: dbErr},
			req:        adminRequest(http.MethodGet, "/api/academic/curriculum/summary?profile_id="+handlerTestUUID(52).String(), ""),
			wantStatus: http.StatusInternalServerError,
			wantCalled: func(f *fakeAcademicService) bool { return f.curriculumSummaryProfileID.Valid },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(&Academic{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if !tt.wantCalled(tt.svc) {
				t.Fatalf("service call state after %s did not match expectation", tt.name)
			}
		})
	}
}

func TestAcademicMoreYearRolloverLowBranches(t *testing.T) {
	sourceID := handlerTestUUID(53)
	targetID := handlerTestUUID(54)
	domainErr := errors.Join(domain.ErrBadRequest, errors.New("confirmation mismatch"))

	tests := []struct {
		name       string
		fn         func(*Academic, http.ResponseWriter, *http.Request)
		svc        *fakeAcademicService
		req        *http.Request
		wantStatus int
		wantBody   string
		check      func(*testing.T, *fakeAcademicService)
	}{
		{
			name:       "preview forbidden skips service",
			fn:         (*Academic).PreviewYearRollover,
			svc:        &fakeAcademicService{},
			req:        httptest.NewRequest(http.MethodPost, "/api/academic/year-rollover/preview", strings.NewReader(`{}`)),
			wantStatus: http.StatusForbidden,
			check: func(t *testing.T, f *fakeAcademicService) {
				t.Helper()
				if f.previewInput.TargetAcademicYearID.Valid {
					t.Fatalf("PreviewYearRollover called service on forbidden request: %+v", f.previewInput)
				}
			},
		},
		{
			name:       "preview invalid json",
			fn:         (*Academic).PreviewYearRollover,
			svc:        &fakeAcademicService{},
			req:        adminRequest(http.MethodPost, "/api/academic/year-rollover/preview", `{`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "preview invalid optional source",
			fn:         (*Academic).PreviewYearRollover,
			svc:        &fakeAcademicService{},
			req:        adminRequest(http.MethodPost, "/api/academic/year-rollover/preview", `{"source_academic_year_id":"bad","target_academic_year_id":"`+targetID.String()+`"}`),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Tahun ajaran sumber tidak valid",
		},
		{
			name:       "preview invalid target",
			fn:         (*Academic).PreviewYearRollover,
			svc:        &fakeAcademicService{},
			req:        adminRequest(http.MethodPost, "/api/academic/year-rollover/preview", `{"source_academic_year_id":"`+sourceID.String()+`","target_academic_year_id":"bad"}`),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Tahun ajaran tujuan tidak valid",
		},
		{
			name:       "preview domain error",
			fn:         (*Academic).PreviewYearRollover,
			svc:        &fakeAcademicService{previewErr: domainErr},
			req:        adminRequest(http.MethodPost, "/api/academic/year-rollover/preview", `{"target_academic_year_id":"`+targetID.String()+`"}`),
			wantStatus: http.StatusBadRequest,
			check: func(t *testing.T, f *fakeAcademicService) {
				t.Helper()
				if f.previewInput.SourceAcademicYearID.Valid || f.previewInput.TargetAcademicYearID != targetID {
					t.Fatalf("preview input = %+v, want empty optional source and target %v", f.previewInput, targetID)
				}
			},
		},
		{
			name:       "apply forbidden skips service",
			fn:         (*Academic).ApplyYearRollover,
			svc:        &fakeAcademicService{},
			req:        httptest.NewRequest(http.MethodPost, "/api/academic/year-rollover/apply", strings.NewReader(`{}`)),
			wantStatus: http.StatusForbidden,
			check: func(t *testing.T, f *fakeAcademicService) {
				t.Helper()
				if f.applyInput.TargetAcademicYearID.Valid {
					t.Fatalf("ApplyYearRollover called service on forbidden request: %+v", f.applyInput)
				}
			},
		},
		{
			name:       "apply invalid json",
			fn:         (*Academic).ApplyYearRollover,
			svc:        &fakeAcademicService{},
			req:        adminRequest(http.MethodPost, "/api/academic/year-rollover/apply", `{`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "apply invalid source",
			fn:         (*Academic).ApplyYearRollover,
			svc:        &fakeAcademicService{},
			req:        adminRequest(http.MethodPost, "/api/academic/year-rollover/apply", `{"source_academic_year_id":"bad","target_academic_year_id":"`+targetID.String()+`"}`),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Tahun ajaran sumber tidak valid",
		},
		{
			name:       "apply invalid target",
			fn:         (*Academic).ApplyYearRollover,
			svc:        &fakeAcademicService{},
			req:        adminRequest(http.MethodPost, "/api/academic/year-rollover/apply", `{"source_academic_year_id":"`+sourceID.String()+`","target_academic_year_id":"bad"}`),
			wantStatus: http.StatusBadRequest,
			wantBody:   "Tahun ajaran tujuan tidak valid",
		},
		{
			name:       "apply forwards confirmation safety token and maps domain error",
			fn:         (*Academic).ApplyYearRollover,
			svc:        &fakeAcademicService{previewErr: domainErr},
			req:        adminRequest(http.MethodPost, "/api/academic/year-rollover/apply", `{"source_academic_year_id":"`+sourceID.String()+`","target_academic_year_id":"`+targetID.String()+`","confirmation":"APPLY","safety_token":"token-123"}`),
			wantStatus: http.StatusBadRequest,
			check: func(t *testing.T, f *fakeAcademicService) {
				t.Helper()
				if f.applyInput.SourceAcademicYearID != sourceID || f.applyInput.TargetAcademicYearID != targetID || f.applyInput.Confirmation != "APPLY" || f.applyInput.SafetyToken != "token-123" {
					t.Fatalf("apply input = %+v, want mapped ids/confirmation/token", f.applyInput)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(&Academic{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want to contain %q", rec.Body.String(), tt.wantBody)
			}
			if tt.check != nil {
				tt.check(t, tt.svc)
			}
		})
	}
}
