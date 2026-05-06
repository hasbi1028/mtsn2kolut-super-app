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
	svc   authService
	audit authAuditWriter
}

type authService interface {
	Login(ctx context.Context, username, password string, meta service.SessionMeta) (domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string, meta service.SessionMeta) (domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID pgtype.UUID) error
	GetAccount(ctx context.Context, userID pgtype.UUID) (db.GetUserAccountSummaryRow, error)
	ListActiveSessions(ctx context.Context, userID pgtype.UUID) ([]db.AuthSession, error)
	RevokeSession(ctx context.Context, userID, sessionID pgtype.UUID) error
	UpdateSessionLabel(ctx context.Context, userID, sessionID pgtype.UUID, deviceLabel string) error
	GetSidebarPreferences(ctx context.Context, userID pgtype.UUID) (service.SidebarPreferences, error)
	UpdateSidebarPreferences(ctx context.Context, userID pgtype.UUID, prefs service.SidebarPreferences) (service.SidebarPreferences, error)
	ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error
}

type authSessionResponse struct {
	ID          string `json:"id"`
	DeviceLabel string `json:"device_label"`
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	LastUsedAt  string `json:"last_used_at,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

type authAccountResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
	ProfileType string   `json:"profile_type"`
	ProfileNama string   `json:"profile_nama"`
	EmployeeID  string   `json:"employee_id,omitempty"`
	StudentID   string   `json:"student_id,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
	IsActive    bool     `json:"is_active"`
	LastLoginAt string   `json:"last_login_at,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
}

type authAuditWriter interface {
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

func NewAuth(svc authService, audit authAuditWriter) *Auth { return &Auth{svc: svc, audit: audit} }

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
	h.auditRefreshTokenEvent(r.Context(), "AUTH_LOGOUT", body.RefreshToken, map[string]any{
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

func (h *Auth) GetAccount(w http.ResponseWriter, r *http.Request) {
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

	row, err := h.svc.GetAccount(r.Context(), userID)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}

	roles := make([]string, 0)
	if len(row.Roles) > 0 {
		_ = json.Unmarshal(row.Roles, &roles)
	}
	api.OK(w, authAccountResponse{
		ID:          pgUUIDString(row.ID),
		Username:    row.Username,
		DisplayName: row.DisplayName,
		Roles:       roles,
		ProfileType: row.ProfileType,
		ProfileNama: row.ProfileNama,
		EmployeeID:  pgUUIDString(row.EmployeeID),
		StudentID:   pgUUIDString(row.StudentID),
		ParentID:    pgUUIDString(row.ParentID),
		IsActive:    row.IsActive,
		LastLoginAt: timestamptzRFC3339(row.LastLoginAt),
		CreatedAt:   timestamptzRFC3339(row.CreatedAt),
	})
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
	items := make([]authSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, authSessionResponse{
			ID:          pgUUIDString(session.ID),
			DeviceLabel: session.DeviceLabel,
			IPAddress:   session.IpAddress,
			UserAgent:   session.UserAgent,
			LastUsedAt:  timestamptzRFC3339(session.LastUsedAt),
			CreatedAt:   timestamptzRFC3339(session.CreatedAt),
			ExpiresAt:   timestamptzRFC3339(session.ExpiresAt),
		})
	}
	api.OK(w, items)
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

func (h *Auth) GetSidebarPreferences(w http.ResponseWriter, r *http.Request) {
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

	prefs, err := h.svc.GetSidebarPreferences(r.Context(), userID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, prefs)
}

func (h *Auth) UpdateSidebarPreferences(w http.ResponseWriter, r *http.Request) {
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

	var body service.SidebarPreferences
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}

	prefs, err := h.svc.UpdateSidebarPreferences(r.Context(), userID, body)
	if errors.Is(err, domain.ErrBadRequest) {
		api.BadRequest(w, "invalid sidebar preferences")
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}

	h.auditClaimsEvent(r.Context(), "AUTH_SIDEBAR_PREFS_UPDATE", map[string]any{
		"pinned_count": len(prefs.PinnedItems),
		"recent_count": len(prefs.RecentItems),
	})
	api.OK(w, prefs)
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
	h.auditClaimsEvent(r.Context(), "AUTH_PASSWORD_CHANGE", map[string]any{
		"scope": "current_user",
	})
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

	return service.SessionMeta{
		IPAddress: trustedClientIP(r),
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

func (h *Auth) auditRefreshTokenEvent(ctx context.Context, action, refreshToken string, extra map[string]any) {
	if h.audit == nil {
		return
	}
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return
	}
	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(refreshToken, claims); err != nil {
		return
	}
	if tokenType, _ := claims["type"].(string); tokenType != "refresh" {
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
