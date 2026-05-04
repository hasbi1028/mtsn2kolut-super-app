package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtEvent struct {
	q    cbtEventStore
	pool *pgxpool.Pool
}

type cbtEventStore interface {
	ListCbtExamEvents(ctx context.Context) ([]db.ListCbtExamEventsRow, error)
	GetCbtExamEvent(ctx context.Context, id pgtype.UUID) (db.GetCbtExamEventRow, error)
	GetCbtEventOverviewSummary(ctx context.Context, id pgtype.UUID) (db.GetCbtEventOverviewSummaryRow, error)
	GetEventResults(ctx context.Context, eventID pgtype.UUID) ([]db.GetEventResultsRow, error)
	GetEventExamCards(ctx context.Context, eventID pgtype.UUID) ([]db.GetEventExamCardsRow, error)
	ListCbtEventPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventPackagesRow, error)
	ListCbtEventSessionsReadiness(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSessionsReadinessRow, error)
	ListCbtEventSubjectMatrix(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectMatrixRow, error)
	CreateCbtExamEvent(ctx context.Context, arg db.CreateCbtExamEventParams) (db.CbtExamEvent, error)
	UpdateCbtExamEventStatus(ctx context.Context, arg db.UpdateCbtExamEventStatusParams) (db.CbtExamEvent, error)
	UpdateCbtExamEvent(ctx context.Context, arg db.UpdateCbtExamEventParams) (db.CbtExamEvent, error)
	DeleteCbtExamEvent(ctx context.Context, id pgtype.UUID) (int64, error)
	ListCbtEventMembers(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventMembersRow, error)
	ListCbtEventMembersByUser(ctx context.Context, userID pgtype.UUID) ([]db.CbtEventMember, error)
	GetCbtEventMember(ctx context.Context, arg db.GetCbtEventMemberParams) (db.GetCbtEventMemberRow, error)
	CreateCbtEventMember(ctx context.Context, arg db.CreateCbtEventMemberParams) (db.CbtEventMember, error)
	UpdateCbtEventMember(ctx context.Context, arg db.UpdateCbtEventMemberParams) (db.CbtEventMember, error)
	DeleteCbtEventMember(ctx context.Context, arg db.DeleteCbtEventMemberParams) error
	ListCbtEventSubjectTargets(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectTargetsRow, error)
	UpsertCbtEventSubjectTarget(ctx context.Context, arg db.UpsertCbtEventSubjectTargetParams) (db.CbtEventSubjectTarget, error)
	DeleteCbtEventSubjectTarget(ctx context.Context, arg db.DeleteCbtEventSubjectTargetParams) (int64, error)
}

func NewCbtEvent(pool *pgxpool.Pool) *CbtEvent {
	return &CbtEvent{q: db.New(pool), pool: pool}
}

func (s *CbtEvent) List(ctx context.Context) ([]db.ListCbtExamEventsRow, error) {
	rows, err := s.q.ListCbtExamEvents(ctx)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamEventsRow{}, nil
	}
	return rows, nil
}

