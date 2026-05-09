package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

const internalAnalyticsBodyLimit = 16 << 10

type InternalAnalytics struct {
	svc internalAnalyticsService
}

type internalAnalyticsService interface {
	CreateEvent(ctx context.Context, in service.CreateInternalAnalyticsEventInput) (service.InternalAnalyticsEventReceipt, error)
	CreatePublicEvent(ctx context.Context, in service.CreatePublicAnalyticsEventInput) (service.InternalAnalyticsEventReceipt, error)
	ListDailyAggregates(ctx context.Context, in service.InternalAnalyticsDailyQuery) (service.InternalAnalyticsDailyResult, error)
	Summary(ctx context.Context, in service.InternalAnalyticsSummaryQuery) (service.InternalAnalyticsSummary, error)
	ExportAggregatesCSV(ctx context.Context, in service.InternalAnalyticsExportQuery) (service.InternalAnalyticsExport, error)
}

func NewInternalAnalytics(svc *service.InternalAnalytics) *InternalAnalytics {
	return &InternalAnalytics{svc: svc}
}

type internalAnalyticsEventRequest struct {
	EventName          string          `json:"event_name"`
	EventGroup         string          `json:"event_group"`
	SourceSurface      string          `json:"source_surface"`
	PermissionCode     string          `json:"permission_code"`
	RouteGroup         string          `json:"route_group"`
	Module             string          `json:"module"`
	Result             string          `json:"result"`
	StatusCodeClass    string          `json:"status_code_class"`
	DurationBucket     string          `json:"duration_bucket"`
	Metadata           json.RawMessage `json:"metadata"`
	RetentionExpiresAt *time.Time      `json:"retention_expires_at"`
}

type publicAnalyticsEventRequest struct {
	EventName  string          `json:"event_name"`
	RouteGroup string          `json:"route_group"`
	Result     string          `json:"result"`
	Metadata   json.RawMessage `json:"metadata"`
}

func (h *InternalAnalytics) CreateEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}

	var body internalAnalyticsEventRequest
	metadata, ok := decodeInternalAnalyticsJSON(w, r, &body)
	if !ok {
		return
	}
	row, err := h.svc.CreateEvent(r.Context(), service.CreateInternalAnalyticsEventInput{
		EventName:          body.EventName,
		EventGroup:         body.EventGroup,
		SourceSurface:      body.SourceSurface,
		ActorUserID:        internalAnalyticsActorUserID(claims),
		ActorRole:          internalAnalyticsActorRole(claims),
		PermissionCode:     body.PermissionCode,
		RouteGroup:         body.RouteGroup,
		Module:             body.Module,
		Result:             body.Result,
		StatusCodeClass:    body.StatusCodeClass,
		DurationBucket:     body.DurationBucket,
		Metadata:           metadata,
		RetentionExpiresAt: body.RetentionExpiresAt,
	})
	if err != nil {
		var validationErr *service.InternalAnalyticsValidationError
		if errors.As(err, &validationErr) {
			api.BadRequest(w, internalAnalyticsValidationMessage(validationErr.Code))
			return
		}
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *InternalAnalytics) CreatePublicEvent(w http.ResponseWriter, r *http.Request) {
	var body publicAnalyticsEventRequest
	metadata, ok := decodePublicAnalyticsJSON(w, r, &body)
	if !ok {
		return
	}
	row, err := h.svc.CreatePublicEvent(r.Context(), service.CreatePublicAnalyticsEventInput{
		EventName:  body.EventName,
		RouteGroup: body.RouteGroup,
		Result:     body.Result,
		Metadata:   metadata,
	})
	if err != nil {
		var validationErr *service.InternalAnalyticsValidationError
		if errors.As(err, &validationErr) {
			api.BadRequest(w, internalAnalyticsValidationMessage(validationErr.Code))
			return
		}
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *InternalAnalytics) ListDailyAggregates(w http.ResponseWriter, r *http.Request) {
	query, ok := parseInternalAnalyticsDailyQuery(w, r)
	if !ok {
		return
	}
	result, err := h.svc.ListDailyAggregates(r.Context(), query)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, result)
}

