package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type governanceStore interface {
	GetGovernanceStats(ctx context.Context) (db.GetGovernanceStatsRow, error)
	ListGovernanceSNPMatrix(ctx context.Context) ([]db.ListGovernanceSNPMatrixRow, error)
	ListGovernanceEmployeeOptions(ctx context.Context) ([]db.ListGovernanceEmployeeOptionsRow, error)

	ListGovernanceUnits(ctx context.Context, search string) ([]db.ListGovernanceUnitsRow, error)
	CreateGovernanceUnit(ctx context.Context, arg db.CreateGovernanceUnitParams) (db.GovernanceUnit, error)
	UpdateGovernanceUnit(ctx context.Context, arg db.UpdateGovernanceUnitParams) (db.GovernanceUnit, error)
	DeleteGovernanceUnit(ctx context.Context, id pgtype.UUID) error

	ListGovernancePositions(ctx context.Context, search string) ([]db.ListGovernancePositionsRow, error)
	CreateGovernancePosition(ctx context.Context, arg db.CreateGovernancePositionParams) (db.GovernancePosition, error)
	UpdateGovernancePosition(ctx context.Context, arg db.UpdateGovernancePositionParams) (db.GovernancePosition, error)
	DeleteGovernancePosition(ctx context.Context, id pgtype.UUID) error

	ListGovernanceAssignments(ctx context.Context, arg db.ListGovernanceAssignmentsParams) ([]db.ListGovernanceAssignmentsRow, error)
	CreateGovernanceAssignment(ctx context.Context, arg db.CreateGovernanceAssignmentParams) (db.GovernanceAssignment, error)
	UpdateGovernanceAssignment(ctx context.Context, arg db.UpdateGovernanceAssignmentParams) (db.GovernanceAssignment, error)
	DeleteGovernanceAssignment(ctx context.Context, id pgtype.UUID) error

	ListGovernanceDocuments(ctx context.Context, arg db.ListGovernanceDocumentsParams) ([]db.ListGovernanceDocumentsRow, error)
	CreateGovernanceDocument(ctx context.Context, arg db.CreateGovernanceDocumentParams) (db.GovernanceDocument, error)
	UpdateGovernanceDocument(ctx context.Context, arg db.UpdateGovernanceDocumentParams) (db.GovernanceDocument, error)
	DeleteGovernanceDocument(ctx context.Context, id pgtype.UUID) error

	ListGovernancePrograms(ctx context.Context, arg db.ListGovernanceProgramsParams) ([]db.ListGovernanceProgramsRow, error)
	CreateGovernanceProgram(ctx context.Context, arg db.CreateGovernanceProgramParams) (db.GovernanceProgram, error)
	UpdateGovernanceProgram(ctx context.Context, arg db.UpdateGovernanceProgramParams) (db.GovernanceProgram, error)
	DeleteGovernanceProgram(ctx context.Context, id pgtype.UUID) error

	ListGovernanceWorkPlanItems(ctx context.Context, arg db.ListGovernanceWorkPlanItemsParams) ([]db.ListGovernanceWorkPlanItemsRow, error)
	CreateGovernanceWorkPlanItem(ctx context.Context, arg db.CreateGovernanceWorkPlanItemParams) (db.GovernanceWorkPlanItem, error)
	UpdateGovernanceWorkPlanItem(ctx context.Context, arg db.UpdateGovernanceWorkPlanItemParams) (db.GovernanceWorkPlanItem, error)
	DeleteGovernanceWorkPlanItem(ctx context.Context, id pgtype.UUID) error

	ListGovernancePerformanceTargets(ctx context.Context, arg db.ListGovernancePerformanceTargetsParams) ([]db.ListGovernancePerformanceTargetsRow, error)
	CreateGovernancePerformanceTarget(ctx context.Context, arg db.CreateGovernancePerformanceTargetParams) (db.GovernancePerformanceTarget, error)
	UpdateGovernancePerformanceTarget(ctx context.Context, arg db.UpdateGovernancePerformanceTargetParams) (db.GovernancePerformanceTarget, error)
	DeleteGovernancePerformanceTarget(ctx context.Context, id pgtype.UUID) error

	ListGovernanceEvidenceItems(ctx context.Context, arg db.ListGovernanceEvidenceItemsParams) ([]db.ListGovernanceEvidenceItemsRow, error)
	CreateGovernanceEvidenceItem(ctx context.Context, arg db.CreateGovernanceEvidenceItemParams) (db.GovernanceEvidenceItem, error)
	UpdateGovernanceEvidenceItem(ctx context.Context, arg db.UpdateGovernanceEvidenceItemParams) (db.GovernanceEvidenceItem, error)
	DeleteGovernanceEvidenceItem(ctx context.Context, id pgtype.UUID) error

	ListGovernanceComplianceActions(ctx context.Context, arg db.ListGovernanceComplianceActionsParams) ([]db.ListGovernanceComplianceActionsRow, error)
	CreateGovernanceComplianceAction(ctx context.Context, arg db.CreateGovernanceComplianceActionParams) (db.GovernanceComplianceAction, error)
	UpdateGovernanceComplianceAction(ctx context.Context, arg db.UpdateGovernanceComplianceActionParams) (db.GovernanceComplianceAction, error)
	DeleteGovernanceComplianceAction(ctx context.Context, id pgtype.UUID) error
}

