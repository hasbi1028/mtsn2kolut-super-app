package handler

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/service"
)

func TestCbtSessionProctorPureHelpers(t *testing.T) {
	if got := firstNonEmptyString("", " \t", "  keep spaces  ", "fallback"); got != "  keep spaces  " {
		t.Fatalf("firstNonEmptyString returned %q, want original first non-empty value", got)
	}
	if got := firstNonEmptyString("", "   "); got != "" {
		t.Fatalf("firstNonEmptyString(empty values) = %q, want empty", got)
	}

	validTime := pgtype.Timestamptz{Time: time.Date(2026, 5, 17, 1, 2, 3, 0, time.UTC), Valid: true}
	invalidTime := pgtype.Timestamptz{Valid: false}
	cases := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "nil", value: nil, want: false},
		{name: "valid timestamptz", value: validTime, want: true},
		{name: "invalid timestamptz", value: invalidTime, want: false},
		{name: "valid timestamptz pointer", value: &validTime, want: true},
		{name: "nil timestamptz pointer", value: (*pgtype.Timestamptz)(nil), want: false},
		{name: "blank string", value: " \n", want: false},
		{name: "nonblank string", value: "2026-05-17T00:00:00Z", want: true},
		{name: "other nonnil type", value: 1, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := proctorAnyTimeValid(tc.value); got != tc.want {
				t.Fatalf("proctorAnyTimeValid(%#v) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestDecodeOptionalProctorActionBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = nil
	body, ok := decodeOptionalProctorActionBody(httptest.NewRecorder(), req)
	if !ok || body != (proctorActionBody{}) {
		t.Fatalf("nil body decode = (%#v, %v), want zero body and ok", body, ok)
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"action_type":"force_submit","action":"mark","reason":" device issue ","notes":"checked","event_id":"evt-1"}`))
	body, ok = decodeOptionalProctorActionBody(httptest.NewRecorder(), req)
	if !ok {
		t.Fatal("valid JSON decode returned ok=false")
	}
	want := proctorActionBody{ActionType: "force_submit", Action: "mark", Reason: " device issue ", Notes: "checked", EventID: "evt-1"}
	if body != want {
		t.Fatalf("valid JSON decode = %#v, want %#v", body, want)
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	body, ok = decodeOptionalProctorActionBody(httptest.NewRecorder(), req)
	if !ok || body != (proctorActionBody{}) {
		t.Fatalf("empty body decode = (%#v, %v), want zero body and ok", body, ok)
	}

	recorder := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"action_type"`))
	if _, ok := decodeOptionalProctorActionBody(recorder, req); ok {
		t.Fatal("malformed JSON decode returned ok=true")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("malformed JSON status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestBuildProctoringReportDataClassifiesEventsAndParticipants(t *testing.T) {
	summary := service.CbtProctoringLiveSummary{
		SessionID: "session-1",
		RoomID:    "room-1",
		LatestEvents: []service.CbtProctoringEventDTO{
			{ID: "info-ack", Severity: "info", AcknowledgedAt: "2026-05-17T00:00:00Z"},
			{ID: "warn", Severity: "warning"},
			{ID: "medium", Severity: "medium"},
			{ID: "critical", Severity: "critical", AcknowledgedAt: "2026-05-17T00:00:00Z"},
			{ID: "technical", Severity: "technical"},
			{ID: "mass-tech", Severity: "warning", IsMassTechnicalIssue: true},
			{ID: "unknown", Severity: "custom"},
		},
		Participants: []service.CbtProctoringParticipantDTO{
			{ParticipantID: "submitted-by-status", ConnectionStatus: "selesai"},
			{ParticipantID: "submitted-by-time", SubmittedAt: "2026-05-17T00:00:00Z"},
			{ParticipantID: "locked-by-time", LockedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
			{ParticipantID: "locked-by-risk", RiskLevel: "locked"},
			{ParticipantID: "pending-sync-status", SyncStatus: "belum_sinkron"},
			{ParticipantID: "held-sync-status", SyncStatus: "tertahan"},
			{ParticipantID: "pending-answers", PendingAnswerCount: 2},
			{ParticipantID: "normal"},
		},
	}

	report := buildProctoringReportData(summary)
	if report["session_id"] != "session-1" || report["room_id"] != "room-1" {
		t.Fatalf("report identifiers = session:%v room:%v, want session-1/room-1", report["session_id"], report["room_id"])
	}
	if generated, ok := report["generated_at"].(string); !ok || strings.TrimSpace(generated) == "" {
		t.Fatalf("generated_at = %#v, want non-empty string", report["generated_at"])
	}

	severity, ok := report["severity_summary"].(map[string]int)
	if !ok {
		t.Fatalf("severity_summary type = %T, want map[string]int", report["severity_summary"])
	}
	wantSeverity := map[string]int{"info": 1, "warning": 2, "medium": 1, "critical": 1, "technical": 1}
	if !reflect.DeepEqual(severity, wantSeverity) {
		t.Fatalf("severity_summary = %#v, want %#v", severity, wantSeverity)
	}

	if got := report["technical_incidents"].([]service.CbtProctoringEventDTO); len(got) != 2 || got[0].ID != "technical" || got[1].ID != "mass-tech" {
		t.Fatalf("technical_incidents = %#v, want technical and mass-tech", got)
	}
	if got := report["unacknowledged_events"].([]service.CbtProctoringEventDTO); len(got) != 2 || got[0].ID != "medium" || got[1].ID != "technical" {
		t.Fatalf("unacknowledged_events = %#v, want unacked medium and technical only", got)
	}
	if got := report["locked_participants"].([]service.CbtProctoringParticipantDTO); len(got) != 2 || got[0].ParticipantID != "locked-by-time" || got[1].ParticipantID != "locked-by-risk" {
		t.Fatalf("locked_participants = %#v, want locked-by-time and locked-by-risk", got)
	}
	if got := report["sync_anomalies"].([]service.CbtProctoringParticipantDTO); len(got) != 3 {
		t.Fatalf("sync_anomalies length = %d, want 3: %#v", len(got), got)
	}
	submitted, ok := report["submitted_summary"].(map[string]int)
	if !ok {
		t.Fatalf("submitted_summary type = %T, want map[string]int", report["submitted_summary"])
	}
	if submitted["submitted"] != 2 || submitted["not_submitted"] != 6 {
		t.Fatalf("submitted_summary = %#v, want submitted=2 not_submitted=6", submitted)
	}
}

func TestProctorEventNeedsReportAction(t *testing.T) {
	for _, severity := range []string{"medium", "critical", "technical"} {
		if !proctorEventNeedsReportAction(service.CbtProctoringEventDTO{Severity: severity}) {
			t.Fatalf("proctorEventNeedsReportAction(%q) = false, want true", severity)
		}
	}
	for _, severity := range []string{"", "info", "warning", "custom"} {
		if proctorEventNeedsReportAction(service.CbtProctoringEventDTO{Severity: severity}) {
			t.Fatalf("proctorEventNeedsReportAction(%q) = true, want false", severity)
		}
	}
}
