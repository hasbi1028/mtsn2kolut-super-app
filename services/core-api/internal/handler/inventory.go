package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Inventory struct{ svc *service.Inventory }

func NewInventory(svc *service.Inventory) *Inventory { return &Inventory{svc: svc} }

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
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, err.Error())
		return
	}
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
		api.BadRequest(w, err.Error())
		return
	}
	api.NoContent(w)
}
