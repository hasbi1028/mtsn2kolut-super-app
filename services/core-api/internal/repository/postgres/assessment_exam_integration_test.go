package db

import (
	"context"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationNonTestAssessmentSubmissionGradeSyncQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	year, class, subject, assignment, teacherID, studentID := createAssessmentExamFixture(t, ctx, tdb, suffix)
	_ = year

	dueAt := pgtype.Timestamptz{Time: time.Date(2026, time.September, 1, 8, 0, 0, 0, time.UTC), Valid: true}
	assessment, err := q.CreateNonTestAssessment(ctx, CreateNonTestAssessmentParams{
		SubjectID:            subject.ID,
		ClassID:              class.ID,
		AssessmentType:       "proyek",
		Title:                "Integration Project " + suffix,
		Description:          "integration non-test assessment",
		InstructionHtml:      "<p>Collect evidence</p>",
		RubricHtml:           "<p>Accuracy and process</p>",
		EvidenceRequirements: "photo and report " + suffix,
		Mode:                 "advance",
		ScoringScale:         "0_100",
		MaxScore:             numericFromInt(100),
		Weight:               numericFromInt(25),
		DueAt:                dueAt,
		Status:               "active",
		CreatedByUsername:    "integration-admin",
		AssessorUsername:     "integration-teacher",
		Checklist:            []byte(`[{"label":"evidence","required":true}]`),
	})
	if err != nil {
		t.Fatalf("create non-test assessment: %v", err)
	}

	generated, err := q.GenerateNonTestSubmissionsForClass(ctx, GenerateNonTestSubmissionsForClassParams{AssessmentID: assessment.ID, ClassID: class.ID})
	if err != nil {
		t.Fatalf("generate non-test submissions: %v", err)
	}
	if len(generated) != 1 || generated[0].StudentID != studentID || generated[0].Status != "assigned" {
		t.Fatalf("generated submissions = %#v, want one assigned submission for student %v", generated, studentID)
	}

	submittedAt := pgtype.Timestamptz{Time: time.Date(2026, time.August, 20, 9, 30, 0, 0, time.UTC), Valid: true}
	gradedAt := pgtype.Timestamptz{Time: time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC), Valid: true}
	reviewed, err := q.UpsertNonTestSubmission(ctx, UpsertNonTestSubmissionParams{
		AssessmentID:     assessment.ID,
		StudentID:        studentID,
		Status:           "reviewed",
		EvidenceUrl:      "https://example.invalid/evidence/" + suffix,
		EvidenceNote:     "complete evidence",
		Score:            numericFromInt(88),
		Feedback:         "solid project",
		SubmittedAt:      submittedAt,
		GradedAt:         gradedAt,
		GradedByUsername: "integration-teacher",
	})
	if err != nil {
		t.Fatalf("upsert non-test submission: %v", err)
	}
	if reviewed.Status != "reviewed" || !numericApprox(reviewed.Score, 88) {
		t.Fatalf("reviewed submission = %#v, want reviewed score 88", reviewed)
	}

	got, err := q.GetNonTestAssessment(ctx, assessment.ID)
	if err != nil {
		t.Fatalf("get non-test assessment: %v", err)
	}
	if got.TotalSubmissions != 1 || got.ReviewedSubmissions != 1 || got.UnsyncedReviewedSubmissions != 1 {
		t.Fatalf("assessment submission stats = total %d reviewed %d unsynced %d, want 1/1/1", got.TotalSubmissions, got.ReviewedSubmissions, got.UnsyncedReviewedSubmissions)
	}

	listed, err := q.ListNonTestAssessments(ctx, ListNonTestAssessmentsParams{
		SubjectID:            subject.ID,
		ClassID:              class.ID,
		StatusFilter:         "active",
		AssessmentTypeFilter: "proyek",
		TeacherEmployeeID:    teacherID,
		SyncFilter:           "needs_sync",
		SearchQuery:          suffix,
		LimitCount:           10,
	})
	if err != nil {
		t.Fatalf("list non-test assessments needing sync: %v", err)
	}
	if !nonTestAssessmentListed(listed, assessment.ID, 1, 1) {
		t.Fatalf("created assessment %v not returned by ListNonTestAssessments needs_sync: %#v", assessment.ID, listed)
	}
	count, err := q.CountNonTestAssessments(ctx, CountNonTestAssessmentsParams{
		SubjectID:            subject.ID,
		ClassID:              class.ID,
		StatusFilter:         "active",
		AssessmentTypeFilter: "proyek",
		TeacherEmployeeID:    teacherID,
		SyncFilter:           "needs_sync",
		SearchQuery:          suffix,
	})
	if err != nil {
		t.Fatalf("count non-test assessments needing sync: %v", err)
	}
	if count != 1 {
		t.Fatalf("count non-test assessments needing sync = %d, want 1", count)
	}

	submissions, err := q.ListNonTestSubmissions(ctx, assessment.ID)
	if err != nil {
		t.Fatalf("list non-test submissions: %v", err)
	}
	if len(submissions) != 1 || submissions[0].StudentID != studentID || submissions[0].Status != "reviewed" || !numericApprox(submissions[0].Score, 88) {
		t.Fatalf("non-test submissions = %#v, want reviewed score 88 for student %v", submissions, studentID)
	}

	ownsAssessment, err := q.TeacherOwnsNonTestAssessment(ctx, TeacherOwnsNonTestAssessmentParams{ID: assessment.ID, TeacherEmployeeID: teacherID})
	if err != nil {
		t.Fatalf("teacher owns non-test assessment: %v", err)
	}
	if !ownsAssessment {
		t.Fatalf("teacher %v should own assessment %v", teacherID, assessment.ID)
	}
	belongs, err := q.StudentBelongsToClass(ctx, StudentBelongsToClassParams{StudentID: studentID, ClassID: class.ID})
	if err != nil {
		t.Fatalf("student belongs to class: %v", err)
	}
	if !belongs {
		t.Fatalf("student %v should belong to class %v", studentID, class.ID)
	}
	gradeAssignment, err := q.GetGradeAssignmentByClassSubject(ctx, GetGradeAssignmentByClassSubjectParams{ClassID: class.ID, SubjectID: subject.ID})
	if err != nil {
		t.Fatalf("get grade assignment by class subject: %v", err)
	}
	if gradeAssignment.ID != assignment.ID || gradeAssignment.TeacherEmployeeID != teacherID {
		t.Fatalf("grade assignment = %#v, want id %v teacher %v", gradeAssignment, assignment.ID, teacherID)
	}

	component, err := q.CreateGradeComponent(ctx, CreateGradeComponentParams{
		AssignmentID: assignment.ID,
		Title:        "Integration Non-Test Sync " + suffix,
		Category:     "project",
		Weight:       25,
		MaxScore:     100,
		IsPublished:  true,
	})
	if err != nil {
		t.Fatalf("create grade component for non-test sync: %v", err)
	}
	if _, err := q.UpsertGradeEntry(ctx, UpsertGradeEntryParams{ComponentID: component.ID, StudentID: studentID, Score: pgtype.Float8{Float64: 88, Valid: true}, Notes: "synced from non-test assessment", GradedBy: "integration-sync"}); err != nil {
		t.Fatalf("upsert grade entry from non-test sync: %v", err)
	}
	marked, err := q.MarkNonTestAssessmentGradeSync(ctx, MarkNonTestAssessmentGradeSyncParams{ID: assessment.ID, GradeComponentID: component.ID, GradeSyncedBy: "integration-sync"})
	if err != nil {
		t.Fatalf("mark non-test assessment grade sync: %v", err)
	}
	if marked.GradeComponentID != component.ID || !marked.GradeSyncedAt.Valid || marked.GradeSyncedBy != "integration-sync" {
		t.Fatalf("marked assessment sync = %#v, want component %v with synced metadata", marked, component.ID)
	}
	componentSource, err := q.GetNonTestGradeComponentSource(ctx, component.ID)
	if err != nil {
		t.Fatalf("get non-test grade component source: %v", err)
	}
	if componentSource != assessment.ID {
		t.Fatalf("non-test grade component source = %v, want assessment %v", componentSource, assessment.ID)
	}
	if _, err := q.GetGradeComponentHighestScore(ctx, component.ID); err != nil {
		t.Fatalf("get grade component highest score: %v", err)
	}
}

