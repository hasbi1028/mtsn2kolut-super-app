package service

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeNonTestAssessmentStore struct {
	listParams                 db.ListNonTestAssessmentsParams
	countParams                db.CountNonTestAssessmentsParams
	createParams               db.CreateNonTestAssessmentParams
	updateParams               db.UpdateNonTestAssessmentParams
	generateParams             db.GenerateNonTestSubmissionsForClassParams
	upsertParams               db.UpsertNonTestSubmissionParams
	getAssessmentID            pgtype.UUID
	assessmentRow              db.GetNonTestAssessmentRow
	getAssessmentErr           error
	submissionRows             []db.ListNonTestSubmissionsRow
	studentBelongs             bool
	teacherOwnsClassSubjectSet bool
	teacherOwnsClassSubject    bool

	gradeAssignment           db.GetGradeAssignmentByClassSubjectRow
	gradeAssignmentErr        error
	finalization              db.GradeAssignmentFinalization
	finalizationErr           error
	finalizationErrSet        bool
	gradeComponent            db.GradeComponent
	getGradeComponentErr      error
	createGradeComponentParam db.CreateGradeComponentParams
	updatePublishParam        db.UpdateGradeComponentPublishStateParams
	upsertGradeParams         []db.UpsertGradeEntryParams
	markGradeSyncParams       db.MarkNonTestAssessmentGradeSyncParams

	createErr error
}

func (f *fakeNonTestAssessmentStore) WithTx(_ pgx.Tx) *db.Queries {
	return nil
}

func (f *fakeNonTestAssessmentStore) ListNonTestAssessments(_ context.Context, arg db.ListNonTestAssessmentsParams) ([]db.ListNonTestAssessmentsRow, error) {
	f.listParams = arg
	return []db.ListNonTestAssessmentsRow{}, nil
}

func (f *fakeNonTestAssessmentStore) CountNonTestAssessments(_ context.Context, arg db.CountNonTestAssessmentsParams) (int64, error) {
	f.countParams = arg
	return 0, nil
}

func (f *fakeNonTestAssessmentStore) GetNonTestAssessment(_ context.Context, id pgtype.UUID) (db.GetNonTestAssessmentRow, error) {
	f.getAssessmentID = id
	if f.getAssessmentErr != nil {
		return db.GetNonTestAssessmentRow{}, f.getAssessmentErr
	}
	if f.assessmentRow.ID.Valid || f.assessmentRow.SubjectID.Valid || f.assessmentRow.ClassID.Valid {
		row := f.assessmentRow
		if !row.ID.Valid {
			row.ID = id
		}
		return row, nil
	}
	return db.GetNonTestAssessmentRow{
		ID:                id,
		ClassID:           uuidForNonTest("22222222-2222-2222-2222-222222222222"),
		CreatedByUsername: "guru.lama",
		MaxScore:          pgNumeric(100),
	}, nil
}

func (f *fakeNonTestAssessmentStore) CreateNonTestAssessment(_ context.Context, arg db.CreateNonTestAssessmentParams) (db.NonTestAssessment, error) {
	f.createParams = arg
	return db.NonTestAssessment{SubjectID: arg.SubjectID, AssessmentType: arg.AssessmentType, Title: arg.Title}, f.createErr
}

func (f *fakeNonTestAssessmentStore) UpdateNonTestAssessment(_ context.Context, arg db.UpdateNonTestAssessmentParams) (db.NonTestAssessment, error) {
	f.updateParams = arg
	return db.NonTestAssessment{ID: arg.ID, SubjectID: arg.SubjectID, AssessmentType: arg.AssessmentType, Title: arg.Title}, nil
}

func (f *fakeNonTestAssessmentStore) DeleteNonTestAssessment(_ context.Context, _ pgtype.UUID) error {
	return nil
}

func (f *fakeNonTestAssessmentStore) ListNonTestSubmissions(_ context.Context, _ pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error) {
	return f.submissionRows, nil
}

func (f *fakeNonTestAssessmentStore) GenerateNonTestSubmissionsForClass(_ context.Context, arg db.GenerateNonTestSubmissionsForClassParams) ([]db.NonTestAssessmentSubmission, error) {
	f.generateParams = arg
	return []db.NonTestAssessmentSubmission{{AssessmentID: arg.AssessmentID, StudentID: uuidForNonTest("33333333-3333-3333-3333-333333333333"), Status: "assigned"}}, nil
}

func (f *fakeNonTestAssessmentStore) UpsertNonTestSubmission(_ context.Context, arg db.UpsertNonTestSubmissionParams) (db.NonTestAssessmentSubmission, error) {
	f.upsertParams = arg
	return db.NonTestAssessmentSubmission{AssessmentID: arg.AssessmentID, StudentID: arg.StudentID, Status: arg.Status}, nil
}

func (f *fakeNonTestAssessmentStore) StudentBelongsToClass(_ context.Context, _ db.StudentBelongsToClassParams) (bool, error) {
	return f.studentBelongs, nil
}