type Governance struct{ q governanceStore }

func NewGovernance(q *db.Queries) *Governance { return &Governance{q: q} }

func (s *Governance) Stats(ctx context.Context) (db.GetGovernanceStatsRow, error) {
	return s.q.GetGovernanceStats(ctx)
}

func (s *Governance) SNPMatrix(ctx context.Context) ([]db.ListGovernanceSNPMatrixRow, error) {
	return s.q.ListGovernanceSNPMatrix(ctx)
}

func (s *Governance) EmployeeOptions(ctx context.Context) ([]db.ListGovernanceEmployeeOptionsRow, error) {
	return s.q.ListGovernanceEmployeeOptions(ctx)
}

func (s *Governance) ListUnits(ctx context.Context, search string) ([]db.ListGovernanceUnitsRow, error) {
	return s.q.ListGovernanceUnits(ctx, strings.TrimSpace(search))
}

func (s *Governance) CreateUnit(ctx context.Context, arg db.CreateGovernanceUnitParams) (db.GovernanceUnit, error) {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.UnitType = normalizeGovernanceText(arg.UnitType, "madrasah")
	arg.Description = strings.TrimSpace(arg.Description)
	if err := validateGovernanceUnit(arg.Code, arg.Name); err != nil {
		return db.GovernanceUnit{}, err
	}
	return s.q.CreateGovernanceUnit(ctx, arg)
}

func (s *Governance) UpdateUnit(ctx context.Context, arg db.UpdateGovernanceUnitParams) (db.GovernanceUnit, error) {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.UnitType = normalizeGovernanceText(arg.UnitType, "madrasah")
	arg.Description = strings.TrimSpace(arg.Description)
	if err := validateGovernanceUnit(arg.Code, arg.Name); err != nil {
		return db.GovernanceUnit{}, err
	}
	return s.q.UpdateGovernanceUnit(ctx, arg)
}

func (s *Governance) DeleteUnit(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernanceUnit(ctx, id)
}

func (s *Governance) ListPositions(ctx context.Context, search string) ([]db.ListGovernancePositionsRow, error) {
	return s.q.ListGovernancePositions(ctx, strings.TrimSpace(search))
}

func (s *Governance) CreatePosition(ctx context.Context, arg db.CreateGovernancePositionParams) (db.GovernancePosition, error) {
	arg.Title = strings.TrimSpace(arg.Title)
	arg.PositionType = normalizeGovernanceText(arg.PositionType, "struktural")
	arg.Description = strings.TrimSpace(arg.Description)
	arg.Tupoksi = strings.TrimSpace(arg.Tupoksi)
	if err := validateGovernancePosition(arg); err != nil {
		return db.GovernancePosition{}, err
	}
	return s.q.CreateGovernancePosition(ctx, arg)
}

func (s *Governance) UpdatePosition(ctx context.Context, arg db.UpdateGovernancePositionParams) (db.GovernancePosition, error) {
	arg.Title = strings.TrimSpace(arg.Title)
	arg.PositionType = normalizeGovernanceText(arg.PositionType, "struktural")
	arg.Description = strings.TrimSpace(arg.Description)
	arg.Tupoksi = strings.TrimSpace(arg.Tupoksi)
	if err := validateGovernancePosition(db.CreateGovernancePositionParams{
		UnitID:           arg.UnitID,
		Title:            arg.Title,
		PositionType:     arg.PositionType,
		ParentPositionID: arg.ParentPositionID,
		Description:      arg.Description,
		Tupoksi:          arg.Tupoksi,
		IsActive:         arg.IsActive,
		SortOrder:        arg.SortOrder,
	}); err != nil {
		return db.GovernancePosition{}, err
	}
	return s.q.UpdateGovernancePosition(ctx, arg)
}

func (s *Governance) DeletePosition(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernancePosition(ctx, id)
}

func (s *Governance) ListAssignments(ctx context.Context, search string, activeOnly bool) ([]db.ListGovernanceAssignmentsRow, error) {
	return s.q.ListGovernanceAssignments(ctx, db.ListGovernanceAssignmentsParams{
		Search:     strings.TrimSpace(search),
		ActiveOnly: activeOnly,
	})
}

func (s *Governance) CreateAssignment(ctx context.Context, arg db.CreateGovernanceAssignmentParams) (db.GovernanceAssignment, error) {
	arg.Notes = strings.TrimSpace(arg.Notes)
	if err := validateGovernanceAssignment(arg.PositionID, arg.EmployeeID, arg.StartDate, arg.EndDate); err != nil {
		return db.GovernanceAssignment{}, err
	}
	return s.q.CreateGovernanceAssignment(ctx, arg)
}

