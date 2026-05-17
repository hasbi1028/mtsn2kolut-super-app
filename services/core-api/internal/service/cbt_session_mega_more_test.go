package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestMegaCbtSessionExamLoginRoomTokenAndAccessBranches(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)
	participant.RoomID = mustUUID(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1")
	participant.RoomToken = " Lab-1 "
	participant.QuestionOrder = []byte(`[]`)
	participant.RandomizeQuestions = false
	participant.RandomizeOptions = false
	questionID := mustUUID(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2")
	store := &fakeExamStore{
		participant: participant,
		questions: []db.GetExamQuestionsRow{{
			ID:           questionID,
			Code:         "Q1",
			QuestionText: "Satu?",
			QuestionType: "multiple_choice",
			Options:      []byte(`[{"label":"A","text":"Ya"}]`),
		}},
		answers: []db.GetParticipantAnswersRow{{QuestionID: questionID}},
		room:    db.CbtExamRoom{ID: participant.RoomID, RoomName: "Lab Komputer"},
		assets:  map[string][]db.CbtQuestionAsset{},
	}
	result, err := (&Exam{q: store}).Login(ctx, " token-with-space ", "lab-1", " device-a ", "10.0.0.1")
	if err != nil {
		t.Fatalf("Login(valid case-insensitive room token) error = %v", err)
	}
	if result.Room == nil || result.Room.RoomName != "Lab Komputer" || result.AnsweredCount != 1 || result.TotalQuestions != 1 {
		t.Fatalf("Login() result = %+v, want room and answer counts", result)
	}
	if store.updateLoginArg.ID != participant.ID || store.updateLoginArg.DeviceFingerprint.String != "device-a" || store.updateLoginArg.LoginIp.String != "10.0.0.1" {
		t.Fatalf("Login() update arg = %+v, want trimmed device and ip", store.updateLoginArg)
	}
	if len(store.events) != 1 || store.events[0].EventType != "login" {
		t.Fatalf("Login() events = %+v, want login only", store.events)
	}

	mismatchStore := &fakeExamStore{participant: participant}
	if _, err := (&Exam{q: mismatchStore}).Login(ctx, "token", "wrong", "device-a", "10.0.0.2"); !errors.Is(err, ErrRoomTokenMismatch) {
		t.Fatalf("Login(room token mismatch) error = %v, want ErrRoomTokenMismatch", err)
	}
	if len(mismatchStore.events) != 1 || mismatchStore.events[0].EventType != "exam_room_token_mismatch" {
		t.Fatalf("Login(room token mismatch) events = %+v, want mismatch event", mismatchStore.events)
	}
	var payload map[string]string
	if err := json.Unmarshal(mismatchStore.events[0].EventData, &payload); err != nil {
		t.Fatalf("mismatch event json error = %v", err)
	}
	if payload["reason"] != "room_token_mismatch" || payload["room_token_configured"] != "true" || payload["ip_hash"] == "10.0.0.2" || payload["room_id"] == "" {
		t.Fatalf("mismatch payload = %+v, want reason/configured/hashed ip/room id", payload)
	}
	if _, err := (&Exam{q: &fakeExamStore{participant: participant}}).Login(ctx, "token", " ", "device-a", "10.0.0.3"); !errors.Is(err, ErrRoomTokenRequired) {
		t.Fatalf("Login(missing room token) error = %v, want ErrRoomTokenRequired", err)
	}

	noRoom := participant
	noRoom.RoomID = pgtype.UUID{}
	if _, err := (&Exam{q: &fakeExamStore{participant: noRoom}}).Login(ctx, "token", "lab-1", "device-a", "10.0.0.4"); !errors.Is(err, ErrExamRoomRequired) {
		t.Fatalf("Login(no room) error = %v, want ErrExamRoomRequired", err)
	}
	future := participant
	future.ScheduledStart = pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}
	if _, err := (&Exam{q: &fakeExamStore{participant: future}}).Login(ctx, "token", "lab-1", "device-a", "10.0.0.5"); !errors.Is(err, ErrExamNotStarted) {
		t.Fatalf("Login(future start) error = %v, want ErrExamNotStarted", err)
	}
}

