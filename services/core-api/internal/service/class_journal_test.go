package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeClassJournalStore struct {
	countSessions int32
	countErr      error

	createArg db.CreateJournalSessionParams
	createRow db.ClassJournalSession
	createErr error

	sessionRow db.GetJournalSessionRow
	sessionErr error

	sessionByDateArg db.GetJournalSessionIDByAssignmentDateParams
	sessionByDateID  pgtype.UUID
	sessionByDateErr error

	timetableArg db.GetRombelTimetableSlotParams
	timetableRow db.GetRombelTimetableSlotRow
	timetableErr error

	listSessionsArg  pgtype.UUID
	listSessionsRows []db.ListJournalSessionsRow
	listSessionsErr  error

	updateArg db.UpdateJournalSessionParams
	updateRow db.ClassJournalSession
	updateErr error

	deleteID  pgtype.UUID
	deleteErr error

	upsertArgs []db.UpsertJournalAttendanceParams
	upsertErr  error

	attendanceRows []db.ListJournalAttendancesRow
	attendanceErr  error

	summaryArg  pgtype.UUID
	summaryRows []db.ListJournalAttendanceSummaryRow
	summaryErr  error

	studentsArg  pgtype.UUID
	studentsRows []db.ListActiveStudentsByClassIDRow
	studentsErr  error

	assignments    []db.ListClassSubjectAssignmentsRow
	assignmentsErr error

	assignmentRow db.GetClassSubjectAssignmentRow
	assignmentErr error
}

func (f *fakeClassJournalStore) CountJournalSessionsForAssignment(ctx context.Context, assignmentID pgtype.UUID) (int32, error) {
	return f.countSessions, f.countErr
}

func (f *fakeClassJournalStore) CreateJournalSession(ctx context.Context, arg db.CreateJournalSessionParams) (db.ClassJournalSession, error) {
	f.createArg = arg
	if f.createErr != nil {
		return db.ClassJournalSession{}, f.createErr
	}
	return f.createRow, nil
}

func (f *fakeClassJournalStore) GetJournalSession(ctx context.Context, id pgtype.UUID) (db.GetJournalSessionRow, error) {
	if f.sessionErr != nil {
		return db.GetJournalSessionRow{}, f.sessionErr
	}
	return f.sessionRow, nil
}

func (f *fakeClassJournalStore) GetJournalSessionIDByAssignmentDate(ctx context.Context, arg db.GetJournalSessionIDByAssignmentDateParams) (pgtype.UUID, error) {
	f.sessionByDateArg = arg
	if f.sessionByDateErr != nil {
		return pgtype.UUID{}, f.sessionByDateErr
	}
	return f.sessionByDateID, nil
}

func (f *fakeClassJournalStore) GetRombelTimetableSlot(ctx context.Context, arg db.GetRombelTimetableSlotParams) (db.GetRombelTimetableSlotRow, error) {
	f.timetableArg = arg
	if f.timetableErr != nil {
		return db.GetRombelTimetableSlotRow{}, f.timetableErr
	}
	return f.timetableRow, nil
}

func (f *fakeClassJournalStore) ListJournalSessions(ctx context.Context, assignmentID pgtype.UUID) ([]db.ListJournalSessionsRow, error) {
	f.listSessionsArg = assignmentID
	if f.listSessionsErr != nil {
		return nil, f.listSessionsErr
	}
	return f.listSessionsRows, nil
}

func (f *fakeClassJournalStore) UpdateJournalSession(ctx context.Context, arg db.UpdateJournalSessionParams) (db.ClassJournalSession, error) {
	f.updateArg = arg
	if f.updateErr != nil {
		return db.ClassJournalSession{}, f.updateErr
	}
	return f.updateRow, nil
}

func (f *fakeClassJournalStore) DeleteJournalSession(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeClassJournalStore) UpsertJournalAttendance(ctx context.Context, arg db.UpsertJournalAttendanceParams) (db.ClassJournalAttendance, error) {
	f.upsertArgs = append(f.upsertArgs, arg)
	if f.upsertErr != nil {
		return db.ClassJournalAttendance{}, f.upsertErr
	}
	return db.ClassJournalAttendance{SessionID: arg.SessionID, StudentID: arg.StudentID, Status: arg.Status, Catatan: arg.Catatan}, nil
}