func (s *CbtEvent) ListForUser(ctx context.Context, userID pgtype.UUID) ([]db.ListCbtExamEventsRow, error) {
	if !userID.Valid {
		return []db.ListCbtExamEventsRow{}, nil
	}
	rows, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	members, err := s.q.ListCbtEventMembersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	allowed := make(map[pgtype.UUID]struct{}, len(members))
	for _, member := range members {
		allowed[member.EventID] = struct{}{}
	}
	filtered := make([]db.ListCbtExamEventsRow, 0, len(rows))
	for _, row := range rows {
		if _, ok := allowed[row.ID]; ok {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

func (s *CbtEvent) Get(ctx context.Context, id pgtype.UUID) (db.GetCbtExamEventRow, error) {
	return s.q.GetCbtExamEvent(ctx, id)
}

type CbtEventOverview struct {
	Event           db.GetCbtEventOverviewSummaryRow      `json:"event"`
	Members         []db.ListCbtEventMembersRow           `json:"members"`
	QuestionTargets []db.ListCbtEventSubjectTargetsRow    `json:"question_targets"`
	SubjectMatrix   []db.ListCbtEventSubjectMatrixRow     `json:"subject_matrix"`
	Packages        []db.ListCbtEventPackagesRow          `json:"packages"`
	Sessions        []db.ListCbtEventSessionsReadinessRow `json:"sessions"`
	Readiness       CbtEventReadiness                     `json:"readiness"`
	BlockingReasons []string                              `json:"blocking_reasons"`
}

type CbtEventReadiness struct {
	AuthoringReady bool `json:"authoring_ready"`
	PackageReady   bool `json:"package_ready"`
	SessionReady   bool `json:"session_ready"`
	RoomReady      bool `json:"room_ready"`
	TokenReady     bool `json:"token_ready"`
	CardReady      bool `json:"card_ready"`
	RuntimeStarted bool `json:"runtime_started"`
	ResultsReady   bool `json:"results_ready"`
}

func (s *CbtEvent) Overview(ctx context.Context, id pgtype.UUID) (CbtEventOverview, error) {
	summary, err := s.q.GetCbtEventOverviewSummary(ctx, id)
	if err != nil {
		return CbtEventOverview{}, err
	}
	members, err := s.ListMembers(ctx, id)
	if err != nil {
		return CbtEventOverview{}, err
	}
	targets, err := s.ListQuestionTargets(ctx, id)
	if err != nil {
		return CbtEventOverview{}, err
	}
	matrix, err := s.ListSubjectMatrix(ctx, id)
	if err != nil {
		return CbtEventOverview{}, err
	}
	packages, err := s.ListPackages(ctx, id)
	if err != nil {
		return CbtEventOverview{}, err
	}
	sessions, err := s.ListSessions(ctx, id)
	if err != nil {
		return CbtEventOverview{}, err
	}
	readiness, blockers := buildCbtEventReadiness(summary, matrix)
	return CbtEventOverview{
		Event:           summary,
		Members:         members,
		QuestionTargets: targets,
		SubjectMatrix:   matrix,
		Packages:        packages,
		Sessions:        sessions,
		Readiness:       readiness,
		BlockingReasons: blockers,
	}, nil
}

func (s *CbtEvent) ListPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventPackagesRow, error) {
	rows, err := s.q.ListCbtEventPackages(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtEventPackagesRow{}, nil
	}
	return rows, nil
}

func (s *CbtEvent) ListSessions(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSessionsReadinessRow, error) {
	rows, err := s.q.ListCbtEventSessionsReadiness(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtEventSessionsReadinessRow{}, nil
	}
	return rows, nil
}

func (s *CbtEvent) ListSubjectMatrix(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectMatrixRow, error) {
	rows, err := s.q.ListCbtEventSubjectMatrix(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtEventSubjectMatrixRow{}, nil
	}
	return rows, nil
}

func (s *CbtEvent) GetResults(ctx context.Context, id pgtype.UUID) ([]db.GetEventResultsRow, error) {
	rows, err := s.q.GetEventResults(ctx, id)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetEventResultsRow{}, nil
	}
	return rows, nil
}

func (s *CbtEvent) GetExamCards(ctx context.Context, id pgtype.UUID) ([]db.GetEventExamCardsRow, error) {
	rows, err := s.q.GetEventExamCards(ctx, id)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetEventExamCardsRow{}, nil
	}
	return rows, nil
}

type CreateCbtEventInput struct {
	Title          string
	ExamType       db.CbtExamType
	Scope          string
	TargetLevels   []string
	AcademicYearID pgtype.UUID
	Status         string
}

func (s *CbtEvent) Create(ctx context.Context, in CreateCbtEventInput) (db.CbtExamEvent, error) {
	status := in.Status
	if status == "" {
		status = "draft"
	}
	if !validCbtEventStatus(status) {
		return db.CbtExamEvent{}, fmt.Errorf("%w: status event CBT tidak valid", domain.ErrBadRequest)
	}
	return s.q.CreateCbtExamEvent(ctx, db.CreateCbtExamEventParams{
		Title:          in.Title,
		ExamType:       in.ExamType,
		Scope:          in.Scope,
		TargetLevels:   in.TargetLevels,
		AcademicYearID: in.AcademicYearID,
		Status:         status,
	})
}

