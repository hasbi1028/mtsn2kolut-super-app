package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeInternalAnalyticsStore struct {
	calls        []db.CreateInternalAnalyticsEventParams
	dailyCalls   []db.ListInternalAnalyticsDailyAggregatesParams
	rollupCalls  []db.ListInternalAnalyticsEventsForRollupParams
	upsertCalls  []db.UpsertInternalAnalyticsDailyAggregateParams
	deleteCalls  []pgtype.Timestamptz
	backlogCalls []pgtype.Timestamptz
	row          db.InternalAnalyticsEvent
	dailyRows    []db.InternalAnalyticsDailyAggregate
	rollupRows   []db.InternalAnalyticsEvent
	backlogRow   db.GetInternalAnalyticsExpiredEventBacklogRow
	deleted      int64
	err          error
	dailyErr     error
	rollupErr    error
	upsertErr    error
	deleteErr    error
	backlogErr   error
}

func (f *fakeInternalAnalyticsStore) CreateInternalAnalyticsEvent(_ context.Context, arg db.CreateInternalAnalyticsEventParams) (db.InternalAnalyticsEvent, error) {
	f.calls = append(f.calls, arg)
	if f.err != nil {
		return db.InternalAnalyticsEvent{}, f.err
	}
	row := f.row
	if !row.ID.Valid {
		row.ID = serviceTestUUID(11)
	}
	if row.EventName == "" {
		row.EventName = arg.EventName
	}
	if row.EventGroup == "" {
		row.EventGroup = arg.EventGroup
	}
	if !row.OccurredAt.Valid {
		row.OccurredAt = pgtype.Timestamptz{Time: fixedAnalyticsNow(), Valid: true}
	}
	if !row.RetentionExpiresAt.Valid {
		row.RetentionExpiresAt = arg.RetentionExpiresAt
	}
	return row, nil
}

func (f *fakeInternalAnalyticsStore) ListInternalAnalyticsDailyAggregates(_ context.Context, arg db.ListInternalAnalyticsDailyAggregatesParams) ([]db.InternalAnalyticsDailyAggregate, error) {
	f.dailyCalls = append(f.dailyCalls, arg)
	if f.dailyErr != nil {
		return nil, f.dailyErr
	}
	return f.dailyRows, nil
}

func (f *fakeInternalAnalyticsStore) ListInternalAnalyticsEventsForRollup(_ context.Context, arg db.ListInternalAnalyticsEventsForRollupParams) ([]db.InternalAnalyticsEvent, error) {
	f.rollupCalls = append(f.rollupCalls, arg)
	if f.rollupErr != nil {
		return nil, f.rollupErr
	}
	return f.rollupRows, nil
}

func (f *fakeInternalAnalyticsStore) UpsertInternalAnalyticsDailyAggregate(_ context.Context, arg db.UpsertInternalAnalyticsDailyAggregateParams) (db.InternalAnalyticsDailyAggregate, error) {
	f.upsertCalls = append(f.upsertCalls, arg)
	if f.upsertErr != nil {
		return db.InternalAnalyticsDailyAggregate{}, f.upsertErr
	}
	return db.InternalAnalyticsDailyAggregate{
		AggregateDate: arg.AggregateDate,
		EventGroup:    arg.EventGroup,
		EventName:     arg.EventName,
		SourceSurface: arg.SourceSurface,
		Role:          arg.Role,
		Result:        arg.Result,
		Count:         arg.EventCount,
		Metadata:      arg.Metadata,
	}, nil
}

func (f *fakeInternalAnalyticsStore) DeleteExpiredInternalAnalyticsEvents(_ context.Context, cutoffAt pgtype.Timestamptz) (int64, error) {
	f.deleteCalls = append(f.deleteCalls, cutoffAt)
	if f.deleteErr != nil {
		return 0, f.deleteErr
	}
	return f.deleted, nil
}

func (f *fakeInternalAnalyticsStore) GetInternalAnalyticsExpiredEventBacklog(_ context.Context, cutoffAt pgtype.Timestamptz) (db.GetInternalAnalyticsExpiredEventBacklogRow, error) {
	f.backlogCalls = append(f.backlogCalls, cutoffAt)
	if f.backlogErr != nil {
		return db.GetInternalAnalyticsExpiredEventBacklogRow{}, f.backlogErr
	}
	return f.backlogRow, nil
}

func fixedAnalyticsNow() time.Time {
	return time.Date(2026, 5, 9, 10, 30, 0, 0, time.UTC)
}

func serviceTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func newTestInternalAnalytics(store *fakeInternalAnalyticsStore) *InternalAnalytics {
	return &InternalAnalytics{
		q:   store,
		now: fixedAnalyticsNow,
	}
}

