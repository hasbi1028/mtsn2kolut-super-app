package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pusaka/backend/internal/api"
	"github.com/pusaka/backend/internal/service"
)

type Attendance struct {
	svc *service.Attendance
}

func NewAttendance(svc *service.Attendance) *Attendance { return &Attendance{svc: svc} }

func (h *Attendance) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
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

func (h *Attendance) ByDate(w http.ResponseWriter, r *http.Request) {
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

func (h *Attendance) ByEmployee(w http.ResponseWriter, r *http.Request) {
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
