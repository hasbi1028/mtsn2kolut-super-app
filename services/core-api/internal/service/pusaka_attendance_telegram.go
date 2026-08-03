package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const (
	attendanceTelegramDefaultTimezone = "Asia/Makassar"
	attendanceTelegramDefaultChatID   = "1450267717"
)

var (
	ErrAttendanceTelegramNotConfigured = errors.New("konfigurasi Telegram belum lengkap")
	ErrAttendanceTelegramDisabled      = errors.New("jadwal Telegram daftar hadir belum aktif")
)

type pusakaAttendanceTelegramStore interface {
	GetPusakaAttendanceTelegramSettings(ctx context.Context) (db.GetPusakaAttendanceTelegramSettingsRow, error)
	UpsertPusakaAttendanceTelegramSettings(ctx context.Context, arg db.UpsertPusakaAttendanceTelegramSettingsParams) (db.UpsertPusakaAttendanceTelegramSettingsRow, error)
	ListPusakaAttendanceTelegramReportRows(ctx context.Context, tanggal pgtype.Date) ([]db.ListPusakaAttendanceTelegramReportRowsRow, error)
	HasPusakaAttendanceTelegramScheduledLog(ctx context.Context, arg db.HasPusakaAttendanceTelegramScheduledLogParams) (bool, error)
	CreatePusakaAttendanceTelegramLog(ctx context.Context, arg db.CreatePusakaAttendanceTelegramLogParams) (db.CreatePusakaAttendanceTelegramLogRow, error)
	ListPusakaAttendanceTelegramLogs(ctx context.Context, arg db.ListPusakaAttendanceTelegramLogsParams) ([]db.ListPusakaAttendanceTelegramLogsRow, error)
}

type PusakaAttendanceTelegram struct {
	q          pusakaAttendanceTelegramStore
	botToken   string
	httpClient *http.Client
	loc        *time.Location
	stop       context.CancelFunc
}

type AttendanceTelegramSettingsResponse struct {
	IsEnabled          bool     `json:"is_enabled"`
	SendTime           string   `json:"send_time"`
	SendTimes          []string `json:"send_times"`
	SendDays           []int32  `json:"send_days"`
	Timezone           string   `json:"timezone"`
	TargetChatID       string   `json:"target_chat_id"`
	TargetChatIDMasked string   `json:"target_chat_id_masked"`
	IncludeCaption     bool     `json:"include_caption"`
	IncludeImage       bool     `json:"include_image"`
	ReportMode         string   `json:"report_mode"`
	BotConfigured      bool     `json:"bot_configured"`
}

type UpdateAttendanceTelegramSettingsInput struct {
	IsEnabled      bool     `json:"is_enabled"`
	SendTime       string   `json:"send_time"`
	SendTimes      []string `json:"send_times"`
	SendDays       []int32  `json:"send_days"`
	Timezone       string   `json:"timezone"`
	TargetChatID   string   `json:"target_chat_id"`
	IncludeCaption bool     `json:"include_caption"`
	IncludeImage   bool     `json:"include_image"`
	ReportMode     string   `json:"report_mode"`
}

type SendAttendanceTelegramReportInput struct {
	Date           string      `json:"date"`
	TargetChatID   string      `json:"target_chat_id"`
	IncludeCaption bool        `json:"include_caption"`
	IncludeImage   bool        `json:"include_image"`
	SendMode       string      `json:"send_mode"`
	RequestedBy    pgtype.UUID `json:"-"`
	ScheduleTime   string      `json:"-"`
}

type AttendanceTelegramReportResult struct {
	ReportDate         string `json:"report_date"`
	TargetChatIDMasked string `json:"target_chat_id_masked"`
	SendMode           string `json:"send_mode"`
	Status             string `json:"status"`
	TelegramMessageID  string `json:"telegram_message_id,omitempty"`
	ErrorMessage       string `json:"error_message,omitempty"`
	TotalEmployees     int    `json:"total_employees"`
	CheckedIn          int    `json:"checked_in"`
	NotCheckedIn       int    `json:"not_checked_in"`
	CheckedOut         int    `json:"checked_out"`
}

