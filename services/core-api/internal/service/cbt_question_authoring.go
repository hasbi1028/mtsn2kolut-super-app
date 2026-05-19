package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	nethtml "golang.org/x/net/html"

	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/platform/logging"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func (s *CbtQuestion) Get(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	return s.q.GetCbtQuestion(ctx, id)
}

func (s *CbtQuestion) GetDetail(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) (db.GetCbtQuestionDetailRow, error) {
	row, err := s.q.GetCbtQuestionDetail(ctx, id)
	if err != nil {
		return db.GetCbtQuestionDetailRow{}, normalizeNoRows(err)
	}
	actor = normalizeCbtQuestionActor(actor)
	canView, canSeeAnswerKey, err := s.questionDetailAccess(ctx, actor, row)
	if err != nil {
		return db.GetCbtQuestionDetailRow{}, err
	}
	if !canView {
		return db.GetCbtQuestionDetailRow{}, domain.ErrForbidden
	}
	if !canSeeAnswerKey {
		row.AnswerKey = ""
		row.RubricHtml = ""
	}
	return row, nil
}

func (s *CbtQuestion) Create(ctx context.Context, input SaveCbtQuestionInput) (db.CbtQuestion, error) {
	return s.createWithAudit(ctx, input, "create", "", nil)
}

func (s *CbtQuestion) createWithAudit(ctx context.Context, input SaveCbtQuestionInput, action string, note string, metadata map[string]any) (db.CbtQuestion, error) {
	actor := normalizeCbtQuestionActor(inputActor(input))
	logging.Info(ctx, "cbt_question_create_request_received", questionInputLogAttrs(input, actor)...)
	if err := s.requireCreateQuestion(ctx, actor, input.EventID, input.SubjectID); err != nil {
		logging.Warn(ctx, "cbt_question_create_authorization_failed", append(questionInputLogAttrs(input, actor), slog.String("error", err.Error()))...)
		return db.CbtQuestion{}, err
	}
	if createInputBypassesWorkflow(input) {
		err := fmt.Errorf("%w: soal baru hanya boleh dibuat sebagai draft atau diajukan review", domain.ErrBadRequest)
		logging.Warn(ctx, "cbt_question_create_validation_failed", append(questionInputLogAttrs(input, actor), slog.String("validation_field", "workflow_status"), slog.String("error", err.Error()))...)
		return db.CbtQuestion{}, err
	}
	params, err := buildCreateQuestionParams(input)
	if err != nil {
		logging.Warn(ctx, "cbt_question_create_validation_failed", append(questionInputLogAttrs(input, actor), slog.String("error", err.Error()))...)
		return db.CbtQuestion{}, err
	}
	return s.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		if err := s.validateQuestionReferences(ctx, store, input); err != nil {
			logging.Warn(ctx, "cbt_question_create_reference_invalid", append(questionInputLogAttrs(input, actor), slog.String("error", err.Error()))...)
			return db.CbtQuestion{}, err
		}
		if strings.TrimSpace(params.Code) == "" {
			code, err := store.GenerateCbtQuestionAcademicCode(ctx, db.GenerateCbtQuestionAcademicCodeParams{
				SubjectID:    params.SubjectID,
				TargetLevel:  params.TargetLevel.String,
				QuestionType: params.QuestionType,
			})
			if err != nil {
				logging.Error(ctx, "cbt_question_create_code_generation_failed", err, questionInputLogAttrs(input, actor)...)
				return db.CbtQuestion{}, err
			}
			params.Code = strings.TrimSpace(code)
			input.Code = params.Code
		}
		if err := s.validateMediaAssetIDs(ctx, store, input.MediaAssetIDs, pgtype.UUID{}, actor); err != nil {
			logging.Warn(ctx, "cbt_question_create_validation_failed", append(questionInputLogAttrs(input, actor), slog.String("validation_field", "media_asset_ids"), slog.String("error", err.Error()))...)
			return db.CbtQuestion{}, err
		}
		if err := store.AcquireCbtQuestionDraftDuplicateLock(ctx, draftDuplicateFingerprint(params)); err != nil {
			logging.Error(ctx, "cbt_question_create_duplicate_lock_failed", err, questionInputLogAttrs(input, actor)...)
			return db.CbtQuestion{}, err
		}
		if duplicate, ok, err := s.findRecentDraftDuplicate(ctx, store, params); err != nil {
			logging.Error(ctx, "cbt_question_create_duplicate_lookup_failed", err, questionInputLogAttrs(input, actor)...)
			return db.CbtQuestion{}, err
		} else if ok {
			logging.Info(ctx, "cbt_question_create_duplicate_reused", append(questionInputLogAttrs(input, actor), slog.String("question_id", cbtQuestionUUIDString(duplicate.ID)))...)
			return duplicate, nil
		}
		row, err := store.CreateCbtQuestion(ctx, params)
		if err != nil {
			logging.Error(ctx, "cbt_question_create_db_failed", err, questionInputLogAttrs(input, actor)...)
			return db.CbtQuestion{}, err
		}
		if err := logQuestionAudit(ctx, store, row.ID, actor.Username, action, note, metadata); err != nil {
			logging.Error(ctx, "cbt_question_create_audit_failed", err, append(questionInputLogAttrs(input, actor), slog.String("question_id", cbtQuestionUUIDString(row.ID)))...)
			return db.CbtQuestion{}, err
		}
		logging.Info(ctx, "cbt_question_create_success", append(questionInputLogAttrs(input, actor), slog.String("question_id", cbtQuestionUUIDString(row.ID)))...)
		return row, nil
	})
}

func questionInputLogAttrs(input SaveCbtQuestionInput, actor CbtQuestionActor) []slog.Attr {
	stemLength := len(strings.TrimSpace(input.StemHTML))
	questionTextLength := len(strings.TrimSpace(input.QuestionText))
	return []slog.Attr{
		slog.String("module", "bank-soal"),
		slog.String("username", actor.Username),
		slog.String("subject_id", cbtQuestionUUIDString(input.SubjectID)),
		slog.Bool("event_id_present", input.EventID.Valid),
		slog.String("event_id", cbtQuestionUUIDString(input.EventID)),
		slog.String("question_type", strings.TrimSpace(input.QuestionType)),
		slog.String("target_level", strings.TrimSpace(input.TargetLevel)),
		slog.String("difficulty", string(input.Difficulty)),
		slog.String("workflow_status", normalizeWorkflowStatus(input.WorkflowStatus)),
		slog.String("authoring_mode", strings.TrimSpace(input.AuthoringMode)),
		slog.Int("options_count", len(input.Options)),
		slog.Int("stem_length", stemLength),
		slog.Int("question_text_length", questionTextLength),
	}
}

func (s *CbtQuestion) validateQuestionReferences(ctx context.Context, store cbtQuestionStore, input SaveCbtQuestionInput) error {
	if !input.SubjectID.Valid {
		return fmt.Errorf("%w: subject_id wajib diisi", domain.ErrBadRequest)
	}
	exists, err := store.CbtQuestionSubjectExists(ctx, input.SubjectID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: subject_id tidak ditemukan. Pilih ulang mapel dari daftar terbaru", domain.ErrBadRequest)
	}
	if input.EventID.Valid {
		exists, err := store.CbtQuestionEventExists(ctx, input.EventID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("%w: event_id tidak ditemukan. Matikan mode khusus kegiatan atau pilih ulang kegiatan", domain.ErrBadRequest)
		}
	}
	return nil
}

