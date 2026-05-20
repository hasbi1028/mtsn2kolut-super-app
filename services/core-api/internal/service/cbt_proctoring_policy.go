package service

import (
	"database/sql"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"
)

type ProctorSeverity string

const (
	ProctorSeverityInfo      ProctorSeverity = "info"
	ProctorSeverityWarning   ProctorSeverity = "warning"
	ProctorSeverityMedium    ProctorSeverity = "medium"
	ProctorSeverityCritical  ProctorSeverity = "critical"
	ProctorSeverityTechnical ProctorSeverity = "technical"
)

type SeverityDecision struct {
	EventType    string
	Severity     ProctorSeverity
	Category     string
	RiskDelta    int32
	LockEligible bool
	RequiresNote bool
	AudioKey     string
	LabelID      string
	MessageID    string
}

type ParticipantRiskState struct {
	ViolationCount int32
	RiskScore      int32
	RiskLevel      string
	LockedAt       sql.NullTime
}

var proctorEventAliases = map[string]string{
	"warning":                          "focus_lost_short",
	"focus_lost":                       "focus_lost_short",
	"window_focus_lost":                "focus_lost_short",
	"app_backgrounded":                 "background_over_threshold",
	"app_background_resume":            "background_over_threshold",
	"picture_in_picture_detected":      "pip_detected",
	"anti_cheat_local_lock":            "root_emulator_strong",
	"split_screen":                     "split_screen_detected",
	"picture_in_picture":               "pip_detected",
	"screen_capture_attempt":           "screenshot_attempt_ambiguous",
	"submit_blocked_pending_sync":      "submit_held_pending_sync",
	"auto_submit_blocked_pending_sync": "submit_held_pending_sync",
	"browser_darurat":                  "web_fallback_used",
}

