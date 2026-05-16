package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const (
	MaintenanceSeverityInfo     = "info"
	MaintenanceSeverityWarning  = "warning"
	MaintenanceSeverityCritical = "critical"

	maintenanceActionCreate     = "create"
	maintenanceActionUpdate     = "update"
	maintenanceActionActivate   = "activate"
	maintenanceActionDeactivate = "deactivate"
)

type maintenanceStore interface {
	GetActiveMaintenanceWindow(ctx context.Context) (db.SystemMaintenanceWindow, error)
	GetMaintenanceWindow(ctx context.Context, id pgtype.UUID) (db.SystemMaintenanceWindow, error)
	ListMaintenanceWindows(ctx context.Context, arg db.ListMaintenanceWindowsParams) ([]db.SystemMaintenanceWindow, error)
	CreateMaintenanceWindow(ctx context.Context, arg db.CreateMaintenanceWindowParams) (db.SystemMaintenanceWindow, error)
	UpdateMaintenanceWindow(ctx context.Context, arg db.UpdateMaintenanceWindowParams) (db.SystemMaintenanceWindow, error)
	SetMaintenanceWindowActive(ctx context.Context, arg db.SetMaintenanceWindowActiveParams) (db.SystemMaintenanceWindow, error)
	CreateMaintenanceAuditLog(ctx context.Context, arg db.CreateMaintenanceAuditLogParams) (db.SystemMaintenanceAuditLog, error)
	ListMaintenanceAuditLogs(ctx context.Context, arg db.ListMaintenanceAuditLogsParams) ([]db.SystemMaintenanceAuditLog, error)
	GetMaintenanceDatabaseTime(ctx context.Context) (pgtype.Timestamptz, error)
	CountActiveCbtSessionsForMaintenance(ctx context.Context) (int64, error)
}

type maintenanceBackupStatusProvider interface {
	Status(ctx context.Context) (SystemBackupStatus, error)
}

type SystemMaintenance struct {
	q      maintenanceStore
	backup maintenanceBackupStatusProvider
	now    func() time.Time
}

type MaintenanceWindowInput struct {
	Title            string
	Message          string
	Mode             string
	AffectedModules  []string
	StartsAt         *time.Time
	EndsAt           *time.Time
	IsActive         bool
	AllowAdminBypass *bool
	BypassRoles      []string
	Severity         string
	ActorUserID      pgtype.UUID
	Reason           string
}

type MaintenanceAuditFilter struct {
	Action string
	FromAt *time.Time
	ToAt   *time.Time
	Limit  int32
	Offset int32
}

type SystemMaintenanceStatus struct {
	Active     bool                   `json:"active"`
	Mode       string                 `json:"mode"`
	ServerTime time.Time              `json:"server_time"`
	Window     *MaintenanceWindowView `json:"window,omitempty"`
}

type MaintenanceWindowView struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Message          string     `json:"message"`
	Mode             string     `json:"mode"`
	AffectedModules  []string   `json:"affected_modules"`
	StartsAt         *time.Time `json:"starts_at,omitempty"`
	EndsAt           *time.Time `json:"ends_at,omitempty"`
	IsActive         bool       `json:"is_active"`
	AllowAdminBypass bool       `json:"allow_admin_bypass"`
	BypassRoles      []string   `json:"bypass_roles"`
	Severity         string     `json:"severity"`
	Status           string     `json:"status"`
	CreatedBy        string     `json:"created_by,omitempty"`
	UpdatedBy        string     `json:"updated_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type MaintenanceAuditLogView struct {
	ID            string         `json:"id"`
	MaintenanceID string         `json:"maintenance_id,omitempty"`
	ActorUserID   string         `json:"actor_user_id,omitempty"`
	Action        string         `json:"action"`
	Reason        string         `json:"reason,omitempty"`
	Metadata      map[string]any `json:"metadata"`
	CreatedAt     time.Time      `json:"created_at"`
}

type MaintenanceHealthSummary struct {
	CoreAPI   MaintenanceHealthComponent `json:"core_api"`
	Database  MaintenanceDatabaseHealth  `json:"database"`
	Backup    MaintenanceBackupHealth    `json:"backup"`
	Disk      MaintenanceDiskHealth      `json:"disk"`
	CBT       MaintenanceCBTHealth       `json:"cbt"`
	Checklist []MaintenanceChecklistItem `json:"checklist"`
}

type MaintenanceHealthComponent struct {
	Status     string    `json:"status"`
	ServerTime time.Time `json:"server_time"`
}

type MaintenanceDatabaseHealth struct {
	Connected bool       `json:"connected"`
	DBTime    *time.Time `json:"db_time,omitempty"`
	Message   string     `json:"message,omitempty"`
}

type MaintenanceBackupHealth struct {
	Available      bool       `json:"available"`
	Health         string     `json:"health"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	LastRunSuccess *bool      `json:"last_run_success,omitempty"`
	LatestBackupAt *time.Time `json:"latest_backup_at,omitempty"`
	BackupCount    int        `json:"backup_count"`
	Warnings       []string   `json:"warnings"`
}