func createInputBypassesWorkflow(input SaveCbtQuestionInput) bool {
	status := strings.TrimSpace(string(input.Status))
	workflowStatus := normalizeWorkflowStatus(input.WorkflowStatus)
	return status == string(db.CbtQuestionStatusEnumPublished) ||
		status == string(db.CbtQuestionStatusEnumArchived) ||
		workflowStatus == "reviewed" ||
		workflowStatus == "approved" ||
		workflowStatus == "published" ||
		workflowStatus == "archived"
}

func (s *CbtQuestion) Update(ctx context.Context, input SaveCbtQuestionInput) (db.CbtQuestion, error) {
	return s.updateWithAudit(ctx, input, "update", strings.TrimSpace(input.ReviewNotes), nil)
}

func (s *CbtQuestion) updateWithAudit(ctx context.Context, input SaveCbtQuestionInput, action string, note string, metadata map[string]any) (db.CbtQuestion, error) {
	actor := normalizeCbtQuestionActor(inputActor(input))
	current, err := s.q.GetCbtQuestion(ctx, input.ID)
	if err != nil {
		return db.CbtQuestion{}, normalizeNoRows(err)
	}
	if err := s.requireModifyQuestion(ctx, actor, current); err != nil {
		return db.CbtQuestion{}, err
	}
	if questionUsageLocked(current.PackageCount, current.AnswerCount) {
		return db.CbtQuestion{}, fmt.Errorf("%w: soal sudah masuk paket ujian atau memiliki jawaban siswa. Duplikat soal untuk membuat revisi baru", domain.ErrConflict)
	}
	params, err := buildUpdateQuestionParams(current, input)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	if !actor.IsAdmin() && (!sameOptionalUUID(current.EventID, params.EventID) || !sameUUID(current.SubjectID, params.SubjectID)) {
		if err := s.requireCreateQuestion(ctx, actor, params.EventID, params.SubjectID); err != nil {
			return db.CbtQuestion{}, err
		}
	}
	return s.withMutationStore(ctx, func(store cbtQuestionStore) (db.CbtQuestion, error) {
		if err := s.validateQuestionReferences(ctx, store, input); err != nil {
			logging.Warn(ctx, "cbt_question_update_reference_invalid", append(questionInputLogAttrs(input, actor), slog.String("question_id", cbtQuestionUUIDString(input.ID)), slog.String("error", err.Error()))...)
			return db.CbtQuestion{}, err
		}
		if err := s.validateMediaAssetIDs(ctx, store, input.MediaAssetIDs, current.ID, actor); err != nil {
			return db.CbtQuestion{}, err
		}
		row, err := store.UpdateCbtQuestion(ctx, params)
		if err != nil {
			return db.CbtQuestion{}, err
		}
		if err := logQuestionAudit(ctx, store, row.ID, actor.Username, action, note, metadata); err != nil {
			return db.CbtQuestion{}, err
		}
		return row, nil
	})
}

func (s *CbtQuestion) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.DeleteWithActor(ctx, id, CbtQuestionActor{Roles: []string{"admin"}})
}

func (s *CbtQuestion) DeleteWithActor(ctx context.Context, id pgtype.UUID, actor CbtQuestionActor) error {
	actor = normalizeCbtQuestionActor(actor)
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return normalizeNoRows(err)
	}
	if err := s.requireModifyQuestion(ctx, actor, current); err != nil {
		return err
	}
	if current.WorkflowStatus != "" && current.WorkflowStatus != "draft" && current.WorkflowStatus != "rejected" {
		return fmt.Errorf("%w: soal hanya dapat dihapus saat masih draft atau sudah ditolak dan belum dipakai", domain.ErrConflict)
	}
	if current.Status != "" && current.Status != db.CbtQuestionStatusEnumDraft {
		return fmt.Errorf("%w: soal hanya dapat dihapus saat belum terbit", domain.ErrConflict)
	}
	if questionUsageLocked(current.PackageCount, current.AnswerCount) {
		return fmt.Errorf("%w: soal sudah masuk paket ujian atau memiliki jawaban siswa. Soal ini tidak dapat dihapus permanen", domain.ErrConflict)
	}
	return s.withMutationStoreExec(ctx, func(store cbtQuestionStore) error {
		if err := logQuestionAudit(ctx, store, id, actor.Username, "delete", "", nil); err != nil {
			return err
		}
		return store.DeleteCbtQuestion(ctx, id)
	})
}

func questionUsageLocked(packageCount, answerCount int32) bool {
	return packageCount > 0 || answerCount > 0
}

func draftDuplicateFingerprint(params db.CreateCbtQuestionParams) string {
	parts := []string{
		params.AuthorUsername,
		fmt.Sprintf("%x", params.SubjectID.Bytes),
		fmt.Sprintf("%t:%x", params.EventID.Valid, params.EventID.Bytes),
		params.QuestionType,
		params.Code,
		params.QuestionText,
		params.StemHtml,
		params.StemLatex,
		params.StimulusHtml,
		params.StimulusLatex,
		params.AnswerKey,
		params.Explanation,
		params.ExplanationHtml,
		params.RubricHtml,
		string(params.Difficulty),
		params.AcademicPhase,
		fmt.Sprintf("%t:%s", params.TargetLevel.Valid, params.TargetLevel.String),
		params.CpRef,
		params.TpRef,
		params.KdRef,
		params.IndicatorRef,
		params.MaterialTopic,
		params.CognitiveLevel,
		fmt.Sprintf("%t", params.HotsFlag),
		string(params.Options),
		string(params.MediaAssetIds),
		params.WorkflowStatus,
	}
	return strings.Join(parts, "\x1f")
}

func (s *CbtQuestion) findRecentDraftDuplicate(ctx context.Context, store cbtQuestionStore, params db.CreateCbtQuestionParams) (db.CbtQuestion, bool, error) {
	if strings.TrimSpace(params.AuthorUsername) == "" || params.Status != db.CbtQuestionStatusEnumDraft {
		return db.CbtQuestion{}, false, nil
	}
	if params.SourceQuestionID.Valid || params.SupersedesQuestionID.Valid {
		return db.CbtQuestion{}, false, nil
	}
	row, err := store.FindRecentCbtQuestionDraftDuplicate(ctx, db.FindRecentCbtQuestionDraftDuplicateParams{
		AuthorUsername:  params.AuthorUsername,
		SubjectID:       params.SubjectID,
		EventID:         params.EventID,
		QuestionType:    params.QuestionType,
		Code:            params.Code,
		QuestionText:    params.QuestionText,
		StemHtml:        params.StemHtml,
		StemLatex:       params.StemLatex,
		StimulusHtml:    params.StimulusHtml,
		StimulusLatex:   params.StimulusLatex,
		AnswerKey:       params.AnswerKey,
		Explanation:     params.Explanation,
		ExplanationHtml: params.ExplanationHtml,
		RubricHtml:      params.RubricHtml,
		Difficulty:      string(params.Difficulty),
		AcademicPhase:   params.AcademicPhase,
		TargetLevel:     params.TargetLevel,
		CpRef:           params.CpRef,
		TpRef:           params.TpRef,
		KdRef:           params.KdRef,
		IndicatorRef:    params.IndicatorRef,
		MaterialTopic:   params.MaterialTopic,
		CognitiveLevel:  params.CognitiveLevel,
		HotsFlag:        params.HotsFlag,
		Options:         params.Options,
		MediaAssetIds:   params.MediaAssetIds,
		WorkflowStatus:  params.WorkflowStatus,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.CbtQuestion{}, false, nil
		}
		return db.CbtQuestion{}, false, err
	}
	return row, true, nil
}

