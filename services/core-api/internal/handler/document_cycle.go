package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type documentCycleService interface {
	Stats(ctx context.Context, periodYear int32) (db.GetDocumentCycleStatsRow, error)
	ListCatalogs(ctx context.Context, search, frequency string, activeOnly bool) ([]db.ListDocumentCycleCatalogsRow, error)
	CreateCatalog(ctx context.Context, arg db.CreateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error)
	UpdateCatalog(ctx context.Context, arg db.UpdateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error)
	DeleteCatalog(ctx context.Context, id pgtype.UUID) error
	ListObligations(ctx context.Context, arg db.ListDocumentCycleObligationsParams) ([]db.ListDocumentCycleObligationsRow, error)
	ListVerificationQueue(ctx context.Context, verifierEmployeeID pgtype.UUID, periodYear int32) ([]db.ListDocumentCycleObligationsRow, error)
	GenerateYear(ctx context.Context, actorUserID pgtype.UUID, periodYear int32) (service.DocumentCycleGenerateResult, error)
	UpdateObligation(ctx context.Context, actorUserID pgtype.UUID, arg db.UpdateDocumentCycleObligationParams) (db.DocumentCycleObligation, error)
	ListEvents(ctx context.Context, obligationID pgtype.UUID, eventType, actor string) ([]db.ListDocumentCycleEventsByObligationRow, error)
	UpdateObligationStatus(ctx context.Context, actorUserID pgtype.UUID, id pgtype.UUID, status, notes string) (db.DocumentCycleObligation, error)
	DeleteObligation(ctx context.Context, id pgtype.UUID) error
}

type DocumentCycle struct {
	svc   documentCycleService
	audit cbtAuthoringAuditWriter
}

func NewDocumentCycle(svc *service.DocumentCycle, audit ...cbtAuthoringAuditWriter) *DocumentCycle {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &DocumentCycle{svc: svc, audit: writer}
}

func documentCycleHasRole(r *http.Request, want string) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(jwt.MapClaims(claims), want)
}

func documentCycleEmployeeID(r *http.Request) (pgtype.UUID, bool) {
	var id pgtype.UUID
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if mw.HasAnyRole(jwt.MapClaims(claims), "admin") {
			return id, false
		}
		if raw, _ := claims["eid"].(string); raw != "" {
			if err := id.Scan(raw); err == nil && id.Valid {
				return id, true
			}
		}
	}
	return id, false
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
		writeClientError(w, err, "Data katalog siklus dokumen tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "DOCUMENT_CYCLE_CATALOG_CREATE", "document_cycle_catalog", pgUUIDString(row.ID), map[string]any{
		"code":            row.Code,
		"title":           row.Title,
		"frequency":       row.Frequency,
		"domain_area":     row.DomainArea,
		"external_system": row.ExternalSystem,
	})
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
		writeClientError(w, err, "Perubahan katalog siklus dokumen tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "DOCUMENT_CYCLE_CATALOG_UPDATE", "document_cycle_catalog", pgUUIDString(row.ID), map[string]any{
		"code":            row.Code,
		"title":           row.Title,
		"frequency":       row.Frequency,
		"domain_area":     row.DomainArea,
		"external_system": row.ExternalSystem,
	})
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
		writeClientError(w, err, "Penghapusan katalog siklus dokumen tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "DOCUMENT_CYCLE_CATALOG_DELETE", "document_cycle_catalog", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
	api.NoContent(w)
}

