package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const (
	accessTokenTTL  = 1 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

// authStore defines the subset of db.Queries that Auth service needs.
type authStore interface {
	GetUserByUsername(ctx context.Context, username string) (db.GetUserByUsernameRow, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.GetUserByIDRow, error)
	GetUserByStudentID(ctx context.Context, studentID pgtype.UUID) (db.GetUserByStudentIDRow, error)
	GetUserAccountSummary(ctx context.Context, id pgtype.UUID) (db.GetUserAccountSummaryRow, error)
	UpdateOwnedEmployeeContact(ctx context.Context, arg db.UpdateOwnedEmployeeContactParams) (db.UpdateOwnedEmployeeContactRow, error)
	UpdateOwnedStudentContact(ctx context.Context, arg db.UpdateOwnedStudentContactParams) (db.UpdateOwnedStudentContactRow, error)
	UpdateOwnedParentContact(ctx context.Context, arg db.UpdateOwnedParentContactParams) (db.UpdateOwnedParentContactRow, error)
	UpdateOwnedEmployeeAvatar(ctx context.Context, arg db.UpdateOwnedEmployeeAvatarParams) (string, error)
	UpdateOwnedStudentAvatar(ctx context.Context, arg db.UpdateOwnedStudentAvatarParams) (string, error)
	UpdateOwnedParentAvatar(ctx context.Context, arg db.UpdateOwnedParentAvatarParams) (string, error)
	ListOwnAccountChangeHistory(ctx context.Context, arg db.ListOwnAccountChangeHistoryParams) ([]db.ListOwnAccountChangeHistoryRow, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error)
	UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error
	ChangeUserPasswordAndInvalidate(ctx context.Context, arg db.ChangeUserPasswordAndInvalidateParams) (int32, error)
	MarkUserPasswordChanged(ctx context.Context, id pgtype.UUID) error
	MarkUserMustChangePassword(ctx context.Context, id pgtype.UUID) error
	GetUserRoles(ctx context.Context, userID pgtype.UUID) ([]db.UserRole, error)
	GetUserRoleCodesFromRbac(ctx context.Context, userID pgtype.UUID) ([]string, error)
	GetUserPermissionCodes(ctx context.Context, userID pgtype.UUID) ([]string, error)
	AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error
	AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error
	IncrementUserAuthVersion(ctx context.Context, id pgtype.UUID) (int32, error)
	MarkUserLastLogin(ctx context.Context, id pgtype.UUID) error
	CreateAuthSession(ctx context.Context, arg db.CreateAuthSessionParams) (db.AuthSession, error)
	GetAuthSession(ctx context.Context, id pgtype.UUID) (db.AuthSession, error)
	RevokeAuthSession(ctx context.Context, id pgtype.UUID) error
	RevokeLiveAuthSessionByHash(ctx context.Context, arg db.RevokeLiveAuthSessionByHashParams) (int64, error)
	TouchAuthSessionLastUsed(ctx context.Context, id pgtype.UUID) (int64, error)
	RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error)
	ListActiveAuthSessionsByUser(ctx context.Context, userID pgtype.UUID) ([]db.AuthSession, error)
	RevokeOwnedAuthSession(ctx context.Context, arg db.RevokeOwnedAuthSessionParams) (int64, error)
	UpdateOwnedAuthSessionLabel(ctx context.Context, arg db.UpdateOwnedAuthSessionLabelParams) (int64, error)
	GetUserUIPreferences(ctx context.Context, userID pgtype.UUID) (db.UserUiPreference, error)
	UpsertUserUIPreferences(ctx context.Context, arg db.UpsertUserUIPreferencesParams) (db.UserUiPreference, error)
}

type authUserRecord struct {
	ID                 pgtype.UUID
	Username           string
	PasswordHash       string
	EmployeeID         pgtype.UUID
	StudentID          pgtype.UUID
	ParentID           pgtype.UUID
	IsActive           bool
	AuthVersion        int32
	MustChangePassword bool
	Roles              []byte
}

type Auth struct {
	q             authStore
	jwtSecret     []byte
	adminPassword string
	avatarDir     string
}

type SessionMeta struct {
	IPAddress   string
	UserAgent   string
	DeviceLabel string
}

type SidebarPreferences struct {
	PinnedItems []string `json:"pinned_items"`
	RecentItems []string `json:"recent_items"`
}

type AccountContactPatch struct {
	Phone   *string
	Email   *string
	Address *string
}

type AccountChangeHistoryItem struct {
	Action              string
	FieldKey            string
	Status              string
	CreatedAt           pgtype.Timestamptz
	ReviewerUsername    string
	ReviewerDisplayName string
	ReviewNote          string
}

