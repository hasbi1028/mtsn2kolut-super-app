package service

import (
	"context"
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
		Actor:          CbtQuestionActor{Username: "teacher", Roles: []string{"guru"}},
	}
}

func TestBuildCreateQuestionParamsLeavesCodeBlankForStoreAcademicCode(t *testing.T) {
	input := validCbtQuestionCodeTestInput()
	input.Code = ""

	params, err := buildCreateQuestionParams(input)
	if err != nil {
		t.Fatalf("buildCreateQuestionParams() error = %v", err)
	}
	if params.Code != "" {
		t.Fatalf("buildCreateQuestionParams code = %q, want blank before store academic code generation", params.Code)
	}
}

func TestCreateQuestionGeneratesAcademicCodeWhenComposerOmitsCode(t *testing.T) {
	input := validCbtQuestionCodeTestInput()
	input.Code = ""
	store := &fakeQuestionStore{generatedCode: "MTK-VII-PG-0007"}
	svc := &CbtQuestion{q: store}

	_, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if store.createParams.Code != "MTK-VII-PG-0007" {
		t.Fatalf("created code = %q, want academic code", store.createParams.Code)
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