func normalizeNoRows(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func (s *CbtQuestion) requireCreateQuestion(ctx context.Context, actor CbtQuestionActor, eventID pgtype.UUID, subjectID pgtype.UUID) error {
	if actor.IsAdmin() {
		return nil
	}
	if !eventID.Valid {
		if actor.HasRole("guru") || actor.HasRole("teacher") || actor.HasPermission("bank_soal.create") {
			return nil
		}
		return domain.ErrForbidden
	}
	members, err := s.q.ListCbtEventMembersByUser(ctx, actor.UserID)
	if err != nil {
		return err
	}
	for _, member := range members {
		if sameUUID(member.EventID, eventID) && member.Role == db.CbtEventMemberRolePembuatSoal && memberSubjectMatches(member.SubjectID, subjectID) {
			return nil
		}
	}
	return domain.ErrForbidden
}

func (s *CbtQuestion) requireModifyQuestion(ctx context.Context, actor CbtQuestionActor, current db.GetCbtQuestionRow) error {
	if actor.IsAdmin() {
		return nil
	}
	if strings.TrimSpace(current.AuthorUsername) == "" || strings.TrimSpace(current.AuthorUsername) != actor.Username {
		return domain.ErrForbidden
	}
	workflowStatus := strings.TrimSpace(current.WorkflowStatus)
	if workflowStatus != "" && workflowStatus != "draft" && workflowStatus != "rejected" {
		if !revisionNeededInlineEditable(current) {
			return fmt.Errorf("%w: soal sedang atau sudah masuk alur review. Duplikat soal untuk membuat revisi baru", domain.ErrConflict)
		}
	}
	if current.Status != "" && current.Status != db.CbtQuestionStatusEnumDraft {
		return fmt.Errorf("%w: soal tidak lagi berstatus draft. Duplikat soal untuk membuat revisi baru", domain.ErrConflict)
	}
	if current.EventID.Valid {
		return s.requireCreateQuestion(ctx, actor, current.EventID, current.SubjectID)
	}
	return nil
}

func revisionNeededInlineEditable(current db.GetCbtQuestionRow) bool {
	workflowStatus := strings.TrimSpace(current.WorkflowStatus)
	status := strings.TrimSpace(string(current.Status))
	return workflowStatus == "revision_needed" &&
		(status == "" || status == string(db.CbtQuestionStatusEnumDraft)) &&
		current.IsLatestVersion &&
		!questionUsageLocked(current.PackageCount, current.AnswerCount)
}

func sameUUID(a, b pgtype.UUID) bool {
	return a.Valid && b.Valid && a.Bytes == b.Bytes
}

func sameOptionalUUID(a, b pgtype.UUID) bool {
	if !a.Valid || !b.Valid {
		return a.Valid == b.Valid
	}
	return a.Bytes == b.Bytes
}

func cbtQuestionUUIDString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", id.Bytes[0:4], id.Bytes[4:6], id.Bytes[6:8], id.Bytes[8:10], id.Bytes[10:16])
}

