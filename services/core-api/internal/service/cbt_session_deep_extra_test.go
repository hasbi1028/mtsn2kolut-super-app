package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtSessionDeepExtraShuffleRoomsEarlyReturnsAndErrors(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(181)
	participantID := cbtSessionTestUUID(182)
	roomID := cbtSessionTestUUID(183)
	boom := errors.New("shuffle failed")

	store := &fakeCbtSessionStore{clearRoomErr: boom}
	if err := shuffleRooms(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("shuffleRooms(clear error) = %v, want %v", err, boom)
	}
	if len(store.assignRoomArgs) != 0 {
		t.Fatalf("shuffleRooms(clear error) assigned rooms = %+v, want none", store.assignRoomArgs)
	}

	store = &fakeCbtSessionStore{byRoomErr: boom}
	if err := shuffleRooms(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("shuffleRooms(participant error) = %v, want %v", err, boom)
	}
	if store.clearRoomID != sessionID || len(store.assignRoomArgs) != 0 {
		t.Fatalf("shuffleRooms(participant error) clear=%v assign=%+v, want clear only", store.clearRoomID, store.assignRoomArgs)
	}

	store = &fakeCbtSessionStore{byRoomRows: []db.ListParticipantsByRoomRow{{ID: participantID}}, roomsErr: boom}
	if err := shuffleRooms(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("shuffleRooms(room error) = %v, want %v", err, boom)
	}

	for _, tt := range []struct {
		name  string
		store *fakeCbtSessionStore
	}{
		{name: "no rooms", store: &fakeCbtSessionStore{byRoomRows: []db.ListParticipantsByRoomRow{{ID: participantID}}}},
		{name: "no participants", store: &fakeCbtSessionStore{roomRows: []db.ListCbtExamRoomsRow{{ID: roomID, Capacity: 1}}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := shuffleRooms(ctx, tt.store, sessionID); err != nil {
				t.Fatalf("shuffleRooms(%s) error = %v, want nil", tt.name, err)
			}
			if len(tt.store.assignRoomArgs) != 0 {
				t.Fatalf("shuffleRooms(%s) assignments = %+v, want none", tt.name, tt.store.assignRoomArgs)
			}
		})
	}

	store = &fakeCbtSessionStore{
		byRoomRows:    []db.ListParticipantsByRoomRow{{ID: participantID}},
		roomRows:      []db.ListCbtExamRoomsRow{{ID: roomID, Capacity: 1}},
		assignRoomErr: boom,
	}
	if err := shuffleRooms(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("shuffleRooms(assign error) = %v, want %v", err, boom)
	}
	if len(store.assignRoomArgs) != 1 || store.assignRoomArgs[0].ID != participantID || store.assignRoomArgs[0].RoomID != roomID {
		t.Fatalf("shuffleRooms(assign error) args = %+v, want attempted participant assignment", store.assignRoomArgs)
	}
}

func TestCbtSessionDeepExtraScoreSessionHelperOrdersAndStopsOnErrors(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(184)

	store := &fakeCbtSessionStore{}
	if err := scoreSession(ctx, store, sessionID); err != nil {
		t.Fatalf("scoreSession() error = %v", err)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("scoreSession() ids correctness=%v scores=%v, want %v", store.correctnessID, store.scoresID, sessionID)
	}

	boom := errors.New("answer correctness failed")
	store = &fakeCbtSessionStore{correctnessErr: boom}
	if err := scoreSession(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("scoreSession(correctness error) = %v, want %v", err, boom)
	}
	if store.scoresID.Valid {
		t.Fatalf("scoreSession(correctness error) scoresID=%v, want skipped", store.scoresID)
	}

	boom = errors.New("participant score failed")
	store = &fakeCbtSessionStore{scoresErr: boom}
	if err := scoreSession(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("scoreSession(score error) = %v, want %v", err, boom)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("scoreSession(score error) ids correctness=%v scores=%v, want both called", store.correctnessID, store.scoresID)
	}
}

