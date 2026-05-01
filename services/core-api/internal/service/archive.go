package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const maxArchiveFileSize = 25 * 1024 * 1024

type archiveStore interface {
	GetArchiveStats(ctx context.Context) (db.GetArchiveStatsRow, error)
	ListArchiveCategories(ctx context.Context, arg db.ListArchiveCategoriesParams) ([]db.ListArchiveCategoriesRow, error)
	GetArchiveCategory(ctx context.Context, id pgtype.UUID) (db.ArchiveCategory, error)
	CreateArchiveCategory(ctx context.Context, arg db.CreateArchiveCategoryParams) (db.ArchiveCategory, error)
	UpdateArchiveCategory(ctx context.Context, arg db.UpdateArchiveCategoryParams) (db.ArchiveCategory, error)
	DeleteArchiveCategory(ctx context.Context, id pgtype.UUID) error
	ListArchiveDocuments(ctx context.Context, arg db.ListArchiveDocumentsParams) ([]db.ListArchiveDocumentsRow, error)
	GetArchiveDocument(ctx context.Context, id pgtype.UUID) (db.ArchiveDocument, error)
	GetArchiveDocumentDetail(ctx context.Context, id pgtype.UUID) (db.GetArchiveDocumentDetailRow, error)
	CreateArchiveDocument(ctx context.Context, arg db.CreateArchiveDocumentParams) (db.ArchiveDocument, error)
	UpdateArchiveDocument(ctx context.Context, arg db.UpdateArchiveDocumentParams) (db.ArchiveDocument, error)
	DeleteArchiveDocument(ctx context.Context, id pgtype.UUID) error
}

type Archive struct {
	q          archiveStore
	storageDir string
}

func NewArchive(q *db.Queries, storageDir string) *Archive {
	if strings.TrimSpace(storageDir) == "" {
		storageDir = "data/archives"
	}
	return &Archive{q: q, storageDir: storageDir}
}

type UploadArchiveDocumentInput struct {
	CategoryID       pgtype.UUID
	Title            string
	ArchiveNumber    string
	DocumentDate     pgtype.Date
	ReceivedDate     pgtype.Date
	Summary          string
	Tags             string
	Status           string
	StorageLocation  string
	RetentionUntil   pgtype.Date
	OriginalName     string
	MimeType         string
	FileSize         int64
	UploadedByUserID pgtype.UUID
	File             io.Reader
}

func (s *Archive) Stats(ctx context.Context) (db.GetArchiveStatsRow, error) {
	return s.q.GetArchiveStats(ctx)
}

func (s *Archive) ListCategories(ctx context.Context, search string, activeOnly bool) ([]db.ListArchiveCategoriesRow, error) {
	return s.q.ListArchiveCategories(ctx, db.ListArchiveCategoriesParams{
		Search:     strings.TrimSpace(search),
		ActiveOnly: activeOnly,
	})
}

func (s *Archive) CreateCategory(ctx context.Context, arg db.CreateArchiveCategoryParams) (db.ArchiveCategory, error) {
	arg.Code = strings.ToUpper(strings.TrimSpace(arg.Code))
	arg.Name = strings.TrimSpace(arg.Name)
	arg.ClassificationCode = strings.TrimSpace(arg.ClassificationCode)
	arg.Description = strings.TrimSpace(arg.Description)
	if err := validateArchiveCategory(arg.Code, arg.Name, arg.RetentionYears); err != nil {
		return db.ArchiveCategory{}, err
	}
	return s.q.CreateArchiveCategory(ctx, arg)
}

func (s *Archive) UpdateCategory(ctx context.Context, arg db.UpdateArchiveCategoryParams) (db.ArchiveCategory, error) {
	arg.Code = strings.ToUpper(strings.TrimSpace(arg.Code))
	arg.Name = strings.TrimSpace(arg.Name)
	arg.ClassificationCode = strings.TrimSpace(arg.ClassificationCode)
	arg.Description = strings.TrimSpace(arg.Description)
	if err := validateArchiveCategory(arg.Code, arg.Name, arg.RetentionYears); err != nil {
		return db.ArchiveCategory{}, err
	}
	return s.q.UpdateArchiveCategory(ctx, arg)
}

func (s *Archive) DeleteCategory(ctx context.Context, id pgtype.UUID) error {
	if !id.Valid {
		return fmt.Errorf("id kategori tidak valid")
	}
	return s.q.DeleteArchiveCategory(ctx, id)
}

func (s *Archive) ListDocuments(ctx context.Context, search string, categoryID pgtype.UUID, status, classificationCode string) ([]db.ListArchiveDocumentsRow, error) {
	return s.q.ListArchiveDocuments(ctx, db.ListArchiveDocumentsParams{
		Search:             strings.TrimSpace(search),
		CategoryID:         categoryID,
		Status:             normalizeArchiveStatusFilter(status),
		ClassificationCode: strings.TrimSpace(classificationCode),
	})
}

