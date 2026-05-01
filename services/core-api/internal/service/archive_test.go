package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeArchiveStore struct {
	stats             db.GetArchiveStatsRow
	statsErr          error
	listCategoriesArg db.ListArchiveCategoriesParams
	listCategories    []db.ListArchiveCategoriesRow
	categoryID        pgtype.UUID
	category          db.ArchiveCategory
	categoryErr       error
	createCategoryArg db.CreateArchiveCategoryParams
	createCategoryErr error
	updateCategoryArg db.UpdateArchiveCategoryParams
	updateCategoryErr error
	deleteCategoryID  pgtype.UUID
	deleteCategoryErr error
	listDocumentsArg  db.ListArchiveDocumentsParams
	listDocuments     []db.ListArchiveDocumentsRow
	documentID        pgtype.UUID
	document          db.ArchiveDocument
	documentErr       error
	detailID          pgtype.UUID
	detail            db.GetArchiveDocumentDetailRow
	detailErr         error
	createDocumentArg db.CreateArchiveDocumentParams
	createDocumentErr error
	updateDocumentArg db.UpdateArchiveDocumentParams
	updateDocumentErr error
	deleteDocumentID  pgtype.UUID
	deleteDocumentErr error
}

func (f *fakeArchiveStore) GetArchiveStats(ctx context.Context) (db.GetArchiveStatsRow, error) {
	return f.stats, f.statsErr
}

func (f *fakeArchiveStore) ListArchiveCategories(ctx context.Context, arg db.ListArchiveCategoriesParams) ([]db.ListArchiveCategoriesRow, error) {
	f.listCategoriesArg = arg
	return f.listCategories, nil
}

func (f *fakeArchiveStore) GetArchiveCategory(ctx context.Context, id pgtype.UUID) (db.ArchiveCategory, error) {
	f.categoryID = id
	return f.category, f.categoryErr
}

func (f *fakeArchiveStore) CreateArchiveCategory(ctx context.Context, arg db.CreateArchiveCategoryParams) (db.ArchiveCategory, error) {
	f.createCategoryArg = arg
	if f.createCategoryErr != nil {
		return db.ArchiveCategory{}, f.createCategoryErr
	}
	return db.ArchiveCategory{
		ID:             archiveTestUUID(1),
		Code:           arg.Code,
		Name:           arg.Name,
		Description:    arg.Description,
		RetentionYears: arg.RetentionYears,
		IsActive:       arg.IsActive,
	}, nil
}

func (f *fakeArchiveStore) UpdateArchiveCategory(ctx context.Context, arg db.UpdateArchiveCategoryParams) (db.ArchiveCategory, error) {
	f.updateCategoryArg = arg
	if f.updateCategoryErr != nil {
		return db.ArchiveCategory{}, f.updateCategoryErr
	}
	return db.ArchiveCategory{
		ID:             arg.ID,
		Code:           arg.Code,
		Name:           arg.Name,
		Description:    arg.Description,
		RetentionYears: arg.RetentionYears,
		IsActive:       arg.IsActive,
	}, nil
}

func (f *fakeArchiveStore) DeleteArchiveCategory(ctx context.Context, id pgtype.UUID) error {
	f.deleteCategoryID = id
	return f.deleteCategoryErr
}

func (f *fakeArchiveStore) ListArchiveDocuments(ctx context.Context, arg db.ListArchiveDocumentsParams) ([]db.ListArchiveDocumentsRow, error) {
	f.listDocumentsArg = arg
	return f.listDocuments, nil
}

func (f *fakeArchiveStore) GetArchiveDocument(ctx context.Context, id pgtype.UUID) (db.ArchiveDocument, error) {
	f.documentID = id
	return f.document, f.documentErr
}

func (f *fakeArchiveStore) GetArchiveDocumentDetail(ctx context.Context, id pgtype.UUID) (db.GetArchiveDocumentDetailRow, error) {
	f.detailID = id
	return f.detail, f.detailErr
}

