# Core API Policy

This service owns the domain model, PostgreSQL schema, migrations, queue state, scheduler, and API contracts.

## Current Baseline — 2026-05-03

- Core API is still the only implemented PostgreSQL owner.
- Library and Inventory routes are grouped behind `admin`/`staf` access.
- Kesiswaan student photo access is scoped through `CanReadKesiswaanStudentPhoto`.
- PUSAKA jobs recover stale `running` rows before claim/scheduler work.
- CBT question answer keys and tokens must stay role-redacted outside admin/authorized author contexts.
- CBT scoring must not mutate `submitted_at`; duplicate submit must remain an explicit conflict.
- `/api/cbt/questions/*` remains the canonical question contract even though web UI `/cbt/questions` is retired.

## Must Keep

- Go code must stay idiomatic and clean.
- `sqlc` is the required pattern for typed SQL access.
- PostgreSQL is owned only here.
- Schema changes must go through migrations.
- Generated repository code must remain generated from `db/queries` and `db/sqlc.yaml`.

## Structure

- handlers: HTTP concerns only
- services: use cases and orchestration
- `db/queries`: explicit SQL
- `internal/repository/postgres`: generated `sqlc` output

## Guardrails

- Do not introduce ORM-led runtime access patterns.
- Do not move SQL into handlers.
- Do not bypass migrations with ad hoc schema edits.
- Prefer transaction boundaries in services where multi-step writes must be atomic.
- Keep error mapping explicit and predictable.
- Keep mobile-facing exam response envelopes and HTTP status semantics covered by handler tests.
- Trust `X-Forwarded-For` only through configured `TRUSTED_PROXY_CIDRS`; rate limits must not accept arbitrary forwarded client IPs from the public internet.
- Keep upload/file handlers on shared allowlist + `nosniff` helpers; do not add ad hoc file serving that bypasses MIME/extension validation or safe disposition.
- Keep BYOD/student-facing handlers (`/api/exam/*`) wrapped with `http.MaxBytesReader` on every JSON-decoded body so a tampered mobile client cannot memory-flood the backend.
- Every migration that introduces a new RBAC permission MUST also `INSERT INTO rbac_role_permissions` for the `admin` role in the same migration. The seed in `069_dynamic_rbac_foundation.sql` only runs once, so later permissions silently miss admin unless explicitly granted (`076_profile_change_review_permission.sql` is the reference pattern).
- Routes that hardcode `adminAccessAllowed(r)` at the handler should also use `requireAdmin` at the route guard. Don't gate admin-only mutators with broad `requireCbt`-style middlewares — the handler check is correct but the route gate becomes misleading and easy to break in refactors.
- HTML sanitization for editorial/UGC content MUST go through `bluemonday` (allowlist) — see `internal/service/website.go` `websiteHTMLPolicy`. Do not add ad hoc regex-based deny-list sanitizers; they consistently miss bypasses (`data:`/`vbscript:` URLs, `style=` payloads, exotic whitespace in event handlers, unbalanced tags).

## Product Context

- This backend must support staged growth toward CBT and school operations.
- Build master data first, then exam authoring, then scheduling, then execution flows.
