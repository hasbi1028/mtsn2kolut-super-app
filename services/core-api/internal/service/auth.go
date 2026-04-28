package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const (
	accessTokenTTL  = 1 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
	authVersionKey  = "auth_version"
)

// authStore defines the subset of db.Queries that Auth service needs.
type authStore interface {
	GetUserByUsername(ctx context.Context, username string) (db.User, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error
	GetSetting(ctx context.Context, key string) (db.AppSetting, error)
	UpsertSetting(ctx context.Context, arg db.UpsertSettingParams) error
	ListSettings(ctx context.Context) ([]db.AppSetting, error)
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	return s.issueTokenPair(ctx, user)
}

func (s *Auth) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	if tokenType, _ := claims["type"].(string); tokenType != "refresh" {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	username, ok := claims["sub"].(string)
	if !ok || username == "" {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}

	if !s.validAuthVersion(ctx, claims) {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	return s.issueTokenPair(ctx, user)
}

func (s *Auth) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	user, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrUnauthorized
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
	return s.bumpAuthVersion(ctx)
}

func (s *Auth) issueTokenPair(ctx context.Context, user db.User) (domain.TokenPair, error) {
	now := time.Now()
	version, err := s.CurrentAuthVersion(ctx)
	if err != nil {
		return domain.TokenPair{}, err
	}

	claims := jwt.MapClaims{
		"sub":  user.Username,
		"uid":  pgUUIDString(user.ID),
		"role": string(user.Role),
		"type": "access",
		"ver":  version,
		"iat":  now.Unix(),
		"exp":  now.Add(accessTokenTTL).Unix(),
	}
	
	if user.EmployeeID.Valid {
		claims["eid"] = pgUUIDString(user.EmployeeID)
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessSigned, err := access.SignedString(s.jwtSecret)
	if err != nil {
		return domain.TokenPair{}, err
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.Username,
		"type": "refresh",
		"ver":  version,
		"iat":  now.Unix(),
		"exp":  now.Add(refreshTokenTTL).Unix(),
	})
	refreshSigned, err := refresh.SignedString(s.jwtSecret)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{AccessToken: accessSigned, RefreshToken: refreshSigned}, nil
}

func (s *Auth) SeedAdmin(ctx context.Context) error {
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
		_, err = s.q.CreateUser(ctx, db.CreateUserParams{
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         db.UserRoleAdmin,
		})
		return err
	}

	// Admin exists — update password
	return s.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           existing.ID,
		PasswordHash: string(hash),
	})
}

func (s *Auth) CurrentAuthVersion(ctx context.Context) (int64, error) {
	row, err := s.q.GetSetting(ctx, authVersionKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	v, err := strconv.ParseInt(row.Value, 10, 64)
	if err != nil {
		return 0, nil
	}
	return v, nil
}

func (s *Auth) bumpAuthVersion(ctx context.Context) error {
	v, err := s.CurrentAuthVersion(ctx)
	if err != nil {
		return err
	}
	return s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: authVersionKey, Value: strconv.FormatInt(v+1, 10)})
}

func (s *Auth) validAuthVersion(ctx context.Context, claims jwt.MapClaims) bool {
	current, err := s.CurrentAuthVersion(ctx)
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
