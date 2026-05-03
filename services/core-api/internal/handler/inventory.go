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
	ListSchoolRooms(ctx context.Context, search, roomType, condition, examEligible string) ([]db.SchoolRoom, error)
	GetSchoolRoom(ctx context.Context, id pgtype.UUID) (db.SchoolRoom, error)
	CreateSchoolRoom(ctx context.Context, arg db.CreateSchoolRoomParams) (db.SchoolRoom, error)
	UpdateSchoolRoom(ctx context.Context, arg db.UpdateSchoolRoomParams) (db.SchoolRoom, error)
	DeleteSchoolRoom(ctx context.Context, id pgtype.UUID) error
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

func (h *Inventory) ListSchoolRooms(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rooms, err := h.svc.ListSchoolRooms(
		r.Context(),
		r.URL.Query().Get("search"),
		r.URL.Query().Get("room_type"),
		r.URL.Query().Get("condition"),
		r.URL.Query().Get("exam_eligible"),
	)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rooms)
}

func (h *Inventory) GetSchoolRoom(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	room, err := h.svc.GetSchoolRoom(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, room)
}

func (h *Inventory) CreateSchoolRoom(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	body, ok := decodeSchoolRoomPayload(w, r)
	if !ok {
		return
	}
	room, err := h.svc.CreateSchoolRoom(r.Context(), body.toCreateParams())
	if err != nil {
		writeClientError(w, err, "Data ruangan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "SCHOOL_ROOM_CREATE", "school_room", pgUUIDString(room.ID), map[string]any{
		"code":             room.Code,
		"name":             room.Name,
		"room_type":        room.RoomType,
		"exam_capacity":    room.ExamCapacity,
		"is_exam_eligible": room.IsExamEligible,
	})
	api.Created(w, room)
}

func (h *Inventory) UpdateSchoolRoom(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	body, ok := decodeSchoolRoomPayload(w, r)
	if !ok {
		return
	}
	arg := body.toUpdateParams()
	arg.ID = id
	room, err := h.svc.UpdateSchoolRoom(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Perubahan ruangan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "SCHOOL_ROOM_UPDATE", "school_room", pgUUIDString(room.ID), map[string]any{
		"code":             room.Code,
		"name":             room.Name,
		"room_type":        room.RoomType,
		"exam_capacity":    room.ExamCapacity,
		"is_exam_eligible": room.IsExamEligible,
	})
	api.OK(w, room)
}

func (h *Inventory) DeleteSchoolRoom(w http.ResponseWriter, r *http.Request) {
	if !inventoryAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.DeleteSchoolRoom(r.Context(), id); err != nil {
		writeClientError(w, err, "Penghapusan ruangan tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "SCHOOL_ROOM_DELETE", "school_room", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
	api.NoContent(w)
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

type schoolRoomPayload struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Building        string `json:"building"`
	Floor           string `json:"floor"`
	RoomType        string `json:"room_type"`
	LocationNote    string `json:"location_note"`
	DefaultCapacity int32  `json:"default_capacity"`
	ExamCapacity    int32  `json:"exam_capacity"`
	Condition       string `json:"condition"`
	IsExamEligible  *bool  `json:"is_exam_eligible"`
	NetworkReady    bool   `json:"network_ready"`
	PowerReady      bool   `json:"power_ready"`
	Notes           string `json:"notes"`
}

func decodeSchoolRoomPayload(w http.ResponseWriter, r *http.Request) (schoolRoomPayload, bool) {
	var body schoolRoomPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return schoolRoomPayload{}, false
	}
	return body, true
}

func (p schoolRoomPayload) examEligible() bool {
	if p.IsExamEligible == nil {
		return true
	}
	return *p.IsExamEligible
}

func (p schoolRoomPayload) toCreateParams() db.CreateSchoolRoomParams {
	return db.CreateSchoolRoomParams{
		Code:            strings.TrimSpace(p.Code),
		Name:            strings.TrimSpace(p.Name),
		Building:        strings.TrimSpace(p.Building),
		Floor:           strings.TrimSpace(p.Floor),
		RoomType:        strings.TrimSpace(p.RoomType),
		LocationNote:    strings.TrimSpace(p.LocationNote),
		DefaultCapacity: p.DefaultCapacity,
		ExamCapacity:    p.ExamCapacity,
		Condition:       strings.TrimSpace(p.Condition),
		IsExamEligible:  p.examEligible(),
		NetworkReady:    p.NetworkReady,
		PowerReady:      p.PowerReady,
		Notes:           strings.TrimSpace(p.Notes),
	}
}

func (p schoolRoomPayload) toUpdateParams() db.UpdateSchoolRoomParams {
	return db.UpdateSchoolRoomParams{
		Code:            strings.TrimSpace(p.Code),
		Name:            strings.TrimSpace(p.Name),
		Building:        strings.TrimSpace(p.Building),
		Floor:           strings.TrimSpace(p.Floor),
		RoomType:        strings.TrimSpace(p.RoomType),
		LocationNote:    strings.TrimSpace(p.LocationNote),
		DefaultCapacity: p.DefaultCapacity,
		ExamCapacity:    p.ExamCapacity,
		Condition:       strings.TrimSpace(p.Condition),
		IsExamEligible:  p.examEligible(),
		NetworkReady:    p.NetworkReady,
		PowerReady:      p.PowerReady,
		Notes:           strings.TrimSpace(p.Notes),
	}
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
