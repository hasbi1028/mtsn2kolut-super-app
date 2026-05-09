package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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
	ListInternalAnalyticsDailyAggregates(ctx context.Context, arg db.ListInternalAnalyticsDailyAggregatesParams) ([]db.InternalAnalyticsDailyAggregate, error)
	ListInternalAnalyticsEventsForRollup(ctx context.Context, arg db.ListInternalAnalyticsEventsForRollupParams) ([]db.InternalAnalyticsEvent, error)
	UpsertInternalAnalyticsDailyAggregate(ctx context.Context, arg db.UpsertInternalAnalyticsDailyAggregateParams) (db.InternalAnalyticsDailyAggregate, error)
	DeleteExpiredInternalAnalyticsEvents(ctx context.Context, cutoffAt pgtype.Timestamptz) (int64, error)
	GetInternalAnalyticsExpiredEventBacklog(ctx context.Context, cutoffAt pgtype.Timestamptz) (db.GetInternalAnalyticsExpiredEventBacklogRow, error)
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

type CreatePublicAnalyticsEventInput struct {
	EventName  string
	RouteGroup string
	Result     string
	Metadata   map[string]any
}

type InternalAnalyticsEventReceipt struct {
	ID                 string    `json:"id"`
	EventName          string    `json:"event_name"`
	EventGroup         string    `json:"event_group"`
	OccurredAt         time.Time `json:"occurred_at"`
	RetentionExpiresAt time.Time `json:"retention_expires_at"`
}

type InternalAnalyticsDailyQuery struct {
	EventGroup    string
	EventName     string
	SourceSurface string
	Role          string
	Result        string
	Days          int32
	Limit         int32
	Offset        int32
}

type InternalAnalyticsSummaryQuery struct {
	Days          int32
	EventGroup    string
	SourceSurface string
	Role          string
	Result        string
}

type InternalAnalyticsExportQuery struct {
	EventGroup    string
	EventName     string
	SourceSurface string
	Role          string
	Result        string
	Days          int32
	Limit         int32
	Offset        int32
	ActorUserID   pgtype.UUID
	ActorRole     string
}

type InternalAnalyticsExport struct {
	Filename string
	Content  []byte
}

type InternalAnalyticsRollupCleanupInput struct {
	StartAt  time.Time
	EndAt    time.Time
	CutoffAt time.Time
	Limit    int32
}

type InternalAnalyticsRollupCleanupResult struct {
	StartAt              time.Time `json:"start_at"`
	EndAt                time.Time `json:"end_at"`
	CutoffAt             time.Time `json:"cutoff_at"`
	EventsScanned        int64     `json:"events_scanned"`
	AggregatesUpserted   int64     `json:"aggregates_upserted"`
	ExpiredEventsDeleted int64     `json:"expired_events_deleted"`
	LimitReached         bool      `json:"limit_reached"`
}

type InternalAnalyticsExpiredBacklog struct {
	ExpiredEventBacklogCount int64      `json:"expired_event_backlog_count"`
	OldestExpiredEventAt     *time.Time `json:"oldest_expired_event_at,omitempty"`
}

type InternalAnalyticsDailyResult struct {
	Days          int32                        `json:"days"`
	EventGroup    string                       `json:"event_group,omitempty"`
	EventName     string                       `json:"event_name,omitempty"`
	SourceSurface string                       `json:"source_surface,omitempty"`
	Items         []InternalAnalyticsDailyItem `json:"items"`
}

type InternalAnalyticsDailyItem struct {
	AggregateDate string `json:"aggregate_date"`
	EventGroup    string `json:"event_group"`
	EventName     string `json:"event_name"`
	SourceSurface string `json:"source_surface"`
	Role          string `json:"role,omitempty"`
	Result        string `json:"result,omitempty"`
	Count         int64  `json:"count"`
}

type InternalAnalyticsSummary struct {
	Days       int32                           `json:"days"`
	EventGroup string                          `json:"event_group,omitempty"`
	TotalCount int64                           `json:"total_count"`
	Groups     []InternalAnalyticsGroupSummary `json:"groups"`
	TopEvents  []InternalAnalyticsEventSummary `json:"top_events"`
}

