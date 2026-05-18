package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeQuestionStore struct {
	current db.GetCbtQuestionRow
	detail  db.GetCbtQuestionDetailRow

	listRows      []db.ListCbtQuestionsRow
	filteredRows  []db.ListCbtQuestionsFilteredRow
	stemRows      []db.ListCbtQuestionStemTextsBySubjectRow
	listFilterArg db.ListCbtQuestionsFilteredParams
	countArg      db.CountCbtQuestionsFilteredParams
	count         int64

	subjectMissing bool
	eventMissing   bool

	createParams  db.CreateCbtQuestionParams
	generatedCode string
	createCalls   int
	createRow     db.CbtQuestion
	createHistory []db.CreateCbtQuestionParams
	duplicateRow  db.CbtQuestion
	duplicateErr  error

	updateParams         db.UpdateCbtQuestionParams
	updateCalls          int
	nextVersionNumber    int32
	markNotLatestID      pgtype.UUID
	markNotLatestCalls   int
	markGroupNotLatestID pgtype.UUID
	markGroupLatestCalls int

	deleteID          pgtype.UUID
	deleteCalls       int
	assets            map[pgtype.UUID]db.CbtQuestionAsset
	assetErr          error
	membersByUser     []db.CbtEventMember
	membersByUsername []db.CbtEventMember
	auditLogs         []db.CbtQuestionAuditLog
	workflowEvents    []db.BankSoalQuestionWorkflowEvent
	workflowEventRows []db.ListBankSoalQuestionWorkflowEventsRow
	versionRows       []db.ListCbtQuestionVersionsRow
	auditErr          error
	auditCalls        int
	canReview         bool
	canApprove        bool
	summaryCounts     db.GetCbtQuestionSummaryCountsRow
	summarySubjects   []db.ListCbtQuestionSummaryBySubjectRow
	summaryCognitive  []db.ListCbtQuestionSummaryByCognitiveLevelRow
	summaryRecent     []db.ListCbtQuestionSummaryRecentRow
	summaryCountsArg  db.GetCbtQuestionSummaryCountsParams
	summarySubjectArg db.ListCbtQuestionSummaryBySubjectParams
	summaryCogArg     db.ListCbtQuestionSummaryByCognitiveLevelParams
	summaryRecentArg  db.ListCbtQuestionSummaryRecentParams
}

func (f *fakeQuestionStore) CbtQuestionSubjectExists(ctx context.Context, id pgtype.UUID) (bool, error) {
	return !f.subjectMissing, nil
}

func (f *fakeQuestionStore) CbtQuestionEventExists(ctx context.Context, id pgtype.UUID) (bool, error) {
	return !f.eventMissing, nil
}

func (f *fakeQuestionStore) GenerateCbtQuestionAcademicCode(ctx context.Context, arg db.GenerateCbtQuestionAcademicCodeParams) (string, error) {
	if strings.TrimSpace(f.generatedCode) != "" {
		return f.generatedCode, nil
	}
	return "MTK-VII-PG-0001", nil
}

func (f *fakeQuestionStore) ListCbtQuestions(ctx context.Context, arg db.ListCbtQuestionsParams) ([]db.ListCbtQuestionsRow, error) {
	return f.listRows, nil
}

func (f *fakeQuestionStore) ListCbtEventMembersByUser(ctx context.Context, userID pgtype.UUID) ([]db.CbtEventMember, error) {
	return f.membersByUser, nil
}

func (f *fakeQuestionStore) ListCbtEventMembersByUsername(ctx context.Context, username string) ([]db.CbtEventMember, error) {
	return f.membersByUsername, nil
}

func (f *fakeQuestionStore) ListCbtQuestionsFiltered(ctx context.Context, arg db.ListCbtQuestionsFilteredParams) ([]db.ListCbtQuestionsFilteredRow, error) {
	f.listFilterArg = arg
	return f.filteredRows, nil
}

func (f *fakeQuestionStore) CountCbtQuestionsFiltered(ctx context.Context, arg db.CountCbtQuestionsFilteredParams) (int64, error) {
	f.countArg = arg
	return f.count, nil
}

func (f *fakeQuestionStore) GetCbtQuestionSummaryCounts(ctx context.Context, arg db.GetCbtQuestionSummaryCountsParams) (db.GetCbtQuestionSummaryCountsRow, error) {
	f.summaryCountsArg = arg
	return f.summaryCounts, nil
}

func (f *fakeQuestionStore) ListCbtQuestionSummaryBySubject(ctx context.Context, arg db.ListCbtQuestionSummaryBySubjectParams) ([]db.ListCbtQuestionSummaryBySubjectRow, error) {
	f.summarySubjectArg = arg
	return f.summarySubjects, nil
}

func (f *fakeQuestionStore) ListCbtQuestionSummaryByCognitiveLevel(ctx context.Context, arg db.ListCbtQuestionSummaryByCognitiveLevelParams) ([]db.ListCbtQuestionSummaryByCognitiveLevelRow, error) {
	f.summaryCogArg = arg
	return f.summaryCognitive, nil
}

func (f *fakeQuestionStore) ListCbtQuestionSummaryRecent(ctx context.Context, arg db.ListCbtQuestionSummaryRecentParams) ([]db.ListCbtQuestionSummaryRecentRow, error) {
	f.summaryRecentArg = arg
	return f.summaryRecent, nil
}

func (f *fakeQuestionStore) ListCbtQuestionStemTextsBySubject(ctx context.Context, subjectID pgtype.UUID) ([]db.ListCbtQuestionStemTextsBySubjectRow, error) {
	return f.stemRows, nil
}

func (f *fakeQuestionStore) GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	if !f.current.ID.Valid {
		return db.GetCbtQuestionRow{}, errors.New("not found")
	}
	return f.current, nil
}

func (f *fakeQuestionStore) GetCbtQuestionDetail(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionDetailRow, error) {
	return f.detail, nil
}

func (f *fakeQuestionStore) GetCbtQuestionAsset(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error) {
	if f.assetErr != nil {
		return db.CbtQuestionAsset{}, f.assetErr
	}
	if f.assets != nil {
		if asset, ok := f.assets[id]; ok {
			return asset, nil
		}
	}
	return db.CbtQuestionAsset{}, pgx.ErrNoRows
}

func (f *fakeQuestionStore) AcquireCbtQuestionDraftDuplicateLock(ctx context.Context, fingerprint string) error {
	return nil
}

func (f *fakeQuestionStore) FindRecentCbtQuestionDraftDuplicate(ctx context.Context, arg db.FindRecentCbtQuestionDraftDuplicateParams) (db.CbtQuestion, error) {
	if f.duplicateErr != nil {
		return db.CbtQuestion{}, f.duplicateErr
	}
	if f.duplicateRow.ID.Valid {
		return f.duplicateRow, nil
	}
	return db.CbtQuestion{}, pgx.ErrNoRows
}

func (f *fakeQuestionStore) CreateCbtQuestion(ctx context.Context, arg db.CreateCbtQuestionParams) (db.CbtQuestion, error) {
	f.createParams = arg
	f.createCalls++
	f.createHistory = append(f.createHistory, arg)
	return f.createRow, nil
}

func (f *fakeQuestionStore) UpdateCbtQuestion(ctx context.Context, arg db.UpdateCbtQuestionParams) (db.CbtQuestion, error) {
	f.updateParams = arg
	f.updateCalls++
	return db.CbtQuestion{
		ID:               arg.ID,
		QuestionText:     arg.QuestionText,
		WorkflowStatus:   arg.WorkflowStatus,
		Status:           arg.Status,
		StemHtml:         arg.StemHtml,
		StimulusHtml:     arg.StimulusHtml,
		ReviewerUsername: arg.ReviewerUsername,
		ApproverUsername: arg.ApproverUsername,
	}, nil
}

func (f *fakeQuestionStore) GetNextCbtQuestionVersionNumber(ctx context.Context, versionGroupID pgtype.UUID) (int32, error) {
	if f.nextVersionNumber > 0 {
		return f.nextVersionNumber, nil
	}
	if f.current.VersionNumber > 0 {
		return f.current.VersionNumber + 1, nil
	}
	return 2, nil
}

func (f *fakeQuestionStore) MarkCbtQuestionVersionNotLatest(ctx context.Context, id pgtype.UUID) error {
	f.markNotLatestID = id
	f.markNotLatestCalls++
	return nil
}

func (f *fakeQuestionStore) MarkCbtQuestionVersionGroupNotLatest(ctx context.Context, versionGroupID pgtype.UUID) error {
	f.markGroupNotLatestID = versionGroupID
	f.markGroupLatestCalls++
	return nil
}

func (f *fakeQuestionStore) DeleteCbtQuestion(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	f.deleteCalls++
	return nil
}

func (f *fakeQuestionStore) CreateCbtQuestionAuditLog(ctx context.Context, arg db.CreateCbtQuestionAuditLogParams) (db.CbtQuestionAuditLog, error) {
	f.auditCalls++
	if f.auditErr != nil {
		return db.CbtQuestionAuditLog{}, f.auditErr
	}
	row := db.CbtQuestionAuditLog{QuestionID: arg.QuestionID, ActorUsername: arg.ActorUsername, Action: arg.Action, Note: arg.Note, Metadata: arg.Metadata}
	f.auditLogs = append(f.auditLogs, row)
	return row, nil
}

func (f *fakeQuestionStore) CreateBankSoalQuestionWorkflowEvent(ctx context.Context, arg db.CreateBankSoalQuestionWorkflowEventParams) (db.BankSoalQuestionWorkflowEvent, error) {
	row := db.BankSoalQuestionWorkflowEvent{
		QuestionID:    arg.QuestionID,
		ActorUserID:   arg.ActorUserID,
		ActorUsername: arg.ActorUsername,
		FromStatus:    arg.FromStatus,
		ToStatus:      arg.ToStatus,
		Action:        arg.Action,
		Note:          arg.Note,
		Metadata:      arg.Metadata,
	}
	f.workflowEvents = append(f.workflowEvents, row)
	return row, nil
}

func (f *fakeQuestionStore) CanBankSoalUserReview(ctx context.Context, arg db.CanBankSoalUserReviewParams) (bool, error) {
	return f.canReview, nil
}

func (f *fakeQuestionStore) CanBankSoalUserApprove(ctx context.Context, arg db.CanBankSoalUserApproveParams) (bool, error) {
	return f.canApprove, nil
}

func (f *fakeQuestionStore) ListCbtQuestionTimeline(ctx context.Context, questionID pgtype.UUID) ([]db.ListCbtQuestionTimelineRow, error) {
	rows := make([]db.ListCbtQuestionTimelineRow, 0, len(f.auditLogs))
	for _, log := range f.auditLogs {
		rows = append(rows, db.ListCbtQuestionTimelineRow{
			ID:               log.ID,
			QuestionID:       log.QuestionID,
			ActorUsername:    log.ActorUsername,
			ActorDisplayName: log.ActorUsername,
			Action:           log.Action,
			Note:             log.Note,
			Metadata:         log.Metadata,
			CreatedAt:        log.CreatedAt,
		})
	}
	return rows, nil
}

func (f *fakeQuestionStore) ListBankSoalQuestionWorkflowEvents(ctx context.Context, questionID pgtype.UUID) ([]db.ListBankSoalQuestionWorkflowEventsRow, error) {
	if f.workflowEventRows != nil {
		return f.workflowEventRows, nil
	}
	rows := make([]db.ListBankSoalQuestionWorkflowEventsRow, 0, len(f.workflowEvents))
	for _, event := range f.workflowEvents {
		rows = append(rows, db.ListBankSoalQuestionWorkflowEventsRow{
			ID:               event.ID,
			QuestionID:       event.QuestionID,
			ActorUserID:      event.ActorUserID,
			ActorUsername:    event.ActorUsername,
			ActorDisplayName: event.ActorUsername,
			FromStatus:       event.FromStatus,
			ToStatus:         event.ToStatus,
			Action:           event.Action,
			Note:             event.Note,
			Metadata:         event.Metadata,
			CreatedAt:        event.CreatedAt,
		})
	}
	return rows, nil
}

func (f *fakeQuestionStore) ListCbtQuestionVersions(ctx context.Context, id pgtype.UUID) ([]db.ListCbtQuestionVersionsRow, error) {
	return f.versionRows, nil
}

func mustQuestionUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("scan uuid %q: %v", value, err)
	}
	return id
}

