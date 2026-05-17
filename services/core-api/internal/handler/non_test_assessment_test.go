package handler

import (
	"context"
	"errors"
	"math/big"
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
	teacherOwnsAssessment   bool
	teacherOwnsErr          error
	assessmentOwnsErr       error
	classSubjectCalls       int
	assessmentCalls         int
	seenAssessmentID        pgtype.UUID
	seenClassID             pgtype.UUID
	seenSubjectID           pgtype.UUID
	seenTeacherID           pgtype.UUID
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
func (f *fakeNonTestAssessmentHandlerService) TeacherOwnsClassSubject(_ context.Context, classID, subjectID, teacherID pgtype.UUID) (bool, error) {
	f.classSubjectCalls++
	f.seenClassID = classID
	f.seenSubjectID = subjectID
	f.seenTeacherID = teacherID
	if f.teacherOwnsErr != nil {
		return false, f.teacherOwnsErr
	}
	return f.teacherOwnsClassSubject, nil
}
func (f *fakeNonTestAssessmentHandlerService) TeacherOwnsAssessment(_ context.Context, assessmentID, teacherID pgtype.UUID) (bool, error) {
	f.assessmentCalls++
	f.seenAssessmentID = assessmentID
	f.seenTeacherID = teacherID
	if f.assessmentOwnsErr != nil {
		return false, f.assessmentOwnsErr
	}
	return f.teacherOwnsAssessment, nil
}

func nonTestNumeric(value int64, exp int32) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(value), Exp: exp, Valid: true}
}

func TestNonTestAssessmentListInputFromRequestParsesFiltersAndBounds(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/non-test-assessments?subject_id=11111111-1111-1111-1111-111111111111&class_id=22222222-2222-2222-2222-222222222222&status=draft&assessment_type=praktik&sync_filter=unsynced&q=observasi&limit=100&offset=7", nil)

	input, err := nonTestAssessmentListInputFromRequest(req)
	if err != nil {
		t.Fatalf("nonTestAssessmentListInputFromRequest() error = %v", err)
	}
	if got := pgUUIDString(input.SubjectID); got != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("SubjectID = %q", got)
	}
	if got := pgUUIDString(input.ClassID); got != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("ClassID = %q", got)
	}
	if input.Status != "draft" || input.AssessmentType != "praktik" || input.SyncFilter != "unsynced" || input.SearchQuery != "observasi" {
		t.Fatalf("filters = %#v", input)
	}
	if input.Limit != 100 || input.Offset != 7 {
		t.Fatalf("pagination = limit %d offset %d, want 100/7", input.Limit, input.Offset)
	}

	defaultsReq := httptest.NewRequest(http.MethodGet, "/api/non-test-assessments?limit=101&offset=-1", nil)
	defaults, err := nonTestAssessmentListInputFromRequest(defaultsReq)
	if err != nil {
		t.Fatalf("defaults parse error = %v", err)
	}
	if defaults.Limit != 25 || defaults.Offset != 0 {
		t.Fatalf("bounded pagination = limit %d offset %d, want defaults 25/0", defaults.Limit, defaults.Offset)
	}
}

