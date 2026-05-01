package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeInventoryStore struct {
	listArg     db.ListInventoryItemsParams
	listRows    []db.InventoryItem
	listErr     error
	itemsByID   map[pgtype.UUID]db.InventoryItem
	getItem     db.InventoryItem
	getErr      error
	getIDs      []pgtype.UUID
	createArg   db.CreateInventoryItemParams
	createErr   error
	updateArgs  []db.UpdateInventoryItemParams
	updateErr   error
	deleteID    pgtype.UUID
	deleteErr   error
	eventArgs   []db.CreateInventoryItemEventParams
	eventErr    error
	eventRows   []db.ListInventoryItemEventsByItemRow
	eventItemID pgtype.UUID
	stats       db.GetInventoryStatsRow
	statsErr    error
}

func (f *fakeInventoryStore) ListInventoryItems(ctx context.Context, arg db.ListInventoryItemsParams) ([]db.InventoryItem, error) {
	f.listArg = arg
	return f.listRows, f.listErr
}

func (f *fakeInventoryStore) GetInventoryItem(ctx context.Context, id pgtype.UUID) (db.InventoryItem, error) {
	f.getIDs = append(f.getIDs, id)
	if f.getErr != nil {
		return db.InventoryItem{}, f.getErr
	}
	if f.itemsByID != nil {
		if item, ok := f.itemsByID[id]; ok {
			return item, nil
		}
	}
	return f.getItem, nil
}

func (f *fakeInventoryStore) CreateInventoryItem(ctx context.Context, arg db.CreateInventoryItemParams) (db.InventoryItem, error) {
	f.createArg = arg
	if f.createErr != nil {
		return db.InventoryItem{}, f.createErr
	}
	return db.InventoryItem{
		ID:          inventoryTestUUID(1),
		Kode:        arg.Kode,
		Nama:        arg.Nama,
		Kategori:    arg.Kategori,
		Lokasi:      arg.Lokasi,
		Kondisi:     arg.Kondisi,
		Satuan:      arg.Satuan,
		JumlahTotal: arg.JumlahTotal,
		JumlahBaik:  arg.JumlahBaik,
		MinStock:    arg.MinStock,
		Catatan:     arg.Catatan,
	}, nil
}

func (f *fakeInventoryStore) UpdateInventoryItem(ctx context.Context, arg db.UpdateInventoryItemParams) (db.InventoryItem, error) {
	f.updateArgs = append(f.updateArgs, arg)
	if f.updateErr != nil {
		return db.InventoryItem{}, f.updateErr
	}
	return db.InventoryItem{
		ID:          arg.ID,
		Kode:        arg.Kode,
		Nama:        arg.Nama,
		Kategori:    arg.Kategori,
		Lokasi:      arg.Lokasi,
		Kondisi:     arg.Kondisi,
		Satuan:      arg.Satuan,
		JumlahTotal: arg.JumlahTotal,
		JumlahBaik:  arg.JumlahBaik,
		MinStock:    arg.MinStock,
		Catatan:     arg.Catatan,
	}, nil
}

func (f *fakeInventoryStore) DeleteInventoryItem(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeInventoryStore) CreateInventoryItemEvent(ctx context.Context, arg db.CreateInventoryItemEventParams) (db.InventoryItemEvent, error) {
	f.eventArgs = append(f.eventArgs, arg)
	if f.eventErr != nil {
		return db.InventoryItemEvent{}, f.eventErr
	}
	return db.InventoryItemEvent{ItemID: arg.ItemID, ActorUserID: arg.ActorUserID, Action: arg.Action, Summary: arg.Summary}, nil
}

func (f *fakeInventoryStore) ListInventoryItemEventsByItem(ctx context.Context, itemID pgtype.UUID) ([]db.ListInventoryItemEventsByItemRow, error) {
	f.eventItemID = itemID
	return f.eventRows, nil
}

func (f *fakeInventoryStore) GetInventoryStats(ctx context.Context) (db.GetInventoryStatsRow, error) {
	return f.stats, f.statsErr
}

func inventoryTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func TestInventoryListItemsTrimsFilters(t *testing.T) {
	store := &fakeInventoryStore{}
	svc := &Inventory{q: store}

	if _, err := svc.ListItems(context.Background(), "  laptop  ", "  elektronik  ", "  baik  "); err != nil {
		t.Fatalf("ListItems() error = %v", err)
	}
	if store.listArg.Search != "laptop" || store.listArg.Kategori != "elektronik" || store.listArg.Kondisi != "baik" {
		t.Fatalf("ListItems() arg = %+v, want trimmed filters", store.listArg)
	}
}