type UploadAccountAvatarInput struct {
	UserID       pgtype.UUID
	OriginalName string
	MimeType     string
	Ext          string
	FileSize     int64
	File         io.Reader
}

const maxAccountAvatarBytes int64 = 2 * 1024 * 1024

var allowedAccountAvatarTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

func NewAuth(q *db.Queries, jwtSecret, adminPassword string, avatarDir ...string) *Auth {
	return &Auth{q: q, jwtSecret: []byte(jwtSecret), adminPassword: adminPassword, avatarDir: normalizeAvatarDir(avatarDir...)}
}

func (s *Auth) Login(ctx context.Context, username, password string, meta SessionMeta) (domain.TokenPair, error) {
	row, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TokenPair{}, domain.ErrUnauthorized
		}
		return domain.TokenPair{}, err
	}
	user := authUserFromUsernameRow(row)

	if !user.IsActive {
		return domain.TokenPair{}, domain.ErrSuspended
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	if err := s.q.MarkUserLastLogin(ctx, user.ID); err != nil {
		return domain.TokenPair{}, err
	}

	return s.issueTokenPair(ctx, user, normalizeSessionMeta(meta))
}

func (s *Auth) LoginStudentByID(ctx context.Context, studentID pgtype.UUID, meta SessionMeta) (domain.TokenPair, error) {
	if !studentID.Valid {
		return domain.TokenPair{}, domain.ErrBadRequest
	}
	row, err := s.q.GetUserByStudentID(ctx, studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.TokenPair{}, err
	}
	user := authUserFromStudentIDRow(row)
	if !user.IsActive {
		return domain.TokenPair{}, domain.ErrSuspended
	}
	if err := s.q.MarkUserLastLogin(ctx, user.ID); err != nil {
		return domain.TokenPair{}, err
	}
	return s.issueTokenPair(ctx, user, normalizeSessionMeta(meta))
}

func (s *Auth) Refresh(ctx context.Context, refreshToken string, meta SessionMeta) (domain.TokenPair, error) {
	claims, err := s.parseTokenClaims(refreshToken)
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	if tokenType, _ := claims["type"].(string); tokenType != "refresh" {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	session, err := s.validateRefreshSession(ctx, refreshToken, claims)
	if err != nil {
		return domain.TokenPair{}, err
	}
	userID, err := userIDFromClaims(claims)
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	row, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	user := authUserFromIDRow(row)

	if !user.IsActive {
		return domain.TokenPair{}, domain.ErrSuspended
	}

	if !s.validAuthVersion(ctx, claims) {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	affected, err := s.q.RevokeLiveAuthSessionByHash(ctx, db.RevokeLiveAuthSessionByHashParams{
		ID:               session.ID,
		RefreshTokenHash: hashToken(refreshToken),
	})
	if err != nil {
		return domain.TokenPair{}, err
	}
	if affected == 0 {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	return s.issueTokenPair(ctx, user, normalizeSessionMeta(meta))
}

func (s *Auth) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}

	claims, err := s.parseTokenClaims(refreshToken)
	if err != nil {
		return nil
	}
	if tokenType, _ := claims["type"].(string); tokenType != "refresh" {
		return nil
	}

	session, err := s.validateRefreshSession(ctx, refreshToken, claims)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil
		}
		return err
	}
	return s.q.RevokeAuthSession(ctx, session.ID)
}

func (s *Auth) LogoutAll(ctx context.Context, userID pgtype.UUID) error {
	if _, err := s.q.RevokeAllAuthSessionsForUser(ctx, userID); err != nil {
		return err
	}
	_, err := s.q.IncrementUserAuthVersion(ctx, userID)
	return err
}

func (s *Auth) ListActiveSessions(ctx context.Context, userID pgtype.UUID) ([]db.AuthSession, error) {
	return s.q.ListActiveAuthSessionsByUser(ctx, userID)
}

func (s *Auth) GetAccount(ctx context.Context, userID pgtype.UUID) (db.GetUserAccountSummaryRow, error) {
	row, err := s.q.GetUserAccountSummary(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetUserAccountSummaryRow{}, domain.ErrUnauthorized
	}
	return row, err
}

func (s *Auth) ListAccountChangeHistory(ctx context.Context, userID pgtype.UUID, limit int32) ([]AccountChangeHistoryItem, error) {
	if !userID.Valid {
		return nil, domain.ErrUnauthorized
	}
	rows, err := s.q.ListOwnAccountChangeHistory(ctx, db.ListOwnAccountChangeHistoryParams{
		UserID:     userID,
		LimitCount: normalizeAccountChangeHistoryLimit(limit),
	})
	if err != nil {
		return nil, err
	}
	items := make([]AccountChangeHistoryItem, 0, len(rows))
	for _, row := range rows {
		item, ok := accountChangeHistoryItemFromRow(row)
		if ok {
			items = append(items, item)
		}
	}
	return items, nil
}

