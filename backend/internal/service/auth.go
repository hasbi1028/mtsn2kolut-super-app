package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/pusaka/backend/internal/domain"
	db "github.com/pusaka/backend/internal/repository/postgres"
)

type Auth struct {
	q             *db.Queries
	jwtSecret     []byte
	adminPassword string // plaintext only used for initial seed
}

func NewAuth(q *db.Queries, jwtSecret, adminPassword string) *Auth {
	return &Auth{q: q, jwtSecret: []byte(jwtSecret), adminPassword: adminPassword}
}

func (s *Auth) SeedAdmin(ctx context.Context) error {
	_, err := s.q.GetSetting(ctx, "admin_username")
	if err == nil {
		return nil // already seeded
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: "admin_username", Value: "admin"}); err != nil {
		return err
	}
	return s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: "admin_password", Value: string(hash)})
}

func (s *Auth) Login(ctx context.Context, username, password string) (string, error) {
	uRow, err := s.q.GetSetting(ctx, "admin_username")
	if err != nil || uRow.Value != username {
		return "", domain.ErrUnauthorized
	}
	pRow, err := s.q.GetSetting(ctx, "admin_password")
	if err != nil {
		return "", domain.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(pRow.Value), []byte(password)); err != nil {
		return "", domain.ErrUnauthorized
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(12 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	return token.SignedString(s.jwtSecret)
}

func (s *Auth) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	if err := s.verifyPassword(ctx, username, oldPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: "admin_password", Value: string(hash)})
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