func TestInventoryGetItemDelegatesToStore(t *testing.T) {
	itemID := inventoryTestUUID(14)
	expectedErr := errors.New("item lookup failed")
	store := &fakeInventoryStore{
		getItem: db.InventoryItem{ID: itemID, Kode: "INV-014", Nama: "Proyektor"},
	}
	svc := &Inventory{q: store}

	item, err := svc.GetItem(context.Background(), itemID)
	if err != nil {
		t.Fatalf("GetItem() error = %v", err)
	}
	if item.ID != itemID || item.Nama != "Proyektor" || len(store.getIDs) != 1 || store.getIDs[0] != itemID {
		t.Fatalf("GetItem() = %+v ids=%+v, want delegated item lookup", item, store.getIDs)
	}

	store.getErr = expectedErr
	if _, err := svc.GetItem(context.Background(), itemID); !errors.Is(err, expectedErr) {
		t.Fatalf("GetItem(error) = %v, want %v", err, expectedErr)
	}
}

func TestValidateInventoryPayload(t *testing.T) {
	tests := []struct {
		name    string
		kode    string
		nama    string
		kondisi string
		total   int32
		baik    int32
		min     int32
		wantErr string
	}{
		{name: "valid with default condition", kode: "INV-1", nama: "Laptop", total: 2, baik: 2},
		{name: "empty kode", nama: "Laptop", total: 2, baik: 2, wantErr: "kode inventaris wajib diisi"},
		{name: "empty nama", kode: "INV-1", total: 2, baik: 2, wantErr: "nama barang wajib diisi"},
		{name: "total below minimum", kode: "INV-1", nama: "Laptop", total: 0, wantErr: "jumlah total minimal 1"},
		{name: "baik below zero", kode: "INV-1", nama: "Laptop", total: 2, baik: -1, wantErr: "jumlah kondisi baik harus antara 0 dan jumlah total"},
		{name: "baik above total", kode: "INV-1", nama: "Laptop", total: 2, baik: 3, wantErr: "jumlah kondisi baik harus antara 0 dan jumlah total"},
		{name: "min stock below zero", kode: "INV-1", nama: "Laptop", total: 2, baik: 2, min: -1, wantErr: "batas restok tidak boleh negatif"},
		{name: "invalid condition", kode: "INV-1", nama: "Laptop", kondisi: "hilang", total: 2, baik: 2, wantErr: "kondisi inventaris tidak valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInventoryPayload(tt.kode, tt.nama, tt.kondisi, tt.total, tt.baik, tt.min)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateInventoryPayload() error = %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("validateInventoryPayload() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestInventoryCreateItemNormalizesAndAudits(t *testing.T) {
	actorID := inventoryTestUUID(9)
	store := &fakeInventoryStore{}
	svc := &Inventory{q: store}

	item, err := svc.CreateItem(context.Background(), actorID, db.CreateInventoryItemParams{
		Kode:        "  INV-001  ",
		Nama:        "  Laptop Guru  ",
		Lokasi:      "   ",
		JumlahTotal: 3,
		JumlahBaik:  2,
		Catatan:     "  Siap dipakai  ",
	})
	if err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}
	if item.Kode != "INV-001" || item.Nama != "Laptop Guru" {
		t.Fatalf("CreateItem() item = %+v, want trimmed kode/nama", item)
	}
	if store.createArg.Kategori != "umum" || store.createArg.Kondisi != "baik" || store.createArg.Satuan != "unit" {
		t.Fatalf("CreateInventoryItem() arg = %+v, want default kategori/kondisi/satuan", store.createArg)
	}
	if store.createArg.Lokasi != "" || store.createArg.Catatan != "Siap dipakai" {
		t.Fatalf("CreateInventoryItem() arg = %+v, want trimmed lokasi/catatan", store.createArg)
	}
	if len(store.eventArgs) != 1 {
		t.Fatalf("CreateInventoryItemEvent() calls = %d, want 1", len(store.eventArgs))
	}
	event := store.eventArgs[0]
	if event.ActorUserID != actorID || event.Action != "create" {
		t.Fatalf("CreateInventoryItemEvent() arg = %+v, want actor/action", event)
	}
	if !strings.Contains(event.Summary, "Belum diatur") || !strings.Contains(event.Summary, "stok layak 2/3") {
		t.Fatalf("CreateInventoryItemEvent() summary = %q, want location fallback and stock summary", event.Summary)
	}
}

func TestInventoryCreateItemStopsOnValidationError(t *testing.T) {
	store := &fakeInventoryStore{}
	svc := &Inventory{q: store}

	_, err := svc.CreateItem(context.Background(), inventoryTestUUID(1), db.CreateInventoryItemParams{
		Kode:        "INV-001",
		Nama:        "Laptop",
		Kondisi:     "hilang",
		JumlahTotal: 1,
		JumlahBaik:  1,
	})
	if err == nil || err.Error() != "kondisi inventaris tidak valid" {
		t.Fatalf("CreateItem() error = %v, want invalid condition", err)
	}
	if store.createArg.Kode != "" {
		t.Fatalf("CreateInventoryItem() was called after validation failure")
	}
}

func TestInventoryCreateItemPropagatesStoreError(t *testing.T) {
	expectedErr := errors.New("create failed")
	store := &fakeInventoryStore{createErr: expectedErr}
	svc := &Inventory{q: store}

	_, err := svc.CreateItem(context.Background(), inventoryTestUUID(1), db.CreateInventoryItemParams{
		Kode:        "INV-001",
		Nama:        "Laptop",
		JumlahTotal: 1,
		JumlahBaik:  1,
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("CreateItem(store error) = %v, want %v", err, expectedErr)
	}
	if len(store.eventArgs) != 0 {
		t.Fatalf("CreateInventoryItemEvent() calls = %d, want 0 when create fails", len(store.eventArgs))
	}
}

func TestInventoryUpdateItemSummarizesChanges(t *testing.T) {
	itemID := inventoryTestUUID(2)
	actorID := inventoryTestUUID(3)
	store := &fakeInventoryStore{
		getItem: db.InventoryItem{
			ID:          itemID,
			Kode:        "INV-001",
			Nama:        "Laptop",
			Kategori:    "elektronik",
			Lokasi:      "",
			Kondisi:     "baik",
			Satuan:      "unit",
			JumlahTotal: 5,
			JumlahBaik:  5,
			MinStock:    1,
		},
	}
	svc := &Inventory{q: store}

	_, err := svc.UpdateItem(context.Background(), actorID, db.UpdateInventoryItemParams{
		ID:          itemID,
		Kode:        " INV-001 ",
		Nama:        " Laptop ",
		Kategori:    " elektronik ",
		Lokasi:      " Lab IPA ",
		Kondisi:     " rusak ",
		Satuan:      " unit ",
		JumlahTotal: 5,
		JumlahBaik:  2,
		MinStock:    2,
		Catatan:     " Perlu servis ",
	})
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}
	if len(store.updateArgs) != 1 {
		t.Fatalf("UpdateInventoryItem() calls = %d, want 1", len(store.updateArgs))
	}
	arg := store.updateArgs[0]
	if arg.Kode != "INV-001" || arg.Nama != "Laptop" || arg.Kategori != "elektronik" || arg.Lokasi != "Lab IPA" {
		t.Fatalf("UpdateInventoryItem() arg = %+v, want normalized text", arg)
	}
	if len(store.eventArgs) != 1 {
		t.Fatalf("CreateInventoryItemEvent() calls = %d, want 1", len(store.eventArgs))
	}
	summary := store.eventArgs[0].Summary
	for _, want := range []string{"lokasi 'Belum diatur' -> 'Lab IPA'", "kondisi 'baik' -> 'rusak'", "stok layak 5/5 -> 2/5", "batas restok 1 -> 2", "catatan diperbarui"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary = %q, want %q", summary, want)
		}
	}
}

func TestInventoryUpdateItemValidationAndStoreErrors(t *testing.T) {
	itemID := inventoryTestUUID(15)
	actorID := inventoryTestUUID(16)
	validArg := db.UpdateInventoryItemParams{
		ID:          itemID,
		Kode:        "INV-015",
		Nama:        "Meja",
		Kondisi:     "baik",
		JumlahTotal: 2,
		JumlahBaik:  2,
	}

	store := &fakeInventoryStore{}
	svc := &Inventory{q: store}
	if _, err := svc.UpdateItem(context.Background(), actorID, db.UpdateInventoryItemParams{Kode: "", Nama: "Meja", JumlahTotal: 1, JumlahBaik: 1}); err == nil || err.Error() != "kode inventaris wajib diisi" {
		t.Fatalf("UpdateItem(validation) = %v, want kode validation", err)
	}
	if len(store.getIDs) != 0 {
		t.Fatalf("GetInventoryItem() calls = %d, want 0 after validation failure", len(store.getIDs))
	}

	expectedErr := errors.New("lookup failed")
	store = &fakeInventoryStore{getErr: expectedErr}
	svc = &Inventory{q: store}
	if _, err := svc.UpdateItem(context.Background(), actorID, validArg); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateItem(get error) = %v, want %v", err, expectedErr)
	}

	expectedErr = errors.New("update failed")
	store = &fakeInventoryStore{
		getItem: db.InventoryItem{
			ID:          itemID,
			Kode:        "INV-015",
			Nama:        "Meja",
			Kondisi:     "baik",
			JumlahTotal: 2,
			JumlahBaik:  2,
		},
		updateErr: expectedErr,
	}
	svc = &Inventory{q: store}
	if _, err := svc.UpdateItem(context.Background(), actorID, validArg); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateItem(update error) = %v, want %v", err, expectedErr)
	}
	if len(store.eventArgs) != 0 {
		t.Fatalf("CreateInventoryItemEvent() calls = %d, want 0 when update fails", len(store.eventArgs))
	}
}

