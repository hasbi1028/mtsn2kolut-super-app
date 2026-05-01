package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeExamStore struct {
	participant    db.GetParticipantByTokenRow
	participantErr error
	updateLoginArg db.UpdateParticipantLoginParams
	updateLoginErr error
	events         []db.InsertParticipantEventParams
	eventErr       error
	questions      []db.GetExamQuestionsRow
	questionsErr   error
	orderArg       db.UpdateParticipantQuestionOrderParams
	answers        []db.GetParticipantAnswersRow
	answersErr     error
	room           db.CbtExamRoom
	roomErr        error
	heartbeatID    pgtype.UUID
	heartbeatErr   error
	appSwitchID    pgtype.UUID
	appSwitchErr   error
	screenshotID   pgtype.UUID
	screenshotErr  error
	answerArg      db.UpsertStudentAnswerParams
	answerErr      error
	submitRow      db.SubmitParticipantExamRow
	submitID       pgtype.UUID
	submitErr      error
	assets         map[string][]db.CbtQuestionAsset
	assetsErr      error
}

func (f *fakeExamStore) GetParticipantByToken(ctx context.Context, token string) (db.GetParticipantByTokenRow, error) {
	return f.participant, f.participantErr
}

func (f *fakeExamStore) UpdateParticipantLogin(ctx context.Context, arg db.UpdateParticipantLoginParams) error {
	f.updateLoginArg = arg
	return f.updateLoginErr
}

func (f *fakeExamStore) InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error {
	f.events = append(f.events, arg)
	return f.eventErr
}

func (f *fakeExamStore) GetExamQuestions(ctx context.Context, packageID pgtype.UUID) ([]db.GetExamQuestionsRow, error) {
	return f.questions, f.questionsErr
}

func (f *fakeExamStore) UpdateParticipantQuestionOrder(ctx context.Context, arg db.UpdateParticipantQuestionOrderParams) error {
	f.orderArg = arg
	return nil
}

func (f *fakeExamStore) GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error) {
	return f.answers, f.answersErr
}

func (f *fakeExamStore) GetCbtExamRoom(ctx context.Context, id pgtype.UUID) (db.CbtExamRoom, error) {
	return f.room, f.roomErr
}

func (f *fakeExamStore) UpdateParticipantHeartbeat(ctx context.Context, participantID pgtype.UUID) error {
	f.heartbeatID = participantID
	return f.heartbeatErr
}

func (f *fakeExamStore) IncrementParticipantAppSwitch(ctx context.Context, participantID pgtype.UUID) error {
	f.appSwitchID = participantID
	return f.appSwitchErr
}

func (f *fakeExamStore) IncrementParticipantScreenshot(ctx context.Context, participantID pgtype.UUID) error {
	f.screenshotID = participantID
	return f.screenshotErr
}

func (f *fakeExamStore) UpsertStudentAnswer(ctx context.Context, arg db.UpsertStudentAnswerParams) error {
	f.answerArg = arg
	return f.answerErr
}

func (f *fakeExamStore) SubmitParticipantExam(ctx context.Context, id pgtype.UUID) (db.SubmitParticipantExamRow, error) {
	f.submitID = id
	return f.submitRow, f.submitErr
}

