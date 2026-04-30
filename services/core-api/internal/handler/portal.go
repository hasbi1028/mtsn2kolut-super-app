package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Portal struct {
	svc *service.Portal
}

func NewPortal(svc *service.Portal) *Portal { return &Portal{svc: svc} }

func (h *Portal) StudentMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	rawID, _ := claims["sid"].(string)
	if rawID == "" {
		api.Forbidden(w)
		return
	}
	var studentID pgtype.UUID
	if err := studentID.Scan(rawID); err != nil {
		api.Forbidden(w)
		return
	}
	student, parents, sessions, err := h.svc.StudentOverview(r.Context(), studentID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"student":  student,
		"parents":  parents,
		"sessions": sessions,
	})
}

func (h *Portal) ParentMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	rawID, _ := claims["pid"].(string)
	if rawID == "" {
		api.Forbidden(w)
		return
	}
	var parentID pgtype.UUID
	if err := parentID.Scan(rawID); err != nil {
		api.Forbidden(w)
		return
	}
	parent, children, err := h.svc.ParentOverview(r.Context(), parentID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"parent":   parent,
		"children": children,
	})
}
