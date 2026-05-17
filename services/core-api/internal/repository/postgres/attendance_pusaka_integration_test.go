package db

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationAttendancePusakaEmployeesRepository(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	pnsID := createAttendancePusakaEmployee(t, ctx, q, suffix+"-pns", "pns", true)
	pppkID := createAttendancePusakaEmployee(t, ctx, q, suffix+"-pppk", "pppk", true)
	honorerID := createAttendancePusakaEmployee(t, ctx, q, suffix+"-honorer", "honorer", true)
	inactiveID := createAttendancePusakaEmployee(t, ctx, q, suffix+"-inactive", "pns", false)

	pnsAccount, err := q.UpsertPusakaAccount(ctx, UpsertPusakaAccountParams{
		EmployeeID:     pnsID,
		PusakaUsername: "it-pusaka-user-" + suffix,
		PusakaPassword: "it-fake-password-" + suffix,
		IsEnabled:      true,
	})
	if err != nil {
		t.Fatalf("upsert pns pusaka account: %v", err)
	}
	if pnsAccount.EmployeeID != pnsID || pnsAccount.PusakaUsername != "it-pusaka-user-"+suffix || !pnsAccount.IsEnabled {
		t.Fatalf("pns pusaka account = %#v", pnsAccount)
	}

	updatedAccount, err := q.UpsertPusakaAccount(ctx, UpsertPusakaAccountParams{
		EmployeeID:     pnsID,
		PusakaUsername: "it-pusaka-user-updated-" + suffix,
		PusakaPassword: "",
		IsEnabled:      true,
	})
	if err != nil {
		t.Fatalf("update pns pusaka account preserving password: %v", err)
	}
	if updatedAccount.PusakaPassword != "it-fake-password-"+suffix || updatedAccount.PusakaUsername != "it-pusaka-user-updated-"+suffix {
		t.Fatalf("updated pusaka account = %#v, want username update and preserved password", updatedAccount)
	}

	if _, err := q.UpsertPusakaAccount(ctx, UpsertPusakaAccountParams{
		EmployeeID:     pppkID,
		PusakaUsername: "it-pusaka-pppk-" + suffix,
		PusakaPassword: "it-fake-password-pppk-" + suffix,
		IsEnabled:      false,
	}); err != nil {
		t.Fatalf("upsert disabled pppk pusaka account: %v", err)
	}

	gotPNS, err := q.GetEmployee(ctx, pnsID)
	if err != nil {
		t.Fatalf("get pns employee: %v", err)
	}
	if gotPNS.PusakaUsername != "it-pusaka-user-updated-"+suffix || gotPNS.PusakaPassword != "it-fake-password-"+suffix || !gotPNS.PusakaIsEnabled {
		t.Fatalf("get pns employee pusaka projection = %#v", gotPNS)
	}
	if gotPNS.PegawaiUid == "" || gotPNS.Nama != "Integration Pusaka "+suffix+"-pns" {
		t.Fatalf("get pns employee identity = %#v", gotPNS)
	}

	activeEmployees, err := q.ListActiveEmployees(ctx)
	if err != nil {
		t.Fatalf("list active pusaka employees: %v", err)
	}
	if !activeEmployeeListed(activeEmployees, pnsID) || activeEmployeeListed(activeEmployees, pppkID) || activeEmployeeListed(activeEmployees, honorerID) || activeEmployeeListed(activeEmployees, inactiveID) {
		t.Fatalf("active pusaka employees = %#v, want only enabled active account", activeEmployees)
	}

	eligibleEmployees, err := q.ListPusakaEligibleEmployeesWithStatus(ctx)
	if err != nil {
		t.Fatalf("list pusaka eligible employees with status: %v", err)
	}
	if !eligibleEmployeeListed(eligibleEmployees, pnsID) || !eligibleEmployeeListed(eligibleEmployees, pppkID) || eligibleEmployeeListed(eligibleEmployees, honorerID) {
		t.Fatalf("eligible employees = %#v, want pns/pppk only", eligibleEmployees)
	}

	allWithStatus, err := q.ListEmployeesWithStatus(ctx)
	if err != nil {
		t.Fatalf("list employees with status: %v", err)
	}
	if !employeeStatusListed(allWithStatus, pnsID, true, true) || !employeeStatusListed(allWithStatus, honorerID, false, false) {
		t.Fatalf("employees with status = %#v, want pusaka eligibility/account flags", allWithStatus)
	}

	createdJob, err := q.CreateJobIfAbsent(ctx, CreateJobIfAbsentParams{
		EmployeeID:  pnsID,
		RunType:     RunTypeEnumCheckin,
		MaxAttempts: 3,
		NotBefore:   pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
	})
	if err != nil {
		t.Fatalf("create employee status job: %v", err)
	}
	allWithStatus, err = q.ListEmployeesWithStatus(ctx)
	if err != nil {
		t.Fatalf("list employees with active job status: %v", err)
	}
	if !employeeStatusHasJob(allWithStatus, pnsID, "queued", "checkin") {
		t.Fatalf("employees with status missing queued checkin job %v: %#v", createdJob.ID, allWithStatus)
	}

	updatedEmployeeID, err := q.UpdateEmployee(ctx, UpdateEmployeeParams{
		ID:             pnsID,
		Nip:            "20" + suffix,
		Nama:           "Integration Pusaka Updated " + suffix,
		UnitKerja:      "MTsN 2 Kolaka Utara",
		EmploymentType: "pns",
		TanggalLahir:   pgDate(1987, time.March, 4),
		JenisKelamin:   "P",
		TempatLahir:    "Lasusua",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("update pns employee: %v", err)
	}
	if updatedEmployeeID != pnsID {
		t.Fatalf("updated employee id = %v, want %v", updatedEmployeeID, pnsID)
	}
	gotUpdated, err := q.GetEmployee(ctx, pnsID)
	if err != nil {
		t.Fatalf("get updated pns employee: %v", err)
	}
	if gotUpdated.Nama != "Integration Pusaka Updated "+suffix || gotUpdated.Nip != "20"+suffix || gotUpdated.JenisKelamin != "P" {
		t.Fatalf("updated employee = %#v", gotUpdated)
	}

	if err := q.DeletePusakaAccountByEmployeeID(ctx, pppkID); err != nil {
		t.Fatalf("delete pppk pusaka account: %v", err)
	}
	if _, err := q.GetPusakaAccountByEmployeeID(ctx, pppkID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("get deleted pppk pusaka account err = %v, want pgx.ErrNoRows", err)
	}

	if count, err := q.CountEmployees(ctx); err != nil {
		t.Fatalf("count employees: %v", err)
	} else if count < 4 {
		t.Fatalf("employee count = %d, want at least created employees", count)
	}
}

func TestIntegrationAttendancePresenceAndTelegramRepository(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	employeeID := createAttendancePusakaEmployee(t, ctx, q, suffix+"-att", "pns", true)
	otherEmployeeID := createAttendancePusakaEmployee(t, ctx, q, suffix+"-att-other", "pppk", true)
	_ = createAttendancePusakaEmployee(t, ctx, q, suffix+"-att-honorer", "honorer", true)

	reportDate := pgDate(2026, time.May, 17)
	otherDate := pgDate(2026, time.May, 18)
	sourceJob := createAttendanceSourceJob(t, ctx, q, employeeID, RunTypeEnumCheckin)
	otherSourceJob := createAttendanceSourceJob(t, ctx, q, otherEmployeeID, RunTypeEnumCheckin)

	first, err := q.UpsertAttendance(ctx, UpsertAttendanceParams{
		EmployeeID:  employeeID,
		Tanggal:     reportDate,
		JamMasuk:    "07:01",
		JamPulang:   "",
		SourceJobID: sourceJob,
	})
	if err != nil {
		t.Fatalf("upsert first attendance: %v", err)
	}
	if first.EmployeeID != employeeID || first.JamMasuk != "07:01" || first.JamPulang != "" {
		t.Fatalf("first attendance = %#v", first)
	}

	merged, err := q.UpsertAttendance(ctx, UpsertAttendanceParams{
		EmployeeID:  employeeID,
		Tanggal:     reportDate,
		JamMasuk:    "",
		JamPulang:   "15:30",
		SourceJobID: sourceJob,
	})
	if err != nil {
		t.Fatalf("merge checkout attendance: %v", err)
	}
	if merged.ID != first.ID || merged.JamMasuk != "07:01" || merged.JamPulang != "15:30" {
		t.Fatalf("merged attendance = %#v, want preserved checkin and updated checkout", merged)
	}

	if _, err := q.UpsertAttendance(ctx, UpsertAttendanceParams{EmployeeID: otherEmployeeID, Tanggal: reportDate, JamMasuk: "07:15", SourceJobID: otherSourceJob}); err != nil {
		t.Fatalf("upsert other employee attendance: %v", err)
	}
	if _, err := q.UpsertAttendance(ctx, UpsertAttendanceParams{EmployeeID: employeeID, Tanggal: otherDate, JamMasuk: "07:05", JamPulang: "15:35", SourceJobID: sourceJob}); err != nil {
		t.Fatalf("upsert other date attendance: %v", err)
	}

	inRangeCount, err := q.CountAttendanceInRange(ctx, CountAttendanceInRangeParams{Tanggal: reportDate, Tanggal_2: otherDate})
	if err != nil {
		t.Fatalf("count attendance in range: %v", err)
	}
	if inRangeCount != 3 {
		t.Fatalf("attendance in range count = %d, want 3", inRangeCount)
	}

	byDate, err := q.ListAttendanceByDate(ctx, reportDate)
	if err != nil {
		t.Fatalf("list attendance by date: %v", err)
	}
	if len(byDate) != 2 || !attendanceByDateListed(byDate, employeeID, "07:01", "15:30") || !attendanceByDateListed(byDate, otherEmployeeID, "07:15", "") {
		t.Fatalf("attendance by date = %#v", byDate)
	}

	byEmployee, err := q.ListAttendanceByEmployee(ctx, ListAttendanceByEmployeeParams{EmployeeID: employeeID, Limit: 10})
	if err != nil {
		t.Fatalf("list attendance by employee: %v", err)
	}
	if len(byEmployee) != 2 || byEmployee[0].Tanggal.Time.Before(byEmployee[1].Tanggal.Time) {
		t.Fatalf("attendance by employee = %#v, want two rows ordered desc", byEmployee)
	}

	inRange, err := q.ListAttendanceInRange(ctx, ListAttendanceInRangeParams{Tanggal: reportDate, Tanggal_2: otherDate})
	if err != nil {
		t.Fatalf("list attendance in range: %v", err)
	}
	if len(inRange) != 3 {
		t.Fatalf("attendance in range len = %d, want 3: %#v", len(inRange), inRange)
	}

	paged, err := q.ListAttendance(ctx, ListAttendanceParams{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("list paged attendance: %v", err)
	}
	if len(paged) < 2 {
		t.Fatalf("paged attendance len = %d, want at least 2", len(paged))
	}

	summary, err := q.GetMonthlyAttendanceSummary(ctx, GetMonthlyAttendanceSummaryParams{Tanggal: reportDate, Tanggal_2: otherDate})
	if err != nil {
		t.Fatalf("get monthly attendance summary: %v", err)
	}
	if !attendanceSummaryListed(summary, employeeID, 2, 2, 0, 0) || !attendanceSummaryListed(summary, otherEmployeeID, 1, 0, 1, 0) {
		t.Fatalf("monthly attendance summary = %#v", summary)
	}

	reportRows, err := q.ListPusakaAttendanceTelegramReportRows(ctx, reportDate)
	if err != nil {
		t.Fatalf("list pusaka attendance telegram report rows: %v", err)
	}
	if !telegramReportListed(reportRows, employeeID, "07:01", "15:30") || !telegramReportListed(reportRows, otherEmployeeID, "07:15", "") || telegramReportContainsName(reportRows, "Integration Pusaka "+suffix+"-att-honorer") {
		t.Fatalf("telegram report rows = %#v, want pns/pppk only with attendance projection", reportRows)
	}

	settings, err := q.UpsertPusakaAttendanceTelegramSettings(ctx, UpsertPusakaAttendanceTelegramSettingsParams{
		IsEnabled:      true,
		SendTime:       pgtype.Time{Microseconds: int64((7*time.Hour + 30*time.Minute) / time.Microsecond), Valid: true},
		SendTimes:      []string{"07:30", "15:45"},
		SendDays:       []int32{1, 2, 3, 4, 5},
		Timezone:       "Asia/Makassar",
		TargetChatID:   "1234567890" + suffix,
		IncludeCaption: true,
		IncludeImage:   false,
		ReportMode:     "ringkas",
	})
	if err != nil {
		t.Fatalf("upsert telegram settings: %v", err)
	}
	if !settings.IsEnabled || settings.TargetChatID != "1234567890"+suffix || settings.ReportMode != "ringkas" || len(settings.SendTimes) != 2 || len(settings.SendDays) != 5 {
		t.Fatalf("telegram settings = %#v", settings)
	}
	gotSettings, err := q.GetPusakaAttendanceTelegramSettings(ctx)
	if err != nil {
		t.Fatalf("get telegram settings: %v", err)
	}
	if gotSettings.TargetChatID != settings.TargetChatID || gotSettings.Timezone != "Asia/Makassar" {
		t.Fatalf("got telegram settings = %#v", gotSettings)
	}

	actor, err := q.CreateUserWithMustChangePassword(ctx, CreateUserWithMustChangePasswordParams{
		Username:           "it-pusaka-telegram-" + suffix,
		PasswordHash:       "hash:v1:" + suffix,
		DisplayName:        pgtype.Text{String: "Integration Pusaka Telegram " + suffix, Valid: true},
		EmployeeID:         employeeID,
		IsActive:           true,
		MustChangePassword: false,
	})
	if err != nil {
		t.Fatalf("create telegram actor user: %v", err)
	}

	logRow, err := q.CreatePusakaAttendanceTelegramLog(ctx, CreatePusakaAttendanceTelegramLogParams{
		ReportDate:        reportDate,
		TargetChatID:      settings.TargetChatID,
		SendMode:          "scheduled",
		ScheduleTime:      "07:30",
		Status:            "success",
		TelegramMessageID: pgtype.Text{String: "msg-" + suffix, Valid: true},
		RequestedBy:       actor.ID,
	})
	if err != nil {
		t.Fatalf("create telegram log: %v", err)
	}
	if logRow.Status != "success" || logRow.RequestedBy != actor.ID || !logRow.TelegramMessageID.Valid {
		t.Fatalf("telegram log = %#v", logRow)
	}
	hasScheduled, err := q.HasPusakaAttendanceTelegramScheduledLog(ctx, HasPusakaAttendanceTelegramScheduledLogParams{ReportDate: reportDate, TargetChatID: settings.TargetChatID, ScheduleTime: "07:30"})
	if err != nil {
		t.Fatalf("has scheduled telegram log: %v", err)
	}
	if !hasScheduled {
		t.Fatalf("scheduled telegram log not found")
	}
	logs, err := q.ListPusakaAttendanceTelegramLogs(ctx, ListPusakaAttendanceTelegramLogsParams{Limit: 10})
	if err != nil {
		t.Fatalf("list telegram logs: %v", err)
	}
	if !telegramLogListedMasked(logs, logRow.ID, settings.TargetChatID) {
		t.Fatalf("telegram logs = %#v, want masked row for %v", logs, logRow.ID)
	}
}

func TestIntegrationEmployeeLinkedUserAccountSummary(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()
	employeeID := createAttendancePusakaEmployee(t, ctx, q, suffix+"-user", "pns", true)

	user, err := q.CreateUserWithMustChangePassword(ctx, CreateUserWithMustChangePasswordParams{
		Username:           "it-employee-user-" + suffix,
		PasswordHash:       "hash:v1:" + suffix,
		DisplayName:        pgtype.Text{String: "", Valid: false},
		EmployeeID:         employeeID,
		IsActive:           true,
		MustChangePassword: true,
	})
	if err != nil {
		t.Fatalf("create employee-linked user: %v", err)
	}
	if err := q.AddUserRole(ctx, AddUserRoleParams{UserID: user.ID, Role: UserRoleGuru}); err != nil {
		t.Fatalf("add legacy guru role: %v", err)
	}

	gotByUsername, err := q.GetUserByUsername(ctx, user.Username)
	if err != nil {
		t.Fatalf("get employee-linked user by username: %v", err)
	}
	if gotByUsername.ID != user.ID || gotByUsername.EmployeeID != employeeID || !gotByUsername.MustChangePassword {
		t.Fatalf("got user by username = %#v", gotByUsername)
	}
	gotByID, err := q.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get employee-linked user by id: %v", err)
	}
	if gotByID.Username != user.Username || gotByID.EmployeeID != employeeID || gotByID.PasswordHash != "hash:v1:"+suffix {
		t.Fatalf("got user by id = %#v", gotByID)
	}
	roles, err := q.GetUserRoles(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user roles: %v", err)
	}
	if len(roles) != 1 || roles[0] != UserRoleGuru {
		t.Fatalf("legacy roles = %#v, want guru", roles)
	}

	summary, err := q.GetUserAccountSummary(ctx, user.ID)
	if err != nil {
		t.Fatalf("get employee-linked user account summary: %v", err)
	}
	if summary.ID != user.ID || summary.ProfileType != "employee" || summary.ProfileNama != "Integration Pusaka "+suffix+"-user" || summary.DisplayName != summary.ProfileNama {
		t.Fatalf("employee user summary = %#v", summary)
	}
	if !summary.EmployeeID.Valid || summary.StudentID.Valid || summary.ParentID.Valid {
		t.Fatalf("employee user summary profile ids = employee:%v student:%v parent:%v", summary.EmployeeID.Valid, summary.StudentID.Valid, summary.ParentID.Valid)
	}
}

func createAttendanceSourceJob(t *testing.T, ctx context.Context, q *Queries, employeeID pgtype.UUID, runType RunTypeEnum) pgtype.UUID {
	t.Helper()
	job, err := q.CreateJobIfAbsent(ctx, CreateJobIfAbsentParams{
		EmployeeID:  employeeID,
		RunType:     runType,
		MaxAttempts: 1,
		NotBefore:   pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
	})
	if err != nil {
		t.Fatalf("create attendance source job for %v: %v", employeeID, err)
	}
	return job.ID
}

func createAttendancePusakaEmployee(t *testing.T, ctx context.Context, q *Queries, suffix, employmentType string, active bool) pgtype.UUID {
	t.Helper()
	employeeID, err := q.CreateEmployee(ctx, CreateEmployeeParams{
		Npsn:           "40404224",
		Nip:            "18" + suffix,
		Nama:           "Integration Pusaka " + suffix,
		UnitKerja:      "MTsN 2 Kolaka Utara",
		EmploymentType: employmentType,
		TanggalLahir:   pgDate(1988, time.January, 2),
		JenisKelamin:   "L",
		TempatLahir:    "Kolaka Utara",
		IsActive:       active,
	})
	if err != nil {
		t.Fatalf("create attendance pusaka employee %q: %v", suffix, err)
	}
	if !employeeID.Valid {
		t.Fatalf("created attendance pusaka employee %q has invalid id", suffix)
	}
	return employeeID
}

func activeEmployeeListed(rows []ListActiveEmployeesRow, id pgtype.UUID) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}
	return false
}

