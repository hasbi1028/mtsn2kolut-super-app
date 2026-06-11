package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/api"
)

func withClaims(req *http.Request, claims jwt.MapClaims) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, claims))
}

func withRouteParam(req *http.Request, key, value string) *http.Request {
	chiCtx := chi.RouteContext(req.Context())
	if chiCtx == nil {
		chiCtx = chi.NewRouteContext()
	}
	chiCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
}

func TestGovernanceStatsForbiddenWithoutStaffAccess(t *testing.T) {
	h := NewGovernance(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/governance/stats", nil)
	rec := httptest.NewRecorder()

	h.Stats(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestGovernanceDeleteUnitForbiddenForGuruRole(t *testing.T) {
	h := NewGovernance(nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/governance/units/11111111-1111-1111-1111-111111111111", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.DeleteUnit(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionGetForbiddenForNonCbtRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"staf"}})
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionScoreForbiddenForNonCbtRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/score", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"kesiswaan"}})
	rec := httptest.NewRecorder()

	h.ScoreSession(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionGuruAwareListForbiddenForNonCbtRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/cbt/sessions", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"staf"}})
	rec := httptest.NewRecorder()

	h.GuruAwareList(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionGuruAwareResultsForbiddenForNonCbtRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/results", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"staf"}})
	rec := httptest.NewRecorder()

	h.GuruAwareResults(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionGuruAwareParticipantsForbiddenForNonCbtRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/participants", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"staf"}})
	rec := httptest.NewRecorder()

	h.GuruAwareParticipants(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionGenerateTokensForbiddenForGuruRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/generate-tokens", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}, "eid": "22222222-2222-2222-2222-222222222222"})
	rec := httptest.NewRecorder()

	h.GenerateTokens(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionCreateRoomForbiddenForGuruRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/rooms", strings.NewReader(`{"room_name":"Ruang A","capacity":20}`))
	req.Header.Set("Content-Type", "application/json")
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}, "eid": "22222222-2222-2222-2222-222222222222"})
	rec := httptest.NewRecorder()

	h.CreateRoom(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionRecordAnswerForbiddenForGuruRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/participants/22222222-2222-2222-2222-222222222222/answer", strings.NewReader(`{"question_id":"33333333-3333-3333-3333-333333333333","answer":"A"}`))
	req.Header.Set("Content-Type", "application/json")
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "11111111-1111-1111-1111-111111111111")
	chiCtx.URLParams.Add("pid", "22222222-2222-2222-2222-222222222222")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}, "eid": "44444444-4444-4444-4444-444444444444"})
	rec := httptest.NewRecorder()

	h.RecordAnswer(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPortalStudentMeForbiddenForGuruClaimWithStudentID(t *testing.T) {
	h := NewPortal(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/portal/student/me", nil)
	req = withClaims(req, jwt.MapClaims{
		"role": "guru",
		"sid":  "11111111-1111-1111-1111-111111111111",
	})
	rec := httptest.NewRecorder()

	h.StudentMe(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPortalTeacherTimetableForbiddenForStudentClaimWithEmployeeID(t *testing.T) {
	h := NewPortal(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/portal/guru/timetable", nil)
	req = withClaims(req, jwt.MapClaims{
		"role": "siswa",
		"eid":  "11111111-1111-1111-1111-111111111111",
	})
	rec := httptest.NewRecorder()

	h.TeacherTimetable(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPortalParentMeForbiddenForStudentClaimWithParentID(t *testing.T) {
	h := NewPortal(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/portal/parent/me", nil)
	req = withClaims(req, jwt.MapClaims{
		"role": "siswa",
		"pid":  "11111111-1111-1111-1111-111111111111",
	})
	rec := httptest.NewRecorder()

	h.ParentMe(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestParentListForbiddenForGuruRole(t *testing.T) {
	h := NewParent(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/parents", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestParentGetForbiddenForParentRole(t *testing.T) {
	h := NewParent(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/parents/11111111-1111-1111-1111-111111111111", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"role": "ortu"})
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestParentListChildrenForbiddenForGuruRole(t *testing.T) {
	h := NewParent(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/parents/11111111-1111-1111-1111-111111111111/children", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.ListChildren(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestStudentGuruAwareListForbiddenForStaffRole(t *testing.T) {
	h := NewStudent(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/students", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"staf"}})
	rec := httptest.NewRecorder()

	h.GuruAwareList(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestStudentListForbiddenForGuruRole(t *testing.T) {
	h := NewStudent(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/students", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestStudentCreateForbiddenForGuruRole(t *testing.T) {
	h := NewStudent(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/students", bytes.NewBufferString(`{"nama":"Siswa","nis":"123"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestAcademicCreateForbiddenForGuruRole(t *testing.T) {
	h := NewAcademic(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/academic/years", bytes.NewBufferString(`{"name":"2026/2027","start_date":"2026-07-01","end_date":"2027-06-30"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withRouteParam(req, "entity", "years")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestEmployeeListForbiddenForGuruRole(t *testing.T) {
	h := NewEmployee(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/employees", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestUserListForbiddenForGuruRole(t *testing.T) {
	h := NewUser(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtEventCreateForbiddenForGuruRole(t *testing.T) {
	h := NewCbtEvent(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/cbt/events", bytes.NewBufferString(`{"title":"UTS","exam_type":"uts"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPusakaSettingListForbiddenForGuruRole(t *testing.T) {
	h := NewSetting(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/pusaka/settings", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestSchoolProfileUpdateForbiddenForGuruRole(t *testing.T) {
	h := NewSetting(nil)
	req := httptest.NewRequest(http.MethodPut, "/api/school-profile", bytes.NewBufferString(`{"school_name":"MTsN 2"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.UpdateSchoolProfile(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestWebsiteAdminListForbiddenForGuruRole(t *testing.T) {
	h := NewWebsite(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/website/content", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestWebsiteMediaUploadForbiddenForGuruRole(t *testing.T) {
	h := NewWebsiteMedia("/tmp")
	req := httptest.NewRequest(http.MethodPost, "/api/website/media", bytes.NewBufferString("not-a-multipart"))
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.Upload(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPusakaJobListForbiddenForGuruRole(t *testing.T) {
	h := NewPusakaJob(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/pusaka/jobs", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPusakaScheduleListForbiddenForGuruRole(t *testing.T) {
	h := NewPusakaSchedule(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/pusaka/schedules", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestEmployeeScheduleListForbiddenForGuruRole(t *testing.T) {
	h := NewEmployeeSchedule(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/pusaka/employees/11111111-1111-1111-1111-111111111111/schedules", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPusakaWorkerStatusForbiddenForGuruRole(t *testing.T) {
	h := NewPusakaWorker(nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/pusaka/worker/status", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.GetStatus(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPusakaSchedulerTickForbiddenForGuruRole(t *testing.T) {
	h := NewPusakaScheduler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/pusaka/scheduler/tick", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.Tick(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPusakaAttendanceListForbiddenForGuruRole(t *testing.T) {
	h := NewPusakaAttendance(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/pusaka/attendance", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionCreateForbiddenForGuruRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/cbt/sessions", bytes.NewBufferString(`{"package_id":"01000000-0000-0000-0000-000000000000"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCbtSessionUpdateStatusForbiddenForGuruRole(t *testing.T) {
	h := NewCbtSession(nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/cbt/sessions/11111111-1111-1111-1111-111111111111/status", bytes.NewBufferString(`{"status":"published"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.UpdateStatus(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestJournalDeleteSessionForbiddenForGuruRole(t *testing.T) {
	h := NewClassJournal(nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/journal/sessions/11111111-1111-1111-1111-111111111111", nil)
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}})
	rec := httptest.NewRecorder()

	h.DeleteSession(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestSettingUpsertBlocksProtectedAuthKeys(t *testing.T) {
	h := NewSetting(nil)
	req := httptest.NewRequest(http.MethodPut, "/api/settings/admin_password", bytes.NewBufferString(`{"value":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	req = withRouteParam(req, "key", "admin_password")
	rec := httptest.NewRecorder()

	h.Upsert(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestArchiveListDocumentsRejectsInvalidCategoryFilter(t *testing.T) {
	h := NewArchive(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/arsip/dokumen?category_id=not-a-uuid", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"staf"}})
	rec := httptest.NewRecorder()

	h.ListDocuments(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "id tidak valid") {
		t.Fatalf("body = %s, want safe archive validation message", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "not-a-uuid") {
		t.Fatalf("body = %s, should not expose raw invalid uuid input", rec.Body.String())
	}
}

func TestDocumentCycleUpdateRejectsInvalidArchiveDocumentID(t *testing.T) {
	h := NewDocumentCycle(nil)
	req := httptest.NewRequest(http.MethodPut, "/api/document-cycles/obligations/11111111-1111-1111-1111-111111111111", bytes.NewBufferString(`{
		"code":"DOC-1",
		"title":"Obligation",
		"period_year":2026,
		"due_date":"2026-06-01",
		"reminder_date":"2026-05-25",
		"archive_document_id":"not-a-uuid"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}})
	req = withRouteParam(req, "id", "11111111-1111-1111-1111-111111111111")
	rec := httptest.NewRecorder()

	h.UpdateObligation(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "id tidak valid") {
		t.Fatalf("body = %s, want safe document-cycle validation message", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "not-a-uuid") {
		t.Fatalf("body = %s, should not expose raw invalid uuid input", rec.Body.String())
	}
}
