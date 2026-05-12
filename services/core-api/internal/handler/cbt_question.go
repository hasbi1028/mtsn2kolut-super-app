package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtQuestion struct {
	svc   cbtQuestionService
	audit cbtAuthoringAuditWriter
}

type cbtAuthoringAuditWriter interface {
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

type cbtQuestionService interface {
	ListFiltered(ctx context.Context, in service.ListCbtQuestionsInput) ([]db.ListCbtQuestionsFilteredRow, int64, error)
	Summary(ctx context.Context, actor service.CbtQuestionActor) (service.CbtQuestionSummary, error)
	GetDetail(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor) (db.GetCbtQuestionDetailRow, error)
	Create(ctx context.Context, input service.SaveCbtQuestionInput) (db.CbtQuestion, error)
	Update(ctx context.Context, input service.SaveCbtQuestionInput) (db.CbtQuestion, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	DeleteWithActor(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor) error
	Timeline(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor) ([]db.CbtQuestionAuditLog, error)
	BulkWorkflow(ctx context.Context, in service.BulkCbtQuestionWorkflowInput) (service.BulkCbtQuestionWorkflowResult, error)
	SubmitReview(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error)
	Approve(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error)
	Reject(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error)
	Publish(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor) (db.CbtQuestion, error)
	Archive(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor) (db.CbtQuestion, error)
	DuplicateAsDraft(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor) (db.CbtQuestion, error)
	DuplicateForRevision(ctx context.Context, id pgtype.UUID, actor service.CbtQuestionActor, reviewNotes string) (db.CbtQuestion, error)
}

type cbtQuestionImportService interface {
	ImportLegacyCSV(ctx context.Context, input service.ImportLegacyQuestionsInput) (service.ImportLegacyQuestionsResult, error)
}

type cbtQuestionExportService interface {
	ExportCSV(ctx context.Context, input service.ListCbtQuestionsInput) (service.ExportCbtQuestionsCSVResult, error)
}

type cbtQuestionTemplateService interface {
	TemplateCSV() (service.ExportCbtQuestionsCSVResult, error)
}

func NewCbtQuestion(svc *service.CbtQuestion, audit ...cbtAuthoringAuditWriter) *CbtQuestion {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &CbtQuestion{svc: svc, audit: writer}
}

type cbtQuestionBody struct {
	SubjectID       string                   `json:"subject_id"`
	EventID         string                   `json:"event_id"`
	AuthoringMode   string                   `json:"authoring_mode"`
	Code            string                   `json:"code"`
	QuestionText    string                   `json:"question_text"`
	QuestionType    string                   `json:"question_type"`
	Options         []service.QuestionOption `json:"options"`
	OptionA         string                   `json:"option_a"`
	OptionB         string                   `json:"option_b"`
	OptionC         string                   `json:"option_c"`
	OptionD         string                   `json:"option_d"`
	OptionE         string                   `json:"option_e"`
	AnswerKey       string                   `json:"answer_key"`
	Explanation     string                   `json:"explanation"`
	Difficulty      string                   `json:"difficulty"`
	Status          string                   `json:"status"`
	StemHTML        string                   `json:"stem_html"`
	StemLatex       string                   `json:"stem_latex"`
	StimulusHTML    string                   `json:"stimulus_html"`
	StimulusLatex   string                   `json:"stimulus_latex"`
	ExplanationHTML string                   `json:"explanation_html"`
	RubricHTML      string                   `json:"rubric_html"`
	AcademicPhase   string                   `json:"academic_phase"`
	GradeLevel      int16                    `json:"grade_level"`
	CPRef           string                   `json:"cp_ref"`
	TPRef           string                   `json:"tp_ref"`
	KDRef           string                   `json:"kd_ref"`
	IndicatorRef    string                   `json:"indicator_ref"`
	MaterialTopic   string                   `json:"material_topic"`
	CognitiveLevel  string                   `json:"cognitive_level"`
	HotsFlag        bool                     `json:"hots_flag"`
	MediaAssetIDs   []string                 `json:"media_asset_ids"`
	WorkflowStatus  string                   `json:"workflow_status"`
	WriterNotes     string                   `json:"writer_notes"`
	ReviewNotes     string                   `json:"review_notes"`
}

func (h *CbtQuestion) Summary(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	summary, err := h.svc.Summary(r.Context(), cbtQuestionActorFromRequest(r))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, serializeQuestionSummary(summary))
}

func (h *CbtQuestion) List(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	input, err := questionListInputFromRequest(r, 25, 100)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	rows, total, err := h.svc.ListFiltered(r.Context(), input)
	if err != nil {
		api.Internal(w, err)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, serializeQuestionListRow(row))
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

func questionListInputFromRequest(r *http.Request, defaultLimit int32, maxLimit int32) (service.ListCbtQuestionsInput, error) {
	eventID := pgtype.UUID{}
	if raw := strings.TrimSpace(r.URL.Query().Get("event_id")); raw != "" {
		parsed, err := parseUUID(raw)
		if err != nil {
			return service.ListCbtQuestionsInput{}, fmt.Errorf("Kegiatan asesmen tidak valid")
		}
		eventID = parsed
	}
	subjectID := pgtype.UUID{}
	if raw := strings.TrimSpace(r.URL.Query().Get("subject_id")); raw != "" {
		parsed, err := parseUUID(raw)
		if err != nil {
			return service.ListCbtQuestionsInput{}, fmt.Errorf("Mata pelajaran tidak valid")
		}
		subjectID = parsed
	}
	limit := defaultLimit
	offset := int32(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		var parsed int
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err == nil && parsed > 0 && parsed <= int(maxLimit) {
			limit = int32(parsed)
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		var parsed int
		if _, err := fmt.Sscanf(raw, "%d", &parsed); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}
	return service.ListCbtQuestionsInput{
		EventID:        eventID,
		QuestionScope:  r.URL.Query().Get("scope"),
		SubjectID:      subjectID,
		WorkflowStatus: r.URL.Query().Get("workflow_status"),
		Status:         r.URL.Query().Get("status"),
		QuestionType:   r.URL.Query().Get("question_type"),
		HotsFilter:     r.URL.Query().Get("hots"),
		RevisionSource: r.URL.Query().Get("revision_source"),
		SearchQuery:    r.URL.Query().Get("q"),
		Limit:          limit,
		Offset:         offset,
		Actor:          cbtQuestionActorFromRequest(r),
	}, nil
}
