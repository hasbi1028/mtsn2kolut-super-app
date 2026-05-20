package db

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationCbtQuestionWorkflowStatusAndAuditQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	fixture := setupCbtQuestionWorkflowFixture(t, ctx, tdb, "WF")

	question := createCbtWorkflowQuestion(t, ctx, q, cbtWorkflowQuestionSeed{
		EventID:        fixture.eventID,
		SubjectID:      fixture.subjectID,
		Code:           "CBT-WF-Q-" + fixture.suffix,
		QuestionText:   "Draft workflow question " + fixture.suffix,
		AuthorUsername: "cbt_" + fixture.suffix,
		Status:         CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		VersionNote:    "draft",
	})

	updated, err := q.UpdateCbtQuestion(ctx, updateParamsFromCbtQuestion(question, cbtQuestionWorkflowUpdate{
		WorkflowStatus:   "submitted",
		Status:           CbtQuestionStatusEnumDraft,
		WriterNotes:      "submitted from integration test",
		ReviewerUsername: "reviewer_" + fixture.suffix,
		ReviewNotes:      "waiting for review",
	}))
	if err != nil {
		t.Fatalf("update cbt question workflow status: %v", err)
	}
	if updated.WorkflowStatus != "submitted" || updated.Version != 2 || updated.WriterNotes != "submitted from integration test" {
		t.Fatalf("updated workflow question mismatch: %#v", updated)
	}

	audit, err := q.CreateCbtQuestionAuditLog(ctx, CreateCbtQuestionAuditLogParams{
		QuestionID:    question.ID,
		ActorUsername: "cbt_" + fixture.suffix,
		Action:        "submit_review",
		Note:          "submitted for review",
		Metadata:      []byte(`{"source":"integration","step":"audit"}`),
	})
	if err != nil {
		t.Fatalf("create cbt question audit log: %v", err)
	}
	if audit.QuestionID != question.ID || audit.Action != "submit_review" || !strings.Contains(string(audit.Metadata), "integration") {
		t.Fatalf("audit log mismatch: %#v", audit)
	}

	workflowEvent, err := q.CreateBankSoalQuestionWorkflowEvent(ctx, CreateBankSoalQuestionWorkflowEventParams{
		QuestionID:    question.ID,
		ActorUserID:   fixture.userID,
		ActorUsername: "cbt_" + fixture.suffix,
		FromStatus:    "draft",
		ToStatus:      "submitted",
		Action:        "submit",
		Note:          "workflow event from integration test",
		Metadata:      []byte(`{"source":"integration","step":"workflow_event"}`),
	})
	if err != nil {
		t.Fatalf("create bank soal question workflow event: %v", err)
	}
	if workflowEvent.QuestionID != question.ID || workflowEvent.ActorUserID != fixture.userID || workflowEvent.ToStatus != "submitted" {
		t.Fatalf("workflow event mismatch: %#v", workflowEvent)
	}

	filtered, err := q.ListCbtQuestionsFiltered(ctx, ListCbtQuestionsFilteredParams{
		IsAdmin:          false,
		ActorUsername:    "cbt_" + fixture.suffix,
		ActorUserID:      fixture.userID,
		ScopeFilter:      "event_pool",
		EventID:          fixture.eventID,
		SubjectID:        fixture.subjectID,
		WorkflowStatuses: []string{"submitted"},
		SortOrder:        "oldest",
		LimitCount:       10,
	})
	if err != nil {
		t.Fatalf("list cbt questions filtered: %v", err)
	}
	if got := findFilteredCbtQuestion(filtered, question.ID); got == nil {
		t.Fatalf("submitted question %v not found in filtered results: %#v", question.ID, filtered)
	} else if got.WorkflowStatus != "submitted" || got.AnswerKey != "A" || got.PackageCount != 0 {
		t.Fatalf("filtered submitted question mismatch: %#v", *got)
	}

	count, err := q.CountCbtQuestionsFiltered(ctx, CountCbtQuestionsFilteredParams{
		ScopeFilter:      "event_pool",
		EventID:          fixture.eventID,
		SubjectID:        fixture.subjectID,
		WorkflowStatuses: []string{"submitted"},
		ActorUsername:    "cbt_" + fixture.suffix,
		ActorUserID:      fixture.userID,
	})
	if err != nil {
		t.Fatalf("count cbt questions filtered: %v", err)
	}
	if count != 1 {
		t.Fatalf("submitted question count = %d, want 1", count)
	}
}

