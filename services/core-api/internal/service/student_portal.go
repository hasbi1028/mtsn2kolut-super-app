package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type studentPortalStore interface {
	GetPortalStudentIDByUserID(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error)
	GetStudentByID(ctx context.Context, id pgtype.UUID) (db.GetStudentByIDRow, error)
	ListStudentTimetable(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error)
	ListStudentExamSessions(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error)
	ListStudentPortalCbtSchedule(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentPortalCbtScheduleRow, error)
	GetStudentPortalCbtParticipant(ctx context.Context, arg db.GetStudentPortalCbtParticipantParams) (db.GetStudentPortalCbtParticipantRow, error)
	InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error
}

type StudentPortal struct {
	q studentPortalStore
}

const studentPortalTokenWindowMinutes = 15

type StudentPortalCbtStatus string

const (
	StudentPortalCbtUpcoming    StudentPortalCbtStatus = "upcoming"
	StudentPortalCbtTokenWindow StudentPortalCbtStatus = "token_window"
	StudentPortalCbtActive      StudentPortalCbtStatus = "active"
	StudentPortalCbtSubmitted   StudentPortalCbtStatus = "submitted"
	StudentPortalCbtClosed      StudentPortalCbtStatus = "closed"
	StudentPortalCbtLocked      StudentPortalCbtStatus = "locked"
)

type StudentPortalCbtScheduleItem struct {
	ParticipantID     string                 `json:"participant_id"`
	SessionID         string                 `json:"session_id"`
	SessionTitle      string                 `json:"session_title"`
	PackageTitle      string                 `json:"package_title"`
	ScheduledStart    string                 `json:"scheduled_start"`
	ScheduledEnd      string                 `json:"scheduled_end"`
	DurationMinutes   int32                  `json:"duration_minutes"`
	RoomID            *string                `json:"room_id"`
	RoomName          *string                `json:"room_name"`
	SeatNo            *int32                 `json:"seat_no"`
	Status            StudentPortalCbtStatus `json:"status"`
	CanRevealToken    bool                   `json:"can_reveal_token"`
	RequiresRoomToken bool                   `json:"requires_room_token"`
	TokenMasked       *string                `json:"token_masked"`
}

type StudentPortalTokenReveal struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

func NewStudentPortal(q *db.Queries) *StudentPortal {
	return &StudentPortal{q: q}
}

func (s *StudentPortal) Profile(ctx context.Context, userID pgtype.UUID) (db.GetStudentByIDRow, error) {
	studentID, err := s.studentIDForUser(ctx, userID)
	if err != nil {
		return db.GetStudentByIDRow{}, err
	}
	return s.q.GetStudentByID(ctx, studentID)
}

func (s *StudentPortal) Schedule(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	studentID, err := s.studentIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.q.ListStudentTimetable(ctx, studentID)
}

func (s *StudentPortal) Results(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error) {
	studentID, err := s.studentIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.q.ListStudentExamSessions(ctx, studentID)
}

func (s *StudentPortal) CbtSchedule(ctx context.Context, userID pgtype.UUID) ([]StudentPortalCbtScheduleItem, error) {
	studentID, err := s.studentIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.ListStudentPortalCbtSchedule(ctx, studentID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	out := make([]StudentPortalCbtScheduleItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, studentPortalCbtScheduleItem(row, now))
	}
	return out, nil
}