func TestNewCbtQuestionAndReadDelegation(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	store := &fakeQuestionStore{
		listRows: []db.ListCbtQuestionsRow{{ID: questionID, Code: "Q-1"}},
		current:  db.GetCbtQuestionRow{ID: questionID, Code: "Q-1", QuestionText: "Soal"},
		detail:   db.GetCbtQuestionDetailRow{ID: questionID, Code: "Q-1", QuestionText: "Soal detail"},
	}
	svc := NewCbtQuestion(nil)
	svc.q = store

	rows, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Code != "Q-1" {
		t.Fatalf("List() = %+v, want one Q-1 row", rows)
	}

	got, err := svc.Get(context.Background(), questionID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != questionID || got.QuestionText != "Soal" {
		t.Fatalf("Get() = %+v, want current row", got)
	}

	detail, err := svc.GetDetail(context.Background(), questionID, CbtQuestionActor{Username: "admin", Roles: []string{"admin"}})
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if detail.ID != questionID || detail.QuestionText != "Soal detail" {
		t.Fatalf("GetDetail() = %+v, want detail row", detail)
	}

	store.versionRows = []db.ListCbtQuestionVersionsRow{{ID: questionID, Code: "Q-1", VersionNumber: 1, IsLatestVersion: true}}
	versions, err := svc.Versions(context.Background(), questionID, CbtQuestionActor{Username: "admin", Roles: []string{"admin"}})
	if err != nil {
		t.Fatalf("Versions() error = %v", err)
	}
	if len(versions) != 1 || versions[0].VersionNumber != 1 || !versions[0].IsLatestVersion {
		t.Fatalf("Versions() = %+v, want one latest v1 row", versions)
	}

	store.workflowEventRows = []db.ListBankSoalQuestionWorkflowEventsRow{{QuestionID: questionID, Action: "submit_for_review", FromStatus: "draft", ToStatus: "submitted"}}
	events, err := svc.WorkflowEvents(context.Background(), questionID, CbtQuestionActor{Username: "admin", Roles: []string{"admin"}})
	if err != nil {
		t.Fatalf("WorkflowEvents() error = %v", err)
	}
	if len(events) != 1 || events[0].Action != "submit_for_review" || events[0].ToStatus != "submitted" {
		t.Fatalf("WorkflowEvents() = %+v, want submitted workflow event", events)
	}
}

func TestCbtQuestionFilterCreateAndDeleteDelegation(t *testing.T) {
	assetID := mustQuestionUUID(t, "11111111-1111-4111-8111-111111111111")
	store := &fakeQuestionStore{
		filteredRows: []db.ListCbtQuestionsFilteredRow{{ID: pgtype.UUID{Valid: true}, Code: "Q-1"}},
		count:        7,
		createRow: db.CbtQuestion{
			ID:           pgtype.UUID{Valid: true},
			QuestionText: "Soal mudah",
		},
		assets: map[pgtype.UUID]db.CbtQuestionAsset{
			assetID: {ID: assetID, UploadedBy: "guru"},
		},
	}
	svc := &CbtQuestion{q: store}

	rows, total, err := svc.ListFiltered(context.Background(), ListCbtQuestionsInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthorUsername: " guru.ipa ",
		QuestionScope:  " event_pool ",
		WorkflowStatus: " draft ",
		Status:         " published ",
		QuestionType:   " multiple_choice ",
		TargetLevel:    " viii ",
		Difficulty:     " hard ",
		CognitiveLevel: " C3 ",
		MaterialTopic:  " bilangan ",
		MetadataFilter: " kurang ",
		HotsFilter:     " true ",
		RevisionSource: " item_analysis ",
		SearchQuery:    " aljabar ",
		SortOrder:      " code_asc ",
		Limit:          25,
		Offset:         5,
	})
	if err != nil {
		t.Fatalf("ListFiltered() error = %v", err)
	}
	if len(rows) != 1 || total != 7 {
		t.Fatalf("ListFiltered() rows/total = %d/%d, want 1/7", len(rows), total)
	}
	if store.listFilterArg.AuthorUsername != "guru.ipa" || store.listFilterArg.ScopeFilter != "event_pool" || store.listFilterArg.WorkflowStatus != "draft" || store.listFilterArg.StatusFilter != "published" || store.listFilterArg.QuestionType != "multiple_choice" || store.listFilterArg.TargetLevel != "VIII" || store.listFilterArg.DifficultyFilter != "hard" || store.listFilterArg.CognitiveLevel != "C3" || store.listFilterArg.MaterialTopic != "bilangan" || store.listFilterArg.MetadataFilter != "gap" || store.listFilterArg.HotsFilter != "true" || store.listFilterArg.RevisionSource != "item_analysis" || store.listFilterArg.SearchQuery != "aljabar" || store.listFilterArg.SortOrder != "code_asc" {
		t.Fatalf("ListFiltered() arg = %+v, want trimmed filters", store.listFilterArg)
	}
	if store.countArg.AuthorUsername != store.listFilterArg.AuthorUsername || store.countArg.ScopeFilter != store.listFilterArg.ScopeFilter || store.countArg.WorkflowStatus != store.listFilterArg.WorkflowStatus || store.countArg.StatusFilter != store.listFilterArg.StatusFilter || store.countArg.TargetLevel != store.listFilterArg.TargetLevel || store.countArg.DifficultyFilter != store.listFilterArg.DifficultyFilter || store.countArg.CognitiveLevel != store.listFilterArg.CognitiveLevel || store.countArg.MaterialTopic != store.listFilterArg.MaterialTopic || store.countArg.MetadataFilter != store.listFilterArg.MetadataFilter || store.countArg.RevisionSource != store.listFilterArg.RevisionSource || store.countArg.SearchQuery != store.listFilterArg.SearchQuery {
		t.Fatalf("ListFiltered() count arg = %+v, want same trimmed filters", store.countArg)
	}

	created, err := svc.Create(context.Background(), SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "single_choice",
		QuestionText:   " Soal mudah ",
		TargetLevel:    " viii ",
		OptionA:        " A ",
		OptionB:        " B ",
		OptionC:        " C ",
		OptionD:        " D ",
		AnswerKey:      " b ",
		MediaAssetIDs:  []string{"11111111-1111-4111-8111-111111111111"},
		AuthorUsername: " guru ",
		Actor:          CbtQuestionActor{Username: "guru", Roles: []string{"guru"}},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !created.ID.Valid || store.createCalls != 1 {
		t.Fatalf("Create() row/calls = %+v/%d, want store row and one call", created, store.createCalls)
	}
	if store.createParams.QuestionType != "multiple_choice" || store.createParams.QuestionText != "Soal mudah" || store.createParams.AnswerKey != "B" {
		t.Fatalf("Create() params = %+v, want normalized type/text/answer", store.createParams)
	}
	if !store.createParams.TargetLevel.Valid || store.createParams.TargetLevel.String != "VIII" {
		t.Fatalf("Create() target_level = %+v, want VIII", store.createParams.TargetLevel)
	}
	if store.createParams.OptionA != "A" || store.createParams.OptionD != "D" || store.createParams.Version != 1 || store.createParams.AuthorUsername != "guru" {
		t.Fatalf("Create() params = %+v, want legacy options/version/author", store.createParams)
	}
	if string(store.createParams.MediaAssetIds) != `["11111111-1111-4111-8111-111111111111"]` {
		t.Fatalf("Create() media_asset_ids = %s, want asset JSON", string(store.createParams.MediaAssetIds))
	}

	deleteID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	store.current = db.GetCbtQuestionRow{ID: deleteID, Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft"}
	if err := svc.Delete(context.Background(), deleteID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if store.deleteID != deleteID || store.deleteCalls != 1 {
		t.Fatalf("Delete() id/calls = %v/%d, want %v/1", store.deleteID, store.deleteCalls, deleteID)
	}
}

func TestCbtQuestionCreateAllowsMissingTargetLevel(t *testing.T) {
	store := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}

	_, err := svc.Create(context.Background(), SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal tanpa tingkat",
		OptionA:        "A",
		OptionB:        "B",
		OptionC:        "C",
		OptionD:        "D",
		AnswerKey:      "A",
		AuthorUsername: "guru",
		Actor:          CbtQuestionActor{Username: "guru", Roles: []string{"guru"}},
	})
	if err != nil {
		t.Fatalf("Create(missing target_level) error = %v", err)
	}
	if store.createCalls != 1 {
		t.Fatalf("Create(missing target_level) calls = %d, want 1", store.createCalls)
	}
	if store.createParams.TargetLevel.Valid {
		t.Fatalf("Create(missing target_level) target_level = %+v, want null", store.createParams.TargetLevel)
	}
}

func TestCbtQuestionCreateReturnsRecentDuplicateDraftWithoutInsert(t *testing.T) {
	duplicateID := pgtype.UUID{Bytes: [16]byte{8}, Valid: true}
	store := &fakeQuestionStore{
		duplicateRow: db.CbtQuestion{ID: duplicateID, Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft"},
		createRow:    db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{9}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}

	row, err := svc.Create(context.Background(), SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Bytes: [16]byte{7}, Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal sama",
		OptionA:        "A",
		OptionB:        "B",
		OptionC:        "C",
		OptionD:        "D",
		AnswerKey:      "A",
		AuthorUsername: "guru",
		Actor:          CbtQuestionActor{Username: "guru", Roles: []string{"guru"}},
	})
	if err != nil {
		t.Fatalf("Create(duplicate draft) error = %v", err)
	}
	if row.ID != duplicateID {
		t.Fatalf("Create(duplicate draft) id = %v, want existing duplicate %v", row.ID, duplicateID)
	}
	if store.createCalls != 0 {
		t.Fatalf("Create(duplicate draft) create calls = %d, want 0", store.createCalls)
	}
}

func TestCbtQuestionCreateDoesNotCollapseDifferentMetadataDraft(t *testing.T) {
	store := &fakeQuestionStore{
		duplicateRow: db.CbtQuestion{},
		createRow:    db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{9}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}

	_, err := svc.Create(context.Background(), SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Bytes: [16]byte{7}, Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal sama tapi metadata berbeda",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
		AnswerKey:      "A",
		MaterialTopic:  "Topik baru",
		AuthorUsername: "guru",
		Actor:          CbtQuestionActor{Username: "guru", Roles: []string{"guru"}},
	})
	if err != nil {
		t.Fatalf("Create(different metadata) error = %v", err)
	}
	if store.createCalls != 1 {
		t.Fatalf("Create(different metadata) create calls = %d, want 1", store.createCalls)
	}
}

func TestCbtQuestionCreateRejectsInvalidTargetLevel(t *testing.T) {
	store := &fakeQuestionStore{}
	svc := &CbtQuestion{q: store}

	_, err := svc.Create(context.Background(), SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal tingkat tidak valid",
		TargetLevel:    "X",
		OptionA:        "A",
		OptionB:        "B",
		OptionC:        "C",
		OptionD:        "D",
		AnswerKey:      "A",
		AuthorUsername: "guru",
		Actor:          CbtQuestionActor{Username: "guru", Roles: []string{"guru"}},
	})
	if err == nil || !strings.Contains(err.Error(), "target_level") {
		t.Fatalf("Create(invalid target_level) error = %v, want target_level validation", err)
	}
	if store.createCalls != 0 {
		t.Fatalf("Create(invalid target_level) calls = %d, want 0", store.createCalls)
	}
}

func TestCbtQuestionCreateRejectsDirectApprovalBypass(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}
	base := SaveCbtQuestionInput{
		SubjectID:      subjectID,
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal lengkap",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "review",
		Actor:          CbtQuestionActor{Username: "admin", Roles: []string{"admin"}},
	}
	if _, err := svc.Create(context.Background(), base); err != nil {
		t.Fatalf("Create(review draft) error = %v", err)
	}

	approved := base
	approved.WorkflowStatus = "approved"
	_, err := svc.Create(context.Background(), approved)
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Create(approved bypass) error = %v, want ErrBadRequest", err)
	}

	published := base
	published.Status = db.CbtQuestionStatusEnumPublished
	published.WorkflowStatus = "approved"
	_, err = svc.Create(context.Background(), published)
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Create(published bypass) error = %v, want ErrBadRequest", err)
	}
	if store.createCalls != 1 {
		t.Fatalf("Create direct bypass create calls = %d, want only initial valid create", store.createCalls)
	}
}