func TestMegaCbtSessionSubmitAnswerBranchesAndCanonicalization(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)
	questionID := mustUUID(t, "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1")
	participant.OptionOrder = []byte(`{"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1":["C","A","B"]}`)

	store := &fakeExamStore{participant: participant}
	if err := (&Exam{q: store}).SubmitAnswer(ctx, participant, questionID, "A, C"); err != nil {
		t.Fatalf("SubmitAnswer(success) error = %v", err)
	}
	if store.questionScopeArg.ID != participant.ID || store.questionScopeArg.QuestionID != questionID {
		t.Fatalf("SubmitAnswer() scope arg = %+v, want participant/question", store.questionScopeArg)
	}
	if store.answerArg.ParticipantID != participant.ID || store.answerArg.QuestionID != questionID || store.answerArg.Answer != "C,B" {
		t.Fatalf("SubmitAnswer() answer arg = %+v, want canonical labels C,B", store.answerArg)
	}
	if len(store.events) != 1 || store.events[0].EventType != "answer" {
		t.Fatalf("SubmitAnswer() events = %+v, want answer event", store.events)
	}
	var answerEvent map[string]string
	if err := json.Unmarshal(store.events[0].EventData, &answerEvent); err != nil {
		t.Fatalf("answer event json error = %v", err)
	}
	if answerEvent["question_id"] != pgUUIDString(questionID) || answerEvent["answer_length"] != "4" || answerEvent["answer_hash"] == "A, C" {
		t.Fatalf("answer event payload = %+v, want id/length/hashed answer", answerEvent)
	}

	if err := (&Exam{q: &fakeExamStore{questionOutsidePackage: true}}).SubmitAnswer(ctx, participant, questionID, "A"); !errors.Is(err, ErrExamQuestionScope) {
		t.Fatalf("SubmitAnswer(outside package) error = %v, want ErrExamQuestionScope", err)
	}
	boom := errors.New("scope failed")
	if err := (&Exam{q: &fakeExamStore{questionScopeErr: boom}}).SubmitAnswer(ctx, participant, questionID, "A"); !errors.Is(err, boom) {
		t.Fatalf("SubmitAnswer(scope error) error = %v, want %v", err, boom)
	}
	if err := (&Exam{q: &fakeExamStore{answerRowsSet: true, answerRows: 0}}).SubmitAnswer(ctx, participant, questionID, "A"); !errors.Is(err, ErrExamAlreadySubmit) {
		t.Fatalf("SubmitAnswer(zero rows) error = %v, want ErrExamAlreadySubmit", err)
	}
	locked := participant
	locked.LockedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if err := (&Exam{q: &fakeExamStore{}}).SubmitAnswer(ctx, locked, questionID, "A"); !errors.Is(err, ErrExamLocked) {
		t.Fatalf("SubmitAnswer(locked) error = %v, want ErrExamLocked", err)
	}
	closed := participant
	closed.JoinedAt = pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour), Valid: true}
	closed.DurationMinutes = 10
	closed.ScheduledEnd = pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}
	if err := (&Exam{q: &fakeExamStore{}}).SubmitAnswer(ctx, closed, questionID, "A"); !errors.Is(err, ErrExamWindowClosed) {
		t.Fatalf("SubmitAnswer(duration closed) error = %v, want ErrExamWindowClosed", err)
	}
}

