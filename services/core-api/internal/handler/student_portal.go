package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type studentPortalService interface {
	PreviewStudents(ctx context.Context) ([]db.ListStudentPortalPreviewStudentsRow, error)
	Profile(ctx context.Context, userID pgtype.UUID) (db.GetStudentByIDRow, error)
	ProfileByStudentID(ctx context.Context, studentID pgtype.UUID) (db.GetStudentByIDRow, error)
	Schedule(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentTimetableRow, error)
	ScheduleByStudentID(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error)
	Results(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error)
	ResultsByStudentID(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error)
	CbtSchedule(ctx context.Context, userID pgtype.UUID) ([]service.StudentPortalCbtScheduleItem, error)
	CbtScheduleByStudentID(ctx context.Context, studentID pgtype.UUID) ([]service.StudentPortalCbtScheduleItem, error)
	RevealCbtToken(ctx context.Context, userID, participantID pgtype.UUID, roomToken, clientIP string) (service.StudentPortalTokenReveal, error)
}

type StudentPortal struct {
	svc studentPortalService
}

func NewStudentPortal(svc *service.StudentPortal) *StudentPortal {
	return &StudentPortal{svc: svc}
}

func (h *StudentPortal) PreviewStudents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.PreviewStudents(r.Context())
	if err != nil {
		writeDomainOrInternal(w, err, "Daftar siswa preview tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"students": studentPortalPreviewStudents(rows)})
}

func (h *StudentPortal) PreviewProfile(w http.ResponseWriter, r *http.Request) {
	studentID, ok := h.previewStudentID(w, r)
	if !ok {
		return
	}
	student, err := h.svc.ProfileByStudentID(r.Context(), studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Data profil siswa tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"student": studentPortalProfile(student), "preview": true})
}

func (h *StudentPortal) PreviewSchedule(w http.ResponseWriter, r *http.Request) {
	studentID, ok := h.previewStudentID(w, r)
	if !ok {
		return
	}
	schedule, err := h.svc.ScheduleByStudentID(r.Context(), studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Jadwal siswa tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"schedule": schedule, "preview": true})
}

func (h *StudentPortal) PreviewResults(w http.ResponseWriter, r *http.Request) {
	studentID, ok := h.previewStudentID(w, r)
	if !ok {
		return
	}
	results, err := h.svc.ResultsByStudentID(r.Context(), studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Hasil siswa tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"results": studentPortalResults(results), "preview": true})
}