type InternalAnalyticsGroupSummary struct {
	EventGroup string `json:"event_group"`
	Count      int64  `json:"count"`
}

type InternalAnalyticsEventSummary struct {
	EventName  string `json:"event_name"`
	EventGroup string `json:"event_group"`
	Count      int64  `json:"count"`
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
	"bank_soal.export":                "bank_soal",
	"bank_soal.review_view":           "bank_soal",
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
	"pusaka.manual_run":               "pusaka",
	"pusaka.scheduler_tick":           "pusaka",
	"pusaka.queue_cancel":             "pusaka",
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
	"security.settings_view":          "security",
	"security.settings_update":        "security",
	"security.analytics_view":         "security",
	"security.analytics_filter":       "security",
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

var internalAnalyticsPublicMetadataKeys = map[string]struct{}{
	"page_key":             {},
	"page_kind":            {},
	"cta_key":              {},
	"cta_group":            {},
	"link_kind":            {},
	"file_kind":            {},
	"download_kind":        {},
	"search_bucket":        {},
	"search_length_bucket": {},
	"form_key":             {},
	"form_step":            {},
	"result":               {},
	"device_class":         {},
	"source_component":     {},
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
	"device_fingerprint_hash":         {},
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
	"user_agent":                      {},
	"ua_raw":                          {},
	"query":                           {},
	"query_string":                    {},
	"raw_query":                       {},
	"url":                             {},
	"full_url":                        {},
	"sql":                             {},
	"stack_trace":                     {},
	"request_body":                    {},
	"response_body":                   {},
	"raw_payload":                     {},
}

func (s *InternalAnalytics) CreatePublicEvent(ctx context.Context, in CreatePublicAnalyticsEventInput) (InternalAnalyticsEventReceipt, error) {
	eventName := strings.TrimSpace(in.EventName)
	if internalAnalyticsAllowedEvents[eventName] != "public" {
		return InternalAnalyticsEventReceipt{}, internalAnalyticsValidation(InternalAnalyticsErrEventNotAllowlisted, eventName)
	}
	metadata := sanitizeInternalAnalyticsPublicMetadata(in.Metadata)
	return s.CreateEvent(ctx, CreateInternalAnalyticsEventInput{
		EventName:     eventName,
		EventGroup:    "public",
		SourceSurface: "public_website",
		RouteGroup:    sanitizeInternalAnalyticsPublicToken(in.RouteGroup, "home"),
		Module:        "public_site",
		Result:        sanitizeInternalAnalyticsPublicResult(in.Result, metadata),
		Metadata:      metadata,
	})
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

func (s *InternalAnalytics) ListDailyAggregates(ctx context.Context, in InternalAnalyticsDailyQuery) (InternalAnalyticsDailyResult, error) {
	if s == nil || s.q == nil {
		return InternalAnalyticsDailyResult{}, fmt.Errorf("internal analytics store unavailable")
	}
	query := normalizeInternalAnalyticsDailyQuery(in)
	endDate := internalAnalyticsDateOnly(s.clockNow())
	startDate := endDate.AddDate(0, 0, -int(query.Days-1))
	rows, err := s.q.ListInternalAnalyticsDailyAggregates(ctx, db.ListInternalAnalyticsDailyAggregatesParams{
		StartDate:     pgtype.Date{Time: startDate, Valid: true},
		EndDate:       pgtype.Date{Time: endDate, Valid: true},
		EventGroup:    query.EventGroup,
		EventName:     query.EventName,
		SourceSurface: query.SourceSurface,
		Role:          query.Role,
		Result:        query.Result,
		OffsetCount:   query.Offset,
		LimitCount:    query.Limit,
	})
	if err != nil {
		if isInternalAnalyticsSchemaUnavailable(err) {
			return emptyInternalAnalyticsDailyResult(query), nil
		}
		return InternalAnalyticsDailyResult{}, err
	}
	items := make([]InternalAnalyticsDailyItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapInternalAnalyticsDailyItem(row))
	}
	return InternalAnalyticsDailyResult{
		Days:          query.Days,
		EventGroup:    query.EventGroup,
		EventName:     query.EventName,
		SourceSurface: query.SourceSurface,
		Items:         items,
	}, nil
}

func (s *InternalAnalytics) Summary(ctx context.Context, in InternalAnalyticsSummaryQuery) (InternalAnalyticsSummary, error) {
	query := InternalAnalyticsDailyQuery{
		EventGroup:    strings.TrimSpace(in.EventGroup),
		SourceSurface: strings.TrimSpace(in.SourceSurface),
		Role:          strings.TrimSpace(in.Role),
		Result:        strings.TrimSpace(in.Result),
		Days:          in.Days,
		Limit:         10000,
	}
	daily, err := s.ListDailyAggregates(ctx, query)
	if err != nil {
		return InternalAnalyticsSummary{}, err
	}
	groupCounts := map[string]int64{}
	eventCounts := map[string]InternalAnalyticsEventSummary{}
	var total int64
	for _, item := range daily.Items {
		total += item.Count
		groupCounts[item.EventGroup] += item.Count
		current := eventCounts[item.EventName]
		current.EventName = item.EventName
		current.EventGroup = item.EventGroup
		current.Count += item.Count
		eventCounts[item.EventName] = current
	}
	groups := make([]InternalAnalyticsGroupSummary, 0, len(groupCounts))
	for group, count := range groupCounts {
		groups = append(groups, InternalAnalyticsGroupSummary{EventGroup: group, Count: count})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Count == groups[j].Count {
			return groups[i].EventGroup < groups[j].EventGroup
		}
		return groups[i].Count > groups[j].Count
	})
	topEvents := make([]InternalAnalyticsEventSummary, 0, len(eventCounts))
	for _, item := range eventCounts {
		topEvents = append(topEvents, item)
	}
	sort.Slice(topEvents, func(i, j int) bool {
		if topEvents[i].Count == topEvents[j].Count {
			return topEvents[i].EventName < topEvents[j].EventName
		}
		return topEvents[i].Count > topEvents[j].Count
	})
	if len(topEvents) > 10 {
		topEvents = topEvents[:10]
	}
	return InternalAnalyticsSummary{
		Days:       daily.Days,
		EventGroup: daily.EventGroup,
		TotalCount: total,
		Groups:     groups,
		TopEvents:  topEvents,
	}, nil
}

