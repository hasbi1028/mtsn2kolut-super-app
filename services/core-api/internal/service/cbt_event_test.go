package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func cbtEventTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

type fakeCbtEventStore7A struct {
	listEventsRows []db.ListCbtExamEventsRow
	getEventRow    db.GetCbtExamEventRow
	overviewRow    db.GetCbtEventOverviewSummaryRow
	resultsRows    []db.GetEventResultsRow
	cardsRows      []db.GetEventExamCardsRow
	packagesRows   []db.ListCbtEventPackagesRow
	sessionsRows   []db.ListCbtEventSessionsReadinessRow
	matrixRows     []db.ListCbtEventSubjectMatrixRow

	questionRows     []db.ListCbtEventQuestionCompletenessRowsRow
	contributionRows []db.ListCbtEventQuestionPoolContributionsRow
	excludedLevels   []string
	requirementsRow  db.GetCbtEventQuestionRequirementsRow
	upsertReqRow     db.UpsertCbtEventQuestionRequirementsRow
	lastUpsertReqArg db.UpsertCbtEventQuestionRequirementsParams
	upsertReqCalls   int

	createEventRow  db.CbtExamEvent
	updateStatusRow db.CbtExamEvent
	updateEventRow  db.CbtExamEvent
	deleteEventRows int64

	membersRows         []db.ListCbtEventMembersRow
	membersByUserRows   []db.CbtEventMember
	getMemberRow        db.GetCbtEventMemberRow
	createMemberRow     db.CbtEventMember
	updateMemberRow     db.CbtEventMember
	lastCreateMemberArg db.CreateCbtEventMemberParams
	lastUpdateMemberArg db.UpdateCbtEventMemberParams
	lastDeleteMemberArg db.DeleteCbtEventMemberParams
	createMemberCalls   int
	updateMemberCalls   int
	deleteMemberCalls   int

	targetRows          []db.ListCbtEventSubjectTargetsRow
	upsertTargetRow     db.CbtEventSubjectTarget
	deleteTargetRows    int64
	lastUpsertTargetArg db.UpsertCbtEventSubjectTargetParams
	lastDeleteTargetArg db.DeleteCbtEventSubjectTargetParams
	upsertTargetCalls   int
	deleteTargetCalls   int

	err error
}

