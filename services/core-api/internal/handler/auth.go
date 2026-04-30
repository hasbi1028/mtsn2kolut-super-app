package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Auth struct {
	svc   *service.Auth
	audit authAuditWriter
}

type authAuditWriter interface {
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

func NewAuth(svc *service.Auth, audit authAuditWriter) *Auth { return &Auth{svc: svc, audit: audit} }

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	pair, err := h.svc.Login(r.Context(), body.Username, body.Password, sessionMetaFromRequest(r))
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
	h.auditTokenPair(r.Context(), "AUTH_LOGIN", pair, map[string]any{
		"username": body.Username,
	})
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
	pair, err := h.svc.Refresh(r.Context(), body.RefreshToken, sessionMetaFromRequest(r))
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
	h.auditTokenPair(r.Context(), "AUTH_REFRESH", pair, nil)
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
	h.auditClaimsEvent(r.Context(), "AUTH_LOGOUT", map[string]any{
		"scope": "single_session",
	})
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
	h.auditClaimsEvent(r.Context(), "AUTH_LOGOUT_ALL", map[string]any{
		"scope": "all_sessions",
	})
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
	h.auditClaimsEvent(r.Context(), "AUTH_SESSION_REVOKE", map[string]any{
		"revoked_session_id": chi.URLParam(r, "id"),
	})
	api.OK(w, map[string]string{"message": "session revoked"})
}

func (h *Auth) UpdateSessionLabel(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		DeviceLabel string `json:"device_label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}

	err = h.svc.UpdateSessionLabel(r.Context(), userID, sessionID, body.DeviceLabel)
	if errors.Is(err, domain.ErrBadRequest) {
		api.BadRequest(w, "device_label required")
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		api.NotFound(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}

	h.auditClaimsEvent(r.Context(), "AUTH_SESSION_RENAME", map[string]any{
		"renamed_session_id": chi.URLParam(r, "id"),
		"device_label":       strings.TrimSpace(body.DeviceLabel),
	})
	api.OK(w, map[string]string{"message": "session label updated"})
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

func sessionMetaFromRequest(r *http.Request) service.SessionMeta {
	userAgent := strings.TrimSpace(r.Header.Get("X-Client-User-Agent"))
	if userAgent == "" {
		userAgent = strings.TrimSpace(r.UserAgent())
	}

	ipAddress := strings.TrimSpace(r.Header.Get("X-Client-IP"))
	if ipAddress == "" {
		if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
			ipAddress = strings.TrimSpace(strings.Split(forwarded, ",")[0])
		}
	}
	if ipAddress == "" {
		ipAddress = strings.TrimSpace(r.RemoteAddr)
	}

	return service.SessionMeta{
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}

func (h *Auth) auditTokenPair(ctx context.Context, action string, pair domain.TokenPair, extra map[string]any) {
	if h.audit == nil {
		return
	}
	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(pair.AccessToken, claims); err != nil {
		return
	}
	h.auditWithClaims(ctx, action, claims, extra)
}

func (h *Auth) auditClaimsEvent(ctx context.Context, action string, extra map[string]any) {
	if h.audit == nil {
		return
	}
	claims, ok := api.ClaimsFromContext(ctx)
	if !ok {
		return
	}
	h.auditWithClaims(ctx, action, claims, extra)
}

func (h *Auth) auditWithClaims(ctx context.Context, action string, claims jwt.MapClaims, extra map[string]any) {
	if h.audit == nil {
		return
	}

	var userID pgtype.UUID
	if uid, _ := claims["uid"].(string); uid != "" {
		_ = userID.Scan(uid)
	}

	entityID := ""
	if sid, _ := claims["ssid"].(string); sid != "" {
		entityID = sid
	}
	if entityID == "" {
		if uid, _ := claims["uid"].(string); uid != "" {
			entityID = uid
		}
	}

	meta := map[string]any{
		"username":   claims["usr"],
		"user_id":    claims["uid"],
		"session_id": claims["ssid"],
	}
	for key, value := range extra {
		meta[key] = value
	}
	rawMeta, err := json.Marshal(meta)
	if err != nil {
		return
	}

	_, _ = h.audit.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:     userID,
		Action:     action,
		EntityType: "auth_session",
		EntityID:   entityID,
		Metadata:   rawMeta,
	})
}
