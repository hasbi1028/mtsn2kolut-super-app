package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"mtsn2kolut-super-app/backend/internal/domain"
)

func TestNonTestAssessmentMoreGetUpdateDeleteLowBranches(t *testing.T) {
	assessmentID := handlerTestUUID(161)
	subjectID := handlerTestUUID(162)
	classID := handlerTestUUID(163)
	teacherID := "33333333-3333-3333-3333-333333333333"

	t.Run("get not found stops before ownership check", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{getErr: domain.ErrNotFound, teacherOwnsAssessment: true}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(nonTestAssessmentAdminRequest(http.MethodGet, "/", ""), "id", pgUUIDString(assessmentID))

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Get() status = %d, want 404; body=%s", rec.Code, rec.Body.String())
		}
		if fake.assessmentCalls != 0 {
			t.Fatalf("ownership called %d times after Get service miss", fake.assessmentCalls)
		}
	})

	t.Run("update rejects invalid route id before body/service", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: true, teacherOwnsClassSubject: true}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(nonTestAssessmentAdminRequest(http.MethodPut, "/", `{"subject_id":"`+pgUUIDString(subjectID)+`"}`), "id", "not-a-uuid")

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Update() invalid id status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if fake.updateInput.ID.Valid {
			t.Fatalf("Update service called despite invalid id: %+v", fake.updateInput)
		}
	})

	t.Run("update stops when assessment ownership denied before class check", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: false, teacherOwnsClassSubject: true}
		h := &NonTestAssessment{svc: fake}
		body := `{"subject_id":"` + pgUUIDString(subjectID) + `","class_id":"` + pgUUIDString(classID) + `","title":"Observasi"}`
		req := withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body)), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID}), "id", pgUUIDString(assessmentID))
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("Update() denied status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
		if fake.assessmentCalls != 1 || fake.classSubjectCalls != 0 {
			t.Fatalf("ownership calls assessment=%d class=%d, want 1/0", fake.assessmentCalls, fake.classSubjectCalls)
		}
	})

	t.Run("delete invalid id and ownership service error", func(t *testing.T) {
		h := &NonTestAssessment{svc: &fakeNonTestAssessmentHandlerService{}}
		rec := httptest.NewRecorder()
		h.Delete(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodDelete, "/", ""), "id", "bad"))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Delete() invalid id status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}

		fake := &fakeNonTestAssessmentHandlerService{assessmentOwnsErr: errors.New("scope store down")}
		h = &NonTestAssessment{svc: fake}
		rec = httptest.NewRecorder()
		req := withRouteParam(withClaims(httptest.NewRequest(http.MethodDelete, "/", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID}), "id", pgUUIDString(assessmentID))
		h.Delete(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Delete() ownership error status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
		if fake.deleteID.Valid {
			t.Fatalf("Delete service called despite ownership error: %s", pgUUIDString(fake.deleteID))
		}
	})
}

func TestNonTestAssessmentMoreGenerateAndUpsertLowBranches(t *testing.T) {
	assessmentID := handlerTestUUID(164)
	studentID := handlerTestUUID(165)
	teacherID := "33333333-3333-3333-3333-333333333333"
	claims := jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID, "usr": "guru.low"}

	t.Run("generate accepts empty body and invalid class id", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: true}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(``)), claims), "id", pgUUIDString(assessmentID))

		h.GenerateSubmissions(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GenerateSubmissions(empty) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if fake.generateAssessmentID != assessmentID || fake.generateClassID.Valid || pgUUIDString(fake.generateTeacherID) != teacherID {
			t.Fatalf("GenerateSubmissions empty forwarded assessment=%v class=%s teacher=%s", fake.generateAssessmentID, pgUUIDString(fake.generateClassID), pgUUIDString(fake.generateTeacherID))
		}

		rec = httptest.NewRecorder()
		req = withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"class_id":"bad"}`)), claims), "id", pgUUIDString(assessmentID))
		h.GenerateSubmissions(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("GenerateSubmissions(bad class) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("generate bad json and invalid route id", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: true}
		h := &NonTestAssessment{svc: fake}
		rec := httptest.NewRecorder()
		req := withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`)), claims), "id", pgUUIDString(assessmentID))
		h.GenerateSubmissions(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("GenerateSubmissions(bad json) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}

		rec = httptest.NewRecorder()
		req = withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`)), claims), "id", "bad")
		h.GenerateSubmissions(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("GenerateSubmissions(bad id) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("upsert rejects invalid times before service", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: true}
		h := &NonTestAssessment{svc: fake}
		body := `{"student_id":"` + pgUUIDString(studentID) + `","status":"submitted","submitted_at":"17/05/2026"}`
		rec := httptest.NewRecorder()
		req := withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body)), claims), "id", pgUUIDString(assessmentID))

		h.UpsertSubmission(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("UpsertSubmission(bad submitted_at) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if fake.upsertInput.AssessmentID.Valid {
			t.Fatalf("Upsert service called despite invalid submitted_at: %+v", fake.upsertInput)
		}

		body = `{"student_id":"` + pgUUIDString(studentID) + `","status":"reviewed","graded_at":"bad-date"}`
		rec = httptest.NewRecorder()
		req = withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body)), claims), "id", pgUUIDString(assessmentID))
		h.UpsertSubmission(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("UpsertSubmission(bad graded_at) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("upsert submitted status auto-fills submitted_at only", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: true}
		h := &NonTestAssessment{svc: fake}
		body := `{"student_id":"` + pgUUIDString(studentID) + `","status":"submitted","evidence_note":"selesai"}`
		rec := httptest.NewRecorder()
		req := withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body)), claims), "id", pgUUIDString(assessmentID))

		h.UpsertSubmission(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("UpsertSubmission(submitted) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if !fake.upsertInput.SubmittedAt.Valid || fake.upsertInput.GradedAt.Valid || fake.upsertInput.GradedByUsername != "guru.low" {
			t.Fatalf("Upsert submitted timestamps/user = submitted %+v graded %+v user %q", fake.upsertInput.SubmittedAt, fake.upsertInput.GradedAt, fake.upsertInput.GradedByUsername)
		}
	})
}
