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
	listRows              []db.ListAssessmentExamsRow
	getRow                db.GetAssessmentExamRow
	getErr                error
	created               db.AssessmentExam
	updated               db.AssessmentExam
	firstSession          db.AssessmentSession
	firstSessionErr       error
	createdSession        db.AssessmentSession
	createdRoom           db.AssessmentRoom
	createdRooms          []db.CreateAssessmentRoomParams
	candidateStudents     []db.ListAssessmentCandidateStudentsByClassIDsRow
	participants          []db.ListAssessmentParticipantsForAssignmentRow
	upsertedParticipants  []db.UpsertAssessmentParticipantParams
	deletedOutsideClasses db.DeleteAssessmentParticipantsOutsideClassIDsParams
	assignedParticipants  []db.AssignAssessmentParticipantRoomParams
	placementRows         []db.ListAssessmentParticipantPlacementsByExamRow
	movedParticipantArg   db.MoveAssessmentParticipantSeatParams
	movedParticipant      db.MoveAssessmentParticipantSeatRow
	cardTargets           []db.ListAssessmentParticipantCardTargetsByExamRow
	createdCards          []db.CreateAssessmentParticipantAccessCardParams
	roomForMove           db.GetAssessmentRoomByIDAndExamRow
	roomForMoveErr        error
	clearedParticipantSet bool
	roomCount             int64
	participantCnt        int64
	cardCount             int64
	packageMapRows        []db.ListAssessmentExamPackageMapsRow
	packageOptions        []db.ListAssessmentPackageOptionsRow
	upsertedPackageMaps   []db.UpsertAssessmentExamPackageMapParams
	deletePackageMapArg   db.DeleteAssessmentExamPackageMapParams
	createArg             db.CreateAssessmentExamParams
	updateArg             db.UpdateAssessmentExamParams
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
	roomID := mustPgUUIDAssessmentTest("33333333-3333-3333-3333-333333333300")
	if len(arg.Code) == 3 && arg.Code[0] == 'R' {
		roomID.Bytes[15] = (arg.Code[1]-'0')*10 + (arg.Code[2] - '0')
	}
	return db.AssessmentRoom{ID: roomID, SessionID: arg.SessionID, Code: arg.Code, Name: arg.Name, Capacity: arg.Capacity}, nil
}

func (f *fakeAssessmentExamStore) ListAssessmentCandidateStudentsByClassIDs(context.Context, []pgtype.UUID) ([]db.ListAssessmentCandidateStudentsByClassIDsRow, error) {
	return f.candidateStudents, nil
}

func (f *fakeAssessmentExamStore) UpsertAssessmentParticipant(_ context.Context, arg db.UpsertAssessmentParticipantParams) (db.AssessmentParticipant, error) {
	f.upsertedParticipants = append(f.upsertedParticipants, arg)
	return db.AssessmentParticipant{ID: arg.StudentID, SessionID: arg.SessionID, StudentID: arg.StudentID, Status: "registered"}, nil
}

func (f *fakeAssessmentExamStore) DeleteAssessmentParticipantsOutsideClassIDs(_ context.Context, arg db.DeleteAssessmentParticipantsOutsideClassIDsParams) (int64, error) {
	f.deletedOutsideClasses = arg
	return 0, nil
}

func (f *fakeAssessmentExamStore) ListAssessmentParticipantsForAssignment(context.Context, pgtype.UUID) ([]db.ListAssessmentParticipantsForAssignmentRow, error) {
	return f.participants, nil
}

func (f *fakeAssessmentExamStore) ClearAssessmentParticipantRooms(context.Context, pgtype.UUID) error {
	f.clearedParticipantSet = true
	return nil
}

func (f *fakeAssessmentExamStore) AssignAssessmentParticipantRoom(_ context.Context, arg db.AssignAssessmentParticipantRoomParams) error {
	f.assignedParticipants = append(f.assignedParticipants, arg)
	return nil
}

func (f *fakeAssessmentExamStore) ListAssessmentParticipantCardTargetsByExam(context.Context, pgtype.UUID) ([]db.ListAssessmentParticipantCardTargetsByExamRow, error) {
	return f.cardTargets, nil
}

func (f *fakeAssessmentExamStore) CreateAssessmentParticipantAccessCard(_ context.Context, arg db.CreateAssessmentParticipantAccessCardParams) (db.AssessmentAccessCard, error) {
	f.createdCards = append(f.createdCards, arg)
	cardID := mustPgUUIDAssessmentTest("cccccccc-0000-0000-0000-000000000001")
	return db.AssessmentAccessCard{ID: cardID, CardType: "participant", SessionID: arg.SessionID, ParticipantID: arg.ParticipantID, TokenHash: arg.TokenHash, PinHash: arg.PinHash, Status: "active"}, nil
}