func TestNonTestAssessmentListInputFromRequestRejectsBadUUIDFilters(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   string
	}{
		{name: "subject", target: "/api/non-test-assessments?subject_id=nope", want: "Mata pelajaran tidak valid"},
		{name: "class", target: "/api/non-test-assessments?class_id=nope", want: "Rombel tidak valid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := nonTestAssessmentListInputFromRequest(httptest.NewRequest(http.MethodGet, tt.target, nil))
			if err == nil || err.Error() != tt.want {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSerializeNonTestAssessmentMapsAllHelperFields(t *testing.T) {
	dueAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC), Valid: true}
	createdAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC), Valid: true}
	updatedAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC), Valid: true}
	syncedAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 3, 8, 0, 0, 0, time.UTC), Valid: true}

	item := serializeNonTestAssessment(db.NonTestAssessment{
		ID:                   handlerTestUUID(1),
		SubjectID:            handlerTestUUID(2),
		ClassID:              handlerTestUUID(3),
		AssessmentType:       "praktik",
		Title:                "Praktik wudu",
		Description:          "Observasi praktik",
		InstructionHtml:      "<p>Instruksi</p>",
		RubricHtml:           "<p>Rubrik</p>",
		EvidenceRequirements: "foto/video",
		Mode:                 "individual",
		ScoringScale:         "score",
		MaxScore:             nonTestNumeric(985, -1),
		Weight:               nonTestNumeric(125, -2),
		DueAt:                dueAt,
		Status:               "published",
		CreatedByUsername:    "admin",
		AssessorUsername:     "guru.fikih",
		Checklist:            []byte(`["niat","tertib"]`),
		GradeComponentID:     handlerTestUUID(4),
		GradeSyncedAt:        syncedAt,
		GradeSyncedBy:        "admin",
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	})

	if item["title"] != "Praktik wudu" || item["instruction_html"] != "<p>Instruksi</p>" || item["rubric_html"] != "<p>Rubrik</p>" {
		t.Fatalf("serialized text fields = %#v", item)
	}
	if item["max_score"] != 98.5 || item["weight"] != 1.25 {
		t.Fatalf("numeric fields = max %#v weight %#v", item["max_score"], item["weight"])
	}
	checklist, ok := item["checklist"].([]any)
	if !ok || len(checklist) != 2 || checklist[0] != "niat" || checklist[1] != "tertib" {
		t.Fatalf("checklist = %#v", item["checklist"])
	}
	if got := pgUUIDString(handlerTestUUID(4)); item["grade_component_id"] != got {
		t.Fatalf("grade_component_id = %#v, want %s", item["grade_component_id"], got)
	}
	if item["due_at"] != dueAt || item["grade_synced_at"] != syncedAt || item["created_at"] != createdAt || item["updated_at"] != updatedAt {
		t.Fatalf("timestamp fields not preserved: %#v", item)
	}
}

func TestSerializeNonTestSubmissionsMapScoresAndStudentFields(t *testing.T) {
	submittedAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC), Valid: true}
	gradedAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC), Valid: true}

	listItem := serializeNonTestSubmissionListRow(db.ListNonTestSubmissionsRow{
		ID:               handlerTestUUID(5),
		AssessmentID:     handlerTestUUID(6),
		StudentID:        handlerTestUUID(7),
		Nis:              "123",
		Nisn:             "456",
		StudentName:      "Siswa A",
		ClassID:          handlerTestUUID(8),
		ClassName:        "VII A",
		ClassLevel:       "VII",
		Status:           "reviewed",
		EvidenceUrl:      "https://example.test/evidence",
		EvidenceNote:     "lengkap",
		Score:            nonTestNumeric(875, -1),
		Feedback:         "baik",
		SubmittedAt:      submittedAt,
		GradedAt:         gradedAt,
		GradedByUsername: "guru.a",
	})
	if listItem["student_name"] != "Siswa A" || listItem["class_name"] != "VII A" || listItem["score"] != 87.5 {
		t.Fatalf("list submission = %#v", listItem)
	}
	if listItem["submitted_at"] != submittedAt || listItem["graded_at"] != gradedAt {
		t.Fatalf("list timestamps = %#v", listItem)
	}

	submission := serializeNonTestSubmission(db.NonTestAssessmentSubmission{
		ID:               handlerTestUUID(9),
		AssessmentID:     handlerTestUUID(10),
		StudentID:        handlerTestUUID(11),
		Status:           "submitted",
		EvidenceUrl:      "https://example.test/submission",
		EvidenceNote:     "catatan",
		Score:            pgtype.Numeric{},
		Feedback:         "",
		SubmittedAt:      submittedAt,
		GradedByUsername: "guru.b",
	})
	if submission["score"] != nil {
		t.Fatalf("nil score = %#v, want nil", submission["score"])
	}
	if submission["evidence_url"] != "https://example.test/submission" || submission["graded_by_username"] != "guru.b" {
		t.Fatalf("submission = %#v", submission)
	}
}