func expectInternalAnalyticsValidationCode(t *testing.T, err error, code string) {
	t.Helper()
	var validationErr *InternalAnalyticsValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want InternalAnalyticsValidationError", err)
	}
	if validationErr.Code != code {
		t.Fatalf("validation code = %q, want %q", validationErr.Code, code)
	}
}

func TestInternalAnalyticsAcceptsAllowlistedSafeEvent(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	receipt, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "public.page_view",
		EventGroup:    "public",
		SourceSurface: "public_website",
		Metadata: map[string]any{
			"route_group":  "home",
			"device_class": "desktop",
		},
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if receipt.EventName != "public.page_view" || receipt.EventGroup != "public" {
		t.Fatalf("receipt = %+v, want safe event receipt", receipt)
	}
	if len(store.calls) != 1 {
		t.Fatalf("CreateInternalAnalyticsEvent calls = %d, want 1", len(store.calls))
	}
	call := store.calls[0]
	if call.EventName != "public.page_view" || call.EventGroup != "public" || call.SourceSurface != "public_website" {
		t.Fatalf("stored event params = %+v, want event identity and source", call)
	}
	wantRetention := fixedAnalyticsNow().Add(90 * 24 * time.Hour)
	if !call.RetentionExpiresAt.Valid || !call.RetentionExpiresAt.Time.Equal(wantRetention) {
		t.Fatalf("retention = %v, want public default %v", call.RetentionExpiresAt, wantRetention)
	}
	var metadata map[string]any
	if err := json.Unmarshal(call.Metadata, &metadata); err != nil {
		t.Fatalf("stored metadata is not json object: %v", err)
	}
	if metadata["route_group"] != "home" || metadata["device_class"] != "desktop" {
		t.Fatalf("metadata = %+v, want sanitized safe keys", metadata)
	}
}

func TestInternalAnalyticsDefaultRetentionByGroup(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		group     string
		wantDays  int
	}{
		{name: "public", eventName: "public.page_view", group: "public", wantDays: 90},
		{name: "internal", eventName: "auth.login_success", group: "auth", wantDays: 180},
		{name: "security", eventName: "security.forbidden", group: "security", wantDays: 180},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeInternalAnalyticsStore{}
			svc := newTestInternalAnalytics(store)
			_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
				EventName:     tt.eventName,
				EventGroup:    tt.group,
				SourceSurface: "core_api",
				Metadata:      map[string]any{"result": "success"},
			})
			if err != nil {
				t.Fatalf("CreateEvent() error = %v", err)
			}
			got := store.calls[0].RetentionExpiresAt.Time
			want := fixedAnalyticsNow().Add(time.Duration(tt.wantDays) * 24 * time.Hour)
			if !got.Equal(want) {
				t.Fatalf("retention = %v, want %v", got, want)
			}
		})
	}
}

func TestInternalAnalyticsRejectsUnknownEvent(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "unknown.event",
		EventGroup:    "unknown",
		SourceSurface: "web_admin",
		Metadata:      map[string]any{},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrEventNotAllowlisted)
	if len(store.calls) != 0 {
		t.Fatalf("store called for rejected event")
	}
}

func TestInternalAnalyticsRejectsGroupMismatch(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "public.page_view",
		EventGroup:    "auth",
		SourceSurface: "public_website",
		Metadata:      map[string]any{},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrGroupMismatch)
	if len(store.calls) != 0 {
		t.Fatalf("store called for rejected event")
	}
}

func TestInternalAnalyticsRejectsForbiddenMetadataKeysCaseInsensitiveAndNested(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "dashboard.view",
		EventGroup:    "dashboard",
		SourceSurface: "web_admin",
		Metadata: map[string]any{
			"safe": map[string]any{
				"Access_Token": "must-not-be-returned",
			},
		},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrForbiddenMetadataKey)
	if strings.Contains(err.Error(), "must-not-be-returned") {
		t.Fatalf("validation error leaked raw sensitive value: %q", err.Error())
	}
	if len(store.calls) != 0 {
		t.Fatalf("store called for rejected metadata")
	}
}

func TestInternalAnalyticsRejectsMetadataTooLarge(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "dashboard.view",
		EventGroup:    "dashboard",
		SourceSurface: "web_admin",
		Metadata: map[string]any{
			"widget_key": strings.Repeat("x", internalAnalyticsMetadataMaxBytes+1),
		},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrMetadataTooLarge)
	if len(store.calls) != 0 {
		t.Fatalf("store called for oversized metadata")
	}
}

