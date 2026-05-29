package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func megaQuestionBase(id pgtype.UUID, workflow string) db.GetCbtQuestionRow {
	return db.GetCbtQuestionRow{
		ID:             id,
		SubjectID:      pgtype.UUID{Bytes: [16]byte{42}, Valid: true},
		Code:           "MEGA-1",
		QuestionText:   "Pilih jawaban benar",
		QuestionType:   "multiple_choice",
		OptionA:        "Benar",
		OptionB:        "Salah",
		AnswerKey:      "A",
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: workflow,
		TargetLevel:    pgtype.Text{String: "VII", Valid: true},
		AuthorUsername: "author",
		VersionNumber:  1,
	}
}

func TestCbtQuestionMegaWorkflowTransitionsWriteAuditAndEvents(t *testing.T) {
	ctx := context.Background()
	questionID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000030001")
	reviewerID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000030002")
	approverID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000030003")

	t.Run("submit for review updates workflow and records audit/event", func(t *testing.T) {
		store := &fakeQuestionStore{current: megaQuestionBase(questionID, "draft")}
		svc := &CbtQuestion{q: store}

		row, err := svc.SubmitForReview(ctx, questionID, CbtQuestionActor{Username: "author"}, "ready")
		if err != nil {
			t.Fatalf("SubmitForReview() error = %v", err)
		}
		if row.WorkflowStatus != "diperiksa" || store.updateParams.WorkflowStatus != "diperiksa" || store.updateParams.ReviewerUsername != "" {
			t.Fatalf("SubmitForReview() workflow/reviewer = row %q params %q reviewer %q, want diperiksa with blank reviewer", row.WorkflowStatus, store.updateParams.WorkflowStatus, store.updateParams.ReviewerUsername)
		}
		if store.auditCalls != 1 || len(store.workflowEvents) != 1 {
			t.Fatalf("SubmitForReview() audit/events = %d/%d, want 1/1", store.auditCalls, len(store.workflowEvents))
		}
		if store.auditLogs[0].Action != "submit_for_review" || store.auditLogs[0].ActorUsername != "author" || store.auditLogs[0].Note != "ready" {
			t.Fatalf("SubmitForReview() audit = %+v", store.auditLogs[0])
		}
		event := store.workflowEvents[0]
		if event.FromStatus != "draft" || event.ToStatus != "diperiksa" || event.Action != "submit_for_review" {
			t.Fatalf("SubmitForReview() event = %+v, want draft->diperiksa submit_for_review", event)
		}
		var meta map[string]any
		if err := json.Unmarshal(event.Metadata, &meta); err != nil {
			t.Fatalf("workflow metadata is not JSON: %v", err)
		}
		if meta["workflow_status"] != "diperiksa" || meta["to_publication_status"] != string(db.CbtQuestionStatusEnumDraft) {
			t.Fatalf("SubmitForReview() metadata = %#v, want workflow/publication status", meta)
		}
	})

	t.Run("revision needed latest draft can be resubmitted after edit", func(t *testing.T) {
		current := megaQuestionBase(questionID, "revision_needed")
		current.IsLatestVersion = true
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		row, err := svc.SubmitForReview(ctx, questionID, CbtQuestionActor{Username: "author"}, "revisi selesai")
		if err != nil {
			t.Fatalf("SubmitForReview(revision_needed safe) error = %v", err)
		}
		if row.WorkflowStatus != "diperiksa" || store.updateParams.WorkflowStatus != "diperiksa" {
			t.Fatalf("SubmitForReview(revision_needed safe) row/update = %+v/%+v, want diperiksa", row, store.updateParams)
		}
		if len(store.workflowEvents) != 1 || store.workflowEvents[0].FromStatus != "revision_needed" || store.workflowEvents[0].ToStatus != "diperiksa" {
			t.Fatalf("SubmitForReview(revision_needed safe) events = %+v, want revision_needed -> diperiksa", store.workflowEvents)
		}
	})

	t.Run("already submitted is idempotent and does not audit", func(t *testing.T) {
		current := megaQuestionBase(questionID, "submitted")
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}

		row, err := svc.SubmitForReview(ctx, questionID, CbtQuestionActor{Username: "author"}, "again")
		if err != nil {
			t.Fatalf("SubmitForReview(already submitted) error = %v", err)
		}
		if row.WorkflowStatus != "submitted" || store.updateCalls != 0 || store.auditCalls != 0 || len(store.workflowEvents) != 0 {
			t.Fatalf("SubmitForReview(already submitted) row/update/audit/events = %+v/%d/%d/%d, want no mutation", row, store.updateCalls, store.auditCalls, len(store.workflowEvents))
		}
	})

	t.Run("reviewer marks submitted question reviewed", func(t *testing.T) {
		store := &fakeQuestionStore{current: megaQuestionBase(questionID, "submitted"), canReview: true}
		svc := &CbtQuestion{q: store}
		actor := CbtQuestionActor{UserID: reviewerID, Username: "reviewer", Permissions: []string{"bank_soal.review"}}

		row, err := svc.MarkReviewed(ctx, questionID, actor, "layak")
		if err != nil {
			t.Fatalf("MarkReviewed() error = %v", err)
		}
		if row.WorkflowStatus != "diperiksa" || row.ReviewerUsername != "reviewer" || store.updateParams.ApproverUsername != "" {
			t.Fatalf("MarkReviewed() row/update = %+v / %+v", row, store.updateParams)
		}
		if store.auditLogs[0].Action != "mark_reviewed" || store.workflowEvents[0].FromStatus != "submitted" || store.workflowEvents[0].ToStatus != "diperiksa" {
			t.Fatalf("MarkReviewed() audit/event = %+v / %+v", store.auditLogs[0], store.workflowEvents[0])
		}
	})

	t.Run("publisher publishes approved question and status", func(t *testing.T) {
		current := megaQuestionBase(questionID, "siap_pakai")
		current.ApproverUsername = "lead"
		store := &fakeQuestionStore{current: current, canApprove: true}
		svc := &CbtQuestion{q: store}
		actor := CbtQuestionActor{UserID: approverID, Username: "publisher", Permissions: []string{"bank_soal.publish"}}

		row, err := svc.Publish(ctx, questionID, actor)
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
		if row.WorkflowStatus != "siap_pakai" || row.Status != db.CbtQuestionStatusEnumPublished || row.ApproverUsername != "publisher" {
			t.Fatalf("Publish() row = %+v, want published status and publisher approver", row)
		}
		if store.updateParams.Status != db.CbtQuestionStatusEnumPublished || store.updateParams.WorkflowStatus != "siap_pakai" {
			t.Fatalf("Publish() update params = %+v, want published/siap_pakai", store.updateParams)
		}
		var meta map[string]any
		if err := json.Unmarshal(store.workflowEvents[0].Metadata, &meta); err != nil {
			t.Fatalf("Publish() workflow metadata is not JSON: %v", err)
		}
		if meta["from_approver_username"] != "lead" || meta["to_approver_username"] != "publisher" || meta["status"] != string(db.CbtQuestionStatusEnumPublished) {
			t.Fatalf("Publish() metadata = %#v, want approver/status transition", meta)
		}
	})
}