func TestCbtQuestionMediaAssetIDsAreValidated(t *testing.T) {
	assetID := mustQuestionUUID(t, "22222222-2222-4222-8222-222222222222")
	questionID := mustQuestionUUID(t, "33333333-3333-4333-8333-333333333333")
	otherQuestionID := mustQuestionUUID(t, "44444444-4444-4444-8444-444444444444")
	base := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Bytes: [16]byte{7}, Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal lengkap",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
		AnswerKey:      "A",
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		Actor:          CbtQuestionActor{Username: "admin", Roles: []string{"admin"}},
	}

	t.Run("rejects invalid uuid before create", func(t *testing.T) {
		store := &fakeQuestionStore{createRow: db.CbtQuestion{ID: questionID}}
		svc := &CbtQuestion{q: store}
		input := base
		input.MediaAssetIDs = []string{"asset-1"}

		_, err := svc.Create(context.Background(), input)
		if !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("Create(invalid asset id) error = %v, want ErrBadRequest", err)
		}
		if store.createCalls != 0 {
			t.Fatalf("Create(invalid asset id) calls = %d, want 0", store.createCalls)
		}
	})

	t.Run("rejects missing asset before create", func(t *testing.T) {
		store := &fakeQuestionStore{createRow: db.CbtQuestion{ID: questionID}}
		svc := &CbtQuestion{q: store}
		input := base
		input.MediaAssetIDs = []string{"22222222-2222-4222-8222-222222222222"}

		_, err := svc.Create(context.Background(), input)
		if !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("Create(missing asset) error = %v, want ErrBadRequest", err)
		}
		if store.createCalls != 0 {
			t.Fatalf("Create(missing asset) calls = %d, want 0", store.createCalls)
		}
	})

	t.Run("rejects asset already bound to another question", func(t *testing.T) {
		store := &fakeQuestionStore{
			current: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru", Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft"},
			assets: map[pgtype.UUID]db.CbtQuestionAsset{
				assetID: {ID: assetID, QuestionID: otherQuestionID, UploadedBy: "guru"},
			},
		}
		svc := &CbtQuestion{q: store}
		input := base
		input.ID = questionID
		input.AuthorUsername = "guru"
		input.Actor = CbtQuestionActor{Username: "guru", Roles: []string{"guru"}}
		input.MediaAssetIDs = []string{"22222222-2222-4222-8222-222222222222"}

		_, err := svc.Update(context.Background(), input)
		if !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("Update(cross question asset) error = %v, want ErrBadRequest", err)
		}
		if store.updateCalls != 0 {
			t.Fatalf("Update(cross question asset) calls = %d, want 0", store.updateCalls)
		}
	})

	t.Run("allows asset bound to same question and actor", func(t *testing.T) {
		store := &fakeQuestionStore{
			current: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru", Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft"},
			assets: map[pgtype.UUID]db.CbtQuestionAsset{
				assetID: {ID: assetID, QuestionID: questionID, UploadedBy: "guru"},
			},
		}
		svc := &CbtQuestion{q: store}
		input := base
		input.ID = questionID
		input.AuthorUsername = "guru"
		input.Actor = CbtQuestionActor{Username: "guru", Roles: []string{"guru"}}
		input.MediaAssetIDs = []string{" 22222222-2222-4222-8222-222222222222 "}

		_, err := svc.Update(context.Background(), input)
		if err != nil {
			t.Fatalf("Update(same question asset) error = %v", err)
		}
		if store.updateCalls != 1 || string(store.updateParams.MediaAssetIds) != `["22222222-2222-4222-8222-222222222222"]` {
			t.Fatalf("Update(same question asset) calls/media = %d/%s, want one normalized asset", store.updateCalls, string(store.updateParams.MediaAssetIds))
		}
	})
}

func TestCbtQuestionDeleteWithActorAllowsDraftUnusedQuestion(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	store := &fakeQuestionStore{
		current: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.a", Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft"},
	}
	svc := &CbtQuestion{q: store}

	if err := svc.DeleteWithActor(context.Background(), questionID, CbtQuestionActor{Username: "guru.a", Roles: []string{"guru"}}); err != nil {
		t.Fatalf("DeleteWithActor(draft unused) error = %v", err)
	}
	if store.deleteCalls != 1 || store.deleteID != questionID {
		t.Fatalf("DeleteWithActor(draft unused) delete = %d/%v, want once for question", store.deleteCalls, store.deleteID)
	}
	if store.auditCalls != 1 || len(store.auditLogs) != 1 || store.auditLogs[0].Action != "delete" {
		t.Fatalf("DeleteWithActor(draft unused) audit = calls %d logs %+v, want delete audit", store.auditCalls, store.auditLogs)
	}
}

func TestCbtQuestionDeleteWithActorRequiresModifyAccess(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	store := &fakeQuestionStore{
		current: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.a", Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft"},
	}
	svc := &CbtQuestion{q: store}

	err := svc.DeleteWithActor(context.Background(), questionID, CbtQuestionActor{Username: "guru.b", Roles: []string{"guru"}})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("DeleteWithActor(other author) error = %v, want forbidden", err)
	}
	if store.deleteCalls != 0 {
		t.Fatalf("DeleteWithActor(other author) delete calls = %d, want 0", store.deleteCalls)
	}
}

func TestCbtQuestionDeleteWithActorRejectsNonDraftOrUsedQuestion(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	tests := []struct {
		name    string
		current db.GetCbtQuestionRow
	}{
		{
			name:    "review workflow",
			current: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.a", Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "review"},
		},
		{
			name:    "published status",
			current: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.a", Status: db.CbtQuestionStatusEnumPublished, WorkflowStatus: "draft"},
		},
		{
			name:    "package usage",
			current: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.a", Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft", PackageCount: 1},
		},
		{
			name:    "student answer usage",
			current: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.a", Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft", AnswerCount: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeQuestionStore{current: tt.current}
			svc := &CbtQuestion{q: store}
			err := svc.DeleteWithActor(context.Background(), questionID, CbtQuestionActor{Username: "guru.a", Roles: []string{"guru"}})
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("DeleteWithActor(%s) error = %v, want conflict", tt.name, err)
			}
			if store.deleteCalls != 0 {
				t.Fatalf("DeleteWithActor(%s) delete calls = %d, want 0", tt.name, store.deleteCalls)
			}
		})
	}
}

func TestCbtQuestionImportLegacyCSVMapsRows(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}
	csvText := strings.Join([]string{
		"kode;soal;opsi_a;opsi_b;opsi_c;opsi_d;jawaban;gambar_soal;gambar_b;bobot;is_rtl",
		"MTK-1;<p>Berapa 2+2?</p>;3;4;5;6;1;https://cdn.test/soal.png;https://cdn.test/b.png;2;true",
		"MTK-1;Berapa 3+3?;5;6;7;8;B;;;;",
		"; ;A;B;C;D;A;;;;",
		"MTK-2;Pilih huruf;A;B;C;D;C;;;;",
	}, "\n")

	got, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   csvText,
		Username:  " guru.cbt ",
		Actor:     CbtQuestionActor{Username: "guru.cbt", Roles: []string{"admin"}},
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV() error = %v", err)
	}
	if got.TotalRows != 4 || got.Imported != 2 || got.Skipped != 2 {
		t.Fatalf("ImportLegacyCSV() result = %+v, want 4 total, 2 imported, 2 skipped", got)
	}
	if len(got.Errors) != 2 || !strings.Contains(strings.Join(got.Errors, "\n"), "kode MTK-1 duplikat") || !strings.Contains(strings.Join(got.Errors, "\n"), "soal kosong") {
		t.Fatalf("ImportLegacyCSV() errors = %+v, want duplicate code and empty question errors", got.Errors)
	}
	if len(got.DuplicateCodes) != 1 || got.DuplicateCodes[0] != "MTK-1" {
		t.Fatalf("ImportLegacyCSV() duplicate codes = %+v, want MTK-1", got.DuplicateCodes)
	}
	if store.createCalls != 2 || len(store.createHistory) != 2 {
		t.Fatalf("CreateCbtQuestion() calls/history = %d/%d, want 2/2", store.createCalls, len(store.createHistory))
	}
	first := store.createHistory[0]
	if first.SubjectID != subjectID || first.Code != "MTK-1" || first.AnswerKey != "B" || first.AuthorUsername != "guru.cbt" {
		t.Fatalf("first imported params = %+v, want subject/code/answer/author mapped", first)
	}
	if !strings.Contains(first.StemHtml, "https://cdn.test/soal.png") || !strings.Contains(string(first.Options), "https://cdn.test/b.png") {
		t.Fatalf("first imported media = stem %q options %s, want legacy image URLs", first.StemHtml, string(first.Options))
	}
	if !strings.Contains(first.WriterNotes, "Bobot legacy: 2") || !strings.Contains(first.WriterNotes, "RTL legacy: ya") {
		t.Fatalf("first imported notes = %q, want legacy metadata notes", first.WriterNotes)
	}
	if store.createHistory[1].Code != "MTK-2" || store.createHistory[1].AnswerKey != "C" {
		t.Fatalf("second imported params = %+v, want MTK-2 with answer C", store.createHistory[1])
	}
}

func TestCbtQuestionImportLegacyCSVSkipsExistingStem(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeQuestionStore{
		stemRows: []db.ListCbtQuestionStemTextsBySubjectRow{
			{QuestionText: "Berapa 2+2?", StemHtml: "<p>Berapa 2+2?</p>"},
		},
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}
	csvText := strings.Join([]string{
		"kode;soal;opsi_a;opsi_b;opsi_c;opsi_d;jawaban",
		"MTK-1;<p>Berapa 2+2?</p>;3;4;5;6;B",
		"MTK-2;Berapa 3+3?;5;6;7;8;B",
	}, "\n")

	got, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   csvText,
		Username:  "guru.cbt",
		Actor:     CbtQuestionActor{Username: "guru.cbt", Roles: []string{"admin"}},
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV() error = %v", err)
	}
	if got.Imported != 1 || got.Skipped != 1 || store.createCalls != 1 {
		t.Fatalf("ImportLegacyCSV() result/calls = %+v/%d, want 1 imported, 1 skipped, 1 create", got, store.createCalls)
	}
	if !strings.Contains(strings.Join(got.Errors, "\n"), "duplikat dengan bank soal") {
		t.Fatalf("ImportLegacyCSV() errors = %+v, want existing duplicate message", got.Errors)
	}
	if store.createHistory[0].Code != "MTK-2" {
		t.Fatalf("created code = %q, want only non-duplicate row", store.createHistory[0].Code)
	}
}

func TestCbtQuestionImportLegacyCSVDryRunDoesNotInsert(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}
	csvText := strings.Join([]string{
		"kode;soal;opsi_a;opsi_b;opsi_c;opsi_d;jawaban",
		"MTK-1;Berapa 2+2?;3;4;5;6;B",
		"MTK-2;Berapa 3+3?;5;6;7;8;B",
	}, "\n")

	got, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   csvText,
		Actor:     CbtQuestionActor{Username: "admin", Roles: []string{"admin"}},
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV(dry run) error = %v", err)
	}
	if !got.DryRun || got.WouldImport != 2 || got.Imported != 0 || got.Skipped != 0 {
		t.Fatalf("ImportLegacyCSV(dry run) result = %+v, want dry_run with 2 would_import and no inserts", got)
	}
	if store.createCalls != 0 || store.auditCalls != 0 {
		t.Fatalf("ImportLegacyCSV(dry run) create/audit calls = %d/%d, want 0/0", store.createCalls, store.auditCalls)
	}
}

func TestCbtQuestionImportLegacyCSVSupportsStructuredTypes(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	store := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}},
	}
	svc := &CbtQuestion{q: store}
	csvText := strings.Join([]string{
		importCSVRow("kode", "tipe", "soal", "opsi_a", "opsi_b", "opsi_c", "opsi_d", "opsi_e", "jawaban", "kiri_a", "kanan_1", "kiri_b", "kanan_2", "distraktor_1", "rubrik"),
		importCSVRow("MTK-G", "pg_kompleks", "Pilih bilangan genap", "1", "2", "3", "4", "6", "B E", "", "", "", "", "", ""),
		importCSVRow("MTK-BS", "benar_salah", "Air membeku pada suhu rendah", "", "", "", "", "", "Benar", "", "", "", "", "", ""),
		importCSVRow("MTK-IS", "isian", "Ibu kota Sulawesi Tenggara", "", "", "", "", "", "Kendari | kendari kota", "", "", "", "", "", ""),
		importCSVRow("MTK-ES", "essay", "Jelaskan fotosintesis", "", "", "", "", "", "", "", "", "", "", "", "<p>Ketepatan konsep</p>"),
		importCSVRow("MTK-M", "menjodohkan", "Pasangkan angka", "", "", "", "", "", "", "Satu", "1", "Dua", "2", "Tiga", ""),
	}, "\n")

	got, err := svc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   csvText,
		Username:  "guru.cbt",
		Actor:     CbtQuestionActor{Username: "guru.cbt", Roles: []string{"admin"}},
	})
	if err != nil {
		t.Fatalf("ImportLegacyCSV() error = %v", err)
	}
	if got.TotalRows != 5 || got.Imported != 5 || got.Skipped != 0 {
		t.Fatalf("ImportLegacyCSV() result = %+v, want 5 imported and no skipped rows", got)
	}
	if store.createCalls != 5 || len(store.createHistory) != 5 {
		t.Fatalf("CreateCbtQuestion() calls/history = %d/%d, want 5/5", store.createCalls, len(store.createHistory))
	}
	if store.createHistory[0].QuestionType != "multiple_answer" || store.createHistory[0].AnswerKey != "B,E" || !strings.Contains(string(store.createHistory[0].Options), `"label":"E"`) {
		t.Fatalf("multiple answer import = type %q answer %q options %s, want type/options/key mapped", store.createHistory[0].QuestionType, store.createHistory[0].AnswerKey, string(store.createHistory[0].Options))
	}
	if store.createHistory[1].QuestionType != "true_false" || store.createHistory[1].AnswerKey != "A" || !strings.Contains(string(store.createHistory[1].Options), "Benar") {
		t.Fatalf("true/false import = %+v options %s, want fixed options and answer A", store.createHistory[1], string(store.createHistory[1].Options))
	}
	if store.createHistory[2].QuestionType != "short_answer" || store.createHistory[2].AnswerKey != "Kendari|kendari kota" {
		t.Fatalf("short answer import = %+v, want aliases preserved", store.createHistory[2])
	}
	if store.createHistory[3].QuestionType != "essay" || store.createHistory[3].RubricHtml != "<p>Ketepatan konsep</p>" {
		t.Fatalf("essay import = %+v, want rubric mapped", store.createHistory[3])
	}
	if store.createHistory[4].QuestionType != "matching" || store.createHistory[4].AnswerKey != "A=1;B=2" || !strings.Contains(string(store.createHistory[4].Options), `"is_distractor":true`) {
		t.Fatalf("matching import = %+v options %s, want default answer and distractor", store.createHistory[4], string(store.createHistory[4].Options))
	}
}

