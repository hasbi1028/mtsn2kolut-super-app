package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const DefaultEmployeeAccountNPSN = "40406031"
const DefaultEmployeeAccountRole = "guru"

type employeeAccountGenerationStore interface {
	ListEmployeeAccountGenerationCandidates(ctx context.Context, npsn string) ([]db.ListEmployeeAccountGenerationCandidatesRow, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error)
	AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error
	AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

type EmployeeAccountGenerator struct {
	q    employeeAccountGenerationStore
	tx   userLifecycleTxStarter
	npsn string
	role string
}

type EmployeeAccountGenerationItem struct {
	EmployeeID     string `json:"employee_id"`
	NIP            string `json:"nip"`
	Nama           string `json:"nama"`
	TanggalLahir   string `json:"tanggal_lahir,omitempty"`
	NomorUrut      int32  `json:"nomor_urut"`
	Username       string `json:"username"`
	Password       string `json:"password,omitempty"`
	Role           string `json:"role"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	ExistingUserID string `json:"existing_user_id,omitempty"`
	UsernameUserID string `json:"username_user_id,omitempty"`
}

type EmployeeAccountGenerationResult struct {
	NPSN               string                          `json:"npsn"`
	DefaultRole        string                          `json:"default_role"`
	PasswordSameAsUser bool                            `json:"password_same_as_username"`
	Total              int                             `json:"total"`
	Ready              int                             `json:"ready"`
	Created            int                             `json:"created"`
	Skipped            int                             `json:"skipped"`
	Failed             int                             `json:"failed"`
	Items              []EmployeeAccountGenerationItem `json:"items"`
}

func NewEmployeeAccountGenerator(q *db.Queries) *EmployeeAccountGenerator {
	return &EmployeeAccountGenerator{q: q, npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}
}

func NewEmployeeAccountGeneratorWithPool(pool *pgxpool.Pool) *EmployeeAccountGenerator {
	return &EmployeeAccountGenerator{q: db.New(pool), tx: pool, npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}
}

func (s *EmployeeAccountGenerator) Preview(ctx context.Context) (EmployeeAccountGenerationResult, error) {
	rows, err := s.q.ListEmployeeAccountGenerationCandidates(ctx, s.npsn)
	if err != nil {
		return EmployeeAccountGenerationResult{}, err
	}
	return s.buildResult(rows, false), nil
}

func (s *EmployeeAccountGenerator) Generate(ctx context.Context, actorID pgtype.UUID) (EmployeeAccountGenerationResult, error) {
	rows, err := s.q.ListEmployeeAccountGenerationCandidates(ctx, s.npsn)
	if err != nil {
		return EmployeeAccountGenerationResult{}, err
	}
	result := s.buildResult(rows, true)
	if result.Ready == 0 {
		return result, nil
	}

	mutate := func(store employeeAccountGenerationStore) error {
		for i := range result.Items {
			item := &result.Items[i]
			if item.Status != "ready" {
				continue
			}
			employeeID, err := uuidFromString(item.EmployeeID)
			if err != nil {
				item.Status = "failed"
				item.Message = "employee_id tidak valid"
				result.Failed++
				result.Ready--
				continue
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(item.Username), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			row, err := store.CreateUser(ctx, db.CreateUserParams{
				Username:     item.Username,
				PasswordHash: string(hash),
				DisplayName:  pgtype.Text{String: item.Nama, Valid: item.Nama != ""},
				EmployeeID:   employeeID,
				IsActive:     true,
			})
			if err != nil {
				item.Status = "failed"
				item.Message = err.Error()
				result.Failed++
				result.Ready--
				continue
			}
			if err := store.AddUserRole(ctx, db.AddUserRoleParams{UserID: row.ID, Role: db.UserRole(s.role)}); err != nil {
				return err
			}
			if err := store.AddUserRbacRoleByCode(ctx, db.AddUserRbacRoleByCodeParams{UserID: row.ID, Code: s.role}); err != nil {
				return err
			}
			if err := auditEmployeeAccountGeneration(ctx, store, actorID, row.ID, item.EmployeeID, item.Username, s.role); err != nil {
				return err
			}
			item.Status = "created"
			item.Message = "akun dibuat"
			item.Password = item.Username
			item.ExistingUserID = uuidEntityID(row.ID)
			result.Created++
			result.Ready--
		}
		return nil
	}
	if s.tx == nil {
		err = mutate(s.q)
	} else {
		tx, txErr := s.tx.Begin(ctx)
		if txErr != nil {
			return result, txErr
		}
		defer tx.Rollback(ctx)
		err = mutate(db.New(tx))
		if err == nil {
			err = tx.Commit(ctx)
		}
	}
	return result, err
}

func (s *EmployeeAccountGenerator) buildResult(rows []db.ListEmployeeAccountGenerationCandidatesRow, includePassword bool) EmployeeAccountGenerationResult {
	result := EmployeeAccountGenerationResult{NPSN: s.npsn, DefaultRole: s.role, PasswordSameAsUser: true, Items: make([]EmployeeAccountGenerationItem, 0, len(rows))}
	for _, row := range rows {
		item := EmployeeAccountGenerationItem{EmployeeID: uuidEntityID(row.EmployeeID), NIP: row.Nip, Nama: row.Nama, TanggalLahir: accountGenerationDateString(row.TanggalLahir), NomorUrut: row.NomorUrut, Username: row.GeneratedUsername, Role: s.role}
		switch {
		case !row.TanggalLahir.Valid:
			item.Status = "skipped"
			item.Message = "tanggal lahir belum diisi"
			result.Skipped++
		case row.ExistingUserID.Valid:
			item.Status = "skipped"
			item.Message = "pegawai sudah punya akun"
			item.ExistingUserID = uuidEntityID(row.ExistingUserID)
			result.Skipped++
		case row.UsernameUserID.Valid:
			item.Status = "skipped"
			item.Message = "username sudah dipakai akun lain"
			item.UsernameUserID = uuidEntityID(row.UsernameUserID)
			result.Skipped++
		default:
			item.Status = "ready"
			item.Message = "siap dibuat"
			if includePassword {
				item.Password = item.Username
			}
			result.Ready++
		}
		result.Items = append(result.Items, item)
	}
	result.Total = len(result.Items)
	return result
}

func uuidFromString(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid uuid %q: %w", value, err)
	}
	return id, nil
}

func accountGenerationDateString(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func auditEmployeeAccountGeneration(ctx context.Context, store employeeAccountGenerationStore, actorID pgtype.UUID, userID pgtype.UUID, employeeID string, username string, role string) error {
	payload, err := json.Marshal(map[string]any{"user_id": uuidEntityID(userID), "username": username, "employee_id": employeeID, "role": role})
	if err != nil {
		return err
	}
	_, err = store.CreateAuditLog(ctx, db.CreateAuditLogParams{UserID: actorID, Action: "USER_EMPLOYEE_ACCOUNT_GENERATED", EntityType: "user", EntityID: uuidEntityID(userID), Metadata: payload})
	return err
}

var _ employeeAccountGenerationStore = (*db.Queries)(nil)
var _ userLifecycleTxStarter = (*pgxpool.Pool)(nil)
