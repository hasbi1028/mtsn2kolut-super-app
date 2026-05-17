package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtSessionAutoAssignSeatsNilPoolOrdersSeatsAndMapsErrors(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(201)
	roomID := cbtSessionTestUUID(202)
	firstID := cbtSessionTestUUID(203)
	secondID := cbtSessionTestUUID(204)
	thirdID := cbtSessionTestUUID(205)

	store := &fakeCbtSessionStore{
		byRoomRows: []db.ListParticipantsByRoomRow{
			{ID: thirdID, RoomID: roomID, Nama: "Budi", Nis: "003"},
			{ID: firstID, RoomID: roomID, Nama: "Ani", Nis: "002"},
			{ID: cbtSessionTestUUID(206), RoomID: pgtype.UUID{}, Nama: "No Room", Nis: "000"},
			{ID: secondID, RoomID: roomID, Nama: "Ani", Nis: "001"},
		},
	}
	if err := (&CbtSession{q: store}).AutoAssignSeats(ctx, sessionID); err != nil {
		t.Fatalf("AutoAssignSeats(nil pool) error = %v", err)
	}
	if store.getID != sessionID || store.clearSeatID != sessionID {
		t.Fatalf("AutoAssignSeats(nil pool) ids get=%v clear=%v, want %v", store.getID, store.clearSeatID, sessionID)
	}
	if len(store.assignSeatArgs) != 3 {
		t.Fatalf("AutoAssignSeats(nil pool) assigned %d seats, want 3", len(store.assignSeatArgs))
	}
	wantOrder := []pgtype.UUID{secondID, firstID, thirdID}
	for i, wantID := range wantOrder {
		got := store.assignSeatArgs[i]
		if got.ID != wantID || got.RoomID != roomID || !got.SeatNo.Valid || got.SeatNo.Int32 != int32(i+1) {
			t.Fatalf("AutoAssignSeats(nil pool) assign[%d] = %+v, want id %v seat %d", i, got, wantID, i+1)
		}
	}

	clearErr := &pgconn.PgError{Code: "23505", Message: "duplicate seat"}
	store = &fakeCbtSessionStore{clearSeatErr: clearErr}
	if err := (&CbtSession{q: store}).AutoAssignSeats(ctx, sessionID); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "bertabrakan") {
		t.Fatalf("AutoAssignSeats(clear pg error) = %v, want mapped conflict", err)
	}

	listErr := errors.New("list participants failed")
	store = &fakeCbtSessionStore{byRoomErr: listErr}
	if err := (&CbtSession{q: store}).AutoAssignSeats(ctx, sessionID); !errors.Is(err, listErr) {
		t.Fatalf("AutoAssignSeats(list error) = %v, want %v", err, listErr)
	}

	assignErr := &pgconn.PgError{Code: "23514", Message: "bad seat"}
	store = &fakeCbtSessionStore{
		byRoomRows:    []db.ListParticipantsByRoomRow{{ID: firstID, RoomID: roomID, Nama: "Ani", Nis: "001"}},
		assignSeatErr: assignErr,
	}
	if err := (&CbtSession{q: store}).AutoAssignSeats(ctx, sessionID); !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "tidak valid") {
		t.Fatalf("AutoAssignSeats(assign pg error) = %v, want mapped bad request", err)
	}
}