func (f *fakeAssessmentExamStore) ListAssessmentParticipantPlacementsByExam(context.Context, pgtype.UUID) ([]db.ListAssessmentParticipantPlacementsByExamRow, error) {
	return f.placementRows, nil
}

func (f *fakeAssessmentExamStore) GetAssessmentRoomByIDAndExam(context.Context, db.GetAssessmentRoomByIDAndExamParams) (db.GetAssessmentRoomByIDAndExamRow, error) {
	return f.roomForMove, f.roomForMoveErr
}

func (f *fakeAssessmentExamStore) MoveAssessmentParticipantSeat(_ context.Context, arg db.MoveAssessmentParticipantSeatParams) (db.MoveAssessmentParticipantSeatRow, error) {
	f.movedParticipantArg = arg
	return f.movedParticipant, nil
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

func (f *fakeAssessmentExamStore) ListAssessmentExamPackageMaps(context.Context, pgtype.UUID) ([]db.ListAssessmentExamPackageMapsRow, error) {
	return f.packageMapRows, nil
}

func (f *fakeAssessmentExamStore) UpsertAssessmentExamPackageMap(_ context.Context, arg db.UpsertAssessmentExamPackageMapParams) (db.AssessmentExamPackageMap, error) {
	f.upsertedPackageMaps = append(f.upsertedPackageMaps, arg)
	return db.AssessmentExamPackageMap{ID: mustPgUUIDAssessmentTest("99999999-9999-9999-9999-999999999999"), ExamID: arg.ExamID, ClassID: arg.ClassID, SubjectID: arg.SubjectID, PackageID: arg.PackageID, SlotLabel: arg.SlotLabel, Notes: arg.Notes}, nil
}

func (f *fakeAssessmentExamStore) DeleteAssessmentExamPackageMap(_ context.Context, arg db.DeleteAssessmentExamPackageMapParams) (int64, error) {
	f.deletePackageMapArg = arg
	return 1, nil
}

func (f *fakeAssessmentExamStore) ListAssessmentPackageOptions(context.Context, pgtype.UUID) ([]db.ListAssessmentPackageOptionsRow, error) {
	return f.packageOptions, nil
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

func TestAssessmentExamAssignmentPreviewBalancesParticipantsAcrossAllRooms(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "UAS Merata",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		participantCnt: 120,
	}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.AssignmentPreview(context.Background(), examID, AssessmentAssignmentRequest{RoomCount: 8, CapacityPerRoom: 30, MixPolicy: AssessmentMixPolicyMixed})
	if err != nil {
		t.Fatalf("AssignmentPreview err = %v", err)
	}
	if result.AssignedTotal != 120 || result.UnassignedTotal != 0 {
		t.Fatalf("result = %+v, want all 120 participants assigned", result)
	}
	for i, room := range result.Rooms {
		if room.AssignedCount != 15 {
			t.Fatalf("room[%d] = %+v, want balanced 15 participants per room", i, room)
		}
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

func TestAssessmentExamAssignmentApplyBalancesParticipantsAcrossAllRooms(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	sessionID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	participants := make([]db.ListAssessmentParticipantsForAssignmentRow, 0, 120)
	for i := 1; i <= 120; i++ {
		participantID := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001")
		participantID.Bytes[14] = byte(i / 256)
		participantID.Bytes[15] = byte(i % 256)
		participants = append(participants, db.ListAssessmentParticipantsForAssignmentRow{
			ParticipantID: participantID,
			SessionID:     sessionID,
			StudentID:     participantID,
			StudentName:   "Siswa",
			ClassCode:     "9A",
			ClassName:     "IX A",
			GradeLevel:    9,
		})
	}
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "UAS Merata",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		firstSession: db.AssessmentSession{ID: sessionID},
		participants: participants,
	}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.AssignmentApply(context.Background(), examID, AssessmentAssignmentRequest{RoomCount: 8, CapacityPerRoom: 30, MixPolicy: AssessmentMixPolicyMixed})
	if err != nil {
		t.Fatalf("AssignmentApply err = %v", err)
	}
	if result.AssignedTotal != 120 || result.UnassignedTotal != 0 || len(store.assignedParticipants) != 120 {
		t.Fatalf("result=%+v assigned=%d, want all 120 assigned", result, len(store.assignedParticipants))
	}
	roomCounts := map[byte]int{}
	for _, assigned := range store.assignedParticipants {
		roomCounts[assigned.RoomID.Bytes[15]]++
	}
	for roomNo := byte(1); roomNo <= 8; roomNo++ {
		if roomCounts[roomNo] != 15 {
			t.Fatalf("roomCounts = %+v, want room %d to contain 15 participants", roomCounts, roomNo)
		}
	}
	for i, room := range result.Rooms {
		if room.AssignedCount != 15 {
			t.Fatalf("result room[%d] = %+v, want balanced 15 participants", i, room)
		}
	}
}

func TestAssessmentExamAssignmentPreviewCountsSelectedClassStudents(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	classID := mustPgUUIDAssessmentTest("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "PAT",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		candidateStudents: []db.ListAssessmentCandidateStudentsByClassIDsRow{
			{StudentID: mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001"), ClassID: classID, StudentName: "A"},
			{StudentID: mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000002"), ClassID: classID, StudentName: "B"},
			{StudentID: mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000003"), ClassID: classID, StudentName: "C"},
		},
	}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.AssignmentPreview(context.Background(), examID, AssessmentAssignmentRequest{RoomCount: 2, CapacityPerRoom: 2, ClassIDs: []string{classID.String()}})
	if err != nil {
		t.Fatalf("AssignmentPreview err = %v", err)
	}
	if result.TotalParticipants != 3 || result.AssignedTotal != 3 || result.UnassignedTotal != 0 {
		t.Fatalf("result = %+v, want three selected-class participants fully assigned", result)
	}
}

func TestAssessmentExamAssignmentApplyEnrollsClassStudentsAndAssignsSeats(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	sessionID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	classID := mustPgUUIDAssessmentTest("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	student1 := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001")
	student2 := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000002")
	student3 := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000003")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "PAT",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		firstSession: db.AssessmentSession{ID: sessionID},
		candidateStudents: []db.ListAssessmentCandidateStudentsByClassIDsRow{
			{StudentID: student1, ClassID: classID, StudentName: "A"},
			{StudentID: student2, ClassID: classID, StudentName: "B"},
			{StudentID: student3, ClassID: classID, StudentName: "C"},
		},
		participants: []db.ListAssessmentParticipantsForAssignmentRow{
			{ParticipantID: student1, SessionID: sessionID, StudentID: student1, StudentName: "A", ClassCode: "7A", ClassName: "VII A", GradeLevel: 7},
			{ParticipantID: student2, SessionID: sessionID, StudentID: student2, StudentName: "B", ClassCode: "7A", ClassName: "VII A", GradeLevel: 7},
			{ParticipantID: student3, SessionID: sessionID, StudentID: student3, StudentName: "C", ClassCode: "7A", ClassName: "VII A", GradeLevel: 7},
		},
	}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.AssignmentApply(context.Background(), examID, AssessmentAssignmentRequest{RoomCount: 2, CapacityPerRoom: 2, ClassIDs: []string{classID.String()}})
	if err != nil {
		t.Fatalf("AssignmentApply err = %v", err)
	}
	if result.AssignedTotal != 3 || len(store.upsertedParticipants) != 3 || len(store.assignedParticipants) != 3 {
		t.Fatalf("result=%+v upserted=%d assigned=%d, want three participants enrolled and assigned", result, len(store.upsertedParticipants), len(store.assignedParticipants))
	}
	if store.deletedOutsideClasses.SessionID != sessionID || len(store.deletedOutsideClasses.ClassIds) != 1 || store.deletedOutsideClasses.ClassIds[0] != classID {
		t.Fatalf("deletedOutsideClasses=%+v, want session scoped sync to selected class", store.deletedOutsideClasses)
	}
	if !store.clearedParticipantSet {
		t.Fatal("participants were not cleared before reassignment")
	}
	if store.assignedParticipants[0].SeatNo.Int32 != 1 || store.assignedParticipants[1].SeatNo.Int32 != 2 || store.assignedParticipants[2].SeatNo.Int32 != 1 {
		t.Fatalf("assigned seats = %+v, want R01 seats 1-2 then R02 seat 1", store.assignedParticipants)
	}
	if store.assignedParticipants[0].RoomID == store.assignedParticipants[2].RoomID {
		t.Fatalf("assigned rooms = %+v, want third participant in second room", store.assignedParticipants)
	}
	if len(result.Rooms[0].ClassSummary) != 1 || result.Rooms[0].ClassSummary[0].ClassCode != "7A" || result.Rooms[0].ClassSummary[0].Count != 2 {
		t.Fatalf("room 1 class summary = %+v, want 2 students from 7A", result.Rooms[0].ClassSummary)
	}
	if len(result.Rooms[1].ClassSummary) != 1 || result.Rooms[1].ClassSummary[0].Count != 1 {
		t.Fatalf("room 2 class summary = %+v, want 1 remaining student", result.Rooms[1].ClassSummary)
	}
	if result.CardCount != 0 {
		t.Fatalf("card count = %d, want no card issuance", result.CardCount)
	}
}

func TestAssessmentExamAssignmentApplyMixedPolicyInterleavesRombelPerRoom(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	sessionID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	classes := []string{"VIIA", "VIIB", "VIIIA", "VIIIC"}
	participants := make([]db.ListAssessmentParticipantsForAssignmentRow, 0, 20)
	for _, classCode := range classes {
		for i := 1; i <= 5; i++ {
			participantID := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001")
			participantID.Bytes[15] = byte(len(participants) + 1)
			grade := int32(7)
			if classCode == "VIIIA" || classCode == "VIIIC" {
				grade = 8
			}
			participants = append(participants, db.ListAssessmentParticipantsForAssignmentRow{
				ParticipantID: participantID,
				SessionID:     sessionID,
				StudentID:     participantID,
				StudentName:   classCode,
				ClassCode:     classCode,
				ClassName:     classCode,
				GradeLevel:    grade,
			})
		}
	}
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "Gladi Campur",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		firstSession: db.AssessmentSession{ID: sessionID},
		participants: participants,
	}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.AssignmentApply(context.Background(), examID, AssessmentAssignmentRequest{RoomCount: 1, CapacityPerRoom: 20, MixPolicy: AssessmentMixPolicyMixed})
	if err != nil {
		t.Fatalf("AssignmentApply err = %v", err)
	}
	if len(result.Rooms) != 1 {
		t.Fatalf("rooms = %+v, want one room", result.Rooms)
	}
	counts := map[string]int64{}
	for _, item := range result.Rooms[0].ClassSummary {
		counts[item.ClassCode] = item.Count
	}
	for _, classCode := range classes {
		if counts[classCode] != 5 {
			t.Fatalf("class summary = %+v, want %s count 5 in the same room", result.Rooms[0].ClassSummary, classCode)
		}
	}
	assignedByClass := map[string]int{}
	for seatIndex, assigned := range store.assignedParticipants {
		original := participants[int(assigned.ParticipantID.Bytes[15])-1]
		assignedByClass[original.ClassCode]++
		if seatIndex < 4 && assignedByClass[original.ClassCode] != 1 {
			t.Fatalf("first four seats = %+v, want one student from each rombel before repeating", store.assignedParticipants[:4])
		}
	}
}

func TestAssessmentExamListParticipantPlacementsReturnsManualEditRows(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	participantID := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001")
	roomID := mustPgUUIDAssessmentTest("33333333-3333-3333-3333-333333333333")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "Gladi Manual",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		placementRows: []db.ListAssessmentParticipantPlacementsByExamRow{{
			ParticipantID: participantID,
			StudentName:   "Ahmad",
			ClassCode:     "VIIA",
			ClassName:     "VII A",
			GradeLevel:    7,
			RoomID:        roomID,
			RoomCode:      "R01",
			RoomName:      "Ruang Ujian 1",
			RoomCapacity:  30,
			SeatNo:        12,
		}},
	}
	svc := NewAssessmentExamWithStore(store)

	items, err := svc.ListParticipantPlacements(context.Background(), examID)
	if err != nil {
		t.Fatalf("ListParticipantPlacements err = %v", err)
	}
	if len(items) != 1 || items[0].ParticipantID != participantID.String() || items[0].RoomCode != "R01" || items[0].SeatNo != 12 {
		t.Fatalf("items = %+v, want manual placement row with room and seat", items)
	}
}

func TestAssessmentExamMoveParticipantSeatValidatesRoomAndSeat(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	participantID := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001")
	roomID := mustPgUUIDAssessmentTest("33333333-3333-3333-3333-333333333333")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{
			ID:        examID,
			Title:     "Gladi Manual",
			Status:    AssessmentExamStatusDraft,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		},
		roomForMove: db.GetAssessmentRoomByIDAndExamRow{ID: roomID, Capacity: 30},
		movedParticipant: db.MoveAssessmentParticipantSeatRow{
			ParticipantID: participantID,
			StudentName:   "Ahmad",
			RoomID:        roomID,
			RoomCode:      "R02",
			SeatNo:        7,
		},
	}
	svc := NewAssessmentExamWithStore(store)

	item, err := svc.MoveParticipantSeat(context.Background(), examID, AssessmentParticipantSeatInput{ParticipantID: participantID.String(), RoomID: roomID.String(), SeatNo: 7})
	if err != nil {
		t.Fatalf("MoveParticipantSeat err = %v", err)
	}
	if item.ParticipantID != participantID.String() || item.RoomID != roomID.String() || item.SeatNo != 7 {
		t.Fatalf("item = %+v, want moved participant", item)
	}
	if store.movedParticipantArg.ParticipantID != participantID || store.movedParticipantArg.RoomID != roomID || store.movedParticipantArg.SeatNo.Int32 != 7 {
		t.Fatalf("move arg = %+v, want participant/room/seat", store.movedParticipantArg)
	}

	_, err = svc.MoveParticipantSeat(context.Background(), examID, AssessmentParticipantSeatInput{ParticipantID: participantID.String(), RoomID: roomID.String(), SeatNo: 31})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("MoveParticipantSeat over capacity err = %v, want bad request", err)
	}
}