func TestIntegrationCbtPackageQuestionOrderingCloneAndSnapshotQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	fixture := setupCbtQuestionWorkflowFixture(t, ctx, tdb, "PQ")

	first := createCbtWorkflowQuestion(t, ctx, q, cbtWorkflowQuestionSeed{
		EventID:        fixture.eventID,
		SubjectID:      fixture.subjectID,
		Code:           "CBT-PQ-A-" + fixture.suffix,
		QuestionText:   "First displayed package question " + fixture.suffix,
		AuthorUsername: "cbt_" + fixture.suffix,
		Status:         CbtQuestionStatusEnumPublished,
		WorkflowStatus: "approved",
		VersionNote:    "first",
	})
	second := createCbtWorkflowQuestion(t, ctx, q, cbtWorkflowQuestionSeed{
		EventID:        fixture.eventID,
		SubjectID:      fixture.subjectID,
		Code:           "CBT-PQ-B-" + fixture.suffix,
		QuestionText:   "Second displayed package question " + fixture.suffix,
		AuthorUsername: "cbt_" + fixture.suffix,
		Status:         CbtQuestionStatusEnumPublished,
		WorkflowStatus: "published",
		VersionNote:    "second",
	})

	pkg, err := q.CreateCbtPackage(ctx, CreateCbtPackageParams{
		EventID:            fixture.eventID,
		SubjectID:          fixture.subjectID,
		Title:              "CBT Ordered Package " + fixture.suffix,
		Description:        "package ordering integration test",
		DurationMinutes:    60,
		RandomizeQuestions: false,
		IsActive:           true,
		SourceMode:         "teacher_class",
		RandomizeOptions:   false,
		DrawPgCount:        2,
		DrawEssayCount:     0,
		RandomSeed:         "seed-" + fixture.suffix,
		CompositionLog:     []byte(`{"source":"integration"}`),
	})
	if err != nil {
		t.Fatalf("create cbt package: %v", err)
	}
	if rows, err := q.AddCbtPackageQuestion(ctx, AddCbtPackageQuestionParams{PackageID: pkg.ID, QuestionID: second.ID, Position: 2, Points: 3}); err != nil || rows != 1 {
		t.Fatalf("add second package question rows=%d err=%v", rows, err)
	}
	if rows, err := q.AddCbtPackageQuestion(ctx, AddCbtPackageQuestionParams{PackageID: pkg.ID, QuestionID: first.ID, Position: 1, Points: 2}); err != nil || rows != 1 {
		t.Fatalf("add first package question rows=%d err=%v", rows, err)
	}

	packageQuestions, err := q.ListCbtPackageQuestionsByPackage(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("list cbt package questions by package: %v", err)
	}
	if len(packageQuestions) != 2 || packageQuestions[0].QuestionID != first.ID || packageQuestions[0].Position != 1 || packageQuestions[0].Points != 2 || packageQuestions[1].QuestionID != second.ID || packageQuestions[1].Position != 2 || packageQuestions[1].Points != 3 {
		t.Fatalf("package question order mismatch: %#v", packageQuestions)
	}

	examQuestions, err := q.GetExamQuestions(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("get exam questions before snapshot: %v", err)
	}
	if len(examQuestions) != 2 || examQuestions[0].ID != first.ID || examQuestions[1].ID != second.ID {
		t.Fatalf("exam question order mismatch before snapshot: %#v", examQuestions)
	}

	cloned, err := q.CloneCbtPackage(ctx, CloneCbtPackageParams{Title: "CBT Ordered Package Clone " + fixture.suffix, SourceID: pkg.ID})
	if err != nil {
		t.Fatalf("clone cbt package: %v", err)
	}
	if rows, err := q.CloneCbtPackageQuestions(ctx, CloneCbtPackageQuestionsParams{TargetID: cloned.ID, SourceID: pkg.ID}); err != nil || rows != 2 {
		t.Fatalf("clone cbt package questions rows=%d err=%v", rows, err)
	}
	clonedQuestions, err := q.ListCbtPackageQuestionsByPackage(ctx, cloned.ID)
	if err != nil {
		t.Fatalf("list cloned cbt package questions: %v", err)
	}
	if len(clonedQuestions) != 2 || clonedQuestions[0].QuestionID != first.ID || clonedQuestions[0].Position != 1 || clonedQuestions[1].QuestionID != second.ID || clonedQuestions[1].Position != 2 {
		t.Fatalf("cloned package question order mismatch: %#v", clonedQuestions)
	}

	lock, err := q.LockCbtPackageForSnapshot(ctx, LockCbtPackageForSnapshotParams{LockedBy: fixture.userID, LockReason: "integration_snapshot", PackageID: pkg.ID})
	if err != nil {
		t.Fatalf("lock cbt package for snapshot: %v", err)
	}
	if !lock.LockedAt.Valid || lock.LockedBy != fixture.userID || lock.SnapshotVersion != 1 || lock.LockReason != "integration_snapshot" {
		t.Fatalf("package lock mismatch: %#v", lock)
	}
	if rows, err := q.CreateCbtPackageQuestionSnapshots(ctx, pkg.ID); err != nil || rows != 2 {
		t.Fatalf("create cbt package question snapshots rows=%d err=%v", rows, err)
	}
	if rows, err := q.AddCbtPackageQuestion(ctx, AddCbtPackageQuestionParams{PackageID: pkg.ID, QuestionID: first.ID, Position: 3, Points: 1}); err != nil || rows != 0 {
		t.Fatalf("add package question to locked package rows=%d err=%v", rows, err)
	}

	if _, err := q.UpdateCbtQuestion(ctx, updateParamsFromCbtQuestion(first, cbtQuestionWorkflowUpdate{
		QuestionText:     "Mutated after snapshot " + fixture.suffix,
		WorkflowStatus:   "approved",
		Status:           CbtQuestionStatusEnumPublished,
		ApproverUsername: "approver_" + fixture.suffix,
		ReviewNotes:      "mutated after snapshot",
	})); err != nil {
		t.Fatalf("update question after snapshot: %v", err)
	}
	snapshotExamQuestions, err := q.GetExamQuestions(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("get exam questions after snapshot: %v", err)
	}
	if len(snapshotExamQuestions) != 2 || snapshotExamQuestions[0].ID != first.ID || snapshotExamQuestions[0].QuestionText != first.QuestionText || snapshotExamQuestions[1].ID != second.ID {
		t.Fatalf("snapshot exam question order/content mismatch: %#v", snapshotExamQuestions)
	}
}

