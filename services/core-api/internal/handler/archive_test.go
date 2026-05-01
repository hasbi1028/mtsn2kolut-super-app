package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeArchiveService struct {
	statsRow db.GetArchiveStatsRow
	statsErr error

	listCategorySearch     string
	listCategoryActiveOnly bool
	listCategoryRows       []db.ListArchiveCategoriesRow
	listCategoryErr        error

	createCategoryArg db.CreateArchiveCategoryParams
	createCategoryRow db.ArchiveCategory
	createCategoryErr error

	updateCategoryArg db.UpdateArchiveCategoryParams
	updateCategoryRow db.ArchiveCategory
	updateCategoryErr error

	deleteCategoryID  pgtype.UUID
	deleteCategoryErr error

	listDocumentSearch             string
	listDocumentCategoryID         pgtype.UUID
	listDocumentStatus             string
	listDocumentClassificationCode string
	listDocumentRows               []db.ListArchiveDocumentsRow
	listDocumentErr                error

	uploadInput   service.UploadArchiveDocumentInput
	uploadContent string
	uploadRow     db.ArchiveDocument
	uploadErr     error

	getDocumentID  pgtype.UUID
	getDocumentRow db.GetArchiveDocumentDetailRow
	getDocumentErr error

	updateDocumentArg db.UpdateArchiveDocumentParams
	updateDocumentRow db.ArchiveDocument
	updateDocumentErr error

	deleteDocumentID  pgtype.UUID
	deleteDocumentErr error

	getFileID  pgtype.UUID
	getFileRow db.ArchiveDocument
	getFileErr error
}

func (f *fakeArchiveService) Stats(_ context.Context) (db.GetArchiveStatsRow, error) {
	return f.statsRow, f.statsErr
}

func (f *fakeArchiveService) ListCategories(_ context.Context, search string, activeOnly bool) ([]db.ListArchiveCategoriesRow, error) {
	f.listCategorySearch = search
	f.listCategoryActiveOnly = activeOnly
	if f.listCategoryErr != nil {
		return nil, f.listCategoryErr
	}
	return f.listCategoryRows, nil
}

func (f *fakeArchiveService) CreateCategory(_ context.Context, arg db.CreateArchiveCategoryParams) (db.ArchiveCategory, error) {
	f.createCategoryArg = arg
	if f.createCategoryErr != nil {
		return db.ArchiveCategory{}, f.createCategoryErr
	}
	return f.createCategoryRow, nil
}

func (f *fakeArchiveService) UpdateCategory(_ context.Context, arg db.UpdateArchiveCategoryParams) (db.ArchiveCategory, error) {
	f.updateCategoryArg = arg
	if f.updateCategoryErr != nil {
		return db.ArchiveCategory{}, f.updateCategoryErr
	}
	return f.updateCategoryRow, nil
}

func (f *fakeArchiveService) DeleteCategory(_ context.Context, id pgtype.UUID) error {
	f.deleteCategoryID = id
	return f.deleteCategoryErr
}

func (f *fakeArchiveService) ListDocuments(_ context.Context, search string, categoryID pgtype.UUID, status, classificationCode string) ([]db.ListArchiveDocumentsRow, error) {
	f.listDocumentSearch = search
	f.listDocumentCategoryID = categoryID
	f.listDocumentStatus = status
	f.listDocumentClassificationCode = classificationCode
	if f.listDocumentErr != nil {
		return nil, f.listDocumentErr
	}
	return f.listDocumentRows, nil
}

func (f *fakeArchiveService) SaveDocument(_ context.Context, input service.UploadArchiveDocumentInput) (db.ArchiveDocument, error) {
	f.uploadInput = input
	if input.File != nil {
		content, err := io.ReadAll(input.File)
		if err != nil {
			return db.ArchiveDocument{}, err
		}
		f.uploadContent = string(content)
	}
	if f.uploadErr != nil {
		return db.ArchiveDocument{}, f.uploadErr
	}
	return f.uploadRow, nil
}

func (f *fakeArchiveService) GetDocument(_ context.Context, id pgtype.UUID) (db.GetArchiveDocumentDetailRow, error) {
	f.getDocumentID = id
	if f.getDocumentErr != nil {
		return db.GetArchiveDocumentDetailRow{}, f.getDocumentErr
	}
	return f.getDocumentRow, nil
}