func (h *DocumentCycle) ListObligations(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(r.URL.Query().Get("owner_unit_id"))
	if err != nil {
		writeClientError(w, err, "Filter siklus dokumen tidak valid")
		return
	}
	responsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(r.URL.Query().Get("responsible_employee_id"))
	if err != nil {
		writeClientError(w, err, "Filter siklus dokumen tidak valid")
		return
	}
	verifierEmployeeID, err := service.ParseGovernanceOptionalUUID(r.URL.Query().Get("verifier_employee_id"))
	if err != nil {
		writeClientError(w, err, "Filter siklus dokumen tidak valid")
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
		VerifierEmployeeID:    verifierEmployeeID,
		ReminderOnly:          boolQuery(r.URL.Query().Get("reminder_only")),
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *DocumentCycle) ListVerificationQueue(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}

	verifierEmployeeID, hasEmployeeID := documentCycleEmployeeID(r)
	if boolQuery(r.URL.Query().Get("all")) {
		if !documentCycleHasRole(r, "admin") {
			api.Forbidden(w)
			return
		}
		verifierEmployeeID = pgtype.UUID{}
		hasEmployeeID = true
	}
	if requestedVerifier := r.URL.Query().Get("verifier_employee_id"); requestedVerifier != "" {
		if !documentCycleHasRole(r, "admin") {
			api.Forbidden(w)
			return
		}
		parsed, err := service.ParseGovernanceOptionalUUID(requestedVerifier)
		if err != nil {
			writeClientError(w, err, "Filter antrian verifikasi tidak valid")
			return
		}
		verifierEmployeeID = parsed
		hasEmployeeID = true
	}
	if !hasEmployeeID {
		if !documentCycleHasRole(r, "admin") {
			api.Forbidden(w)
			return
		}
		verifierEmployeeID = pgtype.UUID{}
	}

	data, err := h.svc.ListVerificationQueue(r.Context(), verifierEmployeeID, int32Query(r.URL.Query().Get("period_year")))
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
		writeClientError(w, err, "Pembuatan siklus dokumen tahunan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "DOCUMENT_CYCLE_YEAR_GENERATE", "document_cycle_generation", strconv.FormatInt(int64(data.PeriodYear), 10), map[string]any{
		"period_year": data.PeriodYear,
		"generated":   data.Generated,
		"inserted":    data.Inserted,
	})
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
		writeClientError(w, err, "Perubahan kewajiban siklus dokumen tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "DOCUMENT_CYCLE_OBLIGATION_UPDATE", "document_cycle_obligation", pgUUIDString(row.ID), map[string]any{
		"period_year":     row.PeriodYear,
		"status":          row.Status,
		"domain_area":     row.DomainArea,
		"external_system": row.ExternalSystem,
	})
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
	data, err := h.svc.ListEvents(r.Context(), id, r.URL.Query().Get("event_type"), r.URL.Query().Get("actor"))
	if err != nil {
		writeClientError(w, err, "Filter audit siklus dokumen tidak valid")
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
		writeClientError(w, err, "Perubahan status kewajiban tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "DOCUMENT_CYCLE_OBLIGATION_STATUS_UPDATE", "document_cycle_obligation", pgUUIDString(row.ID), map[string]any{
		"period_year": row.PeriodYear,
		"status":      row.Status,
		"notes":       body.Notes,
	})
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
		writeClientError(w, err, "Penghapusan kewajiban siklus dokumen tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "DOCUMENT_CYCLE_OBLIGATION_DELETE", "document_cycle_obligation", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
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
		writeClientError(w, err, "Data katalog siklus dokumen tidak valid")
		return documentCycleCatalogParsedRequest{}, false
	}
	defaultResponsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(body.DefaultResponsibleEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data katalog siklus dokumen tidak valid")
		return documentCycleCatalogParsedRequest{}, false
	}
	defaultVerifierEmployeeID, err := service.ParseGovernanceOptionalUUID(body.DefaultVerifierEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data katalog siklus dokumen tidak valid")
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
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	reminderDate, err := service.ParseGovernanceDate(body.ReminderDate)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(body.OwnerUnitID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	responsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(body.ResponsibleEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	verifierEmployeeID, err := service.ParseGovernanceOptionalUUID(body.VerifierEmployeeID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	governanceDocumentID, err := service.ParseGovernanceOptionalUUID(body.GovernanceDocumentID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	workPlanItemID, err := service.ParseGovernanceOptionalUUID(body.WorkPlanItemID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	performanceTargetID, err := service.ParseGovernanceOptionalUUID(body.PerformanceTargetID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	evidenceItemID, err := service.ParseGovernanceOptionalUUID(body.EvidenceItemID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	complianceActionID, err := service.ParseGovernanceOptionalUUID(body.ComplianceActionID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
		return db.UpdateDocumentCycleObligationParams{}, false
	}
	archiveDocumentID, err := service.ParseGovernanceOptionalUUID(body.ArchiveDocumentID)
	if err != nil {
		writeClientError(w, err, "Data kewajiban siklus dokumen tidak valid")
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