func TestAssessmentExamIssueParticipantCardsCreatesSeparateQRPIN(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	sessionID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	participantID := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001")
	studentID := mustPgUUIDAssessmentTest("bbbbbbbb-0000-0000-0000-000000000001")
	roomID := mustPgUUIDAssessmentTest("33333333-3333-3333-3333-333333333333")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{ID: examID, Title: "Gladi", Status: AssessmentExamStatusDraft, CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}},
		cardTargets: []db.ListAssessmentParticipantCardTargetsByExamRow{{
			ParticipantID: participantID,
			SessionID:     sessionID,
			StudentID:     studentID,
			StudentName:   "Ahmad",
			Nis:           "123",
			Nisn:          "456",
			ClassCode:     "VIIA",
			ClassName:     "VII A",
			GradeLevel:    7,
			RoomID:        roomID,
			RoomCode:      "R01",
			RoomName:      "Ruang Ujian 1",
			SeatNo:        3,
			Status:        "registered",
		}},
	}
	svc := NewAssessmentExamWithStore(store)

	cards, err := svc.IssueParticipantCards(context.Background(), examID, AssessmentCardIssueInput{})
	if err != nil {
		t.Fatalf("IssueParticipantCards err = %v", err)
	}
	if len(cards.Cards) != 1 || cards.Cards[0].Token == "" || cards.Cards[0].PIN == "" || cards.Cards[0].QRPath == "" {
		t.Fatalf("cards = %+v, want newly issued raw token, PIN, and QR path", cards)
	}
	if len(store.createdCards) != 1 {
		t.Fatalf("created cards = %d, want one", len(store.createdCards))
	}
	if store.createdCards[0].ParticipantID != participantID || store.createdCards[0].SessionID != sessionID || store.createdCards[0].TokenHash == cards.Cards[0].Token || store.createdCards[0].PinHash == cards.Cards[0].PIN {
		t.Fatalf("created card arg = %+v, card=%+v; want hashed secrets for participant", store.createdCards[0], cards.Cards[0])
	}
}