func (s *Auth) UpdateAccountContact(ctx context.Context, userID pgtype.UUID, patch AccountContactPatch) (db.GetUserAccountSummaryRow, error) {
	current, err := s.GetAccount(ctx, userID)
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}

	phone, err := normalizeContactPatchValue(patch.Phone, current.ContactPhone, 40)
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	email, err := normalizeContactPatchValue(patch.Email, current.ContactEmail, 254)
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	address, err := normalizeContactPatchValue(patch.Address, current.ContactAddress, 500)
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	if patch.Email != nil && email != "" && !looksLikeEmail(email) {
		return db.GetUserAccountSummaryRow{}, domain.ErrBadRequest
	}

	switch current.ProfileType {
	case "employee":
		_, err = s.q.UpdateOwnedEmployeeContact(ctx, db.UpdateOwnedEmployeeContactParams{
			UserID:  userID,
			Phone:   phone,
			Email:   email,
			Address: address,
		})
	case "student":
		if patch.Email != nil {
			return db.GetUserAccountSummaryRow{}, domain.ErrBadRequest
		}
		_, err = s.q.UpdateOwnedStudentContact(ctx, db.UpdateOwnedStudentContactParams{
			UserID:  userID,
			Phone:   phone,
			Address: address,
		})
	case "parent":
		if patch.Email != nil {
			return db.GetUserAccountSummaryRow{}, domain.ErrBadRequest
		}
		_, err = s.q.UpdateOwnedParentContact(ctx, db.UpdateOwnedParentContactParams{
			UserID:  userID,
			Phone:   phone,
			Address: address,
		})
	default:
		return db.GetUserAccountSummaryRow{}, domain.ErrForbidden
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetUserAccountSummaryRow{}, domain.ErrForbidden
	}
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	return s.GetAccount(ctx, userID)
}

func (s *Auth) SaveAccountAvatar(ctx context.Context, input UploadAccountAvatarInput) (db.GetUserAccountSummaryRow, error) {
	if err := validateAccountAvatar(input); err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	current, err := s.GetAccount(ctx, input.UserID)
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	if current.ProfileType != "employee" && current.ProfileType != "student" && current.ProfileType != "parent" {
		return db.GetUserAccountSummaryRow{}, domain.ErrForbidden
	}

	avatarDir := s.accountAvatarDir()
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	token, err := randomHex(16)
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	storedName := "avatar_" + token + strings.ToLower(input.Ext)
	absPath, err := filepath.Abs(filepath.Join(avatarDir, storedName))
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	out, err := os.Create(absPath)
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	copyErr := func() error {
		defer out.Close()
		_, err := io.Copy(out, input.File)
		return err
	}()
	if copyErr != nil {
		_ = os.Remove(absPath)
		return db.GetUserAccountSummaryRow{}, copyErr
	}

	photoURL := accountAvatarURL(storedName)
	if err := s.updateOwnedAccountAvatarURL(ctx, input.UserID, current.ProfileType, photoURL); err != nil {
		_ = os.Remove(absPath)
		return db.GetUserAccountSummaryRow{}, err
	}
	s.removeStoredAccountAvatar(current.PhotoUrl)
	return s.GetAccount(ctx, input.UserID)
}

func (s *Auth) DeleteAccountAvatar(ctx context.Context, userID pgtype.UUID) (db.GetUserAccountSummaryRow, error) {
	current, err := s.GetAccount(ctx, userID)
	if err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	if current.ProfileType != "employee" && current.ProfileType != "student" && current.ProfileType != "parent" {
		return db.GetUserAccountSummaryRow{}, domain.ErrForbidden
	}
	if err := s.updateOwnedAccountAvatarURL(ctx, userID, current.ProfileType, ""); err != nil {
		return db.GetUserAccountSummaryRow{}, err
	}
	s.removeStoredAccountAvatar(current.PhotoUrl)
	return s.GetAccount(ctx, userID)
}

func (s *Auth) AccountAvatarPath(ctx context.Context, userID pgtype.UUID, filename string) (string, bool, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return "", false, nil
	}
	current, err := s.GetAccount(ctx, userID)
	if err != nil {
		return "", false, err
	}
	if current.PhotoUrl != accountAvatarURL(filename) {
		return "", false, nil
	}
	return filepath.Join(s.accountAvatarDir(), filename), true, nil
}