func (f *fakeExamStore) ListCbtQuestionAssetsByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error) {
	return f.assets[pgUUIDString(questionID)], f.assetsErr
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

func TestExamLoginRejectsInvalidStatesAndPropagatesErrors(t *testing.T) {
	ctx := context.Background()
	questionID := mustUUID(t, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	base := examActiveParticipant(t)

	storeErr := errors.New("store failed")
	svc := &Exam{q: &fakeExamStore{participantErr: storeErr}}
	if _, err := svc.Login(ctx, "missing", "device-1", "127.0.0.1"); !errors.Is(err, ErrExamNotFound) {
		t.Fatalf("Login(missing token) error = %v, want ErrExamNotFound", err)
	}

	inactive := base
	inactive.SessionStatus = db.CbtSessionStatusEnumScheduled
	svc = &Exam{q: &fakeExamStore{participant: inactive}}
	if _, err := svc.Login(ctx, "token", "device-1", "127.0.0.1"); !errors.Is(err, ErrExamNotActive) {
		t.Fatalf("Login(inactive) error = %v, want ErrExamNotActive", err)
	}

	bound := base
	bound.DeviceFingerprint = pgtype.Text{String: "other-device", Valid: true}
	svc = &Exam{q: &fakeExamStore{participant: bound}}
	if _, err := svc.Login(ctx, "token", "device-1", "127.0.0.1"); !errors.Is(err, ErrDeviceMismatch) {
		t.Fatalf("Login(device mismatch) error = %v, want ErrDeviceMismatch", err)
	}

	loginErr := errors.New("login update failed")
	svc = &Exam{q: &fakeExamStore{participant: base, updateLoginErr: loginErr}}
	if _, err := svc.Login(ctx, "token", "device-1", "127.0.0.1"); !errors.Is(err, loginErr) {
		t.Fatalf("Login(update error) error = %v, want %v", err, loginErr)
	}

	questionsErr := errors.New("questions failed")
	svc = &Exam{q: &fakeExamStore{participant: base, questionsErr: questionsErr}}
	if _, err := svc.Login(ctx, "token", "device-1", "127.0.0.1"); !errors.Is(err, questionsErr) {
		t.Fatalf("Login(questions error) error = %v, want %v", err, questionsErr)
	}

	roomID := mustUUID(t, "cccccccc-cccc-cccc-cccc-cccccccccccc")
	firstQuestionID := mustUUID(t, "dddddddd-dddd-dddd-dddd-dddddddddddd")
	secondQuestionID := questionID
	withRoom := base
	withRoom.RoomID = roomID
	withRoom.QuestionOrder = nil
	store := &fakeExamStore{
		participant: withRoom,
		questions: []db.GetExamQuestionsRow{
			{ID: firstQuestionID, Code: "Q-1", QuestionText: "Satu", QuestionType: "essay"},
			{ID: secondQuestionID, Code: "Q-2", QuestionText: "Dua", QuestionType: "multiple_choice"},
		},
		answersErr: errors.New("answers ignored"),
		room:       db.CbtExamRoom{ID: roomID, RoomName: "Ruang 1"},
	}
	svc = &Exam{q: store}
	result, err := svc.Login(ctx, "token", "device-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login(room/order) error = %v", err)
	}
	if store.updateLoginArg.ID != withRoom.ID || store.updateLoginArg.DeviceFingerprint.String != "device-1" || store.updateLoginArg.LoginIp.String != "127.0.0.1" {
		t.Fatalf("UpdateParticipantLogin arg = %+v, want bound device and IP", store.updateLoginArg)
	}
	if store.orderArg.ID != withRoom.ID {
		t.Fatalf("UpdateParticipantQuestionOrder id = %v, want %v", store.orderArg.ID, withRoom.ID)
	}
	var savedOrder []string
	if err := json.Unmarshal(store.orderArg.QuestionOrder, &savedOrder); err != nil {
		t.Fatalf("saved question order unmarshal error = %v", err)
	}
	if len(savedOrder) != 2 {
		t.Fatalf("saved question order len = %d, want 2", len(savedOrder))
	}
	if result.Room == nil || result.Room.RoomName != "Ruang 1" {
		t.Fatalf("Login() room = %+v, want Ruang 1", result.Room)
	}
	if result.AnsweredCount != 0 || result.TotalQuestions != 2 {
		t.Fatalf("Login() counts = %d/%d, want 0/2", result.AnsweredCount, result.TotalQuestions)
	}
	if len(store.events) == 0 || store.events[0].EventType != "login" {
		t.Fatalf("events = %+v, want login event", store.events)
	}
}

func TestExamOperationalMethodsHandleErrorsAndEvents(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)
	questionID := mustUUID(t, "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")

	updateErr := errors.New("heartbeat failed")
	svc := &Exam{q: &fakeExamStore{heartbeatErr: updateErr}}
	if err := svc.Heartbeat(ctx, participant.ID); !errors.Is(err, updateErr) {
		t.Fatalf("Heartbeat(update error) = %v, want %v", err, updateErr)
	}

	eventErr := errors.New("event failed")
	svc = &Exam{q: &fakeExamStore{eventErr: eventErr}}
	if err := svc.Heartbeat(ctx, participant.ID); !errors.Is(err, eventErr) {
		t.Fatalf("Heartbeat(event error) = %v, want %v", err, eventErr)
	}

	store := &fakeExamStore{}
	svc = &Exam{q: store}
	if err := svc.Heartbeat(ctx, participant.ID); err != nil {
		t.Fatalf("Heartbeat() error = %v", err)
	}
	if store.heartbeatID != participant.ID || len(store.events) != 1 || store.events[0].EventType != "heartbeat" {
		t.Fatalf("Heartbeat store state = id %v events %+v, want heartbeat event", store.heartbeatID, store.events)
	}

	store = &fakeExamStore{appSwitchErr: errors.New("switch counter ignored"), screenshotErr: errors.New("screenshot counter ignored")}
	svc = &Exam{q: store}
	if err := svc.RecordClientEvent(ctx, participant.ID, "app_switch", map[string]any{"count": 1}); err != nil {
		t.Fatalf("RecordClientEvent(app_switch) error = %v", err)
	}
	if err := svc.RecordClientEvent(ctx, participant.ID, "screenshot_attempt", nil); err != nil {
		t.Fatalf("RecordClientEvent(screenshot_attempt) error = %v", err)
	}
	if store.appSwitchID != participant.ID || store.screenshotID != participant.ID || len(store.events) != 2 {
		t.Fatalf("RecordClientEvent counters/events = app %v screenshot %v events %+v", store.appSwitchID, store.screenshotID, store.events)
	}

	svc = &Exam{q: &fakeExamStore{eventErr: eventErr}}
	if err := svc.RecordClientEvent(ctx, participant.ID, "other", nil); !errors.Is(err, eventErr) {
		t.Fatalf("RecordClientEvent(event error) = %v, want %v", err, eventErr)
	}

	submitted := participant
	submitted.SubmittedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if err := svc.SubmitAnswer(ctx, submitted, questionID, "A"); !errors.Is(err, ErrExamAlreadySubmit) {
		t.Fatalf("SubmitAnswer(submitted) = %v, want ErrExamAlreadySubmit", err)
	}

	closed := participant
	closed.ScheduledEnd = pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}
	if err := svc.SubmitAnswer(ctx, closed, questionID, "A"); !errors.Is(err, ErrExamWindowClosed) {
		t.Fatalf("SubmitAnswer(closed) = %v, want ErrExamWindowClosed", err)
	}

	answerErr := errors.New("answer failed")
	svc = &Exam{q: &fakeExamStore{answerErr: answerErr}}
	if err := svc.SubmitAnswer(ctx, participant, questionID, "A"); !errors.Is(err, answerErr) {
		t.Fatalf("SubmitAnswer(answer error) = %v, want %v", err, answerErr)
	}
	svc = &Exam{q: &fakeExamStore{eventErr: eventErr}}
	if err := svc.SubmitAnswer(ctx, participant, questionID, "A"); !errors.Is(err, eventErr) {
		t.Fatalf("SubmitAnswer(event error) = %v, want %v", err, eventErr)
	}

	store = &fakeExamStore{}
	svc = &Exam{q: store}
	if err := svc.SubmitAnswer(ctx, participant, questionID, "B"); err != nil {
		t.Fatalf("SubmitAnswer() error = %v", err)
	}
	if store.answerArg.ParticipantID != participant.ID || store.answerArg.QuestionID != questionID || store.answerArg.Answer != "B" {
		t.Fatalf("UpsertStudentAnswer arg = %+v, want submitted answer", store.answerArg)
	}
	if len(store.events) != 1 || store.events[0].EventType != "answer" || len(store.events[0].EventData) == 0 {
		t.Fatalf("SubmitAnswer events = %+v, want answer event with data", store.events)
	}

	if err := svc.Submit(ctx, submitted); !errors.Is(err, ErrExamAlreadySubmit) {
		t.Fatalf("Submit(submitted) = %v, want ErrExamAlreadySubmit", err)
	}
	submitErr := errors.New("submit failed")
	svc = &Exam{q: &fakeExamStore{submitErr: submitErr}}
	if err := svc.Submit(ctx, participant); !errors.Is(err, submitErr) {
		t.Fatalf("Submit(submit error) = %v, want %v", err, submitErr)
	}
	svc = &Exam{q: &fakeExamStore{eventErr: eventErr, submitRow: db.SubmitParticipantExamRow{ID: participant.ID, SubmittedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}}}
	if err := svc.Submit(ctx, participant); !errors.Is(err, eventErr) {
		t.Fatalf("Submit(event error) = %v, want %v", err, eventErr)
	}
	store = &fakeExamStore{submitRow: db.SubmitParticipantExamRow{ID: participant.ID, SubmittedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}}
	svc = &Exam{q: store}
	if err := svc.Submit(ctx, participant); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if store.submitID != participant.ID || len(store.events) != 1 || store.events[0].EventType != "submit" {
		t.Fatalf("Submit store state = id %v events %+v, want submit event", store.submitID, store.events)
	}
}

