package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strconv"
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
var importAnswerTokenSeparators = regexp.MustCompile(`[,\|;/+\s]+`)
var importMatchingPairSeparators = regexp.MustCompile(`[;,|]+`)

type cbtQuestionStore interface {
	ListCbtQuestions(ctx context.Context) ([]db.ListCbtQuestionsRow, error)
	ListCbtQuestionsFiltered(ctx context.Context, arg db.ListCbtQuestionsFilteredParams) ([]db.ListCbtQuestionsFilteredRow, error)
	CountCbtQuestionsFiltered(ctx context.Context, arg db.CountCbtQuestionsFilteredParams) (int64, error)
	ListCbtQuestionStemTextsBySubject(ctx context.Context, subjectID pgtype.UUID) ([]db.ListCbtQuestionStemTextsBySubjectRow, error)
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
	Label        string `json:"label"`
	Text         string `json:"text,omitempty"`
	HTML         string `json:"html,omitempty"`
	Latex        string `json:"latex,omitempty"`
	AssetID      string `json:"asset_id,omitempty"`
	MatchLabel   string `json:"match_label,omitempty"`
	MatchText    string `json:"match_text,omitempty"`
	MatchHTML    string `json:"match_html,omitempty"`
	IsDistractor bool   `json:"is_distractor,omitempty"`
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

type ExportCbtQuestionsCSVResult struct {
	Filename string
	Content  []byte
	Count    int
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

func (s *CbtQuestion) ExportCSV(ctx context.Context, in ListCbtQuestionsInput) (ExportCbtQuestionsCSVResult, error) {
	if in.Limit <= 0 || in.Limit > 2000 {
		in.Limit = 2000
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	rows, _, err := s.ListFiltered(ctx, in)
	if err != nil {
		return ExportCbtQuestionsCSVResult{}, err
	}
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write(exportQuestionCSVHeaders()); err != nil {
		return ExportCbtQuestionsCSVResult{}, err
	}
	for _, row := range rows {
		if err := writer.Write(exportQuestionCSVRow(row)); err != nil {
			return ExportCbtQuestionsCSVResult{}, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return ExportCbtQuestionsCSVResult{}, err
	}
	return ExportCbtQuestionsCSVResult{
		Filename: fmt.Sprintf("bank-soal-cbt-%s.csv", time.Now().Format("20060102-150405")),
		Content:  []byte(builder.String()),
		Count:    len(rows),
	}, nil
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

type ImportLegacyQuestionsInput struct {
	SubjectID pgtype.UUID
	CSVText   string
	Username  string
}

type ImportLegacyQuestionsResult struct {
	TotalRows      int      `json:"total_rows"`
	Imported       int      `json:"imported"`
	Skipped        int      `json:"skipped"`
	Errors         []string `json:"errors"`
	DuplicateCodes []string `json:"duplicate_codes"`
}

func (s *CbtQuestion) ImportLegacyCSV(ctx context.Context, input ImportLegacyQuestionsInput) (ImportLegacyQuestionsResult, error) {
	reader := csv.NewReader(strings.NewReader(input.CSVText))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	if strings.Count(firstLine(input.CSVText), ";") > strings.Count(firstLine(input.CSVText), ",") {
		reader.Comma = ';'
	}
	records, err := reader.ReadAll()
	if err != nil {
		return ImportLegacyQuestionsResult{}, fmt.Errorf("CSV tidak valid: %w", err)
	}
	if len(records) < 2 {
		return ImportLegacyQuestionsResult{}, fmt.Errorf("CSV minimal berisi header dan satu baris soal")
	}
	headers := normalizeCSVHeaders(records[0])
	result := ImportLegacyQuestionsResult{TotalRows: len(records) - 1, Errors: []string{}, DuplicateCodes: []string{}}
	existingSignatures, err := s.existingQuestionSignatures(ctx, input.SubjectID)
	if err != nil {
		return ImportLegacyQuestionsResult{}, err
	}
	seen := map[string]bool{}
	seenCodes := map[string]bool{}
	skip := func(message string) {
		result.Skipped++
		result.Errors = append(result.Errors, message)
	}
	for idx, record := range records[1:] {
		rowNumber := idx + 2
		row := csvRow(headers, record)
		stem := strings.TrimSpace(firstCSVValue(row, "soal", "question", "stem", "pertanyaan"))
		if stem == "" {
			skip(fmt.Sprintf("Baris %d: soal kosong", rowNumber))
			continue
		}
		signature := strings.ToLower(strings.Join(strings.Fields(derivePlainText(stem)), " "))
		if seen[signature] {
			skip(fmt.Sprintf("Baris %d: duplikat dalam file import", rowNumber))
			continue
		}
		if existingSignatures[signature] {
			skip(fmt.Sprintf("Baris %d: duplikat dengan bank soal yang sudah ada", rowNumber))
			continue
		}
		seen[signature] = true
		code := strings.TrimSpace(firstCSVValue(row, "kode", "code"))
		if code != "" {
			codeKey := strings.ToUpper(code)
			if seenCodes[codeKey] {
				result.DuplicateCodes = append(result.DuplicateCodes, code)
				skip(fmt.Sprintf("Baris %d: kode %s duplikat dalam file import", rowNumber, code))
				continue
			}
			seenCodes[codeKey] = true
		}

		questionType, ok := normalizeImportQuestionType(firstCSVValue(row, "tipe", "jenis", "bentuk", "type", "question_type", "questiontype"))
		if !ok {
			skip(fmt.Sprintf("Baris %d: tipe soal tidak didukung", rowNumber))
			continue
		}

		options := importQuestionOptions(row, questionType)
		answerKey, ok := normalizeImportAnswerKey(questionType, firstCSVValue(row, "jawaban", "answer", "answer_key", "kunci", "kunci_jawaban"), options)
		if !ok {
			skip(fmt.Sprintf("Baris %d: kunci jawaban tidak valid untuk %s", rowNumber, importQuestionTypeLabel(questionType)))
			continue
		}

		stemHTML := appendLegacyImage(stem, firstCSVValue(row, "gambar", "gambarsoal", "gambar_soal", "image", "question_image"))
		rubricHTML := appendLegacyImage(firstCSVValue(row, "rubrik", "rubric", "pedoman", "pedoman_jawaban"), firstCSVValue(row, "gambar_rubrik", "rubric_image"))
		writerNotes := legacyImportNotes(row)
		_, err := s.Create(ctx, SaveCbtQuestionInput{
			SubjectID:        input.SubjectID,
			AuthoringMode:    "advance",
			Code:             code,
			QuestionText:     derivePlainText(stemHTML),
			QuestionType:     questionType,
			Options:          options,
			AnswerKey:        answerKey,
			Difficulty:       db.CbtQuestionDifficultyEnumMedium,
			Status:           db.CbtQuestionStatusEnumDraft,
			StemHTML:         stemHTML,
			RubricHTML:       rubricHTML,
			WorkflowStatus:   "draft",
			AuthorUsername:   strings.TrimSpace(input.Username),
			ReviewerUsername: "",
			ApproverUsername: "",
			WriterNotes:      writerNotes,
		})
		if err != nil {
			skip(fmt.Sprintf("Baris %d: %s", rowNumber, err.Error()))
			continue
		}
		result.Imported++
	}
	return result, nil
}

func (s *CbtQuestion) existingQuestionSignatures(ctx context.Context, subjectID pgtype.UUID) (map[string]bool, error) {
	rows, err := s.q.ListCbtQuestionStemTextsBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	signatures := make(map[string]bool, len(rows))
	for _, row := range rows {
		for _, value := range []string{row.QuestionText, row.StemHtml} {
			if sig := normalizedQuestionSignature(value); sig != "" {
				signatures[sig] = true
			}
		}
	}
	return signatures, nil
}

func firstLine(value string) string {
	if idx := strings.IndexAny(value, "\r\n"); idx >= 0 {
		return value[:idx]
	}
	return value
}

func normalizeCSVHeaders(headers []string) map[string]int {
	out := make(map[string]int, len(headers))
	for idx, header := range headers {
		key := normalizeCSVKey(header)
		if key != "" {
			out[key] = idx
		}
	}
	return out
}

func normalizeCSVKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "", "-", "", ".", "", "/", "")
	return replacer.Replace(value)
}

func normalizedQuestionSignature(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(derivePlainText(value)), " "))
}

func csvRow(headers map[string]int, record []string) map[string]string {
	row := map[string]string{}
	for key, idx := range headers {
		if idx >= 0 && idx < len(record) {
			row[key] = strings.TrimSpace(record[idx])
		}
	}
	return row
}

func firstCSVValue(row map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(row[normalizeCSVKey(key)]); value != "" {
			return value
		}
	}
	return ""
}

