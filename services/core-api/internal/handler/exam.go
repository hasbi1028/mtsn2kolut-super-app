package handler

import (
	"encoding/json"
	"net/http"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Exam struct {
	svc *service.Exam
}

func NewExam(svc *service.Exam) *Exam { return &Exam{svc: svc} }

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