func (s *Archive) GetDocument(ctx context.Context, id pgtype.UUID) (db.GetArchiveDocumentDetailRow, error) {
	if !id.Valid {
		return db.GetArchiveDocumentDetailRow{}, fmt.Errorf("id arsip tidak valid")
	}
	return s.q.GetArchiveDocumentDetail(ctx, id)
}

func (s *Archive) GetDocumentFile(ctx context.Context, id pgtype.UUID) (db.ArchiveDocument, error) {
	if !id.Valid {
		return db.ArchiveDocument{}, fmt.Errorf("id arsip tidak valid")
	}
	return s.q.GetArchiveDocument(ctx, id)
}

func (s *Archive) SaveDocument(ctx context.Context, input UploadArchiveDocumentInput) (db.ArchiveDocument, error) {
	input = normalizeArchiveDocumentInput(input)
	category, err := s.q.GetArchiveCategory(ctx, input.CategoryID)
	if err != nil {
		return db.ArchiveDocument{}, err
	}
	if !category.IsActive {
		return db.ArchiveDocument{}, fmt.Errorf("kategori arsip tidak aktif")
	}
	if err := validateArchiveDocumentInput(input); err != nil {
		return db.ArchiveDocument{}, err
	}
	if !input.RetentionUntil.Valid && category.RetentionYears > 0 {
		input.RetentionUntil = pgtype.Date{
			Time:  input.ReceivedDate.Time.AddDate(int(category.RetentionYears), 0, 0),
			Valid: true,
		}
	}
	if err := os.MkdirAll(s.storageDir, 0o755); err != nil {
		return db.ArchiveDocument{}, err
	}

	token, err := randomHex(16)
	if err != nil {
		return db.ArchiveDocument{}, err
	}
	storedName := token + "_" + sanitizeFilename(input.OriginalName)
	absPath, err := filepath.Abs(filepath.Join(s.storageDir, storedName))
	if err != nil {
		return db.ArchiveDocument{}, err
	}

	out, err := os.Create(absPath)
	if err != nil {
		return db.ArchiveDocument{}, err
	}
	hash := sha256.New()
	written, copyErr := io.Copy(out, io.TeeReader(input.File, hash))
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(absPath)
		return db.ArchiveDocument{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(absPath)
		return db.ArchiveDocument{}, closeErr
	}
	if written <= 0 {
		_ = os.Remove(absPath)
		return db.ArchiveDocument{}, fmt.Errorf("file arsip kosong")
	}

	row, err := s.q.CreateArchiveDocument(ctx, db.CreateArchiveDocumentParams{
		CategoryID:       input.CategoryID,
		Title:            input.Title,
		ArchiveNumber:    input.ArchiveNumber,
		DocumentDate:     input.DocumentDate,
		ReceivedDate:     input.ReceivedDate,
		Summary:          input.Summary,
		Tags:             input.Tags,
		Status:           input.Status,
		StorageLocation:  input.StorageLocation,
		RetentionUntil:   input.RetentionUntil,
		OriginalName:     input.OriginalName,
		StoredName:       storedName,
		FilePath:         absPath,
		MimeType:         input.MimeType,
		FileSize:         written,
		ChecksumSha256:   hex.EncodeToString(hash.Sum(nil)),
		UploadedByUserID: input.UploadedByUserID,
	})
	if err != nil {
		_ = os.Remove(absPath)
		return db.ArchiveDocument{}, err
	}
	return row, nil
}

func (s *Archive) UpdateDocument(ctx context.Context, arg db.UpdateArchiveDocumentParams) (db.ArchiveDocument, error) {
	arg.Title = strings.TrimSpace(arg.Title)
	arg.ArchiveNumber = strings.TrimSpace(arg.ArchiveNumber)
	arg.Summary = strings.TrimSpace(arg.Summary)
	arg.Tags = normalizeArchiveTags(arg.Tags)
	arg.Status = normalizeArchiveStatus(arg.Status)
	arg.StorageLocation = strings.TrimSpace(arg.StorageLocation)
	if !arg.ReceivedDate.Valid {
		arg.ReceivedDate = pgtype.Date{Time: time.Now(), Valid: true}
	}
	if _, err := s.q.GetArchiveCategory(ctx, arg.CategoryID); err != nil {
		return db.ArchiveDocument{}, err
	}
	if err := validateArchiveDocumentMetadata(arg.CategoryID, arg.Title, arg.ReceivedDate, arg.Status); err != nil {
		return db.ArchiveDocument{}, err
	}
	return s.q.UpdateArchiveDocument(ctx, arg)
}

