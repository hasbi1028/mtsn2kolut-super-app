package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeDocumentCycleStore struct {
	statsPeriod int32
	statsRow    db.GetDocumentCycleStatsRow
	statsErr    error

	listCatalogsArg  db.ListDocumentCycleCatalogsParams
	listCatalogsRows []db.ListDocumentCycleCatalogsRow
	listCatalogsErr  error

	createCatalogArg db.CreateDocumentCycleCatalogParams
	createCatalogRow db.DocumentCycleCatalog
	createCatalogErr error

	updateCatalogArg db.UpdateDocumentCycleCatalogParams
	updateCatalogRow db.DocumentCycleCatalog
	updateCatalogErr error

	deleteCatalogID  pgtype.UUID
	deleteCatalogErr error

	statusLookupID pgtype.UUID
	statusBefore   string
	statusErr      error

	updateObligationArg db.UpdateDocumentCycleObligationParams
	updateObligationRow db.DocumentCycleObligation
	updateObligationErr error

	readinessID  pgtype.UUID
	readinessRow db.GetDocumentCycleObligationCompletionReadinessRow
	readinessErr error

	updateStatusArg db.UpdateDocumentCycleObligationStatusParams
	updateStatusRow db.DocumentCycleObligation
	updateStatusErr error

	eventArg db.CreateDocumentCycleEventParams
	eventErr error

	listObligationsArg  db.ListDocumentCycleObligationsParams
	listObligationsRows []db.ListDocumentCycleObligationsRow
	listObligationsErr  error

	generateArg db.GenerateDocumentCycleYearObligationsParams
	generateRow db.GenerateDocumentCycleYearObligationsRow
	generateErr error

	listEventsArg  db.ListDocumentCycleEventsByObligationParams
	listEventsRows []db.ListDocumentCycleEventsByObligationRow
	listEventsErr  error

	deleteObligationID  pgtype.UUID
	deleteObligationErr error
}

func (f *fakeDocumentCycleStore) GetDocumentCycleStats(ctx context.Context, periodYear int32) (db.GetDocumentCycleStatsRow, error) {
	f.statsPeriod = periodYear
	return f.statsRow, f.statsErr
}

func (f *fakeDocumentCycleStore) ListDocumentCycleCatalogs(ctx context.Context, arg db.ListDocumentCycleCatalogsParams) ([]db.ListDocumentCycleCatalogsRow, error) {
	f.listCatalogsArg = arg
	return f.listCatalogsRows, f.listCatalogsErr
}

func (f *fakeDocumentCycleStore) CreateDocumentCycleCatalog(ctx context.Context, arg db.CreateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error) {
	f.createCatalogArg = arg
	return f.createCatalogRow, f.createCatalogErr
}

func (f *fakeDocumentCycleStore) UpdateDocumentCycleCatalog(ctx context.Context, arg db.UpdateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error) {
	f.updateCatalogArg = arg
	return f.updateCatalogRow, f.updateCatalogErr
}

func (f *fakeDocumentCycleStore) DeleteDocumentCycleCatalog(ctx context.Context, id pgtype.UUID) error {
	f.deleteCatalogID = id
	return f.deleteCatalogErr
}

func (f *fakeDocumentCycleStore) ListDocumentCycleObligations(ctx context.Context, arg db.ListDocumentCycleObligationsParams) ([]db.ListDocumentCycleObligationsRow, error) {
	f.listObligationsArg = arg
	return f.listObligationsRows, f.listObligationsErr
}

func (f *fakeDocumentCycleStore) GenerateDocumentCycleYearObligations(ctx context.Context, arg db.GenerateDocumentCycleYearObligationsParams) (db.GenerateDocumentCycleYearObligationsRow, error) {
	f.generateArg = arg
	return f.generateRow, f.generateErr
}

func (f *fakeDocumentCycleStore) UpdateDocumentCycleObligation(ctx context.Context, arg db.UpdateDocumentCycleObligationParams) (db.DocumentCycleObligation, error) {
	f.updateObligationArg = arg
	return f.updateObligationRow, f.updateObligationErr
}