func (s *Auth) RevokeSession(ctx context.Context, userID, sessionID pgtype.UUID) error {
	affected, err := s.q.RevokeOwnedAuthSession(ctx, db.RevokeOwnedAuthSessionParams{
		UserID: userID,
		ID:     sessionID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Auth) UpdateSessionLabel(ctx context.Context, userID, sessionID pgtype.UUID, deviceLabel string) error {
	deviceLabel = strings.TrimSpace(deviceLabel)
	if deviceLabel == "" {
		return domain.ErrBadRequest
	}

	affected, err := s.q.UpdateOwnedAuthSessionLabel(ctx, db.UpdateOwnedAuthSessionLabelParams{
		UserID:      userID,
		ID:          sessionID,
		DeviceLabel: deviceLabel,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Auth) GetSidebarPreferences(ctx context.Context, userID pgtype.UUID) (SidebarPreferences, error) {
	row, err := s.q.GetUserUIPreferences(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SidebarPreferences{}, nil
		}
		return SidebarPreferences{}, err
	}
	return decodeSidebarPreferences(row.SidebarPinned, row.SidebarRecent)
}

func (s *Auth) UpdateSidebarPreferences(ctx context.Context, userID pgtype.UUID, prefs SidebarPreferences) (SidebarPreferences, error) {
	normalized, err := normalizeSidebarPreferences(prefs)
	if err != nil {
		return SidebarPreferences{}, err
	}
	pinnedJSON, err := json.Marshal(normalized.PinnedItems)
	if err != nil {
		return SidebarPreferences{}, err
	}
	recentJSON, err := json.Marshal(normalized.RecentItems)
	if err != nil {
		return SidebarPreferences{}, err
	}
	row, err := s.q.UpsertUserUIPreferences(ctx, db.UpsertUserUIPreferencesParams{
		UserID:        userID,
		SidebarPinned: pinnedJSON,
		SidebarRecent: recentJSON,
	})
	if err != nil {
		return SidebarPreferences{}, err
	}
	return decodeSidebarPreferences(row.SidebarPinned, row.SidebarRecent)
}

func (s *Auth) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	row, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.ErrUnauthorized
	}
	user := authUserFromUsernameRow(row)

	if !user.IsActive {
		return domain.ErrSuspended
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrUnauthorized
	}

	if err := validatePassword(username, newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if _, err := s.q.ChangeUserPasswordAndInvalidate(ctx, db.ChangeUserPasswordAndInvalidateParams{ID: user.ID, PasswordHash: string(hash)}); err != nil {
		return err
	}
	return nil
}

func (s *Auth) issueTokenPair(ctx context.Context, user authUserRecord, meta SessionMeta) (domain.TokenPair, error) {
	now := time.Now()
	sessionID, err := randomUUID()
	if err != nil {
		return domain.TokenPair{}, err
	}

	roleStrs, err := s.q.GetUserRoleCodesFromRbac(ctx, user.ID)
	if err != nil {
		return domain.TokenPair{}, err
	}
	roleStrs = normalizeStringSet(roleStrs)
	if len(roleStrs) == 0 {
		if len(user.Roles) > 0 {
			_ = json.Unmarshal(user.Roles, &roleStrs)
			roleStrs = normalizeStringSet(roleStrs)
		}
		if len(roleStrs) == 0 {
			roles, err := s.q.GetUserRoles(ctx, user.ID)
			if err != nil {
				return domain.TokenPair{}, err
			}
			for _, r := range roles {
				roleStrs = append(roleStrs, string(r))
			}
			roleStrs = normalizeStringSet(roleStrs)
		}
	}

	primaryRole := ""
	if len(roleStrs) > 0 {
		primaryRole = roleStrs[0]
	}
	permissions, err := s.q.GetUserPermissionCodes(ctx, user.ID)
	if err != nil {
		return domain.TokenPair{}, err
	}
	permissions = normalizeStringSet(permissions)

	userID := pgUUIDString(user.ID)
	claims := jwt.MapClaims{
		"sub":                  userID,
		"uid":                  userID,
		"usr":                  user.Username,
		"ssid":                 pgUUIDString(sessionID),
		"role":                 primaryRole, // backward compatibility
		"roles":                roleStrs,
		"permissions":          permissions,
		"must_change_password": user.MustChangePassword,
		"type":                 "access",
		"ver":                  int64(user.AuthVersion),
		"iat":                  now.Unix(),
		"exp":                  now.Add(accessTokenTTL).Unix(),
	}

	if user.EmployeeID.Valid {
		claims["eid"] = pgUUIDString(user.EmployeeID)
	}
	if user.StudentID.Valid {
		claims["sid"] = pgUUIDString(user.StudentID)
	}
	if user.ParentID.Valid {
		claims["pid"] = pgUUIDString(user.ParentID)
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessSigned, err := access.SignedString(s.jwtSecret)
	if err != nil {
		return domain.TokenPair{}, err
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"uid":  userID,
		"usr":  user.Username,
		"ssid": pgUUIDString(sessionID),
		"type": "refresh",
		"ver":  int64(user.AuthVersion),
		"iat":  now.Unix(),
		"exp":  now.Add(refreshTokenTTL).Unix(),
	})
	refreshSigned, err := refresh.SignedString(s.jwtSecret)
	if err != nil {
		return domain.TokenPair{}, err
	}

	expiresAt := pgtype.Timestamptz{}
	if err := expiresAt.Scan(now.Add(refreshTokenTTL)); err != nil {
		return domain.TokenPair{}, err
	}
	if _, err := s.q.CreateAuthSession(ctx, db.CreateAuthSessionParams{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: hashToken(refreshSigned),
		ExpiresAt:        expiresAt,
		IpAddress:        meta.IPAddress,
		UserAgent:        meta.UserAgent,
		DeviceLabel:      meta.DeviceLabel,
	}); err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{AccessToken: accessSigned, RefreshToken: refreshSigned, MustChangePassword: user.MustChangePassword}, nil
}

func normalizeStringSet(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalizeAccountChangeHistoryLimit(limit int32) int32 {
	if limit < 1 {
		return 30
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func accountChangeHistoryItemFromRow(row db.ListOwnAccountChangeHistoryRow) (AccountChangeHistoryItem, bool) {
	action := normalizeAccountChangeHistoryAction(row.Action)
	if action == "" {
		return AccountChangeHistoryItem{}, false
	}
	return AccountChangeHistoryItem{
		Action:              action,
		FieldKey:            normalizeAccountChangeHistoryField(row.FieldKey),
		Status:              normalizeAccountChangeHistoryStatus(row.Status),
		CreatedAt:           row.CreatedAt,
		ReviewerUsername:    strings.TrimSpace(row.ReviewerUsername),
		ReviewerDisplayName: strings.TrimSpace(row.ReviewerDisplayName),
		ReviewNote:          strings.TrimSpace(row.ReviewNote),
	}, true
}

func normalizeAccountChangeHistoryAction(action string) string {
	switch strings.TrimSpace(action) {
	case "AUTH_ACCOUNT_CONTACT_UPDATE":
		return "contact_update"
	case "AUTH_ACCOUNT_AVATAR_UPDATE":
		return "avatar_update"
	case "AUTH_ACCOUNT_AVATAR_DELETE":
		return "avatar_delete"
	case "ACCOUNT_CHANGE_REQUEST_CREATED":
		return "change_request_created"
	case "ACCOUNT_CHANGE_REQUEST_CANCELLED":
		return "change_request_cancelled"
	case "ACCOUNT_CHANGE_REQUEST_APPROVED":
		return "change_request_approved"
	case "ACCOUNT_CHANGE_REQUEST_REJECTED":
		return "change_request_rejected"
	default:
		return ""
	}
}

func normalizeAccountChangeHistoryField(fieldKey string) string {
	switch strings.TrimSpace(strings.ToLower(fieldKey)) {
	case "contact", "avatar", "nama", "tanggal_lahir", "parent_name":
		return strings.TrimSpace(strings.ToLower(fieldKey))
	default:
		return ""
	}
}

func normalizeAccountChangeHistoryStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "pending", "approved", "rejected", "cancelled", "completed":
		return strings.TrimSpace(strings.ToLower(status))
	default:
		return "completed"
	}
}

func decodeSidebarPreferences(pinnedJSON, recentJSON []byte) (SidebarPreferences, error) {
	var prefs SidebarPreferences
	if len(pinnedJSON) > 0 {
		if err := json.Unmarshal(pinnedJSON, &prefs.PinnedItems); err != nil {
			return SidebarPreferences{}, err
		}
	}
	if len(recentJSON) > 0 {
		if err := json.Unmarshal(recentJSON, &prefs.RecentItems); err != nil {
			return SidebarPreferences{}, err
		}
	}
	return normalizeSidebarPreferences(prefs)
}

func normalizeSidebarPreferences(prefs SidebarPreferences) (SidebarPreferences, error) {
	const maxPinned = 8
	const maxRecent = 12

	normalize := func(items []string, maxCount int, allowRoot bool) ([]string, error) {
		seen := make(map[string]struct{}, len(items))
		normalized := make([]string, 0, len(items))
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			if !strings.HasPrefix(item, "/") {
				return nil, domain.ErrBadRequest
			}
			if !allowRoot && item == "/" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			normalized = append(normalized, item)
			if len(normalized) == maxCount {
				break
			}
		}
		return normalized, nil
	}

	pinned, err := normalize(prefs.PinnedItems, maxPinned, false)
	if err != nil {
		return SidebarPreferences{}, err
	}
	recent, err := normalize(prefs.RecentItems, maxRecent, true)
	if err != nil {
		return SidebarPreferences{}, err
	}
	return SidebarPreferences{PinnedItems: pinned, RecentItems: recent}, nil
}

func (s *Auth) SeedAdmin(ctx context.Context) error {
	if strings.TrimSpace(s.adminPassword) == "" {
		return fmt.Errorf("ADMIN_PASSWORD is required for initial admin bootstrap")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(s.adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	existing, err := s.q.GetUserByUsername(ctx, "admin")
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		// Admin doesn't exist yet — create
		user, err := s.q.CreateUser(ctx, db.CreateUserParams{
			Username:     "admin",
			PasswordHash: string(hash),
			IsActive:     true,
		})
		if err != nil {
			return err
		}
		if err := s.q.AddUserRole(ctx, db.AddUserRoleParams{
			UserID: user.ID,
			Role:   db.UserRoleAdmin,
		}); err != nil {
			return err
		}
		return s.q.AddUserRbacRoleByCode(ctx, db.AddUserRbacRoleByCodeParams{
			UserID: user.ID,
			Code:   string(db.UserRoleAdmin),
		})
	}

	// Update password to match current ADMIN_PASSWORD env
	if err := s.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           existing.ID,
		PasswordHash: string(hash),
	}); err != nil {
		return err
	}

	// Ensure admin has admin role
	if err := s.q.AddUserRole(ctx, db.AddUserRoleParams{
		UserID: existing.ID,
		Role:   db.UserRoleAdmin,
	}); err != nil {
		return err
	}
	return s.q.AddUserRbacRoleByCode(ctx, db.AddUserRbacRoleByCodeParams{
		UserID: existing.ID,
		Code:   string(db.UserRoleAdmin),
	})
}

func (s *Auth) CurrentAuthVersion(ctx context.Context, subject string) (int64, error) {
	var userID pgtype.UUID
	if err := userID.Scan(subject); err != nil {
		return 0, domain.ErrUnauthorized
	}
	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrUnauthorized
		}
		return 0, err
	}
	if !user.IsActive {
		return 0, domain.ErrUnauthorized
	}
	return int64(user.AuthVersion), nil
}

