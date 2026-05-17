package service

import (
	"database/sql"
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