func TestInternalAnalyticsRejectsInvalidSourceSurface(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)

	_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
		EventName:     "dashboard.view",
		EventGroup:    "dashboard",
		SourceSurface: "browser",
		Metadata:      map[string]any{},
	})
	expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrInvalidSourceSurface)
	if len(store.calls) != 0 {
		t.Fatalf("store called for invalid source surface")
	}
}

func TestInternalAnalyticsListsDailyAggregatesWithoutRawMetadata(t *testing.T) {
	store := &fakeInternalAnalyticsStore{
		dailyRows: []db.InternalAnalyticsDailyAggregate{
			{
				AggregateDate: pgtype.Date{Time: time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC), Valid: true},
				EventGroup:    "dashboard",
				EventName:     "dashboard.view",
				SourceSurface: "web_admin",
				Role:          "admin",
				Result:        "success",
				Count:         7,
				Metadata:      []byte(`{"raw_user_agent":"must-not-leak","safe":"ignored"}`),
			},
		},
	}
	svc := newTestInternalAnalytics(store)

	result, err := svc.ListDailyAggregates(context.Background(), InternalAnalyticsDailyQuery{
		EventGroup:    " dashboard ",
		Days:          14,
		SourceSurface: " web_admin ",
		Limit:         25,
	})
	if err != nil {
		t.Fatalf("ListDailyAggregates() error = %v", err)
	}
	if len(store.dailyCalls) != 1 {
		t.Fatalf("ListInternalAnalyticsDailyAggregates calls = %d, want 1", len(store.dailyCalls))
	}
	call := store.dailyCalls[0]
	if call.EventGroup != "dashboard" || call.SourceSurface != "web_admin" || call.LimitCount != 25 {
		t.Fatalf("daily aggregate params = %+v, want trimmed filters and explicit limit", call)
	}
	if !call.EndDate.Time.Equal(time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("end date = %v, want fixed now date", call.EndDate.Time)
	}
	if !call.StartDate.Time.Equal(time.Date(2026, 4, 26, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("start date = %v, want 14 day inclusive window", call.StartDate.Time)
	}
	if len(result.Items) != 1 {
		t.Fatalf("daily result items = %d, want 1", len(result.Items))
	}
	item := result.Items[0]
	if item.AggregateDate != "2026-05-09" || item.Count != 7 || item.Role != "admin" || item.Result != "success" {
		t.Fatalf("daily item = %+v, want mapped aggregate fields", item)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal daily result: %v", err)
	}
	if strings.Contains(string(encoded), "raw_user_agent") || strings.Contains(string(encoded), "must-not-leak") || strings.Contains(string(encoded), "metadata") {
		t.Fatalf("daily aggregate response leaked metadata: %s", encoded)
	}
}

func TestInternalAnalyticsSummaryAggregatesCountsByGroupAndEvent(t *testing.T) {
	store := &fakeInternalAnalyticsStore{
		dailyRows: []db.InternalAnalyticsDailyAggregate{
			{AggregateDate: pgtype.Date{Time: time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC), Valid: true}, EventGroup: "dashboard", EventName: "dashboard.view", SourceSurface: "web_admin", Count: 3},
			{AggregateDate: pgtype.Date{Time: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC), Valid: true}, EventGroup: "dashboard", EventName: "dashboard.view", SourceSurface: "web_admin", Count: 4},
			{AggregateDate: pgtype.Date{Time: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC), Valid: true}, EventGroup: "security", EventName: "security.forbidden", SourceSurface: "core_api", Count: 2},
		},
	}
	svc := newTestInternalAnalytics(store)

	summary, err := svc.Summary(context.Background(), InternalAnalyticsSummaryQuery{Days: 7})
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.TotalCount != 9 {
		t.Fatalf("summary total = %d, want 9", summary.TotalCount)
	}
	if len(summary.Groups) != 2 || summary.Groups[0].EventGroup != "dashboard" || summary.Groups[0].Count != 7 {
		t.Fatalf("summary groups = %+v, want grouped counts sorted by count", summary.Groups)
	}
	if len(summary.TopEvents) != 2 || summary.TopEvents[0].EventName != "dashboard.view" || summary.TopEvents[0].Count != 7 {
		t.Fatalf("summary top events = %+v, want event counts sorted by count", summary.TopEvents)
	}
	if len(store.dailyCalls) != 1 || store.dailyCalls[0].LimitCount != 10000 {
		t.Fatalf("summary daily call = %+v, want bounded aggregate read", store.dailyCalls)
	}
}

