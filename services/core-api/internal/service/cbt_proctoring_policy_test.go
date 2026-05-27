package service

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCbtProctoringPolicyRiskBoundaries(t *testing.T) {
	cases := []struct {
		score int
		want  string
	}{
		{0, "normal"},
		{19, "normal"},
		{20, "warning"},
		{49, "warning"},
		{50, "high"},
		{79, "high"},
		{80, "locked"},
	}
	for _, tc := range cases {
		if got := RiskLevelFromScore(tc.score, 0, sql.NullTime{}); got != tc.want {
			t.Fatalf("RiskLevelFromScore(%d) = %q, want %q", tc.score, got, tc.want)
		}
	}
	if got := RiskLevelFromScore(10, 0, sql.NullTime{Time: time.Now(), Valid: true}); got != "locked" {
		t.Fatalf("RiskLevelFromScore(locked) = %q, want locked", got)
	}
}

func TestCbtProctoringPolicyClassifiesWhitelistedTelemetry(t *testing.T) {
	cases := []struct {
		raw        string
		data       map[string]any
		wantType   string
		wantLevel  ProctorSeverity
		wantRisk   int32
		wantLocked bool
	}{
		{"focus_lost_short", nil, "focus_lost_short", ProctorSeverityWarning, 5, false},
		{"app_switch", nil, "app_switch_once", ProctorSeverityWarning, 10, false},
		{"app_switch_once", map[string]any{"count": 3}, "app_switch_repeated", ProctorSeverityMedium, 25, true},
		{"split_screen_detected", nil, "split_screen_detected", ProctorSeverityMedium, 30, true},
		{"device_mismatch_strong", nil, "device_mismatch_strong", ProctorSeverityCritical, 60, true},
		{"offline_mass", nil, "offline_mass", ProctorSeverityTechnical, 0, false},
		{"pending_sync", nil, "pending_sync", ProctorSeverityTechnical, 0, false},
		{"submit_blocked_pending_sync", nil, "submit_held_pending_sync", ProctorSeverityTechnical, 0, false},
		{"copy_attempt", nil, "copy_attempt", ProctorSeverityWarning, 10, false},
		{"paste_attempt", nil, "paste_attempt", ProctorSeverityMedium, 20, false},
		{"right_click", nil, "context_menu_attempt", ProctorSeverityWarning, 5, false},
	}
	for _, tc := range cases {
		got := ClassifyProctorSeverity(tc.raw, tc.data)
		if got.EventType != tc.wantType || got.Severity != tc.wantLevel || got.RiskDelta != tc.wantRisk || got.LockEligible != tc.wantLocked {
			t.Fatalf("ClassifyProctorSeverity(%q) = %+v, want type=%s severity=%s risk=%d lock=%v", tc.raw, got, tc.wantType, tc.wantLevel, tc.wantRisk, tc.wantLocked)
		}
	}
}

func TestCbtProctoringPolicyRejectsUnknownAndIgnoresClientSeverity(t *testing.T) {
	if normalized, ok := NormalizeProctorEventType("unknown_event"); ok || normalized != "" {
		t.Fatalf("NormalizeProctorEventType(unknown) = %q/%v, want rejected", normalized, ok)
	}
	got := ClassifyProctorSeverity("offline_mass", map[string]any{"severity": "critical", "risk_delta": 99})
	if got.Severity != ProctorSeverityTechnical || got.RiskDelta != 0 || got.LockEligible {
		t.Fatalf("ClassifyProctorSeverity(forged technical) = %+v, want technical zero-risk", got)
	}
}

func TestCbtProctoringPolicyDedupWindows(t *testing.T) {
	if !ShouldDedup("app_switch") {
		t.Fatalf("ShouldDedup(app_switch) = false, want true")
	}
	if got := DedupWindow("focus_lost_short"); got != 30*time.Second {
		t.Fatalf("DedupWindow(focus) = %s, want 30s", got)
	}
	if got := DedupWindow("offline_mass"); got != 2*time.Minute {
		t.Fatalf("DedupWindow(offline_mass) = %s, want 2m", got)
	}
}

