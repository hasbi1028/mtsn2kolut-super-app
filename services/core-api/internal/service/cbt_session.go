package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtSession struct {
	q    cbtSessionStore
	pool *pgxpool.Pool
}

type cbtSessionStore interface {
	ListCbtExamSessions(ctx context.Context) ([]db.ListCbtExamSessionsRow, error)
	GetCbtExamSession(ctx context.Context, id pgtype.UUID) (db.GetCbtExamSessionRow, error)
	CreateCbtExamSession(ctx context.Context, arg db.CreateCbtExamSessionParams) (db.CbtExamSession, error)
	UpdateCbtExamSessionStatus(ctx context.Context, arg db.UpdateCbtExamSessionStatusParams) (db.CbtExamSession, error)
	DeleteCbtExamSession(ctx context.Context, id pgtype.UUID) error
	ListCbtExamParticipants(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamParticipantsRow, error)
	EnrollClassToSession(ctx context.Context, arg db.EnrollClassToSessionParams) error
	EnrollGradeToSession(ctx context.Context, arg db.EnrollGradeToSessionParams) error
	EnrollSchoolToSession(ctx context.Context, sessionID pgtype.UUID) error
	GenerateTokensForSession(ctx context.Context, sessionID pgtype.UUID) error
	RegenerateParticipantToken(ctx context.Context, id pgtype.UUID) (db.RegenerateParticipantTokenRow, error)
	ListCbtExamRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error)
	CreateCbtExamRoom(ctx context.Context, arg db.CreateCbtExamRoomParams) (db.CbtExamRoom, error)
	DeleteCbtExamRoom(ctx context.Context, id pgtype.UUID) error
	GetSchoolRoom(ctx context.Context, id pgtype.UUID) (db.SchoolRoom, error)
	GetCbtRoomProctorDashboard(ctx context.Context, id pgtype.UUID) (db.GetCbtRoomProctorDashboardRow, error)
	ListCbtProctorRooms(ctx context.Context, arg db.ListCbtProctorRoomsParams) ([]db.ListCbtProctorRoomsRow, error)
	ListCbtRoomProctors(ctx context.Context, examRoomID pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error)
	DeleteCbtRoomProctorsByRoom(ctx context.Context, examRoomID pgtype.UUID) error
	CreateCbtRoomProctor(ctx context.Context, arg db.CreateCbtRoomProctorParams) (db.CbtRoomProctor, error)
	GetCbtSessionRoomReadiness(ctx context.Context, targetSessionID pgtype.UUID) (db.GetCbtSessionRoomReadinessRow, error)
	HasSessionRoomProctor(ctx context.Context, arg db.HasSessionRoomProctorParams) (bool, error)
	HasSessionRoomParticipant(ctx context.Context, arg db.HasSessionRoomParticipantParams) (bool, error)
	AssignParticipantSeat(ctx context.Context, arg db.AssignParticipantSeatParams) error
	ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error)
	GetSessionProctoringStatus(ctx context.Context, arg db.GetSessionProctoringStatusParams) ([]db.GetSessionProctoringStatusRow, error)
	SetParticipantSuspiciousFlag(ctx context.Context, arg db.SetParticipantSuspiciousFlagParams) error
	GradeStudentEssay(ctx context.Context, arg db.GradeStudentEssayParams) error
	ListUngradedEssays(ctx context.Context, sessionID pgtype.UUID) ([]db.ListUngradedEssaysRow, error)
	UpsertStudentAnswer(ctx context.Context, arg db.UpsertStudentAnswerParams) error
	UpdateAnswerCorrectness(ctx context.Context, sessionID pgtype.UUID) error
	UpdateParticipantScores(ctx context.Context, sessionID pgtype.UUID) error
	ListCbtExamSessionsByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) ([]db.ListCbtExamSessionsByTeacherRow, error)
	GetSessionResultsByTeacher(ctx context.Context, arg db.GetSessionResultsByTeacherParams) ([]db.GetSessionResultsByTeacherRow, error)
	GetSessionTeacherAccess(ctx context.Context, arg db.GetSessionTeacherAccessParams) (bool, error)
	HasSessionParticipant(ctx context.Context, arg db.HasSessionParticipantParams) (bool, error)
	HasSessionRoom(ctx context.Context, arg db.HasSessionRoomParams) (bool, error)
	HasSessionAnswer(ctx context.Context, arg db.HasSessionAnswerParams) (bool, error)
	GetSessionResults(ctx context.Context, sessionID pgtype.UUID) ([]db.GetSessionResultsRow, error)
	GetParticipantAnswers(ctx context.Context, participantID pgtype.UUID) ([]db.GetParticipantAnswersRow, error)
	WithTx(tx pgx.Tx) *db.Queries
}

type cbtRoomShuffleStore interface {
	ClearParticipantRooms(ctx context.Context, sessionID pgtype.UUID) error
	ListParticipantsByRoom(ctx context.Context, sessionID pgtype.UUID) ([]db.ListParticipantsByRoomRow, error)
	ListCbtExamRooms(ctx context.Context, sessionID pgtype.UUID) ([]db.ListCbtExamRoomsRow, error)
	AssignParticipantRoom(ctx context.Context, arg db.AssignParticipantRoomParams) error
}