func TestCbtQuestionCreateAllowsGuruReusableAndRejectsNonGuru(t *testing.T) {
	store := &fakeQuestionStore{createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{8}, Valid: true}}}
	svc := &CbtQuestion{q: store}
	base := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Bytes: [16]byte{7}, Valid: true},
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal reusable",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
		AnswerKey:      "A",
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		AuthorUsername: "guru.tanpa-event",
		Actor:          CbtQuestionActor{Username: "guru.tanpa-event", Roles: []string{"guru"}},
	}
	_, err := svc.Create(context.Background(), base)
	if err != nil {
		t.Fatalf("Create(guru reusable) error = %v", err)
	}
	if store.createCalls != 1 || store.createHistory[0].EventID.Valid {
		t.Fatalf("Create(guru reusable) calls/event = %d/%+v, want one global reusable create", store.createCalls, store.createHistory[0].EventID)
	}

	permissionOnly := base
	permissionOnly.AuthorUsername = "creator.permission"
	permissionOnly.Actor = CbtQuestionActor{Username: "creator.permission", Roles: []string{}, Permissions: []string{"bank_soal.create"}}
	_, err = svc.Create(context.Background(), permissionOnly)
	if err != nil {
		t.Fatalf("Create(permission-only reusable) error = %v", err)
	}

	blocked := base
	blocked.AuthorUsername = "staf.tu"
	blocked.Actor = CbtQuestionActor{Username: "staf.tu", Roles: []string{"staf"}}
	_, err = svc.Create(context.Background(), blocked)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Create(non-guru reusable) error = %v, want ErrForbidden", err)
	}
}

func TestCbtQuestionGetDetailScopesAndRedactsAnswerKey(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	store := &fakeQuestionStore{detail: db.GetCbtQuestionDetailRow{
		ID:             questionID,
		SubjectID:      subjectID,
		Status:         db.CbtQuestionStatusEnumPublished,
		AuthorUsername: "author",
		AnswerKey:      "A",
		RubricHtml:     "<p>Rubrik</p>",
	}}
	svc := &CbtQuestion{q: store}
	row, err := svc.GetDetail(context.Background(), questionID, CbtQuestionActor{Username: "guru.lain", Roles: []string{"guru"}})
	if err != nil {
		t.Fatalf("GetDetail(published) error = %v", err)
	}
	if row.AnswerKey != "" || row.RubricHtml != "" {
		t.Fatalf("GetDetail(published unauthorized) answer/rubric = %q/%q, want redacted", row.AnswerKey, row.RubricHtml)
	}

	store.detail.Status = db.CbtQuestionStatusEnumDraft
	store.detail.WorkflowStatus = "submitted"
	_, err = svc.GetDetail(context.Background(), questionID, CbtQuestionActor{Username: "guru.lain", Roles: []string{"guru"}})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("GetDetail(submitted other author) error = %v, want ErrForbidden", err)
	}

	store.canReview = true
	reviewerID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	row, err = svc.GetDetail(context.Background(), questionID, CbtQuestionActor{
		UserID:      reviewerID,
		Username:    "reviewer.ipa",
		Roles:       []string{"guru"},
		Permissions: []string{"bank_soal.review"},
	})
	if err != nil {
		t.Fatalf("GetDetail(scoped reviewer) error = %v", err)
	}
	if row.AnswerKey != "A" || row.RubricHtml != "<p>Rubrik</p>" {
		t.Fatalf("GetDetail(scoped reviewer) answer/rubric = %q/%q, want visible", row.AnswerKey, row.RubricHtml)
	}

	store.canReview = false
	store.canApprove = true
	store.detail.WorkflowStatus = "reviewed"
	approverID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	row, err = svc.GetDetail(context.Background(), questionID, CbtQuestionActor{
		UserID:      approverID,
		Username:    "approver.ipa",
		Roles:       []string{"guru"},
		Permissions: []string{"bank_soal.approve"},
	})
	if err != nil {
		t.Fatalf("GetDetail(scoped approver) error = %v", err)
	}
	if row.AnswerKey != "A" || row.RubricHtml != "<p>Rubrik</p>" {
		t.Fatalf("GetDetail(scoped approver) answer/rubric = %q/%q, want visible", row.AnswerKey, row.RubricHtml)
	}
}

func TestCbtQuestionReviewerCanApproveAssignedEventQuestion(t *testing.T) {
	eventID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{5}, Valid: true}
	questionID := pgtype.UUID{Bytes: [16]byte{6}, Valid: true}
	store := &fakeQuestionStore{
		current: db.GetCbtQuestionRow{
			ID: questionID, EventID: eventID, SubjectID: subjectID, QuestionText: "Soal", QuestionType: "multiple_choice",
			Options:   []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"},{"label":"C","text":"C"},{"label":"D","text":"D"}]`),
			AnswerKey: "A", Difficulty: db.CbtQuestionDifficultyEnumMedium, Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "review",
		},
		canApprove: true,
	}
	svc := &CbtQuestion{q: store}
	actorID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	_, err := svc.Approve(context.Background(), questionID, CbtQuestionActor{UserID: actorID, Username: "reviewer", Roles: []string{"guru"}, Permissions: []string{"bank_soal.approve"}}, "siap")
	if err != nil {
		t.Fatalf("Approve(reviewer) error = %v", err)
	}
	if store.updateParams.WorkflowStatus != "approved" {
		t.Fatalf("Approve(reviewer) workflow = %q, want approved", store.updateParams.WorkflowStatus)
	}

	store.updateParams = db.UpdateCbtQuestionParams{}
	store.canReview = false
	_, err = svc.Reject(context.Background(), questionID, CbtQuestionActor{UserID: actorID, Username: "bukan-reviewer", Roles: []string{"guru"}, Permissions: []string{"bank_soal.review"}}, "tolak")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Reject(unauthorized) error = %v, want ErrForbidden", err)
	}
}

func TestCbtQuestionNonAdminCannotReviewGlobalQuestion(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{5}, Valid: true}
	questionID := pgtype.UUID{Bytes: [16]byte{6}, Valid: true}
	store := &fakeQuestionStore{
		current: db.GetCbtQuestionRow{
			ID: questionID, SubjectID: subjectID, QuestionText: "Soal global", QuestionType: "multiple_choice",
			Options:   []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"},{"label":"C","text":"C"},{"label":"D","text":"D"}]`),
			AnswerKey: "A", Difficulty: db.CbtQuestionDifficultyEnumMedium, Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "review",
		},
	}
	svc := &CbtQuestion{q: store}

	_, err := svc.Approve(context.Background(), questionID, CbtQuestionActor{Username: "guru.ipa", Roles: []string{"guru"}}, "siap")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Approve(global non-admin) error = %v, want ErrForbidden", err)
	}
	_, err = svc.Reject(context.Background(), questionID, CbtQuestionActor{Username: "guru.ipa", Roles: []string{"guru"}}, "revisi")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Reject(global non-admin) error = %v, want ErrForbidden", err)
	}
	if store.updateCalls != 0 {
		t.Fatalf("global non-admin review update calls = %d, want 0", store.updateCalls)
	}
}

func workflowQuestionRow(id pgtype.UUID, workflowStatus string) db.GetCbtQuestionRow {
	return db.GetCbtQuestionRow{
		ID:             id,
		SubjectID:      pgtype.UUID{Bytes: [16]byte{55}, Valid: true},
		Code:           "WF-1",
		QuestionText:   "Soal workflow",
		QuestionType:   "multiple_choice",
		Options:        []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"},{"label":"C","text":"C"},{"label":"D","text":"D"}]`),
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: workflowStatus,
		AuthorUsername: "author.guru",
		ReviewNotes:    "catatan awal",
		TargetLevel:    pgtype.Text{String: "VIII", Valid: true},
	}
}

func TestCbtQuestionWorkflowTimelineAndNilRows(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{41}, Valid: true}
	store := &fakeQuestionStore{detail: db.GetCbtQuestionDetailRow{ID: questionID, Status: db.CbtQuestionStatusEnumPublished}}
	svc := &CbtQuestion{q: store}
	actor := CbtQuestionActor{Username: "admin", Roles: []string{"admin"}}

	rows, err := svc.Timeline(context.Background(), questionID, actor)
	if err != nil {
		t.Fatalf("Timeline(nil rows) error = %v", err)
	}
	if rows == nil || len(rows) != 0 {
		t.Fatalf("Timeline(nil rows) = %#v, want empty non-nil slice", rows)
	}

	store.auditLogs = []db.CbtQuestionAuditLog{{QuestionID: questionID, ActorUsername: "admin", Action: "mark_reviewed", Note: "ok"}}
	rows, err = svc.Timeline(context.Background(), questionID, actor)
	if err != nil {
		t.Fatalf("Timeline(log rows) error = %v", err)
	}
	if len(rows) != 1 || rows[0].Action != "mark_reviewed" || rows[0].QuestionID != questionID {
		t.Fatalf("Timeline(log rows) = %+v, want mapped audit row", rows)
	}
}

func TestCbtQuestionRequestRevisionAndMarkReviewedWorkflow(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{42}, Valid: true}
	actor := CbtQuestionActor{Username: "admin", Roles: []string{"admin"}}

	store := &fakeQuestionStore{current: workflowQuestionRow(questionID, "submitted")}
	svc := &CbtQuestion{q: store}
	row, err := svc.RequestRevision(context.Background(), questionID, actor, "perbaiki indikator")
	if err != nil {
		t.Fatalf("RequestRevision() error = %v", err)
	}
	if row.WorkflowStatus != "revision_needed" || row.ReviewerUsername != "admin" || store.updateParams.ApproverUsername != "" {
		t.Fatalf("RequestRevision() row/update = %+v/%+v, want revision_needed by reviewer and cleared approver", row, store.updateParams)
	}
	if len(store.workflowEvents) != 1 || store.workflowEvents[0].FromStatus != "submitted" || store.workflowEvents[0].ToStatus != "revision_needed" || store.workflowEvents[0].Action != "request_revision" {
		t.Fatalf("RequestRevision() workflow events = %+v, want submitted -> revision_needed", store.workflowEvents)
	}
	if store.updateParams.ReviewNotes != "perbaiki indikator" {
		t.Fatalf("RequestRevision() review notes = %q, want provided revision note", store.updateParams.ReviewNotes)
	}

	store = &fakeQuestionStore{current: workflowQuestionRow(questionID, "review")}
	svc = &CbtQuestion{q: store}
	row, err = svc.MarkReviewed(context.Background(), questionID, actor, "layak")
	if err != nil {
		t.Fatalf("MarkReviewed() error = %v", err)
	}
	if row.WorkflowStatus != "reviewed" || row.ReviewerUsername != "admin" {
		t.Fatalf("MarkReviewed() row = %+v, want reviewed by admin", row)
	}
	if len(store.workflowEvents) != 1 || store.workflowEvents[0].Action != "mark_reviewed" || store.workflowEvents[0].ToStatus != "reviewed" {
		t.Fatalf("MarkReviewed() workflow events = %+v, want mark_reviewed event", store.workflowEvents)
	}

	store = &fakeQuestionStore{current: workflowQuestionRow(questionID, "approved")}
	svc = &CbtQuestion{q: store}
	row, err = svc.ReturnToRevision(context.Background(), questionID, actor, "turunkan ke revisi")
	if err != nil {
		t.Fatalf("ReturnToRevision() error = %v", err)
	}
	if row.WorkflowStatus != "revision_needed" || len(store.workflowEvents) != 1 || store.workflowEvents[0].Action != "request_revision" {
		t.Fatalf("ReturnToRevision() row/events = %+v/%+v, want RequestRevision alias", row, store.workflowEvents)
	}
}

