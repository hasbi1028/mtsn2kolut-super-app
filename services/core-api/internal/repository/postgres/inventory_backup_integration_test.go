package db

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationInventoryRepositoryQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	actor, err := q.CreateUserWithMustChangePassword(ctx, CreateUserWithMustChangePasswordParams{
		Username:           "it-inv-actor-" + suffix,
		PasswordHash:       "hash:v1:" + suffix,
		DisplayName:        pgtype.Text{String: "Inventory Actor " + suffix, Valid: true},
		IsActive:           true,
		MustChangePassword: false,
	})
	if err != nil {
		t.Fatalf("create inventory actor user: %v", err)
	}

	projector, err := q.CreateInventoryItem(ctx, CreateInventoryItemParams{
		Kode:        "IT-INV-PROJ-" + suffix,
		Nama:        "Integration Projector " + suffix,
		Kategori:    "elektronik",
		Lokasi:      "Lab Komputer",
		Kondisi:     "baik",
		Satuan:      "unit",
		JumlahTotal: 10,
		JumlahBaik:  8,
		MinStock:    3,
		Catatan:     "created by integration test",
	})
	if err != nil {
		t.Fatalf("create projector inventory item: %v", err)
	}
	if !projector.ID.Valid || projector.Kode != "IT-INV-PROJ-"+suffix || projector.JumlahTotal != 10 {
		t.Fatalf("created projector inventory item = %#v", projector)
	}

	chair, err := q.CreateInventoryItem(ctx, CreateInventoryItemParams{
		Kode:        "IT-INV-CHAIR-" + suffix,
		Nama:        "Integration Chair " + suffix,
		Kategori:    "mebel",
		Lokasi:      "Ruang Guru",
		Kondisi:     "perlu-perawatan",
		Satuan:      "unit",
		JumlahTotal: 5,
		JumlahBaik:  1,
		MinStock:    2,
		Catatan:     "low usable stock",
	})
	if err != nil {
		t.Fatalf("create chair inventory item: %v", err)
	}

	gotProjector, err := q.GetInventoryItem(ctx, projector.ID)
	if err != nil {
		t.Fatalf("get inventory item: %v", err)
	}
	if gotProjector.ID != projector.ID || gotProjector.Nama != projector.Nama {
		t.Fatalf("got projector = %#v, want id %v name %q", gotProjector, projector.ID, projector.Nama)
	}

	updatedProjector, err := q.UpdateInventoryItem(ctx, UpdateInventoryItemParams{
		ID:          projector.ID,
		Kode:        projector.Kode,
		Nama:        "Integration Projector Updated " + suffix,
		Kategori:    "elektronik",
		Lokasi:      "Lab Bahasa",
		Kondisi:     "rusak",
		Satuan:      "unit",
		JumlahTotal: 10,
		JumlahBaik:  2,
		MinStock:    3,
		Catatan:     "updated by integration test",
	})
	if err != nil {
		t.Fatalf("update inventory item: %v", err)
	}
	if updatedProjector.Nama != "Integration Projector Updated "+suffix || updatedProjector.Kondisi != "rusak" || updatedProjector.Lokasi != "Lab Bahasa" {
		t.Fatalf("updated projector = %#v", updatedProjector)
	}

	allItems, err := q.ListInventoryItems(ctx, ListInventoryItemsParams{})
	if err != nil {
		t.Fatalf("list all inventory items: %v", err)
	}
	if !inventoryItemListed(allItems, projector.ID) || !inventoryItemListed(allItems, chair.ID) {
		t.Fatalf("all inventory items missing created rows: %#v", allItems)
	}

	filteredItems, err := q.ListInventoryItems(ctx, ListInventoryItemsParams{Search: "Lab", Kategori: "elektronik", Kondisi: "rusak"})
	if err != nil {
		t.Fatalf("list filtered inventory items: %v", err)
	}
	if len(filteredItems) != 1 || filteredItems[0].ID != projector.ID {
		t.Fatalf("filtered inventory items = %#v, want only updated projector", filteredItems)
	}

	stats, err := q.GetInventoryStats(ctx)
	if err != nil {
		t.Fatalf("get inventory stats: %v", err)
	}
	if stats.TotalJenis != 2 || fmt.Sprint(stats.TotalUnit) != "15" || fmt.Sprint(stats.TotalLayak) != "3" || stats.PerluRestok != 2 || stats.PerluPerawatan != 2 {
		t.Fatalf("inventory stats = %#v, want totals for created and updated items", stats)
	}

	event, err := q.CreateInventoryItemEvent(ctx, CreateInventoryItemEventParams{
		ItemID:      projector.ID,
		ActorUserID: actor.ID,
		Action:      "update",
		Summary:     "Marked projector as rusak during integration test",
	})
	if err != nil {
		t.Fatalf("create inventory item event: %v", err)
	}
	if !event.ID.Valid || event.ItemID != projector.ID || event.ActorUserID != actor.ID {
		t.Fatalf("created inventory event = %#v", event)
	}

	events, err := q.ListInventoryItemEventsByItem(ctx, projector.ID)
	if err != nil {
		t.Fatalf("list inventory item events by item: %v", err)
	}
	if len(events) != 1 || events[0].ID != event.ID || events[0].ActorUsername != actor.Username || events[0].ActorDisplayName != "Inventory Actor "+suffix {
		t.Fatalf("inventory item events = %#v, want created event with actor details", events)
	}

	if err := q.DeleteInventoryItem(ctx, chair.ID); err != nil {
		t.Fatalf("delete inventory item: %v", err)
	}
	allItems, err = q.ListInventoryItems(ctx, ListInventoryItemsParams{})
	if err != nil {
		t.Fatalf("list inventory items after delete: %v", err)
	}
	if inventoryItemListed(allItems, chair.ID) {
		t.Fatalf("deleted inventory item %v still listed: %#v", chair.ID, allItems)
	}
}

