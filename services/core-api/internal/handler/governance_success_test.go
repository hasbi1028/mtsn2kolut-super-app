package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeGovernanceService struct {
	*service.Governance

	statsErr error
	listErr  error

	unitSearch       string
	positionSearch   string
	assignmentSearch string
	assignmentActive bool

	documentSearch string
	documentType   string
	documentSNP    string
	documentYear   int32

	programSearch string
	programStatus string
	programSNP    string
	programYear   int32

	workPlanSearch    string
	workPlanProgramID string
	workPlanStatus    string
	workPlanYear      int32

	targetSearch     string
	targetStatus     string
	targetEmployeeID string
	targetYear       int32

	evidenceSearch string
	evidenceStatus string
	evidenceSNP    string
	evidenceYear   int32

	complianceSearch     string
	complianceStatus     string
	compliancePriority   string
	complianceSourceType string
	complianceSNP        string
	complianceEmployeeID string
	complianceYear       int32

	createUnitArg       db.CreateGovernanceUnitParams
	updateUnitArg       db.UpdateGovernanceUnitParams
	createPositionArg   db.CreateGovernancePositionParams
	updatePositionArg   db.UpdateGovernancePositionParams
	createAssignmentArg db.CreateGovernanceAssignmentParams
	updateAssignmentArg db.UpdateGovernanceAssignmentParams
	createDocumentArg   db.CreateGovernanceDocumentParams
	updateDocumentArg   db.UpdateGovernanceDocumentParams
	createProgramArg    db.CreateGovernanceProgramParams
	updateProgramArg    db.UpdateGovernanceProgramParams
	createWorkPlanArg   db.CreateGovernanceWorkPlanItemParams
	updateWorkPlanArg   db.UpdateGovernanceWorkPlanItemParams
	createTargetArg     db.CreateGovernancePerformanceTargetParams
	updateTargetArg     db.UpdateGovernancePerformanceTargetParams
	createEvidenceArg   db.CreateGovernanceEvidenceItemParams
	updateEvidenceArg   db.UpdateGovernanceEvidenceItemParams
	createComplianceArg db.CreateGovernanceComplianceActionParams
	updateComplianceArg db.UpdateGovernanceComplianceActionParams

	deleteIDs   []pgtype.UUID
	mutationErr error
}

func (f *fakeGovernanceService) Stats(context.Context) (db.GetGovernanceStatsRow, error) {
	if f.statsErr != nil {
		return db.GetGovernanceStatsRow{}, f.statsErr
	}
	return db.GetGovernanceStatsRow{TotalUnits: 2, ActiveAssignments: 3}, nil
}

func (f *fakeGovernanceService) SNPMatrix(context.Context) ([]db.ListGovernanceSNPMatrixRow, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceSNPMatrixRow{{Code: "standar_isi", Name: "Standar Isi"}}, nil
}

func (f *fakeGovernanceService) EmployeeOptions(context.Context) ([]db.ListGovernanceEmployeeOptionsRow, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceEmployeeOptionsRow{{Nama: "Pegawai"}}, nil
}

func (f *fakeGovernanceService) ListUnits(_ context.Context, search string) ([]db.ListGovernanceUnitsRow, error) {
	f.unitSearch = search
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceUnitsRow{{Name: "Kurikulum"}}, nil
}

func (f *fakeGovernanceService) ListPositions(_ context.Context, search string) ([]db.ListGovernancePositionsRow, error) {
	f.positionSearch = search
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernancePositionsRow{{Title: "Waka Kurikulum"}}, nil
}

func (f *fakeGovernanceService) ListAssignments(_ context.Context, search string, activeOnly bool) ([]db.ListGovernanceAssignmentsRow, error) {
	f.assignmentSearch = search
	f.assignmentActive = activeOnly
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceAssignmentsRow{{EmployeeName: "Guru"}}, nil
}

func (f *fakeGovernanceService) ListDocuments(_ context.Context, search, docType, snpStandard string, periodYear int32) ([]db.ListGovernanceDocumentsRow, error) {
	f.documentSearch = search
	f.documentType = docType
	f.documentSNP = snpStandard
	f.documentYear = periodYear
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceDocumentsRow{{Title: "RKJM"}}, nil
}

func (f *fakeGovernanceService) ListPrograms(_ context.Context, search, status, snpStandard string, periodYear int32) ([]db.ListGovernanceProgramsRow, error) {
	f.programSearch = search
	f.programStatus = status
	f.programSNP = snpStandard
	f.programYear = periodYear
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceProgramsRow{{Name: "Program Mutu"}}, nil
}

func (f *fakeGovernanceService) ListWorkPlanItems(_ context.Context, search, programID, status string, periodYear int32) ([]db.ListGovernanceWorkPlanItemsRow, error) {
	f.workPlanSearch = search
	f.workPlanProgramID = programID
	f.workPlanStatus = status
	f.workPlanYear = periodYear
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceWorkPlanItemsRow{{ActivityName: "Supervisi"}}, nil
}

func (f *fakeGovernanceService) ListPerformanceTargets(_ context.Context, search, status, employeeID string, periodYear int32) ([]db.ListGovernancePerformanceTargetsRow, error) {
	f.targetSearch = search
	f.targetStatus = status
	f.targetEmployeeID = employeeID
	f.targetYear = periodYear
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernancePerformanceTargetsRow{{Title: "Target Kinerja"}}, nil
}

func (f *fakeGovernanceService) ListEvidenceItems(_ context.Context, search, status, snpStandard string, periodYear int32) ([]db.ListGovernanceEvidenceItemsRow, error) {
	f.evidenceSearch = search
	f.evidenceStatus = status
	f.evidenceSNP = snpStandard
	f.evidenceYear = periodYear
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceEvidenceItemsRow{{Title: "Bukti"}}, nil
}

func (f *fakeGovernanceService) ListComplianceActions(_ context.Context, search, status, priority, sourceType, snpStandard, responsibleEmployeeID string, periodYear int32) ([]db.ListGovernanceComplianceActionsRow, error) {
	f.complianceSearch = search
	f.complianceStatus = status
	f.compliancePriority = priority
	f.complianceSourceType = sourceType
	f.complianceSNP = snpStandard
	f.complianceEmployeeID = responsibleEmployeeID
	f.complianceYear = periodYear
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListGovernanceComplianceActionsRow{{Title: "Tindak lanjut"}}, nil
}

