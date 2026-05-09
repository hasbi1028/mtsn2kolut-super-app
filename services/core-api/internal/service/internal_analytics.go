package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const (
	InternalAnalyticsErrEventNotAllowlisted  = "event_not_allowlisted"
	InternalAnalyticsErrGroupMismatch        = "event_group_mismatch"
	InternalAnalyticsErrForbiddenMetadataKey = "forbidden_metadata_key"
	InternalAnalyticsErrMetadataTooLarge     = "metadata_too_large"
	InternalAnalyticsErrInvalidSourceSurface = "invalid_source_surface"
	InternalAnalyticsErrInvalidMetadata      = "invalid_metadata"
	InternalAnalyticsErrInvalidRetention     = "invalid_retention"

	internalAnalyticsMetadataMaxBytes = 8 << 10
)

type internalAnalyticsStore interface {
	CreateInternalAnalyticsEvent(ctx context.Context, arg db.CreateInternalAnalyticsEventParams) (db.InternalAnalyticsEvent, error)
}

type InternalAnalytics struct {
	q   internalAnalyticsStore
	now func() time.Time
}

func NewInternalAnalytics(q *db.Queries) *InternalAnalytics {
	return &InternalAnalytics{q: q, now: time.Now}
}

type CreateInternalAnalyticsEventInput struct {
	EventName          string
	EventGroup         string
	SourceSurface      string
	ActorUserID        pgtype.UUID
	ActorRole          string
	PermissionCode     string
	RouteGroup         string
	Module             string
	Result             string
	StatusCodeClass    string
	DurationBucket     string
	Metadata           map[string]any
	RetentionExpiresAt *time.Time
}

type InternalAnalyticsEventReceipt struct {
	ID                 string    `json:"id"`
	EventName          string    `json:"event_name"`
	EventGroup         string    `json:"event_group"`
	OccurredAt         time.Time `json:"occurred_at"`
	RetentionExpiresAt time.Time `json:"retention_expires_at"`
}

type InternalAnalyticsValidationError struct {
	Code  string
	Field string
}

func (e *InternalAnalyticsValidationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Field)
	}
	return e.Code
}

var internalAnalyticsAllowedEvents = map[string]string{
	"public.page_view":                "public",
	"public.cta_click":                "public",
	"public.download":                 "public",
	"public.search":                   "public",
	"public.form_start":               "public",
	"public.form_submit":              "public",
	"auth.login_attempt":              "auth",
	"auth.login_success":              "auth",
	"auth.login_failed":               "auth",
	"auth.refresh_success":            "auth",
	"auth.logout":                     "auth",
	"auth.logout_all":                 "auth",
	"auth.session_revoke":             "auth",
	"auth.password_change":            "auth",
	"auth.suspended_block":            "auth",
	"dashboard.view":                  "dashboard",
	"dashboard.widget_view":           "dashboard",
	"dashboard.filter_change":         "dashboard",
	"dashboard.export":                "dashboard",
	"dashboard.refresh":               "dashboard",
	"bank_soal.list_view":             "bank_soal",
	"bank_soal.editor_open":           "bank_soal",
	"bank_soal.draft_save":            "bank_soal",
	"bank_soal.question_create":       "bank_soal",
	"bank_soal.question_update":       "bank_soal",
	"bank_soal.import_start":          "bank_soal",
	"bank_soal.import_complete":       "bank_soal",
	"bank_soal.review_decision":       "bank_soal",
	"bank_soal.readiness_check":       "bank_soal",
	"bank_soal.asset_upload":          "bank_soal",
	"asesmen.hub_view":                "asesmen",
	"asesmen.package_create":          "asesmen",
	"asesmen.package_update":          "asesmen",
	"asesmen.event_create":            "asesmen",
	"asesmen.session_create":          "asesmen",
	"asesmen.session_update":          "asesmen",
	"asesmen.proctoring_view":         "asesmen",
	"asesmen.participant_action":      "asesmen",
	"asesmen.result_view":             "asesmen",
	"asesmen.result_export":           "asesmen",
	"asesmen.non_test_sync":           "asesmen",
	"pusaka.dashboard_view":           "pusaka",
	"pusaka.job_claimed":              "pusaka",
	"pusaka.job_completed":            "pusaka",
	"pusaka.job_failed":               "pusaka",
	"pusaka.job_retried":              "pusaka",
	"pusaka.stale_recovered":          "pusaka",
	"pusaka.employee_scope_update":    "pusaka",
	"pusaka.settings_update":          "pusaka",
	"pusaka.worker_heartbeat":         "pusaka",
	"users.list_view":                 "users",
	"users.create":                    "users",
	"users.update":                    "users",
	"users.reset_password":            "users",
	"users.suspend_change":            "users",
	"rbac.roles_view":                 "rbac",
	"rbac.permission_update":          "rbac",
	"rbac.permission_denied":          "rbac",
	"security.rate_limited":           "security",
	"security.forbidden":              "security",
	"security.validation_rejected":    "security",
	"security.sensitive_key_rejected": "security",
	"security.suspicious_pattern":     "security",
	"security.export_requested":       "security",
	"security.internal_error_class":   "security",
}