type attendanceReportRow struct {
	No           int
	EmployeeName string
	EmployeeNIP  string
	CheckIn      string
	CheckOut     string
	Status       string
}

type attendanceReport struct {
	Date           time.Time
	GeneratedAt    time.Time
	Rows           []attendanceReportRow
	TotalEmployees int
	CheckedIn      int
	NotCheckedIn   int
	CheckedOut     int
}

func NewPusakaAttendanceTelegram(q *db.Queries, botToken string) *PusakaAttendanceTelegram {
	loc, _ := time.LoadLocation(attendanceTelegramDefaultTimezone)
	if loc == nil {
		loc = time.FixedZone("WITA", 8*3600)
	}
	return &PusakaAttendanceTelegram{q: q, botToken: strings.TrimSpace(botToken), httpClient: &http.Client{Timeout: 30 * time.Second}, loc: loc}
}

func (s *PusakaAttendanceTelegram) Start(ctx context.Context) {
	if s == nil || s.q == nil {
		return
	}
	loopCtx, cancel := context.WithCancel(ctx)
	s.stop = cancel
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-loopCtx.Done():
				return
			case now := <-ticker.C:
				_ = s.RunDue(loopCtx, now)
			}
		}
	}()
}

func (s *PusakaAttendanceTelegram) Stop() {
	if s != nil && s.stop != nil {
		s.stop()
	}
}

func (s *PusakaAttendanceTelegram) GetSettings(ctx context.Context) (AttendanceTelegramSettingsResponse, error) {
	row, err := s.q.GetPusakaAttendanceTelegramSettings(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return AttendanceTelegramSettingsResponse{IsEnabled: false, SendTime: "17:00", SendTimes: []string{"17:00"}, SendDays: defaultAttendanceTelegramSendDays(), Timezone: attendanceTelegramDefaultTimezone, TargetChatID: attendanceTelegramDefaultChatID, TargetChatIDMasked: maskChatID(attendanceTelegramDefaultChatID), IncludeCaption: true, IncludeImage: true, ReportMode: "ringkas", BotConfigured: s.botToken != ""}, nil
	}
	if err != nil {
		return AttendanceTelegramSettingsResponse{}, err
	}
	return s.settingsResponse(row), nil
}

func (s *PusakaAttendanceTelegram) UpdateSettings(ctx context.Context, in UpdateAttendanceTelegramSettingsInput) (AttendanceTelegramSettingsResponse, error) {
	if in.Timezone == "" {
		in.Timezone = attendanceTelegramDefaultTimezone
	}
	if in.Timezone != attendanceTelegramDefaultTimezone {
		return AttendanceTelegramSettingsResponse{}, fmt.Errorf("timezone hanya mendukung %s", attendanceTelegramDefaultTimezone)
	}
	if in.ReportMode == "" {
		in.ReportMode = "ringkas"
	}
	if in.ReportMode != "ringkas" {
		return AttendanceTelegramSettingsResponse{}, errors.New("mode laporan tidak valid")
	}
	if !in.IncludeCaption && !in.IncludeImage {
		return AttendanceTelegramSettingsResponse{}, errors.New("minimal caption atau gambar harus aktif")
	}
	if in.IsEnabled && strings.TrimSpace(in.TargetChatID) == "" {
		return AttendanceTelegramSettingsResponse{}, errors.New("target Telegram wajib diisi saat jadwal aktif")
	}
	sendTimes, err := normalizeSendTimes(in.SendTimes, in.SendTime)
	if err != nil {
		return AttendanceTelegramSettingsResponse{}, err
	}
	sendDays, err := normalizeSendDays(in.SendDays)
	if err != nil {
		return AttendanceTelegramSettingsResponse{}, err
	}
	micros, err := parseHHMMToMicros(sendTimes[0])
	if err != nil {
		return AttendanceTelegramSettingsResponse{}, err
	}
	row, err := s.q.UpsertPusakaAttendanceTelegramSettings(ctx, db.UpsertPusakaAttendanceTelegramSettingsParams{IsEnabled: in.IsEnabled, SendTime: pgtype.Time{Microseconds: micros, Valid: true}, SendTimes: sendTimes, SendDays: sendDays, Timezone: in.Timezone, TargetChatID: strings.TrimSpace(in.TargetChatID), IncludeCaption: in.IncludeCaption, IncludeImage: in.IncludeImage, ReportMode: in.ReportMode})
	if err != nil {
		return AttendanceTelegramSettingsResponse{}, err
	}
	return s.settingsResponseFromUpsert(row), nil
}

