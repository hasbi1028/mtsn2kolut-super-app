# CBT Standardization Sprint 0-5 Implementation Plan

> **For Hermes/Codex:** implement sequentially, keep SvelteKit as BFF only, Go core-api owns PostgreSQL, use explicit sqlc queries/migrations, no direct DB access from web-admin, no secrets in output.

**Goal:** Raise Bank Soal + Asesmen/CBT from mature foundation to operationally standardized madrasah CBT with readiness board, package lock/snapshot, token hardening, runtime randomization/auto-submit, result sync, and remediasi.

**Architecture:** Go core-api owns all business rules and DB changes through migrations/sqlc/service/handler. SvelteKit web-admin provides BFF routes and operator UI only. Production deploy order: backup -> migrations -> backend build/restart/health -> frontend build/restart/smoke.

**Baseline constraints:**
- Do not touch unrelated untracked old plan/seed files.
- Prefer additive migrations and compatibility aliases.
- Existing routes must keep working.
- Use madrasah terms in UI.
- Validate with: `npm --prefix apps/web-admin run check`; `cd services/core-api && /home/servermtsn2kolut/go/bin/sqlc generate -f db/sqlc.yaml && go test ./internal/handler ./internal/service ./internal/repository/postgres && go build -o /tmp/core-api-cbt-standardization ./cmd/api`.

---

## Sprint 0 — Safety Gates + Technical Baseline

**Objective:** Add/confirm documentation and route/API contracts for the CBT standardization work.

**Files:**
- Create/modify docs/contract plan as needed: `docs/contracts/cbt-standardization.md`
- Optionally update existing plan implementation notes.

**Acceptance:**
- Document final target flows and deploy/migration order.
- No runtime behavior change required.

---

## Sprint 1 — Operational Readiness Board

**Objective:** Make `/asesmen/persiapan` a real readiness board for selected/active CBT event/kegiatan.

**Backend:**
- Add read-only summary endpoint, e.g. `GET /api/asesmen/readiness` and/or `GET /api/asesmen/events/{id}/readiness`.
- Summary should include event/kegiatan, package readiness, question target status, sessions, participants, rooms/seats, proctors, cards/tokens/app readiness when available, and issue counts.
- Use existing tables/queries where possible; avoid heavy full scans.

**Frontend:**
- BFF route under `apps/web-admin/src/routes/api/asesmen/readiness/+server.ts`.
- Revise `apps/web-admin/src/routes/asesmen/persiapan/+page.svelte` to show:
  - event selector / current active event,
  - readiness checklist rows: identitas kegiatan, target, bank soal/paket, sesi, peserta, ruang/kursi, pengawas, kartu/token, aplikasi siswa, hasil previous if relevant,
  - statuses: `Siap`, `Perlu dilengkapi`, `Atensi`,
  - CTA `Lengkapi` to the relevant screen.

**Acceptance:**
- Read-only only; no production risk.
- Page handles empty events gracefully.

---

## Sprint 2 — Package Lock & Snapshot

**Objective:** Freeze package/question content for audit once a package is used in scheduled/active/finished sessions.

**Backend:**
- Add migration for package lock/snapshot if not already present. Suggested additive columns/tables:
  - `cbt_packages.locked_at`, `locked_by`, `lock_reason`, `snapshot_version`
  - `cbt_package_question_snapshots` or equivalent table with package_id/question_id/order/points/question_type/stem_html/stimulus_html/options/answer_key/rubric_html/explanation_html/metadata JSONB/snapshot_at.
- Create service method to lock/snapshot package transactionally.
- Automatically lock/snapshot when session transitions to `scheduled` or `active` if not already locked.
- Prevent package question mutations after lock, except explicit admin unlock not required for this sprint.
- Runtime should prefer snapshot content for exam delivery/scoring where practical, while preserving old behavior if no snapshot exists.

**Frontend:**
- Show lock status in package/session screens.
- Warn if attempting to edit locked package.

**Acceptance:**
- Existing packages still work.
- New scheduled/active sessions produce snapshot.
- Tests cover package lock guard and snapshot creation.

---

## Sprint 3 — Token Hardening

**Objective:** Reduce token exposure and prepare hashed-token operation without breaking current production sessions.

**Backend:**
- Add additive columns for hashed tokens/version/reveal/revoke metadata for participant tokens and room tokens.
- New token generation writes hash + metadata and may keep plaintext legacy column only as compatibility during migration.
- Login verifies against hashed token if hash exists, otherwise legacy plaintext fallback.
- List/detail endpoints should redact tokens by default; reveal/print endpoints can show tokens only when operationally intended.
- Audit token reveal/regeneration where audit infrastructure exists.

**Frontend:**
- Ensure general lists do not expose full token unintentionally.
- Print/reveal actions clearly labeled.
- Room token should be treated as pengawas-only operational code, not general student portal info if feasible.

**Acceptance:**
- Existing tokens remain usable through fallback.
- New/regenerated tokens use hash fields.
- Runtime exam login still passes.

---

## Sprint 4 — Runtime Randomization + Auto Submit

**Objective:** Enforce runtime randomization contract and add deadline finalization.

**Backend:**
- Persist per-participant option order when `randomize_options` is true. Reuse/extend existing participant/session runtime storage when possible.
- Enforce draw counts (`draw_pg_count`, `draw_essay_count`) if configured; store deterministic/reproducible draw log per participant/package/session.
- Ensure answer evaluation maps randomized option labels back to canonical answer keys.
- Add endpoint/job/service to finalize overdue sessions/participants:
  - auto-submit participants past deadline,
  - calculate scores,
  - log `auto_submit_deadline`.
- Add admin endpoint e.g. `POST /api/asesmen/sessions/{id}/finalize-overdue`.

**Frontend:**
- Add control/status in session detail/pelaksanaan for auto-finalize overdue participants.
- Show auto-submitted count where easy.

**Acceptance:**
- Scoring remains correct with random options.
- Overdue finalize is idempotent.

---

## Sprint 5 — Result Sync + Remediasi

**Objective:** Turn CBT results into academic follow-up.

**Backend:**
- Add read endpoint for remedial candidates by session/event/subject/class using score thresholds and optionally item metadata/KD/indicator gaps.
- Add sync endpoint to gradebook following existing non-test assessment patterns where available; if full gradebook mapping is too large, add a safe export/sync-preflight endpoint and UI status, not unsafe writes.
- Track sync status metadata if schema already supports it, otherwise add additive columns/table.

**Frontend:**
- Enhance `/asesmen/hasil` or session results with:
  - sync status/preflight,
  - remedial candidates,
  - export/report CTA,
  - links to detail session/results.

**Acceptance:**
- No unsafe broad grade overwrite.
- Preflight/report works even if final write requires operator confirmation.

---

## Final Verification + Production Deploy

1. Check worktree and diff scope.
2. Run validation commands above.
3. Create production DB backup before migrations.
4. Apply migrations using repo's standard migration command/script.
5. Build backend binary and restart `mtsn2kolut-core-api` PM2.
6. Health check `http://127.0.0.1:8080/health`.
7. Build web-admin and restart `mtsn2kolut-web-admin` PM2.
8. Smoke key pages/API locally:
   - `/asesmen/persiapan`
   - `/asesmen/pelaksanaan`
   - backend health
   - at least one readiness API if authenticated smoke feasible; otherwise curl unauthenticated should return auth boundary not 500.
9. Commit with clear message after verification/deploy evidence.