func logQuestionAudit(ctx context.Context, store cbtQuestionStore, questionID pgtype.UUID, actorUsername string, action string, note string, metadata map[string]any) error {
	action = strings.TrimSpace(action)
	if !questionID.Valid || action == "" {
		return nil
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = store.CreateCbtQuestionAuditLog(ctx, db.CreateCbtQuestionAuditLogParams{
		QuestionID:    questionID,
		ActorUsername: strings.TrimSpace(actorUsername),
		Action:        action,
		Note:          strings.TrimSpace(note),
		Metadata:      raw,
	})
	return err
}

func logQuestionWorkflowEvent(ctx context.Context, store cbtQuestionStore, questionID pgtype.UUID, actor CbtQuestionActor, fromStatus, toStatus string, action string, note string, metadata map[string]any) error {
	action = strings.TrimSpace(action)
	toStatus = strings.TrimSpace(toStatus)
	if !questionID.Valid || action == "" || toStatus == "" {
		return nil
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = store.CreateBankSoalQuestionWorkflowEvent(ctx, db.CreateBankSoalQuestionWorkflowEventParams{
		QuestionID:    questionID,
		ActorUserID:   actor.UserID,
		ActorUsername: strings.TrimSpace(actor.Username),
		FromStatus:    strings.TrimSpace(fromStatus),
		ToStatus:      toStatus,
		Action:        action,
		Note:          strings.TrimSpace(note),
		Metadata:      raw,
	})
	return err
}

func (s *CbtQuestion) validateMediaAssetIDs(ctx context.Context, store cbtQuestionStore, rawIDs []string, questionID pgtype.UUID, actor CbtQuestionActor) error {
	ids := normalizeStringList(rawIDs)
	for _, rawID := range ids {
		assetID, err := parseMediaAssetUUID(rawID)
		if err != nil {
			return fmt.Errorf("%w: media_asset_ids harus berisi UUID valid", domain.ErrBadRequest)
		}
		asset, err := store.GetCbtQuestionAsset(ctx, assetID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: media_asset_ids berisi aset yang tidak ditemukan", domain.ErrBadRequest)
			}
			return err
		}
		if asset.QuestionID.Valid && !sameUUID(asset.QuestionID, questionID) {
			return fmt.Errorf("%w: media_asset_ids berisi aset dari soal lain", domain.ErrBadRequest)
		}
		if !actor.IsAdmin() && strings.TrimSpace(asset.UploadedBy) != actor.Username {
			return fmt.Errorf("%w: media_asset_ids berisi aset yang bukan milik pengguna", domain.ErrForbidden)
		}
	}
	return nil
}

func parseMediaAssetUUID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(strings.TrimSpace(value)); err != nil {
		return pgtype.UUID{}, err
	}
	if !id.Valid {
		return pgtype.UUID{}, fmt.Errorf("uuid kosong")
	}
	return id, nil
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func memberSubjectMatches(memberSubjectID pgtype.UUID, subjectID pgtype.UUID) bool {
	return !memberSubjectID.Valid || sameUUID(memberSubjectID, subjectID)
}

func (s *CbtQuestion) requireWorkflowRole(ctx context.Context, username string, current db.GetCbtQuestionRow, role db.CbtEventMemberRole) error {
	username = strings.TrimSpace(username)
	if !current.EventID.Valid || username == "" {
		return domain.ErrForbidden
	}
	members, err := s.q.ListCbtEventMembersByUsername(ctx, username)
	if err != nil {
		return err
	}
	for _, member := range members {
		if sameUUID(member.EventID, current.EventID) && memberSubjectMatches(member.SubjectID, current.SubjectID) {
			if member.Role == role || member.Role == db.CbtEventMemberRolePanitia {
				return nil
			}
		}
	}
	return domain.ErrForbidden
}

func (s *CbtQuestion) questionDetailAccess(ctx context.Context, actor CbtQuestionActor, row db.GetCbtQuestionDetailRow) (bool, bool, error) {
	if actor.CanReadAllBankSoal() {
		return true, true, nil
	}
	if strings.TrimSpace(row.AuthorUsername) != "" && row.AuthorUsername == actor.Username {
		return true, true, nil
	}
	canReviewScoped, err := s.actorCanSeeBankSoalAnswerMaterial(ctx, actor, row.SubjectID, row.TargetLevel, row.WorkflowStatus)
	if err != nil {
		return false, false, err
	}
	if canReviewScoped {
		return true, true, nil
	}
	member, err := s.actorHasEventQuestionRole(ctx, actor, row.EventID, row.SubjectID, db.CbtEventMemberRoleReviewer, db.CbtEventMemberRolePanitia)
	if err != nil {
		return false, false, err
	}
	if member {
		return true, true, nil
	}
	if row.Status == db.CbtQuestionStatusEnumPublished {
		return true, false, nil
	}
	return false, false, nil
}

func (s *CbtQuestion) actorCanSeeBankSoalAnswerMaterial(ctx context.Context, actor CbtQuestionActor, subjectID pgtype.UUID, targetLevel pgtype.Text, workflowStatus string) (bool, error) {
	if !actor.UserID.Valid {
		return false, nil
	}
	workflowStatus = normalizeWorkflowStatus(workflowStatus)
	if actor.HasPermission("bank_soal.review") && workflowStatusIn(workflowStatus, "submitted", "review", "revision_needed", "reviewed") {
		return s.q.CanBankSoalUserReview(ctx, db.CanBankSoalUserReviewParams{
			UserID:     actor.UserID,
			SubjectID:  subjectID,
			GradeLevel: workflowScopeGradeLevel(targetLevel),
		})
	}
	if (actor.HasPermission("bank_soal.approve") || actor.HasPermission("bank_soal.publish")) && workflowStatusIn(workflowStatus, "reviewed", "approved", "published") {
		return s.q.CanBankSoalUserApprove(ctx, db.CanBankSoalUserApproveParams{
			UserID:     actor.UserID,
			SubjectID:  subjectID,
			GradeLevel: workflowScopeGradeLevel(targetLevel),
		})
	}
	return false, nil
}

func (s *CbtQuestion) requireDuplicateSourceAccess(ctx context.Context, actor CbtQuestionActor, current db.GetCbtQuestionRow) error {
	if actor.IsAdmin() {
		return nil
	}
	if strings.TrimSpace(current.AuthorUsername) != "" && current.AuthorUsername == actor.Username {
		if current.EventID.Valid {
			return s.requireCreateQuestion(ctx, actor, current.EventID, current.SubjectID)
		}
		return nil
	}
	member, err := s.actorHasEventQuestionRole(ctx, actor, current.EventID, current.SubjectID, db.CbtEventMemberRoleReviewer, db.CbtEventMemberRolePanitia)
	if err != nil {
		return err
	}
	if member {
		return nil
	}
	return domain.ErrForbidden
}

func (s *CbtQuestion) actorHasEventQuestionRole(ctx context.Context, actor CbtQuestionActor, eventID, subjectID pgtype.UUID, roles ...db.CbtEventMemberRole) (bool, error) {
	if !eventID.Valid || !actor.UserID.Valid {
		return false, nil
	}
	members, err := s.q.ListCbtEventMembersByUser(ctx, actor.UserID)
	if err != nil {
		return false, err
	}
	for _, member := range members {
		if !sameUUID(member.EventID, eventID) || !memberSubjectMatches(member.SubjectID, subjectID) {
			continue
		}
		for _, role := range roles {
			if member.Role == role {
				return true, nil
			}
		}
	}
	return false, nil
}

func buildCreateQuestionParams(input SaveCbtQuestionInput) (db.CreateCbtQuestionParams, error) {
	normalized, err := normalizeQuestionInput(input)
	if err != nil {
		return db.CreateCbtQuestionParams{}, err
	}
	if normalized.Code == "" {
		// The service layer fills an academic display code after reference validation,
		// so code generation can use the canonical subject code from the database.
	}
	optionsJSON, err := EncodeQuestionOptions(normalized.Options)
	if err != nil {
		return db.CreateCbtQuestionParams{}, err
	}
	mediaJSON, err := EncodeStringArray(normalized.MediaAssetIDs)
	if err != nil {
		return db.CreateCbtQuestionParams{}, err
	}
	return db.CreateCbtQuestionParams{
		EventID:              normalized.EventID,
		SubjectID:            normalized.SubjectID,
		Code:                 normalized.Code,
		QuestionText:         normalized.QuestionText,
		QuestionType:         normalized.QuestionType,
		Options:              optionsJSON,
		OptionA:              normalized.OptionA,
		OptionB:              normalized.OptionB,
		OptionC:              normalized.OptionC,
		OptionD:              normalized.OptionD,
		OptionE:              normalized.OptionE,
		AnswerKey:            normalized.AnswerKey,
		Explanation:          normalized.Explanation,
		Difficulty:           normalized.Difficulty,
		Status:               normalized.Status,
		StemHtml:             normalized.StemHTML,
		StemLatex:            normalized.StemLatex,
		StimulusHtml:         normalized.StimulusHTML,
		StimulusLatex:        normalized.StimulusLatex,
		ExplanationHtml:      normalized.ExplanationHTML,
		RubricHtml:           normalized.RubricHTML,
		AcademicPhase:        normalized.AcademicPhase,
		TargetLevel:          questionTargetLevelText(normalized.TargetLevel),
		CpRef:                normalized.CPRef,
		TpRef:                normalized.TPRef,
		KdRef:                normalized.KDRef,
		IndicatorRef:         normalized.IndicatorRef,
		MaterialTopic:        normalized.MaterialTopic,
		CognitiveLevel:       normalized.CognitiveLevel,
		HotsFlag:             normalized.HotsFlag,
		MediaAssetIds:        mediaJSON,
		WorkflowStatus:       normalized.WorkflowStatus,
		Version:              1,
		VersionGroupID:       normalized.VersionGroupID,
		VersionNumber:        normalizeVersionNumber(normalized.VersionNumber),
		SourceQuestionID:     normalized.SourceQuestionID,
		SupersedesQuestionID: normalized.SupersedesQuestionID,
		IsLatestVersion:      true,
		VersionNote:          normalized.VersionNote,
		AuthorUsername:       normalized.AuthorUsername,
		ReviewerUsername:     normalized.ReviewerUsername,
		ReviewedAt:           normalized.reviewedAt(),
		ApproverUsername:     normalized.ApproverUsername,
		ApprovedAt:           normalized.approvedAt(),
		WriterNotes:          normalized.WriterNotes,
		ReviewNotes:          normalized.ReviewNotes,
	}, nil
}

func buildUpdateQuestionParams(current db.GetCbtQuestionRow, input SaveCbtQuestionInput) (db.UpdateCbtQuestionParams, error) {
	normalized, err := normalizeQuestionInput(input)
	if err != nil {
		return db.UpdateCbtQuestionParams{}, err
	}
	if normalized.Code == "" {
		normalized.Code = strings.TrimSpace(current.Code)
	}
	optionsJSON, err := EncodeQuestionOptions(normalized.Options)
	if err != nil {
		return db.UpdateCbtQuestionParams{}, err
	}
	mediaJSON, err := EncodeStringArray(normalized.MediaAssetIDs)
	if err != nil {
		return db.UpdateCbtQuestionParams{}, err
	}

	reviewer := current.ReviewerUsername
	reviewedAt := current.ReviewedAt
	if workflowStatusSetsReviewer(normalized.WorkflowStatus) {
		reviewer = normalized.ReviewerUsername
		reviewedAt = normalized.reviewedAt()
	}

	approver := current.ApproverUsername
	approvedAt := current.ApprovedAt
	if workflowStatusSetsApprover(normalized.WorkflowStatus) || normalized.Status == db.CbtQuestionStatusEnumPublished {
		approver = normalized.ApproverUsername
		approvedAt = normalized.approvedAt()
	}

	return db.UpdateCbtQuestionParams{
		ID:               input.ID,
		EventID:          normalized.EventID,
		SubjectID:        normalized.SubjectID,
		Code:             normalized.Code,
		QuestionText:     normalized.QuestionText,
		QuestionType:     normalized.QuestionType,
		Options:          optionsJSON,
		OptionA:          normalized.OptionA,
		OptionB:          normalized.OptionB,
		OptionC:          normalized.OptionC,
		OptionD:          normalized.OptionD,
		OptionE:          normalized.OptionE,
		AnswerKey:        normalized.AnswerKey,
		Explanation:      normalized.Explanation,
		Difficulty:       normalized.Difficulty,
		Status:           normalized.Status,
		StemHtml:         normalized.StemHTML,
		StemLatex:        normalized.StemLatex,
		StimulusHtml:     normalized.StimulusHTML,
		StimulusLatex:    normalized.StimulusLatex,
		ExplanationHtml:  normalized.ExplanationHTML,
		RubricHtml:       normalized.RubricHTML,
		AcademicPhase:    normalized.AcademicPhase,
		TargetLevel:      questionTargetLevelText(normalized.TargetLevel),
		CpRef:            normalized.CPRef,
		TpRef:            normalized.TPRef,
		KdRef:            normalized.KDRef,
		IndicatorRef:     normalized.IndicatorRef,
		MaterialTopic:    normalized.MaterialTopic,
		CognitiveLevel:   normalized.CognitiveLevel,
		HotsFlag:         normalized.HotsFlag,
		MediaAssetIds:    mediaJSON,
		WorkflowStatus:   normalized.WorkflowStatus,
		ReviewerUsername: reviewer,
		ReviewedAt:       reviewedAt,
		ApproverUsername: approver,
		ApprovedAt:       approvedAt,
		WriterNotes:      normalized.WriterNotes,
		ReviewNotes:      normalized.ReviewNotes,
	}, nil
}

func generateCbtQuestionCode() string {
	var raw [4]byte
	if _, err := rand.Read(raw[:]); err == nil {
		return "SOAL-" + time.Now().Format("20060102-150405") + "-" + strings.ToUpper(hex.EncodeToString(raw[:]))
	}
	return fmt.Sprintf("SOAL-%s-%d", time.Now().Format("20060102-150405"), time.Now().UnixNano())
}

func normalizeQuestionInput(input SaveCbtQuestionInput) (SaveCbtQuestionInput, error) {
	out := input
	out.AuthoringMode = normalizeAuthoringMode(out.AuthoringMode)
	out.Code = strings.TrimSpace(out.Code)
	out.QuestionType = normalizeQuestionType(out.QuestionType)
	out.QuestionText = strings.TrimSpace(out.QuestionText)
	out.Explanation = strings.TrimSpace(out.Explanation)
	out.StemHTML = sanitizeHTML(out.StemHTML)
	out.StemLatex = strings.TrimSpace(out.StemLatex)
	out.StimulusHTML = sanitizeHTML(out.StimulusHTML)
	out.StimulusLatex = strings.TrimSpace(out.StimulusLatex)
	out.ExplanationHTML = sanitizeHTML(out.ExplanationHTML)
	out.RubricHTML = sanitizeHTML(out.RubricHTML)
	out.AcademicPhase = strings.TrimSpace(out.AcademicPhase)
	if targetLevel, ok := normalizeQuestionTargetLevel(out.TargetLevel); ok {
		out.TargetLevel = targetLevel
	} else {
		return SaveCbtQuestionInput{}, fmt.Errorf("target_level hanya boleh berisi VII, VIII, atau IX")
	}
	if out.TargetLevel == "" {
		out.TargetLevel = ""
	}
	out.CPRef = strings.TrimSpace(out.CPRef)
	out.TPRef = strings.TrimSpace(out.TPRef)
	out.KDRef = strings.TrimSpace(out.KDRef)
	out.IndicatorRef = strings.TrimSpace(out.IndicatorRef)
	out.MaterialTopic = strings.TrimSpace(out.MaterialTopic)
	out.CognitiveLevel = strings.TrimSpace(out.CognitiveLevel)
	out.WorkflowStatus = normalizeWorkflowStatus(out.WorkflowStatus)
	out.VersionNumber = normalizeVersionNumber(out.VersionNumber)
	out.VersionNote = strings.TrimSpace(out.VersionNote)
	out.AuthorUsername = strings.TrimSpace(out.AuthorUsername)
	out.ReviewerUsername = strings.TrimSpace(out.ReviewerUsername)
	out.ApproverUsername = strings.TrimSpace(out.ApproverUsername)
	out.WriterNotes = strings.TrimSpace(out.WriterNotes)
	out.ReviewNotes = strings.TrimSpace(out.ReviewNotes)
	out.MediaAssetIDs = normalizeStringList(out.MediaAssetIDs)

	if out.Difficulty == "" {
		out.Difficulty = db.CbtQuestionDifficultyEnumMedium
	}
	if out.Status == "" {
		out.Status = db.CbtQuestionStatusEnumDraft
	}
	if out.AuthoringMode == "beginner" {
		out.Difficulty = db.CbtQuestionDifficultyEnumMedium
		out.Status = db.CbtQuestionStatusEnumDraft
		if out.WorkflowStatus != "review" && out.WorkflowStatus != "submitted" && out.WorkflowStatus != "revision_needed" {
			out.WorkflowStatus = "draft"
			out.ReviewerUsername = ""
		}
		out.ApproverUsername = ""
		out.WriterNotes = ""
		out.ReviewNotes = ""
	}
	if out.WorkflowStatus == "draft" {
		out.ReviewerUsername = ""
	}
	if out.WorkflowStatus != "approved" && out.WorkflowStatus != "published" {
		out.ApproverUsername = ""
	}
	if out.QuestionText == "" && out.StemHTML != "" {
		out.QuestionText = derivePlainText(out.StemHTML)
	}
	if out.QuestionText == "" && out.StimulusHTML != "" {
		out.QuestionText = derivePlainText(out.StimulusHTML)
	}
	if out.QuestionText == "" && out.StemLatex != "" {
		out.QuestionText = out.StemLatex
	}
	if out.QuestionText == "" {
		return SaveCbtQuestionInput{}, fmt.Errorf("question_text atau stem_html/stem_latex wajib diisi")
	}

	options := out.Options
	if len(options) == 0 {
		options = legacyOptions(out.OptionA, out.OptionB, out.OptionC, out.OptionD, out.OptionE, out.QuestionType)
	}
	if fixedOptions := fixedPairQuestionOptions(out.QuestionType); len(options) == 0 && len(fixedOptions) > 0 {
		options = fixedOptions
	}
	normalizedOptions, err := normalizeOptions(options)
	if err != nil {
		return SaveCbtQuestionInput{}, err
	}
	optionA, optionB, optionC, optionD, optionE := legacyOptionColumns(normalizedOptions)

	out.Options = normalizedOptions
	out.OptionA = optionA
	out.OptionB = optionB
	out.OptionC = optionC
	out.OptionD = optionD
	out.OptionE = optionE
	if out.QuestionType == "short_answer" {
		out.AnswerKey = normalizeShortAnswerKey(out.AnswerKey)
	} else if out.QuestionType == "matching" {
		answerKey, err := normalizeMatchingAnswerKey(out.AnswerKey, countMatchingPairs(normalizedOptions))
		if err != nil {
			return SaveCbtQuestionInput{}, err
		}
		out.AnswerKey = answerKey
	} else if out.QuestionType == "ordering" {
		answerKey, err := normalizeOrderingAnswerKey(out.AnswerKey, len(normalizedOptions))
		if err != nil {
			return SaveCbtQuestionInput{}, err
		}
		out.AnswerKey = answerKey
	} else if out.QuestionType == "multiple_answer" {
		out.AnswerKey = normalizeMultipleAnswerKey(out.AnswerKey)
	} else {
		out.AnswerKey = strings.TrimSpace(strings.ToUpper(out.AnswerKey))
	}

	if err := validateQuestion(out); err != nil {
		return SaveCbtQuestionInput{}, err
	}

	return out, nil
}

func questionTargetLevelText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	return pgtype.Text{String: trimmed, Valid: trimmed != ""}
}

