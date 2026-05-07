package service

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type employeeStore interface {
	ListEmployees(ctx context.Context) ([]db.ListEmployeesRow, error)
	ListActiveEmployees(ctx context.Context) ([]db.ListActiveEmployeesRow, error)
	ListPusakaEligibleEmployeesWithStatus(ctx context.Context) ([]db.ListPusakaEligibleEmployeesWithStatusRow, error)
	GetEmployee(ctx context.Context, id pgtype.UUID) (db.GetEmployeeRow, error)
	CreateEmployee(ctx context.Context, arg db.CreateEmployeeParams) (pgtype.UUID, error)
	UpsertPusakaAccount(ctx context.Context, arg db.UpsertPusakaAccountParams) (db.PusakaAccount, error)
	UpdateEmployee(ctx context.Context, arg db.UpdateEmployeeParams) (pgtype.UUID, error)
	DeletePusakaAccountByEmployeeID(ctx context.Context, employeeID pgtype.UUID) error
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
	ListUsersByEmployeeID(ctx context.Context, employeeID pgtype.UUID) ([]db.ListUsersByEmployeeIDRow, error)
	UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) error
	DeleteEmployee(ctx context.Context, id pgtype.UUID) error
	ListEmployeesWithStatus(ctx context.Context) ([]db.ListEmployeesWithStatusRow, error)
	ListEntityAuditLogs(ctx context.Context, arg db.ListEntityAuditLogsParams) ([]db.ListEntityAuditLogsRow, error)
	GetSetting(ctx context.Context, key string) (db.AppSetting, error)
}

type Employee struct {
	q employeeStore
}

func NewEmployee(q *db.Queries) *Employee { return &Employee{q: q} }

func normalizeEmploymentType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pns", "pppk", "honorer":
		return strings.ToLower(strings.TrimSpace(value))
	case "", "lainnya":
		return "lainnya"
	default:
		return ""
	}
}

func pusakaEligible(employmentType string) bool {
	return employmentType == "pns" || employmentType == "pppk"
}

func normalizeEmployeeGender(value string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case "":
		return "", true
	case "L", "P":
		return normalized, true
	default:
		return "", false
	}
}

func normalizeEmployeeNPSN(value string) string {
	normalized := strings.TrimSpace(value)
	if len(normalized) != 8 {
		return DefaultEmployeeAccountNPSN
	}
	for _, r := range normalized {
		if !unicode.IsDigit(r) {
			return DefaultEmployeeAccountNPSN
		}
	}
	return normalized
}

func (s *Employee) employeeIdentityNPSN(ctx context.Context) string {
	setting, err := s.q.GetSetting(ctx, schoolProfilePrefix+"npsn")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DefaultEmployeeAccountNPSN
		}
		return DefaultEmployeeAccountNPSN
	}
	return normalizeEmployeeNPSN(setting.Value)
}

func isPegawaiUIDConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == "uq_employees_pegawai_uid"
}

func (s *Employee) List(ctx context.Context) ([]db.ListEmployeesRow, error) {
	return s.q.ListEmployees(ctx)
}

func (s *Employee) ListActive(ctx context.Context) ([]db.ListActiveEmployeesRow, error) {
	return s.q.ListActiveEmployees(ctx)
}

func (s *Employee) ListPusakaEligibleWithStatus(ctx context.Context) ([]db.ListPusakaEligibleEmployeesWithStatusRow, error) {
	return s.q.ListPusakaEligibleEmployeesWithStatus(ctx)
}

func (s *Employee) Get(ctx context.Context, id pgtype.UUID) (db.GetEmployeeRow, error) {
	return s.q.GetEmployee(ctx, id)
}

func (s *Employee) Create(ctx context.Context, nip, nama, unitKerja, employmentType string, tanggalLahir pgtype.Date, jenisKelamin, tempatLahir, pusakaUsername, pusakaPassword string, isActive bool) (db.GetEmployeeRow, error) {
	normalizedType := normalizeEmploymentType(employmentType)
	if normalizedType == "" {
		return db.GetEmployeeRow{}, errors.New("invalid employment type")
	}
	normalizedGender, ok := normalizeEmployeeGender(jenisKelamin)
	if !ok {
		return db.GetEmployeeRow{}, errors.New("invalid gender")
	}
	if (pusakaUsername != "" || pusakaPassword != "") && !pusakaEligible(normalizedType) {
		return db.GetEmployeeRow{}, errors.New("only pns or pppk employees can have pusaka accounts")
	}
	params := db.CreateEmployeeParams{
		Npsn:           s.employeeIdentityNPSN(ctx),
		Nip:            strings.TrimSpace(nip),
		Nama:           strings.TrimSpace(nama),
		UnitKerja:      strings.TrimSpace(unitKerja),
		EmploymentType: normalizedType,
		TanggalLahir:   tanggalLahir,
		JenisKelamin:   normalizedGender,
		TempatLahir:    strings.TrimSpace(tempatLahir),
		IsActive:       isActive,
	}
	var employeeID pgtype.UUID
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		employeeID, err = s.q.CreateEmployee(ctx, params)
		if err == nil {
			break
		}
		if !isPegawaiUIDConflict(err) {
			return db.GetEmployeeRow{}, err
		}
	}
	if err != nil {
		return db.GetEmployeeRow{}, err
	}
	if pusakaUsername != "" || pusakaPassword != "" {
		if _, err := s.q.UpsertPusakaAccount(ctx, db.UpsertPusakaAccountParams{
			EmployeeID:     employeeID,
			PusakaUsername: pusakaUsername,
			PusakaPassword: pusakaPassword,
			IsEnabled:      true,
		}); err != nil {
			return db.GetEmployeeRow{}, err
		}
		return s.q.GetEmployee(ctx, employeeID)
	}
	return s.q.GetEmployee(ctx, employeeID)
}

