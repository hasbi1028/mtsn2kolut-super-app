package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/time/rate"
	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Exam struct {
	svc          examService
	writeLimiter *examWriteLimiter
}

func NewExam(svc *service.Exam) *Exam { return &Exam{svc: svc, writeLimiter: newExamWriteLimiter()} }

const (
	examLoginBodyLimit  int64 = 4 << 10
	examEventBodyLimit  int64 = 16 << 10
	examAnswerBodyLimit int64 = 64 << 10
	examEmptyBodyLimit  int64 = 1 << 10
)

type examService interface {
	Login(ctx context.Context, token, deviceFingerprint, loginIP string) (service.LoginResult, error)
	GetStatus(ctx context.Context, p db.GetParticipantByTokenRow) (service.StatusResult, error)
	Heartbeat(ctx context.Context, participantID pgtype.UUID) error
	RecordClientEvent(ctx context.Context, participantID pgtype.UUID, eventType string, data map[string]any) error
	SubmitAnswer(ctx context.Context, p db.GetParticipantByTokenRow, questionID pgtype.UUID, answer string) error
	Submit(ctx context.Context, p db.GetParticipantByTokenRow) error
}

func (h *Exam) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token             string `json:"token"`
		DeviceFingerprint string `json:"device_fingerprint"`
	}
	if !decodeExamJSON(w, r, examLoginBodyLimit, &body) {
		return
	}
	token := strings.TrimSpace(body.Token)
	if token == "" {
		api.BadRequest(w, "token required")
		return
	}

	ip := trustedClientIP(r)

	result, err := h.svc.Login(r.Context(), token, body.DeviceFingerprint, ip)
	if err != nil {
		switch err {
		case service.ErrExamNotFound:
			api.Err(w, http.StatusNotFound, "token not found")
		case service.ErrExamNotActive:
			api.Err(w, http.StatusForbidden, "exam session is not active")
		case service.ErrExamNotStarted:
			api.Err(w, http.StatusForbidden, "exam session has not started")
		case service.ErrExamAlreadySubmit:
			api.Err(w, http.StatusConflict, "exam already submitted")
		case service.ErrExamWindowClosed:
			api.Err(w, http.StatusForbidden, "exam window has closed")
		case service.ErrDeviceMismatch:
			api.Err(w, http.StatusConflict, "token already bound to another device")
		case service.ErrDeviceRequired:
			api.BadRequest(w, "device fingerprint required")
		default:
			api.Internal(w, err)
		}
		return
	}
	absolutizeExamLoginResult(r, &result)
	api.OK(w, result)
}

func (h *Exam) Status(w http.ResponseWriter, r *http.Request) {
	p, ok := mw.ParticipantFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	result, err := h.svc.GetStatus(r.Context(), p)
	if err != nil {
		api.Internal(w, err)
		return
	}
	// Status payload currently has no media URLs; keep this path explicit if fields are added.
	api.OK(w, result)
}

func (h *Exam) Heartbeat(w http.ResponseWriter, r *http.Request) {
	p, ok := mw.ParticipantFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	if !h.allowExamWrite(w, p.ID, "heartbeat") {
		return
	}
	if !validateOptionalEmptyExamJSONBody(w, r, examEmptyBodyLimit) {
		return
	}
	if err := h.svc.Heartbeat(r.Context(), p.ID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

func (h *Exam) RecordEvent(w http.ResponseWriter, r *http.Request) {
	p, ok := mw.ParticipantFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	if !h.allowExamWrite(w, p.ID, "event") {
		return
	}
	var body struct {
		EventType string         `json:"event_type"`
		Data      map[string]any `json:"data"`
	}
	if !decodeExamJSON(w, r, examEventBodyLimit, &body) {
		return
	}
	eventType := strings.TrimSpace(body.EventType)
	if eventType == "" {
		api.BadRequest(w, "event_type required")
		return
	}
	if err := h.svc.RecordClientEvent(r.Context(), p.ID, eventType, body.Data); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "recorded"})
}

