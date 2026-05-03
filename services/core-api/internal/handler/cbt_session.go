package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
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
	UpdateSchedule(ctx context.Context, id pgtype.UUID, start, end pgtype.Timestamptz) (service.UpdateCbtSessionScheduleResult, error)
	ListAuditLogs(ctx context.Context, id pgtype.UUID, limit, offset int32) ([]db.ListEntityAuditLogsRow, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	ListParticipants(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error)
	ListParticipantsByTeacher(ctx context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error)
	EnrollClass(ctx context.Context, sessionID, classID pgtype.UUID) error
	EnrollGrade(ctx context.Context, sessionID pgtype.UUID, level string) error
	EnrollSchool(ctx context.Context, sessionID pgtype.UUID) error
	GenerateTokens(ctx context.Context, sessionID pgtype.UUID) error
	RegenerateToken(ctx context.Context, participantID pgtype.UUID) (db.RegenerateParticipantTokenRow, error)
	HasParticipant(ctx context.Context, sessionID, participantID pgtype.UUID) (bool, error)
	HasParticipantByTeacher(ctx context.Context, sessionID, participantID, teacherEmployeeID pgtype.UUID) (bool, error)
	HasRoom(ctx context.Context, sessionID, roomID pgtype.UUID) (bool, error)
	HasAnswer(ctx context.Context, sessionID, answerID pgtype.UUID) (bool, error)
	HasAnswerByTeacher(ctx context.Context, sessionID, answerID, teacherEmployeeID pgtype.UUID) (bool, error)
	AssignSeat(ctx context.Context, participantID, roomID pgtype.UUID, seatNo int32) error
	AutoAssignSeats(ctx context.Context, sessionID pgtype.UUID) error
	ListRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error)
	CreateRoom(ctx context.Context, sessionID pgtype.UUID, roomName string, capacity int32) (db.CbtExamRoom, error)
	DeleteRoom(ctx context.Context, roomID pgtype.UUID) error
	ShuffleRooms(ctx context.Context, sessionID pgtype.UUID) error
	GetProctoringStatus(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error)
	SetSuspiciousFlag(ctx context.Context, participantID pgtype.UUID, flag bool) error
	ListUngradedEssays(ctx context.Context, sessionID pgtype.UUID) ([]db.ListUngradedEssaysRow, error)
	ListUngradedEssaysByTeacher(ctx context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.ListUngradedEssaysRow, error)
	GradeEssay(ctx context.Context, sessionID, answerID pgtype.UUID, manualScore float64, gradedBy string) error
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

type cbtSessionPhysicalRoomService interface {
	CreateRoomFromSchoolRoom(ctx context.Context, sessionID, schoolRoomID pgtype.UUID, roomName string, capacity int32) (db.CbtExamRoom, error)
}

type cbtSessionRoomProctorService interface {
	ListRoomProctors(ctx context.Context, roomID pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error)
	ReplaceRoomProctors(ctx context.Context, roomID, assignedBy, primaryEmployeeID pgtype.UUID, employeeIDs []pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error)
	RoomReadiness(ctx context.Context, sessionID pgtype.UUID) (db.GetCbtSessionRoomReadinessRow, error)
}

type cbtSessionRoomProctorDashboardService interface {
	GetRoomProctoringDashboard(ctx context.Context, roomID pgtype.UUID) (db.GetCbtRoomProctorDashboardRow, error)
	ListProctorRooms(ctx context.Context, employeeID pgtype.UUID, includeAll bool) ([]db.ListCbtProctorRoomsRow, error)
	HasRoomProctor(ctx context.Context, sessionID, roomID, employeeID pgtype.UUID) (bool, error)
	HasRoomParticipant(ctx context.Context, sessionID, roomID, participantID pgtype.UUID) (bool, error)
	GetProctoringStatusForRoom(ctx context.Context, sessionID, roomID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error)
	ListParticipantEventsForRoom(ctx context.Context, sessionID, participantID, roomID pgtype.UUID, limit int32) ([]db.ListSessionParticipantEventsRow, error)
}

