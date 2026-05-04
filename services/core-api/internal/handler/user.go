package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type userStore interface {
	ListUsers(ctx context.Context) ([]db.ListUsersRow, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error)
	AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error
	ListAuditLogs(ctx context.Context, arg db.ListAuditLogsParams) ([]db.ListAuditLogsRow, error)
}

type userLifecycleService interface {
	DeleteAsDeactivate(ctx context.Context, id pgtype.UUID) error
	UpdateStatus(ctx context.Context, id pgtype.UUID, isActive bool) error
}

type userTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type User struct {
	q         userStore
	lifecycle userLifecycleService
	tx        userTxStarter
}

func NewUser(q *db.Queries) *User { return &User{q: q, lifecycle: service.NewUserLifecycle(q)} }

func NewUserWithPool(pool *pgxpool.Pool) *User {
	return &User{q: db.New(pool), lifecycle: service.NewUserLifecycleWithPool(pool), tx: pool}
}

func (h *User) List(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.q.ListUsers(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}

	type userResponse struct {
		ID          pgtype.UUID        `json:"id"`
		Username    string             `json:"username"`
		EmployeeID  pgtype.UUID        `json:"employee_id"`
		StudentID   pgtype.UUID        `json:"student_id"`
		ParentID    pgtype.UUID        `json:"parent_id"`
		ProfileNama string             `json:"profile_nama"`
		CreatedAt   pgtype.Timestamptz `json:"created_at"`
		Roles       []string           `json:"roles"`
	}

	res := make([]userResponse, len(rows))
	for i, row := range rows {
		var roles []string
		if len(row.Roles) > 0 {
			_ = json.Unmarshal(row.Roles, &roles)
		}
		res[i] = userResponse{
			ID:          row.ID,
			Username:    row.Username,
			EmployeeID:  row.EmployeeID,
			StudentID:   row.StudentID,
			ParentID:    row.ParentID,
			ProfileNama: row.ProfileNama,
			CreatedAt:   row.CreatedAt,
			Roles:       roles,
		}
	}
	api.OK(w, res)
}

func (h *User) Create(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		Username   string   `json:"username"`
		Password   string   `json:"password"`
		Roles      []string `json:"roles"`
		EmployeeID string   `json:"employee_id"`
		StudentID  string   `json:"student_id"`
		ParentID   string   `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.Username == "" || body.Password == "" {
		api.BadRequest(w, "username and password required")
		return
	}
	if err := service.ValidatePassword(body.Username, body.Password); err != nil {
		writeClientError(w, err, "Password tidak memenuhi kebijakan keamanan")
		return
	}
	if len(body.Roles) == 0 {
		api.BadRequest(w, "at least one role required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		api.Internal(w, err)
		return
	}

	var empID, stuID, parID pgtype.UUID
	if body.EmployeeID != "" {
		_ = empID.Scan(body.EmployeeID)
	}
	if body.StudentID != "" {
		_ = stuID.Scan(body.StudentID)
	}
	if body.ParentID != "" {
		_ = parID.Scan(body.ParentID)
	}
	if err := validateUserCreate(body.Roles, empID, stuID, parID); err != nil {
		writeClientError(w, err, "Data pengguna tidak valid")
		return
	}

	create := func(store userStore) (db.CreateUserRow, error) {
		row, err := store.CreateUser(r.Context(), db.CreateUserParams{
			Username:     body.Username,
			PasswordHash: string(hash),
			EmployeeID:   empID,
			StudentID:    stuID,
			ParentID:     parID,
			IsActive:     true,
		})
		if err != nil {
			return db.CreateUserRow{}, err
		}
		for _, rStr := range body.Roles {
			if err := store.AddUserRole(r.Context(), db.AddUserRoleParams{
				UserID: row.ID,
				Role:   db.UserRole(rStr),
			}); err != nil {
				return db.CreateUserRow{}, err
			}
		}
		return row, nil
	}

	var row db.CreateUserRow
	if h.tx != nil {
		tx, err := h.tx.Begin(r.Context())
		if err != nil {
			api.Internal(w, err)
			return
		}
		defer tx.Rollback(r.Context())
		row, err = create(db.New(tx))
		if err != nil {
			api.Internal(w, err)
			return
		}
		if err := tx.Commit(r.Context()); err != nil {
			api.Internal(w, err)
			return
		}
	} else {
		row, err = create(h.q)
		if err != nil {
			api.Internal(w, err)
			return
		}
	}

	api.Created(w, row)
}

func validateUserCreate(roles []string, empID, stuID, parID pgtype.UUID) error {
	validRoles := []string{"admin", "guru", "staf", "kesiswaan", "siswa", "ortu"}
	for _, role := range roles {
		if !slices.Contains(validRoles, role) {
			return httpError("role tidak valid")
		}
	}

	hasEmployeeRole := slices.Contains(roles, "guru") || slices.Contains(roles, "staf") || slices.Contains(roles, "kesiswaan")
	hasStudentRole := slices.Contains(roles, "siswa")
	hasParentRole := slices.Contains(roles, "ortu")

	if hasEmployeeRole && !empID.Valid {
		return httpError("role guru/staf/kesiswaan wajib ditautkan ke pegawai")
	}
	if hasStudentRole && !stuID.Valid {
		return httpError("role siswa wajib ditautkan ke siswa")
	}
	if hasParentRole && !parID.Valid {
		return httpError("role ortu wajib ditautkan ke orang tua")
	}
	if empID.Valid && !hasEmployeeRole {
		return httpError("tautan pegawai hanya boleh untuk role guru/staf/kesiswaan")
	}
	if stuID.Valid && !hasStudentRole {
		return httpError("tautan siswa hanya boleh untuk role siswa")
	}
	if parID.Valid && !hasParentRole {
		return httpError("tautan orang tua hanya boleh untuk role ortu")
	}

	linkedProfiles := 0
	if empID.Valid {
		linkedProfiles++
	}
	if stuID.Valid {
		linkedProfiles++
	}
	if parID.Valid {
		linkedProfiles++
	}
	if linkedProfiles > 1 {
		return httpError("satu akun hanya boleh ditautkan ke satu jenis profil")
	}
	return nil
}

type httpError string

func (e httpError) Error() string { return string(e) }

func (h *User) Delete(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.lifecycle.DeleteAsDeactivate(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *User) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if err := h.lifecycle.UpdateStatus(r.Context(), id, body.IsActive); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"id":        id,
		"is_active": body.IsActive,
	})
}

func (h *User) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	q := r.URL.Query()
	limit := int32(pageSize(q.Get("per_page"), 100))
	page := pageNum(q.Get("page"), 1)
	offset := int32((page - 1) * int(limit))

	rows, err := h.q.ListAuditLogs(r.Context(), db.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}