func TestAbsInt32(t *testing.T) {
	for _, tt := range []struct {
		in   int32
		want int32
	}{
		{in: -12, want: 12},
		{in: 0, want: 0},
		{in: 9, want: 9},
	} {
		if got := absInt32(tt.in); got != tt.want {
			t.Fatalf("absInt32(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestNonTestAssessmentAccessChecksUseClaimsAndServiceScope(t *testing.T) {
	assessmentID := handlerTestUUID(21)
	classID := handlerTestUUID(22)
	subjectID := handlerTestUUID(23)
	teacherID := "33333333-3333-3333-3333-333333333333"

	t.Run("admin bypasses assessment ownership service", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{}
		h := &NonTestAssessment{svc: fake}
		req := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"roles": []any{"admin"}})
		rec := httptest.NewRecorder()

		if !h.requireAssessmentTeacherOrAdmin(rec, req, assessmentID) {
			t.Fatalf("admin access returned false; status=%d body=%s", rec.Code, rec.Body.String())
		}
		if fake.assessmentCalls != 0 {
			t.Fatalf("assessment ownership called %d times for admin", fake.assessmentCalls)
		}
	})

	t.Run("guru requires teacher employee id", func(t *testing.T) {
		h := &NonTestAssessment{svc: &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: true}}
		req := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"roles": []any{"guru"}})
		rec := httptest.NewRecorder()

		if h.requireAssessmentTeacherOrAdmin(rec, req, assessmentID) {
			t.Fatalf("access returned true without eid")
		}
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("guru assessment ownership uses eid claim", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: true}
		h := &NonTestAssessment{svc: fake}
		req := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID})
		rec := httptest.NewRecorder()

		if !h.requireAssessmentTeacherOrAdmin(rec, req, assessmentID) {
			t.Fatalf("access returned false; status=%d body=%s", rec.Code, rec.Body.String())
		}
		if fake.assessmentCalls != 1 || pgUUIDString(fake.seenAssessmentID) != pgUUIDString(assessmentID) || pgUUIDString(fake.seenTeacherID) != teacherID {
			t.Fatalf("ownership call mismatch: calls=%d assessment=%s teacher=%s", fake.assessmentCalls, pgUUIDString(fake.seenAssessmentID), pgUUIDString(fake.seenTeacherID))
		}
	})

	t.Run("service error maps to internal", func(t *testing.T) {
		h := &NonTestAssessment{svc: &fakeNonTestAssessmentHandlerService{assessmentOwnsErr: errors.New("store down")}}
		req := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID})
		rec := httptest.NewRecorder()

		if h.requireAssessmentTeacherOrAdmin(rec, req, assessmentID) {
			t.Fatalf("access returned true on service error")
		}
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("guru class subject ownership uses route IDs and eid claim", func(t *testing.T) {
		fake := &fakeNonTestAssessmentHandlerService{teacherOwnsClassSubject: true}
		h := &NonTestAssessment{svc: fake}
		req := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID})
		rec := httptest.NewRecorder()

		if !h.requireClassSubjectTeacherOrAdmin(rec, req, classID, subjectID) {
			t.Fatalf("class subject access returned false; status=%d body=%s", rec.Code, rec.Body.String())
		}
		if fake.classSubjectCalls != 1 || pgUUIDString(fake.seenClassID) != pgUUIDString(classID) || pgUUIDString(fake.seenSubjectID) != pgUUIDString(subjectID) || pgUUIDString(fake.seenTeacherID) != teacherID {
			t.Fatalf("class subject call mismatch: calls=%d class=%s subject=%s teacher=%s", fake.classSubjectCalls, pgUUIDString(fake.seenClassID), pgUUIDString(fake.seenSubjectID), pgUUIDString(fake.seenTeacherID))
		}
	})

	t.Run("class subject ownership denied", func(t *testing.T) {
		h := &NonTestAssessment{svc: &fakeNonTestAssessmentHandlerService{teacherOwnsClassSubject: false}}
		req := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID})
		rec := httptest.NewRecorder()

		if h.requireClassSubjectTeacherOrAdmin(rec, req, classID, subjectID) {
			t.Fatalf("class subject access returned true when service denied")
		}
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})
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
