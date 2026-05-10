package handler

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeExamService struct {
	loginResult                 service.LoginResult
	loginErr                    error
	statusResult                service.StatusResult
	statusErr                   error
	heartbeatErr                error
	eventErr                    error
	answerErr                   error
	submitErr                   error
	commands                    []service.ParticipantCommand
	commandsErr                 error
	commandAckErr               error
	lastLoginToken              string
	lastLoginRoomToken          string
	lastLoginDeviceFingerprint  string
	lastLoginIP                 string
	lastStatusParticipantID     pgtype.UUID
	lastHeartbeatParticipantID  pgtype.UUID
	lastEventParticipantID      pgtype.UUID
	lastEventType               string
	lastEventData               map[string]any
	lastAnswerParticipantID     pgtype.UUID
	lastAnswerQuestionID        pgtype.UUID
	lastAnswerText              string
	lastSubmitParticipantID     pgtype.UUID
	lastCommandsParticipantID   pgtype.UUID
	lastCommandAckParticipantID pgtype.UUID
	lastCommandAckID            string
	lastCommandAckStatus        string
	submitCalls                 int
}

func (f *fakeExamService) Login(ctx context.Context, token, roomToken, deviceFingerprint, loginIP string) (service.LoginResult, error) {
	f.lastLoginToken = token
	f.lastLoginRoomToken = roomToken
	f.lastLoginDeviceFingerprint = deviceFingerprint
	f.lastLoginIP = loginIP
	return f.loginResult, f.loginErr
}

func (f *fakeExamService) GetStatus(ctx context.Context, p db.GetParticipantByTokenRow) (service.StatusResult, error) {
	f.lastStatusParticipantID = p.ID
	return f.statusResult, f.statusErr
}

func (f *fakeExamService) Heartbeat(ctx context.Context, participantID pgtype.UUID) error {
	f.lastHeartbeatParticipantID = participantID
	return f.heartbeatErr
}

func (f *fakeExamService) RecordClientEvent(ctx context.Context, participantID pgtype.UUID, eventType string, data map[string]any) error {
	f.lastEventParticipantID = participantID
	f.lastEventType = eventType
	f.lastEventData = data
	return f.eventErr
}

func (f *fakeExamService) SubmitAnswer(ctx context.Context, p db.GetParticipantByTokenRow, questionID pgtype.UUID, answer string) error {
	f.lastAnswerParticipantID = p.ID
	f.lastAnswerQuestionID = questionID
	f.lastAnswerText = answer
	return f.answerErr
}

func (f *fakeExamService) Submit(ctx context.Context, p db.GetParticipantByTokenRow) error {
	f.lastSubmitParticipantID = p.ID
	f.submitCalls++
	return f.submitErr
}

func (f *fakeExamService) ListPendingCommands(ctx context.Context, participantID pgtype.UUID) ([]service.ParticipantCommand, error) {
	f.lastCommandsParticipantID = participantID
	return f.commands, f.commandsErr
}

func (f *fakeExamService) AcknowledgeCommand(ctx context.Context, participantID pgtype.UUID, commandID, status string) error {
	f.lastCommandAckParticipantID = participantID
	f.lastCommandAckID = commandID
	f.lastCommandAckStatus = status
	return f.commandAckErr
}

func TestAbsolutizeExamAssetURLHonorsForwardedHeaders(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/24")
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", nil)
	req.Host = "internal:8080"
	req.RemoteAddr = "10.0.0.10:1234"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "cbt.mtsn2kolut.sch.id")

	got := absolutizeExamAssetURL(publicAPIBaseURL(req), "/api/cbt/assets/asset-1/file")
	want := "https://cbt.mtsn2kolut.sch.id/api/cbt/assets/asset-1/file"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamAssetURLFallsBackToRequestHost(t *testing.T) {
	req := httptest.NewRequest("POST", "http://localhost:8080/api/exam/login", nil)
	req.Host = "localhost:8080"

	got := absolutizeExamAssetURL(publicAPIBaseURL(req), "api/cbt/assets/asset-2/file")
	want := "http://localhost:8080/api/cbt/assets/asset-2/file"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamAssetURLLeavesAbsoluteURLUntouched(t *testing.T) {
	req := httptest.NewRequest("POST", "http://localhost:8080/api/exam/login", nil)

	got := absolutizeExamAssetURL(publicAPIBaseURL(req), "https://cdn.example.com/file.png")
	want := "https://cdn.example.com/file.png"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL() = %q, want %q", got, want)
	}
}