func normalizeLegacyAnswer(value string) (string, bool) {
	value = strings.ToUpper(strings.TrimSpace(value))
	switch value {
	case "A", "B", "C", "D", "E", "F":
		return value, true
	case "0":
		return "A", true
	case "1":
		return "B", true
	case "2":
		return "C", true
	case "3":
		return "D", true
	case "4":
		return "E", true
	case "5":
		return "F", true
	default:
		return "", false
	}
}

func normalizeImportQuestionType(value string) (string, bool) {
	switch normalizeImportToken(value) {
	case "":
		return "multiple_choice", true
	case "pg", "pilihanganda", "multiplechoice", "singlechoice":
		return "multiple_choice", true
	case "pgkompleks", "jawabanganda", "jawabanmajemuk", "pilihangandakompleks", "multipleanswer", "multipleanswers":
		return "multiple_answer", true
	case "benarsalah", "bs", "truefalse":
		return "true_false", true
	case "setujutidaksetuju", "sts", "agreedisagree":
		return "agree_disagree", true
	case "isian", "isiansingkat", "jawabansingkat", "shortanswer":
		return "short_answer", true
	case "essay", "esai", "uraian":
		return "essay", true
	case "menjodohkan", "jodohkan", "matching", "match":
		return "matching", true
	default:
		return "", false
	}
}

