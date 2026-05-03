package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type NonTestAssessment struct {
	svc   nonTestAssessmentService
	audit cbtAuthoringAuditWriter
}

type nonTestAssessmentService interface {
	List(ctx context.Context, in service.ListNonTestAssessmentsInput) ([]db.ListNonTestAssessmentsRow, int64, error)
	Get(ctx context.Context, id pgtype.UUID) (db.GetNonTestAssessmentRow, error)
	Create(ctx context.Context, input service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error)
	Update(ctx context.Context, input service.SaveNonTestAssessmentInput) (db.NonTestAssessment, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	ListSubmissions(ctx context.Context, assessmentID pgtype.UUID) ([]db.ListNonTestSubmissionsRow, error)
	GenerateSubmissions(ctx context.Context, assessmentID pgtype.UUID, classID pgtype.UUID) ([]db.NonTestAssessmentSubmission, error)
	UpsertSubmission(ctx context.Context, input service.SaveNonTestSubmissionInput) (db.NonTestAssessmentSubmission, error)
	SyncToGrade(ctx context.Context, assessmentID, teacherEmployeeID pgtype.UUID, syncedBy string, publish bool) (service.SyncNonTestAssessmentToGradeResult, error)
}

func NewNonTestAssessment(svc *service.NonTestAssessment, audit ...cbtAuthoringAuditWriter) *NonTestAssessment {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &NonTestAssessment{svc: svc, audit: writer}
}

type nonTestAssessmentBody struct {
	SubjectID            string   `json:"subject_id"`
	ClassID              string   `json:"class_id"`
	AssessmentType       string   `json:"assessment_type"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	InstructionHTML      string   `json:"instruction_html"`
	RubricHTML           string   `json:"rubric_html"`
	EvidenceRequirements string   `json:"evidence_requirements"`
	Mode                 string   `json:"mode"`
	ScoringScale         string   `json:"scoring_scale"`
	MaxScore             float64  `json:"max_score"`
	Weight               float64  `json:"weight"`
	DueAt                string   `json:"due_at"`
	Status               string   `json:"status"`
	AssessorUsername     string   `json:"assessor_username"`
	Checklist            []string `json:"checklist"`
}

type nonTestSubmissionBody struct {
	StudentID    string   `json:"student_id"`
	Status       string   `json:"status"`
	EvidenceURL  string   `json:"evidence_url"`
	EvidenceNote string   `json:"evidence_note"`
	Score        *float64 `json:"score"`
	Feedback     string   `json:"feedback"`
	SubmittedAt  string   `json:"submitted_at"`
	GradedAt     string   `json:"graded_at"`
}

type nonTestGenerateSubmissionsBody struct {
	ClassID string `json:"class_id"`
}

type nonTestSyncGradeBody struct {
	IsPublished *bool `json:"is_published"`
}

func (h *NonTestAssessment) List(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	input, err := nonTestAssessmentListInputFromRequest(r)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	rows, total, err := h.svc.List(r.Context(), input)
	if err != nil {
		api.Internal(w, err)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, serializeNonTestAssessmentListRow(row))
	}
	api.OK(w, map[string]any{
		"items": items,
		"meta": map[string]any{
			"total":  total,
			"limit":  input.Limit,
			"offset": input.Offset,
		},
	})
}

func (h *NonTestAssessment) Get(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeClientError(w, err, "Asesmen non-tes tidak ditemukan")
		return
	}
	api.OK(w, serializeNonTestAssessmentDetailRow(row))
}

func (h *NonTestAssessment) Create(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	input, err := h.inputFromRequest(r, pgtype.UUID{})
	if err != nil {
		writeClientError(w, err, "Data asesmen non-tes tidak valid")
		return
	}
	row, err := h.svc.Create(r.Context(), input)
	if err != nil {
		writeClientError(w, err, "Pembuatan asesmen non-tes tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_NON_TEST_ASSESSMENT_CREATE", "non_test_assessment", pgUUIDString(row.ID), map[string]any{
		"subject_id":      pgUUIDString(row.SubjectID),
		"assessment_type": row.AssessmentType,
		"title":           row.Title,
	})
	api.Created(w, serializeNonTestAssessment(row))
}

func (h *NonTestAssessment) Update(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	input, err := h.inputFromRequest(r, id)
	if err != nil {
		writeClientError(w, err, "Data asesmen non-tes tidak valid")
		return
	}
	row, err := h.svc.Update(r.Context(), input)
	if err != nil {
		writeClientError(w, err, "Perubahan asesmen non-tes tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_NON_TEST_ASSESSMENT_UPDATE", "non_test_assessment", pgUUIDString(row.ID), map[string]any{
		"subject_id":      pgUUIDString(row.SubjectID),
		"assessment_type": row.AssessmentType,
		"status":          row.Status,
	})
	api.OK(w, serializeNonTestAssessment(row))
}

func (h *NonTestAssessment) Delete(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeClientError(w, err, "Penghapusan asesmen non-tes tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_NON_TEST_ASSESSMENT_DELETE", "non_test_assessment", pgUUIDString(id), nil)
	api.NoContent(w)
}

func (h *NonTestAssessment) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.ListSubmissions(r.Context(), id)
	if err != nil {
		writeClientError(w, err, "Data pengumpulan asesmen non-tes tidak valid")
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, serializeNonTestSubmissionListRow(row))
	}
	api.OK(w, map[string]any{"items": items})
}

func (h *NonTestAssessment) GenerateSubmissions(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	assessmentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body nonTestGenerateSubmissionsBody
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			api.BadRequest(w, "invalid json")
			return
		}
	}
	classID := pgtype.UUID{}
	if strings.TrimSpace(body.ClassID) != "" {
		classID, err = parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "class_id invalid")
			return
		}
	}
	rows, err := h.svc.GenerateSubmissions(r.Context(), assessmentID, classID)
	if err != nil {
		writeClientError(w, err, "Penyiapan siswa asesmen non-tes tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_NON_TEST_SUBMISSIONS_GENERATE", "non_test_assessment", pgUUIDString(assessmentID), map[string]any{
		"class_id":      pgUUIDString(classID),
		"created_count": len(rows),
	})
	api.OK(w, map[string]any{
		"status":        "generated",
		"created_count": len(rows),
	})
}

func (h *NonTestAssessment) SyncGrade(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	assessmentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	publish := true
	var body nonTestSyncGradeBody
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			api.BadRequest(w, "invalid json")
			return
		}
	}
	if body.IsPublished != nil {
		publish = *body.IsPublished
	}
	result, err := h.svc.SyncToGrade(r.Context(), assessmentID, gradeTeacherEmployeeID(r), currentUsername(r), publish)
	if err != nil {
		writeClientError(w, err, "Sinkron nilai non-tes tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_NON_TEST_GRADE_SYNC", "non_test_assessment", pgUUIDString(assessmentID), map[string]any{
		"grade_component_id": result.GradeComponentID,
		"synced_entries":     result.SyncedEntries,
		"skipped_entries":    result.SkippedEntries,
		"is_published":       result.IsPublished,
	})
	api.OK(w, result)
}

func (h *NonTestAssessment) UpsertSubmission(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	assessmentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body nonTestSubmissionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	studentID, err := parseUUID(body.StudentID)
	if err != nil {
		api.BadRequest(w, "student_id invalid")
		return
	}
	submittedAt, err := parseOptionalTime(body.SubmittedAt)
	if err != nil {
		api.BadRequest(w, "submitted_at invalid")
		return
	}
	gradedAt, err := parseOptionalTime(body.GradedAt)
	if err != nil {
		api.BadRequest(w, "graded_at invalid")
		return
	}
	if strings.TrimSpace(body.Status) == "submitted" && !submittedAt.Valid {
		_ = submittedAt.Scan(time.Now())
	}
	if strings.TrimSpace(body.Status) == "reviewed" && !gradedAt.Valid {
		_ = gradedAt.Scan(time.Now())
	}
	row, err := h.svc.UpsertSubmission(r.Context(), service.SaveNonTestSubmissionInput{
		AssessmentID:     assessmentID,
		StudentID:        studentID,
		Status:           body.Status,
		EvidenceURL:      body.EvidenceURL,
		EvidenceNote:     body.EvidenceNote,
		Score:            body.Score,
		Feedback:         body.Feedback,
		SubmittedAt:      submittedAt,
		GradedAt:         gradedAt,
		GradedByUsername: currentUsername(r),
	})
	if err != nil {
		writeClientError(w, err, "Data pengumpulan asesmen non-tes tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_NON_TEST_SUBMISSION_UPSERT", "non_test_assessment", pgUUIDString(assessmentID), map[string]any{
		"student_id": pgUUIDString(row.StudentID),
		"status":     row.Status,
	})
	api.OK(w, serializeNonTestSubmission(row))
}

func (h *NonTestAssessment) inputFromRequest(r *http.Request, id pgtype.UUID) (service.SaveNonTestAssessmentInput, error) {
	var body nonTestAssessmentBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return service.SaveNonTestAssessmentInput{}, fmt.Errorf("invalid json")
	}
	subjectID, err := parseUUID(body.SubjectID)
	if err != nil {
		return service.SaveNonTestAssessmentInput{}, fmt.Errorf("subject_id invalid")
	}
	classID := pgtype.UUID{}
	if strings.TrimSpace(body.ClassID) != "" {
		if classID, err = parseUUID(body.ClassID); err != nil {
			return service.SaveNonTestAssessmentInput{}, fmt.Errorf("class_id invalid")
		}
	}
	dueAt, err := parseOptionalTime(body.DueAt)
	if err != nil {
		return service.SaveNonTestAssessmentInput{}, fmt.Errorf("due_at invalid")
	}
	checklist, err := json.Marshal(body.Checklist)
	if err != nil {
		return service.SaveNonTestAssessmentInput{}, err
	}
	return service.SaveNonTestAssessmentInput{
		ID:                   id,
		SubjectID:            subjectID,
		ClassID:              classID,
		AssessmentType:       body.AssessmentType,
		Title:                body.Title,
		Description:          body.Description,
		InstructionHTML:      body.InstructionHTML,
		RubricHTML:           body.RubricHTML,
		EvidenceRequirements: body.EvidenceRequirements,
		Mode:                 body.Mode,
		ScoringScale:         body.ScoringScale,
		MaxScore:             body.MaxScore,
		Weight:               body.Weight,
		DueAt:                dueAt,
		Status:               body.Status,
		CreatedByUsername:    currentUsername(r),
		AssessorUsername:     body.AssessorUsername,
		Checklist:            checklist,
	}, nil
}

func nonTestAssessmentListInputFromRequest(r *http.Request) (service.ListNonTestAssessmentsInput, error) {
	subjectID := pgtype.UUID{}
	if raw := strings.TrimSpace(r.URL.Query().Get("subject_id")); raw != "" {
		parsed, err := parseUUID(raw)
		if err != nil {
			return service.ListNonTestAssessmentsInput{}, fmt.Errorf("subject_id invalid")
		}
		subjectID = parsed
	}
	classID := pgtype.UUID{}
	if raw := strings.TrimSpace(r.URL.Query().Get("class_id")); raw != "" {
		parsed, err := parseUUID(raw)
		if err != nil {
			return service.ListNonTestAssessmentsInput{}, fmt.Errorf("class_id invalid")
		}
		classID = parsed
	}
	limit := int32(25)
	offset := int32(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		var parsed int
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		var parsed int
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}
	return service.ListNonTestAssessmentsInput{
		SubjectID:      subjectID,
		ClassID:        classID,
		Status:         r.URL.Query().Get("status"),
		AssessmentType: r.URL.Query().Get("assessment_type"),
		SyncFilter:     r.URL.Query().Get("sync_filter"),
		SearchQuery:    r.URL.Query().Get("q"),
		Limit:          limit,
		Offset:         offset,
	}, nil
}

func serializeNonTestAssessmentListRow(row db.ListNonTestAssessmentsRow) map[string]any {
	return map[string]any{
		"id":                            pgUUIDString(row.ID),
		"subject_id":                    pgUUIDString(row.SubjectID),
		"subject_name":                  row.SubjectName,
		"subject_code":                  row.SubjectCode,
		"class_id":                      pgUUIDString(row.ClassID),
		"class_name":                    row.ClassName,
		"class_level":                   row.ClassLevel,
		"assessment_type":               row.AssessmentType,
		"title":                         row.Title,
		"description":                   row.Description,
		"instruction_html":              row.InstructionHtml,
		"rubric_html":                   row.RubricHtml,
		"evidence_requirements":         row.EvidenceRequirements,
		"mode":                          row.Mode,
		"scoring_scale":                 row.ScoringScale,
		"max_score":                     numericToFloat(row.MaxScore),
		"weight":                        numericToFloat(row.Weight),
		"due_at":                        row.DueAt,
		"status":                        row.Status,
		"created_by_username":           row.CreatedByUsername,
		"assessor_username":             row.AssessorUsername,
		"checklist":                     decodeJSONBytes(row.Checklist),
		"grade_component_id":            pgUUIDString(row.GradeComponentID),
		"grade_synced_at":               row.GradeSyncedAt,
		"grade_synced_by":               row.GradeSyncedBy,
		"created_at":                    row.CreatedAt,
		"updated_at":                    row.UpdatedAt,
		"total_submissions":             row.TotalSubmissions,
		"reviewed_submissions":          row.ReviewedSubmissions,
		"last_reviewed_at":              row.LastReviewedAt,
		"unsynced_reviewed_submissions": row.UnsyncedReviewedSubmissions,
	}
}

func serializeNonTestAssessmentDetailRow(row db.GetNonTestAssessmentRow) map[string]any {
	item := map[string]any{
		"id":                            pgUUIDString(row.ID),
		"subject_id":                    pgUUIDString(row.SubjectID),
		"subject_name":                  row.SubjectName,
		"subject_code":                  row.SubjectCode,
		"class_id":                      pgUUIDString(row.ClassID),
		"class_name":                    row.ClassName,
		"class_level":                   row.ClassLevel,
		"assessment_type":               row.AssessmentType,
		"title":                         row.Title,
		"description":                   row.Description,
		"instruction_html":              row.InstructionHtml,
		"rubric_html":                   row.RubricHtml,
		"evidence_requirements":         row.EvidenceRequirements,
		"mode":                          row.Mode,
		"scoring_scale":                 row.ScoringScale,
		"max_score":                     numericToFloat(row.MaxScore),
		"weight":                        numericToFloat(row.Weight),
		"due_at":                        row.DueAt,
		"status":                        row.Status,
		"created_by_username":           row.CreatedByUsername,
		"assessor_username":             row.AssessorUsername,
		"checklist":                     decodeJSONBytes(row.Checklist),
		"grade_component_id":            pgUUIDString(row.GradeComponentID),
		"grade_synced_at":               row.GradeSyncedAt,
		"grade_synced_by":               row.GradeSyncedBy,
		"created_at":                    row.CreatedAt,
		"updated_at":                    row.UpdatedAt,
		"total_submissions":             row.TotalSubmissions,
		"reviewed_submissions":          row.ReviewedSubmissions,
		"last_reviewed_at":              row.LastReviewedAt,
		"unsynced_reviewed_submissions": row.UnsyncedReviewedSubmissions,
	}
	return item
}

func serializeNonTestAssessment(row db.NonTestAssessment) map[string]any {
	return map[string]any{
		"id":                    pgUUIDString(row.ID),
		"subject_id":            pgUUIDString(row.SubjectID),
		"class_id":              pgUUIDString(row.ClassID),
		"assessment_type":       row.AssessmentType,
		"title":                 row.Title,
		"description":           row.Description,
		"instruction_html":      row.InstructionHtml,
		"rubric_html":           row.RubricHtml,
		"evidence_requirements": row.EvidenceRequirements,
		"mode":                  row.Mode,
		"scoring_scale":         row.ScoringScale,
		"max_score":             numericToFloat(row.MaxScore),
		"weight":                numericToFloat(row.Weight),
		"due_at":                row.DueAt,
		"status":                row.Status,
		"created_by_username":   row.CreatedByUsername,
		"assessor_username":     row.AssessorUsername,
		"checklist":             decodeJSONBytes(row.Checklist),
		"grade_component_id":    pgUUIDString(row.GradeComponentID),
		"grade_synced_at":       row.GradeSyncedAt,
		"grade_synced_by":       row.GradeSyncedBy,
		"created_at":            row.CreatedAt,
		"updated_at":            row.UpdatedAt,
	}
}

func serializeNonTestSubmissionListRow(row db.ListNonTestSubmissionsRow) map[string]any {
	return map[string]any{
		"id":                 pgUUIDString(row.ID),
		"assessment_id":      pgUUIDString(row.AssessmentID),
		"student_id":         pgUUIDString(row.StudentID),
		"nis":                row.Nis,
		"nisn":               row.Nisn,
		"student_name":       row.StudentName,
		"class_id":           pgUUIDString(row.ClassID),
		"class_name":         row.ClassName,
		"class_level":        row.ClassLevel,
		"status":             row.Status,
		"evidence_url":       row.EvidenceUrl,
		"evidence_note":      row.EvidenceNote,
		"score":              numericToFloat(row.Score),
		"feedback":           row.Feedback,
		"submitted_at":       row.SubmittedAt,
		"graded_at":          row.GradedAt,
		"graded_by_username": row.GradedByUsername,
		"created_at":         row.CreatedAt,
		"updated_at":         row.UpdatedAt,
	}
}

func serializeNonTestSubmission(row db.NonTestAssessmentSubmission) map[string]any {
	return map[string]any{
		"id":                 pgUUIDString(row.ID),
		"assessment_id":      pgUUIDString(row.AssessmentID),
		"student_id":         pgUUIDString(row.StudentID),
		"status":             row.Status,
		"evidence_url":       row.EvidenceUrl,
		"evidence_note":      row.EvidenceNote,
		"score":              numericToFloat(row.Score),
		"feedback":           row.Feedback,
		"submitted_at":       row.SubmittedAt,
		"graded_at":          row.GradedAt,
		"graded_by_username": row.GradedByUsername,
		"created_at":         row.CreatedAt,
		"updated_at":         row.UpdatedAt,
	}
}

func parseOptionalTime(raw string) (pgtype.Timestamptz, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Timestamptz{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		var out pgtype.Timestamptz
		if scanErr := out.Scan(parsed); scanErr != nil {
			return pgtype.Timestamptz{}, scanErr
		}
		return out, nil
	}
	location, locErr := time.LoadLocation("Asia/Makassar")
	if locErr != nil {
		location = time.Local
	}
	formats := []string{"2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"}
	var err error
	for _, format := range formats {
		var parsed time.Time
		parsed, err = time.ParseInLocation(format, raw, location)
		if err != nil {
			continue
		}
		var out pgtype.Timestamptz
		if scanErr := out.Scan(parsed); scanErr != nil {
			return pgtype.Timestamptz{}, scanErr
		}
		return out, nil
	}
	return pgtype.Timestamptz{}, err
}

func numericToFloat(value pgtype.Numeric) any {
	if !value.Valid || value.Int == nil {
		return nil
	}
	ratio := new(big.Rat).SetInt(value.Int)
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absInt32(value.Exp))), nil)
	if value.Exp >= 0 {
		ratio.Mul(ratio, new(big.Rat).SetInt(scale))
	} else {
		ratio.Quo(ratio, new(big.Rat).SetInt(scale))
	}
	out, _ := ratio.Float64()
	return out
}

func absInt32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}