func (f *fakeGovernanceService) CreateUnit(_ context.Context, arg db.CreateGovernanceUnitParams) (db.GovernanceUnit, error) {
	f.createUnitArg = arg
	return db.GovernanceUnit{ID: handlerTestUUID(201), Code: arg.Code, Name: arg.Name, UnitType: arg.UnitType}, f.mutationErr
}

func (f *fakeGovernanceService) UpdateUnit(_ context.Context, arg db.UpdateGovernanceUnitParams) (db.GovernanceUnit, error) {
	f.updateUnitArg = arg
	return db.GovernanceUnit{ID: arg.ID, Code: arg.Code, Name: arg.Name, UnitType: arg.UnitType}, f.mutationErr
}

func (f *fakeGovernanceService) CreatePosition(_ context.Context, arg db.CreateGovernancePositionParams) (db.GovernancePosition, error) {
	f.createPositionArg = arg
	return db.GovernancePosition{ID: handlerTestUUID(202), Title: arg.Title}, f.mutationErr
}

func (f *fakeGovernanceService) UpdatePosition(_ context.Context, arg db.UpdateGovernancePositionParams) (db.GovernancePosition, error) {
	f.updatePositionArg = arg
	return db.GovernancePosition{ID: arg.ID, Title: arg.Title}, f.mutationErr
}

func (f *fakeGovernanceService) CreateAssignment(_ context.Context, arg db.CreateGovernanceAssignmentParams) (db.GovernanceAssignment, error) {
	f.createAssignmentArg = arg
	return db.GovernanceAssignment{ID: handlerTestUUID(203), EmployeeID: arg.EmployeeID}, f.mutationErr
}

func (f *fakeGovernanceService) UpdateAssignment(_ context.Context, arg db.UpdateGovernanceAssignmentParams) (db.GovernanceAssignment, error) {
	f.updateAssignmentArg = arg
	return db.GovernanceAssignment{ID: arg.ID, EmployeeID: arg.EmployeeID}, f.mutationErr
}

func (f *fakeGovernanceService) CreateDocument(_ context.Context, arg db.CreateGovernanceDocumentParams) (db.GovernanceDocument, error) {
	f.createDocumentArg = arg
	return db.GovernanceDocument{ID: handlerTestUUID(204), Title: arg.Title}, f.mutationErr
}

func (f *fakeGovernanceService) UpdateDocument(_ context.Context, arg db.UpdateGovernanceDocumentParams) (db.GovernanceDocument, error) {
	f.updateDocumentArg = arg
	return db.GovernanceDocument{ID: arg.ID, Title: arg.Title}, f.mutationErr
}

func (f *fakeGovernanceService) CreateProgram(_ context.Context, arg db.CreateGovernanceProgramParams) (db.GovernanceProgram, error) {
	f.createProgramArg = arg
	return db.GovernanceProgram{ID: handlerTestUUID(205), PeriodYear: arg.PeriodYear, Code: arg.Code, Name: arg.Name, Status: arg.Status}, f.mutationErr
}

func (f *fakeGovernanceService) UpdateProgram(_ context.Context, arg db.UpdateGovernanceProgramParams) (db.GovernanceProgram, error) {
	f.updateProgramArg = arg
	return db.GovernanceProgram{ID: arg.ID, PeriodYear: arg.PeriodYear, Code: arg.Code, Name: arg.Name, Status: arg.Status}, f.mutationErr
}

func (f *fakeGovernanceService) CreateWorkPlanItem(_ context.Context, arg db.CreateGovernanceWorkPlanItemParams) (db.GovernanceWorkPlanItem, error) {
	f.createWorkPlanArg = arg
	return db.GovernanceWorkPlanItem{ID: handlerTestUUID(206), PeriodYear: arg.PeriodYear, ActivityName: arg.ActivityName, Status: arg.Status}, f.mutationErr
}

func (f *fakeGovernanceService) UpdateWorkPlanItem(_ context.Context, arg db.UpdateGovernanceWorkPlanItemParams) (db.GovernanceWorkPlanItem, error) {
	f.updateWorkPlanArg = arg
	return db.GovernanceWorkPlanItem{ID: arg.ID, PeriodYear: arg.PeriodYear, ActivityName: arg.ActivityName, Status: arg.Status}, f.mutationErr
}

func (f *fakeGovernanceService) CreatePerformanceTarget(_ context.Context, arg db.CreateGovernancePerformanceTargetParams) (db.GovernancePerformanceTarget, error) {
	f.createTargetArg = arg
	return db.GovernancePerformanceTarget{ID: handlerTestUUID(207), PeriodYear: arg.PeriodYear, Title: arg.Title, Status: arg.Status}, f.mutationErr
}

func (f *fakeGovernanceService) UpdatePerformanceTarget(_ context.Context, arg db.UpdateGovernancePerformanceTargetParams) (db.GovernancePerformanceTarget, error) {
	f.updateTargetArg = arg
	return db.GovernancePerformanceTarget{ID: arg.ID, PeriodYear: arg.PeriodYear, Title: arg.Title, Status: arg.Status}, f.mutationErr
}

func (f *fakeGovernanceService) CreateEvidenceItem(_ context.Context, arg db.CreateGovernanceEvidenceItemParams) (db.GovernanceEvidenceItem, error) {
	f.createEvidenceArg = arg
	return db.GovernanceEvidenceItem{ID: handlerTestUUID(208), Title: arg.Title}, f.mutationErr
}

func (f *fakeGovernanceService) UpdateEvidenceItem(_ context.Context, arg db.UpdateGovernanceEvidenceItemParams) (db.GovernanceEvidenceItem, error) {
	f.updateEvidenceArg = arg
	return db.GovernanceEvidenceItem{ID: arg.ID, Title: arg.Title}, f.mutationErr
}

func (f *fakeGovernanceService) CreateComplianceAction(_ context.Context, arg db.CreateGovernanceComplianceActionParams) (db.GovernanceComplianceAction, error) {
	f.createComplianceArg = arg
	return db.GovernanceComplianceAction{ID: handlerTestUUID(209), Title: arg.Title}, f.mutationErr
}