func TestCbtQuestionBulkWorkflowActionNormalizationAndItemResults(t *testing.T) {
	firstID := pgtype.UUID{Bytes: [16]byte{43}, Valid: true}
	secondID := pgtype.UUID{Bytes: [16]byte{44}, Valid: true}
	store := &fakeQuestionStore{current: workflowQuestionRow(firstID, "submitted")}
	svc := &CbtQuestion{q: store}
	actor := CbtQuestionActor{Username: "admin", Roles: []string{"admin"}}

	result, err := svc.BulkWorkflow(context.Background(), BulkCbtQuestionWorkflowInput{
		QuestionIDs: []pgtype.UUID{firstID, secondID},
		Action:      " RETURN_REVISION ",
		Notes:       "perlu revisi",
		Actor:       actor,
	})
	if err != nil {
		t.Fatalf("BulkWorkflow(return revision alias) error = %v", err)
	}
	if result.Action != "request_revision" || result.Total != 2 || result.Success != 2 || result.Failed != 0 {
		t.Fatalf("BulkWorkflow() result = %+v, want normalized action and two successes", result)
	}
	if len(result.Items) != 2 || !result.Items[0].OK || result.Items[0].Workflow != "revision_needed" || result.Items[1].QuestionID != secondID {
		t.Fatalf("BulkWorkflow() items = %+v, want per-question success items", result.Items)
	}
	if store.updateCalls != 2 || len(store.workflowEvents) != 2 {
		t.Fatalf("BulkWorkflow() update/event calls = %d/%d, want 2/2", store.updateCalls, len(store.workflowEvents))
	}

	_, err = svc.BulkWorkflow(context.Background(), BulkCbtQuestionWorkflowInput{QuestionIDs: []pgtype.UUID{firstID}, Action: "bogus", Actor: actor})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("BulkWorkflow(unsupported) error = %v, want ErrBadRequest", err)
	}
}

func TestCbtQuestionApplyBulkWorkflowItemDispatchAndUnsupported(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{45}, Valid: true}
	actor := CbtQuestionActor{Username: "admin", Roles: []string{"admin"}}

	store := &fakeQuestionStore{current: workflowQuestionRow(questionID, "submitted")}
	svc := &CbtQuestion{q: store}
	if _, err := svc.applyBulkWorkflowItem(context.Background(), questionID, "mark_reviewed", actor, "ok"); err != nil {
		t.Fatalf("applyBulkWorkflowItem(mark_reviewed) error = %v", err)
	}
	if store.updateParams.WorkflowStatus != "reviewed" || len(store.workflowEvents) != 1 || store.workflowEvents[0].Action != "mark_reviewed" {
		t.Fatalf("applyBulkWorkflowItem(mark_reviewed) update/events = %+v/%+v", store.updateParams, store.workflowEvents)
	}

	store = &fakeQuestionStore{current: workflowQuestionRow(questionID, "approved")}
	svc = &CbtQuestion{q: store}
	if _, err := svc.applyBulkWorkflowItem(context.Background(), questionID, "publish", actor, "ignored"); err != nil {
		t.Fatalf("applyBulkWorkflowItem(publish) error = %v", err)
	}
	if store.updateParams.WorkflowStatus != "published" || store.updateParams.Status != db.CbtQuestionStatusEnumPublished {
		t.Fatalf("applyBulkWorkflowItem(publish) update = %+v, want published status", store.updateParams)
	}

	_, err := svc.applyBulkWorkflowItem(context.Background(), questionID, "unsupported", actor, "")
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("applyBulkWorkflowItem(unsupported) error = %v, want ErrBadRequest", err)
	}
}

func importCSVRow(values ...string) string {
	return strings.Join(values, ";")
}

func TestCbtQuestionExportCSVMapsStructuredTypes(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	multipleAnswerOptions, err := EncodeQuestionOptions([]QuestionOption{
		{Label: "A", Text: "Satu"},
		{Label: "B", Text: "Dua"},
		{Label: "C", Text: "Tiga"},
		{Label: "D", Text: "Empat"},
		{Label: "E", Text: "Enam"},
	})
	if err != nil {
		t.Fatalf("EncodeQuestionOptions(multiple answer) error = %v", err)
	}
	matchingOptions, err := EncodeQuestionOptions([]QuestionOption{
		{Label: "A", Text: "Satu", MatchLabel: "1", MatchText: "1"},
		{Label: "B", Text: "Dua", MatchLabel: "2", MatchText: "2"},
		{MatchLabel: "3", MatchText: "Tiga", IsDistractor: true},
	})
	if err != nil {
		t.Fatalf("EncodeQuestionOptions(matching) error = %v", err)
	}
	store := &fakeQuestionStore{
		filteredRows: []db.ListCbtQuestionsFilteredRow{
			{
				SubjectID:      subjectID,
				SubjectName:    "Matematika",
				Code:           "MTK-G",
				QuestionText:   "Pilih bilangan genap",
				QuestionType:   "multiple_answer",
				Options:        multipleAnswerOptions,
				AnswerKey:      "B,E",
				Difficulty:     db.CbtQuestionDifficultyEnumMedium,
				Status:         db.CbtQuestionStatusEnumDraft,
				WorkflowStatus: "draft",
				TargetLevel:    pgtype.Text{String: "VIII", Valid: true},
				HotsFlag:       true,
			},
			{
				SubjectID:      subjectID,
				SubjectName:    "Matematika",
				Code:           "MTK-M",
				QuestionText:   "Pasangkan angka",
				QuestionType:   "matching",
				Options:        matchingOptions,
				AnswerKey:      "A=1;B=2",
				Difficulty:     db.CbtQuestionDifficultyEnumMedium,
				Status:         db.CbtQuestionStatusEnumDraft,
				WorkflowStatus: "draft",
			},
		},
		count: 2,
	}
	svc := &CbtQuestion{q: store}

	got, err := svc.ExportCSV(context.Background(), ListCbtQuestionsInput{
		SubjectID:      subjectID,
		AuthorUsername: " guru.ipa ",
		WorkflowStatus: " draft ",
		Limit:          0,
	})
	if err != nil {
		t.Fatalf("ExportCSV() error = %v", err)
	}
	if got.Count != 2 || !strings.HasPrefix(got.Filename, "bank-soal-") {
		t.Fatalf("ExportCSV() result = %+v, want count and generated filename", got)
	}
	if store.listFilterArg.LimitCount != 2000 || store.listFilterArg.WorkflowStatus != "draft" || store.listFilterArg.AuthorUsername != "guru.ipa" {
		t.Fatalf("ExportCSV() list arg = %+v, want default export limit and trimmed workflow", store.listFilterArg)
	}
	records, err := csv.NewReader(strings.NewReader(string(got.Content))).ReadAll()
	if err != nil {
		t.Fatalf("ExportCSV() CSV parse error = %v content=%s", err, string(got.Content))
	}
	if len(records) != 3 {
		t.Fatalf("ExportCSV() records = %d, want header + 2 rows", len(records))
	}
	header := csvHeaderIndex(records[0])
	if _, ok := header["grade_level"]; ok {
		t.Fatalf("ExportCSV() header exposed legacy grade_level: %+v", records[0])
	}
	if _, ok := header["target_level"]; !ok {
		t.Fatalf("ExportCSV() header missing target_level: %+v", records[0])
	}
	if records[1][header["tipe"]] != "pg_kompleks" || records[1][header["opsi_e"]] != "Enam" || records[1][header["jawaban"]] != "B,E" || records[1][header["target_level"]] != "VIII" || records[1][header["hots_flag"]] != "true" {
		t.Fatalf("multiple answer CSV row = %+v, want type/options/key/metadata mapped", records[1])
	}
	if records[2][header["tipe"]] != "menjodohkan" || records[2][header["kiri_a"]] != "Satu" || records[2][header["kanan_2"]] != "2" || records[2][header["distraktor_1"]] != "Tiga" {
		t.Fatalf("matching CSV row = %+v, want pairs and distractor mapped", records[2])
	}

	importStore := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{9}, Valid: true}},
	}
	importSvc := &CbtQuestion{q: importStore}
	imported, err := importSvc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   string(got.Content),
		Username:  "guru.import",
		Actor:     CbtQuestionActor{Username: "guru.import", Roles: []string{"admin"}},
	})
	if err != nil {
		t.Fatalf("roundtrip ImportLegacyCSV() error = %v", err)
	}
	if imported.Imported != 2 || imported.Skipped != 0 {
		t.Fatalf("roundtrip import result = %+v, want 2 imported and no skipped rows", imported)
	}
	if importStore.createHistory[0].QuestionType != "multiple_answer" || importStore.createHistory[0].AnswerKey != "B,E" || importStore.createHistory[0].TargetLevel.String != "VIII" || !importStore.createHistory[0].HotsFlag {
		t.Fatalf("roundtrip multiple answer params = %+v, want type/key/metadata preserved", importStore.createHistory[0])
	}
	if importStore.createHistory[1].QuestionType != "matching" || importStore.createHistory[1].AnswerKey != "A=1;B=2" || !strings.Contains(string(importStore.createHistory[1].Options), `"is_distractor":true`) {
		t.Fatalf("roundtrip matching params = %+v options %s, want matching pair and distractor preserved", importStore.createHistory[1], string(importStore.createHistory[1].Options))
	}
}

func TestCbtQuestionTemplateCSVRoundtripsThroughImport(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	svc := &CbtQuestion{q: &fakeQuestionStore{}}
	template, err := svc.TemplateCSV()
	if err != nil {
		t.Fatalf("TemplateCSV() error = %v", err)
	}
	if template.Count < 6 || template.Filename != "template-bank-soal.csv" {
		t.Fatalf("TemplateCSV() result = %+v, want sample rows and stable filename", template)
	}
	records, err := csv.NewReader(strings.NewReader(string(template.Content))).ReadAll()
	if err != nil {
		t.Fatalf("TemplateCSV() CSV parse error = %v content=%s", err, string(template.Content))
	}
	header := csvHeaderIndex(records[0])
	if header["tipe"] == 0 || header["rubrik"] == 0 || header["distraktor_1"] == 0 {
		t.Fatalf("TemplateCSV() header = %+v, want multi-type columns", records[0])
	}
	if _, ok := header["grade_level"]; ok {
		t.Fatalf("TemplateCSV() header exposed legacy grade_level: %+v", records[0])
	}
	if _, ok := header["target_level"]; !ok {
		t.Fatalf("TemplateCSV() header missing target_level: %+v", records[0])
	}

	importStore := &fakeQuestionStore{
		createRow: db.CbtQuestion{ID: pgtype.UUID{Bytes: [16]byte{10}, Valid: true}},
	}
	importSvc := &CbtQuestion{q: importStore}
	imported, err := importSvc.ImportLegacyCSV(context.Background(), ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		CSVText:   string(template.Content),
		Username:  "guru.template",
		Actor:     CbtQuestionActor{Username: "guru.template", Roles: []string{"admin"}},
	})
	if err != nil {
		t.Fatalf("template ImportLegacyCSV() error = %v", err)
	}
	if imported.Imported != template.Count || imported.Skipped != 0 {
		t.Fatalf("template import result = %+v, want all template rows imported", imported)
	}
	types := make(map[string]bool, len(importStore.createHistory))
	for _, params := range importStore.createHistory {
		types[params.QuestionType] = true
	}
	for _, questionType := range []string{"multiple_choice", "multiple_answer", "true_false", "short_answer", "matching", "essay"} {
		if !types[questionType] {
			t.Fatalf("template import types = %+v, want %s", types, questionType)
		}
	}
}

func csvHeaderIndex(headers []string) map[string]int {
	out := make(map[string]int, len(headers))
	for idx, header := range headers {
		out[header] = idx
	}
	return out
}

func TestNormalizeQuestionInputSanitizesDangerousHTML(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		StemHTML:       `<p onclick="alert(1)" style="text-align:center;color:#166534">Halo</p><script>alert(2)</script>`,
		StimulusHTML:   `<img src="javascript:alert(1)" onerror="alert(1)" style="background:url(javascript:alert(2))"><table><tr><td>Data</td></tr></table>`,
		QuestionText:   "",
		Options:        []QuestionOption{{Label: "A", HTML: `<span onclick="x()"><a href=javascript:alert(1)>Aman</a></span>`}, {Label: "B", Text: "B"}},
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
	}

	got, err := normalizeQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	for _, htmlValue := range []string{got.StemHTML, got.StimulusHTML, got.Options[0].HTML} {
		assertQuestionHTMLSafe(t, htmlValue)
	}
	if !strings.Contains(got.StemHTML, `<p style="text-align: center; color: #166534">Halo</p>`) {
		t.Fatalf("StemHTML = %q, want rich paragraph formatting preserved", got.StemHTML)
	}
	if !strings.Contains(got.StimulusHTML, `<table><tr><td>Data</td></tr></table>`) {
		t.Fatalf("StimulusHTML = %q, want table preserved after unsafe image is dropped", got.StimulusHTML)
	}
	if !strings.Contains(got.Options[0].HTML, `<span>Aman</span>`) {
		t.Fatalf("Option HTML = %q, want unsafe link dropped with text preserved", got.Options[0].HTML)
	}
}