func (f *fakeClassJournalStore) ListJournalAttendances(ctx context.Context, sessionID pgtype.UUID) ([]db.ListJournalAttendancesRow, error) {
	if f.attendanceErr != nil {
		return nil, f.attendanceErr
	}
	return f.attendanceRows, nil
}

func (f *fakeClassJournalStore) ListJournalAttendanceSummary(ctx context.Context, assignmentID pgtype.UUID) ([]db.ListJournalAttendanceSummaryRow, error) {
	f.summaryArg = assignmentID
	if f.summaryErr != nil {
		return nil, f.summaryErr
	}
	return f.summaryRows, nil
}

func (f *fakeClassJournalStore) ListActiveStudentsByClassID(ctx context.Context, classID pgtype.UUID) ([]db.ListActiveStudentsByClassIDRow, error) {
	f.studentsArg = classID
	if f.studentsErr != nil {
		return nil, f.studentsErr
	}
	return f.studentsRows, nil
}

func (f *fakeClassJournalStore) ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error) {
	if f.assignmentsErr != nil {
		return nil, f.assignmentsErr
	}
	return f.assignments, nil
}

func (f *fakeClassJournalStore) GetClassSubjectAssignment(ctx context.Context, id pgtype.UUID) (db.GetClassSubjectAssignmentRow, error) {
	if f.assignmentErr != nil {
		return db.GetClassSubjectAssignmentRow{}, f.assignmentErr
	}
	return f.assignmentRow, nil
}

func TestClassJournalOverviewFiltersTeacherAndLoadsSelection(t *testing.T) {
	teacherID := documentCycleTestUUID(41)
	assignmentID := documentCycleTestUUID(42)
	otherAssignmentID := documentCycleTestUUID(43)
	store := &fakeClassJournalStore{
		assignments: []db.ListClassSubjectAssignmentsRow{
			{ID: assignmentID, TeacherEmployeeID: teacherID},
			{ID: otherAssignmentID, TeacherEmployeeID: documentCycleTestUUID(44)},
		},
		assignmentRow: db.GetClassSubjectAssignmentRow{ID: assignmentID, TeacherEmployeeID: teacherID},
		listSessionsRows: []db.ListJournalSessionsRow{
			{ID: documentCycleTestUUID(45), AssignmentID: assignmentID, Materi: "Pecahan"},
		},
		summaryRows: []db.ListJournalAttendanceSummaryRow{
			{StudentID: documentCycleTestUUID(46), Nama: "Siswa A", Hadir: 1},
		},
	}
	svc := &ClassJournal{q: store}

	overview, err := svc.Overview(context.Background(), assignmentID, teacherID)
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if len(overview.Assignments) != 1 || overview.Assignments[0].ID != assignmentID {
		t.Fatalf("assignments = %+v, want only teacher assignment", overview.Assignments)
	}
	if len(overview.Sessions) != 1 || store.listSessionsArg != assignmentID {
		t.Fatalf("sessions were not loaded for selected assignment")
	}
	if len(overview.Summary) != 1 || store.summaryArg != assignmentID {
		t.Fatalf("summary was not loaded for selected assignment")
	}
}

func TestClassJournalOverviewRejectsTeacherOutsideAssignment(t *testing.T) {
	teacherID := documentCycleTestUUID(47)
	store := &fakeClassJournalStore{
		assignments: []db.ListClassSubjectAssignmentsRow{
			{ID: documentCycleTestUUID(48), TeacherEmployeeID: teacherID},
		},
		assignmentRow: db.GetClassSubjectAssignmentRow{
			ID:                documentCycleTestUUID(49),
			TeacherEmployeeID: documentCycleTestUUID(50),
		},
	}
	svc := &ClassJournal{q: store}

	_, err := svc.Overview(context.Background(), documentCycleTestUUID(49), teacherID)
	if err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("Overview() error = %v, want akses ditolak", err)
	}
}

