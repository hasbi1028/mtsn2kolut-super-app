package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationCbtRepositoryQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	year, err := q.CreateAcademicYear(ctx, CreateAcademicYearParams{
		Name:      "CBT Integration Year " + suffix,
		StartDate: pgDate(2026, time.July, 1),
		EndDate:   pgDate(2027, time.June, 30),
		IsActive:  false,
	})
	if err != nil {
		t.Fatalf("create academic year: %v", err)
	}

	class, err := q.CreateSchoolClass(ctx, CreateSchoolClassParams{
		AcademicYearID: year.ID,
		Code:           "CBT-" + suffix,
		Name:           "CBT Class " + suffix,
		Level:          "VII",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create school class: %v", err)
	}

	subject, err := q.CreateSubject(ctx, CreateSubjectParams{
		Code:                "CBT-SUB-" + suffix,
		Name:                "CBT Subject " + suffix,
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
		Title:          "CBT Event " + suffix,
		ExamType:       CbtExamTypeUts,
		Scope:          "grade",
		TargetLevels:   []string{"VII"},
		AcademicYearID: year.ID,
		Status:         "draft",
	})
	if err != nil {
		t.Fatalf("create cbt exam event: %v", err)
	}
	if event.Title != "CBT Event "+suffix || event.ExamType != CbtExamTypeUts || len(event.TargetLevels) != 1 || event.TargetLevels[0] != "VII" {
		t.Fatalf("created event mismatch: %#v", event)
	}

	gotEvent, err := q.GetCbtExamEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("get cbt exam event: %v", err)
	}
	if gotEvent.ID != event.ID || gotEvent.AcademicYearID != year.ID || gotEvent.AcademicYearName.String != year.Name {
		t.Fatalf("got event mismatch: %#v", gotEvent)
	}

	updatedEvent, err := q.UpdateCbtExamEvent(ctx, UpdateCbtExamEventParams{
		ID:             event.ID,
		Title:          "CBT Event Updated " + suffix,
		ExamType:       CbtExamTypeUas,
		Scope:          "grade",
		TargetLevels:   []string{"VII"},
		AcademicYearID: year.ID,
	})
	if err != nil {
		t.Fatalf("update cbt exam event: %v", err)
	}
	if updatedEvent.Title != "CBT Event Updated "+suffix || updatedEvent.ExamType != CbtExamTypeUas {
		t.Fatalf("updated event mismatch: %#v", updatedEvent)
	}
	if _, err := q.UpdateCbtExamEventStatus(ctx, UpdateCbtExamEventStatusParams{ID: event.ID, Status: "active"}); err != nil {
		t.Fatalf("update cbt exam event status: %v", err)
	}

	member, err := q.CreateCbtEventMember(ctx, CreateCbtEventMemberParams{
		EventID:    event.ID,
		UserID:     userID,
		EmployeeID: employeeID,
		SubjectID:  subject.ID,
		Role:       CbtEventMemberRolePembuatSoal,
	})
	if err != nil {
		t.Fatalf("create cbt event member: %v", err)
	}
	gotMember, err := q.GetCbtEventMember(ctx, GetCbtEventMemberParams{ID: member.ID, EventID: event.ID})
	if err != nil {
		t.Fatalf("get cbt event member: %v", err)
	}
	if gotMember.Username != "cbt_"+suffix || gotMember.EmployeeName != "CBT Teacher "+suffix || gotMember.SubjectCode != subject.Code {
		t.Fatalf("got member mismatch: %#v", gotMember)
	}
	members, err := q.ListCbtEventMembers(ctx, event.ID)
	if err != nil {
		t.Fatalf("list cbt event members: %v", err)
	}
	if len(members) != 1 || members[0].ID != member.ID {
		t.Fatalf("members = %#v, want member %v", members, member.ID)
	}
	membersByUser, err := q.ListCbtEventMembersByUser(ctx, userID)
	if err != nil {
		t.Fatalf("list cbt event members by user: %v", err)
	}
	if len(membersByUser) != 1 || membersByUser[0].ID != member.ID {
		t.Fatalf("members by user = %#v, want member %v", membersByUser, member.ID)
	}
	membersByUsername, err := q.ListCbtEventMembersByUsername(ctx, "cbt_"+suffix)
	if err != nil {
		t.Fatalf("list cbt event members by username: %v", err)
	}
	if len(membersByUsername) != 1 || membersByUsername[0].ID != member.ID {
		t.Fatalf("members by username = %#v, want member %v", membersByUsername, member.ID)
	}

	target, err := q.UpsertCbtEventSubjectTarget(ctx, UpsertCbtEventSubjectTargetParams{EventID: event.ID, SubjectID: subject.ID, TargetQuestions: 1})
	if err != nil {
		t.Fatalf("upsert cbt event subject target: %v", err)
	}
	if target.TargetQuestions != 1 {
		t.Fatalf("target questions = %d, want 1", target.TargetQuestions)
	}
	requirements, err := q.UpsertCbtEventQuestionRequirements(ctx, UpsertCbtEventQuestionRequirementsParams{
		EventID:      event.ID,
		ScopeMode:    "per_rombel",
		TargetPg:     1,
		TargetEssay:  0,
		StatusFilter: "published_only",
	})
	if err != nil {
		t.Fatalf("upsert cbt event question requirements: %v", err)
	}
	if requirements.TargetPg != 1 || requirements.ScopeMode != "per_rombel" {
		t.Fatalf("requirements mismatch: %#v", requirements)
	}
	gotRequirements, err := q.GetCbtEventQuestionRequirements(ctx, event.ID)
	if err != nil {
		t.Fatalf("get cbt event question requirements: %v", err)
	}
	if gotRequirements.TargetPg != 1 || gotRequirements.StatusFilter != "published_only" {
		t.Fatalf("got requirements mismatch: %#v", gotRequirements)
	}

	question := createCbtIntegrationQuestion(t, ctx, q, event.ID, subject.ID, suffix)
	questionByID, err := q.GetCbtQuestion(ctx, question.ID)
	if err != nil {
		t.Fatalf("get cbt question: %v", err)
	}
	if questionByID.ID != question.ID || questionByID.SubjectID != subject.ID || questionByID.AnswerKey != "A" {
		t.Fatalf("question by id mismatch: %#v", questionByID)
	}
	questionDetail, err := q.GetCbtQuestionDetail(ctx, question.ID)
	if err != nil {
		t.Fatalf("get cbt question detail: %v", err)
	}
	if questionDetail.ID != question.ID || questionDetail.EventID != event.ID || questionDetail.AuthorUsername != "cbt_"+suffix {
		t.Fatalf("question detail mismatch: %#v", questionDetail)
	}
	nextVersion, err := q.GetNextCbtQuestionVersionNumber(ctx, question.VersionGroupID)
	if err != nil {
		t.Fatalf("get next cbt question version number: %v", err)
	}
	if nextVersion != 2 {
		t.Fatalf("next version = %d, want 2", nextVersion)
	}
	questions, err := q.ListCbtQuestions(ctx, ListCbtQuestionsParams{IsAdmin: true, ActorUsername: "cbt_" + suffix, ActorUserID: userID, CanReviewAnswer: true, CanApproveAnswer: true})
	if err != nil {
		t.Fatalf("list cbt questions: %v", err)
	}
	if !cbtQuestionListed(questions, question.ID) {
		t.Fatalf("question %v not found in ListCbtQuestions", question.ID)
	}
	questionCounts, err := q.GetCbtQuestionSummaryCounts(ctx, GetCbtQuestionSummaryCountsParams{ActorUsername: "cbt_" + suffix, IsAdmin: true, ActorUserID: userID, CanUseInPackage: true})
	if err != nil {
		t.Fatalf("get cbt question summary counts: %v", err)
	}
	if questionCounts.Total < 1 || questionCounts.Published < 1 || questionCounts.PackageReady < 1 {
		t.Fatalf("question counts mismatch: %#v", questionCounts)
	}

	pkg, err := q.CreateCbtPackage(ctx, CreateCbtPackageParams{
		EventID:            event.ID,
		SubjectID:          subject.ID,
		Title:              "CBT Package " + suffix,
		Description:        "integration package",
		DurationMinutes:    90,
		RandomizeQuestions: true,
		IsActive:           true,
		SourceMode:         "teacher_class",
		RandomizeOptions:   true,
		DrawPgCount:        1,
		DrawEssayCount:     0,
		RandomSeed:         "seed-" + suffix,
		CompositionLog:     []byte(`{"source":"integration"}`),
	})
	if err != nil {
		t.Fatalf("create cbt package: %v", err)
	}
	if rows, err := q.AddCbtPackageQuestion(ctx, AddCbtPackageQuestionParams{PackageID: pkg.ID, QuestionID: question.ID, Position: 1, Points: 2}); err != nil || rows != 1 {
		t.Fatalf("add cbt package question rows=%d err=%v", rows, err)
	}
	packageDetail, err := q.GetCbtPackageDetail(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("get cbt package detail: %v", err)
	}
	if packageDetail.ID != pkg.ID || packageDetail.QuestionCount != 1 || packageDetail.SubjectID != subject.ID {
		t.Fatalf("package detail mismatch: %#v", packageDetail)
	}
	packageQuality, err := q.GetCbtPackageQuestionQuality(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("get cbt package question quality: %v", err)
	}
	if packageQuality.TotalQuestions != 1 || packageQuality.PublishedQuestions != 1 || packageQuality.UnpublishedQuestions != 0 {
		t.Fatalf("package quality mismatch: %#v", packageQuality)
	}
	packageQuestions, err := q.ListCbtPackageQuestionsByPackage(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("list cbt package questions by package: %v", err)
	}
	if len(packageQuestions) != 1 || packageQuestions[0].QuestionID != question.ID || packageQuestions[0].Points != 2 {
		t.Fatalf("package questions mismatch: %#v", packageQuestions)
	}
	examQuestions, err := q.GetExamQuestions(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("get exam questions: %v", err)
	}
	if len(examQuestions) != 1 || examQuestions[0].ID != question.ID || examQuestions[0].Code == "" {
		t.Fatalf("exam questions mismatch: %#v", examQuestions)
	}

	sessionStart := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Second)
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
		Title:           "CBT Session " + suffix,
		ScheduledStart:  pgTimestamptz(sessionStart),
		ScheduledEnd:    pgTimestamptz(sessionStart.Add(90 * time.Minute)),
		Status:          CbtSessionStatusEnumDraft,
	})
	if err != nil {
		t.Fatalf("create cbt exam session: %v", err)
	}
	gotSession, err := q.GetCbtExamSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("get cbt exam session: %v", err)
	}
	if gotSession.ID != session.ID || gotSession.PackageTitle != pkg.Title || gotSession.ClassCode != class.Code {
		t.Fatalf("got session mismatch: %#v", gotSession)
	}
	if err := q.EnrollClassToSession(ctx, EnrollClassToSessionParams{SessionID: session.ID, ClassID: class.ID}); err != nil {
		t.Fatalf("enroll class to session: %v", err)
	}
	participants, err := q.ListCbtExamParticipants(ctx, session.ID)
	if err != nil {
		t.Fatalf("list cbt exam participants: %v", err)
	}
	if len(participants) != 1 || participants[0].StudentID != student.ID || participants[0].Token == "" {
		t.Fatalf("participants mismatch: %#v", participants)
	}
	participantID := participants[0].ID

	if rows, err := q.UpsertStudentAnswer(ctx, UpsertStudentAnswerParams{ParticipantID: participantID, QuestionID: question.ID, Answer: "A"}); err != nil || rows != 1 {
		t.Fatalf("upsert student answer rows=%d err=%v", rows, err)
	}
	if err := q.UpdateParticipantAnswerCorrectness(ctx, participantID); err != nil {
		t.Fatalf("update participant answer correctness: %v", err)
	}
	if err := q.UpdateParticipantScores(ctx, participantID); err != nil {
		t.Fatalf("update participant scores: %v", err)
	}
	if _, err := q.SubmitParticipantExam(ctx, participantID); err != nil {
		t.Fatalf("submit participant exam: %v", err)
	}

	answers, err := q.GetParticipantAnswers(ctx, participantID)
	if err != nil {
		t.Fatalf("get participant answers: %v", err)
	}
	if len(answers) != 1 || answers[0].QuestionID != question.ID || answers[0].Answer != "A" {
		t.Fatalf("participant answers mismatch: %#v", answers)
	}
	sessionResults, err := q.GetSessionResults(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session results: %v", err)
	}
	if len(sessionResults) != 1 || sessionResults[0].ParticipantID != participantID || !sessionResults[0].SubmittedAt.Valid {
		t.Fatalf("session results mismatch: %#v", sessionResults)
	}
	eventResults, err := q.GetEventResults(ctx, event.ID)
	if err != nil {
		t.Fatalf("get event results: %v", err)
	}
	if len(eventResults) != 1 || eventResults[0].ParticipantID != participantID || !eventResults[0].SubmittedAt.Valid {
		t.Fatalf("event results mismatch: %#v", eventResults)
	}
	examCards, err := q.GetEventExamCards(ctx, event.ID)
	if err != nil {
		t.Fatalf("get event exam cards: %v", err)
	}
	if len(examCards) != 1 || examCards[0].ParticipantID != participantID || examCards[0].Nis != student.Nis {
		t.Fatalf("exam cards mismatch: %#v", examCards)
	}

	overview, err := q.GetCbtEventOverviewSummary(ctx, event.ID)
	if err != nil {
		t.Fatalf("get cbt event overview summary: %v", err)
	}
	if overview.MemberCount != 1 || overview.TargetSubjectCount != 1 || overview.TotalQuestions != 1 || overview.PackageCount != 1 || overview.SessionCount != 1 || overview.ParticipantCount != 1 || overview.SubmittedCount != 1 {
		t.Fatalf("overview mismatch: %#v", overview)
	}
	subjectTargets, err := q.ListCbtEventSubjectTargets(ctx, event.ID)
	if err != nil {
		t.Fatalf("list cbt event subject targets: %v", err)
	}
	if len(subjectTargets) != 1 || subjectTargets[0].SubjectID != subject.ID || subjectTargets[0].PublishedCount != 1 {
		t.Fatalf("subject targets mismatch: %#v", subjectTargets)
	}
	subjectMatrix, err := q.ListCbtEventSubjectMatrix(ctx, event.ID)
	if err != nil {
		t.Fatalf("list cbt event subject matrix: %v", err)
	}
	if len(subjectMatrix) != 1 || subjectMatrix[0].SubjectID != subject.ID || subjectMatrix[0].PackageCount != 1 || subjectMatrix[0].SessionCount != 1 {
		t.Fatalf("subject matrix mismatch: %#v", subjectMatrix)
	}
	readiness, err := q.ListCbtEventSessionsReadiness(ctx, event.ID)
	if err != nil {
		t.Fatalf("list cbt event sessions readiness: %v", err)
	}
	if len(readiness) != 1 || readiness[0].ID != session.ID || readiness[0].QuestionCount != 1 || readiness[0].ParticipantCount != 1 || readiness[0].SubmittedCount != 1 {
		t.Fatalf("session readiness mismatch: %#v", readiness)
	}
	packagesReadiness, err := q.ListCbtPackageReadiness(ctx, event.ID)
	if err != nil {
		t.Fatalf("list cbt package readiness: %v", err)
	}
	if len(packagesReadiness) != 1 || packagesReadiness[0].ID != pkg.ID || packagesReadiness[0].QuestionCount != 1 || packagesReadiness[0].SessionCount != 1 {
		t.Fatalf("package readiness mismatch: %#v", packagesReadiness)
	}
}