func TestQuestionHTMLSanitizerBlocksCommonBypassPayloads(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		required string
	}{
		{
			name:     "encoded javascript URL",
			input:    `<a href="jav&#x61;script:alert(1)">Klik</a>`,
			required: `Klik`,
		},
		{
			name:     "svg onload content",
			input:    `<svg><g onload="alert(1)"><text>Jangan</text></g></svg><p>Aman</p>`,
			required: `<p>Aman</p>`,
		},
		{
			name:     "malformed img handler",
			input:    `<p><strong>Utuh<img src=x onerror=alert(1)</p>`,
			required: `Utuh`,
		},
		{
			name:     "data html URL",
			input:    `<a href="data:text/html,&lt;script&gt;alert(1)&lt;/script&gt;">Data</a>`,
			required: `Data`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeHTML(tt.input)
			assertQuestionHTMLSafe(t, got)
			if !strings.Contains(got, tt.required) {
				t.Fatalf("sanitizeHTML() = %q, want to preserve %q", got, tt.required)
			}
		})
	}
}

func assertQuestionHTMLSafe(t *testing.T, value string) {
	t.Helper()
	lowered := strings.ToLower(value)
	for _, forbidden := range []string{
		"javascript:", "vbscript:", "data:text/html", "<script", "<style", "<iframe", "<object", "<embed", "<svg",
		" onload", " onclick", " onerror", "srcdoc", "formaction", "xlink:href",
	} {
		if strings.Contains(lowered, forbidden) {
			t.Fatalf("sanitized HTML %q still contains forbidden fragment %q", value, forbidden)
		}
	}
}

func TestValidateQuestionRequiresApprovedBeforePublish(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal uji",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}},
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumPublished,
		WorkflowStatus: "draft",
	}
	_, err := normalizeQuestionInput(input)
	if err == nil {
		t.Fatal("normalizeQuestionInput() error = nil, want publish gating error")
	}
}

func TestSubmitReviewUpdatesWorkflowAndReviewer(t *testing.T) {
	store := &fakeQuestionStore{
		current: db.GetCbtQuestionRow{
			ID:             pgtype.UUID{Valid: true},
			SubjectID:      pgtype.UUID{Valid: true},
			QuestionText:   "Soal",
			QuestionType:   "multiple_choice",
			Options:        []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"}]`),
			AnswerKey:      "A",
			Difficulty:     db.CbtQuestionDifficultyEnumMedium,
			Status:         db.CbtQuestionStatusEnumDraft,
			WorkflowStatus: "draft",
		},
	}
	svc := &CbtQuestion{q: store}

	_, err := svc.SubmitReview(context.Background(), store.current.ID, CbtQuestionActor{Username: "reviewer1", Roles: []string{"admin"}}, "cek redaksi")
	if err != nil {
		t.Fatalf("SubmitReview() error = %v", err)
	}
	if store.updateParams.WorkflowStatus != "submitted" {
		t.Fatalf("WorkflowStatus = %q, want submitted", store.updateParams.WorkflowStatus)
	}
	if store.updateParams.ReviewerUsername != "" {
		t.Fatalf("ReviewerUsername = %q, want empty until reviewer action", store.updateParams.ReviewerUsername)
	}
}

func TestCbtQuestionWorkflowActions(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	versionGroupID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	current := db.GetCbtQuestionRow{
		ID:              questionID,
		SubjectID:       subjectID,
		Code:            "Q-1",
		QuestionText:    "Soal",
		QuestionType:    "multiple_choice",
		Options:         []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"},{"label":"C","text":"C"},{"label":"D","text":"D"}]`),
		OptionA:         "A",
		OptionB:         "B",
		OptionC:         "C",
		OptionD:         "D",
		AnswerKey:       "A",
		Difficulty:      db.CbtQuestionDifficultyEnumMedium,
		Status:          db.CbtQuestionStatusEnumDraft,
		WorkflowStatus:  "review",
		ReviewNotes:     "catatan lama",
		VersionGroupID:  versionGroupID,
		VersionNumber:   3,
		IsLatestVersion: true,
	}

	t.Run("approve updates workflow and notes", func(t *testing.T) {
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		_, err := svc.Approve(context.Background(), questionID, CbtQuestionActor{Username: "waka", Roles: []string{"admin"}}, "siap")
		if err != nil {
			t.Fatalf("Approve() error = %v", err)
		}
		if store.updateParams.WorkflowStatus != "approved" || store.updateParams.ReviewerUsername != "waka" {
			t.Fatalf("Approve() params = %+v, want approved reviewer waka", store.updateParams)
		}
		if !store.updateParams.ReviewedAt.Valid || store.updateParams.ReviewNotes != "siap" {
			t.Fatalf("Approve() review timestamp/notes = %v/%q, want valid/siap", store.updateParams.ReviewedAt, store.updateParams.ReviewNotes)
		}
		if len(store.workflowEvents) != 1 || store.workflowEvents[0].FromStatus != "review" || store.workflowEvents[0].ToStatus != "approved" {
			t.Fatalf("Approve() workflow events = %+v, want review -> approved", store.workflowEvents)
		}
		var metadata map[string]string
		if err := json.Unmarshal(store.workflowEvents[0].Metadata, &metadata); err != nil {
			t.Fatalf("Approve() workflow metadata decode error = %v", err)
		}
		if metadata["to_approver_username"] != "waka" || metadata["to_publication_status"] != "draft" {
			t.Fatalf("Approve() workflow metadata = %+v, want approver/status transition", metadata)
		}
	})

	t.Run("reject updates workflow reviewer and notes", func(t *testing.T) {
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		_, err := svc.Reject(context.Background(), questionID, CbtQuestionActor{Username: "waka", Roles: []string{"admin"}}, "perbaiki opsi C")
		if err != nil {
			t.Fatalf("Reject() error = %v", err)
		}
		if store.updateParams.WorkflowStatus != "rejected" || store.updateParams.ReviewerUsername != "waka" {
			t.Fatalf("Reject() params = %+v, want rejected reviewer waka", store.updateParams)
		}
		if !store.updateParams.ReviewedAt.Valid || store.updateParams.ReviewNotes != "perbaiki opsi C" {
			t.Fatalf("Reject() review timestamp/notes = %v/%q, want valid/notes", store.updateParams.ReviewedAt, store.updateParams.ReviewNotes)
		}
		if store.updateParams.ApproverUsername != "" {
			t.Fatalf("Reject() approver = %q, want cleared", store.updateParams.ApproverUsername)
		}
	})

	t.Run("publish sets published status and approver", func(t *testing.T) {
		approved := current
		approved.WorkflowStatus = "approved"
		store := &fakeQuestionStore{current: approved}
		svc := &CbtQuestion{q: store}

		_, err := svc.Publish(context.Background(), questionID, CbtQuestionActor{Username: "kepala", Roles: []string{"admin"}})
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
		if store.updateParams.Status != db.CbtQuestionStatusEnumPublished || store.updateParams.WorkflowStatus != "published" {
			t.Fatalf("Publish() params = %+v, want published workflow", store.updateParams)
		}
		if store.updateParams.ApproverUsername != "kepala" || !store.updateParams.ApprovedAt.Valid {
			t.Fatalf("Publish() approver = %q/%v, want kepala with timestamp", store.updateParams.ApproverUsername, store.updateParams.ApprovedAt)
		}
	})

	t.Run("publish allows granular permission without admin role", func(t *testing.T) {
		approved := current
		approved.WorkflowStatus = "approved"
		store := &fakeQuestionStore{current: approved, canApprove: true}
		svc := &CbtQuestion{q: store}
		publisherID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}

		_, err := svc.Publish(context.Background(), questionID, CbtQuestionActor{UserID: publisherID, Username: "publisher", Permissions: []string{"bank_soal.publish"}})
		if err != nil {
			t.Fatalf("Publish(permission) error = %v", err)
		}
		if store.updateParams.Status != db.CbtQuestionStatusEnumPublished || store.updateParams.ApproverUsername != "publisher" {
			t.Fatalf("Publish(permission) params = %+v, want published by publisher", store.updateParams)
		}
	})

	t.Run("publish rejects reviewer without publish permission", func(t *testing.T) {
		approved := current
		approved.WorkflowStatus = "approved"
		store := &fakeQuestionStore{current: approved}
		svc := &CbtQuestion{q: store}

		_, err := svc.Publish(context.Background(), questionID, CbtQuestionActor{Username: "reviewer", Permissions: []string{"bank_soal.review"}})
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("Publish(reviewer) error = %v, want ErrForbidden", err)
		}
	})

	t.Run("archive sets archived status", func(t *testing.T) {
		approved := current
		approved.WorkflowStatus = "approved"
		store := &fakeQuestionStore{current: approved}
		svc := &CbtQuestion{q: store}

		_, err := svc.Archive(context.Background(), questionID, CbtQuestionActor{Username: "admin", Roles: []string{"admin"}})
		if err != nil {
			t.Fatalf("Archive() error = %v", err)
		}
		if store.updateParams.Status != db.CbtQuestionStatusEnumArchived {
			t.Fatalf("Archive() status = %q, want archived", store.updateParams.Status)
		}
	})

	t.Run("archive rejects used question", func(t *testing.T) {
		used := current
		used.PackageCount = 1
		store := &fakeQuestionStore{current: used}
		svc := &CbtQuestion{q: store}

		_, err := svc.Archive(context.Background(), questionID, CbtQuestionActor{Username: "admin", Roles: []string{"admin"}})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("Archive(used) error = %v, want ErrConflict", err)
		}
		if store.updateCalls != 0 {
			t.Fatalf("Archive(used) update calls = %d, want 0", store.updateCalls)
		}
	})

	t.Run("duplicate creates clean draft copy", func(t *testing.T) {
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		_, err := svc.DuplicateAsDraft(context.Background(), questionID, CbtQuestionActor{Username: "guru", Roles: []string{"admin"}})
		if err != nil {
			t.Fatalf("DuplicateAsDraft() error = %v", err)
		}
		if store.createCalls != 1 {
			t.Fatalf("CreateCbtQuestion() calls = %d, want 1", store.createCalls)
		}
		if store.createParams.Code != "Q-1-COPY" || store.createParams.Status != db.CbtQuestionStatusEnumDraft || store.createParams.WorkflowStatus != "draft" {
			t.Fatalf("DuplicateAsDraft() create params = %+v, want clean draft copy", store.createParams)
		}
		if store.createParams.ReviewerUsername != "" || store.createParams.ApproverUsername != "" || store.createParams.ReviewNotes != "" {
			t.Fatalf("DuplicateAsDraft() reviewer/approver/notes = %q/%q/%q, want cleared", store.createParams.ReviewerUsername, store.createParams.ApproverUsername, store.createParams.ReviewNotes)
		}
		if store.createParams.VersionGroupID.Valid || store.createParams.VersionNumber != 1 || store.createParams.SourceQuestionID.Valid || store.createParams.SupersedesQuestionID.Valid {
			t.Fatalf("DuplicateAsDraft() lineage = %+v/%d/%+v/%+v, want independent v1 group", store.createParams.VersionGroupID, store.createParams.VersionNumber, store.createParams.SourceQuestionID, store.createParams.SupersedesQuestionID)
		}
	})

	t.Run("duplicate for revision creates rejected draft copy with notes", func(t *testing.T) {
		store := &fakeQuestionStore{current: current, nextVersionNumber: 4}
		svc := &CbtQuestion{q: store}

		_, err := svc.DuplicateForRevision(context.Background(), questionID, CbtQuestionActor{Username: "reviewer", Roles: []string{"admin"}}, "Daya pembeda rendah")
		if err != nil {
			t.Fatalf("DuplicateForRevision() error = %v", err)
		}
		if store.createCalls != 1 {
			t.Fatalf("CreateCbtQuestion() calls = %d, want 1", store.createCalls)
		}
		if !strings.HasPrefix(store.createParams.Code, "Q-1-REV-") || store.createParams.Status != db.CbtQuestionStatusEnumDraft || store.createParams.WorkflowStatus != "rejected" {
			t.Fatalf("DuplicateForRevision() create params = %+v, want rejected draft revision copy", store.createParams)
		}
		if store.createParams.ReviewerUsername != "reviewer" || store.createParams.ApproverUsername != "" || store.createParams.ReviewNotes != "Daya pembeda rendah" {
			t.Fatalf("DuplicateForRevision() reviewer/approver/notes = %q/%q/%q, want reviewer/no approver/notes", store.createParams.ReviewerUsername, store.createParams.ApproverUsername, store.createParams.ReviewNotes)
		}
		if store.createParams.VersionGroupID != versionGroupID || store.createParams.VersionNumber != 4 || store.createParams.SourceQuestionID != questionID || store.createParams.SupersedesQuestionID != questionID || store.createParams.VersionNote != "Daya pembeda rendah" {
			t.Fatalf("DuplicateForRevision() lineage = %+v/%d/%+v/%+v/%q, want same group v4 source/supersedes note", store.createParams.VersionGroupID, store.createParams.VersionNumber, store.createParams.SourceQuestionID, store.createParams.SupersedesQuestionID, store.createParams.VersionNote)
		}
		if store.markGroupLatestCalls != 1 || store.markGroupNotLatestID != versionGroupID {
			t.Fatalf("DuplicateForRevision() mark group latest calls/id = %d/%+v, want version group marked not latest", store.markGroupLatestCalls, store.markGroupNotLatestID)
		}
		if store.markNotLatestCalls != 0 {
			t.Fatalf("DuplicateForRevision() source-only mark calls = %d, want group mark only", store.markNotLatestCalls)
		}
	})
}