var internalAnalyticsAllowedSourceSurfaces = map[string]struct{}{
	"public_website": {},
	"web_admin":      {},
	"core_api":       {},
	"mobile_app":     {},
	"system":         {},
}

var internalAnalyticsForbiddenMetadataKeys = map[string]struct{}{
	"password":                        {},
	"passphrase":                      {},
	"pin":                             {},
	"otp":                             {},
	"token":                           {},
	"access_token":                    {},
	"refresh_token":                   {},
	"exam_token":                      {},
	"csrf_token":                      {},
	"session_token":                   {},
	"cookie":                          {},
	"set_cookie":                      {},
	"authorization":                   {},
	"auth_header":                     {},
	"bearer":                          {},
	"jwt":                             {},
	"secret":                          {},
	"api_key":                         {},
	"apikey":                          {},
	"client_secret":                   {},
	"nik":                             {},
	"nip":                             {},
	"nisn":                            {},
	"nisn_full":                       {},
	"nip_full":                        {},
	"device_fingerprint":              {},
	"fingerprint":                     {},
	"device_id_hash_from_fingerprint": {},
	"pusaka_username":                 {},
	"pusaka_password":                 {},
	"pusaka_credential":               {},
	"credential_pusaka":               {},
	"raw_ip":                          {},
	"ip_address_raw":                  {},
	"remote_addr":                     {},
	"x_forwarded_for_raw":             {},
	"raw_user_agent":                  {},
	"user_agent_raw":                  {},
	"ua_raw":                          {},
	"sql":                             {},
	"stack_trace":                     {},
	"request_body":                    {},
	"response_body":                   {},
	"raw_payload":                     {},
}

