package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var (
	ErrCbtAccessCardInvalid = errors.New("cbt access card invalid")
	ErrCbtAccessCardLocked  = errors.New("cbt access card locked")
	ErrCbtAccessCardExpired = errors.New("cbt access card expired")
)

type cbtAccessCardStore interface {
	ListCbtParticipantAccessCardTargetsByEvent(ctx context.Context, arg db.ListCbtParticipantAccessCardTargetsByEventParams) ([]db.ListCbtParticipantAccessCardTargetsByEventRow, error)
	ListCbtProctorAccessCardTargetsByEvent(ctx context.Context, arg db.ListCbtProctorAccessCardTargetsByEventParams) ([]db.ListCbtProctorAccessCardTargetsByEventRow, error)
	RevokeActiveParticipantAccessCards(ctx context.Context, participantID pgtype.UUID) error
	RevokeActiveProctorAccessCards(ctx context.Context, roomProctorID pgtype.UUID) error
	CreateCbtExamAccessCard(ctx context.Context, arg db.CreateCbtExamAccessCardParams) (db.CbtExamAccessCard, error)
	GetCbtAccessCardByTokenHash(ctx context.Context, tokenHash string) (db.GetCbtAccessCardByTokenHashRow, error)
	MarkCbtAccessCardVerified(ctx context.Context, id pgtype.UUID) error
	RecordCbtAccessCardFailedAttempt(ctx context.Context, id pgtype.UUID) (db.RecordCbtAccessCardFailedAttemptRow, error)
	InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error
}

type CbtAccessCard struct{ q cbtAccessCardStore }

type CbtAccessCardIssueInput struct {
	EventID      pgtype.UUID
	SessionID    pgtype.UUID
	RoomID       pgtype.UUID
	Regenerate   bool
	GeneratedBy  pgtype.UUID
	ExpiresHours int32
}

type CbtAccessCardVerifyInput struct {
	Token    string
	PIN      string
	CardType string
	ClientIP string
}

type CbtParticipantAccessCard struct {
	CardID         *string `json:"card_id,omitempty"`
	ParticipantID  string  `json:"participant_id"`
	SessionID      string  `json:"session_id"`
	EventID        string  `json:"event_id"`
	StudentID      string  `json:"student_id"`
	NIS            string  `json:"nis"`
	NISN           string  `json:"nisn"`
	StudentName    string  `json:"student_name"`
	ClassName      string  `json:"class_name"`
	ClassCode      string  `json:"class_code"`
	RoomID         *string `json:"room_id,omitempty"`
	RoomName       string  `json:"room_name"`
	SeatNo         *int32  `json:"seat_no,omitempty"`
	SessionTitle   string  `json:"session_title"`
	PackageTitle   string  `json:"package_title"`
	ScheduledStart string  `json:"scheduled_start"`
	ScheduledEnd   string  `json:"scheduled_end"`
	Status         string  `json:"status"`
	FailedAttempts int32   `json:"failed_attempts"`
	CardCreatedAt  string  `json:"card_created_at,omitempty"`
	CardExpiresAt  string  `json:"card_expires_at,omitempty"`
	CardVerifiedAt string  `json:"card_verified_at,omitempty"`
	Token          string  `json:"token,omitempty"`
	PIN            string  `json:"pin,omitempty"`
	QRPath         string  `json:"qr_path,omitempty"`
}

type CbtProctorAccessCard struct {
	CardID         *string `json:"card_id,omitempty"`
	RoomProctorID  string  `json:"room_proctor_id"`
	RoomID         string  `json:"room_id"`
	SessionID      string  `json:"session_id"`
	EventID        string  `json:"event_id"`
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	NIP            string  `json:"nip"`
	ProctorRole    string  `json:"proctor_role"`
	RoomName       string  `json:"room_name"`
	SessionTitle   string  `json:"session_title"`
	PackageTitle   string  `json:"package_title"`
	ScheduledStart string  `json:"scheduled_start"`
	ScheduledEnd   string  `json:"scheduled_end"`
	Status         string  `json:"status"`
	FailedAttempts int32   `json:"failed_attempts"`
	CardCreatedAt  string  `json:"card_created_at,omitempty"`
	CardExpiresAt  string  `json:"card_expires_at,omitempty"`
	CardVerifiedAt string  `json:"card_verified_at,omitempty"`
	Token          string  `json:"token,omitempty"`
	PIN            string  `json:"pin,omitempty"`
	QRPath         string  `json:"qr_path,omitempty"`
}

