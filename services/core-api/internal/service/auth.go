package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error
	GetUserRoles(ctx context.Context, userID pgtype.UUID) ([]db.UserRole, error)
	AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error
	IncrementUserAuthVersion(ctx context.Context, id pgtype.UUID) (int32, error)
	CreateAuthSession(ctx context.Context, arg db.CreateAuthSessionParams) (db.AuthSession, error)
	GetAuthSession(ctx context.Context, id pgtype.UUID) (db.AuthSession, error)
	RevokeAuthSession(ctx context.Context, id pgtype.UUID) error
	RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error)
}

type Auth struct {
	q             authStore
	jwtSecret     []byte
	adminPassword string
}

func NewAuth(q *db.Queries, jwtSecret, adminPassword string) *Auth {
	return &Auth{q: q, jwtSecret: []byte(jwtSecret), adminPassword: adminPassword}
}

func (s *Auth) Login(ctx context.Context, username, password string) (domain.TokenPair, error) {
	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TokenPair{}, domain.ErrUnauthorized
		}
		return domain.TokenPair{}, err
	}

	if !user.IsActive {
		return domain.TokenPair{}, domain.ErrSuspended
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	return s.issueTokenPair(ctx, user)
}

func (s *Auth) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
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
	username, ok := claims["sub"].(string)
	if !ok || username == "" {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	if !user.IsActive {
		return domain.TokenPair{}, domain.ErrSuspended
	}

	if !s.validAuthVersion(ctx, claims) {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	if err := s.q.RevokeAuthSession(ctx, session.ID); err != nil {
		return domain.TokenPair{}, err
	}
	return s.issueTokenPair(ctx, user)
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

func (s *Auth) LogoutAll(ctx context.Context, username string) error {
	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.ErrUnauthorized
	}
	if _, err := s.q.RevokeAllAuthSessionsForUser(ctx, user.ID); err != nil {
		return err
	}
	_, err = s.q.IncrementUserAuthVersion(ctx, user.ID)
	return err
}

func (s *Auth) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.ErrUnauthorized
	}

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

	if err := s.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: string(hash),
	}); err != nil {
		return err
	}
	_, err = s.q.IncrementUserAuthVersion(ctx, user.ID)
	return err
}

func (s *Auth) issueTokenPair(ctx context.Context, user db.GetUserByUsernameRow) (domain.TokenPair, error) {
	now := time.Now()
	sessionID, err := randomUUID()
	if err != nil {
		return domain.TokenPair{}, err
	}

	var roleStrs []string
	if len(user.Roles) > 0 {
		_ = json.Unmarshal(user.Roles, &roleStrs)
	}

	if len(roleStrs) == 0 {
		// Fallback to fetch roles if not joined or empty
		roles, err := s.q.GetUserRoles(ctx, user.ID)
		if err == nil {
			for _, r := range roles {
				roleStrs = append(roleStrs, string(r))
			}
		}
	}

	primaryRole := ""
	if len(roleStrs) > 0 {
		primaryRole = roleStrs[0]
	}

	claims := jwt.MapClaims{
		"sub":   user.Username,
		"uid":   pgUUIDString(user.ID),
		"role":  primaryRole, // backward compatibility
		"roles": roleStrs,
		"type":  "access",
		"ver":   int64(user.AuthVersion),
		"iat":   now.Unix(),
		"exp":   now.Add(accessTokenTTL).Unix(),
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
		"sub":  user.Username,
		"sid":  pgUUIDString(sessionID),
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
	}); err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{AccessToken: accessSigned, RefreshToken: refreshSigned}, nil
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
		return s.q.AddUserRole(ctx, db.AddUserRoleParams{
			UserID: user.ID,
			Role:   db.UserRoleAdmin,
		})
	}

	// Ensure admin has admin role
	return s.q.AddUserRole(ctx, db.AddUserRoleParams{
		UserID: existing.ID,
		Role:   db.UserRoleAdmin,
	})
}

func (s *Auth) CurrentAuthVersion(ctx context.Context, username string) (int64, error) {
	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrUnauthorized
		}
		return 0, err
	}
	return int64(user.AuthVersion), nil
}

func (s *Auth) validAuthVersion(ctx context.Context, claims jwt.MapClaims) bool {
	username, _ := claims["sub"].(string)
	if strings.TrimSpace(username) == "" {
		return false
	}
	current, err := s.CurrentAuthVersion(ctx, username)
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

func validatePassword(username, password string) error {
	trimmed := strings.TrimSpace(password)
	if len(trimmed) < 8 {
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
	rawSID, _ := claims["sid"].(string)
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
	if session.RefreshTokenHash != hashToken(refreshToken) {
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
