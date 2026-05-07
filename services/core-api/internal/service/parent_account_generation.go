package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const ParentAccountRole = "ortu"

type parentAccountGenerationStore interface {
	ListParentAccountGenerationCandidates(ctx context.Context) ([]db.ListParentAccountGenerationCandidatesRow, error)
	CreateUserWithMustChangePassword(ctx context.Context, arg db.CreateUserWithMustChangePasswordParams) (db.CreateUserWithMustChangePasswordRow, error)
	AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error
	AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

type ParentAccountGenerator struct {
	q                 parentAccountGenerationStore
	tx                userLifecycleTxStarter
	role              string
	passwordGenerator func() (string, error)
}

type ParentAccountGenerationCandidate struct {
	ParentID          string `json:"parent_id"`
	Nama              string `json:"nama"`
	Phone             string `json:"phone,omitempty"`
	ChildCount        int32  `json:"child_count"`
	BasisStudentID    string `json:"basis_student_id,omitempty"`
	BasisStudentNISN  string `json:"basis_student_nisn,omitempty"`
	GeneratedUsername string `json:"generated_username,omitempty"`
	TemporaryPassword string `json:"temporary_password,omitempty"`
	Role              string `json:"role"`
	Status            string `json:"status"`
	Reason            string `json:"reason,omitempty"`
	UserID            string `json:"user_id,omitempty"`
	ExistingUserID    string `json:"existing_user_id,omitempty"`
}

type ParentAccountGenerationResult struct {
	Role       string                             `json:"role"`
	Total      int                                `json:"total"`
	Ready      int                                `json:"ready"`
	Created    int                                `json:"created"`
	Skipped    int                                `json:"skipped"`
	Failed     int                                `json:"failed"`
	Candidates []ParentAccountGenerationCandidate `json:"candidates"`
}

func NewParentAccountGenerator(q *db.Queries) *ParentAccountGenerator {
	return &ParentAccountGenerator{q: q, role: ParentAccountRole}
}

func NewParentAccountGeneratorWithPool(pool *pgxpool.Pool) *ParentAccountGenerator {
	return &ParentAccountGenerator{q: db.New(pool), tx: pool, role: ParentAccountRole}
}

func (s *ParentAccountGenerator) Preview(ctx context.Context) (ParentAccountGenerationResult, error) {
	rows, err := s.q.ListParentAccountGenerationCandidates(ctx)
	if err != nil {
		return ParentAccountGenerationResult{}, err
	}
	return s.buildResult(rows), nil
}

func (s *ParentAccountGenerator) Generate(ctx context.Context, actorID pgtype.UUID) (ParentAccountGenerationResult, error) {
	rows, err := s.q.ListParentAccountGenerationCandidates(ctx)
	if err != nil {
		return ParentAccountGenerationResult{}, err
	}
	result := s.buildResult(rows)
	if result.Ready == 0 {
		return result, nil
	}

	mutate := func(store parentAccountGenerationStore) error {
		for i := range result.Candidates {
			item := &result.Candidates[i]
			if item.Status != "ready" {
				continue
			}
			parentID, err := uuidFromString(item.ParentID)
			if err != nil {
				markParentAccountGenerationFailed(&result, item, "parent_id tidak valid")
				continue
			}
			password, err := s.generatePassword()
			if err != nil {
				return err
			}
			if err := ValidatePassword(item.GeneratedUsername, password); err != nil {
				return err
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			row, err := store.CreateUserWithMustChangePassword(ctx, db.CreateUserWithMustChangePasswordParams{
				Username:           item.GeneratedUsername,
				PasswordHash:       string(hash),
				DisplayName:        pgtype.Text{String: item.Nama, Valid: strings.TrimSpace(item.Nama) != ""},
				ParentID:           parentID,
				IsActive:           true,
				MustChangePassword: true,
			})
			if err != nil {
				markParentAccountGenerationFailed(&result, item, safeAccountGenerationError(err))
				continue
			}
			if err := addLegacyAccountRoleIfSafe(ctx, store, row.ID, s.accountRole()); err != nil {
				return err
			}
			if err := store.AddUserRbacRoleByCode(ctx, db.AddUserRbacRoleByCodeParams{UserID: row.ID, Code: s.accountRole()}); err != nil {
				return err
			}
			if err := auditParentAccountGeneration(ctx, store, actorID, row.ID, parentID, item.GeneratedUsername, s.accountRole()); err != nil {
				return err
			}
			item.Status = "created"
			item.Reason = ""
			item.UserID = uuidEntityID(row.ID)
			item.ExistingUserID = uuidEntityID(row.ID)
			item.TemporaryPassword = password
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

func (s *ParentAccountGenerator) buildResult(rows []db.ListParentAccountGenerationCandidatesRow) ParentAccountGenerationResult {
	result := ParentAccountGenerationResult{Role: s.accountRole(), Candidates: make([]ParentAccountGenerationCandidate, 0, len(rows))}
	takenByBase := map[string]map[string]bool{}
	for _, row := range rows {
		base := strings.TrimSpace(row.BaseUsername)
		item := ParentAccountGenerationCandidate{
			ParentID:         uuidEntityID(row.ParentID),
			Nama:             row.Nama,
			Phone:            row.Phone,
			ChildCount:       row.ChildCount,
			BasisStudentID:   uuidEntityID(row.BasisStudentID),
			BasisStudentNISN: row.BasisStudentNisn,
			Role:             s.accountRole(),
		}
		mergeTakenUsernames(takenByBase, base, row.UsernameCollisions)
		switch {
		case row.ExistingUserID.Valid:
			item.Status = "skipped"
			item.Reason = "orang tua sudah punya akun"
			item.ExistingUserID = uuidEntityID(row.ExistingUserID)
			result.Skipped++
		case row.ChildCount == 0:
			item.Status = "skipped"
			item.Reason = "orang tua belum terhubung ke siswa"
			result.Skipped++
		case base == "":
			item.Status = "skipped"
			item.Reason = "NISN anak/nomor HP belum tersedia"
			result.Skipped++
		default:
			item.Status = "ready"
			item.GeneratedUsername = nextAvailableAccountUsername(base, takenByBase[base])
			result.Ready++
		}
		result.Candidates = append(result.Candidates, item)
	}
	result.Total = len(result.Candidates)
	return result
}

func (s *ParentAccountGenerator) accountRole() string {
	if strings.TrimSpace(s.role) == "" {
		return ParentAccountRole
	}
	return strings.TrimSpace(s.role)
}

func (s *ParentAccountGenerator) generatePassword() (string, error) {
	if s.passwordGenerator != nil {
		return s.passwordGenerator()
	}
	return randomTemporaryAccountPassword()
}

func markParentAccountGenerationFailed(result *ParentAccountGenerationResult, item *ParentAccountGenerationCandidate, reason string) {
	item.Status = "failed"
	item.Reason = reason
	result.Failed++
	result.Ready--
}

func auditParentAccountGeneration(ctx context.Context, store parentAccountGenerationStore, actorID pgtype.UUID, userID pgtype.UUID, parentID pgtype.UUID, username string, role string) error {
	payload, err := json.Marshal(map[string]any{
		"user_id":   uuidEntityID(userID),
		"parent_id": uuidEntityID(parentID),
		"username":  username,
		"role":      role,
	})
	if err != nil {
		return err
	}
	_, err = store.CreateAuditLog(ctx, db.CreateAuditLogParams{UserID: actorID, Action: "PARENT_ACCOUNT_GENERATED", EntityType: "user", EntityID: uuidEntityID(userID), Metadata: payload})
	return err
}

var _ parentAccountGenerationStore = (*db.Queries)(nil)