func TestIntegrationCbtExamSubmissionScoringAndGradePreflightQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	_, class, subject, assignment, teacherID, studentID := createAssessmentExamFixture(t, ctx, tdb, "CBT-"+suffix)

	question, err := q.CreateCbtQuestion(ctx, CreateCbtQuestionParams{
		SubjectID:        subject.ID,
		Code:             "IT-CBT-Q-" + suffix,
		QuestionText:     "2 + 2 = ? " + suffix,
		QuestionType:     "single_choice",
		Options:          []byte(`["A","B","C","D"]`),
		OptionA:          "3",
		OptionB:          "4",
		OptionC:          "5",
		OptionD:          "6",
		AnswerKey:        "B",
		Explanation:      "basic arithmetic",
		Difficulty:       CbtQuestionDifficultyEnumEasy,
		Status:           CbtQuestionStatusEnumPublished,
		AcademicPhase:    "fase_d",
		TargetLevel:      pgtype.Text{String: "VIII", Valid: true},
		MaterialTopic:    "integration scoring",
		CognitiveLevel:   "C2",
		MediaAssetIds:    []byte(`[]`),
		WorkflowStatus:   "approved",
		Version:          1,
		AuthorUsername:   "integration-author",
		ReviewerUsername: "integration-reviewer",
		ReviewedAt:       pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		ApproverUsername: "integration-approver",
		ApprovedAt:       pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		VersionNumber:    1,
		IsLatestVersion:  true,
	})
	if err != nil {
		t.Fatalf("create cbt question: %v", err)
	}

	pkg, err := q.CreateCbtPackage(ctx, CreateCbtPackageParams{
		SubjectID:          subject.ID,
		Title:              "Integration CBT Package " + suffix,
		Description:        "integration CBT package",
		DurationMinutes:    30,
		RandomizeQuestions: false,
		IsActive:           true,
		SourceMode:         "teacher_class",
		RandomizeOptions:   false,
		RandomSeed:         "seed-" + suffix,
		CompositionLog:     []byte(`{"source":"integration"}`),
	})
	if err != nil {
		t.Fatalf("create cbt package: %v", err)
	}
	if rows, err := q.AddCbtPackageQuestion(ctx, AddCbtPackageQuestionParams{PackageID: pkg.ID, QuestionID: question.ID, Position: 1, Points: 10}); err != nil {
		t.Fatalf("add cbt package question: %v", err)
	} else if rows != 1 {
		t.Fatalf("add cbt package question rows = %d, want 1", rows)
	}

	session, err := q.CreateCbtExamSession(ctx, CreateCbtExamSessionParams{
		PackageID:       pkg.ID,
		ClassID:         class.ID,
		ScopeType:       "class",
		ScopeRef:        "integration",
		MixPolicy:       "same_class",
		AssignmentMode:  "manual",
		AllowCrossGrade: false,
		IsSpecialEvent:  false,
		Title:           "Integration CBT Session " + suffix,
		ScheduledStart:  pgtype.Timestamptz{Time: time.Now().UTC().Add(-time.Minute), Valid: true},
		ScheduledEnd:    pgtype.Timestamptz{Time: time.Now().UTC().Add(time.Hour), Valid: true},
		Status:          CbtSessionStatusEnumActive,
	})
	if err != nil {
		t.Fatalf("create cbt exam session: %v", err)
	}
	if err := q.EnrollClassToSession(ctx, EnrollClassToSessionParams{SessionID: session.ID, ClassID: class.ID}); err != nil {
		t.Fatalf("enroll class to session: %v", err)
	}
	participants, err := q.ListCbtExamParticipants(ctx, session.ID)
	if err != nil {
		t.Fatalf("list cbt exam participants: %v", err)
	}
	if len(participants) != 1 || participants[0].StudentID != studentID {
		t.Fatalf("participants = %#v, want one participant for student %v", participants, studentID)
	}
	participantID := participants[0].ID

	if hasParticipant, err := q.HasSessionParticipant(ctx, HasSessionParticipantParams{SessionID: session.ID, ID: participantID}); err != nil {
		t.Fatalf("has session participant: %v", err)
	} else if !hasParticipant {
		t.Fatalf("session %v should have participant %v", session.ID, participantID)
	}
	if belongs, err := q.QuestionBelongsToParticipantPackage(ctx, QuestionBelongsToParticipantPackageParams{ID: participantID, QuestionID: question.ID}); err != nil {
		t.Fatalf("question belongs to participant package: %v", err)
	} else if !belongs {
		t.Fatalf("question %v should belong to participant %v package", question.ID, participantID)
	}

	if rows, err := q.UpsertStudentAnswer(ctx, UpsertStudentAnswerParams{ParticipantID: participantID, QuestionID: question.ID, Answer: "B"}); err != nil {
		t.Fatalf("upsert student answer: %v", err)
	} else if rows != 1 {
		t.Fatalf("upsert student answer rows = %d, want 1", rows)
	}
	if err := q.UpdateAnswerCorrectness(ctx, session.ID); err != nil {
		t.Fatalf("update answer correctness: %v", err)
	}
	answers, err := q.GetParticipantAnswers(ctx, participantID)
	if err != nil {
		t.Fatalf("get participant answers: %v", err)
	}
	if len(answers) != 1 || answers[0].QuestionID != question.ID || !answers[0].IsCorrect.Bool {
		t.Fatalf("participant answers = %#v, want one correct answer for question %v", answers, question.ID)
	}

	submitted, err := q.SubmitParticipantExam(ctx, participantID)
	if err != nil {
		t.Fatalf("submit participant exam: %v", err)
	}
	if submitted.ID != participantID || !submitted.SubmittedAt.Valid || !numericApprox(submitted.Score, 100) {
		t.Fatalf("submitted participant = %#v, want score 100", submitted)
	}
	results, err := q.GetSessionResults(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session results: %v", err)
	}
	if len(results) != 1 || results[0].ParticipantID != participantID || results[0].TotalAnswers != 1 || results[0].CorrectAnswers != 1 || !numericApprox(results[0].Score, 100) {
		t.Fatalf("session results = %#v, want one 100 score result", results)
	}
	teacherResults, err := q.GetSessionResultsByTeacher(ctx, GetSessionResultsByTeacherParams{SessionID: session.ID, TeacherEmployeeID: teacherID})
	if err != nil {
		t.Fatalf("get session results by teacher: %v", err)
	}
	if len(teacherResults) != 1 || teacherResults[0].ParticipantID != participantID || teacherResults[0].CorrectAnswers != 1 {
		t.Fatalf("teacher session results = %#v, want participant %v", teacherResults, participantID)
	}

	component, err := q.CreateGradeComponent(ctx, CreateGradeComponentParams{AssignmentID: assignment.ID, Title: "Integration CBT Result " + suffix, Category: "quiz", Weight: 30, MaxScore: 100, IsPublished: true})
	if err != nil {
		t.Fatalf("create cbt grade component: %v", err)
	}
	if _, err := q.UpsertGradeEntry(ctx, UpsertGradeEntryParams{ComponentID: component.ID, StudentID: studentID, Score: pgtype.Float8{Float64: 100, Valid: true}, Notes: "synced from cbt session", GradedBy: "integration-cbt-sync"}); err != nil {
		t.Fatalf("upsert cbt grade entry: %v", err)
	}
	if _, err := tdb.Pool.Exec(ctx, `
		INSERT INTO cbt_result_sync_runs (session_id, grade_assignment_id, grade_component_id, status, candidate_count, synced_count, skipped_count, threshold, notes, created_by)
		VALUES ($1, $2, $3, 'synced', 1, 1, 0, 75, 'integration sync run', 'integration-cbt-sync')
	`, session.ID, assignment.ID, component.ID); err != nil {
		t.Fatalf("insert cbt result sync run fixture: %v", err)
	}
	preflight, err := q.GetCbtSessionGradeSyncPreflight(ctx, session.ID)
	if err != nil {
		t.Fatalf("get cbt session grade sync preflight: %v", err)
	}
	if preflight.SessionID != session.ID || preflight.GradeAssignmentID != assignment.ID || preflight.GradeComponentID != component.ID || preflight.LatestSyncStatus != "synced" || preflight.ParticipantCount != 1 || preflight.SubmittedCount != 1 || preflight.ScoredCount != 1 || preflight.MissingScoreCount != 0 {
		t.Fatalf("cbt grade sync preflight = %#v, want synced one-participant summary", preflight)
	}
}