func TestAssessmentExamIssueParticipantCardsSkipsExistingUnlessRegenerate(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	sessionID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	participantID := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001")
	cardID := mustPgUUIDAssessmentTest("cccccccc-0000-0000-0000-000000000001")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow: db.GetAssessmentExamRow{ID: examID, Title: "Gladi", Status: AssessmentExamStatusDraft, CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}},
		cardTargets: []db.ListAssessmentParticipantCardTargetsByExamRow{{
			ParticipantID: participantID,
			SessionID:     sessionID,
			StudentName:   "Ahmad",
			CardID:        cardID,
			CardStatus:    "active",
		}},
	}
	svc := NewAssessmentExamWithStore(store)

	cards, err := svc.IssueParticipantCards(context.Background(), examID, AssessmentCardIssueInput{})
	if err != nil {
		t.Fatalf("IssueParticipantCards existing err = %v", err)
	}
	if len(cards.Cards) != 1 || cards.Cards[0].Token != "" || cards.Cards[0].PIN != "" || len(store.createdCards) != 0 {
		t.Fatalf("cards=%+v created=%d, want existing card without raw secrets and no insert", cards, len(store.createdCards))
	}

	cards, err = svc.IssueParticipantCards(context.Background(), examID, AssessmentCardIssueInput{Regenerate: true})
	if err != nil {
		t.Fatalf("IssueParticipantCards regenerate err = %v", err)
	}
	if len(cards.Cards) != 1 || cards.Cards[0].Token == "" || cards.Cards[0].PIN == "" || len(store.createdCards) != 1 {
		t.Fatalf("cards=%+v created=%d, want regenerated raw secrets", cards, len(store.createdCards))
	}
}

