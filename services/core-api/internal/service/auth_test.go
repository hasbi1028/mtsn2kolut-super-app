package service

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeStore struct {
	settings map[string]string
	users    map[string]db.User
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		settings: map[string]string{},
		users:    map[string]db.User{},
	}
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

func (f *fakeStore) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	u, ok := f.users[username]
	if !ok {
		return db.User{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	var id pgtype.UUID
	_ = id.Scan("11111111-1111-1111-1111-111111111111")
	u := db.User{
		ID:           id,
		Username:     arg.Username,
		PasswordHash: arg.PasswordHash,
		Role:         arg.Role,
		EmployeeID:   arg.EmployeeID,
	}
	f.users[arg.Username] = u
	return u, nil
}

func (f *fakeStore) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	for k, u := range f.users {
		if u.ID == arg.ID {
			u.PasswordHash = arg.PasswordHash
			f.users[k] = u
			return nil
		}
	}
	return pgx.ErrNoRows
}

func TestAuthSeedAdminCreatesUser(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	u, ok := store.users["admin"]
	if !ok {
		t.Fatal("admin user not created")
	}
	if u.Role != db.UserRoleAdmin {
		t.Fatalf("admin role = %q, want admin", u.Role)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("admin")); err != nil {
		t.Fatalf("admin password hash mismatch: %v", err)
	}
}

func TestAuthRefreshRejectsOldTokenAfterPasswordChange(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}

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
