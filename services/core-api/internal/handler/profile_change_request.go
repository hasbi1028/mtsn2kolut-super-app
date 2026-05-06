package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

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
	ListAdmin(ctx context.Context, filter service.ProfileChangeRequestListFilter, limit, offset int32) ([]service.ProfileChangeRequestListItem, error)
	CountAdmin(ctx context.Context, filter service.ProfileChangeRequestListFilter) (int32, error)
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
	FieldLabel           string `json:"field_label,omitempty"`
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
	rows, err := h.svc.ListAdmin(r.Context(), profileChangeRequestListFilterFromQuery(q), limit, offset)
	if err != nil {
		writeDomainOrInternal(w, err, "daftar permintaan perubahan data tidak valid")
		return
	}
	api.OK(w, profileChangeRequestResponses(rows))
}

func (h *ProfileChangeRequest) CountPendingAdmin(w http.ResponseWriter, r *http.Request) {
	filter := profileChangeRequestListFilterFromQuery(r.URL.Query())
	filter.Status = string(db.ProfileChangeRequestStatusPending)
	count, err := h.svc.CountAdmin(r.Context(), filter)
	if err != nil {
		writeDomainOrInternal(w, err, "jumlah permintaan perubahan data tidak valid")
		return
	}
	api.OK(w, map[string]int32{"pending": count})
}

func (h *ProfileChangeRequest) ExportAdminCSV(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.ListAdmin(r.Context(), profileChangeRequestListFilterFromQuery(r.URL.Query()), 5000, 0)
	if err != nil {
		writeDomainOrInternal(w, err, "export permintaan perubahan data tidak valid")
		return
	}

	filename := "profile-change-requests-" + time.Now().UTC().Format("20060102T150405Z") + ".csv"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	writer := csv.NewWriter(w)
	_ = writer.Write([]string{
		"request_id",
		"status",
		"created_at",
		"profile_type",
		"profile_name",
		"requester",
		"requester_username",
		"field",
		"current_value_masked",
		"requested_value_masked",
		"reason_summary",
		"reviewer_username",
		"review_note_summary",
		"reviewed_at",
		"updated_at",
	})
	for _, row := range rows {
		_ = writer.Write([]string{
			safeProfileChangeCSVCell(pgUUIDString(row.ID)),
			safeProfileChangeCSVCell(string(row.Status)),
			safeProfileChangeCSVCell(timestamptzRFC3339(row.CreatedAt)),
			safeProfileChangeCSVCell(row.ProfileType),
			safeProfileChangeCSVCell(row.ProfileNama),
			safeProfileChangeCSVCell(row.RequesterDisplayName),
			safeProfileChangeCSVCell(row.RequesterUsername),
			safeProfileChangeCSVCell(row.FieldLabel),
			safeProfileChangeCSVCell(maskProfileChangeExportValue(row.FieldKey, row.CurrentValue)),
			safeProfileChangeCSVCell(maskProfileChangeExportValue(row.FieldKey, row.RequestedValue)),
			safeProfileChangeCSVCell(safeProfileChangeExportSummary(row.Reason)),
			safeProfileChangeCSVCell(row.ReviewerUsername),
			safeProfileChangeCSVCell(safeProfileChangeExportSummary(row.ReviewNote)),
			safeProfileChangeCSVCell(timestamptzRFC3339(row.ReviewedAt)),
			safeProfileChangeCSVCell(timestamptzRFC3339(row.UpdatedAt)),
		})
	}
	writer.Flush()
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

func profileChangeRequestListFilterFromQuery(q url.Values) service.ProfileChangeRequestListFilter {
	fieldKey := q.Get("field")
	if fieldKey == "" {
		fieldKey = q.Get("field_key")
	}
	return service.ProfileChangeRequestListFilter{
		Status:      q.Get("status"),
		ProfileType: q.Get("profile_type"),
		FieldKey:    fieldKey,
		Search:      q.Get("search"),
	}
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
		FieldLabel:           item.FieldLabel,
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

var profileChangeExportSensitiveNumberPattern = regexp.MustCompile(`\d{4,}`)

func safeProfileChangeCSVCell(value string) string {
	if value == "" {
		return ""
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}

func safeProfileChangeExportSummary(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return ""
	}
	value = profileChangeExportSensitiveNumberPattern.ReplaceAllString(value, "****")
	if utf8.RuneCountInString(value) <= 140 {
		return value
	}
	runes := []rune(value)
	return string(runes[:140]) + "..."
}

func maskProfileChangeExportValue(fieldKey, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.TrimSpace(strings.ToLower(fieldKey)) == "tanggal_lahir" {
		if len(value) >= 4 {
			return value[:4] + "-**-**"
		}
		return "****-**-**"
	}
	words := strings.Fields(value)
	if len(words) == 0 {
		return ""
	}
	masked := make([]string, 0, len(words))
	for _, word := range words {
		masked = append(masked, maskProfileChangeWord(word))
	}
	return strings.Join(masked, " ")
}

func maskProfileChangeWord(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) == 1 {
		return "*"
	}
	maskLen := len(runes) - 1
	if maskLen > 6 {
		maskLen = 6
	}
	return string(runes[0]) + strings.Repeat("*", maskLen)
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
