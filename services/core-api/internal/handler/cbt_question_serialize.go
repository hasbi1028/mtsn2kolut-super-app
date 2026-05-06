package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func cbtQuestionUsageLocked(packageCount, answerCount int32) bool {
	return packageCount > 0 || answerCount > 0
}

func cbtQuestionUsageMap(packageCount, answerCount int32) map[string]any {
	return map[string]any{
		"package_count": packageCount,
		"answer_count":  answerCount,
		"is_locked":     cbtQuestionUsageLocked(packageCount, answerCount),
	}
}

func parseBoolFormValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func hasAnyRole(r *http.Request, allowed ...string) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, allowed...)
}

func cbtAuditAuthoringEvent(audit cbtAuthoringAuditWriter, ctx context.Context, action, entityType, entityID string, extra map[string]any) {
	if audit == nil {
		return
	}
	claims, ok := api.ClaimsFromContext(ctx)
	if !ok {
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		return
	}
	meta := map[string]any{
		"username":   claims["usr"],
		"user_id":    claims["uid"],
		"session_id": claims["ssid"],
	}
	for key, value := range extra {
		meta[key] = value
	}
	rawMeta, err := json.Marshal(meta)
	if err != nil {
		return
	}
	_, _ = audit.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   rawMeta,
	})
}

func serializeQuestionSummary(summary service.CbtQuestionSummary) map[string]any {
	bySubject := make([]map[string]any, 0, len(summary.BySubject))
	for _, row := range summary.BySubject {
		bySubject = append(bySubject, map[string]any{
			"subject_id":   pgUUIDString(row.SubjectID),
			"subject_name": row.SubjectName,
			"subject_code": row.SubjectCode,
			"total":        row.Total,
		})
	}
	byCognitive := make([]map[string]any, 0, len(summary.ByCognitiveLevel))
	for _, row := range summary.ByCognitiveLevel {
		byCognitive = append(byCognitive, map[string]any{
			"cognitive_level": row.CognitiveLevel,
			"total":           row.Total,
		})
	}
	recent := make([]map[string]any, 0, len(summary.Recent))
	for _, row := range summary.Recent {
		recent = append(recent, map[string]any{
			"id":                pgUUIDString(row.ID),
			"code":              row.Code,
			"subject_id":        pgUUIDString(row.SubjectID),
			"subject_name":      row.SubjectName,
			"subject_code":      row.SubjectCode,
			"material_topic":    row.MaterialTopic,
			"workflow_status":   row.WorkflowStatus,
			"status":            row.Status,
			"author_username":   row.AuthorUsername,
			"reviewer_username": row.ReviewerUsername,
			"created_at":        row.CreatedAt,
			"updated_at":        row.UpdatedAt,
		})
	}
	return map[string]any{
		"counts": map[string]any{
			"all":           summary.Counts.Total,
			"draft":         summary.Counts.Draft,
			"review":        summary.Counts.Review,
			"rejected":      summary.Counts.Rejected,
			"approved":      summary.Counts.Approved,
			"published":     summary.Counts.Published,
			"package_usage": summary.Counts.PackageUsage,
		},
		"by_subject":         bySubject,
		"by_cognitive_level": byCognitive,
		"recent":             recent,
	}
}

