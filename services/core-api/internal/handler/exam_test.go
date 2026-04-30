package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeExamService struct {
	loginResult  service.LoginResult
	loginErr     error
	statusResult service.StatusResult
	statusErr    error
	answerErr    error
	submitErr    error
}

func (f *fakeExamService) Login(ctx context.Context, token, deviceFingerprint, loginIP string) (service.LoginResult, error) {
	return f.loginResult, f.loginErr
}

func (f *fakeExamService) GetStatus(ctx context.Context, p db.GetParticipantByTokenRow) (service.StatusResult, error) {
	return f.statusResult, f.statusErr
}

func (f *fakeExamService) Heartbeat(ctx context.Context, participantID pgtype.UUID) error {
	return nil
}

func (f *fakeExamService) RecordClientEvent(ctx context.Context, participantID pgtype.UUID, eventType string, data map[string]any) error {
	return nil
}

func (f *fakeExamService) SubmitAnswer(ctx context.Context, p db.GetParticipantByTokenRow, questionID pgtype.UUID, answer string) error {
	return f.answerErr
}

func (f *fakeExamService) Submit(ctx context.Context, p db.GetParticipantByTokenRow) error {
	return f.submitErr
}

func TestAbsolutizeExamAssetURLHonorsForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", nil)
	req.Host = "internal:8080"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "cbt.mtsn2kolut.sch.id")

	got := absolutizeExamAssetURL(req, "/api/cbt/assets/asset-1/file")
	want := "https://cbt.mtsn2kolut.sch.id/api/cbt/assets/asset-1/file"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamAssetURLFallsBackToRequestHost(t *testing.T) {
	req := httptest.NewRequest("POST", "http://localhost:8080/api/exam/login", nil)
	req.Host = "localhost:8080"

	got := absolutizeExamAssetURL(req, "api/cbt/assets/asset-2/file")
	want := "http://localhost:8080/api/cbt/assets/asset-2/file"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamAssetURLLeavesAbsoluteURLUntouched(t *testing.T) {
	req := httptest.NewRequest("POST", "http://localhost:8080/api/exam/login", nil)

	got := absolutizeExamAssetURL(req, "https://cdn.example.com/file.png")
	want := "https://cdn.example.com/file.png"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamLoginResultRewritesAllMediaFields(t *testing.T) {
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "mobile-api.example.sch.id")

	result := service.LoginResult{
		Questions: []service.ExamQuestion{
			{
				StemMediaURL:     "/api/cbt/assets/image-1/file",
				StimulusMediaURL: "/api/cbt/assets/image-2/file",
				StemAudioURL:     "/api/cbt/assets/audio-1/file",
				StimulusAudioURL: "/api/cbt/assets/audio-2/file",
			},
		},
	}

	absolutizeExamLoginResult(req, &result)

	question := result.Questions[0]
	if question.StemMediaURL != "https://mobile-api.example.sch.id/api/cbt/assets/image-1/file" {
		t.Fatalf("StemMediaURL = %q", question.StemMediaURL)
	}
	if question.StimulusMediaURL != "https://mobile-api.example.sch.id/api/cbt/assets/image-2/file" {
		t.Fatalf("StimulusMediaURL = %q", question.StimulusMediaURL)
	}
	if question.StemAudioURL != "https://mobile-api.example.sch.id/api/cbt/assets/audio-1/file" {
		t.Fatalf("StemAudioURL = %q", question.StemAudioURL)
	}
	if question.StimulusAudioURL != "https://mobile-api.example.sch.id/api/cbt/assets/audio-2/file" {
		t.Fatalf("StimulusAudioURL = %q", question.StimulusAudioURL)
	}
}

func TestExamLoginWritesWrappedJSONWithAbsoluteMediaURLs(t *testing.T) {
	h := &Exam{
		svc: &fakeExamService{
			loginResult: service.LoginResult{
				ParticipantID: "participant-1",
				Student:       service.StudentInfo{NIS: "12345", Nama: "Ahmad"},
				Session:       service.SessionInfo{ID: "session-1", Title: "Ujian IPA"},
				Questions: []service.ExamQuestion{
					{
						ID:               "q-1",
						QuestionText:     "Soal 1",
						StemMediaURL:     "/api/cbt/assets/image-1/file",
						StimulusAudioURL: "/api/cbt/assets/audio-1/file",
					},
				},
			},
		},
	}

	body := bytes.NewBufferString(`{"token":"a1b2c3d4","device_fingerprint":"device-1"}`)
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "cbt.mtsn2kolut.sch.id")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Data service.LoginResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if payload.Data.Session.Title != "Ujian IPA" {
		t.Fatalf("Session.Title = %q", payload.Data.Session.Title)
	}
	if len(payload.Data.Questions) != 1 {
		t.Fatalf("len(Questions) = %d, want 1", len(payload.Data.Questions))
	}
	if payload.Data.Questions[0].StemMediaURL != "https://cbt.mtsn2kolut.sch.id/api/cbt/assets/image-1/file" {
		t.Fatalf("StemMediaURL = %q", payload.Data.Questions[0].StemMediaURL)
	}
	if payload.Data.Questions[0].StimulusAudioURL != "https://cbt.mtsn2kolut.sch.id/api/cbt/assets/audio-1/file" {
		t.Fatalf("StimulusAudioURL = %q", payload.Data.Questions[0].StimulusAudioURL)
	}
}