func (s *Auth) ValidateAccessSession(ctx context.Context, subject, sessionRef string) (bool, error) {
	var userID pgtype.UUID
	if err := userID.Scan(subject); err != nil {
		return false, domain.ErrUnauthorized
	}

	var sessionID pgtype.UUID
	if err := sessionID.Scan(sessionRef); err != nil {
		return false, domain.ErrUnauthorized
	}

	session, err := s.q.GetAuthSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, domain.ErrUnauthorized
		}
		return false, err
	}

	if !session.ExpiresAt.Valid || session.ExpiresAt.Time.Before(time.Now()) {
		return false, nil
	}
	if session.RevokedAt.Valid {
		return false, nil
	}
	if pgUUIDString(session.UserID) != pgUUIDString(userID) {
		return false, nil
	}
	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, domain.ErrUnauthorized
		}
		return false, err
	}
	if !user.IsActive {
		return false, nil
	}
	if _, err := s.q.TouchAuthSessionLastUsed(ctx, sessionID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Auth) validAuthVersion(ctx context.Context, claims jwt.MapClaims) bool {
	subject, _ := claims["sub"].(string)
	if strings.TrimSpace(subject) == "" {
		return false
	}
	current, err := s.CurrentAuthVersion(ctx, subject)
	if err != nil {
		return false
	}
	raw, ok := claims["ver"]
	if !ok {
		return false
	}
	claimed, ok := asInt64(raw)
	if !ok {
		return false
	}
	return claimed == current
}

func asInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case float64:
		return int64(x), true
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func normalizeContactPatchValue(value *string, current string, maxLen int) (string, error) {
	if value == nil {
		return strings.TrimSpace(current), nil
	}
	trimmed := strings.TrimSpace(*value)
	if len(trimmed) > maxLen {
		return "", domain.ErrBadRequest
	}
	return trimmed, nil
}

func looksLikeEmail(email string) bool {
	if strings.ContainsAny(email, " \t\r\n") {
		return false
	}
	at := strings.Index(email, "@")
	return at > 0 && at < len(email)-1 && strings.Contains(email[at+1:], ".")
}

func (s *Auth) updateOwnedAccountAvatarURL(ctx context.Context, userID pgtype.UUID, profileType, photoURL string) error {
	var err error
	switch profileType {
	case "employee":
		_, err = s.q.UpdateOwnedEmployeeAvatar(ctx, db.UpdateOwnedEmployeeAvatarParams{
			UserID:   userID,
			PhotoUrl: photoURL,
		})
	case "student":
		_, err = s.q.UpdateOwnedStudentAvatar(ctx, db.UpdateOwnedStudentAvatarParams{
			UserID:   userID,
			PhotoUrl: photoURL,
		})
	case "parent":
		_, err = s.q.UpdateOwnedParentAvatar(ctx, db.UpdateOwnedParentAvatarParams{
			UserID:   userID,
			PhotoUrl: photoURL,
		})
	default:
		return domain.ErrForbidden
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrForbidden
	}
	return err
}