type CbtAccessCardVerifyResult struct {
	CardID          string  `json:"card_id"`
	CardType        string  `json:"card_type"`
	EventID         string  `json:"event_id"`
	EventTitle      string  `json:"event_title"`
	SessionID       string  `json:"session_id"`
	SessionTitle    string  `json:"session_title"`
	SessionStatus   string  `json:"session_status"`
	ScheduledStart  string  `json:"scheduled_start"`
	ScheduledEnd    string  `json:"scheduled_end"`
	RoomID          *string `json:"room_id,omitempty"`
	RoomName        string  `json:"room_name"`
	ParticipantID   *string `json:"participant_id,omitempty"`
	StudentID       *string `json:"student_id,omitempty"`
	NIS             string  `json:"nis,omitempty"`
	NISN            string  `json:"nisn,omitempty"`
	StudentName     string  `json:"student_name,omitempty"`
	ClassName       string  `json:"class_name,omitempty"`
	SeatNo          *int32  `json:"seat_no,omitempty"`
	RoomProctorID   *string `json:"room_proctor_id,omitempty"`
	EmployeeID      *string `json:"employee_id,omitempty"`
	ProctorName     string  `json:"proctor_name,omitempty"`
	ProctorRole     string  `json:"proctor_role,omitempty"`
	PackageTitle    string  `json:"package_title"`
	DurationMinutes int32   `json:"duration_minutes"`
}

func NewCbtAccessCard(q *db.Queries) *CbtAccessCard { return &CbtAccessCard{q: q} }

func (s *CbtAccessCard) ListParticipantCards(ctx context.Context, in CbtAccessCardIssueInput) ([]CbtParticipantAccessCard, error) {
	if !in.EventID.Valid {
		return nil, domain.ErrBadRequest
	}
	rows, err := s.q.ListCbtParticipantAccessCardTargetsByEvent(ctx, db.ListCbtParticipantAccessCardTargetsByEventParams{EventID: in.EventID, SessionID: in.SessionID, RoomID: in.RoomID})
	if err != nil {
		return nil, err
	}
	out := make([]CbtParticipantAccessCard, 0, len(rows))
	for _, row := range rows {
		out = append(out, participantCardDTO(row, "", ""))
	}
	return out, nil
}

func (s *CbtAccessCard) IssueParticipantCards(ctx context.Context, in CbtAccessCardIssueInput) ([]CbtParticipantAccessCard, error) {
	if !in.EventID.Valid {
		return nil, domain.ErrBadRequest
	}
	rows, err := s.q.ListCbtParticipantAccessCardTargetsByEvent(ctx, db.ListCbtParticipantAccessCardTargetsByEventParams{EventID: in.EventID, SessionID: in.SessionID, RoomID: in.RoomID})
	if err != nil {
		return nil, err
	}
	expiresAt := cardExpiresAt(in.ExpiresHours)
	out := make([]CbtParticipantAccessCard, 0, len(rows))
	for _, row := range rows {
		if !in.Regenerate && row.CardID.Valid {
			out = append(out, participantCardDTO(row, "", ""))
			continue
		}
		token, pin, err := newCardSecrets()
		if err != nil {
			return nil, err
		}
		if err := s.q.RevokeActiveParticipantAccessCards(ctx, row.ParticipantID); err != nil {
			return nil, err
		}
		card, err := s.q.CreateCbtExamAccessCard(ctx, db.CreateCbtExamAccessCardParams{CardType: "participant", EventID: row.EventID, SessionID: row.SessionID, RoomID: row.RoomID, ParticipantID: row.ParticipantID, TokenHash: hashCardSecret(token), PinHash: hashCardSecret(pin), ExpiresAt: expiresAt, GeneratedBy: in.GeneratedBy})
		if err != nil {
			return nil, err
		}
		row.CardID, row.CardStatus, row.FailedAttempts, row.CardCreatedAt, row.CardExpiresAt = card.ID, card.Status, card.FailedAttempts, card.CreatedAt, card.ExpiresAt
		item := participantCardDTO(row, token, pin)
		item.QRPath = "/ujian?card=" + token
		out = append(out, item)
	}
	return out, nil
}

func (s *CbtAccessCard) ListProctorCards(ctx context.Context, in CbtAccessCardIssueInput) ([]CbtProctorAccessCard, error) {
	if !in.EventID.Valid {
		return nil, domain.ErrBadRequest
	}
	rows, err := s.q.ListCbtProctorAccessCardTargetsByEvent(ctx, db.ListCbtProctorAccessCardTargetsByEventParams{EventID: in.EventID, SessionID: in.SessionID, RoomID: in.RoomID})
	if err != nil {
		return nil, err
	}
	out := make([]CbtProctorAccessCard, 0, len(rows))
	for _, row := range rows {
		out = append(out, proctorCardDTO(row, "", ""))
	}
	return out, nil
}