func TestCbtQuestionBulkWorkflowReturnsPerItemResults(t *testing.T) {
	questionID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	store := &fakeQuestionStore{current: db.GetCbtQuestionRow{
		ID:             questionID,
		SubjectID:      pgtype.UUID{Bytes: [16]byte{3}, Valid: true},
		QuestionText:   "Soal",
		QuestionType:   "multiple_choice",
		Options:        []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"}]`),
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "review",
	}}
	svc := &CbtQuestion{q: store}

	got, err := svc.BulkWorkflow(context.Background(), BulkCbtQuestionWorkflowInput{
		QuestionIDs: []pgtype.UUID{questionID},
		Action:      "approve",
		Notes:       "siap",
		Actor:       CbtQuestionActor{Username: "reviewer", Roles: []string{"admin"}},
	})
	if err != nil {
		t.Fatalf("BulkWorkflow() error = %v", err)
	}
	if got.Total != 1 || got.Success != 1 || got.Failed != 0 || len(got.Items) != 1 || !got.Items[0].OK {
		t.Fatalf("BulkWorkflow() result = %+v, want one successful item", got)
	}
	if store.updateParams.WorkflowStatus != "approved" || store.auditCalls != 1 || store.auditLogs[0].Action != "approve" {
		t.Fatalf("BulkWorkflow() update/audit = %+v/%+v, want approved audit", store.updateParams, store.auditLogs)
	}
}

func TestNormalizeQuestionInputBeginnerDefaultsToDraft(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal mudah",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
		AnswerKey:      "A",
		Status:         db.CbtQuestionStatusEnumPublished,
		WorkflowStatus: "approved",
		Difficulty:     db.CbtQuestionDifficultyEnumHard,
		WriterNotes:    "should be cleared",
	}

	got, err := normalizeQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	if got.Status != db.CbtQuestionStatusEnumDraft {
		t.Fatalf("Status = %q, want draft", got.Status)
	}
	if got.WorkflowStatus != "draft" {
		t.Fatalf("WorkflowStatus = %q, want draft", got.WorkflowStatus)
	}
	if got.Difficulty != db.CbtQuestionDifficultyEnumMedium {
		t.Fatalf("Difficulty = %q, want medium", got.Difficulty)
	}
	if got.WriterNotes != "" {
		t.Fatalf("WriterNotes = %q, want cleared", got.WriterNotes)
	}
}

func TestNormalizeQuestionInputClearsDraftReviewActors(t *testing.T) {
	got, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:        pgtype.UUID{Valid: true},
		AuthoringMode:    "advance",
		QuestionType:     "multiple_choice",
		QuestionText:     "Soal draft",
		Options:          []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
		AnswerKey:        "A",
		Status:           db.CbtQuestionStatusEnumDraft,
		WorkflowStatus:   "draft",
		ReviewerUsername: "guru",
		ApproverUsername: "admin",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	if got.ReviewerUsername != "" || got.ApproverUsername != "" {
		t.Fatalf("draft actors = reviewer %q approver %q, want cleared", got.ReviewerUsername, got.ApproverUsername)
	}
}

func TestNormalizeQuestionInputBeginnerSupportsJuknisTypes(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "short_answer",
		QuestionText:   "Jawab singkat",
		AnswerKey:      " Fotosintesis | fotosintesis | foto sintesis ",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
	}

	got, err := normalizeQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	if got.QuestionType != "short_answer" || got.AnswerKey != "Fotosintesis|foto sintesis" {
		t.Fatalf("normalizeQuestionInput() type/key = %q/%q, want short_answer normalized aliases", got.QuestionType, got.AnswerKey)
	}
}

func TestNormalizeShortAnswerKey(t *testing.T) {
	got := normalizeShortAnswerKey("  Iman kepada Allah | iman   kepada allah | Iman\u00a0Kepada\u00a0Rasul  | ")
	want := "Iman kepada Allah|Iman Kepada Rasul"
	if got != want {
		t.Fatalf("normalizeShortAnswerKey() = %q, want %q", got, want)
	}
	if got := normalizeShortAnswerComparable("  Iman\u00a0  Kepada   Allah "); got != "iman kepada allah" {
		t.Fatalf("normalizeShortAnswerComparable() = %q, want normalized lowercase whitespace", got)
	}
}

func TestCbtQuestionBuildUpdatePublishedParams(t *testing.T) {
	currentReviewedAt := pgtype.Timestamptz{Valid: true}
	currentApprovedAt := pgtype.Timestamptz{Valid: true}
	current := db.GetCbtQuestionRow{
		ID:               pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		ReviewerUsername: "old-reviewer",
		ReviewedAt:       currentReviewedAt,
		ApproverUsername: "old-approver",
		ApprovedAt:       currentApprovedAt,
	}

	params, err := buildUpdateQuestionParams(current, SaveCbtQuestionInput{
		ID:               current.ID,
		SubjectID:        pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		AuthoringMode:    "advance",
		QuestionType:     "essay",
		QuestionText:     "Uraikan proses fotosintesis",
		Status:           db.CbtQuestionStatusEnumPublished,
		WorkflowStatus:   "approved",
		RubricHTML:       "<p>Rubrik lengkap</p>",
		ReviewerUsername: "reviewer-baru",
		ApproverUsername: "approver-baru",
		ReviewNotes:      " siap ",
	})
	if err != nil {
		t.Fatalf("buildUpdateQuestionParams() error = %v", err)
	}
	if params.ID != current.ID || params.QuestionType != "essay" || params.Status != db.CbtQuestionStatusEnumPublished || params.WorkflowStatus != "approved" {
		t.Fatalf("buildUpdateQuestionParams() identity/status = %+v, want published approved essay", params)
	}
	if params.ReviewerUsername != "reviewer-baru" || !params.ReviewedAt.Valid {
		t.Fatalf("buildUpdateQuestionParams() reviewer = %q/%v, want reviewer-baru with timestamp", params.ReviewerUsername, params.ReviewedAt)
	}
	if params.ApproverUsername != "approver-baru" || !params.ApprovedAt.Valid {
		t.Fatalf("buildUpdateQuestionParams() approver = %q/%v, want approver-baru with timestamp", params.ApproverUsername, params.ApprovedAt)
	}
	if params.ReviewNotes != "siap" || params.RubricHtml != "<p>Rubrik lengkap</p>" {
		t.Fatalf("buildUpdateQuestionParams() notes/rubric = %q/%q, want trimmed sanitized fields", params.ReviewNotes, params.RubricHtml)
	}
}

func TestNormalizeQuestionInputValidationMatrix(t *testing.T) {
	tests := []struct {
		name    string
		input   SaveCbtQuestionInput
		wantErr string
	}{
		{
			name:    "missing question text",
			input:   SaveCbtQuestionInput{SubjectID: pgtype.UUID{Valid: true}, AuthoringMode: "advance", QuestionType: "essay"},
			wantErr: "question_text atau stem_html/stem_latex wajib diisi",
		},
		{
			name: "objective requires two options",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "multiple_answer",
				QuestionText:   "Pilih",
				Options:        []QuestionOption{{Label: "A", Text: "A"}},
				AnswerKey:      "A",
				WorkflowStatus: "review",
			},
			wantErr: "opsi jawaban minimal 2 untuk tipe soal objektif",
		},
		{
			name: "multiple answer requires two keys",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "multiple_answer",
				QuestionText:   "Pilih semua",
				Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
				AnswerKey:      "A",
				WorkflowStatus: "review",
			},
			wantErr: "multiple_answer membutuhkan minimal 2 kunci jawaban",
		},
		{
			name: "beginner multiple choice requires four options",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "beginner",
				QuestionType:   "multiple_choice",
				QuestionText:   "Pilih",
				Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}},
				AnswerKey:      "A",
				WorkflowStatus: "review",
			},
			wantErr: "mode beginner membutuhkan minimal 4 opsi untuk pilihan ganda",
		},
		{
			name: "short answer requires key",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "short_answer",
				QuestionText:   "Jawab",
				WorkflowStatus: "review",
			},
			wantErr: "answer_key wajib diisi untuk short_answer",
		},
		{
			name: "objective answer key must match option label",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "multiple_choice",
				QuestionText:   "Pilih",
				Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}},
				AnswerKey:      "F",
				WorkflowStatus: "review",
			},
			wantErr: "answer_key harus sesuai label opsi yang tersedia",
		},
		{
			name: "matching requires complete pair content",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "beginner",
				QuestionType:   "matching",
				QuestionText:   "Cocokkan",
				Options:        []QuestionOption{{Label: "A", Text: "Istilah", MatchLabel: "1"}},
				AnswerKey:      "A=1",
				WorkflowStatus: "review",
			},
			wantErr: "setiap pasangan menjodohkan wajib memiliki kolom kiri dan kanan",
		},
		{
			name: "matching rejects invalid non-empty draft key",
			input: SaveCbtQuestionInput{
				SubjectID:     pgtype.UUID{Valid: true},
				AuthoringMode: "advance",
				QuestionType:  "matching",
				QuestionText:  "Cocokkan",
				Options: []QuestionOption{
					{Label: "A", Text: "Satu", MatchLabel: "1", MatchText: "One"},
					{Label: "B", Text: "Dua", MatchLabel: "2", MatchText: "Two"},
				},
				AnswerKey:      "A=1;B=9",
				Status:         db.CbtQuestionStatusEnumDraft,
				WorkflowStatus: "draft",
			},
			wantErr: "answer_key menjodohkan harus sesuai label pasangan",
		},
		{
			name: "matching review requires answer key",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "matching",
				QuestionText:   "Cocokkan",
				Options:        []QuestionOption{{Label: "A", Text: "Satu", MatchLabel: "1", MatchText: "One"}, {Label: "B", Text: "Dua", MatchLabel: "2", MatchText: "Two"}},
				WorkflowStatus: "review",
			},
			wantErr: "answer_key menjodohkan tidak valid",
		},
		{
			name: "matching rejects duplicated right key",
			input: SaveCbtQuestionInput{
				SubjectID:      pgtype.UUID{Valid: true},
				AuthoringMode:  "advance",
				QuestionType:   "matching",
				QuestionText:   "Cocokkan",
				Options:        []QuestionOption{{Label: "A", Text: "Satu", MatchLabel: "1", MatchText: "One"}, {Label: "B", Text: "Dua", MatchLabel: "2", MatchText: "Two"}},
				AnswerKey:      "A=1;B=1",
				WorkflowStatus: "review",
			},
			wantErr: "answer_key menjodohkan harus sesuai label pasangan",
		},
		{
			name: "unsupported type",
			input: SaveCbtQuestionInput{
				SubjectID:     pgtype.UUID{Valid: true},
				AuthoringMode: "advance",
				QuestionType:  "hotspot",
				QuestionText:  "Cocokkan",
			},
			wantErr: "question_type tidak didukung",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeQuestionInput(tt.input)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("normalizeQuestionInput() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeQuestionInputAllowsEssayWithoutRubric(t *testing.T) {
	input, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "essay",
		QuestionText:   "Uraikan hikmah salat berjamaah",
		WorkflowStatus: "approved",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(approved essay without rubric) error = %v", err)
	}
	if input.QuestionType != "essay" || input.RubricHTML != "" {
		t.Fatalf("normalizeQuestionInput() = type %q rubric %q, want essay with empty rubric", input.QuestionType, input.RubricHTML)
	}
}

