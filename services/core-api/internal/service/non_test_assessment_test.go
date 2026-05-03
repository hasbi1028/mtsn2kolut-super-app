package service

import (
	"context"
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeNonTestAssessmentStore struct {
	listParams   db.ListNonTestAssessmentsParams
	countParams  db.CountNonTestAssessmentsParams
	createParams db.CreateNonTestAssessmentParams
	updateParams db.UpdateNonTestAssessmentParams
	createErr    error
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
	return db.GetNonTestAssessmentRow{ID: id, CreatedByUsername: "guru.lama"}, nil
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
	return []db.ListNonTestSubmissionsRow{}, nil
}

func (f *fakeNonTestAssessmentStore) UpsertNonTestSubmission(_ context.Context, arg db.UpsertNonTestSubmissionParams) (db.NonTestAssessmentSubmission, error) {
	return db.NonTestAssessmentSubmission{AssessmentID: arg.AssessmentID, StudentID: arg.StudentID, Status: arg.Status}, nil
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
	if store.listParams.LimitCount != 25 || store.listParams.OffsetCount != 0 || store.listParams.SearchQuery != "portofolio" {
		t.Fatalf("paging/search = %+v, want clamped limit/offset and trimmed search", store.listParams)
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
