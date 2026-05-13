package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"mtsn2kolut-super-app/backend/internal/domain"
)

const (
	DefaultSystemBackupDir           = "/home/servermtsn2kolut/backups/mtsn2kolut-super-app/postgresql"
	DefaultSystemBackupTimerName     = "mtsn2kolut-postgresql-backup.timer"
	DefaultSystemBackupServiceName   = "mtsn2kolut-postgresql-backup.service"
	DefaultSystemBackupRetentionDays = 30
	DefaultSystemBackupScriptPath    = "/home/servermtsn2kolut/mtsn2kolut-super-app/deploy/backup-postgresql.sh"

	systemBackupSchedule          = "*-*-* 00:00:00"
	systemBackupTimezone          = "WITA"
	systemBackupStatusTimeout     = 2 * time.Second
	systemBackupRunTimeout        = 10 * time.Minute
	systemBackupStaleAfter        = 24 * time.Hour
	systemBackupLatestSymlinkName = "latest.dump"
)

type SystemBackupCommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

type SystemBackupConfig struct {
	BackupDir     string
	ScriptPath    string
	TimerName     string
	ServiceName   string
	RetentionDays int
	RunTimeout    time.Duration
	Now           func() time.Time
	CommandRunner SystemBackupCommandRunner
}

type SystemBackup struct {
	cfg     SystemBackupConfig
	mu      sync.Mutex
	running bool
	jobs    map[string]SystemBackupJob
}

type SystemBackupFile struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Kind            string    `json:"kind"`
	SizeBytes       int64     `json:"size_bytes"`
	CreatedAt       time.Time `json:"created_at"`
	SHA256          string    `json:"sha256,omitempty"`
	SHA256Available bool      `json:"sha256_available"`
	IsLatest        bool      `json:"is_latest"`
	Downloadable    bool      `json:"downloadable"`
}

type SystemBackupListMeta struct {
	Total int `json:"total"`
}

type SystemBackupList struct {
	Items []SystemBackupFile   `json:"items"`
	Meta  SystemBackupListMeta `json:"meta"`
}

type SystemBackupStatus struct {
	TimerName          string            `json:"timer_name"`
	ServiceName        string            `json:"service_name"`
	TimerEnabled       bool              `json:"timer_enabled"`
	TimerActive        bool              `json:"timer_active"`
	Schedule           string            `json:"schedule"`
	Timezone           string            `json:"timezone"`
	LastRunAt          *time.Time        `json:"last_run_at,omitempty"`
	LastRunSuccess     *bool             `json:"last_run_success,omitempty"`
	NextRunAt          *time.Time        `json:"next_run_at,omitempty"`
	LatestBackup       *SystemBackupFile `json:"latest_backup,omitempty"`
	RetentionDays      int               `json:"retention_days"`
	BackupCount        int               `json:"backup_count"`
	BackupDirSizeBytes int64             `json:"backup_dir_size_bytes"`
	Health             string            `json:"health"`
	Warnings           []string          `json:"warnings"`
}

type SystemBackupDownload struct {
	Path      string
	Filename  string
	SizeBytes int64
	ModTime   time.Time
}

type SystemBackupRunRequest struct {
	Reason string `json:"reason"`
}

type SystemBackupJob struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Reason      string     `json:"reason,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Output      string     `json:"output,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type systemdBackupStatus struct {
	TimerEnabled   bool
	TimerActive    bool
	LastRunAt      *time.Time
	LastRunSuccess *bool
	NextRunAt      *time.Time
	Warnings       []string
}

func NewSystemBackup(cfg SystemBackupConfig) *SystemBackup {
	if strings.TrimSpace(cfg.BackupDir) == "" {
		cfg.BackupDir = DefaultSystemBackupDir
	}
	if strings.TrimSpace(cfg.ScriptPath) == "" {
		cfg.ScriptPath = DefaultSystemBackupScriptPath
	}
	if strings.TrimSpace(cfg.TimerName) == "" {
		cfg.TimerName = DefaultSystemBackupTimerName
	}
	if strings.TrimSpace(cfg.ServiceName) == "" {
		cfg.ServiceName = DefaultSystemBackupServiceName
	}
	if cfg.RetentionDays <= 0 {
		cfg.RetentionDays = DefaultSystemBackupRetentionDays
	}
	if cfg.RunTimeout <= 0 {
		cfg.RunTimeout = systemBackupRunTimeout
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.CommandRunner == nil {
		cfg.CommandRunner = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		}
	}
	return &SystemBackup{cfg: cfg, jobs: map[string]SystemBackupJob{}}
}

