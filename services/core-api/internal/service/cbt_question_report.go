package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const bankSoalReportTZ = "Asia/Makassar"

type BankSoalReportService struct {
	pool *pgxpool.Pool
}

func NewBankSoalReportService(pool *pgxpool.Pool) *BankSoalReportService {
	return &BankSoalReportService{pool: pool}
}

type BankSoalReportFilters struct {
	Report           string   `json:"report"`
	PeriodPreset     string   `json:"period_preset"`
	StartDate        string   `json:"start_date,omitempty"`
	EndDate          string   `json:"end_date,omitempty"`
	EventID          string   `json:"event_id,omitempty"`
	SubjectID        string   `json:"subject_id,omitempty"`
	TargetLevel      string   `json:"target_level,omitempty"`
	AuthorUsername   string   `json:"author_username,omitempty"`
	WorkflowStatus   string   `json:"workflow_status,omitempty"`
	WorkflowStatuses []string `json:"workflow_statuses,omitempty"`
	IncludeSystem    bool     `json:"include_system"`
	GroupBy          string   `json:"group_by,omitempty"`
}

type BankSoalReportRow struct {
	No          int    `json:"no"`
	Primary     string `json:"primary"`
	Secondary   string `json:"secondary,omitempty"`
	SubjectName string `json:"subject_name,omitempty"`
	LevelName   string `json:"level_name,omitempty"`
	Task        string `json:"task,omitempty"`
	Total       int64  `json:"total"`
	PG          int64  `json:"pg"`
	Essay       int64  `json:"essay"`
	Other       int64  `json:"other"`
	Draft       int64  `json:"draft"`
	Submitted   int64  `json:"submitted"`
	Review      int64  `json:"review"`
	Revision    int64  `json:"revision_needed"`
	Approved    int64  `json:"approved"`
	Published   int64  `json:"published"`
	Rejected    int64  `json:"rejected"`
	TargetPG    int64  `json:"target_pg,omitempty"`
	TargetEssay int64  `json:"target_essay,omitempty"`
	Shortage    int64  `json:"shortage,omitempty"`
	Volume      int64  `json:"volume,omitempty"`
	Unit        string `json:"unit,omitempty"`
	Statuses    string `json:"statuses,omitempty"`
	FirstInput  string `json:"first_input,omitempty"`
	LastInput   string `json:"last_input,omitempty"`
	Notes       string `json:"notes,omitempty"`
	SystemRow   bool   `json:"system_row"`
	DataWarning bool   `json:"data_warning"`
}

type BankSoalReportStatus struct {
	Status string `json:"status"`
	Total  int64  `json:"total"`
}

type BankSoalReportSummary struct {
	Total        int64 `json:"total"`
	PG           int64 `json:"pg"`
	Essay        int64 `json:"essay"`
	Other        int64 `json:"other"`
	Authors      int64 `json:"authors"`
	Subjects     int64 `json:"subjects"`
	MissingLevel int64 `json:"missing_level"`
	SystemRows   int64 `json:"system_rows"`
	Shortage     int64 `json:"shortage"`
}

type BankSoalReportResult struct {
	Filters     BankSoalReportFilters  `json:"filters"`
	GeneratedAt string                 `json:"generated_at"`
	PeriodLabel string                 `json:"period_label"`
	Title       string                 `json:"title"`
	Summary     BankSoalReportSummary  `json:"summary"`
	Statuses    []BankSoalReportStatus `json:"statuses"`
	Rows        []BankSoalReportRow    `json:"rows"`
	AccessNote  string                 `json:"access_note"`
}

type BankSoalReportExport struct {
	FileName    string
	ContentType string
	Data        []byte
}

