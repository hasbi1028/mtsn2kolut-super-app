package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeCbtQuestionAssetStore struct {
	createArg   db.CreateCbtQuestionAssetParams
	createErr   error
	getID       pgtype.UUID
	listID      pgtype.UUID
	questionID  pgtype.UUID
	questionRow db.GetCbtQuestionRow
	questions   []db.GetExamQuestionsRow
	examErr     error
}

func (f *fakeCbtQuestionAssetStore) CreateCbtQuestionAsset(ctx context.Context, arg db.CreateCbtQuestionAssetParams) (db.CbtQuestionAsset, error) {
	f.createArg = arg
	if f.createErr != nil {
		return db.CbtQuestionAsset{}, f.createErr
	}
	return db.CbtQuestionAsset{
		QuestionID:   arg.QuestionID,
		OriginalName: arg.OriginalName,
		StoredName:   arg.StoredName,
		MimeType:     arg.MimeType,
		FileSize:     arg.FileSize,
		StoragePath:  arg.StoragePath,
		Purpose:      arg.Purpose,
		UploadedBy:   arg.UploadedBy,
	}, nil
}

func (f *fakeCbtQuestionAssetStore) GetCbtQuestionAsset(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error) {
	f.getID = id
	return db.CbtQuestionAsset{ID: id, OriginalName: "asset.png"}, nil
}

func (f *fakeCbtQuestionAssetStore) ListCbtQuestionAssetsByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error) {
	f.listID = questionID
	return []db.CbtQuestionAsset{{QuestionID: questionID, OriginalName: "asset.png"}}, nil
}

func (f *fakeCbtQuestionAssetStore) GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	f.questionID = id
	return f.questionRow, nil
}

func (f *fakeCbtQuestionAssetStore) GetExamQuestions(ctx context.Context, packageID pgtype.UUID) ([]db.GetExamQuestionsRow, error) {
	return f.questions, f.examErr
}

func TestCbtQuestionAssetSaveValidatesWritesAndCreatesRecord(t *testing.T) {
	questionID := mustUUID(t, "10101010-1010-1010-1010-101010101010")
	store := &fakeCbtQuestionAssetStore{}
	dir := t.TempDir()
	svc := NewCbtQuestionAsset(nil, "")
	if svc.assetDir != "data/cbt-assets" {
		t.Fatalf("NewCbtQuestionAsset(default) assetDir = %q, want data/cbt-assets", svc.assetDir)
	}
	svc = NewCbtQuestionAsset(nil, dir)
	svc.q = store

	asset, err := svc.Save(context.Background(), UploadCbtQuestionAssetInput{
		QuestionID:   questionID,
		OriginalName: " soal 1?.PNG ",
		MimeType:     "image/png",
		FileSize:     7,
		Purpose:      " Stimulus ",
		UploadedBy:   "guru",
		File:         strings.NewReader("content"),
	})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if asset.QuestionID != questionID || asset.Purpose != "stimulus" || asset.UploadedBy != "guru" {
		t.Fatalf("Save() asset = %+v, want normalized stored asset", asset)
	}
	if store.createArg.QuestionID != questionID || store.createArg.Purpose != "stimulus" || store.createArg.MimeType != "image/png" {
		t.Fatalf("Save() create arg = %+v, want normalized create params", store.createArg)
	}
	if !strings.HasSuffix(store.createArg.StoredName, "_soal_1_.PNG") || strings.Contains(store.createArg.StoredName, "?") {
		t.Fatalf("Save() stored name = %q, want token plus sanitized name", store.createArg.StoredName)
	}
	data, err := os.ReadFile(store.createArg.StoragePath)
	if err != nil {
		t.Fatalf("ReadFile(saved asset) error = %v", err)
	}
	if string(data) != "content" {
		t.Fatalf("saved asset content = %q, want content", string(data))
	}
}

func TestCbtQuestionAssetSaveRejectsInvalidInputAndCreateErrors(t *testing.T) {
	questionID := mustUUID(t, "11111111-1111-1111-1111-111111111112")
	valid := UploadCbtQuestionAssetInput{
		QuestionID:   questionID,
		OriginalName: "asset.pdf",
		MimeType:     "application/pdf",
		FileSize:     4,
		Purpose:      "rubric",
		UploadedBy:   "guru",
		File:         strings.NewReader("data"),
	}
	tests := []struct {
		name    string
		input   UploadCbtQuestionAssetInput
		wantErr string
	}{
		{name: "missing file", input: func() UploadCbtQuestionAssetInput { in := valid; in.File = nil; return in }(), wantErr: "file wajib diisi"},
		{name: "zero size", input: func() UploadCbtQuestionAssetInput { in := valid; in.FileSize = 0; return in }(), wantErr: "file_size tidak valid"},
		{name: "too large", input: func() UploadCbtQuestionAssetInput { in := valid; in.FileSize = 10*1024*1024 + 1; return in }(), wantErr: "ukuran file maksimal 10MB"},
		{name: "missing original name", input: func() UploadCbtQuestionAssetInput { in := valid; in.OriginalName = ""; return in }(), wantErr: "original_name wajib diisi"},
		{name: "unsupported mime", input: func() UploadCbtQuestionAssetInput { in := valid; in.MimeType = "text/plain"; return in }(), wantErr: "mime_type tidak didukung"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &CbtQuestionAsset{q: &fakeCbtQuestionAssetStore{}, assetDir: t.TempDir()}
			if _, err := svc.Save(context.Background(), tt.input); err == nil || err.Error() != tt.wantErr {
				t.Fatalf("Save() error = %v, want %q", err, tt.wantErr)
			}
		})
	}

	svc := &CbtQuestionAsset{q: &fakeCbtQuestionAssetStore{createErr: errors.New("create failed")}, assetDir: t.TempDir()}
	if _, err := svc.Save(context.Background(), valid); err == nil || err.Error() != "create failed" {
		t.Fatalf("Save(create error) = %v, want create failed", err)
	}
}

