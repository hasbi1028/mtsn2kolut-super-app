package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakePusakaAttendanceTelegramService struct {
	settings service.AttendanceTelegramSettingsResponse
	result   service.AttendanceTelegramReportResult
	logs     []db.ListPusakaAttendanceTelegramLogsRow
	err      error

	updateIn   service.UpdateAttendanceTelegramSettingsInput
	listLimit  int32
	listOffset int32
	sendIn     service.SendAttendanceTelegramReportInput
	runDueAt   time.Time
}

func (f *fakePusakaAttendanceTelegramService) GetSettings(context.Context) (service.AttendanceTelegramSettingsResponse, error) {
	return f.settings, f.err
}
func (f *fakePusakaAttendanceTelegramService) UpdateSettings(_ context.Context, in service.UpdateAttendanceTelegramSettingsInput) (service.AttendanceTelegramSettingsResponse, error) {
	f.updateIn = in
	return f.settings, f.err
}
func (f *fakePusakaAttendanceTelegramService) ListLogs(_ context.Context, limit, offset int32) ([]db.ListPusakaAttendanceTelegramLogsRow, error) {
	f.listLimit, f.listOffset = limit, offset
	return f.logs, f.err
}
func (f *fakePusakaAttendanceTelegramService) SendReport(_ context.Context, in service.SendAttendanceTelegramReportInput) (service.AttendanceTelegramReportResult, error) {
	f.sendIn = in
	return f.result, f.err
}
func (f *fakePusakaAttendanceTelegramService) RunDue(_ context.Context, now time.Time) error {
	f.runDueAt = now
	return f.err
}

func TestPusakaAttendanceTelegramHandlers(t *testing.T) {
	userID := handlerTestUUID(211)
	fake := &fakePusakaAttendanceTelegramService{
		settings: service.AttendanceTelegramSettingsResponse{IsEnabled: true, SendTime: "17:00", SendTimes: []string{"17:00"}, SendDays: []int32{1, 2, 3, 4, 5, 6}, Timezone: "Asia/Makassar", TargetChatID: "1450267717", TargetChatIDMasked: "******7717", IncludeCaption: true, IncludeImage: true, ReportMode: "ringkas", BotConfigured: true},
		result:   service.AttendanceTelegramReportResult{ReportDate: "2026-05-17", TargetChatIDMasked: "******7717", SendMode: "manual", Status: "success", TotalEmployees: 3, CheckedIn: 2},
	}
	h := NewPusakaAttendanceTelegram(fake)

	rec := httptest.NewRecorder()
	h.GetSettings(rec, httptest.NewRequest(http.MethodGet, "/api/pusaka/attendance-telegram/settings", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "17:00") {
		t.Fatalf("GetSettings status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.UpdateSettings(rec, httptest.NewRequest(http.MethodPut, "/api/pusaka/attendance-telegram/settings", strings.NewReader(`{"is_enabled":true,"send_time":"17:30","send_times":["17:30"],"send_days":[1,2],"timezone":"Asia/Makassar","target_chat_id":"123","include_caption":true,"include_image":false,"report_mode":"ringkas"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateSettings status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !fake.updateIn.IsEnabled || fake.updateIn.SendTime != "17:30" || fake.updateIn.TargetChatID != "123" || fake.updateIn.IncludeImage {
		t.Fatalf("UpdateSettings input = %+v", fake.updateIn)
	}

	rec = httptest.NewRecorder()
	h.ListLogs(rec, httptest.NewRequest(http.MethodGet, "/api/pusaka/attendance-telegram/logs?limit=5&offset=2", nil))
	if rec.Code != http.StatusOK || fake.listLimit != 5 || fake.listOffset != 2 {
		t.Fatalf("ListLogs status=%d limit=%d offset=%d body=%s", rec.Code, fake.listLimit, fake.listOffset, rec.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/api/pusaka/attendance-telegram/send", strings.NewReader(`{"date":"2026-05-17","target_chat_id":"123","include_caption":true}`))
	req = withClaims(req, jwt.MapClaims{"uid": userID.String()})
	rec = httptest.NewRecorder()
	h.SendNow(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("SendNow status=%d body=%s", rec.Code, rec.Body.String())
	}
	if fake.sendIn.Date != "2026-05-17" || fake.sendIn.TargetChatID != "123" || fake.sendIn.SendMode != "manual" || fake.sendIn.RequestedBy != userID {
		t.Fatalf("SendNow input = %+v", fake.sendIn)
	}

	fixedNow := time.Date(2026, 5, 17, 9, 30, 0, 0, time.UTC)
	oldNow := timeNow
	timeNow = func() time.Time { return fixedNow }
	defer func() { timeNow = oldNow }()
	rec = httptest.NewRecorder()
	h.Tick(rec, httptest.NewRequest(http.MethodPost, "/api/pusaka/attendance-telegram/tick", nil))
	if rec.Code != http.StatusOK || !fake.runDueAt.Equal(fixedNow) {
		t.Fatalf("Tick status=%d runDueAt=%v body=%s", rec.Code, fake.runDueAt, rec.Body.String())
	}
}

func TestPusakaAttendanceTelegramHandlerErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func(*PusakaAttendanceTelegram, http.ResponseWriter, *http.Request)
		req  *http.Request
		err  error
		want int
	}{
		{"get internal", (*PusakaAttendanceTelegram).GetSettings, httptest.NewRequest(http.MethodGet, "/settings", nil), errors.New("db down"), http.StatusInternalServerError},
		{"update invalid json", (*PusakaAttendanceTelegram).UpdateSettings, httptest.NewRequest(http.MethodPut, "/settings", strings.NewReader(`{`)), nil, http.StatusBadRequest},
		{"update validation", (*PusakaAttendanceTelegram).UpdateSettings, httptest.NewRequest(http.MethodPut, "/settings", strings.NewReader(`{"is_enabled":true}`)), errors.New("target wajib"), http.StatusBadRequest},
		{"logs internal", (*PusakaAttendanceTelegram).ListLogs, httptest.NewRequest(http.MethodGet, "/logs?limit=999&offset=-1", nil), errors.New("db down"), http.StatusInternalServerError},
		{"send not configured", (*PusakaAttendanceTelegram).SendNow, httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(`{"date":"2026-05-17"}`)), service.ErrAttendanceTelegramNotConfigured, http.StatusBadRequest},
		{"send validation", (*PusakaAttendanceTelegram).SendNow, httptest.NewRequest(http.MethodPost, "/send", strings.NewReader(`{`)), errors.New("tanggal tidak valid"), http.StatusBadRequest},
		{"tick internal", (*PusakaAttendanceTelegram).Tick, httptest.NewRequest(http.MethodPost, "/tick", nil), errors.New("db down"), http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakePusakaAttendanceTelegramService{err: tc.err}
			rec := httptest.NewRecorder()
			tc.fn(NewPusakaAttendanceTelegram(fake), rec, tc.req)
			if rec.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, tc.want, rec.Body.String())
			}
			if tc.name == "logs internal" && (fake.listLimit != 20 || fake.listOffset != 0) {
				t.Fatalf("ListLogs fallback limit/offset = %d/%d", fake.listLimit, fake.listOffset)
			}
		})
	}
}
