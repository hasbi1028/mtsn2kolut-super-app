package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAcademicCurriculumLessonYearAndImportEndpoints(t *testing.T) {
	profileID := handlerTestUUID(51)
	yearID := handlerTestUUID(52)
	lessonID := handlerTestUUID(53)
	fake := &fakeAcademicService{}
	h := &Academic{svc: fake}

	rec := httptest.NewRecorder()
	h.GetCurriculumOverview(rec, adminRequest(http.MethodGet, "/api/academic/curriculum/overview?profile_id="+profileID.String()+"&level=vii", ""))
	if rec.Code != http.StatusOK || fake.curriculumOverviewProfileID != profileID || fake.curriculumOverviewLevel != "VII" || !strings.Contains(rec.Body.String(), "Kurikulum Merdeka") {
		t.Fatalf("GetCurriculumOverview() status/args/body = %d/%v/%q/%s, want mapped overview", rec.Code, fake.curriculumOverviewProfileID, fake.curriculumOverviewLevel, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListCurriculumProfiles(rec, adminRequest(http.MethodGet, "/api/academic/curriculum/profiles", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Kurikulum Merdeka") {
		t.Fatalf("ListCurriculumProfiles() status/body = %d/%s, want profiles", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListCurriculumAllocations(rec, adminRequest(http.MethodGet, "/api/academic/curriculum/allocations?profile_id="+profileID.String()+"&level=ix", ""))
	if rec.Code != http.StatusOK || fake.curriculumAllocProfileID != profileID || fake.curriculumAllocLevel != "IX" || !strings.Contains(rec.Body.String(), "total_weekly_hours") {
		t.Fatalf("ListCurriculumAllocations() status/args/body = %d/%v/%q/%s, want mapped allocations", rec.Code, fake.curriculumAllocProfileID, fake.curriculumAllocLevel, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetCurriculumSummary(rec, adminRequest(http.MethodGet, "/api/academic/curriculum/summary?profile_id="+profileID.String(), ""))
	if rec.Code != http.StatusOK || fake.curriculumSummaryProfileID != profileID || !strings.Contains(rec.Body.String(), "status_label") {
		t.Fatalf("GetCurriculumSummary() status/profile/body = %d/%v/%s, want summary", rec.Code, fake.curriculumSummaryProfileID, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetLessonPeriods(rec, adminRequest(http.MethodGet, "/api/academic/lesson-periods", ""))
	if rec.Code != http.StatusOK || !fake.lessonCalled {
		t.Fatalf("GetLessonPeriods() status/called = %d/%v, want 200/true", rec.Code, fake.lessonCalled)
	}

	lessonBody := `{"academic_year_id":"` + yearID.String() + `","day_of_week":2,"period_number":1,"start_time":"07:00","end_time":"07:40","activity_type":" lesson ","label":" Jam 1 ","is_counted_as_lesson":false}`
	rec = httptest.NewRecorder()
	h.CreateLessonPeriod(rec, adminRequest(http.MethodPost, "/api/academic/lesson-periods", lessonBody))
	if rec.Code != http.StatusCreated || fake.createLessonArg.AcademicYearID != yearID || fake.createLessonArg.ActivityType != "lesson" || fake.createLessonArg.Label != "Jam 1" || fake.createLessonArg.IsCountedAsLesson {
		t.Fatalf("CreateLessonPeriod() status/arg = %d/%+v, want mapped trimmed lesson", rec.Code, fake.createLessonArg)
	}

	rec = httptest.NewRecorder()
	h.UpdateLessonPeriod(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/academic/lesson-periods/"+lessonID.String(), lessonBody), "id", lessonID.String()))
	if rec.Code != http.StatusOK || fake.updateLessonArg.ID != lessonID || fake.updateLessonArg.ActivityType != "lesson" {
		t.Fatalf("UpdateLessonPeriod() status/arg = %d/%+v, want mapped update", rec.Code, fake.updateLessonArg)
	}

	rec = httptest.NewRecorder()
	h.DeleteLessonPeriod(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/academic/lesson-periods/"+lessonID.String(), ""), "id", lessonID.String()))
	if rec.Code != http.StatusNoContent || fake.deleteLessonID != lessonID {
		t.Fatalf("DeleteLessonPeriod() status/id = %d/%v, want 204/%v", rec.Code, fake.deleteLessonID, lessonID)
	}

	rec = httptest.NewRecorder()
	h.ActivateYear(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/years/"+yearID.String()+"/activate", `{"confirmation":"AKTIFKAN"}`), "id", yearID.String()))
	if rec.Code != http.StatusOK || fake.activateYearID != yearID || fake.activateConfirm != "AKTIFKAN" {
		t.Fatalf("ActivateYear() status/id/confirm = %d/%v/%q, want mapped activation", rec.Code, fake.activateYearID, fake.activateConfirm)
	}

	rec = httptest.NewRecorder()
	h.PreviewYearRollover(rec, adminRequest(http.MethodPost, "/api/academic/year-rollover/preview", `{"source_academic_year_id":"`+yearID.String()+`","target_academic_year_id":"`+profileID.String()+`"}`))
	if rec.Code != http.StatusOK || fake.previewInput.SourceAcademicYearID != yearID || fake.previewInput.TargetAcademicYearID != profileID {
		t.Fatalf("PreviewYearRollover() status/input = %d/%+v, want mapped preview", rec.Code, fake.previewInput)
	}

	rec = httptest.NewRecorder()
	h.ApplyYearRollover(rec, adminRequest(http.MethodPost, "/api/academic/year-rollover/apply", `{"source_academic_year_id":"`+yearID.String()+`","target_academic_year_id":"`+profileID.String()+`","confirmation":"LANJUT","safety_token":"token"}`))
	if rec.Code != http.StatusOK || fake.applyInput.SourceAcademicYearID != yearID || fake.applyInput.TargetAcademicYearID != profileID || fake.applyInput.SafetyToken != "token" {
		t.Fatalf("ApplyYearRollover() status/input = %d/%+v, want mapped apply", rec.Code, fake.applyInput)
	}

	rec = httptest.NewRecorder()
	h.DryRunAcademicImport(rec, adminRequest(http.MethodPost, "/api/academic/import/dry-run", `{"kind":"subjects","rows":[{"code":"IPA"}]}`))
	if rec.Code != http.StatusOK || fake.dryRunInput.Kind != "subjects" || !strings.Contains(rec.Body.String(), "add_count") {
		t.Fatalf("DryRunAcademicImport() status/input/body = %d/%+v/%s, want dry-run result", rec.Code, fake.dryRunInput, rec.Body.String())
	}
}

func TestAcademicAdditionalEndpointValidationAndErrorPaths(t *testing.T) {
	profileID := handlerTestUUID(61)
	yearID := handlerTestUUID(62)
	lessonID := handlerTestUUID(63)
	errDB := errors.New("db down")
	plain := func(method, target, body string) *http.Request {
		return httptest.NewRequest(method, target, strings.NewReader(body))
	}
	tests := []struct {
		name string
		fn   func(*Academic, http.ResponseWriter, *http.Request)
		svc  *fakeAcademicService
		req  *http.Request
		want int
	}{
		{name: "curriculum overview forbidden", fn: (*Academic).GetCurriculumOverview, req: plain(http.MethodGet, "/api/academic/curriculum/overview", ""), want: http.StatusForbidden},
		{name: "curriculum overview bad level", fn: (*Academic).GetCurriculumOverview, req: adminRequest(http.MethodGet, "/api/academic/curriculum/overview?level=X", ""), want: http.StatusBadRequest},
		{name: "curriculum overview service error", fn: (*Academic).GetCurriculumOverview, svc: &fakeAcademicService{curriculumErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic/curriculum/overview", ""), want: http.StatusInternalServerError},
		{name: "curriculum profiles service error", fn: (*Academic).ListCurriculumProfiles, svc: &fakeAcademicService{curriculumErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic/curriculum/profiles", ""), want: http.StatusInternalServerError},
		{name: "curriculum allocations missing profile", fn: (*Academic).ListCurriculumAllocations, req: adminRequest(http.MethodGet, "/api/academic/curriculum/allocations", ""), want: http.StatusBadRequest},
		{name: "curriculum allocations bad level", fn: (*Academic).ListCurriculumAllocations, req: adminRequest(http.MethodGet, "/api/academic/curriculum/allocations?profile_id="+profileID.String()+"&level=X", ""), want: http.StatusBadRequest},
		{name: "curriculum summary bad profile", fn: (*Academic).GetCurriculumSummary, req: adminRequest(http.MethodGet, "/api/academic/curriculum/summary?profile_id=bad", ""), want: http.StatusBadRequest},
		{name: "lesson periods forbidden", fn: (*Academic).GetLessonPeriods, req: plain(http.MethodGet, "/api/academic/lesson-periods", ""), want: http.StatusForbidden},
		{name: "lesson periods service error", fn: (*Academic).GetLessonPeriods, svc: &fakeAcademicService{lessonErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic/lesson-periods", ""), want: http.StatusInternalServerError},
		{name: "create lesson forbidden", fn: (*Academic).CreateLessonPeriod, req: plain(http.MethodPost, "/api/academic/lesson-periods", `{}`), want: http.StatusForbidden},
		{name: "create lesson invalid year", fn: (*Academic).CreateLessonPeriod, req: adminRequest(http.MethodPost, "/api/academic/lesson-periods", `{"academic_year_id":"bad"}`), want: http.StatusBadRequest},
		{name: "create lesson invalid start", fn: (*Academic).CreateLessonPeriod, req: adminRequest(http.MethodPost, "/api/academic/lesson-periods", `{"academic_year_id":"`+yearID.String()+`","start_time":"bad","end_time":"08:00"}`), want: http.StatusBadRequest},
		{name: "create lesson invalid end", fn: (*Academic).CreateLessonPeriod, req: adminRequest(http.MethodPost, "/api/academic/lesson-periods", `{"academic_year_id":"`+yearID.String()+`","start_time":"07:00","end_time":"bad"}`), want: http.StatusBadRequest},
		{name: "create lesson service error", fn: (*Academic).CreateLessonPeriod, svc: &fakeAcademicService{createErr: errDB}, req: adminRequest(http.MethodPost, "/api/academic/lesson-periods", `{"academic_year_id":"`+yearID.String()+`","start_time":"07:00","end_time":"08:00"}`), want: http.StatusBadRequest},
		{name: "update lesson bad id", fn: (*Academic).UpdateLessonPeriod, req: withRouteParam(adminRequest(http.MethodPatch, "/api/academic/lesson-periods/bad", `{}`), "id", "bad"), want: http.StatusBadRequest},
		{name: "update lesson invalid json", fn: (*Academic).UpdateLessonPeriod, req: withRouteParam(adminRequest(http.MethodPatch, "/api/academic/lesson-periods/"+lessonID.String(), `{`), "id", lessonID.String()), want: http.StatusBadRequest},
		{name: "delete lesson bad id", fn: (*Academic).DeleteLessonPeriod, req: withRouteParam(adminRequest(http.MethodDelete, "/api/academic/lesson-periods/bad", ""), "id", "bad"), want: http.StatusBadRequest},
		{name: "delete lesson service error", fn: (*Academic).DeleteLessonPeriod, svc: &fakeAcademicService{deleteErr: errDB}, req: withRouteParam(adminRequest(http.MethodDelete, "/api/academic/lesson-periods/"+lessonID.String(), ""), "id", lessonID.String()), want: http.StatusBadRequest},
		{name: "activate bad id", fn: (*Academic).ActivateYear, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/years/bad/activate", `{}`), "id", "bad"), want: http.StatusBadRequest},
		{name: "activate invalid json", fn: (*Academic).ActivateYear, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/years/"+yearID.String()+"/activate", `{`), "id", yearID.String()), want: http.StatusBadRequest},
		{name: "activate service error", fn: (*Academic).ActivateYear, svc: &fakeAcademicService{activateErr: errDB}, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/years/"+yearID.String()+"/activate", `{"confirmation":"x"}`), "id", yearID.String()), want: http.StatusInternalServerError},
		{name: "preview invalid target", fn: (*Academic).PreviewYearRollover, req: adminRequest(http.MethodPost, "/api/academic/year-rollover/preview", `{"target_academic_year_id":"bad"}`), want: http.StatusBadRequest},
		{name: "apply invalid source", fn: (*Academic).ApplyYearRollover, req: adminRequest(http.MethodPost, "/api/academic/year-rollover/apply", `{"source_academic_year_id":"bad","target_academic_year_id":"`+yearID.String()+`"}`), want: http.StatusBadRequest},
		{name: "dry-run service error", fn: (*Academic).DryRunAcademicImport, svc: &fakeAcademicService{dryRunErr: errDB}, req: adminRequest(http.MethodPost, "/api/academic/import/dry-run", `{"kind":"subjects"}`), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakeAcademicService{}
			}
			rec := httptest.NewRecorder()
			tt.fn(&Academic{svc: svc}, rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
