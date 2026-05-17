package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestRombelConstructorsAndSimpleWrappers(t *testing.T) {
	if NewRombel(nil).q == nil {
		t.Fatal("NewRombel(nil).q = nil")
	}
	withPool := NewRombelWithPool(nil)
	if withPool.q == nil || withPool.tx != nil {
		t.Fatalf("NewRombelWithPool(nil) = %+v, want query store and nil tx", withPool)
	}

	classID := documentCycleTestUUID(211)
	assignmentID := documentCycleTestUUID(212)
	store := &fakeRombelWrapperStore{
		rombels: []db.ListRombelsRow{{ID: classID, Code: "VII.A"}},
		detail:  db.GetRombelDetailRow{ID: classID, Code: "VII.A"},
		assignments: []db.ListRombelSubjectAssignmentsRow{{
			ID: assignmentID,
		}},
		assignment:     db.GetRombelSubjectAssignmentRow{ID: assignmentID, ClassID: classID},
		timetableSlots: []db.ListRombelTimetableSlotsRow{{ID: documentCycleTestUUID(213), AssignmentID: assignmentID}},
		timetableSlot:  db.GetRombelTimetableSlotRow{ID: documentCycleTestUUID(214), AssignmentID: assignmentID},
	}
	svc := &Rombel{q: store}

	if rows, err := svc.List(context.Background()); err != nil || len(rows) != 1 || rows[0].Code != "VII.A" {
		t.Fatalf("List() = %+v, %v", rows, err)
	}
	if row, err := svc.Get(context.Background(), classID); err != nil || row.ID != classID {
		t.Fatalf("Get() = %+v, %v", row, err)
	}
	if rows, err := svc.ListSubjectAssignments(context.Background(), classID); err != nil || len(rows) != 1 || rows[0].ID != assignmentID {
		t.Fatalf("ListSubjectAssignments() = %+v, %v", rows, err)
	}
	if row, err := svc.GetSubjectAssignment(context.Background(), db.GetRombelSubjectAssignmentParams{ClassID: classID, ID: assignmentID}); err != nil || row.ID != assignmentID {
		t.Fatalf("GetSubjectAssignment() = %+v, %v", row, err)
	}
	if rows, err := svc.ListTimetableSlots(context.Background(), classID); err != nil || len(rows) != 1 || rows[0].AssignmentID != assignmentID {
		t.Fatalf("ListTimetableSlots() = %+v, %v", rows, err)
	}
	if row, err := svc.GetTimetableSlot(context.Background(), db.GetRombelTimetableSlotParams{ClassID: classID, ID: store.timetableSlot.ID}); err != nil || row.ID != store.timetableSlot.ID {
		t.Fatalf("GetTimetableSlot() = %+v, %v", row, err)
	}
}

func TestRombelSubjectAssignmentMatrixAndCellHelpers(t *testing.T) {
	yearID := documentCycleTestUUID(221)
	classID := documentCycleTestUUID(222)
	subjectID := documentCycleTestUUID(223)
	teacherID := documentCycleTestUUID(224)
	assignmentID := documentCycleTestUUID(225)
	store := &fakeRombelWrapperStore{
		activeYear:       db.AcademicYear{ID: yearID, Name: "2025/2026"},
		classes:          []db.ListSubjectAssignmentMatrixClassesRow{{ID: classID, Name: "VII A"}},
		subjects:         []db.ListSubjectAssignmentMatrixSubjectsRow{{ID: subjectID, Name: "Prakarya", IsChoiceSubject: true}},
		teachers:         []db.ListSubjectAssignmentMatrixTeachersRow{{ID: teacherID, Nama: "Guru"}},
		cells:            []db.ListSubjectAssignmentMatrixCellsRow{{ClassID: classID, SubjectID: subjectID}},
		upsertAssignment: db.ClassSubjectAssignment{ID: assignmentID, ClassID: classID, SubjectID: subjectID, TeacherEmployeeID: teacherID},
		matrixCell:       db.GetSubjectAssignmentMatrixCellRow{ClassID: classID, SubjectID: subjectID, AssignmentID: assignmentID},
	}
	svc := &Rombel{q: store}

	matrix, err := svc.GetSubjectAssignmentMatrix(context.Background())
	if err != nil {
		t.Fatalf("GetSubjectAssignmentMatrix() error = %v", err)
	}
	if matrix.AcademicYearID != yearID || matrix.AcademicYearName != "2025/2026" || len(matrix.Classes) != 1 || len(matrix.Subjects) != 1 || len(matrix.Teachers) != 1 || len(matrix.Cells) != 1 {
		t.Fatalf("matrix = %+v", matrix)
	}
	if choice, err := svc.isChoiceSubject(context.Background(), subjectID); err != nil || !choice {
		t.Fatalf("isChoiceSubject() = %v, %v; want true, nil", choice, err)
	}
	if choice, err := svc.isChoiceSubject(context.Background(), documentCycleTestUUID(226)); err != nil || choice {
		t.Fatalf("isChoiceSubject(unknown) = %v, %v; want false, nil", choice, err)
	}

	cell, err := svc.UpdateSubjectAssignmentMatrixCell(context.Background(), SubjectAssignmentMatrixCellInput{ClassID: classID, SubjectID: subjectID, TeacherEmployeeID: teacherID})
	if err != nil {
		t.Fatalf("UpdateSubjectAssignmentMatrixCell() error = %v", err)
	}
	if cell.AssignmentID != assignmentID || store.upsertCellArg.TeacherEmployeeID != teacherID {
		t.Fatalf("cell=%+v upsertArg=%+v", cell, store.upsertCellArg)
	}
}