func (s *Governance) UpdateAssignment(ctx context.Context, arg db.UpdateGovernanceAssignmentParams) (db.GovernanceAssignment, error) {
	arg.Notes = strings.TrimSpace(arg.Notes)
	if err := validateGovernanceAssignment(arg.PositionID, arg.EmployeeID, arg.StartDate, arg.EndDate); err != nil {
		return db.GovernanceAssignment{}, err
	}
	return s.q.UpdateGovernanceAssignment(ctx, arg)
}

func (s *Governance) DeleteAssignment(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernanceAssignment(ctx, id)
}

func (s *Governance) ListDocuments(ctx context.Context, search, docType, snpStandard string, periodYear int32) ([]db.ListGovernanceDocumentsRow, error) {
	return s.q.ListGovernanceDocuments(ctx, db.ListGovernanceDocumentsParams{
		Search:      strings.TrimSpace(search),
		DocType:     strings.TrimSpace(docType),
		PeriodYear:  periodYear,
		SnpStandard: normalizeSNPStandard(snpStandard),
	})
}

func (s *Governance) CreateDocument(ctx context.Context, arg db.CreateGovernanceDocumentParams) (db.GovernanceDocument, error) {
	arg.DocType = normalizeGovernanceText(arg.DocType, "lainnya")
	arg.Title = strings.TrimSpace(arg.Title)
	arg.PeriodLabel = strings.TrimSpace(arg.PeriodLabel)
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.Status = normalizeGovernanceText(arg.Status, "draft")
	arg.DocumentUrl = strings.TrimSpace(arg.DocumentUrl)
	arg.Summary = strings.TrimSpace(arg.Summary)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	if err := validateGovernanceDocument(arg.DocType, arg.Title, arg.PeriodYear, arg.Status, arg.SnpStandard); err != nil {
		return db.GovernanceDocument{}, err
	}
	return s.q.CreateGovernanceDocument(ctx, arg)
}

func (s *Governance) UpdateDocument(ctx context.Context, arg db.UpdateGovernanceDocumentParams) (db.GovernanceDocument, error) {
	arg.DocType = normalizeGovernanceText(arg.DocType, "lainnya")
	arg.Title = strings.TrimSpace(arg.Title)
	arg.PeriodLabel = strings.TrimSpace(arg.PeriodLabel)
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.Status = normalizeGovernanceText(arg.Status, "draft")
	arg.DocumentUrl = strings.TrimSpace(arg.DocumentUrl)
	arg.Summary = strings.TrimSpace(arg.Summary)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	if err := validateGovernanceDocument(arg.DocType, arg.Title, arg.PeriodYear, arg.Status, arg.SnpStandard); err != nil {
		return db.GovernanceDocument{}, err
	}
	return s.q.UpdateGovernanceDocument(ctx, arg)
}

func (s *Governance) DeleteDocument(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernanceDocument(ctx, id)
}

func (s *Governance) ListPrograms(ctx context.Context, search, status, snpStandard string, periodYear int32) ([]db.ListGovernanceProgramsRow, error) {
	return s.q.ListGovernancePrograms(ctx, db.ListGovernanceProgramsParams{
		Search:      strings.TrimSpace(search),
		PeriodYear:  periodYear,
		Status:      strings.TrimSpace(status),
		SnpStandard: normalizeSNPStandard(snpStandard),
	})
}

func (s *Governance) CreateProgram(ctx context.Context, arg db.CreateGovernanceProgramParams) (db.GovernanceProgram, error) {
	arg = normalizeGovernanceProgramCreate(arg)
	if err := validateGovernanceProgram(arg.PeriodYear, arg.Code, arg.Name, arg.Status, arg.ProgressPercent, arg.SnpStandard); err != nil {
		return db.GovernanceProgram{}, err
	}
	return s.q.CreateGovernanceProgram(ctx, arg)
}

func (s *Governance) UpdateProgram(ctx context.Context, arg db.UpdateGovernanceProgramParams) (db.GovernanceProgram, error) {
	arg = normalizeGovernanceProgramUpdate(arg)
	if err := validateGovernanceProgram(arg.PeriodYear, arg.Code, arg.Name, arg.Status, arg.ProgressPercent, arg.SnpStandard); err != nil {
		return db.GovernanceProgram{}, err
	}
	return s.q.UpdateGovernanceProgram(ctx, arg)
}

func (s *Governance) DeleteProgram(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernanceProgram(ctx, id)
}

func (s *Governance) ListWorkPlanItems(ctx context.Context, search, programID, status string, periodYear int32) ([]db.ListGovernanceWorkPlanItemsRow, error) {
	parsedProgramID, err := ParseGovernanceOptionalUUID(programID)
	if err != nil {
		return nil, err
	}
	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus != "" && !validGovernanceProgramStatuses[normalizedStatus] {
		return nil, fmt.Errorf("status item RKT/RKJM tidak valid")
	}
	return s.q.ListGovernanceWorkPlanItems(ctx, db.ListGovernanceWorkPlanItemsParams{
		Search:     strings.TrimSpace(search),
		PeriodYear: periodYear,
		ProgramID:  parsedProgramID,
		Status:     normalizedStatus,
	})
}

