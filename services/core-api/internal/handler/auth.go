package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"path/filepath"
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
	ListAccountChangeHistory(ctx context.Context, userID pgtype.UUID, limit int32) ([]service.AccountChangeHistoryItem, error)
	UpdateAccountContact(ctx context.Context, userID pgtype.UUID, patch service.AccountContactPatch) (db.GetUserAccountSummaryRow, error)
	SaveAccountAvatar(ctx context.Context, input service.UploadAccountAvatarInput) (db.GetUserAccountSummaryRow, error)
	DeleteAccountAvatar(ctx context.Context, userID pgtype.UUID) (db.GetUserAccountSummaryRow, error)
	AccountAvatarPath(ctx context.Context, userID pgtype.UUID, filename string) (string, bool, error)
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
	ID                 string                     `json:"id"`
	Username           string                     `json:"username"`
	DisplayName        string                     `json:"display_name"`
	Roles              []string                   `json:"roles"`
	ProfileType        string                     `json:"profile_type"`
	ProfileNama        string                     `json:"profile_nama"`
	PhotoURL           string                     `json:"photo_url"`
	AvatarURL          string                     `json:"avatar_url"`
	Contact            authAccountContactResponse `json:"contact"`
	EmployeeID         string                     `json:"employee_id,omitempty"`
	StudentID          string                     `json:"student_id,omitempty"`
	ParentID           string                     `json:"parent_id,omitempty"`
	IsActive           bool                       `json:"is_active"`
	MustChangePassword bool                       `json:"must_change_password"`
	PasswordChangedAt  string                     `json:"password_changed_at,omitempty"`
	LastLoginAt        string                     `json:"last_login_at,omitempty"`
	CreatedAt          string                     `json:"created_at,omitempty"`
}

type authAccountContactResponse struct {
	Phone          string   `json:"phone"`
	Email          string   `json:"email,omitempty"`
	Address        string   `json:"address"`
	EditableFields []string `json:"editable_fields"`
}

type authAccountChangeHistoryResponse struct {
	Action              string `json:"action"`
	FieldKey            string `json:"field_key"`
	Status              string `json:"status"`
	CreatedAt           string `json:"created_at"`
	ReviewerUsername    string `json:"reviewer_username,omitempty"`
	ReviewerDisplayName string `json:"reviewer_display_name,omitempty"`
	ReviewNote          string `json:"review_note,omitempty"`
}

const maxAccountAvatarUploadBytes int64 = 2 * 1024 * 1024

var accountAvatarUploadTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

type authAuditWriter interface {
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

func NewAuth(svc authService, audit authAuditWriter) *Auth { return &Auth{svc: svc, audit: audit} }

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
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
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
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
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
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

	api.OK(w, authAccountResponseFromRow(row))
}

func (h *Auth) GetAccountChangeHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentAccountUserID(w, r)
	if !ok {
		return
	}
	limit := int32(pageSize(r.URL.Query().Get("per_page"), 30))

	items, err := h.svc.ListAccountChangeHistory(r.Context(), userID, limit)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}

	api.OK(w, authAccountChangeHistoryResponses(items))
}

func (h *Auth) UpdateAccountContact(w http.ResponseWriter, r *http.Request) {
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

	patch, ok := decodeAccountContactPatch(w, r)
	if !ok {
		return
	}

	row, err := h.svc.UpdateAccountContact(r.Context(), userID, patch)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if errors.Is(err, domain.ErrForbidden) {
		api.Forbidden(w)
		return
	}
	if errors.Is(err, domain.ErrBadRequest) {
		api.BadRequest(w, "kontak pribadi tidak valid")
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}

	h.auditClaimsEvent(r.Context(), "AUTH_ACCOUNT_CONTACT_UPDATE", map[string]any{
		"profile_type": row.ProfileType,
		"fields":       contactPatchFields(patch),
	})
	api.OK(w, authAccountResponseFromRow(row))
}

func (h *Auth) UploadAccountAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentAccountUserID(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 3<<20)
	if err := r.ParseMultipartForm(3 << 20); err != nil {
		api.BadRequest(w, "multipart form tidak valid (maks 2 MB)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "file wajib diisi")
		return
	}
	defer file.Close()

	validated, err := validateUploadedFile(header.Filename, file, maxAccountAvatarUploadBytes, true)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	if !accountAvatarUploadAllowed(validated) {
		api.BadRequest(w, "hanya file JPG, PNG, atau WebP yang diperbolehkan")
		return
	}

	row, err := h.svc.SaveAccountAvatar(r.Context(), service.UploadAccountAvatarInput{
		UserID:       userID,
		OriginalName: header.Filename,
		MimeType:     validated.MimeType,
		Ext:          validated.Ext,
		FileSize:     int64(len(validated.Data)),
		File:         bytes.NewReader(validated.Data),
	})
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if errors.Is(err, domain.ErrForbidden) {
		api.Forbidden(w)
		return
	}
	if errors.Is(err, domain.ErrBadRequest) {
		api.BadRequest(w, "foto profil tidak valid")
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}

	h.auditClaimsEvent(r.Context(), "AUTH_ACCOUNT_AVATAR_UPDATE", map[string]any{
		"profile_type": row.ProfileType,
		"photo_url":    row.PhotoUrl,
	})
	api.OK(w, authAccountResponseFromRow(row))
}

func (h *Auth) DeleteAccountAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentAccountUserID(w, r)
	if !ok {
		return
	}

	row, err := h.svc.DeleteAccountAvatar(r.Context(), userID)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if errors.Is(err, domain.ErrForbidden) {
		api.Forbidden(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}

	h.auditClaimsEvent(r.Context(), "AUTH_ACCOUNT_AVATAR_DELETE", map[string]any{
		"profile_type": row.ProfileType,
	})
	api.OK(w, authAccountResponseFromRow(row))
}

