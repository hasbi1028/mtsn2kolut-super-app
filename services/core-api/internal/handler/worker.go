package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type pusakaWorkerJobService interface {
	Claim(ctx context.Context, workerID string) (db.ClaimJobRow, error)
	Complete(ctx context.Context, id pgtype.UUID, workerID string) error
	CompleteWithAttendance(ctx context.Context, id pgtype.UUID, workerID string, attendance *service.PusakaJobAttendanceInput) error
	Fail(ctx context.Context, id pgtype.UUID, workerID string, errorMessage string, retryAfter pgtype.Text) error
	Stats(ctx context.Context) (db.GetJobStatsRow, error)
}

type pusakaWorkerAttendanceService interface {
	Upsert(ctx context.Context, p db.UpsertAttendanceParams) (db.AttendanceRecord, error)
}

type PusakaWorker struct {
	jobs pusakaWorkerJobService
	att  pusakaWorkerAttendanceService
	sett settingService
}

func NewPusakaWorker(jobs *service.PusakaJob, att *service.PusakaAttendance, sett *service.Setting) *PusakaWorker {
	return &PusakaWorker{jobs: jobs, att: att, sett: sett}
}

func (h *PusakaWorker) Claim(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
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

func (h *PusakaWorker) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	// Optional attendance data in body
	var body struct {
		WorkerID  string `json:"worker_id"`
		Tanggal   string `json:"tanggal"`
		JamMasuk  string `json:"jam_masuk"`
		JamPulang string `json:"jam_pulang"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	workerID := workerIDFromRequest(r, body.WorkerID)
	if workerID == "" {
		api.BadRequest(w, "worker_id required")
		return
	}

	var attendance *service.PusakaJobAttendanceInput
	if body.Tanggal != "" {
		var tanggal pgtype.Date
		if err := tanggal.Scan(body.Tanggal); err != nil {
			api.BadRequest(w, "invalid tanggal")
			return
		}
		attendance = &service.PusakaJobAttendanceInput{
			Tanggal:   tanggal,
			JamMasuk:  body.JamMasuk,
			JamPulang: body.JamPulang,
		}
	}

	if err := h.jobs.CompleteWithAttendance(r.Context(), id, workerID, attendance); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			api.Conflict(w, "job is not running for this worker")
			return
		}
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *PusakaWorker) Fail(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		WorkerID       string          `json:"worker_id"`
		Error          string          `json:"error"`
		RetryAfterSecs json.RawMessage `json:"retry_after_secs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	retryAfterSecs, err := parseWorkerRetryAfterSecs(body.RetryAfterSecs)
	if err != nil {
		api.BadRequest(w, "invalid retry_after_secs")
		return
	}
	if retryAfterSecs == "" {
		retryAfterSecs = "60"
	}
	workerID := workerIDFromRequest(r, body.WorkerID)
	if workerID == "" {
		api.BadRequest(w, "worker_id required")
		return
	}
	retryAfter := pgtype.Text{String: retryAfterSecs, Valid: true}
	if err := h.jobs.Fail(r.Context(), id, workerID, body.Error, retryAfter); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			api.Conflict(w, "job is not running for this worker")
			return
		}
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func parseWorkerRetryAfterSecs(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return "", nil
	}
	var rawValue string
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		rawValue = strings.TrimSpace(text)
	} else {
		var number json.Number
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&number); err != nil {
			return "", errors.New("retry_after_secs must be string or number")
		}
		rawValue = number.String()
	}
	if rawValue == "" {
		return "", nil
	}
	retryAfter, err := strconv.Atoi(rawValue)
	if err != nil || retryAfter < 0 || retryAfter > 86400 {
		return "", errors.New("retry_after_secs must be an integer between 0 and 86400")
	}
	return strconv.Itoa(retryAfter), nil
}

func workerIDFromRequest(r *http.Request, bodyWorkerID string) string {
	if workerID := strings.TrimSpace(bodyWorkerID); workerID != "" {
		return workerID
	}
	return strings.TrimSpace(r.Header.Get("X-Worker-ID"))
}

func (h *PusakaWorker) UpsertAttendance(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct {
		EmployeeID  string `json:"employee_id"`
		Tanggal     string `json:"tanggal"`
		JamMasuk    string `json:"jam_masuk"`
		JamPulang   string `json:"jam_pulang"`
		SourceJobID string `json:"source_job_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
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

func (h *PusakaWorker) Config(w http.ResponseWriter, r *http.Request) {
	rows, err := h.sett.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}

	cfg := map[string]string{}
	for _, row := range rows {
		switch row.Key {
		case "max_concurrent",
			"headless",
			"pusaka_geo_base_lat",
			"pusaka_geo_base_lng",
			"pusaka_geo_default_radius_m",
			"pusaka_geo_checkin_radius_m",
			"pusaka_geo_checkout_radius_m":
			cfg[row.Key] = row.Value
		}
	}

	api.OK(w, cfg)
}

func (h *PusakaWorker) Heartbeat(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct {
		WorkerID         string `json:"worker_id"`
		ActiveConsumers  int    `json:"active_consumers"`
		TargetConcurrent int    `json:"target_concurrency"`
		Headless         bool   `json:"headless"`
		LastSyncAt       string `json:"last_sync_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.WorkerID == "" {
		api.BadRequest(w, "worker_id required")
		return
	}

	payload := map[string]any{
		"worker_id":          body.WorkerID,
		"active_consumers":   body.ActiveConsumers,
		"target_concurrency": body.TargetConcurrent,
		"headless":           body.Headless,
		"last_sync_at":       body.LastSyncAt,
		"reported_at":        time.Now().UTC().Format(time.RFC3339),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		api.Internal(w, err)
		return
	}
	if err := h.sett.Upsert(r.Context(), "worker_status:"+body.WorkerID, string(raw)); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]bool{"ok": true})
}

func (h *PusakaWorker) GetStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.sett.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	stats, err := h.jobs.Stats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}

	activeWorkers := make([]map[string]any, 0)
	cutoff := time.Now().UTC().Add(-2 * time.Minute)

	for _, row := range rows {
		if !strings.HasPrefix(row.Key, "worker_status:") {
			continue
		}

		var workerData map[string]any
		if err := json.Unmarshal([]byte(row.Value), &workerData); err != nil {
			continue
		}

		// Check if worker is still active
		if reportedAtStr, ok := workerData["reported_at"].(string); ok {
			reportedAt, err := time.Parse(time.RFC3339, reportedAtStr)
			if err != nil || reportedAt.Before(cutoff) {
				continue
			}
		} else {
			continue
		}

		activeWorkers = append(activeWorkers, workerData)
	}

	api.OK(w, map[string]any{
		"active_workers": activeWorkers,
		"total":          len(activeWorkers),
		"queue": map[string]int64{
			"queued":  stats.Queued,
			"running": stats.Running,
			"success": stats.Success,
			"failed":  stats.Failed,
		},
		"last_checked": time.Now().UTC(),
	})
}
