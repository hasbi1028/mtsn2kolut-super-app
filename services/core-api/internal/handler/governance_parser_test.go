package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const governanceParserUUID = "01000000-0000-0000-0000-000000000000"

func TestGovernanceQueryAndBoolHelpers(t *testing.T) {
	falseValue := false
	trueValue := true
	if boolDefault(nil, true) != true || boolDefault(&falseValue, true) != false || boolDefault(&trueValue, false) != true {
		t.Fatal("boolDefault() did not respect nil fallback and explicit values")
	}
	for _, value := range []string{"1", "true", "YES", " y "} {
		if !boolQuery(value) {
			t.Fatalf("boolQuery(%q) = false, want true", value)
		}
	}
	for _, value := range []string{"0", "false", "no", "bad"} {
		if boolQuery(value) {
			t.Fatalf("boolQuery(%q) = true, want false", value)
		}
	}
	if got := int32Query(" 2026 "); got != 2026 {
		t.Fatalf("int32Query() = %d, want 2026", got)
	}
	if got := int32Query("bad"); got != 0 {
		t.Fatalf("int32Query(bad) = %d, want 0", got)
	}
}

func TestGovernanceParsePositionAndAssignmentRequests(t *testing.T) {
	rec := httptest.NewRecorder()
	unitID, parentID, ok := parsePositionIDs(rec, governancePositionRequest{
		UnitID:           governanceParserUUID,
		ParentPositionID: governanceParserUUID,
	})
	if !ok || !unitID.Valid || !parentID.Valid {
		t.Fatalf("parsePositionIDs() = %v/%v/%v, want valid ids", unitID, parentID, ok)
	}
	rec = httptest.NewRecorder()
	if _, _, ok := parsePositionIDs(rec, governancePositionRequest{UnitID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parsePositionIDs(bad) ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	assignment, ok := parseAssignmentRequest(rec, governanceAssignmentRequest{
		PositionID:             governanceParserUUID,
		EmployeeID:             governanceParserUUID,
		StartDate:              "2026-05-01",
		EndDate:                "2026-12-31",
		DecreeOutgoingLetterID: governanceParserUUID,
		Notes:                  "SK tugas",
	})
	if !ok || !assignment.PositionID.Valid || !assignment.EmployeeID.Valid || assignment.Notes != "SK tugas" {
		t.Fatalf("parseAssignmentRequest() = %+v/%v, want mapped assignment", assignment, ok)
	}
	if !assignment.StartDate.Valid || assignment.StartDate.Time.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("parseAssignmentRequest() start_date = %v, want 2026-05-01", assignment.StartDate)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseAssignmentRequest(rec, governanceAssignmentRequest{PositionID: governanceParserUUID, EmployeeID: governanceParserUUID, StartDate: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseAssignmentRequest(bad date) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
}

func TestGovernanceParseDocumentProgramAndWorkPlanRequests(t *testing.T) {
	rec := httptest.NewRecorder()
	document, ok := parseDocumentRequest(rec, governanceDocumentRequest{
		DocType:          "rkt",
		Title:            "RKT 2026",
		PeriodYear:       2026,
		PeriodLabel:      "2026",
		OwnerUnitID:      governanceParserUUID,
		SnpStandard:      "skl",
		Status:           "draft",
		DocumentUrl:      "/docs/rkt.pdf",
		OutgoingLetterID: governanceParserUUID,
		Summary:          "ringkasan",
	})
	if !ok || document.DocType != "rkt" || document.OwnerUnitID.Valid != true || document.OutgoingLetterID.Valid != true {
		t.Fatalf("parseDocumentRequest() = %+v/%v, want mapped document", document, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseDocumentRequest(rec, governanceDocumentRequest{OwnerUnitID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseDocumentRequest(bad uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	program, ok := parseProgramRequest(rec, governanceProgramRequest{
		PeriodYear:            2026,
		Code:                  "P-1",
		Name:                  "Mutu Madrasah",
		SourceDocumentID:      governanceParserUUID,
		OwnerUnitID:           governanceParserUUID,
		ResponsiblePositionID: governanceParserUUID,
		ResponsibleEmployeeID: governanceParserUUID,
		SnpStandard:           "skl",
		IkuCode:               "IKU-1",
		Indicator:             "Persentase",
		TargetValue:           "90",
		TargetUnit:            "persen",
		Status:                "planned",
		ProgressPercent:       10,
		RealizationSummary:    "berjalan",
		EvidenceUrl:           "/evidence",
		DueDate:               "2026-06-30",
	})
	if !ok || program.Code != "P-1" || !program.SourceDocumentID.Valid || !program.DueDate.Valid {
		t.Fatalf("parseProgramRequest() = %+v/%v, want mapped program", program, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseProgramRequest(rec, governanceProgramRequest{DueDate: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseProgramRequest(bad date) ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	workPlan, ok := parseWorkPlanItemRequest(rec, governanceWorkPlanItemRequest{
		PeriodYear:            2026,
		ProgramID:             governanceParserUUID,
		SourceDocumentID:      governanceParserUUID,
		OwnerUnitID:           governanceParserUUID,
		ResponsibleEmployeeID: governanceParserUUID,
		EvidenceItemID:        governanceParserUUID,
		ActivityCode:          "A-1",
		ActivityName:          "Workshop",
		OutputIndicator:       "Dokumen",
		TargetVolume:          "1",
		TargetUnit:            "paket",
		BudgetSource:          "BOS",
		BudgetAmount:          1000,
		RealizationAmount:     500,
		Status:                "planned",
		ProgressPercent:       25,
		StartDate:             "2026-05-01",
		EndDate:               "2026-05-10",
		EvidenceUrl:           "/bukti",
		Notes:                 "catatan",
	})
	if !ok || workPlan.ActivityCode != "A-1" || !workPlan.ProgramID.Valid || !workPlan.StartDate.Valid || !workPlan.EndDate.Valid {
		t.Fatalf("parseWorkPlanItemRequest() = %+v/%v, want mapped work plan", workPlan, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseWorkPlanItemRequest(rec, governanceWorkPlanItemRequest{ProgramID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseWorkPlanItemRequest(bad uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
}

func TestGovernanceParseTargetEvidenceAndComplianceRequests(t *testing.T) {
	rec := httptest.NewRecorder()
	target, ok := parsePerformanceTargetRequest(rec, governancePerformanceTargetRequest{
		PeriodYear:      2026,
		EmployeeID:      governanceParserUUID,
		PositionID:      governanceParserUUID,
		ProgramID:       governanceParserUUID,
		ParentTargetID:  governanceParserUUID,
		Aspect:          "hasil_kerja",
		Title:           "Target SKP",
		Indicator:       "Indikator",
		TargetValue:     "90",
		TargetUnit:      "persen",
		Status:          "planned",
		ProgressPercent: 30,
		EvidenceUrl:     "/bukti",
		ReviewNotes:     "review",
		DueDate:         "2026-05-20",
	})
	if !ok || target.Title != "Target SKP" || !target.EmployeeID.Valid || !target.DueDate.Valid {
		t.Fatalf("parsePerformanceTargetRequest() = %+v/%v, want mapped target", target, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parsePerformanceTargetRequest(rec, governancePerformanceTargetRequest{EmployeeID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parsePerformanceTargetRequest(bad uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	evidence, ok := parseEvidenceItemRequest(rec, governanceEvidenceItemRequest{
		PeriodYear:          2026,
		Title:               "Dokumen EDM",
		EvidenceType:        "dokumen",
		SnpStandard:         "skl",
		OwnerUnitID:         governanceParserUUID,
		DocumentID:          governanceParserUUID,
		ProgramID:           governanceParserUUID,
		PerformanceTargetID: governanceParserUUID,
		SourceModule:        "governance",
		EvidenceUrl:         "/edm",
		Status:              "available",
		Notes:               "lengkap",
	})
	if !ok || evidence.Title != "Dokumen EDM" || !evidence.OwnerUnitID.Valid || !evidence.PerformanceTargetID.Valid {
		t.Fatalf("parseEvidenceItemRequest() = %+v/%v, want mapped evidence", evidence, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseEvidenceItemRequest(rec, governanceEvidenceItemRequest{OwnerUnitID: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseEvidenceItemRequest(bad uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}

	rec = httptest.NewRecorder()
	action, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{
		PeriodYear:            2026,
		SourceType:            "manual",
		SourceRefID:           governanceParserUUID,
		SnpStandard:           "skl",
		ProgramID:             governanceParserUUID,
		DocumentID:            governanceParserUUID,
		PerformanceTargetID:   governanceParserUUID,
		EvidenceItemID:        governanceParserUUID,
		OwnerUnitID:           governanceParserUUID,
		ResponsibleEmployeeID: governanceParserUUID,
		Title:                 "Tindak lanjut",
		Description:           "Lengkapi lampiran",
		Priority:              "high",
		Status:                "open",
		DueDate:               "2026-05-30",
		CompletedAt:           "2026-06-01T08:00:00Z",
		FollowUpNotes:         "catatan",
		EvidenceUrl:           "/bukti",
	})
	if !ok || action.Title != "Tindak lanjut" || !action.SourceRefID.Valid || !action.CompletedAt.Valid {
		t.Fatalf("parseComplianceActionRequest() = %+v/%v, want mapped compliance action", action, ok)
	}
	rec = httptest.NewRecorder()
	if _, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{CompletedAt: "bad"}); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseComplianceActionRequest(bad timestamp) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
}

func TestGovernanceParsersRejectInvalidNestedFields(t *testing.T) {
	tests := []struct {
		name  string
		parse func(*httptest.ResponseRecorder) bool
	}{
		{
			name: "position parent",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, _, ok := parsePositionIDs(rec, governancePositionRequest{UnitID: governanceParserUUID, ParentPositionID: "bad"})
				return ok
			},
		},
		{
			name: "assignment employee",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseAssignmentRequest(rec, governanceAssignmentRequest{PositionID: governanceParserUUID, EmployeeID: "bad"})
				return ok
			},
		},
		{
			name: "assignment end date",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseAssignmentRequest(rec, governanceAssignmentRequest{PositionID: governanceParserUUID, EmployeeID: governanceParserUUID, StartDate: "2026-05-01", EndDate: "bad"})
				return ok
			},
		},
		{
			name: "assignment decree",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseAssignmentRequest(rec, governanceAssignmentRequest{PositionID: governanceParserUUID, EmployeeID: governanceParserUUID, StartDate: "2026-05-01", DecreeOutgoingLetterID: "bad"})
				return ok
			},
		},
		{
			name: "document outgoing letter",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseDocumentRequest(rec, governanceDocumentRequest{OwnerUnitID: governanceParserUUID, OutgoingLetterID: "bad"})
				return ok
			},
		},
		{
			name: "program owner",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseProgramRequest(rec, governanceProgramRequest{SourceDocumentID: governanceParserUUID, OwnerUnitID: "bad"})
				return ok
			},
		},
		{
			name: "program responsible position",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseProgramRequest(rec, governanceProgramRequest{OwnerUnitID: governanceParserUUID, ResponsiblePositionID: "bad"})
				return ok
			},
		},
		{
			name: "program responsible employee",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseProgramRequest(rec, governanceProgramRequest{OwnerUnitID: governanceParserUUID, ResponsiblePositionID: governanceParserUUID, ResponsibleEmployeeID: "bad"})
				return ok
			},
		},
		{
			name: "work plan source document",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseWorkPlanItemRequest(rec, governanceWorkPlanItemRequest{ProgramID: governanceParserUUID, SourceDocumentID: "bad"})
				return ok
			},
		},
		{
			name: "work plan owner",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseWorkPlanItemRequest(rec, governanceWorkPlanItemRequest{ProgramID: governanceParserUUID, SourceDocumentID: governanceParserUUID, OwnerUnitID: "bad"})
				return ok
			},
		},
		{
			name: "work plan responsible employee",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseWorkPlanItemRequest(rec, governanceWorkPlanItemRequest{ProgramID: governanceParserUUID, SourceDocumentID: governanceParserUUID, OwnerUnitID: governanceParserUUID, ResponsibleEmployeeID: "bad"})
				return ok
			},
		},
		{
			name: "work plan evidence",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseWorkPlanItemRequest(rec, governanceWorkPlanItemRequest{ProgramID: governanceParserUUID, SourceDocumentID: governanceParserUUID, OwnerUnitID: governanceParserUUID, ResponsibleEmployeeID: governanceParserUUID, EvidenceItemID: "bad"})
				return ok
			},
		},
		{
			name: "work plan start date",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseWorkPlanItemRequest(rec, governanceWorkPlanItemRequest{StartDate: "bad"})
				return ok
			},
		},
		{
			name: "work plan end date",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseWorkPlanItemRequest(rec, governanceWorkPlanItemRequest{StartDate: "2026-05-01", EndDate: "bad"})
				return ok
			},
		},
		{
			name: "target position",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parsePerformanceTargetRequest(rec, governancePerformanceTargetRequest{EmployeeID: governanceParserUUID, PositionID: "bad"})
				return ok
			},
		},
		{
			name: "target program",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parsePerformanceTargetRequest(rec, governancePerformanceTargetRequest{EmployeeID: governanceParserUUID, PositionID: governanceParserUUID, ProgramID: "bad"})
				return ok
			},
		},
		{
			name: "target parent",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parsePerformanceTargetRequest(rec, governancePerformanceTargetRequest{EmployeeID: governanceParserUUID, PositionID: governanceParserUUID, ProgramID: governanceParserUUID, ParentTargetID: "bad"})
				return ok
			},
		},
		{
			name: "target due date",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parsePerformanceTargetRequest(rec, governancePerformanceTargetRequest{DueDate: "bad"})
				return ok
			},
		},
		{
			name: "evidence document",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseEvidenceItemRequest(rec, governanceEvidenceItemRequest{OwnerUnitID: governanceParserUUID, DocumentID: "bad"})
				return ok
			},
		},
		{
			name: "evidence program",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseEvidenceItemRequest(rec, governanceEvidenceItemRequest{OwnerUnitID: governanceParserUUID, DocumentID: governanceParserUUID, ProgramID: "bad"})
				return ok
			},
		},
		{
			name: "evidence target",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseEvidenceItemRequest(rec, governanceEvidenceItemRequest{OwnerUnitID: governanceParserUUID, DocumentID: governanceParserUUID, ProgramID: governanceParserUUID, PerformanceTargetID: "bad"})
				return ok
			},
		},
		{
			name: "compliance source ref",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{SourceRefID: "bad"})
				return ok
			},
		},
		{
			name: "compliance program",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{SourceRefID: governanceParserUUID, ProgramID: "bad"})
				return ok
			},
		},
		{
			name: "compliance document",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{SourceRefID: governanceParserUUID, ProgramID: governanceParserUUID, DocumentID: "bad"})
				return ok
			},
		},
		{
			name: "compliance performance target",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{SourceRefID: governanceParserUUID, ProgramID: governanceParserUUID, DocumentID: governanceParserUUID, PerformanceTargetID: "bad"})
				return ok
			},
		},
		{
			name: "compliance evidence",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{SourceRefID: governanceParserUUID, ProgramID: governanceParserUUID, DocumentID: governanceParserUUID, PerformanceTargetID: governanceParserUUID, EvidenceItemID: "bad"})
				return ok
			},
		},
		{
			name: "compliance owner",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{SourceRefID: governanceParserUUID, ProgramID: governanceParserUUID, DocumentID: governanceParserUUID, PerformanceTargetID: governanceParserUUID, EvidenceItemID: governanceParserUUID, OwnerUnitID: "bad"})
				return ok
			},
		},
		{
			name: "compliance responsible employee",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{SourceRefID: governanceParserUUID, ProgramID: governanceParserUUID, DocumentID: governanceParserUUID, PerformanceTargetID: governanceParserUUID, EvidenceItemID: governanceParserUUID, OwnerUnitID: governanceParserUUID, ResponsibleEmployeeID: "bad"})
				return ok
			},
		},
		{
			name: "compliance due date",
			parse: func(rec *httptest.ResponseRecorder) bool {
				_, ok := parseComplianceActionRequest(rec, governanceComplianceActionRequest{DueDate: "bad"})
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			if ok := tt.parse(rec); ok || rec.Code != http.StatusBadRequest {
				t.Fatalf("%s ok/status = %v/%d, want false/400", tt.name, ok, rec.Code)
			}
		})
	}
}

