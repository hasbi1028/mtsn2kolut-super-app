package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	deleteID  pgtype.UUID
	deleteErr error

	listParticipantsSessionID pgtype.UUID
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

	regenerateParticipantID pgtype.UUID
	regenerateErr           error

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

	proctoringSessionID      pgtype.UUID
	roomDashboardID          pgtype.UUID
	proctorRoomsEmployeeID   pgtype.UUID
	proctorRoomsIncludeAll   bool
	roomProctorSessionID     pgtype.UUID
	roomProctorRoomID        pgtype.UUID
	roomProctorEmployeeID    pgtype.UUID
	roomProctorAllowed       bool
	roomParticipantSessionID pgtype.UUID
	roomParticipantRoomID    pgtype.UUID
	roomParticipantID        pgtype.UUID
	roomParticipantAllowed   bool
	roomProctoringSessionID  pgtype.UUID
	roomProctoringRoomID     pgtype.UUID
	roomEventsSessionID      pgtype.UUID
	roomEventsRoomID         pgtype.UUID
	proctoringErr            error
	handoverRoomID           pgtype.UUID
	operationalRecapID       pgtype.UUID
	saveHandoverRoomID       pgtype.UUID
	saveHandoverUpdatedBy    pgtype.UUID
	saveHandoverInput        service.SaveCbtRoomHandoverInput
	saveHandoverErr          error
	lockHandoverRoomID       pgtype.UUID
	lockHandoverLockedBy     pgtype.UUID
	lockHandoverErr          error

	flagParticipantID pgtype.UUID
	flagValue         bool
	flagErr           error

	ungradedSessionID pgtype.UUID
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

func (f *fakeCbtSessionService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeCbtSessionService) ListParticipants(_ context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error) {
	f.listParticipantsSessionID = sessionID
	if f.listParticipantsErr != nil {
		return nil, f.listParticipantsErr
	}
	return []db.ListCbtExamParticipantsRow{}, nil
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
	return []db.ListCbtRoomProctorsRow{{ExamRoomID: roomID, Nama: "Pengawas"}}, nil
}

func (f *fakeCbtSessionService) ReplaceRoomProctors(_ context.Context, roomID, assignedBy, primaryEmployeeID pgtype.UUID, employeeIDs []pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error) {
	return []db.ListCbtRoomProctorsRow{{ExamRoomID: roomID, EmployeeID: primaryEmployeeID}}, nil
}

func (f *fakeCbtSessionService) RoomReadiness(_ context.Context, sessionID pgtype.UUID) (db.GetCbtSessionRoomReadinessRow, error) {
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

func TestCbtSessionAdminLifecycleHandlersForwardValidRequests(t *testing.T) {
	sessionID := handlerTestUUID(90)
	packageID := handlerTestUUID(91)
	classID := handlerTestUUID(92)
	fake := &fakeCbtSessionService{}
	h := &CbtSession{svc: fake}

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
		{name: "create service error", handler: (*CbtSession).Create, svc: &fakeCbtSessionService{createErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions", validCreate), wantStatus: http.StatusInternalServerError},
		{name: "create with event service error", handler: (*CbtSession).Create, svc: &fakeCbtSessionService{createErr: errDB}, req: adminRoute(http.MethodPost, "/api/cbt/sessions", `{"package_id":"`+packageID.String()+`","class_id":"`+classID.String()+`","event_id":"`+eventID.String()+`","scheduled_start":"2026-05-01T08:00:00Z","scheduled_end":"2026-05-01T09:00:00Z"}`), wantStatus: http.StatusInternalServerError},
		{name: "update status invalid id", handler: (*CbtSession).UpdateStatus, req: adminRoute(http.MethodPatch, "/api/cbt/sessions/bad/status", `{"status":"active"}`, "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "update status invalid json", handler: (*CbtSession).UpdateStatus, req: adminRoute(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/status", `{`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "update status service error", handler: (*CbtSession).UpdateStatus, svc: &fakeCbtSessionService{updateStatusErr: errDB}, req: adminRoute(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/status", `{"status":"active"}`, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
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
		{name: "score invalid id", handler: (*CbtSession).ScoreSession, req: adminRoute(http.MethodPost, "/api/cbt/sessions/bad/score", "", "id", "bad"), wantStatus: http.StatusBadRequest},
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
		{name: "guru results session error", handler: (*CbtSession).GuruAwareResults, svc: &fakeCbtSessionService{getErr: errDB}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
		{name: "guru results rows error", handler: (*CbtSession).GuruAwareResults, svc: &fakeCbtSessionService{resultsByTeacherErr: errDB}, req: guruRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/results", "", validGuruClaims, "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
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
	if fake.resultsByTeacherSessionID != sessionID || fake.resultsByTeacherID != teacherID {
		t.Fatalf("GetResultsByTeacher args = %v/%v, want session/teacher", fake.resultsByTeacherSessionID, fake.resultsByTeacherID)
	}

	rec = httptest.NewRecorder()
	h.GuruAwareParticipants(rec, guruReq(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/participants", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GuruAwareParticipants() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.checkSessionID != sessionID || fake.checkTeacherID != teacherID || fake.listParticipantsSessionID != sessionID {
		t.Fatalf("GuruAwareParticipants args = check:%v/%v list:%v, want session/teacher/list", fake.checkSessionID, fake.checkTeacherID, fake.listParticipantsSessionID)
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
