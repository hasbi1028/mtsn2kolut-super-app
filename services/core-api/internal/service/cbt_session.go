package service

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtSession struct {
	q    cbtSessionStore
	pool *pgxpool.Pool
}

type CbtRoomAssignmentInput struct {
	MixPolicy       string                  `json:"mix_policy"`
	AssignmentMode  string                  `json:"assignment_mode"`
	AllowCrossGrade bool                    `json:"allow_cross_grade"`
	IsSpecialEvent  bool                    `json:"is_special_event"`
	Assignments     []CbtRoomAssignmentSeat `json:"assignments,omitempty"`
}

type CbtRoomAssignmentPreview struct {
	Summary     CbtRoomAssignmentSummary `json:"summary"`
	Rooms       []CbtRoomAssignmentRoom  `json:"rooms"`
	Warnings    []string                 `json:"warnings"`
	Assignments []CbtRoomAssignmentSeat  `json:"assignments,omitempty"`
}

type CbtRoomAssignmentSummary struct {
	ParticipantCount int    `json:"participant_count"`
	RoomCount        int    `json:"room_count"`
	CapacityTotal    int    `json:"capacity_total"`
	AssignedCount    int    `json:"assigned_count"`
	UnassignedCount  int    `json:"unassigned_count"`
	MixPolicy        string `json:"mix_policy"`
	AssignmentMode   string `json:"assignment_mode"`
	AllowCrossGrade  bool   `json:"allow_cross_grade"`
	IsSpecialEvent   bool   `json:"is_special_event"`
}

type CbtRoomAssignmentRoom struct {
	RoomID           string         `json:"room_id"`
	RoomName         string         `json:"room_name"`
	Capacity         int            `json:"capacity"`
	ParticipantCount int            `json:"participant_count"`
	Levels           map[string]int `json:"levels"`
	Classes          []string       `json:"classes"`
}

type CbtRoomAssignmentSeat struct {
	ParticipantID    string `json:"participant_id"`
	ParticipantName  string `json:"participant_name,omitempty"`
	ParticipantNis   string `json:"participant_nis,omitempty"`
	ParticipantClass string `json:"participant_class,omitempty"`
	RoomID           string `json:"room_id"`
	RoomName         string `json:"room_name,omitempty"`
	SeatNo           int32  `json:"seat_no"`
}

type UpdateCbtSessionScheduleResult struct {
	Before  db.GetCbtExamSessionRow `json:"before"`
	Session db.CbtExamSession       `json:"session"`
}

type cbtSessionStore interface {
	ListCbtExamSessions(ctx context.Context) ([]db.ListCbtExamSessionsRow, error)
	GetCbtExamSession(ctx context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error)
	ListCbtPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, error)
	GetCbtPackageQuestionQuality(ctx context.Context, id pgtype.UUID) (db.GetCbtPackageQuestionQualityRow, error)
	CreateCbtExamSession(ctx context.Context, arg db.CreateCbtExamSessionParams) (db.CbtExamSession, error)
	UpdateCbtExamSessionStatus(ctx context.Context, arg db.UpdateCbtExamSessionStatusParams) (db.CbtExamSession, error)
	UpdateCbtExamSessionSchedule(ctx context.Context, arg db.UpdateCbtExamSessionScheduleParams) (db.CbtExamSession, error)
	ListEntityAuditLogs(ctx context.Context, arg db.ListEntityAuditLogsParams) ([]db.ListEntityAuditLogsRow, error)
	DeleteCbtExamSession(ctx context.Context, id pgtype.UUID) (int64, error)
	ListCbtExamParticipants(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error)
	ListCbtExamParticipantsByTeacher(ctx context.Context, arg db.ListCbtExamParticipantsByTeacherParams) ([]db.ListCbtExamParticipantsByTeacherRow, error)
	EnrollClassToSession(ctx context.Context, arg db.EnrollClassToSessionParams) error
	EnrollGradeToSession(ctx context.Context, arg db.EnrollGradeToSessionParams) error
	EnrollSchoolToSession(ctx context.Context, sessionID pgtype.UUID) error
	GenerateTokensForSession(ctx context.Context, sessionID pgtype.UUID) error
	RegenerateParticipantToken(ctx context.Context, id pgtype.UUID) (db.RegenerateParticipantTokenRow, error)
	ListCbtExamRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error)
	CreateCbtExamRoom(ctx context.Context, arg db.CreateCbtExamRoomParams) (db.CbtExamRoom, error)
	DeleteCbtExamRoom(ctx context.Context, id pgtype.UUID) error
	GetCbtExamRoomSetupContext(ctx context.Context, id pgtype.UUID) (db.GetCbtExamRoomSetupContextRow, error)
	GetCbtExamRoomSetupUsage(ctx context.Context, id pgtype.UUID) (db.GetCbtExamRoomSetupUsageRow, error)
	GetSchoolRoom(ctx context.Context, id pgtype.UUID) (db.SchoolRoom, error)
	GetCbtRoomProctorDashboard(ctx context.Context, id pgtype.UUID) (db.GetCbtRoomProctorDashboardRow, error)
	GetCbtRoomHandover(ctx context.Context, id pgtype.UUID) (db.GetCbtRoomHandoverRow, error)
	GetCbtSessionOperationalRecap(ctx context.Context, id pgtype.UUID) (db.GetCbtSessionOperationalRecapRow, error)
	ListCbtSessionRoomOperationalRecap(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtSessionRoomOperationalRecapRow, error)
	UpsertCbtRoomHandover(ctx context.Context, arg db.UpsertCbtRoomHandoverParams) (db.CbtRoomHandover, error)
	LockCbtRoomHandover(ctx context.Context, arg db.LockCbtRoomHandoverParams) (db.CbtRoomHandover, error)
	ListCbtProctorRooms(ctx context.Context, arg db.ListCbtProctorRoomsParams) ([]db.ListCbtProctorRoomsRow, error)
	ListCbtRoomProctors(ctx context.Context, examRoomID pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error)
	DeleteCbtRoomProctorsByRoom(ctx context.Context, examRoomID pgtype.UUID) error
	CreateCbtRoomProctor(ctx context.Context, arg db.CreateCbtRoomProctorParams) (db.CbtRoomProctor, error)
	GetCbtSessionRoomReadiness(ctx context.Context, targetSessionID pgtype.UUID) (db.GetCbtSessionRoomReadinessRow, error)
	HasSessionRoomProctor(ctx context.Context, arg db.HasSessionRoomProctorParams) (bool, error)
	HasSessionRoomParticipant(ctx context.Context, arg db.HasSessionRoomParticipantParams) (bool, error)
	AssignParticipantSeat(ctx context.Context, arg db.AssignParticipantSeatParams) error
	ClearParticipantRooms(ctx context.Context, sessionID pgtype.UUID) error
	ClearParticipantSeatsForSession(ctx context.Context, sessionID pgtype.UUID) error
	HasOverlappingCbtRoomProctor(ctx context.Context, arg db.HasOverlappingCbtRoomProctorParams) (bool, error)
	ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error)
	GetSessionProctoringStatus(ctx context.Context, arg db.GetSessionProctoringStatusParams) ([]db.GetSessionProctoringStatusRow, error)
	SetParticipantSuspiciousFlag(ctx context.Context, arg db.SetParticipantSuspiciousFlagParams) error
	GradeStudentEssay(ctx context.Context, arg db.GradeStudentEssayParams) error
	ListUngradedEssays(ctx context.Context, sessionID pgtype.UUID) ([]db.ListUngradedEssaysRow, error)
	QuestionBelongsToParticipantPackage(ctx context.Context, arg db.QuestionBelongsToParticipantPackageParams) (bool, error)
	UpsertStudentAnswer(ctx context.Context, arg db.UpsertStudentAnswerParams) (int64, error)
	UpdateAnswerCorrectness(ctx context.Context, sessionID pgtype.UUID) error
	UpdateParticipantScores(ctx context.Context, sessionID pgtype.UUID) error
	ListCbtExamSessionsByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamSessionsByTeacherRow, error)
	GetSessionResultsByTeacher(ctx context.Context, arg db.GetSessionResultsByTeacherParams) ([]db.GetSessionResultsByTeacherRow, error)
	GetSessionTeacherAccess(ctx context.Context, arg db.GetSessionTeacherAccessParams) (bool, error)
	HasSessionParticipant(ctx context.Context, arg db.HasSessionParticipantParams) (bool, error)
	HasSessionParticipantByTeacher(ctx context.Context, arg db.HasSessionParticipantByTeacherParams) (bool, error)
	HasSessionRoom(ctx context.Context, arg db.HasSessionRoomParams) (bool, error)
	HasSessionAnswer(ctx context.Context, arg db.HasSessionAnswerParams) (bool, error)
	HasSessionAnswerByTeacher(ctx context.Context, arg db.HasSessionAnswerByTeacherParams) (bool, error)
	GetSessionResults(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionResultsRow, error)
	GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error)
	ListUngradedEssaysByTeacher(ctx context.Context, arg db.ListUngradedEssaysByTeacherParams) ([]db.ListUngradedEssaysByTeacherRow, error)
	WithTx(tx pgx.Tx) *db.Queries
}

type cbtRoomShuffleStore interface {
	GetCbtExamSession(ctx context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error)
	ClearParticipantRooms(ctx context.Context, sessionID pgtype.UUID) error
	ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error)
	ListCbtExamRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error)
	AssignParticipantRoom(ctx context.Context, arg db.AssignParticipantRoomParams) error
}

type cbtRoomAssignmentStore interface {
	GetCbtExamSession(ctx context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error)
	ClearParticipantRooms(ctx context.Context, sessionID pgtype.UUID) error
	ClearParticipantSeatsForSession(ctx context.Context, sessionID pgtype.UUID) error
	ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error)
	ListCbtExamRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error)
	AssignParticipantSeat(ctx context.Context, arg db.AssignParticipantSeatParams) error
}

type cbtSeatAssignmentStore interface {
	ClearParticipantSeatsForSession(ctx context.Context, sessionID pgtype.UUID) error
	ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error)
	AssignParticipantSeat(ctx context.Context, arg db.AssignParticipantSeatParams) error
}

type cbtScoreStore interface {
	UpdateAnswerCorrectness(ctx context.Context, sessionID pgtype.UUID) error
	UpdateParticipantScores(ctx context.Context, sessionID pgtype.UUID) error
}

type cbtFinalizeOverdueStore interface {
	UpdateAnswerCorrectness(ctx context.Context, sessionID pgtype.UUID) error
	FinalizeOverdueParticipants(ctx context.Context, sessionID pgtype.UUID) (int32, error)
}