func (s *Governance) CreateWorkPlanItem(ctx context.Context, arg db.CreateGovernanceWorkPlanItemParams) (db.GovernanceWorkPlanItem, error) {
	arg = normalizeGovernanceWorkPlanItemCreate(arg)
	if err := validateGovernanceWorkPlanItem(arg.PeriodYear, arg.ProgramID, arg.ActivityCode, arg.ActivityName, arg.Status, arg.ProgressPercent, arg.BudgetAmount, arg.RealizationAmount, arg.StartDate, arg.EndDate); err != nil {
		return db.GovernanceWorkPlanItem{}, err
	}
	return s.q.CreateGovernanceWorkPlanItem(ctx, arg)
}

func (s *Governance) UpdateWorkPlanItem(ctx context.Context, arg db.UpdateGovernanceWorkPlanItemParams) (db.GovernanceWorkPlanItem, error) {
	arg = normalizeGovernanceWorkPlanItemUpdate(arg)
	if err := validateGovernanceWorkPlanItem(arg.PeriodYear, arg.ProgramID, arg.ActivityCode, arg.ActivityName, arg.Status, arg.ProgressPercent, arg.BudgetAmount, arg.RealizationAmount, arg.StartDate, arg.EndDate); err != nil {
		return db.GovernanceWorkPlanItem{}, err
	}
	return s.q.UpdateGovernanceWorkPlanItem(ctx, arg)
}

func (s *Governance) DeleteWorkPlanItem(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernanceWorkPlanItem(ctx, id)
}

func (s *Governance) ListPerformanceTargets(ctx context.Context, search, status, employeeID string, periodYear int32) ([]db.ListGovernancePerformanceTargetsRow, error) {
	parsedEmployeeID, err := ParseGovernanceOptionalUUID(employeeID)
	if err != nil {
		return nil, err
	}
	return s.q.ListGovernancePerformanceTargets(ctx, db.ListGovernancePerformanceTargetsParams{
		Search:     strings.TrimSpace(search),
		PeriodYear: periodYear,
		Status:     strings.TrimSpace(status),
		EmployeeID: parsedEmployeeID,
	})
}

func (s *Governance) CreatePerformanceTarget(ctx context.Context, arg db.CreateGovernancePerformanceTargetParams) (db.GovernancePerformanceTarget, error) {
	arg = normalizeGovernancePerformanceTargetCreate(arg)
	if err := validateGovernancePerformanceTarget(arg.PeriodYear, arg.EmployeeID, arg.Aspect, arg.Title, arg.Status, arg.ProgressPercent); err != nil {
		return db.GovernancePerformanceTarget{}, err
	}
	return s.q.CreateGovernancePerformanceTarget(ctx, arg)
}

func (s *Governance) UpdatePerformanceTarget(ctx context.Context, arg db.UpdateGovernancePerformanceTargetParams) (db.GovernancePerformanceTarget, error) {
	arg = normalizeGovernancePerformanceTargetUpdate(arg)
	if err := validateGovernancePerformanceTarget(arg.PeriodYear, arg.EmployeeID, arg.Aspect, arg.Title, arg.Status, arg.ProgressPercent); err != nil {
		return db.GovernancePerformanceTarget{}, err
	}
	return s.q.UpdateGovernancePerformanceTarget(ctx, arg)
}

func (s *Governance) DeletePerformanceTarget(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernancePerformanceTarget(ctx, id)
}

func (s *Governance) ListEvidenceItems(ctx context.Context, search, status, snpStandard string, periodYear int32) ([]db.ListGovernanceEvidenceItemsRow, error) {
	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus != "" && !validGovernanceEvidenceStatuses[normalizedStatus] {
		return nil, fmt.Errorf("status bukti mutu tidak valid")
	}
	normalizedSNP := normalizeSNPStandard(snpStandard)
	if !validGovernanceSNPStandards[normalizedSNP] {
		return nil, fmt.Errorf("standar SNP tidak valid")
	}
	return s.q.ListGovernanceEvidenceItems(ctx, db.ListGovernanceEvidenceItemsParams{
		Search:      strings.TrimSpace(search),
		PeriodYear:  periodYear,
		Status:      normalizedStatus,
		SnpStandard: normalizedSNP,
	})
}

func (s *Governance) CreateEvidenceItem(ctx context.Context, arg db.CreateGovernanceEvidenceItemParams) (db.GovernanceEvidenceItem, error) {
	arg = normalizeGovernanceEvidenceItemCreate(arg)
	if err := validateGovernanceEvidenceItem(arg.PeriodYear, arg.Title, arg.EvidenceType, arg.Status, arg.SnpStandard, arg.SourceModule); err != nil {
		return db.GovernanceEvidenceItem{}, err
	}
	return s.q.CreateGovernanceEvidenceItem(ctx, arg)
}