func (f *fakeArchiveService) UpdateDocument(_ context.Context, arg db.UpdateArchiveDocumentParams) (db.ArchiveDocument, error) {
	f.updateDocumentArg = arg
	if f.updateDocumentErr != nil {
		return db.ArchiveDocument{}, f.updateDocumentErr
	}
	return f.updateDocumentRow, nil
}

func (f *fakeArchiveService) DeleteDocument(_ context.Context, id pgtype.UUID) error {
	f.deleteDocumentID = id
	return f.deleteDocumentErr
}

func (f *fakeArchiveService) GetDocumentFile(_ context.Context, id pgtype.UUID) (db.ArchiveDocument, error) {
	f.getFileID = id
	if f.getFileErr != nil {
		return db.ArchiveDocument{}, f.getFileErr
	}
	return f.getFileRow, nil
}

func newArchiveMultipartRequest(t *testing.T, fields map[string]string, withFile bool) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", key, err)
		}
	}
	if withFile {
		part, err := writer.CreateFormFile("file", "arsip.pdf")
		if err != nil {
			t.Fatalf("CreateFormFile() error = %v", err)
		}
		if _, err := part.Write([]byte("archive")); err != nil {
			t.Fatalf("multipart file write error = %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/tu/archives/documents", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return withClaims(req, jwt.MapClaims{
		"roles": []any{"staf"},
		"uid":   "01000000-0000-0000-0000-000000000000",
	})
}

func TestArchiveCategoryHandlersForwardSuccessPaths(t *testing.T) {
	categoryID := handlerTestUUID(50)
	category := db.ArchiveCategory{ID: categoryID, Code: "AK", Name: "Akademik", RetentionYears: 5, IsActive: true}
	fake := &fakeArchiveService{
		statsRow:          db.GetArchiveStatsRow{ActiveCategories: 2, TotalDocuments: 4},
		listCategoryRows:  []db.ListArchiveCategoriesRow{{ID: categoryID, Code: "AK", Name: "Akademik", IsActive: true}},
		createCategoryRow: category,
		updateCategoryRow: category,
	}
	h := &Archive{svc: fake}

	rec := httptest.NewRecorder()
	h.Stats(rec, adminRequest(http.MethodGet, "/api/tu/archives/stats", ""))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "total_documents") {
		t.Fatalf("Stats() status/body = %d/%s, want stats payload", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListCategories(rec, adminRequest(http.MethodGet, "/api/tu/archives/categories?search=aka&active_only=true", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListCategories() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listCategorySearch != "aka" || !fake.listCategoryActiveOnly {
		t.Fatalf("ListCategories filters = %q/%v, want search active", fake.listCategorySearch, fake.listCategoryActiveOnly)
	}

	rec = httptest.NewRecorder()
	h.CreateCategory(rec, adminRequest(http.MethodPost, "/api/tu/archives/categories", `{"code":"ak","name":"Akademik","classification_code":"420","description":"Dokumen akademik","retention_years":5,"is_active":false}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateCategory() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createCategoryArg.Code != "ak" || fake.createCategoryArg.Name != "Akademik" || fake.createCategoryArg.IsActive {
		t.Fatalf("CreateCategory arg = %+v, want request fields and inactive flag", fake.createCategoryArg)
	}

	rec = httptest.NewRecorder()
	updateReq := withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/categories/"+categoryID.String(), `{"code":"ak","name":"Akademik Baru","retention_years":7}`), "id", categoryID.String())
	h.UpdateCategory(rec, updateReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateCategory() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateCategoryArg.ID != categoryID || fake.updateCategoryArg.Name != "Akademik Baru" || fake.updateCategoryArg.RetentionYears != 7 || !fake.updateCategoryArg.IsActive {
		t.Fatalf("UpdateCategory arg = %+v, want id/body/default active", fake.updateCategoryArg)
	}

	rec = httptest.NewRecorder()
	h.DeleteCategory(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/tu/archives/categories/"+categoryID.String(), ""), "id", categoryID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteCategory() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteCategoryID != categoryID {
		t.Fatalf("DeleteCategory id = %v, want %v", fake.deleteCategoryID, categoryID)
	}
}

func TestArchiveDocumentHandlersForwardSuccessPaths(t *testing.T) {
	categoryID := handlerTestUUID(60)
	documentID := handlerTestUUID(61)
	document := db.ArchiveDocument{ID: documentID, CategoryID: categoryID, Title: "SK Tim", OriginalName: "arsip.pdf", MimeType: "application/pdf", FileSize: 7}
	fake := &fakeArchiveService{
		listDocumentRows:  []db.ListArchiveDocumentsRow{{ID: documentID, CategoryID: categoryID, Title: "SK Tim", Status: "active"}},
		uploadRow:         document,
		getDocumentRow:    db.GetArchiveDocumentDetailRow{ID: documentID, CategoryID: categoryID, Title: "SK Tim", Status: "active"},
		updateDocumentRow: document,
	}
	h := &Archive{svc: fake}

	rec := httptest.NewRecorder()
	h.ListDocuments(rec, adminRequest(http.MethodGet, "/api/tu/archives/documents?search=sk&category_id="+categoryID.String()+"&status=active&classification_code=420", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListDocuments() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listDocumentSearch != "sk" || fake.listDocumentCategoryID != categoryID || fake.listDocumentStatus != "active" || fake.listDocumentClassificationCode != "420" {
		t.Fatalf("ListDocuments filters = %q/%v/%q/%q, want forwarded filters", fake.listDocumentSearch, fake.listDocumentCategoryID, fake.listDocumentStatus, fake.listDocumentClassificationCode)
	}

	rec = httptest.NewRecorder()
	uploadReq := newArchiveMultipartRequest(t, map[string]string{
		"category_id":      categoryID.String(),
		"title":            "SK Tim",
		"archive_number":   "AK-001",
		"document_date":    "2026-04-01",
		"received_date":    "2026-04-02",
		"summary":          "Rangkuman",
		"tags":             "sk,tim",
		"status":           "active",
		"storage_location": "Lemari A",
		"retention_until":  "2031-04-02",
	}, true)
	h.UploadDocument(rec, uploadReq)
	if rec.Code != http.StatusCreated {
		t.Fatalf("UploadDocument() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.uploadInput.CategoryID != categoryID || fake.uploadInput.Title != "SK Tim" || fake.uploadInput.FileSize != 7 || fake.uploadContent != "archive" || !fake.uploadInput.UploadedByUserID.Valid {
		t.Fatalf("UploadDocument input = %+v content=%q, want request fields/user/file", fake.uploadInput, fake.uploadContent)
	}

	rec = httptest.NewRecorder()
	h.GetDocument(rec, withRouteParam(adminRequest(http.MethodGet, "/api/tu/archives/documents/"+documentID.String(), ""), "id", documentID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetDocument() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.getDocumentID != documentID {
		t.Fatalf("GetDocument id = %v, want %v", fake.getDocumentID, documentID)
	}

	rec = httptest.NewRecorder()
	updateReq := withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{"category_id":"`+categoryID.String()+`","title":"SK Tim Revisi","archive_number":"AK-002","document_date":"2026-04-03","received_date":"2026-04-04","summary":"Ringkas","tags":"sk","status":"borrowed","storage_location":"Lemari B","retention_until":"2031-04-04"}`), "id", documentID.String())
	h.UpdateDocument(rec, updateReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateDocument() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateDocumentArg.ID != documentID || fake.updateDocumentArg.CategoryID != categoryID || fake.updateDocumentArg.Title != "SK Tim Revisi" || fake.updateDocumentArg.Status != "borrowed" {
		t.Fatalf("UpdateDocument arg = %+v, want route/body fields", fake.updateDocumentArg)
	}

	rec = httptest.NewRecorder()
	h.DeleteDocument(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/tu/archives/documents/"+documentID.String(), ""), "id", documentID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteDocument() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteDocumentID != documentID {
		t.Fatalf("DeleteDocument id = %v, want %v", fake.deleteDocumentID, documentID)
	}
}

func TestArchiveMutationHandlersWriteAuditEvents(t *testing.T) {
	categoryID := handlerTestUUID(62)
	documentID := handlerTestUUID(63)
	category := db.ArchiveCategory{ID: categoryID, Code: "AK", Name: "Akademik", RetentionYears: 5, IsActive: true}
	document := db.ArchiveDocument{ID: documentID, CategoryID: categoryID, Title: "SK Tim", ArchiveNumber: "AK-001", OriginalName: "arsip.pdf", MimeType: "application/pdf", FileSize: 7, Status: "active"}
	fake := &fakeArchiveService{
		createCategoryRow: category,
		updateCategoryRow: db.ArchiveCategory{ID: categoryID, Code: "AK", Name: "Akademik Baru", RetentionYears: 7, IsActive: true},
		uploadRow:         document,
		updateDocumentRow: db.ArchiveDocument{ID: documentID, CategoryID: categoryID, Title: "SK Tim Revisi", ArchiveNumber: "AK-002", Status: "borrowed"},
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &Archive{svc: fake, audit: audit}
	req := func(method, target, body string) *http.Request {
		return withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"staf"},
			"uid":   "01000000-0000-0000-0000-000000000000",
			"sub":   "01000000-0000-0000-0000-000000000000",
			"usr":   "staf.arsip",
			"ssid":  "sess-archive-1",
		})
	}

	rec := httptest.NewRecorder()
	h.CreateCategory(rec, req(http.MethodPost, "/api/tu/archives/categories", `{"code":"ak","name":"Akademik","classification_code":"420","description":"Dokumen akademik","retention_years":5,"is_active":false}`))
	rec = httptest.NewRecorder()
	h.UpdateCategory(rec, withRouteParam(req(http.MethodPatch, "/api/tu/archives/categories/"+categoryID.String(), `{"code":"ak","name":"Akademik Baru","retention_years":7}`), "id", categoryID.String()))
	rec = httptest.NewRecorder()
	h.DeleteCategory(rec, withRouteParam(req(http.MethodDelete, "/api/tu/archives/categories/"+categoryID.String(), ""), "id", categoryID.String()))

	rec = httptest.NewRecorder()
	uploadReq := newArchiveMultipartRequest(t, map[string]string{
		"category_id":      categoryID.String(),
		"title":            "SK Tim",
		"archive_number":   "AK-001",
		"document_date":    "2026-04-01",
		"received_date":    "2026-04-02",
		"summary":          "Rangkuman",
		"tags":             "sk,tim",
		"status":           "active",
		"storage_location": "Lemari A",
		"retention_until":  "2031-04-02",
	}, true)
	uploadReq = withClaims(uploadReq, jwt.MapClaims{
		"roles": []any{"staf"},
		"uid":   "01000000-0000-0000-0000-000000000000",
		"sub":   "01000000-0000-0000-0000-000000000000",
		"usr":   "staf.arsip",
		"ssid":  "sess-archive-1",
	})
	h.UploadDocument(rec, uploadReq)
	rec = httptest.NewRecorder()
	h.UpdateDocument(rec, withRouteParam(req(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{"category_id":"`+categoryID.String()+`","title":"SK Tim Revisi","archive_number":"AK-002","document_date":"2026-04-03","received_date":"2026-04-04","summary":"Ringkas","tags":"sk","status":"borrowed","storage_location":"Lemari B","retention_until":"2031-04-04"}`), "id", documentID.String()))
	rec = httptest.NewRecorder()
	h.DeleteDocument(rec, withRouteParam(req(http.MethodDelete, "/api/tu/archives/documents/"+documentID.String(), ""), "id", documentID.String()))

	if len(audit.entries) != 6 {
		t.Fatalf("audit entries = %d, want 6", len(audit.entries))
	}
	if audit.entries[0].Action != "ARCHIVE_CATEGORY_CREATE" || audit.entries[3].Action != "ARCHIVE_DOCUMENT_UPLOAD" {
		t.Fatalf("unexpected archive audit actions = %+v", audit.entries)
	}
	meta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if meta["name"] != "Akademik" || meta["username"] != "staf.arsip" {
		t.Fatalf("archive category create metadata = %+v, want name/username", meta)
	}
	meta = mustAuditMetadataMap(t, audit.entries[4].Metadata)
	if meta["archive_number"] != "AK-002" || meta["status"] != "borrowed" {
		t.Fatalf("archive document update metadata = %+v, want archive_number/status", meta)
	}
}

func TestArchiveFileServesStoredDocument(t *testing.T) {
	documentID := handlerTestUUID(70)
	filePath := t.TempDir() + "/arsip.pdf"
	if err := os.WriteFile(filePath, []byte("archive"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	fake := &fakeArchiveService{getFileRow: db.ArchiveDocument{
		ID:           documentID,
		OriginalName: "arsip.pdf",
		FilePath:     filePath,
		MimeType:     "application/pdf",
		FileSize:     7,
	}}
	rec := httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodGet, "/api/tu/archives/documents/"+documentID.String()+"/file", ""), "id", documentID.String())

	(&Archive{svc: fake}).File(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("File() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "archive" || fake.getFileID != documentID {
		t.Fatalf("File() body/id = %q/%v, want archive/id", rec.Body.String(), fake.getFileID)
	}
	if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, `filename=arsip.pdf`) {
		t.Fatalf("Content-Disposition = %q, want archive filename", got)
	}
}

func TestArchiveHandlersMapValidationAndServiceErrors(t *testing.T) {
	categoryID := handlerTestUUID(80)
	documentID := handlerTestUUID(81)
	tests := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		req  *http.Request
		want int
	}{
		{
			name: "stats internal",
			fn:   (&Archive{svc: &fakeArchiveService{statsErr: errors.New("db down")}}).Stats,
			req:  adminRequest(http.MethodGet, "/api/tu/archives/stats", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "list category internal",
			fn:   (&Archive{svc: &fakeArchiveService{listCategoryErr: errors.New("db down")}}).ListCategories,
			req:  adminRequest(http.MethodGet, "/api/tu/archives/categories", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "create category invalid json",
			fn:   (&Archive{svc: &fakeArchiveService{}}).CreateCategory,
			req:  adminRequest(http.MethodPost, "/api/tu/archives/categories", `{`),
			want: http.StatusBadRequest,
		},
		{
			name: "create category service error",
			fn:   (&Archive{svc: &fakeArchiveService{createCategoryErr: errors.New("code wajib diisi")}}).CreateCategory,
			req:  adminRequest(http.MethodPost, "/api/tu/archives/categories", `{"code":"","name":""}`),
			want: http.StatusBadRequest,
		},
		{
			name: "update category invalid id",
			fn:   (&Archive{svc: &fakeArchiveService{}}).UpdateCategory,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/categories/bad", `{}`), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "update category invalid json",
			fn:   (&Archive{svc: &fakeArchiveService{}}).UpdateCategory,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/categories/"+categoryID.String(), `{`), "id", categoryID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "update category service error",
			fn:   (&Archive{svc: &fakeArchiveService{updateCategoryErr: errors.New("kode wajib diisi")}}).UpdateCategory,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/categories/"+categoryID.String(), `{"code":"","name":""}`), "id", categoryID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "delete category invalid id",
			fn:   (&Archive{svc: &fakeArchiveService{}}).DeleteCategory,
			req:  withRouteParam(adminRequest(http.MethodDelete, "/api/tu/archives/categories/bad", ""), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "delete category service error",
			fn:   (&Archive{svc: &fakeArchiveService{deleteCategoryErr: errors.New("masih dipakai")}}).DeleteCategory,
			req:  withRouteParam(adminRequest(http.MethodDelete, "/api/tu/archives/categories/"+categoryID.String(), ""), "id", categoryID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "list documents invalid category",
			fn:   (&Archive{svc: &fakeArchiveService{}}).ListDocuments,
			req:  adminRequest(http.MethodGet, "/api/tu/archives/documents?category_id=bad", ""),
			want: http.StatusBadRequest,
		},
		{
			name: "list documents internal",
			fn:   (&Archive{svc: &fakeArchiveService{listDocumentErr: errors.New("db down")}}).ListDocuments,
			req:  adminRequest(http.MethodGet, "/api/tu/archives/documents", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "get document not found",
			fn:   (&Archive{svc: &fakeArchiveService{getDocumentErr: pgx.ErrNoRows}}).GetDocument,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/tu/archives/documents/"+documentID.String(), ""), "id", documentID.String()),
			want: http.StatusNotFound,
		},
		{
			name: "get document invalid id",
			fn:   (&Archive{svc: &fakeArchiveService{}}).GetDocument,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/tu/archives/documents/bad", ""), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "get document internal",
			fn:   (&Archive{svc: &fakeArchiveService{getDocumentErr: errors.New("db down")}}).GetDocument,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/tu/archives/documents/"+documentID.String(), ""), "id", documentID.String()),
			want: http.StatusInternalServerError,
		},
		{
			name: "update document invalid id",
			fn:   (&Archive{svc: &fakeArchiveService{}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/bad", `{}`), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "update document invalid json",
			fn:   (&Archive{svc: &fakeArchiveService{}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{`), "id", documentID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "update document invalid category",
			fn:   (&Archive{svc: &fakeArchiveService{}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{"category_id":"bad"}`), "id", documentID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "update document invalid date",
			fn:   (&Archive{svc: &fakeArchiveService{}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{"category_id":"`+categoryID.String()+`","document_date":"bad"}`), "id", documentID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "update document invalid received date",
			fn:   (&Archive{svc: &fakeArchiveService{}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{"received_date":"bad"}`), "id", documentID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "update document invalid retention date",
			fn:   (&Archive{svc: &fakeArchiveService{}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{"retention_until":"bad"}`), "id", documentID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "update document not found",
			fn:   (&Archive{svc: &fakeArchiveService{updateDocumentErr: pgx.ErrNoRows}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{"category_id":"`+categoryID.String()+`","title":"SK","received_date":"2026-01-01"}`), "id", documentID.String()),
			want: http.StatusNotFound,
		},
		{
			name: "update document service error",
			fn:   (&Archive{svc: &fakeArchiveService{updateDocumentErr: errors.New("judul wajib diisi")}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPatch, "/api/tu/archives/documents/"+documentID.String(), `{"title":""}`), "id", documentID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "delete document invalid id",
			fn:   (&Archive{svc: &fakeArchiveService{}}).DeleteDocument,
			req:  withRouteParam(adminRequest(http.MethodDelete, "/api/tu/archives/documents/bad", ""), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "delete document not found",
			fn:   (&Archive{svc: &fakeArchiveService{deleteDocumentErr: pgx.ErrNoRows}}).DeleteDocument,
			req:  withRouteParam(adminRequest(http.MethodDelete, "/api/tu/archives/documents/"+documentID.String(), ""), "id", documentID.String()),
			want: http.StatusNotFound,
		},
		{
			name: "delete document service error",
			fn:   (&Archive{svc: &fakeArchiveService{deleteDocumentErr: errors.New("masih dipinjam")}}).DeleteDocument,
			req:  withRouteParam(adminRequest(http.MethodDelete, "/api/tu/archives/documents/"+documentID.String(), ""), "id", documentID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "file invalid id",
			fn:   (&Archive{svc: &fakeArchiveService{}}).File,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/tu/archives/documents/bad/file", ""), "id", "bad"),
			want: http.StatusBadRequest,
		},
		{
			name: "file missing row",
			fn:   (&Archive{svc: &fakeArchiveService{getFileErr: errors.New("not found")}}).File,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/tu/archives/documents/"+documentID.String()+"/file", ""), "id", documentID.String()),
			want: http.StatusNotFound,
		},
		{
			name: "file missing disk path",
			fn:   (&Archive{svc: &fakeArchiveService{getFileRow: db.ArchiveDocument{ID: documentID, FilePath: "/path/does/not/exist"}}}).File,
			req:  withRouteParam(adminRequest(http.MethodGet, "/api/tu/archives/documents/"+documentID.String()+"/file", ""), "id", documentID.String()),
			want: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.want, rec.Body.String())
			}
		})
	}

	rec := httptest.NewRecorder()
	(&Archive{svc: &fakeArchiveService{}}).UploadDocument(rec, adminRequest(http.MethodPost, "/api/tu/archives/documents", "not multipart"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadDocument(invalid multipart) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Archive{svc: &fakeArchiveService{}}).UploadDocument(rec, newArchiveMultipartRequest(t, map[string]string{"category_id": "bad"}, true))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadDocument(invalid category) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Archive{svc: &fakeArchiveService{}}).UploadDocument(rec, newArchiveMultipartRequest(t, map[string]string{"document_date": "bad"}, true))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadDocument(invalid date) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Archive{svc: &fakeArchiveService{}}).UploadDocument(rec, newArchiveMultipartRequest(t, map[string]string{"received_date": "bad"}, true))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadDocument(invalid received date) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Archive{svc: &fakeArchiveService{}}).UploadDocument(rec, newArchiveMultipartRequest(t, map[string]string{"retention_until": "bad"}, true))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadDocument(invalid retention date) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Archive{svc: &fakeArchiveService{}}).UploadDocument(rec, newArchiveMultipartRequest(t, map[string]string{}, false))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadDocument(missing file) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Archive{svc: &fakeArchiveService{uploadErr: errors.New("file terlalu besar")}}).UploadDocument(rec, newArchiveMultipartRequest(t, map[string]string{}, true))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("UploadDocument(service error) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	if got := archiveClientMessage(errors.New("pesan aman")); got != "pesan aman" {
		t.Fatalf("archiveClientMessage() = %q, want pesan aman", got)
	}
}