type cbtSessionRoomHandoverService interface {
	GetRoomHandover(ctx context.Context, roomID pgtype.UUID) (db.GetCbtRoomHandoverRow, error)
	GetSessionOperationalRecap(ctx context.Context, sessionID pgtype.UUID) (db.GetCbtSessionOperationalRecapRow, []db.ListCbtSessionRoomOperationalRecapRow, error)
	SaveRoomHandover(ctx context.Context, roomID, updatedBy pgtype.UUID, in service.SaveCbtRoomHandoverInput) (db.CbtRoomHandover, error)
	LockRoomHandover(ctx context.Context, roomID, lockedBy pgtype.UUID) (db.CbtRoomHandover, error)
}

type cbtSessionItemAnalysisService interface {
	GetItemAnalysis(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionItemAnalysisRow, error)
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

func cbtSessionEmployeeID(r *http.Request) pgtype.UUID {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return pgtype.UUID{}
	}
	eidRaw, _ := claims["eid"].(string)
	if strings.TrimSpace(eidRaw) == "" {
		return pgtype.UUID{}
	}
	var eid pgtype.UUID
	if err := eid.Scan(strings.TrimSpace(eidRaw)); err != nil {
		return pgtype.UUID{}
	}
	return eid
}

func cbtSessionActorUserID(r *http.Request) pgtype.UUID {
	var uid pgtype.UUID
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if raw, ok := claims["uid"].(string); ok {
			_ = uid.Scan(raw)
		}
	}
	return uid
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

func (h *CbtSession) requireSessionParticipantForTeacherOrAdmin(w http.ResponseWriter, r *http.Request, sessionID, participantID pgtype.UUID) bool {
	if adminAccessAllowed(r) {
		return h.requireSessionParticipant(w, r, sessionID, participantID)
	}
	teacherID := cbtSessionTeacherID(r)
	if !teacherID.Valid {
		api.Forbidden(w)
		return false
	}
	ok, err := h.svc.HasParticipantByTeacher(r.Context(), sessionID, participantID, teacherID)
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

func (h *CbtSession) requireSessionAnswerForTeacherOrAdmin(w http.ResponseWriter, r *http.Request, sessionID, answerID pgtype.UUID) bool {
	if adminAccessAllowed(r) {
		return h.requireSessionAnswer(w, r, sessionID, answerID)
	}
	teacherID := cbtSessionTeacherID(r)
	if !teacherID.Valid {
		api.Forbidden(w)
		return false
	}
	ok, err := h.svc.HasAnswerByTeacher(r.Context(), sessionID, answerID, teacherID)
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

func (h *CbtSession) requireSessionRoomProctorOrAdmin(w http.ResponseWriter, r *http.Request, sessionID, roomID pgtype.UUID) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return false
	}
	if cbtSessionHasAnyRole(claims, "admin") {
		return true
	}
	employeeID := cbtSessionEmployeeID(r)
	if !employeeID.Valid {
		api.Forbidden(w)
		return false
	}
	roomSvc, ok := h.svc.(cbtSessionRoomProctorDashboardService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor dashboard service unavailable"))
		return false
	}
	allowed, err := roomSvc.HasRoomProctor(r.Context(), sessionID, roomID, employeeID)
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

func (h *CbtSession) requireSessionRoomParticipant(w http.ResponseWriter, r *http.Request, sessionID, roomID, participantID pgtype.UUID) bool {
	roomSvc, ok := h.svc.(cbtSessionRoomProctorDashboardService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor dashboard service unavailable"))
		return false
	}
	allowed, err := roomSvc.HasRoomParticipant(r.Context(), sessionID, roomID, participantID)
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
		api.BadRequest(w, "scheduled_start invalid - use RFC3339")
		return
	}
	end, err := time.Parse(time.RFC3339, body.ScheduledEnd)
	if err != nil {
		api.BadRequest(w, "scheduled_end invalid - use RFC3339")
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
		if errors.Is(err, domain.ErrConflict) {
			message := strings.TrimPrefix(safeClientMessage(err, "Sesi CBT tidak valid"), "conflict: ")
			api.Conflict(w, message)
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			api.NotFound(w)
			return
		}
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
		if errors.Is(err, domain.ErrConflict) {
			message := strings.TrimPrefix(safeClientMessage(err, "Status sesi tidak dapat diperbarui"), "conflict: ")
			api.Conflict(w, message)
			return
		}
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtSession) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
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
		ScheduledStart string `json:"scheduled_start"`
		ScheduledEnd   string `json:"scheduled_end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	start, err := time.Parse(time.RFC3339, body.ScheduledStart)
	if err != nil {
		api.BadRequest(w, "scheduled_start invalid - use RFC3339")
		return
	}
	end, err := time.Parse(time.RFC3339, body.ScheduledEnd)
	if err != nil {
		api.BadRequest(w, "scheduled_end invalid - use RFC3339")
		return
	}
	if !end.After(start) {
		api.BadRequest(w, "scheduled_end must be after scheduled_start")
		return
	}
	result, err := h.svc.UpdateSchedule(
		r.Context(),
		id,
		pgtype.Timestamptz{Time: start, Valid: true},
		pgtype.Timestamptz{Time: end, Valid: true},
	)
	if err != nil {
		if errors.Is(err, domain.ErrBadRequest) {
			message := strings.TrimPrefix(safeClientMessage(err, "Jadwal sesi tidak valid"), "bad request: ")
			api.BadRequest(w, message)
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			message := strings.TrimPrefix(safeClientMessage(err, "Jadwal sesi tidak dapat diperbarui"), "conflict: ")
			api.Conflict(w, message)
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			api.NotFound(w)
			return
		}
		api.Internal(w, err)
		return
	}
	h.auditEvent(r.Context(), "CBT_SESSION_SCHEDULE_UPDATE", "cbt_session", pgUUIDString(id), map[string]any{
		"title":                    result.Session.Title,
		"status":                   result.Session.Status,
		"previous_scheduled_start": timestamptzRFC3339(result.Before.ScheduledStart),
		"previous_scheduled_end":   timestamptzRFC3339(result.Before.ScheduledEnd),
		"new_scheduled_start":      timestamptzRFC3339(result.Session.ScheduledStart),
		"new_scheduled_end":        timestamptzRFC3339(result.Session.ScheduledEnd),
		"actor_session_id":         cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user":         cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, result.Session)
}

func (h *CbtSession) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, id) {
		return
	}
	q := r.URL.Query()
	limit := int32(pageSize(q.Get("per_page"), 50))
	page := pageNum(q.Get("page"), 1)
	offset := int32((page - 1) * int(limit))

	rows, err := h.svc.ListAuditLogs(r.Context(), id, limit, offset)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, serializeCbtSessionAuditLogs(rows))
}

type cbtSessionAuditLogResponse struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	Username   string          `json:"username"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Metadata   json.RawMessage `json:"metadata"`
	CreatedAt  string          `json:"created_at"`
}

