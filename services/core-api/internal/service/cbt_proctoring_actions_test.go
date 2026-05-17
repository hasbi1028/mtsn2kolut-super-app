package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtProctoringActionHelpersNormalizeAndClassify(t *testing.T) {
	cases := []struct {
		raw       string
		want      string
		requires  bool
		eventType string
		severity  ProctorSeverity
	}{
		{raw: " Warn ", want: "warn_student", eventType: "proctor_warning", severity: ProctorSeverityWarning},
		{raw: "unlock-access", want: "unlock_access", requires: true, eventType: "proctor_unlock", severity: ProctorSeverityInfo},
		{raw: "LOCK", want: "hold_access", requires: true, eventType: "proctor_hold_access", severity: ProctorSeverityCritical},
		{raw: "reset_device", want: "reset_device_binding", requires: true, eventType: "proctor_reset_access", severity: ProctorSeverityWarning},
		{raw: "force_submit", want: "force_submit", requires: true, eventType: "proctor_force_submit", severity: ProctorSeverityCritical},
		{raw: "technical", want: "mark_technical_issue", eventType: "proctor_mark_technical", severity: ProctorSeverityTechnical},
		{raw: "incident", want: "mark_incident", eventType: "proctor_incident_action", severity: ProctorSeverityInfo},
		{raw: "escalate", want: "escalate_to_committee", requires: true, eventType: "proctor_escalate", severity: ProctorSeverityCritical},
		{raw: "clear", want: "clear_after_check", eventType: "proctor_clear_after_check", severity: ProctorSeverityInfo},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			got := normalizeProctorActionType(tc.raw)
			if got != tc.want {
				t.Fatalf("normalizeProctorActionType(%q) = %q, want %q", tc.raw, got, tc.want)
			}
			if requires := proctorActionRequiresNotes(got); requires != tc.requires {
				t.Fatalf("proctorActionRequiresNotes(%q) = %v, want %v", got, requires, tc.requires)
			}
			if eventType := proctorActionEventType(got); eventType != tc.eventType {
				t.Fatalf("proctorActionEventType(%q) = %q, want %q", got, eventType, tc.eventType)
			}
			if severity := proctorActionSeverity(got); severity != tc.severity {
				t.Fatalf("proctorActionSeverity(%q) = %q, want %q", got, severity, tc.severity)
			}
		})
	}
	if got := normalizeProctorActionType("approve"); got != "" {
		t.Fatalf("normalizeProctorActionType(unknown) = %q, want empty", got)
	}
}

func TestCbtProctoringStatusHelpers(t *testing.T) {
	now := time.Now()
	submitted := pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true}
	online := pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true}
	late := pgtype.Timestamptz{Time: now.Add(-5 * time.Minute), Valid: true}
	offline := pgtype.Timestamptz{Time: now.Add(-10 * time.Minute), Valid: true}

	connectionCases := []struct {
		name          string
		submittedAt   pgtype.Timestamptz
		lastHeartbeat pgtype.Timestamptz
		want          string
	}{
		{name: "submitted wins", submittedAt: submitted, lastHeartbeat: offline, want: "selesai"},
		{name: "missing heartbeat", want: "terputus"},
		{name: "online", lastHeartbeat: online, want: "online"},
		{name: "late", lastHeartbeat: late, want: "terlambat"},
		{name: "offline", lastHeartbeat: offline, want: "terputus"},
	}
	for _, tc := range connectionCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := proctorConnectionStatus(tc.submittedAt, tc.lastHeartbeat); got != tc.want {
				t.Fatalf("proctorConnectionStatus() = %q, want %q", got, tc.want)
			}
		})
	}

	syncCases := []struct {
		state   string
		pending int32
		want    string
	}{
		{state: "synced", pending: 9, want: "sinkron"},
		{state: "pending", want: "belum_sinkron"},
		{state: "failed", want: "tertahan"},
		{state: "", pending: 1, want: "belum_sinkron"},
		{state: "unknown", want: "tidak_diketahui"},
	}
	for _, tc := range syncCases {
		if got := proctorSyncStatus(tc.state, tc.pending); got != tc.want {
			t.Fatalf("proctorSyncStatus(%q, %d) = %q, want %q", tc.state, tc.pending, got, tc.want)
		}
	}
}