func (s *CbtAccessCard) IssueProctorCards(ctx context.Context, in CbtAccessCardIssueInput) ([]CbtProctorAccessCard, error) {
	if !in.EventID.Valid {
		return nil, domain.ErrBadRequest
	}
	rows, err := s.q.ListCbtProctorAccessCardTargetsByEvent(ctx, db.ListCbtProctorAccessCardTargetsByEventParams{EventID: in.EventID, SessionID: in.SessionID, RoomID: in.RoomID})
	if err != nil {
		return nil, err
	}
	expiresAt := cardExpiresAt(in.ExpiresHours)
	out := make([]CbtProctorAccessCard, 0, len(rows))
	for _, row := range rows {
		if !in.Regenerate && row.CardID.Valid {
			out = append(out, proctorCardDTO(row, "", ""))
			continue
		}
		token, pin, err := newCardSecrets()
		if err != nil {
			return nil, err
		}
		if err := s.q.RevokeActiveProctorAccessCards(ctx, row.RoomProctorID); err != nil {
			return nil, err
		}
		card, err := s.q.CreateCbtExamAccessCard(ctx, db.CreateCbtExamAccessCardParams{CardType: "proctor", EventID: row.EventID, SessionID: row.SessionID, RoomID: row.RoomID, RoomProctorID: row.RoomProctorID, TokenHash: hashCardSecret(token), PinHash: hashCardSecret(pin), ExpiresAt: expiresAt, GeneratedBy: in.GeneratedBy})
		if err != nil {
			return nil, err
		}
		row.CardID, row.CardStatus, row.FailedAttempts, row.CardCreatedAt, row.CardExpiresAt = card.ID, card.Status, card.FailedAttempts, card.CreatedAt, card.ExpiresAt
		item := proctorCardDTO(row, token, pin)
		item.QRPath = "/pengawas-ujian?card=" + token
		out = append(out, item)
	}
	return out, nil
}

func (s *CbtAccessCard) Verify(ctx context.Context, in CbtAccessCardVerifyInput) (CbtAccessCardVerifyResult, error) {
	token := strings.TrimSpace(in.Token)
	pin := strings.TrimSpace(in.PIN)
	if token == "" || pin == "" {
		return CbtAccessCardVerifyResult{}, ErrCbtAccessCardInvalid
	}
	row, err := s.q.GetCbtAccessCardByTokenHash(ctx, hashCardSecret(token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CbtAccessCardVerifyResult{}, ErrCbtAccessCardInvalid
		}
		return CbtAccessCardVerifyResult{}, err
	}
	if in.CardType != "" && row.CardType != in.CardType {
		return CbtAccessCardVerifyResult{}, ErrCbtAccessCardInvalid
	}
	if row.RevokedAt.Valid || row.CardStatus == "revoked" {
		return CbtAccessCardVerifyResult{}, ErrCbtAccessCardInvalid
	}
	if row.CardStatus == "locked" {
		return CbtAccessCardVerifyResult{}, ErrCbtAccessCardLocked
	}
	if row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
		return CbtAccessCardVerifyResult{}, ErrCbtAccessCardExpired
	}
	if subtleStringMismatch(row.PinHash, hashCardSecret(pin)) {
		_, _ = s.q.RecordCbtAccessCardFailedAttempt(ctx, row.ID)
		if row.ParticipantID.Valid {
			s.logParticipantCardEvent(ctx, row.ParticipantID, "cbt_access_card_pin_failed", in.ClientIP)
		}
		return CbtAccessCardVerifyResult{}, ErrCbtAccessCardInvalid
	}
	if err := s.q.MarkCbtAccessCardVerified(ctx, row.ID); err != nil {
		return CbtAccessCardVerifyResult{}, err
	}
	if row.ParticipantID.Valid {
		s.logParticipantCardEvent(ctx, row.ParticipantID, "cbt_access_card_verified", in.ClientIP)
	}
	return verifyResultDTO(row), nil
}

func (s *CbtAccessCard) logParticipantCardEvent(ctx context.Context, participantID pgtype.UUID, eventType, clientIP string) {
	_ = s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{ParticipantID: participantID, EventType: eventType, EventData: marshalJSON(map[string]string{"ip_hash": hashString(clientIP)})})
}

