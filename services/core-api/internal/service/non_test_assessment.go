package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

var (
	allowedNonTestAssessmentTypes = map[string]bool{
		"praktik":    true,
		"portofolio": true,
		"proyek":     true,
		"penugasan":  true,
		"observasi":  true,
		"lainnya":    true,
	}
	allowedNonTestAssessmentModes = map[string]bool{
		"beginner": true,
		"advance":  true,
	}
	allowedNonTestAssessmentStatuses = map[string]bool{
		"draft":    true,
		"active":   true,
		"closed":   true,
		"archived": true,
	}
	allowedNonTestSubmissionStatuses = map[string]bool{
		"assigned":  true,
		"submitted": true,
		"reviewed":  true,
		"returned":  true,
	}
	allowedNonTestAssessmentSyncFilters = map[string]bool{
		"needs_sync": true,
	}
)

type nonTestAssessmentStore interface {
	WithTx(tx pgx.Tx) *db.Queries
	ListNonTestAssessments(ctx context.Context, arg db.ListNonTestAssessmentsParams) ([]db.ListNonTestAssessmentsRow, error)
	CountNonTestAssessments(ctx context.Context, arg db.CountNonTestAssessmentsParams) (int64, error)
	GetNonTestAssessment(ctx context.Context, id pgtype.UUID) (db.GetNonTestAssessmentRow, error)
	CreateNonTestAssessment(ctx context.Context, arg db.CreateNonTestAssessmentParams) (db.NonTestAssessment, error)
	UpdateNonTestAssessment(ctx context.Context, arg db.UpdateNonTestAssessmentParams) (db.NonTestAssessment, error)
	DeleteNonTestAssessment(ctx context.Context, id pgtype.UUID) error
	ListNonTestSubmissions(ctx context.Context, assessmentID pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error)
	GenerateNonTestSubmissionsForClass(ctx context.Context, arg db.GenerateNonTestSubmissionsForClassParams) ([]db.NonTestAssessmentSubmission, error)
	UpsertNonTestSubmission(ctx context.Context, arg db.UpsertNonTestSubmissionParams) (db.NonTestAssessmentSubmission, error)
	StudentBelongsToClass(ctx context.Context, arg db.StudentBelongsToClassParams) (bool, error)
	GetGradeAssignmentByClassSubject(ctx context.Context, arg db.GetGradeAssignmentByClassSubjectParams) (db.GetGradeAssignmentByClassSubjectRow, error)
	TeacherOwnsClassSubject(ctx context.Context, arg db.TeacherOwnsClassSubjectParams) (bool, error)
	TeacherOwnsNonTestAssessment(ctx context.Context, arg db.TeacherOwnsNonTestAssessmentParams) (bool, error)
	GetGradeAssignmentFinalization(ctx context.Context, assignmentID pgtype.UUID) (db.GradeAssignmentFinalization, error)
	GetGradeComponent(ctx context.Context, id pgtype.UUID) (db.GradeComponent, error)
	CreateGradeComponent(ctx context.Context, arg db.CreateGradeComponentParams) (db.GradeComponent, error)
	UpdateGradeComponentPublishState(ctx context.Context, arg db.UpdateGradeComponentPublishStateParams) (db.GradeComponent, error)
	UpsertGradeEntry(ctx context.Context, arg db.UpsertGradeEntryParams) (db.GradeEntry, error)
	MarkNonTestAssessmentGradeSync(ctx context.Context, arg db.MarkNonTestAssessmentGradeSyncParams) (db.NonTestAssessment, error)
}

type nonTestAssessmentTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type NonTestAssessment struct {
	q  nonTestAssessmentStore
	tx nonTestAssessmentTxStarter
}

func NewNonTestAssessment(q *db.Queries) *NonTestAssessment {
	return &NonTestAssessment{q: q}
}

func NewNonTestAssessmentWithPool(pool *pgxpool.Pool) *NonTestAssessment {
	return &NonTestAssessment{q: db.New(pool), tx: pool}
}