func eligibleEmployeeListed(rows []ListPusakaEligibleEmployeesWithStatusRow, id pgtype.UUID) bool {
	for _, row := range rows {
		if row.ID == id {
			return row.PusakaEligible
		}
	}
	return false
}

func employeeStatusListed(rows []ListEmployeesWithStatusRow, id pgtype.UUID, eligible, hasAccount bool) bool {
	for _, row := range rows {
		if row.ID == id {
			return row.PusakaEligible == eligible && fmt.Sprint(row.HasPusakaAccount) == fmt.Sprint(hasAccount)
		}
	}
	return false
}

func employeeStatusHasJob(rows []ListEmployeesWithStatusRow, id pgtype.UUID, status, runType string) bool {
	for _, row := range rows {
		if row.ID == id {
			return fmt.Sprint(row.ActiveStatus) == status && fmt.Sprint(row.ActiveRunType) == runType && fmt.Sprint(row.LastStatus) == status && fmt.Sprint(row.LastRunType) == runType
		}
	}
	return false
}

func attendanceByDateListed(rows []ListAttendanceByDateRow, id pgtype.UUID, checkin, checkout string) bool {
	for _, row := range rows {
		if row.EmployeeID == id {
			return row.JamMasuk == checkin && row.JamPulang == checkout
		}
	}
	return false
}