func (f *fakeCbtEventStore7A) ListCbtExamEvents(ctx context.Context) ([]db.ListCbtExamEventsRow, error) {
	return f.listEventsRows, f.err
}
func (f *fakeCbtEventStore7A) GetCbtExamEvent(ctx context.Context, id pgtype.UUID) (db.GetCbtExamEventRow, error) {
	return f.getEventRow, f.err
}
func (f *fakeCbtEventStore7A) GetCbtEventOverviewSummary(ctx context.Context, id pgtype.UUID) (db.GetCbtEventOverviewSummaryRow, error) {
	return f.overviewRow, f.err
}
func (f *fakeCbtEventStore7A) GetEventResults(ctx context.Context, eventID pgtype.UUID) ([]db.GetEventResultsRow, error) {
	return f.resultsRows, f.err
}
func (f *fakeCbtEventStore7A) GetEventExamCards(ctx context.Context, eventID pgtype.UUID) ([]db.GetEventExamCardsRow, error) {
	return f.cardsRows, f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventPackagesRow, error) {
	return f.packagesRows, f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventSessionsReadiness(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSessionsReadinessRow, error) {
	return f.sessionsRows, f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventSubjectMatrix(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectMatrixRow, error) {
	return f.matrixRows, f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventQuestionCompletenessRows(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventQuestionCompletenessRowsRow, error) {
	return f.questionRows, f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventQuestionPoolContributions(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventQuestionPoolContributionsRow, error) {
	return f.contributionRows, f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventQuestionCompletenessExcludedLevels(ctx context.Context, id pgtype.UUID) ([]string, error) {
	return f.excludedLevels, f.err
}
func (f *fakeCbtEventStore7A) GetCbtEventQuestionRequirements(ctx context.Context, id pgtype.UUID) (db.GetCbtEventQuestionRequirementsRow, error) {
	return f.requirementsRow, f.err
}
func (f *fakeCbtEventStore7A) UpsertCbtEventQuestionRequirements(ctx context.Context, arg db.UpsertCbtEventQuestionRequirementsParams) (db.UpsertCbtEventQuestionRequirementsRow, error) {
	f.upsertReqCalls++
	f.lastUpsertReqArg = arg
	return f.upsertReqRow, f.err
}
func (f *fakeCbtEventStore7A) CreateCbtExamEvent(ctx context.Context, arg db.CreateCbtExamEventParams) (db.CbtExamEvent, error) {
	return f.createEventRow, f.err
}
func (f *fakeCbtEventStore7A) UpdateCbtExamEventStatus(ctx context.Context, arg db.UpdateCbtExamEventStatusParams) (db.CbtExamEvent, error) {
	return f.updateStatusRow, f.err
}
func (f *fakeCbtEventStore7A) UpdateCbtExamEvent(ctx context.Context, arg db.UpdateCbtExamEventParams) (db.CbtExamEvent, error) {
	return f.updateEventRow, f.err
}
func (f *fakeCbtEventStore7A) DeleteCbtExamEvent(ctx context.Context, id pgtype.UUID) (int64, error) {
	return f.deleteEventRows, f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventMembers(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventMembersRow, error) {
	return f.membersRows, f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventMembersByUser(ctx context.Context, userID pgtype.UUID) ([]db.CbtEventMember, error) {
	return f.membersByUserRows, f.err
}
func (f *fakeCbtEventStore7A) GetCbtEventMember(ctx context.Context, arg db.GetCbtEventMemberParams) (db.GetCbtEventMemberRow, error) {
	return f.getMemberRow, f.err
}
func (f *fakeCbtEventStore7A) CreateCbtEventMember(ctx context.Context, arg db.CreateCbtEventMemberParams) (db.CbtEventMember, error) {
	f.createMemberCalls++
	f.lastCreateMemberArg = arg
	return f.createMemberRow, f.err
}
func (f *fakeCbtEventStore7A) UpdateCbtEventMember(ctx context.Context, arg db.UpdateCbtEventMemberParams) (db.CbtEventMember, error) {
	f.updateMemberCalls++
	f.lastUpdateMemberArg = arg
	return f.updateMemberRow, f.err
}
func (f *fakeCbtEventStore7A) DeleteCbtEventMember(ctx context.Context, arg db.DeleteCbtEventMemberParams) error {
	f.deleteMemberCalls++
	f.lastDeleteMemberArg = arg
	return f.err
}
func (f *fakeCbtEventStore7A) ListCbtEventSubjectTargets(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectTargetsRow, error) {
	return f.targetRows, f.err
}
func (f *fakeCbtEventStore7A) UpsertCbtEventSubjectTarget(ctx context.Context, arg db.UpsertCbtEventSubjectTargetParams) (db.CbtEventSubjectTarget, error) {
	f.upsertTargetCalls++
	f.lastUpsertTargetArg = arg
	return f.upsertTargetRow, f.err
}
func (f *fakeCbtEventStore7A) DeleteCbtEventSubjectTarget(ctx context.Context, arg db.DeleteCbtEventSubjectTargetParams) (int64, error) {
	f.deleteTargetCalls++
	f.lastDeleteTargetArg = arg
	return f.deleteTargetRows, f.err
}

func TestCbtEventQuestionCompletenessSummarizesRows(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(1)
	req := db.GetCbtEventQuestionRequirementsRow{EventID: eventID, ScopeMode: "per_rombel", TargetPg: 20, TargetEssay: 5, StatusFilter: "published_only"}
	store := &fakeCbtEventStore7A{
		requirementsRow: req,
		questionRows: []db.ListCbtEventQuestionCompletenessRowsRow{
			{Level: "VII", ClassName: "VII A", SubjectName: "Math", TargetPg: 10, AvailablePg: 7, TargetEssay: 2, AvailableEssay: 5},
			{Level: "VIII", ClassName: "VIII A", SubjectName: "IPA", TargetPg: 8, AvailablePg: 10, TargetEssay: 1, AvailableEssay: 1},
		},
		excludedLevels: []string{"IX"},
	}
	svc := &CbtEvent{q: store}

	got, err := svc.QuestionCompleteness(ctx, eventID)
	if err != nil {
		t.Fatalf("QuestionCompleteness() error = %v", err)
	}
	if got.Requirements != req {
		t.Fatalf("requirements = %#v, want %#v", got.Requirements, req)
	}
	if len(got.Contributions) != 0 {
		t.Fatalf("contributions len = %d, want 0", len(got.Contributions))
	}
	if !reflect.DeepEqual(got.ExcludedLevels, []string{"IX"}) {
		t.Fatalf("excluded levels = %#v", got.ExcludedLevels)
	}
	if len(got.Rows) != 2 {
		t.Fatalf("rows len = %d, want 2", len(got.Rows))
	}
	if got.Rows[0].MissingPg != 3 || got.Rows[0].MissingEssay != 0 || got.Rows[0].Complete {
		t.Fatalf("first row missing/complete = pg %d essay %d complete %v", got.Rows[0].MissingPg, got.Rows[0].MissingEssay, got.Rows[0].Complete)
	}
	if got.Rows[1].MissingPg != 0 || got.Rows[1].MissingEssay != 0 || !got.Rows[1].Complete {
		t.Fatalf("second row missing/complete = pg %d essay %d complete %v", got.Rows[1].MissingPg, got.Rows[1].MissingEssay, got.Rows[1].Complete)
	}
	wantSummary := CbtQuestionCompletenessSummary{TotalRows: 2, CompleteRows: 1, IncompleteRows: 1, TargetPg: 18, AvailablePg: 17, MissingPg: 3, TargetEssay: 3, AvailableEssay: 6, MissingEssay: 0}
	if got.Summary != wantSummary {
		t.Fatalf("summary = %#v, want %#v", got.Summary, wantSummary)
	}
	if maxInt32(-5, 0) != 0 || maxInt32(6, 0) != 6 {
		t.Fatalf("maxInt32 floor behavior failed")
	}
}

func TestCbtEventQuestionRequirementsValidationAndDefaults(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(2)
	store := &fakeCbtEventStore7A{requirementsRow: db.GetCbtEventQuestionRequirementsRow{EventID: eventID, TargetPg: 20}}
	svc := &CbtEvent{q: store}

	gotReq, err := svc.GetQuestionRequirements(ctx, eventID)
	if err != nil || gotReq.EventID != eventID {
		t.Fatalf("GetQuestionRequirements() = %#v, %v", gotReq, err)
	}
	_, err = svc.UpsertQuestionRequirements(ctx, eventID, SaveCbtEventQuestionRequirementsInput{TargetPg: 12, TargetEssay: 3})
	if err != nil {
		t.Fatalf("UpsertQuestionRequirements(defaults) error = %v", err)
	}
	wantArg := db.UpsertCbtEventQuestionRequirementsParams{EventID: eventID, ScopeMode: "per_rombel", TargetPg: 12, TargetEssay: 3, StatusFilter: "published_only"}
	if store.lastUpsertReqArg != wantArg {
		t.Fatalf("upsert arg = %#v, want %#v", store.lastUpsertReqArg, wantArg)
	}

	badInputs := []SaveCbtEventQuestionRequirementsInput{
		{ScopeMode: "bad", TargetPg: 1},
		{StatusFilter: "bad", TargetPg: 1},
		{TargetPg: -1},
		{TargetEssay: -1},
	}
	for _, in := range badInputs {
		before := store.upsertReqCalls
		if _, err := svc.UpsertQuestionRequirements(ctx, eventID, in); !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("UpsertQuestionRequirements(%#v) error = %v, want ErrBadRequest", in, err)
		}
		if store.upsertReqCalls != before {
			t.Fatalf("invalid input %#v called store", in)
		}
	}
}

func TestCbtEventCanReadChecksMembership(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(3)
	userID := cbtEventTestUUID(4)
	store := &fakeCbtEventStore7A{membersByUserRows: []db.CbtEventMember{{EventID: cbtEventTestUUID(9)}, {EventID: eventID}}}
	svc := &CbtEvent{q: store}

	ok, err := svc.CanRead(ctx, eventID, userID)
	if err != nil || !ok {
		t.Fatalf("CanRead(matching) = %v, %v; want true, nil", ok, err)
	}
	ok, err = svc.CanRead(ctx, cbtEventTestUUID(10), userID)
	if err != nil || ok {
		t.Fatalf("CanRead(non-matching) = %v, %v; want false, nil", ok, err)
	}
	ok, err = svc.CanRead(ctx, pgtype.UUID{}, userID)
	if err != nil || ok {
		t.Fatalf("CanRead(invalid event) = %v, %v; want false, nil", ok, err)
	}

	boom := errors.New("boom")
	svc = &CbtEvent{q: &fakeCbtEventStore7A{err: boom}}
	if _, err := svc.CanRead(ctx, eventID, userID); !errors.Is(err, boom) {
		t.Fatalf("CanRead(store error) = %v, want boom", err)
	}
}

func TestCbtEventSopReadinessBuildsStageStatuses(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(5)
	store := &fakeCbtEventStore7A{
		overviewRow: db.GetCbtEventOverviewSummaryRow{
			ID: eventID, Status: "active", MemberCount: 2,
			TargetQuestionCount: 10, TotalQuestions: 9, PublishedQuestions: 6, ReviewQuestions: 2, ApprovedQuestions: 1,
			PackageCount: 2, ActivePackageCount: 1, EmptyPackageCount: 1,
			SessionCount: 1, ScheduledSessionCount: 1, RoomCount: 1, ParticipantCount: 4,
			UnassignedParticipantCount: 1, MissingSeatCount: 1, RoomsWithoutProctor: 1, TokenReadyCount: 2,
			SubmittedCount: 3, ScoredCount: 1,
		},
		matrixRows: []db.ListCbtEventSubjectMatrixRow{{SubjectName: "Math", TargetQuestions: 10, ShortageCount: 4, AuthorCount: 0, ReviewerCount: 0}},
	}
	svc := &CbtEvent{q: store}

	got, err := svc.SopReadiness(ctx, eventID)
	if err != nil {
		t.Fatalf("SopReadiness() error = %v", err)
	}
	if got.EventID != uuidString(eventID) {
		t.Fatalf("EventID = %q, want %q", got.EventID, uuidString(eventID))
	}
	if len(got.Stages) != 10 {
		t.Fatalf("stages len = %d, want 10", len(got.Stages))
	}
	stages := map[string]CbtSopStageReadiness{}
	for _, stage := range got.Stages {
		stages[stage.Key] = stage
	}
	assertStage := func(key string, status CbtSopStageStatus, blocking, warning int32) {
		t.Helper()
		stage, ok := stages[key]
		if !ok {
			t.Fatalf("missing stage %q", key)
		}
		if stage.Status != status || stage.BlockingCount != blocking || stage.WarningCount != warning {
			t.Fatalf("stage %s = status %s blocking %d warning %d", key, stage.Status, stage.BlockingCount, stage.WarningCount)
		}
		if stage.NextActions == nil {
			t.Fatalf("stage %s NextActions is nil", key)
		}
	}
	assertStage("question_authoring", CbtSopStageWarning, 4, 2)
	assertStage("question_verification", CbtSopStageWarning, 2, 1)
	assertStage("package_ready", CbtSopStageWarning, 1, 1)
	assertStage("participants_rooms_ready", CbtSopStageWarning, 3, 1)
	assertStage("tokens_cards_ready", CbtSopStageWarning, 2, 1)
	assertStage("execution", CbtSopStageWarning, 0, 1)
	assertStage("grading", CbtSopStageWarning, 2, 3)
	assertStage("result_verification", CbtSopStageWarning, 0, 1)
	assertStage("final_archive", CbtSopStageWarning, 0, 0)
}

func TestCbtEventMemberMethodsPassThroughParams(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(6)
	memberID := cbtEventTestUUID(7)
	userID := cbtEventTestUUID(8)
	employeeID := cbtEventTestUUID(9)
	subjectID := cbtEventTestUUID(10)
	store := &fakeCbtEventStore7A{}
	svc := &CbtEvent{q: store}
	input := SaveCbtEventMemberInput{UserID: userID, EmployeeID: employeeID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer}

	if _, err := svc.CreateMember(ctx, eventID, input); err != nil {
		t.Fatalf("CreateMember() error = %v", err)
	}
	if want := (db.CreateCbtEventMemberParams{EventID: eventID, UserID: userID, EmployeeID: employeeID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer}); store.lastCreateMemberArg != want {
		t.Fatalf("create arg = %#v, want %#v", store.lastCreateMemberArg, want)
	}
	if _, err := svc.UpdateMember(ctx, eventID, memberID, input); err != nil {
		t.Fatalf("UpdateMember() error = %v", err)
	}
	if want := (db.UpdateCbtEventMemberParams{ID: memberID, EventID: eventID, UserID: userID, EmployeeID: employeeID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer}); store.lastUpdateMemberArg != want {
		t.Fatalf("update arg = %#v, want %#v", store.lastUpdateMemberArg, want)
	}
	if err := svc.DeleteMember(ctx, eventID, memberID); err != nil {
		t.Fatalf("DeleteMember() error = %v", err)
	}
	if want := (db.DeleteCbtEventMemberParams{ID: memberID, EventID: eventID}); store.lastDeleteMemberArg != want {
		t.Fatalf("delete arg = %#v, want %#v", store.lastDeleteMemberArg, want)
	}
}

func TestCbtEventQuestionTargetsValidationAndDelete(t *testing.T) {
	ctx := context.Background()
	eventID := cbtEventTestUUID(11)
	subjectID := cbtEventTestUUID(12)
	store := &fakeCbtEventStore7A{deleteTargetRows: 1}
	svc := &CbtEvent{q: store}

	if _, err := svc.UpsertQuestionTarget(ctx, eventID, SaveCbtEventSubjectTargetInput{SubjectID: subjectID, TargetQuestions: 25}); err != nil {
		t.Fatalf("UpsertQuestionTarget() error = %v", err)
	}
	if want := (db.UpsertCbtEventSubjectTargetParams{EventID: eventID, SubjectID: subjectID, TargetQuestions: 25}); store.lastUpsertTargetArg != want {
		t.Fatalf("upsert target arg = %#v, want %#v", store.lastUpsertTargetArg, want)
	}
	for _, in := range []SaveCbtEventSubjectTargetInput{{SubjectID: pgtype.UUID{}, TargetQuestions: 1}, {SubjectID: subjectID, TargetQuestions: 0}, {SubjectID: subjectID, TargetQuestions: -1}} {
		before := store.upsertTargetCalls
		if _, err := svc.UpsertQuestionTarget(ctx, eventID, in); !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("UpsertQuestionTarget(%#v) error = %v, want ErrBadRequest", in, err)
		}
		if store.upsertTargetCalls != before {
			t.Fatalf("invalid target input %#v called store", in)
		}
	}
	if _, err := svc.UpsertQuestionTarget(ctx, pgtype.UUID{}, SaveCbtEventSubjectTargetInput{SubjectID: subjectID, TargetQuestions: 1}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("UpsertQuestionTarget(invalid event) error = %v, want ErrBadRequest", err)
	}

	if err := svc.DeleteQuestionTarget(ctx, eventID, subjectID); err != nil {
		t.Fatalf("DeleteQuestionTarget() error = %v", err)
	}
	if want := (db.DeleteCbtEventSubjectTargetParams{EventID: eventID, SubjectID: subjectID}); store.lastDeleteTargetArg != want {
		t.Fatalf("delete target arg = %#v, want %#v", store.lastDeleteTargetArg, want)
	}
	store.deleteTargetRows = 0
	if err := svc.DeleteQuestionTarget(ctx, eventID, subjectID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeleteQuestionTarget(not found) error = %v, want ErrNotFound", err)
	}
	boom := errors.New("boom")
	store.err = boom
	if err := svc.DeleteQuestionTarget(ctx, eventID, subjectID); !errors.Is(err, boom) {
		t.Fatalf("DeleteQuestionTarget(store error) = %v, want boom", err)
	}
}

func TestCbtEventBuildReadinessBlockers(t *testing.T) {
	summary := db.GetCbtEventOverviewSummaryRow{TargetQuestionCount: 5, PublishedQuestions: 3, EmptyPackageCount: 1, UnassignedParticipantCount: 2, MissingSeatCount: 1, RoomsWithoutProctor: 1, ParticipantCount: 3, TokenReadyCount: 1}
	matrix := []db.ListCbtEventSubjectMatrixRow{{SubjectName: "Math", TargetQuestions: 5, ShortageCount: 2}}
	readiness, blockers := buildCbtEventReadiness(summary, matrix)
	if readiness.AuthoringReady || readiness.PackageReady || readiness.RoomReady || readiness.TokenReady || readiness.CardReady || readiness.ResultsReady {
		t.Fatalf("readiness unexpectedly ready: %#v", readiness)
	}
	joined := strings.Join(blockers, "|")
	for _, want := range []string{"event belum memiliki anggota/panitia", "jumlah soal terbit belum memenuhi target event", "mapel Math kurang 2 soal terbit", "belum ada paket aktif yang berisi soal", "belum ada sesi ujian untuk event", "2 peserta belum masuk ruang", "1 peserta belum memiliki nomor kursi", "1 ruang belum memiliki pengawas/proktor", "2 token peserta belum siap"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("blockers %q missing %q", joined, want)
		}
	}
}
