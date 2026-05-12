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

func TestStudentPortalDerivesStudentIDFromUserAccount(t *testing.T) {
	userID := testGenerationUUID(31)
	studentID := testGenerationUUID(32)
	store := &fakeStudentPortalStore{studentID: studentID}
	svc := &StudentPortal{q: store}

	profile, err := svc.Profile(context.Background(), userID)
	if err != nil {
		t.Fatalf("Profile() error = %v", err)
	}
	if store.studentIDUserID != userID {
		t.Fatalf("GetPortalStudentIDByUserID userID = %s, want %s", store.studentIDUserID.String(), userID.String())
	}
	if store.profileStudentID != studentID || profile.ID != studentID {
		t.Fatalf("Profile() used studentID = %s/%s, want %s", store.profileStudentID.String(), profile.ID.String(), studentID.String())
	}

	if _, err := svc.Schedule(context.Background(), userID); err != nil {
		t.Fatalf("Schedule() error = %v", err)
	}
	if store.scheduleStudentID != studentID {
		t.Fatalf("Schedule() used studentID = %s, want %s", store.scheduleStudentID.String(), studentID.String())
	}

	if _, err := svc.Results(context.Background(), userID); err != nil {
		t.Fatalf("Results() error = %v", err)
	}
	if store.resultsStudentID != studentID {
		t.Fatalf("Results() used studentID = %s, want %s", store.resultsStudentID.String(), studentID.String())
	}
}

func TestStudentPortalRejectsUsersWithoutLinkedStudent(t *testing.T) {
	svc := &StudentPortal{q: &fakeStudentPortalStore{studentIDErr: pgx.ErrNoRows}}

	if _, err := svc.Profile(context.Background(), testGenerationUUID(33)); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Profile() error = %v, want ErrForbidden", err)
	}
}

func TestStudentPortalCbtScheduleMasksTokensAndComputesStatus(t *testing.T) {
	userID := testGenerationUUID(36)
	studentID := testGenerationUUID(37)
	roomID := testGenerationUUID(38)
	now := time.Now()
	store := &fakeStudentPortalStore{
		studentID: studentID,
		cbtRows: []db.ListStudentPortalCbtScheduleRow{
			{
				ParticipantID:   testGenerationUUID(39),
				SessionID:       testGenerationUUID(40),
				Token:           "abcdef1234567890",
				RoomID:          roomID,
				SeatNo:          pgtype.Int4{Int32: 12, Valid: true},
				SessionTitle:    "Ujian IPA",
				SessionStatus:   db.CbtSessionStatusEnumActive,
				ScheduledStart:  pgtype.Timestamptz{Time: now.Add(10 * time.Minute), Valid: true},
				ScheduledEnd:    pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
				PackageTitle:    "Paket IPA",
				DurationMinutes: 60,
				RoomName:        "Lab 1",
				RoomToken:       "ROOM-1",
			},
			{
				ParticipantID: testGenerationUUID(41),
				SessionID:     testGenerationUUID(42),
				Token:         "secret-token",
				SubmittedAt:   pgtype.Timestamptz{Time: now, Valid: true},
				SessionStatus: db.CbtSessionStatusEnumActive,
			},
		},
	}
	svc := &StudentPortal{q: store}

	items, err := svc.CbtSchedule(context.Background(), userID)
	if err != nil {
		t.Fatalf("CbtSchedule() error = %v", err)
	}
	if store.cbtStudentID != studentID {
		t.Fatalf("ListStudentPortalCbtSchedule studentID = %s, want %s", store.cbtStudentID.String(), studentID.String())
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Status != StudentPortalCbtTokenWindow || !items[0].CanRevealToken || !items[0].RequiresRoomToken {
		t.Fatalf("first status/reveal = %+v, want token window with room token", items[0])
	}
	if items[0].TokenMasked == nil || *items[0].TokenMasked == "abcdef1234567890" || !strings.Contains(*items[0].TokenMasked, "•") {
		t.Fatalf("masked token = %v, want masked value", items[0].TokenMasked)
	}
	if items[1].Status != StudentPortalCbtSubmitted || items[1].CanRevealToken {
		t.Fatalf("submitted item = %+v, want submitted without reveal", items[1])
	}
}

func TestStudentPortalRevealCbtTokenValidatesOwnershipWindowAndRoomToken(t *testing.T) {
	userID := testGenerationUUID(43)
	studentID := testGenerationUUID(44)
	participantID := testGenerationUUID(45)
	sessionID := testGenerationUUID(46)
	roomID := testGenerationUUID(47)
	now := time.Now()
	baseRow := db.GetStudentPortalCbtParticipantRow{
		ParticipantID:  participantID,
		SessionID:      sessionID,
		StudentID:      studentID,
		Token:          "student-token",
		RoomID:         roomID,
		SessionStatus:  db.CbtSessionStatusEnumActive,
		ScheduledStart: pgtype.Timestamptz{Time: now.Add(5 * time.Minute), Valid: true},
		ScheduledEnd:   pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
		RoomToken:      "ROOM-1",
	}

	store := &fakeStudentPortalStore{studentID: studentID, revealRow: baseRow}
	svc := &StudentPortal{q: store}
	got, err := svc.RevealCbtToken(context.Background(), userID, participantID, " room-1 ", "203.0.113.5")
	if err != nil {
		t.Fatalf("RevealCbtToken() error = %v", err)
	}
	if got.Token != "student-token" || got.ExpiresAt == "" {
		t.Fatalf("RevealCbtToken() = %+v, want token and expiry", got)
	}
	if store.revealArg.ParticipantID != participantID || store.revealArg.StudentID != studentID {
		t.Fatalf("reveal arg = %+v, want participant and owned student", store.revealArg)
	}
	if len(store.participantLogs) != 1 || store.participantLogs[0].EventType != "student_portal_token_reveal" {
		t.Fatalf("events = %+v, want reveal event", store.participantLogs)
	}
	if strings.Contains(string(store.participantLogs[0].EventData), "ROOM-1") || strings.Contains(string(store.participantLogs[0].EventData), "203.0.113.5") {
		t.Fatalf("reveal event leaked sensitive data: %s", string(store.participantLogs[0].EventData))
	}

	store = &fakeStudentPortalStore{studentID: studentID, revealRow: baseRow}
	svc = &StudentPortal{q: store}
	if _, err := svc.RevealCbtToken(context.Background(), userID, participantID, "WRONG", "203.0.113.5"); !errors.Is(err, ErrRoomTokenMismatch) {
		t.Fatalf("RevealCbtToken(wrong room) error = %v, want ErrRoomTokenMismatch", err)
	}
	if len(store.participantLogs) != 1 || store.participantLogs[0].EventType != "student_portal_room_token_mismatch" {
		t.Fatalf("events = %+v, want mismatch event", store.participantLogs)
	}
	var eventData map[string]string
	if err := json.Unmarshal(store.participantLogs[0].EventData, &eventData); err != nil {
		t.Fatalf("event data unmarshal error = %v", err)
	}
	if eventData["reason"] != "room_token_mismatch" || eventData["ip_hash"] == "203.0.113.5" {
		t.Fatalf("event data = %+v, want mismatch reason and hashed IP", eventData)
	}

	closedRow := baseRow
	closedRow.ScheduledStart = pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true}
	store = &fakeStudentPortalStore{studentID: studentID, revealRow: closedRow}
	svc = &StudentPortal{q: store}
	if _, err := svc.RevealCbtToken(context.Background(), userID, participantID, "ROOM-1", ""); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("RevealCbtToken(outside window) error = %v, want ErrForbidden", err)
	}

	store = &fakeStudentPortalStore{studentID: studentID, revealErr: pgx.ErrNoRows}
	svc = &StudentPortal{q: store}
	if _, err := svc.RevealCbtToken(context.Background(), userID, participantID, "ROOM-1", ""); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("RevealCbtToken(other participant) error = %v, want ErrForbidden", err)
	}
}

