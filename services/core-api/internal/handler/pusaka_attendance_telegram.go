package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type PusakaAttendanceTelegram struct {
	svc pusakaAttendanceTelegramService
}

type pusakaAttendanceTelegramService interface {
	GetSettings(ctx context.Context) (service.AttendanceTelegramSettingsResponse, error)
	UpdateSettings(ctx context.Context, in service.UpdateAttendanceTelegramSettingsInput) (service.AttendanceTelegramSettingsResponse, error)
	ListLogs(ctx context.Context, limit, offset int32) ([]db.ListPusakaAttendanceTelegramLogsRow, error)
	SendReport(ctx context.Context, in service.SendAttendanceTelegramReportInput) (service.AttendanceTelegramReportResult, error)
	RunDue(ctx context.Context, now time.Time) error
}

func NewPusakaAttendanceTelegram(svc pusakaAttendanceTelegramService) *PusakaAttendanceTelegram {
	return &PusakaAttendanceTelegram{svc: svc}
}

func (h *PusakaAttendanceTelegram) GetSettings(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.GetSettings(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, out)
}

func (h *PusakaAttendanceTelegram) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var in service.UpdateAttendanceTelegramSettingsInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		api.BadRequest(w, "payload tidak valid")
		return
	}
	out, err := h.svc.UpdateSettings(r.Context(), in)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, out)
}

func (h *PusakaAttendanceTelegram) ListLogs(w http.ResponseWriter, r *http.Request) {
	limit := int32(20)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	rows, err := h.svc.ListLogs(r.Context(), limit, offset)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *PusakaAttendanceTelegram) SendNow(w http.ResponseWriter, r *http.Request) {
	var in service.SendAttendanceTelegramReportInput
	_ = json.NewDecoder(r.Body).Decode(&in)
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if raw, ok := claims["uid"].(string); ok && raw != "" {
			_ = in.RequestedBy.Scan(raw)
		}
	}
	if !in.RequestedBy.Valid {
		in.RequestedBy = pgtype.UUID{}
	}
	in.SendMode = "manual"
	out, err := h.svc.SendReport(r.Context(), in)
	if err != nil {
		if errors.Is(err, service.ErrAttendanceTelegramNotConfigured) {
			api.BadRequest(w, "Telegram belum dikonfigurasi di backend (TELEGRAM_BOT_TOKEN/target chat)")
			return
		}
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, out)
}

func (h *PusakaAttendanceTelegram) Tick(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.RunDue(r.Context(), timeNow()); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]bool{"ok": true})
}

var timeNow = func() time.Time { return time.Now() }