type ListNonTestAssessmentsInput struct {
	SubjectID         pgtype.UUID
	ClassID           pgtype.UUID
	Status            string
	AssessmentType    string
	SyncFilter        string
	SearchQuery       string
	Limit             int32
	Offset            int32
	TeacherEmployeeID pgtype.UUID
}

type SaveNonTestAssessmentInput struct {
	ID                   pgtype.UUID
	SubjectID            pgtype.UUID
	ClassID              pgtype.UUID
	AssessmentType       string
	Title                string
	Description          string
	InstructionHTML      string
	RubricHTML           string
	EvidenceRequirements string
	Mode                 string
	ScoringScale         string
	MaxScore             float64
	Weight               float64
	DueAt                pgtype.Timestamptz
	Status               string
	CreatedByUsername    string
	AssessorUsername     string
	Checklist            []byte
}

type SaveNonTestSubmissionInput struct {
	AssessmentID      pgtype.UUID
	StudentID         pgtype.UUID
	Status            string
	EvidenceURL       string
	EvidenceNote      string
	Score             *float64
	Feedback          string
	SubmittedAt       pgtype.Timestamptz
	GradedAt          pgtype.Timestamptz
	GradedByUsername  string
	TeacherEmployeeID pgtype.UUID
}

type SyncNonTestAssessmentToGradeResult struct {
	AssessmentID     string `json:"assessment_id"`
	AssignmentID     string `json:"assignment_id"`
	GradeComponentID string `json:"grade_component_id"`
	CreatedComponent bool   `json:"created_component"`
	SyncedEntries    int    `json:"synced_entries"`
	SkippedEntries   int    `json:"skipped_entries"`
	IsPublished      bool   `json:"is_published"`
}

