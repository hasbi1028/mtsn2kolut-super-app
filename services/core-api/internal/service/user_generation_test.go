package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestEmployeeAccountGeneratorUsesSchoolProfileNPSNWhenConfigured(t *testing.T) {
	store := &fakeEmployeeAccountGenerationStore{settings: map[string]db.AppSetting{
		schoolProfilePrefix + "npsn": {Key: schoolProfilePrefix + "npsn", Value: "12345678"},
	}}
	generator := &EmployeeAccountGenerator{q: store, npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}

	result, err := generator.Preview(context.Background())
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}

	if result.NPSN != "12345678" {
		t.Fatalf("NPSN = %q, want %q", result.NPSN, "12345678")
	}
	if store.requestedNPSN != "12345678" {
		t.Fatalf("ListEmployeeAccountGenerationCandidates npsn = %q, want %q", store.requestedNPSN, "12345678")
	}
}

func TestEmployeeAccountGeneratorFallsBackToDefaultNPSNWhenSettingMissing(t *testing.T) {
	store := &fakeEmployeeAccountGenerationStore{settings: map[string]db.AppSetting{}}
	generator := &EmployeeAccountGenerator{q: store, npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}

	result, err := generator.Preview(context.Background())
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}

	if result.NPSN != DefaultEmployeeAccountNPSN {
		t.Fatalf("NPSN = %q, want default %q", result.NPSN, DefaultEmployeeAccountNPSN)
	}
	if store.requestedNPSN != DefaultEmployeeAccountNPSN {
		t.Fatalf("ListEmployeeAccountGenerationCandidates npsn = %q, want default %q", store.requestedNPSN, DefaultEmployeeAccountNPSN)
	}
}

