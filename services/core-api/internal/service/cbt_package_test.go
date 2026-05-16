package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeCbtPackageCreateStore struct {
	createParams db.CreateCbtPackageParams
	createRow    db.CbtPackage
	createErr    error
	createCalls  int

	questions   map[pgtype.UUID]db.GetCbtQuestionRow
	questionErr error

	addParams []db.AddCbtPackageQuestionParams
	addErr    error
}

func TestCbtPackageDeleteRejectsSessionUsage(t *testing.T) {
	packageID := pgtype.UUID{Bytes: [16]byte{8}, Valid: true}
	store := &fakeCbtPackageStore{usageCount: 2}
	svc := &CbtPackage{q: store}

	err := svc.Delete(context.Background(), packageID)
	if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "sesi ujian") {
		t.Fatalf("Delete(used package) error = %v, want usage conflict", err)
	}
	if store.deleteID.Valid {
		t.Fatalf("Delete(used package) delete id = %v, want no delete", store.deleteID)
	}
}

func (f *fakeCbtPackageCreateStore) CreateCbtPackage(_ context.Context, arg db.CreateCbtPackageParams) (db.CbtPackage, error) {
	f.createParams = arg
	f.createCalls++
	if f.createErr != nil {
		return db.CbtPackage{}, f.createErr
	}
	return f.createRow, nil
}

func (f *fakeCbtPackageCreateStore) GetCbtQuestion(_ context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	if f.questionErr != nil {
		return db.GetCbtQuestionRow{}, f.questionErr
	}
	return f.questions[id], nil
}

func (f *fakeCbtPackageCreateStore) AddCbtPackageQuestion(_ context.Context, arg db.AddCbtPackageQuestionParams) (int64, error) {
	f.addParams = append(f.addParams, arg)
	if f.addErr != nil {
		return 0, f.addErr
	}
	return 1, nil
}

func TestCreateCbtPackageRequiresEligibleOfficialQuestions(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	packageID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	publishedQuestionID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	draftQuestionID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	approvedQuestionID := pgtype.UUID{Bytes: [16]byte{5}, Valid: true}
	workflowPublishedQuestionID := pgtype.UUID{Bytes: [16]byte{6}, Valid: true}

	t.Run("accepts published questions", func(t *testing.T) {
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				publishedQuestionID: {
					ID:        publishedQuestionID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumPublished,
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			SubjectID:   subjectID,
			Title:       "PAT IPA",
			QuestionIDs: []pgtype.UUID{publishedQuestionID},
		})
		if err != nil {
			t.Fatalf("createCbtPackage() error = %v", err)
		}
		if len(store.addParams) != 1 || store.addParams[0].QuestionID != publishedQuestionID {
			t.Fatalf("AddCbtPackageQuestion() params = %+v, want published question added", store.addParams)
		}
	})

	t.Run("accepts approved workflow questions before publish", func(t *testing.T) {
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				approvedQuestionID: {
					ID:             approvedQuestionID,
					SubjectID:      subjectID,
					Status:         db.CbtQuestionStatusEnumDraft,
					WorkflowStatus: "approved",
				},
				workflowPublishedQuestionID: {
					ID:             workflowPublishedQuestionID,
					SubjectID:      subjectID,
					Status:         db.CbtQuestionStatusEnumDraft,
					WorkflowStatus: "published",
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			SubjectID:   subjectID,
			Title:       "PAT IPA",
			QuestionIDs: []pgtype.UUID{approvedQuestionID, workflowPublishedQuestionID},
		})
		if err != nil {
			t.Fatalf("createCbtPackage() error = %v", err)
		}
		if len(store.addParams) != 2 || store.addParams[0].QuestionID != approvedQuestionID || store.addParams[1].QuestionID != workflowPublishedQuestionID {
			t.Fatalf("AddCbtPackageQuestion() params = %+v, want approved/published workflow questions added", store.addParams)
		}
	})

	t.Run("rejects unsafe workflow questions", func(t *testing.T) {
		unsafeStatuses := []string{"draft", "submitted", "revision_needed", "rejected", "archived"}
		for _, workflowStatus := range unsafeStatuses {
			t.Run(workflowStatus, func(t *testing.T) {
				store := &fakeCbtPackageCreateStore{
					createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
					questions: map[pgtype.UUID]db.GetCbtQuestionRow{
						draftQuestionID: {
							ID:             draftQuestionID,
							SubjectID:      subjectID,
							Status:         db.CbtQuestionStatusEnumDraft,
							WorkflowStatus: workflowStatus,
						},
					},
				}

				_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
					SubjectID:   subjectID,
					Title:       "PAT IPA",
					QuestionIDs: []pgtype.UUID{draftQuestionID},
				})
				if err == nil || !strings.Contains(err.Error(), "draft/submitted/revision_needed/rejected") {
					t.Fatalf("createCbtPackage() error = %v, want unsafe workflow rejection", err)
				}
				if len(store.addParams) != 0 {
					t.Fatalf("AddCbtPackageQuestion() calls = %d, want 0", len(store.addParams))
				}
			})
		}
	})

	t.Run("rejects draft questions", func(t *testing.T) {
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				draftQuestionID: {
					ID:        draftQuestionID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumDraft,
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			SubjectID:   subjectID,
			Title:       "PAT IPA",
			QuestionIDs: []pgtype.UUID{draftQuestionID},
		})
		if err == nil || !strings.Contains(err.Error(), "approved/published") {
			t.Fatalf("createCbtPackage() error = %v, want unpublished question rejection", err)
		}
		if len(store.addParams) != 0 {
			t.Fatalf("AddCbtPackageQuestion() calls = %d, want 0", len(store.addParams))
		}
	})

	t.Run("rejects event questions in global package", func(t *testing.T) {
		eventID := pgtype.UUID{Bytes: [16]byte{5}, Valid: true}
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				publishedQuestionID: {
					ID:        publishedQuestionID,
					EventID:   eventID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumPublished,
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			SubjectID:   subjectID,
			Title:       "Paket Umum IPA",
			QuestionIDs: []pgtype.UUID{publishedQuestionID},
		})
		if err == nil || !strings.Contains(err.Error(), "paket umum") {
			t.Fatalf("createCbtPackage() error = %v, want global package event-question rejection", err)
		}
		if len(store.addParams) != 0 {
			t.Fatalf("AddCbtPackageQuestion() calls = %d, want 0", len(store.addParams))
		}
	})
}