type cbtResultFollowUpStore interface {
	GetCbtSessionGradeSyncPreflight(ctx context.Context, id pgtype.UUID) (db.GetCbtSessionGradeSyncPreflightRow, error)
	ListCbtSessionRemedialCandidates(ctx context.Context, arg db.ListCbtSessionRemedialCandidatesParams) ([]db.ListCbtSessionRemedialCandidatesRow, error)
}

type cbtItemAnalysisStore interface {
	GetSessionItemAnalysis(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionItemAnalysisRow, error)
}

type cbtParticipantForceSubmitStore interface {
	UpdateParticipantAnswerCorrectness(ctx context.Context, participantID pgtype.UUID) error
	ForceSubmitParticipant(ctx context.Context, arg db.ForceSubmitParticipantParams) (db.ForceSubmitParticipantRow, error)
	InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error
}

type cbtParticipantEventStore interface {
	ResetParticipantRuntimeAccess(ctx context.Context, id pgtype.UUID) error
	UnlockParticipantAntiCheat(ctx context.Context, id pgtype.UUID) (db.UnlockParticipantAntiCheatRow, error)
	ListSessionParticipantEvents(ctx context.Context, arg db.ListSessionParticipantEventsParams) ([]db.ListSessionParticipantEventsRow, error)
	InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error
}

type cbtRoomWebFallbackPolicyStore interface {
	UpdateCbtRoomWebFallbackPolicy(ctx context.Context, arg db.UpdateCbtRoomWebFallbackPolicyParams) (db.CbtExamRoom, error)
}

const (
	ParticipantCommandWarningMessage = "warning_message"
	ParticipantCommandReconnect      = "reconnect"
	ParticipantCommandUnlockNotice   = "unlock_notice"
)

func NewCbtSession(pool *pgxpool.Pool) *CbtSession {
	return &CbtSession{q: db.New(pool), pool: pool}
}

func (s *CbtSession) List(ctx context.Context) ([]db.ListCbtExamSessionsRow, error) {
	rows, err := s.q.ListCbtExamSessions(ctx)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamSessionsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) Get(ctx context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error) {
	return s.q.GetCbtExamSession(ctx, id)
}

type CreateCbtSessionInput struct {
	PackageID       pgtype.UUID
	ClassID         pgtype.UUID
	EventID         pgtype.UUID
	ScopeType       string
	ScopeRef        string
	MixPolicy       string
	AssignmentMode  string
	AllowCrossGrade bool
	IsSpecialEvent  bool
	Title           string
	ScheduledStart  pgtype.Timestamptz
	ScheduledEnd    pgtype.Timestamptz
	Status          db.CbtSessionStatusEnum
}

func (s *CbtSession) Create(ctx context.Context, in CreateCbtSessionInput) (db.CbtExamSession, error) {
	if in.Status == "" {
		in.Status = db.CbtSessionStatusEnumDraft
	}
	if !validCbtSessionStatus(in.Status) {
		return db.CbtExamSession{}, fmt.Errorf("%w: status sesi CBT tidak valid", domain.ErrBadRequest)
	}
	scopeType := normalizeScopeType(in.ScopeType)
	mixPolicy := normalizeMixPolicy(in.MixPolicy, scopeType)
	assignmentMode := normalizeAssignmentMode(in.AssignmentMode)
	if err := s.ensureCbtPackageReadyForSession(ctx, in.PackageID, in.EventID); err != nil {
		return db.CbtExamSession{}, err
	}
	return s.q.CreateCbtExamSession(ctx, db.CreateCbtExamSessionParams{
		PackageID:       in.PackageID,
		ClassID:         in.ClassID,
		EventID:         in.EventID,
		ScopeType:       scopeType,
		ScopeRef:        in.ScopeRef,
		MixPolicy:       mixPolicy,
		AssignmentMode:  assignmentMode,
		AllowCrossGrade: in.AllowCrossGrade,
		IsSpecialEvent:  in.IsSpecialEvent,
		Title:           in.Title,
		ScheduledStart:  in.ScheduledStart,
		ScheduledEnd:    in.ScheduledEnd,
		Status:          in.Status,
	})
}

func (s *CbtSession) ensureCbtPackageReadyForSession(ctx context.Context, packageID, eventID pgtype.UUID) error {
	if err := s.ensureCbtPackageEventForSession(ctx, packageID, eventID); err != nil {
		return err
	}
	quality, err := s.q.GetCbtPackageQuestionQuality(ctx, packageID)
	if err != nil {
		return err
	}
	return validateCbtPackageQualityForSession(quality)
}

func (s *CbtSession) ensureCbtPackageEventForSession(ctx context.Context, packageID, eventID pgtype.UUID) error {
	pkg, err := s.findCbtPackage(ctx, packageID)
	if err != nil {
		return err
	}
	if !eventID.Valid && pkg.EventID.Valid {
		return fmt.Errorf("%w: paket khusus event hanya boleh dipakai pada sesi event yang sama", domain.ErrBadRequest)
	}
	return nil
}

func (s *CbtSession) findCbtPackage(ctx context.Context, packageID pgtype.UUID) (db.ListCbtPackagesRow, error) {
	rows, err := s.q.ListCbtPackages(ctx, pgtype.UUID{})
	if err != nil {
		return db.ListCbtPackagesRow{}, err
	}
	for _, row := range rows {
		if sameUUID(row.ID, packageID) {
			return row, nil
		}
	}
	return db.ListCbtPackagesRow{}, domain.ErrNotFound
}

func validateCbtPackageQualityForSession(quality db.GetCbtPackageQuestionQualityRow) error {
	switch {
	case !quality.IsActive:
		return fmt.Errorf("%w: paket soal tidak aktif", domain.ErrConflict)
	case quality.TotalQuestions == 0:
		return fmt.Errorf("%w: paket soal belum memiliki soal", domain.ErrConflict)
	case quality.PublishedQuestions == 0:
		return fmt.Errorf("%w: paket soal belum memiliki soal terbit", domain.ErrConflict)
	case quality.UnpublishedQuestions > 0:
		return fmt.Errorf("%w: paket soal masih memiliki %d soal belum terbit", domain.ErrConflict, quality.UnpublishedQuestions)
	default:
		return nil
	}
}