func (s *CbtEvent) UpdateStatus(ctx context.Context, id pgtype.UUID, status string) (db.CbtExamEvent, error) {
	if !validCbtEventStatus(status) {
		return db.CbtExamEvent{}, fmt.Errorf("%w: status event CBT tidak valid", domain.ErrBadRequest)
	}
	return s.q.UpdateCbtExamEventStatus(ctx, db.UpdateCbtExamEventStatusParams{
		ID:     id,
		Status: status,
	})
}

func (s *CbtEvent) CanRead(ctx context.Context, eventID, userID pgtype.UUID) (bool, error) {
	if !eventID.Valid || !userID.Valid {
		return false, nil
	}
	members, err := s.q.ListCbtEventMembersByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, member := range members {
		if sameUUID(member.EventID, eventID) {
			return true, nil
		}
	}
	return false, nil
}

func validCbtEventStatus(status string) bool {
	switch status {
	case "draft", "active", "finished", "cancelled":
		return true
	default:
		return false
	}
}

func buildCbtEventReadiness(summary db.GetCbtEventOverviewSummaryRow, matrix []db.ListCbtEventSubjectMatrixRow) (CbtEventReadiness, []string) {
	readiness := CbtEventReadiness{
		AuthoringReady: summary.TargetQuestionCount > 0 && summary.PublishedQuestions >= summary.TargetQuestionCount,
		PackageReady:   summary.ActivePackageCount > 0 && summary.EmptyPackageCount == 0,
		SessionReady:   summary.SessionCount > 0,
		RoomReady:      summary.SessionCount > 0 && summary.RoomCount > 0 && summary.RoomsWithoutProctor == 0 && summary.UnassignedParticipantCount == 0 && summary.MissingSeatCount == 0,
		TokenReady:     summary.ParticipantCount > 0 && summary.TokenReadyCount == summary.ParticipantCount,
		CardReady:      summary.ParticipantCount > 0 && summary.TokenReadyCount == summary.ParticipantCount && summary.MissingSeatCount == 0,
		RuntimeStarted: summary.JoinedCount > 0,
		ResultsReady:   summary.SubmittedCount > 0 && summary.ScoredCount >= summary.SubmittedCount,
	}
	blockers := make([]string, 0, 8)
	if summary.MemberCount == 0 {
		blockers = append(blockers, "event belum memiliki anggota/panitia")
	}
	if !readiness.AuthoringReady {
		blockers = append(blockers, "jumlah soal terbit belum memenuhi target event")
	}
	for _, row := range matrix {
		if row.ShortageCount > 0 {
			blockers = append(blockers, fmt.Sprintf("mapel %s kurang %d soal terbit", row.SubjectName, row.ShortageCount))
		}
		if row.TargetQuestions > 0 && row.AuthorCount == 0 {
			blockers = append(blockers, fmt.Sprintf("mapel %s belum memiliki pembuat soal", row.SubjectName))
		}
		if row.TargetQuestions > 0 && row.ReviewerCount == 0 {
			blockers = append(blockers, fmt.Sprintf("mapel %s belum memiliki reviewer", row.SubjectName))
		}
	}
	if !readiness.PackageReady {
		blockers = append(blockers, "belum ada paket aktif yang berisi soal")
	}
	if !readiness.SessionReady {
		blockers = append(blockers, "belum ada sesi ujian untuk event")
	}
	if summary.ParticipantCount == 0 {
		blockers = append(blockers, "belum ada peserta ujian")
	}
	if summary.UnassignedParticipantCount > 0 {
		blockers = append(blockers, fmt.Sprintf("%d peserta belum masuk ruang", summary.UnassignedParticipantCount))
	}
	if summary.MissingSeatCount > 0 {
		blockers = append(blockers, fmt.Sprintf("%d peserta belum memiliki nomor kursi", summary.MissingSeatCount))
	}
	if summary.RoomsWithoutProctor > 0 {
		blockers = append(blockers, fmt.Sprintf("%d ruang belum memiliki pengawas/proktor", summary.RoomsWithoutProctor))
	}
	if summary.ParticipantCount > 0 && summary.TokenReadyCount < summary.ParticipantCount {
		blockers = append(blockers, fmt.Sprintf("%d token peserta belum siap", summary.ParticipantCount-summary.TokenReadyCount))
	}
	return readiness, blockers
}