func (f *fakeGovernanceService) UpdateComplianceAction(_ context.Context, arg db.UpdateGovernanceComplianceActionParams) (db.GovernanceComplianceAction, error) {
	f.updateComplianceArg = arg
	return db.GovernanceComplianceAction{ID: arg.ID, Title: arg.Title}, f.mutationErr
}

func (f *fakeGovernanceService) DeleteUnit(_ context.Context, id pgtype.UUID) error {
	f.deleteIDs = append(f.deleteIDs, id)
	return f.mutationErr
}

func (f *fakeGovernanceService) DeletePosition(ctx context.Context, id pgtype.UUID) error {
	return f.DeleteUnit(ctx, id)
}

func (f *fakeGovernanceService) DeleteAssignment(ctx context.Context, id pgtype.UUID) error {
	return f.DeleteUnit(ctx, id)
}

func (f *fakeGovernanceService) DeleteDocument(ctx context.Context, id pgtype.UUID) error {
	return f.DeleteUnit(ctx, id)
}

func (f *fakeGovernanceService) DeleteProgram(ctx context.Context, id pgtype.UUID) error {
	return f.DeleteUnit(ctx, id)
}

func (f *fakeGovernanceService) DeleteWorkPlanItem(ctx context.Context, id pgtype.UUID) error {
	return f.DeleteUnit(ctx, id)
}

func (f *fakeGovernanceService) DeletePerformanceTarget(ctx context.Context, id pgtype.UUID) error {
	return f.DeleteUnit(ctx, id)
}

func (f *fakeGovernanceService) DeleteEvidenceItem(ctx context.Context, id pgtype.UUID) error {
	return f.DeleteUnit(ctx, id)
}

func (f *fakeGovernanceService) DeleteComplianceAction(ctx context.Context, id pgtype.UUID) error {
	return f.DeleteUnit(ctx, id)
}