func (s *BankSoalReportService) Report(ctx context.Context, filters BankSoalReportFilters, actor CbtQuestionActor) (BankSoalReportResult, error) {
	filters = normalizeBankSoalReportFilters(filters)
	startAt, endAt, periodLabel, err := bankSoalReportRange(filters)
	if err != nil {
		return BankSoalReportResult{}, err
	}
	allAccess := bankSoalReportCanReadAll(actor)
	if !allAccess {
		filters.AuthorUsername = strings.TrimSpace(actor.Username)
	}

	rows, err := s.queryRows(ctx, filters, startAt, endAt)
	if err != nil {
		return BankSoalReportResult{}, err
	}
	result := BankSoalReportResult{
		Filters:     filters,
		GeneratedAt: time.Now().In(bankSoalReportLocation()).Format(time.RFC3339),
		PeriodLabel: periodLabel,
		Title:       bankSoalReportTitle(filters.Report),
		Rows:        rows,
		AccessNote:  bankSoalReportAccessNote(allAccess),
	}
	result.Summary = summarizeBankSoalRows(rows)
	result.Statuses = statusesFromRows(rows)
	return result, nil
}

func (s *BankSoalReportService) Export(ctx context.Context, filters BankSoalReportFilters, format string, actor CbtQuestionActor) (BankSoalReportExport, error) {
	result, err := s.Report(ctx, filters, actor)
	if err != nil {
		return BankSoalReportExport{}, err
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "png"
	}
	stamp := time.Now().In(bankSoalReportLocation()).Format("20060102-150405")
	base := fmt.Sprintf("laporan-bank-soal-%s-%s", result.Filters.Report, stamp)
	switch format {
	case "csv", "xlsx", "excel":
		data, err := renderBankSoalReportCSV(result)
		if err != nil {
			return BankSoalReportExport{}, err
		}
		return BankSoalReportExport{FileName: base + ".csv", ContentType: "text/csv; charset=utf-8", Data: data}, nil
	case "html":
		return BankSoalReportExport{FileName: base + ".html", ContentType: "text/html; charset=utf-8", Data: []byte(renderBankSoalReportHTML(result))}, nil
	case "pdf":
		data, err := renderBankSoalReportPDF(result)
		if err != nil {
			return BankSoalReportExport{}, err
		}
		return BankSoalReportExport{FileName: base + ".pdf", ContentType: "application/pdf", Data: data}, nil
	case "png", "image":
		data, err := renderBankSoalReportPNG(result)
		if err != nil {
			return BankSoalReportExport{}, err
		}
		return BankSoalReportExport{FileName: base + ".png", ContentType: "image/png", Data: data}, nil
	default:
		return BankSoalReportExport{}, fmt.Errorf("format export tidak didukung")
	}
}

func normalizeBankSoalReportFilters(f BankSoalReportFilters) BankSoalReportFilters {
	f.Report = strings.ToLower(strings.TrimSpace(f.Report))
	if f.Report == "" {
		f.Report = "input"
	}
	allowed := map[string]bool{"input": true, "progress": true, "revision": true, "reviewer": true, "readiness": true, "honor": true}
	if !allowed[f.Report] {
		f.Report = "input"
	}
	f.PeriodPreset = strings.ToLower(strings.TrimSpace(f.PeriodPreset))
	if f.PeriodPreset == "" {
		f.PeriodPreset = "this_month"
	}
	f.TargetLevel = strings.TrimSpace(f.TargetLevel)
	f.AuthorUsername = strings.TrimSpace(f.AuthorUsername)
	f.WorkflowStatus = strings.TrimSpace(f.WorkflowStatus)
	f.WorkflowStatuses = normalizeBankSoalWorkflowStatuses(f.WorkflowStatus, f.WorkflowStatuses)
	f.WorkflowStatus = strings.Join(f.WorkflowStatuses, ",")
	f.SubjectID = strings.TrimSpace(f.SubjectID)
	f.EventID = strings.TrimSpace(f.EventID)
	return f
}