func (f *fakeNonTestAssessmentStore) GetGradeAssignmentByClassSubject(_ context.Context, arg db.GetGradeAssignmentByClassSubjectParams) (db.GetGradeAssignmentByClassSubjectRow, error) {
	if f.gradeAssignmentErr != nil {
		return db.GetGradeAssignmentByClassSubjectRow{}, f.gradeAssignmentErr
	}
	if f.gradeAssignment.ID.Valid {
		return f.gradeAssignment, nil
	}
	return db.GetGradeAssignmentByClassSubjectRow{
		ID:        uuidForNonTest("44444444-4444-4444-4444-444444444444"),
		ClassID:   arg.ClassID,
		SubjectID: arg.SubjectID,
	}, nil
}

func (f *fakeNonTestAssessmentStore) TeacherOwnsClassSubject(_ context.Context, arg db.TeacherOwnsClassSubjectParams) (bool, error) {
	if f.teacherOwnsClassSubjectSet {
		return f.teacherOwnsClassSubject, nil
	}
	return arg.TeacherEmployeeID.Valid, nil
}

func (f *fakeNonTestAssessmentStore) TeacherOwnsNonTestAssessment(_ context.Context, arg db.TeacherOwnsNonTestAssessmentParams) (bool, error) {
	return arg.TeacherEmployeeID.Valid, nil
}

func (f *fakeNonTestAssessmentStore) GetGradeAssignmentFinalization(_ context.Context, _ pgtype.UUID) (db.GradeAssignmentFinalization, error) {
	if f.finalizationErrSet {
		return f.finalization, f.finalizationErr
	}
	return db.GradeAssignmentFinalization{}, pgx.ErrNoRows
}

func (f *fakeNonTestAssessmentStore) GetGradeComponent(_ context.Context, id pgtype.UUID) (db.GradeComponent, error) {
	if f.getGradeComponentErr != nil {
		return db.GradeComponent{}, f.getGradeComponentErr
	}
	if f.gradeComponent.ID.Valid && f.gradeComponent.ID == id {
		return f.gradeComponent, nil
	}
	return db.GradeComponent{}, pgx.ErrNoRows
}

func (f *fakeNonTestAssessmentStore) CreateGradeComponent(_ context.Context, arg db.CreateGradeComponentParams) (db.GradeComponent, error) {
	f.createGradeComponentParam = arg
	component := db.GradeComponent{
		ID:           uuidForNonTest("55555555-5555-5555-5555-555555555555"),
		AssignmentID: arg.AssignmentID,
		Title:        arg.Title,
		Category:     arg.Category,
		Weight:       arg.Weight,
		MaxScore:     arg.MaxScore,
		IsPublished:  arg.IsPublished,
	}
	f.gradeComponent = component
	return component, nil
}

func (f *fakeNonTestAssessmentStore) UpdateGradeComponentPublishState(_ context.Context, arg db.UpdateGradeComponentPublishStateParams) (db.GradeComponent, error) {
	f.updatePublishParam = arg
	f.gradeComponent.IsPublished = arg.IsPublished
	return f.gradeComponent, nil
}

func (f *fakeNonTestAssessmentStore) UpsertGradeEntry(_ context.Context, arg db.UpsertGradeEntryParams) (db.GradeEntry, error) {
	f.upsertGradeParams = append(f.upsertGradeParams, arg)
	return db.GradeEntry{ComponentID: arg.ComponentID, StudentID: arg.StudentID, Score: arg.Score, Notes: arg.Notes, GradedBy: arg.GradedBy}, nil
}

func (f *fakeNonTestAssessmentStore) MarkNonTestAssessmentGradeSync(_ context.Context, arg db.MarkNonTestAssessmentGradeSyncParams) (db.NonTestAssessment, error) {
	f.markGradeSyncParams = arg
	return db.NonTestAssessment{ID: arg.ID, GradeComponentID: arg.GradeComponentID, GradeSyncedBy: arg.GradeSyncedBy}, nil
}

