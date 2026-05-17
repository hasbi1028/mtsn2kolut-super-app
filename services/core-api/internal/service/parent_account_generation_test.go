package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestParentAccountGenerationPreviewUsesChildOrPhoneBasisAndSkipReasons(t *testing.T) {
	parentID := testGenerationUUID(31)
	phoneFallbackID := testGenerationUUID(32)
	existingParentID := testGenerationUUID(33)
	noChildID := testGenerationUUID(34)
	noBasisID := testGenerationUUID(35)
	existingUserID := testGenerationUUID(95)

	store := &fakeParentAccountGenerationStore{
		parentRows: []db.ListParentAccountGenerationCandidatesRow{
			{
				ParentID:           parentID,
				Nama:               "Ibu Alya",
				Phone:              "081234567890",
				ChildCount:         2,
				BasisStudentID:     testGenerationUUID(41),
				BasisStudentNisn:   "1111222233",
				BaseUsername:       "ortu1111222233",
				UsernameCollisions: []string{"ortu1111222233", "ortu1111222233-01"},
			},
			{
				ParentID:     phoneFallbackID,
				Nama:         "Bapak Bima",
				Phone:        "0812-3456-7890",
				ChildCount:   1,
				BaseUsername: "ortu34567890",
			},
			{
				ParentID:       existingParentID,
				Nama:           "Wali Citra",
				Phone:          "082200000000",
				ChildCount:     1,
				BaseUsername:   "ortu3333444455",
				ExistingUserID: existingUserID,
			},
			{
				ParentID:     noChildID,
				Nama:         "Wali Tanpa Anak",
				Phone:        "083300000000",
				ChildCount:   0,
				BaseUsername: "ortu33000000",
			},
			{
				ParentID:   noBasisID,
				Nama:       "Wali Tanpa Basis",
				ChildCount: 1,
			},
		},
	}
	generator := &ParentAccountGenerator{q: store, role: ParentAccountRole}

	result, err := generator.Preview(context.Background())
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}

	if result.Total != 5 || result.Ready != 2 || result.Skipped != 3 {
		t.Fatalf("summary = total %d ready %d skipped %d, want 5/2/3", result.Total, result.Ready, result.Skipped)
	}
	assertParentCandidate(t, result.Candidates[0], parentID, "ortu1111222233-02", "ready", "")
	assertParentCandidate(t, result.Candidates[1], phoneFallbackID, "ortu34567890", "ready", "")
	assertParentCandidate(t, result.Candidates[2], existingParentID, "", "skipped", "orang tua sudah punya akun")
	assertParentCandidate(t, result.Candidates[3], noChildID, "", "skipped", "orang tua belum terhubung ke siswa")
	assertParentCandidate(t, result.Candidates[4], noBasisID, "", "skipped", "NISN anak/nomor HP belum tersedia")
}