type cbtScoreStore interface {
	UpdateAnswerCorrectness(ctx context.Context, sessionID pgtype.UUID) error
	UpdateParticipantScores(ctx context.Context, sessionID pgtype.UUID) error
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

func (s *CbtSession) UpdateStatus(ctx context.Context, id pgtype.UUID, status db.CbtSessionStatusEnum) (db.CbtExamSession, error) {
	if status == db.CbtSessionStatusEnumActive {
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

func validateCbtSessionActivationReadiness(readiness db.GetCbtSessionRoomReadinessRow) error {
	switch {
	case readiness.ParticipantCount == 0:
		return fmt.Errorf("%w: sesi belum memiliki peserta", domain.ErrConflict)
	case readiness.RoomCount == 0:
		return fmt.Errorf("%w: sesi belum memiliki ruangan ujian", domain.ErrConflict)
	case readiness.TotalCapacity < readiness.ParticipantCount:
		return fmt.Errorf("%w: kapasitas ruangan belum cukup untuk seluruh peserta", domain.ErrConflict)
	case readiness.UnassignedParticipantCount > 0:
		return fmt.Errorf("%w: masih ada %d peserta belum mendapat ruangan", domain.ErrConflict, readiness.UnassignedParticipantCount)
	case readiness.MissingSeatCount > 0:
		return fmt.Errorf("%w: masih ada %d peserta belum mendapat nomor meja", domain.ErrConflict, readiness.MissingSeatCount)
	case readiness.RoomsWithoutProctor > 0:
		return fmt.Errorf("%w: masih ada %d ruangan belum punya pengawas", domain.ErrConflict, readiness.RoomsWithoutProctor)
	default:
		return nil
	}
}

func (s *CbtSession) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteCbtExamSession(ctx, id)
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
	return s.q.GenerateTokensForSession(ctx, sessionID)
}

type RegenerateTokenResult struct {
	ID    pgtype.UUID `json:"id"`
	Token string      `json:"token"`
}

func (s *CbtSession) RegenerateToken(ctx context.Context, participantID pgtype.UUID) (db.RegenerateParticipantTokenRow, error) {
	return s.q.RegenerateParticipantToken(ctx, participantID)
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
	roomName = strings.TrimSpace(roomName)
	if capacity <= 0 {
		capacity = 30
	}
	return s.q.CreateCbtExamRoom(ctx, db.CreateCbtExamRoomParams{
		SessionID:        sessionID,
		SchoolRoomID:     pgtype.UUID{},
		RoomName:         roomName,
		RoomNameSnapshot: roomName,
		Capacity:         capacity,
	})
}

func (s *CbtSession) CreateRoomFromSchoolRoom(ctx context.Context, sessionID, schoolRoomID pgtype.UUID, roomName string, capacity int32) (db.CbtExamRoom, error) {
	schoolRoom, err := s.q.GetSchoolRoom(ctx, schoolRoomID)
	if err != nil {
		return db.CbtExamRoom{}, err
	}
	if !schoolRoom.IsExamEligible || schoolRoom.Condition == "rusak" {
		return db.CbtExamRoom{}, fmt.Errorf("ruangan fisik belum layak dipakai untuk ujian")
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
	return s.q.CreateCbtExamRoom(ctx, db.CreateCbtExamRoomParams{
		SessionID:        sessionID,
		SchoolRoomID:     schoolRoomID,
		RoomName:         roomName,
		RoomNameSnapshot: roomName,
		Capacity:         capacity,
	})
}

func (s *CbtSession) DeleteRoom(ctx context.Context, roomID pgtype.UUID) error {
	return s.q.DeleteCbtExamRoom(ctx, roomID)
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
		return nil, err
	}

	ordered := normalizeRoomProctorIDs(primaryEmployeeID, employeeIDs)
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
			return nil, err
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
	return s.q.AssignParticipantSeat(ctx, db.AssignParticipantSeatParams{
		ID:     participantID,
		RoomID: roomID,
		SeatNo: pgtype.Int4{Int32: seatNo, Valid: seatNo > 0},
	})
}

// ShuffleRooms randomly assigns participants to rooms respecting capacity.
// If a room is full, remaining participants are left unassigned.
func (s *CbtSession) ShuffleRooms(ctx context.Context, sessionID pgtype.UUID) error {
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
		for range int(room.Capacity) {
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
	participants, err := s.q.ListParticipantsByRoom(ctx, sessionID)
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
			if err := s.AssignSeat(ctx, participant.ID, participant.RoomID, int32(idx+1)); err != nil {
				return err
			}
		}
	}
	return nil
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

// --- Answers & Scoring ---

func (s *CbtSession) RecordAnswer(ctx context.Context, participantID, questionID pgtype.UUID, answer string) error {
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