func TestNonTestAssessmentCreateDefaultsAndTrims(t *testing.T) {
	store := &fakeNonTestAssessmentStore{}
	svc := &NonTestAssessment{q: store}
	subjectID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")

	_, err := svc.Create(context.Background(), SaveNonTestAssessmentInput{
		SubjectID:         subjectID,
		Title:             "  Praktik membaca teks  ",
		CreatedByUsername: " guru.arab ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if store.createParams.AssessmentType != "penugasan" ||
		store.createParams.Title != "Praktik membaca teks" ||
		store.createParams.Mode != "beginner" ||
		store.createParams.Status != "draft" ||
		store.createParams.ScoringScale != "0_100" {
		t.Fatalf("Create params = %+v, want normalized defaults", store.createParams)
	}
	if got := numericForNonTest(t, store.createParams.MaxScore); got != 100 {
		t.Fatalf("MaxScore = %v, want 100", got)
	}
	if got := numericForNonTest(t, store.createParams.Weight); got != 1 {
		t.Fatalf("Weight = %v, want 1", got)
	}
	if string(store.createParams.Checklist) != "[]" {
		t.Fatalf("Checklist = %s, want []", string(store.createParams.Checklist))
	}
}

func TestNonTestAssessmentRejectsInvalidChecklist(t *testing.T) {
	store := &fakeNonTestAssessmentStore{}
	svc := &NonTestAssessment{q: store}
	subjectID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")

	_, err := svc.Create(context.Background(), SaveNonTestAssessmentInput{
		SubjectID: subjectID,
		Title:     "Observasi",
		Checklist: []byte(`{"a":true}`),
	})
	if err == nil {
		t.Fatalf("Create() error = %v, want invalid checklist error", err)
	}
	if store.createParams.Title != "" {
		t.Fatalf("Create called despite invalid checklist: %+v", store.createParams)
	}
}

func TestNonTestAssessmentListIgnoresInvalidFilters(t *testing.T) {
	store := &fakeNonTestAssessmentStore{}
	svc := &NonTestAssessment{q: store}

	_, _, err := svc.List(context.Background(), ListNonTestAssessmentsInput{
		Status:         "selesai",
		AssessmentType: "kiosk",
		SyncFilter:     "stale",
		Limit:          500,
		Offset:         -10,
		SearchQuery:    "  portofolio  ",
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.listParams.StatusFilter != "" || store.listParams.AssessmentTypeFilter != "" {
		t.Fatalf("filters = %q/%q, want invalid filters cleared", store.listParams.StatusFilter, store.listParams.AssessmentTypeFilter)
	}
	if store.listParams.SyncFilter != "" || store.countParams.SyncFilter != "" {
		t.Fatalf("sync filters = %q/%q, want invalid sync filter cleared", store.listParams.SyncFilter, store.countParams.SyncFilter)
	}
	if store.listParams.LimitCount != 25 || store.listParams.OffsetCount != 0 || store.listParams.SearchQuery != "portofolio" {
		t.Fatalf("paging/search = %+v, want clamped limit/offset and trimmed search", store.listParams)
	}
}

func TestNonTestAssessmentListForwardsNeedsSyncFilter(t *testing.T) {
	store := &fakeNonTestAssessmentStore{}
	svc := &NonTestAssessment{q: store}

	_, _, err := svc.List(context.Background(), ListNonTestAssessmentsInput{
		SyncFilter: " needs_sync ",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.listParams.SyncFilter != "needs_sync" || store.countParams.SyncFilter != "needs_sync" {
		t.Fatalf("sync filters = %q/%q, want needs_sync forwarded", store.listParams.SyncFilter, store.countParams.SyncFilter)
	}
}

func TestNonTestAssessmentGetForwardsID(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	store := &fakeNonTestAssessmentStore{
		assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, Title: "Portofolio Quran"},
	}
	svc := &NonTestAssessment{q: store}

	row, err := svc.Get(context.Background(), assessmentID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if store.getAssessmentID != assessmentID || row.ID != assessmentID || row.Title != "Portofolio Quran" {
		t.Fatalf("Get() row/id = %+v/%v, want stored row and forwarded id", row, store.getAssessmentID)
	}
}

func TestNonTestAssessmentUpdatePreservesCreatorAndNormalizesParams(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	store := &fakeNonTestAssessmentStore{
		assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, CreatedByUsername: "guru.awal"},
	}
	svc := &NonTestAssessment{q: store}

	row, err := svc.Update(context.Background(), SaveNonTestAssessmentInput{
		ID:             assessmentID,
		SubjectID:      subjectID,
		ClassID:        classID,
		AssessmentType: " proyek ",
		Title:          "  Proyek Sains  ",
		Mode:           " advance ",
		Status:         " active ",
		MaxScore:       80,
		Weight:         3,
		Checklist:      []byte(`["rubrik"]`),
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if row.ID != assessmentID || store.getAssessmentID != assessmentID {
		t.Fatalf("Update() row/get id = %+v/%v, want assessment id", row, store.getAssessmentID)
	}
	if store.updateParams.ID != assessmentID || store.updateParams.SubjectID != subjectID || store.updateParams.ClassID != classID ||
		store.updateParams.AssessmentType != "proyek" || store.updateParams.Title != "Proyek Sains" ||
		store.updateParams.Mode != "advance" || store.updateParams.Status != "active" || store.updateParams.CreatedByUsername != "guru.awal" {
		t.Fatalf("Update params = %+v, want normalized fields and preserved creator", store.updateParams)
	}
	if got := numericForNonTest(t, store.updateParams.MaxScore); got != 80 {
		t.Fatalf("Update MaxScore = %v, want 80", got)
	}
	if got := numericForNonTest(t, store.updateParams.Weight); got != 3 {
		t.Fatalf("Update Weight = %v, want 3", got)
	}
}

func TestNonTestAssessmentUpdateRejectsInvalidIDBeforeStore(t *testing.T) {
	store := &fakeNonTestAssessmentStore{}
	svc := &NonTestAssessment{q: store}

	_, err := svc.Update(context.Background(), SaveNonTestAssessmentInput{})
	if err == nil || !strings.Contains(err.Error(), "id asesmen") {
		t.Fatalf("Update() error = %v, want id validation", err)
	}
	if store.getAssessmentID.Valid || store.updateParams.ID.Valid {
		t.Fatalf("store called despite invalid id: get=%v update=%+v", store.getAssessmentID, store.updateParams)
	}
}

func TestNonTestAssessmentGenerateSubmissionsUsesAssessmentClassFallback(t *testing.T) {
	store := &fakeNonTestAssessmentStore{}
	svc := &NonTestAssessment{q: store}
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")

	rows, err := svc.GenerateSubmissions(context.Background(), assessmentID, pgtype.UUID{}, pgtype.UUID{})
	if err != nil {
		t.Fatalf("GenerateSubmissions() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("GenerateSubmissions() rows = %d, want 1", len(rows))
	}
	if store.generateParams.AssessmentID != assessmentID {
		t.Fatalf("assessment id = %v, want %v", store.generateParams.AssessmentID, assessmentID)
	}
	if !store.generateParams.ClassID.Valid {
		t.Fatalf("class id = %v, want assessment class forwarded", store.generateParams.ClassID)
	}
}

func TestNonTestAssessmentGenerateSubmissionsRejectsCrossClass(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	classID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	otherClassID := mustUUIDForNonTest(t, "99999999-9999-9999-9999-999999999999")
	store := &fakeNonTestAssessmentStore{
		assessmentRow:              db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333"), ClassID: classID},
		teacherOwnsClassSubjectSet: true,
		teacherOwnsClassSubject:    false,
	}
	svc := &NonTestAssessment{q: store}

	_, err := svc.GenerateSubmissions(context.Background(), assessmentID, otherClassID, pgtype.UUID{})
	if err == nil || !strings.Contains(err.Error(), "class_id harus sesuai") {
		t.Fatalf("GenerateSubmissions() error = %v, want cross-class rejection", err)
	}
	if store.generateParams.AssessmentID.Valid {
		t.Fatalf("generate called despite cross-class input: %+v", store.generateParams)
	}
}

func TestNonTestAssessmentGenerateSubmissionsRejectsOtherTeacherScope(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	classID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	teacherID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	store := &fakeNonTestAssessmentStore{
		assessmentRow:              db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333"), ClassID: classID},
		teacherOwnsClassSubjectSet: true,
		teacherOwnsClassSubject:    false,
	}
	svc := &NonTestAssessment{q: store}

	_, err := svc.GenerateSubmissions(context.Background(), assessmentID, pgtype.UUID{}, teacherID)
	if err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("GenerateSubmissions() error = %v, want teacher scope rejection", err)
	}
}

func TestNonTestAssessmentUpsertSubmissionRejectsScoreAboveMax(t *testing.T) {
	store := &fakeNonTestAssessmentStore{studentBelongs: true, teacherOwnsClassSubjectSet: true, teacherOwnsClassSubject: false}
	svc := &NonTestAssessment{q: store}

	score := 101.0
	_, err := svc.UpsertSubmission(context.Background(), SaveNonTestSubmissionInput{
		AssessmentID: mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111"),
		StudentID:    mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333"),
		Status:       "reviewed",
		Score:        &score,
	})
	if err == nil {
		t.Fatal("UpsertSubmission() error = nil, want score above max error")
	}
	if store.upsertParams.Status != "" {
		t.Fatalf("upsert called despite invalid score: %+v", store.upsertParams)
	}
}

func TestNonTestAssessmentUpsertSubmissionRequiresScoreForReviewed(t *testing.T) {
	store := &fakeNonTestAssessmentStore{studentBelongs: true}
	svc := &NonTestAssessment{q: store}

	_, err := svc.UpsertSubmission(context.Background(), SaveNonTestSubmissionInput{
		AssessmentID: mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111"),
		StudentID:    mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333"),
		Status:       "reviewed",
	})
	if err == nil {
		t.Fatal("UpsertSubmission() error = nil, want missing score error")
	}
}

func TestNonTestAssessmentUpsertSubmissionRejectsStudentOutsideAssessmentClass(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	studentID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	store := &fakeNonTestAssessmentStore{studentBelongs: false}
	svc := &NonTestAssessment{q: store}

	_, err := svc.UpsertSubmission(context.Background(), SaveNonTestSubmissionInput{
		AssessmentID: assessmentID,
		StudentID:    studentID,
		Status:       "assigned",
	})
	if err == nil || !strings.Contains(err.Error(), "siswa tidak berada") {
		t.Fatalf("UpsertSubmission() error = %v, want class membership rejection", err)
	}
	if store.upsertParams.AssessmentID.Valid {
		t.Fatalf("upsert called despite student class mismatch: %+v", store.upsertParams)
	}
}

func TestNonTestAssessmentUpsertSubmissionRejectsOtherTeacherScope(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	studentID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	teacherID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	store := &fakeNonTestAssessmentStore{studentBelongs: true}
	svc := &NonTestAssessment{q: store}

	_, err := svc.UpsertSubmission(context.Background(), SaveNonTestSubmissionInput{
		AssessmentID:      assessmentID,
		StudentID:         studentID,
		Status:            "assigned",
		TeacherEmployeeID: teacherID,
	})
	if err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("UpsertSubmission() error = %v, want teacher scope rejection", err)
	}
	if store.upsertParams.AssessmentID.Valid {
		t.Fatalf("upsert called despite teacher scope mismatch: %+v", store.upsertParams)
	}
}

func TestNonTestAssessmentSyncToGradeCreatesComponentAndEntries(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	assignmentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	studentID := mustUUIDForNonTest(t, "66666666-6666-6666-6666-666666666666")
	store := &fakeNonTestAssessmentStore{
		assessmentRow: db.GetNonTestAssessmentRow{
			ID:             assessmentID,
			SubjectID:      subjectID,
			ClassID:        classID,
			AssessmentType: "praktik",
			Title:          "Praktik Wudhu",
			MaxScore:       pgNumeric(100),
			Weight:         pgNumeric(2),
		},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{
			ID:        assignmentID,
			ClassID:   classID,
			SubjectID: subjectID,
		},
		submissionRows: []db.ListNonTestSubmissionsRow{
			{StudentID: studentID, Status: "reviewed", Score: pgNumeric(87), Feedback: "Sudah baik"},
			{StudentID: uuidForNonTest("77777777-7777-7777-7777-777777777777"), Status: "assigned"},
		},
	}
	svc := &NonTestAssessment{q: store}

	result, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, " guru.fiqih ", true)
	if err != nil {
		t.Fatalf("SyncToGrade() error = %v", err)
	}
	if !result.CreatedComponent || result.SyncedEntries != 1 || result.SkippedEntries != 1 {
		t.Fatalf("SyncToGrade() result = %+v, want one created component and one synced entry", result)
	}
	if store.createGradeComponentParam.AssignmentID != assignmentID ||
		store.createGradeComponentParam.Title != "Non-Tes: Praktik Wudhu" ||
		store.createGradeComponentParam.Category != "practice" ||
		store.createGradeComponentParam.Weight != 2 ||
		store.createGradeComponentParam.MaxScore != 100 ||
		!store.createGradeComponentParam.IsPublished {
		t.Fatalf("CreateGradeComponent params = %+v, want non-test practice component", store.createGradeComponentParam)
	}
	if len(store.upsertGradeParams) != 1 {
		t.Fatalf("grade entries = %d, want 1", len(store.upsertGradeParams))
	}
	entry := store.upsertGradeParams[0]
	if entry.StudentID != studentID || entry.Score.Float64 != 87 || entry.GradedBy != "guru.fiqih" || entry.Notes != "Sudah baik" {
		t.Fatalf("grade entry = %+v, want reviewed submission mapped", entry)
	}
	if store.markGradeSyncParams.ID != assessmentID || !store.markGradeSyncParams.GradeComponentID.Valid || store.markGradeSyncParams.GradeSyncedBy != "guru.fiqih" {
		t.Fatalf("MarkNonTestAssessmentGradeSync params = %+v, want sync marker", store.markGradeSyncParams)
	}
}

func TestNonTestAssessmentSyncToGradeRejectsOtherTeacherAssignment(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	teacherID := mustUUIDForNonTest(t, "88888888-8888-8888-8888-888888888888")
	store := &fakeNonTestAssessmentStore{
		assessmentRow: db.GetNonTestAssessmentRow{
			ID:        assessmentID,
			SubjectID: subjectID,
			ClassID:   classID,
			MaxScore:  pgNumeric(100),
			Weight:    pgNumeric(1),
		},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{
			ID:                mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444"),
			ClassID:           classID,
			SubjectID:         subjectID,
			TeacherEmployeeID: mustUUIDForNonTest(t, "99999999-9999-9999-9999-999999999999"),
		},
		submissionRows: []db.ListNonTestSubmissionsRow{
			{StudentID: mustUUIDForNonTest(t, "66666666-6666-6666-6666-666666666666"), Status: "reviewed", Score: pgNumeric(80)},
		},
	}
	svc := &NonTestAssessment{q: store}

	_, err := svc.SyncToGrade(context.Background(), assessmentID, teacherID, "guru", true)
	if err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("SyncToGrade() error = %v, want access denied", err)
	}
	if len(store.upsertGradeParams) != 0 {
		t.Fatalf("grade entries written despite access denial: %+v", store.upsertGradeParams)
	}
}

func TestNonTestAssessmentSyncToGradeUsesExistingComponentFallbackNotesAndSystemUser(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	assignmentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	componentID := mustUUIDForNonTest(t, "55555555-5555-5555-5555-555555555555")
	studentID := mustUUIDForNonTest(t, "66666666-6666-6666-6666-666666666666")
	store := &fakeNonTestAssessmentStore{
		assessmentRow: db.GetNonTestAssessmentRow{
			ID:               assessmentID,
			SubjectID:        subjectID,
			ClassID:          classID,
			AssessmentType:   "portofolio",
			Title:            "Non-Tes: Hafalan",
			MaxScore:         pgNumeric(100),
			Weight:           pgNumeric(1),
			GradeComponentID: componentID,
		},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		gradeComponent:  db.GradeComponent{ID: componentID, AssignmentID: assignmentID, Title: "Existing", Category: "project", Weight: 1, MaxScore: 100, IsPublished: false},
		submissionRows: []db.ListNonTestSubmissionsRow{
			{StudentID: studentID, Status: "reviewed", Score: pgNumeric(75), EvidenceNote: "catatan bukti"},
		},
	}
	svc := &NonTestAssessment{q: store}

	result, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, "   ", true)
	if err != nil {
		t.Fatalf("SyncToGrade() error = %v", err)
	}
	if result.CreatedComponent || !result.IsPublished || result.GradeComponentID != componentID.String() {
		t.Fatalf("SyncToGrade() result = %+v, want existing component published", result)
	}
	if store.createGradeComponentParam.AssignmentID.Valid {
		t.Fatalf("created component despite reusable component: %+v", store.createGradeComponentParam)
	}
	if store.updatePublishParam.ID != componentID || !store.updatePublishParam.IsPublished {
		t.Fatalf("UpdateGradeComponentPublishState params = %+v, want publish existing component", store.updatePublishParam)
	}
	if len(store.upsertGradeParams) != 1 || store.upsertGradeParams[0].Notes != "catatan bukti" || store.upsertGradeParams[0].GradedBy != "system" {
		t.Fatalf("grade entry = %+v, want evidence-note fallback and system grader", store.upsertGradeParams)
	}
}

func TestNonTestAssessmentSyncToGradeCreatesComponentWhenStoredComponentMissing(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	assignmentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	store := &fakeNonTestAssessmentStore{
		assessmentRow: db.GetNonTestAssessmentRow{
			ID:               assessmentID,
			SubjectID:        subjectID,
			ClassID:          classID,
			AssessmentType:   "lainnya",
			Title:            "Jurnal",
			MaxScore:         pgNumeric(50),
			Weight:           pgNumeric(1),
			GradeComponentID: mustUUIDForNonTest(t, "99999999-9999-9999-9999-999999999999"),
		},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		submissionRows: []db.ListNonTestSubmissionsRow{
			{StudentID: mustUUIDForNonTest(t, "66666666-6666-6666-6666-666666666666"), Status: "reviewed", Score: pgNumeric(40)},
		},
	}
	svc := &NonTestAssessment{q: store}

	result, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, "guru", false)
	if err != nil {
		t.Fatalf("SyncToGrade() error = %v", err)
	}
	if !result.CreatedComponent || store.createGradeComponentParam.Category != "other" || store.createGradeComponentParam.Title != "Non-Tes: Jurnal" {
		t.Fatalf("SyncToGrade/create component = %+v/%+v, want replacement component", result, store.createGradeComponentParam)
	}
}

func TestNonTestAssessmentSyncToGradeRejectsInvalidAndBlockedStates(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	assignmentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	sentinelErr := errors.New("boom")
	tests := []struct {
		name  string
		store *fakeNonTestAssessmentStore
		id    pgtype.UUID
		want  string
	}{
		{name: "invalid assessment id", store: &fakeNonTestAssessmentStore{}, want: "id asesmen"},
		{name: "missing class", id: assessmentID, store: &fakeNonTestAssessmentStore{assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, MaxScore: pgNumeric(100)}}, want: "kelas asesmen"},
		{name: "missing assignment", id: assessmentID, store: &fakeNonTestAssessmentStore{assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, MaxScore: pgNumeric(100)}, gradeAssignmentErr: pgx.ErrNoRows}, want: "assignment kelas-mapel"},
		{name: "assignment lookup error", id: assessmentID, store: &fakeNonTestAssessmentStore{assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, MaxScore: pgNumeric(100)}, gradeAssignmentErr: sentinelErr}, want: "boom"},
		{name: "finalized assignment", id: assessmentID, store: &fakeNonTestAssessmentStore{assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, MaxScore: pgNumeric(100)}, gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID}, finalizationErrSet: true}, want: "difinalisasi"},
		{name: "bad max score", id: assessmentID, store: &fakeNonTestAssessmentStore{assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID}, gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID}}, want: "skor maksimum"},
		{name: "no reviewed entries", id: assessmentID, store: &fakeNonTestAssessmentStore{assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, MaxScore: pgNumeric(100)}, gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID}, submissionRows: []db.ListNonTestSubmissionsRow{{StudentID: uuidForNonTest("66666666-6666-6666-6666-666666666666"), Status: "assigned"}}}, want: "belum ada nilai reviewed"},
		{name: "submission above max", id: assessmentID, store: &fakeNonTestAssessmentStore{assessmentRow: db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, MaxScore: pgNumeric(100)}, gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID}, submissionRows: []db.ListNonTestSubmissionsRow{{StudentID: uuidForNonTest("66666666-6666-6666-6666-666666666666"), Status: "reviewed", Score: pgNumeric(101)}}}, want: "nilai melebihi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &NonTestAssessment{q: tt.store}
			_, err := svc.SyncToGrade(context.Background(), tt.id, pgtype.UUID{}, "guru", true)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("SyncToGrade() error = %v, want containing %q", err, tt.want)
			}
			if len(tt.store.upsertGradeParams) != 0 || tt.store.markGradeSyncParams.ID.Valid {
				t.Fatalf("writes occurred despite error: entries=%+v sync=%+v", tt.store.upsertGradeParams, tt.store.markGradeSyncParams)
			}
		})
	}
}

