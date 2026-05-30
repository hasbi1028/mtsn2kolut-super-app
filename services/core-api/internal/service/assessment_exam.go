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

const (
	AssessmentExamStatusDraft    = "draft"
	AssessmentExamStatusReady    = "ready"
	AssessmentExamStatusRunning  = "running"
	AssessmentExamStatusFinished = "finished"
	AssessmentExamStatusArchived = "archived"
)

var validAssessmentExamStatuses = map[string]bool{
	AssessmentExamStatusDraft:    true,
	AssessmentExamStatusReady:    true,
	AssessmentExamStatusRunning:  true,
	AssessmentExamStatusFinished: true,
	AssessmentExamStatusArchived: true,
}

type assessmentExamStore interface {
	ListAssessmentExams(ctx context.Context, arg db.ListAssessmentExamsParams) ([]db.ListAssessmentExamsRow, error)
	GetAssessmentExam(ctx context.Context, id pgtype.UUID) (db.GetAssessmentExamRow, error)
	CreateAssessmentExam(ctx context.Context, arg db.CreateAssessmentExamParams) (db.AssessmentExam, error)
	UpdateAssessmentExam(ctx context.Context, arg db.UpdateAssessmentExamParams) (db.AssessmentExam, error)
	GetFirstAssessmentSessionByExam(ctx context.Context, examID pgtype.UUID) (db.AssessmentSession, error)
	CreateAssessmentSession(ctx context.Context, arg db.CreateAssessmentSessionParams) (db.AssessmentSession, error)
	CreateAssessmentRoom(ctx context.Context, arg db.CreateAssessmentRoomParams) (db.AssessmentRoom, error)
	UpsertAssessmentRoom(ctx context.Context, arg db.UpsertAssessmentRoomParams) (db.AssessmentRoom, error)
	CountAssessmentRoomsByExam(ctx context.Context, examID pgtype.UUID) (int64, error)
	CountAssessmentParticipantsByExam(ctx context.Context, examID pgtype.UUID) (int64, error)
	CountAssessmentCardsByExam(ctx context.Context, examID pgtype.UUID) (int64, error)
}

type AssessmentExam struct {
	q assessmentExamStore
}

func NewAssessmentExam(q *db.Queries) *AssessmentExam {
	return &AssessmentExam{q: q}
}

func NewAssessmentExamWithStore(q assessmentExamStore) *AssessmentExam {
	return &AssessmentExam{q: q}
}

type AssessmentExamInput struct {
	Title      string
	SubjectID  pgtype.UUID
	GradeLevel pgtype.Int2
	Status     string
	StartsAt   *time.Time
	EndsAt     *time.Time
	ActorID    pgtype.UUID
}

