package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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

	getUserByUsernameErr error
	getUserByIDErr       error
	createUserErr        error
	updatePasswordErr    error
	addRoleErr           error
	incrementVersionErr  error
	createSessionErr     error
	getSessionErr        error
	revokeSessionErr     error
	revokeAllErr         error
	listSessionsErr      error
	revokeOwnedErr       error
	updateLabelErr       error
	getPrefsErr          error
	upsertPrefsErr       error
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
	if f.getUserByUsernameErr != nil {
		return db.GetUserByUsernameRow{}, f.getUserByUsernameErr
	}
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
	if f.getUserByIDErr != nil {
		return db.GetUserByIDRow{}, f.getUserByIDErr
	}
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

func (f *fakeStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error) {
	if f.createUserErr != nil {
		return db.CreateUserRow{}, f.createUserErr
	}
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
	return db.CreateUserRow{
		ID:          u.ID,
		Username:    u.Username,
		EmployeeID:  u.EmployeeID,
		StudentID:   u.StudentID,
		ParentID:    u.ParentID,
		IsActive:    u.IsActive,
		AuthVersion: u.AuthVersion,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (f *fakeStore) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	if f.updatePasswordErr != nil {
		return f.updatePasswordErr
	}
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
	if f.addRoleErr != nil {
		return f.addRoleErr
	}
	f.userRoles[arg.UserID] = append(f.userRoles[arg.UserID], arg.Role)
	return nil
}

func (f *fakeStore) IncrementUserAuthVersion(ctx context.Context, id pgtype.UUID) (int32, error) {
	if f.incrementVersionErr != nil {
		return 0, f.incrementVersionErr
	}
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
	if f.createSessionErr != nil {
		return db.AuthSession{}, f.createSessionErr
	}
	now := pgtype.Timestamptz{}
	_ = now.Scan(time.Now())
	session := db.AuthSession{
		ID:               arg.ID,
		UserID:           arg.UserID,
		RefreshTokenHash: arg.RefreshTokenHash,
		ExpiresAt:        arg.ExpiresAt,
		LastUsedAt:       now,
		IpAddress:        arg.IpAddress,
		UserAgent:        arg.UserAgent,
		DeviceLabel:      arg.DeviceLabel,
	}
	f.authSessions[arg.ID] = session
	return session, nil
}

func (f *fakeStore) GetAuthSession(ctx context.Context, id pgtype.UUID) (db.AuthSession, error) {
	if f.getSessionErr != nil {
		return db.AuthSession{}, f.getSessionErr
	}
	session, ok := f.authSessions[id]
	if !ok {
		return db.AuthSession{}, pgx.ErrNoRows
	}
	return session, nil
}

func (f *fakeStore) RevokeAuthSession(ctx context.Context, id pgtype.UUID) error {
	if f.revokeSessionErr != nil {
		return f.revokeSessionErr
	}
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

func (f *fakeStore) RevokeLiveAuthSessionByHash(ctx context.Context, arg db.RevokeLiveAuthSessionByHashParams) (int64, error) {
	if f.revokeSessionErr != nil {
		return 0, f.revokeSessionErr
	}
	session, ok := f.authSessions[arg.ID]
	if !ok || session.RevokedAt.Valid || !session.ExpiresAt.Valid || session.ExpiresAt.Time.Before(time.Now()) || session.RefreshTokenHash != arg.RefreshTokenHash {
		return 0, nil
	}
	now := pgtype.Timestamptz{}
	_ = now.Scan(time.Now())
	session.RevokedAt = now
	f.authSessions[arg.ID] = session
	return 1, nil
}

func (f *fakeStore) TouchAuthSessionLastUsed(ctx context.Context, id pgtype.UUID) (int64, error) {
	session, ok := f.authSessions[id]
	if !ok || session.RevokedAt.Valid || !session.ExpiresAt.Valid || session.ExpiresAt.Time.Before(time.Now()) {
		return 0, nil
	}
	now := pgtype.Timestamptz{}
	_ = now.Scan(time.Now())
	session.LastUsedAt = now
	f.authSessions[id] = session
	return 1, nil
}

func (f *fakeStore) RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error) {
	if f.revokeAllErr != nil {
		return 0, f.revokeAllErr
	}
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
	if f.listSessionsErr != nil {
		return nil, f.listSessionsErr
	}
	items := make([]db.AuthSession, 0)
	for _, session := range f.authSessions {
		if session.UserID == userID && !session.RevokedAt.Valid {
			items = append(items, session)
		}
	}
	return items, nil
}

func (f *fakeStore) RevokeOwnedAuthSession(ctx context.Context, arg db.RevokeOwnedAuthSessionParams) (int64, error) {
	if f.revokeOwnedErr != nil {
		return 0, f.revokeOwnedErr
	}
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
	if f.updateLabelErr != nil {
		return 0, f.updateLabelErr
	}
	session, ok := f.authSessions[arg.ID]
	if !ok || session.UserID != arg.UserID || session.RevokedAt.Valid {
		return 0, nil
	}
	session.DeviceLabel = arg.DeviceLabel
	f.authSessions[arg.ID] = session
	return 1, nil
}

func (f *fakeStore) GetUserUIPreferences(ctx context.Context, userID pgtype.UUID) (db.UserUiPreference, error) {
	if f.getPrefsErr != nil {
		return db.UserUiPreference{}, f.getPrefsErr
	}
	prefs, ok := f.uiPrefs[userID]
	if !ok {
		return db.UserUiPreference{}, pgx.ErrNoRows
	}
	return prefs, nil
}

func (f *fakeStore) UpsertUserUIPreferences(ctx context.Context, arg db.UpsertUserUIPreferencesParams) (db.UserUiPreference, error) {
	if f.upsertPrefsErr != nil {
		return db.UserUiPreference{}, f.upsertPrefsErr
	}
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

func TestAuthSessionManagementAndVersionHelpers(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "admin"}
	if err := svc.SeedAdmin(context.Background()); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	pair, err := svc.Login(context.Background(), "admin", "admin", SessionMeta{DeviceLabel: " Laptop "})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	admin := store.users["admin"]
	if version, err := svc.CurrentAuthVersion(context.Background(), admin.ID.String()); err != nil || version != 0 {
		t.Fatalf("CurrentAuthVersion() = %d/%v, want 0 nil", version, err)
	}
	if _, err := svc.CurrentAuthVersion(context.Background(), "bad"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("CurrentAuthVersion(bad) = %v, want unauthorized", err)
	}
	if _, err := svc.CurrentAuthVersion(context.Background(), "99999999-9999-9999-9999-999999999999"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("CurrentAuthVersion(missing) = %v, want unauthorized", err)
	}

	sessions, err := svc.ListActiveSessions(context.Background(), admin.ID)
	if err != nil || len(sessions) != 1 {
		t.Fatalf("ListActiveSessions() = %d/%v, want one session", len(sessions), err)
	}
	if sessions[0].DeviceLabel != "Laptop" {
		t.Fatalf("session device label = %q, want trimmed Laptop", sessions[0].DeviceLabel)
	}
	if err := svc.UpdateSessionLabel(context.Background(), admin.ID, sessions[0].ID, " "); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("UpdateSessionLabel(blank) = %v, want bad request", err)
	}
	if err := svc.UpdateSessionLabel(context.Background(), documentCycleTestUUID(1), sessions[0].ID, "HP"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("UpdateSessionLabel(non-owned) = %v, want not found", err)
	}
	if err := svc.RevokeSession(context.Background(), documentCycleTestUUID(1), sessions[0].ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("RevokeSession(non-owned) = %v, want not found", err)
	}
	if err := svc.RevokeSession(context.Background(), admin.ID, sessions[0].ID); err != nil {
		t.Fatalf("RevokeSession(owned) error = %v", err)
	}
	if active, err := svc.ListActiveSessions(context.Background(), admin.ID); err != nil || len(active) != 0 {
		t.Fatalf("ListActiveSessions(after revoke) = %d/%v, want none", len(active), err)
	}
	if err := svc.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("Logout(revoked token) error = %v, want nil", err)
	}
}

