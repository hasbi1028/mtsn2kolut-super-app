package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func kesiswaanTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func kesiswaanTestDate() pgtype.Date {
	return pgtype.Date{Time: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC), Valid: true}
}

type fakeKesiswaanStore struct {
	statsRow        db.GetKesiswaanStatsRow
	statsErr        error
	statsTeacherID  pgtype.UUID
	teacherStatsRow db.GetKesiswaanStatsByTeacherRow
	teacherStatsErr error

	classOptions    []db.ListKesiswaanClassOptionsRow
	classOptionsErr error

	listStudentsArg db.ListKesiswaanStudentsParams
	listStudentsRow []db.ListKesiswaanStudentsRow
	listStudentsErr error
	profileArg      db.UpdateKesiswaanStudentProfileParams
	profileErr      error
	photoArg        db.UpdateKesiswaanStudentPhotoParams
	photoErr        error
	canReadArg      db.CanReadKesiswaanStudentPhotoParams
	canRead         bool
	canReadErr      error

	categorySearch string
	categoryCreate db.CreateViolationCategoryParams
	categoryUpdate db.UpdateViolationCategoryParams
	categoryDelete pgtype.UUID

	violationListArg db.ListStudentViolationsParams
	violationCreate  db.CreateStudentViolationParams
	violationUpdate  db.UpdateStudentViolationParams
	violationDelete  pgtype.UUID

	achievementListArg db.ListStudentAchievementsParams
	achievementCreate  db.CreateStudentAchievementParams
	achievementUpdate  db.UpdateStudentAchievementParams
	achievementDelete  pgtype.UUID

	extraListArg db.ListExtracurricularsParams
	extraCreate  db.CreateExtracurricularParams
	extraUpdate  db.UpdateExtracurricularParams
	extraDelete  pgtype.UUID

	memberListArg db.ListExtracurricularMembersParams
	memberCreate  db.CreateExtracurricularMemberParams
	memberUpdate  db.UpdateExtracurricularMemberParams
	memberDelete  pgtype.UUID

	counselingListArg db.ListCounselingSessionsParams
	counselingCreate  db.CreateCounselingSessionParams
	counselingUpdate  db.UpdateCounselingSessionParams
	counselingDelete  pgtype.UUID

	transferListArg db.ListStudentTransfersParams
	transferCreate  db.CreateStudentTransferParams
	transferErr     error
	transferRow     db.StudentTransfer
	lifecycleArg    db.UpdateKesiswaanStudentLifecycleParams
	lifecycleErr    error
}

func (f *fakeKesiswaanStore) GetKesiswaanStats(ctx context.Context) (db.GetKesiswaanStatsRow, error) {
	return f.statsRow, f.statsErr
}

func (f *fakeKesiswaanStore) GetKesiswaanStatsByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) (db.GetKesiswaanStatsByTeacherRow, error) {
	f.statsTeacherID = teacherEmployeeID
	return f.teacherStatsRow, f.teacherStatsErr
}

func (f *fakeKesiswaanStore) ListKesiswaanClassOptions(ctx context.Context) ([]db.ListKesiswaanClassOptionsRow, error) {
	return f.classOptions, f.classOptionsErr
}

func (f *fakeKesiswaanStore) ListKesiswaanStudents(ctx context.Context, arg db.ListKesiswaanStudentsParams) ([]db.ListKesiswaanStudentsRow, error) {
	f.listStudentsArg = arg
	return f.listStudentsRow, f.listStudentsErr
}

func (f *fakeKesiswaanStore) UpdateKesiswaanStudentProfile(ctx context.Context, arg db.UpdateKesiswaanStudentProfileParams) (db.Student, error) {
	f.profileArg = arg
	return db.Student{ID: arg.ID, Nik: arg.Nik, TempatLahir: arg.TempatLahir, Alamat: arg.Alamat, Agama: arg.Agama, Phone: arg.Phone}, f.profileErr
}

func (f *fakeKesiswaanStore) UpdateKesiswaanStudentPhoto(ctx context.Context, arg db.UpdateKesiswaanStudentPhotoParams) (db.Student, error) {
	f.photoArg = arg
	return db.Student{ID: arg.ID, PhotoUrl: arg.PhotoUrl}, f.photoErr
}

func (f *fakeKesiswaanStore) CanReadKesiswaanStudentPhoto(ctx context.Context, arg db.CanReadKesiswaanStudentPhotoParams) (bool, error) {
	f.canReadArg = arg
	return f.canRead, f.canReadErr
}

func (f *fakeKesiswaanStore) ListViolationCategories(ctx context.Context, search string) ([]db.ViolationCategory, error) {
	f.categorySearch = search
	return []db.ViolationCategory{{Code: "T01", Name: "Terlambat"}}, nil
}

func (f *fakeKesiswaanStore) CreateViolationCategory(ctx context.Context, arg db.CreateViolationCategoryParams) (db.ViolationCategory, error) {
	f.categoryCreate = arg
	return db.ViolationCategory{Code: arg.Code, Name: arg.Name, Point: arg.Point, Severity: arg.Severity, Description: arg.Description, IsActive: arg.IsActive}, nil
}

func (f *fakeKesiswaanStore) UpdateViolationCategory(ctx context.Context, arg db.UpdateViolationCategoryParams) (db.ViolationCategory, error) {
	f.categoryUpdate = arg
	return db.ViolationCategory{ID: arg.ID, Code: arg.Code, Name: arg.Name, Point: arg.Point, Severity: arg.Severity, Description: arg.Description, IsActive: arg.IsActive}, nil
}

func (f *fakeKesiswaanStore) DeleteViolationCategory(ctx context.Context, id pgtype.UUID) error {
	f.categoryDelete = id
	return nil
}

func (f *fakeKesiswaanStore) ListStudentViolations(ctx context.Context, arg db.ListStudentViolationsParams) ([]db.ListStudentViolationsRow, error) {
	f.violationListArg = arg
	return []db.ListStudentViolationsRow{{ID: kesiswaanTestUUID(11), Status: arg.Status}}, nil
}

func (f *fakeKesiswaanStore) CreateStudentViolation(ctx context.Context, arg db.CreateStudentViolationParams) (db.StudentViolation, error) {
	f.violationCreate = arg
	return db.StudentViolation{StudentID: arg.StudentID, IncidentDate: arg.IncidentDate, Points: arg.Points, Description: arg.Description, ActionTaken: arg.ActionTaken, Status: arg.Status}, nil
}

func (f *fakeKesiswaanStore) UpdateStudentViolation(ctx context.Context, arg db.UpdateStudentViolationParams) (db.StudentViolation, error) {
	f.violationUpdate = arg
	return db.StudentViolation{ID: arg.ID, StudentID: arg.StudentID, IncidentDate: arg.IncidentDate, Points: arg.Points, Description: arg.Description, ActionTaken: arg.ActionTaken, Status: arg.Status}, nil
}

func (f *fakeKesiswaanStore) DeleteStudentViolation(ctx context.Context, id pgtype.UUID) error {
	f.violationDelete = id
	return nil
}

func (f *fakeKesiswaanStore) ListStudentAchievements(ctx context.Context, arg db.ListStudentAchievementsParams) ([]db.ListStudentAchievementsRow, error) {
	f.achievementListArg = arg
	return []db.ListStudentAchievementsRow{{ID: kesiswaanTestUUID(12), Level: arg.Level}}, nil
}

func (f *fakeKesiswaanStore) CreateStudentAchievement(ctx context.Context, arg db.CreateStudentAchievementParams) (db.StudentAchievement, error) {
	f.achievementCreate = arg
	return db.StudentAchievement{StudentID: arg.StudentID, AchievementDate: arg.AchievementDate, Title: arg.Title, Level: arg.Level, Category: arg.Category, Organizer: arg.Organizer, Description: arg.Description, DocumentUrl: arg.DocumentUrl}, nil
}

func (f *fakeKesiswaanStore) UpdateStudentAchievement(ctx context.Context, arg db.UpdateStudentAchievementParams) (db.StudentAchievement, error) {
	f.achievementUpdate = arg
	return db.StudentAchievement{ID: arg.ID, StudentID: arg.StudentID, AchievementDate: arg.AchievementDate, Title: arg.Title, Level: arg.Level, Category: arg.Category, Organizer: arg.Organizer, Description: arg.Description, DocumentUrl: arg.DocumentUrl}, nil
}