func TestGovernanceListHandlersForwardFilters(t *testing.T) {
	employeeID := handlerTestUUID(210)
	programID := handlerTestUUID(211)
	fake := &fakeGovernanceService{}
	h := &Governance{svc: fake}

	tests := []struct {
		name string
		fn   http.HandlerFunc
		url  string
	}{
		{name: "stats", fn: h.Stats, url: "/api/governance/stats"},
		{name: "snp", fn: h.SNPMatrix, url: "/api/governance/snp-matrix"},
		{name: "employees", fn: h.EmployeeOptions, url: "/api/governance/employee-options"},
		{name: "units", fn: h.ListUnits, url: "/api/governance/units?search=kurikulum"},
		{name: "positions", fn: h.ListPositions, url: "/api/governance/positions?search=waka"},
		{name: "assignments", fn: h.ListAssignments, url: "/api/governance/assignments?search=guru&active_only=yes"},
		{name: "documents", fn: h.ListDocuments, url: "/api/governance/documents?search=rkjm&doc_type=rencana&snp_standard=standar_isi&period_year=2026"},
		{name: "programs", fn: h.ListPrograms, url: "/api/governance/programs?search=mutu&status=active&snp_standard=standar_proses&period_year=2027"},
		{name: "work plans", fn: h.ListWorkPlanItems, url: "/api/governance/work-plan-items?search=supervisi&program_id=" + programID.String() + "&status=active&period_year=2028"},
		{name: "targets", fn: h.ListPerformanceTargets, url: "/api/governance/performance-targets?search=target&status=review&employee_id=" + employeeID.String() + "&period_year=2029"},
		{name: "evidence", fn: h.ListEvidenceItems, url: "/api/governance/evidence-items?search=bukti&status=verified&snp_standard=standar_ptk&period_year=2030"},
		{name: "compliance", fn: h.ListComplianceActions, url: "/api/governance/compliance-actions?search=tl&status=open&priority=high&source_type=audit&snp_standard=standar_sarpras&responsible_employee_id=" + employeeID.String() + "&period_year=2031"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, adminRequest(http.MethodGet, tt.url, ""))
			if rec.Code != http.StatusOK {
				t.Fatalf("%s status = %d, want 200; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}

	if fake.unitSearch != "kurikulum" || fake.positionSearch != "waka" {
		t.Fatalf("unit/position search = %q/%q, want forwarded search", fake.unitSearch, fake.positionSearch)
	}
	if fake.assignmentSearch != "guru" || !fake.assignmentActive {
		t.Fatalf("assignment filters = %q/%v, want search active", fake.assignmentSearch, fake.assignmentActive)
	}
	if fake.documentSearch != "rkjm" || fake.documentType != "rencana" || fake.documentSNP != "standar_isi" || fake.documentYear != 2026 {
		t.Fatalf("document filters = %q/%q/%q/%d, want forwarded document filters", fake.documentSearch, fake.documentType, fake.documentSNP, fake.documentYear)
	}
	if fake.programSearch != "mutu" || fake.programStatus != "active" || fake.programSNP != "standar_proses" || fake.programYear != 2027 {
		t.Fatalf("program filters = %q/%q/%q/%d, want forwarded program filters", fake.programSearch, fake.programStatus, fake.programSNP, fake.programYear)
	}
	if fake.workPlanSearch != "supervisi" || fake.workPlanProgramID != programID.String() || fake.workPlanStatus != "active" || fake.workPlanYear != 2028 {
		t.Fatalf("work plan filters = %q/%q/%q/%d, want forwarded filters", fake.workPlanSearch, fake.workPlanProgramID, fake.workPlanStatus, fake.workPlanYear)
	}
	if fake.targetSearch != "target" || fake.targetStatus != "review" || fake.targetEmployeeID != employeeID.String() || fake.targetYear != 2029 {
		t.Fatalf("target filters = %q/%q/%q/%d, want forwarded filters", fake.targetSearch, fake.targetStatus, fake.targetEmployeeID, fake.targetYear)
	}
	if fake.evidenceSearch != "bukti" || fake.evidenceStatus != "verified" || fake.evidenceSNP != "standar_ptk" || fake.evidenceYear != 2030 {
		t.Fatalf("evidence filters = %q/%q/%q/%d, want forwarded filters", fake.evidenceSearch, fake.evidenceStatus, fake.evidenceSNP, fake.evidenceYear)
	}
	if fake.complianceSearch != "tl" || fake.complianceStatus != "open" || fake.compliancePriority != "high" || fake.complianceSourceType != "audit" || fake.complianceSNP != "standar_sarpras" || fake.complianceEmployeeID != employeeID.String() || fake.complianceYear != 2031 {
		t.Fatalf("compliance filters = %+v, want forwarded compliance filters", fake)
	}
}

func TestGovernanceCreateUpdateAndDeleteCoreEntities(t *testing.T) {
	unitID := handlerTestUUID(220)
	positionID := handlerTestUUID(221)
	employeeID := handlerTestUUID(222)
	documentID := handlerTestUUID(223)
	programID := handlerTestUUID(224)
	fake := &fakeGovernanceService{}
	h := &Governance{svc: fake}

	createCases := []struct {
		name  string
		fn    http.HandlerFunc
		url   string
		body  string
		check func(t *testing.T)
	}{
		{
			name: "unit",
			fn:   h.CreateUnit,
			url:  "/api/governance/units",
			body: `{"code":"KUR","name":"Kurikulum","unit_type":"bidang","description":"Akademik","is_active":false,"sort_order":2}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createUnitArg.Code != "KUR" || fake.createUnitArg.Name != "Kurikulum" || fake.createUnitArg.IsActive {
					t.Fatalf("CreateUnit arg = %+v, want body fields and inactive", fake.createUnitArg)
				}
			},
		},
		{
			name: "position",
			fn:   h.CreatePosition,
			url:  "/api/governance/positions",
			body: `{"unit_id":"` + unitID.String() + `","title":"Waka Kurikulum","position_type":"struktural","tupoksi":"Koordinasi","sort_order":1}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createPositionArg.UnitID != unitID || fake.createPositionArg.Title != "Waka Kurikulum" || !fake.createPositionArg.IsActive {
					t.Fatalf("CreatePosition arg = %+v, want mapped position", fake.createPositionArg)
				}
			},
		},
		{
			name: "assignment",
			fn:   h.CreateAssignment,
			url:  "/api/governance/assignments",
			body: `{"position_id":"` + positionID.String() + `","employee_id":"` + employeeID.String() + `","start_date":"2026-01-02","end_date":"2026-12-31","decree_outgoing_letter_id":"` + documentID.String() + `","notes":"SK Kepala"}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createAssignmentArg.PositionID != positionID || fake.createAssignmentArg.EmployeeID != employeeID || !fake.createAssignmentArg.StartDate.Valid || fake.createAssignmentArg.Notes != "SK Kepala" {
					t.Fatalf("CreateAssignment arg = %+v, want mapped assignment", fake.createAssignmentArg)
				}
			},
		},
		{
			name: "document",
			fn:   h.CreateDocument,
			url:  "/api/governance/documents",
			body: `{"doc_type":"rencana","title":"RKJM","period_year":2026,"period_label":"2026","owner_unit_id":"` + unitID.String() + `","snp_standard":"standar_isi","status":"draft","document_url":"https://example.test/rkjm","outgoing_letter_id":"` + documentID.String() + `","summary":"Ringkas"}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createDocumentArg.OwnerUnitID != unitID || fake.createDocumentArg.Title != "RKJM" || fake.createDocumentArg.PeriodYear != 2026 || !fake.createDocumentArg.CreatedByUserID.Valid {
					t.Fatalf("CreateDocument arg = %+v, want mapped document and actor", fake.createDocumentArg)
				}
			},
		},
		{
			name: "program",
			fn:   h.CreateProgram,
			url:  "/api/governance/programs",
			body: `{"period_year":2026,"code":"PRG-1","name":"Program Mutu","source_document_id":"` + documentID.String() + `","owner_unit_id":"` + unitID.String() + `","responsible_position_id":"` + positionID.String() + `","responsible_employee_id":"` + employeeID.String() + `","snp_standard":"standar_proses","status":"active","progress_percent":40,"due_date":"2026-10-10"}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createProgramArg.Code != "PRG-1" || fake.createProgramArg.Name != "Program Mutu" || fake.createProgramArg.OwnerUnitID != unitID || !fake.createProgramArg.DueDate.Valid {
					t.Fatalf("CreateProgram arg = %+v, want mapped program", fake.createProgramArg)
				}
			},
		},
	}
	for _, tt := range createCases {
		t.Run("create "+tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, adminRequest(http.MethodPost, tt.url, tt.body))
			if rec.Code != http.StatusCreated {
				t.Fatalf("Create %s status = %d, want 201; body=%s", tt.name, rec.Code, rec.Body.String())
			}
			tt.check(t)
		})
	}

	updateCases := []struct {
		name  string
		fn    http.HandlerFunc
		url   string
		body  string
		check func(t *testing.T)
	}{
		{
			name: "unit",
			fn:   h.UpdateUnit,
			url:  "/api/governance/units/" + unitID.String(),
			body: `{"code":"KUR2","name":"Kurikulum Baru","unit_type":"bidang","is_active":true}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.updateUnitArg.ID != unitID || fake.updateUnitArg.Code != "KUR2" || fake.updateUnitArg.Name != "Kurikulum Baru" {
					t.Fatalf("UpdateUnit arg = %+v, want route/body fields", fake.updateUnitArg)
				}
			},
		},
		{
			name: "position",
			fn:   h.UpdatePosition,
			url:  "/api/governance/positions/" + positionID.String(),
			body: `{"unit_id":"` + unitID.String() + `","title":"Koordinator Kurikulum","position_type":"struktural","is_active":true}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.updatePositionArg.ID != positionID || fake.updatePositionArg.UnitID != unitID || fake.updatePositionArg.Title != "Koordinator Kurikulum" {
					t.Fatalf("UpdatePosition arg = %+v, want route/body fields", fake.updatePositionArg)
				}
			},
		},
		{
			name: "assignment",
			fn:   h.UpdateAssignment,
			url:  "/api/governance/assignments/" + handlerTestUUID(225).String(),
			body: `{"position_id":"` + positionID.String() + `","employee_id":"` + employeeID.String() + `","start_date":"2026-02-01","notes":"Revisi"}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.updateAssignmentArg.PositionID != positionID || fake.updateAssignmentArg.EmployeeID != employeeID || fake.updateAssignmentArg.Notes != "Revisi" {
					t.Fatalf("UpdateAssignment arg = %+v, want mapped assignment", fake.updateAssignmentArg)
				}
			},
		},
		{
			name: "document",
			fn:   h.UpdateDocument,
			url:  "/api/governance/documents/" + documentID.String(),
			body: `{"doc_type":"laporan","title":"Laporan EDM","period_year":2027,"owner_unit_id":"` + unitID.String() + `","snp_standard":"standar_isi","status":"final"}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.updateDocumentArg.ID != documentID || fake.updateDocumentArg.Title != "Laporan EDM" || fake.updateDocumentArg.PeriodYear != 2027 {
					t.Fatalf("UpdateDocument arg = %+v, want mapped document", fake.updateDocumentArg)
				}
			},
		},
		{
			name: "program",
			fn:   h.UpdateProgram,
			url:  "/api/governance/programs/" + programID.String(),
			body: `{"period_year":2027,"code":"PRG-2","name":"Program Supervisi","owner_unit_id":"` + unitID.String() + `","status":"active","progress_percent":80}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.updateProgramArg.ID != programID || fake.updateProgramArg.Code != "PRG-2" || fake.updateProgramArg.ProgressPercent != 80 {
					t.Fatalf("UpdateProgram arg = %+v, want mapped program", fake.updateProgramArg)
				}
			},
		},
	}
	for _, tt := range updateCases {
		t.Run("update "+tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(http.MethodPatch, tt.url, tt.body)
			req = withRouteParam(req, "id", tt.url[strings.LastIndex(tt.url, "/")+1:])
			tt.fn(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("Update %s status = %d, want 200; body=%s", tt.name, rec.Code, rec.Body.String())
			}
			tt.check(t)
		})
	}

	deleteTargets := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "unit", fn: h.DeleteUnit},
		{name: "position", fn: h.DeletePosition},
		{name: "assignment", fn: h.DeleteAssignment},
		{name: "document", fn: h.DeleteDocument},
		{name: "program", fn: h.DeleteProgram},
		{name: "work_plan", fn: h.DeleteWorkPlanItem},
		{name: "target", fn: h.DeletePerformanceTarget},
		{name: "evidence", fn: h.DeleteEvidenceItem},
		{name: "compliance", fn: h.DeleteComplianceAction},
	}
	for _, tt := range deleteTargets {
		t.Run("delete "+tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(http.MethodDelete, "/api/governance/"+tt.name+"/"+unitID.String(), "")
			req = withRouteParam(req, "id", unitID.String())
			tt.fn(rec, req)
			if rec.Code != http.StatusNoContent {
				t.Fatalf("Delete %s status = %d, want 204; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}
	if len(fake.deleteIDs) != len(deleteTargets) {
		t.Fatalf("delete calls = %d, want %d", len(fake.deleteIDs), len(deleteTargets))
	}
}

func TestGovernanceCreateUpdateOperationalEntities(t *testing.T) {
	unitID := handlerTestUUID(240)
	employeeID := handlerTestUUID(241)
	documentID := handlerTestUUID(242)
	programID := handlerTestUUID(243)
	workPlanID := handlerTestUUID(244)
	targetID := handlerTestUUID(245)
	evidenceID := handlerTestUUID(246)
	complianceID := handlerTestUUID(247)
	fake := &fakeGovernanceService{}
	h := &Governance{svc: fake}

	workPlanBody := `{"period_year":2026,"program_id":"` + programID.String() + `","source_document_id":"` + documentID.String() + `","owner_unit_id":"` + unitID.String() + `","responsible_employee_id":"` + employeeID.String() + `","evidence_item_id":"` + evidenceID.String() + `","activity_code":"WP-1","activity_name":"Supervisi EDM","output_indicator":"Laporan","target_volume":"1","target_unit":"dokumen","budget_source":"BOS","budget_amount":1000,"realization_amount":250,"status":"running","progress_percent":25,"start_date":"2026-05-01","end_date":"2026-05-31","evidence_url":"/bukti","notes":"catatan"}`
	rec := httptest.NewRecorder()
	h.CreateWorkPlanItem(rec, adminRequest(http.MethodPost, "/api/governance/work-plan-items", workPlanBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateWorkPlanItem() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createWorkPlanArg.ProgramID != programID || fake.createWorkPlanArg.ActivityName != "Supervisi EDM" || fake.createWorkPlanArg.BudgetAmount != 1000 || !fake.createWorkPlanArg.StartDate.Valid {
		t.Fatalf("CreateWorkPlanItem() arg = %+v, want mapped work plan", fake.createWorkPlanArg)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodPut, "/api/governance/work-plan-items/"+workPlanID.String(), workPlanBody), "id", workPlanID.String())
	h.UpdateWorkPlanItem(rec, req)
	if rec.Code != http.StatusOK || fake.updateWorkPlanArg.ID != workPlanID || fake.updateWorkPlanArg.ProgressPercent != 25 {
		t.Fatalf("UpdateWorkPlanItem() status/arg = %d/%+v, want 200 and mapped update", rec.Code, fake.updateWorkPlanArg)
	}

	targetBody := `{"period_year":2026,"employee_id":"` + employeeID.String() + `","program_id":"` + programID.String() + `","parent_target_id":"` + targetID.String() + `","aspect":"hasil","title":"Target EDM","indicator":"Nilai EDM","target_value":"90","target_unit":"persen","status":"running","progress_percent":35,"evidence_url":"/evidence","review_notes":"cek","due_date":"2026-06-30"}`
	rec = httptest.NewRecorder()
	h.CreatePerformanceTarget(rec, adminRequest(http.MethodPost, "/api/governance/performance-targets", targetBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreatePerformanceTarget() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createTargetArg.EmployeeID != employeeID || fake.createTargetArg.Title != "Target EDM" || !fake.createTargetArg.CreatedByUserID.Valid || !fake.createTargetArg.DueDate.Valid {
		t.Fatalf("CreatePerformanceTarget() arg = %+v, want mapped target with actor/date", fake.createTargetArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/governance/performance-targets/"+targetID.String(), targetBody), "id", targetID.String())
	h.UpdatePerformanceTarget(rec, req)
	if rec.Code != http.StatusOK || fake.updateTargetArg.ID != targetID || fake.updateTargetArg.ProgressPercent != 35 {
		t.Fatalf("UpdatePerformanceTarget() status/arg = %d/%+v, want 200 and mapped update", rec.Code, fake.updateTargetArg)
	}

	evidenceBody := `{"period_year":2026,"title":"Bukti EDM","evidence_type":"document","snp_standard":"skl","owner_unit_id":"` + unitID.String() + `","document_id":"` + documentID.String() + `","program_id":"` + programID.String() + `","performance_target_id":"` + targetID.String() + `","source_module":"governance","evidence_url":"/evidence","status":"available","notes":"lengkap"}`
	rec = httptest.NewRecorder()
	h.CreateEvidenceItem(rec, adminRequest(http.MethodPost, "/api/governance/evidence-items", evidenceBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateEvidenceItem() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createEvidenceArg.Title != "Bukti EDM" || fake.createEvidenceArg.ProgramID != programID || !fake.createEvidenceArg.CreatedByUserID.Valid {
		t.Fatalf("CreateEvidenceItem() arg = %+v, want mapped evidence with actor", fake.createEvidenceArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/governance/evidence-items/"+evidenceID.String(), evidenceBody), "id", evidenceID.String())
	h.UpdateEvidenceItem(rec, req)
	if rec.Code != http.StatusOK || fake.updateEvidenceArg.ID != evidenceID || fake.updateEvidenceArg.Title != "Bukti EDM" {
		t.Fatalf("UpdateEvidenceItem() status/arg = %d/%+v, want 200 and mapped update", rec.Code, fake.updateEvidenceArg)
	}

	complianceBody := `{"period_year":2026,"source_type":"manual","source_ref_id":"` + evidenceID.String() + `","snp_standard":"skl","program_id":"` + programID.String() + `","document_id":"` + documentID.String() + `","performance_target_id":"` + targetID.String() + `","evidence_item_id":"` + evidenceID.String() + `","owner_unit_id":"` + unitID.String() + `","responsible_employee_id":"` + employeeID.String() + `","title":"Tindak Lanjut EDM","description":"Lengkapi bukti","priority":"high","status":"open","due_date":"2026-07-01","completed_at":"2026-07-02T08:00:00Z","follow_up_notes":"catatan","evidence_url":"/tl"}`
	rec = httptest.NewRecorder()
	h.CreateComplianceAction(rec, adminRequest(http.MethodPost, "/api/governance/compliance-actions", complianceBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateComplianceAction() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createComplianceArg.Title != "Tindak Lanjut EDM" || fake.createComplianceArg.ResponsibleEmployeeID != employeeID || !fake.createComplianceArg.CreatedByUserID.Valid || !fake.createComplianceArg.CompletedAt.Valid {
		t.Fatalf("CreateComplianceAction() arg = %+v, want mapped compliance with actor/timestamp", fake.createComplianceArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/api/governance/compliance-actions/"+complianceID.String(), complianceBody), "id", complianceID.String())
	h.UpdateComplianceAction(rec, req)
	if rec.Code != http.StatusOK || fake.updateComplianceArg.ID != complianceID || fake.updateComplianceArg.Title != "Tindak Lanjut EDM" {
		t.Fatalf("UpdateComplianceAction() status/arg = %d/%+v, want 200 and mapped update", rec.Code, fake.updateComplianceArg)
	}
}

func TestGovernanceMutationHandlersWriteAuditEvents(t *testing.T) {
	unitID := handlerTestUUID(250)
	documentID := handlerTestUUID(251)
	positionID := handlerTestUUID(252)
	employeeID := handlerTestUUID(253)
	programID := handlerTestUUID(254)
	workPlanID := handlerTestUUID(255)
	targetID := handlerTestUUID(156)
	fake := &fakeGovernanceService{}
	audit := &fakeCbtSessionAuditWriter{}
	h := &Governance{svc: fake, audit: audit}
	auditedAdminRequest := func(method, target, body string) *http.Request {
		return withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"admin"},
			"uid":   "01000000-0000-0000-0000-000000000000",
			"sub":   "01000000-0000-0000-0000-000000000000",
			"usr":   "admin.tatakelola",
			"ssid":  "sess-governance-1",
		})
	}

	unitBody := `{"code":"KUR","name":"Kurikulum","unit_type":"bidang","description":"Akademik","is_active":false,"sort_order":2}`
	rec := httptest.NewRecorder()
	h.CreateUnit(rec, auditedAdminRequest(http.MethodPost, "/api/governance/units", unitBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateUnit status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.UpdateUnit(rec, withRouteParam(auditedAdminRequest(http.MethodPatch, "/api/governance/units/"+unitID.String(), `{"code":"KUR2","name":"Kurikulum Baru","unit_type":"bidang","is_active":true}`), "id", unitID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateUnit status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.DeleteUnit(rec, withRouteParam(auditedAdminRequest(http.MethodDelete, "/api/governance/units/"+unitID.String(), ""), "id", unitID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteUnit status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	programBody := `{"period_year":2026,"code":"PRG-1","name":"Program Mutu","source_document_id":"` + documentID.String() + `","owner_unit_id":"` + unitID.String() + `","responsible_position_id":"` + positionID.String() + `","responsible_employee_id":"` + employeeID.String() + `","snp_standard":"standar_proses","status":"active","progress_percent":40,"due_date":"2026-10-10"}`
	rec = httptest.NewRecorder()
	h.CreateProgram(rec, auditedAdminRequest(http.MethodPost, "/api/governance/programs", programBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateProgram status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.UpdateProgram(rec, withRouteParam(auditedAdminRequest(http.MethodPatch, "/api/governance/programs/"+programID.String(), `{"period_year":2027,"code":"PRG-2","name":"Program Supervisi","owner_unit_id":"`+unitID.String()+`","status":"active","progress_percent":80}`), "id", programID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateProgram status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.DeleteProgram(rec, withRouteParam(auditedAdminRequest(http.MethodDelete, "/api/governance/programs/"+programID.String(), ""), "id", programID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteProgram status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	workPlanBody := `{"period_year":2026,"program_id":"` + programID.String() + `","source_document_id":"` + documentID.String() + `","owner_unit_id":"` + unitID.String() + `","responsible_employee_id":"` + employeeID.String() + `","activity_code":"WP-1","activity_name":"Supervisi EDM","output_indicator":"Laporan","target_volume":"1","target_unit":"dokumen","budget_source":"BOS","budget_amount":1000,"realization_amount":250,"status":"running","progress_percent":25,"start_date":"2026-05-01","end_date":"2026-05-31","evidence_url":"/bukti","notes":"catatan"}`
	rec = httptest.NewRecorder()
	h.CreateWorkPlanItem(rec, auditedAdminRequest(http.MethodPost, "/api/governance/work-plan-items", workPlanBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateWorkPlanItem status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.UpdateWorkPlanItem(rec, withRouteParam(auditedAdminRequest(http.MethodPut, "/api/governance/work-plan-items/"+workPlanID.String(), workPlanBody), "id", workPlanID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateWorkPlanItem status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.DeleteWorkPlanItem(rec, withRouteParam(auditedAdminRequest(http.MethodDelete, "/api/governance/work-plan-items/"+workPlanID.String(), ""), "id", workPlanID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteWorkPlanItem status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	targetBody := `{"period_year":2026,"employee_id":"` + employeeID.String() + `","program_id":"` + programID.String() + `","parent_target_id":"` + targetID.String() + `","aspect":"hasil","title":"Target EDM","indicator":"Nilai EDM","target_value":"90","target_unit":"persen","status":"running","progress_percent":35,"evidence_url":"/evidence","review_notes":"cek","due_date":"2026-06-30"}`
	rec = httptest.NewRecorder()
	h.CreatePerformanceTarget(rec, auditedAdminRequest(http.MethodPost, "/api/governance/performance-targets", targetBody))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreatePerformanceTarget status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.UpdatePerformanceTarget(rec, withRouteParam(auditedAdminRequest(http.MethodPut, "/api/governance/performance-targets/"+targetID.String(), targetBody), "id", targetID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdatePerformanceTarget status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.DeletePerformanceTarget(rec, withRouteParam(auditedAdminRequest(http.MethodDelete, "/api/governance/performance-targets/"+targetID.String(), ""), "id", targetID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeletePerformanceTarget status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	if len(audit.entries) != 12 {
		t.Fatalf("audit entries = %d, want 12", len(audit.entries))
	}

	actions := []string{
		"GOVERNANCE_UNIT_CREATE",
		"GOVERNANCE_UNIT_UPDATE",
		"GOVERNANCE_UNIT_DELETE",
		"GOVERNANCE_PROGRAM_CREATE",
		"GOVERNANCE_PROGRAM_UPDATE",
		"GOVERNANCE_PROGRAM_DELETE",
		"GOVERNANCE_WORK_PLAN_CREATE",
		"GOVERNANCE_WORK_PLAN_UPDATE",
		"GOVERNANCE_WORK_PLAN_DELETE",
		"GOVERNANCE_TARGET_CREATE",
		"GOVERNANCE_TARGET_UPDATE",
		"GOVERNANCE_TARGET_DELETE",
	}
	for i, action := range actions {
		if audit.entries[i].Action != action {
			t.Fatalf("audit[%d].Action = %q, want %q", i, audit.entries[i].Action, action)
		}
	}

	meta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if meta["username"] != "admin.tatakelola" || meta["name"] != "Kurikulum" {
		t.Fatalf("unit create metadata = %+v, want username/name", meta)
	}
	meta = mustAuditMetadataMap(t, audit.entries[5].Metadata)
	if meta["deleted_by"] != "admin.tatakelola" {
		t.Fatalf("program delete metadata = %+v, want deleted_by", meta)
	}
	meta = mustAuditMetadataMap(t, audit.entries[6].Metadata)
	if meta["activity_name"] != "Supervisi EDM" {
		t.Fatalf("work plan create metadata = %+v, want activity_name", meta)
	}
	meta = mustAuditMetadataMap(t, audit.entries[9].Metadata)
	if meta["title"] != "Target EDM" {
		t.Fatalf("target create metadata = %+v, want title", meta)
	}
}

func TestGovernanceHandlersMapServiceErrors(t *testing.T) {
	id := handlerTestUUID(230)
	unitID := handlerTestUUID(231)
	positionID := handlerTestUUID(232)
	employeeID := handlerTestUUID(233)
	documentID := handlerTestUUID(234)
	programID := handlerTestUUID(235)
	evidenceID := handlerTestUUID(236)
	targetID := handlerTestUUID(237)
	validationErr := errors.New("validasi gagal")
	unitBody := `{"code":"KUR","name":"Kurikulum","unit_type":"bidang"}`
	positionBody := `{"unit_id":"` + unitID.String() + `","title":"Waka Kurikulum","position_type":"struktural"}`
	assignmentBody := `{"position_id":"` + positionID.String() + `","employee_id":"` + employeeID.String() + `","start_date":"2026-01-02"}`
	documentBody := `{"doc_type":"rencana","title":"RKJM","period_year":2026,"owner_unit_id":"` + unitID.String() + `","snp_standard":"standar_isi","status":"draft"}`
	programBody := `{"period_year":2026,"code":"PRG-1","name":"Program Mutu","owner_unit_id":"` + unitID.String() + `","responsible_position_id":"` + positionID.String() + `","responsible_employee_id":"` + employeeID.String() + `","status":"active"}`
	workPlanBody := `{"period_year":2026,"program_id":"` + programID.String() + `","source_document_id":"` + documentID.String() + `","owner_unit_id":"` + unitID.String() + `","responsible_employee_id":"` + employeeID.String() + `","evidence_item_id":"` + evidenceID.String() + `","activity_code":"WP-1","activity_name":"Supervisi EDM","status":"running"}`
	targetBody := `{"period_year":2026,"employee_id":"` + employeeID.String() + `","program_id":"` + programID.String() + `","parent_target_id":"` + targetID.String() + `","aspect":"hasil","title":"Target EDM","status":"running"}`
	evidenceBody := `{"period_year":2026,"program_id":"` + programID.String() + `","work_plan_item_id":"` + id.String() + `","performance_target_id":"` + targetID.String() + `","document_id":"` + documentID.String() + `","owner_unit_id":"` + unitID.String() + `","uploaded_by_employee_id":"` + employeeID.String() + `","title":"Bukti EDM","evidence_url":"/bukti","status":"verified"}`
	complianceBody := `{"period_year":2026,"source_type":"manual","source_ref_id":"` + evidenceID.String() + `","snp_standard":"skl","program_id":"` + programID.String() + `","document_id":"` + documentID.String() + `","performance_target_id":"` + targetID.String() + `","evidence_item_id":"` + evidenceID.String() + `","owner_unit_id":"` + unitID.String() + `","responsible_employee_id":"` + employeeID.String() + `","title":"Tindak Lanjut EDM","status":"open"}`
	tests := []struct {
		name string
		fn   http.HandlerFunc
		req  *http.Request
		want int
	}{
		{
			name: "stats internal",
			fn:   (&Governance{svc: &fakeGovernanceService{statsErr: errors.New("db down")}}).Stats,
			req:  adminRequest(http.MethodGet, "/api/governance/stats", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "list internal",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListPrograms,
			req:  adminRequest(http.MethodGet, "/api/governance/programs", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "snp matrix internal",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).SNPMatrix,
			req:  adminRequest(http.MethodGet, "/api/governance/snp-matrix", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "employee options internal",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).EmployeeOptions,
			req:  adminRequest(http.MethodGet, "/api/governance/employee-options", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "units internal",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListUnits,
			req:  adminRequest(http.MethodGet, "/api/governance/units", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "positions internal",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListPositions,
			req:  adminRequest(http.MethodGet, "/api/governance/positions", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "assignments internal",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListAssignments,
			req:  adminRequest(http.MethodGet, "/api/governance/assignments", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "documents internal",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListDocuments,
			req:  adminRequest(http.MethodGet, "/api/governance/documents", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "work plan client error",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListWorkPlanItems,
			req:  adminRequest(http.MethodGet, "/api/governance/work-plan-items", ""),
			want: http.StatusBadRequest,
		},
		{
			name: "targets client error",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListPerformanceTargets,
			req:  adminRequest(http.MethodGet, "/api/governance/performance-targets", ""),
			want: http.StatusBadRequest,
		},
		{
			name: "evidence client error",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListEvidenceItems,
			req:  adminRequest(http.MethodGet, "/api/governance/evidence-items", ""),
			want: http.StatusBadRequest,
		},
		{
			name: "compliance client error",
			fn:   (&Governance{svc: &fakeGovernanceService{listErr: errors.New("db down")}}).ListComplianceActions,
			req:  adminRequest(http.MethodGet, "/api/governance/compliance-actions", ""),
			want: http.StatusBadRequest,
		},
		{
			name: "create client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: errors.New("nama wajib diisi")}}).CreateUnit,
			req:  adminRequest(http.MethodPost, "/api/governance/units", `{"code":"KUR","name":"Kurikulum"}`),
			want: http.StatusBadRequest,
		},
		{
			name: "update unit client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdateUnit,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/units/"+unitID.String(), unitBody), "id", unitID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "create position client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).CreatePosition,
			req:  adminRequest(http.MethodPost, "/api/governance/positions", positionBody),
			want: http.StatusBadRequest,
		},
		{
			name: "update position client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdatePosition,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/positions/"+positionID.String(), positionBody), "id", positionID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "create assignment client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).CreateAssignment,
			req:  adminRequest(http.MethodPost, "/api/governance/assignments", assignmentBody),
			want: http.StatusBadRequest,
		},
		{
			name: "update assignment client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdateAssignment,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/assignments/"+id.String(), assignmentBody), "id", id.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "create document client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).CreateDocument,
			req:  adminRequest(http.MethodPost, "/api/governance/documents", documentBody),
			want: http.StatusBadRequest,
		},
		{
			name: "update document client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdateDocument,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/documents/"+documentID.String(), documentBody), "id", documentID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "create program client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).CreateProgram,
			req:  adminRequest(http.MethodPost, "/api/governance/programs", programBody),
			want: http.StatusBadRequest,
		},
		{
			name: "update program client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdateProgram,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/programs/"+programID.String(), programBody), "id", programID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "create work plan client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).CreateWorkPlanItem,
			req:  adminRequest(http.MethodPost, "/api/governance/work-plan-items", workPlanBody),
			want: http.StatusBadRequest,
		},
		{
			name: "update work plan client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdateWorkPlanItem,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/work-plan-items/"+id.String(), workPlanBody), "id", id.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "create target client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).CreatePerformanceTarget,
			req:  adminRequest(http.MethodPost, "/api/governance/performance-targets", targetBody),
			want: http.StatusBadRequest,
		},
		{
			name: "update target client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdatePerformanceTarget,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/performance-targets/"+targetID.String(), targetBody), "id", targetID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "create evidence client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).CreateEvidenceItem,
			req:  adminRequest(http.MethodPost, "/api/governance/evidence-items", evidenceBody),
			want: http.StatusBadRequest,
		},
		{
			name: "update evidence client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdateEvidenceItem,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/evidence-items/"+evidenceID.String(), evidenceBody), "id", evidenceID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "create compliance client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).CreateComplianceAction,
			req:  adminRequest(http.MethodPost, "/api/governance/compliance-actions", complianceBody),
			want: http.StatusBadRequest,
		},
		{
			name: "update compliance client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: validationErr}}).UpdateComplianceAction,
			req:  withRouteParam(adminRequest(http.MethodPut, "/api/governance/compliance-actions/"+id.String(), complianceBody), "id", id.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "delete client error",
			fn:   (&Governance{svc: &fakeGovernanceService{mutationErr: errors.New("masih dipakai")}}).DeleteProgram,
			req:  withRouteParam(adminRequest(http.MethodDelete, "/api/governance/programs/"+id.String(), ""), "id", id.String()),
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
