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

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestExamSubmitExtraBranches(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)
	locked := participant
	locked.LockedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}

	store := &fakeExamStore{}
	if err := (&Exam{q: store}).Submit(ctx, locked); !errors.Is(err, ErrExamLocked) {
		t.Fatalf("Submit(locked) = %v, want ErrExamLocked", err)
	}
	if store.submitID.Valid || store.correctnessID.Valid || len(store.events) != 0 {
		t.Fatalf("Submit(locked) mutated store submit=%v correctness=%v events=%+v", store.submitID, store.correctnessID, store.events)
	}

	submittedAt := time.Date(2026, 5, 17, 11, 12, 13, 0, time.UTC)
	store = &fakeExamStore{submitRow: db.SubmitParticipantExamRow{ID: participant.ID, SubmittedAt: pgtype.Timestamptz{Time: submittedAt, Valid: true}}}
	if err := submitExamWithStore(ctx, store, participant.ID); err != nil {
		t.Fatalf("submitExamWithStore() error = %v", err)
	}
	if store.submitID != participant.ID || store.correctnessID != participant.ID || len(store.events) != 1 || store.events[0].EventType != "submit" {
		t.Fatalf("submitExamWithStore store = submit %v correctness %v events %+v, want submit flow", store.submitID, store.correctnessID, store.events)
	}
	var payload map[string]string
	if err := json.Unmarshal(store.events[0].EventData, &payload); err != nil {
		t.Fatalf("submit event JSON = %v; data=%s", err, string(store.events[0].EventData))
	}
	if payload["submitted_at"] != submittedAt.Format(time.RFC3339) {
		t.Fatalf("submit event payload = %+v, want submitted_at", payload)
	}
}

func TestExamRecordClientEventProctorStoreDedupAndTechnicalBranches(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)
	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)

	dedupStore := &fakeExamProctorStore{
		riskRow: db.GetCbtParticipantRiskForUpdateRow{
			ID:             participant.ID,
			SessionID:      participant.SessionID,
			RoomID:         participant.RoomID,
			ViolationCount: 1,
			RiskScore:      30,
			RiskLevel:      "warning",
			SyncState:      "synced",
		},
		recentRow: db.GetRecentCbtProctorEventByDedupKeyRow{ID: mustUUID(t, "81000000-0000-0000-0000-000000000001")},
	}
	if err := recordProctorTelemetryWithStore(ctx, dedupStore, participant.ID, "app_switch", map[string]any{"count": 3, "correlation_id": "episode-1", "event_at": now.Format(time.RFC3339)}, now.Add(time.Minute)); err != nil {
		t.Fatalf("recordProctorTelemetryWithStore(dedup) error = %v", err)
	}
	if dedupStore.recentArg.ParticipantID != participant.ID || !strings.Contains(dedupStore.recentArg.DedupKey, "episode-1") {
		t.Fatalf("dedup lookup arg = %+v, want participant and correlation key", dedupStore.recentArg)
	}
	if dedupStore.updateArg.RiskScore != 30 || dedupStore.updateArg.ViolationCount != 1 || dedupStore.updateArg.AppSwitchIncrement != 0 || dedupStore.updateArg.SyncState != "synced" {
		t.Fatalf("dedup update arg = %+v, want unchanged risk/counters", dedupStore.updateArg)
	}
	if dedupStore.createArg.EventType != "app_switch_repeated" || dedupStore.createArg.RiskDelta != 0 || dedupStore.createArg.OriginalEventAt.Time.UTC() != now {
		t.Fatalf("dedup event arg = %+v, want repeated event with zero risk and original timestamp", dedupStore.createArg)
	}
	if !strings.Contains(string(dedupStore.createArg.EventData), `"deduped":true`) {
		t.Fatalf("dedup event data = %s, want deduped true", string(dedupStore.createArg.EventData))
	}

	technicalStore := &fakeExamProctorStore{
		riskRow: db.GetCbtParticipantRiskForUpdateRow{
			ID:                 participant.ID,
			ViolationCount:     2,
			RiskScore:          55,
			RiskLevel:          "high",
			PendingAnswerCount: 4,
			SyncState:          "failed",
		},
		recentErr: pgx.ErrNoRows,
	}
	if err := recordProctorTelemetryWithStore(ctx, technicalStore, participant.ID, "pending_sync", map[string]any{"pending_answer_count": -1, "sync_state": "unknown", "last_synced_at": "not-time"}, now); err != nil {
		t.Fatalf("recordProctorTelemetryWithStore(technical) error = %v", err)
	}
	if technicalStore.updateArg.RiskScore != 55 || technicalStore.updateArg.ViolationCount != 2 || technicalStore.updateArg.RiskLevel != "high" || technicalStore.updateArg.SuspiciousFlag {
		t.Fatalf("technical update arg = %+v, want unchanged risk and no suspicious flag", technicalStore.updateArg)
	}
	if technicalStore.updateArg.PendingAnswerCount != 4 || technicalStore.updateArg.SyncState != "pending" {
		t.Fatalf("technical sync arg = pending_count %d sync %q, want fallback count and pending", technicalStore.updateArg.PendingAnswerCount, technicalStore.updateArg.SyncState)
	}
	if technicalStore.createArg.Severity != string(ProctorSeverityTechnical) || technicalStore.createArg.RiskDelta != 0 {
		t.Fatalf("technical create arg = %+v, want technical event with zero risk", technicalStore.createArg)
	}
}