func TestEmployeeAccountGenerationPreviewKeepsBackendGeneratedBirthYearUsername(t *testing.T) {
	store := &fakeEmployeeAccountGenerationStore{
		settings: map[string]db.AppSetting{schoolProfilePrefix + "npsn": {Key: schoolProfilePrefix + "npsn", Value: "40406031"}},
		rows: []db.ListEmployeeAccountGenerationCandidatesRow{
			{EmployeeID: testGenerationUUID(1), Nip: "1", Nama: "A", TanggalLahir: validDate(1990, 5, 7), NomorUrut: 1, GeneratedUsername: "4040603190001"},
			{EmployeeID: testGenerationUUID(2), Nip: "2", Nama: "B", TanggalLahir: validDate(1990, 9, 9), NomorUrut: 2, GeneratedUsername: "4040603190002"},
			{EmployeeID: testGenerationUUID(3), Nip: "3", Nama: "C", TanggalLahir: validDate(1985, 1, 1), NomorUrut: 1, GeneratedUsername: "4040603185001"},
		},
	}
	generator := &EmployeeAccountGenerator{q: store, npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}

	result, err := generator.Preview(context.Background())
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}

	got := []string{result.Items[0].Username, result.Items[1].Username, result.Items[2].Username}
	want := []string{"4040603190001", "4040603190002", "4040603185001"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("username[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

type fakeEmployeeAccountGenerationStore struct {
	settings      map[string]db.AppSetting
	rows          []db.ListEmployeeAccountGenerationCandidatesRow
	requestedNPSN string
	createErr     error
	roleErr       error
	rbacRoleErr   error
	auditErr      error

	createCalls   []db.CreateUserParams
	roleCalls     []db.AddUserRoleParams
	rbacRoleCalls []db.AddUserRbacRoleByCodeParams
	auditCalls    []db.CreateAuditLogParams
}

func (f *fakeEmployeeAccountGenerationStore) GetSetting(ctx context.Context, key string) (db.AppSetting, error) {
	if setting, ok := f.settings[key]; ok {
		return setting, nil
	}
	return db.AppSetting{}, pgx.ErrNoRows
}

func (f *fakeEmployeeAccountGenerationStore) ListEmployeeAccountGenerationCandidates(ctx context.Context, npsn string) ([]db.ListEmployeeAccountGenerationCandidatesRow, error) {
	f.requestedNPSN = npsn
	return f.rows, nil
}

func (f *fakeEmployeeAccountGenerationStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error) {
	f.createCalls = append(f.createCalls, arg)
	return db.CreateUserRow{ID: testGenerationUUID(99), Username: arg.Username, DisplayName: arg.DisplayName, EmployeeID: arg.EmployeeID, IsActive: arg.IsActive}, f.createErr
}

func (f *fakeEmployeeAccountGenerationStore) AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error {
	f.roleCalls = append(f.roleCalls, arg)
	return f.roleErr
}

func (f *fakeEmployeeAccountGenerationStore) AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error {
	f.rbacRoleCalls = append(f.rbacRoleCalls, arg)
	return f.rbacRoleErr
}

func (f *fakeEmployeeAccountGenerationStore) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.auditCalls = append(f.auditCalls, arg)
	return db.AuditLog{}, f.auditErr
}

func TestEmployeeAccountGeneratorConstructorsInitializeDefaults(t *testing.T) {
	if generator := NewEmployeeAccountGenerator(nil); generator == nil || generator.q == nil || generator.tx != nil || generator.npsn != DefaultEmployeeAccountNPSN || generator.role != DefaultEmployeeAccountRole {
		t.Fatalf("NewEmployeeAccountGenerator(nil) = %+v, want defaults and no tx", generator)
	}
	if generator := NewEmployeeAccountGeneratorWithPool(nil); generator == nil || generator.q == nil || generator.tx == nil || generator.npsn != DefaultEmployeeAccountNPSN || generator.role != DefaultEmployeeAccountRole {
		t.Fatalf("NewEmployeeAccountGeneratorWithPool(nil) = %+v, want defaults and tx", generator)
	}
}

func TestEmployeeAccountGenerationBuildResultClassifiesEveryCandidate(t *testing.T) {
	existingID := testGenerationUUID(80)
	usernameOwnerID := testGenerationUUID(81)
	generator := &EmployeeAccountGenerator{role: "guru"}

	result := generator.buildResult([]db.ListEmployeeAccountGenerationCandidatesRow{
		{EmployeeID: testGenerationUUID(1), Nip: "1", Nama: "Ready", TanggalLahir: validDate(1990, 1, 2), NomorUrut: 1, GeneratedUsername: "4040603190001"},
		{EmployeeID: testGenerationUUID(2), Nip: "2", Nama: "No Birth", NomorUrut: 2, GeneratedUsername: "4040603190002"},
		{EmployeeID: testGenerationUUID(3), Nip: "3", Nama: "Has User", TanggalLahir: validDate(1991, 2, 3), NomorUrut: 3, GeneratedUsername: "4040603191003", ExistingUserID: existingID},
		{EmployeeID: testGenerationUUID(4), Nip: "4", Nama: "Username Taken", TanggalLahir: validDate(1992, 3, 4), NomorUrut: 4, GeneratedUsername: "4040603192004", UsernameUserID: usernameOwnerID},
	}, true, "40406031")

	if result.Total != 4 || result.Ready != 1 || result.Skipped != 3 || result.Created != 0 || result.Failed != 0 {
		t.Fatalf("result counts = %+v, want one ready and three skipped", result)
	}
	if result.Items[0].Status != "ready" || result.Items[0].Password != result.Items[0].Username || result.Items[0].TanggalLahir != "1990-01-02" {
		t.Fatalf("ready item = %+v, want password and formatted birth date", result.Items[0])
	}
	if result.Items[1].Message != "tanggal lahir belum diisi" || result.Items[2].ExistingUserID == "" || result.Items[3].UsernameUserID == "" {
		t.Fatalf("skipped items = %+v, want missing-birth/existing/username-owner reasons", result.Items[1:])
	}
}

func TestEmployeeAccountGeneratorGenerateCreatesReadyAccountsAndAudits(t *testing.T) {
	actorID := testGenerationUUID(82)
	store := &fakeEmployeeAccountGenerationStore{rows: []db.ListEmployeeAccountGenerationCandidatesRow{{EmployeeID: testGenerationUUID(83), Nip: "1", Nama: "Guru Satu", TanggalLahir: validDate(1990, 1, 1), NomorUrut: 1, GeneratedUsername: "4040603190001"}}}
	generator := &EmployeeAccountGenerator{q: store, npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}

	result, err := generator.Generate(context.Background(), actorID)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if result.Created != 1 || result.Ready != 0 || result.Failed != 0 || result.Items[0].Status != "created" || result.Items[0].Password != result.Items[0].Username || result.Items[0].ExistingUserID == "" {
		t.Fatalf("Generate() result = %+v, want created account", result)
	}
	if len(store.createCalls) != 1 || store.createCalls[0].Username != "4040603190001" || !store.createCalls[0].IsActive || !store.createCalls[0].DisplayName.Valid {
		t.Fatalf("create calls = %+v, want active user with display name", store.createCalls)
	}
	if len(store.roleCalls) != 1 || store.roleCalls[0].Role != db.UserRole(DefaultEmployeeAccountRole) || len(store.rbacRoleCalls) != 1 || store.rbacRoleCalls[0].Code != DefaultEmployeeAccountRole {
		t.Fatalf("role calls = legacy %+v rbac %+v, want default role", store.roleCalls, store.rbacRoleCalls)
	}
	if len(store.auditCalls) != 1 || store.auditCalls[0].Action != "USER_EMPLOYEE_ACCOUNT_GENERATED" || store.auditCalls[0].UserID != actorID {
		t.Fatalf("audit calls = %+v, want generation audit", store.auditCalls)
	}
	var metadata map[string]any
	if err := json.Unmarshal(store.auditCalls[0].Metadata, &metadata); err != nil || metadata["username"] != "4040603190001" || metadata["role"] != DefaultEmployeeAccountRole {
		t.Fatalf("audit metadata = %s err=%v, want username and role", string(store.auditCalls[0].Metadata), err)
	}
}

func TestEmployeeAccountGeneratorGenerateMarksCreateAndInvalidUUIDFailures(t *testing.T) {
	store := &fakeEmployeeAccountGenerationStore{
		createErr: errors.New("duplicate username"),
		rows: []db.ListEmployeeAccountGenerationCandidatesRow{
			{EmployeeID: testGenerationUUID(84), Nip: "1", Nama: "Create Fails", TanggalLahir: validDate(1990, 1, 1), NomorUrut: 1, GeneratedUsername: "4040603190001"},
			{Nip: "2", Nama: "Bad UUID", TanggalLahir: validDate(1991, 1, 1), NomorUrut: 2, GeneratedUsername: "4040603191002"},
		},
	}
	generator := &EmployeeAccountGenerator{q: store, npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}

	result, err := generator.Generate(context.Background(), testGenerationUUID(85))
	if err != nil {
		t.Fatalf("Generate() error = %v, want item failures but no fatal error", err)
	}
	if result.Failed != 2 || result.Created != 0 || result.Ready != 0 || result.Items[0].Status != "failed" || result.Items[0].Message != "duplicate username" || result.Items[1].Message != "employee_id tidak valid" {
		t.Fatalf("Generate() result = %+v, want create and uuid failures", result)
	}
	if len(store.roleCalls) != 0 || len(store.auditCalls) != 0 {
		t.Fatalf("role/audit calls = %+v/%+v, want none for failed items", store.roleCalls, store.auditCalls)
	}
}

func TestAuditEmployeeAccountGenerationPropagatesStoreError(t *testing.T) {
	expected := errors.New("audit down")
	store := &fakeEmployeeAccountGenerationStore{auditErr: expected}
	err := auditEmployeeAccountGeneration(context.Background(), store, testGenerationUUID(86), testGenerationUUID(87), testGenerationUUID(88).String(), "user1", "guru")
	if !errors.Is(err, expected) {
		t.Fatalf("auditEmployeeAccountGeneration() error = %v, want %v", err, expected)
	}
}

func testGenerationUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func validDate(year int, month int, day int) pgtype.Date {
	var date pgtype.Date
	_ = date.Scan(testDateString(year, month, day))
	return date
}

func testDateString(year int, month int, day int) string {
	return fmtDate(year, month, day)
}

func fmtDate(year int, month int, day int) string {
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}