type fakeStudentPortalStore struct {
	studentID       pgtype.UUID
	studentIDErr    error
	studentIDUserID pgtype.UUID

	profileStudentID  pgtype.UUID
	scheduleStudentID pgtype.UUID
	resultsStudentID  pgtype.UUID
	cbtStudentID      pgtype.UUID
	cbtRows           []db.ListStudentPortalCbtScheduleRow

	revealArg       db.GetStudentPortalCbtParticipantParams
	revealRow       db.GetStudentPortalCbtParticipantRow
	revealErr       error
	participantLogs []db.InsertParticipantEventParams
}

func (f *fakeStudentPortalStore) GetPortalStudentIDByUserID(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error) {
	f.studentIDUserID = userID
	if f.studentIDErr != nil {
		return pgtype.UUID{}, f.studentIDErr
	}
	return f.studentID, nil
}

func (f *fakeStudentPortalStore) ListStudentPortalPreviewStudents(ctx context.Context) ([]db.ListStudentPortalPreviewStudentsRow, error) {
	return []db.ListStudentPortalPreviewStudentsRow{}, nil
}

func (f *fakeStudentPortalStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (db.GetStudentByIDRow, error) {
	f.profileStudentID = id
	return db.GetStudentByIDRow{ID: id, Nama: "Siswa A"}, nil
}

func (f *fakeStudentPortalStore) ListStudentTimetable(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	f.scheduleStudentID = studentID
	return []db.ListStudentTimetableRow{{ID: testGenerationUUID(34), SubjectName: "IPA"}}, nil
}

func (f *fakeStudentPortalStore) ListStudentExamSessions(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error) {
	f.resultsStudentID = studentID
	return []db.ListStudentExamSessionsRow{{SessionID: testGenerationUUID(35), SessionTitle: "Ujian IPA", Token: "secret-token"}}, nil
}

func (f *fakeStudentPortalStore) ListStudentPortalCbtSchedule(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentPortalCbtScheduleRow, error) {
	f.cbtStudentID = studentID
	return f.cbtRows, nil
}

func (f *fakeStudentPortalStore) GetStudentPortalCbtParticipant(ctx context.Context, arg db.GetStudentPortalCbtParticipantParams) (db.GetStudentPortalCbtParticipantRow, error) {
	f.revealArg = arg
	if f.revealErr != nil {
		return db.GetStudentPortalCbtParticipantRow{}, f.revealErr
	}
	return f.revealRow, nil
}

func (f *fakeStudentPortalStore) InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error {
	f.participantLogs = append(f.participantLogs, arg)
	return nil
}
