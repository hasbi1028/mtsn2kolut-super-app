package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pusaka/backend/internal/api"
	"github.com/pusaka/backend/internal/domain"
	db "github.com/pusaka/backend/internal/repository/postgres"
	"github.com/pusaka/backend/internal/service"
)

type Worker struct {
	jobs *service.Job
	att  *service.Attendance
}

func NewWorker(jobs *service.Job, att *service.Attendance) *Worker {
	return &Worker{jobs: jobs, att: att}
}

func (h *Worker) Claim(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WorkerID string `json:"worker_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.WorkerID == "" {
		api.BadRequest(w, "worker_id required")
		return
	}
	job, err := h.jobs.Claim(r.Context(), body.WorkerID)
	if errors.Is(err, domain.ErrNoJob) {
		api.JSON(w, http.StatusNoContent, nil)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, job)
}

func (h *Worker) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}

	// Optional attendance data in body
	var body struct {
		Tanggal   string `json:"tanggal"`
		JamMasuk  string `json:"jam_masuk"`
		JamPulang string `json:"jam_pulang"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if body.Tanggal != "" {
		// Need employee_id from job for attendance upsert
		job, err := h.jobs.Get(r.Context(), id)
		if err != nil {
			api.Internal(w, err)
			return
		}
		var tanggal pgtype.Date
		if err := tanggal.Scan(body.Tanggal); err != nil {
			api.BadRequest(w, "invalid tanggal")
			return
		}
		if _, err := h.att.Upsert(r.Context(), db.UpsertAttendanceParams{
			EmployeeID:  job.EmployeeID,
			Tanggal:     tanggal,
			JamMasuk:    body.JamMasuk,
			JamPulang:   body.JamPulang,
			SourceJobID: id,
		}); err != nil {
			api.Internal(w, err)
			return
		}
	}

	if err := h.jobs.Complete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *Worker) Fail(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Error          string `json:"error"`
		RetryAfterSecs string `json:"retry_after_secs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.RetryAfterSecs == "" {
		body.RetryAfterSecs = "60"
	}
	retryAfter := pgtype.Text{String: body.RetryAfterSecs, Valid: true}
	if err := h.jobs.Fail(r.Context(), id, body.Error, retryAfter); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *Worker) UpsertAttendance(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID  string `json:"employee_id"`
		Tanggal     string `json:"tanggal"`
		JamMasuk    string `json:"jam_masuk"`
		JamPulang   string `json:"jam_pulang"`
		SourceJobID string `json:"source_job_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	empID, err := parseUUID(body.EmployeeID)
	if err != nil {
		api.BadRequest(w, "invalid employee_id")
		return
	}
	var tanggal pgtype.Date
	if err := tanggal.Scan(body.Tanggal); err != nil {
		api.BadRequest(w, "invalid tanggal, use YYYY-MM-DD")
		return
	}
	var sourceJobID pgtype.UUID
	if body.SourceJobID != "" {
		if err := sourceJobID.Scan(body.SourceJobID); err != nil {
			api.BadRequest(w, "invalid source_job_id")
			return
		}
	}
	record, err := h.att.Upsert(r.Context(), db.UpsertAttendanceParams{
		EmployeeID:  empID,
		Tanggal:     tanggal,
		JamMasuk:    body.JamMasuk,
		JamPulang:   body.JamPulang,
		SourceJobID: sourceJobID,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, record)
}