func TestIntegrationSystemMaintenanceRepositoryQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	actor, err := q.CreateUserWithMustChangePassword(ctx, CreateUserWithMustChangePasswordParams{
		Username:           "it-maint-actor-" + suffix,
		PasswordHash:       "hash:v1:" + suffix,
		DisplayName:        pgtype.Text{String: "Maintenance Actor " + suffix, Valid: true},
		IsActive:           true,
		MustChangePassword: false,
	})
	if err != nil {
		t.Fatalf("create maintenance actor user: %v", err)
	}

	dbTime, err := q.GetMaintenanceDatabaseTime(ctx)
	if err != nil {
		t.Fatalf("get maintenance database time: %v", err)
	}
	if !dbTime.Valid {
		t.Fatalf("maintenance database time is invalid: %#v", dbTime)
	}

	activeSessions, err := q.CountActiveCbtSessionsForMaintenance(ctx)
	if err != nil {
		t.Fatalf("count active CBT sessions for maintenance: %v", err)
	}
	if activeSessions != 0 {
		t.Fatalf("active CBT session count = %d, want 0 in isolated integration database", activeSessions)
	}

	now := time.Now().UTC()
	window, err := q.CreateMaintenanceWindow(ctx, CreateMaintenanceWindowParams{
		Title:            "Integration Maintenance " + suffix,
		Message:          "System maintenance integration test",
		Mode:             "module",
		AffectedModules:  []string{"inventory", "backup"},
		StartsAt:         pgtype.Timestamptz{Time: now.Add(-time.Minute), Valid: true},
		EndsAt:           pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true},
		IsActive:         true,
		AllowAdminBypass: true,
		BypassRoles:      []string{"admin", "staf"},
		Severity:         "warning",
		ActorUserID:      actor.ID,
	})
	if err != nil {
		t.Fatalf("create maintenance window: %v", err)
	}
	if !window.ID.Valid || window.CreatedBy != actor.ID || window.UpdatedBy != actor.ID || !window.IsActive {
		t.Fatalf("created maintenance window = %#v", window)
	}
	if !slices.Contains(window.AffectedModules, "inventory") || !slices.Contains(window.AffectedModules, "backup") {
		t.Fatalf("maintenance window affected modules = %v", window.AffectedModules)
	}

	gotWindow, err := q.GetMaintenanceWindow(ctx, window.ID)
	if err != nil {
		t.Fatalf("get maintenance window: %v", err)
	}
	if gotWindow.ID != window.ID || gotWindow.Title != window.Title || gotWindow.Mode != "module" {
		t.Fatalf("got maintenance window = %#v", gotWindow)
	}

	activeWindow, err := q.GetActiveMaintenanceWindow(ctx)
	if err != nil {
		t.Fatalf("get active maintenance window: %v", err)
	}
	if activeWindow.ID != window.ID {
		t.Fatalf("active maintenance window = %#v, want %v", activeWindow, window.ID)
	}

	windows, err := q.ListMaintenanceWindows(ctx, ListMaintenanceWindowsParams{OffsetCount: 0, LimitCount: 10})
	if err != nil {
		t.Fatalf("list maintenance windows: %v", err)
	}
	if !maintenanceWindowListed(windows, window.ID) {
		t.Fatalf("created maintenance window %v not listed: %#v", window.ID, windows)
	}

	updated, err := q.UpdateMaintenanceWindow(ctx, UpdateMaintenanceWindowParams{
		Title:            "Integration Maintenance Updated " + suffix,
		Message:          "Updated system maintenance integration test",
		Mode:             "read_only",
		AffectedModules:  []string{"global"},
		StartsAt:         pgtype.Timestamptz{Time: now.Add(-2 * time.Minute), Valid: true},
		EndsAt:           pgtype.Timestamptz{Time: now.Add(2 * time.Hour), Valid: true},
		AllowAdminBypass: false,
		BypassRoles:      []string{"admin"},
		Severity:         "critical",
		ActorUserID:      actor.ID,
		ID:               window.ID,
	})
	if err != nil {
		t.Fatalf("update maintenance window: %v", err)
	}
	if updated.Title != "Integration Maintenance Updated "+suffix || updated.Mode != "read_only" || updated.AllowAdminBypass || updated.Severity != "critical" {
		t.Fatalf("updated maintenance window = %#v", updated)
	}

	audit, err := q.CreateMaintenanceAuditLog(ctx, CreateMaintenanceAuditLogParams{
		MaintenanceID: window.ID,
		ActorUserID:   actor.ID,
		Action:        "update",
		Reason:        pgtype.Text{String: "integration maintenance audit", Valid: true},
		Metadata:      []byte(`{"source":"integration-test","module":"backup"}`),
	})
	if err != nil {
		t.Fatalf("create maintenance audit log: %v", err)
	}
	if !audit.ID.Valid || audit.MaintenanceID != window.ID || audit.ActorUserID != actor.ID || string(audit.Metadata) == "" {
		t.Fatalf("created maintenance audit log = %#v", audit)
	}

	auditLogs, err := q.ListMaintenanceAuditLogs(ctx, ListMaintenanceAuditLogsParams{ActionFilter: "update", OffsetCount: 0, LimitCount: 10})
	if err != nil {
		t.Fatalf("list maintenance audit logs: %v", err)
	}
	if !maintenanceAuditLogListed(auditLogs, audit.ID) {
		t.Fatalf("created maintenance audit log %v not listed: %#v", audit.ID, auditLogs)
	}

	inactive, err := q.SetMaintenanceWindowActive(ctx, SetMaintenanceWindowActiveParams{IsActive: false, ActorUserID: actor.ID, ID: window.ID})
	if err != nil {
		t.Fatalf("set maintenance window inactive: %v", err)
	}
	if inactive.IsActive {
		t.Fatalf("maintenance window still active after deactivation: %#v", inactive)
	}
	_, err = q.GetActiveMaintenanceWindow(ctx)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("get active maintenance window after deactivation error = %v, want pgx.ErrNoRows", err)
	}
}

func inventoryItemListed(items []InventoryItem, id pgtype.UUID) bool {
	return slices.ContainsFunc(items, func(item InventoryItem) bool {
		return item.ID == id
	})
}

func maintenanceWindowListed(windows []SystemMaintenanceWindow, id pgtype.UUID) bool {
	return slices.ContainsFunc(windows, func(window SystemMaintenanceWindow) bool {
		return window.ID == id
	})
}

func maintenanceAuditLogListed(logs []SystemMaintenanceAuditLog, id pgtype.UUID) bool {
	return slices.ContainsFunc(logs, func(log SystemMaintenanceAuditLog) bool {
		return log.ID == id
	})
}
