package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeGovernanceStore struct {
	unitSearch            string
	createUnitArg         db.CreateGovernanceUnitParams
	updateUnitArg         db.UpdateGovernanceUnitParams
	deleteUnitID          pgtype.UUID
	positionSearch        string
	createPositionArg     db.CreateGovernancePositionParams
	updatePositionArg     db.UpdateGovernancePositionParams
	deletePositionID      pgtype.UUID
	assignmentsArg        db.ListGovernanceAssignmentsParams
	createAssignmentArg   db.CreateGovernanceAssignmentParams
	updateAssignmentArg   db.UpdateGovernanceAssignmentParams
	deleteAssignmentID    pgtype.UUID
	documentsArg          db.ListGovernanceDocumentsParams
	createDocumentArg     db.CreateGovernanceDocumentParams
	updateDocumentArg     db.UpdateGovernanceDocumentParams
	deleteDocumentID      pgtype.UUID
	programsArg           db.ListGovernanceProgramsParams
	createProgramArg      db.CreateGovernanceProgramParams
	updateProgramArg      db.UpdateGovernanceProgramParams
	deleteProgramID       pgtype.UUID
	workPlanArg           db.ListGovernanceWorkPlanItemsParams
	createWorkPlanArg     db.CreateGovernanceWorkPlanItemParams
	updateWorkPlanArg     db.UpdateGovernanceWorkPlanItemParams
	deleteWorkPlanID      pgtype.UUID
	targetsArg            db.ListGovernancePerformanceTargetsParams
	createTargetArg       db.CreateGovernancePerformanceTargetParams
	updateTargetArg       db.UpdateGovernancePerformanceTargetParams
	deleteTargetID        pgtype.UUID
	evidenceArg           db.ListGovernanceEvidenceItemsParams
	createEvidenceArg     db.CreateGovernanceEvidenceItemParams
	updateEvidenceArg     db.UpdateGovernanceEvidenceItemParams
	deleteEvidenceID      pgtype.UUID
	complianceArg         db.ListGovernanceComplianceActionsParams
	createComplianceArg   db.CreateGovernanceComplianceActionParams
	updateComplianceArg   db.UpdateGovernanceComplianceActionParams
	deleteComplianceID    pgtype.UUID
	statsCalled           bool
	snpMatrixCalled       bool
	employeeOptionsCalled bool
}

func (f *fakeGovernanceStore) GetGovernanceStats(ctx context.Context) (db.GetGovernanceStatsRow, error) {
	f.statsCalled = true
	return db.GetGovernanceStatsRow{}, nil
}

func (f *fakeGovernanceStore) ListGovernanceSNPMatrix(ctx context.Context) ([]db.ListGovernanceSNPMatrixRow, error) {
	f.snpMatrixCalled = true
	return []db.ListGovernanceSNPMatrixRow{{}}, nil
}

func (f *fakeGovernanceStore) ListGovernanceEmployeeOptions(ctx context.Context) ([]db.ListGovernanceEmployeeOptionsRow, error) {
	f.employeeOptionsCalled = true
	return []db.ListGovernanceEmployeeOptionsRow{{}}, nil
}

