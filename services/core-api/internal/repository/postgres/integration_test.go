package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const integrationDBEnv = "CORE_API_INTEGRATION_DB"

const defaultIntegrationAdminDSN = "host=/var/run/postgresql user=servermtsn2kolut dbname=postgres sslmode=disable"

type integrationTestDB struct {
	Pool *pgxpool.Pool
	Q    *Queries
}

func setupIntegrationTestDB(t *testing.T) *integrationTestDB {
	t.Helper()

	if os.Getenv(integrationDBEnv) != "1" {
		t.Skipf("skipping PostgreSQL integration test; set %s=1 to enable", integrationDBEnv)
	}

	adminDSN := strings.TrimSpace(os.Getenv("CORE_API_TEST_ADMIN_DSN"))
	if adminDSN == "" {
		adminDSN = defaultIntegrationAdminDSN
	}
	if databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); databaseURL != "" && adminDSN == databaseURL {
		t.Fatalf("refusing to use DATABASE_URL as integration-test admin DSN; set CORE_API_TEST_ADMIN_DSN to a non-production admin database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	adminConfig, err := pgxpool.ParseConfig(adminDSN)
	if err != nil {
		t.Fatalf("parse admin DSN: %v", err)
	}
	adminPool, err := pgxpool.NewWithConfig(ctx, adminConfig)
	if err != nil {
		t.Fatalf("connect admin database: %v", err)
	}
	defer adminPool.Close()

	dbName := uniqueIntegrationDBName(t)
	if _, err := adminPool.Exec(ctx, `CREATE DATABASE `+quoteIdentifier(dbName)); err != nil {
		t.Fatalf("create temp integration database %q: %v", dbName, err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		_, _ = adminPool.Exec(cleanupCtx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, dbName)
		_, _ = adminPool.Exec(cleanupCtx, `DROP DATABASE IF EXISTS `+quoteIdentifier(dbName))
	})

	testConfig, err := pgxpool.ParseConfig(adminDSN)
	if err != nil {
		t.Fatalf("parse test DSN: %v", err)
	}
	testConfig.ConnConfig.Database = dbName
	testPool, err := pgxpool.NewWithConfig(ctx, testConfig)
	if err != nil {
		t.Fatalf("connect temp integration database %q: %v", dbName, err)
	}
	t.Cleanup(testPool.Close)

	applyMigrations(t, ctx, testPool)

	return &integrationTestDB{
		Pool: testPool,
		Q:    New(testPool),
	}
}

func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join("..", "..", "..", "db", "migrations", "*.sql"))
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("no migrations found")
	}
	sort.Strings(matches)

	for _, migration := range matches {
		sql, err := os.ReadFile(migration)
		if err != nil {
			t.Fatalf("read migration %s: %v", migration, err)
		}
		if err := execMigration(ctx, pool, string(sql)); err != nil {
			t.Fatalf("apply migration %s: %v", filepath.Base(migration), err)
		}
	}
}

func execMigration(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	if !strings.Contains(strings.ToUpper(sql), "CONCURRENTLY") {
		_, err := pool.Exec(ctx, sql)
		return err
	}

	for _, stmt := range strings.Split(sql, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func uniqueIntegrationDBName(t *testing.T) string {
	t.Helper()

	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("generate random database suffix: %v", err)
	}
	return fmt.Sprintf("core_api_it_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b[:]))
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
