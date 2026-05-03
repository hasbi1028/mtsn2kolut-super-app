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

type fakeGradeStore struct {
	highestScore          float64
	highestScoreErr       error
	createArg             db.CreateGradeComponentParams
	createErr             error
	updateArg             db.UpdateGradeComponentParams
	updateErr             error
	publishArg            db.UpdateGradeComponentPublishStateParams
	publishErr            error
	deleteComponentID     pgtype.UUID
	deleteComponentErr    error
	upsertEntryArg        db.UpsertGradeEntryParams
	upsertEntryErr        error
	updateComponent       db.GradeComponent
	assignments           []db.ListClassSubjectAssignmentsRow
	assignmentsErr        error
	assignmentStatuses    []db.ListGradeAssignmentStatusesRow
	assignmentStatusesErr error
	component             db.GradeComponent
	componentErr          error
	nonTestSourceID       pgtype.UUID
	nonTestSourceErr      error
	listComponents        []db.ListGradeComponentsRow
	listComponentsErr     error
	listSummary           []db.ListGradebookSummaryRow
	listSummaryErr        error
	listEntries           []db.ListGradeEntriesByComponentRow
	listEntriesErr        error
	finalization          db.GradeAssignmentFinalization
	finalizationErr       error
	finalizationUpsertArg db.UpsertGradeAssignmentFinalizationParams
	finalizationDeleteID  pgtype.UUID
	finalizationDeleteErr error
}

func (f *fakeGradeStore) ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error) {
	return f.assignments, f.assignmentsErr
}

func (f *fakeGradeStore) ListGradeAssignmentStatuses(ctx context.Context) ([]db.ListGradeAssignmentStatusesRow, error) {
	return f.assignmentStatuses, f.assignmentStatusesErr
}

func (f *fakeGradeStore) ListGradeComponents(ctx context.Context, arg db.ListGradeComponentsParams) ([]db.ListGradeComponentsRow, error) {
	return f.listComponents, f.listComponentsErr
}

func (f *fakeGradeStore) GetGradeComponent(ctx context.Context, id pgtype.UUID) (db.GradeComponent, error) {
	return f.component, f.componentErr
}

func (f *fakeGradeStore) GetNonTestGradeComponentSource(ctx context.Context, gradeComponentID pgtype.UUID) (pgtype.UUID, error) {
	if f.nonTestSourceErr != nil {
		return pgtype.UUID{}, f.nonTestSourceErr
	}
	if f.nonTestSourceID.Valid {
		return f.nonTestSourceID, nil
	}
	return pgtype.UUID{}, pgx.ErrNoRows
}

func (f *fakeGradeStore) GetGradeComponentHighestScore(ctx context.Context, componentID pgtype.UUID) (float64, error) {
	return f.highestScore, f.highestScoreErr
}

func (f *fakeGradeStore) GetGradeAssignmentFinalization(ctx context.Context, assignmentID pgtype.UUID) (db.GradeAssignmentFinalization, error) {
	if f.finalizationErr != nil {
		return db.GradeAssignmentFinalization{}, f.finalizationErr
	}
	return f.finalization, nil
}

func (f *fakeGradeStore) UpsertGradeAssignmentFinalization(ctx context.Context, arg db.UpsertGradeAssignmentFinalizationParams) (db.GradeAssignmentFinalization, error) {
	f.finalizationUpsertArg = arg
	return db.GradeAssignmentFinalization{
		AssignmentID: arg.AssignmentID,
		FinalizedBy:  arg.FinalizedBy,
		Notes:        arg.Notes,
	}, nil
}

func (f *fakeGradeStore) DeleteGradeAssignmentFinalization(ctx context.Context, assignmentID pgtype.UUID) error {
	f.finalizationDeleteID = assignmentID
	return f.finalizationDeleteErr
}