func serializeCbtSessionAuditLogs(rows []db.ListEntityAuditLogsRow) []cbtSessionAuditLogResponse {
	items := make([]cbtSessionAuditLogResponse, 0, len(rows))
	for _, row := range rows {
		metadata := json.RawMessage(`{}`)
		if len(row.Metadata) > 0 && json.Valid(row.Metadata) {
			metadata = json.RawMessage(row.Metadata)
		}
		items = append(items, cbtSessionAuditLogResponse{
			ID:         pgUUIDString(row.ID),
			UserID:     pgUUIDString(row.UserID),
			Username:   row.Username.String,
			Action:     row.Action,
			EntityType: row.EntityType,
			EntityID:   row.EntityID,
			Metadata:   metadata,
			CreatedAt:  timestamptzRFC3339(row.CreatedAt),
		})
	}
	return items
}

func timestamptzRFC3339(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
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
	api.OK(w, serializeParticipantListRows(rows, adminAccessAllowed(r)))
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
		if errors.Is(err, domain.ErrConflict) {
			message := strings.TrimPrefix(safeClientMessage(err, "Token sesi tidak dapat diperbarui"), "conflict: ")
			api.Conflict(w, message)
			return
		}
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
		if errors.Is(err, domain.ErrConflict) {
			message := strings.TrimPrefix(safeClientMessage(err, "Token peserta tidak dapat diperbarui"), "conflict: ")
			api.Conflict(w, message)
			return
		}
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
		RoomName     string `json:"room_name"`
		SchoolRoomID string `json:"school_room_id"`
		Capacity     int32  `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	body.RoomName = strings.TrimSpace(body.RoomName)
	if strings.TrimSpace(body.SchoolRoomID) != "" {
		schoolRoomID, err := parseUUID(body.SchoolRoomID)
		if err != nil {
			api.BadRequest(w, "school_room_id invalid")
			return
		}
		physicalRoomSvc, ok := h.svc.(cbtSessionPhysicalRoomService)
		if !ok {
			api.Internal(w, fmt.Errorf("cbt physical room service unavailable"))
			return
		}
		room, err := physicalRoomSvc.CreateRoomFromSchoolRoom(r.Context(), sessionID, schoolRoomID, body.RoomName, body.Capacity)
		if err != nil {
			api.Internal(w, err)
			return
		}
		h.auditEvent(r.Context(), "CBT_SESSION_ROOM_LINK_CREATE", "cbt_session_room", pgUUIDString(room.ID), map[string]any{
			"session_id":     pgUUIDString(sessionID),
			"school_room_id": pgUUIDString(schoolRoomID),
			"room_name":      room.RoomName,
			"capacity":       room.Capacity,
		})
		api.Created(w, room)
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

func (h *CbtSession) GetRoomReadiness(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	roomSvc, ok := h.svc.(cbtSessionRoomProctorService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room readiness service unavailable"))
		return
	}
	readiness, err := roomSvc.RoomReadiness(r.Context(), sessionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, readiness)
}

func (h *CbtSession) ListRoomProctors(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, ok := h.requireSessionRoomParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	roomSvc, ok := h.svc.(cbtSessionRoomProctorService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor service unavailable"))
		return
	}
	rows, err := roomSvc.ListRoomProctors(r.Context(), roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) ReplaceRoomProctors(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	sessionID, roomID, ok := h.requireSessionRoomParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	var body struct {
		PrimaryEmployeeID string   `json:"primary_employee_id"`
		EmployeeIDs       []string `json:"employee_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	var primaryEmployeeID pgtype.UUID
	if strings.TrimSpace(body.PrimaryEmployeeID) != "" {
		parsed, err := parseUUID(body.PrimaryEmployeeID)
		if err != nil {
			api.BadRequest(w, "primary_employee_id invalid")
			return
		}
		primaryEmployeeID = parsed
	}
	employeeIDs := make([]pgtype.UUID, 0, len(body.EmployeeIDs))
	for _, raw := range body.EmployeeIDs {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		parsed, err := parseUUID(raw)
		if err != nil {
			api.BadRequest(w, "employee_ids invalid")
			return
		}
		employeeIDs = append(employeeIDs, parsed)
	}
	roomSvc, ok := h.svc.(cbtSessionRoomProctorService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor service unavailable"))
		return
	}
	rows, err := roomSvc.ReplaceRoomProctors(r.Context(), roomID, cbtSessionActorUserID(r), primaryEmployeeID, employeeIDs)
	if err != nil {
		api.Internal(w, err)
		return
	}
	h.auditEvent(r.Context(), "CBT_SESSION_ROOM_PROCTORS_REPLACE", "cbt_session_room", pgUUIDString(roomID), map[string]any{
		"session_id":          pgUUIDString(sessionID),
		"primary_employee_id": body.PrimaryEmployeeID,
		"employee_ids":        body.EmployeeIDs,
		"actor_username":      currentUsername(r),
	})
	api.OK(w, rows)
}

