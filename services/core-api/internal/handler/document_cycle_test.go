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

func TestDocumentCycleParseCatalogRequest(t *testing.T) {
	handler := &DocumentCycle{}
	body := `{
		"code":"RKT",
		"title":"Rencana Kerja Tahunan",
		"frequency":"annual",
		"domain_area":"governance",
		"external_system":"EDM",
		"snp_standard":"skl",
		"regulation_ref":"Permenag",
		"default_owner_unit_id":"01000000-0000-0000-0000-000000000000",
		"default_responsible_employee_id":"02000000-0000-0000-0000-000000000000",
		"default_verifier_employee_id":"03000000-0000-0000-0000-000000000000",
		"description":"Lampiran EDM",
		"sort_order":7
	}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/document-cycle/catalogs", strings.NewReader(body))

	got, ok := handler.parseCatalogRequest(rec, req, false)
	if !ok {
		t.Fatalf("parseCatalogRequest(create) ok = false; body=%s", rec.Body.String())
	}
	if got.Code != "RKT" || got.Title != "Rencana Kerja Tahunan" || got.Frequency != "annual" || got.DomainArea != "governance" {
		t.Fatalf("parseCatalogRequest(create) = %+v, want mapped fields", got)
	}
	if got.DeadlineDaysAfterPeriod != 5 || got.ReminderDaysBeforeDue != 3 || !got.IsActive || got.SortOrder != 7 {
		t.Fatalf("parseCatalogRequest(create) defaults = %+v, want deadline 5 reminder 3 active true sort 7", got)
	}
	if !got.DefaultOwnerUnitID.Valid || !got.DefaultResponsibleEmployeeID.Valid || !got.DefaultVerifierEmployeeID.Valid {
		t.Fatalf("parseCatalogRequest(create) ids = %v/%v/%v, want valid ids", got.DefaultOwnerUnitID, got.DefaultResponsibleEmployeeID, got.DefaultVerifierEmployeeID)
	}

	updateBody := `{
		"code":"BAP",
		"title":"Berita Acara",
		"frequency":"monthly",
		"is_active":false
	}`
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/document-cycle/catalogs/1", strings.NewReader(updateBody))
	got, ok = handler.parseCatalogRequest(rec, req, true)
	if !ok {
		t.Fatalf("parseCatalogRequest(update) ok = false; body=%s", rec.Body.String())
	}
	if got.Code != "BAP" || got.DeadlineDaysAfterPeriod != 0 || got.ReminderDaysBeforeDue != 0 || got.IsActive {
		t.Fatalf("parseCatalogRequest(update) = %+v, want no create defaults and inactive", got)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/document-cycle/catalogs", strings.NewReader(`{`))
	if _, ok := handler.parseCatalogRequest(rec, req, false); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseCatalogRequest(invalid json) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/document-cycle/catalogs", strings.NewReader(`{"default_owner_unit_id":"bad"}`))
	if _, ok := handler.parseCatalogRequest(rec, req, false); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseCatalogRequest(invalid uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
	for _, field := range []string{"default_responsible_employee_id", "default_verifier_employee_id"} {
		t.Run("invalid "+field, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/document-cycle/catalogs", strings.NewReader(`{"`+field+`":"bad"}`))
			if _, ok := handler.parseCatalogRequest(rec, req, false); ok || rec.Code != http.StatusBadRequest {
				t.Fatalf("parseCatalogRequest(%s) ok/status = %v/%d, want false/400", field, ok, rec.Code)
			}
		})
	}
}

func TestDocumentCycleParseObligationRequest(t *testing.T) {
	handler := &DocumentCycle{}
	id := handlerTestUUID(9)
	body := `{
		"due_date":"2026-05-10",
		"reminder_date":"2026-05-07",
		"domain_area":"governance",
		"external_system":"EDM",
		"owner_unit_id":"01000000-0000-0000-0000-000000000000",
		"responsible_employee_id":"02000000-0000-0000-0000-000000000000",
		"verifier_employee_id":"03000000-0000-0000-0000-000000000000",
		"governance_document_id":"04000000-0000-0000-0000-000000000000",
		"work_plan_item_id":"05000000-0000-0000-0000-000000000000",
		"performance_target_id":"06000000-0000-0000-0000-000000000000",
		"evidence_item_id":"07000000-0000-0000-0000-000000000000",
		"compliance_action_id":"08000000-0000-0000-0000-000000000000",
		"archive_document_id":"09000000-0000-0000-0000-000000000000",
		"notes":"menunggu lampiran",
		"verification_notes":"cek kepala"
	}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/document-cycle/obligations/1", strings.NewReader(body))

	got, ok := handler.parseObligationRequest(rec, req, id)
	if !ok {
		t.Fatalf("parseObligationRequest() ok = false; body=%s", rec.Body.String())
	}
	if got.ID != id || got.DomainArea != "governance" || got.ExternalSystem != "EDM" || got.Notes != "menunggu lampiran" || got.VerificationNotes != "cek kepala" {
		t.Fatalf("parseObligationRequest() = %+v, want mapped fields", got)
	}
	if !got.DueDate.Valid || got.DueDate.Time.Format("2006-01-02") != "2026-05-10" {
		t.Fatalf("parseObligationRequest() due_date = %v, want 2026-05-10", got.DueDate)
	}
	validIDs := []pgtype.UUID{
		got.OwnerUnitID,
		got.ResponsibleEmployeeID,
		got.VerifierEmployeeID,
		got.GovernanceDocumentID,
		got.WorkPlanItemID,
		got.PerformanceTargetID,
		got.EvidenceItemID,
		got.ComplianceActionID,
		got.ArchiveDocumentID,
	}
	for i, parsedID := range validIDs {
		if !parsedID.Valid {
			t.Fatalf("parseObligationRequest() id[%d] invalid, want valid", i)
		}
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/document-cycle/obligations/1", strings.NewReader(`{`))
	if _, ok := handler.parseObligationRequest(rec, req, id); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseObligationRequest(invalid json) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/document-cycle/obligations/1", strings.NewReader(`{"due_date":"bad","reminder_date":"2026-05-07"}`))
	if _, ok := handler.parseObligationRequest(rec, req, id); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseObligationRequest(invalid date) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/document-cycle/obligations/1", strings.NewReader(`{"due_date":"2026-05-10","reminder_date":"2026-05-07","owner_unit_id":"bad"}`))
	if _, ok := handler.parseObligationRequest(rec, req, id); ok || rec.Code != http.StatusBadRequest {
		t.Fatalf("parseObligationRequest(invalid uuid) ok/status = %v/%d, want false/400", ok, rec.Code)
	}
	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "invalid reminder date", body: `{"due_date":"2026-05-10","reminder_date":"bad"}`},
		{name: "invalid responsible employee", body: `{"due_date":"2026-05-10","reminder_date":"2026-05-07","responsible_employee_id":"bad"}`},
		{name: "invalid verifier employee", body: `{"due_date":"2026-05-10","reminder_date":"2026-05-07","verifier_employee_id":"bad"}`},
		{name: "invalid governance document", body: `{"due_date":"2026-05-10","reminder_date":"2026-05-07","governance_document_id":"bad"}`},
		{name: "invalid work plan item", body: `{"due_date":"2026-05-10","reminder_date":"2026-05-07","work_plan_item_id":"bad"}`},
		{name: "invalid performance target", body: `{"due_date":"2026-05-10","reminder_date":"2026-05-07","performance_target_id":"bad"}`},
		{name: "invalid evidence item", body: `{"due_date":"2026-05-10","reminder_date":"2026-05-07","evidence_item_id":"bad"}`},
		{name: "invalid compliance action", body: `{"due_date":"2026-05-10","reminder_date":"2026-05-07","compliance_action_id":"bad"}`},
		{name: "invalid archive document", body: `{"due_date":"2026-05-10","reminder_date":"2026-05-07","archive_document_id":"bad"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/document-cycle/obligations/1", strings.NewReader(tc.body))
			if _, ok := handler.parseObligationRequest(rec, req, id); ok || rec.Code != http.StatusBadRequest {
				t.Fatalf("parseObligationRequest(%s) ok/status = %v/%d, want false/400", tc.name, ok, rec.Code)
			}
		})
	}
}

type fakeDocumentCycleHandlerService struct {
	statsYear int32
	statsRow  db.GetDocumentCycleStatsRow
	statsErr  error

	listCatalogsSearch     string
	listCatalogsFrequency  string
	listCatalogsActiveOnly bool
	listCatalogsRows       []db.ListDocumentCycleCatalogsRow
	listCatalogsErr        error

	createCatalogArg db.CreateDocumentCycleCatalogParams
	createCatalogRow db.DocumentCycleCatalog
	createCatalogErr error

	updateCatalogArg db.UpdateDocumentCycleCatalogParams
	updateCatalogRow db.DocumentCycleCatalog
	updateCatalogErr error

	deleteCatalogID  pgtype.UUID
	deleteCatalogErr error

	listObligationsArg  db.ListDocumentCycleObligationsParams
	listObligationsRows []db.ListDocumentCycleObligationsRow
	listObligationsErr  error

	queueVerifierID pgtype.UUID
	queuePeriodYear int32
	queueRows       []db.ListDocumentCycleObligationsRow
	queueErr        error

	generateActor pgtype.UUID
	generateYear  int32
	generateRow   service.DocumentCycleGenerateResult
	generateErr   error

	updateObligationActor pgtype.UUID
	updateObligationArg   db.UpdateDocumentCycleObligationParams
	updateObligationRow   db.DocumentCycleObligation
	updateObligationErr   error

	eventsObligationID pgtype.UUID
	eventsType         string
	eventsActor        string
	eventsRows         []db.ListDocumentCycleEventsByObligationRow
	eventsErr          error

	statusActor pgtype.UUID
	statusID    pgtype.UUID
	statusValue string
	statusNotes string
	statusRow   db.DocumentCycleObligation
	statusErr   error

	deleteObligationID  pgtype.UUID
	deleteObligationErr error
}

func (f *fakeDocumentCycleHandlerService) Stats(ctx context.Context, periodYear int32) (db.GetDocumentCycleStatsRow, error) {
	f.statsYear = periodYear
	return f.statsRow, f.statsErr
}

func (f *fakeDocumentCycleHandlerService) ListCatalogs(ctx context.Context, search, frequency string, activeOnly bool) ([]db.ListDocumentCycleCatalogsRow, error) {
	f.listCatalogsSearch = search
	f.listCatalogsFrequency = frequency
	f.listCatalogsActiveOnly = activeOnly
	return f.listCatalogsRows, f.listCatalogsErr
}

func (f *fakeDocumentCycleHandlerService) CreateCatalog(ctx context.Context, arg db.CreateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error) {
	f.createCatalogArg = arg
	if f.createCatalogErr != nil {
		return db.DocumentCycleCatalog{}, f.createCatalogErr
	}
	return f.createCatalogRow, nil
}

func (f *fakeDocumentCycleHandlerService) UpdateCatalog(ctx context.Context, arg db.UpdateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error) {
	f.updateCatalogArg = arg
	if f.updateCatalogErr != nil {
		return db.DocumentCycleCatalog{}, f.updateCatalogErr
	}
	return f.updateCatalogRow, nil
}

func (f *fakeDocumentCycleHandlerService) DeleteCatalog(ctx context.Context, id pgtype.UUID) error {
	f.deleteCatalogID = id
	return f.deleteCatalogErr
}

func (f *fakeDocumentCycleHandlerService) ListObligations(ctx context.Context, arg db.ListDocumentCycleObligationsParams) ([]db.ListDocumentCycleObligationsRow, error) {
	f.listObligationsArg = arg
	return f.listObligationsRows, f.listObligationsErr
}

func (f *fakeDocumentCycleHandlerService) ListVerificationQueue(ctx context.Context, verifierEmployeeID pgtype.UUID, periodYear int32) ([]db.ListDocumentCycleObligationsRow, error) {
	f.queueVerifierID = verifierEmployeeID
	f.queuePeriodYear = periodYear
	return f.queueRows, f.queueErr
}

func (f *fakeDocumentCycleHandlerService) GenerateYear(ctx context.Context, actorUserID pgtype.UUID, periodYear int32) (service.DocumentCycleGenerateResult, error) {
	f.generateActor = actorUserID
	f.generateYear = periodYear
	if f.generateErr != nil {
		return service.DocumentCycleGenerateResult{}, f.generateErr
	}
	return f.generateRow, nil
}

func (f *fakeDocumentCycleHandlerService) UpdateObligation(ctx context.Context, actorUserID pgtype.UUID, arg db.UpdateDocumentCycleObligationParams) (db.DocumentCycleObligation, error) {
	f.updateObligationActor = actorUserID
	f.updateObligationArg = arg
	if f.updateObligationErr != nil {
		return db.DocumentCycleObligation{}, f.updateObligationErr
	}
	return f.updateObligationRow, nil
}

func (f *fakeDocumentCycleHandlerService) ListEvents(ctx context.Context, obligationID pgtype.UUID, eventType, actor string) ([]db.ListDocumentCycleEventsByObligationRow, error) {
	f.eventsObligationID = obligationID
	f.eventsType = eventType
	f.eventsActor = actor
	return f.eventsRows, f.eventsErr
}

func (f *fakeDocumentCycleHandlerService) UpdateObligationStatus(ctx context.Context, actorUserID pgtype.UUID, id pgtype.UUID, status, notes string) (db.DocumentCycleObligation, error) {
	f.statusActor = actorUserID
	f.statusID = id
	f.statusValue = status
	f.statusNotes = notes
	if f.statusErr != nil {
		return db.DocumentCycleObligation{}, f.statusErr
	}
	return f.statusRow, nil
}

func (f *fakeDocumentCycleHandlerService) DeleteObligation(ctx context.Context, id pgtype.UUID) error {
	f.deleteObligationID = id
	return f.deleteObligationErr
}

func staffDocumentCycleRequest(method, target, body string, employeeID pgtype.UUID) *http.Request {
	claims := jwt.MapClaims{"roles": []any{"staf"}}
	if employeeID.Valid {
		claims["eid"] = employeeID.String()
	}
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, claims)
}

func TestDocumentCycleHandlersForwardQueriesAndBodies(t *testing.T) {
	catalogID := handlerTestUUID(71)
	obligationID := handlerTestUUID(72)
	ownerUnitID := handlerTestUUID(73)
	responsibleEmployeeID := handlerTestUUID(74)
	verifierEmployeeID := handlerTestUUID(75)
	actorID := handlerTestUUID(1)
	svc := &fakeDocumentCycleHandlerService{
		statsRow:         db.GetDocumentCycleStatsRow{ActiveCatalogs: 2},
		createCatalogRow: db.DocumentCycleCatalog{ID: catalogID, Code: "RKT"},
		updateCatalogRow: db.DocumentCycleCatalog{ID: catalogID, Code: "BAP"},
		listObligationsRows: []db.ListDocumentCycleObligationsRow{
			{ID: obligationID, CatalogCode: "RKT"},
		},
		generateRow:         service.DocumentCycleGenerateResult{PeriodYear: 2027, Generated: 4, Inserted: 3},
		updateObligationRow: db.DocumentCycleObligation{ID: obligationID, Status: "draft"},
		statusRow:           db.DocumentCycleObligation{ID: obligationID, Status: "waiting_verification"},
		eventsRows:          []db.ListDocumentCycleEventsByObligationRow{{ID: handlerTestUUID(76), ObligationID: obligationID, EventType: "updated"}},
		listCatalogsRows:    []db.ListDocumentCycleCatalogsRow{{ID: catalogID, Code: "RKT"}},
		queueRows:           []db.ListDocumentCycleObligationsRow{{ID: obligationID, VerifierEmployeeID: verifierEmployeeID}},
		deleteCatalogID:     pgtype.UUID{},
		deleteObligationID:  pgtype.UUID{},
		deleteCatalogErr:    nil,
		deleteObligationErr: nil,
		listCatalogsErr:     nil,
		listObligationsErr:  nil,
		updateObligationErr: nil,
		statusErr:           nil,
		eventsErr:           nil,
		createCatalogErr:    nil,
		updateCatalogErr:    nil,
		queueErr:            nil,
		statsErr:            nil,
		generateErr:         nil,
	}
	h := &DocumentCycle{svc: svc}

	rec := httptest.NewRecorder()
	h.Stats(rec, adminRequest(http.MethodGet, "/document-cycle/stats?period_year=2026", ""))
	if rec.Code != http.StatusOK || svc.statsYear != 2026 {
		t.Fatalf("Stats() status/year = %d/%d, want 200/2026", rec.Code, svc.statsYear)
	}

	rec = httptest.NewRecorder()
	h.ListCatalogs(rec, adminRequest(http.MethodGet, "/document-cycle/catalogs?search=rapat&frequency=monthly&active_only=true", ""))
	if rec.Code != http.StatusOK || svc.listCatalogsSearch != "rapat" || svc.listCatalogsFrequency != "monthly" || !svc.listCatalogsActiveOnly {
		t.Fatalf("ListCatalogs() status/args = %d/%q/%q/%v", rec.Code, svc.listCatalogsSearch, svc.listCatalogsFrequency, svc.listCatalogsActiveOnly)
	}

	rec = httptest.NewRecorder()
	h.CreateCatalog(rec, adminRequest(http.MethodPost, "/document-cycle/catalogs", `{"code":"rkt","title":"Rencana Kerja","frequency":"annual"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateCatalog() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if svc.createCatalogArg.Code != "rkt" || svc.createCatalogArg.DeadlineDaysAfterPeriod != 5 || svc.createCatalogArg.ReminderDaysBeforeDue != 3 || !svc.createCatalogArg.IsActive {
		t.Fatalf("CreateCatalog() arg = %+v, want parsed body with create defaults", svc.createCatalogArg)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodPut, "/document-cycle/catalogs/"+catalogID.String(), `{"code":"BAP","title":"Berita Acara","frequency":"monthly","is_active":false}`), "id", catalogID.String())
	h.UpdateCatalog(rec, req)
	if rec.Code != http.StatusOK || svc.updateCatalogArg.ID != catalogID || svc.updateCatalogArg.IsActive {
		t.Fatalf("UpdateCatalog() status/arg = %d/%+v, want 200 and inactive catalog", rec.Code, svc.updateCatalogArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/document-cycle/catalogs/"+catalogID.String(), ""), "id", catalogID.String())
	h.DeleteCatalog(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteCatalogID != catalogID {
		t.Fatalf("DeleteCatalog() status/id = %d/%v, want 204/%v", rec.Code, svc.deleteCatalogID, catalogID)
	}

	rec = httptest.NewRecorder()
	url := "/document-cycle/obligations?search=rkt&status=draft&frequency=annual&domain_area=governance&external_system=edm&period_year=2026&owner_unit_id=" + ownerUnitID.String() + "&responsible_employee_id=" + responsibleEmployeeID.String() + "&verifier_employee_id=" + verifierEmployeeID.String() + "&reminder_only=true"
	h.ListObligations(rec, adminRequest(http.MethodGet, url, ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListObligations() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.listObligationsArg.Search != "rkt" || svc.listObligationsArg.Status != "draft" || svc.listObligationsArg.PeriodYear != 2026 || !svc.listObligationsArg.ReminderOnly {
		t.Fatalf("ListObligations() arg = %+v, want query filters", svc.listObligationsArg)
	}
	if svc.listObligationsArg.OwnerUnitID != ownerUnitID || svc.listObligationsArg.ResponsibleEmployeeID != responsibleEmployeeID || svc.listObligationsArg.VerifierEmployeeID != verifierEmployeeID {
		t.Fatalf("ListObligations() uuid filters = %+v, want parsed filters", svc.listObligationsArg)
	}

	rec = httptest.NewRecorder()
	h.GenerateYear(rec, adminRequest(http.MethodPost, "/document-cycle/generate", `{"period_year":2027}`))
	if rec.Code != http.StatusCreated || svc.generateYear != 2027 || svc.generateActor != actorID {
		t.Fatalf("GenerateYear() status/year/actor = %d/%d/%v, want 201/2027/%v", rec.Code, svc.generateYear, svc.generateActor, actorID)
	}

	obligationBody := `{
		"due_date":"2026-05-20",
		"reminder_date":"2026-05-15",
		"domain_area":"governance",
		"external_system":"edm",
		"owner_unit_id":"` + ownerUnitID.String() + `",
		"responsible_employee_id":"` + responsibleEmployeeID.String() + `",
		"verifier_employee_id":"` + verifierEmployeeID.String() + `",
		"notes":"lampiran rapat",
		"verification_notes":"cek waka"
	}`
	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPut, "/document-cycle/obligations/"+obligationID.String(), obligationBody), "id", obligationID.String())
	h.UpdateObligation(rec, req)
	if rec.Code != http.StatusOK || svc.updateObligationArg.ID != obligationID || svc.updateObligationActor != actorID {
		t.Fatalf("UpdateObligation() status/id/actor = %d/%v/%v, want 200/%v/%v", rec.Code, svc.updateObligationArg.ID, svc.updateObligationActor, obligationID, actorID)
	}
	if svc.updateObligationArg.Notes != "lampiran rapat" || !svc.updateObligationArg.DueDate.Valid || !svc.updateObligationArg.ReminderDate.Valid {
		t.Fatalf("UpdateObligation() arg = %+v, want parsed obligation body", svc.updateObligationArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodGet, "/document-cycle/obligations/"+obligationID.String()+"/events?event_type=updated&actor=kepala", ""), "id", obligationID.String())
	h.ListEvents(rec, req)
	if rec.Code != http.StatusOK || svc.eventsObligationID != obligationID || svc.eventsType != "updated" || svc.eventsActor != "kepala" {
		t.Fatalf("ListEvents() status/args = %d/%v/%q/%q", rec.Code, svc.eventsObligationID, svc.eventsType, svc.eventsActor)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodPost, "/document-cycle/obligations/"+obligationID.String()+"/status", `{"status":"waiting_verification","notes":"siap dicek"}`), "id", obligationID.String())
	h.UpdateObligationStatus(rec, req)
	if rec.Code != http.StatusOK || svc.statusID != obligationID || svc.statusValue != "waiting_verification" || svc.statusNotes != "siap dicek" || svc.statusActor != actorID {
		t.Fatalf("UpdateObligationStatus() status/args = %d/%v/%q/%q/%v", rec.Code, svc.statusID, svc.statusValue, svc.statusNotes, svc.statusActor)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/document-cycle/obligations/"+obligationID.String(), ""), "id", obligationID.String())
	h.DeleteObligation(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteObligationID != obligationID {
		t.Fatalf("DeleteObligation() status/id = %d/%v, want 204/%v", rec.Code, svc.deleteObligationID, obligationID)
	}
}

func TestDocumentCycleVerificationQueueAdminGuruRemainsAdminWide(t *testing.T) {
	svc := &fakeDocumentCycleHandlerService{
		queueRows: []db.ListDocumentCycleObligationsRow{{ID: handlerTestUUID(81)}},
	}
	h := &DocumentCycle{svc: svc}

	req := httptest.NewRequest(http.MethodGet, "/document-cycle/verification-queue?period_year=2026", nil)
	req = withClaims(req, jwt.MapClaims{
		"roles": []any{"admin", "guru"},
	})
	rec := httptest.NewRecorder()

	h.ListVerificationQueue(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ListVerificationQueue(admin+guru) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.queuePeriodYear != 2026 {
		t.Fatalf("ListVerificationQueue(admin+guru) periodYear = %d, want 2026", svc.queuePeriodYear)
	}
	if svc.queueVerifierID.Valid {
		t.Fatalf("ListVerificationQueue(admin+guru) verifierEmployeeID = %v, want invalid admin-wide scope", svc.queueVerifierID)
	}
}

func TestDocumentCycleVerificationQueueRoleScope(t *testing.T) {
	employeeID := handlerTestUUID(81)
	requestedVerifierID := handlerTestUUID(82)
	svc := &fakeDocumentCycleHandlerService{}
	h := &DocumentCycle{svc: svc}

	rec := httptest.NewRecorder()
	h.ListVerificationQueue(rec, staffDocumentCycleRequest(http.MethodGet, "/document-cycle/verification-queue?period_year=2026", "", employeeID))
	if rec.Code != http.StatusOK || svc.queueVerifierID != employeeID || svc.queuePeriodYear != 2026 {
		t.Fatalf("staff queue status/args = %d/%v/%d, want 200/%v/2026", rec.Code, svc.queueVerifierID, svc.queuePeriodYear, employeeID)
	}

	rec = httptest.NewRecorder()
	h.ListVerificationQueue(rec, staffDocumentCycleRequest(http.MethodGet, "/document-cycle/verification-queue?all=true", "", employeeID))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("staff all queue status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ListVerificationQueue(rec, staffDocumentCycleRequest(http.MethodGet, "/document-cycle/verification-queue?verifier_employee_id="+requestedVerifierID.String(), "", employeeID))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("staff requested verifier status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ListVerificationQueue(rec, staffDocumentCycleRequest(http.MethodGet, "/document-cycle/verification-queue", "", pgtype.UUID{}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("staff missing eid status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ListVerificationQueue(rec, adminRequest(http.MethodGet, "/document-cycle/verification-queue?all=true&period_year=2027", ""))
	if rec.Code != http.StatusOK || svc.queueVerifierID.Valid || svc.queuePeriodYear != 2027 {
		t.Fatalf("admin all queue status/args = %d/%v/%d, want 200/null/2027", rec.Code, svc.queueVerifierID, svc.queuePeriodYear)
	}

	rec = httptest.NewRecorder()
	h.ListVerificationQueue(rec, adminRequest(http.MethodGet, "/document-cycle/verification-queue?verifier_employee_id="+requestedVerifierID.String(), ""))
	if rec.Code != http.StatusOK || svc.queueVerifierID != requestedVerifierID {
		t.Fatalf("admin requested verifier status/arg = %d/%v, want 200/%v", rec.Code, svc.queueVerifierID, requestedVerifierID)
	}

	rec = httptest.NewRecorder()
	h.ListVerificationQueue(rec, adminRequest(http.MethodGet, "/document-cycle/verification-queue?verifier_employee_id=bad", ""))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("admin invalid requested verifier status = %d, want 400", rec.Code)
	}
}

func TestDocumentCycleMutationHandlersWriteAuditEvents(t *testing.T) {
	catalogID := handlerTestUUID(83)
	obligationID := handlerTestUUID(84)
	actorID := handlerTestUUID(85)
	svc := &fakeDocumentCycleHandlerService{
		createCatalogRow: db.DocumentCycleCatalog{
			ID:             catalogID,
			Code:           "RKT",
			Title:          "Rencana Kerja Tahunan",
			Frequency:      "annual",
			DomainArea:     "governance",
			ExternalSystem: "edm",
		},
		updateCatalogRow: db.DocumentCycleCatalog{
			ID:             catalogID,
			Code:           "BAP",
			Title:          "Berita Acara Pemeriksaan",
			Frequency:      "monthly",
			DomainArea:     "governance",
			ExternalSystem: "simpatika",
		},
		generateRow: service.DocumentCycleGenerateResult{
			PeriodYear: 2027,
			Generated:  4,
			Inserted:   3,
		},
		updateObligationRow: db.DocumentCycleObligation{
			ID:             obligationID,
			PeriodYear:     2027,
			Status:         "draft",
			DomainArea:     "governance",
			ExternalSystem: "edm",
		},
		statusRow: db.DocumentCycleObligation{
			ID:         obligationID,
			PeriodYear: 2027,
			Status:     "waiting_verification",
		},
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &DocumentCycle{svc: svc, audit: audit}

	adminAuditReq := func(method, target, body string) *http.Request {
		return withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"admin"},
			"uid":   actorID.String(),
			"sub":   actorID.String(),
			"usr":   "operator.dokumen",
			"ssid":  "sess-doc-cycle-1",
		})
	}

	rec := httptest.NewRecorder()
	h.CreateCatalog(rec, adminAuditReq(http.MethodPost, "/document-cycle/catalogs", `{"code":"rkt","title":"Rencana Kerja Tahunan","frequency":"annual","domain_area":"governance","external_system":"edm"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateCatalog() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminAuditReq(http.MethodPut, "/document-cycle/catalogs/"+catalogID.String(), `{"code":"BAP","title":"Berita Acara Pemeriksaan","frequency":"monthly","domain_area":"governance","external_system":"simpatika"}`), "id", catalogID.String())
	h.UpdateCatalog(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateCatalog() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminAuditReq(http.MethodDelete, "/document-cycle/catalogs/"+catalogID.String(), ""), "id", catalogID.String())
	h.DeleteCatalog(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteCatalog() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GenerateYear(rec, adminAuditReq(http.MethodPost, "/document-cycle/generate", `{"period_year":2027}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("GenerateYear() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminAuditReq(http.MethodPut, "/document-cycle/obligations/"+obligationID.String(), `{"due_date":"2027-05-20","reminder_date":"2027-05-15","domain_area":"governance","external_system":"edm"}`), "id", obligationID.String())
	h.UpdateObligation(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateObligation() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminAuditReq(http.MethodPost, "/document-cycle/obligations/"+obligationID.String()+"/status", `{"status":"waiting_verification","notes":"siap diverifikasi"}`), "id", obligationID.String())
	h.UpdateObligationStatus(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateObligationStatus() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminAuditReq(http.MethodDelete, "/document-cycle/obligations/"+obligationID.String(), ""), "id", obligationID.String())
	h.DeleteObligation(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteObligation() status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	if len(audit.entries) != 7 {
		t.Fatalf("audit entries = %d, want 7", len(audit.entries))
	}
	if audit.entries[0].Action != "DOCUMENT_CYCLE_CATALOG_CREATE" || audit.entries[3].Action != "DOCUMENT_CYCLE_YEAR_GENERATE" || audit.entries[6].Action != "DOCUMENT_CYCLE_OBLIGATION_DELETE" {
		t.Fatalf("audit actions = %#v", []string{
			audit.entries[0].Action,
			audit.entries[1].Action,
			audit.entries[2].Action,
			audit.entries[3].Action,
			audit.entries[4].Action,
			audit.entries[5].Action,
			audit.entries[6].Action,
		})
	}

	catalogMeta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if catalogMeta["code"] != "RKT" || catalogMeta["title"] != "Rencana Kerja Tahunan" || catalogMeta["username"] != "operator.dokumen" {
		t.Fatalf("catalog create metadata = %+v", catalogMeta)
	}
	generateMeta := mustAuditMetadataMap(t, audit.entries[3].Metadata)
	if generateMeta["period_year"] != float64(2027) || generateMeta["inserted"] != float64(3) {
		t.Fatalf("generate metadata = %+v", generateMeta)
	}
	deleteMeta := mustAuditMetadataMap(t, audit.entries[6].Metadata)
	if deleteMeta["deleted_by"] != "operator.dokumen" {
		t.Fatalf("delete metadata = %+v", deleteMeta)
	}
}

func TestDocumentCycleHandlersForbidMissingClaims(t *testing.T) {
	h := &DocumentCycle{}
	id := handlerTestUUID(89).String()
	tests := []struct {
		name   string
		fn     http.HandlerFunc
		method string
		target string
		body   string
		id     string
	}{
		{name: "stats", fn: h.Stats, method: http.MethodGet, target: "/document-cycle/stats"},
		{name: "list catalogs", fn: h.ListCatalogs, method: http.MethodGet, target: "/document-cycle/catalogs"},
		{name: "create catalog", fn: h.CreateCatalog, method: http.MethodPost, target: "/document-cycle/catalogs", body: `{}`},
		{name: "update catalog", fn: h.UpdateCatalog, method: http.MethodPut, target: "/document-cycle/catalogs/" + id, body: `{}`, id: id},
		{name: "delete catalog", fn: h.DeleteCatalog, method: http.MethodDelete, target: "/document-cycle/catalogs/" + id, id: id},
		{name: "list obligations", fn: h.ListObligations, method: http.MethodGet, target: "/document-cycle/obligations"},
		{name: "verification queue", fn: h.ListVerificationQueue, method: http.MethodGet, target: "/document-cycle/verification-queue"},
		{name: "generate year", fn: h.GenerateYear, method: http.MethodPost, target: "/document-cycle/generate", body: `{}`},
		{name: "update obligation", fn: h.UpdateObligation, method: http.MethodPut, target: "/document-cycle/obligations/" + id, body: `{}`, id: id},
		{name: "list events", fn: h.ListEvents, method: http.MethodGet, target: "/document-cycle/obligations/" + id + "/events", id: id},
		{name: "update status", fn: h.UpdateObligationStatus, method: http.MethodPost, target: "/document-cycle/obligations/" + id + "/status", body: `{}`, id: id},
		{name: "delete obligation", fn: h.DeleteObligation, method: http.MethodDelete, target: "/document-cycle/obligations/" + id, id: id},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body))
			if tt.id != "" {
				req = withRouteParam(req, "id", tt.id)
			}
			tt.fn(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("%s status = %d, want 403; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDocumentCycleRoleHelpers(t *testing.T) {
	adminReq := adminRequest(http.MethodGet, "/document-cycle/verification-queue", "")
	if !documentCycleHasRole(adminReq, "admin") {
		t.Fatal("documentCycleHasRole(admin) = false, want true")
	}
	if id, ok := documentCycleEmployeeID(adminReq); ok || id.Valid {
		t.Fatalf("documentCycleEmployeeID(admin) = %v/%v, want no employee scope", id, ok)
	}

	employeeID := handlerTestUUID(90)
	staffReq := staffDocumentCycleRequest(http.MethodGet, "/document-cycle/verification-queue", "", employeeID)
	if !documentCycleHasRole(staffReq, "staf") {
		t.Fatal("documentCycleHasRole(staf) = false, want true")
	}
	if id, ok := documentCycleEmployeeID(staffReq); !ok || id != employeeID {
		t.Fatalf("documentCycleEmployeeID(staf) = %v/%v, want %v/true", id, ok, employeeID)
	}

	invalidEIDReq := withClaims(httptest.NewRequest(http.MethodGet, "/document-cycle/verification-queue", nil), jwt.MapClaims{
		"roles": []any{"staf"},
		"eid":   "bad",
	})
	if id, ok := documentCycleEmployeeID(invalidEIDReq); ok || id.Valid {
		t.Fatalf("documentCycleEmployeeID(invalid eid) = %v/%v, want empty false", id, ok)
	}

	plainReq := httptest.NewRequest(http.MethodGet, "/document-cycle/verification-queue", nil)
	if documentCycleHasRole(plainReq, "admin") {
		t.Fatal("documentCycleHasRole(no claims) = true, want false")
	}
	if id, ok := documentCycleEmployeeID(plainReq); ok || id.Valid {
		t.Fatalf("documentCycleEmployeeID(no claims) = %v/%v, want empty false", id, ok)
	}
}

func TestDocumentCycleHandlersRejectInvalidInputs(t *testing.T) {
	h := &DocumentCycle{}
	validID := handlerTestUUID(91).String()
	tests := []struct {
		name       string
		fn         http.HandlerFunc
		method     string
		target     string
		body       string
		id         string
		wantStatus int
	}{
		{name: "create catalog invalid json", fn: h.CreateCatalog, method: http.MethodPost, target: "/document-cycle/catalogs", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "update catalog invalid id", fn: h.UpdateCatalog, method: http.MethodPut, target: "/document-cycle/catalogs/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "delete catalog invalid id", fn: h.DeleteCatalog, method: http.MethodDelete, target: "/document-cycle/catalogs/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "list obligations invalid owner", fn: h.ListObligations, method: http.MethodGet, target: "/document-cycle/obligations?owner_unit_id=bad", wantStatus: http.StatusBadRequest},
		{name: "list obligations invalid responsible", fn: h.ListObligations, method: http.MethodGet, target: "/document-cycle/obligations?responsible_employee_id=bad", wantStatus: http.StatusBadRequest},
		{name: "list obligations invalid verifier", fn: h.ListObligations, method: http.MethodGet, target: "/document-cycle/obligations?verifier_employee_id=bad", wantStatus: http.StatusBadRequest},
		{name: "generate invalid json", fn: h.GenerateYear, method: http.MethodPost, target: "/document-cycle/generate", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "update obligation invalid id", fn: h.UpdateObligation, method: http.MethodPut, target: "/document-cycle/obligations/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update obligation invalid json", fn: h.UpdateObligation, method: http.MethodPut, target: "/document-cycle/obligations/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "events invalid id", fn: h.ListEvents, method: http.MethodGet, target: "/document-cycle/obligations/bad/events", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "status invalid id", fn: h.UpdateObligationStatus, method: http.MethodPost, target: "/document-cycle/obligations/bad/status", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "status invalid json", fn: h.UpdateObligationStatus, method: http.MethodPost, target: "/document-cycle/obligations/" + validID + "/status", body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete obligation invalid id", fn: h.DeleteObligation, method: http.MethodDelete, target: "/document-cycle/obligations/bad", id: "bad", wantStatus: http.StatusBadRequest},
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

func TestDocumentCycleHandlersMapServiceErrors(t *testing.T) {
	catalogID := handlerTestUUID(92)
	obligationID := handlerTestUUID(93)
	serviceErr := errors.New("service failed")
	validCatalogBody := `{"code":"RKT","title":"Rencana Kerja","frequency":"annual"}`
	validObligationBody := `{"due_date":"2026-05-10","reminder_date":"2026-05-07"}`
	tests := []struct {
		name       string
		svc        *fakeDocumentCycleHandlerService
		call       func(*DocumentCycle, *httptest.ResponseRecorder)
		wantStatus int
	}{
		{
			name: "stats",
			svc:  &fakeDocumentCycleHandlerService{statsErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				h.Stats(rec, adminRequest(http.MethodGet, "/document-cycle/stats", ""))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "list catalogs",
			svc:  &fakeDocumentCycleHandlerService{listCatalogsErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				h.ListCatalogs(rec, adminRequest(http.MethodGet, "/document-cycle/catalogs", ""))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "create catalog",
			svc:  &fakeDocumentCycleHandlerService{createCatalogErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				h.CreateCatalog(rec, adminRequest(http.MethodPost, "/document-cycle/catalogs", validCatalogBody))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "update catalog",
			svc:  &fakeDocumentCycleHandlerService{updateCatalogErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				req := withRouteParam(adminRequest(http.MethodPut, "/document-cycle/catalogs/"+catalogID.String(), validCatalogBody), "id", catalogID.String())
				h.UpdateCatalog(rec, req)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "delete catalog",
			svc:  &fakeDocumentCycleHandlerService{deleteCatalogErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				req := withRouteParam(adminRequest(http.MethodDelete, "/document-cycle/catalogs/"+catalogID.String(), ""), "id", catalogID.String())
				h.DeleteCatalog(rec, req)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "list obligations",
			svc:  &fakeDocumentCycleHandlerService{listObligationsErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				h.ListObligations(rec, adminRequest(http.MethodGet, "/document-cycle/obligations", ""))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "verification queue",
			svc:  &fakeDocumentCycleHandlerService{queueErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				h.ListVerificationQueue(rec, adminRequest(http.MethodGet, "/document-cycle/verification-queue", ""))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "generate year",
			svc:  &fakeDocumentCycleHandlerService{generateErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				h.GenerateYear(rec, adminRequest(http.MethodPost, "/document-cycle/generate", `{"period_year":2026}`))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "update obligation",
			svc:  &fakeDocumentCycleHandlerService{updateObligationErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				req := withRouteParam(adminRequest(http.MethodPut, "/document-cycle/obligations/"+obligationID.String(), validObligationBody), "id", obligationID.String())
				h.UpdateObligation(rec, req)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "list events",
			svc:  &fakeDocumentCycleHandlerService{eventsErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				req := withRouteParam(adminRequest(http.MethodGet, "/document-cycle/obligations/"+obligationID.String()+"/events", ""), "id", obligationID.String())
				h.ListEvents(rec, req)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "update status",
			svc:  &fakeDocumentCycleHandlerService{statusErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				req := withRouteParam(adminRequest(http.MethodPost, "/document-cycle/obligations/"+obligationID.String()+"/status", `{"status":"submitted"}`), "id", obligationID.String())
				h.UpdateObligationStatus(rec, req)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "delete obligation",
			svc:  &fakeDocumentCycleHandlerService{deleteObligationErr: serviceErr},
			call: func(h *DocumentCycle, rec *httptest.ResponseRecorder) {
				req := withRouteParam(adminRequest(http.MethodDelete, "/document-cycle/obligations/"+obligationID.String(), ""), "id", obligationID.String())
				h.DeleteObligation(rec, req)
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &DocumentCycle{svc: tt.svc}
			rec := httptest.NewRecorder()
			tt.call(h, rec)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