func TestClassJournalCreateSessionTrimsAndSeedsAttendances(t *testing.T) {
	teacherID := documentCycleTestUUID(51)
	assignmentID := documentCycleTestUUID(52)
	classID := documentCycleTestUUID(53)
	sessionID := documentCycleTestUUID(54)
	firstStudentID := documentCycleTestUUID(55)
	secondStudentID := documentCycleTestUUID(56)
	store := &fakeClassJournalStore{
		countSessions: 2,
		assignmentRow: db.GetClassSubjectAssignmentRow{
			ID:                assignmentID,
			ClassID:           classID,
			TeacherEmployeeID: teacherID,
		},
		createRow: db.ClassJournalSession{ID: sessionID, AssignmentID: assignmentID, PertemuanKe: 3},
		studentsRows: []db.ListActiveStudentsByClassIDRow{
			{ID: firstStudentID, Nama: "Siswa A"},
			{ID: secondStudentID, Nama: "Siswa B"},
		},
		sessionRow: db.GetJournalSessionRow{
			ID:                sessionID,
			AssignmentID:      assignmentID,
			TeacherEmployeeID: teacherID,
			Materi:            "Pecahan",
		},
		attendanceRows: []db.ListJournalAttendancesRow{
			{SessionID: sessionID, StudentID: firstStudentID, Status: db.JournalAttendanceStatusHadir},
			{SessionID: sessionID, StudentID: secondStudentID, Status: db.JournalAttendanceStatusHadir},
		},
	}
	svc := &ClassJournal{q: store}

	detail, err := svc.CreateSession(
		context.Background(),
		assignmentID,
		documentCycleTestDate(2026, 5, 2),
		"  Pecahan  ",
		"  Diskusi kelompok  ",
		"  Catatan wali kelas  ",
		true,
		teacherID,
	)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if store.createArg.PertemuanKe != 3 {
		t.Fatalf("pertemuan_ke = %d, want 3", store.createArg.PertemuanKe)
	}
	if store.createArg.Materi != "Pecahan" || store.createArg.Kegiatan != "Diskusi kelompok" || store.createArg.Catatan != "Catatan wali kelas" {
		t.Fatalf("create arg was not normalized: %+v", store.createArg)
	}
	if store.studentsArg != classID {
		t.Fatalf("students class id = %v, want %v", store.studentsArg, classID)
	}
	if len(store.upsertArgs) != 2 {
		t.Fatalf("upsert attendance count = %d, want 2", len(store.upsertArgs))
	}
	for _, arg := range store.upsertArgs {
		if arg.SessionID != sessionID || arg.Status != db.JournalAttendanceStatusHadir {
			t.Fatalf("attendance seed arg = %+v, want hadir for session", arg)
		}
	}
	if detail.Session.ID != sessionID || len(detail.Attendances) != 2 {
		t.Fatalf("detail = %+v, want created session with attendances", detail)
	}
}

func TestClassJournalCreateSessionMapsDuplicateDate(t *testing.T) {
	store := &fakeClassJournalStore{
		assignmentRow: db.GetClassSubjectAssignmentRow{TeacherEmployeeID: documentCycleTestUUID(57)},
		createErr:     errors.New("violates uq_journal_session_date unique constraint"),
	}
	svc := &ClassJournal{q: store}

	_, err := svc.CreateSession(context.Background(), documentCycleTestUUID(58), documentCycleTestDate(2026, 5, 3), "materi", "", "", true, pgtype.UUID{})
	if err == nil || !strings.Contains(err.Error(), "tanggal ini sudah ada") {
		t.Fatalf("CreateSession() error = %v, want duplicate-date message", err)
	}
}