func TestCbtSessionDeepExtraGradeEssayNilPoolAndHelperEdgeErrors(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(185)
	answerID := cbtSessionTestUUID(186)

	store := &fakeCbtSessionStore{}
	if err := gradeEssayAndRefreshScore(ctx, store, sessionID, answerID, 0, ""); err != nil {
		t.Fatalf("gradeEssayAndRefreshScore(zero score) error = %v", err)
	}
	if store.gradeArg.ID != answerID || store.gradeArg.GradedBy.String != "" || !store.gradeArg.GradedBy.Valid || testNumericFloat64(t, store.gradeArg.ManualScore) != 0 {
		t.Fatalf("gradeEssayAndRefreshScore(zero score) grade arg = %+v, want zero score and valid empty grader", store.gradeArg)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("gradeEssayAndRefreshScore(zero score) score refresh = %v/%v, want %v", store.correctnessID, store.scoresID, sessionID)
	}

	boom := errors.New("nil-pool score refresh failed")
	store = &fakeCbtSessionStore{scoresErr: boom}
	if err := (&CbtSession{q: store}).GradeEssay(ctx, sessionID, answerID, 12.5, "guru"); !errors.Is(err, boom) {
		t.Fatalf("GradeEssay(nil pool score error) = %v, want %v", err, boom)
	}
	if store.gradeArg.ID != answerID || store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("GradeEssay(nil pool score error) calls grade=%+v correctness=%v scores=%v", store.gradeArg, store.correctnessID, store.scoresID)
	}
}

func TestCbtSessionDeepExtraForceSubmitNilPoolErrorBranches(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(187)
	participantID := cbtSessionTestUUID(188)

	boom := errors.New("participant correctness failed")
	store := &fakeCbtSessionStore{participantCorrectnessErr: boom}
	if row, err := forceSubmitParticipant(ctx, store, sessionID, participantID, "actor"); !errors.Is(err, boom) || row.ID.Valid {
		t.Fatalf("forceSubmitParticipant(correctness error) = %+v/%v, want zero/%v", row, err, boom)
	}
	if strings.Join(store.forceSubmitCalls, ",") != "update_correctness" || len(store.insertedEvents) != 0 {
		t.Fatalf("forceSubmitParticipant(correctness error) calls=%q events=%+v, want correctness only", strings.Join(store.forceSubmitCalls, ","), store.insertedEvents)
	}

	boom = errors.New("force submit failed")
	store = &fakeCbtSessionStore{forceSubmitErr: boom}
	if _, err := forceSubmitParticipant(ctx, store, sessionID, participantID, "actor"); !errors.Is(err, boom) {
		t.Fatalf("forceSubmitParticipant(force error) = %v, want %v", err, boom)
	}
	if strings.Join(store.forceSubmitCalls, ",") != "update_correctness,force_submit" || len(store.insertedEvents) != 0 {
		t.Fatalf("forceSubmitParticipant(force error) calls=%q events=%+v, want no event", strings.Join(store.forceSubmitCalls, ","), store.insertedEvents)
	}

	boom = errors.New("event insert failed")
	store = &fakeCbtSessionStore{forceSubmitRow: db.ForceSubmitParticipantRow{ID: participantID}, insertEventErr: boom}
	if _, err := forceSubmitParticipant(ctx, store, sessionID, participantID, " actor "); !errors.Is(err, boom) {
		t.Fatalf("forceSubmitParticipant(event error) = %v, want %v", err, boom)
	}
	if len(store.insertedEvents) != 1 || store.insertedEvents[0].ParticipantID != participantID || store.insertedEvents[0].EventType != "proctor_force_submit" {
		t.Fatalf("forceSubmitParticipant(event error) inserted events = %+v, want attempted force-submit event", store.insertedEvents)
	}
}