func TestMegaCbtSessionSubmitAndStatusCommandBranches(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)
	participantID := participant.ID
	questionID := mustUUID(t, "cccccccc-cccc-4ccc-8ccc-ccccccccccc1")
	submittedAt := pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}
	store := &fakeExamStore{
		submitRow: db.SubmitParticipantExamRow{ID: participantID, SubmittedAt: submittedAt, Score: pgNumeric(80)},
	}
	if err := (&Exam{q: store}).Submit(ctx, participant); err != nil {
		t.Fatalf("Submit(success) error = %v", err)
	}
	if store.submitID != participantID || store.correctnessID != participantID || len(store.events) != 1 || store.events[0].EventType != "submit" {
		t.Fatalf("Submit(success) calls submit=%v correctness=%v events=%+v", store.submitID, store.correctnessID, store.events)
	}

	if err := (&Exam{q: &fakeExamStore{submitErr: pgx.ErrNoRows}}).Submit(ctx, participant); !errors.Is(err, ErrExamAlreadySubmit) {
		t.Fatalf("Submit(no rows) error = %v, want ErrExamAlreadySubmit", err)
	}
	correctnessErr := errors.New("correctness failed")
	store = &fakeExamStore{submitRow: db.SubmitParticipantExamRow{ID: participantID, SubmittedAt: submittedAt}, correctnessErr: correctnessErr}
	if err := (&Exam{q: store}).Submit(ctx, participant); !errors.Is(err, correctnessErr) {
		t.Fatalf("Submit(correctness error) error = %v, want %v", err, correctnessErr)
	}
	if len(store.events) != 0 {
		t.Fatalf("Submit(correctness error) events = %+v, want none", store.events)
	}

	commandID := mustUUID(t, "cccccccc-cccc-4ccc-8ccc-ccccccccccc2")
	commandAt := pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Minute), Valid: true}
	store = &fakeExamStore{
		questions: []db.GetExamQuestionsRow{{ID: questionID}},
		answers:   []db.GetParticipantAnswersRow{{QuestionID: questionID}},
		commandRows: []db.ListPendingParticipantCommandsRow{{
			ID:        commandID,
			EventType: "participant_command",
			EventData: []byte(`{"command_type":"reconnect","message":"Login ulang sekarang","severity":"warning","actor":"proctor-1"}`),
			CreatedAt: commandAt,
		}},
	}
	status, err := (&Exam{q: store}).GetStatus(ctx, participant)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.AnsweredCount != 1 || status.TotalQuestions != 1 || len(status.Commands) != 1 || status.Commands[0].LegacyLabel != "Login ulang" || store.commandParticipantID != participantID {
		t.Fatalf("GetStatus() = %+v commandParticipantID=%v, want counts and command", status, store.commandParticipantID)
	}
	cmdErr := errors.New("commands failed")
	if _, err := (&Exam{q: &fakeExamStore{commandErr: cmdErr}}).GetStatus(ctx, participant); !errors.Is(err, cmdErr) {
		t.Fatalf("GetStatus(command error) error = %v, want %v", err, cmdErr)
	}
	if err := (&Exam{q: &fakeExamStore{}}).AcknowledgeCommand(ctx, participantID, "  ", "seen"); err == nil || !strings.Contains(err.Error(), "command id required") {
		t.Fatalf("AcknowledgeCommand(blank id) error = %v, want command id required", err)
	}
}