func TestCbtQuestionAssetOpenRejectsStoragePathOutsideAssetDir(t *testing.T) {
	dir := t.TempDir()
	svc := NewCbtQuestionAsset(nil, dir)
	outside, err := os.CreateTemp(t.TempDir(), "outside-*.png")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	outside.Close()

	_, err = svc.Open(db.CbtQuestionAsset{StoragePath: outside.Name()})
	if err == nil || !strings.Contains(err.Error(), "di luar direktori aset") {
		t.Fatalf("Open(outside storage_path) error = %v, want path defense error", err)
	}
}

func TestCbtQuestionAssetGetListAndHelpers(t *testing.T) {
	questionID := mustUUID(t, "12121212-1212-1212-1212-121212121212")
	assetID := mustUUID(t, "13131313-1313-1313-1313-131313131313")
	store := &fakeCbtQuestionAssetStore{}
	svc := &CbtQuestionAsset{q: store}

	if got, err := svc.Get(context.Background(), assetID); err != nil || got.ID != assetID || store.getID != assetID {
		t.Fatalf("Get() = %+v/%v id=%v, want asset", got, err, store.getID)
	}
	if rows, err := svc.ListByQuestion(context.Background(), questionID); err != nil || len(rows) != 1 || store.listID != questionID {
		t.Fatalf("ListByQuestion() = %d rows/%v id=%v, want asset list", len(rows), err, store.listID)
	}
	if got := normalizeAssetPurpose(" explanation "); got != "explanation" {
		t.Fatalf("normalizeAssetPurpose() = %q, want explanation", got)
	}
	if got := normalizeAssetPurpose("unknown"); got != "general" {
		t.Fatalf("normalizeAssetPurpose(unknown) = %q, want general", got)
	}
	if got := sanitizeFilename(" .. "); got != "asset.bin" {
		t.Fatalf("sanitizeFilename(blank) = %q, want asset.bin", got)
	}
	if token, err := randomHex(4); err != nil || len(token) != 8 {
		t.Fatalf("randomHex(4) = %q/%v, want 8 hex chars", token, err)
	}
}

func TestCbtQuestionAssetAccessibleByPackageReturnsTrueWhenQuestionIncluded(t *testing.T) {
	questionID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	packageID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	svc := &CbtQuestionAsset{
		q: &fakeCbtQuestionAssetStore{
			questions: []db.GetExamQuestionsRow{
				{ID: questionID},
			},
		},
	}

	allowed, err := svc.AccessibleByPackage(context.Background(), questionID, packageID)
	if err != nil {
		t.Fatalf("AccessibleByPackage() error = %v", err)
	}
	if !allowed {
		t.Fatal("AccessibleByPackage() = false, want true")
	}
}

func TestCbtQuestionAssetAccessibleByPackageReturnsFalseWhenQuestionExcluded(t *testing.T) {
	questionID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	packageID := mustUUID(t, "44444444-4444-4444-4444-444444444444")
	svc := &CbtQuestionAsset{
		q: &fakeCbtQuestionAssetStore{
			questions: []db.GetExamQuestionsRow{
				{ID: mustUUID(t, "55555555-5555-5555-5555-555555555555")},
			},
		},
	}

	allowed, err := svc.AccessibleByPackage(context.Background(), questionID, packageID)
	if err != nil {
		t.Fatalf("AccessibleByPackage() error = %v", err)
	}
	if allowed {
		t.Fatal("AccessibleByPackage() = true, want false")
	}
}

func TestCbtQuestionAssetAccessibleByPackageReturnsFalseForInvalidIDs(t *testing.T) {
	svc := &CbtQuestionAsset{q: &fakeCbtQuestionAssetStore{}}

	allowed, err := svc.AccessibleByPackage(context.Background(), pgtype.UUID{}, pgtype.UUID{})
	if err != nil {
		t.Fatalf("AccessibleByPackage() error = %v", err)
	}
	if allowed {
		t.Fatal("AccessibleByPackage() = true, want false")
	}
}
