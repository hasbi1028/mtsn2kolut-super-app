package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Job struct {
	svc *service.Job
}

func NewJob(svc *service.Job) *Job { return &Job{svc: svc} }

func (h *Job) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	limit := int32(pageSize(q.Get("per_page"), 20))
	page := pageNum(q.Get("page"), 1)
	offset := int32((page - 1) * int(limit))

	rows, total, err := h.svc.List(r.Context(), status, limit, offset)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.JSON(w, http.StatusOK, api.PagedResponse{
		Data: rows,
		Meta: api.PageMeta{Total: total, Page: page, PerPage: int(limit)},
	})
}

func (h *Job) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID  string `json:"employee_id"`
		RunType     string `json:"run_type"`
		MaxAttempts int32  `json:"max_attempts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.MaxAttempts == 0 {
		body.MaxAttempts = 3
	}
	id, err := parseUUID(body.EmployeeID)
	if err != nil {
		api.BadRequest(w, "invalid employee_id")
		return
	}
	job, err := h.svc.Create(r.Context(), id, body.RunType, body.MaxAttempts)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			api.Conflict(w, "job queued/running already exists")
			return
		}
		api.Internal(w, err)
		return
	}
	api.Created(w, job)
}

func (h *Job) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, stats)
}

func (h *Job) RunAll(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RunType     string `json:"run_type"`
		MaxAttempts int32  `json:"max_attempts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.RunType == "" {
		body.RunType = "morning"
	}
	if body.MaxAttempts == 0 {
		body.MaxAttempts = 3
	}
	inserted, skipped, err := h.svc.RunAll(r.Context(), body.RunType, body.MaxAttempts)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, map[string]any{"inserted": inserted, "skipped": skipped})
}

func (h *Job) CancelEmployee(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID string `json:"employee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	id, err := parseUUID(body.EmployeeID)
	if err != nil {
		api.BadRequest(w, "invalid employee_id")
		return
	}
	cancelled, err := h.svc.CancelEmployee(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"ok": true, "cancelled": cancelled})
}

func (h *Job) CancelAll(w http.ResponseWriter, r *http.Request) {
	cancelled, err := h.svc.CancelAll(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"ok": true, "cancelled": cancelled})
}

func (h *Job) SyncAttendance(w http.ResponseWriter, r *http.Request) {
	inserted, skipped, err := h.svc.RunAll(r.Context(), "scrape", 3)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"ok": true,
		"inserted": inserted,
		"skipped": skipped,
		"message": "Attendance sync jobs created",
	})
}

func pageSize(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 200 {
		return def
	}
	return n
}

func pageNum(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return def
	}
	return n
}
