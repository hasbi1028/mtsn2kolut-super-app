package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var stripHTMLTags = regexp.MustCompile(`(?s)<[^>]*>`)
var stripDangerousBlockPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
	regexp.MustCompile(`(?is)<iframe[^>]*>.*?</iframe>`),
	regexp.MustCompile(`(?is)<object[^>]*>.*?</object>`),
	regexp.MustCompile(`(?is)<embed[^>]*>.*?</embed>`),
}
var stripEventHandlers = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*(".*?"|'.*?'|[^\s>]+)`)
var stripDangerousURLs = regexp.MustCompile(`(?i)\s(href|src)\s*=\s*(['"])\s*javascript:[^'"]*['"]`)

type cbtQuestionStore interface {
	ListCbtQuestions(ctx context.Context) ([]db.ListCbtQuestionsRow, error)
	ListCbtQuestionsFiltered(ctx context.Context, arg db.ListCbtQuestionsFilteredParams) ([]db.ListCbtQuestionsFilteredRow, error)
	CountCbtQuestionsFiltered(ctx context.Context, arg db.CountCbtQuestionsFilteredParams) (int64, error)
	GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error)
	GetCbtQuestionDetail(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionDetailRow, error)
	CreateCbtQuestion(ctx context.Context, arg db.CreateCbtQuestionParams) (db.CbtQuestion, error)
	UpdateCbtQuestion(ctx context.Context, arg db.UpdateCbtQuestionParams) (db.CbtQuestion, error)
	DeleteCbtQuestion(ctx context.Context, id pgtype.UUID) error
}

type CbtQuestion struct {
	q cbtQuestionStore
}

func NewCbtQuestion(q *db.Queries) *CbtQuestion { return &CbtQuestion{q: q} }

type QuestionOption struct {
	Label   string `json:"label"`
	Text    string `json:"text,omitempty"`
	HTML    string `json:"html,omitempty"`
	Latex   string `json:"latex,omitempty"`
	AssetID string `json:"asset_id,omitempty"`
}

type SaveCbtQuestionInput struct {
	ID               pgtype.UUID
	SubjectID        pgtype.UUID
	AuthoringMode    string
	Code             string
	QuestionText     string
	QuestionType     string
	Options          []QuestionOption
	OptionA          string
	OptionB          string
	OptionC          string
	OptionD          string
	OptionE          string
	AnswerKey        string
	Explanation      string
	Difficulty       db.CbtQuestionDifficultyEnum
	Status           db.CbtQuestionStatusEnum
	StemHTML         string
	StemLatex        string
	StimulusHTML     string
	StimulusLatex    string
	ExplanationHTML  string
	RubricHTML       string
	AcademicPhase    string
	GradeLevel       pgtype.Int2
	CPRef            string
	TPRef            string
	KDRef            string
	IndicatorRef     string
	MaterialTopic    string
	CognitiveLevel   string
	HotsFlag         bool
	MediaAssetIDs    []string
	WorkflowStatus   string
	AuthorUsername   string
	ReviewerUsername string
	ApproverUsername string
	WriterNotes      string
	ReviewNotes      string
}

func (s *CbtQuestion) List(ctx context.Context) ([]db.ListCbtQuestionsRow, error) {
	return s.q.ListCbtQuestions(ctx)
}

type ListCbtQuestionsInput struct {
	SubjectID      pgtype.UUID
	WorkflowStatus string
	QuestionType   string
	HotsFilter     string
	SearchQuery    string
	Limit          int32
	Offset         int32
}

func (s *CbtQuestion) ListFiltered(ctx context.Context, in ListCbtQuestionsInput) ([]db.ListCbtQuestionsFilteredRow, int64, error) {
	rows, err := s.q.ListCbtQuestionsFiltered(ctx, db.ListCbtQuestionsFilteredParams{
		SubjectID:      in.SubjectID,
		WorkflowStatus: strings.TrimSpace(in.WorkflowStatus),
		QuestionType:   strings.TrimSpace(in.QuestionType),
		HotsFilter:     strings.TrimSpace(in.HotsFilter),
		SearchQuery:    strings.TrimSpace(in.SearchQuery),
		LimitCount:     in.Limit,
		OffsetCount:    in.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountCbtQuestionsFiltered(ctx, db.CountCbtQuestionsFilteredParams{
		SubjectID:      in.SubjectID,
		WorkflowStatus: strings.TrimSpace(in.WorkflowStatus),
		QuestionType:   strings.TrimSpace(in.QuestionType),
		HotsFilter:     strings.TrimSpace(in.HotsFilter),
		SearchQuery:    strings.TrimSpace(in.SearchQuery),
	})
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *CbtQuestion) Get(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	return s.q.GetCbtQuestion(ctx, id)
}

func (s *CbtQuestion) GetDetail(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionDetailRow, error) {
	return s.q.GetCbtQuestionDetail(ctx, id)
}

func (s *CbtQuestion) Create(ctx context.Context, input SaveCbtQuestionInput) (db.CbtQuestion, error) {
	params, err := buildCreateQuestionParams(input)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	return s.q.CreateCbtQuestion(ctx, params)
}

func (s *CbtQuestion) Update(ctx context.Context, input SaveCbtQuestionInput) (db.CbtQuestion, error) {
	current, err := s.q.GetCbtQuestion(ctx, input.ID)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	params, err := buildUpdateQuestionParams(current, input)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	return s.q.UpdateCbtQuestion(ctx, params)
}

func (s *CbtQuestion) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteCbtQuestion(ctx, id)
}

func (s *CbtQuestion) SubmitReview(ctx context.Context, id pgtype.UUID, username string, reviewNotes string) (db.CbtQuestion, error) {
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, username)
	input.WorkflowStatus = "review"
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.Update(ctx, input)
}

func (s *CbtQuestion) Approve(ctx context.Context, id pgtype.UUID, username string, reviewNotes string) (db.CbtQuestion, error) {
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, username)
	input.WorkflowStatus = "approved"
	input.ReviewNotes = mergeNotes(current.ReviewNotes, reviewNotes)
	return s.Update(ctx, input)
}

func (s *CbtQuestion) Publish(ctx context.Context, id pgtype.UUID, username string) (db.CbtQuestion, error) {
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, username)
	input.Status = db.CbtQuestionStatusEnumPublished
	return s.Update(ctx, input)
}

func (s *CbtQuestion) Archive(ctx context.Context, id pgtype.UUID, username string) (db.CbtQuestion, error) {
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, username)
	input.Status = db.CbtQuestionStatusEnumArchived
	return s.Update(ctx, input)
}

func (s *CbtQuestion) DuplicateAsDraft(ctx context.Context, id pgtype.UUID, username string) (db.CbtQuestion, error) {
	current, err := s.q.GetCbtQuestion(ctx, id)
	if err != nil {
		return db.CbtQuestion{}, err
	}
	input := questionInputFromCurrent(current, username)
	input.ID = pgtype.UUID{}
	input.Status = db.CbtQuestionStatusEnumDraft
	input.WorkflowStatus = "draft"
	input.ReviewerUsername = ""
	input.ApproverUsername = ""
	input.ReviewNotes = ""
	if input.Code != "" {
		input.Code = input.Code + "-COPY"
	}
	return s.Create(ctx, input)
}

func buildCreateQuestionParams(input SaveCbtQuestionInput) (db.CreateCbtQuestionParams, error) {
	normalized, err := normalizeQuestionInput(input)
	if err != nil {
		return db.CreateCbtQuestionParams{}, err
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
		GradeLevel:       normalized.GradeLevel,
		CpRef:            normalized.CPRef,
		TpRef:            normalized.TPRef,
		KdRef:            normalized.KDRef,
		IndicatorRef:     normalized.IndicatorRef,
		MaterialTopic:    normalized.MaterialTopic,
		CognitiveLevel:   normalized.CognitiveLevel,
		HotsFlag:         normalized.HotsFlag,
		MediaAssetIds:    mediaJSON,
		WorkflowStatus:   normalized.WorkflowStatus,
		Version:          1,
		AuthorUsername:   normalized.AuthorUsername,
		ReviewerUsername: normalized.ReviewerUsername,
		ReviewedAt:       normalized.reviewedAt(),
		ApproverUsername: normalized.ApproverUsername,
		ApprovedAt:       normalized.approvedAt(),
		WriterNotes:      normalized.WriterNotes,
		ReviewNotes:      normalized.ReviewNotes,
	}, nil
}

func buildUpdateQuestionParams(current db.GetCbtQuestionRow, input SaveCbtQuestionInput) (db.UpdateCbtQuestionParams, error) {
	normalized, err := normalizeQuestionInput(input)
	if err != nil {
		return db.UpdateCbtQuestionParams{}, err
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
	if normalized.WorkflowStatus == "approved" || normalized.WorkflowStatus == "review" {
		reviewer = normalized.ReviewerUsername
		reviewedAt = normalized.reviewedAt()
	}

	approver := current.ApproverUsername
	approvedAt := current.ApprovedAt
	if normalized.Status == db.CbtQuestionStatusEnumPublished {
		approver = normalized.ApproverUsername
		approvedAt = normalized.approvedAt()
	}

	return db.UpdateCbtQuestionParams{
		ID:               input.ID,
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
		GradeLevel:       normalized.GradeLevel,
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
	out.CPRef = strings.TrimSpace(out.CPRef)
	out.TPRef = strings.TrimSpace(out.TPRef)
	out.KDRef = strings.TrimSpace(out.KDRef)
	out.IndicatorRef = strings.TrimSpace(out.IndicatorRef)
	out.MaterialTopic = strings.TrimSpace(out.MaterialTopic)
	out.CognitiveLevel = strings.TrimSpace(out.CognitiveLevel)
	out.WorkflowStatus = normalizeWorkflowStatus(out.WorkflowStatus)
	out.AuthorUsername = strings.TrimSpace(out.AuthorUsername)
	out.ReviewerUsername = strings.TrimSpace(out.ReviewerUsername)
	out.ApproverUsername = strings.TrimSpace(out.ApproverUsername)
	out.WriterNotes = strings.TrimSpace(out.WriterNotes)
	out.ReviewNotes = strings.TrimSpace(out.ReviewNotes)

	if out.Difficulty == "" {
		out.Difficulty = db.CbtQuestionDifficultyEnumMedium
	}
	if out.Status == "" {
		out.Status = db.CbtQuestionStatusEnumDraft
	}
	if out.AuthoringMode == "beginner" {
		out.Difficulty = db.CbtQuestionDifficultyEnumMedium
		out.Status = db.CbtQuestionStatusEnumDraft
		out.WorkflowStatus = "draft"
		out.ReviewerUsername = ""
		out.ApproverUsername = ""
		out.WriterNotes = ""
		out.ReviewNotes = ""
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
	if out.QuestionType == "true_false" && len(options) == 0 {
		options = []QuestionOption{
			{Label: "A", Text: "Benar"},
			{Label: "B", Text: "Salah"},
		}
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
	out.AnswerKey = strings.TrimSpace(strings.ToUpper(out.AnswerKey))

	if err := validateQuestion(out); err != nil {
		return SaveCbtQuestionInput{}, err
	}

	return out, nil
}

func validateQuestion(input SaveCbtQuestionInput) error {
	if input.AuthoringMode == "beginner" {
		switch input.QuestionType {
		case "multiple_choice", "essay":
		default:
			return fmt.Errorf("mode beginner hanya mendukung pilihan ganda dan essay")
		}
	}

	switch input.QuestionType {
	case "multiple_choice", "single_choice", "multiple_answer", "true_false":
		if len(input.Options) < 2 {
			return fmt.Errorf("opsi jawaban minimal 2 untuk tipe soal objektif")
		}
		if input.AuthoringMode == "beginner" && len(input.Options) < 4 {
			return fmt.Errorf("mode beginner membutuhkan minimal 4 opsi untuk pilihan ganda")
		}
		if input.AnswerKey == "" {
			return fmt.Errorf("answer_key wajib diisi")
		}
	case "short_answer":
		if input.AnswerKey == "" {
			return fmt.Errorf("answer_key wajib diisi untuk short_answer")
		}
	case "essay":
		if input.WorkflowStatus == "approved" || input.Status == db.CbtQuestionStatusEnumPublished {
			if input.RubricHTML == "" {
				return fmt.Errorf("rubric_html wajib diisi untuk essay yang di-approve atau dipublish")
			}
		}
	default:
		return fmt.Errorf("question_type tidak didukung")
	}

	if input.Status == db.CbtQuestionStatusEnumPublished && input.WorkflowStatus != "approved" {
		return fmt.Errorf("soal hanya boleh dipublish jika workflow_status sudah approved")
	}

	return nil
}

func normalizeQuestionType(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "multiple_choice", "single_choice":
		return "multiple_choice"
	case "multiple_answer", "true_false", "short_answer", "essay":
		return value
	default:
		return strings.TrimSpace(strings.ToLower(value))
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
	case "review", "approved", "rejected":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "draft"
	}
}

func normalizeOptions(options []QuestionOption) ([]QuestionOption, error) {
	normalized := make([]QuestionOption, 0, len(options))
	for idx, option := range options {
		label := strings.TrimSpace(strings.ToUpper(option.Label))
		if label == "" {
			label = string(rune('A' + idx))
		}
		text := strings.TrimSpace(option.Text)
		htmlText := strings.TrimSpace(option.HTML)
		latex := strings.TrimSpace(option.Latex)
		if text == "" && htmlText == "" && latex == "" {
			continue
		}
		normalized = append(normalized, QuestionOption{
			Label:   label,
			Text:    text,
			HTML:    sanitizeHTML(htmlText),
			Latex:   latex,
			AssetID: strings.TrimSpace(option.AssetID),
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
	plain := stripHTMLTags.ReplaceAllString(value, " ")
	plain = html.UnescapeString(strings.TrimSpace(plain))
	plain = strings.Join(strings.Fields(plain), " ")
	if len(plain) > 500 {
		return plain[:500]
	}
	return plain
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
	for _, pattern := range stripDangerousBlockPatterns {
		value = pattern.ReplaceAllString(value, "")
	}
	value = stripEventHandlers.ReplaceAllString(value, "")
	value = stripDangerousURLs.ReplaceAllString(value, "")
	return strings.TrimSpace(value)
}

func mergeNotes(existing string, incoming string) string {
	incoming = strings.TrimSpace(incoming)
	if incoming == "" {
		return existing
	}
	return incoming
}

func questionInputFromCurrent(current db.GetCbtQuestionRow, username string) SaveCbtQuestionInput {
	return SaveCbtQuestionInput{
		ID:               current.ID,
		SubjectID:        current.SubjectID,
		AuthoringMode:    "advance",
		Code:             current.Code,
		QuestionText:     current.QuestionText,
		QuestionType:     current.QuestionType,
		Options:          decodeQuestionOptions(current.Options),
		OptionA:          current.OptionA,
		OptionB:          current.OptionB,
		OptionC:          current.OptionC,
		OptionD:          current.OptionD,
		OptionE:          current.OptionE,
		AnswerKey:        current.AnswerKey,
		Explanation:      current.Explanation,
		Difficulty:       current.Difficulty,
		Status:           current.Status,
		StemHTML:         current.StemHtml,
		StemLatex:        current.StemLatex,
		StimulusHTML:     current.StimulusHtml,
		StimulusLatex:    current.StimulusLatex,
		ExplanationHTML:  current.ExplanationHtml,
		RubricHTML:       current.RubricHtml,
		AcademicPhase:    current.AcademicPhase,
		GradeLevel:       current.GradeLevel,
		CPRef:            current.CpRef,
		TPRef:            current.TpRef,
		KDRef:            current.KdRef,
		IndicatorRef:     current.IndicatorRef,
		MaterialTopic:    current.MaterialTopic,
		CognitiveLevel:   current.CognitiveLevel,
		HotsFlag:         current.HotsFlag,
		MediaAssetIDs:    decodeStringArray(current.MediaAssetIds),
		WorkflowStatus:   current.WorkflowStatus,
		AuthorUsername:   current.AuthorUsername,
		ReviewerUsername: username,
		ApproverUsername: username,
		WriterNotes:      current.WriterNotes,
		ReviewNotes:      current.ReviewNotes,
	}
}

func suggestQuestionAuthoringMode(questionType, stemLatex, stimulusLatex, academicPhase, cpRef, tpRef, kdRef, indicatorRef, materialTopic, cognitiveLevel string, hotsFlag bool, workflowStatus, writerNotes, reviewNotes, rubricHTML string) string {
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