func normalizeBankSoalWorkflowStatuses(primary string, values []string) []string {
	allowed := map[string]bool{
		"konsep":     true,
		"diperiksa":  true,
		"siap_pakai": true,
	}
	seen := map[string]bool{}
	out := []string{}
	add := func(raw string) {
		for _, part := range strings.Split(raw, ",") {
			status := normalizeWorkflowStatus(part)
			if status == "" || !allowed[status] || seen[status] {
				continue
			}
			seen[status] = true
			out = append(out, status)
		}
	}
	add(primary)
	for _, value := range values {
		add(value)
	}
	return out
}

func bankSoalReportRange(f BankSoalReportFilters) (time.Time, time.Time, string, error) {
	loc := bankSoalReportLocation()
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	switch f.PeriodPreset {
	case "all", "semua":
		start := time.Date(2000, 1, 1, 0, 0, 0, 0, loc)
		return start, now.AddDate(1, 0, 0), "Semua periode", nil
	case "today":
		return startOfDay, startOfDay.AddDate(0, 0, 1), startOfDay.Format("02 Jan 2006") + " WITA", nil
	case "this_week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := startOfDay.AddDate(0, 0, -(weekday - 1))
		return start, start.AddDate(0, 0, 7), start.Format("02 Jan 2006") + " s.d. " + start.AddDate(0, 0, 6).Format("02 Jan 2006") + " WITA", nil
	case "custom":
		if f.StartDate == "" || f.EndDate == "" {
			return time.Time{}, time.Time{}, "", fmt.Errorf("tanggal awal dan akhir wajib diisi")
		}
		start, err := time.ParseInLocation("2006-01-02", f.StartDate, loc)
		if err != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("tanggal awal tidak valid")
		}
		end, err := time.ParseInLocation("2006-01-02", f.EndDate, loc)
		if err != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("tanggal akhir tidak valid")
		}
		end = end.AddDate(0, 0, 1)
		if !end.After(start) {
			return time.Time{}, time.Time{}, "", fmt.Errorf("rentang tanggal tidak valid")
		}
		if end.Sub(start) > 370*24*time.Hour {
			return time.Time{}, time.Time{}, "", fmt.Errorf("rentang laporan maksimal 1 tahun")
		}
		return start, end, start.Format("02 Jan 2006") + " s.d. " + end.AddDate(0, 0, -1).Format("02 Jan 2006") + " WITA", nil
	default:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 1, 0), start.Format("02 Jan 2006") + " s.d. " + now.Format("02 Jan 2006 15:04") + " WITA", nil
	}
}

func bankSoalReportLocation() *time.Location {
	loc, err := time.LoadLocation(bankSoalReportTZ)
	if err != nil {
		return time.FixedZone("WITA", 8*3600)
	}
	return loc
}
func bankSoalReportCanReadAll(a CbtQuestionActor) bool {
	return a.IsAdmin() || a.CanReadAllBankSoal() || a.HasPermission("bank_soal.analytics") || a.HasPermission("bank_soal.review") || a.HasPermission("bank_soal.approve") || a.HasPermission("bank_soal.publish") || a.HasPermission("bank_soal.settings")
}
func bankSoalReportAccessNote(all bool) string {
	if all {
		return "Anda melihat semua data karena memiliki izin laporan Bank Soal."
	}
	return "Anda melihat data soal milik sendiri."
}
func bankSoalReportTitle(report string) string {
	titles := map[string]string{"input": "Laporan Input Soal", "progress": "Progres Mapel", "revision": "Soal Perlu Revisi", "reviewer": "Reviewer & Verifikasi", "readiness": "Kesiapan Paket", "honor": "Honor/Tugas"}
	if v := titles[report]; v != "" {
		return v
	}
	return "Laporan Bank Soal"
}

