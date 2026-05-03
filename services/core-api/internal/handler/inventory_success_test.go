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

type fakeInventoryService struct {
	*service.Inventory

	statsRow      db.GetInventoryStatsRow
	statsErr      error
	listSearch    string
	listKategori  string
	listKondisi   string
	listRows      []db.InventoryItem
	listErr       error
	eventsID      pgtype.UUID
	eventsRows    []db.ListInventoryItemEventsByItemRow
	eventsErr     error
	createActorID pgtype.UUID
	createArg     db.CreateInventoryItemParams
	createRow     db.InventoryItem
	createErr     error
	batchActorID  pgtype.UUID
	batchIDs      []pgtype.UUID
	batchLokasi   *string
	batchKondisi  *string
	batchUpdated  int
	batchErr      error
	updateActorID pgtype.UUID
	updateArg     db.UpdateInventoryItemParams
	updateRow     db.InventoryItem
	updateErr     error
	deleteActorID pgtype.UUID
	deleteID      pgtype.UUID
	deleteErr     error
	roomRows      []db.SchoolRoom
	roomSearch    string
	roomType      string
	roomCondition string
	roomEligible  string
	roomID        pgtype.UUID
	roomCreate    db.CreateSchoolRoomParams
	roomUpdate    db.UpdateSchoolRoomParams
	roomRow       db.SchoolRoom
}

func (f *fakeInventoryService) Stats(context.Context) (db.GetInventoryStatsRow, error) {
	return f.statsRow, f.statsErr
}

func (f *fakeInventoryService) ListItems(_ context.Context, search, kategori, kondisi string) ([]db.InventoryItem, error) {
	f.listSearch = search
	f.listKategori = kategori
	f.listKondisi = kondisi
	return f.listRows, f.listErr
}

func (f *fakeInventoryService) ListItemEvents(_ context.Context, itemID pgtype.UUID) ([]db.ListInventoryItemEventsByItemRow, error) {
	f.eventsID = itemID
	return f.eventsRows, f.eventsErr
}

func (f *fakeInventoryService) CreateItem(_ context.Context, actorUserID pgtype.UUID, arg db.CreateInventoryItemParams) (db.InventoryItem, error) {
	f.createActorID = actorUserID
	f.createArg = arg
	return f.createRow, f.createErr
}

func (f *fakeInventoryService) BatchUpdateItems(_ context.Context, actorUserID pgtype.UUID, ids []pgtype.UUID, lokasi *string, kondisi *string) (int, error) {
	f.batchActorID = actorUserID
	f.batchIDs = ids
	f.batchLokasi = lokasi
	f.batchKondisi = kondisi
	return f.batchUpdated, f.batchErr
}

func (f *fakeInventoryService) UpdateItem(_ context.Context, actorUserID pgtype.UUID, arg db.UpdateInventoryItemParams) (db.InventoryItem, error) {
	f.updateActorID = actorUserID
	f.updateArg = arg
	return f.updateRow, f.updateErr
}

func (f *fakeInventoryService) DeleteItem(_ context.Context, actorUserID pgtype.UUID, id pgtype.UUID) error {
	f.deleteActorID = actorUserID
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeInventoryService) ListSchoolRooms(_ context.Context, search, roomType, condition, examEligible string) ([]db.SchoolRoom, error) {
	f.roomSearch = search
	f.roomType = roomType
	f.roomCondition = condition
	f.roomEligible = examEligible
	return f.roomRows, nil
}

func (f *fakeInventoryService) GetSchoolRoom(_ context.Context, id pgtype.UUID) (db.SchoolRoom, error) {
	f.roomID = id
	return f.roomRow, nil
}

func (f *fakeInventoryService) CreateSchoolRoom(_ context.Context, arg db.CreateSchoolRoomParams) (db.SchoolRoom, error) {
	f.roomCreate = arg
	if f.roomRow.ID.Valid {
		return f.roomRow, nil
	}
	return db.SchoolRoom{ID: handlerTestUUID(150), Code: arg.Code, Name: arg.Name, ExamCapacity: arg.ExamCapacity, Condition: arg.Condition, IsExamEligible: arg.IsExamEligible}, nil
}

func (f *fakeInventoryService) UpdateSchoolRoom(_ context.Context, arg db.UpdateSchoolRoomParams) (db.SchoolRoom, error) {
	f.roomUpdate = arg
	return db.SchoolRoom{ID: arg.ID, Code: arg.Code, Name: arg.Name, ExamCapacity: arg.ExamCapacity, Condition: arg.Condition, IsExamEligible: arg.IsExamEligible}, nil
}

func (f *fakeInventoryService) DeleteSchoolRoom(_ context.Context, id pgtype.UUID) error {
	f.roomID = id
	return nil
}