func (s *Employee) Update(ctx context.Context, p db.UpdateEmployeeParams) (db.GetEmployeeRow, error) {
	p.EmploymentType = normalizeEmploymentType(p.EmploymentType)
	if p.EmploymentType == "" {
		return db.GetEmployeeRow{}, errors.New("invalid employment type")
	}
	normalizedGender, ok := normalizeEmployeeGender(p.JenisKelamin)
	if !ok {
		return db.GetEmployeeRow{}, errors.New("invalid gender")
	}
	p.Nip = strings.TrimSpace(p.Nip)
	p.Nama = strings.TrimSpace(p.Nama)
	p.UnitKerja = strings.TrimSpace(p.UnitKerja)
	p.JenisKelamin = normalizedGender
	p.TempatLahir = strings.TrimSpace(p.TempatLahir)
	existing, err := s.q.GetEmployee(ctx, p.ID)
	if err != nil {
		return db.GetEmployeeRow{}, err
	}
	if !pusakaEligible(p.EmploymentType) && existing.PusakaUsername != "" {
		return db.GetEmployeeRow{}, errors.New("disable or remove the pusaka account before changing employee type")
	}
	employeeID, err := s.q.UpdateEmployee(ctx, p)
	if err != nil {
		return db.GetEmployeeRow{}, err
	}
	return s.q.GetEmployee(ctx, employeeID)
}

func (s *Employee) UpsertPusakaAccount(ctx context.Context, employeeID pgtype.UUID, username, password string, isEnabled bool) error {
	employee, err := s.q.GetEmployee(ctx, employeeID)
	if err != nil {
		return err
	}
	if !pusakaEligible(employee.EmploymentType) {
		return errors.New("only pns or pppk employees can have pusaka accounts")
	}
	_, err = s.q.UpsertPusakaAccount(ctx, db.UpsertPusakaAccountParams{
		EmployeeID:     employeeID,
		PusakaUsername: username,
		PusakaPassword: password,
		IsEnabled:      isEnabled,
	})
	return err
}

func (s *Employee) SetPusakaAccountEnabled(ctx context.Context, employeeID pgtype.UUID, isEnabled bool) error {
	employee, err := s.q.GetEmployee(ctx, employeeID)
	if err != nil {
		return err
	}
	if employee.PusakaUsername == "" {
		return errors.New("pusaka account is not configured")
	}
	return s.UpsertPusakaAccount(ctx, employeeID, employee.PusakaUsername, employee.PusakaPassword, isEnabled)
}

func (s *Employee) DeletePusakaAccount(ctx context.Context, employeeID pgtype.UUID) error {
	employee, err := s.q.GetEmployee(ctx, employeeID)
	if err != nil {
		return err
	}
	if employee.PusakaUsername == "" {
		return errors.New("pusaka account is not configured")
	}
	return s.q.DeletePusakaAccountByEmployeeID(ctx, employeeID)
}

func (s *Employee) CreateAuditLog(ctx context.Context, userID pgtype.UUID, action, entityType, entityID string, metadata []byte) error {
	_, err := s.q.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   metadata,
	})
	return err
}

func (s *Employee) SetActive(ctx context.Context, id pgtype.UUID, isActive bool) error {
	emp, err := s.q.GetEmployee(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.q.UpdateEmployee(ctx, db.UpdateEmployeeParams{
		ID:             emp.ID,
		Nip:            emp.Nip,
		Nama:           emp.Nama,
		UnitKerja:      emp.UnitKerja,
		EmploymentType: emp.EmploymentType,
		TanggalLahir:   emp.TanggalLahir,
		JenisKelamin:   emp.JenisKelamin,
		TempatLahir:    emp.TempatLahir,
		IsActive:       isActive,
	}); err != nil {
		return err
	}
	if emp.PusakaUsername != "" || emp.PusakaPassword != "" {
		if err := s.UpsertPusakaAccount(ctx, id, emp.PusakaUsername, emp.PusakaPassword, isActive); err != nil {
			return err
		}
	}
	users, err := s.q.ListUsersByEmployeeID(ctx, id)
	if err != nil {
		return err
	}
	for _, user := range users {
		if err := s.q.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
			ID:       user.ID,
			IsActive: isActive,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Employee) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteEmployee(ctx, id)
}

func (s *Employee) ListWithStatus(ctx context.Context) ([]db.ListEmployeesWithStatusRow, error) {
	return s.q.ListEmployeesWithStatus(ctx)
}

func (s *Employee) ListPusakaAuditLogs(ctx context.Context, employeeID string, limit, offset int32) ([]db.ListEntityAuditLogsRow, error) {
	return s.q.ListEntityAuditLogs(ctx, db.ListEntityAuditLogsParams{
		EntityType: "pusaka_account",
		EntityID:   employeeID,
		Limit:      limit,
		Offset:     offset,
	})
}