func TestCbtQuestionMegaWorkflowAndMutationGuards(t *testing.T) {
	ctx := context.Background()
	questionID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000031001")

	t.Run("reviewer cannot review own question", func(t *testing.T) {
		store := &fakeQuestionStore{current: megaQuestionBase(questionID, "submitted"), canReview: true}
		svc := &CbtQuestion{q: store}
		_, err := svc.MarkReviewed(ctx, questionID, CbtQuestionActor{UserID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, Username: "author", Permissions: []string{"bank_soal.review"}}, "self")
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("MarkReviewed(self) error = %v, want ErrForbidden", err)
		}
		if store.updateCalls != 0 || store.auditCalls != 0 {
			t.Fatalf("MarkReviewed(self) update/audit = %d/%d, want no mutation", store.updateCalls, store.auditCalls)
		}
	})

	t.Run("publish requires approved workflow", func(t *testing.T) {
		store := &fakeQuestionStore{current: megaQuestionBase(questionID, "reviewed"), canApprove: true}
		svc := &CbtQuestion{q: store}
		_, err := svc.Publish(ctx, questionID, CbtQuestionActor{Username: "admin", Roles: []string{"admin"}})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("Publish(reviewed) error = %v, want ErrConflict", err)
		}
		if store.updateCalls != 0 {
			t.Fatalf("Publish(reviewed) updateCalls = %d, want 0", store.updateCalls)
		}
	})

	t.Run("update locked question fails before mutation", func(t *testing.T) {
		current := megaQuestionBase(questionID, "draft")
		current.PackageCount = 1
		store := &fakeQuestionStore{current: current}
		svc := &CbtQuestion{q: store}
		_, err := svc.Update(ctx, SaveCbtQuestionInput{
			ID:             questionID,
			SubjectID:      current.SubjectID,
			QuestionText:   current.QuestionText,
			QuestionType:   current.QuestionType,
			OptionA:        current.OptionA,
			OptionB:        current.OptionB,
			AnswerKey:      current.AnswerKey,
			WorkflowStatus: "draft",
			AuthorUsername: "author",
			Actor:          CbtQuestionActor{Username: "author"},
		})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("Update(locked) error = %v, want ErrConflict", err)
		}
		if store.updateCalls != 0 || store.auditCalls != 0 {
			t.Fatalf("Update(locked) update/audit = %d/%d, want no mutation", store.updateCalls, store.auditCalls)
		}
	})

	t.Run("delete rejects non draft workflow", func(t *testing.T) {
		store := &fakeQuestionStore{current: megaQuestionBase(questionID, "submitted")}
		svc := &CbtQuestion{q: store}
		err := svc.DeleteWithActor(ctx, questionID, CbtQuestionActor{Username: "admin", Roles: []string{"admin"}})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("DeleteWithActor(submitted) error = %v, want ErrConflict", err)
		}
		if store.deleteCalls != 0 || store.auditCalls != 0 {
			t.Fatalf("DeleteWithActor(submitted) delete/audit = %d/%d, want no mutation", store.deleteCalls, store.auditCalls)
		}
	})

	t.Run("delete draft records audit before delete", func(t *testing.T) {
		store := &fakeQuestionStore{current: megaQuestionBase(questionID, "draft")}
		svc := &CbtQuestion{q: store}
		err := svc.DeleteWithActor(ctx, questionID, CbtQuestionActor{Username: "author"})
		if err != nil {
			t.Fatalf("DeleteWithActor(draft) error = %v", err)
		}
		if store.deleteCalls != 1 || store.deleteID != questionID || store.auditCalls != 1 {
			t.Fatalf("DeleteWithActor(draft) delete/audit = %d/%v/%d, want one delete id and one audit", store.deleteCalls, store.deleteID, store.auditCalls)
		}
		if store.auditLogs[0].Action != "delete" || store.auditLogs[0].ActorUsername != "author" {
			t.Fatalf("DeleteWithActor(draft) audit = %+v, want author delete", store.auditLogs[0])
		}
	})
}

