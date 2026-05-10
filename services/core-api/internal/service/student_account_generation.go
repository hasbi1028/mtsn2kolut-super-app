package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const StudentAccountRole = "siswa"

type studentAccountGenerationStore interface {
	ListStudentAccountGenerationCandidates(ctx context.Context) ([]db.ListStudentAccountGenerationCandidatesRow, error)
	CreateUserWithMustChangePassword(ctx context.Context, arg db.CreateUserWithMustChangePasswordParams) (db.CreateUserWithMustChangePasswordRow, error)
	AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error
	AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

type StudentAccountGenerator struct {
	q                 studentAccountGenerationStore
	tx                userLifecycleTxStarter
	role              string
	passwordGenerator func() (string, error)
}

type StudentAccountGenerationCandidate struct {
	StudentID         string `json:"student_id"`
	NIS               string `json:"nis"`
	NISN              string `json:"nisn"`
	Nama              string `json:"nama"`
	GeneratedUsername string `json:"generated_username,omitempty"`
	TemporaryPassword string `json:"temporary_password,omitempty"`
	Role              string `json:"role"`
	Status            string `json:"status"`
	Reason            string `json:"reason,omitempty"`
	UserID            string `json:"user_id,omitempty"`
	ExistingUserID    string `json:"existing_user_id,omitempty"`
}

type StudentAccountGenerationResult struct {
	Role       string                              `json:"role"`
	Total      int                                 `json:"total"`
	Ready      int                                 `json:"ready"`
	Created    int                                 `json:"created"`
	Skipped    int                                 `json:"skipped"`
	Failed     int                                 `json:"failed"`
	Candidates []StudentAccountGenerationCandidate `json:"candidates"`
}

func NewStudentAccountGenerator(q *db.Queries) *StudentAccountGenerator {
	return &StudentAccountGenerator{q: q, role: StudentAccountRole}
}

func NewStudentAccountGeneratorWithPool(pool *pgxpool.Pool) *StudentAccountGenerator {
	return &StudentAccountGenerator{q: db.New(pool), tx: pool, role: StudentAccountRole}
}

func (s *StudentAccountGenerator) Preview(ctx context.Context) (StudentAccountGenerationResult, error) {
	rows, err := s.q.ListStudentAccountGenerationCandidates(ctx)
	if err != nil {
		return StudentAccountGenerationResult{}, err
	}
	return s.buildResult(rows), nil
}

func (s *StudentAccountGenerator) Generate(ctx context.Context, actorID pgtype.UUID) (StudentAccountGenerationResult, error) {
	rows, err := s.q.ListStudentAccountGenerationCandidates(ctx)
	if err != nil {
		return StudentAccountGenerationResult{}, err
	}
	result := s.buildResult(rows)
	if result.Ready == 0 {
		return result, nil
	}

	mutate := func(store studentAccountGenerationStore) error {
		for i := range result.Candidates {
			item := &result.Candidates[i]
			if item.Status != "ready" {
				continue
			}
			studentID, err := uuidFromString(item.StudentID)
			if err != nil {
				markStudentAccountGenerationFailed(&result, item, "student_id tidak valid")
				continue
			}
			password, err := s.studentInitialPassword(item)
			if err != nil {
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
				StudentID:          studentID,
				IsActive:           true,
				MustChangePassword: true,
			})
			if err != nil {
				markStudentAccountGenerationFailed(&result, item, safeAccountGenerationError(err))
				continue
			}
			if err := addLegacyAccountRoleIfSafe(ctx, store, row.ID, s.accountRole()); err != nil {
				return err
			}
			if err := store.AddUserRbacRoleByCode(ctx, db.AddUserRbacRoleByCodeParams{UserID: row.ID, Code: s.accountRole()}); err != nil {
				return err
			}
			if err := auditStudentAccountGeneration(ctx, store, actorID, row.ID, studentID, item.GeneratedUsername, s.accountRole()); err != nil {
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

func (s *StudentAccountGenerator) buildResult(rows []db.ListStudentAccountGenerationCandidatesRow) StudentAccountGenerationResult {
	result := StudentAccountGenerationResult{Role: s.accountRole(), Candidates: make([]StudentAccountGenerationCandidate, 0, len(rows))}
	takenByBase := map[string]map[string]bool{}
	for _, row := range rows {
		base := strings.TrimSpace(row.BaseUsername)
		item := StudentAccountGenerationCandidate{
			StudentID: uuidEntityID(row.StudentID),
			NIS:       row.Nis,
			NISN:      row.Nisn,
			Nama:      row.Nama,
			Role:      s.accountRole(),
		}
		mergeTakenUsernames(takenByBase, base, row.UsernameCollisions)
		switch {
		case row.ExistingUserID.Valid:
			item.Status = "skipped"
			item.Reason = "siswa sudah punya akun"
			item.ExistingUserID = uuidEntityID(row.ExistingUserID)
			result.Skipped++
		case strings.TrimSpace(row.Nisn) == "":
			item.Status = "skipped"
			item.Reason = "NISN belum diisi"
			result.Skipped++
		case base == "":
			item.Status = "skipped"
			item.Reason = "NISN/NIS belum diisi"
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

func (s *StudentAccountGenerator) accountRole() string {
	if strings.TrimSpace(s.role) == "" {
		return StudentAccountRole
	}
	return strings.TrimSpace(s.role)
}

func (s *StudentAccountGenerator) studentInitialPassword(item *StudentAccountGenerationCandidate) (string, error) {
	if s.passwordGenerator != nil {
		return s.passwordGenerator()
	}
	password := strings.TrimSpace(item.NISN)
	if password == "" {
		return "", fmt.Errorf("NISN belum diisi")
	}
	return password, nil
}

func (s *StudentAccountGenerator) generatePassword() (string, error) {
	if s.passwordGenerator != nil {
		return s.passwordGenerator()
	}
	return randomTemporaryAccountPassword()
}

func markStudentAccountGenerationFailed(result *StudentAccountGenerationResult, item *StudentAccountGenerationCandidate, reason string) {
	item.Status = "failed"
	item.Reason = reason
	result.Failed++
	result.Ready--
}

func auditStudentAccountGeneration(ctx context.Context, store studentAccountGenerationStore, actorID pgtype.UUID, userID pgtype.UUID, studentID pgtype.UUID, username string, role string) error {
	payload, err := json.Marshal(map[string]any{
		"user_id":    uuidEntityID(userID),
		"student_id": uuidEntityID(studentID),
		"username":   username,
		"role":       role,
	})
	if err != nil {
		return err
	}
	_, err = store.CreateAuditLog(ctx, db.CreateAuditLogParams{UserID: actorID, Action: "STUDENT_ACCOUNT_GENERATED", EntityType: "user", EntityID: uuidEntityID(userID), Metadata: payload})
	return err
}

type accountGenerationRoleStore interface {
	AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error
}

func addLegacyAccountRoleIfSafe(ctx context.Context, store accountGenerationRoleStore, userID pgtype.UUID, role string) error {
	err := store.AddUserRole(ctx, db.AddUserRoleParams{UserID: userID, Role: db.UserRole(role)})
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "user_role") || strings.Contains(message, "invalid input value for enum") {
		return nil
	}
	return err
}

func mergeTakenUsernames(takenByBase map[string]map[string]bool, base string, usernames []string) {
	if base == "" {
		return
	}
	if _, ok := takenByBase[base]; !ok {
		takenByBase[base] = map[string]bool{}
	}
	for _, username := range usernames {
		username = strings.TrimSpace(username)
		if username != "" {
			takenByBase[base][username] = true
		}
	}
}

func nextAvailableAccountUsername(base string, taken map[string]bool) string {
	if taken == nil {
		taken = map[string]bool{}
	}
	if !taken[base] {
		taken[base] = true
		return base
	}
	for sequence := 1; ; sequence++ {
		candidate := fmt.Sprintf("%s-%02d", base, sequence)
		if !taken[candidate] {
			taken[candidate] = true
			return candidate
		}
	}
}

func randomTemporaryAccountPassword() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#%?"
	const length = 12
	out := make([]byte, length)
	for i := range out {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}

func safeAccountGenerationError(err error) string {
	if err == nil {
		return "gagal membuat akun"
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "gagal membuat akun"
	}
	return message
}

var _ studentAccountGenerationStore = (*db.Queries)(nil)
