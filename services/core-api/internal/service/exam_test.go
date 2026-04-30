package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeExamStore struct {
	participant db.GetParticipantByTokenRow
	questions   []db.GetExamQuestionsRow
	answers     []db.GetParticipantAnswersRow
	assets      map[string][]db.CbtQuestionAsset
}

func (f *fakeExamStore) GetParticipantByToken(ctx context.Context, token string) (db.GetParticipantByTokenRow, error) {
	return f.participant, nil
}

func (f *fakeExamStore) UpdateParticipantLogin(ctx context.Context, arg db.UpdateParticipantLoginParams) error {
	return nil
}

func (f *fakeExamStore) InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error {
	return nil
}

func (f *fakeExamStore) GetExamQuestions(ctx context.Context, packageID pgtype.UUID) ([]db.GetExamQuestionsRow, error) {
	return f.questions, nil
}

func (f *fakeExamStore) UpdateParticipantQuestionOrder(ctx context.Context, arg db.UpdateParticipantQuestionOrderParams) error {
	return nil
}

func (f *fakeExamStore) GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error) {
	return f.answers, nil
}

func (f *fakeExamStore) GetCbtExamRoom(ctx context.Context, id pgtype.UUID) (db.CbtExamRoom, error) {
	return db.CbtExamRoom{}, nil
}

func (f *fakeExamStore) UpdateParticipantHeartbeat(ctx context.Context, participantID pgtype.UUID) error {
	return nil
}

func (f *fakeExamStore) IncrementParticipantAppSwitch(ctx context.Context, participantID pgtype.UUID) error {
	return nil
}

func (f *fakeExamStore) IncrementParticipantScreenshot(ctx context.Context, participantID pgtype.UUID) error {
	return nil
}

func (f *fakeExamStore) UpsertStudentAnswer(ctx context.Context, arg db.UpsertStudentAnswerParams) error {
	return nil
}

func (f *fakeExamStore) SubmitParticipantExam(ctx context.Context, id pgtype.UUID) (db.SubmitParticipantExamRow, error) {
	return db.SubmitParticipantExamRow{}, nil
}

func (f *fakeExamStore) ListCbtQuestionAssetsByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error) {
	return f.assets[pgUUIDString(questionID)], nil
}

func TestExamLoginIncludesMobileContractFields(t *testing.T) {
	ctx := context.Background()
	participantID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	sessionID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	packageID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	questionID := mustUUID(t, "44444444-4444-4444-4444-444444444444")

	start := mustTimestamp(t, "2026-05-01T08:00:00Z")
	end := mustTimestamp(t, "2026-05-01T10:00:00Z")
	joined := mustTimestamp(t, "2026-05-01T08:05:00Z")
	orderJSON, _ := json.Marshal([]string{pgUUIDString(questionID)})

	store := &fakeExamStore{
		participant: db.GetParticipantByTokenRow{
			ID:              participantID,
			SessionID:       sessionID,
			Token:           "a1b2c3d4",
			Nis:             "12345",
			Nama:            "Ahmad",
			SessionStatus:   db.CbtSessionStatusEnumActive,
			SessionTitle:    "Ujian Matematika",
			PackageTitle:    "Paket Matematika VII",
			ScheduledStart:  start,
			ScheduledEnd:    end,
			PackageID:       packageID,
			DurationMinutes: 90,
			JoinedAt:        joined,
			QuestionOrder:   orderJSON,
		},
		questions: []db.GetExamQuestionsRow{
			{
				ID:           questionID,
				Code:         "Q-1",
				QuestionText: "2 + 2 = ...",
				QuestionType: "multiple_choice",
				Options:      []byte(`[{"label":"A","text":"3"},{"label":"B","text":"4"}]`),
				StemHtml:     "<p>2 + 2 = ...</p>",
				StimulusHtml: "<p>Perhatikan gambar.</p>",
			},
		},
		answers: []db.GetParticipantAnswersRow{
			{QuestionID: questionID},
		},
		assets: map[string][]db.CbtQuestionAsset{
			pgUUIDString(questionID): {
				{
					ID:       mustUUID(t, "55555555-5555-5555-5555-555555555555"),
					MimeType: "image/png",
					Purpose:  "stimulus",
				},
				{
					ID:       mustUUID(t, "66666666-6666-6666-6666-666666666666"),
					MimeType: "audio/mpeg",
					Purpose:  "general",
				},
			},
		},
	}

	svc := &Exam{q: store}
	result, err := svc.Login(ctx, "a1b2c3d4", "device-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.Session.Title != "Ujian Matematika" {
		t.Fatalf("Session.Title = %q, want %q", result.Session.Title, "Ujian Matematika")
	}
	if result.AnsweredCount != 1 {
		t.Fatalf("AnsweredCount = %d, want 1", result.AnsweredCount)
	}
	if len(result.Questions) != 1 {
		t.Fatalf("len(Questions) = %d, want 1", len(result.Questions))
	}

	question := result.Questions[0]
	if question.StemHTML == "" || question.StimulusHTML == "" {
		t.Fatalf("expected rich content fields to be preserved, got stem=%q stimulus=%q", question.StemHTML, question.StimulusHTML)
	}
	if question.StimulusMediaURL != "/api/cbt/assets/55555555-5555-5555-5555-555555555555/file" {
		t.Fatalf("StimulusMediaURL = %q", question.StimulusMediaURL)
	}
	if question.StemAudioURL != "/api/cbt/assets/66666666-6666-6666-6666-666666666666/file" {
		t.Fatalf("StemAudioURL = %q", question.StemAudioURL)
	}
}

func TestExamStatusIncludesIsSubmitted(t *testing.T) {
	ctx := context.Background()
	participantID := mustUUID(t, "77777777-7777-7777-7777-777777777777")
	packageID := mustUUID(t, "88888888-8888-8888-8888-888888888888")
	submitted := mustTimestamp(t, "2026-05-01T09:15:00Z")

	store := &fakeExamStore{
		participant: db.GetParticipantByTokenRow{
			ID:              participantID,
			PackageID:       packageID,
			SubmittedAt:     submitted,
			DurationMinutes: 90,
		},
		questions: []db.GetExamQuestionsRow{{ID: mustUUID(t, "99999999-9999-9999-9999-999999999999")}},
		answers:   []db.GetParticipantAnswersRow{{QuestionID: mustUUID(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")}},
	}

	svc := &Exam{q: store}
	status, err := svc.GetStatus(ctx, store.participant)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if !status.IsSubmitted {
		t.Fatal("IsSubmitted = false, want true")
	}
	if status.SubmittedAt == "" {
		t.Fatal("SubmittedAt = empty, want RFC3339 timestamp")
	}
}

func mustUUID(t *testing.T, raw string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		t.Fatalf("uuid scan failed: %v", err)
	}
	return id
}

func mustTimestamp(t *testing.T, raw string) pgtype.Timestamptz {
	t.Helper()
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("time parse failed: %v", err)
	}
	var ts pgtype.Timestamptz
	if err := ts.Scan(value); err != nil {
		t.Fatalf("timestamp scan failed: %v", err)
	}
	return ts
}
