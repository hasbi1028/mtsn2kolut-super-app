package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type fakeClockHealthDB struct {
	fakeHealthDB
	row fakeClockHealthRow
	sql string
}

func (f *fakeClockHealthDB) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	f.sql = sql
	return f.row
}

type fakeClockHealthRow struct {
	now      time.Time
	timezone string
	err      error
}

func (r fakeClockHealthRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*(dest[0].(*time.Time)) = r.now
	*(dest[1].(*string)) = r.timezone
	return nil
}

func TestHealthClockHealthIncludesDatabaseClockDiagnostics(t *testing.T) {
	serverNow := time.Date(2026, time.May, 17, 12, 0, 0, 123000000, time.FixedZone("WITA", 8*60*60))
	dbNow := serverNow.Add(90 * time.Second)
	pool := &fakeClockHealthDB{row: fakeClockHealthRow{now: dbNow, timezone: "Asia/Makassar"}}
	h := &Health{pool: pool}

	clock := h.clockHealth(context.Background(), serverNow)

	if pool.sql != "SELECT now(), current_setting('TimeZone')" {
		t.Fatalf("clock query SQL = %q, want current database time query", pool.sql)
	}
	if clock["status"] != "ok" {
		t.Fatalf("clock status = %v, want ok", clock["status"])
	}
	if clock["server_local"] != serverNow.Format(time.RFC3339) || clock["server_utc"] != serverNow.UTC().Format(time.RFC3339) {
		t.Fatalf("server clock fields = local:%v utc:%v, want formatted server times", clock["server_local"], clock["server_utc"])
	}
	if clock["db_now"] != dbNow.Format(time.RFC3339) || clock["db_timezone"] != "Asia/Makassar" || clock["db_server_drift_ms"] != int64(90000) {
		t.Fatalf("database clock fields = %+v, want db time/timezone/drift", clock)
	}
}

func TestHealthClockHealthDegradesOnLargeDriftOrQueryFailure(t *testing.T) {
	serverNow := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.UTC)

	largeDrift := (&Health{pool: &fakeClockHealthDB{row: fakeClockHealthRow{now: serverNow.Add(-3 * time.Minute), timezone: "UTC"}}}).clockHealth(context.Background(), serverNow)
	if largeDrift["status"] != "degraded" || largeDrift["db_server_drift_ms"] != int64(180000) {
		t.Fatalf("large drift clock = %+v, want degraded with absolute drift", largeDrift)
	}

	queryFailure := (&Health{pool: &fakeClockHealthDB{row: fakeClockHealthRow{err: errors.New("scan failed")}}}).clockHealth(context.Background(), serverNow)
	if queryFailure["status"] != "degraded" || queryFailure["db_error"] != "clock query failed" {
		t.Fatalf("query failure clock = %+v, want degraded query error", queryFailure)
	}
	if _, ok := queryFailure["db_now"]; ok {
		t.Fatalf("query failure clock = %+v, want no db_now on failed scan", queryFailure)
	}
}

func TestHealthClockHealthSkipsDatabaseClockWhenPoolCannotQuery(t *testing.T) {
	serverNow := time.Date(2026, time.May, 17, 12, 0, 0, 0, time.UTC)
	clock := (&Health{pool: fakeHealthDB{}}).clockHealth(context.Background(), serverNow)

	if clock["status"] != "ok" || clock["server_utc"] != serverNow.UTC().Format(time.RFC3339) {
		t.Fatalf("clock = %+v, want ok server-only diagnostics", clock)
	}
	if _, ok := clock["db_now"]; ok {
		t.Fatalf("clock = %+v, want no database fields when pool cannot query rows", clock)
	}
}