func TestInternalAnalyticsExportAggregatesCSVUsesOnlyAggregatesAndAudits(t *testing.T) {
	store := &fakeInternalAnalyticsStore{
		dailyRows: []db.InternalAnalyticsDailyAggregate{
			{
				AggregateDate: pgtype.Date{Time: time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC), Valid: true},
				EventGroup:    "dashboard",
				EventName:     "=dashboard.view",
				SourceSurface: "+web_admin",
				Role:          "-admin",
				Result:        "@success",
				Count:         12,
				Metadata:      []byte(`{"raw_user_agent":"must-not-leak","actor_user_id":"must-not-leak"}`),
			},
		},
	}
	svc := newTestInternalAnalytics(store)

	exported, err := svc.ExportAggregatesCSV(context.Background(), InternalAnalyticsExportQuery{
		Days:        7,
		EventGroup:  "dashboard",
		ActorUserID: serviceTestUUID(19),
		ActorRole:   "admin",
	})
	if err != nil {
		t.Fatalf("ExportAggregatesCSV() error = %v", err)
	}
	if exported.Filename == "" || !strings.Contains(exported.Filename, "internal-analytics-aggregate") {
		t.Fatalf("filename = %q, want aggregate analytics CSV filename", exported.Filename)
	}
	csvText := string(exported.Content)
	for _, forbidden := range []string{"metadata", "actor_user_id", "raw_user_agent", "must-not-leak", "session_id", "ip_address", "user_agent"} {
		if strings.Contains(csvText, forbidden) {
			t.Fatalf("export leaked forbidden field %q in CSV:\n%s", forbidden, csvText)
		}
	}
	records, err := csv.NewReader(bytes.NewReader(exported.Content)).ReadAll()
	if err != nil {
		t.Fatalf("export CSV is not parseable: %v\n%s", err, csvText)
	}
	wantHeader := []string{"aggregate_date", "event_group", "event_name", "source_surface", "role", "result", "count"}
	if got := strings.Join(records[0], ","); got != strings.Join(wantHeader, ",") {
		t.Fatalf("CSV header = %v, want %v", records[0], wantHeader)
	}
	for _, cell := range records[1][2:6] {
		if !strings.HasPrefix(cell, "'") {
			t.Fatalf("CSV cell %q was not formula-injection escaped", cell)
		}
	}
	if len(store.dailyCalls) != 1 || store.dailyCalls[0].EventGroup != "dashboard" || store.dailyCalls[0].LimitCount != 10000 {
		t.Fatalf("daily aggregate export call = %+v, want bounded aggregate-only read", store.dailyCalls)
	}
	if len(store.calls) != 1 {
		t.Fatalf("analytics audit event calls = %d, want 1", len(store.calls))
	}
	audit := store.calls[0]
	if audit.EventName != "security.export_requested" || audit.EventGroup != "security" || audit.SourceSurface != "core_api" {
		t.Fatalf("audit event = %+v, want security.export_requested from core_api", audit)
	}
	var auditMetadata map[string]any
	if err := json.Unmarshal(audit.Metadata, &auditMetadata); err != nil {
		t.Fatalf("audit metadata json = %v", err)
	}
	if auditMetadata["export_type"] != "aggregate_csv" || auditMetadata["event_group"] != "dashboard" || auditMetadata["days"].(float64) != 7 {
		t.Fatalf("audit metadata = %+v, want safe aggregate export metadata", auditMetadata)
	}
	if strings.Contains(string(audit.Metadata), "actor_user_id") || strings.Contains(string(audit.Metadata), "metadata") {
		t.Fatalf("audit metadata leaked raw identifiers: %s", string(audit.Metadata))
	}
}

