package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtPackage struct {
	pool *pgxpool.Pool
	q    cbtPackageStore
}

type cbtPackageStore interface {
	ListCbtPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, error)
	ListCbtPackageQuestions(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackageQuestionsRow, error)
	GetCbtPackageUsage(ctx context.Context, id pgtype.UUID) (int32, error)
	DeleteCbtPackage(ctx context.Context, id pgtype.UUID) (int64, error)
	WithTx(tx pgx.Tx) *db.Queries
}

type cbtPackageSnapshotStore interface {
	LockCbtPackageForSnapshot(ctx context.Context, arg db.LockCbtPackageForSnapshotParams) (db.LockCbtPackageForSnapshotRow, error)
	CreateCbtPackageQuestionSnapshots(ctx context.Context, packageID pgtype.UUID) (int64, error)
}

type cbtPackageCreateStore interface {
	CreateCbtPackage(ctx context.Context, arg db.CreateCbtPackageParams) (db.CbtPackage, error)
	GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error)
	AddCbtPackageQuestion(ctx context.Context, arg db.AddCbtPackageQuestionParams) error
}

func NewCbtPackage(pool *pgxpool.Pool) *CbtPackage {
	return &CbtPackage{pool: pool, q: db.New(pool)}
}

func (s *CbtPackage) List(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, []db.ListCbtPackageQuestionsRow, error) {
	packages, err := s.q.ListCbtPackages(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}
	questions, err := s.q.ListCbtPackageQuestions(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}
	return packages, questions, nil
}

type CreateCbtPackageInput struct {
	EventID            pgtype.UUID
	SubjectID          pgtype.UUID
	Title              string
	Description        string
	DurationMinutes    int32
	RandomizeQuestions bool
	RandomizeOptions   bool
	SourceMode         string
	DrawPgCount        int32
	DrawEssayCount     int32
	RandomSeed         string
	IsActive           bool
	QuestionIDs        []pgtype.UUID
	QuestionWeights    map[string]int32
}