func (h *Exam) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	p, ok := mw.ParticipantFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	if !h.allowExamWrite(w, p.ID, "answer") {
		return
	}
	var body struct {
		QuestionID string `json:"question_id"`
		Answer     string `json:"answer"`
	}
	if !decodeExamJSON(w, r, examAnswerBodyLimit, &body) {
		return
	}
	questionID, err := parseUUID(body.QuestionID)
	if err != nil {
		api.BadRequest(w, "question_id invalid")
		return
	}
	if err := h.svc.SubmitAnswer(r.Context(), p, questionID, body.Answer); err != nil {
		switch err {
		case service.ErrExamAlreadySubmit:
			api.Err(w, http.StatusConflict, "exam already submitted")
		case service.ErrExamWindowClosed:
			api.Err(w, http.StatusForbidden, "exam window has closed")
		case service.ErrExamNotStarted:
			api.Err(w, http.StatusForbidden, "exam session has not started")
		case service.ErrExamQuestionScope:
			api.Err(w, http.StatusBadRequest, "question is not part of this exam")
		default:
			api.Internal(w, err)
		}
		return
	}
	api.OK(w, map[string]string{"status": "recorded"})
}

func (h *Exam) Submit(w http.ResponseWriter, r *http.Request) {
	p, ok := mw.ParticipantFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	if !h.allowExamWrite(w, p.ID, "submit") {
		return
	}
	if !validateOptionalEmptyExamJSONBody(w, r, examEmptyBodyLimit) {
		return
	}
	if err := h.svc.Submit(r.Context(), p); err != nil {
		switch err {
		case service.ErrExamAlreadySubmit:
			api.Err(w, http.StatusConflict, "exam already submitted")
		case service.ErrExamWindowClosed:
			api.Err(w, http.StatusForbidden, "exam window has closed")
		case service.ErrExamNotStarted:
			api.Err(w, http.StatusForbidden, "exam session has not started")
		default:
			api.Internal(w, err)
		}
		return
	}
	api.OK(w, map[string]string{"status": "submitted"})
}

func decodeExamJSON(w http.ResponseWriter, r *http.Request, limit int64, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		writeExamDecodeError(w, err)
		return false
	}

	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		writeExamDecodeError(w, err)
		return false
	}
	return true
}

func validateOptionalEmptyExamJSONBody(w http.ResponseWriter, r *http.Request, limit int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	dec := json.NewDecoder(r.Body)
	var body map[string]json.RawMessage
	if err := dec.Decode(&body); err != nil {
		if errors.Is(err, io.EOF) {
			return true
		}
		writeExamDecodeError(w, err)
		return false
	}

	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		writeExamDecodeError(w, err)
		return false
	}
	if body == nil || len(body) > 0 {
		api.BadRequest(w, "request body must be empty")
		return false
	}
	return true
}

func writeExamDecodeError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		api.Err(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	api.BadRequest(w, "invalid json")
}

func absolutizeExamLoginResult(r *http.Request, result *service.LoginResult) {
	base := publicAPIBaseURL(r)
	for i := range result.Questions {
		result.Questions[i].StemMediaURL = absolutizeExamAssetURL(base, result.Questions[i].StemMediaURL)
		result.Questions[i].StimulusMediaURL = absolutizeExamAssetURL(base, result.Questions[i].StimulusMediaURL)
		result.Questions[i].StemAudioURL = absolutizeExamAssetURL(base, result.Questions[i].StemAudioURL)
		result.Questions[i].StimulusAudioURL = absolutizeExamAssetURL(base, result.Questions[i].StimulusAudioURL)
	}
}

func absolutizeExamAssetURL(base, value string) string {
	return joinBaseURL(base, value)
}

func (h *Exam) allowExamWrite(w http.ResponseWriter, participantID pgtype.UUID, action string) bool {
	if h.writeLimiter == nil {
		return true
	}
	if !h.writeLimiter.allow(pgUUIDString(participantID)+":"+action, action) {
		api.TooManyRequests(w)
		return false
	}
	return true
}

type examWriteLimiter struct {
	mu          sync.Mutex
	items       map[string]*examWriteLimiterItem
	lastCleanup time.Time
}

type examWriteLimiterItem struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newExamWriteLimiter() *examWriteLimiter {
	return &examWriteLimiter{items: map[string]*examWriteLimiterItem{}}
}

func (l *examWriteLimiter) allow(key, action string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.lastCleanup) >= time.Minute {
		for key, item := range l.items {
			if now.Sub(item.lastSeen) > 10*time.Minute {
				delete(l.items, key)
			}
		}
		l.lastCleanup = now
	}
	item, ok := l.items[key]
	if !ok {
		item = &examWriteLimiterItem{limiter: newExamActionLimiter(action)}
		l.items[key] = item
	}
	item.lastSeen = now
	return item.limiter.Allow()
}

func newExamActionLimiter(action string) *rate.Limiter {
	switch action {
	case "submit":
		return rate.NewLimiter(rate.Every(5*time.Second), 5)
	default:
		return rate.NewLimiter(rate.Limit(2), 60)
	}
}
