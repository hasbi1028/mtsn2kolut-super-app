package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtSessionScoreSessionHelperPaths(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(100)

	t.Run("success updates correctness before scores", func(t *testing.T) {
		store := &fakeCbtSessionStore{}
		if err := scoreSession(ctx, store, sessionID); err != nil {
			t.Fatalf("scoreSession() error = %v, want nil", err)
		}
		if store.correctnessID != sessionID || store.scoresID != sessionID {
			t.Fatalf("scoreSession() ids correctness=%v scores=%v, want %v", store.correctnessID, store.scoresID, sessionID)
		}
	})

	t.Run("correctness error skips score update", func(t *testing.T) {
		boom := errors.New("correctness failed")
		store := &fakeCbtSessionStore{correctnessErr: boom}
		if err := scoreSession(ctx, store, sessionID); !errors.Is(err, boom) {
			t.Fatalf("scoreSession() error = %v, want %v", err, boom)
		}
		if store.correctnessID != sessionID {
			t.Fatalf("scoreSession() correctness id = %v, want %v", store.correctnessID, sessionID)
		}
		if store.scoresID.Valid {
			t.Fatalf("scoreSession() scores id = %v, want score update skipped", store.scoresID)
		}
	})

	t.Run("score update error is returned after correctness", func(t *testing.T) {
		boom := errors.New("scores failed")
		store := &fakeCbtSessionStore{scoresErr: boom}
		if err := scoreSession(ctx, store, sessionID); !errors.Is(err, boom) {
			t.Fatalf("scoreSession() error = %v, want %v", err, boom)
		}
		if store.correctnessID != sessionID || store.scoresID != sessionID {
			t.Fatalf("scoreSession() ids correctness=%v scores=%v, want %v", store.correctnessID, store.scoresID, sessionID)
		}
	})
}

type gradeEssayErrorStore struct {
	*fakeCbtSessionStore
	err error
}

func (s *gradeEssayErrorStore) GradeStudentEssay(ctx context.Context, arg db.GradeStudentEssayParams) error {
	s.fakeCbtSessionStore.gradeArg = arg
	return s.err
}

func TestCbtSessionGradeEssayScoringPathsWithoutPool(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(110)
	answerID := cbtSessionTestUUID(111)

	t.Run("success grades essay and refreshes session score", func(t *testing.T) {
		store := &fakeCbtSessionStore{}
		svc := &CbtSession{q: store}

		if err := svc.GradeEssay(ctx, sessionID, answerID, 88.25, "guru-a"); err != nil {
			t.Fatalf("GradeEssay() error = %v, want nil", err)
		}
		if store.gradeArg.ID != answerID || store.gradeArg.GradedBy.String != "guru-a" || !store.gradeArg.GradedBy.Valid {
			t.Fatalf("GradeEssay() grade arg = %+v, want answer and grader", store.gradeArg)
		}
		if got := testNumericFloat64(t, store.gradeArg.ManualScore); got != 88.25 {
			t.Fatalf("GradeEssay() manual score = %v, want 88.25", got)
		}
		if store.correctnessID != sessionID || store.scoresID != sessionID {
			t.Fatalf("GradeEssay() scoring ids correctness=%v scores=%v, want %v", store.correctnessID, store.scoresID, sessionID)
		}
	})

	t.Run("grade error skips scoring", func(t *testing.T) {
		boom := errors.New("grade failed")
		base := &fakeCbtSessionStore{}
		store := &gradeEssayErrorStore{fakeCbtSessionStore: base, err: boom}
		svc := &CbtSession{q: store}

		if err := svc.GradeEssay(ctx, sessionID, answerID, 70, "guru-b"); !errors.Is(err, boom) {
			t.Fatalf("GradeEssay() error = %v, want %v", err, boom)
		}
		if base.gradeArg.ID != answerID {
			t.Fatalf("GradeEssay() grade arg id = %v, want %v", base.gradeArg.ID, answerID)
		}
		if base.correctnessID.Valid || base.scoresID.Valid {
			t.Fatalf("GradeEssay() scoring ids correctness=%v scores=%v, want skipped", base.correctnessID, base.scoresID)
		}
	})

	t.Run("score error is returned after grade", func(t *testing.T) {
		boom := errors.New("refresh scores failed")
		store := &fakeCbtSessionStore{scoresErr: boom}
		svc := &CbtSession{q: store}

		if err := svc.GradeEssay(ctx, sessionID, answerID, 72, "guru-c"); !errors.Is(err, boom) {
			t.Fatalf("GradeEssay() error = %v, want %v", err, boom)
		}
		if store.gradeArg.ID != answerID || store.correctnessID != sessionID || store.scoresID != sessionID {
			t.Fatalf("GradeEssay() calls grade=%v correctness=%v scores=%v", store.gradeArg.ID, store.correctnessID, store.scoresID)
		}
	})
}