func normalizeImportAnswerKey(questionType string, value string, options []QuestionOption) (string, bool) {
	switch questionType {
	case "multiple_choice":
		return normalizeImportOptionAnswer(value, options, 1)
	case "multiple_answer":
		return normalizeImportOptionAnswer(value, options, 2)
	case "true_false":
		return normalizeImportFixedPairAnswer(value, "benar", "salah")
	case "agree_disagree":
		return normalizeImportFixedPairAnswer(value, "setuju", "tidaksetuju")
	case "short_answer":
		answerKey := normalizeShortAnswerKey(value)
		return answerKey, answerKey != ""
	case "matching":
		if !importMatchingOptionsComplete(options) {
			return "", false
		}
		return normalizeImportMatchingAnswerKey(value, countMatchingPairs(options))
	case "essay":
		return "", true
	default:
		return "", false
	}
}

func normalizeImportOptionAnswer(value string, options []QuestionOption, minKeys int) (string, bool) {
	available := importedOptionLabels(options)
	rawTokens := splitImportAnswerTokens(value)
	if len(rawTokens) == 1 && len(rawTokens[0]) > 1 && importTokenIsCompactLabels(rawTokens[0]) {
		rawTokens = strings.Split(strings.ToUpper(rawTokens[0]), "")
	}
	seen := make(map[string]bool, len(rawTokens))
	for _, token := range rawTokens {
		label, ok := normalizeLegacyAnswer(token)
		if !ok || !available[label] {
			return "", false
		}
		seen[label] = true
	}
	if len(seen) < minKeys {
		return "", false
	}
	ordered := make([]string, 0, len(seen))
	for _, label := range []string{"A", "B", "C", "D", "E", "F"} {
		if seen[label] {
			ordered = append(ordered, label)
		}
	}
	return strings.Join(ordered, ","), true
}

func normalizeImportFixedPairAnswer(value string, firstToken string, secondToken string) (string, bool) {
	switch normalizeImportToken(value) {
	case "a", firstToken, "true", "ya", "y", "1":
		return "A", true
	case "b", secondToken, "false", "tidak", "no", "n", "0":
		return "B", true
	default:
		return "", false
	}
}

