# Deployment

## Runtime Topology

This monorepo is deployed to three separate VPS targets:

- frontend VPS for `apps/web-admin`
- backend VPS for `services/core-api`
- worker VPS for `services/pusaka-worker`

`apps/mobile` is not deployed as a VPS service. It is built as an internal Flutter Android APK and distributed through school-controlled channels for BYOD CBT trials.

## Rules

- One repository does not mean one server.
- Backend owns migration execution.
- Frontend and worker should be deployable independently after backend compatibility is ensured.
- Flutter APK rollout should be treated as a client release and checked against `docs/exam-api.md`.

## Safe Order

1. Deploy backend code
2. Run PostgreSQL migrations from backend path
3. Set or verify backend env such as `TRUSTED_PROXY_CIDRS` and `WORKER_API_KEY`
4. Restart backend and verify health
5. Deploy frontend
6. Set or verify worker env, then deploy worker

For CBT exam windows, add a post-deploy rehearsal:

1. Use `docs/cbt-operator-runbook.md` as the end-to-end operator sequence.
2. Run the admin/guru visibility checklist in `docs/cbt-smoke-checklist.md`.
3. Optionally run the web-admin staging scaffold in `apps/web-admin/docs/cbt-role-smoke.md` when staging credentials and browser tooling are available.
4. Verify at least one Flutter token login, answer save, heartbeat, and submit.

Token reprint language for CBT deploys:

- Migration `060_cbt_exam_token_hardening.sql` does not rotate existing tokens. It only rejects duplicate non-empty legacy tokens and enforces uniqueness for stored non-empty tokens.
- Reprint exam cards only after an explicit operator regeneration/repair changes participant tokens, or after card-visible data such as room/seat changes.
- If migration 060 fails because duplicate tokens exist, repair or regenerate only the affected draft/scheduled session tokens, then print the affected cards from the repaired state.

## CBT Migration Preflight

Run this preflight on staging first, then production, before CBT migrations that affect tokens, participants, seats, event members, or question scope.

1. Run a PostgreSQL backup first with `make ops-backup` from the repo root or the backend VPS backup script.
2. Confirm no active exam window is running or about to start. Do not migrate while students may login, heartbeat, save answers, or submit.
3. Check duplicate participant tokens before migration 060:

```sql
SELECT token, COUNT(*)
FROM cbt_exam_participants
WHERE token <> ''
GROUP BY token
HAVING COUNT(*) > 1;
```

4. Check invalid seats before migration 061:

```sql
SELECT id, session_id, room_id, seat_no
FROM cbt_exam_participants
WHERE seat_no IS NOT NULL
  AND seat_no <= 0;
```

5. Check duplicate seats before migration 061:

```sql
SELECT room_id, seat_no, COUNT(*)
FROM cbt_exam_participants
WHERE room_id IS NOT NULL
  AND seat_no IS NOT NULL
GROUP BY room_id, seat_no
HAVING COUNT(*) > 1;
```

6. Migration `061_cbt_participant_seat_invariants.sql` stops if non-positive seats or duplicate `(room_id, seat_no)` values exist, then adds a positive-seat constraint and unique room-seat index.
7. Migration `062_cbt_event_members_question_scope.sql` adds event members with roles `panitia`, `pembuat_soal`, `reviewer`, `proktor`, `pengawas`, and `korektor`; only `pembuat_soal`, `reviewer`, and `korektor` may carry `subject_id` scope.
8. After migration, run backend health, frontend smoke, and CBT smoke checks before activating or reopening exam sessions.

## Makefile Target Map

Use these targets from the repo root unless noted otherwise:

| Target/command | Purpose |
|----------------|---------|
| `make ops-backup` | Run PostgreSQL backup before migration/deploy. |
| `make db-migrate` | Apply PostgreSQL migrations using production/staging `DATABASE_URL` or backend `.env`. |
| `make db-migrate-local` | Apply migrations to an explicitly allowed local development database. |
| `make ops-health` | Check backend, frontend, and worker health shortcuts. |
| `make ops-health-backend` | Check backend health only. |
| `make ops-health-frontend` | Check frontend health only. |
| `make ops-health-worker` | Check worker health only. |
| `make check` | Run the full repository check suite currently wired in Makefile. |
| `make ci-check` | Run the lighter CI-style checks without vet/audit extras. |
| `make check-web` | Run SvelteKit check for web-admin. |
| `make test-mobile` | Run Flutter tests in `apps/mobile`. |
| `make check-mobile` | Run Flutter analyzer in `apps/mobile`. |
| `make mobile-release-apk API_BASE_URL=https://api.example.sch.id` | Build signed/unsigned release APK according to local Android signing setup and HTTPS backend base URL. |
| `cd apps/mobile && flutter pub get` | Install Flutter dependencies before analyze/test/build. |
| `cd apps/mobile && flutter build apk --release --dart-define=API_BASE_URL=https://...` | Direct release APK build command when not using Makefile. |

## Must Keep

- Do not run PostgreSQL migrations from frontend or worker paths.
- Do not introduce frontend-owned runtime database state.
- Do not let worker bypass backend API contracts.
- Do not route live Flutter exam traffic through SvelteKit; the app should use the backend exam API directly.

## Security Env Checklist

Backend VPS:

- `TRUSTED_PROXY_CIDRS` must contain only trusted reverse proxy/load balancer CIDRs that are allowed to supply `X-Forwarded-For`; leave empty for direct backend access.
- `WORKER_API_KEY` must be present and must match the worker VPS value before worker restart.
- Upload/file responses must keep backend MIME/extension allowlist validation and `X-Content-Type-Options: nosniff` behavior intact.

Worker VPS:

- `WORKER_API_KEY` is required outside local/test environments; do not log or commit it.
- `WORKER_API_TIMEOUT_MS` should be explicit for production if the default `10000` ms is too short/long for the backend link; valid range is 1000-60000 ms.
- `WORKER_LOG_PATH` should point to a non-public service log path with rotation/retention.
- `SCREENSHOT_DIR` should point to a non-public directory with restricted permissions; screenshots are operational evidence and must not be served by web-admin.

For operational details, see `deploy/DEPLOY.md`.

## Rate Limiting and Proxy Scaling Notes

Current sensitive-entrypoint rate limiting is intentionally in-memory and per backend process. This is acceptable for the current single backend VPS/process deployment, but it is not a multi-instance scaling boundary.

Before adding extra backend replicas, PM2 cluster mode, or a load balancer that fans traffic out to multiple backend processes:

- Move rate-limit counters to a shared store such as Redis or PostgreSQL-backed advisory/counter storage, or enforce equivalent limits at the trusted edge proxy.
- Keep backend application limits enabled as a defense-in-depth layer after introducing an edge/shared limiter.
- Validate login, refresh, public registration, and exam token login limits in staging before exam windows.
- Document the source of truth for client IP extraction so frontend, backend, and reverse proxy logs can be correlated during incidents.

`TRUSTED_PROXY_CIDRS` must stay narrow. Only list reverse proxy or load balancer source CIDRs that the school controls. Do not add broad public CIDRs, CDN-wide ranges that are not actually the direct trusted hop, or `0.0.0.0/0`; otherwise forged `X-Forwarded-For` values can distort rate limiting and audit context.
