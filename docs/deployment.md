# Deployment

## Runtime Topology

This monorepo is deployed to three separate VPS targets:

- frontend VPS for `apps/web-admin`
- backend VPS for `services/core-api`
- worker VPS for `services/pusaka-worker`

## Rules

- One repository does not mean one server.
- Backend owns migration execution.
- Frontend and worker should be deployable independently after backend compatibility is ensured.

## Safe Order

1. Deploy backend code
2. Run PostgreSQL migrations from backend path
3. Restart backend and verify health
4. Deploy frontend
5. Deploy worker

## Must Keep

- Do not run PostgreSQL migrations from frontend or worker paths.
- Do not introduce frontend-owned runtime database state.
- Do not let worker bypass backend API contracts.

For operational details, see `deploy/DEPLOY.md`.
