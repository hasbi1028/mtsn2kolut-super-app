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

1. Print/regenerate token artifacts after backend migration changes.
2. Run the admin/guru visibility checklist in `docs/cbt-smoke-checklist.md`.
3. Verify at least one Flutter token login, answer save, heartbeat, and submit.

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
