package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

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
)

type nonTestAssessmentStore interface {
	ListNonTestAssessments(ctx context.Context, arg db.ListNonTestAssessmentsParams) ([]db.ListNonTestAssessmentsRow, error)
	CountNonTestAssessments(ctx context.Context, arg db.CountNonTestAssessmentsParams) (int64, error)
	GetNonTestAssessment(ctx context.Context, id pgtype.UUID) (db.GetNonTestAssessmentRow, error)
	CreateNonTestAssessment(ctx context.Context, arg db.CreateNonTestAssessmentParams) (db.NonTestAssessment, error)
	UpdateNonTestAssessment(ctx context.Context, arg db.UpdateNonTestAssessmentParams) (db.NonTestAssessment, error)
	DeleteNonTestAssessment(ctx context.Context, id pgtype.UUID) error
	ListNonTestSubmissions(ctx context.Context, assessmentID pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error)
	UpsertNonTestSubmission(ctx context.Context, arg db.UpsertNonTestSubmissionParams) (db.NonTestAssessmentSubmission, error)
}

type NonTestAssessment struct {
	q nonTestAssessmentStore
}

func NewNonTestAssessment(q *db.Queries) *NonTestAssessment {
	return &NonTestAssessment{q: q}
}

type ListNonTestAssessmentsInput struct {
	SubjectID      pgtype.UUID
	ClassID        pgtype.UUID
	Status         string
	AssessmentType string
	SearchQuery    string
	Limit          int32
	Offset         int32
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
	AssessmentID     pgtype.UUID
	StudentID        pgtype.UUID
	Status           string
	EvidenceURL      string
	EvidenceNote     string
	Score            *float64
	Feedback         string
	SubmittedAt      pgtype.Timestamptz
	GradedAt         pgtype.Timestamptz
	GradedByUsername string
}

func (s *NonTestAssessment) List(ctx context.Context, in ListNonTestAssessmentsInput) ([]db.ListNonTestAssessmentsRow, int64, error) {
	normalized := normalizeNonTestListInput(in)
	rows, err := s.q.ListNonTestAssessments(ctx, db.ListNonTestAssessmentsParams{
		SubjectID:            normalized.SubjectID,
		ClassID:              normalized.ClassID,
		StatusFilter:         normalized.Status,
		AssessmentTypeFilter: normalized.AssessmentType,
		SearchQuery:          normalized.SearchQuery,
		LimitCount:           normalized.Limit,
		OffsetCount:          normalized.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountNonTestAssessments(ctx, db.CountNonTestAssessmentsParams{
		SubjectID:            normalized.SubjectID,
		ClassID:              normalized.ClassID,
		StatusFilter:         normalized.Status,
		AssessmentTypeFilter: normalized.AssessmentType,
		SearchQuery:          normalized.SearchQuery,
	})
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
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

func (s *NonTestAssessment) UpsertSubmission(ctx context.Context, input SaveNonTestSubmissionInput) (db.NonTestAssessmentSubmission, error) {
	normalized, err := normalizeNonTestSubmissionInput(input)
	if err != nil {
		return db.NonTestAssessmentSubmission{}, err
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

func normalizeNonTestListInput(in ListNonTestAssessmentsInput) ListNonTestAssessmentsInput {
	in.Status = strings.TrimSpace(in.Status)
	in.AssessmentType = strings.TrimSpace(in.AssessmentType)
	in.SearchQuery = strings.TrimSpace(in.SearchQuery)
	if !allowedNonTestAssessmentStatuses[in.Status] {
		in.Status = ""
	}
	if !allowedNonTestAssessmentTypes[in.AssessmentType] {
		in.AssessmentType = ""
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
