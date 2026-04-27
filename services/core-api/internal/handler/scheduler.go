package handler

import (
	"net/http"
	"time"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Scheduler struct {
	svc *service.Scheduler
}

func NewScheduler(svc *service.Scheduler) *Scheduler { return &Scheduler{svc: svc} }

func (h *Scheduler) Tick(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.Tick(r.Context(), time.Now())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, result)
}