func (s *CbtEvent) Update(ctx context.Context, id pgtype.UUID, in CreateCbtEventInput) (db.CbtExamEvent, error) {
	return s.q.UpdateCbtExamEvent(ctx, db.UpdateCbtExamEventParams{
		ID:             id,
		Title:          in.Title,
		ExamType:       in.ExamType,
		Scope:          in.Scope,
		TargetLevels:   in.TargetLevels,
		AcademicYearID: in.AcademicYearID,
	})
}

func (s *CbtEvent) Delete(ctx context.Context, id pgtype.UUID) error {
	rows, err := s.q.DeleteCbtExamEvent(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("%w: event tidak ditemukan atau bukan draft", domain.ErrConflict)
	}
	return nil
}

type SaveCbtEventMemberInput struct {
	UserID     pgtype.UUID
	EmployeeID pgtype.UUID
	SubjectID  pgtype.UUID
	Role       db.CbtEventMemberRole
}

func (s *CbtEvent) ListMembers(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventMembersRow, error) {
	rows, err := s.q.ListCbtEventMembers(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtEventMembersRow{}, nil
	}
	return rows, nil
}

func (s *CbtEvent) CreateMember(ctx context.Context, eventID pgtype.UUID, in SaveCbtEventMemberInput) (db.CbtEventMember, error) {
	return s.q.CreateCbtEventMember(ctx, db.CreateCbtEventMemberParams{
		EventID:    eventID,
		UserID:     in.UserID,
		EmployeeID: in.EmployeeID,
		SubjectID:  in.SubjectID,
		Role:       in.Role,
	})
}

func (s *CbtEvent) UpdateMember(ctx context.Context, eventID pgtype.UUID, id pgtype.UUID, in SaveCbtEventMemberInput) (db.CbtEventMember, error) {
	return s.q.UpdateCbtEventMember(ctx, db.UpdateCbtEventMemberParams{
		ID:         id,
		EventID:    eventID,
		UserID:     in.UserID,
		EmployeeID: in.EmployeeID,
		SubjectID:  in.SubjectID,
		Role:       in.Role,
	})
}

func (s *CbtEvent) DeleteMember(ctx context.Context, eventID pgtype.UUID, id pgtype.UUID) error {
	return s.q.DeleteCbtEventMember(ctx, db.DeleteCbtEventMemberParams{ID: id, EventID: eventID})
}

func (s *CbtEvent) ListQuestionTargets(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventSubjectTargetsRow, error) {
	rows, err := s.q.ListCbtEventSubjectTargets(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtEventSubjectTargetsRow{}, nil
	}
	return rows, nil
}

type SaveCbtEventSubjectTargetInput struct {
	SubjectID       pgtype.UUID
	TargetQuestions int32
}

func (s *CbtEvent) UpsertQuestionTarget(ctx context.Context, eventID pgtype.UUID, in SaveCbtEventSubjectTargetInput) (db.CbtEventSubjectTarget, error) {
	if !eventID.Valid || !in.SubjectID.Valid {
		return db.CbtEventSubjectTarget{}, domain.ErrBadRequest
	}
	if in.TargetQuestions <= 0 {
		return db.CbtEventSubjectTarget{}, fmt.Errorf("%w: target_questions harus lebih dari 0", domain.ErrBadRequest)
	}
	return s.q.UpsertCbtEventSubjectTarget(ctx, db.UpsertCbtEventSubjectTargetParams{
		EventID:         eventID,
		SubjectID:       in.SubjectID,
		TargetQuestions: in.TargetQuestions,
	})
}

func (s *CbtEvent) DeleteQuestionTarget(ctx context.Context, eventID pgtype.UUID, subjectID pgtype.UUID) error {
	rows, err := s.q.DeleteCbtEventSubjectTarget(ctx, db.DeleteCbtEventSubjectTargetParams{EventID: eventID, SubjectID: subjectID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}