func (s *CbtSession) UpdateStatus(ctx context.Context, id pgtype.UUID, status db.CbtSessionStatusEnum) (db.CbtExamSession, error) {
	if !validCbtSessionStatus(status) {
		return db.CbtExamSession{}, fmt.Errorf("%w: status sesi CBT tidak valid", domain.ErrBadRequest)
	}
	var session db.GetCbtExamSessionRow
	if status == db.CbtSessionStatusEnumScheduled || status == db.CbtSessionStatusEnumActive {
		var err error
		session, err = s.q.GetCbtExamSession(ctx, id)
		if err != nil {
			return db.CbtExamSession{}, err
		}
		if err := s.ensureCbtPackageReadyForSession(ctx, session.PackageID, session.EventID); err != nil {
			return db.CbtExamSession{}, err
		}
	}
	if status == db.CbtSessionStatusEnumActive {
		if session.ScheduledEnd.Valid && time.Now().After(session.ScheduledEnd.Time) {
			return db.CbtExamSession{}, fmt.Errorf("%w: jadwal sesi sudah berakhir", domain.ErrConflict)
		}
		readiness, err := s.q.GetCbtSessionRoomReadiness(ctx, id)
		if err != nil {
			return db.CbtExamSession{}, err
		}
		if err := validateCbtSessionActivationReadiness(readiness); err != nil {
			return db.CbtExamSession{}, err
		}
	}
	if status == db.CbtSessionStatusEnumScheduled || status == db.CbtSessionStatusEnumActive {
		if s.pool != nil {
			conn, err := s.pool.Acquire(ctx)
			if err != nil {
				return db.CbtExamSession{}, err
			}
			defer conn.Release()

			tx, err := conn.Begin(ctx)
			if err != nil {
				return db.CbtExamSession{}, err
			}
			defer tx.Rollback(ctx) //nolint:errcheck

			qtx := s.q.WithTx(tx)
			if _, err := lockCbtPackageSnapshot(ctx, qtx, session.PackageID, pgtype.UUID{}, "session_"+string(status)); err != nil {
				return db.CbtExamSession{}, err
			}
			updated, err := qtx.UpdateCbtExamSessionStatus(ctx, db.UpdateCbtExamSessionStatusParams{
				ID:     id,
				Status: status,
			})
			if err != nil {
				return db.CbtExamSession{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return db.CbtExamSession{}, err
			}
			return updated, nil
		}
		snapshotStore, ok := s.q.(cbtPackageSnapshotStore)
		if ok {
			if _, err := lockCbtPackageSnapshot(ctx, snapshotStore, session.PackageID, pgtype.UUID{}, "session_"+string(status)); err != nil {
				return db.CbtExamSession{}, err
			}
		}
	}
	return s.q.UpdateCbtExamSessionStatus(ctx, db.UpdateCbtExamSessionStatusParams{
		ID:     id,
		Status: status,
	})
}

func (s *CbtSession) UpdateSchedule(ctx context.Context, id pgtype.UUID, start, end pgtype.Timestamptz) (UpdateCbtSessionScheduleResult, error) {
	if !start.Valid || !end.Valid || !end.Time.After(start.Time) {
		return UpdateCbtSessionScheduleResult{}, fmt.Errorf("%w: jadwal selesai harus setelah jadwal mulai", domain.ErrBadRequest)
	}
	if time.Now().After(end.Time) {
		return UpdateCbtSessionScheduleResult{}, fmt.Errorf("%w: jadwal baru sudah berakhir", domain.ErrConflict)
	}
	session, err := s.q.GetCbtExamSession(ctx, id)
	if err != nil {
		return UpdateCbtSessionScheduleResult{}, err
	}
	if session.Status != db.CbtSessionStatusEnumDraft && session.Status != db.CbtSessionStatusEnumScheduled {
		return UpdateCbtSessionScheduleResult{}, fmt.Errorf("%w: hanya sesi draft atau terjadwal yang boleh diubah jadwalnya", domain.ErrConflict)
	}
	updated, err := s.q.UpdateCbtExamSessionSchedule(ctx, db.UpdateCbtExamSessionScheduleParams{
		ID:             id,
		ScheduledStart: start,
		ScheduledEnd:   end,
	})
	if err != nil {
		return UpdateCbtSessionScheduleResult{}, err
	}
	return UpdateCbtSessionScheduleResult{Before: session, Session: updated}, nil
}

func (s *CbtSession) ListAuditLogs(ctx context.Context, id pgtype.UUID, limit, offset int32) ([]db.ListEntityAuditLogsRow, error) {
	return s.q.ListEntityAuditLogs(ctx, db.ListEntityAuditLogsParams{
		EntityType: "cbt_session",
		EntityID:   pgUUIDString(id),
		Limit:      limit,
		Offset:     offset,
	})
}

func validateCbtSessionActivationReadiness(readiness db.GetCbtSessionRoomReadinessRow) error {
	switch {
	case readiness.ParticipantCount == 0:
		return fmt.Errorf("%w: sesi belum memiliki peserta", domain.ErrConflict)
	case readiness.RoomCount == 0:
		return fmt.Errorf("%w: sesi belum memiliki ruangan ujian", domain.ErrConflict)
	case readiness.TotalCapacity < readiness.ParticipantCount:
		return fmt.Errorf("%w: kapasitas ruangan belum cukup untuk seluruh peserta", domain.ErrConflict)
	case readiness.OverCapacityRoomCount > 0:
		return fmt.Errorf("%w: ada %d ruangan melebihi kapasitas efektif", domain.ErrConflict, readiness.OverCapacityRoomCount)
	case readiness.UnassignedParticipantCount > 0:
		return fmt.Errorf("%w: masih ada %d peserta belum mendapat ruangan", domain.ErrConflict, readiness.UnassignedParticipantCount)
	case readiness.MissingSeatCount > 0:
		return fmt.Errorf("%w: masih ada %d peserta belum mendapat nomor meja", domain.ErrConflict, readiness.MissingSeatCount)
	case readiness.RoomsWithoutProctor > 0:
		return fmt.Errorf("%w: masih ada %d ruangan belum punya pengawas", domain.ErrConflict, readiness.RoomsWithoutProctor)
	case readiness.NetworkNotReadyRoomCount > 0:
		return fmt.Errorf("%w: ada %d ruangan CBT dengan jaringan belum siap", domain.ErrConflict, readiness.NetworkNotReadyRoomCount)
	case readiness.PowerNotReadyRoomCount > 0:
		return fmt.Errorf("%w: ada %d ruangan CBT dengan listrik belum siap", domain.ErrConflict, readiness.PowerNotReadyRoomCount)
	default:
		return nil
	}
}

func (s *CbtSession) Delete(ctx context.Context, id pgtype.UUID) error {
	rows, err := s.q.DeleteCbtExamSession(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: sesi tidak ditemukan atau bukan draft", domain.ErrConflict)
	}
	return nil
}

// --- Participants ---

func (s *CbtSession) ListParticipants(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error) {
	rows, err := s.q.ListCbtExamParticipants(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamParticipantsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) ListParticipantsByTeacher(ctx context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error) {
	rows, err := s.q.ListCbtExamParticipantsByTeacher(ctx, db.ListCbtExamParticipantsByTeacherParams{
		TeacherEmployeeID: teacherEmployeeID,
		SessionID:         sessionID,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamParticipantsRow{}, nil
	}
	out := make([]db.ListCbtExamParticipantsRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, db.ListCbtExamParticipantsRow(row))
	}
	return out, nil
}

func (s *CbtSession) EnrollClass(ctx context.Context, sessionID, classID pgtype.UUID) error {
	if err := s.ensureSessionSetupMutable(ctx, sessionID); err != nil {
		return err
	}
	return s.q.EnrollClassToSession(ctx, db.EnrollClassToSessionParams{
		SessionID: sessionID,
		ClassID:   classID,
	})
}

func (s *CbtSession) EnrollGrade(ctx context.Context, sessionID pgtype.UUID, level string) error {
	if err := s.ensureSessionSetupMutable(ctx, sessionID); err != nil {
		return err
	}
	return s.q.EnrollGradeToSession(ctx, db.EnrollGradeToSessionParams{
		SessionID: sessionID,
		Level:     level,
	})
}

func (s *CbtSession) EnrollSchool(ctx context.Context, sessionID pgtype.UUID) error {
	if err := s.ensureSessionSetupMutable(ctx, sessionID); err != nil {
		return err
	}
	return s.q.EnrollSchoolToSession(ctx, sessionID)
}

func normalizeScopeType(value string) string {
	switch value {
	case "class", "grade", "school", "custom":
		return value
	default:
		return "class"
	}
}

func normalizeMixPolicy(value string, scopeType string) string {
	switch scopeType {
	case "class":
		return "same_class"
	case "grade":
		return "same_grade"
	case "school", "custom":
		return "mixed_scope"
	default:
		switch value {
		case "same_class", "same_grade", "mixed_scope":
			return value
		default:
			return "same_grade"
		}
	}
}

func normalizeAssignmentMode(value string) string {
	switch value {
	case "manual", "random_balanced", "random_by_gender", "random_by_accommodation":
		return value
	default:
		return "random_balanced"
	}
}

func validCbtSessionStatus(status db.CbtSessionStatusEnum) bool {
	switch status {
	case db.CbtSessionStatusEnumDraft,
		db.CbtSessionStatusEnumScheduled,
		db.CbtSessionStatusEnumActive,
		db.CbtSessionStatusEnumFinished,
		db.CbtSessionStatusEnumCancelled:
		return true
	default:
		return false
	}
}

func (s *CbtSession) GenerateTokens(ctx context.Context, sessionID pgtype.UUID) error {
	if err := s.ensureSessionTokenMutable(ctx, sessionID); err != nil {
		return err
	}
	return s.q.GenerateTokensForSession(ctx, sessionID)
}

type RegenerateTokenResult struct {
	ID    pgtype.UUID `json:"id"`
	Token string      `json:"token"`
}

func (s *CbtSession) RegenerateToken(ctx context.Context, participantID pgtype.UUID) (db.RegenerateParticipantTokenRow, error) {
	row, err := s.q.RegenerateParticipantToken(ctx, participantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.RegenerateParticipantTokenRow{}, fmt.Errorf("%w: token peserta hanya boleh diubah sebelum sesi aktif", domain.ErrConflict)
	}
	return row, err
}

func (s *CbtSession) ensureSessionTokenMutable(ctx context.Context, sessionID pgtype.UUID) error {
	session, err := s.q.GetCbtExamSession(ctx, sessionID)
	if err != nil {
		return err
	}
	switch session.Status {
	case db.CbtSessionStatusEnumDraft, db.CbtSessionStatusEnumScheduled:
		return nil
	default:
		return fmt.Errorf("%w: token sesi hanya boleh dibuat sebelum sesi aktif", domain.ErrConflict)
	}
}

func (s *CbtSession) ResetParticipantRuntimeAccess(ctx context.Context, participantID pgtype.UUID, actor string) error {
	q, ok := s.q.(cbtParticipantEventStore)
	if !ok {
		return fmt.Errorf("cbt participant event store unavailable")
	}
	if err := q.ResetParticipantRuntimeAccess(ctx, participantID); err != nil {
		return err
	}
	return q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "proctor_reset_access",
		EventData:     marshalJSON(map[string]string{"actor": actor}),
	})
}

func (s *CbtSession) UnlockParticipantAntiCheat(ctx context.Context, participantID pgtype.UUID, actor, notes string) (db.UnlockParticipantAntiCheatRow, error) {
	q, ok := s.q.(cbtParticipantEventStore)
	if !ok {
		return db.UnlockParticipantAntiCheatRow{}, fmt.Errorf("cbt participant event store unavailable")
	}
	row, err := q.UnlockParticipantAntiCheat(ctx, participantID)
	if err != nil {
		return db.UnlockParticipantAntiCheatRow{}, err
	}
	if err := q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "proctor_unlock",
		EventData:     marshalJSON(map[string]string{"actor": actor, "notes": strings.TrimSpace(notes)}),
	}); err != nil {
		return db.UnlockParticipantAntiCheatRow{}, err
	}
	return row, nil
}

func (s *CbtSession) AcknowledgeProctorEvent(ctx context.Context, participantID pgtype.UUID, eventID, actor, notes string) error {
	q, ok := s.q.(cbtParticipantEventStore)
	if !ok {
		return fmt.Errorf("cbt participant event store unavailable")
	}
	return q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "proctor_acknowledge",
		EventData: marshalJSON(map[string]string{
			"actor":    actor,
			"event_id": strings.TrimSpace(eventID),
			"notes":    strings.TrimSpace(notes),
		}),
	})
}

func (s *CbtSession) RecordIncidentAction(ctx context.Context, participantID pgtype.UUID, eventID, action, actor, notes string) error {
	q, ok := s.q.(cbtParticipantEventStore)
	if !ok {
		return fmt.Errorf("cbt participant event store unavailable")
	}
	action = normalizeIncidentAction(action)
	if action == "" {
		return fmt.Errorf("%w: tindakan insiden tidak valid", domain.ErrBadRequest)
	}
	return q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "proctor_incident_action",
		EventData: marshalJSON(map[string]string{
			"actor":    strings.TrimSpace(actor),
			"event_id": strings.TrimSpace(eventID),
			"action":   action,
			"notes":    strings.TrimSpace(notes),
		}),
	})
}

func (s *CbtSession) SendParticipantCommand(ctx context.Context, participantID pgtype.UUID, commandType, message, actor string) error {
	q, ok := s.q.(cbtParticipantEventStore)
	if !ok {
		return fmt.Errorf("cbt participant event store unavailable")
	}
	commandType = normalizeParticipantCommand(commandType)
	if commandType == "" {
		return fmt.Errorf("%w: perintah peserta tidak valid", domain.ErrBadRequest)
	}
	message = strings.TrimSpace(message)
	if message == "" {
		message = defaultParticipantCommandMessage(commandType)
	}
	return q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "participant_command",
		EventData: marshalJSON(map[string]string{
			"actor":        strings.TrimSpace(actor),
			"command_type": commandType,
			"message":      message,
			"severity":     participantCommandSeverity(commandType),
			"issued_at":    time.Now().Format(time.RFC3339),
		}),
	})
}

func (s *CbtSession) HasParticipant(ctx context.Context, sessionID, participantID pgtype.UUID) (bool, error) {
	return s.q.HasSessionParticipant(ctx, db.HasSessionParticipantParams{
		SessionID: sessionID,
		ID:        participantID,
	})
}

func (s *CbtSession) HasParticipantByTeacher(ctx context.Context, sessionID, participantID, teacherEmployeeID pgtype.UUID) (bool, error) {
	return s.q.HasSessionParticipantByTeacher(ctx, db.HasSessionParticipantByTeacherParams{
		TeacherEmployeeID: teacherEmployeeID,
		SessionID:         sessionID,
		ParticipantID:     participantID,
	})
}

