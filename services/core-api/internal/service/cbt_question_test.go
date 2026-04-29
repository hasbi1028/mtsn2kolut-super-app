package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeQuestionStore struct {
	current      db.GetCbtQuestionRow
	updateParams db.UpdateCbtQuestionParams
	updateCalls  int
}

func (f *fakeQuestionStore) ListCbtQuestions(ctx context.Context) ([]db.ListCbtQuestionsRow, error) {
	return nil, nil
}

func (f *fakeQuestionStore) ListCbtQuestionsFiltered(ctx context.Context, arg db.ListCbtQuestionsFilteredParams) ([]db.ListCbtQuestionsFilteredRow, error) {
	return nil, nil
}

func (f *fakeQuestionStore) CountCbtQuestionsFiltered(ctx context.Context, arg db.CountCbtQuestionsFilteredParams) (int64, error) {
	return 0, nil
}

func (f *fakeQuestionStore) GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	if !f.current.ID.Valid {
		return db.GetCbtQuestionRow{}, errors.New("not found")
	}
	return f.current, nil
}

func (f *fakeQuestionStore) GetCbtQuestionDetail(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionDetailRow, error) {
	return db.GetCbtQuestionDetailRow{}, nil
}

func (f *fakeQuestionStore) CreateCbtQuestion(ctx context.Context, arg db.CreateCbtQuestionParams) (db.CbtQuestion, error) {
	return db.CbtQuestion{}, nil
}

func (f *fakeQuestionStore) UpdateCbtQuestion(ctx context.Context, arg db.UpdateCbtQuestionParams) (db.CbtQuestion, error) {
	f.updateParams = arg
	f.updateCalls++
	return db.CbtQuestion{
		ID:             arg.ID,
		QuestionText:   arg.QuestionText,
		WorkflowStatus: arg.WorkflowStatus,
		Status:         arg.Status,
		StemHtml:       arg.StemHtml,
		StimulusHtml:   arg.StimulusHtml,
	}, nil
}

func (f *fakeQuestionStore) DeleteCbtQuestion(ctx context.Context, id pgtype.UUID) error {
	return nil
}

func TestNormalizeQuestionInputSanitizesDangerousHTML(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		StemHTML:       `<p onclick="alert(1)">Halo</p><script>alert(2)</script>`,
		StimulusHTML:   `<img src="javascript:alert(1)" onerror="alert(1)">`,
		QuestionText:   "",
		Options:        []QuestionOption{{Label: "A", HTML: `<span onclick="x()">Aman</span>`}, {Label: "B", Text: "B"}},
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
	}

	got, err := normalizeQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	if got.StemHTML != `<p>Halo</p>` {
		t.Fatalf("StemHTML = %q, want sanitized paragraph", got.StemHTML)
	}
	if got.StimulusHTML != `<img>` {
		t.Fatalf("StimulusHTML = %q, want sanitized img", got.StimulusHTML)
	}
	if got.Options[0].HTML != `<span>Aman</span>` {
		t.Fatalf("Option HTML = %q, want sanitized span", got.Options[0].HTML)
	}
}

func TestValidateQuestionRequiresApprovedBeforePublish(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "advance",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal uji",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}},
		AnswerKey:      "A",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumPublished,
		WorkflowStatus: "draft",
	}
	_, err := normalizeQuestionInput(input)
	if err == nil {
		t.Fatal("normalizeQuestionInput() error = nil, want publish gating error")
	}
}

func TestSubmitReviewUpdatesWorkflowAndReviewer(t *testing.T) {
	store := &fakeQuestionStore{
		current: db.GetCbtQuestionRow{
			ID:             pgtype.UUID{Valid: true},
			SubjectID:      pgtype.UUID{Valid: true},
			QuestionText:   "Soal",
			QuestionType:   "multiple_choice",
			Options:        []byte(`[{"label":"A","text":"A"},{"label":"B","text":"B"}]`),
			AnswerKey:      "A",
			Difficulty:     db.CbtQuestionDifficultyEnumMedium,
			Status:         db.CbtQuestionStatusEnumDraft,
			WorkflowStatus: "draft",
		},
	}
	svc := &CbtQuestion{q: store}

	_, err := svc.SubmitReview(context.Background(), store.current.ID, "reviewer1", "cek redaksi")
	if err != nil {
		t.Fatalf("SubmitReview() error = %v", err)
	}
	if store.updateParams.WorkflowStatus != "review" {
		t.Fatalf("WorkflowStatus = %q, want review", store.updateParams.WorkflowStatus)
	}
	if store.updateParams.ReviewerUsername != "reviewer1" {
		t.Fatalf("ReviewerUsername = %q, want reviewer1", store.updateParams.ReviewerUsername)
	}
}

func TestNormalizeQuestionInputBeginnerDefaultsToDraft(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "multiple_choice",
		QuestionText:   "Soal mudah",
		Options:        []QuestionOption{{Label: "A", Text: "A"}, {Label: "B", Text: "B"}, {Label: "C", Text: "C"}, {Label: "D", Text: "D"}},
		AnswerKey:      "A",
		Status:         db.CbtQuestionStatusEnumPublished,
		WorkflowStatus: "approved",
		Difficulty:     db.CbtQuestionDifficultyEnumHard,
		WriterNotes:    "should be cleared",
	}

	got, err := normalizeQuestionInput(input)
	if err != nil {
		t.Fatalf("normalizeQuestionInput() error = %v", err)
	}
	if got.Status != db.CbtQuestionStatusEnumDraft {
		t.Fatalf("Status = %q, want draft", got.Status)
	}
	if got.WorkflowStatus != "draft" {
		t.Fatalf("WorkflowStatus = %q, want draft", got.WorkflowStatus)
	}
	if got.Difficulty != db.CbtQuestionDifficultyEnumMedium {
		t.Fatalf("Difficulty = %q, want medium", got.Difficulty)
	}
	if got.WriterNotes != "" {
		t.Fatalf("WriterNotes = %q, want cleared", got.WriterNotes)
	}
}

func TestNormalizeQuestionInputBeginnerRejectsAdvancedType(t *testing.T) {
	input := SaveCbtQuestionInput{
		SubjectID:      pgtype.UUID{Valid: true},
		AuthoringMode:  "beginner",
		QuestionType:   "short_answer",
		QuestionText:   "Jawab singkat",
		AnswerKey:      "42",
		Difficulty:     db.CbtQuestionDifficultyEnumMedium,
		Status:         db.CbtQuestionStatusEnumDraft,
		WorkflowStatus: "draft",
	}

	_, err := normalizeQuestionInput(input)
	if err == nil {
		t.Fatal("normalizeQuestionInput() error = nil, want beginner type rejection")
	}
}
