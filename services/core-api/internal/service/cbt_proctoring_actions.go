package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtProctorActor struct {
	UserID     pgtype.UUID
	Username   string
	EmployeeID pgtype.UUID
	RequestID  string
	SourceIP   string
}

type CbtProctorActionInput struct {
	SessionID     pgtype.UUID
	ParticipantID pgtype.UUID
	EventID       pgtype.UUID
	ActionType    string
	Reason        string
	Notes         string
}

type CbtProctorActionResult struct {
	Action      db.CbtProctorAction `json:"action"`
	Event       any                 `json:"event"`
	Participant any                 `json:"participant"`
}

type CbtProctoringEventDTO struct {
	ID                    string `json:"event_id"`
	LegacyID              string `json:"id"`
	ParticipantID         string `json:"participant_id"`
	StudentID             string `json:"student_id"`
	NIS                   string `json:"nis"`
	Nama                  string `json:"nama"`
	ParticipantName       string `json:"participant_name"`
	SessionID             string `json:"session_id"`
	RoomID                string `json:"room_id"`
	RoomName              string `json:"room_name"`
	EventType             string `json:"event_type"`
	Severity              string `json:"severity"`
	Category              string `json:"category"`
	RiskDelta             int32  `json:"risk_delta"`
	LabelID               string `json:"label_id"`
	MessageID             string `json:"message_id"`
	AudioKey              string `json:"audio_key"`
	RequiresNote          bool   `json:"requires_note"`
	IsMassTechnicalIssue  bool   `json:"is_mass_technical_issue"`
	AcknowledgedAt        string `json:"acknowledged_at"`
	AcknowledgedBy        string `json:"acknowledged_by"`
	AcknowledgeNote       string `json:"acknowledge_note"`
	ActorUsernameSnapshot string `json:"actor_username_snapshot"`
	EventData             any    `json:"event_data"`
	CreatedAt             string `json:"created_at"`
}

type CbtProctoringParticipantDTO struct {
	ParticipantID       string                 `json:"participant_id"`
	StudentID           string                 `json:"student_id"`
	NIS                 string                 `json:"nis"`
	Nama                string                 `json:"nama"`
	RoomID              string                 `json:"room_id"`
	RoomName            string                 `json:"room_name"`
	SeatNo              any                    `json:"seat_no"`
	SubmittedAt         any                    `json:"submitted_at"`
	LastHeartbeat       any                    `json:"last_heartbeat"`
	ConnectionStatus    string                 `json:"connection_status"`
	SyncStatus          string                 `json:"sync_status"`
	RiskLevel           string                 `json:"risk_level"`
	RiskScore           int32                  `json:"risk_score"`
	ViolationCount      int32                  `json:"violation_count"`
	LockedAt            any                    `json:"locked_at"`
	LockedReason        any                    `json:"locked_reason"`
	NeedsAction         bool                   `json:"needs_action"`
	ActionPriority      int                    `json:"action_priority"`
	NeedsActionReason   string                 `json:"needs_action_reason"`
	LatestEvent         *CbtProctoringEventDTO `json:"latest_event"`
	UnacknowledgedCount int                    `json:"unacknowledged_count"`
	PendingAnswerCount  int32                  `json:"pending_answer_count"`
	LastLocalSaveAt     any                    `json:"last_local_save_at"`
	LastSyncedAt        any                    `json:"last_synced_at"`
	AnsweredCount       int32                  `json:"answered_count"`
	Score               any                    `json:"score"`
	AppSwitchCount      int32                  `json:"app_switch_count"`
	ScreenshotAttempt   int32                  `json:"screenshot_attempt"`
	SuspiciousFlag      bool                   `json:"suspicious_flag"`
}

type CbtProctoringRoomSummaryDTO struct {
	RoomID           string `json:"room_id"`
	RoomName         string `json:"room_name"`
	ParticipantCount int    `json:"participant_count"`
	NormalCount      int    `json:"normal_count"`
	WarningCount     int    `json:"warning_count"`
	HighCount        int    `json:"high_count"`
	CriticalCount    int    `json:"critical_count"`
	TechnicalCount   int    `json:"technical_count"`
	OfflineCount     int    `json:"offline_count"`
	SubmittedCount   int    `json:"submitted_count"`
	PendingSyncCount int    `json:"pending_sync_count"`
	NeedsActionCount int    `json:"needs_action_count"`
}

