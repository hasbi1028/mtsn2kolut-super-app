package service

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestExportQuestionTypeAliasExtra(t *testing.T) {
	cases := map[string]string{
		"multiple_choice":  "pg",
		"single_choice":    "pg",
		"multiple_answer":  "pg_kompleks",
		"true_false":       "benar_salah",
		"agree_disagree":   "setuju_tidak_setuju",
		"short_answer":     "isian",
		"matching":         "menjodohkan",
		"essay":            "essay",
		"custom_free_text": "custom_free_text",
	}
	for input, want := range cases {
		t.Run(input, func(t *testing.T) {
			if got := exportQuestionTypeAlias(input); got != want {
				t.Fatalf("exportQuestionTypeAlias(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func TestExportQuestionCSVRowExtraUsesAliasesAndContentFallbacks(t *testing.T) {
	row := db.ListCbtQuestionsFilteredRow{
		Code:           "Q-EXP-1",
		QuestionType:   "single_choice",
		SubjectName:    "Matematika",
		QuestionText:   "Plain stem",
		OptionA:        "Fallback A",
		OptionB:        "Fallback B",
		AnswerKey:      "B",
		Explanation:    "Plain explanation",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		TargetLevel:    pgtype.Text{String: "vii", Valid: true},
		CpRef:          "CP-1",
		TpRef:          "TP-1",
		KdRef:          "KD-1",
		IndicatorRef:   "IND-1",
		MaterialTopic:  "Peluang",
		CognitiveLevel: "C3",
		HotsFlag:       true,
	}

	values := exportQuestionCSVRow(row)
	if len(values) != len(exportQuestionCSVHeaders()) {
		t.Fatalf("exportQuestionCSVRow() length = %d, want header length %d", len(values), len(exportQuestionCSVHeaders()))
	}
	byHeader := map[string]string{}
	for idx, header := range exportQuestionCSVHeaders() {
		byHeader[header] = values[idx]
	}
	checks := map[string]string{
		"kode":            "Q-EXP-1",
		"tipe":            "pg",
		"mapel":           "Matematika",
		"soal":            "Plain stem",
		"opsi_a":          "Fallback A",
		"opsi_b":          "Fallback B",
		"jawaban":         "B",
		"pembahasan":      "Plain explanation",
		"target_level":    "VII",
		"material_topic":  "Peluang",
		"cognitive_level": "C3",
		"hots_flag":       "true",
	}
	for header, want := range checks {
		if got := byHeader[header]; got != want {
			t.Fatalf("exportQuestionCSVRow()[%s] = %q, want %q", header, got, want)
		}
	}
}

func TestTemplateCSVExtraIsReadableAndUsesImportAliases(t *testing.T) {
	result, err := NewCbtQuestion(nil).TemplateCSV()
	if err != nil {
		t.Fatalf("TemplateCSV() error = %v", err)
	}
	if result.Filename != "template-bank-soal.csv" || result.Count < 5 {
		t.Fatalf("TemplateCSV() result = %+v, want template filename and multiple sample rows", result)
	}
	records, err := csv.NewReader(strings.NewReader(string(result.Content))).ReadAll()
	if err != nil {
		t.Fatalf("template CSV is not readable: %v\n%s", err, result.Content)
	}
	if len(records) != result.Count+1 {
		t.Fatalf("template record count = %d, want %d", len(records), result.Count+1)
	}
	if records[0][0] != "kode" || records[0][1] != "tipe" || records[0][3] != "soal" {
		t.Fatalf("template header = %#v, want import-compatible leading headers", records[0][:4])
	}
	seenTypes := map[string]bool{}
	for _, record := range records[1:] {
		seenTypes[record[1]] = true
	}
	for _, want := range []string{"pg", "pg_kompleks", "benar_salah", "isian", "menjodohkan"} {
		if !seenTypes[want] {
			t.Fatalf("template types = %#v, missing %q", seenTypes, want)
		}
	}
}
