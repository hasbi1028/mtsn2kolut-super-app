package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type extraNonTestAssessmentHandlerService struct {
	createInput             service.SaveNonTestAssessmentInput
	createRow               db.NonTestAssessment
	createErr               error
	teacherOwnsClassSubject bool
	teacherOwnsAssessment   bool
	classSubjectCalls       int
	syncAssessmentID        pgtype.UUID
	syncTeacherID           pgtype.UUID
	syncUsername            string
	syncPublish             bool
	syncResult              service.SyncNonTestAssessmentToGradeResult
	syncErr                 error
}

func (f *extraNonTestAssessmentHandlerService) List(context.Context, service.ListNonTestAssessmentsInput) ([]db.ListNonTestAssessmentsRow, int64, error) {
	return nil, 0, nil
}
func (f *extraNonTestAssessmentHandlerService) Get(context.Context, pgtype.UUID) (db.GetNonTestAssessmentRow, error) {
	return db.GetNonTestAssessmentRow{}, nil
}
func (f *extraNonTestAssessmentHandlerService) Create(_ context.Context, input service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	f.createInput = input
	if f.createRow.ID.Valid || f.createRow.SubjectID.Valid {
		return f.createRow, f.createErr
	}
	return db.NonTestAssessment{ID: handlerTestUUID(81), SubjectID: input.SubjectID, ClassID: input.ClassID, AssessmentType: input.AssessmentType, Title: input.Title, DueAt: input.DueAt, CreatedByUsername: input.CreatedByUsername, Checklist: input.Checklist}, f.createErr
}
func (f *extraNonTestAssessmentHandlerService) Update(context.Context, service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	return db.NonTestAssessment{}, nil
}
func (f *extraNonTestAssessmentHandlerService) Delete(context.Context, pgtype.UUID) error { return nil }
func (f *extraNonTestAssessmentHandlerService) ListSubmissions(context.Context, pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error) {
	return nil, nil
}
func (f *extraNonTestAssessmentHandlerService) GenerateSubmissions(context.Context, pgtype.UUID, pgtype.UUID, pgtype.UUID) ([]db.NonTestAssessmentSubmission, error) {
	return nil, nil
}
func (f *extraNonTestAssessmentHandlerService) UpsertSubmission(context.Context, service.SaveNonTestSubmissionInput) (db.NonTestAssessmentSubmission, error) {
	return db.NonTestAssessmentSubmission{}, nil
}
func (f *extraNonTestAssessmentHandlerService) SyncToGrade(_ context.Context, assessmentID, teacherID pgtype.UUID, username string, publish bool) (service.SyncNonTestAssessmentToGradeResult, error) {
	f.syncAssessmentID = assessmentID
	f.syncTeacherID = teacherID
	f.syncUsername = username
	f.syncPublish = publish
	return f.syncResult, f.syncErr
}
func (f *extraNonTestAssessmentHandlerService) TeacherOwnsClassSubject(context.Context, pgtype.UUID, pgtype.UUID, pgtype.UUID) (bool, error) {
	f.classSubjectCalls++
	return f.teacherOwnsClassSubject, nil
}
func (f *extraNonTestAssessmentHandlerService) TeacherOwnsAssessment(context.Context, pgtype.UUID, pgtype.UUID) (bool, error) {
	return f.teacherOwnsAssessment, nil
}

func TestNonTestAssessmentCreateForwardsParsedBodyAndTeacherScope(t *testing.T) {
	subjectID := "11111111-1111-1111-1111-111111111111"
	classID := "22222222-2222-2222-2222-222222222222"
	teacherID := "33333333-3333-3333-3333-333333333333"
	fake := &extraNonTestAssessmentHandlerService{teacherOwnsClassSubject: true}
	h := &NonTestAssessment{svc: fake}
	body := `{"subject_id":"` + subjectID + `","class_id":"` + classID + `","assessment_type":"praktik","title":" Praktik tayamum ","description":" desc ","instruction_html":"<p>Ikuti</p>","rubric_html":"<p>Rubrik</p>","evidence_requirements":"foto","mode":"advance","scoring_scale":"0_100","max_score":90,"weight":2,"due_at":"2026-05-17T09:30","status":"active","assessor_username":" guru.asessor ","checklist":["niat","tertib"]}`
	req := withClaims(httptest.NewRequest(http.MethodPost, "/api/non-test-assessments", strings.NewReader(body)), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID, "usr": "guru.pai"})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Create() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.classSubjectCalls != 1 {
		t.Fatalf("TeacherOwnsClassSubject calls = %d, want 1", fake.classSubjectCalls)
	}
	if got := pgUUIDString(fake.createInput.SubjectID); got != subjectID {
		t.Fatalf("SubjectID = %s, want %s", got, subjectID)
	}
	if got := pgUUIDString(fake.createInput.ClassID); got != classID {
		t.Fatalf("ClassID = %s, want %s", got, classID)
	}
	if fake.createInput.Title != " Praktik tayamum " || fake.createInput.CreatedByUsername != "guru.pai" || fake.createInput.MaxScore != 90 || fake.createInput.Weight != 2 {
		t.Fatalf("Create input = %+v, want body fields and current username forwarded", fake.createInput)
	}
	if !fake.createInput.DueAt.Valid || fake.createInput.DueAt.Time.Location() == time.UTC {
		t.Fatalf("DueAt = %+v, want parsed local non-RFC3339 time", fake.createInput.DueAt)
	}
	if string(fake.createInput.Checklist) != `["niat","tertib"]` {
		t.Fatalf("Checklist = %s, want marshaled checklist", string(fake.createInput.Checklist))
	}
}