func TestExamRecordClientEventProctorStoreErrorBranches(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)

	stateErr := errors.New("risk state failed")
	if err := recordProctorTelemetryWithStore(ctx, &fakeExamProctorStore{riskErr: stateErr}, participant.ID, "focus_lost_short", nil, time.Now()); !errors.Is(err, stateErr) {
		t.Fatalf("recordProctorTelemetryWithStore(risk error) = %v, want %v", err, stateErr)
	}

	recentErr := errors.New("dedup lookup failed")
	if err := recordProctorTelemetryWithStore(ctx, &fakeExamProctorStore{recentErr: recentErr}, participant.ID, "focus_lost_short", nil, time.Now()); !errors.Is(err, recentErr) {
		t.Fatalf("recordProctorTelemetryWithStore(dedup error) = %v, want %v", err, recentErr)
	}

	updateErr := errors.New("risk update failed")
	store := &fakeExamProctorStore{recentErr: pgx.ErrNoRows, updateErr: updateErr}
	if err := recordProctorTelemetryWithStore(ctx, store, participant.ID, "focus_lost_short", nil, time.Now()); !errors.Is(err, updateErr) || store.createCalls != 0 {
		t.Fatalf("recordProctorTelemetryWithStore(update error) err/createCalls = %v/%d, want update error before event", err, store.createCalls)
	}

	createErr := errors.New("create event failed")
	store = &fakeExamProctorStore{recentErr: pgx.ErrNoRows, createErr: createErr}
	if err := recordProctorTelemetryWithStore(ctx, store, participant.ID, "focus_lost_short", nil, time.Now()); !errors.Is(err, createErr) || store.updateCalls != 1 {
		t.Fatalf("recordProctorTelemetryWithStore(create error) err/updateCalls = %v/%d, want create error after update", err, store.updateCalls)
	}
}

func TestExamNormalizeSyncStateAndAntiCheatReasonExtraBranches(t *testing.T) {
	syncTests := []struct {
		raw     string
		pending int32
		fb      string
		want    string
	}{
		{raw: " sinkron ", want: "synced"},
		{raw: "belum_sinkron", want: "pending"},
		{raw: "tertahan", want: "failed"},
		{raw: "", fb: "failed", want: "failed"},
		{raw: "unknown", pending: 2, fb: "synced", want: "pending"},
		{raw: "garbage", pending: 1, want: "pending"},
		{raw: "garbage", want: "unknown"},
	}
	for _, tt := range syncTests {
		if got := normalizeSyncState(tt.raw, tt.pending, tt.fb); got != tt.want {
			t.Fatalf("normalizeSyncState(%q,%d,%q) = %q, want %q", tt.raw, tt.pending, tt.fb, got, tt.want)
		}
	}

	if got := antiCheatReason(nil); got != "anti_cheat_violation" {
		t.Fatalf("antiCheatReason(nil) = %q, want default", got)
	}
	if got := antiCheatReason(map[string]any{"reason": "  split_screen_detected  "}); got != "split_screen_detected" {
		t.Fatalf("antiCheatReason(trimmed) = %q, want split_screen_detected", got)
	}
	if got := antiCheatReason(map[string]any{"reason": 123}); got != "anti_cheat_violation" {
		t.Fatalf("antiCheatReason(non-string) = %q, want default", got)
	}
	if got := antiCheatRiskWeight(map[string]any{"reason": "app_backgrounded"}); got != 30 {
		t.Fatalf("antiCheatRiskWeight(app_backgrounded) = %d, want 30", got)
	}
}

type fakeExamProctorStore struct {
	riskRow db.GetCbtParticipantRiskForUpdateRow
	riskErr error

	recentArg db.GetRecentCbtProctorEventByDedupKeyParams
	recentRow db.GetRecentCbtProctorEventByDedupKeyRow
	recentErr error

	updateArg   db.UpdateCbtParticipantProctorRiskParams
	updateRow   db.UpdateCbtParticipantProctorRiskRow
	updateErr   error
	updateCalls int

	createArg   db.CreateCbtParticipantProctorEventParams
	createRow   db.CreateCbtParticipantProctorEventRow
	createErr   error
	createCalls int
}

func (f *fakeExamProctorStore) GetCbtParticipantRiskForUpdate(context.Context, pgtype.UUID) (db.GetCbtParticipantRiskForUpdateRow, error) {
	return f.riskRow, f.riskErr
}

func (f *fakeExamProctorStore) GetRecentCbtProctorEventByDedupKey(_ context.Context, arg db.GetRecentCbtProctorEventByDedupKeyParams) (db.GetRecentCbtProctorEventByDedupKeyRow, error) {
	f.recentArg = arg
	return f.recentRow, f.recentErr
}

func (f *fakeExamProctorStore) UpdateCbtParticipantProctorRisk(_ context.Context, arg db.UpdateCbtParticipantProctorRiskParams) (db.UpdateCbtParticipantProctorRiskRow, error) {
	f.updateArg = arg
	f.updateCalls++
	return f.updateRow, f.updateErr
}

func (f *fakeExamProctorStore) CreateCbtParticipantProctorEvent(_ context.Context, arg db.CreateCbtParticipantProctorEventParams) (db.CreateCbtParticipantProctorEventRow, error) {
	f.createArg = arg
	f.createCalls++
	return f.createRow, f.createErr
}
