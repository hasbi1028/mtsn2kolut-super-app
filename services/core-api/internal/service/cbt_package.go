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

type cbtPackageDetailStore interface {
	GetCbtPackageDetail(ctx context.Context, id pgtype.UUID) (db.GetCbtPackageDetailRow, error)
	ListCbtPackageQuestionsByPackage(ctx context.Context, packageID pgtype.UUID) ([]db.ListCbtPackageQuestionsByPackageRow, error)
}

type cbtPackageReadinessStore interface {
	ListCbtPackageReadiness(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackageReadinessRow, error)
}

type cbtPackageEditStore interface {
	LockCbtPackageForEdit(ctx context.Context, id pgtype.UUID) (db.LockCbtPackageForEditRow, error)
	UpdateCbtPackageMetadata(ctx context.Context, arg db.UpdateCbtPackageMetadataParams) (db.CbtPackage, error)
	GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error)
	DeleteCbtPackageQuestions(ctx context.Context, packageID pgtype.UUID) (int64, error)
	AddCbtPackageQuestion(ctx context.Context, arg db.AddCbtPackageQuestionParams) error
	CloneCbtPackage(ctx context.Context, arg db.CloneCbtPackageParams) (db.CbtPackage, error)
	CloneCbtPackageQuestions(ctx context.Context, arg db.CloneCbtPackageQuestionsParams) (int64, error)
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

type UpdateCbtPackageInput struct {
	ID                 pgtype.UUID
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
}

type ReplaceCbtPackageQuestionsInput struct {
	PackageID       pgtype.UUID
	QuestionIDs     []pgtype.UUID
	QuestionWeights map[string]int32
}

type CloneCbtPackageInput struct {
	SourceID pgtype.UUID
	Title    string
}

type CbtPackageDetailResult struct {
	Package   db.GetCbtPackageDetailRow                `json:"package"`
	Questions []db.ListCbtPackageQuestionsByPackageRow `json:"questions"`
	Readiness CbtPackageReadinessStatus                `json:"readiness"`
}

type CbtPackageReadinessStatus struct {
	Status            string `json:"status"`
	TargetPgCount     int32  `json:"target_pg_count"`
	TargetEssayCount  int32  `json:"target_essay_count"`
	MissingPgCount    int32  `json:"missing_pg_count"`
	MissingEssayCount int32  `json:"missing_essay_count"`
	QuestionCount     int32  `json:"question_count"`
	PgCount           int32  `json:"pg_count"`
	EssayCount        int32  `json:"essay_count"`
	TotalPoints       int32  `json:"total_points"`
	PublishedCount    int32  `json:"published_count"`
	UnpublishedCount  int32  `json:"unpublished_count"`
	MetadataGapCount  int32  `json:"metadata_gap_count"`
	SessionCount      int32  `json:"session_count"`
	Locked            bool   `json:"locked"`
	Ready             bool   `json:"ready"`
}

type CbtPackageReadinessItem struct {
	Package   db.ListCbtPackageReadinessRow `json:"package"`
	Readiness CbtPackageReadinessStatus     `json:"readiness"`
}

type CbtPackageReadinessOverview struct {
	Items []CbtPackageReadinessItem `json:"items"`
}