func TestMegaCbtSessionResetRuntimeAccessAndAnswerHelperBranches(t *testing.T) {
	ctx := context.Background()
	participantID := cbtSessionTestUUID(81)
	sessionID := cbtSessionTestUUID(82)
	answerID := cbtSessionTestUUID(83)
	teacherID := cbtSessionTestUUID(84)
	store := &fakeCbtSessionStore{}
	if err := (&CbtSession{q: store}).ResetParticipantRuntimeAccess(ctx, participantID, "  proctor-a  "); err != nil {
		t.Fatalf("ResetParticipantRuntimeAccess() error = %v", err)
	}
	if store.resetParticipantID != participantID || len(store.insertedEvents) != 1 || store.insertedEvents[0].EventType != "proctor_reset_access" {
		t.Fatalf("ResetParticipantRuntimeAccess() reset=%v events=%+v, want reset and event", store.resetParticipantID, store.insertedEvents)
	}
	var payload map[string]string
	if err := json.Unmarshal(store.insertedEvents[0].EventData, &payload); err != nil || payload["actor"] != "  proctor-a  " {
		t.Fatalf("ResetParticipantRuntimeAccess() payload=%+v err=%v, want actor preserved", payload, err)
	}
	resetErr := errors.New("reset failed")
	store = &fakeCbtSessionStore{resetParticipantErr: resetErr}
	if err := (&CbtSession{q: store}).ResetParticipantRuntimeAccess(ctx, participantID, "actor"); !errors.Is(err, resetErr) {
		t.Fatalf("ResetParticipantRuntimeAccess(reset error) error = %v, want %v", err, resetErr)
	}
	if len(store.insertedEvents) != 0 {
		t.Fatalf("ResetParticipantRuntimeAccess(reset error) inserted events = %+v, want none", store.insertedEvents)
	}
	insertErr := errors.New("insert failed")
	store = &fakeCbtSessionStore{insertEventErr: insertErr}
	if err := (&CbtSession{q: store}).ResetParticipantRuntimeAccess(ctx, participantID, "actor"); !errors.Is(err, insertErr) {
		t.Fatalf("ResetParticipantRuntimeAccess(insert error) error = %v, want %v", err, insertErr)
	}

	store = &fakeCbtSessionStore{sessionParticipant: true, sessionAnswer: true}
	if ok, err := (&CbtSession{q: store}).HasParticipantByTeacher(ctx, sessionID, participantID, teacherID); err != nil || !ok || store.teacherParticipantArg.TeacherEmployeeID != teacherID || store.teacherParticipantArg.ParticipantID != participantID {
		t.Fatalf("HasParticipantByTeacher() ok=%v err=%v arg=%+v", ok, err, store.teacherParticipantArg)
	}
	if ok, err := (&CbtSession{q: store}).HasAnswerByTeacher(ctx, sessionID, answerID, teacherID); err != nil || !ok || store.teacherAnswerArg.TeacherEmployeeID != teacherID || store.teacherAnswerArg.AnswerID != answerID {
		t.Fatalf("HasAnswerByTeacher() ok=%v err=%v arg=%+v", ok, err, store.teacherAnswerArg)
	}
	if rows, err := (&CbtSession{q: &fakeCbtSessionStore{resultsErr: errors.New("results failed")}}).GetResults(ctx, sessionID); err == nil || rows != nil {
		t.Fatalf("GetResults(error) rows=%+v err=%v, want nil error", rows, err)
	}
	if rows, err := (&CbtSession{q: &fakeCbtSessionStore{answersErr: errors.New("answers failed")}}).GetParticipantAnswers(ctx, participantID); err == nil || rows != nil {
		t.Fatalf("GetParticipantAnswers(error) rows=%+v err=%v, want nil error", rows, err)
	}
}

