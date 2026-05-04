package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeAttendanceStore struct {
	listRows      []db.ListAttendanceRow
	listErr       error
	listArg       db.ListAttendanceParams
	count         int64
	countErr      error
	byDateArg     pgtype.Date
	rangeArg      db.ListAttendanceInRangeParams
	byEmployeeArg db.ListAttendanceByEmployeeParams
	summaryArg    db.GetMonthlyAttendanceSummaryParams
	upsertArg     db.UpsertAttendanceParams
}

func (f *fakeAttendanceStore) ListAttendance(ctx context.Context, arg db.ListAttendanceParams) ([]db.ListAttendanceRow, error) {
	f.listArg = arg
	return f.listRows, f.listErr
}

func (f *fakeAttendanceStore) CountAttendance(ctx context.Context) (int64, error) {
	return f.count, f.countErr
}

func (f *fakeAttendanceStore) ListAttendanceByDate(ctx context.Context, tanggal pgtype.Date) ([]db.ListAttendanceByDateRow, error) {
	f.byDateArg = tanggal
	return []db.ListAttendanceByDateRow{{EmployeeNama: "Guru"}}, nil
}

func (f *fakeAttendanceStore) ListAttendanceInRange(ctx context.Context, arg db.ListAttendanceInRangeParams) ([]db.ListAttendanceInRangeRow, error) {
	f.rangeArg = arg
	return []db.ListAttendanceInRangeRow{{EmployeeNama: "Guru"}}, nil
}

func (f *fakeAttendanceStore) ListAttendanceByEmployee(ctx context.Context, arg db.ListAttendanceByEmployeeParams) ([]db.ListAttendanceByEmployeeRow, error) {
	f.byEmployeeArg = arg
	return []db.ListAttendanceByEmployeeRow{{EmployeeID: arg.EmployeeID, EmployeeNama: "Guru"}}, nil
}

func (f *fakeAttendanceStore) GetMonthlyAttendanceSummary(ctx context.Context, arg db.GetMonthlyAttendanceSummaryParams) ([]db.GetMonthlyAttendanceSummaryRow, error) {
	f.summaryArg = arg
	return []db.GetMonthlyAttendanceSummaryRow{{EmployeeNama: "Guru", TotalDays: 2}}, nil
}

func (f *fakeAttendanceStore) UpsertAttendance(ctx context.Context, arg db.UpsertAttendanceParams) (db.AttendanceRecord, error) {
	f.upsertArg = arg
	return db.AttendanceRecord{EmployeeID: arg.EmployeeID, Tanggal: arg.Tanggal, JamMasuk: arg.JamMasuk}, nil
}

func TestPusakaAttendanceServiceForwardsStoreCalls(t *testing.T) {
	employeeID := documentCycleTestUUID(181)
	start := pgtype.Date{Valid: true}
	end := pgtype.Date{Valid: true}
	store := &fakeAttendanceStore{
		listRows: []db.ListAttendanceRow{{EmployeeID: employeeID, EmployeeNama: "Guru"}},
		count:    3,
	}
	svc := &PusakaAttendance{q: store}
	if NewPusakaAttendance(nil) == nil {
		t.Fatal("NewPusakaAttendance(nil) = nil, want service")
	}

	rows, count, err := svc.List(context.Background(), 20, 5)
	if err != nil || len(rows) != 1 || count != 3 {
		t.Fatalf("List() = %d rows/%d/%v, want 1/3/nil", len(rows), count, err)
	}
	if store.listArg.Limit != 20 || store.listArg.Offset != 5 {
		t.Fatalf("List() arg = %+v, want limit/offset", store.listArg)
	}
	if rows, err := svc.ByDate(context.Background(), start); err != nil || len(rows) != 1 || store.byDateArg != start {
		t.Fatalf("ByDate() = %d rows/%v date=%v, want 1/nil/%v", len(rows), err, store.byDateArg, start)
	}
	if rows, err := svc.ListInRange(context.Background(), start, end); err != nil || len(rows) != 1 || store.rangeArg.Tanggal != start || store.rangeArg.Tanggal_2 != end {
		t.Fatalf("ListInRange() = %d rows/%v arg=%+v, want range", len(rows), err, store.rangeArg)
	}
	if rows, err := svc.ByEmployee(context.Background(), employeeID, 10, 2); err != nil || len(rows) != 1 || store.byEmployeeArg.EmployeeID != employeeID || store.byEmployeeArg.Limit != 10 {
		t.Fatalf("ByEmployee() = %d rows/%v arg=%+v, want employee paging", len(rows), err, store.byEmployeeArg)
	}
	if rows, err := svc.GetSummary(context.Background(), start, end); err != nil || len(rows) != 1 || store.summaryArg.Tanggal != start || store.summaryArg.Tanggal_2 != end {
		t.Fatalf("GetSummary() = %d rows/%v arg=%+v, want range", len(rows), err, store.summaryArg)
	}
	if _, err := svc.Upsert(context.Background(), db.UpsertAttendanceParams{EmployeeID: employeeID, Tanggal: start, JamMasuk: "07:00"}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if store.upsertArg.EmployeeID != employeeID || store.upsertArg.JamMasuk != "07:00" {
		t.Fatalf("Upsert() arg = %+v, want forwarded attendance", store.upsertArg)
	}
}

func TestPusakaAttendanceListPropagatesErrors(t *testing.T) {
	svc := &PusakaAttendance{q: &fakeAttendanceStore{listErr: errors.New("list failed")}}
	if _, _, err := svc.List(context.Background(), 1, 0); err == nil || err.Error() != "list failed" {
		t.Fatalf("List(list error) = %v, want list failed", err)
	}
	svc = &PusakaAttendance{q: &fakeAttendanceStore{countErr: errors.New("count failed")}}
	if _, _, err := svc.List(context.Background(), 1, 0); err == nil || err.Error() != "count failed" {
		t.Fatalf("List(count error) = %v, want count failed", err)
	}
}

type fakeCbtEventStore struct {
	listRows        []db.ListCbtExamEventsRow
	listErr         error
	getID           pgtype.UUID
	overviewRow     db.GetCbtEventOverviewSummaryRow
	overviewID      pgtype.UUID
	resultsRows     []db.GetEventResultsRow
	resultsErr      error
	resultsID       pgtype.UUID
	cardsRows       []db.GetEventExamCardsRow
	cardsErr        error
	cardsID         pgtype.UUID
	createArg       db.CreateCbtExamEventParams
	updateStatusArg db.UpdateCbtExamEventStatusParams
	updateArg       db.UpdateCbtExamEventParams
	deleteID        pgtype.UUID
	packagesRows    []db.ListCbtEventPackagesRow
	sessionsRows    []db.ListCbtEventSessionsReadinessRow
	matrixRows      []db.ListCbtEventSubjectMatrixRow
	membersByUser   []db.CbtEventMember
}

func (f *fakeCbtEventStore) ListCbtExamEvents(ctx context.Context) ([]db.ListCbtExamEventsRow, error) {
	return f.listRows, f.listErr
}

func (f *fakeCbtEventStore) GetCbtExamEvent(ctx context.Context, id pgtype.UUID) (db.GetCbtExamEventRow, error) {
	f.getID = id
	return db.GetCbtExamEventRow{ID: id, Title: "PAT"}, nil
}

func (f *fakeCbtEventStore) GetCbtEventOverviewSummary(ctx context.Context, id pgtype.UUID) (db.GetCbtEventOverviewSummaryRow, error) {
	f.overviewID = id
	if f.overviewRow.ID.Valid {
		return f.overviewRow, nil
	}
	return db.GetCbtEventOverviewSummaryRow{ID: id, Title: "PAT"}, nil
}

func (f *fakeCbtEventStore) GetEventResults(ctx context.Context, eventID pgtype.UUID) ([]db.GetEventResultsRow, error) {
	f.resultsID = eventID
	return f.resultsRows, f.resultsErr
}

func (f *fakeCbtEventStore) GetEventExamCards(ctx context.Context, eventID pgtype.UUID) ([]db.GetEventExamCardsRow, error) {
	f.cardsID = eventID
	return f.cardsRows, f.cardsErr
}

func (f *fakeCbtEventStore) CreateCbtExamEvent(ctx context.Context, arg db.CreateCbtExamEventParams) (db.CbtExamEvent, error) {
	f.createArg = arg
	return db.CbtExamEvent{Title: arg.Title, ExamType: arg.ExamType, Scope: arg.Scope, Status: arg.Status, TargetLevels: arg.TargetLevels}, nil
}

func (f *fakeCbtEventStore) UpdateCbtExamEventStatus(ctx context.Context, arg db.UpdateCbtExamEventStatusParams) (db.CbtExamEvent, error) {
	f.updateStatusArg = arg
	return db.CbtExamEvent{ID: arg.ID, Status: arg.Status}, nil
}

func (f *fakeCbtEventStore) UpdateCbtExamEvent(ctx context.Context, arg db.UpdateCbtExamEventParams) (db.CbtExamEvent, error) {
	f.updateArg = arg
	return db.CbtExamEvent{ID: arg.ID, Title: arg.Title, ExamType: arg.ExamType, Scope: arg.Scope, TargetLevels: arg.TargetLevels}, nil
}

func (f *fakeCbtEventStore) DeleteCbtExamEvent(ctx context.Context, id pgtype.UUID) (int64, error) {
	f.deleteID = id
	return 1, nil
}

func (f *fakeCbtEventStore) ListCbtEventMembers(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventMembersRow, error) {
	return nil, nil
}

func (f *fakeCbtEventStore) ListCbtEventMembersByUser(ctx context.Context, userID pgtype.UUID) ([]db.CbtEventMember, error) {
	return f.membersByUser, nil
}

func (f *fakeCbtEventStore) ListCbtEventPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventPackagesRow, error) {
	return f.packagesRows, nil
}

func (f *fakeCbtEventStore) ListCbtEventSessionsReadiness(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSessionsReadinessRow, error) {
	return f.sessionsRows, nil
}

func (f *fakeCbtEventStore) ListCbtEventSubjectMatrix(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectMatrixRow, error) {
	return f.matrixRows, nil
}

func (f *fakeCbtEventStore) GetCbtEventMember(ctx context.Context, arg db.GetCbtEventMemberParams) (db.GetCbtEventMemberRow, error) {
	return db.GetCbtEventMemberRow{}, nil
}

func (f *fakeCbtEventStore) CreateCbtEventMember(ctx context.Context, arg db.CreateCbtEventMemberParams) (db.CbtEventMember, error) {
	return db.CbtEventMember{EventID: arg.EventID, UserID: arg.UserID, SubjectID: arg.SubjectID, Role: arg.Role}, nil
}

