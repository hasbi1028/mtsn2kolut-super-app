package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeAssessmentExamStore struct {
	listRows        []db.ListAssessmentExamsRow
	getRow          db.GetAssessmentExamRow
	getErr          error
	created         db.AssessmentExam
	updated         db.AssessmentExam
	firstSession    db.AssessmentSession
	firstSessionErr error
	createdSession  db.AssessmentSession
	createdRoom     db.AssessmentRoom
	createdRooms    []db.CreateAssessmentRoomParams
	roomCount       int64
	participantCnt  int64
	cardCount       int64
	createArg       db.CreateAssessmentExamParams
	updateArg       db.UpdateAssessmentExamParams
}

func (f *fakeAssessmentExamStore) ListAssessmentExams(context.Context, db.ListAssessmentExamsParams) ([]db.ListAssessmentExamsRow, error) {
	return f.listRows, nil
}

func (f *fakeAssessmentExamStore) GetAssessmentExam(context.Context, pgtype.UUID) (db.GetAssessmentExamRow, error) {
	return f.getRow, f.getErr
}

func (f *fakeAssessmentExamStore) CreateAssessmentExam(_ context.Context, arg db.CreateAssessmentExamParams) (db.AssessmentExam, error) {
	f.createArg = arg
	return f.created, nil
}

func (f *fakeAssessmentExamStore) UpdateAssessmentExam(_ context.Context, arg db.UpdateAssessmentExamParams) (db.AssessmentExam, error) {
	f.updateArg = arg
	return f.updated, nil
}

func (f *fakeAssessmentExamStore) GetFirstAssessmentSessionByExam(context.Context, pgtype.UUID) (db.AssessmentSession, error) {
	return f.firstSession, f.firstSessionErr
}

func (f *fakeAssessmentExamStore) CreateAssessmentSession(context.Context, db.CreateAssessmentSessionParams) (db.AssessmentSession, error) {
	return f.createdSession, nil
}

func (f *fakeAssessmentExamStore) CreateAssessmentRoom(context.Context, db.CreateAssessmentRoomParams) (db.AssessmentRoom, error) {
	return f.createdRoom, nil
}

func (f *fakeAssessmentExamStore) UpsertAssessmentRoom(_ context.Context, arg db.UpsertAssessmentRoomParams) (db.AssessmentRoom, error) {
	f.createdRooms = append(f.createdRooms, db.CreateAssessmentRoomParams{
		SessionID: arg.SessionID,
		Code:      arg.Code,
		Name:      arg.Name,
		Capacity:  arg.Capacity,
	})
	return db.AssessmentRoom{ID: mustPgUUIDAssessmentTest("33333333-3333-3333-3333-333333333333"), SessionID: arg.SessionID, Code: arg.Code, Name: arg.Name, Capacity: arg.Capacity}, nil
}

func (f *fakeAssessmentExamStore) CountAssessmentRoomsByExam(context.Context, pgtype.UUID) (int64, error) {
	return f.roomCount, nil
}

func (f *fakeAssessmentExamStore) CountAssessmentParticipantsByExam(context.Context, pgtype.UUID) (int64, error) {
	return f.participantCnt, nil
}

func (f *fakeAssessmentExamStore) CountAssessmentCardsByExam(context.Context, pgtype.UUID) (int64, error) {
	return f.cardCount, nil
}

func TestAssessmentExamCreateValidation(t *testing.T) {
	svc := NewAssessmentExamWithStore(&fakeAssessmentExamStore{})
	_, err := svc.Create(context.Background(), AssessmentExamInput{Title: "  "})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Create empty title err = %v, want bad request", err)
	}
	_, err = svc.Create(context.Background(), AssessmentExamInput{
		Title:      "Ujian",
		GradeLevel: pgtype.Int2{Int16: 10, Valid: true},
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Create invalid grade err = %v, want bad request", err)
	}
}

func TestAssessmentExamCreateNormalizesDraft(t *testing.T) {
	id := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		created: db.AssessmentExam{ID: id},
		getRow: db.GetAssessmentExamRow{
			ID:        id,
			Title:     "PAT Semester",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
	}
	svc := NewAssessmentExamWithStore(store)
	view, err := svc.Create(context.Background(), AssessmentExamInput{Title: "  PAT Semester  "})
	if err != nil {
		t.Fatalf("Create err = %v", err)
	}
	if view.Title != "PAT Semester" || view.Status != AssessmentExamStatusDraft {
		t.Fatalf("view = %+v, want normalized draft title/status", view)
	}
	if store.createArg.Title != "PAT Semester" {
		t.Fatalf("create title = %q, want trimmed", store.createArg.Title)
	}
}

