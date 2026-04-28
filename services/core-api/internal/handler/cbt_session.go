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
		EventID        string `json:"event_id"`
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

	var classID pgtype.UUID
	if body.ClassID != "" {
		classID, err = parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "class_id invalid")
			return
		}
	}

	var eventID pgtype.UUID
	if body.EventID != "" {
		eventID, err = parseUUID(body.EventID)
		if err != nil {
			api.BadRequest(w, "event_id invalid")
			return
		}
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
		EventID:        eventID,
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

// --- Participants ---

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

func (h *CbtSession) EnrollGrade(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	var body struct {
		Level string `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.Level == "" {
		api.BadRequest(w, "level required (e.g. VII, VIII, IX)")
		return
	}
	if err := h.svc.EnrollGrade(r.Context(), sessionID, body.Level); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "enrolled"})
}

func (h *CbtSession) EnrollSchool(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if err := h.svc.EnrollSchool(r.Context(), sessionID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "enrolled"})
}

func (h *CbtSession) GenerateTokens(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if err := h.svc.GenerateTokens(r.Context(), sessionID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "tokens generated"})
}

func (h *CbtSession) RegenerateToken(w http.ResponseWriter, r *http.Request) {
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	row, err := h.svc.RegenerateToken(r.Context(), pid)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

// --- Rooms ---

func (h *CbtSession) ListRooms(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	rows, err := h.svc.ListRooms(r.Context(), sessionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) CreateRoom(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	var body struct {
		RoomName string `json:"room_name"`
		Capacity int32  `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.RoomName == "" {
		api.BadRequest(w, "room_name required")
		return
	}
	if body.Capacity <= 0 {
		body.Capacity = 30
	}
	room, err := h.svc.CreateRoom(r.Context(), sessionID, body.RoomName, body.Capacity)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, room)
}

func (h *CbtSession) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := parseUUID(chi.URLParam(r, "rid"))
	if err != nil {
		api.BadRequest(w, "invalid room id")
		return
	}
	if err := h.svc.DeleteRoom(r.Context(), roomID); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *CbtSession) ShuffleRooms(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if err := h.svc.ShuffleRooms(r.Context(), sessionID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "shuffled"})
}

// --- Proctoring ---

func (h *CbtSession) GetProctoringStatus(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	rows, err := h.svc.GetProctoringStatus(r.Context(), sessionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) FlagParticipant(w http.ResponseWriter, r *http.Request) {
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	var body struct {
		Flag bool `json:"flag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if err := h.svc.SetSuspiciousFlag(r.Context(), pid, body.Flag); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]bool{"suspicious_flag": body.Flag})
}

// --- Essay Grading ---

func (h *CbtSession) ListUngradedEssays(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	rows, err := h.svc.ListUngradedEssays(r.Context(), sessionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) GradeEssay(w http.ResponseWriter, r *http.Request) {
	answerID, err := parseUUID(chi.URLParam(r, "aid"))
	if err != nil {
		api.BadRequest(w, "invalid answer id")
		return
	}
	var body struct {
		ManualScore float64 `json:"manual_score"`
		GradedBy    string  `json:"graded_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.ManualScore < 0 || body.ManualScore > 100 {
		api.BadRequest(w, "manual_score must be 0–100")
		return
	}
	if err := h.svc.GradeEssay(r.Context(), answerID, body.ManualScore, body.GradedBy); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"status": "graded", "manual_score": body.ManualScore})
}

// --- Answers & Scoring (existing) ---

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
