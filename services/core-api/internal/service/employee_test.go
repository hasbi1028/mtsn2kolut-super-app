package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestEmployeeNormalizeEmploymentTypeAndEligibility(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		wantType     string
		wantEligible bool
	}{
		{name: "pns", value: " PNS ", wantType: "pns", wantEligible: true},
		{name: "pppk", value: "pppk", wantType: "pppk", wantEligible: true},
		{name: "honorer", value: "HONORER", wantType: "honorer", wantEligible: false},
		{name: "empty defaults lainnya", value: "", wantType: "lainnya", wantEligible: false},
		{name: "lainnya", value: "lainnya", wantType: "lainnya", wantEligible: false},
		{name: "invalid", value: "kontrak", wantType: "", wantEligible: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeEmploymentType(tt.value)
			if got != tt.wantType {
				t.Fatalf("normalizeEmploymentType(%q) = %q, want %q", tt.value, got, tt.wantType)
			}
			if eligible := pusakaEligible(got); eligible != tt.wantEligible {
				t.Fatalf("pusakaEligible(%q) = %v, want %v", got, eligible, tt.wantEligible)
			}
		})
	}
}

type fakeEmployeeStore struct {
	listRows        []db.ListEmployeesRow
	activeRows      []db.ListActiveEmployeesRow
	eligibleRows    []db.ListPusakaEligibleEmployeesWithStatusRow
	employee        db.GetEmployeeRow
	getID           pgtype.UUID
	getErr          error
	createID        pgtype.UUID
	createArg       db.CreateEmployeeParams
	createErr       error
	upsertArgs      []db.UpsertPusakaAccountParams
	upsertErr       error
	updateArg       db.UpdateEmployeeParams
	updateErr       error
	deletePusakaID  pgtype.UUID
	auditArg        db.CreateAuditLogParams
	usersByEmployee []db.ListUsersByEmployeeIDRow
	usersErr        error
	userStatusArgs  []db.UpdateUserStatusParams
	userStatusErr   error
	deleteID        pgtype.UUID
	withStatusRows  []db.ListEmployeesWithStatusRow
	entityAuditArg  db.ListEntityAuditLogsParams
	entityAuditRows []db.ListEntityAuditLogsRow
	setting         db.AppSetting
	settingErr      error
}

func (f *fakeEmployeeStore) ListEmployees(ctx context.Context) ([]db.ListEmployeesRow, error) {
	return f.listRows, nil
}

func (f *fakeEmployeeStore) ListActiveEmployees(ctx context.Context) ([]db.ListActiveEmployeesRow, error) {
	return f.activeRows, nil
}

func (f *fakeEmployeeStore) ListPusakaEligibleEmployeesWithStatus(ctx context.Context) ([]db.ListPusakaEligibleEmployeesWithStatusRow, error) {
	return f.eligibleRows, nil
}

func (f *fakeEmployeeStore) GetEmployee(ctx context.Context, id pgtype.UUID) (db.GetEmployeeRow, error) {
	f.getID = id
	if f.getErr != nil {
		return db.GetEmployeeRow{}, f.getErr
	}
	employee := f.employee
	if !employee.ID.Valid {
		employee.ID = id
	}
	return employee, nil
}

func (f *fakeEmployeeStore) CreateEmployee(ctx context.Context, arg db.CreateEmployeeParams) (pgtype.UUID, error) {
	f.createArg = arg
	if f.createErr != nil {
		return pgtype.UUID{}, f.createErr
	}
	return f.createID, nil
}

func (f *fakeEmployeeStore) UpsertPusakaAccount(ctx context.Context, arg db.UpsertPusakaAccountParams) (db.PusakaAccount, error) {
	f.upsertArgs = append(f.upsertArgs, arg)
	if f.upsertErr != nil {
		return db.PusakaAccount{}, f.upsertErr
	}
	return db.PusakaAccount{EmployeeID: arg.EmployeeID, PusakaUsername: arg.PusakaUsername, PusakaPassword: arg.PusakaPassword, IsEnabled: arg.IsEnabled}, nil
}

func (f *fakeEmployeeStore) UpdateEmployee(ctx context.Context, arg db.UpdateEmployeeParams) (pgtype.UUID, error) {
	f.updateArg = arg
	if f.updateErr != nil {
		return pgtype.UUID{}, f.updateErr
	}
	return arg.ID, nil
}