type CbtProctoringLiveSummary struct {
	SessionID           string                        `json:"session_id"`
	RoomID              string                        `json:"room_id,omitempty"`
	GeneratedAt         string                        `json:"generated_at"`
	Counts              map[string]int                `json:"counts"`
	Rooms               []CbtProctoringRoomSummaryDTO `json:"rooms"`
	Participants        []CbtProctoringParticipantDTO `json:"participants"`
	LatestEvents        []CbtProctoringEventDTO       `json:"latest_events"`
	AlarmEvents         []CbtProctoringEventDTO       `json:"alarm_events"`
	UnacknowledgedCount int                           `json:"unacknowledged_count"`
	Actions             []db.CbtProctorAction         `json:"actions"`
}

type cbtProctoringStore interface {
	GetCbtParticipantProctorScope(ctx context.Context, arg db.GetCbtParticipantProctorScopeParams) (db.GetCbtParticipantProctorScopeRow, error)
	GetCbtProctorEventScope(ctx context.Context, id pgtype.UUID) (db.GetCbtProctorEventScopeRow, error)
	AcknowledgeCbtProctorEvent(ctx context.Context, arg db.AcknowledgeCbtProctorEventParams) (db.AcknowledgeCbtProctorEventRow, error)
	CreateCbtParticipantProctorEvent(ctx context.Context, arg db.CreateCbtParticipantProctorEventParams) (db.CreateCbtParticipantProctorEventRow, error)
	GetCbtParticipantRiskForUpdate(ctx context.Context, id pgtype.UUID) (db.GetCbtParticipantRiskForUpdateRow, error)
	CreateCbtProctorAction(ctx context.Context, arg db.CreateCbtProctorActionParams) (db.CbtProctorAction, error)
	UnlockParticipantAccessForProctor(ctx context.Context, id pgtype.UUID) (db.UnlockParticipantAccessForProctorRow, error)
	HoldParticipantAccessForProctor(ctx context.Context, arg db.HoldParticipantAccessForProctorParams) (db.HoldParticipantAccessForProctorRow, error)
	ResetParticipantDeviceBindingForProctor(ctx context.Context, id pgtype.UUID) (db.ResetParticipantDeviceBindingForProctorRow, error)
	ForceSubmitParticipant(ctx context.Context, arg db.ForceSubmitParticipantParams) (db.ForceSubmitParticipantRow, error)
	UpdateParticipantAnswerCorrectness(ctx context.Context, participantID pgtype.UUID) error
	ListCbtProctorEventsBySession(ctx context.Context, arg db.ListCbtProctorEventsBySessionParams) ([]db.ListCbtProctorEventsBySessionRow, error)
	ListCbtProctorEventsByRoom(ctx context.Context, arg db.ListCbtProctorEventsByRoomParams) ([]db.ListCbtProctorEventsByRoomRow, error)
	ListCbtProctorActionsBySession(ctx context.Context, arg db.ListCbtProctorActionsBySessionParams) ([]db.CbtProctorAction, error)
	ListCbtProctorActionsByRoom(ctx context.Context, arg db.ListCbtProctorActionsByRoomParams) ([]db.CbtProctorAction, error)
}

func (s *CbtSession) GetProctorEventScope(ctx context.Context, eventID pgtype.UUID) (db.GetCbtProctorEventScopeRow, error) {
	q, ok := s.q.(cbtProctoringStore)
	if !ok {
		return db.GetCbtProctorEventScopeRow{}, fmt.Errorf("cbt proctoring store unavailable")
	}
	return q.GetCbtProctorEventScope(ctx, eventID)
}

func (s *CbtSession) GetParticipantProctorScope(ctx context.Context, sessionID, participantID pgtype.UUID) (db.GetCbtParticipantProctorScopeRow, error) {
	q, ok := s.q.(cbtProctoringStore)
	if !ok {
		return db.GetCbtParticipantProctorScopeRow{}, fmt.Errorf("cbt proctoring store unavailable")
	}
	return q.GetCbtParticipantProctorScope(ctx, db.GetCbtParticipantProctorScopeParams{SessionID: sessionID, ParticipantID: participantID})
}

