# Question Level Normalization Option C Implementation Plan

> **For Hermes:** Use test-driven-development skill for backend behavior and verify SvelteKit UI with check/build.

**Goal:** Normalize Bank Soal/Paket Builder level metadata so `target_level` (VII/VIII/IX) is the single source of truth while `grade_level` remains temporary internal compatibility.

**Architecture:** Public UI/API writes `target_level` only. Backend derives/synchronizes `grade_level` for legacy compatibility. Readiness, Paket Builder filters, and official UX use `target_level`; destructive DB column removal is deferred.

**Tech Stack:** Go core-api, PostgreSQL/sqlc, SvelteKit web-admin, PM2.

---

## Tahap 0 — Audit & Safety Baseline

**Objective:** Confirm current schema/data/code usage before normalization.

**Checks:**
- Count mismatch `target_level` vs `grade_level` in `cbt_questions`.
- Search code references to `grade_level` and `target_level`.
- Confirm no credentials are printed.

**Expected output:** mismatch counts only, no connection strings.

**Audit result 2026-05-15:**
- `total_questions=34`
- `missing_target_level=34`
- `legacy_grade_present=34`
- `mismatched=0`
- Code still has legacy `grade_level` references, so `grade_level` must remain internal compatibility for now.

## Tahap 1 — Normalize without dropping columns

**Objective:** Make `target_level` the official field now, keep `grade_level` derived/internal.

**Backend:**
- Add idempotent migration `105_cbt_question_level_sync.sql`:
  - Backfill `target_level` from `grade_level` when missing.
  - Sync `grade_level` from `target_level` where missing/mismatched.
  - Add/refresh check constraint for `target_level` values VII/VIII/IX/null.
- Update create/update normalization:
  - Accept `target_level` as official input.
  - If `target_level` exists, derive `grade_level` 7/8/9.
  - If `target_level` missing but legacy `grade_level` provided, derive `target_level` for compatibility.
  - If both provided but mismatched, prefer `target_level` and sync `grade_level`.
- Update tests for target-level wins, legacy grade_level compatibility, invalid target_level rejection.

**Frontend:**
- Bank Soal form shows one field: `Kelas/Tingkat Soal`.
- Remove separate `Tingkat angka` field from simple/visible UX.
- Frontend sends `target_level`; derived `grade_level` may be omitted or synchronized only internally.
- Badge/filter labels consistently use `Kelas/Tingkat Soal`.

**Verification:**
```bash
cd services/core-api
/home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml
go test ./internal/handler ./internal/service ./internal/repository/postgres
go build -o /tmp/core-api-question-level-normalization ./cmd/api
cd ../..
npm --prefix apps/web-admin run check
npm --prefix apps/web-admin run build
```

## Tahap 2 — Hide `grade_level` from public API responses

**Objective:** Public API/UI consumers see `target_level` as the single field. DB/query structs can still hold `grade_level` internally.

**Backend:**
- Update public question serializers to omit `grade_level` from JSON maps.
- Keep `target_level` in list/detail/package pool/package question responses.
- Add handler tests proving public question list/detail response does not include `grade_level`.

**Frontend:**
- Keep TypeScript `grade_level?: number | null` optional for temporary backward compatibility only.
- New UI must not depend on `grade_level` for display/save.

**Verification:** same as Tahap 1.

## Tahap 3 — Audit and normalize import/export/reporting paths

**Objective:** Ensure any non-interactive surface (template/import/export/reporting/print/readiness) uses `target_level`, not legacy `grade_level`.

**Scope:**
- Search remaining `grade_level` references and classify:
  - allowed internal DB/sqlc/query compatibility;
  - allowed legacy request input/fallback;
  - not allowed public/export/reporting output.
- For import templates or exports, prefer column/key `target_level` / `tingkat` with values `VII`, `VIII`, `IX`.
- For reporting/print/readiness labels, display `target_level` and use `grade_level` only as fallback for old data/link compatibility.
- Add tests around any export/public reporting serializer changed in this stage.

**Verification:** same as Tahap 1.

**Tahap 3 result 2026-05-15:**
- Bank Soal CSV export/template now uses `target_level`; `grade_level` header is blocked by tests.
- CSV import accepts official `target_level`/`tingkat` values (`VII`, `VIII`, `IX`) and still accepts legacy numeric `grade_level`/`kelas` fallback.
- CBT package snapshot metadata no longer writes `grade_level` for newly snapshotted package questions.
- Kelengkapan Soal composer links now prefill `target_level`, not `grade_level`.
- Mapel/KD coverage and internal analytics documentation prefer `target_level`; frontend `grade_level` references left only as optional legacy fallback.

## Tahap 4 — Final dependency removal and DB column drop

**Objective:** Remove `grade_level` from application dependencies and drop `cbt_questions.grade_level` only after code no longer compiles against it.

**Safety rules:**
- Re-audit all `grade_level` references first.
- Replace DB/query filtering with `target_level` semantics.
- Keep request compatibility only through request/body/CSV parsing helpers that convert legacy numeric input into `target_level`; do not write/read `cbt_questions.grade_level`.
- Create a DB backup before production migration.
- Validate migration in a transaction before production apply.
- Run sqlc after query changes to prove no generated Go code depends on the dropped column.

**Verification:** same as Tahap 1 plus production health and protected route smoke tests after deploy.

**Tahap 4 implementation result 2026-05-15:**
- Added migration `services/core-api/db/migrations/106_drop_cbt_questions_grade_level.sql` to drop the legacy `cbt_questions.grade_level` column and `idx_cbt_questions_grade_level`, with rollback notes.
- Removed app DB read/write dependency on `cbt_questions.grade_level`; sqlc-generated structs no longer expose `GradeLevel` for `cbt_questions` query results/params.
- Kept legacy compatibility only at request/CSV parsing boundaries: incoming `grade_level` can still be translated to `target_level`, but it is not persisted to a DB column.
- Updated readiness/completeness SQL to compare question `target_level` with class/event roman level values.
- Updated Bank Soal frontend display paths to rely on `target_level`; optional public `grade_level` fallback was removed from the main Bank Soal display models.
- Production dry-run transaction passed: after applying migration inside `BEGIN`, `information_schema.columns` returned `grade_level_still_exists_after_migration = false`, then `ROLLBACK`.
- Verification passed before production apply: sqlc generate, Go handler/service/repository tests, Go build, Svelte check, and SvelteKit production build.
