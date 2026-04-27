package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Setting struct {
	svc *service.Setting
}

func NewSetting(svc *service.Setting) *Setting { return &Setting{svc: svc} }

func (h *Setting) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Setting) Upsert(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	// block auth keys from being set via this endpoint
	if key == "admin_password" || key == "admin_username" {
		api.Forbidden(w)
		return
	}
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if err := h.svc.Upsert(r.Context(), key, body.Value); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"key": key, "value": body.Value})
}
