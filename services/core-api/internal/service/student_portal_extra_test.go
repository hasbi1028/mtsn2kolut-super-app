package service

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestNewStudentPortalSetsStore(t *testing.T) {
	store := &fakeStudentPortalStore{}
	svc := NewStudentPortal(nil)
	if svc == nil {
		t.Fatalf("NewStudentPortal(nil) = nil")
	}
	if svc.q == nil {
		t.Fatalf("NewStudentPortal(nil).q = nil, want queries store")
	}

	svc = &StudentPortal{q: store}
	if svc.q != store {
		t.Fatalf("StudentPortal store = %T, want injected fake store", svc.q)
	}
}

func TestStudentPortalCbtRevealExpiresAtPrefersEndThenStartThenNow(t *testing.T) {
	now := time.Now()
	start := now.Add(5 * time.Minute).Truncate(time.Second)
	end := now.Add(45 * time.Minute).Truncate(time.Second)

	withEnd := db.GetStudentPortalCbtParticipantRow{
		ScheduledStart: pgtype.Timestamptz{Time: start, Valid: true},
		ScheduledEnd:   pgtype.Timestamptz{Time: end, Valid: true},
	}
	if got := studentPortalCbtRevealExpiresAt(withEnd); !got.Equal(end) {
		t.Fatalf("expires with scheduled end = %s, want %s", got, end)
	}

	withStartOnly := db.GetStudentPortalCbtParticipantRow{
		ScheduledStart: pgtype.Timestamptz{Time: start, Valid: true},
	}
	wantFromStart := start.Add(studentPortalTokenWindowMinutes * time.Minute)
	if got := studentPortalCbtRevealExpiresAt(withStartOnly); !got.Equal(wantFromStart) {
		t.Fatalf("expires with scheduled start only = %s, want %s", got, wantFromStart)
	}

	beforeFallback := time.Now()
	fallback := studentPortalCbtRevealExpiresAt(db.GetStudentPortalCbtParticipantRow{})
	afterFallback := time.Now()
	if fallback.Before(beforeFallback.Add(studentPortalTokenWindowMinutes*time.Minute)) || fallback.After(afterFallback.Add(studentPortalTokenWindowMinutes*time.Minute+time.Second)) {
		t.Fatalf("fallback expires = %s, want about now + %d minutes", fallback, studentPortalTokenWindowMinutes)
	}
}

func TestStudentPortalCbtScheduleItemExtraStatusPaths(t *testing.T) {
	now := time.Now()
	base := db.ListStudentPortalCbtScheduleRow{
		ParticipantID:   testGenerationUUID(71),
		SessionID:       testGenerationUUID(72),
		SessionStatus:   db.CbtSessionStatusEnumDraft,
		ScheduledStart:  pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
		ScheduledEnd:    pgtype.Timestamptz{Time: now.Add(2 * time.Hour), Valid: true},
		Token:           "",
		SessionTitle:    "",
		PackageTitle:    "",
		DurationMinutes: 90,
	}

	item := studentPortalCbtScheduleItem(base, now)
	if item.Status != StudentPortalCbtUpcoming || item.SessionTitle != "Sesi Ujian" || item.TokenMasked != nil || item.CanRevealToken || item.RequiresRoomToken {
		t.Fatalf("draft upcoming item = %+v, want default title without reveal/token", item)
	}

	locked := base
	locked.LockedAt = pgtype.Timestamptz{Time: now, Valid: true}
	locked.SessionStatus = db.CbtSessionStatusEnumActive
	locked.RoomID = testGenerationUUID(73)
	locked.RoomToken = "ROOM"
	if item := studentPortalCbtScheduleItem(locked, now); item.Status != StudentPortalCbtLocked || item.CanRevealToken {
		t.Fatalf("locked item = %+v, want locked without reveal", item)
	}

	finished := base
	finished.SessionStatus = db.CbtSessionStatusEnumFinished
	if item := studentPortalCbtScheduleItem(finished, now); item.Status != StudentPortalCbtClosed {
		t.Fatalf("finished item status = %q, want closed", item.Status)
	}
}
