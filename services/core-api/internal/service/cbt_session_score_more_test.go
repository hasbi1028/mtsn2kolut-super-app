package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type scoreOrderStore struct {
	calls          []string
	correctnessID  pgtype.UUID
	scoresID       pgtype.UUID
	correctnessErr error
	scoresErr      error
}

func (s *scoreOrderStore) UpdateAnswerCorrectness(ctx context.Context, sessionID pgtype.UUID) error {
	s.calls = append(s.calls, "correctness")
	s.correctnessID = sessionID
	return s.correctnessErr
}

func (s *scoreOrderStore) UpdateParticipantScores(ctx context.Context, sessionID pgtype.UUID) error {
	s.calls = append(s.calls, "scores")
	s.scoresID = sessionID
	return s.scoresErr
}

func TestCbtSessionScoreMoreScoreSessionHelperCallOrder(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(211)

	store := &scoreOrderStore{}
	if err := scoreSession(ctx, store, sessionID); err != nil {
		t.Fatalf("scoreSession() error = %v, want nil", err)
	}
	if got, want := strings.Join(store.calls, ","), "correctness,scores"; got != want {
		t.Fatalf("scoreSession() call order = %q, want %q", got, want)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("scoreSession() ids correctness=%v scores=%v, want %v", store.correctnessID, store.scoresID, sessionID)
	}

	boom := errors.New("correctness failed")
	store = &scoreOrderStore{correctnessErr: boom}
	if err := scoreSession(ctx, store, sessionID); !errors.Is(err, boom) {
		t.Fatalf("scoreSession(correctness error) = %v, want %v", err, boom)
	}
	if got, want := strings.Join(store.calls, ","), "correctness"; got != want {
		t.Fatalf("scoreSession(correctness error) call order = %q, want %q", got, want)
	}
	if store.scoresID.Valid {
		t.Fatalf("scoreSession(correctness error) scores id = %v, want skipped", store.scoresID)
	}
}

func TestCbtSessionScoreMoreGradeEssayRefreshesCorrectnessBeforeScores(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(212)
	answerID := cbtSessionTestUUID(213)
	store := &fakeCbtSessionStore{}

	if err := gradeEssayAndRefreshScore(ctx, store, sessionID, answerID, 99.75, "grader"); err != nil {
		t.Fatalf("gradeEssayAndRefreshScore() error = %v, want nil", err)
	}
	if store.gradeArg.ID != answerID || store.gradeArg.GradedBy.String != "grader" || !store.gradeArg.GradedBy.Valid {
		t.Fatalf("gradeEssayAndRefreshScore() grade arg = %+v, want answer/grader forwarded", store.gradeArg)
	}
	if got := testNumericFloat64(t, store.gradeArg.ManualScore); got != 99.75 {
		t.Fatalf("gradeEssayAndRefreshScore() manual score = %v, want 99.75", got)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("gradeEssayAndRefreshScore() score refresh ids = %v/%v, want %v", store.correctnessID, store.scoresID, sessionID)
	}
}

type forceSubmitUnavailableStore struct {
	cbtSessionStore
}

func TestCbtSessionScoreMoreForceSubmitUnavailableStoreWithoutPool(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(214)
	participantID := cbtSessionTestUUID(215)
	base := &fakeCbtSessionStore{}
	svc := &CbtSession{q: &forceSubmitUnavailableStore{cbtSessionStore: base}}

	row, err := svc.ForceSubmitParticipant(ctx, sessionID, participantID, "proctor")
	if err == nil || !strings.Contains(err.Error(), "cbt participant force submit store unavailable") {
		t.Fatalf("ForceSubmitParticipant(unavailable store) error = %v, want unavailable store error", err)
	}
	if row != (db.ForceSubmitParticipantRow{}) {
		t.Fatalf("ForceSubmitParticipant(unavailable store) row = %+v, want zero row", row)
	}
	if base.participantCorrectnessID.Valid || base.forceSubmitArg.ID.Valid || len(base.insertedEvents) != 0 {
		t.Fatalf("ForceSubmitParticipant(unavailable store) touched base store correctness=%v force=%+v events=%+v", base.participantCorrectnessID, base.forceSubmitArg, base.insertedEvents)
	}
}