func TestCbtSessionShuffleRoomsCoreStopsOnStoreErrors(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(211)
	participantID := cbtSessionTestUUID(212)
	roomID := cbtSessionTestUUID(213)

	clearErr := errors.New("clear rooms failed")
	store := &fakeCbtSessionStore{clearRoomErr: clearErr}
	if err := shuffleRooms(ctx, store, sessionID); !errors.Is(err, clearErr) {
		t.Fatalf("shuffleRooms(clear error) = %v, want %v", err, clearErr)
	}
	if len(store.assignRoomArgs) != 0 {
		t.Fatalf("shuffleRooms(clear error) assigned rooms = %+v, want none", store.assignRoomArgs)
	}

	listErr := errors.New("list by room failed")
	store = &fakeCbtSessionStore{byRoomErr: listErr}
	if err := shuffleRooms(ctx, store, sessionID); !errors.Is(err, listErr) {
		t.Fatalf("shuffleRooms(list participants error) = %v, want %v", err, listErr)
	}

	roomsErr := errors.New("list rooms failed")
	store = &fakeCbtSessionStore{byRoomRows: []db.ListParticipantsByRoomRow{{ID: participantID}}, roomsErr: roomsErr}
	if err := shuffleRooms(ctx, store, sessionID); !errors.Is(err, roomsErr) {
		t.Fatalf("shuffleRooms(list rooms error) = %v, want %v", err, roomsErr)
	}

	assignErr := errors.New("assign room failed")
	store = &fakeCbtSessionStore{
		byRoomRows:    []db.ListParticipantsByRoomRow{{ID: participantID}},
		roomRows:      []db.ListCbtExamRoomsRow{{ID: roomID, Capacity: 1}},
		assignRoomErr: assignErr,
	}
	if err := shuffleRooms(ctx, store, sessionID); !errors.Is(err, assignErr) {
		t.Fatalf("shuffleRooms(assign error) = %v, want %v", err, assignErr)
	}
	if len(store.assignRoomArgs) != 1 || store.assignRoomArgs[0].ID != participantID || store.assignRoomArgs[0].RoomID != roomID {
		t.Fatalf("shuffleRooms(assign error) args = %+v, want attempted participant/room assignment", store.assignRoomArgs)
	}

	store = &fakeCbtSessionStore{byRoomRows: []db.ListParticipantsByRoomRow{{ID: participantID}}}
	if err := shuffleRooms(ctx, store, sessionID); err != nil {
		t.Fatalf("shuffleRooms(no rooms) error = %v, want nil", err)
	}
	if len(store.assignRoomArgs) != 0 {
		t.Fatalf("shuffleRooms(no rooms) assigned = %+v, want none", store.assignRoomArgs)
	}
}

func TestCbtSessionReplaceRoomProctorsPreTransactionErrors(t *testing.T) {
	ctx := context.Background()
	roomID := cbtSessionTestUUID(221)
	assignedBy := cbtSessionTestUUID(222)
	primaryID := cbtSessionTestUUID(223)
	sessionID := cbtSessionTestUUID(224)

	activeStore := &fakeCbtSessionStore{roomSetupRow: db.GetCbtExamRoomSetupContextRow{ID: roomID, SessionID: sessionID, SessionStatus: db.CbtSessionStatusEnumActive}}
	if _, err := (&CbtSession{q: activeStore}).ReplaceRoomProctors(ctx, roomID, assignedBy, primaryID, nil); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("ReplaceRoomProctors(active room) = %v, want ErrConflict", err)
	}
	if activeStore.proctorOverlapArg.EmployeeID.Valid {
		t.Fatalf("ReplaceRoomProctors(active room) checked overlaps = %+v, want blocked before overlap", activeStore.proctorOverlapArg)
	}

	overlapStore := &fakeCbtSessionStore{
		roomSetupRow:   db.GetCbtExamRoomSetupContextRow{ID: roomID, SessionID: sessionID, SessionStatus: db.CbtSessionStatusEnumScheduled},
		proctorOverlap: true,
	}
	if _, err := (&CbtSession{q: overlapStore}).ReplaceRoomProctors(ctx, roomID, assignedBy, primaryID, []pgtype.UUID{primaryID}); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "bertugas") {
		t.Fatalf("ReplaceRoomProctors(overlap) = %v, want overlap conflict", err)
	}
	if overlapStore.proctorOverlapArg.SessionID != sessionID || overlapStore.proctorOverlapArg.EmployeeID != primaryID || overlapStore.proctorOverlapArg.ExamRoomID != roomID {
		t.Fatalf("ReplaceRoomProctors(overlap) arg = %+v, want session/employee/room", overlapStore.proctorOverlapArg)
	}
}

