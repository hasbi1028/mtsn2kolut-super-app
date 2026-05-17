package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type nonTestAssessmentMoreStore struct {
	*fakeNonTestAssessmentStore
	deleteID             pgtype.UUID
	deleteErr            error
	submissionsID        pgtype.UUID
	submissionsErr       error
	teacherAssessmentArg db.TeacherOwnsNonTestAssessmentParams
	teacherAssessmentOwn bool
	teacherAssessmentErr error
	upsertGradeErr       error
	markGradeSyncErr     error
}

func (s *nonTestAssessmentMoreStore) DeleteNonTestAssessment(ctx context.Context, id pgtype.UUID) error {
	s.deleteID = id
	return s.deleteErr
}

func (s *nonTestAssessmentMoreStore) ListNonTestSubmissions(ctx context.Context, assessmentID pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error) {
	s.submissionsID = assessmentID
	if s.submissionsErr != nil {
		return nil, s.submissionsErr
	}
	return s.submissionRows, nil
}

func (s *nonTestAssessmentMoreStore) TeacherOwnsNonTestAssessment(ctx context.Context, arg db.TeacherOwnsNonTestAssessmentParams) (bool, error) {
	s.teacherAssessmentArg = arg
	if s.teacherAssessmentErr != nil {
		return false, s.teacherAssessmentErr
	}
	return s.teacherAssessmentOwn, nil
}

func (s *nonTestAssessmentMoreStore) UpsertGradeEntry(ctx context.Context, arg db.UpsertGradeEntryParams) (db.GradeEntry, error) {
	s.upsertGradeParams = append(s.upsertGradeParams, arg)
	if s.upsertGradeErr != nil {
		return db.GradeEntry{}, s.upsertGradeErr
	}
	return db.GradeEntry{ComponentID: arg.ComponentID, StudentID: arg.StudentID, Score: arg.Score, Notes: arg.Notes, GradedBy: arg.GradedBy}, nil
}

func (s *nonTestAssessmentMoreStore) MarkNonTestAssessmentGradeSync(ctx context.Context, arg db.MarkNonTestAssessmentGradeSyncParams) (db.NonTestAssessment, error) {
	s.markGradeSyncParams = arg
	if s.markGradeSyncErr != nil {
		return db.NonTestAssessment{}, s.markGradeSyncErr
	}
	return db.NonTestAssessment{ID: arg.ID, GradeComponentID: arg.GradeComponentID, GradeSyncedBy: arg.GradeSyncedBy}, nil
}

func newNonTestMoreStore(base *fakeNonTestAssessmentStore) *nonTestAssessmentMoreStore {
	if base == nil {
		base = &fakeNonTestAssessmentStore{}
	}
	return &nonTestAssessmentMoreStore{fakeNonTestAssessmentStore: base}
}

func TestNonTestAssessmentMoreTeacherOwnsAssessmentForwardsParamsAndErrors(t *testing.T) {
	assessmentID := uuidForNonTest("11111111-1111-1111-1111-111111111111")
	teacherID := uuidForNonTest("22222222-2222-2222-2222-222222222222")
	store := newNonTestMoreStore(nil)
	store.teacherAssessmentOwn = true
	svc := &NonTestAssessment{q: store}

	ok, err := svc.TeacherOwnsAssessment(context.Background(), assessmentID, teacherID)
	if err != nil || !ok {
		t.Fatalf("TeacherOwnsAssessment() = %v/%v, want true nil", ok, err)
	}
	if store.teacherAssessmentArg.ID != assessmentID || store.teacherAssessmentArg.TeacherEmployeeID != teacherID {
		t.Fatalf("TeacherOwnsAssessment arg = %+v, want forwarded assessment/teacher IDs", store.teacherAssessmentArg)
	}

	boom := errors.New("scope lookup failed")
	store = newNonTestMoreStore(nil)
	store.teacherAssessmentErr = boom
	svc = &NonTestAssessment{q: store}
	ok, err = svc.TeacherOwnsAssessment(context.Background(), assessmentID, teacherID)
	if ok || !errors.Is(err, boom) {
		t.Fatalf("TeacherOwnsAssessment(error) = %v/%v, want false/%v", ok, err, boom)
	}
}