func TestRombelSubjectAssignmentMatrixCellValidationAndDeletion(t *testing.T) {
	classID := documentCycleTestUUID(231)
	subjectID := documentCycleTestUUID(232)
	assignmentID := documentCycleTestUUID(233)
	svc := &Rombel{q: &fakeRombelWrapperStore{}}
	if _, err := svc.UpdateSubjectAssignmentMatrixCell(context.Background(), SubjectAssignmentMatrixCellInput{}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("missing IDs error = %v, want bad request", err)
	}
	negative := -1.0
	if _, err := svc.UpdateSubjectAssignmentMatrixCell(context.Background(), SubjectAssignmentMatrixCellInput{ClassID: classID, SubjectID: subjectID, AdditionalWeeklyHours: &negative}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("negative hours error = %v, want bad request", err)
	}

	current := db.ClassSubjectAssignment{ID: assignmentID, ClassID: classID, SubjectID: subjectID}
	store := &fakeRombelWrapperStore{byClassSubject: current, matrixCell: db.GetSubjectAssignmentMatrixCellRow{ClassID: classID, SubjectID: subjectID, AssignmentID: assignmentID}, deleteAssignmentRows: 1}
	svc = &Rombel{q: store}
	if _, err := svc.UpdateSubjectAssignmentMatrixCell(context.Background(), SubjectAssignmentMatrixCellInput{ClassID: classID, SubjectID: subjectID}); err != nil {
		t.Fatalf("clearing assignment error = %v", err)
	}
	if !store.deleteAssignmentCalled || store.deleteAssignmentArg.ID != assignmentID {
		t.Fatalf("assignment not deleted: called=%v arg=%+v", store.deleteAssignmentCalled, store.deleteAssignmentArg)
	}
	positive := 1.0
	store = &fakeRombelWrapperStore{byClassSubject: current}
	svc = &Rombel{q: store}
	if _, err := svc.UpdateSubjectAssignmentMatrixCell(context.Background(), SubjectAssignmentMatrixCellInput{ClassID: classID, SubjectID: subjectID, AdditionalWeeklyHours: &positive}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("positive hours without teacher error = %v, want bad request", err)
	}
}

func TestRombelNumericAndUUIDHelpers(t *testing.T) {
	n := float64ToNumeric(1.236)
	if got := numericToFloat64(n); got != 1.24 {
		t.Fatalf("float64ToNumeric(1.236) as float = %v, want 1.24", got)
	}
	a := documentCycleTestUUID(241)
	b := a
	invalid := pgtype.UUID{}
	if !sameRombelUUID(a, b) {
		t.Fatal("sameRombelUUID(equal valid) = false")
	}
	if sameRombelUUID(a, invalid) {
		t.Fatal("sameRombelUUID(valid, invalid) = true")
	}
	if !sameRombelUUID(invalid, pgtype.UUID{}) {
		t.Fatal("sameRombelUUID(two invalid UUIDs) = false")
	}
}