func (s *CbtPackage) Create(ctx context.Context, input CreateCbtPackageInput) (db.CbtPackage, error) {
	if err := validateCbtPackageTitle(input.Title); err != nil {
		return db.CbtPackage{}, err
	}
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
	if err := validateCbtPackageTitle(input.Title); err != nil {
		return db.CbtPackage{}, err
	}
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
		Title:              strings.TrimSpace(input.Title),
		Description:        strings.TrimSpace(input.Description),
		DurationMinutes:    input.DurationMinutes,
		RandomizeQuestions: input.RandomizeQuestions,
		IsActive:           input.IsActive,
		SourceMode:         input.SourceMode,
		RandomizeOptions:   input.RandomizeOptions,
		DrawPgCount:        input.DrawPgCount,
		DrawEssayCount:     input.DrawEssayCount,
		RandomSeed:         strings.TrimSpace(input.RandomSeed),
		CompositionLog:     compositionLog,
	})
	if err != nil {
		return db.CbtPackage{}, err
	}

	selections, err := validateCbtPackageQuestionSelection(ctx, q, input.EventID, input.SubjectID, input.QuestionIDs, input.QuestionWeights)
	if err != nil {
		return db.CbtPackage{}, err
	}
	for i, selection := range selections {
		if err := q.AddCbtPackageQuestion(ctx, db.AddCbtPackageQuestionParams{
			PackageID:  pkg.ID,
			QuestionID: selection.QuestionID,
			Position:   int32(i + 1),
			Points:     selection.Points,
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

type cbtPackageQuestionSelection struct {
	QuestionID pgtype.UUID
	Points     int32
}

func getCbtPackageDetail(ctx context.Context, q cbtPackageDetailStore, id pgtype.UUID) (CbtPackageDetailResult, error) {
	pkg, err := q.GetCbtPackageDetail(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CbtPackageDetailResult{}, domain.ErrNotFound
		}
		return CbtPackageDetailResult{}, err
	}
	questions, err := q.ListCbtPackageQuestionsByPackage(ctx, id)
	if err != nil {
		return CbtPackageDetailResult{}, err
	}
	readiness := cbtPackageReadinessStatusFromQuestions(questions, pkg.SessionCount, pkg.LockedAt.Valid)
	return CbtPackageDetailResult{Package: pkg, Questions: questions, Readiness: readiness}, nil
}

func updateCbtPackageMetadata(ctx context.Context, q cbtPackageEditStore, input UpdateCbtPackageInput) error {
	current, err := q.LockCbtPackageForEdit(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if current.LockedAt.Valid {
		return fmt.Errorf("%w: paket soal sudah terkunci; buat clone/revisi untuk mengubah metadata", domain.ErrConflict)
	}
	_, err = q.UpdateCbtPackageMetadata(ctx, db.UpdateCbtPackageMetadataParams{
		ID:                 input.ID,
		Title:              strings.TrimSpace(input.Title),
		Description:        strings.TrimSpace(input.Description),
		DurationMinutes:    input.DurationMinutes,
		RandomizeQuestions: input.RandomizeQuestions,
		RandomizeOptions:   input.RandomizeOptions,
		SourceMode:         input.SourceMode,
		DrawPgCount:        input.DrawPgCount,
		DrawEssayCount:     input.DrawEssayCount,
		RandomSeed:         strings.TrimSpace(input.RandomSeed),
		IsActive:           input.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: paket soal sudah terkunci atau tidak ditemukan", domain.ErrConflict)
		}
		return err
	}
	return nil
}

func replaceCbtPackageQuestions(ctx context.Context, q cbtPackageEditStore, input ReplaceCbtPackageQuestionsInput) error {
	current, err := q.LockCbtPackageForEdit(ctx, input.PackageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if current.LockedAt.Valid {
		return fmt.Errorf("%w: paket soal sudah terkunci; buat clone/revisi untuk mengubah soal", domain.ErrConflict)
	}
	selections, err := validateCbtPackageQuestionSelection(ctx, q, current.EventID, current.SubjectID, input.QuestionIDs, input.QuestionWeights)
	if err != nil {
		return err
	}
	if _, err := q.DeleteCbtPackageQuestions(ctx, input.PackageID); err != nil {
		return err
	}
	for i, selection := range selections {
		if err := q.AddCbtPackageQuestion(ctx, db.AddCbtPackageQuestionParams{
			PackageID:  input.PackageID,
			QuestionID: selection.QuestionID,
			Position:   int32(i + 1),
			Points:     selection.Points,
		}); err != nil {
			return err
		}
	}
	return nil
}

func cloneCbtPackage(ctx context.Context, q cbtPackageEditStore, input CloneCbtPackageInput) (db.CbtPackage, error) {
	pkg, err := q.CloneCbtPackage(ctx, db.CloneCbtPackageParams{
		SourceID: input.SourceID,
		Title:    strings.TrimSpace(input.Title),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.CbtPackage{}, domain.ErrNotFound
		}
		return db.CbtPackage{}, err
	}
	if _, err := q.CloneCbtPackageQuestions(ctx, db.CloneCbtPackageQuestionsParams{
		SourceID: input.SourceID,
		TargetID: pkg.ID,
	}); err != nil {
		return db.CbtPackage{}, err
	}
	return pkg, nil
}

func validateCbtPackageQuestionSelection(ctx context.Context, q interface {
	GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error)
}, eventID, subjectID pgtype.UUID, questionIDs []pgtype.UUID, weights map[string]int32) ([]cbtPackageQuestionSelection, error) {
	selections := make([]cbtPackageQuestionSelection, 0, len(questionIDs))
	seen := map[string]struct{}{}
	for _, questionID := range questionIDs {
		key := pgUUIDString(questionID)
		if key == "" {
			return nil, fmt.Errorf("%w: soal tidak valid", domain.ErrBadRequest)
		}
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("%w: soal paket tidak boleh duplikat", domain.ErrBadRequest)
		}
		seen[key] = struct{}{}

		question, err := q.GetCbtQuestion(ctx, questionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrNotFound
			}
			return nil, err
		}
		if question.SubjectID != subjectID {
			return nil, fmt.Errorf("semua soal harus dari mapel yang sama")
		}
		if question.Status != db.CbtQuestionStatusEnumPublished {
			return nil, fmt.Errorf("semua soal paket harus berstatus terbit")
		}
		switch {
		case eventID.Valid && question.EventID.Valid && !sameUUID(question.EventID, eventID):
			return nil, fmt.Errorf("%w: soal paket event harus berasal dari event yang sama", domain.ErrBadRequest)
		case !eventID.Valid && question.EventID.Valid:
			return nil, fmt.Errorf("%w: paket umum tidak boleh memakai soal khusus event", domain.ErrBadRequest)
		}
		points := int32(1)
		if weights != nil {
			if value, ok := weights[key]; ok {
				points = value
			}
		}
		if points < 1 || points > 100 {
			return nil, fmt.Errorf("bobot soal harus 1-100")
		}
		selections = append(selections, cbtPackageQuestionSelection{QuestionID: questionID, Points: points})
	}
	return selections, nil
}

func cbtPackageReadinessStatusFromQuestions(questions []db.ListCbtPackageQuestionsByPackageRow, sessionCount int32, locked bool) CbtPackageReadinessStatus {
	var pgCount, essayCount, totalPoints, publishedCount, unpublishedCount, metadataGapCount int32
	for _, question := range questions {
		switch question.QuestionType {
		case "multiple_choice":
			pgCount++
		case "essay":
			essayCount++
		}
		totalPoints += question.Points
		if question.Status == db.CbtQuestionStatusEnumPublished {
			publishedCount++
		} else {
			unpublishedCount++
		}
		if strings.TrimSpace(question.CpRef) == "" || (strings.TrimSpace(question.TpRef) == "" && strings.TrimSpace(question.KdRef) == "") || strings.TrimSpace(question.CognitiveLevel) == "" {
			metadataGapCount++
		}
	}
	return cbtPackageReadinessStatus(int32(len(questions)), pgCount, essayCount, totalPoints, publishedCount, unpublishedCount, metadataGapCount, sessionCount, locked)
}

func cbtPackageReadinessStatus(questionCount, pgCount, essayCount, totalPoints, publishedCount, unpublishedCount, metadataGapCount, sessionCount int32, locked bool) CbtPackageReadinessStatus {
	const targetPgCount int32 = 20
	const targetEssayCount int32 = 5
	missingPg := targetPgCount - pgCount
	if missingPg < 0 {
		missingPg = 0
	}
	missingEssay := targetEssayCount - essayCount
	if missingEssay < 0 {
		missingEssay = 0
	}
	ready := questionCount > 0 && missingPg == 0 && missingEssay == 0 && unpublishedCount == 0 && metadataGapCount == 0
	status := "kurang"
	switch {
	case locked:
		status = "locked"
	case questionCount == 0:
		status = "kosong"
	case ready:
		status = "siap"
	}
	return CbtPackageReadinessStatus{
		Status:            status,
		TargetPgCount:     targetPgCount,
		TargetEssayCount:  targetEssayCount,
		MissingPgCount:    missingPg,
		MissingEssayCount: missingEssay,
		QuestionCount:     questionCount,
		PgCount:           pgCount,
		EssayCount:        essayCount,
		TotalPoints:       totalPoints,
		PublishedCount:    publishedCount,
		UnpublishedCount:  unpublishedCount,
		MetadataGapCount:  metadataGapCount,
		SessionCount:      sessionCount,
		Locked:            locked,
		Ready:             ready,
	}
}

func (s *CbtPackage) Detail(ctx context.Context, id pgtype.UUID) (CbtPackageDetailResult, error) {
	store, ok := s.q.(cbtPackageDetailStore)
	if !ok {
		return CbtPackageDetailResult{}, fmt.Errorf("cbt package detail store unavailable")
	}
	return getCbtPackageDetail(ctx, store, id)
}

func (s *CbtPackage) Readiness(ctx context.Context, eventID pgtype.UUID) (CbtPackageReadinessOverview, error) {
	store, ok := s.q.(cbtPackageReadinessStore)
	if !ok {
		return CbtPackageReadinessOverview{}, fmt.Errorf("cbt package readiness store unavailable")
	}
	rows, err := store.ListCbtPackageReadiness(ctx, eventID)
	if err != nil {
		return CbtPackageReadinessOverview{}, err
	}
	items := make([]CbtPackageReadinessItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, CbtPackageReadinessItem{
			Package: row,
			Readiness: cbtPackageReadinessStatus(
				row.QuestionCount,
				row.PgCount,
				row.EssayCount,
				row.TotalPoints,
				row.PublishedCount,
				row.UnpublishedCount,
				row.MetadataGapCount,
				row.SessionCount,
				row.LockedAt.Valid,
			),
		})
	}
	return CbtPackageReadinessOverview{Items: items}, nil
}