func (f *fakeGradeStore) CreateGradeComponent(ctx context.Context, arg db.CreateGradeComponentParams) (db.GradeComponent, error) {
	f.createArg = arg
	if f.createErr != nil {
		return db.GradeComponent{}, f.createErr
	}
	return db.GradeComponent{AssignmentID: arg.AssignmentID, Title: arg.Title, Category: arg.Category, Weight: arg.Weight, MaxScore: arg.MaxScore}, nil
}

func (f *fakeGradeStore) UpdateGradeComponent(ctx context.Context, arg db.UpdateGradeComponentParams) (db.GradeComponent, error) {
	f.updateArg = arg
	if f.updateErr != nil {
		return db.GradeComponent{}, f.updateErr
	}
	return f.updateComponent, nil
}

func (f *fakeGradeStore) UpdateGradeComponentPublishState(ctx context.Context, arg db.UpdateGradeComponentPublishStateParams) (db.GradeComponent, error) {
	f.publishArg = arg
	if f.publishErr != nil {
		return db.GradeComponent{}, f.publishErr
	}
	return db.GradeComponent{ID: arg.ID, AssignmentID: f.component.AssignmentID, Title: f.component.Title, IsPublished: arg.IsPublished}, nil
}

func (f *fakeGradeStore) DeleteGradeComponent(ctx context.Context, id pgtype.UUID) error {
	f.deleteComponentID = id
	return f.deleteComponentErr
}

func (f *fakeGradeStore) ListGradebookSummary(ctx context.Context, arg db.ListGradebookSummaryParams) ([]db.ListGradebookSummaryRow, error) {
	return f.listSummary, f.listSummaryErr
}

func (f *fakeGradeStore) ListGradeEntriesByComponent(ctx context.Context, componentID pgtype.UUID) ([]db.ListGradeEntriesByComponentRow, error) {
	return f.listEntries, f.listEntriesErr
}

func (f *fakeGradeStore) UpsertGradeEntry(ctx context.Context, arg db.UpsertGradeEntryParams) (db.GradeEntry, error) {
	f.upsertEntryArg = arg
	if f.upsertEntryErr != nil {
		return db.GradeEntry{}, f.upsertEntryErr
	}
	return db.GradeEntry{ComponentID: arg.ComponentID, StudentID: arg.StudentID, Score: arg.Score, Notes: arg.Notes, GradedBy: arg.GradedBy}, nil
}

