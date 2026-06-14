package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type PusakaJob struct {
	svc pusakaJobService
}

type pusakaJobService interface {
	List(ctx context.Context, status string, limit, offset int32) ([]db.ListJobsRow, int64, error)
	Create(ctx context.Context, employeeID pgtype.UUID, runType string, maxAttempts int32) (db.Job, error)
	Stats(ctx context.Context) (db.GetJobStatsRow, error)
	RunAll(ctx context.Context, runType string, maxAttempts int32) (inserted, skipped int, err error)
	CancelEmployee(ctx context.Context, employeeID pgtype.UUID) (int64, error)
	CancelAll(ctx context.Context) (int64, error)
}

func NewPusakaJob(svc *service.PusakaJob) *PusakaJob { return &PusakaJob{svc: svc} }

func (h *PusakaJob) List(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
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

func (h *PusakaJob) Create(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		EmployeeID  string `json:"employee_id"`
		RunType     string `json:"run_type"`
		MaxAttempts int32  `json:"max_attempts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
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

func (h *PusakaJob) Stats(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, stats)
}

func (h *PusakaJob) RunAll(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		RunType     string `json:"run_type"`
		MaxAttempts int32  `json:"max_attempts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
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

func (h *PusakaJob) CancelEmployee(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		EmployeeID string `json:"employee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
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

func (h *PusakaJob) CancelAll(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	cancelled, err := h.svc.CancelAll(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"ok": true, "cancelled": cancelled})
}

func (h *PusakaJob) SyncAttendance(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	inserted, skipped, err := h.svc.RunAll(r.Context(), "scrape", 3)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"ok":       true,
		"inserted": inserted,
		"skipped":  skipped,
		"message":  "Attendance sync jobs created",
	})
}