func TestAbsolutizeExamAssetURLCoversEdgeBranches(t *testing.T) {
	req := httptest.NewRequest("POST", "https://localhost:8443/api/exam/login", nil)
	req.TLS = &tls.ConnectionState{}
	got := absolutizeExamAssetURL(publicAPIBaseURL(req), "/api/cbt/assets/secure/file")
	want := "https://localhost:8443/api/cbt/assets/secure/file"
	if got != want {
		t.Fatalf("absolutizeExamAssetURL(https fallback) = %q, want %q", got, want)
	}

	req = &http.Request{Header: http.Header{}}
	if got := absolutizeExamAssetURL(publicAPIBaseURL(req), "api/cbt/assets/no-host/file"); got != "api/cbt/assets/no-host/file" {
		t.Fatalf("absolutizeExamAssetURL(no host) = %q, want original relative path", got)
	}

	req = httptest.NewRequest("POST", "http://localhost:8080/api/exam/login", nil)
	if got := absolutizeExamAssetURL(publicAPIBaseURL(req), ""); got != "" {
		t.Fatalf("absolutizeExamAssetURL(empty) = %q, want empty", got)
	}
	if got := absolutizeExamAssetURL(publicAPIBaseURL(req), "http://%zz"); got != "http://%zz" {
		t.Fatalf("absolutizeExamAssetURL(malformed absolute) = %q, want original malformed URL", got)
	}
	got = absolutizeExamAssetURL(publicAPIBaseURL(req), "https://cdn.example.com/file.png?existing=1")
	if got != "https://cdn.example.com/file.png?existing=1" {
		t.Fatalf("absolutizeExamAssetURL(existing query) = %q, want unchanged URL", got)
	}
}

