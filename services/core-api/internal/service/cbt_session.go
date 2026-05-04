package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
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
	ClearParticipantSeatsForSession(ctx context.Context, sessionID pgtype.UUID) error
	HasOverlappingCbtRoomProctor(ctx context.Context, arg db.HasOverlappingCbtRoomProctorParams) (bool, error)
	ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error)
	GetSessionProctoringStatus(ctx context.Context, arg db.GetSessionProctoringStatusParams) ([]db.GetSessionProctoringStatusRow, error)
	SetParticipantSuspiciousFlag(ctx context.Context, arg db.SetParticipantSuspiciousFlagParams) error
	GradeStudentEssay(ctx context.Context, arg db.GradeStudentEssayParams) error
	ListUngradedEssays(ctx context.Context, sessionID pgtype.UUID) ([]db.ListUngradedEssaysRow, error)
	QuestionBelongsToParticipantPackage(ctx context.Context, arg db.QuestionBelongsToParticipantPackageParams) (bool, error)
	UpsertStudentAnswer(ctx context.Context, arg db.UpsertStudentAnswerParams) error
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
	ClearParticipantRooms(ctx context.Context, sessionID pgtype.UUID) error
	ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error)
	ListCbtExamRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error)
	AssignParticipantRoom(ctx context.Context, arg db.AssignParticipantRoomParams) error
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
	ListSessionParticipantEvents(ctx context.Context, arg db.ListSessionParticipantEventsParams) ([]db.ListSessionParticipantEventsRow, error)
	InsertParticipantEvent(ctx context.Context, arg db.InsertParticipantEventParams) error
}

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
	switch {
	case eventID.Valid && !sameUUID(pkg.EventID, eventID):
		return fmt.Errorf("%w: paket sesi harus berasal dari event yang sama", domain.ErrBadRequest)
	case !eventID.Valid && pkg.EventID.Valid:
		return fmt.Errorf("%w: paket khusus event hanya boleh dipakai pada sesi event yang sama", domain.ErrBadRequest)
	default:
		return nil
	}
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
	return s.q.EnrollClassToSession(ctx, db.EnrollClassToSessionParams{
		SessionID: sessionID,
		ClassID:   classID,
	})
}

func (s *CbtSession) EnrollGrade(ctx context.Context, sessionID pgtype.UUID, level string) error {
	return s.q.EnrollGradeToSession(ctx, db.EnrollGradeToSessionParams{
		SessionID: sessionID,
		Level:     level,
	})
}

func (s *CbtSession) EnrollSchool(ctx context.Context, sessionID pgtype.UUID) error {
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

	indices := rand.Perm(len(participants))

	slot := 0
	for _, room := range rooms {
		for range int(cbtRoomEffectiveCapacity(room)) {
			if slot >= len(indices) {
				break
			}
			p := participants[indices[slot]]
			if err := q.AssignParticipantRoom(ctx, db.AssignParticipantRoomParams{
				ID:     p.ID,
				RoomID: room.ID,
			}); err != nil {
				return err
			}
			slot++
		}
	}

	return nil
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
	return s.q.UpsertStudentAnswer(ctx, db.UpsertStudentAnswerParams{
		ParticipantID: participantID,
		QuestionID:    questionID,
		Answer:        answer,
	})
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

// shuffleUUIDs returns a new slice with UUIDs in random order.
func shuffleUUIDs(ids []pgtype.UUID) []pgtype.UUID {
	out := make([]pgtype.UUID, len(ids))
	copy(out, ids)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
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
