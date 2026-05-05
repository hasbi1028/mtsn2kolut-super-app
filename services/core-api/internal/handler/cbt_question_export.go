package handler

import (
	"fmt"
	"net/http"
	"strings"

	"mtsn2kolut-super-app/backend/internal/api"
)

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
