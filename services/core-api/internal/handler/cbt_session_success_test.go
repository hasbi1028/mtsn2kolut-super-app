package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeCbtSessionService struct {
	checkSessionID pgtype.UUID
	checkTeacherID pgtype.UUID
	checkAllowed   bool
	checkErr       error

	hasParticipantSessionID pgtype.UUID
	hasParticipantID        pgtype.UUID
	hasParticipantTeacherID pgtype.UUID
	hasParticipant          bool
	hasParticipantSet       bool
	hasParticipantErr       error

	hasRoomSessionID pgtype.UUID
	hasRoomID        pgtype.UUID
	hasRoom          bool
	hasRoomSet       bool
	hasRoomErr       error

	hasAnswerSessionID pgtype.UUID
	hasAnswerID        pgtype.UUID
	hasAnswerTeacherID pgtype.UUID
	hasAnswer          bool
	hasAnswerSet       bool
	hasAnswerErr       error

	listCalled bool
	listErr    error

	getID  pgtype.UUID
	getRow db.GetCbtExamSessionRow
	getErr error

	createInput service.CreateCbtSessionInput
	createErr   error

	updateStatusID     pgtype.UUID
	updateStatusStatus db.CbtSessionStatusEnum
	updateStatusErr    error

	updateScheduleID    pgtype.UUID
	updateScheduleStart pgtype.Timestamptz
	updateScheduleEnd   pgtype.Timestamptz
	updateScheduleErr   error

	listAuditID     pgtype.UUID
	listAuditLimit  int32
	listAuditOffset int32
	listAuditRows   []db.ListEntityAuditLogsRow
	listAuditErr    error

	deleteID  pgtype.UUID
	deleteErr error

	listParticipantsSessionID pgtype.UUID
	listParticipantsTeacherID pgtype.UUID
	listParticipantsRows      []db.ListCbtExamParticipantsRow
	listParticipantsErr       error

	enrollClassSessionID pgtype.UUID
	enrollClassID        pgtype.UUID
	enrollClassErr       error

	enrollGradeSessionID pgtype.UUID
	enrollGradeLevel     string
	enrollGradeErr       error

	enrollSchoolSessionID pgtype.UUID
	enrollSchoolErr       error

	generateTokensSessionID pgtype.UUID
	generateTokensErr       error

	regenerateParticipantID  pgtype.UUID
	regenerateErr            error
	resetParticipantID       pgtype.UUID
	resetActor               string
	resetErr                 error
	incidentParticipantID    pgtype.UUID
	incidentEventID          string
	incidentAction           string
	incidentActor            string
	incidentNotes            string
	incidentErr              error
	commandParticipantID     pgtype.UUID
	commandType              string
	commandMessage           string
	commandActor             string
	commandErr               error
	forceSubmitSessionID     pgtype.UUID
	forceSubmitParticipantID pgtype.UUID
	forceSubmitActor         string
	forceSubmitErr           error

	assignSeatParticipantID pgtype.UUID
	assignSeatRoomID        pgtype.UUID
	assignSeatNo            int32
	assignSeatErr           error

	autoAssignSessionID pgtype.UUID
	autoAssignErr       error

	listRoomsSessionID pgtype.UUID
	listRoomsErr       error

	createRoomSessionID pgtype.UUID
	createRoomName      string
	createRoomCapacity  int32
	createRoomErr       error

	deleteRoomID  pgtype.UUID
	deleteRoomErr error

	shuffleSessionID pgtype.UUID
	shuffleErr       error

	proctoringSessionID        pgtype.UUID
	roomDashboardID            pgtype.UUID
	proctorRoomsEmployeeID     pgtype.UUID
	proctorRoomsIncludeAll     bool
	roomProctorSessionID       pgtype.UUID
	roomProctorRoomID          pgtype.UUID
	roomProctorEmployeeID      pgtype.UUID
	roomProctorAllowed         bool
	roomParticipantSessionID   pgtype.UUID
	roomParticipantRoomID      pgtype.UUID
	roomParticipantID          pgtype.UUID
	roomParticipantAllowed     bool
	roomParticipantErr         error
	roomProctoringSessionID    pgtype.UUID
	roomProctoringRoomID       pgtype.UUID
	roomEventsSessionID        pgtype.UUID
	roomEventsRoomID           pgtype.UUID
	roomEventsParticipantID    pgtype.UUID
	roomEventsLimit            int32
	proctoringErr              error
	liveSummarySessionID       pgtype.UUID
	liveSummaryRoomID          pgtype.UUID
	liveSummary                service.CbtProctoringLiveSummary
	liveSummaryErr             error
	participantScopeSessionID  pgtype.UUID
	participantScopeID         pgtype.UUID
	participantScope           db.GetCbtParticipantProctorScopeRow
	participantScopeErr        error
	proctorEventScopeID        pgtype.UUID
	proctorEventScope          db.GetCbtProctorEventScopeRow
	proctorEventScopeErr       error
	ackEventID                 pgtype.UUID
	ackActor                   service.CbtProctorActor
	ackNote                    string
	ackRow                     db.AcknowledgeCbtProctorEventRow
	ackErr                     error
	proctorActionInput         service.CbtProctorActionInput
	proctorActionActor         service.CbtProctorActor
	proctorActionResult        service.CbtProctorActionResult
	proctorActionErr           error
	listRoomProctorsRoomID     pgtype.UUID
	replaceProctorsRoomID      pgtype.UUID
	replaceProctorsAssignedBy  pgtype.UUID
	replaceProctorsPrimaryID   pgtype.UUID
	replaceProctorsEmployeeIDs []pgtype.UUID
	roomReadinessSessionID     pgtype.UUID
	handoverRoomID             pgtype.UUID
	operationalRecapID         pgtype.UUID
	saveHandoverRoomID         pgtype.UUID
	saveHandoverUpdatedBy      pgtype.UUID
	saveHandoverInput          service.SaveCbtRoomHandoverInput
	saveHandoverErr            error
	lockHandoverRoomID         pgtype.UUID
	lockHandoverLockedBy       pgtype.UUID
	lockHandoverErr            error

	flagParticipantID pgtype.UUID
	flagValue         bool
	flagErr           error

	ungradedSessionID pgtype.UUID
	ungradedTeacherID pgtype.UUID
	ungradedErr       error

	gradeAnswerID pgtype.UUID
	gradeScore    float64
	gradeBy       string
	gradeErr      error

	recordParticipantID pgtype.UUID
	recordQuestionID    pgtype.UUID
	recordAnswer        string
	recordErr           error

	scoreSessionID pgtype.UUID
	scoreErr       error

	listByTeacherID  pgtype.UUID
	listByTeacherErr error

	resultsByTeacherSessionID pgtype.UUID
	resultsByTeacherID        pgtype.UUID
	resultsByTeacherErr       error

	getResultsSessionID pgtype.UUID
	getResultsErr       error

	participantAnswersID  pgtype.UUID
	participantAnswersErr error

	finalizeOverdueSessionID pgtype.UUID
	finalizeOverdueResult    service.CbtFinalizeOverdueResult
	finalizeOverdueErr       error

	gradeSyncPreflightSessionID pgtype.UUID
	gradeSyncPreflightRow       db.GetCbtSessionGradeSyncPreflightRow
	gradeSyncPreflightErr       error
	remedialSessionID           pgtype.UUID
	remedialThreshold           float64
	remedialRows                []db.ListCbtSessionRemedialCandidatesRow
	remedialErr                 error

	itemAnalysisSessionID pgtype.UUID
	itemAnalysisRows      []db.GetSessionItemAnalysisRow
	itemAnalysisErr       error
}

type fakeCbtSessionAuditWriter struct {
	entries []db.CreateAuditLogParams
}

func (f *fakeCbtSessionAuditWriter) CreateAuditLog(_ context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.entries = append(f.entries, arg)
	return db.AuditLog{}, nil
}

func mustAuditMetadataMap(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	meta := map[string]any{}
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("json.Unmarshal(audit metadata) error = %v", err)
	}
	return meta
}

func TestNewCbtSessionAcceptsOptionalAuditWriter(t *testing.T) {
	writer := &fakeCbtSessionAuditWriter{}
	h := NewCbtSession(&service.CbtSession{}, writer)
	if h == nil {
		t.Fatal("NewCbtSession() = nil")
	}
	if h.audit != writer {
		t.Fatalf("NewCbtSession() audit = %T, want provided writer", h.audit)
	}
}

func TestCbtSessionAuditEventBranches(t *testing.T) {
	userID := handlerTestUUID(250)
	ctx := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/score", nil), jwt.MapClaims{
		"sub":  userID.String(),
		"uid":  userID.String(),
		"usr":  "operator.cbt",
		"ssid": "session-1",
	}).Context()

	(&CbtSession{}).auditEvent(ctx, "CBT_SESSION_SCORE", "cbt_session", "session-1", map[string]any{"ok": true})

	writer := &fakeCbtSessionAuditWriter{}
	h := &CbtSession{audit: writer}
	h.auditEvent(context.Background(), "CBT_SESSION_SCORE", "cbt_session", "session-1", nil)
	h.auditEvent(withClaims(httptest.NewRequest(http.MethodPost, "/score", nil), jwt.MapClaims{"sub": "not-a-uuid"}).Context(), "CBT_SESSION_SCORE", "cbt_session", "session-1", nil)
	h.auditEvent(ctx, "CBT_SESSION_SCORE", "cbt_session", "session-1", map[string]any{"bad": func() {}})
	if len(writer.entries) != 0 {
		t.Fatalf("audit entries = %d, want none for missing/invalid/unmarshalable metadata branches", len(writer.entries))
	}

	h.auditEvent(ctx, "CBT_SESSION_SCORE", "cbt_session", "session-1", map[string]any{"scored_by": "operator.cbt"})
	if len(writer.entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(writer.entries))
	}
	got := writer.entries[0]
	if got.UserID != userID || got.Action != "CBT_SESSION_SCORE" || got.EntityType != "cbt_session" || got.EntityID != "session-1" {
		t.Fatalf("audit arg = %+v, want score audit for user/session", got)
	}
	meta := mustAuditMetadataMap(t, got.Metadata)
	if meta["username"] != "operator.cbt" || meta["session_id"] != "session-1" || meta["scored_by"] != "operator.cbt" {
		t.Fatalf("audit metadata = %+v, want merged actor metadata", meta)
	}
}

func TestCbtAuditClaimStringBranches(t *testing.T) {
	if got := cbtAuditClaimString(context.Background(), "ssid"); got != "" {
		t.Fatalf("cbtAuditClaimString(no claims) = %q, want empty", got)
	}
	ctx := withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/sessions", nil), jwt.MapClaims{
		"ssid": "session-1",
		"uid":  123,
	}).Context()
	if got := cbtAuditClaimString(ctx, "ssid"); got != "session-1" {
		t.Fatalf("cbtAuditClaimString(ssid) = %q, want session-1", got)
	}
	if got := cbtAuditClaimString(ctx, "uid"); got != "" {
		t.Fatalf("cbtAuditClaimString(non-string uid) = %q, want empty", got)
	}
}

func (f *fakeCbtSessionService) CheckTeacherAccess(_ context.Context, sessionID, teacherEmployeeID pgtype.UUID) (bool, error) {
	f.checkSessionID = sessionID
	f.checkTeacherID = teacherEmployeeID
	if f.checkErr != nil {
		return false, f.checkErr
	}
	return f.checkAllowed, nil
}

func (f *fakeCbtSessionService) HasParticipant(_ context.Context, sessionID, participantID pgtype.UUID) (bool, error) {
	f.hasParticipantSessionID = sessionID
	f.hasParticipantID = participantID
	if f.hasParticipantErr != nil {
		return false, f.hasParticipantErr
	}
	if f.hasParticipantSet {
		return f.hasParticipant, nil
	}
	return true, nil
}

func (f *fakeCbtSessionService) HasParticipantByTeacher(_ context.Context, sessionID, participantID, teacherEmployeeID pgtype.UUID) (bool, error) {
	f.hasParticipantSessionID = sessionID
	f.hasParticipantID = participantID
	f.hasParticipantTeacherID = teacherEmployeeID
	if f.hasParticipantErr != nil {
		return false, f.hasParticipantErr
	}
	if f.hasParticipantSet {
		return f.hasParticipant, nil
	}
	return true, nil
}

func (f *fakeCbtSessionService) HasRoom(_ context.Context, sessionID, roomID pgtype.UUID) (bool, error) {
	f.hasRoomSessionID = sessionID
	f.hasRoomID = roomID
	if f.hasRoomErr != nil {
		return false, f.hasRoomErr
	}
	if f.hasRoomSet {
		return f.hasRoom, nil
	}
	return true, nil
}

func (f *fakeCbtSessionService) HasAnswer(_ context.Context, sessionID, answerID pgtype.UUID) (bool, error) {
	f.hasAnswerSessionID = sessionID
	f.hasAnswerID = answerID
	if f.hasAnswerErr != nil {
		return false, f.hasAnswerErr
	}
	if f.hasAnswerSet {
		return f.hasAnswer, nil
	}
	return true, nil
}

func (f *fakeCbtSessionService) HasAnswerByTeacher(_ context.Context, sessionID, answerID, teacherEmployeeID pgtype.UUID) (bool, error) {
	f.hasAnswerSessionID = sessionID
	f.hasAnswerID = answerID
	f.hasAnswerTeacherID = teacherEmployeeID
	if f.hasAnswerErr != nil {
		return false, f.hasAnswerErr
	}
	if f.hasAnswerSet {
		return f.hasAnswer, nil
	}
	return true, nil
}

func (f *fakeCbtSessionService) List(context.Context) ([]db.ListCbtExamSessionsRow, error) {
	f.listCalled = true
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListCbtExamSessionsRow{}, nil
}

func (f *fakeCbtSessionService) Get(_ context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error) {
	f.getID = id
	if f.getErr != nil {
		return db.GetCbtExamSessionRow{}, f.getErr
	}
	return f.getRow, nil
}

func (f *fakeCbtSessionService) Create(_ context.Context, in service.CreateCbtSessionInput) (db.CbtExamSession, error) {
	f.createInput = in
	if f.createErr != nil {
		return db.CbtExamSession{}, f.createErr
	}
	return db.CbtExamSession{ID: handlerTestUUID(101), Title: in.Title, Status: in.Status}, nil
}

