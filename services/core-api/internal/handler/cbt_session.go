package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtSession struct {
	svc   cbtSessionService
	audit cbtSessionAuditWriter
}

type cbtSessionAuditWriter interface {
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

type cbtSessionService interface {
	CheckTeacherAccess(ctx context.Context, sessionID, teacherEmployeeID pgtype.UUID) (bool, error)
	List(ctx context.Context) ([]db.ListCbtExamSessionsRow, error)
	Get(ctx context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error)
	Create(ctx context.Context, in service.CreateCbtSessionInput) (db.CbtExamSession, error)
	UpdateStatus(ctx context.Context, id pgtype.UUID, status db.CbtSessionStatusEnum) (db.CbtExamSession, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	ListParticipants(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error)
	EnrollClass(ctx context.Context, sessionID, classID pgtype.UUID) error
	EnrollGrade(ctx context.Context, sessionID pgtype.UUID, level string) error
	EnrollSchool(ctx context.Context, sessionID pgtype.UUID) error
	GenerateTokens(ctx context.Context, sessionID pgtype.UUID) error
	RegenerateToken(ctx context.Context, participantID pgtype.UUID) (db.RegenerateParticipantTokenRow, error)
	HasParticipant(ctx context.Context, sessionID, participantID pgtype.UUID) (bool, error)
	HasRoom(ctx context.Context, sessionID, roomID pgtype.UUID) (bool, error)
	HasAnswer(ctx context.Context, sessionID, answerID pgtype.UUID) (bool, error)
	AssignSeat(ctx context.Context, participantID, roomID pgtype.UUID, seatNo int32) error
	AutoAssignSeats(ctx context.Context, sessionID pgtype.UUID) error
	ListRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error)
	CreateRoom(ctx context.Context, sessionID pgtype.UUID, roomName string, capacity int32) (db.CbtExamRoom, error)
	DeleteRoom(ctx context.Context, roomID pgtype.UUID) error
	ShuffleRooms(ctx context.Context, sessionID pgtype.UUID) error
	GetProctoringStatus(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error)
	SetSuspiciousFlag(ctx context.Context, participantID pgtype.UUID, flag bool) error
	ListUngradedEssays(ctx context.Context, sessionID pgtype.UUID) ([]db.ListUngradedEssaysRow, error)
	GradeEssay(ctx context.Context, answerID pgtype.UUID, manualScore float64, gradedBy string) error
	RecordAnswer(ctx context.Context, participantID, questionID pgtype.UUID, answer string) error
	ScoreSession(ctx context.Context, sessionID pgtype.UUID) error
	ListByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamSessionsByTeacherRow, error)
	GetResultsByTeacher(ctx context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.GetSessionResultsByTeacherRow, error)
	GetResults(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionResultsRow, error)
	GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error)
}

type cbtSessionProctorControlService interface {
	ResetParticipantRuntimeAccess(ctx context.Context, participantID pgtype.UUID, actor string) error
	ListParticipantEvents(ctx context.Context, sessionID, participantID pgtype.UUID, limit int32) ([]db.ListSessionParticipantEventsRow, error)
	ForceSubmitParticipant(ctx context.Context, sessionID, participantID pgtype.UUID, actor string) (db.ForceSubmitParticipantRow, error)
}

func NewCbtSession(svc *service.CbtSession, audit ...cbtSessionAuditWriter) *CbtSession {
	var writer cbtSessionAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &CbtSession{svc: svc, audit: writer}
}

func cbtSessionHasAnyRole(claims jwt.MapClaims, roles ...string) bool {
	return mw.HasAnyRole(claims, roles...)
}

func cbtSessionTeacherID(r *http.Request) pgtype.UUID {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok || cbtSessionHasAnyRole(claims, "admin") || !cbtSessionHasAnyRole(claims, "guru") {
		return pgtype.UUID{}
	}
	if eidRaw, _ := claims["eid"].(string); eidRaw != "" {
		var eid pgtype.UUID
		if err := eid.Scan(eidRaw); err == nil {
			return eid
		}
	}
	return pgtype.UUID{}
}

func (h *CbtSession) requireSessionTeacherOrAdmin(w http.ResponseWriter, r *http.Request, sessionID pgtype.UUID) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return false
	}
	if cbtSessionHasAnyRole(claims, "admin") {
		return true
	}
	teacherID := cbtSessionTeacherID(r)
	if !teacherID.Valid {
		api.Forbidden(w)
		return false
	}
	allowed, err := h.svc.CheckTeacherAccess(r.Context(), sessionID, teacherID)
	if err != nil {
		api.Internal(w, err)
		return false
	}
	if !allowed {
		api.Forbidden(w)
		return false
	}
	return true
}