func (f *fakeGovernanceStore) ListGovernanceUnits(ctx context.Context, search string) ([]db.ListGovernanceUnitsRow, error) {
	f.unitSearch = search
	return []db.ListGovernanceUnitsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernanceUnit(ctx context.Context, arg db.CreateGovernanceUnitParams) (db.GovernanceUnit, error) {
	f.createUnitArg = arg
	return db.GovernanceUnit{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernanceUnit(ctx context.Context, arg db.UpdateGovernanceUnitParams) (db.GovernanceUnit, error) {
	f.updateUnitArg = arg
	return db.GovernanceUnit{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernanceUnit(ctx context.Context, id pgtype.UUID) error {
	f.deleteUnitID = id
	return nil
}

func (f *fakeGovernanceStore) ListGovernancePositions(ctx context.Context, search string) ([]db.ListGovernancePositionsRow, error) {
	f.positionSearch = search
	return []db.ListGovernancePositionsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernancePosition(ctx context.Context, arg db.CreateGovernancePositionParams) (db.GovernancePosition, error) {
	f.createPositionArg = arg
	return db.GovernancePosition{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernancePosition(ctx context.Context, arg db.UpdateGovernancePositionParams) (db.GovernancePosition, error) {
	f.updatePositionArg = arg
	return db.GovernancePosition{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernancePosition(ctx context.Context, id pgtype.UUID) error {
	f.deletePositionID = id
	return nil
}

func (f *fakeGovernanceStore) ListGovernanceAssignments(ctx context.Context, arg db.ListGovernanceAssignmentsParams) ([]db.ListGovernanceAssignmentsRow, error) {
	f.assignmentsArg = arg
	return []db.ListGovernanceAssignmentsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernanceAssignment(ctx context.Context, arg db.CreateGovernanceAssignmentParams) (db.GovernanceAssignment, error) {
	f.createAssignmentArg = arg
	return db.GovernanceAssignment{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernanceAssignment(ctx context.Context, arg db.UpdateGovernanceAssignmentParams) (db.GovernanceAssignment, error) {
	f.updateAssignmentArg = arg
	return db.GovernanceAssignment{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernanceAssignment(ctx context.Context, id pgtype.UUID) error {
	f.deleteAssignmentID = id
	return nil
}

func (f *fakeGovernanceStore) ListGovernanceDocuments(ctx context.Context, arg db.ListGovernanceDocumentsParams) ([]db.ListGovernanceDocumentsRow, error) {
	f.documentsArg = arg
	return []db.ListGovernanceDocumentsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernanceDocument(ctx context.Context, arg db.CreateGovernanceDocumentParams) (db.GovernanceDocument, error) {
	f.createDocumentArg = arg
	return db.GovernanceDocument{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernanceDocument(ctx context.Context, arg db.UpdateGovernanceDocumentParams) (db.GovernanceDocument, error) {
	f.updateDocumentArg = arg
	return db.GovernanceDocument{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernanceDocument(ctx context.Context, id pgtype.UUID) error {
	f.deleteDocumentID = id
	return nil
}

func (f *fakeGovernanceStore) ListGovernancePrograms(ctx context.Context, arg db.ListGovernanceProgramsParams) ([]db.ListGovernanceProgramsRow, error) {
	f.programsArg = arg
	return []db.ListGovernanceProgramsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernanceProgram(ctx context.Context, arg db.CreateGovernanceProgramParams) (db.GovernanceProgram, error) {
	f.createProgramArg = arg
	return db.GovernanceProgram{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernanceProgram(ctx context.Context, arg db.UpdateGovernanceProgramParams) (db.GovernanceProgram, error) {
	f.updateProgramArg = arg
	return db.GovernanceProgram{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernanceProgram(ctx context.Context, id pgtype.UUID) error {
	f.deleteProgramID = id
	return nil
}

func (f *fakeGovernanceStore) ListGovernanceWorkPlanItems(ctx context.Context, arg db.ListGovernanceWorkPlanItemsParams) ([]db.ListGovernanceWorkPlanItemsRow, error) {
	f.workPlanArg = arg
	return []db.ListGovernanceWorkPlanItemsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernanceWorkPlanItem(ctx context.Context, arg db.CreateGovernanceWorkPlanItemParams) (db.GovernanceWorkPlanItem, error) {
	f.createWorkPlanArg = arg
	return db.GovernanceWorkPlanItem{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernanceWorkPlanItem(ctx context.Context, arg db.UpdateGovernanceWorkPlanItemParams) (db.GovernanceWorkPlanItem, error) {
	f.updateWorkPlanArg = arg
	return db.GovernanceWorkPlanItem{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernanceWorkPlanItem(ctx context.Context, id pgtype.UUID) error {
	f.deleteWorkPlanID = id
	return nil
}

func (f *fakeGovernanceStore) ListGovernancePerformanceTargets(ctx context.Context, arg db.ListGovernancePerformanceTargetsParams) ([]db.ListGovernancePerformanceTargetsRow, error) {
	f.targetsArg = arg
	return []db.ListGovernancePerformanceTargetsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernancePerformanceTarget(ctx context.Context, arg db.CreateGovernancePerformanceTargetParams) (db.GovernancePerformanceTarget, error) {
	f.createTargetArg = arg
	return db.GovernancePerformanceTarget{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernancePerformanceTarget(ctx context.Context, arg db.UpdateGovernancePerformanceTargetParams) (db.GovernancePerformanceTarget, error) {
	f.updateTargetArg = arg
	return db.GovernancePerformanceTarget{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernancePerformanceTarget(ctx context.Context, id pgtype.UUID) error {
	f.deleteTargetID = id
	return nil
}

func (f *fakeGovernanceStore) ListGovernanceEvidenceItems(ctx context.Context, arg db.ListGovernanceEvidenceItemsParams) ([]db.ListGovernanceEvidenceItemsRow, error) {
	f.evidenceArg = arg
	return []db.ListGovernanceEvidenceItemsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernanceEvidenceItem(ctx context.Context, arg db.CreateGovernanceEvidenceItemParams) (db.GovernanceEvidenceItem, error) {
	f.createEvidenceArg = arg
	return db.GovernanceEvidenceItem{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernanceEvidenceItem(ctx context.Context, arg db.UpdateGovernanceEvidenceItemParams) (db.GovernanceEvidenceItem, error) {
	f.updateEvidenceArg = arg
	return db.GovernanceEvidenceItem{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernanceEvidenceItem(ctx context.Context, id pgtype.UUID) error {
	f.deleteEvidenceID = id
	return nil
}

func (f *fakeGovernanceStore) ListGovernanceComplianceActions(ctx context.Context, arg db.ListGovernanceComplianceActionsParams) ([]db.ListGovernanceComplianceActionsRow, error) {
	f.complianceArg = arg
	return []db.ListGovernanceComplianceActionsRow{{}}, nil
}

func (f *fakeGovernanceStore) CreateGovernanceComplianceAction(ctx context.Context, arg db.CreateGovernanceComplianceActionParams) (db.GovernanceComplianceAction, error) {
	f.createComplianceArg = arg
	return db.GovernanceComplianceAction{}, nil
}

func (f *fakeGovernanceStore) UpdateGovernanceComplianceAction(ctx context.Context, arg db.UpdateGovernanceComplianceActionParams) (db.GovernanceComplianceAction, error) {
	f.updateComplianceArg = arg
	return db.GovernanceComplianceAction{}, nil
}

func (f *fakeGovernanceStore) DeleteGovernanceComplianceAction(ctx context.Context, id pgtype.UUID) error {
	f.deleteComplianceID = id
	return nil
}

func TestGovernanceServiceForwardsStoreCallsAndNormalizesInputs(t *testing.T) {
	id := governanceTestUUID(31)
	employeeID := governanceTestUUID(32)
	start := governanceTestDate(2026, time.January, 1)
	end := governanceTestDate(2026, time.December, 31)
	currentYear := int32(time.Now().Year())
	store := &fakeGovernanceStore{}
	svc := &Governance{q: store}
	if NewGovernance(nil) == nil {
		t.Fatal("NewGovernance(nil) = nil, want service")
	}

	if _, err := svc.Stats(context.Background()); err != nil || !store.statsCalled {
		t.Fatalf("Stats() = %v called=%v, want called", err, store.statsCalled)
	}
	if rows, err := svc.SNPMatrix(context.Background()); err != nil || len(rows) != 1 || !store.snpMatrixCalled {
		t.Fatalf("SNPMatrix() = %d rows/%v called=%v, want rows", len(rows), err, store.snpMatrixCalled)
	}
	if rows, err := svc.EmployeeOptions(context.Background()); err != nil || len(rows) != 1 || !store.employeeOptionsCalled {
		t.Fatalf("EmployeeOptions() = %d rows/%v called=%v, want rows", len(rows), err, store.employeeOptionsCalled)
	}

	if _, err := svc.ListUnits(context.Background(), " unit "); err != nil || store.unitSearch != "unit" {
		t.Fatalf("ListUnits() = %v search=%q, want trimmed", err, store.unitSearch)
	}
	if _, err := svc.CreateUnit(context.Background(), db.CreateGovernanceUnitParams{Code: " KAMAD ", Name: " Kepala ", Description: " desc "}); err != nil {
		t.Fatalf("CreateUnit() error = %v", err)
	}
	if store.createUnitArg.Code != "KAMAD" || store.createUnitArg.Name != "Kepala" || store.createUnitArg.UnitType != "madrasah" || store.createUnitArg.Description != "desc" {
		t.Fatalf("CreateUnit() arg = %+v, want normalized unit", store.createUnitArg)
	}
	if _, err := svc.UpdateUnit(context.Background(), db.UpdateGovernanceUnitParams{ID: id, Code: " TU ", Name: " Tata Usaha "}); err != nil {
		t.Fatalf("UpdateUnit() error = %v", err)
	}
	if store.updateUnitArg.ID != id || store.updateUnitArg.Code != "TU" {
		t.Fatalf("UpdateUnit() arg = %+v, want id and trimmed code", store.updateUnitArg)
	}
	if err := svc.DeleteUnit(context.Background(), id); err != nil || store.deleteUnitID != id {
		t.Fatalf("DeleteUnit() = %v id=%v, want nil/%v", err, store.deleteUnitID, id)
	}

	if _, err := svc.ListPositions(context.Background(), " waka "); err != nil || store.positionSearch != "waka" {
		t.Fatalf("ListPositions() = %v search=%q, want trimmed", err, store.positionSearch)
	}
	if _, err := svc.CreatePosition(context.Background(), db.CreateGovernancePositionParams{UnitID: id, Title: " Waka ", Description: " desc ", Tupoksi: " tugas "}); err != nil {
		t.Fatalf("CreatePosition() error = %v", err)
	}
	if store.createPositionArg.Title != "Waka" || store.createPositionArg.PositionType != "struktural" || store.createPositionArg.Tupoksi != "tugas" {
		t.Fatalf("CreatePosition() arg = %+v, want normalized position", store.createPositionArg)
	}
	if _, err := svc.UpdatePosition(context.Background(), db.UpdateGovernancePositionParams{ID: id, UnitID: id, Title: " Waka Baru "}); err != nil {
		t.Fatalf("UpdatePosition() error = %v", err)
	}
	if store.updatePositionArg.ID != id || store.updatePositionArg.Title != "Waka Baru" {
		t.Fatalf("UpdatePosition() arg = %+v, want id and trimmed title", store.updatePositionArg)
	}
	if err := svc.DeletePosition(context.Background(), id); err != nil || store.deletePositionID != id {
		t.Fatalf("DeletePosition() = %v id=%v, want nil/%v", err, store.deletePositionID, id)
	}

	if _, err := svc.ListAssignments(context.Background(), " guru ", true); err != nil || store.assignmentsArg.Search != "guru" || !store.assignmentsArg.ActiveOnly {
		t.Fatalf("ListAssignments() = %v arg=%+v, want trimmed active search", err, store.assignmentsArg)
	}
	if _, err := svc.CreateAssignment(context.Background(), db.CreateGovernanceAssignmentParams{PositionID: id, EmployeeID: employeeID, StartDate: start, EndDate: end, Notes: " catatan "}); err != nil {
		t.Fatalf("CreateAssignment() error = %v", err)
	}
	if store.createAssignmentArg.Notes != "catatan" {
		t.Fatalf("CreateAssignment() arg = %+v, want trimmed notes", store.createAssignmentArg)
	}
	if _, err := svc.UpdateAssignment(context.Background(), db.UpdateGovernanceAssignmentParams{ID: id, PositionID: id, EmployeeID: employeeID, StartDate: start, Notes: " update "}); err != nil {
		t.Fatalf("UpdateAssignment() error = %v", err)
	}
	if store.updateAssignmentArg.ID != id || store.updateAssignmentArg.Notes != "update" {
		t.Fatalf("UpdateAssignment() arg = %+v, want id and trimmed notes", store.updateAssignmentArg)
	}
	if err := svc.DeleteAssignment(context.Background(), id); err != nil || store.deleteAssignmentID != id {
		t.Fatalf("DeleteAssignment() = %v id=%v, want nil/%v", err, store.deleteAssignmentID, id)
	}

	if _, err := svc.ListDocuments(context.Background(), " rkt ", " rkt ", " SKL ", 2026); err != nil || store.documentsArg.Search != "rkt" || store.documentsArg.SnpStandard != "skl" {
		t.Fatalf("ListDocuments() = %v arg=%+v, want normalized filters", err, store.documentsArg)
	}
	if _, err := svc.CreateDocument(context.Background(), db.CreateGovernanceDocumentParams{DocType: " rkt ", Title: " RKT ", SnpStandard: " SKL ", Status: "", DocumentUrl: " /doc ", Summary: " ringkas "}); err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}
	if store.createDocumentArg.PeriodYear != currentYear || store.createDocumentArg.Status != "draft" || store.createDocumentArg.Title != "RKT" || store.createDocumentArg.SnpStandard != "skl" {
		t.Fatalf("CreateDocument() arg = %+v, want defaults and trimmed fields", store.createDocumentArg)
	}
	if _, err := svc.UpdateDocument(context.Background(), db.UpdateGovernanceDocumentParams{ID: id, DocType: " rkt ", Title: " RKT Baru ", SnpStandard: " skl ", Status: "final"}); err != nil {
		t.Fatalf("UpdateDocument() error = %v", err)
	}
	if store.updateDocumentArg.ID != id || store.updateDocumentArg.PeriodYear != currentYear || store.updateDocumentArg.Title != "RKT Baru" {
		t.Fatalf("UpdateDocument() arg = %+v, want id/default year", store.updateDocumentArg)
	}
	if err := svc.DeleteDocument(context.Background(), id); err != nil || store.deleteDocumentID != id {
		t.Fatalf("DeleteDocument() = %v id=%v, want nil/%v", err, store.deleteDocumentID, id)
	}

	if _, err := svc.ListPrograms(context.Background(), " p ", "planned", "SKL", 2026); err != nil || store.programsArg.Search != "p" || store.programsArg.SnpStandard != "skl" {
		t.Fatalf("ListPrograms() = %v arg=%+v, want normalized filters", err, store.programsArg)
	}
	if _, err := svc.CreateProgram(context.Background(), db.CreateGovernanceProgramParams{Code: " P-1 ", Name: " Program ", SnpStandard: " SKL "}); err != nil {
		t.Fatalf("CreateProgram() error = %v", err)
	}
	if store.createProgramArg.PeriodYear != currentYear || store.createProgramArg.Code != "P-1" || store.createProgramArg.Status != "planned" {
		t.Fatalf("CreateProgram() arg = %+v, want normalized program", store.createProgramArg)
	}
	if _, err := svc.UpdateProgram(context.Background(), db.UpdateGovernanceProgramParams{ID: id, Code: " P-2 ", Name: " Program 2 ", SnpStandard: " SKL "}); err != nil {
		t.Fatalf("UpdateProgram() error = %v", err)
	}
	if store.updateProgramArg.ID != id || store.updateProgramArg.Code != "P-2" {
		t.Fatalf("UpdateProgram() arg = %+v, want id and trimmed code", store.updateProgramArg)
	}
	if err := svc.DeleteProgram(context.Background(), id); err != nil || store.deleteProgramID != id {
		t.Fatalf("DeleteProgram() = %v id=%v, want nil/%v", err, store.deleteProgramID, id)
	}

	if _, err := svc.ListWorkPlanItems(context.Background(), " kegiatan ", "00000000-0000-0000-0000-000000000001", "planned", 2026); err != nil || store.workPlanArg.Search != "kegiatan" || !store.workPlanArg.ProgramID.Valid {
		t.Fatalf("ListWorkPlanItems() = %v arg=%+v, want parsed program filter", err, store.workPlanArg)
	}
	if _, err := svc.CreateWorkPlanItem(context.Background(), db.CreateGovernanceWorkPlanItemParams{ProgramID: id, ActivityCode: " A-1 ", ActivityName: " Kegiatan "}); err != nil {
		t.Fatalf("CreateWorkPlanItem() error = %v", err)
	}
	if store.createWorkPlanArg.PeriodYear != currentYear || store.createWorkPlanArg.ActivityCode != "A-1" || store.createWorkPlanArg.Status != "planned" {
		t.Fatalf("CreateWorkPlanItem() arg = %+v, want normalized work plan", store.createWorkPlanArg)
	}
	if _, err := svc.UpdateWorkPlanItem(context.Background(), db.UpdateGovernanceWorkPlanItemParams{ID: id, ProgramID: id, ActivityCode: " A-2 ", ActivityName: " Kegiatan 2 "}); err != nil {
		t.Fatalf("UpdateWorkPlanItem() error = %v", err)
	}
	if store.updateWorkPlanArg.ID != id || store.updateWorkPlanArg.ActivityCode != "A-2" {
		t.Fatalf("UpdateWorkPlanItem() arg = %+v, want id and code", store.updateWorkPlanArg)
	}
	if err := svc.DeleteWorkPlanItem(context.Background(), id); err != nil || store.deleteWorkPlanID != id {
		t.Fatalf("DeleteWorkPlanItem() = %v id=%v, want nil/%v", err, store.deleteWorkPlanID, id)
	}

	if _, err := svc.ListPerformanceTargets(context.Background(), " target ", "planned", "00000000-0000-0000-0000-000000000001", 2026); err != nil || store.targetsArg.Search != "target" || !store.targetsArg.EmployeeID.Valid {
		t.Fatalf("ListPerformanceTargets() = %v arg=%+v, want parsed employee", err, store.targetsArg)
	}
	if _, err := svc.CreatePerformanceTarget(context.Background(), db.CreateGovernancePerformanceTargetParams{EmployeeID: employeeID, Title: " Target "}); err != nil {
		t.Fatalf("CreatePerformanceTarget() error = %v", err)
	}
	if store.createTargetArg.PeriodYear != currentYear || store.createTargetArg.Aspect != "hasil_kerja" || store.createTargetArg.Title != "Target" {
		t.Fatalf("CreatePerformanceTarget() arg = %+v, want normalized target", store.createTargetArg)
	}
	if _, err := svc.UpdatePerformanceTarget(context.Background(), db.UpdateGovernancePerformanceTargetParams{ID: id, EmployeeID: employeeID, Title: " Target 2 "}); err != nil {
		t.Fatalf("UpdatePerformanceTarget() error = %v", err)
	}
	if store.updateTargetArg.ID != id || store.updateTargetArg.Title != "Target 2" {
		t.Fatalf("UpdatePerformanceTarget() arg = %+v, want id and title", store.updateTargetArg)
	}
	if err := svc.DeletePerformanceTarget(context.Background(), id); err != nil || store.deleteTargetID != id {
		t.Fatalf("DeletePerformanceTarget() = %v id=%v, want nil/%v", err, store.deleteTargetID, id)
	}

	if _, err := svc.ListEvidenceItems(context.Background(), " bukti ", "needed", "SKL", 2026); err != nil || store.evidenceArg.Search != "bukti" || store.evidenceArg.SnpStandard != "skl" {
		t.Fatalf("ListEvidenceItems() = %v arg=%+v, want normalized filters", err, store.evidenceArg)
	}
	if _, err := svc.CreateEvidenceItem(context.Background(), db.CreateGovernanceEvidenceItemParams{Title: " Bukti ", SnpStandard: " SKL "}); err != nil {
		t.Fatalf("CreateEvidenceItem() error = %v", err)
	}
	if store.createEvidenceArg.PeriodYear != currentYear || store.createEvidenceArg.Title != "Bukti" || store.createEvidenceArg.EvidenceType != "dokumen" || store.createEvidenceArg.SourceModule != "governance" {
		t.Fatalf("CreateEvidenceItem() arg = %+v, want normalized evidence", store.createEvidenceArg)
	}
	if _, err := svc.UpdateEvidenceItem(context.Background(), db.UpdateGovernanceEvidenceItemParams{ID: id, Title: " Bukti 2 ", SnpStandard: " SKL "}); err != nil {
		t.Fatalf("UpdateEvidenceItem() error = %v", err)
	}
	if store.updateEvidenceArg.ID != id || store.updateEvidenceArg.Title != "Bukti 2" {
		t.Fatalf("UpdateEvidenceItem() arg = %+v, want id and title", store.updateEvidenceArg)
	}
	if err := svc.DeleteEvidenceItem(context.Background(), id); err != nil || store.deleteEvidenceID != id {
		t.Fatalf("DeleteEvidenceItem() = %v id=%v, want nil/%v", err, store.deleteEvidenceID, id)
	}

	if _, err := svc.ListComplianceActions(context.Background(), " tindak ", "open", "medium", "manual", "SKL", "00000000-0000-0000-0000-000000000001", 2026); err != nil || store.complianceArg.Search != "tindak" || store.complianceArg.SnpStandard != "skl" || !store.complianceArg.ResponsibleEmployeeID.Valid {
		t.Fatalf("ListComplianceActions() = %v arg=%+v, want normalized filters", err, store.complianceArg)
	}
	if _, err := svc.CreateComplianceAction(context.Background(), db.CreateGovernanceComplianceActionParams{SnpStandard: " SKL ", Title: " Tindak ", Description: " desc "}); err != nil {
		t.Fatalf("CreateComplianceAction() error = %v", err)
	}
	if store.createComplianceArg.PeriodYear != currentYear || store.createComplianceArg.SourceType != "manual" || store.createComplianceArg.Priority != "medium" || store.createComplianceArg.Status != "open" {
		t.Fatalf("CreateComplianceAction() arg = %+v, want normalized compliance", store.createComplianceArg)
	}
	if _, err := svc.UpdateComplianceAction(context.Background(), db.UpdateGovernanceComplianceActionParams{ID: id, SnpStandard: " SKL ", Title: " Tindak 2 ", Status: "done"}); err != nil {
		t.Fatalf("UpdateComplianceAction() error = %v", err)
	}
	if store.updateComplianceArg.ID != id || store.updateComplianceArg.Title != "Tindak 2" || !store.updateComplianceArg.CompletedAt.Valid {
		t.Fatalf("UpdateComplianceAction() arg = %+v, want id/title/completed timestamp", store.updateComplianceArg)
	}
	if err := svc.DeleteComplianceAction(context.Background(), id); err != nil || store.deleteComplianceID != id {
		t.Fatalf("DeleteComplianceAction() = %v id=%v, want nil/%v", err, store.deleteComplianceID, id)
	}
}

func TestGovernanceServiceListFilterValidation(t *testing.T) {
	svc := &Governance{q: &fakeGovernanceStore{}}
	if _, err := svc.ListWorkPlanItems(context.Background(), "", "bad", "", 2026); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListWorkPlanItems(bad id) = %v, want id tidak valid", err)
	}
	if _, err := svc.ListWorkPlanItems(context.Background(), "", "", "bad", 2026); err == nil || err.Error() != "status item RKT/RKJM tidak valid" {
		t.Fatalf("ListWorkPlanItems(bad status) = %v, want status error", err)
	}
	if _, err := svc.ListPerformanceTargets(context.Background(), "", "", "bad", 2026); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListPerformanceTargets(bad id) = %v, want id tidak valid", err)
	}
	if _, err := svc.ListEvidenceItems(context.Background(), "", "bad", "skl", 2026); err == nil || err.Error() != "status bukti mutu tidak valid" {
		t.Fatalf("ListEvidenceItems(bad status) = %v, want status error", err)
	}
	if _, err := svc.ListEvidenceItems(context.Background(), "", "", "bad", 2026); err == nil || err.Error() != "standar SNP tidak valid" {
		t.Fatalf("ListEvidenceItems(bad snp) = %v, want snp error", err)
	}
	if _, err := svc.ListComplianceActions(context.Background(), "", "", "", "", "skl", "bad", 2026); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ListComplianceActions(bad employee id) = %v, want id tidak valid", err)
	}
	if _, err := svc.ListComplianceActions(context.Background(), "", "bad", "", "", "skl", "", 2026); err == nil || err.Error() != "status tindak lanjut tidak valid" {
		t.Fatalf("ListComplianceActions(bad status) = %v, want status error", err)
	}
	if _, err := svc.ListComplianceActions(context.Background(), "", "", "bad", "", "skl", "", 2026); err == nil || err.Error() != "prioritas tindak lanjut tidak valid" {
		t.Fatalf("ListComplianceActions(bad priority) = %v, want priority error", err)
	}
	if _, err := svc.ListComplianceActions(context.Background(), "", "", "", "bad", "skl", "", 2026); err == nil || err.Error() != "sumber tindak lanjut tidak valid" {
		t.Fatalf("ListComplianceActions(bad source) = %v, want source error", err)
	}
	if _, err := svc.ListComplianceActions(context.Background(), "", "", "", "", "bad", "", 2026); err == nil || err.Error() != "standar SNP tidak valid" {
		t.Fatalf("ListComplianceActions(bad snp) = %v, want snp error", err)
	}
}