func serializeQuestionListRow(row db.ListCbtQuestionsFilteredRow, includeAnswerKey ...bool) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	return map[string]any{
		"id":                pgUUIDString(row.ID),
		"event_id":          pgUUIDString(row.EventID),
		"authoring_mode":    suggestedMode,
		"suggested_mode":    suggestedMode,
		"subject_id":        pgUUIDString(row.SubjectID),
		"subject_name":      row.SubjectName,
		"subject_code":      row.SubjectCode,
		"code":              row.Code,
		"question_text":     row.QuestionText,
		"question_type":     row.QuestionType,
		"options":           decodeJSONBytes(row.Options),
		"option_a":          row.OptionA,
		"option_b":          row.OptionB,
		"option_c":          row.OptionC,
		"option_d":          row.OptionD,
		"option_e":          row.OptionE,
		"answer_key":        row.AnswerKey,
		"explanation":       row.Explanation,
		"difficulty":        row.Difficulty,
		"status":            row.Status,
		"created_at":        row.CreatedAt,
		"updated_at":        row.UpdatedAt,
		"stem_html":         row.StemHtml,
		"stem_latex":        row.StemLatex,
		"stimulus_html":     row.StimulusHtml,
		"stimulus_latex":    row.StimulusLatex,
		"explanation_html":  row.ExplanationHtml,
		"rubric_html":       row.RubricHtml,
		"academic_phase":    row.AcademicPhase,
		"grade_level":       nullableInt(row.GradeLevel),
		"cp_ref":            row.CpRef,
		"tp_ref":            row.TpRef,
		"kd_ref":            row.KdRef,
		"indicator_ref":     row.IndicatorRef,
		"material_topic":    row.MaterialTopic,
		"cognitive_level":   row.CognitiveLevel,
		"hots_flag":         row.HotsFlag,
		"media_asset_ids":   decodeJSONBytes(row.MediaAssetIds),
		"workflow_status":   row.WorkflowStatus,
		"version":           row.Version,
		"author_username":   row.AuthorUsername,
		"reviewer_username": row.ReviewerUsername,
		"reviewed_at":       row.ReviewedAt,
		"approver_username": row.ApproverUsername,
		"approved_at":       row.ApprovedAt,
		"writer_notes":      row.WriterNotes,
		"review_notes":      row.ReviewNotes,
		"package_count":     row.PackageCount,
		"answer_count":      row.AnswerCount,
		"is_locked":         cbtQuestionUsageLocked(row.PackageCount, row.AnswerCount),
		"usage":             cbtQuestionUsageMap(row.PackageCount, row.AnswerCount),
	}
}

func questionAnswerKeyAllowed(r *http.Request, authorUsername string) bool {
	if adminAccessAllowed(r) {
		return true
	}
	return strings.TrimSpace(currentUsername(r)) != "" && strings.TrimSpace(currentUsername(r)) == strings.TrimSpace(authorUsername)
}

func serializeQuestionDetailRow(row db.GetCbtQuestionDetailRow, includeAnswerKey bool) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	answerKey := ""
	if includeAnswerKey {
		answerKey = row.AnswerKey
	}
	return map[string]any{
		"id":                pgUUIDString(row.ID),
		"event_id":          pgUUIDString(row.EventID),
		"authoring_mode":    suggestedMode,
		"suggested_mode":    suggestedMode,
		"subject_id":        pgUUIDString(row.SubjectID),
		"subject_name":      row.SubjectName,
		"subject_code":      row.SubjectCode,
		"code":              row.Code,
		"question_text":     row.QuestionText,
		"question_type":     row.QuestionType,
		"options":           decodeJSONBytes(row.Options),
		"option_a":          row.OptionA,
		"option_b":          row.OptionB,
		"option_c":          row.OptionC,
		"option_d":          row.OptionD,
		"option_e":          row.OptionE,
		"answer_key":        answerKey,
		"explanation":       row.Explanation,
		"difficulty":        row.Difficulty,
		"status":            row.Status,
		"created_at":        row.CreatedAt,
		"updated_at":        row.UpdatedAt,
		"stem_html":         row.StemHtml,
		"stem_latex":        row.StemLatex,
		"stimulus_html":     row.StimulusHtml,
		"stimulus_latex":    row.StimulusLatex,
		"explanation_html":  row.ExplanationHtml,
		"rubric_html":       row.RubricHtml,
		"academic_phase":    row.AcademicPhase,
		"grade_level":       nullableInt(row.GradeLevel),
		"cp_ref":            row.CpRef,
		"tp_ref":            row.TpRef,
		"kd_ref":            row.KdRef,
		"indicator_ref":     row.IndicatorRef,
		"material_topic":    row.MaterialTopic,
		"cognitive_level":   row.CognitiveLevel,
		"hots_flag":         row.HotsFlag,
		"media_asset_ids":   decodeJSONBytes(row.MediaAssetIds),
		"workflow_status":   row.WorkflowStatus,
		"version":           row.Version,
		"author_username":   row.AuthorUsername,
		"reviewer_username": row.ReviewerUsername,
		"reviewed_at":       row.ReviewedAt,
		"approver_username": row.ApproverUsername,
		"approved_at":       row.ApprovedAt,
		"writer_notes":      row.WriterNotes,
		"review_notes":      row.ReviewNotes,
		"package_count":     row.PackageCount,
		"answer_count":      row.AnswerCount,
		"is_locked":         cbtQuestionUsageLocked(row.PackageCount, row.AnswerCount),
		"usage":             cbtQuestionUsageMap(row.PackageCount, row.AnswerCount),
	}
}