func TestCbtProctoringPolicyTechnicalNeverAutoLocks(t *testing.T) {
	decision := ClassifyProctorSeverity("offline_mass", nil)
	state := ParticipantRiskState{ViolationCount: 20, RiskScore: 100}
	if ShouldAutoLock(decision, state) {
		t.Fatalf("ShouldAutoLock(technical) = true, want false")
	}
}

func TestCbtProctoringPolicyNormalizationAndDataDrivenClassification(t *testing.T) {
	cases := []struct {
		name      string
		raw       string
		data      map[string]any
		wantType  string
		severity  ProctorSeverity
		riskDelta int32
	}{
		{name: "trim lower dash alias", raw: " Window-Focus-Lost ", wantType: "focus_lost_short", severity: ProctorSeverityWarning, riskDelta: 5},
		{name: "background alias", raw: "app background resume", wantType: "background_over_threshold", severity: ProctorSeverityMedium, riskDelta: 30},
		{name: "screenshot platform callback bool", raw: "screen_capture_attempt", data: map[string]any{"valid_platform_callback": true}, wantType: "screenshot_attempt_valid", severity: ProctorSeverityMedium, riskDelta: 25},
		{name: "screenshot platform callback string remains ambiguous", raw: "screenshot_attempt", data: map[string]any{"valid_platform_callback": "true"}, wantType: "screenshot_attempt_ambiguous", severity: ProctorSeverityWarning, riskDelta: 5},
		{name: "app switch json number count", raw: "app_switch_once", data: map[string]any{"count": json.Number("3")}, wantType: "app_switch_repeated", severity: ProctorSeverityMedium, riskDelta: 25},
		{name: "root emulator local lock alias", raw: "anti_cheat_local_lock", wantType: "root_emulator_strong", severity: ProctorSeverityCritical, riskDelta: 70},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			normalized, ok := NormalizeProctorEventType(tc.raw)
			if !ok {
				t.Fatalf("NormalizeProctorEventType(%q) rejected", tc.raw)
			}
			decision := ClassifyProctorSeverity(normalized, tc.data)
			if decision.EventType != tc.wantType || decision.Severity != tc.severity || decision.RiskDelta != tc.riskDelta {
				t.Fatalf("ClassifyProctorSeverity(%q) = %+v, want type=%s severity=%s risk=%d", tc.raw, decision, tc.wantType, tc.severity, tc.riskDelta)
			}
		})
	}
}

func TestCbtProctoringPolicyWhitelistContractIsCompleteAndSafe(t *testing.T) {
	validSeverity := map[ProctorSeverity]bool{
		ProctorSeverityInfo:      true,
		ProctorSeverityWarning:   true,
		ProctorSeverityMedium:    true,
		ProctorSeverityCritical:  true,
		ProctorSeverityTechnical: true,
	}
	validAudioKey := map[string]bool{"none": true, "warning": true, "medium": true, "critical": true, "technical": true}
	for key, decision := range proctorEventWhitelist {
		t.Run(key, func(t *testing.T) {
			if normalized, ok := NormalizeProctorEventType(key); !ok || normalized != decision.EventType {
				t.Fatalf("NormalizeProctorEventType(%q) = %q/%v, want %q/true", key, normalized, ok, decision.EventType)
			}
			classified := ClassifyProctorSeverity(key, nil)
			if classified.EventType != decision.EventType {
				t.Fatalf("ClassifyProctorSeverity(%q).EventType = %q, want %q", key, classified.EventType, decision.EventType)
			}
			if decision.EventType == "" || decision.Category == "" || decision.LabelID == "" || decision.MessageID == "" {
				t.Fatalf("decision has incomplete event/category/label/message contract: %+v", decision)
			}
			if !validSeverity[decision.Severity] {
				t.Fatalf("unsupported severity %q in decision %+v", decision.Severity, decision)
			}
			if !validAudioKey[decision.AudioKey] {
				t.Fatalf("unsupported audio key %q in decision %+v", decision.AudioKey, decision)
			}
			if decision.Severity == ProctorSeverityTechnical && (decision.RiskDelta != 0 || decision.LockEligible) {
				t.Fatalf("technical event can raise risk or lock: %+v", decision)
			}
			if strings.TrimSpace(decision.LabelID) != decision.LabelID || strings.TrimSpace(decision.MessageID) != decision.MessageID {
				t.Fatalf("label/message ids must be canonical without whitespace: %+v", decision)
			}
		})
	}
}

