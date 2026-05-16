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

func (s *CbtQuestion) Timeline(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) ([]db.ListCbtQuestionTimelineRow, error) {
	if _, err := s.GetDetail(ctx, id, actor); err != nil {
		return nil, err
	}
	rows, err := s.q.ListCbtQuestionTimeline(ctx, id)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtQuestionTimelineRow{}, nil
	}
	return rows, nil
}

func (s *CbtQuestion) WorkflowEvents(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) ([]db.ListBankSoalQuestionWorkflowEventsRow, error) {
	if _, err := s.GetDetail(ctx, id, actor); err != nil {
		return nil, err
	}
	rows, err := s.q.ListBankSoalQuestionWorkflowEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListBankSoalQuestionWorkflowEventsRow{}, nil
	}
	return rows, nil
}

func (s *CbtQuestion) Versions(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) ([]db.ListCbtQuestionVersionsRow, error) {
	if _, err := s.GetDetail(ctx, id, actor); err != nil {
		return nil, err
	}
	rows, err := s.q.ListCbtQuestionVersions(ctx, id)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtQuestionVersionsRow{}, nil
	}
	return rows, nil
}

func (s *CbtQuestion) BulkWorkflow(ctx context.Context, in BulkCbtQuestionWorkflowInput) (BulkCbtQuestionWorkflowResult, error) {
	action := normalizeWorkflowAction(in.Action)
	switch action {
	case "submit_for_review", "request_revision", "mark_reviewed", "reject", "approve", "publish", "archive":
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
	case "submit_for_review":
		return s.SubmitForReview(ctx, id, actor, notes)
	case "request_revision":
		return s.RequestRevision(ctx, id, actor, notes)
	case "mark_reviewed":
		return s.MarkReviewed(ctx, id, actor, notes)
	case "approve":
		return s.Approve(ctx, id, actor, notes)
	case "reject":
		return s.Reject(ctx, id, actor, notes)
	case "publish":
		return s.Publish(ctx, id, actor)
	case "archive":
		return s.Archive(ctx, id, actor)
	default:
		return db.CbtQuestion{}, fmt.Errorf("%w: action tidak didukung", domain.ErrBadRequest)
	}
}

func normalizeWorkflowAction(action string) string {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "submit_review":
		return "submit_for_review"
	case "return_revision":
		return "request_revision"
	default:
		return strings.TrimSpace(strings.ToLower(action))
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

func normalizeQuestionTargetLevel(value string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case "":
		return "", true
	case "VII", "VIII", "IX":
		return normalized, true
	default:
		return normalized, false
	}
}

func normalizeQuestionTargetLevelFilter(value string) string {
	normalized, ok := normalizeQuestionTargetLevel(value)
	if !ok {
		return ""
	}
	return normalized
}

func normalizeQuestionDifficultyFilter(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case string(db.CbtQuestionDifficultyEnumEasy):
		return string(db.CbtQuestionDifficultyEnumEasy)
	case string(db.CbtQuestionDifficultyEnumMedium):
		return string(db.CbtQuestionDifficultyEnumMedium)
	case string(db.CbtQuestionDifficultyEnumHard):
		return string(db.CbtQuestionDifficultyEnumHard)
	default:
		return ""
	}
}

func normalizeQuestionMetadataFilter(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "complete", "lengkap":
		return "complete"
	case "gap", "missing", "kurang":
		return "gap"
	default:
		return ""
	}
}

func normalizeQuestionSortOrder(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "code_asc", "updated_desc", "created_asc", "difficulty_asc", "type_asc":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return ""
	}
}

func (s *CbtQuestion) SubmitReview(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	return s.SubmitForReview(ctx, id, actor, reviewNotes)
}

