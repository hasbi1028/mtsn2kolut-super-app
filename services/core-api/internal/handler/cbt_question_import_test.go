package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/service"
)

func newCbtQuestionImportMultipartRequest(t *testing.T, fields map[string]string, filename string, content string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", key, err)
		}
	}
	if filename != "" {
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			t.Fatalf("CreateFormFile() error = %v", err)
		}
		if _, err := part.Write([]byte(content)); err != nil {
			t.Fatalf("file write error = %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart close error = %v", err)
	}
	req := adminRequest(http.MethodPost, "/api/cbt/questions/import-legacy", "")
	req.Body = ioNopCloserBytes(body.Bytes())
	req.ContentLength = int64(body.Len())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func ioNopCloserBytes(data []byte) *readCloserBytes {
	return &readCloserBytes{Reader: bytes.NewReader(data)}
}

type readCloserBytes struct{ *bytes.Reader }

func (r *readCloserBytes) Close() error { return nil }

func TestCbtQuestionImportLegacyCSVExtraRejectsBadSubjectEventAndMissingFile(t *testing.T) {
	validSubject := handlerTestUUID(61).String()
	validEvent := handlerTestUUID(62).String()
	cases := []struct {
		name   string
		fields map[string]string
		file   bool
		want   string
	}{
		{name: "bad subject", fields: map[string]string{"subject_id": "not-a-uuid"}, file: true, want: "Mata pelajaran tidak valid"},
		{name: "bad event", fields: map[string]string{"subject_id": validSubject, "event_id": "not-a-uuid"}, file: true, want: "Kegiatan asesmen tidak valid"},
		{name: "missing file", fields: map[string]string{"subject_id": validSubject, "event_id": validEvent}, file: false, want: "file CSV wajib diisi"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCbtQuestionService{}
			h := &CbtQuestion{svc: fake}
			filename := ""
			if tt.file {
				filename = "legacy.csv"
			}
			req := newCbtQuestionImportMultipartRequest(t, tt.fields, filename, "kode,soal\nQ,Soal")
			rec := httptest.NewRecorder()

			h.ImportLegacyCSV(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("ImportLegacyCSV status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tt.want) {
				t.Fatalf("ImportLegacyCSV body = %s, want substring %q", rec.Body.String(), tt.want)
			}
			if fake.importInput.SubjectID.Valid {
				t.Fatalf("ImportLegacyCSV reached service with input %+v", fake.importInput)
			}
		})
	}
}

func TestCbtQuestionImportLegacyCSVExtraSuccessPassesParsedInput(t *testing.T) {
	subjectID := handlerTestUUID(71)
	eventID := handlerTestUUID(72)
	csvText := "kode,soal\nQ-1,Soal"
	fake := &fakeCbtQuestionService{
		importResult: service.ImportLegacyQuestionsResult{TotalRows: 1, DryRun: true, WouldImport: 1, Errors: []string{}, DuplicateCodes: []string{}},
	}
	h := &CbtQuestion{svc: fake}
	req := newCbtQuestionImportMultipartRequest(t, map[string]string{
		"subject_id": subjectID.String(),
		"event_id":   eventID.String(),
		"dry_run":    "true",
	}, "legacy.csv", csvText)
	rec := httptest.NewRecorder()

	h.ImportLegacyCSV(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ImportLegacyCSV status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !sameHandlerTestUUID(fake.importInput.SubjectID, subjectID) || !sameHandlerTestUUID(fake.importInput.EventID, eventID) {
		t.Fatalf("service input IDs = subject %+v event %+v, want %+v %+v", fake.importInput.SubjectID, fake.importInput.EventID, subjectID, eventID)
	}
	if fake.importInput.CSVText != csvText || !fake.importInput.DryRun {
		t.Fatalf("service input CSV/dry_run = %q/%v, want %q/true", fake.importInput.CSVText, fake.importInput.DryRun, csvText)
	}
	if !fake.importInput.Actor.IsAdmin() {
		t.Fatalf("service actor = %+v, want admin actor from request", fake.importInput.Actor)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response JSON error = %v; body=%s", err, rec.Body.String())
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("response body = %#v, want data object", body)
	}
	if data["total_rows"] != float64(1) || data["would_import"] != float64(1) || data["dry_run"] != true {
		t.Fatalf("response data = %#v, want import result", data)
	}
}

func sameHandlerTestUUID(a, b pgtype.UUID) bool {
	return a.Valid && b.Valid && a.Bytes == b.Bytes
}
