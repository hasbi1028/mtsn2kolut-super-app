package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type inventoryStore interface {
	ListInventoryItems(ctx context.Context, arg db.ListInventoryItemsParams) ([]db.InventoryItem, error)
	GetInventoryItem(ctx context.Context, id pgtype.UUID) (db.InventoryItem, error)
	CreateInventoryItem(ctx context.Context, arg db.CreateInventoryItemParams) (db.InventoryItem, error)
	UpdateInventoryItem(ctx context.Context, arg db.UpdateInventoryItemParams) (db.InventoryItem, error)
	DeleteInventoryItem(ctx context.Context, id pgtype.UUID) error
	CreateInventoryItemEvent(ctx context.Context, arg db.CreateInventoryItemEventParams) (db.InventoryItemEvent, error)
	ListInventoryItemEventsByItem(ctx context.Context, itemID pgtype.UUID) ([]db.ListInventoryItemEventsByItemRow, error)
	GetInventoryStats(ctx context.Context) (db.GetInventoryStatsRow, error)
	ListSchoolRooms(ctx context.Context, arg db.ListSchoolRoomsParams) ([]db.SchoolRoom, error)
	GetSchoolRoom(ctx context.Context, id pgtype.UUID) (db.SchoolRoom, error)
	CreateSchoolRoom(ctx context.Context, arg db.CreateSchoolRoomParams) (db.SchoolRoom, error)
	UpdateSchoolRoom(ctx context.Context, arg db.UpdateSchoolRoomParams) (db.SchoolRoom, error)
	DeleteSchoolRoom(ctx context.Context, id pgtype.UUID) error
}

type Inventory struct{ q inventoryStore }

func NewInventory(q *db.Queries) *Inventory { return &Inventory{q: q} }

func (s *Inventory) ListItems(ctx context.Context, search, kategori, kondisi string) ([]db.InventoryItem, error) {
	return s.q.ListInventoryItems(ctx, db.ListInventoryItemsParams{
		Search:   strings.TrimSpace(search),
		Kategori: strings.TrimSpace(kategori),
		Kondisi:  strings.TrimSpace(kondisi),
	})
}

func (s *Inventory) ListSchoolRooms(ctx context.Context, search, roomType, condition, examEligible string) ([]db.SchoolRoom, error) {
	return s.q.ListSchoolRooms(ctx, db.ListSchoolRoomsParams{
		Search:       strings.TrimSpace(search),
		RoomType:     strings.TrimSpace(roomType),
		Condition:    strings.TrimSpace(condition),
		ExamEligible: normalizeSchoolRoomExamEligibleFilter(examEligible),
	})
}

func (s *Inventory) GetSchoolRoom(ctx context.Context, id pgtype.UUID) (db.SchoolRoom, error) {
	return s.q.GetSchoolRoom(ctx, id)
}

func (s *Inventory) CreateSchoolRoom(ctx context.Context, arg db.CreateSchoolRoomParams) (db.SchoolRoom, error) {
	normalized, err := normalizeCreateSchoolRoom(arg)
	if err != nil {
		return db.SchoolRoom{}, err
	}
	return s.q.CreateSchoolRoom(ctx, normalized)
}

func (s *Inventory) UpdateSchoolRoom(ctx context.Context, arg db.UpdateSchoolRoomParams) (db.SchoolRoom, error) {
	normalized, err := normalizeUpdateSchoolRoom(arg)
	if err != nil {
		return db.SchoolRoom{}, err
	}
	return s.q.UpdateSchoolRoom(ctx, normalized)
}

func (s *Inventory) DeleteSchoolRoom(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteSchoolRoom(ctx, id)
}

func (s *Inventory) GetItem(ctx context.Context, id pgtype.UUID) (db.InventoryItem, error) {
	return s.q.GetInventoryItem(ctx, id)
}

func (s *Inventory) CreateItem(ctx context.Context, actorUserID pgtype.UUID, arg db.CreateInventoryItemParams) (db.InventoryItem, error) {
	if err := validateInventoryPayload(arg.Kode, arg.Nama, arg.Kondisi, arg.JumlahTotal, arg.JumlahBaik, arg.MinStock); err != nil {
		return db.InventoryItem{}, err
	}
	arg.Kode = strings.TrimSpace(arg.Kode)
	arg.Nama = strings.TrimSpace(arg.Nama)
	arg.Kategori = normalizeInventoryText(arg.Kategori, "umum")
	arg.Lokasi = strings.TrimSpace(arg.Lokasi)
	arg.Kondisi = normalizeInventoryText(arg.Kondisi, "baik")
	arg.Satuan = normalizeInventoryText(arg.Satuan, "unit")
	arg.Catatan = strings.TrimSpace(arg.Catatan)
	item, err := s.q.CreateInventoryItem(ctx, arg)
	if err != nil {
		return db.InventoryItem{}, err
	}
	_, _ = s.q.CreateInventoryItemEvent(ctx, db.CreateInventoryItemEventParams{
		ItemID:      item.ID,
		ActorUserID: actorUserID,
		Action:      "create",
		Summary:     fmt.Sprintf("Barang dibuat dengan lokasi '%s', kondisi '%s', dan stok layak %d/%d.", fallbackInventoryText(item.Lokasi), item.Kondisi, item.JumlahBaik, item.JumlahTotal),
	})
	return item, nil
}