func (s *CbtQuestion) SubmitForReview(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	fromStatus := strings.TrimSpace(current.WorkflowStatus)
	if !workflowStatusIn(fromStatus, "draft", "revision_needed", "rejected") {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya draft, revision_needed, atau rejected yang dapat diajukan review", domain.ErrConflict)
	}
	if !actor.IsAdmin() && (strings.TrimSpace(current.AuthorUsername) == "" || current.AuthorUsername != actor.Username) {
		return db.CbtQuestion{}, domain.ErrForbidden
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = workflowMutationActor(actor)
	input.WorkflowStatus = "submitted"
	input.ReviewerUsername = ""
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.updateWithWorkflowAudit(ctx, input, actor, "submit_for_review", reviewNotes, map[string]any{"workflow_status": "submitted"}, fromStatus, "submitted")
}

func (s *CbtQuestion) RequestRevision(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	fromStatus := strings.TrimSpace(current.WorkflowStatus)
	if workflowStatusIn(fromStatus, "submitted", "review") {
		if err := s.requireBankSoalReviewer(ctx, actor, current); err != nil {
			return db.CbtQuestion{}, err
		}
	} else if workflowStatusIn(fromStatus, "reviewed", "approved") {
		if err := s.requireBankSoalApprover(ctx, actor, current, false); err != nil {
			return db.CbtQuestion{}, err
		}
	} else {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal submitted, review, reviewed, atau approved yang dapat diminta revisi", domain.ErrConflict)
	}
	if current.Status != "" && current.Status != db.CbtQuestionStatusEnumDraft {
		return db.CbtQuestion{}, fmt.Errorf("%w: soal yang sudah terbit harus dibuat sebagai revisi baru", domain.ErrConflict)
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = workflowMutationActor(actor)
	input.WorkflowStatus = "revision_needed"
	input.ReviewerUsername = actor.Username
	input.ApproverUsername = ""
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	if strings.TrimSpace(input.ReviewNotes) == "" {
		input.ReviewNotes = "Perlu revisi berdasarkan review."
	}
	return s.updateWithWorkflowAudit(ctx, input, actor, "request_revision", reviewNotes, map[string]any{"workflow_status": "revision_needed"}, fromStatus, "revision_needed")
}

func (s *CbtQuestion) MarkReviewed(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	fromStatus := strings.TrimSpace(current.WorkflowStatus)
	if !workflowStatusIn(fromStatus, "submitted", "review") {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal submitted atau review yang dapat ditandai layak", domain.ErrConflict)
	}
	if err := s.requireBankSoalReviewer(ctx, actor, current); err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = workflowMutationActor(actor)
	input.WorkflowStatus = "reviewed"
	input.ReviewerUsername = actor.Username
	input.ApproverUsername = ""
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.updateWithWorkflowAudit(ctx, input, actor, "mark_reviewed", reviewNotes, map[string]any{"workflow_status": "reviewed"}, fromStatus, "reviewed")
}

func (s *CbtQuestion) Approve(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	fromStatus := strings.TrimSpace(current.WorkflowStatus)
	if !workflowStatusIn(fromStatus, "reviewed", "submitted", "review") {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal reviewed yang dapat disetujui", domain.ErrConflict)
	}
	if err := s.requireBankSoalApprover(ctx, actor, current, false); err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = workflowMutationActor(actor)
	input.WorkflowStatus = "approved"
	input.ApproverUsername = actor.Username
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.updateWithWorkflowAudit(ctx, input, actor, "approve", reviewNotes, map[string]any{"workflow_status": "approved"}, fromStatus, "approved")
}

func (s *CbtQuestion) Reject(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	fromStatus := strings.TrimSpace(current.WorkflowStatus)
	if !workflowStatusIn(fromStatus, "submitted", "review", "reviewed") {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal submitted, review, atau reviewed yang dapat ditolak", domain.ErrConflict)
	}
	if err := s.requireBankSoalReviewer(ctx, actor, current); err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = workflowMutationActor(actor)
	input.WorkflowStatus = "rejected"
	input.ReviewerUsername = actor.Username
	input.ApproverUsername = ""
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.updateWithWorkflowAudit(ctx, input, actor, "reject", reviewNotes, map[string]any{"workflow_status": "rejected"}, fromStatus, "rejected")
}

func (s *CbtQuestion) ReturnToRevision(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error) {
	return s.RequestRevision(ctx, id, actor, reviewNotes)
}

func (s *CbtQuestion) Publish(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	fromStatus := strings.TrimSpace(current.WorkflowStatus)
	if current.WorkflowStatus != "approved" {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal approved yang dapat dipublish", domain.ErrConflict)
	}
	if err := s.requireBankSoalApprover(ctx, actor, current, true); err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = workflowMutationActor(actor)
	input.WorkflowStatus = "published"
	input.ApproverUsername = actor.Username
	input.Status = db.CbtQuestionStatusEnumPublished
	return s.updateWithWorkflowAudit(ctx, input, actor, "publish", "", map[string]any{"workflow_status": "published", "status": db.CbtQuestionStatusEnumPublished}, fromStatus, "published")
}

func (s *CbtQuestion) Archive(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	fromStatus := strings.TrimSpace(current.WorkflowStatus)
	if !workflowStatusIn(fromStatus, "approved", "published", "rejected") {
		return db.CbtQuestion{}, fmt.Errorf("%w: hanya soal approved, published, atau rejected yang dapat diarsipkan", domain.ErrConflict)
	}
	if err := s.requireBankSoalApprover(ctx, actor, current, false); err != nil {
		return db.CbtQuestion{}, err
	}
	if questionUsageLocked(current.PackageCount, current.AnswerCount) {
		return db.CbtQuestion{}, fmt.Errorf("%w: soal sudah masuk paket ujian atau memiliki jawaban siswa. Duplikat soal untuk membuat revisi baru", domain.ErrConflict)
	}
	input := questionInputFromCurrent(current, actor.Username)
	input.Actor = workflowMutationActor(actor)
	input.WorkflowStatus = "archived"
	input.Status = db.CbtQuestionStatusEnumArchived
	return s.updateWithWorkflowAudit(ctx, input, actor, "archive", "", map[string]any{"workflow_status": "archived", "status": db.CbtQuestionStatusEnumArchived}, fromStatus, "archived")
}

func workflowMutationActor(actor CbtQuestionActor) CbtQuestionActor {
	actor = normalizeCbtQuestionActor(actor)
	return CbtQuestionActor{
		UserID:      actor.UserID,
		Username:    actor.Username,
		Roles:       []string{"admin"},
		Permissions: actor.Permissions,
	}
}

func workflowStatusIn(value string, allowed ...string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	for _, item := range allowed {
		if value == strings.TrimSpace(strings.ToLower(item)) {
			return true
		}
	}
	return false
}

func (s *CbtQuestion) requireBankSoalReviewer(ctx context.Context, actor CbtQuestionActor, current db.GetCbtQuestionRow) error {
	actor = normalizeCbtQuestionActor(actor)
	if actor.IsAdmin() {
		return nil
	}
	if strings.TrimSpace(current.AuthorUsername) != "" && current.AuthorUsername == actor.Username {
		return domain.ErrForbidden
	}
	if !actor.HasPermission("bank_soal.review") || !actor.UserID.Valid {
		return domain.ErrForbidden
	}
	ok, err := s.q.CanBankSoalUserReview(ctx, db.CanBankSoalUserReviewParams{
		UserID:     actor.UserID,
		SubjectID:  current.SubjectID,
		GradeLevel: workflowScopeGradeLevel(current.TargetLevel),
	})
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *CbtQuestion) requireBankSoalApprover(ctx context.Context, actor CbtQuestionActor, current db.GetCbtQuestionRow, requirePublish bool) error {
	actor = normalizeCbtQuestionActor(actor)
	if actor.IsAdmin() {
		return nil
	}
	if strings.TrimSpace(current.AuthorUsername) != "" && current.AuthorUsername == actor.Username {
		return domain.ErrForbidden
	}
	hasPermission := actor.HasPermission("bank_soal.approve") || actor.HasPermission("bank_soal.publish")
	if requirePublish {
		hasPermission = actor.HasPermission("bank_soal.publish")
	}
	if !hasPermission || !actor.UserID.Valid {
		return domain.ErrForbidden
	}
	ok, err := s.q.CanBankSoalUserApprove(ctx, db.CanBankSoalUserApproveParams{
		UserID:     actor.UserID,
		SubjectID:  current.SubjectID,
		GradeLevel: workflowScopeGradeLevel(current.TargetLevel),
	})
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func workflowScopeGradeLevel(targetLevel pgtype.Text) pgtype.Int2 {
	if !targetLevel.Valid {
		return pgtype.Int2{}
	}
	switch strings.TrimSpace(targetLevel.String) {
	case "VII":
		return pgtype.Int2{Int16: 7, Valid: true}
	case "VIII":
		return pgtype.Int2{Int16: 8, Valid: true}
	case "IX":
		return pgtype.Int2{Int16: 9, Valid: true}
	default:
		return pgtype.Int2{}
	}
}

func enrichWorkflowEventMetadata(metadata map[string]any, current db.GetCbtQuestionRow, row db.CbtQuestion) map[string]any {
	enriched := map[string]any{}
	for key, value := range metadata {
		enriched[key] = value
	}
	if strings.TrimSpace(current.ReviewerUsername) != "" || strings.TrimSpace(row.ReviewerUsername) != "" {
		enriched["from_reviewer_username"] = strings.TrimSpace(current.ReviewerUsername)
		enriched["to_reviewer_username"] = strings.TrimSpace(row.ReviewerUsername)
	}
	if strings.TrimSpace(current.ApproverUsername) != "" || strings.TrimSpace(row.ApproverUsername) != "" {
		enriched["from_approver_username"] = strings.TrimSpace(current.ApproverUsername)
		enriched["to_approver_username"] = strings.TrimSpace(row.ApproverUsername)
	}
	if strings.TrimSpace(string(current.Status)) != "" || strings.TrimSpace(string(row.Status)) != "" {
		enriched["from_publication_status"] = strings.TrimSpace(string(current.Status))
		enriched["to_publication_status"] = strings.TrimSpace(string(row.Status))
	}
	return enriched
}

func (s *CbtQuestion) updateWithWorkflowAudit(ctx context.Context, input SaveCbtQuestionInput, actor CbtQuestionActor, action string, note string, metadata map[string]any, fromStatus string, toStatus string) (db.CbtQuestion, error) {
	actor = normalizeCbtQuestionActor(actor)
	mutationActor := normalizeCbtQuestionActor(inputActor(input))
	current, err := s.q.GetCbtQuestion(ctx, input.ID)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	if err := s.requireModifyQuestion(ctx, mutationActor, current); err != nil {
		return db.CbtQuestion{}, err
	}
	if questionUsageLocked(current.PackageCount, current.AnswerCount) {
		return db.CbtQuestion{}, fmt.Errorf("%w: soal sudah masuk paket ujian atau memiliki jawaban siswa. Duplikat soal untuk membuat revisi baru", domain.ErrConflict)
	}
	params, err := buildUpdateQuestionParams(current, input)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	return s.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		if err := s.validateMediaAssetIDs(ctx, store, input.MediaAssetIDs, current.ID, mutationActor); err != nil {
			return db.CbtQuestion{}, err
		}
		row, err := store.UpdateCbtQuestion(ctx, params)
		if err != nil {
			return db.CbtQuestion{}, err
		}
		if err := logQuestionAudit(ctx, store, row.ID, actor.Username, action, note, metadata); err != nil {
			return db.CbtQuestion{}, err
		}
		eventMetadata := enrichWorkflowEventMetadata(metadata, current, row)
		if err := logQuestionWorkflowEvent(ctx, store, row.ID, actor, fromStatus, toStatus, action, note, eventMetadata); err != nil {
			return db.CbtQuestion{}, err
		}
		return row, nil
	})
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
	input.VersionGroupID = pgtype.UUID{}
	input.VersionNumber = 1
	input.SourceQuestionID = pgtype.UUID{}
	input.SupersedesQuestionID = pgtype.UUID{}
	input.VersionNote = ""
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
	versionGroupID := current.VersionGroupID
	if !versionGroupID.Valid {
		versionGroupID = current.ID
	}
	input.VersionGroupID = versionGroupID
	input.VersionNumber = 1
	input.SourceQuestionID = current.ID
	input.SupersedesQuestionID = current.ID
	input.VersionNote = reviewNotes

	if err := s.requireCreateQuestion(ctx, actor, input.EventID, input.SubjectID); err != nil {
		return db.CbtQuestion{}, err
	}
	if createInputBypassesWorkflow(input) {
		return db.CbtQuestion{}, fmt.Errorf("%w: soal baru hanya boleh dibuat sebagai draft atau diajukan review", domain.ErrBadRequest)
	}
	params, err := buildCreateQuestionParams(input)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	return s.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		if err := s.validateMediaAssetIDs(ctx, store, input.MediaAssetIDs, current.ID, actor); err != nil {
			return db.CbtQuestion{}, err
		}
		nextVersion, err := store.GetNextCbtQuestionVersionNumber(ctx, versionGroupID)
		if err != nil {
			return db.CbtQuestion{}, err
		}
		if nextVersion < 2 {
			nextVersion = normalizeVersionNumber(current.VersionNumber) + 1
		}
		params.VersionNumber = nextVersion
		if err := store.MarkCbtQuestionVersionGroupNotLatest(ctx, versionGroupID); err != nil {
			return db.CbtQuestion{}, err
		}
		row, err := store.CreateCbtQuestion(ctx, params)
		if err != nil {
			return db.CbtQuestion{}, err
		}
		if err := logQuestionAudit(ctx, store, row.ID, actor.Username, "revision", reviewNotes, map[string]any{"source_question_id": cbtQuestionUUIDString(id)}); err != nil {
			return db.CbtQuestion{}, err
		}
		return row, nil
	})
}
