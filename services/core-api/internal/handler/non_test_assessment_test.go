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
	listInput               service.ListNonTestAssessmentsInput
	listRows                []db.ListNonTestAssessmentsRow
	listTotal               int64
	listErr                 error
	getID                   pgtype.UUID
	getRow                  db.GetNonTestAssessmentRow
	getErr                  error
	updateInput             service.SaveNonTestAssessmentInput
	updateRow               db.NonTestAssessment
	updateErr               error
	deleteID                pgtype.UUID
	deleteErr               error
	listSubmissionsID       pgtype.UUID
	listSubmissionsRows     []db.ListNonTestSubmissionsRow
	listSubmissionsErr      error
	generateAssessmentID    pgtype.UUID
	generateClassID         pgtype.UUID
	generateRows            []db.NonTestAssessmentSubmission
	generateErr             error
	syncAssessmentID        pgtype.UUID
	syncTeacherID           pgtype.UUID
	syncUsername            string
	syncPublish             bool
	syncResult              service.SyncNonTestAssessmentToGradeResult
	syncErr                 error
	upsertRow               db.NonTestAssessmentSubmission
	upsertErr               error
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

func (f *fakeNonTestAssessmentHandlerService) List(_ context.Context, input service.ListNonTestAssessmentsInput) ([]db.ListNonTestAssessmentsRow, int64, error) {
	f.listInput = input
	return f.listRows, f.listTotal, f.listErr
}
func (f *fakeNonTestAssessmentHandlerService) Get(_ context.Context, id pgtype.UUID) (db.GetNonTestAssessmentRow, error) {
	f.getID = id
	return f.getRow, f.getErr
}
func (f *fakeNonTestAssessmentHandlerService) Create(context.Context, service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	return db.NonTestAssessment{ID: handlerTestUUID(40)}, nil
}
func (f *fakeNonTestAssessmentHandlerService) Update(_ context.Context, input service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	f.updateInput = input
	if f.updateRow.ID.Valid || f.updateRow.SubjectID.Valid {
		return f.updateRow, f.updateErr
	}
	return db.NonTestAssessment{ID: input.ID, SubjectID: input.SubjectID, ClassID: input.ClassID, Title: input.Title, AssessmentType: input.AssessmentType, Status: input.Status}, f.updateErr
}
func (f *fakeNonTestAssessmentHandlerService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}
func (f *fakeNonTestAssessmentHandlerService) ListSubmissions(_ context.Context, id pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error) {
	f.listSubmissionsID = id
	return f.listSubmissionsRows, f.listSubmissionsErr
}
func (f *fakeNonTestAssessmentHandlerService) GenerateSubmissions(_ context.Context, assessmentID pgtype.UUID, classID, teacherID pgtype.UUID) ([]db.NonTestAssessmentSubmission, error) {
	f.generateAssessmentID = assessmentID
	f.generateClassID = classID
	f.generateTeacherID = teacherID
	return f.generateRows, f.generateErr
}
func (f *fakeNonTestAssessmentHandlerService) UpsertSubmission(_ context.Context, input service.SaveNonTestSubmissionInput) (db.NonTestAssessmentSubmission, error) {
	f.upsertInput = input
	if f.upsertRow.AssessmentID.Valid || f.upsertRow.StudentID.Valid {
		return f.upsertRow, f.upsertErr
	}
	return db.NonTestAssessmentSubmission{AssessmentID: input.AssessmentID, StudentID: input.StudentID, Status: input.Status}, f.upsertErr
}
func (f *fakeNonTestAssessmentHandlerService) SyncToGrade(_ context.Context, assessmentID, teacherID pgtype.UUID, username string, publish bool) (service.SyncNonTestAssessmentToGradeResult, error) {
	f.syncAssessmentID = assessmentID
	f.syncTeacherID = teacherID
	f.syncUsername = username
	f.syncPublish = publish
	return f.syncResult, f.syncErr
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

func nonTestAssessmentAdminRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "usr": "admin.test"})
}