func (f *fakeDocumentCycleStore) GetDocumentCycleObligationStatus(ctx context.Context, id pgtype.UUID) (string, error) {
	f.statusLookupID = id
	return f.statusBefore, f.statusErr
}

func (f *fakeDocumentCycleStore) GetDocumentCycleObligationCompletionReadiness(ctx context.Context, id pgtype.UUID) (db.GetDocumentCycleObligationCompletionReadinessRow, error) {
	f.readinessID = id
	return f.readinessRow, f.readinessErr
}

func (f *fakeDocumentCycleStore) UpdateDocumentCycleObligationStatus(ctx context.Context, arg db.UpdateDocumentCycleObligationStatusParams) (db.DocumentCycleObligation, error) {
	f.updateStatusArg = arg
	if f.updateStatusErr != nil {
		return db.DocumentCycleObligation{}, f.updateStatusErr
	}
	return f.updateStatusRow, nil
}

func (f *fakeDocumentCycleStore) DeleteDocumentCycleObligation(ctx context.Context, id pgtype.UUID) error {
	f.deleteObligationID = id
	return f.deleteObligationErr
}

func (f *fakeDocumentCycleStore) CreateDocumentCycleEvent(ctx context.Context, arg db.CreateDocumentCycleEventParams) (db.DocumentCycleEvent, error) {
	f.eventArg = arg
	if f.eventErr != nil {
		return db.DocumentCycleEvent{}, f.eventErr
	}
	return db.DocumentCycleEvent{}, nil
}

func (f *fakeDocumentCycleStore) ListDocumentCycleEventsByObligation(ctx context.Context, arg db.ListDocumentCycleEventsByObligationParams) ([]db.ListDocumentCycleEventsByObligationRow, error) {
	f.listEventsArg = arg
	return f.listEventsRows, f.listEventsErr
}

func documentCycleTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func documentCycleTestDate(year int, month int, day int) pgtype.Date {
	return pgtype.Date{Time: time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func TestDocumentCycleStatsCatalogsAndGenerateYear(t *testing.T) {
	actorID := documentCycleTestUUID(31)
	catalogID := documentCycleTestUUID(32)
	store := &fakeDocumentCycleStore{
		statsRow: db.GetDocumentCycleStatsRow{ActiveCatalogs: 3, TotalObligations: 9},
		listCatalogsRows: []db.ListDocumentCycleCatalogsRow{
			{ID: catalogID, Code: "RKT", Frequency: "annual"},
		},
		createCatalogRow: db.DocumentCycleCatalog{ID: catalogID, Code: "RKT"},
		updateCatalogRow: db.DocumentCycleCatalog{ID: catalogID, Code: "BAP"},
		generateRow:      db.GenerateDocumentCycleYearObligationsRow{Generated: 5, Inserted: 2},
	}
	svc := &DocumentCycle{q: store}

	stats, err := svc.Stats(context.Background(), 0)
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.TotalObligations != 9 || store.statsPeriod == 0 {
		t.Fatalf("Stats() = %+v period=%d, want normalized current year and stats row", stats, store.statsPeriod)
	}

	catalogs, err := svc.ListCatalogs(context.Background(), " rkt ", " Annual ", true)
	if err != nil {
		t.Fatalf("ListCatalogs() error = %v", err)
	}
	if len(catalogs) != 1 || store.listCatalogsArg.Search != "rkt" || store.listCatalogsArg.Frequency != "annual" || !store.listCatalogsArg.ActiveOnly {
		t.Fatalf("ListCatalogs() rows/arg = %+v/%+v, want trimmed filters", catalogs, store.listCatalogsArg)
	}

	created, err := svc.CreateCatalog(context.Background(), db.CreateDocumentCycleCatalogParams{
		Code:                    " rkt ",
		Title:                   " RKT Tahunan ",
		Frequency:               " annual ",
		DomainArea:              " governance ",
		ExternalSystem:          " edm ",
		SnpStandard:             " SKL ",
		RegulationRef:           " regulasi ",
		DeadlineDaysAfterPeriod: 7,
		ReminderDaysBeforeDue:   3,
		Description:             " deskripsi ",
		IsActive:                true,
	})
	if err != nil {
		t.Fatalf("CreateCatalog() error = %v", err)
	}
	if created.ID != catalogID || store.createCatalogArg.Code != "RKT" || store.createCatalogArg.Title != "RKT Tahunan" || store.createCatalogArg.ExternalSystem != "edm" || store.createCatalogArg.SnpStandard != "skl" {
		t.Fatalf("CreateCatalog() row/arg = %+v/%+v, want normalized catalog", created, store.createCatalogArg)
	}

	updated, err := svc.UpdateCatalog(context.Background(), db.UpdateDocumentCycleCatalogParams{
		ID:                      catalogID,
		Code:                    " bap ",
		Title:                   " Berita Acara ",
		Frequency:               " monthly ",
		DomainArea:              " kurikulum ",
		ExternalSystem:          "",
		SnpStandard:             " isi ",
		DeadlineDaysAfterPeriod: 5,
		ReminderDaysBeforeDue:   2,
	})
	if err != nil {
		t.Fatalf("UpdateCatalog() error = %v", err)
	}
	if updated.Code != "BAP" || store.updateCatalogArg.Code != "BAP" || store.updateCatalogArg.DomainArea != "kurikulum" {
		t.Fatalf("UpdateCatalog() row/arg = %+v/%+v, want normalized update", updated, store.updateCatalogArg)
	}

	if err := svc.DeleteCatalog(context.Background(), catalogID); err != nil {
		t.Fatalf("DeleteCatalog() error = %v", err)
	}
	if store.deleteCatalogID != catalogID {
		t.Fatalf("DeleteCatalog() id = %v, want %v", store.deleteCatalogID, catalogID)
	}

	result, err := svc.GenerateYear(context.Background(), actorID, 2026)
	if err != nil {
		t.Fatalf("GenerateYear() error = %v", err)
	}
	if result.PeriodYear != 2026 || result.Generated != 5 || result.Inserted != 2 || store.generateArg.CreatedByUserID != actorID {
		t.Fatalf("GenerateYear() result/arg = %+v/%+v, want generated summary", result, store.generateArg)
	}
	if _, err := svc.GenerateYear(context.Background(), actorID, 1999); err == nil || err.Error() != "tahun dokumen tidak valid" {
		t.Fatalf("GenerateYear(invalid) error = %v, want invalid year", err)
	}
}

func TestDocumentCycleCatalogValidationAndNormalization(t *testing.T) {
	valid := db.CreateDocumentCycleCatalogParams{
		Code:                    "RKT",
		Title:                   "RKT",
		Frequency:               "annual",
		DomainArea:              "governance",
		ExternalSystem:          "edm",
		SnpStandard:             "skl",
		DeadlineDaysAfterPeriod: 7,
		ReminderDaysBeforeDue:   3,
	}
	tests := []struct {
		name    string
		mutate  func(*db.CreateDocumentCycleCatalogParams)
		wantErr string
	}{
		{name: "empty code", mutate: func(p *db.CreateDocumentCycleCatalogParams) { p.Code = "" }, wantErr: "kode dokumen wajib diisi"},
		{name: "empty title", mutate: func(p *db.CreateDocumentCycleCatalogParams) { p.Title = "" }, wantErr: "nama dokumen wajib diisi"},
		{name: "invalid frequency", mutate: func(p *db.CreateDocumentCycleCatalogParams) { p.Frequency = "yearly" }, wantErr: "frekuensi dokumen tidak valid"},
		{name: "invalid domain", mutate: func(p *db.CreateDocumentCycleCatalogParams) { p.DomainArea = "lain" }, wantErr: "bidang dokumen tidak valid"},
		{name: "invalid external", mutate: func(p *db.CreateDocumentCycleCatalogParams) { p.ExternalSystem = "google" }, wantErr: "sistem eksternal dokumen tidak valid"},
		{name: "invalid snp", mutate: func(p *db.CreateDocumentCycleCatalogParams) { p.SnpStandard = "lain" }, wantErr: "standar SNP tidak valid"},
		{name: "invalid deadline", mutate: func(p *db.CreateDocumentCycleCatalogParams) { p.DeadlineDaysAfterPeriod = 366 }, wantErr: "batas jatuh tempo harus 0 sampai 365 hari"},
		{name: "invalid reminder", mutate: func(p *db.CreateDocumentCycleCatalogParams) { p.ReminderDaysBeforeDue = 61 }, wantErr: "batas pengingat harus 0 sampai 60 hari"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := valid
			tt.mutate(&arg)
			err := validateDocumentCycleCatalog(arg.Code, arg.Title, arg.Frequency, arg.DomainArea, arg.ExternalSystem, arg.SnpStandard, arg.DeadlineDaysAfterPeriod, arg.ReminderDaysBeforeDue)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("validateDocumentCycleCatalog() error = %v, want %q", err, tt.wantErr)
			}
		})
	}

	if got := normalizeDocumentCycleFrequency(""); got != "monthly" {
		t.Fatalf("normalizeDocumentCycleFrequency(empty) = %q, want monthly", got)
	}
	if got := normalizeDocumentCycleDomainArea(""); got != "governance" {
		t.Fatalf("normalizeDocumentCycleDomainArea(empty) = %q, want governance", got)
	}
	if got := normalizeDocumentCycleStatus(""); got != "not_started" {
		t.Fatalf("normalizeDocumentCycleStatus(empty) = %q, want not_started", got)
	}
	if got := documentCycleStatusLabel("unknown"); got != "unknown" {
		t.Fatalf("documentCycleStatusLabel(unknown) = %q, want unknown", got)
	}
}