type AssessmentExamView struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	SubjectID        string     `json:"subject_id,omitempty"`
	GradeLevel       *int16     `json:"grade_level,omitempty"`
	Status           string     `json:"status"`
	StartsAt         *time.Time `json:"starts_at,omitempty"`
	EndsAt           *time.Time `json:"ends_at,omitempty"`
	CreatedBy        string     `json:"created_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	SessionCount     int64      `json:"session_count"`
	RoomCount        int64      `json:"room_count"`
	ParticipantCount int64      `json:"participant_count"`
	CardCount        int64      `json:"card_count,omitempty"`
}

type AssessmentPrepareRoomsResult struct {
	ExamID       string `json:"exam_id"`
	SessionID    string `json:"session_id"`
	RoomID       string `json:"room_id"`
	RoomCount    int64  `json:"room_count"`
	Message      string `json:"message"`
	AlreadyReady bool   `json:"already_ready"`
}

const (
	AssessmentMixPolicyMixed        = "mixed"
	AssessmentMixPolicyClassGrouped = "class_grouped"
)

type AssessmentIssueCardsResult struct {
	ExamID           string `json:"exam_id"`
	RoomCount        int64  `json:"room_count"`
	ParticipantCount int64  `json:"participant_count"`
	CardCount        int64  `json:"card_count"`
	Message          string `json:"message"`
}

type AssessmentAssignmentRequest struct {
	RoomCount       int32    `json:"room_count"`
	CapacityPerRoom int32    `json:"capacity_per_room"`
	ClassIDs        []string `json:"class_ids,omitempty"`
	MixPolicy       string   `json:"mix_policy,omitempty"`
}

type AssessmentAssignmentRoom struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Capacity      int32  `json:"capacity"`
	AssignedCount int64  `json:"assigned_count"`
}

type AssessmentAssignmentResult struct {
	ExamID            string                     `json:"exam_id"`
	SessionID         string                     `json:"session_id,omitempty"`
	MixPolicy         string                     `json:"mix_policy"`
	RoomCount         int32                      `json:"room_count"`
	CapacityPerRoom   int32                      `json:"capacity_per_room"`
	TotalParticipants int64                      `json:"total_participants"`
	AssignedTotal     int64                      `json:"assigned_total"`
	UnassignedTotal   int64                      `json:"unassigned_total"`
	ParticipantCount  int64                      `json:"participant_count"`
	CardCount         int64                      `json:"card_count"`
	Rooms             []AssessmentAssignmentRoom `json:"rooms"`
	Message           string                     `json:"message"`
	Applied           bool                       `json:"applied"`
}

func (s *AssessmentExam) List(ctx context.Context, search, status string, limit, offset int32) ([]AssessmentExamView, error) {
	status = strings.TrimSpace(status)
	if status != "" && !validAssessmentExamStatuses[status] {
		return nil, fmt.Errorf("%w: status asesmen tidak valid", domain.ErrBadRequest)
	}
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.q.ListAssessmentExams(ctx, db.ListAssessmentExamsParams{
		Search:       strings.TrimSpace(search),
		StatusFilter: status,
		LimitCount:   limit,
		OffsetCount:  offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]AssessmentExamView, 0, len(rows))
	for _, row := range rows {
		items = append(items, assessmentExamListRowView(row))
	}
	return items, nil
}

func (s *AssessmentExam) Get(ctx context.Context, id pgtype.UUID) (AssessmentExamView, error) {
	if !id.Valid {
		return AssessmentExamView{}, fmt.Errorf("%w: id asesmen tidak valid", domain.ErrBadRequest)
	}
	row, err := s.q.GetAssessmentExam(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return AssessmentExamView{}, domain.ErrNotFound
	}
	if err != nil {
		return AssessmentExamView{}, err
	}
	return assessmentExamGetRowView(row), nil
}

func (s *AssessmentExam) Create(ctx context.Context, input AssessmentExamInput) (AssessmentExamView, error) {
	normalized, err := normalizeAssessmentExamInput(input, true)
	if err != nil {
		return AssessmentExamView{}, err
	}
	row, err := s.q.CreateAssessmentExam(ctx, db.CreateAssessmentExamParams{
		Title:      normalized.Title,
		SubjectID:  normalized.SubjectID,
		GradeLevel: normalized.GradeLevel,
		StartsAt:   assessmentTimestamptzFromPtr(normalized.StartsAt),
		EndsAt:     assessmentTimestamptzFromPtr(normalized.EndsAt),
		CreatedBy:  normalized.ActorID,
	})
	if err != nil {
		return AssessmentExamView{}, err
	}
	return s.Get(ctx, row.ID)
}

func (s *AssessmentExam) Update(ctx context.Context, id pgtype.UUID, input AssessmentExamInput) (AssessmentExamView, error) {
	if !id.Valid {
		return AssessmentExamView{}, fmt.Errorf("%w: id asesmen tidak valid", domain.ErrBadRequest)
	}
	normalized, err := normalizeAssessmentExamInput(input, false)
	if err != nil {
		return AssessmentExamView{}, err
	}
	row, err := s.q.UpdateAssessmentExam(ctx, db.UpdateAssessmentExamParams{
		ID:         id,
		Title:      normalized.Title,
		SubjectID:  normalized.SubjectID,
		GradeLevel: normalized.GradeLevel,
		Status:     normalized.Status,
		StartsAt:   assessmentTimestamptzFromPtr(normalized.StartsAt),
		EndsAt:     assessmentTimestamptzFromPtr(normalized.EndsAt),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return AssessmentExamView{}, domain.ErrNotFound
	}
	if err != nil {
		return AssessmentExamView{}, err
	}
	return s.Get(ctx, row.ID)
}

func (s *AssessmentExam) PrepareRooms(ctx context.Context, id pgtype.UUID) (AssessmentPrepareRoomsResult, error) {
	exam, err := s.Get(ctx, id)
	if err != nil {
		return AssessmentPrepareRoomsResult{}, err
	}
	session, err := s.q.GetFirstAssessmentSessionByExam(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		session, err = s.q.CreateAssessmentSession(ctx, db.CreateAssessmentSessionParams{
			ExamID:   id,
			Title:    "Sesi Utama",
			StartsAt: assessmentTimestamptzFromPtr(exam.StartsAt),
			EndsAt:   assessmentTimestamptzFromPtr(exam.EndsAt),
		})
	}
	if err != nil {
		return AssessmentPrepareRoomsResult{}, err
	}
	room, err := s.q.CreateAssessmentRoom(ctx, db.CreateAssessmentRoomParams{
		SessionID: session.ID,
		Code:      "R01",
		Name:      "Ruang Ujian 1",
		Capacity:  30,
	})
	if err != nil {
		return AssessmentPrepareRoomsResult{}, err
	}
	roomCount, err := s.q.CountAssessmentRoomsByExam(ctx, id)
	if err != nil {
		return AssessmentPrepareRoomsResult{}, err
	}
	return AssessmentPrepareRoomsResult{
		ExamID:       exam.ID,
		SessionID:    assessmentUUIDString(session.ID),
		RoomID:       assessmentUUIDString(room.ID),
		RoomCount:    roomCount,
		Message:      "Fondasi ruang ujian awal sudah siap. Pembagian peserta otomatis menyusul pada tahap berikutnya.",
		AlreadyReady: roomCount > 1,
	}, nil
}

func (s *AssessmentExam) IssueCards(ctx context.Context, id pgtype.UUID) (AssessmentIssueCardsResult, error) {
	exam, err := s.Get(ctx, id)
	if err != nil {
		return AssessmentIssueCardsResult{}, err
	}
	roomCount, err := s.q.CountAssessmentRoomsByExam(ctx, id)
	if err != nil {
		return AssessmentIssueCardsResult{}, err
	}
	participantCount, err := s.q.CountAssessmentParticipantsByExam(ctx, id)
	if err != nil {
		return AssessmentIssueCardsResult{}, err
	}
	cardCount, err := s.q.CountAssessmentCardsByExam(ctx, id)
	if err != nil {
		return AssessmentIssueCardsResult{}, err
	}
	return AssessmentIssueCardsResult{
		ExamID:           exam.ID,
		RoomCount:        roomCount,
		ParticipantCount: participantCount,
		CardCount:        cardCount,
		Message:          "Penerbitan QR+PIN belum diaktifkan sampai peserta ujian tersambung. Tidak ada token mentah yang dibuat pada tahap ini.",
	}, nil
}

func (s *AssessmentExam) AssignmentPreview(ctx context.Context, id pgtype.UUID, input AssessmentAssignmentRequest) (AssessmentAssignmentResult, error) {
	exam, err := s.Get(ctx, id)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	normalized, err := normalizeAssessmentAssignmentRequest(input)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	participantCount, err := s.q.CountAssessmentParticipantsByExam(ctx, id)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	result := buildAssessmentAssignmentResult(exam.ID, "", normalized, participantCount, false)
	return result, nil
}

func (s *AssessmentExam) AssignmentApply(ctx context.Context, id pgtype.UUID, input AssessmentAssignmentRequest) (AssessmentAssignmentResult, error) {
	exam, err := s.Get(ctx, id)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	normalized, err := normalizeAssessmentAssignmentRequest(input)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	session, err := s.q.GetFirstAssessmentSessionByExam(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		session, err = s.q.CreateAssessmentSession(ctx, db.CreateAssessmentSessionParams{
			ExamID:   id,
			Title:    "Sesi Utama",
			StartsAt: assessmentTimestamptzFromPtr(exam.StartsAt),
			EndsAt:   assessmentTimestamptzFromPtr(exam.EndsAt),
		})
	}
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	participantCount, err := s.q.CountAssessmentParticipantsByExam(ctx, id)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	result := buildAssessmentAssignmentResult(exam.ID, assessmentUUIDString(session.ID), normalized, participantCount, true)
	for _, room := range result.Rooms {
		if _, err := s.q.UpsertAssessmentRoom(ctx, db.UpsertAssessmentRoomParams{
			SessionID: session.ID,
			Code:      room.Code,
			Name:      room.Name,
			Capacity:  room.Capacity,
		}); err != nil {
			return AssessmentAssignmentResult{}, err
		}
	}
	cardCount, err := s.q.CountAssessmentCardsByExam(ctx, id)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	result.CardCount = cardCount
	result.Message = "Ruang ujian tersimpan. Peserta dan kartu ujian belum dibuat pada tahap fondasi ini."
	return result, nil
}

func normalizeAssessmentAssignmentRequest(input AssessmentAssignmentRequest) (AssessmentAssignmentRequest, error) {
	if input.RoomCount < 1 || input.RoomCount > 20 {
		return input, fmt.Errorf("%w: jumlah ruang harus 1 sampai 20", domain.ErrBadRequest)
	}
	if input.CapacityPerRoom < 1 || input.CapacityPerRoom > 50 {
		return input, fmt.Errorf("%w: kapasitas per ruang harus 1 sampai 50", domain.ErrBadRequest)
	}
	input.MixPolicy = strings.TrimSpace(input.MixPolicy)
	if input.MixPolicy == "" {
		input.MixPolicy = AssessmentMixPolicyMixed
	}
	if input.MixPolicy != AssessmentMixPolicyMixed && input.MixPolicy != AssessmentMixPolicyClassGrouped {
		return input, fmt.Errorf("%w: kebijakan campur peserta tidak valid", domain.ErrBadRequest)
	}
	return input, nil
}

func buildAssessmentAssignmentResult(examID, sessionID string, input AssessmentAssignmentRequest, totalParticipants int64, applied bool) AssessmentAssignmentResult {
	rooms := make([]AssessmentAssignmentRoom, 0, input.RoomCount)
	remaining := totalParticipants
	for i := int32(1); i <= input.RoomCount; i++ {
		assigned := int64(0)
		if remaining > 0 {
			assigned = int64(input.CapacityPerRoom)
			if remaining < assigned {
				assigned = remaining
			}
			remaining -= assigned
		}
		code := fmt.Sprintf("R%02d", i)
		rooms = append(rooms, AssessmentAssignmentRoom{
			Code:          code,
			Name:          fmt.Sprintf("Ruang Ujian %d", i),
			Capacity:      input.CapacityPerRoom,
			AssignedCount: assigned,
		})
	}
	assignedTotal := totalParticipants - remaining
	message := "Preview ruang ujian siap. Belum ada peserta terdaftar pada asesmen ini."
	if totalParticipants > 0 {
		message = "Preview ruang ujian siap. Simpan untuk membuat ruang; pembagian peserta detail menyusul pada tahap berikutnya."
	}
	return AssessmentAssignmentResult{
		ExamID:            examID,
		SessionID:         sessionID,
		MixPolicy:         input.MixPolicy,
		RoomCount:         input.RoomCount,
		CapacityPerRoom:   input.CapacityPerRoom,
		TotalParticipants: totalParticipants,
		AssignedTotal:     assignedTotal,
		UnassignedTotal:   remaining,
		ParticipantCount:  totalParticipants,
		Rooms:             rooms,
		Message:           message,
		Applied:           applied,
	}
}

func normalizeAssessmentExamInput(input AssessmentExamInput, create bool) (AssessmentExamInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return input, fmt.Errorf("%w: nama ujian wajib diisi", domain.ErrBadRequest)
	}
	if input.GradeLevel.Valid && (input.GradeLevel.Int16 < 7 || input.GradeLevel.Int16 > 9) {
		return input, fmt.Errorf("%w: tingkat kelas harus 7 sampai 9", domain.ErrBadRequest)
	}
	if input.StartsAt != nil && input.EndsAt != nil && !input.EndsAt.After(*input.StartsAt) {
		return input, fmt.Errorf("%w: jam selesai harus setelah jam mulai", domain.ErrBadRequest)
	}
	input.Status = strings.TrimSpace(input.Status)
	if create || input.Status == "" {
		input.Status = AssessmentExamStatusDraft
	}
	if !validAssessmentExamStatuses[input.Status] {
		return input, fmt.Errorf("%w: status asesmen tidak valid", domain.ErrBadRequest)
	}
	return input, nil
}

func assessmentExamListRowView(row db.ListAssessmentExamsRow) AssessmentExamView {
	return AssessmentExamView{
		ID:               assessmentUUIDString(row.ID),
		Title:            row.Title,
		SubjectID:        assessmentUUIDString(row.SubjectID),
		GradeLevel:       assessmentInt2Ptr(row.GradeLevel),
		Status:           row.Status,
		StartsAt:         assessmentTimestamptzPtr(row.StartsAt),
		EndsAt:           assessmentTimestamptzPtr(row.EndsAt),
		CreatedBy:        assessmentUUIDString(row.CreatedBy),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		SessionCount:     row.SessionCount,
		RoomCount:        row.RoomCount,
		ParticipantCount: row.ParticipantCount,
	}
}

func assessmentExamGetRowView(row db.GetAssessmentExamRow) AssessmentExamView {
	view := AssessmentExamView{
		ID:               assessmentUUIDString(row.ID),
		Title:            row.Title,
		SubjectID:        assessmentUUIDString(row.SubjectID),
		GradeLevel:       assessmentInt2Ptr(row.GradeLevel),
		Status:           row.Status,
		StartsAt:         assessmentTimestamptzPtr(row.StartsAt),
		EndsAt:           assessmentTimestamptzPtr(row.EndsAt),
		CreatedBy:        assessmentUUIDString(row.CreatedBy),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		SessionCount:     row.SessionCount,
		RoomCount:        row.RoomCount,
		ParticipantCount: row.ParticipantCount,
		CardCount:        row.CardCount,
	}
	return view
}

func assessmentTimestamptzFromPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}

func assessmentTimestamptzPtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func assessmentInt2Ptr(value pgtype.Int2) *int16 {
	if !value.Valid {
		return nil
	}
	return &value.Int16
}

func assessmentUUIDString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return value.String()
}