func (s *SystemBackup) Status(ctx context.Context) (SystemBackupStatus, error) {
	list, err := s.List(ctx)
	if err != nil {
		return SystemBackupStatus{}, err
	}

	status := SystemBackupStatus{
		TimerName:          s.cfg.TimerName,
		ServiceName:        s.cfg.ServiceName,
		Schedule:           systemBackupSchedule,
		Timezone:           systemBackupTimezone,
		RetentionDays:      s.cfg.RetentionDays,
		BackupCount:        len(list.Items),
		BackupDirSizeBytes: backupListSize(list.Items),
		Health:             "ok",
		Warnings:           []string{},
	}
	systemd := s.systemdStatus(ctx)
	status.TimerEnabled = systemd.TimerEnabled
	status.TimerActive = systemd.TimerActive
	status.LastRunAt = systemd.LastRunAt
	status.LastRunSuccess = systemd.LastRunSuccess
	status.NextRunAt = systemd.NextRunAt
	status.Warnings = append(status.Warnings, systemd.Warnings...)

	for _, item := range list.Items {
		if item.IsLatest {
			latest := item
			status.LatestBackup = &latest
			break
		}
	}
	if status.LatestBackup == nil && len(list.Items) > 0 {
		latest := list.Items[0]
		status.LatestBackup = &latest
	}

	if status.LatestBackup == nil {
		status.Warnings = append(status.Warnings, "Belum ada file backup PostgreSQL .dump yang tersedia.")
	} else if s.cfg.Now().Sub(status.LatestBackup.CreatedAt) > systemBackupStaleAfter {
		status.Warnings = append(status.Warnings, "Backup terbaru lebih lama dari 24 jam.")
	}
	if !status.TimerActive {
		status.Warnings = append(status.Warnings, "Timer backup PostgreSQL tidak aktif atau belum dapat diverifikasi.")
	}
	if !status.TimerEnabled {
		status.Warnings = append(status.Warnings, "Timer backup PostgreSQL belum terdeteksi enabled.")
	}
	if status.LastRunSuccess != nil && !*status.LastRunSuccess {
		status.Warnings = append(status.Warnings, "Eksekusi backup terakhir tidak berhasil menurut systemd.")
	}

	status.Warnings = uniqueStrings(status.Warnings)
	status.Health = systemBackupHealth(status)
	return status, nil
}

func (s *SystemBackup) List(ctx context.Context) (SystemBackupList, error) {
	_ = ctx
	dir, err := s.backupDir()
	if err != nil {
		return SystemBackupList{}, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return SystemBackupList{Items: []SystemBackupFile{}, Meta: SystemBackupListMeta{Total: 0}}, nil
		}
		return SystemBackupList{}, err
	}

	latestName := s.latestTargetName(dir)
	items := make([]SystemBackupFile, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".dump") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		path := filepath.Join(dir, name)
		sha := readBackupSHA256(path)
		items = append(items, SystemBackupFile{
			ID:              name,
			Name:            name,
			Kind:            classifyBackupKind(name),
			SizeBytes:       info.Size(),
			CreatedAt:       info.ModTime(),
			SHA256:          sha,
			SHA256Available: sha != "",
			IsLatest:        name == latestName,
			Downloadable:    true,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		return items[i].Name > items[j].Name
	})
	return SystemBackupList{Items: items, Meta: SystemBackupListMeta{Total: len(items)}}, nil
}

func (s *SystemBackup) RunManual(ctx context.Context, req SystemBackupRunRequest) (SystemBackupJob, error) {
	reason := sanitizeBackupReason(req.Reason)
	if len(reason) > 200 {
		return SystemBackupJob{}, domain.ErrBadRequest
	}
	scriptPath, err := cleanBackupScriptPath(s.cfg.ScriptPath)
	if err != nil {
		return SystemBackupJob{}, err
	}
	if _, err := os.Stat(scriptPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return SystemBackupJob{}, domain.ErrNotFound
		}
		return SystemBackupJob{}, err
	}

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return SystemBackupJob{}, domain.ErrConflict
	}
	s.running = true
	job := SystemBackupJob{
		ID:        fmt.Sprintf("backup-%s", s.cfg.Now().Format("20060102-150405")),
		Status:    "running",
		Reason:    reason,
		StartedAt: s.cfg.Now(),
	}
	s.jobs[job.ID] = job
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	runCtx, cancel := context.WithTimeout(ctx, s.cfg.RunTimeout)
	defer cancel()
	output, runErr := s.cfg.CommandRunner(runCtx, "bash", scriptPath)
	completedAt := s.cfg.Now()
	job.CompletedAt = &completedAt
	job.Output = sanitizeBackupCommandOutput(string(output))
	if runErr != nil {
		job.Status = "failed"
		job.Error = sanitizeBackupCommandOutput(runErr.Error())
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			job.Error = "backup manual melewati batas waktu eksekusi"
		}
	} else {
		job.Status = "success"
		if strings.TrimSpace(job.Output) == "" {
			job.Output = "Backup manual selesai. Periksa daftar backup terbaru."
		}
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()
	if runErr != nil {
		return job, runErr
	}
	return job, nil
}

