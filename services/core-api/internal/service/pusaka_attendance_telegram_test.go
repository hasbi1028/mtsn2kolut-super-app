package service

import (
	"bytes"
	"context"
	"errors"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeAttendanceTelegramStore struct {
	settings    db.GetPusakaAttendanceTelegramSettingsRow
	settingsErr error

	upsertArg db.UpsertPusakaAttendanceTelegramSettingsParams
	upsertRow db.UpsertPusakaAttendanceTelegramSettingsRow
	upsertErr error

	reportRows []db.ListPusakaAttendanceTelegramReportRowsRow
	reportErr  error

	hasLogArg db.HasPusakaAttendanceTelegramScheduledLogParams
	hasLog    bool
	hasLogErr error

	logArgs []db.CreatePusakaAttendanceTelegramLogParams
	logErr  error

	listLogsArg db.ListPusakaAttendanceTelegramLogsParams
	listLogs    []db.ListPusakaAttendanceTelegramLogsRow
	listLogsErr error
}

func (f *fakeAttendanceTelegramStore) GetPusakaAttendanceTelegramSettings(ctx context.Context) (db.GetPusakaAttendanceTelegramSettingsRow, error) {
	return f.settings, f.settingsErr
}

func (f *fakeAttendanceTelegramStore) UpsertPusakaAttendanceTelegramSettings(ctx context.Context, arg db.UpsertPusakaAttendanceTelegramSettingsParams) (db.UpsertPusakaAttendanceTelegramSettingsRow, error) {
	f.upsertArg = arg
	if f.upsertErr != nil {
		return db.UpsertPusakaAttendanceTelegramSettingsRow{}, f.upsertErr
	}
	if f.upsertRow.Timezone == "" {
		f.upsertRow = db.UpsertPusakaAttendanceTelegramSettingsRow{IsEnabled: arg.IsEnabled, SendTime: arg.SendTime, Timezone: arg.Timezone, TargetChatID: arg.TargetChatID, SendTimes: arg.SendTimes, SendDays: arg.SendDays, IncludeCaption: arg.IncludeCaption, IncludeImage: arg.IncludeImage, ReportMode: arg.ReportMode}
	}
	return f.upsertRow, nil
}

func (f *fakeAttendanceTelegramStore) ListPusakaAttendanceTelegramReportRows(ctx context.Context, tanggal pgtype.Date) ([]db.ListPusakaAttendanceTelegramReportRowsRow, error) {
	return f.reportRows, f.reportErr
}

func (f *fakeAttendanceTelegramStore) HasPusakaAttendanceTelegramScheduledLog(ctx context.Context, arg db.HasPusakaAttendanceTelegramScheduledLogParams) (bool, error) {
	f.hasLogArg = arg
	return f.hasLog, f.hasLogErr
}

func (f *fakeAttendanceTelegramStore) CreatePusakaAttendanceTelegramLog(ctx context.Context, arg db.CreatePusakaAttendanceTelegramLogParams) (db.CreatePusakaAttendanceTelegramLogRow, error) {
	f.logArgs = append(f.logArgs, arg)
	return db.CreatePusakaAttendanceTelegramLogRow{Status: arg.Status, TelegramMessageID: arg.TelegramMessageID, ErrorMessage: arg.ErrorMessage}, f.logErr
}

func (f *fakeAttendanceTelegramStore) ListPusakaAttendanceTelegramLogs(ctx context.Context, arg db.ListPusakaAttendanceTelegramLogsParams) ([]db.ListPusakaAttendanceTelegramLogsRow, error) {
	f.listLogsArg = arg
	return f.listLogs, f.listLogsErr
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func newTelegramTestService(store *fakeAttendanceTelegramStore, token string, rt http.RoundTripper) *PusakaAttendanceTelegram {
	loc, _ := time.LoadLocation(attendanceTelegramDefaultTimezone)
	if rt == nil {
		rt = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":77}}`))}, nil
		})
	}
	return &PusakaAttendanceTelegram{q: store, botToken: token, httpClient: &http.Client{Transport: rt}, loc: loc}
}