func (f *fakeCbtSessionService) UpdateStatus(_ context.Context, id pgtype.UUID, status db.CbtSessionStatusEnum) (db.CbtExamSession, error) {
	f.updateStatusID = id
	f.updateStatusStatus = status
	if f.updateStatusErr != nil {
		return db.CbtExamSession{}, f.updateStatusErr
	}
	return db.CbtExamSession{ID: id, Status: status}, nil
}

func (f *fakeCbtSessionService) UpdateSchedule(_ context.Context, id pgtype.UUID, start, end pgtype.Timestamptz) (service.UpdateCbtSessionScheduleResult, error) {
	f.updateScheduleID = id
	f.updateScheduleStart = start
	f.updateScheduleEnd = end
	if f.updateScheduleErr != nil {
		return service.UpdateCbtSessionScheduleResult{}, f.updateScheduleErr
	}
	return service.UpdateCbtSessionScheduleResult{
		Before: db.GetCbtExamSessionRow{
			ID:             id,
			Title:          "Sesi Lama",
			Status:         db.CbtSessionStatusEnumScheduled,
			ScheduledStart: pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC), Valid: true},
			ScheduledEnd:   pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC), Valid: true},
		},
		Session: db.CbtExamSession{ID: id, Title: "Sesi Baru", Status: db.CbtSessionStatusEnumScheduled, ScheduledStart: start, ScheduledEnd: end},
	}, nil
}

func (f *fakeCbtSessionService) ListAuditLogs(_ context.Context, id pgtype.UUID, limit, offset int32) ([]db.ListEntityAuditLogsRow, error) {
	f.listAuditID = id
	f.listAuditLimit = limit
	f.listAuditOffset = offset
	return f.listAuditRows, f.listAuditErr
}

func (f *fakeCbtSessionService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeCbtSessionService) ListParticipants(_ context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error) {
	f.listParticipantsSessionID = sessionID
	if f.listParticipantsErr != nil {
		return nil, f.listParticipantsErr
	}
	return f.listParticipantsRows, nil
}

func (f *fakeCbtSessionService) ListParticipantsByTeacher(_ context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error) {
	f.listParticipantsSessionID = sessionID
	f.listParticipantsTeacherID = teacherEmployeeID
	if f.listParticipantsErr != nil {
		return nil, f.listParticipantsErr
	}
	return f.listParticipantsRows, nil
}

func (f *fakeCbtSessionService) EnrollClass(_ context.Context, sessionID, classID pgtype.UUID) error {
	f.enrollClassSessionID = sessionID
	f.enrollClassID = classID
	return f.enrollClassErr
}

func (f *fakeCbtSessionService) EnrollGrade(_ context.Context, sessionID pgtype.UUID, level string) error {
	f.enrollGradeSessionID = sessionID
	f.enrollGradeLevel = level
	return f.enrollGradeErr
}

func (f *fakeCbtSessionService) EnrollSchool(_ context.Context, sessionID pgtype.UUID) error {
	f.enrollSchoolSessionID = sessionID
	return f.enrollSchoolErr
}

func (f *fakeCbtSessionService) GenerateTokens(_ context.Context, sessionID pgtype.UUID) error {
	f.generateTokensSessionID = sessionID
	return f.generateTokensErr
}

func (f *fakeCbtSessionService) RegenerateToken(_ context.Context, participantID pgtype.UUID) (db.RegenerateParticipantTokenRow, error) {
	f.regenerateParticipantID = participantID
	if f.regenerateErr != nil {
		return db.RegenerateParticipantTokenRow{}, f.regenerateErr
	}
	return db.RegenerateParticipantTokenRow{ID: participantID, Token: "TOKEN"}, nil
}

func (f *fakeCbtSessionService) ResetParticipantRuntimeAccess(_ context.Context, participantID pgtype.UUID, actor string) error {
	f.resetParticipantID = participantID
	f.resetActor = actor
	return f.resetErr
}

func (f *fakeCbtSessionService) UnlockParticipantAntiCheat(_ context.Context, participantID pgtype.UUID, actor, notes string) (db.UnlockParticipantAntiCheatRow, error) {
	f.resetParticipantID = participantID
	f.resetActor = actor
	return db.UnlockParticipantAntiCheatRow{ID: participantID}, f.resetErr
}

func (f *fakeCbtSessionService) AcknowledgeProctorEvent(_ context.Context, participantID pgtype.UUID, eventID, actor, notes string) error {
	f.incidentParticipantID = participantID
	f.incidentEventID = eventID
	f.incidentActor = actor
	f.incidentNotes = notes
	return f.incidentErr
}

func (f *fakeCbtSessionService) RecordIncidentAction(_ context.Context, participantID pgtype.UUID, eventID, action, actor, notes string) error {
	f.incidentParticipantID = participantID
	f.incidentEventID = eventID
	f.incidentAction = action
	f.incidentActor = actor
	f.incidentNotes = notes
	return f.incidentErr
}

func (f *fakeCbtSessionService) SendParticipantCommand(_ context.Context, participantID pgtype.UUID, commandType, message, actor string) error {
	f.commandParticipantID = participantID
	f.commandType = commandType
	f.commandMessage = message
	f.commandActor = actor
	return f.commandErr
}

func (f *fakeCbtSessionService) ListParticipantEvents(_ context.Context, sessionID, participantID pgtype.UUID, limit int32) ([]db.ListSessionParticipantEventsRow, error) {
	f.roomEventsSessionID = sessionID
	f.roomEventsParticipantID = participantID
	f.roomEventsLimit = limit
	return []db.ListSessionParticipantEventsRow{}, nil
}

func (f *fakeCbtSessionService) ForceSubmitParticipant(_ context.Context, sessionID, participantID pgtype.UUID, actor string) (db.ForceSubmitParticipantRow, error) {
	f.forceSubmitSessionID = sessionID
	f.forceSubmitParticipantID = participantID
	f.forceSubmitActor = actor
	if f.forceSubmitErr != nil {
		return db.ForceSubmitParticipantRow{}, f.forceSubmitErr
	}
	return db.ForceSubmitParticipantRow{ID: participantID}, nil
}

func (f *fakeCbtSessionService) AssignSeat(_ context.Context, participantID, roomID pgtype.UUID, seatNo int32) error {
	f.assignSeatParticipantID = participantID
	f.assignSeatRoomID = roomID
	f.assignSeatNo = seatNo
	return f.assignSeatErr
}

func (f *fakeCbtSessionService) AutoAssignSeats(_ context.Context, sessionID pgtype.UUID) error {
	f.autoAssignSessionID = sessionID
	return f.autoAssignErr
}

func (f *fakeCbtSessionService) ListRooms(_ context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error) {
	f.listRoomsSessionID = sessionID
	if f.listRoomsErr != nil {
		return nil, f.listRoomsErr
	}
	return []db.ListCbtExamRoomsRow{}, nil
}

func (f *fakeCbtSessionService) CreateRoom(_ context.Context, sessionID pgtype.UUID, roomName string, capacity int32) (db.CbtExamRoom, error) {
	f.createRoomSessionID = sessionID
	f.createRoomName = roomName
	f.createRoomCapacity = capacity
	if f.createRoomErr != nil {
		return db.CbtExamRoom{}, f.createRoomErr
	}
	return db.CbtExamRoom{ID: handlerTestUUID(102), SessionID: sessionID, RoomName: roomName, Capacity: capacity}, nil
}

func (f *fakeCbtSessionService) DeleteRoom(_ context.Context, roomID pgtype.UUID) error {
	f.deleteRoomID = roomID
	return f.deleteRoomErr
}

func (f *fakeCbtSessionService) ShuffleRooms(_ context.Context, sessionID pgtype.UUID) error {
	f.shuffleSessionID = sessionID
	return f.shuffleErr
}

func (f *fakeCbtSessionService) GetProctoringStatus(_ context.Context, sessionID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error) {
	f.proctoringSessionID = sessionID
	if f.proctoringErr != nil {
		return nil, f.proctoringErr
	}
	return []db.GetSessionProctoringStatusRow{}, nil
}

func (f *fakeCbtSessionService) GetProctoringLiveSummary(_ context.Context, sessionID, roomID pgtype.UUID) (service.CbtProctoringLiveSummary, error) {
	f.liveSummarySessionID = sessionID
	f.liveSummaryRoomID = roomID
	if f.liveSummaryErr != nil {
		return service.CbtProctoringLiveSummary{}, f.liveSummaryErr
	}
	if f.liveSummary.SessionID != "" || f.liveSummary.RoomID != "" || len(f.liveSummary.Participants) > 0 || len(f.liveSummary.LatestEvents) > 0 || len(f.liveSummary.Actions) > 0 {
		return f.liveSummary, nil
	}
	return service.CbtProctoringLiveSummary{SessionID: pgUUIDString(sessionID), RoomID: pgUUIDString(roomID), Counts: map[string]int{"participants": 1}}, nil
}

func (f *fakeCbtSessionService) GetParticipantProctorScope(_ context.Context, sessionID, participantID pgtype.UUID) (db.GetCbtParticipantProctorScopeRow, error) {
	f.participantScopeSessionID = sessionID
	f.participantScopeID = participantID
	if f.participantScopeErr != nil {
		return db.GetCbtParticipantProctorScopeRow{}, f.participantScopeErr
	}
	if f.participantScope.ParticipantID.Valid {
		return f.participantScope, nil
	}
	return db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: f.roomProctorRoomID, Nama: "Peserta"}, nil
}

func (f *fakeCbtSessionService) GetProctorEventScope(_ context.Context, eventID pgtype.UUID) (db.GetCbtProctorEventScopeRow, error) {
	f.proctorEventScopeID = eventID
	if f.proctorEventScopeErr != nil {
		return db.GetCbtProctorEventScopeRow{}, f.proctorEventScopeErr
	}
	if f.proctorEventScope.ID.Valid {
		return f.proctorEventScope, nil
	}
	return db.GetCbtProctorEventScopeRow{ID: eventID, ParticipantID: f.participantScopeID, SessionID: f.proctoringSessionID, RoomID: f.roomProctorRoomID, Severity: "warning", Nama: "Peserta"}, nil
}

func (f *fakeCbtSessionService) AcknowledgeProctorEventByID(_ context.Context, eventID pgtype.UUID, actor service.CbtProctorActor, note string) (db.AcknowledgeCbtProctorEventRow, error) {
	f.ackEventID = eventID
	f.ackActor = actor
	f.ackNote = note
	if f.ackErr != nil {
		return db.AcknowledgeCbtProctorEventRow{}, f.ackErr
	}
	if f.ackRow.ID.Valid {
		return f.ackRow, nil
	}
	return db.AcknowledgeCbtProctorEventRow{ID: eventID, ParticipantID: f.participantScopeID, AcknowledgedBy: actor.UserID, AcknowledgeNote: note}, nil
}

func (f *fakeCbtSessionService) ExecuteProctorAction(_ context.Context, in service.CbtProctorActionInput, actor service.CbtProctorActor) (service.CbtProctorActionResult, error) {
	f.proctorActionInput = in
	f.proctorActionActor = actor
	if f.proctorActionErr != nil {
		return service.CbtProctorActionResult{}, f.proctorActionErr
	}
	if f.proctorActionResult.Action.ID.Valid {
		return f.proctorActionResult, nil
	}
	return service.CbtProctorActionResult{Action: db.CbtProctorAction{ID: handlerTestUUID(220), SessionID: in.SessionID, ParticipantID: in.ParticipantID, EventID: in.EventID, ActionType: in.ActionType, Reason: in.Reason, Notes: in.Notes, ActorUserID: actor.UserID, ActorUsernameSnapshot: actor.Username}}, nil
}

func (f *fakeCbtSessionService) GetRoomProctoringDashboard(_ context.Context, roomID pgtype.UUID) (db.GetCbtRoomProctorDashboardRow, error) {
	f.roomDashboardID = roomID
	return db.GetCbtRoomProctorDashboardRow{ID: roomID, RoomName: "Ruang 1"}, nil
}

func (f *fakeCbtSessionService) ListProctorRooms(_ context.Context, employeeID pgtype.UUID, includeAll bool) ([]db.ListCbtProctorRoomsRow, error) {
	f.proctorRoomsEmployeeID = employeeID
	f.proctorRoomsIncludeAll = includeAll
	return []db.ListCbtProctorRoomsRow{{ID: handlerTestUUID(103), RoomName: "Ruang 1"}}, nil
}

func (f *fakeCbtSessionService) HasRoomProctor(_ context.Context, sessionID, roomID, employeeID pgtype.UUID) (bool, error) {
	f.roomProctorSessionID = sessionID
	f.roomProctorRoomID = roomID
	f.roomProctorEmployeeID = employeeID
	return f.roomProctorAllowed, nil
}

func (f *fakeCbtSessionService) HasRoomParticipant(_ context.Context, sessionID, roomID, participantID pgtype.UUID) (bool, error) {
	f.roomParticipantSessionID = sessionID
	f.roomParticipantRoomID = roomID
	f.roomParticipantID = participantID
	if f.roomParticipantErr != nil {
		return false, f.roomParticipantErr
	}
	if f.roomParticipantAllowed {
		return true, nil
	}
	return false, nil
}

func (f *fakeCbtSessionService) GetProctoringStatusForRoom(_ context.Context, sessionID, roomID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error) {
	f.roomProctoringSessionID = sessionID
	f.roomProctoringRoomID = roomID
	return []db.GetSessionProctoringStatusRow{}, nil
}

func (f *fakeCbtSessionService) ListParticipantEventsForRoom(_ context.Context, sessionID, participantID, roomID pgtype.UUID, limit int32) ([]db.ListSessionParticipantEventsRow, error) {
	f.roomEventsSessionID = sessionID
	f.roomEventsRoomID = roomID
	return []db.ListSessionParticipantEventsRow{}, nil
}

func (f *fakeCbtSessionService) GetRoomHandover(_ context.Context, roomID pgtype.UUID) (db.GetCbtRoomHandoverRow, error) {
	f.handoverRoomID = roomID
	return db.GetCbtRoomHandoverRow{RoomID: roomID, RoomName: "Ruang 1"}, nil
}