func TestClassJournalOpenSessionFromTimetableSlotReturnsExisting(t *testing.T) {
	classID := documentCycleTestUUID(81)
	slotID := documentCycleTestUUID(82)
	assignmentID := documentCycleTestUUID(83)
	sessionID := documentCycleTestUUID(84)
	teacherID := documentCycleTestUUID(85)
	tanggal := documentCycleTestDate(2026, 5, 6)
	store := &fakeClassJournalStore{
		timetableRow: db.GetRombelTimetableSlotRow{
			ID:                slotID,
			ClassID:           classID,
			AssignmentID:      assignmentID,
			TeacherEmployeeID: teacherID,
			SubjectName:       "IPA",
		},
		sessionByDateID: sessionID,
		sessionRow: db.GetJournalSessionRow{
			ID:                sessionID,
			AssignmentID:      assignmentID,
			TeacherEmployeeID: teacherID,
			Materi:            "Ekosistem",
		},
		attendanceRows: []db.ListJournalAttendancesRow{
			{SessionID: sessionID, StudentID: documentCycleTestUUID(86), Status: db.JournalAttendanceStatusHadir},
		},
	}
	svc := &ClassJournal{q: store}

	result, err := svc.OpenSessionFromTimetableSlot(context.Background(), classID, slotID, tanggal, "  Materi baru  ", "", "", true, teacherID)
	if err != nil {
		t.Fatalf("OpenSessionFromTimetableSlot() error = %v", err)
	}
	if result.Created {
		t.Fatal("Created = true, want false for existing assignment/date session")
	}
	if result.Session.ID != sessionID || result.TimetableSlot.ID != slotID || len(result.Attendances) != 1 {
		t.Fatalf("result = %+v, want existing session detail and slot", result)
	}
	if store.timetableArg.ClassID != classID || store.timetableArg.ID != slotID {
		t.Fatalf("timetable arg = %+v, want scoped class/slot", store.timetableArg)
	}
	if store.sessionByDateArg.AssignmentID != assignmentID || store.sessionByDateArg.Tanggal != tanggal {
		t.Fatalf("session by date arg = %+v, want assignment/date", store.sessionByDateArg)
	}
	if store.createArg.AssignmentID.Valid {
		t.Fatalf("create arg = %+v, want no create for existing session", store.createArg)
	}
}

func TestClassJournalOpenSessionFromTimetableSlotCreatesAndSeeds(t *testing.T) {
	classID := documentCycleTestUUID(87)
	slotID := documentCycleTestUUID(88)
	assignmentID := documentCycleTestUUID(89)
	sessionID := documentCycleTestUUID(90)
	teacherID := documentCycleTestUUID(91)
	studentID := documentCycleTestUUID(92)
	tanggal := documentCycleTestDate(2026, 5, 7)
	store := &fakeClassJournalStore{
		countSessions:    4,
		sessionByDateErr: pgx.ErrNoRows,
		timetableRow:     db.GetRombelTimetableSlotRow{ID: slotID, ClassID: classID, AssignmentID: assignmentID, TeacherEmployeeID: teacherID},
		createRow:        db.ClassJournalSession{ID: sessionID, AssignmentID: assignmentID, PertemuanKe: 5},
		studentsRows:     []db.ListActiveStudentsByClassIDRow{{ID: studentID, Nama: "Siswa A"}},
		sessionRow:       db.GetJournalSessionRow{ID: sessionID, AssignmentID: assignmentID, TeacherEmployeeID: teacherID, Materi: "Pecahan"},
		attendanceRows:   []db.ListJournalAttendancesRow{{SessionID: sessionID, StudentID: studentID, Status: db.JournalAttendanceStatusHadir}},
	}
	svc := &ClassJournal{q: store}

	result, err := svc.OpenSessionFromTimetableSlot(context.Background(), classID, slotID, tanggal, "  Pecahan  ", "  Diskusi  ", "  Siap  ", false, teacherID)
	if err != nil {
		t.Fatalf("OpenSessionFromTimetableSlot() error = %v", err)
	}
	if !result.Created || result.Session.ID != sessionID {
		t.Fatalf("result = %+v, want created session %v", result, sessionID)
	}
	if store.createArg.AssignmentID != assignmentID || store.createArg.Tanggal != tanggal || store.createArg.PertemuanKe != 5 {
		t.Fatalf("create arg = %+v, want assignment/date/pertemuan from slot", store.createArg)
	}
	if store.createArg.Materi != "Pecahan" || store.createArg.Kegiatan != "Diskusi" || store.createArg.Catatan != "Siap" || store.createArg.GuruHadir {
		t.Fatalf("create arg normalization = %+v, want trimmed values and guru_hadir false", store.createArg)
	}
	if store.studentsArg != classID || len(store.upsertArgs) != 1 || store.upsertArgs[0].StudentID != studentID {
		t.Fatalf("attendance seed args class=%v upserts=%+v, want one class student", store.studentsArg, store.upsertArgs)
	}
}

