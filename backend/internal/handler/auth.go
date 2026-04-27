package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pusaka/backend/internal/api"
	"github.com/pusaka/backend/internal/domain"
	"github.com/pusaka/backend/internal/service"
)

type Auth struct {
	svc *service.Auth
}

func NewAuth(svc *service.Auth) *Auth { return &Auth{svc: svc} }

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	token, err := h.svc.Login(r.Context(), body.Username, body.Password)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"token": token})
}

func (h *Auth) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.NewPassword == "" {
		api.BadRequest(w, "new_password required")
		return
	}
	err := h.svc.ChangePassword(r.Context(), body.Username, body.OldPassword, body.NewPassword)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"message": "password changed"})
}