func TestCbtSessionForceSubmitParticipantExtraPathsWithoutPool(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(120)
	participantID := cbtSessionTestUUID(121)

	t.Run("correctness error skips force submit and event", func(t *testing.T) {
		boom := errors.New("participant correctness failed")
		store := &fakeCbtSessionStore{participantCorrectnessErr: boom}
		svc := &CbtSession{q: store}

		if _, err := svc.ForceSubmitParticipant(ctx, sessionID, participantID, "proctor"); !errors.Is(err, boom) {
			t.Fatalf("ForceSubmitParticipant() error = %v, want %v", err, boom)
		}
		if got, want := strings.Join(store.forceSubmitCalls, ","), "update_correctness"; got != want {
			t.Fatalf("ForceSubmitParticipant() call order = %q, want %q", got, want)
		}
		if store.forceSubmitArg.ID.Valid || len(store.insertedEvents) != 0 {
			t.Fatalf("ForceSubmitParticipant() force arg=%+v events=%+v, want skipped", store.forceSubmitArg, store.insertedEvents)
		}
	})

	t.Run("force submit error skips event", func(t *testing.T) {
		boom := errors.New("force submit failed")
		store := &fakeCbtSessionStore{forceSubmitErr: boom}
		svc := &CbtSession{q: store}

		if _, err := svc.ForceSubmitParticipant(ctx, sessionID, participantID, "proctor"); !errors.Is(err, boom) {
			t.Fatalf("ForceSubmitParticipant() error = %v, want %v", err, boom)
		}
		if got, want := strings.Join(store.forceSubmitCalls, ","), "update_correctness,force_submit"; got != want {
			t.Fatalf("ForceSubmitParticipant() call order = %q, want %q", got, want)
		}
		if len(store.insertedEvents) != 0 {
			t.Fatalf("ForceSubmitParticipant() events=%+v, want skipped", store.insertedEvents)
		}
	})

	t.Run("event insert error is returned after submitted row", func(t *testing.T) {
		boom := errors.New("event insert failed")
		store := &fakeCbtSessionStore{
			forceSubmitRow: db.ForceSubmitParticipantRow{ID: participantID, SubmittedAt: pgtype.Timestamptz{Time: time.Unix(200, 0), Valid: true}},
			insertEventErr: boom,
		}
		svc := &CbtSession{q: store}

		row, err := svc.ForceSubmitParticipant(ctx, sessionID, participantID, "proctor")
		if !errors.Is(err, boom) {
			t.Fatalf("ForceSubmitParticipant() error = %v, want %v", err, boom)
		}
		if row.ID.Valid {
			t.Fatalf("ForceSubmitParticipant() row on event error = %+v, want zero row", row)
		}
		if len(store.insertedEvents) != 1 || store.insertedEvents[0].EventType != "proctor_force_submit" {
			t.Fatalf("ForceSubmitParticipant() events=%+v, want attempted force-submit event", store.insertedEvents)
		}
	})
}