const bankSoalBaseFilterSQL = `
WITH filtered AS (
  SELECT q.id, q.author_username, q.reviewer_username, q.approver_username, q.subject_id::text AS subject_id,
         COALESCE(NULLIF(s.name,''), '[Mapel tidak tercatat]') AS subject_name,
         COALESCE(NULLIF(q.target_level,''), 'Belum tercatat') AS level_name,
         COALESCE(NULLIF(q.question_type::text,''), 'multiple_choice') AS qtype,
         COALESCE(NULLIF(q.workflow_status::text,''), 'konsep') AS workflow,
         q.status::text AS publication_status,
         (q.created_at AT TIME ZONE 'Asia/Makassar') AS created_wita,
         COALESCE(NULLIF(e.nama,''), NULLIF(u.display_name,''), NULLIF(q.author_username,''), '[Tidak tercatat]') AS author_name
  FROM cbt_questions q
  LEFT JOIN subjects s ON s.id = q.subject_id
  LEFT JOIN users u ON u.username = q.author_username
  LEFT JOIN employees e ON e.id = u.employee_id
  WHERE q.created_at >= $1 AND q.created_at < $2
    AND ($3::boolean OR COALESCE(q.author_username,'') <> 'system')
    AND ($4::text = '' OR q.subject_id::text = $4)
    AND ($5::text = '' OR q.target_level = $5)
    AND ($6::text = '' OR q.author_username = $6)
    AND (cardinality($7::text[]) = 0 OR q.workflow_status::text = ANY($7::text[]))
    AND ($8::text = '' OR COALESCE(q.event_id::text,'') = $8)
)
`