func validateQuestion(input SaveCbtQuestionInput) error {
	requiresCompleteContent := input.WorkflowStatus != "draft" || input.Status != db.CbtQuestionStatusEnumDraft
	if input.AuthoringMode == "beginner" {
		if !beginnerSupportsQuestionType(input.QuestionType) {
			return fmt.Errorf("mode beginner belum mendukung tipe soal ini")
		}
	}

	switch input.QuestionType {
	case "multiple_choice", "single_choice", "multiple_answer", "true_false", "agree_disagree", "ordering":
		if !requiresCompleteContent {
			return nil
		}
		if len(input.Options) < 2 {
			return fmt.Errorf("opsi jawaban minimal 2 untuk tipe soal objektif")
		}
		if input.AuthoringMode == "beginner" && !isFixedPairQuestionType(input.QuestionType) && len(input.Options) < 4 {
			return fmt.Errorf("mode beginner membutuhkan minimal 4 opsi untuk pilihan ganda")
		}
		if input.AnswerKey == "" {
			return fmt.Errorf("answer_key wajib diisi")
		}
		if err := validateObjectiveAnswerKey(input.Options, input.AnswerKey, input.QuestionType); err != nil {
			return err
		}
	case "short_answer":
		if requiresCompleteContent && input.AnswerKey == "" {
			return fmt.Errorf("answer_key wajib diisi untuk short_answer")
		}
	case "matching":
		if !requiresCompleteContent {
			return nil
		}
		if err := validateMatchingQuestion(input.Options, input.AnswerKey); err != nil {
			return err
		}
	case "essay":
		// Rubrik essay bersifat opsional. Guru pembuat soal dapat menilai sendiri
		// tanpa rubrik; UI tetap memberi pengingat kualitas non-blocking.
	default:
		return fmt.Errorf("question_type tidak didukung")
	}

	if input.Status == db.CbtQuestionStatusEnumPublished && input.WorkflowStatus != "approved" && input.WorkflowStatus != "published" {
		return fmt.Errorf("soal hanya boleh dipublish jika workflow_status sudah approved")
	}

	return nil
}

