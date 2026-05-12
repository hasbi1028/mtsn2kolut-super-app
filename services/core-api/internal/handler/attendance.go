package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type PusakaAttendance struct {
	svc pusakaAttendanceService
}

func NewPusakaAttendance(svc *service.PusakaAttendance) *PusakaAttendance {
	return &PusakaAttendance{svc: svc}
}

type pusakaAttendanceService interface {
	List(ctx context.Context, limit, offset int32) ([]db.ListAttendanceRow, int64, error)
	ListInRange(ctx context.Context, start, end pgtype.Date) ([]db.ListAttendanceInRangeRow, error)
	ByDate(ctx context.Context, date pgtype.Date) ([]db.ListAttendanceByDateRow, error)
	GetSummary(ctx context.Context, start, end pgtype.Date) ([]db.GetMonthlyAttendanceSummaryRow, error)
	ByEmployee(ctx context.Context, empID pgtype.UUID, limit, offset int32) ([]db.ListAttendanceByEmployeeRow, error)
}

func (h *PusakaAttendance) List(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	startStr := q.Get("start_date")
	endStr := q.Get("end_date")

	if startStr != "" && endStr != "" {
		var start, end pgtype.Date
		if err := start.Scan(startStr); err != nil {
			api.BadRequest(w, "Tanggal awal tidak valid")
			return
		}
		if err := end.Scan(endStr); err != nil {
			api.BadRequest(w, "Tanggal akhir tidak valid")
			return
		}
		rows, err := h.svc.ListInRange(r.Context(), start, end)
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.OK(w, rows)
		return
	}

	limit := int32(pageSize(q.Get("per_page"), 50))
	page := pageNum(q.Get("page"), 1)
	offset := int32((page - 1) * int(limit))

	rows, total, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.JSON(w, http.StatusOK, api.PagedResponse{
		Data: rows,
		Meta: api.PageMeta{Total: total, Page: page, PerPage: int(limit)},
	})
}

func (h *PusakaAttendance) ByDate(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	dateStr := chi.URLParam(r, "date")
	var d pgtype.Date
	if err := d.Scan(dateStr); err != nil {
		api.BadRequest(w, "Tanggal tidak valid, gunakan format YYYY-MM-DD")
		return
	}
	rows, err := h.svc.ByDate(r.Context(), d)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *PusakaAttendance) GetSummary(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	startStr := q.Get("start_date")
	endStr := q.Get("end_date")

	if startStr == "" || endStr == "" {
		api.BadRequest(w, "start_date and end_date required")
		return
	}

	var start, end pgtype.Date
	if err := start.Scan(startStr); err != nil {
		api.BadRequest(w, "Tanggal awal tidak valid")
		return
	}
	if err := end.Scan(endStr); err != nil {
		api.BadRequest(w, "Tanggal akhir tidak valid")
		return
	}

	rows, err := h.svc.GetSummary(r.Context(), start, end)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *PusakaAttendance) ByEmployee(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	q := r.URL.Query()
	limit := int32(pageSize(q.Get("per_page"), 50))
	page := pageNum(q.Get("page"), 1)
	offset := int32((page - 1) * int(limit))

	rows, err := h.svc.ByEmployee(r.Context(), id, limit, offset)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}