func (s *InternalAnalytics) ExportAggregatesCSV(ctx context.Context, in InternalAnalyticsExportQuery) (InternalAnalyticsExport, error) {
	if err := s.recordAggregateExportEvent(ctx, in); err != nil {
		if !isInternalAnalyticsSchemaUnavailable(err) {
			return InternalAnalyticsExport{}, err
		}
	}
	query := InternalAnalyticsDailyQuery{
		EventGroup:    in.EventGroup,
		EventName:     in.EventName,
		SourceSurface: in.SourceSurface,
		Role:          in.Role,
		Result:        in.Result,
		Days:          in.Days,
		Limit:         in.Limit,
		Offset:        in.Offset,
	}
	if query.Limit <= 0 {
		query.Limit = 10000
	}
	daily, err := s.ListDailyAggregates(ctx, query)
	if err != nil {
		return InternalAnalyticsExport{}, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	header := []string{"aggregate_date", "event_group", "event_name", "source_surface", "role", "result", "count"}
	if err := writer.Write(header); err != nil {
		return InternalAnalyticsExport{}, err
	}
	for _, item := range daily.Items {
		record := []string{
			safeInternalAnalyticsCSVCell(item.AggregateDate),
			safeInternalAnalyticsCSVCell(item.EventGroup),
			safeInternalAnalyticsCSVCell(item.EventName),
			safeInternalAnalyticsCSVCell(item.SourceSurface),
			safeInternalAnalyticsCSVCell(item.Role),
			safeInternalAnalyticsCSVCell(item.Result),
			fmt.Sprintf("%d", item.Count),
		}
		if err := writer.Write(record); err != nil {
			return InternalAnalyticsExport{}, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return InternalAnalyticsExport{}, err
	}

	return InternalAnalyticsExport{
		Filename: fmt.Sprintf("internal-analytics-aggregate-%s.csv", s.clockNow().Format("20060102")),
		Content:  buf.Bytes(),
	}, nil
}

func emptyInternalAnalyticsDailyResult(query InternalAnalyticsDailyQuery) InternalAnalyticsDailyResult {
	return InternalAnalyticsDailyResult{
		Days:          query.Days,
		EventGroup:    query.EventGroup,
		EventName:     query.EventName,
		SourceSurface: query.SourceSurface,
		Items:         []InternalAnalyticsDailyItem{},
	}
}

func (s *InternalAnalytics) RollupAndCleanup(ctx context.Context, in InternalAnalyticsRollupCleanupInput) (InternalAnalyticsRollupCleanupResult, error) {
	if s == nil || s.q == nil {
		return InternalAnalyticsRollupCleanupResult{}, fmt.Errorf("internal analytics store unavailable")
	}
	normalized := normalizeInternalAnalyticsRollupCleanupInput(in, s.clockNow())
	rows, err := s.q.ListInternalAnalyticsEventsForRollup(ctx, db.ListInternalAnalyticsEventsForRollupParams{
		StartAt:    pgtype.Timestamptz{Time: normalized.StartAt, Valid: true},
		EndAt:      pgtype.Timestamptz{Time: normalized.EndAt, Valid: true},
		LimitCount: normalized.Limit,
	})
	if err != nil {
		return InternalAnalyticsRollupCleanupResult{}, err
	}

	counts := map[internalAnalyticsAggregateKey]int64{}
	for _, row := range rows {
		if !row.OccurredAt.Valid {
			continue
		}
		key := internalAnalyticsAggregateKey{
			date:          internalAnalyticsDateOnly(row.OccurredAt.Time),
			eventGroup:    row.EventGroup,
			eventName:     row.EventName,
			sourceSurface: row.SourceSurface,
			role:          pgTextString(row.ActorRole),
			result:        pgTextString(row.Result),
		}
		counts[key]++
	}

	keys := make([]internalAnalyticsAggregateKey, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].sortKey() < keys[j].sortKey()
	})
	for _, key := range keys {
		if _, err := s.q.UpsertInternalAnalyticsDailyAggregate(ctx, db.UpsertInternalAnalyticsDailyAggregateParams{
			AggregateDate:   pgtype.Date{Time: key.date, Valid: true},
			EventGroup:      key.eventGroup,
			EventName:       key.eventName,
			SourceSurface:   key.sourceSurface,
			Role:            key.role,
			Result:          key.result,
			EventCount:      counts[key],
			TotalDurationMs: pgtype.Int8{},
			Metadata:        []byte(`{}`),
		}); err != nil {
			return InternalAnalyticsRollupCleanupResult{}, err
		}
	}

	deleted, err := s.q.DeleteExpiredInternalAnalyticsEvents(ctx, pgtype.Timestamptz{Time: normalized.CutoffAt, Valid: true})
	if err != nil {
		return InternalAnalyticsRollupCleanupResult{}, err
	}
	return InternalAnalyticsRollupCleanupResult{
		StartAt:              normalized.StartAt,
		EndAt:                normalized.EndAt,
		CutoffAt:             normalized.CutoffAt,
		EventsScanned:        int64(len(rows)),
		AggregatesUpserted:   int64(len(keys)),
		ExpiredEventsDeleted: deleted,
		LimitReached:         len(rows) == int(normalized.Limit),
	}, nil
}

