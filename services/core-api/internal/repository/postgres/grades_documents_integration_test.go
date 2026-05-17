package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationGradesDocumentsGovernanceArchiveRepositoryQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	user, err := q.CreateUserWithMustChangePassword(ctx, CreateUserWithMustChangePasswordParams{
		Username:           "it-doc-user-" + suffix,
		PasswordHash:       "hash:v1:" + suffix,
		DisplayName:        pgtype.Text{String: "Integration Document User " + suffix, Valid: true},
		IsActive:           true,
		MustChangePassword: false,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	year, err := q.CreateAcademicYear(ctx, CreateAcademicYearParams{
		Name:      "Integration Grades Documents Year " + suffix,
		StartDate: pgDate(2026, time.July, 1),
		EndDate:   pgDate(2027, time.June, 30),
		IsActive:  false,
	})
	if err != nil {
		t.Fatalf("create academic year: %v", err)
	}
	class, err := q.CreateSchoolClass(ctx, CreateSchoolClassParams{
		AcademicYearID: year.ID,
		Code:           "IT-GD-CLS-" + suffix,
		Name:           "Integration Grades Documents Class " + suffix,
		Level:          "8",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create school class: %v", err)
	}
	subject, err := q.CreateSubject(ctx, CreateSubjectParams{
		Code:                "IT-GD-SUB-" + suffix,
		Name:                "Integration Grades Documents Subject " + suffix,
		Category:            "umum",
		IsAssessmentSubject: true,
		IsReportSubject:     true,
		IsScheduleActivity:  false,
		CountsForRanking:    true,
		IsLocalContent:      false,
		IsChoiceSubject:     false,
		DefaultWeeklyHours:  2,
		DisplayOrder:        9997,
		IsActive:            true,
	})
	if err != nil {
		t.Fatalf("create subject: %v", err)
	}
	teacherID := insertIntegrationEmployee(t, ctx, tdb, "IT-GD-T-"+suffix, "Integration Grades Teacher "+suffix, employeeUID('3', suffix))
	studentID := insertIntegrationStudent(t, ctx, tdb, "IT-GD-NIS-"+suffix, "Integration Grades Student "+suffix, class.ID)
	assignment, err := q.CreateRombelSubjectAssignment(ctx, CreateRombelSubjectAssignmentParams{
		ClassID:           class.ID,
		SubjectID:         subject.ID,
		TeacherEmployeeID: teacherID,
	})
	if err != nil {
		t.Fatalf("create rombel subject assignment: %v", err)
	}

	component, err := q.CreateGradeComponent(ctx, CreateGradeComponentParams{
		AssignmentID: assignment.ID,
		Title:        "Integration Tugas " + suffix,
		Category:     "assignment",
		Weight:       40,
		MaxScore:     100,
		IsPublished:  true,
	})
	if err != nil {
		t.Fatalf("create grade component: %v", err)
	}
	if _, err := q.UpsertGradeEntry(ctx, UpsertGradeEntryParams{
		ComponentID: component.ID,
		StudentID:   studentID,
		Score:       pgtype.Float8{Float64: 90, Valid: true},
		Notes:       "integration grade entry",
		GradedBy:    "integration-test",
	}); err != nil {
		t.Fatalf("upsert grade entry: %v", err)
	}
	if _, err := q.UpsertGradeStudentSubjectDescription(ctx, UpsertGradeStudentSubjectDescriptionParams{
		AssignmentID: assignment.ID,
		StudentID:    studentID,
		Description:  "Strong integration performance",
	}); err != nil {
		t.Fatalf("upsert grade description: %v", err)
	}
	if _, err := q.UpsertGradeAssignmentFinalization(ctx, UpsertGradeAssignmentFinalizationParams{
		AssignmentID: assignment.ID,
		FinalizedBy:  "integration-test",
		Notes:        "finalized by integration test",
	}); err != nil {
		t.Fatalf("finalize grade assignment: %v", err)
	}

	components, err := q.ListGradeComponents(ctx, ListGradeComponentsParams{AssignmentID: assignment.ID, PublishedOnly: true})
	if err != nil {
		t.Fatalf("list grade components: %v", err)
	}
	if !gradeComponentListed(components, component.ID) {
		t.Fatalf("created grade component %v not returned by ListGradeComponents", component.ID)
	}
	entries, err := q.ListGradeEntriesByComponent(ctx, component.ID)
	if err != nil {
		t.Fatalf("list grade entries by component: %v", err)
	}
	if len(entries) != 1 || entries[0].StudentID != studentID || entries[0].Score != 90 {
		t.Fatalf("grade entries = %#v, want one score 90 for student %v", entries, studentID)
	}
	summary, err := q.ListGradebookSummary(ctx, ListGradebookSummaryParams{AssignmentID: assignment.ID, PublishedOnly: true})
	if err != nil {
		t.Fatalf("list gradebook summary: %v", err)
	}
	if len(summary) != 1 || summary[0].StudentID != studentID || summary[0].FinalScore != 90 || summary[0].ReportDescription == "" {
		t.Fatalf("gradebook summary = %#v", summary)
	}
	statuses, err := q.ListGradeAssignmentStatuses(ctx)
	if err != nil {
		t.Fatalf("list grade assignment statuses: %v", err)
	}
	if !gradeAssignmentStatusListed(statuses, assignment.ID) {
		t.Fatalf("finalized assignment %v not returned with expected grade status aggregates", assignment.ID)
	}

	unit, err := q.CreateGovernanceUnit(ctx, CreateGovernanceUnitParams{
		Code:        "IT-GD-UNIT-" + suffix,
		Name:        "Integration Governance Unit " + suffix,
		UnitType:    "team",
		Description: "integration governance unit",
		IsActive:    true,
		SortOrder:   9997,
	})
	if err != nil {
		t.Fatalf("create governance unit: %v", err)
	}
	position, err := q.CreateGovernancePosition(ctx, CreateGovernancePositionParams{
		UnitID:       unit.ID,
		Title:        "Integration Governance Position " + suffix,
		PositionType: "coordinator",
		Description:  "integration governance position",
		Tupoksi:      "integration tupoksi",
		IsActive:     true,
		SortOrder:    9997,
	})
	if err != nil {
		t.Fatalf("create governance position: %v", err)
	}
	assignmentGov, err := q.CreateGovernanceAssignment(ctx, CreateGovernanceAssignmentParams{
		PositionID: position.ID,
		EmployeeID: teacherID,
		StartDate:  pgDate(2026, time.July, 1),
		Notes:      "integration governance assignment",
	})
	if err != nil {
		t.Fatalf("create governance assignment: %v", err)
	}
	govDoc, err := q.CreateGovernanceDocument(ctx, CreateGovernanceDocumentParams{
		DocType:         "rkt",
		Title:           "Integration Governance Document " + suffix,
		PeriodYear:      2026,
		PeriodLabel:     "2026",
		OwnerUnitID:     unit.ID,
		SnpStandard:     "pengelolaan",
		Status:          "final",
		DocumentUrl:     "https://example.invalid/governance/" + suffix,
		Summary:         "integration governance document",
		CreatedByUserID: user.ID,
	})
	if err != nil {
		t.Fatalf("create governance document: %v", err)
	}
	evidence, err := q.CreateGovernanceEvidenceItem(ctx, CreateGovernanceEvidenceItemParams{
		PeriodYear:      2026,
		Title:           "Integration Evidence " + suffix,
		EvidenceType:    "dokumen",
		SnpStandard:     "pengelolaan",
		OwnerUnitID:     unit.ID,
		DocumentID:      govDoc.ID,
		SourceModule:    "integration",
		EvidenceUrl:     "https://example.invalid/evidence/" + suffix,
		Status:          "verified",
		Notes:           "integration evidence",
		CreatedByUserID: user.ID,
	})
	if err != nil {
		t.Fatalf("create governance evidence: %v", err)
	}
	compliance, err := q.CreateGovernanceComplianceAction(ctx, CreateGovernanceComplianceActionParams{
		PeriodYear:            2026,
		SourceType:            "audit",
		SnpStandard:           "pengelolaan",
		EvidenceItemID:        evidence.ID,
		OwnerUnitID:           unit.ID,
		ResponsibleEmployeeID: teacherID,
		Title:                 "Integration Compliance Action " + suffix,
		Description:           "integration compliance action",
		Priority:              "urgent",
		Status:                "open",
		DueDate:               pgDate(2026, time.August, 10),
		FollowUpNotes:         "integration follow up",
		EvidenceUrl:           "https://example.invalid/compliance/" + suffix,
		CreatedByUserID:       user.ID,
	})
	if err != nil {
		t.Fatalf("create governance compliance action: %v", err)
	}

	units, err := q.ListGovernanceUnits(ctx, "Integration Governance Unit "+suffix)
	if err != nil {
		t.Fatalf("list governance units: %v", err)
	}
	if !governanceUnitListed(units, unit.ID) {
		t.Fatalf("created governance unit %v not returned by ListGovernanceUnits", unit.ID)
	}
	assignments, err := q.ListGovernanceAssignments(ctx, ListGovernanceAssignmentsParams{ActiveOnly: true, Search: suffix})
	if err != nil {
		t.Fatalf("list governance assignments: %v", err)
	}
	if !governanceAssignmentListed(assignments, assignmentGov.ID) {
		t.Fatalf("created governance assignment %v not returned by ListGovernanceAssignments", assignmentGov.ID)
	}
	documents, err := q.ListGovernanceDocuments(ctx, ListGovernanceDocumentsParams{Search: suffix, DocType: "rkt", PeriodYear: 2026, SnpStandard: "pengelolaan"})
	if err != nil {
		t.Fatalf("list governance documents: %v", err)
	}
	if !governanceDocumentListed(documents, govDoc.ID) {
		t.Fatalf("created governance document %v not returned by ListGovernanceDocuments", govDoc.ID)
	}
	evidenceItems, err := q.ListGovernanceEvidenceItems(ctx, ListGovernanceEvidenceItemsParams{Search: suffix, PeriodYear: 2026, Status: "verified", SnpStandard: "pengelolaan"})
	if err != nil {
		t.Fatalf("list governance evidence items: %v", err)
	}
	if !governanceEvidenceListed(evidenceItems, evidence.ID) {
		t.Fatalf("created governance evidence %v not returned by ListGovernanceEvidenceItems", evidence.ID)
	}
	complianceActions, err := q.ListGovernanceComplianceActions(ctx, ListGovernanceComplianceActionsParams{Search: suffix, PeriodYear: 2026, Status: "open", Priority: "urgent", SourceType: "audit", SnpStandard: "pengelolaan", ResponsibleEmployeeID: teacherID})
	if err != nil {
		t.Fatalf("list governance compliance actions: %v", err)
	}
	if !governanceComplianceActionListed(complianceActions, compliance.ID) {
		t.Fatalf("created compliance action %v not returned by ListGovernanceComplianceActions", compliance.ID)
	}
	govStats, err := q.GetGovernanceStats(ctx)
	if err != nil {
		t.Fatalf("get governance stats: %v", err)
	}
	if govStats.TotalUnits < 1 || govStats.TotalPositions < 1 || govStats.ActiveAssignments < 1 || govStats.FinalDocuments < 1 || govStats.EvidenceItems < 1 || govStats.CriticalComplianceActions < 1 {
		t.Fatalf("governance stats missing created rows: %#v", govStats)
	}

	category, err := q.CreateArchiveCategory(ctx, CreateArchiveCategoryParams{
		Code:               "IT-GD-ARC-" + suffix,
		Name:               "Integration Archive Category " + suffix,
		ClassificationCode: "",
		Description:        "integration archive category",
		RetentionYears:     5,
		IsActive:           true,
	})
	if err != nil {
		t.Fatalf("create archive category: %v", err)
	}
	archiveDoc, err := q.CreateArchiveDocument(ctx, CreateArchiveDocumentParams{
		CategoryID:       category.ID,
		Title:            "Integration Archive Document " + suffix,
		ArchiveNumber:    "ARC/IT/" + suffix,
		DocumentDate:     pgDate(2026, time.July, 2),
		ReceivedDate:     pgDate(time.Now().UTC().Year(), time.July, 3),
		Summary:          "integration archive document summary",
		Tags:             "integration," + suffix,
		Status:           "active",
		StorageLocation:  "Integration Rack",
		RetentionUntil:   pgDate(2031, time.July, 3),
		OriginalName:     "integration.pdf",
		StoredName:       "integration-" + suffix + ".pdf",
		FilePath:         "/tmp/integration-" + suffix + ".pdf",
		MimeType:         "application/pdf",
		FileSize:         4096,
		ChecksumSha256:   "sha256-" + suffix,
		UploadedByUserID: user.ID,
	})
	if err != nil {
		t.Fatalf("create archive document: %v", err)
	}
	categories, err := q.ListArchiveCategories(ctx, ListArchiveCategoriesParams{Search: suffix, ActiveOnly: true})
	if err != nil {
		t.Fatalf("list archive categories: %v", err)
	}
	if !archiveCategoryListed(categories, category.ID) {
		t.Fatalf("created archive category %v not returned by ListArchiveCategories", category.ID)
	}
	archiveDocuments, err := q.ListArchiveDocuments(ctx, ListArchiveDocumentsParams{Search: suffix, CategoryID: category.ID, Status: "active"})
	if err != nil {
		t.Fatalf("list archive documents: %v", err)
	}
	if !archiveDocumentListed(archiveDocuments, archiveDoc.ID) {
		t.Fatalf("created archive document %v not returned by ListArchiveDocuments", archiveDoc.ID)
	}
	archiveDetail, err := q.GetArchiveDocumentDetail(ctx, archiveDoc.ID)
	if err != nil {
		t.Fatalf("get archive document detail: %v", err)
	}
	if archiveDetail.ID != archiveDoc.ID || archiveDetail.CategoryCode != category.Code || archiveDetail.UploadedByUsername != user.Username {
		t.Fatalf("archive detail = %#v", archiveDetail)
	}
	archiveStats, err := q.GetArchiveStats(ctx)
	if err != nil {
		t.Fatalf("get archive stats: %v", err)
	}
	if archiveStats.TotalDocuments < 1 || archiveStats.ActiveDocuments < 1 || archiveStats.TotalFileSize < archiveDoc.FileSize {
		t.Fatalf("archive stats missing created document: %#v", archiveStats)
	}

	catalog, err := q.CreateDocumentCycleCatalog(ctx, CreateDocumentCycleCatalogParams{
		Code:                         "IT-GD-CYC-" + suffix,
		Title:                        "Integration Document Cycle " + suffix,
		Frequency:                    "monthly",
		DomainArea:                   "governance",
		ExternalSystem:               "edm",
		SnpStandard:                  "pengelolaan",
		RegulationRef:                "Integration regulation " + suffix,
		DefaultOwnerUnitID:           unit.ID,
		DefaultResponsibleEmployeeID: teacherID,
		DefaultVerifierEmployeeID:    teacherID,
		DeadlineDaysAfterPeriod:      7,
		ReminderDaysBeforeDue:        2,
		Description:                  "integration document cycle catalog",
		IsActive:                     true,
		SortOrder:                    9997,
	})
	if err != nil {
		t.Fatalf("create document cycle catalog: %v", err)
	}
	obligation, err := q.EnsureDocumentCycleObligation(ctx, EnsureDocumentCycleObligationParams{
		CatalogID:             catalog.ID,
		PeriodYear:            2026,
		PeriodLabel:           "Juli 2026 " + suffix,
		PeriodStart:           pgDate(2026, time.July, 1),
		PeriodEnd:             pgDate(2026, time.July, 31),
		DueDate:               pgDate(2026, time.August, 7),
		ReminderDate:          pgDate(2026, time.August, 5),
		OwnerUnitID:           unit.ID,
		ResponsibleEmployeeID: teacherID,
		VerifierEmployeeID:    teacherID,
		Status:                "draft",
		Notes:                 "integration document cycle obligation",
		CreatedByUserID:       user.ID,
	})
	if err != nil {
		t.Fatalf("ensure document cycle obligation: %v", err)
	}
	obligation, err = q.UpdateDocumentCycleObligation(ctx, UpdateDocumentCycleObligationParams{
		DueDate:               pgDate(2026, time.August, 7),
		ReminderDate:          pgDate(2026, time.August, 5),
		DomainArea:            "governance",
		ExternalSystem:        "edm",
		OwnerUnitID:           unit.ID,
		ResponsibleEmployeeID: teacherID,
		VerifierEmployeeID:    teacherID,
		GovernanceDocumentID:  govDoc.ID,
		EvidenceItemID:        evidence.ID,
		ComplianceActionID:    compliance.ID,
		ArchiveDocumentID:     archiveDoc.ID,
		Notes:                 "integration document cycle obligation linked",
		VerificationNotes:     "integration verification notes",
		ID:                    obligation.ID,
	})
	if err != nil {
		t.Fatalf("update document cycle obligation links: %v", err)
	}
	if _, err := q.CreateDocumentCycleEvent(ctx, CreateDocumentCycleEventParams{
		ObligationID: obligation.ID,
		EventType:    "status_changed",
		FromStatus:   "not_started",
		ToStatus:     "draft",
		Notes:        "integration document cycle event",
		ActorUserID:  user.ID,
	}); err != nil {
		t.Fatalf("create document cycle event: %v", err)
	}
	catalogs, err := q.ListDocumentCycleCatalogs(ctx, ListDocumentCycleCatalogsParams{Search: suffix, Frequency: "monthly", ActiveOnly: true})
	if err != nil {
		t.Fatalf("list document cycle catalogs: %v", err)
	}
	if !documentCycleCatalogListed(catalogs, catalog.ID) {
		t.Fatalf("created document cycle catalog %v not returned by ListDocumentCycleCatalogs", catalog.ID)
	}
	obligations, err := q.ListDocumentCycleObligations(ctx, ListDocumentCycleObligationsParams{
		Search:                suffix,
		Status:                "draft",
		Frequency:             "monthly",
		DomainArea:            "governance",
		ExternalSystem:        "edm",
		PeriodYear:            2026,
		OwnerUnitID:           unit.ID,
		ResponsibleEmployeeID: teacherID,
		VerifierEmployeeID:    teacherID,
	})
	if err != nil {
		t.Fatalf("list document cycle obligations: %v", err)
	}
	if !documentCycleObligationListed(obligations, obligation.ID, govDoc.ID, archiveDoc.ID, evidence.ID, compliance.ID) {
		t.Fatalf("linked document cycle obligation %v not returned with expected joins: %#v", obligation.ID, obligations)
	}
	events, err := q.ListDocumentCycleEventsByObligation(ctx, ListDocumentCycleEventsByObligationParams{ObligationID: obligation.ID, EventType: "status_changed", Actor: user.Username})
	if err != nil {
		t.Fatalf("list document cycle events: %v", err)
	}
	if len(events) != 1 || events[0].ObligationID != obligation.ID || events[0].ActorUsername != user.Username {
		t.Fatalf("document cycle events = %#v", events)
	}
	cycleStats, err := q.GetDocumentCycleStats(ctx, 2026)
	if err != nil {
		t.Fatalf("get document cycle stats: %v", err)
	}
	if cycleStats.ActiveCatalogs < 1 || cycleStats.TotalObligations < 1 || cycleStats.LinkedArchiveObligations < 1 || cycleStats.LinkedEvidenceObligations < 1 || cycleStats.LinkedGovernanceDocumentObligations < 1 || cycleStats.LinkedComplianceActionObligations < 1 || cycleStats.ExternalTrackerObligations < 1 {
		t.Fatalf("document cycle stats missing linked obligation: %#v", cycleStats)
	}
}

func gradeComponentListed(items []ListGradeComponentsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func gradeAssignmentStatusListed(items []ListGradeAssignmentStatusesRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.AssignmentID == id && item.ComponentCount == 1 && item.PublishedComponentCount == 1 && item.StudentCount == 1 && item.ReadyStudentCount == 1 && item.IsFinalized {
			return true
		}
	}
	return false
}

func governanceUnitListed(items []ListGovernanceUnitsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func governanceAssignmentListed(items []ListGovernanceAssignmentsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func governanceDocumentListed(items []ListGovernanceDocumentsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func governanceEvidenceListed(items []ListGovernanceEvidenceItemsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func governanceComplianceActionListed(items []ListGovernanceComplianceActionsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func archiveCategoryListed(items []ListArchiveCategoriesRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id && item.DocumentCount == 1 {
			return true
		}
	}
	return false
}

func archiveDocumentListed(items []ListArchiveDocumentsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func documentCycleCatalogListed(items []ListDocumentCycleCatalogsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func documentCycleObligationListed(items []ListDocumentCycleObligationsRow, id, governanceDocumentID, archiveDocumentID, evidenceItemID, complianceActionID pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id && item.GovernanceDocumentID == governanceDocumentID && item.ArchiveDocumentID == archiveDocumentID && item.EvidenceItemID == evidenceItemID && item.ComplianceActionID == complianceActionID {
			return true
		}
	}
	return false
}