func TestInventoryDeleteItemCreatesAuditEventBeforeDelete(t *testing.T) {
	itemID := inventoryTestUUID(4)
	actorID := inventoryTestUUID(5)
	store := &fakeInventoryStore{
		getItem: db.InventoryItem{
			ID:          itemID,
			Lokasi:      "Gudang TU",
			Kondisi:     "perlu-perawatan",
			JumlahTotal: 4,
			JumlahBaik:  2,
		},
	}
	svc := &Inventory{q: store}

	if err := svc.DeleteItem(context.Background(), actorID, itemID); err != nil {
		t.Fatalf("DeleteItem() error = %v", err)
	}
	if store.deleteID != itemID {
		t.Fatalf("DeleteInventoryItem() id = %v, want %v", store.deleteID, itemID)
	}
	if len(store.eventArgs) != 1 {
		t.Fatalf("CreateInventoryItemEvent() calls = %d, want 1", len(store.eventArgs))
	}
	event := store.eventArgs[0]
	if event.Action != "delete" || event.ActorUserID != actorID || event.ItemID != itemID {
		t.Fatalf("CreateInventoryItemEvent() arg = %+v, want delete event", event)
	}
	if !strings.Contains(event.Summary, "Gudang TU") || !strings.Contains(event.Summary, "stok layak 2/4") {
		t.Fatalf("delete summary = %q, want last known location and stock", event.Summary)
	}
}

