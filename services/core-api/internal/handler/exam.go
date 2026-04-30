package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Exam struct {
	svc examService
}

func NewExam(svc *service.Exam) *Exam { return &Exam{svc: svc} }

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
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.Token == "" {
		api.BadRequest(w, "token required")
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	result, err := h.svc.Login(r.Context(), body.Token, body.DeviceFingerprint, ip)
	if err != nil {
		switch err {
		case service.ErrExamNotFound:
			api.Err(w, http.StatusNotFound, "token not found")
		case service.ErrExamNotActive:
			api.Err(w, http.StatusForbidden, "exam session is not active")
		case service.ErrDeviceMismatch:
			api.Err(w, http.StatusConflict, "token already bound to another device")
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
	api.OK(w, result)
}

func (h *Exam) Heartbeat(w http.ResponseWriter, r *http.Request) {
	p, ok := mw.ParticipantFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
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
	var body struct {
		EventType string         `json:"event_type"`
		Data      map[string]any `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if err := h.svc.RecordClientEvent(r.Context(), p.ID, body.EventType, body.Data); err != nil {
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
	var body struct {
		QuestionID string `json:"question_id"`
		Answer     string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
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
	if err := h.svc.Submit(r.Context(), p); err != nil {
		switch err {
		case service.ErrExamAlreadySubmit:
			api.Err(w, http.StatusConflict, "exam already submitted")
		default:
			api.Internal(w, err)
		}
		return
	}
	api.OK(w, map[string]string{"status": "submitted"})
}

func absolutizeExamLoginResult(r *http.Request, result *service.LoginResult) {
	for i := range result.Questions {
		result.Questions[i].StemMediaURL = absolutizeExamAssetURL(r, result.Questions[i].StemMediaURL)
		result.Questions[i].StimulusMediaURL = absolutizeExamAssetURL(r, result.Questions[i].StimulusMediaURL)
		result.Questions[i].StemAudioURL = absolutizeExamAssetURL(r, result.Questions[i].StemAudioURL)
		result.Questions[i].StimulusAudioURL = absolutizeExamAssetURL(r, result.Questions[i].StimulusAudioURL)
	}
}

func absolutizeExamAssetURL(r *http.Request, value string) string {
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	if host == "" {
		return value
	}
	if strings.HasPrefix(value, "/") {
		return scheme + "://" + host + value
	}
	return scheme + "://" + host + "/" + value
}