func TestAssessmentExamPrepareRoomsCreatesDefaultRoom(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	sessionID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	roomID := mustPgUUIDAssessmentTest("33333333-3333-3333-3333-333333333333")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "PAT",
			Status:    AssessmentExamStatusDraft,
			StartsAt:  pgtype.Timestamptz{Time: now, Valid: true},
			EndsAt:    pgtype.Timestamptz{Time: now.Add(90 * time.Minute), Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		firstSessionErr: pgx.ErrNoRows,
		createdSession:  db.AssessmentSession{ID: sessionID},
		createdRoom:     db.AssessmentRoom{ID: roomID},
		roomCount:       1,
	}
	svc := NewAssessmentExamWithStore(store)
	result, err := svc.PrepareRooms(context.Background(), examID)
	if err != nil {
		t.Fatalf("PrepareRooms err = %v", err)
	}
	if result.SessionID == "" || result.RoomID == "" || result.RoomCount != 1 {
		t.Fatalf("result = %+v, want default session and room", result)
	}
}

func TestAssessmentExamAssignmentPreviewEmptyParticipants(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "PAT",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		participantCnt: 0,
	}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.AssignmentPreview(context.Background(), examID, AssessmentAssignmentRequest{RoomCount: 3, CapacityPerRoom: 8})
	if err != nil {
		t.Fatalf("AssignmentPreview err = %v", err)
	}
	if result.TotalParticipants != 0 || len(result.Rooms) != 3 {
		t.Fatalf("result = %+v, want three empty rooms and zero participants", result)
	}
	for i, room := range result.Rooms {
		wantCode := []string{"R01", "R02", "R03"}[i]
		if room.Code != wantCode || room.Capacity != 8 || room.AssignedCount != 0 {
			t.Fatalf("room[%d] = %+v, want %s capacity 8 empty", i, room, wantCode)
		}
	}
	if result.Message == "" {
		t.Fatal("message is empty, want clear no-participant message")
	}
}

func TestAssessmentExamAssignmentPreviewReportsCapacityShortage(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "PAT",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		participantCnt: 65,
	}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.AssignmentPreview(context.Background(), examID, AssessmentAssignmentRequest{RoomCount: 2, CapacityPerRoom: 30, MixPolicy: AssessmentMixPolicyMixed})
	if err != nil {
		t.Fatalf("AssignmentPreview err = %v", err)
	}
	if result.AssignedTotal != 60 || result.UnassignedTotal != 5 {
		t.Fatalf("result = %+v, want 60 assigned and 5 unassigned", result)
	}
	if result.Rooms[0].AssignedCount != 30 || result.Rooms[1].AssignedCount != 30 {
		t.Fatalf("rooms = %+v, want capacity-respecting deterministic fill", result.Rooms)
	}
}

func TestAssessmentExamAssignmentPreviewValidation(t *testing.T) {
	svc := NewAssessmentExamWithStore(&fakeAssessmentExamStore{})
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	cases := []AssessmentAssignmentRequest{
		{RoomCount: 0, CapacityPerRoom: 8},
		{RoomCount: 21, CapacityPerRoom: 8},
		{RoomCount: 2, CapacityPerRoom: 0},
		{RoomCount: 2, CapacityPerRoom: 51},
		{RoomCount: 2, CapacityPerRoom: 8, MixPolicy: "invalid"},
	}
	for _, tc := range cases {
		_, err := svc.AssignmentPreview(context.Background(), examID, tc)
		if !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("AssignmentPreview(%+v) err = %v, want bad request", tc, err)
		}
	}
}

func TestAssessmentExamAssignmentApplyUpsertsRoomsOnly(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	sessionID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "PAT",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		firstSessionErr: pgx.ErrNoRows,
		createdSession:  db.AssessmentSession{ID: sessionID},
	}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.AssignmentApply(context.Background(), examID, AssessmentAssignmentRequest{RoomCount: 2, CapacityPerRoom: 12, MixPolicy: AssessmentMixPolicyClassGrouped})
	if err != nil {
		t.Fatalf("AssignmentApply err = %v", err)
	}
	if result.SessionID != sessionID.String() || len(store.createdRooms) != 2 {
		t.Fatalf("result=%+v createdRooms=%+v, want two upserted rooms on created session", result, store.createdRooms)
	}
	if store.createdRooms[0].Code != "R01" || store.createdRooms[1].Code != "R02" || store.createdRooms[0].Capacity != 12 {
		t.Fatalf("createdRooms=%+v, want deterministic R01/R02 capacity 12", store.createdRooms)
	}
	if result.CardCount != 0 {
		t.Fatalf("card count = %d, want no card issuance", result.CardCount)
	}
}

func mustPgUUIDAssessmentTest(raw string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		panic(err)
	}
	return id
}