func TestCbtQuestionMegaVisibilityAndValidationBranches(t *testing.T) {
	ctx := context.Background()
	questionID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000032001")

	t.Run("published question redacts answer material for public actor", func(t *testing.T) {
		store := &fakeQuestionStore{detail: db.GetCbtQuestionDetailRow{ID: questionID, Status: db.CbtQuestionStatusEnumPublished, WorkflowStatus: "published", AuthorUsername: "author", AnswerKey: "A", RubricHtml: "<p>rubrik</p>"}}
		svc := &CbtQuestion{q: store}
		row, err := svc.GetDetail(ctx, questionID, CbtQuestionActor{Username: "reader"})
		if err != nil {
			t.Fatalf("GetDetail(public published) error = %v", err)
		}
		if row.AnswerKey != "" || row.RubricHtml != "" {
			t.Fatalf("GetDetail(public published) answer/rubric = %q/%q, want redacted", row.AnswerKey, row.RubricHtml)
		}
	})

	t.Run("unpublished question is hidden from unrelated actor", func(t *testing.T) {
		store := &fakeQuestionStore{detail: db.GetCbtQuestionDetailRow{ID: questionID, Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "draft", AuthorUsername: "author", AnswerKey: "A"}}
		svc := &CbtQuestion{q: store}
		_, err := svc.GetDetail(ctx, questionID, CbtQuestionActor{Username: "reader"})
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("GetDetail(unpublished unrelated) error = %v, want ErrForbidden", err)
		}
	})

	t.Run("scoped reviewer can see submitted answer material", func(t *testing.T) {
		store := &fakeQuestionStore{detail: db.GetCbtQuestionDetailRow{ID: questionID, Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "submitted", AuthorUsername: "author", AnswerKey: "A", RubricHtml: "rubric"}, canReview: true}
		svc := &CbtQuestion{q: store}
		row, err := svc.GetDetail(ctx, questionID, CbtQuestionActor{UserID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true}, Username: "reviewer", Permissions: []string{"bank_soal.review"}})
		if err != nil {
			t.Fatalf("GetDetail(scoped reviewer) error = %v", err)
		}
		if row.AnswerKey != "A" || row.RubricHtml != "rubric" {
			t.Fatalf("GetDetail(scoped reviewer) answer/rubric = %q/%q, want visible", row.AnswerKey, row.RubricHtml)
		}
	})

	t.Run("validation rejects invalid target level and multiple answer key", func(t *testing.T) {
		store := &fakeQuestionStore{createRow: db.CbtQuestion{ID: questionID}}
		svc := &CbtQuestion{q: store}
		base := SaveCbtQuestionInput{
			SubjectID:      pgtype.UUID{Bytes: [16]byte{3}, Valid: true},
			AuthoringMode:  "advance",
			QuestionText:   "Pilih semua bilangan prima",
			QuestionType:   "multiple_answer",
			OptionA:        "2",
			OptionB:        "4",
			OptionC:        "5",
			AnswerKey:      "A",
			WorkflowStatus: "submitted",
			AuthorUsername: "guru",
			Actor:          CbtQuestionActor{Username: "guru", Roles: []string{"guru"}},
		}

		invalidLevel := base
		invalidLevel.TargetLevel = "X"
		_, err := svc.Create(ctx, invalidLevel)
		if err == nil || !strings.Contains(err.Error(), "target_level") {
			t.Fatalf("Create(invalid target_level) error = %v, want target_level validation", err)
		}

		_, err = svc.Create(ctx, base)
		if err == nil || !strings.Contains(err.Error(), "multiple_answer") {
			t.Fatalf("Create(multiple_answer single key) error = %v, want multiple_answer validation", err)
		}
		if store.createCalls != 0 {
			t.Fatalf("Create(validation errors) createCalls = %d, want 0", store.createCalls)
		}
	})
}