func (s *CbtPackage) UpdateMetadata(ctx context.Context, input UpdateCbtPackageInput) (CbtPackageDetailResult, error) {
	if err := validateCbtPackageMetadataInput(input.Title, input.DurationMinutes, input.SourceMode, input.DrawPgCount, input.DrawEssayCount); err != nil {
		return CbtPackageDetailResult{}, err
	}
	if input.SourceMode == "" {
		input.SourceMode = "teacher_class"
	}

	if s.pool == nil {
		store, ok := s.q.(interface {
			cbtPackageEditStore
			cbtPackageDetailStore
		})
		if !ok {
			return CbtPackageDetailResult{}, fmt.Errorf("cbt package edit store unavailable")
		}
		if err := updateCbtPackageMetadata(ctx, store, input); err != nil {
			return CbtPackageDetailResult{}, err
		}
		return getCbtPackageDetail(ctx, store, input.ID)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CbtPackageDetailResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.q.WithTx(tx)
	if err := updateCbtPackageMetadata(ctx, qtx, input); err != nil {
		return CbtPackageDetailResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CbtPackageDetailResult{}, err
	}
	return s.Detail(ctx, input.ID)
}

func (s *CbtPackage) ReplaceQuestions(ctx context.Context, input ReplaceCbtPackageQuestionsInput) (CbtPackageDetailResult, error) {
	if s.pool == nil {
		store, ok := s.q.(interface {
			cbtPackageEditStore
			cbtPackageDetailStore
		})
		if !ok {
			return CbtPackageDetailResult{}, fmt.Errorf("cbt package edit store unavailable")
		}
		if err := replaceCbtPackageQuestions(ctx, store, input); err != nil {
			return CbtPackageDetailResult{}, err
		}
		return getCbtPackageDetail(ctx, store, input.PackageID)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CbtPackageDetailResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.q.WithTx(tx)
	if err := replaceCbtPackageQuestions(ctx, qtx, input); err != nil {
		return CbtPackageDetailResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CbtPackageDetailResult{}, err
	}
	return s.Detail(ctx, input.PackageID)
}

func (s *CbtPackage) Clone(ctx context.Context, input CloneCbtPackageInput) (CbtPackageDetailResult, error) {
	if strings.TrimSpace(input.Title) != "" {
		if err := validateCbtPackageTitle(input.Title); err != nil {
			return CbtPackageDetailResult{}, err
		}
	}

	if s.pool == nil {
		store, ok := s.q.(interface {
			cbtPackageEditStore
			cbtPackageDetailStore
		})
		if !ok {
			return CbtPackageDetailResult{}, fmt.Errorf("cbt package clone store unavailable")
		}
		pkg, err := cloneCbtPackage(ctx, store, input)
		if err != nil {
			return CbtPackageDetailResult{}, err
		}
		return getCbtPackageDetail(ctx, store, pkg.ID)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CbtPackageDetailResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.q.WithTx(tx)
	pkg, err := cloneCbtPackage(ctx, qtx, input)
	if err != nil {
		return CbtPackageDetailResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CbtPackageDetailResult{}, err
	}
	return s.Detail(ctx, pkg.ID)
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

func validateCbtPackageTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: nama paket wajib diisi", domain.ErrBadRequest)
	}
	return nil
}

func validateCbtPackageMetadataInput(title string, duration int32, sourceMode string, drawPgCount, drawEssayCount int32) error {
	if err := validateCbtPackageTitle(title); err != nil {
		return err
	}
	if err := validateCbtPackageDuration(duration); err != nil {
		return err
	}
	if sourceMode == "" {
		sourceMode = "teacher_class"
	}
	if !validCbtPackageSourceMode(sourceMode) {
		return fmt.Errorf("%w: mode sumber paket tidak valid", domain.ErrBadRequest)
	}
	if drawPgCount < 0 || drawEssayCount < 0 {
		return fmt.Errorf("%w: jumlah draw soal tidak boleh negatif", domain.ErrBadRequest)
	}
	return nil
}

func normalizeCbtPackageDuration(duration int32) int32 {
	if duration == 0 {
		return 60
	}
	return duration
}