func TestAssessmentExamListParticipantCardsNeverReturnsSecrets(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	sessionID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	participantID := mustPgUUIDAssessmentTest("aaaaaaaa-0000-0000-0000-000000000001")
	cardID := mustPgUUIDAssessmentTest("cccccccc-0000-0000-0000-000000000001")
	now := time.Date(2026, 5, 30, 8, 0, 0, 0, time.UTC)
	store := &fakeAssessmentExamStore{
		getRow:      db.GetAssessmentExamRow{ID: examID, Title: "Gladi", Status: AssessmentExamStatusDraft, CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}},
		cardTargets: []db.ListAssessmentParticipantCardTargetsByExamRow{{ParticipantID: participantID, SessionID: sessionID, StudentName: "Ahmad", CardID: cardID, CardStatus: "active"}},
	}
	svc := NewAssessmentExamWithStore(store)

	cards, err := svc.ListParticipantCards(context.Background(), examID)
	if err != nil {
		t.Fatalf("ListParticipantCards err = %v", err)
	}
	if len(cards) != 1 || cards[0].CardID == "" || cards[0].Token != "" || cards[0].PIN != "" {
		t.Fatalf("cards = %+v, want card metadata without raw secrets", cards)
	}
}

func mustPgUUIDAssessmentTest(raw string) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		panic(err)
	}
	return id
}