func (s *PusakaAttendanceTelegram) ListLogs(ctx context.Context, limit, offset int32) ([]db.ListPusakaAttendanceTelegramLogsRow, error) {
	return s.q.ListPusakaAttendanceTelegramLogs(ctx, db.ListPusakaAttendanceTelegramLogsParams{Limit: limit, Offset: offset})
}

func (s *PusakaAttendanceTelegram) RunDue(ctx context.Context, now time.Time) error {
	settings, err := s.q.GetPusakaAttendanceTelegramSettings(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !settings.IsEnabled {
		return nil
	}
	localNow := now.In(s.loc)
	if !sendDayEnabled(settings.SendDays, int32(localNow.Weekday())) {
		return nil
	}
	current := localNow.Format("15:04")
	sendTimes, err := normalizeSendTimes(settings.SendTimes, microsToHHMM(settings.SendTime.Microseconds))
	if err != nil {
		return err
	}
	for _, scheduled := range sendTimes {
		if current != scheduled {
			continue
		}
		reportDate := datePg(localNow)
		exists, err := s.q.HasPusakaAttendanceTelegramScheduledLog(ctx, db.HasPusakaAttendanceTelegramScheduledLogParams{ReportDate: reportDate, TargetChatID: settings.TargetChatID, ScheduleTime: scheduled})
		if err != nil || exists {
			return err
		}
		_, err = s.SendReport(ctx, SendAttendanceTelegramReportInput{Date: localNow.Format("2006-01-02"), TargetChatID: settings.TargetChatID, IncludeCaption: settings.IncludeCaption, IncludeImage: settings.IncludeImage, SendMode: "scheduled", ScheduleTime: scheduled})
		return err
	}
	return nil
}

func (s *PusakaAttendanceTelegram) SendReport(ctx context.Context, in SendAttendanceTelegramReportInput) (AttendanceTelegramReportResult, error) {
	if in.SendMode == "" {
		in.SendMode = "manual"
	}
	if in.Date == "" {
		in.Date = time.Now().In(s.loc).Format("2006-01-02")
	}
	date, err := time.ParseInLocation("2006-01-02", in.Date, s.loc)
	if err != nil {
		return AttendanceTelegramReportResult{}, errors.New("tanggal tidak valid, gunakan YYYY-MM-DD")
	}
	if !in.IncludeCaption && !in.IncludeImage {
		in.IncludeCaption = true
		in.IncludeImage = true
	}
	settings, _ := s.GetSettings(ctx)
	chatID := strings.TrimSpace(in.TargetChatID)
	if chatID == "" {
		chatID = strings.TrimSpace(settings.TargetChatID)
	}
	if chatID == "" {
		chatID = attendanceTelegramDefaultChatID
	}

	report, err := s.buildReport(ctx, date)
	if err != nil {
		return AttendanceTelegramReportResult{}, err
	}
	result := AttendanceTelegramReportResult{ReportDate: date.Format("2006-01-02"), TargetChatIDMasked: maskChatID(chatID), SendMode: in.SendMode, TotalEmployees: report.TotalEmployees, CheckedIn: report.CheckedIn, NotCheckedIn: report.NotCheckedIn, CheckedOut: report.CheckedOut}
	caption := buildAttendanceCaption(report)
	var imageBytes []byte
	if in.IncludeImage {
		imageBytes, err = renderAttendancePNG(report)
		if err != nil {
			return result, err
		}
	}
	msgID, sendErr := s.sendTelegram(ctx, chatID, caption, imageBytes, in.IncludeImage)
	status := "success"
	var errText pgtype.Text
	if sendErr != nil {
		status = "failed"
		errText = pgtype.Text{String: sendErr.Error(), Valid: true}
		result.ErrorMessage = sendErr.Error()
	}
	var msgText pgtype.Text
	if msgID != "" {
		msgText = pgtype.Text{String: msgID, Valid: true}
		result.TelegramMessageID = msgID
	}
	_, _ = s.q.CreatePusakaAttendanceTelegramLog(ctx, db.CreatePusakaAttendanceTelegramLogParams{ReportDate: datePg(date), TargetChatID: chatID, SendMode: in.SendMode, ScheduleTime: in.ScheduleTime, Status: status, TelegramMessageID: msgText, ErrorMessage: errText, RequestedBy: in.RequestedBy})
	result.Status = status
	return result, sendErr
}

func (s *PusakaAttendanceTelegram) buildReport(ctx context.Context, date time.Time) (attendanceReport, error) {
	rows, err := s.q.ListPusakaAttendanceTelegramReportRows(ctx, datePg(date))
	if err != nil {
		return attendanceReport{}, err
	}
	report := attendanceReport{Date: date, GeneratedAt: time.Now().In(s.loc), TotalEmployees: len(rows)}
	for _, r := range rows {
		if excludedAttendanceReportName(r.EmployeeNama) {
			continue
		}
		in := cleanTime(r.JamMasuk)
		out := cleanTime(r.JamPulang)
		status := "Belum"
		if in != "-" && out != "-" {
			status = "Lengkap"
		} else if in != "-" {
			status = "Masuk"
		}
		if in != "-" {
			report.CheckedIn++
		} else {
			report.NotCheckedIn++
		}
		if out != "-" {
			report.CheckedOut++
		}
		report.Rows = append(report.Rows, attendanceReportRow{No: len(report.Rows) + 1, EmployeeName: r.EmployeeNama, EmployeeNIP: r.EmployeeNip, CheckIn: in, CheckOut: out, Status: status})
	}
	report.TotalEmployees = len(report.Rows)
	return report, nil
}

func excludedAttendanceReportName(name string) bool {
	normalized := strings.ToUpper(strings.Join(strings.Fields(name), " "))
	if normalized == "" {
		return false
	}
	for _, marker := range []string{"DEV", "TEST", "DUMMY", "CBT"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func (s *PusakaAttendanceTelegram) sendTelegram(ctx context.Context, chatID, caption string, imageBytes []byte, includeImage bool) (string, error) {
	if s.botToken == "" {
		return "", ErrAttendanceTelegramNotConfigured
	}
	if strings.TrimSpace(chatID) == "" {
		return "", ErrAttendanceTelegramNotConfigured
	}
	if !includeImage || len(imageBytes) == 0 {
		return s.telegramRequest(ctx, "sendMessage", map[string]string{"chat_id": chatID, "text": caption})
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("chat_id", chatID)
	_ = mw.WriteField("caption", caption)
	fw, err := mw.CreateFormFile("photo", "daftar-hadir-ringkas.png")
	if err != nil {
		return "", err
	}
	_, _ = fw.Write(imageBytes)
	_ = mw.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+s.botToken+"/sendPhoto", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return s.doTelegram(req)
}

// SendAlertText mengirim pesan teks singkat (alert worker offline/online)
// ke chat id Telegram yang dikonfigurasi. Dipakai scheduler health monitoring.
func (s *PusakaAttendanceTelegram) SendAlertText(ctx context.Context, text string) (string, error) {
	if s == nil || s.botToken == "" {
		return "", ErrAttendanceTelegramNotConfigured
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return "", err
	}
	chatID := strings.TrimSpace(settings.TargetChatID)
	if chatID == "" {
		chatID = attendanceTelegramDefaultChatID
	}
	return s.sendTelegram(ctx, chatID, text, nil, false)
}

func (s *PusakaAttendanceTelegram) telegramRequest(ctx context.Context, method string, fields map[string]string) (string, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	_ = mw.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+s.botToken+"/"+method, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return s.doTelegram(req)
}

func (s *PusakaAttendanceTelegram) doTelegram(req *http.Request) (string, error) {
	res, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("telegram HTTP %d", res.StatusCode)
	}
	var payload struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
		Description string `json:"description"`
	}
	_ = json.Unmarshal(b, &payload)
	if !payload.OK {
		if payload.Description != "" {
			return "", errors.New(payload.Description)
		}
		return "", errors.New("telegram response tidak valid")
	}
	if payload.Result.MessageID > 0 {
		return strconv.Itoa(payload.Result.MessageID), nil
	}
	return "", nil
}

func (s *PusakaAttendanceTelegram) settingsResponse(row db.GetPusakaAttendanceTelegramSettingsRow) AttendanceTelegramSettingsResponse {
	sendTime := microsToHHMM(row.SendTime.Microseconds)
	sendTimes, _ := normalizeSendTimes(row.SendTimes, sendTime)
	sendDays, _ := normalizeSendDays(row.SendDays)
	return AttendanceTelegramSettingsResponse{IsEnabled: row.IsEnabled, SendTime: sendTime, SendTimes: sendTimes, SendDays: sendDays, Timezone: row.Timezone, TargetChatID: row.TargetChatID, TargetChatIDMasked: maskChatID(row.TargetChatID), IncludeCaption: row.IncludeCaption, IncludeImage: row.IncludeImage, ReportMode: row.ReportMode, BotConfigured: s.botToken != ""}
}

func (s *PusakaAttendanceTelegram) settingsResponseFromUpsert(row db.UpsertPusakaAttendanceTelegramSettingsRow) AttendanceTelegramSettingsResponse {
	sendTime := microsToHHMM(row.SendTime.Microseconds)
	sendTimes, _ := normalizeSendTimes(row.SendTimes, sendTime)
	sendDays, _ := normalizeSendDays(row.SendDays)
	return AttendanceTelegramSettingsResponse{IsEnabled: row.IsEnabled, SendTime: sendTime, SendTimes: sendTimes, SendDays: sendDays, Timezone: row.Timezone, TargetChatID: row.TargetChatID, TargetChatIDMasked: maskChatID(row.TargetChatID), IncludeCaption: row.IncludeCaption, IncludeImage: row.IncludeImage, ReportMode: row.ReportMode, BotConfigured: s.botToken != ""}
}

func defaultAttendanceTelegramSendDays() []int32 {
	return []int32{1, 2, 3, 4, 5, 6}
}

func normalizeSendDays(values []int32) ([]int32, error) {
	if len(values) == 0 {
		return defaultAttendanceTelegramSendDays(), nil
	}
	seen := map[int32]bool{}
	out := make([]int32, 0, len(values))
	for _, day := range values {
		if day < 0 || day > 6 {
			return nil, errors.New("hari kirim tidak valid")
		}
		if !seen[day] {
			seen[day] = true
			out = append(out, day)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("minimal satu hari kirim harus dipilih")
	}
	return out, nil
}

func sendDayEnabled(values []int32, day int32) bool {
	sendDays, err := normalizeSendDays(values)
	if err != nil {
		return false
	}
	for _, enabled := range sendDays {
		if enabled == day {
			return true
		}
	}
	return false
}

func normalizeSendTimes(values []string, fallback string) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(values)+1)
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		micros, err := parseHHMMToMicros(value)
		if err != nil {
			return nil, err
		}
		normalized := microsToHHMM(micros)
		if !seen[normalized] {
			seen[normalized] = true
			out = append(out, normalized)
		}
	}
	if len(out) == 0 {
		if strings.TrimSpace(fallback) == "" {
			fallback = "17:00"
		}
		micros, err := parseHHMMToMicros(fallback)
		if err != nil {
			return nil, err
		}
		out = append(out, microsToHHMM(micros))
	}
	if len(out) > 8 {
		return nil, errors.New("maksimal 8 jadwal kirim per hari")
	}
	return out, nil
}