func (s *NonTestAssessment) List(ctx context.Context, in ListNonTestAssessmentsInput) ([]db.ListNonTestAssessmentsRow, int64, error) {
	normalized := normalizeNonTestListInput(in)
	rows, err := s.q.ListNonTestAssessments(ctx, db.ListNonTestAssessmentsParams{
		SubjectID:            normalized.SubjectID,
		ClassID:              normalized.ClassID,
		StatusFilter:         normalized.Status,
		AssessmentTypeFilter: normalized.AssessmentType,
		SyncFilter:           normalized.SyncFilter,
		SearchQuery:          normalized.SearchQuery,
		LimitCount:           normalized.Limit,
		OffsetCount:          normalized.Offset,
		TeacherEmployeeID:    normalized.TeacherEmployeeID,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountNonTestAssessments(ctx, db.CountNonTestAssessmentsParams{
		SubjectID:            normalized.SubjectID,
		ClassID:              normalized.ClassID,
		StatusFilter:         normalized.Status,
		AssessmentTypeFilter: normalized.AssessmentType,
		SyncFilter:           normalized.SyncFilter,
		SearchQuery:          normalized.SearchQuery,
		TeacherEmployeeID:    normalized.TeacherEmployeeID,
	})
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *NonTestAssessment) TeacherOwnsClassSubject(ctx context.Context, classID, subjectID, teacherEmployeeID pgtype.UUID) (bool, error) {
	if !classID.Valid || !subjectID.Valid || !teacherEmployeeID.Valid {
		return false, nil
	}
	return s.q.TeacherOwnsClassSubject(ctx, db.TeacherOwnsClassSubjectParams{
		ClassID:           classID,
		SubjectID:         subjectID,
		TeacherEmployeeID: teacherEmployeeID,
	})
}

func (s *NonTestAssessment) TeacherOwnsAssessment(ctx context.Context, assessmentID, teacherEmployeeID pgtype.UUID) (bool, error) {
	if !assessmentID.Valid || !teacherEmployeeID.Valid {
		return false, nil
	}
	return s.q.TeacherOwnsNonTestAssessment(ctx, db.TeacherOwnsNonTestAssessmentParams{
		ID:                assessmentID,
		TeacherEmployeeID: teacherEmployeeID,
	})
}

func (s *NonTestAssessment) Get(ctx context.Context, id pgtype.UUID) (db.GetNonTestAssessmentRow, error) {
	return s.q.GetNonTestAssessment(ctx, id)
}

func (s *NonTestAssessment) Create(ctx context.Context, input SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	normalized, err := normalizeNonTestAssessmentInput(input)
	if err != nil {
		return db.NonTestAssessment{}, err
	}
	return s.q.CreateNonTestAssessment(ctx, db.CreateNonTestAssessmentParams{
		SubjectID:            normalized.SubjectID,
		ClassID:              normalized.ClassID,
		AssessmentType:       normalized.AssessmentType,
		Title:                normalized.Title,
		Description:          normalized.Description,
		InstructionHtml:      normalized.InstructionHTML,
		RubricHtml:           normalized.RubricHTML,
		EvidenceRequirements: normalized.EvidenceRequirements,
		Mode:                 normalized.Mode,
		ScoringScale:         normalized.ScoringScale,
		MaxScore:             pgNumeric(normalized.MaxScore),
		Weight:               pgNumeric(normalized.Weight),
		DueAt:                normalized.DueAt,
		Status:               normalized.Status,
		CreatedByUsername:    normalized.CreatedByUsername,
		AssessorUsername:     normalized.AssessorUsername,
		Checklist:            normalized.Checklist,
	})
}

func (s *NonTestAssessment) Update(ctx context.Context, input SaveNonTestAssessmentInput) (db.NonTestAssessment, error) {
	if !input.ID.Valid {
		return db.NonTestAssessment{}, errors.New("id asesmen tidak valid")
	}
	current, err := s.q.GetNonTestAssessment(ctx, input.ID)
	if err != nil {
		return db.NonTestAssessment{}, err
	}
	if strings.TrimSpace(input.CreatedByUsername) == "" {
		input.CreatedByUsername = current.CreatedByUsername
	}
	normalized, err := normalizeNonTestAssessmentInput(input)
	if err != nil {
		return db.NonTestAssessment{}, err
	}
	return s.q.UpdateNonTestAssessment(ctx, db.UpdateNonTestAssessmentParams{
		ID:                   normalized.ID,
		SubjectID:            normalized.SubjectID,
		ClassID:              normalized.ClassID,
		AssessmentType:       normalized.AssessmentType,
		Title:                normalized.Title,
		Description:          normalized.Description,
		InstructionHtml:      normalized.InstructionHTML,
		RubricHtml:           normalized.RubricHTML,
		EvidenceRequirements: normalized.EvidenceRequirements,
		Mode:                 normalized.Mode,
		ScoringScale:         normalized.ScoringScale,
		MaxScore:             pgNumeric(normalized.MaxScore),
		Weight:               pgNumeric(normalized.Weight),
		DueAt:                normalized.DueAt,
		Status:               normalized.Status,
		CreatedByUsername:    normalized.CreatedByUsername,
		AssessorUsername:     normalized.AssessorUsername,
		Checklist:            normalized.Checklist,
	})
}

func (s *NonTestAssessment) Delete(ctx context.Context, id pgtype.UUID) error {
	if !id.Valid {
		return errors.New("id asesmen tidak valid")
	}
	return s.q.DeleteNonTestAssessment(ctx, id)
}

func (s *NonTestAssessment) ListSubmissions(ctx context.Context, assessmentID pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error) {
	if !assessmentID.Valid {
		return nil, errors.New("id asesmen tidak valid")
	}
	return s.q.ListNonTestSubmissions(ctx, assessmentID)
}

func (s *NonTestAssessment) GenerateSubmissions(ctx context.Context, assessmentID pgtype.UUID, classID, teacherEmployeeID pgtype.UUID) ([]db.NonTestAssessmentSubmission, error) {
	if !assessmentID.Valid {
		return nil, errors.New("id asesmen tidak valid")
	}
	assessment, err := s.q.GetNonTestAssessment(ctx, assessmentID)
	if err != nil {
		return nil, err
	}
	if !assessment.ClassID.Valid {
		return nil, errors.New("kelas asesmen wajib dipilih sebelum menyiapkan siswa")
	}
	if classID.Valid && classID != assessment.ClassID {
		return nil, errors.New("class_id harus sesuai dengan kelas asesmen")
	}
	if teacherEmployeeID.Valid {
		owns, err := s.TeacherOwnsClassSubject(ctx, assessment.ClassID, assessment.SubjectID, teacherEmployeeID)
		if err != nil {
			return nil, err
		}
		if !owns {
			return nil, errors.New("akses ditolak")
		}
	}
	return s.q.GenerateNonTestSubmissionsForClass(ctx, db.GenerateNonTestSubmissionsForClassParams{
		AssessmentID: assessmentID,
		ClassID:      assessment.ClassID,
	})
}

func (s *NonTestAssessment) UpsertSubmission(ctx context.Context, input SaveNonTestSubmissionInput) (db.NonTestAssessmentSubmission, error) {
	normalized, err := normalizeNonTestSubmissionInput(input)
	if err != nil {
		return db.NonTestAssessmentSubmission{}, err
	}
	assessment, err := s.q.GetNonTestAssessment(ctx, normalized.AssessmentID)
	if err != nil {
		return db.NonTestAssessmentSubmission{}, err
	}
	if normalized.Score != nil {
		maxScore, ok := numericFloat64(assessment.MaxScore)
		if ok && *normalized.Score > maxScore {
			return db.NonTestAssessmentSubmission{}, errors.New("nilai melebihi skor maksimum asesmen")
		}
	}
	if !assessment.ClassID.Valid {
		return db.NonTestAssessmentSubmission{}, errors.New("kelas asesmen wajib dipilih sebelum menyimpan pengumpulan")
	}
	if normalized.TeacherEmployeeID.Valid {
		owns, err := s.TeacherOwnsClassSubject(ctx, assessment.ClassID, assessment.SubjectID, normalized.TeacherEmployeeID)
		if err != nil {
			return db.NonTestAssessmentSubmission{}, err
		}
		if !owns {
			return db.NonTestAssessmentSubmission{}, errors.New("akses ditolak")
		}
	}
	belongs, err := s.q.StudentBelongsToClass(ctx, db.StudentBelongsToClassParams{
		StudentID: normalized.StudentID,
		ClassID:   assessment.ClassID,
	})
	if err != nil {
		return db.NonTestAssessmentSubmission{}, err
	}
	if !belongs {
		return db.NonTestAssessmentSubmission{}, errors.New("siswa tidak berada di kelas asesmen")
	}
	score := pgtype.Numeric{}
	if normalized.Score != nil {
		score = pgNumeric(*normalized.Score)
	}
	return s.q.UpsertNonTestSubmission(ctx, db.UpsertNonTestSubmissionParams{
		AssessmentID:     normalized.AssessmentID,
		StudentID:        normalized.StudentID,
		Status:           normalized.Status,
		EvidenceUrl:      normalized.EvidenceURL,
		EvidenceNote:     normalized.EvidenceNote,
		Score:            score,
		Feedback:         normalized.Feedback,
		SubmittedAt:      normalized.SubmittedAt,
		GradedAt:         normalized.GradedAt,
		GradedByUsername: normalized.GradedByUsername,
	})
}

func (s *NonTestAssessment) SyncToGrade(ctx context.Context, assessmentID, teacherEmployeeID pgtype.UUID, syncedBy string, publish bool) (SyncNonTestAssessmentToGradeResult, error) {
	if !assessmentID.Valid {
		return SyncNonTestAssessmentToGradeResult{}, errors.New("id asesmen tidak valid")
	}
	syncedBy = strings.TrimSpace(syncedBy)
	if syncedBy == "" {
		syncedBy = "system"
	}
	if s.tx != nil {
		tx, err := s.tx.Begin(ctx)
		if err != nil {
			return SyncNonTestAssessmentToGradeResult{}, err
		}
		defer func() { _ = tx.Rollback(ctx) }()

		result, err := s.syncToGradeWithStore(ctx, s.q.WithTx(tx), assessmentID, teacherEmployeeID, syncedBy, publish)
		if err != nil {
			return SyncNonTestAssessmentToGradeResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return SyncNonTestAssessmentToGradeResult{}, err
		}
		return result, nil
	}
	return s.syncToGradeWithStore(ctx, s.q, assessmentID, teacherEmployeeID, syncedBy, publish)
}

type nonTestGradeEntryDraft struct {
	studentID pgtype.UUID
	score     float64
	notes     string
}

func (s *NonTestAssessment) syncToGradeWithStore(ctx context.Context, q nonTestAssessmentStore, assessmentID, teacherEmployeeID pgtype.UUID, syncedBy string, publish bool) (SyncNonTestAssessmentToGradeResult, error) {
	assessment, err := q.GetNonTestAssessment(ctx, assessmentID)
	if err != nil {
		return SyncNonTestAssessmentToGradeResult{}, err
	}
	if !assessment.ClassID.Valid {
		return SyncNonTestAssessmentToGradeResult{}, errors.New("kelas asesmen wajib dipilih sebelum sinkron nilai")
	}
	assignment, err := q.GetGradeAssignmentByClassSubject(ctx, db.GetGradeAssignmentByClassSubjectParams{
		ClassID:   assessment.ClassID,
		SubjectID: assessment.SubjectID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return SyncNonTestAssessmentToGradeResult{}, errors.New("assignment kelas-mapel belum tersedia di master akademik")
	}
	if err != nil {
		return SyncNonTestAssessmentToGradeResult{}, err
	}
	if teacherEmployeeID.Valid && assignment.TeacherEmployeeID != teacherEmployeeID {
		return SyncNonTestAssessmentToGradeResult{}, errors.New("akses ditolak")
	}
	if _, err := q.GetGradeAssignmentFinalization(ctx, assignment.ID); err == nil {
		return SyncNonTestAssessmentToGradeResult{}, errors.New("assignment sudah difinalisasi, buka finalisasi terlebih dahulu")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return SyncNonTestAssessmentToGradeResult{}, err
	}
	maxScore, ok := numericFloat64(assessment.MaxScore)
	if !ok || maxScore <= 0 {
		return SyncNonTestAssessmentToGradeResult{}, errors.New("skor maksimum asesmen tidak valid")
	}
	weight, ok := numericFloat64(assessment.Weight)
	if !ok || weight <= 0 {
		weight = 1
	}
	submissions, err := q.ListNonTestSubmissions(ctx, assessmentID)
	if err != nil {
		return SyncNonTestAssessmentToGradeResult{}, err
	}
	entries := make([]nonTestGradeEntryDraft, 0, len(submissions))
	skipped := 0
	for _, row := range submissions {
		score, ok := numericFloat64(row.Score)
		if row.Status != "reviewed" || !ok {
			skipped++
			continue
		}
		if score < 0 {
			return SyncNonTestAssessmentToGradeResult{}, errors.New("nilai tidak boleh negatif")
		}
		if score > maxScore {
			return SyncNonTestAssessmentToGradeResult{}, fmt.Errorf("nilai melebihi skor maksimum asesmen %.2f", maxScore)
		}
		notes := strings.TrimSpace(row.Feedback)
		if notes == "" {
			notes = strings.TrimSpace(row.EvidenceNote)
		}
		entries = append(entries, nonTestGradeEntryDraft{
			studentID: row.StudentID,
			score:     score,
			notes:     notes,
		})
	}
	if len(entries) == 0 {
		return SyncNonTestAssessmentToGradeResult{}, errors.New("belum ada nilai reviewed yang siap dikirim ke nilai")
	}
	component, createdComponent, err := s.ensureNonTestGradeComponent(ctx, q, assessment, assignment.ID, maxScore, weight, publish)
	if err != nil {
		return SyncNonTestAssessmentToGradeResult{}, err
	}
	for _, entry := range entries {
		if entry.score > component.MaxScore {
			return SyncNonTestAssessmentToGradeResult{}, fmt.Errorf("nilai melebihi skor maksimum komponen %.2f", component.MaxScore)
		}
		if _, err := q.UpsertGradeEntry(ctx, db.UpsertGradeEntryParams{
			ComponentID: component.ID,
			StudentID:   entry.studentID,
			Score:       pgtype.Float8{Float64: entry.score, Valid: true},
			Notes:       entry.notes,
			GradedBy:    syncedBy,
		}); err != nil {
			return SyncNonTestAssessmentToGradeResult{}, err
		}
	}
	if _, err := q.MarkNonTestAssessmentGradeSync(ctx, db.MarkNonTestAssessmentGradeSyncParams{
		ID:               assessmentID,
		GradeComponentID: component.ID,
		GradeSyncedBy:    syncedBy,
	}); err != nil {
		return SyncNonTestAssessmentToGradeResult{}, err
	}
	return SyncNonTestAssessmentToGradeResult{
		AssessmentID:     assessmentID.String(),
		AssignmentID:     assignment.ID.String(),
		GradeComponentID: component.ID.String(),
		CreatedComponent: createdComponent,
		SyncedEntries:    len(entries),
		SkippedEntries:   skipped,
		IsPublished:      component.IsPublished,
	}, nil
}

func (s *NonTestAssessment) ensureNonTestGradeComponent(ctx context.Context, q nonTestAssessmentStore, assessment db.GetNonTestAssessmentRow, assignmentID pgtype.UUID, maxScore, weight float64, publish bool) (db.GradeComponent, bool, error) {
	if assessment.GradeComponentID.Valid {
		component, err := q.GetGradeComponent(ctx, assessment.GradeComponentID)
		if err == nil && component.AssignmentID == assignmentID {
			if component.IsPublished != publish {
				component, err = q.UpdateGradeComponentPublishState(ctx, db.UpdateGradeComponentPublishStateParams{
					ID:          component.ID,
					IsPublished: publish,
				})
				if err != nil {
					return db.GradeComponent{}, false, err
				}
			}
			return component, false, nil
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return db.GradeComponent{}, false, err
		}
	}
	component, err := q.CreateGradeComponent(ctx, db.CreateGradeComponentParams{
		AssignmentID: assignmentID,
		Title:        nonTestAssessmentGradeTitle(assessment.Title),
		Category:     nonTestAssessmentGradeCategory(assessment.AssessmentType),
		Weight:       weight,
		MaxScore:     maxScore,
		IsPublished:  publish,
	})
	if err != nil {
		return db.GradeComponent{}, false, err
	}
	return component, true, nil
}

func normalizeNonTestListInput(in ListNonTestAssessmentsInput) ListNonTestAssessmentsInput {
	in.Status = strings.TrimSpace(in.Status)
	in.AssessmentType = strings.TrimSpace(in.AssessmentType)
	in.SyncFilter = strings.TrimSpace(in.SyncFilter)
	in.SearchQuery = strings.TrimSpace(in.SearchQuery)
	if !allowedNonTestAssessmentStatuses[in.Status] {
		in.Status = ""
	}
	if !allowedNonTestAssessmentTypes[in.AssessmentType] {
		in.AssessmentType = ""
	}
	if !allowedNonTestAssessmentSyncFilters[in.SyncFilter] {
		in.SyncFilter = ""
	}
	if in.Limit <= 0 || in.Limit > 100 {
		in.Limit = 25
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	return in
}

func normalizeNonTestAssessmentInput(input SaveNonTestAssessmentInput) (SaveNonTestAssessmentInput, error) {
	input.AssessmentType = strings.TrimSpace(input.AssessmentType)
	if input.AssessmentType == "" {
		input.AssessmentType = "penugasan"
	}
	if !allowedNonTestAssessmentTypes[input.AssessmentType] {
		return SaveNonTestAssessmentInput{}, errors.New("bentuk asesmen non-tes tidak valid")
	}
	if !input.SubjectID.Valid {
		return SaveNonTestAssessmentInput{}, errors.New("mata pelajaran wajib dipilih")
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return SaveNonTestAssessmentInput{}, errors.New("judul asesmen wajib diisi")
	}
	input.Description = strings.TrimSpace(input.Description)
	input.InstructionHTML = strings.TrimSpace(input.InstructionHTML)
	input.RubricHTML = strings.TrimSpace(input.RubricHTML)
	input.EvidenceRequirements = strings.TrimSpace(input.EvidenceRequirements)
	input.Mode = strings.TrimSpace(input.Mode)
	if input.Mode == "" {
		input.Mode = "beginner"
	}
	if !allowedNonTestAssessmentModes[input.Mode] {
		return SaveNonTestAssessmentInput{}, errors.New("mode asesmen tidak valid")
	}
	input.ScoringScale = strings.TrimSpace(input.ScoringScale)
	if input.ScoringScale == "" {
		input.ScoringScale = "0_100"
	}
	if input.MaxScore <= 0 {
		input.MaxScore = 100
	}
	if input.Weight <= 0 {
		input.Weight = 1
	}
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = "draft"
	}
	if !allowedNonTestAssessmentStatuses[input.Status] {
		return SaveNonTestAssessmentInput{}, errors.New("status asesmen tidak valid")
	}
	input.CreatedByUsername = strings.TrimSpace(input.CreatedByUsername)
	input.AssessorUsername = strings.TrimSpace(input.AssessorUsername)
	checklist, err := normalizeChecklistJSON(input.Checklist)
	if err != nil {
		return SaveNonTestAssessmentInput{}, err
	}
	input.Checklist = checklist
	return input, nil
}

func normalizeNonTestSubmissionInput(input SaveNonTestSubmissionInput) (SaveNonTestSubmissionInput, error) {
	if !input.AssessmentID.Valid {
		return SaveNonTestSubmissionInput{}, errors.New("id asesmen tidak valid")
	}
	if !input.StudentID.Valid {
		return SaveNonTestSubmissionInput{}, errors.New("siswa wajib dipilih")
	}
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = "assigned"
	}
	if !allowedNonTestSubmissionStatuses[input.Status] {
		return SaveNonTestSubmissionInput{}, errors.New("status pengumpulan tidak valid")
	}
	input.EvidenceURL = strings.TrimSpace(input.EvidenceURL)
	input.EvidenceNote = strings.TrimSpace(input.EvidenceNote)
	input.Feedback = strings.TrimSpace(input.Feedback)
	input.GradedByUsername = strings.TrimSpace(input.GradedByUsername)
	if input.Score != nil && *input.Score < 0 {
		return SaveNonTestSubmissionInput{}, errors.New("nilai tidak boleh negatif")
	}
	if input.Status == "reviewed" && input.Score == nil {
		return SaveNonTestSubmissionInput{}, errors.New("nilai wajib diisi sebelum ditandai sudah dinilai")
	}
	return input, nil
}

func normalizeChecklistJSON(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return []byte("[]"), nil
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, errors.New("checklist observasi harus berupa JSON valid")
	}
	if _, ok := decoded.([]any); !ok {
		return nil, errors.New("checklist observasi harus berupa daftar")
	}
	return raw, nil
}

func nonTestAssessmentGradeTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Non-Tes"
	}
	if strings.HasPrefix(strings.ToLower(title), "non-tes:") {
		return title
	}
	return "Non-Tes: " + title
}

func nonTestAssessmentGradeCategory(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "praktik":
		return "practice"
	case "proyek", "portofolio":
		return "project"
	case "penugasan":
		return "assignment"
	case "observasi":
		return "attitude"
	default:
		return "other"
	}
}

func numericFloat64(value pgtype.Numeric) (float64, bool) {
	if !value.Valid || value.Int == nil {
		return 0, false
	}
	ratio := new(big.Rat).SetInt(value.Int)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absInt32Service(value.Exp))), nil)
	if value.Exp >= 0 {
		ratio.Mul(ratio, new(big.Rat).SetInt(scale))
	} else {
		ratio.Quo(ratio, new(big.Rat).SetInt(scale))
	}
	out, _ := ratio.Float64()
	return out, true
}

func absInt32Service(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}
