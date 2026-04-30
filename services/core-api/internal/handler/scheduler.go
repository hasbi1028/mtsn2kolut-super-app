package handler

import (
	"net/http"
	"time"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type PusakaScheduler struct {
	svc *service.PusakaScheduler
}

func NewPusakaScheduler(svc *service.PusakaScheduler) *PusakaScheduler { return &PusakaScheduler{svc: svc} }

func (h *PusakaScheduler) Tick(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.Tick(r.Context(), time.Now())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, result)
}