func TestGovernanceHandlersRejectInvalidInputs(t *testing.T) {
	h := &Governance{}
	validID := handlerTestUUID(101).String()
	tests := []struct {
		name       string
		fn         http.HandlerFunc
		method     string
		target     string
		body       string
		id         string
		wantStatus int
	}{
		{name: "create unit invalid json", fn: h.CreateUnit, method: http.MethodPost, target: "/governance/units", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create unit invalid parent", fn: h.CreateUnit, method: http.MethodPost, target: "/governance/units", body: `{"parent_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update unit invalid id", fn: h.UpdateUnit, method: http.MethodPut, target: "/governance/units/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update unit invalid json", fn: h.UpdateUnit, method: http.MethodPut, target: "/governance/units/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "update unit invalid parent", fn: h.UpdateUnit, method: http.MethodPut, target: "/governance/units/" + validID, body: `{"parent_id":"bad"}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete unit invalid id", fn: h.DeleteUnit, method: http.MethodDelete, target: "/governance/units/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "create position invalid json", fn: h.CreatePosition, method: http.MethodPost, target: "/governance/positions", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create position invalid unit", fn: h.CreatePosition, method: http.MethodPost, target: "/governance/positions", body: `{"unit_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update position invalid id", fn: h.UpdatePosition, method: http.MethodPut, target: "/governance/positions/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update position invalid json", fn: h.UpdatePosition, method: http.MethodPut, target: "/governance/positions/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "update position invalid unit", fn: h.UpdatePosition, method: http.MethodPut, target: "/governance/positions/" + validID, body: `{"unit_id":"bad"}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete position invalid id", fn: h.DeletePosition, method: http.MethodDelete, target: "/governance/positions/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "create assignment invalid json", fn: h.CreateAssignment, method: http.MethodPost, target: "/governance/assignments", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create assignment invalid date", fn: h.CreateAssignment, method: http.MethodPost, target: "/governance/assignments", body: `{"position_id":"` + governanceParserUUID + `","employee_id":"` + governanceParserUUID + `","start_date":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update assignment invalid id", fn: h.UpdateAssignment, method: http.MethodPut, target: "/governance/assignments/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update assignment invalid json", fn: h.UpdateAssignment, method: http.MethodPut, target: "/governance/assignments/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "update assignment invalid date", fn: h.UpdateAssignment, method: http.MethodPut, target: "/governance/assignments/" + validID, body: `{"position_id":"` + governanceParserUUID + `","employee_id":"` + governanceParserUUID + `","start_date":"bad"}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete assignment invalid id", fn: h.DeleteAssignment, method: http.MethodDelete, target: "/governance/assignments/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "create document invalid json", fn: h.CreateDocument, method: http.MethodPost, target: "/governance/documents", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create document invalid owner", fn: h.CreateDocument, method: http.MethodPost, target: "/governance/documents", body: `{"owner_unit_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update document invalid id", fn: h.UpdateDocument, method: http.MethodPut, target: "/governance/documents/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update document invalid json", fn: h.UpdateDocument, method: http.MethodPut, target: "/governance/documents/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "update document invalid owner", fn: h.UpdateDocument, method: http.MethodPut, target: "/governance/documents/" + validID, body: `{"owner_unit_id":"bad"}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete document invalid id", fn: h.DeleteDocument, method: http.MethodDelete, target: "/governance/documents/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "create program invalid json", fn: h.CreateProgram, method: http.MethodPost, target: "/governance/programs", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create program invalid source", fn: h.CreateProgram, method: http.MethodPost, target: "/governance/programs", body: `{"source_document_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update program invalid id", fn: h.UpdateProgram, method: http.MethodPut, target: "/governance/programs/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update program invalid json", fn: h.UpdateProgram, method: http.MethodPut, target: "/governance/programs/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "update program invalid source", fn: h.UpdateProgram, method: http.MethodPut, target: "/governance/programs/" + validID, body: `{"source_document_id":"bad"}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete program invalid id", fn: h.DeleteProgram, method: http.MethodDelete, target: "/governance/programs/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "create work plan invalid json", fn: h.CreateWorkPlanItem, method: http.MethodPost, target: "/governance/work-plan", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create work plan invalid program", fn: h.CreateWorkPlanItem, method: http.MethodPost, target: "/governance/work-plan", body: `{"program_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update work plan invalid id", fn: h.UpdateWorkPlanItem, method: http.MethodPut, target: "/governance/work-plan/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update work plan invalid json", fn: h.UpdateWorkPlanItem, method: http.MethodPut, target: "/governance/work-plan/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "update work plan invalid program", fn: h.UpdateWorkPlanItem, method: http.MethodPut, target: "/governance/work-plan/" + validID, body: `{"program_id":"bad"}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete work plan invalid id", fn: h.DeleteWorkPlanItem, method: http.MethodDelete, target: "/governance/work-plan/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "create target invalid json", fn: h.CreatePerformanceTarget, method: http.MethodPost, target: "/governance/performance-targets", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create target invalid employee", fn: h.CreatePerformanceTarget, method: http.MethodPost, target: "/governance/performance-targets", body: `{"employee_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update target invalid id", fn: h.UpdatePerformanceTarget, method: http.MethodPut, target: "/governance/performance-targets/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update target invalid json", fn: h.UpdatePerformanceTarget, method: http.MethodPut, target: "/governance/performance-targets/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "update target invalid employee", fn: h.UpdatePerformanceTarget, method: http.MethodPut, target: "/governance/performance-targets/" + validID, body: `{"employee_id":"bad"}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete target invalid id", fn: h.DeletePerformanceTarget, method: http.MethodDelete, target: "/governance/performance-targets/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "create evidence invalid json", fn: h.CreateEvidenceItem, method: http.MethodPost, target: "/governance/evidence", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create evidence invalid owner", fn: h.CreateEvidenceItem, method: http.MethodPost, target: "/governance/evidence", body: `{"owner_unit_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update evidence invalid id", fn: h.UpdateEvidenceItem, method: http.MethodPut, target: "/governance/evidence/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update evidence invalid json", fn: h.UpdateEvidenceItem, method: http.MethodPut, target: "/governance/evidence/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "update evidence invalid owner", fn: h.UpdateEvidenceItem, method: http.MethodPut, target: "/governance/evidence/" + validID, body: `{"owner_unit_id":"bad"}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete evidence invalid id", fn: h.DeleteEvidenceItem, method: http.MethodDelete, target: "/governance/evidence/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "create compliance invalid json", fn: h.CreateComplianceAction, method: http.MethodPost, target: "/governance/compliance", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create compliance invalid completed at", fn: h.CreateComplianceAction, method: http.MethodPost, target: "/governance/compliance", body: `{"completed_at":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "update compliance invalid id", fn: h.UpdateComplianceAction, method: http.MethodPut, target: "/governance/compliance/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update compliance invalid json", fn: h.UpdateComplianceAction, method: http.MethodPut, target: "/governance/compliance/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete compliance invalid id", fn: h.DeleteComplianceAction, method: http.MethodDelete, target: "/governance/compliance/bad", id: "bad", wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(tt.method, tt.target, tt.body)
			if tt.id != "" {
				req = withRouteParam(req, "id", tt.id)
			}
			tt.fn(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