func (h *StudentPortal) PreviewCbtSchedule(w http.ResponseWriter, r *http.Request) {
	studentID, ok := h.previewStudentID(w, r)
	if !ok {
		return
	}
	schedule, err := h.svc.CbtScheduleByStudentID(r.Context(), studentID)
	if err != nil {
		writeDomainOrInternal(w, err, "Jadwal CBT siswa tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"schedule": schedule, "preview": true, "token_reveal_disabled": true})
}

func (h *StudentPortal) Profile(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authorizedUserID(w, r)
	if !ok {
		return
	}
	student, err := h.svc.Profile(r.Context(), userID)
	if err != nil {
		writeDomainOrInternal(w, err, "Data profil siswa tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"student": studentPortalProfile(student)})
}

func (h *StudentPortal) Schedule(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authorizedUserID(w, r)
	if !ok {
		return
	}
	schedule, err := h.svc.Schedule(r.Context(), userID)
	if err != nil {
		writeDomainOrInternal(w, err, "Jadwal siswa tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"schedule": schedule})
}

func (h *StudentPortal) Results(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authorizedUserID(w, r)
	if !ok {
		return
	}
	results, err := h.svc.Results(r.Context(), userID)
	if err != nil {
		writeDomainOrInternal(w, err, "Hasil siswa tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"results": studentPortalResults(results)})
}

func (h *StudentPortal) CbtSchedule(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authorizedUserID(w, r)
	if !ok {
		return
	}
	schedule, err := h.svc.CbtSchedule(r.Context(), userID)
	if err != nil {
		writeDomainOrInternal(w, err, "Jadwal CBT siswa tidak tersedia")
		return
	}
	api.OK(w, map[string]any{"schedule": schedule})
}

func (h *StudentPortal) RevealCbtToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authorizedUserID(w, r)
	if !ok {
		return
	}
	participantID, err := parseUUID(strings.TrimSpace(chi.URLParam(r, "participantID")))
	if err != nil {
		api.BadRequest(w, "Peserta ujian tidak valid")
		return
	}
	var body struct {
		RoomToken string `json:"room_token"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	result, err := h.svc.RevealCbtToken(r.Context(), userID, participantID, body.RoomToken, trustedClientIP(r))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomTokenRequired):
			api.BadRequest(w, "room token required")
		case errors.Is(err, service.ErrRoomTokenMismatch):
			api.Err(w, http.StatusForbidden, "room token mismatch")
		default:
			writeDomainOrInternal(w, err, "Token ujian tidak dapat dibuka")
		}
		return
	}
	api.OK(w, result)
}

func (h *StudentPortal) authorizedUserID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	userID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	claims, _ := api.ClaimsFromContext(r.Context())
	if !studentPortalAccessAllowed(claims) {
		api.Forbidden(w)
		return pgtype.UUID{}, false
	}
	return userID, true
}

func studentPortalAccessAllowed(claims jwt.MapClaims) bool {
	return mw.HasAnyRole(claims, "siswa") || mw.HasAnyPermission(claims, "student_portal.read")
}

type studentPortalPreviewStudentDTO struct {
	ID        string `json:"id"`
	NIS       string `json:"nis"`
	NISN      string `json:"nisn"`
	Nama      string `json:"nama"`
	ClassID   string `json:"class_id"`
	ClassName string `json:"class_name"`
	ClassCode string `json:"class_code"`
	Status    string `json:"status"`
	IsActive  bool   `json:"is_active"`
}

func studentPortalPreviewStudents(rows []db.ListStudentPortalPreviewStudentsRow) []studentPortalPreviewStudentDTO {
	out := make([]studentPortalPreviewStudentDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, studentPortalPreviewStudentDTO{
			ID:        portalUUIDString(row.ID),
			NIS:       row.Nis,
			NISN:      row.Nisn,
			Nama:      row.Nama,
			ClassID:   portalUUIDString(row.ClassID),
			ClassName: portalTextString(row.ClassName),
			ClassCode: portalTextString(row.ClassCode),
			Status:    string(row.Status),
			IsActive:  row.IsActive,
		})
	}
	return out
}

func (h *StudentPortal) previewStudentID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	studentID, err := parseUUID(strings.TrimSpace(chi.URLParam(r, "studentID")))
	if err != nil {
		api.BadRequest(w, "ID siswa tidak valid")
		return pgtype.UUID{}, false
	}
	return studentID, true
}

type studentPortalProfileDTO struct {
	ID                string `json:"id"`
	NIS               string `json:"nis"`
	NISN              string `json:"nisn"`
	Nama              string `json:"nama"`
	Gender            string `json:"gender"`
	ParentName        string `json:"parent_name"`
	ParentPhone       string `json:"parent_phone"`
	ClassID           string `json:"class_id"`
	ClassName         string `json:"class_name"`
	ClassCode         string `json:"class_code"`
	LinkedParentNames string `json:"linked_parent_names"`
	LinkedParentCount int64  `json:"linked_parent_count"`
	IsActive          bool   `json:"is_active"`
	Status            string `json:"status"`
}

func studentPortalProfile(row db.GetStudentByIDRow) studentPortalProfileDTO {
	return studentPortalProfileDTO{
		ID:                portalUUIDString(row.ID),
		NIS:               row.Nis,
		NISN:              row.Nisn,
		Nama:              row.Nama,
		Gender:            string(row.Gender),
		ParentName:        row.ParentName,
		ParentPhone:       row.ParentPhone,
		ClassID:           portalUUIDString(row.ClassID),
		ClassName:         portalTextString(row.ClassName),
		ClassCode:         portalTextString(row.ClassCode),
		LinkedParentNames: string(row.LinkedParentNames),
		LinkedParentCount: row.LinkedParentCount,
		IsActive:          row.IsActive,
		Status:            string(row.Status),
	}
}

type studentPortalResultDTO struct {
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

func studentPortalResults(rows []db.ListStudentExamSessionsRow) []studentPortalResultDTO {
	out := make([]studentPortalResultDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, studentPortalResultDTO{
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

func portalUUIDString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return id.String()
}

func portalTextString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