func TestCbtSessionUpdateStatusExtraPathsWithoutPool(t *testing.T) {
	ctx := context.Background()
	sessionID := cbtSessionTestUUID(130)
	packageID := cbtSessionTestUUID(131)

	t.Run("invalid status is bad request and skips store", func(t *testing.T) {
		store := &fakeCbtSessionStore{}
		svc := &CbtSession{q: store}
		_, err := svc.UpdateStatus(ctx, sessionID, db.CbtSessionStatusEnum("bogus"))
		if !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("UpdateStatus() error = %v, want ErrBadRequest", err)
		}
		if store.getID.Valid || store.updateStatusArg.ID.Valid {
			t.Fatalf("UpdateStatus() touched store get=%v update=%+v, want skipped", store.getID, store.updateStatusArg)
		}
	})

	t.Run("scheduled validates package readiness before status update", func(t *testing.T) {
		store := &fakeCbtSessionStore{
			sessionRow:        db.GetCbtExamSessionRow{ID: sessionID, PackageID: packageID, Status: db.CbtSessionStatusEnumDraft},
			packageRows:       []db.ListCbtPackagesRow{{ID: packageID}},
			packageQualityRow: db.GetCbtPackageQuestionQualityRow{IsActive: true, TotalQuestions: 10, PublishedQuestions: 10},
		}
		svc := &CbtSession{q: store}

		row, err := svc.UpdateStatus(ctx, sessionID, db.CbtSessionStatusEnumScheduled)
		if err != nil {
			t.Fatalf("UpdateStatus(scheduled) error = %v, want nil", err)
		}
		if row.Status != db.CbtSessionStatusEnumScheduled || store.getID != sessionID || store.packageQualityID != packageID || store.updateStatusArg.ID != sessionID || store.updateStatusArg.Status != db.CbtSessionStatusEnumScheduled {
			t.Fatalf("UpdateStatus(scheduled) row=%+v get=%v quality=%v update=%+v", row, store.getID, store.packageQualityID, store.updateStatusArg)
		}
	})

	t.Run("inactive package blocks scheduling", func(t *testing.T) {
		store := &fakeCbtSessionStore{
			sessionRow:        db.GetCbtExamSessionRow{ID: sessionID, PackageID: packageID, Status: db.CbtSessionStatusEnumDraft},
			packageRows:       []db.ListCbtPackagesRow{{ID: packageID}},
			packageQualityRow: db.GetCbtPackageQuestionQualityRow{IsActive: false, TotalQuestions: 10, PublishedQuestions: 10},
		}
		svc := &CbtSession{q: store}

		_, err := svc.UpdateStatus(ctx, sessionID, db.CbtSessionStatusEnumScheduled)
		if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "paket soal tidak aktif") {
			t.Fatalf("UpdateStatus(inactive package) error = %v, want package conflict", err)
		}
		if store.updateStatusArg.ID.Valid {
			t.Fatalf("UpdateStatus(inactive package) update=%+v, want skipped", store.updateStatusArg)
		}
	})

	t.Run("active rejects ended schedule before readiness", func(t *testing.T) {
		store := &fakeCbtSessionStore{
			sessionRow: db.GetCbtExamSessionRow{
				ID:           sessionID,
				PackageID:    packageID,
				Status:       db.CbtSessionStatusEnumScheduled,
				ScheduledEnd: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
			},
			packageRows:       []db.ListCbtPackagesRow{{ID: packageID}},
			packageQualityRow: db.GetCbtPackageQuestionQualityRow{IsActive: true, TotalQuestions: 10, PublishedQuestions: 10},
		}
		svc := &CbtSession{q: store}

		_, err := svc.UpdateStatus(ctx, sessionID, db.CbtSessionStatusEnumActive)
		if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "jadwal sesi sudah berakhir") {
			t.Fatalf("UpdateStatus(active ended) error = %v, want ended schedule conflict", err)
		}
		if store.roomReadinessID.Valid || store.updateStatusArg.ID.Valid {
			t.Fatalf("UpdateStatus(active ended) readiness=%v update=%+v, want skipped", store.roomReadinessID, store.updateStatusArg)
		}
	})

	t.Run("active validates room readiness before status update", func(t *testing.T) {
		store := &fakeCbtSessionStore{
			sessionRow: db.GetCbtExamSessionRow{
				ID:           sessionID,
				PackageID:    packageID,
				Status:       db.CbtSessionStatusEnumScheduled,
				ScheduledEnd: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
			},
			packageRows:       []db.ListCbtPackagesRow{{ID: packageID}},
			packageQualityRow: db.GetCbtPackageQuestionQualityRow{IsActive: true, TotalQuestions: 10, PublishedQuestions: 10},
			roomReadinessRow:  db.GetCbtSessionRoomReadinessRow{ParticipantCount: 2, RoomCount: 0, TotalCapacity: 0},
		}
		svc := &CbtSession{q: store}

		_, err := svc.UpdateStatus(ctx, sessionID, db.CbtSessionStatusEnumActive)
		if !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "sesi belum memiliki ruangan ujian") {
			t.Fatalf("UpdateStatus(active not ready) error = %v, want readiness conflict", err)
		}
		if store.roomReadinessID != sessionID || store.updateStatusArg.ID.Valid {
			t.Fatalf("UpdateStatus(active not ready) readiness=%v update=%+v, want readiness only", store.roomReadinessID, store.updateStatusArg)
		}
	})
}