func normalizeImportMatchingAnswerKey(value string, pairCount int) (string, bool) {
	if pairCount < 2 {
		return "", false
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return buildMatchingAnswerKey(pairCount), true
	}
	labels := make(map[string]bool, pairCount)
	for i := 0; i < pairCount; i++ {
		labels[string(rune('A'+i))] = true
	}
	matches := make(map[string]string, pairCount)
	seenRight := make(map[string]bool, pairCount)
	for _, pair := range importMatchingPairSeparators.Split(value, -1) {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.FieldsFunc(pair, func(r rune) bool {
			return r == '=' || r == ':' || r == '-'
		})
		if len(parts) != 2 {
			return "", false
		}
		left := strings.TrimSpace(strings.ToUpper(parts[0]))
		right := strings.TrimSpace(parts[1])
		rightIndex, err := strconv.Atoi(right)
		normalizedRight := strconv.Itoa(rightIndex)
		if !labels[left] || err != nil || rightIndex < 1 || rightIndex > pairCount || seenRight[normalizedRight] {
			return "", false
		}
		matches[left] = normalizedRight
		seenRight[normalizedRight] = true
	}
	if len(matches) != pairCount {
		return "", false
	}
	ordered := make([]string, 0, pairCount)
	for i := 0; i < pairCount; i++ {
		left := string(rune('A' + i))
		ordered = append(ordered, fmt.Sprintf("%s=%s", left, matches[left]))
	}
	return strings.Join(ordered, ";"), true
}

func splitImportAnswerTokens(value string) []string {
	tokens := importAnswerTokenSeparators.Split(strings.TrimSpace(value), -1)
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token != "" {
			out = append(out, token)
		}
	}
	return out
}

func importTokenIsCompactLabels(value string) bool {
	value = strings.TrimSpace(strings.ToUpper(value))
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < 'A' || r > 'F' {
			return false
		}
	}
	return true
}

func importedOptionLabels(options []QuestionOption) map[string]bool {
	available := make(map[string]bool, len(options))
	for _, option := range options {
		label := strings.TrimSpace(strings.ToUpper(option.Label))
		if label != "" && optionContent(option) != "" {
			available[label] = true
		}
	}
	return available
}

func importMatchingOptionsComplete(options []QuestionOption) bool {
	pairCount := 0
	for _, option := range options {
		if option.IsDistractor {
			if strings.TrimSpace(option.MatchLabel) == "" || matchingOptionContent(option) == "" {
				return false
			}
			continue
		}
		if optionContent(option) == "" || matchingOptionContent(option) == "" {
			return false
		}
		pairCount++
	}
	return pairCount >= 2
}

func importQuestionOptions(row map[string]string, questionType string) []QuestionOption {
	switch questionType {
	case "multiple_choice", "multiple_answer":
		return legacyImportOptions(row)
	case "true_false", "agree_disagree":
		return fixedPairQuestionOptions(questionType)
	case "matching":
		return legacyImportMatchingOptions(row)
	default:
		return nil
	}
}

func legacyImportOptions(row map[string]string) []QuestionOption {
	labels := []string{"A", "B", "C", "D", "E", "F"}
	options := make([]QuestionOption, 0, len(labels))
	for _, label := range labels {
		option := legacyImportOption(label, importOptionText(row, label), importOptionImage(row, label))
		if optionContent(option) != "" {
			options = append(options, option)
		}
	}
	return options
}

func legacyImportMatchingOptions(row map[string]string) []QuestionOption {
	options := make([]QuestionOption, 0, 10)
	pairCount := 0
	for idx, label := range []string{"A", "B", "C", "D", "E", "F"} {
		left := firstCSVValue(row,
			"kiri"+strings.ToLower(label),
			"kiri_"+strings.ToLower(label),
			"left"+strings.ToLower(label),
			"left_"+strings.ToLower(label),
			"pasangan"+strings.ToLower(label),
			"pasangan_"+strings.ToLower(label),
			"match_left_"+strings.ToLower(label),
		)
		rightLabel := strconv.Itoa(idx + 1)
		right := firstCSVValue(row,
			"kanan"+rightLabel,
			"kanan_"+rightLabel,
			"right"+rightLabel,
			"right_"+rightLabel,
			"pasangan"+rightLabel,
			"pasangan_"+rightLabel,
			"match_right_"+rightLabel,
		)
		if strings.TrimSpace(left) == "" && strings.TrimSpace(right) == "" {
			continue
		}
		pairCount++
		pairLabel := string(rune('A' + pairCount - 1))
		options = append(options, QuestionOption{
			Label:      pairLabel,
			Text:       left,
			MatchLabel: strconv.Itoa(pairCount),
			MatchText:  right,
		})
	}
	for idx := 1; idx <= 4; idx++ {
		value := firstCSVValue(row,
			fmt.Sprintf("distraktor%d", idx),
			fmt.Sprintf("distraktor_%d", idx),
			fmt.Sprintf("distractor%d", idx),
			fmt.Sprintf("distractor_%d", idx),
			fmt.Sprintf("kananextra%d", idx),
			fmt.Sprintf("kanan_extra_%d", idx),
			fmt.Sprintf("rightextra%d", idx),
			fmt.Sprintf("right_extra_%d", idx),
		)
		if strings.TrimSpace(value) == "" {
			continue
		}
		options = append(options, QuestionOption{
			MatchLabel:   strconv.Itoa(pairCount + idx),
			MatchText:    value,
			IsDistractor: true,
		})
	}
	return options
}