func TestClassJournalOpenSessionFromTimetableSlotRejectsTeacherMismatch(t *testing.T) {
	teacherID := documentCycleTestUUID(93)
	otherTeacherID := documentCycleTestUUID(94)
	store := &fakeClassJournalStore{
		timetableRow: db.GetRombelTimetableSlotRow{
			ID:                documentCycleTestUUID(95),
			ClassID:           documentCycleTestUUID(96),
			AssignmentID:      documentCycleTestUUID(97),
			TeacherEmployeeID: otherTeacherID,
		},
	}
	svc := &ClassJournal{q: store}

	_, err := svc.OpenSessionFromTimetableSlot(context.Background(), store.timetableRow.ClassID, store.timetableRow.ID, documentCycleTestDate(2026, 5, 8), "", "", "", true, teacherID)
	if err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("OpenSessionFromTimetableSlot() error = %v, want akses ditolak", err)
	}
}

func TestClassJournalUpdateDeleteAndBulkUpsert(t *testing.T) {
	teacherID := documentCycleTestUUID(59)
	sessionID := documentCycleTestUUID(60)
	firstStudentID := documentCycleTestUUID(61)
	secondStudentID := documentCycleTestUUID(62)
	store := &fakeClassJournalStore{
		sessionRow: db.GetJournalSessionRow{
			ID:                sessionID,
			TeacherEmployeeID: teacherID,
		},
		updateRow: db.ClassJournalSession{ID: sessionID, Materi: "Bab 2"},
	}
	svc := &ClassJournal{q: store}

	if _, err := svc.UpdateSession(context.Background(), sessionID, "  Bab 2  ", "  Latihan  ", "  selesai  ", false, teacherID); err != nil {
		t.Fatalf("UpdateSession() error = %v", err)
	}
	if store.updateArg.Materi != "Bab 2" || store.updateArg.Kegiatan != "Latihan" || store.updateArg.Catatan != "selesai" {
		t.Fatalf("update arg was not normalized: %+v", store.updateArg)
	}
	if err := svc.DeleteSession(context.Background(), sessionID, teacherID); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if store.deleteID != sessionID {
		t.Fatalf("delete id = %v, want %v", store.deleteID, sessionID)
	}

	err := svc.BulkUpsertAttendances(context.Background(), sessionID, []JournalAttendanceEntry{
		{StudentID: firstStudentID.String(), Status: " sakit ", Catatan: "  izin dokter  "},
		{StudentID: secondStudentID.String(), Status: "ALPHA", Catatan: "  tanpa kabar  "},
	}, teacherID)
	if err != nil {
		t.Fatalf("BulkUpsertAttendances() error = %v", err)
	}
	if len(store.upsertArgs) != 2 {
		t.Fatalf("upsert count = %d, want 2", len(store.upsertArgs))
	}
	if store.upsertArgs[0].Status != db.JournalAttendanceStatusSakit || store.upsertArgs[0].Catatan != "izin dokter" {
		t.Fatalf("first attendance arg = %+v, want sakit and trimmed note", store.upsertArgs[0])
	}
	if store.upsertArgs[1].Status != db.JournalAttendanceStatusAlpha || store.upsertArgs[1].Catatan != "tanpa kabar" {
		t.Fatalf("second attendance arg = %+v, want alpha and trimmed note", store.upsertArgs[1])
	}
}

func TestClassJournalRejectsTeacherMismatchAndInvalidAttendance(t *testing.T) {
	teacherID := documentCycleTestUUID(63)
	otherTeacherID := documentCycleTestUUID(64)
	sessionID := documentCycleTestUUID(65)
	store := &fakeClassJournalStore{
		sessionRow: db.GetJournalSessionRow{
			ID:                sessionID,
			TeacherEmployeeID: otherTeacherID,
		},
	}
	svc := &ClassJournal{q: store}

	if _, err := svc.GetSession(context.Background(), sessionID, teacherID); err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("GetSession() error = %v, want akses ditolak", err)
	}
	if err := svc.DeleteSession(context.Background(), sessionID, teacherID); err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("DeleteSession() error = %v, want akses ditolak", err)
	}

	store.sessionRow.TeacherEmployeeID = teacherID
	err := svc.BulkUpsertAttendances(context.Background(), sessionID, []JournalAttendanceEntry{
		{StudentID: "not-a-uuid", Status: "hadir"},
	}, teacherID)
	if err == nil || !strings.Contains(err.Error(), "student_id tidak valid") {
		t.Fatalf("BulkUpsertAttendances() error = %v, want invalid student id", err)
	}

	err = svc.BulkUpsertAttendances(context.Background(), sessionID, []JournalAttendanceEntry{
		{StudentID: documentCycleTestUUID(66).String(), Status: "mangkir"},
	}, teacherID)
	if err == nil || !strings.Contains(err.Error(), "status kehadiran tidak valid") {
		t.Fatalf("BulkUpsertAttendances() error = %v, want invalid status", err)
	}
}