func (f *fakeArchiveStore) CreateArchiveDocument(ctx context.Context, arg db.CreateArchiveDocumentParams) (db.ArchiveDocument, error) {
	f.createDocumentArg = arg
	if f.createDocumentErr != nil {
		return db.ArchiveDocument{}, f.createDocumentErr
	}
	return db.ArchiveDocument{
		ID:               archiveTestUUID(2),
		CategoryID:       arg.CategoryID,
		Title:            arg.Title,
		ArchiveNumber:    arg.ArchiveNumber,
		DocumentDate:     arg.DocumentDate,
		ReceivedDate:     arg.ReceivedDate,
		Summary:          arg.Summary,
		Tags:             arg.Tags,
		Status:           arg.Status,
		StorageLocation:  arg.StorageLocation,
		RetentionUntil:   arg.RetentionUntil,
		OriginalName:     arg.OriginalName,
		StoredName:       arg.StoredName,
		FilePath:         arg.FilePath,
		MimeType:         arg.MimeType,
		FileSize:         arg.FileSize,
		ChecksumSha256:   arg.ChecksumSha256,
		UploadedByUserID: arg.UploadedByUserID,
	}, nil
}

func (f *fakeArchiveStore) UpdateArchiveDocument(ctx context.Context, arg db.UpdateArchiveDocumentParams) (db.ArchiveDocument, error) {
	f.updateDocumentArg = arg
	if f.updateDocumentErr != nil {
		return db.ArchiveDocument{}, f.updateDocumentErr
	}
	return db.ArchiveDocument{
		ID:              arg.ID,
		CategoryID:      arg.CategoryID,
		Title:           arg.Title,
		ArchiveNumber:   arg.ArchiveNumber,
		DocumentDate:    arg.DocumentDate,
		ReceivedDate:    arg.ReceivedDate,
		Tags:            arg.Tags,
		Status:          arg.Status,
		StorageLocation: arg.StorageLocation,
		RetentionUntil:  arg.RetentionUntil,
	}, nil
}

func (f *fakeArchiveStore) DeleteArchiveDocument(ctx context.Context, id pgtype.UUID) error {
	f.deleteDocumentID = id
	return f.deleteDocumentErr
}

func archiveTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func archiveTestDate(year int, month time.Month, day int) pgtype.Date {
	return pgtype.Date{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func TestArchiveCategoryMethodsNormalizeAndValidate(t *testing.T) {
	store := &fakeArchiveStore{}
	svc := &Archive{q: store}

	if _, err := svc.ListCategories(context.Background(), "  EDM  ", true); err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if store.listCategoriesArg.Search != "EDM" || !store.listCategoriesArg.ActiveOnly {
		t.Fatalf("ListCategories() arg = %+v, want trimmed search and active only", store.listCategoriesArg)
	}

	_, err := svc.CreateCategory(context.Background(), db.CreateArchiveCategoryParams{
		Code:               " adm ",
		Name:               "  Administrasi  ",
		ClassificationCode: "  400  ",
		Description:        "  Surat TU  ",
		RetentionYears:     5,
		IsActive:           true,
	})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}
	if store.createCategoryArg.Code != "ADM" || store.createCategoryArg.Name != "Administrasi" {
		t.Fatalf("CreateCategory() arg = %+v, want normalized code/name", store.createCategoryArg)
	}
	if store.createCategoryArg.ClassificationCode != "400" || store.createCategoryArg.Description != "Surat TU" {
		t.Fatalf("CreateCategory() arg = %+v, want trimmed classification/description", store.createCategoryArg)
	}

	if _, err := svc.CreateCategory(context.Background(), db.CreateArchiveCategoryParams{Name: "Arsip", RetentionYears: -1}); err == nil || err.Error() != "kode kategori wajib diisi" {
		t.Fatalf("CreateCategory(invalid) error = %v, want required code", err)
	}

	categoryID := archiveTestUUID(3)
	_, err = svc.UpdateCategory(context.Background(), db.UpdateArchiveCategoryParams{
		ID:             categoryID,
		Code:           " ed ",
		Name:           "  EDM  ",
		RetentionYears: 3,
	})
	if err != nil {
		t.Fatalf("UpdateCategory() error = %v", err)
	}
	if store.updateCategoryArg.ID != categoryID || store.updateCategoryArg.Code != "ED" || store.updateCategoryArg.Name != "EDM" {
		t.Fatalf("UpdateCategory() arg = %+v, want id and normalized text", store.updateCategoryArg)
	}

	if err := svc.DeleteCategory(context.Background(), pgtype.UUID{}); err == nil || err.Error() != "id kategori tidak valid" {
		t.Fatalf("DeleteCategory(invalid) error = %v, want invalid id", err)
	}
	if err := svc.DeleteCategory(context.Background(), categoryID); err != nil {
		t.Fatalf("DeleteCategory() error = %v", err)
	}
	if store.deleteCategoryID != categoryID {
		t.Fatalf("DeleteCategory() id = %v, want %v", store.deleteCategoryID, categoryID)
	}
}

func TestArchiveDocumentListAndGetValidateIDs(t *testing.T) {
	categoryID := archiveTestUUID(4)
	documentID := archiveTestUUID(5)
	store := &fakeArchiveStore{}
	svc := &Archive{q: store}

	if _, err := svc.ListDocuments(context.Background(), "  rapat  ", categoryID, "  BORROWED  ", "  400  "); err != nil {
		t.Fatalf("ListDocuments() error = %v", err)
	}
	if store.listDocumentsArg.Search != "rapat" || store.listDocumentsArg.Status != "borrowed" || store.listDocumentsArg.ClassificationCode != "400" {
		t.Fatalf("ListDocuments() arg = %+v, want normalized filters", store.listDocumentsArg)
	}
	if _, err := svc.ListDocuments(context.Background(), "", categoryID, "tidak-valid", ""); err != nil {
		t.Fatalf("ListDocuments(invalid status filter) error = %v", err)
	}
	if store.listDocumentsArg.Status != "" {
		t.Fatalf("ListDocuments() invalid status = %q, want empty filter", store.listDocumentsArg.Status)
	}

	if _, err := svc.GetDocument(context.Background(), pgtype.UUID{}); err == nil || err.Error() != "id arsip tidak valid" {
		t.Fatalf("GetDocument(invalid) error = %v, want invalid id", err)
	}
	if _, err := svc.GetDocument(context.Background(), documentID); err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if store.detailID != documentID {
		t.Fatalf("GetArchiveDocumentDetail() id = %v, want %v", store.detailID, documentID)
	}

	if _, err := svc.GetDocumentFile(context.Background(), pgtype.UUID{}); err == nil || err.Error() != "id arsip tidak valid" {
		t.Fatalf("GetDocumentFile(invalid) error = %v, want invalid id", err)
	}
	if _, err := svc.GetDocumentFile(context.Background(), documentID); err != nil {
		t.Fatalf("GetDocumentFile() error = %v", err)
	}
	if store.documentID != documentID {
		t.Fatalf("GetArchiveDocument() id = %v, want %v", store.documentID, documentID)
	}
}

func TestArchiveSaveDocumentStoresFileAndMetadata(t *testing.T) {
	categoryID := archiveTestUUID(6)
	receivedDate := archiveTestDate(2026, time.January, 2)
	store := &fakeArchiveStore{
		category: db.ArchiveCategory{
			ID:             categoryID,
			RetentionYears: 2,
			IsActive:       true,
		},
	}
	storageDir := t.TempDir()
	svc := &Archive{q: store, storageDir: storageDir}

	row, err := svc.SaveDocument(context.Background(), UploadArchiveDocumentInput{
		CategoryID:       categoryID,
		Title:            "  Notulen Rapat EDM  ",
		ArchiveNumber:    "  ARS/001  ",
		DocumentDate:     archiveTestDate(2026, time.January, 1),
		ReceivedDate:     receivedDate,
		Summary:          "  Ringkasan  ",
		Tags:             " EDM, rapat, edm,  ",
		Status:           "",
		StorageLocation:  "  Lemari A  ",
		OriginalName:     "  Surat Final.pdf  ",
		FileSize:         int64(len("arsip\n")),
		UploadedByUserID: archiveTestUUID(7),
		File:             strings.NewReader("arsip\n"),
	})
	if err != nil {
		t.Fatalf("SaveDocument() error = %v", err)
	}
	if row.Title != "Notulen Rapat EDM" || row.ArchiveNumber != "ARS/001" || row.Tags != "EDM, rapat" {
		t.Fatalf("SaveDocument() row = %+v, want normalized metadata", row)
	}
	arg := store.createDocumentArg
	if arg.CategoryID != categoryID || arg.Status != "active" || arg.StorageLocation != "Lemari A" {
		t.Fatalf("CreateArchiveDocument() arg = %+v, want normalized category/status/location", arg)
	}
	if arg.MimeType != "application/pdf" {
		t.Fatalf("CreateArchiveDocument() mime = %q, want application/pdf", arg.MimeType)
	}
	if !arg.RetentionUntil.Valid || arg.RetentionUntil.Time.Format("2006-01-02") != "2028-01-02" {
		t.Fatalf("CreateArchiveDocument() retention = %v, want 2028-01-02", arg.RetentionUntil.Time)
	}
	if arg.FileSize != int64(len("arsip\n")) || len(arg.ChecksumSha256) != 64 {
		t.Fatalf("CreateArchiveDocument() file size/checksum = %d/%q, want actual size and sha256", arg.FileSize, arg.ChecksumSha256)
	}
	if !strings.HasPrefix(arg.FilePath, storageDir) {
		t.Fatalf("CreateArchiveDocument() filepath = %q, want under %q", arg.FilePath, storageDir)
	}
	if !strings.HasSuffix(arg.StoredName, "Surat_Final.pdf") {
		t.Fatalf("CreateArchiveDocument() stored name = %q, want sanitized original suffix", arg.StoredName)
	}
	if _, err := os.Stat(arg.FilePath); err != nil {
		t.Fatalf("stored archive file stat error = %v", err)
	}
}

func TestArchiveSaveDocumentRejectsInactiveCategoryAndInvalidInput(t *testing.T) {
	categoryID := archiveTestUUID(8)
	store := &fakeArchiveStore{
		category: db.ArchiveCategory{ID: categoryID, IsActive: false},
	}
	svc := &Archive{q: store, storageDir: t.TempDir()}

	_, err := svc.SaveDocument(context.Background(), UploadArchiveDocumentInput{
		CategoryID:   categoryID,
		Title:        "Dokumen",
		ReceivedDate: archiveTestDate(2026, time.January, 2),
		OriginalName: "dokumen.pdf",
		FileSize:     4,
		File:         strings.NewReader("data"),
	})
	if err == nil || err.Error() != "kategori arsip tidak aktif" {
		t.Fatalf("SaveDocument(inactive category) error = %v, want inactive category", err)
	}

	store.category.IsActive = true
	_, err = svc.SaveDocument(context.Background(), UploadArchiveDocumentInput{
		CategoryID:   categoryID,
		Title:        "Dokumen",
		ReceivedDate: archiveTestDate(2026, time.January, 2),
		OriginalName: "dokumen.exe",
		FileSize:     4,
		File:         strings.NewReader("data"),
	})
	if err == nil || err.Error() != "jenis file arsip tidak didukung" {
		t.Fatalf("SaveDocument(invalid mime) error = %v, want unsupported mime", err)
	}
}

func TestArchiveUpdateDocumentNormalizesAndValidates(t *testing.T) {
	categoryID := archiveTestUUID(9)
	documentID := archiveTestUUID(10)
	store := &fakeArchiveStore{
		category: db.ArchiveCategory{ID: categoryID, IsActive: true},
	}
	svc := &Archive{q: store}

	_, err := svc.UpdateDocument(context.Background(), db.UpdateArchiveDocumentParams{
		ID:              documentID,
		CategoryID:      categoryID,
		Title:           "  SK Tim  ",
		ArchiveNumber:   "  SK/001  ",
		ReceivedDate:    archiveTestDate(2026, time.February, 3),
		Summary:         "  Ringkas  ",
		Tags:            " SK, tim, sk ",
		Status:          " BORROWED ",
		StorageLocation: "  Lemari B  ",
	})
	if err != nil {
		t.Fatalf("UpdateDocument() error = %v", err)
	}
	arg := store.updateDocumentArg
	if arg.ID != documentID || arg.Title != "SK Tim" || arg.ArchiveNumber != "SK/001" {
		t.Fatalf("UpdateDocument() arg = %+v, want normalized title/archive number", arg)
	}
	if arg.Tags != "SK, tim" || arg.Status != "borrowed" || arg.StorageLocation != "Lemari B" {
		t.Fatalf("UpdateDocument() arg = %+v, want normalized tags/status/location", arg)
	}
	if store.categoryID != categoryID {
		t.Fatalf("GetArchiveCategory() id = %v, want %v", store.categoryID, categoryID)
	}

	_, err = svc.UpdateDocument(context.Background(), db.UpdateArchiveDocumentParams{
		ID:           documentID,
		CategoryID:   categoryID,
		Title:        "",
		ReceivedDate: archiveTestDate(2026, time.February, 3),
		Status:       "active",
	})
	if err == nil || err.Error() != "judul arsip wajib diisi" {
		t.Fatalf("UpdateDocument(invalid) error = %v, want title validation", err)
	}
}

func TestArchiveDeleteDocumentRemovesStoredFile(t *testing.T) {
	documentID := archiveTestUUID(11)
	dir := t.TempDir()
	path := filepath.Join(dir, "arsip.pdf")
	if err := os.WriteFile(path, []byte("arsip"), 0o600); err != nil {
		t.Fatalf("setup archive file: %v", err)
	}
	store := &fakeArchiveStore{
		document: db.ArchiveDocument{ID: documentID, FilePath: path},
	}
	svc := &Archive{q: store}

	if err := svc.DeleteDocument(context.Background(), documentID); err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
	if store.deleteDocumentID != documentID {
		t.Fatalf("DeleteArchiveDocument() id = %v, want %v", store.deleteDocumentID, documentID)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("archive file stat error = %v, want file removed", err)
	}

	if err := svc.DeleteDocument(context.Background(), pgtype.UUID{}); err == nil || err.Error() != "id arsip tidak valid" {
		t.Fatalf("DeleteDocument(invalid) error = %v, want invalid id", err)
	}
}

func TestArchiveStatsAndDocumentErrorBranches(t *testing.T) {
	categoryID := archiveTestUUID(12)
	documentID := archiveTestUUID(13)
	receivedDate := archiveTestDate(2026, time.March, 4)
	expectedErr := errors.New("store failed")

	store := &fakeArchiveStore{stats: db.GetArchiveStatsRow{TotalDocuments: 7}}
	svc := &Archive{q: store, storageDir: t.TempDir()}
	stats, err := svc.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.TotalDocuments != 7 {
		t.Fatalf("Stats().TotalDocuments = %d, want 7", stats.TotalDocuments)
	}
	store.statsErr = expectedErr
	if _, err := svc.Stats(context.Background()); !errors.Is(err, expectedErr) {
		t.Fatalf("Stats(error) = %v, want %v", err, expectedErr)
	}

	input := UploadArchiveDocumentInput{
		CategoryID:       categoryID,
		Title:            "Dokumen",
		ReceivedDate:     receivedDate,
		OriginalName:     "dokumen.pdf",
		FileSize:         int64(len("arsip")),
		UploadedByUserID: archiveTestUUID(14),
		File:             strings.NewReader("arsip"),
	}

	store = &fakeArchiveStore{categoryErr: expectedErr}
	svc = &Archive{q: store, storageDir: t.TempDir()}
	if _, err := svc.SaveDocument(context.Background(), input); !errors.Is(err, expectedErr) {
		t.Fatalf("SaveDocument(category error) = %v, want %v", err, expectedErr)
	}

	store = &fakeArchiveStore{category: db.ArchiveCategory{ID: categoryID, IsActive: true}}
	svc = &Archive{q: store, storageDir: t.TempDir()}
	emptyInput := input
	emptyInput.File = strings.NewReader("")
	if _, err := svc.SaveDocument(context.Background(), emptyInput); err == nil || err.Error() != "file arsip kosong" {
		t.Fatalf("SaveDocument(empty reader) = %v, want empty file error", err)
	}

	store = &fakeArchiveStore{
		category:          db.ArchiveCategory{ID: categoryID, IsActive: true},
		createDocumentErr: expectedErr,
	}
	svc = &Archive{q: store, storageDir: t.TempDir()}
	if _, err := svc.SaveDocument(context.Background(), input); !errors.Is(err, expectedErr) {
		t.Fatalf("SaveDocument(create error) = %v, want %v", err, expectedErr)
	}
	if store.createDocumentArg.FilePath == "" {
		t.Fatal("SaveDocument(create error) did not build file path")
	}
	if _, err := os.Stat(store.createDocumentArg.FilePath); !os.IsNotExist(err) {
		t.Fatalf("stored file after create error stat = %v, want removed", err)
	}

	store = &fakeArchiveStore{categoryErr: expectedErr}
	svc = &Archive{q: store}
	if _, err := svc.UpdateDocument(context.Background(), db.UpdateArchiveDocumentParams{
		ID:           documentID,
		CategoryID:   categoryID,
		Title:        "Dokumen",
		ReceivedDate: receivedDate,
		Status:       "active",
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateDocument(category error) = %v, want %v", err, expectedErr)
	}

	store = &fakeArchiveStore{
		category:          db.ArchiveCategory{ID: categoryID, IsActive: true},
		updateDocumentErr: expectedErr,
	}
	svc = &Archive{q: store}
	if _, err := svc.UpdateDocument(context.Background(), db.UpdateArchiveDocumentParams{
		ID:           documentID,
		CategoryID:   categoryID,
		Title:        "Dokumen",
		ReceivedDate: receivedDate,
		Status:       "active",
	}); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateDocument(update error) = %v, want %v", err, expectedErr)
	}

	store = &fakeArchiveStore{documentErr: expectedErr}
	svc = &Archive{q: store}
	if err := svc.DeleteDocument(context.Background(), documentID); !errors.Is(err, expectedErr) {
		t.Fatalf("DeleteDocument(get error) = %v, want %v", err, expectedErr)
	}

	store = &fakeArchiveStore{document: db.ArchiveDocument{ID: documentID}, deleteDocumentErr: expectedErr}
	svc = &Archive{q: store}
	if err := svc.DeleteDocument(context.Background(), documentID); !errors.Is(err, expectedErr) {
		t.Fatalf("DeleteDocument(delete error) = %v, want %v", err, expectedErr)
	}
}

func TestArchiveValidationHelperBranches(t *testing.T) {
	if err := validateArchiveCategory("ADM", "", 1); err == nil || err.Error() != "nama kategori wajib diisi" {
		t.Fatalf("validateArchiveCategory(empty name) = %v, want name error", err)
	}
	if err := validateArchiveCategory("ADM", "Administrasi", -1); err == nil || err.Error() != "masa retensi tidak boleh negatif" {
		t.Fatalf("validateArchiveCategory(negative retention) = %v, want retention error", err)
	}

	validCategoryID := archiveTestUUID(15)
	validDate := archiveTestDate(2026, time.April, 5)
	tests := []struct {
		name    string
		input   UploadArchiveDocumentInput
		wantErr string
	}{
		{name: "missing file", input: UploadArchiveDocumentInput{CategoryID: validCategoryID, Title: "Dokumen", ReceivedDate: validDate, OriginalName: "a.pdf", FileSize: 1}, wantErr: "file arsip wajib diisi"},
		{name: "bad file size", input: UploadArchiveDocumentInput{CategoryID: validCategoryID, Title: "Dokumen", ReceivedDate: validDate, OriginalName: "a.pdf", File: strings.NewReader("a")}, wantErr: "ukuran file arsip tidak valid"},
		{name: "too large", input: UploadArchiveDocumentInput{CategoryID: validCategoryID, Title: "Dokumen", ReceivedDate: validDate, OriginalName: "a.pdf", FileSize: maxArchiveFileSize + 1, File: strings.NewReader("a")}, wantErr: "ukuran file arsip maksimal 25MB"},
		{name: "missing original name", input: UploadArchiveDocumentInput{CategoryID: validCategoryID, Title: "Dokumen", ReceivedDate: validDate, FileSize: 1, File: strings.NewReader("a")}, wantErr: "nama file arsip wajib diisi"},
		{name: "missing category", input: UploadArchiveDocumentInput{Title: "Dokumen", ReceivedDate: validDate, OriginalName: "a.pdf", MimeType: "application/pdf", FileSize: 1, File: strings.NewReader("a")}, wantErr: "kategori arsip wajib dipilih"},
		{name: "bad status", input: UploadArchiveDocumentInput{CategoryID: validCategoryID, Title: "Dokumen", ReceivedDate: validDate, OriginalName: "a.pdf", MimeType: "application/pdf", Status: "lost", FileSize: 1, File: strings.NewReader("a")}, wantErr: "status arsip tidak valid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArchiveDocumentInput(tt.input)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("validateArchiveDocumentInput() = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestArchiveParseAndNormalizeHelpers(t *testing.T) {
	id, err := ParseArchiveOptionalUUID("  00000000-0000-0000-0000-000000000001  ")
	if err != nil {
		t.Fatalf("ParseArchiveOptionalUUID() error = %v", err)
	}
	if !id.Valid {
		t.Fatalf("ParseArchiveOptionalUUID() valid = false, want true")
	}
	if emptyID, err := ParseArchiveOptionalUUID(" "); err != nil || emptyID.Valid {
		t.Fatalf("ParseArchiveOptionalUUID(empty) = %v, %v; want invalid nil", emptyID, err)
	}
	if _, err := ParseArchiveOptionalUUID("bad"); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ParseArchiveOptionalUUID(bad) error = %v, want invalid id", err)
	}

	date, err := ParseArchiveOptionalDate(" 2026-05-01 ")
	if err != nil {
		t.Fatalf("ParseArchiveOptionalDate() error = %v", err)
	}
	if !date.Valid || date.Time.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("ParseArchiveOptionalDate() = %v, want 2026-05-01", date.Time)
	}
	if _, err := ParseArchiveOptionalDate("not-a-date"); err == nil || err.Error() != "format tanggal tidak valid" {
		t.Fatalf("ParseArchiveOptionalDate(bad) error = %v, want invalid date", err)
	}

	if got := normalizeArchiveTags(" EDM, rapat, edm, , Komite "); got != "EDM, rapat, Komite" {
		t.Fatalf("normalizeArchiveTags() = %q, want deduped tags", got)
	}
	if got := normalizeArchiveStatus(" "); got != "active" {
		t.Fatalf("normalizeArchiveStatus(empty) = %q, want active", got)
	}
	if got := normalizeArchiveStatusFilter(" DISPOSED "); got != "disposed" {
		t.Fatalf("normalizeArchiveStatusFilter() = %q, want disposed", got)
	}
	if got := normalizeArchiveMimeType("", "surat.docx"); got != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		t.Fatalf("normalizeArchiveMimeType(docx) = %q, want docx mime", got)
	}
	if !archiveMimeAllowed("image/png") || archiveMimeAllowed("application/x-msdownload") {
		t.Fatalf("archiveMimeAllowed() image/x-msdownload result unexpected")
	}
}