func (f *fakeCbtSessionService) GetSessionOperationalRecap(_ context.Context, sessionID pgtype.UUID) (db.GetCbtSessionOperationalRecapRow, []db.ListCbtSessionRoomOperationalRecapRow, error) {
	f.operationalRecapID = sessionID
	return db.GetCbtSessionOperationalRecapRow{SessionID: sessionID, SessionTitle: "Sesi", RoomCount: 1}, []db.ListCbtSessionRoomOperationalRecapRow{{SessionID: sessionID, RoomName: "Ruang 1"}}, nil
}

func (f *fakeCbtSessionService) SaveRoomHandover(_ context.Context, roomID, updatedBy pgtype.UUID, in service.SaveCbtRoomHandoverInput) (db.CbtRoomHandover, error) {
	f.saveHandoverRoomID = roomID
	f.saveHandoverUpdatedBy = updatedBy
	f.saveHandoverInput = in
	if f.saveHandoverErr != nil {
		return db.CbtRoomHandover{}, f.saveHandoverErr
	}
	return db.CbtRoomHandover{ExamRoomID: roomID, AttendanceChecked: in.AttendanceChecked, IncidentNotes: in.IncidentNotes}, nil
}

func (f *fakeCbtSessionService) LockRoomHandover(_ context.Context, roomID, lockedBy pgtype.UUID) (db.CbtRoomHandover, error) {
	f.lockHandoverRoomID = roomID
	f.lockHandoverLockedBy = lockedBy
	if f.lockHandoverErr != nil {
		return db.CbtRoomHandover{}, f.lockHandoverErr
	}
	return db.CbtRoomHandover{ExamRoomID: roomID, LockedBy: lockedBy}, nil
}

func (f *fakeCbtSessionService) ListRoomProctors(_ context.Context, roomID pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error) {
	f.listRoomProctorsRoomID = roomID
	return []db.ListCbtRoomProctorsRow{{ExamRoomID: roomID, Nama: "Pengawas"}}, nil
}

func (f *fakeCbtSessionService) ReplaceRoomProctors(_ context.Context, roomID, assignedBy, primaryEmployeeID pgtype.UUID, employeeIDs []pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error) {
	f.replaceProctorsRoomID = roomID
	f.replaceProctorsAssignedBy = assignedBy
	f.replaceProctorsPrimaryID = primaryEmployeeID
	f.replaceProctorsEmployeeIDs = append([]pgtype.UUID(nil), employeeIDs...)
	return []db.ListCbtRoomProctorsRow{{ExamRoomID: roomID, EmployeeID: primaryEmployeeID}}, nil
}

func (f *fakeCbtSessionService) RoomReadiness(_ context.Context, sessionID pgtype.UUID) (db.GetCbtSessionRoomReadinessRow, error) {
	f.roomReadinessSessionID = sessionID
	return db.GetCbtSessionRoomReadinessRow{RoomCount: 1, ParticipantCount: 1, TotalCapacity: 30}, nil
}

func (f *fakeCbtSessionService) SetSuspiciousFlag(_ context.Context, participantID pgtype.UUID, flag bool) error {
	f.flagParticipantID = participantID
	f.flagValue = flag
	return f.flagErr
}

func (f *fakeCbtSessionService) ListUngradedEssays(_ context.Context, sessionID pgtype.UUID) ([]db.ListUngradedEssaysRow, error) {
	f.ungradedSessionID = sessionID
	if f.ungradedErr != nil {
		return nil, f.ungradedErr
	}
	return []db.ListUngradedEssaysRow{}, nil
}

func (f *fakeCbtSessionService) ListUngradedEssaysByTeacher(_ context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.ListUngradedEssaysRow, error) {
	f.ungradedSessionID = sessionID
	f.ungradedTeacherID = teacherEmployeeID
	if f.ungradedErr != nil {
		return nil, f.ungradedErr
	}
	return []db.ListUngradedEssaysRow{}, nil
}

func (f *fakeCbtSessionService) GradeEssay(_ context.Context, sessionID, answerID pgtype.UUID, manualScore float64, gradedBy string) error {
	f.gradeAnswerID = answerID
	f.scoreSessionID = sessionID
	f.gradeScore = manualScore
	f.gradeBy = gradedBy
	return f.gradeErr
}

func (f *fakeCbtSessionService) RecordAnswer(_ context.Context, participantID, questionID pgtype.UUID, answer string) error {
	f.recordParticipantID = participantID
	f.recordQuestionID = questionID
	f.recordAnswer = answer
	return f.recordErr
}

func (f *fakeCbtSessionService) ScoreSession(_ context.Context, sessionID pgtype.UUID) error {
	f.scoreSessionID = sessionID
	return f.scoreErr
}

func (f *fakeCbtSessionService) ListByTeacher(_ context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamSessionsByTeacherRow, error) {
	f.listByTeacherID = teacherEmployeeID
	if f.listByTeacherErr != nil {
		return nil, f.listByTeacherErr
	}
	return []db.ListCbtExamSessionsByTeacherRow{}, nil
}

func (f *fakeCbtSessionService) GetResultsByTeacher(_ context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.GetSessionResultsByTeacherRow, error) {
	f.resultsByTeacherSessionID = sessionID
	f.resultsByTeacherID = teacherEmployeeID
	if f.resultsByTeacherErr != nil {
		return nil, f.resultsByTeacherErr
	}
	return []db.GetSessionResultsByTeacherRow{}, nil
}

func (f *fakeCbtSessionService) GetResults(_ context.Context, sessionID pgtype.UUID) ([]db.GetSessionResultsRow, error) {
	f.getResultsSessionID = sessionID
	if f.getResultsErr != nil {
		return nil, f.getResultsErr
	}
	return []db.GetSessionResultsRow{}, nil
}

func (f *fakeCbtSessionService) GetParticipantAnswers(_ context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error) {
	f.participantAnswersID = participantID
	if f.participantAnswersErr != nil {
		return nil, f.participantAnswersErr
	}
	return []db.GetParticipantAnswersRow{}, nil
}

func (f *fakeCbtSessionService) FinalizeOverdue(_ context.Context, sessionID pgtype.UUID) (service.CbtFinalizeOverdueResult, error) {
	f.finalizeOverdueSessionID = sessionID
	if f.finalizeOverdueErr != nil {
		return service.CbtFinalizeOverdueResult{}, f.finalizeOverdueErr
	}
	if f.finalizeOverdueResult.SessionID != "" || f.finalizeOverdueResult.FinalizedCount != 0 {
		return f.finalizeOverdueResult, nil
	}
	return service.CbtFinalizeOverdueResult{SessionID: pgUUIDString(sessionID), FinalizedCount: 2}, nil
}

func (f *fakeCbtSessionService) GetGradeSyncPreflight(_ context.Context, sessionID pgtype.UUID) (db.GetCbtSessionGradeSyncPreflightRow, error) {
	f.gradeSyncPreflightSessionID = sessionID
	if f.gradeSyncPreflightErr != nil {
		return db.GetCbtSessionGradeSyncPreflightRow{}, f.gradeSyncPreflightErr
	}
	if f.gradeSyncPreflightRow.SessionID.Valid {
		return f.gradeSyncPreflightRow, nil
	}
	return db.GetCbtSessionGradeSyncPreflightRow{SessionID: sessionID, SessionTitle: "Sesi IPA", SessionStatus: db.CbtSessionStatusEnumFinished, ParticipantCount: 3, SubmittedCount: 3, ScoredCount: 2, MissingScoreCount: 1}, nil
}

func (f *fakeCbtSessionService) ListRemedialCandidates(_ context.Context, sessionID pgtype.UUID, threshold float64) ([]db.ListCbtSessionRemedialCandidatesRow, error) {
	f.remedialSessionID = sessionID
	f.remedialThreshold = threshold
	if f.remedialErr != nil {
		return nil, f.remedialErr
	}
	return f.remedialRows, nil
}

func (f *fakeCbtSessionService) GetItemAnalysis(_ context.Context, sessionID pgtype.UUID) ([]db.GetSessionItemAnalysisRow, error) {
	f.itemAnalysisSessionID = sessionID
	if f.itemAnalysisErr != nil {
		return nil, f.itemAnalysisErr
	}
	return f.itemAnalysisRows, nil
}