func (s *Governance) UpdateEvidenceItem(ctx context.Context, arg db.UpdateGovernanceEvidenceItemParams) (db.GovernanceEvidenceItem, error) {
	arg = normalizeGovernanceEvidenceItemUpdate(arg)
	if err := validateGovernanceEvidenceItem(arg.PeriodYear, arg.Title, arg.EvidenceType, arg.Status, arg.SnpStandard, arg.SourceModule); err != nil {
		return db.GovernanceEvidenceItem{}, err
	}
	return s.q.UpdateGovernanceEvidenceItem(ctx, arg)
}

func (s *Governance) DeleteEvidenceItem(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernanceEvidenceItem(ctx, id)
}

func (s *Governance) ListComplianceActions(ctx context.Context, search, status, priority, sourceType, snpStandard, responsibleEmployeeID string, periodYear int32) ([]db.ListGovernanceComplianceActionsRow, error) {
	parsedEmployeeID, err := ParseGovernanceOptionalUUID(responsibleEmployeeID)
	if err != nil {
		return nil, err
	}
	normalizedStatus := strings.TrimSpace(status)
	if normalizedStatus != "" && !validGovernanceComplianceActionStatuses[normalizedStatus] {
		return nil, fmt.Errorf("status tindak lanjut tidak valid")
	}
	normalizedPriority := strings.TrimSpace(priority)
	if normalizedPriority != "" && !validGovernanceComplianceActionPriorities[normalizedPriority] {
		return nil, fmt.Errorf("prioritas tindak lanjut tidak valid")
	}
	normalizedSourceType := strings.TrimSpace(sourceType)
	if normalizedSourceType != "" && !validGovernanceComplianceActionSources[normalizedSourceType] {
		return nil, fmt.Errorf("sumber tindak lanjut tidak valid")
	}
	normalizedSNP := normalizeSNPStandard(snpStandard)
	if !validGovernanceSNPStandards[normalizedSNP] {
		return nil, fmt.Errorf("standar SNP tidak valid")
	}
	return s.q.ListGovernanceComplianceActions(ctx, db.ListGovernanceComplianceActionsParams{
		Search:                strings.TrimSpace(search),
		PeriodYear:            periodYear,
		Status:                normalizedStatus,
		Priority:              normalizedPriority,
		SourceType:            normalizedSourceType,
		SnpStandard:           normalizedSNP,
		ResponsibleEmployeeID: parsedEmployeeID,
	})
}

func (s *Governance) CreateComplianceAction(ctx context.Context, arg db.CreateGovernanceComplianceActionParams) (db.GovernanceComplianceAction, error) {
	arg = normalizeGovernanceComplianceActionCreate(arg)
	if err := validateGovernanceComplianceAction(arg.PeriodYear, arg.SourceType, arg.SnpStandard, arg.Title, arg.Priority, arg.Status); err != nil {
		return db.GovernanceComplianceAction{}, err
	}
	return s.q.CreateGovernanceComplianceAction(ctx, arg)
}

func (s *Governance) UpdateComplianceAction(ctx context.Context, arg db.UpdateGovernanceComplianceActionParams) (db.GovernanceComplianceAction, error) {
	arg = normalizeGovernanceComplianceActionUpdate(arg)
	if err := validateGovernanceComplianceAction(arg.PeriodYear, arg.SourceType, arg.SnpStandard, arg.Title, arg.Priority, arg.Status); err != nil {
		return db.GovernanceComplianceAction{}, err
	}
	return s.q.UpdateGovernanceComplianceAction(ctx, arg)
}

func (s *Governance) DeleteComplianceAction(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGovernanceComplianceAction(ctx, id)
}

func ParseGovernanceOptionalUUID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if strings.TrimSpace(raw) == "" {
		return id, nil
	}
	if err := id.Scan(strings.TrimSpace(raw)); err != nil {
		return pgtype.UUID{}, fmt.Errorf("id tidak valid")
	}
	return id, nil
}

func ParseGovernanceDate(raw string) (pgtype.Date, error) {
	var date pgtype.Date
	if strings.TrimSpace(raw) == "" {
		return date, fmt.Errorf("tanggal wajib diisi")
	}
	if err := date.Scan(strings.TrimSpace(raw)); err != nil {
		return pgtype.Date{}, fmt.Errorf("format tanggal tidak valid")
	}
	return date, nil
}

func ParseGovernanceOptionalDate(raw string) (pgtype.Date, error) {
	var date pgtype.Date
	if strings.TrimSpace(raw) == "" {
		return date, nil
	}
	if err := date.Scan(strings.TrimSpace(raw)); err != nil {
		return pgtype.Date{}, fmt.Errorf("format tanggal tidak valid")
	}
	return date, nil
}

func ParseGovernanceOptionalTimestamp(raw string) (pgtype.Timestamptz, error) {
	var timestamp pgtype.Timestamptz
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return timestamp, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		date, dateErr := time.Parse("2006-01-02", trimmed)
		if dateErr != nil {
			return pgtype.Timestamptz{}, fmt.Errorf("format waktu tidak valid")
		}
		parsed = date
	}
	return pgtype.Timestamptz{Time: parsed, Valid: true}, nil
}