func (f *fakeKesiswaanStore) DeleteStudentAchievement(ctx context.Context, id pgtype.UUID) error {
	f.achievementDelete = id
	return nil
}

func (f *fakeKesiswaanStore) ListExtracurriculars(ctx context.Context, arg db.ListExtracurricularsParams) ([]db.ListExtracurricularsRow, error) {
	f.extraListArg = arg
	return []db.ListExtracurricularsRow{{ID: kesiswaanTestUUID(13), Code: "PRM", Name: "Pramuka"}}, nil
}

func (f *fakeKesiswaanStore) CreateExtracurricular(ctx context.Context, arg db.CreateExtracurricularParams) (db.Extracurricular, error) {
	f.extraCreate = arg
	return db.Extracurricular{Code: arg.Code, Name: arg.Name, Category: arg.Category, Description: arg.Description, ScheduleText: arg.ScheduleText, IsActive: arg.IsActive}, nil
}

func (f *fakeKesiswaanStore) UpdateExtracurricular(ctx context.Context, arg db.UpdateExtracurricularParams) (db.Extracurricular, error) {
	f.extraUpdate = arg
	return db.Extracurricular{ID: arg.ID, Code: arg.Code, Name: arg.Name, Category: arg.Category, Description: arg.Description, ScheduleText: arg.ScheduleText, IsActive: arg.IsActive}, nil
}

func (f *fakeKesiswaanStore) DeleteExtracurricular(ctx context.Context, id pgtype.UUID) error {
	f.extraDelete = id
	return nil
}

func (f *fakeKesiswaanStore) ListExtracurricularMembers(ctx context.Context, arg db.ListExtracurricularMembersParams) ([]db.ListExtracurricularMembersRow, error) {
	f.memberListArg = arg
	return []db.ListExtracurricularMembersRow{{ID: kesiswaanTestUUID(14), Status: arg.Status}}, nil
}

func (f *fakeKesiswaanStore) CreateExtracurricularMember(ctx context.Context, arg db.CreateExtracurricularMemberParams) (db.ExtracurricularMember, error) {
	f.memberCreate = arg
	return db.ExtracurricularMember{ExtracurricularID: arg.ExtracurricularID, StudentID: arg.StudentID, JoinedAt: arg.JoinedAt, Role: arg.Role, Status: arg.Status, Notes: arg.Notes}, nil
}

func (f *fakeKesiswaanStore) UpdateExtracurricularMember(ctx context.Context, arg db.UpdateExtracurricularMemberParams) (db.ExtracurricularMember, error) {
	f.memberUpdate = arg
	return db.ExtracurricularMember{ID: arg.ID, ExtracurricularID: arg.ExtracurricularID, StudentID: arg.StudentID, JoinedAt: arg.JoinedAt, Role: arg.Role, Status: arg.Status, Notes: arg.Notes}, nil
}

func (f *fakeKesiswaanStore) DeleteExtracurricularMember(ctx context.Context, id pgtype.UUID) error {
	f.memberDelete = id
	return nil
}

func (f *fakeKesiswaanStore) ListCounselingSessions(ctx context.Context, arg db.ListCounselingSessionsParams) ([]db.ListCounselingSessionsRow, error) {
	f.counselingListArg = arg
	return []db.ListCounselingSessionsRow{{ID: kesiswaanTestUUID(15), Status: arg.Status}}, nil
}

func (f *fakeKesiswaanStore) CreateCounselingSession(ctx context.Context, arg db.CreateCounselingSessionParams) (db.CounselingSession, error) {
	f.counselingCreate = arg
	return db.CounselingSession{StudentID: arg.StudentID, SessionDate: arg.SessionDate, Topic: arg.Topic, Summary: arg.Summary, FollowUp: arg.FollowUp, Status: arg.Status, IsConfidential: arg.IsConfidential}, nil
}

func (f *fakeKesiswaanStore) UpdateCounselingSession(ctx context.Context, arg db.UpdateCounselingSessionParams) (db.CounselingSession, error) {
	f.counselingUpdate = arg
	return db.CounselingSession{ID: arg.ID, StudentID: arg.StudentID, SessionDate: arg.SessionDate, Topic: arg.Topic, Summary: arg.Summary, FollowUp: arg.FollowUp, Status: arg.Status, IsConfidential: arg.IsConfidential}, nil
}

func (f *fakeKesiswaanStore) DeleteCounselingSession(ctx context.Context, id pgtype.UUID) error {
	f.counselingDelete = id
	return nil
}

func (f *fakeKesiswaanStore) ListStudentTransfers(ctx context.Context, arg db.ListStudentTransfersParams) ([]db.ListStudentTransfersRow, error) {
	f.transferListArg = arg
	return []db.ListStudentTransfersRow{{ID: kesiswaanTestUUID(16), TransferType: arg.TransferType}}, nil
}

func (f *fakeKesiswaanStore) CreateStudentTransfer(ctx context.Context, arg db.CreateStudentTransferParams) (db.StudentTransfer, error) {
	f.transferCreate = arg
	if f.transferRow.ID.Valid {
		return f.transferRow, f.transferErr
	}
	return db.StudentTransfer{ID: kesiswaanTestUUID(17), StudentID: arg.StudentID, TransferType: arg.TransferType}, f.transferErr
}

func (f *fakeKesiswaanStore) UpdateKesiswaanStudentLifecycle(ctx context.Context, arg db.UpdateKesiswaanStudentLifecycleParams) (db.Student, error) {
	f.lifecycleArg = arg
	return db.Student{ID: arg.ID, Status: arg.Status, IsActive: arg.IsActive}, f.lifecycleErr
}

func TestKesiswaanParseHelpers(t *testing.T) {
	id, err := ParseKesiswaanOptionalUUID(" 00000000-0000-0000-0000-000000000001 ")
	if err != nil {
		t.Fatalf("ParseKesiswaanOptionalUUID() error = %v", err)
	}
	if !id.Valid {
		t.Fatalf("ParseKesiswaanOptionalUUID() valid = false, want true")
	}
	if emptyID, err := ParseKesiswaanOptionalUUID(" "); err != nil || emptyID.Valid {
		t.Fatalf("ParseKesiswaanOptionalUUID(empty) = %v, %v; want invalid nil", emptyID, err)
	}
	if _, err := ParseKesiswaanOptionalUUID("bad"); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ParseKesiswaanOptionalUUID(bad) error = %v, want invalid id", err)
	}

	date, err := ParseKesiswaanDate(" 2026-05-01 ")
	if err != nil {
		t.Fatalf("ParseKesiswaanDate() error = %v", err)
	}
	if !date.Valid || date.Time.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("ParseKesiswaanDate() = %v, want 2026-05-01", date.Time)
	}
	if _, err := ParseKesiswaanDate(" "); err == nil || err.Error() != "tanggal wajib diisi" {
		t.Fatalf("ParseKesiswaanDate(empty) error = %v, want required date", err)
	}
	if _, err := ParseKesiswaanDate("bad"); err == nil || err.Error() != "format tanggal tidak valid" {
		t.Fatalf("ParseKesiswaanDate(bad) error = %v, want invalid date", err)
	}

	if emptyDate, err := ParseKesiswaanOptionalDate(" "); err != nil || emptyDate.Valid {
		t.Fatalf("ParseKesiswaanOptionalDate(empty) = %v, %v; want invalid nil", emptyDate, err)
	}
	if optionalDate, err := ParseKesiswaanOptionalDate(" 2026-05-02 "); err != nil || !optionalDate.Valid || optionalDate.Time.Format("2006-01-02") != "2026-05-02" {
		t.Fatalf("ParseKesiswaanOptionalDate(valid) = %v, %v; want 2026-05-02 nil", optionalDate, err)
	}
	if _, err := ParseKesiswaanOptionalDate("bad"); err == nil || err.Error() != "format tanggal tidak valid" {
		t.Fatalf("ParseKesiswaanOptionalDate(bad) error = %v, want invalid optional date", err)
	}

	value := int32(3)
	if got := KesiswaanOptionalInt4(&value); !got.Valid || got.Int32 != 3 {
		t.Fatalf("KesiswaanOptionalInt4() = %+v, want valid 3", got)
	}
	if got := KesiswaanOptionalInt4(nil); got.Valid {
		t.Fatalf("KesiswaanOptionalInt4(nil) valid = true, want false")
	}
	if got := normalizeKesiswaanText("  ACTIVE  ", "fallback"); got != "active" {
		t.Fatalf("normalizeKesiswaanText() = %q, want active", got)
	}
	if got := normalizeKesiswaanText(" ", "fallback"); got != "fallback" {
		t.Fatalf("normalizeKesiswaanText(empty) = %q, want fallback", got)
	}
}