func seedCbtIntegrationEmployeeAndUser(t *testing.T, ctx context.Context, tdb *integrationTestDB, suffix string) (pgtype.UUID, pgtype.UUID) {
	t.Helper()

	var employeeID pgtype.UUID
	if err := tdb.Pool.QueryRow(ctx, `
		INSERT INTO employees (pegawai_uid, nip, nama, unit_kerja, is_active)
		VALUES ($1, $2, $3, 'CBT', TRUE)
		RETURNING id
	`, "4040603199001", "CBT-NIP-"+suffix, "CBT Teacher "+suffix).Scan(&employeeID); err != nil {
		t.Fatalf("seed employee: %v", err)
	}

	var userID pgtype.UUID
	if err := tdb.Pool.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, display_name, employee_id, is_active, must_change_password, password_changed_at)
		VALUES ($1, 'integration-hash', $2, $3, TRUE, FALSE, NOW())
		RETURNING id
	`, "cbt_"+suffix, "CBT Teacher "+suffix, employeeID).Scan(&userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return employeeID, userID
}

func seedCbtIntegrationClassAssignment(t *testing.T, ctx context.Context, tdb *integrationTestDB, classID, subjectID, employeeID pgtype.UUID) {
	t.Helper()
	if _, err := tdb.Pool.Exec(ctx, `
		INSERT INTO class_subject_assignments (class_id, subject_id, teacher_employee_id)
		VALUES ($1, $2, $3)
	`, classID, subjectID, employeeID); err != nil {
		t.Fatalf("seed class subject assignment: %v", err)
	}
}

func seedCbtIntegrationStudent(t *testing.T, ctx context.Context, q *Queries, classID pgtype.UUID, suffix string) Student {
	t.Helper()
	student, err := q.CreateStudent(ctx, CreateStudentParams{
		Nis:         "CBT-NIS-" + suffix,
		Nisn:        "CBT-NISN-" + suffix,
		Nama:        "CBT Student " + suffix,
		Gender:      GenderEnumL,
		ParentName:  "CBT Parent " + suffix,
		ParentPhone: "081234" + suffix[len(suffix)-6:],
		ClassID:     classID,
		IsActive:    true,
		Status:      StudentStatusEnumActive,
	})
	if err != nil {
		t.Fatalf("create cbt integration student: %v", err)
	}
	return student
}

func createCbtIntegrationQuestion(t *testing.T, ctx context.Context, q *Queries, eventID, subjectID pgtype.UUID, suffix string) CbtQuestion {
	t.Helper()
	question, err := q.CreateCbtQuestion(ctx, CreateCbtQuestionParams{
		EventID:              eventID,
		SubjectID:            subjectID,
		Code:                 "CBT-Q-" + suffix,
		QuestionText:         "What is 2 + 2?",
		QuestionType:         "multiple_choice",
		Options:              []byte(`[{"key":"A","text":"4"},{"key":"B","text":"5"}]`),
		OptionA:              "4",
		OptionB:              "5",
		OptionC:              "6",
		OptionD:              "7",
		OptionE:              "",
		AnswerKey:            "A",
		Explanation:          "Basic arithmetic",
		Difficulty:           CbtQuestionDifficultyEnumEasy,
		Status:               CbtQuestionStatusEnumPublished,
		StemHtml:             "<p>What is 2 + 2?</p>",
		StemLatex:            "",
		StimulusHtml:         "",
		StimulusLatex:        "",
		ExplanationHtml:      "<p>2 + 2 = 4</p>",
		RubricHtml:           "",
		AcademicPhase:        "D",
		TargetLevel:          pgtype.Text{String: "VII", Valid: true},
		CpRef:                "CP-IT",
		TpRef:                "TP-IT",
		KdRef:                "KD-IT",
		IndicatorRef:         "IND-IT",
		MaterialTopic:        "Arithmetic",
		CognitiveLevel:       "C1",
		HotsFlag:             false,
		MediaAssetIds:        []byte(`[]`),
		WorkflowStatus:       "approved",
		Version:              1,
		AuthorUsername:       "cbt_" + suffix,
		ReviewerUsername:     "reviewer_" + suffix,
		ReviewedAt:           pgTimestamptz(time.Now().UTC()),
		ApproverUsername:     "approver_" + suffix,
		ApprovedAt:           pgTimestamptz(time.Now().UTC()),
		WriterNotes:          "integration writer note",
		ReviewNotes:          "integration review note",
		VersionGroupID:       pgtype.UUID{},
		VersionNumber:        1,
		SourceQuestionID:     pgtype.UUID{},
		SupersedesQuestionID: pgtype.UUID{},
		IsLatestVersion:      true,
		VersionNote:          "initial",
	})
	if err != nil {
		t.Fatalf("create cbt question: %v", err)
	}
	return question
}

func pgTimestamptz(ts time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: ts, Valid: true}
}

func cbtQuestionListed(questions []ListCbtQuestionsRow, id pgtype.UUID) bool {
	for _, question := range questions {
		if question.ID == id {
			return true
		}
	}
	return false
}
