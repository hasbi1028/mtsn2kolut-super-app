package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type PusakaAttendance struct {
	svc *service.PusakaAttendance
}

func NewPusakaAttendance(svc *service.PusakaAttendance) *PusakaAttendance { return &PusakaAttendance{svc: svc} }

func (h *PusakaAttendance) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	startStr := q.Get("start_date")
	endStr := q.Get("end_date")

	if startStr != "" && endStr != "" {
		var start, end pgtype.Date
		if err := start.Scan(startStr); err != nil {
			api.BadRequest(w, "invalid start_date")
			return
		}
		if err := end.Scan(endStr); err != nil {
			api.BadRequest(w, "invalid end_date")
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
	dateStr := chi.URLParam(r, "date")
	var d pgtype.Date
	if err := d.Scan(dateStr); err != nil {
		api.BadRequest(w, "invalid date, use YYYY-MM-DD")
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
	q := r.URL.Query()
	startStr := q.Get("start_date")
	endStr := q.Get("end_date")

	if startStr == "" || endStr == "" {
		api.BadRequest(w, "start_date and end_date required")
		return
	}

	var start, end pgtype.Date
	if err := start.Scan(startStr); err != nil {
		api.BadRequest(w, "invalid start_date")
		return
	}
	if err := end.Scan(endStr); err != nil {
		api.BadRequest(w, "invalid end_date")
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
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
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