func (s *Archive) DeleteDocument(ctx context.Context, id pgtype.UUID) error {
	if !id.Valid {
		return fmt.Errorf("id arsip tidak valid")
	}
	document, err := s.q.GetArchiveDocument(ctx, id)
	if err != nil {
		return err
	}
	if err := s.q.DeleteArchiveDocument(ctx, id); err != nil {
		return err
	}
	if strings.TrimSpace(document.FilePath) != "" {
		_ = os.Remove(document.FilePath)
	}
	return nil
}

func ParseArchiveOptionalUUID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if strings.TrimSpace(raw) == "" {
		return id, nil
	}
	if err := id.Scan(strings.TrimSpace(raw)); err != nil {
		return pgtype.UUID{}, fmt.Errorf("id tidak valid")
	}
	return id, nil
}

func ParseArchiveOptionalDate(raw string) (pgtype.Date, error) {
	var date pgtype.Date
	if strings.TrimSpace(raw) == "" {
		return date, nil
	}
	if err := date.Scan(strings.TrimSpace(raw)); err != nil {
		return pgtype.Date{}, fmt.Errorf("format tanggal tidak valid")
	}
	return date, nil
}

func validateArchiveCategory(code, name string, retentionYears int32) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("kode kategori wajib diisi")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nama kategori wajib diisi")
	}
	if retentionYears < 0 {
		return fmt.Errorf("masa retensi tidak boleh negatif")
	}
	return nil
}

func normalizeArchiveDocumentInput(input UploadArchiveDocumentInput) UploadArchiveDocumentInput {
	input.Title = strings.TrimSpace(input.Title)
	input.ArchiveNumber = strings.TrimSpace(input.ArchiveNumber)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Tags = normalizeArchiveTags(input.Tags)
	input.Status = normalizeArchiveStatus(input.Status)
	input.StorageLocation = strings.TrimSpace(input.StorageLocation)
	input.OriginalName = strings.TrimSpace(input.OriginalName)
	input.MimeType = normalizeArchiveMimeType(input.MimeType, input.OriginalName)
	if !input.ReceivedDate.Valid {
		input.ReceivedDate = pgtype.Date{Time: time.Now(), Valid: true}
	}
	return input
}

func validateArchiveDocumentInput(input UploadArchiveDocumentInput) error {
	if input.File == nil {
		return fmt.Errorf("file arsip wajib diisi")
	}
	if input.FileSize <= 0 {
		return fmt.Errorf("ukuran file arsip tidak valid")
	}
	if input.FileSize > maxArchiveFileSize {
		return fmt.Errorf("ukuran file arsip maksimal 25MB")
	}
	if strings.TrimSpace(input.OriginalName) == "" {
		return fmt.Errorf("nama file arsip wajib diisi")
	}
	if !archiveMimeAllowed(input.MimeType) {
		return fmt.Errorf("jenis file arsip tidak didukung")
	}
	return validateArchiveDocumentMetadata(input.CategoryID, input.Title, input.ReceivedDate, input.Status)
}

func validateArchiveDocumentMetadata(categoryID pgtype.UUID, title string, receivedDate pgtype.Date, status string) error {
	if !categoryID.Valid {
		return fmt.Errorf("kategori arsip wajib dipilih")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("judul arsip wajib diisi")
	}
	if !receivedDate.Valid {
		return fmt.Errorf("tanggal arsip diterima wajib diisi")
	}
	switch normalizeArchiveStatus(status) {
	case "active", "borrowed", "disposed":
		return nil
	default:
		return fmt.Errorf("status arsip tidak valid")
	}
}

func normalizeArchiveStatus(value string) string {
	status := strings.TrimSpace(strings.ToLower(value))
	if status == "" {
		return "active"
	}
	return status
}

func normalizeArchiveStatusFilter(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "active", "borrowed", "disposed":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return ""
	}
}

func normalizeArchiveTags(value string) string {
	parts := strings.Split(value, ",")
	tags := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if seen[key] {
			continue
		}
		seen[key] = true
		tags = append(tags, tag)
	}
	return strings.Join(tags, ", ")
}

func normalizeArchiveMimeType(value, filename string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value != "" {
		return value
	}
	if ext := strings.ToLower(filepath.Ext(filename)); ext != "" {
		if detected := mime.TypeByExtension(ext); detected != "" {
			return strings.Split(detected, ";")[0]
		}
	}
	return ""
}

func archiveMimeAllowed(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	if strings.HasPrefix(value, "image/") {
		return true
	}
	switch value {
	case "application/pdf",
		"text/plain",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/zip":
		return true
	default:
		return false
	}
}
