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
	"github.com/jackc/pgx/v5/pgtype"

	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeCbtQuestionAssetService struct {
	saveInput   service.UploadCbtQuestionAssetInput
	saveContent string
	saveRow     db.CbtQuestionAsset
	saveErr     error

	questionID  pgtype.UUID
	questionRow db.GetCbtQuestionRow
	questionErr error

	listQuestionID pgtype.UUID
	listRows       []db.CbtQuestionAsset
	listErr        error

	getID  pgtype.UUID
	getRow db.CbtQuestionAsset
	getErr error

	accessQuestionID pgtype.UUID
	accessPackageID  pgtype.UUID
	accessAllowed    bool
	accessErr        error
	openErr          error
}

func (f *fakeCbtQuestionAssetService) Save(_ context.Context, input service.UploadCbtQuestionAssetInput) (db.CbtQuestionAsset, error) {
	f.saveInput = input
	if input.File != nil {
		content, err := io.ReadAll(input.File)
		if err != nil {
			return db.CbtQuestionAsset{}, err
		}
		f.saveContent = string(content)
	}
	if f.saveErr != nil {
		return db.CbtQuestionAsset{}, f.saveErr
	}
	return f.saveRow, nil
}

func (f *fakeCbtQuestionAssetService) GetQuestion(_ context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	f.questionID = id
	if f.questionErr != nil {
		return db.GetCbtQuestionRow{}, f.questionErr
	}
	return f.questionRow, nil
}

func (f *fakeCbtQuestionAssetService) ListByQuestion(_ context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error) {
	f.listQuestionID = questionID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listRows, nil
}

func (f *fakeCbtQuestionAssetService) Get(_ context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error) {
	f.getID = id
	if f.getErr != nil {
		return db.CbtQuestionAsset{}, f.getErr
	}
	return f.getRow, nil
}

func (f *fakeCbtQuestionAssetService) AccessibleByPackage(_ context.Context, questionID, packageID pgtype.UUID) (bool, error) {
	f.accessQuestionID = questionID
	f.accessPackageID = packageID
	if f.accessErr != nil {
		return false, f.accessErr
	}
	return f.accessAllowed, nil
}

func (f *fakeCbtQuestionAssetService) Open(asset db.CbtQuestionAsset) (*os.File, error) {
	if f.openErr != nil {
		return nil, f.openErr
	}
	return os.Open(asset.StoragePath)
}