type cbtQuestionWorkflowFixture struct {
	suffix    string
	subjectID pgtype.UUID
	eventID   pgtype.UUID
	userID    pgtype.UUID
}

func setupCbtQuestionWorkflowFixture(t *testing.T, ctx context.Context, tdb *integrationTestDB, prefix string) cbtQuestionWorkflowFixture {
	t.Helper()

	q := tdb.Q
	suffix := prefix + "-" + integrationSuffix()
	year, err := q.CreateAcademicYear(ctx, CreateAcademicYearParams{
		Name:      "CBT Question Workflow Year " + suffix,
		StartDate: pgDate(2026, time.July, 1),
		EndDate:   pgDate(2027, time.June, 30),
		IsActive:  false,
	})
	if err != nil {
		t.Fatalf("create academic year: %v", err)
	}
	subject, err := q.CreateSubject(ctx, CreateSubjectParams{
		Code:                "CBT-QW-SUB-" + suffix,
		Name:                "CBT Question Workflow Subject " + suffix,
		Category:            "umum",
		IsAssessmentSubject: true,
		IsReportSubject:     true,
		IsScheduleActivity:  false,
		CountsForRanking:    true,
		IsLocalContent:      false,
		IsChoiceSubject:     false,
		DefaultWeeklyHours:  2,
		DisplayOrder:        9999,
		IsActive:            true,
	})
	if err != nil {
		t.Fatalf("create subject: %v", err)
	}
	_, userID := seedCbtIntegrationEmployeeAndUser(t, ctx, tdb, suffix)
	event, err := q.CreateCbtExamEvent(ctx, CreateCbtExamEventParams{
		Title:          "CBT Question Workflow Event " + suffix,
		ExamType:       CbtExamTypeUts,
		Scope:          "grade",
		TargetLevels:   []string{"VII"},
		AcademicYearID: year.ID,
		Status:         "draft",
	})
	if err != nil {
		t.Fatalf("create cbt exam event: %v", err)
	}

	return cbtQuestionWorkflowFixture{suffix: suffix, subjectID: subject.ID, eventID: event.ID, userID: userID}
}

type cbtWorkflowQuestionSeed struct {
	EventID        pgtype.UUID
	SubjectID      pgtype.UUID
	Code           string
	QuestionText   string
	AuthorUsername string
	Status         CbtQuestionStatusEnum
	WorkflowStatus string
	VersionNote    string
}

