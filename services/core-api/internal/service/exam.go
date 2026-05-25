package service

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var (
	ErrExamNotFound         = errors.New("exam token not found")
	ErrExamNotActive        = errors.New("exam session is not active")
	ErrExamAlreadySubmit    = errors.New("exam already submitted")
	ErrExamWindowClosed     = errors.New("exam window has closed")
	ErrExamNotStarted       = errors.New("exam session has not started")
	ErrDeviceMismatch       = errors.New("token already bound to another device")
	ErrDeviceRequired       = errors.New("device fingerprint required")
	ErrExamRoomRequired     = errors.New("exam room has not been assigned")
	ErrRoomTokenRequired    = errors.New("room token required")
	ErrRoomTokenMismatch    = errors.New("room token mismatch")
	ErrWebFallbackDisabled  = errors.New("web fallback is not enabled for exam room")
	ErrExamQuestionScope    = errors.New("question is not part of participant exam")
	ErrExamLocked           = errors.New("exam locked by anti-cheat policy")
	ErrExamInvalidTelemetry = errors.New("invalid exam telemetry event")
)

type Exam struct {
	q    examStore
	pool *pgxpool.Pool
}

type examStore interface {
	GetParticipantByToken(ctx context.Context, token string) (db.GetParticipantByTokenRow, error)
	UpdateParticipantLogin(ctx context.Context, arg db.UpdateParticipantLoginParams) (pgtype.UUID, error)
	InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error
	GetExamQuestions(ctx context.Context, packageID pgtype.UUID) ([]db.GetExamQuestionsRow, error)
	SetParticipantQuestionOrderIfEmpty(ctx context.Context, arg db.SetParticipantQuestionOrderIfEmptyParams) ([]byte, error)
	GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error)
	GetCbtExamRoom(ctx context.Context, id pgtype.UUID) (db.CbtExamRoom, error)
	UpdateParticipantHeartbeat(ctx context.Context, participantID pgtype.UUID) error
	IncrementParticipantAppSwitch(ctx context.Context, participantID pgtype.UUID) error
	IncrementParticipantScreenshot(ctx context.Context, participantID pgtype.UUID) error
	IncrementParticipantAntiCheatViolation(ctx context.Context, arg db.IncrementParticipantAntiCheatViolationParams) (db.IncrementParticipantAntiCheatViolationRow, error)
	QuestionBelongsToParticipantPackage(ctx context.Context, arg db.QuestionBelongsToParticipantPackageParams) (bool, error)
	UpsertStudentAnswer(ctx context.Context, arg db.UpsertStudentAnswerParams) (int64, error)
	UpdateParticipantAnswerCorrectness(ctx context.Context, participantID pgtype.UUID) error
	SubmitParticipantExam(ctx context.Context, id pgtype.UUID) (db.SubmitParticipantExamRow, error)
	ListCbtQuestionAssetsByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error)
	ListPendingParticipantCommands(ctx context.Context, participantID pgtype.UUID) ([]db.ListPendingParticipantCommandsRow, error)
}

type examTxStore interface {
	WithTx(tx pgx.Tx) *db.Queries
}

type examRuntimePlanStore interface {
	SetParticipantRuntimePlanIfEmpty(ctx context.Context, arg db.SetParticipantRuntimePlanIfEmptyParams) (db.SetParticipantRuntimePlanIfEmptyRow, error)
	SetParticipantOptionOrderIfEmpty(ctx context.Context, arg db.SetParticipantOptionOrderIfEmptyParams) ([]byte, error)
}

type examProctorTelemetryStore interface {
	GetCbtParticipantRiskForUpdate(ctx context.Context, id pgtype.UUID) (db.GetCbtParticipantRiskForUpdateRow, error)
	GetRecentCbtProctorEventByDedupKey(ctx context.Context, arg db.GetRecentCbtProctorEventByDedupKeyParams) (db.GetRecentCbtProctorEventByDedupKeyRow, error)
	UpdateCbtParticipantProctorRisk(ctx context.Context, arg db.UpdateCbtParticipantProctorRiskParams) (db.UpdateCbtParticipantProctorRiskRow, error)
	CreateCbtParticipantProctorEvent(ctx context.Context, arg db.CreateCbtParticipantProctorEventParams) (db.CreateCbtParticipantProctorEventRow, error)
}

func NewExam(pool *pgxpool.Pool) *Exam {
	return &Exam{q: db.New(pool), pool: pool}
}

func (s *Exam) GetParticipantByToken(ctx context.Context, token string) (db.GetParticipantByTokenRow, error) {
	return s.q.GetParticipantByToken(ctx, token)
}

type LoginResult struct {
	ParticipantID        string         `json:"participant_id"`
	Student              StudentInfo    `json:"student"`
	Session              SessionInfo    `json:"session"`
	Room                 *RoomInfo      `json:"room,omitempty"`
	Questions            []ExamQuestion `json:"questions"`
	AnsweredCount        int            `json:"answered_count"`
	TotalQuestions       int            `json:"total_questions"`
	TimeRemainingSeconds int64          `json:"time_remaining_seconds"`
	AntiCheat            AntiCheatState `json:"anti_cheat"`
}

type AntiCheatState struct {
	ViolationCount int    `json:"violation_count"`
	RiskScore      int    `json:"risk_score"`
	RiskLevel      string `json:"risk_level"`
	Locked         bool   `json:"locked"`
	LockedAt       string `json:"locked_at,omitempty"`
	LockedReason   string `json:"locked_reason,omitempty"`
}