func TestNonTestAssessmentListEndpointSuccessAndError(t *testing.T) {
	teacherID := "33333333-3333-3333-3333-333333333333"
	fake := &fakeNonTestAssessmentHandlerService{
		listRows: []db.ListNonTestAssessmentsRow{{
			ID:             handlerTestUUID(41),
			SubjectID:      handlerTestUUID(42),
			SubjectName:    "Fikih",
			AssessmentType: "praktik",
			Title:          "Praktik salat",
			MaxScore:       nonTestNumeric(100, 0),
		}},
		listTotal: 9,
	}
	h := &NonTestAssessment{svc: fake}
	req := httptest.NewRequest(http.MethodGet, "/api/non-test-assessments?status=active&limit=10&offset=2", nil)
	req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listInput.Status != "active" || fake.listInput.Limit != 10 || fake.listInput.Offset != 2 || pgUUIDString(fake.listInput.TeacherEmployeeID) != teacherID {
		t.Fatalf("List input = %+v, want query filters and teacher id", fake.listInput)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Praktik salat") || !strings.Contains(body, `"total":9`) {
		t.Fatalf("List body = %s, want item and total", body)
	}

	rec = httptest.NewRecorder()
	h.List(rec, nonTestAssessmentAdminRequest(http.MethodGet, "/api/non-test-assessments?subject_id=bad", ""))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("List() invalid filter status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestNonTestAssessmentGetEndpointSuccessAndErrors(t *testing.T) {
	assessmentID := handlerTestUUID(43)
	fake := &fakeNonTestAssessmentHandlerService{
		teacherOwnsAssessment: true,
		getRow: db.GetNonTestAssessmentRow{
			ID:             assessmentID,
			SubjectID:      handlerTestUUID(44),
			SubjectName:    "Akidah",
			AssessmentType: "observasi",
			Title:          "Observasi sikap",
			MaxScore:       nonTestNumeric(100, 0),
		},
	}
	h := &NonTestAssessment{svc: fake}
	req := withRouteParam(nonTestAssessmentAdminRequest(http.MethodGet, "/api/non-test-assessments/"+pgUUIDString(assessmentID), ""), "id", pgUUIDString(assessmentID))
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Get() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.getID != assessmentID || !strings.Contains(rec.Body.String(), "Observasi sikap") {
		t.Fatalf("Get forwarded id/body mismatch: id=%v body=%s", fake.getID, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Get(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodGet, "/api/non-test-assessments/bad", ""), "id", "bad"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Get() invalid id status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestNonTestAssessmentUpdateAndDeleteEndpointPaths(t *testing.T) {
	assessmentID := handlerTestUUID(45)
	subjectID := handlerTestUUID(46)
	classID := handlerTestUUID(47)
	fake := &fakeNonTestAssessmentHandlerService{teacherOwnsAssessment: true, teacherOwnsClassSubject: true}
	h := &NonTestAssessment{svc: fake}
	body := `{"subject_id":"` + pgUUIDString(subjectID) + `","class_id":"` + pgUUIDString(classID) + `","assessment_type":"proyek","title":" Proyek IPA ","status":"active","checklist":["proposal"]}`
	req := withRouteParam(nonTestAssessmentAdminRequest(http.MethodPut, "/api/non-test-assessments/"+pgUUIDString(assessmentID), body), "id", pgUUIDString(assessmentID))
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Update() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateInput.ID != assessmentID || fake.updateInput.SubjectID != subjectID || fake.updateInput.ClassID != classID || fake.updateInput.Title != " Proyek IPA " {
		t.Fatalf("Update input = %+v, want body and route id forwarded", fake.updateInput)
	}

	rec = httptest.NewRecorder()
	h.Update(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodPut, "/api/non-test-assessments/"+pgUUIDString(assessmentID), `{`), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Update() bad json status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Delete(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodDelete, "/api/non-test-assessments/"+pgUUIDString(assessmentID), ""), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Delete() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID != assessmentID {
		t.Fatalf("Delete id = %v, want %v", fake.deleteID, assessmentID)
	}

	fake.deleteErr = errors.New("akses ditolak")
	rec = httptest.NewRecorder()
	h.Delete(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodDelete, "/api/non-test-assessments/"+pgUUIDString(assessmentID), ""), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Delete() service error status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestNonTestAssessmentSubmissionEndpoints(t *testing.T) {
	assessmentID := handlerTestUUID(48)
	classID := handlerTestUUID(49)
	studentID := handlerTestUUID(50)
	teacherID := "33333333-3333-3333-3333-333333333333"
	fake := &fakeNonTestAssessmentHandlerService{
		teacherOwnsAssessment: true,
		listSubmissionsRows: []db.ListNonTestSubmissionsRow{{
			ID:           handlerTestUUID(51),
			AssessmentID: assessmentID,
			StudentID:    studentID,
			StudentName:  "Siswa Non Tes",
			Status:       "reviewed",
			Score:        nonTestNumeric(875, -1),
		}},
		generateRows: []db.NonTestAssessmentSubmission{{AssessmentID: assessmentID, StudentID: studentID}},
		syncResult: service.SyncNonTestAssessmentToGradeResult{
			AssessmentID:     pgUUIDString(assessmentID),
			GradeComponentID: pgUUIDString(handlerTestUUID(52)),
			SyncedEntries:    1,
			IsPublished:      false,
		},
	}
	h := &NonTestAssessment{svc: fake}
	claims := jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID, "usr": "guru.non.tes"}

	rec := httptest.NewRecorder()
	h.ListSubmissions(rec, withRouteParam(withClaims(httptest.NewRequest(http.MethodGet, "/", nil), claims), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusOK || fake.listSubmissionsID != assessmentID || !strings.Contains(rec.Body.String(), "Siswa Non Tes") {
		t.Fatalf("ListSubmissions status/id/body = %d/%v/%s", rec.Code, fake.listSubmissionsID, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	generateBody := `{"class_id":"` + pgUUIDString(classID) + `"}`
	h.GenerateSubmissions(rec, withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(generateBody)), claims), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusOK {
		t.Fatalf("GenerateSubmissions status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.generateAssessmentID != assessmentID || fake.generateClassID != classID || pgUUIDString(fake.generateTeacherID) != teacherID || !strings.Contains(rec.Body.String(), `"created_count":1`) {
		t.Fatalf("GenerateSubmissions forwarded assessment=%v class=%v teacher=%s body=%s", fake.generateAssessmentID, fake.generateClassID, pgUUIDString(fake.generateTeacherID), rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.SyncGrade(rec, withRouteParam(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"is_published":false}`)), claims), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusOK {
		t.Fatalf("SyncGrade status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.syncAssessmentID != assessmentID || pgUUIDString(fake.syncTeacherID) != teacherID || fake.syncUsername != "guru.non.tes" || fake.syncPublish {
		t.Fatalf("SyncGrade forwarded assessment=%v teacher=%s user=%q publish=%v", fake.syncAssessmentID, pgUUIDString(fake.syncTeacherID), fake.syncUsername, fake.syncPublish)
	}

	score := "88.5"
	upsertBody := `{"student_id":"` + pgUUIDString(studentID) + `","status":"reviewed","score":` + score + `,"feedback":" Mantap "}`
	rec = httptest.NewRecorder()
	h.UpsertSubmission(rec, withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/", strings.NewReader(upsertBody)), claims), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpsertSubmission status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.upsertInput.AssessmentID != assessmentID || fake.upsertInput.StudentID != studentID || fake.upsertInput.Score == nil || *fake.upsertInput.Score != 88.5 || !fake.upsertInput.GradedAt.Valid || fake.upsertInput.GradedByUsername != "guru.non.tes" {
		t.Fatalf("Upsert input = %+v, want reviewed submission with auto graded_at", fake.upsertInput)
	}

	rec = httptest.NewRecorder()
	h.UpsertSubmission(rec, withRouteParam(withClaims(httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"student_id":"bad"}`)), claims), "id", pgUUIDString(assessmentID)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UpsertSubmission bad student status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestNonTestAssessmentEndpointServiceErrors(t *testing.T) {
	assessmentID := handlerTestUUID(53)
	tests := []struct {
		name string
		run  func(*NonTestAssessment, *httptest.ResponseRecorder)
	}{
		{name: "list submissions", run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
			h.ListSubmissions(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodGet, "/", ""), "id", pgUUIDString(assessmentID)))
		}},
		{name: "generate", run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
			h.GenerateSubmissions(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodPost, "/", "{}"), "id", pgUUIDString(assessmentID)))
		}},
		{name: "sync", run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
			h.SyncGrade(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodPost, "/", "{}"), "id", pgUUIDString(assessmentID)))
		}},
		{name: "upsert", run: func(h *NonTestAssessment, rec *httptest.ResponseRecorder) {
			h.UpsertSubmission(rec, withRouteParam(nonTestAssessmentAdminRequest(http.MethodPut, "/", `{"student_id":"`+pgUUIDString(handlerTestUUID(54))+`"}`), "id", pgUUIDString(assessmentID)))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeNonTestAssessmentHandlerService{
				listSubmissionsErr: errors.New("akses ditolak"),
				generateErr:        errors.New("akses ditolak"),
				syncErr:            errors.New("akses ditolak"),
				upsertErr:          errors.New("akses ditolak"),
			}
			h := &NonTestAssessment{svc: fake}
			rec := httptest.NewRecorder()
			tt.run(h, rec)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("%s status = %d, want 403; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}
}