func TestCbtProctoringPolicyScreenshotRequiresBooleanPlatformCallback(t *testing.T) {
	cases := []struct {
		name     string
		data     map[string]any
		wantType string
		wantRisk int32
	}{
		{name: "missing", data: nil, wantType: "screenshot_attempt_ambiguous", wantRisk: 5},
		{name: "false", data: map[string]any{"valid_platform_callback": false}, wantType: "screenshot_attempt_ambiguous", wantRisk: 5},
		{name: "string true is not trusted", data: map[string]any{"valid_platform_callback": "true"}, wantType: "screenshot_attempt_ambiguous", wantRisk: 5},
		{name: "numeric true is not trusted", data: map[string]any{"valid_platform_callback": 1}, wantType: "screenshot_attempt_ambiguous", wantRisk: 5},
		{name: "boolean true", data: map[string]any{"valid_platform_callback": true}, wantType: "screenshot_attempt_valid", wantRisk: 25},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyProctorSeverity("screenshot_attempt", tc.data)
			if got.EventType != tc.wantType || got.RiskDelta != tc.wantRisk {
				t.Fatalf("ClassifyProctorSeverity(screenshot_attempt) = %+v, want type=%s risk=%d", got, tc.wantType, tc.wantRisk)
			}
		})
	}
}

func TestCbtProctoringPolicyAutoLockRules(t *testing.T) {
	cases := []struct {
		name  string
		event string
		state ParticipantRiskState
		want  bool
	}{
		{name: "already locked stays locked", event: "focus_lost_short", state: ParticipantRiskState{LockedAt: sql.NullTime{Time: time.Now(), Valid: true}}, want: true},
		{name: "token reuse immediate", event: "token_reuse_confirmed", state: ParticipantRiskState{RiskScore: 0}, want: true},
		{name: "root emulator immediate", event: "root_emulator_strong", state: ParticipantRiskState{RiskScore: 0}, want: true},
		{name: "device mismatch needs corroborating score", event: "device_mismatch_strong", state: ParticipantRiskState{RiskScore: 10}, want: false},
		{name: "device mismatch corroborated by accumulated score", event: "device_mismatch_strong", state: ParticipantRiskState{RiskScore: 25}, want: true},
		{name: "lock eligible third violation", event: "split_screen_detected", state: ParticipantRiskState{ViolationCount: 2, RiskScore: 10}, want: true},
		{name: "non lock eligible still locks at score threshold", event: "screenshot_attempt_valid", state: ParticipantRiskState{ViolationCount: 0, RiskScore: 60}, want: true},
		{name: "warning below threshold does not lock", event: "focus_lost_short", state: ParticipantRiskState{ViolationCount: 0, RiskScore: 10}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decision := ClassifyProctorSeverity(tc.event, nil)
			if got := ShouldAutoLock(decision, tc.state); got != tc.want {
				t.Fatalf("ShouldAutoLock(%s, %+v) = %v, want %v", tc.event, tc.state, got, tc.want)
			}
		})
	}
}

func TestCbtProctoringPolicyDedupKeyUsesCorrelationOrEpisode(t *testing.T) {
	participantID := "participant-1"
	if got := proctorDedupKey(participantID, "focus_lost_short", map[string]any{"correlation_id": " focus-episode "}); got != "participant-1:focus_lost_short:focus-episode" {
		t.Fatalf("proctorDedupKey(correlation) = %q", got)
	}
	if got := proctorDedupKey(participantID, "focus_lost_short", map[string]any{"episode_id": "episode-2"}); got != "participant-1:focus_lost_short:episode-2" {
		t.Fatalf("proctorDedupKey(episode) = %q", got)
	}
	if got := proctorDedupKey(participantID, "focus_lost_short", nil); got != "participant-1:focus_lost_short" {
		t.Fatalf("proctorDedupKey(default) = %q", got)
	}
}
