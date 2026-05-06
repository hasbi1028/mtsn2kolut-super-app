package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type profileChangeRequestService interface {
	Create(ctx context.Context, requesterUserID pgtype.UUID, input service.CreateProfileChangeRequestInput) (db.ProfileChangeRequest, error)
	ListOwn(ctx context.Context, requesterUserID pgtype.UUID) ([]service.ProfileChangeRequestListItem, error)
	ListSelfRequestableFields(ctx context.Context, requesterUserID pgtype.UUID) ([]service.ProfileChangeRequestField, error)
	CancelOwn(ctx context.Context, requesterUserID, id pgtype.UUID) (db.ProfileChangeRequest, error)
	ListAdmin(ctx context.Context, status string, limit, offset int32) ([]service.ProfileChangeRequestListItem, error)
	Review(ctx context.Context, reviewerUserID, id pgtype.UUID, input service.ReviewProfileChangeRequestInput) (db.ProfileChangeRequest, error)
}

type ProfileChangeRequest struct {
	svc profileChangeRequestService
}

type profileChangeRequestResponse struct {
	ID                   string `json:"id"`
	RequesterUserID      string `json:"requester_user_id"`
	RequesterUsername    string `json:"requester_username,omitempty"`
	RequesterDisplayName string `json:"requester_display_name,omitempty"`
	ProfileType          string `json:"profile_type"`
	ProfileNama          string `json:"profile_nama,omitempty"`
	TargetEmployeeID     string `json:"target_employee_id,omitempty"`
	TargetStudentID      string `json:"target_student_id,omitempty"`
	TargetParentID       string `json:"target_parent_id,omitempty"`
	FieldKey             string `json:"field_key"`
	CurrentValue         string `json:"current_value"`
	RequestedValue       string `json:"requested_value"`
	Reason               string `json:"reason"`
	Status               string `json:"status"`
	ReviewerUserID       string `json:"reviewer_user_id,omitempty"`
	ReviewerUsername     string `json:"reviewer_username,omitempty"`
	ReviewNote           string `json:"review_note,omitempty"`
	ReviewedAt           string `json:"reviewed_at,omitempty"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

type profileChangeRequestFieldResponse struct {
	ProfileType        string `json:"profile_type"`
	FieldKey           string `json:"field_key"`
	Label              string `json:"label"`
	ValueType          string `json:"value_type"`
	SelfRequestable    bool   `json:"self_requestable"`
	ReviewerPermission string `json:"reviewer_permission,omitempty"`
	IsActive           bool   `json:"is_active"`
}

func NewProfileChangeRequest(svc profileChangeRequestService) *ProfileChangeRequest {
	return &ProfileChangeRequest{svc: svc}
}

func (h *ProfileChangeRequest) CreateOwn(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentRequestUserID(w, r)
	if !ok {
		return
	}
	var body struct {
		ProfileType    string `json:"profile_type"`
		FieldKey       string `json:"field_key"`
		RequestedValue string `json:"requested_value"`
		Reason         string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.Create(r.Context(), userID, service.CreateProfileChangeRequestInput{
		ProfileType:    body.ProfileType,
		FieldKey:       body.FieldKey,
		RequestedValue: body.RequestedValue,
		Reason:         body.Reason,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "permintaan perubahan data tidak valid")
		return
	}
	api.Created(w, profileChangeRequestResponseFromDB(row))
}

func (h *ProfileChangeRequest) ListOwn(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentRequestUserID(w, r)
	if !ok {
		return
	}
	rows, err := h.svc.ListOwn(r.Context(), userID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, profileChangeRequestResponses(rows))
}

func (h *ProfileChangeRequest) ListSelfRequestableFields(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentRequestUserID(w, r)
	if !ok {
		return
	}
	rows, err := h.svc.ListSelfRequestableFields(r.Context(), userID)
	if err != nil {
		writeDomainOrInternal(w, err, "daftar field perubahan data tidak valid")
		return
	}
	api.OK(w, profileChangeRequestFieldResponses(rows))
}

func (h *ProfileChangeRequest) CancelOwn(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentRequestUserID(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.CancelOwn(r.Context(), userID, id)
	if err != nil {
		writeDomainOrInternal(w, err, "permintaan perubahan data tidak dapat dibatalkan")
		return
	}
	api.OK(w, profileChangeRequestResponseFromDB(row))
}

func (h *ProfileChangeRequest) ListAdmin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := int32(pageSize(q.Get("per_page"), 50))
	page := pageNum(q.Get("page"), 1)
	offset := int32((page - 1) * int(limit))
	rows, err := h.svc.ListAdmin(r.Context(), q.Get("status"), limit, offset)
	if err != nil {
		writeDomainOrInternal(w, err, "daftar permintaan perubahan data tidak valid")
		return
	}
	api.OK(w, profileChangeRequestResponses(rows))
}

func (h *ProfileChangeRequest) Review(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := currentRequestUserID(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Status     string `json:"status"`
		ReviewNote string `json:"review_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.Review(r.Context(), reviewerID, id, service.ReviewProfileChangeRequestInput{
		Status:              body.Status,
		ReviewNote:          body.ReviewNote,
		ReviewerRoles:       currentClaimStrings(r, "roles", "role"),
		ReviewerPermissions: currentClaimStrings(r, "permissions", "permission"),
	})
	if err != nil {
		writeDomainOrInternal(w, err, "review permintaan perubahan data tidak valid")
		return
	}
	api.OK(w, profileChangeRequestResponseFromDB(row))
}