func attendanceSummaryListed(rows []GetMonthlyAttendanceSummaryRow, id pgtype.UUID, total, complete, missingCheckout, missingCheckin int32) bool {
	for _, row := range rows {
		if row.EmployeeID == id {
			return row.TotalDays == total && row.CompleteDays == complete && row.MissingCheckout == missingCheckout && row.MissingCheckin == missingCheckin
		}
	}
	return false
}

func telegramReportListed(rows []ListPusakaAttendanceTelegramReportRowsRow, id pgtype.UUID, checkin, checkout string) bool {
	for _, row := range rows {
		if row.EmployeeID == id {
			return row.JamMasuk == checkin && row.JamPulang == checkout
		}
	}
	return false
}

func telegramReportContainsName(rows []ListPusakaAttendanceTelegramReportRowsRow, name string) bool {
	for _, row := range rows {
		if row.EmployeeNama == name {
			return true
		}
	}
	return false
}

func telegramLogListedMasked(rows []ListPusakaAttendanceTelegramLogsRow, id pgtype.UUID, rawChatID string) bool {
	want := rawChatID[len(rawChatID)-4:]
	for _, row := range rows {
		if row.ID == id {
			return row.TargetChatIDMasked != rawChatID && len(row.TargetChatIDMasked) == len(rawChatID) && row.TargetChatIDMasked[len(row.TargetChatIDMasked)-4:] == want
		}
	}
	return false
}
