package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

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
		Filename: fmt.Sprintf("bank-soal-%s.csv", time.Now().Format("20060102-150405")),
		Content:  []byte(builder.String()),
		Count:    len(rows),
	}, nil
}

func (s *CbtQuestion) TemplateCSV() (ExportCbtQuestionsCSVResult, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	records := templateQuestionCSVRecords()
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return ExportCbtQuestionsCSVResult{}, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return ExportCbtQuestionsCSVResult{}, err
	}
	return ExportCbtQuestionsCSVResult{
		Filename: "template-bank-soal.csv",
		Content:  []byte(builder.String()),
		Count:    len(records) - 1,
	}, nil
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

func templateQuestionCSVRecords() [][]string {
	return [][]string{
		exportQuestionCSVHeaders(),
		templateQuestionCSVRow(map[string]string{
			"kode":        "TPL-PG-001",
			"tipe":        "pg",
			"soal":        "Contoh soal pilihan ganda",
			"opsi_a":      "Opsi A",
			"opsi_b":      "Opsi B benar",
			"opsi_c":      "Opsi C",
			"opsi_d":      "Opsi D",
			"jawaban":     "B",
			"kesulitan":   "sedang",
			"grade_level": "8",
		}),
		templateQuestionCSVRow(map[string]string{
			"kode":        "TPL-PGK-001",
			"tipe":        "pg_kompleks",
			"soal":        "Contoh soal pilihan ganda kompleks",
			"opsi_a":      "Pernyataan A benar",
			"opsi_b":      "Pernyataan B",
			"opsi_c":      "Pernyataan C benar",
			"opsi_d":      "Pernyataan D",
			"jawaban":     "A,C",
			"kesulitan":   "sedang",
			"grade_level": "8",
			"hots_flag":   "false",
		}),
		templateQuestionCSVRow(map[string]string{
			"kode":      "TPL-BS-001",
			"tipe":      "benar_salah",
			"soal":      "Contoh pernyataan benar/salah",
			"jawaban":   "Benar",
			"kesulitan": "mudah",
		}),
		templateQuestionCSVRow(map[string]string{
			"kode":           "TPL-ISIAN-001",
			"tipe":           "isian",
			"soal":           "Contoh soal isian singkat",
			"jawaban":        "Jawaban utama | alias jawaban",
			"kesulitan":      "sedang",
			"material_topic": "Topik materi",
		}),
		templateQuestionCSVRow(map[string]string{
			"kode":         "TPL-JODOH-001",
			"tipe":         "menjodohkan",
			"soal":         "Contoh soal menjodohkan",
			"kiri_a":       "Istilah A",
			"kanan_1":      "Pasangan A",
			"kiri_b":       "Istilah B",
			"kanan_2":      "Pasangan B",
			"distraktor_1": "Distraktor kanan",
			"jawaban":      "A=1;B=2",
			"kesulitan":    "sedang",
		}),
		templateQuestionCSVRow(map[string]string{
			"kode":        "TPL-ESSAY-001",
			"tipe":        "essay",
			"soal":        "Contoh soal uraian",
			"rubrik":      "Tuliskan pedoman koreksi atau rubrik singkat",
			"pembahasan":  "Opsional: catatan pembahasan",
			"kesulitan":   "sedang",
			"grade_level": "9",
		}),
	}
}

func templateQuestionCSVRow(values map[string]string) []string {
	headers := exportQuestionCSVHeaders()
	row := make([]string, len(headers))
	for idx, header := range headers {
		row[idx] = values[header]
	}
	return row
}