var proctorEventWhitelist = map[string]SeverityDecision{
	"heartbeat": {
		EventType: "heartbeat", Severity: ProctorSeverityInfo, Category: "connection", AudioKey: "none", LabelID: "heartbeat", MessageID: "status_koneksi",
	},
	"focus_lost_short": {
		EventType: "focus_lost_short", Severity: ProctorSeverityWarning, Category: "focus", RiskDelta: 5, AudioKey: "warning", LabelID: "focus_lost_short", MessageID: "fokus_berpindah",
	},
	"app_switch_once": {
		EventType: "app_switch_once", Severity: ProctorSeverityWarning, Category: "app_switch", RiskDelta: 10, AudioKey: "warning", LabelID: "app_switch_once", MessageID: "aplikasi_berpindah",
	},
	"app_switch": {
		EventType: "app_switch_once", Severity: ProctorSeverityWarning, Category: "app_switch", RiskDelta: 10, AudioKey: "warning", LabelID: "app_switch_once", MessageID: "aplikasi_berpindah",
	},
	"app_switch_repeated": {
		EventType: "app_switch_repeated", Severity: ProctorSeverityMedium, Category: "app_switch", RiskDelta: 25, LockEligible: true, RequiresNote: true, AudioKey: "medium", LabelID: "app_switch_repeated", MessageID: "aplikasi_berulang",
	},
	"background_over_threshold": {
		EventType: "background_over_threshold", Severity: ProctorSeverityMedium, Category: "app_background", RiskDelta: 30, LockEligible: true, RequiresNote: true, AudioKey: "medium", LabelID: "background_over_threshold", MessageID: "latar_belakang_terlalu_lama",
	},
	"split_screen_detected": {
		EventType: "split_screen_detected", Severity: ProctorSeverityMedium, Category: "windowing", RiskDelta: 30, LockEligible: true, RequiresNote: true, AudioKey: "medium", LabelID: "split_screen_detected", MessageID: "layar_terbagi",
	},
	"pip_detected": {
		EventType: "pip_detected", Severity: ProctorSeverityMedium, Category: "windowing", RiskDelta: 25, LockEligible: true, RequiresNote: true, AudioKey: "medium", LabelID: "pip_detected", MessageID: "picture_in_picture",
	},
	"overlay_suspicious_confirmed": {
		EventType: "overlay_suspicious_confirmed", Severity: ProctorSeverityMedium, Category: "overlay", RiskDelta: 25, RequiresNote: true, AudioKey: "medium", LabelID: "overlay_suspicious_confirmed", MessageID: "overlay_mencurigakan",
	},
	"screenshot_attempt": {
		EventType: "screenshot_attempt_ambiguous", Severity: ProctorSeverityWarning, Category: "screenshot", RiskDelta: 5, AudioKey: "warning", LabelID: "screenshot_attempt_ambiguous", MessageID: "indikasi_tangkapan_layar",
	},
	"screenshot_attempt_ambiguous": {
		EventType: "screenshot_attempt_ambiguous", Severity: ProctorSeverityWarning, Category: "screenshot", RiskDelta: 5, AudioKey: "warning", LabelID: "screenshot_attempt_ambiguous", MessageID: "indikasi_tangkapan_layar",
	},
	"screenshot_attempt_valid": {
		EventType: "screenshot_attempt_valid", Severity: ProctorSeverityMedium, Category: "screenshot", RiskDelta: 25, RequiresNote: true, AudioKey: "medium", LabelID: "screenshot_attempt_valid", MessageID: "percobaan_tangkapan_layar_valid",
	},
	"device_mismatch_weak": {
		EventType: "device_mismatch_weak", Severity: ProctorSeverityMedium, Category: "device", RiskDelta: 25, RequiresNote: true, AudioKey: "medium", LabelID: "device_mismatch_weak", MessageID: "perangkat_tidak_sesuai",
	},
	"device_mismatch_strong": {
		EventType: "device_mismatch_strong", Severity: ProctorSeverityCritical, Category: "device", RiskDelta: 60, LockEligible: true, RequiresNote: true, AudioKey: "critical", LabelID: "device_mismatch_strong", MessageID: "perangkat_tidak_sesuai_kuat",
	},
	"token_reuse_confirmed": {
		EventType: "token_reuse_confirmed", Severity: ProctorSeverityCritical, Category: "token_reuse", RiskDelta: 80, LockEligible: true, RequiresNote: true, AudioKey: "critical", LabelID: "token_reuse_confirmed", MessageID: "token_digunakan_ulang",
	},
	"root_emulator_weak": {
		EventType: "root_emulator_weak", Severity: ProctorSeverityMedium, Category: "device_integrity", RiskDelta: 25, RequiresNote: true, AudioKey: "medium", LabelID: "root_emulator_weak", MessageID: "integritas_perangkat_lemah",
	},
	"root_emulator_strong": {
		EventType: "root_emulator_strong", Severity: ProctorSeverityCritical, Category: "device_integrity", RiskDelta: 70, LockEligible: true, RequiresNote: true, AudioKey: "critical", LabelID: "root_emulator_strong", MessageID: "integritas_perangkat_kuat",
	},
	"offline_short": {
		EventType: "offline_short", Severity: ProctorSeverityTechnical, Category: "connection", AudioKey: "technical", LabelID: "offline_short", MessageID: "kontak_terlambat",
	},
	"offline_mass": {
		EventType: "offline_mass", Severity: ProctorSeverityTechnical, Category: "technical_mass", AudioKey: "technical", LabelID: "offline_mass", MessageID: "gangguan_teknis_massal",
	},
	"pending_sync": {
		EventType: "pending_sync", Severity: ProctorSeverityTechnical, Category: "sync", AudioKey: "technical", LabelID: "pending_sync", MessageID: "jawaban_belum_terkirim",
	},
	"submit_held_pending_sync": {
		EventType: "submit_held_pending_sync", Severity: ProctorSeverityTechnical, Category: "sync", RequiresNote: true, AudioKey: "technical", LabelID: "submit_held_pending_sync", MessageID: "submit_ditahan_sinkronisasi",
	},
	"web_fallback_used": {
		EventType: "web_fallback_used", Severity: ProctorSeverityInfo, Category: "web_fallback", AudioKey: "none", LabelID: "web_fallback_used", MessageID: "browser_darurat_digunakan",
	},
	"web_visibility_hidden": {
		EventType: "web_visibility_hidden", Severity: ProctorSeverityWarning, Category: "web_fallback", AudioKey: "warning", LabelID: "web_visibility_hidden", MessageID: "browser_tidak_terlihat",
	},
	"web_visibility_visible": {
		EventType: "web_visibility_visible", Severity: ProctorSeverityInfo, Category: "web_fallback", AudioKey: "none", LabelID: "web_visibility_visible", MessageID: "browser_terlihat_kembali",
	},
	"web_focus_lost": {
		EventType: "web_focus_lost", Severity: ProctorSeverityWarning, Category: "web_fallback", AudioKey: "warning", LabelID: "web_focus_lost", MessageID: "fokus_browser_berpindah",
	},
	"web_focus_restored": {
		EventType: "web_focus_restored", Severity: ProctorSeverityInfo, Category: "web_fallback", AudioKey: "none", LabelID: "web_focus_restored", MessageID: "fokus_browser_kembali",
	},
	"web_fullscreen_exit": {
		EventType: "web_fullscreen_exit", Severity: ProctorSeverityWarning, Category: "web_fallback", RiskDelta: 5, AudioKey: "warning", LabelID: "web_fullscreen_exit", MessageID: "fullscreen_browser_keluar",
	},
	"web_fullscreen_restored": {
		EventType: "web_fullscreen_restored", Severity: ProctorSeverityInfo, Category: "web_fallback", AudioKey: "none", LabelID: "web_fullscreen_restored", MessageID: "fullscreen_browser_kembali",
	},
	"web_pending_answer_saved": {
		EventType: "web_pending_answer_saved", Severity: ProctorSeverityTechnical, Category: "sync", AudioKey: "technical", LabelID: "web_pending_answer_saved", MessageID: "jawaban_browser_lokal",
	},
	"web_pending_answer_flushed": {
		EventType: "web_pending_answer_flushed", Severity: ProctorSeverityInfo, Category: "sync", AudioKey: "none", LabelID: "web_pending_answer_flushed", MessageID: "jawaban_browser_terkirim",
	},
	"web_connection_degraded": {
		EventType: "web_connection_degraded", Severity: ProctorSeverityTechnical, Category: "connection", AudioKey: "technical", LabelID: "web_connection_degraded", MessageID: "koneksi_browser_menurun",
	},
	"web_connection_restored": {
		EventType: "web_connection_restored", Severity: ProctorSeverityInfo, Category: "connection", AudioKey: "none", LabelID: "web_connection_restored", MessageID: "koneksi_browser_pulih",
	},
}

