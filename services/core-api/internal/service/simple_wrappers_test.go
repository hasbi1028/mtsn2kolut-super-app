package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeParentStore struct {
	listRows    []db.Parent
	getID       pgtype.UUID
	createArg   db.CreateParentParams
	updateArg   db.UpdateParentParams
	deleteID    pgtype.UUID
	linkArg     db.LinkParentStudentParams
	unlinkArg   db.UnlinkParentStudentParams
	childrenID  pgtype.UUID
	childrenRow []db.ListParentChildrenRow
}

func (f *fakeParentStore) ListParents(ctx context.Context) ([]db.Parent, error) {
	return f.listRows, nil
}

func (f *fakeParentStore) GetParent(ctx context.Context, id pgtype.UUID) (db.Parent, error) {
	f.getID = id
	return db.Parent{ID: id, Nama: "Wali"}, nil
}

func (f *fakeParentStore) CreateParent(ctx context.Context, arg db.CreateParentParams) (db.Parent, error) {
	f.createArg = arg
	return db.Parent{Nama: arg.Nama, Phone: arg.Phone, Address: arg.Address}, nil
}

func (f *fakeParentStore) UpdateParent(ctx context.Context, arg db.UpdateParentParams) (db.Parent, error) {
	f.updateArg = arg
	return db.Parent{ID: arg.ID, Nama: arg.Nama, Phone: arg.Phone, Address: arg.Address}, nil
}

func (f *fakeParentStore) DeleteParent(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return nil
}

func (f *fakeParentStore) LinkParentStudent(ctx context.Context, arg db.LinkParentStudentParams) error {
	f.linkArg = arg
	return nil
}

func (f *fakeParentStore) UnlinkParentStudent(ctx context.Context, arg db.UnlinkParentStudentParams) error {
	f.unlinkArg = arg
	return nil
}

func (f *fakeParentStore) ListParentChildren(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	f.childrenID = parentID
	return f.childrenRow, nil
}

func TestParentServiceForwardsStoreCalls(t *testing.T) {
	parentID := documentCycleTestUUID(101)
	studentID := documentCycleTestUUID(102)
	store := &fakeParentStore{
		listRows:    []db.Parent{{ID: parentID, Nama: "Wali"}},
		childrenRow: []db.ListParentChildrenRow{{ID: studentID, Nama: "Siswa"}},
	}
	svc := &Parent{q: store}

	rows, err := svc.List(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("List() = %d rows, %v; want 1 nil", len(rows), err)
	}
	got, err := svc.Get(context.Background(), parentID)
	if err != nil || got.ID != parentID || store.getID != parentID {
		t.Fatalf("Get() = %+v, %v; getID=%v, want %v", got, err, store.getID, parentID)
	}
	if _, err := svc.Create(context.Background(), "Wali", "0812", "Lasusua"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if store.createArg.Nama != "Wali" || store.createArg.Phone != "0812" || store.createArg.Address != "Lasusua" {
		t.Fatalf("Create() arg = %+v, want forwarded fields", store.createArg)
	}
	if _, err := svc.Update(context.Background(), parentID, "Wali Baru", "0821", "Pakue"); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if store.updateArg.ID != parentID || store.updateArg.Nama != "Wali Baru" {
		t.Fatalf("Update() arg = %+v, want forwarded id/name", store.updateArg)
	}
	if err := svc.LinkStudent(context.Background(), parentID, studentID); err != nil {
		t.Fatalf("LinkStudent() error = %v", err)
	}
	if store.linkArg.ParentID != parentID || store.linkArg.StudentID != studentID {
		t.Fatalf("LinkStudent() arg = %+v, want forwarded ids", store.linkArg)
	}
	if err := svc.UnlinkStudent(context.Background(), parentID, studentID); err != nil {
		t.Fatalf("UnlinkStudent() error = %v", err)
	}
	if store.unlinkArg.ParentID != parentID || store.unlinkArg.StudentID != studentID {
		t.Fatalf("UnlinkStudent() arg = %+v, want forwarded ids", store.unlinkArg)
	}
	children, err := svc.ListChildren(context.Background(), parentID)
	if err != nil || len(children) != 1 || store.childrenID != parentID {
		t.Fatalf("ListChildren() = %d rows, %v; parentID=%v, want 1 nil/%v", len(children), err, store.childrenID, parentID)
	}
	if err := svc.Delete(context.Background(), parentID); err != nil || store.deleteID != parentID {
		t.Fatalf("Delete() = %v, id=%v; want nil/%v", err, store.deleteID, parentID)
	}
}

type fakeStudentStore struct {
	listRows        []db.ListStudentsRow
	listTeacherID   pgtype.UUID
	listTeacherRows []db.ListStudentsByTeacherRow
	createArg       db.CreateStudentParams
	updateArg       db.UpdateStudentParams
	deleteID        pgtype.UUID
	statusArg       db.UpdateStudentStatusParams
	getID           pgtype.UUID
	lifecycleArg    db.UpdateStudentLifecycleParams
	usersByStudent  []db.ListUsersByStudentIDRow
	usersErr        error
	userStatusArgs  []db.UpdateUserStatusParams
	userStatusErr   error
}

func (f *fakeStudentStore) ListStudents(ctx context.Context) ([]db.ListStudentsRow, error) {
	return f.listRows, nil
}

func (f *fakeStudentStore) ListStudentsByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListStudentsByTeacherRow, error) {
	f.listTeacherID = teacherEmployeeID
	return f.listTeacherRows, nil
}