func TestDocumentCycleListAndUpdateObligation(t *testing.T) {
	obligationID := documentCycleTestUUID(33)
	actorID := documentCycleTestUUID(34)
	store := &fakeDocumentCycleStore{
		listObligationsRows: []db.ListDocumentCycleObligationsRow{{ID: obligationID, Status: "draft"}},
		updateObligationRow: db.DocumentCycleObligation{ID: obligationID, Status: "draft"},
	}
	svc := &DocumentCycle{q: store}

	rows, err := svc.ListObligations(context.Background(), db.ListDocumentCycleObligationsParams{
		Search:         " rkt ",
		Status:         " Draft ",
		Frequency:      " Annual ",
		DomainArea:     " Governance ",
		ExternalSystem: " EDM ",
	})
	if err != nil {
		t.Fatalf("ListObligations() error = %v", err)
	}
	if len(rows) != 1 || store.listObligationsArg.Search != "rkt" || store.listObligationsArg.Status != "draft" || store.listObligationsArg.PeriodYear == 0 {
		t.Fatalf("ListObligations() rows/arg = %+v/%+v, want normalized filters", rows, store.listObligationsArg)
	}

	due := documentCycleTestDate(2026, 5, 10)
	reminder := documentCycleTestDate(2026, 5, 7)
	row, err := svc.UpdateObligation(context.Background(), actorID, db.UpdateDocumentCycleObligationParams{
		ID:                obligationID,
		DueDate:           due,
		ReminderDate:      reminder,
		DomainArea:        " governance ",
		ExternalSystem:    " edm ",
		Notes:             " catatan ",
		VerificationNotes: " cek ",
	})
	if err != nil {
		t.Fatalf("UpdateObligation() error = %v", err)
	}
	if row.ID != obligationID || store.updateObligationArg.DomainArea != "governance" || store.updateObligationArg.ExternalSystem != "edm" || store.updateObligationArg.Notes != "catatan" || store.updateObligationArg.VerificationNotes != "cek" {
		t.Fatalf("UpdateObligation() row/arg = %+v/%+v, want normalized update", row, store.updateObligationArg)
	}
	if store.eventArg.EventType != "updated" || store.eventArg.ActorUserID != actorID {
		t.Fatalf("UpdateObligation() event = %+v, want updated event by actor", store.eventArg)
	}

	if _, err := svc.UpdateObligation(context.Background(), actorID, db.UpdateDocumentCycleObligationParams{DueDate: due, ReminderDate: reminder, DomainArea: "bad", ExternalSystem: "edm"}); err == nil || err.Error() != "bidang dokumen tidak valid" {
		t.Fatalf("UpdateObligation(invalid domain) error = %v, want invalid domain", err)
	}
	if _, err := svc.UpdateObligation(context.Background(), actorID, db.UpdateDocumentCycleObligationParams{DueDate: due, ReminderDate: reminder, ExternalSystem: "bad"}); err == nil || err.Error() != "sistem eksternal dokumen tidak valid" {
		t.Fatalf("UpdateObligation(invalid external) error = %v, want invalid external", err)
	}
	if _, err := svc.UpdateObligation(context.Background(), actorID, db.UpdateDocumentCycleObligationParams{DueDate: due}); err == nil || err.Error() != "tanggal pengingat dan jatuh tempo wajib diisi" {
		t.Fatalf("UpdateObligation(missing reminder) error = %v, want date required", err)
	}
	if _, err := svc.UpdateObligation(context.Background(), actorID, db.UpdateDocumentCycleObligationParams{DueDate: reminder, ReminderDate: due}); err == nil || err.Error() != "tanggal pengingat tidak boleh setelah jatuh tempo" {
		t.Fatalf("UpdateObligation(reminder after due) error = %v, want date order error", err)
	}

	if err := svc.DeleteObligation(context.Background(), obligationID); err != nil {
		t.Fatalf("DeleteObligation() error = %v", err)
	}
	if store.deleteObligationID != obligationID {
		t.Fatalf("DeleteObligation() id = %v, want %v", store.deleteObligationID, obligationID)
	}
}