func TestNonTestAssessmentEnsureGradeComponentPropagatesLookupError(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	assignmentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	componentID := mustUUIDForNonTest(t, "55555555-5555-5555-5555-555555555555")
	store := &fakeNonTestAssessmentStore{getGradeComponentErr: errors.New("component lookup failed")}
	svc := &NonTestAssessment{q: store}

	_, _, err := svc.ensureNonTestGradeComponent(context.Background(), store, db.GetNonTestAssessmentRow{
		ID: assessmentID, GradeComponentID: componentID, Title: "Praktik", AssessmentType: "praktik",
	}, assignmentID, 100, 1, true)
	if err == nil || !strings.Contains(err.Error(), "component lookup failed") {
		t.Fatalf("ensureNonTestGradeComponent() error = %v, want lookup error", err)
	}
}

func TestNewNonTestAssessmentWithPoolInitializesQueriesAndTxStarter(t *testing.T) {
	svc := NewNonTestAssessmentWithPool(nil)
	if svc == nil || svc.q == nil || svc.tx == nil {
		t.Fatalf("NewNonTestAssessmentWithPool(nil) = %+v, want service with queries and tx starter", svc)
	}
}

func TestNonTestAssessmentValidationRejectsBadAssessmentFields(t *testing.T) {
	subjectID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	tests := []struct {
		name  string
		input SaveNonTestAssessmentInput
		want  string
	}{
		{name: "invalid type", input: SaveNonTestAssessmentInput{SubjectID: subjectID, AssessmentType: "ujian", Title: "Judul"}, want: "bentuk asesmen"},
		{name: "missing subject", input: SaveNonTestAssessmentInput{Title: "Judul"}, want: "mata pelajaran"},
		{name: "blank title", input: SaveNonTestAssessmentInput{SubjectID: subjectID, Title: "   "}, want: "judul asesmen"},
		{name: "invalid mode", input: SaveNonTestAssessmentInput{SubjectID: subjectID, Title: "Judul", Mode: "expert"}, want: "mode asesmen"},
		{name: "invalid status", input: SaveNonTestAssessmentInput{SubjectID: subjectID, Title: "Judul", Status: "published"}, want: "status asesmen"},
		{name: "invalid checklist json", input: SaveNonTestAssessmentInput{SubjectID: subjectID, Title: "Judul", Checklist: []byte(`[`)}, want: "checklist observasi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeNonTestAssessmentInput(tt.input)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("normalizeNonTestAssessmentInput() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestNonTestAssessmentSubmissionValidationDefaultsAndRejectsBadFields(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	studentID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	normalized, err := normalizeNonTestSubmissionInput(SaveNonTestSubmissionInput{
		AssessmentID:     assessmentID,
		StudentID:        studentID,
		EvidenceURL:      " https://example.test/evidence ",
		EvidenceNote:     " catatan ",
		Feedback:         " bagus ",
		GradedByUsername: " guru ",
	})
	if err != nil {
		t.Fatalf("normalizeNonTestSubmissionInput() error = %v", err)
	}
	if normalized.Status != "assigned" || normalized.EvidenceURL != "https://example.test/evidence" || normalized.EvidenceNote != "catatan" || normalized.Feedback != "bagus" || normalized.GradedByUsername != "guru" {
		t.Fatalf("normalized submission = %+v, want defaults and trimmed text", normalized)
	}

	negative := -1.0
	tests := []struct {
		name  string
		input SaveNonTestSubmissionInput
		want  string
	}{
		{name: "missing assessment", input: SaveNonTestSubmissionInput{StudentID: studentID}, want: "id asesmen"},
		{name: "missing student", input: SaveNonTestSubmissionInput{AssessmentID: assessmentID}, want: "siswa wajib"},
		{name: "invalid status", input: SaveNonTestSubmissionInput{AssessmentID: assessmentID, StudentID: studentID, Status: "done"}, want: "status pengumpulan"},
		{name: "negative score", input: SaveNonTestSubmissionInput{AssessmentID: assessmentID, StudentID: studentID, Score: &negative}, want: "nilai tidak boleh negatif"},
		{name: "reviewed without score", input: SaveNonTestSubmissionInput{AssessmentID: assessmentID, StudentID: studentID, Status: "reviewed"}, want: "nilai wajib diisi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeNonTestSubmissionInput(tt.input)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("normalizeNonTestSubmissionInput() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestNonTestAssessmentSimpleValidatorsAndHelpers(t *testing.T) {
	svc := &NonTestAssessment{q: &fakeNonTestAssessmentStore{}}
	if err := svc.Delete(context.Background(), pgtype.UUID{}); err == nil || !strings.Contains(err.Error(), "id asesmen") {
		t.Fatalf("Delete(invalid) error = %v, want id validation", err)
	}
	if _, err := svc.ListSubmissions(context.Background(), pgtype.UUID{}); err == nil || !strings.Contains(err.Error(), "id asesmen") {
		t.Fatalf("ListSubmissions(invalid) error = %v, want id validation", err)
	}
	if ok, err := svc.TeacherOwnsClassSubject(context.Background(), pgtype.UUID{}, uuidForNonTest("22222222-2222-2222-2222-222222222222"), uuidForNonTest("33333333-3333-3333-3333-333333333333")); err != nil || ok {
		t.Fatalf("TeacherOwnsClassSubject(invalid) = %v/%v, want false nil", ok, err)
	}
	if ok, err := svc.TeacherOwnsAssessment(context.Background(), uuidForNonTest("11111111-1111-1111-1111-111111111111"), pgtype.UUID{}); err != nil || ok {
		t.Fatalf("TeacherOwnsAssessment(invalid) = %v/%v, want false nil", ok, err)
	}

	if got := nonTestAssessmentGradeTitle("  Non-Tes: Praktik  "); got != "Non-Tes: Praktik" {
		t.Fatalf("nonTestAssessmentGradeTitle(existing) = %q", got)
	}
	if got := nonTestAssessmentGradeTitle(" "); got != "Non-Tes" {
		t.Fatalf("nonTestAssessmentGradeTitle(blank) = %q", got)
	}
	categories := map[string]string{"praktik": "practice", "portofolio": "project", "proyek": "project", "penugasan": "assignment", "observasi": "attitude", "lainnya": "other"}
	for in, want := range categories {
		if got := nonTestAssessmentGradeCategory(in); got != want {
			t.Fatalf("nonTestAssessmentGradeCategory(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNonTestAssessmentUpsertSubmissionSuccessForwardsNormalizedParams(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	studentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	teacherID := mustUUIDForNonTest(t, "55555555-5555-5555-5555-555555555555")
	store := &fakeNonTestAssessmentStore{
		assessmentRow:  db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, MaxScore: pgNumeric(100)},
		studentBelongs: true,
	}
	svc := &NonTestAssessment{q: store}
	score := 88.25

	row, err := svc.UpsertSubmission(context.Background(), SaveNonTestSubmissionInput{
		AssessmentID:      assessmentID,
		StudentID:         studentID,
		Status:            " reviewed ",
		EvidenceURL:       " https://example.test/bukti ",
		EvidenceNote:      " lengkap ",
		Score:             &score,
		Feedback:          " baik ",
		GradedByUsername:  " guru ",
		TeacherEmployeeID: teacherID,
	})
	if err != nil {
		t.Fatalf("UpsertSubmission() error = %v", err)
	}
	if row.AssessmentID != assessmentID || row.StudentID != studentID || row.Status != "reviewed" {
		t.Fatalf("UpsertSubmission() row = %+v, want returned submission ids/status", row)
	}
	if store.upsertParams.Status != "reviewed" || store.upsertParams.EvidenceUrl != "https://example.test/bukti" || store.upsertParams.EvidenceNote != "lengkap" || store.upsertParams.Feedback != "baik" || store.upsertParams.GradedByUsername != "guru" {
		t.Fatalf("Upsert params = %+v, want normalized text forwarded", store.upsertParams)
	}
	if got := numericForNonTest(t, store.upsertParams.Score); got != score {
		t.Fatalf("Upsert score = %v, want %v", got, score)
	}
}

func mustUUIDForNonTest(t *testing.T, raw string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		t.Fatalf("Scan(%q) error = %v", raw, err)
	}
	return id
}

func uuidForNonTest(raw string) pgtype.UUID {
	var id pgtype.UUID
	_ = id.Scan(raw)
	return id
}

func numericForNonTest(t *testing.T, value pgtype.Numeric) float64 {
	t.Helper()
	if !value.Valid || value.Int == nil {
		t.Fatalf("numeric invalid: %+v", value)
	}
	ratio := new(big.Rat).SetInt(value.Int)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absInt32ForNonTest(value.Exp))), nil)
	if value.Exp >= 0 {
		ratio.Mul(ratio, new(big.Rat).SetInt(scale))
	} else {
		ratio.Quo(ratio, new(big.Rat).SetInt(scale))
	}
	out, _ := ratio.Float64()
	return out
}

func absInt32ForNonTest(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}
