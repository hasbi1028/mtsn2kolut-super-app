package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestStudentAccountGenerationPreviewBuildsDeterministicUsernames(t *testing.T) {
	studentID := testGenerationUUID(11)
	existingStudentID := testGenerationUUID(12)
	duplicateAID := testGenerationUUID(13)
	duplicateBID := testGenerationUUID(14)
	missingBasisID := testGenerationUUID(15)
	existingUserID := testGenerationUUID(91)

	store := &fakeStudentAccountGenerationStore{
		studentRows: []db.ListStudentAccountGenerationCandidatesRow{
			{
				StudentID:          studentID,
				Nis:                "001",
				Nisn:               "1234567890",
				Nama:               "Alya",
				BaseUsername:       "1234567890",
				UsernameCollisions: []string{"1234567890"},
			},
			{
				StudentID:        existingStudentID,
				Nis:              "002",
				Nisn:             "2222222222",
				Nama:             "Bima",
				BaseUsername:     "2222222222",
				ExistingUserID:   existingUserID,
				UsernameUserID:   existingUserID,
				ExistingUsername: "2222222222",
			},
			{
				StudentID:    duplicateAID,
				Nis:          "003",
				Nisn:         "9876543210",
				Nama:         "Citra",
				BaseUsername: "9876543210",
			},
			{
				StudentID:    duplicateBID,
				Nis:          "004",
				Nisn:         "9876543210",
				Nama:         "Danu",
				BaseUsername: "9876543210",
			},
			{
				StudentID:    missingBasisID,
				Nama:         "Eka",
				BaseUsername: "",
			},
		},
	}
	generator := &StudentAccountGenerator{q: store, role: StudentAccountRole}

	result, err := generator.Preview(context.Background())
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}

	if result.Total != 5 || result.Ready != 3 || result.Skipped != 2 {
		t.Fatalf("summary = total %d ready %d skipped %d, want 5/3/2", result.Total, result.Ready, result.Skipped)
	}
	assertStudentCandidate(t, result.Candidates[0], studentID, "1234567890-01", "ready", "")
	assertStudentCandidate(t, result.Candidates[1], existingStudentID, "", "skipped", "siswa sudah punya akun")
	assertStudentCandidate(t, result.Candidates[2], duplicateAID, "9876543210", "ready", "")
	assertStudentCandidate(t, result.Candidates[3], duplicateBID, "9876543210-01", "ready", "")
	assertStudentCandidate(t, result.Candidates[4], missingBasisID, "", "skipped", "NISN belum diisi")
}

func TestStudentAccountGenerationGenerateUsesNISNAsDefaultInitialPassword(t *testing.T) {
	studentID := testGenerationUUID(31)
	userID := testGenerationUUID(95)
	actorID := testGenerationUUID(96)
	store := &fakeStudentAccountGenerationStore{
		studentRows: []db.ListStudentAccountGenerationCandidatesRow{{
			StudentID:    studentID,
			Nis:          "201",
			Nisn:         "0099887766",
			Nama:         "Gita",
			BaseUsername: "0099887766",
		}},
		createUserID: userID,
	}
	generator := &StudentAccountGenerator{q: store, role: StudentAccountRole}

	result, err := generator.Generate(context.Background(), actorID)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.Created != 1 || result.Candidates[0].TemporaryPassword != "0099887766" {
		t.Fatalf("result = %+v, want one created candidate with NISN initial password", result)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(store.createArg.PasswordHash), []byte("0099887766")); err != nil {
		t.Fatalf("password hash does not match NISN: %v", err)
	}
}

