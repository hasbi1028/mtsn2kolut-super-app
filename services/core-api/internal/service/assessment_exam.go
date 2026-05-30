package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

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
	ListAssessmentCandidateStudentsByClassIDs(ctx context.Context, classIds []pgtype.UUID) ([]db.ListAssessmentCandidateStudentsByClassIDsRow, error)
	UpsertAssessmentParticipant(ctx context.Context, arg db.UpsertAssessmentParticipantParams) (db.AssessmentParticipant, error)
	ListAssessmentParticipantsForAssignment(ctx context.Context, sessionID pgtype.UUID) ([]db.ListAssessmentParticipantsForAssignmentRow, error)
	ClearAssessmentParticipantRooms(ctx context.Context, sessionID pgtype.UUID) error
	AssignAssessmentParticipantRoom(ctx context.Context, arg db.AssignAssessmentParticipantRoomParams) error
	CountAssessmentRoomsByExam(ctx context.Context, examID pgtype.UUID) (int64, error)
	CountAssessmentParticipantsByExam(ctx context.Context, examID pgtype.UUID) (int64, error)
	CountAssessmentCardsByExam(ctx context.Context, examID pgtype.UUID) (int64, error)
	ListAssessmentParticipantPlacementsByExam(ctx context.Context, examID pgtype.UUID) ([]db.ListAssessmentParticipantPlacementsByExamRow, error)
	GetAssessmentRoomByIDAndExam(ctx context.Context, arg db.GetAssessmentRoomByIDAndExamParams) (db.GetAssessmentRoomByIDAndExamRow, error)
	MoveAssessmentParticipantSeat(ctx context.Context, arg db.MoveAssessmentParticipantSeatParams) (db.MoveAssessmentParticipantSeatRow, error)
	ListAssessmentParticipantCardTargetsByExam(ctx context.Context, examID pgtype.UUID) ([]db.ListAssessmentParticipantCardTargetsByExamRow, error)
	CreateAssessmentParticipantAccessCard(ctx context.Context, arg db.CreateAssessmentParticipantAccessCardParams) (db.AssessmentAccessCard, error)
	ListAssessmentExamPackageMaps(ctx context.Context, examID pgtype.UUID) ([]db.ListAssessmentExamPackageMapsRow, error)
	UpsertAssessmentExamPackageMap(ctx context.Context, arg db.UpsertAssessmentExamPackageMapParams) (db.AssessmentExamPackageMap, error)
	DeleteAssessmentExamPackageMap(ctx context.Context, arg db.DeleteAssessmentExamPackageMapParams) (int64, error)
	ListAssessmentPackageOptions(ctx context.Context, subjectID pgtype.UUID) ([]db.ListAssessmentPackageOptionsRow, error)
}

type AssessmentExam struct {
	q    assessmentExamStore
	pool *pgxpool.Pool
}

func NewAssessmentExam(q *db.Queries) *AssessmentExam {
	return &AssessmentExam{q: q}
}

