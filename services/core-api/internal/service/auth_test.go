package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeStore struct {
	settings     map[string]string
	users        map[string]db.User
	userRoles    map[pgtype.UUID][]db.UserRole
	authSessions map[pgtype.UUID]db.AuthSession
	uiPrefs      map[pgtype.UUID]db.UserUiPreference
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		settings:     map[string]string{},
		users:        map[string]db.User{},
		userRoles:    map[pgtype.UUID][]db.UserRole{},
		authSessions: map[pgtype.UUID]db.AuthSession{},
		uiPrefs:      map[pgtype.UUID]db.UserUiPreference{},
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

func (f *fakeStore) GetUserByUsername(ctx context.Context, username string) (db.GetUserByUsernameRow, error) {
	u, ok := f.users[username]
	if !ok {
		return db.GetUserByUsernameRow{}, pgx.ErrNoRows
	}
	roles := f.userRoles[u.ID]
	rolesJSON, _ := json.Marshal(roles)
	return db.GetUserByUsernameRow{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		EmployeeID:   u.EmployeeID,
		StudentID:    u.StudentID,
		ParentID:     u.ParentID,
		IsActive:     u.IsActive,
		AuthVersion:  u.AuthVersion,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Roles:        rolesJSON,
	}, nil
}

func (f *fakeStore) GetUserByID(ctx context.Context, id pgtype.UUID) (db.GetUserByIDRow, error) {
	for _, u := range f.users {
		if u.ID == id {
			roles := f.userRoles[u.ID]
			rolesJSON, _ := json.Marshal(roles)
			return db.GetUserByIDRow{
				ID:           u.ID,
				Username:     u.Username,
				PasswordHash: u.PasswordHash,
				EmployeeID:   u.EmployeeID,
				StudentID:    u.StudentID,
				ParentID:     u.ParentID,
				IsActive:     u.IsActive,
				AuthVersion:  u.AuthVersion,
				CreatedAt:    u.CreatedAt,
				UpdatedAt:    u.UpdatedAt,
				Roles:        rolesJSON,
			}, nil
		}
	}
	return db.GetUserByIDRow{}, pgx.ErrNoRows
}

func (f *fakeStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	var id pgtype.UUID
	_ = id.Scan("11111111-1111-1111-1111-111111111111")
	u := db.User{
		ID:           id,
		Username:     arg.Username,
		PasswordHash: arg.PasswordHash,
		EmployeeID:   arg.EmployeeID,
		StudentID:    arg.StudentID,
		ParentID:     arg.ParentID,
		IsActive:     arg.IsActive,
		AuthVersion:  0,
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

func (f *fakeStore) GetUserRoles(ctx context.Context, userID pgtype.UUID) ([]db.UserRole, error) {
	return f.userRoles[userID], nil
}

func (f *fakeStore) AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error {
	f.userRoles[arg.UserID] = append(f.userRoles[arg.UserID], arg.Role)
	return nil
}

func (f *fakeStore) IncrementUserAuthVersion(ctx context.Context, id pgtype.UUID) (int32, error) {
	for key, user := range f.users {
		if user.ID == id {
			user.AuthVersion++
			f.users[key] = user
			return user.AuthVersion, nil
		}
	}
	return 0, pgx.ErrNoRows
}

func (f *fakeStore) CreateAuthSession(ctx context.Context, arg db.CreateAuthSessionParams) (db.AuthSession, error) {
	now := pgtype.Timestamptz{}
	_ = now.Scan(time.Now())
	session := db.AuthSession{
		ID:               arg.ID,
		UserID:           arg.UserID,
		RefreshTokenHash: arg.RefreshTokenHash,
		ExpiresAt:        arg.ExpiresAt,
		LastUsedAt:       now,
	}
	f.authSessions[arg.ID] = session
	return session, nil
}

func (f *fakeStore) GetAuthSession(ctx context.Context, id pgtype.UUID) (db.AuthSession, error) {
	session, ok := f.authSessions[id]
	if !ok {
		return db.AuthSession{}, pgx.ErrNoRows
	}
	return session, nil
}

func (f *fakeStore) RevokeAuthSession(ctx context.Context, id pgtype.UUID) error {
	session, ok := f.authSessions[id]
	if !ok {
		return pgx.ErrNoRows
	}
	now := pgtype.Timestamptz{}
	_ = now.Scan(time.Now())
	session.RevokedAt = now
	f.authSessions[id] = session
	return nil
}

func (f *fakeStore) RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error) {
	var count int64
	for key, session := range f.authSessions {
		if session.UserID == userID && !session.RevokedAt.Valid {
			now := pgtype.Timestamptz{}
			_ = now.Scan(time.Now())
			session.RevokedAt = now
			f.authSessions[key] = session
			count++
		}
	}
	return count, nil
}

func (f *fakeStore) ListActiveAuthSessionsByUser(ctx context.Context, userID pgtype.UUID) ([]db.AuthSession, error) {
	items := make([]db.AuthSession, 0)
	for _, session := range f.authSessions {
		if session.UserID == userID && !session.RevokedAt.Valid {
			items = append(items, session)
		}
	}
	return items, nil
}

func (f *fakeStore) RevokeOwnedAuthSession(ctx context.Context, arg db.RevokeOwnedAuthSessionParams) (int64, error) {
	session, ok := f.authSessions[arg.ID]
	if !ok || session.UserID != arg.UserID || session.RevokedAt.Valid {
		return 0, nil
	}
	now := pgtype.Timestamptz{}
	_ = now.Scan(time.Now())
	session.RevokedAt = now
	f.authSessions[arg.ID] = session
	return 1, nil
}

func (f *fakeStore) UpdateOwnedAuthSessionLabel(ctx context.Context, arg db.UpdateOwnedAuthSessionLabelParams) (int64, error) {
	session, ok := f.authSessions[arg.ID]
	if !ok || session.UserID != arg.UserID || session.RevokedAt.Valid {
		return 0, nil
	}
	session.DeviceLabel = arg.DeviceLabel
	f.authSessions[arg.ID] = session
	return 1, nil
}

func (f *fakeStore) GetUserUIPreferences(ctx context.Context, userID pgtype.UUID) (db.UserUiPreference, error) {
	prefs, ok := f.uiPrefs[userID]
	if !ok {
		return db.UserUiPreference{}, pgx.ErrNoRows
	}
	return prefs, nil
}

func (f *fakeStore) UpsertUserUIPreferences(ctx context.Context, arg db.UpsertUserUIPreferencesParams) (db.UserUiPreference, error) {
	now := pgtype.Timestamptz{}
	_ = now.Scan(time.Now())
	prefs, ok := f.uiPrefs[arg.UserID]
	if !ok {
		prefs = db.UserUiPreference{
			UserID:    arg.UserID,
			CreatedAt: now,
		}
	}
	prefs.SidebarPinned = arg.SidebarPinned
	prefs.SidebarRecent = arg.SidebarRecent
	prefs.UpdatedAt = now
	f.uiPrefs[arg.UserID] = prefs
	return prefs, nil
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
	roles := store.userRoles[u.ID]
	hasAdmin := false
	for _, r := range roles {
		if r == db.UserRoleAdmin {
			hasAdmin = true
			break
		}
	}
	if !hasAdmin {
		t.Fatal("admin role not assigned")
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

	pair, err := svc.Login(context.Background(), "admin", "admin", SessionMeta{})
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

	if err := svc.ChangePassword(context.Background(), "admin", "admin", "newpass123"); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if got := store.users["admin"].AuthVersion; got != 1 {
		t.Fatalf("auth_version after change = %d, want 1", got)
	}
	if _, err := svc.Refresh(context.Background(), pair.RefreshToken, SessionMeta{}); err == nil {
		t.Fatalf("Refresh() error = nil, want unauthorized")
	}
}

func TestAuthSidebarPreferencesRoundTrip(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	user := store.users["admin"]

	got, err := svc.UpdateSidebarPreferences(context.Background(), user.ID, SidebarPreferences{
		PinnedItems: []string{"/grades", "/grades", "/", "/settings"},
		RecentItems: []string{"/", "/inventory", "/inventory", "/grades"},
	})
	if err != nil {
		t.Fatalf("UpdateSidebarPreferences() error = %v", err)
	}
	if len(got.PinnedItems) != 2 || got.PinnedItems[0] != "/grades" || got.PinnedItems[1] != "/settings" {
		t.Fatalf("unexpected pinned items: %#v", got.PinnedItems)
	}
	if len(got.RecentItems) != 3 || got.RecentItems[0] != "/" || got.RecentItems[1] != "/inventory" || got.RecentItems[2] != "/grades" {
		t.Fatalf("unexpected recent items: %#v", got.RecentItems)
	}

	loaded, err := svc.GetSidebarPreferences(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetSidebarPreferences() error = %v", err)
	}
	if len(loaded.PinnedItems) != 2 || loaded.PinnedItems[0] != "/grades" || loaded.PinnedItems[1] != "/settings" {
		t.Fatalf("unexpected loaded pinned items: %#v", loaded.PinnedItems)
	}
}

func TestAuthSidebarPreferencesRejectInvalidPath(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	user := store.users["admin"]

	_, err := svc.UpdateSidebarPreferences(context.Background(), user.ID, SidebarPreferences{
		PinnedItems: []string{"grades"},
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestAuthLoginRejectsSuspendedAccount(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	var id pgtype.UUID
	_ = id.Scan("22222222-2222-2222-2222-222222222222")
	store.users["guru"] = db.User{
		ID:           id,
		Username:     "guru",
		PasswordHash: string(hash),
		IsActive:     false,
	}

	_, err = svc.Login(context.Background(), "guru", "password123", SessionMeta{})
	if !errors.Is(err, domain.ErrSuspended) {
		t.Fatalf("Login() error = %v, want ErrSuspended", err)
	}
}

func TestAuthChangePasswordRejectsWeakPassword(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	var id pgtype.UUID
	_ = id.Scan("33333333-3333-3333-3333-333333333333")
	store.users["guru"] = db.User{
		ID:           id,
		Username:     "guru",
		PasswordHash: string(hash),
		IsActive:     true,
	}

	err = svc.ChangePassword(context.Background(), "guru", "password123", "12345678")
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("ChangePassword() error = %v, want ErrWeakPassword", err)
	}
}

func TestAuthSeedAdminDoesNotOverwriteExistingPassword(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "new-admin-password"}

	oldHash, err := bcrypt.GenerateFromPassword([]byte("old-admin-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	var id pgtype.UUID
	_ = id.Scan("44444444-4444-4444-4444-444444444444")
	store.users["admin"] = db.User{
		ID:           id,
		Username:     "admin",
		PasswordHash: string(oldHash),
		IsActive:     true,
	}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}

	admin := store.users["admin"]
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte("old-admin-password")); err != nil {
		t.Fatalf("admin password was unexpectedly changed: %v", err)
	}
}

func TestAuthLogoutRevokesRefreshSession(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}

	pair, err := svc.Login(context.Background(), "admin", "admin", SessionMeta{})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(pair.RefreshToken, claims); err != nil {
		t.Fatalf("ParseUnverified(refresh) error = %v", err)
	}
	rawSID, _ := claims["ssid"].(string)
	var sid pgtype.UUID
	if err := sid.Scan(rawSID); err != nil {
		t.Fatalf("sid Scan() error = %v", err)
	}

	if err := svc.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	session, err := store.GetAuthSession(context.Background(), sid)
	if err != nil {
		t.Fatalf("GetAuthSession() error = %v", err)
	}
	if !session.RevokedAt.Valid {
		t.Fatal("session was not revoked")
	}
	if _, err := svc.Refresh(context.Background(), pair.RefreshToken, SessionMeta{}); err == nil {
		t.Fatal("Refresh() succeeded after logout revoke")
	}
}

func TestAuthLogoutAllRevokesUserSessionsAndBumpsVersion(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}

	first, err := svc.Login(context.Background(), "admin", "admin", SessionMeta{})
	if err != nil {
		t.Fatalf("first Login() error = %v", err)
	}
	second, err := svc.Login(context.Background(), "admin", "admin", SessionMeta{})
	if err != nil {
		t.Fatalf("second Login() error = %v", err)
	}

	admin := store.users["admin"]
	if err := svc.LogoutAll(context.Background(), admin.ID); err != nil {
		t.Fatalf("LogoutAll() error = %v", err)
	}
	if store.users["admin"].AuthVersion != 1 {
		t.Fatalf("auth_version = %d, want 1", store.users["admin"].AuthVersion)
	}
	if _, err := svc.Refresh(context.Background(), first.RefreshToken, SessionMeta{}); err == nil {
		t.Fatal("first Refresh() succeeded after LogoutAll")
	}
	if _, err := svc.Refresh(context.Background(), second.RefreshToken, SessionMeta{}); err == nil {
		t.Fatal("second Refresh() succeeded after LogoutAll")
	}
}

func TestAuthUpdateSessionLabelUpdatesOwnedSession(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}

	pair, err := svc.Login(context.Background(), "admin", "admin", SessionMeta{})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(pair.RefreshToken, claims); err != nil {
		t.Fatalf("ParseUnverified(refresh) error = %v", err)
	}
	rawSID, _ := claims["ssid"].(string)
	var sid pgtype.UUID
	if err := sid.Scan(rawSID); err != nil {
		t.Fatalf("sid Scan() error = %v", err)
	}

	admin := store.users["admin"]
	if err := svc.UpdateSessionLabel(context.Background(), admin.ID, sid, "Laptop Ruang Guru"); err != nil {
		t.Fatalf("UpdateSessionLabel() error = %v", err)
	}

	session, err := store.GetAuthSession(context.Background(), sid)
	if err != nil {
		t.Fatalf("GetAuthSession() error = %v", err)
	}
	if session.DeviceLabel != "Laptop Ruang Guru" {
		t.Fatalf("session.DeviceLabel = %q, want %q", session.DeviceLabel, "Laptop Ruang Guru")
	}
}

func TestAuthValidateAccessSessionRejectsRevokedSession(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}

	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}

	pair, err := svc.Login(context.Background(), "admin", "admin", SessionMeta{})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(pair.AccessToken, claims); err != nil {
		t.Fatalf("ParseUnverified(access) error = %v", err)
	}
	sub, _ := claims["sub"].(string)
	ssid, _ := claims["ssid"].(string)

	ok, err := svc.ValidateAccessSession(context.Background(), sub, ssid)
	if err != nil || !ok {
		t.Fatalf("ValidateAccessSession() before revoke = (%v, %v), want (true, nil)", ok, err)
	}

	if err := svc.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	ok, err = svc.ValidateAccessSession(context.Background(), sub, ssid)
	if err != nil {
		t.Fatalf("ValidateAccessSession() after revoke error = %v", err)
	}
	if ok {
		t.Fatal("ValidateAccessSession() = true after revoke, want false")
	}
}
