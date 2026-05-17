package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func TestExamExtraCommandsHandlerBranches(t *testing.T) {
	participant := db.GetParticipantByTokenRow{ID: handlerTestUUID(210)}

	t.Run("unauthorized without participant", func(t *testing.T) {
		h := &Exam{svc: &fakeExamService{}}
		req := httptest.NewRequest(http.MethodGet, "/api/exam/commands", nil)
		rec := httptest.NewRecorder()
		h.Commands(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("Commands status = %d, want 401; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("wraps pending commands", func(t *testing.T) {
		svc := &fakeExamService{commands: []service.ParticipantCommand{{ID: "cmd-1", Type: "warning_message", Message: "Fokus ke ujian"}}}
		h := &Exam{svc: svc}
		req := httptest.NewRequest(http.MethodGet, "/api/exam/commands", nil)
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()
		h.Commands(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Commands status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastCommandsParticipantID != participant.ID || !strings.Contains(rec.Body.String(), "cmd-1") || !strings.Contains(rec.Body.String(), "commands") {
			t.Fatalf("Commands participant/body = %v/%s", svc.lastCommandsParticipantID, rec.Body.String())
		}
	})

	t.Run("maps service error", func(t *testing.T) {
		h := &Exam{svc: &fakeExamService{commandsErr: errors.New("commands failed")}}
		req := httptest.NewRequest(http.MethodGet, "/api/exam/commands", nil)
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()
		h.Commands(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Commands error status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestExamExtraAcknowledgeCommandBranches(t *testing.T) {
	participant := db.GetParticipantByTokenRow{ID: handlerTestUUID(211)}

	t.Run("unauthorized without participant", func(t *testing.T) {
		h := &Exam{svc: &fakeExamService{}}
		req := httptest.NewRequest(http.MethodPost, "/api/exam/commands/cmd-1/ack", nil)
		rec := httptest.NewRecorder()
		h.AcknowledgeCommand(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("Acknowledge status = %d, want 401; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("uses chi route param and body status", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest(http.MethodPost, "/api/exam/commands/ignored/ack", bytes.NewBufferString(`{"status":"acted"}`))
		req = withRouteParam(req, "cid", "cmd-route")
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()
		h.AcknowledgeCommand(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Acknowledge status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastCommandAckParticipantID != participant.ID || svc.lastCommandAckID != "cmd-route" || svc.lastCommandAckStatus != "acted" {
			t.Fatalf("ack args = %v/%q/%q", svc.lastCommandAckParticipantID, svc.lastCommandAckID, svc.lastCommandAckStatus)
		}
	})

	t.Run("falls back to path and default seen on empty body", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest(http.MethodPost, "/api/exam/commands/cmd-path/ack", nil)
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()
		h.AcknowledgeCommand(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Acknowledge fallback status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastCommandAckID != "cmd-path" || svc.lastCommandAckStatus != "seen" {
			t.Fatalf("ack fallback = %q/%q, want cmd-path/seen", svc.lastCommandAckID, svc.lastCommandAckStatus)
		}
	})

	t.Run("requires command id", func(t *testing.T) {
		h := &Exam{svc: &fakeExamService{}}
		req := httptest.NewRequest(http.MethodPost, "/api/exam/commands", nil)
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()
		h.AcknowledgeCommand(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Acknowledge missing id status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("rejects invalid optional json before service", func(t *testing.T) {
		svc := &fakeExamService{}
		h := &Exam{svc: svc}
		req := httptest.NewRequest(http.MethodPost, "/api/exam/commands/cmd-1/ack", bytes.NewBufferString(`{"status":"seen"} {}`))
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()
		h.AcknowledgeCommand(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Acknowledge invalid json status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}
		if svc.lastCommandAckID != "" {
			t.Fatalf("Acknowledge service called with id %q, want blocked", svc.lastCommandAckID)
		}
	})

	t.Run("maps service error", func(t *testing.T) {
		h := &Exam{svc: &fakeExamService{commandAckErr: errors.New("ack failed")}}
		req := httptest.NewRequest(http.MethodPost, "/api/exam/commands/cmd-1/ack", nil)
		req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
		rec := httptest.NewRecorder()
		h.AcknowledgeCommand(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("Acknowledge service error status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestExamExtraOptionalDecodeAndCommandPathHelpers(t *testing.T) {
	if got := examCommandIDFromPath("/api/exam/commands/cmd-123/ack"); got != "cmd-123" {
		t.Fatalf("examCommandIDFromPath = %q, want cmd-123", got)
	}
	if got := examCommandIDFromPath("/api/exam/no-commands"); got != "" {
		t.Fatalf("examCommandIDFromPath(no command) = %q, want empty", got)
	}

	t.Run("optional decode accepts empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		var body struct {
			Status string `json:"status"`
		}
		if !decodeOptionalExamJSON(rec, req, examEmptyBodyLimit, &body) || rec.Code != http.StatusOK {
			t.Fatalf("decodeOptionalExamJSON(empty) ok/code = false/%d, body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("optional decode rejects oversized body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"status":"`+strings.Repeat("x", 2<<10)+`"}`))
		rec := httptest.NewRecorder()
		var body struct {
			Status string `json:"status"`
		}
		if decodeOptionalExamJSON(rec, req, examEmptyBodyLimit, &body) || rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("decodeOptionalExamJSON(oversized) code = %d, want 413; body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestExamExtraActionLimiterPolicies(t *testing.T) {
	submitLimiter := newExamActionLimiter("submit")
	for i := 0; i < 5; i++ {
		if !submitLimiter.Allow() {
			t.Fatalf("submit limiter attempt %d denied, want initial burst allowed", i+1)
		}
	}
	if submitLimiter.Allow() {
		t.Fatalf("submit limiter sixth immediate attempt allowed, want denied")
	}

	defaultLimiter := newExamActionLimiter("answer")
	allowed := 0
	for i := 0; i < 61; i++ {
		if defaultLimiter.Allow() {
			allowed++
		}
	}
	if allowed != 60 {
		t.Fatalf("default limiter immediate allowed = %d, want burst 60", allowed)
	}

	limiter := newExamWriteLimiter()
	limiter.items["old"] = &examWriteLimiterItem{limiter: newExamActionLimiter("answer"), lastSeen: time.Now().Add(-11 * time.Minute)}
	limiter.lastCleanup = time.Now().Add(-2 * time.Minute)
	if !limiter.allow("new", "answer") {
		t.Fatalf("write limiter unexpectedly denied new key")
	}
	if _, ok := limiter.items["old"]; ok {
		t.Fatalf("write limiter did not cleanup stale key")
	}
}

func TestExamExtraAcknowledgeResponseJSON(t *testing.T) {
	participant := db.GetParticipantByTokenRow{ID: handlerTestUUID(212)}
	h := &Exam{svc: &fakeExamService{}}
	req := httptest.NewRequest(http.MethodPost, "/api/exam/commands/cmd-json/ack", bytes.NewBufferString(`{"status":"done"}`))
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()
	h.AcknowledgeCommand(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Data["status"] != "done" {
		t.Fatalf("ack response status = %q, want done", payload.Data["status"])
	}
}

func TestExamExtraRouteContextParamPrecedence(t *testing.T) {
	participant := db.GetParticipantByTokenRow{ID: handlerTestUUID(213)}
	svc := &fakeExamService{}
	h := &Exam{svc: svc}
	req := httptest.NewRequest(http.MethodPost, "/api/exam/commands/path-id/ack", nil)
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("cid", "chi-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	req = req.WithContext(context.WithValue(req.Context(), mw.ExamParticipantKey, participant))
	rec := httptest.NewRecorder()
	h.AcknowledgeCommand(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.lastCommandAckID != "chi-id" {
		t.Fatalf("ack command id = %q, want chi-id", svc.lastCommandAckID)
	}
}
