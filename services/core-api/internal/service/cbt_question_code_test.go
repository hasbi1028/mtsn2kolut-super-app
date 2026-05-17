package service

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func validCbtQuestionCodeTestInput() SaveCbtQuestionInput {
	return SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		QuestionText:   "What is the correct answer?",
		QuestionType:   "multiple_choice",
		Options:        []QuestionOption{{Label: "A", Text: "Alpha"}, {Label: "B", Text: "Beta"}, {Label: "C", Text: "Gamma"}, {Label: "D", Text: "Delta"}},
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
		TargetLevel:    "VII",
		AuthorUsername: "teacher",
	}
}

func TestBuildCreateQuestionParamsGeneratesCodeWhenComposerOmitsCode(t *testing.T) {
	input := validCbtQuestionCodeTestInput()
	input.Code = ""

	params, err := buildCreateQuestionParams(input)
	if err != nil {
		t.Fatalf("buildCreateQuestionParams() error = %v", err)
	}
	if !strings.HasPrefix(params.Code, "SOAL-") {
		t.Fatalf("generated code = %q, want SOAL-*", params.Code)
	}
	if strings.TrimSpace(params.Code) == "" {
		t.Fatal("generated code must not be blank")
	}
}

func TestBuildUpdateQuestionParamsKeepsCurrentCodeWhenComposerOmitsCode(t *testing.T) {
	input := validCbtQuestionCodeTestInput()
	input.Code = ""
	current := db.GetCbtQuestionRow{ID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true}, Code: "KEEP-ME"}

	params, err := buildUpdateQuestionParams(current, input)
	if err != nil {
		t.Fatalf("buildUpdateQuestionParams() error = %v", err)
	}
	if params.Code != "KEEP-ME" {
		t.Fatalf("update code = %q, want existing code", params.Code)
	}
}