func TestClassJournalPropagatesOverviewAndDetailStoreErrors(t *testing.T) {
	ctx := context.Background()
	assignmentID := documentCycleTestUUID(67)
	sessionID := documentCycleTestUUID(68)
	teacherID := documentCycleTestUUID(69)
	expectedErr := errors.New("store failed")

	svc := &ClassJournal{q: &fakeClassJournalStore{assignmentsErr: expectedErr}}
	if _, err := svc.Overview(ctx, pgtype.UUID{}, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview() error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{assignmentErr: expectedErr}}
	if _, err := svc.Overview(ctx, assignmentID, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview() assignment error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{listSessionsErr: expectedErr}}
	if _, err := svc.Overview(ctx, assignmentID, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview() sessions error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{summaryErr: expectedErr}}
	if _, err := svc.Overview(ctx, assignmentID, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Overview() summary error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{sessionErr: expectedErr}}
	if _, err := svc.GetSession(ctx, sessionID, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("GetSession() session error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		sessionRow:    db.GetJournalSessionRow{ID: sessionID},
		attendanceErr: expectedErr,
	}}
	if _, err := svc.GetSession(ctx, sessionID, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("GetSession() attendance error = %v, want %v", err, expectedErr)
	}
}

func TestClassJournalCreateSessionErrorBranches(t *testing.T) {
	ctx := context.Background()
	assignmentID := documentCycleTestUUID(70)
	sessionID := documentCycleTestUUID(71)
	classID := documentCycleTestUUID(72)
	teacherID := documentCycleTestUUID(73)
	otherTeacherID := documentCycleTestUUID(74)
	expectedErr := errors.New("store failed")

	svc := &ClassJournal{q: &fakeClassJournalStore{assignmentErr: expectedErr}}
	if _, err := svc.CreateSession(ctx, assignmentID, documentCycleTestDate(2026, 5, 4), "materi", "", "", true, pgtype.UUID{}); err == nil || !strings.Contains(err.Error(), "assignment tidak ditemukan") {
		t.Fatalf("CreateSession() assignment error = %v, want assignment tidak ditemukan", err)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		assignmentRow: db.GetClassSubjectAssignmentRow{ID: assignmentID, TeacherEmployeeID: otherTeacherID},
	}}
	if _, err := svc.CreateSession(ctx, assignmentID, documentCycleTestDate(2026, 5, 4), "materi", "", "", true, teacherID); err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("CreateSession() teacher mismatch error = %v, want akses ditolak", err)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		assignmentRow: db.GetClassSubjectAssignmentRow{ID: assignmentID, TeacherEmployeeID: teacherID},
		countErr:      expectedErr,
	}}
	if _, err := svc.CreateSession(ctx, assignmentID, documentCycleTestDate(2026, 5, 4), "materi", "", "", true, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("CreateSession() count error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		assignmentRow: db.GetClassSubjectAssignmentRow{ID: assignmentID, TeacherEmployeeID: teacherID},
		createErr:     expectedErr,
	}}
	if _, err := svc.CreateSession(ctx, assignmentID, documentCycleTestDate(2026, 5, 4), "materi", "", "", true, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("CreateSession() create error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		assignmentRow: db.GetClassSubjectAssignmentRow{ID: assignmentID, ClassID: classID, TeacherEmployeeID: teacherID},
		createRow:     db.ClassJournalSession{ID: sessionID, AssignmentID: assignmentID},
		studentsErr:   expectedErr,
	}}
	if _, err := svc.CreateSession(ctx, assignmentID, documentCycleTestDate(2026, 5, 4), "materi", "", "", true, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("CreateSession() students error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		assignmentRow: db.GetClassSubjectAssignmentRow{ID: assignmentID, ClassID: classID, TeacherEmployeeID: teacherID},
		createRow:     db.ClassJournalSession{ID: sessionID, AssignmentID: assignmentID},
		sessionRow:    db.GetJournalSessionRow{ID: sessionID, TeacherEmployeeID: teacherID},
		attendanceErr: expectedErr,
	}}
	if _, err := svc.CreateSession(ctx, assignmentID, documentCycleTestDate(2026, 5, 4), "materi", "", "", true, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("CreateSession() detail attendance error = %v, want %v", err, expectedErr)
	}
}

