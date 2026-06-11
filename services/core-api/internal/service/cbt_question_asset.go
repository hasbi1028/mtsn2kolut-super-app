package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type cbtQuestionAssetStore interface {
	GetCbtQuestionAsset(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error)
	GetExamQuestions(ctx context.Context, packageID pgtype.UUID) ([]db.GetExamQuestionsRow, error)
}

type CbtQuestionAsset struct {
	q        cbtQuestionAssetStore
	assetDir string
}

func NewCbtQuestionAsset(q *db.Queries, assetDir string) *CbtQuestionAsset {
	if strings.TrimSpace(assetDir) == "" {
		assetDir = "data/cbt-assets"
	}
	return &CbtQuestionAsset{q: q, assetDir: assetDir}
}

func (s *CbtQuestionAsset) Get(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error) {
	return s.q.GetCbtQuestionAsset(ctx, id)
}

func (s *CbtQuestionAsset) Open(asset db.CbtQuestionAsset) (*os.File, error) {
	assetDir, err := filepath.Abs(s.assetDir)
	if err != nil {
		return nil, err
	}
	storagePath, err := filepath.Abs(asset.StoragePath)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(assetDir, storagePath)
	if err != nil {
		return nil, err
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." || filepath.IsAbs(rel) {
		return nil, fmt.Errorf("storage_path di luar direktori aset")
	}
	return os.Open(storagePath)
}

func (s *CbtQuestionAsset) AccessibleByPackage(ctx context.Context, questionID, packageID pgtype.UUID) (bool, error) {
	if !questionID.Valid || !packageID.Valid {
		return false, nil
	}
	questions, err := s.q.GetExamQuestions(ctx, packageID)
	if err != nil {
		return false, err
	}
	for _, question := range questions {
		if question.ID == questionID {
			return true, nil
		}
	}
	return false, nil
}