func (f *fakeCbtEventStore) UpdateCbtEventMember(ctx context.Context, arg db.UpdateCbtEventMemberParams) (db.CbtEventMember, error) {
	return db.CbtEventMember{ID: arg.ID, EventID: arg.EventID, UserID: arg.UserID, SubjectID: arg.SubjectID, Role: arg.Role}, nil
}

func (f *fakeCbtEventStore) DeleteCbtEventMember(ctx context.Context, arg db.DeleteCbtEventMemberParams) error {
	return nil
}

func (f *fakeCbtEventStore) ListCbtEventSubjectTargets(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectTargetsRow, error) {
	return []db.ListCbtEventSubjectTargetsRow{}, nil
}

func (f *fakeCbtEventStore) UpsertCbtEventSubjectTarget(ctx context.Context, arg db.UpsertCbtEventSubjectTargetParams) (db.CbtEventSubjectTarget, error) {
	return db.CbtEventSubjectTarget{EventID: arg.EventID, SubjectID: arg.SubjectID, TargetQuestions: arg.TargetQuestions}, nil
}

func (f *fakeCbtEventStore) DeleteCbtEventSubjectTarget(ctx context.Context, arg db.DeleteCbtEventSubjectTargetParams) (int64, error) {
	return 1, nil
}