func (s *InternalAnalytics) ExpiredBacklog(ctx context.Context) (InternalAnalyticsExpiredBacklog, error) {
	if s == nil || s.q == nil {
		return InternalAnalyticsExpiredBacklog{}, fmt.Errorf("internal analytics store unavailable")
	}
	cutoff := s.clockNow()
	row, err := s.q.GetInternalAnalyticsExpiredEventBacklog(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	if err != nil {
		return InternalAnalyticsExpiredBacklog{}, err
	}
	return InternalAnalyticsExpiredBacklog{
		ExpiredEventBacklogCount: row.ExpiredCount,
		OldestExpiredEventAt:     internalAnalyticsNullableTime(row.OldestExpiredAt),
	}, nil
}

func (s *InternalAnalytics) recordAggregateExportEvent(ctx context.Context, in InternalAnalyticsExportQuery) error {
	_, err := s.CreateEvent(ctx, CreateInternalAnalyticsEventInput{
		EventName:     "security.export_requested",
		EventGroup:    "security",
		SourceSurface: "core_api",
		ActorUserID:   in.ActorUserID,
		ActorRole:     in.ActorRole,
		Module:        "internal_analytics",
		Result:        "success",
		Metadata: map[string]any{
			"export_type":    "aggregate_csv",
			"event_group":    strings.TrimSpace(in.EventGroup),
			"source_surface": strings.TrimSpace(in.SourceSurface),
			"days":           normalizeInternalAnalyticsDailyQuery(InternalAnalyticsDailyQuery{Days: in.Days}).Days,
		},
	})
	return err
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

func normalizeInternalAnalyticsDailyQuery(in InternalAnalyticsDailyQuery) InternalAnalyticsDailyQuery {
	days := in.Days
	if days <= 0 {
		days = 30
	}
	if days > 180 {
		days = 180
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 500
	}
	if limit > 10000 {
		limit = 10000
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}
	return InternalAnalyticsDailyQuery{
		EventGroup:    strings.TrimSpace(in.EventGroup),
		EventName:     strings.TrimSpace(in.EventName),
		SourceSurface: strings.TrimSpace(in.SourceSurface),
		Role:          strings.TrimSpace(in.Role),
		Result:        strings.TrimSpace(in.Result),
		Days:          days,
		Limit:         limit,
		Offset:        offset,
	}
}

func normalizeInternalAnalyticsRollupCleanupInput(in InternalAnalyticsRollupCleanupInput, now time.Time) InternalAnalyticsRollupCleanupInput {
	now = now.UTC()
	endAt := in.EndAt.UTC()
	if endAt.IsZero() {
		endAt = internalAnalyticsDateOnly(now)
	}
	startAt := in.StartAt.UTC()
	if startAt.IsZero() {
		startAt = endAt.AddDate(0, 0, -1)
	}
	if !endAt.After(startAt) {
		endAt = startAt.Add(24 * time.Hour)
	}
	cutoffAt := in.CutoffAt.UTC()
	if cutoffAt.IsZero() {
		cutoffAt = now
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 10000
	}
	if limit > 10000 {
		limit = 10000
	}
	return InternalAnalyticsRollupCleanupInput{StartAt: startAt, EndAt: endAt, CutoffAt: cutoffAt, Limit: limit}
}

func internalAnalyticsDateOnly(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func internalAnalyticsText(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	return pgtype.Text{String: value, Valid: value != ""}
}

type internalAnalyticsAggregateKey struct {
	date          time.Time
	eventGroup    string
	eventName     string
	sourceSurface string
	role          string
	result        string
}

func (k internalAnalyticsAggregateKey) sortKey() string {
	return k.date.Format("2006-01-02") + "\x00" + k.eventGroup + "\x00" + k.eventName + "\x00" + k.sourceSurface + "\x00" + k.role + "\x00" + k.result
}

func pgTextString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
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

func mapInternalAnalyticsDailyItem(row db.InternalAnalyticsDailyAggregate) InternalAnalyticsDailyItem {
	return InternalAnalyticsDailyItem{
		AggregateDate: formatInternalAnalyticsDate(row.AggregateDate),
		EventGroup:    row.EventGroup,
		EventName:     row.EventName,
		SourceSurface: row.SourceSurface,
		Role:          strings.TrimSpace(row.Role),
		Result:        strings.TrimSpace(row.Result),
		Count:         row.Count,
	}
}

func formatInternalAnalyticsDate(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return internalAnalyticsDateOnly(value.Time).Format("2006-01-02")
}

func internalAnalyticsValidation(code, field string) *InternalAnalyticsValidationError {
	return &InternalAnalyticsValidationError{Code: code, Field: field}
}

func isInternalAnalyticsSchemaUnavailable(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	switch pgErr.Code {
	case "42P01", "42703":
		return true
	default:
		return false
	}
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
	normalized := camelToSnakeInternalAnalyticsKey(strings.TrimSpace(key))
	normalized = strings.ToLower(normalized)
	normalized = strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(normalized)
	return strings.Trim(normalized, "_")
}

func camelToSnakeInternalAnalyticsKey(key string) string {
	var b strings.Builder
	var previous rune
	for i, r := range key {
		if i > 0 && r >= 'A' && r <= 'Z' && ((previous >= 'a' && previous <= 'z') || (previous >= '0' && previous <= '9')) {
			b.WriteRune('_')
		}
		b.WriteRune(r)
		previous = r
	}
	return b.String()
}

func safeInternalAnalyticsCSVCell(value string) string {
	if value == "" {
		return ""
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r', '\n':
		return "'" + value
	default:
		return value
	}
}

func internalAnalyticsNullableTime(value any) *time.Time {
	switch typed := value.(type) {
	case nil:
		return nil
	case time.Time:
		if typed.IsZero() {
			return nil
		}
		t := typed.UTC()
		return &t
	case pgtype.Timestamptz:
		if !typed.Valid || typed.Time.IsZero() {
			return nil
		}
		t := typed.Time.UTC()
		return &t
	case *time.Time:
		if typed == nil || typed.IsZero() {
			return nil
		}
		t := typed.UTC()
		return &t
	default:
		return nil
	}
}

func sanitizeInternalAnalyticsPublicMetadata(input map[string]any) map[string]any {
	output := map[string]any{}
	for key, value := range input {
		normalizedKey := normalizeInternalAnalyticsMetadataKey(key)
		if _, ok := internalAnalyticsPublicMetadataKeys[normalizedKey]; !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			safe := sanitizeInternalAnalyticsPublicToken(typed, "")
			if safe != "" {
				output[normalizedKey] = safe
			}
		case bool:
			output[normalizedKey] = typed
		case int:
			output[normalizedKey] = typed
		case int32:
			output[normalizedKey] = typed
		case int64:
			output[normalizedKey] = typed
		case float64:
			if typed >= 0 && typed <= 10000 && typed == float64(int64(typed)) {
				output[normalizedKey] = int64(typed)
			}
		}
	}
	return output
}

func sanitizeInternalAnalyticsPublicResult(result string, metadata map[string]any) string {
	result = sanitizeInternalAnalyticsPublicToken(result, "")
	if result == "" {
		if raw, ok := metadata["result"].(string); ok {
			result = sanitizeInternalAnalyticsPublicToken(raw, "")
		}
	}
	switch result {
	case "success", "failed", "error", "validation_failed", "started":
		return result
	default:
		return ""
	}
}

func sanitizeInternalAnalyticsPublicToken(value string, fallback string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" || internalAnalyticsLooksSensitivePublicValue(normalized) {
		return fallback
	}
	normalized = strings.NewReplacer("-", "_", " ", "_", "/", "_").Replace(normalized)
	var b strings.Builder
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
		if b.Len() >= 64 {
			break
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return fallback
	}
	return out
}

func internalAnalyticsLooksSensitivePublicValue(value string) bool {
	if strings.Contains(value, "://") || strings.Contains(value, "?") || strings.Contains(value, "=") || strings.Contains(value, "@") {
		return true
	}
	consecutiveDigits := 0
	for _, r := range value {
		if r >= '0' && r <= '9' {
			consecutiveDigits++
			if consecutiveDigits >= 8 {
				return true
			}
			continue
		}
		consecutiveDigits = 0
	}
	return false
}