func serializeQuestionModel(row db.CbtQuestion) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	return map[string]any{
		"id":                pgUUIDString(row.ID),
		"event_id":          pgUUIDString(row.EventID),
		"authoring_mode":    suggestedMode,
		"suggested_mode":    suggestedMode,
		"subject_id":        pgUUIDString(row.SubjectID),
		"code":              row.Code,
		"question_text":     row.QuestionText,
		"question_type":     row.QuestionType,
		"options":           decodeJSONBytes(row.Options),
		"option_a":          row.OptionA,
		"option_b":          row.OptionB,
		"option_c":          row.OptionC,
		"option_d":          row.OptionD,
		"option_e":          row.OptionE,
		"answer_key":        row.AnswerKey,
		"explanation":       row.Explanation,
		"difficulty":        row.Difficulty,
		"status":            row.Status,
		"created_at":        row.CreatedAt,
		"updated_at":        row.UpdatedAt,
		"stem_html":         row.StemHtml,
		"stem_latex":        row.StemLatex,
		"stimulus_html":     row.StimulusHtml,
		"stimulus_latex":    row.StimulusLatex,
		"explanation_html":  row.ExplanationHtml,
		"rubric_html":       row.RubricHtml,
		"academic_phase":    row.AcademicPhase,
		"grade_level":       nullableInt(row.GradeLevel),
		"cp_ref":            row.CpRef,
		"tp_ref":            row.TpRef,
		"kd_ref":            row.KdRef,
		"indicator_ref":     row.IndicatorRef,
		"material_topic":    row.MaterialTopic,
		"cognitive_level":   row.CognitiveLevel,
		"hots_flag":         row.HotsFlag,
		"media_asset_ids":   decodeJSONBytes(row.MediaAssetIds),
		"workflow_status":   row.WorkflowStatus,
		"version":           row.Version,
		"author_username":   row.AuthorUsername,
		"reviewer_username": row.ReviewerUsername,
		"reviewed_at":       row.ReviewedAt,
		"approver_username": row.ApproverUsername,
		"approved_at":       row.ApprovedAt,
		"writer_notes":      row.WriterNotes,
		"review_notes":      row.ReviewNotes,
	}
}

func decodeJSONBytes(raw []byte) any {
	if len(raw) == 0 {
		return []any{}
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return []any{}
	}
	return v
}

func nullableInt(value pgtype.Int2) any {
	if !value.Valid {
		return nil
	}
	return value.Int16
}

func serviceAuthoringModeFromRow(questionType, stemLatex, stimulusLatex, academicPhase, cpRef, tpRef, kdRef, indicatorRef, materialTopic, cognitiveLevel string, hotsFlag bool, workflowStatus, writerNotes, reviewNotes, rubricHTML string) string {
	if questionType != "multiple_choice" && questionType != "essay" {
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