func TestClassJournalMutationErrorBranches(t *testing.T) {
	ctx := context.Background()
	sessionID := documentCycleTestUUID(75)
	studentID := documentCycleTestUUID(76)
	teacherID := documentCycleTestUUID(77)
	otherTeacherID := documentCycleTestUUID(78)
	expectedErr := errors.New("store failed")

	svc := &ClassJournal{q: &fakeClassJournalStore{sessionErr: expectedErr}}
	if _, err := svc.UpdateSession(ctx, sessionID, "", "", "", true, pgtype.UUID{}); err == nil || !strings.Contains(err.Error(), "sesi tidak ditemukan") {
		t.Fatalf("UpdateSession() session error = %v, want sesi tidak ditemukan", err)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		sessionRow: db.GetJournalSessionRow{ID: sessionID, TeacherEmployeeID: otherTeacherID},
	}}
	if _, err := svc.UpdateSession(ctx, sessionID, "", "", "", true, teacherID); err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("UpdateSession() teacher mismatch error = %v, want akses ditolak", err)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		sessionRow: db.GetJournalSessionRow{ID: sessionID, TeacherEmployeeID: teacherID},
		updateErr:  expectedErr,
	}}
	if _, err := svc.UpdateSession(ctx, sessionID, "materi", "kegiatan", "catatan", true, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateSession() update error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{sessionErr: expectedErr}}
	if err := svc.DeleteSession(ctx, sessionID, teacherID); err == nil || !strings.Contains(err.Error(), "sesi tidak ditemukan") {
		t.Fatalf("DeleteSession() session error = %v, want sesi tidak ditemukan", err)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{deleteErr: expectedErr}}
	if err := svc.DeleteSession(ctx, sessionID, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("DeleteSession() delete error = %v, want %v", err, expectedErr)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{sessionErr: expectedErr}}
	if err := svc.BulkUpsertAttendances(ctx, sessionID, nil, pgtype.UUID{}); err == nil || !strings.Contains(err.Error(), "sesi tidak ditemukan") {
		t.Fatalf("BulkUpsertAttendances() session error = %v, want sesi tidak ditemukan", err)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		sessionRow: db.GetJournalSessionRow{ID: sessionID, TeacherEmployeeID: otherTeacherID},
	}}
	if err := svc.BulkUpsertAttendances(ctx, sessionID, nil, teacherID); err == nil || !strings.Contains(err.Error(), "akses ditolak") {
		t.Fatalf("BulkUpsertAttendances() teacher mismatch error = %v, want akses ditolak", err)
	}

	svc = &ClassJournal{q: &fakeClassJournalStore{
		sessionRow: db.GetJournalSessionRow{ID: sessionID, TeacherEmployeeID: teacherID},
		upsertErr:  expectedErr,
	}}
	err := svc.BulkUpsertAttendances(ctx, sessionID, []JournalAttendanceEntry{
		{StudentID: studentID.String(), Status: "izin", Catatan: "rapat keluarga"},
	}, teacherID)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("BulkUpsertAttendances() upsert error = %v, want %v", err, expectedErr)
	}
	storeUpserts := svc.q.(*fakeClassJournalStore).upsertArgs
	if len(storeUpserts) != 1 || storeUpserts[0].Status != db.JournalAttendanceStatusIzin {
		t.Fatalf("upsert args = %+v, want izin status", storeUpserts)
	}
}
