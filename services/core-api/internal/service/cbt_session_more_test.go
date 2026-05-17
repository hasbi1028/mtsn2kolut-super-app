package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtSessionListParticipantsByTeacherConvertsRowsAndNormalizesNil(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(141)
	teacherID := cbtSessionTestUUID(142)
	participantID := cbtSessionTestUUID(143)
	studentID := cbtSessionTestUUID(144)
	roomID := cbtSessionTestUUID(145)
	store := &fakeCbtSessionStore{
		participantTeacherRows: []db.ListCbtExamParticipantsByTeacherRow{{
			ID:             participantID,
			SessionID:      sessionID,
			StudentID:      studentID,
			Nis:            "001",
			Nama:           "Siswa Guru",
			Token:          "TOKEN-1",
			RoomID:         roomID,
			SeatNo:         pgtype.Int4{Int32: 7, Valid: true},
			Score:          pgNumeric(91.5),
			SuspiciousFlag: true,
			RiskLevel:      "medium",
			RoomName:       "Lab 1",
		}},
	}
	svc := &CbtSession{q: store}

	rows, err := svc.ListParticipantsByTeacher(ctx, sessionID, teacherID)
	if err != nil {
		t.Fatalf("ListParticipantsByTeacher() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListParticipantsByTeacher() rows = %d, want 1", len(rows))
	}
	got := rows[0]
	if got.ID != participantID || got.SessionID != sessionID || got.StudentID != studentID || got.RoomID != roomID || got.Nama != "Siswa Guru" || got.RoomName != "Lab 1" || got.SeatNo.Int32 != 7 || !got.SuspiciousFlag || got.RiskLevel != "medium" {
		t.Fatalf("ListParticipantsByTeacher() row = %+v, want converted participant row", got)
	}
	if testNumericFloat64(t, got.Score) != 91.5 {
		t.Fatalf("ListParticipantsByTeacher() score = %v, want 91.5", got.Score)
	}
	if store.participantTeacherArg != (db.ListCbtExamParticipantsByTeacherParams{TeacherEmployeeID: teacherID, SessionID: sessionID}) {
		t.Fatalf("ListParticipantsByTeacher() arg = %+v, want session and teacher", store.participantTeacherArg)
	}

	store.participantTeacherRows = nil
	rows, err = svc.ListParticipantsByTeacher(ctx, sessionID, teacherID)
	if err != nil {
		t.Fatalf("ListParticipantsByTeacher(nil rows) error = %v", err)
	}
	if rows == nil || len(rows) != 0 {
		t.Fatalf("ListParticipantsByTeacher(nil rows) = %#v, want empty slice", rows)
	}

	boom := errors.New("teacher participants failed")
	store.participantsErr = boom
	if _, err := svc.ListParticipantsByTeacher(ctx, sessionID, teacherID); !errors.Is(err, boom) {
		t.Fatalf("ListParticipantsByTeacher(error) = %v, want %v", err, boom)
	}
}

func TestCbtSessionGradeEssayRefreshesScoresAndStopsOnErrors(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(151)
	answerID := cbtSessionTestUUID(152)
	store := &fakeCbtSessionStore{}
	svc := &CbtSession{q: store}

	if err := svc.GradeEssay(ctx, sessionID, answerID, 88.25, " guru mapel "); err != nil {
		t.Fatalf("GradeEssay() error = %v", err)
	}
	if store.gradeArg.ID != answerID || store.gradeArg.GradedBy.String != " guru mapel " || !store.gradeArg.GradedBy.Valid || testNumericFloat64(t, store.gradeArg.ManualScore) != 88.25 {
		t.Fatalf("GradeEssay() grade arg = %+v, want answer, grader, and score", store.gradeArg)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("GradeEssay() refresh ids = correctness %v scores %v, want %v", store.correctnessID, store.scoresID, sessionID)
	}

	gradeErr := errors.New("grade failed")
	store = &fakeCbtSessionStore{gradeErr: gradeErr}
	if err := (&CbtSession{q: store}).GradeEssay(ctx, sessionID, answerID, 77, "guru"); !errors.Is(err, gradeErr) {
		t.Fatalf("GradeEssay(grade error) = %v, want %v", err, gradeErr)
	}
	if store.correctnessID.Valid || store.scoresID.Valid {
		t.Fatalf("GradeEssay(grade error) refreshed scores correctness=%v scores=%v, want skipped", store.correctnessID, store.scoresID)
	}

	correctnessErr := errors.New("correctness failed")
	store = &fakeCbtSessionStore{correctnessErr: correctnessErr}
	if err := (&CbtSession{q: store}).GradeEssay(ctx, sessionID, answerID, 77, "guru"); !errors.Is(err, correctnessErr) {
		t.Fatalf("GradeEssay(correctness error) = %v, want %v", err, correctnessErr)
	}
	if store.gradeArg.ID != answerID || store.correctnessID != sessionID || store.scoresID.Valid {
		t.Fatalf("GradeEssay(correctness error) calls grade=%+v correctness=%v scores=%v, want grade then correctness only", store.gradeArg, store.correctnessID, store.scoresID)
	}

	scoreErr := errors.New("scores failed")
	store = &fakeCbtSessionStore{scoresErr: scoreErr}
	if err := (&CbtSession{q: store}).GradeEssay(ctx, sessionID, answerID, 77, "guru"); !errors.Is(err, scoreErr) {
		t.Fatalf("GradeEssay(score error) = %v, want %v", err, scoreErr)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("GradeEssay(score error) refresh ids = %v/%v, want %v", store.correctnessID, store.scoresID, sessionID)
	}
}