func validateObjectiveAnswerKey(options []QuestionOption, answerKey string, questionType string) error {
	available := make(map[string]bool, len(options))
	for _, option := range options {
		label := strings.TrimSpace(strings.ToUpper(option.Label))
		if label != "" {
			available[label] = true
		}
	}

	keys := []string{answerKey}
	if questionType == "multiple_answer" || questionType == "ordering" {
		keys = strings.Split(answerKey, ",")
	}
	seen := map[string]bool{}
	validKeyCount := 0
	for _, key := range keys {
		key = strings.TrimSpace(strings.ToUpper(key))
		if key == "" || !available[key] {
			return fmt.Errorf("answer_key harus sesuai label opsi yang tersedia")
		}
		if seen[key] {
			return fmt.Errorf("answer_key tidak boleh memuat label duplikat")
		}
		seen[key] = true
		validKeyCount++
	}
	if questionType == "multiple_answer" && validKeyCount < 2 {
		return fmt.Errorf("multiple_answer membutuhkan minimal 2 kunci jawaban")
	}
	if questionType == "ordering" && validKeyCount != len(available) {
		return fmt.Errorf("ordering membutuhkan urutan semua label opsi")
	}
	return nil
}

func validateMatchingQuestion(options []QuestionOption, answerKey string) error {
	pairCount := 0
	leftLabels := make(map[string]bool, len(options))
	rightLabels := make(map[string]bool, len(options))
	correctRightLabels := make(map[string]bool, len(options))
	for _, option := range options {
		left := strings.TrimSpace(strings.ToUpper(option.Label))
		right := strings.TrimSpace(option.MatchLabel)
		if option.IsDistractor {
			if right == "" || matchingOptionContent(option) == "" {
				return fmt.Errorf("distraktor menjodohkan wajib memiliki label dan teks kanan")
			}
			if rightLabels[right] {
				return fmt.Errorf("label pasangan menjodohkan tidak boleh duplikat")
			}
			rightLabels[right] = true
			continue
		}
		if left == "" || right == "" || optionContent(option) == "" || matchingOptionContent(option) == "" {
			return fmt.Errorf("setiap pasangan menjodohkan wajib memiliki kolom kiri dan kanan")
		}
		if leftLabels[left] || rightLabels[right] {
			return fmt.Errorf("label pasangan menjodohkan tidak boleh duplikat")
		}
		leftLabels[left] = true
		rightLabels[right] = true
		correctRightLabels[right] = true
		pairCount++
	}
	if pairCount < 2 {
		return fmt.Errorf("menjodohkan membutuhkan minimal 2 pasangan")
	}
	remainingRightLabels := make(map[string]bool, len(correctRightLabels))
	for right := range correctRightLabels {
		remainingRightLabels[right] = true
	}
	for _, pair := range strings.Split(answerKey, ";") {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return fmt.Errorf("answer_key menjodohkan tidak valid")
		}
		left := strings.TrimSpace(strings.ToUpper(parts[0]))
		right := strings.TrimSpace(parts[1])
		if !leftLabels[left] || !remainingRightLabels[right] {
			return fmt.Errorf("answer_key menjodohkan harus sesuai label pasangan")
		}
		delete(leftLabels, left)
		delete(remainingRightLabels, right)
	}
	if len(leftLabels) != 0 || len(remainingRightLabels) != 0 {
		return fmt.Errorf("answer_key menjodohkan harus memetakan semua pasangan")
	}
	return nil
}

func optionContent(option QuestionOption) string {
	return strings.TrimSpace(firstQuestionOptionContent(option.Text, option.HTML, option.Latex))
}

func matchingOptionContent(option QuestionOption) string {
	return strings.TrimSpace(firstQuestionOptionContent(option.MatchText, option.MatchHTML))
}

func firstQuestionOptionContent(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func normalizeShortAnswerKey(value string) string {
	aliases := shortAnswerAliases(value)
	for i, alias := range aliases {
		aliases[i] = strings.Join(strings.Fields(strings.ReplaceAll(alias, "\u00a0", " ")), " ")
	}
	return strings.Join(aliases, "|")
}

func normalizeMatchingAnswerKey(value string, optionCount int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if optionCount <= 0 {
		return "", fmt.Errorf("answer_key menjodohkan tidak valid")
	}
	pairs := strings.Split(value, ";")
	labels := make(map[string]bool, optionCount)
	matches := make(map[string]string, optionCount)
	for i := 0; i < optionCount; i++ {
		labels[string(rune('A'+i))] = true
	}
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return "", fmt.Errorf("answer_key menjodohkan tidak valid")
		}
		left := strings.TrimSpace(strings.ToUpper(parts[0]))
		right := strings.TrimSpace(parts[1])
		if !labels[left] {
			return "", fmt.Errorf("answer_key menjodohkan harus sesuai label pasangan")
		}
		rightIndex, err := strconv.Atoi(right)
		if err != nil || rightIndex < 1 || rightIndex > optionCount {
			return "", fmt.Errorf("answer_key menjodohkan harus sesuai label pasangan")
		}
		matches[left] = strconv.Itoa(rightIndex)
	}
	if len(matches) != optionCount {
		return "", fmt.Errorf("answer_key menjodohkan harus memetakan semua pasangan")
	}
	ordered := make([]string, 0, optionCount)
	for i := 0; i < optionCount; i++ {
		left := string(rune('A' + i))
		ordered = append(ordered, fmt.Sprintf("%s=%s", left, matches[left]))
	}
	return strings.Join(ordered, ";"), nil
}