func TestCbtSessionDeepExtraAutoAssignSeatsErrorBranchesAndEmptyGroups(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(189)
	participantID := cbtSessionTestUUID(190)
	roomID := cbtSessionTestUUID(191)
	boom := errors.New("seat assignment failed")

	store := &fakeCbtSessionStore{clearSeatErr: boom}
	if err := autoAssignSeats(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("autoAssignSeats(clear error) = %v, want %v", err, boom)
	}
	if store.clearSeatID != sessionID || len(store.assignSeatArgs) != 0 {
		t.Fatalf("autoAssignSeats(clear error) clear=%v assign=%+v, want clear only", store.clearSeatID, store.assignSeatArgs)
	}

	store = &fakeCbtSessionStore{byRoomErr: boom}
	if err := autoAssignSeats(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("autoAssignSeats(list error) = %v, want %v", err, boom)
	}
	if len(store.assignSeatArgs) != 0 {
		t.Fatalf("autoAssignSeats(list error) assignments=%+v, want none", store.assignSeatArgs)
	}

	store = &fakeCbtSessionStore{byRoomRows: []db.ListParticipantsByRoomRow{{ID: participantID, RoomID: roomID}}, assignSeatErr: boom}
	if err := autoAssignSeats(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("autoAssignSeats(assign error) = %v, want %v", err, boom)
	}
	if len(store.assignSeatArgs) != 1 || store.assignSeatArgs[0].ID != participantID || store.assignSeatArgs[0].SeatNo.Int32 != 1 {
		t.Fatalf("autoAssignSeats(assign error) args=%+v, want first seat attempted", store.assignSeatArgs)
	}

	store = &fakeCbtSessionStore{byRoomRows: []db.ListParticipantsByRoomRow{{ID: participantID}}}
	if err := autoAssignSeats(ctx, store, sessionID); err != nil {
		t.Fatalf("autoAssignSeats(only unassigned participants) error = %v, want nil", err)
	}
	if len(store.assignSeatArgs) != 0 {
		t.Fatalf("autoAssignSeats(only unassigned participants) args=%+v, want no assignments", store.assignSeatArgs)
	}
}

func TestCbtSessionDeepExtraReplaceRoomProctorsNilPoolAndOrderingBranches(t *testing.T) {
	ctx := context.Background()
	roomID := cbtSessionTestUUID(192)
	sessionID := cbtSessionTestUUID(193)
	primary := cbtSessionTestUUID(194)
	backup := cbtSessionTestUUID(195)
	assignedBy := cbtSessionTestUUID(196)

	ordered := normalizeRoomProctorIDs(pgtype.UUID{}, []pgtype.UUID{backup, pgtype.UUID{}, backup, primary})
	if len(ordered) != 2 || ordered[0] != backup || ordered[1] != primary {
		t.Fatalf("normalizeRoomProctorIDs(no primary) = %+v, want unique valid input order", ordered)
	}

	store := &fakeCbtSessionStore{
		roomSetupRow: db.GetCbtExamRoomSetupContextRow{ID: roomID, SessionID: sessionID, SessionStatus: db.CbtSessionStatusEnumScheduled},
	}
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("ReplaceRoomProctors(nil pool after validations) did not panic; update test if fake-store transaction path is added")
		}
		if store.proctorOverlapArg.SessionID != sessionID || store.proctorOverlapArg.EmployeeID != backup || store.proctorOverlapArg.ExamRoomID != roomID {
			t.Fatalf("ReplaceRoomProctors(nil pool) last overlap arg = %+v, want checked unique backup before pool acquire", store.proctorOverlapArg)
		}
	}()
	_, _ = (&CbtSession{q: store}).ReplaceRoomProctors(ctx, roomID, assignedBy, primary, []pgtype.UUID{backup, primary, backup, pgtype.UUID{}})
}

func TestCbtSessionDeepExtraPublicScoreAndShuffleNilPoolPanicsAfterValidation(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(197)
	for _, tt := range []struct {
		name string
		run  func()
	}{
		{name: "ScoreSession", run: func() { _ = (&CbtSession{q: &fakeCbtSessionStore{}}).ScoreSession(ctx, sessionID) }},
		{name: "ShuffleRooms", run: func() {
			_ = (&CbtSession{q: &fakeCbtSessionStore{sessionRow: db.GetCbtExamSessionRow{ID: sessionID, Status: db.CbtSessionStatusEnumScheduled}}}).ShuffleRooms(ctx, sessionID)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered == nil {
					t.Fatalf("%s(nil pool) did not panic; update test if nil-pool transaction path is added", tt.name)
				}
			}()
			tt.run()
		})
	}
}

func TestCbtSessionDeepExtraAutoAssignSeatsPublicNilPoolStillValidatesMutability(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(198)
	store := &fakeCbtSessionStore{sessionRow: db.GetCbtExamSessionRow{ID: sessionID, Status: db.CbtSessionStatusEnumFinished}}
	if err := (&CbtSession{q: store}).AutoAssignSeats(ctx, sessionID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("AutoAssignSeats(finished session) = %v, want ErrConflict", err)
	}
	if store.clearSeatID.Valid || len(store.assignSeatArgs) != 0 {
		t.Fatalf("AutoAssignSeats(finished session) touched seat store clear=%v args=%+v, want blocked", store.clearSeatID, store.assignSeatArgs)
	}
}