func TestDocumentCycleUpdateStatusCreatesTransitionEvent(t *testing.T) {
	obligationID := documentCycleTestUUID(1)
	actorID := documentCycleTestUUID(2)
	store := &fakeDocumentCycleStore{
		statusBefore: "waiting_verification",
		readinessRow: db.GetDocumentCycleObligationCompletionReadinessRow{
			ID:                    obligationID,
			ResponsibleEmployeeID: documentCycleTestUUID(7),
			VerifierEmployeeID:    documentCycleTestUUID(8),
			ArchiveDocumentID:     documentCycleTestUUID(9),
		},
		updateStatusRow: db.DocumentCycleObligation{
			ID:     obligationID,
			Status: "completed",
		},
	}
	svc := &DocumentCycle{q: store}

	row, err := svc.UpdateObligationStatus(context.Background(), actorID, obligationID, "completed", "  Dokumen sudah diverifikasi.  ")
	if err != nil {
		t.Fatalf("UpdateObligationStatus() error = %v", err)
	}
	if row.Status != "completed" {
		t.Fatalf("UpdateObligationStatus() status = %q, want completed", row.Status)
	}
	if store.statusLookupID != obligationID {
		t.Fatalf("GetDocumentCycleObligationStatus() id = %v, want %v", store.statusLookupID, obligationID)
	}
	if store.readinessID != obligationID {
		t.Fatalf("GetDocumentCycleObligationCompletionReadiness() id = %v, want %v", store.readinessID, obligationID)
	}
	if store.updateStatusArg.Status != "completed" {
		t.Fatalf("UpdateDocumentCycleObligationStatus() status = %q, want completed", store.updateStatusArg.Status)
	}
	if store.updateStatusArg.Notes != "Dokumen sudah diverifikasi." {
		t.Fatalf("UpdateDocumentCycleObligationStatus() notes = %q", store.updateStatusArg.Notes)
	}
	if store.eventArg.EventType != "status_changed" {
		t.Fatalf("CreateDocumentCycleEvent() event_type = %q, want status_changed", store.eventArg.EventType)
	}
	if store.eventArg.FromStatus != "waiting_verification" || store.eventArg.ToStatus != "completed" {
		t.Fatalf("CreateDocumentCycleEvent() transition = %q -> %q, want waiting_verification -> completed", store.eventArg.FromStatus, store.eventArg.ToStatus)
	}
	if store.eventArg.ActorUserID != actorID {
		t.Fatalf("CreateDocumentCycleEvent() actor = %v, want %v", store.eventArg.ActorUserID, actorID)
	}
}