func TestNonTestAssessmentMoreDeleteForwardsIDAndPropagatesError(t *testing.T) {
	assessmentID := uuidForNonTest("11111111-1111-1111-1111-111111111111")
	store := newNonTestMoreStore(nil)
	svc := &NonTestAssessment{q: store}

	if err := svc.Delete(context.Background(), assessmentID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleteID != assessmentID {
		t.Fatalf("Delete() id = %v, want %v", store.deleteID, assessmentID)
	}

	deleteErr := errors.New("delete failed")
	store = newNonTestMoreStore(nil)
	store.deleteErr = deleteErr
	svc = &NonTestAssessment{q: store}
	if err := svc.Delete(context.Background(), assessmentID); !errors.Is(err, deleteErr) {
		t.Fatalf("Delete(error) = %v, want %v", err, deleteErr)
	}
}

func TestNonTestAssessmentMoreListSubmissionsForwardsIDRowsAndErrors(t *testing.T) {
	assessmentID := uuidForNonTest("11111111-1111-1111-1111-111111111111")
	studentID := uuidForNonTest("22222222-2222-2222-2222-222222222222")
	store := newNonTestMoreStore(&fakeNonTestAssessmentStore{submissionRows: []db.ListNonTestSubmissionsRow{{AssessmentID: assessmentID, StudentID: studentID, Status: "reviewed"}}})
	svc := &NonTestAssessment{q: store}

	rows, err := svc.ListSubmissions(context.Background(), assessmentID)
	if err != nil {
		t.Fatalf("ListSubmissions() error = %v", err)
	}
	if store.submissionsID != assessmentID || len(rows) != 1 || rows[0].StudentID != studentID {
		t.Fatalf("ListSubmissions() id/rows = %v/%+v, want stored reviewed row", store.submissionsID, rows)
	}

	listErr := errors.New("list submissions failed")
	store = newNonTestMoreStore(nil)
	store.submissionsErr = listErr
	svc = &NonTestAssessment{q: store}
	if _, err := svc.ListSubmissions(context.Background(), assessmentID); !errors.Is(err, listErr) {
		t.Fatalf("ListSubmissions(error) = %v, want %v", err, listErr)
	}
}

func TestNonTestAssessmentMoreSyncToGradePropagatesUpsertErrorBeforeMark(t *testing.T) {
	assessmentID := uuidForNonTest("11111111-1111-1111-1111-111111111111")
	subjectID := uuidForNonTest("22222222-2222-2222-2222-222222222222")
	classID := uuidForNonTest("33333333-3333-3333-3333-333333333333")
	assignmentID := uuidForNonTest("44444444-4444-4444-4444-444444444444")
	upsertErr := errors.New("grade upsert failed")
	store := newNonTestMoreStore(&fakeNonTestAssessmentStore{
		assessmentRow:   db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, AssessmentType: "praktik", Title: "Praktik", MaxScore: pgNumeric(100), Weight: pgNumeric(1)},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		submissionRows:  []db.ListNonTestSubmissionsRow{{StudentID: uuidForNonTest("55555555-5555-5555-5555-555555555555"), Status: "reviewed", Score: pgNumeric(80)}},
	})
	store.upsertGradeErr = upsertErr
	svc := &NonTestAssessment{q: store}

	_, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, "guru", true)
	if !errors.Is(err, upsertErr) {
		t.Fatalf("SyncToGrade(upsert error) = %v, want %v", err, upsertErr)
	}
	if len(store.upsertGradeParams) != 1 || store.markGradeSyncParams.ID.Valid {
		t.Fatalf("SyncToGrade(upsert error) writes = entries %+v mark %+v, want one attempted entry and no mark", store.upsertGradeParams, store.markGradeSyncParams)
	}
}