func (s *CbtSession) AcknowledgeProctorEventByID(ctx context.Context, eventID pgtype.UUID, actor CbtProctorActor, note string) (db.AcknowledgeCbtProctorEventRow, error) {
	q, ok := s.q.(cbtProctoringStore)
	if !ok {
		return db.AcknowledgeCbtProctorEventRow{}, fmt.Errorf("cbt proctoring store unavailable")
	}
	if !actor.UserID.Valid {
		return db.AcknowledgeCbtProctorEventRow{}, fmt.Errorf("%w: aktor pengawas wajib", domain.ErrUnauthorized)
	}
	scope, err := q.GetCbtProctorEventScope(ctx, eventID)
	if err != nil {
		return db.AcknowledgeCbtProctorEventRow{}, err
	}
	note = strings.TrimSpace(note)
	if (scope.RequiresNote || scope.Severity == string(ProctorSeverityCritical)) && note == "" {
		return db.AcknowledgeCbtProctorEventRow{}, fmt.Errorf("%w: catatan wajib untuk peringatan ini", domain.ErrBadRequest)
	}
	row, err := q.AcknowledgeCbtProctorEvent(ctx, db.AcknowledgeCbtProctorEventParams{
		ID:              eventID,
		AcknowledgedBy:  actor.UserID,
		AcknowledgeNote: note,
	})
	if err != nil {
		return db.AcknowledgeCbtProctorEventRow{}, err
	}
	_, _ = q.CreateCbtParticipantProctorEvent(ctx, db.CreateCbtParticipantProctorEventParams{
		ParticipantID:         scope.ParticipantID,
		EventType:             "proctor_acknowledge",
		EventData:             marshalJSON(map[string]any{"event_id": pgUUIDString(eventID), "actor": actor.Username, "notes": note}),
		Severity:              string(ProctorSeverityInfo),
		Category:              "proctor_action",
		RiskDelta:             0,
		DedupKey:              "",
		RequiresNote:          false,
		ActorUserID:           actor.UserID,
		ActorUsernameSnapshot: actor.Username,
		ActorEmployeeID:       actor.EmployeeID,
		RequestID:             actor.RequestID,
		SourceIp:              actor.SourceIP,
	})
	return row, nil
}

func (s *CbtSession) ExecuteProctorAction(ctx context.Context, in CbtProctorActionInput, actor CbtProctorActor) (CbtProctorActionResult, error) {
	if !actor.UserID.Valid {
		return CbtProctorActionResult{}, fmt.Errorf("%w: aktor tindakan pengawas wajib", domain.ErrUnauthorized)
	}
	actionType := normalizeProctorActionType(in.ActionType)
	if actionType == "" {
		return CbtProctorActionResult{}, fmt.Errorf("%w: tindakan pengawas tidak valid", domain.ErrBadRequest)
	}
	reason := strings.TrimSpace(in.Reason)
	notes := strings.TrimSpace(in.Notes)
	if reason == "" {
		return CbtProctorActionResult{}, fmt.Errorf("%w: alasan tindakan wajib diisi", domain.ErrBadRequest)
	}
	if proctorActionRequiresNotes(actionType) && notes == "" {
		return CbtProctorActionResult{}, fmt.Errorf("%w: catatan tindakan wajib diisi", domain.ErrBadRequest)
	}
	if s.pool == nil {
		q, ok := s.q.(cbtProctoringStore)
		if !ok {
			return CbtProctorActionResult{}, fmt.Errorf("cbt proctoring store unavailable")
		}
		return executeProctorActionWithStore(ctx, q, in, actor, actionType, reason, notes)
	}
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return CbtProctorActionResult{}, err
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return CbtProctorActionResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	result, err := executeProctorActionWithStore(ctx, s.q.WithTx(tx), in, actor, actionType, reason, notes)
	if err != nil {
		return CbtProctorActionResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CbtProctorActionResult{}, err
	}
	return result, nil
}