func TestKesiswaanStatsAndListMethodsForwardFilters(t *testing.T) {
	teacherID := kesiswaanTestUUID(21)
	classID := kesiswaanTestUUID(22)
	studentID := kesiswaanTestUUID(23)
	extraID := kesiswaanTestUUID(24)
	store := &fakeKesiswaanStore{
		statsRow:        db.GetKesiswaanStatsRow{ActiveStudents: 10},
		teacherStatsRow: db.GetKesiswaanStatsByTeacherRow{ActiveStudents: 4},
		classOptions:    []db.ListKesiswaanClassOptionsRow{{ID: classID, Code: "7A", Name: "VII A", Level: "VII"}},
		listStudentsRow: []db.ListKesiswaanStudentsRow{{ID: studentID, Nama: "Siswa A"}},
	}
	svc := &Kesiswaan{q: store, photoDir: t.TempDir()}

	stats, err := svc.Stats(context.Background(), pgtype.UUID{})
	if err != nil {
		t.Fatalf("Stats(admin) error = %v", err)
	}
	if stats.ActiveStudents != 10 {
		t.Fatalf("Stats(admin).ActiveStudents = %d, want 10", stats.ActiveStudents)
	}
	stats, err = svc.Stats(context.Background(), teacherID)
	if err != nil {
		t.Fatalf("Stats(teacher) error = %v", err)
	}
	if stats.ActiveStudents != 4 || store.statsTeacherID != teacherID {
		t.Fatalf("Stats(teacher) = %+v, teacher arg %v; want active 4 and teacher id", stats, store.statsTeacherID)
	}

	classes, err := svc.ClassOptions(context.Background())
	if err != nil || len(classes) != 1 {
		t.Fatalf("ClassOptions() = %d rows, %v; want 1 nil", len(classes), err)
	}
	if _, err := svc.ListStudents(context.Background(), "  Siswa  ", "ACTIVE", classID.String(), teacherID); err != nil {
		t.Fatalf("ListStudents() error = %v", err)
	}
	if store.listStudentsArg.Search != "Siswa" || store.listStudentsArg.Status != "active" || store.listStudentsArg.ClassID != classID || store.listStudentsArg.TeacherEmployeeID != teacherID {
		t.Fatalf("ListStudents() arg = %+v, want normalized filters", store.listStudentsArg)
	}
	if _, err := svc.ListStudents(context.Background(), "", "blocked", "", pgtype.UUID{}); err == nil || err.Error() != "status siswa tidak valid" {
		t.Fatalf("ListStudents(invalid status) error = %v, want status siswa tidak valid", err)
	}

	if _, err := svc.ListCategories(context.Background(), "  telat  "); err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if store.categorySearch != "telat" {
		t.Fatalf("ListCategories() search = %q, want trimmed telat", store.categorySearch)
	}
	if _, err := svc.ListViolations(context.Background(), "  kasus  ", "OPEN", studentID.String(), teacherID); err != nil {
		t.Fatalf("ListViolations() error = %v", err)
	}
	if store.violationListArg.Search != "kasus" || store.violationListArg.Status != "open" || store.violationListArg.StudentID != studentID || store.violationListArg.TeacherEmployeeID != teacherID {
		t.Fatalf("ListViolations() arg = %+v, want normalized filters", store.violationListArg)
	}
	if _, err := svc.ListViolations(context.Background(), "", "pending", "", pgtype.UUID{}); err == nil || err.Error() != "status pelanggaran tidak valid" {
		t.Fatalf("ListViolations(invalid status) error = %v, want invalid status", err)
	}
	if _, err := svc.ListViolations(context.Background(), "", "open", "bad-id", pgtype.UUID{}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListViolations(invalid student id) error = %v, want invalid id", err)
	}

	if _, err := svc.ListAchievements(context.Background(), "  juara  ", "SCHOOL", studentID.String(), teacherID); err != nil {
		t.Fatalf("ListAchievements() error = %v", err)
	}
	if store.achievementListArg.Search != "juara" || store.achievementListArg.Level != "school" || store.achievementListArg.StudentID != studentID {
		t.Fatalf("ListAchievements() arg = %+v, want normalized filters", store.achievementListArg)
	}
	if _, err := svc.ListAchievements(context.Background(), "", "desa", "", pgtype.UUID{}); err == nil || err.Error() != "tingkat prestasi tidak valid" {
		t.Fatalf("ListAchievements(invalid level) error = %v, want invalid level", err)
	}
	if _, err := svc.ListAchievements(context.Background(), "", "school", "bad-id", pgtype.UUID{}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListAchievements(invalid student id) error = %v, want invalid id", err)
	}

	if _, err := svc.ListExtracurriculars(context.Background(), "  pramuka  ", true); err != nil {
		t.Fatalf("ListExtracurriculars() error = %v", err)
	}
	if store.extraListArg.Search != "pramuka" || !store.extraListArg.ActiveOnly {
		t.Fatalf("ListExtracurriculars() arg = %+v, want normalized filters", store.extraListArg)
	}
	if _, err := svc.ListExtracurricularMembers(context.Background(), "  ketua  ", "ACTIVE", extraID.String(), studentID.String(), teacherID); err != nil {
		t.Fatalf("ListExtracurricularMembers() error = %v", err)
	}
	if store.memberListArg.Search != "ketua" || store.memberListArg.Status != "active" || store.memberListArg.ExtracurricularID != extraID || store.memberListArg.StudentID != studentID {
		t.Fatalf("ListExtracurricularMembers() arg = %+v, want normalized filters", store.memberListArg)
	}
	if _, err := svc.ListExtracurricularMembers(context.Background(), "", "pending", "", "", pgtype.UUID{}); err == nil || err.Error() != "status anggota ekskul tidak valid" {
		t.Fatalf("ListExtracurricularMembers(invalid status) error = %v, want invalid status", err)
	}
	if _, err := svc.ListExtracurricularMembers(context.Background(), "", "active", "bad-id", "", pgtype.UUID{}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListExtracurricularMembers(invalid extracurricular id) error = %v, want invalid id", err)
	}
	if _, err := svc.ListExtracurricularMembers(context.Background(), "", "active", extraID.String(), "bad-id", pgtype.UUID{}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListExtracurricularMembers(invalid student id) error = %v, want invalid id", err)
	}

	if _, err := svc.ListCounselingSessions(context.Background(), "  belajar  ", "OPEN", studentID.String(), true, teacherID); err != nil {
		t.Fatalf("ListCounselingSessions() error = %v", err)
	}
	if store.counselingListArg.Search != "belajar" || store.counselingListArg.Status != "open" || !store.counselingListArg.CanReadConfidential || store.counselingListArg.TeacherEmployeeID != teacherID {
		t.Fatalf("ListCounselingSessions() arg = %+v, want normalized filters", store.counselingListArg)
	}
	if _, err := svc.ListCounselingSessions(context.Background(), "", "pending", "", false, pgtype.UUID{}); err == nil || err.Error() != "status konseling tidak valid" {
		t.Fatalf("ListCounselingSessions(invalid status) error = %v, want invalid status", err)
	}
	if _, err := svc.ListCounselingSessions(context.Background(), "", "open", "bad-id", false, pgtype.UUID{}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListCounselingSessions(invalid student id) error = %v, want invalid id", err)
	}

	if _, err := svc.ListStudentTransfers(context.Background(), "  pindah  ", "OUT", studentID.String(), teacherID); err != nil {
		t.Fatalf("ListStudentTransfers() error = %v", err)
	}
	if store.transferListArg.Search != "pindah" || store.transferListArg.TransferType != "out" || store.transferListArg.StudentID != studentID || store.transferListArg.TeacherEmployeeID != teacherID {
		t.Fatalf("ListStudentTransfers() arg = %+v, want normalized filters", store.transferListArg)
	}
	if _, err := svc.ListStudentTransfers(context.Background(), "", "stay", "", pgtype.UUID{}); err == nil || err.Error() != "jenis mutasi tidak valid" {
		t.Fatalf("ListStudentTransfers(invalid type) error = %v, want invalid type", err)
	}
	if _, err := svc.ListStudentTransfers(context.Background(), "", "out", "bad-id", pgtype.UUID{}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListStudentTransfers(invalid student id) error = %v, want invalid id", err)
	}
}