func (h *InternalAnalytics) Summary(w http.ResponseWriter, r *http.Request) {
	days, ok := parseInternalAnalyticsPositiveInt(w, r, "days", 30, 180)
	if !ok {
		return
	}
	result, err := h.svc.Summary(r.Context(), service.InternalAnalyticsSummaryQuery{
		Days:          int32(days),
		EventGroup:    strings.TrimSpace(r.URL.Query().Get("event_group")),
		SourceSurface: strings.TrimSpace(r.URL.Query().Get("source_surface")),
		Role:          strings.TrimSpace(r.URL.Query().Get("role")),
		Result:        strings.TrimSpace(r.URL.Query().Get("result")),
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, result)
}

func (h *InternalAnalytics) ExportAggregates(w http.ResponseWriter, r *http.Request) {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	query, ok := parseInternalAnalyticsDailyQuery(w, r)
	if !ok {
		return
	}
	exported, err := h.svc.ExportAggregatesCSV(r.Context(), service.InternalAnalyticsExportQuery{
		EventGroup:    query.EventGroup,
		EventName:     query.EventName,
		SourceSurface: query.SourceSurface,
		Role:          query.Role,
		Result:        query.Result,
		Days:          query.Days,
		Limit:         query.Limit,
		Offset:        query.Offset,
		ActorUserID:   internalAnalyticsActorUserID(claims),
		ActorRole:     internalAnalyticsActorRole(claims),
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	filename := strings.ReplaceAll(exported.Filename, `"`, "")
	if filename == "" {
		filename = "internal-analytics-aggregate.csv"
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(exported.Content)
}

func parseInternalAnalyticsDailyQuery(w http.ResponseWriter, r *http.Request) (service.InternalAnalyticsDailyQuery, bool) {
	days, ok := parseInternalAnalyticsPositiveInt(w, r, "days", 30, 180)
	if !ok {
		return service.InternalAnalyticsDailyQuery{}, false
	}
	limit, ok := parseInternalAnalyticsPositiveInt(w, r, "limit", 500, 10000)
	if !ok {
		return service.InternalAnalyticsDailyQuery{}, false
	}
	offset, ok := parseInternalAnalyticsNonNegativeInt(w, r, "offset", 0, 100000)
	if !ok {
		return service.InternalAnalyticsDailyQuery{}, false
	}
	values := r.URL.Query()
	return service.InternalAnalyticsDailyQuery{
		EventGroup:    strings.TrimSpace(values.Get("event_group")),
		EventName:     strings.TrimSpace(values.Get("event_name")),
		SourceSurface: strings.TrimSpace(values.Get("source_surface")),
		Role:          strings.TrimSpace(values.Get("role")),
		Result:        strings.TrimSpace(values.Get("result")),
		Days:          int32(days),
		Limit:         int32(limit),
		Offset:        int32(offset),
	}, true
}

func parseInternalAnalyticsPositiveInt(w http.ResponseWriter, r *http.Request, key string, fallback int, max int) (int, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 || value > max {
		api.BadRequest(w, key+" tidak valid")
		return 0, false
	}
	return value, true
}

func parseInternalAnalyticsNonNegativeInt(w http.ResponseWriter, r *http.Request, key string, fallback int, max int) (int, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 || value > max {
		api.BadRequest(w, key+" tidak valid")
		return 0, false
	}
	return value, true
}

func decodeInternalAnalyticsJSON(w http.ResponseWriter, r *http.Request, body *internalAnalyticsEventRequest) (map[string]any, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, internalAnalyticsBodyLimit)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		writeInternalAnalyticsDecodeError(w, err)
		return nil, false
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		writeInternalAnalyticsDecodeError(w, err)
		return nil, false
	}
	if !jsonRawObject(raw) {
		api.BadRequest(w, "request body must be a json object")
		return nil, false
	}

	bodyDec := json.NewDecoder(bytes.NewReader(raw))
	bodyDec.DisallowUnknownFields()
	if err := bodyDec.Decode(body); err != nil {
		writeInternalAnalyticsDecodeError(w, err)
		return nil, false
	}

	metadata := map[string]any{}
	if len(body.Metadata) == 0 {
		return metadata, true
	}
	if !jsonRawObject(body.Metadata) {
		api.BadRequest(w, "metadata must be a json object")
		return nil, false
	}
	metadataDec := json.NewDecoder(bytes.NewReader(body.Metadata))
	if err := metadataDec.Decode(&metadata); err != nil {
		writeInternalAnalyticsDecodeError(w, err)
		return nil, false
	}
	return metadata, true
}

func decodePublicAnalyticsJSON(w http.ResponseWriter, r *http.Request, body *publicAnalyticsEventRequest) (map[string]any, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		writeInternalAnalyticsDecodeError(w, err)
		return nil, false
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		writeInternalAnalyticsDecodeError(w, err)
		return nil, false
	}
	if !jsonRawObject(raw) {
		api.BadRequest(w, "request body must be a json object")
		return nil, false
	}

	bodyDec := json.NewDecoder(bytes.NewReader(raw))
	bodyDec.DisallowUnknownFields()
	if err := bodyDec.Decode(body); err != nil {
		writeInternalAnalyticsDecodeError(w, err)
		return nil, false
	}

	metadata := map[string]any{}
	if len(body.Metadata) == 0 {
		return metadata, true
	}
	if !jsonRawObject(body.Metadata) {
		api.BadRequest(w, "metadata must be a json object")
		return nil, false
	}
	metadataDec := json.NewDecoder(bytes.NewReader(body.Metadata))
	if err := metadataDec.Decode(&metadata); err != nil {
		writeInternalAnalyticsDecodeError(w, err)
		return nil, false
	}
	return metadata, true
}

func writeInternalAnalyticsDecodeError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		api.Err(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	api.BadRequest(w, "invalid json")
}

func jsonRawObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) > 0 && trimmed[0] == '{'
}