func executeProctorActionWithStore(ctx context.Context, q cbtProctoringStore, in CbtProctorActionInput, actor CbtProctorActor, actionType, reason, notes string) (CbtProctorActionResult, error) {
	scope, err := q.GetCbtParticipantProctorScope(ctx, db.GetCbtParticipantProctorScopeParams{SessionID: in.SessionID, ParticipantID: in.ParticipantID})
	if err != nil {
		return CbtProctorActionResult{}, err
	}
	state, err := q.GetCbtParticipantRiskForUpdate(ctx, in.ParticipantID)
	if err != nil {
		return CbtProctorActionResult{}, err
	}
	var participant any = map[string]any{
		"id":                   pgUUIDString(state.ID),
		"risk_score":           state.RiskScore,
		"risk_level":           state.RiskLevel,
		"violation_count":      state.ViolationCount,
		"locked_at":            state.LockedAt,
		"locked_reason":        state.LockedReason,
		"sync_state":           state.SyncState,
		"pending_answer_count": state.PendingAnswerCount,
	}
	switch actionType {
	case "unlock_access":
		row, err := q.UnlockParticipantAccessForProctor(ctx, in.ParticipantID)
		if err != nil {
			return CbtProctorActionResult{}, err
		}
		participant = row
	case "hold_access":
		row, err := q.HoldParticipantAccessForProctor(ctx, db.HoldParticipantAccessForProctorParams{
			ID:           in.ParticipantID,
			LockedReason: pgtype.Text{String: reason, Valid: true},
		})
		if err != nil {
			return CbtProctorActionResult{}, err
		}
		participant = row
	case "reset_device_binding":
		row, err := q.ResetParticipantDeviceBindingForProctor(ctx, in.ParticipantID)
		if err != nil {
			return CbtProctorActionResult{}, err
		}
		participant = row
	case "force_submit":
		if state.PendingAnswerCount > 0 || state.SyncState == "pending" || state.SyncState == "failed" {
			if !strings.Contains(strings.ToLower(reason+" "+notes), "sinkron") && !strings.Contains(strings.ToLower(reason+" "+notes), "pending") {
				return CbtProctorActionResult{}, fmt.Errorf("%w: force submit saat jawaban belum sinkron wajib menyebut risiko sinkronisasi", domain.ErrBadRequest)
			}
		}
		row, err := q.ForceSubmitParticipant(ctx, db.ForceSubmitParticipantParams{SessionID: in.SessionID, ID: in.ParticipantID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return CbtProctorActionResult{}, fmt.Errorf("%w: peserta sudah submit atau tidak ada", domain.ErrConflict)
			}
			return CbtProctorActionResult{}, err
		}
		if err := q.UpdateParticipantAnswerCorrectness(ctx, in.ParticipantID); err != nil {
			return CbtProctorActionResult{}, err
		}
		participant = row
	case "mark_technical_issue", "mark_incident", "warn_student", "escalate_to_committee", "clear_after_check":
		// Ledger/timeline only. Technical marking remains zero-risk by invariant.
	default:
		return CbtProctorActionResult{}, fmt.Errorf("%w: tindakan pengawas tidak valid", domain.ErrBadRequest)
	}
	action, err := q.CreateCbtProctorAction(ctx, db.CreateCbtProctorActionParams{
		SessionID:             in.SessionID,
		RoomID:                scope.RoomID,
		ParticipantID:         in.ParticipantID,
		EventID:               in.EventID,
		ActionType:            actionType,
		Reason:                reason,
		Notes:                 notes,
		ActorUserID:           actor.UserID,
		ActorUsernameSnapshot: actor.Username,
		ActorEmployeeID:       actor.EmployeeID,
		RequestID:             actor.RequestID,
		SourceIp:              actor.SourceIP,
	})
	if err != nil {
		return CbtProctorActionResult{}, err
	}
	eventRow, err := q.CreateCbtParticipantProctorEvent(ctx, db.CreateCbtParticipantProctorEventParams{
		ParticipantID: in.ParticipantID,
		EventType:     proctorActionEventType(actionType),
		EventData: marshalJSON(map[string]any{
			"action_id":   pgUUIDString(action.ID),
			"action_type": actionType,
			"reason":      reason,
			"notes":       notes,
			"actor":       actor.Username,
			"session_id":  pgUUIDString(in.SessionID),
			"room_id":     pgUUIDString(scope.RoomID),
		}),
		Severity:              string(proctorActionSeverity(actionType)),
		Category:              "proctor_action",
		RiskDelta:             0,
		DedupKey:              "",
		RequiresNote:          false,
		ActorUserID:           actor.UserID,
		ActorUsernameSnapshot: actor.Username,
		ActorEmployeeID:       actor.EmployeeID,
		RequestID:             actor.RequestID,
		SourceIp:              actor.SourceIP,
	})
	if err != nil {
		return CbtProctorActionResult{}, err
	}
	return CbtProctorActionResult{Action: action, Event: eventRow, Participant: participant}, nil
}