func (f *fakeStudentStore) CreateStudent(ctx context.Context, arg db.CreateStudentParams) (db.Student, error) {
	f.createArg = arg
	return db.Student{Nis: arg.Nis, Nama: arg.Nama, ClassID: arg.ClassID, Status: arg.Status}, nil
}

func (f *fakeStudentStore) UpdateStudent(ctx context.Context, arg db.UpdateStudentParams) (db.Student, error) {
	f.updateArg = arg
	return db.Student{ID: arg.ID, Nis: arg.Nis, Nama: arg.Nama, ClassID: arg.ClassID, Status: arg.Status}, nil
}

func (f *fakeStudentStore) DeleteStudent(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return nil
}

func (f *fakeStudentStore) UpdateStudentStatus(ctx context.Context, arg db.UpdateStudentStatusParams) error {
	f.statusArg = arg
	return nil
}

func (f *fakeStudentStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (db.GetStudentByIDRow, error) {
	f.getID = id
	return db.GetStudentByIDRow{ID: id, Nama: "Siswa"}, nil
}

func (f *fakeStudentStore) UpdateStudentLifecycle(ctx context.Context, arg db.UpdateStudentLifecycleParams) error {
	f.lifecycleArg = arg
	return nil
}

func (f *fakeStudentStore) ListUsersByStudentID(ctx context.Context, studentID pgtype.UUID) ([]db.ListUsersByStudentIDRow, error) {
	return f.usersByStudent, f.usersErr
}

func (f *fakeStudentStore) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) error {
	f.userStatusArgs = append(f.userStatusArgs, arg)
	return f.userStatusErr
}

