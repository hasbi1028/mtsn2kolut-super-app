package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Notification struct {
	svc notificationService
}

type notificationService interface {
	List(ctx context.Context, userID pgtype.UUID, unreadOnly bool, limit, offset int32) ([]db.AppNotification, error)
	CountUnread(ctx context.Context, userID pgtype.UUID) (int64, error)
	MarkRead(ctx context.Context, userID, notificationID pgtype.UUID) (db.AppNotification, error)
	MarkAllRead(ctx context.Context, userID pgtype.UUID) (int64, error)
}

func NewNotification(svc *service.Notification) *Notification {
	return &Notification{svc: svc}
}

func (h *Notification) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := notificationUserID(w, r)
	if !ok {
		return
	}
	rows, err := h.svc.List(
		r.Context(),
		userID,
		boolQuery(r.URL.Query().Get("unread_only")),
		int32Query(r.URL.Query().Get("limit")),
		int32Query(r.URL.Query().Get("offset")),
	)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, notificationResponses(rows))
}

func (h *Notification) CountUnread(w http.ResponseWriter, r *http.Request) {
	userID, ok := notificationUserID(w, r)
	if !ok {
		return
	}
	count, err := h.svc.CountUnread(r.Context(), userID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]int64{"unread": count})
}

func (h *Notification) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := notificationUserID(w, r)
	if !ok {
		return
	}
	var notificationID pgtype.UUID
	if err := notificationID.Scan(chi.URLParam(r, "id")); err != nil {
		api.BadRequest(w, "id notifikasi tidak valid")
		return
	}
	row, err := h.svc.MarkRead(r.Context(), userID, notificationID)
	if errors.Is(err, domain.ErrNotFound) {
		api.NotFound(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, notificationResponseFromRow(row))
}

func (h *Notification) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := notificationUserID(w, r)
	if !ok {
		return
	}
	count, err := h.svc.MarkAllRead(r.Context(), userID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]int64{"marked": count})
}

func notificationUserID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	userID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	return userID, true
}

type notificationResponse struct {
	ID         pgtype.UUID        `json:"id"`
	UserID     pgtype.UUID        `json:"user_id"`
	Category   string             `json:"category"`
	Title      string             `json:"title"`
	Body       string             `json:"body"`
	EntityType string             `json:"entity_type"`
	EntityID   string             `json:"entity_id"`
	LinkPath   string             `json:"link_path"`
	ReadAt     pgtype.Timestamptz `json:"read_at"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
}

func notificationResponses(rows []db.AppNotification) []notificationResponse {
	out := make([]notificationResponse, len(rows))
	for i, row := range rows {
		out[i] = notificationResponseFromRow(row)
	}
	return out
}

func notificationResponseFromRow(row db.AppNotification) notificationResponse {
	return notificationResponse{
		ID:         row.ID,
		UserID:     row.UserID,
		Category:   row.Category,
		Title:      row.Title,
		Body:       row.Body,
		EntityType: row.EntityType,
		EntityID:   row.EntityID,
		LinkPath:   row.LinkPath,
		ReadAt:     row.ReadAt,
		CreatedAt:  row.CreatedAt,
	}
}