func TestCbtQuestionNormalizeAndEncodingHelpers(t *testing.T) {
	draft, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "beginner",
		QuestionType:  "multiple_choice",
		QuestionText:  "Draft awal",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(draft partial) error = %v", err)
	}
	if draft.WorkflowStatus != "draft" || len(draft.Options) != 0 {
		t.Fatalf("normalizeQuestionInput(draft partial) workflow/options = %q/%d, want draft/0", draft.WorkflowStatus, len(draft.Options))
	}

	matchingDraft, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "advance",
		QuestionType:  "matching",
		QuestionText:  "Draft menjodohkan",
		Options:       []QuestionOption{{Label: "A", Text: "Satu"}, {Label: "B", Text: "Dua"}},
		Status:        db.CbtQuestionStatusEnumDraft,
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(matching draft empty key) error = %v", err)
	}
	if matchingDraft.AnswerKey != "" {
		t.Fatalf("normalizeQuestionInput(matching draft empty key) answer = %q, want empty draft key", matchingDraft.AnswerKey)
	}

	trueFalse, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "advance",
		QuestionType:  "true_false",
		QuestionText:  "MTsN 2 berada di Kolaka Utara",
		AnswerKey:     "A",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(true_false) error = %v", err)
	}
	if len(trueFalse.Options) != 2 || trueFalse.Options[0].Text != "Benar" || trueFalse.Options[1].Text != "Salah" {
		t.Fatalf("normalizeQuestionInput(true_false) options = %+v, want Benar/Salah", trueFalse.Options)
	}

	agreeDisagree, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "beginner",
		QuestionType:  "agree_disagree",
		QuestionText:  "Kebersihan kelas adalah tanggung jawab bersama",
		AnswerKey:     "b",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(agree_disagree) error = %v", err)
	}
	if len(agreeDisagree.Options) != 2 || agreeDisagree.Options[0].Text != "Setuju" || agreeDisagree.Options[1].Text != "Tidak Setuju" || agreeDisagree.AnswerKey != "B" {
		t.Fatalf("normalizeQuestionInput(agree_disagree) options/key = %+v/%q, want Setuju/Tidak Setuju with key B", agreeDisagree.Options, agreeDisagree.AnswerKey)
	}

	matching, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "beginner",
		QuestionType:  "matching",
		QuestionText:  "Jodohkan istilah dengan pengertiannya",
		Options: []QuestionOption{
			{Label: "A", HTML: "<p>Fotosintesis</p>", MatchLabel: "1", MatchHTML: "<p>Proses membuat makanan</p>"},
			{Label: "B", Text: "Evaporasi", MatchLabel: "2", MatchText: "Penguapan"},
			{MatchLabel: "3", MatchText: "Distraktor kanan", IsDistractor: true},
		},
		AnswerKey:      "b=2; a=1",
		WorkflowStatus: "review",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(matching) error = %v", err)
	}
	if matching.AnswerKey != "A=1;B=2" || len(matching.Options) != 3 || matching.Options[0].MatchText != "" || matching.Options[0].MatchHTML != "<p>Proses membuat makanan</p>" || !matching.Options[2].IsDistractor {
		t.Fatalf("normalizeQuestionInput(matching) key/options = %q/%+v, want canonical matching options", matching.AnswerKey, matching.Options)
	}

	fromStem, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "advance",
		QuestionType:  "essay",
		StemHTML:      "<p>Jelaskan&nbsp;EDM</p>",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(stem) error = %v", err)
	}
	if fromStem.QuestionText != "Jelaskan EDM" {
		t.Fatalf("normalizeQuestionInput(stem) QuestionText = %q, want derived plain text", fromStem.QuestionText)
	}

	if got := normalizeQuestionType(" SINGLE_CHOICE "); got != "multiple_choice" {
		t.Fatalf("normalizeQuestionType(single_choice) = %q, want multiple_choice", got)
	}
	if got := normalizeQuestionType(" TRUE_FALSE "); got != "true_false" {
		t.Fatalf("normalizeQuestionType(true_false) = %q, want true_false", got)
	}
	if got := normalizeQuestionType(" AGREE_DISAGREE "); got != "agree_disagree" {
		t.Fatalf("normalizeQuestionType(agree_disagree) = %q, want agree_disagree", got)
	}
	if got := normalizeQuestionType("matching"); got != "matching" {
		t.Fatalf("normalizeQuestionType(matching) = %q, want matching", got)
	}
	if got := normalizeQuestionType("ordering"); got != "ordering" {
		t.Fatalf("normalizeQuestionType(unknown) = %q, want ordering passthrough", got)
	}
	if got := normalizeAuthoringMode("ADVANCE"); got != "advance" {
		t.Fatalf("normalizeAuthoringMode() = %q, want advance", got)
	}
	if got := normalizeWorkflowStatus("approved"); got != "approved" {
		t.Fatalf("normalizeWorkflowStatus(approved) = %q, want approved", got)
	}
	if got := normalizeWorkflowStatus("published"); got != "published" {
		t.Fatalf("normalizeWorkflowStatus(published) = %q, want published", got)
	}

	sixOptions, err := normalizeQuestionInput(SaveCbtQuestionInput{
		SubjectID:     pgtype.UUID{Valid: true},
		AuthoringMode: "beginner",
		QuestionType:  "multiple_choice",
		QuestionText:  "Pilih jawaban",
		Options: []QuestionOption{
			{Label: "A", Text: "A"},
			{Label: "B", Text: "B"},
			{Label: "C", Text: "C"},
			{Label: "D", Text: "D"},
			{Label: "E", Text: "E"},
			{Label: "F", Text: "F"},
		},
		AnswerKey: "f",
	})
	if err != nil {
		t.Fatalf("normalizeQuestionInput(six options) error = %v", err)
	}
	if len(sixOptions.Options) != 6 || sixOptions.AnswerKey != "F" {
		t.Fatalf("normalizeQuestionInput(six options) = %d/%q, want 6/F", len(sixOptions.Options), sixOptions.AnswerKey)
	}

	a, b, c, d, e := legacyOptionColumns([]QuestionOption{
		{Text: "teks"},
		{HTML: "<b>html</b>"},
		{Latex: "x^2"},
	})
	if a != "teks" || b != "html" || c != "x^2" || d != "" || e != "" {
		t.Fatalf("legacyOptionColumns() = %q/%q/%q/%q/%q, want text/html/latex/empty/empty", a, b, c, d, e)
	}

	optionsJSON, err := EncodeQuestionOptions([]QuestionOption{{Label: "A", Text: "Satu"}})
	if err != nil {
		t.Fatalf("EncodeQuestionOptions() error = %v", err)
	}
	var decodedOptions []QuestionOption
	if err := json.Unmarshal(optionsJSON, &decodedOptions); err != nil || len(decodedOptions) != 1 || decodedOptions[0].Label != "A" {
		t.Fatalf("EncodeQuestionOptions() json = %s decoded=%+v err=%v, want one option", string(optionsJSON), decodedOptions, err)
	}
	emptyOptionsJSON, err := EncodeQuestionOptions(nil)
	if string(mustBytes(t, emptyOptionsJSON, err)) != "[]" {
		t.Fatalf("EncodeQuestionOptions(nil) = %s, want []", string(emptyOptionsJSON))
	}
	assetJSON, err := EncodeStringArray([]string{"asset-1"})
	if string(mustBytes(t, assetJSON, err)) != `["asset-1"]` {
		t.Fatalf("EncodeStringArray() = %s, want asset JSON", string(assetJSON))
	}
	if got := decodeQuestionOptions([]byte(`[{"label":"B","text":"Dua"}]`)); len(got) != 1 || got[0].Label != "B" {
		t.Fatalf("decodeQuestionOptions(valid) = %+v, want one B option", got)
	}
	if got := decodeQuestionOptions([]byte(`bad`)); got != nil {
		t.Fatalf("decodeQuestionOptions(invalid) = %+v, want nil", got)
	}
	if got := decodeStringArray([]byte(`["a","b"]`)); strings.Join(got, ",") != "a,b" {
		t.Fatalf("decodeStringArray(valid) = %+v, want a,b", got)
	}
	if got := decodeStringArray([]byte(`bad`)); got != nil {
		t.Fatalf("decodeStringArray(invalid) = %+v, want nil", got)
	}
	if got := mergeNotes("lama", " baru "); got != "baru" {
		t.Fatalf("mergeNotes() = %q, want incoming note only", got)
	}
	if got := mergeNotes("lama", " "); got != "lama" {
		t.Fatalf("mergeNotes(empty incoming) = %q, want existing", got)
	}
}

func TestCbtFixedPairScoringSqlContract(t *testing.T) {
	sqlBytes, err := os.ReadFile("../../db/queries/cbt_sessions.sql")
	if err != nil {
		t.Fatalf("Read cbt_sessions.sql error = %v", err)
	}
	sqlText := string(sqlBytes)

	for _, phrase := range []string{
		"WHEN lower(btrim(sa.answer)) = 'true' THEN 'A'",
		"WHEN lower(btrim(sa.answer)) = 'false' THEN 'B'",
	} {
		if !strings.Contains(sqlText, phrase) {
			t.Fatalf("cbt_sessions.sql missing fixed-pair scoring phrase %q", phrase)
		}
	}
	for _, alternatives := range [][]string{
		{"WHEN q.question_type = 'true_false' THEN", "WHEN items.question_type = 'true_false' THEN"},
		{"WHEN q.question_type = 'agree_disagree' THEN", "WHEN items.question_type = 'agree_disagree' THEN"},
		{"upper(btrim(sa.answer)) = upper(btrim(q.answer_key))", "upper(btrim(sa.answer)) = upper(btrim(items.answer_key))"},
	} {
		found := false
		for _, phrase := range alternatives {
			if strings.Contains(sqlText, phrase) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("cbt_sessions.sql missing fixed-pair scoring phrase alternatives %q", alternatives)
		}
	}
}

func TestCbtQuestionSuggestAuthoringMode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		hots bool
		want string
	}{
		{name: "beginner", args: []string{"multiple_choice", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "beginner"},
		{name: "type beginner", args: []string{"short_answer", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "beginner"},
		{name: "latex advance", args: []string{"essay", "", " y ", "", "", "", "", "", "", "", "draft", "", "", ""}, want: "advance"},
		{name: "metadata advance", args: []string{"essay", "", "", "", "", "TP", "", "", "", "", "draft", "", "", ""}, want: "advance"},
		{name: "hots advance", args: []string{"essay", "", "", "", "", "", "", "", "", "", "draft", "", "", ""}, hots: true, want: "advance"},
		{name: "workflow advance", args: []string{"essay", "", "", "", "", "", "", "", "", "", "review", "", "", ""}, want: "advance"},
		{name: "notes advance", args: []string{"essay", "", "", "", "", "", "", "", "", "", "draft", "", "review", ""}, want: "advance"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestQuestionAuthoringMode(
				tt.args[0],
				tt.args[1],
				tt.args[2],
				tt.args[3],
				tt.args[4],
				tt.args[5],
				tt.args[6],
				tt.args[7],
				tt.args[8],
				tt.args[9],
				tt.hots,
				tt.args[10],
				tt.args[11],
				tt.args[12],
				tt.args[13],
			)
			if got != tt.want {
				t.Fatalf("suggestQuestionAuthoringMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func mustBytes(t *testing.T, value []byte, err error) []byte {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	return value
}

func TestNormalizeQuestionInputQuestionTypeContracts(t *testing.T) {
	base := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionText:   "Kontrak tipe soal harus deterministik",
		WorkflowStatus: "review",
		Options: []QuestionOption{
			{Label: "A", Text: "Alpha"},
			{Label: "B", Text: "Beta"},
			{Label: "C", Text: "Gamma"},
			{Label: "D", Text: "Delta"},
		},
	}

	t.Run("multiple answer sorts labels deterministically", func(t *testing.T) {
		input := base
		input.QuestionType = "multiple_answer"
		input.AnswerKey = " c, a "
		got, err := normalizeQuestionInput(input)
		if err != nil {
			t.Fatalf("normalizeQuestionInput() error = %v", err)
		}
		if got.AnswerKey != "A,C" {
			t.Fatalf("AnswerKey = %q, want A,C", got.AnswerKey)
		}
	})

	t.Run("ordering preserves exact sequence and requires every label once", func(t *testing.T) {
		input := base
		input.QuestionType = "ordering"
		input.AnswerKey = " B, A, D, C "
		got, err := normalizeQuestionInput(input)
		if err != nil {
			t.Fatalf("normalizeQuestionInput() error = %v", err)
		}
		if got.AnswerKey != "B,A,D,C" {
			t.Fatalf("AnswerKey = %q, want B,A,D,C", got.AnswerKey)
		}

		input.AnswerKey = "B,A,A,C"
		if _, err := normalizeQuestionInput(input); err == nil {
			t.Fatalf("normalizeQuestionInput() duplicate ordering labels succeeded, want error")
		}
	})

	t.Run("fixed pairs keep canonical labels and options", func(t *testing.T) {
		input := base
		input.Options = nil
		input.QuestionType = "agree_disagree"
		input.AnswerKey = " b "
		got, err := normalizeQuestionInput(input)
		if err != nil {
			t.Fatalf("normalizeQuestionInput() error = %v", err)
		}
		if got.AnswerKey != "B" || len(got.Options) != 2 || got.Options[0].Text != "Setuju" || got.Options[1].Text != "Tidak Setuju" {
			t.Fatalf("fixed pair contract = key %q options %+v", got.AnswerKey, got.Options)
		}
	})

	t.Run("short answer aliases are normalized", func(t *testing.T) {
		input := base
		input.QuestionType = "short_answer"
		input.Options = nil
		input.AnswerKey = " Fotosintesis | foto  sintesis | fotosintesis "
		got, err := normalizeQuestionInput(input)
		if err != nil {
			t.Fatalf("normalizeQuestionInput() error = %v", err)
		}
		if got.AnswerKey != "Fotosintesis|foto sintesis" {
			t.Fatalf("AnswerKey = %q, want normalized aliases", got.AnswerKey)
		}
	})
}