func TestStudentAccountGenerationGenerateCreatesMustChangePasswordUserAndAudit(t *testing.T) {
	studentID := testGenerationUUID(21)
	userID := testGenerationUUID(92)
	actorID := testGenerationUUID(93)
	store := &fakeStudentAccountGenerationStore{
		studentRows: []db.ListStudentAccountGenerationCandidatesRow{{
			StudentID:    studentID,
			Nis:          "101",
			Nisn:         "0011223344",
			Nama:         "Fahri",
			BaseUsername: "0011223344",
		}},
		createUserID: userID,
	}
	generator := &StudentAccountGenerator{
		q:                 store,
		role:              StudentAccountRole,
		passwordGenerator: fixedAccountPassword("TempPass123!"),
	}

	result, err := generator.Generate(context.Background(), actorID)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.Created != 1 || result.Ready != 0 || result.Candidates[0].Status != "created" {
		t.Fatalf("result = %+v, want one created candidate", result)
	}
	if result.Candidates[0].TemporaryPassword != "TempPass123!" {
		t.Fatalf("temporary password = %q, want generated password", result.Candidates[0].TemporaryPassword)
	}
	if store.createArg.Username != "0011223344" || store.createArg.StudentID != studentID || store.createArg.ParentID.Valid || !store.createArg.IsActive || !store.createArg.MustChangePassword {
		t.Fatalf("CreateUserWithMustChangePassword arg = %+v, want active student must-change user", store.createArg)
	}
	if store.createArg.PasswordHash == "" || store.createArg.PasswordHash == "TempPass123!" {
		t.Fatalf("password hash = %q, want non-plaintext hash", store.createArg.PasswordHash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(store.createArg.PasswordHash), []byte("TempPass123!")); err != nil {
		t.Fatalf("password hash does not match generated password: %v", err)
	}
	if len(store.rbacRoles) != 1 || store.rbacRoles[0].Code != StudentAccountRole {
		t.Fatalf("rbac roles = %+v, want siswa dynamic role", store.rbacRoles)
	}
	if len(store.legacyRoles) != 1 || store.legacyRoles[0].Role != db.UserRoleSiswa {
		t.Fatalf("legacy roles = %+v, want safe siswa compatibility role", store.legacyRoles)
	}
	if len(store.auditLogs) != 1 || store.auditLogs[0].Action != "STUDENT_ACCOUNT_GENERATED" {
		t.Fatalf("audit logs = %+v, want STUDENT_ACCOUNT_GENERATED", store.auditLogs)
	}
	var metadata map[string]any
	if err := json.Unmarshal(store.auditLogs[0].Metadata, &metadata); err != nil {
		t.Fatalf("audit metadata json = %v", err)
	}
	if metadata["student_id"] != studentID.String() || metadata["username"] != "0011223344" {
		t.Fatalf("audit metadata = %+v, want student_id and username", metadata)
	}
}

func TestStudentAccountGenerationGenerateMarksCreateFailuresWithoutLeakingPassword(t *testing.T) {
	studentID := testGenerationUUID(41)
	store := &fakeStudentAccountGenerationStore{
		studentRows: []db.ListStudentAccountGenerationCandidatesRow{{
			StudentID:    studentID,
			Nis:          "301",
			Nisn:         "9988776655",
			Nama:         "Hana",
			BaseUsername: "9988776655",
		}},
		createErr: errors.New("duplicate username"),
	}
	generator := &StudentAccountGenerator{q: store, role: StudentAccountRole}

	result, err := generator.Generate(context.Background(), testGenerationUUID(97))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if result.Created != 0 || result.Failed != 1 || result.Ready != 0 || result.Candidates[0].Status != "failed" || result.Candidates[0].Reason != "duplicate username" {
		t.Fatalf("result = %+v, want failed candidate with safe create error", result)
	}
	if result.Candidates[0].TemporaryPassword != "" || len(store.legacyRoles) != 0 || len(store.rbacRoles) != 0 || len(store.auditLogs) != 0 {
		t.Fatalf("side effects after create failure: candidate=%+v legacy=%d rbac=%d audit=%d; want no password/roles/audit", result.Candidates[0], len(store.legacyRoles), len(store.rbacRoles), len(store.auditLogs))
	}
}

func TestStudentAccountGenerationPasswordHelpers(t *testing.T) {
	custom := &StudentAccountGenerator{passwordGenerator: fixedAccountPassword("CustomPass123!")}
	if got, err := custom.generatePassword(); err != nil || got != "CustomPass123!" {
		t.Fatalf("generatePassword(custom) = %q, %v; want custom password", got, err)
	}

	generated, err := (&StudentAccountGenerator{}).generatePassword()
	if err != nil {
		t.Fatalf("generatePassword(default) error = %v", err)
	}
	assertTemporaryPasswordShape(t, generated)
}

type fakeStudentAccountGenerationStore struct {
	studentRows []db.ListStudentAccountGenerationCandidatesRow
	listErr     error

	createArg    db.CreateUserWithMustChangePasswordParams
	createUserID pgtype.UUID
	createErr    error

	legacyRoles []db.AddUserRoleParams
	rbacRoles   []db.AddUserRbacRoleByCodeParams
	roleErr     error

	auditLogs []db.CreateAuditLogParams
	auditErr  error
}

func (f *fakeStudentAccountGenerationStore) ListStudentAccountGenerationCandidates(ctx context.Context) ([]db.ListStudentAccountGenerationCandidatesRow, error) {
	return f.studentRows, f.listErr
}

func (f *fakeStudentAccountGenerationStore) CreateUserWithMustChangePassword(ctx context.Context, arg db.CreateUserWithMustChangePasswordParams) (db.CreateUserWithMustChangePasswordRow, error) {
	f.createArg = arg
	if !f.createUserID.Valid {
		f.createUserID = testGenerationUUID(94)
	}
	return db.CreateUserWithMustChangePasswordRow{ID: f.createUserID, Username: arg.Username, StudentID: arg.StudentID, ParentID: arg.ParentID, IsActive: arg.IsActive, MustChangePassword: arg.MustChangePassword}, f.createErr
}

func (f *fakeStudentAccountGenerationStore) AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error {
	f.legacyRoles = append(f.legacyRoles, arg)
	return f.roleErr
}

func (f *fakeStudentAccountGenerationStore) AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error {
	f.rbacRoles = append(f.rbacRoles, arg)
	return f.roleErr
}

func (f *fakeStudentAccountGenerationStore) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.auditLogs = append(f.auditLogs, arg)
	return db.AuditLog{}, f.auditErr
}

func fixedAccountPassword(password string) func() (string, error) {
	return func() (string, error) { return password, nil }
}

func assertTemporaryPasswordShape(t *testing.T, password string) {
	t.Helper()
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#%?"
	if len(password) != 12 {
		t.Fatalf("temporary password length = %d, want 12 (%q)", len(password), password)
	}
	for _, char := range password {
		if !strings.ContainsRune(alphabet, char) {
			t.Fatalf("temporary password %q contains disallowed rune %q", password, char)
		}
	}
}

func assertStudentCandidate(t *testing.T, got StudentAccountGenerationCandidate, studentID pgtype.UUID, username string, status string, reason string) {
	t.Helper()
	if got.StudentID != studentID.String() || got.GeneratedUsername != username || got.Status != status || got.Reason != reason {
		t.Fatalf("candidate = %+v, want student=%s username=%q status=%q reason=%q", got, studentID.String(), username, status, reason)
	}
}