func (h *Auth) AccountAvatarFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.currentAccountUserID(w, r)
	if !ok {
		return
	}
	filename := chi.URLParam(r, "filename")
	path, found, err := h.svc.AccountAvatarPath(r.Context(), userID, filename)
	if errors.Is(err, domain.ErrUnauthorized) {
		api.Unauthorized(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	if !found {
		api.NotFound(w)
		return
	}
	secureFileResponseHeaders(w, mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))), filename)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	http.ServeFile(w, r, path)
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
		api.BadRequest(w, "ID sesi ujian tidak valid")
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
	api.OK(w, map[string]string{"message": "Sesi masuk dicabut"})
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
		api.BadRequest(w, "ID sesi ujian tidak valid")
		return
	}

	var body struct {
		DeviceLabel string `json:"device_label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
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
	api.OK(w, map[string]string{"message": "Label sesi masuk diperbarui"})
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
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}

	prefs, err := h.svc.UpdateSidebarPreferences(r.Context(), userID, body)
	if errors.Is(err, domain.ErrBadRequest) {
		api.BadRequest(w, "Pengaturan menu samping tidak valid")
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
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
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
		api.BadRequest(w, "password baru minimal 8 karakter dan maksimal 72 karakter, tidak boleh sama dengan username, dan tidak boleh hanya angka")
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

func (h *Auth) currentAccountUserID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	userID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	return userID, true
}

func accountAvatarUploadAllowed(upload validatedUpload) bool {
	expectedMime, ok := accountAvatarUploadTypes[strings.ToLower(upload.Ext)]
	return ok && expectedMime == upload.MimeType
}

func authAccountResponseFromRow(row db.GetUserAccountSummaryRow) authAccountResponse {
	return authAccountResponse{
		ID:          pgUUIDString(row.ID),
		Username:    row.Username,
		DisplayName: row.DisplayName,
		Roles:       stringSliceFromJSONValue(row.Roles),
		ProfileType: row.ProfileType,
		ProfileNama: row.ProfileNama,
		PhotoURL:    row.PhotoUrl,
		AvatarURL:   row.PhotoUrl,
		Contact: authAccountContactResponse{
			Phone:          row.ContactPhone,
			Email:          row.ContactEmail,
			Address:        row.ContactAddress,
			EditableFields: contactEditableFields(row.ProfileType),
		},
		EmployeeID:         pgUUIDString(row.EmployeeID),
		StudentID:          pgUUIDString(row.StudentID),
		ParentID:           pgUUIDString(row.ParentID),
		IsActive:           row.IsActive,
		MustChangePassword: row.MustChangePassword,
		PasswordChangedAt:  timestamptzRFC3339(row.PasswordChangedAt),
		LastLoginAt:        timestamptzRFC3339(row.LastLoginAt),
		CreatedAt:          timestamptzRFC3339(row.CreatedAt),
	}
}

func stringSliceFromJSONValue(value any) []string {
	out := make([]string, 0)
	var payload []byte
	switch v := value.(type) {
	case nil:
		return out
	case []byte:
		payload = v
	case string:
		payload = []byte(v)
	default:
		var err error
		payload, err = json.Marshal(v)
		if err != nil {
			return out
		}
	}
	if len(payload) == 0 {
		return out
	}
	_ = json.Unmarshal(payload, &out)
	return out
}

func authAccountChangeHistoryResponses(items []service.AccountChangeHistoryItem) []authAccountChangeHistoryResponse {
	responses := make([]authAccountChangeHistoryResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, authAccountChangeHistoryResponse{
			Action:              item.Action,
			FieldKey:            item.FieldKey,
			Status:              item.Status,
			CreatedAt:           timestamptzRFC3339(item.CreatedAt),
			ReviewerUsername:    item.ReviewerUsername,
			ReviewerDisplayName: item.ReviewerDisplayName,
			ReviewNote:          item.ReviewNote,
		})
	}
	return responses
}

func contactEditableFields(profileType string) []string {
	switch profileType {
	case "employee":
		return []string{"phone", "email", "address"}
	case "student", "parent":
		return []string{"phone", "address"}
	default:
		return []string{}
	}
}

func decodeAccountContactPatch(w http.ResponseWriter, r *http.Request) (service.AccountContactPatch, bool) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return service.AccountContactPatch{}, false
	}

	var patch service.AccountContactPatch
	for key, value := range raw {
		parsed, err := decodeContactString(value)
		if err != nil {
			api.BadRequest(w, "kontak pribadi tidak valid")
			return service.AccountContactPatch{}, false
		}
		switch key {
		case "phone":
			patch.Phone = parsed
		case "email":
			patch.Email = parsed
		case "address":
			patch.Address = parsed
		default:
			api.BadRequest(w, "field "+key+" tidak dapat diubah dari akun saya")
			return service.AccountContactPatch{}, false
		}
	}
	return patch, true
}

func decodeContactString(raw json.RawMessage) (*string, error) {
	if strings.TrimSpace(string(raw)) == "null" {
		empty := ""
		return &empty, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

func contactPatchFields(patch service.AccountContactPatch) []string {
	fields := make([]string, 0, 3)
	if patch.Phone != nil {
		fields = append(fields, "phone")
	}
	if patch.Email != nil {
		fields = append(fields, "email")
	}
	if patch.Address != nil {
		fields = append(fields, "address")
	}
	return fields
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