func normalizeGovernanceText(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func validateGovernanceUnit(code, name string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("kode unit wajib diisi")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nama unit wajib diisi")
	}
	return nil
}

func validateGovernancePosition(arg db.CreateGovernancePositionParams) error {
	if !arg.UnitID.Valid {
		return fmt.Errorf("unit kerja wajib dipilih")
	}
	if strings.TrimSpace(arg.Title) == "" {
		return fmt.Errorf("nama jabatan wajib diisi")
	}
	return nil
}

func validateGovernanceAssignment(positionID, employeeID pgtype.UUID, startDate, endDate pgtype.Date) error {
	if !positionID.Valid {
		return fmt.Errorf("jabatan wajib dipilih")
	}
	if !employeeID.Valid {
		return fmt.Errorf("pegawai wajib dipilih")
	}
	if !startDate.Valid {
		return fmt.Errorf("tanggal mulai wajib diisi")
	}
	if endDate.Valid && endDate.Time.Before(startDate.Time) {
		return fmt.Errorf("tanggal selesai tidak boleh sebelum tanggal mulai")
	}
	return nil
}

func validateGovernanceDocument(docType, title string, year int32, status, snpStandard string) error {
	if !validGovernanceDocumentTypes[docType] {
		return fmt.Errorf("jenis dokumen tidak valid")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("judul dokumen wajib diisi")
	}
	if year < 2000 {
		return fmt.Errorf("tahun periode tidak valid")
	}
	if !validGovernanceDocumentStatuses[status] {
		return fmt.Errorf("status dokumen tidak valid")
	}
	if !validGovernanceSNPStandards[snpStandard] {
		return fmt.Errorf("standar SNP tidak valid")
	}
	return nil
}

func validateGovernanceProgram(year int32, code, name, status string, progress int32, snpStandard string) error {
	if year < 2000 {
		return fmt.Errorf("tahun program tidak valid")
	}
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("kode program wajib diisi")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nama program wajib diisi")
	}
	if !validGovernanceProgramStatuses[status] {
		return fmt.Errorf("status program tidak valid")
	}
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progres program harus 0 sampai 100")
	}
	if !validGovernanceSNPStandards[snpStandard] {
		return fmt.Errorf("standar SNP tidak valid")
	}
	return nil
}

func validateGovernancePerformanceTarget(year int32, employeeID pgtype.UUID, aspect, title, status string, progress int32) error {
	if year < 2000 {
		return fmt.Errorf("tahun target kinerja tidak valid")
	}
	if !employeeID.Valid {
		return fmt.Errorf("pegawai wajib dipilih")
	}
	if !validGovernancePerformanceAspects[aspect] {
		return fmt.Errorf("aspek target kinerja tidak valid")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("target kerja wajib diisi")
	}
	if !validGovernanceProgramStatuses[status] {
		return fmt.Errorf("status target kinerja tidak valid")
	}
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progres target kinerja harus 0 sampai 100")
	}
	return nil
}

func validateGovernanceWorkPlanItem(year int32, programID pgtype.UUID, code, name, status string, progress int32, budgetAmount, realizationAmount int64, startDate, endDate pgtype.Date) error {
	if year < 2000 {
		return fmt.Errorf("tahun item RKT/RKJM tidak valid")
	}
	if !programID.Valid {
		return fmt.Errorf("program wajib dipilih")
	}
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("kode kegiatan wajib diisi")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nama kegiatan wajib diisi")
	}
	if !validGovernanceProgramStatuses[status] {
		return fmt.Errorf("status item RKT/RKJM tidak valid")
	}
	if progress < 0 || progress > 100 {
		return fmt.Errorf("progres item RKT/RKJM harus 0 sampai 100")
	}
	if budgetAmount < 0 {
		return fmt.Errorf("anggaran tidak boleh negatif")
	}
	if realizationAmount < 0 {
		return fmt.Errorf("realisasi anggaran tidak boleh negatif")
	}
	if startDate.Valid && endDate.Valid && endDate.Time.Before(startDate.Time) {
		return fmt.Errorf("tanggal selesai tidak boleh sebelum tanggal mulai")
	}
	return nil
}

func validateGovernanceEvidenceItem(year int32, title, evidenceType, status, snpStandard, sourceModule string) error {
	if year < 2000 {
		return fmt.Errorf("tahun bukti mutu tidak valid")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("judul bukti mutu wajib diisi")
	}
	if !validGovernanceEvidenceTypes[evidenceType] {
		return fmt.Errorf("jenis bukti mutu tidak valid")
	}
	if !validGovernanceEvidenceStatuses[status] {
		return fmt.Errorf("status bukti mutu tidak valid")
	}
	if !validGovernanceSNPStandards[snpStandard] {
		return fmt.Errorf("standar SNP tidak valid")
	}
	if strings.TrimSpace(sourceModule) == "" {
		return fmt.Errorf("sumber bukti mutu wajib diisi")
	}
	return nil
}

