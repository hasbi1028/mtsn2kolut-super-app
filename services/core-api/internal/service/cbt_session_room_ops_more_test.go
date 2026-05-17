package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtSessionRoomOpsMoreAutoAssignSeatsUsesNilPoolAndSeparateRoomOrdering(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(1)
	roomA := cbtSessionTestUUID(2)
	roomB := cbtSessionTestUUID(3)
	ann02 := cbtSessionTestUUID(4)
	ann01 := cbtSessionTestUUID(5)
	bob := cbtSessionTestUUID(6)
	cici := cbtSessionTestUUID(7)
	unassigned := cbtSessionTestUUID(8)

	store := &fakeCbtSessionStore{
		sessionRow: db.GetCbtExamSessionRow{ID: sessionID, Status: db.CbtSessionStatusEnumDraft},
		byRoomRows: []db.ListParticipantsByRoomRow{
			{ID: bob, RoomID: roomA, Nama: "Bob", Nis: "003"},
			{ID: unassigned, Nama: "Tanpa Ruang", Nis: "999"},
			{ID: ann02, RoomID: roomA, Nama: "Ann", Nis: "002"},
			{ID: cici, RoomID: roomB, Nama: "Cici", Nis: "004"},
			{ID: ann01, RoomID: roomA, Nama: "Ann", Nis: "001"},
		},
	}

	if err := (&CbtSession{q: store}).AutoAssignSeats(ctx, sessionID); err != nil {
		t.Fatalf("AutoAssignSeats(nil pool) error = %v", err)
	}
	if store.getID != sessionID || store.clearSeatID != sessionID {
		t.Fatalf("AutoAssignSeats(nil pool) ids get=%v clear=%v, want %v", store.getID, store.clearSeatID, sessionID)
	}
	if len(store.assignSeatArgs) != 4 {
		t.Fatalf("AutoAssignSeats(nil pool) assignments = %d, want 4", len(store.assignSeatArgs))
	}

	seats := map[pgtype.UUID]db.AssignParticipantSeatParams{}
	for _, arg := range store.assignSeatArgs {
		if !arg.RoomID.Valid || !arg.SeatNo.Valid {
			t.Fatalf("AutoAssignSeats(nil pool) invalid assignment = %+v", arg)
		}
		seats[arg.ID] = arg
	}
	wantSeats := map[pgtype.UUID]int32{ann01: 1, ann02: 2, bob: 3, cici: 1}
	for id, wantSeat := range wantSeats {
		got, ok := seats[id]
		if !ok || got.SeatNo.Int32 != wantSeat {
			t.Fatalf("AutoAssignSeats(nil pool) participant %v seat=%+v present=%v, want %d", id, got, ok, wantSeat)
		}
	}
	if seats[ann01].RoomID != roomA || seats[cici].RoomID != roomB {
		t.Fatalf("AutoAssignSeats(nil pool) room preservation failed: ann=%+v cici=%+v", seats[ann01], seats[cici])
	}
	if _, ok := seats[unassigned]; ok {
		t.Fatalf("AutoAssignSeats(nil pool) assigned participant without room: %+v", seats[unassigned])
	}
}

func TestCbtSessionRoomOpsMoreShuffleRoomsCapacityOverrideAndMutableGuard(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(11)
	roomA := cbtSessionTestUUID(12)
	roomB := cbtSessionTestUUID(13)
	participants := []db.ListParticipantsByRoomRow{
		{ID: cbtSessionTestUUID(14)},
		{ID: cbtSessionTestUUID(15)},
		{ID: cbtSessionTestUUID(16)},
		{ID: cbtSessionTestUUID(17)},
	}

	store := &fakeCbtSessionStore{
		byRoomRows: participants,
		roomRows: []db.ListCbtExamRoomsRow{
			{ID: roomA, Capacity: 10, CapacityOverride: pgtype.Int4{Int32: 1, Valid: true}},
			{ID: roomB, Capacity: 2},
		},
	}
	if err := shuffleRooms(ctx, store, sessionID); err != nil {
		t.Fatalf("shuffleRooms(capacity override) error = %v", err)
	}
	if store.clearRoomID != sessionID {
		t.Fatalf("shuffleRooms(capacity override) clear id = %v, want %v", store.clearRoomID, sessionID)
	}
	if len(store.assignRoomArgs) != 3 {
		t.Fatalf("shuffleRooms(capacity override) assignments = %d, want 3 capped by effective capacity", len(store.assignRoomArgs))
	}
	counts := map[pgtype.UUID]int{}
	seen := map[pgtype.UUID]bool{}
	for _, arg := range store.assignRoomArgs {
		counts[arg.RoomID]++
		if seen[arg.ID] {
			t.Fatalf("shuffleRooms(capacity override) assigned participant twice: %v", arg.ID)
		}
		seen[arg.ID] = true
	}
	if counts[roomA] != 1 || counts[roomB] != 2 {
		t.Fatalf("shuffleRooms(capacity override) room counts = %+v, want roomA=1 roomB=2", counts)
	}

	activeStore := &fakeCbtSessionStore{sessionRow: db.GetCbtExamSessionRow{ID: sessionID, Status: db.CbtSessionStatusEnumActive}}
	if err := (&CbtSession{q: activeStore}).ShuffleRooms(ctx, sessionID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("ShuffleRooms(active session) error = %v, want ErrConflict", err)
	}
	if activeStore.clearRoomID.Valid {
		t.Fatalf("ShuffleRooms(active session) clear id = %v, want no shuffle work", activeStore.clearRoomID)
	}
}