func TestGradeUpdateComponentRejectsMaxScoreBelowHighestExistingScore(t *testing.T) {
	store := &fakeGradeStore{
		highestScore:    88,
		component:       db.GradeComponent{AssignmentID: pgtype.UUID{}},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &Grade{q: store}

	_, err := svc.UpdateComponent(context.Background(), db.UpdateGradeComponentParams{
		ID:       pgtype.UUID{},
		Title:    "Ulangan Harian 1",
		Category: "quiz",
		Weight:   1,
		MaxScore: 80,
	}, pgtype.UUID{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nilai tertinggi 88.00") {
		t.Fatalf("error = %q, want highest-score guard", err.Error())
	}
}

func TestGradeUpdateComponentNormalizesAndForwardsValues(t *testing.T) {
	store := &fakeGradeStore{
		highestScore:    -1,
		component:       db.GradeComponent{AssignmentID: pgtype.UUID{}},
		finalizationErr: pgx.ErrNoRows,
		updateComponent: db.GradeComponent{
			Title:       "Tugas Proyek",
			Category:    "project",
			Weight:      2,
			MaxScore:    100,
			IsPublished: false,
		},
	}
	svc := &Grade{q: store}

	_, err := svc.UpdateComponent(context.Background(), db.UpdateGradeComponentParams{
		ID:       pgtype.UUID{},
		Title:    "  Tugas Proyek  ",
		Category: "PROJECT",
		Weight:   2,
		MaxScore: 100,
	}, pgtype.UUID{})
	if err != nil {
		t.Fatalf("UpdateComponent() error = %v", err)
	}
	if store.updateArg.Title != "Tugas Proyek" {
		t.Fatalf("title = %q, want %q", store.updateArg.Title, "Tugas Proyek")
	}
	if store.updateArg.Category != "project" {
		t.Fatalf("category = %q, want %q", store.updateArg.Category, "project")
	}
}

func TestGradeRejectsDirectMutationForNonTestSourcedComponent(t *testing.T) {
	assignmentID := pgtype.UUID{Bytes: [16]byte{61}, Valid: true}
	componentID := pgtype.UUID{Bytes: [16]byte{62}, Valid: true}
	studentID := pgtype.UUID{Bytes: [16]byte{63}, Valid: true}
	store := &fakeGradeStore{
		component:       db.GradeComponent{ID: componentID, AssignmentID: assignmentID, Title: "Praktik", MaxScore: 100},
		finalizationErr: pgx.ErrNoRows,
		nonTestSourceID: pgtype.UUID{Bytes: [16]byte{64}, Valid: true},
	}
	svc := &Grade{q: store}

	_, err := svc.UpdateComponent(context.Background(), db.UpdateGradeComponentParams{
		ID:       componentID,
		Title:    "Praktik Revisi",
		Category: "practice",
		Weight:   1,
		MaxScore: 100,
	}, pgtype.UUID{})
	if err == nil || !strings.Contains(err.Error(), "asesmen non-tes") {
		t.Fatalf("UpdateComponent() error = %v, want non-test source guard", err)
	}
	if store.updateArg.ID.Valid {
		t.Fatal("UpdateComponent() forwarded repository update for non-test sourced component")
	}

	err = svc.DeleteComponent(context.Background(), componentID, pgtype.UUID{})
	if err == nil || !strings.Contains(err.Error(), "asesmen non-tes") {
		t.Fatalf("DeleteComponent() error = %v, want non-test source guard", err)
	}
	if store.deleteComponentID.Valid {
		t.Fatal("DeleteComponent() forwarded repository delete for non-test sourced component")
	}

	_, err = svc.UpsertEntry(context.Background(), componentID, studentID, pgtype.UUID{}, 88, "catatan", "guru")
	if err == nil || !strings.Contains(err.Error(), "asesmen non-tes") {
		t.Fatalf("UpsertEntry() error = %v, want non-test source guard", err)
	}
	if store.upsertEntryArg.ComponentID.Valid {
		t.Fatal("UpsertEntry() forwarded repository entry write for non-test sourced component")
	}
}

func TestGradeFinalizeAssignmentRejectsWhenNotReady(t *testing.T) {
	assignmentID := pgtype.UUID{Valid: true}
	store := &fakeGradeStore{
		listComponents: []db.ListGradeComponentsRow{
			{IsPublished: false},
		},
		listSummary: []db.ListGradebookSummaryRow{
			{ComponentCount: 1, FilledCount: 0},
		},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &Grade{q: store}

	_, err := svc.FinalizeAssignment(context.Background(), assignmentID, "guru-a", "cek awal", pgtype.UUID{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "belum siap difinalisasi") {
		t.Fatalf("error = %q, want readiness guard", err.Error())
	}
}

func TestGradeFinalizeAssignmentPersistsCheckpointWhenReady(t *testing.T) {
	assignmentID := pgtype.UUID{Valid: true}
	store := &fakeGradeStore{
		listComponents: []db.ListGradeComponentsRow{
			{IsPublished: true},
		},
		listSummary: []db.ListGradebookSummaryRow{
			{ComponentCount: 1, FilledCount: 1},
		},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &Grade{q: store}

	_, err := svc.FinalizeAssignment(context.Background(), assignmentID, "guru-a", "siap cetak", pgtype.UUID{})
	if err != nil {
		t.Fatalf("FinalizeAssignment() error = %v", err)
	}
	if store.finalizationUpsertArg.FinalizedBy != "guru-a" {
		t.Fatalf("finalized_by = %q, want %q", store.finalizationUpsertArg.FinalizedBy, "guru-a")
	}
	if store.finalizationUpsertArg.Notes != "siap cetak" {
		t.Fatalf("notes = %q, want %q", store.finalizationUpsertArg.Notes, "siap cetak")
	}
}

func TestGradeCreateComponentRejectsWhenAssignmentAlreadyFinalized(t *testing.T) {
	assignmentID := pgtype.UUID{Valid: true}
	store := &fakeGradeStore{
		finalizationErr: nil,
		finalization:    db.GradeAssignmentFinalization{AssignmentID: assignmentID, FinalizedBy: "guru-a"},
	}
	svc := &Grade{q: store}

	_, err := svc.CreateComponent(context.Background(), db.CreateGradeComponentParams{
		AssignmentID: assignmentID,
		Title:        "UH 1",
		Category:     "quiz",
		Weight:       1,
		MaxScore:     100,
	}, pgtype.UUID{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "sudah difinalisasi") {
		t.Fatalf("error = %q, want finalized guard", err.Error())
	}
}

func TestGradeReopenAssignmentDeletesFinalization(t *testing.T) {
	assignmentID := pgtype.UUID{Valid: true}
	store := &fakeGradeStore{finalizationErr: errors.New("unused")}
	svc := &Grade{q: store}

	if err := svc.ReopenAssignment(context.Background(), assignmentID, pgtype.UUID{}); err != nil {
		t.Fatalf("ReopenAssignment() error = %v", err)
	}
	if store.finalizationDeleteID != assignmentID {
		t.Fatal("delete finalization assignment id was not forwarded")
	}
}

func TestGradeOverviewBuildsAssignmentStatuses(t *testing.T) {
	assignmentID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	store := &fakeGradeStore{
		assignmentStatuses: []db.ListGradeAssignmentStatusesRow{
			{
				AssignmentID:            assignmentID,
				ClassName:               "VII A",
				ClassCode:               "7A",
				SubjectName:             "Matematika",
				SubjectCode:             "MTK",
				TeacherName:             "Ibu Guru",
				ComponentCount:          2,
				PublishedComponentCount: 2,
				DraftComponentCount:     0,
				StudentCount:            10,
				ReadyStudentCount:       10,
				IncompleteStudentCount:  0,
				MissingGradeCount:       0,
				IsFinalized:             true,
				FinalizedBy:             pgtype.Text{String: "ibu.guru", Valid: true},
				Notes:                   pgtype.Text{String: "siap cetak", Valid: true},
				FinalizedAt:             pgtype.Timestamptz{Valid: true},
			},
		},
	}
	svc := &Grade{q: store}

	overview, err := svc.Overview(context.Background(), pgtype.UUID{}, pgtype.UUID{}, false, pgtype.UUID{})
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if len(overview.AssignmentStatuses) != 1 {
		t.Fatalf("assignment status count = %d, want 1", len(overview.AssignmentStatuses))
	}
	if !overview.AssignmentStatuses[0].Ready {
		t.Fatal("expected assignment status to be ready")
	}
	if !overview.AssignmentStatuses[0].IsFinalized {
		t.Fatal("expected assignment status to be finalized")
	}
	if overview.AssignmentStatuses[0].Finalization == nil || overview.AssignmentStatuses[0].Finalization.FinalizedBy != "ibu.guru" {
		t.Fatal("expected finalization metadata to be mapped")
	}
}

func TestGradeOverviewFiltersAssignmentsForTeacher(t *testing.T) {
	teacherID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	store := &fakeGradeStore{
		assignments: []db.ListClassSubjectAssignmentsRow{
			{ID: pgtype.UUID{Bytes: [16]byte{10}, Valid: true}, TeacherEmployeeID: teacherID},
			{ID: pgtype.UUID{Bytes: [16]byte{11}, Valid: true}, TeacherEmployeeID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true}},
		},
		assignmentStatuses: []db.ListGradeAssignmentStatusesRow{
			{AssignmentID: pgtype.UUID{Bytes: [16]byte{10}, Valid: true}, TeacherEmployeeID: teacherID},
			{AssignmentID: pgtype.UUID{Bytes: [16]byte{11}, Valid: true}, TeacherEmployeeID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true}},
		},
	}
	svc := &Grade{q: store}

	overview, err := svc.Overview(context.Background(), pgtype.UUID{}, pgtype.UUID{}, false, teacherID)
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if len(overview.Assignments) != 1 {
		t.Fatalf("assignment count = %d, want 1", len(overview.Assignments))
	}
	if len(overview.AssignmentStatuses) != 1 {
		t.Fatalf("assignment status count = %d, want 1", len(overview.AssignmentStatuses))
	}
}

func TestGradeCreateComponentRejectsTeacherOutsideAssignment(t *testing.T) {
	assignmentID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	teacherID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeGradeStore{
		assignments: []db.ListClassSubjectAssignmentsRow{
			{ID: assignmentID, TeacherEmployeeID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
		},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &Grade{q: store}

	_, err := svc.CreateComponent(context.Background(), db.CreateGradeComponentParams{
		AssignmentID: assignmentID,
		Title:        "UH 1",
		Category:     "quiz",
		Weight:       1,
		MaxScore:     100,
	}, teacherID)
	if err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("error = %v, want akses ditolak", err)
	}
}

func TestGradeCreatePublishDeleteAndUpsertEntryForwardNormalizedValues(t *testing.T) {
	assignmentID := pgtype.UUID{Bytes: [16]byte{21}, Valid: true}
	componentID := pgtype.UUID{Bytes: [16]byte{22}, Valid: true}
	studentID := pgtype.UUID{Bytes: [16]byte{23}, Valid: true}
	store := &fakeGradeStore{
		component:       db.GradeComponent{ID: componentID, AssignmentID: assignmentID, Title: "UH 1", MaxScore: 100},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &Grade{q: store}

	if _, err := svc.CreateComponent(context.Background(), db.CreateGradeComponentParams{
		AssignmentID: assignmentID,
		Title:        "  UH 1  ",
		Category:     "QUIZ",
		Weight:       2,
		MaxScore:     100,
	}, pgtype.UUID{}); err != nil {
		t.Fatalf("CreateComponent() error = %v", err)
	}
	if store.createArg.Title != "UH 1" || store.createArg.Category != "quiz" {
		t.Fatalf("CreateComponent() arg = %+v, want normalized title/category", store.createArg)
	}

	if _, err := svc.SetComponentPublished(context.Background(), componentID, true, pgtype.UUID{}); err != nil {
		t.Fatalf("SetComponentPublished() error = %v", err)
	}
	if store.publishArg.ID != componentID || !store.publishArg.IsPublished {
		t.Fatalf("SetComponentPublished() arg = %+v, want published component", store.publishArg)
	}

	teacherID := pgtype.UUID{Bytes: [16]byte{24}, Valid: true}
	store = &fakeGradeStore{
		component:       db.GradeComponent{ID: componentID, AssignmentID: assignmentID, Title: "UH 1", MaxScore: 100},
		finalizationErr: pgx.ErrNoRows,
		assignments: []db.ListClassSubjectAssignmentsRow{
			{ID: assignmentID, TeacherEmployeeID: teacherID},
		},
	}
	svc = &Grade{q: store}
	if _, err := svc.SetComponentPublished(context.Background(), componentID, false, teacherID); err != nil {
		t.Fatalf("SetComponentPublished(teacher scoped) error = %v", err)
	}
	if store.publishArg.ID != componentID || store.publishArg.IsPublished {
		t.Fatalf("SetComponentPublished(teacher scoped) arg = %+v, want unpublished component", store.publishArg)
	}

	if err := svc.DeleteComponent(context.Background(), componentID, pgtype.UUID{}); err != nil {
		t.Fatalf("DeleteComponent() error = %v", err)
	}
	if store.deleteComponentID != componentID {
		t.Fatalf("DeleteComponent() id = %v, want %v", store.deleteComponentID, componentID)
	}

	entry, err := svc.UpsertEntry(context.Background(), componentID, studentID, pgtype.UUID{}, 88.5, "  bagus  ", "  guru-a  ")
	if err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}
	if entry.ComponentID != componentID || store.upsertEntryArg.StudentID != studentID {
		t.Fatalf("UpsertEntry() row/arg = %+v/%+v, want forwarded IDs", entry, store.upsertEntryArg)
	}
	if !store.upsertEntryArg.Score.Valid || store.upsertEntryArg.Score.Float64 != 88.5 || store.upsertEntryArg.Notes != "bagus" || store.upsertEntryArg.GradedBy != "guru-a" {
		t.Fatalf("UpsertEntry() arg = %+v, want normalized score metadata", store.upsertEntryArg)
	}

	if _, err := svc.UpsertEntry(context.Background(), componentID, studentID, pgtype.UUID{}, -1, "", ""); err == nil || !strings.Contains(err.Error(), "tidak boleh negatif") {
		t.Fatalf("UpsertEntry(negative) error = %v, want negative guard", err)
	}
	if _, err := svc.UpsertEntry(context.Background(), componentID, studentID, pgtype.UUID{}, 101, "", ""); err == nil || !strings.Contains(err.Error(), "melebihi skor maksimum") {
		t.Fatalf("UpsertEntry(too high) error = %v, want max-score guard", err)
	}
}

func TestGradeOverviewLoadsSelectedAssignmentAndComponentDetails(t *testing.T) {
	assignmentID := pgtype.UUID{Bytes: [16]byte{31}, Valid: true}
	componentID := pgtype.UUID{Bytes: [16]byte{32}, Valid: true}
	store := &fakeGradeStore{
		component: db.GradeComponent{ID: componentID, AssignmentID: assignmentID, Title: "UH 1", MaxScore: 100},
		listComponents: []db.ListGradeComponentsRow{
			{ID: componentID, AssignmentID: assignmentID, Title: "UH 1", IsPublished: true},
		},
		listSummary: []db.ListGradebookSummaryRow{
			{ComponentCount: 1, FilledCount: 1},
		},
		listEntries: []db.ListGradeEntriesByComponentRow{
			{StudentID: pgtype.UUID{Bytes: [16]byte{33}, Valid: true}},
		},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &Grade{q: store}

	overview, err := svc.Overview(context.Background(), assignmentID, componentID, true, pgtype.UUID{})
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if len(overview.Components) != 1 || len(overview.Summary) != 1 || len(overview.Entries) != 1 {
		t.Fatalf("Overview() components/summary/entries = %d/%d/%d, want 1/1/1", len(overview.Components), len(overview.Summary), len(overview.Entries))
	}
	if !overview.Readiness.Ready {
		t.Fatal("Overview() readiness = false, want ready")
	}
}

func TestGradeOverviewAndAccessErrorBranches(t *testing.T) {
	assignmentID := pgtype.UUID{Bytes: [16]byte{41}, Valid: true}
	componentID := pgtype.UUID{Bytes: [16]byte{42}, Valid: true}
	teacherID := pgtype.UUID{Bytes: [16]byte{43}, Valid: true}
	expectedErr := errors.New("assignments failed")
	svc := &Grade{q: &fakeGradeStore{assignmentsErr: expectedErr}}
	if _, err := svc.Overview(context.Background(), pgtype.UUID{}, pgtype.UUID{}, false, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview(assignments error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("statuses failed")
	svc = &Grade{q: &fakeGradeStore{assignmentStatusesErr: expectedErr}}
	if _, err := svc.Overview(context.Background(), pgtype.UUID{}, pgtype.UUID{}, false, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview(statuses error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("components failed")
	svc = &Grade{q: &fakeGradeStore{listComponentsErr: expectedErr}}
	if _, err := svc.Overview(context.Background(), assignmentID, pgtype.UUID{}, false, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview(components error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("summary failed")
	svc = &Grade{q: &fakeGradeStore{
		listComponents: []db.ListGradeComponentsRow{{ID: componentID, AssignmentID: assignmentID, IsPublished: true}},
		listSummaryErr: expectedErr,
	}}
	if _, err := svc.Overview(context.Background(), assignmentID, pgtype.UUID{}, false, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview(summary error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("finalization failed")
	svc = &Grade{q: &fakeGradeStore{
		listComponents:  []db.ListGradeComponentsRow{{ID: componentID, AssignmentID: assignmentID, IsPublished: true}},
		listSummary:     []db.ListGradebookSummaryRow{{ComponentCount: 1, FilledCount: 1}},
		finalizationErr: expectedErr,
	}}
	if _, err := svc.Overview(context.Background(), assignmentID, pgtype.UUID{}, false, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview(finalization error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("component failed")
	svc = &Grade{q: &fakeGradeStore{componentErr: expectedErr}}
	if _, err := svc.Overview(context.Background(), pgtype.UUID{}, componentID, false, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview(component access error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("entries failed")
	svc = &Grade{q: &fakeGradeStore{
		component:      db.GradeComponent{ID: componentID, AssignmentID: assignmentID},
		listEntriesErr: expectedErr,
	}}
	if _, err := svc.Overview(context.Background(), pgtype.UUID{}, componentID, false, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview(entries error) = %v, want %v", err, expectedErr)
	}

	svc = &Grade{q: &fakeGradeStore{
		component: db.GradeComponent{ID: componentID, AssignmentID: assignmentID},
		assignments: []db.ListClassSubjectAssignmentsRow{
			{ID: assignmentID, TeacherEmployeeID: teacherID},
		},
		listEntries: []db.ListGradeEntriesByComponentRow{{StudentID: pgtype.UUID{Bytes: [16]byte{44}, Valid: true}}},
	}}
	overview, err := svc.Overview(context.Background(), pgtype.UUID{}, componentID, false, teacherID)
	if err != nil {
		t.Fatalf("Overview(component teacher access) error = %v", err)
	}
	if len(overview.Entries) != 1 || !overview.Entries[0].StudentID.Valid {
		t.Fatalf("Overview(component teacher access) entries = %+v, want scoped entry", overview.Entries)
	}
}

func TestGradeMutationValidationAndStoreErrorBranches(t *testing.T) {
	assignmentID := pgtype.UUID{Bytes: [16]byte{51}, Valid: true}
	componentID := pgtype.UUID{Bytes: [16]byte{52}, Valid: true}
	studentID := pgtype.UUID{Bytes: [16]byte{53}, Valid: true}
	teacherID := pgtype.UUID{Bytes: [16]byte{54}, Valid: true}
	validCreate := db.CreateGradeComponentParams{AssignmentID: assignmentID, Title: "UH 1", Category: "quiz", Weight: 1, MaxScore: 100}
	validUpdate := db.UpdateGradeComponentParams{ID: componentID, Title: "UH 1", Category: "quiz", Weight: 1, MaxScore: 100}

	svc := &Grade{q: &fakeGradeStore{
		assignments: []db.ListClassSubjectAssignmentsRow{{ID: pgtype.UUID{Bytes: [16]byte{55}, Valid: true}, TeacherEmployeeID: teacherID}},
	}}
	if err := svc.ReopenAssignment(context.Background(), assignmentID, teacherID); err == nil || !strings.Contains(err.Error(), "assignment tidak ditemukan") {
		t.Fatalf("ReopenAssignment(not found) = %v, want assignment tidak ditemukan", err)
	}

	expectedErr := errors.New("assignment access failed")
	svc = &Grade{q: &fakeGradeStore{assignmentsErr: expectedErr}}
	if _, err := svc.CreateComponent(context.Background(), validCreate, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("CreateComponent(access error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("editable check failed")
	svc = &Grade{q: &fakeGradeStore{finalizationErr: expectedErr}}
	if _, err := svc.CreateComponent(context.Background(), validCreate, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("CreateComponent(editable error) = %v, want %v", err, expectedErr)
	}

	svc = &Grade{q: &fakeGradeStore{finalizationErr: pgx.ErrNoRows}}
	if _, err := svc.CreateComponent(context.Background(), db.CreateGradeComponentParams{AssignmentID: assignmentID, Title: " ", Weight: 1, MaxScore: 100}, pgtype.UUID{}); err == nil || !strings.Contains(err.Error(), "judul komponen wajib diisi") {
		t.Fatalf("CreateComponent(blank title) = %v, want title validation", err)
	}
	if _, err := svc.CreateComponent(context.Background(), db.CreateGradeComponentParams{AssignmentID: assignmentID, Title: "UH", Weight: -1, MaxScore: 100}, pgtype.UUID{}); err == nil || !strings.Contains(err.Error(), "bobot tidak boleh negatif") {
		t.Fatalf("CreateComponent(negative weight) = %v, want weight validation", err)
	}
	if _, err := svc.CreateComponent(context.Background(), db.CreateGradeComponentParams{AssignmentID: assignmentID, Title: "UH", Weight: 1, MaxScore: 0}, pgtype.UUID{}); err == nil || !strings.Contains(err.Error(), "skor maksimum harus lebih dari 0") {
		t.Fatalf("CreateComponent(zero max) = %v, want max validation", err)
	}

	expectedErr = errors.New("component failed")
	svc = &Grade{q: &fakeGradeStore{componentErr: expectedErr}}
	if _, err := svc.UpdateComponent(context.Background(), validUpdate, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateComponent(component error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("highest score failed")
	svc = &Grade{q: &fakeGradeStore{
		component:       db.GradeComponent{ID: componentID, AssignmentID: assignmentID, MaxScore: 100},
		finalizationErr: pgx.ErrNoRows,
		highestScoreErr: expectedErr,
	}}
	if _, err := svc.UpdateComponent(context.Background(), validUpdate, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateComponent(highest score error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("publish failed")
	svc = &Grade{q: &fakeGradeStore{
		component:       db.GradeComponent{ID: componentID, AssignmentID: assignmentID},
		finalizationErr: pgx.ErrNoRows,
		publishErr:      expectedErr,
	}}
	if _, err := svc.SetComponentPublished(context.Background(), componentID, true, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("SetComponentPublished(publish error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("delete failed")
	svc = &Grade{q: &fakeGradeStore{
		component:          db.GradeComponent{ID: componentID, AssignmentID: assignmentID},
		finalizationErr:    pgx.ErrNoRows,
		deleteComponentErr: expectedErr,
	}}
	if err := svc.DeleteComponent(context.Background(), componentID, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("DeleteComponent(delete error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("entry failed")
	svc = &Grade{q: &fakeGradeStore{
		component:       db.GradeComponent{ID: componentID, AssignmentID: assignmentID, MaxScore: 100},
		finalizationErr: pgx.ErrNoRows,
		upsertEntryErr:  expectedErr,
	}}
	if _, err := svc.UpsertEntry(context.Background(), componentID, studentID, pgtype.UUID{}, 80, "", "guru"); !errors.Is(err, expectedErr) {
		t.Fatalf("UpsertEntry(store error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("reopen failed")
	svc = &Grade{q: &fakeGradeStore{finalizationDeleteErr: expectedErr}}
	if err := svc.ReopenAssignment(context.Background(), assignmentID, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("ReopenAssignment(delete error) = %v, want %v", err, expectedErr)
	}

	if got := normalizeGradeCategory("unknown"); got != "other" {
		t.Fatalf("normalizeGradeCategory(unknown) = %q, want other", got)
	}
}