type MaintenanceDiskHealth struct {
	Health             string   `json:"health"`
	BackupDirSizeBytes int64    `json:"backup_dir_size_bytes"`
	Warnings           []string `json:"warnings"`
}

type MaintenanceCBTHealth struct {
	ActiveSessionCount int64  `json:"active_session_count"`
	Health             string `json:"health"`
	Message            string `json:"message"`
}

type MaintenanceChecklistItem struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	OK       bool   `json:"ok"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func NewSystemMaintenance(q *db.Queries, backup maintenanceBackupStatusProvider) *SystemMaintenance {
	return &SystemMaintenance{q: q, backup: backup, now: time.Now}
}

func NewSystemMaintenanceWithStore(q maintenanceStore, backup maintenanceBackupStatusProvider) *SystemMaintenance {
	return &SystemMaintenance{q: q, backup: backup, now: time.Now}
}

func (s *SystemMaintenance) Status(ctx context.Context) (SystemMaintenanceStatus, error) {
	now := s.currentTime()
	status := SystemMaintenanceStatus{Active: false, Mode: MaintenanceModeOff, ServerTime: now}
	row, err := s.q.GetActiveMaintenanceWindow(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return status, nil
	}
	if err != nil {
		return status, err
	}
	view := s.windowView(row, now)
	status.Active = true
	status.Mode = view.Mode
	status.Window = &view
	return status, nil
}

func (s *SystemMaintenance) ListWindows(ctx context.Context, limit, offset int32) ([]MaintenanceWindowView, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.q.ListMaintenanceWindows(ctx, db.ListMaintenanceWindowsParams{
		LimitCount:  limit,
		OffsetCount: offset,
	})
	if err != nil {
		return nil, err
	}
	now := s.currentTime()
	items := make([]MaintenanceWindowView, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.windowView(row, now))
	}
	return items, nil
}

func (s *SystemMaintenance) CreateWindow(ctx context.Context, in MaintenanceWindowInput) (MaintenanceWindowView, error) {
	normalized, err := s.validateWindowInput(in, false)
	if err != nil {
		return MaintenanceWindowView{}, err
	}
	row, err := s.q.CreateMaintenanceWindow(ctx, db.CreateMaintenanceWindowParams{
		Title:            normalized.Title,
		Message:          normalized.Message,
		Mode:             normalized.Mode,
		AffectedModules:  normalized.AffectedModules,
		StartsAt:         timestamptzFromPtr(normalized.StartsAt),
		EndsAt:           timestamptzFromPtr(normalized.EndsAt),
		IsActive:         normalized.IsActive,
		AllowAdminBypass: derefBool(normalized.AllowAdminBypass, true),
		BypassRoles:      normalized.BypassRoles,
		Severity:         normalized.Severity,
		ActorUserID:      normalized.ActorUserID,
	})
	if err != nil {
		return MaintenanceWindowView{}, err
	}
	if err := s.audit(ctx, row.ID, normalized.ActorUserID, maintenanceActionCreate, normalized.Reason, row); err != nil {
		return MaintenanceWindowView{}, err
	}
	return s.windowView(row, s.currentTime()), nil
}

func (s *SystemMaintenance) UpdateWindow(ctx context.Context, id pgtype.UUID, in MaintenanceWindowInput) (MaintenanceWindowView, error) {
	if !id.Valid {
		return MaintenanceWindowView{}, fmt.Errorf("%w: id maintenance tidak valid", domain.ErrBadRequest)
	}
	normalized, err := s.validateWindowInput(in, true)
	if err != nil {
		return MaintenanceWindowView{}, err
	}
	row, err := s.q.UpdateMaintenanceWindow(ctx, db.UpdateMaintenanceWindowParams{
		ID:               id,
		Title:            normalized.Title,
		Message:          normalized.Message,
		Mode:             normalized.Mode,
		AffectedModules:  normalized.AffectedModules,
		StartsAt:         timestamptzFromPtr(normalized.StartsAt),
		EndsAt:           timestamptzFromPtr(normalized.EndsAt),
		AllowAdminBypass: derefBool(normalized.AllowAdminBypass, true),
		BypassRoles:      normalized.BypassRoles,
		Severity:         normalized.Severity,
		ActorUserID:      normalized.ActorUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MaintenanceWindowView{}, domain.ErrNotFound
	}
	if err != nil {
		return MaintenanceWindowView{}, err
	}
	if err := s.audit(ctx, row.ID, normalized.ActorUserID, maintenanceActionUpdate, normalized.Reason, row); err != nil {
		return MaintenanceWindowView{}, err
	}
	return s.windowView(row, s.currentTime()), nil
}

func (s *SystemMaintenance) ActivateWindow(ctx context.Context, id, actor pgtype.UUID, reason string) (MaintenanceWindowView, error) {
	if !id.Valid {
		return MaintenanceWindowView{}, fmt.Errorf("%w: id maintenance tidak valid", domain.ErrBadRequest)
	}
	row, err := s.q.GetMaintenanceWindow(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return MaintenanceWindowView{}, domain.ErrNotFound
	}
	if err != nil {
		return MaintenanceWindowView{}, err
	}
	if ended(row.EndsAt, s.currentTime()) {
		return MaintenanceWindowView{}, fmt.Errorf("%w: jendela maintenance sudah berakhir", domain.ErrConflict)
	}
	row, err = s.q.SetMaintenanceWindowActive(ctx, db.SetMaintenanceWindowActiveParams{
		ID:          id,
		IsActive:    true,
		ActorUserID: actor,
	})
	if err != nil {
		return MaintenanceWindowView{}, err
	}
	if err := s.audit(ctx, row.ID, actor, maintenanceActionActivate, reason, row); err != nil {
		return MaintenanceWindowView{}, err
	}
	slog.Info("maintenance window activated", "maintenance_id", maintenanceUUIDString(row.ID), "mode", row.Mode)
	return s.windowView(row, s.currentTime()), nil
}

func (s *SystemMaintenance) DeactivateWindow(ctx context.Context, id, actor pgtype.UUID, reason string) (MaintenanceWindowView, error) {
	if !id.Valid {
		return MaintenanceWindowView{}, fmt.Errorf("%w: id maintenance tidak valid", domain.ErrBadRequest)
	}
	row, err := s.q.SetMaintenanceWindowActive(ctx, db.SetMaintenanceWindowActiveParams{
		ID:          id,
		IsActive:    false,
		ActorUserID: actor,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MaintenanceWindowView{}, domain.ErrNotFound
	}
	if err != nil {
		return MaintenanceWindowView{}, err
	}
	if err := s.audit(ctx, row.ID, actor, maintenanceActionDeactivate, reason, row); err != nil {
		return MaintenanceWindowView{}, err
	}
	slog.Info("maintenance window deactivated", "maintenance_id", maintenanceUUIDString(row.ID), "mode", row.Mode)
	return s.windowView(row, s.currentTime()), nil
}

func (s *SystemMaintenance) ListAuditLogs(ctx context.Context, filter MaintenanceAuditFilter) ([]MaintenanceAuditLogView, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	action := strings.TrimSpace(strings.ToLower(filter.Action))
	rows, err := s.q.ListMaintenanceAuditLogs(ctx, db.ListMaintenanceAuditLogsParams{
		ActionFilter: action,
		FromAt:       timestamptzFromPtr(filter.FromAt),
		ToAt:         timestamptzFromPtr(filter.ToAt),
		LimitCount:   filter.Limit,
		OffsetCount:  filter.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]MaintenanceAuditLogView, 0, len(rows))
	for _, row := range rows {
		items = append(items, auditLogView(row))
	}
	return items, nil
}

func (s *SystemMaintenance) HealthSummary(ctx context.Context) (MaintenanceHealthSummary, error) {
	now := s.currentTime()
	summary := MaintenanceHealthSummary{
		CoreAPI: MaintenanceHealthComponent{Status: "ok", ServerTime: now},
		Database: MaintenanceDatabaseHealth{
			Connected: false,
			Message:   "Belum terverifikasi",
		},
		Backup: MaintenanceBackupHealth{
			Available: false,
			Health:    "warning",
			Warnings:  []string{"Status backup belum tersedia."},
		},
		Disk: MaintenanceDiskHealth{
			Health:   "warning",
			Warnings: []string{"Status disk bebas belum tersedia dari layanan sistem."},
		},
		CBT: MaintenanceCBTHealth{
			Health:  "ok",
			Message: "Tidak ada sesi CBT aktif.",
		},
	}

	if dbTime, err := s.q.GetMaintenanceDatabaseTime(ctx); err == nil {
		summary.Database.Connected = true
		summary.Database.DBTime = timePtr(dbTime)
		summary.Database.Message = "Koneksi database aktif."
	} else {
		summary.Database.Message = "Koneksi database belum dapat diverifikasi."
	}

	if activeSessions, err := s.q.CountActiveCbtSessionsForMaintenance(ctx); err == nil {
		summary.CBT.ActiveSessionCount = activeSessions
		if activeSessions > 0 {
			summary.CBT.Health = "warning"
			summary.CBT.Message = fmt.Sprintf("%d sesi CBT sedang aktif. Hindari maintenance global kecuali darurat.", activeSessions)
		}
	} else {
		summary.CBT.Health = "warning"
		summary.CBT.Message = "Jumlah sesi CBT aktif belum dapat diverifikasi."
	}

	if s.backup != nil {
		if backup, err := s.backup.Status(ctx); err == nil {
			summary.Backup = MaintenanceBackupHealth{
				Available:      true,
				Health:         backup.Health,
				LastRunAt:      backup.LastRunAt,
				LastRunSuccess: backup.LastRunSuccess,
				BackupCount:    backup.BackupCount,
				Warnings:       backup.Warnings,
			}
			if backup.LatestBackup != nil {
				latest := backup.LatestBackup.CreatedAt
				summary.Backup.LatestBackupAt = &latest
			}
			summary.Disk = MaintenanceDiskHealth{
				Health:             backup.Health,
				BackupDirSizeBytes: backup.BackupDirSizeBytes,
				Warnings:           backup.Warnings,
			}
		} else {
			summary.Backup.Warnings = []string{"Status backup belum dapat dimuat."}
			summary.Disk.Warnings = []string{"Status disk backup belum dapat dimuat."}
		}
	}

	summary.Checklist = []MaintenanceChecklistItem{
		{
			Key:      "core_api",
			Label:    "Backend merespons",
			OK:       summary.CoreAPI.Status == "ok",
			Severity: "critical",
			Message:  "Core API dapat melayani request Maintenance Center.",
		},
		{
			Key:      "database",
			Label:    "Database terhubung",
			OK:       summary.Database.Connected,
			Severity: "critical",
			Message:  summary.Database.Message,
		},
		{
			Key:      "latest_backup",
			Label:    "Backup terakhir tersedia",
			OK:       summary.Backup.Available && summary.Backup.LatestBackupAt != nil,
			Severity: "warning",
			Message:  maintenanceBackupChecklistMessage(summary.Backup),
		},
		{
			Key:      "disk",
			Label:    "Storage backup tidak error",
			OK:       summary.Disk.Health != "error",
			Severity: "warning",
			Message:  "Pantau ukuran direktori backup sebelum maintenance besar.",
		},
		{
			Key:      "active_cbt",
			Label:    "Tidak ada sesi CBT aktif",
			OK:       summary.CBT.ActiveSessionCount == 0,
			Severity: "critical",
			Message:  summary.CBT.Message,
		},
		{
			Key:      "admin_bypass",
			Label:    "Bypass admin tetap aktif",
			OK:       true,
			Severity: "critical",
			Message:  "Form maintenance memakai bypass admin aktif secara default agar operator tidak terkunci.",
		},
	}

	return summary, nil
}

func (s *SystemMaintenance) validateWindowInput(in MaintenanceWindowInput, updating bool) (MaintenanceWindowInput, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Message = strings.TrimSpace(in.Message)
	in.Mode = strings.TrimSpace(strings.ToLower(in.Mode))
	in.Severity = strings.TrimSpace(strings.ToLower(in.Severity))
	in.Reason = strings.TrimSpace(in.Reason)

	if in.Title == "" {
		return in, fmt.Errorf("%w: judul maintenance wajib diisi", domain.ErrBadRequest)
	}
	if in.Message == "" {
		return in, fmt.Errorf("%w: pesan maintenance wajib diisi", domain.ErrBadRequest)
	}
	switch in.Mode {
	case MaintenanceModeGlobal, MaintenanceModeModule, MaintenanceModeReadOnly:
	default:
		return in, fmt.Errorf("%w: mode maintenance tidak valid", domain.ErrBadRequest)
	}
	switch in.Severity {
	case "":
		in.Severity = MaintenanceSeverityInfo
	case MaintenanceSeverityInfo, MaintenanceSeverityWarning, MaintenanceSeverityCritical:
	default:
		return in, fmt.Errorf("%w: severity maintenance tidak valid", domain.ErrBadRequest)
	}
	if in.StartsAt != nil && in.EndsAt != nil && !in.EndsAt.After(*in.StartsAt) {
		return in, fmt.Errorf("%w: waktu selesai harus setelah waktu mulai", domain.ErrBadRequest)
	}
	if in.IsActive && in.EndsAt != nil && in.EndsAt.Before(s.currentTime()) {
		return in, fmt.Errorf("%w: jendela maintenance sudah berakhir", domain.ErrConflict)
	}
	modules, err := normalizeMaintenanceModules(in.Mode, in.AffectedModules)
	if err != nil {
		return in, err
	}
	in.AffectedModules = modules
	in.BypassRoles = normalizeBypassRoles(in.BypassRoles)
	if in.AllowAdminBypass == nil {
		defaultBypass := true
		in.AllowAdminBypass = &defaultBypass
	}
	if updating {
		in.IsActive = false
	}
	return in, nil
}

func normalizeMaintenanceModules(mode string, modules []string) ([]string, error) {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(modules))
	for _, module := range modules {
		module = strings.TrimSpace(strings.ToLower(strings.ReplaceAll(module, "-", "_")))
		if module == "" {
			continue
		}
		if !IsValidMaintenanceModule(module) {
			return nil, fmt.Errorf("%w: modul maintenance tidak valid", domain.ErrBadRequest)
		}
		if _, ok := seen[module]; ok {
			continue
		}
		seen[module] = struct{}{}
		normalized = append(normalized, module)
	}
	if len(normalized) == 0 {
		if mode == MaintenanceModeModule {
			return nil, fmt.Errorf("%w: pilih minimal satu modul", domain.ErrBadRequest)
		}
		return []string{MaintenanceModuleGlobal}, nil
	}
	return normalized, nil
}

func normalizeBypassRoles(roles []string) []string {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(roles)+2)
	for _, role := range roles {
		role = strings.TrimSpace(strings.ToLower(role))
		if role == "" {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		normalized = append(normalized, role)
	}
	if len(normalized) == 0 {
		return []string{"superadmin", "admin"}
	}
	return normalized
}

func (s *SystemMaintenance) audit(ctx context.Context, maintenanceID, actor pgtype.UUID, action, reason string, row db.SystemMaintenanceWindow) error {
	metadata, err := json.Marshal(map[string]any{
		"maintenance_id": maintenanceUUIDString(row.ID),
		"title":          row.Title,
		"mode":           row.Mode,
		"severity":       row.Severity,
		"is_active":      row.IsActive,
		"modules":        row.AffectedModules,
	})
	if err != nil {
		return err
	}
	_, err = s.q.CreateMaintenanceAuditLog(ctx, db.CreateMaintenanceAuditLogParams{
		MaintenanceID: maintenanceID,
		ActorUserID:   actor,
		Action:        action,
		Reason:        textFromString(reason),
		Metadata:      metadata,
	})
	return err
}

func (s *SystemMaintenance) windowView(row db.SystemMaintenanceWindow, now time.Time) MaintenanceWindowView {
	return MaintenanceWindowView{
		ID:               maintenanceUUIDString(row.ID),
		Title:            row.Title,
		Message:          row.Message,
		Mode:             row.Mode,
		AffectedModules:  append([]string{}, row.AffectedModules...),
		StartsAt:         timePtr(row.StartsAt),
		EndsAt:           timePtr(row.EndsAt),
		IsActive:         row.IsActive,
		AllowAdminBypass: row.AllowAdminBypass,
		BypassRoles:      append([]string{}, row.BypassRoles...),
		Severity:         row.Severity,
		Status:           maintenanceWindowLifecycleStatus(row, now),
		CreatedBy:        maintenanceUUIDString(row.CreatedBy),
		UpdatedBy:        maintenanceUUIDString(row.UpdatedBy),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func maintenanceWindowLifecycleStatus(row db.SystemMaintenanceWindow, now time.Time) string {
	if !row.IsActive {
		return "inactive"
	}
	if row.EndsAt.Valid && row.EndsAt.Time.Before(now) {
		return "ended"
	}
	if row.StartsAt.Valid && row.StartsAt.Time.After(now) {
		return "scheduled"
	}
	return "active_now"
}

func auditLogView(row db.SystemMaintenanceAuditLog) MaintenanceAuditLogView {
	metadata := map[string]any{}
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &metadata)
	}
	return MaintenanceAuditLogView{
		ID:            maintenanceUUIDString(row.ID),
		MaintenanceID: maintenanceUUIDString(row.MaintenanceID),
		ActorUserID:   maintenanceUUIDString(row.ActorUserID),
		Action:        row.Action,
		Reason:        row.Reason.String,
		Metadata:      redactMaintenanceMetadata(metadata),
		CreatedAt:     row.CreatedAt.Time,
	}
}

func maintenanceBackupChecklistMessage(backup MaintenanceBackupHealth) string {
	if !backup.Available {
		return "Status backup belum tersedia dari Backup Center."
	}
	if backup.LatestBackupAt == nil {
		return "Backup terakhir belum ditemukan."
	}
	if backup.LastRunSuccess != nil && !*backup.LastRunSuccess {
		return "Backup terakhir tercatat gagal. Jalankan backup manual sebelum maintenance besar."
	}
	return "Backup terakhir tersedia untuk preflight maintenance."
}

func ended(value pgtype.Timestamptz, now time.Time) bool {
	return value.Valid && value.Time.Before(now)
}

func (s *SystemMaintenance) currentTime() time.Time {
	if s.now == nil {
		return time.Now()
	}
	return s.now()
}

func derefBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func timestamptzFromPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func textFromString(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func maintenanceUUIDString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	b := value.Bytes
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3],
		b[4], b[5],
		b[6], b[7],
		b[8], b[9],
		b[10], b[11], b[12], b[13], b[14], b[15],
	)
}

func redactMaintenanceMetadata(metadata map[string]any) map[string]any {
	redacted := make(map[string]any, len(metadata))
	for key, value := range metadata {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "credential") {
			redacted[key] = "[REDACTED]"
			continue
		}
		if nested, ok := value.(map[string]any); ok {
			redacted[key] = redactMaintenanceMetadata(nested)
			continue
		}
		redacted[key] = value
	}
	return redacted
}