func (f *fakeEmployeeStore) DeletePusakaAccountByEmployeeID(ctx context.Context, employeeID pgtype.UUID) error {
	f.deletePusakaID = employeeID
	return nil
}

func (f *fakeEmployeeStore) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.auditArg = arg
	return db.AuditLog{UserID: arg.UserID, Action: arg.Action, EntityType: arg.EntityType, EntityID: arg.EntityID}, nil
}

func (f *fakeEmployeeStore) ListUsersByEmployeeID(ctx context.Context, employeeID pgtype.UUID) ([]db.ListUsersByEmployeeIDRow, error) {
	return f.usersByEmployee, f.usersErr
}

func (f *fakeEmployeeStore) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) error {
	f.userStatusArgs = append(f.userStatusArgs, arg)
	return f.userStatusErr
}

func (f *fakeEmployeeStore) DeleteEmployee(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return nil
}

func (f *fakeEmployeeStore) ListEmployeesWithStatus(ctx context.Context) ([]db.ListEmployeesWithStatusRow, error) {
	return f.withStatusRows, nil
}

func (f *fakeEmployeeStore) ListEntityAuditLogs(ctx context.Context, arg db.ListEntityAuditLogsParams) ([]db.ListEntityAuditLogsRow, error) {
	f.entityAuditArg = arg
	return f.entityAuditRows, nil
}

func (f *fakeEmployeeStore) GetSetting(ctx context.Context, key string) (db.AppSetting, error) {
	if f.settingErr != nil {
		return db.AppSetting{}, f.settingErr
	}
	if f.setting.Key == "" {
		return db.AppSetting{Key: key, Value: DefaultEmployeeAccountNPSN}, nil
	}
	return f.setting, nil
}