func (s *CbtSession) HasRoom(ctx context.Context, sessionID, roomID pgtype.UUID) (bool, error) {
	return s.q.HasSessionRoom(ctx, db.HasSessionRoomParams{
		SessionID: sessionID,
		ID:        roomID,
	})
}

func (s *CbtSession) HasAnswer(ctx context.Context, sessionID, answerID pgtype.UUID) (bool, error) {
	return s.q.HasSessionAnswer(ctx, db.HasSessionAnswerParams{
		SessionID: sessionID,
		ID:        answerID,
	})
}

func (s *CbtSession) HasAnswerByTeacher(ctx context.Context, sessionID, answerID, teacherEmployeeID pgtype.UUID) (bool, error) {
	return s.q.HasSessionAnswerByTeacher(ctx, db.HasSessionAnswerByTeacherParams{
		TeacherEmployeeID: teacherEmployeeID,
		SessionID:         sessionID,
		AnswerID:          answerID,
	})
}

// --- Rooms ---

func (s *CbtSession) ListRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error) {
	rows, err := s.q.ListCbtExamRooms(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamRoomsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) CreateRoom(ctx context.Context, sessionID pgtype.UUID, roomName string, capacity int32) (db.CbtExamRoom, error) {
	if err := s.ensureSessionSetupMutable(ctx, sessionID); err != nil {
		return db.CbtExamRoom{}, err
	}
	roomName = strings.TrimSpace(roomName)
	if capacity <= 0 {
		capacity = 30
	}
	room, err := s.q.CreateCbtExamRoom(ctx, db.CreateCbtExamRoomParams{
		SessionID:        sessionID,
		SchoolRoomID:     pgtype.UUID{},
		RoomName:         roomName,
		RoomNameSnapshot: roomName,
		Capacity:         capacity,
	})
	if err != nil {
		return db.CbtExamRoom{}, mapCbtRoomSetupError(err)
	}
	return room, nil
}

func (s *CbtSession) CreateRoomFromSchoolRoom(ctx context.Context, sessionID, schoolRoomID pgtype.UUID, roomName string, capacity int32) (db.CbtExamRoom, error) {
	if err := s.ensureSessionSetupMutable(ctx, sessionID); err != nil {
		return db.CbtExamRoom{}, err
	}
	schoolRoom, err := s.q.GetSchoolRoom(ctx, schoolRoomID)
	if err != nil {
		return db.CbtExamRoom{}, err
	}
	if !schoolRoom.IsExamEligible || schoolRoom.Condition == "rusak" {
		return db.CbtExamRoom{}, fmt.Errorf("%w: ruangan fisik belum layak dipakai untuk ujian", domain.ErrConflict)
	}
	if !schoolRoom.NetworkReady {
		return db.CbtExamRoom{}, fmt.Errorf("%w: jaringan ruangan fisik belum siap untuk CBT", domain.ErrConflict)
	}
	if !schoolRoom.PowerReady {
		return db.CbtExamRoom{}, fmt.Errorf("%w: listrik ruangan fisik belum siap untuk CBT", domain.ErrConflict)
	}
	roomName = strings.TrimSpace(roomName)
	if roomName == "" {
		roomName = schoolRoom.Name
	}
	if capacity <= 0 {
		capacity = schoolRoom.ExamCapacity
	}
	if capacity <= 0 {
		capacity = schoolRoom.DefaultCapacity
	}
	if capacity <= 0 {
		capacity = 30
	}
	room, err := s.q.CreateCbtExamRoom(ctx, db.CreateCbtExamRoomParams{
		SessionID:        sessionID,
		SchoolRoomID:     schoolRoomID,
		RoomName:         roomName,
		RoomNameSnapshot: roomName,
		Capacity:         capacity,
	})
	if err != nil {
		return db.CbtExamRoom{}, mapCbtRoomSetupError(err)
	}
	return room, nil
}

func (s *CbtSession) DeleteRoom(ctx context.Context, roomID pgtype.UUID) error {
	ctxRow, err := s.q.GetCbtExamRoomSetupContext(ctx, roomID)
	if err != nil {
		return err
	}
	if err := ensureCbtSessionSetupStatusMutable(ctxRow.SessionStatus); err != nil {
		return err
	}
	usage, err := s.q.GetCbtExamRoomSetupUsage(ctx, roomID)
	if err != nil {
		return err
	}
	switch {
	case usage.ParticipantCount > 0:
		return fmt.Errorf("%w: ruangan masih memiliki %d peserta", domain.ErrConflict, usage.ParticipantCount)
	case usage.ProctorCount > 0:
		return fmt.Errorf("%w: ruangan masih memiliki %d pengawas", domain.ErrConflict, usage.ProctorCount)
	case usage.HandoverCount > 0:
		return fmt.Errorf("%w: ruangan sudah memiliki serah terima", domain.ErrConflict)
	}
	if err := s.q.DeleteCbtExamRoom(ctx, roomID); err != nil {
		return mapCbtRoomSetupError(err)
	}
	return nil
}

func (s *CbtSession) ListRoomProctors(ctx context.Context, roomID pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error) {
	rows, err := s.q.ListCbtRoomProctors(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtRoomProctorsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) ReplaceRoomProctors(ctx context.Context, roomID, assignedBy, primaryEmployeeID pgtype.UUID, employeeIDs []pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error) {
	room, err := s.q.GetCbtExamRoomSetupContext(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if err := ensureCbtSessionSetupStatusMutable(room.SessionStatus); err != nil {
		return nil, err
	}
	ordered := normalizeRoomProctorIDs(primaryEmployeeID, employeeIDs)
	for _, employeeID := range ordered {
		overlaps, err := s.q.HasOverlappingCbtRoomProctor(ctx, db.HasOverlappingCbtRoomProctorParams{
			SessionID:  room.SessionID,
			EmployeeID: employeeID,
			ExamRoomID: roomID,
		})
		if err != nil {
			return nil, err
		}
		if overlaps {
			return nil, fmt.Errorf("%w: pengawas sudah bertugas pada sesi CBT lain yang waktunya bertabrakan", domain.ErrConflict)
		}
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.q.WithTx(tx)
	if err := qtx.DeleteCbtRoomProctorsByRoom(ctx, roomID); err != nil {
		return nil, mapCbtRoomSetupError(err)
	}

	for index, employeeID := range ordered {
		role := "pendamping"
		if index == 0 && primaryEmployeeID.Valid && employeeID == primaryEmployeeID {
			role = "utama"
		}
		if _, err := qtx.CreateCbtRoomProctor(ctx, db.CreateCbtRoomProctorParams{
			ExamRoomID: roomID,
			EmployeeID: employeeID,
			Role:       role,
			AssignedBy: assignedBy,
		}); err != nil {
			return nil, mapCbtRoomSetupError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.ListRoomProctors(ctx, roomID)
}

func (s *CbtSession) RoomReadiness(ctx context.Context, sessionID pgtype.UUID) (db.GetCbtSessionRoomReadinessRow, error) {
	return s.q.GetCbtSessionRoomReadiness(ctx, sessionID)
}

func (s *CbtSession) GetRoomProctoringDashboard(ctx context.Context, roomID pgtype.UUID) (db.GetCbtRoomProctorDashboardRow, error) {
	return s.q.GetCbtRoomProctorDashboard(ctx, roomID)
}

func (s *CbtSession) UpdateRoomWebFallbackPolicy(ctx context.Context, sessionID, roomID, actorUserID pgtype.UUID, allow bool, reason string) (db.CbtExamRoom, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return db.CbtExamRoom{}, domain.ErrBadRequest
	}
	q, ok := s.q.(cbtRoomWebFallbackPolicyStore)
	if !ok {
		return db.CbtExamRoom{}, fmt.Errorf("cbt web fallback policy store unavailable")
	}
	room, err := q.UpdateCbtRoomWebFallbackPolicy(ctx, db.UpdateCbtRoomWebFallbackPolicyParams{
		SessionID:        sessionID,
		RoomID:           roomID,
		AllowWebFallback: allow,
		ActorUserID:      actorUserID,
		Reason:           reason,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.CbtExamRoom{}, domain.ErrNotFound
		}
		return db.CbtExamRoom{}, err
	}
	return room, nil
}

type SaveCbtRoomHandoverInput struct {
	AttendanceChecked     bool
	AllSubmittedChecked   bool
	DeviceIssueChecked    bool
	RoomCleanChecked      bool
	TokenReturnedChecked  bool
	AssetsReturnedChecked bool
	IncidentNotes         string
	OperatorNotes         string
	HandoverNotes         string
}

func (s *CbtSession) GetRoomHandover(ctx context.Context, roomID pgtype.UUID) (db.GetCbtRoomHandoverRow, error) {
	return s.q.GetCbtRoomHandover(ctx, roomID)
}

func (s *CbtSession) GetSessionOperationalRecap(ctx context.Context, sessionID pgtype.UUID) (db.GetCbtSessionOperationalRecapRow, []db.ListCbtSessionRoomOperationalRecapRow, error) {
	recap, err := s.q.GetCbtSessionOperationalRecap(ctx, sessionID)
	if err != nil {
		return db.GetCbtSessionOperationalRecapRow{}, nil, err
	}
	rooms, err := s.q.ListCbtSessionRoomOperationalRecap(ctx, sessionID)
	if err != nil {
		return db.GetCbtSessionOperationalRecapRow{}, nil, err
	}
	if rooms == nil {
		rooms = []db.ListCbtSessionRoomOperationalRecapRow{}
	}
	return recap, rooms, nil
}

func (s *CbtSession) SaveRoomHandover(ctx context.Context, roomID, updatedBy pgtype.UUID, in SaveCbtRoomHandoverInput) (db.CbtRoomHandover, error) {
	row, err := s.q.UpsertCbtRoomHandover(ctx, db.UpsertCbtRoomHandoverParams{
		ExamRoomID:            roomID,
		AttendanceChecked:     in.AttendanceChecked,
		AllSubmittedChecked:   in.AllSubmittedChecked,
		DeviceIssueChecked:    in.DeviceIssueChecked,
		RoomCleanChecked:      in.RoomCleanChecked,
		TokenReturnedChecked:  in.TokenReturnedChecked,
		AssetsReturnedChecked: in.AssetsReturnedChecked,
		IncidentNotes:         strings.TrimSpace(in.IncidentNotes),
		OperatorNotes:         strings.TrimSpace(in.OperatorNotes),
		HandoverNotes:         strings.TrimSpace(in.HandoverNotes),
		UpdatedBy:             updatedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.CbtRoomHandover{}, fmt.Errorf("%w: serah terima ruang sudah dikunci", domain.ErrConflict)
		}
		return db.CbtRoomHandover{}, err
	}
	return row, nil
}

func (s *CbtSession) LockRoomHandover(ctx context.Context, roomID, lockedBy pgtype.UUID) (db.CbtRoomHandover, error) {
	row, err := s.q.LockCbtRoomHandover(ctx, db.LockCbtRoomHandoverParams{
		ExamRoomID: roomID,
		LockedBy:   lockedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.CbtRoomHandover{}, fmt.Errorf("%w: serah terima ruang sudah dikunci", domain.ErrConflict)
		}
		return db.CbtRoomHandover{}, err
	}
	return row, nil
}

func (s *CbtSession) ListProctorRooms(ctx context.Context, employeeID pgtype.UUID, includeAll bool) ([]db.ListCbtProctorRoomsRow, error) {
	rows, err := s.q.ListCbtProctorRooms(ctx, db.ListCbtProctorRoomsParams{
		EmployeeID: employeeID,
		IncludeAll: includeAll,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtProctorRoomsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) HasRoomProctor(ctx context.Context, sessionID, roomID, employeeID pgtype.UUID) (bool, error) {
	return s.q.HasSessionRoomProctor(ctx, db.HasSessionRoomProctorParams{
		SessionID:  sessionID,
		RoomID:     roomID,
		EmployeeID: employeeID,
	})
}

func (s *CbtSession) HasRoomParticipant(ctx context.Context, sessionID, roomID, participantID pgtype.UUID) (bool, error) {
	return s.q.HasSessionRoomParticipant(ctx, db.HasSessionRoomParticipantParams{
		SessionID:     sessionID,
		RoomID:        roomID,
		ParticipantID: participantID,
	})
}

func normalizeRoomProctorIDs(primary pgtype.UUID, ids []pgtype.UUID) []pgtype.UUID {
	seen := make(map[[16]byte]bool, len(ids)+1)
	out := make([]pgtype.UUID, 0, len(ids)+1)
	add := func(id pgtype.UUID) {
		if !id.Valid || seen[id.Bytes] {
			return
		}
		seen[id.Bytes] = true
		out = append(out, id)
	}
	add(primary)
	for _, id := range ids {
		add(id)
	}
	return out
}

func (s *CbtSession) AssignSeat(ctx context.Context, participantID, roomID pgtype.UUID, seatNo int32) error {
	room, err := s.q.GetCbtExamRoomSetupContext(ctx, roomID)
	if err != nil {
		return err
	}
	if err := ensureCbtSessionSetupStatusMutable(room.SessionStatus); err != nil {
		return err
	}
	if seatNo <= 0 {
		return fmt.Errorf("%w: nomor meja harus lebih dari 0", domain.ErrBadRequest)
	}
	belongs, err := s.q.HasSessionParticipant(ctx, db.HasSessionParticipantParams{
		SessionID: room.SessionID,
		ID:        participantID,
	})
	if err != nil {
		return err
	}
	if !belongs {
		return fmt.Errorf("%w: peserta dan ruang harus berada pada sesi CBT yang sama", domain.ErrBadRequest)
	}
	if err := s.q.AssignParticipantSeat(ctx, db.AssignParticipantSeatParams{
		ID:     participantID,
		RoomID: roomID,
		SeatNo: pgtype.Int4{Int32: seatNo, Valid: seatNo > 0},
	}); err != nil {
		return mapCbtRoomSetupError(err)
	}
	return nil
}

// ShuffleRooms randomly assigns participants to rooms respecting capacity.
// If a room is full, remaining participants are left unassigned.

func (s *CbtSession) PreviewRoomAssignment(ctx context.Context, sessionID pgtype.UUID, input CbtRoomAssignmentInput) (CbtRoomAssignmentPreview, error) {
	session, err := s.q.GetCbtExamSession(ctx, sessionID)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	participants, err := s.q.ListParticipantsByRoom(ctx, sessionID)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	rooms, err := s.q.ListCbtExamRooms(ctx, sessionID)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	return buildCbtRoomAssignmentPreview(session, participants, rooms, input)
}

func (s *CbtSession) ApplyRoomAssignment(ctx context.Context, sessionID pgtype.UUID, input CbtRoomAssignmentInput) (CbtRoomAssignmentPreview, error) {
	if err := s.ensureSessionSetupMutable(ctx, sessionID); err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	if s.pool == nil {
		return applyCbtRoomAssignment(ctx, s.q, sessionID, input)
	}
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	preview, err := applyCbtRoomAssignment(ctx, s.q.WithTx(tx), sessionID, input)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	return preview, nil
}

func applyCbtRoomAssignment(ctx context.Context, q cbtRoomAssignmentStore, sessionID pgtype.UUID, input CbtRoomAssignmentInput) (CbtRoomAssignmentPreview, error) {
	session, err := q.GetCbtExamSession(ctx, sessionID)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	if err := ensureCbtSessionSetupStatusMutable(session.Status); err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	participants, err := q.ListParticipantsByRoom(ctx, sessionID)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	rooms, err := q.ListCbtExamRooms(ctx, sessionID)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	preview, err := buildCbtRoomAssignmentPreview(session, participants, rooms, input)
	if err != nil {
		return CbtRoomAssignmentPreview{}, err
	}
	if err := q.ClearParticipantRooms(ctx, sessionID); err != nil {
		return CbtRoomAssignmentPreview{}, mapCbtRoomSetupError(err)
	}
	if err := q.ClearParticipantSeatsForSession(ctx, sessionID); err != nil {
		return CbtRoomAssignmentPreview{}, mapCbtRoomSetupError(err)
	}
	for _, assignment := range preview.Assignments {
		var participantID, roomID pgtype.UUID
		if err := participantID.Scan(assignment.ParticipantID); err != nil {
			return CbtRoomAssignmentPreview{}, err
		}
		if err := roomID.Scan(assignment.RoomID); err != nil {
			return CbtRoomAssignmentPreview{}, err
		}
		if err := q.AssignParticipantSeat(ctx, db.AssignParticipantSeatParams{ID: participantID, RoomID: roomID, SeatNo: pgtype.Int4{Int32: assignment.SeatNo, Valid: assignment.SeatNo > 0}}); err != nil {
			return CbtRoomAssignmentPreview{}, mapCbtRoomSetupError(err)
		}
	}
	return preview, nil
}

func (s *CbtSession) ShuffleRooms(ctx context.Context, sessionID pgtype.UUID) error {
	if err := s.ensureSessionSetupMutable(ctx, sessionID); err != nil {
		return err
	}
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

	qtx := s.q.WithTx(tx)

	if err := shuffleRooms(ctx, qtx, sessionID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func shuffleRooms(ctx context.Context, q cbtRoomShuffleStore, sessionID pgtype.UUID) error {
	session, err := q.GetCbtExamSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if err := q.ClearParticipantRooms(ctx, sessionID); err != nil {
		return err
	}

	participants, err := q.ListParticipantsByRoom(ctx, sessionID)
	if err != nil {
		return err
	}
	rooms, err := q.ListCbtExamRooms(ctx, sessionID)
	if err != nil {
		return err
	}
	if len(rooms) == 0 || len(participants) == 0 {
		return nil
	}

	assignments := planCbtRoomShuffle(session, participants, rooms)
	for _, assignment := range assignments {
		if err := q.AssignParticipantRoom(ctx, assignment); err != nil {
			return err
		}
	}
	return nil
}

func planCbtRoomShuffle(session db.GetCbtExamSessionRow, participants []db.ListParticipantsByRoomRow, rooms []db.ListCbtExamRoomsRow) []db.AssignParticipantRoomParams {
	placements := planCbtRoomPlacements(session, participants, rooms)
	assignments := make([]db.AssignParticipantRoomParams, 0, len(placements))
	for _, placement := range placements {
		assignments = append(assignments, db.AssignParticipantRoomParams{ID: placement.ParticipantID, RoomID: rooms[placement.RoomIndex].ID})
	}
	return assignments
}

func cbtShufflePolicyGroupKey(session db.GetCbtExamSessionRow, participant db.ListParticipantsByRoomRow) string {
	switch session.MixPolicy {
	case "same_class":
		if participant.ClassID.Valid {
			return "class:" + participant.ClassID.String()
		}
		if strings.TrimSpace(participant.ClassCode) != "" {
			return "class_code:" + strings.TrimSpace(participant.ClassCode)
		}
		return "class:unknown"
	case "mixed_scope":
		if session.IsSpecialEvent && session.AllowCrossGrade {
			return "mixed"
		}
	}
	level := strings.TrimSpace(participant.ClassLevel)
	if level == "" {
		level = "unknown"
	}
	return "grade:" + level
}

func buildCbtRoomAssignmentPreview(session db.GetCbtExamSessionRow, participants []db.ListParticipantsByRoomRow, rooms []db.ListCbtExamRoomsRow, input CbtRoomAssignmentInput) (CbtRoomAssignmentPreview, error) {
	policySession := session
	policySession.MixPolicy = normalizeRoomAssignmentMixPolicy(input.MixPolicy, session.MixPolicy)
	policySession.AssignmentMode = normalizeAssignmentMode(firstNonEmpty(input.AssignmentMode, session.AssignmentMode))
	policySession.AllowCrossGrade = input.AllowCrossGrade || session.AllowCrossGrade
	policySession.IsSpecialEvent = input.IsSpecialEvent || session.IsSpecialEvent
	if policySession.MixPolicy == "mixed_scope" && hasMultipleClassLevels(participants) && (!policySession.IsSpecialEvent || !policySession.AllowCrossGrade) {
		return CbtRoomAssignmentPreview{}, fmt.Errorf("%w: campur lintas tingkat hanya boleh untuk sesi khusus dengan izin campur tingkat", domain.ErrBadRequest)
	}

	assignments := planCbtBalancedRoomAssignment(policySession, participants, rooms)
	previewAssignments := assignmentSeatsFromPlan(assignments, participants, rooms)
	if len(input.Assignments) > 0 {
		manualAssignments, err := normalizeManualRoomAssignments(input.Assignments, participants, rooms)
		if err != nil {
			return CbtRoomAssignmentPreview{}, err
		}
		assignments = manualAssignments.db
		previewAssignments = manualAssignments.preview
	}

	preview := CbtRoomAssignmentPreview{
		Summary: CbtRoomAssignmentSummary{
			ParticipantCount: len(participants),
			RoomCount:        len(rooms),
			CapacityTotal:    int(cbtRoomsTotalEffectiveCapacity(rooms)),
			AssignedCount:    len(assignments),
			UnassignedCount:  max(0, len(participants)-len(assignments)),
			MixPolicy:        policySession.MixPolicy,
			AssignmentMode:   policySession.AssignmentMode,
			AllowCrossGrade:  policySession.AllowCrossGrade,
			IsSpecialEvent:   policySession.IsSpecialEvent,
		},
		Rooms:       summarizeCbtRoomAssignments(rooms, participants, assignments),
		Assignments: previewAssignments,
	}
	if preview.Summary.UnassignedCount > 0 {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("%d peserta belum mendapat ruang karena kapasitas kurang.", preview.Summary.UnassignedCount))
	}
	if policySession.MixPolicy == "mixed_scope" && hasMultipleClassLevels(participants) {
		preview.Warnings = append(preview.Warnings, "Ruang dapat berisi peserta lintas tingkat dan rombel. Gunakan hanya sesuai keputusan panitia.")
	}
	if countUnknownParticipantCohort(participants) > 0 {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("%d peserta belum memiliki data tingkat/rombel lengkap.", countUnknownParticipantCohort(participants)))
	}
	return preview, nil
}

func normalizeRoomAssignmentMixPolicy(value, fallback string) string {
	switch strings.TrimSpace(value) {
	case "same_class", "same_grade", "mixed_scope":
		return strings.TrimSpace(value)
	}
	switch strings.TrimSpace(fallback) {
	case "same_class", "same_grade", "mixed_scope":
		return strings.TrimSpace(fallback)
	default:
		return "same_grade"
	}
}

func planCbtBalancedRoomAssignment(session db.GetCbtExamSessionRow, participants []db.ListParticipantsByRoomRow, rooms []db.ListCbtExamRoomsRow) []db.AssignParticipantSeatParams {
	placements := planCbtRoomPlacements(session, participants, rooms)
	assignments := make([]db.AssignParticipantSeatParams, 0, len(placements))
	for _, placement := range placements {
		assignments = append(assignments, db.AssignParticipantSeatParams{
			ID:     placement.ParticipantID,
			RoomID: rooms[placement.RoomIndex].ID,
			SeatNo: pgtype.Int4{Int32: placement.SeatNo, Valid: true},
		})
	}
	return assignments
}

type cbtRoomPlacement struct {
	ParticipantID pgtype.UUID
	RoomIndex     int
	SeatNo        int32
}

func planCbtRoomPlacements(session db.GetCbtExamSessionRow, participants []db.ListParticipantsByRoomRow, rooms []db.ListCbtExamRoomsRow) []cbtRoomPlacement {
	if len(participants) == 0 || len(rooms) == 0 {
		return nil
	}
	placements := make([]cbtRoomPlacement, 0, min(len(participants), int(cbtRoomsTotalEffectiveCapacity(rooms))))
	roomUsed := make([]int32, len(rooms))
	seatNo := make([]int32, len(rooms))
	if session.MixPolicy == "mixed_scope" && session.IsSpecialEvent && session.AllowCrossGrade {
		appendBalancedCbtRoomPlacements(&placements, participants, rooms, allAvailableRoomIndexes(rooms), roomUsed, seatNo)
		return placements
	}

	groups := make(map[string][]db.ListParticipantsByRoomRow)
	keys := make([]string, 0)
	for _, participant := range participants {
		key := cbtShufflePolicyGroupKey(session, participant)
		if _, ok := groups[key]; !ok {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], participant)
	}
	sort.Strings(keys)
	roomCursor := 0
	for _, key := range keys {
		group := groups[key]
		roomIndexes := roomIndexesForParticipantGroup(rooms, roomCursor, len(group))
		if len(roomIndexes) == 0 {
			break
		}
		appendBalancedCbtRoomPlacements(&placements, group, rooms, roomIndexes, roomUsed, seatNo)
		roomCursor = roomIndexes[len(roomIndexes)-1] + 1
	}
	return placements
}

func nextAvailableRoomIndex(rooms []db.ListCbtExamRoomsRow, used []int32, start int) int {
	if len(rooms) == 0 {
		return -1
	}
	for offset := 0; offset < len(rooms); offset++ {
		idx := (start + offset) % len(rooms)
		if used[idx] < cbtRoomEffectiveCapacity(rooms[idx]) {
			return idx
		}
	}
	return -1
}

func nextAvailableSelectedRoomIndex(rooms []db.ListCbtExamRoomsRow, used []int32, selected []int, start int) int {
	if len(selected) == 0 {
		return -1
	}
	for offset := 0; offset < len(selected); offset++ {
		selectedPos := (start + offset) % len(selected)
		idx := selected[selectedPos]
		if used[idx] < cbtRoomEffectiveCapacity(rooms[idx]) {
			return idx
		}
	}
	return -1
}

func selectedRoomPosition(selected []int, roomIdx int) int {
	for pos, idx := range selected {
		if idx == roomIdx {
			return pos
		}
	}
	return 0
}

func allAvailableRoomIndexes(rooms []db.ListCbtExamRoomsRow) []int {
	indexes := make([]int, 0, len(rooms))
	for idx := range rooms {
		if cbtRoomEffectiveCapacity(rooms[idx]) > 0 {
			indexes = append(indexes, idx)
		}
	}
	return indexes
}

func roomIndexesForParticipantGroup(rooms []db.ListCbtExamRoomsRow, start int, participantCount int) []int {
	if participantCount <= 0 {
		return nil
	}
	indexes := make([]int, 0)
	capacity := int32(0)
	for idx := start; idx < len(rooms); idx++ {
		roomCapacity := cbtRoomEffectiveCapacity(rooms[idx])
		if roomCapacity <= 0 {
			continue
		}
		indexes = append(indexes, idx)
		capacity += roomCapacity
		if capacity >= int32(participantCount) {
			break
		}
	}
	return indexes
}

func appendBalancedCbtRoomPlacements(placements *[]cbtRoomPlacement, participants []db.ListParticipantsByRoomRow, rooms []db.ListCbtExamRoomsRow, selectedRooms []int, roomUsed []int32, seatNo []int32) {
	if len(participants) == 0 || len(selectedRooms) == 0 {
		return
	}
	cohorts := make(map[string][]db.ListParticipantsByRoomRow)
	keys := make([]string, 0)
	for _, participant := range participants {
		key := cbtRoomAssignmentCohortKey(participant)
		if _, ok := cohorts[key]; !ok {
			keys = append(keys, key)
		}
		cohorts[key] = append(cohorts[key], participant)
	}
	sort.Strings(keys)
	for cohortOffset, key := range keys {
		group := cohorts[key]
		order := cryptoPermInts(len(group))
		nextRoomPos := cohortOffset % len(selectedRooms)
		for _, participantIdx := range order {
			roomIdx := nextAvailableSelectedRoomIndex(rooms, roomUsed, selectedRooms, nextRoomPos)
			if roomIdx < 0 {
				return
			}
			seatNo[roomIdx]++
			*placements = append(*placements, cbtRoomPlacement{
				ParticipantID: group[participantIdx].ID,
				RoomIndex:     roomIdx,
				SeatNo:        seatNo[roomIdx],
			})
			roomUsed[roomIdx]++
			nextRoomPos = (selectedRoomPosition(selectedRooms, roomIdx) + 1) % len(selectedRooms)
		}
	}
}

func cbtRoomAssignmentCohortKey(participant db.ListParticipantsByRoomRow) string {
	if participant.ClassID.Valid {
		return "class:" + participant.ClassID.String()
	}
	classCode := strings.TrimSpace(participant.ClassCode)
	if classCode != "" {
		level := strings.TrimSpace(participant.ClassLevel)
		if level != "" {
			return "class_code:" + level + ":" + classCode
		}
		return "class_code:" + classCode
	}
	level := strings.TrimSpace(participant.ClassLevel)
	if level != "" {
		return "level:" + level
	}
	return "unknown"
}

func summarizeCbtRoomAssignments(rooms []db.ListCbtExamRoomsRow, participants []db.ListParticipantsByRoomRow, assignments []db.AssignParticipantSeatParams) []CbtRoomAssignmentRoom {
	participantByID := make(map[pgtype.UUID]db.ListParticipantsByRoomRow, len(participants))
	for _, participant := range participants {
		participantByID[participant.ID] = participant
	}
	out := make([]CbtRoomAssignmentRoom, len(rooms))
	roomIndex := make(map[pgtype.UUID]int, len(rooms))
	classSets := make([]map[string]bool, len(rooms))
	for i, room := range rooms {
		roomIndex[room.ID] = i
		classSets[i] = map[string]bool{}
		out[i] = CbtRoomAssignmentRoom{RoomID: pgUUIDString(room.ID), RoomName: room.RoomName, Capacity: int(cbtRoomEffectiveCapacity(room)), Levels: map[string]int{}, Classes: []string{}}
	}
	for _, assignment := range assignments {
		idx, ok := roomIndex[assignment.RoomID]
		if !ok {
			continue
		}
		participant, ok := participantByID[assignment.ID]
		if !ok {
			continue
		}
		out[idx].ParticipantCount++
		level := strings.TrimSpace(participant.ClassLevel)
		if level == "" {
			level = "Tanpa tingkat"
		}
		out[idx].Levels[level]++
		classCode := strings.TrimSpace(participant.ClassCode)
		if classCode == "" {
			classCode = "Tanpa rombel"
		}
		if !classSets[idx][classCode] {
			classSets[idx][classCode] = true
			out[idx].Classes = append(out[idx].Classes, classCode)
		}
	}
	for i := range out {
		sort.Strings(out[i].Classes)
	}
	return out
}

func assignmentSeatsFromPlan(assignments []db.AssignParticipantSeatParams, participants []db.ListParticipantsByRoomRow, rooms []db.ListCbtExamRoomsRow) []CbtRoomAssignmentSeat {
	participantByID := make(map[pgtype.UUID]db.ListParticipantsByRoomRow, len(participants))
	for _, participant := range participants {
		participantByID[participant.ID] = participant
	}
	roomByID := make(map[pgtype.UUID]db.ListCbtExamRoomsRow, len(rooms))
	for _, room := range rooms {
		roomByID[room.ID] = room
	}
	out := make([]CbtRoomAssignmentSeat, 0, len(assignments))
	for _, assignment := range assignments {
		participant, _ := participantByID[assignment.ID]
		room, _ := roomByID[assignment.RoomID]
		out = append(out, assignmentSeatFromRow(participant, room, assignment))
	}
	return out
}

func assignmentSeatFromRow(participant db.ListParticipantsByRoomRow, room db.ListCbtExamRoomsRow, assignment db.AssignParticipantSeatParams) CbtRoomAssignmentSeat {
	return CbtRoomAssignmentSeat{
		ParticipantID:    pgUUIDString(assignment.ID),
		ParticipantName:  strings.TrimSpace(participant.Nama),
		ParticipantNis:   strings.TrimSpace(participant.Nis),
		ParticipantClass: formatParticipantClass(participant),
		RoomID:           pgUUIDString(assignment.RoomID),
		RoomName:         room.RoomName,
		SeatNo:           assignment.SeatNo.Int32,
	}
}

type normalizedManualRoomAssignments struct {
	db      []db.AssignParticipantSeatParams
	preview []CbtRoomAssignmentSeat
}

func normalizeManualRoomAssignments(assignments []CbtRoomAssignmentSeat, participants []db.ListParticipantsByRoomRow, rooms []db.ListCbtExamRoomsRow) (normalizedManualRoomAssignments, error) {
	if len(assignments) == 0 {
		return normalizedManualRoomAssignments{}, nil
	}
	participantByID := make(map[string]db.ListParticipantsByRoomRow, len(participants))
	for _, participant := range participants {
		participantByID[pgUUIDString(participant.ID)] = participant
	}
	roomByID := make(map[string]db.ListCbtExamRoomsRow, len(rooms))
	for _, room := range rooms {
		roomByID[pgUUIDString(room.ID)] = room
	}
	seenParticipants := make(map[string]bool, len(assignments))
	seenSeats := make(map[string]map[int32]bool, len(rooms))
	outDB := make([]db.AssignParticipantSeatParams, 0, len(assignments))
	outPreview := make([]CbtRoomAssignmentSeat, 0, len(assignments))
	for _, assignment := range assignments {
		participantID := strings.TrimSpace(assignment.ParticipantID)
		roomID := strings.TrimSpace(assignment.RoomID)
		if participantID == "" || roomID == "" || assignment.SeatNo <= 0 {
			return normalizedManualRoomAssignments{}, fmt.Errorf("%w: data pembagian ruang manual tidak valid", domain.ErrBadRequest)
		}
		participant, ok := participantByID[participantID]
		if !ok {
			return normalizedManualRoomAssignments{}, fmt.Errorf("%w: peserta pembagian ruang manual tidak valid", domain.ErrBadRequest)
		}
		room, ok := roomByID[roomID]
		if !ok {
			return normalizedManualRoomAssignments{}, fmt.Errorf("%w: ruang pembagian manual tidak valid", domain.ErrBadRequest)
		}
		if assignment.SeatNo > cbtRoomEffectiveCapacity(room) {
			return normalizedManualRoomAssignments{}, fmt.Errorf("%w: nomor kursi melebihi kapasitas ruang", domain.ErrBadRequest)
		}
		if seenParticipants[participantID] {
			return normalizedManualRoomAssignments{}, fmt.Errorf("%w: peserta tidak boleh mendapat dua ruang atau dua kursi", domain.ErrConflict)
		}
		if seenSeats[roomID] == nil {
			seenSeats[roomID] = map[int32]bool{}
		}
		if seenSeats[roomID][assignment.SeatNo] {
			return normalizedManualRoomAssignments{}, fmt.Errorf("%w: nomor kursi pada ruang yang sama tidak boleh duplikat", domain.ErrConflict)
		}
		seenParticipants[participantID] = true
		seenSeats[roomID][assignment.SeatNo] = true
		seatNo := pgtype.Int4{Int32: assignment.SeatNo, Valid: true}
		outDB = append(outDB, db.AssignParticipantSeatParams{ID: participant.ID, RoomID: room.ID, SeatNo: seatNo})
		outPreview = append(outPreview, assignmentSeatFromRow(participant, room, db.AssignParticipantSeatParams{ID: participant.ID, RoomID: room.ID, SeatNo: seatNo}))
	}
	if len(seenParticipants) != len(participants) {
		return normalizedManualRoomAssignments{}, fmt.Errorf("%w: masih ada peserta yang belum diberi ruang atau kursi", domain.ErrBadRequest)
	}
	return normalizedManualRoomAssignments{db: outDB, preview: outPreview}, nil
}

func formatParticipantClass(participant db.ListParticipantsByRoomRow) string {
	parts := []string{}
	if level := strings.TrimSpace(participant.ClassLevel); level != "" {
		parts = append(parts, level)
	}
	if code := strings.TrimSpace(participant.ClassCode); code != "" {
		parts = append(parts, code)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

func hasMultipleClassLevels(participants []db.ListParticipantsByRoomRow) bool {
	levels := map[string]bool{}
	for _, participant := range participants {
		level := strings.TrimSpace(participant.ClassLevel)
		if level == "" {
			continue
		}
		levels[level] = true
		if len(levels) > 1 {
			return true
		}
	}
	return false
}

func countUnknownParticipantCohort(participants []db.ListParticipantsByRoomRow) int {
	count := 0
	for _, participant := range participants {
		if strings.TrimSpace(participant.ClassLevel) == "" || strings.TrimSpace(participant.ClassCode) == "" {
			count++
		}
	}
	return count
}

func cbtRoomsTotalEffectiveCapacity(rooms []db.ListCbtExamRoomsRow) int32 {
	var total int32
	for _, room := range rooms {
		if capacity := cbtRoomEffectiveCapacity(room); capacity > 0 {
			total += capacity
		}
	}
	return total
}

func (s *CbtSession) AutoAssignSeats(ctx context.Context, sessionID pgtype.UUID) error {
	if err := s.ensureSessionSetupMutable(ctx, sessionID); err != nil {
		return err
	}
	if s.pool == nil {
		return autoAssignSeats(ctx, s.q, sessionID)
	}
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

	if err := autoAssignSeats(ctx, s.q.WithTx(tx), sessionID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func autoAssignSeats(ctx context.Context, q cbtSeatAssignmentStore, sessionID pgtype.UUID) error {
	if err := q.ClearParticipantSeatsForSession(ctx, sessionID); err != nil {
		return mapCbtRoomSetupError(err)
	}
	participants, err := q.ListParticipantsByRoom(ctx, sessionID)
	if err != nil {
		return err
	}
	grouped := make(map[string][]db.ListParticipantsByRoomRow)
	for _, participant := range participants {
		if !participant.RoomID.Valid {
			continue
		}
		key := participant.RoomID.String()
		grouped[key] = append(grouped[key], participant)
	}
	for _, rows := range grouped {
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Nama == rows[j].Nama {
				return rows[i].Nis < rows[j].Nis
			}
			return rows[i].Nama < rows[j].Nama
		})
		for idx, participant := range rows {
			if err := q.AssignParticipantSeat(ctx, db.AssignParticipantSeatParams{
				ID:     participant.ID,
				RoomID: participant.RoomID,
				SeatNo: pgtype.Int4{Int32: int32(idx + 1), Valid: true},
			}); err != nil {
				return mapCbtRoomSetupError(err)
			}
		}
	}
	return nil
}

func cbtRoomEffectiveCapacity(room db.ListCbtExamRoomsRow) int32 {
	if room.CapacityOverride.Valid && room.CapacityOverride.Int32 > 0 {
		return room.CapacityOverride.Int32
	}
	return room.Capacity
}

func (s *CbtSession) ensureSessionSetupMutable(ctx context.Context, sessionID pgtype.UUID) error {
	session, err := s.q.GetCbtExamSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return ensureCbtSessionSetupStatusMutable(session.Status)
}

func ensureCbtSessionSetupStatusMutable(status db.CbtSessionStatusEnum) error {
	switch status {
	case db.CbtSessionStatusEnumDraft, db.CbtSessionStatusEnumScheduled:
		return nil
	default:
		return fmt.Errorf("%w: pengaturan ruangan, kursi, dan pengawas hanya boleh diubah saat sesi draft atau terjadwal", domain.ErrConflict)
	}
}

func mapCbtRoomSetupError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505":
		return fmt.Errorf("%w: data pengaturan ruang CBT sudah ada atau bertabrakan", domain.ErrConflict)
	case "23503":
		return fmt.Errorf("%w: referensi pengaturan ruang CBT tidak valid", domain.ErrBadRequest)
	case "23514":
		return fmt.Errorf("%w: nilai pengaturan ruang CBT tidak valid", domain.ErrBadRequest)
	default:
		return err
	}
}

// --- Proctoring ---

func (s *CbtSession) GetProctoringStatus(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error) {
	return s.GetProctoringStatusForRoom(ctx, sessionID, pgtype.UUID{})
}

func (s *CbtSession) GetProctoringStatusForRoom(ctx context.Context, sessionID, roomID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error) {
	rows, err := s.q.GetSessionProctoringStatus(ctx, db.GetSessionProctoringStatusParams{
		SessionID: sessionID,
		RoomID:    roomID,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetSessionProctoringStatusRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) ListParticipantEvents(ctx context.Context, sessionID, participantID pgtype.UUID, limit int32) ([]db.ListSessionParticipantEventsRow, error) {
	return s.ListParticipantEventsForRoom(ctx, sessionID, participantID, pgtype.UUID{}, limit)
}

func (s *CbtSession) ListParticipantEventsForRoom(ctx context.Context, sessionID, participantID, roomID pgtype.UUID, limit int32) ([]db.ListSessionParticipantEventsRow, error) {
	q, ok := s.q.(cbtParticipantEventStore)
	if !ok {
		return nil, fmt.Errorf("cbt participant event store unavailable")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := q.ListSessionParticipantEvents(ctx, db.ListSessionParticipantEventsParams{
		SessionID:     sessionID,
		ParticipantID: participantID,
		RoomID:        roomID,
		LimitCount:    limit,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListSessionParticipantEventsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) SetSuspiciousFlag(ctx context.Context, participantID pgtype.UUID, flag bool) error {
	return s.q.SetParticipantSuspiciousFlag(ctx, db.SetParticipantSuspiciousFlagParams{
		ID:             participantID,
		SuspiciousFlag: flag,
	})
}

func (s *CbtSession) ForceSubmitParticipant(ctx context.Context, sessionID, participantID pgtype.UUID, actor string) (db.ForceSubmitParticipantRow, error) {
	if s.pool == nil {
		q, ok := s.q.(cbtParticipantForceSubmitStore)
		if !ok {
			return db.ForceSubmitParticipantRow{}, fmt.Errorf("cbt participant force submit store unavailable")
		}
		return forceSubmitParticipant(ctx, q, sessionID, participantID, actor)
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return db.ForceSubmitParticipantRow{}, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return db.ForceSubmitParticipantRow{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := s.q.WithTx(tx)
	row, err := forceSubmitParticipant(ctx, qtx, sessionID, participantID, actor)
	if err != nil {
		return db.ForceSubmitParticipantRow{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return db.ForceSubmitParticipantRow{}, err
	}
	return row, nil
}

func forceSubmitParticipant(ctx context.Context, q cbtParticipantForceSubmitStore, sessionID, participantID pgtype.UUID, actor string) (db.ForceSubmitParticipantRow, error) {
	if err := q.UpdateParticipantAnswerCorrectness(ctx, participantID); err != nil {
		return db.ForceSubmitParticipantRow{}, err
	}
	row, err := q.ForceSubmitParticipant(ctx, db.ForceSubmitParticipantParams{
		SessionID: sessionID,
		ID:        participantID,
	})
	if err != nil {
		return db.ForceSubmitParticipantRow{}, err
	}
	if err := q.InsertParticipantEvent(ctx, db.InsertParticipantEventParams{
		ParticipantID: participantID,
		EventType:     "proctor_force_submit",
		EventData:     marshalJSON(map[string]string{"actor": actor}),
	}); err != nil {
		return db.ForceSubmitParticipantRow{}, err
	}
	return row, nil
}

func normalizeIncidentAction(action string) string {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "reviewed", "cleared", "warning_given", "locked", "submitted", "escalated":
		return strings.TrimSpace(strings.ToLower(action))
	default:
		return ""
	}
}

func normalizeParticipantCommand(commandType string) string {
	switch strings.TrimSpace(strings.ToLower(commandType)) {
	case ParticipantCommandWarningMessage:
		return ParticipantCommandWarningMessage
	case ParticipantCommandReconnect:
		return ParticipantCommandReconnect
	case ParticipantCommandUnlockNotice:
		return ParticipantCommandUnlockNotice
	default:
		return ""
	}
}

func defaultParticipantCommandMessage(commandType string) string {
	switch commandType {
	case ParticipantCommandReconnect:
		return "Silakan hubungi pengawas untuk login ulang setelah akses diverifikasi."
	case ParticipantCommandUnlockNotice:
		return "Akses ujian sudah dibuka. Lanjutkan hanya setelah pengawas memberi arahan."
	default:
		return "Tetap di aplikasi ujian dan ikuti arahan pengawas."
	}
}

func participantCommandSeverity(commandType string) string {
	switch commandType {
	case ParticipantCommandReconnect:
		return "warning"
	case ParticipantCommandUnlockNotice:
		return "info"
	default:
		return "warning"
	}
}

// --- Essay Grading ---

func (s *CbtSession) GradeEssay(ctx context.Context, sessionID, answerID pgtype.UUID, manualScore float64, gradedBy string) error {
	if s.pool == nil {
		return gradeEssayAndRefreshScore(ctx, s.q, sessionID, answerID, manualScore, gradedBy)
	}

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

	qtx := s.q.WithTx(tx)
	if err := gradeEssayAndRefreshScore(ctx, qtx, sessionID, answerID, manualScore, gradedBy); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func gradeEssayAndRefreshScore(ctx context.Context, q cbtSessionStore, sessionID, answerID pgtype.UUID, manualScore float64, gradedBy string) error {
	if err := q.GradeStudentEssay(ctx, db.GradeStudentEssayParams{
		ID:          answerID,
		ManualScore: pgNumeric(manualScore),
		GradedBy:    pgtype.Text{String: gradedBy, Valid: true},
	}); err != nil {
		return err
	}
	return scoreSession(ctx, q, sessionID)
}

func pgNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(strconv.FormatFloat(f, 'f', -1, 64))
	return n
}

func (s *CbtSession) ListUngradedEssays(ctx context.Context, sessionID pgtype.UUID) ([]db.ListUngradedEssaysRow, error) {
	rows, err := s.q.ListUngradedEssays(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListUngradedEssaysRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) ListUngradedEssaysByTeacher(ctx context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.ListUngradedEssaysRow, error) {
	rows, err := s.q.ListUngradedEssaysByTeacher(ctx, db.ListUngradedEssaysByTeacherParams{
		TeacherEmployeeID: teacherEmployeeID,
		SessionID:         sessionID,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListUngradedEssaysRow{}, nil
	}
	out := make([]db.ListUngradedEssaysRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, db.ListUngradedEssaysRow(row))
	}
	return out, nil
}

// --- Answers & Scoring ---

func (s *CbtSession) RecordAnswer(ctx context.Context, participantID, questionID pgtype.UUID, answer string) error {
	belongs, err := s.q.QuestionBelongsToParticipantPackage(ctx, db.QuestionBelongsToParticipantPackageParams{
		ID:         participantID,
		QuestionID: questionID,
	})
	if err != nil {
		return err
	}
	if !belongs {
		return ErrExamQuestionScope
	}
	rows, err := s.q.UpsertStudentAnswer(ctx, db.UpsertStudentAnswerParams{
		ParticipantID: participantID,
		QuestionID:    questionID,
		Answer:        answer,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrExamAlreadySubmit
	}
	return nil
}

// ScoreSession marks is_correct for all answers then updates participant scores in a transaction.
func (s *CbtSession) ScoreSession(ctx context.Context, sessionID pgtype.UUID) error {
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

	qtx := s.q.WithTx(tx)

	if err := scoreSession(ctx, qtx, sessionID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func scoreSession(ctx context.Context, q cbtScoreStore, sessionID pgtype.UUID) error {
	if err := q.UpdateAnswerCorrectness(ctx, sessionID); err != nil {
		return err
	}
	if err := q.UpdateParticipantScores(ctx, sessionID); err != nil {
		return err
	}
	return nil
}

type CbtFinalizeOverdueResult struct {
	SessionID      string `json:"session_id"`
	FinalizedCount int32  `json:"finalized_count"`
}

func (s *CbtSession) FinalizeOverdue(ctx context.Context, sessionID pgtype.UUID) (CbtFinalizeOverdueResult, error) {
	if s.pool == nil {
		q, ok := s.q.(cbtFinalizeOverdueStore)
		if !ok {
			return CbtFinalizeOverdueResult{}, fmt.Errorf("cbt finalize store unavailable")
		}
		count, err := finalizeOverdueWithStore(ctx, q, sessionID)
		if err != nil {
			return CbtFinalizeOverdueResult{}, err
		}
		return CbtFinalizeOverdueResult{SessionID: pgUUIDString(sessionID), FinalizedCount: count}, nil
	}

	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return CbtFinalizeOverdueResult{}, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return CbtFinalizeOverdueResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	count, err := finalizeOverdueWithStore(ctx, s.q.WithTx(tx), sessionID)
	if err != nil {
		return CbtFinalizeOverdueResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CbtFinalizeOverdueResult{}, err
	}
	return CbtFinalizeOverdueResult{SessionID: pgUUIDString(sessionID), FinalizedCount: count}, nil
}

func finalizeOverdueWithStore(ctx context.Context, q cbtFinalizeOverdueStore, sessionID pgtype.UUID) (int32, error) {
	if err := q.UpdateAnswerCorrectness(ctx, sessionID); err != nil {
		return 0, err
	}
	return q.FinalizeOverdueParticipants(ctx, sessionID)
}

func (s *CbtSession) GetGradeSyncPreflight(ctx context.Context, sessionID pgtype.UUID) (db.GetCbtSessionGradeSyncPreflightRow, error) {
	q, ok := s.q.(cbtResultFollowUpStore)
	if !ok {
		return db.GetCbtSessionGradeSyncPreflightRow{}, fmt.Errorf("cbt result follow-up store unavailable")
	}
	return q.GetCbtSessionGradeSyncPreflight(ctx, sessionID)
}

func (s *CbtSession) ListRemedialCandidates(ctx context.Context, sessionID pgtype.UUID, threshold float64) ([]db.ListCbtSessionRemedialCandidatesRow, error) {
	q, ok := s.q.(cbtResultFollowUpStore)
	if !ok {
		return nil, fmt.Errorf("cbt result follow-up store unavailable")
	}
	if threshold <= 0 || threshold > 100 {
		threshold = 75
	}
	rows, err := q.ListCbtSessionRemedialCandidates(ctx, db.ListCbtSessionRemedialCandidatesParams{
		SessionID: sessionID,
		Threshold: pgNumeric(threshold),
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtSessionRemedialCandidatesRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) ListByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamSessionsByTeacherRow, error) {
	rows, err := s.q.ListCbtExamSessionsByTeacher(ctx, teacherEmployeeID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamSessionsByTeacherRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) GetResultsByTeacher(ctx context.Context, sessionID, teacherEmployeeID pgtype.UUID) ([]db.GetSessionResultsByTeacherRow, error) {
	rows, err := s.q.GetSessionResultsByTeacher(ctx, db.GetSessionResultsByTeacherParams{
		SessionID:         sessionID,
		TeacherEmployeeID: teacherEmployeeID,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetSessionResultsByTeacherRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) CheckTeacherAccess(ctx context.Context, sessionID, teacherEmployeeID pgtype.UUID) (bool, error) {
	return s.q.GetSessionTeacherAccess(ctx, db.GetSessionTeacherAccessParams{
		ID:                sessionID,
		TeacherEmployeeID: teacherEmployeeID,
	})
}

func (s *CbtSession) GetResults(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionResultsRow, error) {
	rows, err := s.q.GetSessionResults(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetSessionResultsRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) GetItemAnalysis(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionItemAnalysisRow, error) {
	q, ok := s.q.(cbtItemAnalysisStore)
	if !ok {
		return nil, fmt.Errorf("cbt item analysis store unavailable")
	}
	rows, err := q.GetSessionItemAnalysis(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetSessionItemAnalysisRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error) {
	rows, err := s.q.GetParticipantAnswers(ctx, participantID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetParticipantAnswersRow{}, nil
	}
	return rows, nil
}

// shuffleUUIDs returns a new slice with UUIDs in random order using crypto/rand.
func shuffleUUIDs(ids []pgtype.UUID) []pgtype.UUID {
	out := make([]pgtype.UUID, len(ids))
	copy(out, ids)
	for i := len(out) - 1; i > 0; i-- {
		value, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return out
		}
		j := int(value.Int64())
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// cryptoPermInts returns [0,n) shuffled via Fisher–Yates seeded by crypto/rand.
func cryptoPermInts(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	for i := n - 1; i > 0; i-- {
		value, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return out
		}
		j := int(value.Int64())
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// UUIDsToJSON serialises a UUID slice as a JSON array of strings.
func UUIDsToJSON(ids []pgtype.UUID) ([]byte, error) {
	strs := make([]string, len(ids))
	for i, u := range ids {
		strs[i] = pgUUIDString(u)
	}
	return json.Marshal(strs)
}