func TestAbsolutizeExamLoginResultRewritesAllMediaFields(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/24")
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", nil)
	req.RemoteAddr = "10.0.0.10:1234"
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
	t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/24")
	svc := &fakeExamService{
		loginResult: service.LoginResult{
			ParticipantID: "participant-1",
			Student:       service.StudentInfo{NIS: "12345", Nama: "Ahmad"},
			Session:       service.SessionInfo{ID: "session-1", Title: "Ujian IPA"},
			Questions: []service.ExamQuestion{
				{
					ID:               "q-1",
					QuestionType:     "multiple_choice",
					QuestionText:     "Soal 1",
					Options:          json.RawMessage(`[{"label":"A","text":"Satu"},{"label":"B","text":"Dua"}]`),
					StemHTML:         "<p>Soal 1</p>",
					StimulusHTML:     "<p>Bacaan singkat</p>",
					StemMediaURL:     "/api/cbt/assets/image-1/file",
					StimulusMediaURL: "/api/cbt/assets/image-2/file",
					StemAudioURL:     "/api/cbt/assets/audio-2/file",
					StimulusAudioURL: "/api/cbt/assets/audio-1/file",
				},
			},
		},
	}
	h := &Exam{svc: svc}

	body := bytes.NewBufferString(`{"token":"a1b2c3d4","room_token":"ROOM-1","device_fingerprint":"device-1"}`)
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", body)
	req.RemoteAddr = "10.0.0.10:1234"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "cbt.mtsn2kolut.sch.id")
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
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
	if payload.Data.Questions[0].StimulusMediaURL != "https://cbt.mtsn2kolut.sch.id/api/cbt/assets/image-2/file" {
		t.Fatalf("StimulusMediaURL = %q", payload.Data.Questions[0].StimulusMediaURL)
	}
	if payload.Data.Questions[0].StemAudioURL != "https://cbt.mtsn2kolut.sch.id/api/cbt/assets/audio-2/file" {
		t.Fatalf("StemAudioURL = %q", payload.Data.Questions[0].StemAudioURL)
	}
	if payload.Data.Questions[0].StimulusAudioURL != "https://cbt.mtsn2kolut.sch.id/api/cbt/assets/audio-1/file" {
		t.Fatalf("StimulusAudioURL = %q", payload.Data.Questions[0].StimulusAudioURL)
	}
	if payload.Data.Questions[0].QuestionType != "multiple_choice" {
		t.Fatalf("QuestionType = %q", payload.Data.Questions[0].QuestionType)
	}
	if payload.Data.Questions[0].StemHTML != "<p>Soal 1</p>" {
		t.Fatalf("StemHTML = %q", payload.Data.Questions[0].StemHTML)
	}
	if payload.Data.Questions[0].StimulusHTML != "<p>Bacaan singkat</p>" {
		t.Fatalf("StimulusHTML = %q", payload.Data.Questions[0].StimulusHTML)
	}
	var options []map[string]string
	if err := json.Unmarshal(payload.Data.Questions[0].Options, &options); err != nil {
		t.Fatalf("question options unmarshal failed: %v", err)
	}
	if len(options) != 2 || options[1]["label"] != "B" || options[1]["text"] != "Dua" {
		t.Fatalf("question options = %+v, want mobile option label/text contract", options)
	}
	if svc.lastLoginToken != "a1b2c3d4" {
		t.Fatalf("login token = %q, want %q", svc.lastLoginToken, "a1b2c3d4")
	}
	if svc.lastLoginRoomToken != "ROOM-1" {
		t.Fatalf("room token = %q, want %q", svc.lastLoginRoomToken, "ROOM-1")
	}
	if svc.lastLoginDeviceFingerprint != "device-1" {
		t.Fatalf("device fingerprint = %q, want %q", svc.lastLoginDeviceFingerprint, "device-1")
	}
	if svc.lastLoginIP != "203.0.113.10" {
		t.Fatalf("login ip = %q, want %q", svc.lastLoginIP, "203.0.113.10")
	}
}