func (h *CbtSession) requireSessionRoomParams(w http.ResponseWriter, r *http.Request) (pgtype.UUID, pgtype.UUID, bool) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	roomID, err := parseUUID(chi.URLParam(r, "rid"))
	if err != nil {
		api.BadRequest(w, "invalid room id")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	return sessionID, roomID, true
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.GetProctoringStatus(r.Context(), sessionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, serializeProctoringRows(rows, adminAccessAllowed(r)))
}

func (h *CbtSession) ListParticipantEvents(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
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
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
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

func (h *CbtSession) GetRoomProctoringDashboard(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, ok := h.requireSessionRoomParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomProctorOrAdmin(w, r, sessionID, roomID) {
		return
	}
	roomSvc, ok := h.svc.(cbtSessionRoomProctorDashboardService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor dashboard service unavailable"))
		return
	}
	proctorSvc, ok := h.svc.(cbtSessionRoomProctorService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor service unavailable"))
		return
	}
	room, err := roomSvc.GetRoomProctoringDashboard(r.Context(), roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	proctors, err := proctorSvc.ListRoomProctors(r.Context(), roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	participants, err := roomSvc.GetProctoringStatusForRoom(r.Context(), sessionID, roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	events, err := roomSvc.ListParticipantEventsForRoom(r.Context(), sessionID, pgtype.UUID{}, roomID, 100)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"room":         room,
		"proctors":     proctors,
		"participants": serializeProctoringRows(participants, adminAccessAllowed(r)),
		"events":       events,
	})
}

func (h *CbtSession) ListMyProctorRooms(w http.ResponseWriter, r *http.Request) {
	if !cbtOpsAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	includeAll := cbtSessionHasAnyRole(claims, "admin")
	employeeID := cbtSessionEmployeeID(r)
	if !includeAll && !employeeID.Valid {
		api.Forbidden(w)
		return
	}
	roomSvc, ok := h.svc.(cbtSessionRoomProctorDashboardService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor dashboard service unavailable"))
		return
	}
	rows, err := roomSvc.ListProctorRooms(r.Context(), employeeID, includeAll)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtSession) GetRoomProctorPrintPack(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, ok := h.requireSessionRoomParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomProctorOrAdmin(w, r, sessionID, roomID) {
		return
	}
	roomSvc, ok := h.svc.(cbtSessionRoomProctorDashboardService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor dashboard service unavailable"))
		return
	}
	proctorSvc, ok := h.svc.(cbtSessionRoomProctorService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room proctor service unavailable"))
		return
	}
	room, err := roomSvc.GetRoomProctoringDashboard(r.Context(), roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	proctors, err := proctorSvc.ListRoomProctors(r.Context(), roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	participants, err := roomSvc.GetProctoringStatusForRoom(r.Context(), sessionID, roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"room":         room,
		"proctors":     proctors,
		"participants": participants,
	})
}

func (h *CbtSession) GetRoomHandover(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, ok := h.requireSessionRoomParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomProctorOrAdmin(w, r, sessionID, roomID) {
		return
	}
	handoverSvc, ok := h.svc.(cbtSessionRoomHandoverService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room handover service unavailable"))
		return
	}
	row, err := handoverSvc.GetRoomHandover(r.Context(), roomID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtSession) GetSessionOperationalRecap(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	if !h.requireSessionTeacherOrAdmin(w, r, sessionID) {
		return
	}
	handoverSvc, ok := h.svc.(cbtSessionRoomHandoverService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt session operational recap service unavailable"))
		return
	}
	recap, rooms, err := handoverSvc.GetSessionOperationalRecap(r.Context(), sessionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"recap": recap,
		"rooms": rooms,
	})
}

func (h *CbtSession) SaveRoomHandover(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, ok := h.requireSessionRoomParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomProctorOrAdmin(w, r, sessionID, roomID) {
		return
	}
	var body struct {
		AttendanceChecked     bool   `json:"attendance_checked"`
		AllSubmittedChecked   bool   `json:"all_submitted_checked"`
		DeviceIssueChecked    bool   `json:"device_issue_checked"`
		RoomCleanChecked      bool   `json:"room_clean_checked"`
		TokenReturnedChecked  bool   `json:"token_returned_checked"`
		AssetsReturnedChecked bool   `json:"assets_returned_checked"`
		IncidentNotes         string `json:"incident_notes"`
		OperatorNotes         string `json:"operator_notes"`
		HandoverNotes         string `json:"handover_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	handoverSvc, ok := h.svc.(cbtSessionRoomHandoverService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room handover service unavailable"))
		return
	}
	row, err := handoverSvc.SaveRoomHandover(r.Context(), roomID, cbtSessionActorUserID(r), service.SaveCbtRoomHandoverInput{
		AttendanceChecked:     body.AttendanceChecked,
		AllSubmittedChecked:   body.AllSubmittedChecked,
		DeviceIssueChecked:    body.DeviceIssueChecked,
		RoomCleanChecked:      body.RoomCleanChecked,
		TokenReturnedChecked:  body.TokenReturnedChecked,
		AssetsReturnedChecked: body.AssetsReturnedChecked,
		IncidentNotes:         body.IncidentNotes,
		OperatorNotes:         body.OperatorNotes,
		HandoverNotes:         body.HandoverNotes,
	})
	if err != nil {
		writeClientError(w, err, "Serah terima ruang tidak valid")
		return
	}
	h.auditEvent(r.Context(), "CBT_SESSION_ROOM_HANDOVER_SAVE", "cbt_session_room", pgUUIDString(roomID), map[string]any{
		"session_id":              pgUUIDString(sessionID),
		"attendance_checked":      body.AttendanceChecked,
		"all_submitted_checked":   body.AllSubmittedChecked,
		"device_issue_checked":    body.DeviceIssueChecked,
		"room_clean_checked":      body.RoomCleanChecked,
		"token_returned_checked":  body.TokenReturnedChecked,
		"assets_returned_checked": body.AssetsReturnedChecked,
		"actor_username":          currentUsername(r),
		"actor_session_id":        cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user":        cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, row)
}

func (h *CbtSession) LockRoomHandover(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, ok := h.requireSessionRoomParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomProctorOrAdmin(w, r, sessionID, roomID) {
		return
	}
	handoverSvc, ok := h.svc.(cbtSessionRoomHandoverService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt room handover service unavailable"))
		return
	}
	row, err := handoverSvc.LockRoomHandover(r.Context(), roomID, cbtSessionActorUserID(r))
	if err != nil {
		writeClientError(w, err, "Penguncian serah terima ruang tidak valid")
		return
	}
	h.auditEvent(r.Context(), "CBT_SESSION_ROOM_HANDOVER_LOCK", "cbt_session_room", pgUUIDString(roomID), map[string]any{
		"session_id":       pgUUIDString(sessionID),
		"actor_username":   currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, row)
}

func (h *CbtSession) FlagRoomParticipant(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, pid, ok := h.requireRoomParticipantControlParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomProctorOrAdmin(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomParticipant(w, r, sessionID, roomID, pid) {
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
	h.auditEvent(r.Context(), "CBT_SESSION_ROOM_PARTICIPANT_FLAG", "cbt_session_room", pgUUIDString(roomID), map[string]any{
		"session_id":       pgUUIDString(sessionID),
		"participant_id":   pgUUIDString(pid),
		"suspicious_flag":  body.Flag,
		"actor_username":   currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, map[string]bool{"suspicious_flag": body.Flag})
}

func (h *CbtSession) ResetRoomParticipantAccess(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, pid, ok := h.requireRoomParticipantControlParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomProctorOrAdmin(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomParticipant(w, r, sessionID, roomID, pid) {
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
	h.auditEvent(r.Context(), "CBT_SESSION_ROOM_PARTICIPANT_RESET_ACCESS", "cbt_session_room", pgUUIDString(roomID), map[string]any{
		"session_id":       pgUUIDString(sessionID),
		"participant_id":   pgUUIDString(pid),
		"actor_username":   currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, map[string]string{"status": "reset"})
}

func (h *CbtSession) ForceSubmitRoomParticipant(w http.ResponseWriter, r *http.Request) {
	sessionID, roomID, pid, ok := h.requireRoomParticipantControlParams(w, r)
	if !ok {
		return
	}
	if !h.requireSessionRoom(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomProctorOrAdmin(w, r, sessionID, roomID) {
		return
	}
	if !h.requireSessionRoomParticipant(w, r, sessionID, roomID, pid) {
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
	h.auditEvent(r.Context(), "CBT_SESSION_ROOM_PARTICIPANT_FORCE_SUBMIT", "cbt_session_room", pgUUIDString(roomID), map[string]any{
		"session_id":       pgUUIDString(sessionID),
		"participant_id":   pgUUIDString(pid),
		"score":            row.Score,
		"actor_username":   currentUsername(r),
		"actor_session_id": cbtAuditClaimString(r.Context(), "ssid"),
		"actor_claim_user": cbtAuditClaimString(r.Context(), "uid"),
	})
	api.OK(w, row)
}

func (h *CbtSession) requireRoomParticipantControlParams(w http.ResponseWriter, r *http.Request) (pgtype.UUID, pgtype.UUID, pgtype.UUID, bool) {
	sessionID, roomID, ok := h.requireSessionRoomParams(w, r)
	if !ok {
		return pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, false
	}
	pid, err := parseUUID(chi.URLParam(r, "pid"))
	if err != nil {
		api.BadRequest(w, "invalid participant id")
		return pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, false
	}
	return sessionID, roomID, pid, true
}

// --- Essay Grading ---

func (h *CbtSession) ListUngradedEssays(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}
	var rows []db.ListUngradedEssaysRow
	if adminAccessAllowed(r) {
		rows, err = h.svc.ListUngradedEssays(r.Context(), sessionID)
	} else {
		teacherID := cbtSessionTeacherID(r)
		if !teacherID.Valid {
			api.Forbidden(w)
			return
		}
		hasAccess, accessErr := h.svc.CheckTeacherAccess(r.Context(), sessionID, teacherID)
		if accessErr != nil {
			api.Internal(w, accessErr)
			return
		}
		if !hasAccess {
			api.Forbidden(w)
			return
		}
		rows, err = h.svc.ListUngradedEssaysByTeacher(r.Context(), sessionID, teacherID)
	}
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
	answerID, err := parseUUID(chi.URLParam(r, "aid"))
	if err != nil {
		api.BadRequest(w, "invalid answer id")
		return
	}
	if !h.requireSessionAnswerForTeacherOrAdmin(w, r, sessionID, answerID) {
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
	if err := h.svc.GradeEssay(r.Context(), sessionID, answerID, body.ManualScore, currentUsername(r)); err != nil {
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
		if errors.Is(err, service.ErrExamQuestionScope) {
			api.BadRequest(w, "question is not part of this exam")
			return
		}
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
		hasAccess, err := h.svc.CheckTeacherAccess(r.Context(), id, eid)
		if err != nil {
			api.Internal(w, err)
			return
		}
		if !hasAccess {
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
		rows, err := h.svc.ListParticipantsByTeacher(r.Context(), id, eid)
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.OK(w, serializeParticipantListRows(rows, false))
		return
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

func (h *CbtSession) GetItemAnalysis(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	itemAnalysisSvc, ok := h.svc.(cbtSessionItemAnalysisService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt item analysis service unavailable"))
		return
	}
	rows, err := itemAnalysisSvc.GetItemAnalysis(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	includeAnswerKey := adminAccessAllowed(r)
	for _, row := range rows {
		items = append(items, serializeSessionItemAnalysisRow(row, includeAnswerKey))
	}
	api.OK(w, map[string]any{"items": items})
}

func serializeSessionItemAnalysisRow(row db.GetSessionItemAnalysisRow, includeAnswerKey bool) map[string]any {
	answerKey := ""
	if includeAnswerKey {
		answerKey = row.AnswerKey
	}
	return map[string]any{
		"position":             row.Position,
		"points":               row.Points,
		"question_id":          pgUUIDString(row.QuestionID),
		"question_code":        row.QuestionCode,
		"question_text":        row.QuestionText,
		"question_type":        row.QuestionType,
		"difficulty":           row.Difficulty,
		"answer_key":           answerKey,
		"cp_ref":               row.CpRef,
		"tp_ref":               row.TpRef,
		"kd_ref":               row.KdRef,
		"material_topic":       row.MaterialTopic,
		"cognitive_level":      row.CognitiveLevel,
		"hots_flag":            row.HotsFlag,
		"submitted_count":      row.SubmittedCount,
		"answered_count":       row.AnsweredCount,
		"blank_count":          row.BlankCount,
		"correct_count":        row.CorrectCount,
		"incorrect_count":      row.IncorrectCount,
		"unscored_count":       row.UnscoredCount,
		"avg_manual_score":     row.AvgManualScore,
		"difficulty_index":     row.DifficultyIndex,
		"top_group_count":      row.TopGroupCount,
		"top_correct_count":    row.TopCorrectCount,
		"bottom_group_count":   row.BottomGroupCount,
		"bottom_correct_count": row.BottomCorrectCount,
		"discrimination_index": row.DiscriminationIndex,
		"answer_distribution":  decodeJSONBytes(row.AnswerDistribution),
		"recommendation":       sessionItemAnalysisRecommendation(row),
		"recommendation_tone":  sessionItemAnalysisTone(row),
	}
}

func sessionItemAnalysisRecommendation(row db.GetSessionItemAnalysisRow) string {
	switch {
	case row.SubmittedCount == 0:
		return "Belum ada submit"
	case row.QuestionType == "essay" && row.UnscoredCount > 0:
		return "Koreksi uraian belum lengkap"
	case row.AnsweredCount == 0:
		return "Belum dijawab"
	case row.DiscriminationIndex < -0.05:
		return "Cek kunci/rubrik"
	case row.DifficultyIndex < 0.20:
		return "Terlalu sulit"
	case row.DifficultyIndex > 0.90:
		return "Terlalu mudah"
	case row.SubmittedCount >= 3 && row.DiscriminationIndex < 0.15:
		return "Daya pembeda rendah"
	case row.BlankCount > row.SubmittedCount/2:
		return "Banyak jawaban kosong"
	default:
		return "Baik"
	}
}

func sessionItemAnalysisTone(row db.GetSessionItemAnalysisRow) string {
	switch sessionItemAnalysisRecommendation(row) {
	case "Baik":
		return "success"
	case "Belum ada submit":
		return "info"
	case "Cek kunci/rubrik", "Belum dijawab":
		return "danger"
	default:
		return "warning"
	}
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
	var participants []db.ListCbtExamParticipantsRow
	if adminAccessAllowed(r) {
		participants, err = h.svc.ListParticipants(r.Context(), id)
	} else {
		teacherID := cbtSessionTeacherID(r)
		if !teacherID.Valid {
			api.Forbidden(w)
			return
		}
		participants, err = h.svc.ListParticipantsByTeacher(r.Context(), id, teacherID)
	}
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
		"participants": serializeParticipantListRows(participants, adminAccessAllowed(r)),
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
	if !h.requireSessionParticipantForTeacherOrAdmin(w, r, sessionID, pid) {
		return
	}
	rows, err := h.svc.GetParticipantAnswers(r.Context(), pid)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, serializeParticipantAnswerRows(rows, adminAccessAllowed(r)))
}

func serializeProctoringRows(rows []db.GetSessionProctoringStatusRow, includeToken bool) []map[string]any {
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		token := ""
		if includeToken {
			token = row.Token
		}
		items = append(items, map[string]any{
			"participant_id":     pgUUIDString(row.ParticipantID),
			"student_id":         pgUUIDString(row.StudentID),
			"nis":                row.Nis,
			"nama":               row.Nama,
			"token":              token,
			"room_id":            pgUUIDString(row.RoomID),
			"room_name":          row.RoomName,
			"seat_no":            row.SeatNo,
			"submitted_at":       row.SubmittedAt,
			"last_heartbeat":     row.LastHeartbeat,
			"app_switch_count":   row.AppSwitchCount,
			"screenshot_attempt": row.ScreenshotAttempt,
			"suspicious_flag":    row.SuspiciousFlag,
			"answered_count":     row.AnsweredCount,
			"score":              row.Score,
		})
	}
	return items
}

func serializeParticipantListRows(rows []db.ListCbtExamParticipantsRow, includeToken bool) []map[string]any {
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		token := ""
		if includeToken {
			token = row.Token
		}
		items = append(items, map[string]any{
			"id":                 pgUUIDString(row.ID),
			"session_id":         pgUUIDString(row.SessionID),
			"student_id":         pgUUIDString(row.StudentID),
			"nis":                row.Nis,
			"nama":               row.Nama,
			"gender":             row.Gender,
			"token":              token,
			"room_id":            pgUUIDString(row.RoomID),
			"seat_no":            row.SeatNo,
			"joined_at":          row.JoinedAt,
			"submitted_at":       row.SubmittedAt,
			"score":              row.Score,
			"app_switch_count":   row.AppSwitchCount,
			"screenshot_attempt": row.ScreenshotAttempt,
			"suspicious_flag":    row.SuspiciousFlag,
			"last_heartbeat":     row.LastHeartbeat,
			"created_at":         row.CreatedAt,
			"room_name":          row.RoomName,
		})
	}
	return items
}

func serializeParticipantAnswerRows(rows []db.GetParticipantAnswersRow, includeAnswerKey bool) []map[string]any {
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		answerKey := ""
		if includeAnswerKey {
			answerKey = row.AnswerKey
		}
		items = append(items, map[string]any{
			"id":             pgUUIDString(row.ID),
			"participant_id": pgUUIDString(row.ParticipantID),
			"question_id":    pgUUIDString(row.QuestionID),
			"question_code":  row.QuestionCode,
			"question_text":  row.QuestionText,
			"option_a":       row.OptionA,
			"option_b":       row.OptionB,
			"option_c":       row.OptionC,
			"option_d":       row.OptionD,
			"option_e":       row.OptionE,
			"answer_key":     answerKey,
			"answer":         row.Answer,
			"is_correct":     row.IsCorrect,
			"answered_at":    row.AnsweredAt,
		})
	}
	return items
}