func importOptionText(row map[string]string, label string) string {
	lower := strings.ToLower(label)
	return firstCSVValue(row,
		"opsi"+lower,
		"opsi_"+lower,
		"option"+lower,
		"option_"+lower,
		"jawaban"+lower,
		"jawaban_"+lower,
	)
}

func importOptionImage(row map[string]string, label string) string {
	lower := strings.ToLower(label)
	return firstCSVValue(row,
		"gambar"+lower,
		"gambar_"+lower,
		"image"+lower,
		"image_"+lower,
	)
}

func normalizeImportToken(value string) string {
	value = normalizeCSVKey(value)
	value = strings.ReplaceAll(value, "_", "")
	return value
}

func importQuestionTypeLabel(questionType string) string {
	switch questionType {
	case "multiple_choice":
		return "Pilihan Ganda"
	case "multiple_answer":
		return "Pilihan Ganda Kompleks"
	case "true_false":
		return "Benar/Salah"
	case "agree_disagree":
		return "Setuju/Tidak Setuju"
	case "short_answer":
		return "Isian Singkat"
	case "matching":
		return "Menjodohkan"
	case "essay":
		return "Essay"
	default:
		return questionType
	}
}

func exportQuestionCSVHeaders() []string {
	headers := []string{
		"kode", "tipe", "mapel", "soal", "stimulus",
		"opsi_a", "opsi_b", "opsi_c", "opsi_d", "opsi_e", "opsi_f",
		"jawaban",
	}
	for idx, label := range []string{"a", "b", "c", "d", "e", "f"} {
		headers = append(headers, "kiri_"+label, "kanan_"+strconv.Itoa(idx+1))
	}
	headers = append(headers,
		"distraktor_1", "distraktor_2", "distraktor_3", "distraktor_4",
		"rubrik", "pembahasan", "kesulitan", "status", "workflow_status",
		"grade_level", "cp_ref", "tp_ref", "kd_ref", "indicator_ref",
		"material_topic", "cognitive_level", "hots_flag",
	)
	return headers
}

func exportQuestionCSVRow(row db.ListCbtQuestionsFilteredRow) []string {
	options := decodeQuestionOptions(row.Options)
	optionByLabel := exportQuestionOptionMap(options)
	matchingPairs, matchingDistractors := exportMatchingOptions(options)
	values := []string{
		row.Code,
		exportQuestionTypeAlias(row.QuestionType),
		row.SubjectName,
		firstExportContent(row.StemHtml, row.QuestionText),
		row.StimulusHtml,
		exportObjectiveOption(optionByLabel, "A", row.OptionA),
		exportObjectiveOption(optionByLabel, "B", row.OptionB),
		exportObjectiveOption(optionByLabel, "C", row.OptionC),
		exportObjectiveOption(optionByLabel, "D", row.OptionD),
		exportObjectiveOption(optionByLabel, "E", row.OptionE),
		exportObjectiveOption(optionByLabel, "F", ""),
		row.AnswerKey,
	}
	for idx := 0; idx < 6; idx++ {
		if idx < len(matchingPairs) {
			values = append(values, optionContent(matchingPairs[idx]), matchingOptionContent(matchingPairs[idx]))
			continue
		}
		values = append(values, "", "")
	}
	for idx := 0; idx < 4; idx++ {
		if idx < len(matchingDistractors) {
			values = append(values, matchingOptionContent(matchingDistractors[idx]))
			continue
		}
		values = append(values, "")
	}
	values = append(values,
		row.RubricHtml,
		firstExportContent(row.ExplanationHtml, row.Explanation),
		string(row.Difficulty),
		string(row.Status),
		row.WorkflowStatus,
		exportInt2(row.GradeLevel),
		row.CpRef,
		row.TpRef,
		row.KdRef,
		row.IndicatorRef,
		row.MaterialTopic,
		row.CognitiveLevel,
		strconv.FormatBool(row.HotsFlag),
	)
	return values
}