func validateAccountAvatar(input UploadAccountAvatarInput) error {
	if !input.UserID.Valid {
		return domain.ErrUnauthorized
	}
	if input.File == nil {
		return fmt.Errorf("file wajib diisi")
	}
	if input.FileSize <= 0 {
		return fmt.Errorf("ukuran file tidak valid")
	}
	if input.FileSize > maxAccountAvatarBytes {
		return fmt.Errorf("ukuran foto maksimal 2MB")
	}
	ext := strings.ToLower(strings.TrimSpace(input.Ext))
	expectedMime, ok := allowedAccountAvatarTypes[ext]
	if !ok || expectedMime != input.MimeType {
		return fmt.Errorf("hanya file JPG, PNG, atau WebP yang diperbolehkan")
	}
	return nil
}

func normalizeAvatarDir(values ...string) string {
	if len(values) > 0 && strings.TrimSpace(values[0]) != "" {
		return strings.TrimSpace(values[0])
	}
	return "uploads/avatars"
}

func (s *Auth) accountAvatarDir() string {
	return normalizeAvatarDir(s.avatarDir)
}

func accountAvatarURL(filename string) string {
	return "/api/auth/account/avatar/" + filename
}

func accountAvatarFilename(photoURL string) (string, bool) {
	const prefix = "/api/auth/account/avatar/"
	photoURL = strings.TrimSpace(photoURL)
	if !strings.HasPrefix(photoURL, prefix) {
		return "", false
	}
	filename := strings.TrimPrefix(photoURL, prefix)
	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return "", false
	}
	return filename, true
}

func (s *Auth) removeStoredAccountAvatar(photoURL string) {
	filename, ok := accountAvatarFilename(photoURL)
	if !ok {
		return
	}
	_ = os.Remove(filepath.Join(s.accountAvatarDir(), filename))
}

func validatePassword(username, password string) error {
	trimmed := strings.TrimSpace(password)
	if len(trimmed) < 8 {
		return domain.ErrWeakPassword
	}
	// bcrypt silently truncates inputs longer than 72 bytes, which would let two
	// different passwords share the same hash if they share a 72-byte prefix.
	// Reject up front so the user gets a clear error instead of a confusing
	// "password works on first try, fails on retype" symptom.
	if len(trimmed) > 72 {
		return domain.ErrWeakPassword
	}

	lowerPassword := strings.ToLower(trimmed)
	lowerUsername := strings.ToLower(strings.TrimSpace(username))
	if lowerUsername != "" && lowerPassword == lowerUsername {
		return domain.ErrWeakPassword
	}

	allDigits := true
	for _, r := range trimmed {
		if r < '0' || r > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		return domain.ErrWeakPassword
	}

	return nil
}

func ValidatePassword(username, password string) error {
	return validatePassword(username, password)
}