func NormalizeProctorEventType(raw string) (string, bool) {
	clean := strings.TrimSpace(strings.ToLower(raw))
	clean = strings.ReplaceAll(clean, "-", "_")
	clean = strings.ReplaceAll(clean, " ", "_")
	if clean == "" {
		return "", false
	}
	if alias, ok := proctorEventAliases[clean]; ok {
		clean = alias
	}
	decision, ok := proctorEventWhitelist[clean]
	if !ok {
		return "", false
	}
	return decision.EventType, true
}

func ClassifyProctorSeverity(eventType string, data map[string]any) SeverityDecision {
	normalized, ok := NormalizeProctorEventType(eventType)
	if !ok {
		return SeverityDecision{}
	}
	decision := proctorEventWhitelist[normalized]
	if decision.EventType == "" {
		for _, candidate := range proctorEventWhitelist {
			if candidate.EventType == normalized {
				decision = candidate
				break
			}
		}
	}
	if decision.EventType == "" {
		return SeverityDecision{}
	}
	if normalized == "screenshot_attempt_ambiguous" && hasBoolTrue(data, "valid_platform_callback") {
		decision = proctorEventWhitelist["screenshot_attempt_valid"]
	}
	if normalized == "app_switch_once" && intFromAny(data["count"]) >= 3 {
		decision = proctorEventWhitelist["app_switch_repeated"]
	}
	if decision.Severity == ProctorSeverityTechnical {
		decision.RiskDelta = 0
		decision.LockEligible = false
	}
	return decision
}

func ShouldDedup(eventType string) bool {
	normalized, ok := NormalizeProctorEventType(eventType)
	if !ok {
		return false
	}
	return DedupWindow(normalized) > 0
}

func DedupWindow(eventType string) time.Duration {
	normalized, _ := NormalizeProctorEventType(eventType)
	switch normalized {
	case "focus_lost_short":
		return 30 * time.Second
	case "offline_mass":
		return 2 * time.Minute
	case "heartbeat":
		return 20 * time.Second
	default:
		return time.Minute
	}
}

func RiskLevelFromScore(score int, violationCount int, lockedAt sql.NullTime) string {
	if lockedAt.Valid || score >= 80 {
		return "locked"
	}
	if score >= 50 || violationCount >= 2 {
		return "high"
	}
	if score >= 20 || violationCount >= 1 {
		return "warning"
	}
	return "normal"
}

func ShouldAutoLock(decision SeverityDecision, participantState ParticipantRiskState) bool {
	if decision.Severity == ProctorSeverityTechnical {
		return false
	}
	if participantState.LockedAt.Valid {
		return true
	}
	nextScore := int(participantState.RiskScore + decision.RiskDelta)
	nextViolations := int(participantState.ViolationCount)
	if decision.RiskDelta > 0 {
		nextViolations += 1
	}
	switch decision.EventType {
	case "token_reuse_confirmed", "root_emulator_strong":
		return true
	case "device_mismatch_strong":
		return hasCorroboratingCriticalScore(nextScore)
	}
	if !decision.LockEligible {
		return nextScore >= 80
	}
	return nextScore >= 80 || nextViolations >= 3
}

func proctorDedupKey(participantID, eventType string, data map[string]any) string {
	correlation := strings.TrimSpace(stringFromAny(data["correlation_id"]))
	if correlation == "" {
		correlation = strings.TrimSpace(stringFromAny(data["episode_id"]))
	}
	if correlation != "" {
		return participantID + ":" + eventType + ":" + correlation
	}
	return participantID + ":" + eventType
}

func hasCorroboratingCriticalScore(score int) bool {
	return score >= 80
}

func hasBoolTrue(data map[string]any, key string) bool {
	value, ok := data[key]
	if !ok {
		return false
	}
	typed, ok := value.(bool)
	return ok && typed
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(math.Round(typed))
	case json.Number:
		parsed, _ := strconv.Atoi(typed.String())
		return parsed
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(typed))
		return parsed
	default:
		return 0
	}
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}