func TestAuthValidateAccessSessionRejectsInvalidExpiredAndWrongUser(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret")}
	userID := documentCycleTestUUID(11)
	otherUserID := documentCycleTestUUID(12)
	sessionID := documentCycleTestUUID(13)
	if ok, err := svc.ValidateAccessSession(context.Background(), "bad", sessionID.String()); !errors.Is(err, domain.ErrUnauthorized) || ok {
		t.Fatalf("ValidateAccessSession(bad subject) = %v/%v, want unauthorized false", ok, err)
	}
	if ok, err := svc.ValidateAccessSession(context.Background(), userID.String(), "bad"); !errors.Is(err, domain.ErrUnauthorized) || ok {
		t.Fatalf("ValidateAccessSession(bad session) = %v/%v, want unauthorized false", ok, err)
	}
	if ok, err := svc.ValidateAccessSession(context.Background(), userID.String(), sessionID.String()); !errors.Is(err, domain.ErrUnauthorized) || ok {
		t.Fatalf("ValidateAccessSession(missing) = %v/%v, want unauthorized false", ok, err)
	}

	store.authSessions[sessionID] = db.AuthSession{ID: sessionID, UserID: userID}
	if ok, err := svc.ValidateAccessSession(context.Background(), userID.String(), sessionID.String()); err != nil || ok {
		t.Fatalf("ValidateAccessSession(no expiry) = %v/%v, want false nil", ok, err)
	}
	store.authSessions[sessionID] = db.AuthSession{ID: sessionID, UserID: userID, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}}
	if ok, err := svc.ValidateAccessSession(context.Background(), userID.String(), sessionID.String()); err != nil || ok {
		t.Fatalf("ValidateAccessSession(expired) = %v/%v, want false nil", ok, err)
	}
	store.authSessions[sessionID] = db.AuthSession{ID: sessionID, UserID: userID, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}, RevokedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}
	if ok, err := svc.ValidateAccessSession(context.Background(), userID.String(), sessionID.String()); err != nil || ok {
		t.Fatalf("ValidateAccessSession(revoked) = %v/%v, want false nil", ok, err)
	}
	store.authSessions[sessionID] = db.AuthSession{ID: sessionID, UserID: otherUserID, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}}
	if ok, err := svc.ValidateAccessSession(context.Background(), userID.String(), sessionID.String()); err != nil || ok {
		t.Fatalf("ValidateAccessSession(wrong user) = %v/%v, want false nil", ok, err)
	}
}