func telegramSettings(enabled bool, sendTimes []string, sendDays []int32) db.GetPusakaAttendanceTelegramSettingsRow {
	micros, _ := parseHHMMToMicros("17:00")
	return db.GetPusakaAttendanceTelegramSettingsRow{IsEnabled: enabled, SendTime: pgtype.Time{Microseconds: micros, Valid: true}, Timezone: attendanceTelegramDefaultTimezone, TargetChatID: "123456789", SendTimes: sendTimes, SendDays: sendDays, IncludeCaption: true, IncludeImage: false, ReportMode: "ringkas"}
}

func TestPusakaAttendanceTelegramPresentationHelpers(t *testing.T) {
	date := time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC)
	if got := datePg(date); !got.Valid || got.Time.Location() != time.UTC || got.Time.Hour() != 0 || got.Time.Day() != 17 {
		t.Fatalf("datePg() = %+v", got)
	}
	for input, want := range map[string]string{
		" 07:15:30 WITA ":         "07:15:30",
		"masuk 06:57:00.123 WITA": "06:57:00",
		"08:05":                   "08:05",
		"":                        "-",
	} {
		if got := cleanTime(input); got != want {
			t.Fatalf("cleanTime(%q) = %q, want %q", input, got, want)
		}
	}
	for input, want := range map[string]string{
		"123":         "***",
		" 987654321 ": "*****4321",
		"":            "",
	} {
		if got := maskChatID(input); got != want {
			t.Fatalf("maskChatID(%q) = %q, want %q", input, got, want)
		}
	}
	if got := truncate("Siti Aminah", 20); got != "Siti Aminah" {
		t.Fatalf("truncate short = %q", got)
	}
	if got := truncate("Nama Pegawai Sangat Panjang", 12); got != "Nama Pegawa…" {
		t.Fatalf("truncate long = %q", got)
	}
}

