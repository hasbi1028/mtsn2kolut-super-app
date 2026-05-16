package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtApproval struct {
	svc cbtApprovalService
}

type cbtApprovalService interface {
	List(ctx context.Context, in service.ListCbtApprovalInput) ([]db.ListCbtApprovalRecordsRow, error)
	Approve(ctx context.Context, in service.SaveCbtApprovalInput) (db.CbtApprovalRecord, error)
	Revoke(ctx context.Context, id pgtype.UUID, actorUserID pgtype.UUID, notes string) (db.CbtApprovalRecord, error)
}

func NewCbtApproval(svc *service.CbtApproval) *CbtApproval { return &CbtApproval{svc: svc} }

func (h *CbtApproval) List(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var entityID pgtype.UUID
	if raw := r.URL.Query().Get("entity_id"); raw != "" {
		parsed, err := parseUUID(raw)
		if err != nil {
			api.BadRequest(w, "ID data pengesahan tidak valid")
			return
		}
		entityID = parsed
	}
	rows, err := h.svc.List(r.Context(), service.ListCbtApprovalInput{
		EntityType: r.URL.Query().Get("entity_type"),
		EntityID:   entityID,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "Daftar pengesahan tidak dapat diakses")
		return
	}
	api.OK(w, rows)
}

func (h *CbtApproval) Approve(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	actorID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	var body struct {
		EntityType   string `json:"entity_type"`
		EntityID     string `json:"entity_id"`
		ApprovalType string `json:"approval_type"`
		Notes        string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data pengesahan tidak valid")
		return
	}
	entityID, err := parseUUID(body.EntityID)
	if err != nil {
		api.BadRequest(w, "ID data pengesahan tidak valid")
		return
	}
	row, err := h.svc.Approve(r.Context(), service.SaveCbtApprovalInput{
		EntityType:   body.EntityType,
		EntityID:     entityID,
		ApprovalType: body.ApprovalType,
		Notes:        body.Notes,
		ActorUserID:  actorID,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "Pengesahan tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *CbtApproval) Revoke(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	actorID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID pengesahan tidak valid")
		return
	}
	var body struct {
		Notes string `json:"notes"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	row, err := h.svc.Revoke(r.Context(), id, actorID, body.Notes)
	if err != nil {
		writeDomainOrInternal(w, err, "Cabut pengesahan tidak valid")
		return
	}
	api.OK(w, row)
}
