package handler

import (
	"context"
	"net/http"
	"time"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type pusakaSchedulerTickService interface {
	Tick(ctx context.Context, now time.Time) (service.PusakaSchedulerResult, error)
}

type PusakaScheduler struct {
	svc pusakaSchedulerTickService
}

func NewPusakaScheduler(svc *service.PusakaScheduler) *PusakaScheduler {
	return &PusakaScheduler{svc: svc}
}

func (h *PusakaScheduler) Tick(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	result, err := h.svc.Tick(r.Context(), time.Now())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, result)
}