func (s *BankSoalReportService) queryRows(ctx context.Context, f BankSoalReportFilters, startAt, endAt time.Time) ([]BankSoalReportRow, error) {
	selectSQL := bankSoalReportSelectSQL(f.Report)
	query := bankSoalBaseFilterSQL + selectSQL
	rows, err := s.pool.Query(ctx, query, startAt, endAt, f.IncludeSystem, f.SubjectID, f.TargetLevel, f.AuthorUsername, f.WorkflowStatuses, f.EventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BankSoalReportRow{}
	for rows.Next() {
		var r BankSoalReportRow
		if err := rows.Scan(&r.Primary, &r.Secondary, &r.SubjectName, &r.LevelName, &r.Task, &r.Total, &r.PG, &r.Essay, &r.Other, &r.Draft, &r.Submitted, &r.Review, &r.Revision, &r.Approved, &r.Published, &r.Rejected, &r.TargetPG, &r.TargetEssay, &r.Shortage, &r.Volume, &r.Unit, &r.FirstInput, &r.LastInput, &r.Notes); err != nil {
			return nil, err
		}
		r.No = len(out) + 1
		r.SystemRow = r.Secondary == "system" || r.Primary == "system"
		r.DataWarning = strings.Contains(r.LevelName, "Belum") || strings.Contains(r.SubjectName, "tidak tercatat")
		r.Statuses = statusString(r)
		out = append(out, r)
	}
	return out, rows.Err()
}

func bankSoalReportSelectSQL(report string) string {
	commonCounts := `COUNT(*)::bigint AS total, COUNT(*) FILTER (WHERE qtype='multiple_choice')::bigint AS pg, COUNT(*) FILTER (WHERE qtype='essay')::bigint AS essay, COUNT(*) FILTER (WHERE qtype NOT IN ('multiple_choice','essay'))::bigint AS other, COUNT(*) FILTER (WHERE workflow='konsep')::bigint AS draft, 0::bigint AS submitted, COUNT(*) FILTER (WHERE workflow='diperiksa')::bigint AS review, COUNT(*) FILTER (WHERE workflow='konsep' AND NULLIF(btrim(COALESCE(reviewer_username,'')), '') IS NOT NULL) AS revision, COUNT(*) FILTER (WHERE workflow='siap_pakai')::bigint AS approved, COUNT(*) FILTER (WHERE publication_status='published')::bigint AS published, 0::bigint AS rejected`
	suffix := `, to_char(MIN(created_wita), 'YYYY-MM-DD HH24:MI') AS first_input, to_char(MAX(created_wita), 'YYYY-MM-DD HH24:MI') AS last_input`
	switch report {
	case "progress":
		return ` SELECT subject_name AS primary, '' AS secondary, subject_name, level_name, 'Progres mapel' AS task, ` + commonCounts + `, 20::bigint AS target_pg, 5::bigint AS target_essay, GREATEST(20 - COUNT(*) FILTER (WHERE qtype='multiple_choice'),0)::bigint + GREATEST(5 - COUNT(*) FILTER (WHERE qtype='essay'),0)::bigint AS shortage, COUNT(*)::bigint AS volume, 'soal' AS unit` + suffix + `, CASE WHEN (COUNT(*) FILTER (WHERE qtype='multiple_choice') >= 20 AND COUNT(*) FILTER (WHERE qtype='essay') >= 5) THEN 'Lengkap' ELSE 'Belum lengkap' END AS notes FROM filtered GROUP BY subject_name, level_name ORDER BY subject_name, level_name`
	case "revision":
		return ` SELECT author_name AS primary, COALESCE(author_username,'-') AS secondary, subject_name, level_name, 'Revisi soal' AS task, ` + commonCounts + `, 0::bigint, 0::bigint, COUNT(*)::bigint AS shortage, COUNT(*)::bigint AS volume, 'soal revisi' AS unit` + suffix + `, 'Perlu ditindaklanjuti pembuat soal' AS notes FROM filtered WHERE workflow='konsep' AND NULLIF(btrim(COALESCE(reviewer_username,'')), '') IS NOT NULL GROUP BY author_name, author_username, subject_name, level_name ORDER BY last_input DESC`
	case "reviewer":
		return ` SELECT COALESCE(NULLIF(reviewer_username,''), NULLIF(approver_username,''), 'Belum ditugaskan') AS primary, '' AS secondary, subject_name, level_name, 'Review/verifikasi' AS task, ` + commonCounts + `, 0::bigint, 0::bigint, COUNT(*) FILTER (WHERE workflow='diperiksa')::bigint AS shortage, COUNT(*)::bigint AS volume, 'soal' AS unit` + suffix + `, 'Beban reviewer dan status verifikasi' AS notes FROM filtered WHERE workflow IN ('diperiksa','siap_pakai') OR (workflow='konsep' AND NULLIF(btrim(COALESCE(reviewer_username,'')), '') IS NOT NULL) GROUP BY primary, subject_name, level_name ORDER BY primary, subject_name, level_name`
	case "readiness":
		return ` SELECT subject_name AS primary, '' AS secondary, subject_name, level_name, 'Siap paket' AS task, ` + commonCounts + `, 20::bigint, 5::bigint, GREATEST(20 - COUNT(*) FILTER (WHERE qtype='multiple_choice' AND workflow='siap_pakai'),0)::bigint + GREATEST(5 - COUNT(*) FILTER (WHERE qtype='essay' AND workflow='siap_pakai'),0)::bigint AS shortage, COUNT(*) FILTER (WHERE workflow='siap_pakai')::bigint AS volume, 'soal siap' AS unit` + suffix + `, CASE WHEN COUNT(*) FILTER (WHERE workflow='siap_pakai') > 0 THEN 'Ada soal siap paket' ELSE 'Belum siap' END AS notes FROM filtered GROUP BY subject_name, level_name ORDER BY subject_name, level_name`
	case "honor":
		return ` SELECT author_name AS primary, COALESCE(author_username,'-') AS secondary, subject_name, level_name, 'Pembuat soal' AS task, ` + commonCounts + `, 0::bigint, 0::bigint, 0::bigint AS shortage, COUNT(*)::bigint AS volume, 'soal' AS unit` + suffix + `, 'Volume input soal; nominal honor ditentukan kebijakan madrasah' AS notes FROM filtered GROUP BY author_name, author_username, subject_name, level_name ORDER BY author_name, subject_name, level_name`
	default:
		return ` SELECT author_name AS primary, COALESCE(author_username,'-') AS secondary, subject_name, level_name, 'Input soal' AS task, ` + commonCounts + `, 0::bigint, 0::bigint, 0::bigint AS shortage, COUNT(*)::bigint AS volume, 'soal' AS unit` + suffix + `, '' AS notes FROM filtered GROUP BY author_name, author_username, subject_name, level_name ORDER BY CASE WHEN COALESCE(author_username,'')='system' THEN 1 ELSE 0 END, author_name, subject_name, level_name`
	}
}

func summarizeBankSoalRows(rows []BankSoalReportRow) BankSoalReportSummary {
	var s BankSoalReportSummary
	authors := map[string]bool{}
	subjects := map[string]bool{}
	for _, r := range rows {
		s.Total += r.Total
		s.PG += r.PG
		s.Essay += r.Essay
		s.Other += r.Other
		s.Shortage += r.Shortage
		if r.Primary != "" {
			authors[r.Primary] = true
		}
		if r.SubjectName != "" {
			subjects[r.SubjectName] = true
		}
		if r.DataWarning {
			s.MissingLevel++
		}
		if r.SystemRow {
			s.SystemRows += r.Total
		}
	}
	s.Authors = int64(len(authors))
	s.Subjects = int64(len(subjects))
	return s
}
func statusesFromRows(rows []BankSoalReportRow) []BankSoalReportStatus {
	m := map[string]int64{}
	for _, r := range rows {
		m["konsep"] += r.Draft
		m["diperiksa"] += r.Review
		m["siap_pakai"] += r.Approved
	}
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if v > 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	out := []BankSoalReportStatus{}
	for _, k := range keys {
		out = append(out, BankSoalReportStatus{Status: k, Total: m[k]})
	}
	return out
}
func statusString(r BankSoalReportRow) string {
	parts := []string{}
	vals := []struct {
		k string
		v int64
	}{{"konsep", r.Draft}, {"diperiksa", r.Review}, {"siap_pakai", r.Approved}}
	for _, x := range vals {
		if x.v > 0 {
			parts = append(parts, x.k+":"+strconv.FormatInt(x.v, 10))
		}
	}
	return strings.Join(parts, ", ")
}

func renderBankSoalReportCSV(r BankSoalReportResult) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"No", "Laporan", "Pembuat/Utama", "Username/Sekunder", "Mapel", "Tingkat", "Tugas", "Total", "PG", "Essay", "Lain", "Konsep", "Diperiksa", "Siap Pakai", "Perlu Revisi", "Target PG", "Target Essay", "Kurang", "Volume", "Satuan", "Input Pertama", "Input Terakhir", "Keterangan"})
	for _, row := range r.Rows {
		_ = w.Write([]string{strconv.Itoa(row.No), r.Title, row.Primary, row.Secondary, row.SubjectName, row.LevelName, row.Task, i64(row.Total), i64(row.PG), i64(row.Essay), i64(row.Other), i64(row.Draft), i64(row.Review), i64(row.Approved), i64(row.Revision), i64(row.TargetPG), i64(row.TargetEssay), i64(row.Shortage), i64(row.Volume), row.Unit, row.FirstInput, row.LastInput, row.Notes})
	}
	w.Flush()
	return b.Bytes(), w.Error()
}
func i64(v int64) string { return strconv.FormatInt(v, 10) }