func (s *SystemBackup) Job(ctx context.Context, id string) (SystemBackupJob, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[strings.TrimSpace(id)]
	if !ok {
		return SystemBackupJob{}, domain.ErrNotFound
	}
	return job, nil
}

func (s *SystemBackup) Download(ctx context.Context, id string) (SystemBackupDownload, error) {
	_ = ctx
	name, err := cleanBackupID(id)
	if err != nil {
		return SystemBackupDownload{}, err
	}
	dir, err := s.backupDir()
	if err != nil {
		return SystemBackupDownload{}, err
	}
	path := filepath.Join(dir, name)
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return SystemBackupDownload{}, domain.ErrNotFound
		}
		return SystemBackupDownload{}, err
	}

	filename := name
	if info.Mode()&fs.ModeSymlink != 0 {
		if name != systemBackupLatestSymlinkName {
			return SystemBackupDownload{}, domain.ErrForbidden
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return SystemBackupDownload{}, domain.ErrNotFound
			}
			return SystemBackupDownload{}, err
		}
		if !pathInside(dir, resolved) || strings.ToLower(filepath.Ext(resolved)) != ".dump" {
			return SystemBackupDownload{}, domain.ErrForbidden
		}
		stat, err := os.Stat(resolved)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return SystemBackupDownload{}, domain.ErrNotFound
			}
			return SystemBackupDownload{}, err
		}
		if !stat.Mode().IsRegular() {
			return SystemBackupDownload{}, domain.ErrNotFound
		}
		path = resolved
		info = stat
		filename = filepath.Base(resolved)
	} else if !info.Mode().IsRegular() {
		return SystemBackupDownload{}, domain.ErrNotFound
	}

	return SystemBackupDownload{
		Path:      path,
		Filename:  filename,
		SizeBytes: info.Size(),
		ModTime:   info.ModTime(),
	}, nil
}

func (s *SystemBackup) backupDir() (string, error) {
	return filepath.Abs(filepath.Clean(s.cfg.BackupDir))
}

func (s *SystemBackup) latestTargetName(dir string) string {
	path := filepath.Join(dir, systemBackupLatestSymlinkName)
	info, err := os.Lstat(path)
	if err != nil {
		return ""
	}
	if info.Mode()&fs.ModeSymlink == 0 {
		if info.Mode().IsRegular() {
			return systemBackupLatestSymlinkName
		}
		return ""
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return ""
	}
	if !pathInside(dir, resolved) || strings.ToLower(filepath.Ext(resolved)) != ".dump" {
		return ""
	}
	return filepath.Base(resolved)
}

func (s *SystemBackup) systemdStatus(ctx context.Context) systemdBackupStatus {
	status := systemdBackupStatus{Warnings: []string{}}
	if !validSystemdUnitName(s.cfg.TimerName, ".timer") || !validSystemdUnitName(s.cfg.ServiceName, ".service") {
		status.Warnings = append(status.Warnings, "Nama unit systemd backup tidak valid.")
		return status
	}

	timerFields, err := s.systemctlShow(ctx, s.cfg.TimerName, "UnitFileState", "ActiveState", "NextElapseUSecRealtime", "LastTriggerUSec")
	if err != nil {
		status.Warnings = append(status.Warnings, "Status timer backup PostgreSQL belum dapat dibaca.")
	} else {
		status.TimerEnabled = timerFields["UnitFileState"] == "enabled"
		status.TimerActive = timerFields["ActiveState"] == "active"
		status.NextRunAt = parseSystemdTimestamp(timerFields["NextElapseUSecRealtime"])
		status.LastRunAt = parseSystemdTimestamp(timerFields["LastTriggerUSec"])
	}

	serviceFields, err := s.systemctlShow(ctx, s.cfg.ServiceName, "Result", "ExecMainStatus")
	if err != nil {
		status.Warnings = append(status.Warnings, "Status service backup PostgreSQL belum dapat dibaca.")
		return status
	}
	success := serviceFields["Result"] == "success"
	if serviceFields["Result"] == "" && serviceFields["ExecMainStatus"] == "0" {
		success = true
	}
	if serviceFields["Result"] != "" || serviceFields["ExecMainStatus"] != "" {
		status.LastRunSuccess = &success
	}
	return status
}

