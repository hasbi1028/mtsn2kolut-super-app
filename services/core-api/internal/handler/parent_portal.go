package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type parentPortalService interface {
	Children(ctx context.Context, userID pgtype.UUID) ([]db.ListParentChildrenRow, error)
	ChildrenByParentID(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error)
	ChildProfile(ctx context.Context, userID, studentID pgtype.UUID) (db.GetParentPortalChildProfileRow, error)
	ChildProfileByParentID(ctx context.Context, parentID, studentID pgtype.UUID) (db.GetParentPortalChildProfileRow, error)
	ChildSchedule(ctx context.Context, userID, studentID pgtype.UUID) ([]db.ListParentPortalChildTimetableRow, error)
	ChildScheduleByParentID(ctx context.Context, parentID, studentID pgtype.UUID) ([]db.ListParentPortalChildTimetableRow, error)
	ChildResults(ctx context.Context, userID, studentID pgtype.UUID) ([]db.ListParentPortalChildExamSessionsRow, error)
	ChildResultsByParentID(ctx context.Context, parentID, studentID pgtype.UUID) ([]db.ListParentPortalChildExamSessionsRow, error)
}

type ParentPortal struct {
	svc parentPortalService
}

func NewParentPortal(svc *service.ParentPortal) *ParentPortal {
	return &ParentPortal{svc: svc}
}

func (h *ParentPortal) PreviewChildren(w http.ResponseWriter, r *http.Request) {
	parentID, ok := h.previewParentID(w, r)
	if !ok {
		return
	}
	children, err := h.svc.ChildrenByParentID(r.Context(), parentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Data anak tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"children": parentPortalChildren(children), "preview": true})
}

func (h *ParentPortal) PreviewChildProfile(w http.ResponseWriter, r *http.Request) {
	parentID, studentID, ok := h.previewChildRequest(w, r)
	if !ok {
		return
	}
	student, err := h.svc.ChildProfileByParentID(r.Context(), parentID, studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Data profil anak tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"student": parentPortalChildProfile(student), "preview": true})
}

func (h *ParentPortal) PreviewChildSchedule(w http.ResponseWriter, r *http.Request) {
	parentID, studentID, ok := h.previewChildRequest(w, r)
	if !ok {
		return
	}
	schedule, err := h.svc.ChildScheduleByParentID(r.Context(), parentID, studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Jadwal anak tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"schedule": schedule, "preview": true})
}

func (h *ParentPortal) PreviewChildResults(w http.ResponseWriter, r *http.Request) {
	parentID, studentID, ok := h.previewChildRequest(w, r)
	if !ok {
		return
	}
	results, err := h.svc.ChildResultsByParentID(r.Context(), parentID, studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Hasil anak tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"results": parentPortalResults(results), "preview": true})
}

func (h *ParentPortal) Children(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authorizedUserID(w, r)
	if !ok {
		return
	}
	children, err := h.svc.Children(r.Context(), userID)
	if err != nil {
		writeDomainOrInternal(w, err, "Data anak tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"children": parentPortalChildren(children)})
}

func (h *ParentPortal) ChildProfile(w http.ResponseWriter, r *http.Request) {
	userID, studentID, ok := h.authorizedChildRequest(w, r)
	if !ok {
		return
	}
	student, err := h.svc.ChildProfile(r.Context(), userID, studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Data profil anak tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"student": parentPortalChildProfile(student)})
}

func (h *ParentPortal) ChildSchedule(w http.ResponseWriter, r *http.Request) {
	userID, studentID, ok := h.authorizedChildRequest(w, r)
	if !ok {
		return
	}
	schedule, err := h.svc.ChildSchedule(r.Context(), userID, studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Jadwal anak tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"schedule": schedule})
}

func (h *ParentPortal) ChildResults(w http.ResponseWriter, r *http.Request) {
	userID, studentID, ok := h.authorizedChildRequest(w, r)
	if !ok {
		return
	}
	results, err := h.svc.ChildResults(r.Context(), userID, studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Hasil anak tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"results": parentPortalResults(results)})
}

func (h *ParentPortal) authorizedUserID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	userID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	claims, _ := api.ClaimsFromContext(r.Context())
	if !parentPortalAccessAllowed(claims) {
		api.Forbidden(w)
		return pgtype.UUID{}, false
	}
	return userID, true
}

