package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeExamTelemetryStore struct {
	fakeExamStore

	state         db.GetCbtParticipantRiskForUpdateRow
	stateErr      error
	dedupRow      db.GetRecentCbtProctorEventByDedupKeyRow
	dedupErr      error
	dedupArg      db.GetRecentCbtProctorEventByDedupKeyParams
	riskArg       db.UpdateCbtParticipantProctorRiskParams
	riskErr       error
	proctorEvents []db.CreateCbtParticipantProctorEventParams
	proctorErr    error
}

func (f *fakeExamTelemetryStore) GetCbtParticipantRiskForUpdate(ctx context.Context, id pgtype.UUID) (db.GetCbtParticipantRiskForUpdateRow, error) {
	if f.stateErr != nil {
		return db.GetCbtParticipantRiskForUpdateRow{}, f.stateErr
	}
	f.state.ID = id
	return f.state, nil
}

func (f *fakeExamTelemetryStore) GetRecentCbtProctorEventByDedupKey(ctx context.Context, arg db.GetRecentCbtProctorEventByDedupKeyParams) (db.GetRecentCbtProctorEventByDedupKeyRow, error) {
	f.dedupArg = arg
	return f.dedupRow, f.dedupErr
}

func (f *fakeExamTelemetryStore) UpdateCbtParticipantProctorRisk(ctx context.Context, arg db.UpdateCbtParticipantProctorRiskParams) (db.UpdateCbtParticipantProctorRiskRow, error) {
	f.riskArg = arg
	return db.UpdateCbtParticipantProctorRiskRow{
		ID:                 arg.ID,
		ViolationCount:     arg.ViolationCount,
		RiskScore:          arg.RiskScore,
		RiskLevel:          arg.RiskLevel,
		LockedAt:           arg.LockedAt,
		LockedReason:       arg.LockedReason,
		LastLocalSaveAt:    arg.LastLocalSaveAt,
		LastSyncedAt:       arg.LastSyncedAt,
		PendingAnswerCount: arg.PendingAnswerCount,
		SyncState:          arg.SyncState,
	}, f.riskErr
}

func (f *fakeExamTelemetryStore) CreateCbtParticipantProctorEvent(ctx context.Context, arg db.CreateCbtParticipantProctorEventParams) (db.CreateCbtParticipantProctorEventRow, error) {
	f.proctorEvents = append(f.proctorEvents, arg)
	return db.CreateCbtParticipantProctorEventRow{
		ID:            arg.ParticipantID,
		ParticipantID: arg.ParticipantID,
		EventType:     arg.EventType,
		EventData:     arg.EventData,
		Severity:      arg.Severity,
		Category:      arg.Category,
		RiskDelta:     arg.RiskDelta,
		DedupKey:      arg.DedupKey,
		RequiresNote:  arg.RequiresNote,
	}, f.proctorErr
}