func TestCbtSessionFinalizeGradeSyncRemedialAndItemAnalysisHandlers(t *testing.T) {
	sessionID := handlerTestUUID(240)
	teacherID := handlerTestUUID(241)
	fake := &fakeCbtSessionService{
		checkAllowed: true,
		remedialRows: []db.ListCbtSessionRemedialCandidatesRow{{
			ParticipantID:  handlerTestUUID(242),
			StudentID:      handlerTestUUID(243),
			Nis:            "12345",
			Nama:           "Siswa Remedial",
			ClassCode:      "VII-A",
			ClassName:      "VII A",
			QuestionCount:  10,
			IncorrectCount: 4,
			BlankCount:     1,
			KdGaps:         []byte(`["KD 3.1"]`),
		}},
		itemAnalysisRows: []db.GetSessionItemAnalysisRow{{
			Position:            1,
			Points:              5,
			QuestionID:          handlerTestUUID(244),
			QuestionCode:        "IPA-001",
			QuestionText:        "Apa fungsi akar?",
			QuestionType:        "multiple_choice",
			Difficulty:          db.CbtQuestionDifficultyEnumMedium,
			AnswerKey:           "A",
			SubmittedCount:      5,
			AnsweredCount:       5,
			CorrectCount:        4,
			IncorrectCount:      1,
			DifficultyIndex:     0.8,
			DiscriminationIndex: 0.4,
			AnswerDistribution:  []byte(`{"A":4,"B":1}`),
		}},
	}
	h := &CbtSession{svc: fake}

	rec := httptest.NewRecorder()
	h.FinalizeOverdue(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/finalize-overdue", ""), "id", sessionID.String()))
	if rec.Code != http.StatusOK || fake.finalizeOverdueSessionID != sessionID || !strings.Contains(rec.Body.String(), `"finalized_count":2`) {
		t.Fatalf("FinalizeOverdue status/body/arg = %d/%s/%v, want 200 result for session", rec.Code, rec.Body.String(), fake.finalizeOverdueSessionID)
	}

	teacherReq := withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/grade-sync/preflight", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String(), "usr": "guru.ipa"})
	rec = httptest.NewRecorder()
	h.GetGradeSyncPreflight(rec, withRouteParam(teacherReq, "id", sessionID.String()))
	if rec.Code != http.StatusOK || fake.gradeSyncPreflightSessionID != sessionID || fake.checkSessionID != sessionID || fake.checkTeacherID != teacherID {
		t.Fatalf("GetGradeSyncPreflight status/args = %d/%v/%v/%v; body=%s", rec.Code, fake.gradeSyncPreflightSessionID, fake.checkSessionID, fake.checkTeacherID, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Sesi IPA") || !strings.Contains(rec.Body.String(), `"missing_score_count":1`) {
		t.Fatalf("GetGradeSyncPreflight body = %s, want serialized preflight", rec.Body.String())
	}

	teacherReq = withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/remedial?threshold=68.5", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String(), "usr": "guru.ipa"})
	rec = httptest.NewRecorder()
	h.ListRemedialCandidates(rec, withRouteParam(teacherReq, "id", sessionID.String()))
	if rec.Code != http.StatusOK || fake.remedialSessionID != sessionID || fake.remedialThreshold != 68.5 {
		t.Fatalf("ListRemedialCandidates status/args = %d/%v/%v; body=%s", rec.Code, fake.remedialSessionID, fake.remedialThreshold, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Siswa Remedial") || !strings.Contains(rec.Body.String(), `"question_count":10`) {
		t.Fatalf("ListRemedialCandidates body = %s, want remedial row", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListRemedialCandidates(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/remedial?threshold=101", ""), "id", sessionID.String()))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("ListRemedialCandidates invalid threshold status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetItemAnalysis(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/item-analysis", ""), "id", sessionID.String()))
	if rec.Code != http.StatusOK || fake.itemAnalysisSessionID != sessionID {
		t.Fatalf("GetItemAnalysis status/arg = %d/%v; body=%s", rec.Code, fake.itemAnalysisSessionID, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("GetItemAnalysis json.Unmarshal error = %v; body=%s", err, rec.Body.String())
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("GetItemAnalysis data = %#v, want object", body["data"])
	}
	items := requireAnySlice(t, data["items"], "items")
	first, ok := items[0].(map[string]any)
	if !ok || first["answer_key"] != "A" || first["recommendation"] != "Baik" || first["recommendation_tone"] != "success" {
		t.Fatalf("GetItemAnalysis first item = %#v, want admin answer key and recommendation", items[0])
	}

	rec = httptest.NewRecorder()
	h.GetItemAnalysis(rec, withRouteParam(withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/item-analysis", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()}), "id", sessionID.String()))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("GetItemAnalysis guru status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtSessionAdminLifecycleHandlersForwardValidRequests(t *testing.T) {
	sessionID := handlerTestUUID(90)
	packageID := handlerTestUUID(91)
	classID := handlerTestUUID(92)
	fake := &fakeCbtSessionService{}
	audit := &fakeCbtSessionAuditWriter{}
	h := &CbtSession{svc: fake, audit: audit}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/cbt/sessions", ""))
	if rec.Code != http.StatusOK || !fake.listCalled {
		t.Fatalf("List() status/called = %d/%v, want 200/true", rec.Code, fake.listCalled)
	}

	rec = httptest.NewRecorder()
	h.Create(rec, adminRequest(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"`+classID.String()+`","title":"Ujian IPA","scheduled_start":"2026-05-01T08:00:00Z","scheduled_end":"2026-05-01T09:30:00Z"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createInput.PackageID != packageID || fake.createInput.ClassID != classID || fake.createInput.ScopeRef != classID.String() || fake.createInput.Status != db.CbtSessionStatusEnumDraft {
		t.Fatalf("Create input = %+v, want package/class scope and default draft", fake.createInput)
	}
	if !fake.createInput.ScheduledStart.Valid || !fake.createInput.ScheduledEnd.Valid {
		t.Fatalf("Create schedule = %+v/%+v, want valid timestamptz", fake.createInput.ScheduledStart, fake.createInput.ScheduledEnd)
	}

	rec = httptest.NewRecorder()
	h.UpdateStatus(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/status", `{"status":"active"}`), "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateStatus() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateStatusID != sessionID || fake.updateStatusStatus != db.CbtSessionStatusEnumActive {
		t.Fatalf("UpdateStatus args = %v/%s, want session/active", fake.updateStatusID, fake.updateStatusStatus)
	}

	rec = httptest.NewRecorder()
	h.UpdateSchedule(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/schedule", `{"scheduled_start":"2026-05-02T08:00:00Z","scheduled_end":"2026-05-02T09:30:00Z"}`), "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateSchedule() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateScheduleID != sessionID || !fake.updateScheduleStart.Valid || !fake.updateScheduleEnd.Valid {
		t.Fatalf("UpdateSchedule args = %v/%+v/%+v, want valid schedule", fake.updateScheduleID, fake.updateScheduleStart, fake.updateScheduleEnd)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "CBT_SESSION_SCHEDULE_UPDATE" || audit.entries[0].EntityID != pgUUIDString(sessionID) {
		t.Fatalf("UpdateSchedule audit = %+v, want schedule update audit", audit.entries)
	}
	scheduleMeta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if scheduleMeta["previous_scheduled_start"] != "2026-05-01T08:00:00Z" || scheduleMeta["new_scheduled_start"] != "2026-05-02T08:00:00Z" {
		t.Fatalf("UpdateSchedule audit metadata = %+v, want previous and new schedule", scheduleMeta)
	}

	fake.listAuditRows = []db.ListEntityAuditLogsRow{{
		ID:         handlerTestUUID(102),
		UserID:     handlerTestUUID(103),
		Username:   pgtype.Text{String: "operator.cbt", Valid: true},
		Action:     "CBT_SESSION_SCHEDULE_UPDATE",
		EntityType: "cbt_session",
		EntityID:   pgUUIDString(sessionID),
		Metadata:   []byte(`{"new_scheduled_start":"2026-05-02T08:00:00Z"}`),
		CreatedAt:  pgtype.Timestamptz{Time: time.Date(2026, 5, 2, 7, 0, 0, 0, time.UTC), Valid: true},
	}}
	rec = httptest.NewRecorder()
	h.ListAuditLogs(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/audit-logs?page=2&per_page=25", ""), "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListAuditLogs() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listAuditID != sessionID || fake.listAuditLimit != 25 || fake.listAuditOffset != 25 {
		t.Fatalf("ListAuditLogs args = %v/%d/%d, want session page params", fake.listAuditID, fake.listAuditLimit, fake.listAuditOffset)
	}
	if !strings.Contains(rec.Body.String(), "CBT_SESSION_SCHEDULE_UPDATE") || !strings.Contains(rec.Body.String(), "operator.cbt") {
		t.Fatalf("ListAuditLogs body = %s, want serialized audit row", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Delete(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/sessions/"+sessionID.String(), ""), "id", sessionID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Delete() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID != sessionID {
		t.Fatalf("Delete id = %v, want %v", fake.deleteID, sessionID)
	}
}

func TestCbtSessionOperationalHandlersForwardValidRequests(t *testing.T) {
	sessionID := handlerTestUUID(93)
	classID := handlerTestUUID(94)
	participantID := handlerTestUUID(95)
	roomID := handlerTestUUID(96)
	answerID := handlerTestUUID(97)
	questionID := handlerTestUUID(98)
	fake := &fakeCbtSessionService{}
	audit := &fakeCbtSessionAuditWriter{}
	h := &CbtSession{svc: fake, audit: audit}

	run := func(name string, fn http.HandlerFunc, req *http.Request, want int) {
		t.Helper()
		rec := httptest.NewRecorder()
		fn(rec, req)
		if rec.Code != want {
			t.Fatalf("%s status = %d, want %d; body=%s", name, rec.Code, want, rec.Body.String())
		}
	}

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(adminRequest(method, target, body), pairs...)
	}

	run("Get", h.Get, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String(), "", "id", sessionID.String()), http.StatusOK)
	run("ListParticipants", h.ListParticipants, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", "id", sessionID.String()), http.StatusOK)
	run("Enroll class", h.Enroll, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"class","class_id":"`+classID.String()+`"}`, "id", sessionID.String()), http.StatusOK)
	run("EnrollClass alias", h.EnrollClass, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-class", `{"scope_type":"class","class_id":"`+classID.String()+`"}`, "id", sessionID.String()), http.StatusOK)
	run("Enroll grade", h.Enroll, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"grade","level":"VIII"}`, "id", sessionID.String()), http.StatusOK)
	run("Enroll school", h.EnrollSchool, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-school", "", "id", sessionID.String()), http.StatusOK)
	run("EnrollGrade", h.EnrollGrade, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-grade", `{"level":"IX"}`, "id", sessionID.String()), http.StatusOK)
	run("GenerateTokens", h.GenerateTokens, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/tokens", "", "id", sessionID.String()), http.StatusOK)
	run("RegenerateToken", h.RegenerateToken, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/token", "", "id", sessionID.String(), "pid", participantID.String()), http.StatusOK)
	run("AssignSeat", h.AssignSeat, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/seat", `{"room_id":"`+roomID.String()+`","seat_no":12}`, "id", sessionID.String(), "pid", participantID.String()), http.StatusOK)
	run("AutoAssignSeats", h.AutoAssignSeats, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/seats/auto", "", "id", sessionID.String()), http.StatusOK)
	run("ListRooms", h.ListRooms, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms", "", "id", sessionID.String()), http.StatusOK)
	run("CreateRoom", h.CreateRoom, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms", `{"room_name":"Ruang 1","capacity":0}`, "id", sessionID.String()), http.StatusCreated)
	run("DeleteRoom", h.DeleteRoom, adminRoute(http.MethodDelete, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String(), "", "id", sessionID.String(), "rid", roomID.String()), http.StatusNoContent)
	run("ShuffleRooms", h.ShuffleRooms, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms/shuffle", "", "id", sessionID.String()), http.StatusOK)
	run("GetProctoringStatus", h.GetProctoringStatus, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring", "", "id", sessionID.String()), http.StatusOK)
	run("ListMyProctorRooms", h.ListMyProctorRooms, adminRoute(http.MethodGet, "/api/cbt/proctoring/my-rooms", ""), http.StatusOK)
	run("GetRoomProctoringDashboard", h.GetRoomProctoringDashboard, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/proctoring", "", "id", sessionID.String(), "rid", roomID.String()), http.StatusOK)
	run("GetRoomProctorPrintPack", h.GetRoomProctorPrintPack, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/print-pack", "", "id", sessionID.String(), "rid", roomID.String()), http.StatusOK)
	run("GetSessionOperationalRecap", h.GetSessionOperationalRecap, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/operational-recap", "", "id", sessionID.String()), http.StatusOK)
	run("FlagParticipant", h.FlagParticipant, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/flag", `{"flag":true}`, "id", sessionID.String(), "pid", participantID.String()), http.StatusOK)
	run("ListUngradedEssays", h.ListUngradedEssays, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/essays/ungraded", "", "id", sessionID.String()), http.StatusOK)
	gradeReq := withRouteParams(withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/answers/"+answerID.String()+"/grade", strings.NewReader(`{"manual_score":87.5,"graded_by":"spoofed.actor"}`)), jwt.MapClaims{"roles": []any{"admin"}, "usr": "pengawas.utama", "uid": "01000000-0000-0000-0000-000000000000", "sub": "01000000-0000-0000-0000-000000000000", "ssid": "session-admin-1"}), "id", sessionID.String(), "aid", answerID.String())
	run("GradeEssay", h.GradeEssay, gradeReq, http.StatusOK)
	run("RecordAnswer", h.RecordAnswer, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", `{"question_id":"`+questionID.String()+`","answer":"B"}`, "id", sessionID.String(), "pid", participantID.String()), http.StatusOK)
	run("ScoreSession", h.ScoreSession, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/score", "", "id", sessionID.String()), http.StatusOK)
	run("GetResults", h.GetResults, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", "id", sessionID.String()), http.StatusOK)
	run("GetMinutes", h.GetMinutes, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/minutes", "", "id", sessionID.String()), http.StatusOK)
	run("GetParticipantAnswers", h.GetParticipantAnswers, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", "", "id", sessionID.String(), "pid", participantID.String()), http.StatusOK)
	run("GetRoomHandover", h.GetRoomHandover, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/handover", "", "id", sessionID.String(), "rid", roomID.String()), http.StatusOK)
	run("SaveRoomHandover", h.SaveRoomHandover, adminRoute(http.MethodPut, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/handover", `{"attendance_checked":true,"all_submitted_checked":true,"device_issue_checked":true,"room_clean_checked":true,"token_returned_checked":true,"assets_returned_checked":true,"incident_notes":"  aman  ","operator_notes":"reset 1","handover_notes":"selesai"}`, "id", sessionID.String(), "rid", roomID.String()), http.StatusOK)
	run("LockRoomHandover", h.LockRoomHandover, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/handover/lock", "", "id", sessionID.String(), "rid", roomID.String()), http.StatusOK)

	if fake.enrollClassSessionID != sessionID || fake.enrollClassID != classID {
		t.Fatalf("EnrollClass args = %v/%v, want session/class", fake.enrollClassSessionID, fake.enrollClassID)
	}
	if fake.enrollGradeSessionID != sessionID || fake.enrollGradeLevel != "IX" {
		t.Fatalf("EnrollGrade args = %v/%q, want session/IX", fake.enrollGradeSessionID, fake.enrollGradeLevel)
	}
	if fake.enrollSchoolSessionID != sessionID || fake.generateTokensSessionID != sessionID || fake.autoAssignSessionID != sessionID || fake.shuffleSessionID != sessionID || fake.scoreSessionID != sessionID {
		t.Fatalf("session action ids = school:%v tokens:%v auto:%v shuffle:%v score:%v, want session id", fake.enrollSchoolSessionID, fake.generateTokensSessionID, fake.autoAssignSessionID, fake.shuffleSessionID, fake.scoreSessionID)
	}
	if fake.regenerateParticipantID != participantID || fake.assignSeatParticipantID != participantID || fake.assignSeatRoomID != roomID || fake.assignSeatNo != 12 {
		t.Fatalf("participant token/seat args = %v/%v/%v/%d, want participant room seat", fake.regenerateParticipantID, fake.assignSeatParticipantID, fake.assignSeatRoomID, fake.assignSeatNo)
	}
	if fake.createRoomSessionID != sessionID || fake.createRoomName != "Ruang 1" || fake.createRoomCapacity != 30 || fake.deleteRoomID != roomID {
		t.Fatalf("room args = %v/%q/%d delete:%v, want default capacity and room delete", fake.createRoomSessionID, fake.createRoomName, fake.createRoomCapacity, fake.deleteRoomID)
	}
	if fake.roomDashboardID != roomID || fake.roomProctoringSessionID != sessionID || fake.roomProctoringRoomID != roomID || fake.roomEventsRoomID != roomID {
		t.Fatalf("room dashboard args = dashboard:%v proctoring:%v/%v events:%v, want room-scoped dashboard", fake.roomDashboardID, fake.roomProctoringSessionID, fake.roomProctoringRoomID, fake.roomEventsRoomID)
	}
	if fake.handoverRoomID != roomID || fake.saveHandoverRoomID != roomID || fake.lockHandoverRoomID != roomID || !fake.saveHandoverInput.AllSubmittedChecked || fake.saveHandoverInput.IncidentNotes != "  aman  " {
		t.Fatalf("handover args = get:%v save:%v lock:%v input:%+v, want room and forwarded payload", fake.handoverRoomID, fake.saveHandoverRoomID, fake.lockHandoverRoomID, fake.saveHandoverInput)
	}
	if fake.operationalRecapID != sessionID {
		t.Fatalf("operational recap id = %v, want %v", fake.operationalRecapID, sessionID)
	}
	if !fake.proctorRoomsIncludeAll || fake.proctorRoomsEmployeeID.Valid {
		t.Fatalf("proctor rooms args = includeAll:%v employee:%v, want admin all rooms without employee filter", fake.proctorRoomsIncludeAll, fake.proctorRoomsEmployeeID)
	}
	if fake.flagParticipantID != participantID || !fake.flagValue || fake.gradeAnswerID != answerID || fake.gradeScore != 87.5 || fake.gradeBy != "pengawas.utama" {
		t.Fatalf("flag/grade args = flag:%v/%v grade:%v/%v/%q, want forwarded values", fake.flagParticipantID, fake.flagValue, fake.gradeAnswerID, fake.gradeScore, fake.gradeBy)
	}
	if fake.recordParticipantID != participantID || fake.recordQuestionID != questionID || fake.recordAnswer != "B" || fake.participantAnswersID != participantID {
		t.Fatalf("answer args = %v/%v/%q participantAnswers:%v, want forwarded answer ids", fake.recordParticipantID, fake.recordQuestionID, fake.recordAnswer, fake.participantAnswersID)
	}
	if len(audit.entries) != 5 {
		t.Fatalf("audit entries len = %d, want 5", len(audit.entries))
	}
	if audit.entries[0].Action != "CBT_SESSION_PARTICIPANT_FLAG" || audit.entries[0].EntityID != pgUUIDString(sessionID) {
		t.Fatalf("flag audit entry = %+v, want flagged cbt session audit", audit.entries[0])
	}
	flagMeta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if flagMeta["participant_id"] != pgUUIDString(participantID) || flagMeta["suspicious_flag"] != true {
		t.Fatalf("flag audit metadata = %+v, want participant and suspicious flag", flagMeta)
	}
	if audit.entries[1].Action != "CBT_SESSION_ESSAY_GRADE" || audit.entries[1].EntityID != pgUUIDString(sessionID) {
		t.Fatalf("grade audit entry = %+v, want essay grade cbt session audit", audit.entries[1])
	}
	gradeMeta := mustAuditMetadataMap(t, audit.entries[1].Metadata)
	if gradeMeta["answer_id"] != pgUUIDString(answerID) || gradeMeta["graded_by"] != "pengawas.utama" {
		t.Fatalf("grade audit metadata = %+v, want answer id and jwt actor", gradeMeta)
	}
	if audit.entries[2].Action != "CBT_SESSION_SCORE" || audit.entries[2].EntityID != pgUUIDString(sessionID) {
		t.Fatalf("score audit entry = %+v, want session score audit", audit.entries[2])
	}
	if audit.entries[3].Action != "CBT_SESSION_ROOM_HANDOVER_SAVE" || audit.entries[3].EntityID != pgUUIDString(roomID) {
		t.Fatalf("handover save audit entry = %+v, want room handover save audit", audit.entries[3])
	}
	if audit.entries[4].Action != "CBT_SESSION_ROOM_HANDOVER_LOCK" || audit.entries[4].EntityID != pgUUIDString(roomID) {
		t.Fatalf("handover lock audit entry = %+v, want room handover lock audit", audit.entries[4])
	}
}

func TestCbtSessionAdminParticipantResponsesMaskTokens(t *testing.T) {
	sessionID := handlerTestUUID(226)
	participantID := handlerTestUUID(227)
	studentID := handlerTestUUID(228)
	rawToken := "abcdef1234567890abcdef1234567890"

	participants := serializeParticipantListRows([]db.ListCbtExamParticipantsRow{
		{ID: participantID, SessionID: sessionID, StudentID: studentID, Nis: "12345", Nama: "Ahmad", Token: rawToken},
	}, true)
	if got := participants[0]["token"]; got == "" || got == rawToken || strings.Contains(got.(string), "34567890") {
		t.Fatalf("participant list token = %#v, want masked token without raw value", got)
	}

	proctorRows := serializeProctoringRows([]db.GetSessionProctoringStatusRow{
		{ParticipantID: participantID, StudentID: studentID, Nis: "12345", Nama: "Ahmad", Token: rawToken},
	}, true)
	if got := proctorRows[0]["token"]; got == "" || got == rawToken || strings.Contains(got.(string), "34567890") {
		t.Fatalf("proctoring token = %#v, want masked token without raw value", got)
	}

	regenerated := serializeRegeneratedParticipantToken(db.RegenerateParticipantTokenRow{ID: participantID, Token: rawToken})
	if got := regenerated["token"]; got == "" || got == rawToken || strings.Contains(got.(string), "34567890") {
		t.Fatalf("regenerated token response = %#v, want masked token without raw value", got)
	}
	if regenerated["token_revealed"] != false {
		t.Fatalf("token_revealed = %#v, want false", regenerated["token_revealed"])
	}
}

func TestCbtSessionRoomProctorAndParticipantEventEndpoints(t *testing.T) {
	sessionID := handlerTestUUID(201)
	roomID := handlerTestUUID(202)
	participantID := handlerTestUUID(203)
	primaryEmployeeID := handlerTestUUID(204)
	secondaryEmployeeID := handlerTestUUID(205)
	actorID := handlerTestUUID(206)
	fake := &fakeCbtSessionService{}
	audit := &fakeCbtSessionAuditWriter{}
	h := &CbtSession{svc: fake, audit: audit}

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(
			withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
				"roles": []any{"admin"},
				"uid":   actorID.String(),
				"sub":   actorID.String(),
				"usr":   "admin.cbt",
				"ssid":  "session-admin-1",
			}),
			pairs...,
		)
	}

	rec := httptest.NewRecorder()
	h.GetRoomReadiness(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/readiness", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomReadiness status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.roomReadinessSessionID != sessionID {
		t.Fatalf("RoomReadiness session = %v, want %v", fake.roomReadinessSessionID, sessionID)
	}

	rec = httptest.NewRecorder()
	h.ListRoomProctors(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/proctors", "", "id", sessionID.String(), "rid", roomID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListRoomProctors status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.hasRoomSessionID != sessionID || fake.hasRoomID != roomID || fake.listRoomProctorsRoomID != roomID {
		t.Fatalf("ListRoomProctors args = hasRoom:%v/%v list:%v, want session/room", fake.hasRoomSessionID, fake.hasRoomID, fake.listRoomProctorsRoomID)
	}

	rec = httptest.NewRecorder()
	h.ReplaceRoomProctors(rec, adminRoute(http.MethodPut, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/proctors", `{"primary_employee_id":"`+primaryEmployeeID.String()+`","employee_ids":["`+primaryEmployeeID.String()+`","","`+secondaryEmployeeID.String()+`"]}`, "id", sessionID.String(), "rid", roomID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ReplaceRoomProctors status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.replaceProctorsRoomID != roomID || fake.replaceProctorsAssignedBy != actorID || fake.replaceProctorsPrimaryID != primaryEmployeeID {
		t.Fatalf("ReplaceRoomProctors ids = room:%v assigned:%v primary:%v, want room/actor/primary", fake.replaceProctorsRoomID, fake.replaceProctorsAssignedBy, fake.replaceProctorsPrimaryID)
	}
	if len(fake.replaceProctorsEmployeeIDs) != 2 || fake.replaceProctorsEmployeeIDs[0] != primaryEmployeeID || fake.replaceProctorsEmployeeIDs[1] != secondaryEmployeeID {
		t.Fatalf("ReplaceRoomProctors employee ids = %v, want primary and secondary only", fake.replaceProctorsEmployeeIDs)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "CBT_SESSION_ROOM_PROCTORS_REPLACE" || audit.entries[0].EntityID != pgUUIDString(roomID) {
		t.Fatalf("ReplaceRoomProctors audit = %+v, want room proctor replacement audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	h.ListParticipantEvents(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring/events?participant_id="+participantID.String()+"&limit=25", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListParticipantEvents status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.hasParticipantSessionID != sessionID || fake.hasParticipantID != participantID || fake.roomEventsSessionID != sessionID || fake.roomEventsParticipantID != participantID || fake.roomEventsLimit != 25 {
		t.Fatalf("ListParticipantEvents args = participant check:%v/%v events:%v/%v/%d, want session/participant/limit", fake.hasParticipantSessionID, fake.hasParticipantID, fake.roomEventsSessionID, fake.roomEventsParticipantID, fake.roomEventsLimit)
	}
}

func TestCbtSessionRealtimeProctorLiveActionAndReportEndpoints(t *testing.T) {
	sessionID := handlerTestUUID(221)
	roomID := handlerTestUUID(222)
	participantID := handlerTestUUID(223)
	eventID := handlerTestUUID(224)
	actorID := handlerTestUUID(225)
	fake := &fakeCbtSessionService{
		liveSummary: service.CbtProctoringLiveSummary{
			SessionID: sessionID.String(),
			RoomID:    roomID.String(),
			Counts:    map[string]int{"participants": 3, "needs_action": 2},
			Participants: []service.CbtProctoringParticipantDTO{
				{ParticipantID: participantID.String(), Nama: "Ahmad", ConnectionStatus: "selesai", SyncStatus: "sinkron"},
				{ParticipantID: handlerTestUUID(226).String(), Nama: "Budi", ConnectionStatus: "online", SyncStatus: "belum_sinkron", PendingAnswerCount: 2},
				{ParticipantID: handlerTestUUID(227).String(), Nama: "Cici", ConnectionStatus: "online", RiskLevel: "locked", LockedAt: "2026-05-03T08:15:00Z"},
			},
			LatestEvents: []service.CbtProctoringEventDTO{
				{ID: eventID.String(), ParticipantID: participantID.String(), Severity: "technical", IsMassTechnicalIssue: true, AcknowledgedAt: ""},
				{ID: handlerTestUUID(228).String(), ParticipantID: participantID.String(), Severity: "warning", AcknowledgedAt: "2026-05-03T08:10:00Z"},
				{ID: handlerTestUUID(229).String(), ParticipantID: participantID.String(), Severity: "critical", AcknowledgedAt: ""},
			},
			Actions: []db.CbtProctorAction{{ID: handlerTestUUID(230), SessionID: sessionID, ParticipantID: participantID, ActionType: "unlock_access", Reason: "Validasi pengawas"}},
		},
		participantScope:  db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID, Nama: "Ahmad", RoomName: "Ruang 1"},
		proctorEventScope: db.GetCbtProctorEventScopeRow{ID: eventID, ParticipantID: participantID, SessionID: sessionID, RoomID: roomID, Severity: "technical", Category: "device", RequiresNote: true, Nama: "Ahmad", RoomName: "Ruang 1"},
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &CbtSession{svc: fake, audit: audit}

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(
			withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
				"roles": []any{"admin"},
				"uid":   actorID.String(),
				"sub":   actorID.String(),
				"usr":   "proctor.admin",
				"ssid":  "session-admin-live",
			}),
			pairs...,
		)
	}

	rec := httptest.NewRecorder()
	h.GetProctoringLiveSummary(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring/live", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetProctoringLiveSummary status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.liveSummarySessionID != sessionID || fake.liveSummaryRoomID.Valid {
		t.Fatalf("live summary args = %v/%v, want session and empty room", fake.liveSummarySessionID, fake.liveSummaryRoomID)
	}
	if !strings.Contains(rec.Body.String(), `"needs_action":2`) {
		t.Fatalf("live summary body = %s, want returned summary counts", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetRoomProctoringLiveSummary(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/proctoring/live", "", "id", sessionID.String(), "rid", roomID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomProctoringLiveSummary status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.hasRoomSessionID != sessionID || fake.hasRoomID != roomID || fake.liveSummarySessionID != sessionID || fake.liveSummaryRoomID != roomID {
		t.Fatalf("room live args = hasRoom:%v/%v live:%v/%v, want session/room", fake.hasRoomSessionID, fake.hasRoomID, fake.liveSummarySessionID, fake.liveSummaryRoomID)
	}

	rec = httptest.NewRecorder()
	h.AcknowledgeProctorEvent(rec, adminRoute(http.MethodPost, "/api/cbt/proctoring/events/"+eventID.String()+"/ack", `{"notes":"  sudah dicek  "}`, "event_id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("AcknowledgeProctorEvent status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.proctorEventScopeID != eventID || fake.ackEventID != eventID || fake.ackActor.UserID != actorID || fake.ackActor.Username != "proctor.admin" || fake.ackNote != "sudah dicek" {
		t.Fatalf("ack args = scope:%v ack:%v actor:%+v note:%q, want event/actor/trimmed note", fake.proctorEventScopeID, fake.ackEventID, fake.ackActor, fake.ackNote)
	}

	actionBody := `{"action_type":"unlock_access","event_id":"` + eventID.String() + `","reason":"Peserta sudah diverifikasi","notes":"Izinkan lanjut"}`
	rec = httptest.NewRecorder()
	h.ExecuteProctorAction(rec, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/proctoring/actions", actionBody, "id", sessionID.String(), "pid", participantID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ExecuteProctorAction status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.participantScopeSessionID != sessionID || fake.participantScopeID != participantID || fake.proctorActionInput.EventID != eventID || fake.proctorActionInput.ActionType != "unlock_access" || fake.proctorActionActor.UserID != actorID {
		t.Fatalf("proctor action args = scope:%v/%v input:%+v actor:%+v, want participant/event action by actor", fake.participantScopeSessionID, fake.participantScopeID, fake.proctorActionInput, fake.proctorActionActor)
	}

	rec = httptest.NewRecorder()
	h.GetProctoringReportData(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring/report", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetProctoringReportData status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json.Unmarshal(report) error = %v; body=%s", err, rec.Body.String())
	}
	report, _ := envelope["data"].(map[string]any)
	severity, _ := report["severity_summary"].(map[string]any)
	submitted, _ := report["submitted_summary"].(map[string]any)
	if severity["technical"] != float64(1) || severity["critical"] != float64(1) || submitted["submitted"] != float64(1) || submitted["not_submitted"] != float64(2) {
		t.Fatalf("report summaries = severity:%+v submitted:%+v, want technical/critical and submitted counts", severity, submitted)
	}
	if got := len(report["technical_incidents"].([]any)); got != 1 {
		t.Fatalf("technical_incidents len = %d, want 1", got)
	}
	if got := len(report["unacknowledged_events"].([]any)); got != 2 {
		t.Fatalf("unacknowledged_events len = %d, want 2", got)
	}
	if got := len(report["sync_anomalies"].([]any)); got != 1 {
		t.Fatalf("sync_anomalies len = %d, want 1", got)
	}

	rec = httptest.NewRecorder()
	h.GetRoomProctoringReportData(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/proctoring/report", "", "id", sessionID.String(), "rid", roomID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomProctoringReportData status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.liveSummaryRoomID != roomID {
		t.Fatalf("room report live summary room = %v, want %v", fake.liveSummaryRoomID, roomID)
	}

	rec = httptest.NewRecorder()
	h.ForceSubmitParticipant(rec, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/force-submit", `{"reason":"Selesai manual"}`, "id", sessionID.String(), "pid", participantID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ForceSubmitParticipant status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.proctorActionInput.ActionType != "force_submit" || fake.proctorActionInput.Reason != "Selesai manual" || fake.proctorActionInput.Notes == "" || fake.proctorActionActor.Username != "proctor.admin" {
		t.Fatalf("force submit legacy action = input:%+v actor:%+v, want force_submit with default notes", fake.proctorActionInput, fake.proctorActionActor)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "CBT_SESSION_PARTICIPANT_FORCE_SUBMIT" || audit.entries[0].EntityID != pgUUIDString(sessionID) {
		t.Fatalf("force submit audit = %+v, want force-submit session audit", audit.entries)
	}
}

func TestCbtSessionProctorActorFromRequestUsesClaimsHeadersAndIP(t *testing.T) {
	userID := handlerTestUUID(207)
	employeeID := handlerTestUUID(208)
	req := withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/proctoring/actions", nil), jwt.MapClaims{
		"roles": []any{"guru"},
		"uid":   userID.String(),
		"sub":   handlerTestUUID(209).String(),
		"eid":   "  " + employeeID.String() + "  ",
		"usr":   "pengawas.ruang",
	})
	req.Header.Set("X-Request-ID", "  req-123  ")
	req.RemoteAddr = "198.51.100.10:4567"

	actor := cbtProctorActorFromRequest(req)
	if actor.UserID != userID || actor.EmployeeID != employeeID || actor.Username != "pengawas.ruang" {
		t.Fatalf("actor identity = user:%v employee:%v username:%q, want claims", actor.UserID, actor.EmployeeID, actor.Username)
	}
	if actor.RequestID != "req-123" {
		t.Fatalf("actor request id = %q, want trimmed req-123", actor.RequestID)
	}
	if !strings.Contains(actor.SourceIP, "198.51.100.10") {
		t.Fatalf("actor source ip = %q, want remote address/client IP", actor.SourceIP)
	}
}

func TestCbtSessionRequireRoomParticipantControlParams(t *testing.T) {
	sessionID := handlerTestUUID(210)
	roomID := handlerTestUUID(211)
	participantID := handlerTestUUID(212)
	h := &CbtSession{}

	req := withRouteParams(httptest.NewRequest(http.MethodPost, "/", nil), "id", sessionID.String(), "rid", roomID.String(), "pid", participantID.String())
	rec := httptest.NewRecorder()
	gotSession, gotRoom, gotParticipant, ok := h.requireRoomParticipantControlParams(rec, req)
	if !ok || gotSession != sessionID || gotRoom != roomID || gotParticipant != participantID {
		t.Fatalf("requireRoomParticipantControlParams valid = %v/%v/%v/%v, want parsed ids", gotSession, gotRoom, gotParticipant, ok)
	}

	req = withRouteParams(httptest.NewRequest(http.MethodPost, "/", nil), "id", sessionID.String(), "rid", roomID.String(), "pid", "not-a-uuid")
	rec = httptest.NewRecorder()
	_, _, _, ok = h.requireRoomParticipantControlParams(rec, req)
	if ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("requireRoomParticipantControlParams invalid pid = ok:%v status:%d, want false/400", ok, rec.Code)
	}
}

func TestCbtSessionItemAnalysisHelpersAndEndpoint(t *testing.T) {
	sessionID := handlerTestUUID(213)
	questionID := handlerTestUUID(214)
	base := db.GetSessionItemAnalysisRow{
		Position:            1,
		Points:              2,
		QuestionID:          questionID,
		QuestionCode:        "Q-001",
		QuestionText:        "Apa jawaban yang benar?",
		QuestionType:        "multiple_choice",
		Difficulty:          db.CbtQuestionDifficultyEnumMedium,
		AnswerKey:           "B",
		CpRef:               "CP-1",
		TpRef:               "TP-1",
		KdRef:               "KD-1",
		MaterialTopic:       "Bilangan",
		CognitiveLevel:      "C2",
		HotsFlag:            true,
		SubmittedCount:      10,
		AnsweredCount:       8,
		BlankCount:          2,
		CorrectCount:        6,
		IncorrectCount:      2,
		DifficultyIndex:     0.60,
		TopGroupCount:       3,
		TopCorrectCount:     3,
		BottomGroupCount:    3,
		BottomCorrectCount:  1,
		DiscriminationIndex: 0.35,
		AnswerDistribution:  []byte(`{"A":2,"B":6}`),
	}

	cases := []struct {
		name string
		row  db.GetSessionItemAnalysisRow
		want string
		tone string
	}{
		{name: "no submit", row: func() db.GetSessionItemAnalysisRow { r := base; r.SubmittedCount = 0; return r }(), want: "Belum ada submit", tone: "info"},
		{name: "essay unscored", row: func() db.GetSessionItemAnalysisRow {
			r := base
			r.QuestionType = "essay"
			r.UnscoredCount = 1
			return r
		}(), want: "Koreksi uraian belum lengkap", tone: "warning"},
		{name: "unanswered", row: func() db.GetSessionItemAnalysisRow { r := base; r.AnsweredCount = 0; return r }(), want: "Belum dijawab", tone: "danger"},
		{name: "negative discrimination", row: func() db.GetSessionItemAnalysisRow { r := base; r.DiscriminationIndex = -0.10; return r }(), want: "Cek kunci/rubrik", tone: "danger"},
		{name: "too hard", row: func() db.GetSessionItemAnalysisRow { r := base; r.DifficultyIndex = 0.19; return r }(), want: "Terlalu sulit", tone: "warning"},
		{name: "too easy", row: func() db.GetSessionItemAnalysisRow { r := base; r.DifficultyIndex = 0.91; return r }(), want: "Terlalu mudah", tone: "warning"},
		{name: "low discrimination", row: func() db.GetSessionItemAnalysisRow { r := base; r.DiscriminationIndex = 0.14; return r }(), want: "Daya pembeda rendah", tone: "warning"},
		{name: "many blanks", row: func() db.GetSessionItemAnalysisRow { r := base; r.BlankCount = 6; return r }(), want: "Banyak jawaban kosong", tone: "warning"},
		{name: "good", row: base, want: "Baik", tone: "success"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sessionItemAnalysisRecommendation(tc.row); got != tc.want {
				t.Fatalf("recommendation = %q, want %q", got, tc.want)
			}
			if got := sessionItemAnalysisTone(tc.row); got != tc.tone {
				t.Fatalf("tone = %q, want %q", got, tc.tone)
			}
		})
	}

	serialized := serializeSessionItemAnalysisRow(base, false)
	if serialized["answer_key"] != "" || serialized["question_id"] != questionID.String() || serialized["recommendation"] != "Baik" || serialized["recommendation_tone"] != "success" {
		t.Fatalf("serialized redacted row = %+v, want redacted answer key and identifiers/recommendation", serialized)
	}
	if distribution, ok := serialized["answer_distribution"].(map[string]any); !ok || distribution["B"].(float64) != 6 {
		t.Fatalf("serialized answer_distribution = %#v, want decoded JSON distribution", serialized["answer_distribution"])
	}
	serialized = serializeSessionItemAnalysisRow(base, true)
	if serialized["answer_key"] != "B" {
		t.Fatalf("serialized answer_key = %#v, want included key", serialized["answer_key"])
	}

	fake := &fakeCbtSessionService{itemAnalysisRows: []db.GetSessionItemAnalysisRow{base}}
	req := withRouteParams(adminRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/item-analysis", ""), "id", sessionID.String())
	rec := httptest.NewRecorder()
	(&CbtSession{svc: fake}).GetItemAnalysis(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetItemAnalysis status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.itemAnalysisSessionID != sessionID {
		t.Fatalf("GetItemAnalysis session id = %v, want %v", fake.itemAnalysisSessionID, sessionID)
	}
	if !strings.Contains(rec.Body.String(), `"answer_key":"B"`) || !strings.Contains(rec.Body.String(), `"recommendation":"Baik"`) {
		t.Fatalf("GetItemAnalysis body = %s, want serialized item with answer key and recommendation", rec.Body.String())
	}
}

func TestCbtSessionGetMinutesRedactsTokensForTeacher(t *testing.T) {
	sessionID := handlerTestUUID(230)
	teacherID := handlerTestUUID(231)
	participantID := handlerTestUUID(232)
	studentID := handlerTestUUID(233)
	fake := &fakeCbtSessionService{
		checkAllowed: true,
		getRow: db.GetCbtExamSessionRow{
			ID:    sessionID,
			Title: "Asesmen IPA",
		},
		listParticipantsRows: []db.ListCbtExamParticipantsRow{
			{
				ID:        participantID,
				SessionID: sessionID,
				StudentID: studentID,
				Nis:       "12345",
				Nama:      "Ahmad",
				Token:     "secret-token",
			},
		},
	}
	h := &CbtSession{svc: fake}
	req := withRouteParams(
		withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/minutes", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()}),
		"id", sessionID.String(),
	)
	rec := httptest.NewRecorder()

	h.GetMinutes(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetMinutes(guru) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listParticipantsTeacherID != teacherID {
		t.Fatalf("ListParticipantsByTeacher teacher = %v, want %v", fake.listParticipantsTeacherID, teacherID)
	}
	var payload struct {
		Data struct {
			Participants []map[string]any `json:"participants"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode GetMinutes response: %v body=%s", err, rec.Body.String())
	}
	if len(payload.Data.Participants) != 1 {
		t.Fatalf("participants len = %d, want 1", len(payload.Data.Participants))
	}
	if got := payload.Data.Participants[0]["token"]; got != "" {
		t.Fatalf("guru minutes token = %#v, want redacted empty string", got)
	}
}

func TestCbtSessionResetParticipantAccessScopesGuruToParticipantClass(t *testing.T) {
	sessionID := handlerTestUUID(234)
	participantID := handlerTestUUID(235)
	teacherID := handlerTestUUID(236)
	fake := &fakeCbtSessionService{
		checkAllowed:      true,
		hasParticipantSet: true,
		hasParticipant:    false,
	}
	h := &CbtSession{svc: fake}
	req := withRouteParams(
		withClaims(
			httptest.NewRequest(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/reset-access", nil),
			jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String(), "usr": "guru.ipa"},
		),
		"id", sessionID.String(),
		"pid", participantID.String(),
	)
	rec := httptest.NewRecorder()

	h.ResetParticipantAccess(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("ResetParticipantAccess(guru outside participant class) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	if fake.hasParticipantTeacherID != teacherID {
		t.Fatalf("teacher participant scope check = %v, want %v", fake.hasParticipantTeacherID, teacherID)
	}
	if fake.resetParticipantID.Valid {
		t.Fatalf("reset called despite teacher scope denial: participant=%v actor=%q", fake.resetParticipantID, fake.resetActor)
	}
}

func TestCbtSessionRoomProctorDashboardScopesAssignedProctor(t *testing.T) {
	sessionID := handlerTestUUID(108)
	roomID := handlerTestUUID(109)
	employeeID := handlerTestUUID(110)

	req := withRouteParams(
		withClaims(
			httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/proctoring", nil),
			jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()},
		),
		"id", sessionID.String(),
		"rid", roomID.String(),
	)
	fake := &fakeCbtSessionService{roomProctorAllowed: true}
	rec := httptest.NewRecorder()
	(&CbtSession{svc: fake}).GetRoomProctoringDashboard(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomProctoringDashboard(assigned proctor) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	printReq := withRouteParams(
		withClaims(
			httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/print-pack", nil),
			jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()},
		),
		"id", sessionID.String(),
		"rid", roomID.String(),
	)
	rec = httptest.NewRecorder()
	(&CbtSession{svc: fake}).GetRoomProctorPrintPack(rec, printReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomProctorPrintPack(assigned proctor) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	handoverReq := withRouteParams(
		withClaims(
			httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/handover", nil),
			jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()},
		),
		"id", sessionID.String(),
		"rid", roomID.String(),
	)
	rec = httptest.NewRecorder()
	(&CbtSession{svc: fake}).GetRoomHandover(rec, handoverReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomHandover(assigned proctor) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	saveReq := withRouteParams(
		withClaims(
			httptest.NewRequest(http.MethodPut, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/handover", strings.NewReader(`{"attendance_checked":true}`)),
			jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String(), "uid": handlerTestUUID(251).String(), "sub": handlerTestUUID(251).String()},
		),
		"id", sessionID.String(),
		"rid", roomID.String(),
	)
	rec = httptest.NewRecorder()
	(&CbtSession{svc: fake}).SaveRoomHandover(rec, saveReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("SaveRoomHandover(assigned proctor) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.roomProctorSessionID != sessionID || fake.roomProctorRoomID != roomID || fake.roomProctorEmployeeID != employeeID {
		t.Fatalf("room proctor check = %v/%v/%v, want session/room/employee", fake.roomProctorSessionID, fake.roomProctorRoomID, fake.roomProctorEmployeeID)
	}

	denied := &fakeCbtSessionService{}
	rec = httptest.NewRecorder()
	(&CbtSession{svc: denied}).GetRoomProctoringDashboard(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("GetRoomProctoringDashboard(unassigned proctor) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	(&CbtSession{svc: denied}).GetRoomProctorPrintPack(rec, printReq)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("GetRoomProctorPrintPack(unassigned proctor) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	(&CbtSession{svc: denied}).GetRoomHandover(rec, handoverReq)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("GetRoomHandover(unassigned proctor) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtSessionListMyProctorRoomsScopesByActor(t *testing.T) {
	employeeID := handlerTestUUID(111)
	req := withClaims(
		httptest.NewRequest(http.MethodGet, "/api/cbt/proctoring/my-rooms", nil),
		jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()},
	)
	fake := &fakeCbtSessionService{}
	rec := httptest.NewRecorder()
	(&CbtSession{svc: fake}).ListMyProctorRooms(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ListMyProctorRooms(guru) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.proctorRoomsEmployeeID != employeeID || fake.proctorRoomsIncludeAll {
		t.Fatalf("guru proctor rooms args = %v includeAll:%v, want employee scoped", fake.proctorRoomsEmployeeID, fake.proctorRoomsIncludeAll)
	}

	adminFake := &fakeCbtSessionService{}
	rec = httptest.NewRecorder()
	(&CbtSession{svc: adminFake}).ListMyProctorRooms(rec, adminRequest(http.MethodGet, "/api/cbt/proctoring/my-rooms", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListMyProctorRooms(admin) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !adminFake.proctorRoomsIncludeAll || adminFake.proctorRoomsEmployeeID.Valid {
		t.Fatalf("admin proctor rooms args = %v includeAll:%v, want all rooms", adminFake.proctorRoomsEmployeeID, adminFake.proctorRoomsIncludeAll)
	}

	rec = httptest.NewRecorder()
	(&CbtSession{svc: &fakeCbtSessionService{}}).ListMyProctorRooms(
		rec,
		withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/proctoring/my-rooms", nil), jwt.MapClaims{"roles": []any{"staf"}}),
	)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("ListMyProctorRooms(staf missing eid) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtSessionOperationalHandlersMapServiceErrors(t *testing.T) {
	sessionID := handlerTestUUID(112)
	classID := handlerTestUUID(113)
	participantID := handlerTestUUID(114)
	roomID := handlerTestUUID(115)
	answerID := handlerTestUUID(116)
	questionID := handlerTestUUID(117)
	errDB := errors.New("db down")

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(adminRequest(method, target, body), pairs...)
	}
	tests := []struct {
		name    string
		handler func(*CbtSession, http.ResponseWriter, *http.Request)
		svc     *fakeCbtSessionService
		req     *http.Request
	}{
		{
			name:    "get",
			handler: (*CbtSession).Get,
			svc:     &fakeCbtSessionService{getErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String(), "", "id", sessionID.String()),
		},
		{
			name:    "list participants",
			handler: (*CbtSession).ListParticipants,
			svc:     &fakeCbtSessionService{listParticipantsErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", "id", sessionID.String()),
		},
		{
			name:    "enroll class",
			handler: (*CbtSession).Enroll,
			svc:     &fakeCbtSessionService{enrollClassErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"class","class_id":"`+classID.String()+`"}`, "id", sessionID.String()),
		},
		{
			name:    "enroll grade",
			handler: (*CbtSession).EnrollGrade,
			svc:     &fakeCbtSessionService{enrollGradeErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-grade", `{"level":"VIII"}`, "id", sessionID.String()),
		},
		{
			name:    "enroll school",
			handler: (*CbtSession).EnrollSchool,
			svc:     &fakeCbtSessionService{enrollSchoolErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-school", "", "id", sessionID.String()),
		},
		{
			name:    "generate tokens",
			handler: (*CbtSession).GenerateTokens,
			svc:     &fakeCbtSessionService{generateTokensErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/tokens", "", "id", sessionID.String()),
		},
		{
			name:    "regenerate token",
			handler: (*CbtSession).RegenerateToken,
			svc:     &fakeCbtSessionService{regenerateErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/token", "", "id", sessionID.String(), "pid", participantID.String()),
		},
		{
			name:    "assign seat",
			handler: (*CbtSession).AssignSeat,
			svc:     &fakeCbtSessionService{assignSeatErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/seat", `{"room_id":"`+roomID.String()+`","seat_no":4}`, "id", sessionID.String(), "pid", participantID.String()),
		},
		{
			name:    "auto assign seats",
			handler: (*CbtSession).AutoAssignSeats,
			svc:     &fakeCbtSessionService{autoAssignErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/seats/auto", "", "id", sessionID.String()),
		},
		{
			name:    "list rooms",
			handler: (*CbtSession).ListRooms,
			svc:     &fakeCbtSessionService{listRoomsErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms", "", "id", sessionID.String()),
		},
		{
			name:    "create room",
			handler: (*CbtSession).CreateRoom,
			svc:     &fakeCbtSessionService{createRoomErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms", `{"room_name":"Ruang 1","capacity":20}`, "id", sessionID.String()),
		},
		{
			name:    "delete room",
			handler: (*CbtSession).DeleteRoom,
			svc:     &fakeCbtSessionService{deleteRoomErr: errDB},
			req:     adminRoute(http.MethodDelete, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String(), "", "id", sessionID.String(), "rid", roomID.String()),
		},
		{
			name:    "shuffle rooms",
			handler: (*CbtSession).ShuffleRooms,
			svc:     &fakeCbtSessionService{shuffleErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms/shuffle", "", "id", sessionID.String()),
		},
		{
			name:    "proctoring status",
			handler: (*CbtSession).GetProctoringStatus,
			svc:     &fakeCbtSessionService{proctoringErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring", "", "id", sessionID.String()),
		},
		{
			name:    "flag participant",
			handler: (*CbtSession).FlagParticipant,
			svc:     &fakeCbtSessionService{flagErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/flag", `{"flag":true}`, "id", sessionID.String(), "pid", participantID.String()),
		},
		{
			name:    "ungraded essays",
			handler: (*CbtSession).ListUngradedEssays,
			svc:     &fakeCbtSessionService{ungradedErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/essays/ungraded", "", "id", sessionID.String()),
		},
		{
			name:    "grade essay",
			handler: (*CbtSession).GradeEssay,
			svc:     &fakeCbtSessionService{gradeErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/answers/"+answerID.String()+"/grade", `{"manual_score":90,"graded_by":"guru"}`, "id", sessionID.String(), "aid", answerID.String()),
		},
		{
			name:    "record answer",
			handler: (*CbtSession).RecordAnswer,
			svc:     &fakeCbtSessionService{recordErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", `{"question_id":"`+questionID.String()+`","answer":"A"}`, "id", sessionID.String(), "pid", participantID.String()),
		},
		{
			name:    "score session",
			handler: (*CbtSession).ScoreSession,
			svc:     &fakeCbtSessionService{scoreErr: errDB},
			req:     adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/score", "", "id", sessionID.String()),
		},
		{
			name:    "get results session error",
			handler: (*CbtSession).GetResults,
			svc:     &fakeCbtSessionService{getErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", "id", sessionID.String()),
		},
		{
			name:    "get results rows error",
			handler: (*CbtSession).GetResults,
			svc:     &fakeCbtSessionService{getResultsErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", "id", sessionID.String()),
		},
		{
			name:    "get minutes participants error",
			handler: (*CbtSession).GetMinutes,
			svc:     &fakeCbtSessionService{listParticipantsErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/minutes", "", "id", sessionID.String()),
		},
		{
			name:    "get minutes rooms error",
			handler: (*CbtSession).GetMinutes,
			svc:     &fakeCbtSessionService{listRoomsErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/minutes", "", "id", sessionID.String()),
		},
		{
			name:    "participant answers",
			handler: (*CbtSession).GetParticipantAnswers,
			svc:     &fakeCbtSessionService{participantAnswersErr: errDB},
			req:     adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", "", "id", sessionID.String(), "pid", participantID.String()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&CbtSession{svc: tt.svc}, rec, tt.req)
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500; body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionAdminValidationAndServiceErrors(t *testing.T) {
	sessionID := handlerTestUUID(118)
	packageID := handlerTestUUID(119)
	classID := handlerTestUUID(120)
	eventID := handlerTestUUID(121)
	errDB := errors.New("db down")
	validCreate := `{"package_id":"` + packageID.String() + `","class_id":"` + classID.String() + `","title":"Ujian IPA","scheduled_start":"2026-05-01T08:00:00Z","scheduled_end":"2026-05-01T09:30:00Z"}`

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(adminRequest(method, target, body), pairs...)
	}
	plainRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(httptest.NewRequest(method, target, strings.NewReader(body)), pairs...)
	}
	tests := []struct {
		name       string
		handler    func(*CbtSession, http.ResponseWriter, *http.Request)
		svc        *fakeCbtSessionService
		req        *http.Request
		wantStatus int
	}{
		{name: "list forbidden", handler: (*CbtSession).List, req: plainRoute(http.MethodGet, "/api/cbt/sessions", ""), wantStatus: http.StatusForbidden},
		{name: "list service error", handler: (*CbtSession).List, svc: &fakeCbtSessionService{listErr: errDB}, req: adminRoute(http.MethodGet, "/api/cbt/sessions", ""), wantStatus: http.StatusInternalServerError},
		{name: "create forbidden", handler: (*CbtSession).Create, req: plainRoute(http.MethodPost, "/api/cbt/sessions", validCreate), wantStatus: http.StatusForbidden},
		{name: "create invalid json", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{`), wantStatus: http.StatusBadRequest},
		{name: "create invalid package", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"bad"}`), wantStatus: http.StatusBadRequest},
		{name: "create invalid class", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"bad"}`), wantStatus: http.StatusBadRequest},
		{name: "create missing class for class scope", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","scope_type":"class"}`), wantStatus: http.StatusBadRequest},
		{name: "create cross grade without special event", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"`+classID.String()+`","allow_cross_grade":true}`), wantStatus: http.StatusBadRequest},
		{name: "create grade missing scope ref", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","scope_type":"grade"}`), wantStatus: http.StatusBadRequest},
		{name: "create invalid event", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"`+classID.String()+`","event_id":"bad"}`), wantStatus: http.StatusBadRequest},
		{name: "create invalid start", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"`+classID.String()+`","scheduled_start":"bad"}`), wantStatus: http.StatusBadRequest},
		{name: "create invalid end", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"`+classID.String()+`","scheduled_start":"2026-05-01T08:00:00Z","scheduled_end":"bad"}`), wantStatus: http.StatusBadRequest},
		{name: "create end before start", handler: (*CbtSession).Create, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"`+classID.String()+`","scheduled_start":"2026-05-01T09:00:00Z","scheduled_end":"2026-05-01T08:00:00Z"}`), wantStatus: http.StatusBadRequest},
		{name: "create package quality conflict", handler: (*CbtSession).Create, svc: &fakeCbtSessionService{createErr: errors.Join(domain.ErrConflict, errors.New("paket soal belum memiliki soal terbit"))}, req: adminRoute(http.MethodPost, "/api/cbt/sessions", validCreate), wantStatus: http.StatusConflict},
		{name: "create package event mismatch bad request", handler: (*CbtSession).Create, svc: &fakeCbtSessionService{createErr: errors.Join(domain.ErrBadRequest, errors.New("paket khusus event hanya boleh dipakai pada sesi event yang sama"))}, req: adminRoute(http.MethodPost, "/api/cbt/sessions", validCreate), wantStatus: http.StatusBadRequest},
		{name: "create service error", handler: (*CbtSession).Create, svc: &fakeCbtSessionService{createErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions", validCreate), wantStatus: http.StatusInternalServerError},
		{name: "create with event service error", handler: (*CbtSession).Create, svc: &fakeCbtSessionService{createErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"`+classID.String()+`","event_id":"`+eventID.String()+`","scheduled_start":"2026-05-01T08:00:00Z","scheduled_end":"2026-05-01T09:00:00Z"}`), wantStatus: http.StatusInternalServerError},
		{name: "update status invalid id", handler: (*CbtSession).UpdateStatus, req: adminRoute(http.MethodPatch, "/api/cbt/sessions/bad/status", `{"status":"active"}`, "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "update status invalid json", handler: (*CbtSession).UpdateStatus, req: adminRoute(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/status", `{`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "update status service error", handler: (*CbtSession).UpdateStatus, svc: &fakeCbtSessionService{updateStatusErr: errDB}, req: adminRoute(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/status", `{"status":"active"}`, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "update status bad request", handler: (*CbtSession).UpdateStatus, svc: &fakeCbtSessionService{updateStatusErr: errors.Join(domain.ErrBadRequest, errors.New("status sesi CBT tidak valid"))}, req: adminRoute(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/status", `{"status":"archived"}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "update status readiness conflict", handler: (*CbtSession).UpdateStatus, svc: &fakeCbtSessionService{updateStatusErr: errors.Join(domain.ErrConflict, errors.New("masih ada 1 ruangan belum punya pengawas"))}, req: adminRoute(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/status", `{"status":"active"}`, "id", sessionID.String()), wantStatus: http.StatusConflict},
		{name: "delete invalid id", handler: (*CbtSession).Delete, req: adminRoute(http.MethodDelete, "/api/cbt/sessions/bad", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "delete service error", handler: (*CbtSession).Delete, svc: &fakeCbtSessionService{deleteErr: errDB}, req: adminRoute(http.MethodDelete, "/api/cbt/sessions/"+sessionID.String(), "", "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakeCbtSessionService{}
			}
			rec := httptest.NewRecorder()
			tt.handler(&CbtSession{svc: svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionOperationalValidationAndAccessBranches(t *testing.T) {
	sessionID := handlerTestUUID(122)
	participantID := handlerTestUUID(123)
	roomID := handlerTestUUID(124)
	answerID := handlerTestUUID(125)
	questionID := handlerTestUUID(126)
	classID := handlerTestUUID(127)
	teacherID := handlerTestUUID(128)
	errDB := errors.New("db down")

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(adminRequest(method, target, body), pairs...)
	}
	plainRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(httptest.NewRequest(method, target, strings.NewReader(body)), pairs...)
	}
	guruRoute := func(method, target, body string, claims jwt.MapClaims, pairs ...string) *http.Request {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, claims))
		return withRouteParams(req, pairs...)
	}
	validGuruClaims := jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()}

	tests := []struct {
		name       string
		handler    func(*CbtSession, http.ResponseWriter, *http.Request)
		svc        *fakeCbtSessionService
		req        *http.Request
		wantStatus int
	}{
		{name: "get without claims unauthorized", handler: (*CbtSession).Get, req: plainRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String(), "", "id", sessionID.String()), wantStatus: http.StatusUnauthorized},
		{name: "get guru missing eid forbidden", handler: (*CbtSession).Get, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String(), "", jwt.MapClaims{"roles": []any{"guru"}}, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "get guru check error", handler: (*CbtSession).Get, svc: &fakeCbtSessionService{checkErr: errDB}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String(), "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "get guru denied", handler: (*CbtSession).Get, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String(), "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "list participants invalid id", handler: (*CbtSession).ListParticipants, req: adminRoute(http.MethodGet, "/api/cbt/sessions/bad/participants", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "list participants guru denied", handler: (*CbtSession).ListParticipants, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "enroll invalid session", handler: (*CbtSession).Enroll, req: adminRoute(http.MethodPost, "/api/cbt/sessions/bad/enroll", `{}`, "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "enroll invalid json", handler: (*CbtSession).Enroll, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "enroll invalid class id", handler: (*CbtSession).Enroll, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"class","class_id":"bad"}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "enroll grade missing level", handler: (*CbtSession).Enroll, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"grade"}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "enroll grade service error", handler: (*CbtSession).Enroll, svc: &fakeCbtSessionService{enrollGradeErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"grade","level":"VIII"}`, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "enroll class setup conflict", handler: (*CbtSession).Enroll, svc: &fakeCbtSessionService{enrollClassErr: errors.Join(domain.ErrConflict, errors.New("sesi aktif"))}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"class","class_id":"`+classID.String()+`"}`, "id", sessionID.String()), wantStatus: http.StatusConflict},
		{name: "enroll school service error", handler: (*CbtSession).Enroll, svc: &fakeCbtSessionService{enrollSchoolErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"school"}`, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "enroll invalid scope", handler: (*CbtSession).Enroll, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll", `{"scope_type":"room"}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "enroll grade invalid json", handler: (*CbtSession).EnrollGrade, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-grade", `{`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "enroll grade missing level", handler: (*CbtSession).EnrollGrade, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-grade", `{}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "regenerate invalid participant", handler: (*CbtSession).RegenerateToken, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/bad/token", "", "id", sessionID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "regenerate participant outside session", handler: (*CbtSession).RegenerateToken, svc: &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: false}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/token", "", "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusForbidden},
		{name: "regenerate participant membership error", handler: (*CbtSession).RegenerateToken, svc: &fakeCbtSessionService{hasParticipantErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/token", "", "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusInternalServerError},
		{name: "assign invalid participant", handler: (*CbtSession).AssignSeat, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/bad/seat", `{}`, "id", sessionID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "assign participant outside session", handler: (*CbtSession).AssignSeat, svc: &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: false}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/seat", `{"room_id":"`+roomID.String()+`","seat_no":1}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusForbidden},
		{name: "assign invalid json", handler: (*CbtSession).AssignSeat, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/seat", `{`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "assign invalid room", handler: (*CbtSession).AssignSeat, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/seat", `{"room_id":"bad","seat_no":1}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "assign room outside session", handler: (*CbtSession).AssignSeat, svc: &fakeCbtSessionService{hasRoomSet: true, hasRoom: false}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/seat", `{"room_id":"`+roomID.String()+`","seat_no":1}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusForbidden},
		{name: "assign room membership error", handler: (*CbtSession).AssignSeat, svc: &fakeCbtSessionService{hasRoomErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/seat", `{"room_id":"`+roomID.String()+`","seat_no":1}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusInternalServerError},
		{name: "assign invalid seat", handler: (*CbtSession).AssignSeat, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/seat", `{"room_id":"`+roomID.String()+`","seat_no":0}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "create room invalid json", handler: (*CbtSession).CreateRoom, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms", `{`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "create room missing name", handler: (*CbtSession).CreateRoom, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms", `{"capacity":20}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "delete room invalid room", handler: (*CbtSession).DeleteRoom, req: adminRoute(http.MethodDelete, "/api/cbt/sessions/"+sessionID.String()+"/rooms/bad", "", "id", sessionID.String(), "rid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "delete room outside session", handler: (*CbtSession).DeleteRoom, svc: &fakeCbtSessionService{hasRoomSet: true, hasRoom: false}, req: adminRoute(http.MethodDelete, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String(), "", "id", sessionID.String(), "rid", roomID.String()), wantStatus: http.StatusForbidden},
		{name: "delete room membership error", handler: (*CbtSession).DeleteRoom, svc: &fakeCbtSessionService{hasRoomErr: errDB}, req: adminRoute(http.MethodDelete, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String(), "", "id", sessionID.String(), "rid", roomID.String()), wantStatus: http.StatusInternalServerError},
		{name: "flag invalid participant", handler: (*CbtSession).FlagParticipant, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/bad/flag", `{}`, "id", sessionID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "flag participant outside session", handler: (*CbtSession).FlagParticipant, svc: &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: false}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/flag", `{"flag":true}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusForbidden},
		{name: "flag participant membership error", handler: (*CbtSession).FlagParticipant, svc: &fakeCbtSessionService{hasParticipantErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/flag", `{"flag":true}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusInternalServerError},
		{name: "flag invalid json", handler: (*CbtSession).FlagParticipant, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/flag", `{`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "grade invalid answer id", handler: (*CbtSession).GradeEssay, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/answers/bad/grade", `{}`, "id", sessionID.String(), "aid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "grade answer outside session", handler: (*CbtSession).GradeEssay, svc: &fakeCbtSessionService{hasAnswerSet: true, hasAnswer: false}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/answers/"+answerID.String()+"/grade", `{"manual_score":50}`, "id", sessionID.String(), "aid", answerID.String()), wantStatus: http.StatusForbidden},
		{name: "grade answer membership error", handler: (*CbtSession).GradeEssay, svc: &fakeCbtSessionService{hasAnswerErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/answers/"+answerID.String()+"/grade", `{"manual_score":50}`, "id", sessionID.String(), "aid", answerID.String()), wantStatus: http.StatusInternalServerError},
		{name: "grade invalid json", handler: (*CbtSession).GradeEssay, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/answers/"+answerID.String()+"/grade", `{`, "id", sessionID.String(), "aid", answerID.String()), wantStatus: http.StatusBadRequest},
		{name: "grade score below range", handler: (*CbtSession).GradeEssay, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/answers/"+answerID.String()+"/grade", `{"manual_score":-1}`, "id", sessionID.String(), "aid", answerID.String()), wantStatus: http.StatusBadRequest},
		{name: "grade score above range", handler: (*CbtSession).GradeEssay, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/answers/"+answerID.String()+"/grade", `{"manual_score":101}`, "id", sessionID.String(), "aid", answerID.String()), wantStatus: http.StatusBadRequest},
		{name: "record answer invalid participant", handler: (*CbtSession).RecordAnswer, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/bad/answers", `{}`, "id", sessionID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "record answer participant outside session", handler: (*CbtSession).RecordAnswer, svc: &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: false}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", `{"question_id":"`+questionID.String()+`","answer":"A"}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusForbidden},
		{name: "record answer invalid json", handler: (*CbtSession).RecordAnswer, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", `{`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "record answer invalid question", handler: (*CbtSession).RecordAnswer, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", `{"question_id":"bad"}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "record answer question outside package", handler: (*CbtSession).RecordAnswer, svc: &fakeCbtSessionService{recordErr: service.ErrExamQuestionScope}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", `{"question_id":"`+questionID.String()+`","answer":"A"}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "score invalid id", handler: (*CbtSession).ScoreSession, req: adminRoute(http.MethodPost, "/api/cbt/sessions/bad/score", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "score guru forbidden", handler: (*CbtSession).ScoreSession, req: guruRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/score", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "get results forbidden", handler: (*CbtSession).GetResults, req: plainRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "get results invalid id", handler: (*CbtSession).GetResults, req: adminRoute(http.MethodGet, "/api/cbt/sessions/bad/results", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "get minutes invalid id", handler: (*CbtSession).GetMinutes, req: adminRoute(http.MethodGet, "/api/cbt/sessions/bad/minutes", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "get minutes session error", handler: (*CbtSession).GetMinutes, svc: &fakeCbtSessionService{getErr: errDB}, req: adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/minutes", "", "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "participant answers invalid id", handler: (*CbtSession).GetParticipantAnswers, req: adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants/bad/answers", "", "id", sessionID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "participant answers outside session", handler: (*CbtSession).GetParticipantAnswers, svc: &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: false}, req: adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", "", "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusForbidden},
		{name: "guru list forbidden without claims", handler: (*CbtSession).GuruAwareList, req: plainRoute(http.MethodGet, "/api/cbt/sessions", ""), wantStatus: http.StatusForbidden},
		{name: "guru list missing eid", handler: (*CbtSession).GuruAwareList, req: guruRoute(http.MethodGet, "/api/cbt/sessions", "", jwt.MapClaims{"roles": []any{"guru"}}), wantStatus: http.StatusForbidden},
		{name: "guru list invalid eid", handler: (*CbtSession).GuruAwareList, req: guruRoute(http.MethodGet, "/api/cbt/sessions", "", jwt.MapClaims{"roles": []any{"guru"}, "eid": "bad"}), wantStatus: http.StatusForbidden},
		{name: "guru list service error", handler: (*CbtSession).GuruAwareList, svc: &fakeCbtSessionService{listByTeacherErr: errDB}, req: guruRoute(http.MethodGet, "/api/cbt/sessions", "", validGuruClaims), wantStatus: http.StatusInternalServerError},
		{name: "guru results invalid id", handler: (*CbtSession).GuruAwareResults, req: guruRoute(http.MethodGet, "/api/cbt/sessions/bad/results", "", validGuruClaims, "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "guru results missing eid", handler: (*CbtSession).GuruAwareResults, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", jwt.MapClaims{"roles": []any{"guru"}}, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "guru results invalid eid", handler: (*CbtSession).GuruAwareResults, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", jwt.MapClaims{"roles": []any{"guru"}, "eid": "bad"}, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "guru results check error", handler: (*CbtSession).GuruAwareResults, svc: &fakeCbtSessionService{checkErr: errDB}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "guru results denied", handler: (*CbtSession).GuruAwareResults, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "guru results session error", handler: (*CbtSession).GuruAwareResults, svc: &fakeCbtSessionService{checkAllowed: true, getErr: errDB}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "guru results rows error", handler: (*CbtSession).GuruAwareResults, svc: &fakeCbtSessionService{checkAllowed: true, resultsByTeacherErr: errDB}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "guru participants invalid id", handler: (*CbtSession).GuruAwareParticipants, req: guruRoute(http.MethodGet, "/api/cbt/sessions/bad/participants", "", validGuruClaims, "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "guru participants missing eid", handler: (*CbtSession).GuruAwareParticipants, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", jwt.MapClaims{"roles": []any{"guru"}}, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "guru participants invalid eid", handler: (*CbtSession).GuruAwareParticipants, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", jwt.MapClaims{"roles": []any{"guru"}, "eid": "bad"}, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "guru participants check error", handler: (*CbtSession).GuruAwareParticipants, svc: &fakeCbtSessionService{checkErr: errDB}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "guru participants denied", handler: (*CbtSession).GuruAwareParticipants, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "guru participants no cbt role", handler: (*CbtSession).GuruAwareParticipants, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", jwt.MapClaims{"roles": []any{"staf"}}, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "enroll class alias invalid class", handler: (*CbtSession).EnrollClass, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-class", `{"scope_type":"class","class_id":"bad"}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "enroll class alias valid service error", handler: (*CbtSession).EnrollClass, svc: &fakeCbtSessionService{enrollClassErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/enroll-class", `{"scope_type":"class","class_id":"`+classID.String()+`"}`, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "enroll school invalid id", handler: (*CbtSession).EnrollSchool, req: adminRoute(http.MethodPost, "/api/cbt/sessions/bad/enroll-school", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "list rooms invalid id", handler: (*CbtSession).ListRooms, req: adminRoute(http.MethodGet, "/api/cbt/sessions/bad/rooms", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "list rooms guru denied", handler: (*CbtSession).ListRooms, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "shuffle invalid id", handler: (*CbtSession).ShuffleRooms, req: adminRoute(http.MethodPost, "/api/cbt/sessions/bad/rooms/shuffle", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "proctoring invalid id", handler: (*CbtSession).GetProctoringStatus, req: adminRoute(http.MethodGet, "/api/cbt/sessions/bad/proctoring", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "proctoring guru denied", handler: (*CbtSession).GetProctoringStatus, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "ungraded invalid id", handler: (*CbtSession).ListUngradedEssays, req: adminRoute(http.MethodGet, "/api/cbt/sessions/bad/essays/ungraded", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "ungraded guru denied", handler: (*CbtSession).ListUngradedEssays, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/essays/ungraded", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "generate tokens invalid id", handler: (*CbtSession).GenerateTokens, req: adminRoute(http.MethodPost, "/api/cbt/sessions/bad/tokens", "", "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "get minutes guru denied", handler: (*CbtSession).GetMinutes, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/minutes", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusForbidden},
		{name: "participant answers guru denied", handler: (*CbtSession).GetParticipantAnswers, svc: &fakeCbtSessionService{checkAllowed: false}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/answers", "", validGuruClaims, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakeCbtSessionService{checkAllowed: true}
			}
			rec := httptest.NewRecorder()
			tt.handler(&CbtSession{svc: svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}

}

func TestCbtSessionGuruAwareHandlersForwardTeacherScope(t *testing.T) {
	sessionID := handlerTestUUID(110)
	teacherID := handlerTestUUID(111)
	fake := &fakeCbtSessionService{checkAllowed: true}
	h := &CbtSession{svc: fake}
	guruReq := func(method, target, body string, pairs ...string) *http.Request {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
			"roles": []any{"guru"},
			"eid":   teacherID.String(),
		}))
		return withRouteParams(req, pairs...)
	}

	rec := httptest.NewRecorder()
	h.GuruAwareList(rec, guruReq(http.MethodGet, "/api/cbt/sessions", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("GuruAwareList() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listByTeacherID != teacherID {
		t.Fatalf("ListByTeacher id = %v, want %v", fake.listByTeacherID, teacherID)
	}

	rec = httptest.NewRecorder()
	h.GuruAwareResults(rec, guruReq(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GuruAwareResults() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.checkSessionID != sessionID || fake.checkTeacherID != teacherID || fake.resultsByTeacherSessionID != sessionID || fake.resultsByTeacherID != teacherID {
		t.Fatalf("GuruAwareResults args = check:%v/%v results:%v/%v, want session/teacher", fake.checkSessionID, fake.checkTeacherID, fake.resultsByTeacherSessionID, fake.resultsByTeacherID)
	}

	rec = httptest.NewRecorder()
	h.GuruAwareParticipants(rec, guruReq(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GuruAwareParticipants() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.checkSessionID != sessionID || fake.checkTeacherID != teacherID || fake.listParticipantsSessionID != sessionID || fake.listParticipantsTeacherID != teacherID {
		t.Fatalf("GuruAwareParticipants args = check:%v/%v list:%v/%v, want session/teacher/list", fake.checkSessionID, fake.checkTeacherID, fake.listParticipantsSessionID, fake.listParticipantsTeacherID)
	}
}

func TestCbtSessionGuruAwareHandlersAdminGuruRemainAdminWide(t *testing.T) {
	sessionID := handlerTestUUID(112)
	fake := &fakeCbtSessionService{checkAllowed: true}
	h := &CbtSession{svc: fake}
	adminGuruReq := func(method, target, body string, pairs ...string) *http.Request {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, jwt.MapClaims{
			"roles": []any{"admin", "guru"},
		}))
		return withRouteParams(req, pairs...)
	}

	rec := httptest.NewRecorder()
	h.GuruAwareList(rec, adminGuruReq(http.MethodGet, "/api/cbt/sessions", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("GuruAwareList() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.listCalled {
		t.Fatal("List() was not called for admin+guru branch")
	}
	if fake.listByTeacherID.Valid {
		t.Fatalf("ListByTeacher id = %v, want invalid admin-wide scope", fake.listByTeacherID)
	}

	rec = httptest.NewRecorder()
	h.GuruAwareResults(rec, adminGuruReq(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GuruAwareResults() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.getResultsSessionID != sessionID {
		t.Fatalf("GetResults session = %v, want %v", fake.getResultsSessionID, sessionID)
	}
	if fake.resultsByTeacherID.Valid || fake.resultsByTeacherSessionID.Valid {
		t.Fatalf("GetResultsByTeacher args = %v/%v, want invalid admin-wide scope", fake.resultsByTeacherSessionID, fake.resultsByTeacherID)
	}

	rec = httptest.NewRecorder()
	h.GuruAwareParticipants(rec, adminGuruReq(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GuruAwareParticipants() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listParticipantsSessionID != sessionID {
		t.Fatalf("ListParticipants session = %v, want %v", fake.listParticipantsSessionID, sessionID)
	}
	if fake.checkSessionID.Valid || fake.checkTeacherID.Valid {
		t.Fatalf("CheckTeacherAccess args = %v/%v, want invalid admin-wide scope", fake.checkSessionID, fake.checkTeacherID)
	}
}