func TestRombelSubjectAssignmentUpdateAndDeleteRules(t *testing.T) {
	classID := documentCycleTestUUID(251)
	assignmentID := documentCycleTestUUID(252)
	subjectID := documentCycleTestUUID(253)
	newSubjectID := documentCycleTestUUID(254)
	store := &fakeRombelWrapperStore{assignment: db.GetRombelSubjectAssignmentRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID}}
	svc := &Rombel{q: store}
	if _, err := svc.UpdateSubjectAssignment(context.Background(), db.UpdateRombelSubjectAssignmentParams{ClassID: classID, ID: assignmentID, SubjectID: subjectID}); err != nil {
		t.Fatalf("UpdateSubjectAssignment(same subject) error = %v", err)
	}
	if store.countDependentsCalled {
		t.Fatal("CountRombelSubjectAssignmentDependents called for same subject")
	}

	store = &fakeRombelWrapperStore{
		assignment: db.GetRombelSubjectAssignmentRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID},
		dependents: db.CountRombelSubjectAssignmentDependentsRow{TotalTimetableSlots: 1},
	}
	svc = &Rombel{q: store}
	if _, err := svc.UpdateSubjectAssignment(context.Background(), db.UpdateRombelSubjectAssignmentParams{ClassID: classID, ID: assignmentID, SubjectID: newSubjectID}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("UpdateSubjectAssignment(changed used subject) error = %v, want conflict", err)
	}
	if err := svc.DeleteSubjectAssignment(context.Background(), db.DeleteRombelSubjectAssignmentParams{ClassID: classID, ID: assignmentID}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("DeleteSubjectAssignment(used) error = %v, want conflict", err)
	}

	store = &fakeRombelWrapperStore{deleteAssignmentRows: 0}
	svc = &Rombel{q: store}
	if err := svc.DeleteSubjectAssignment(context.Background(), db.DeleteRombelSubjectAssignmentParams{ClassID: classID, ID: assignmentID}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeleteSubjectAssignment(no rows) error = %v, want not found", err)
	}
}

func TestRombelTimetableCreateUpdateDeleteRules(t *testing.T) {
	classID := documentCycleTestUUID(161)
	assignmentID := documentCycleTestUUID(162)
	teacherID := documentCycleTestUUID(163)
	slotID := documentCycleTestUUID(164)
	start := pgtype.Time{Microseconds: 8 * 60 * 60 * 1_000_000, Valid: true}
	end := pgtype.Time{Microseconds: 9 * 60 * 60 * 1_000_000, Valid: true}
	store := &fakeRombelWrapperStore{
		assignment:          db.GetRombelSubjectAssignmentRow{ID: assignmentID, ClassID: classID, TeacherEmployeeID: teacherID},
		createTimetableRow:  db.CreateRombelTimetableSlotRow{ID: slotID, AssignmentID: assignmentID},
		updateTimetableRow:  db.UpdateRombelTimetableSlotRow{ID: slotID, AssignmentID: assignmentID},
		timetableSlot:       db.GetRombelTimetableSlotRow{ID: slotID, AssignmentID: assignmentID},
		deleteTimetableRows: 1,
	}
	svc := &Rombel{q: store}
	created, err := svc.CreateTimetableSlot(context.Background(), db.CreateRombelTimetableSlotParams{ClassID: classID, AssignmentID: assignmentID, DayOfWeek: 1, StartTime: start, EndTime: end, RoomLabel: " Lab "})
	if err != nil || created.ID != slotID {
		t.Fatalf("CreateTimetableSlot() = %+v, %v", created, err)
	}
	if len(store.lockKeys) != 3 || store.conflictArg.ClassID != classID || store.roomConflictArg.RoomLabel != "Lab" {
		t.Fatalf("locks/conflict args not set: locks=%v conflict=%+v room=%+v", store.lockKeys, store.conflictArg, store.roomConflictArg)
	}
	if _, err := svc.UpdateTimetableSlot(context.Background(), db.UpdateRombelTimetableSlotParams{ClassID: classID, ID: slotID, AssignmentID: assignmentID, DayOfWeek: 2, StartTime: start, EndTime: end}); err != nil {
		t.Fatalf("UpdateTimetableSlot() error = %v", err)
	}
	if err := svc.DeleteTimetableSlot(context.Background(), db.DeleteRombelTimetableSlotParams{ClassID: classID, ID: slotID}); err != nil {
		t.Fatalf("DeleteTimetableSlot() error = %v", err)
	}

	store = &fakeRombelWrapperStore{assignment: db.GetRombelSubjectAssignmentRow{ID: assignmentID, ClassID: classID, TeacherEmployeeID: teacherID}, timetableConflicts: 1}
	svc = &Rombel{q: store}
	if _, err := svc.CreateTimetableSlot(context.Background(), db.CreateRombelTimetableSlotParams{ClassID: classID, AssignmentID: assignmentID, DayOfWeek: 1, StartTime: start, EndTime: end}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("CreateTimetableSlot(conflict) error = %v, want conflict", err)
	}
	store = &fakeRombelWrapperStore{deleteTimetableRows: 0}
	svc = &Rombel{q: store}
	if err := svc.DeleteTimetableSlot(context.Background(), db.DeleteRombelTimetableSlotParams{ClassID: classID, ID: slotID}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeleteTimetableSlot(no rows) error = %v, want not found", err)
	}
}