func (h *ParentPortal) authorizedChildRequest(w http.ResponseWriter, r *http.Request) (pgtype.UUID, pgtype.UUID, bool) {
	userID, ok := h.authorizedUserID(w, r)
	if !ok {
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	studentID, err := parseUUID(chi.URLParam(r, "studentID"))
	if err != nil {
		api.BadRequest(w, "invalid student_id")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	return userID, studentID, true
}

func (h *ParentPortal) previewParentID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	parentID, err := parseUUID(chi.URLParam(r, "parentID"))
	if err != nil {
		api.BadRequest(w, "ID orang tua tidak valid")
		return pgtype.UUID{}, false
	}
	return parentID, true
}

func (h *ParentPortal) previewChildRequest(w http.ResponseWriter, r *http.Request) (pgtype.UUID, pgtype.UUID, bool) {
	parentID, ok := h.previewParentID(w, r)
	if !ok {
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	studentID, err := parseUUID(chi.URLParam(r, "studentID"))
	if err != nil {
		api.BadRequest(w, "invalid student_id")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	return parentID, studentID, true
}

func parentPortalAccessAllowed(claims jwt.MapClaims) bool {
	return mw.HasAnyRole(claims, "ortu") || mw.HasAnyPermission(claims, "parent_portal.read")
}

type parentPortalChildDTO struct {
	ID               string `json:"id"`
	NIS              string `json:"nis"`
	Nama             string `json:"nama"`
	ClassID          string `json:"class_id"`
	ClassName        string `json:"class_name"`
	Relationship     string `json:"relationship"`
	IsPrimaryContact bool   `json:"is_primary_contact"`
	Notes            string `json:"notes"`
}

func parentPortalChildren(rows []db.ListParentChildrenRow) []parentPortalChildDTO {
	out := make([]parentPortalChildDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, parentPortalChildDTO{
			ID:               portalUUIDString(row.ID),
			NIS:              row.Nis,
			Nama:             row.Nama,
			ClassID:          portalUUIDString(row.ClassID),
			ClassName:        portalTextString(row.ClassName),
			Relationship:     parentRelationshipString(row.Relationship),
			IsPrimaryContact: row.IsPrimaryContact,
			Notes:            row.Notes,
		})
	}
	return out
}

type parentPortalChildProfileDTO struct {
	ID          string `json:"id"`
	NIS         string `json:"nis"`
	NISN        string `json:"nisn"`
	Nama        string `json:"nama"`
	Gender      string `json:"gender"`
	ParentName  string `json:"parent_name"`
	ParentPhone string `json:"parent_phone"`
	ClassID     string `json:"class_id"`
	ClassName   string `json:"class_name"`
	ClassCode   string `json:"class_code"`
	IsActive    bool   `json:"is_active"`
	Status      string `json:"status"`
}

func parentPortalChildProfile(row db.GetParentPortalChildProfileRow) parentPortalChildProfileDTO {
	return parentPortalChildProfileDTO{
		ID:          portalUUIDString(row.ID),
		NIS:         row.Nis,
		NISN:        row.Nisn,
		Nama:        row.Nama,
		Gender:      string(row.Gender),
		ParentName:  row.ParentName,
		ParentPhone: row.ParentPhone,
		ClassID:     portalUUIDString(row.ClassID),
		ClassName:   portalTextString(row.ClassName),
		ClassCode:   portalTextString(row.ClassCode),
		IsActive:    row.IsActive,
		Status:      string(row.Status),
	}
}

type parentPortalResultDTO struct {
	ParticipantID   string             `json:"participant_id"`
	SessionID       string             `json:"session_id"`
	RoomID          string             `json:"room_id"`
	SeatNo          pgtype.Int4        `json:"seat_no"`
	JoinedAt        pgtype.Timestamptz `json:"joined_at"`
	SubmittedAt     pgtype.Timestamptz `json:"submitted_at"`
	Score           pgtype.Numeric     `json:"score"`
	SessionTitle    string             `json:"session_title"`
	SessionStatus   string             `json:"session_status"`
	ScheduledStart  pgtype.Timestamptz `json:"scheduled_start"`
	ScheduledEnd    pgtype.Timestamptz `json:"scheduled_end"`
	PackageTitle    string             `json:"package_title"`
	DurationMinutes int32              `json:"duration_minutes"`
	RoomName        string             `json:"room_name"`
}

func parentPortalResults(rows []db.ListParentPortalChildExamSessionsRow) []parentPortalResultDTO {
	out := make([]parentPortalResultDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, parentPortalResultDTO{
			ParticipantID:   portalUUIDString(row.ParticipantID),
			SessionID:       portalUUIDString(row.SessionID),
			RoomID:          portalUUIDString(row.RoomID),
			SeatNo:          row.SeatNo,
			JoinedAt:        row.JoinedAt,
			SubmittedAt:     row.SubmittedAt,
			Score:           row.Score,
			SessionTitle:    row.SessionTitle,
			SessionStatus:   string(row.SessionStatus),
			ScheduledStart:  row.ScheduledStart,
			ScheduledEnd:    row.ScheduledEnd,
			PackageTitle:    row.PackageTitle,
			DurationMinutes: row.DurationMinutes,
			RoomName:        row.RoomName,
		})
	}
	return out
}

func parentRelationshipString(value any) string {
	if value == nil {
		return ""
	}
	if raw, ok := value.(string); ok {
		return raw
	}
	return ""
}