func parseHHMMToMicros(v string) (int64, error) {
	parts := strings.Split(strings.TrimSpace(v), ":")
	if len(parts) < 2 {
		return 0, errors.New("jam kirim tidak valid")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, errors.New("jam kirim tidak valid")
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, errors.New("jam kirim tidak valid")
	}
	return int64((time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) / time.Microsecond), nil
}
func microsToHHMM(us int64) string {
	d := time.Duration(us) * time.Microsecond
	h := int(d / time.Hour)
	m := int((d % time.Hour) / time.Minute)
	return fmt.Sprintf("%02d:%02d", h, m)
}
func datePg(t time.Time) pgtype.Date {
	return pgtype.Date{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), Valid: true}
}
func cleanTime(v string) string {
	v = strings.TrimSpace(strings.ReplaceAll(v, "WITA", ""))
	if v == "" {
		return "-"
	}
	for _, field := range strings.Fields(v) {
		if strings.Contains(field, ":") {
			v = field
			break
		}
	}
	if dot := strings.IndexByte(v, '.'); dot > 0 {
		v = v[:dot]
	}
	if len(v) >= 8 && v[2] == ':' && v[5] == ':' {
		return v[:8]
	}
	if len(v) >= 5 && v[2] == ':' {
		return v[:5]
	}
	return v
}
func maskChatID(v string) string {
	v = strings.TrimSpace(v)
	if len(v) <= 4 {
		return strings.Repeat("*", len(v))
	}
	return strings.Repeat("*", len(v)-4) + v[len(v)-4:]
}