func TestAuthSidebarDecodeAndUtilityHelpers(t *testing.T) {
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret")}
	userID := documentCycleTestUUID(21)
	store.uiPrefs[userID] = db.UserUiPreference{UserID: userID, SidebarPinned: []byte(`{`)}
	if _, err := svc.GetSidebarPreferences(context.Background(), userID); err == nil {
		t.Fatal("GetSidebarPreferences(invalid json) error = nil, want json error")
	}

	tests := []struct {
		input any
		want  int64
		ok    bool
	}{
		{input: int64(7), want: 7, ok: true},
		{input: 8, want: 8, ok: true},
		{input: float64(9), want: 9, ok: true},
		{input: "10", want: 10, ok: true},
		{input: "bad", ok: false},
		{input: true, ok: false},
	}
	for _, tt := range tests {
		got, ok := asInt64(tt.input)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("asInt64(%#v) = %d/%v, want %d/%v", tt.input, got, ok, tt.want, tt.ok)
		}
	}

	if _, err := userIDFromClaims(jwt.MapClaims{}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("userIDFromClaims(empty) = %v, want unauthorized", err)
	}
	if _, err := userIDFromClaims(jwt.MapClaims{"sub": "bad"}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("userIDFromClaims(bad) = %v, want unauthorized", err)
	}
	if id, err := userIDFromClaims(jwt.MapClaims{"sub": userID.String()}); err != nil || id != userID {
		t.Fatalf("userIDFromClaims(valid) = %v/%v, want user id", id, err)
	}

	if err := validatePassword("admin", "admin"); !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("validatePassword(username) = %v, want weak password", err)
	}
	if err := validatePassword("admin", "abc12345"); err != nil {
		t.Fatalf("validatePassword(valid) = %v, want nil", err)
	}

	meta := normalizeSessionMeta(SessionMeta{IPAddress: " 1.2.3.4 ", UserAgent: " Mozilla/5.0 (Windows NT) Chrome/120 ", DeviceLabel: " "})
	if meta.IPAddress != "1.2.3.4" || meta.DeviceLabel != "Chrome di Windows" {
		t.Fatalf("normalizeSessionMeta() = %+v, want trimmed Chrome Windows label", meta)
	}
	for _, tt := range []struct {
		ua   string
		want string
	}{
		{ua: "Mozilla/5.0 (Android) Firefox/120", want: "Firefox di Android"},
		{ua: "Mozilla/5.0 (iPhone) Version/17 Safari/604", want: "Safari di iPhone/iPad"},
		{ua: "Mozilla/5.0 (Macintosh) Edg/120", want: "Edge di Mac"},
		{ua: "curl/8.0", want: "Browser di Perangkat Tidak Dikenal"},
	} {
		if got := deriveDeviceLabel(tt.ua); got != tt.want {
			t.Fatalf("deriveDeviceLabel(%q) = %q, want %q", tt.ua, got, tt.want)
		}
	}
}