func TestExamRecordClientEventWithTelemetryStoreEscalatesAndSanitizes(t *testing.T) {
	ctx := context.Background()
	participantID := mustUUID(t, "10101010-1010-1010-1010-101010101010")
	sessionID := mustUUID(t, "20202020-2020-2020-2020-202020202020")
	roomID := mustUUID(t, "30303030-3030-3030-3030-303030303030")
	eventAt := "2026-05-01T08:30:00Z"
	store := &fakeExamTelemetryStore{
		state: db.GetCbtParticipantRiskForUpdateRow{
			SessionID:       sessionID,
			RoomID:          roomID,
			ViolationCount:  1,
			RiskScore:       40,
			RiskLevel:       "warning",
			SyncState:       "synced",
			LastLocalSaveAt: pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC), Valid: true},
		},
		dedupErr: pgx.ErrNoRows,
	}
	svc := &Exam{q: store}

	err := svc.RecordClientEvent(ctx, participantID, "DEVICE-MISMATCH-STRONG", map[string]any{
		"severity":             "info",
		"risk_delta":           0,
		"correlation_id":       "device-check-1",
		"original_event_at":    eventAt,
		"last_synced_at":       "2026-05-01T08:29:45Z",
		"pending_answer_count": 2,
		"sync_state":           "pending",
	})
	if err != nil {
		t.Fatalf("RecordClientEvent() error = %v", err)
	}

	if store.dedupArg.DedupKey != pgUUIDString(participantID)+":device_mismatch_strong:device-check-1" {
		t.Fatalf("dedup key = %q", store.dedupArg.DedupKey)
	}
	if store.riskArg.RiskScore != 100 || store.riskArg.ViolationCount != 2 || store.riskArg.RiskLevel != "locked" || !store.riskArg.LockedAt.Valid {
		t.Fatalf("risk update = %+v, want capped score 100, +1 violation and locked", store.riskArg)
	}
	if store.riskArg.LockedReason.String != "device_mismatch_strong" || !store.riskArg.SuspiciousFlag {
		t.Fatalf("lock reason/suspicious = %q/%v", store.riskArg.LockedReason.String, store.riskArg.SuspiciousFlag)
	}
	if store.riskArg.PendingAnswerCount != 2 || store.riskArg.SyncState != "pending" || !store.riskArg.LastSyncedAt.Valid {
		t.Fatalf("sync risk fields = count %d state %q lastSynced %+v", store.riskArg.PendingAnswerCount, store.riskArg.SyncState, store.riskArg.LastSyncedAt)
	}
	if len(store.proctorEvents) != 1 {
		t.Fatalf("proctor events len = %d, want 1", len(store.proctorEvents))
	}
	event := store.proctorEvents[0]
	if event.EventType != "device_mismatch_strong" || event.Severity != "critical" || event.RiskDelta != 60 || !event.RequiresNote {
		t.Fatalf("event classification = %+v", event)
	}
	if event.OriginalEventAt.Time.Format(time.RFC3339) != eventAt || event.CorrelationID != "device-check-1" {
		t.Fatalf("event timing/correlation = %s/%q", event.OriginalEventAt.Time.Format(time.RFC3339), event.CorrelationID)
	}
	var payload map[string]any
	if err := json.Unmarshal(event.EventData, &payload); err != nil {
		t.Fatalf("event payload unmarshal = %v", err)
	}
	if payload["risk_delta"].(float64) != 60 || payload["risk_score_after"].(float64) != 100 || payload["deduped"].(bool) {
		t.Fatalf("event payload = %v", payload)
	}
}