func TestEmployeeServiceForwardsStoreCallsAndPusakaRules(t *testing.T) {
	employeeID := documentCycleTestUUID(241)
	userID := documentCycleTestUUID(242)
	otherUserID := documentCycleTestUUID(243)
	birthDate := pgtype.Date{Time: time.Date(1986, 5, 4, 0, 0, 0, 0, time.UTC), Valid: true}
	store := &fakeEmployeeStore{
		listRows:        []db.ListEmployeesRow{{ID: employeeID, Nama: "Guru"}},
		activeRows:      []db.ListActiveEmployeesRow{{ID: employeeID, Nama: "Guru"}},
		eligibleRows:    []db.ListPusakaEligibleEmployeesWithStatusRow{{ID: employeeID, Nama: "Guru", PusakaEligible: true}},
		employee:        db.GetEmployeeRow{ID: employeeID, Nip: "1980", Nama: "Guru", UnitKerja: "MTsN", EmploymentType: "pns", PusakaUsername: "guru", PusakaPassword: "secret", IsActive: true},
		createID:        employeeID,
		usersByEmployee: []db.ListUsersByEmployeeIDRow{{ID: userID}, {ID: otherUserID}},
		withStatusRows:  []db.ListEmployeesWithStatusRow{{ID: employeeID, Nama: "Guru"}},
		entityAuditRows: []db.ListEntityAuditLogsRow{{EntityID: "emp-1"}},
	}
	svc := &Employee{q: store}
	if NewEmployee(nil) == nil {
		t.Fatal("NewEmployee(nil) = nil, want service")
	}

	if rows, err := svc.List(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("List() = %d rows/%v, want 1 nil", len(rows), err)
	}
	if rows, err := svc.ListActive(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("ListActive() = %d rows/%v, want 1 nil", len(rows), err)
	}
	if rows, err := svc.ListPusakaEligibleWithStatus(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("ListPusakaEligibleWithStatus() = %d rows/%v, want 1 nil", len(rows), err)
	}
	if got, err := svc.Get(context.Background(), employeeID); err != nil || got.ID != employeeID {
		t.Fatalf("Get() = %+v/%v, want employee", got, err)
	}
	if _, err := svc.Create(context.Background(), "1980", "Guru", "MTsN", " PPPK ", birthDate, "L", "Kolaka", "akun", "rahasia", true); err != nil {
		t.Fatalf("Create(pppk pusaka) error = %v", err)
	}
	if store.createArg.EmploymentType != "pppk" || store.createArg.Nama != "Guru" || store.createArg.TanggalLahir != birthDate || store.createArg.JenisKelamin != "L" || store.createArg.TempatLahir != "Kolaka" || store.createArg.Npsn != DefaultEmployeeAccountNPSN {
		t.Fatalf("Create() arg = %+v, want normalized employee", store.createArg)
	}
	if len(store.upsertArgs) != 1 || store.upsertArgs[0].EmployeeID != employeeID || store.upsertArgs[0].PusakaUsername != "akun" || !store.upsertArgs[0].IsEnabled {
		t.Fatalf("Create() upsert args = %+v, want pusaka account", store.upsertArgs)
	}
	if _, err := svc.Create(context.Background(), "", "Staf", "TU", "honorer", pgtype.Date{}, "", "", "", "", true); err != nil {
		t.Fatalf("Create(honorer no pusaka) error = %v", err)
	}
	if store.createArg.Nip != "" {
		t.Fatalf("Create(honorer) NIP = %q, want optional empty NIP", store.createArg.Nip)
	}
	if _, err := svc.Create(context.Background(), "1982", "Kontrak", "TU", "kontrak", pgtype.Date{}, "", "", "", "", true); err == nil || err.Error() != "invalid employment type" {
		t.Fatalf("Create(invalid type) = %v, want invalid employment type", err)
	}
	if _, err := svc.Create(context.Background(), "1983", "Honorer", "TU", "honorer", pgtype.Date{}, "X", "", "", "", true); err == nil || err.Error() != "invalid gender" {
		t.Fatalf("Create(invalid gender) = %v, want invalid gender", err)
	}
	if _, err := svc.Create(context.Background(), "1983", "Honorer", "TU", "honorer", pgtype.Date{}, "", "", "akun", "", true); err == nil || err.Error() != "only pns or pppk employees can have pusaka accounts" {
		t.Fatalf("Create(noneligible pusaka) = %v, want eligibility error", err)
	}

	if _, err := svc.Update(context.Background(), db.UpdateEmployeeParams{ID: employeeID, Nip: "1980", Nama: "Guru Baru", UnitKerja: "MTsN", EmploymentType: "PNS", IsActive: true}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if store.updateArg.EmploymentType != "pns" || store.updateArg.Nama != "Guru Baru" {
		t.Fatalf("Update() arg = %+v, want normalized update", store.updateArg)
	}
	if _, err := svc.Update(context.Background(), db.UpdateEmployeeParams{ID: employeeID, EmploymentType: "pns", JenisKelamin: "X"}); err == nil || err.Error() != "invalid gender" {
		t.Fatalf("Update(invalid gender) = %v, want invalid gender", err)
	}
	if _, err := svc.Update(context.Background(), db.UpdateEmployeeParams{ID: employeeID, EmploymentType: "kontrak"}); err == nil || err.Error() != "invalid employment type" {
		t.Fatalf("Update(invalid type) = %v, want invalid employment type", err)
	}
	if _, err := svc.Update(context.Background(), db.UpdateEmployeeParams{ID: employeeID, EmploymentType: "honorer"}); err == nil || err.Error() != "disable or remove the pusaka account before changing employee type" {
		t.Fatalf("Update(remove eligibility with pusaka) = %v, want account removal error", err)
	}

	store.employee.EmploymentType = "honorer"
	if err := svc.UpsertPusakaAccount(context.Background(), employeeID, "akun", "secret", true); err == nil || err.Error() != "only pns or pppk employees can have pusaka accounts" {
		t.Fatalf("UpsertPusakaAccount(honorer) = %v, want eligibility error", err)
	}
	store.employee.EmploymentType = "pns"
	if err := svc.UpsertPusakaAccount(context.Background(), employeeID, "akun2", "secret2", false); err != nil {
		t.Fatalf("UpsertPusakaAccount(pns) error = %v", err)
	}
	if got := store.upsertArgs[len(store.upsertArgs)-1]; got.PusakaUsername != "akun2" || got.IsEnabled {
		t.Fatalf("UpsertPusakaAccount() arg = %+v, want disabled akun2", got)
	}

	store.employee.PusakaUsername = ""
	if err := svc.SetPusakaAccountEnabled(context.Background(), employeeID, false); err == nil || err.Error() != "pusaka account is not configured" {
		t.Fatalf("SetPusakaAccountEnabled(no account) = %v, want configured error", err)
	}
	store.employee.PusakaUsername = "guru"
	if err := svc.SetPusakaAccountEnabled(context.Background(), employeeID, false); err != nil {
		t.Fatalf("SetPusakaAccountEnabled() error = %v", err)
	}
	if got := store.upsertArgs[len(store.upsertArgs)-1]; got.PusakaUsername != "guru" || got.IsEnabled {
		t.Fatalf("SetPusakaAccountEnabled() upsert = %+v, want disabled existing account", got)
	}

	store.employee.PusakaUsername = ""
	if err := svc.DeletePusakaAccount(context.Background(), employeeID); err == nil || err.Error() != "pusaka account is not configured" {
		t.Fatalf("DeletePusakaAccount(no account) = %v, want configured error", err)
	}
	store.employee.PusakaUsername = "guru"
	if err := svc.DeletePusakaAccount(context.Background(), employeeID); err != nil || store.deletePusakaID != employeeID {
		t.Fatalf("DeletePusakaAccount() = %v id=%v, want nil/%v", err, store.deletePusakaID, employeeID)
	}
	if err := svc.CreateAuditLog(context.Background(), userID, "pusaka.updated", "pusaka_account", "emp-1", []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("CreateAuditLog() error = %v", err)
	}
	if store.auditArg.UserID != userID || store.auditArg.Action != "pusaka.updated" || store.auditArg.EntityID != "emp-1" {
		t.Fatalf("CreateAuditLog() arg = %+v, want audit fields", store.auditArg)
	}

	store.userStatusArgs = nil
	if err := svc.SetActive(context.Background(), employeeID, false); err != nil {
		t.Fatalf("SetActive(false) error = %v", err)
	}
	if store.updateArg.ID != employeeID || store.updateArg.IsActive {
		t.Fatalf("SetActive() update = %+v, want inactive employee", store.updateArg)
	}
	if got := store.upsertArgs[len(store.upsertArgs)-1]; got.EmployeeID != employeeID || got.IsEnabled {
		t.Fatalf("SetActive() pusaka update = %+v, want disabled account", got)
	}
	if len(store.userStatusArgs) != 2 || store.userStatusArgs[0].IsActive || store.userStatusArgs[1].ID != otherUserID {
		t.Fatalf("SetActive() user updates = %+v, want two inactive users", store.userStatusArgs)
	}
	if err := svc.Delete(context.Background(), employeeID); err != nil || store.deleteID != employeeID {
		t.Fatalf("Delete() = %v id=%v, want nil/%v", err, store.deleteID, employeeID)
	}
	if rows, err := svc.ListWithStatus(context.Background()); err != nil || len(rows) != 1 {
		t.Fatalf("ListWithStatus() = %d rows/%v, want 1 nil", len(rows), err)
	}
	if rows, err := svc.ListPusakaAuditLogs(context.Background(), "emp-1", 10, 5); err != nil || len(rows) != 1 {
		t.Fatalf("ListPusakaAuditLogs() = %d rows/%v, want 1 nil", len(rows), err)
	}
	if store.entityAuditArg.EntityType != "pusaka_account" || store.entityAuditArg.EntityID != "emp-1" || store.entityAuditArg.Limit != 10 || store.entityAuditArg.Offset != 5 {
		t.Fatalf("ListPusakaAuditLogs() arg = %+v, want pusaka account audit paging", store.entityAuditArg)
	}
}

func TestEmployeeServicePropagatesStoreErrors(t *testing.T) {
	employeeID := documentCycleTestUUID(251)
	store := &fakeEmployeeStore{employee: db.GetEmployeeRow{ID: employeeID, EmploymentType: "pns"}}
	svc := &Employee{q: store}

	store.createErr = errors.New("create failed")
	if _, err := svc.Create(context.Background(), "1", "Guru", "MTsN", "pns", pgtype.Date{}, "", "", "", "", true); err == nil || err.Error() != "create failed" {
		t.Fatalf("Create(store error) = %v, want create failed", err)
	}
	store.createErr = nil
	store.upsertErr = errors.New("upsert failed")
	if _, err := svc.Create(context.Background(), "1", "Guru", "MTsN", "pns", pgtype.Date{}, "", "", "akun", "secret", true); err == nil || err.Error() != "upsert failed" {
		t.Fatalf("Create(upsert error) = %v, want upsert failed", err)
	}
	store.upsertErr = nil
	store.getErr = errors.New("get failed")
	if _, err := svc.Update(context.Background(), db.UpdateEmployeeParams{ID: employeeID, EmploymentType: "pns"}); err == nil || err.Error() != "get failed" {
		t.Fatalf("Update(get error) = %v, want get failed", err)
	}
	if err := svc.UpsertPusakaAccount(context.Background(), employeeID, "akun", "secret", true); err == nil || err.Error() != "get failed" {
		t.Fatalf("UpsertPusakaAccount(get error) = %v, want get failed", err)
	}
	if err := svc.SetActive(context.Background(), employeeID, true); err == nil || err.Error() != "get failed" {
		t.Fatalf("SetActive(get error) = %v, want get failed", err)
	}
	store.getErr = nil
	store.updateErr = errors.New("update failed")
	if _, err := svc.Update(context.Background(), db.UpdateEmployeeParams{ID: employeeID, EmploymentType: "pns"}); err == nil || err.Error() != "update failed" {
		t.Fatalf("Update(update error) = %v, want update failed", err)
	}
	if err := svc.SetActive(context.Background(), employeeID, true); err == nil || err.Error() != "update failed" {
		t.Fatalf("SetActive(update error) = %v, want update failed", err)
	}
	store.updateErr = nil
	store.usersErr = errors.New("users failed")
	if err := svc.SetActive(context.Background(), employeeID, true); err == nil || err.Error() != "users failed" {
		t.Fatalf("SetActive(users error) = %v, want users failed", err)
	}
	store.usersErr = nil
	store.usersByEmployee = []db.ListUsersByEmployeeIDRow{{ID: documentCycleTestUUID(252)}}
	store.userStatusErr = errors.New("status failed")
	if err := svc.SetActive(context.Background(), employeeID, true); err == nil || err.Error() != "status failed" {
		t.Fatalf("SetActive(user status error) = %v, want status failed", err)
	}
}

type fakeEmployeeScheduleStore struct {
	listID    pgtype.UUID
	listRows  []db.ListEmployeeSchedulesRow
	upsertArg db.UpsertEmployeeScheduleParams
	deleteArg db.DeleteEmployeeScheduleParams
}

func (f *fakeEmployeeScheduleStore) ListEmployeeSchedules(ctx context.Context, employeeID pgtype.UUID) ([]db.ListEmployeeSchedulesRow, error) {
	f.listID = employeeID
	return f.listRows, nil
}

func (f *fakeEmployeeScheduleStore) UpsertEmployeeSchedule(ctx context.Context, arg db.UpsertEmployeeScheduleParams) (db.EmployeeSchedule, error) {
	f.upsertArg = arg
	return db.EmployeeSchedule{EmployeeID: arg.EmployeeID, RunType: arg.RunType, RunTime: arg.RunTime, IsEnabled: arg.IsEnabled, RandomWindowMinutes: arg.RandomWindowMinutes, DayOfWeek: arg.DayOfWeek}, nil
}

func (f *fakeEmployeeScheduleStore) DeleteEmployeeSchedule(ctx context.Context, arg db.DeleteEmployeeScheduleParams) error {
	f.deleteArg = arg
	return nil
}

func TestEmployeeScheduleServiceForwardsStoreCalls(t *testing.T) {
	employeeID := documentCycleTestUUID(253)
	scheduleID := documentCycleTestUUID(254)
	store := &fakeEmployeeScheduleStore{listRows: []db.ListEmployeeSchedulesRow{{EmployeeID: employeeID, RunType: db.RunTypeEnumMorning}}}
	svc := &EmployeeSchedule{q: store}
	if NewEmployeeSchedule(nil) == nil {
		t.Fatal("NewEmployeeSchedule(nil) = nil, want service")
	}

	if rows, err := svc.List(context.Background(), employeeID); err != nil || len(rows) != 1 || store.listID != employeeID {
		t.Fatalf("List() = %d rows/%v id=%v, want schedule", len(rows), err, store.listID)
	}
	if _, err := svc.Upsert(context.Background(), db.UpsertEmployeeScheduleParams{EmployeeID: employeeID, RunType: db.RunTypeEnumMorning, RunTime: "07:00", IsEnabled: true, RandomWindowMinutes: 10, DayOfWeek: 1}); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if store.upsertArg.EmployeeID != employeeID || store.upsertArg.RunTime != "07:00" || store.upsertArg.RandomWindowMinutes != 10 {
		t.Fatalf("Upsert() arg = %+v, want forwarded schedule", store.upsertArg)
	}
	if err := svc.Delete(context.Background(), scheduleID, employeeID); err != nil || store.deleteArg.ID != scheduleID || store.deleteArg.EmployeeID != employeeID {
		t.Fatalf("Delete() = %v arg=%+v, want ids", err, store.deleteArg)
	}
}