func exportQuestionTypeAlias(questionType string) string {
	switch questionType {
	case "multiple_choice", "single_choice":
		return "pg"
	case "multiple_answer":
		return "pg_kompleks"
	case "true_false":
		return "benar_salah"
	case "agree_disagree":
		return "setuju_tidak_setuju"
	case "short_answer":
		return "isian"
	case "matching":
		return "menjodohkan"
	case "essay":
		return "essay"
	default:
		return questionType
	}
}

func exportQuestionOptionMap(options []QuestionOption) map[string]QuestionOption {
	out := make(map[string]QuestionOption, len(options))
	for _, option := range options {
		if option.IsDistractor {
			continue
		}
		label := strings.TrimSpace(strings.ToUpper(option.Label))
		if label != "" {
			out[label] = option
		}
	}
	return out
}

func exportObjectiveOption(options map[string]QuestionOption, label string, fallback string) string {
	if option, ok := options[label]; ok {
		return optionContent(option)
	}
	return strings.TrimSpace(fallback)
}

func exportMatchingOptions(options []QuestionOption) ([]QuestionOption, []QuestionOption) {
	pairs := make([]QuestionOption, 0, len(options))
	distractors := make([]QuestionOption, 0, len(options))
	for _, option := range options {
		if option.IsDistractor {
			if matchingOptionContent(option) != "" {
				distractors = append(distractors, option)
			}
			continue
		}
		if optionContent(option) != "" || matchingOptionContent(option) != "" {
			pairs = append(pairs, option)
		}
	}
	return pairs, distractors
}

func firstExportContent(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func exportInt2(value pgtype.Int2) string {
	if !value.Valid {
		return ""
	}
	return strconv.Itoa(int(value.Int16))
}

func legacyImportOption(label, text, imageURL string) QuestionOption {
	text = strings.TrimSpace(text)
	htmlText := appendLegacyImage(text, imageURL)
	if imageURL != "" {
		return QuestionOption{Label: label, HTML: htmlText}
	}
	return QuestionOption{Label: label, Text: text}
}

func appendLegacyImage(content, imageURL string) string {
	content = strings.TrimSpace(content)
	imageURL = strings.TrimSpace(imageURL)
	if content == "" && imageURL == "" {
		return ""
	}
	if strings.Contains(strings.ToLower(content), "<img") || imageURL == "" {
		return content
	}
	safeSrc := html.EscapeString(imageURL)
	if content == "" {
		return fmt.Sprintf(`<p><img src="%s" alt="Media soal" /></p>`, safeSrc)
	}
	return fmt.Sprintf(`%s<p><img src="%s" alt="Media soal" /></p>`, content, safeSrc)
}

func legacyImportNotes(row map[string]string) string {
	notes := []string{"Import CSV legacy CBT lama."}
	if value := firstCSVValue(row, "bobot", "weight", "points"); value != "" {
		notes = append(notes, "Bobot legacy: "+value)
	}
	if value := firstCSVValue(row, "isrtl", "is_rtl", "rtl"); value == "true" || value == "1" {
		notes = append(notes, "RTL legacy: ya")
	}
	if value := firstCSVValue(row, "isshuffle", "is_shuffle", "shuffle"); value == "true" || value == "1" {
		notes = append(notes, "Shuffle opsi legacy: ya")
	}
	return strings.Join(notes, "\n")
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
		if out.WorkflowStatus != "review" {
			out.WorkflowStatus = "draft"
			out.ReviewerUsername = ""
		}
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
		out.AnswerKey = normalizeMatchingAnswerKey(out.AnswerKey, countMatchingPairs(normalizedOptions))
	} else {
		out.AnswerKey = strings.TrimSpace(strings.ToUpper(out.AnswerKey))
	}

	if err := validateQuestion(out); err != nil {
		return SaveCbtQuestionInput{}, err
	}

	return out, nil
}