func TestDocumentCycleUpdateStatusRejectsIncompleteCompletion(t *testing.T) {
	obligationID := documentCycleTestUUID(10)
	store := &fakeDocumentCycleStore{statusBefore: "waiting_verification"}
	svc := &DocumentCycle{q: store}

	_, err := svc.UpdateObligationStatus(context.Background(), documentCycleTestUUID(11), obligationID, "completed", "")
	if err == nil {
		t.Fatalf("UpdateObligationStatus() error = nil, want completion readiness error")
	}
	want := "dokumen belum siap diselesaikan: lengkapi PIC penyusun, verifikator, arsip digital"
	if err.Error() != want {
		t.Fatalf("UpdateObligationStatus() error = %q, want %q", err.Error(), want)
	}
	if store.updateStatusArg.ID.Valid {
		t.Fatalf("UpdateDocumentCycleObligationStatus() was called for incomplete completion")
	}
	if store.eventArg.ObligationID.Valid {
		t.Fatalf("CreateDocumentCycleEvent() was called for incomplete completion")
	}
}

func TestDocumentCycleUpdateStatusDraftSkipsCompletionReadiness(t *testing.T) {
	obligationID := documentCycleTestUUID(12)
	store := &fakeDocumentCycleStore{
		statusBefore: "not_started",
		updateStatusRow: db.DocumentCycleObligation{
			ID:     obligationID,
			Status: "draft",
		},
	}
	svc := &DocumentCycle{q: store}

	_, err := svc.UpdateObligationStatus(context.Background(), documentCycleTestUUID(13), obligationID, "draft", "")
	if err != nil {
		t.Fatalf("UpdateObligationStatus() error = %v", err)
	}
	if store.readinessID.Valid {
		t.Fatalf("GetDocumentCycleObligationCompletionReadiness() was called for draft status")
	}
	if store.updateStatusArg.Status != "draft" {
		t.Fatalf("UpdateDocumentCycleObligationStatus() status = %q, want draft", store.updateStatusArg.Status)
	}
}

func TestDocumentCycleUpdateStatusRequiresArchiveForCompletion(t *testing.T) {
	obligationID := documentCycleTestUUID(14)
	store := &fakeDocumentCycleStore{
		statusBefore: "waiting_verification",
		readinessRow: db.GetDocumentCycleObligationCompletionReadinessRow{
			ID:                    obligationID,
			ResponsibleEmployeeID: documentCycleTestUUID(15),
			VerifierEmployeeID:    documentCycleTestUUID(16),
			EvidenceItemID:        documentCycleTestUUID(17),
		},
	}
	svc := &DocumentCycle{q: store}

	_, err := svc.UpdateObligationStatus(context.Background(), documentCycleTestUUID(18), obligationID, "completed", "")
	if err == nil || err.Error() != "dokumen belum siap diselesaikan: lengkapi arsip digital" {
		t.Fatalf("UpdateObligationStatus() error = %v, want archive readiness error", err)
	}
	if store.updateStatusArg.ID.Valid {
		t.Fatalf("UpdateDocumentCycleObligationStatus() was called without archive readiness")
	}
}

func TestDocumentCycleUpdateStatusRejectsInvalidTransition(t *testing.T) {
	obligationID := documentCycleTestUUID(19)
	store := &fakeDocumentCycleStore{statusBefore: "draft"}
	svc := &DocumentCycle{q: store}

	_, err := svc.UpdateObligationStatus(context.Background(), documentCycleTestUUID(20), obligationID, "completed", "")
	if err == nil {
		t.Fatalf("UpdateObligationStatus() error = nil, want invalid transition error")
	}
	want := "alur status dokumen tidak valid: Sedang Dibuat -> Selesai"
	if err.Error() != want {
		t.Fatalf("UpdateObligationStatus() error = %q, want %q", err.Error(), want)
	}
	if store.readinessID.Valid {
		t.Fatalf("GetDocumentCycleObligationCompletionReadiness() was called for invalid transition")
	}
	if store.updateStatusArg.ID.Valid {
		t.Fatalf("UpdateDocumentCycleObligationStatus() was called for invalid transition")
	}
	if store.eventArg.ObligationID.Valid {
		t.Fatalf("CreateDocumentCycleEvent() was called for invalid transition")
	}
}

