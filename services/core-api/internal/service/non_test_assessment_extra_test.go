package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestNonTestAssessmentSyncToGradeDefaultsInvalidWeightAndCountsSkipped(t *testing.T) {
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
			AssessmentType: "observasi",
			Title:          "  Sikap Harian  ",
			MaxScore:       pgNumeric(100),
			Weight:         pgNumeric(0),
		},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		submissionRows: []db.ListNonTestSubmissionsRow{
			{StudentID: studentID, Status: "reviewed", Score: pgNumeric(92), Feedback: "  sangat baik  "},
			{StudentID: uuidForNonTest("77777777-7777-7777-7777-777777777777"), Status: "submitted", Score: pgNumeric(80)},
			{StudentID: uuidForNonTest("88888888-8888-8888-8888-888888888888"), Status: "reviewed"},
		},
	}
	svc := &NonTestAssessment{q: store}

	result, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, " waka ", false)
	if err != nil {
		t.Fatalf("SyncToGrade() error = %v", err)
	}
	if !result.CreatedComponent || result.SyncedEntries != 1 || result.SkippedEntries != 2 || result.IsPublished {
		t.Fatalf("SyncToGrade() result = %+v, want one entry, two skipped, unpublished component", result)
	}
	if store.createGradeComponentParam.Weight != 1 || store.createGradeComponentParam.Category != "attitude" || store.createGradeComponentParam.Title != "Non-Tes: Sikap Harian" {
		t.Fatalf("CreateGradeComponent params = %+v, want default weight and attitude title/category", store.createGradeComponentParam)
	}
	if len(store.upsertGradeParams) != 1 || store.upsertGradeParams[0].Notes != "sangat baik" || store.upsertGradeParams[0].GradedBy != "waka" {
		t.Fatalf("grade params = %+v, want trimmed notes and syncedBy", store.upsertGradeParams)
	}
}

func TestNonTestAssessmentSyncToGradeRejectsFinalizationLookupError(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	assignmentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	store := &fakeNonTestAssessmentStore{
		assessmentRow:      db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, MaxScore: pgNumeric(100), Weight: pgNumeric(1)},
		gradeAssignment:    db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		finalizationErrSet: true,
		finalizationErr:    errors.New("finalization lookup failed"),
	}
	svc := &NonTestAssessment{q: store}

	_, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, "guru", true)
	if err == nil || !strings.Contains(err.Error(), "finalization lookup failed") {
		t.Fatalf("SyncToGrade() error = %v, want finalization lookup error", err)
	}
	if store.createGradeComponentParam.AssignmentID.Valid || len(store.upsertGradeParams) != 0 || store.markGradeSyncParams.ID.Valid {
		t.Fatalf("writes occurred despite finalization lookup error: component=%+v entries=%+v sync=%+v", store.createGradeComponentParam, store.upsertGradeParams, store.markGradeSyncParams)
	}
}

func TestNonTestAssessmentSyncToGradeRejectsNegativeReviewedScore(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	assignmentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	store := &fakeNonTestAssessmentStore{
		assessmentRow:   db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, MaxScore: pgNumeric(100), Weight: pgNumeric(1)},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		submissionRows:  []db.ListNonTestSubmissionsRow{{StudentID: uuidForNonTest("66666666-6666-6666-6666-666666666666"), Status: "reviewed", Score: pgNumeric(-1)}},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &NonTestAssessment{q: store}

	_, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, "guru", true)
	if err == nil || !strings.Contains(err.Error(), "nilai tidak boleh negatif") {
		t.Fatalf("SyncToGrade() error = %v, want negative score error", err)
	}
	if store.createGradeComponentParam.AssignmentID.Valid || len(store.upsertGradeParams) != 0 {
		t.Fatalf("writes occurred despite negative score: component=%+v entries=%+v", store.createGradeComponentParam, store.upsertGradeParams)
	}
}

func TestNonTestAssessmentSyncToGradeRejectsExistingComponentMaxBelowEntry(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	subjectID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	classID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	assignmentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	componentID := mustUUIDForNonTest(t, "55555555-5555-5555-5555-555555555555")
	store := &fakeNonTestAssessmentStore{
		assessmentRow: db.GetNonTestAssessmentRow{
			ID:               assessmentID,
			SubjectID:        subjectID,
			ClassID:          classID,
			MaxScore:         pgNumeric(100),
			Weight:           pgNumeric(1),
			GradeComponentID: componentID,
		},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		gradeComponent:  db.GradeComponent{ID: componentID, AssignmentID: assignmentID, MaxScore: 50, Weight: 1, IsPublished: true},
		submissionRows:  []db.ListNonTestSubmissionsRow{{StudentID: uuidForNonTest("66666666-6666-6666-6666-666666666666"), Status: "reviewed", Score: pgNumeric(75)}},
	}
	svc := &NonTestAssessment{q: store}

	_, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, "guru", true)
	if err == nil || !strings.Contains(err.Error(), "skor maksimum komponen") {
		t.Fatalf("SyncToGrade() error = %v, want component max error", err)
	}
	if len(store.upsertGradeParams) != 0 || store.markGradeSyncParams.ID.Valid {
		t.Fatalf("writes occurred despite component max error: entries=%+v sync=%+v", store.upsertGradeParams, store.markGradeSyncParams)
	}
}

func TestNonTestAssessmentEnsureGradeComponentCreatesWhenStoredComponentAssignmentDiffers(t *testing.T) {
	assessmentID := mustUUIDForNonTest(t, "11111111-1111-1111-1111-111111111111")
	oldAssignmentID := mustUUIDForNonTest(t, "22222222-2222-2222-2222-222222222222")
	newAssignmentID := mustUUIDForNonTest(t, "33333333-3333-3333-3333-333333333333")
	componentID := mustUUIDForNonTest(t, "44444444-4444-4444-4444-444444444444")
	store := &fakeNonTestAssessmentStore{
		gradeComponent: db.GradeComponent{ID: componentID, AssignmentID: oldAssignmentID, MaxScore: 100, Weight: 1, IsPublished: false},
	}
	svc := &NonTestAssessment{q: store}

	component, created, err := svc.ensureNonTestGradeComponent(context.Background(), store, db.GetNonTestAssessmentRow{
		ID: assessmentID, GradeComponentID: componentID, Title: "Proyek", AssessmentType: "proyek",
	}, newAssignmentID, 80, 3, true)
	if err != nil {
		t.Fatalf("ensureNonTestGradeComponent() error = %v", err)
	}
	if !created || component.AssignmentID != newAssignmentID || component.ID == componentID {
		t.Fatalf("ensureNonTestGradeComponent() component=%+v created=%v, want new component for new assignment", component, created)
	}
	if store.updatePublishParam.ID.Valid {
		t.Fatalf("updated old component publish state despite assignment mismatch: %+v", store.updatePublishParam)
	}
	if store.createGradeComponentParam.AssignmentID != newAssignmentID || store.createGradeComponentParam.Category != "project" || store.createGradeComponentParam.MaxScore != 80 || store.createGradeComponentParam.Weight != 3 || !store.createGradeComponentParam.IsPublished {
		t.Fatalf("CreateGradeComponent params = %+v, want replacement project component", store.createGradeComponentParam)
	}
}