func newMultipartAssetRequest(t *testing.T, fields map[string]string, withFile bool) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", key, err)
		}
	}
	if withFile {
		part, err := writer.CreateFormFile("file", "diagram.png")
		if err != nil {
			t.Fatalf("CreateFormFile() error = %v", err)
		}
		if _, err := part.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\b\x02\x00\x00\x00")); err != nil {
			t.Fatalf("multipart file write error = %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/cbt/assets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "usr": "operator.cbt"})
}

func newOversizedMultipartAssetRequest(t *testing.T, questionID pgtype.UUID) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("question_id", questionID.String()); err != nil {
		t.Fatalf("WriteField(question_id) error = %v", err)
	}
	part, err := writer.CreateFormFile("file", "diagram.png")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(bytes.Repeat([]byte("x"), (12<<20)+1)); err != nil {
		t.Fatalf("multipart file write error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart Close() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/cbt/assets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "usr": "operator.cbt"})
}

func TestCbtQuestionAssetUploadForwardsValidMultipart(t *testing.T) {
	questionID := handlerTestUUID(1)
	assetID := handlerTestUUID(2)
	fake := &fakeCbtQuestionAssetService{
		questionRow: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "operator.cbt"},
		saveRow: db.CbtQuestionAsset{
			ID:           assetID,
			QuestionID:   questionID,
			OriginalName: "diagram.png",
			MimeType:     "image/png",
			FileSize:     3,
			Purpose:      "stimulus",
		},
	}
	h := &CbtQuestionAsset{svc: fake}
	req := newMultipartAssetRequest(t, map[string]string{
		"question_id": questionID.String(),
		"purpose":     "stimulus",
	}, true)
	rec := httptest.NewRecorder()

	h.Upload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Upload() status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if fake.saveInput.QuestionID != questionID {
		t.Fatalf("Save.QuestionID = %v, want %v", fake.saveInput.QuestionID, questionID)
	}
	if fake.saveInput.OriginalName != "diagram.png" || fake.saveInput.Purpose != "stimulus" || fake.saveInput.UploadedBy != "operator.cbt" {
		t.Fatalf("Save input = %+v, want filename/purpose/uploader forwarded", fake.saveInput)
	}
	if fake.saveInput.FileSize == 0 || !strings.HasPrefix(fake.saveContent, "\x89PNG") {
		t.Fatalf("Save file size/content = %d/%q, want detected png", fake.saveInput.FileSize, fake.saveContent)
	}
	if !strings.Contains(rec.Body.String(), `/api/cbt/assets/`+assetID.String()+`/file`) {
		t.Fatalf("Upload() body missing asset URL: %s", rec.Body.String())
	}
}

func TestCbtQuestionAssetUploadRejectsInvalidRequests(t *testing.T) {
	h := &CbtQuestionAsset{svc: &fakeCbtQuestionAssetService{}}
	tests := []struct {
		name string
		req  *http.Request
	}{
		{
			name: "forbidden without cbt role",
			req:  httptest.NewRequest(http.MethodPost, "/api/cbt/assets", nil),
		},
		{
			name: "invalid multipart",
			req:  withClaims(httptest.NewRequest(http.MethodPost, "/api/cbt/assets", strings.NewReader("not multipart")), jwt.MapClaims{"roles": []any{"admin"}}),
		},
		{
			name: "missing file",
			req:  newMultipartAssetRequest(t, map[string]string{"question_id": handlerTestUUID(1).String()}, false),
		},
		{
			name: "invalid question id",
			req:  newMultipartAssetRequest(t, map[string]string{"question_id": "bad"}, true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.Upload(rec, tt.req)
			if rec.Code != http.StatusBadRequest && rec.Code != http.StatusForbidden {
				t.Fatalf("Upload() status = %d, want 400/403; body=%s", rec.Code, rec.Body.String())
			}
		})
	}

	failing := &CbtQuestionAsset{svc: &fakeCbtQuestionAssetService{saveErr: errors.New("mime_type tidak didukung")}}
	rec := httptest.NewRecorder()
	failing.Upload(rec, newMultipartAssetRequest(t, map[string]string{"question_id": handlerTestUUID(11).String()}, true))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Upload(service error) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtQuestionAssetUploadRejectsOversizedMultipart(t *testing.T) {
	questionID := handlerTestUUID(252)
	fake := &fakeCbtQuestionAssetService{questionRow: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "operator.cbt"}}
	h := &CbtQuestionAsset{svc: fake}
	rec := httptest.NewRecorder()

	h.Upload(rec, newOversizedMultipartAssetRequest(t, questionID))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Upload(oversized) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if fake.saveInput.FileSize != 0 {
		t.Fatalf("Upload(oversized) reached service with input %+v", fake.saveInput)
	}
}

func TestCbtQuestionAssetListHandlesQueryAndStoreErrors(t *testing.T) {
	questionID := handlerTestUUID(3)
	fake := &fakeCbtQuestionAssetService{
		questionRow: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "operator.cbt"},
		listRows: []db.CbtQuestionAsset{
			{
				ID:           handlerTestUUID(4),
				QuestionID:   questionID,
				OriginalName: "audio.mp3",
				MimeType:     "audio/mpeg",
				FileSize:     9,
				Purpose:      "supporting",
			},
		},
	}
	h := &CbtQuestionAsset{svc: fake}
	rec := httptest.NewRecorder()
	req := adminRequest(http.MethodGet, "/api/cbt/assets?question_id="+questionID.String(), "")

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listQuestionID != questionID {
		t.Fatalf("ListByQuestion id = %v, want %v", fake.listQuestionID, questionID)
	}
	if !strings.Contains(rec.Body.String(), "audio.mp3") {
		t.Fatalf("List() body missing asset name: %s", rec.Body.String())
	}

	invalidCases := []struct {
		name   string
		target string
		want   int
		svc    *fakeCbtQuestionAssetService
	}{
		{name: "missing question id", target: "/api/cbt/assets", want: http.StatusBadRequest, svc: &fakeCbtQuestionAssetService{}},
		{name: "invalid question id", target: "/api/cbt/assets?question_id=bad", want: http.StatusBadRequest, svc: &fakeCbtQuestionAssetService{}},
		{name: "store error", target: "/api/cbt/assets?question_id=" + questionID.String(), want: http.StatusInternalServerError, svc: &fakeCbtQuestionAssetService{listErr: errors.New("db down")}},
	}
	for _, tt := range invalidCases {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&CbtQuestionAsset{svc: tt.svc}).List(rec, adminRequest(http.MethodGet, tt.target, ""))
			if rec.Code != tt.want {
				t.Fatalf("List() status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestCbtQuestionAssetTeacherScopeRequiresQuestionOwnership(t *testing.T) {
	questionID := handlerTestUUID(12)

	t.Run("upload forbids teacher without question ownership", func(t *testing.T) {
		fake := &fakeCbtQuestionAssetService{questionRow: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.lain"}}
		h := &CbtQuestionAsset{svc: fake}
		req := newMultipartAssetRequest(t, map[string]string{"question_id": questionID.String()}, true)
		req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa", "sub": "01000000-0000-0000-0000-000000000010"})
		rec := httptest.NewRecorder()
		h.Upload(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("Upload() status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("upload forbids teacher orphan asset", func(t *testing.T) {
		h := &CbtQuestionAsset{svc: &fakeCbtQuestionAssetService{}}
		req := newMultipartAssetRequest(t, map[string]string{}, true)
		req = withClaims(req, jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa", "sub": "01000000-0000-0000-0000-000000000010"})
		rec := httptest.NewRecorder()
		h.Upload(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("Upload(orphan) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("list forbids teacher without question ownership", func(t *testing.T) {
		fake := &fakeCbtQuestionAssetService{questionRow: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.lain"}}
		h := &CbtQuestionAsset{svc: fake}
		req := withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/assets?question_id="+questionID.String(), nil), jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa", "sub": "01000000-0000-0000-0000-000000000010"})
		rec := httptest.NewRecorder()
		h.List(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("List() status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestCbtQuestionAssetRequireScopeBranches(t *testing.T) {
	questionID := handlerTestUUID(13)
	tests := []struct {
		name       string
		claims     jwt.MapClaims
		questionID pgtype.UUID
		svc        *fakeCbtQuestionAssetService
		wantOK     bool
		wantStatus int
	}{
		{
			name:       "admin bypasses author lookup",
			claims:     jwt.MapClaims{"roles": []any{"admin"}, "usr": "admin.cbt"},
			questionID: pgtype.UUID{},
			svc:        &fakeCbtQuestionAssetService{questionErr: errors.New("should not be called")},
			wantOK:     true,
		},
		{
			name:       "missing claims forbidden",
			questionID: questionID,
			svc:        &fakeCbtQuestionAssetService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "non guru forbidden",
			claims:     jwt.MapClaims{"roles": []any{"staf"}, "usr": "tu"},
			questionID: questionID,
			svc:        &fakeCbtQuestionAssetService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "guru orphan asset forbidden",
			claims:     jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"},
			questionID: pgtype.UUID{},
			svc:        &fakeCbtQuestionAssetService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "guru missing username forbidden",
			claims:     jwt.MapClaims{"roles": []any{"guru"}},
			questionID: questionID,
			svc:        &fakeCbtQuestionAssetService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "question lookup error internal",
			claims:     jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"},
			questionID: questionID,
			svc:        &fakeCbtQuestionAssetService{questionErr: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "author mismatch forbidden",
			claims:     jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"},
			questionID: questionID,
			svc:        &fakeCbtQuestionAssetService{questionRow: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.matematika"}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "author allowed",
			claims:     jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"},
			questionID: questionID,
			svc:        &fakeCbtQuestionAssetService{questionRow: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.ipa"}},
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/cbt/assets", nil)
			if tt.claims != nil {
				req = withClaims(req, tt.claims)
			}
			rec := httptest.NewRecorder()
			gotOK := (&CbtQuestionAsset{svc: tt.svc}).requireQuestionAssetScope(rec, req, tt.questionID)
			if gotOK != tt.wantOK {
				t.Fatalf("requireQuestionAssetScope() = %v, want %v", gotOK, tt.wantOK)
			}
			if !tt.wantOK && rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtQuestionAssetFileServesFileAndChecksParticipantAccess(t *testing.T) {
	assetID := handlerTestUUID(5)
	questionID := handlerTestUUID(6)
	filePath := t.TempDir() + "/asset.txt"
	if err := os.WriteFile(filePath, []byte("asset-body"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	baseRow := db.CbtQuestionAsset{
		ID:          assetID,
		QuestionID:  questionID,
		StoredName:  "asset.txt",
		MimeType:    "text/plain",
		FileSize:    int64(len("asset-body")),
		StoragePath: filePath,
	}

	h := &CbtQuestionAsset{svc: &fakeCbtQuestionAssetService{getRow: baseRow}}
	rec := httptest.NewRecorder()
	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/cbt/assets/"+assetID.String()+"/file", nil), "id", assetID.String())
	req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "usr": "operator.cbt"})

	h.File(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("File() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "asset-body" {
		t.Fatalf("File() body = %q, want asset-body", rec.Body.String())
	}
	if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, `filename="asset.txt"`) {
		t.Fatalf("Content-Disposition = %q, want inline filename", got)
	}

	deniedFake := &fakeCbtQuestionAssetService{getRow: baseRow, accessAllowed: false}
	participantReq := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/cbt/assets/"+assetID.String()+"/file", nil), "id", assetID.String())
	participantReq = participantReq.WithContext(context.WithValue(participantReq.Context(), mw.ExamParticipantKey, db.GetParticipantByTokenRow{
		PackageID: handlerTestUUID(7),
	}))
	rec = httptest.NewRecorder()
	(&CbtQuestionAsset{svc: deniedFake}).File(rec, participantReq)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("File(participant denied) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	if deniedFake.accessQuestionID != questionID || deniedFake.accessPackageID != handlerTestUUID(7) {
		t.Fatalf("AccessibleByPackage ids = %v/%v, want question/package", deniedFake.accessQuestionID, deniedFake.accessPackageID)
	}

	teacherDeniedFake := &fakeCbtQuestionAssetService{
		getRow:      baseRow,
		questionRow: db.GetCbtQuestionRow{ID: questionID, AuthorUsername: "guru.lain"},
	}
	teacherReq := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/cbt/assets/"+assetID.String()+"/file", nil), "id", assetID.String())
	teacherReq = withClaims(teacherReq, jwt.MapClaims{"roles": []any{"guru"}, "usr": "guru.ipa"})
	rec = httptest.NewRecorder()
	(&CbtQuestionAsset{svc: teacherDeniedFake}).File(rec, teacherReq)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("File(non-author teacher) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	if teacherDeniedFake.questionID != questionID {
		t.Fatalf("GetQuestion id = %v, want %v", teacherDeniedFake.questionID, questionID)
	}
}

func TestCbtQuestionAssetFileRejectsInvalidAndMissingAssets(t *testing.T) {
	assetID := handlerTestUUID(8)
	tests := []struct {
		name string
		id   string
		svc  *fakeCbtQuestionAssetService
		want int
	}{
		{name: "invalid id", id: "bad", svc: &fakeCbtQuestionAssetService{}, want: http.StatusBadRequest},
		{name: "missing row", id: assetID.String(), svc: &fakeCbtQuestionAssetService{getErr: errors.New("not found")}, want: http.StatusNotFound},
		{name: "missing file", id: assetID.String(), svc: &fakeCbtQuestionAssetService{getRow: db.CbtQuestionAsset{ID: assetID, StoragePath: "/path/does/not/exist"}}, want: http.StatusNotFound},
		{name: "access error", id: assetID.String(), svc: &fakeCbtQuestionAssetService{getRow: db.CbtQuestionAsset{ID: assetID, QuestionID: handlerTestUUID(9)}, accessErr: errors.New("db down")}, want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/cbt/assets/"+tt.id+"/file", nil), "id", tt.id)
			if tt.name == "access error" {
				req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, db.GetParticipantByTokenRow{PackageID: handlerTestUUID(10)}))
			}
			if tt.name == "missing file" {
				req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}, "usr": "operator.cbt"})
			}
			rec := httptest.NewRecorder()
			(&CbtQuestionAsset{svc: tt.svc}).File(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("File() status = %d, want %d; body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