func (s *CbtSession) GetProctoringLiveSummary(ctx context.Context, sessionID, roomID pgtype.UUID) (CbtProctoringLiveSummary, error) {
	participants, err := s.GetProctoringStatusForRoom(ctx, sessionID, roomID)
	if err != nil {
		return CbtProctoringLiveSummary{}, err
	}
	q, ok := s.q.(cbtProctoringStore)
	if !ok {
		return CbtProctoringLiveSummary{}, fmt.Errorf("cbt proctoring store unavailable")
	}
	var events []CbtProctoringEventDTO
	if roomID.Valid {
		rows, err := q.ListCbtProctorEventsByRoom(ctx, db.ListCbtProctorEventsByRoomParams{SessionID: sessionID, RoomID: roomID, LimitCount: 300})
		if err != nil {
			return CbtProctoringLiveSummary{}, err
		}
		events = proctorEventDTOsFromRoomRows(rows)
	} else {
		rows, err := q.ListCbtProctorEventsBySession(ctx, db.ListCbtProctorEventsBySessionParams{SessionID: sessionID, LimitCount: 300})
		if err != nil {
			return CbtProctoringLiveSummary{}, err
		}
		events = proctorEventDTOsFromSessionRows(rows)
	}
	actions, err := listProctorActions(ctx, q, sessionID, roomID, 200)
	if err != nil {
		return CbtProctoringLiveSummary{}, err
	}
	return buildProctoringLiveSummary(sessionID, roomID, participants, events, actions), nil
}

func listProctorActions(ctx context.Context, q cbtProctoringStore, sessionID, roomID pgtype.UUID, limit int32) ([]db.CbtProctorAction, error) {
	if roomID.Valid {
		return q.ListCbtProctorActionsByRoom(ctx, db.ListCbtProctorActionsByRoomParams{SessionID: sessionID, RoomID: roomID, LimitCount: limit})
	}
	return q.ListCbtProctorActionsBySession(ctx, db.ListCbtProctorActionsBySessionParams{SessionID: sessionID, LimitCount: limit})
}

