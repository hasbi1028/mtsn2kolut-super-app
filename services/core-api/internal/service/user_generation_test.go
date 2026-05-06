package service

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestEmployeeAccountGeneratorPreviewBuildsNPSNBirthDateUsernameAndSkipsUnsafeRows(t *testing.T) {
	generator := &EmployeeAccountGenerator{npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}
	readyEmployee := userLifecycleTestUUID(11)
	linkedEmployee := userLifecycleTestUUID(12)
	linkedUser := userLifecycleTestUUID(13)
	usernameTakenEmployee := userLifecycleTestUUID(14)
	usernameTakenUser := userLifecycleTestUUID(15)
	missingBirthEmployee := userLifecycleTestUUID(16)

	result := generator.buildResult([]db.ListEmployeeAccountGenerationCandidatesRow{
		{EmployeeID: readyEmployee, Nip: "198", Nama: "Hasbi Awal", TanggalLahir: pgtype.Date{Time: time.Date(1992, 10, 28, 0, 0, 0, 0, time.UTC), Valid: true}, NomorUrut: 1, GeneratedUsername: "40406031281092001"},
		{EmployeeID: linkedEmployee, Nip: "199", Nama: "Pegawai Berakun", TanggalLahir: pgtype.Date{Time: time.Date(1990, 1, 2, 0, 0, 0, 0, time.UTC), Valid: true}, NomorUrut: 2, GeneratedUsername: "40406031020190002", ExistingUserID: linkedUser},
		{EmployeeID: usernameTakenEmployee, Nip: "200", Nama: "Username Dipakai", TanggalLahir: pgtype.Date{Time: time.Date(1991, 3, 4, 0, 0, 0, 0, time.UTC), Valid: true}, NomorUrut: 3, GeneratedUsername: "40406031040391003", UsernameUserID: usernameTakenUser},
		{EmployeeID: missingBirthEmployee, Nip: "201", Nama: "Belum Ada Tanggal", NomorUrut: 4},
	}, false)

	if result.Total != 4 || result.Ready != 1 || result.Skipped != 3 || result.Created != 0 || result.DefaultRole != "guru" {
		t.Fatalf("result summary = %+v, want 1 ready and 3 skipped", result)
	}
	if got := result.Items[0].Username; got != "40406031281092001" {
		t.Fatalf("ready username = %q, want NPSN+DDMMYY+urut", got)
	}
	if result.Items[0].Password != "" {
		t.Fatalf("preview must not expose password, got %q", result.Items[0].Password)
	}
	if result.Items[1].Status != "skipped" || result.Items[2].Status != "skipped" || result.Items[3].Message != "tanggal lahir belum diisi" {
		t.Fatalf("skip statuses = %+v", result.Items)
	}
}

func TestEmployeeAccountGeneratorGeneratePreviewIncludesInitialPasswordOnlyForReadyRows(t *testing.T) {
	generator := &EmployeeAccountGenerator{npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}
	result := generator.buildResult([]db.ListEmployeeAccountGenerationCandidatesRow{{EmployeeID: userLifecycleTestUUID(21), TanggalLahir: pgtype.Date{Time: time.Date(1992, 10, 28, 0, 0, 0, 0, time.UTC), Valid: true}, NomorUrut: 1, GeneratedUsername: "40406031281092001"}}, true)
	if len(result.Items) != 1 || result.Items[0].Password != result.Items[0].Username {
		t.Fatalf("generate result item = %+v, want password equal username", result.Items)
	}
}
