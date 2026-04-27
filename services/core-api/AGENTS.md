# Core API Policy

This service owns the domain model, PostgreSQL schema, migrations, queue state, scheduler, and API contracts.

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

## Product Context

- This backend must support staged growth toward CBT and school operations.
- Build master data first, then exam authoring, then scheduling, then execution flows.
