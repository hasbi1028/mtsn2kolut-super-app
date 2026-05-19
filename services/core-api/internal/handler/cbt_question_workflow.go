package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

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
		api.BadRequest(w, "Data yang dikirim tidak valid")
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
	normalizedAction := strings.TrimSpace(strings.ToLower(action))
	if normalizedAction == "submit_review" {
		normalizedAction = "submit_for_review"
	}
	if normalizedAction == "return_revision" {
		normalizedAction = "request_revision"
	}
	if normalizedAction == "submit_for_review" && !hasAnyPermission(r, "bank_soal.submit", "bank_soal.update", "bank_soal.update_own") && !hasAnyRole(r, "admin", "guru") {
		api.Forbidden(w)
		return
	}
	if normalizedAction == "publish" && !hasAnyPermission(r, "bank_soal.publish") && !hasAnyRole(r, "admin") {
		api.Forbidden(w)
		return
	}
	if normalizedAction == "archive" && !hasAnyPermission(r, "bank_soal.approve", "bank_soal.publish") && !hasAnyRole(r, "admin") {
		api.Forbidden(w)
		return
	}
	if normalizedAction == "restore_archive" && !hasAnyPermission(r, "bank_soal.approve", "bank_soal.publish") && !hasAnyRole(r, "admin") {
		api.Forbidden(w)
		return
	}
	if normalizedAction == "approve" && !hasAnyPermission(r, "bank_soal.approve", "bank_soal.publish") && !hasAnyRole(r, "admin") {
		api.Forbidden(w)
		return
	}
	if normalizedAction == "request_revision" && !hasAnyPermission(r, "bank_soal.review", "bank_soal.approve", "bank_soal.publish") && !hasAnyRole(r, "admin") {
		api.Forbidden(w)
		return
	}
	if (normalizedAction == "mark_reviewed" || normalizedAction == "reject") && !hasAnyPermission(r, "bank_soal.review") && !hasAnyRole(r, "admin", "guru") {
		api.Forbidden(w)
		return
	}
	result, err := h.svc.BulkWorkflow(r.Context(), service.BulkCbtQuestionWorkflowInput{
		QuestionIDs: ids,
		Action:      normalizedAction,
		Notes:       body.Notes,
		Actor:       cbtQuestionActorFromRequest(r),
	})
	if err != nil {
		writeClientError(w, err, "Aksi massal alur verifikasi soal tidak valid")
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
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		Action string `json:"action"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	actor := cbtQuestionActorFromRequest(r)
	switch strings.TrimSpace(body.Action) {
	case "submit_review", "submit_for_review":
		if !hasAnyPermission(r, "bank_soal.submit", "bank_soal.update", "bank_soal.update_own") && !hasAnyRole(r, "admin", "guru") {
			api.Forbidden(w)
			return
		}
		if !h.requireQuestionAuthorOrAdmin(w, r, id) {
			return
		}
		row, err := h.svc.SubmitForReview(r.Context(), id, actor, body.Notes)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_SUBMIT_REVIEW", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"workflow_status": row.WorkflowStatus,
			"review_notes":    body.Notes,
		})
		api.OK(w, serializeQuestionModel(row))
	case "request_revision", "return_revision":
		if !hasAnyPermission(r, "bank_soal.review", "bank_soal.approve", "bank_soal.publish") && !hasAnyRole(r, "admin") {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.RequestRevision(r.Context(), id, actor, body.Notes)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_REQUEST_REVISION", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"workflow_status": row.WorkflowStatus,
			"review_notes":    body.Notes,
		})
		api.OK(w, serializeQuestionModel(row))
	case "mark_reviewed":
		if !hasAnyPermission(r, "bank_soal.review") && !hasAnyRole(r, "admin", "guru") {
			api.Forbidden(w)
			return
		}
		if !actor.IsAdmin() && !actor.UserID.Valid {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.MarkReviewed(r.Context(), id, actor, body.Notes)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_MARK_REVIEWED", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"workflow_status": row.WorkflowStatus,
			"review_notes":    body.Notes,
		})
		api.OK(w, serializeQuestionModel(row))
	case "approve":
		if !hasAnyPermission(r, "bank_soal.approve", "bank_soal.publish") && !hasAnyRole(r, "admin") {
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
		if !hasAnyPermission(r, "bank_soal.review") && !hasAnyRole(r, "admin", "guru") {
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
		if !hasAnyPermission(r, "bank_soal.publish") && !hasAnyRole(r, "admin") {
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
		if !hasAnyPermission(r, "bank_soal.approve", "bank_soal.publish") && !hasAnyRole(r, "admin") {
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
	case "restore_archive":
		if !hasAnyPermission(r, "bank_soal.approve", "bank_soal.publish") && !hasAnyRole(r, "admin") {
			api.Forbidden(w)
			return
		}
		row, err := h.svc.RestoreArchive(r.Context(), id, actor, body.Notes)
		if err != nil {
			writeClientError(w, err, "Aksi workflow soal CBT tidak valid")
			return
		}
		cbtAuditAuthoringEvent(h.audit, r.Context(), "CBT_QUESTION_RESTORE_ARCHIVE", "cbt_question", pgUUIDString(row.ID), map[string]any{
			"status":          row.Status,
			"workflow_status": row.WorkflowStatus,
			"review_notes":    body.Notes,
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
		api.BadRequest(w, "ID data tidak valid")
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
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		Notes string `json:"notes"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
			api.BadRequest(w, "Data yang dikirim tidak valid")
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