func internalAnalyticsActorUserID(claims jwt.MapClaims) pgtype.UUID {
	for _, key := range []string{"sub", "uid"} {
		if raw, ok := claims[key].(string); ok {
			var id pgtype.UUID
			if err := id.Scan(strings.TrimSpace(raw)); err == nil && id.Valid {
				return id
			}
		}
	}
	return pgtype.UUID{}
}

func internalAnalyticsActorRole(claims jwt.MapClaims) string {
	if rawRoles, ok := claims["roles"].([]any); ok {
		for _, rawRole := range rawRoles {
			if role, ok := rawRole.(string); ok && strings.TrimSpace(role) != "" {
				return strings.TrimSpace(role)
			}
		}
	}
	if rawRoles, ok := claims["roles"].([]string); ok {
		for _, role := range rawRoles {
			if strings.TrimSpace(role) != "" {
				return strings.TrimSpace(role)
			}
		}
	}
	if role, ok := claims["role"].(string); ok {
		return strings.TrimSpace(role)
	}
	return ""
}

func internalAnalyticsValidationMessage(code string) string {
	switch code {
	case service.InternalAnalyticsErrEventNotAllowlisted:
		return "internal analytics event is not allowlisted"
	case service.InternalAnalyticsErrGroupMismatch:
		return "internal analytics event_group does not match event_name"
	case service.InternalAnalyticsErrForbiddenMetadataKey:
		return "internal analytics metadata contains forbidden sensitive key"
	case service.InternalAnalyticsErrMetadataTooLarge:
		return "internal analytics metadata is too large"
	case service.InternalAnalyticsErrInvalidSourceSurface:
		return "internal analytics source_surface is invalid"
	case service.InternalAnalyticsErrInvalidRetention:
		return "internal analytics retention_expires_at is invalid"
	default:
		return "internal analytics payload is invalid"
	}
}