func TestCbtSessionForceSubmitParticipantNilPoolSequenceAndErrors(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(231)
	participantID := cbtSessionTestUUID(232)
	row := db.ForceSubmitParticipantRow{ID: participantID}

	store := &fakeCbtSessionStore{forceSubmitRow: row}
	got, err := (&CbtSession{q: store}).ForceSubmitParticipant(ctx, sessionID, participantID, "proctor-a")
	if err != nil {
		t.Fatalf("ForceSubmitParticipant(nil pool) error = %v", err)
	}
	if got.ID != participantID || store.participantCorrectnessID != participantID || store.forceSubmitArg.SessionID != sessionID || store.forceSubmitArg.ID != participantID {
		t.Fatalf("ForceSubmitParticipant(nil pool) row=%+v correctness=%v arg=%+v, want participant/session", got, store.participantCorrectnessID, store.forceSubmitArg)
	}
	wantCalls := []string{"update_correctness", "force_submit"}
	if len(store.forceSubmitCalls) != len(wantCalls) || store.forceSubmitCalls[0] != wantCalls[0] || store.forceSubmitCalls[1] != wantCalls[1] {
		t.Fatalf("ForceSubmitParticipant(nil pool) calls = %+v, want %+v", store.forceSubmitCalls, wantCalls)
	}
	if len(store.insertedEvents) != 1 || store.insertedEvents[0].ParticipantID != participantID || store.insertedEvents[0].EventType != "proctor_force_submit" || !strings.Contains(string(store.insertedEvents[0].EventData), "proctor-a") {
		t.Fatalf("ForceSubmitParticipant(nil pool) event = %+v, want force-submit event with actor", store.insertedEvents)
	}

	correctnessErr := errors.New("participant correctness failed")
	store = &fakeCbtSessionStore{participantCorrectnessErr: correctnessErr}
	if _, err := (&CbtSession{q: store}).ForceSubmitParticipant(ctx, sessionID, participantID, "proctor-a"); !errors.Is(err, correctnessErr) {
		t.Fatalf("ForceSubmitParticipant(correctness error) = %v, want %v", err, correctnessErr)
	}
	if store.forceSubmitArg.ID.Valid || len(store.insertedEvents) != 0 {
		t.Fatalf("ForceSubmitParticipant(correctness error) forceArg=%+v events=%+v, want stopped", store.forceSubmitArg, store.insertedEvents)
	}

	forceErr := errors.New("force submit failed")
	store = &fakeCbtSessionStore{forceSubmitErr: forceErr}
	if _, err := (&CbtSession{q: store}).ForceSubmitParticipant(ctx, sessionID, participantID, "proctor-a"); !errors.Is(err, forceErr) {
		t.Fatalf("ForceSubmitParticipant(force error) = %v, want %v", err, forceErr)
	}
	if len(store.insertedEvents) != 0 {
		t.Fatalf("ForceSubmitParticipant(force error) events=%+v, want none", store.insertedEvents)
	}

	eventErr := errors.New("event insert failed")
	store = &fakeCbtSessionStore{forceSubmitRow: row, insertEventErr: eventErr}
	if _, err := (&CbtSession{q: store}).ForceSubmitParticipant(ctx, sessionID, participantID, "proctor-a"); !errors.Is(err, eventErr) {
		t.Fatalf("ForceSubmitParticipant(event error) = %v, want %v", err, eventErr)
	}
	if len(store.insertedEvents) != 1 {
		t.Fatalf("ForceSubmitParticipant(event error) inserted %d events, want attempted one", len(store.insertedEvents))
	}
}

func TestCbtSessionFinalizeOverdueNilPoolResultAndErrors(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(241)

	store := &fakeCbtSessionStore{finalizeOverdueCount: 7}
	got, err := (&CbtSession{q: store}).FinalizeOverdue(ctx, sessionID)
	if err != nil {
		t.Fatalf("FinalizeOverdue(nil pool) error = %v", err)
	}
	if got.SessionID != pgUUIDString(sessionID) || got.FinalizedCount != 7 || store.correctnessID != sessionID || store.finalizeOverdueID != sessionID {
		t.Fatalf("FinalizeOverdue(nil pool) = %+v correctness=%v finalize=%v, want count/session and both steps", got, store.correctnessID, store.finalizeOverdueID)
	}

	correctnessErr := errors.New("correctness failed")
	store = &fakeCbtSessionStore{correctnessErr: correctnessErr}
	if _, err := (&CbtSession{q: store}).FinalizeOverdue(ctx, sessionID); !errors.Is(err, correctnessErr) {
		t.Fatalf("FinalizeOverdue(correctness error) = %v, want %v", err, correctnessErr)
	}
	if store.finalizeOverdueID.Valid {
		t.Fatalf("FinalizeOverdue(correctness error) finalize id = %v, want skipped", store.finalizeOverdueID)
	}

	finalizeErr := errors.New("finalize failed")
	store = &fakeCbtSessionStore{finalizeOverdueErr: finalizeErr}
	if _, err := (&CbtSession{q: store}).FinalizeOverdue(ctx, sessionID); !errors.Is(err, finalizeErr) {
		t.Fatalf("FinalizeOverdue(finalize error) = %v, want %v", err, finalizeErr)
	}
	if store.correctnessID != sessionID || store.finalizeOverdueID != sessionID {
		t.Fatalf("FinalizeOverdue(finalize error) ids correctness=%v finalize=%v, want both attempted", store.correctnessID, store.finalizeOverdueID)
	}
}
