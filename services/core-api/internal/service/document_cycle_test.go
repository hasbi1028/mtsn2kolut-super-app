package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeDocumentCycleStore struct {
	statusLookupID pgtype.UUID
	statusBefore   string
	statusErr      error

	readinessID  pgtype.UUID
	readinessRow db.GetDocumentCycleObligationCompletionReadinessRow
	readinessErr error

	updateStatusArg db.UpdateDocumentCycleObligationStatusParams
	updateStatusRow db.DocumentCycleObligation
	updateStatusErr error

	eventArg db.CreateDocumentCycleEventParams
	eventErr error

	listEventsID   pgtype.UUID
	listEventsRows []db.ListDocumentCycleEventsByObligationRow
	listEventsErr  error
}

func (f *fakeDocumentCycleStore) GetDocumentCycleStats(ctx context.Context, periodYear int32) (db.GetDocumentCycleStatsRow, error) {
	return db.GetDocumentCycleStatsRow{}, nil
}

func (f *fakeDocumentCycleStore) ListDocumentCycleCatalogs(ctx context.Context, arg db.ListDocumentCycleCatalogsParams) ([]db.ListDocumentCycleCatalogsRow, error) {
	return nil, nil
}

func (f *fakeDocumentCycleStore) CreateDocumentCycleCatalog(ctx context.Context, arg db.CreateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error) {
	return db.DocumentCycleCatalog{}, nil
}

func (f *fakeDocumentCycleStore) UpdateDocumentCycleCatalog(ctx context.Context, arg db.UpdateDocumentCycleCatalogParams) (db.DocumentCycleCatalog, error) {
	return db.DocumentCycleCatalog{}, nil
}

func (f *fakeDocumentCycleStore) DeleteDocumentCycleCatalog(ctx context.Context, id pgtype.UUID) error {
	return nil
}

func (f *fakeDocumentCycleStore) ListDocumentCycleObligations(ctx context.Context, arg db.ListDocumentCycleObligationsParams) ([]db.ListDocumentCycleObligationsRow, error) {
	return nil, nil
}

func (f *fakeDocumentCycleStore) GenerateDocumentCycleYearObligations(ctx context.Context, arg db.GenerateDocumentCycleYearObligationsParams) (db.GenerateDocumentCycleYearObligationsRow, error) {
	return db.GenerateDocumentCycleYearObligationsRow{}, nil
}

func (f *fakeDocumentCycleStore) UpdateDocumentCycleObligation(ctx context.Context, arg db.UpdateDocumentCycleObligationParams) (db.DocumentCycleObligation, error) {
	return db.DocumentCycleObligation{}, nil
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
	return nil
}

func (f *fakeDocumentCycleStore) CreateDocumentCycleEvent(ctx context.Context, arg db.CreateDocumentCycleEventParams) (db.DocumentCycleEvent, error) {
	f.eventArg = arg
	if f.eventErr != nil {
		return db.DocumentCycleEvent{}, f.eventErr
	}
	return db.DocumentCycleEvent{}, nil
}

func (f *fakeDocumentCycleStore) ListDocumentCycleEventsByObligation(ctx context.Context, obligationID pgtype.UUID) ([]db.ListDocumentCycleEventsByObligationRow, error) {
	f.listEventsID = obligationID
	return f.listEventsRows, f.listEventsErr
}

func documentCycleTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
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

	rows, err := svc.ListEvents(context.Background(), obligationID)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if store.listEventsID != obligationID {
		t.Fatalf("ListDocumentCycleEventsByObligation() id = %v, want %v", store.listEventsID, obligationID)
	}
	if len(rows) != 1 || rows[0].EventType != "generated" {
		t.Fatalf("ListEvents() rows = %+v, want one generated event", rows)
	}
}
