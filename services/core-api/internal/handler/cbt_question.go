package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
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
			return service.ListCbtQuestionsInput{}, fmt.Errorf("event_id invalid")
		}
		eventID = parsed
	}
	subjectID := pgtype.UUID{}
	if raw := strings.TrimSpace(r.URL.Query().Get("subject_id")); raw != "" {
		parsed, err := parseUUID(raw)
		if err != nil {
			return service.ListCbtQuestionsInput{}, fmt.Errorf("subject_id invalid")
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

func (h *CbtQuestion) ExportCSV(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	exportSvc, ok := h.svc.(cbtQuestionExportService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt question export service unavailable"))
		return
	}
	input, err := questionListInputFromRequest(r, 2000, 2000)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	if !adminAccessAllowed(r) {
		username := strings.TrimSpace(currentUsername(r))
		if username == "" {
			api.Forbidden(w)
			return
		}
		input.AuthorUsername = username
	}
	result, err := exportSvc.ExportCSV(r.Context(), input)
	if err != nil {
		api.Internal(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, result.Filename))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(result.Content)
}

func (h *CbtQuestion) TemplateCSV(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	templateSvc, ok := h.svc.(cbtQuestionTemplateService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt question template service unavailable"))
		return
	}
	result, err := templateSvc.TemplateCSV()
	if err != nil {
		api.Internal(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, result.Filename))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(result.Content)
}

func (h *CbtQuestion) Get(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.GetDetail(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		writeDomainOrInternal(w, err, "Detail soal CBT tidak dapat diakses")
		return
	}
	api.OK(w, serializeQuestionDetailRow(row, true))
}

func (h *CbtQuestion) Timeline(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.Timeline(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		writeDomainOrInternal(w, err, "Timeline soal CBT tidak dapat diakses")
		return
	}
	api.OK(w, rows)
}

func (h *CbtQuestion) Create(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	body, err := decodeQuestionBody(r)
	if err != nil {
		writeClientError(w, err, "Data soal CBT tidak valid")
		return
	}
	input, err := questionInputFromBody(r, body, pgtype.UUID{})
	if err != nil {
		writeClientError(w, err, "Data soal CBT tidak valid")
		return
	}
	row, err := h.svc.Create(r.Context(), input)
	if err != nil {
		writeClientError(w, err, "Pembuatan soal CBT tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_CREATE", "cbt_question", pgUUIDString(row.ID), map[string]any{
		"subject_id":      pgUUIDString(row.SubjectID),
		"question_type":   row.QuestionType,
		"workflow_status": row.WorkflowStatus,
		"author_username": row.AuthorUsername,
	})
	api.Created(w, serializeQuestionModel(row))
}

func (h *CbtQuestion) ImportLegacyCSV(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		api.BadRequest(w, "multipart import tidak valid")
		return
	}
	subjectID, err := parseUUID(r.FormValue("subject_id"))
	if err != nil {
		api.BadRequest(w, "subject_id invalid")
		return
	}
	eventID := pgtype.UUID{}
	if raw := strings.TrimSpace(r.FormValue("event_id")); raw != "" {
		parsed, err := parseUUID(raw)
		if err != nil {
			api.BadRequest(w, "event_id invalid")
			return
		}
		eventID = parsed
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "file CSV wajib diisi")
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, 5<<20))
	if err != nil {
		api.BadRequest(w, "file CSV tidak dapat dibaca")
		return
	}
	importSvc, ok := h.svc.(cbtQuestionImportService)
	if !ok {
		api.Internal(w, fmt.Errorf("cbt question import service unavailable"))
		return
	}
	result, err := importSvc.ImportLegacyCSV(r.Context(), service.ImportLegacyQuestionsInput{
		SubjectID: subjectID,
		EventID:   eventID,
		CSVText:   string(raw),
		Username:  currentUsername(r),
		Actor:     cbtQuestionActorFromRequest(r),
		DryRun:    parseBoolFormValue(r.FormValue("dry_run")),
	})
	if err != nil {
		writeClientError(w, err, "Import bank soal legacy tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_IMPORT_LEGACY", "cbt_question", "", map[string]any{
		"subject_id": pgUUIDString(subjectID),
		"event_id":   pgUUIDString(eventID),
		"imported":   result.Imported,
		"skipped":    result.Skipped,
	})
	api.OK(w, result)
}

func (h *CbtQuestion) Update(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if !h.requireQuestionAuthorOrAdmin(w, r, id) {
		return
	}
	current, err := h.svc.GetDetail(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		writeDomainOrInternal(w, err, "Detail soal CBT tidak dapat diakses")
		return
	}
	body, err := decodeQuestionBody(r)
	if err != nil {
		writeClientError(w, err, "Data soal CBT tidak valid")
		return
	}
	if !h.directQuestionUpdateAllowed(w, r, current, body) {
		return
	}
	input, err := questionInputFromBody(r, body, id)
	if err != nil {
		writeClientError(w, err, "Data soal CBT tidak valid")
		return
	}
	row, err := h.svc.Update(r.Context(), input)
	if err != nil {
		writeClientError(w, err, "Perubahan soal CBT tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_UPDATE", "cbt_question", pgUUIDString(row.ID), map[string]any{
		"subject_id":      pgUUIDString(row.SubjectID),
		"question_type":   row.QuestionType,
		"workflow_status": row.WorkflowStatus,
		"author_username": row.AuthorUsername,
	})
	api.OK(w, serializeQuestionModel(row))
}

func (h *CbtQuestion) Delete(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if !h.requireQuestionAuthorOrAdmin(w, r, id) {
		return
	}
	current, err := h.svc.GetDetail(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		api.Internal(w, err)
		return
	}
	if cbtQuestionUsageLocked(current.PackageCount, current.AnswerCount) {
		api.Conflict(w, "Soal sudah masuk paket ujian atau memiliki jawaban siswa. Duplikat soal untuk membuat revisi baru.")
		return
	}
	if err := h.svc.DeleteWithActor(r.Context(), id, cbtQuestionActorFromRequest(r)); err != nil {
		writeDomainOrInternal(w, err, "Hapus soal CBT tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_DELETE", "cbt_question", pgUUIDString(id), nil)
	api.NoContent(w)
}

func (h *CbtQuestion) BulkWorkflowAction(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		Action      string   `json:"action"`
		Notes       string   `json:"notes"`
		QuestionIDs []string `json:"question_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	ids := make([]pgtype.UUID, 0, len(body.QuestionIDs))
	for _, raw := range body.QuestionIDs {
		id, err := parseUUID(raw)
		if err != nil {
			api.BadRequest(w, "question_ids invalid")
			return
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 || len(ids) > 100 {
		api.BadRequest(w, "question_ids wajib berisi 1-100 id")
		return
	}
	action := strings.TrimSpace(body.Action)
	if action == "publish" && !hasAnyRole(r, "admin") {
		api.Forbidden(w)
		return
	}
	if (action == "approve" || action == "reject") && !hasAnyRole(r, "admin", "guru") {
		api.Forbidden(w)
		return
	}
	result, err := h.svc.BulkWorkflow(r.Context(), service.BulkCbtQuestionWorkflowInput{
		QuestionIDs: ids,
		Action:      action,
		Notes:       body.Notes,
		Actor:       cbtQuestionActorFromRequest(r),
	})
	if err != nil {
		writeClientError(w, err, "Aksi bulk workflow soal CBT tidak valid")
		return
	}
	api.OK(w, result)
}

func (h *CbtQuestion) WorkflowAction(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Action string `json:"action"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	actor := cbtQuestionActorFromRequest(r)
	switch strings.TrimSpace(body.Action) {
	case "submit_review":
		if !hasAnyRole(r, "admin", "guru") {
			api.Forbidden(w)
			return
		}
		if !h.requireQuestionAuthorOrAdmin(w, r, id) {
			return
		}
		row, err := h.svc.SubmitReview(r.Context(), id, actor, body.Notes)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_SUBMIT_REVIEW", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"workflow_status": row.WorkflowStatus,
			"review_notes":    body.Notes,
		})
		api.OK(w, serializeQuestionModel(row))
	case "approve":
		if !hasAnyRole(r, "admin", "guru") {
			api.Forbidden(w)
			return
		}
		if !actor.IsAdmin() && !actor.UserID.Valid {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.Approve(r.Context(), id, actor, body.Notes)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_APPROVE", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"workflow_status": row.WorkflowStatus,
			"review_notes":    body.Notes,
		})
		api.OK(w, serializeQuestionModel(row))
	case "reject":
		if !hasAnyRole(r, "admin", "guru") {
			api.Forbidden(w)
			return
		}
		if !actor.IsAdmin() && !actor.UserID.Valid {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.Reject(r.Context(), id, actor, body.Notes)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_REJECT", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"workflow_status": row.WorkflowStatus,
			"review_notes":    body.Notes,
		})
		api.OK(w, serializeQuestionModel(row))
	case "publish":
		if !hasAnyRole(r, "admin") {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.Publish(r.Context(), id, actor)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_PUBLISH", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"status": row.Status,
		})
		api.OK(w, serializeQuestionModel(row))
	case "archive":
		if !hasAnyRole(r, "admin") {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.Archive(r.Context(), id, actor)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_ARCHIVE", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"status": row.Status,
		})
		api.OK(w, serializeQuestionModel(row))
	default:
		api.BadRequest(w, "action tidak didukung")
	}
}

