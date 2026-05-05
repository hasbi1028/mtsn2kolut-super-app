package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

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