func (s *StudentPortal) RevealCbtToken(ctx context.Context, userID, participantID pgtype.UUID, roomToken, clientIP string) (StudentPortalTokenReveal, error) {
	studentID, err := s.studentIDForUser(ctx, userID)
	if err != nil {
		return StudentPortalTokenReveal{}, err
	}
	if !participantID.Valid {
		return StudentPortalTokenReveal{}, domain.ErrBadRequest
	}
	row, err := s.q.GetStudentPortalCbtParticipant(ctx, db.GetStudentPortalCbtParticipantParams{
		ParticipantID: participantID,
		StudentID:     studentID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StudentPortalTokenReveal{}, domain.ErrForbidden
		}
		return StudentPortalTokenReveal{}, err
	}
	if !studentPortalCbtCanReveal(row, time.Now()) {
		return StudentPortalTokenReveal{}, fmt.Errorf("%w: token siswa belum dapat dibuka", domain.ErrForbidden)
	}
	if !row.RoomID.Valid {
		return StudentPortalTokenReveal{}, fmt.Errorf("%w: ruang ujian belum ditetapkan", domain.ErrForbidden)
	}
	if err := validateStudentPortalRoomToken(row, roomToken); err != nil {
		s.recordStudentPortalRoomTokenMismatch(ctx, row, userID, clientIP, err)
		return StudentPortalTokenReveal{}, err
	}

	expiresAt := studentPortalCbtRevealExpiresAt(row)
	_ = s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: row.ParticipantID,
		EventType:     "student_portal_token_reveal",
		EventData: marshalJSON(map[string]string{
			"session_id":    pgUUIDString(row.SessionID),
			"room_id":       pgUUIDString(row.RoomID),
			"student_id":    pgUUIDString(row.StudentID),
			"actor_user_id": pgUUIDString(userID),
			"ip_hash":       hashString(clientIP),
			"expires_at":    expiresAt.Format(time.RFC3339),
		}),
	})
	return StudentPortalTokenReveal{Token: row.Token, ExpiresAt: expiresAt.Format(time.RFC3339)}, nil
}

func (s *StudentPortal) studentIDForUser(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error) {
	if !userID.Valid {
		return pgtype.UUID{}, domain.ErrUnauthorized
	}
	studentID, err := s.q.GetPortalStudentIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgtype.UUID{}, domain.ErrForbidden
		}
		return pgtype.UUID{}, err
	}
	if !studentID.Valid {
		return pgtype.UUID{}, domain.ErrForbidden
	}
	return studentID, nil
}

var _ studentPortalStore = (*db.Queries)(nil)

func studentPortalCbtScheduleItem(row db.ListStudentPortalCbtScheduleRow, now time.Time) StudentPortalCbtScheduleItem {
	return StudentPortalCbtScheduleItem{
		ParticipantID:     pgUUIDString(row.ParticipantID),
		SessionID:         pgUUIDString(row.SessionID),
		SessionTitle:      firstNonEmpty(row.SessionTitle, row.PackageTitle, "Sesi Ujian"),
		PackageTitle:      row.PackageTitle,
		ScheduledStart:    timeString(row.ScheduledStart),
		ScheduledEnd:      timeString(row.ScheduledEnd),
		DurationMinutes:   row.DurationMinutes,
		RoomID:            uuidStringPtr(row.RoomID),
		RoomName:          textPtr(row.RoomName),
		SeatNo:            int4Ptr(row.SeatNo),
		Status:            studentPortalCbtStatus(row, now),
		CanRevealToken:    studentPortalCbtCanRevealSchedule(row, now),
		RequiresRoomToken: row.RoomID.Valid && strings.TrimSpace(row.RoomToken) != "",
		TokenMasked:       textPtr(maskExamToken(row.Token)),
	}
}

func studentPortalCbtStatus(row db.ListStudentPortalCbtScheduleRow, now time.Time) StudentPortalCbtStatus {
	if row.SubmittedAt.Valid {
		return StudentPortalCbtSubmitted
	}
	if row.LockedAt.Valid {
		return StudentPortalCbtLocked
	}
	if row.SessionStatus == db.CbtSessionStatusEnumFinished || row.SessionStatus == db.CbtSessionStatusEnumCancelled || portalCbtEnded(row.ScheduledEnd, now) {
		return StudentPortalCbtClosed
	}
	if row.SessionStatus == db.CbtSessionStatusEnumActive && portalCbtStarted(row.ScheduledStart, now) {
		return StudentPortalCbtActive
	}
	if studentPortalCbtCanRevealSchedule(row, now) {
		return StudentPortalCbtTokenWindow
	}
	return StudentPortalCbtUpcoming
}