func TestExamTokenScopedEndpointsForwardParticipantContext(t *testing.T) {
	participant := db.GetParticipantByTokenRow{ID: handlerTestUUID(131)}
	svc := &fakeExamService{}
	h := &Exam{svc: svc}

	req := httptest.NewRequest("GET", "http://internal/api/exam/status", nil)
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	h.Status(httptest.NewRecorder(), req)

	req = httptest.NewRequest("GET", "http://internal/api/exam/commands", nil)
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	h.Commands(httptest.NewRecorder(), req)

	req = httptest.NewRequest("POST", "http://internal/api/exam/commands/cmd-1/ack", bytes.NewBufferString(`{"status":"read"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	h.AcknowledgeCommand(httptest.NewRecorder(), req)

	req = httptest.NewRequest("POST", "http://internal/api/exam/heartbeat", nil)
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	h.Heartbeat(httptest.NewRecorder(), req)

	req = httptest.NewRequest("POST", "http://internal/api/exam/event", bytes.NewBufferString(`{"event_type":"warning","data":{"reason":"manual_submit"}}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	h.RecordEvent(httptest.NewRecorder(), req)

	req = httptest.NewRequest("POST", "http://internal/api/exam/answer", bytes.NewBufferString(`{"question_id":"11111111-1111-1111-1111-111111111111","answer":"B"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	h.SubmitAnswer(httptest.NewRecorder(), req)

	req = httptest.NewRequest("POST", "http://internal/api/exam/submit", nil)
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	h.Submit(httptest.NewRecorder(), req)

	want := participant.ID.String()
	for endpoint, got := range map[string]string{
		"status":      svc.lastStatusParticipantID.String(),
		"heartbeat":   svc.lastHeartbeatParticipantID.String(),
		"event":       svc.lastEventParticipantID.String(),
		"answer":      svc.lastAnswerParticipantID.String(),
		"submit":      svc.lastSubmitParticipantID.String(),
		"commands":    svc.lastCommandsParticipantID.String(),
		"command_ack": svc.lastCommandAckParticipantID.String(),
	} {
		if got != want {
			t.Fatalf("%s participant id = %q, want %q", endpoint, got, want)
		}
	}
	if svc.lastCommandAckID != "cmd-1" || svc.lastCommandAckStatus != "read" {
		t.Fatalf("command ack = %q/%q, want cmd-1/read", svc.lastCommandAckID, svc.lastCommandAckStatus)
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
				Commands: []service.ParticipantCommand{{
					ID:      "cmd-1",
					Type:    "warning_message",
					Message: "Tetap di aplikasi ujian.",
				}},
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
	if len(payload.Data.Commands) != 1 || payload.Data.Commands[0].ID != "cmd-1" {
		t.Fatalf("Commands = %+v, want one pending command", payload.Data.Commands)
	}
}

func TestExamStatusMapsUnexpectedServiceError(t *testing.T) {
	h := &Exam{
		svc: &fakeExamService{
			statusErr: errors.New("status lookup failed"),
		},
	}

	req := httptest.NewRequest("GET", "http://internal/api/exam/status", nil)
	var participant db.GetParticipantByTokenRow
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.Status(rec, req)

	if rec.Code != 500 {
		t.Fatalf("status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "internal server error" {
		t.Fatalf("error = %q, want %q", payload.Error, "internal server error")
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
			name:       "session not started",
			err:        service.ErrExamNotStarted,
			wantStatus: 403,
			wantError:  "exam session has not started",
		},
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
			name:       "device mismatch",
			err:        service.ErrDeviceMismatch,
			wantStatus: 409,
			wantError:  "token already bound to another device",
		},
		{
			name:       "device required",
			err:        service.ErrDeviceRequired,
			wantStatus: 400,
			wantError:  "device fingerprint required",
		},
		{
			name:       "exam room required",
			err:        service.ErrExamRoomRequired,
			wantStatus: 403,
			wantError:  "exam room has not been assigned",
		},
		{
			name:       "room token required",
			err:        service.ErrRoomTokenRequired,
			wantStatus: 400,
			wantError:  "room token required",
		},
		{
			name:       "room token mismatch",
			err:        service.ErrRoomTokenMismatch,
			wantStatus: 403,
			wantError:  "room token mismatch",
		},
		{
			name:       "unexpected error",
			err:        errors.New("database down"),
			wantStatus: 500,
			wantError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Exam{svc: &fakeExamService{loginErr: tt.err}}
			body := bytes.NewBufferString(`{"token":"a1b2c3d4","room_token":"ROOM-1","device_fingerprint":"device-1"}`)
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

func TestExamLoginRejectsInvalidJSON(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", bytes.NewBufferString(`{"token":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "invalid json" {
		t.Fatalf("error = %q, want %q", payload.Error, "invalid json")
	}
}

func TestExamLoginRequiresToken(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("POST", "http://internal/api/exam/login", bytes.NewBufferString(`{"device_fingerprint":"device-1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "token required" {
		t.Fatalf("error = %q, want %q", payload.Error, "token required")
	}
}

func TestExamLoginHardensJSONBody(t *testing.T) {
	t.Run("rejects oversized body", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		body := bytes.NewBufferString(`{"token":"` + strings.Repeat("a", 5000) + `","device_fingerprint":"device-1"}`)
		req := httptest.NewRequest("POST", "http://internal/api/exam/login", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want 413; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastLoginToken != "" {
			t.Fatalf("service Login called with token %q, want blocked before service", svc.lastLoginToken)
		}
	})

	t.Run("rejects trailing json", func(t *testing.T) {
		h := &Exam{svc: &fakeExamService{}}
		req := httptest.NewRequest("POST", "http://internal/api/exam/login", bytes.NewBufferString(`{"token":"a1b2c3d4","room_token":"ROOM-1","device_fingerprint":"device-1"} {}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		var payload struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("json unmarshal failed: %v", err)
		}
		if payload.Error != "invalid json" {
			t.Fatalf("error = %q, want invalid json", payload.Error)
		}
	})

	t.Run("trims token before service lookup", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest("POST", "http://internal/api/exam/login", bytes.NewBufferString(`{"token":"  a1b2c3d4  ","room_token":"  ROOM-1  ","device_fingerprint":"device-1"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastLoginToken != "a1b2c3d4" {
			t.Fatalf("login token = %q, want trimmed token", svc.lastLoginToken)
		}
		if svc.lastLoginRoomToken != "  ROOM-1  " {
			t.Fatalf("room token = %q, want forwarded request value", svc.lastLoginRoomToken)
		}
	})
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

func TestExamSubmitAnswerRequiresParticipantContext(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("POST", "http://internal/api/exam/answer", bytes.NewBufferString(`{"question_id":"11111111-1111-1111-1111-111111111111","answer":"B"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.SubmitAnswer(rec, req)

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
			name:       "session not started",
			err:        service.ErrExamNotStarted,
			wantStatus: 403,
			wantError:  "exam session has not started",
		},
		{
			name:       "question outside exam",
			err:        service.ErrExamQuestionScope,
			wantStatus: 400,
			wantError:  "question is not part of this exam",
		},
		{
			name:       "unexpected error",
			err:        errors.New("save failed"),
			wantStatus: 500,
			wantError:  "internal server error",
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

func TestExamSubmitAnswerRejectsInvalidJSON(t *testing.T) {
	participant := db.GetParticipantByTokenRow{}
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("POST", "http://internal/api/exam/answer", bytes.NewBufferString(`{"question_id":`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.SubmitAnswer(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "invalid json" {
		t.Fatalf("error = %q, want %q", payload.Error, "invalid json")
	}
}

func TestExamSubmitAnswerRejectsInvalidQuestionID(t *testing.T) {
	participant := db.GetParticipantByTokenRow{}
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("POST", "http://internal/api/exam/answer", bytes.NewBufferString(`{"question_id":"not-a-uuid","answer":"B"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.SubmitAnswer(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "question_id invalid" {
		t.Fatalf("error = %q, want %q", payload.Error, "question_id invalid")
	}
}

func TestExamSubmitAnswerRejectsOversizedBody(t *testing.T) {
	participant := db.GetParticipantByTokenRow{}
	svc := &fakeExamService{}
	h := &Exam{svc: svc}
	req := httptest.NewRequest("POST", "http://internal/api/exam/answer", bytes.NewBufferString(`{"question_id":"11111111-1111-1111-1111-111111111111","answer":"`+strings.Repeat("x", 70<<10)+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.SubmitAnswer(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413; body=%s", rec.Code, rec.Body.String())
	}
	if svc.lastAnswerText != "" {
		t.Fatalf("service SubmitAnswer called with answer length %d, want blocked before service", len(svc.lastAnswerText))
	}
}

func TestExamSubmitAnswerWritesWrappedJSON(t *testing.T) {
	participant := db.GetParticipantByTokenRow{}
	svc := &fakeExamService{}
	h := &Exam{svc: svc}
	body := bytes.NewBufferString(`{"question_id":"11111111-1111-1111-1111-111111111111","answer":"B"}`)
	req := httptest.NewRequest("POST", "http://internal/api/exam/answer", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.SubmitAnswer(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Data["status"] != "recorded" {
		t.Fatalf("status = %q, want %q", payload.Data["status"], "recorded")
	}
	if svc.lastAnswerQuestionID.String() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("question id = %q, want %q", svc.lastAnswerQuestionID.String(), "11111111-1111-1111-1111-111111111111")
	}
	if svc.lastAnswerText != "B" {
		t.Fatalf("answer = %q, want %q", svc.lastAnswerText, "B")
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
			name:       "window closed",
			err:        service.ErrExamWindowClosed,
			wantStatus: 403,
			wantError:  "exam window has closed",
		},
		{
			name:       "session not started",
			err:        service.ErrExamNotStarted,
			wantStatus: 403,
			wantError:  "exam session has not started",
		},
		{
			name:       "unexpected error",
			err:        errors.New("submit failed"),
			wantStatus: 500,
			wantError:  "internal server error",
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

func TestExamSubmitWritesWrappedJSON(t *testing.T) {
	participant := db.GetParticipantByTokenRow{}
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("POST", "http://internal/api/exam/submit", bytes.NewBufferString(`{}`))
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.Submit(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Data["status"] != "submitted" {
		t.Fatalf("status = %q, want %q", payload.Data["status"], "submitted")
	}
}

func TestExamSubmitRejectsUnexpectedOrOversizedBody(t *testing.T) {
	participant := db.GetParticipantByTokenRow{ID: handlerTestUUID(132)}

	t.Run("rejects non empty json object", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest("POST", "http://internal/api/exam/submit", bytes.NewBufferString(`{"force":true}`))
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()

		h.Submit(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if svc.submitCalls != 0 {
			t.Fatalf("Submit service calls = %d, want blocked before service", svc.submitCalls)
		}
	})

	t.Run("rejects oversized body", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest("POST", "http://internal/api/exam/submit", bytes.NewBufferString(`{"padding":"`+strings.Repeat("x", 2<<10)+`"}`))
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()

		h.Submit(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want 413; body=%s", rec.Code, rec.Body.String())
		}
		if svc.submitCalls != 0 {
			t.Fatalf("Submit service calls = %d, want blocked before service", svc.submitCalls)
		}
	})
}

func TestExamSubmitUsesWriteLimiter(t *testing.T) {
	participant := db.GetParticipantByTokenRow{ID: handlerTestUUID(130)}
	fake := &fakeExamService{}
	h := &Exam{svc: fake, writeLimiter: newExamWriteLimiter()}

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "http://internal/api/exam/submit", nil)
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()
		h.Submit(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Submit attempt %d status = %d, want 200; body=%s", i+1, rec.Code, rec.Body.String())
		}
	}

	req := httptest.NewRequest("POST", "http://internal/api/exam/submit", nil)
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()
	h.Submit(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("Submit rate-limited status = %d, want 429; body=%s", rec.Code, rec.Body.String())
	}
	if fake.submitCalls != 5 {
		t.Fatalf("Submit service calls = %d, want 5 before limiter blocks", fake.submitCalls)
	}
}

func TestExamSubmitRequiresParticipantContext(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest("POST", "http://internal/api/exam/submit", nil)
	rec := httptest.NewRecorder()

	h.Submit(rec, req)

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

func TestExamHeartbeatWritesWrappedSuccessJSON(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	var participant db.GetParticipantByTokenRow
	req := httptest.NewRequest("POST", "http://internal/api/exam/heartbeat", bytes.NewBufferString(`{}`))
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.Heartbeat(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Data["status"] != "ok" {
		t.Fatalf("status = %q, want %q", payload.Data["status"], "ok")
	}
}

func TestExamHeartbeatRejectsUnexpectedOrOversizedBody(t *testing.T) {
	participant := db.GetParticipantByTokenRow{ID: handlerTestUUID(133)}

	t.Run("rejects non empty json object", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest("POST", "http://internal/api/exam/heartbeat", bytes.NewBufferString(`{"status":"manual"}`))
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()

		h.Heartbeat(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastHeartbeatParticipantID.Valid {
			t.Fatalf("Heartbeat service called with participant %s, want blocked before service", svc.lastHeartbeatParticipantID.String())
		}
	})

	t.Run("rejects json null", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest("POST", "http://internal/api/exam/heartbeat", bytes.NewBufferString(`null`))
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()

		h.Heartbeat(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastHeartbeatParticipantID.Valid {
			t.Fatalf("Heartbeat service called with participant %s, want blocked before service", svc.lastHeartbeatParticipantID.String())
		}
	})

	t.Run("rejects oversized body", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest("POST", "http://internal/api/exam/heartbeat", bytes.NewBufferString(`{"padding":"`+strings.Repeat("x", 2<<10)+`"}`))
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()

		h.Heartbeat(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want 413; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastHeartbeatParticipantID.Valid {
			t.Fatalf("Heartbeat service called with participant %s, want blocked before service", svc.lastHeartbeatParticipantID.String())
		}
	})
}

func TestExamHeartbeatMapsUnexpectedServiceError(t *testing.T) {
	h := &Exam{svc: &fakeExamService{heartbeatErr: errors.New("heartbeat failed")}}
	var participant db.GetParticipantByTokenRow
	req := httptest.NewRequest("POST", "http://internal/api/exam/heartbeat", nil)
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.Heartbeat(rec, req)

	if rec.Code != 500 {
		t.Fatalf("status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "internal server error" {
		t.Fatalf("error = %q, want %q", payload.Error, "internal server error")
	}
}

func TestExamRecordEventWritesWrappedSuccessJSON(t *testing.T) {
	svc := &fakeExamService{}
	h := &Exam{svc: svc}
	var participant db.GetParticipantByTokenRow
	body := bytes.NewBufferString(`{"event_type":"warning","data":{"reason":"test"}}`)
	req := httptest.NewRequest("POST", "http://internal/api/exam/event", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.RecordEvent(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Data["status"] != "recorded" {
		t.Fatalf("status = %q, want %q", payload.Data["status"], "recorded")
	}
	if svc.lastEventType != "warning" {
		t.Fatalf("event type = %q, want %q", svc.lastEventType, "warning")
	}
	if reason, _ := svc.lastEventData["reason"].(string); reason != "test" {
		t.Fatalf("event data reason = %v, want test", svc.lastEventData["reason"])
	}
}

func TestExamRecordEventHardensJSONBody(t *testing.T) {
	participant := db.GetParticipantByTokenRow{}

	t.Run("rejects oversized body", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest("POST", "http://internal/api/exam/event", bytes.NewBufferString(`{"event_type":"warning","data":{"padding":"`+strings.Repeat("x", 20<<10)+`"}}`))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()

		h.RecordEvent(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want 413; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastEventType != "" {
			t.Fatalf("service RecordClientEvent called with event type %q, want blocked before service", svc.lastEventType)
		}
	})

	t.Run("rejects trailing json", func(t *testing.T) {
		h := &Exam{svc: &fakeExamService{}}
		req := httptest.NewRequest("POST", "http://internal/api/exam/event", bytes.NewBufferString(`{"event_type":"warning","data":{"reason":"test"}} {}`))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()

		h.RecordEvent(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("trims event type before service write", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest("POST", "http://internal/api/exam/event", bytes.NewBufferString(`{"event_type":"  warning  ","data":{"reason":"test"}}`))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()

		h.RecordEvent(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastEventType != "warning" {
			t.Fatalf("event type = %q, want trimmed warning", svc.lastEventType)
		}
	})
}

func TestExamRecordEventMapsUnexpectedServiceError(t *testing.T) {
	h := &Exam{svc: &fakeExamService{eventErr: errors.New("event write failed")}}
	var participant db.GetParticipantByTokenRow
	body := bytes.NewBufferString(`{"event_type":"warning","data":{"reason":"test"}}`)
	req := httptest.NewRequest("POST", "http://internal/api/exam/event", body)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.RecordEvent(rec, req)

	if rec.Code != 500 {
		t.Fatalf("status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "internal server error" {
		t.Fatalf("error = %q, want %q", payload.Error, "internal server error")
	}
}

func TestExamRecordEventRejectsInvalidJSON(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	var participant db.GetParticipantByTokenRow
	req := httptest.NewRequest("POST", "http://internal/api/exam/event", bytes.NewBufferString(`{"event_type":`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.RecordEvent(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "invalid json" {
		t.Fatalf("error = %q, want %q", payload.Error, "invalid json")
	}
}

func TestExamRecordEventRequiresEventType(t *testing.T) {
	h := &Exam{svc: &fakeExamService{}}
	var participant db.GetParticipantByTokenRow
	req := httptest.NewRequest("POST", "http://internal/api/exam/event", bytes.NewBufferString(`{"data":{"reason":"test"}}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()

	h.RecordEvent(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Error != "event_type required" {
		t.Fatalf("error = %q, want %q", payload.Error, "event_type required")
	}
}
