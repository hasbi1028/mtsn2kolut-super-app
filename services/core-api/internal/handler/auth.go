package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/service"
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
	pair, err := h.svc.Login(r.Context(), body.Username, body.Password)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if errors.Is(err, domain.ErrSuspended) {
		api.Err(w, http.StatusForbidden, "account is suspended")
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, pair)
}

func (h *Auth) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		api.BadRequest(w, "refresh_token required")
		return
	}
	pair, err := h.svc.Refresh(r.Context(), body.RefreshToken)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if errors.Is(err, domain.ErrSuspended) {
		api.Err(w, http.StatusForbidden, "account is suspended")
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, pair)
}

func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := h.svc.Logout(r.Context(), body.RefreshToken); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"message": "logged out"})
}

func (h *Auth) LogoutAll(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return
	}

	if err := h.svc.LogoutAll(r.Context(), userID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			api.Unauthorized(w)
			return
		}
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"message": "all sessions logged out"})
}

func (h *Auth) ListSessions(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return
	}

	sessions, err := h.svc.ListActiveSessions(r.Context(), userID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, sessions)
}

func (h *Auth) RevokeSession(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return
	}

	var sessionID pgtype.UUID
	if err := sessionID.Scan(chi.URLParam(r, "id")); err != nil {
		api.BadRequest(w, "invalid session id")
		return
	}

	err = h.svc.RevokeSession(r.Context(), userID, sessionID)
	if errors.Is(err, domain.ErrNotFound) {
		api.NotFound(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"message": "session revoked"})
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

	// Username from JWT claims takes priority; body username is used for internal-key requests
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if username, _ := claims["usr"].(string); username != "" {
			body.Username = username
		}
	}

	if body.Username == "" {
		api.BadRequest(w, "username required")
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
	if errors.Is(err, domain.ErrSuspended) {
		api.Err(w, http.StatusForbidden, "account is suspended")
		return
	}
	if errors.Is(err, domain.ErrWeakPassword) {
		api.BadRequest(w, "password baru minimal 8 karakter, tidak boleh sama dengan username, dan tidak boleh hanya angka")
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]string{"message": "password changed"})
}

func authUserID(claims map[string]any) (pgtype.UUID, error) {
	var userID pgtype.UUID
	raw, _ := claims["sub"].(string)
	if raw == "" {
		raw, _ = claims["uid"].(string)
	}
	if raw == "" {
		return pgtype.UUID{}, domain.ErrUnauthorized
	}
	if err := userID.Scan(raw); err != nil {
		return pgtype.UUID{}, domain.ErrUnauthorized
	}
	return userID, nil
}
