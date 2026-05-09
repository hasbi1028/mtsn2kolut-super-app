package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