func buildProctoringLiveSummary(sessionID, roomID pgtype.UUID, rows []db.GetSessionProctoringStatusRow, events []CbtProctoringEventDTO, actions []db.CbtProctorAction) CbtProctoringLiveSummary {
	latestByParticipant := map[string]*CbtProctoringEventDTO{}
	unackedByParticipant := map[string]int{}
	alarmEvents := make([]CbtProctoringEventDTO, 0)
	for i := range events {
		event := events[i]
		if _, exists := latestByParticipant[event.ParticipantID]; !exists {
			copyEvent := event
			latestByParticipant[event.ParticipantID] = &copyEvent
		}
		if event.AcknowledgedAt == "" && proctorEventNeedsAction(event) {
			unackedByParticipant[event.ParticipantID]++
			alarmEvents = append(alarmEvents, event)
		}
	}
	participants := make([]CbtProctoringParticipantDTO, 0, len(rows))
	counts := map[string]int{"normal": 0, "warning": 0, "high": 0, "critical": 0, "technical": 0, "offline": 0, "submitted": 0, "pending_sync": 0}
	roomMap := map[string]*CbtProctoringRoomSummaryDTO{}
	for _, row := range rows {
		connection := proctorConnectionStatus(row.SubmittedAt, row.LastHeartbeat)
		syncStatus := proctorSyncStatus(row.SyncState, row.PendingAnswerCount)
		latest := latestByParticipant[pgUUIDString(row.ParticipantID)]
		needsAction, priority, reason := proctorParticipantNeedsAction(row, connection, syncStatus, unackedByParticipant[pgUUIDString(row.ParticipantID)], latest)
		dto := CbtProctoringParticipantDTO{
			ParticipantID:       pgUUIDString(row.ParticipantID),
			StudentID:           pgUUIDString(row.StudentID),
			NIS:                 row.Nis,
			Nama:                row.Nama,
			RoomID:              pgUUIDString(row.RoomID),
			RoomName:            row.RoomName,
			SeatNo:              row.SeatNo,
			SubmittedAt:         row.SubmittedAt,
			LastHeartbeat:       row.LastHeartbeat,
			ConnectionStatus:    connection,
			SyncStatus:          syncStatus,
			RiskLevel:           row.RiskLevel,
			RiskScore:           row.RiskScore,
			ViolationCount:      row.ViolationCount,
			LockedAt:            row.LockedAt,
			LockedReason:        row.LockedReason,
			NeedsAction:         needsAction,
			ActionPriority:      priority,
			NeedsActionReason:   reason,
			LatestEvent:         latest,
			UnacknowledgedCount: unackedByParticipant[pgUUIDString(row.ParticipantID)],
			PendingAnswerCount:  row.PendingAnswerCount,
			LastLocalSaveAt:     row.LastLocalSaveAt,
			LastSyncedAt:        row.LastSyncedAt,
			AnsweredCount:       row.AnsweredCount,
			Score:               row.Score,
			AppSwitchCount:      row.AppSwitchCount,
			ScreenshotAttempt:   row.ScreenshotAttempt,
			SuspiciousFlag:      row.SuspiciousFlag,
		}
		participants = append(participants, dto)
		counts[row.RiskLevel]++
		if row.LockedAt.Valid || row.RiskLevel == "locked" {
			counts["critical"]++
		}
		if connection == "terputus" || connection == "terlambat" {
			counts["offline"]++
		}
		if row.SubmittedAt.Valid {
			counts["submitted"]++
		}
		if syncStatus == "belum_sinkron" || syncStatus == "tertahan" {
			counts["pending_sync"]++
		}
		key := pgUUIDString(row.RoomID)
		if key == "" {
			key = "unassigned"
		}
		room := roomMap[key]
		if room == nil {
			room = &CbtProctoringRoomSummaryDTO{RoomID: pgUUIDString(row.RoomID), RoomName: firstNonEmpty(row.RoomName, "Tanpa ruang")}
			roomMap[key] = room
		}
		room.ParticipantCount++
		if row.RiskLevel == "normal" {
			room.NormalCount++
		}
		if row.RiskLevel == "warning" {
			room.WarningCount++
		}
		if row.RiskLevel == "high" {
			room.HighCount++
		}
		if row.LockedAt.Valid || row.RiskLevel == "locked" {
			room.CriticalCount++
		}
		if connection == "terputus" || connection == "terlambat" {
			room.OfflineCount++
		}
		if row.SubmittedAt.Valid {
			room.SubmittedCount++
		}
		if syncStatus == "belum_sinkron" || syncStatus == "tertahan" {
			room.PendingSyncCount++
		}
		if needsAction {
			room.NeedsActionCount++
		}
	}
	sort.Slice(participants, func(i, j int) bool {
		if participants[i].ActionPriority == participants[j].ActionPriority {
			return participants[i].Nama < participants[j].Nama
		}
		return participants[i].ActionPriority > participants[j].ActionPriority
	})
	rooms := make([]CbtProctoringRoomSummaryDTO, 0, len(roomMap))
	for _, room := range roomMap {
		rooms = append(rooms, *room)
	}
	sort.Slice(rooms, func(i, j int) bool {
		if rooms[i].NeedsActionCount == rooms[j].NeedsActionCount {
			return rooms[i].RoomName < rooms[j].RoomName
		}
		return rooms[i].NeedsActionCount > rooms[j].NeedsActionCount
	})
	return CbtProctoringLiveSummary{
		SessionID:           pgUUIDString(sessionID),
		RoomID:              pgUUIDString(roomID),
		GeneratedAt:         time.Now().Format(time.RFC3339),
		Counts:              counts,
		Rooms:               rooms,
		Participants:        participants,
		LatestEvents:        firstProctorEvents(events, 50),
		AlarmEvents:         firstProctorEvents(alarmEvents, 50),
		UnacknowledgedCount: len(alarmEvents),
		Actions:             actions,
	}
}

