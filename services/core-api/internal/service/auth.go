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

type authStore interface {
	GetSetting(ctx context.Context, key string) (db.AppSetting, error)
	UpsertSetting(ctx context.Context, arg db.UpsertSettingParams) error
}

type Auth struct {
	q             authStore
	jwtSecret     []byte
	adminPassword string
}

func NewAuth(q *db.Queries, jwtSecret, adminPassword string) *Auth {
	return &Auth{q: q, jwtSecret: []byte(jwtSecret), adminPassword: adminPassword}
}

func (s *Auth) SeedAdmin(ctx context.Context) error {
	_, err := s.q.GetSetting(ctx, "admin_username")
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if s.adminPassword == "" {
		return errors.New("ADMIN_PASSWORD is required to seed initial admin user")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(s.adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: "admin_username", Value: "admin"}); err != nil {
		return err
	}
	if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: authVersionKey, Value: "0"}); err != nil {
		return err
	}
	return s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: "admin_password", Value: string(hash)})
}

func (s *Auth) Login(ctx context.Context, username, password string) (domain.TokenPair, error) {
	uRow, err := s.q.GetSetting(ctx, "admin_username")
	if err != nil || uRow.Value != username {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	pRow, err := s.q.GetSetting(ctx, "admin_password")
	if err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(pRow.Value), []byte(password)); err != nil {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	return s.issueTokenPair(ctx, username)
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
	if !s.validAuthVersion(ctx, claims) {
		return domain.TokenPair{}, domain.ErrUnauthorized
	}
	return s.issueTokenPair(ctx, username)
}

func (s *Auth) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	if err := s.verifyPassword(ctx, username, oldPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: "admin_password", Value: string(hash)}); err != nil {
		return err
	}
	return s.bumpAuthVersion(ctx)
}

func (s *Auth) CurrentAuthVersion(ctx context.Context) (int64, error) {
	return s.currentAuthVersion(ctx)
}

func (s *Auth) issueTokenPair(ctx context.Context, username string) (domain.TokenPair, error) {
	now := time.Now()
	version, err := s.currentAuthVersion(ctx)
	if err != nil {
		return domain.TokenPair{}, err
	}

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  username,
		"type": "access",
		"ver":  version,
		"iat":  now.Unix(),
		"exp":  now.Add(accessTokenTTL).Unix(),
	})
	accessSigned, err := access.SignedString(s.jwtSecret)
	if err != nil {
		return domain.TokenPair{}, err
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  username,
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

func (s *Auth) currentAuthVersion(ctx context.Context) (int64, error) {
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
	v, err := s.currentAuthVersion(ctx)
	if err != nil {
		return err
	}
	return s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: authVersionKey, Value: strconv.FormatInt(v+1, 10)})
}

func (s *Auth) validAuthVersion(ctx context.Context, claims jwt.MapClaims) bool {
	current, err := s.currentAuthVersion(ctx)
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

func (s *Auth) verifyPassword(ctx context.Context, username, password string) error {
	uRow, err := s.q.GetSetting(ctx, "admin_username")
	if err != nil || uRow.Value != username {
		return domain.ErrUnauthorized
	}
	pRow, err := s.q.GetSetting(ctx, "admin_password")
	if err != nil {
		return domain.ErrUnauthorized
	}
	return bcrypt.CompareHashAndPassword([]byte(pRow.Value), []byte(password))
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
