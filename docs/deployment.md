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
3. Restart backend and verify health
4. Deploy frontend
5. Deploy worker

For CBT exam windows, add a post-deploy rehearsal:

1. Print/regenerate token artifacts after backend migration changes.
2. Run the admin/guru visibility checklist in `docs/cbt-smoke-checklist.md`.
3. Verify at least one Flutter token login, answer save, heartbeat, and submit.

## Must Keep

- Do not run PostgreSQL migrations from frontend or worker paths.
- Do not introduce frontend-owned runtime database state.
- Do not let worker bypass backend API contracts.
- Do not route live Flutter exam traffic through SvelteKit; the app should use the backend exam API directly.

For operational details, see `deploy/DEPLOY.md`.