func TestPusakaAttendanceCaptionAndPNGRendering(t *testing.T) {
	report := attendanceReport{
		Date:           time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		GeneratedAt:    time.Date(2026, 5, 17, 10, 11, 12, 0, time.FixedZone("WITA", 8*3600)),
		TotalEmployees: 2,
		CheckedIn:      1,
		NotCheckedIn:   1,
		CheckedOut:     1,
		Rows: []attendanceReportRow{
			{No: 1, EmployeeName: "Andi", CheckIn: "07:01", CheckOut: "15:30", Status: "lengkap"},
			{No: 2, EmployeeName: "Budi", CheckIn: "-", CheckOut: "-", Status: "belum"},
		},
	}
	caption := buildAttendanceCaption(report)
	for _, want := range []string{"Daftar Hadir Ringkas", "Tanggal: 17 May 2026", "Total Pegawai: 2", "Sudah Masuk: 1", "Belum Masuk: 1", "Sudah Pulang: 1"} {
		if !strings.Contains(caption, want) {
			t.Fatalf("caption %q missing %q", caption, want)
		}
	}

	pngBytes, err := renderAttendancePNG(report)
	if err != nil {
		t.Fatalf("renderAttendancePNG() error = %v", err)
	}
	if !bytes.HasPrefix(pngBytes, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatalf("renderAttendancePNG() did not return PNG bytes")
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
	if got, want := img.Bounds().Dx(), 1200; got != want {
		t.Fatalf("PNG width = %d, want %d", got, want)
	}
	if got, want := img.Bounds().Dy(), 220+len(report.Rows)*30+50; got != want {
		t.Fatalf("PNG height = %d, want %d", got, want)
	}
}

func TestPusakaAttendanceStatusColors(t *testing.T) {
	cases := []struct {
		status string
		want   color.RGBA
	}{
		{"lengkap", color.RGBA{22, 101, 52, 255}},
		{"MASUK", color.RGBA{180, 83, 9, 255}},
		{"belum", color.RGBA{100, 116, 139, 255}},
	}
	for _, tc := range cases {
		got, ok := statusColor(tc.status).(color.RGBA)
		if !ok || got != tc.want {
			t.Fatalf("statusColor(%q) = %#v, want %#v", tc.status, statusColor(tc.status), tc.want)
		}
	}
}

func TestPusakaAttendanceTelegramConstructorStartStop(t *testing.T) {
	svc := NewPusakaAttendanceTelegram(nil, "  token-123  ")
	if svc == nil || svc.botToken != "token-123" || svc.httpClient == nil || svc.loc == nil {
		t.Fatalf("constructor did not initialize service correctly: %#v", svc)
	}
	svc.Start(context.Background())
	svc.Stop()

	store := &fakeAttendanceTelegramStore{}
	svc = newTelegramTestService(store, "token", nil)
	ctx, cancel := context.WithCancel(context.Background())
	svc.Start(ctx)
	if svc.stop == nil {
		t.Fatal("Start() did not install stop callback")
	}
	svc.Stop()
	cancel()
}

func TestPusakaAttendanceTelegramSettingsFlows(t *testing.T) {
	ctx := context.Background()

	store := &fakeAttendanceTelegramStore{settingsErr: pgx.ErrNoRows}
	svc := newTelegramTestService(store, "bot-token", nil)
	got, err := svc.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings default error = %v", err)
	}
	if got.IsEnabled || got.SendTime != "17:00" || got.TargetChatID != attendanceTelegramDefaultChatID || !got.BotConfigured || got.TargetChatIDMasked != "******7717" {
		t.Fatalf("unexpected default settings: %+v", got)
	}

	store = &fakeAttendanceTelegramStore{settings: telegramSettings(true, []string{"06:30", "06:30", "17:00"}, []int32{1, 1, 5})}
	svc = newTelegramTestService(store, "", nil)
	got, err = svc.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings row error = %v", err)
	}
	if !got.IsEnabled || got.SendTime != "17:00" || strings.Join(got.SendTimes, ",") != "06:30,17:00" || got.BotConfigured {
		t.Fatalf("unexpected row settings: %+v", got)
	}

	updated, err := svc.UpdateSettings(ctx, UpdateAttendanceTelegramSettingsInput{IsEnabled: true, SendTimes: []string{" 06:30 ", "17:00", "06:30"}, SendDays: []int32{1, 2, 2}, TargetChatID: " 987654321 ", IncludeCaption: true})
	if err != nil {
		t.Fatalf("UpdateSettings error = %v", err)
	}
	if store.upsertArg.Timezone != attendanceTelegramDefaultTimezone || store.upsertArg.TargetChatID != "987654321" || strings.Join(store.upsertArg.SendTimes, ",") != "06:30,17:00" || len(store.upsertArg.SendDays) != 2 || store.upsertArg.SendDays[0] != 1 || store.upsertArg.SendDays[1] != 2 {
		t.Fatalf("unexpected upsert arg: %+v", store.upsertArg)
	}
	if !updated.IsEnabled || updated.SendTime != "06:30" || updated.TargetChatIDMasked != "*****4321" || updated.ReportMode != "ringkas" {
		t.Fatalf("unexpected update response: %+v", updated)
	}

	invalidInputs := []UpdateAttendanceTelegramSettingsInput{
		{Timezone: "UTC", IncludeCaption: true},
		{ReportMode: "detail", IncludeCaption: true},
		{IncludeCaption: false, IncludeImage: false},
		{IsEnabled: true, IncludeCaption: true},
		{TargetChatID: "1", IncludeCaption: true, SendTimes: []string{"25:00"}},
		{TargetChatID: "1", IncludeCaption: true, SendDays: []int32{7}},
	}
	for _, input := range invalidInputs {
		if _, err := svc.UpdateSettings(ctx, input); err == nil {
			t.Fatalf("UpdateSettings(%+v) expected error", input)
		}
	}
}