func TestDocumentCycleUpdateStatusStopsWhenStatusLookupFails(t *testing.T) {
	obligationID := documentCycleTestUUID(3)
	store := &fakeDocumentCycleStore{statusErr: errors.New("lookup failed")}
	svc := &DocumentCycle{q: store}

	_, err := svc.UpdateObligationStatus(context.Background(), documentCycleTestUUID(4), obligationID, "completed", "")
	if err == nil || err.Error() != "lookup failed" {
		t.Fatalf("UpdateObligationStatus() error = %v, want lookup failed", err)
	}
	if store.updateStatusArg.ID.Valid {
		t.Fatalf("UpdateDocumentCycleObligationStatus() was called after lookup failure")
	}
	if store.eventArg.ObligationID.Valid {
		t.Fatalf("CreateDocumentCycleEvent() was called after lookup failure")
	}
}

func TestDocumentCycleListEventsDelegatesToStore(t *testing.T) {
	obligationID := documentCycleTestUUID(5)
	store := &fakeDocumentCycleStore{
		listEventsRows: []db.ListDocumentCycleEventsByObligationRow{
			{ID: documentCycleTestUUID(6), ObligationID: obligationID, EventType: "generated"},
		},
	}
	svc := &DocumentCycle{q: store}

	rows, err := svc.ListEvents(context.Background(), obligationID, "", "")
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if store.listEventsArg.ObligationID != obligationID {
		t.Fatalf("ListDocumentCycleEventsByObligation() id = %v, want %v", store.listEventsArg.ObligationID, obligationID)
	}
	if store.listEventsArg.EventType != "" || store.listEventsArg.Actor != "" {
		t.Fatalf("ListDocumentCycleEventsByObligation() filters = %+v, want empty event/actor filters", store.listEventsArg)
	}
	if len(rows) != 1 || rows[0].EventType != "generated" {
		t.Fatalf("ListEvents() rows = %+v, want one generated event", rows)
	}
}

func TestDocumentCycleListEventsPassesFilters(t *testing.T) {
	obligationID := documentCycleTestUUID(21)
	store := &fakeDocumentCycleStore{}
	svc := &DocumentCycle{q: store}

	_, err := svc.ListEvents(context.Background(), obligationID, " status_changed ", " kepala ")
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if store.listEventsArg.EventType != "status_changed" || store.listEventsArg.Actor != "kepala" {
		t.Fatalf("ListDocumentCycleEventsByObligation() filters = %+v, want status_changed/kepala", store.listEventsArg)
	}
}

func TestDocumentCycleListEventsRejectsInvalidEventType(t *testing.T) {
	store := &fakeDocumentCycleStore{}
	svc := &DocumentCycle{q: store}

	_, err := svc.ListEvents(context.Background(), documentCycleTestUUID(22), "invalid", "")
	if err == nil || err.Error() != "jenis audit dokumen tidak valid" {
		t.Fatalf("ListEvents() error = %v, want invalid event type", err)
	}
	if store.listEventsArg.ObligationID.Valid {
		t.Fatalf("ListDocumentCycleEventsByObligation() was called for invalid event type")
	}
}

func TestDocumentCycleListVerificationQueueFiltersWaitingByVerifier(t *testing.T) {
	verifierID := documentCycleTestUUID(23)
	store := &fakeDocumentCycleStore{
		listObligationsRows: []db.ListDocumentCycleObligationsRow{
			{ID: documentCycleTestUUID(24), Status: "waiting_verification"},
		},
	}
	svc := &DocumentCycle{q: store}

	rows, err := svc.ListVerificationQueue(context.Background(), verifierID, 2026)
	if err != nil {
		t.Fatalf("ListVerificationQueue() error = %v", err)
	}
	if store.listObligationsArg.Status != "waiting_verification" {
		t.Fatalf("ListDocumentCycleObligations() status = %q, want waiting_verification", store.listObligationsArg.Status)
	}
	if store.listObligationsArg.VerifierEmployeeID != verifierID {
		t.Fatalf("ListDocumentCycleObligations() verifier = %v, want %v", store.listObligationsArg.VerifierEmployeeID, verifierID)
	}
	if len(rows) != 1 {
		t.Fatalf("ListVerificationQueue() rows = %d, want 1", len(rows))
	}
}