func buildAttendanceCaption(r attendanceReport) string {
	return fmt.Sprintf("📋 Daftar Hadir Ringkas\nMTsN 2 Kolaka Utara\nTanggal: %s\n\nTotal Pegawai: %d\nSudah Masuk: %d\nBelum Masuk: %d\nSudah Pulang: %d\n\nDikirim otomatis dari Sistem MTsN 2 Kolaka Utara.", r.Date.Format("02 Jan 2006"), r.TotalEmployees, r.CheckedIn, r.NotCheckedIn, r.CheckedOut)
}

func renderAttendancePNG(r attendanceReport) ([]byte, error) {
	width := 1200
	rowH := 30
	height := 220 + len(r.Rows)*rowH + 50
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{248, 250, 252, 255}}, image.Point{}, draw.Src)
	fillRect(img, 30, 30, width-60, 96, color.RGBA{15, 118, 110, 255})
	drawText(img, 58, 55, "MTSN 2 KOLAKA UTARA", 3, color.White)
	drawText(img, 58, 88, "DAFTAR HADIR RINGKAS - "+r.Date.Format("2006-01-02"), 2, color.White)
	cards := []string{fmt.Sprintf("TOTAL %d", r.TotalEmployees), fmt.Sprintf("MASUK %d", r.CheckedIn), fmt.Sprintf("BELUM %d", r.NotCheckedIn), fmt.Sprintf("PULANG %d", r.CheckedOut)}
	for i, c := range cards {
		x := 30 + i*285
		fillRect(img, x, 140, 265, 54, color.RGBA{255, 255, 255, 255})
		strokeRect(img, x, 140, 265, 54, color.RGBA{203, 213, 225, 255})
		drawText(img, x+18, 158, c, 2, color.RGBA{15, 23, 42, 255})
	}
	y := 220
	fillRect(img, 30, y, width-60, rowH, color.RGBA{226, 232, 240, 255})
	headers := []string{"NO", "NAMA PEGAWAI", "MASUK", "PULANG", "STATUS"}
	xs := []int{50, 110, 735, 860, 990}
	for i, h := range headers {
		drawText(img, xs[i], y+9, h, 2, color.RGBA{15, 23, 42, 255})
	}
	y += rowH
	for i, row := range r.Rows {
		bg := color.RGBA{255, 255, 255, 255}
		if i%2 == 1 {
			bg = color.RGBA{241, 245, 249, 255}
		}
		fillRect(img, 30, y, width-60, rowH, bg)
		strokeRect(img, 30, y, width-60, rowH, color.RGBA{226, 232, 240, 255})
		drawText(img, 50, y+8, fmt.Sprintf("%02d", row.No), 2, color.RGBA{51, 65, 85, 255})
		drawText(img, 110, y+8, truncate(row.EmployeeName, 46), 2, color.RGBA{15, 23, 42, 255})
		drawText(img, 735, y+8, row.CheckIn, 2, color.RGBA{15, 23, 42, 255})
		drawText(img, 860, y+8, row.CheckOut, 2, color.RGBA{15, 23, 42, 255})
		drawText(img, 990, y+8, strings.ToUpper(row.Status), 2, statusColor(row.Status))
		y += rowH
	}
	drawText(img, 30, height-28, "GENERATED "+r.GeneratedAt.Format("2006-01-02 15:04:05 WITA"), 2, color.RGBA{100, 116, 139, 255})
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}
func statusColor(s string) color.Color {
	switch strings.ToLower(s) {
	case "lengkap":
		return color.RGBA{22, 101, 52, 255}
	case "masuk":
		return color.RGBA{180, 83, 9, 255}
	default:
		return color.RGBA{100, 116, 139, 255}
	}
}
func fillRect(img *image.RGBA, x, y, w, h int, c color.Color) {
	draw.Draw(img, image.Rect(x, y, x+w, y+h), &image.Uniform{c}, image.Point{}, draw.Src)
}
func strokeRect(img *image.RGBA, x, y, w, h int, c color.Color) {
	fillRect(img, x, y, w, 1, c)
	fillRect(img, x, y+h-1, w, 1, c)
	fillRect(img, x, y, 1, h, c)
	fillRect(img, x+w-1, y, 1, h, c)
}
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

