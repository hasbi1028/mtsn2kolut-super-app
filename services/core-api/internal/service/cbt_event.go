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
	ListCbtEventQuestionCompletenessRows(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventQuestionCompletenessRowsRow, error)
	ListCbtEventQuestionPoolContributions(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtEventQuestionPoolContributionsRow, error)
	ListCbtEventQuestionCompletenessExcludedLevels(ctx context.Context, id pgtype.UUID) ([]string, error)
	GetCbtEventQuestionRequirements(ctx context.Context, id pgtype.UUID) (db.GetCbtEventQuestionRequirementsRow, error)
	UpsertCbtEventQuestionRequirements(ctx context.Context, arg db.UpsertCbtEventQuestionRequirementsParams) (db.UpsertCbtEventQuestionRequirementsRow, error)
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

type CbtSopStageStatus string

const (
	CbtSopStageReady   CbtSopStageStatus = "ready"
	CbtSopStageWarning CbtSopStageStatus = "warning"
	CbtSopStageBlocked CbtSopStageStatus = "blocked"
	CbtSopStageRunning CbtSopStageStatus = "running"
)

type CbtSopNextAction struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type CbtSopStageReadiness struct {
	Key           string             `json:"key"`
	Label         string             `json:"label"`
	Status        CbtSopStageStatus  `json:"status"`
	BlockingCount int32              `json:"blocking_count"`
	WarningCount  int32              `json:"warning_count"`
	NextActions   []CbtSopNextAction `json:"next_actions"`
}

