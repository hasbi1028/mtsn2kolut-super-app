package service

import (
	"context"
	"encoding/json"
	"math/rand"
	"sort"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtSession struct {
	q    *db.Queries
	pool *pgxpool.Pool
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
	return s.q.UpdateCbtExamSessionStatus(ctx, db.UpdateCbtExamSessionStatusParams{
		ID:     id,
		Status: status,
	})
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
	return s.q.CreateCbtExamRoom(ctx, db.CreateCbtExamRoomParams{
		SessionID: sessionID,
		RoomName:  roomName,
		Capacity:  capacity,
	})
}

func (s *CbtSession) DeleteRoom(ctx context.Context, roomID pgtype.UUID) error {
	return s.q.DeleteCbtExamRoom(ctx, roomID)
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

	// Clear existing room assignments
	if err := qtx.ClearParticipantRooms(ctx, sessionID); err != nil {
		return err
	}

	// Load participants and rooms
	participants, err := qtx.ListParticipantsByRoom(ctx, sessionID)
	if err != nil {
		return err
	}
	rooms, err := qtx.ListCbtExamRooms(ctx, sessionID)
	if err != nil {
		return err
	}
	if len(rooms) == 0 || len(participants) == 0 {
		return tx.Commit(ctx)
	}

	// Shuffle participants randomly
	indices := rand.Perm(len(participants))

	slot := 0
	for _, room := range rooms {
		for range int(room.Capacity) {
			if slot >= len(indices) {
				break
			}
			p := participants[indices[slot]]
			if err := qtx.AssignParticipantRoom(ctx, db.AssignParticipantRoomParams{
				ID:     p.ID,
				RoomID: room.ID,
			}); err != nil {
				return err
			}
			slot++
		}
	}

	return tx.Commit(ctx)
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
	rows, err := s.q.GetSessionProctoringStatus(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetSessionProctoringStatusRow{}, nil
	}
	return rows, nil
}

func (s *CbtSession) SetSuspiciousFlag(ctx context.Context, participantID pgtype.UUID, flag bool) error {
	return s.q.SetParticipantSuspiciousFlag(ctx, db.SetParticipantSuspiciousFlagParams{
		ID:             participantID,
		SuspiciousFlag: flag,
	})
}

// --- Essay Grading ---

func (s *CbtSession) GradeEssay(ctx context.Context, answerID pgtype.UUID, manualScore float64, gradedBy string) error {
	return s.q.GradeStudentEssay(ctx, db.GradeStudentEssayParams{
		ID:          answerID,
		ManualScore: pgNumeric(manualScore),
		GradedBy:    pgtype.Text{String: gradedBy, Valid: true},
	})
}

func pgNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(f)
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

	if err := qtx.UpdateAnswerCorrectness(ctx, sessionID); err != nil {
		return err
	}
	if err := qtx.UpdateParticipantScores(ctx, sessionID); err != nil {
		return err
	}

	return tx.Commit(ctx)
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
