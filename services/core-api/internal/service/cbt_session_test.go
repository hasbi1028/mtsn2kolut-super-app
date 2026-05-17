package service

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func cbtSessionTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func TestCbtSessionNormalizeScopeMixAndAssignment(t *testing.T) {
	scopeTests := []struct {
		name string
		in   string
		want string
	}{
		{name: "class", in: "class", want: "class"},
		{name: "grade", in: "grade", want: "grade"},
		{name: "school", in: "school", want: "school"},
		{name: "custom", in: "custom", want: "custom"},
		{name: "fallback", in: "invalid", want: "class"},
	}
	for _, tt := range scopeTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeScopeType(tt.in); got != tt.want {
				t.Fatalf("normalizeScopeType(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}

	mixTests := []struct {
		name      string
		value     string
		scopeType string
		want      string
	}{
		{name: "class forced", value: "mixed_scope", scopeType: "class", want: "same_class"},
		{name: "grade forced", value: "same_class", scopeType: "grade", want: "same_grade"},
		{name: "school forced", value: "same_class", scopeType: "school", want: "mixed_scope"},
		{name: "custom forced", value: "same_grade", scopeType: "custom", want: "mixed_scope"},
		{name: "unknown accepts explicit", value: "same_class", scopeType: "unknown", want: "same_class"},
		{name: "unknown fallback", value: "bad", scopeType: "unknown", want: "same_grade"},
	}
	for _, tt := range mixTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeMixPolicy(tt.value, tt.scopeType); got != tt.want {
				t.Fatalf("normalizeMixPolicy(%q, %q) = %q, want %q", tt.value, tt.scopeType, got, tt.want)
			}
		})
	}

	modeTests := []struct {
		name string
		in   string
		want string
	}{
		{name: "manual", in: "manual", want: "manual"},
		{name: "random balanced", in: "random_balanced", want: "random_balanced"},
		{name: "random by gender", in: "random_by_gender", want: "random_by_gender"},
		{name: "random by accommodation", in: "random_by_accommodation", want: "random_by_accommodation"},
		{name: "fallback", in: "bad", want: "random_balanced"},
	}
	for _, tt := range modeTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeAssignmentMode(tt.in); got != tt.want {
				t.Fatalf("normalizeAssignmentMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCbtSessionPgNumeric(t *testing.T) {
	n := pgNumeric(87.5)
	if !n.Valid {
		t.Fatal("pgNumeric() valid = false, want true")
	}
	f, _ := n.Int.Float64()
	scale := new(big.Float).SetInt64(1)
	for i := int32(0); i < -n.Exp; i++ {
		scale.Mul(scale, big.NewFloat(10))
	}
	divisor, _ := scale.Float64()
	if got := f / divisor; got != 87.5 {
		t.Fatalf("pgNumeric(87.5) = %v, want 87.5", got)
	}
}

func TestValidateCbtSessionActivationReadiness(t *testing.T) {
	valid := db.GetCbtSessionRoomReadinessRow{
		RoomCount:                  2,
		TotalCapacity:              60,
		ParticipantCount:           50,
		AssignedParticipantCount:   50,
		UnassignedParticipantCount: 0,
		MissingSeatCount:           0,
		RoomsWithoutProctor:        0,
		ProctorAssignmentCount:     2,
	}

	tests := []struct {
		name    string
		row     db.GetCbtSessionRoomReadinessRow
		wantErr string
	}{
		{name: "ready", row: valid},
		{name: "no participants", row: db.GetCbtSessionRoomReadinessRow{RoomCount: 1, TotalCapacity: 30}, wantErr: "sesi belum memiliki peserta"},
		{name: "no rooms", row: db.GetCbtSessionRoomReadinessRow{ParticipantCount: 10}, wantErr: "sesi belum memiliki ruangan ujian"},
		{name: "capacity too small", row: db.GetCbtSessionRoomReadinessRow{RoomCount: 1, TotalCapacity: 5, ParticipantCount: 10}, wantErr: "kapasitas ruangan belum cukup"},
		{name: "room over capacity", row: db.GetCbtSessionRoomReadinessRow{RoomCount: 2, TotalCapacity: 60, ParticipantCount: 50, OverCapacityRoomCount: 1}, wantErr: "1 ruangan melebihi kapasitas efektif"},
		{name: "unassigned participants", row: db.GetCbtSessionRoomReadinessRow{RoomCount: 1, TotalCapacity: 30, ParticipantCount: 20, UnassignedParticipantCount: 3}, wantErr: "3 peserta belum mendapat ruangan"},
		{name: "missing seats", row: db.GetCbtSessionRoomReadinessRow{RoomCount: 1, TotalCapacity: 30, ParticipantCount: 20, MissingSeatCount: 4}, wantErr: "4 peserta belum mendapat nomor meja"},
		{name: "rooms without proctors", row: db.GetCbtSessionRoomReadinessRow{RoomCount: 2, TotalCapacity: 60, ParticipantCount: 50, RoomsWithoutProctor: 1}, wantErr: "1 ruangan belum punya pengawas"},
		{name: "network not ready", row: db.GetCbtSessionRoomReadinessRow{RoomCount: 2, TotalCapacity: 60, ParticipantCount: 50, NetworkNotReadyRoomCount: 1}, wantErr: "1 ruangan CBT dengan jaringan belum siap"},
		{name: "power not ready", row: db.GetCbtSessionRoomReadinessRow{RoomCount: 2, TotalCapacity: 60, ParticipantCount: 50, PowerNotReadyRoomCount: 1}, wantErr: "1 ruangan CBT dengan listrik belum siap"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCbtSessionActivationReadiness(tt.row)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateCbtSessionActivationReadiness() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("validateCbtSessionActivationReadiness() error = %v, want ErrConflict", err)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateCbtSessionActivationReadiness() error = %q, want contains %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func decodeParticipantEventData(t *testing.T, arg db.InsertParticipantEventParams) map[string]string {
	t.Helper()
	var got map[string]string
	if err := json.Unmarshal(arg.EventData, &got); err != nil {
		t.Fatalf("event data unmarshal error = %v", err)
	}
	return got
}

func TestCbtSessionParticipantEventListingUsesFakeStore(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(10)
	participantID := cbtSessionTestUUID(11)
	roomID := cbtSessionTestUUID(12)
	store := &fakeCbtSessionStore{
		participantEventRows: []db.ListSessionParticipantEventsRow{{ID: cbtSessionTestUUID(13), EventType: "focus_lost"}},
	}
	svc := &CbtSession{q: store}

	rows, err := svc.ListParticipantEvents(ctx, sessionID, participantID, 0)
	if err != nil {
		t.Fatalf("ListParticipantEvents() error = %v", err)
	}
	if len(rows) != 1 || rows[0].EventType != "focus_lost" {
		t.Fatalf("ListParticipantEvents() rows = %#v", rows)
	}
	if store.participantEventsArg.SessionID != sessionID || store.participantEventsArg.ParticipantID != participantID || store.participantEventsArg.RoomID.Valid || store.participantEventsArg.LimitCount != 100 {
		t.Fatalf("ListParticipantEvents() params = %#v", store.participantEventsArg)
	}

	store.participantEventRows = nil
	rows, err = svc.ListParticipantEventsForRoom(ctx, sessionID, participantID, roomID, 501)
	if err != nil {
		t.Fatalf("ListParticipantEventsForRoom() error = %v", err)
	}
	if rows == nil || len(rows) != 0 {
		t.Fatalf("ListParticipantEventsForRoom() nil rows normalized to %#v, want empty slice", rows)
	}
	if store.participantEventsArg.RoomID != roomID || store.participantEventsArg.LimitCount != 100 {
		t.Fatalf("ListParticipantEventsForRoom() params = %#v", store.participantEventsArg)
	}
}

func TestCbtSessionParticipantEventActionsUseFakeStore(t *testing.T) {
	ctx := context.Background()
	participantID := cbtSessionTestUUID(20)
	store := &fakeCbtSessionStore{
		unlockParticipantRow: db.UnlockParticipantAntiCheatRow{ID: participantID, RiskLevel: "low"},
	}
	svc := &CbtSession{q: store}

	if err := svc.ResetParticipantRuntimeAccess(ctx, participantID, " proctor-a "); err != nil {
		t.Fatalf("ResetParticipantRuntimeAccess() error = %v", err)
	}
	if store.resetParticipantID != participantID || len(store.insertedEvents) != 1 || store.insertedEvents[0].EventType != "proctor_reset_access" {
		t.Fatalf("ResetParticipantRuntimeAccess() calls = reset %v events %#v", store.resetParticipantID, store.insertedEvents)
	}
	if got := decodeParticipantEventData(t, store.insertedEvents[0]); got["actor"] != " proctor-a " {
		t.Fatalf("ResetParticipantRuntimeAccess() event data = %#v", got)
	}

	row, err := svc.UnlockParticipantAntiCheat(ctx, participantID, " proctor-b ", "  verified  ")
	if err != nil {
		t.Fatalf("UnlockParticipantAntiCheat() error = %v", err)
	}
	if row.ID != participantID || store.unlockParticipantID != participantID || len(store.insertedEvents) != 2 || store.insertedEvents[1].EventType != "proctor_unlock" {
		t.Fatalf("UnlockParticipantAntiCheat() row/calls = %#v unlock %v events %#v", row, store.unlockParticipantID, store.insertedEvents)
	}
	if got := decodeParticipantEventData(t, store.insertedEvents[1]); got["actor"] != " proctor-b " || got["notes"] != "verified" {
		t.Fatalf("UnlockParticipantAntiCheat() event data = %#v", got)
	}

	if err := svc.AcknowledgeProctorEvent(ctx, participantID, " event-1 ", "actor-c", " noted "); err != nil {
		t.Fatalf("AcknowledgeProctorEvent() error = %v", err)
	}
	if len(store.insertedEvents) != 3 || store.insertedEvents[2].EventType != "proctor_acknowledge" {
		t.Fatalf("AcknowledgeProctorEvent() events = %#v", store.insertedEvents)
	}
	if got := decodeParticipantEventData(t, store.insertedEvents[2]); got["actor"] != "actor-c" || got["event_id"] != "event-1" || got["notes"] != "noted" {
		t.Fatalf("AcknowledgeProctorEvent() event data = %#v", got)
	}

	if err := svc.RecordIncidentAction(ctx, participantID, " event-2 ", " WARNING_GIVEN ", " actor-d ", " ok "); err != nil {
		t.Fatalf("RecordIncidentAction() error = %v", err)
	}
	if len(store.insertedEvents) != 4 || store.insertedEvents[3].EventType != "proctor_incident_action" {
		t.Fatalf("RecordIncidentAction() events = %#v", store.insertedEvents)
	}
	if got := decodeParticipantEventData(t, store.insertedEvents[3]); got["actor"] != "actor-d" || got["event_id"] != "event-2" || got["action"] != "warning_given" || got["notes"] != "ok" {
		t.Fatalf("RecordIncidentAction() event data = %#v", got)
	}

	if err := svc.RecordIncidentAction(ctx, participantID, "event-3", "bad", "actor", ""); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("RecordIncidentAction() invalid action error = %v, want ErrBadRequest", err)
	}
	if len(store.insertedEvents) != 4 {
		t.Fatalf("RecordIncidentAction() invalid action inserted event: %#v", store.insertedEvents)
	}

	if err := svc.SendParticipantCommand(ctx, participantID, " reconnect ", "  ", " actor-e "); err != nil {
		t.Fatalf("SendParticipantCommand() error = %v", err)
	}
	if len(store.insertedEvents) != 5 || store.insertedEvents[4].EventType != "participant_command" {
		t.Fatalf("SendParticipantCommand() events = %#v", store.insertedEvents)
	}
	if got := decodeParticipantEventData(t, store.insertedEvents[4]); got["actor"] != "actor-e" || got["command_type"] != ParticipantCommandReconnect || got["severity"] != "warning" || got["message"] != defaultParticipantCommandMessage(ParticipantCommandReconnect) || got["issued_at"] == "" {
		t.Fatalf("SendParticipantCommand() event data = %#v", got)
	} else if _, err := time.Parse(time.RFC3339, got["issued_at"]); err != nil {
		t.Fatalf("SendParticipantCommand() issued_at = %q, want RFC3339: %v", got["issued_at"], err)
	}

	if err := svc.SendParticipantCommand(ctx, participantID, "invalid", "msg", "actor"); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("SendParticipantCommand() invalid command error = %v, want ErrBadRequest", err)
	}
	if len(store.insertedEvents) != 5 {
		t.Fatalf("SendParticipantCommand() invalid command inserted event: %#v", store.insertedEvents)
	}
}

func TestCbtSessionForceSubmitParticipantUsesFakeStoreWithoutPool(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(30)
	participantID := cbtSessionTestUUID(31)
	store := &fakeCbtSessionStore{
		forceSubmitRow: db.ForceSubmitParticipantRow{ID: participantID, SubmittedAt: pgtype.Timestamptz{Time: time.Unix(100, 0), Valid: true}},
	}
	svc := &CbtSession{q: store}

	row, err := svc.ForceSubmitParticipant(ctx, sessionID, participantID, " proctor ")
	if err != nil {
		t.Fatalf("ForceSubmitParticipant() error = %v", err)
	}
	if row.ID != participantID {
		t.Fatalf("ForceSubmitParticipant() row = %#v", row)
	}
	if got, want := strings.Join(store.forceSubmitCalls, ","), "update_correctness,force_submit"; got != want {
		t.Fatalf("ForceSubmitParticipant() call order = %q, want %q", got, want)
	}
	if store.participantCorrectnessID != participantID || store.forceSubmitArg.SessionID != sessionID || store.forceSubmitArg.ID != participantID {
		t.Fatalf("ForceSubmitParticipant() params = correctness %v force %#v", store.participantCorrectnessID, store.forceSubmitArg)
	}
	if len(store.insertedEvents) != 1 || store.insertedEvents[0].ParticipantID != participantID || store.insertedEvents[0].EventType != "proctor_force_submit" {
		t.Fatalf("ForceSubmitParticipant() insert events = %#v", store.insertedEvents)
	}
	if got := decodeParticipantEventData(t, store.insertedEvents[0]); got["actor"] != " proctor " {
		t.Fatalf("ForceSubmitParticipant() event data = %#v", got)
	}
}

func TestCbtSessionTeacherAccessWrappersUseFakeStore(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(40)
	participantID := cbtSessionTestUUID(41)
	answerID := cbtSessionTestUUID(42)
	teacherID := cbtSessionTestUUID(43)
	store := &fakeCbtSessionStore{sessionParticipant: true, sessionAnswer: true}
	svc := &CbtSession{q: store}

	ok, err := svc.HasParticipantByTeacher(ctx, sessionID, participantID, teacherID)
	if err != nil || !ok {
		t.Fatalf("HasParticipantByTeacher() = %v, %v; want true, nil", ok, err)
	}
	if store.teacherParticipantArg != (db.HasSessionParticipantByTeacherParams{TeacherEmployeeID: teacherID, SessionID: sessionID, ParticipantID: participantID}) {
		t.Fatalf("HasParticipantByTeacher() params = %#v", store.teacherParticipantArg)
	}

	ok, err = svc.HasAnswerByTeacher(ctx, sessionID, answerID, teacherID)
	if err != nil || !ok {
		t.Fatalf("HasAnswerByTeacher() = %v, %v; want true, nil", ok, err)
	}
	if store.teacherAnswerArg != (db.HasSessionAnswerByTeacherParams{TeacherEmployeeID: teacherID, SessionID: sessionID, AnswerID: answerID}) {
		t.Fatalf("HasAnswerByTeacher() params = %#v", store.teacherAnswerArg)
	}
}

func TestCbtSessionRoomWrappersUseFakeStore(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(50)
	roomID := cbtSessionTestUUID(51)
	proctorID := cbtSessionTestUUID(52)
	store := &fakeCbtSessionStore{
		roomReadinessRow: db.GetCbtSessionRoomReadinessRow{RoomCount: 2, ParticipantCount: 30, TotalCapacity: 60},
		roomProctorRows:  []db.ListCbtRoomProctorsRow{{ID: proctorID, ExamRoomID: roomID, Role: "utama"}},
	}
	svc := &CbtSession{q: store}

	readiness, err := svc.RoomReadiness(ctx, sessionID)
	if err != nil {
		t.Fatalf("RoomReadiness() error = %v", err)
	}
	if readiness.RoomCount != 2 || store.roomReadinessID != sessionID {
		t.Fatalf("RoomReadiness() = %#v arg %v", readiness, store.roomReadinessID)
	}

	proctors, err := svc.ListRoomProctors(ctx, roomID)
	if err != nil {
		t.Fatalf("ListRoomProctors() error = %v", err)
	}
	if len(proctors) != 1 || proctors[0].ID != proctorID || store.roomProctorListID != roomID {
		t.Fatalf("ListRoomProctors() = %#v arg %v", proctors, store.roomProctorListID)
	}

	store.roomProctorRows = nil
	proctors, err = svc.ListRoomProctors(ctx, roomID)
	if err != nil {
		t.Fatalf("ListRoomProctors() nil rows error = %v", err)
	}
	if proctors == nil || len(proctors) != 0 {
		t.Fatalf("ListRoomProctors() nil rows normalized to %#v, want empty slice", proctors)
	}
}

func TestCbtSessionFinalizeOverdueWithFakeStore(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(60)
	store := &fakeCbtSessionStore{finalizeOverdueCount: 3}
	svc := &CbtSession{q: store}

	result, err := svc.FinalizeOverdue(ctx, sessionID)
	if err != nil {
		t.Fatalf("FinalizeOverdue() error = %v", err)
	}
	if result.SessionID != pgUUIDString(sessionID) || result.FinalizedCount != 3 {
		t.Fatalf("FinalizeOverdue() = %+v, want session id and finalized count", result)
	}
	if store.correctnessID != sessionID || store.finalizeOverdueID != sessionID {
		t.Fatalf("FinalizeOverdue() calls correctness=%v finalize=%v, want %v", store.correctnessID, store.finalizeOverdueID, sessionID)
	}

	boom := errors.New("correctness failed")
	store = &fakeCbtSessionStore{correctnessErr: boom, finalizeOverdueCount: 9}
	if count, err := finalizeOverdueWithStore(ctx, store, sessionID); !errors.Is(err, boom) || count != 0 {
		t.Fatalf("finalizeOverdueWithStore(correctness error) = %d/%v, want 0/%v", count, err, boom)
	}
	if store.finalizeOverdueID.Valid {
		t.Fatalf("finalizeOverdueWithStore(correctness error) finalized with id %v, want skipped", store.finalizeOverdueID)
	}

	boom = errors.New("finalize failed")
	store = &fakeCbtSessionStore{finalizeOverdueErr: boom}
	if count, err := finalizeOverdueWithStore(ctx, store, sessionID); !errors.Is(err, boom) || count != 0 || store.correctnessID != sessionID || store.finalizeOverdueID != sessionID {
		t.Fatalf("finalizeOverdueWithStore(finalize error) = %d/%v correctness=%v finalize=%v, want error after correctness", count, err, store.correctnessID, store.finalizeOverdueID)
	}
}

func TestCbtSessionResultFollowUpPathsUseFakeStore(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(70)
	studentID := cbtSessionTestUUID(71)
	participantID := cbtSessionTestUUID(72)
	store := &fakeCbtSessionStore{
		preflightRow: db.GetCbtSessionGradeSyncPreflightRow{
			SessionID:         sessionID,
			SessionTitle:      "PAS Fikih",
			ParticipantCount:  12,
			SubmittedCount:    10,
			ScoredCount:       9,
			MissingScoreCount: 1,
		},
		remedialRows: []db.ListCbtSessionRemedialCandidatesRow{{
			ParticipantID:  participantID,
			StudentID:      studentID,
			Nis:            "001",
			Nama:           "Siswa Remedial",
			ClassCode:      "9A",
			Score:          pgNumeric(64.5),
			QuestionCount:  20,
			IncorrectCount: 7,
		}},
	}
	svc := &CbtSession{q: store}

	preflight, err := svc.GetGradeSyncPreflight(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetGradeSyncPreflight() error = %v", err)
	}
	if preflight.SessionID != sessionID || preflight.SessionTitle != "PAS Fikih" || preflight.MissingScoreCount != 1 || store.preflightID != sessionID {
		t.Fatalf("GetGradeSyncPreflight() = %+v id=%v, want forwarded preflight row", preflight, store.preflightID)
	}

	rows, err := svc.ListRemedialCandidates(ctx, sessionID, 0)
	if err != nil {
		t.Fatalf("ListRemedialCandidates(default threshold) error = %v", err)
	}
	if len(rows) != 1 || rows[0].ParticipantID != participantID || testNumericFloat64(t, store.remedialArg.Threshold) != 75 {
		t.Fatalf("ListRemedialCandidates(default threshold) rows=%+v arg=%+v, want one row threshold 75", rows, store.remedialArg)
	}

	store.remedialRows = nil
	rows, err = svc.ListRemedialCandidates(ctx, sessionID, 101)
	if err != nil {
		t.Fatalf("ListRemedialCandidates(high threshold) error = %v", err)
	}
	if rows == nil || len(rows) != 0 || testNumericFloat64(t, store.remedialArg.Threshold) != 75 {
		t.Fatalf("ListRemedialCandidates(high threshold) rows=%+v arg=%+v, want empty slice threshold 75", rows, store.remedialArg)
	}

	store.remedialRows = []db.ListCbtSessionRemedialCandidatesRow{{ParticipantID: participantID}}
	rows, err = svc.ListRemedialCandidates(ctx, sessionID, 68.25)
	if err != nil {
		t.Fatalf("ListRemedialCandidates(custom threshold) error = %v", err)
	}
	if len(rows) != 1 || testNumericFloat64(t, store.remedialArg.Threshold) != 68.25 || store.remedialArg.SessionID != sessionID {
		t.Fatalf("ListRemedialCandidates(custom threshold) rows=%+v arg=%+v, want custom threshold and session", rows, store.remedialArg)
	}
}

func TestCbtSessionItemAnalysisUsesFakeStore(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(80)
	questionID := cbtSessionTestUUID(81)
	store := &fakeCbtSessionStore{
		itemAnalysisRows: []db.GetSessionItemAnalysisRow{{
			Position:            1,
			QuestionID:          questionID,
			QuestionCode:        "Q-001",
			QuestionType:        "multiple_choice",
			DifficultyIndex:     0.72,
			DiscriminationIndex: 0.41,
			AnswerDistribution:  []byte(`{\"A\":3,\"B\":7}`),
		}},
	}
	svc := &CbtSession{q: store}

	rows, err := svc.GetItemAnalysis(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetItemAnalysis() error = %v", err)
	}
	if len(rows) != 1 || rows[0].QuestionID != questionID || rows[0].QuestionCode != "Q-001" || store.itemAnalysisID != sessionID {
		t.Fatalf("GetItemAnalysis() rows=%+v id=%v, want forwarded item analysis", rows, store.itemAnalysisID)
	}

	store.itemAnalysisRows = nil
	rows, err = svc.GetItemAnalysis(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetItemAnalysis(nil rows) error = %v", err)
	}
	if rows == nil || len(rows) != 0 {
		t.Fatalf("GetItemAnalysis(nil rows) = %+v, want empty slice", rows)
	}
}

func TestCbtSessionListUngradedEssaysByTeacherUsesFakeStore(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(90)
	teacherID := cbtSessionTestUUID(91)
	answerID := cbtSessionTestUUID(92)
	participantID := cbtSessionTestUUID(93)
	questionID := cbtSessionTestUUID(94)
	store := &fakeCbtSessionStore{
		ungradedTeacherRows: []db.ListUngradedEssaysByTeacherRow{{
			AnswerID:      answerID,
			ParticipantID: participantID,
			QuestionID:    questionID,
			Answer:        "uraian",
			QuestionCode:  "ESS-1",
			Nis:           "009",
			Nama:          "Siswa Essay",
			RoomName:      "Lab 1",
		}},
	}
	svc := &CbtSession{q: store}

	rows, err := svc.ListUngradedEssaysByTeacher(ctx, sessionID, teacherID)
	if err != nil {
		t.Fatalf("ListUngradedEssaysByTeacher() error = %v", err)
	}
	if len(rows) != 1 || rows[0].AnswerID != answerID || rows[0].ParticipantID != participantID || rows[0].QuestionCode != "ESS-1" {
		t.Fatalf("ListUngradedEssaysByTeacher() rows = %+v, want converted teacher essay row", rows)
	}
	if store.ungradedTeacherArg.SessionID != sessionID || store.ungradedTeacherArg.TeacherEmployeeID != teacherID {
		t.Fatalf("ListUngradedEssaysByTeacher() arg = %+v, want session and teacher", store.ungradedTeacherArg)
	}

	store.ungradedTeacherRows = nil
	rows, err = svc.ListUngradedEssaysByTeacher(ctx, sessionID, teacherID)
	if err != nil {
		t.Fatalf("ListUngradedEssaysByTeacher(nil rows) error = %v", err)
	}
	if rows == nil || len(rows) != 0 {
		t.Fatalf("ListUngradedEssaysByTeacher(nil rows) = %+v, want empty slice", rows)
	}

	boom := errors.New("teacher essay list failed")
	store.ungradedErr = boom
	if _, err := svc.ListUngradedEssaysByTeacher(ctx, sessionID, teacherID); !errors.Is(err, boom) {
		t.Fatalf("ListUngradedEssaysByTeacher(error) = %v, want %v", err, boom)
	}
}

func testNumericFloat64(t *testing.T, n pgtype.Numeric) float64 {
	t.Helper()
	if !n.Valid {
		t.Fatal("numeric is invalid")
	}
	f, _ := n.Int.Float64()
	scale := new(big.Float).SetInt64(1)
	for i := int32(0); i < -n.Exp; i++ {
		scale.Mul(scale, big.NewFloat(10))
	}
	divisor, _ := scale.Float64()
	return f / divisor
}

func TestCbtSessionUUIDHelpers(t *testing.T) {
	ids := []pgtype.UUID{cbtSessionTestUUID(1), cbtSessionTestUUID(2), pgtype.UUID{}}
	jsonBytes, err := UUIDsToJSON(ids)
	if err != nil {
		t.Fatalf("UUIDsToJSON() error = %v", err)
	}
	var got []string
	if err := json.Unmarshal(jsonBytes, &got); err != nil {
		t.Fatalf("UUIDsToJSON() output unmarshal error = %v", err)
	}
	want := []string{
		"01000000-0000-0000-0000-000000000000",
		"02000000-0000-0000-0000-000000000000",
		"",
	}
	if len(got) != len(want) {
		t.Fatalf("UUIDsToJSON() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("UUIDsToJSON()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	shuffled := shuffleUUIDs(ids)
	if len(shuffled) != len(ids) {
		t.Fatalf("shuffleUUIDs() len = %d, want %d", len(shuffled), len(ids))
	}
	if ids[0] != cbtSessionTestUUID(1) || ids[1] != cbtSessionTestUUID(2) {
		t.Fatalf("shuffleUUIDs() mutated input = %v", ids)
	}
	counts := map[string]int{}
	for _, id := range ids {
		counts[pgUUIDString(id)]++
	}
	for _, id := range shuffled {
		counts[pgUUIDString(id)]--
	}
	for id, count := range counts {
		if count != 0 {
			t.Fatalf("shuffleUUIDs() changed membership for %q: count delta %d", id, count)
		}
	}
}