type fakeRombelWrapperStore struct {
	noopRombelStore
	rombels                []db.ListRombelsRow
	detail                 db.GetRombelDetailRow
	assignments            []db.ListRombelSubjectAssignmentsRow
	assignment             db.GetRombelSubjectAssignmentRow
	activeYear             db.AcademicYear
	activeYearErr          error
	classes                []db.ListSubjectAssignmentMatrixClassesRow
	subjects               []db.ListSubjectAssignmentMatrixSubjectsRow
	teachers               []db.ListSubjectAssignmentMatrixTeachersRow
	cells                  []db.ListSubjectAssignmentMatrixCellsRow
	matrixCell             db.GetSubjectAssignmentMatrixCellRow
	byClassSubject         db.ClassSubjectAssignment
	byClassSubjectErr      error
	upsertAssignment       db.ClassSubjectAssignment
	upsertCellArg          db.UpsertSubjectAssignmentMatrixCellParams
	deleteAssignmentCalled bool
	deleteAssignmentArg    db.DeleteRombelSubjectAssignmentParams
	deleteAssignmentRows   int64
	dependents             db.CountRombelSubjectAssignmentDependentsRow
	countDependentsCalled  bool
	timetableSlots         []db.ListRombelTimetableSlotsRow
	timetableSlot          db.GetRombelTimetableSlotRow
	createTimetableRow     db.CreateRombelTimetableSlotRow
	updateTimetableRow     db.UpdateRombelTimetableSlotRow
	deleteTimetableRows    int64
	lockKeys               []string
	timetableConflicts     int32
	roomConflicts          int32
	conflictArg            db.CountTimetableConflictsParams
	roomConflictArg        db.CountTimetableRoomConflictsParams
}