func TestCbtSessionHandoverSaveAndLockMapNoRowsAndPropagateGenericErrors(t *testing.T) {
	ctx := context.Background()
	roomID := cbtSessionTestUUID(161)
	actorID := cbtSessionTestUUID(162)
	store := &fakeCbtSessionStore{}
	svc := &CbtSession{q: store}

	saved, err := svc.SaveRoomHandover(ctx, roomID, actorID, SaveCbtRoomHandoverInput{
		AttendanceChecked:     true,
		AllSubmittedChecked:   true,
		DeviceIssueChecked:    false,
		RoomCleanChecked:      true,
		TokenReturnedChecked:  true,
		AssetsReturnedChecked: true,
		IncidentNotes:         "  incident notes  ",
		OperatorNotes:         "\noperator notes\t",
		HandoverNotes:         "  final notes  ",
	})
	if err != nil {
		t.Fatalf("SaveRoomHandover() error = %v", err)
	}
	if saved.ExamRoomID != roomID || saved.UpdatedBy != actorID || !saved.AttendanceChecked || !saved.AssetsReturnedChecked {
		t.Fatalf("SaveRoomHandover() row = %+v, want saved handover fields", saved)
	}
	if store.saveHandoverArg.ExamRoomID != roomID || store.saveHandoverArg.UpdatedBy != actorID || store.saveHandoverArg.IncidentNotes != "incident notes" || store.saveHandoverArg.OperatorNotes != "operator notes" || store.saveHandoverArg.HandoverNotes != "final notes" {
		t.Fatalf("SaveRoomHandover() arg = %+v, want trimmed notes", store.saveHandoverArg)
	}

	locked, err := svc.LockRoomHandover(ctx, roomID, actorID)
	if err != nil {
		t.Fatalf("LockRoomHandover() error = %v", err)
	}
	if locked.ExamRoomID != roomID || locked.LockedBy != actorID || store.lockHandoverArg != (db.LockCbtRoomHandoverParams{ExamRoomID: roomID, LockedBy: actorID}) {
		t.Fatalf("LockRoomHandover() row=%+v arg=%+v, want room and locker", locked, store.lockHandoverArg)
	}

	store = &fakeCbtSessionStore{saveHandoverErr: pgx.ErrNoRows}
	if _, err := (&CbtSession{q: store}).SaveRoomHandover(ctx, roomID, actorID, SaveCbtRoomHandoverInput{}); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "sudah dikunci") {
		t.Fatalf("SaveRoomHandover(no rows) = %v, want locked conflict", err)
	}
	genericSaveErr := errors.New("save failed")
	store = &fakeCbtSessionStore{saveHandoverErr: genericSaveErr}
	if _, err := (&CbtSession{q: store}).SaveRoomHandover(ctx, roomID, actorID, SaveCbtRoomHandoverInput{}); !errors.Is(err, genericSaveErr) {
		t.Fatalf("SaveRoomHandover(generic error) = %v, want %v", err, genericSaveErr)
	}

	store = &fakeCbtSessionStore{lockHandoverErr: pgx.ErrNoRows}
	if _, err := (&CbtSession{q: store}).LockRoomHandover(ctx, roomID, actorID); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "sudah dikunci") {
		t.Fatalf("LockRoomHandover(no rows) = %v, want locked conflict", err)
	}
	genericLockErr := errors.New("lock failed")
	store = &fakeCbtSessionStore{lockHandoverErr: genericLockErr}
	if _, err := (&CbtSession{q: store}).LockRoomHandover(ctx, roomID, actorID); !errors.Is(err, genericLockErr) {
		t.Fatalf("LockRoomHandover(generic error) = %v, want %v", err, genericLockErr)
	}
}

func TestCbtSessionPublicRoomAndScoreWrappersGuardBeforeTransactions(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(171)
	activeStore := &fakeCbtSessionStore{sessionRow: db.GetCbtExamSessionRow{ID: sessionID, Status: db.CbtSessionStatusEnumActive}}
	if err := (&CbtSession{q: activeStore}).ShuffleRooms(ctx, sessionID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("ShuffleRooms(active session) = %v, want ErrConflict before transaction", err)
	}
	if activeStore.clearRoomID.Valid || len(activeStore.assignRoomArgs) != 0 {
		t.Fatalf("ShuffleRooms(active session) touched shuffle store clear=%v assign=%+v, want blocked", activeStore.clearRoomID, activeStore.assignRoomArgs)
	}

	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("ScoreSession(nil pool) did not panic; update test if fake-store public scoring path is added")
		}
	}()
	_ = (&CbtSession{q: &fakeCbtSessionStore{}}).ScoreSession(ctx, sessionID)
}