func (h *CbtSession) requireSessionParticipant(w http.ResponseWriter, r *http.Request, sessionID, participantID pgtype.UUID) bool {
	ok, err := h.svc.HasParticipant(r.Context(), sessionID, participantID)
	if err != nil {
		api.Internal(w, err)
		return false
	}
	if !ok {
		api.Forbidden(w)
		return false
	}
	return true
}

func (h *CbtSession) requireSessionRoom(w http.ResponseWriter, r *http.Request, sessionID, roomID pgtype.UUID) bool {
	ok, err := h.svc.HasRoom(r.Context(), sessionID, roomID)
	if err != nil {
		api.Internal(w, err)
		return false
	}
	if !ok {
		api.Forbidden(w)
		return false
	}
	return true
}

func (h *CbtSession) requireSessionAnswer(w http.ResponseWriter, r *http.Request, sessionID, answerID pgtype.UUID) bool {
	ok, err := h.svc.HasAnswer(r.Context(), sessionID, answerID)
	if err != nil {
		api.Internal(w, err)
		return false
	}
	if !ok {
		api.Forbidden(w)
		return false
	}
	return true
}

func (h *CbtSession) List(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
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
	if !h.requireSessionTeacherOrAdmin(w, r, id) {
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		PackageID       string `json:"package_id"`
		ClassID         string `json:"class_id"`
		EventID         string `json:"event_id"`
		ScopeType       string `json:"scope_type"`
		ScopeRef        string `json:"scope_ref"`
		MixPolicy       string `json:"mix_policy"`
		AssignmentMode  string `json:"assignment_mode"`
		AllowCrossGrade bool   `json:"allow_cross_grade"`
		IsSpecialEvent  bool   `json:"is_special_event"`
		Title           string `json:"title"`
		ScheduledStart  string `json:"scheduled_start"`
		ScheduledEnd    string `json:"scheduled_end"`
		Status          string `json:"status"`
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

	scopeType := body.ScopeType
	if scopeType == "" {
		scopeType = "class"
	}

	var classID pgtype.UUID
	if body.ClassID != "" {
		classID, err = parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "class_id invalid")
			return
		}
	}
	if scopeType == "class" && !classID.Valid {
		api.BadRequest(w, "class_id wajib diisi untuk scope_type=class")
		return
	}
	if body.AllowCrossGrade && !body.IsSpecialEvent {
		api.BadRequest(w, "allow_cross_grade hanya boleh untuk special event")
		return
	}

	scopeRef := body.ScopeRef
	if scopeType == "class" && scopeRef == "" && classID.Valid {
		scopeRef = classID.String()
	}
	if (scopeType == "grade" || scopeType == "custom") && scopeRef == "" {
		api.BadRequest(w, "scope_ref wajib diisi untuk scope_type grade/custom")
		return
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
		PackageID:       packageID,
		ClassID:         classID,
		EventID:         eventID,
		ScopeType:       scopeType,
		ScopeRef:        scopeRef,
		MixPolicy:       body.MixPolicy,
		AssignmentMode:  body.AssignmentMode,
		AllowCrossGrade: body.AllowCrossGrade,
		IsSpecialEvent:  body.IsSpecialEvent,
		Title:           body.Title,
		ScheduledStart:  startTz,
		ScheduledEnd:    endTz,
		Status:          status,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *CbtSession) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
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
	if !h.requireSessionTeacherOrAdmin(w, r, id) {
		return
	}
	rows, err := h.svc.ListParticipants(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) Enroll(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	var body struct {
		ScopeType string `json:"scope_type"`
		ClassID   string `json:"class_id"`
		Level     string `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	switch body.ScopeType {
	case "", "class":
		classID, err := parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "class_id invalid")
			return
		}
		if err := h.svc.EnrollClass(r.Context(), sessionID, classID); err != nil {
			api.Internal(w, err)
			return
		}
	case "grade":
		if body.Level == "" {
			api.BadRequest(w, "level required (e.g. VII, VIII, IX)")
			return
		}
		if err := h.svc.EnrollGrade(r.Context(), sessionID, body.Level); err != nil {
			api.Internal(w, err)
			return
		}
	case "school":
		if err := h.svc.EnrollSchool(r.Context(), sessionID); err != nil {
			api.Internal(w, err)
			return
		}
	default:
		api.BadRequest(w, "scope_type harus class, grade, atau school")
		return
	}
	api.OK(w, map[string]string{"status": "enrolled"})
}

func (h *CbtSession) EnrollClass(w http.ResponseWriter, r *http.Request) {
	h.Enroll(w, r)
}

func (h *CbtSession) EnrollGrade(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	if err := h.svc.EnrollSchool(r.Context(), sessionID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "enrolled"})
}

func (h *CbtSession) GenerateTokens(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	if err := h.svc.GenerateTokens(r.Context(), sessionID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "tokens generated"})
}

func (h *CbtSession) RegenerateToken(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	if !h.requireSessionParticipant(w, r, sessionID, pid) {
		return
	}
	row, err := h.svc.RegenerateToken(r.Context(), pid)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtSession) ResetParticipantAccess(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	if !h.requireSessionParticipant(w, r, sessionID, pid) {
		return
	}
	proctorSvc, ok := h.svc.(cbtSessionProctorControlService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt proctor control service unavailable"))
		return
	}
	if err := proctorSvc.ResetParticipantRuntimeAccess(r.Context(), pid, currentUsername(r)); err != nil {
		api.Internal(w, err)
		return
	}
	h.auditEvent(r.Context(), "CBT_SESSION_PARTICIPANT_RESET_ACCESS", "cbt_session", pgUUIDString(sessionID), map[string]any{
		"participant_id":   pgUUIDString(pid),
		"actor_username":   currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, map[string]string{"status": "reset"})
}

func (h *CbtSession) AssignSeat(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	if !h.requireSessionParticipant(w, r, sessionID, pid) {
		return
	}
	var body struct {
		RoomID string `json:"room_id"`
		SeatNo int32  `json:"seat_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	roomID, err := parseUUID(body.RoomID)
	if err != nil {
		api.BadRequest(w, "room_id invalid")
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if body.SeatNo <= 0 {
		api.BadRequest(w, "seat_no harus lebih dari 0")
		return
	}
	if err := h.svc.AssignSeat(r.Context(), pid, roomID, body.SeatNo); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"status": "assigned", "seat_no": body.SeatNo})
}