func TestInternalAnalyticsRollupAndCleanupAggregatesEventsAndDeletesExpired(t *testing.T) {
	startAt := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)
	endAt := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	cutoffAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	store := &fakeInternalAnalyticsStore{
		deleted: 4,
		rollupRows: []db.InternalAnalyticsEvent{
			{
				EventName:     "dashboard.view",
				EventGroup:    "dashboard",
				OccurredAt:    pgtype.Timestamptz{Time: time.Date(2026, 5, 8, 3, 0, 0, 0, time.UTC), Valid: true},
				SourceSurface: "web_admin",
				ActorRole:     pgtype.Text{String: "admin", Valid: true},
				Result:        pgtype.Text{String: "success", Valid: true},
				Metadata:      []byte(`{"raw_user_agent":"ignored"}`),
			},
			{
				EventName:     "dashboard.view",
				EventGroup:    "dashboard",
				OccurredAt:    pgtype.Timestamptz{Time: time.Date(2026, 5, 8, 4, 0, 0, 0, time.UTC), Valid: true},
				SourceSurface: "web_admin",
				ActorRole:     pgtype.Text{String: "admin", Valid: true},
				Result:        pgtype.Text{String: "success", Valid: true},
			},
			{
				EventName:     "security.forbidden",
				EventGroup:    "security",
				OccurredAt:    pgtype.Timestamptz{Time: time.Date(2026, 5, 9, 1, 0, 0, 0, time.UTC), Valid: true},
				SourceSurface: "core_api",
				Result:        pgtype.Text{String: "blocked", Valid: true},
			},
		},
	}
	svc := newTestInternalAnalytics(store)

	result, err := svc.RollupAndCleanup(context.Background(), InternalAnalyticsRollupCleanupInput{
		StartAt:  startAt,
		EndAt:    endAt,
		CutoffAt: cutoffAt,
		Limit:    500,
	})
	if err != nil {
		t.Fatalf("RollupAndCleanup() error = %v", err)
	}
	if result.EventsScanned != 3 || result.AggregatesUpserted != 2 || result.ExpiredEventsDeleted != 4 {
		t.Fatalf("rollup result = %+v, want scanned/upserted/deleted counts", result)
	}
	if len(store.rollupCalls) != 1 || !store.rollupCalls[0].StartAt.Time.Equal(startAt) || !store.rollupCalls[0].EndAt.Time.Equal(endAt) || store.rollupCalls[0].LimitCount != 500 {
		t.Fatalf("rollup query = %+v, want bounded requested window", store.rollupCalls)
	}
	if len(store.upsertCalls) != 2 {
		t.Fatalf("upsert calls = %d, want 2 aggregate keys", len(store.upsertCalls))
	}
	first := store.upsertCalls[0]
	if first.EventName != "dashboard.view" || first.EventCount != 2 || first.Role != "admin" || first.Result != "success" {
		t.Fatalf("first aggregate upsert = %+v, want grouped dashboard count", first)
	}
	if strings.Contains(string(first.Metadata), "raw_user_agent") {
		t.Fatalf("aggregate metadata must stay empty/safe, got %s", string(first.Metadata))
	}
	if len(store.deleteCalls) != 1 || !store.deleteCalls[0].Time.Equal(cutoffAt) {
		t.Fatalf("delete calls = %+v, want explicit cutoff", store.deleteCalls)
	}
}

func TestInternalAnalyticsExpiredBacklogSummaryIsAggregateOnly(t *testing.T) {
	oldest := time.Date(2026, 5, 1, 2, 0, 0, 0, time.UTC)
	store := &fakeInternalAnalyticsStore{
		backlogRow: db.GetInternalAnalyticsExpiredEventBacklogRow{
			ExpiredCount:    9,
			OldestExpiredAt: pgtype.Timestamptz{Time: oldest, Valid: true},
		},
	}
	svc := newTestInternalAnalytics(store)

	backlog, err := svc.ExpiredBacklog(context.Background())
	if err != nil {
		t.Fatalf("ExpiredBacklog() error = %v", err)
	}
	if backlog.ExpiredEventBacklogCount != 9 || backlog.OldestExpiredEventAt == nil || !backlog.OldestExpiredEventAt.Equal(oldest) {
		t.Fatalf("backlog = %+v, want aggregate expired count and oldest date", backlog)
	}
	if len(store.backlogCalls) != 1 || !store.backlogCalls[0].Time.Equal(fixedAnalyticsNow()) {
		t.Fatalf("backlog calls = %+v, want current clock cutoff", store.backlogCalls)
	}
}

func TestInternalAnalyticsRejectsExpandedSensitiveMetadataKeys(t *testing.T) {
	tests := []string{
		"token",
		"Authorization",
		"bearer",
		"cookie",
		"NIP",
		"nisn",
		"nik",
		"deviceFingerprint",
		"device-fingerprint",
		"rawUserAgent",
		"raw_user_agent",
		"query_string",
		"rawQuery",
		"full_url",
	}
	for _, key := range tests {
		t.Run(key, func(t *testing.T) {
			store := &fakeInternalAnalyticsStore{}
			svc := newTestInternalAnalytics(store)
			_, err := svc.CreateEvent(context.Background(), CreateInternalAnalyticsEventInput{
				EventName:     "security.validation_rejected",
				EventGroup:    "security",
				SourceSurface: "core_api",
				Metadata:      map[string]any{key: "must-not-store"},
			})
			expectInternalAnalyticsValidationCode(t, err, InternalAnalyticsErrForbiddenMetadataKey)
			if len(store.calls) != 0 {
				t.Fatalf("store called for forbidden metadata key %q", key)
			}
		})
	}
}