func TestInventoryDeleteItemPropagatesStoreErrors(t *testing.T) {
	itemID := inventoryTestUUID(17)
	actorID := inventoryTestUUID(18)

	expectedErr := errors.New("lookup failed")
	store := &fakeInventoryStore{getErr: expectedErr}
	svc := &Inventory{q: store}
	if err := svc.DeleteItem(context.Background(), actorID, itemID); !errors.Is(err, expectedErr) {
		t.Fatalf("DeleteItem(get error) = %v, want %v", err, expectedErr)
	}
	if store.deleteID.Valid {
		t.Fatalf("DeleteInventoryItem() id = %v, want zero when lookup fails", store.deleteID)
	}

	expectedErr = errors.New("delete failed")
	store = &fakeInventoryStore{
		getItem:   db.InventoryItem{ID: itemID, Kondisi: "baik", JumlahTotal: 1, JumlahBaik: 1},
		deleteErr: expectedErr,
	}
	svc = &Inventory{q: store}
	if err := svc.DeleteItem(context.Background(), actorID, itemID); !errors.Is(err, expectedErr) {
		t.Fatalf("DeleteItem(delete error) = %v, want %v", err, expectedErr)
	}
	if len(store.eventArgs) != 1 {
		t.Fatalf("CreateInventoryItemEvent() calls = %d, want delete audit before delete failure", len(store.eventArgs))
	}
}