func (h *CbtSession) AutoAssignSeats(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	if err := h.svc.AutoAssignSeats(r.Context(), sessionID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"status": "auto_assigned"})
}

// --- Rooms ---

func (h *CbtSession) ListRooms(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	roomID, err := parseUUID(chi.URLParam(r, "rid"))
	if err != nil {
		api.BadRequest(w, "invalid room id")
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if err := h.svc.DeleteRoom(r.Context(), roomID); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *CbtSession) ShuffleRooms(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
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
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	rows, err := h.svc.GetProctoringStatus(r.Context(), sessionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) ListParticipantEvents(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	var participantID pgtype.UUID
	if raw := r.URL.Query().Get("participant_id"); raw != "" {
		participantID, err = parseUUID(raw)
		if err != nil {
			api.BadRequest(w, "participant_id invalid")
			return
		}
		if !h.requireSessionParticipant(w, r, sessionID, participantID) {
			return
		}
	}
	limit := int32(100)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		var parsed int32
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err == nil {
			limit = parsed
		}
	}
	proctorSvc, ok := h.svc.(cbtSessionProctorControlService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt proctor control service unavailable"))
		return
	}
	rows, err := proctorSvc.ListParticipantEvents(r.Context(), sessionID, participantID, limit)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) FlagParticipant(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	if !h.requireSessionParticipant(w, r, sessionID, pid) {
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
	h.auditEvent(r.Context(), "CBT_SESSION_PARTICIPANT_FLAG", "cbt_session", pgUUIDString(sessionID), map[string]any{
		"participant_id":   pgUUIDString(pid),
		"suspicious_flag":  body.Flag,
		"actor_username":   currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, map[string]bool{"suspicious_flag": body.Flag})
}

func (h *CbtSession) ForceSubmitParticipant(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	if !h.requireSessionParticipant(w, r, sessionID, pid) {
		return
	}
	proctorSvc, ok := h.svc.(cbtSessionProctorControlService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt proctor control service unavailable"))
		return
	}
	row, err := proctorSvc.ForceSubmitParticipant(r.Context(), sessionID, pid, currentUsername(r))
	if err != nil {
		api.Internal(w, err)
		return
	}
	h.auditEvent(r.Context(), "CBT_SESSION_PARTICIPANT_FORCE_SUBMIT", "cbt_session", pgUUIDString(sessionID), map[string]any{
		"participant_id":   pgUUIDString(pid),
		"score":            row.Score,
		"actor_username":   currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, row)
}

// --- Essay Grading ---

func (h *CbtSession) ListUngradedEssays(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
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
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	answerID, err := parseUUID(chi.URLParam(r, "aid"))
	if err != nil {
		api.BadRequest(w, "invalid answer id")
		return
	}
	if !h.requireSessionAnswer(w, r, sessionID, answerID) {
		return
	}
	var body struct {
		ManualScore float64 `json:"manual_score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.ManualScore < 0 || body.ManualScore > 100 {
		api.BadRequest(w, "manual_score must be 0–100")
		return
	}
	if err := h.svc.GradeEssay(r.Context(), answerID, body.ManualScore, currentUsername(r)); err != nil {
		api.Internal(w, err)
		return
	}
	h.auditEvent(r.Context(), "CBT_SESSION_ESSAY_GRADE", "cbt_session", pgUUIDString(sessionID), map[string]any{
		"answer_id":        pgUUIDString(answerID),
		"manual_score":     body.ManualScore,
		"graded_by":        currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, map[string]any{"status": "graded", "manual_score": body.ManualScore})
}

// --- Answers & Scoring (existing) ---

func (h *CbtSession) RecordAnswer(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	participantID, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	if !h.requireSessionParticipant(w, r, sessionID, participantID) {
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
	if !h.requireSessionTeacherOrAdmin(w, r, id) {
		return
	}
	if err := h.svc.ScoreSession(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	h.auditEvent(r.Context(), "CBT_SESSION_SCORE", "cbt_session", pgUUIDString(id), map[string]any{
		"scored_by":        currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, map[string]string{"status": "scored"})
}

func (h *CbtSession) auditEvent(ctx context.Context, action, entityType, entityID string, extra map[string]any) {
	if h.audit == nil {
		return
	}
	claims, ok := api.ClaimsFromContext(ctx)
	if !ok {
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		return
	}

	meta := map[string]any{
		"username":   claims["usr"],
		"user_id":    claims["uid"],
		"session_id": claims["ssid"],
	}
	for key, value := range extra {
		meta[key] = value
	}
	rawMeta, err := json.Marshal(meta)
	if err != nil {
		return
	}
	_, _ = h.audit.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   rawMeta,
	})
}

func cbtAuditClaimString(ctx context.Context, key string) string {
	claims, ok := api.ClaimsFromContext(ctx)
	if !ok {
		return ""
	}
	value, _ := claims[key].(string)
	return value
}

func (h *CbtSession) GuruAwareList(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if ok && cbtSessionHasAnyRole(claims, "admin") {
		h.List(w, r)
		return
	}
	if ok && cbtSessionHasAnyRole(claims, "guru") {
		eidRaw, _ := claims["eid"].(string)
		if eidRaw == "" {
			api.Forbidden(w)
			return
		}
		var eid pgtype.UUID
		if err := eid.Scan(eidRaw); err != nil {
			api.Forbidden(w)
			return
		}
		rows, err := h.svc.ListByTeacher(r.Context(), eid)
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.OK(w, rows)
		return
	}
	api.Forbidden(w)
}

func (h *CbtSession) GuruAwareResults(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if ok && cbtSessionHasAnyRole(claims, "admin") {
		h.GetResults(w, r)
		return
	}
	if ok && cbtSessionHasAnyRole(claims, "guru") {
		eidRaw, _ := claims["eid"].(string)
		if eidRaw == "" {
			api.Forbidden(w)
			return
		}
		var eid pgtype.UUID
		if err := eid.Scan(eidRaw); err != nil {
			api.Forbidden(w)
			return
		}
		session, err := h.svc.Get(r.Context(), id)
		if err != nil {
			api.Internal(w, err)
			return
		}
		results, err := h.svc.GetResultsByTeacher(r.Context(), id, eid)
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.OK(w, map[string]any{
			"session": session,
			"results": results,
		})
		return
	}
	api.Forbidden(w)
}

func (h *CbtSession) GuruAwareParticipants(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if ok && cbtSessionHasAnyRole(claims, "admin") {
		h.ListParticipants(w, r)
		return
	}
	if ok && cbtSessionHasAnyRole(claims, "guru") {
		eidRaw, _ := claims["eid"].(string)
		if eidRaw == "" {
			api.Forbidden(w)
			return
		}
		var eid pgtype.UUID
		if err := eid.Scan(eidRaw); err != nil {
			api.Forbidden(w)
			return
		}
		hasAccess, err := h.svc.CheckTeacherAccess(r.Context(), id, eid)
		if err != nil {
			api.Internal(w, err)
			return
		}
		if !hasAccess {
			api.Forbidden(w)
			return
		}
	}
	if !ok || !cbtSessionHasAnyRole(claims, "admin", "guru") {
		api.Forbidden(w)
		return
	}
	h.ListParticipants(w, r)
}

func (h *CbtSession) GetResults(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
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

func (h *CbtSession) GetMinutes(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, id) {
		return
	}
	session, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	participants, err := h.svc.ListParticipants(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	rooms, err := h.svc.ListRooms(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"session":      session,
		"participants": participants,
		"rooms":        rooms,
	})
}

func (h *CbtSession) GetParticipantAnswers(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return
	}
	if !h.requireSessionParticipant(w, r, sessionID, pid) {
		return
	}
	rows, err := h.svc.GetParticipantAnswers(r.Context(), pid)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}