func (s *Auth) parseTokenClaims(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (s *Auth) validateRefreshSession(ctx context.Context, refreshToken string, claims jwt.MapClaims) (db.AuthSession, error) {
	rawSID, _ := claims["ssid"].(string)
	if strings.TrimSpace(rawSID) == "" {
		return db.AuthSession{}, domain.ErrUnauthorized
	}
	var sessionID pgtype.UUID
	if err := sessionID.Scan(rawSID); err != nil {
		return db.AuthSession{}, domain.ErrUnauthorized
	}

	session, err := s.q.GetAuthSession(ctx, sessionID)
	if err != nil {
		return db.AuthSession{}, domain.ErrUnauthorized
	}
	if session.RevokedAt.Valid {
		return db.AuthSession{}, domain.ErrUnauthorized
	}
	if !session.ExpiresAt.Valid || session.ExpiresAt.Time.Before(time.Now()) {
		return db.AuthSession{}, domain.ErrUnauthorized
	}
	if subtle.ConstantTimeCompare([]byte(session.RefreshTokenHash), []byte(hashToken(refreshToken))) != 1 {
		return db.AuthSession{}, domain.ErrUnauthorized
	}
	return session, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomUUID() (pgtype.UUID, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return pgtype.UUID{}, err
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80

	hexValue := hex.EncodeToString(buf)
	uuidValue := fmt.Sprintf("%s-%s-%s-%s-%s",
		hexValue[0:8],
		hexValue[8:12],
		hexValue[12:16],
		hexValue[16:20],
		hexValue[20:32],
	)
	var id pgtype.UUID
	if err := id.Scan(uuidValue); err != nil {
		return pgtype.UUID{}, err
	}
	return id, nil
}

func userIDFromClaims(claims jwt.MapClaims) (pgtype.UUID, error) {
	subject, _ := claims["sub"].(string)
	if strings.TrimSpace(subject) == "" {
		return pgtype.UUID{}, domain.ErrUnauthorized
	}
	var userID pgtype.UUID
	if err := userID.Scan(subject); err != nil {
		return pgtype.UUID{}, domain.ErrUnauthorized
	}
	return userID, nil
}

func authUserFromUsernameRow(row db.GetUserByUsernameRow) authUserRecord {
	return authUserRecord{
		ID:                 row.ID,
		Username:           row.Username,
		PasswordHash:       row.PasswordHash,
		EmployeeID:         row.EmployeeID,
		StudentID:          row.StudentID,
		ParentID:           row.ParentID,
		IsActive:           row.IsActive,
		AuthVersion:        row.AuthVersion,
		MustChangePassword: row.MustChangePassword,
		Roles:              authRolesBytes(row.Roles),
	}
}

func authUserFromIDRow(row db.GetUserByIDRow) authUserRecord {
	return authUserRecord{
		ID:                 row.ID,
		Username:           row.Username,
		PasswordHash:       row.PasswordHash,
		EmployeeID:         row.EmployeeID,
		StudentID:          row.StudentID,
		ParentID:           row.ParentID,
		IsActive:           row.IsActive,
		AuthVersion:        row.AuthVersion,
		MustChangePassword: row.MustChangePassword,
		Roles:              authRolesBytes(row.Roles),
	}
}

func authUserFromStudentIDRow(row db.GetUserByStudentIDRow) authUserRecord {
	return authUserRecord{
		ID:                 row.ID,
		Username:           row.Username,
		PasswordHash:       row.PasswordHash,
		EmployeeID:         row.EmployeeID,
		StudentID:          row.StudentID,
		ParentID:           row.ParentID,
		IsActive:           row.IsActive,
		AuthVersion:        row.AuthVersion,
		MustChangePassword: row.MustChangePassword,
		Roles:              authRolesBytes(row.Roles),
	}
}

func authRolesBytes(value any) []byte {
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		payload, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		return payload
	}
}

func normalizeSessionMeta(meta SessionMeta) SessionMeta {
	meta.IPAddress = strings.TrimSpace(meta.IPAddress)
	meta.UserAgent = strings.TrimSpace(meta.UserAgent)
	meta.DeviceLabel = strings.TrimSpace(meta.DeviceLabel)
	if meta.DeviceLabel == "" {
		meta.DeviceLabel = deriveDeviceLabel(meta.UserAgent)
	}
	return meta
}

func deriveDeviceLabel(userAgent string) string {
	ua := strings.ToLower(userAgent)

	osLabel := "Perangkat Tidak Dikenal"
	switch {
	case strings.Contains(ua, "android"):
		osLabel = "Android"
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipad"), strings.Contains(ua, "ios"):
		osLabel = "iPhone/iPad"
	case strings.Contains(ua, "windows"):
		osLabel = "Windows"
	case strings.Contains(ua, "mac os"), strings.Contains(ua, "macintosh"):
		osLabel = "Mac"
	case strings.Contains(ua, "linux"):
		osLabel = "Linux"
	}

	browserLabel := "Browser"
	switch {
	case strings.Contains(ua, "edg/"):
		browserLabel = "Edge"
	case strings.Contains(ua, "firefox/"):
		browserLabel = "Firefox"
	case strings.Contains(ua, "chrome/") && !strings.Contains(ua, "edg/"):
		browserLabel = "Chrome"
	case strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome/"):
		browserLabel = "Safari"
	}

	return browserLabel + " di " + osLabel
}

// randomHex generates a cryptographically secure random hex string.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
