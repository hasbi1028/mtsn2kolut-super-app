package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestInternalAnalyticsRollupLoopConfigNormalizesEdges(t *testing.T) {
	cfg := normalizeInternalAnalyticsRollupLoopConfig(InternalAnalyticsRollupLoopConfig{
		Interval: -time.Second,
		Lookback: -time.Hour,
		Limit:    -1,
	})
	if cfg.Interval != defaultInternalAnalyticsRollupInterval || cfg.Lookback != 48*time.Hour || cfg.Limit != 10000 {
		t.Fatalf("normalized defaults = %+v, want default interval/lookback/limit", cfg)
	}

	cfg = normalizeInternalAnalyticsRollupLoopConfig(InternalAnalyticsRollupLoopConfig{
		Interval: time.Second,
		Lookback: time.Hour,
		Limit:    20000,
	})
	if cfg.Interval != minInternalAnalyticsRollupInterval || cfg.Lookback != 24*time.Hour || cfg.Limit != 10000 {
		t.Fatalf("normalized clamps = %+v, want min interval/lookback and max limit", cfg)
	}

	cfg = normalizeInternalAnalyticsRollupLoopConfig(InternalAnalyticsRollupLoopConfig{Lookback: 60 * 24 * time.Hour, Limit: 5})
	if cfg.Lookback != 30*24*time.Hour || cfg.Limit != 5 {
		t.Fatalf("normalized large lookback/explicit limit = %+v, want 30d and limit 5", cfg)
	}
}

func TestInternalAnalyticsRunRollupLoopOnceUsesDateLookbackAndProvidedClock(t *testing.T) {
	store := &fakeInternalAnalyticsStore{deleted: 2}
	svc := newTestInternalAnalytics(store)
	now := time.Date(2026, 5, 17, 15, 45, 0, 0, time.FixedZone("WITA", 8*3600))
	store.rollupRows = []db.InternalAnalyticsEvent{{
		OccurredAt:    pgtype.Timestamptz{Time: now.Add(-2 * time.Hour), Valid: true},
		EventGroup:    "auth",
		EventName:     "auth.login_success",
		SourceSurface: "core_api",
		ActorRole:     pgtype.Text{String: " admin ", Valid: true},
		Result:        pgtype.Text{String: " success ", Valid: true},
	}}

	svc.runRollupLoopOnce(context.Background(), InternalAnalyticsRollupLoopConfig{Lookback: 72 * time.Hour, Limit: 25}, now)
	if len(store.rollupCalls) != 1 || len(store.upsertCalls) != 1 || len(store.deleteCalls) != 1 {
		t.Fatalf("rollup/upsert/delete calls = %d/%d/%d, want one each", len(store.rollupCalls), len(store.upsertCalls), len(store.deleteCalls))
	}
	call := store.rollupCalls[0]
	wantEnd := now.UTC()
	wantStart := internalAnalyticsDateOnly(wantEnd.Add(-72 * time.Hour))
	if !call.StartAt.Time.Equal(wantStart) || !call.EndAt.Time.Equal(wantEnd) || call.LimitCount != 25 {
		t.Fatalf("rollup call = %+v, want start %v end %v limit 25", call, wantStart, wantEnd)
	}
	if !store.deleteCalls[0].Time.Equal(wantEnd) {
		t.Fatalf("delete cutoff = %v, want %v", store.deleteCalls[0].Time, wantEnd)
	}
	upsert := store.upsertCalls[0]
	if upsert.Role != "admin" || upsert.Result != "success" || upsert.EventCount != 1 {
		t.Fatalf("upsert = %+v, want trimmed role/result and one count", upsert)
	}
}

func TestInternalAnalyticsStartRollupLoopRunImmediatelyCanBeCancelled(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	cancelLoop := svc.StartRollupLoop(ctx, InternalAnalyticsRollupLoopConfig{RunImmediately: true, Limit: 3})
	defer cancelLoop()

	deadline := time.After(2 * time.Second)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for len(store.rollupCalls) == 0 {
		select {
		case <-deadline:
			t.Fatal("StartRollupLoop(RunImmediately) did not invoke rollup before deadline")
		case <-tick.C:
		}
	}
	cancelLoop()
	if store.rollupCalls[0].LimitCount != 3 {
		t.Fatalf("immediate rollup limit = %d, want configured limit 3", store.rollupCalls[0].LimitCount)
	}
}