func buildMatchingAnswerKey(optionCount int) string {
	pairs := make([]string, 0, optionCount)
	for i := 0; i < optionCount; i++ {
		pairs = append(pairs, fmt.Sprintf("%s=%d", string(rune('A'+i)), i+1))
	}
	return strings.Join(pairs, ";")
}

func normalizeMultipleAnswerKey(value string) string {
	parts := strings.Split(value, ",")
	seen := map[string]bool{}
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		label := strings.TrimSpace(strings.ToUpper(part))
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return strings.Join(labels, ",")
}

func normalizeOrderingAnswerKey(value string, optionCount int) (string, error) {
	labels := make(map[string]bool, optionCount)
	for i := 0; i < optionCount; i++ {
		labels[string(rune('A'+i))] = true
	}
	ordered := make([]string, 0, optionCount)
	seen := map[string]bool{}
	for _, part := range strings.Split(value, ",") {
		label := strings.TrimSpace(strings.ToUpper(part))
		if label == "" {
			continue
		}
		if !labels[label] || seen[label] {
			return "", fmt.Errorf("answer_key ordering harus memuat tiap label opsi tepat satu kali")
		}
		seen[label] = true
		ordered = append(ordered, label)
	}
	if len(ordered) != optionCount {
		return "", fmt.Errorf("answer_key ordering harus memuat semua label opsi")
	}
	return strings.Join(ordered, ","), nil
}

func countMatchingPairs(options []QuestionOption) int {
	count := 0
	for _, option := range options {
		if !option.IsDistractor {
			count++
		}
	}
	return count
}

func shortAnswerAliases(value string) []string {
	parts := strings.Split(value, "|")
	seen := make(map[string]bool, len(parts))
	aliases := make([]string, 0, len(parts))
	for _, part := range parts {
		alias := strings.TrimSpace(strings.ReplaceAll(part, "\u00a0", " "))
		if alias == "" {
			continue
		}
		normalized := normalizeShortAnswerComparable(alias)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		aliases = append(aliases, alias)
	}
	return aliases
}

func normalizeShortAnswerComparable(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.ReplaceAll(value, "\u00a0", " ")), " "))
}

func beginnerSupportsQuestionType(questionType string) bool {
	switch questionType {
	case "multiple_choice", "multiple_answer", "true_false", "agree_disagree", "matching", "ordering", "short_answer", "essay":
		return true
	default:
		return false
	}
}

func normalizeQuestionType(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "", "multiple_choice", "single_choice":
		return "multiple_choice"
	case "multiple_answer", "true_false", "agree_disagree", "matching", "ordering", "short_answer", "essay":
		return normalized
	default:
		return normalized
	}
}

func fixedPairQuestionOptions(questionType string) []QuestionOption {
	switch questionType {
	case "true_false":
		return []QuestionOption{
			{Label: "A", Text: "Benar"},
			{Label: "B", Text: "Salah"},
		}
	case "agree_disagree":
		return []QuestionOption{
			{Label: "A", Text: "Setuju"},
			{Label: "B", Text: "Tidak Setuju"},
		}
	default:
		return nil
	}
}

func isFixedPairQuestionType(questionType string) bool {
	switch questionType {
	case "true_false", "agree_disagree":
		return true
	default:
		return false
	}
}

func normalizeAuthoringMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "advance":
		return "advance"
	default:
		return "beginner"
	}
}

func normalizeWorkflowStatus(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "draft":
		return "draft"
	case "review", "submitted", "revision_needed", "reviewed", "approved", "published", "rejected", "archived":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "draft"
	}
}

func workflowStatusSetsReviewer(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "revision_needed", "reviewed", "approved", "rejected":
		return true
	default:
		return false
	}
}

func workflowStatusSetsApprover(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "approved", "published":
		return true
	default:
		return false
	}
}

func normalizeVersionNumber(value int32) int32 {
	if value < 1 {
		return 1
	}
	return value
}

func normalizeOptions(options []QuestionOption) ([]QuestionOption, error) {
	normalized := make([]QuestionOption, 0, len(options))
	for idx, option := range options {
		label := strings.TrimSpace(strings.ToUpper(option.Label))
		if label == "" && !option.IsDistractor {
			label = string(rune('A' + idx))
		}
		text := strings.TrimSpace(option.Text)
		htmlText := strings.TrimSpace(option.HTML)
		latex := strings.TrimSpace(option.Latex)
		matchLabel := strings.TrimSpace(option.MatchLabel)
		if matchLabel == "" && (strings.TrimSpace(option.MatchText) != "" || strings.TrimSpace(option.MatchHTML) != "") {
			matchLabel = fmt.Sprintf("%d", idx+1)
		}
		matchText := strings.TrimSpace(option.MatchText)
		matchHTML := strings.TrimSpace(option.MatchHTML)
		if text == "" && htmlText == "" && latex == "" && matchText == "" && matchHTML == "" {
			continue
		}
		normalized = append(normalized, QuestionOption{
			Label:        label,
			Text:         text,
			HTML:         sanitizeHTML(htmlText),
			Latex:        latex,
			AssetID:      strings.TrimSpace(option.AssetID),
			MatchLabel:   matchLabel,
			MatchText:    matchText,
			MatchHTML:    sanitizeHTML(matchHTML),
			IsDistractor: option.IsDistractor,
		})
	}
	return normalized, nil
}

func legacyOptions(a, b, c, d, e, questionType string) []QuestionOption {
	raw := []string{a, b, c, d, e}
	options := make([]QuestionOption, 0, len(raw))
	for idx, value := range raw {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		options = append(options, QuestionOption{
			Label: string(rune('A' + idx)),
			Text:  value,
		})
	}
	return options
}

func legacyOptionColumns(options []QuestionOption) (string, string, string, string, string) {
	values := [5]string{}
	for i := 0; i < len(options) && i < 5; i++ {
		if options[i].Text != "" {
			values[i] = options[i].Text
			continue
		}
		if options[i].HTML != "" {
			values[i] = derivePlainText(options[i].HTML)
			continue
		}
		values[i] = options[i].Latex
	}
	return values[0], values[1], values[2], values[3], values[4]
}

func derivePlainText(value string) string {
	plain := plainTextFromHTML(value)
	plain = html.UnescapeString(strings.TrimSpace(plain))
	plain = strings.Join(strings.Fields(plain), " ")
	if len(plain) > 500 {
		return plain[:500]
	}
	return plain
}

func plainTextFromHTML(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	nodes, err := nethtml.ParseFragment(strings.NewReader(value), &nethtml.Node{
		Type: nethtml.ElementNode,
		Data: "body",
	})
	if err != nil {
		return bankSoalPlainTextPolicy.Sanitize(value)
	}
	var builder strings.Builder
	for _, node := range nodes {
		appendHTMLText(&builder, node)
	}
	return builder.String()
}

