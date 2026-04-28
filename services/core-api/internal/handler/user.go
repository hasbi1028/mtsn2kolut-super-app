package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type User struct {
	q *db.Queries
}

func NewUser(q *db.Queries) *User { return &User{q: q} }

func (h *User) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.q.ListUsers(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *User) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		Role       string `json:"role"`
		EmployeeID string `json:"employee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.Username == "" || body.Password == "" {
		api.BadRequest(w, "username and password required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		api.Internal(w, err)
		return
	}

	var empID pgtype.UUID
	if body.EmployeeID != "" {
		_ = empID.Scan(body.EmployeeID)
	}

	role := db.UserRole(body.Role)
	if role == "" {
		role = db.UserRoleGuru
	}

	row, err := h.q.CreateUser(r.Context(), db.CreateUserParams{
		Username:     body.Username,
		PasswordHash: string(hash),
		Role:         role,
		EmployeeID:   empID,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *User) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.q.DeleteUser(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

func (h *User) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
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

