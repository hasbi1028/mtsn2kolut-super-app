package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtQuestion struct {
	svc *service.CbtQuestion
}

func NewCbtQuestion(svc *service.CbtQuestion) *CbtQuestion { return &CbtQuestion{svc: svc} }

type cbtQuestionBody struct {
	SubjectID       string                   `json:"subject_id"`
	AuthoringMode   string                   `json:"authoring_mode"`
	Code            string                   `json:"code"`
	QuestionText    string                   `json:"question_text"`
	QuestionType    string                   `json:"question_type"`
	Options         []service.QuestionOption `json:"options"`
	OptionA         string                   `json:"option_a"`
	OptionB         string                   `json:"option_b"`
	OptionC         string                   `json:"option_c"`
	OptionD         string                   `json:"option_d"`
	OptionE         string                   `json:"option_e"`
	AnswerKey       string                   `json:"answer_key"`
	Explanation     string                   `json:"explanation"`
	Difficulty      string                   `json:"difficulty"`
	Status          string                   `json:"status"`
	StemHTML        string                   `json:"stem_html"`
	StemLatex       string                   `json:"stem_latex"`
	StimulusHTML    string                   `json:"stimulus_html"`
	StimulusLatex   string                   `json:"stimulus_latex"`
	ExplanationHTML string                   `json:"explanation_html"`
	RubricHTML      string                   `json:"rubric_html"`
	AcademicPhase   string                   `json:"academic_phase"`
	GradeLevel      int16                    `json:"grade_level"`
	CPRef           string                   `json:"cp_ref"`
	TPRef           string                   `json:"tp_ref"`
	KDRef           string                   `json:"kd_ref"`
	IndicatorRef    string                   `json:"indicator_ref"`
	MaterialTopic   string                   `json:"material_topic"`
	CognitiveLevel  string                   `json:"cognitive_level"`
	HotsFlag        bool                     `json:"hots_flag"`
	MediaAssetIDs   []string                 `json:"media_asset_ids"`
	WorkflowStatus  string                   `json:"workflow_status"`
	WriterNotes     string                   `json:"writer_notes"`
	ReviewNotes     string                   `json:"review_notes"`
}

func (h *CbtQuestion) List(w http.ResponseWriter, r *http.Request) {
	subjectID := pgtype.UUID{}
	if raw := strings.TrimSpace(r.URL.Query().Get("subject_id")); raw != "" {
		parsed, err := parseUUID(raw)
		if err != nil {
			api.BadRequest(w, "subject_id invalid")
			return
		}
		subjectID = parsed
	}
	limit := int32(25)
	offset := int32(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		var parsed int
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		var parsed int
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}
	rows, total, err := h.svc.ListFiltered(r.Context(), service.ListCbtQuestionsInput{
		SubjectID:      subjectID,
		WorkflowStatus: r.URL.Query().Get("workflow_status"),
		QuestionType:   r.URL.Query().Get("question_type"),
		HotsFilter:     r.URL.Query().Get("hots"),
		SearchQuery:    r.URL.Query().Get("q"),
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, serializeQuestionListRow(row))
	}
	api.OK(w, map[string]any{
		"items": items,
		"meta": map[string]any{
			"total": total,
			"limit": limit,
			"offset": offset,
		},
	})
}

func (h *CbtQuestion) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.GetDetail(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, serializeQuestionDetailRow(row))
}

func (h *CbtQuestion) Create(w http.ResponseWriter, r *http.Request) {
	body, err := decodeQuestionBody(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	input, err := questionInputFromBody(r, body, pgtype.UUID{})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	row, err := h.svc.Create(r.Context(), input)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, serializeQuestionModel(row))
}

func (h *CbtQuestion) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	body, err := decodeQuestionBody(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	input, err := questionInputFromBody(r, body, id)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	row, err := h.svc.Update(r.Context(), input)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, serializeQuestionModel(row))
}

