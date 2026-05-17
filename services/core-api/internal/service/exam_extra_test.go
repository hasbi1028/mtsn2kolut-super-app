package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestExamExtraSubmitBranchesAndEventPayloads(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)
	locked := participant
	locked.LockedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}

	svc := &Exam{q: &fakeExamStore{}}
	if err := svc.Submit(ctx, locked); !errors.Is(err, ErrExamLocked) {
		t.Fatalf("Submit(locked) = %v, want ErrExamLocked", err)
	}

	submittedAt := time.Date(2026, 5, 17, 10, 30, 0, 0, time.UTC)
	store := &fakeExamStore{submitRow: db.SubmitParticipantExamRow{ID: participant.ID, SubmittedAt: pgtype.Timestamptz{Time: submittedAt, Valid: true}}}
	svc = &Exam{q: store}
	if err := svc.Submit(ctx, participant); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if store.submitID != participant.ID || store.correctnessID != participant.ID {
		t.Fatalf("Submit store ids = submit:%v correctness:%v, want participant", store.submitID, store.correctnessID)
	}
	if len(store.events) != 1 || store.events[0].EventType != "submit" {
		t.Fatalf("Submit events = %+v, want one submit event", store.events)
	}
	var payload map[string]string
	if err := json.Unmarshal(store.events[0].EventData, &payload); err != nil {
		t.Fatalf("submit event json.Unmarshal error = %v; data=%s", err, string(store.events[0].EventData))
	}
	if payload["submitted_at"] != submittedAt.Format(time.RFC3339) {
		t.Fatalf("submit event payload = %+v, want submitted_at timestamp", payload)
	}
}

func TestExamExtraRecordClientEventFallbackValidationAndSanitization(t *testing.T) {
	ctx := context.Background()
	participant := examActiveParticipant(t)

	store := &fakeExamStore{}
	svc := &Exam{q: store}
	if err := svc.RecordClientEvent(ctx, participant.ID, "not-a-real-event", nil); !errors.Is(err, ErrExamInvalidTelemetry) {
		t.Fatalf("RecordClientEvent(invalid) = %v, want ErrExamInvalidTelemetry", err)
	}
	if len(store.events) != 0 || store.appSwitchID.Valid || store.screenshotID.Valid {
		t.Fatalf("invalid telemetry mutated store: events=%+v app=%v screenshot=%v", store.events, store.appSwitchID, store.screenshotID)
	}

	store = &fakeExamStore{}
	svc = &Exam{q: store}
	if err := svc.RecordClientEvent(ctx, participant.ID, "focus_lost_short", map[string]any{
		"reason":       "window_focus_lost",
		"severity":     "critical-spoof",
		"risk_delta":   999,
		"risk_score":   999,
		"risk_level":   "locked",
		"locked_at":    "2026-05-17T10:00:00Z",
		"custom_field": "kept",
	}); err != nil {
		t.Fatalf("RecordClientEvent(fallback focus_lost_short) error = %v", err)
	}
	if store.antiCheatViolationArg.ID != participant.ID || store.antiCheatViolationArg.RiskScore == 0 || store.antiCheatViolationArg.LockedReason.String != "focus_lost_short" {
		t.Fatalf("anti cheat increment arg = %+v, want participant risk increment for normalized event", store.antiCheatViolationArg)
	}
	if len(store.events) != 1 || store.events[0].EventType != "focus_lost_short" {
		t.Fatalf("events = %+v, want normalized focus_lost_short event", store.events)
	}
	body := string(store.events[0].EventData)
	for _, forbidden := range []string{"critical-spoof", `"risk_delta":999`, `"risk_score":999`, `"risk_level":"locked"`, `"locked_at"`} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("event data = %s, contains spoofed/sensitive field %s", body, forbidden)
		}
	}
	if !strings.Contains(body, `"custom_field":"kept"`) || !strings.Contains(body, `"event_type":"focus_lost_short"`) {
		t.Fatalf("event data = %s, want sanitized custom payload plus decision metadata", body)
	}
}

func TestExamExtraOptionalDecodeHelpers(t *testing.T) {
	valid, ok := optionalProctorTimestamp(map[string]any{"event_at": "2026-05-17T10:15:30.123Z"}, "event_at")
	if !ok || !valid.Valid || valid.Time.UTC().Format(time.RFC3339Nano) != "2026-05-17T10:15:30.123Z" {
		t.Fatalf("optionalProctorTimestamp(valid) = %+v/%v, want parsed timestamp", valid, ok)
	}
	if got, ok := optionalProctorTimestamp(map[string]any{"event_at": "not-time"}, "event_at"); ok || got.Valid {
		t.Fatalf("optionalProctorTimestamp(invalid) = %+v/%v, want empty false", got, ok)
	}
	if got := parseOptionOrder([]byte(`{"question-1":["B","A"]}`)); len(got["question-1"]) != 2 || got["question-1"][0] != "B" {
		t.Fatalf("parseOptionOrder(valid) = %+v, want decoded labels", got)
	}
	if got := parseOptionOrder([]byte(`not-json`)); len(got) != 0 {
		t.Fatalf("parseOptionOrder(invalid) = %+v, want empty map", got)
	}
}