func NewAssessmentExamWithPool(pool *pgxpool.Pool) *AssessmentExam {
	return &AssessmentExam{q: db.New(pool), pool: pool}
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

type AssessmentAssignmentClassSummary struct {
	ClassCode  string `json:"class_code"`
	ClassName  string `json:"class_name"`
	GradeLevel int32  `json:"grade_level"`
	Count      int64  `json:"count"`
}

type AssessmentAssignmentRoom struct {
	Code          string                             `json:"code"`
	Name          string                             `json:"name"`
	Capacity      int32                              `json:"capacity"`
	AssignedCount int64                              `json:"assigned_count"`
	GradeLevels   []int32                            `json:"grade_levels,omitempty"`
	ClassSummary  []AssessmentAssignmentClassSummary `json:"class_summary,omitempty"`
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

type AssessmentParticipantPlacementView struct {
	ParticipantID string `json:"participant_id"`
	SessionID     string `json:"session_id"`
	RoomID        string `json:"room_id,omitempty"`
	StudentID     string `json:"student_id"`
	StudentName   string `json:"student_name"`
	NIS           string `json:"nis,omitempty"`
	NISN          string `json:"nisn,omitempty"`
	ClassCode     string `json:"class_code"`
	ClassName     string `json:"class_name"`
	GradeLevel    int32  `json:"grade_level"`
	RoomCode      string `json:"room_code,omitempty"`
	RoomName      string `json:"room_name,omitempty"`
	RoomCapacity  int32  `json:"room_capacity,omitempty"`
	SeatNo        int32  `json:"seat_no,omitempty"`
	Status        string `json:"status"`
}

type AssessmentParticipantSeatInput struct {
	ParticipantID string `json:"participant_id"`
	RoomID        string `json:"room_id"`
	SeatNo        int32  `json:"seat_no"`
}

type AssessmentCardIssueInput struct {
	Regenerate   bool  `json:"regenerate"`
	ExpiresHours int32 `json:"expires_hours,omitempty"`
}

type AssessmentParticipantCardView struct {
	CardID         string `json:"card_id,omitempty"`
	ParticipantID  string `json:"participant_id"`
	SessionID      string `json:"session_id"`
	RoomID         string `json:"room_id,omitempty"`
	StudentID      string `json:"student_id"`
	StudentName    string `json:"student_name"`
	NIS            string `json:"nis,omitempty"`
	NISN           string `json:"nisn,omitempty"`
	ClassCode      string `json:"class_code"`
	ClassName      string `json:"class_name"`
	GradeLevel     int32  `json:"grade_level"`
	RoomCode       string `json:"room_code,omitempty"`
	RoomName       string `json:"room_name,omitempty"`
	SeatNo         int32  `json:"seat_no,omitempty"`
	Status         string `json:"status"`
	FailedAttempts int32  `json:"failed_attempts"`
	CardCreatedAt  string `json:"card_created_at,omitempty"`
	CardExpiresAt  string `json:"card_expires_at,omitempty"`
	Token          string `json:"token,omitempty"`
	PIN            string `json:"pin,omitempty"`
	QRPath         string `json:"qr_path,omitempty"`
}

type AssessmentParticipantCardIssueResult struct {
	ExamID  string                          `json:"exam_id"`
	Count   int                             `json:"count"`
	Cards   []AssessmentParticipantCardView `json:"cards"`
	Message string                          `json:"message"`
}

type AssessmentPackageMapInput struct {
	ID        string `json:"id,omitempty"`
	ClassID   string `json:"class_id"`
	SubjectID string `json:"subject_id"`
	PackageID string `json:"package_id"`
	SlotLabel string `json:"slot_label,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

type AssessmentPackageMapRequest struct {
	Items []AssessmentPackageMapInput `json:"items"`
}

type AssessmentPackageMapView struct {
	ID              string `json:"id"`
	ExamID          string `json:"exam_id"`
	ClassID         string `json:"class_id"`
	ClassCode       string `json:"class_code"`
	ClassName       string `json:"class_name"`
	GradeLevel      int32  `json:"grade_level"`
	SubjectID       string `json:"subject_id"`
	SubjectCode     string `json:"subject_code"`
	SubjectName     string `json:"subject_name"`
	PackageID       string `json:"package_id"`
	PackageTitle    string `json:"package_title"`
	DurationMinutes int32  `json:"duration_minutes"`
	SlotLabel       string `json:"slot_label,omitempty"`
	Notes           string `json:"notes,omitempty"`
}

type AssessmentPackageOptionView struct {
	ID              string `json:"id"`
	EventID         string `json:"event_id,omitempty"`
	SubjectID       string `json:"subject_id"`
	SubjectCode     string `json:"subject_code"`
	SubjectName     string `json:"subject_name"`
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	DurationMinutes int32  `json:"duration_minutes"`
	QuestionCount   int32  `json:"question_count"`
	SessionCount    int32  `json:"session_count"`
	Locked          bool   `json:"locked"`
	SnapshotVersion int32  `json:"snapshot_version"`
}

type AssessmentPackageMapSaveResult struct {
	ExamID  string `json:"exam_id"`
	Count   int    `json:"count"`
	Message string `json:"message"`
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
	participantCount, err := s.assignmentPreviewParticipantCount(ctx, id, normalized)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	result := buildAssessmentAssignmentResult(exam.ID, "", normalized, participantCount, false)
	if len(normalized.ClassIDs) > 0 && participantCount > 0 {
		result.Message = "Preview peserta dan ruang siap. Simpan untuk mendaftarkan siswa dari rombel terpilih dan menempatkan mereka ke ruang."
	}
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
	if len(normalized.ClassIDs) > 0 {
		if err := s.upsertAssessmentParticipantsFromClasses(ctx, session.ID, normalized); err != nil {
			return AssessmentAssignmentResult{}, err
		}
	}
	participants, err := s.q.ListAssessmentParticipantsForAssignment(ctx, session.ID)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	participants = orderAssessmentParticipantsForMixPolicy(participants, normalized.MixPolicy)
	result := buildAssessmentAssignmentResult(exam.ID, assessmentUUIDString(session.ID), normalized, int64(len(participants)), true)
	result.Rooms = attachAssessmentRoomComposition(result.Rooms, participants)
	roomsByCode := make(map[string]db.AssessmentRoom, len(result.Rooms))
	for _, room := range result.Rooms {
		savedRoom, err := s.q.UpsertAssessmentRoom(ctx, db.UpsertAssessmentRoomParams{
			SessionID: session.ID,
			Code:      room.Code,
			Name:      room.Name,
			Capacity:  room.Capacity,
		})
		if err != nil {
			return AssessmentAssignmentResult{}, err
		}
		roomsByCode[room.Code] = savedRoom
	}
	if err := s.assignAssessmentParticipantsToRooms(ctx, session.ID, participants, result.Rooms, roomsByCode); err != nil {
		return AssessmentAssignmentResult{}, err
	}
	cardCount, err := s.q.CountAssessmentCardsByExam(ctx, id)
	if err != nil {
		return AssessmentAssignmentResult{}, err
	}
	result.CardCount = cardCount
	if len(participants) > 0 {
		result.Message = "Ruang dan peserta tersimpan. Siswa sudah masuk ruang secara berurutan; kartu/QR+PIN belum diterbitkan."
	} else {
		result.Message = "Ruang ujian tersimpan. Belum ada peserta dari rombel terpilih; kartu/QR+PIN belum diterbitkan."
	}
	return result, nil
}

func (s *AssessmentExam) ListPackageMaps(ctx context.Context, id pgtype.UUID) ([]AssessmentPackageMapView, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.q.ListAssessmentExamPackageMaps(ctx, id)
	if err != nil {
		return nil, err
	}
	items := make([]AssessmentPackageMapView, 0, len(rows))
	for _, row := range rows {
		items = append(items, assessmentPackageMapView(row))
	}
	return items, nil
}

func (s *AssessmentExam) SavePackageMaps(ctx context.Context, id pgtype.UUID, input AssessmentPackageMapRequest) (AssessmentPackageMapSaveResult, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return AssessmentPackageMapSaveResult{}, err
	}
	if s.pool != nil {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return AssessmentPackageMapSaveResult{}, err
		}
		defer tx.Rollback(ctx) //nolint:errcheck
		result, err := s.savePackageMapsInStore(ctx, db.New(tx), id, input)
		if err != nil {
			return AssessmentPackageMapSaveResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return AssessmentPackageMapSaveResult{}, err
		}
		return result, nil
	}
	return s.savePackageMapsInStore(ctx, s.q, id, input)
}

func (s *AssessmentExam) savePackageMapsInStore(ctx context.Context, store assessmentExamStore, id pgtype.UUID, input AssessmentPackageMapRequest) (AssessmentPackageMapSaveResult, error) {
	if len(input.Items) == 0 {
		return AssessmentPackageMapSaveResult{}, fmt.Errorf("%w: minimal satu paket per rombel harus dipilih", domain.ErrBadRequest)
	}
	seen := map[string]bool{}
	for _, item := range input.Items {
		classID, err := assessmentUUIDFromString(item.ClassID, "rombel")
		if err != nil {
			return AssessmentPackageMapSaveResult{}, err
		}
		subjectID, err := assessmentUUIDFromString(item.SubjectID, "mata pelajaran")
		if err != nil {
			return AssessmentPackageMapSaveResult{}, err
		}
		packageID, err := assessmentUUIDFromString(item.PackageID, "paket soal")
		if err != nil {
			return AssessmentPackageMapSaveResult{}, err
		}
		key := assessmentUUIDString(classID) + ":" + assessmentUUIDString(subjectID)
		if seen[key] {
			return AssessmentPackageMapSaveResult{}, fmt.Errorf("%w: satu rombel hanya boleh punya satu paket untuk mapel yang sama", domain.ErrBadRequest)
		}
		seen[key] = true
		_, err = store.UpsertAssessmentExamPackageMap(ctx, db.UpsertAssessmentExamPackageMapParams{
			ExamID:    id,
			ClassID:   classID,
			SubjectID: subjectID,
			PackageID: packageID,
			SlotLabel: strings.TrimSpace(item.SlotLabel),
			Notes:     strings.TrimSpace(item.Notes),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return AssessmentPackageMapSaveResult{}, fmt.Errorf("%w: paket tidak aktif atau mapel paket tidak sesuai", domain.ErrBadRequest)
		}
		if err != nil {
			return AssessmentPackageMapSaveResult{}, err
		}
	}
	return AssessmentPackageMapSaveResult{ExamID: assessmentUUIDString(id), Count: len(input.Items), Message: "Paket per rombel tersimpan. Paket yang sama boleh dipakai bersama oleh beberapa rombel."}, nil
}

func (s *AssessmentExam) DeletePackageMap(ctx context.Context, examID, mapID pgtype.UUID) error {
	if _, err := s.Get(ctx, examID); err != nil {
		return err
	}
	if !mapID.Valid {
		return fmt.Errorf("%w: id pemetaan paket tidak valid", domain.ErrBadRequest)
	}
	rows, err := s.q.DeleteAssessmentExamPackageMap(ctx, db.DeleteAssessmentExamPackageMapParams{ExamID: examID, ID: mapID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *AssessmentExam) ListPackageOptions(ctx context.Context, subjectID pgtype.UUID) ([]AssessmentPackageOptionView, error) {
	rows, err := s.q.ListAssessmentPackageOptions(ctx, subjectID)
	if err != nil {
		return nil, err
	}
	items := make([]AssessmentPackageOptionView, 0, len(rows))
	for _, row := range rows {
		items = append(items, assessmentPackageOptionView(row))
	}
	return items, nil
}

func (s *AssessmentExam) ListParticipantCards(ctx context.Context, id pgtype.UUID) ([]AssessmentParticipantCardView, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.q.ListAssessmentParticipantCardTargetsByExam(ctx, id)
	if err != nil {
		return nil, err
	}
	cards := make([]AssessmentParticipantCardView, 0, len(rows))
	for _, row := range rows {
		cards = append(cards, assessmentParticipantCardTargetView(row, "", ""))
	}
	return cards, nil
}

func (s *AssessmentExam) issueParticipantCardsInStore(ctx context.Context, store assessmentExamStore, id pgtype.UUID, input AssessmentCardIssueInput, refreshAfterLock bool) ([]AssessmentParticipantCardView, int, error) {
	rows, err := store.ListAssessmentParticipantCardTargetsByExam(ctx, id)
	if err != nil {
		return nil, 0, err
	}
	if refreshAfterLock {
		// The first SELECT ... FOR UPDATE serializes issuance on participant rows.
		// Re-read in a fresh statement after locks are acquired so a concurrent
		// non-regenerate request sees cards inserted by the prior transaction and
		// does not overwrite/return a second raw token+PIN set.
		rows, err = store.ListAssessmentParticipantCardTargetsByExam(ctx, id)
		if err != nil {
			return nil, 0, err
		}
	}
	cards := make([]AssessmentParticipantCardView, 0, len(rows))
	if len(rows) == 0 {
		return cards, 0, nil
	}
	expiresAt := cardExpiresAt(input.ExpiresHours)
	issued := 0
	for _, row := range rows {
		if row.CardID.Valid && !input.Regenerate {
			cards = append(cards, assessmentParticipantCardTargetView(row, "", ""))
			continue
		}
		token, pin, err := newCardSecrets()
		if err != nil {
			return nil, 0, err
		}
		card, err := store.CreateAssessmentParticipantAccessCard(ctx, db.CreateAssessmentParticipantAccessCardParams{
			SessionID:     row.SessionID,
			ParticipantID: row.ParticipantID,
			TokenHash:     hashCardSecret(token),
			PinHash:       hashCardSecret(pin),
			ExpiresAt:     expiresAt,
		})
		if err != nil {
			return nil, 0, err
		}
		row.CardID = card.ID
		row.CardStatus = card.Status
		row.FailedAttempts = card.FailedAttempts
		row.CardCreatedAt = card.CreatedAt
		row.CardExpiresAt = card.ExpiresAt
		item := assessmentParticipantCardTargetView(row, token, pin)
		item.QRPath = "/ujian?card=" + token
		cards = append(cards, item)
		issued++
	}
	return cards, issued, nil
}

func (s *AssessmentExam) IssueParticipantCards(ctx context.Context, id pgtype.UUID, input AssessmentCardIssueInput) (AssessmentParticipantCardIssueResult, error) {
	exam, err := s.Get(ctx, id)
	if err != nil {
		return AssessmentParticipantCardIssueResult{}, err
	}
	var cards []AssessmentParticipantCardView
	var issued int
	if s.pool != nil {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return AssessmentParticipantCardIssueResult{}, err
		}
		defer tx.Rollback(ctx) //nolint:errcheck
		cards, issued, err = s.issueParticipantCardsInStore(ctx, db.New(tx), id, input, true)
		if err != nil {
			return AssessmentParticipantCardIssueResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return AssessmentParticipantCardIssueResult{}, err
		}
	} else {
		cards, issued, err = s.issueParticipantCardsInStore(ctx, s.q, id, input, false)
		if err != nil {
			return AssessmentParticipantCardIssueResult{}, err
		}
	}
	if len(cards) == 0 {
		return AssessmentParticipantCardIssueResult{ExamID: exam.ID, Cards: []AssessmentParticipantCardView{}, Message: "Belum ada peserta untuk diterbitkan kartu."}, nil
	}
	message := fmt.Sprintf("%d kartu peserta siap. PIN hanya tampil pada hasil terbitkan ini; cetak/simpan PDF sekarang.", len(cards))
	if issued == 0 {
		message = "Kartu peserta sudah pernah diterbitkan. Untuk melihat PIN mentah lagi, gunakan regenerasi dengan sengaja."
	}
	return AssessmentParticipantCardIssueResult{ExamID: exam.ID, Count: len(cards), Cards: cards, Message: message}, nil
}

func (s *AssessmentExam) ListParticipantPlacements(ctx context.Context, id pgtype.UUID) ([]AssessmentParticipantPlacementView, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.q.ListAssessmentParticipantPlacementsByExam(ctx, id)
	if err != nil {
		return nil, err
	}
	items := make([]AssessmentParticipantPlacementView, 0, len(rows))
	for _, row := range rows {
		items = append(items, assessmentParticipantPlacementListRowView(row))
	}
	return items, nil
}

func (s *AssessmentExam) MoveParticipantSeat(ctx context.Context, examID pgtype.UUID, input AssessmentParticipantSeatInput) (AssessmentParticipantPlacementView, error) {
	if _, err := s.Get(ctx, examID); err != nil {
		return AssessmentParticipantPlacementView{}, err
	}
	participantID, err := assessmentUUIDFromString(input.ParticipantID, "peserta")
	if err != nil {
		return AssessmentParticipantPlacementView{}, err
	}
	roomID, err := assessmentUUIDFromString(input.RoomID, "ruang")
	if err != nil {
		return AssessmentParticipantPlacementView{}, err
	}
	room, err := s.q.GetAssessmentRoomByIDAndExam(ctx, db.GetAssessmentRoomByIDAndExamParams{ExamID: examID, RoomID: roomID})
	if errors.Is(err, pgx.ErrNoRows) {
		return AssessmentParticipantPlacementView{}, fmt.Errorf("%w: ruang ujian tidak ditemukan pada asesmen ini", domain.ErrBadRequest)
	}
	if err != nil {
		return AssessmentParticipantPlacementView{}, err
	}
	if input.SeatNo < 1 || input.SeatNo > room.Capacity {
		return AssessmentParticipantPlacementView{}, fmt.Errorf("%w: nomor kursi harus 1 sampai %d", domain.ErrBadRequest, room.Capacity)
	}
	row, err := s.q.MoveAssessmentParticipantSeat(ctx, db.MoveAssessmentParticipantSeatParams{
		ExamID:        examID,
		ParticipantID: participantID,
		RoomID:        roomID,
		SeatNo:        pgtype.Int4{Int32: input.SeatNo, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return AssessmentParticipantPlacementView{}, fmt.Errorf("%w: peserta tidak ditemukan pada asesmen ini", domain.ErrBadRequest)
	}
	if err != nil {
		return AssessmentParticipantPlacementView{}, err
	}
	return assessmentParticipantPlacementMoveRowView(row), nil
}

func (s *AssessmentExam) assignmentPreviewParticipantCount(ctx context.Context, examID pgtype.UUID, input AssessmentAssignmentRequest) (int64, error) {
	if len(input.ClassIDs) == 0 {
		return s.q.CountAssessmentParticipantsByExam(ctx, examID)
	}
	classIDs, err := assessmentUUIDsFromStrings(input.ClassIDs)
	if err != nil {
		return 0, err
	}
	students, err := s.q.ListAssessmentCandidateStudentsByClassIDs(ctx, classIDs)
	if err != nil {
		return 0, err
	}
	return int64(len(students)), nil
}

func (s *AssessmentExam) upsertAssessmentParticipantsFromClasses(ctx context.Context, sessionID pgtype.UUID, input AssessmentAssignmentRequest) error {
	classIDs, err := assessmentUUIDsFromStrings(input.ClassIDs)
	if err != nil {
		return err
	}
	students, err := s.q.ListAssessmentCandidateStudentsByClassIDs(ctx, classIDs)
	if err != nil {
		return err
	}
	for _, student := range students {
		if _, err := s.q.UpsertAssessmentParticipant(ctx, db.UpsertAssessmentParticipantParams{SessionID: sessionID, StudentID: student.StudentID}); err != nil {
			return err
		}
	}
	return nil
}

func (s *AssessmentExam) assignAssessmentParticipantsToRooms(ctx context.Context, sessionID pgtype.UUID, participants []db.ListAssessmentParticipantsForAssignmentRow, rooms []AssessmentAssignmentRoom, roomsByCode map[string]db.AssessmentRoom) error {
	if len(participants) == 0 {
		return nil
	}
	if err := s.q.ClearAssessmentParticipantRooms(ctx, sessionID); err != nil {
		return err
	}
	participantIndex := 0
	for _, plannedRoom := range rooms {
		savedRoom, ok := roomsByCode[plannedRoom.Code]
		if !ok {
			return fmt.Errorf("%w: ruang %s belum tersimpan", domain.ErrBadRequest, plannedRoom.Code)
		}
		for seat := int32(1); seat <= plannedRoom.Capacity && participantIndex < len(participants); seat++ {
			participant := participants[participantIndex]
			if err := s.q.AssignAssessmentParticipantRoom(ctx, db.AssignAssessmentParticipantRoomParams{
				RoomID:        savedRoom.ID,
				SeatNo:        pgtype.Int4{Int32: seat, Valid: true},
				ParticipantID: participant.ParticipantID,
				SessionID:     sessionID,
			}); err != nil {
				return err
			}
			participantIndex++
		}
	}
	return nil
}

func orderAssessmentParticipantsForMixPolicy(participants []db.ListAssessmentParticipantsForAssignmentRow, mixPolicy string) []db.ListAssessmentParticipantsForAssignmentRow {
	if mixPolicy != AssessmentMixPolicyMixed || len(participants) < 2 {
		return participants
	}
	groups := make(map[string][]db.ListAssessmentParticipantsForAssignmentRow)
	order := make([]string, 0)
	for _, participant := range participants {
		key := assessmentParticipantClassKey(participant)
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], participant)
	}
	if len(order) < 2 {
		return participants
	}
	mixed := make([]db.ListAssessmentParticipantsForAssignmentRow, 0, len(participants))
	for {
		added := false
		for _, key := range order {
			bucket := groups[key]
			if len(bucket) == 0 {
				continue
			}
			mixed = append(mixed, bucket[0])
			groups[key] = bucket[1:]
			added = true
		}
		if !added {
			break
		}
	}
	return mixed
}

func assessmentParticipantClassKey(participant db.ListAssessmentParticipantsForAssignmentRow) string {
	classCode := strings.TrimSpace(participant.ClassCode)
	className := strings.TrimSpace(participant.ClassName)
	if classCode == "" {
		classCode = "Tanpa Rombel"
	}
	return fmt.Sprintf("%d|%s|%s", participant.GradeLevel, classCode, className)
}

func attachAssessmentRoomComposition(rooms []AssessmentAssignmentRoom, participants []db.ListAssessmentParticipantsForAssignmentRow) []AssessmentAssignmentRoom {
	if len(rooms) == 0 || len(participants) == 0 {
		return rooms
	}
	participantIndex := 0
	for roomIndex := range rooms {
		classCounts := map[string]*AssessmentAssignmentClassSummary{}
		classOrder := make([]string, 0)
		gradeSeen := map[int32]bool{}
		grades := make([]int32, 0)
		for seat := int32(1); seat <= rooms[roomIndex].Capacity && participantIndex < len(participants); seat++ {
			participant := participants[participantIndex]
			classCode := strings.TrimSpace(participant.ClassCode)
			className := strings.TrimSpace(participant.ClassName)
			if classCode == "" {
				classCode = "Tanpa Rombel"
			}
			key := fmt.Sprintf("%d|%s|%s", participant.GradeLevel, classCode, className)
			if _, ok := classCounts[key]; !ok {
				classCounts[key] = &AssessmentAssignmentClassSummary{ClassCode: classCode, ClassName: className, GradeLevel: participant.GradeLevel}
				classOrder = append(classOrder, key)
			}
			classCounts[key].Count++
			if participant.GradeLevel > 0 && !gradeSeen[participant.GradeLevel] {
				gradeSeen[participant.GradeLevel] = true
				grades = append(grades, participant.GradeLevel)
			}
			participantIndex++
		}
		sort.Slice(grades, func(i, j int) bool { return grades[i] < grades[j] })
		rooms[roomIndex].GradeLevels = grades
		rooms[roomIndex].ClassSummary = make([]AssessmentAssignmentClassSummary, 0, len(classOrder))
		for _, key := range classOrder {
			rooms[roomIndex].ClassSummary = append(rooms[roomIndex].ClassSummary, *classCounts[key])
		}
	}
	return rooms
}

func assessmentUUIDFromString(raw, label string) (pgtype.UUID, error) {
	var id pgtype.UUID
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return id, fmt.Errorf("%w: id %s wajib diisi", domain.ErrBadRequest, label)
	}
	if err := id.Scan(trimmed); err != nil || !id.Valid {
		return id, fmt.Errorf("%w: id %s tidak valid", domain.ErrBadRequest, label)
	}
	return id, nil
}

func assessmentUUIDsFromStrings(rawIDs []string) ([]pgtype.UUID, error) {
	ids := make([]pgtype.UUID, 0, len(rawIDs))
	seen := make(map[string]bool, len(rawIDs))
	for _, raw := range rawIDs {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		var id pgtype.UUID
		if err := id.Scan(trimmed); err != nil || !id.Valid {
			return nil, fmt.Errorf("%w: rombel peserta tidak valid", domain.ErrBadRequest)
		}
		seen[trimmed] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("%w: pilih minimal satu rombel peserta", domain.ErrBadRequest)
	}
	return ids, nil
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

func assessmentParticipantCardTargetView(row db.ListAssessmentParticipantCardTargetsByExamRow, token, pin string) AssessmentParticipantCardView {
	return AssessmentParticipantCardView{
		CardID:         assessmentUUIDString(row.CardID),
		ParticipantID:  assessmentUUIDString(row.ParticipantID),
		SessionID:      assessmentUUIDString(row.SessionID),
		RoomID:         assessmentUUIDString(row.RoomID),
		StudentID:      assessmentUUIDString(row.StudentID),
		StudentName:    row.StudentName,
		NIS:            row.Nis,
		NISN:           row.Nisn,
		ClassCode:      row.ClassCode,
		ClassName:      row.ClassName,
		GradeLevel:     row.GradeLevel,
		RoomCode:       row.RoomCode,
		RoomName:       row.RoomName,
		SeatNo:         row.SeatNo,
		Status:         firstNonEmpty(row.CardStatus, "not_issued"),
		FailedAttempts: row.FailedAttempts,
		CardCreatedAt:  timeString(row.CardCreatedAt),
		CardExpiresAt:  timeString(row.CardExpiresAt),
		Token:          token,
		PIN:            pin,
	}
}

func assessmentParticipantPlacementListRowView(row db.ListAssessmentParticipantPlacementsByExamRow) AssessmentParticipantPlacementView {
	roomID := row.RoomID
	if row.RoomIDActual.Valid {
		roomID = row.RoomIDActual
	}
	return AssessmentParticipantPlacementView{
		ParticipantID: assessmentUUIDString(row.ParticipantID),
		SessionID:     assessmentUUIDString(row.SessionID),
		RoomID:        assessmentUUIDString(roomID),
		StudentID:     assessmentUUIDString(row.StudentID),
		StudentName:   row.StudentName,
		NIS:           row.Nis,
		NISN:          row.Nisn,
		ClassCode:     row.ClassCode,
		ClassName:     row.ClassName,
		GradeLevel:    row.GradeLevel,
		RoomCode:      row.RoomCode,
		RoomName:      row.RoomName,
		RoomCapacity:  row.RoomCapacity,
		SeatNo:        row.SeatNo,
		Status:        row.Status,
	}
}

func assessmentParticipantPlacementMoveRowView(row db.MoveAssessmentParticipantSeatRow) AssessmentParticipantPlacementView {
	roomID := row.RoomID
	if row.RoomIDActual.Valid {
		roomID = row.RoomIDActual
	}
	return AssessmentParticipantPlacementView{
		ParticipantID: assessmentUUIDString(row.ParticipantID),
		SessionID:     assessmentUUIDString(row.SessionID),
		RoomID:        assessmentUUIDString(roomID),
		StudentID:     assessmentUUIDString(row.StudentID),
		StudentName:   row.StudentName,
		NIS:           row.Nis,
		NISN:          row.Nisn,
		ClassCode:     row.ClassCode,
		ClassName:     row.ClassName,
		GradeLevel:    row.GradeLevel,
		RoomCode:      row.RoomCode,
		RoomName:      row.RoomName,
		RoomCapacity:  row.RoomCapacity,
		SeatNo:        row.SeatNo,
		Status:        row.Status,
	}
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
		CardCount:        row.CardCount,
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

func assessmentPackageMapView(row db.ListAssessmentExamPackageMapsRow) AssessmentPackageMapView {
	return AssessmentPackageMapView{
		ID: assessmentUUIDString(row.ID), ExamID: assessmentUUIDString(row.ExamID), ClassID: assessmentUUIDString(row.ClassID),
		ClassCode: row.ClassCode, ClassName: row.ClassName, GradeLevel: row.GradeLevel,
		SubjectID: assessmentUUIDString(row.SubjectID), SubjectCode: row.SubjectCode, SubjectName: row.SubjectName,
		PackageID: assessmentUUIDString(row.PackageID), PackageTitle: row.PackageTitle, DurationMinutes: row.DurationMinutes,
		SlotLabel: row.SlotLabel, Notes: row.Notes,
	}
}

func assessmentPackageOptionView(row db.ListAssessmentPackageOptionsRow) AssessmentPackageOptionView {
	return AssessmentPackageOptionView{
		ID: assessmentUUIDString(row.ID), EventID: assessmentUUIDString(row.EventID), SubjectID: assessmentUUIDString(row.SubjectID),
		SubjectCode: row.SubjectCode, SubjectName: row.SubjectName, Title: row.Title, Description: row.Description,
		DurationMinutes: row.DurationMinutes, QuestionCount: row.QuestionCount, SessionCount: row.SessionCount,
		Locked: row.LockedAt.Valid, SnapshotVersion: row.SnapshotVersion,
	}
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
