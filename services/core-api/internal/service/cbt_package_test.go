package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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

func TestValidateCbtPackageMetadataInput(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		duration       int32
		sourceMode     string
		drawPgCount    int32
		drawEssayCount int32
		wantErr        string
	}{
		{name: "valid default source mode", title: " PAT IPA ", duration: 0},
		{name: "valid explicit source mode", title: "PAT IPA", duration: 360, sourceMode: "event_pool", drawPgCount: 20, drawEssayCount: 5},
		{name: "blank title", title: "  ", wantErr: "nama paket wajib diisi"},
		{name: "duration too low", title: "PAT IPA", duration: -1, wantErr: "durasi paket CBT"},
		{name: "duration too high", title: "PAT IPA", duration: 361, wantErr: "durasi paket CBT"},
		{name: "invalid source mode", title: "PAT IPA", sourceMode: "student_upload", wantErr: "mode sumber paket tidak valid"},
		{name: "negative pg draw", title: "PAT IPA", drawPgCount: -1, wantErr: "jumlah draw soal tidak boleh negatif"},
		{name: "negative essay draw", title: "PAT IPA", drawEssayCount: -1, wantErr: "jumlah draw soal tidak boleh negatif"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCbtPackageMetadataInput(tt.title, tt.duration, tt.sourceMode, tt.drawPgCount, tt.drawEssayCount)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateCbtPackageMetadataInput() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateCbtPackageMetadataInput() error = %v, want ErrBadRequest containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestCbtPackageUpdateMetadataNormalizesAndTrimsInput(t *testing.T) {
	packageID := pgtype.UUID{Bytes: [16]byte{21}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{22}, Valid: true}
	store := &fakeCbtPackageEditStore{
		lockRow: db.LockCbtPackageForEditRow{ID: packageID, SubjectID: subjectID},
		detailRow: db.GetCbtPackageDetailRow{
			ID:            packageID,
			SubjectID:     subjectID,
			SubjectName:   "IPA",
			SubjectCode:   "IPA",
			Title:         "PAT IPA",
			SessionCount:  2,
			QuestionCount: 2,
		},
		detailQuestions: []db.ListCbtPackageQuestionsByPackageRow{
			{PackageID: packageID, QuestionID: pgtype.UUID{Bytes: [16]byte{23}, Valid: true}, SubjectID: subjectID, QuestionType: "multiple_choice", Status: db.CbtQuestionStatusEnumPublished, TargetLevel: pgtype.Text{String: "VIII", Valid: true}, CpRef: "CP-1", TpRef: "TP-1", CognitiveLevel: "C2", Points: 3},
			{PackageID: packageID, QuestionID: pgtype.UUID{Bytes: [16]byte{24}, Valid: true}, SubjectID: subjectID, QuestionType: "essay", Status: db.CbtQuestionStatusEnumPublished, TargetLevel: pgtype.Text{String: "VIII", Valid: true}, CpRef: "CP-2", KdRef: "KD-1", CognitiveLevel: "C4", Points: 5},
		},
	}
	svc := &CbtPackage{q: store}

	got, err := svc.UpdateMetadata(context.Background(), UpdateCbtPackageInput{
		ID:                 packageID,
		Title:              "  PAT IPA  ",
		Description:        "  paket akhir tahun  ",
		DurationMinutes:    0,
		RandomizeQuestions: true,
		RandomizeOptions:   true,
		SourceMode:         "",
		DrawPgCount:        20,
		DrawEssayCount:     5,
		RandomSeed:         "  seed-1  ",
		IsActive:           true,
	})
	if err != nil {
		t.Fatalf("UpdateMetadata() error = %v", err)
	}
	if store.updateCalls != 1 {
		t.Fatalf("UpdateCbtPackageMetadata calls = %d, want 1", store.updateCalls)
	}
	if store.updateArg.Title != "PAT IPA" || store.updateArg.Description != "paket akhir tahun" || store.updateArg.RandomSeed != "seed-1" {
		t.Fatalf("UpdateCbtPackageMetadata trims = %+v, want trimmed title/description/random seed", store.updateArg)
	}
	if store.updateArg.DurationMinutes != 60 || store.updateArg.SourceMode != "teacher_class" {
		t.Fatalf("UpdateCbtPackageMetadata normalized = %+v, want duration 60 and teacher_class source", store.updateArg)
	}
	if !store.updateArg.RandomizeQuestions || !store.updateArg.RandomizeOptions || !store.updateArg.IsActive || store.updateArg.DrawPgCount != 20 || store.updateArg.DrawEssayCount != 5 {
		t.Fatalf("UpdateCbtPackageMetadata params = %+v, want metadata flags preserved", store.updateArg)
	}
	if got.Package.ID != packageID || got.Readiness.TotalPoints != 8 || got.Readiness.SessionCount != 2 || got.Readiness.MetadataGapCount != 0 {
		t.Fatalf("UpdateMetadata() result = %+v, want refreshed detail/readiness", got)
	}
}

func TestCbtPackageUpdateMetadataRejectsLockedPackage(t *testing.T) {
	packageID := pgtype.UUID{Bytes: [16]byte{31}, Valid: true}
	store := &fakeCbtPackageEditStore{lockRow: db.LockCbtPackageForEditRow{ID: packageID, LockedAt: pgtype.Timestamptz{Time: time.Unix(1700000000, 0), Valid: true}}}
	svc := &CbtPackage{q: store}

	_, err := svc.UpdateMetadata(context.Background(), UpdateCbtPackageInput{ID: packageID, Title: "PAT IPA", DurationMinutes: 90})
	if err == nil || !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "terkunci") {
		t.Fatalf("UpdateMetadata(locked) error = %v, want lock conflict", err)
	}
	if store.updateCalls != 0 {
		t.Fatalf("UpdateCbtPackageMetadata calls = %d, want 0 for locked package", store.updateCalls)
	}
}

func TestCbtPackageReplaceQuestionsValidatesLockAndSelection(t *testing.T) {
	packageID := pgtype.UUID{Bytes: [16]byte{41}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{42}, Valid: true}
	firstQuestionID := pgtype.UUID{Bytes: [16]byte{43}, Valid: true}
	secondQuestionID := pgtype.UUID{Bytes: [16]byte{44}, Valid: true}
	store := &fakeCbtPackageEditStore{
		lockRow: db.LockCbtPackageForEditRow{ID: packageID, SubjectID: subjectID},
		questions: map[pgtype.UUID]db.GetCbtQuestionRow{
			firstQuestionID:  {ID: firstQuestionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
			secondQuestionID: {ID: secondQuestionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
		},
		detailRow: db.GetCbtPackageDetailRow{ID: packageID, SubjectID: subjectID, SubjectName: "IPA", SubjectCode: "IPA"},
	}
	svc := &CbtPackage{q: store}

	_, err := svc.ReplaceQuestions(context.Background(), ReplaceCbtPackageQuestionsInput{
		PackageID:   packageID,
		QuestionIDs: []pgtype.UUID{firstQuestionID, secondQuestionID},
		QuestionWeights: map[string]int32{
			pgUUIDString(firstQuestionID):  2,
			pgUUIDString(secondQuestionID): 4,
		},
	})
	if err != nil {
		t.Fatalf("ReplaceQuestions() error = %v", err)
	}
	if store.deleteQuestionCalls != 1 {
		t.Fatalf("DeleteCbtPackageQuestions calls = %d, want 1", store.deleteQuestionCalls)
	}
	if len(store.addArgs) != 2 || store.addArgs[0].Position != 1 || store.addArgs[0].Points != 2 || store.addArgs[1].Position != 2 || store.addArgs[1].Points != 4 {
		t.Fatalf("AddCbtPackageQuestion args = %+v, want ordered replacement with weights", store.addArgs)
	}

	lockedStore := &fakeCbtPackageEditStore{lockRow: db.LockCbtPackageForEditRow{ID: packageID, LockedAt: pgtype.Timestamptz{Time: time.Unix(1700000000, 0), Valid: true}}}
	_, err = (&CbtPackage{q: lockedStore}).ReplaceQuestions(context.Background(), ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{firstQuestionID}})
	if err == nil || !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "terkunci") {
		t.Fatalf("ReplaceQuestions(locked) error = %v, want lock conflict", err)
	}
	if lockedStore.deleteQuestionCalls != 0 || len(lockedStore.addArgs) != 0 {
		t.Fatalf("ReplaceQuestions(locked) mutating calls = delete %d add %d, want none", lockedStore.deleteQuestionCalls, len(lockedStore.addArgs))
	}
}

func TestCbtPackageCloneTrimsTitleAndRefreshesDetail(t *testing.T) {
	sourceID := pgtype.UUID{Bytes: [16]byte{51}, Valid: true}
	targetID := pgtype.UUID{Bytes: [16]byte{52}, Valid: true}
	store := &fakeCbtPackageEditStore{cloneRow: db.CbtPackage{ID: targetID, Title: "PAT IPA - Revisi"}, detailRow: db.GetCbtPackageDetailRow{ID: targetID, Title: "PAT IPA - Revisi", SubjectName: "IPA", SubjectCode: "IPA"}}
	svc := &CbtPackage{q: store}

	got, err := svc.Clone(context.Background(), CloneCbtPackageInput{SourceID: sourceID, Title: "  PAT IPA - Revisi  "})
	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	if store.cloneArg.SourceID != sourceID || store.cloneArg.Title != "PAT IPA - Revisi" {
		t.Fatalf("CloneCbtPackage arg = %+v, want source and trimmed title", store.cloneArg)
	}
	if store.cloneQuestionsArg.SourceID != sourceID || store.cloneQuestionsArg.TargetID != targetID {
		t.Fatalf("CloneCbtPackageQuestions arg = %+v, want source/target clone", store.cloneQuestionsArg)
	}
	if got.Package.ID != targetID {
		t.Fatalf("Clone() package ID = %v, want %v", got.Package.ID, targetID)
	}
}

func TestCbtPackageReadinessBuildsOverviewStatuses(t *testing.T) {
	lockedAt := pgtype.Timestamptz{Time: time.Unix(1700000000, 0), Valid: true}
	store := &fakeCbtPackageReadinessStore{rows: []db.ListCbtPackageReadinessRow{
		{ID: pgtype.UUID{Bytes: [16]byte{61}, Valid: true}, Title: "Kosong"},
		{ID: pgtype.UUID{Bytes: [16]byte{62}, Valid: true}, Title: "Siap", QuestionCount: 25, PgCount: 20, EssayCount: 5, TotalPoints: 30, PublishedCount: 25},
		{ID: pgtype.UUID{Bytes: [16]byte{63}, Valid: true}, Title: "Terkunci", LockedAt: lockedAt, QuestionCount: 25, PgCount: 20, EssayCount: 5, PublishedCount: 25},
		{ID: pgtype.UUID{Bytes: [16]byte{64}, Valid: true}, Title: "Kurang", QuestionCount: 3, PgCount: 2, EssayCount: 1, PublishedCount: 2, UnpublishedCount: 1, MetadataGapCount: 1},
	}}
	svc := &CbtPackage{q: store}

	got, err := svc.Readiness(context.Background(), pgtype.UUID{})
	if err != nil {
		t.Fatalf("Readiness() error = %v", err)
	}
	if len(got.Items) != 4 {
		t.Fatalf("Readiness() items = %d, want 4", len(got.Items))
	}
	statuses := []string{got.Items[0].Readiness.Status, got.Items[1].Readiness.Status, got.Items[2].Readiness.Status, got.Items[3].Readiness.Status}
	want := []string{"kosong", "siap", "locked", "kurang"}
	for i := range want {
		if statuses[i] != want[i] {
			t.Fatalf("Readiness status[%d] = %q, want %q (all statuses %v)", i, statuses[i], want[i], statuses)
		}
	}
	if !got.Items[1].Readiness.Ready || got.Items[1].Readiness.MissingPgCount != 0 || got.Items[1].Readiness.MissingEssayCount != 0 {
		t.Fatalf("ready item readiness = %+v, want ready with no missing targets", got.Items[1].Readiness)
	}
	if !got.Items[2].Readiness.Locked || !got.Items[2].Readiness.Ready {
		t.Fatalf("locked item readiness = %+v, want locked flag while preserving ready calculation", got.Items[2].Readiness)
	}
}

type fakeCbtPackageEditStore struct {
	lockRow db.LockCbtPackageForEditRow
	lockErr error

	updateArg   db.UpdateCbtPackageMetadataParams
	updateErr   error
	updateCalls int

	questions   map[pgtype.UUID]db.GetCbtQuestionRow
	questionErr error

	deleteQuestionCalls int
	deleteQuestionErr   error
	addArgs             []db.AddCbtPackageQuestionParams
	addErr              error

	cloneArg          db.CloneCbtPackageParams
	cloneRow          db.CbtPackage
	cloneErr          error
	cloneQuestionsArg db.CloneCbtPackageQuestionsParams
	cloneQuestionsErr error

	detailRow       db.GetCbtPackageDetailRow
	detailErr       error
	detailQuestions []db.ListCbtPackageQuestionsByPackageRow
	detailQErr      error
}

func (f *fakeCbtPackageEditStore) ListCbtPackages(context.Context, pgtype.UUID) ([]db.ListCbtPackagesRow, error) {
	return nil, nil
}

func (f *fakeCbtPackageEditStore) ListCbtPackageQuestions(context.Context, pgtype.UUID) ([]db.ListCbtPackageQuestionsRow, error) {
	return nil, nil
}

func (f *fakeCbtPackageEditStore) GetCbtPackageUsage(context.Context, pgtype.UUID) (int32, error) {
	return 0, nil
}

func (f *fakeCbtPackageEditStore) DeleteCbtPackage(context.Context, pgtype.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeCbtPackageEditStore) WithTx(pgx.Tx) *db.Queries { return nil }

func (f *fakeCbtPackageEditStore) LockCbtPackageForEdit(context.Context, pgtype.UUID) (db.LockCbtPackageForEditRow, error) {
	return f.lockRow, f.lockErr
}

func (f *fakeCbtPackageEditStore) UpdateCbtPackageMetadata(_ context.Context, arg db.UpdateCbtPackageMetadataParams) (db.CbtPackage, error) {
	f.updateArg = arg
	f.updateCalls++
	return db.CbtPackage{ID: arg.ID, Title: arg.Title, Description: arg.Description, DurationMinutes: arg.DurationMinutes}, f.updateErr
}

func (f *fakeCbtPackageEditStore) GetCbtQuestion(_ context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	if f.questionErr != nil {
		return db.GetCbtQuestionRow{}, f.questionErr
	}
	if question, ok := f.questions[id]; ok {
		return question, nil
	}
	return db.GetCbtQuestionRow{}, pgx.ErrNoRows
}

func (f *fakeCbtPackageEditStore) DeleteCbtPackageQuestions(context.Context, pgtype.UUID) (int64, error) {
	f.deleteQuestionCalls++
	return 1, f.deleteQuestionErr
}

func (f *fakeCbtPackageEditStore) AddCbtPackageQuestion(_ context.Context, arg db.AddCbtPackageQuestionParams) (int64, error) {
	f.addArgs = append(f.addArgs, arg)
	if f.addErr != nil {
		return 0, f.addErr
	}
	return 1, nil
}

func (f *fakeCbtPackageEditStore) CloneCbtPackage(_ context.Context, arg db.CloneCbtPackageParams) (db.CbtPackage, error) {
	f.cloneArg = arg
	if f.cloneRow.ID.Valid {
		return f.cloneRow, f.cloneErr
	}
	return db.CbtPackage{ID: pgtype.UUID{Bytes: [16]byte{250}, Valid: true}, Title: arg.Title}, f.cloneErr
}

func (f *fakeCbtPackageEditStore) CloneCbtPackageQuestions(_ context.Context, arg db.CloneCbtPackageQuestionsParams) (int64, error) {
	f.cloneQuestionsArg = arg
	return 1, f.cloneQuestionsErr
}

func (f *fakeCbtPackageEditStore) GetCbtPackageDetail(context.Context, pgtype.UUID) (db.GetCbtPackageDetailRow, error) {
	return f.detailRow, f.detailErr
}

func (f *fakeCbtPackageEditStore) ListCbtPackageQuestionsByPackage(context.Context, pgtype.UUID) ([]db.ListCbtPackageQuestionsByPackageRow, error) {
	return f.detailQuestions, f.detailQErr
}

type fakeCbtPackageReadinessStore struct {
	rows []db.ListCbtPackageReadinessRow
	err  error
}

func (f *fakeCbtPackageReadinessStore) ListCbtPackages(context.Context, pgtype.UUID) ([]db.ListCbtPackagesRow, error) {
	return nil, nil
}

func (f *fakeCbtPackageReadinessStore) ListCbtPackageQuestions(context.Context, pgtype.UUID) ([]db.ListCbtPackageQuestionsRow, error) {
	return nil, nil
}

func (f *fakeCbtPackageReadinessStore) GetCbtPackageUsage(context.Context, pgtype.UUID) (int32, error) {
	return 0, nil
}

func (f *fakeCbtPackageReadinessStore) DeleteCbtPackage(context.Context, pgtype.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeCbtPackageReadinessStore) WithTx(pgx.Tx) *db.Queries { return nil }

func (f *fakeCbtPackageReadinessStore) ListCbtPackageReadiness(context.Context, pgtype.UUID) ([]db.ListCbtPackageReadinessRow, error) {
	return f.rows, f.err
}