func (s *Inventory) UpdateItem(ctx context.Context, actorUserID pgtype.UUID, arg db.UpdateInventoryItemParams) (db.InventoryItem, error) {
	if err := validateInventoryPayload(arg.Kode, arg.Nama, arg.Kondisi, arg.JumlahTotal, arg.JumlahBaik, arg.MinStock); err != nil {
		return db.InventoryItem{}, err
	}
	before, err := s.q.GetInventoryItem(ctx, arg.ID)
	if err != nil {
		return db.InventoryItem{}, err
	}
	arg.Kode = strings.TrimSpace(arg.Kode)
	arg.Nama = strings.TrimSpace(arg.Nama)
	arg.Kategori = normalizeInventoryText(arg.Kategori, "umum")
	arg.Lokasi = strings.TrimSpace(arg.Lokasi)
	arg.Kondisi = normalizeInventoryText(arg.Kondisi, "baik")
	arg.Satuan = normalizeInventoryText(arg.Satuan, "unit")
	arg.Catatan = strings.TrimSpace(arg.Catatan)
	item, err := s.q.UpdateInventoryItem(ctx, arg)
	if err != nil {
		return db.InventoryItem{}, err
	}
	_, _ = s.q.CreateInventoryItemEvent(ctx, db.CreateInventoryItemEventParams{
		ItemID:      item.ID,
		ActorUserID: actorUserID,
		Action:      "update",
		Summary:     summarizeInventoryChange(before, item),
	})
	return item, nil
}

func (s *Inventory) DeleteItem(ctx context.Context, actorUserID pgtype.UUID, id pgtype.UUID) error {
	item, err := s.q.GetInventoryItem(ctx, id)
	if err != nil {
		return err
	}
	_, _ = s.q.CreateInventoryItemEvent(ctx, db.CreateInventoryItemEventParams{
		ItemID:      id,
		ActorUserID: actorUserID,
		Action:      "delete",
		Summary:     fmt.Sprintf("Barang dihapus dari inventaris. Lokasi terakhir '%s', kondisi '%s', stok layak %d/%d.", fallbackInventoryText(item.Lokasi), item.Kondisi, item.JumlahBaik, item.JumlahTotal),
	})
	return s.q.DeleteInventoryItem(ctx, id)
}

func (s *Inventory) Stats(ctx context.Context) (db.GetInventoryStatsRow, error) {
	return s.q.GetInventoryStats(ctx)
}

func (s *Inventory) ListItemEvents(ctx context.Context, itemID pgtype.UUID) ([]db.ListInventoryItemEventsByItemRow, error) {
	return s.q.ListInventoryItemEventsByItem(ctx, itemID)
}

func (s *Inventory) BatchUpdateItems(ctx context.Context, actorUserID pgtype.UUID, ids []pgtype.UUID, lokasi *string, kondisi *string) (int, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("pilih minimal satu barang")
	}
	if lokasi == nil && kondisi == nil {
		return 0, fmt.Errorf("perubahan batch belum diisi")
	}

	updated := 0
	for _, id := range ids {
		current, err := s.q.GetInventoryItem(ctx, id)
		if err != nil {
			return updated, err
		}

		nextLokasi := current.Lokasi
		if lokasi != nil {
			nextLokasi = strings.TrimSpace(*lokasi)
		}

		nextKondisi := current.Kondisi
		if kondisi != nil {
			nextKondisi = normalizeInventoryText(*kondisi, current.Kondisi)
		}

		_, err = s.UpdateItem(ctx, actorUserID, db.UpdateInventoryItemParams{
			ID:          current.ID,
			Kode:        current.Kode,
			Nama:        current.Nama,
			Kategori:    current.Kategori,
			Lokasi:      nextLokasi,
			Kondisi:     nextKondisi,
			Satuan:      current.Satuan,
			JumlahTotal: current.JumlahTotal,
			JumlahBaik:  current.JumlahBaik,
			MinStock:    current.MinStock,
			Catatan:     current.Catatan,
		})
		if err != nil {
			return updated, err
		}
		updated++
	}

	return updated, nil
}