func TestParentAccountGenerationGenerateCreatesMustChangePasswordUserAndAudit(t *testing.T) {
	parentID := testGenerationUUID(51)
	userID := testGenerationUUID(96)
	actorID := testGenerationUUID(97)
	store := &fakeParentAccountGenerationStore{
		parentRows: []db.ListParentAccountGenerationCandidatesRow{{
			ParentID:     parentID,
			Nama:         "Ibu Danu",
			Phone:        "081299991111",
			ChildCount:   1,
			BaseUsername: "ortu4444555566",
		}},
		createUserID: userID,
	}
	generator := &ParentAccountGenerator{
		q:                 store,
		role:              ParentAccountRole,
		passwordGenerator: fixedAccountPassword("TempPass456!"),
	}

	result, err := generator.Generate(context.Background(), actorID)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.Created != 1 || result.Ready != 0 || result.Candidates[0].Status != "created" {
		t.Fatalf("result = %+v, want one created candidate", result)
	}
	if result.Candidates[0].TemporaryPassword != "TempPass456!" {
		t.Fatalf("temporary password = %q, want generated password", result.Candidates[0].TemporaryPassword)
	}
	if store.createArg.Username != "ortu4444555566" || store.createArg.ParentID != parentID || store.createArg.StudentID.Valid || !store.createArg.IsActive || !store.createArg.MustChangePassword {
		t.Fatalf("CreateUserWithMustChangePassword arg = %+v, want active parent must-change user", store.createArg)
	}
	if store.createArg.PasswordHash == "" || store.createArg.PasswordHash == "TempPass456!" {
		t.Fatalf("password hash = %q, want non-plaintext hash", store.createArg.PasswordHash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(store.createArg.PasswordHash), []byte("TempPass456!")); err != nil {
		t.Fatalf("password hash does not match generated password: %v", err)
	}
	if len(store.rbacRoles) != 1 || store.rbacRoles[0].Code != ParentAccountRole {
		t.Fatalf("rbac roles = %+v, want ortu dynamic role", store.rbacRoles)
	}
	if len(store.legacyRoles) != 1 || store.legacyRoles[0].Role != db.UserRoleOrtu {
		t.Fatalf("legacy roles = %+v, want safe ortu compatibility role", store.legacyRoles)
	}
	if len(store.auditLogs) != 1 || store.auditLogs[0].Action != "PARENT_ACCOUNT_GENERATED" {
		t.Fatalf("audit logs = %+v, want PARENT_ACCOUNT_GENERATED", store.auditLogs)
	}
	var metadata map[string]any
	if err := json.Unmarshal(store.auditLogs[0].Metadata, &metadata); err != nil {
		t.Fatalf("audit metadata json = %v", err)
	}
	if metadata["parent_id"] != parentID.String() || metadata["username"] != "ortu4444555566" {
		t.Fatalf("audit metadata = %+v, want parent_id and username", metadata)
	}
}

func TestParentAccountGenerationGenerateRejectsWeakGeneratedPasswordBeforeUserCreate(t *testing.T) {
	parentID := testGenerationUUID(61)
	store := &fakeParentAccountGenerationStore{
		parentRows: []db.ListParentAccountGenerationCandidatesRow{{
			ParentID:     parentID,
			Nama:         "Ibu Eka",
			ChildCount:   1,
			BaseUsername: "ortu5555666677",
		}},
	}
	generator := &ParentAccountGenerator{
		q:                 store,
		role:              ParentAccountRole,
		passwordGenerator: fixedAccountPassword("short"),
	}

	result, err := generator.Generate(context.Background(), testGenerationUUID(99))
	if err == nil {
		t.Fatalf("Generate(weak password) error = nil, want validation error")
	}
	if result.Ready != 1 || result.Created != 0 || store.createArg.Username != "" || len(store.legacyRoles) != 0 || len(store.rbacRoles) != 0 || len(store.auditLogs) != 0 {
		t.Fatalf("Generate(weak password) result=%+v store=%+v, want validation before any create/role/audit", result, store)
	}
}

func TestParentAccountGenerationHelpers(t *testing.T) {
	custom := &ParentAccountGenerator{passwordGenerator: fixedAccountPassword("ParentPass123!")}
	if got, err := custom.generatePassword(); err != nil || got != "ParentPass123!" {
		t.Fatalf("generatePassword(custom) = %q, %v; want custom password", got, err)
	}

	generated, err := (&ParentAccountGenerator{}).generatePassword()
	if err != nil {
		t.Fatalf("generatePassword(default) error = %v", err)
	}
	assertTemporaryPasswordShape(t, generated)

	result := ParentAccountGenerationResult{Ready: 1}
	item := ParentAccountGenerationCandidate{Status: "ready"}
	markParentAccountGenerationFailed(&result, &item, "parent_id tidak valid")
	if result.Ready != 0 || result.Failed != 1 || item.Status != "failed" || item.Reason != "parent_id tidak valid" {
		t.Fatalf("markParentAccountGenerationFailed result=%+v item=%+v, want failed and decremented ready", result, item)
	}
}

type fakeParentAccountGenerationStore struct {
	parentRows []db.ListParentAccountGenerationCandidatesRow
	listErr    error

	createArg    db.CreateUserWithMustChangePasswordParams
	createUserID pgtype.UUID
	createErr    error

	legacyRoles []db.AddUserRoleParams
	rbacRoles   []db.AddUserRbacRoleByCodeParams
	roleErr     error

	auditLogs []db.CreateAuditLogParams
	auditErr  error
}

func (f *fakeParentAccountGenerationStore) ListParentAccountGenerationCandidates(ctx context.Context) ([]db.ListParentAccountGenerationCandidatesRow, error) {
	return f.parentRows, f.listErr
}

func (f *fakeParentAccountGenerationStore) CreateUserWithMustChangePassword(ctx context.Context, arg db.CreateUserWithMustChangePasswordParams) (db.CreateUserWithMustChangePasswordRow, error) {
	f.createArg = arg
	if !f.createUserID.Valid {
		f.createUserID = testGenerationUUID(98)
	}
	return db.CreateUserWithMustChangePasswordRow{ID: f.createUserID, Username: arg.Username, StudentID: arg.StudentID, ParentID: arg.ParentID, IsActive: arg.IsActive, MustChangePassword: arg.MustChangePassword}, f.createErr
}

func (f *fakeParentAccountGenerationStore) AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error {
	f.legacyRoles = append(f.legacyRoles, arg)
	return f.roleErr
}

func (f *fakeParentAccountGenerationStore) AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error {
	f.rbacRoles = append(f.rbacRoles, arg)
	return f.roleErr
}

func (f *fakeParentAccountGenerationStore) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.auditLogs = append(f.auditLogs, arg)
	return db.AuditLog{}, f.auditErr
}

func assertParentCandidate(t *testing.T, got ParentAccountGenerationCandidate, parentID pgtype.UUID, username string, status string, reason string) {
	t.Helper()
	if got.ParentID != parentID.String() || got.GeneratedUsername != username || got.Status != status || got.Reason != reason {
		t.Fatalf("candidate = %+v, want parent=%s username=%q status=%q reason=%q", got, parentID.String(), username, status, reason)
	}
}