func TestExamStatusWritesWrappedJSON(t *testing.T) {
	h := &Exam{
		svc: &fakeExamService{
			statusResult: service.StatusResult{
				AnsweredCount:        7,
				TotalQuestions:       20,
				TimeRemainingSeconds: 1800,
				IsSubmitted:          true,
				SubmittedAt:          "2026-05-01T08:30:00Z",
			},
		},
	}

	req := httptest.NewRequest("GET", "http://internal/api/exam/status", nil)
	var participant db.GetParticipantByTokenRow
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.Status(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Data service.StatusResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if !payload.Data.IsSubmitted {
		t.Fatal("IsSubmitted = false, want true")
	}
	if payload.Data.SubmittedAt != "2026-05-01T08:30:00Z" {
		t.Fatalf("SubmittedAt = %q", payload.Data.SubmittedAt)
	}
}

func TestExamLoginMapsKnownServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{
			name:       "token not found",
			err:        service.ErrExamNotFound,
			wantStatus: 404,
			wantError:  "token not found",
		},
		{
			name:       "session not active",
			err:        service.ErrExamNotActive,
			wantStatus: 403,
			wantError:  "exam session is not active",
		},
		{
			name:       "device mismatch",
			err:        service.ErrDeviceMismatch,
			wantStatus: 409,
			wantError:  "token already bound to another device",
		},
		{
			name:       "unexpected error",
			err:        errors.New("database down"),
			wantStatus: 500,
			wantError:  "database down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Exam{svc: &fakeExamService{loginErr: tt.err}}
			body := bytes.NewBufferString(`{"token":"a1b2c3d4","device_fingerprint":"device-1"}`)
			req := httptest.NewRequest("POST", "http://internal/api/exam/login", body)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Login(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json unmarshal failed: %v", err)
			}
			if payload.Error != tt.wantError {
				t.Fatalf("error = %q, want %q", payload.Error, tt.wantError)
			}
		})
	}
}

func TestExamStatusRequiresParticipantContext(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("GET", "http://internal/api/exam/status", nil)
	rec := httptest.NewRecorder()

	h.Status(rec, req)

	if rec.Code != 401 {
		t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "unauthorized" {
		t.Fatalf("error = %q, want %q", payload.Error, "unauthorized")
	}
}

func TestExamSubmitAnswerMapsKnownServiceErrors(t *testing.T) {
	participant := db.GetParticipantByTokenRow{}
	questionID := "11111111-1111-1111-1111-111111111111"
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{
			name:       "already submitted",
			err:        service.ErrExamAlreadySubmit,
			wantStatus: 409,
			wantError:  "exam already submitted",
		},
		{
			name:       "window closed",
			err:        service.ErrExamWindowClosed,
			wantStatus: 403,
			wantError:  "exam window has closed",
		},
		{
			name:       "unexpected error",
			err:        errors.New("save failed"),
			wantStatus: 500,
			wantError:  "save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Exam{svc: &fakeExamService{answerErr: tt.err}}
			body := bytes.NewBufferString(`{"question_id":"` + questionID + `","answer":"B"}`)
			req := httptest.NewRequest("POST", "http://internal/api/exam/answer", body)
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
			rec := httptest.NewRecorder()

			h.SubmitAnswer(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json unmarshal failed: %v", err)
			}
			if payload.Error != tt.wantError {
				t.Fatalf("error = %q, want %q", payload.Error, tt.wantError)
			}
		})
	}
}

func TestExamSubmitMapsKnownServiceErrors(t *testing.T) {
	participant := db.GetParticipantByTokenRow{}
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{
			name:       "already submitted",
			err:        service.ErrExamAlreadySubmit,
			wantStatus: 409,
			wantError:  "exam already submitted",
		},
		{
			name:       "unexpected error",
			err:        errors.New("submit failed"),
			wantStatus: 500,
			wantError:  "submit failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Exam{svc: &fakeExamService{submitErr: tt.err}}
			req := httptest.NewRequest("POST", "http://internal/api/exam/submit", nil)
			req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
			rec := httptest.NewRecorder()

			h.Submit(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json unmarshal failed: %v", err)
			}
			if payload.Error != tt.wantError {
				t.Fatalf("error = %q, want %q", payload.Error, tt.wantError)
			}
		})
	}
}

func TestExamHeartbeatRequiresParticipantContext(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("POST", "http://internal/api/exam/heartbeat", nil)
	rec := httptest.NewRecorder()

	h.Heartbeat(rec, req)

	if rec.Code != 401 {
		t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "unauthorized" {
		t.Fatalf("error = %q, want %q", payload.Error, "unauthorized")
	}
}

func TestExamRecordEventRequiresParticipantContext(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	body := bytes.NewBufferString(`{"event_type":"warning","data":{"reason":"test"}}`)
	req := httptest.NewRequest("POST", "http://internal/api/exam/event", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RecordEvent(rec, req)

	if rec.Code != 401 {
		t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "unauthorized" {
		t.Fatalf("error = %q, want %q", payload.Error, "unauthorized")
	}
}