func TestCbtProctoringNeedsActionHelpers(t *testing.T) {
	lockedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	critical := &CbtProctoringEventDTO{Severity: string(ProctorSeverityCritical)}
	medium := &CbtProctoringEventDTO{Severity: string(ProctorSeverityMedium)}
	technical := &CbtProctoringEventDTO{Severity: string(ProctorSeverityTechnical)}
	ackedCritical := &CbtProctoringEventDTO{Severity: string(ProctorSeverityCritical), AcknowledgedAt: "2026-05-01T08:00:00Z"}

	cases := []struct {
		name           string
		row            db.GetSessionProctoringStatusRow
		connection     string
		syncStatus     string
		unacknowledged int
		latest         *CbtProctoringEventDTO
		wantNeeds      bool
		wantPriority   int
		wantReason     string
	}{
		{name: "locked at", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal", LockedAt: lockedAt}, wantNeeds: true, wantPriority: 100, wantReason: "Akses ditahan"},
		{name: "locked level", row: db.GetSessionProctoringStatusRow{RiskLevel: "locked"}, wantNeeds: true, wantPriority: 100, wantReason: "Akses ditahan"},
		{name: "critical latest", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal"}, latest: critical, wantNeeds: true, wantPriority: 95, wantReason: "Peringatan kritis belum dicek"},
		{name: "high risk", row: db.GetSessionProctoringStatusRow{RiskLevel: "high"}, latest: ackedCritical, wantNeeds: true, wantPriority: 80, wantReason: "Risiko tinggi"},
		{name: "medium latest", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal"}, latest: medium, wantNeeds: true, wantPriority: 70, wantReason: "Peringatan belum dicek"},
		{name: "technical latest", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal"}, latest: technical, wantNeeds: true, wantPriority: 70, wantReason: "Peringatan belum dicek"},
		{name: "other unacked", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal"}, unacknowledged: 2, wantNeeds: true, wantPriority: 60, wantReason: "Ada peringatan belum dicek"},
		{name: "sync held", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal"}, syncStatus: "tertahan", wantNeeds: true, wantPriority: 50, wantReason: "Jawaban belum terkirim"},
		{name: "disconnected", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal"}, connection: "terputus", wantNeeds: true, wantPriority: 40, wantReason: "Terputus"},
		{name: "late", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal"}, connection: "terlambat", wantNeeds: true, wantPriority: 30, wantReason: "Kontak terlambat"},
		{name: "warning", row: db.GetSessionProctoringStatusRow{RiskLevel: "warning"}, wantNeeds: true, wantPriority: 20, wantReason: "Perlu perhatian"},
		{name: "normal", row: db.GetSessionProctoringStatusRow{RiskLevel: "normal"}, connection: "online", syncStatus: "sinkron", wantNeeds: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			needs, priority, reason := proctorParticipantNeedsAction(tc.row, tc.connection, tc.syncStatus, tc.unacknowledged, tc.latest)
			if needs != tc.wantNeeds || priority != tc.wantPriority || reason != tc.wantReason {
				t.Fatalf("proctorParticipantNeedsAction() = %v/%d/%q, want %v/%d/%q", needs, priority, reason, tc.wantNeeds, tc.wantPriority, tc.wantReason)
			}
		})
	}

	eventCases := []struct {
		severity string
		want     bool
	}{
		{severity: string(ProctorSeverityMedium), want: true},
		{severity: string(ProctorSeverityCritical), want: true},
		{severity: string(ProctorSeverityTechnical), want: true},
		{severity: string(ProctorSeverityWarning), want: false},
		{severity: string(ProctorSeverityInfo), want: false},
	}
	for _, tc := range eventCases {
		if got := proctorEventNeedsAction(CbtProctoringEventDTO{Severity: tc.severity}); got != tc.want {
			t.Fatalf("proctorEventNeedsAction(%q) = %v, want %v", tc.severity, got, tc.want)
		}
	}
}

func TestCbtProctoringEventDTOMapsPayloadDefaultsAndTimestamps(t *testing.T) {
	id := cbtProctoringActionTestUUID(t, "11111111-1111-1111-1111-111111111111")
	participantID := cbtProctoringActionTestUUID(t, "22222222-2222-2222-2222-222222222222")
	studentID := cbtProctoringActionTestUUID(t, "33333333-3333-3333-3333-333333333333")
	sessionID := cbtProctoringActionTestUUID(t, "44444444-4444-4444-4444-444444444444")
	roomID := cbtProctoringActionTestUUID(t, "55555555-5555-5555-5555-555555555555")
	ackBy := cbtProctoringActionTestUUID(t, "66666666-6666-6666-6666-666666666666")
	createdAt := cbtProctoringActionTestTimestamp(t, "2026-05-01T08:15:30Z")
	ackedAt := cbtProctoringActionTestTimestamp(t, "2026-05-01T08:20:00Z")
	payload := []byte(`{"label_id":"label.focus","message_id":"message.focus","audio_key":"alarm.warning","extra":7}`)

	dto := proctorEventDTO(id, participantID, studentID, sessionID, roomID, "12345", "Siswa A", "Ruang 1", "offline_mass", "", "technical_mass", 0, payload, createdAt, ackedAt, ackBy, "checked", true, "proctor-1")

	if dto.ID != pgUUIDString(id) || dto.LegacyID != dto.ID || dto.ParticipantID != pgUUIDString(participantID) || dto.StudentID != pgUUIDString(studentID) || dto.SessionID != pgUUIDString(sessionID) || dto.RoomID != pgUUIDString(roomID) {
		t.Fatalf("uuid fields not mapped correctly: %+v", dto)
	}
	if dto.NIS != "12345" || dto.Nama != "Siswa A" || dto.ParticipantName != "Siswa A" || dto.RoomName != "Ruang 1" {
		t.Fatalf("identity fields not mapped correctly: %+v", dto)
	}
	if dto.EventType != "offline_mass" || dto.Severity != "info" || dto.Category != "technical_mass" || dto.RiskDelta != 0 || !dto.IsMassTechnicalIssue {
		t.Fatalf("classification fields = %+v", dto)
	}
	if dto.LabelID != "label.focus" || dto.MessageID != "message.focus" || dto.AudioKey != "alarm.warning" {
		t.Fatalf("payload-derived fields = %+v", dto)
	}
	if dto.AcknowledgedAt != "2026-05-01T08:20:00Z" || dto.CreatedAt != "2026-05-01T08:15:30Z" || dto.AcknowledgedBy != pgUUIDString(ackBy) || dto.AcknowledgeNote != "checked" || !dto.RequiresNote || dto.ActorUsernameSnapshot != "proctor-1" {
		t.Fatalf("ack/timestamp fields = %+v", dto)
	}
	data, ok := dto.EventData.(map[string]any)
	if !ok || data["extra"].(float64) != 7 {
		t.Fatalf("EventData = %#v, want unmarshaled map", dto.EventData)
	}

	fallback := proctorEventDTO(id, participantID, studentID, sessionID, roomID, "", "", "", "focus_lost_short", "warning", "", 5, []byte(`not-json`), createdAt, pgtype.Timestamptz{}, pgtype.UUID{}, "", false, "")
	if fallback.LabelID != "focus_lost_short" || fallback.MessageID != "focus_lost_short" || fallback.AudioKey != "warning" || fallback.Category != "timeline" || fallback.AcknowledgedAt != "" || fallback.AcknowledgedBy != "" {
		t.Fatalf("fallback dto = %+v", fallback)
	}
}

func TestCbtProctoringEventDTOsFromRows(t *testing.T) {
	sessionRows := []db.ListCbtProctorEventsBySessionRow{{
		ID: cbtProctoringActionTestUUID(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), ParticipantID: cbtProctoringActionTestUUID(t, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), StudentID: cbtProctoringActionTestUUID(t, "cccccccc-cccc-cccc-cccc-cccccccccccc"), SessionID: cbtProctoringActionTestUUID(t, "dddddddd-dddd-dddd-dddd-dddddddddddd"), RoomID: cbtProctoringActionTestUUID(t, "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
		Nis: "111", Nama: "Siswa Session", RoomName: "Lab", EventType: "app_switch_once", Severity: "warning", Category: "integrity", RiskDelta: 10, EventData: []byte(`{"label_id":"label.session"}`), CreatedAt: cbtProctoringActionTestTimestamp(t, "2026-05-01T09:00:00Z"), RequiresNote: true, ActorUsernameSnapshot: "proctor-session",
	}}
	roomRows := []db.ListCbtProctorEventsByRoomRow{{
		ID: cbtProctoringActionTestUUID(t, "abababab-abab-abab-abab-abababababab"), ParticipantID: cbtProctoringActionTestUUID(t, "bcbcbcbc-bcbc-bcbc-bcbc-bcbcbcbcbcbc"), StudentID: cbtProctoringActionTestUUID(t, "cdcdcdcd-cdcd-cdcd-cdcd-cdcdcdcdcdcd"), SessionID: cbtProctoringActionTestUUID(t, "dededede-dede-dede-dede-dededededede"), RoomID: cbtProctoringActionTestUUID(t, "efefefef-efef-efef-efef-efefefefefef"),
		Nis: "222", Nama: "Siswa Room", RoomName: "Lab 2", EventType: "pending_sync", Severity: "technical", Category: "technical", RiskDelta: 0, EventData: []byte(`{"message_id":"message.room"}`), CreatedAt: cbtProctoringActionTestTimestamp(t, "2026-05-01T09:05:00Z"), ActorUsernameSnapshot: "proctor-room",
	}}

	sessionDTOs := proctorEventDTOsFromSessionRows(sessionRows)
	roomDTOs := proctorEventDTOsFromRoomRows(roomRows)
	if len(sessionDTOs) != 1 || sessionDTOs[0].Nama != "Siswa Session" || sessionDTOs[0].LabelID != "label.session" || !sessionDTOs[0].RequiresNote || sessionDTOs[0].ActorUsernameSnapshot != "proctor-session" {
		t.Fatalf("session DTOs = %+v", sessionDTOs)
	}
	if len(roomDTOs) != 1 || roomDTOs[0].Nama != "Siswa Room" || roomDTOs[0].MessageID != "message.room" || roomDTOs[0].ActorUsernameSnapshot != "proctor-room" {
		t.Fatalf("room DTOs = %+v", roomDTOs)
	}
}

func TestCbtProctoringFirstEventsAndTimeString(t *testing.T) {
	events := []CbtProctoringEventDTO{{ID: "1"}, {ID: "2"}, {ID: "3"}}
	if got := firstProctorEvents(events, 5); len(got) != 3 {
		t.Fatalf("firstProctorEvents(no trim) len = %d, want 3", len(got))
	}
	got := firstProctorEvents(events, 2)
	if len(got) != 2 || got[0].ID != "1" || got[1].ID != "2" {
		t.Fatalf("firstProctorEvents(trim) = %+v", got)
	}
	if got := proctorTimeString(pgtype.Timestamptz{}); got != "" {
		t.Fatalf("proctorTimeString(invalid) = %q, want empty", got)
	}
	if got := proctorTimeString(cbtProctoringActionTestTimestamp(t, "2026-05-01T10:00:00Z")); got != "2026-05-01T10:00:00Z" {
		t.Fatalf("proctorTimeString(valid) = %q", got)
	}
}

func TestCbtProctoringBuildLiveSummaryMapsParticipantsRoomsAndAlarms(t *testing.T) {
	sessionID := cbtProctoringActionTestUUID(t, "10101010-1010-1010-1010-101010101010")
	roomA := cbtProctoringActionTestUUID(t, "20202020-2020-2020-2020-202020202020")
	roomB := cbtProctoringActionTestUUID(t, "30303030-3030-3030-3030-303030303030")
	participantA := cbtProctoringActionTestUUID(t, "40404040-4040-4040-4040-404040404040")
	participantB := cbtProctoringActionTestUUID(t, "50505050-5050-5050-5050-505050505050")
	studentA := cbtProctoringActionTestUUID(t, "60606060-6060-6060-6060-606060606060")
	studentB := cbtProctoringActionTestUUID(t, "70707070-7070-7070-7070-707070707070")
	now := time.Now()

	rows := []db.GetSessionProctoringStatusRow{
		{ParticipantID: participantA, StudentID: studentA, Nis: "001", Nama: "Budi", RoomID: roomA, RoomName: "Ruang A", SeatNo: pgtype.Int4{Int32: 1, Valid: true}, LastHeartbeat: pgtype.Timestamptz{Time: now.Add(-10 * time.Minute), Valid: true}, RiskLevel: "normal", SyncState: "synced", AnsweredCount: 10, AppSwitchCount: 1},
		{ParticipantID: participantB, StudentID: studentB, Nis: "002", Nama: "Ani", RoomID: roomB, RoomName: "Ruang B", SubmittedAt: pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true}, LastHeartbeat: pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true}, RiskLevel: "high", RiskScore: 55, ViolationCount: 2, SyncState: "pending", PendingAnswerCount: 3, ScreenshotAttempt: 1, SuspiciousFlag: true},
	}
	events := []CbtProctoringEventDTO{
		{ID: "evt-a", ParticipantID: pgUUIDString(participantA), Severity: string(ProctorSeverityTechnical), EventType: "pending_sync"},
		{ID: "evt-b", ParticipantID: pgUUIDString(participantB), Severity: string(ProctorSeverityWarning), EventType: "focus_lost_short"},
	}
	actions := []db.CbtProctorAction{{ActionType: "warn_student", Reason: "focus"}}

	summary := buildProctoringLiveSummary(sessionID, pgtype.UUID{}, rows, events, actions)
	if summary.SessionID != pgUUIDString(sessionID) || summary.RoomID != "" || summary.GeneratedAt == "" {
		t.Fatalf("summary ids/timestamp = %+v", summary)
	}
	if summary.Counts["normal"] != 1 || summary.Counts["high"] != 1 || summary.Counts["offline"] != 1 || summary.Counts["submitted"] != 1 || summary.Counts["pending_sync"] != 1 {
		t.Fatalf("summary counts = %+v", summary.Counts)
	}
	if len(summary.Participants) != 2 || summary.Participants[0].ParticipantID != pgUUIDString(participantB) || !summary.Participants[0].NeedsAction || summary.Participants[0].NeedsActionReason != "Risiko tinggi" || summary.Participants[1].ConnectionStatus != "terputus" {
		t.Fatalf("participants sorted/mapped = %+v", summary.Participants)
	}
	if summary.Participants[1].LatestEvent == nil || summary.Participants[1].LatestEvent.ID != "evt-a" || summary.Participants[1].UnacknowledgedCount != 1 {
		t.Fatalf("latest/unacked participant A = %+v", summary.Participants[1])
	}
	if len(summary.Rooms) != 2 || summary.Rooms[0].RoomID != pgUUIDString(roomA) || summary.Rooms[0].OfflineCount != 1 || summary.Rooms[0].NeedsActionCount != 1 || summary.Rooms[1].PendingSyncCount != 1 {
		t.Fatalf("rooms = %+v", summary.Rooms)
	}
	if len(summary.LatestEvents) != 2 || len(summary.AlarmEvents) != 1 || summary.AlarmEvents[0].ID != "evt-a" || summary.UnacknowledgedCount != 1 || len(summary.Actions) != 1 {
		t.Fatalf("events/actions summary = latest:%+v alarms:%+v actions:%+v unacked:%d", summary.LatestEvents, summary.AlarmEvents, summary.Actions, summary.UnacknowledgedCount)
	}
}

func cbtProctoringActionTestUUID(t *testing.T, raw string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		t.Fatalf("uuid scan failed: %v", err)
	}
	return id
}

func cbtProctoringActionTestTimestamp(t *testing.T, raw string) pgtype.Timestamptz {
	t.Helper()
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("time parse failed: %v", err)
	}
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func TestCbtProctoringEventDTOJSONShape(t *testing.T) {
	dto := CbtProctoringEventDTO{ID: "event-1", LegacyID: "event-1", ParticipantName: "Siswa", RequiresNote: true}
	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("json marshal dto: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("json unmarshal dto: %v", err)
	}
	if decoded["event_id"] != "event-1" || decoded["id"] != "event-1" || decoded["participant_name"] != "Siswa" || decoded["requires_note"] != true {
		t.Fatalf("json shape = %s", string(body))
	}
}