func renderBankSoalReportPNG(r BankSoalReportResult) ([]byte, error) {
	htmlData := renderBankSoalReportHTML(r)
	dir, err := os.MkdirTemp("", "bank-soal-report-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	htmlPath := filepath.Join(dir, "report.html")
	pngPath := filepath.Join(dir, "report.png")
	if err := os.WriteFile(htmlPath, []byte(htmlData), 0600); err != nil {
		return nil, err
	}
	chrome, err := bankSoalReportChrome()
	if err != nil {
		return nil, err
	}
	// Use a dynamic viewport height so the exported PNG becomes a long-image
	// attachment: all report rows are visible in a single image instead of being
	// clipped by the old fixed 1800px screenshot viewport.
	height := bankSoalReportPNGHeight(len(r.Rows))
	cmd := exec.Command(chrome, "--headless", "--no-sandbox", "--disable-gpu", "--hide-scrollbars", fmt.Sprintf("--window-size=1900,%d", height), "--screenshot="+pngPath, "file://"+htmlPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("render gambar gagal: %s", string(out))
	}
	return os.ReadFile(pngPath)
}

func renderBankSoalReportPDF(r BankSoalReportResult) ([]byte, error) {
	htmlData := renderBankSoalReportHTML(r)
	dir, err := os.MkdirTemp("", "bank-soal-report-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	htmlPath := filepath.Join(dir, "report.html")
	pdfPath := filepath.Join(dir, "report.pdf")
	if err := os.WriteFile(htmlPath, []byte(htmlData), 0600); err != nil {
		return nil, err
	}
	chrome, err := bankSoalReportChrome()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(chrome, "--headless", "--no-sandbox", "--disable-gpu", "--print-to-pdf="+pdfPath, "file://"+htmlPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("render PDF gagal: %s", string(out))
	}
	return os.ReadFile(pdfPath)
}

func bankSoalReportChrome() (string, error) {
	if p, err := exec.LookPath("google-chrome"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("chromium"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("renderer gambar/PDF tidak tersedia")
}

func bankSoalReportPNGHeight(rowCount int) int {
	// Compact export header is ~150px, table header ~44px, each row ~30px.
	// Keep a small bottom margin and clamp to Chromium's practical screenshot size.
	height := 230 + rowCount*32
	if height < 900 {
		return 900
	}
	if height > 16000 {
		return 16000
	}
	return height
}

func renderBankSoalReportHTML(r BankSoalReportResult) string {
	rows := ""
	for _, row := range r.Rows {
		cls := ""
		if row.SystemRow {
			cls = "system"
		}
		if row.DataWarning {
			cls += " warning"
		}
		rows += fmt.Sprintf(`<tr class="%s"><td>%d</td><td><b>%s</b></td><td>%s</td><td>%s</td><td class="num">%d</td><td class="num">%d</td><td class="num">%d</td><td>%s</td><td>%s</td></tr>`, cls, row.No, esc(row.Primary), esc(bankSoalReportShortSubject(row.SubjectName)), esc(row.LevelName), row.Total, row.PG, row.Essay, esc(bankSoalReportShortStatuses(row.Statuses)), esc(bankSoalReportShortDate(row.LastInput)))
	}
	return fmt.Sprintf(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<style>
	*{box-sizing:border-box}
	body{font-family:Arial,sans-serif;background:#eef2f7;margin:0;color:#0f172a}
	.page{width:1840px;padding:18px}
	.banner{display:flex;align-items:center;gap:14px;flex-wrap:wrap;background:white;border:1px solid #d1fae5;border-left:10px solid #0f766e;border-radius:16px;padding:14px 18px;margin-bottom:12px;box-shadow:0 8px 24px #0f172a12;font-size:22px;line-height:1.18}
	.banner b{font-size:27px;color:#0f766e}
	.banner span{color:#334155;border-left:1px solid #cbd5e1;padding-left:14px;font-weight:700}
	table{width:100%%;border-collapse:separate;border-spacing:0;background:white;border-radius:16px;overflow:hidden;box-shadow:0 10px 28px #0f172a14;table-layout:fixed}
	th{background:#0f172a;color:white;text-align:left;padding:8px 9px;font-size:18px;line-height:1.08}
	td{padding:6px 9px;border-bottom:1px solid #e5e7eb;font-size:17px;line-height:1.08;vertical-align:middle;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
	tr:nth-child(even) td{background:#f8fafc}.system td{background:#eef6ff!important}.warning td{background:#fff7ed!important}.num{text-align:right;font-variant-numeric:tabular-nums}
	th:nth-child(1),td:nth-child(1){width:54px;text-align:center;color:#64748b}th:nth-child(2),td:nth-child(2){width:430px}th:nth-child(3),td:nth-child(3){width:240px}th:nth-child(4),td:nth-child(4){width:85px;text-align:center}th:nth-child(5),td:nth-child(5),th:nth-child(6),td:nth-child(6),th:nth-child(7),td:nth-child(7){width:82px}th:nth-child(8),td:nth-child(8){width:520px}th:nth-child(9),td:nth-child(9){width:150px;text-align:right;color:#475569}
	.footer{margin-top:8px;color:#64748b;font-size:14px;text-align:right}
</style>
</head>
<body>
<div class="page">
	<div class="banner"><b>%s</b><span>%d soal</span><span>%d pembuat/grup</span><span>%d mapel</span><span>%d baris</span><span>%s</span></div>
	<table><thead><tr><th>No</th><th>Utama/Pembuat</th><th>Mapel</th><th>Tingkat</th><th>Total</th><th>PG</th><th>Essay</th><th>Status</th><th>Terakhir</th></tr></thead><tbody>%s</tbody></table>
	<div class="footer">Bank Soal MTsN 2 Kolaka Utara · Dibuat %s WITA</div>
</div>
</body>
</html>`, esc(r.Title), r.Summary.Total, r.Summary.Authors, r.Summary.Subjects, len(r.Rows), esc(r.PeriodLabel), rows, esc(r.GeneratedAt))
}

func bankSoalReportShortSubject(value string) string {
	switch strings.TrimSpace(value) {
	case "Pendidikan Jasmani, Olahraga, dan Kesehatan", "Pendidikan Jasmani Olahraga dan Kesehatan":
		return "PJOK"
	case "Ilmu Pengetahuan Alam":
		return "IPA"
	case "Ilmu Pengetahuan Sosial":
		return "IPS"
	case "Bahasa Indonesia":
		return "B. Indonesia"
	case "Bahasa Inggris":
		return "B. Inggris"
	case "Bahasa Arab":
		return "B. Arab"
	case "Pendidikan Pancasila dan Kewarganegaraan", "Pendidikan Pancasila":
		return "PPKn"
	case "Mulok Kewirausahaan", "Muatan Lokal Kewirausahaan":
		return "Mulok KWU"
	case "Sejarah Kebudayaan Islam":
		return "SKI"
	case "Al-Qur'an Hadis", "Qur'an Hadits":
		return "Qurdis"
	case "Akidah Akhlak":
		return "Akidah"
	case "Matematika":
		return "MTK"
	default:
		return strings.TrimSpace(value)
	}
}

func bankSoalReportShortStatuses(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	aliases := map[string]string{
		"konsep":     "Konsep",
		"diperiksa":  "Diperiksa",
		"siap_pakai": "Siap",
	}
	parts := strings.Split(text, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		pair := strings.SplitN(item, ":", 2)
		label := strings.TrimSpace(pair[0])
		if alias, ok := aliases[label]; ok {
			label = alias
		}
		if len(pair) == 2 && strings.TrimSpace(pair[1]) != "" {
			out = append(out, label+" "+strings.TrimSpace(pair[1]))
		} else {
			out = append(out, label)
		}
	}
	return strings.Join(out, " · ")
}

func bankSoalReportShortDate(value string) string {
	text := strings.TrimSpace(value)
	if len(text) >= 16 && text[4] == '-' && text[7] == '-' {
		return text[8:10] + "/" + text[5:7] + " " + text[11:16]
	}
	return text
}

func esc(s string) string { return html.EscapeString(s) }