func validateInventoryPayload(kode, nama, kondisi string, jumlahTotal, jumlahBaik, minStock int32) error {
	if strings.TrimSpace(kode) == "" {
		return fmt.Errorf("kode inventaris wajib diisi")
	}
	if strings.TrimSpace(nama) == "" {
		return fmt.Errorf("nama barang wajib diisi")
	}
	if jumlahTotal < 1 {
		return fmt.Errorf("jumlah total minimal 1")
	}
	if jumlahBaik < 0 || jumlahBaik > jumlahTotal {
		return fmt.Errorf("jumlah kondisi baik harus antara 0 dan jumlah total")
	}
	if minStock < 0 {
		return fmt.Errorf("batas restok tidak boleh negatif")
	}
	normalizedCondition := normalizeInventoryText(kondisi, "baik")
	if normalizedCondition != "baik" && normalizedCondition != "perlu-perawatan" && normalizedCondition != "rusak" {
		return fmt.Errorf("kondisi inventaris tidak valid")
	}
	return nil
}

func normalizeCreateSchoolRoom(arg db.CreateSchoolRoomParams) (db.CreateSchoolRoomParams, error) {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.Building = strings.TrimSpace(arg.Building)
	arg.Floor = strings.TrimSpace(arg.Floor)
	arg.RoomType = normalizeInventoryText(arg.RoomType, "kelas")
	arg.LocationNote = strings.TrimSpace(arg.LocationNote)
	arg.Condition = normalizeInventoryText(arg.Condition, "baik")
	arg.Notes = strings.TrimSpace(arg.Notes)
	if arg.DefaultCapacity <= 0 {
		arg.DefaultCapacity = 30
	}
	if arg.ExamCapacity <= 0 {
		arg.ExamCapacity = arg.DefaultCapacity
	}
	if err := validateSchoolRoomPayload(arg.Code, arg.Name, arg.Condition, arg.DefaultCapacity, arg.ExamCapacity); err != nil {
		return db.CreateSchoolRoomParams{}, err
	}
	return arg, nil
}

func normalizeUpdateSchoolRoom(arg db.UpdateSchoolRoomParams) (db.UpdateSchoolRoomParams, error) {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.Building = strings.TrimSpace(arg.Building)
	arg.Floor = strings.TrimSpace(arg.Floor)
	arg.RoomType = normalizeInventoryText(arg.RoomType, "kelas")
	arg.LocationNote = strings.TrimSpace(arg.LocationNote)
	arg.Condition = normalizeInventoryText(arg.Condition, "baik")
	arg.Notes = strings.TrimSpace(arg.Notes)
	if arg.DefaultCapacity <= 0 {
		arg.DefaultCapacity = 30
	}
	if arg.ExamCapacity <= 0 {
		arg.ExamCapacity = arg.DefaultCapacity
	}
	if err := validateSchoolRoomPayload(arg.Code, arg.Name, arg.Condition, arg.DefaultCapacity, arg.ExamCapacity); err != nil {
		return db.UpdateSchoolRoomParams{}, err
	}
	return arg, nil
}

func validateSchoolRoomPayload(code, name, condition string, defaultCapacity, examCapacity int32) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("kode ruangan wajib diisi")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nama ruangan wajib diisi")
	}
	if defaultCapacity < 1 {
		return fmt.Errorf("kapasitas normal minimal 1")
	}
	if examCapacity < 1 {
		return fmt.Errorf("kapasitas ujian minimal 1")
	}
	normalizedCondition := normalizeInventoryText(condition, "baik")
	if normalizedCondition != "baik" && normalizedCondition != "perlu-perawatan" && normalizedCondition != "rusak" {
		return fmt.Errorf("kondisi ruangan tidak valid")
	}
	return nil
}

func normalizeSchoolRoomExamEligibleFilter(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "y":
		return "true"
	case "false", "0", "no", "n":
		return "false"
	default:
		return ""
	}
}

func normalizeInventoryText(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func fallbackInventoryText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "Belum diatur"
	}
	return trimmed
}

func summarizeInventoryChange(before, after db.InventoryItem) string {
	changes := make([]string, 0, 5)
	if before.Lokasi != after.Lokasi {
		changes = append(changes, fmt.Sprintf("lokasi '%s' -> '%s'", fallbackInventoryText(before.Lokasi), fallbackInventoryText(after.Lokasi)))
	}
	if before.Kondisi != after.Kondisi {
		changes = append(changes, fmt.Sprintf("kondisi '%s' -> '%s'", before.Kondisi, after.Kondisi))
	}
	if before.JumlahBaik != after.JumlahBaik || before.JumlahTotal != after.JumlahTotal {
		changes = append(changes, fmt.Sprintf("stok layak %d/%d -> %d/%d", before.JumlahBaik, before.JumlahTotal, after.JumlahBaik, after.JumlahTotal))
	}
	if before.MinStock != after.MinStock {
		changes = append(changes, fmt.Sprintf("batas restok %d -> %d", before.MinStock, after.MinStock))
	}
	if before.Catatan != after.Catatan && strings.TrimSpace(after.Catatan) != "" {
		changes = append(changes, "catatan diperbarui")
	}
	if len(changes) == 0 {
		return "Data barang diperbarui tanpa perubahan lokasi, kondisi, atau stok utama."
	}
	return "Perubahan inventaris: " + strings.Join(changes, "; ") + "."
}