type failingReader struct{ err error }

func (r failingReader) Read(p []byte) (int, error) {
	return 0, r.err
}

func TestKesiswaanStatsListProfileAndPhotoErrorBranches(t *testing.T) {
	ctx := context.Background()
	teacherID := kesiswaanTestUUID(25)
	studentID := kesiswaanTestUUID(26)
	validDate := kesiswaanTestDate()
	expectedErr := errors.New("store failed")

	store := &fakeKesiswaanStore{statsErr: expectedErr}
	svc := &Kesiswaan{q: store, photoDir: t.TempDir()}
	if _, err := svc.Stats(ctx, pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Stats(admin error) = %v, want %v", err, expectedErr)
	}

	store = &fakeKesiswaanStore{teacherStatsErr: expectedErr}
	svc = &Kesiswaan{q: store, photoDir: t.TempDir()}
	if _, err := svc.Stats(ctx, teacherID); !errors.Is(err, expectedErr) {
		t.Fatalf("Stats(teacher error) = %v, want %v", err, expectedErr)
	}

	store = &fakeKesiswaanStore{listStudentsErr: expectedErr}
	svc = &Kesiswaan{q: store, photoDir: t.TempDir()}
	if _, err := svc.ListStudents(ctx, "", "active", "", pgtype.UUID{}); !errors.Is(err, expectedErr) {
		t.Fatalf("ListStudents(store error) = %v, want %v", err, expectedErr)
	}
	if _, err := svc.ListStudents(ctx, "", "", "bad-id", pgtype.UUID{}); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListStudents(bad class id) = %v, want id tidak valid", err)
	}

	store = &fakeKesiswaanStore{profileErr: expectedErr}
	svc = &Kesiswaan{q: store, photoDir: t.TempDir()}
	if _, err := svc.UpdateStudentProfile(ctx, db.UpdateKesiswaanStudentProfileParams{Nik: "123"}); err == nil || err.Error() != "NIK harus 16 digit" {
		t.Fatalf("UpdateStudentProfile(validation) = %v, want NIK harus 16 digit", err)
	}
	if _, err := svc.UpdateStudentProfile(ctx, db.UpdateKesiswaanStudentProfileParams{
		ID:           studentID,
		Nik:          "1234567890123456",
		TempatLahir:  "Kolaka",
		TanggalLahir: validDate,
		Alamat:       "Lasusua",
		Agama:        "Islam",
		AnakKe:       pgtype.Int4{Int32: 1, Valid: true},
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateStudentProfile(store error) = %v, want %v", err, expectedErr)
	}

	copyErr := errors.New("copy failed")
	store = &fakeKesiswaanStore{}
	svc = &Kesiswaan{q: store, photoDir: t.TempDir()}
	if _, err := svc.SaveStudentPhoto(ctx, UploadStudentPhotoInput{
		StudentID:    studentID,
		OriginalName: "foto.png",
		MimeType:     "image/png",
		FileSize:     10,
		File:         failingReader{err: copyErr},
	}); !errors.Is(err, copyErr) {
		t.Fatalf("SaveStudentPhoto(copy error) = %v, want %v", err, copyErr)
	}

	if _, err := svc.SaveStudentPhoto(ctx, UploadStudentPhotoInput{}); err == nil || err.Error() != "siswa wajib dipilih" {
		t.Fatalf("SaveStudentPhoto(validation) = %v, want siswa wajib dipilih", err)
	}

	blockingFile := filepath.Join(t.TempDir(), "photo-dir-file")
	if err := os.WriteFile(blockingFile, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}
	svc = &Kesiswaan{q: &fakeKesiswaanStore{}, photoDir: filepath.Join(blockingFile, "photos")}
	if _, err := svc.SaveStudentPhoto(ctx, UploadStudentPhotoInput{
		StudentID:    studentID,
		OriginalName: "foto.png",
		MimeType:     "image/png",
		FileSize:     int64(len("png")),
		File:         strings.NewReader("png"),
	}); err == nil {
		t.Fatal("SaveStudentPhoto(mkdir error) error = nil, want directory creation error")
	}

	readOnlyDir := t.TempDir()
	if err := os.Chmod(readOnlyDir, 0o555); err != nil {
		t.Fatalf("chmod readonly photo dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(readOnlyDir, 0o755)
	})
	svc = &Kesiswaan{q: &fakeKesiswaanStore{}, photoDir: readOnlyDir}
	if _, err := svc.SaveStudentPhoto(ctx, UploadStudentPhotoInput{
		StudentID:    studentID,
		OriginalName: "foto.png",
		MimeType:     "image/png",
		FileSize:     int64(len("png")),
		File:         strings.NewReader("png"),
	}); err == nil {
		t.Fatal("SaveStudentPhoto(create error) error = nil, want file creation error")
	}

	store = &fakeKesiswaanStore{photoErr: expectedErr}
	svc = &Kesiswaan{q: store, photoDir: t.TempDir()}
	if _, err := svc.SaveStudentPhoto(ctx, UploadStudentPhotoInput{
		StudentID:    studentID,
		OriginalName: "foto.png",
		MimeType:     "image/png",
		FileSize:     int64(len("png")),
		File:         strings.NewReader("png"),
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("SaveStudentPhoto(store error) = %v, want %v", err, expectedErr)
	}

	store = &fakeKesiswaanStore{canReadErr: expectedErr}
	svc = &Kesiswaan{q: store, photoDir: t.TempDir()}
	if ok, err := svc.CanReadStudentPhoto(ctx, "foto.png", teacherID); ok || !errors.Is(err, expectedErr) {
		t.Fatalf("CanReadStudentPhoto(error) = %v/%v, want false/%v", ok, err, expectedErr)
	}
}

func TestKesiswaanMutationsNormalizeForwardAndDelete(t *testing.T) {
	store := &fakeKesiswaanStore{canRead: true}
	svc := &Kesiswaan{q: store, photoDir: t.TempDir()}
	id := kesiswaanTestUUID(31)
	studentID := kesiswaanTestUUID(32)
	categoryID := kesiswaanTestUUID(33)
	extraID := kesiswaanTestUUID(34)
	validDate := kesiswaanTestDate()

	if _, err := svc.UpdateStudentProfile(context.Background(), db.UpdateKesiswaanStudentProfileParams{
		ID:           studentID,
		Nik:          "  1234567890123456  ",
		TempatLahir:  "  Kolaka  ",
		TanggalLahir: validDate,
		Alamat:       "  Lasusua  ",
		Agama:        "  Islam  ",
		AnakKe:       pgtype.Int4{Int32: 1, Valid: true},
		Phone:        "  0812  ",
		ParentName:   "  Orang Tua  ",
		ParentPhone:  "  0821  ",
	}); err != nil {
		t.Fatalf("UpdateStudentProfile() error = %v", err)
	}
	if store.profileArg.Nik != "1234567890123456" || store.profileArg.TempatLahir != "Kolaka" || store.profileArg.ParentName != "Orang Tua" {
		t.Fatalf("UpdateStudentProfile() arg = %+v, want trimmed fields", store.profileArg)
	}

	photo, err := svc.SaveStudentPhoto(context.Background(), UploadStudentPhotoInput{
		StudentID:    studentID,
		OriginalName: "../Siswa A.png",
		MimeType:     "image/png",
		FileSize:     int64(len("png")),
		File:         strings.NewReader("png"),
	})
	if err != nil {
		t.Fatalf("SaveStudentPhoto() error = %v", err)
	}
	if photo.ID != studentID || store.photoArg.ID != studentID || !strings.HasPrefix(store.photoArg.PhotoUrl, "/api/kesiswaan/student-photos/") || !strings.HasSuffix(store.photoArg.PhotoUrl, "Siswa_A.png") {
		t.Fatalf("SaveStudentPhoto() photo/arg = %+v/%+v, want stored API URL", photo, store.photoArg)
	}
	canRead, err := svc.CanReadStudentPhoto(context.Background(), strings.TrimPrefix(store.photoArg.PhotoUrl, "/api/kesiswaan/student-photos/"), kesiswaanTestUUID(35))
	if err != nil || !canRead {
		t.Fatalf("CanReadStudentPhoto() = %v, %v; want true nil", canRead, err)
	}
	if store.canReadArg.PhotoUrl != store.photoArg.PhotoUrl {
		t.Fatalf("CanReadStudentPhoto() arg = %+v, want photo url %q", store.canReadArg, store.photoArg.PhotoUrl)
	}

	if _, err := svc.CreateCategory(context.Background(), db.CreateViolationCategoryParams{Code: " T01 ", Name: " Telat ", Point: 5, Severity: " RINGAN ", Description: " Terlambat ", IsActive: true}); err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}
	if store.categoryCreate.Code != "T01" || store.categoryCreate.Name != "Telat" || store.categoryCreate.Severity != "ringan" || store.categoryCreate.Description != "Terlambat" {
		t.Fatalf("CreateCategory() arg = %+v, want normalized", store.categoryCreate)
	}
	if _, err := svc.UpdateCategory(context.Background(), db.UpdateViolationCategoryParams{ID: categoryID, Code: " T02 ", Name: " Atribut ", Point: 3, Severity: "SEDANG", Description: " Rapikan ", IsActive: true}); err != nil {
		t.Fatalf("UpdateCategory() error = %v", err)
	}
	if store.categoryUpdate.Code != "T02" || store.categoryUpdate.Severity != "sedang" {
		t.Fatalf("UpdateCategory() arg = %+v, want normalized", store.categoryUpdate)
	}
	if err := svc.DeleteCategory(context.Background(), categoryID); err != nil || store.categoryDelete != categoryID {
		t.Fatalf("DeleteCategory() = %v, id %v; want nil/%v", err, store.categoryDelete, categoryID)
	}

	if _, err := svc.CreateViolation(context.Background(), db.CreateStudentViolationParams{StudentID: studentID, CategoryID: categoryID, IncidentDate: validDate, Points: 5, Description: "  Datang terlambat  ", ActionTaken: "  Pembinaan  "}); err != nil {
		t.Fatalf("CreateViolation() error = %v", err)
	}
	if store.violationCreate.Description != "Datang terlambat" || store.violationCreate.ActionTaken != "Pembinaan" || store.violationCreate.Status != "open" {
		t.Fatalf("CreateViolation() arg = %+v, want normalized", store.violationCreate)
	}
	if _, err := svc.UpdateViolation(context.Background(), db.UpdateStudentViolationParams{ID: id, StudentID: studentID, CategoryID: categoryID, IncidentDate: validDate, Points: 4, Description: "  sudah dibina  ", ActionTaken: "  wali kelas  ", Status: "RESOLVED"}); err != nil {
		t.Fatalf("UpdateViolation() error = %v", err)
	}
	if store.violationUpdate.Description != "sudah dibina" || store.violationUpdate.ActionTaken != "wali kelas" || store.violationUpdate.Status != "resolved" {
		t.Fatalf("UpdateViolation() arg = %+v, want normalized", store.violationUpdate)
	}
	if err := svc.DeleteViolation(context.Background(), id); err != nil || store.violationDelete != id {
		t.Fatalf("DeleteViolation() = %v, id %v; want nil/%v", err, store.violationDelete, id)
	}

	if _, err := svc.CreateAchievement(context.Background(), db.CreateStudentAchievementParams{StudentID: studentID, AchievementDate: validDate, Title: " Juara 1 ", Level: "SCHOOL", Category: " MTQ ", Organizer: " Kemenag ", Description: "  baik  ", DocumentUrl: " /doc.pdf "}); err != nil {
		t.Fatalf("CreateAchievement() error = %v", err)
	}
	if store.achievementCreate.Title != "Juara 1" || store.achievementCreate.Level != "school" || store.achievementCreate.DocumentUrl != "/doc.pdf" {
		t.Fatalf("CreateAchievement() arg = %+v, want normalized", store.achievementCreate)
	}
	if _, err := svc.UpdateAchievement(context.Background(), db.UpdateStudentAchievementParams{ID: id, StudentID: studentID, AchievementDate: validDate, Title: " Juara 2 ", Level: "DISTRICT", Category: " Sains ", Organizer: " KKM ", Description: "  final  ", DocumentUrl: " /piagam.pdf "}); err != nil {
		t.Fatalf("UpdateAchievement() error = %v", err)
	}
	if store.achievementUpdate.Title != "Juara 2" || store.achievementUpdate.Level != "district" || store.achievementUpdate.DocumentUrl != "/piagam.pdf" {
		t.Fatalf("UpdateAchievement() arg = %+v, want normalized", store.achievementUpdate)
	}
	if err := svc.DeleteAchievement(context.Background(), id); err != nil || store.achievementDelete != id {
		t.Fatalf("DeleteAchievement() = %v, id %v; want nil/%v", err, store.achievementDelete, id)
	}

	if _, err := svc.CreateExtracurricular(context.Background(), db.CreateExtracurricularParams{Code: " PRM ", Name: " Pramuka ", Category: " Wajib ", Description: "  latihan  ", ScheduleText: " Jumat ", IsActive: true}); err != nil {
		t.Fatalf("CreateExtracurricular() error = %v", err)
	}
	if store.extraCreate.Code != "PRM" || store.extraCreate.Name != "Pramuka" || store.extraCreate.ScheduleText != "Jumat" {
		t.Fatalf("CreateExtracurricular() arg = %+v, want normalized", store.extraCreate)
	}
	if _, err := svc.UpdateExtracurricular(context.Background(), db.UpdateExtracurricularParams{ID: extraID, Code: " PMR ", Name: " PMR ", Category: " Pilihan ", Description: "  kesehatan  ", ScheduleText: " Sabtu ", IsActive: true}); err != nil {
		t.Fatalf("UpdateExtracurricular() error = %v", err)
	}
	if store.extraUpdate.Code != "PMR" || store.extraUpdate.Category != "Pilihan" || store.extraUpdate.ScheduleText != "Sabtu" {
		t.Fatalf("UpdateExtracurricular() arg = %+v, want normalized", store.extraUpdate)
	}
	if err := svc.DeleteExtracurricular(context.Background(), extraID); err != nil || store.extraDelete != extraID {
		t.Fatalf("DeleteExtracurricular() = %v, id %v; want nil/%v", err, store.extraDelete, extraID)
	}

	if _, err := svc.CreateExtracurricularMember(context.Background(), db.CreateExtracurricularMemberParams{ExtracurricularID: extraID, StudentID: studentID, JoinedAt: validDate, Role: "", Status: "", Notes: "  aktif  "}); err != nil {
		t.Fatalf("CreateExtracurricularMember() error = %v", err)
	}
	if store.memberCreate.Role != "member" || store.memberCreate.Status != "active" || store.memberCreate.Notes != "aktif" {
		t.Fatalf("CreateExtracurricularMember() arg = %+v, want normalized", store.memberCreate)
	}
	if _, err := svc.UpdateExtracurricularMember(context.Background(), db.UpdateExtracurricularMemberParams{ID: id, ExtracurricularID: extraID, StudentID: studentID, JoinedAt: validDate, Role: "LEADER", Status: "INACTIVE", Notes: "  nonaktif  "}); err != nil {
		t.Fatalf("UpdateExtracurricularMember() error = %v", err)
	}
	if store.memberUpdate.Role != "leader" || store.memberUpdate.Status != "inactive" || store.memberUpdate.Notes != "nonaktif" {
		t.Fatalf("UpdateExtracurricularMember() arg = %+v, want normalized", store.memberUpdate)
	}
	if err := svc.DeleteExtracurricularMember(context.Background(), id); err != nil || store.memberDelete != id {
		t.Fatalf("DeleteExtracurricularMember() = %v, id %v; want nil/%v", err, store.memberDelete, id)
	}

	if _, err := svc.CreateCounselingSession(context.Background(), db.CreateCounselingSessionParams{StudentID: studentID, SessionDate: validDate, Topic: " Belajar ", Summary: "  ringkas  ", FollowUp: "  pantau  ", Status: ""}); err != nil {
		t.Fatalf("CreateCounselingSession() error = %v", err)
	}
	if store.counselingCreate.Topic != "Belajar" || store.counselingCreate.Summary != "ringkas" || store.counselingCreate.Status != "open" {
		t.Fatalf("CreateCounselingSession() arg = %+v, want normalized", store.counselingCreate)
	}
	if _, err := svc.UpdateCounselingSession(context.Background(), db.UpdateCounselingSessionParams{ID: id, StudentID: studentID, SessionDate: validDate, Topic: " Lanjutan ", Summary: "  ringkas 2  ", FollowUp: "  selesai  ", Status: "RESOLVED"}); err != nil {
		t.Fatalf("UpdateCounselingSession() error = %v", err)
	}
	if store.counselingUpdate.Topic != "Lanjutan" || store.counselingUpdate.Summary != "ringkas 2" || store.counselingUpdate.Status != "resolved" {
		t.Fatalf("UpdateCounselingSession() arg = %+v, want normalized", store.counselingUpdate)
	}
	if err := svc.DeleteCounselingSession(context.Background(), id); err != nil || store.counselingDelete != id {
		t.Fatalf("DeleteCounselingSession() = %v, id %v; want nil/%v", err, store.counselingDelete, id)
	}
}

