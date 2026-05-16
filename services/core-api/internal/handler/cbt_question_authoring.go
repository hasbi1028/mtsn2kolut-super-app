package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func (h *CbtQuestion) Get(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
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
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	rows, err := h.svc.Timeline(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		writeDomainOrInternal(w, err, "Timeline soal CBT tidak dapat diakses")
		return
	}
	api.OK(w, rows)
}

func (h *CbtQuestion) WorkflowEvents(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	rows, err := h.svc.WorkflowEvents(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		writeDomainOrInternal(w, err, "Riwayat workflow soal CBT tidak dapat diakses")
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, serializeQuestionWorkflowEventRow(row))
	}
	api.OK(w, map[string]any{"items": items})
}

func (h *CbtQuestion) Versions(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	rows, err := h.svc.Versions(r.Context(), id, cbtQuestionActorFromRequest(r))
	if err != nil {
		writeDomainOrInternal(w, err, "Riwayat versi soal CBT tidak dapat diakses")
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, serializeQuestionVersionRow(row))
	}
	api.OK(w, map[string]any{"items": items})
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

func (h *CbtQuestion) Update(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
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
		api.BadRequest(w, "ID data tidak valid")
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

	targetLevel := strings.TrimSpace(body.TargetLevel)
	if targetLevel == "" && body.GradeLevel > 0 {
		targetLevel = legacyQuestionTargetLevelFromInt(body.GradeLevel)
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
		TargetLevel:      targetLevel,
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

func legacyQuestionTargetLevelFromInt(value int16) string {
	switch value {
	case 7:
		return "VII"
	case 8:
		return "VIII"
	case 9:
		return "IX"
	default:
		return ""
	}
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
	if raw, ok := claims["roles"].([]string); ok {
		roles = append(roles, raw...)
	}
	if role, _ := claims["role"].(string); role != "" {
		roles = append(roles, role)
	}
	permissions := []string{}
	if raw, ok := claims["permissions"].([]any); ok {
		for _, value := range raw {
			if permission, ok := value.(string); ok {
				permissions = append(permissions, permission)
			}
		}
	}
	if raw, ok := claims["permissions"].([]string); ok {
		permissions = append(permissions, raw...)
	}
	if permission, _ := claims["permission"].(string); permission != "" {
		permissions = append(permissions, permission)
	}
	return service.CbtQuestionActor{UserID: userID, Username: currentUsername(r), Roles: roles, Permissions: permissions}
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