func TestInternalAnalyticsPublicMetadataSanitizersDropUnsafeInputs(t *testing.T) {
	metadata := sanitizeInternalAnalyticsPublicMetadata(map[string]any{
		"CTA Group":            " Admissions/PPDB ",
		"DeviceClass":          "Desktop",
		"unsafe":               "ignored",
		"result":               "https://example.test/?token=secret",
		"search_bucket":        "12345678",
		"search_length_bucket": float64(42),
		"form_step":            -1,
		"page_kind":            1.5,
	})
	if metadata["cta_group"] != "admissions_ppdb" || metadata["device_class"] != "desktop" || metadata["search_length_bucket"] != int64(42) {
		t.Fatalf("sanitized metadata = %+v, want normalized safe tokens and bounded integer float", metadata)
	}
	if _, ok := metadata["unsafe"]; ok {
		t.Fatalf("sanitized metadata kept unallowlisted key: %+v", metadata)
	}
	if _, ok := metadata["result"]; ok {
		t.Fatalf("sanitized metadata kept URL-like result: %+v", metadata)
	}
	if _, ok := metadata["search_bucket"]; ok {
		t.Fatalf("sanitized metadata kept sensitive digit sequence: %+v", metadata)
	}
	if got := sanitizeInternalAnalyticsPublicResult("", map[string]any{"result": "validation_failed"}); got != "validation_failed" {
		t.Fatalf("sanitize result fallback = %q, want validation_failed", got)
	}
	if got := sanitizeInternalAnalyticsPublicToken("token@example.test", "fallback"); got != "fallback" {
		t.Fatalf("sanitize token(email) = %q, want fallback", got)
	}
}

func TestInternalAnalyticsCreatePublicEventUsesSanitizedRouteResultAndMetadata(t *testing.T) {
	store := &fakeInternalAnalyticsStore{}
	svc := newTestInternalAnalytics(store)
	_, err := svc.CreatePublicEvent(context.Background(), CreatePublicAnalyticsEventInput{
		EventName:  " public.page_view ",
		RouteGroup: " PPDB/Home ",
		Result:     "success",
		Metadata: map[string]any{
			"device_class": "Mobile Phone",
			"raw_query":    "token=secret",
		},
	})
	if err != nil {
		t.Fatalf("CreatePublicEvent() error = %v", err)
	}
	if len(store.calls) != 1 {
		t.Fatalf("CreatePublicEvent store calls = %d, want 1", len(store.calls))
	}
	call := store.calls[0]
	if call.EventGroup != "public" || call.SourceSurface != "public_website" || !call.RouteGroup.Valid || call.RouteGroup.String != "ppdb_home" || !call.Result.Valid || call.Result.String != "success" {
		t.Fatalf("public event params = %+v, want public source with sanitized route/result", call)
	}
	var metadata map[string]any
	if err := json.Unmarshal(call.Metadata, &metadata); err != nil {
		t.Fatalf("metadata unmarshal: %v", err)
	}
	if metadata["device_class"] != "mobile_phone" {
		t.Fatalf("metadata = %+v, want sanitized device_class", metadata)
	}
	if _, ok := metadata["raw_query"]; ok {
		t.Fatalf("metadata = %+v, want raw_query dropped", metadata)
	}
}

func TestInternalAnalyticsClockAndCleanupInputHelpers(t *testing.T) {
	local := time.Date(2026, 5, 17, 23, 59, 0, 0, time.FixedZone("WITA", 8*3600))
	if got := internalAnalyticsDateOnly(local); !got.Equal(time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("internalAnalyticsDateOnly(%v) = %v, want UTC date boundary", local, got)
	}
	if got := (&InternalAnalytics{now: func() time.Time { return local }}).clockNow(); got.Location() != time.UTC || !got.Equal(local.UTC()) {
		t.Fatalf("clockNow() = %v (%v), want UTC fixed clock", got, got.Location())
	}

	normalized := normalizeInternalAnalyticsRollupCleanupInput(InternalAnalyticsRollupCleanupInput{Limit: -1}, fixedAnalyticsNow())
	if normalized.Limit != 10000 || !normalized.EndAt.Equal(internalAnalyticsDateOnly(fixedAnalyticsNow())) || !normalized.StartAt.Equal(normalized.EndAt.AddDate(0, 0, -1)) || !normalized.CutoffAt.Equal(fixedAnalyticsNow()) {
		t.Fatalf("normalized cleanup input = %+v, want default dates/cutoff/limit", normalized)
	}
	start := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	normalized = normalizeInternalAnalyticsRollupCleanupInput(InternalAnalyticsRollupCleanupInput{StartAt: start, EndAt: start, Limit: 20000}, fixedAnalyticsNow())
	if !normalized.EndAt.Equal(start.Add(24*time.Hour)) || normalized.Limit != 10000 {
		t.Fatalf("normalized invalid range/large limit = %+v, want end advanced by 24h and limit clamped", normalized)
	}
}