func currentRequestUserID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	userID, err := currentActorUUID(r)
	if err != nil {
		api.Unauthorized(w)
		return pgtype.UUID{}, false
	}
	return userID, true
}

func profileChangeRequestResponses(items []service.ProfileChangeRequestListItem) []profileChangeRequestResponse {
	responses := make([]profileChangeRequestResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, profileChangeRequestResponseFromListItem(item))
	}
	return responses
}

func profileChangeRequestFieldResponses(items []service.ProfileChangeRequestField) []profileChangeRequestFieldResponse {
	responses := make([]profileChangeRequestFieldResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, profileChangeRequestFieldResponse{
			ProfileType:        item.ProfileType,
			FieldKey:           item.FieldKey,
			Label:              item.Label,
			ValueType:          item.ValueType,
			SelfRequestable:    item.SelfRequestable,
			ReviewerPermission: item.ReviewerPermission,
			IsActive:           item.IsActive,
		})
	}
	return responses
}

func profileChangeRequestResponseFromListItem(item service.ProfileChangeRequestListItem) profileChangeRequestResponse {
	return profileChangeRequestResponse{
		ID:                   pgUUIDString(item.ID),
		RequesterUserID:      pgUUIDString(item.RequesterUserID),
		RequesterUsername:    item.RequesterUsername,
		RequesterDisplayName: item.RequesterDisplayName,
		ProfileType:          item.ProfileType,
		ProfileNama:          item.ProfileNama,
		TargetEmployeeID:     pgUUIDString(item.TargetEmployeeID),
		TargetStudentID:      pgUUIDString(item.TargetStudentID),
		TargetParentID:       pgUUIDString(item.TargetParentID),
		FieldKey:             item.FieldKey,
		CurrentValue:         item.CurrentValue,
		RequestedValue:       item.RequestedValue,
		Reason:               item.Reason,
		Status:               string(item.Status),
		ReviewerUserID:       pgUUIDString(item.ReviewerUserID),
		ReviewerUsername:     item.ReviewerUsername,
		ReviewNote:           item.ReviewNote,
		ReviewedAt:           timestamptzRFC3339(item.ReviewedAt),
		CreatedAt:            timestamptzRFC3339(item.CreatedAt),
		UpdatedAt:            timestamptzRFC3339(item.UpdatedAt),
	}
}

func currentClaimStrings(r *http.Request, arrayKey, scalarKey string) []string {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return nil
	}
	values := make([]string, 0)
	if raw, ok := claims[arrayKey].([]any); ok {
		for _, item := range raw {
			if value, ok := item.(string); ok && strings.TrimSpace(value) != "" {
				values = append(values, value)
			}
		}
	}
	if raw, ok := claims[arrayKey].([]string); ok {
		for _, value := range raw {
			if strings.TrimSpace(value) != "" {
				values = append(values, value)
			}
		}
	}
	if value, ok := claims[scalarKey].(string); ok && strings.TrimSpace(value) != "" {
		values = append(values, value)
	}
	return values
}

func profileChangeRequestResponseFromDB(row db.ProfileChangeRequest) profileChangeRequestResponse {
	return profileChangeRequestResponse{
		ID:               pgUUIDString(row.ID),
		RequesterUserID:  pgUUIDString(row.RequesterUserID),
		ProfileType:      row.ProfileType,
		TargetEmployeeID: pgUUIDString(row.TargetEmployeeID),
		TargetStudentID:  pgUUIDString(row.TargetStudentID),
		TargetParentID:   pgUUIDString(row.TargetParentID),
		FieldKey:         row.FieldKey,
		CurrentValue:     row.CurrentValue,
		RequestedValue:   row.RequestedValue,
		Reason:           row.Reason,
		Status:           string(row.Status),
		ReviewerUserID:   pgUUIDString(row.ReviewerUserID),
		ReviewNote:       row.ReviewNote,
		ReviewedAt:       timestamptzRFC3339(row.ReviewedAt),
		CreatedAt:        timestamptzRFC3339(row.CreatedAt),
		UpdatedAt:        timestamptzRFC3339(row.UpdatedAt),
	}
}