func (f *fakeRombelWrapperStore) ListRombels(context.Context) ([]db.ListRombelsRow, error) {
	return f.rombels, nil
}
func (f *fakeRombelWrapperStore) GetRombelDetail(context.Context, pgtype.UUID) (db.GetRombelDetailRow, error) {
	return f.detail, nil
}
func (f *fakeRombelWrapperStore) ListRombelSubjectAssignments(context.Context, pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error) {
	return f.assignments, nil
}
func (f *fakeRombelWrapperStore) GetRombelSubjectAssignment(context.Context, db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error) {
	return f.assignment, nil
}
func (f *fakeRombelWrapperStore) GetActiveAcademicYear(context.Context) (db.AcademicYear, error) {
	if f.activeYearErr != nil {
		return db.AcademicYear{}, f.activeYearErr
	}
	return f.activeYear, nil
}
func (f *fakeRombelWrapperStore) ListSubjectAssignmentMatrixClasses(context.Context, pgtype.UUID) ([]db.ListSubjectAssignmentMatrixClassesRow, error) {
	return f.classes, nil
}
func (f *fakeRombelWrapperStore) ListSubjectAssignmentMatrixSubjects(context.Context) ([]db.ListSubjectAssignmentMatrixSubjectsRow, error) {
	return f.subjects, nil
}
func (f *fakeRombelWrapperStore) ListSubjectAssignmentMatrixTeachers(context.Context) ([]db.ListSubjectAssignmentMatrixTeachersRow, error) {
	return f.teachers, nil
}
func (f *fakeRombelWrapperStore) ListSubjectAssignmentMatrixCells(context.Context, pgtype.UUID) ([]db.ListSubjectAssignmentMatrixCellsRow, error) {
	return f.cells, nil
}
func (f *fakeRombelWrapperStore) GetSubjectAssignmentMatrixCell(context.Context, db.GetSubjectAssignmentMatrixCellParams) (db.GetSubjectAssignmentMatrixCellRow, error) {
	return f.matrixCell, nil
}
func (f *fakeRombelWrapperStore) GetSubjectAssignmentByClassSubject(context.Context, db.GetSubjectAssignmentByClassSubjectParams) (db.ClassSubjectAssignment, error) {
	if f.byClassSubjectErr != nil {
		return db.ClassSubjectAssignment{}, f.byClassSubjectErr
	}
	if !f.byClassSubject.ID.Valid {
		return db.ClassSubjectAssignment{}, pgx.ErrNoRows
	}
	return f.byClassSubject, nil
}
func (f *fakeRombelWrapperStore) UpsertSubjectAssignmentMatrixCell(_ context.Context, arg db.UpsertSubjectAssignmentMatrixCellParams) (db.ClassSubjectAssignment, error) {
	f.upsertCellArg = arg
	return f.upsertAssignment, nil
}
func (f *fakeRombelWrapperStore) CountRombelSubjectAssignmentDependents(context.Context, db.CountRombelSubjectAssignmentDependentsParams) (db.CountRombelSubjectAssignmentDependentsRow, error) {
	f.countDependentsCalled = true
	return f.dependents, nil
}
func (f *fakeRombelWrapperStore) DeleteRombelSubjectAssignment(_ context.Context, arg db.DeleteRombelSubjectAssignmentParams) (int64, error) {
	f.deleteAssignmentCalled = true
	f.deleteAssignmentArg = arg
	return f.deleteAssignmentRows, nil
}
func (f *fakeRombelWrapperStore) UpdateRombelSubjectAssignment(context.Context, db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error) {
	return db.UpdateRombelSubjectAssignmentRow{}, nil
}
func (f *fakeRombelWrapperStore) ListRombelTimetableSlots(context.Context, pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error) {
	return f.timetableSlots, nil
}
func (f *fakeRombelWrapperStore) GetRombelTimetableSlot(context.Context, db.GetRombelTimetableSlotParams) (db.GetRombelTimetableSlotRow, error) {
	return f.timetableSlot, nil
}
func (f *fakeRombelWrapperStore) CreateRombelTimetableSlot(context.Context, db.CreateRombelTimetableSlotParams) (db.CreateRombelTimetableSlotRow, error) {
	return f.createTimetableRow, nil
}
func (f *fakeRombelWrapperStore) UpdateRombelTimetableSlot(context.Context, db.UpdateRombelTimetableSlotParams) (db.UpdateRombelTimetableSlotRow, error) {
	return f.updateTimetableRow, nil
}
func (f *fakeRombelWrapperStore) DeleteRombelTimetableSlot(context.Context, db.DeleteRombelTimetableSlotParams) (int64, error) {
	return f.deleteTimetableRows, nil
}
func (f *fakeRombelWrapperStore) LockTimetableMutationScope(_ context.Context, key string) (int64, error) {
	f.lockKeys = append(f.lockKeys, key)
	return 1, nil
}
func (f *fakeRombelWrapperStore) CountTimetableConflicts(_ context.Context, arg db.CountTimetableConflictsParams) (int32, error) {
	f.conflictArg = arg
	return f.timetableConflicts, nil
}
func (f *fakeRombelWrapperStore) CountTimetableRoomConflicts(_ context.Context, arg db.CountTimetableRoomConflictsParams) (int32, error) {
	f.roomConflictArg = arg
	return f.roomConflicts, nil
}

var _ = errors.Is