func (s *InternalAnalytics) CreateEvent(ctx context.Context, in CreateInternalAnalyticsEventInput) (InternalAnalyticsEventReceipt, error) {
	eventName := strings.TrimSpace(in.EventName)
	eventGroup := strings.TrimSpace(in.EventGroup)
	sourceSurface := strings.TrimSpace(in.SourceSurface)

	expectedGroup, ok := internalAnalyticsAllowedEvents[eventName]
	if !ok {
		return InternalAnalyticsEventReceipt{}, internalAnalyticsValidation(InternalAnalyticsErrEventNotAllowlisted, eventName)
	}
	if eventGroup != expectedGroup {
		return InternalAnalyticsEventReceipt{}, internalAnalyticsValidation(InternalAnalyticsErrGroupMismatch, eventGroup)
	}
	if _, ok := internalAnalyticsAllowedSourceSurfaces[sourceSurface]; !ok {
		return InternalAnalyticsEventReceipt{}, internalAnalyticsValidation(InternalAnalyticsErrInvalidSourceSurface, sourceSurface)
	}

	metadata := in.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	if key, ok := findForbiddenInternalAnalyticsMetadataKey(metadata); ok {
		return InternalAnalyticsEventReceipt{}, internalAnalyticsValidation(InternalAnalyticsErrForbiddenMetadataKey, key)
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return InternalAnalyticsEventReceipt{}, internalAnalyticsValidation(InternalAnalyticsErrInvalidMetadata, "")
	}
	if len(metadataJSON) > internalAnalyticsMetadataMaxBytes {
		return InternalAnalyticsEventReceipt{}, internalAnalyticsValidation(InternalAnalyticsErrMetadataTooLarge, "")
	}

	now := s.clockNow()
	retentionExpiresAt := defaultInternalAnalyticsRetention(eventGroup, now)
	if in.RetentionExpiresAt != nil && !in.RetentionExpiresAt.IsZero() {
		retentionExpiresAt = in.RetentionExpiresAt.UTC()
		if !retentionExpiresAt.After(now) {
			return InternalAnalyticsEventReceipt{}, internalAnalyticsValidation(InternalAnalyticsErrInvalidRetention, "")
		}
	}

	if s == nil || s.q == nil {
		return InternalAnalyticsEventReceipt{}, fmt.Errorf("internal analytics store unavailable")
	}
	row, err := s.q.CreateInternalAnalyticsEvent(ctx, db.CreateInternalAnalyticsEventParams{
		EventName:          eventName,
		EventGroup:         eventGroup,
		SourceSurface:      sourceSurface,
		ActorUserID:        in.ActorUserID,
		ActorRole:          internalAnalyticsText(in.ActorRole),
		PermissionCode:     internalAnalyticsText(in.PermissionCode),
		RouteGroup:         internalAnalyticsText(in.RouteGroup),
		Module:             internalAnalyticsText(in.Module),
		Result:             internalAnalyticsText(in.Result),
		StatusCodeClass:    internalAnalyticsText(in.StatusCodeClass),
		DurationBucket:     internalAnalyticsText(in.DurationBucket),
		Metadata:           metadataJSON,
		RetentionExpiresAt: pgtype.Timestamptz{Time: retentionExpiresAt, Valid: true},
	})
	if err != nil {
		return InternalAnalyticsEventReceipt{}, err
	}
	return mapInternalAnalyticsReceipt(row), nil
}

func (s *InternalAnalytics) clockNow() time.Time {
	if s != nil && s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

func defaultInternalAnalyticsRetention(eventGroup string, now time.Time) time.Time {
	switch eventGroup {
	case "public":
		return now.Add(90 * 24 * time.Hour)
	case "security":
		return now.Add(180 * 24 * time.Hour)
	default:
		return now.Add(180 * 24 * time.Hour)
	}
}

func internalAnalyticsText(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	return pgtype.Text{String: value, Valid: value != ""}
}

func mapInternalAnalyticsReceipt(row db.InternalAnalyticsEvent) InternalAnalyticsEventReceipt {
	var occurredAt time.Time
	if row.OccurredAt.Valid {
		occurredAt = row.OccurredAt.Time
	}
	var retentionExpiresAt time.Time
	if row.RetentionExpiresAt.Valid {
		retentionExpiresAt = row.RetentionExpiresAt.Time
	}
	return InternalAnalyticsEventReceipt{
		ID:                 row.ID.String(),
		EventName:          row.EventName,
		EventGroup:         row.EventGroup,
		OccurredAt:         occurredAt,
		RetentionExpiresAt: retentionExpiresAt,
	}
}

func internalAnalyticsValidation(code, field string) *InternalAnalyticsValidationError {
	return &InternalAnalyticsValidationError{Code: code, Field: field}
}

func findForbiddenInternalAnalyticsMetadataKey(value any) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			normalized := normalizeInternalAnalyticsMetadataKey(key)
			if _, ok := internalAnalyticsForbiddenMetadataKeys[normalized]; ok {
				return normalized, true
			}
			if nestedKey, ok := findForbiddenInternalAnalyticsMetadataKey(child); ok {
				return nestedKey, true
			}
		}
	case []any:
		for _, child := range typed {
			if nestedKey, ok := findForbiddenInternalAnalyticsMetadataKey(child); ok {
				return nestedKey, true
			}
		}
	}
	return "", false
}

func normalizeInternalAnalyticsMetadataKey(key string) string {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(normalized)
	return strings.Trim(normalized, "_")
}