func createAssessmentExamFixture(t *testing.T, ctx context.Context, tdb *integrationTestDB, suffix string) (AcademicYear, SchoolClass, Subject, CreateRombelSubjectAssignmentRow, pgtype.UUID, pgtype.UUID) {
	t.Helper()
	q := tdb.Q
	year, err := q.CreateAcademicYear(ctx, CreateAcademicYearParams{
		Name:      "Integration Assessment Exam Year " + suffix,
		StartDate: pgDate(2026, time.July, 1),
		EndDate:   pgDate(2027, time.June, 30),
		IsActive:  false,
	})
	if err != nil {
		t.Fatalf("create academic year: %v", err)
	}
	class, err := q.CreateSchoolClass(ctx, CreateSchoolClassParams{
		AcademicYearID: year.ID,
		Code:           "IT-AE-CLS-" + suffix,
		Name:           "Integration Assessment Exam Class " + suffix,
		Level:          "8",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create school class: %v", err)
	}
	subject, err := q.CreateSubject(ctx, CreateSubjectParams{
		Code:                "IT-AE-SUB-" + suffix,
		Name:                "Integration Assessment Exam Subject " + suffix,
		Category:            "umum",
		IsAssessmentSubject: true,
		IsReportSubject:     true,
		IsScheduleActivity:  false,
		CountsForRanking:    true,
		IsLocalContent:      false,
		IsChoiceSubject:     false,
		DefaultWeeklyHours:  2,
		DisplayOrder:        9996,
		IsActive:            true,
	})
	if err != nil {
		t.Fatalf("create subject: %v", err)
	}
	teacherID := insertIntegrationEmployee(t, ctx, tdb, "IT-AE-T-"+suffix, "Integration Assessment Exam Teacher "+suffix, employeeUID('5', suffix))
	studentID := insertIntegrationStudent(t, ctx, tdb, "IT-AE-NIS-"+suffix, "Integration Assessment Exam Student "+suffix, class.ID)
	assignment, err := q.CreateRombelSubjectAssignment(ctx, CreateRombelSubjectAssignmentParams{ClassID: class.ID, SubjectID: subject.ID, TeacherEmployeeID: teacherID})
	if err != nil {
		t.Fatalf("create rombel subject assignment: %v", err)
	}
	return year, class, subject, assignment, teacherID, studentID
}

func numericFromInt(value int64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(value), Exp: 0, Valid: true}
}

func numericApprox(n pgtype.Numeric, want float64) bool {
	got, ok := numericToFloat64(n)
	return ok && math.Abs(got-want) < 0.0001
}

func numericToFloat64(n pgtype.Numeric) (float64, bool) {
	if !n.Valid || n.Int == nil {
		return 0, false
	}
	rat := new(big.Rat).SetInt(n.Int)
	if n.Exp > 0 {
		rat.Mul(rat, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)))
	} else if n.Exp < 0 {
		rat.Quo(rat, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)))
	}
	value, _ := rat.Float64()
	return value, true
}

func nonTestAssessmentListed(items []ListNonTestAssessmentsRow, id pgtype.UUID, totalSubmissions, reviewedSubmissions int32) bool {
	for _, item := range items {
		if item.ID == id && item.TotalSubmissions == totalSubmissions && item.ReviewedSubmissions == reviewedSubmissions {
			return true
		}
	}
	return false
}
