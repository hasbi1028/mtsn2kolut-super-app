package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func (s *CbtQuestion) Timeline(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) ([]db.CbtQuestionAuditLog, error) {
	if _, err := s.GetDetail(ctx, id, actor); err != nil {
		return nil, err
	}
	rows, err := s.q.ListCbtQuestionTimeline(ctx, id)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.CbtQuestionAuditLog{}, nil
	}
	return rows, nil
}

func (s *CbtQuestion) BulkWorkflow(ctx context.Context, in BulkCbtQuestionWorkflowInput) (BulkCbtQuestionWorkflowResult, error) {
	action := strings.TrimSpace(in.Action)
	switch action {
	case "approve", "reject", "publish":
	default:
		return BulkCbtQuestionWorkflowResult{}, fmt.Errorf("%w: action tidak didukung", domain.ErrBadRequest)
	}
	result := BulkCbtQuestionWorkflowResult{
		Action: action,
		Total:  len(in.QuestionIDs),
		Items:  make([]BulkCbtQuestionWorkflowItem, 0, len(in.QuestionIDs)),
	}
	for _, id := range in.QuestionIDs {
		item := BulkCbtQuestionWorkflowItem{QuestionID: id}
		row, err := s.applyBulkWorkflowItem(ctx, id, action, in.Actor, in.Notes)
		if err != nil {
			item.Error = err.Error()
			result.Failed++
		} else {
			item.OK = true
			item.Status = string(row.Status)
			item.Workflow = row.WorkflowStatus
			result.Success++
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func (s *CbtQuestion) applyBulkWorkflowItem(ctx context.Context, id pgtype.UUID, action string, actor CbtQuestionActor, notes string) (db.CbtQuestion, error) {
	switch action {
	case "approve":
		return s.Approve(ctx, id, actor, notes)
	case "reject":
		return s.Reject(ctx, id, actor, notes)
	case "publish":
		return s.Publish(ctx, id, actor)
	default:
		return db.CbtQuestion{}, fmt.Errorf("%w: action tidak didukung", domain.ErrBadRequest)
	}
}

func normalizeRevisionSource(value string) string {
	normalized := strings.TrimSpace(value)
	switch normalized {
	case "item_analysis", "reviewer", "workflow":
		return normalized
	default:
		return ""
	}
}

func normalizeQuestionScope(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "global":
		return "global"
	case "event_pool":
		return "event_pool"
	default:
		return ""
	}
}

func normalizeQuestionStatusFilter(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case string(db.CbtQuestionStatusEnumDraft):
		return string(db.CbtQuestionStatusEnumDraft)
	case string(db.CbtQuestionStatusEnumPublished):
		return string(db.CbtQuestionStatusEnumPublished)
	case string(db.CbtQuestionStatusEnumArchived):
		return string(db.CbtQuestionStatusEnumArchived)
	default:
		return ""
	}
}

func (s *CbtQuestion) SubmitReview(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	if current.WorkflowStatus != "draft" && current.WorkflowStatus != "rejected" {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya draft atau rejected yang dapat diajukan review", domain.ErrConflict)
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = actor
	input.WorkflowStatus = "review"
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.updateWithAudit(ctx, input, "submit_review", reviewNotes, map[string]any{"workflow_status": "review"})
}

func (s *CbtQuestion) Approve(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	if current.WorkflowStatus != "review" {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal review yang dapat disetujui", domain.ErrConflict)
	}
	if !actor.IsAdmin() {
		if err := s.requireWorkflowRole(ctx, actor.Username, current, db.CbtEventMemberRoleReviewer); err != nil {
			return db.CbtQuestion{}, err
		}
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = CbtQuestionActor{Username: actor.Username, Roles: []string{"admin"}}
	input.WorkflowStatus = "approved"
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.updateWithAudit(ctx, input, "approve", reviewNotes, map[string]any{"workflow_status": "approved"})
}

func (s *CbtQuestion) Reject(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	if current.WorkflowStatus != "review" {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal review yang dapat ditolak", domain.ErrConflict)
	}
	if !actor.IsAdmin() {
		if err := s.requireWorkflowRole(ctx, actor.Username, current, db.CbtEventMemberRoleReviewer); err != nil {
			return db.CbtQuestion{}, err
		}
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = CbtQuestionActor{Username: actor.Username, Roles: []string{"admin"}}
	input.WorkflowStatus = "rejected"
	input.ReviewerUsername = actor.Username
	input.ApproverUsername = ""
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.updateWithAudit(ctx, input, "reject", reviewNotes, map[string]any{"workflow_status": "rejected"})
}

func (s *CbtQuestion) ReturnToRevision(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	if !actor.IsAdmin() && !actor.HasPermission("bank_soal.review") && !actor.HasPermission("bank_soal.publish") {
		return db.CbtQuestion{}, domain.ErrForbidden
	}
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	if current.WorkflowStatus != "approved" {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal approved yang dapat dikembalikan ke revisi", domain.ErrConflict)
	}
	if current.Status != "" && current.Status != db.CbtQuestionStatusEnumDraft {
		return db.CbtQuestion{}, fmt.Errorf("%w: soal yang sudah terbit harus dibuat sebagai revisi baru", domain.ErrConflict)
	}
	if questionUsageLocked(current.PackageCount, current.AnswerCount) {
		return db.CbtQuestion{}, fmt.Errorf("%w: soal sudah masuk paket ujian atau memiliki jawaban siswa. Buat revisi baru agar riwayat ujian tetap aman", domain.ErrConflict)
	}
	input := questionInputFromCurrent(current, actor.Username)
	// Returning an approved question is a reviewer/publisher workflow operation, not a
	// normal author edit. Use an internal admin-scoped actor for the shared update path
	// while preserving the real username in audit fields.
	input.Actor = CbtQuestionActor{UserID: actor.UserID, Username: actor.Username, Roles: []string{"admin"}, Permissions: actor.Permissions}
	input.WorkflowStatus = "rejected"
	input.ReviewerUsername = actor.Username
	input.ApproverUsername = ""
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	if strings.TrimSpace(input.ReviewNotes) == "" {
		input.ReviewNotes = "Dikembalikan ke revisi setelah disetujui."
	}
	return s.updateWithAudit(ctx, input, "return_revision", reviewNotes, map[string]any{"workflow_status": "rejected"})
}

func (s *CbtQuestion) Publish(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	if !actor.CanPublishBankSoal() {
		return db.CbtQuestion{}, domain.ErrForbidden
	}
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	if current.WorkflowStatus != "approved" {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal approved yang dapat dipublish", domain.ErrConflict)
	}
	input := questionInputFromCurrent(current, actor.Username)
	// Publish already checked bank_soal.publish/admin above. Use an internal admin-scoped
	// actor for the shared update path so permission-based publishers can transition
	// approved questions without being blocked by author-only draft modification rules.
	input.Actor = CbtQuestionActor{UserID: actor.UserID, Username: actor.Username, Roles: []string{"admin"}, Permissions: actor.Permissions}
	input.Status = db.CbtQuestionStatusEnumPublished
	return s.updateWithAudit(ctx, input, "publish", "", map[string]any{"status": db.CbtQuestionStatusEnumPublished})
}

func (s *CbtQuestion) Archive(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	if !actor.CanPublishBankSoal() {
		return db.CbtQuestion{}, domain.ErrForbidden
	}
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	if questionUsageLocked(current.PackageCount, current.AnswerCount) {
		return db.CbtQuestion{}, fmt.Errorf("%w: soal sudah masuk paket ujian atau memiliki jawaban siswa. Duplikat soal untuk membuat revisi baru", domain.ErrConflict)
	}
	input := questionInputFromCurrent(current, actor.Username)
	// Archive shares the publish-level privilege and uses the update path; keep the
	// human actor username for audit while bypassing author-only draft modification rules.
	input.Actor = CbtQuestionActor{UserID: actor.UserID, Username: actor.Username, Roles: []string{"admin"}, Permissions: actor.Permissions}
	input.Status = db.CbtQuestionStatusEnumArchived
	return s.Update(ctx, input)
}

func (s *CbtQuestion) DuplicateAsDraft(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	if err := s.requireDuplicateSourceAccess(ctx, actor, current); err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = actor
	input.ID = pgtype.UUID{}
	input.Status = db.CbtQuestionStatusEnumDraft
	input.WorkflowStatus = "draft"
	input.ReviewerUsername = ""
	input.ApproverUsername = ""
	input.ReviewNotes = ""
	if input.Code != "" {
		input.Code = input.Code + "-COPY"
	}
	return s.createWithAudit(ctx, input, "duplicate", "", map[string]any{"source_question_id": cbtQuestionUUIDString(id)})
}

func (s *CbtQuestion) DuplicateForRevision(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	if err := s.requireDuplicateSourceAccess(ctx, actor, current); err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = actor
	input.ID = pgtype.UUID{}
	input.Status = db.CbtQuestionStatusEnumDraft
	input.WorkflowStatus = "rejected"
	input.ReviewerUsername = actor.Username
	input.ApproverUsername = ""
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	if strings.TrimSpace(input.ReviewNotes) == "" {
		input.ReviewNotes = "Perlu revisi berdasarkan analisis butir."
	}
	if input.Code != "" {
		input.Code = fmt.Sprintf("%s-REV-%s", input.Code, time.Now().Format("20060102150405"))
	}
	return s.createWithAudit(ctx, input, "revision", reviewNotes, map[string]any{"source_question_id": cbtQuestionUUIDString(id)})
}