func normalizeProctorActionType(raw string) string {
	clean := strings.TrimSpace(strings.ToLower(raw))
	clean = strings.ReplaceAll(clean, "-", "_")
	switch clean {
	case "warn", "warning":
		return "warn_student"
	case "unlock", "unlock_access":
		return "unlock_access"
	case "hold", "lock", "hold_access":
		return "hold_access"
	case "reset", "reset_access", "reset_device", "reset_device_binding":
		return "reset_device_binding"
	case "force_submit":
		return "force_submit"
	case "technical", "mark_technical_issue":
		return "mark_technical_issue"
	case "incident", "mark_incident":
		return "mark_incident"
	case "escalate", "escalate_to_committee":
		return "escalate_to_committee"
	case "clear", "clear_after_check":
		return "clear_after_check"
	default:
		return ""
	}
}

func proctorActionRequiresNotes(actionType string) bool {
	switch actionType {
	case "unlock_access", "reset_device_binding", "force_submit", "hold_access", "escalate_to_committee":
		return true
	default:
		return false
	}
}

func proctorActionEventType(actionType string) string {
	switch actionType {
	case "unlock_access":
		return "proctor_unlock"
	case "reset_device_binding":
		return "proctor_reset_access"
	case "force_submit":
		return "proctor_force_submit"
	case "hold_access":
		return "proctor_hold_access"
	case "mark_technical_issue":
		return "proctor_mark_technical"
	case "warn_student":
		return "proctor_warning"
	case "escalate_to_committee":
		return "proctor_escalate"
	case "clear_after_check":
		return "proctor_clear_after_check"
	default:
		return "proctor_incident_action"
	}
}

func proctorActionSeverity(actionType string) ProctorSeverity {
	switch actionType {
	case "force_submit", "hold_access", "escalate_to_committee":
		return ProctorSeverityCritical
	case "mark_technical_issue":
		return ProctorSeverityTechnical
	case "warn_student", "reset_device_binding":
		return ProctorSeverityWarning
	default:
		return ProctorSeverityInfo
	}
}

func proctorEventDTOsFromSessionRows(rows []db.ListCbtProctorEventsBySessionRow) []CbtProctoringEventDTO {
	items := make([]CbtProctoringEventDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, proctorEventDTO(row.ID, row.ParticipantID, row.StudentID, row.SessionID, row.RoomID, row.Nis, row.Nama, row.RoomName, row.EventType, row.Severity, row.Category, row.RiskDelta, row.EventData, row.CreatedAt, row.AcknowledgedAt, row.AcknowledgedBy, row.AcknowledgeNote, row.RequiresNote, row.ActorUsernameSnapshot))
	}
	return items
}

func proctorEventDTOsFromRoomRows(rows []db.ListCbtProctorEventsByRoomRow) []CbtProctoringEventDTO {
	items := make([]CbtProctoringEventDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, proctorEventDTO(row.ID, row.ParticipantID, row.StudentID, row.SessionID, row.RoomID, row.Nis, row.Nama, row.RoomName, row.EventType, row.Severity, row.Category, row.RiskDelta, row.EventData, row.CreatedAt, row.AcknowledgedAt, row.AcknowledgedBy, row.AcknowledgeNote, row.RequiresNote, row.ActorUsernameSnapshot))
	}
	return items
}

