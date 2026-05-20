package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type BankSoalReport struct {
	svc *service.BankSoalReportService
}

func NewBankSoalReport(svc *service.BankSoalReportService) *BankSoalReport {
	return &BankSoalReport{svc: svc}
}

func (h *BankSoalReport) Get(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	filters := bankSoalReportFiltersFromRequest(r)
	result, err := h.svc.Report(r.Context(), filters, cbtQuestionActorFromRequest(r))
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, result)
}

func (h *BankSoalReport) Export(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	filters := bankSoalReportFiltersFromRequest(r)
	format := strings.TrimSpace(r.URL.Query().Get("format"))
	if r.Method == http.MethodPost {
		var body struct {
			Format  string                        `json:"format"`
			Report  string                        `json:"report"`
			Filters service.BankSoalReportFilters `json:"filters"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if strings.TrimSpace(body.Format) != "" {
				format = body.Format
			}
			if strings.TrimSpace(body.Report) != "" {
				body.Filters.Report = body.Report
			}
			if strings.TrimSpace(body.Filters.Report) != "" {
				filters = body.Filters
			}
		}
	}
	export, err := h.svc.Export(r.Context(), filters, format, cbtQuestionActorFromRequest(r))
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", export.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+export.FileName+"\"")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(export.Data)
}

func bankSoalReportFiltersFromRequest(r *http.Request) service.BankSoalReportFilters {
	q := r.URL.Query()
	includeSystem := false
	switch strings.ToLower(strings.TrimSpace(q.Get("include_system"))) {
	case "1", "true", "yes", "ya":
		includeSystem = true
	}
	workflowStatuses := q["workflow_status"]
	if len(workflowStatuses) == 0 {
		workflowStatuses = q["workflow_status[]"]
	}
	return service.BankSoalReportFilters{
		Report:           q.Get("report"),
		PeriodPreset:     q.Get("period_preset"),
		StartDate:        q.Get("start_date"),
		EndDate:          q.Get("end_date"),
		EventID:          q.Get("event_id"),
		SubjectID:        q.Get("subject_id"),
		TargetLevel:      q.Get("target_level"),
		AuthorUsername:   q.Get("author_username"),
		WorkflowStatus:   q.Get("workflow_status"),
		WorkflowStatuses: workflowStatuses,
		IncludeSystem:    includeSystem,
		GroupBy:          q.Get("group_by"),
	}
}