func TestCreateCbtPackageEventQuestionScope(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	packageID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	eventID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	otherEventID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	globalQuestionID := pgtype.UUID{Bytes: [16]byte{5}, Valid: true}
	eventQuestionID := pgtype.UUID{Bytes: [16]byte{6}, Valid: true}
	otherEventQuestionID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}

	t.Run("event package accepts reusable global and same-event questions", func(t *testing.T) {
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, EventID: eventID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				globalQuestionID: {
					ID:        globalQuestionID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumPublished,
				},
				eventQuestionID: {
					ID:        eventQuestionID,
					EventID:   eventID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumPublished,
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			EventID:     eventID,
			SubjectID:   subjectID,
			Title:       "Paket Event IPA",
			QuestionIDs: []pgtype.UUID{globalQuestionID, eventQuestionID},
		})
		if err != nil {
			t.Fatalf("createCbtPackage() error = %v", err)
		}
		if len(store.addParams) != 2 {
			t.Fatalf("AddCbtPackageQuestion() calls = %d, want 2", len(store.addParams))
		}
		if store.addParams[0].QuestionID != globalQuestionID || store.addParams[1].QuestionID != eventQuestionID {
			t.Fatalf("AddCbtPackageQuestion() params = %+v, want global then same-event questions", store.addParams)
		}
	})

	t.Run("event package rejects questions from another event", func(t *testing.T) {
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, EventID: eventID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				otherEventQuestionID: {
					ID:        otherEventQuestionID,
					EventID:   otherEventID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumPublished,
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			EventID:     eventID,
			SubjectID:   subjectID,
			Title:       "Paket Event IPA",
			QuestionIDs: []pgtype.UUID{otherEventQuestionID},
		})
		if err == nil || !strings.Contains(err.Error(), "event yang sama") {
			t.Fatalf("createCbtPackage() error = %v, want other-event question rejection", err)
		}
		if len(store.addParams) != 0 {
			t.Fatalf("AddCbtPackageQuestion() calls = %d, want 0", len(store.addParams))
		}
	})
}

func TestCbtPackageReadinessCountsMissingTargetLevelAsMetadataGap(t *testing.T) {
	questions := []db.ListCbtPackageQuestionsByPackageRow{
		{
			QuestionID:     pgtype.UUID{Bytes: [16]byte{8}, Valid: true},
			QuestionType:   "multiple_choice",
			Status:         db.CbtQuestionStatusEnumPublished,
			TargetLevel:    pgtype.Text{},
			CpRef:          "CP-1",
			TpRef:          "TP-1",
			CognitiveLevel: "C2",
			Points:         1,
		},
		{
			QuestionID:     pgtype.UUID{Bytes: [16]byte{9}, Valid: true},
			QuestionType:   "essay",
			Status:         db.CbtQuestionStatusEnumPublished,
			TargetLevel:    pgtype.Text{String: "VIII", Valid: true},
			CpRef:          "CP-1",
			KdRef:          "KD-1",
			CognitiveLevel: "C4",
			Points:         2,
		},
	}

	readiness := cbtPackageReadinessStatusFromQuestions(questions, 0, false)

	if readiness.MetadataGapCount != 1 {
		t.Fatalf("MetadataGapCount = %d, want 1 for missing target_level", readiness.MetadataGapCount)
	}
	if readiness.PgCount != 1 || readiness.EssayCount != 1 || readiness.TotalPoints != 3 {
		t.Fatalf("readiness counts = %+v, want package counts preserved", readiness)
	}
}