func TestCbtSessionRoomOpsMoreReplaceRoomProctorsNormalizesBeforeNilPoolTransaction(t *testing.T) {
	ctx := context.Background()
	roomID := cbtSessionTestUUID(21)
	sessionID := cbtSessionTestUUID(22)
	assignedBy := cbtSessionTestUUID(23)
	primary := cbtSessionTestUUID(24)
	backup := cbtSessionTestUUID(25)
	third := cbtSessionTestUUID(26)
	store := &fakeCbtSessionStore{
		roomSetupRow: db.GetCbtExamRoomSetupContextRow{ID: roomID, SessionID: sessionID, SessionStatus: db.CbtSessionStatusEnumScheduled},
	}

	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("ReplaceRoomProctors(nil pool) did not panic after validation; update test if nil-pool transaction path is added")
		}
		if store.roomSetupID != roomID {
			t.Fatalf("ReplaceRoomProctors(nil pool) setup id = %v, want %v", store.roomSetupID, roomID)
		}
		if store.proctorOverlapArg.SessionID != sessionID || store.proctorOverlapArg.EmployeeID != third || store.proctorOverlapArg.ExamRoomID != roomID {
			t.Fatalf("ReplaceRoomProctors(nil pool) last overlap arg = %+v, want final unique proctor before pool acquire", store.proctorOverlapArg)
		}
	}()
	_, _ = (&CbtSession{q: store}).ReplaceRoomProctors(ctx, roomID, assignedBy, primary, []pgtype.UUID{backup, primary, pgtype.UUID{}, backup, third})
}

func TestCbtSessionRoomOpsMoreFinalizeOverdueNilPoolZeroAndFinalizeError(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(31)

	store := &fakeCbtSessionStore{}
	got, err := (&CbtSession{q: store}).FinalizeOverdue(ctx, sessionID)
	if err != nil {
		t.Fatalf("FinalizeOverdue(nil pool zero) error = %v", err)
	}
	if got.SessionID != pgUUIDString(sessionID) || got.FinalizedCount != 0 {
		t.Fatalf("FinalizeOverdue(nil pool zero) = %+v, want session and zero count", got)
	}
	if store.correctnessID != sessionID || store.finalizeOverdueID != sessionID {
		t.Fatalf("FinalizeOverdue(nil pool zero) calls correctness=%v finalize=%v, want %v", store.correctnessID, store.finalizeOverdueID, sessionID)
	}

	boom := errors.New("overdue finalize failed")
	store = &fakeCbtSessionStore{finalizeOverdueErr: boom}
	if got, err := (&CbtSession{q: store}).FinalizeOverdue(ctx, sessionID); !errors.Is(err, boom) || got.FinalizedCount != 0 || got.SessionID != "" {
		t.Fatalf("FinalizeOverdue(nil pool finalize error) = %+v/%v, want zero result/%v", got, err, boom)
	}
	if store.correctnessID != sessionID || store.finalizeOverdueID != sessionID {
		t.Fatalf("FinalizeOverdue(nil pool finalize error) calls correctness=%v finalize=%v, want both", store.correctnessID, store.finalizeOverdueID)
	}
}