func TestNonTestAssessmentMoreSyncToGradePropagatesMarkErrorAfterEntries(t *testing.T) {
	assessmentID := uuidForNonTest("11111111-1111-1111-1111-111111111111")
	subjectID := uuidForNonTest("22222222-2222-2222-2222-222222222222")
	classID := uuidForNonTest("33333333-3333-3333-3333-333333333333")
	assignmentID := uuidForNonTest("44444444-4444-4444-4444-444444444444")
	markErr := errors.New("mark failed")
	store := newNonTestMoreStore(&fakeNonTestAssessmentStore{
		assessmentRow:   db.GetNonTestAssessmentRow{ID: assessmentID, SubjectID: subjectID, ClassID: classID, AssessmentType: "observasi", Title: "Sikap", MaxScore: pgNumeric(100), Weight: pgNumeric(1)},
		gradeAssignment: db.GetGradeAssignmentByClassSubjectRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		submissionRows:  []db.ListNonTestSubmissionsRow{{StudentID: uuidForNonTest("55555555-5555-5555-5555-555555555555"), Status: "reviewed", Score: pgNumeric(90), Feedback: " baik "}},
	})
	store.markGradeSyncErr = markErr
	svc := &NonTestAssessment{q: store}

	_, err := svc.SyncToGrade(context.Background(), assessmentID, pgtype.UUID{}, " guru ", false)
	if !errors.Is(err, markErr) {
		t.Fatalf("SyncToGrade(mark error) = %v, want %v", err, markErr)
	}
	if len(store.upsertGradeParams) != 1 || store.upsertGradeParams[0].Notes != "baik" || store.markGradeSyncParams.ID != assessmentID {
		t.Fatalf("SyncToGrade(mark error) entries/mark = %+v/%+v, want entry written then mark attempted", store.upsertGradeParams, store.markGradeSyncParams)
	}
}

func TestNonTestAssessmentMoreListKeepsValidFiltersAndTeacherScope(t *testing.T) {
	subjectID := uuidForNonTest("11111111-1111-1111-1111-111111111111")
	classID := uuidForNonTest("22222222-2222-2222-2222-222222222222")
	teacherID := uuidForNonTest("33333333-3333-3333-3333-333333333333")
	store := &fakeNonTestAssessmentStore{}
	svc := &NonTestAssessment{q: store}

	_, _, err := svc.List(context.Background(), ListNonTestAssessmentsInput{
		SubjectID:         subjectID,
		ClassID:           classID,
		Status:            " active ",
		AssessmentType:    " praktik ",
		SearchQuery:       "  wudhu  ",
		Limit:             100,
		Offset:            5,
		TeacherEmployeeID: teacherID,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.listParams.SubjectID != subjectID || store.listParams.ClassID != classID || store.listParams.StatusFilter != "active" || store.listParams.AssessmentTypeFilter != "praktik" || store.listParams.SearchQuery != "wudhu" || store.listParams.LimitCount != 100 || store.listParams.OffsetCount != 5 || store.listParams.TeacherEmployeeID != teacherID {
		t.Fatalf("List() params = %+v, want normalized valid filters and teacher scope", store.listParams)
	}
	if store.countParams.TeacherEmployeeID != teacherID || store.countParams.SearchQuery != "wudhu" {
		t.Fatalf("Count params = %+v, want same teacher/search filters", store.countParams)
	}
}

func TestNonTestAssessmentMoreNormalizeSubmissionRejectsInvalidScoreWithMessage(t *testing.T) {
	assessmentID := uuidForNonTest("11111111-1111-1111-1111-111111111111")
	studentID := uuidForNonTest("22222222-2222-2222-2222-222222222222")
	negative := -0.25
	_, err := normalizeNonTestSubmissionInput(SaveNonTestSubmissionInput{AssessmentID: assessmentID, StudentID: studentID, Status: "submitted", Score: &negative})
	if err == nil || !strings.Contains(err.Error(), "nilai tidak boleh negatif") {
		t.Fatalf("normalizeNonTestSubmissionInput(negative) error = %v, want negative score message", err)
	}
}