func TestKesiswaanMutationValidationBranches(t *testing.T) {
	svc := &Kesiswaan{q: &fakeKesiswaanStore{}, photoDir: t.TempDir()}
	id := kesiswaanTestUUID(36)
	studentID := kesiswaanTestUUID(37)
	extraID := kesiswaanTestUUID(38)
	validDate := kesiswaanTestDate()

	if _, err := svc.CreateCategory(context.Background(), db.CreateViolationCategoryParams{Name: "Telat", Point: 1}); err == nil || err.Error() != "kode kategori wajib diisi" {
		t.Fatalf("CreateCategory(invalid) = %v, want code validation", err)
	}
	if _, err := svc.UpdateCategory(context.Background(), db.UpdateViolationCategoryParams{ID: id, Code: "T01", Name: "Telat", Point: 1, Severity: "parah"}); err == nil || err.Error() != "tingkat pelanggaran tidak valid" {
		t.Fatalf("UpdateCategory(invalid) = %v, want severity validation", err)
	}
	if _, err := svc.CreateViolation(context.Background(), db.CreateStudentViolationParams{CategoryID: id, IncidentDate: validDate, Points: 1}); err == nil || err.Error() != "siswa wajib dipilih" {
		t.Fatalf("CreateViolation(invalid) = %v, want student validation", err)
	}
	if _, err := svc.UpdateViolation(context.Background(), db.UpdateStudentViolationParams{ID: id, StudentID: studentID, IncidentDate: validDate, Points: 1, Status: "pending"}); err == nil || err.Error() != "status pelanggaran tidak valid" {
		t.Fatalf("UpdateViolation(invalid) = %v, want status validation", err)
	}
	if _, err := svc.CreateAchievement(context.Background(), db.CreateStudentAchievementParams{StudentID: studentID, AchievementDate: validDate, Level: "school"}); err == nil || err.Error() != "judul prestasi wajib diisi" {
		t.Fatalf("CreateAchievement(invalid) = %v, want title validation", err)
	}
	if _, err := svc.UpdateAchievement(context.Background(), db.UpdateStudentAchievementParams{ID: id, StudentID: studentID, AchievementDate: validDate, Title: "Juara", Level: "desa"}); err == nil || err.Error() != "tingkat prestasi tidak valid" {
		t.Fatalf("UpdateAchievement(invalid) = %v, want level validation", err)
	}
	if _, err := svc.CreateExtracurricular(context.Background(), db.CreateExtracurricularParams{Name: "Pramuka"}); err == nil || err.Error() != "kode ekskul wajib diisi" {
		t.Fatalf("CreateExtracurricular(invalid) = %v, want code validation", err)
	}
	if _, err := svc.UpdateExtracurricular(context.Background(), db.UpdateExtracurricularParams{ID: extraID, Code: "PRM"}); err == nil || err.Error() != "nama ekskul wajib diisi" {
		t.Fatalf("UpdateExtracurricular(invalid) = %v, want name validation", err)
	}
	if _, err := svc.CreateExtracurricularMember(context.Background(), db.CreateExtracurricularMemberParams{ExtracurricularID: extraID, StudentID: studentID, JoinedAt: validDate, Role: "coach", Status: "active"}); err == nil || err.Error() != "peran anggota ekskul tidak valid" {
		t.Fatalf("CreateExtracurricularMember(invalid) = %v, want role validation", err)
	}
	if _, err := svc.UpdateExtracurricularMember(context.Background(), db.UpdateExtracurricularMemberParams{ID: id, ExtracurricularID: extraID, StudentID: studentID, JoinedAt: validDate, Role: "member", Status: "pending"}); err == nil || err.Error() != "status anggota ekskul tidak valid" {
		t.Fatalf("UpdateExtracurricularMember(invalid) = %v, want status validation", err)
	}
	if _, err := svc.CreateCounselingSession(context.Background(), db.CreateCounselingSessionParams{StudentID: studentID, SessionDate: validDate}); err == nil || err.Error() != "topik konseling wajib diisi" {
		t.Fatalf("CreateCounselingSession(invalid) = %v, want topic validation", err)
	}
	if _, err := svc.UpdateCounselingSession(context.Background(), db.UpdateCounselingSessionParams{ID: id, StudentID: studentID, SessionDate: validDate, Topic: "Belajar", Status: "pending"}); err == nil || err.Error() != "status konseling tidak valid" {
		t.Fatalf("UpdateCounselingSession(invalid) = %v, want status validation", err)
	}
	if _, err := svc.CreateStudentTransfer(context.Background(), db.CreateStudentTransferParams{TransferType: "out", TransferDate: validDate, DestinationSchool: "MTs Baru"}); err == nil || err.Error() != "siswa wajib dipilih" {
		t.Fatalf("CreateStudentTransfer(invalid) = %v, want student validation", err)
	}
}