func newCardSecrets() (string, string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := "eca_" + hex.EncodeToString(b)
	pin, err := randomDigits(4)
	return token, pin, err
}

func randomDigits(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, v := range b {
		out[i] = byte('0' + int(v)%10)
	}
	return string(out), nil
}

func hashCardSecret(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func cardExpiresAt(hours int32) pgtype.Timestamptz {
	if hours <= 0 {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: time.Now().Add(time.Duration(hours) * time.Hour), Valid: true}
}

func participantCardDTO(row db.ListCbtParticipantAccessCardTargetsByEventRow, token, pin string) CbtParticipantAccessCard {
	return CbtParticipantAccessCard{CardID: uuidStringPtr(row.CardID), ParticipantID: pgUUIDString(row.ParticipantID), SessionID: pgUUIDString(row.SessionID), EventID: pgUUIDString(row.EventID), StudentID: pgUUIDString(row.StudentID), NIS: row.Nis, NISN: row.Nisn, StudentName: row.StudentName, ClassName: row.ClassName, ClassCode: row.ClassCode, RoomID: uuidStringPtr(row.RoomID), RoomName: row.RoomName, SeatNo: int4Ptr(row.SeatNo), SessionTitle: firstNonEmpty(row.SessionTitle, row.PackageTitle, "Sesi CBT"), PackageTitle: row.PackageTitle, ScheduledStart: timeString(row.ScheduledStart), ScheduledEnd: timeString(row.ScheduledEnd), Status: firstNonEmpty(row.CardStatus, "not_issued"), FailedAttempts: row.FailedAttempts, CardCreatedAt: timeString(row.CardCreatedAt), CardExpiresAt: timeString(row.CardExpiresAt), CardVerifiedAt: timeString(row.CardVerifiedAt), Token: token, PIN: pin}
}

func proctorCardDTO(row db.ListCbtProctorAccessCardTargetsByEventRow, token, pin string) CbtProctorAccessCard {
	return CbtProctorAccessCard{CardID: uuidStringPtr(row.CardID), RoomProctorID: pgUUIDString(row.RoomProctorID), RoomID: pgUUIDString(row.RoomID), SessionID: pgUUIDString(row.SessionID), EventID: pgUUIDString(row.EventID), EmployeeID: pgUUIDString(row.EmployeeID), EmployeeName: row.EmployeeName, NIP: row.Nip.String, ProctorRole: row.ProctorRole, RoomName: row.RoomName, SessionTitle: firstNonEmpty(row.SessionTitle, row.PackageTitle, "Sesi CBT"), PackageTitle: row.PackageTitle, ScheduledStart: timeString(row.ScheduledStart), ScheduledEnd: timeString(row.ScheduledEnd), Status: firstNonEmpty(row.CardStatus, "not_issued"), FailedAttempts: row.FailedAttempts, CardCreatedAt: timeString(row.CardCreatedAt), CardExpiresAt: timeString(row.CardExpiresAt), CardVerifiedAt: timeString(row.CardVerifiedAt), Token: token, PIN: pin}
}

func verifyResultDTO(row db.GetCbtAccessCardByTokenHashRow) CbtAccessCardVerifyResult {
	return CbtAccessCardVerifyResult{CardID: pgUUIDString(row.ID), CardType: row.CardType, EventID: pgUUIDString(row.EventID), EventTitle: row.EventTitle, SessionID: pgUUIDString(row.SessionID), SessionTitle: firstNonEmpty(row.SessionTitle, row.PackageTitle, "Sesi CBT"), SessionStatus: string(row.SessionStatus), ScheduledStart: timeString(row.ScheduledStart), ScheduledEnd: timeString(row.ScheduledEnd), RoomID: uuidStringPtr(row.RoomID), RoomName: row.RoomName, ParticipantID: uuidStringPtr(row.ParticipantID), StudentID: uuidStringPtr(row.StudentID), NIS: row.Nis, NISN: row.Nisn, StudentName: row.StudentName, ClassName: row.ClassName, SeatNo: int4Ptr(row.SeatNo), RoomProctorID: uuidStringPtr(row.RoomProctorID), EmployeeID: uuidStringPtr(row.ProctorEmployeeID), ProctorName: row.ProctorName, ProctorRole: row.ProctorRole, PackageTitle: row.PackageTitle, DurationMinutes: row.DurationMinutes}
}

var _ cbtAccessCardStore = (*db.Queries)(nil)