func TestMegaCbtSessionProctorScopeAndActionBranches(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(91)
	participantID := cbtSessionTestUUID(92)
	roomID := cbtSessionTestUUID(93)
	eventID := cbtSessionTestUUID(94)
	actor := CbtProctorActor{UserID: cbtSessionTestUUID(95), EmployeeID: cbtSessionTestUUID(96), Username: "proctor", RequestID: "req-1", SourceIP: "127.0.0.1"}

	store := &fakeCbtSessionStore{participantProctorScopeRow: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID}}
	scope, err := (&CbtSession{q: store}).GetParticipantProctorScope(ctx, sessionID, participantID)
	if err != nil || scope.ParticipantID != participantID || store.participantProctorScopeArg.SessionID != sessionID || store.participantProctorScopeArg.ParticipantID != participantID {
		t.Fatalf("GetParticipantProctorScope() scope=%+v err=%v arg=%+v", scope, err, store.participantProctorScopeArg)
	}
	store.proctorEventScopeRow = db.GetCbtProctorEventScopeRow{ID: eventID, ParticipantID: participantID, Severity: string(ProctorSeverityCritical), RequiresNote: true}
	if _, err := (&CbtSession{q: store}).AcknowledgeProctorEventByID(ctx, eventID, actor, " "); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("AcknowledgeProctorEventByID(critical requires note) error = %v, want ErrBadRequest", err)
	}
	store.proctorEventScopeRow = db.GetCbtProctorEventScopeRow{ID: eventID, ParticipantID: participantID, Severity: "warning", RequiresNote: false}
	ack, err := (&CbtSession{q: store}).AcknowledgeProctorEventByID(ctx, eventID, actor, " noted ")
	if err != nil {
		t.Fatalf("AcknowledgeProctorEventByID(success) error = %v", err)
	}
	if ack.ID != eventID || store.ackProctorEventArg.AcknowledgeNote != "noted" || len(store.proctorEventArgs) != 1 || store.proctorEventArgs[0].ActorUsernameSnapshot != "proctor" {
		t.Fatalf("AcknowledgeProctorEventByID(success) ack=%+v arg=%+v events=%+v", ack, store.ackProctorEventArg, store.proctorEventArgs)
	}

	store = &fakeCbtSessionStore{
		participantProctorScopeRow: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID},
		participantRiskRow:         db.GetCbtParticipantRiskForUpdateRow{ID: participantID, SessionID: sessionID, RoomID: roomID, PendingAnswerCount: 2, SyncState: "pending"},
	}
	_, err = (&CbtSession{q: store}).ExecuteProctorAction(ctx, CbtProctorActionInput{SessionID: sessionID, ParticipantID: participantID, ActionType: "force_submit", Reason: "selesai", Notes: "tanpa kata kunci"}, actor)
	if !errors.Is(err, domain.ErrBadRequest) || store.forceSubmitArg.ID.Valid {
		t.Fatalf("ExecuteProctorAction(force submit pending without sync warning) err=%v forceArg=%+v, want bad request before submit", err, store.forceSubmitArg)
	}
	result, err := (&CbtSession{q: store}).ExecuteProctorAction(ctx, CbtProctorActionInput{SessionID: sessionID, ParticipantID: participantID, EventID: eventID, ActionType: "force_submit", Reason: "risiko sinkron pending", Notes: "disetujui"}, actor)
	if err != nil {
		t.Fatalf("ExecuteProctorAction(force submit success) error = %v", err)
	}
	if result.Action.ActionType != "force_submit" || store.forceSubmitArg.ID != participantID || store.forceSubmitArg.SessionID != sessionID || store.proctorCorrectnessID.Valid {
		t.Fatalf("ExecuteProctorAction(force submit success) result=%+v forceArg=%+v proctorCorrectness=%v", result, store.forceSubmitArg, store.proctorCorrectnessID)
	}
	if store.participantCorrectnessID != participantID || len(store.proctorEventArgs) != 1 || store.proctorEventArgs[0].EventType != "proctor_force_submit" {
		t.Fatalf("ExecuteProctorAction(force submit success) correctness=%v events=%+v", store.participantCorrectnessID, store.proctorEventArgs)
	}

	store = &fakeCbtSessionStore{
		participantProctorScopeRow: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID},
		participantRiskRow:         db.GetCbtParticipantRiskForUpdateRow{ID: participantID, SessionID: sessionID, RoomID: roomID},
	}
	if result, err := (&CbtSession{q: store}).ExecuteProctorAction(ctx, CbtProctorActionInput{SessionID: sessionID, ParticipantID: participantID, ActionType: "reset_device_binding", Reason: "ganti perangkat", Notes: "diverifikasi pengawas"}, actor); err != nil || result.Action.ActionType != "reset_device_binding" || store.resetProctorID != participantID {
		t.Fatalf("ExecuteProctorAction(reset device) result=%+v err=%v resetID=%v", result, err, store.resetProctorID)
	}
}