func (h *CbtQuestion) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *CbtQuestion) WorkflowAction(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Action string `json:"action"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	username := currentUsername(r)
	role := currentRole(r)
	switch strings.TrimSpace(body.Action) {
	case "submit_review":
		row, err := h.svc.SubmitReview(r.Context(), id, username, body.Notes)
		if err != nil {
			api.BadRequest(w, err.Error())
			return
		}
		api.OK(w, serializeQuestionModel(row))
	case "approve":
		if role != "admin" {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.Approve(r.Context(), id, username, body.Notes)
		if err != nil {
			api.BadRequest(w, err.Error())
			return
		}
		api.OK(w, serializeQuestionModel(row))
	case "publish":
		if role != "admin" {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.Publish(r.Context(), id, username)
		if err != nil {
			api.BadRequest(w, err.Error())
			return
		}
		api.OK(w, serializeQuestionModel(row))
	case "archive":
		if role != "admin" {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.Archive(r.Context(), id, username)
		if err != nil {
			api.BadRequest(w, err.Error())
			return
		}
		api.OK(w, serializeQuestionModel(row))
	default:
		api.BadRequest(w, "action tidak didukung")
	}
}

func (h *CbtQuestion) Duplicate(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.DuplicateAsDraft(r.Context(), id, currentUsername(r))
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, serializeQuestionModel(row))
}

func decodeQuestionBody(r *http.Request) (cbtQuestionBody, error) {
	var body cbtQuestionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return cbtQuestionBody{}, err
	}
	return body, nil
}

func questionInputFromBody(r *http.Request, body cbtQuestionBody, id pgtype.UUID) (service.SaveCbtQuestionInput, error) {
	subjectID, err := parseUUID(body.SubjectID)
	if err != nil {
		return service.SaveCbtQuestionInput{}, err
	}

	gradeLevel := pgtype.Int2{}
	if body.GradeLevel > 0 {
		gradeLevel = pgtype.Int2{Int16: body.GradeLevel, Valid: true}
	}

	username := currentUsername(r)
	return service.SaveCbtQuestionInput{
		ID:               id,
		SubjectID:        subjectID,
		AuthoringMode:    body.AuthoringMode,
		Code:             body.Code,
		QuestionText:     body.QuestionText,
		QuestionType:     body.QuestionType,
		Options:          body.Options,
		OptionA:          body.OptionA,
		OptionB:          body.OptionB,
		OptionC:          body.OptionC,
		OptionD:          body.OptionD,
		OptionE:          body.OptionE,
		AnswerKey:        strings.TrimSpace(body.AnswerKey),
		Explanation:      body.Explanation,
		Difficulty:       db.CbtQuestionDifficultyEnum(body.Difficulty),
		Status:           db.CbtQuestionStatusEnum(body.Status),
		StemHTML:         body.StemHTML,
		StemLatex:        body.StemLatex,
		StimulusHTML:     body.StimulusHTML,
		StimulusLatex:    body.StimulusLatex,
		ExplanationHTML:  body.ExplanationHTML,
		RubricHTML:       body.RubricHTML,
		AcademicPhase:    body.AcademicPhase,
		GradeLevel:       gradeLevel,
		CPRef:            body.CPRef,
		TPRef:            body.TPRef,
		KDRef:            body.KDRef,
		IndicatorRef:     body.IndicatorRef,
		MaterialTopic:    body.MaterialTopic,
		CognitiveLevel:   body.CognitiveLevel,
		HotsFlag:         body.HotsFlag,
		MediaAssetIDs:    body.MediaAssetIDs,
		WorkflowStatus:   body.WorkflowStatus,
		AuthorUsername:   username,
		ReviewerUsername: username,
		ApproverUsername: username,
		WriterNotes:      body.WriterNotes,
		ReviewNotes:      body.ReviewNotes,
	}, nil
}

func currentUsername(r *http.Request) string {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return ""
	}
	if sub, _ := claims["sub"].(string); sub != "" {
		return sub
	}
	return ""
}

func currentRole(r *http.Request) string {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return ""
	}
	if role, _ := claims["role"].(string); role != "" {
		return role
	}
	return ""
}

func serializeQuestionListRow(row db.ListCbtQuestionsFilteredRow) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	return map[string]any{
		"id":                pgUUIDString(row.ID),
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
	}
}

func serializeQuestionDetailRow(row db.GetCbtQuestionDetailRow) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	return map[string]any{
		"id":                pgUUIDString(row.ID),
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
	}
}

func serializeQuestionModel(row db.CbtQuestion) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	return map[string]any{
		"id":                pgUUIDString(row.ID),
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
