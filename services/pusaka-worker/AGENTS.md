# Worker Policy

This service is the worker runtime for asynchronous job execution.

## Current Baseline — 2026-05-03

- Worker runtime must call only canonical `/api/pusaka/worker/*` routes.
- Backend now recovers stale `running` jobs, but the worker must still report `complete` or `fail` explicitly for every claimed job.
- Worker heartbeat remains the operational signal for status pages and health checks.
- Worker must not depend on SvelteKit or frontend routes.

## Role

- Claim jobs from the Go API
- Execute worker tasks safely
- Report completion, failure, and retry state back to the Go API

## Must Keep

- Never access PostgreSQL directly.
- Never become a second owner of business state.
- Keep communication direct to `services/core-api`.
- Keep concurrency configurable, but prefer safe defaults.
- Keep shutdown behavior graceful; avoid abandoning claimed jobs without a final fail attempt when possible.

## Implementation Direction

- Favor operational robustness over feature sprawl.
- Keep retries, claim loops, and failure reporting explicit.
- Avoid coupling worker internals to frontend code or UI concerns.