func TestNonTestAssessmentCreateRejectsInvalidDueAtBeforeService(t *testing.T) {
	fake := &extraNonTestAssessmentHandlerService{teacherOwnsClassSubject: true}
	h := &NonTestAssessment{svc: fake}
	body := `{"subject_id":"11111111-1111-1111-1111-111111111111","class_id":"22222222-2222-2222-2222-222222222222","title":"Observasi","due_at":"17/05/2026"}`
	req := nonTestAssessmentAdminRequest(http.MethodPost, "/api/non-test-assessments", body)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Create() status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createInput.SubjectID.Valid {
		t.Fatalf("service Create called despite invalid due_at: %+v", fake.createInput)
	}
}

func TestParseOptionalTimeAcceptsSupportedFormatsAndRejectsInvalid(t *testing.T) {
	cases := []string{"2026-05-17T10:15:30Z", "2026-05-17T10:15", "2026-05-17 10:15", "2026-05-17"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			got, err := parseOptionalTime("  " + raw + "  ")
			if err != nil || !got.Valid {
				t.Fatalf("parseOptionalTime(%q) = %+v/%v, want valid", raw, got, err)
			}
		})
	}
	blank, err := parseOptionalTime("   ")
	if err != nil || blank.Valid {
		t.Fatalf("parseOptionalTime(blank) = %+v/%v, want invalid nil error", blank, err)
	}
	if _, err := parseOptionalTime("2026/05/17"); err == nil {
		t.Fatalf("parseOptionalTime(invalid) error = nil, want error")
	}
}

func TestNonTestAssessmentSyncGradeDefaultsPublishAndMapsErrors(t *testing.T) {
	assessmentID := handlerTestUUID(82)
	teacherID := "33333333-3333-3333-3333-333333333333"
	fake := &extraNonTestAssessmentHandlerService{
		teacherOwnsAssessment: true,
		syncResult: service.SyncNonTestAssessmentToGradeResult{
			AssessmentID:     pgUUIDString(assessmentID),
			GradeComponentID: "55555555-5555-5555-5555-555555555555",
			SyncedEntries:    2,
			SkippedEntries:   1,
			IsPublished:      true,
		},
	}
	h := &NonTestAssessment{svc: fake}
	claims := jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID, "usr": "guru.sync"}
	req := withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`)), claims), "id", pgUUIDString(assessmentID))
	rec := httptest.NewRecorder()

	h.SyncGrade(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("SyncGrade() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.syncAssessmentID != assessmentID || pgUUIDString(fake.syncTeacherID) != teacherID || fake.syncUsername != "guru.sync" || !fake.syncPublish {
		t.Fatalf("SyncGrade forwarded assessment=%v teacher=%s username=%q publish=%v", fake.syncAssessmentID, pgUUIDString(fake.syncTeacherID), fake.syncUsername, fake.syncPublish)
	}
	if !strings.Contains(rec.Body.String(), `"skipped_entries":1`) || !strings.Contains(rec.Body.String(), `"is_published":true`) {
		t.Fatalf("SyncGrade body = %s, want sync result JSON", rec.Body.String())
	}

	fake.syncErr = errors.New("akses ditolak")
	rec = httptest.NewRecorder()
	h.SyncGrade(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodPost, "/", `{}`), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("SyncGrade() service error status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.SyncGrade(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodPost, "/", `{`), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("SyncGrade() bad JSON status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}