func TestStudentServiceForwardsStoreCallsAndLifecycleStatus(t *testing.T) {
	studentID := documentCycleTestUUID(111)
	teacherID := documentCycleTestUUID(112)
	userOneID := documentCycleTestUUID(113)
	userTwoID := documentCycleTestUUID(114)
	classID := documentCycleTestUUID(115)
	store := &fakeStudentStore{
		listRows:        []db.ListStudentsRow{{ID: studentID, Nama: "Siswa"}},
		listTeacherRows: []db.ListStudentsByTeacherRow{{ID: studentID, Nama: "Siswa"}},
		usersByStudent: []db.ListUsersByStudentIDRow{
			{ID: userOneID},
			{ID: userTwoID},
		},
	}
	svc := &Student{q: store}

	rows, err := svc.List(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("List() = %d rows, %v; want 1 nil", len(rows), err)
	}
	teacherRows, err := svc.ListByTeacher(context.Background(), teacherID)
	if err != nil || len(teacherRows) != 1 || store.listTeacherID != teacherID {
		t.Fatalf("ListByTeacher() = %d rows, %v; teacherID=%v, want 1 nil/%v", len(teacherRows), err, store.listTeacherID, teacherID)
	}
	if _, err := svc.Create(context.Background(), db.CreateStudentParams{Nis: "001", Nama: "Siswa", ClassID: classID, Status: db.StudentStatusEnumActive}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if store.createArg.Nis != "001" || store.createArg.ClassID != classID {
		t.Fatalf("Create() arg = %+v, want forwarded student", store.createArg)
	}
	if _, err := svc.Update(context.Background(), db.UpdateStudentParams{ID: studentID, Nis: "002", Nama: "Siswa B", ClassID: classID, Status: db.StudentStatusEnumAlumni}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if store.updateArg.ID != studentID || store.updateArg.Nis != "002" {
		t.Fatalf("Update() arg = %+v, want forwarded update", store.updateArg)
	}
	if err := svc.UpdateStatus(context.Background(), studentID, db.StudentStatusEnumProspective); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if store.statusArg.ID != studentID || store.statusArg.Status != db.StudentStatusEnumProspective {
		t.Fatalf("UpdateStatus() arg = %+v, want prospective status", store.statusArg)
	}
	if got, err := svc.Get(context.Background(), studentID); err != nil || got.ID != studentID || store.getID != studentID {
		t.Fatalf("Get() = %+v, %v; getID=%v, want %v", got, err, store.getID, studentID)
	}

	if err := svc.UpdateLifecycle(context.Background(), studentID, db.StudentStatusEnumActive); err != nil {
		t.Fatalf("UpdateLifecycle(active) error = %v", err)
	}
	if store.lifecycleArg.Status != db.StudentStatusEnumActive || !store.lifecycleArg.IsActive {
		t.Fatalf("UpdateLifecycle(active) arg = %+v, want active true", store.lifecycleArg)
	}
	if len(store.userStatusArgs) != 2 || !store.userStatusArgs[0].IsActive || store.userStatusArgs[1].ID != userTwoID {
		t.Fatalf("UpdateLifecycle(active) user status args = %+v, want two active updates", store.userStatusArgs)
	}

	store.userStatusArgs = nil
	if err := svc.UpdateLifecycle(context.Background(), studentID, db.StudentStatusEnumMutated); err != nil {
		t.Fatalf("UpdateLifecycle(mutated) error = %v", err)
	}
	if store.lifecycleArg.Status != db.StudentStatusEnumMutated || store.lifecycleArg.IsActive {
		t.Fatalf("UpdateLifecycle(mutated) arg = %+v, want inactive mutated", store.lifecycleArg)
	}
	if len(store.userStatusArgs) != 2 || store.userStatusArgs[0].IsActive {
		t.Fatalf("UpdateLifecycle(mutated) user status args = %+v, want inactive updates", store.userStatusArgs)
	}

	store.usersErr = errors.New("users failed")
	if err := svc.UpdateLifecycle(context.Background(), studentID, db.StudentStatusEnumActive); err == nil || err.Error() != "users failed" {
		t.Fatalf("UpdateLifecycle(users error) = %v, want users failed", err)
	}
	store.usersErr = nil
	store.userStatusErr = errors.New("status failed")
	if err := svc.UpdateLifecycle(context.Background(), studentID, db.StudentStatusEnumActive); err == nil || err.Error() != "status failed" {
		t.Fatalf("UpdateLifecycle(user status error) = %v, want status failed", err)
	}
	if err := svc.Delete(context.Background(), studentID); err != nil || store.deleteID != studentID {
		t.Fatalf("Delete() = %v, id=%v; want nil/%v", err, store.deleteID, studentID)
	}
}

type fakePusakaScheduleStore struct {
	listRows  []db.Schedule
	getID     pgtype.UUID
	createArg db.CreateScheduleParams
	updateArg db.UpdateScheduleByIDParams
	deleteID  pgtype.UUID
}

func (f *fakePusakaScheduleStore) ListSchedules(ctx context.Context) ([]db.Schedule, error) {
	return f.listRows, nil
}

func (f *fakePusakaScheduleStore) GetSchedule(ctx context.Context, id pgtype.UUID) (db.Schedule, error) {
	f.getID = id
	return db.Schedule{ID: id, Label: "Pagi"}, nil
}

func (f *fakePusakaScheduleStore) CreateSchedule(ctx context.Context, arg db.CreateScheduleParams) (db.Schedule, error) {
	f.createArg = arg
	return db.Schedule{Label: arg.Label, RunTime: arg.RunTime, RunType: arg.RunType, IsEnabled: arg.IsEnabled}, nil
}

func (f *fakePusakaScheduleStore) UpdateScheduleByID(ctx context.Context, arg db.UpdateScheduleByIDParams) (db.Schedule, error) {
	f.updateArg = arg
	return db.Schedule{ID: arg.ID, Label: arg.Label, RunTime: arg.RunTime, IsEnabled: arg.IsEnabled}, nil
}

func (f *fakePusakaScheduleStore) DeleteScheduleByID(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return nil
}

func TestPusakaScheduleServiceForwardsStoreCalls(t *testing.T) {
	scheduleID := documentCycleTestUUID(121)
	store := &fakePusakaScheduleStore{listRows: []db.Schedule{{ID: scheduleID, Label: "Pagi"}}}
	svc := &PusakaSchedule{q: store}

	rows, err := svc.List(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("List() = %d rows, %v; want 1 nil", len(rows), err)
	}
	got, err := svc.Get(context.Background(), scheduleID)
	if err != nil || got.ID != scheduleID || store.getID != scheduleID {
		t.Fatalf("Get() = %+v, %v; getID=%v, want %v", got, err, store.getID, scheduleID)
	}
	if _, err := svc.Create(context.Background(), db.CreateScheduleParams{Label: "Pagi", RunTime: "07:00", RunType: db.RunTypeEnumMorning, IsEnabled: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if store.createArg.Label != "Pagi" || store.createArg.RunType != db.RunTypeEnumMorning {
		t.Fatalf("Create() arg = %+v, want schedule fields", store.createArg)
	}
	if _, err := svc.UpdateByID(context.Background(), db.UpdateScheduleByIDParams{ID: scheduleID, Label: "Siang", RunTime: "12:00", IsEnabled: false}); err != nil {
		t.Fatalf("UpdateByID() error = %v", err)
	}
	if store.updateArg.ID != scheduleID || store.updateArg.Label != "Siang" || store.updateArg.IsEnabled {
		t.Fatalf("UpdateByID() arg = %+v, want updated schedule", store.updateArg)
	}
	if err := svc.DeleteByID(context.Background(), scheduleID); err != nil || store.deleteID != scheduleID {
		t.Fatalf("DeleteByID() = %v, id=%v; want nil/%v", err, store.deleteID, scheduleID)
	}
}

type fakePortalStore struct {
	studentID           pgtype.UUID
	parentID            pgtype.UUID
	parentChildrenID    pgtype.UUID
	studentParentsID    pgtype.UUID
	studentSessionsID   pgtype.UUID
	studentTimetableID  pgtype.UUID
	teacherTimetableID  pgtype.UUID
	childrenTimetableID pgtype.UUID
	studentErr          error
	parentErr           error
	childrenErr         error
	studentParentsErr   error
	sessionsErr         error
}

func (f *fakePortalStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (db.GetStudentByIDRow, error) {
	f.studentID = id
	if f.studentErr != nil {
		return db.GetStudentByIDRow{}, f.studentErr
	}
	return db.GetStudentByIDRow{ID: id, Nama: "Siswa"}, nil
}

func (f *fakePortalStore) ListStudentParents(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentParentsRow, error) {
	f.studentParentsID = studentID
	if f.studentParentsErr != nil {
		return nil, f.studentParentsErr
	}
	return []db.ListStudentParentsRow{{ID: documentCycleTestUUID(131), Nama: "Wali"}}, nil
}

func (f *fakePortalStore) ListStudentExamSessions(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error) {
	f.studentSessionsID = studentID
	if f.sessionsErr != nil {
		return nil, f.sessionsErr
	}
	return []db.ListStudentExamSessionsRow{{SessionTitle: "Ujian"}}, nil
}

func (f *fakePortalStore) ListStudentTimetable(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	f.studentTimetableID = studentID
	return []db.ListStudentTimetableRow{{SubjectName: "Matematika"}}, nil
}

func (f *fakePortalStore) ListTeacherTimetable(ctx context.Context, employeeID pgtype.UUID) ([]db.ListTeacherTimetableRow, error) {
	f.teacherTimetableID = employeeID
	return []db.ListTeacherTimetableRow{{SubjectName: "IPA"}}, nil
}

func (f *fakePortalStore) ListParentChildrenTimetable(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenTimetableRow, error) {
	f.childrenTimetableID = parentID
	return []db.ListParentChildrenTimetableRow{{SubjectName: "IPS"}}, nil
}

func (f *fakePortalStore) GetParent(ctx context.Context, id pgtype.UUID) (db.Parent, error) {
	f.parentID = id
	if f.parentErr != nil {
		return db.Parent{}, f.parentErr
	}
	return db.Parent{ID: id, Nama: "Wali"}, nil
}

func (f *fakePortalStore) ListParentChildren(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	f.parentChildrenID = parentID
	if f.childrenErr != nil {
		return nil, f.childrenErr
	}
	return []db.ListParentChildrenRow{{ID: documentCycleTestUUID(132), Nama: "Siswa"}}, nil
}

func TestPortalServiceComposesOverviewAndTimetableCalls(t *testing.T) {
	studentID := documentCycleTestUUID(141)
	parentID := documentCycleTestUUID(142)
	teacherID := documentCycleTestUUID(143)
	store := &fakePortalStore{}
	svc := &Portal{q: store}

	student, parents, sessions, err := svc.StudentOverview(context.Background(), studentID)
	if err != nil {
		t.Fatalf("StudentOverview() error = %v", err)
	}
	if student.ID != studentID || len(parents) != 1 || len(sessions) != 1 || store.studentID != studentID || store.studentParentsID != studentID || store.studentSessionsID != studentID {
		t.Fatalf("StudentOverview() = %+v/%d/%d, store=%+v; want composed student overview", student, len(parents), len(sessions), store)
	}
	if rows, err := svc.StudentTimetable(context.Background(), studentID); err != nil || len(rows) != 1 || store.studentTimetableID != studentID {
		t.Fatalf("StudentTimetable() = %d rows, %v; id=%v, want 1 nil/%v", len(rows), err, store.studentTimetableID, studentID)
	}
	if rows, err := svc.TeacherTimetable(context.Background(), teacherID); err != nil || len(rows) != 1 || store.teacherTimetableID != teacherID {
		t.Fatalf("TeacherTimetable() = %d rows, %v; id=%v, want 1 nil/%v", len(rows), err, store.teacherTimetableID, teacherID)
	}
	if rows, err := svc.ParentChildrenTimetable(context.Background(), parentID); err != nil || len(rows) != 1 || store.childrenTimetableID != parentID {
		t.Fatalf("ParentChildrenTimetable() = %d rows, %v; id=%v, want 1 nil/%v", len(rows), err, store.childrenTimetableID, parentID)
	}
	parent, children, err := svc.ParentOverview(context.Background(), parentID)
	if err != nil {
		t.Fatalf("ParentOverview() error = %v", err)
	}
	if parent.ID != parentID || len(children) != 1 || store.parentID != parentID || store.parentChildrenID != parentID {
		t.Fatalf("ParentOverview() = %+v/%d, store=%+v; want composed parent overview", parent, len(children), store)
	}
}

func TestPortalServiceStopsOverviewOnStoreErrors(t *testing.T) {
	studentID := documentCycleTestUUID(151)
	parentID := documentCycleTestUUID(152)

	svc := &Portal{q: &fakePortalStore{studentErr: errors.New("student failed")}}
	if _, _, _, err := svc.StudentOverview(context.Background(), studentID); err == nil || err.Error() != "student failed" {
		t.Fatalf("StudentOverview(student error) = %v, want student failed", err)
	}
	svc = &Portal{q: &fakePortalStore{studentParentsErr: errors.New("parents failed")}}
	if _, _, _, err := svc.StudentOverview(context.Background(), studentID); err == nil || err.Error() != "parents failed" {
		t.Fatalf("StudentOverview(parents error) = %v, want parents failed", err)
	}
	svc = &Portal{q: &fakePortalStore{sessionsErr: errors.New("sessions failed")}}
	if _, _, _, err := svc.StudentOverview(context.Background(), studentID); err == nil || err.Error() != "sessions failed" {
		t.Fatalf("StudentOverview(sessions error) = %v, want sessions failed", err)
	}
	svc = &Portal{q: &fakePortalStore{parentErr: errors.New("parent failed")}}
	if _, _, err := svc.ParentOverview(context.Background(), parentID); err == nil || err.Error() != "parent failed" {
		t.Fatalf("ParentOverview(parent error) = %v, want parent failed", err)
	}
	svc = &Portal{q: &fakePortalStore{childrenErr: errors.New("children failed")}}
	if _, _, err := svc.ParentOverview(context.Background(), parentID); err == nil || err.Error() != "children failed" {
		t.Fatalf("ParentOverview(children error) = %v, want children failed", err)
	}
}

type fakeAcademicStore struct {
	years            []db.AcademicYear
	classes          []db.ListSchoolClassesRow
	subjects         []db.Subject
	assignments      []db.ListClassSubjectAssignmentsRow
	timetableSlots   []db.ListTimetableSlotsRow
	stats            db.GetAcademicStatsRow
	createYearArg    db.CreateAcademicYearParams
	createClassArg   db.CreateSchoolClassParams
	createSubjectArg db.CreateSubjectParams
	createAssignArg  db.CreateClassSubjectAssignmentParams
	createSlotArg    db.CreateTimetableSlotParams
	getSlotID        pgtype.UUID
	getSlotErr       error
	updateSlotArg    db.UpdateTimetableSlotParams
	deleteYearID     pgtype.UUID
	deleteClassID    pgtype.UUID
	deleteSubjectID  pgtype.UUID
	deleteAssignID   pgtype.UUID
	deleteSlotID     pgtype.UUID
	getAssignID      pgtype.UUID
	getAssignRow     db.GetClassSubjectAssignmentRow
	getAssignErr     error
	conflictArgs     []db.CountTimetableConflictsParams
	conflictCount    int32
	conflictErr      error
	roomArgs         []db.CountTimetableRoomConflictsParams
	roomCount        int32
	roomErr          error
}

func (f *fakeAcademicStore) ListAcademicYears(ctx context.Context) ([]db.AcademicYear, error) {
	return f.years, nil
}

func (f *fakeAcademicStore) ListSchoolClasses(ctx context.Context) ([]db.ListSchoolClassesRow, error) {
	return f.classes, nil
}

func (f *fakeAcademicStore) ListSubjects(ctx context.Context) ([]db.Subject, error) {
	return f.subjects, nil
}

func (f *fakeAcademicStore) ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error) {
	return f.assignments, nil
}

func (f *fakeAcademicStore) ListTimetableSlots(ctx context.Context) ([]db.ListTimetableSlotsRow, error) {
	return f.timetableSlots, nil
}

func (f *fakeAcademicStore) GetAcademicStats(ctx context.Context) (db.GetAcademicStatsRow, error) {
	return f.stats, nil
}

func (f *fakeAcademicStore) CreateAcademicYear(ctx context.Context, arg db.CreateAcademicYearParams) (db.AcademicYear, error) {
	f.createYearArg = arg
	return db.AcademicYear{Name: arg.Name, IsActive: arg.IsActive}, nil
}

func (f *fakeAcademicStore) CreateSchoolClass(ctx context.Context, arg db.CreateSchoolClassParams) (db.SchoolClass, error) {
	f.createClassArg = arg
	return db.SchoolClass{AcademicYearID: arg.AcademicYearID, Code: arg.Code, Name: arg.Name, Level: arg.Level, IsActive: arg.IsActive}, nil
}

func (f *fakeAcademicStore) CreateSubject(ctx context.Context, arg db.CreateSubjectParams) (db.Subject, error) {
	f.createSubjectArg = arg
	return db.Subject{Code: arg.Code, Name: arg.Name, IsActive: arg.IsActive}, nil
}

func (f *fakeAcademicStore) CreateClassSubjectAssignment(ctx context.Context, arg db.CreateClassSubjectAssignmentParams) (db.ClassSubjectAssignment, error) {
	f.createAssignArg = arg
	return db.ClassSubjectAssignment{ClassID: arg.ClassID, SubjectID: arg.SubjectID, TeacherEmployeeID: arg.TeacherEmployeeID}, nil
}

func (f *fakeAcademicStore) CreateTimetableSlot(ctx context.Context, arg db.CreateTimetableSlotParams) (db.TimetableSlot, error) {
	f.createSlotArg = arg
	return db.TimetableSlot{AssignmentID: arg.AssignmentID, DayOfWeek: arg.DayOfWeek, StartTime: arg.StartTime, EndTime: arg.EndTime, RoomLabel: arg.RoomLabel}, nil
}

func (f *fakeAcademicStore) GetTimetableSlot(ctx context.Context, id pgtype.UUID) (db.TimetableSlot, error) {
	f.getSlotID = id
	if f.getSlotErr != nil {
		return db.TimetableSlot{}, f.getSlotErr
	}
	return db.TimetableSlot{ID: id}, nil
}

func (f *fakeAcademicStore) UpdateTimetableSlot(ctx context.Context, arg db.UpdateTimetableSlotParams) (db.TimetableSlot, error) {
	f.updateSlotArg = arg
	return db.TimetableSlot{ID: arg.ID, AssignmentID: arg.AssignmentID, DayOfWeek: arg.DayOfWeek, StartTime: arg.StartTime, EndTime: arg.EndTime, RoomLabel: arg.RoomLabel}, nil
}

func (f *fakeAcademicStore) DeleteAcademicYear(ctx context.Context, id pgtype.UUID) error {
	f.deleteYearID = id
	return nil
}

func (f *fakeAcademicStore) DeleteSchoolClass(ctx context.Context, id pgtype.UUID) error {
	f.deleteClassID = id
	return nil
}

func (f *fakeAcademicStore) DeleteSubject(ctx context.Context, id pgtype.UUID) error {
	f.deleteSubjectID = id
	return nil
}

func (f *fakeAcademicStore) DeleteClassSubjectAssignment(ctx context.Context, id pgtype.UUID) error {
	f.deleteAssignID = id
	return nil
}

func (f *fakeAcademicStore) DeleteTimetableSlot(ctx context.Context, id pgtype.UUID) error {
	f.deleteSlotID = id
	return nil
}

func (f *fakeAcademicStore) GetClassSubjectAssignment(ctx context.Context, id pgtype.UUID) (db.GetClassSubjectAssignmentRow, error) {
	f.getAssignID = id
	if f.getAssignErr != nil {
		return db.GetClassSubjectAssignmentRow{}, f.getAssignErr
	}
	return f.getAssignRow, nil
}

func (f *fakeAcademicStore) CountTimetableConflicts(ctx context.Context, arg db.CountTimetableConflictsParams) (int32, error) {
	f.conflictArgs = append(f.conflictArgs, arg)
	return f.conflictCount, f.conflictErr
}

func (f *fakeAcademicStore) CountTimetableRoomConflicts(ctx context.Context, arg db.CountTimetableRoomConflictsParams) (int32, error) {
	f.roomArgs = append(f.roomArgs, arg)
	return f.roomCount, f.roomErr
}

func TestAcademicServiceForwardsStoreCallsAndChecksTimetableAvailability(t *testing.T) {
	yearID := documentCycleTestUUID(161)
	classID := documentCycleTestUUID(162)
	subjectID := documentCycleTestUUID(163)
	teacherID := documentCycleTestUUID(164)
	assignmentID := documentCycleTestUUID(165)
	slotID := documentCycleTestUUID(166)
	start, err := ParseAcademicTimeInput("07:30")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(start) error = %v", err)
	}
	end, err := ParseAcademicTimeInput("08:10")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(end) error = %v", err)
	}
	store := &fakeAcademicStore{
		years:          []db.AcademicYear{{ID: yearID, Name: "2026/2027"}},
		classes:        []db.ListSchoolClassesRow{{ID: classID, Name: "VII A"}},
		subjects:       []db.Subject{{ID: subjectID, Name: "Matematika"}},
		assignments:    []db.ListClassSubjectAssignmentsRow{{ID: assignmentID, ClassID: classID, SubjectID: subjectID, TeacherEmployeeID: teacherID}},
		timetableSlots: []db.ListTimetableSlotsRow{{ID: slotID, AssignmentID: assignmentID, RoomLabel: "R1"}},
		stats:          db.GetAcademicStatsRow{TotalStudents: 10, TotalClasses: 2, TotalSubjects: 3, TotalYears: 1},
		getAssignRow:   db.GetClassSubjectAssignmentRow{ID: assignmentID, ClassID: classID, TeacherEmployeeID: teacherID},
	}
	svc := &Academic{q: store}
	if NewAcademic(nil) == nil {
		t.Fatal("NewAcademic(nil) = nil, want service")
	}

	if rows, err := svc.ListYears(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("ListYears() = %d rows, %v; want 1 nil", len(rows), err)
	}
	if rows, err := svc.ListClasses(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("ListClasses() = %d rows, %v; want 1 nil", len(rows), err)
	}
	if rows, err := svc.ListSubjects(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("ListSubjects() = %d rows, %v; want 1 nil", len(rows), err)
	}
	if rows, err := svc.ListAssignments(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("ListAssignments() = %d rows, %v; want 1 nil", len(rows), err)
	}
	if rows, err := svc.ListTimetableSlots(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("ListTimetableSlots() = %d rows, %v; want 1 nil", len(rows), err)
	}
	if stats, err := svc.GetStats(context.Background()); err != nil || stats.TotalStudents != 10 {
		t.Fatalf("GetStats() = %+v, %v; want total students 10", stats, err)
	}

	if _, err := svc.CreateYear(context.Background(), db.CreateAcademicYearParams{Name: "2026/2027", IsActive: true}); err != nil {
		t.Fatalf("CreateYear() error = %v", err)
	}
	if store.createYearArg.Name != "2026/2027" || !store.createYearArg.IsActive {
		t.Fatalf("CreateYear() arg = %+v, want forwarded year", store.createYearArg)
	}
	if _, err := svc.CreateClass(context.Background(), db.CreateSchoolClassParams{AcademicYearID: yearID, Code: "7A", Name: "VII A", Level: "7", IsActive: true}); err != nil {
		t.Fatalf("CreateClass() error = %v", err)
	}
	if store.createClassArg.AcademicYearID != yearID || store.createClassArg.Code != "7A" {
		t.Fatalf("CreateClass() arg = %+v, want forwarded class", store.createClassArg)
	}
	if _, err := svc.CreateSubject(context.Background(), db.CreateSubjectParams{Code: "MTK", Name: "Matematika", IsActive: true}); err != nil {
		t.Fatalf("CreateSubject() error = %v", err)
	}
	if store.createSubjectArg.Code != "MTK" || store.createSubjectArg.Name != "Matematika" {
		t.Fatalf("CreateSubject() arg = %+v, want forwarded subject", store.createSubjectArg)
	}
	if _, err := svc.CreateAssignment(context.Background(), db.CreateClassSubjectAssignmentParams{ClassID: classID, SubjectID: subjectID, TeacherEmployeeID: teacherID}); err != nil {
		t.Fatalf("CreateAssignment() error = %v", err)
	}
	if store.createAssignArg.ClassID != classID || store.createAssignArg.TeacherEmployeeID != teacherID {
		t.Fatalf("CreateAssignment() arg = %+v, want forwarded assignment", store.createAssignArg)
	}

	if _, err := svc.CreateTimetableSlot(context.Background(), db.CreateTimetableSlotParams{AssignmentID: assignmentID, DayOfWeek: 1, StartTime: start, EndTime: end, RoomLabel: " R1 ", Notes: "pagi"}); err != nil {
		t.Fatalf("CreateTimetableSlot() error = %v", err)
	}
	if store.createSlotArg.AssignmentID != assignmentID || store.createSlotArg.RoomLabel != " R1 " {
		t.Fatalf("CreateTimetableSlot() arg = %+v, want forwarded slot", store.createSlotArg)
	}
	if len(store.conflictArgs) != 1 || store.conflictArgs[0].ClassID != classID || store.conflictArgs[0].TeacherEmployeeID != teacherID {
		t.Fatalf("CountTimetableConflicts() args = %+v, want class/teacher from assignment", store.conflictArgs)
	}
	if len(store.roomArgs) != 1 || store.roomArgs[0].RoomLabel != "R1" {
		t.Fatalf("CountTimetableRoomConflicts() args = %+v, want trimmed room", store.roomArgs)
	}

	if _, err := svc.UpdateTimetableSlot(context.Background(), db.UpdateTimetableSlotParams{ID: slotID, AssignmentID: assignmentID, DayOfWeek: 2, StartTime: start, EndTime: end, RoomLabel: "R2", Notes: "siang"}); err != nil {
		t.Fatalf("UpdateTimetableSlot() error = %v", err)
	}
	if store.getSlotID != slotID || store.updateSlotArg.ID != slotID || store.updateSlotArg.DayOfWeek != 2 {
		t.Fatalf("UpdateTimetableSlot() store = get %v update %+v, want slot update", store.getSlotID, store.updateSlotArg)
	}
	if got := store.conflictArgs[len(store.conflictArgs)-1].ExcludeSlotID; got != slotID {
		t.Fatalf("UpdateTimetableSlot() exclude slot = %v, want %v", got, slotID)
	}

	if err := svc.DeleteYear(context.Background(), yearID); err != nil || store.deleteYearID != yearID {
		t.Fatalf("DeleteYear() = %v, id=%v; want nil/%v", err, store.deleteYearID, yearID)
	}
	if err := svc.DeleteClass(context.Background(), classID); err != nil || store.deleteClassID != classID {
		t.Fatalf("DeleteClass() = %v, id=%v; want nil/%v", err, store.deleteClassID, classID)
	}
	if err := svc.DeleteSubject(context.Background(), subjectID); err != nil || store.deleteSubjectID != subjectID {
		t.Fatalf("DeleteSubject() = %v, id=%v; want nil/%v", err, store.deleteSubjectID, subjectID)
	}
	if err := svc.DeleteAssignment(context.Background(), assignmentID); err != nil || store.deleteAssignID != assignmentID {
		t.Fatalf("DeleteAssignment() = %v, id=%v; want nil/%v", err, store.deleteAssignID, assignmentID)
	}
	if err := svc.DeleteTimetableSlot(context.Background(), slotID); err != nil || store.deleteSlotID != slotID {
		t.Fatalf("DeleteTimetableSlot() = %v, id=%v; want nil/%v", err, store.deleteSlotID, slotID)
	}
}

func TestAcademicTimetableAvailabilityFailures(t *testing.T) {
	assignmentID := documentCycleTestUUID(171)
	classID := documentCycleTestUUID(172)
	teacherID := documentCycleTestUUID(173)
	slotID := documentCycleTestUUID(174)
	start, err := ParseAcademicTimeInput("07:30")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(start) error = %v", err)
	}
	end, err := ParseAcademicTimeInput("08:10")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(end) error = %v", err)
	}
	slot := db.CreateTimetableSlotParams{AssignmentID: assignmentID, DayOfWeek: 1, StartTime: start, EndTime: end, RoomLabel: "R1"}
	baseStore := func() *fakeAcademicStore {
		return &fakeAcademicStore{
			getAssignRow: db.GetClassSubjectAssignmentRow{ID: assignmentID, ClassID: classID, TeacherEmployeeID: teacherID},
		}
	}

	svc := &Academic{q: &fakeAcademicStore{getAssignErr: errors.New("missing")}}
	if _, err := svc.CreateTimetableSlot(context.Background(), slot); err == nil || err.Error() != "assignment tidak ditemukan" {
		t.Fatalf("CreateTimetableSlot(assignment missing) = %v, want assignment error", err)
	}
	svc = &Academic{q: &fakeAcademicStore{getSlotErr: errors.New("slot missing")}}
	if _, err := svc.UpdateTimetableSlot(context.Background(), db.UpdateTimetableSlotParams{ID: slotID, AssignmentID: assignmentID}); err == nil || err.Error() != "slot missing" {
		t.Fatalf("UpdateTimetableSlot(get slot error) = %v, want slot missing", err)
	}
	store := baseStore()
	store.conflictErr = errors.New("count failed")
	svc = &Academic{q: store}
	if _, err := svc.CreateTimetableSlot(context.Background(), slot); err == nil || err.Error() != "count failed" {
		t.Fatalf("CreateTimetableSlot(conflict count error) = %v, want count failed", err)
	}
	store = baseStore()
	store.conflictCount = 1
	svc = &Academic{q: store}
	if _, err := svc.CreateTimetableSlot(context.Background(), slot); err == nil || err.Error() != "slot bentrok dengan jadwal kelas atau guru pada waktu yang sama" {
		t.Fatalf("CreateTimetableSlot(class conflict) = %v, want class/guru conflict", err)
	}
	store = baseStore()
	store.roomErr = errors.New("room failed")
	svc = &Academic{q: store}
	if _, err := svc.CreateTimetableSlot(context.Background(), slot); err == nil || err.Error() != "room failed" {
		t.Fatalf("CreateTimetableSlot(room count error) = %v, want room failed", err)
	}
	store = baseStore()
	store.roomCount = 1
	svc = &Academic{q: store}
	if _, err := svc.CreateTimetableSlot(context.Background(), slot); err == nil || err.Error() != "slot bentrok dengan penggunaan ruang pada waktu yang sama" {
		t.Fatalf("CreateTimetableSlot(room conflict) = %v, want room conflict", err)
	}
}

func TestParseAcademicTimeInput(t *testing.T) {
	got, err := ParseAcademicTimeInput("07:30")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(07:30) error = %v", err)
	}
	if !got.Valid {
		t.Fatalf("ParseAcademicTimeInput(07:30) valid = false, want true")
	}
	got, err = ParseAcademicTimeInput(" 07:30:15 ")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(07:30:15) error = %v", err)
	}
	if !got.Valid {
		t.Fatalf("ParseAcademicTimeInput(07:30:15) valid = false, want true")
	}
	if _, err := ParseAcademicTimeInput(" "); err == nil || err.Error() != "waktu wajib diisi" {
		t.Fatalf("ParseAcademicTimeInput(empty) = %v, want required error", err)
	}
	if _, err := ParseAcademicTimeInput("bukan-jam"); err == nil || err.Error() != "format waktu tidak valid" {
		t.Fatalf("ParseAcademicTimeInput(invalid) = %v, want format error", err)
	}
}