func TestPusakaAttendanceTelegramListLogsAndBuildReport(t *testing.T) {
	ctx := context.Background()
	store := &fakeAttendanceTelegramStore{
		listLogs: []db.ListPusakaAttendanceTelegramLogsRow{{TargetChatIDMasked: "***1234", Status: "success"}},
		reportRows: []db.ListPusakaAttendanceTelegramReportRowsRow{
			{EmployeeNama: "Andi", EmployeeNip: "1", JamMasuk: "07:00:01 WITA", JamPulang: "15:10"},
			{EmployeeNama: "Budi", EmployeeNip: "2", JamMasuk: "08:05", JamPulang: ""},
			{EmployeeNama: "Cici", EmployeeNip: "3", JamMasuk: "", JamPulang: ""},
		},
	}
	svc := newTelegramTestService(store, "token", nil)

	logs, err := svc.ListLogs(ctx, 25, 5)
	if err != nil || len(logs) != 1 || store.listLogsArg.Limit != 25 || store.listLogsArg.Offset != 5 {
		t.Fatalf("ListLogs() logs=%+v arg=%+v err=%v", logs, store.listLogsArg, err)
	}

	report, err := svc.buildReport(ctx, time.Date(2026, 5, 17, 0, 0, 0, 0, svc.loc))
	if err != nil {
		t.Fatalf("buildReport error = %v", err)
	}
	if report.TotalEmployees != 3 || report.CheckedIn != 2 || report.NotCheckedIn != 1 || report.CheckedOut != 1 {
		t.Fatalf("unexpected report counters: %+v", report)
	}
	if report.Rows[0].Status != "Lengkap" || report.Rows[1].Status != "Masuk" || report.Rows[2].Status != "Belum" || report.Rows[0].CheckIn != "07:00:01" {
		t.Fatalf("unexpected report rows: %+v", report.Rows)
	}
}

func TestPusakaAttendanceTelegramSendReportSuccessAndFailure(t *testing.T) {
	ctx := context.Background()
	store := &fakeAttendanceTelegramStore{settingsErr: pgx.ErrNoRows, reportRows: []db.ListPusakaAttendanceTelegramReportRowsRow{{EmployeeNama: "Andi", JamMasuk: "07:00", JamPulang: "15:00"}}}
	var paths []string
	var form map[string][]string
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		paths = append(paths, req.URL.Path)
		if err := req.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		form = req.MultipartForm.Value
		if strings.HasSuffix(req.URL.Path, "/sendPhoto") && len(req.MultipartForm.File["photo"]) != 1 {
			t.Fatalf("sendPhoto missing uploaded photo: %#v", req.MultipartForm.File)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":88}}`))}, nil
	})
	svc := newTelegramTestService(store, "token", rt)

	res, err := svc.SendReport(ctx, SendAttendanceTelegramReportInput{Date: "2026-05-17", TargetChatID: "99999", IncludeCaption: true, IncludeImage: false})
	if err != nil {
		t.Fatalf("SendReport text error = %v", err)
	}
	if res.Status != "success" || res.TelegramMessageID != "88" || res.TargetChatIDMasked != "*9999" || len(store.logArgs) != 1 || store.logArgs[0].Status != "success" {
		t.Fatalf("unexpected text result=%+v logs=%+v", res, store.logArgs)
	}
	if !strings.HasSuffix(paths[0], "/sendMessage") || form["chat_id"][0] != "99999" || !strings.Contains(form["text"][0], "Daftar Hadir Ringkas") {
		t.Fatalf("unexpected text request path=%v form=%+v", paths, form)
	}

	res, err = svc.SendReport(ctx, SendAttendanceTelegramReportInput{Date: "2026-05-17", TargetChatID: "99999", IncludeCaption: true, IncludeImage: true})
	if err != nil || !strings.HasSuffix(paths[1], "/sendPhoto") || !strings.Contains(form["caption"][0], "Daftar Hadir Ringkas") || res.TelegramMessageID != "88" {
		t.Fatalf("unexpected photo result paths=%v form=%+v result=%+v err=%v", paths, form, res, err)
	}

	store = &fakeAttendanceTelegramStore{settingsErr: pgx.ErrNoRows, reportRows: []db.ListPusakaAttendanceTelegramReportRowsRow{{EmployeeNama: "Andi"}}}
	svc = newTelegramTestService(store, "", nil)
	res, err = svc.SendReport(ctx, SendAttendanceTelegramReportInput{Date: "2026-05-17", TargetChatID: "99999", IncludeCaption: true})
	if !errors.Is(err, ErrAttendanceTelegramNotConfigured) || res.Status != "failed" || len(store.logArgs) != 1 || store.logArgs[0].Status != "failed" || !store.logArgs[0].ErrorMessage.Valid {
		t.Fatalf("failure not returned/logged: result=%+v logs=%+v err=%v", res, store.logArgs, err)
	}
	if _, err := svc.SendReport(ctx, SendAttendanceTelegramReportInput{Date: "17-05-2026", IncludeCaption: true}); err == nil {
		t.Fatal("SendReport invalid date expected error")
	}
}

