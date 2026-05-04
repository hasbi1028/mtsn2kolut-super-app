package handler

import (
	"context"
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

type fakeNonTestAssessmentHandlerService struct {
	teacherOwnsClassSubject bool
	generateTeacherID       pgtype.UUID
	upsertInput             service.SaveNonTestSubmissionInput
}

func (f *fakeNonTestAssessmentHandlerService) List(context.Context, service.ListNonTestAssessmentsInput) ([]db.ListNonTestAssessmentsRow, int64, error) {
	return nil, 0, nil
}
func (f *fakeNonTestAssessmentHandlerService) Get(context.Context, pgtype.UUID) (db.GetNonTestAssessmentRow, error) {
	return db.GetNonTestAssessmentRow{}, nil
}
func (f *fakeNonTestAssessmentHandlerService) Create(context.Context, service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	return db.NonTestAssessment{ID: handlerTestUUID(40)}, nil
}
func (f *fakeNonTestAssessmentHandlerService) Update(context.Context, service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	return db.NonTestAssessment{}, nil
}
func (f *fakeNonTestAssessmentHandlerService) Delete(context.Context, pgtype.UUID) error { return nil }
func (f *fakeNonTestAssessmentHandlerService) ListSubmissions(context.Context, pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error) {
	return nil, nil
}
func (f *fakeNonTestAssessmentHandlerService) GenerateSubmissions(_ context.Context, _ pgtype.UUID, _, teacherID pgtype.UUID) ([]db.NonTestAssessmentSubmission, error) {
	f.generateTeacherID = teacherID
	return nil, nil
}
func (f *fakeNonTestAssessmentHandlerService) UpsertSubmission(_ context.Context, input service.SaveNonTestSubmissionInput) (db.NonTestAssessmentSubmission, error) {
	f.upsertInput = input
	return db.NonTestAssessmentSubmission{AssessmentID: input.AssessmentID, StudentID: input.StudentID, Status: input.Status}, nil
}
func (f *fakeNonTestAssessmentHandlerService) SyncToGrade(context.Context, pgtype.UUID, pgtype.UUID, string, bool) (service.SyncNonTestAssessmentToGradeResult, error) {
	return service.SyncNonTestAssessmentToGradeResult{}, nil
}
func (f *fakeNonTestAssessmentHandlerService) TeacherOwnsClassSubject(context.Context, pgtype.UUID, pgtype.UUID, pgtype.UUID) (bool, error) {
	return f.teacherOwnsClassSubject, nil
}
func (f *fakeNonTestAssessmentHandlerService) TeacherOwnsAssessment(context.Context, pgtype.UUID, pgtype.UUID) (bool, error) {
	return true, nil
}

func TestSerializeNonTestAssessmentRowsIncludeSyncFreshness(t *testing.T) {
	lastReviewedAt := pgtype.Timestamptz{
		Time:  time.Date(2026, 5, 3, 10, 30, 0, 0, time.UTC),
		Valid: true,
	}

	listItem := serializeNonTestAssessmentListRow(db.ListNonTestAssessmentsRow{
		ID:                          handlerTestUUID(31),
		SubjectID:                   handlerTestUUID(32),
		LastReviewedAt:              lastReviewedAt,
		UnsyncedReviewedSubmissions: 2,
	})
	if got, ok := listItem["last_reviewed_at"].(pgtype.Timestamptz); !ok || !got.Valid || !got.Time.Equal(lastReviewedAt.Time) {
		t.Fatalf("list last_reviewed_at = %#v, want %v", listItem["last_reviewed_at"], lastReviewedAt.Time)
	}
	if got := listItem["unsynced_reviewed_submissions"]; got != int32(2) {
		t.Fatalf("list unsynced_reviewed_submissions = %#v, want 2", got)
	}

	detailItem := serializeNonTestAssessmentDetailRow(db.GetNonTestAssessmentRow{
		ID:                          handlerTestUUID(33),
		SubjectID:                   handlerTestUUID(34),
		LastReviewedAt:              lastReviewedAt,
		UnsyncedReviewedSubmissions: 3,
	})
	if got, ok := detailItem["last_reviewed_at"].(pgtype.Timestamptz); !ok || !got.Valid || !got.Time.Equal(lastReviewedAt.Time) {
		t.Fatalf("detail last_reviewed_at = %#v, want %v", detailItem["last_reviewed_at"], lastReviewedAt.Time)
	}
	if got := detailItem["unsynced_reviewed_submissions"]; got != int32(3) {
		t.Fatalf("detail unsynced_reviewed_submissions = %#v, want 3", got)
	}
}

func TestNonTestAssessmentGuruCreateForbiddenForUnassignedClassSubject(t *testing.T) {
	h := &NonTestAssessment{svc: &fakeNonTestAssessmentHandlerService{teacherOwnsClassSubject: false}}
	req := httptest.NewRequest(http.MethodPost, "/api/non-test-assessments", strings.NewReader(`{"subject_id":"11111111-1111-1111-1111-111111111111","class_id":"22222222-2222-2222-2222-222222222222","title":"Praktik kelas B"}`))
	req = withClaims(req, jwt.MapClaims{
		"roles": []any{"guru"},
		"eid":   "33333333-3333-3333-3333-333333333333",
		"usr":   "guru.a",
	})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("Create() status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}