func TestAssessmentExamPackageMapsSharedPackagePerRombel(t *testing.T) {
	examID := mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111")
	classA := mustPgUUIDAssessmentTest("aaaaaaaa-1111-1111-1111-111111111111")
	classB := mustPgUUIDAssessmentTest("bbbbbbbb-1111-1111-1111-111111111111")
	subjectID := mustPgUUIDAssessmentTest("22222222-2222-2222-2222-222222222222")
	packageID := mustPgUUIDAssessmentTest("33333333-3333-3333-3333-333333333333")
	store := &fakeAssessmentExamStore{getRow: db.GetAssessmentExamRow{ID: examID, Title: "PAT", Status: "draft"}}
	svc := NewAssessmentExamWithStore(store)

	result, err := svc.SavePackageMaps(context.Background(), examID, AssessmentPackageMapRequest{Items: []AssessmentPackageMapInput{
		{ClassID: assessmentUUIDString(classA), SubjectID: assessmentUUIDString(subjectID), PackageID: assessmentUUIDString(packageID), SlotLabel: "Hari 1 - Matematika"},
		{ClassID: assessmentUUIDString(classB), SubjectID: assessmentUUIDString(subjectID), PackageID: assessmentUUIDString(packageID), SlotLabel: "Hari 1 - Matematika"},
	}})
	if err != nil {
		t.Fatalf("SavePackageMaps returned error: %v", err)
	}
	if result.Count != 2 {
		t.Fatalf("saved count = %d, want 2", result.Count)
	}
	if len(store.upsertedPackageMaps) != 2 {
		t.Fatalf("upsert calls = %d, want 2", len(store.upsertedPackageMaps))
	}
	for i, arg := range store.upsertedPackageMaps {
		if arg.ExamID != examID || arg.SubjectID != subjectID || arg.PackageID != packageID {
			t.Fatalf("upsert[%d] = %+v, want shared package for same subject", i, arg)
		}
	}
}

func TestAssessmentExamPackageMapsRejectMissingIDs(t *testing.T) {
	svc := NewAssessmentExamWithStore(&fakeAssessmentExamStore{})
	_, err := svc.SavePackageMaps(context.Background(), mustPgUUIDAssessmentTest("11111111-1111-1111-1111-111111111111"), AssessmentPackageMapRequest{Items: []AssessmentPackageMapInput{{ClassID: "", SubjectID: "bad", PackageID: ""}}})
	if err == nil {
		t.Fatal("SavePackageMaps error = nil, want validation error")
	}
}