var pixelFont = map[rune][]string{
	' ': {"00000", "00000", "00000", "00000", "00000", "00000", "00000"}, '-': {"00000", "00000", "00000", "11110", "00000", "00000", "00000"}, '.': {"00000", "00000", "00000", "00000", "00000", "01100", "01100"}, ',': {"00000", "00000", "00000", "00000", "00000", "01100", "01000"}, ':': {"00000", "01100", "01100", "00000", "01100", "01100", "00000"}, '/': {"00001", "00010", "00100", "01000", "10000", "00000", "00000"}, '…': {"00000", "00000", "00000", "00000", "00000", "10101", "10101"},
	'0': {"01110", "10001", "10011", "10101", "11001", "10001", "01110"}, '1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"}, '2': {"01110", "10001", "00001", "00010", "00100", "01000", "11111"}, '3': {"11110", "00001", "00001", "01110", "00001", "00001", "11110"}, '4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"}, '5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"}, '6': {"00110", "01000", "10000", "11110", "10001", "10001", "01110"}, '7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"}, '8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"}, '9': {"01110", "10001", "10001", "01111", "00001", "00010", "01100"},
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"}, 'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"}, 'C': {"01110", "10001", "10000", "10000", "10000", "10001", "01110"}, 'D': {"11110", "10001", "10001", "10001", "10001", "10001", "11110"}, 'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"}, 'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"}, 'G': {"01110", "10001", "10000", "10111", "10001", "10001", "01110"}, 'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"}, 'I': {"01110", "00100", "00100", "00100", "00100", "00100", "01110"}, 'J': {"00001", "00001", "00001", "00001", "10001", "10001", "01110"}, 'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"}, 'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"}, 'M': {"10001", "11011", "10101", "10101", "10001", "10001", "10001"}, 'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"}, 'O': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"}, 'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"}, 'Q': {"01110", "10001", "10001", "10001", "10101", "10010", "01101"}, 'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"}, 'S': {"01111", "10000", "10000", "01110", "00001", "00001", "11110"}, 'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"}, 'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"}, 'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"}, 'W': {"10001", "10001", "10001", "10101", "10101", "10101", "01010"}, 'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"}, 'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"}, 'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
}

func drawText(img *image.RGBA, x, y int, text string, scale int, c color.Color) {
	cx := x
	for _, r := range strings.ToUpper(text) {
		if r == '\n' {
			y += 8 * scale
			cx = x
			continue
		}
		glyph, ok := pixelFont[r]
		if !ok {
			glyph = pixelFont[' ']
		}
		for gy, line := range glyph {
			for gx, p := range line {
				if p == '1' {
					fillRect(img, cx+gx*scale, y+gy*scale, scale, scale, c)
				}
			}
		}
		cx += 6 * scale
	}
}
