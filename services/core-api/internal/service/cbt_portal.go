package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var (
	ErrCbtPortalLoginInvalid = errors.New("cbt portal login invalid")
	ErrCbtPortalDisabled     = errors.New("cbt portal direct login disabled")
)

type cbtPortalStore interface {
	GetCbtPortalStudentByNisn(ctx context.Context, btrim string) (db.GetCbtPortalStudentByNisnRow, error)
	ListCbtPortalParticipantsByStudent(ctx context.Context, studentID pgtype.UUID) ([]db.ListCbtPortalParticipantsByStudentRow, error)
	InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error
}

type CbtPortal struct {
	q cbtPortalStore
}

type CbtPortalLoginResult struct {
	Student  CbtPortalStudent        `json:"student"`
	Schedule []CbtPortalScheduleItem `json:"schedule"`
}

type CbtPortalStudent struct {
	ID        string `json:"id"`
	NIS       string `json:"nis"`
	NISN      string `json:"nisn"`
	Nama      string `json:"nama"`
	ClassID   string `json:"class_id"`
	ClassName string `json:"class_name"`
	ClassCode string `json:"class_code"`
}

type CbtPortalScheduleItem struct {
	ParticipantID   string                 `json:"participant_id"`
	SessionID       string                 `json:"session_id"`
	SessionTitle    string                 `json:"session_title"`
	PackageTitle    string                 `json:"package_title"`
	ScheduledStart  string                 `json:"scheduled_start"`
	ScheduledEnd    string                 `json:"scheduled_end"`
	DurationMinutes int32                  `json:"duration_minutes"`
	RoomID          *string                `json:"room_id"`
	RoomName        *string                `json:"room_name"`
	SeatNo          *int32                 `json:"seat_no"`
	Status          StudentPortalCbtStatus `json:"status"`
	CanStart        bool                   `json:"can_start"`
	AccessMode      string                 `json:"access_mode"`
}

func NewCbtPortal(q *db.Queries) *CbtPortal {
	return &CbtPortal{q: q}
}

func (s *CbtPortal) LoginByNISN(ctx context.Context, nisn, code, clientIP string) (CbtPortalLoginResult, error) {
	nisn = normalizeCbtPortalNISN(nisn)
	code = normalizeCbtPortalNISN(code)
	if nisn == "" || code == "" || subtleStringMismatch(nisn, code) {
		return CbtPortalLoginResult{}, ErrCbtPortalLoginInvalid
	}
	student, err := s.q.GetCbtPortalStudentByNisn(ctx, nisn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CbtPortalLoginResult{}, ErrCbtPortalLoginInvalid
		}
		return CbtPortalLoginResult{}, err
	}
	if !student.IsActive {
		return CbtPortalLoginResult{}, ErrCbtPortalLoginInvalid
	}
	schedule, err := s.ScheduleByStudentID(ctx, student.ID)
	if err != nil {
		return CbtPortalLoginResult{}, err
	}
	return CbtPortalLoginResult{Student: cbtPortalStudentDTO(student), Schedule: schedule}, nil
}

func (s *CbtPortal) ScheduleByStudentID(ctx context.Context, studentID pgtype.UUID) ([]CbtPortalScheduleItem, error) {
	if !studentID.Valid {
		return nil, domain.ErrBadRequest
	}
	rows, err := s.q.ListCbtPortalParticipantsByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	items := make([]CbtPortalScheduleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, cbtPortalScheduleItem(row, now))
	}
	return items, nil
}

func (s *CbtPortal) AssertCanStart(ctx context.Context, studentID, participantID pgtype.UUID, clientIP string) error {
	if !studentID.Valid || !participantID.Valid {
		return domain.ErrBadRequest
	}
	rows, err := s.q.ListCbtPortalParticipantsByStudent(ctx, studentID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.ParticipantID == participantID {
			if !cbtPortalCanStart(row, time.Now()) {
				return fmt.Errorf("%w: sesi belum dapat dimulai", domain.ErrForbidden)
			}
			_ = s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
				ParticipantID: participantID,
				EventType:     "cbt_portal_direct_start_attempt",
				EventData: marshalJSON(map[string]string{
					"session_id":  pgUUIDString(row.SessionID),
					"student_id":  pgUUIDString(studentID),
					"access_mode": strings.TrimSpace(row.AccessMode),
					"ip_hash":     hashString(clientIP),
				}),
			})
			return nil
		}
	}
	return domain.ErrForbidden
}

func cbtPortalStudentDTO(row db.GetCbtPortalStudentByNisnRow) CbtPortalStudent {
	return CbtPortalStudent{
		ID:        pgUUIDString(row.ID),
		NIS:       row.Nis,
		NISN:      row.Nisn,
		Nama:      row.Nama,
		ClassID:   pgUUIDString(row.ClassID),
		ClassName: row.ClassName,
		ClassCode: row.ClassCode,
	}
}

func cbtPortalScheduleItem(row db.ListCbtPortalParticipantsByStudentRow, now time.Time) CbtPortalScheduleItem {
	status := cbtPortalStatus(row, now)
	return CbtPortalScheduleItem{
		ParticipantID:   pgUUIDString(row.ParticipantID),
		SessionID:       pgUUIDString(row.SessionID),
		SessionTitle:    firstNonEmpty(row.SessionTitle, row.PackageTitle, "Sesi CBT"),
		PackageTitle:    row.PackageTitle,
		ScheduledStart:  timeString(row.ScheduledStart),
		ScheduledEnd:    timeString(row.ScheduledEnd),
		DurationMinutes: row.DurationMinutes,
		RoomID:          uuidStringPtr(row.RoomID),
		RoomName:        textPtr(row.RoomName),
		SeatNo:          int4Ptr(row.SeatNo),
		Status:          status,
		CanStart:        cbtPortalCanStart(row, now),
		AccessMode:      row.AccessMode,
	}
}

func cbtPortalStatus(row db.ListCbtPortalParticipantsByStudentRow, now time.Time) StudentPortalCbtStatus {
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
	return StudentPortalCbtUpcoming
}

func cbtPortalCanStart(row db.ListCbtPortalParticipantsByStudentRow, now time.Time) bool {
	if row.SubmittedAt.Valid || row.LockedAt.Valid {
		return false
	}
	if row.SessionStatus != db.CbtSessionStatusEnumActive {
		return false
	}
	if !row.NisnDirectLoginEnabled || !row.StudentPortalDirectLoginEnabled || row.RequireRoomTokenForWeb {
		return false
	}
	if row.AccessMode != "simulation" && row.AccessMode != "web_fallback" {
		return false
	}
	if row.RoomID.Valid && !row.RoomAllowWebFallback {
		return false
	}
	return portalCbtStarted(row.ScheduledStart, now) && !portalCbtEnded(row.ScheduledEnd, now)
}

func normalizeCbtPortalNISN(value string) string {
	return strings.TrimSpace(value)
}

func subtleStringMismatch(a, b string) bool {
	if len(a) != len(b) {
		return true
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff != 0
}

var _ cbtPortalStore = (*db.Queries)(nil)
