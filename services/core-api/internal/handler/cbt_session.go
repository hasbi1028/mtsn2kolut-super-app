package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtSession struct {
	svc *service.CbtSession
}

func NewCbtSession(svc *service.CbtSession) *CbtSession { return &CbtSession{svc: svc} }

func (h *CbtSession) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtSession) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PackageID      string `json:"package_id"`
		ClassID        string `json:"class_id"`
		Title          string `json:"title"`
		ScheduledStart string `json:"scheduled_start"`
		ScheduledEnd   string `json:"scheduled_end"`
		Status         string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}

	packageID, err := parseUUID(body.PackageID)
	if err != nil {
		api.BadRequest(w, "package_id invalid")
		return
	}
	classID, err := parseUUID(body.ClassID)
	if err != nil {
		api.BadRequest(w, "class_id invalid")
		return
	}

	start, err := time.Parse(time.RFC3339, body.ScheduledStart)
	if err != nil {
		api.BadRequest(w, "scheduled_start invalid — use RFC3339")
		return
	}
	end, err := time.Parse(time.RFC3339, body.ScheduledEnd)
	if err != nil {
		api.BadRequest(w, "scheduled_end invalid — use RFC3339")
		return
	}
	if !end.After(start) {
		api.BadRequest(w, "scheduled_end must be after scheduled_start")
		return
	}

	status := db.CbtSessionStatusEnum(body.Status)
	if status == "" {
		status = db.CbtSessionStatusEnumDraft
	}

	var startTz, endTz pgtype.Timestamptz
	startTz.Time = start
	startTz.Valid = true
	endTz.Time = end
	endTz.Valid = true

	row, err := h.svc.Create(r.Context(), service.CreateCbtSessionInput{
		PackageID:      packageID,
		ClassID:        classID,
		Title:          body.Title,
		ScheduledStart: startTz,
		ScheduledEnd:   endTz,
		Status:         status,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *CbtSession) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.UpdateStatus(r.Context(), id, db.CbtSessionStatusEnum(body.Status))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtSession) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *CbtSession) ListParticipants(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.ListParticipants(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) EnrollClass(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	var body struct {
		ClassID string `json:"class_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	classID, err := parseUUID(body.ClassID)
	if err != nil {
		api.BadRequest(w, "class_id invalid")
		return
	}
	if err := h.svc.EnrollClass(r.Context(), sessionID, classID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "enrolled"})
}

func (h *CbtSession) RecordAnswer(w http.ResponseWriter, r *http.Request) {
	participantID, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
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
	if err := h.svc.RecordAnswer(r.Context(), participantID, questionID, body.Answer); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "recorded"})
}

func (h *CbtSession) ScoreSession(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.ScoreSession(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "scored"})
}

func (h *CbtSession) GetResults(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	session, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	results, err := h.svc.GetResults(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"session": session,
		"results": results,
	})
}

func (h *CbtSession) GetParticipantAnswers(w http.ResponseWriter, r *http.Request) {
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	rows, err := h.svc.GetParticipantAnswers(r.Context(), pid)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}