func studentPortalCbtCanRevealSchedule(row db.ListStudentPortalCbtScheduleRow, now time.Time) bool {
	if row.SubmittedAt.Valid || row.LockedAt.Valid || !row.RoomID.Valid || strings.TrimSpace(row.RoomToken) == "" {
		return false
	}
	if row.SessionStatus != db.CbtSessionStatusEnumActive {
		return false
	}
	if !row.ScheduledStart.Valid {
		return false
	}
	if now.Before(row.ScheduledStart.Time.Add(-studentPortalTokenWindowMinutes * time.Minute)) {
		return false
	}
	return !portalCbtEnded(row.ScheduledEnd, now)
}

func studentPortalCbtCanReveal(row db.GetStudentPortalCbtParticipantRow, now time.Time) bool {
	if row.SubmittedAt.Valid || row.LockedAt.Valid || !row.RoomID.Valid || strings.TrimSpace(row.RoomToken) == "" {
		return false
	}
	if row.SessionStatus != db.CbtSessionStatusEnumActive || !row.ScheduledStart.Valid {
		return false
	}
	if now.Before(row.ScheduledStart.Time.Add(-studentPortalTokenWindowMinutes * time.Minute)) {
		return false
	}
	return !portalCbtEnded(row.ScheduledEnd, now)
}

func validateStudentPortalRoomToken(row db.GetStudentPortalCbtParticipantRow, roomToken string) error {
	roomToken = strings.TrimSpace(roomToken)
	if roomToken == "" {
		return ErrRoomTokenRequired
	}
	if strings.TrimSpace(row.RoomToken) == "" || !strings.EqualFold(strings.TrimSpace(row.RoomToken), roomToken) {
		return ErrRoomTokenMismatch
	}
	return nil
}

func (s *StudentPortal) recordStudentPortalRoomTokenMismatch(ctx context.Context, row db.GetStudentPortalCbtParticipantRow, userID pgtype.UUID, clientIP string, cause error) {
	_ = s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: row.ParticipantID,
		EventType:     "student_portal_room_token_mismatch",
		EventData: marshalJSON(map[string]string{
			"reason":                roomTokenFailureReason(cause),
			"session_id":            pgUUIDString(row.SessionID),
			"room_id":               pgUUIDString(row.RoomID),
			"student_id":            pgUUIDString(row.StudentID),
			"actor_user_id":         pgUUIDString(userID),
			"ip_hash":               hashString(clientIP),
			"room_token_configured": strconv.FormatBool(strings.TrimSpace(row.RoomToken) != ""),
		}),
	})
}

func studentPortalCbtRevealExpiresAt(row db.GetStudentPortalCbtParticipantRow) time.Time {
	if row.ScheduledEnd.Valid {
		return row.ScheduledEnd.Time
	}
	if row.ScheduledStart.Valid {
		return row.ScheduledStart.Time.Add(studentPortalTokenWindowMinutes * time.Minute)
	}
	return time.Now().Add(studentPortalTokenWindowMinutes * time.Minute)
}

func portalCbtStarted(start pgtype.Timestamptz, now time.Time) bool {
	return !start.Valid || !now.Before(start.Time)
}

func portalCbtEnded(end pgtype.Timestamptz, now time.Time) bool {
	return end.Valid && now.After(end.Time)
}

func timeString(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.RFC3339)
}

func uuidStringPtr(value pgtype.UUID) *string {
	if !value.Valid {
		return nil
	}
	out := value.String()
	return &out
}

func textPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func int4Ptr(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	out := value.Int32
	return &out
}

func maskExamToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 4 {
		return strings.Repeat("•", len(token))
	}
	prefix := token[:4]
	groups := []string{strings.ToUpper(prefix)}
	remaining := len(token) - 4
	for remaining > 0 {
		width := 4
		if remaining < width {
			width = remaining
		}
		groups = append(groups, strings.Repeat("•", width))
		remaining -= width
	}
	return strings.Join(groups, "-")
}
