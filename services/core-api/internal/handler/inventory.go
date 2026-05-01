package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Inventory struct {
	svc   inventoryService
	audit cbtAuthoringAuditWriter
}

type inventoryService interface {
	Stats(ctx context.Context) (db.GetInventoryStatsRow, error)
	ListItems(ctx context.Context, search, kategori, kondisi string) ([]db.InventoryItem, error)
	ListItemEvents(ctx context.Context, itemID pgtype.UUID) ([]db.ListInventoryItemEventsByItemRow, error)
	CreateItem(ctx context.Context, actorUserID pgtype.UUID, arg db.CreateInventoryItemParams) (db.InventoryItem, error)
	BatchUpdateItems(ctx context.Context, actorUserID pgtype.UUID, ids []pgtype.UUID, lokasi *string, kondisi *string) (int, error)
	UpdateItem(ctx context.Context, actorUserID pgtype.UUID, arg db.UpdateInventoryItemParams) (db.InventoryItem, error)
	DeleteItem(ctx context.Context, actorUserID pgtype.UUID, id pgtype.UUID) error
}

func NewInventory(svc *service.Inventory, audit ...cbtAuthoringAuditWriter) *Inventory {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &Inventory{svc: svc, audit: writer}
}

func inventoryActorUserID(r *http.Request) pgtype.UUID {
	var uid pgtype.UUID
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if raw, ok := claims["uid"].(string); ok {
			_ = uid.Scan(raw)
		}
	}
	return uid
}

func inventoryAccessAllowed(r *http.Request) bool {
	return libraryAccessAllowed(r)
}

func (h *Inventory) Stats(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.Stats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Inventory) ListItems(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	search := r.URL.Query().Get("search")
	kategori := r.URL.Query().Get("kategori")
	kondisi := r.URL.Query().Get("kondisi")
	items, err := h.svc.ListItems(r.Context(), search, kategori, kondisi)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, items)
}

func (h *Inventory) ListItemEvents(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	events, err := h.svc.ListItemEvents(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, events)
}

func (h *Inventory) CreateItem(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		Kode        string `json:"kode"`
		Nama        string `json:"nama"`
		Kategori    string `json:"kategori"`
		Lokasi      string `json:"lokasi"`
		Kondisi     string `json:"kondisi"`
		Satuan      string `json:"satuan"`
		JumlahTotal int32  `json:"jumlah_total"`
		JumlahBaik  int32  `json:"jumlah_baik"`
		MinStock    int32  `json:"min_stock"`
		Catatan     string `json:"catatan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	item, err := h.svc.CreateItem(r.Context(), inventoryActorUserID(r), db.CreateInventoryItemParams{
		Kode:        strings.TrimSpace(body.Kode),
		Nama:        strings.TrimSpace(body.Nama),
		Kategori:    strings.TrimSpace(body.Kategori),
		Lokasi:      strings.TrimSpace(body.Lokasi),
		Kondisi:     strings.TrimSpace(body.Kondisi),
		Satuan:      strings.TrimSpace(body.Satuan),
		JumlahTotal: body.JumlahTotal,
		JumlahBaik:  body.JumlahBaik,
		MinStock:    body.MinStock,
		Catatan:     strings.TrimSpace(body.Catatan),
	})
	if err != nil {
		writeClientError(w, err, "Data inventaris tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "INVENTORY_ITEM_CREATE", "inventory_item", pgUUIDString(item.ID), map[string]any{
		"kode":     item.Kode,
		"nama":     item.Nama,
		"kategori": item.Kategori,
		"lokasi":   item.Lokasi,
		"kondisi":  item.Kondisi,
	})
	api.Created(w, item)
}

func (h *Inventory) BatchUpdateItems(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		ItemIDs []string `json:"item_ids"`
		Lokasi  *string  `json:"lokasi"`
		Kondisi *string  `json:"kondisi"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	ids := make([]pgtype.UUID, 0, len(body.ItemIDs))
	for _, rawID := range body.ItemIDs {
		id, err := parseUUID(rawID)
		if err != nil {
			api.BadRequest(w, "item_ids invalid")
			return
		}
		ids = append(ids, id)
	}
	updated, err := h.svc.BatchUpdateItems(r.Context(), inventoryActorUserID(r), ids, body.Lokasi, body.Kondisi)
	if err != nil {
		writeClientError(w, err, "Perubahan inventaris tidak valid")
		return
	}
	meta := map[string]any{
		"updated_count": updated,
		"item_ids":      body.ItemIDs,
	}
	if body.Lokasi != nil {
		meta["lokasi"] = strings.TrimSpace(*body.Lokasi)
	}
	if body.Kondisi != nil {
		meta["kondisi"] = strings.TrimSpace(*body.Kondisi)
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "INVENTORY_ITEM_BATCH_UPDATE", "inventory_item", "", meta)
	api.OK(w, map[string]any{"updated": updated})
}

func (h *Inventory) UpdateItem(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Kode        string `json:"kode"`
		Nama        string `json:"nama"`
		Kategori    string `json:"kategori"`
		Lokasi      string `json:"lokasi"`
		Kondisi     string `json:"kondisi"`
		Satuan      string `json:"satuan"`
		JumlahTotal int32  `json:"jumlah_total"`
		JumlahBaik  int32  `json:"jumlah_baik"`
		MinStock    int32  `json:"min_stock"`
		Catatan     string `json:"catatan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	item, err := h.svc.UpdateItem(r.Context(), inventoryActorUserID(r), db.UpdateInventoryItemParams{
		ID:          id,
		Kode:        strings.TrimSpace(body.Kode),
		Nama:        strings.TrimSpace(body.Nama),
		Kategori:    strings.TrimSpace(body.Kategori),
		Lokasi:      strings.TrimSpace(body.Lokasi),
		Kondisi:     strings.TrimSpace(body.Kondisi),
		Satuan:      strings.TrimSpace(body.Satuan),
		JumlahTotal: body.JumlahTotal,
		JumlahBaik:  body.JumlahBaik,
		MinStock:    body.MinStock,
		Catatan:     strings.TrimSpace(body.Catatan),
	})
	if err != nil {
		writeClientError(w, err, "Perubahan inventaris tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "INVENTORY_ITEM_UPDATE", "inventory_item", pgUUIDString(item.ID), map[string]any{
		"kode":     item.Kode,
		"nama":     item.Nama,
		"kategori": item.Kategori,
		"lokasi":   item.Lokasi,
		"kondisi":  item.Kondisi,
	})
	api.OK(w, item)
}

func (h *Inventory) DeleteItem(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.DeleteItem(r.Context(), inventoryActorUserID(r), id); err != nil {
		writeClientError(w, err, "Penghapusan inventaris tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "INVENTORY_ITEM_DELETE", "inventory_item", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
	api.NoContent(w)
}
