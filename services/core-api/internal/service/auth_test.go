package service

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeStore struct {
	settings map[string]string
}

func newFakeStore() *fakeStore {
	return &fakeStore{settings: map[string]string{}}
}

func (f *fakeStore) GetSetting(ctx context.Context, key string) (db.AppSetting, error) {
	v, ok := f.settings[key]
	if !ok {
		return db.AppSetting{}, pgx.ErrNoRows
	}
	return db.AppSetting{Key: key, Value: v}, nil
}

func (f *fakeStore) ListSettings(ctx context.Context) ([]db.AppSetting, error) {
	items := make([]db.AppSetting, 0, len(f.settings))
	for key, value := range f.settings {
		items = append(items, db.AppSetting{Key: key, Value: value})
	}
	return items, nil
}

func (f *fakeStore) UpsertSetting(ctx context.Context, arg db.UpsertSettingParams) error {
	f.settings[arg.Key] = arg.Value
	return nil
}

func TestAuthSeedAdminSetsAuthVersion(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	if got := store.settings["admin_username"]; got != "admin" {
		t.Fatalf("admin_username = %q", got)
	}
	if got := store.settings["auth_version"]; got != "0" {
		t.Fatalf("auth_version = %q, want 0", got)
	}
	if _, err := bcrypt.Cost([]byte(store.settings["admin_password"])); err != nil {
		t.Fatalf("admin_password not bcrypt hash: %v", err)
	}
}

func TestAuthRefreshRejectsOldTokenAfterPasswordChange(t *testing.T) {
	store := newFakeStore()
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	store.settings["admin_username"] = "admin"
	store.settings["admin_password"] = string(hash)
	store.settings["auth_version"] = "0"

	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	pair, err := svc.Login(context.Background(), "admin", "admin")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(pair.RefreshToken, claims); err != nil {
		t.Fatalf("ParseUnverified(refresh) error = %v", err)
	}
	if got := claims["ver"]; got != float64(0) {
		t.Fatalf("refresh ver = %v, want 0", got)
	}

	if err := svc.ChangePassword(context.Background(), "admin", "admin", "newpass"); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if got := store.settings["auth_version"]; got != "1" {
		t.Fatalf("auth_version after change = %q, want 1", got)
	}
	if _, err := svc.Refresh(context.Background(), pair.RefreshToken); err == nil {
		t.Fatalf("Refresh() error = nil, want unauthorized")
	}
}
