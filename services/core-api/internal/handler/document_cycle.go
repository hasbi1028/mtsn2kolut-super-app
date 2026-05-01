package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type DocumentCycle struct{ svc *service.DocumentCycle }

func NewDocumentCycle(svc *service.DocumentCycle) *DocumentCycle {
	return &DocumentCycle{svc: svc}
}

func (h *DocumentCycle) Stats(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.Stats(r.Context(), int32Query(r.URL.Query().Get("period_year")))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *DocumentCycle) ListCatalogs(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.ListCatalogs(
		r.Context(),
		r.URL.Query().Get("search"),
		r.URL.Query().Get("frequency"),
		boolQuery(r.URL.Query().Get("active_only")),
	)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *DocumentCycle) CreateCatalog(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	params, ok := h.parseCatalogRequest(w, r, false)
	if !ok {
		return
	}
	row, err := h.svc.CreateCatalog(r.Context(), db.CreateDocumentCycleCatalogParams{
		Code:                         params.Code,
		Title:                        params.Title,
		Frequency:                    params.Frequency,
		DomainArea:                   params.DomainArea,
		ExternalSystem:               params.ExternalSystem,
		SnpStandard:                  params.SnpStandard,
		RegulationRef:                params.RegulationRef,
		DefaultOwnerUnitID:           params.DefaultOwnerUnitID,
		DefaultResponsibleEmployeeID: params.DefaultResponsibleEmployeeID,
		DefaultVerifierEmployeeID:    params.DefaultVerifierEmployeeID,
		DeadlineDaysAfterPeriod:      params.DeadlineDaysAfterPeriod,
		ReminderDaysBeforeDue:        params.ReminderDaysBeforeDue,
		Description:                  params.Description,
		IsActive:                     params.IsActive,
		SortOrder:                    params.SortOrder,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *DocumentCycle) UpdateCatalog(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	params, ok := h.parseCatalogRequest(w, r, true)
	if !ok {
		return
	}
	row, err := h.svc.UpdateCatalog(r.Context(), db.UpdateDocumentCycleCatalogParams{
		ID:                           id,
		Code:                         params.Code,
		Title:                        params.Title,
		Frequency:                    params.Frequency,
		DomainArea:                   params.DomainArea,
		ExternalSystem:               params.ExternalSystem,
		SnpStandard:                  params.SnpStandard,
		RegulationRef:                params.RegulationRef,
		DefaultOwnerUnitID:           params.DefaultOwnerUnitID,
		DefaultResponsibleEmployeeID: params.DefaultResponsibleEmployeeID,
		DefaultVerifierEmployeeID:    params.DefaultVerifierEmployeeID,
		DeadlineDaysAfterPeriod:      params.DeadlineDaysAfterPeriod,
		ReminderDaysBeforeDue:        params.ReminderDaysBeforeDue,
		Description:                  params.Description,
		IsActive:                     params.IsActive,
		SortOrder:                    params.SortOrder,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *DocumentCycle) DeleteCatalog(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteCatalog(r.Context(), id); err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.NoContent(w)
}

func (h *DocumentCycle) ListObligations(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(r.URL.Query().Get("owner_unit_id"))
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	responsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(r.URL.Query().Get("responsible_employee_id"))
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	data, err := h.svc.ListObligations(r.Context(), db.ListDocumentCycleObligationsParams{
		Search:                r.URL.Query().Get("search"),
		Status:                r.URL.Query().Get("status"),
		Frequency:             r.URL.Query().Get("frequency"),
		DomainArea:            r.URL.Query().Get("domain_area"),
		ExternalSystem:        r.URL.Query().Get("external_system"),
		PeriodYear:            int32Query(r.URL.Query().Get("period_year")),
		OwnerUnitID:           ownerUnitID,
		ResponsibleEmployeeID: responsibleEmployeeID,
		ReminderOnly:          boolQuery(r.URL.Query().Get("reminder_only")),
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *DocumentCycle) GenerateYear(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body documentCycleGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	data, err := h.svc.GenerateYear(r.Context(), inventoryActorUserID(r), body.PeriodYear)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, data)
}

func (h *DocumentCycle) UpdateObligation(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	params, ok := h.parseObligationRequest(w, r, id)
	if !ok {
		return
	}
	row, err := h.svc.UpdateObligation(r.Context(), inventoryActorUserID(r), params)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *DocumentCycle) ListEvents(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	data, err := h.svc.ListEvents(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *DocumentCycle) UpdateObligationStatus(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body documentCycleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.UpdateObligationStatus(r.Context(), inventoryActorUserID(r), id, body.Status, body.Notes)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *DocumentCycle) DeleteObligation(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteObligation(r.Context(), id); err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.NoContent(w)
}

func (h *DocumentCycle) parseCatalogRequest(w http.ResponseWriter, r *http.Request, update bool) (documentCycleCatalogParsedRequest, bool) {
	var body documentCycleCatalogRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return documentCycleCatalogParsedRequest{}, false
	}
	defaultOwnerUnitID, err := service.ParseGovernanceOptionalUUID(body.DefaultOwnerUnitID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return documentCycleCatalogParsedRequest{}, false
	}
	defaultResponsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(body.DefaultResponsibleEmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return documentCycleCatalogParsedRequest{}, false
	}
	defaultVerifierEmployeeID, err := service.ParseGovernanceOptionalUUID(body.DefaultVerifierEmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return documentCycleCatalogParsedRequest{}, false
	}
	deadlineDays := body.DeadlineDaysAfterPeriod
	reminderDays := body.ReminderDaysBeforeDue
	if !update {
		if deadlineDays == 0 {
			deadlineDays = 5
		}
		if reminderDays == 0 {
			reminderDays = 3
		}
	}
	return documentCycleCatalogParsedRequest{
		Code:                         body.Code,
		Title:                        body.Title,
		Frequency:                    body.Frequency,
		DomainArea:                   body.DomainArea,
		ExternalSystem:               body.ExternalSystem,
		SnpStandard:                  body.SnpStandard,
		RegulationRef:                body.RegulationRef,
		DefaultOwnerUnitID:           defaultOwnerUnitID,
		DefaultResponsibleEmployeeID: defaultResponsibleEmployeeID,
		DefaultVerifierEmployeeID:    defaultVerifierEmployeeID,
		DeadlineDaysAfterPeriod:      deadlineDays,
		ReminderDaysBeforeDue:        reminderDays,
		Description:                  body.Description,
		IsActive:                     boolDefault(body.IsActive, true),
		SortOrder:                    body.SortOrder,
	}, true
}

func (h *DocumentCycle) parseObligationRequest(w http.ResponseWriter, r *http.Request, id pgtype.UUID) (db.UpdateDocumentCycleObligationParams, bool) {
	var body documentCycleObligationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	dueDate, err := service.ParseGovernanceDate(body.DueDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	reminderDate, err := service.ParseGovernanceDate(body.ReminderDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(body.OwnerUnitID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	responsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(body.ResponsibleEmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	verifierEmployeeID, err := service.ParseGovernanceOptionalUUID(body.VerifierEmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	governanceDocumentID, err := service.ParseGovernanceOptionalUUID(body.GovernanceDocumentID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	workPlanItemID, err := service.ParseGovernanceOptionalUUID(body.WorkPlanItemID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	performanceTargetID, err := service.ParseGovernanceOptionalUUID(body.PerformanceTargetID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	evidenceItemID, err := service.ParseGovernanceOptionalUUID(body.EvidenceItemID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	complianceActionID, err := service.ParseGovernanceOptionalUUID(body.ComplianceActionID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	archiveDocumentID, err := service.ParseGovernanceOptionalUUID(body.ArchiveDocumentID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	return db.UpdateDocumentCycleObligationParams{
		ID:                    id,
		DueDate:               dueDate,
		ReminderDate:          reminderDate,
		DomainArea:            body.DomainArea,
		ExternalSystem:        body.ExternalSystem,
		OwnerUnitID:           ownerUnitID,
		ResponsibleEmployeeID: responsibleEmployeeID,
		VerifierEmployeeID:    verifierEmployeeID,
		GovernanceDocumentID:  governanceDocumentID,
		WorkPlanItemID:        workPlanItemID,
		PerformanceTargetID:   performanceTargetID,
		EvidenceItemID:        evidenceItemID,
		ComplianceActionID:    complianceActionID,
		ArchiveDocumentID:     archiveDocumentID,
		Notes:                 body.Notes,
		VerificationNotes:     body.VerificationNotes,
	}, true
}

type documentCycleCatalogRequest struct {
	Code                         string `json:"code"`
	Title                        string `json:"title"`
	Frequency                    string `json:"frequency"`
	DomainArea                   string `json:"domain_area"`
	ExternalSystem               string `json:"external_system"`
	SnpStandard                  string `json:"snp_standard"`
	RegulationRef                string `json:"regulation_ref"`
	DefaultOwnerUnitID           string `json:"default_owner_unit_id"`
	DefaultResponsibleEmployeeID string `json:"default_responsible_employee_id"`
	DefaultVerifierEmployeeID    string `json:"default_verifier_employee_id"`
	DeadlineDaysAfterPeriod      int32  `json:"deadline_days_after_period"`
	ReminderDaysBeforeDue        int32  `json:"reminder_days_before_due"`
	Description                  string `json:"description"`
	IsActive                     *bool  `json:"is_active"`
	SortOrder                    int32  `json:"sort_order"`
}

type documentCycleCatalogParsedRequest struct {
	Code                         string
	Title                        string
	Frequency                    string
	DomainArea                   string
	ExternalSystem               string
	SnpStandard                  string
	RegulationRef                string
	DefaultOwnerUnitID           pgtype.UUID
	DefaultResponsibleEmployeeID pgtype.UUID
	DefaultVerifierEmployeeID    pgtype.UUID
	DeadlineDaysAfterPeriod      int32
	ReminderDaysBeforeDue        int32
	Description                  string
	IsActive                     bool
	SortOrder                    int32
}

type documentCycleObligationRequest struct {
	DueDate               string `json:"due_date"`
	ReminderDate          string `json:"reminder_date"`
	DomainArea            string `json:"domain_area"`
	ExternalSystem        string `json:"external_system"`
	OwnerUnitID           string `json:"owner_unit_id"`
	ResponsibleEmployeeID string `json:"responsible_employee_id"`
	VerifierEmployeeID    string `json:"verifier_employee_id"`
	GovernanceDocumentID  string `json:"governance_document_id"`
	WorkPlanItemID        string `json:"work_plan_item_id"`
	PerformanceTargetID   string `json:"performance_target_id"`
	EvidenceItemID        string `json:"evidence_item_id"`
	ComplianceActionID    string `json:"compliance_action_id"`
	ArchiveDocumentID     string `json:"archive_document_id"`
	Notes                 string `json:"notes"`
	VerificationNotes     string `json:"verification_notes"`
}

type documentCycleStatusRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

type documentCycleGenerateRequest struct {
	PeriodYear int32 `json:"period_year"`
}