func TestCbtQuestionMegaBulkWorkflowResults(t *testing.T) {
	ctx := context.Background()
	questionID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000033001")
	secondID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000033002")

	_, err := (&CbtQuestion{}).BulkWorkflow(ctx, BulkCbtQuestionWorkflowInput{Action: "not-real", QuestionIDs: []pgtype.UUID{questionID}})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("BulkWorkflow(invalid action) error = %v, want ErrBadRequest", err)
	}

	failedStore := &fakeQuestionStore{current: megaQuestionBase(questionID, "reviewed"), canApprove: true}
	failedSvc := &CbtQuestion{q: failedStore}
	failed, err := failedSvc.BulkWorkflow(ctx, BulkCbtQuestionWorkflowInput{
		Action:      "publish",
		QuestionIDs: []pgtype.UUID{questionID},
		Actor:       CbtQuestionActor{Username: "admin", Roles: []string{"admin"}},
	})
	if err != nil {
		t.Fatalf("BulkWorkflow(failed item) error = %v", err)
	}
	if failed.Success != 0 || failed.Failed != 1 || len(failed.Items) != 1 || failed.Items[0].OK || failed.Items[0].Error == "" {
		t.Fatalf("BulkWorkflow(failed item) result = %+v, want one recorded item failure", failed)
	}
	if failedStore.updateCalls != 0 || failedStore.auditCalls != 0 {
		t.Fatalf("BulkWorkflow(failed item) update/audit = %d/%d, want no mutation", failedStore.updateCalls, failedStore.auditCalls)
	}

	store := &fakeQuestionStore{current: megaQuestionBase(questionID, "draft")}
	svc := &CbtQuestion{q: store}
	result, err := svc.BulkWorkflow(ctx, BulkCbtQuestionWorkflowInput{
		Action:      "submit_review",
		QuestionIDs: []pgtype.UUID{questionID, secondID},
		Actor:       CbtQuestionActor{Username: "author"},
		Notes:       "bulk submit",
	})
	if err != nil {
		t.Fatalf("BulkWorkflow(submit alias) error = %v", err)
	}
	if result.Action != "submit_for_review" || result.Total != 2 || result.Success != 2 || result.Failed != 0 {
		t.Fatalf("BulkWorkflow(submit alias) result = %+v, want normalized action with two successes against fake current", result)
	}
	if len(result.Items) != 2 || !result.Items[0].OK || result.Items[0].Workflow != "diperiksa" {
		t.Fatalf("BulkWorkflow(submit alias) items = %+v, want diperiksa successes", result.Items)
	}
	if store.updateCalls != 2 || store.auditCalls != 2 || len(store.workflowEvents) != 2 {
		t.Fatalf("BulkWorkflow(submit alias) update/audit/events = %d/%d/%d, want 2/2/2", store.updateCalls, store.auditCalls, len(store.workflowEvents))
	}
}
