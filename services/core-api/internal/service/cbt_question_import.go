package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type ImportLegacyQuestionsInput struct {
	SubjectID pgtype.UUID
	EventID   pgtype.UUID
	CSVText   string
	Username  string
	Actor     CbtQuestionActor
	DryRun    bool
}

type ImportLegacyQuestionsResult struct {
	TotalRows      int      `json:"total_rows"`
	DryRun         bool     `json:"dry_run"`
	WouldImport    int      `json:"would_import"`
	Imported       int      `json:"imported"`
	Skipped        int      `json:"skipped"`
	Errors         []string `json:"errors"`
	DuplicateCodes []string `json:"duplicate_codes"`
}

func (s *CbtQuestion) ImportLegacyCSV(ctx context.Context, input ImportLegacyQuestionsInput) (ImportLegacyQuestionsResult, error) {
	actor := normalizeCbtQuestionActor(input.Actor)
	if actor.Username == "" {
		actor.Username = strings.TrimSpace(input.Username)
	}
	if err := s.requireCreateQuestion(ctx, actor, input.EventID, input.SubjectID); err != nil {
		return ImportLegacyQuestionsResult{}, err
	}
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
	result := ImportLegacyQuestionsResult{TotalRows: len(records) - 1, DryRun: input.DryRun, Errors: []string{}, DuplicateCodes: []string{}}
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
		stimulusHTML := firstCSVValue(row, "stimulus", "stimulus_html", "stimulushtml")
		explanationHTML := firstCSVValue(row, "pembahasan", "explanation", "explanation_html", "explanationhtml")
		rubricHTML := appendLegacyImage(firstCSVValue(row, "rubrik", "rubric", "pedoman", "pedoman_jawaban"), firstCSVValue(row, "gambar_rubrik", "rubric_image"))
		writerNotes := legacyImportNotes(row)
		saveInput := SaveCbtQuestionInput{
			EventID:          input.EventID,
			SubjectID:        input.SubjectID,
			AuthoringMode:    "advance",
			Code:             code,
			QuestionText:     derivePlainText(stemHTML),
			QuestionType:     questionType,
			Options:          options,
			AnswerKey:        answerKey,
			Difficulty:       importQuestionDifficulty(row),
			Status:           db.CbtQuestionStatusEnumDraft,
			StemHTML:         stemHTML,
			StimulusHTML:     stimulusHTML,
			ExplanationHTML:  explanationHTML,
			RubricHTML:       rubricHTML,
			GradeLevel:       importGradeLevel(row),
			TargetLevel:      importTargetLevel(row),
			CPRef:            firstCSVValue(row, "cp_ref", "cpref", "cp"),
			TPRef:            firstCSVValue(row, "tp_ref", "tpref", "tp"),
			KDRef:            firstCSVValue(row, "kd_ref", "kdref", "kd"),
			IndicatorRef:     firstCSVValue(row, "indicator_ref", "indicatorref", "indikator", "indikator_ref"),
			MaterialTopic:    firstCSVValue(row, "material_topic", "materialtopic", "materi", "topik"),
			CognitiveLevel:   firstCSVValue(row, "cognitive_level", "cognitivelevel", "level_kognitif", "levelkognitif"),
			HotsFlag:         importBoolean(row, "hots_flag", "hotsflag", "hots"),
			WorkflowStatus:   "draft",
			AuthorUsername:   actor.Username,
			ReviewerUsername: "",
			ApproverUsername: "",
			WriterNotes:      writerNotes,
			Actor:            actor,
		}
		if input.DryRun {
			result.WouldImport++
			continue
		}
		_, err := s.createWithAudit(ctx, saveInput, "import", "", map[string]any{"event_id": cbtQuestionUUIDString(input.EventID), "subject_id": cbtQuestionUUIDString(input.SubjectID)})
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

func importQuestionDifficulty(row map[string]string) db.CbtQuestionDifficultyEnum {
	switch normalizeImportToken(firstCSVValue(row, "kesulitan", "difficulty")) {
	case "easy", "mudah", "rendah":
		return db.CbtQuestionDifficultyEnumEasy
	case "hard", "sulit", "tinggi":
		return db.CbtQuestionDifficultyEnumHard
	case "medium", "sedang", "menengah":
		return db.CbtQuestionDifficultyEnumMedium
	default:
		return db.CbtQuestionDifficultyEnumMedium
	}
}

func importTargetLevel(row map[string]string) string {
	value := strings.TrimSpace(firstCSVValue(row, "target_level", "targetlevel", "tingkat_soal", "tingkatsoal", "tingkat", "kelas", "level"))
	if normalized, ok := normalizeQuestionTargetLevel(value); ok {
		return normalized
	}
	legacyGrade := importGradeLevel(row)
	return questionTargetLevelFromGradeLevel(legacyGrade)
}

func importGradeLevel(row map[string]string) pgtype.Int2 {
	value := strings.TrimSpace(firstCSVValue(row, "grade_level", "gradelevel", "kelas", "tingkat"))
	if value == "" {
		return pgtype.Int2{}
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 || parsed > 12 {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: int16(parsed), Valid: true}
}

func importBoolean(row map[string]string, keys ...string) bool {
	switch normalizeImportToken(firstCSVValue(row, keys...)) {
	case "true", "1", "ya", "y", "yes", "hots":
		return true
	default:
		return false
	}
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