func (s *SystemBackup) systemctlShow(ctx context.Context, unit string, properties ...string) (map[string]string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, systemBackupStatusTimeout)
	defer cancel()
	args := []string{"show", unit}
	for _, property := range properties {
		args = append(args, "--property="+property)
	}
	out, err := s.cfg.CommandRunner(timeoutCtx, "systemctl", args...)
	if err != nil {
		return nil, err
	}
	return parseSystemctlShow(string(out)), nil
}

func parseSystemctlShow(raw string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return fields
}

func parseSystemdTimestamp(raw string) *time.Time {
	value := strings.TrimSpace(raw)
	if value == "" || value == "n/a" || value == "0" {
		return nil
	}
	parts := strings.Fields(value)
	if len(parts) >= 4 && isWeekday(parts[0]) {
		value = strings.Join(parts[1:], " ")
	}
	if strings.HasSuffix(value, " WITA") {
		trimmed := strings.TrimSuffix(value, " WITA")
		if parsed, err := time.ParseInLocation("2006-01-02 15:04:05", trimmed, time.FixedZone("WITA", 8*60*60)); err == nil {
			return &parsed
		}
	}
	layouts := []string{
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func isWeekday(value string) bool {
	switch value {
	case "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun":
		return true
	default:
		return false
	}
}

func cleanBackupID(id string) (string, error) {
	name := strings.TrimSpace(id)
	if name == "" || name == "." || name == ".." {
		return "", domain.ErrBadRequest
	}
	if strings.ContainsAny(name, `/\`+"\x00") || filepath.Base(name) != name {
		return "", domain.ErrForbidden
	}
	if strings.ToLower(filepath.Ext(name)) != ".dump" {
		return "", domain.ErrForbidden
	}
	return name, nil
}

func pathInside(baseDir, candidate string) bool {
	base, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		base, err = filepath.Abs(baseDir)
		if err != nil {
			return false
		}
	}
	target, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel == "." || (!filepath.IsAbs(rel) && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func classifyBackupKind(name string) string {
	switch {
	case strings.HasPrefix(name, "pusaka_"):
		return "scheduled"
	case strings.HasPrefix(name, "pre-"):
		return "manual-pre-change"
	default:
		return "manual"
	}
}

func readBackupSHA256(path string) string {
	data, err := os.ReadFile(path + ".sha256")
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return ""
	}
	candidate := strings.ToLower(strings.TrimSpace(fields[0]))
	if !isSHA256Hex(candidate) {
		return ""
	}
	return candidate
}

func isSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func validSystemdUnitName(name, suffix string) bool {
	value := strings.TrimSpace(name)
	if value == "" || strings.HasPrefix(value, "-") || !strings.HasSuffix(value, suffix) {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '.', '_', '-', '@':
			continue
		default:
			return false
		}
	}
	return true
}

func sanitizeBackupReason(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\x00", "")
	return strings.Join(strings.Fields(value), " ")
}

func cleanBackupScriptPath(path string) (string, error) {
	value := filepath.Clean(strings.TrimSpace(path))
	if value == "" || !filepath.IsAbs(value) || strings.Contains(value, "\x00") {
		return "", domain.ErrBadRequest
	}
	return value, nil
}

func sanitizeBackupCommandOutput(value string) string {
	value = strings.ReplaceAll(value, "\x00", "")
	lines := strings.Split(value, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = redactBackupSecretLine(strings.TrimSpace(line))
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}
	if len(cleaned) > 20 {
		cleaned = cleaned[len(cleaned)-20:]
	}
	out := strings.Join(cleaned, "\n")
	if len(out) > 4000 {
		out = out[len(out)-4000:]
	}
	return out
}

func redactBackupSecretLine(line string) string {
	lower := strings.ToLower(line)
	secretMarkers := []string{"password", "database_url", "postgres_password", "pgpassword", "token", "secret"}
	for _, marker := range secretMarkers {
		if strings.Contains(lower, marker) {
			return "[REDACTED]"
		}
	}
	return line
}

func backupListSize(items []SystemBackupFile) int64 {
	var total int64
	for _, item := range items {
		total += item.SizeBytes
	}
	return total
}

func systemBackupHealth(status SystemBackupStatus) string {
	if !status.TimerActive || (status.LastRunSuccess != nil && !*status.LastRunSuccess) {
		return "error"
	}
	if len(status.Warnings) > 0 {
		return "warning"
	}
	return "ok"
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