func (h *CbtQuestion) Duplicate(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.DuplicateAsDraft(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		writeClientError(w, err, "Duplikasi soal CBT tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_DUPLICATE", "cbt_question", pgUUIDString(row.ID), map[string]any{
		"source_question_id": pgUUIDString(id),
		"workflow_status":    row.WorkflowStatus,
		"author_username":    row.AuthorUsername,
	})
	api.Created(w, serializeQuestionModel(row))
}

func (h *CbtQuestion) MarkRevision(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Notes string `json:"notes"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
			api.BadRequest(w, "invalid json")
			return
		}
	}
	row, err := h.svc.DuplicateForRevision(r.Context(), id, cbtQuestionActorFromRequest(r), body.Notes)
	if err != nil {
		writeClientError(w, err, "Draft revisi soal CBT tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_MARK_REVISION", "cbt_question", pgUUIDString(row.ID), map[string]any{
		"source_question_id": pgUUIDString(id),
		"workflow_status":    row.WorkflowStatus,
		"review_notes":       body.Notes,
		"author_username":    row.AuthorUsername,
	})
	api.Created(w, serializeQuestionModel(row))
}

func decodeQuestionBody(r *http.Request) (cbtQuestionBody, error) {
	var body cbtQuestionBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return cbtQuestionBody{}, err
	}
	return body, nil
}

func questionInputFromBody(r *http.Request, body cbtQuestionBody, id pgtype.UUID) (service.SaveCbtQuestionInput, error) {
	var eventID pgtype.UUID
	if strings.TrimSpace(body.EventID) != "" {
		parsed, err := parseUUID(body.EventID)
		if err != nil {
			return service.SaveCbtQuestionInput{}, err
		}
		eventID = parsed
	}
	subjectID, err := parseUUID(body.SubjectID)
	if err != nil {
		return service.SaveCbtQuestionInput{}, err
	}

	gradeLevel := pgtype.Int2{}
	if body.GradeLevel > 0 {
		gradeLevel = pgtype.Int2{Int16: body.GradeLevel, Valid: true}
	}

	username := currentUsername(r)
	return service.SaveCbtQuestionInput{
		ID:               id,
		EventID:          eventID,
		SubjectID:        subjectID,
		AuthoringMode:    body.AuthoringMode,
		Code:             body.Code,
		QuestionText:     body.QuestionText,
		QuestionType:     body.QuestionType,
		Options:          body.Options,
		OptionA:          body.OptionA,
		OptionB:          body.OptionB,
		OptionC:          body.OptionC,
		OptionD:          body.OptionD,
		OptionE:          body.OptionE,
		AnswerKey:        strings.TrimSpace(body.AnswerKey),
		Explanation:      body.Explanation,
		Difficulty:       db.CbtQuestionDifficultyEnum(body.Difficulty),
		Status:           db.CbtQuestionStatusEnum(body.Status),
		StemHTML:         body.StemHTML,
		StemLatex:        body.StemLatex,
		StimulusHTML:     body.StimulusHTML,
		StimulusLatex:    body.StimulusLatex,
		ExplanationHTML:  body.ExplanationHTML,
		RubricHTML:       body.RubricHTML,
		AcademicPhase:    body.AcademicPhase,
		GradeLevel:       gradeLevel,
		CPRef:            body.CPRef,
		TPRef:            body.TPRef,
		KDRef:            body.KDRef,
		IndicatorRef:     body.IndicatorRef,
		MaterialTopic:    body.MaterialTopic,
		CognitiveLevel:   body.CognitiveLevel,
		HotsFlag:         body.HotsFlag,
		MediaAssetIDs:    body.MediaAssetIDs,
		WorkflowStatus:   body.WorkflowStatus,
		AuthorUsername:   username,
		ReviewerUsername: username,
		ApproverUsername: username,
		WriterNotes:      body.WriterNotes,
		ReviewNotes:      body.ReviewNotes,
		Actor:            cbtQuestionActorFromRequest(r),
	}, nil
}

func cbtQuestionActorFromRequest(r *http.Request) service.CbtQuestionActor {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return service.CbtQuestionActor{}
	}
	userID, _ := authUserID(claims)
	roles := []string{}
	if raw, ok := claims["roles"].([]any); ok {
		for _, value := range raw {
			if role, ok := value.(string); ok {
				roles = append(roles, role)
			}
		}
	}
	if role, _ := claims["role"].(string); role != "" {
		roles = append(roles, role)
	}
	return service.CbtQuestionActor{UserID: userID, Username: currentUsername(r), Roles: roles}
}

func currentUsername(r *http.Request) string {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return ""
	}
	if usr, _ := claims["usr"].(string); usr != "" {
		return usr
	}
	if sub, _ := claims["sub"].(string); sub != "" {
		return sub
	}
	return ""
}

func (h *CbtQuestion) requireQuestionAuthorOrAdmin(w http.ResponseWriter, r *http.Request, id pgtype.UUID) bool {
	if hasAnyRole(r, "admin") {
		return true
	}
	if !hasAnyRole(r, "guru") {
		api.Forbidden(w)
		return false
	}
	username := currentUsername(r)
	if strings.TrimSpace(username) == "" {
		api.Forbidden(w)
		return false
	}
	row, err := h.svc.GetDetail(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		writeDomainOrInternal(w, err, "Detail soal CBT tidak dapat diakses")
		return false
	}
	if strings.TrimSpace(row.AuthorUsername) != username {
		api.Forbidden(w)
		return false
	}
	return true
}

func (h *CbtQuestion) directQuestionUpdateAllowed(w http.ResponseWriter, r *http.Request, current db.GetCbtQuestionDetailRow, body cbtQuestionBody) bool {
	if cbtQuestionUsageLocked(current.PackageCount, current.AnswerCount) {
		api.Conflict(w, "Soal sudah masuk paket ujian atau memiliki jawaban siswa. Duplikat soal untuk membuat revisi baru.")
		return false
	}
	requestedWorkflow := strings.ToLower(strings.TrimSpace(body.WorkflowStatus))
	currentWorkflow := strings.ToLower(strings.TrimSpace(current.WorkflowStatus))
	if requestedWorkflow == "approved" && currentWorkflow != "approved" {
		api.Conflict(w, "Setujui soal melalui aksi workflow, bukan edit langsung.")
		return false
	}
	if db.CbtQuestionStatusEnum(strings.TrimSpace(body.Status)) == db.CbtQuestionStatusEnumPublished && current.Status != db.CbtQuestionStatusEnumPublished {
		api.Conflict(w, "Terbitkan soal melalui aksi workflow, bukan edit langsung.")
		return false
	}
	if !hasAnyRole(r, "admin") {
		if currentWorkflow != "" && currentWorkflow != "draft" && currentWorkflow != "rejected" {
			api.Conflict(w, "Soal sedang atau sudah masuk alur review. Duplikat soal untuk membuat revisi baru.")
			return false
		}
		if current.Status != "" && current.Status != db.CbtQuestionStatusEnumDraft {
			api.Conflict(w, "Soal tidak lagi berstatus draft. Duplikat soal untuk membuat revisi baru.")
			return false
		}
	}
	return true
}

func cbtQuestionUsageLocked(packageCount, answerCount int32) bool {
	return packageCount > 0 || answerCount > 0
}

func cbtQuestionUsageMap(packageCount, answerCount int32) map[string]any {
	return map[string]any{
		"package_count": packageCount,
		"answer_count":  answerCount,
		"is_locked":     cbtQuestionUsageLocked(packageCount, answerCount),
	}
}

func parseBoolFormValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func hasAnyRole(r *http.Request, allowed ...string) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, allowed...)
}

func cbtAuditAuthoringEvent(audit cbtAuthoringAuditWriter, ctx context.Context, action, entityType, entityID string, extra map[string]any) {
	if audit == nil {
		return
	}
	claims, ok := api.ClaimsFromContext(ctx)
	if !ok {
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		return
	}
	meta := map[string]any{
		"username":   claims["usr"],
		"user_id":    claims["uid"],
		"session_id": claims["ssid"],
	}
	for key, value := range extra {
		meta[key] = value
	}
	rawMeta, err := json.Marshal(meta)
	if err != nil {
		return
	}
	_, _ = audit.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   rawMeta,
	})
}

func serializeQuestionListRow(row db.ListCbtQuestionsFilteredRow, includeAnswerKey ...bool) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	return map[string]any{
		"id":                pgUUIDString(row.ID),
		"event_id":          pgUUIDString(row.EventID),
		"authoring_mode":    suggestedMode,
		"suggested_mode":    suggestedMode,
		"subject_id":        pgUUIDString(row.SubjectID),
		"subject_name":      row.SubjectName,
		"subject_code":      row.SubjectCode,
		"code":              row.Code,
		"question_text":     row.QuestionText,
		"question_type":     row.QuestionType,
		"options":           decodeJSONBytes(row.Options),
		"option_a":          row.OptionA,
		"option_b":          row.OptionB,
		"option_c":          row.OptionC,
		"option_d":          row.OptionD,
		"option_e":          row.OptionE,
		"answer_key":        row.AnswerKey,
		"explanation":       row.Explanation,
		"difficulty":        row.Difficulty,
		"status":            row.Status,
		"created_at":        row.CreatedAt,
		"updated_at":        row.UpdatedAt,
		"stem_html":         row.StemHtml,
		"stem_latex":        row.StemLatex,
		"stimulus_html":     row.StimulusHtml,
		"stimulus_latex":    row.StimulusLatex,
		"explanation_html":  row.ExplanationHtml,
		"rubric_html":       row.RubricHtml,
		"academic_phase":    row.AcademicPhase,
		"grade_level":       nullableInt(row.GradeLevel),
		"cp_ref":            row.CpRef,
		"tp_ref":            row.TpRef,
		"kd_ref":            row.KdRef,
		"indicator_ref":     row.IndicatorRef,
		"material_topic":    row.MaterialTopic,
		"cognitive_level":   row.CognitiveLevel,
		"hots_flag":         row.HotsFlag,
		"media_asset_ids":   decodeJSONBytes(row.MediaAssetIds),
		"workflow_status":   row.WorkflowStatus,
		"version":           row.Version,
		"author_username":   row.AuthorUsername,
		"reviewer_username": row.ReviewerUsername,
		"reviewed_at":       row.ReviewedAt,
		"approver_username": row.ApproverUsername,
		"approved_at":       row.ApprovedAt,
		"writer_notes":      row.WriterNotes,
		"review_notes":      row.ReviewNotes,
		"package_count":     row.PackageCount,
		"answer_count":      row.AnswerCount,
		"is_locked":         cbtQuestionUsageLocked(row.PackageCount, row.AnswerCount),
		"usage":             cbtQuestionUsageMap(row.PackageCount, row.AnswerCount),
	}
}

func questionAnswerKeyAllowed(r *http.Request, authorUsername string) bool {
	if adminAccessAllowed(r) {
		return true
	}
	return strings.TrimSpace(currentUsername(r)) != "" && strings.TrimSpace(currentUsername(r)) == strings.TrimSpace(authorUsername)
}

func serializeQuestionDetailRow(row db.GetCbtQuestionDetailRow, includeAnswerKey bool) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	answerKey := ""
	if includeAnswerKey {
		answerKey = row.AnswerKey
	}
	return map[string]any{
		"id":                pgUUIDString(row.ID),
		"event_id":          pgUUIDString(row.EventID),
		"authoring_mode":    suggestedMode,
		"suggested_mode":    suggestedMode,
		"subject_id":        pgUUIDString(row.SubjectID),
		"subject_name":      row.SubjectName,
		"subject_code":      row.SubjectCode,
		"code":              row.Code,
		"question_text":     row.QuestionText,
		"question_type":     row.QuestionType,
		"options":           decodeJSONBytes(row.Options),
		"option_a":          row.OptionA,
		"option_b":          row.OptionB,
		"option_c":          row.OptionC,
		"option_d":          row.OptionD,
		"option_e":          row.OptionE,
		"answer_key":        answerKey,
		"explanation":       row.Explanation,
		"difficulty":        row.Difficulty,
		"status":            row.Status,
		"created_at":        row.CreatedAt,
		"updated_at":        row.UpdatedAt,
		"stem_html":         row.StemHtml,
		"stem_latex":        row.StemLatex,
		"stimulus_html":     row.StimulusHtml,
		"stimulus_latex":    row.StimulusLatex,
		"explanation_html":  row.ExplanationHtml,
		"rubric_html":       row.RubricHtml,
		"academic_phase":    row.AcademicPhase,
		"grade_level":       nullableInt(row.GradeLevel),
		"cp_ref":            row.CpRef,
		"tp_ref":            row.TpRef,
		"kd_ref":            row.KdRef,
		"indicator_ref":     row.IndicatorRef,
		"material_topic":    row.MaterialTopic,
		"cognitive_level":   row.CognitiveLevel,
		"hots_flag":         row.HotsFlag,
		"media_asset_ids":   decodeJSONBytes(row.MediaAssetIds),
		"workflow_status":   row.WorkflowStatus,
		"version":           row.Version,
		"author_username":   row.AuthorUsername,
		"reviewer_username": row.ReviewerUsername,
		"reviewed_at":       row.ReviewedAt,
		"approver_username": row.ApproverUsername,
		"approved_at":       row.ApprovedAt,
		"writer_notes":      row.WriterNotes,
		"review_notes":      row.ReviewNotes,
		"package_count":     row.PackageCount,
		"answer_count":      row.AnswerCount,
		"is_locked":         cbtQuestionUsageLocked(row.PackageCount, row.AnswerCount),
		"usage":             cbtQuestionUsageMap(row.PackageCount, row.AnswerCount),
	}
}

func serializeQuestionModel(row db.CbtQuestion) map[string]any {
	suggestedMode := serviceAuthoringModeFromRow(row.QuestionType, row.StemLatex, row.StimulusLatex, row.AcademicPhase, row.CpRef, row.TpRef, row.KdRef, row.IndicatorRef, row.MaterialTopic, row.CognitiveLevel, row.HotsFlag, row.WorkflowStatus, row.WriterNotes, row.ReviewNotes, row.RubricHtml)
	return map[string]any{
		"id":                pgUUIDString(row.ID),
		"event_id":          pgUUIDString(row.EventID),
		"authoring_mode":    suggestedMode,
		"suggested_mode":    suggestedMode,
		"subject_id":        pgUUIDString(row.SubjectID),
		"code":              row.Code,
		"question_text":     row.QuestionText,
		"question_type":     row.QuestionType,
		"options":           decodeJSONBytes(row.Options),
		"option_a":          row.OptionA,
		"option_b":          row.OptionB,
		"option_c":          row.OptionC,
		"option_d":          row.OptionD,
		"option_e":          row.OptionE,
		"answer_key":        row.AnswerKey,
		"explanation":       row.Explanation,
		"difficulty":        row.Difficulty,
		"status":            row.Status,
		"created_at":        row.CreatedAt,
		"updated_at":        row.UpdatedAt,
		"stem_html":         row.StemHtml,
		"stem_latex":        row.StemLatex,
		"stimulus_html":     row.StimulusHtml,
		"stimulus_latex":    row.StimulusLatex,
		"explanation_html":  row.ExplanationHtml,
		"rubric_html":       row.RubricHtml,
		"academic_phase":    row.AcademicPhase,
		"grade_level":       nullableInt(row.GradeLevel),
		"cp_ref":            row.CpRef,
		"tp_ref":            row.TpRef,
		"kd_ref":            row.KdRef,
		"indicator_ref":     row.IndicatorRef,
		"material_topic":    row.MaterialTopic,
		"cognitive_level":   row.CognitiveLevel,
		"hots_flag":         row.HotsFlag,
		"media_asset_ids":   decodeJSONBytes(row.MediaAssetIds),
		"workflow_status":   row.WorkflowStatus,
		"version":           row.Version,
		"author_username":   row.AuthorUsername,
		"reviewer_username": row.ReviewerUsername,
		"reviewed_at":       row.ReviewedAt,
		"approver_username": row.ApproverUsername,
		"approved_at":       row.ApprovedAt,
		"writer_notes":      row.WriterNotes,
		"review_notes":      row.ReviewNotes,
	}
}

func decodeJSONBytes(raw []byte) any {
	if len(raw) == 0 {
		return []any{}
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return []any{}
	}
	return v
}

func nullableInt(value pgtype.Int2) any {
	if !value.Valid {
		return nil
	}
	return value.Int16
}

func serviceAuthoringModeFromRow(questionType, stemLatex, stimulusLatex, academicPhase, cpRef, tpRef, kdRef, indicatorRef, materialTopic, cognitiveLevel string, hotsFlag bool, workflowStatus, writerNotes, reviewNotes, rubricHTML string) string {
	if questionType != "multiple_choice" && questionType != "essay" {
		return "advance"
	}
	if strings.TrimSpace(stemLatex) != "" || strings.TrimSpace(stimulusLatex) != "" {
		return "advance"
	}
	if strings.TrimSpace(academicPhase) != "" || strings.TrimSpace(cpRef) != "" || strings.TrimSpace(tpRef) != "" ||
		strings.TrimSpace(kdRef) != "" || strings.TrimSpace(indicatorRef) != "" || strings.TrimSpace(materialTopic) != "" ||
		strings.TrimSpace(cognitiveLevel) != "" || hotsFlag {
		return "advance"
	}
	if strings.TrimSpace(workflowStatus) != "" && strings.TrimSpace(workflowStatus) != "draft" {
		return "advance"
	}
	if strings.TrimSpace(writerNotes) != "" || strings.TrimSpace(reviewNotes) != "" || strings.TrimSpace(rubricHTML) != "" {
		return "advance"
	}
	return "beginner"
}
