package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var nonSafeFilename = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type cbtQuestionAssetStore interface {
	CreateCbtQuestionAsset(ctx context.Context, arg db.CreateCbtQuestionAssetParams) (db.CbtQuestionAsset, error)
	GetCbtQuestionAsset(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error)
	ListCbtQuestionAssetsByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error)
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

type UploadCbtQuestionAssetInput struct {
	QuestionID   pgtype.UUID
	OriginalName string
	MimeType     string
	FileSize     int64
	Purpose      string
	UploadedBy   string
	File         io.Reader
}

func (s *CbtQuestionAsset) Save(ctx context.Context, input UploadCbtQuestionAssetInput) (db.CbtQuestionAsset, error) {
	if err := validateAssetInput(input); err != nil {
		return db.CbtQuestionAsset{}, err
	}
	if err := os.MkdirAll(s.assetDir, 0o755); err != nil {
		return db.CbtQuestionAsset{}, err
	}

	token, err := randomHex(16)
	if err != nil {
		return db.CbtQuestionAsset{}, err
	}
	safeName := sanitizeFilename(input.OriginalName)
	storedName := token + "_" + safeName
	absPath, err := filepath.Abs(filepath.Join(s.assetDir, storedName))
	if err != nil {
		return db.CbtQuestionAsset{}, err
	}
	out, err := os.Create(absPath)
	if err != nil {
		return db.CbtQuestionAsset{}, err
	}
	defer out.Close()
	if _, err := io.Copy(out, input.File); err != nil {
		return db.CbtQuestionAsset{}, err
	}

	return s.q.CreateCbtQuestionAsset(ctx, db.CreateCbtQuestionAssetParams{
		QuestionID:   input.QuestionID,
		OriginalName: input.OriginalName,
		StoredName:   storedName,
		MimeType:     input.MimeType,
		FileSize:     input.FileSize,
		StoragePath:  absPath,
		Purpose:      normalizeAssetPurpose(input.Purpose),
		UploadedBy:   input.UploadedBy,
	})
}

func (s *CbtQuestionAsset) Get(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error) {
	return s.q.GetCbtQuestionAsset(ctx, id)
}

func (s *CbtQuestionAsset) ListByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error) {
	return s.q.ListCbtQuestionAssetsByQuestion(ctx, questionID)
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

func validateAssetInput(input UploadCbtQuestionAssetInput) error {
	if input.File == nil {
		return fmt.Errorf("file wajib diisi")
	}
	if input.FileSize <= 0 {
		return fmt.Errorf("file_size tidak valid")
	}
	if input.FileSize > 10*1024*1024 {
		return fmt.Errorf("ukuran file maksimal 10MB")
	}
	if input.OriginalName == "" {
		return fmt.Errorf("original_name wajib diisi")
	}
	switch {
	case strings.HasPrefix(input.MimeType, "image/"),
		strings.HasPrefix(input.MimeType, "audio/"),
		input.MimeType == "application/pdf":
		return nil
	default:
		return fmt.Errorf("mime_type tidak didukung")
	}
}

func normalizeAssetPurpose(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "stimulus", "option", "explanation", "rubric", "supporting":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "general"
	}
}

func sanitizeFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "asset.bin"
	}
	value = nonSafeFilename.ReplaceAllString(value, "_")
	value = strings.Trim(value, "._")
	if value == "" {
		return "asset.bin"
	}
	return value
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