func TestAuthAdditionalErrorBranches(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("store failed")

	if err := (&Auth{q: newFakeStore(), jwtSecret: []byte("secret")}).SeedAdmin(ctx); err == nil || !strings.Contains(err.Error(), "ADMIN_PASSWORD") {
		t.Fatalf("SeedAdmin(empty password) = %v, want required password error", err)
	}
	if err := (&Auth{q: &fakeStore{
		settings:             map[string]string{},
		users:                map[string]db.User{},
		userRoles:            map[pgtype.UUID][]db.UserRole{},
		authSessions:         map[pgtype.UUID]db.AuthSession{},
		uiPrefs:              map[pgtype.UUID]db.UserUiPreference{},
		getUserByUsernameErr: expectedErr,
	}, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}).SeedAdmin(ctx); !errors.Is(err, expectedErr) {
		t.Fatalf("SeedAdmin(get user error) = %v, want %v", err, expectedErr)
	}
	store := newFakeStore()
	store.createUserErr = expectedErr
	if err := (&Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}).SeedAdmin(ctx); !errors.Is(err, expectedErr) {
		t.Fatalf("SeedAdmin(create user error) = %v, want %v", err, expectedErr)
	}
	store = newFakeStore()
	store.addRoleErr = expectedErr
	if err := (&Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}).SeedAdmin(ctx); !errors.Is(err, expectedErr) {
		t.Fatalf("SeedAdmin(add role error) = %v, want %v", err, expectedErr)
	}

	store = newFakeStore()
	store.getUserByUsernameErr = expectedErr
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}
	if _, err := svc.Login(ctx, "admin", "adminpass123", SessionMeta{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Login(store error) = %v, want %v", err, expectedErr)
	}

	store = newFakeStore()
	svc = &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}
	if _, err := svc.Login(ctx, "missing", "password123", SessionMeta{}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Login(missing) = %v, want unauthorized", err)
	}
	if err := svc.SeedAdmin(ctx); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	if _, err := svc.Login(ctx, "admin", "wrongpass", SessionMeta{}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Login(wrong password) = %v, want unauthorized", err)
	}
	store.createSessionErr = expectedErr
	if _, err := svc.Login(ctx, "admin", "adminpass123", SessionMeta{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Login(create session error) = %v, want %v", err, expectedErr)
	}

	store = newFakeStore()
	svc = &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}
	if err := svc.SeedAdmin(ctx); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	pair, err := svc.Login(ctx, "admin", "adminpass123", SessionMeta{})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if _, err := svc.Refresh(ctx, pair.AccessToken, SessionMeta{}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Refresh(access token) = %v, want unauthorized", err)
	}
	store.getUserByIDErr = expectedErr
	if _, err := svc.Refresh(ctx, pair.RefreshToken, SessionMeta{}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Refresh(get user error) = %v, want unauthorized", err)
	}
	store.getUserByIDErr = nil
	admin := store.users["admin"]
	admin.IsActive = false
	store.users["admin"] = admin
	if _, err := svc.Refresh(ctx, pair.RefreshToken, SessionMeta{}); !errors.Is(err, domain.ErrSuspended) {
		t.Fatalf("Refresh(suspended) = %v, want suspended", err)
	}
	admin.IsActive = true
	store.users["admin"] = admin
	store.revokeSessionErr = expectedErr
	if _, err := svc.Refresh(ctx, pair.RefreshToken, SessionMeta{}); !errors.Is(err, expectedErr) {
		t.Fatalf("Refresh(revoke error) = %v, want %v", err, expectedErr)
	}

	store = newFakeStore()
	svc = &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}
	if err := svc.SeedAdmin(ctx); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	admin = store.users["admin"]
	store.revokeAllErr = expectedErr
	if err := svc.LogoutAll(ctx, admin.ID); !errors.Is(err, expectedErr) {
		t.Fatalf("LogoutAll(revoke all error) = %v, want %v", err, expectedErr)
	}
	store.revokeAllErr = nil
	store.incrementVersionErr = expectedErr
	if err := svc.LogoutAll(ctx, admin.ID); !errors.Is(err, expectedErr) {
		t.Fatalf("LogoutAll(increment error) = %v, want %v", err, expectedErr)
	}

	store.listSessionsErr = expectedErr
	if _, err := svc.ListActiveSessions(ctx, admin.ID); !errors.Is(err, expectedErr) {
		t.Fatalf("ListActiveSessions(error) = %v, want %v", err, expectedErr)
	}
	store.listSessionsErr = nil
	store.revokeOwnedErr = expectedErr
	if err := svc.RevokeSession(ctx, admin.ID, documentCycleTestUUID(31)); !errors.Is(err, expectedErr) {
		t.Fatalf("RevokeSession(error) = %v, want %v", err, expectedErr)
	}
	store.revokeOwnedErr = nil
	store.updateLabelErr = expectedErr
	if err := svc.UpdateSessionLabel(ctx, admin.ID, documentCycleTestUUID(31), "HP"); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateSessionLabel(error) = %v, want %v", err, expectedErr)
	}

	if got, err := svc.GetSidebarPreferences(ctx, documentCycleTestUUID(32)); err != nil || len(got.PinnedItems) != 0 || len(got.RecentItems) != 0 {
		t.Fatalf("GetSidebarPreferences(missing) = %+v/%v, want empty nil", got, err)
	}
	store.getPrefsErr = expectedErr
	if _, err := svc.GetSidebarPreferences(ctx, admin.ID); !errors.Is(err, expectedErr) {
		t.Fatalf("GetSidebarPreferences(error) = %v, want %v", err, expectedErr)
	}
	store.getPrefsErr = nil
	store.upsertPrefsErr = expectedErr
	if _, err := svc.UpdateSidebarPreferences(ctx, admin.ID, SidebarPreferences{PinnedItems: []string{"/settings"}}); !errors.Is(err, expectedErr) {
		t.Fatalf("UpdateSidebarPreferences(error) = %v, want %v", err, expectedErr)
	}

	if err := svc.ChangePassword(ctx, "missing", "oldpass123", "newpass123"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("ChangePassword(missing) = %v, want unauthorized", err)
	}
	if err := svc.ChangePassword(ctx, "admin", "wrongpass", "newpass123"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("ChangePassword(wrong old password) = %v, want unauthorized", err)
	}
	store.updatePasswordErr = expectedErr
	if err := svc.ChangePassword(ctx, "admin", "adminpass123", "newpass123"); !errors.Is(err, expectedErr) {
		t.Fatalf("ChangePassword(update error) = %v, want %v", err, expectedErr)
	}
	store.updatePasswordErr = nil
	store.incrementVersionErr = expectedErr
	if err := svc.ChangePassword(ctx, "admin", "adminpass123", "newpass123"); !errors.Is(err, expectedErr) {
		t.Fatalf("ChangePassword(increment error) = %v, want %v", err, expectedErr)
	}
}