type CbtSopReadiness struct {
	EventID string                 `json:"event_id"`
	Stages  []CbtSopStageReadiness `json:"stages"`
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

func (s *CbtEvent) SopReadiness(ctx context.Context, id pgtype.UUID) (CbtSopReadiness, error) {
	overview, err := s.Overview(ctx, id)
	if err != nil {
		return CbtSopReadiness{}, err
	}
	return buildCbtSopReadiness(overview), nil
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

type CbtQuestionCompleteness struct {
	Requirements   db.GetCbtEventQuestionRequirementsRow         `json:"requirements"`
	Summary        CbtQuestionCompletenessSummary                `json:"summary"`
	Rows           []CbtQuestionCompletenessRow                  `json:"rows"`
	Contributions  []db.ListCbtEventQuestionPoolContributionsRow `json:"contributions"`
	ExcludedLevels []string                                      `json:"excluded_levels"`
}

type CbtQuestionCompletenessSummary struct {
	TotalRows      int32 `json:"total_rows"`
	CompleteRows   int32 `json:"complete_rows"`
	IncompleteRows int32 `json:"incomplete_rows"`
	TargetPg       int32 `json:"target_pg"`
	AvailablePg    int32 `json:"available_pg"`
	MissingPg      int32 `json:"missing_pg"`
	TargetEssay    int32 `json:"target_essay"`
	AvailableEssay int32 `json:"available_essay"`
	MissingEssay   int32 `json:"missing_essay"`
}

type CbtQuestionCompletenessRow struct {
	Level             string      `json:"level"`
	ClassID           pgtype.UUID `json:"class_id"`
	ClassCode         string      `json:"class_code"`
	ClassName         string      `json:"class_name"`
	SubjectID         pgtype.UUID `json:"subject_id"`
	SubjectName       string      `json:"subject_name"`
	SubjectCode       string      `json:"subject_code"`
	TeacherEmployeeID pgtype.UUID `json:"teacher_employee_id"`
	TeacherName       string      `json:"teacher_name"`
	TeacherUsername   string      `json:"teacher_username"`
	ScopeMode         string      `json:"scope_mode"`
	StatusFilter      string      `json:"status_filter"`
	TargetPg          int32       `json:"target_pg"`
	AvailablePg       int32       `json:"available_pg"`
	MissingPg         int32       `json:"missing_pg"`
	TargetEssay       int32       `json:"target_essay"`
	AvailableEssay    int32       `json:"available_essay"`
	MissingEssay      int32       `json:"missing_essay"`
	Complete          bool        `json:"complete"`
}

func (s *CbtEvent) QuestionCompleteness(ctx context.Context, eventID pgtype.UUID) (CbtQuestionCompleteness, error) {
	requirements, err := s.q.GetCbtEventQuestionRequirements(ctx, eventID)
	if err != nil {
		return CbtQuestionCompleteness{}, err
	}
	rows, err := s.q.ListCbtEventQuestionCompletenessRows(ctx, eventID)
	if err != nil {
		return CbtQuestionCompleteness{}, err
	}
	excluded, err := s.q.ListCbtEventQuestionCompletenessExcludedLevels(ctx, eventID)
	if err != nil {
		return CbtQuestionCompleteness{}, err
	}
	contributions, err := s.q.ListCbtEventQuestionPoolContributions(ctx, eventID)
	if err != nil {
		return CbtQuestionCompleteness{}, err
	}
	if contributions == nil {
		contributions = []db.ListCbtEventQuestionPoolContributionsRow{}
	}
	out := CbtQuestionCompleteness{Requirements: requirements, Rows: []CbtQuestionCompletenessRow{}, Contributions: contributions, ExcludedLevels: excluded}
	for _, row := range rows {
		missingPg := maxInt32(row.TargetPg-row.AvailablePg, 0)
		missingEssay := maxInt32(row.TargetEssay-row.AvailableEssay, 0)
		complete := missingPg == 0 && missingEssay == 0
		out.Rows = append(out.Rows, CbtQuestionCompletenessRow{
			Level:             row.Level,
			ClassID:           row.ClassID,
			ClassCode:         row.ClassCode,
			ClassName:         row.ClassName,
			SubjectID:         row.SubjectID,
			SubjectName:       row.SubjectName,
			SubjectCode:       row.SubjectCode,
			TeacherEmployeeID: row.TeacherEmployeeID,
			TeacherName:       row.TeacherName,
			TeacherUsername:   row.TeacherUsername,
			ScopeMode:         row.ScopeMode,
			StatusFilter:      row.StatusFilter,
			TargetPg:          row.TargetPg,
			AvailablePg:       row.AvailablePg,
			MissingPg:         missingPg,
			TargetEssay:       row.TargetEssay,
			AvailableEssay:    row.AvailableEssay,
			MissingEssay:      missingEssay,
			Complete:          complete,
		})
		out.Summary.TotalRows++
		if complete {
			out.Summary.CompleteRows++
		} else {
			out.Summary.IncompleteRows++
		}
		out.Summary.TargetPg += row.TargetPg
		out.Summary.AvailablePg += row.AvailablePg
		out.Summary.MissingPg += missingPg
		out.Summary.TargetEssay += row.TargetEssay
		out.Summary.AvailableEssay += row.AvailableEssay
		out.Summary.MissingEssay += missingEssay
	}
	return out, nil
}

func maxInt32(v, floor int32) int32 {
	if v < floor {
		return floor
	}
	return v
}

type SaveCbtEventQuestionRequirementsInput struct {
	ScopeMode    string `json:"scope_mode"`
	TargetPg     int32  `json:"target_pg"`
	TargetEssay  int32  `json:"target_essay"`
	StatusFilter string `json:"status_filter"`
}

func (s *CbtEvent) GetQuestionRequirements(ctx context.Context, eventID pgtype.UUID) (db.GetCbtEventQuestionRequirementsRow, error) {
	return s.q.GetCbtEventQuestionRequirements(ctx, eventID)
}

func (s *CbtEvent) UpsertQuestionRequirements(ctx context.Context, eventID pgtype.UUID, in SaveCbtEventQuestionRequirementsInput) (db.UpsertCbtEventQuestionRequirementsRow, error) {
	scopeMode := in.ScopeMode
	if scopeMode == "" {
		scopeMode = "per_rombel"
	}
	if !validCbtQuestionRequirementScopeMode(scopeMode) {
		return db.UpsertCbtEventQuestionRequirementsRow{}, fmt.Errorf("%w: mode monitoring tidak valid", domain.ErrBadRequest)
	}
	statusFilter := in.StatusFilter
	if statusFilter == "" {
		statusFilter = "published_only"
	}
	if !validCbtQuestionRequirementStatusFilter(statusFilter) {
		return db.UpsertCbtEventQuestionRequirementsRow{}, fmt.Errorf("%w: filter status soal tidak valid", domain.ErrBadRequest)
	}
	if in.TargetPg < 0 || in.TargetEssay < 0 {
		return db.UpsertCbtEventQuestionRequirementsRow{}, fmt.Errorf("%w: target soal tidak boleh negatif", domain.ErrBadRequest)
	}
	return s.q.UpsertCbtEventQuestionRequirements(ctx, db.UpsertCbtEventQuestionRequirementsParams{
		EventID:      eventID,
		ScopeMode:    scopeMode,
		TargetPg:     in.TargetPg,
		TargetEssay:  in.TargetEssay,
		StatusFilter: statusFilter,
	})
}

func validCbtQuestionRequirementScopeMode(value string) bool {
	switch value {
	case "per_rombel", "per_level", "pool_level_subject":
		return true
	default:
		return false
	}
}

func validCbtQuestionRequirementStatusFilter(value string) bool {
	switch value {
	case "published_only", "all_progress":
		return true
	default:
		return false
	}
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

func buildCbtSopReadiness(overview CbtEventOverview) CbtSopReadiness {
	summary := overview.Event
	readiness := overview.Readiness
	eventID := uuidString(summary.ID)
	action := func(label, href string) CbtSopNextAction {
		return CbtSopNextAction{Label: label, Href: href}
	}
	stage := func(key, label string, status CbtSopStageStatus, blocking, warning int32, actions ...CbtSopNextAction) CbtSopStageReadiness {
		if actions == nil {
			actions = []CbtSopNextAction{}
		}
		return CbtSopStageReadiness{
			Key:           key,
			Label:         label,
			Status:        status,
			BlockingCount: blocking,
			WarningCount:  warning,
			NextActions:   actions,
		}
	}
	statusFor := func(ready bool, warning bool) CbtSopStageStatus {
		if ready {
			return CbtSopStageReady
		}
		if warning {
			return CbtSopStageWarning
		}
		return CbtSopStageBlocked
	}

	authoringWarning := summary.PublishedQuestions > 0 || summary.TotalQuestions > 0
	verificationReady := summary.TargetQuestionCount > 0 && summary.PublishedQuestions >= summary.TargetQuestionCount && summary.ReviewQuestions == 0
	verificationWarning := summary.ReviewQuestions > 0 || summary.ApprovedQuestions > 0 || summary.PublishedQuestions > 0
	packageWarning := summary.PackageCount > 0
	roomWarning := summary.SessionCount > 0 || summary.ParticipantCount > 0 || summary.RoomCount > 0
	tokenWarning := summary.TokenReadyCount > 0 || summary.ParticipantCount > 0
	executionStatus := CbtSopStageBlocked
	if summary.ActiveSessionCount > 0 || readiness.RuntimeStarted {
		executionStatus = CbtSopStageRunning
	} else if summary.FinishedSessionCount > 0 {
		executionStatus = CbtSopStageReady
	} else if summary.ScheduledSessionCount > 0 {
		executionStatus = CbtSopStageWarning
	}
	gradingStatus := statusFor(readiness.ResultsReady, summary.SubmittedCount > 0 || summary.ScoredCount > 0)
	resultStatus := statusFor(readiness.ResultsReady && summary.Status == "finished", readiness.ResultsReady || summary.ScoredCount > 0)
	archiveStatus := statusFor(summary.Status == "finished", readiness.ResultsReady || summary.SubmittedCount > 0)

	return CbtSopReadiness{
		EventID: eventID,
		Stages: []CbtSopStageReadiness{
			stage("draft", "Draft", CbtSopStageReady, 0, 0,
				action("Cek identitas kegiatan", fmt.Sprintf("/asesmen/kegiatan/%s", eventID)),
			),
			stage("question_authoring", "Pengisian Soal", statusFor(readiness.AuthoringReady, authoringWarning), maxInt32(summary.TargetQuestionCount-summary.PublishedQuestions, 0), summary.ReviewQuestions,
				action("Lengkapi soal sesuai target", fmt.Sprintf("/bank-soal/tambah?event_id=%s", eventID)),
				action("Cek kelengkapan soal", fmt.Sprintf("/asesmen/kegiatan/%s", eventID)),
			),
			stage("question_verification", "Telaah/Verifikasi Soal", statusFor(verificationReady, verificationWarning), summary.ReviewQuestions, summary.ApprovedQuestions,
				action("Buka antrean verifikasi", "/bank-soal/verifikasi"),
				action("Cek target soal kegiatan", fmt.Sprintf("/asesmen/kegiatan/%s", eventID)),
			),
			stage("package_ready", "Paket Siap", statusFor(readiness.PackageReady, packageWarning), summary.EmptyPackageCount, summary.PackageCount-summary.ActivePackageCount,
				action("Kelola paket kegiatan", fmt.Sprintf("/asesmen/paket?event_id=%s", eventID)),
			),
			stage("participants_rooms_ready", "Peserta & Ruang Siap", statusFor(readiness.RoomReady, roomWarning), summary.UnassignedParticipantCount+summary.MissingSeatCount+summary.RoomsWithoutProctor, summary.RoomCount,
				action("Cek sesi dan ruang", fmt.Sprintf("/asesmen/sesi?event_id=%s&readiness=not_ready", eventID)),
			),
			stage("tokens_cards_ready", "Token & Kartu Siap", statusFor(readiness.TokenReady && readiness.CardReady, tokenWarning), maxInt32(summary.ParticipantCount-summary.TokenReadyCount, 0), summary.MissingSeatCount,
				action("Cetak kartu ujian", fmt.Sprintf("/asesmen/kegiatan/%s/exam-cards", eventID)),
			),
			stage("execution", "Pelaksanaan", executionStatus, 0, summary.ScheduledSessionCount,
				action("Buka pengawasan ruang", "/asesmen/pengawasan"),
				action("Buka sesi kegiatan", fmt.Sprintf("/asesmen/sesi?event_id=%s", eventID)),
			),
			stage("grading", "Koreksi", gradingStatus, maxInt32(summary.SubmittedCount-summary.ScoredCount, 0), summary.SubmittedCount,
				action("Cek hasil kegiatan", fmt.Sprintf("/asesmen/kegiatan/%s#hasil", eventID)),
			),
			stage("result_verification", "Verifikasi Hasil", resultStatus, 0, summary.ScoredCount,
				action("Verifikasi rekap hasil", fmt.Sprintf("/asesmen/kegiatan/%s#hasil", eventID)),
			),
			stage("final_archive", "Final & Arsip", archiveStatus, 0, summary.FinishedSessionCount,
				action("Buka checklist arsip", fmt.Sprintf("/asesmen/kegiatan/%s/archive", eventID)),
			),
		},
	}
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
		return fmt.Errorf("%w: event tidak ditemukan atau masih memiliki sesi", domain.ErrConflict)
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
