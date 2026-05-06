package service

import (
	"context"
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
	return db.CreateUserRow{ID: testGenerationUUID(99), Username: arg.Username, DisplayName: arg.DisplayName, EmployeeID: arg.EmployeeID, IsActive: arg.IsActive}, nil
}

func (f *fakeEmployeeAccountGenerationStore) AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error {
	return nil
}

func (f *fakeEmployeeAccountGenerationStore) AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error {
	return nil
}

func (f *fakeEmployeeAccountGenerationStore) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	return db.AuditLog{}, nil
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