func inventoryTestItem(id pgtype.UUID, nama string) db.InventoryItem {
	return db.InventoryItem{
		ID:          id,
		Kode:        "INV-001",
		Nama:        nama,
		Kategori:    "laboratorium",
		Lokasi:      "Lab IPA",
		Kondisi:     "baik",
		Satuan:      "unit",
		JumlahTotal: 4,
		JumlahBaik:  3,
		MinStock:    1,
		Catatan:     "Aman",
	}
}

func TestInventorySuccessHandlersForwardPayloads(t *testing.T) {
	itemID := handlerTestUUID(118)
	itemID2 := handlerTestUUID(119)
	location := "Perpustakaan"
	condition := "perlu-perawatan"
	fake := &fakeInventoryService{
		Inventory:    &service.Inventory{},
		statsRow:     db.GetInventoryStatsRow{TotalJenis: 2, PerluRestok: 1},
		listRows:     []db.InventoryItem{inventoryTestItem(itemID, "Mikroskop")},
		eventsRows:   []db.ListInventoryItemEventsByItemRow{{ID: handlerTestUUID(120), ItemID: itemID, Action: "create", Summary: "Barang dibuat"}},
		createRow:    inventoryTestItem(itemID, "Mikroskop"),
		batchUpdated: 2,
		updateRow:    inventoryTestItem(itemID, "Mikroskop Revisi"),
	}
	h := &Inventory{svc: fake}

	rec := httptest.NewRecorder()
	h.Stats(rec, adminRequest(http.MethodGet, "/api/inventory/stats", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("Stats status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListItems(rec, adminRequest(http.MethodGet, "/api/inventory/items?search=mikro&kategori=laboratorium&kondisi=baik", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListItems status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listSearch != "mikro" || fake.listKategori != "laboratorium" || fake.listKondisi != "baik" {
		t.Fatalf("ListItems filters = (%q, %q, %q), want query filters", fake.listSearch, fake.listKategori, fake.listKondisi)
	}

	rec = httptest.NewRecorder()
	h.ListItemEvents(rec, withRouteParam(adminRequest(http.MethodGet, "/api/inventory/items/"+itemID.String()+"/events", ""), "id", itemID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListItemEvents status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.eventsID != itemID {
		t.Fatalf("ListItemEvents id = %v, want %v", fake.eventsID, itemID)
	}

	rec = httptest.NewRecorder()
	h.CreateItem(rec, adminRequest(http.MethodPost, "/api/inventory/items", `{"kode":" INV-002 ","nama":" Mikroskop ","kategori":" lab ","lokasi":" Lab IPA ","kondisi":" baik ","satuan":" unit ","jumlah_total":5,"jumlah_baik":4,"min_stock":1,"catatan":" Baru "}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateItem status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.createActorID.Valid || fake.createArg.Kode != "INV-002" || fake.createArg.Nama != "Mikroskop" || fake.createArg.Kategori != "lab" || fake.createArg.Lokasi != "Lab IPA" || fake.createArg.Kondisi != "baik" || fake.createArg.Satuan != "unit" || fake.createArg.Catatan != "Baru" {
		t.Fatalf("CreateItem actor=%v arg=%+v, want trimmed payload with actor", fake.createActorID, fake.createArg)
	}

	rec = httptest.NewRecorder()
	h.BatchUpdateItems(rec, adminRequest(http.MethodPatch, "/api/inventory/items/batch", `{"item_ids":["`+itemID.String()+`","`+itemID2.String()+`"],"lokasi":"`+location+`","kondisi":"`+condition+`"}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("BatchUpdateItems status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.batchActorID.Valid || len(fake.batchIDs) != 2 || *fake.batchLokasi != location || *fake.batchKondisi != condition || !strings.Contains(rec.Body.String(), `"updated":2`) {
		t.Fatalf("BatchUpdateItems actor=%v ids=%v lokasi=%v kondisi=%v body=%s", fake.batchActorID, fake.batchIDs, fake.batchLokasi, fake.batchKondisi, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdateItem(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/inventory/items/"+itemID.String(), `{"kode":"INV-002","nama":"Mikroskop Revisi","kategori":"lab","lokasi":"Ruang TU","kondisi":"baik","satuan":"unit","jumlah_total":5,"jumlah_baik":5,"min_stock":1,"catatan":"Lengkap"}`), "id", itemID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateItem status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.updateActorID.Valid || fake.updateArg.ID != itemID || fake.updateArg.Nama != "Mikroskop Revisi" || fake.updateArg.Lokasi != "Ruang TU" {
		t.Fatalf("UpdateItem actor=%v arg=%+v, want route id and decoded payload", fake.updateActorID, fake.updateArg)
	}

	rec = httptest.NewRecorder()
	h.DeleteItem(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/inventory/items/"+itemID.String(), ""), "id", itemID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteItem status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.deleteActorID.Valid || fake.deleteID != itemID {
		t.Fatalf("DeleteItem actor=%v id=%v, want actor and item id", fake.deleteActorID, fake.deleteID)
	}
}

func TestInventoryMutationHandlersWriteAuditEvents(t *testing.T) {
	itemID := handlerTestUUID(143)
	itemID2 := handlerTestUUID(144)
	location := "Perpustakaan"
	condition := "perlu-perawatan"
	fake := &fakeInventoryService{
		Inventory:    &service.Inventory{},
		createRow:    inventoryTestItem(itemID, "Mikroskop"),
		batchUpdated: 2,
		updateRow:    inventoryTestItem(itemID, "Mikroskop Revisi"),
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &Inventory{svc: fake, audit: audit}
	auditedStaffRequest := func(method, target, body string) *http.Request {
		return withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"staf"},
			"uid":   "01000000-0000-0000-0000-000000000000",
			"sub":   "01000000-0000-0000-0000-000000000000",
			"usr":   "staf.sarpras",
			"ssid":  "sess-inventory-1",
		})
	}

	rec := httptest.NewRecorder()
	h.CreateItem(rec, auditedStaffRequest(http.MethodPost, "/api/inventory/items", `{"kode":"INV-002","nama":"Mikroskop","kategori":"lab","lokasi":"Lab IPA","kondisi":"baik","satuan":"unit","jumlah_total":5,"jumlah_baik":4,"min_stock":1,"catatan":"Baru"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateItem status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.BatchUpdateItems(rec, auditedStaffRequest(http.MethodPatch, "/api/inventory/items/batch", `{"item_ids":["`+itemID.String()+`","`+itemID2.String()+`"],"lokasi":"`+location+`","kondisi":"`+condition+`"}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("BatchUpdateItems status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdateItem(rec, withRouteParam(auditedStaffRequest(http.MethodPatch, "/api/inventory/items/"+itemID.String(), `{"kode":"INV-002","nama":"Mikroskop Revisi","kategori":"lab","lokasi":"Ruang TU","kondisi":"baik","satuan":"unit","jumlah_total":5,"jumlah_baik":5,"min_stock":1,"catatan":"Lengkap"}`), "id", itemID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateItem status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.DeleteItem(rec, withRouteParam(auditedStaffRequest(http.MethodDelete, "/api/inventory/items/"+itemID.String(), ""), "id", itemID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteItem status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}

	if len(audit.entries) != 4 {
		t.Fatalf("audit entries = %d, want 4", len(audit.entries))
	}

	checks := []struct {
		index      int
		action     string
		entityType string
		entityID   string
	}{
		{0, "INVENTORY_ITEM_CREATE", "inventory_item", itemID.String()},
		{1, "INVENTORY_ITEM_BATCH_UPDATE", "inventory_item", ""},
		{2, "INVENTORY_ITEM_UPDATE", "inventory_item", itemID.String()},
		{3, "INVENTORY_ITEM_DELETE", "inventory_item", itemID.String()},
	}
	for _, check := range checks {
		got := audit.entries[check.index]
		if got.Action != check.action || got.EntityType != check.entityType || got.EntityID != check.entityID {
			t.Fatalf("audit[%d] = %+v, want action/type/id %q/%q/%q", check.index, got, check.action, check.entityType, check.entityID)
		}
	}

	meta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if meta["username"] != "staf.sarpras" || meta["nama"] != "Mikroskop" {
		t.Fatalf("create item metadata = %+v, want username/nama", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[1].Metadata)
	itemIDs, ok := meta["item_ids"].([]any)
	if !ok || len(itemIDs) != 2 || meta["updated_count"] != float64(2) {
		t.Fatalf("batch update metadata = %+v, want item_ids and updated_count", meta)
	}

	meta = mustAuditMetadataMap(t, audit.entries[3].Metadata)
	if meta["deleted_by"] != "staf.sarpras" {
		t.Fatalf("delete item metadata = %+v, want deleted_by", meta)
	}
}

func TestInventoryHandlersMapServiceErrors(t *testing.T) {
	itemID := handlerTestUUID(121)
	errDB := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*Inventory, http.ResponseWriter, *http.Request)
		svc        *fakeInventoryService
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "stats",
			handler:    (*Inventory).Stats,
			svc:        &fakeInventoryService{Inventory: &service.Inventory{}, statsErr: errDB},
			req:        adminRequest(http.MethodGet, "/api/inventory/stats", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "list",
			handler:    (*Inventory).ListItems,
			svc:        &fakeInventoryService{Inventory: &service.Inventory{}, listErr: errDB},
			req:        adminRequest(http.MethodGet, "/api/inventory/items", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "events",
			handler:    (*Inventory).ListItemEvents,
			svc:        &fakeInventoryService{Inventory: &service.Inventory{}, eventsErr: errDB},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/inventory/items/"+itemID.String()+"/events", ""), "id", itemID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "create validation",
			handler:    (*Inventory).CreateItem,
			svc:        &fakeInventoryService{Inventory: &service.Inventory{}, createErr: errors.New("kode inventaris wajib diisi")},
			req:        adminRequest(http.MethodPost, "/api/inventory/items", `{}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "batch validation",
			handler:    (*Inventory).BatchUpdateItems,
			svc:        &fakeInventoryService{Inventory: &service.Inventory{}, batchErr: errors.New("pilih minimal satu barang")},
			req:        adminRequest(http.MethodPatch, "/api/inventory/items/batch", `{"item_ids":[],"lokasi":"Lab"}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update validation",
			handler:    (*Inventory).UpdateItem,
			svc:        &fakeInventoryService{Inventory: &service.Inventory{}, updateErr: errors.New("kondisi inventaris tidak valid")},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/inventory/items/"+itemID.String(), `{}`), "id", itemID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete validation",
			handler:    (*Inventory).DeleteItem,
			svc:        &fakeInventoryService{Inventory: &service.Inventory{}, deleteErr: errors.New("barang tidak ditemukan")},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/api/inventory/items/"+itemID.String(), ""), "id", itemID.String()),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Inventory{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestInventoryValidationBranches(t *testing.T) {
	itemID := handlerTestUUID(122)
	plainRequest := func(method, target, body string) *http.Request {
		return httptest.NewRequest(method, target, strings.NewReader(body))
	}
	tests := []struct {
		name       string
		handler    func(*Inventory, http.ResponseWriter, *http.Request)
		req        *http.Request
		wantStatus int
	}{
		{name: "list forbidden", handler: (*Inventory).ListItems, req: plainRequest(http.MethodGet, "/api/inventory/items", ""), wantStatus: http.StatusForbidden},
		{name: "events forbidden", handler: (*Inventory).ListItemEvents, req: withRouteParam(plainRequest(http.MethodGet, "/api/inventory/items/"+itemID.String()+"/events", ""), "id", itemID.String()), wantStatus: http.StatusForbidden},
		{name: "events invalid id", handler: (*Inventory).ListItemEvents, req: withRouteParam(adminRequest(http.MethodGet, "/api/inventory/items/bad/events", ""), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "create forbidden", handler: (*Inventory).CreateItem, req: plainRequest(http.MethodPost, "/api/inventory/items", `{}`), wantStatus: http.StatusForbidden},
		{name: "create invalid json", handler: (*Inventory).CreateItem, req: adminRequest(http.MethodPost, "/api/inventory/items", `{`), wantStatus: http.StatusBadRequest},
		{name: "batch forbidden", handler: (*Inventory).BatchUpdateItems, req: plainRequest(http.MethodPatch, "/api/inventory/items/batch", `{}`), wantStatus: http.StatusForbidden},
		{name: "batch invalid json", handler: (*Inventory).BatchUpdateItems, req: adminRequest(http.MethodPatch, "/api/inventory/items/batch", `{`), wantStatus: http.StatusBadRequest},
		{name: "batch invalid item id", handler: (*Inventory).BatchUpdateItems, req: adminRequest(http.MethodPatch, "/api/inventory/items/batch", `{"item_ids":["bad"]}`), wantStatus: http.StatusBadRequest},
		{name: "update forbidden", handler: (*Inventory).UpdateItem, req: withRouteParam(plainRequest(http.MethodPatch, "/api/inventory/items/"+itemID.String(), `{}`), "id", itemID.String()), wantStatus: http.StatusForbidden},
		{name: "update invalid id", handler: (*Inventory).UpdateItem, req: withRouteParam(adminRequest(http.MethodPatch, "/api/inventory/items/bad", `{}`), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "update invalid json", handler: (*Inventory).UpdateItem, req: withRouteParam(adminRequest(http.MethodPatch, "/api/inventory/items/"+itemID.String(), `{`), "id", itemID.String()), wantStatus: http.StatusBadRequest},
		{name: "delete forbidden", handler: (*Inventory).DeleteItem, req: withRouteParam(plainRequest(http.MethodDelete, "/api/inventory/items/"+itemID.String(), ""), "id", itemID.String()), wantStatus: http.StatusForbidden},
		{name: "delete invalid id", handler: (*Inventory).DeleteItem, req: withRouteParam(adminRequest(http.MethodDelete, "/api/inventory/items/bad", ""), "id", "bad"), wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&Inventory{svc: &fakeInventoryService{Inventory: &service.Inventory{}}}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