func TestCbtEventServiceForwardsStoreCalls(t *testing.T) {
	eventID := documentCycleTestUUID(191)
	yearID := documentCycleTestUUID(192)
	store := &fakeCbtEventStore{
		listRows:    []db.ListCbtExamEventsRow{{ID: eventID, Title: "PAT"}},
		resultsRows: []db.GetEventResultsRow{{SessionTitle: "Sesi 1"}},
		cardsRows:   []db.GetEventExamCardsRow{{SessionTitle: "Sesi 1"}},
	}
	svc := &CbtEvent{q: store}
	if NewCbtEvent(nil) == nil {
		t.Fatal("NewCbtEvent(nil) = nil, want service")
	}

	if rows, err := svc.List(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("List() = %d rows/%v, want 1 nil", len(rows), err)
	}
	store.listRows = nil
	if rows, err := svc.List(context.Background()); err != nil || len(rows) != 0 {
		t.Fatalf("List(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	otherEventID := documentCycleTestUUID(193)
	userID := documentCycleTestUUID(194)
	store.listRows = []db.ListCbtExamEventsRow{{ID: eventID, Title: "PAT"}, {ID: otherEventID, Title: "Tryout"}}
	store.membersByUser = []db.CbtEventMember{{EventID: eventID, UserID: userID}}
	if rows, err := svc.ListForUser(context.Background(), userID); err != nil || len(rows) != 1 || rows[0].ID != eventID {
		t.Fatalf("ListForUser() = %+v/%v, want only member event %v", rows, err, eventID)
	}
	if got, err := svc.Get(context.Background(), eventID); err != nil || got.ID != eventID || store.getID != eventID {
		t.Fatalf("Get() = %+v/%v id=%v, want event", got, err, store.getID)
	}
	if rows, err := svc.GetResults(context.Background(), eventID); err != nil || len(rows) != 1 || store.resultsID != eventID {
		t.Fatalf("GetResults() = %d rows/%v id=%v, want results", len(rows), err, store.resultsID)
	}
	store.resultsRows = nil
	if rows, err := svc.GetResults(context.Background(), eventID); err != nil || len(rows) != 0 {
		t.Fatalf("GetResults(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if rows, err := svc.GetExamCards(context.Background(), eventID); err != nil || len(rows) != 1 || store.cardsID != eventID {
		t.Fatalf("GetExamCards() = %d rows/%v id=%v, want cards", len(rows), err, store.cardsID)
	}
	store.cardsRows = nil
	if rows, err := svc.GetExamCards(context.Background(), eventID); err != nil || len(rows) != 0 {
		t.Fatalf("GetExamCards(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if _, err := svc.Create(context.Background(), CreateCbtEventInput{Title: "PAT", ExamType: db.CbtExamTypeUas, Scope: "grade", TargetLevels: []string{"9"}, AcademicYearID: yearID}); err != nil {
		t.Fatalf("Create(default status) error = %v", err)
	}
	if store.createArg.Status != "draft" || store.createArg.AcademicYearID != yearID {
		t.Fatalf("Create(default status) arg = %+v, want draft/year", store.createArg)
	}
	if _, err := svc.Create(context.Background(), CreateCbtEventInput{Title: "PAT", ExamType: db.CbtExamTypeUts, Scope: "school", Status: "active"}); err != nil {
		t.Fatalf("Create(status) error = %v", err)
	}
	if store.createArg.Status != "active" {
		t.Fatalf("Create(status) arg = %+v, want active", store.createArg)
	}
	if _, err := svc.Create(context.Background(), CreateCbtEventInput{Title: "PAT", ExamType: db.CbtExamTypeUts, Scope: "school", Status: "published"}); err == nil {
		t.Fatal("Create(invalid status) error = nil, want validation error")
	}
	if _, err := svc.UpdateStatus(context.Background(), eventID, "active"); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if store.updateStatusArg.ID != eventID || store.updateStatusArg.Status != "active" {
		t.Fatalf("UpdateStatus() arg = %+v, want status", store.updateStatusArg)
	}
	if _, err := svc.UpdateStatus(context.Background(), eventID, "archived"); err == nil {
		t.Fatal("UpdateStatus(invalid) error = nil, want validation error")
	}
	if _, err := svc.Update(context.Background(), eventID, CreateCbtEventInput{Title: "PAS", ExamType: db.CbtExamTypeUas, Scope: "school", AcademicYearID: yearID}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if store.updateArg.ID != eventID || store.updateArg.Title != "PAS" {
		t.Fatalf("Update() arg = %+v, want update", store.updateArg)
	}
	if err := svc.Delete(context.Background(), eventID); err != nil || store.deleteID != eventID {
		t.Fatalf("Delete() = %v id=%v, want nil/%v", err, store.deleteID, eventID)
	}
}

func TestCbtEventOverviewBuildsReadinessAndBlockers(t *testing.T) {
	eventID := documentCycleTestUUID(194)
	store := &fakeCbtEventStore{
		overviewRow: db.GetCbtEventOverviewSummaryRow{
			ID:                         eventID,
			Title:                      "PAT",
			MemberCount:                1,
			TargetQuestionCount:        10,
			PublishedQuestions:         8,
			ActivePackageCount:         1,
			SessionCount:               1,
			RoomCount:                  1,
			ParticipantCount:           2,
			AssignedParticipantCount:   1,
			UnassignedParticipantCount: 1,
			TokenReadyCount:            1,
		},
		matrixRows: []db.ListCbtEventSubjectMatrixRow{{SubjectName: "Matematika", TargetQuestions: 10, PublishedQuestions: 8, ShortageCount: 2}},
	}
	svc := &CbtEvent{q: store}

	got, err := svc.Overview(context.Background(), eventID)
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if got.Event.ID != eventID || got.Readiness.AuthoringReady || !got.Readiness.PackageReady || !got.Readiness.SessionReady {
		t.Fatalf("Overview() = %+v, want event with package/session ready and authoring blocked", got)
	}
	if len(got.BlockingReasons) == 0 {
		t.Fatal("Overview() blockers empty, want readiness blockers")
	}
}

func TestCbtEventListPropagatesErrors(t *testing.T) {
	eventID := documentCycleTestUUID(193)
	svc := &CbtEvent{q: &fakeCbtEventStore{listErr: errors.New("list failed")}}
	if _, err := svc.List(context.Background()); err == nil || err.Error() != "list failed" {
		t.Fatalf("List(error) = %v, want list failed", err)
	}
	svc = &CbtEvent{q: &fakeCbtEventStore{resultsErr: errors.New("results failed")}}
	if _, err := svc.GetResults(context.Background(), eventID); err == nil || err.Error() != "results failed" {
		t.Fatalf("GetResults(error) = %v, want results failed", err)
	}
	svc = &CbtEvent{q: &fakeCbtEventStore{cardsErr: errors.New("cards failed")}}
	if _, err := svc.GetExamCards(context.Background(), eventID); err == nil || err.Error() != "cards failed" {
		t.Fatalf("GetExamCards(error) = %v, want cards failed", err)
	}
}

type fakeCbtPackageStore struct {
	packagesRows  []db.ListCbtPackagesRow
	packagesErr   error
	questionsRows []db.ListCbtPackageQuestionsRow
	questionsErr  error
	deleteID      pgtype.UUID
	createArg     db.CreateCbtPackageParams
	createErr     error
	packageRow    db.CbtPackage
	questionRow   db.GetCbtQuestionRow
	questionErr   error
	addArgs       []db.AddCbtPackageQuestionParams
	addErr        error
}

func (f *fakeCbtPackageStore) ListCbtPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, error) {
	return f.packagesRows, f.packagesErr
}

func (f *fakeCbtPackageStore) ListCbtPackageQuestions(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackageQuestionsRow, error) {
	return f.questionsRows, f.questionsErr
}

func (f *fakeCbtPackageStore) DeleteCbtPackage(ctx context.Context, id pgtype.UUID) (int64, error) {
	f.deleteID = id
	return 1, nil
}

func (f *fakeCbtPackageStore) CreateCbtPackage(ctx context.Context, arg db.CreateCbtPackageParams) (db.CbtPackage, error) {
	f.createArg = arg
	if f.packageRow.ID.Valid {
		return f.packageRow, f.createErr
	}
	return db.CbtPackage{ID: documentCycleTestUUID(203), EventID: arg.EventID, SubjectID: arg.SubjectID, Title: arg.Title}, f.createErr
}

func (f *fakeCbtPackageStore) GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	if f.questionRow.ID.Valid {
		f.questionRow.ID = id
		return f.questionRow, f.questionErr
	}
	return db.GetCbtQuestionRow{ID: id, EventID: f.createArg.EventID, SubjectID: f.createArg.SubjectID, Status: db.CbtQuestionStatusEnumPublished}, f.questionErr
}

func (f *fakeCbtPackageStore) AddCbtPackageQuestion(ctx context.Context, arg db.AddCbtPackageQuestionParams) error {
	f.addArgs = append(f.addArgs, arg)
	return f.addErr
}

func (f *fakeCbtPackageStore) WithTx(tx pgx.Tx) *db.Queries {
	return nil
}

func TestCbtPackageServiceListDeleteAndValidation(t *testing.T) {
	packageID := documentCycleTestUUID(201)
	questionID := documentCycleTestUUID(202)
	store := &fakeCbtPackageStore{
		packagesRows:  []db.ListCbtPackagesRow{{ID: packageID, Title: "Paket"}},
		questionsRows: []db.ListCbtPackageQuestionsRow{{PackageID: packageID, QuestionID: questionID}},
	}
	svc := &CbtPackage{q: store}
	if NewCbtPackage(nil) == nil {
		t.Fatal("NewCbtPackage(nil) = nil, want service")
	}

	packages, questions, err := svc.List(context.Background(), pgtype.UUID{})
	if err != nil || len(packages) != 1 || len(questions) != 1 {
		t.Fatalf("List() = %d packages/%d questions/%v, want 1/1/nil", len(packages), len(questions), err)
	}
	if err := svc.Delete(context.Background(), packageID); err != nil || store.deleteID != packageID {
		t.Fatalf("Delete() = %v id=%v, want nil/%v", err, store.deleteID, packageID)
	}
	if _, err := svc.Create(context.Background(), CreateCbtPackageInput{Title: "Kosong"}); err == nil || err.Error() != "question_ids wajib diisi" {
		t.Fatalf("Create(no questions) = %v, want validation error", err)
	}
}

func TestCbtPackageListPropagatesErrors(t *testing.T) {
	svc := &CbtPackage{q: &fakeCbtPackageStore{packagesErr: errors.New("packages failed")}}
	if _, _, err := svc.List(context.Background(), pgtype.UUID{}); err == nil || err.Error() != "packages failed" {
		t.Fatalf("List(package error) = %v, want packages failed", err)
	}
	svc = &CbtPackage{q: &fakeCbtPackageStore{questionsErr: errors.New("questions failed")}}
	if _, _, err := svc.List(context.Background(), pgtype.UUID{}); err == nil || err.Error() != "questions failed" {
		t.Fatalf("List(question error) = %v, want questions failed", err)
	}
}

func TestCreateCbtPackageHelperValidatesQuestionsAndPositions(t *testing.T) {
	eventID := documentCycleTestUUID(211)
	subjectID := documentCycleTestUUID(204)
	packageID := documentCycleTestUUID(205)
	firstQuestionID := documentCycleTestUUID(206)
	secondQuestionID := documentCycleTestUUID(207)
	store := &fakeCbtPackageStore{
		packageRow:  db.CbtPackage{ID: packageID, EventID: eventID, SubjectID: subjectID, Title: "Paket"},
		questionRow: db.GetCbtQuestionRow{ID: firstQuestionID, EventID: eventID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
	}

	got, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
		EventID:            eventID,
		SubjectID:          subjectID,
		Title:              "Paket PAT",
		Description:        "Soal akhir tahun",
		DurationMinutes:    90,
		RandomizeQuestions: true,
		IsActive:           true,
		QuestionIDs:        []pgtype.UUID{firstQuestionID, secondQuestionID},
		QuestionWeights: map[string]int32{
			pgUUIDString(firstQuestionID):  2,
			pgUUIDString(secondQuestionID): 5,
		},
	})
	if err != nil {
		t.Fatalf("createCbtPackage() error = %v", err)
	}
	if got.ID != packageID || store.createArg.EventID != eventID || store.createArg.SubjectID != subjectID || store.createArg.Title != "Paket PAT" || !store.createArg.RandomizeQuestions {
		t.Fatalf("createCbtPackage() got/createArg = %+v/%+v, want created package", got, store.createArg)
	}
	if len(store.addArgs) != 2 {
		t.Fatalf("AddCbtPackageQuestion calls = %d, want 2", len(store.addArgs))
	}
	if store.addArgs[0].QuestionID != firstQuestionID || store.addArgs[0].Position != 1 || store.addArgs[0].Points != 2 ||
		store.addArgs[1].QuestionID != secondQuestionID || store.addArgs[1].Position != 2 || store.addArgs[1].Points != 5 {
		t.Fatalf("AddCbtPackageQuestion args = %+v, want ordered positions", store.addArgs)
	}
}

func TestCreateCbtPackageHelperRejectsCrossEventQuestion(t *testing.T) {
	eventID := documentCycleTestUUID(212)
	otherEventID := documentCycleTestUUID(213)
	subjectID := documentCycleTestUUID(214)
	questionID := documentCycleTestUUID(215)
	store := &fakeCbtPackageStore{
		questionRow: db.GetCbtQuestionRow{ID: questionID, EventID: otherEventID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
	}

	_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
		EventID:     eventID,
		SubjectID:   subjectID,
		Title:       "Paket Event",
		QuestionIDs: []pgtype.UUID{questionID},
	})
	if err == nil || !strings.Contains(err.Error(), "event yang sama") {
		t.Fatalf("createCbtPackage(cross event) error = %v, want same-event validation", err)
	}
}

func TestCreateCbtPackageHelperPropagatesStoreErrors(t *testing.T) {
	subjectID := documentCycleTestUUID(208)
	questionID := documentCycleTestUUID(209)
	baseInput := CreateCbtPackageInput{
		SubjectID:   subjectID,
		Title:       "Paket",
		QuestionIDs: []pgtype.UUID{questionID},
	}

	tests := []struct {
		name    string
		store   *fakeCbtPackageStore
		wantErr string
	}{
		{name: "create error", store: &fakeCbtPackageStore{createErr: errors.New("create failed")}, wantErr: "create failed"},
		{name: "question error", store: &fakeCbtPackageStore{questionErr: errors.New("question failed")}, wantErr: "question failed"},
		{name: "subject mismatch", store: &fakeCbtPackageStore{questionRow: db.GetCbtQuestionRow{ID: questionID, SubjectID: documentCycleTestUUID(210)}}, wantErr: "semua soal harus dari mapel yang sama"},
		{name: "invalid weight", store: &fakeCbtPackageStore{}, wantErr: "bobot soal harus 1-100"},
		{name: "add error", store: &fakeCbtPackageStore{addErr: errors.New("add failed")}, wantErr: "add failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := baseInput
			if tt.name == "invalid weight" {
				input.QuestionWeights = map[string]int32{pgUUIDString(questionID): 0}
			}
			_, err := createCbtPackage(context.Background(), tt.store, input)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("createCbtPackage() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

type fakeCbtSessionStore struct {
	listRows               []db.ListCbtExamSessionsRow
	listErr                error
	getID                  pgtype.UUID
	sessionRow             db.GetCbtExamSessionRow
	packageRows            []db.ListCbtPackagesRow
	packageRowsErr         error
	packageQualityID       pgtype.UUID
	packageQualityRow      db.GetCbtPackageQuestionQualityRow
	packageQualityErr      error
	createArg              db.CreateCbtExamSessionParams
	updateStatusArg        db.UpdateCbtExamSessionStatusParams
	updateScheduleArg      db.UpdateCbtExamSessionScheduleParams
	auditArg               db.ListEntityAuditLogsParams
	auditRows              []db.ListEntityAuditLogsRow
	auditErr               error
	deleteID               pgtype.UUID
	participantRows        []db.ListCbtExamParticipantsRow
	participantTeacherRows []db.ListCbtExamParticipantsByTeacherRow
	participantTeacherArg  db.ListCbtExamParticipantsByTeacherParams
	participantsErr        error
	enrollClassArg         db.EnrollClassToSessionParams
	enrollGradeArg         db.EnrollGradeToSessionParams
	enrollSchoolID         pgtype.UUID
	generateID             pgtype.UUID
	regenerateID           pgtype.UUID
	roomRows               []db.ListCbtExamRoomsRow
	roomsErr               error
	createRoomArg          db.CreateCbtExamRoomParams
	schoolRoomRow          db.SchoolRoom
	roomProctorRows        []db.ListCbtRoomProctorsRow
	roomDashboardRow       db.GetCbtRoomProctorDashboardRow
	roomHandoverRow        db.GetCbtRoomHandoverRow
	roomHandoverID         pgtype.UUID
	operationalRecapID     pgtype.UUID
	operationalRecapRow    db.GetCbtSessionOperationalRecapRow
	operationalRoomRows    []db.ListCbtSessionRoomOperationalRecapRow
	operationalRoomID      pgtype.UUID
	saveHandoverArg        db.UpsertCbtRoomHandoverParams
	saveHandoverErr        error
	lockHandoverArg        db.LockCbtRoomHandoverParams
	lockHandoverErr        error
	proctorRoomRows        []db.ListCbtProctorRoomsRow
	proctorRoomArg         db.ListCbtProctorRoomsParams
	proctorRoomErr         error
	roomReadinessRow       db.GetCbtSessionRoomReadinessRow
	roomReadinessID        pgtype.UUID
	roomProctorArg         db.HasSessionRoomProctorParams
	roomProctor            bool
	roomParticipantArg     db.HasSessionRoomParticipantParams
	roomParticipant        bool
	deleteRoomID           pgtype.UUID
	clearRoomID            pgtype.UUID
	clearRoomErr           error
	assignRoomArgs         []db.AssignParticipantRoomParams
	assignRoomErr          error
	assignSeatArgs         []db.AssignParticipantSeatParams
	byRoomRows             []db.ListParticipantsByRoomRow
	byRoomErr              error
	proctorRows            []db.GetSessionProctoringStatusRow
	proctorStatusArg       db.GetSessionProctoringStatusParams
	proctorErr             error
	flagArg                db.SetParticipantSuspiciousFlagParams
	gradeArg               db.GradeStudentEssayParams
	ungradedRows           []db.ListUngradedEssaysRow
	ungradedTeacherRows    []db.ListUngradedEssaysByTeacherRow
	ungradedTeacherArg     db.ListUngradedEssaysByTeacherParams
	ungradedErr            error
	answerArg              db.UpsertStudentAnswerParams
	teacherRows            []db.ListCbtExamSessionsByTeacherRow
	teacherRowsErr         error
	teacherID              pgtype.UUID
	resultsTeacherArg      db.GetSessionResultsByTeacherParams
	resultsTeacherErr      error
	teacherAccess          bool
	teacherAccessArg       db.GetSessionTeacherAccessParams
	sessionParticipantArg  db.HasSessionParticipantParams
	teacherParticipantArg  db.HasSessionParticipantByTeacherParams
	sessionParticipant     bool
	sessionRoomArg         db.HasSessionRoomParams
	sessionRoom            bool
	sessionAnswerArg       db.HasSessionAnswerParams
	teacherAnswerArg       db.HasSessionAnswerByTeacherParams
	sessionAnswer          bool
	questionScopeArg       db.QuestionBelongsToParticipantPackageParams
	questionOutsidePackage bool
	questionScopeErr       error
	resultsRows            []db.GetSessionResultsRow
	resultsErr             error
	resultsID              pgtype.UUID
	answersRows            []db.GetParticipantAnswersRow
	answersErr             error
	answersID              pgtype.UUID
	correctnessID          pgtype.UUID
	correctnessErr         error
	scoresID               pgtype.UUID
	scoresErr              error
}

func (f *fakeCbtSessionStore) ListCbtExamSessions(ctx context.Context) ([]db.ListCbtExamSessionsRow, error) {
	return f.listRows, f.listErr
}

func (f *fakeCbtSessionStore) GetCbtExamSession(ctx context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error) {
	f.getID = id
	if f.sessionRow.ID.Valid {
		return f.sessionRow, nil
	}
	return db.GetCbtExamSessionRow{ID: id, PackageID: documentCycleTestUUID(229), Title: "Sesi", Status: db.CbtSessionStatusEnumScheduled}, nil
}

func (f *fakeCbtSessionStore) ListCbtPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, error) {
	if f.packageRowsErr != nil {
		return nil, f.packageRowsErr
	}
	if f.packageRows != nil {
		return f.packageRows, nil
	}
	if f.sessionRow.PackageID.Valid {
		return []db.ListCbtPackagesRow{{ID: f.sessionRow.PackageID, EventID: f.sessionRow.EventID}}, nil
	}
	return []db.ListCbtPackagesRow{{ID: documentCycleTestUUID(229)}}, nil
}

func (f *fakeCbtSessionStore) GetCbtPackageQuestionQuality(ctx context.Context, id pgtype.UUID) (db.GetCbtPackageQuestionQualityRow, error) {
	f.packageQualityID = id
	if f.packageQualityErr != nil {
		return db.GetCbtPackageQuestionQualityRow{}, f.packageQualityErr
	}
	if !f.packageQualityRow.IsActive && f.packageQualityRow.TotalQuestions == 0 && f.packageQualityRow.PublishedQuestions == 0 && f.packageQualityRow.UnpublishedQuestions == 0 {
		return db.GetCbtPackageQuestionQualityRow{IsActive: true, TotalQuestions: 1, PublishedQuestions: 1}, nil
	}
	return f.packageQualityRow, nil
}

func (f *fakeCbtSessionStore) CreateCbtExamSession(ctx context.Context, arg db.CreateCbtExamSessionParams) (db.CbtExamSession, error) {
	f.createArg = arg
	return db.CbtExamSession{PackageID: arg.PackageID, ClassID: arg.ClassID, Title: arg.Title, ScopeType: arg.ScopeType, MixPolicy: arg.MixPolicy, AssignmentMode: arg.AssignmentMode}, nil
}

func (f *fakeCbtSessionStore) UpdateCbtExamSessionStatus(ctx context.Context, arg db.UpdateCbtExamSessionStatusParams) (db.CbtExamSession, error) {
	f.updateStatusArg = arg
	return db.CbtExamSession{ID: arg.ID, Status: arg.Status}, nil
}

func (f *fakeCbtSessionStore) UpdateCbtExamSessionSchedule(ctx context.Context, arg db.UpdateCbtExamSessionScheduleParams) (db.CbtExamSession, error) {
	f.updateScheduleArg = arg
	return db.CbtExamSession{ID: arg.ID, ScheduledStart: arg.ScheduledStart, ScheduledEnd: arg.ScheduledEnd}, nil
}

func (f *fakeCbtSessionStore) ListEntityAuditLogs(ctx context.Context, arg db.ListEntityAuditLogsParams) ([]db.ListEntityAuditLogsRow, error) {
	f.auditArg = arg
	return f.auditRows, f.auditErr
}

func (f *fakeCbtSessionStore) DeleteCbtExamSession(ctx context.Context, id pgtype.UUID) (int64, error) {
	f.deleteID = id
	return 1, nil
}

func (f *fakeCbtSessionStore) ListCbtExamParticipants(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error) {
	return f.participantRows, f.participantsErr
}

func (f *fakeCbtSessionStore) ListCbtExamParticipantsByTeacher(ctx context.Context, arg db.ListCbtExamParticipantsByTeacherParams) ([]db.ListCbtExamParticipantsByTeacherRow, error) {
	f.participantTeacherArg = arg
	return f.participantTeacherRows, f.participantsErr
}

func (f *fakeCbtSessionStore) EnrollClassToSession(ctx context.Context, arg db.EnrollClassToSessionParams) error {
	f.enrollClassArg = arg
	return nil
}

func (f *fakeCbtSessionStore) EnrollGradeToSession(ctx context.Context, arg db.EnrollGradeToSessionParams) error {
	f.enrollGradeArg = arg
	return nil
}

func (f *fakeCbtSessionStore) EnrollSchoolToSession(ctx context.Context, sessionID pgtype.UUID) error {
	f.enrollSchoolID = sessionID
	return nil
}

func (f *fakeCbtSessionStore) GenerateTokensForSession(ctx context.Context, sessionID pgtype.UUID) error {
	f.generateID = sessionID
	return nil
}

func (f *fakeCbtSessionStore) RegenerateParticipantToken(ctx context.Context, id pgtype.UUID) (db.RegenerateParticipantTokenRow, error) {
	f.regenerateID = id
	return db.RegenerateParticipantTokenRow{ID: id, Token: "ABC123"}, nil
}

func (f *fakeCbtSessionStore) ListCbtExamRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error) {
	return f.roomRows, f.roomsErr
}

func (f *fakeCbtSessionStore) CreateCbtExamRoom(ctx context.Context, arg db.CreateCbtExamRoomParams) (db.CbtExamRoom, error) {
	f.createRoomArg = arg
	return db.CbtExamRoom{SessionID: arg.SessionID, RoomName: arg.RoomName, Capacity: arg.Capacity}, nil
}

func (f *fakeCbtSessionStore) DeleteCbtExamRoom(ctx context.Context, id pgtype.UUID) error {
	f.deleteRoomID = id
	return nil
}

func (f *fakeCbtSessionStore) GetSchoolRoom(ctx context.Context, id pgtype.UUID) (db.SchoolRoom, error) {
	if f.schoolRoomRow.ID.Valid {
		return f.schoolRoomRow, nil
	}
	return db.SchoolRoom{ID: id, Name: "Lab Komputer", DefaultCapacity: 30, ExamCapacity: 30, IsExamEligible: true}, nil
}

func (f *fakeCbtSessionStore) GetCbtRoomProctorDashboard(ctx context.Context, id pgtype.UUID) (db.GetCbtRoomProctorDashboardRow, error) {
	if f.roomDashboardRow.ID.Valid {
		return f.roomDashboardRow, nil
	}
	return db.GetCbtRoomProctorDashboardRow{ID: id, RoomName: "Ruang 1"}, nil
}

func (f *fakeCbtSessionStore) GetCbtRoomHandover(ctx context.Context, id pgtype.UUID) (db.GetCbtRoomHandoverRow, error) {
	f.roomHandoverID = id
	if f.roomHandoverRow.RoomID.Valid {
		return f.roomHandoverRow, nil
	}
	return db.GetCbtRoomHandoverRow{RoomID: id, RoomName: "Ruang 1"}, nil
}

func (f *fakeCbtSessionStore) GetCbtSessionOperationalRecap(ctx context.Context, id pgtype.UUID) (db.GetCbtSessionOperationalRecapRow, error) {
	f.operationalRecapID = id
	if f.operationalRecapRow.SessionID.Valid {
		return f.operationalRecapRow, nil
	}
	return db.GetCbtSessionOperationalRecapRow{SessionID: id, SessionTitle: "Sesi", RoomCount: 1}, nil
}

func (f *fakeCbtSessionStore) ListCbtSessionRoomOperationalRecap(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtSessionRoomOperationalRecapRow, error) {
	f.operationalRoomID = sessionID
	return f.operationalRoomRows, nil
}

func (f *fakeCbtSessionStore) UpsertCbtRoomHandover(ctx context.Context, arg db.UpsertCbtRoomHandoverParams) (db.CbtRoomHandover, error) {
	f.saveHandoverArg = arg
	if f.saveHandoverErr != nil {
		return db.CbtRoomHandover{}, f.saveHandoverErr
	}
	return db.CbtRoomHandover{
		ExamRoomID:            arg.ExamRoomID,
		AttendanceChecked:     arg.AttendanceChecked,
		AllSubmittedChecked:   arg.AllSubmittedChecked,
		DeviceIssueChecked:    arg.DeviceIssueChecked,
		RoomCleanChecked:      arg.RoomCleanChecked,
		TokenReturnedChecked:  arg.TokenReturnedChecked,
		AssetsReturnedChecked: arg.AssetsReturnedChecked,
		IncidentNotes:         arg.IncidentNotes,
		OperatorNotes:         arg.OperatorNotes,
		HandoverNotes:         arg.HandoverNotes,
		UpdatedBy:             arg.UpdatedBy,
	}, nil
}

func (f *fakeCbtSessionStore) LockCbtRoomHandover(ctx context.Context, arg db.LockCbtRoomHandoverParams) (db.CbtRoomHandover, error) {
	f.lockHandoverArg = arg
	if f.lockHandoverErr != nil {
		return db.CbtRoomHandover{}, f.lockHandoverErr
	}
	return db.CbtRoomHandover{ExamRoomID: arg.ExamRoomID, LockedBy: arg.LockedBy}, nil
}

func (f *fakeCbtSessionStore) ListCbtProctorRooms(ctx context.Context, arg db.ListCbtProctorRoomsParams) ([]db.ListCbtProctorRoomsRow, error) {
	f.proctorRoomArg = arg
	return f.proctorRoomRows, f.proctorRoomErr
}

func (f *fakeCbtSessionStore) ListCbtRoomProctors(ctx context.Context, examRoomID pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error) {
	return f.roomProctorRows, nil
}

func (f *fakeCbtSessionStore) DeleteCbtRoomProctorsByRoom(ctx context.Context, examRoomID pgtype.UUID) error {
	return nil
}

func (f *fakeCbtSessionStore) CreateCbtRoomProctor(ctx context.Context, arg db.CreateCbtRoomProctorParams) (db.CbtRoomProctor, error) {
	return db.CbtRoomProctor{ExamRoomID: arg.ExamRoomID, EmployeeID: arg.EmployeeID, Role: arg.Role, AssignedBy: arg.AssignedBy}, nil
}

func (f *fakeCbtSessionStore) GetCbtSessionRoomReadiness(ctx context.Context, targetSessionID pgtype.UUID) (db.GetCbtSessionRoomReadinessRow, error) {
	f.roomReadinessID = targetSessionID
	return f.roomReadinessRow, nil
}

func (f *fakeCbtSessionStore) HasSessionRoomProctor(ctx context.Context, arg db.HasSessionRoomProctorParams) (bool, error) {
	f.roomProctorArg = arg
	return f.roomProctor, nil
}

func (f *fakeCbtSessionStore) HasSessionRoomParticipant(ctx context.Context, arg db.HasSessionRoomParticipantParams) (bool, error) {
	f.roomParticipantArg = arg
	return f.roomParticipant, nil
}

func (f *fakeCbtSessionStore) ClearParticipantRooms(ctx context.Context, sessionID pgtype.UUID) error {
	f.clearRoomID = sessionID
	return f.clearRoomErr
}

func (f *fakeCbtSessionStore) AssignParticipantRoom(ctx context.Context, arg db.AssignParticipantRoomParams) error {
	f.assignRoomArgs = append(f.assignRoomArgs, arg)
	return f.assignRoomErr
}

func (f *fakeCbtSessionStore) AssignParticipantSeat(ctx context.Context, arg db.AssignParticipantSeatParams) error {
	f.assignSeatArgs = append(f.assignSeatArgs, arg)
	return nil
}

func (f *fakeCbtSessionStore) ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error) {
	return f.byRoomRows, f.byRoomErr
}

func (f *fakeCbtSessionStore) GetSessionProctoringStatus(ctx context.Context, arg db.GetSessionProctoringStatusParams) ([]db.GetSessionProctoringStatusRow, error) {
	f.proctorStatusArg = arg
	return f.proctorRows, f.proctorErr
}

func (f *fakeCbtSessionStore) SetParticipantSuspiciousFlag(ctx context.Context, arg db.SetParticipantSuspiciousFlagParams) error {
	f.flagArg = arg
	return nil
}

func (f *fakeCbtSessionStore) GradeStudentEssay(ctx context.Context, arg db.GradeStudentEssayParams) error {
	f.gradeArg = arg
	return nil
}

func (f *fakeCbtSessionStore) ListUngradedEssays(ctx context.Context, sessionID pgtype.UUID) ([]db.ListUngradedEssaysRow, error) {
	return f.ungradedRows, f.ungradedErr
}

func (f *fakeCbtSessionStore) ListUngradedEssaysByTeacher(ctx context.Context, arg db.ListUngradedEssaysByTeacherParams) ([]db.ListUngradedEssaysByTeacherRow, error) {
	f.ungradedTeacherArg = arg
	return f.ungradedTeacherRows, f.ungradedErr
}

func (f *fakeCbtSessionStore) UpsertStudentAnswer(ctx context.Context, arg db.UpsertStudentAnswerParams) error {
	f.answerArg = arg
	return nil
}

func (f *fakeCbtSessionStore) QuestionBelongsToParticipantPackage(ctx context.Context, arg db.QuestionBelongsToParticipantPackageParams) (bool, error) {
	f.questionScopeArg = arg
	if f.questionScopeErr != nil {
		return false, f.questionScopeErr
	}
	return !f.questionOutsidePackage, nil
}

func (f *fakeCbtSessionStore) ListCbtExamSessionsByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamSessionsByTeacherRow, error) {
	f.teacherID = teacherEmployeeID
	return f.teacherRows, f.teacherRowsErr
}

func (f *fakeCbtSessionStore) GetSessionResultsByTeacher(ctx context.Context, arg db.GetSessionResultsByTeacherParams) ([]db.GetSessionResultsByTeacherRow, error) {
	f.resultsTeacherArg = arg
	return nil, f.resultsTeacherErr
}

func (f *fakeCbtSessionStore) GetSessionTeacherAccess(ctx context.Context, arg db.GetSessionTeacherAccessParams) (bool, error) {
	f.teacherAccessArg = arg
	return f.teacherAccess, nil
}

func (f *fakeCbtSessionStore) HasSessionParticipant(ctx context.Context, arg db.HasSessionParticipantParams) (bool, error) {
	f.sessionParticipantArg = arg
	return f.sessionParticipant, nil
}

func (f *fakeCbtSessionStore) HasSessionParticipantByTeacher(ctx context.Context, arg db.HasSessionParticipantByTeacherParams) (bool, error) {
	f.teacherParticipantArg = arg
	return f.sessionParticipant, nil
}

func (f *fakeCbtSessionStore) HasSessionRoom(ctx context.Context, arg db.HasSessionRoomParams) (bool, error) {
	f.sessionRoomArg = arg
	return f.sessionRoom, nil
}

func (f *fakeCbtSessionStore) HasSessionAnswer(ctx context.Context, arg db.HasSessionAnswerParams) (bool, error) {
	f.sessionAnswerArg = arg
	return f.sessionAnswer, nil
}

func (f *fakeCbtSessionStore) HasSessionAnswerByTeacher(ctx context.Context, arg db.HasSessionAnswerByTeacherParams) (bool, error) {
	f.teacherAnswerArg = arg
	return f.sessionAnswer, nil
}

func (f *fakeCbtSessionStore) GetSessionResults(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionResultsRow, error) {
	f.resultsID = sessionID
	return f.resultsRows, f.resultsErr
}

func (f *fakeCbtSessionStore) GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error) {
	f.answersID = participantID
	return f.answersRows, f.answersErr
}

