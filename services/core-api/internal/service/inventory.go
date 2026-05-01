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
	GetInventoryStats(ctx context.Context) (db.GetInventoryStatsRow, error)
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

func (s *Inventory) GetItem(ctx context.Context, id pgtype.UUID) (db.InventoryItem, error) {
	return s.q.GetInventoryItem(ctx, id)
}

func (s *Inventory) CreateItem(ctx context.Context, arg db.CreateInventoryItemParams) (db.InventoryItem, error) {
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
	return s.q.CreateInventoryItem(ctx, arg)
}

func (s *Inventory) UpdateItem(ctx context.Context, arg db.UpdateInventoryItemParams) (db.InventoryItem, error) {
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
	return s.q.UpdateInventoryItem(ctx, arg)
}

func (s *Inventory) DeleteItem(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteInventoryItem(ctx, id)
}

func (s *Inventory) Stats(ctx context.Context) (db.GetInventoryStatsRow, error) {
	return s.q.GetInventoryStats(ctx)
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

func normalizeInventoryText(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