func proctorEventDTO(id, participantID, studentID, sessionID, roomID pgtype.UUID, nis, nama, roomName, eventType, severity, category string, riskDelta int32, eventData []byte, createdAt, acknowledgedAt pgtype.Timestamptz, acknowledgedBy pgtype.UUID, acknowledgeNote string, requiresNote bool, actor string) CbtProctoringEventDTO {
	data := map[string]any{}
	_ = json.Unmarshal(eventData, &data)
	labelID := firstNonEmpty(stringFromAny(data["label_id"]), eventType)
	messageID := firstNonEmpty(stringFromAny(data["message_id"]), eventType)
	audioKey := firstNonEmpty(stringFromAny(data["audio_key"]), severity)
	return CbtProctoringEventDTO{
		ID:                    pgUUIDString(id),
		LegacyID:              pgUUIDString(id),
		ParticipantID:         pgUUIDString(participantID),
		StudentID:             pgUUIDString(studentID),
		NIS:                   nis,
		Nama:                  nama,
		ParticipantName:       nama,
		SessionID:             pgUUIDString(sessionID),
		RoomID:                pgUUIDString(roomID),
		RoomName:              roomName,
		EventType:             eventType,
		Severity:              firstNonEmpty(severity, "info"),
		Category:              firstNonEmpty(category, "timeline"),
		RiskDelta:             riskDelta,
		LabelID:               labelID,
		MessageID:             messageID,
		AudioKey:              audioKey,
		RequiresNote:          requiresNote,
		IsMassTechnicalIssue:  category == "technical_mass" || eventType == "offline_mass",
		AcknowledgedAt:        proctorTimeString(acknowledgedAt),
		AcknowledgedBy:        pgUUIDString(acknowledgedBy),
		AcknowledgeNote:       acknowledgeNote,
		ActorUsernameSnapshot: actor,
		EventData:             data,
		CreatedAt:             proctorTimeString(createdAt),
	}
}

func proctorConnectionStatus(submittedAt, lastHeartbeat pgtype.Timestamptz) string {
	if submittedAt.Valid {
		return "selesai"
	}
	if !lastHeartbeat.Valid {
		return "terputus"
	}
	minutes := time.Since(lastHeartbeat.Time).Minutes()
	if minutes <= 2 {
		return "online"
	}
	if minutes <= 7 {
		return "terlambat"
	}
	return "terputus"
}

func proctorSyncStatus(syncState string, pendingAnswerCount int32) string {
	switch syncState {
	case "synced":
		return "sinkron"
	case "pending":
		return "belum_sinkron"
	case "failed":
		return "tertahan"
	default:
		if pendingAnswerCount > 0 {
			return "belum_sinkron"
		}
		return "tidak_diketahui"
	}
}

func proctorParticipantNeedsAction(row db.GetSessionProctoringStatusRow, connection, syncStatus string, unacknowledged int, latest *CbtProctoringEventDTO) (bool, int, string) {
	if row.LockedAt.Valid || row.RiskLevel == "locked" {
		return true, 100, "Akses ditahan"
	}
	if latest != nil && latest.Severity == string(ProctorSeverityCritical) && latest.AcknowledgedAt == "" {
		return true, 95, "Peringatan kritis belum dicek"
	}
	if row.RiskLevel == "high" {
		return true, 80, "Risiko tinggi"
	}
	if latest != nil && (latest.Severity == string(ProctorSeverityMedium) || latest.Severity == string(ProctorSeverityTechnical)) && latest.AcknowledgedAt == "" {
		return true, 70, "Peringatan belum dicek"
	}
	if unacknowledged > 0 {
		return true, 60, "Ada peringatan belum dicek"
	}
	if syncStatus == "tertahan" || syncStatus == "belum_sinkron" {
		return true, 50, "Jawaban belum terkirim"
	}
	if connection == "terputus" {
		return true, 40, "Terputus"
	}
	if connection == "terlambat" {
		return true, 30, "Kontak terlambat"
	}
	if row.RiskLevel == "warning" {
		return true, 20, "Perlu perhatian"
	}
	return false, 0, ""
}

func proctorEventNeedsAction(event CbtProctoringEventDTO) bool {
	switch event.Severity {
	case string(ProctorSeverityMedium), string(ProctorSeverityCritical), string(ProctorSeverityTechnical):
		return true
	default:
		return false
	}
}

func firstProctorEvents(events []CbtProctoringEventDTO, limit int) []CbtProctoringEventDTO {
	if len(events) <= limit {
		return events
	}
	return events[:limit]
}

func proctorTimeString(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.RFC3339)
}