func validateGovernanceComplianceAction(year int32, sourceType, snpStandard, title, priority, status string) error {
	if year < 2000 {
		return fmt.Errorf("tahun tindak lanjut tidak valid")
	}
	if !validGovernanceComplianceActionSources[sourceType] {
		return fmt.Errorf("sumber tindak lanjut tidak valid")
	}
	if !validGovernanceSNPStandards[snpStandard] {
		return fmt.Errorf("standar SNP tidak valid")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("judul tindak lanjut wajib diisi")
	}
	if !validGovernanceComplianceActionPriorities[priority] {
		return fmt.Errorf("prioritas tindak lanjut tidak valid")
	}
	if !validGovernanceComplianceActionStatuses[status] {
		return fmt.Errorf("status tindak lanjut tidak valid")
	}
	return nil
}

func normalizeSNPStandard(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeGovernanceProgramCreate(arg db.CreateGovernanceProgramParams) db.CreateGovernanceProgramParams {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.IkuCode = strings.TrimSpace(arg.IkuCode)
	arg.Indicator = strings.TrimSpace(arg.Indicator)
	arg.TargetValue = strings.TrimSpace(arg.TargetValue)
	arg.TargetUnit = strings.TrimSpace(arg.TargetUnit)
	arg.Status = normalizeGovernanceText(arg.Status, "planned")
	arg.RealizationSummary = strings.TrimSpace(arg.RealizationSummary)
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	return arg
}

func normalizeGovernanceProgramUpdate(arg db.UpdateGovernanceProgramParams) db.UpdateGovernanceProgramParams {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.IkuCode = strings.TrimSpace(arg.IkuCode)
	arg.Indicator = strings.TrimSpace(arg.Indicator)
	arg.TargetValue = strings.TrimSpace(arg.TargetValue)
	arg.TargetUnit = strings.TrimSpace(arg.TargetUnit)
	arg.Status = normalizeGovernanceText(arg.Status, "planned")
	arg.RealizationSummary = strings.TrimSpace(arg.RealizationSummary)
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	return arg
}

func normalizeGovernanceWorkPlanItemCreate(arg db.CreateGovernanceWorkPlanItemParams) db.CreateGovernanceWorkPlanItemParams {
	arg.ActivityCode = strings.TrimSpace(arg.ActivityCode)
	arg.ActivityName = strings.TrimSpace(arg.ActivityName)
	arg.OutputIndicator = strings.TrimSpace(arg.OutputIndicator)
	arg.TargetVolume = strings.TrimSpace(arg.TargetVolume)
	arg.TargetUnit = strings.TrimSpace(arg.TargetUnit)
	arg.BudgetSource = strings.TrimSpace(arg.BudgetSource)
	arg.Status = normalizeGovernanceText(arg.Status, "planned")
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	arg.Notes = strings.TrimSpace(arg.Notes)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	return arg
}

func normalizeGovernanceWorkPlanItemUpdate(arg db.UpdateGovernanceWorkPlanItemParams) db.UpdateGovernanceWorkPlanItemParams {
	arg.ActivityCode = strings.TrimSpace(arg.ActivityCode)
	arg.ActivityName = strings.TrimSpace(arg.ActivityName)
	arg.OutputIndicator = strings.TrimSpace(arg.OutputIndicator)
	arg.TargetVolume = strings.TrimSpace(arg.TargetVolume)
	arg.TargetUnit = strings.TrimSpace(arg.TargetUnit)
	arg.BudgetSource = strings.TrimSpace(arg.BudgetSource)
	arg.Status = normalizeGovernanceText(arg.Status, "planned")
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	arg.Notes = strings.TrimSpace(arg.Notes)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	return arg
}

func normalizeGovernancePerformanceTargetCreate(arg db.CreateGovernancePerformanceTargetParams) db.CreateGovernancePerformanceTargetParams {
	arg.Aspect = normalizeGovernanceText(arg.Aspect, "hasil_kerja")
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Indicator = strings.TrimSpace(arg.Indicator)
	arg.TargetValue = strings.TrimSpace(arg.TargetValue)
	arg.TargetUnit = strings.TrimSpace(arg.TargetUnit)
	arg.Status = normalizeGovernanceText(arg.Status, "planned")
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	arg.ReviewNotes = strings.TrimSpace(arg.ReviewNotes)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	return arg
}

func normalizeGovernancePerformanceTargetUpdate(arg db.UpdateGovernancePerformanceTargetParams) db.UpdateGovernancePerformanceTargetParams {
	arg.Aspect = normalizeGovernanceText(arg.Aspect, "hasil_kerja")
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Indicator = strings.TrimSpace(arg.Indicator)
	arg.TargetValue = strings.TrimSpace(arg.TargetValue)
	arg.TargetUnit = strings.TrimSpace(arg.TargetUnit)
	arg.Status = normalizeGovernanceText(arg.Status, "planned")
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	arg.ReviewNotes = strings.TrimSpace(arg.ReviewNotes)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	return arg
}

func normalizeGovernanceEvidenceItemCreate(arg db.CreateGovernanceEvidenceItemParams) db.CreateGovernanceEvidenceItemParams {
	arg.Title = strings.TrimSpace(arg.Title)
	arg.EvidenceType = normalizeGovernanceText(arg.EvidenceType, "dokumen")
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.SourceModule = normalizeGovernanceText(arg.SourceModule, "governance")
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	arg.Status = normalizeGovernanceText(arg.Status, "needed")
	arg.Notes = strings.TrimSpace(arg.Notes)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	return arg
}

func normalizeGovernanceEvidenceItemUpdate(arg db.UpdateGovernanceEvidenceItemParams) db.UpdateGovernanceEvidenceItemParams {
	arg.Title = strings.TrimSpace(arg.Title)
	arg.EvidenceType = normalizeGovernanceText(arg.EvidenceType, "dokumen")
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.SourceModule = normalizeGovernanceText(arg.SourceModule, "governance")
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	arg.Status = normalizeGovernanceText(arg.Status, "needed")
	arg.Notes = strings.TrimSpace(arg.Notes)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	return arg
}

func normalizeGovernanceComplianceActionCreate(arg db.CreateGovernanceComplianceActionParams) db.CreateGovernanceComplianceActionParams {
	arg.SourceType = normalizeGovernanceText(arg.SourceType, "manual")
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Description = strings.TrimSpace(arg.Description)
	arg.Priority = normalizeGovernanceText(arg.Priority, "medium")
	arg.Status = normalizeGovernanceText(arg.Status, "open")
	arg.FollowUpNotes = strings.TrimSpace(arg.FollowUpNotes)
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	arg.CompletedAt = normalizeGovernanceComplianceCompletedAt(arg.Status, arg.CompletedAt)
	return arg
}

func normalizeGovernanceComplianceActionUpdate(arg db.UpdateGovernanceComplianceActionParams) db.UpdateGovernanceComplianceActionParams {
	arg.SourceType = normalizeGovernanceText(arg.SourceType, "manual")
	arg.SnpStandard = normalizeSNPStandard(arg.SnpStandard)
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Description = strings.TrimSpace(arg.Description)
	arg.Priority = normalizeGovernanceText(arg.Priority, "medium")
	arg.Status = normalizeGovernanceText(arg.Status, "open")
	arg.FollowUpNotes = strings.TrimSpace(arg.FollowUpNotes)
	arg.EvidenceUrl = strings.TrimSpace(arg.EvidenceUrl)
	if arg.PeriodYear == 0 {
		arg.PeriodYear = int32(time.Now().Year())
	}
	arg.CompletedAt = normalizeGovernanceComplianceCompletedAt(arg.Status, arg.CompletedAt)
	return arg
}

func normalizeGovernanceComplianceCompletedAt(status string, completedAt pgtype.Timestamptz) pgtype.Timestamptz {
	if status != "done" {
		return pgtype.Timestamptz{}
	}
	if completedAt.Valid {
		return completedAt
	}
	return pgtype.Timestamptz{Time: time.Now(), Valid: true}
}

var validGovernanceDocumentTypes = map[string]bool{
	"visi_misi": true,
	"rkjm":      true,
	"rkt":       true,
	"renstra":   true,
	"perkin":    true,
	"iku":       true,
	"sk":        true,
	"sop":       true,
	"snp":       true,
	"lainnya":   true,
}

var validGovernanceDocumentStatuses = map[string]bool{
	"draft": true,
	"final": true,
	"arsip": true,
}

var validGovernanceProgramStatuses = map[string]bool{
	"planned":     true,
	"in_progress": true,
	"done":        true,
	"blocked":     true,
}

var validGovernanceSNPStandards = map[string]bool{
	"":            true,
	"skl":         true,
	"isi":         true,
	"proses":      true,
	"penilaian":   true,
	"ptk":         true,
	"sarpras":     true,
	"pengelolaan": true,
	"pembiayaan":  true,
}

var validGovernancePerformanceAspects = map[string]bool{
	"hasil_kerja":    true,
	"perilaku_kerja": true,
	"tambahan":       true,
}

var validGovernanceEvidenceTypes = map[string]bool{
	"dokumen": true,
	"foto":    true,
	"tautan":  true,
	"laporan": true,
	"arsip":   true,
	"lainnya": true,
}

var validGovernanceEvidenceStatuses = map[string]bool{
	"needed":    true,
	"collected": true,
	"verified":  true,
	"gap":       true,
}

var validGovernanceComplianceActionSources = map[string]bool{
	"alignment_gap": true,
	"snp_gap":       true,
	"audit":         true,
	"document":      true,
	"program":       true,
	"performance":   true,
	"evidence":      true,
	"manual":        true,
}

var validGovernanceComplianceActionPriorities = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
	"urgent": true,
}

var validGovernanceComplianceActionStatuses = map[string]bool{
	"open":             true,
	"in_progress":      true,
	"waiting_evidence": true,
	"done":             true,
	"cancelled":        true,
}