func TestExamRecordClientEventTelemetryDedupAndTechnicalDoNotRaiseRisk(t *testing.T) {
	ctx := context.Background()
	participantID := mustUUID(t, "40404040-4040-4040-4040-404040404040")

	store := &fakeExamTelemetryStore{
		state: db.GetCbtParticipantRiskForUpdateRow{RiskScore: 10, RiskLevel: "normal", SyncState: "synced"},
		// nil dedup error means an event already exists in the dedup window.
	}
	svc := &Exam{q: store}
	if err := svc.RecordClientEvent(ctx, participantID, "app_switch", map[string]any{"count": 3}); err != nil {
		t.Fatalf("RecordClientEvent(deduped app switch) error = %v", err)
	}
	if store.riskArg.RiskScore != 10 || store.riskArg.ViolationCount != 0 || store.riskArg.AppSwitchIncrement != 0 {
		t.Fatalf("dedup risk update = %+v, want unchanged risk and no counter increment", store.riskArg)
	}
	if len(store.proctorEvents) != 1 || store.proctorEvents[0].RiskDelta != 0 {
		t.Fatalf("dedup event = %+v, want zero risk delta", store.proctorEvents)
	}
	var payload map[string]any
	if err := json.Unmarshal(store.proctorEvents[0].EventData, &payload); err != nil {
		t.Fatalf("dedup payload unmarshal = %v", err)
	}
	if !payload["deduped"].(bool) || payload["risk_delta"].(float64) != 25 {
		t.Fatalf("dedup payload = %v, want client-visible decision and deduped=true", payload)
	}

	store = &fakeExamTelemetryStore{
		state:    db.GetCbtParticipantRiskForUpdateRow{RiskScore: 75, ViolationCount: 2, RiskLevel: "high", SyncState: "failed", PendingAnswerCount: 1},
		dedupErr: pgx.ErrNoRows,
	}
	svc = &Exam{q: store}
	if err := svc.RecordClientEvent(ctx, participantID, "offline_mass", map[string]any{"pending_answer_count": 7, "sync_state": "pending"}); err != nil {
		t.Fatalf("RecordClientEvent(technical) error = %v", err)
	}
	if store.riskArg.RiskScore != 75 || store.riskArg.ViolationCount != 2 || store.riskArg.RiskLevel != "high" || store.riskArg.LockedAt.Valid {
		t.Fatalf("technical risk update = %+v, want unchanged non-locked risk", store.riskArg)
	}
	if store.riskArg.PendingAnswerCount != 7 || store.riskArg.SyncState != "pending" || store.proctorEvents[0].RiskDelta != 0 {
		t.Fatalf("technical sync/event = %+v event %+v", store.riskArg, store.proctorEvents[0])
	}

	store = &fakeExamTelemetryStore{
		state:    db.GetCbtParticipantRiskForUpdateRow{RiskScore: 0, RiskLevel: "normal", SyncState: "synced"},
		dedupErr: pgx.ErrNoRows,
	}
	svc = &Exam{q: store}
	if err := svc.RecordClientEvent(ctx, participantID, "web_focus_lost", map[string]any{"reason": "window_blur"}); err != nil {
		t.Fatalf("RecordClientEvent(web_focus_lost) error = %v", err)
	}
	if store.riskArg.AppSwitchIncrement != 1 || store.riskArg.ScreenshotIncrement != 0 {
		t.Fatalf("web focus lost counters = %+v, want app switch increment only", store.riskArg)
	}
	if store.proctorEvents[0].EventType != "web_focus_lost" {
		t.Fatalf("web focus event = %q, want web_focus_lost", store.proctorEvents[0].EventType)
	}
}

func TestExamRecordClientEventTelemetryPropagatesStoreErrors(t *testing.T) {
	ctx := context.Background()
	participantID := mustUUID(t, "50505050-5050-5050-5050-505050505050")

	stateErr := errors.New("risk state failed")
	svc := &Exam{q: &fakeExamTelemetryStore{stateErr: stateErr}}
	if err := svc.RecordClientEvent(ctx, participantID, "focus_lost_short", nil); !errors.Is(err, stateErr) {
		t.Fatalf("RecordClientEvent(state error) = %v, want %v", err, stateErr)
	}

	dedupErr := errors.New("dedup failed")
	svc = &Exam{q: &fakeExamTelemetryStore{dedupErr: dedupErr}}
	if err := svc.RecordClientEvent(ctx, participantID, "focus_lost_short", nil); !errors.Is(err, dedupErr) {
		t.Fatalf("RecordClientEvent(dedup error) = %v, want %v", err, dedupErr)
	}

	riskErr := errors.New("risk update failed")
	svc = &Exam{q: &fakeExamTelemetryStore{dedupErr: pgx.ErrNoRows, riskErr: riskErr}}
	if err := svc.RecordClientEvent(ctx, participantID, "focus_lost_short", nil); !errors.Is(err, riskErr) {
		t.Fatalf("RecordClientEvent(risk update error) = %v, want %v", err, riskErr)
	}

	proctorErr := errors.New("event insert failed")
	svc = &Exam{q: &fakeExamTelemetryStore{dedupErr: pgx.ErrNoRows, proctorErr: proctorErr}}
	if err := svc.RecordClientEvent(ctx, participantID, "focus_lost_short", nil); !errors.Is(err, proctorErr) {
		t.Fatalf("RecordClientEvent(event insert error) = %v, want %v", err, proctorErr)
	}
}
