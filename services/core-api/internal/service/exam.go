package service

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var (
	ErrExamNotFound      = errors.New("exam token not found")
	ErrExamNotActive     = errors.New("exam session is not active")
	ErrExamAlreadySubmit = errors.New("exam already submitted")
	ErrExamWindowClosed  = errors.New("exam window has closed")
	ErrExamNotStarted    = errors.New("exam session has not started")
	ErrDeviceMismatch    = errors.New("token already bound to another device")
	ErrDeviceRequired    = errors.New("device fingerprint required")
	ErrExamQuestionScope = errors.New("question is not part of participant exam")
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
	QuestionBelongsToParticipantPackage(ctx context.Context, arg db.QuestionBelongsToParticipantPackageParams) (bool, error)
	UpsertStudentAnswer(ctx context.Context, arg db.UpsertStudentAnswerParams) (int64, error)
	UpdateParticipantAnswerCorrectness(ctx context.Context, participantID pgtype.UUID) error
	SubmitParticipantExam(ctx context.Context, id pgtype.UUID) (db.SubmitParticipantExamRow, error)
	ListCbtQuestionAssetsByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error)
}

type examTxStore interface {
	WithTx(tx pgx.Tx) *db.Queries
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

func (s *Exam) Login(ctx context.Context, token, deviceFingerprint, loginIP string) (LoginResult, error) {
	deviceFingerprint = strings.TrimSpace(deviceFingerprint)
	if deviceFingerprint == "" {
		return LoginResult{}, ErrDeviceRequired
	}
	p, err := s.q.GetParticipantByToken(ctx, token)
	if err != nil {
		return LoginResult{}, ErrExamNotFound
	}
	if p.SessionStatus != db.CbtSessionStatusEnumActive {
		return LoginResult{}, ErrExamNotActive
	}
	if p.ScheduledStart.Valid && time.Now().Before(p.ScheduledStart.Time) {
		return LoginResult{}, ErrExamNotStarted
	}
	if p.SubmittedAt.Valid {
		return LoginResult{}, ErrExamAlreadySubmit
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
		ID:                p.ID,
		DeviceFingerprint: pgtype.Text{String: deviceFingerprint, Valid: true},
		LoginIp:           pgtype.Text{String: loginIP, Valid: true},
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, ErrDeviceMismatch
		}
		return LoginResult{}, err
	}

	// Log login event
	_ = s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: p.ID,
		EventType:     "login",
		EventData:     marshalJSON(map[string]string{"ip": loginIP, "device_hash": hashString(deviceFingerprint)}),
	})

	// Load questions for this package
	questions, err := s.q.GetExamQuestions(ctx, p.PackageID)
	if err != nil {
		return LoginResult{}, err
	}

	// Build ordered question list — use stored question_order if available.
	ordered := orderQuestions(questions, p.QuestionOrder, p.RandomizeQuestions)

	// If no question_order yet, save the order for this participant
	if len(p.QuestionOrder) == 0 {
		ids := make([]string, len(ordered))
		for i, q := range ordered {
			ids[i] = pgUUIDString(q.ID)
		}
		orderJSON, err := json.Marshal(ids)
		if err != nil {
			return LoginResult{}, err
		}
		if savedOrder, err := s.q.SetParticipantQuestionOrderIfEmpty(ctx, db.SetParticipantQuestionOrderIfEmptyParams{
			ID:            p.ID,
			QuestionOrder: orderJSON,
		}); err == nil && len(savedOrder) > 0 {
			ordered = orderQuestions(questions, savedOrder, p.RandomizeQuestions)
		} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, err
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
		Questions:            toExamQuestions(s.q, ctx, ordered),
		AnsweredCount:        len(answers),
		TotalQuestions:       len(ordered),
		TimeRemainingSeconds: remaining,
	}

	if p.RoomID.Valid {
		room, err := s.q.GetCbtExamRoom(ctx, p.RoomID)
		if err == nil {
			result.Room = &RoomInfo{RoomName: room.RoomName}
		}
	}

	return result, nil
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
	switch eventType {
	case "app_switch":
		_ = s.q.IncrementParticipantAppSwitch(ctx, participantID)
	case "screenshot_attempt":
		_ = s.q.IncrementParticipantScreenshot(ctx, participantID)
	}
	return s.q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     eventType,
		EventData:     marshalJSON(data),
	})
}

func (s *Exam) SubmitAnswer(ctx context.Context, p db.GetParticipantByTokenRow, questionID pgtype.UUID, answer string) error {
	if p.SubmittedAt.Valid {
		return ErrExamAlreadySubmit
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
	rows, err := s.q.UpsertStudentAnswer(ctx, db.UpsertStudentAnswerParams{
		ParticipantID: p.ID,
		QuestionID:    questionID,
		Answer:        answer,
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
	AnsweredCount        int    `json:"answered_count"`
	TotalQuestions       int    `json:"total_questions"`
	SubmittedAt          string `json:"submitted_at,omitempty"`
	TimeRemainingSeconds int64  `json:"time_remaining_seconds"`
	IsSubmitted          bool   `json:"is_submitted"`
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
	}
	if p.SubmittedAt.Valid {
		result.SubmittedAt = p.SubmittedAt.Time.Format(time.RFC3339)
	}
	return result, nil
}

func (s *Exam) Submit(ctx context.Context, p db.GetParticipantByTokenRow) error {
	if p.SubmittedAt.Valid {
		return ErrExamAlreadySubmit
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

func orderQuestions(questions []db.GetExamQuestionsRow, orderJSON []byte, randomize bool) []db.GetExamQuestionsRow {
	if len(orderJSON) == 0 {
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

	// Reorder by stored question_order
	var ids []string
	if err := json.Unmarshal(orderJSON, &ids); err != nil {
		return questions
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

func toExamQuestions(q examStore, ctx context.Context, rows []db.GetExamQuestionsRow) []ExamQuestion {
	out := make([]ExamQuestion, len(rows))
	for i, r := range rows {
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