func TestKesiswaanCreateStudentTransferRequiresPoolAfterValidation(t *testing.T) {
	svc := &Kesiswaan{q: &fakeKesiswaanStore{}, photoDir: t.TempDir()}

	_, err := svc.CreateStudentTransfer(context.Background(), db.CreateStudentTransferParams{
		StudentID:         kesiswaanTestUUID(41),
		TransferDate:      kesiswaanTestDate(),
		TransferType:      "OUT",
		DestinationSchool: " MTs Baru ",
		Reason:            "  pindah domisili  ",
		DocumentRef:       " SK-1 ",
		Notes:             "  lengkap  ",
	})
	if err == nil || err.Error() != "layanan mutasi siswa belum siap" {
		t.Fatalf("CreateStudentTransfer() error = %v, want service not ready", err)
	}
}

func TestKesiswaanCreateStudentTransferValidationErrorsBeforePool(t *testing.T) {
	svc := &Kesiswaan{q: &fakeKesiswaanStore{}, photoDir: t.TempDir()}
	studentID := kesiswaanTestUUID(44)
	validDate := kesiswaanTestDate()

	tests := []struct {
		name    string
		arg     db.CreateStudentTransferParams
		wantErr string
	}{
		{name: "missing student", arg: db.CreateStudentTransferParams{TransferDate: validDate, TransferType: "out", DestinationSchool: "MTs Baru"}, wantErr: "siswa wajib dipilih"},
		{name: "invalid type", arg: db.CreateStudentTransferParams{StudentID: studentID, TransferDate: validDate, TransferType: "stay"}, wantErr: "jenis mutasi tidak valid"},
		{name: "out missing destination after trim", arg: db.CreateStudentTransferParams{StudentID: studentID, TransferDate: validDate, TransferType: " OUT ", DestinationSchool: "   "}, wantErr: "sekolah tujuan wajib diisi untuk mutasi keluar"},
		{name: "in missing previous school after trim", arg: db.CreateStudentTransferParams{StudentID: studentID, TransferDate: validDate, TransferType: " IN ", PreviousSchool: "   "}, wantErr: "sekolah asal wajib diisi untuk mutasi masuk"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateStudentTransfer(context.Background(), tt.arg)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("CreateStudentTransfer() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestCreateStudentTransferUpdatesStudentLifecycle(t *testing.T) {
	studentID := kesiswaanTestUUID(42)
	validDate := kesiswaanTestDate()

	tests := []struct {
		name         string
		transferType string
		wantStatus   db.StudentStatusEnum
		wantActive   bool
	}{
		{name: "out transfer marks student mutated", transferType: "out", wantStatus: db.StudentStatusEnumMutated, wantActive: false},
		{name: "in transfer keeps student active", transferType: "in", wantStatus: db.StudentStatusEnumActive, wantActive: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeKesiswaanStore{}
			row, err := createStudentTransfer(context.Background(), store, db.CreateStudentTransferParams{
				StudentID:         studentID,
				TransferDate:      validDate,
				TransferType:      tt.transferType,
				PreviousSchool:    "MTs Lama",
				DestinationSchool: "MTs Baru",
			})
			if err != nil {
				t.Fatalf("createStudentTransfer() error = %v", err)
			}
			if row.StudentID != studentID || store.transferCreate.TransferType != tt.transferType {
				t.Fatalf("transfer row/arg = %+v/%+v, want forwarded transfer", row, store.transferCreate)
			}
			if store.lifecycleArg.ID != studentID || store.lifecycleArg.Status != tt.wantStatus || store.lifecycleArg.IsActive != tt.wantActive {
				t.Fatalf("lifecycle arg = %+v, want %s active=%v", store.lifecycleArg, tt.wantStatus, tt.wantActive)
			}
		})
	}
}

func TestCreateStudentTransferPropagatesStoreErrors(t *testing.T) {
	studentID := kesiswaanTestUUID(43)
	validDate := kesiswaanTestDate()
	arg := db.CreateStudentTransferParams{
		StudentID:         studentID,
		TransferDate:      validDate,
		TransferType:      "out",
		DestinationSchool: "MTs Baru",
	}

	store := &fakeKesiswaanStore{transferErr: errors.New("create failed")}
	if _, err := createStudentTransfer(context.Background(), store, arg); err == nil || err.Error() != "create failed" {
		t.Fatalf("createStudentTransfer(create error) = %v, want create failed", err)
	}

	store = &fakeKesiswaanStore{lifecycleErr: errors.New("lifecycle failed")}
	if _, err := createStudentTransfer(context.Background(), store, arg); err == nil || err.Error() != "lifecycle failed" {
		t.Fatalf("createStudentTransfer(lifecycle error) = %v, want lifecycle failed", err)
	}
}

func TestKesiswaanNormalizeCreateParams(t *testing.T) {
	violation := normalizeViolationCategoryCreate(db.CreateViolationCategoryParams{
		Code:        "  T01  ",
		Name:        "  Terlambat  ",
		Severity:    "",
		Description: "  Masuk terlambat  ",
	})
	if violation.Code != "T01" || violation.Name != "Terlambat" || violation.Severity != "ringan" || violation.Description != "Masuk terlambat" {
		t.Fatalf("normalizeViolationCategoryCreate() = %+v, want trimmed defaults", violation)
	}

	studentViolation := normalizeStudentViolationCreate(db.CreateStudentViolationParams{
		Description: "  Catatan  ",
		ActionTaken: "  Pembinaan  ",
	})
	if studentViolation.Description != "Catatan" || studentViolation.ActionTaken != "Pembinaan" || studentViolation.Status != "open" || !studentViolation.IncidentDate.Valid {
		t.Fatalf("normalizeStudentViolationCreate() = %+v, want defaults", studentViolation)
	}

	achievement := normalizeStudentAchievementCreate(db.CreateStudentAchievementParams{
		Title:       "  Juara 1  ",
		Category:    "  MTQ  ",
		Organizer:   "  Kemenag  ",
		Description: "  Tingkat kabupaten  ",
		DocumentUrl: "  /doc.pdf  ",
	})
	if achievement.Title != "Juara 1" || achievement.Level != "school" || achievement.Category != "MTQ" || achievement.Organizer != "Kemenag" || achievement.Description != "Tingkat kabupaten" || achievement.DocumentUrl != "/doc.pdf" || !achievement.AchievementDate.Valid {
		t.Fatalf("normalizeStudentAchievementCreate() = %+v, want trimmed defaults", achievement)
	}

	extra := normalizeExtracurricularCreate(db.CreateExtracurricularParams{
		Code:         "  PRM  ",
		Name:         "  Pramuka  ",
		Category:     "  Wajib  ",
		Description:  "  Ekskul wajib  ",
		ScheduleText: "  Jumat  ",
	})
	if extra.Code != "PRM" || extra.Name != "Pramuka" || extra.Category != "Wajib" || extra.Description != "Ekskul wajib" || extra.ScheduleText != "Jumat" {
		t.Fatalf("normalizeExtracurricularCreate() = %+v, want trimmed fields", extra)
	}

	member := normalizeExtracurricularMemberCreate(db.CreateExtracurricularMemberParams{
		Role:  "",
		Notes: "  Aktif  ",
	})
	if member.Role != "member" || member.Status != "active" || member.Notes != "Aktif" || !member.JoinedAt.Valid {
		t.Fatalf("normalizeExtracurricularMemberCreate() = %+v, want defaults", member)
	}

	counseling := normalizeCounselingCreate(db.CreateCounselingSessionParams{
		Topic:    "  Belajar  ",
		Summary:  "  Ringkas  ",
		FollowUp: "  Pantau  ",
	})
	if counseling.Topic != "Belajar" || counseling.Summary != "Ringkas" || counseling.FollowUp != "Pantau" || counseling.Status != "open" || !counseling.SessionDate.Valid {
		t.Fatalf("normalizeCounselingCreate() = %+v, want trimmed defaults", counseling)
	}

	transfer := normalizeStudentTransferCreate(db.CreateStudentTransferParams{
		TransferType:      "",
		PreviousSchool:    "  MTs Lama  ",
		DestinationSchool: "  MTs Baru  ",
		Reason:            "  Pindah domisili  ",
		DocumentRef:       "  DOC-1  ",
		Notes:             "  Lengkap  ",
	})
	if transfer.TransferType != "out" || transfer.PreviousSchool != "MTs Lama" || transfer.DestinationSchool != "MTs Baru" || transfer.Reason != "Pindah domisili" || transfer.DocumentRef != "DOC-1" || transfer.Notes != "Lengkap" || !transfer.TransferDate.Valid {
		t.Fatalf("normalizeStudentTransferCreate() = %+v, want trimmed defaults", transfer)
	}
}

func TestKesiswaanValidationHelpers(t *testing.T) {
	validID := kesiswaanTestUUID(1)
	validDate := kesiswaanTestDate()

	tests := []struct {
		name    string
		err     error
		wantErr string
	}{
		{name: "student profile valid empty nik", err: validateKesiswaanStudentProfile(db.UpdateKesiswaanStudentProfileParams{})},
		{name: "student profile invalid nik length", err: validateKesiswaanStudentProfile(db.UpdateKesiswaanStudentProfileParams{Nik: "123"}), wantErr: "NIK harus 16 digit"},
		{name: "student profile invalid nik chars", err: validateKesiswaanStudentProfile(db.UpdateKesiswaanStudentProfileParams{Nik: "12345678901234AB"}), wantErr: "NIK hanya boleh berisi angka"},
		{name: "student profile invalid child order", err: validateKesiswaanStudentProfile(db.UpdateKesiswaanStudentProfileParams{AnakKe: pgtype.Int4{Int32: 0, Valid: true}}), wantErr: "anak ke harus lebih dari 0"},
		{name: "photo valid", err: validateStudentPhoto(UploadStudentPhotoInput{StudentID: validID, File: strings.NewReader("img"), FileSize: 3, MimeType: "image/png"})},
		{name: "photo missing student", err: validateStudentPhoto(UploadStudentPhotoInput{File: strings.NewReader("img"), FileSize: 3, MimeType: "image/png"}), wantErr: "siswa wajib dipilih"},
		{name: "photo missing file", err: validateStudentPhoto(UploadStudentPhotoInput{StudentID: validID, FileSize: 3, MimeType: "image/png"}), wantErr: "file wajib diisi"},
		{name: "photo invalid size", err: validateStudentPhoto(UploadStudentPhotoInput{StudentID: validID, File: strings.NewReader("img"), FileSize: 0, MimeType: "image/png"}), wantErr: "ukuran file tidak valid"},
		{name: "photo too large", err: validateStudentPhoto(UploadStudentPhotoInput{StudentID: validID, File: strings.NewReader("img"), FileSize: 3*1024*1024 + 1, MimeType: "image/png"}), wantErr: "ukuran foto maksimal 3MB"},
		{name: "photo non image", err: validateStudentPhoto(UploadStudentPhotoInput{StudentID: validID, File: strings.NewReader("img"), FileSize: 3, MimeType: "application/pdf"}), wantErr: "hanya file gambar yang diperbolehkan"},
		{name: "violation category valid", err: validateViolationCategory("T01", "Terlambat", 5, "ringan")},
		{name: "violation category empty code", err: validateViolationCategory("", "Terlambat", 5, "ringan"), wantErr: "kode kategori wajib diisi"},
		{name: "violation category empty name", err: validateViolationCategory("T01", "", 5, "ringan"), wantErr: "nama kategori wajib diisi"},
		{name: "violation category negative point", err: validateViolationCategory("T01", "Terlambat", -1, "ringan"), wantErr: "poin kategori tidak valid"},
		{name: "violation category invalid severity", err: validateViolationCategory("T01", "Terlambat", 5, "sedikit"), wantErr: "tingkat pelanggaran tidak valid"},
		{name: "student violation valid", err: validateStudentViolation(validID, validDate, 5, "open")},
		{name: "student violation missing student", err: validateStudentViolation(pgtype.UUID{}, validDate, 5, "open"), wantErr: "siswa wajib dipilih"},
		{name: "student violation missing date", err: validateStudentViolation(validID, pgtype.Date{}, 5, "open"), wantErr: "tanggal kejadian wajib diisi"},
		{name: "student violation negative points", err: validateStudentViolation(validID, validDate, -1, "open"), wantErr: "poin pelanggaran tidak valid"},
		{name: "student violation invalid status", err: validateStudentViolation(validID, validDate, 5, "pending"), wantErr: "status pelanggaran tidak valid"},
		{name: "achievement valid", err: validateStudentAchievement(validID, validDate, "Juara", "school")},
		{name: "achievement missing student", err: validateStudentAchievement(pgtype.UUID{}, validDate, "Juara", "school"), wantErr: "siswa wajib dipilih"},
		{name: "achievement missing date", err: validateStudentAchievement(validID, pgtype.Date{}, "Juara", "school"), wantErr: "tanggal prestasi wajib diisi"},
		{name: "achievement missing title", err: validateStudentAchievement(validID, validDate, "", "school"), wantErr: "judul prestasi wajib diisi"},
		{name: "achievement invalid level", err: validateStudentAchievement(validID, validDate, "Juara", "desa"), wantErr: "tingkat prestasi tidak valid"},
		{name: "extracurricular valid", err: validateExtracurricular("PRM", "Pramuka")},
		{name: "extracurricular empty code", err: validateExtracurricular("", "Pramuka"), wantErr: "kode ekskul wajib diisi"},
		{name: "extracurricular empty name", err: validateExtracurricular("PRM", ""), wantErr: "nama ekskul wajib diisi"},
		{name: "member valid", err: validateExtracurricularMember(validID, validID, validDate, "leader", "active")},
		{name: "member missing extracurricular", err: validateExtracurricularMember(pgtype.UUID{}, validID, validDate, "leader", "active"), wantErr: "ekskul wajib dipilih"},
		{name: "member missing student", err: validateExtracurricularMember(validID, pgtype.UUID{}, validDate, "leader", "active"), wantErr: "siswa wajib dipilih"},
		{name: "member missing joined date", err: validateExtracurricularMember(validID, validID, pgtype.Date{}, "leader", "active"), wantErr: "tanggal bergabung wajib diisi"},
		{name: "member invalid role", err: validateExtracurricularMember(validID, validID, validDate, "coach", "active"), wantErr: "peran anggota ekskul tidak valid"},
		{name: "member invalid status", err: validateExtracurricularMember(validID, validID, validDate, "leader", "pending"), wantErr: "status anggota ekskul tidak valid"},
		{name: "counseling valid", err: validateCounselingSession(validID, validDate, "Belajar", "open")},
		{name: "counseling missing student", err: validateCounselingSession(pgtype.UUID{}, validDate, "Belajar", "open"), wantErr: "siswa wajib dipilih"},
		{name: "counseling missing date", err: validateCounselingSession(validID, pgtype.Date{}, "Belajar", "open"), wantErr: "tanggal konseling wajib diisi"},
		{name: "counseling missing topic", err: validateCounselingSession(validID, validDate, "", "open"), wantErr: "topik konseling wajib diisi"},
		{name: "counseling invalid status", err: validateCounselingSession(validID, validDate, "Belajar", "pending"), wantErr: "status konseling tidak valid"},
		{name: "transfer valid out", err: validateStudentTransfer(db.CreateStudentTransferParams{StudentID: validID, TransferDate: validDate, TransferType: "out", DestinationSchool: "MTs Baru"})},
		{name: "transfer valid in", err: validateStudentTransfer(db.CreateStudentTransferParams{StudentID: validID, TransferDate: validDate, TransferType: "in", PreviousSchool: "MTs Lama"})},
		{name: "transfer missing student", err: validateStudentTransfer(db.CreateStudentTransferParams{TransferDate: validDate, TransferType: "out", DestinationSchool: "MTs Baru"}), wantErr: "siswa wajib dipilih"},
		{name: "transfer missing date", err: validateStudentTransfer(db.CreateStudentTransferParams{StudentID: validID, TransferType: "out", DestinationSchool: "MTs Baru"}), wantErr: "tanggal mutasi wajib diisi"},
		{name: "transfer invalid type", err: validateStudentTransfer(db.CreateStudentTransferParams{StudentID: validID, TransferDate: validDate, TransferType: "stay"}), wantErr: "jenis mutasi tidak valid"},
		{name: "transfer out missing destination", err: validateStudentTransfer(db.CreateStudentTransferParams{StudentID: validID, TransferDate: validDate, TransferType: "out"}), wantErr: "sekolah tujuan wajib diisi untuk mutasi keluar"},
		{name: "transfer in missing previous school", err: validateStudentTransfer(db.CreateStudentTransferParams{StudentID: validID, TransferDate: validDate, TransferType: "in"}), wantErr: "sekolah asal wajib diisi untuk mutasi masuk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr == "" {
				if tt.err != nil {
					t.Fatalf("validation error = %v", tt.err)
				}
				return
			}
			if tt.err == nil || tt.err.Error() != tt.wantErr {
				t.Fatalf("validation error = %v, want %q", tt.err, tt.wantErr)
			}
		})
	}
}

func TestKesiswaanPhotoPathGuardsUnsafeFilenames(t *testing.T) {
	svc := &Kesiswaan{photoDir: filepath.Join("data", "student-photos")}

	path, ok := svc.StudentPhotoPath(" siswa.jpg ")
	if !ok {
		t.Fatal("StudentPhotoPath(valid) ok = false, want true")
	}
	if path != filepath.Join("data", "student-photos", "siswa.jpg") {
		t.Fatalf("StudentPhotoPath(valid) = %q, want joined path", path)
	}

	for _, filename := range []string{"", "../siswa.jpg", "kelas/siswa.jpg", "siswa..jpg"} {
		if path, ok := svc.StudentPhotoPath(filename); ok || path != "" {
			t.Fatalf("StudentPhotoPath(%q) = %q, %v; want empty false", filename, path, ok)
		}
		if canRead, err := svc.CanReadStudentPhoto(context.Background(), filename, pgtype.UUID{}); err != nil || canRead {
			t.Fatalf("CanReadStudentPhoto(%q) = %v, %v; want false nil", filename, canRead, err)
		}
	}
}