func appendHTMLText(builder *strings.Builder, node *nethtml.Node) {
	if node == nil {
		return
	}
	switch node.Type {
	case nethtml.TextNode:
		builder.WriteString(node.Data)
		builder.WriteByte(' ')
	case nethtml.ElementNode:
		if node.Data == "script" || node.Data == "style" || node.Data == "svg" {
			return
		}
		if node.Data == "br" {
			builder.WriteByte(' ')
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		appendHTMLText(builder, child)
	}
	if node.Type == nethtml.ElementNode && isPlainTextBlockElement(node.Data) {
		builder.WriteByte(' ')
	}
}

func isPlainTextBlockElement(tag string) bool {
	switch tag {
	case "address", "article", "aside", "blockquote", "br", "caption", "div", "figcaption", "figure",
		"footer", "h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "li", "main", "ol", "p",
		"pre", "section", "table", "tbody", "td", "tfoot", "th", "thead", "tr", "ul":
		return true
	default:
		return false
	}
}

func (s SaveCbtQuestionInput) reviewedAt() pgtype.Timestamptz {
	if s.ReviewerUsername == "" {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: time.Now(), Valid: true}
}

func (s SaveCbtQuestionInput) approvedAt() pgtype.Timestamptz {
	if s.ApproverUsername == "" {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: time.Now(), Valid: true}
}

func EncodeQuestionOptions(options []QuestionOption) ([]byte, error) {
	if len(options) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(options)
}

func EncodeStringArray(values []string) ([]byte, error) {
	if len(values) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(values)
}

func sanitizeHTML(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.TrimSpace(bankSoalHTMLPolicy.Sanitize(value))
}

func mergeNotes(existing string, incoming string) string {
	incoming = strings.TrimSpace(incoming)
	if incoming == "" {
		return existing
	}
	return incoming
}

func cbtQuestionFromCurrent(current db.GetCbtQuestionRow) db.CbtQuestion {
	return db.CbtQuestion{
		ID:                   current.ID,
		EventID:              current.EventID,
		SubjectID:            current.SubjectID,
		Code:                 current.Code,
		QuestionText:         current.QuestionText,
		QuestionType:         current.QuestionType,
		Options:              current.Options,
		OptionA:              current.OptionA,
		OptionB:              current.OptionB,
		OptionC:              current.OptionC,
		OptionD:              current.OptionD,
		OptionE:              current.OptionE,
		AnswerKey:            current.AnswerKey,
		Explanation:          current.Explanation,
		Difficulty:           current.Difficulty,
		Status:               current.Status,
		CreatedAt:            current.CreatedAt,
		UpdatedAt:            current.UpdatedAt,
		StemHtml:             current.StemHtml,
		StemLatex:            current.StemLatex,
		StimulusHtml:         current.StimulusHtml,
		StimulusLatex:        current.StimulusLatex,
		ExplanationHtml:      current.ExplanationHtml,
		RubricHtml:           current.RubricHtml,
		AcademicPhase:        current.AcademicPhase,
		TargetLevel:          current.TargetLevel,
		CpRef:                current.CpRef,
		TpRef:                current.TpRef,
		KdRef:                current.KdRef,
		IndicatorRef:         current.IndicatorRef,
		MaterialTopic:        current.MaterialTopic,
		CognitiveLevel:       current.CognitiveLevel,
		HotsFlag:             current.HotsFlag,
		MediaAssetIds:        current.MediaAssetIds,
		WorkflowStatus:       current.WorkflowStatus,
		Version:              current.Version,
		VersionGroupID:       current.VersionGroupID,
		VersionNumber:        current.VersionNumber,
		SourceQuestionID:     current.SourceQuestionID,
		SupersedesQuestionID: current.SupersedesQuestionID,
		IsLatestVersion:      current.IsLatestVersion,
		VersionNote:          current.VersionNote,
		AuthorUsername:       current.AuthorUsername,
		ReviewerUsername:     current.ReviewerUsername,
		ReviewedAt:           current.ReviewedAt,
		ApproverUsername:     current.ApproverUsername,
		ApprovedAt:           current.ApprovedAt,
		WriterNotes:          current.WriterNotes,
		ReviewNotes:          current.ReviewNotes,
	}
}

func questionInputFromCurrent(current db.GetCbtQuestionRow, username string) SaveCbtQuestionInput {
	return SaveCbtQuestionInput{
		ID:                   current.ID,
		EventID:              current.EventID,
		SubjectID:            current.SubjectID,
		AuthoringMode:        "advance",
		Code:                 current.Code,
		QuestionText:         current.QuestionText,
		QuestionType:         current.QuestionType,
		Options:              decodeQuestionOptions(current.Options),
		OptionA:              current.OptionA,
		OptionB:              current.OptionB,
		OptionC:              current.OptionC,
		OptionD:              current.OptionD,
		OptionE:              current.OptionE,
		AnswerKey:            current.AnswerKey,
		Explanation:          current.Explanation,
		Difficulty:           current.Difficulty,
		Status:               current.Status,
		StemHTML:             current.StemHtml,
		StemLatex:            current.StemLatex,
		StimulusHTML:         current.StimulusHtml,
		StimulusLatex:        current.StimulusLatex,
		ExplanationHTML:      current.ExplanationHtml,
		RubricHTML:           current.RubricHtml,
		AcademicPhase:        current.AcademicPhase,
		TargetLevel:          current.TargetLevel.String,
		CPRef:                current.CpRef,
		TPRef:                current.TpRef,
		KDRef:                current.KdRef,
		IndicatorRef:         current.IndicatorRef,
		MaterialTopic:        current.MaterialTopic,
		CognitiveLevel:       current.CognitiveLevel,
		HotsFlag:             current.HotsFlag,
		MediaAssetIDs:        decodeStringArray(current.MediaAssetIds),
		WorkflowStatus:       current.WorkflowStatus,
		VersionGroupID:       current.VersionGroupID,
		VersionNumber:        normalizeVersionNumber(current.VersionNumber),
		SourceQuestionID:     current.SourceQuestionID,
		SupersedesQuestionID: current.SupersedesQuestionID,
		IsLatestVersion:      current.IsLatestVersion,
		VersionNote:          current.VersionNote,
		AuthorUsername:       current.AuthorUsername,
		ReviewerUsername:     username,
		ApproverUsername:     username,
		WriterNotes:          current.WriterNotes,
		ReviewNotes:          current.ReviewNotes,
	}
}

func suggestQuestionAuthoringMode(questionType, stemLatex, stimulusLatex, academicPhase, cpRef, tpRef, kdRef, indicatorRef, materialTopic, cognitiveLevel string, hotsFlag bool, workflowStatus, writerNotes, reviewNotes, rubricHTML string) string {
	if !beginnerSupportsQuestionType(questionType) {
		return "advance"
	}
	if strings.TrimSpace(stemLatex) != "" || strings.TrimSpace(stimulusLatex) != "" {
		return "advance"
	}
	if strings.TrimSpace(academicPhase) != "" || strings.TrimSpace(cpRef) != "" || strings.TrimSpace(tpRef) != "" ||
		strings.TrimSpace(kdRef) != "" || strings.TrimSpace(indicatorRef) != "" || strings.TrimSpace(materialTopic) != "" ||
		strings.TrimSpace(cognitiveLevel) != "" || hotsFlag {
		return "advance"
	}
	if strings.TrimSpace(workflowStatus) != "" && strings.TrimSpace(workflowStatus) != "draft" {
		return "advance"
	}
	if strings.TrimSpace(writerNotes) != "" || strings.TrimSpace(reviewNotes) != "" || strings.TrimSpace(rubricHTML) != "" {
		return "advance"
	}
	return "beginner"
}

func decodeQuestionOptions(raw []byte) []QuestionOption {
	if len(raw) == 0 {
		return nil
	}
	var options []QuestionOption
	if err := json.Unmarshal(raw, &options); err != nil {
		return nil
	}
	return options
}

func decodeStringArray(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	return values
}
