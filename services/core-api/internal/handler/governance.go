package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Governance struct{ svc *service.Governance }

func NewGovernance(svc *service.Governance) *Governance { return &Governance{svc: svc} }

func governanceAccessAllowed(r *http.Request) bool {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if rawRoles, ok := claims["roles"].([]any); ok {
			for _, role := range rawRoles {
				if role == "admin" || role == "staf" {
					return true
				}
			}
		}
		if role, _ := claims["role"].(string); role == "admin" || role == "staf" {
			return true
		}
		return false
	}
	return true
}

func (h *Governance) Stats(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.Stats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Governance) SNPMatrix(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.SNPMatrix(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Governance) EmployeeOptions(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.EmployeeOptions(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Governance) ListUnits(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.ListUnits(r.Context(), r.URL.Query().Get("search"))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreateUnit(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governanceUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	parentID, err := service.ParseGovernanceOptionalUUID(body.ParentID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	row, err := h.svc.CreateUnit(r.Context(), db.CreateGovernanceUnitParams{
		Code:        body.Code,
		Name:        body.Name,
		UnitType:    body.UnitType,
		ParentID:    parentID,
		Description: body.Description,
		IsActive:    boolDefault(body.IsActive, true),
		SortOrder:   body.SortOrder,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governanceUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	parentID, err := service.ParseGovernanceOptionalUUID(body.ParentID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	row, err := h.svc.UpdateUnit(r.Context(), db.UpdateGovernanceUnitParams{
		ID:          id,
		Code:        body.Code,
		Name:        body.Name,
		UnitType:    body.UnitType,
		ParentID:    parentID,
		Description: body.Description,
		IsActive:    boolDefault(body.IsActive, true),
		SortOrder:   body.SortOrder,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeleteUnit(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeleteUnit)
}

func (h *Governance) ListPositions(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.ListPositions(r.Context(), r.URL.Query().Get("search"))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreatePosition(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governancePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	unitID, parentPositionID, ok := parsePositionIDs(w, body)
	if !ok {
		return
	}
	row, err := h.svc.CreatePosition(r.Context(), db.CreateGovernancePositionParams{
		UnitID:           unitID,
		Title:            body.Title,
		PositionType:     body.PositionType,
		ParentPositionID: parentPositionID,
		Description:      body.Description,
		Tupoksi:          body.Tupoksi,
		IsActive:         boolDefault(body.IsActive, true),
		SortOrder:        body.SortOrder,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governancePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	unitID, parentPositionID, ok := parsePositionIDs(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdatePosition(r.Context(), db.UpdateGovernancePositionParams{
		ID:               id,
		UnitID:           unitID,
		Title:            body.Title,
		PositionType:     body.PositionType,
		ParentPositionID: parentPositionID,
		Description:      body.Description,
		Tupoksi:          body.Tupoksi,
		IsActive:         boolDefault(body.IsActive, true),
		SortOrder:        body.SortOrder,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeletePosition(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeletePosition)
}

func (h *Governance) ListAssignments(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	activeOnly := boolQuery(r.URL.Query().Get("active_only"))
	data, err := h.svc.ListAssignments(r.Context(), r.URL.Query().Get("search"), activeOnly)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governanceAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseAssignmentRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.CreateAssignment(r.Context(), arg)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdateAssignment(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governanceAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	createArg, ok := parseAssignmentRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateAssignment(r.Context(), db.UpdateGovernanceAssignmentParams{
		ID:                     id,
		PositionID:             createArg.PositionID,
		EmployeeID:             createArg.EmployeeID,
		StartDate:              createArg.StartDate,
		EndDate:                createArg.EndDate,
		DecreeOutgoingLetterID: createArg.DecreeOutgoingLetterID,
		Notes:                  createArg.Notes,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeleteAssignment)
}

func (h *Governance) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	data, err := h.svc.ListDocuments(
		r.Context(),
		q.Get("search"),
		q.Get("doc_type"),
		q.Get("snp_standard"),
		int32Query(q.Get("period_year")),
	)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreateDocument(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governanceDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseDocumentRequest(w, body)
	if !ok {
		return
	}
	arg.CreatedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreateDocument(r.Context(), arg)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governanceDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	createArg, ok := parseDocumentRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateDocument(r.Context(), db.UpdateGovernanceDocumentParams{
		ID:               id,
		DocType:          createArg.DocType,
		Title:            createArg.Title,
		PeriodYear:       createArg.PeriodYear,
		PeriodLabel:      createArg.PeriodLabel,
		OwnerUnitID:      createArg.OwnerUnitID,
		SnpStandard:      createArg.SnpStandard,
		Status:           createArg.Status,
		DocumentUrl:      createArg.DocumentUrl,
		OutgoingLetterID: createArg.OutgoingLetterID,
		Summary:          createArg.Summary,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeleteDocument)
}

func (h *Governance) ListPrograms(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	data, err := h.svc.ListPrograms(
		r.Context(),
		q.Get("search"),
		q.Get("status"),
		q.Get("snp_standard"),
		int32Query(q.Get("period_year")),
	)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreateProgram(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governanceProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseProgramRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.CreateProgram(r.Context(), arg)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdateProgram(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governanceProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	createArg, ok := parseProgramRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateProgram(r.Context(), db.UpdateGovernanceProgramParams{
		ID:                    id,
		PeriodYear:            createArg.PeriodYear,
		Code:                  createArg.Code,
		Name:                  createArg.Name,
		SourceDocumentID:      createArg.SourceDocumentID,
		OwnerUnitID:           createArg.OwnerUnitID,
		ResponsiblePositionID: createArg.ResponsiblePositionID,
		ResponsibleEmployeeID: createArg.ResponsibleEmployeeID,
		SnpStandard:           createArg.SnpStandard,
		IkuCode:               createArg.IkuCode,
		Indicator:             createArg.Indicator,
		TargetValue:           createArg.TargetValue,
		TargetUnit:            createArg.TargetUnit,
		Status:                createArg.Status,
		ProgressPercent:       createArg.ProgressPercent,
		RealizationSummary:    createArg.RealizationSummary,
		EvidenceUrl:           createArg.EvidenceUrl,
		DueDate:               createArg.DueDate,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeleteProgram(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeleteProgram)
}

func (h *Governance) ListWorkPlanItems(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	data, err := h.svc.ListWorkPlanItems(
		r.Context(),
		q.Get("search"),
		q.Get("program_id"),
		q.Get("status"),
		int32Query(q.Get("period_year")),
	)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreateWorkPlanItem(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governanceWorkPlanItemRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseWorkPlanItemRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.CreateWorkPlanItem(r.Context(), arg)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdateWorkPlanItem(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governanceWorkPlanItemRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	createArg, ok := parseWorkPlanItemRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateWorkPlanItem(r.Context(), db.UpdateGovernanceWorkPlanItemParams{
		ID:                    id,
		PeriodYear:            createArg.PeriodYear,
		ProgramID:             createArg.ProgramID,
		SourceDocumentID:      createArg.SourceDocumentID,
		OwnerUnitID:           createArg.OwnerUnitID,
		ResponsibleEmployeeID: createArg.ResponsibleEmployeeID,
		EvidenceItemID:        createArg.EvidenceItemID,
		ActivityCode:          createArg.ActivityCode,
		ActivityName:          createArg.ActivityName,
		OutputIndicator:       createArg.OutputIndicator,
		TargetVolume:          createArg.TargetVolume,
		TargetUnit:            createArg.TargetUnit,
		BudgetSource:          createArg.BudgetSource,
		BudgetAmount:          createArg.BudgetAmount,
		RealizationAmount:     createArg.RealizationAmount,
		Status:                createArg.Status,
		ProgressPercent:       createArg.ProgressPercent,
		StartDate:             createArg.StartDate,
		EndDate:               createArg.EndDate,
		EvidenceUrl:           createArg.EvidenceUrl,
		Notes:                 createArg.Notes,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeleteWorkPlanItem(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeleteWorkPlanItem)
}

func (h *Governance) ListPerformanceTargets(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	data, err := h.svc.ListPerformanceTargets(
		r.Context(),
		q.Get("search"),
		q.Get("status"),
		q.Get("employee_id"),
		int32Query(q.Get("period_year")),
	)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreatePerformanceTarget(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governancePerformanceTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parsePerformanceTargetRequest(w, body)
	if !ok {
		return
	}
	arg.CreatedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreatePerformanceTarget(r.Context(), arg)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdatePerformanceTarget(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governancePerformanceTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	createArg, ok := parsePerformanceTargetRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdatePerformanceTarget(r.Context(), db.UpdateGovernancePerformanceTargetParams{
		ID:              id,
		PeriodYear:      createArg.PeriodYear,
		EmployeeID:      createArg.EmployeeID,
		PositionID:      createArg.PositionID,
		ProgramID:       createArg.ProgramID,
		ParentTargetID:  createArg.ParentTargetID,
		Aspect:          createArg.Aspect,
		Title:           createArg.Title,
		Indicator:       createArg.Indicator,
		TargetValue:     createArg.TargetValue,
		TargetUnit:      createArg.TargetUnit,
		Status:          createArg.Status,
		ProgressPercent: createArg.ProgressPercent,
		EvidenceUrl:     createArg.EvidenceUrl,
		ReviewNotes:     createArg.ReviewNotes,
		DueDate:         createArg.DueDate,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeletePerformanceTarget(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeletePerformanceTarget)
}

func (h *Governance) ListEvidenceItems(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	data, err := h.svc.ListEvidenceItems(
		r.Context(),
		q.Get("search"),
		q.Get("status"),
		q.Get("snp_standard"),
		int32Query(q.Get("period_year")),
	)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreateEvidenceItem(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governanceEvidenceItemRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseEvidenceItemRequest(w, body)
	if !ok {
		return
	}
	arg.CreatedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreateEvidenceItem(r.Context(), arg)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdateEvidenceItem(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governanceEvidenceItemRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	createArg, ok := parseEvidenceItemRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateEvidenceItem(r.Context(), db.UpdateGovernanceEvidenceItemParams{
		ID:                  id,
		PeriodYear:          createArg.PeriodYear,
		Title:               createArg.Title,
		EvidenceType:        createArg.EvidenceType,
		SnpStandard:         createArg.SnpStandard,
		OwnerUnitID:         createArg.OwnerUnitID,
		DocumentID:          createArg.DocumentID,
		ProgramID:           createArg.ProgramID,
		PerformanceTargetID: createArg.PerformanceTargetID,
		SourceModule:        createArg.SourceModule,
		EvidenceUrl:         createArg.EvidenceUrl,
		Status:              createArg.Status,
		Notes:               createArg.Notes,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeleteEvidenceItem(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeleteEvidenceItem)
}

func (h *Governance) ListComplianceActions(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	data, err := h.svc.ListComplianceActions(
		r.Context(),
		q.Get("search"),
		q.Get("status"),
		q.Get("priority"),
		q.Get("source_type"),
		q.Get("snp_standard"),
		q.Get("responsible_employee_id"),
		int32Query(q.Get("period_year")),
	)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, data)
}

func (h *Governance) CreateComplianceAction(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body governanceComplianceActionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseComplianceActionRequest(w, body)
	if !ok {
		return
	}
	arg.CreatedByUserID = inventoryActorUserID(r)
	row, err := h.svc.CreateComplianceAction(r.Context(), arg)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *Governance) UpdateComplianceAction(w http.ResponseWriter, r *http.Request) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body governanceComplianceActionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	createArg, ok := parseComplianceActionRequest(w, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateComplianceAction(r.Context(), db.UpdateGovernanceComplianceActionParams{
		ID:                    id,
		PeriodYear:            createArg.PeriodYear,
		SourceType:            createArg.SourceType,
		SourceRefID:           createArg.SourceRefID,
		SnpStandard:           createArg.SnpStandard,
		ProgramID:             createArg.ProgramID,
		DocumentID:            createArg.DocumentID,
		PerformanceTargetID:   createArg.PerformanceTargetID,
		EvidenceItemID:        createArg.EvidenceItemID,
		OwnerUnitID:           createArg.OwnerUnitID,
		ResponsibleEmployeeID: createArg.ResponsibleEmployeeID,
		Title:                 createArg.Title,
		Description:           createArg.Description,
		Priority:              createArg.Priority,
		Status:                createArg.Status,
		DueDate:               createArg.DueDate,
		CompletedAt:           createArg.CompletedAt,
		FollowUpNotes:         createArg.FollowUpNotes,
		EvidenceUrl:           createArg.EvidenceUrl,
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func (h *Governance) DeleteComplianceAction(w http.ResponseWriter, r *http.Request) {
	h.deleteByID(w, r, h.svc.DeleteComplianceAction)
}

func (h *Governance) deleteByID(w http.ResponseWriter, r *http.Request, deleteFn func(context.Context, pgtype.UUID) error) {
	if !governanceAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := deleteFn(r.Context(), id); err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.NoContent(w)
}

type governanceUnitRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	UnitType    string `json:"unit_type"`
	ParentID    string `json:"parent_id"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int32  `json:"sort_order"`
}

type governancePositionRequest struct {
	UnitID           string `json:"unit_id"`
	Title            string `json:"title"`
	PositionType     string `json:"position_type"`
	ParentPositionID string `json:"parent_position_id"`
	Description      string `json:"description"`
	Tupoksi          string `json:"tupoksi"`
	IsActive         *bool  `json:"is_active"`
	SortOrder        int32  `json:"sort_order"`
}

type governanceAssignmentRequest struct {
	PositionID             string `json:"position_id"`
	EmployeeID             string `json:"employee_id"`
	StartDate              string `json:"start_date"`
	EndDate                string `json:"end_date"`
	DecreeOutgoingLetterID string `json:"decree_outgoing_letter_id"`
	Notes                  string `json:"notes"`
}

type governanceDocumentRequest struct {
	DocType          string `json:"doc_type"`
	Title            string `json:"title"`
	PeriodYear       int32  `json:"period_year"`
	PeriodLabel      string `json:"period_label"`
	OwnerUnitID      string `json:"owner_unit_id"`
	SnpStandard      string `json:"snp_standard"`
	Status           string `json:"status"`
	DocumentUrl      string `json:"document_url"`
	OutgoingLetterID string `json:"outgoing_letter_id"`
	Summary          string `json:"summary"`
}

type governanceProgramRequest struct {
	PeriodYear            int32  `json:"period_year"`
	Code                  string `json:"code"`
	Name                  string `json:"name"`
	SourceDocumentID      string `json:"source_document_id"`
	OwnerUnitID           string `json:"owner_unit_id"`
	ResponsiblePositionID string `json:"responsible_position_id"`
	ResponsibleEmployeeID string `json:"responsible_employee_id"`
	SnpStandard           string `json:"snp_standard"`
	IkuCode               string `json:"iku_code"`
	Indicator             string `json:"indicator"`
	TargetValue           string `json:"target_value"`
	TargetUnit            string `json:"target_unit"`
	Status                string `json:"status"`
	ProgressPercent       int32  `json:"progress_percent"`
	RealizationSummary    string `json:"realization_summary"`
	EvidenceUrl           string `json:"evidence_url"`
	DueDate               string `json:"due_date"`
}

type governanceWorkPlanItemRequest struct {
	PeriodYear            int32  `json:"period_year"`
	ProgramID             string `json:"program_id"`
	SourceDocumentID      string `json:"source_document_id"`
	OwnerUnitID           string `json:"owner_unit_id"`
	ResponsibleEmployeeID string `json:"responsible_employee_id"`
	EvidenceItemID        string `json:"evidence_item_id"`
	ActivityCode          string `json:"activity_code"`
	ActivityName          string `json:"activity_name"`
	OutputIndicator       string `json:"output_indicator"`
	TargetVolume          string `json:"target_volume"`
	TargetUnit            string `json:"target_unit"`
	BudgetSource          string `json:"budget_source"`
	BudgetAmount          int64  `json:"budget_amount"`
	RealizationAmount     int64  `json:"realization_amount"`
	Status                string `json:"status"`
	ProgressPercent       int32  `json:"progress_percent"`
	StartDate             string `json:"start_date"`
	EndDate               string `json:"end_date"`
	EvidenceUrl           string `json:"evidence_url"`
	Notes                 string `json:"notes"`
}

type governancePerformanceTargetRequest struct {
	PeriodYear      int32  `json:"period_year"`
	EmployeeID      string `json:"employee_id"`
	PositionID      string `json:"position_id"`
	ProgramID       string `json:"program_id"`
	ParentTargetID  string `json:"parent_target_id"`
	Aspect          string `json:"aspect"`
	Title           string `json:"title"`
	Indicator       string `json:"indicator"`
	TargetValue     string `json:"target_value"`
	TargetUnit      string `json:"target_unit"`
	Status          string `json:"status"`
	ProgressPercent int32  `json:"progress_percent"`
	EvidenceUrl     string `json:"evidence_url"`
	ReviewNotes     string `json:"review_notes"`
	DueDate         string `json:"due_date"`
}

type governanceEvidenceItemRequest struct {
	PeriodYear          int32  `json:"period_year"`
	Title               string `json:"title"`
	EvidenceType        string `json:"evidence_type"`
	SnpStandard         string `json:"snp_standard"`
	OwnerUnitID         string `json:"owner_unit_id"`
	DocumentID          string `json:"document_id"`
	ProgramID           string `json:"program_id"`
	PerformanceTargetID string `json:"performance_target_id"`
	SourceModule        string `json:"source_module"`
	EvidenceUrl         string `json:"evidence_url"`
	Status              string `json:"status"`
	Notes               string `json:"notes"`
}

type governanceComplianceActionRequest struct {
	PeriodYear            int32  `json:"period_year"`
	SourceType            string `json:"source_type"`
	SourceRefID           string `json:"source_ref_id"`
	SnpStandard           string `json:"snp_standard"`
	ProgramID             string `json:"program_id"`
	DocumentID            string `json:"document_id"`
	PerformanceTargetID   string `json:"performance_target_id"`
	EvidenceItemID        string `json:"evidence_item_id"`
	OwnerUnitID           string `json:"owner_unit_id"`
	ResponsibleEmployeeID string `json:"responsible_employee_id"`
	Title                 string `json:"title"`
	Description           string `json:"description"`
	Priority              string `json:"priority"`
	Status                string `json:"status"`
	DueDate               string `json:"due_date"`
	CompletedAt           string `json:"completed_at"`
	FollowUpNotes         string `json:"follow_up_notes"`
	EvidenceUrl           string `json:"evidence_url"`
}

func parsePositionIDs(w http.ResponseWriter, body governancePositionRequest) (unitID, parentPositionID pgtype.UUID, ok bool) {
	parsedUnitID, err := service.ParseGovernanceOptionalUUID(body.UnitID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return unitID, parentPositionID, false
	}
	parsedParentID, err := service.ParseGovernanceOptionalUUID(body.ParentPositionID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return unitID, parentPositionID, false
	}
	return parsedUnitID, parsedParentID, true
}

func parseAssignmentRequest(w http.ResponseWriter, body governanceAssignmentRequest) (db.CreateGovernanceAssignmentParams, bool) {
	positionID, err := service.ParseGovernanceOptionalUUID(body.PositionID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceAssignmentParams{}, false
	}
	employeeID, err := service.ParseGovernanceOptionalUUID(body.EmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceAssignmentParams{}, false
	}
	startDate, err := service.ParseGovernanceDate(body.StartDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceAssignmentParams{}, false
	}
	endDate, err := service.ParseGovernanceOptionalDate(body.EndDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceAssignmentParams{}, false
	}
	decreeID, err := service.ParseGovernanceOptionalUUID(body.DecreeOutgoingLetterID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceAssignmentParams{}, false
	}
	return db.CreateGovernanceAssignmentParams{
		PositionID:             positionID,
		EmployeeID:             employeeID,
		StartDate:              startDate,
		EndDate:                endDate,
		DecreeOutgoingLetterID: decreeID,
		Notes:                  body.Notes,
	}, true
}

func parseDocumentRequest(w http.ResponseWriter, body governanceDocumentRequest) (db.CreateGovernanceDocumentParams, bool) {
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(body.OwnerUnitID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceDocumentParams{}, false
	}
	outgoingLetterID, err := service.ParseGovernanceOptionalUUID(body.OutgoingLetterID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceDocumentParams{}, false
	}
	return db.CreateGovernanceDocumentParams{
		DocType:          body.DocType,
		Title:            body.Title,
		PeriodYear:       body.PeriodYear,
		PeriodLabel:      body.PeriodLabel,
		OwnerUnitID:      ownerUnitID,
		SnpStandard:      body.SnpStandard,
		Status:           body.Status,
		DocumentUrl:      body.DocumentUrl,
		OutgoingLetterID: outgoingLetterID,
		Summary:          body.Summary,
	}, true
}

func parseProgramRequest(w http.ResponseWriter, body governanceProgramRequest) (db.CreateGovernanceProgramParams, bool) {
	sourceDocumentID, err := service.ParseGovernanceOptionalUUID(body.SourceDocumentID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceProgramParams{}, false
	}
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(body.OwnerUnitID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceProgramParams{}, false
	}
	responsiblePositionID, err := service.ParseGovernanceOptionalUUID(body.ResponsiblePositionID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceProgramParams{}, false
	}
	responsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(body.ResponsibleEmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceProgramParams{}, false
	}
	dueDate, err := service.ParseGovernanceOptionalDate(body.DueDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceProgramParams{}, false
	}
	return db.CreateGovernanceProgramParams{
		PeriodYear:            body.PeriodYear,
		Code:                  body.Code,
		Name:                  body.Name,
		SourceDocumentID:      sourceDocumentID,
		OwnerUnitID:           ownerUnitID,
		ResponsiblePositionID: responsiblePositionID,
		ResponsibleEmployeeID: responsibleEmployeeID,
		SnpStandard:           body.SnpStandard,
		IkuCode:               body.IkuCode,
		Indicator:             body.Indicator,
		TargetValue:           body.TargetValue,
		TargetUnit:            body.TargetUnit,
		Status:                body.Status,
		ProgressPercent:       body.ProgressPercent,
		RealizationSummary:    body.RealizationSummary,
		EvidenceUrl:           body.EvidenceUrl,
		DueDate:               dueDate,
	}, true
}

func parseWorkPlanItemRequest(w http.ResponseWriter, body governanceWorkPlanItemRequest) (db.CreateGovernanceWorkPlanItemParams, bool) {
	programID, err := service.ParseGovernanceOptionalUUID(body.ProgramID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceWorkPlanItemParams{}, false
	}
	sourceDocumentID, err := service.ParseGovernanceOptionalUUID(body.SourceDocumentID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceWorkPlanItemParams{}, false
	}
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(body.OwnerUnitID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceWorkPlanItemParams{}, false
	}
	responsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(body.ResponsibleEmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceWorkPlanItemParams{}, false
	}
	evidenceItemID, err := service.ParseGovernanceOptionalUUID(body.EvidenceItemID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceWorkPlanItemParams{}, false
	}
	startDate, err := service.ParseGovernanceOptionalDate(body.StartDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceWorkPlanItemParams{}, false
	}
	endDate, err := service.ParseGovernanceOptionalDate(body.EndDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceWorkPlanItemParams{}, false
	}
	return db.CreateGovernanceWorkPlanItemParams{
		PeriodYear:            body.PeriodYear,
		ProgramID:             programID,
		SourceDocumentID:      sourceDocumentID,
		OwnerUnitID:           ownerUnitID,
		ResponsibleEmployeeID: responsibleEmployeeID,
		EvidenceItemID:        evidenceItemID,
		ActivityCode:          body.ActivityCode,
		ActivityName:          body.ActivityName,
		OutputIndicator:       body.OutputIndicator,
		TargetVolume:          body.TargetVolume,
		TargetUnit:            body.TargetUnit,
		BudgetSource:          body.BudgetSource,
		BudgetAmount:          body.BudgetAmount,
		RealizationAmount:     body.RealizationAmount,
		Status:                body.Status,
		ProgressPercent:       body.ProgressPercent,
		StartDate:             startDate,
		EndDate:               endDate,
		EvidenceUrl:           body.EvidenceUrl,
		Notes:                 body.Notes,
	}, true
}

func parsePerformanceTargetRequest(w http.ResponseWriter, body governancePerformanceTargetRequest) (db.CreateGovernancePerformanceTargetParams, bool) {
	employeeID, err := service.ParseGovernanceOptionalUUID(body.EmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernancePerformanceTargetParams{}, false
	}
	positionID, err := service.ParseGovernanceOptionalUUID(body.PositionID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernancePerformanceTargetParams{}, false
	}
	programID, err := service.ParseGovernanceOptionalUUID(body.ProgramID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernancePerformanceTargetParams{}, false
	}
	parentTargetID, err := service.ParseGovernanceOptionalUUID(body.ParentTargetID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernancePerformanceTargetParams{}, false
	}
	dueDate, err := service.ParseGovernanceOptionalDate(body.DueDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernancePerformanceTargetParams{}, false
	}
	return db.CreateGovernancePerformanceTargetParams{
		PeriodYear:      body.PeriodYear,
		EmployeeID:      employeeID,
		PositionID:      positionID,
		ProgramID:       programID,
		ParentTargetID:  parentTargetID,
		Aspect:          body.Aspect,
		Title:           body.Title,
		Indicator:       body.Indicator,
		TargetValue:     body.TargetValue,
		TargetUnit:      body.TargetUnit,
		Status:          body.Status,
		ProgressPercent: body.ProgressPercent,
		EvidenceUrl:     body.EvidenceUrl,
		ReviewNotes:     body.ReviewNotes,
		DueDate:         dueDate,
	}, true
}

func parseEvidenceItemRequest(w http.ResponseWriter, body governanceEvidenceItemRequest) (db.CreateGovernanceEvidenceItemParams, bool) {
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(body.OwnerUnitID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceEvidenceItemParams{}, false
	}
	documentID, err := service.ParseGovernanceOptionalUUID(body.DocumentID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceEvidenceItemParams{}, false
	}
	programID, err := service.ParseGovernanceOptionalUUID(body.ProgramID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceEvidenceItemParams{}, false
	}
	performanceTargetID, err := service.ParseGovernanceOptionalUUID(body.PerformanceTargetID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceEvidenceItemParams{}, false
	}
	return db.CreateGovernanceEvidenceItemParams{
		PeriodYear:          body.PeriodYear,
		Title:               body.Title,
		EvidenceType:        body.EvidenceType,
		SnpStandard:         body.SnpStandard,
		OwnerUnitID:         ownerUnitID,
		DocumentID:          documentID,
		ProgramID:           programID,
		PerformanceTargetID: performanceTargetID,
		SourceModule:        body.SourceModule,
		EvidenceUrl:         body.EvidenceUrl,
		Status:              body.Status,
		Notes:               body.Notes,
	}, true
}

func parseComplianceActionRequest(w http.ResponseWriter, body governanceComplianceActionRequest) (db.CreateGovernanceComplianceActionParams, bool) {
	sourceRefID, err := service.ParseGovernanceOptionalUUID(body.SourceRefID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	programID, err := service.ParseGovernanceOptionalUUID(body.ProgramID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	documentID, err := service.ParseGovernanceOptionalUUID(body.DocumentID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	performanceTargetID, err := service.ParseGovernanceOptionalUUID(body.PerformanceTargetID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	evidenceItemID, err := service.ParseGovernanceOptionalUUID(body.EvidenceItemID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	ownerUnitID, err := service.ParseGovernanceOptionalUUID(body.OwnerUnitID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	responsibleEmployeeID, err := service.ParseGovernanceOptionalUUID(body.ResponsibleEmployeeID)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	dueDate, err := service.ParseGovernanceOptionalDate(body.DueDate)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	completedAt, err := service.ParseGovernanceOptionalTimestamp(body.CompletedAt)
	if err != nil {
		api.BadRequest(w, err.Error())
		return db.CreateGovernanceComplianceActionParams{}, false
	}
	return db.CreateGovernanceComplianceActionParams{
		PeriodYear:            body.PeriodYear,
		SourceType:            body.SourceType,
		SourceRefID:           sourceRefID,
		SnpStandard:           body.SnpStandard,
		ProgramID:             programID,
		DocumentID:            documentID,
		PerformanceTargetID:   performanceTargetID,
		EvidenceItemID:        evidenceItemID,
		OwnerUnitID:           ownerUnitID,
		ResponsibleEmployeeID: responsibleEmployeeID,
		Title:                 body.Title,
		Description:           body.Description,
		Priority:              body.Priority,
		Status:                body.Status,
		DueDate:               dueDate,
		CompletedAt:           completedAt,
		FollowUpNotes:         body.FollowUpNotes,
		EvidenceUrl:           body.EvidenceUrl,
	}, true
}

func boolDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func boolQuery(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func int32Query(value string) int32 {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return int32(parsed)
}