func TestPusakaAttendanceTelegramRunDue(t *testing.T) {
	ctx := context.Background()
	store := &fakeAttendanceTelegramStore{settings: telegramSettings(true, []string{"06:30", "17:00"}, []int32{1}), reportRows: []db.ListPusakaAttendanceTelegramReportRowsRow{{EmployeeNama: "Andi", JamMasuk: "07:00"}}}
	svc := newTelegramTestService(store, "token", nil)
	mondayDue := time.Date(2026, 5, 18, 17, 0, 0, 0, svc.loc)
	if err := svc.RunDue(ctx, mondayDue); err != nil {
		t.Fatalf("RunDue due error = %v", err)
	}
	if len(store.logArgs) != 1 || store.logArgs[0].SendMode != "scheduled" || store.logArgs[0].ScheduleTime != "17:00" || store.hasLogArg.ScheduleTime != "17:00" {
		t.Fatalf("scheduled send not logged correctly: hasArg=%+v logs=%+v", store.hasLogArg, store.logArgs)
	}

	store.logArgs = nil
	store.hasLog = true
	if err := svc.RunDue(ctx, mondayDue); err != nil || len(store.logArgs) != 0 {
		t.Fatalf("RunDue existing log err=%v logs=%+v", err, store.logArgs)
	}

	store.hasLog = false
	if err := svc.RunDue(ctx, mondayDue.Add(30*time.Minute)); err != nil || len(store.logArgs) != 0 {
		t.Fatalf("RunDue off minute err=%v logs=%+v", err, store.logArgs)
	}
}

func TestPusakaAttendanceTelegramTelegramRequestAndDoTelegramErrors(t *testing.T) {
	ctx := context.Background()
	svc := newTelegramTestService(&fakeAttendanceTelegramStore{}, "token", roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(req.URL.Path, "/sendMessage") {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		if err := req.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if req.MultipartForm.Value["chat_id"][0] != "1" || req.MultipartForm.Value["text"][0] != "hello" {
			t.Fatalf("unexpected request fields: %+v", req.MultipartForm.Value)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":{"message_id":123}}`))}, nil
	}))
	msgID, err := svc.telegramRequest(ctx, "sendMessage", map[string]string{"chat_id": "1", "text": "hello"})
	if err != nil || msgID != "123" {
		t.Fatalf("telegramRequest() msgID=%q err=%v", msgID, err)
	}

	for name, rt := range map[string]http.RoundTripper{
		"transport": roundTripFunc(func(req *http.Request) (*http.Response, error) { return nil, errors.New("boom") }),
		"http": roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader("bad"))}, nil
		}),
		"api": roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"ok":false,"description":"forbidden"}`))}, nil
		}),
		"invalid": roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`not-json`))}, nil
		}),
	} {
		svc = newTelegramTestService(&fakeAttendanceTelegramStore{}, "token", rt)
		if _, err := svc.telegramRequest(ctx, "sendMessage", map[string]string{"chat_id": "1"}); err == nil {
			t.Fatalf("%s response expected error", name)
		}
	}
}