func createCbtWorkflowQuestion(t *testing.T, ctx context.Context, q *Queries, seed cbtWorkflowQuestionSeed) CbtQuestion {
	t.Helper()
	question, err := q.CreateCbtQuestion(ctx, CreateCbtQuestionParams{
		EventID:              seed.EventID,
		SubjectID:            seed.SubjectID,
		Code:                 seed.Code,
		QuestionText:         seed.QuestionText,
		QuestionType:         "multiple_choice",
		Options:              []byte(`[{"key":"A","text":"4"},{"key":"B","text":"5"}]`),
		OptionA:              "4",
		OptionB:              "5",
		OptionC:              "6",
		OptionD:              "7",
		OptionE:              "",
		AnswerKey:            "A",
		Explanation:          "Basic arithmetic",
		Difficulty:           CbtQuestionDifficultyEnumEasy,
		Status:               seed.Status,
		StemHtml:             "<p>" + seed.QuestionText + "</p>",
		StemLatex:            "",
		StimulusHtml:         "",
		StimulusLatex:        "",
		ExplanationHtml:      "<p>2 + 2 = 4</p>",
		RubricHtml:           "<p>Rubric</p>",
		AcademicPhase:        "D",
		TargetLevel:          pgtype.Text{String: "VII", Valid: true},
		CpRef:                "CP-WF",
		TpRef:                "TP-WF",
		KdRef:                "KD-WF",
		IndicatorRef:         "IND-WF",
		MaterialTopic:        "Arithmetic workflow",
		CognitiveLevel:       "C1",
		HotsFlag:             false,
		MediaAssetIds:        []byte(`[]`),
		WorkflowStatus:       seed.WorkflowStatus,
		Version:              1,
		AuthorUsername:       seed.AuthorUsername,
		ReviewerUsername:     "",
		ReviewedAt:           pgtype.Timestamptz{},
		ApproverUsername:     "",
		ApprovedAt:           pgtype.Timestamptz{},
		WriterNotes:          "initial writer note",
		ReviewNotes:          "",
		VersionGroupID:       pgtype.UUID{},
		VersionNumber:        1,
		SourceQuestionID:     pgtype.UUID{},
		SupersedesQuestionID: pgtype.UUID{},
		IsLatestVersion:      true,
		VersionNote:          seed.VersionNote,
	})
	if err != nil {
		t.Fatalf("create cbt workflow question: %v", err)
	}
	return question
}

type cbtQuestionWorkflowUpdate struct {
	QuestionText     string
	Status           CbtQuestionStatusEnum
	WorkflowStatus   string
	ReviewerUsername string
	ReviewedAt       pgtype.Timestamptz
	ApproverUsername string
	ApprovedAt       pgtype.Timestamptz
	WriterNotes      string
	ReviewNotes      string
}

func updateParamsFromCbtQuestion(question CbtQuestion, update cbtQuestionWorkflowUpdate) UpdateCbtQuestionParams {
	questionText := question.QuestionText
	if update.QuestionText != "" {
		questionText = update.QuestionText
	}
	status := question.Status
	if update.Status != "" {
		status = update.Status
	}
	workflowStatus := question.WorkflowStatus
	if update.WorkflowStatus != "" {
		workflowStatus = update.WorkflowStatus
	}
	writerNotes := question.WriterNotes
	if update.WriterNotes != "" {
		writerNotes = update.WriterNotes
	}
	reviewNotes := question.ReviewNotes
	if update.ReviewNotes != "" {
		reviewNotes = update.ReviewNotes
	}

	return UpdateCbtQuestionParams{
		ID:               question.ID,
		EventID:          question.EventID,
		SubjectID:        question.SubjectID,
		Code:             question.Code,
		QuestionText:     questionText,
		QuestionType:     question.QuestionType,
		Options:          question.Options,
		OptionA:          question.OptionA,
		OptionB:          question.OptionB,
		OptionC:          question.OptionC,
		OptionD:          question.OptionD,
		OptionE:          question.OptionE,
		AnswerKey:        question.AnswerKey,
		Explanation:      question.Explanation,
		Difficulty:       question.Difficulty,
		Status:           status,
		StemHtml:         question.StemHtml,
		StemLatex:        question.StemLatex,
		StimulusHtml:     question.StimulusHtml,
		StimulusLatex:    question.StimulusLatex,
		ExplanationHtml:  question.ExplanationHtml,
		RubricHtml:       question.RubricHtml,
		AcademicPhase:    question.AcademicPhase,
		TargetLevel:      question.TargetLevel,
		CpRef:            question.CpRef,
		TpRef:            question.TpRef,
		KdRef:            question.KdRef,
		IndicatorRef:     question.IndicatorRef,
		MaterialTopic:    question.MaterialTopic,
		CognitiveLevel:   question.CognitiveLevel,
		HotsFlag:         question.HotsFlag,
		MediaAssetIds:    question.MediaAssetIds,
		WorkflowStatus:   workflowStatus,
		ReviewerUsername: update.ReviewerUsername,
		ReviewedAt:       update.ReviewedAt,
		ApproverUsername: update.ApproverUsername,
		ApprovedAt:       update.ApprovedAt,
		WriterNotes:      writerNotes,
		ReviewNotes:      reviewNotes,
	}
}

func findFilteredCbtQuestion(questions []ListCbtQuestionsFilteredRow, id pgtype.UUID) *ListCbtQuestionsFilteredRow {
	for i := range questions {
		if questions[i].ID == id {
			return &questions[i]
		}
	}
	return nil
}