func validateQuestion(input SaveCbtQuestionInput) error {
	requiresCompleteContent := input.WorkflowStatus != "draft" || input.Status != db.CbtQuestionStatusEnumDraft
	if input.AuthoringMode == "beginner" {
		if !beginnerSupportsQuestionType(input.QuestionType) {
			return fmt.Errorf("mode beginner belum mendukung tipe soal ini")
		}
	}

	switch input.QuestionType {
	case "multiple_choice", "single_choice", "multiple_answer", "true_false", "agree_disagree":
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
		if requiresCompleteContent {
			if input.RubricHTML == "" {
				return fmt.Errorf("rubric_html wajib diisi untuk essay yang diajukan review, di-approve, atau dipublish")
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

func validateObjectiveAnswerKey(options []QuestionOption, answerKey string, questionType string) error {
	available := make(map[string]bool, len(options))
	for _, option := range options {
		label := strings.TrimSpace(strings.ToUpper(option.Label))
		if label != "" {
			available[label] = true
		}
	}

	keys := []string{answerKey}
	if questionType == "multiple_answer" {
		keys = strings.Split(answerKey, ",")
	}
	validKeyCount := 0
	for _, key := range keys {
		key = strings.TrimSpace(strings.ToUpper(key))
		if key == "" || !available[key] {
			return fmt.Errorf("answer_key harus sesuai label opsi yang tersedia")
		}
		validKeyCount++
	}
	if questionType == "multiple_answer" && validKeyCount < 2 {
		return fmt.Errorf("multiple_answer membutuhkan minimal 2 kunci jawaban")
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
	for _, pair := range strings.Split(answerKey, ";") {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return fmt.Errorf("answer_key menjodohkan tidak valid")
		}
		left := strings.TrimSpace(strings.ToUpper(parts[0]))
		right := strings.TrimSpace(parts[1])
		if !leftLabels[left] || !correctRightLabels[right] {
			return fmt.Errorf("answer_key menjodohkan harus sesuai label pasangan")
		}
		delete(leftLabels, left)
	}
	if len(leftLabels) != 0 {
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
	return strings.Join(aliases, "|")
}

func normalizeMatchingAnswerKey(value string, optionCount int) string {
	if optionCount <= 0 {
		return ""
	}
	fallback := buildMatchingAnswerKey(optionCount)
	pairs := strings.Split(value, ";")
	labels := make(map[string]bool, optionCount)
	matches := make(map[string]string, optionCount)
	for i := 0; i < optionCount; i++ {
		labels[string(rune('A'+i))] = true
	}
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			continue
		}
		left := strings.TrimSpace(strings.ToUpper(parts[0]))
		right := strings.TrimSpace(parts[1])
		if !labels[left] {
			continue
		}
		rightIndex, err := strconv.Atoi(right)
		if err != nil || rightIndex < 1 || rightIndex > optionCount {
			continue
		}
		matches[left] = strconv.Itoa(rightIndex)
	}
	if len(matches) != optionCount {
		return fallback
	}
	ordered := make([]string, 0, optionCount)
	for i := 0; i < optionCount; i++ {
		left := string(rune('A' + i))
		ordered = append(ordered, fmt.Sprintf("%s=%s", left, matches[left]))
	}
	return strings.Join(ordered, ";")
}

func buildMatchingAnswerKey(optionCount int) string {
	pairs := make([]string, 0, optionCount)
	for i := 0; i < optionCount; i++ {
		pairs = append(pairs, fmt.Sprintf("%s=%d", string(rune('A'+i)), i+1))
	}
	return strings.Join(pairs, ";")
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
	case "multiple_choice", "multiple_answer", "true_false", "agree_disagree", "matching", "short_answer", "essay":
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
	case "multiple_answer", "true_false", "agree_disagree", "matching", "short_answer", "essay":
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