func (f *fakeCbtSessionStore) WithTx(tx pgx.Tx) *db.Queries {
	return nil
}

func (f *fakeCbtSessionStore) UpdateAnswerCorrectness(ctx context.Context, sessionID pgtype.UUID) error {
	f.correctnessID = sessionID
	return f.correctnessErr
}

func (f *fakeCbtSessionStore) UpdateParticipantScores(ctx context.Context, sessionID pgtype.UUID) error {
	f.scoresID = sessionID
	return f.scoresErr
}

func TestCbtSessionServiceForwardsStoreCalls(t *testing.T) {
	sessionID := documentCycleTestUUID(211)
	packageID := documentCycleTestUUID(212)
	classID := documentCycleTestUUID(213)
	eventID := documentCycleTestUUID(214)
	participantID := documentCycleTestUUID(215)
	questionID := documentCycleTestUUID(216)
	roomID := documentCycleTestUUID(217)
	answerID := documentCycleTestUUID(218)
	teacherID := documentCycleTestUUID(219)
	store := &fakeCbtSessionStore{
		listRows: []db.ListCbtExamSessionsRow{{ID: sessionID, Title: "Sesi"}},
		packageRows: []db.ListCbtPackagesRow{
			{ID: packageID, EventID: eventID, Title: "Paket"},
			{ID: documentCycleTestUUID(229), Title: "Paket Default"},
		},
		participantRows: []db.ListCbtExamParticipantsRow{{ID: participantID}},
		roomRows:        []db.ListCbtExamRoomsRow{{ID: roomID, RoomName: "R1"}},
		proctorRoomRows: []db.ListCbtProctorRoomsRow{{ID: roomID, RoomName: "R1"}},
		operationalRoomRows: []db.ListCbtSessionRoomOperationalRecapRow{{
			RoomID: roomID, RoomName: "R1", ParticipantCount: 1,
		}},
		proctorRows:        []db.GetSessionProctoringStatusRow{{ParticipantID: participantID}},
		ungradedRows:       []db.ListUngradedEssaysRow{{AnswerID: answerID}},
		teacherRows:        []db.ListCbtExamSessionsByTeacherRow{{ID: sessionID}},
		resultsRows:        []db.GetSessionResultsRow{{ParticipantID: participantID}},
		answersRows:        []db.GetParticipantAnswersRow{{ID: answerID}},
		teacherAccess:      true,
		sessionParticipant: true,
		sessionRoom:        true,
		sessionAnswer:      true,
		roomProctor:        true,
		roomParticipant:    true,
		roomReadinessRow: db.GetCbtSessionRoomReadinessRow{
			RoomCount:                  1,
			TotalCapacity:              30,
			ParticipantCount:           1,
			AssignedParticipantCount:   1,
			UnassignedParticipantCount: 0,
			MissingSeatCount:           0,
			RoomsWithoutProctor:        0,
			ProctorAssignmentCount:     1,
		},
	}
	svc := &CbtSession{q: store}
	if NewCbtSession(nil) == nil {
		t.Fatal("NewCbtSession(nil) = nil, want service")
	}

	if rows, err := svc.List(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("List() = %d rows/%v, want 1 nil", len(rows), err)
	}
	store.listRows = nil
	if rows, err := svc.List(context.Background()); err != nil || len(rows) != 0 {
		t.Fatalf("List(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if got, err := svc.Get(context.Background(), sessionID); err != nil || got.ID != sessionID || store.getID != sessionID {
		t.Fatalf("Get() = %+v/%v id=%v, want session", got, err, store.getID)
	}
	if _, err := svc.Create(context.Background(), CreateCbtSessionInput{
		PackageID:       packageID,
		ClassID:         classID,
		EventID:         eventID,
		ScopeType:       "grade",
		ScopeRef:        "9",
		MixPolicy:       "ignored",
		AssignmentMode:  "manual",
		AllowCrossGrade: true,
		IsSpecialEvent:  true,
		Title:           "Sesi",
		Status:          db.CbtSessionStatusEnumDraft,
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if store.createArg.PackageID != packageID || store.createArg.ScopeType != "grade" || store.createArg.MixPolicy != "same_grade" || store.createArg.AssignmentMode != "manual" {
		t.Fatalf("Create() arg = %+v, want normalized session", store.createArg)
	}
	if store.packageQualityID != packageID {
		t.Fatalf("Create() package quality id = %v, want %v", store.packageQualityID, packageID)
	}
	if _, err := svc.UpdateStatus(context.Background(), sessionID, db.CbtSessionStatusEnumActive); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if store.updateStatusArg.ID != sessionID || store.updateStatusArg.Status != db.CbtSessionStatusEnumActive {
		t.Fatalf("UpdateStatus() arg = %+v, want active status", store.updateStatusArg)
	}
	if store.roomReadinessID != sessionID {
		t.Fatalf("UpdateStatus() readiness id = %v, want %v", store.roomReadinessID, sessionID)
	}
	if err := svc.Delete(context.Background(), sessionID); err != nil || store.deleteID != sessionID {
		t.Fatalf("Delete() = %v id=%v, want nil/%v", err, store.deleteID, sessionID)
	}
	if rows, err := svc.ListParticipants(context.Background(), sessionID); err != nil || len(rows) != 1 {
		t.Fatalf("ListParticipants() = %d rows/%v, want 1 nil", len(rows), err)
	}
	store.participantRows = nil
	if rows, err := svc.ListParticipants(context.Background(), sessionID); err != nil || len(rows) != 0 {
		t.Fatalf("ListParticipants(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if err := svc.EnrollClass(context.Background(), sessionID, classID); err != nil || store.enrollClassArg.SessionID != sessionID || store.enrollClassArg.ClassID != classID {
		t.Fatalf("EnrollClass() = %v arg=%+v, want ids", err, store.enrollClassArg)
	}
	if err := svc.EnrollGrade(context.Background(), sessionID, "9"); err != nil || store.enrollGradeArg.SessionID != sessionID || store.enrollGradeArg.Level != "9" {
		t.Fatalf("EnrollGrade() = %v arg=%+v, want level", err, store.enrollGradeArg)
	}
	if err := svc.EnrollSchool(context.Background(), sessionID); err != nil || store.enrollSchoolID != sessionID {
		t.Fatalf("EnrollSchool() = %v id=%v, want %v", err, store.enrollSchoolID, sessionID)
	}
	store.sessionRow = db.GetCbtExamSessionRow{ID: sessionID, PackageID: packageID, Status: db.CbtSessionStatusEnumScheduled}
	if err := svc.GenerateTokens(context.Background(), sessionID); err != nil || store.generateID != sessionID {
		t.Fatalf("GenerateTokens() = %v id=%v, want %v", err, store.generateID, sessionID)
	}
	if got, err := svc.RegenerateToken(context.Background(), participantID); err != nil || got.ID != participantID || got.Token != "ABC123" || store.regenerateID != participantID {
		t.Fatalf("RegenerateToken() = %+v/%v id=%v, want token", got, err, store.regenerateID)
	}
	if ok, err := svc.HasParticipant(context.Background(), sessionID, participantID); err != nil || !ok || store.sessionParticipantArg.SessionID != sessionID || store.sessionParticipantArg.ID != participantID {
		t.Fatalf("HasParticipant() = %v/%v arg=%+v, want true session/participant", ok, err, store.sessionParticipantArg)
	}
	if rows, err := svc.ListRooms(context.Background(), sessionID); err != nil || len(rows) != 1 {
		t.Fatalf("ListRooms() = %d rows/%v, want 1 nil", len(rows), err)
	}
	store.roomRows = nil
	if rows, err := svc.ListRooms(context.Background(), sessionID); err != nil || len(rows) != 0 {
		t.Fatalf("ListRooms(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if _, err := svc.CreateRoom(context.Background(), sessionID, "R1", 30); err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}
	if store.createRoomArg.SessionID != sessionID || store.createRoomArg.RoomName != "R1" || store.createRoomArg.Capacity != 30 {
		t.Fatalf("CreateRoom() arg = %+v, want room", store.createRoomArg)
	}
	if err := svc.DeleteRoom(context.Background(), roomID); err != nil || store.deleteRoomID != roomID {
		t.Fatalf("DeleteRoom() = %v id=%v, want nil/%v", err, store.deleteRoomID, roomID)
	}
	if ok, err := svc.HasRoom(context.Background(), sessionID, roomID); err != nil || !ok || store.sessionRoomArg.SessionID != sessionID || store.sessionRoomArg.ID != roomID {
		t.Fatalf("HasRoom() = %v/%v arg=%+v, want true session/room", ok, err, store.sessionRoomArg)
	}
	if got, err := svc.GetRoomProctoringDashboard(context.Background(), roomID); err != nil || got.ID != roomID {
		t.Fatalf("GetRoomProctoringDashboard() = %+v/%v, want room dashboard", got, err)
	}
	if got, err := svc.GetRoomHandover(context.Background(), roomID); err != nil || got.RoomID != roomID || store.roomHandoverID != roomID {
		t.Fatalf("GetRoomHandover() = %+v/%v id=%v, want room handover", got, err, store.roomHandoverID)
	}
	if recap, roomRows, err := svc.GetSessionOperationalRecap(context.Background(), sessionID); err != nil || recap.SessionID != sessionID || len(roomRows) != 1 || store.operationalRecapID != sessionID || store.operationalRoomID != sessionID {
		t.Fatalf("GetSessionOperationalRecap() = %+v/%d/%v ids=%v/%v, want recap and room rows", recap, len(roomRows), err, store.operationalRecapID, store.operationalRoomID)
	}
	if got, err := svc.SaveRoomHandover(context.Background(), roomID, teacherID, SaveCbtRoomHandoverInput{
		AttendanceChecked:     true,
		AllSubmittedChecked:   true,
		DeviceIssueChecked:    true,
		RoomCleanChecked:      true,
		TokenReturnedChecked:  true,
		AssetsReturnedChecked: true,
		IncidentNotes:         "  kejadian  ",
		OperatorNotes:         "  operator  ",
		HandoverNotes:         "  selesai  ",
	}); err != nil || got.ExamRoomID != roomID {
		t.Fatalf("SaveRoomHandover() = %+v/%v, want saved handover", got, err)
	}
	if store.saveHandoverArg.ExamRoomID != roomID || store.saveHandoverArg.UpdatedBy != teacherID || store.saveHandoverArg.IncidentNotes != "kejadian" || !store.saveHandoverArg.AssetsReturnedChecked {
		t.Fatalf("SaveRoomHandover() arg = %+v, want trimmed room handover", store.saveHandoverArg)
	}
	if got, err := svc.LockRoomHandover(context.Background(), roomID, teacherID); err != nil || got.ExamRoomID != roomID || store.lockHandoverArg.LockedBy != teacherID {
		t.Fatalf("LockRoomHandover() = %+v/%v arg=%+v, want locked handover", got, err, store.lockHandoverArg)
	}
	if rows, err := svc.ListProctorRooms(context.Background(), teacherID, false); err != nil || len(rows) != 1 || store.proctorRoomArg.EmployeeID != teacherID || store.proctorRoomArg.IncludeAll {
		t.Fatalf("ListProctorRooms() = %d rows/%v arg=%+v, want employee scoped room", len(rows), err, store.proctorRoomArg)
	}
	store.proctorRoomRows = nil
	if rows, err := svc.ListProctorRooms(context.Background(), teacherID, true); err != nil || len(rows) != 0 || store.proctorRoomArg.EmployeeID != teacherID || !store.proctorRoomArg.IncludeAll {
		t.Fatalf("ListProctorRooms(nil rows) = %d rows/%v arg=%+v, want empty all-rooms query", len(rows), err, store.proctorRoomArg)
	}
	if ok, err := svc.HasRoomProctor(context.Background(), sessionID, roomID, teacherID); err != nil || !ok || store.roomProctorArg.EmployeeID != teacherID {
		t.Fatalf("HasRoomProctor() = %v/%v arg=%+v, want teacher proctor", ok, err, store.roomProctorArg)
	}
	if ok, err := svc.HasRoomParticipant(context.Background(), sessionID, roomID, participantID); err != nil || !ok || store.roomParticipantArg.ParticipantID != participantID {
		t.Fatalf("HasRoomParticipant() = %v/%v arg=%+v, want participant in room", ok, err, store.roomParticipantArg)
	}
	if err := svc.AssignSeat(context.Background(), participantID, roomID, 12); err != nil {
		t.Fatalf("AssignSeat() error = %v", err)
	}
	if got := store.assignSeatArgs[len(store.assignSeatArgs)-1]; got.ID != participantID || got.RoomID != roomID || got.SeatNo.Int32 != 12 || !got.SeatNo.Valid {
		t.Fatalf("AssignSeat() arg = %+v, want seat 12", got)
	}
	if rows, err := svc.GetProctoringStatus(context.Background(), sessionID); err != nil || len(rows) != 1 {
		t.Fatalf("GetProctoringStatus() = %d rows/%v, want 1 nil", len(rows), err)
	}
	if store.proctorStatusArg.SessionID != sessionID || store.proctorStatusArg.RoomID.Valid {
		t.Fatalf("GetProctoringStatus() arg = %+v, want session without room", store.proctorStatusArg)
	}
	if _, err := svc.GetProctoringStatusForRoom(context.Background(), sessionID, roomID); err != nil || store.proctorStatusArg.RoomID != roomID {
		t.Fatalf("GetProctoringStatusForRoom() = %v arg=%+v, want room filter", err, store.proctorStatusArg)
	}
	store.proctorRows = nil
	if rows, err := svc.GetProctoringStatus(context.Background(), sessionID); err != nil || len(rows) != 0 {
		t.Fatalf("GetProctoringStatus(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if err := svc.SetSuspiciousFlag(context.Background(), participantID, true); err != nil || store.flagArg.ID != participantID || !store.flagArg.SuspiciousFlag {
		t.Fatalf("SetSuspiciousFlag() = %v arg=%+v, want true flag", err, store.flagArg)
	}
	if err := svc.GradeEssay(context.Background(), sessionID, answerID, 87.5, "guru"); err != nil || store.gradeArg.ID != answerID || store.gradeArg.GradedBy.String != "guru" {
		t.Fatalf("GradeEssay() = %v arg=%+v, want graded answer", err, store.gradeArg)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("GradeEssay() score refresh ids = %v/%v, want %v", store.correctnessID, store.scoresID, sessionID)
	}
	if ok, err := svc.HasAnswer(context.Background(), sessionID, answerID); err != nil || !ok || store.sessionAnswerArg.SessionID != sessionID || store.sessionAnswerArg.ID != answerID {
		t.Fatalf("HasAnswer() = %v/%v arg=%+v, want true session/answer", ok, err, store.sessionAnswerArg)
	}
	if rows, err := svc.ListUngradedEssays(context.Background(), sessionID); err != nil || len(rows) != 1 {
		t.Fatalf("ListUngradedEssays() = %d rows/%v, want 1 nil", len(rows), err)
	}
	store.ungradedRows = nil
	if rows, err := svc.ListUngradedEssays(context.Background(), sessionID); err != nil || len(rows) != 0 {
		t.Fatalf("ListUngradedEssays(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if err := svc.RecordAnswer(context.Background(), participantID, questionID, "A"); err != nil ||
		store.questionScopeArg.ID != participantID || store.questionScopeArg.QuestionID != questionID ||
		store.answerArg.ParticipantID != participantID || store.answerArg.QuestionID != questionID || store.answerArg.Answer != "A" {
		t.Fatalf("RecordAnswer() = %v scope=%+v arg=%+v, want scoped answer", err, store.questionScopeArg, store.answerArg)
	}
	if err := (&CbtSession{q: &fakeCbtSessionStore{questionOutsidePackage: true}}).RecordAnswer(context.Background(), participantID, questionID, "A"); !errors.Is(err, ErrExamQuestionScope) {
		t.Fatalf("RecordAnswer(outside package) = %v, want ErrExamQuestionScope", err)
	}
	if rows, err := svc.ListByTeacher(context.Background(), teacherID); err != nil || len(rows) != 1 || store.teacherID != teacherID {
		t.Fatalf("ListByTeacher() = %d rows/%v id=%v, want teacher rows", len(rows), err, store.teacherID)
	}
	store.teacherRows = nil
	if rows, err := svc.ListByTeacher(context.Background(), teacherID); err != nil || len(rows) != 0 {
		t.Fatalf("ListByTeacher(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if rows, err := svc.GetResultsByTeacher(context.Background(), sessionID, teacherID); err != nil || len(rows) != 0 || store.resultsTeacherArg.SessionID != sessionID || store.resultsTeacherArg.TeacherEmployeeID != teacherID {
		t.Fatalf("GetResultsByTeacher() = %d rows/%v arg=%+v, want args", len(rows), err, store.resultsTeacherArg)
	}
	if ok, err := svc.CheckTeacherAccess(context.Background(), sessionID, teacherID); err != nil || !ok || store.teacherAccessArg.ID != sessionID || store.teacherAccessArg.TeacherEmployeeID != teacherID {
		t.Fatalf("CheckTeacherAccess() = %v/%v arg=%+v, want access", ok, err, store.teacherAccessArg)
	}
	if rows, err := svc.GetResults(context.Background(), sessionID); err != nil || len(rows) != 1 || store.resultsID != sessionID {
		t.Fatalf("GetResults() = %d rows/%v id=%v, want results", len(rows), err, store.resultsID)
	}
	store.resultsRows = nil
	if rows, err := svc.GetResults(context.Background(), sessionID); err != nil || len(rows) != 0 {
		t.Fatalf("GetResults(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
	if rows, err := svc.GetParticipantAnswers(context.Background(), participantID); err != nil || len(rows) != 1 || store.answersID != participantID {
		t.Fatalf("GetParticipantAnswers() = %d rows/%v id=%v, want answers", len(rows), err, store.answersID)
	}
	store.answersRows = nil
	if rows, err := svc.GetParticipantAnswers(context.Background(), participantID); err != nil || len(rows) != 0 {
		t.Fatalf("GetParticipantAnswers(nil rows) = %d rows/%v, want empty nil", len(rows), err)
	}
}

func TestCbtSessionCreateRejectsUnsafePackageQuality(t *testing.T) {
	packageID := documentCycleTestUUID(229)
	tests := []struct {
		name    string
		row     db.GetCbtPackageQuestionQualityRow
		err     error
		wantErr string
	}{
		{
			name:    "inactive package",
			row:     db.GetCbtPackageQuestionQualityRow{IsActive: false, TotalQuestions: 1, PublishedQuestions: 1},
			wantErr: "paket soal tidak aktif",
		},
		{
			name:    "empty package",
			row:     db.GetCbtPackageQuestionQualityRow{IsActive: true},
			wantErr: "paket soal belum memiliki soal",
		},
		{
			name:    "no published questions",
			row:     db.GetCbtPackageQuestionQualityRow{IsActive: true, TotalQuestions: 2},
			wantErr: "paket soal belum memiliki soal terbit",
		},
		{
			name:    "contains unpublished questions",
			row:     db.GetCbtPackageQuestionQualityRow{IsActive: true, TotalQuestions: 3, PublishedQuestions: 2, UnpublishedQuestions: 1},
			wantErr: "paket soal masih memiliki 1 soal belum terbit",
		},
		{
			name:    "quality query error",
			err:     errors.New("quality failed"),
			wantErr: "quality failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeCbtSessionStore{packageQualityRow: tt.row, packageQualityErr: tt.err}
			svc := &CbtSession{q: store}
			_, err := svc.Create(context.Background(), CreateCbtSessionInput{
				PackageID: packageID,
				ScopeType: "school",
				Title:     "Sesi",
				Status:    db.CbtSessionStatusEnumDraft,
			})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Create() error = %v, want contains %q", err, tt.wantErr)
			}
			if store.packageQualityID != packageID {
				t.Fatalf("Create() package quality id = %v, want %v", store.packageQualityID, packageID)
			}
			if store.createArg.PackageID.Valid {
				t.Fatalf("Create() called CreateCbtExamSession with %+v, want blocked before insert", store.createArg)
			}
		})
	}
}

func TestCbtSessionCreateRejectsPackageEventMismatch(t *testing.T) {
	packageID := documentCycleTestUUID(235)
	eventID := documentCycleTestUUID(236)
	otherEventID := documentCycleTestUUID(237)
	tests := []struct {
		name           string
		sessionEventID pgtype.UUID
		packageEventID pgtype.UUID
		wantErr        string
	}{
		{
			name:           "event session cannot use global package",
			sessionEventID: eventID,
			wantErr:        "event yang sama",
		},
		{
			name:           "event session cannot use other event package",
			sessionEventID: eventID,
			packageEventID: otherEventID,
			wantErr:        "event yang sama",
		},
		{
			name:           "non event session cannot use event package",
			packageEventID: eventID,
			wantErr:        "sesi event yang sama",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeCbtSessionStore{
				packageRows: []db.ListCbtPackagesRow{{ID: packageID, EventID: tt.packageEventID}},
			}
			svc := &CbtSession{q: store}

			_, err := svc.Create(context.Background(), CreateCbtSessionInput{
				PackageID: packageID,
				EventID:   tt.sessionEventID,
				ScopeType: "school",
				Title:     "Sesi",
				Status:    db.CbtSessionStatusEnumDraft,
			})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Create() error = %v, want contains %q", err, tt.wantErr)
			}
			if store.packageQualityID.Valid {
				t.Fatalf("Create() package quality id = %v, want blocked before quality check", store.packageQualityID)
			}
		})
	}
}

func TestCbtSessionUpdateStatusRechecksPackageQuality(t *testing.T) {
	sessionID := documentCycleTestUUID(231)
	packageID := documentCycleTestUUID(232)
	store := &fakeCbtSessionStore{
		sessionRow:        db.GetCbtExamSessionRow{ID: sessionID, PackageID: packageID, Title: "Sesi lama"},
		packageQualityRow: db.GetCbtPackageQuestionQualityRow{IsActive: true, TotalQuestions: 3, PublishedQuestions: 2, UnpublishedQuestions: 1},
	}
	svc := &CbtSession{q: store}

	_, err := svc.UpdateStatus(context.Background(), sessionID, db.CbtSessionStatusEnumScheduled)
	if err == nil || !strings.Contains(err.Error(), "paket soal masih memiliki 1 soal belum terbit") {
		t.Fatalf("UpdateStatus(scheduled) error = %v, want package quality conflict", err)
	}
	if store.getID != sessionID {
		t.Fatalf("UpdateStatus() session id = %v, want %v", store.getID, sessionID)
	}
	if store.packageQualityID != packageID {
		t.Fatalf("UpdateStatus() package quality id = %v, want %v", store.packageQualityID, packageID)
	}
	if store.updateStatusArg.ID.Valid {
		t.Fatalf("UpdateStatus() called update with %+v, want blocked before status write", store.updateStatusArg)
	}
}

func TestCbtSessionUpdateStatusRejectsExpiredActivation(t *testing.T) {
	sessionID := documentCycleTestUUID(233)
	packageID := documentCycleTestUUID(234)
	store := &fakeCbtSessionStore{
		sessionRow: db.GetCbtExamSessionRow{
			ID:           sessionID,
			PackageID:    packageID,
			Title:        "Sesi lewat jadwal",
			ScheduledEnd: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
		},
		roomReadinessRow: db.GetCbtSessionRoomReadinessRow{
			ParticipantCount:           1,
			RoomCount:                  1,
			TotalCapacity:              1,
			AssignedParticipantCount:   1,
			UnassignedParticipantCount: 0,
			MissingSeatCount:           0,
			RoomsWithoutProctor:        0,
		},
	}
	svc := &CbtSession{q: store}

	_, err := svc.UpdateStatus(context.Background(), sessionID, db.CbtSessionStatusEnumActive)
	if err == nil || !strings.Contains(err.Error(), "jadwal sesi sudah berakhir") {
		t.Fatalf("UpdateStatus(active) error = %v, want expired schedule conflict", err)
	}
	if store.roomReadinessID.Valid {
		t.Fatalf("UpdateStatus() readiness id = %v, want blocked before readiness check", store.roomReadinessID)
	}
	if store.updateStatusArg.ID.Valid {
		t.Fatalf("UpdateStatus() called update with %+v, want blocked before status write", store.updateStatusArg)
	}
}

func TestCbtSessionUpdateScheduleGuardsStatusAndWindow(t *testing.T) {
	sessionID := documentCycleTestUUID(242)
	packageID := documentCycleTestUUID(243)
	start := pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}
	end := pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true}
	store := &fakeCbtSessionStore{
		sessionRow: db.GetCbtExamSessionRow{
			ID:        sessionID,
			PackageID: packageID,
			Status:    db.CbtSessionStatusEnumScheduled,
		},
	}
	svc := &CbtSession{q: store}

	result, err := svc.UpdateSchedule(context.Background(), sessionID, start, end)
	if err != nil {
		t.Fatalf("UpdateSchedule() error = %v", err)
	}
	if result.Before.ID != sessionID || result.Session.ID != sessionID {
		t.Fatalf("UpdateSchedule() result = %+v, want before and updated session ids", result)
	}
	if store.updateScheduleArg.ID != sessionID || !store.updateScheduleArg.ScheduledStart.Time.Equal(start.Time) || !store.updateScheduleArg.ScheduledEnd.Time.Equal(end.Time) {
		t.Fatalf("UpdateSchedule() arg = %+v, want schedule params", store.updateScheduleArg)
	}

	store.updateScheduleArg = db.UpdateCbtExamSessionScheduleParams{}
	store.sessionRow.Status = db.CbtSessionStatusEnumActive
	_, err = svc.UpdateSchedule(context.Background(), sessionID, start, end)
	if err == nil || !strings.Contains(err.Error(), "hanya sesi draft atau terjadwal") {
		t.Fatalf("UpdateSchedule(active) error = %v, want status conflict", err)
	}
	if store.updateScheduleArg.ID.Valid {
		t.Fatalf("UpdateSchedule(active) wrote %+v, want blocked before update", store.updateScheduleArg)
	}
}

func TestCbtSessionListAuditLogsScopesEntity(t *testing.T) {
	sessionID := documentCycleTestUUID(244)
	store := &fakeCbtSessionStore{
		auditRows: []db.ListEntityAuditLogsRow{{Action: "CBT_SESSION_SCHEDULE_UPDATE"}},
	}
	svc := &CbtSession{q: store}

	rows, err := svc.ListAuditLogs(context.Background(), sessionID, 25, 50)
	if err != nil {
		t.Fatalf("ListAuditLogs() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Action != "CBT_SESSION_SCHEDULE_UPDATE" {
		t.Fatalf("ListAuditLogs() rows = %+v, want audit rows", rows)
	}
	if store.auditArg.EntityType != "cbt_session" || store.auditArg.EntityID != pgUUIDString(sessionID) || store.auditArg.Limit != 25 || store.auditArg.Offset != 50 {
		t.Fatalf("ListAuditLogs() arg = %+v, want cbt session entity scope", store.auditArg)
	}
}

func TestCbtSessionShuffleRoomsHelperAssignsByCapacity(t *testing.T) {
	sessionID := documentCycleTestUUID(230)
	roomOneID := documentCycleTestUUID(231)
	roomTwoID := documentCycleTestUUID(232)
	participantOneID := documentCycleTestUUID(233)
	participantTwoID := documentCycleTestUUID(234)
	participantThreeID := documentCycleTestUUID(235)
	store := &fakeCbtSessionStore{
		byRoomRows: []db.ListParticipantsByRoomRow{
			{ID: participantOneID, Nama: "A"},
			{ID: participantTwoID, Nama: "B"},
			{ID: participantThreeID, Nama: "C"},
		},
		roomRows: []db.ListCbtExamRoomsRow{
			{ID: roomOneID, Capacity: 1},
			{ID: roomTwoID, Capacity: 5},
		},
	}

	if err := shuffleRooms(context.Background(), store, sessionID); err != nil {
		t.Fatalf("shuffleRooms() error = %v", err)
	}
	if store.clearRoomID != sessionID {
		t.Fatalf("ClearParticipantRooms id = %v, want %v", store.clearRoomID, sessionID)
	}
	if len(store.assignRoomArgs) != 3 {
		t.Fatalf("assigned rooms = %d, want 3", len(store.assignRoomArgs))
	}
	seen := map[pgtype.UUID]bool{}
	for _, arg := range store.assignRoomArgs {
		seen[arg.ID] = true
		if arg.RoomID != roomOneID && arg.RoomID != roomTwoID {
			t.Fatalf("AssignParticipantRoom arg = %+v, want one of test rooms", arg)
		}
	}
	for _, id := range []pgtype.UUID{participantOneID, participantTwoID, participantThreeID} {
		if !seen[id] {
			t.Fatalf("participant %v was not assigned; args=%+v", id, store.assignRoomArgs)
		}
	}
}

func TestCbtSessionShuffleRoomsHelperHandlesEmptyAndErrors(t *testing.T) {
	sessionID := documentCycleTestUUID(236)
	expectedErr := errors.New("clear failed")
	if err := shuffleRooms(context.Background(), &fakeCbtSessionStore{clearRoomErr: expectedErr}, sessionID); !errors.Is(err, expectedErr) {
		t.Fatalf("shuffleRooms(clear error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("participants failed")
	if err := shuffleRooms(context.Background(), &fakeCbtSessionStore{byRoomErr: expectedErr}, sessionID); !errors.Is(err, expectedErr) {
		t.Fatalf("shuffleRooms(participants error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("rooms failed")
	if err := shuffleRooms(context.Background(), &fakeCbtSessionStore{
		byRoomRows: []db.ListParticipantsByRoomRow{{ID: documentCycleTestUUID(237)}},
		roomsErr:   expectedErr,
	}, sessionID); !errors.Is(err, expectedErr) {
		t.Fatalf("shuffleRooms(rooms error) = %v, want %v", err, expectedErr)
	}

	store := &fakeCbtSessionStore{roomRows: []db.ListCbtExamRoomsRow{{ID: documentCycleTestUUID(238), Capacity: 1}}}
	if err := shuffleRooms(context.Background(), store, sessionID); err != nil {
		t.Fatalf("shuffleRooms(empty participants) error = %v", err)
	}
	if len(store.assignRoomArgs) != 0 {
		t.Fatalf("shuffleRooms(empty participants) assigned %+v, want none", store.assignRoomArgs)
	}

	expectedErr = errors.New("assign failed")
	store = &fakeCbtSessionStore{
		byRoomRows:    []db.ListParticipantsByRoomRow{{ID: documentCycleTestUUID(239)}},
		roomRows:      []db.ListCbtExamRoomsRow{{ID: documentCycleTestUUID(240), Capacity: 1}},
		assignRoomErr: expectedErr,
	}
	if err := shuffleRooms(context.Background(), store, sessionID); !errors.Is(err, expectedErr) {
		t.Fatalf("shuffleRooms(assign error) = %v, want %v", err, expectedErr)
	}
}

func TestCbtSessionScoreSessionHelperPropagatesSteps(t *testing.T) {
	sessionID := documentCycleTestUUID(241)
	expectedErr := errors.New("correctness failed")
	if err := scoreSession(context.Background(), &fakeCbtSessionStore{correctnessErr: expectedErr}, sessionID); !errors.Is(err, expectedErr) {
		t.Fatalf("scoreSession(correctness error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("scores failed")
	if err := scoreSession(context.Background(), &fakeCbtSessionStore{scoresErr: expectedErr}, sessionID); !errors.Is(err, expectedErr) {
		t.Fatalf("scoreSession(scores error) = %v, want %v", err, expectedErr)
	}

	store := &fakeCbtSessionStore{}
	if err := scoreSession(context.Background(), store, sessionID); err != nil {
		t.Fatalf("scoreSession() error = %v", err)
	}
	if store.correctnessID != sessionID || store.scoresID != sessionID {
		t.Fatalf("scoreSession ids = correctness %v scores %v, want %v", store.correctnessID, store.scoresID, sessionID)
	}
}

func TestCbtSessionAutoAssignSeatsSortsWithinRooms(t *testing.T) {
	sessionID := documentCycleTestUUID(221)
	roomID := documentCycleTestUUID(222)
	otherRoomID := documentCycleTestUUID(223)
	anaID := documentCycleTestUUID(224)
	budiID := documentCycleTestUUID(225)
	ciciID := documentCycleTestUUID(226)
	unassignedID := documentCycleTestUUID(227)
	store := &fakeCbtSessionStore{
		byRoomRows: []db.ListParticipantsByRoomRow{
			{ID: budiID, RoomID: roomID, Nama: "Budi", Nis: "002"},
			{ID: unassignedID, Nama: "Tanpa Ruang", Nis: "004"},
			{ID: anaID, RoomID: roomID, Nama: "Ana", Nis: "001"},
			{ID: ciciID, RoomID: otherRoomID, Nama: "Cici", Nis: "003"},
		},
	}
	svc := &CbtSession{q: store}

	if err := svc.AutoAssignSeats(context.Background(), sessionID); err != nil {
		t.Fatalf("AutoAssignSeats() error = %v", err)
	}
	if len(store.assignSeatArgs) != 3 {
		t.Fatalf("AutoAssignSeats() assigned %d seats, want 3", len(store.assignSeatArgs))
	}
	seats := map[pgtype.UUID]db.AssignParticipantSeatParams{}
	for _, arg := range store.assignSeatArgs {
		seats[arg.ID] = arg
	}
	if got := seats[anaID]; got.RoomID != roomID || got.SeatNo.Int32 != 1 {
		t.Fatalf("Ana seat = %+v, want room one seat 1", got)
	}
	if got := seats[budiID]; got.RoomID != roomID || got.SeatNo.Int32 != 2 {
		t.Fatalf("Budi seat = %+v, want room one seat 2", got)
	}
	if got := seats[ciciID]; got.RoomID != otherRoomID || got.SeatNo.Int32 != 1 {
		t.Fatalf("Cici seat = %+v, want other room seat 1", got)
	}
}

func TestCbtSessionListMethodsPropagateErrors(t *testing.T) {
	sessionID := documentCycleTestUUID(231)
	participantID := documentCycleTestUUID(232)
	teacherID := documentCycleTestUUID(233)

	svc := &CbtSession{q: &fakeCbtSessionStore{listErr: errors.New("sessions failed")}}
	if _, err := svc.List(context.Background()); err == nil || err.Error() != "sessions failed" {
		t.Fatalf("List(error) = %v, want sessions failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{participantsErr: errors.New("participants failed")}}
	if _, err := svc.ListParticipants(context.Background(), sessionID); err == nil || err.Error() != "participants failed" {
		t.Fatalf("ListParticipants(error) = %v, want participants failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{roomsErr: errors.New("rooms failed")}}
	if _, err := svc.ListRooms(context.Background(), sessionID); err == nil || err.Error() != "rooms failed" {
		t.Fatalf("ListRooms(error) = %v, want rooms failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{byRoomErr: errors.New("by-room failed")}}
	if err := svc.AutoAssignSeats(context.Background(), sessionID); err == nil || err.Error() != "by-room failed" {
		t.Fatalf("AutoAssignSeats(error) = %v, want by-room failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{proctorErr: errors.New("proctor failed")}}
	if _, err := svc.GetProctoringStatus(context.Background(), sessionID); err == nil || err.Error() != "proctor failed" {
		t.Fatalf("GetProctoringStatus(error) = %v, want proctor failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{proctorRoomErr: errors.New("proctor rooms failed")}}
	if _, err := svc.ListProctorRooms(context.Background(), teacherID, false); err == nil || err.Error() != "proctor rooms failed" {
		t.Fatalf("ListProctorRooms(error) = %v, want proctor rooms failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{ungradedErr: errors.New("ungraded failed")}}
	if _, err := svc.ListUngradedEssays(context.Background(), sessionID); err == nil || err.Error() != "ungraded failed" {
		t.Fatalf("ListUngradedEssays(error) = %v, want ungraded failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{teacherRowsErr: errors.New("teacher failed")}}
	if _, err := svc.ListByTeacher(context.Background(), teacherID); err == nil || err.Error() != "teacher failed" {
		t.Fatalf("ListByTeacher(error) = %v, want teacher failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{resultsTeacherErr: errors.New("teacher results failed")}}
	if _, err := svc.GetResultsByTeacher(context.Background(), sessionID, teacherID); err == nil || err.Error() != "teacher results failed" {
		t.Fatalf("GetResultsByTeacher(error) = %v, want teacher results failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{resultsErr: errors.New("results failed")}}
	if _, err := svc.GetResults(context.Background(), sessionID); err == nil || err.Error() != "results failed" {
		t.Fatalf("GetResults(error) = %v, want results failed", err)
	}
	svc = &CbtSession{q: &fakeCbtSessionStore{answersErr: errors.New("answers failed")}}
	if _, err := svc.GetParticipantAnswers(context.Background(), participantID); err == nil || err.Error() != "answers failed" {
		t.Fatalf("GetParticipantAnswers(error) = %v, want answers failed", err)
	}
}