type ParticipantCommand struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Message     string         `json:"message"`
	Severity    string         `json:"severity,omitempty"`
	IssuedAt    string         `json:"issued_at"`
	Actor       string         `json:"actor,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
	RawEvent    string         `json:"raw_event,omitempty"`
	LegacyLabel string         `json:"legacy_label,omitempty"`
}

type StudentInfo struct {
	NIS  string `json:"nis"`
	Nama string `json:"nama"`
}

type SessionInfo struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	ScheduledStart  string `json:"scheduled_start"`
	ScheduledEnd    string `json:"scheduled_end"`
	DurationMinutes int32  `json:"duration_minutes"`
}

type RoomInfo struct {
	RoomName string `json:"room_name"`
}

type ExamQuestion struct {
	ID               string          `json:"id"`
	Code             string          `json:"code"`
	QuestionText     string          `json:"question_text"`
	QuestionType     string          `json:"question_type"`
	Options          json.RawMessage `json:"options"`
	StemHTML         string          `json:"stem_html"`
	StimulusHTML     string          `json:"stimulus_html"`
	StemMediaURL     string          `json:"stem_media_url"`
	StimulusMediaURL string          `json:"stimulus_media_url"`
	StemAudioURL     string          `json:"stem_audio_url"`
	StimulusAudioURL string          `json:"stimulus_audio_url"`
	// Legacy fields — included for backward compat, empty for new questions
	OptionA string `json:"option_a,omitempty"`
	OptionB string `json:"option_b,omitempty"`
	OptionC string `json:"option_c,omitempty"`
	OptionD string `json:"option_d,omitempty"`
	OptionE string `json:"option_e,omitempty"`
}

type ExamLoginRequest struct {
	Token              string
	RoomToken          string
	DeviceFingerprint  string
	LoginIP            string
	ClientType         string
	BrowserFingerprint string
	UserAgent          string
}

func (s *Exam) Login(ctx context.Context, token, roomToken, deviceFingerprint, loginIP string) (LoginResult, error) {
	return s.LoginWithClient(ctx, ExamLoginRequest{
		Token:             token,
		RoomToken:         roomToken,
		DeviceFingerprint: deviceFingerprint,
		LoginIP:           loginIP,
	})
}

func (s *Exam) LoginWithClient(ctx context.Context, req ExamLoginRequest) (LoginResult, error) {
	clientType := normalizeExamClientType(req.ClientType)
	deviceFingerprint := strings.TrimSpace(req.DeviceFingerprint)
	browserFingerprint := strings.TrimSpace(req.BrowserFingerprint)
	if clientType == "web_fallback" && deviceFingerprint == "" {
		deviceFingerprint = browserFingerprint
	}
	if deviceFingerprint == "" {
		return LoginResult{}, ErrDeviceRequired
	}
	p, err := s.q.GetParticipantByToken(ctx, strings.TrimSpace(req.Token))
	if err != nil {
		return LoginResult{}, ErrExamNotFound
	}
	if !p.RoomID.Valid {
		return LoginResult{}, ErrExamRoomRequired
	}
	if err := validateParticipantRoomToken(p, req.RoomToken); err != nil {
		s.recordRoomTokenMismatch(ctx, p, req.LoginIP, err)
		return LoginResult{}, err
	}
	return s.loginParticipant(ctx, p, req, deviceFingerprint, browserFingerprint, clientType, "login")
}

func (s *Exam) LoginCbtPortalDirect(ctx context.Context, participantID, studentID pgtype.UUID, deviceFingerprint, loginIP, browserFingerprint, userAgent string) (LoginResult, error) {
	deviceFingerprint = strings.TrimSpace(deviceFingerprint)
	browserFingerprint = strings.TrimSpace(browserFingerprint)
	if deviceFingerprint == "" {
		deviceFingerprint = browserFingerprint
	}
	if deviceFingerprint == "" {
		return LoginResult{}, ErrDeviceRequired
	}
	store, ok := s.q.(interface {
		GetCbtPortalExamParticipant(context.Context, db.GetCbtPortalExamParticipantParams) (db.GetCbtPortalExamParticipantRow, error)
	})
	if !ok {
		return LoginResult{}, ErrExamNotFound
	}
	row, err := store.GetCbtPortalExamParticipant(ctx, db.GetCbtPortalExamParticipantParams{ParticipantID: participantID, StudentID: studentID})
	if err != nil {
		return LoginResult{}, ErrExamNotFound
	}
	p := cbtPortalExamParticipantToLoginRow(row)
	if !p.NisnDirectLoginEnabled || !p.StudentPortalDirectLoginEnabled || p.RequireRoomTokenForWeb {
		return LoginResult{}, ErrCbtPortalDisabled
	}
	if p.AccessMode != "simulation" && p.AccessMode != "web_fallback" {
		return LoginResult{}, ErrCbtPortalDisabled
	}
	if p.RoomID.Valid && !p.RoomAllowWebFallback {
		return LoginResult{}, ErrWebFallbackDisabled
	}
	return s.loginParticipant(ctx, p, ExamLoginRequest{
		DeviceFingerprint:  deviceFingerprint,
		LoginIP:            loginIP,
		ClientType:         "cbt_portal",
		BrowserFingerprint: browserFingerprint,
		UserAgent:          userAgent,
	}, deviceFingerprint, browserFingerprint, "cbt_portal", "cbt_portal_direct_login")
}

func (s *Exam) loginParticipant(ctx context.Context, p db.GetParticipantByTokenRow, req ExamLoginRequest, deviceFingerprint, browserFingerprint, clientType, loginEventType string) (LoginResult, error) {
	if p.SessionStatus != db.CbtSessionStatusEnumActive {
		return LoginResult{}, ErrExamNotActive
	}
	if clientType == "web_fallback" && !p.RoomAllowWebFallback {
		return LoginResult{}, ErrWebFallbackDisabled
	}
	if p.ScheduledStart.Valid && time.Now().Before(p.ScheduledStart.Time) {
		return LoginResult{}, ErrExamNotStarted
	}
	if p.SubmittedAt.Valid {
		return LoginResult{}, ErrExamAlreadySubmit
	}
	if p.LockedAt.Valid {
		return LoginResult{}, ErrExamLocked
	}
	if p.ScheduledEnd.Valid && time.Now().After(p.ScheduledEnd.Time) {
		return LoginResult{}, ErrExamWindowClosed
	}

	// Device binding: if already bound, reject different device
	if p.DeviceFingerprint.Valid && strings.TrimSpace(p.DeviceFingerprint.String) != "" &&
		strings.TrimSpace(p.DeviceFingerprint.String) != deviceFingerprint {
		return LoginResult{}, ErrDeviceMismatch
	}

	// Record login and bind device
	if _, err := s.q.UpdateParticipantLogin(ctx, db.UpdateParticipantLoginParams{
		ID:                     p.ID,
		DeviceFingerprint:      pgtype.Text{String: deviceFingerprint, Valid: true},
		LoginIp:                pgtype.Text{String: req.LoginIP, Valid: true},
		ClientType:             clientType,
		BrowserFingerprintHash: safeHash(browserFingerprint),
		ClientUserAgentHash:    safeHash(req.UserAgent),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, ErrDeviceMismatch
		}
		return LoginResult{}, err
	}

	// Log login event
	_ = s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: p.ID,
		EventType:     loginEventType,
		EventData: marshalJSON(map[string]string{
			"ip":          req.LoginIP,
			"device_hash": hashString(deviceFingerprint),
			"client_type": clientType,
		}),
	})
	if clientType == "web_fallback" || clientType == "cbt_portal" {
		_ = s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
			ParticipantID: p.ID,
			EventType:     "web_fallback_used",
			EventData: marshalJSON(map[string]string{
				"client_type":              clientType,
				"browser_fingerprint_hash": safeHash(browserFingerprint),
				"user_agent_hash":          safeHash(req.UserAgent),
				"session_id":               pgUUIDString(p.SessionID),
				"room_id":                  pgUUIDString(p.RoomID),
				"reason":                   "browser_darurat_login",
			}),
		})
	}

	// Load questions for this package
	questions, err := s.q.GetExamQuestions(ctx, p.PackageID)
	if err != nil {
		return LoginResult{}, err
	}

	// Build ordered question list — use stored question_order if available.
	ordered := orderQuestionsWithDraw(questions, p.QuestionOrder, p.RandomizeQuestions, p.DrawPgCount, p.DrawEssayCount)
	optionOrder := parseOptionOrder(p.OptionOrder)

	// If no question_order yet, save the order for this participant
	if len(p.QuestionOrder) == 0 {
		ids := make([]string, len(ordered))
		for i, q := range ordered {
			ids[i] = pgUUIDString(q.ID)
		}
		orderJSON := marshalJSON(ids)
		optionOrder = ensureOptionOrder(ordered, optionOrder, p.RandomizeOptions)
		if runtimeStore, ok := s.q.(examRuntimePlanStore); ok {
			if savedPlan, err := runtimeStore.SetParticipantRuntimePlanIfEmpty(ctx, db.SetParticipantRuntimePlanIfEmptyParams{
				ID:              p.ID,
				QuestionOrder:   orderJSON,
				OptionOrder:     marshalJSON(optionOrder),
				QuestionDrawLog: marshalJSON([]string{}),
			}); err == nil && len(savedPlan.QuestionOrder) > 0 {
				ordered = orderQuestionsWithDraw(questions, savedPlan.QuestionOrder, false, 0, 0)
				optionOrder = parseOptionOrder(savedPlan.OptionOrder)
			} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return LoginResult{}, err
			}
		} else if savedOrder, err := s.q.SetParticipantQuestionOrderIfEmpty(ctx, db.SetParticipantQuestionOrderIfEmptyParams{
			ID:            p.ID,
			QuestionOrder: orderJSON,
		}); err == nil && len(savedOrder) > 0 {
			ordered = orderQuestionsWithDraw(questions, savedOrder, false, 0, 0)
		} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, err
		}
	} else if p.RandomizeOptions && len(optionOrder) == 0 {
		optionOrder = ensureOptionOrder(ordered, optionOrder, p.RandomizeOptions)
		if runtimeStore, ok := s.q.(examRuntimePlanStore); ok {
			if savedOrder, err := runtimeStore.SetParticipantOptionOrderIfEmpty(ctx, db.SetParticipantOptionOrderIfEmptyParams{
				ID:          p.ID,
				OptionOrder: marshalJSON(optionOrder),
			}); err == nil && len(savedOrder) > 0 {
				optionOrder = parseOptionOrder(savedOrder)
			} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return LoginResult{}, err
			}
		}
	}

	// Count answered questions
	answers, _ := s.q.GetParticipantAnswers(ctx, p.ID)

	// Calculate time remaining
	remaining := calcRemaining(p.ScheduledEnd, p.DurationMinutes, p.JoinedAt)

	result := LoginResult{
		ParticipantID: pgUUIDString(p.ID),
		Student: StudentInfo{
			NIS:  p.Nis,
			Nama: p.Nama,
		},
		Session: SessionInfo{
			ID:              pgUUIDString(p.SessionID),
			Title:           firstNonEmpty(p.SessionTitle, p.PackageTitle, "Sesi Ujian"),
			ScheduledStart:  p.ScheduledStart.Time.Format(time.RFC3339),
			ScheduledEnd:    p.ScheduledEnd.Time.Format(time.RFC3339),
			DurationMinutes: p.DurationMinutes,
		},
		Questions:            toExamQuestions(s.q, ctx, ordered, optionOrder),
		AnsweredCount:        len(answers),
		TotalQuestions:       len(ordered),
		TimeRemainingSeconds: remaining,
		AntiCheat:            antiCheatStateFromParticipant(p.ViolationCount, p.RiskScore, p.RiskLevel, p.LockedAt, p.LockedReason),
	}

	if p.RoomID.Valid {
		room, err := s.q.GetCbtExamRoom(ctx, p.RoomID)
		if err == nil {
			result.Room = &RoomInfo{RoomName: room.RoomName}
		}
	}

	return result, nil
}

func (s *Exam) SubmitCbtPortalAnswer(ctx context.Context, participantID, studentID, questionID pgtype.UUID, answer string) error {
	p, err := s.getCbtPortalParticipantForRuntime(ctx, participantID, studentID)
	if err != nil {
		return err
	}
	return s.SubmitAnswer(ctx, p, questionID, answer)
}

func (s *Exam) SubmitCbtPortalExam(ctx context.Context, participantID, studentID pgtype.UUID) error {
	p, err := s.getCbtPortalParticipantForRuntime(ctx, participantID, studentID)
	if err != nil {
		return err
	}
	return s.Submit(ctx, p)
}

func (s *Exam) getCbtPortalParticipantForRuntime(ctx context.Context, participantID, studentID pgtype.UUID) (db.GetParticipantByTokenRow, error) {
	store, ok := s.q.(interface {
		GetCbtPortalExamParticipant(context.Context, db.GetCbtPortalExamParticipantParams) (db.GetCbtPortalExamParticipantRow, error)
	})
	if !ok {
		return db.GetParticipantByTokenRow{}, ErrExamNotFound
	}
	row, err := store.GetCbtPortalExamParticipant(ctx, db.GetCbtPortalExamParticipantParams{ParticipantID: participantID, StudentID: studentID})
	if err != nil {
		return db.GetParticipantByTokenRow{}, ErrExamNotFound
	}
	p := cbtPortalExamParticipantToLoginRow(row)
	if !p.NisnDirectLoginEnabled || !p.StudentPortalDirectLoginEnabled || p.RequireRoomTokenForWeb {
		return db.GetParticipantByTokenRow{}, ErrCbtPortalDisabled
	}
	if p.AccessMode != "simulation" && p.AccessMode != "web_fallback" {
		return db.GetParticipantByTokenRow{}, ErrCbtPortalDisabled
	}
	if p.SessionStatus != db.CbtSessionStatusEnumActive {
		return db.GetParticipantByTokenRow{}, ErrExamNotActive
	}
	if p.SubmittedAt.Valid {
		return db.GetParticipantByTokenRow{}, ErrExamAlreadySubmit
	}
	if p.LockedAt.Valid {
		return db.GetParticipantByTokenRow{}, ErrExamLocked
	}
	if p.ScheduledEnd.Valid && time.Now().After(p.ScheduledEnd.Time) {
		return db.GetParticipantByTokenRow{}, ErrExamWindowClosed
	}
	return p, nil
}

func cbtPortalExamParticipantToLoginRow(row db.GetCbtPortalExamParticipantRow) db.GetParticipantByTokenRow {
	return db.GetParticipantByTokenRow{
		ID: row.ID, SessionID: row.SessionID, StudentID: row.StudentID, Token: row.Token,
		RoomID: row.RoomID, SeatNo: row.SeatNo, DeviceFingerprint: row.DeviceFingerprint,
		QuestionOrder: row.QuestionOrder, OptionOrder: row.OptionOrder, QuestionDrawLog: row.QuestionDrawLog,
		ClientType: row.ClientType, BrowserFingerprintHash: row.BrowserFingerprintHash, ClientUserAgentHash: row.ClientUserAgentHash,
		JoinedAt: row.JoinedAt, SubmittedAt: row.SubmittedAt, Score: row.Score,
		AppSwitchCount: row.AppSwitchCount, ScreenshotAttempt: row.ScreenshotAttempt, SuspiciousFlag: row.SuspiciousFlag,
		ViolationCount: row.ViolationCount, RiskScore: row.RiskScore, RiskLevel: row.RiskLevel,
		LockedAt: row.LockedAt, LockedReason: row.LockedReason, LastHeartbeat: row.LastHeartbeat,
		RoomToken: row.RoomToken, Nis: row.Nis, Nama: row.Nama, Gender: row.Gender,
		SessionStatus: row.SessionStatus, SessionTitle: row.SessionTitle, ScheduledStart: row.ScheduledStart,
		ScheduledEnd: row.ScheduledEnd, PackageID: row.PackageID, PackageTitle: row.PackageTitle,
		DurationMinutes: row.DurationMinutes, RandomizeQuestions: row.RandomizeQuestions,
		RandomizeOptions: row.RandomizeOptions, DrawPgCount: row.DrawPgCount, DrawEssayCount: row.DrawEssayCount,
		RoomAllowWebFallback: row.RoomAllowWebFallback, AccessMode: row.AccessMode,
		StudentPortalDirectLoginEnabled: row.StudentPortalDirectLoginEnabled,
		RequireRoomTokenForWeb:          row.RequireRoomTokenForWeb,
		NisnDirectLoginEnabled:          row.NisnDirectLoginEnabled,
	}
}

func validateParticipantRoomToken(p db.GetParticipantByTokenRow, roomToken string) error {
	roomToken = strings.TrimSpace(roomToken)
	if roomToken == "" {
		return ErrRoomTokenRequired
	}
	if strings.TrimSpace(p.RoomToken) == "" || !strings.EqualFold(strings.TrimSpace(p.RoomToken), roomToken) {
		return ErrRoomTokenMismatch
	}
	return nil
}

func (s *Exam) recordRoomTokenMismatch(ctx context.Context, p db.GetParticipantByTokenRow, loginIP string, cause error) {
	_ = s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: p.ID,
		EventType:     "exam_room_token_mismatch",
		EventData: marshalJSON(map[string]string{
			"reason":                roomTokenFailureReason(cause),
			"session_id":            pgUUIDString(p.SessionID),
			"room_id":               pgUUIDString(p.RoomID),
			"student_id":            pgUUIDString(p.StudentID),
			"ip_hash":               hashString(loginIP),
			"room_token_configured": strconv.FormatBool(strings.TrimSpace(p.RoomToken) != ""),
		}),
	})
}

func roomTokenFailureReason(err error) string {
	switch {
	case errors.Is(err, ErrRoomTokenRequired):
		return "room_token_required"
	case errors.Is(err, ErrRoomTokenMismatch):
		return "room_token_mismatch"
	default:
		return "room_token_invalid"
	}
}

func (s *Exam) Heartbeat(ctx context.Context, participantID pgtype.UUID) error {
	err := s.q.UpdateParticipantHeartbeat(ctx, participantID)
	if err != nil {
		return err
	}
	return s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "heartbeat",
	})
}

func (s *Exam) RecordClientEvent(ctx context.Context, participantID pgtype.UUID, eventType string, data map[string]any) error {
	normalized, ok := NormalizeProctorEventType(eventType)
	if !ok {
		return ErrExamInvalidTelemetry
	}
	if s.pool != nil {
		conn, err := s.pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx) //nolint:errcheck

		qtxProvider, ok := s.q.(examTxStore)
		if !ok {
			return errors.New("exam store does not support transactions")
		}
		if err := recordProctorTelemetryWithStore(ctx, qtxProvider.WithTx(tx), participantID, normalized, data, time.Now()); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	if q, ok := s.q.(examProctorTelemetryStore); ok {
		return recordProctorTelemetryWithStore(ctx, q, participantID, normalized, data, time.Now())
	}
	decision := ClassifyProctorSeverity(normalized, data)
	switch decision.Category {
	case "app_switch":
		_ = s.q.IncrementParticipantAppSwitch(ctx, participantID)
	case "screenshot":
		_ = s.q.IncrementParticipantScreenshot(ctx, participantID)
	}
	if decision.RiskDelta > 0 && decision.Severity != ProctorSeverityTechnical {
		_ = s.incrementAntiCheatRisk(ctx, participantID, decision.RiskDelta, decision.EventType)
	}
	return s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     decision.EventType,
		EventData:     marshalJSON(proctorEventDataWithDecision(sanitizeProctorEventData(data), decision, false)),
	})
}

func recordProctorTelemetryWithStore(ctx context.Context, q examProctorTelemetryStore, participantID pgtype.UUID, eventType string, data map[string]any, now time.Time) error {
	decision := ClassifyProctorSeverity(eventType, data)
	if decision.EventType == "" {
		return ErrExamInvalidTelemetry
	}
	state, err := q.GetCbtParticipantRiskForUpdate(ctx, participantID)
	if err != nil {
		return err
	}
	cleanData := sanitizeProctorEventData(data)
	eventAt := timestampFromProctorData(cleanData, now)
	dedupKey := proctorDedupKey(pgUUIDString(participantID), decision.EventType, cleanData)
	deduped := false
	if ShouldDedup(decision.EventType) && dedupKey != "" {
		_, err := q.GetRecentCbtProctorEventByDedupKey(ctx, db.GetRecentCbtProctorEventByDedupKeyParams{
			ParticipantID: participantID,
			DedupKey:      dedupKey,
			SinceAt:       pgtype.Timestamptz{Time: eventAt.Add(-DedupWindow(decision.EventType)), Valid: true},
		})
		if err == nil {
			deduped = true
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	riskDelta := decision.RiskDelta
	if deduped || decision.Severity == ProctorSeverityTechnical {
		riskDelta = 0
	}
	violationCount := state.ViolationCount
	if riskDelta > 0 && decision.Severity != ProctorSeverityTechnical {
		violationCount += 1
	}
	nextScore := state.RiskScore + riskDelta
	if nextScore > 100 {
		nextScore = 100
	}
	lockedAt := state.LockedAt
	lockedReason := state.LockedReason
	shouldLock := ShouldAutoLock(decision, ParticipantRiskState{
		ViolationCount: state.ViolationCount,
		RiskScore:      state.RiskScore,
		RiskLevel:      state.RiskLevel,
		LockedAt:       sqlNullTimeFromTimestamptz(state.LockedAt),
	})
	if riskDelta == 0 && !state.LockedAt.Valid {
		shouldLock = false
	}
	if shouldLock && !lockedAt.Valid {
		lockedAt = pgtype.Timestamptz{Time: now, Valid: true}
		lockedReason = pgtype.Text{String: decision.EventType, Valid: true}
	}
	riskLevel := RiskLevelFromScore(int(nextScore), int(violationCount), sqlNullTimeFromTimestamptz(lockedAt))
	if decision.Severity == ProctorSeverityTechnical && !state.LockedAt.Valid {
		riskLevel = RiskLevelFromScore(int(state.RiskScore), int(state.ViolationCount), sql.NullTime{})
	}
	lastLocalSaveAt := state.LastLocalSaveAt
	if parsed, ok := optionalProctorTimestamp(cleanData, "last_local_save_at"); ok {
		lastLocalSaveAt = parsed
	}
	lastSyncedAt := state.LastSyncedAt
	if parsed, ok := optionalProctorTimestamp(cleanData, "last_synced_at"); ok {
		lastSyncedAt = parsed
	}
	pendingAnswerCount := state.PendingAnswerCount
	if rawCount, ok := cleanData["pending_answer_count"]; ok {
		if count := intFromAny(rawCount); count >= 0 {
			pendingAnswerCount = int32(count)
		}
	}
	syncState := normalizeSyncState(stringFromAny(cleanData["sync_state"]), pendingAnswerCount, state.SyncState)
	appSwitchIncrement := int32(0)
	if decision.Category == "app_switch" && !deduped {
		appSwitchIncrement = 1
	}
	screenshotIncrement := int32(0)
	if decision.Category == "screenshot" && !deduped {
		screenshotIncrement = 1
	}
	if _, err := q.UpdateCbtParticipantProctorRisk(ctx, db.UpdateCbtParticipantProctorRiskParams{
		ViolationCount:      violationCount,
		RiskScore:           nextScore,
		RiskLevel:           riskLevel,
		LockedAt:            lockedAt,
		LockedReason:        lockedReason,
		SuspiciousFlag:      riskDelta > 0 || decision.Severity == ProctorSeverityCritical || decision.Severity == ProctorSeverityMedium,
		AppSwitchIncrement:  appSwitchIncrement,
		ScreenshotIncrement: screenshotIncrement,
		LastLocalSaveAt:     lastLocalSaveAt,
		LastSyncedAt:        lastSyncedAt,
		PendingAnswerCount:  pendingAnswerCount,
		SyncState:           syncState,
		ID:                  participantID,
	}); err != nil {
		return err
	}
	eventData := proctorEventDataWithDecision(cleanData, decision, deduped)
	eventData["session_id"] = pgUUIDString(state.SessionID)
	eventData["room_id"] = pgUUIDString(state.RoomID)
	eventData["risk_level_after"] = riskLevel
	eventData["risk_score_after"] = nextScore
	eventData["violation_count_after"] = violationCount
	_, err = q.CreateCbtParticipantProctorEvent(ctx, db.CreateCbtParticipantProctorEventParams{
		ParticipantID:         participantID,
		EventType:             decision.EventType,
		EventData:             marshalJSON(eventData),
		Severity:              string(decision.Severity),
		Category:              decision.Category,
		RiskDelta:             riskDelta,
		DedupKey:              dedupKey,
		CorrelationID:         stringFromAny(cleanData["correlation_id"]),
		OriginalEventAt:       pgtype.Timestamptz{Time: eventAt, Valid: true},
		RequiresNote:          decision.RequiresNote,
		ActorUsernameSnapshot: "",
		RequestID:             "",
		SourceIp:              "",
	})
	return err
}

func sanitizeProctorEventData(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}
	var cleaned map[string]any
	raw, err := json.Marshal(data)
	if err == nil {
		_ = json.Unmarshal(raw, &cleaned)
	}
	if cleaned == nil {
		cleaned = map[string]any{}
	}
	for _, key := range []string{"severity", "risk_delta", "risk_score", "risk_level", "lock_eligible", "locked_at", "locked_reason"} {
		delete(cleaned, key)
	}
	return cleaned
}

func proctorEventDataWithDecision(data map[string]any, decision SeverityDecision, deduped bool) map[string]any {
	payload := map[string]any{}
	for key, value := range data {
		payload[key] = value
	}
	payload["event_type"] = decision.EventType
	payload["severity"] = string(decision.Severity)
	payload["category"] = decision.Category
	payload["risk_delta"] = decision.RiskDelta
	payload["label_id"] = decision.LabelID
	payload["message_id"] = decision.MessageID
	payload["audio_key"] = decision.AudioKey
	payload["requires_note"] = decision.RequiresNote
	payload["deduped"] = deduped
	return payload
}

func timestampFromProctorData(data map[string]any, fallback time.Time) time.Time {
	for _, key := range []string{"original_event_at", "event_at", "created_at"} {
		if parsed, ok := optionalProctorTimestamp(data, key); ok {
			return parsed.Time
		}
	}
	return fallback
}

func optionalProctorTimestamp(data map[string]any, key string) (pgtype.Timestamptz, bool) {
	raw := strings.TrimSpace(stringFromAny(data[key]))
	if raw == "" {
		return pgtype.Timestamptz{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return pgtype.Timestamptz{}, false
	}
	return pgtype.Timestamptz{Time: parsed, Valid: true}, true
}

func normalizeSyncState(raw string, pendingAnswerCount int32, fallback string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "synced", "sinkron":
		return "synced"
	case "pending", "belum_sinkron":
		return "pending"
	case "failed", "tertahan":
		return "failed"
	case "unknown", "":
		if pendingAnswerCount > 0 {
			return "pending"
		}
		if fallback != "" {
			return fallback
		}
		return "unknown"
	default:
		if pendingAnswerCount > 0 {
			return "pending"
		}
		return "unknown"
	}
}

func sqlNullTimeFromTimestamptz(value pgtype.Timestamptz) sql.NullTime {
	return sql.NullTime{Time: value.Time, Valid: value.Valid}
}

func (s *Exam) SubmitAnswer(ctx context.Context, p db.GetParticipantByTokenRow, questionID pgtype.UUID, answer string) error {
	if p.SubmittedAt.Valid {
		return ErrExamAlreadySubmit
	}
	if p.LockedAt.Valid {
		return ErrExamLocked
	}
	if examWindowNotStarted(p.ScheduledStart, time.Now()) {
		return ErrExamNotStarted
	}
	if examWindowClosed(p.ScheduledEnd, p.DurationMinutes, p.JoinedAt, time.Now()) {
		return ErrExamWindowClosed
	}
	belongs, err := s.q.QuestionBelongsToParticipantPackage(ctx, db.QuestionBelongsToParticipantPackageParams{
		ID:         p.ID,
		QuestionID: questionID,
	})
	if err != nil {
		return err
	}
	if !belongs {
		return ErrExamQuestionScope
	}
	storedAnswer := canonicalizeRandomizedAnswer(p.OptionOrder, questionID, answer)
	rows, err := s.q.UpsertStudentAnswer(ctx, db.UpsertStudentAnswerParams{
		ParticipantID: p.ID,
		QuestionID:    questionID,
		Answer:        storedAnswer,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrExamAlreadySubmit
	}
	return s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: p.ID,
		EventType:     "answer",
		EventData: marshalJSON(map[string]string{
			"question_id":   pgUUIDString(questionID),
			"answer_length": strconv.Itoa(len([]rune(answer))),
			"answer_hash":   hashString(answer),
			"status":        "recorded",
		}),
	})
}

type StatusResult struct {
	AnsweredCount        int                  `json:"answered_count"`
	TotalQuestions       int                  `json:"total_questions"`
	SubmittedAt          string               `json:"submitted_at,omitempty"`
	TimeRemainingSeconds int64                `json:"time_remaining_seconds"`
	IsSubmitted          bool                 `json:"is_submitted"`
	AntiCheat            AntiCheatState       `json:"anti_cheat"`
	Commands             []ParticipantCommand `json:"commands"`
}

func (s *Exam) GetStatus(ctx context.Context, p db.GetParticipantByTokenRow) (StatusResult, error) {
	questions, err := s.q.GetExamQuestions(ctx, p.PackageID)
	if err != nil {
		return StatusResult{}, err
	}
	answers, _ := s.q.GetParticipantAnswers(ctx, p.ID)

	result := StatusResult{
		AnsweredCount:        len(answers),
		TotalQuestions:       len(questions),
		TimeRemainingSeconds: calcRemaining(p.ScheduledEnd, p.DurationMinutes, p.JoinedAt),
		IsSubmitted:          p.SubmittedAt.Valid,
		AntiCheat:            antiCheatStateFromParticipant(p.ViolationCount, p.RiskScore, p.RiskLevel, p.LockedAt, p.LockedReason),
	}
	if p.SubmittedAt.Valid {
		result.SubmittedAt = p.SubmittedAt.Time.Format(time.RFC3339)
	}
	commands, err := s.ListPendingCommands(ctx, p.ID)
	if err != nil {
		return StatusResult{}, err
	}
	result.Commands = commands
	return result, nil
}

func (s *Exam) ListPendingCommands(ctx context.Context, participantID pgtype.UUID) ([]ParticipantCommand, error) {
	rows, err := s.q.ListPendingParticipantCommands(ctx, participantID)
	if err != nil {
		return nil, err
	}
	commands := make([]ParticipantCommand, 0, len(rows))
	for _, row := range rows {
		commands = append(commands, participantCommandFromEvent(row))
	}
	return commands, nil
}

func (s *Exam) AcknowledgeCommand(ctx context.Context, participantID pgtype.UUID, commandID, status string) error {
	commandID = strings.TrimSpace(commandID)
	if commandID == "" {
		return errors.New("command id required")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "seen"
	}
	return s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "participant_command_ack",
		EventData: marshalJSON(map[string]string{
			"command_id": commandID,
			"status":     status,
		}),
	})
}

func (s *Exam) Submit(ctx context.Context, p db.GetParticipantByTokenRow) error {
	if p.SubmittedAt.Valid {
		return ErrExamAlreadySubmit
	}
	if p.LockedAt.Valid {
		return ErrExamLocked
	}
	if examWindowNotStarted(p.ScheduledStart, time.Now()) {
		return ErrExamNotStarted
	}
	if examWindowClosed(p.ScheduledEnd, p.DurationMinutes, p.JoinedAt, time.Now()) {
		return ErrExamWindowClosed
	}
	if s.pool != nil {
		conn, err := s.pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx) //nolint:errcheck

		qtxProvider, ok := s.q.(examTxStore)
		if !ok {
			return errors.New("exam store does not support transactions")
		}
		if err := submitExamWithStore(ctx, qtxProvider.WithTx(tx), p.ID); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	return submitExamWithStore(ctx, s.q, p.ID)
}

func submitExamWithStore(ctx context.Context, q examStore, participantID pgtype.UUID) error {
	row, err := q.SubmitParticipantExam(ctx, participantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrExamAlreadySubmit
		}
		return err
	}
	if err := q.UpdateParticipantAnswerCorrectness(ctx, participantID); err != nil {
		return err
	}
	return q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "submit",
		EventData:     marshalJSON(map[string]string{"submitted_at": row.SubmittedAt.Time.Format(time.RFC3339)}),
	})
}

func (s *Exam) incrementAntiCheatRisk(ctx context.Context, participantID pgtype.UUID, riskWeight int32, reason string) error {
	if riskWeight <= 0 {
		riskWeight = 20
	}
	if reason == "" {
		reason = "anti_cheat_violation"
	}
	_, err := s.q.IncrementParticipantAntiCheatViolation(ctx, db.IncrementParticipantAntiCheatViolationParams{
		ID:             participantID,
		RiskScore:      riskWeight,
		ViolationCount: 3,
		LockedReason:   pgtype.Text{String: reason, Valid: true},
	})
	return err
}

func antiCheatRiskWeight(data map[string]any) int32 {
	severity, _ := data["severity"].(string)
	reason := antiCheatReason(data)
	switch {
	case severity == "critical" || reason == "anti_cheat_local_lock":
		return 80
	case reason == "split_screen_detected" || reason == "picture_in_picture_detected":
		return 25
	case reason == "window_focus_lost" || reason == "app_backgrounded":
		return 30
	default:
		return 25
	}
}

func antiCheatReason(data map[string]any) string {
	if data == nil {
		return "anti_cheat_violation"
	}
	if reason, ok := data["reason"].(string); ok && strings.TrimSpace(reason) != "" {
		return strings.TrimSpace(reason)
	}
	return "anti_cheat_violation"
}

func antiCheatStateFromParticipant(violationCount, riskScore int32, riskLevel string, lockedAt pgtype.Timestamptz, lockedReason pgtype.Text) AntiCheatState {
	state := AntiCheatState{
		ViolationCount: int(violationCount),
		RiskScore:      int(riskScore),
		RiskLevel:      firstNonEmpty(riskLevel, "normal"),
		Locked:         lockedAt.Valid,
	}
	if lockedAt.Valid {
		state.LockedAt = lockedAt.Time.Format(time.RFC3339)
	}
	if lockedReason.Valid {
		state.LockedReason = lockedReason.String
	}
	return state
}

func participantCommandFromEvent(row db.ListPendingParticipantCommandsRow) ParticipantCommand {
	payload := map[string]any{}
	if len(row.EventData) > 0 {
		_ = json.Unmarshal(row.EventData, &payload)
	}
	commandType := stringFromMap(payload, "command_type")
	if commandType == "" {
		commandType = stringFromMap(payload, "type")
	}
	if commandType == "" {
		commandType = "warning_message"
	}
	message := stringFromMap(payload, "message")
	command := ParticipantCommand{
		ID:       pgUUIDString(row.ID),
		Type:     commandType,
		Message:  message,
		Severity: stringFromMap(payload, "severity"),
		IssuedAt: "",
		Actor:    stringFromMap(payload, "actor"),
		Payload:  payload,
		RawEvent: row.EventType,
	}
	if row.CreatedAt.Valid {
		command.IssuedAt = row.CreatedAt.Time.Format(time.RFC3339)
	}
	switch command.Type {
	case "reconnect":
		command.LegacyLabel = "Login ulang"
	case "unlock_notice":
		command.LegacyLabel = "Akses dibuka"
	default:
		command.LegacyLabel = "Peringatan"
	}
	if len(command.Payload) == 0 {
		command.Payload = nil
	}
	return command
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, ok := values[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

// --- helpers ---

func calcRemaining(end pgtype.Timestamptz, durationMin int32, joinedAt pgtype.Timestamptz) int64 {
	deadline, ok := examDeadline(end, durationMin, joinedAt)
	if !ok {
		return 0
	}
	remaining := time.Until(deadline)
	if remaining < 0 {
		return 0
	}
	return int64(remaining.Seconds())
}

func examWindowClosed(end pgtype.Timestamptz, durationMin int32, joinedAt pgtype.Timestamptz, now time.Time) bool {
	deadline, ok := examDeadline(end, durationMin, joinedAt)
	return ok && now.After(deadline)
}

func examWindowNotStarted(start pgtype.Timestamptz, now time.Time) bool {
	return start.Valid && now.Before(start.Time)
}

func examDeadline(end pgtype.Timestamptz, durationMin int32, joinedAt pgtype.Timestamptz) (time.Time, bool) {
	if durationMin > 0 && joinedAt.Valid {
		byDuration := joinedAt.Time.Add(time.Duration(durationMin) * time.Minute)
		if end.Valid && end.Time.Before(byDuration) {
			return end.Time, true
		}
		return byDuration, true
	} else if end.Valid {
		return end.Time, true
	}
	return time.Time{}, false
}

func orderQuestionsWithDraw(questions []db.GetExamQuestionsRow, orderJSON []byte, randomize bool, drawPG, drawEssay int32) []db.GetExamQuestionsRow {
	if len(orderJSON) == 0 {
		return defaultQuestionOrder(applyQuestionDraw(questions, drawPG, drawEssay), randomize)
	}

	// Reorder by stored question_order. Treat an empty JSON array as missing order
	// so a reset/retake cannot accidentally produce an empty exam package.
	var ids []string
	if err := json.Unmarshal(orderJSON, &ids); err != nil {
		return applyQuestionDraw(questions, drawPG, drawEssay)
	}
	if len(ids) == 0 {
		return defaultQuestionOrder(applyQuestionDraw(questions, drawPG, drawEssay), randomize)
	}
	idMap := make(map[string]db.GetExamQuestionsRow, len(questions))
	for _, q := range questions {
		idMap[pgUUIDString(q.ID)] = q
	}
	ordered := make([]db.GetExamQuestionsRow, 0, len(ids))
	for _, id := range ids {
		if q, ok := idMap[id]; ok {
			ordered = append(ordered, q)
		}
	}
	if len(ordered) == 0 {
		return defaultQuestionOrder(applyQuestionDraw(questions, drawPG, drawEssay), randomize)
	}
	return ordered
}

func orderQuestions(questions []db.GetExamQuestionsRow, orderJSON []byte, randomize bool) []db.GetExamQuestionsRow {
	return orderQuestionsWithDraw(questions, orderJSON, randomize, 0, 0)
}

func applyQuestionDraw(questions []db.GetExamQuestionsRow, drawPG, drawEssay int32) []db.GetExamQuestionsRow {
	if drawPG <= 0 && drawEssay <= 0 {
		return questions
	}
	objective := make([]db.GetExamQuestionsRow, 0, len(questions))
	essay := make([]db.GetExamQuestionsRow, 0, len(questions))
	other := make([]db.GetExamQuestionsRow, 0)
	for _, question := range questions {
		switch question.QuestionType {
		case "essay":
			essay = append(essay, question)
		case "multiple_choice", "multiple_answer", "true_false", "agree_disagree", "matching", "ordering", "short_answer":
			objective = append(objective, question)
		default:
			other = append(other, question)
		}
	}
	selected := make([]db.GetExamQuestionsRow, 0, len(questions))
	selected = append(selected, drawQuestionGroup(objective, drawPG)...)
	selected = append(selected, drawQuestionGroup(essay, drawEssay)...)
	if drawPG <= 0 && len(objective) == 0 {
		selected = append(selected, other...)
	}
	if len(selected) == 0 {
		return questions
	}
	position := make(map[string]int, len(questions))
	for i, question := range questions {
		position[pgUUIDString(question.ID)] = i
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return position[pgUUIDString(selected[i].ID)] < position[pgUUIDString(selected[j].ID)]
	})
	return selected
}

func drawQuestionGroup(questions []db.GetExamQuestionsRow, drawCount int32) []db.GetExamQuestionsRow {
	if drawCount <= 0 || int(drawCount) >= len(questions) {
		return questions
	}
	indices := shuffleInts(len(questions))
	out := make([]db.GetExamQuestionsRow, 0, drawCount)
	for i := 0; i < int(drawCount) && i < len(indices); i++ {
		out = append(out, questions[indices[i]])
	}
	return out
}

func defaultQuestionOrder(questions []db.GetExamQuestionsRow, randomize bool) []db.GetExamQuestionsRow {
	if !randomize {
		return questions
	}
	indices := shuffleInts(len(questions))
	ordered := make([]db.GetExamQuestionsRow, len(questions))
	for i, idx := range indices {
		ordered[i] = questions[idx]
	}
	return ordered
}

func shuffleInts(n int) []int {
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	for i := n - 1; i > 0; i-- {
		value, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return idx
		}
		j := int(value.Int64())
		idx[i], idx[j] = idx[j], idx[i]
	}
	return idx
}

func parseOptionOrder(raw []byte) map[string][]string {
	if len(raw) == 0 {
		return map[string][]string{}
	}
	var order map[string][]string
	if err := json.Unmarshal(raw, &order); err != nil || order == nil {
		return map[string][]string{}
	}
	return order
}

func ensureOptionOrder(rows []db.GetExamQuestionsRow, existing map[string][]string, randomize bool) map[string][]string {
	if !randomize {
		return map[string][]string{}
	}
	if existing == nil {
		existing = map[string][]string{}
	}
	for _, row := range rows {
		if !supportsOptionRandomization(row.QuestionType) {
			continue
		}
		id := pgUUIDString(row.ID)
		if len(existing[id]) > 0 {
			continue
		}
		labels := optionLabelsForQuestion(row)
		if len(labels) <= 1 {
			continue
		}
		indices := shuffleInts(len(labels))
		shuffled := make([]string, 0, len(labels))
		for _, idx := range indices {
			shuffled = append(shuffled, labels[idx])
		}
		existing[id] = shuffled
	}
	return existing
}

func supportsOptionRandomization(questionType string) bool {
	return questionType == "multiple_choice" || questionType == "multiple_answer"
}

func optionLabelsForQuestion(row db.GetExamQuestionsRow) []string {
	options := decodeQuestionOptions(row.Options)
	if len(options) == 0 {
		options = legacyOptions(row.OptionA, row.OptionB, row.OptionC, row.OptionD, row.OptionE, row.QuestionType)
	}
	labels := make([]string, 0, len(options))
	for _, option := range options {
		label := strings.ToUpper(strings.TrimSpace(option.Label))
		if label != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

func toExamQuestions(q examStore, ctx context.Context, rows []db.GetExamQuestionsRow, optionOrders ...map[string][]string) []ExamQuestion {
	optionOrder := map[string][]string{}
	if len(optionOrders) > 0 && optionOrders[0] != nil {
		optionOrder = optionOrders[0]
	}
	out := make([]ExamQuestion, len(rows))
	for i, r := range rows {
		r = applyExamOptionOrder(r, optionOrder[pgUUIDString(r.ID)])
		opts := json.RawMessage(r.Options)
		if len(opts) == 0 {
			opts = json.RawMessage("[]")
		}
		stemMediaURL, stimulusMediaURL, stemAudioURL, stimulusAudioURL := resolveExamQuestionAssets(q, ctx, r.ID)
		out[i] = ExamQuestion{
			ID:               pgUUIDString(r.ID),
			Code:             r.Code,
			QuestionText:     r.QuestionText,
			QuestionType:     r.QuestionType,
			Options:          opts,
			StemHTML:         r.StemHtml,
			StimulusHTML:     r.StimulusHtml,
			StemMediaURL:     stemMediaURL,
			StimulusMediaURL: stimulusMediaURL,
			StemAudioURL:     stemAudioURL,
			StimulusAudioURL: stimulusAudioURL,
			OptionA:          r.OptionA,
			OptionB:          r.OptionB,
			OptionC:          r.OptionC,
			OptionD:          r.OptionD,
			OptionE:          r.OptionE,
		}
	}
	return out
}

func applyExamOptionOrder(row db.GetExamQuestionsRow, order []string) db.GetExamQuestionsRow {
	if len(order) == 0 || !supportsOptionRandomization(row.QuestionType) {
		return row
	}
	options := decodeQuestionOptions(row.Options)
	if len(options) == 0 {
		options = legacyOptions(row.OptionA, row.OptionB, row.OptionC, row.OptionD, row.OptionE, row.QuestionType)
	}
	byLabel := make(map[string]QuestionOption, len(options))
	for _, option := range options {
		byLabel[strings.ToUpper(strings.TrimSpace(option.Label))] = option
	}
	ordered := make([]QuestionOption, 0, len(order))
	for i, canonicalLabel := range order {
		option, ok := byLabel[strings.ToUpper(strings.TrimSpace(canonicalLabel))]
		if !ok {
			continue
		}
		option.Label = labelFromIndex(i)
		ordered = append(ordered, option)
	}
	if len(ordered) == 0 {
		return row
	}
	if raw, err := EncodeQuestionOptions(ordered); err == nil {
		row.Options = raw
	}
	row.OptionA, row.OptionB, row.OptionC, row.OptionD, row.OptionE = legacyOptionColumns(ordered)
	return row
}

func canonicalizeRandomizedAnswer(optionOrderJSON []byte, questionID pgtype.UUID, answer string) string {
	order := parseOptionOrder(optionOrderJSON)[pgUUIDString(questionID)]
	if len(order) == 0 {
		return answer
	}
	parts := strings.Split(answer, ",")
	out := make([]string, 0, len(parts))
	changed := false
	for _, part := range parts {
		label := strings.ToUpper(strings.TrimSpace(part))
		if len(label) == 1 {
			idx := int(label[0] - 'A')
			if idx >= 0 && idx < len(order) {
				out = append(out, order[idx])
				changed = true
				continue
			}
		}
		if strings.TrimSpace(part) != "" {
			out = append(out, strings.TrimSpace(part))
		}
	}
	if !changed {
		return answer
	}
	return strings.Join(out, ",")
}

func labelFromIndex(idx int) string {
	if idx < 0 || idx >= 26 {
		return ""
	}
	return string(rune('A' + idx))
}

func resolveExamQuestionAssets(q examStore, ctx context.Context, questionID pgtype.UUID) (string, string, string, string) {
	if q == nil {
		return "", "", "", ""
	}
	assets, err := q.ListCbtQuestionAssetsByQuestion(ctx, questionID)
	if err != nil {
		return "", "", "", ""
	}

	var stemMediaURL, stimulusMediaURL, stemAudioURL, stimulusAudioURL string
	for _, asset := range assets {
		url := "/api/cbt/assets/" + pgUUIDString(asset.ID) + "/file"
		isStimulus := strings.EqualFold(asset.Purpose, "stimulus")
		switch {
		case strings.HasPrefix(asset.MimeType, "image/"):
			if isStimulus && stimulusMediaURL == "" {
				stimulusMediaURL = url
				continue
			}
			if stemMediaURL == "" {
				stemMediaURL = url
				continue
			}
			if stimulusMediaURL == "" {
				stimulusMediaURL = url
			}
		case strings.HasPrefix(asset.MimeType, "audio/"):
			if isStimulus && stimulusAudioURL == "" {
				stimulusAudioURL = url
				continue
			}
			if stemAudioURL == "" {
				stemAudioURL = url
				continue
			}
			if stimulusAudioURL == "" {
				stimulusAudioURL = url
			}
		}
	}

	return stemMediaURL, stimulusMediaURL, stemAudioURL, stimulusAudioURL
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func marshalJSON(v any) []byte {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func safeHash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return hashString(value)
}

func normalizeExamClientType(raw string) string {
	clean := strings.TrimSpace(strings.ToLower(raw))
	clean = strings.ReplaceAll(clean, "-", "_")
	clean = strings.ReplaceAll(clean, " ", "_")
	switch clean {
	case "android", "flutter_android", "native_android":
		return "android"
	case "windows", "flutter_windows", "native_windows", "desktop":
		return "windows"
	case "web", "browser", "browser_darurat", "web_fallback":
		return "web_fallback"
	default:
		return "unknown"
	}
}