func (s *CbtPackage) Create(ctx context.Context, input CreateCbtPackageInput) (db.CbtPackage, error) {
	if err := validateCbtPackageDuration(input.DurationMinutes); err != nil {
		return db.CbtPackage{}, err
	}
	input.DurationMinutes = normalizeCbtPackageDuration(input.DurationMinutes)
	if len(input.QuestionIDs) == 0 {
		return db.CbtPackage{}, fmt.Errorf("question_ids wajib diisi")
	}
	if input.SourceMode == "" {
		input.SourceMode = "teacher_class"
	}
	if !validCbtPackageSourceMode(input.SourceMode) {
		return db.CbtPackage{}, fmt.Errorf("%w: mode sumber paket tidak valid", domain.ErrBadRequest)
	}
	if input.DrawPgCount < 0 || input.DrawEssayCount < 0 {
		return db.CbtPackage{}, fmt.Errorf("%w: jumlah draw soal tidak boleh negatif", domain.ErrBadRequest)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.CbtPackage{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)
	pkg, err := createCbtPackage(ctx, qtx, input)
	if err != nil {
		return db.CbtPackage{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return db.CbtPackage{}, err
	}

	return pkg, nil
}

func createCbtPackage(ctx context.Context, q cbtPackageCreateStore, input CreateCbtPackageInput) (db.CbtPackage, error) {
	if err := validateCbtPackageDuration(input.DurationMinutes); err != nil {
		return db.CbtPackage{}, err
	}
	input.DurationMinutes = normalizeCbtPackageDuration(input.DurationMinutes)
	if input.SourceMode == "" {
		input.SourceMode = "teacher_class"
	}
	if !validCbtPackageSourceMode(input.SourceMode) {
		return db.CbtPackage{}, fmt.Errorf("%w: mode sumber paket tidak valid", domain.ErrBadRequest)
	}
	if input.DrawPgCount < 0 || input.DrawEssayCount < 0 {
		return db.CbtPackage{}, fmt.Errorf("%w: jumlah draw soal tidak boleh negatif", domain.ErrBadRequest)
	}
	compositionLog, err := json.Marshal(map[string]any{
		"source_mode":             input.SourceMode,
		"randomize_questions":     input.RandomizeQuestions,
		"randomize_options":       input.RandomizeOptions,
		"draw_pg_count":           input.DrawPgCount,
		"draw_essay_count":        input.DrawEssayCount,
		"random_seed":             input.RandomSeed,
		"selected_question_count": len(input.QuestionIDs),
	})
	if err != nil {
		return db.CbtPackage{}, err
	}
	pkg, err := q.CreateCbtPackage(ctx, db.CreateCbtPackageParams{
		EventID:            input.EventID,
		SubjectID:          input.SubjectID,
		Title:              input.Title,
		Description:        input.Description,
		DurationMinutes:    input.DurationMinutes,
		RandomizeQuestions: input.RandomizeQuestions,
		IsActive:           input.IsActive,
		SourceMode:         input.SourceMode,
		RandomizeOptions:   input.RandomizeOptions,
		DrawPgCount:        input.DrawPgCount,
		DrawEssayCount:     input.DrawEssayCount,
		RandomSeed:         input.RandomSeed,
		CompositionLog:     compositionLog,
	})
	if err != nil {
		return db.CbtPackage{}, err
	}

	for i, questionID := range input.QuestionIDs {
		question, err := q.GetCbtQuestion(ctx, questionID)
		if err != nil {
			return db.CbtPackage{}, err
		}
		if question.SubjectID != input.SubjectID {
			return db.CbtPackage{}, fmt.Errorf("semua soal harus dari mapel yang sama")
		}
		if question.Status != db.CbtQuestionStatusEnumPublished {
			return db.CbtPackage{}, fmt.Errorf("semua soal paket harus berstatus terbit")
		}
		switch {
		case input.EventID.Valid && question.EventID.Valid && !sameUUID(question.EventID, input.EventID):
			return db.CbtPackage{}, fmt.Errorf("%w: soal paket event harus berasal dari event yang sama", domain.ErrBadRequest)
		case !input.EventID.Valid && question.EventID.Valid:
			return db.CbtPackage{}, fmt.Errorf("%w: paket umum tidak boleh memakai soal khusus event", domain.ErrBadRequest)
		}
		points := int32(1)
		if input.QuestionWeights != nil {
			if value, ok := input.QuestionWeights[pgUUIDString(questionID)]; ok {
				points = value
			}
		}
		if points < 1 || points > 100 {
			return db.CbtPackage{}, fmt.Errorf("bobot soal harus 1-100")
		}
		if err := q.AddCbtPackageQuestion(ctx, db.AddCbtPackageQuestionParams{
			PackageID:  pkg.ID,
			QuestionID: questionID,
			Position:   int32(i + 1),
			Points:     points,
		}); err != nil {
			return db.CbtPackage{}, err
		}
	}
	return pkg, nil
}

func validCbtPackageSourceMode(value string) bool {
	switch value {
	case "teacher_class", "level_subject_teachers", "event_pool":
		return true
	default:
		return false
	}
}

type CbtPackageSnapshotResult struct {
	PackageID         string `json:"package_id"`
	LockedAt          string `json:"locked_at"`
	LockReason        string `json:"lock_reason"`
	SnapshotVersion   int32  `json:"snapshot_version"`
	SnapshotRowsAdded int64  `json:"snapshot_rows_added"`
}

func (s *CbtPackage) LockAndSnapshot(ctx context.Context, packageID, lockedBy pgtype.UUID, reason string) (CbtPackageSnapshotResult, error) {
	if s.pool == nil {
		store, ok := s.q.(cbtPackageSnapshotStore)
		if !ok {
			return CbtPackageSnapshotResult{}, fmt.Errorf("cbt package snapshot store unavailable")
		}
		return lockCbtPackageSnapshot(ctx, store, packageID, lockedBy, reason)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CbtPackageSnapshotResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	result, err := lockCbtPackageSnapshot(ctx, s.q.WithTx(tx), packageID, lockedBy, reason)
	if err != nil {
		return CbtPackageSnapshotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CbtPackageSnapshotResult{}, err
	}
	return result, nil
}

func lockCbtPackageSnapshot(ctx context.Context, q cbtPackageSnapshotStore, packageID, lockedBy pgtype.UUID, reason string) (CbtPackageSnapshotResult, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "session_scheduled"
	}
	row, err := q.LockCbtPackageForSnapshot(ctx, db.LockCbtPackageForSnapshotParams{
		PackageID:  packageID,
		LockedBy:   lockedBy,
		LockReason: reason,
	})
	if err != nil {
		return CbtPackageSnapshotResult{}, err
	}
	inserted, err := q.CreateCbtPackageQuestionSnapshots(ctx, packageID)
	if err != nil {
		return CbtPackageSnapshotResult{}, err
	}
	lockedAt := ""
	if row.LockedAt.Valid {
		lockedAt = row.LockedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return CbtPackageSnapshotResult{
		PackageID:         pgUUIDString(row.ID),
		LockedAt:          lockedAt,
		LockReason:        row.LockReason,
		SnapshotVersion:   row.SnapshotVersion,
		SnapshotRowsAdded: inserted,
	}, nil
}

func (s *CbtPackage) Delete(ctx context.Context, id pgtype.UUID) error {
	sessionCount, err := s.q.GetCbtPackageUsage(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if sessionCount > 0 {
		return fmt.Errorf("%w: paket CBT sudah digunakan oleh sesi ujian dan tidak dapat dihapus", domain.ErrConflict)
	}
	rows, err := s.q.DeleteCbtPackage(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: paket CBT tidak ditemukan atau sudah terkunci", domain.ErrConflict)
	}
	return nil
}

func validateCbtPackageDuration(duration int32) error {
	if duration == 0 {
		return nil
	}
	if duration < 1 || duration > 360 {
		return fmt.Errorf("%w: durasi paket CBT harus 1-360 menit", domain.ErrBadRequest)
	}
	return nil
}

func normalizeCbtPackageDuration(duration int32) int32 {
	if duration == 0 {
		return 60
	}
	return duration
}