func TestExamStatusAndHelpersCoverFallbackBranches(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)
	questionsErr := errors.New("questions failed")
	svc := &Exam{q: &fakeExamStore{questionsErr: questionsErr}}
	if _, err := svc.GetStatus(ctx, participant); !errors.Is(err, questionsErr) {
		t.Fatalf("GetStatus(questions error) = %v, want %v", err, questionsErr)
	}

	firstID := mustUUID(t, "abababab-abab-abab-abab-abababababab")
	secondID := mustUUID(t, "babababa-baba-baba-baba-babababababa")
	rows := []db.GetExamQuestionsRow{
		{ID: firstID, Code: "Q-1"},
		{ID: secondID, Code: "Q-2"},
	}
	if got := orderQuestions(rows, []byte("not-json")); len(got) != 2 || got[0].ID != firstID || got[1].ID != secondID {
		t.Fatalf("orderQuestions(invalid json) = %+v, want original order", got)
	}
	orderJSON, _ := json.Marshal([]string{pgUUIDString(secondID), "missing"})
	if got := orderQuestions(rows, orderJSON); len(got) != 1 || got[0].ID != secondID {
		t.Fatalf("orderQuestions(partial) = %+v, want only second question", got)
	}
	if got := orderQuestions(rows, nil); len(got) != 2 {
		t.Fatalf("orderQuestions(empty) len = %d, want 2", len(got))
	}

	now := time.Now()
	endBeforeDuration := calcRemaining(
		pgtype.Timestamptz{Time: now.Add(5 * time.Minute), Valid: true},
		30,
		pgtype.Timestamptz{Time: now, Valid: true},
	)
	durationOnly := calcRemaining(pgtype.Timestamptz{}, 30, pgtype.Timestamptz{Time: now, Valid: true})
	endOnly := calcRemaining(pgtype.Timestamptz{Time: now.Add(10 * time.Minute), Valid: true}, 0, pgtype.Timestamptz{})
	past := calcRemaining(pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true}, 0, pgtype.Timestamptz{})
	none := calcRemaining(pgtype.Timestamptz{}, 0, pgtype.Timestamptz{})
	if endBeforeDuration <= 0 || durationOnly <= 0 || endOnly <= 0 || past != 0 || none != 0 {
		t.Fatalf("calcRemaining branches = %d/%d/%d/%d/%d, want positive positive positive 0 0", endBeforeDuration, durationOnly, endOnly, past, none)
	}

	if stemMedia, stimulusMedia, stemAudio, stimulusAudio := resolveExamQuestionAssets(nil, ctx, firstID); stemMedia != "" || stimulusMedia != "" || stemAudio != "" || stimulusAudio != "" {
		t.Fatalf("resolveExamQuestionAssets(nil) = %q/%q/%q/%q, want empty", stemMedia, stimulusMedia, stemAudio, stimulusAudio)
	}
	store := &fakeExamStore{assetsErr: errors.New("assets failed")}
	if stemMedia, stimulusMedia, stemAudio, stimulusAudio := resolveExamQuestionAssets(store, ctx, firstID); stemMedia != "" || stimulusMedia != "" || stemAudio != "" || stimulusAudio != "" {
		t.Fatalf("resolveExamQuestionAssets(error) = %q/%q/%q/%q, want empty", stemMedia, stimulusMedia, stemAudio, stimulusAudio)
	}

	imageOne := mustUUID(t, "11111111-2222-3333-4444-555555555555")
	imageTwo := mustUUID(t, "22222222-3333-4444-5555-666666666666")
	audioOne := mustUUID(t, "33333333-4444-5555-6666-777777777777")
	audioTwo := mustUUID(t, "44444444-5555-6666-7777-888888888888")
	store = &fakeExamStore{assets: map[string][]db.CbtQuestionAsset{
		pgUUIDString(firstID): {
			{ID: imageOne, MimeType: "image/png", Purpose: "stem"},
			{ID: imageTwo, MimeType: "image/jpeg", Purpose: "other"},
			{ID: audioOne, MimeType: "audio/mpeg", Purpose: "stem"},
			{ID: audioTwo, MimeType: "audio/ogg", Purpose: "other"},
			{ID: mustUUID(t, "55555555-6666-7777-8888-999999999999"), MimeType: "application/pdf", Purpose: "stimulus"},
		},
	}}
	stemMedia, stimulusMedia, stemAudio, stimulusAudio := resolveExamQuestionAssets(store, ctx, firstID)
	if stemMedia != "/api/cbt/assets/11111111-2222-3333-4444-555555555555/file" ||
		stimulusMedia != "/api/cbt/assets/22222222-3333-4444-5555-666666666666/file" ||
		stemAudio != "/api/cbt/assets/33333333-4444-5555-6666-777777777777/file" ||
		stimulusAudio != "/api/cbt/assets/44444444-5555-6666-7777-888888888888/file" {
		t.Fatalf("resolveExamQuestionAssets() = %q/%q/%q/%q, want fallback media/audio URLs", stemMedia, stimulusMedia, stemAudio, stimulusAudio)
	}

	examQuestions := toExamQuestions(store, ctx, []db.GetExamQuestionsRow{{ID: firstID, Code: "Q-1"}})
	if len(examQuestions) != 1 || string(examQuestions[0].Options) != "[]" || examQuestions[0].StemMediaURL == "" || examQuestions[0].StimulusAudioURL == "" {
		t.Fatalf("toExamQuestions() = %+v, want default options and resolved assets", examQuestions)
	}
	if got := firstNonEmpty(" ", "\t", "fallback"); got != "fallback" {
		t.Fatalf("firstNonEmpty() = %q, want fallback", got)
	}
	if got := firstNonEmpty(" ", "\t"); got != "" {
		t.Fatalf("firstNonEmpty(empty) = %q, want empty", got)
	}
	if got := marshalJSON(nil); got != nil {
		t.Fatalf("marshalJSON(nil) = %q, want nil", got)
	}
}

func examActiveParticipant(t *testing.T) db.GetParticipantByTokenRow {
	t.Helper()
	now := time.Now()
	return db.GetParticipantByTokenRow{
		ID:              mustUUID(t, "01010101-0101-0101-0101-010101010101"),
		SessionID:       mustUUID(t, "02020202-0202-0202-0202-020202020202"),
		StudentID:       mustUUID(t, "03030303-0303-0303-0303-030303030303"),
		Token:           "token",
		Nis:             "12345",
		Nama:            "Ahmad",
		SessionStatus:   db.CbtSessionStatusEnumActive,
		SessionTitle:    "Sesi Ujian",
		ScheduledStart:  pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true},
		ScheduledEnd:    pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
		PackageID:       mustUUID(t, "04040404-0404-0404-0404-040404040404"),
		PackageTitle:    "Paket Ujian",
		DurationMinutes: 90,
		JoinedAt:        pgtype.Timestamptz{Time: now, Valid: true},
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
