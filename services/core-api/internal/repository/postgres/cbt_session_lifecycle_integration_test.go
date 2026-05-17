package db

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationCbtSessionLifecycle(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := "SL-" + integrationSuffix()

	year, err := q.CreateAcademicYear(ctx, CreateAcademicYearParams{
		Name:      "CBT Session Lifecycle Year " + suffix,
		StartDate: pgDate(2026, time.July, 1),
		EndDate:   pgDate(2027, time.June, 30),
		IsActive:  false,
	})
	if err != nil {
		t.Fatalf("create academic year: %v", err)
	}
	class, err := q.CreateSchoolClass(ctx, CreateSchoolClassParams{
		AcademicYearID: year.ID,
		Code:           "CBT-SL-" + suffix,
		Name:           "CBT Session Lifecycle Class " + suffix,
		Level:          "VII",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create school class: %v", err)
	}
	subject, err := q.CreateSubject(ctx, CreateSubjectParams{
		Code:                "CBT-SL-SUB-" + suffix,
		Name:                "CBT Session Lifecycle Subject " + suffix,
		Category:            "umum",
		IsAssessmentSubject: true,
		IsReportSubject:     true,
		IsScheduleActivity:  false,
		CountsForRanking:    true,
		IsLocalContent:      false,
		IsChoiceSubject:     false,
		DefaultWeeklyHours:  2,
		DisplayOrder:        9999,
		IsActive:            true,
	})
	if err != nil {
		t.Fatalf("create subject: %v", err)
	}

	employeeID, userID := seedCbtIntegrationEmployeeAndUser(t, ctx, tdb, suffix)
	seedCbtIntegrationClassAssignment(t, ctx, tdb, class.ID, subject.ID, employeeID)
	student := seedCbtIntegrationStudent(t, ctx, q, class.ID, suffix)

	event, err := q.CreateCbtExamEvent(ctx, CreateCbtExamEventParams{
		Title:          "CBT Session Lifecycle Event " + suffix,
		ExamType:       CbtExamTypeUts,
		Scope:          "grade",
		TargetLevels:   []string{"VII"},
		AcademicYearID: year.ID,
		Status:         "active",
	})
	if err != nil {
		t.Fatalf("create cbt exam event: %v", err)
	}
	question := createCbtIntegrationQuestion(t, ctx, q, event.ID, subject.ID, suffix)
	pkg, err := q.CreateCbtPackage(ctx, CreateCbtPackageParams{
		EventID:            event.ID,
		SubjectID:          subject.ID,
		Title:              "CBT Session Lifecycle Package " + suffix,
		Description:        "session lifecycle integration package",
		DurationMinutes:    45,
		RandomizeQuestions: false,
		IsActive:           true,
		SourceMode:         "teacher_class",
		RandomizeOptions:   false,
		DrawPgCount:        1,
		DrawEssayCount:     0,
		RandomSeed:         "seed-" + suffix,
		CompositionLog:     []byte(`{"source":"session_lifecycle_integration"}`),
	})
	if err != nil {
		t.Fatalf("create cbt package: %v", err)
	}
	if rows, err := q.AddCbtPackageQuestion(ctx, AddCbtPackageQuestionParams{PackageID: pkg.ID, QuestionID: question.ID, Position: 1, Points: 4}); err != nil || rows != 1 {
		t.Fatalf("add cbt package question rows=%d err=%v", rows, err)
	}

	sessionStart := time.Now().UTC().Add(30 * time.Minute).Truncate(time.Second)
	session, err := q.CreateCbtExamSession(ctx, CreateCbtExamSessionParams{
		PackageID:       pkg.ID,
		ClassID:         class.ID,
		EventID:         event.ID,
		ScopeType:       "class",
		ScopeRef:        class.ID.String(),
		MixPolicy:       "same_class",
		AssignmentMode:  "manual",
		AllowCrossGrade: false,
		IsSpecialEvent:  false,
		Title:           "CBT Session Lifecycle " + suffix,
		ScheduledStart:  pgTimestamptz(sessionStart),
		ScheduledEnd:    pgTimestamptz(sessionStart.Add(45 * time.Minute)),
		Status:          CbtSessionStatusEnumDraft,
	})
	if err != nil {
		t.Fatalf("create cbt exam session: %v", err)
	}
	if err := q.EnrollClassToSession(ctx, EnrollClassToSessionParams{SessionID: session.ID, ClassID: class.ID}); err != nil {
		t.Fatalf("enroll class to session: %v", err)
	}
	if err := q.GenerateTokensForSession(ctx, session.ID); err != nil {
		t.Fatalf("generate tokens for session: %v", err)
	}

	participants, err := q.ListCbtExamParticipants(ctx, session.ID)
	if err != nil {
		t.Fatalf("list cbt exam participants: %v", err)
	}
	if len(participants) != 1 || participants[0].StudentID != student.ID || participants[0].Token == "" {
		t.Fatalf("participants mismatch: %#v", participants)
	}
	participantID := participants[0].ID

	schoolRoomID := seedCbtIntegrationSchoolRoom(t, ctx, tdb, suffix)
	room, err := q.CreateCbtExamRoom(ctx, CreateCbtExamRoomParams{
		SessionID:        session.ID,
		SchoolRoomID:     schoolRoomID,
		RoomName:         "Lifecycle Room " + suffix,
		RoomNameSnapshot: "Lifecycle Room Snapshot " + suffix,
		Capacity:         1,
	})
	if err != nil {
		t.Fatalf("create cbt exam room: %v", err)
	}
	if _, err := q.CreateCbtRoomProctor(ctx, CreateCbtRoomProctorParams{ExamRoomID: room.ID, EmployeeID: employeeID, Role: "utama", AssignedBy: userID}); err != nil {
		t.Fatalf("create cbt room proctor: %v", err)
	}
	if err := q.AssignParticipantSeat(ctx, AssignParticipantSeatParams{ID: participantID, RoomID: room.ID, SeatNo: pgtype.Int4{Int32: 1, Valid: true}}); err != nil {
		t.Fatalf("assign participant seat: %v", err)
	}

	assertSessionLifecycleReadModels(t, ctx, q, session.ID, room.ID, employeeID, participantID)

	scheduled, err := q.UpdateCbtExamSessionStatus(ctx, UpdateCbtExamSessionStatusParams{ID: session.ID, Status: CbtSessionStatusEnumScheduled})
	if err != nil {
		t.Fatalf("update session status scheduled: %v", err)
	}
	if scheduled.Status != CbtSessionStatusEnumScheduled {
		t.Fatalf("scheduled status = %s", scheduled.Status)
	}
	token, err := q.RegenerateParticipantToken(ctx, participantID)
	if err != nil {
		t.Fatalf("regenerate participant token: %v", err)
	}
	if token.Token == "" || token.Token == participants[0].Token {
		t.Fatalf("regenerated token not changed: before=%q after=%q", participants[0].Token, token.Token)
	}
	active, err := q.UpdateCbtExamSessionStatus(ctx, UpdateCbtExamSessionStatusParams{ID: session.ID, Status: CbtSessionStatusEnumActive})
	if err != nil {
		t.Fatalf("update session status active: %v", err)
	}
	if active.Status != CbtSessionStatusEnumActive {
		t.Fatalf("active status = %s", active.Status)
	}

	if _, err := q.UpdateParticipantLogin(ctx, UpdateParticipantLoginParams{ID: participantID, DeviceFingerprint: pgtype.Text{String: "device-" + suffix, Valid: true}, LoginIp: pgtype.Text{String: "127.0.0.1", Valid: true}}); err != nil {
		t.Fatalf("update participant login: %v", err)
	}
	if err := q.UpdateParticipantHeartbeat(ctx, participantID); err != nil {
		t.Fatalf("update participant heartbeat: %v", err)
	}
	portalParticipant, err := q.GetStudentPortalCbtParticipant(ctx, GetStudentPortalCbtParticipantParams{ParticipantID: participantID, StudentID: student.ID})
	if err != nil {
		t.Fatalf("get student portal participant: %v", err)
	}
	if portalParticipant.ParticipantID != participantID || !portalParticipant.RoomID.Valid || portalParticipant.RoomName != room.RoomName {
		t.Fatalf("student portal participant mismatch: %#v", portalParticipant)
	}

	if _, err := q.SetParticipantRuntimePlanIfEmpty(ctx, SetParticipantRuntimePlanIfEmptyParams{ID: participantID, QuestionOrder: []byte(`[` + `"` + question.ID.String() + `"` + `]`), OptionOrder: []byte(`{"A":["A","B"]}`), QuestionDrawLog: []byte(`{"source":"integration"}`)}); err != nil {
		t.Fatalf("set participant runtime plan: %v", err)
	}
	belongs, err := q.QuestionBelongsToParticipantPackage(ctx, QuestionBelongsToParticipantPackageParams{ID: participantID, QuestionID: question.ID})
	if err != nil {
		t.Fatalf("question belongs to participant package: %v", err)
	}
	if !belongs {
		t.Fatalf("question %v should belong to participant %v package", question.ID, participantID)
	}
	if rows, err := q.UpsertStudentAnswer(ctx, UpsertStudentAnswerParams{ParticipantID: participantID, QuestionID: question.ID, Answer: "A"}); err != nil || rows != 1 {
		t.Fatalf("upsert student answer rows=%d err=%v", rows, err)
	}
	if err := q.UpdateAnswerCorrectness(ctx, session.ID); err != nil {
		t.Fatalf("update answer correctness: %v", err)
	}
	answers, err := q.GetParticipantAnswers(ctx, participantID)
	if err != nil {
		t.Fatalf("get participant answers: %v", err)
	}
	if len(answers) != 1 || answers[0].QuestionID != question.ID || answers[0].Answer != "A" || !answers[0].IsCorrect.Bool {
		t.Fatalf("participant answers mismatch: %#v", answers)
	}
	if hasAnswer, err := q.HasSessionAnswer(ctx, HasSessionAnswerParams{SessionID: session.ID, ID: answers[0].ID}); err != nil || !hasAnswer {
		t.Fatalf("has session answer = %v err=%v", hasAnswer, err)
	}
	if hasAnswer, err := q.HasSessionAnswerByTeacher(ctx, HasSessionAnswerByTeacherParams{TeacherEmployeeID: employeeID, SessionID: session.ID, AnswerID: answers[0].ID}); err != nil || !hasAnswer {
		t.Fatalf("has session answer by teacher = %v err=%v", hasAnswer, err)
	}

	proctorEvent, err := q.CreateCbtParticipantProctorEvent(ctx, CreateCbtParticipantProctorEventParams{
		ParticipantID:         participantID,
		EventType:             "app_switch",
		EventData:             []byte(`{"source":"integration","count":1}`),
		Severity:              "warning",
		Category:              "integrity",
		RiskDelta:             25,
		DedupKey:              "app-switch-" + suffix,
		CorrelationID:         "corr-" + suffix,
		OriginalEventAt:       pgTimestamptz(time.Now().UTC()),
		RequiresNote:          true,
		ActorUserID:           userID,
		ActorUsernameSnapshot: "cbt_" + suffix,
		ActorEmployeeID:       employeeID,
		RequestID:             "req-" + suffix,
		SourceIp:              "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("create cbt participant proctor event: %v", err)
	}
	if _, err := q.CreateCbtProctorAction(ctx, CreateCbtProctorActionParams{SessionID: session.ID, RoomID: room.ID, ParticipantID: participantID, EventID: proctorEvent.ID, ActionType: "warn_student", Reason: "integration", Notes: "warned from integration", ActorUserID: userID, ActorUsernameSnapshot: "cbt_" + suffix, ActorEmployeeID: employeeID, RequestID: "req-action-" + suffix, SourceIp: "127.0.0.1"}); err != nil {
		t.Fatalf("create cbt proctor action: %v", err)
	}
	if _, err := q.AcknowledgeCbtProctorEvent(ctx, AcknowledgeCbtProctorEventParams{AcknowledgedBy: userID, AcknowledgeNote: "acknowledged from integration", ID: proctorEvent.ID}); err != nil {
		t.Fatalf("acknowledge cbt proctor event: %v", err)
	}
	if err := q.IncrementParticipantAppSwitch(ctx, participantID); err != nil {
		t.Fatalf("increment participant app switch: %v", err)
	}
	risk, err := q.UpdateCbtParticipantProctorRisk(ctx, UpdateCbtParticipantProctorRiskParams{ViolationCount: 1, RiskScore: 25, RiskLevel: "warning", LockedAt: pgtype.Timestamptz{}, LockedReason: pgtype.Text{}, SuspiciousFlag: true, AppSwitchIncrement: 0, ScreenshotIncrement: 0, LastLocalSaveAt: pgTimestamptz(time.Now().UTC()), LastSyncedAt: pgTimestamptz(time.Now().UTC()), PendingAnswerCount: 0, SyncState: "synced", ID: participantID})
	if err != nil {
		t.Fatalf("update cbt participant proctor risk: %v", err)
	}
	if risk.RiskLevel != "warning" || risk.RiskScore != 25 {
		t.Fatalf("risk mismatch: %#v", risk)
	}
	locked, err := q.HoldParticipantAccessForProctor(ctx, HoldParticipantAccessForProctorParams{LockedReason: pgtype.Text{String: "integration hold", Valid: true}, ID: participantID})
	if err != nil {
		t.Fatalf("hold participant access for proctor: %v", err)
	}
	if locked.RiskLevel != "locked" || !locked.LockedAt.Valid {
		t.Fatalf("locked mismatch: %#v", locked)
	}
	unlocked, err := q.UnlockParticipantAccessForProctor(ctx, participantID)
	if err != nil {
		t.Fatalf("unlock participant access for proctor: %v", err)
	}
	if unlocked.LockedAt.Valid || unlocked.RiskLevel != "warning" {
		t.Fatalf("unlocked mismatch: %#v", unlocked)
	}
	if _, err := q.ResetParticipantDeviceBindingForProctor(ctx, participantID); err != nil {
		t.Fatalf("reset participant device binding for proctor: %v", err)
	}

	assertProctorReadModels(t, ctx, q, session.ID, room.ID, participantID, proctorEvent.ID, "app-switch-"+suffix)

	submitted, err := q.SubmitParticipantExam(ctx, participantID)
	if err != nil {
		t.Fatalf("submit participant exam: %v", err)
	}
	if !submitted.SubmittedAt.Valid {
		t.Fatalf("submitted participant missing submitted_at: %#v", submitted)
	}
	if err := q.UpdateParticipantScores(ctx, session.ID); err != nil {
		t.Fatalf("update participant scores: %v", err)
	}
	results, err := q.GetSessionResults(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session results: %v", err)
	}
	if len(results) != 1 || results[0].ParticipantID != participantID || !results[0].SubmittedAt.Valid {
		t.Fatalf("session results mismatch: %#v", results)
	}
	teacherResults, err := q.GetSessionResultsByTeacher(ctx, GetSessionResultsByTeacherParams{SessionID: session.ID, TeacherEmployeeID: employeeID})
	if err != nil {
		t.Fatalf("get session results by teacher: %v", err)
	}
	if len(teacherResults) != 1 || teacherResults[0].ParticipantID != participantID || teacherResults[0].CorrectAnswers != 1 {
		t.Fatalf("teacher results mismatch: %#v", teacherResults)
	}
	analysis, err := q.GetSessionItemAnalysis(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session item analysis: %v", err)
	}
	if len(analysis) != 1 || analysis[0].QuestionID != question.ID || analysis[0].CorrectCount != 1 {
		t.Fatalf("item analysis mismatch: %#v", analysis)
	}
	remedial, err := q.ListCbtSessionRemedialCandidates(ctx, ListCbtSessionRemedialCandidatesParams{SessionID: session.ID, Threshold: pgtype.Numeric{Int: big.NewInt(101), Valid: true}})
	if err != nil {
		t.Fatalf("list cbt session remedial candidates: %v", err)
	}
	if len(remedial) != 1 || remedial[0].ParticipantID != participantID || remedial[0].QuestionCount != 1 {
		t.Fatalf("remedial candidates mismatch: %#v", remedial)
	}
	if count, err := q.FinalizeOverdueParticipants(ctx, session.ID); err != nil || count != 0 {
		t.Fatalf("finalize overdue participants count=%d err=%v", count, err)
	}
	if _, err := q.UpdateCbtExamSessionStatus(ctx, UpdateCbtExamSessionStatusParams{ID: session.ID, Status: CbtSessionStatusEnumFinished}); err != nil {
		t.Fatalf("update session status finished: %v", err)
	}
}

func assertSessionLifecycleReadModels(t *testing.T, ctx context.Context, q *Queries, sessionID, roomID, employeeID, participantID pgtype.UUID) {
	t.Helper()
	if hasRoom, err := q.HasSessionRoom(ctx, HasSessionRoomParams{SessionID: sessionID, ID: roomID}); err != nil || !hasRoom {
		t.Fatalf("has session room = %v err=%v", hasRoom, err)
	}
	if hasParticipant, err := q.HasSessionParticipant(ctx, HasSessionParticipantParams{SessionID: sessionID, ID: participantID}); err != nil || !hasParticipant {
		t.Fatalf("has session participant = %v err=%v", hasParticipant, err)
	}
	if hasParticipant, err := q.HasSessionParticipantByTeacher(ctx, HasSessionParticipantByTeacherParams{TeacherEmployeeID: employeeID, SessionID: sessionID, ParticipantID: participantID}); err != nil || !hasParticipant {
		t.Fatalf("has session participant by teacher = %v err=%v", hasParticipant, err)
	}
	if hasRoomParticipant, err := q.HasSessionRoomParticipant(ctx, HasSessionRoomParticipantParams{SessionID: sessionID, RoomID: roomID, ParticipantID: participantID}); err != nil || !hasRoomParticipant {
		t.Fatalf("has session room participant = %v err=%v", hasRoomParticipant, err)
	}
	if hasRoomProctor, err := q.HasSessionRoomProctor(ctx, HasSessionRoomProctorParams{SessionID: sessionID, RoomID: roomID, EmployeeID: employeeID}); err != nil || !hasRoomProctor {
		t.Fatalf("has session room proctor = %v err=%v", hasRoomProctor, err)
	}
	if overlap, err := q.HasOverlappingCbtRoomProctor(ctx, HasOverlappingCbtRoomProctorParams{SessionID: sessionID, EmployeeID: employeeID, ExamRoomID: roomID}); err != nil || overlap {
		t.Fatalf("has overlapping room proctor = %v err=%v", overlap, err)
	}
	rooms, err := q.ListCbtExamRooms(ctx, sessionID)
	if err != nil {
		t.Fatalf("list cbt exam rooms: %v", err)
	}
	if len(rooms) != 1 || rooms[0].ID != roomID || rooms[0].ParticipantCount != 1 || rooms[0].ProctorCount != 1 {
		t.Fatalf("rooms mismatch: %#v", rooms)
	}
	proctorRooms, err := q.ListCbtProctorRooms(ctx, ListCbtProctorRoomsParams{IncludeAll: false, EmployeeID: employeeID})
	if err != nil {
		t.Fatalf("list cbt proctor rooms: %v", err)
	}
	if len(proctorRooms) != 1 || proctorRooms[0].ID != roomID || proctorRooms[0].ActorRole != "utama" {
		t.Fatalf("proctor rooms mismatch: %#v", proctorRooms)
	}
	roomProctors, err := q.ListCbtRoomProctors(ctx, roomID)
	if err != nil {
		t.Fatalf("list cbt room proctors: %v", err)
	}
	if len(roomProctors) != 1 || roomProctors[0].EmployeeID != employeeID || roomProctors[0].Role != "utama" {
		t.Fatalf("room proctors mismatch: %#v", roomProctors)
	}
	readiness, err := q.GetCbtSessionRoomReadiness(ctx, sessionID)
	if err != nil {
		t.Fatalf("get cbt session room readiness: %v", err)
	}
	if readiness.RoomCount != 1 || readiness.ParticipantCount != 1 || readiness.AssignedParticipantCount != 1 || readiness.RoomsWithoutProctor != 0 {
		t.Fatalf("room readiness mismatch: %#v", readiness)
	}
	participantsByRoom, err := q.ListParticipantsByRoom(ctx, sessionID)
	if err != nil {
		t.Fatalf("list participants by room: %v", err)
	}
	if len(participantsByRoom) != 1 || participantsByRoom[0].ID != participantID || !participantsByRoom[0].SeatNo.Valid {
		t.Fatalf("participants by room mismatch: %#v", participantsByRoom)
	}
}

func assertProctorReadModels(t *testing.T, ctx context.Context, q *Queries, sessionID, roomID, participantID, eventID pgtype.UUID, dedupKey string) {
	t.Helper()
	if eventScope, err := q.GetCbtProctorEventScope(ctx, eventID); err != nil || eventScope.ID != eventID || eventScope.ParticipantID != participantID {
		t.Fatalf("get cbt proctor event scope = %#v err=%v", eventScope, err)
	}
	if participantScope, err := q.GetCbtParticipantProctorScope(ctx, GetCbtParticipantProctorScopeParams{SessionID: sessionID, ParticipantID: participantID}); err != nil || participantScope.ParticipantID != participantID || participantScope.RoomName == "" {
		t.Fatalf("get cbt participant proctor scope = %#v err=%v", participantScope, err)
	}
	if risk, err := q.GetCbtParticipantRiskForUpdate(ctx, participantID); err != nil || risk.ID != participantID || risk.RiskScore != 25 {
		t.Fatalf("get cbt participant risk for update = %#v err=%v", risk, err)
	}
	recent, err := q.GetRecentCbtProctorEventByDedupKey(ctx, GetRecentCbtProctorEventByDedupKeyParams{ParticipantID: participantID, DedupKey: dedupKey, SinceAt: pgTimestamptz(time.Now().UTC().Add(-time.Hour))})
	if err != nil {
		t.Fatalf("get recent cbt proctor event by dedup key: %v", err)
	}
	if recent.ID != eventID {
		t.Fatalf("recent dedup event mismatch: %#v", recent)
	}
	sessionEvents, err := q.ListCbtProctorEventsBySession(ctx, ListCbtProctorEventsBySessionParams{SessionID: sessionID, ParticipantID: pgtype.UUID{}, LimitCount: 10})
	if err != nil {
		t.Fatalf("list cbt proctor events by session: %v", err)
	}
	if len(sessionEvents) != 1 || sessionEvents[0].ID != eventID || !sessionEvents[0].AcknowledgedAt.Valid {
		t.Fatalf("session proctor events mismatch: %#v", sessionEvents)
	}
	roomEvents, err := q.ListCbtProctorEventsByRoom(ctx, ListCbtProctorEventsByRoomParams{SessionID: sessionID, RoomID: roomID, ParticipantID: pgtype.UUID{}, LimitCount: 10})
	if err != nil {
		t.Fatalf("list cbt proctor events by room: %v", err)
	}
	if len(roomEvents) != 1 || roomEvents[0].ID != eventID || roomEvents[0].RoomID != roomID {
		t.Fatalf("room proctor events mismatch: %#v", roomEvents)
	}
	sessionActions, err := q.ListCbtProctorActionsBySession(ctx, ListCbtProctorActionsBySessionParams{SessionID: sessionID, LimitCount: 10})
	if err != nil {
		t.Fatalf("list cbt proctor actions by session: %v", err)
	}
	if len(sessionActions) != 1 || sessionActions[0].EventID != eventID || sessionActions[0].ActionType != "warn_student" {
		t.Fatalf("session proctor actions mismatch: %#v", sessionActions)
	}
	roomActions, err := q.ListCbtProctorActionsByRoom(ctx, ListCbtProctorActionsByRoomParams{SessionID: sessionID, RoomID: roomID, LimitCount: 10})
	if err != nil {
		t.Fatalf("list cbt proctor actions by room: %v", err)
	}
	if len(roomActions) != 1 || roomActions[0].EventID != eventID {
		t.Fatalf("room proctor actions mismatch: %#v", roomActions)
	}
	sessionParticipantEvents, err := q.ListSessionParticipantEvents(ctx, ListSessionParticipantEventsParams{SessionID: sessionID, ParticipantID: pgtype.UUID{}, RoomID: pgtype.UUID{}, LimitCount: 10})
	if err != nil {
		t.Fatalf("list session participant events: %v", err)
	}
	if len(sessionParticipantEvents) != 1 || sessionParticipantEvents[0].ID != eventID {
		t.Fatalf("session participant events mismatch: %#v", sessionParticipantEvents)
	}
	status, err := q.GetSessionProctoringStatus(ctx, GetSessionProctoringStatusParams{SessionID: sessionID, RoomID: pgtype.UUID{}})
	if err != nil {
		t.Fatalf("get session proctoring status: %v", err)
	}
	if len(status) != 1 || status[0].ParticipantID != participantID || status[0].SuspiciousFlag != true {
		t.Fatalf("session proctoring status mismatch: %#v", status)
	}
	recap, err := q.GetCbtSessionOperationalRecap(ctx, sessionID)
	if err != nil {
		t.Fatalf("get cbt session operational recap: %v", err)
	}
	if recap.ParticipantCount != 1 || recap.IncidentEventCount != 1 || recap.AppSwitchCount != 1 {
		t.Fatalf("session operational recap mismatch: %#v", recap)
	}
	roomRecaps, err := q.ListCbtSessionRoomOperationalRecap(ctx, sessionID)
	if err != nil {
		t.Fatalf("list cbt session room operational recap: %v", err)
	}
	if len(roomRecaps) != 1 || roomRecaps[0].RoomID != roomID || roomRecaps[0].IncidentEventCount != 1 {
		t.Fatalf("room operational recap mismatch: %#v", roomRecaps)
	}
}

func seedCbtIntegrationSchoolRoom(t *testing.T, ctx context.Context, tdb *integrationTestDB, suffix string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := tdb.Pool.QueryRow(ctx, `
		INSERT INTO school_rooms (code, name, building, room_type, default_capacity, exam_capacity, network_ready, power_ready)
		VALUES ($1, $2, 'Integration Building', 'kelas', 1, 1, TRUE, TRUE)
		RETURNING id
	`, "CBT-SL-ROOM-"+suffix, "CBT Session Lifecycle Room "+suffix).Scan(&id); err != nil {
		t.Fatalf("seed school room: %v", err)
	}
	return id
}