func TestAuthTokenValidationHelperBranches(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret"), adminPassword: "adminpass123"}
	if _, err := svc.parseTokenClaims("not-a-token"); err == nil {
		t.Fatal("parseTokenClaims(invalid) error = nil, want error")
	}

	if _, err := svc.validateRefreshSession(ctx, "refresh", jwt.MapClaims{}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("validateRefreshSession(no ssid) = %v, want unauthorized", err)
	}
	if _, err := svc.validateRefreshSession(ctx, "refresh", jwt.MapClaims{"ssid": "bad"}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("validateRefreshSession(bad ssid) = %v, want unauthorized", err)
	}

	if err := svc.SeedAdmin(ctx); err != nil {
		t.Fatalf("SeedAdmin() error = %v", err)
	}
	pair, err := svc.Login(ctx, "admin", "adminpass123", SessionMeta{})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	claims, err := svc.parseTokenClaims(pair.RefreshToken)
	if err != nil {
		t.Fatalf("parseTokenClaims(refresh) error = %v", err)
	}
	session, err := svc.validateRefreshSession(ctx, pair.RefreshToken, claims)
	if err != nil {
		t.Fatalf("validateRefreshSession(valid) error = %v", err)
	}
	store.authSessions[session.ID] = db.AuthSession{ID: session.ID, UserID: session.UserID, ExpiresAt: session.ExpiresAt, RefreshTokenHash: "wrong"}
	if _, err := svc.validateRefreshSession(ctx, pair.RefreshToken, claims); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("validateRefreshSession(hash mismatch) = %v, want unauthorized", err)
	}
	store.authSessions[session.ID] = db.AuthSession{ID: session.ID, UserID: session.UserID, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}, RefreshTokenHash: hashToken(pair.RefreshToken)}
	if _, err := svc.validateRefreshSession(ctx, pair.RefreshToken, claims); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("validateRefreshSession(expired) = %v, want unauthorized", err)
	}
	store.authSessions[session.ID] = db.AuthSession{ID: session.ID, UserID: session.UserID, ExpiresAt: session.ExpiresAt, RevokedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}, RefreshTokenHash: hashToken(pair.RefreshToken)}
	if _, err := svc.validateRefreshSession(ctx, pair.RefreshToken, claims); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("validateRefreshSession(revoked) = %v, want unauthorized", err)
	}

	if svc.validAuthVersion(ctx, jwt.MapClaims{}) {
		t.Fatal("validAuthVersion(empty) = true, want false")
	}
	admin := store.users["admin"]
	if svc.validAuthVersion(ctx, jwt.MapClaims{"sub": admin.ID.String()}) {
		t.Fatal("validAuthVersion(missing version) = true, want false")
	}
	if svc.validAuthVersion(ctx, jwt.MapClaims{"sub": admin.ID.String(), "ver": "bad"}) {
		t.Fatal("validAuthVersion(bad version) = true, want false")
	}
	if svc.validAuthVersion(ctx, jwt.MapClaims{"sub": admin.ID.String(), "ver": int64(99)}) {
		t.Fatal("validAuthVersion(mismatch) = true, want false")
	}
	if !svc.validAuthVersion(ctx, jwt.MapClaims{"sub": admin.ID.String(), "ver": int64(admin.AuthVersion)}) {
		t.Fatal("validAuthVersion(valid) = false, want true")
	}
}