func TestInventoryBatchUpdateItems(t *testing.T) {
	actorID := inventoryTestUUID(6)
	firstID := inventoryTestUUID(7)
	secondID := inventoryTestUUID(8)
	store := &fakeInventoryStore{
		itemsByID: map[pgtype.UUID]db.InventoryItem{
			firstID: {
				ID:          firstID,
				Kode:        "INV-001",
				Nama:        "Laptop 1",
				Kategori:    "elektronik",
				Lokasi:      "Ruang Guru",
				Kondisi:     "baik",
				Satuan:      "unit",
				JumlahTotal: 1,
				JumlahBaik:  1,
			},
			secondID: {
				ID:          secondID,
				Kode:        "INV-002",
				Nama:        "Laptop 2",
				Kategori:    "elektronik",
				Lokasi:      "Ruang Guru",
				Kondisi:     "baik",
				Satuan:      "unit",
				JumlahTotal: 1,
				JumlahBaik:  1,
			},
		},
	}
	svc := &Inventory{q: store}

	if updated, err := svc.BatchUpdateItems(context.Background(), actorID, nil, nil, nil); err == nil || updated != 0 {
		t.Fatalf("BatchUpdateItems(empty) = %d, %v; want validation error", updated, err)
	}
	if updated, err := svc.BatchUpdateItems(context.Background(), actorID, []pgtype.UUID{firstID}, nil, nil); err == nil || updated != 0 {
		t.Fatalf("BatchUpdateItems(no changes) = %d, %v; want validation error", updated, err)
	}

	lokasi := "  Lab Komputer  "
	kondisi := "  perlu-perawatan  "
	updated, err := svc.BatchUpdateItems(context.Background(), actorID, []pgtype.UUID{firstID, secondID}, &lokasi, &kondisi)
	if err != nil {
		t.Fatalf("BatchUpdateItems() error = %v", err)
	}
	if updated != 2 {
		t.Fatalf("BatchUpdateItems() updated = %d, want 2", updated)
	}
	if len(store.updateArgs) != 2 {
		t.Fatalf("UpdateInventoryItem() calls = %d, want 2", len(store.updateArgs))
	}
	for _, arg := range store.updateArgs {
		if arg.Lokasi != "Lab Komputer" || arg.Kondisi != "perlu-perawatan" {
			t.Fatalf("UpdateInventoryItem() arg = %+v, want batch location/condition", arg)
		}
	}
	if len(store.eventArgs) != 2 {
		t.Fatalf("CreateInventoryItemEvent() calls = %d, want 2", len(store.eventArgs))
	}
}

func TestInventoryBatchUpdateStopsOnStoreError(t *testing.T) {
	firstID := inventoryTestUUID(10)
	storeErr := errors.New("lookup failed")
	store := &fakeInventoryStore{
		getErr: storeErr,
	}
	svc := &Inventory{q: store}

	lokasi := "Lab"
	updated, err := svc.BatchUpdateItems(context.Background(), inventoryTestUUID(12), []pgtype.UUID{firstID}, &lokasi, nil)
	if err == nil || !errors.Is(err, storeErr) {
		t.Fatalf("BatchUpdateItems() error = %v, want %v", err, storeErr)
	}
	if updated != 0 {
		t.Fatalf("BatchUpdateItems() updated = %d, want 0 before lookup failure", updated)
	}
}

func TestInventoryStatsAndEventsDelegateToStore(t *testing.T) {
	itemID := inventoryTestUUID(13)
	store := &fakeInventoryStore{
		stats:     db.GetInventoryStatsRow{TotalJenis: 4, PerluRestok: 1},
		eventRows: []db.ListInventoryItemEventsByItemRow{{ItemID: itemID, Action: "create"}},
	}
	svc := &Inventory{q: store}

	stats, err := svc.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.TotalJenis != 4 || stats.PerluRestok != 1 {
		t.Fatalf("Stats() = %+v, want store stats", stats)
	}

	events, err := svc.ListItemEvents(context.Background(), itemID)
	if err != nil {
		t.Fatalf("ListItemEvents() error = %v", err)
	}
	if store.eventItemID != itemID || len(events) != 1 || events[0].Action != "create" {
		t.Fatalf("ListItemEvents() = %+v, store item id = %v; want delegated event rows", events, store.eventItemID)
	}
}
